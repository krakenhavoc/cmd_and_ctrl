# Sound assets

Audio cues for in-game events (S09 sound pass). Each event ships
two takes so `play()` can pick a variant at random and avoid the
"machine-gun loop" effect on rapid-fire events (e.g. drawing an
opening hand of seven).

## Source / license

Suno-generated for this project. Personal / hobby use only — not
licensed for redistribution outside this repo.

## File layout

All files are mono-ish MP3 served by Vite at `/sounds/<name>.mp3`.
The `SoundName` enum in [../src/lib/sounds.ts](../src/lib/sounds.ts)
owns the canonical list; add a new take by appending to the variant
array there.

| Event                     | Variants                                       |
| ------------------------- | ---------------------------------------------- |
| Card draw                 | `draw-1.mp3`, `draw-2.mp3`                     |
| Card play                 | `play-1.mp3`, `play-2.mp3`                     |
| Tap (rotate 90°)          | `tap-1.mp3`, `tap-2.mp3`                       |
| Untap all (start of turn) | `untap_all-1.mp3`, `untap_all-2.mp3`           |
| Damage / life loss        | `damage-1.mp3`, `damage-2.mp3`                 |
| Heal / life gain          | `heal-1.mp3`, `heal-2.mp3`                     |
| Attack declared           | `attack-1.mp3`, `attack-2.mp3`                 |
| Block declared            | `block-1.mp3`, `block-2.mp3`                   |
| Combat damage resolves    | `combat_resolve-1.mp3`, `combat_resolve-2.mp3` |
| Turn change               | `turn_change-1.mp3`, `turn_change-2.mp3`       |
| Mulligan / shuffle        | `shuffle-1.mp3`, `shuffle-2.mp3`               |
| Game over (win)           | `win-1.mp3`, `win-2.mp3`                       |
| Game over (loss)          | `loss-1.mp3`, `loss-2.mp3`                     |

## Bonus

- `ambient_bronze_thunder.mp3` — ambient loop from the original
  asset pack, not referenced by the `SoundName` enum yet. Left in
  place for a future ambient-music toggle (candidate for S11.5's
  music volume slider).

## Note on the source dump

The source asset pack named its damage cues `damage popup 1.mp3`
and `damage output 2.mp3` (no matching `damage popup 2` or
`damage output 1`). Both are mapped here as generic damage variants
(`damage-1.mp3`, `damage-2.mp3`). Revisit if a take feels out of
character for the event.
