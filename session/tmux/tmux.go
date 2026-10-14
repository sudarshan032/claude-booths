package tmux

import (
	"bytes"
	"claude-booths/cmd"
	"claude-booths/log"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/creack/pty"
)

const (
	ProgramClaude = "claude"
	ProgramAider  = "aider"
	ProgramGemini = "gemini"
)

// agentOf names the agent a program command runs, by substring so wrappers, paths and flags
// ("~/bin/claude-trusted", "claude --model x") still count.
func agentOf(program string) string {
	p := strings.ToLower(program)
	for _, a := range []string{ProgramClaude, ProgramAider, ProgramGemini} {
		if strings.Contains(p, a) {
			return a
		}
	}
	return ""
}

// hasPermissionPrompt reports whether the agent is waiting for a yes on a permission or
// confirmation prompt (what auto-yes answers with Enter). Only the bottom of the pane counts:
// an answered question stays in scrollback.
func hasPermissionPrompt(program, content string) bool {
	switch agentOf(program) {
	case ProgramClaude:
		b := strings.Join(paneBottom(content, screenLines), "\n")
		return strings.Contains(b, "No, and tell Claude what to do differently") || // before 2.1
			(strings.Contains(b, "Do you want to") && strings.Contains(b, "1. Yes"))
	case ProgramAider:
		last := paneBottom(content, 1)
		return len(last) == 1 && strings.Contains(last[0], "(Y)es/(N)o")
	case ProgramGemini:
		return strings.Contains(strings.Join(paneBottom(content, screenLines), "\n"), "Yes, allow once")
	}
	return false
}

// TmuxSession represents a managed tmux session
type TmuxSession struct {
	// Initialized by NewTmuxSession
	//
	// The name of the tmux session and the sanitized name used for tmux commands.
	sanitizedName string
	program       string
	// ptyFactory is used to create a PTY for the tmux session.
	ptyFactory PtyFactory
	// cmdExec is used to execute commands in the tmux session.
	cmdExec cmd.Executor

	// Initialized by Start or Restore
	//
	// ptmx is a PTY is running the tmux attach command. This can be resized to change the
	// stdout dimensions of the tmux pane. On detach, we close it and set a new one.
	// This should never be nil.
	ptmx *os.File
	// client is the `tmux attach-session` process running in ptmx.
	client *exec.Cmd
	// monitor monitors the tmux pane content and sends signals to the UI when it's status changes
	monitor *statusMonitor

	// Output streaming. tmux pushes every screen change to the client running in ptmx, so a
	// single pump goroutine reads it: any output marks the pane dirty (and is forwarded to
	// stdout while attached). CapturePaneContent only runs `tmux capture-pane` when dirty,
	// so idle sessions cost no subprocesses instead of ~12 per second.
	pumpDone  chan struct{} // closed when the pump for the current ptmx exits
	pumpGen   atomic.Uint64 // which pump is current; a stale one must not touch streaming
	streaming atomic.Bool   // a pump is running, so dirty is trustworthy
	dirty     atomic.Bool   // output arrived since the last capture
	attached  atomic.Bool   // forward pump output to stdout
	cacheMu   sync.Mutex
	cache     *string // last capture-pane result

	// Initialized by Attach
	// Deinitilaized by Detach
	//
	// Channel to be closed at the very end of detaching. Used to signal callers.
	attachCh chan struct{}
	// While attached, we use some goroutines to manage the window size and stdin/stdout. This stuff
	// is used to terminate them on Detach. We don't want them to outlive the attached window.
	ctx    context.Context
	cancel func()
	wg     *sync.WaitGroup
}

const TmuxPrefix = "claudebooths_"

// StatusStyle is applied as each session's tmux status-style (the bar shown while attached,
// if status is on). Set by the app from the active theme; "" leaves tmux's own setting.
var StatusStyle string

// ApplyStatusStyle sets StatusStyle on this session.
func (t *TmuxSession) ApplyStatusStyle() {
	if StatusStyle == "" {
		return
	}
	cmd := exec.Command("tmux", "set-option", "-t", t.sanitizedName, "status-style", StatusStyle)
	if err := t.cmdExec.Run(cmd); err != nil {
		log.InfoLog.Printf("could not set status-style for %s: %v", t.sanitizedName, err)
	}
}

