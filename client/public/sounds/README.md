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
| Your turn begins          | `your_turn-1.mp3`, `your_turn-2.mp3`           |
| You have priority         | `priority-1.mp3`, `priority-2.mp3`             |
| A prompt is waiting       | `prompt-1.mp3`, `prompt-2.mp3`                 |
| Timer warning             | `timer_warning-1.mp3`, `timer_warning-2.mp3`   |

### Files not made yet

These four manifest entries ship before their files do (ADR 0209 §6).
Drop the files in this folder under exactly these names and they play;
until then a missing file is never an error:

| Entry           | Files                                        | Until the files exist                |
| --------------- | -------------------------------------------- | ------------------------------------ |
| `your_turn`     | `your_turn-1.mp3`, `your_turn-2.mp3`         | plays `turn_change` on your own turn |
| `priority`      | `priority-1.mp3`, `priority-2.mp3`           | silent                               |
| `prompt`        | `prompt-1.mp3`, `prompt-2.mp3`               | silent                               |
| `timer_warning` | `timer_warning-1.mp3`, `timer_warning-2.mp3` | silent (no turn clock exists yet)    |

The first time the browser fails to load one it remembers that for the
session and stops asking. One take is enough: with `-1` present and `-2`
absent only `-1` plays.

## Categories and ducking

Every `SoundName` belongs to one category (`SOUND_CATEGORY` in
[../src/lib/sounds.ts](../src/lib/sounds.ts)); Settings has a volume for
each, and `play()` multiplies master x category x the caller's hint.

| Category    | Sounds                                                                           |
| ----------- | -------------------------------------------------------------------------------- |
| `interface` | `turn_change`, `win`, `loss`, `your_turn`, `priority`, `prompt`, `timer_warning` |
| `cards`     | `draw`, `play`, `tap`, `untap_all`, `shuffle`                                    |
| `combat`    | `attack`, `block`, `combat_resolve`, `damage`                                    |
| `effects`   | `heal`; the generic effect cues and signature effects to come                    |
| `voice`     | planeswalker lines to come                                                       |

An `interface` sound ducks the other buses, music included, by about 6 dB
while it plays. A sound passed an event `batch` plays once per batch.

## Asset spec (ADR 0209 §6)

Ian makes the files. Every file is original or CC0; none comes from
Wizards.

- **Names:** the manifest key plus a take number: `destroy-1.mp3`,
  `your_turn-2.mp3`.
- **Format:** MP3, 44.1 kHz.
- **Loudness:** about -16 LUFS integrated, true peak under -1 dBTP.
- **Length:** a generic cue at most 1.5 s, a signature effect at most
  2.5 s, a voice line at most 3 s. Interface cues are generic cues: keep
  them short, because they duck everything else.

## Ambient music

- `ambient_bronze_thunder.mp3` — looping ambient bed for the app
  shell. Played by [../src/lib/music.ts](../src/lib/music.ts),
  armed on first user gesture from `App.svelte`, and gated by the
  `audio.masterVolume × audio.musicVolume` product from the
  Settings panel. Not part of the `SoundName` enum — SFX and
  music use separate playback paths.

## Note on the source dump

The source asset pack named its damage cues `damage popup 1.mp3`
and `damage output 2.mp3` (no matching `damage popup 2` or
`damage output 1`). Both are mapped here as generic damage variants
(`damage-1.mp3`, `damage-2.mp3`). Revisit if a take feels out of
character for the event.
