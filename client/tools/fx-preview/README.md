# fx preview harness

A page that plays each table-effect cue ([ADR 0209](../../../docs/decisions/0209-table-effects-sound-and-signature-animations.md),
epic #4399) on a fake four-seat table, through the real layers the board mounts:
`TableEffects`, `EliminationFx`, `ExtortScales` and `DiceLayer`. Use it to look at a cue
while you build it, and to check it before review. It is a dev tool (#4868): it has its own
Vite config, is not a route of the app, and never ships in `npm run build`.

## Run it

```sh
cd client
npm run fx-preview          # http://localhost:5174
npm run fx-preview -- --port 5180
```

It needs no Go server. Card images are placeholders the harness's dev server draws: a
colour and the card's name, darker on a back face.

- A button plays one scenario. **Play all** plays them in turn.
- **Effects**, **Speed** and **Motion and sound** change the same settings the game reads,
  so Short, Off, speed and the three reduced-motion signals can be checked. They are stored
  for this page's origin only, apart from the app's.
- **Pause** holds every animation and the timers that start and clear a cue's beats
  (`clock.ts`). A click anywhere else or Escape skips the cue, as at the table.
- The URL says what to play on load: `#counters` plays that scenario, and `#counters@300`
  pauses it 300 ms after its frame lands, for a screenshot of one moment.

## Add a scenario

Every new cue gets one: a cue issue's "done when" includes "the harness has a scenario for
it". Add it to `SCENARIOS` in `scenarios.ts`:

1. Copy the scenario nearest to yours.
2. `before` is the table just before the moment and `after` the table the frame shows: a
   list of cards (`FxCard` in `table.ts`). A card is on seat 1's battlefield unless it says
   otherwise (`seat`, `zone`); `pt`, `counter`, `attachedTo`, `face`, `commander` and
   `token` draw the rest. Seat 0 is the viewer.
3. `log` is the entries the server sends for it. Copy their shape from the cue's render test
   (`src/lib/*.render.test.ts`); the harness stamps `seq` and `turn`.
4. List what it shows in `shows`. A cue drawn by a new layer also needs that layer mounted
   in `Stage.svelte`, and a name in `CueName`.

`npm test` runs `scenarios.test.ts`, which fails while an `EffectKind` has no scenario, a
scenario's log does not plan the cue it lists, or a log names a card its tables do not hold.
`npm run check` and `npm run lint` cover this folder too.

## What it does not show

The board's other motion (a spell's flight to the stack, the stack linger's badges, combat
strikes, keyword flourishes on a tile) and the signature scenes are not mounted. The fake
board draws only what the cues measure, so it does not look like the real one.
