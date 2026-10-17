# Builds docs/assets/wordmark.svg and wordmark-dark.svg: light, widely spaced Space Grotesk caps
# (OFL) with the two O's drawn as booths, and a small Geist Mono "claude" label above.
# Usage: uvx --from fonttools --with brotli python docs/demo/wordmark.py <SpaceGrotesk[wght].ttf> docs/assets
import sys
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
from pathlib import Path
SG = sys.argv[1]; OUT = sys.argv[2]
caps = instantiateVariableFont(TTFont(SG), {"wght": 300})
mono = instantiateVariableFont(TTFont(Path(__file__).parent.parent / "assets/fonts/geist-mono.woff2"), {"wght": 700})
def glyph(font, ch, x, y, scale):
    gs = font.getGlyphSet(); name = font.getBestCmap()[ord(ch)]
    pen = SVGPathPen(gs); tp = TransformPen(pen, (scale, 0, 0, -scale, x, y))
    gs[name].draw(tp); return pen.getCommands(), gs[name].width * scale
INDIGO, INDIGO_D, AMBER = "#5B4BE0", "#8B7CFF", "#F5A524"
TRACK = 220            # letter spacing, font units (0.22 em): sign-over-the-door spacing
CAP, BASE = 700, 1000  # Space Grotesk cap height; baseline y
STROKE = 60            # booth outline, about the light stem (64)
BW, BH = 500, 714      # booth outer size: cap height plus the round overshoot
SIDE = 60              # booth side bearing, like a letter's
def build(ink, booth, label):
    parts = []; x = 0
    def letter(ch):
        nonlocal x
        d, adv = glyph(caps, ch, x, BASE, 1); parts.append(f'<path d="{d}" fill="{ink}"/>'); x += adv + TRACK
    def boothshape(lit):
        nonlocal x
        x += SIDE
        r = BW / 2; x0, y1, y0 = x, BASE, BASE - BH
        outer = f"M{x0:.1f} {y1:.1f}V{y0 + r:.1f}A{r:.1f} {r:.1f} 0 0 1 {x0 + BW:.1f} {y0 + r:.1f}V{y1:.1f}Z"
        if lit:
            parts.append(f'<path d="{outer}" fill="{AMBER}"/>')
            ww, wh = 200, 110
            parts.append(f'<rect x="{x0 + (BW - ww) / 2:.1f}" y="{y0 + BH * .3:.1f}" width="{ww}" height="{wh}" rx="28" fill="{ink}"/>')
        else:
            ri = r - STROKE; yi1 = y1 - STROKE
            inner = f"M{x0 + STROKE:.1f} {yi1:.1f}V{y0 + r:.1f}A{ri:.1f} {ri:.1f} 0 0 1 {x0 + BW - STROKE:.1f} {y0 + r:.1f}V{yi1:.1f}Z"
            parts.append(f'<path d="{outer}{inner}" fill="{booth}" fill-rule="evenodd"/>')
        x += BW + SIDE + TRACK
    letter("B"); boothshape(False); boothshape(True)
    for ch in "THS": letter(ch)
    width = x - TRACK
    lx, ly, ls = 50, BASE - CAP - 120, .24  # label baseline above the booths' tops
    for ch in "claude":
        d, adv = glyph(mono, ch, lx, ly, ls); parts.append(f'<path d="{d}" fill="{label}"/>'); lx += adv + 70
    top = ly - 560 * ls - 20
    return width, BASE + 20 - top, top, parts
def save(name, ink, booth, label):
    w, h, top, parts = build(ink, booth, label)
    pad = 20
    svg = f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{-pad} {top - pad:.1f} {w + 2*pad:.1f} {h + 2*pad:.1f}" role="img" aria-label="Claude Booths">{"".join(parts)}</svg>\n'
    open(f"{OUT}/{name}", "w").write(svg); print(name, round(w), round(h))
save("wordmark.svg", "#141220", INDIGO, INDIGO)
save("wordmark-dark.svg", "#F3F2FA", INDIGO_D, INDIGO_D)
