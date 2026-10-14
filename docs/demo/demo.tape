# Regenerate docs/assets/demo.gif and the website screenshots:
#   vhs docs/demo/demo.tape
# Needs vhs, ttyd, ffmpeg, go, tmux and a logged-in Claude Code. Real Claude Code answers, so
# long waits are cut with Hide/Show; everything on screen is what the app really showed.
Output docs/assets/demo.gif

Set Shell bash
Set FontFamily "Menlo"
Set FontSize 16
Set Width 1440
Set Height 900
Set Padding 24
Set Margin 32
Set MarginFill "#F1EFFF"
Set WindowBar Colorful
Set BorderRadius 12
Set TypingSpeed 35ms
Set Framerate 24
Set Theme { "name": "booths", "background": "#15131F", "foreground": "#E9E7F2", "cursor": "#8B7CFF", "selection": "#2B2650", "black": "#15131F", "red": "#FF7A7A", "green": "#5BD68A", "yellow": "#F5A524", "blue": "#8B7CFF", "magenta": "#B4A9FF", "cyan": "#7FD4FF", "white": "#E9E7F2", "brightBlack": "#5D5874", "brightRed": "#FF9C9C", "brightGreen": "#86E3A9", "brightYellow": "#FFC266", "brightBlue": "#ADA2FF", "brightMagenta": "#CFC7FF", "brightCyan": "#A6E1FF", "brightWhite": "#FFFFFF" }

Hide
Type "source docs/demo/setup.sh && clear" Enter
Sleep 6s
Type "claude-booths" Enter
Sleep 2s
Show
Sleep 2.5s

# booth 1
Type "N" Sleep 500ms
Type "fix-redirect" Sleep 300ms Enter Sleep 700ms
Type "Fix the open redirect in src/auth.ts: only allow same-site paths. Keep it short." Sleep 500ms
Screenshot docs/assets/shot-new.png
Tab Enter Sleep 200ms Enter
Sleep 4s

# booth 2
Type "N" Sleep 400ms
Type "dark-mode" Sleep 300ms Enter Sleep 600ms
Type "Add a dark mode toggle to the Header that remembers the choice. Keep it short." Sleep 400ms
Tab Enter Sleep 200ms Enter
Sleep 4s

# booth 3
Type "N" Sleep 400ms
Type "checkout-tests" Sleep 300ms Enter Sleep 600ms
Type "Write node:test tests for total() in src/checkout.js and run them. Keep it short." Sleep 400ms
Tab Enter Sleep 200ms Enter
Sleep 5s

# everyone is working
Up Sleep 2s
Up Sleep 2s
Screenshot docs/assets/shot-booths.png

# skip the wait (real Claude Code takes ~30-60 s)
Hide
Sleep 75s
Show
Sleep 1.5s
Down Sleep 2s
Down Sleep 2s
Up Up Sleep 2s

# review the diff
Tab Sleep 3s
Screenshot docs/assets/shot-diff.png
Tab Tab Sleep 800ms

# step inside, ask for more, step out
Enter Sleep 1.5s
Type "Thanks. Add a one-line comment explaining the check." Sleep 400ms Enter
Sleep 4s
Ctrl+Q Sleep 2s

# themes
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 800ms
Type "t" Sleep 2.5s
