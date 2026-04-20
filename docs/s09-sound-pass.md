# S09 Sound Pass — assets + wiring plan

The S09 polish-pass shipped GSAP integration, tap/untap animation,
hand deal-in/out, ETB pulse, damage popups, and combat arrows. The
last checklist item — **sound** — was deferred until the asset pack
existed.

**Status: ✅ done (S09 sound pass closed).** The Suno pack landed
as MP3s with two takes per event plus a bonus ambient loop
(#106), the playback module (`play`, `setMuted`,
`armAudioOnFirstGesture`) landed alongside it, and every Svelte
seam + the toolbar mute toggle are wired (#107).

Deviations from the original Suno brief:
- **MP3, not WAV.** Smaller files, universally supported; the
  module doesn't care about container.
- **Two takes per event.** `sounds.ts` exposes them as a variant
  array and picks uniformly at random on each `play()`, which
  breaks up the obvious-loop effect on rapid-fire events (e.g.
  the opening-hand draw flurry).
- **Damage variant discrepancy.** The source pack delivered
  `damage popup 1` and `damage output 2` (no matching partners);
  both were taken as generic damage takes — noted in the sounds
  [README](../client/public/sounds/README.md) in case a future
  pass wants to split them back apart.

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

## 2. Wiring plan

### Asset placement — ✅ landed
- Files live in [client/public/sounds/](../client/public/sounds/)
  as MP3, two variants per event (e.g. `draw-1.mp3`, `draw-2.mp3`).
  Vite serves them at `/sounds/<name>-<n>.mp3` with no extra config.
- [client/public/sounds/README.md](../client/public/sounds/README.md)
  lists source + license (Suno-generated, personal use) and the
  full event → filename table.

### Sound module — ✅ landed in [client/src/lib/sounds.ts](../client/src/lib/sounds.ts)
- Registry: `SOUND_MANIFEST: Record<SoundName, string[]>` — each
  event points at an array of variant URLs; `play()` picks one at
  random.
- `armAudioOnFirstGesture()` installs a one-shot pointerdown/keydown
  listener that preloads + decodes every variant (autoplay-policy
  unlock). Call once on app mount.
- `play(name, opts?: { volume?: number; rate?: number })` clones the
  preloaded Audio so overlapping plays don't cut each other off.
- `isMuted()` / `setMuted(bool)` / `toggleMuted()` — persisted to
  `localStorage["cmdctrl.muted"]`.
- Rate-limit: same-name calls within 40ms are dropped to prevent
  the opening-hand draw flurry from stacking seven sounds.

Explicit mute short-circuits `play()` to a no-op. Reduced-motion
is NOT auto-coupled — S11.5 can layer its own preference on top
via `setMuted()`.

### Event hooks — ✅ landed (#107)
Each existing GSAP / state seam fires one `play(...)` call:

| Seam | Sound | Notes |
|---|---|---|
| Card.svelte `animateTap` $effect | `tap` | Fires only on the untapped→tapped edge; initial mount + manual untap stay silent. |
| Hand.svelte `use:handLifecycle` on `.deal-wrap` | `draw` / `play` | Action's mount = draw, destroy = play. Mass events (opening hand, mulligan) collapse via 40ms rate-limit. |
| `etbPulse` action mount | — | No sound; the scale-pulse rides on top of the `play` cue. |
| PlayerHeader life_history popup spawn | `damage` / `heal` | Picked from `last.delta` sign. |
| `declare_attacker` dispatch in Game.svelte | `attack` | |
| `declare_blocker` dispatch in Game.svelte | `block` | |
| Step-change $effect (`turn.step` watcher) into `combat_damage` | `combat_resolve` | |
| Step-change $effect into `untap` | `untap_all` + `turn_change` | |
| `shuffle_library` / `mulligan` / `mulliganDecide` dispatches | `shuffle` | |
| Game-end $effect (`state === "ended"` edge) | `win` / `loss` | Viewer-relative; spectators stay silent. |

### UI — ✅ landed (#107)
- Mute toggle (🔊/🔇) lives in the left `.toolbar-group` of
  `Game.svelte` — wired to `toggleMuted()` / `isMuted()`,
  persisted to `localStorage["cmdctrl.muted"]`.

### Verification — ✅ done
- Manual 2-player walkthrough confirmed cues fire on the correct
  edges with no snapshot-rebuild double-triggers.
- `armAudioOnFirstGesture()` is called from a bare `$effect` on
  `Game.svelte` mount — any pointerdown / keydown on the page
  unlocks the audio context before the first cue.

### Risks — resolved in `sounds.ts`
- Simultaneous same-name calls: **handled** — `play()` rate-limits
  repeat triggers of the same name within 40ms.
- Browser audio decode latency on first hit: **handled** —
  `armAudioOnFirstGesture()` pre-decodes every variant, and `play()`
  clones from the pool.

---

Kept here as the historical record of the S09 sound pass — the
Suno prompt list (§1) + the event → cue mapping (§2) are both
useful to future-us if the sound design gets refreshed.
