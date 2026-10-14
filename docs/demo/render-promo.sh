#!/bin/bash
# Renders promo.html to docs/assets/promo.gif, promo.mp4 and og.png (needs Google Chrome and ffmpeg).
set -euo pipefail
cd "$(dirname "$0")"
CHROME=${CHROME:-"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"}
FPS=15
frames=$(mktemp -d)
for i in $(seq 0 $((6 * FPS - 1))); do
  t=$(awk "BEGIN { printf \"%.4f\", $i / $FPS }")
  "$CHROME" --headless=new --disable-gpu --hide-scrollbars --force-device-scale-factor=1 --virtual-time-budget=1500 \
    --window-size=1200,630 --screenshot="$frames/$(printf %03d "$i").png" "file://$PWD/promo.html?t=$t" 2>/dev/null
done
cp "$frames/074.png" ../assets/og.png  # every booth lit and done
ffmpeg -v error -y -framerate $FPS -i "$frames/%03d.png" -vf "scale=900:-1:flags=lanczos,split[a][b];[a]palettegen=stats_mode=diff[p];[b][p]paletteuse=dither=bayer:bayer_scale=4" -loop 0 ../assets/promo.gif
ffmpeg -v error -y -framerate $FPS -i "$frames/%03d.png" -c:v libx264 -pix_fmt yuv420p -crf 20 -movflags +faststart ../assets/promo.mp4
rm -rf "$frames"
