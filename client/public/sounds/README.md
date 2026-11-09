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
fallback until the files land (ADR 0209 §6), or nothing when it has none.
`sounds.test.ts` fails if a manifest take has no file, and if an entry with
no takes is missing from its `NOT_MADE_YET` list. In the PR that adds the
files, list the takes in the manifest and take the entry out of
`NOT_MADE_YET`.

| Entry              | Files                                                               | Spec                                                                                                                                                                                                                                                                                                                                                                                                                    | Falls back to |
| ------------------ | ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| `extort`           | `extort-1.mp3`                                                      | About 2 s. A soul sucked out (~1.2 s rising whoosh), then a coin clinking onto a brass scale, with the clink at 1.2-1.4 s. Plays when an extort drain resolves (#4544).                                                                                                                                                                                                                                                 | `damage`      |
| `token_create`     | `token_create-1.mp3` (and an optional `token_create-2.mp3`)         | At most 0.6 s. A soft poof or shimmer. Plays once when one or more tokens are created together, as the first of them pops in with a soft glow (#4441): ten tokens made at once are one sound.                                                                                                                                                                                                                           | none (silent) |
| `counter_add`      | `counter_add-1.mp3` (and an optional `counter_add-2.mp3`)           | At most 0.4 s. A short rising tick. Plays once when +1/+1 counters are put on one or more permanents together, as their power/toughness pulses green (#4442): a proliferate is one sound.                                                                                                                                                                                                                               | none (silent) |
| `counter_remove`   | `counter_remove-1.mp3` (and an optional `counter_remove-2.mp3`)     | At most 0.4 s. A short falling tick. Plays once when -1/-1 counters are put on one or more permanents together, as their power/toughness pulses purple (#4442). A batch with both kinds plays the first one's.                                                                                                                                                                                                          | none (silent) |
| `commander_cast`   | `commander_cast-1.mp3` (and an optional `commander_cast-2.mp3`)     | At most 1.2 s. A horn or choir swell. Plays once when a commander is cast from the command zone, as golden light rises from where it sat to the stack (#4443).                                                                                                                                                                                                                                                          | none (silent) |
| `commander_return` | `commander_return-1.mp3` (and an optional `commander_return-2.mp3`) | At most 0.8 s. A soft whoosh home. Plays once when a commander is put back into the command zone (from its graveyard or exile, or instead of its hand or library), as it drifts home with a soft trail (#4443).                                                                                                                                                                                                         | none (silent) |
| `attach_equip`     | `attach_equip-1.mp3` (and an optional `attach_equip-2.mp3`)         | At most 0.6 s. A metal clang or buckle. Plays once when an Equipment (or a Fortification) becomes attached to a permanent, as it slides onto its host and clasps with a spark (#4444).                                                                                                                                                                                                                                  | none (silent) |
| `attach_aura`      | `attach_aura-1.mp3` (and an optional `attach_aura-2.mp3`)           | At most 0.6 s. A soft shimmer. Plays once when an Aura enchants a permanent (resolving onto it, entering attached, or a Role or a token copy created on it), as it slides onto its host and a soft light rings it (#4444).                                                                                                                                                                                              | none (silent) |
| `mill`             | `mill-1.mp3` (and an optional `mill-2.mp3`)                         | At most 0.8 s. A papery card slide. Plays once when cards are milled into a graveyard (or otherwise put there from a library, a surveil's included), as they slide from the library into the graveyard pile one after another, or as one sweep for five or more (#4445): each player milling at once is one sound.                                                                                                      | none (silent) |
| `discard`          | `discard-1.mp3` (and an optional `discard-2.mp3`)                   | At most 0.5 s. A soft card toss. Plays once when cards are discarded into a graveyard (an effect's discard, the hand-size discard, a cost, a cycling), as they are tossed from the hand onto the graveyard pile (#4445): a whole hand discarded at once is one sound.                                                                                                                                                   | none (silent) |
| `reanimate`        | `reanimate-1.mp3` (and an optional `reanimate-2.mp3`)               | At most 1 s. A reversed whoosh with a low rumble. Plays once when cards return from a graveyard to the battlefield (Reanimate, Living Death, unearth, a land put back) or to a hand (Regrowth, at 60% volume), as each rises out of the graveyard pile in a dark-green glow and goes to where it went (#4446): one effect's cards are one sound.                                                                        | none (silent) |
| `library_search`   | `library_search-1.mp3` (and an optional `library_search-2.mp3`)     | At most 0.6 s. A quick card riffle (a short cut of `shuffle` would do). Plays once when a player searches a library (a tutor, a fetch land's effect, a search of another player's library, a search that finds nothing), as the top cards of that library's pile riffle up and settle (#4447): one effect's searches are one sound. The search's shuffle plays nothing of its own.                                      | none (silent) |
| `transform`        | `transform-1.mp3` (and an optional `transform-2.mp3`)               | At most 0.8 s. A flip whoosh with a short sting. Plays once when a double-faced permanent turns over to its other face (a transform, a convert, two cards melded into one), as its card flips on its vertical axis to the other face (#4448): Moonmist turning every Human over is one sound. Not for a face-down permanent turned face up, a disturb cast or a card returned transformed.                              | none (silent) |
| `day_night`        | `day_night-1.mp3` (and an optional `day_night-2.mp3`)               | At most 0.8 s. A bell: low and soft for night, brighter for day. Plays once when the game becomes day or night, as a sky tint washes across the table (#4448). The werewolves it turns over flip a beat later and play nothing of their own: night falling is one sound.                                                                                                                                                | none (silent) |
| `fetch_crack`      | `fetch_crack-1.mp3` (and an optional `fetch_crack-2.mp3`)           | At most 0.7 s, a little quieter than the other effect cues. A dry branch or wax-seal crack, then a short card-slide or map-rustle "search". Plays once when a land is sacrificed as a cost of its own ability (a fetch land, Evolving Wilds, Prismatic Vista), as the land splits along a crack and falls away, with no altar (#4427). The search that follows, when the ability resolves, plays `library_search`.      | none (silent) |
| `treasure_crack`   | `treasure_crack-1.mp3` (and an optional `treasure_crack-2.mp3`)     | At most 0.5 s, a little quieter than the other effect cues. A coin clink. Plays once when a Treasure is sacrificed for its own mana, by hand or by the auto-tapper paying for a spell, as it splits and falls away (#4427): two Treasures spent on one spell are one sound. Any other permanent sacrificed for its own ability that is not a land, Clue or Food (a Blood, a Map, Lotus Petal, Mind Stone) plays it too. | none (silent) |
| `clue_crack`       | `clue_crack-1.mp3` (and an optional `clue_crack-2.mp3`)             | At most 0.5 s, a little quieter than the other effect cues. A paper crumple. Plays once when a Clue is sacrificed for its own ability, as it splits and falls away (#4427).                                                                                                                                                                                                                                             | none (silent) |
| `food_crack`       | `food_crack-1.mp3` (and an optional `food_crack-2.mp3`)             | At most 0.5 s, a little quieter than the other effect cues. A crunch or a bite. Plays once when a Food is sacrificed for its own ability, as it splits and falls away (#4427).                                                                                                                                                                                                                                          | none (silent) |
| `fizzle`           | `fizzle-1.mp3` (and an optional `fizzle-2.mp3`)                     | At most 0.8 s. A quiet version of `counter` without the slap. Plays once when a spell or ability leaves the stack with no legal target (CR 608.2b), as it dissolves where it sat (#4450).                                                                                                                                                                                                                               | none (silent) |
| `drain_table`      | `drain_table-1.mp3` (and an optional `drain_table-2.mp3`)           | At most 1 s. A single low siphon. Plays once when one source makes two or more players lose life at once (Syr Konrad, Zulaport Cutthroat), as a thin dark line runs from it to each of them (#4450).                                                                                                                                                                                                                    | none (silent) |
| `monarch`          | `monarch-1.mp3` (and an optional `monarch-2.mp3`)                   | At most 1 s. A bright crown chime. Plays once when a player becomes the monarch, with a short glint on their panel (#4450).                                                                                                                                                                                                                                                                                             | none (silent) |
| `initiative`       | `initiative-1.mp3` (and an optional `initiative-2.mp3`)             | At most 1 s. A darker chime than `monarch`. Plays once when a player takes the initiative, with a short violet glint on their panel (#4450).                                                                                                                                                                                                                                                                            | none (silent) |
| `dice_roll`        | `dice_roll-1.mp3` (and an optional `dice_roll-2.mp3`)               | At most 1 s. A dice rattle. Plays once for each roll as its dice start to tumble, however many dice it rolls: a card's roll, a table roll or the opening roll (#4450).                                                                                                                                                                                                                                                  | none (silent) |
| `coin_flip`        | `coin_flip-1.mp3` (and an optional `coin_flip-2.mp3`)               | At most 1 s. A coin ring. Plays once for each flip as its coins start to spin, however many coins it flips (#4450).                                                                                                                                                                                                                                                                                                     | none (silent) |

### Hits on a player and elimination (#4543, #4449)

Combat damage that lands on a player plays one of three sounds in the
`combat` category, by how much that player takes in the beat: **light** for
exactly 1, **medium** for 2 to 3 and **heavy** for 4 or more. Each is one
take, and each falls back to `damage` if its file fails to load
(`SOUND_FALLBACK`). A player leaving the game plays `player_eliminated`, one
take, in the `effects` category, with no fallback. A player who loses by
drawing from an empty library (CR 704.5b, cause `empty_draw`) plays
`player_decked` instead (#4879), two takes in the `effects` category, falling
back to `player_eliminated`.

| Entry               | Files                     | When it plays                                                                                                                          | Falls back to       | Source                                           |
| ------------------- | ------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ------------------- | ------------------------------------------------ |
| `face_hit_light`    | `face_hit_light-1.mp3`    | 1 combat damage hits a player: a deflection, a dink                                                                                    | `damage`            | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `face_hit_medium`   | `face_hit_medium-1.mp3`   | 2 or 3 combat damage hits a player: a blunt punch, no gore                                                                             | `damage`            | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `face_hit_heavy`    | `face_hit_heavy-1.mp3`    | 4 or more combat damage hits a player: a bone and blood crush (the skull crush)                                                        | `damage`            | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `player_eliminated` | `player_eliminated-1.mp3` | A player leaves the game, once for each: their avatar shatters like rock, a tombstone stays                                            | none                | ElevenLabs (Ian's paid subscription), 2026-10-29 |
| `player_decked`     | `player_decked-1.mp3`     | A player is decked (drew from an empty library): a falling, wavering moan, then a flutter of cards at about 0.7 s; the library's ghost | `player_eliminated` | Synthesised                                      |
| `player_decked`     | `player_decked-2.mp3`     | The second take of the same                                                                                                            | `player_eliminated` | Synthesised                                      |

A face hit plays at the strike's contact, once per player a combat damage
beat hits, and the tier is picked by all the combat damage that player takes
in that beat (two attackers of 2 are one blow of 4). The life popup then does
not also play `damage`. With combat motion off, reduced motion or a hidden
tab it plays when the life loss shows. Commander damage is combat damage, so
it gets a tier like any other. Other life loss (a drain, life paid,
noncombat damage) keeps `damage`.

The `combat_resolve` war cry (5 to 7 s) would bury a face hit, so on a beat
that hits a player it starts 300 ms after the last (staggered) face hit, at
0.6 of its volume (`WAR_CRY_AFTER_HIT_MS`, `WAR_CRY_AFTER_HIT_VOLUME` in
`faceHit.ts`, #4560). A beat that hits no player plays it at contact, at full
volume. The step-change fallback (motion off, reduced motion, a hidden tab)
keeps the same order.

`extort` plays once when an extort trigger resolves and drains, as the Scales
of Souls effect starts (`ExtortScales.svelte`, #4544). The life popups on the
drained opponents and the extort player stay quiet for that drain, so this is
its only sound. Brief for the file: soul sucked out (~1.2 s rising whoosh) then
a coin clinking onto a brass scale. **The coin hit must land 1.2-1.4 s into the
file**: the animation's scale levels (the coin falls and the light bursts) at
exactly 1.25 s after the sound starts (`EXTORT_HIT_MS` in
[../../src/lib/extort.ts](../../src/lib/extort.ts)), at animations.speed 1. A
short tail after the clink is fine; the animation runs 2.6 s in all. In the
reduced-motion and Short settings the scene is a still scale and the sound
still plays. Until the file is added it plays `damage`. To add it: put
`extort-1.mp3` here, list `"/sounds/extort-1.mp3"` in `SOUND_MANIFEST.extort`
(`src/lib/sounds.ts`), and remove the row above.

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

| Category    | Sounds                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `interface` | `turn_change`, `win`, `loss`, `your_turn`, `priority`, `prompt`, `timer_warning`                                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| `cards`     | `draw`, `play`, `tap`, `untap_all`, `shuffle`                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| `combat`    | `attack`, `block`, `combat_resolve`, `damage`, `face_hit_light`, `face_hit_medium`, `face_hit_heavy`                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| `effects`   | `heal`, `player_eliminated`, `player_decked`, `destroy`, `exile`, `sacrifice`, `sacrifice_knife`, `bounce`, `counter`, `board_wipe`, `pw_enter`, `pw_leave`, `token_create`, `counter_add`, `counter_remove`, `commander_cast`, `commander_return`, `attach_equip`, `attach_aura`, `mill`, `discard`, `reanimate`, `library_search`, `transform`, `day_night`, `fetch_crack`, `treasure_crack`, `clue_crack`, `food_crack`, `extort`, `sig_blasphemous_act`, `sig_toxic_deluge`, `sig_bitter_triumph`, `sig_nicol_bolas`, `sig_dragons_approach` |
| `voice`     | planeswalker lines to come                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                       |

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

## Length caps

`combat_resolve` (the war cry; takes run 5 and 7 s) is capped at 2 s with a
300 ms fade-out, in code, so its tail doesn't cover the sounds after combat.
The files are untouched. The caps live in `SOUND_MAX_MS` and `SOUND_FADE_MS`
in `client/src/lib/sounds.ts` (#4567).

## Files waiting to be wired

Delivered by Ian on 2026-11-04 (ElevenLabs Sound Effects, his Creator plan, which allows commercial use; normalised to −16 LUFS, peak −1 dBTP). They are not in `SoundName` yet. The cue that plays them is #4427, which is claimed on `claude/brave-galileo-t0um7r-stacked`.

| Sound         | Files                                         | When it plays                            |
| ------------- | --------------------------------------------- | ---------------------------------------- |
| `food_enters` | `food_enters-1.mp3`, `-2`, `-3` (about 1.2 s) | A Food token enters                      |
| `food_sac`    | `food_sac-1.mp3`, `-2`, `-3` (about 1.2 s)    | A Food is sacrificed for its own ability |
