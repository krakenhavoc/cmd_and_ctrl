# Adding a signature effect

A signature is a short scene a named card plays over the table in place of
its generic effect cue ([ADR 0209](decisions/0209-table-effects-sound-and-signature-animations.md) §4,
#4397). It is client work only: no engine, wire or snapshot change. Once the
scene kit exists, a new signature built from its parts is `tier:1-mechanical`.

## What exists

| File | What it holds |
|---|---|
| `client/src/lib/fx/signatures.ts` | `SIGNATURES` (oracle_id to scene), `SCENES` (each scene's trigger, sound, length, first beat and loader), and every rule: matching, caps, Short, still, lazy loading |
| `client/src/lib/fx/scenes/kit.ts` | The parts: `overlay`, `canvasPart` (a particle canvas), `rain` (fire, ash), `fill` (tar, water), `ooze` (sludge welling up across the board), `sink` (a card going down into it), `shadowPass`, `strike`, `flyLand`, `breath` (fire from the flying dragon onto a seat) |
| `client/src/lib/fx/tints.ts` | The tints a scene leaves on its survivors until the turn ends, by instance ID; `components/board/FxTint.svelte` draws one on a tile |
| `client/src/lib/fx/scenes/<scene>.ts` | One scene each, exporting `play(ctx)` |
| `client/src/lib/components/board/TableEffects.svelte` | Draws the scene at z 57 and folds it into the generic cues |
| `client/src/lib/sounds.ts` | The `sig_<scene>` sound, in the `effects` category |

The scenes so far: `blasphemous_act`, `toxic_deluge`, `bitter_triumph`,
`dragons_approach`, and `nicol_bolas` (shared by every catalog printing named
Nicol Bolas).

## Recipe

1. **Check the card is in the catalog.** Find its Spec in
   `server/internal/cards/effects/<snake_name>.go` and copy `OracleID` and
   `Name` from it. Never guess an oracle_id, and leave out a card that is not
   in the catalog.
2. **Pick a trigger.**
   - `resolve`: the card's spell resolved. The scene gets every permanent that
     left the battlefield with the spell as its `source`, where each stood. It
     replaces that batch's board-wipe sweep; each card still shows its own cue,
     lighter, from `cardsAtMs`. Set `needsCards: true` when the scene draws on
     a card and has nothing to show without one (a spear needs a target).
     Set `ownsCards: true` when the scene draws the cards' exits itself (Toxic
     Deluge sinks them): their generic cues are then left out while it plays.
     Set `tint` when the creatures the spell did not take should keep a mark
     until the turn ends (Toxic Deluge's `deluge`; add a kind to `TintKind`
     and its look to `FxTint.svelte`).
   - `enter`: the card entered the battlefield, its spell resolving or not.
     It lands on its tile, in place of the planeswalker ring, and gets the
     players' seats (`ctx.seats`, each marked a foe or not). Casting the card
     plays nothing.
3. **Add the scene's spec** to `SCENES` in `signatures.ts`: its `sound`
   (`sig_<scene>`), `ms` (at most 3500), `firstBeatMs` (what Short plays, at
   most 800), `cardsAtMs`, and `load: () => import("./scenes/<scene>")`.
   Add the scene's ID to the `SceneID` union.
4. **Map the card** in `SIGNATURES`: `"<oracle_id>": { names: ["<Name>"], scene: "<scene>" }`.
   List each face's name for a card with two. Several cards may share one scene.
5. **Write the scene**, `client/src/lib/fx/scenes/<scene>.ts`, exporting
   `play(ctx: SceneContext)`. Build it from the kit's parts, with times in the
   scene's own milliseconds at speed 1 (the kit applies the cap, Short and
   `animations.speed`). When `ctx.short` is true, draw the first beat only,
   ending by `firstBeatMs`. Nothing may end after `ms`. Only import from
   `./kit` and from modules the board already loads, so the scene stays its
   own lazy chunk (vite.config.ts names it `assets/sig-*`, and the service
   worker leaves it out of the precache).
6. **Add the sound.** In `client/src/lib/sounds.ts`, add `sig_<scene>` to
   `SoundName`, `SOUND_CATEGORY` (`effects`), `SOUND_MANIFEST` (two takes) and
   `SOUND_FALLBACK` (the generic cue it replaces, so it is never quieter than
   no signature). Add a row to `client/public/sounds/README.md`.
7. **Test.** `fx/signatures.test.ts` checks each oracle_id and name against the
   server's catalog and oracle fixtures, and every scene's caps;
   `fx/scenes/kit.test.ts` draws every scene and fails if any part runs past
   its length in Full or its first beat in Short. Add a `kit.test.ts` case for
   what is particular to the new scene, then run, from `client/`:
   `npx vitest run src/lib/fx src/lib/signatures.render.test.ts`,
   `npm run check` and `npm run lint`.

## Rules

- Every shape is drawn for C2 (CSS, inline SVG, canvas). No Wizards art, and
  no card art beyond the card's own image, which the board already shows.
- A signature never delays the game: the state has already changed under it,
  a click or Escape skips it, and a scene that has not loaded yet leaves the
  generic cues to play.
- Off, reduced motion (the setting, `animations.enabled` off or the OS signal)
  play no scene: the generic still cues show, with the signature's sound.
- Keep it light: a few dozen animated elements, transforms and opacity, and at
  most one blur per scene. Many small things (fire, sludge, sparks) go on a
  `canvasPart`, planned up front from the seed and painted from scratch each
  frame, rather than as DOM nodes.
- Look at it in a browser, in Full and in Short, before the PR: the tests
  hold the timing, not the look.
