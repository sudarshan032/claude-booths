# Claude Booths brand and design system

One source of truth for how Claude Booths looks and sounds: the app (`ui/theme`), the website
(`docs/index.html`), the README and any promo material use these values.

## Idea

A **booth** is a small private room for one conversation. The app is a row of booths: each one has
its own chat, its own git worktree and its own tmux session. A lit booth is one that's working.
Everything in the brand comes from that picture: arched booths, a warm lamp for activity, a mint
floor line that ties them together.

## Logo

<img src="docs/assets/logo.svg" width="96" alt="Claude Booths logo">

`docs/assets/logo.svg`: three arched booths on a mint floor; the middle booth is lit and holds a
chat bubble. Use it on dark backgrounds as it is (it carries its own ink tile).

- Minimum size 24 px. Keep clear space of at least 1/8 of its width on every side.
- Don't recolor, stretch, rotate, add effects or put it in another container.
- The wordmark is the name set in Space Grotesk Bold next to the logo: **Claude Booths**.
  In code, paths and commands it is always lowercase: `claude-booths`.

## Color

| Token | Hex | Role |
|---|---|---|
| Ink | `#0F0D1A` | Background (web, promo); text on Accent in the app |
| Ink 2 / Ink 3 | `#16132A` / `#1C1834` | Raised surfaces, cards, code blocks |
| Line | `#2A2640` | Borders, grid, separators |
| Text | `#ECEAF6` | Body text |
| Muted | `#8D88A8` | Secondary text, hints |
| **Booth Violet** | `#8B7CFF` | Primary accent: borders, selection, links, focus |
| **Lamp** | `#FFC861` | Activity, "lit" booths, highlighted actions, prompts |
| **Mint** | `#4DF0C0` | Success, ready state, additions, the floor line |
| Coral | `#FF6B8B` | Errors, deletions, destructive confirms |
| Sky | `#7FD4FF` | Info, diff hunks |

Light terminals use Violet `#5B4BE0` and darker neutrals (see the `claude-booths` palette in
`ui/theme/theme.go`). Violet + Lamp + Mint are the brand; use Coral and Sky only for meaning.
Brand gradient (headlines only): Violet → Sky → Mint → Lamp at 100°.

## Type

- **Space Grotesk** (500, 700) for headings and UI text on the web. Headlines are tight: `letter-spacing: -0.04em`.
- **JetBrains Mono** (400, 600) for commands, keys and code. Small uppercase mono labels (`letter-spacing: .14em`) in Mint are section eyebrows.
- The terminal app uses whatever font your terminal uses.

## Shape and motion

- Radius 18 px for cards, 14 px for buttons and inputs, 99 px for pills. Booths are arches (full semicircle top).
- Motion means activity: booths light up in turn, previews stream. Ease out, 0.25–0.8 s; loops are slow (6 s+).
- Always honor `prefers-reduced-motion`: no loops, no tilt, content visible immediately.

## Voice

Plain, short, concrete. Say what happens on screen. Use the product's own words:

| Say | Don't say |
|---|---|
| booth, open a booth, step inside, step out | instance, session spawn, attach/detach (in user-facing copy) |
| chat, agent (the CLI running in a booth) | swarm, army, squad |
| "Each booth gets its own git worktree." | "Leverages isolated workspace paradigms." |

"Claude" is Anthropic's trademark. Never use Anthropic's logo, colors or type, and keep the line
*"Not affiliated with or endorsed by Anthropic."* wherever the brand appears publicly.

## Assets

| File | Use | Regenerate |
|---|---|---|
| `docs/assets/logo.svg` | Logo, favicon | edit by hand |
| `docs/assets/demo.gif` | README / site demo of the real app | `vhs docs/demo/demo.tape` |
| `docs/assets/promo.gif`, `promo.mp4` | 6 s loop for posts and ads (1200×630 source) | `docs/demo/render-promo.sh` |
| `docs/assets/og.png` | Social preview card (1200×630) | `docs/demo/render-promo.sh` |
| `docs/index.html` | Website (GitHub Pages, `/docs`) | edit by hand |

The demo uses a scripted stand-in agent (`docs/demo/agent`) in a throwaway `HOME`, so recordings
are repeatable and never show real projects or accounts.
