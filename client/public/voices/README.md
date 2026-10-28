# Voice lines

Planeswalker lines for ADR 0209 §5. The wiring is in `client/src/lib/fx/`; this
directory holds the files. **It ships empty.** Shipping actual lines waits on
Luke's OK, so do not add one to a PR until he has given it.

## Add a line

1. Put the MP3 here, named `<walker>-<moment>-<take>.mp3`
   (for example `example-enter-1.mp3`). `<moment>` is `enter`, `leave` or
   `ultimate`.
2. Add one entry to `VOICE_MANIFEST` in `client/src/lib/fx/voices.ts`, keyed by
   the card's Scryfall `oracle_id` (never the `scryfall_id`):

   ```ts
   "<oracle_id>": {
     enter: ["/voices/example-enter-1.mp3", "/voices/example-enter-2.mp3"],
     ultimate: ["/voices/example-ultimate-1.mp3"],
   },
   ```

   A moment you leave out is silent. More than one take is picked at random.

That is all: the game log drives the rest. A file that is not there is silent,
never an error.

## When a line plays

- `enter`: a planeswalker enters the battlefield. `leave`: it leaves it.
- `ultimate`: the walker's most expensive minus loyalty ability is activated
  (strictly the largest fixed minus cost; a -X ability or a tie has none).
- A walker speaks at most once every 20 seconds, and only one line plays at a
  time across the table. Lines use the `voice` volume and the mute toggle.

## Rules for the files

- Original TTS voices made for C2. Never a recording or a clone of a real
  actor's voice, and nothing taken from Wizards media.
- Never quote card text. Short, original lines in the C2 voice.
- At most 3 seconds.
- MP3, 44.1 kHz, about -16 LUFS integrated, peaks under -1 dBTP.

The general asset spec is in `client/public/sounds/README.md`.
