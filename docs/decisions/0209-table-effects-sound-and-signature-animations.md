# ADR 0209 — Table effects: sound categories, effect cues and signature animations

**Status:** Accepted · 2026-10-27 · S76 — Cards: incubate, face-down and outside the game.
**Issue:** [#4399](https://github.com/krakenhavoc/cmd_and_ctrl/issues/4399) (the epic: #4394–#4398).
**Numbering:** checked on 2026-10-27. `develop` ends at 0208. `git ls-tree` over every remote head's `docs/decisions/` shows no file above 0208; the heads checked are listed in the PR.
**Builds on:** [docs/s09-sound-pass.md](../s09-sound-pass.md) (the sound registry, two takes per event), [ADR 0119](0119-a-stack-you-can-follow.md) §3 (information that motion carries must still arrive without it), [ADR 0134](0134-combat-you-can-watch.md) (the combat cue director and the `animations.combat` setting, the precedent for §1's setting), DESIGN.md §4 (motion and reduced motion).

---

## Context

The table plays 13 sounds (`client/src/lib/sounds.ts`). The settings have a master volume, an effects volume and a music volume, plus one mute toggle. The game animations in `client/src/lib/animations.ts` cover tapping, cards entering and leaving the hand, an ETB pulse, damage popups and combat strikes.

Ian (owner) asked on 2026-10-27 for three things:

1. Finer control over the sounds.
2. Cues that tell a player what happened, before they read the log. These include "it's your turn" and "you have priority", a board wipe that reads as one event across the whole board, and an exile that never looks like a death.
3. Later, effects of their own for named cards. Examples: Blasphemous Act rains fire on the battlefield, Toxic Deluge bubbles up like a tar pit, Bitter Triumph drives a spear through the card, Dragon's Approach passes a dragon's shadow over the board, and Nicol Bolas swoops around the table before he lands. Planeswalkers may also speak a short line when they enter or leave.

The server already gives the client what it needs to group events. Every event carries `Batch`, the run of events the rules treat as one occurrence (CR 603.2c), and `Source`, the card whose effect produced it (`server/internal/game/events.go`). A board wipe is therefore one `Batch` of `zone_move` events, from the battlefield to the graveyard, that share one `Source`.

## Decision

### 1. Effects never slow the table

- **The animation plays over a state that has already changed.** The game state updates when the event arrives, and the animation runs on top. Nothing waits for it, including input, priority passes and bot seats.
- **Any player can skip** an effect that is playing, with a click or Escape, without affecting anyone else.
- **Length caps:** a generic cue runs at most 1.5 s and a signature effect at most 2.5 s. Both scale with `animations.speed`.
- **A setting,** `animations.effects`: Full, Short or Off. Short plays only the first beat of a signature, up to 0.8 s. Off uses the generic cues in §3 for every card.
- **Reduced motion.** With reduced motion on, or `animations.enabled` off, a signature or board effect becomes a brief tint and the generic cue becomes the existing outcome badge. The sound still plays unless its category is muted. This follows DESIGN.md §4 and ADR 0119 §3.

### 2. Sound categories

`audio` gains one volume (0–100) per category. The master volume and mute stay on top of them.

| Category | Plays |
|---|---|
| `interface` | Your turn, you have priority, a prompt is waiting for you, a timer warning |
| `cards` | Draw, play, tap, shuffle, counters, tokens |
| `combat` | Attack, block, combat damage |
| `effects` | The generic cues in §3 and every signature effect in §4 |
| `voice` | Planeswalker lines (§5) |
| `music` | The existing music bus |

The existing `effectsVolume` becomes the `cards` volume during the settings migration, so a returning player hears no change.

Bursts are capped. One logical sound plays at most once per `Batch`, so twelve creatures dying make one board-wipe sound. A sound in the `interface` category ducks the others by about 6 dB while it plays.

### 3. Generic effect cues (every card gets them)

| Cue | Detected from | Look and sound (default) |
|---|---|---|
| Destroy | `zone_move` battlefield→graveyard caused by a destroy | A dark slash on the card; a raven's caw |
| Dies (SBA, damage) | `zone_move` battlefield→graveyard with no destroy | The existing crumble |
| Exile | `zone_move` to exile | A portal-like fade toward a point, cold light; a distinct chime. It must never be confused with a death. |
| Sacrifice | `sacrifice` | A downward pull; a low thud |
| Bounce | `zone_move` battlefield→hand | A lift and fly-back to the owner's seat |
| Counterspell | `counter_spell` | The stack item shatters |
| Board wipe | 3 or more permanents leave the battlefield in one `Batch` from one `Source` | One sweep across the whole board, then each card's per-card cue at reduced weight; a low rumble |
| Planeswalker enters / leaves | `etb` / `zone_move` of a planeswalker | A short sting and glow / a crack |

Some events may not say whether a move was a destroy, an SBA or a sacrifice. If so, the server adds an additive `cause` field to the `zone_move` event (for example `destroy`, `sba`, `sacrifice`, `exile_effect`). It does not add a new event kind. This is the only engine change in this ADR.

Amendment 2026-10-27 (#4395): the engine did not say. A destroy, a lethal-damage death, a sacrifice, an exile and a bounce all emitted the same `zone_move`, and the wire log dropped `Batch` on purpose, on the argument that adjacency was enough. Now every battlefield exit carries `leave_cause` (`destroy`, `sba`, `sacrifice`, `exile_effect`, `bounce`, `other`) and `leave_source` on the engine event, set at the two movers (`executeBattlefieldLeaveLocked`, `executeZoneRouteLocked`) from the route the verb chose, and the public log forwards `batch`, `cause` and `source`. The cause is the instruction, not the destination: read `new_zone` for the exile look. Lethal damage and deathtouch are `sba`, although CR 701.8b calls them a destruction, because the cue for them is the death one. A board wipe is three or more `zone` entries with one `batch` and one `source`; a sacrifice wipe is the same over `sacrifice` entries. Wire format: [protocol.md](../protocol.md), "Batch, cause and source".

Amendment 2026-10-28 (#4396): the generic cues are built (`client/src/lib/tableEffects.ts`, `components/board/TableEffects.svelte`, z 57). What the build settled:

- **The setting.** `animations.effects` (Full, Short, Off; settings version 30, default Full) governs everything this layer animates, the generic cues included, not only signatures (§4). Short cuts each cue to at most 0.8 s, and a board wipe is compressed so the sweep and its last per-card cue fit it. Off, reduced motion (`accessibility.reduceMotion`, the OS signal) and `animations.enabled` off are one state, "still": no motion, a badge where the permanent was, held 1.2 s whatever the speed. A board wipe is one badge and a static tint. The sound plays in every mode. This reads §1's "Off uses the generic cues in §3 for every card" as "Off shows each cue in its still form": Off is the one switch for table-effect motion. If the owner would rather Off keep the generic motion and drop only signatures, the gate is one function (`effectMotion`).
- **Looks.** Destroy: a dark slash along the card's diagonal, the card darkening and sinking. Exile: the card drawn into a cold portal, shrinking and turning to cold light. Death by state-based action: the existing crumble (`animations.crumble`, now taking the effects gate). Sacrifice: the card pulled down into a dark well. Bounce: a lift, then a flight to the owner's seat. Counterspell: cracks and a flash over the stack item's last box, glass flying out; the stack linger's "Countered" badge stays the badge. Planeswalker: a ring and glow on entering, a crack and a shake on dying (an exiled or bounced one keeps the exile or bounce look). Board wipe: one sweep across the table, tinted dark, cold (mostly exile) or cool (mostly bounce), then each card's cue lighter, staggered from the sweep.
- **Combat.** A death by state-based action in a combat damage step is not drawn here while the strike layer is playing (it crumbles those creatures, ADR 0134); with combat motion off this layer draws it.
- **Skip.** A click anywhere or Escape clears the cues playing, for the local player only. The layer takes no pointer events, so the click also reaches the board.
- **Sound.** One logical sound per `batch` and game, as §2 says. A wipe plays `board_wipe` once and each per-card sound once at under half volume.

### 4. Signature effects for named cards

`client/src/lib/fx/signatures.ts` maps a card's `oracle_id` to a signature: a short scene plus its sound, in the `effects` category. The scene runs when that card's spell resolves, or when its ability produces the batch the scene names. A card with no signature uses the generic cues in §3.

A signature is client data and code. It does not change rules, the snapshot or the wire. Adding a card's signature is a small job with no engine work, tiered `tier:1-mechanical` once the scene kit exists.

The scene kit is a small set of reusable parts: a full-board overlay layer, particle rain (fire, ash, droplets), a rising fill (tar, water), a shadow pass, a strike through a card, and a "fly in, then land as the card". Signatures are built from these parts. The overlay goes on the table's z-index ladder between the action dock (55) and hints (58), as layer 57.

The first batch is the cards Ian named: Blasphemous Act, Toxic Deluge, Bitter Triumph, Dragon's Approach and Nicol Bolas (each printing named Nicol Bolas that is in the catalog, sharing one scene).

### 5. Voice lines (wired, not shipped)

`client/src/lib/fx/voices.ts` maps a planeswalker's `oracle_id` to lines for each moment: `enter`, `leave`, `ultimate`. Each moment has one or more takes, with the files under `client/public/voices/`. Lines play in the `voice` category.

- Each walker has a cooldown: a walker speaks at most once every 20 seconds.
- Only one voice line plays at a time across the whole table.

The wiring ships with an empty manifest. Whether lines ship at all waits on Luke's OK. Lines are original, short and in the C2 voice: TTS voices made for C2, not a recording or clone of any actor's voice from Wizards media, and no quoted card text. Adding a line is then a manifest entry plus a file.

### 6. Assets

Ian produces the sound files and voice files himself. The asset spec lives in `client/public/sounds/README.md`:

- **File names:** the manifest key plus a take number (for example `destroy-1.mp3`).
- **Format:** MP3, 44.1 kHz.
- **Loudness:** about −16 LUFS integrated, peaks under −1 dBTP.
- **Length:** a generic cue at most 1.5 s, a signature at most 2.5 s, a voice line at most 3 s.

Until a file exists, its manifest entry falls back to the closest existing sound, or stays silent. A missing file is never an error. Every art and sound asset is original or CC0. None comes from Wizards (PLAN.md §3, the Fan Content Policy).

## Consequences

- Layers 1 and 2 (§2, §3) give every card clearer feedback without per-card work. Layer 3 (§4) is open-ended, but each card in it is small and isolated.
- A board wipe is one visual event, so a 4-player table no longer gets a wall of overlapping crumbles.
- The voice manifest can stay empty indefinitely at no cost.
- The client bundle grows by the scene kit, and assets load lazily per signature. Signatures are fetched the first time their card appears in a game, not when the app starts.