// ErrSessionNotFound is returned when the tmux session backing an instance is gone, which
// happens whenever the tmux server dies (reboot, crash, `tmux kill-server`).
var ErrSessionNotFound = errors.New("tmux session no longer exists")

var whiteSpaceRegex = regexp.MustCompile(`\s+`)

func toClaudeBoothsTmuxName(str string) string {
	str = whiteSpaceRegex.ReplaceAllString(str, "")
	str = strings.ReplaceAll(str, ".", "_") // tmux replaces all . with _
	return fmt.Sprintf("%s%s", TmuxPrefix, str)
}

// NewTmuxSession creates a new TmuxSession with the given name and program.
func NewTmuxSession(name string, program string) *TmuxSession {
	return newTmuxSession(name, program, MakePtyFactory(), cmd.MakeExecutor())
}

// NewTmuxSessionWithDeps creates a new TmuxSession with provided dependencies for testing.
func NewTmuxSessionWithDeps(name string, program string, ptyFactory PtyFactory, cmdExec cmd.Executor) *TmuxSession {
	return newTmuxSession(name, program, ptyFactory, cmdExec)
}

func newTmuxSession(name string, program string, ptyFactory PtyFactory, cmdExec cmd.Executor) *TmuxSession {
	return &TmuxSession{
		sanitizedName: toClaudeBoothsTmuxName(name),
		program:       program,
		ptyFactory:    ptyFactory,
		cmdExec:       cmdExec,
	}
}

// Start creates and starts a new tmux session, then attaches to it. Program is the command to run in
// the session (ex. claude). workdir is the git worktree directory.
func (t *TmuxSession) Start(workDir string) error {
	// Check if the session already exists
	if t.DoesSessionExist() {
		return fmt.Errorf("tmux session already exists: %s", t.sanitizedName)
	}

	// Create a new detached tmux session and start claude in it
	cmd := exec.Command("tmux", "new-session", "-d", "-s", t.sanitizedName, "-c", workDir, t.program)

	ptmx, err := t.ptyFactory.Start(cmd)
	if err != nil {
		// Cleanup any partially created session if any exists.
		if t.DoesSessionExist() {
			cleanupCmd := exec.Command("tmux", "kill-session", "-t="+t.sanitizedName)
			if cleanupErr := t.cmdExec.Run(cleanupCmd); cleanupErr != nil {
				err = fmt.Errorf("%v (cleanup error: %v)", err, cleanupErr)
			}
		}
		return fmt.Errorf("error starting tmux session: %w", err)
	}

	// Poll for session existence with exponential backoff
	timeout := time.After(2 * time.Second)
	sleepDuration := 5 * time.Millisecond
	for !t.DoesSessionExist() {
		select {
		case <-timeout:
			if cleanupErr := t.Close(); cleanupErr != nil {
				err = fmt.Errorf("%v (cleanup error: %v)", err, cleanupErr)
			}
			return fmt.Errorf("timed out waiting for tmux session %s: %v", t.sanitizedName, err)
		default:
			time.Sleep(sleepDuration)
			// Exponential backoff up to 50ms max
			if sleepDuration < 50*time.Millisecond {
				sleepDuration *= 2
			}
		}
	}
	_ = ptmx.Close()

	// Set history limit to enable scrollback (default is 2000, we'll use 10000 for more history)
	historyCmd := exec.Command("tmux", "set-option", "-t", t.sanitizedName, "history-limit", "10000")
	if err := t.cmdExec.Run(historyCmd); err != nil {
		log.InfoLog.Printf("Warning: failed to set history-limit for session %s: %v", t.sanitizedName, err)
	}

	// Enable mouse scrolling for the session
	mouseCmd := exec.Command("tmux", "set-option", "-t", t.sanitizedName, "mouse", "on")
	if err := t.cmdExec.Run(mouseCmd); err != nil {
		log.InfoLog.Printf("Warning: failed to enable mouse scrolling for session %s: %v", t.sanitizedName, err)
	}

	t.ApplyStatusStyle()

	err = t.Restore()
	if err != nil {
		if cleanupErr := t.Close(); cleanupErr != nil {
			err = fmt.Errorf("%v (cleanup error: %v)", err, cleanupErr)
		}
		return fmt.Errorf("error restoring tmux session: %w", err)
	}

	return nil
}

