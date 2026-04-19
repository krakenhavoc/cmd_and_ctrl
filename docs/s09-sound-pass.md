# S09 Sound Pass — assets + wiring plan

The S09 polish-pass shipped GSAP integration, tap/untap animation,
hand deal-in/out, ETB pulse, damage popups, and combat arrows. The
last checklist item — **sound** — is deferred until the asset pack
exists. This doc captures both halves so the work picks up cleanly:

1. The Suno prompts a friend with Suno is generating from
2. The wiring plan once the WAVs land

Not a blocker for closing S09; the visual pass is a complete S09
deliverable on its own.

---

## 1. Suno prompts (one per event)

Format request for every clip:
- Mono WAV, 44.1 kHz
- Filenames as listed (e.g. `draw.wav`)
- Quiet master with headroom — many SFX layer in close succession
- No music tail, no fades — clean ends only
- Under 1 second except the game-over cues

| Event | Filename | Prompt |
|---|---|---|
| Card draw | `draw.wav` | Quick paper card sliding off a deck, soft fwip, dry, 0.4 seconds, no music, foley only |
| Card play | `play.wav` | Heavy paper card slapping onto a wood table, satisfying thud with subtle low end, 0.6 seconds, foley |
| Tap (turn 90°) | `tap.wav` | Soft mechanical click, leather creak, 0.2 seconds, intimate, no reverb |
| Untap all (start of turn) | `untap_all.wav` | Series of three quick wooden ticks ascending in pitch, 0.5 seconds total, hollow, no music |
| Damage / life loss | `damage.wav` | Low impact thud with a metallic ringing tail, ominous, 0.5 seconds, no music |
| Heal / life gain | `heal.wav` | Soft warm chime, single bell tone with a gentle sparkle, 0.6 seconds, hopeful |
| Attack declared | `attack.wav` | Sharp blade unsheathe, single tense steel ring, 0.5 seconds, dry, no reverb |
| Block declared | `block.wav` | Solid shield brace, wood-on-metal clack, 0.4 seconds, grounded |
| Combat damage resolves | `combat_resolve.wav` | Heavy crack of a battle hit followed by a brief crowd gasp, 0.8 seconds, cinematic but contained |
| Turn change | `turn_change.wav` | Brief gong-like tone with soft ambient sweep, 1.0 second, ceremonial but quiet, suitable as repeating cue |
| Mulligan / shuffle | `shuffle.wav` | Hands shuffling a deck of cards, riffle bridge, 1.2 seconds, foley, intimate mic |
| Game over (win) | `win.wav` | Single triumphant horn note that holds and gently fades, 1.5 seconds, warm |
| Game over (loss) | `loss.wav` | Descending muted brass minor third, 1.2 seconds, somber but short |

---

## 2. Wiring plan (once WAVs land)

### Asset placement
- Drop all `*.wav` files into `client/public/sounds/` so Vite serves
  them under `/sounds/<name>.wav` with no extra config.
- Add a `client/public/sounds/README.md` listing source + license
  attribution (Suno-generated for this project, personal use).

### Sound module — `client/src/lib/sounds.ts`
- Single registry of `name → HTMLAudioElement` pre-loaded on first
  user gesture (autoplay-policy unlock).
- `play(name: string, opts?: { volume?: number; rate?: number })`
  helper that clones a fresh `Audio` per call so overlapping plays
  don't cut each other off (e.g. two cards drawn back-to-back).
- Mute toggle persisted to `localStorage["cmdctrl.muted"]`.
- Reduced-motion / explicit-mute short-circuits to a no-op.

### Event hooks
Each existing GSAP / state seam gets one `play(...)` call:

| Seam | Sound |
|---|---|
| `animateTap` (Card.svelte $effect) | `tap` |
| `dealIn` first-fire in Hand | `draw` |
| `dealOut` first-fire in Hand | `play` |
| `etbPulse` action mount | (no sound — visual only, matches play) |
| PlayerHeader life_history popup spawn | `damage` if delta < 0, `heal` if > 0 |
| Combat-mode `declare_attacker` action | `attack` |
| Combat-mode `declare_blocker` action | `block` |
| AdvanceStep into `combat_damage` (snapshot detect) | `combat_resolve` |
| AdvanceStep into `untap` (turn-cursor detect) | `untap_all` + `turn_change` |
| `mulligan` / `shuffle_library` action | `shuffle` |
| GameView.state transitions to `ended` | `win` if viewer survives, `loss` otherwise |

### UI
- Mute toggle in the `Game.svelte` toolbar (small icon button) so
  every page has one click to silence.

### Verification
- Manual: play a full 2-player turn end-to-end, listen for each
  event firing exactly once. No double-trigger on snapshot rebuilds.
- The autoplay-policy unlock (any `pointerdown` arms the audio
  context) needs to land BEFORE the first sound — easiest: install
  it inside the GameClient init or on the route mount.

### Risks
- Many events firing simultaneously (e.g. opening hand of 7 → 7
  draw sounds at once) — rate-limit at the play() seam (drop calls
  for the same `name` within 40ms).
- Browser audio decode latency on first hit — pre-decode all
  sounds during the unlock gesture, not lazily.

---

Once assets + module land, this file moves to `docs/decisions/`
as the audio-design ADR (or just gets deleted if the work is small
enough not to warrant a permanent record).
