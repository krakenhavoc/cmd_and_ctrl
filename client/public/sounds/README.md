# Sound assets

Audio cues for in-game events (S09 sound pass, ADR 0209). Most events
ship two takes so `play()` can pick a variant at random and avoid the
"machine-gun loop" effect on rapid-fire events (e.g. drawing an
opening hand of seven).

## Source / license

Two sets, by when they were made:

- **The S09 table sounds** (`draw` through `loss` in the first table
  below) and `ambient_bronze_thunder.mp3`: Suno-generated for this
  project. Personal / hobby use only, not licensed for redistribution
  outside this repo.
- **The ADR 0209 cues** (interface, effects and signatures): a mix,
  recorded per file in the source column of the tables below.
  - _Synthesised_: original, synthesised in code for this game.
  - _ElevenLabs_: ElevenLabs Sound Effects from the owner's Creator plan,
    which allows commercial use.
  - _ElevenLabs edit_: an original edit layered from ElevenLabs takes and
    stems.
  - _TangoFlux_: one take, `pw_enter-2.mp3`, generated with TangoFlux. Its
    model licence is non-commercial. That is fine for C2, which is free and
    never charges (PLAN.md §3), but replace the take if that ever changes.
  - _ElevenLabs (Ian's paid subscription), 2026-10-29_: the four files of
    the hit and elimination cues (`face_hit_light-1.mp3`,
    `face_hit_medium-1.mp3`, `face_hit_heavy-1.mp3` and
    `player_eliminated-1.mp3`, #4543 and #4449), made by Ian on his paid
    ElevenLabs plan, which allows commercial use.

None comes from Wizards.

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

A manifest entry lists exactly the takes that exist, except the entries
under "Files not made yet" below. One with a single take plays it every
time.

### Files not made yet

An entry here would be in `SOUND_MANIFEST` with no takes, so nothing is
fetched (a listed file that is absent logs a 404) and it would play its
fallback until the files land (ADR 0209 §6). `sounds.test.ts` fails if a
manifest take has no file. In the PR that adds the files, list the takes in
the manifest.

None at present.

### Hits on a player and elimination (#4543, #4449)

Combat damage that lands on a player plays one of three sounds in the
`combat` category, by how much that player takes in the beat: **light** for
exactly 1, **medium** for 2 to 3 and **heavy** for 4 or more. Each is one
take, and each falls back to `damage` if its file fails to load
(`SOUND_FALLBACK`). A player leaving the game plays `player_eliminated`, one
take, in the `effects` category, with no fallback.

| Entry               | Files                     | When it plays                                                                               | Falls back to | Source                                           |
| ------------------- | ------------------------- | ------------------------------------------------------------------------------------------- | ------------- | ------------------------------------------------ |
| `face_hit_light`    | `face_hit_light-1.mp3`    | 1 combat damage hits a player: a deflection, a dink                                         | `damage`      | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `face_hit_medium`   | `face_hit_medium-1.mp3`   | 2 or 3 combat damage hits a player: a blunt punch, no gore                                  | `damage`      | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `face_hit_heavy`    | `face_hit_heavy-1.mp3`    | 4 or more combat damage hits a player: a bone and blood crush (the skull crush)             | `damage`      | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `player_eliminated` | `player_eliminated-1.mp3` | A player leaves the game, once for each: their avatar shatters like rock, a tombstone stays | none          | ElevenLabs (Ian's paid subscription), 2026-10-29 |

A face hit plays at the strike's contact, once per player a combat damage
beat hits, and the tier is picked by all the combat damage that player takes
in that beat (two attackers of 2 are one blow of 4). The life popup then does
not also play `damage`. With combat motion off, reduced motion or a hidden
tab it plays when the life loss shows. Commander damage is combat damage, so
it gets a tier like any other. Other life loss (a drain, life paid,
noncombat damage) keeps `damage`.

### Interface cues (ADR 0209 §2)

| Entry           | Files                                        | When it plays                                            | Source      |
| --------------- | -------------------------------------------- | -------------------------------------------------------- | ----------- |
| `your_turn`     | `your_turn-1.mp3`                            | Your turn begins                                         | Synthesised |
| `priority`      | `priority-1.mp3`                             | You have a real decision to make (never on auto-pass)    | Synthesised |
| `prompt`        | `prompt-1.mp3`, `prompt-2.mp3`               | A prompt is waiting for you                              | Synthesised |
| `timer_warning` | `timer_warning-1.mp3`, `timer_warning-2.mp3` | The turn timer runs low (nothing plays it yet: no clock) | Synthesised |

### Generic effect cues (#4396)

The table's effect cues (ADR 0209 §3), in the `effects` category.

| Entry             | Files                                            | When it plays                                                                                   | Source                            |
| ----------------- | ------------------------------------------------ | ----------------------------------------------------------------------------------------------- | --------------------------------- |
| `destroy`         | `destroy-1.mp3`, `destroy-2.mp3`                 | A permanent is destroyed (the card shatters from its centre outward): a raven's caw, then blood | ElevenLabs edit                   |
| `exile`           | `exile-1.mp3`                                    | A permanent is exiled, including a destroy turned into exile (the white void)                   | ElevenLabs edit                   |
| `sacrifice`       | `sacrifice-1.mp3`, `sacrifice-2.mp3`             | A permanent is sacrificed: plays as the altar rises                                             | ElevenLabs                        |
| `sacrifice_knife` | `sacrifice_knife-1.mp3`, `sacrifice_knife-2.mp3` | The sacrifice's knife lands, about 0.5 s in, ending on the stone altar                          | -1 ElevenLabs edit, -2 ElevenLabs |
| `bounce`          | `bounce-1.mp3`, `bounce-2.mp3`                   | A permanent flies back to its owner's hand                                                      | ElevenLabs                        |
| `counter`         | `counter-1.mp3`, `counter-2.mp3`                 | A spell or ability is countered: a palm slap, a glass crack, a shatter                          | -1 ElevenLabs edit, -2 ElevenLabs |
| `board_wipe`      | `board_wipe-1.mp3`, `board_wipe-2.mp3`           | Three or more permanents leave in one batch from one source; it plays once, as the sweep runs   | ElevenLabs                        |
| `pw_enter`        | `pw_enter-1.mp3`, `pw_enter-2.mp3`               | A planeswalker enters the battlefield                                                           | -1 ElevenLabs, -2 TangoFlux       |
| `pw_leave`        | `pw_leave-1.mp3`                                 | A planeswalker dies (it cracks, then leaves through a rift)                                     | ElevenLabs                        |

A death by state-based action (lethal damage, zero toughness) has no new
sound: it is the existing crumble. Inside a board wipe each per-card sound
plays at less than half volume, and one sound plays once per batch.

### Signature effects (#4397)

A named card's own scene (ADR 0209 §4, `src/lib/fx/signatures.ts`) plays
one sound for its batch, in the `effects` category, in place of the generic
cue's. These are fetched the first time the card appears in a game, not at
the first gesture. If its files fail to load, a signature falls back to the
generic cue it replaces (`SOUND_FALLBACK`).

| Entry                  | Files                                                      | When it plays                                                       | Falls back to | Source                            |
| ---------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------- | ------------- | --------------------------------- |
| `sig_blasphemous_act`  | `sig_blasphemous_act-1.mp3`, `sig_blasphemous_act-2.mp3`   | Blasphemous Act resolves: fireballs crash down and the ground burns | `board_wipe`  | -1 ElevenLabs edit, -2 ElevenLabs |
| `sig_toxic_deluge`     | `sig_toxic_deluge-1.mp3`                                   | Toxic Deluge resolves: a tar pit wells up and takes the creatures   | `board_wipe`  | ElevenLabs                        |
| `sig_bitter_triumph`   | `sig_bitter_triumph-1.mp3`, `sig_bitter_triumph-2.mp3`     | Bitter Triumph resolves: a spear drives through the card            | `destroy`     | ElevenLabs edit                   |
| `sig_nicol_bolas`      | `sig_nicol_bolas-1.mp3`, `sig_nicol_bolas-2.mp3`           | Nicol Bolas enters: he swoops, breathes fire, lands                 | `pw_enter`    | -1 ElevenLabs edit, -2 ElevenLabs |
| `sig_dragons_approach` | `sig_dragons_approach-1.mp3`, `sig_dragons_approach-2.mp3` | Dragon's Approach resolves: a dragon's shadow passes over the board | `board_wipe`  | -1 ElevenLabs, -2 ElevenLabs edit |

A missing file is never an error (ADR 0209 §6): the first time the browser
fails to load one it remembers that for the session and plays the entry's
other takes, then its fallback, then nothing.

## Categories and ducking

Every `SoundName` belongs to one category (`SOUND_CATEGORY` in
[../src/lib/sounds.ts](../src/lib/sounds.ts)); Settings has a volume for
each, and `play()` multiplies master x category x the caller's hint.

| Category    | Sounds                                                                                                                                                                                                                                                 |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `interface` | `turn_change`, `win`, `loss`, `your_turn`, `priority`, `prompt`, `timer_warning`                                                                                                                                                                       |
| `cards`     | `draw`, `play`, `tap`, `untap_all`, `shuffle`                                                                                                                                                                                                          |
| `combat`    | `attack`, `block`, `combat_resolve`, `damage`, `face_hit_light`, `face_hit_medium`, `face_hit_heavy`                                                                                                                                                   |
| `effects`   | `heal`, `player_eliminated`, `destroy`, `exile`, `sacrifice`, `sacrifice_knife`, `bounce`, `counter`, `board_wipe`, `pw_enter`, `pw_leave`, `sig_blasphemous_act`, `sig_toxic_deluge`, `sig_bitter_triumph`, `sig_nicol_bolas`, `sig_dragons_approach` |
| `voice`     | planeswalker lines to come                                                                                                                                                                                                                             |

An `interface` sound ducks the other buses, music included, by about 6 dB
while it plays. A sound passed an event `batch` plays once per batch.

## Asset spec (ADR 0209 §6)

Ian makes the files, and the source of each new one goes in its table
above. None comes from Wizards.

- **Names:** the manifest key plus a take number: `destroy-1.mp3`,
  `your_turn-2.mp3`.
- **Format:** MP3, 44.1 kHz.
- **Loudness:** about -16 LUFS integrated, true peak under -1 dBTP.
- **Length:** no longer than its animation: a generic cue under 2 s
  (`src/lib/effectCueMotion.ts`), a signature effect at most 3.5 s
  (`SIGNATURE_CAP_MS`), a voice line at most 3 s. Interface cues are generic cues: keep
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