// Restore attaches to an existing session and restores the window size
func (t *TmuxSession) Restore() error {
	// attach-session against a missing session still forks a process successfully, so the
	// PTY start below would report no error while leaving us attached to nothing. Check
	// first so callers can tell "session is gone" apart from "PTY failed".
	if !t.DoesSessionExist() {
		return ErrSessionNotFound
	}

	client := exec.Command("tmux", "attach-session", "-t", t.sanitizedName)
	ptmx, err := t.ptyFactory.Start(client)
	if err != nil {
		return fmt.Errorf("error opening PTY: %w", err)
	}
	t.ptmx = ptmx
	t.client = client
	t.monitor = newStatusMonitor()
	t.startPump(ptmx)
	return nil
}

// startPump makes a goroutine the only reader of ptmx (see the streaming fields).
func (t *TmuxSession) startPump(ptmx *os.File) {
	done := make(chan struct{})
	t.pumpDone = done
	gen := t.pumpGen.Add(1)
	t.dirty.Store(true)
	t.streaming.Store(true)
	go func() {
		defer close(done)
		defer func() {
			if t.pumpGen.Load() == gen {
				t.streaming.Store(false)
			}
		}()
		buf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				t.dirty.Store(true)
				if t.attached.Load() {
					_, _ = os.Stdout.Write(buf[:n])
				}
			}
			if err != nil {
				return
			}
		}
	}()
}

type statusMonitor struct {
	// Store hashes to save memory.
	prevOutputHash []byte
	// answeredHash is the screen Enter was last tapped on, so auto-yes answers each prompt once.
	answeredHash []byte
}

func newStatusMonitor() *statusMonitor {
	return &statusMonitor{}
}

// hash hashes the string.
func (m *statusMonitor) hash(s string) []byte {
	h := sha256.New()
	// TODO: this allocation sucks since the string is probably large. Ideally, we hash the string directly.
	h.Write([]byte(s))
	return h.Sum(nil)
}

// TapEnter sends an enter keystroke to the tmux pane.
func (t *TmuxSession) TapEnter() error {
	t.monitor.answeredHash = t.monitor.prevOutputHash // HasUpdated won't report this screen's prompt again
	_, err := t.ptmx.Write([]byte{0x0D})
	if err != nil {
		return fmt.Errorf("error sending enter keystroke to PTY: %w", err)
	}
	return nil
}

func (t *TmuxSession) SendKeys(keys string) error {
	_, err := t.ptmx.Write([]byte(keys))
	return err
}

// HasUpdated checks if the tmux pane content has changed since the last tick. It also returns true if
// the tmux pane has a prompt for aider or claude code.
func (t *TmuxSession) HasUpdated() (updated bool, hasPrompt bool) {
	content, err := t.CapturePaneContent()
	if err != nil {
		log.ErrorLog.Printf("error capturing pane content in status monitor: %v", err)
		return false, false
	}

	hasPrompt = hasPermissionPrompt(t.program, content)

	h := t.monitor.hash(content)
	if hasPrompt && bytes.Equal(h, t.monitor.answeredHash) {
		hasPrompt = false // already answered; the agent just hasn't redrawn yet
	}
	if !bytes.Equal(h, t.monitor.prevOutputHash) {
		t.monitor.prevOutputHash = h
		return true, hasPrompt
	}
	return false, hasPrompt
}

