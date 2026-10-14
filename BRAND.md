# Claude Booths brand and design system

The single source of truth for how Claude Booths looks and sounds. The website (`docs/index.html`),
the README, promo material and the TUI default theme (`ui/theme/theme.go`) all use these values.

## Idea

A **booth** is a small private room for one conversation. The product is a row of booths: each
has its own chat, git worktree and tmux session. A lit booth is one that is working. The brand is
that picture, drawn simply: arched booths in indigo, a lamp-amber booth for activity, on white.

## Logo

The o's in "booths" are booths: the first an empty booth, the second lit (a session is working).

| Asset | Use |
|---|---|
| <img src="docs/assets/wordmark.svg" width="220" alt=""> `docs/assets/wordmark.svg` | **Primary logo**: website, README, social, print. Light backgrounds |
| `docs/assets/wordmark-dark.svg` | Same, for dark backgrounds |
| <img src="docs/assets/logo-tile.svg" width="40" alt=""> `docs/assets/logo-tile.svg` (= `logo.svg`) | App icon, favicon, GitHub avatar |
| Pixel banner (in `ui/logo.go`) | The terminal's big banner on the welcome screen |
| `❯ claude/b[ ][▮]ths` prompt logo (`ui.PromptLogo`) | Everywhere else in the terminal: booth list title, help, narrow panes |

Primary wordmark construction: Geist Black (900) "b ths" with drawn booths, small Geist Mono Bold
"claude" label above in indigo. Letters keep their natural spacing (−12/1000 em tracking) and each
booth has 46/1000 em of air on both sides so letters never touch the booths. Glyphs are outlined,
so the SVGs need no font. Regenerate with `docs/demo/wordmark.py` if the shapes change.

- Minimum size: wordmark 96 px wide, tile 16 px. Clear space: the width of one booth on every side.
- Don't recolor, outline, stretch or rotate; never light the first booth; keep the lamp amber.

## Color

Two brand colors, one neutral ramp, two semantic colors. Nothing else.

| Role | Token | Hex | Use |
|---|---|---|---|
| **Primary** | `--primary` | `#5B4BE0` | Buttons, links, focus ring, selection, booth outlines. 6.2:1 on white |
| Primary hover | `--primary-hover` | `#4A3BCB` | Hover/pressed |
| Primary tint | `--primary-tint` | `#F1EFFF` | Subtle fills, step markers, closing band |
| **Secondary** | `--secondary` | `#F5A524` | The lit booth, "working" state. Decoration only, never text on white |
| Secondary text | `--secondary-text` | `#B45309` | Amber text on white (5.0:1), e.g. "Testing" pills |
| Secondary tint | `--secondary-tint` | `#FEF3D7` | Amber pill background |
| Background | `--bg` | `#FFFFFF` | Page |
| Surface | `--surface` / `--surface-2` | `#FAFAFC` / `#F4F4F7` | Inputs, code chips, segmented controls |
| Border | `--border` | `#E6E4EE` | Hairlines, table rules |
| Text | `--text` | `#141220` | Body and headings |
| Text secondary | `--text-2` | `#3F3C52` | Supporting copy. 10:1 on white |
| Success | `--ok` | `#15803D` | Verified, additions |
| Danger | `--bad` | `#DC2626` | Errors, deletions |

**Terminal colors** (the app runs in a dark terminal; the website's app window uses these):
background `#15131F`, primary `#8B7CFF`, secondary `#F5A524`, text `#E9E7F2`, dim `#8F8BA6`,
success `#5BD68A`, danger `#FF7A7A`, selection `#2B2650`. The TUI's `claude-booths` theme uses the
same primary/secondary with `AdaptiveColor` for light terminals.

## Type

One family, like the best developer-tool sites: **Geist** for everything (headings, body, UI) and
**Geist Mono** for code, keys and the app window. They're drawn as one family, so commands sit
naturally in running text. Hierarchy comes from size and weight only, never from a second face.

| Role | Face | Weight / tracking |
|---|---|---|
| Display (h1) | Geist | 700, −0.045em, line height 1.02 |
| Headings (h2/h3) | Geist | 700 / 600, −0.035em to −0.02em |
| Body and UI | Geist | 400 (650 for buttons and labels), normal tracking |
| Code, keys, app window | Geist Mono | 400 / 600 |

Both are OFL, self-hosted from `docs/assets/fonts/` and subset to Latin (17 + 13 KB), preloaded.
Scale (desktop): 76 / 46 / 26 / 20 / 17 / 15. Body line height 1.6, measure under 65ch,
`text-wrap: balance` on headings.

**Keycaps** are one component (`.k`): indigo-tinted cap with a bottom edge, Geist Mono 600.
Combos sit side by side (`ctrl` `q`); key + description pairs sit in an indigo-50 pill.

## Layout and components

- Container 1180 px, 16 px gutters on phones. Sections are separated by 112 px of space, not boxes.
- Radius: 8 px (small controls), 12 px (buttons, inputs, code), 20 px (media). Pills are fully round.
- Buttons: primary (indigo fill, white text) for the one main action per view; ghost (white, border) for the rest. Minimum height 44 px.
- Structure comes from hairlines (`--border`) and whitespace. No drop shadows except the app window.
- Icons: 1.8 px stroke line icons in `--primary`, inline SVG. No emoji as icons.
- Focus: 3 px indigo-100 halo + 1.5 px indigo ring on every interactive element.

## Motion

- Motion shows state: booths stream output, a lit booth means work in progress.
- 150–200 ms for hover and press; the demo only animates while it's on screen and the tab is visible.
- `prefers-reduced-motion`: no loops, the demo shows its finished state.

## Performance budget (website)

One HTML file with inline CSS/JS, ~29 KB of self-hosted fonts, inline SVG icons, no frameworks, no third-party
requests. Images are lazy and sized (no layout shift). Target: under 60 KB before images, first
paint in one round trip.

## Voice

Plain, short, concrete; say what happens on screen.

| Say | Don't say |
|---|---|
| booth, open a booth, step inside, step out | instance, spawn, attach/detach (in user copy) |
| chat, agent (the CLI running in a booth) | swarm, army, squad |
| "Each booth gets its own git worktree." | "Leverages isolated workspace paradigms." |

Only claim an agent or OS as supported after the full flow was tested on it (see the website's
agents table). "Claude" is Anthropic's trademark: never use Anthropic's logo, colors or type, and
keep *"Not affiliated with or endorsed by Anthropic."* wherever the brand appears publicly.

## Assets

| File | Use | Regenerate |
|---|---|---|
| `docs/assets/logo.svg`, `logo-tile.svg` | Mark, tile, favicon | by hand |
| `docs/assets/fonts/*.woff2` | Geist, Geist Mono (OFL) | pyftsubset to Latin, see git log |
| `docs/assets/demo.gif` | README demo of the real app | `vhs docs/demo/demo.tape` |
| `docs/assets/promo.gif`, `promo.mp4` | 6 s loop for posts and ads (1200×630) | `docs/demo/render-promo.sh` |
| `docs/assets/og.png` | Social preview card | `docs/demo/render-promo.sh` |
| `docs/index.html` | Website (GitHub Pages from `/docs`) | by hand |
