# Regenerate docs/assets/demo.gif:  vhs docs/demo/demo.tape   (needs vhs, ttyd, ffmpeg, go, tmux)
Output docs/assets/demo.gif

Set Shell bash
Set FontFamily "Menlo"
Set FontSize 16
Set Width 1440
Set Height 960
Set Padding 28
Set Margin 36
Set MarginFill "#1C1834"
Set WindowBar Colorful
Set BorderRadius 14
Set TypingSpeed 40ms
Set Framerate 30
Set Theme { "name": "booths", "background": "#0F0D1A", "foreground": "#ECEAF6", "cursor": "#8B7CFF", "selection": "#2B2550", "black": "#0F0D1A", "red": "#FF6B8B", "green": "#4DF0C0", "yellow": "#FFC861", "blue": "#7FD4FF", "magenta": "#8B7CFF", "cyan": "#7FD4FF", "white": "#ECEAF6", "brightBlack": "#5A5577", "brightRed": "#FF8FA8", "brightGreen": "#7DF5D2", "brightYellow": "#FFD98A", "brightBlue": "#A6E1FF", "brightMagenta": "#ADA2FF", "brightCyan": "#A6E1FF", "brightWhite": "#FFFFFF" }

Hide
Type "source docs/demo/setup.sh && clear" Enter
Sleep 4s
Show

Type "claude-booths" Sleep 400ms Enter
Sleep 2s

# booth 1
Type "N" Sleep 500ms
Type "fix-login" Sleep 300ms Enter Sleep 600ms
Type "Fix the login redirect after sign-in" Sleep 400ms
Tab Enter Sleep 200ms Enter
Sleep 2.5s

# booth 2
Type "N" Sleep 400ms
Type "dark-mode" Sleep 300ms Enter Sleep 600ms
Type "Add a dark mode toggle to the header" Sleep 400ms
Tab Enter Sleep 200ms Enter
Sleep 2.5s

# booth 3
Type "N" Sleep 400ms
Type "checkout-tests" Sleep 300ms Enter Sleep 600ms
Type "Write tests for the checkout flow" Sleep 400ms
Tab Enter Sleep 200ms Enter
Sleep 5s

# hop between booths: every preview is live
Up Sleep 1.5s
Up Sleep 1.5s

# review the diff
Tab Sleep 2.5s
Tab Tab Sleep 800ms

# step inside a booth, talk, step out with Ctrl+Q
Enter Sleep 1.2s
Type "also cover the /admin route" Sleep 300ms Enter
Sleep 3.5s
Ctrl+Q Sleep 1.5s

# eight built-in themes
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 900ms
Type "t" Sleep 2.5s