func (t *TmuxSession) Attach() (chan struct{}, error) {
	t.attachCh = make(chan struct{})

	t.wg = &sync.WaitGroup{}
	t.wg.Add(1)
	t.ctx, t.cancel = context.WithCancel(context.Background())

	// The pump forwards ptmx output to stdout while attached. The first goroutine
	// terminates when that pump does (ptmx closed). We use the waitgroup to wait for it.
	// The 2nd one returns when you press escape to Detach. It doesn't need to be
	// in the waitgroup because is the goroutine doing the Detaching; it waits for
	// all the other ones.
	pumpDone := t.pumpDone
	t.attached.Store(true)
	go func() {
		defer t.wg.Done()
		<-pumpDone
		// When the pump returns, it means the connection was closed
		// This could be due to normal detach or Ctrl-D
		// Check if the context is done to determine if it was a normal detach
		select {
		case <-t.ctx.Done():
			// Normal detach, do nothing
		default:
			// If context is not done, it was likely an abnormal termination (Ctrl-D)
			// Print warning message
			fmt.Fprintf(os.Stderr, "\n\033[31mError: Session terminated without detaching. Use Ctrl-Q to properly detach from tmux sessions.\033[0m\n")
		}
	}()

	go func() {
		// Close the channel after 50ms
		timeoutCh := make(chan struct{})
		go func() {
			time.Sleep(50 * time.Millisecond)
			close(timeoutCh)
		}()

		// Read input from stdin and check for Ctrl+q
		buf := make([]byte, 4096) // large enough that pastes are forwarded in one write
		for {
			nr, err := os.Stdin.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				continue
			}

			// Nuke the first bytes of stdin, up to 64, to prevent tmux from reading it.
			// When we attach, there tends to be terminal control sequences like ?[?62c0;95;0c or
			// ]10;rgb:f8f8f8. The control sequences depend on the terminal (warp vs iterm). We should use regex ideally
			// but this works well for now. Log this for debugging.
			//
			// There seems to always be control characters, but I think it's possible for there not to be. The heuristic
			// here can be: if there's characters within 50ms, then assume they are control characters and nuke them.
			select {
			case <-timeoutCh:
			default:
				log.InfoLog.Printf("nuked first stdin: %s", buf[:nr])
				continue
			}

			// Ctrl+Q detaches. It can arrive inside a larger read (typed right after other
			// keys, or with a focus event), so look for it anywhere; forward what precedes it.
			if i := detachKeyIndex(buf[:nr]); i >= 0 {
				if i > 0 {
					_, _ = t.ptmx.Write(buf[:i])
				}
				t.Detach()
				return
			}

			// Forward other input to tmux
			_, _ = t.ptmx.Write(buf[:nr])
		}
	}()

	t.monitorWindowSize()
	return t.attachCh, nil
}

// detachKeyEncodings are the ways a terminal can send Ctrl+Q: the plain control byte, and
// the CSI-u / xterm modifyOtherKeys forms used when extended key reporting is on.
var detachKeyEncodings = [][]byte{{0x11}, []byte("\x1b[113;5u"), []byte("\x1b[27;5;113~")}

// detachKeyIndex returns where the first Ctrl+Q starts in b, or -1.
func detachKeyIndex(b []byte) int {
	first := -1
	for _, enc := range detachKeyEncodings {
		if i := bytes.Index(b, enc); i >= 0 && (first < 0 || i < first) {
			first = i
		}
	}
	return first
}

// restoreTerminalModes undoes modes the tmux client turned on in the real terminal. It is
// killed on detach rather than exiting cleanly, so without this its cursor stays visible
// over the TUI and focus reporting stays on. (Bracketed paste and mouse are left alone: the
// TUI uses them too.)
func restoreTerminalModes() {
	_, _ = os.Stdout.WriteString("\x1b[?25l\x1b[?1004l")
}

// stopClient ends the tmux client running in ptmx and waits for its pump to exit.
// Closing a pty master on macOS does not interrupt a Read blocked on it, so for an idle
// session (no output pending) the pump would wait forever and detaching would hang.
// Killing the client hangs up the pty, which ends the Read.
func (t *TmuxSession) stopClient() error {
	t.attached.Store(false)
	if c := t.client; c != nil && c.Process != nil {
		_ = c.Process.Kill()
		go func() { _ = c.Wait() }() // reap it
	}
	t.client = nil
	var err error
	if t.ptmx != nil {
		err = t.ptmx.Close()
	}
	if done := t.pumpDone; done != nil {
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			log.WarningLog.Printf("output pump for %s did not stop; continuing", t.sanitizedName)
		}
	}
	return err
}

