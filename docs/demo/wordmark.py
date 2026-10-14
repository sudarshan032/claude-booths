# Builds docs/assets/wordmark.svg and wordmark-dark.svg from Geist (OFL) outlines.
# Usage: uvx --from fonttools --with brotli python docs/demo/wordmark.py <dir with geist-full.woff2 + geistmono-full.woff2> docs/assets
import sys
from fontTools.ttLib import TTFont
from fontTools.varLib.instancer import instantiateVariableFont
from fontTools.pens.svgPathPen import SVGPathPen
from fontTools.pens.transformPen import TransformPen
S = sys.argv[1]; OUT = sys.argv[2]
def inst(path, w):
    return instantiateVariableFont(TTFont(path), {"wght": w})
heavy = inst(f"{S}/geist-full.woff2", 900); mono = inst(f"{S}/geistmono-full.woff2", 700)
def glyph(font, ch, x, y, scale):
    gs = font.getGlyphSet(); name = font.getBestCmap()[ord(ch)]
    pen = SVGPathPen(gs); tp = TransformPen(pen, (scale, 0, 0, -scale, x, y))
    gs[name].draw(tp); return pen.getCommands(), gs[name].width * scale
INDIGO, INDIGO_D, AMBER = "#5B4BE0", "#8B7CFF", "#F5A524"
def build(ink, booth, label, track=-12, size=1.0, stacked=True):
    # units: em = 1000 * size. track in font units between glyphs.
    sc = size; x = 0; base = 1000 * sc if stacked else 760 * sc
    parts = []
    xh = 530 * sc; stem = 128 * sc               # x-height, booth stroke ~ Geist Black stem
    bw, bh = 470 * sc, 600 * sc                  # booth outer width/height (a little taller than x-height)
    gap = 46 * sc                                # space around booths, so letters don't crowd
    def letter(ch):
        nonlocal x
        d, adv = glyph(heavy, ch, x, base, sc); parts.append(f'<path d="{d}" fill="{ink}"/>'); x += adv + track * sc
    def boothshape(lit):
        nonlocal x
        x += gap
        r = bw / 2; x0, y1 = x, base; y0 = base - bh
        outer = f"M{x0:.1f} {y1:.1f}V{y0 + r:.1f}A{r:.1f} {r:.1f} 0 0 1 {x0 + bw:.1f} {y0 + r:.1f}V{y1:.1f}Z"
        if lit:
            parts.append(f'<path d="{outer}" fill="{AMBER}"/>')
            ww, wh = 170 * sc, 96 * sc
            parts.append(f'<rect x="{x0 + (bw - ww) / 2:.1f}" y="{y0 + bh * .3:.1f}" width="{ww:.1f}" height="{wh:.1f}" rx="{24 * sc:.1f}" fill="{ink}"/>')
        else:
            ri = r - stem; xi0 = x0 + stem; yi1 = y1 - stem * .9
            inner = f"M{xi0:.1f} {yi1:.1f}V{y0 + r:.1f}A{ri:.1f} {ri:.1f} 0 0 1 {x0 + bw - stem:.1f} {y0 + r:.1f}V{yi1:.1f}Z"
            parts.append(f'<path d="{outer}{inner}" fill="{booth}" fill-rule="evenodd"/>')
        x += bw + gap
    letter("b"); boothshape(False); boothshape(True)
    for ch in "ths": letter(ch)
    width = x
    if stacked:
        lx = 18 * sc; ly = base - 830 * sc  # label baseline above the word, clear of the b ascender
        for ch in "claude":
            d, adv = glyph(mono, ch, lx, ly, sc * .34); parts.append(f'<path d="{d}" fill="{label}"/>'); lx += adv
        top = ly - 560 * sc * .34 - 30 * sc
    else:
        top = base - 760 * sc
    h = base + 30 * sc - top
    return width, h, top, parts
def save(name, ink, booth, label, stacked=True, bg=None):
    w, h, top, parts = build(ink, booth, label, stacked=stacked)
    pad = 20
    bgr = f'<rect x="{-pad}" y="{top - pad:.1f}" width="{w + 2*pad:.1f}" height="{h + 2*pad:.1f}" fill="{bg}"/>' if bg else ""
    svg = f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="{-pad} {top - pad:.1f} {w + 2*pad:.1f} {h + 2*pad:.1f}" role="img" aria-label="Claude Booths">{bgr}{"".join(parts)}</svg>\n'
    open(f"{OUT}/{name}", "w").write(svg); print(name, round(w), round(h))
save("wordmark.svg", "#141220", INDIGO, INDIGO)
save("wordmark-dark.svg", "#F3F2FA", INDIGO_D, INDIGO_D)