// DetachSafely disconnects from the current tmux session without panicking
func (t *TmuxSession) DetachSafely() error {
	// Only detach if we're actually attached
	if t.attachCh == nil {
		return nil // Already detached
	}

	var errs []error

	// Close the attached pty session.
	if err := t.stopClient(); err != nil {
		errs = append(errs, fmt.Errorf("error closing attach pty session: %w", err))
	}
	t.ptmx = nil

	// Clean up attach state
	if t.attachCh != nil {
		close(t.attachCh)
		t.attachCh = nil
	}

	if t.cancel != nil {
		t.cancel()
		t.cancel = nil
	}

	if t.wg != nil {
		t.wg.Wait()
		t.wg = nil
	}

	t.ctx = nil

	restoreTerminalModes()

	if len(errs) > 0 {
		return fmt.Errorf("errors during detach: %v", errs)
	}
	return nil
}

// Detach disconnects from the current tmux session. It panics if detaching fails. At the moment, there's no
// way to recover from a failed detach.
func (t *TmuxSession) Detach() {
	// TODO: control flow is a bit messy here. If there's an error,
	// I'm not sure if we get into a bad state. Needs testing.
	defer func() {
		close(t.attachCh)
		t.attachCh = nil
		t.cancel = nil
		t.ctx = nil
		t.wg = nil
	}()

	// Close the attached pty session; the old pump must be gone before Restore starts the next one.
	err := t.stopClient()
	if err != nil {
		// This is a fatal error. We can't detach if we can't close the PTY. It's better to just panic and have the
		// user re-invoke the program than to ruin their terminal pane.
		msg := fmt.Sprintf("error closing attach pty session: %v", err)
		log.ErrorLog.Println(msg)
		panic(msg)
	}
	// Attach goroutines should die on EOF due to the ptmx closing. Call
	// t.Restore to set a new t.ptmx.
	if err = t.Restore(); err != nil {
		// This is a fatal error. Our invariant that a started TmuxSession always has a valid ptmx is violated.
		msg := fmt.Sprintf("error closing attach pty session: %v", err)
		log.ErrorLog.Println(msg)
		panic(msg)
	}

	// Cancel goroutines created by Attach.
	t.cancel()
	t.wg.Wait()
	restoreTerminalModes()
}

// Close terminates the tmux session and cleans up resources
func (t *TmuxSession) Close() error {
	var errs []error

	if err := t.stopClient(); err != nil && !errors.Is(err, os.ErrClosed) {
		errs = append(errs, fmt.Errorf("error closing PTY: %w", err))
	}
	t.ptmx = nil

	// A session whose program already exited is gone: that's the goal, not a failure.
	// "-t=" matches exactly; plain "-t" falls back to a prefix match and could kill another booth.
	if t.DoesSessionExist() {
		cmd := exec.Command("tmux", "kill-session", "-t="+t.sanitizedName)
		if err := t.cmdExec.Run(cmd); err != nil && t.DoesSessionExist() {
			errs = append(errs, fmt.Errorf("error killing tmux session: %w", err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}

	errMsg := "multiple errors occurred during cleanup:"
	for _, err := range errs {
		errMsg += "\n  - " + err.Error()
	}
	return errors.New(errMsg)
}

// Gone reports whether the session's tmux session has disappeared (its program exited, or
// someone killed it). Cheap while the output stream is alive: tmux is only asked once the
// stream has ended on its own.
func (t *TmuxSession) Gone() bool {
	return !t.streaming.Load() && !t.attached.Load() && !t.DoesSessionExist()
}

// SetDetachedSize set the width and height of the session while detached. This makes the
// tmux output conform to the specified shape.
func (t *TmuxSession) SetDetachedSize(width, height int) error {
	return t.updateWindowSize(width, height)
}

// updateWindowSize updates the window size of the PTY.
func (t *TmuxSession) updateWindowSize(cols, rows int) error {
	return pty.Setsize(t.ptmx, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
		X:    0,
		Y:    0,
	})
}

func (t *TmuxSession) DoesSessionExist() bool {
	// Using "-t name" does a prefix match, which is wrong. `-t=` does an exact match.
	existsCmd := exec.Command("tmux", "has-session", fmt.Sprintf("-t=%s", t.sanitizedName))
	return t.cmdExec.Run(existsCmd) == nil
}

// CapturePaneContent captures the content of the tmux pane
func (t *TmuxSession) CapturePaneContent() (string, error) {
	// Nothing streamed since the last capture: the pane is unchanged.
	if t.streaming.Load() && !t.dirty.Swap(false) {
		t.cacheMu.Lock()
		cached := t.cache
		t.cacheMu.Unlock()
		if cached != nil {
			return *cached, nil
		}
	}
	// Add -e flag to preserve escape sequences (ANSI color codes)
	cmd := exec.Command("tmux", "capture-pane", "-p", "-e", "-J", "-t", t.sanitizedName)
	output, err := t.cmdExec.Output(cmd)
	if err != nil {
		t.dirty.Store(true) // retry next time
		return "", fmt.Errorf("error capturing pane content: %v", err)
	}
	content := string(output)
	t.cacheMu.Lock()
	t.cache = &content
	t.cacheMu.Unlock()
	return content, nil
}

// CapturePaneContentWithOptions captures the pane content with additional options
// start and end specify the starting and ending line numbers (use "-" for the start/end of history)
func (t *TmuxSession) CapturePaneContentWithOptions(start, end string) (string, error) {
	// Add -e flag to preserve escape sequences (ANSI color codes)
	cmd := exec.Command("tmux", "capture-pane", "-p", "-e", "-J", "-S", start, "-E", end, "-t", t.sanitizedName)
	output, err := t.cmdExec.Output(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to capture tmux pane content with options: %v", err)
	}
	return string(output), nil
}

// CleanupSessions kills all tmux sessions that start with "session-"
func CleanupSessions(cmdExec cmd.Executor) error {
	// First try to list sessions
	cmd := exec.Command("tmux", "ls")
	output, err := cmdExec.Output(cmd)

	// If there's an error and it's because no server is running, that's fine
	// Exit code 1 typically means no sessions exist
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return nil // No sessions to clean up
		}
		return fmt.Errorf("failed to list tmux sessions: %v", err)
	}

	re := regexp.MustCompile(fmt.Sprintf(`%s.*:`, TmuxPrefix))
	matches := re.FindAllString(string(output), -1)
	for i, match := range matches {
		matches[i] = match[:strings.Index(match, ":")]
	}

	for _, match := range matches {
		log.InfoLog.Printf("cleaning up session: %s", match)
		if err := cmdExec.Run(exec.Command("tmux", "kill-session", "-t="+match)); err != nil {
			return fmt.Errorf("failed to kill tmux session %s: %v", match, err)
		}
	}
	return nil
}

// TerminalSessionName is the tmux session behind a booth's Terminal tab.
func TerminalSessionName(boothTitle string) string {
	return toClaudeBoothsTmuxName("term_" + boothTitle)
}

// CleanupOrphanTerminals kills Terminal-tab sessions (claudebooths_term_*) whose booth no
// longer exists, e.g. left behind by an older version or a crash. Agent sessions are never
// touched.
func CleanupOrphanTerminals(cmdExec cmd.Executor, boothTitles []string) {
	out, err := cmdExec.Output(exec.Command("tmux", "list-sessions", "-F", "#{session_name}"))
	if err != nil {
		return // no tmux server
	}
	keep := make(map[string]bool, len(boothTitles))
	for _, title := range boothTitles {
		keep[TerminalSessionName(title)] = true
	}
	for _, name := range strings.Fields(string(out)) {
		if !strings.HasPrefix(name, TmuxPrefix+"term_") || keep[name] {
			continue
		}
		if err := cmdExec.Run(exec.Command("tmux", "kill-session", "-t="+name)); err != nil {
			log.WarningLog.Printf("could not remove orphan terminal session %s: %v", name, err)
		}
	}
}
