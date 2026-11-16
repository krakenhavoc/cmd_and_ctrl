Sprint: S81 — Formats: Standard 1v1 and booster draft (epic #5431)

Ian's feedback after drafting on the live table, worked through as a clickable mockup with him before building. The mockup is a private artifact on Ian's account; ask him to share it if you want to click through it. Everything it settled is written out below, so this issue is the spec.

The pages concerned: `DraftRoom.svelte` (pick screen), `DraftBuild.svelte` (deckbuilding), `DraftSplit.svelte`, `DraftPool.svelte`, `DraftPack.svelte`, `DraftArchetypesPanel.svelte`, `DraftStats.svelte`.

## 1. Drag to Main or Sideboard

Today a pick can be dragged between Main and Sideboard on the pick screen, but only with a mouse (HTML5 drag and drop) and not from the pack.

- [ ] Drag a card from the **pack** onto Main or Sideboard to pick it straight into that pile. A plain click still picks into Main. (The split is client-side, `draftSplit.ts`, so picking to the sideboard is the pick plus `moveTo(id, "side")`.)
- [ ] Drag between Main and Sideboard on the pick screen, in Deck view (below) and on the deckbuilding screen.
- [ ] Use pointer events, not HTML5 drag and drop, so it works on touch. On touch: hold about 380 ms to lift the card (it also shows the preview), then drag. A swipe that moves before the hold is a scroll.
- [ ] A drop target highlights while a card is over it. A dragged card leaves a faded copy in place and a tilted copy follows the pointer.
- [ ] Keep the click and Enter fallback (`Move to sideboard: <name>`) and the screen-reader announcement.

## 2. Large preview on hover

- [ ] After a short delay, a large copy of the card appears **beside** the hovered card, never covering it: to the right, flipped left if it doesn't fit. On a phone it is centred above or below the card.
- [ ] Moving from card to card swaps the preview at once, with no second delay.
- [ ] Keyboard focus shows it, and so does a touch hold. Escape or scrolling hides it.
- [ ] Use the existing "Hover preview delay" setting (`Settings.svelte`) rather than a new one. The mockup defaulted to 250 ms with a 340 px preview, and Ian liked that.
- [ ] Hide it while dragging.

## 3. Why the card is rated what it is

Under the large preview there is a small box: **Rating 58.1 (B+)** and two to four short lines saying why. The lines are built from the numbers and the card's text, never written per card:

- [ ] Win rate when drawn (GIH WR) against the set's average. The set average isn't on the wire yet: compute it from the set's ratings on the server and send it with the draft view, or with the ratings.
- [ ] How early it goes, from `avg_seen`: "usually gone by pick N", or "often comes back around late". This line depends on #5726 (ALSA check); don't ship it until that is settled.
- [ ] What it does: removal, a flier, card draw, a counterspell, or a plain creature.
- [ ] Two colours: "strong, but only if you end up in both".
- [ ] A small footer: GIH WR, ALSA, mana value, and the source (17Lands, with its date; or the in-house rating).

## 4. Colour pairs panel collapses sideways

On a wide screen, collapsing the panel only hides its body; the 240 px column stays, so the pack gains nothing.

- [ ] Collapsed, it becomes a 44 px vertical strip with a rotated "Colour pairs" label and the expand button. The pack grid takes the width (about one more card per row at 1440 px).
- [ ] The strip shows one colour pip for each colour with three or more Main picks, so the hint survives while it's closed.
- [ ] On a narrow screen, collapsed is a single short horizontal bar, as now.
- [ ] Each pair shows how many of your Main picks fit it, highlighted at three or more.

## 5. Deck view during the draft

- [ ] A **Deck view** button in the bottom-right corner (with the pick count, keyboard `D`) swaps the pack for the deckbuilding layout: Main in columns by mana value (cards stacked so each name shows), Sideboard below. Drag, click and preview work there.
- [ ] A note at the top says the pick clock is still running. **Back to the pack** returns. Moves made in Deck view carry back.
- [ ] If the clock runs out in Deck view, the auto-pick happens as now and a toast names the card.

## 6. Colours and curve always shown; full stats on a screen of their own

Ian, after trying both: the curve and colours are needed all the time; the rest is wanted in full, but only when asked for.

- [ ] **Always shown** (Main only): **Colours** and **Curve**, compact. They sit under the picks on the pick screen, and in a strip across the top of Deck view and the deckbuilding screen. They replace the separate curve panel.
- [ ] A **Full stats** button beside Deck view (keyboard `S`) opens a full-screen overlay (`role="dialog"`, `aria-modal`). The page behind it doesn't scroll, so there is still one scroll bar. It has a sticky header with "Deck stats", the Main and Sideboard counts, the pick clock while drafting, and **Close** (`Esc`, or `S` again). Focus goes to Close on open and back to the button on close. It opens closed on page load; it isn't remembered. It holds Colours, Card types and a larger Curve with the count on each bar, then **More detail**, laid out in full (no dropdown):
- [ ] **Colours:** a bar split by coloured mana symbols in Main's costs, a row per colour (symbols and card count), and "17 basics by pips: 9 Swamp · 8 Plains" (largest-remainder split).
- [ ] **Card types:** creatures, instants, sorceries, enchantments, artifacts, planeswalkers and lands (a card with two types counts under both), plus "other spells". Below 13 creatures it says "Most limited decks want 15–17 creatures; you have N."
- [ ] **Curve:** the existing chart (`DraftStats`), plus the average mana value.
- [ ] **More detail** (in the overlay only). Every row lists the names of the cards it counts, and a zero row is shown faded so you can see what the deck is missing:
  - Type lines: Artifact Creature, Enchantment Creature, Enchantment (Aura), Artifact (Equipment) and so on.
  - Creatures: average power/toughness; small (power ≤ 2), medium (3–4), big (5+); evasive (flying, trample, menace, can't be blocked); the eight most common creature types.
  - What they do: removal, fliers, card draw, combat tricks, counterspells, bounce, token makers, mana and ramp, life gain, discard, answers to artifacts or enchantments.
  - Timing and colour: instant speed (instants and flash), sorcery speed, two or more colours, colourless.

`DraftCard` has `type_line`, `mana_cost`, `mana_value` and `colors`, but no rules text, power or toughness. Add `oracle_text`, `power`, `toughness` and `keywords` to the draft view's card as an additive wire field, with `docs/protocol.md` updated. Before matching words in the rules text for "what they do", check whether the catalog or Scryfall data already tags these roles (keywords, `produced_mana`) and prefer that.

## 7. The pick clock as a fuse

- [ ] A full-width bar under the header, pinned to the top when the page scrolls. It burns from the left: the remaining part shrinks toward the right, with a small spark at the burning end.
- [ ] Colour by time left: blue down to about 55 %, orange by 35 %, red by 15 %. It flashes in the last 8 seconds.
- [ ] Driven from the server's deadline (the existing `secondsLeft`), smooth (requestAnimationFrame), not stepped per second. Keep the text clock beside it.
- [ ] With reduced motion, no spark animation and no flashing; the colour still changes.

## 8. Page behaviour Ian flagged

- [ ] **One scroll bar.** The picks column and the pairs panel are `position: sticky` with their own `max-height` and `overflow-y: auto`, so the page has two or three scroll bars. Drop the inner scrolling. A column shorter than the window sticks under the fuse. A taller one scrolls with the page until its end shows, then sticks there (`top: min(24px, 100vh - height - 16px)`, set from a ResizeObserver).
- [ ] **Corner buttons don't hide content.** Full stats and Deck view are half transparent (opacity 0.5 with a backdrop blur) and fully opaque on hover or focus. The page gets about 96 px of bottom padding so the last row can scroll clear of them.

## Test plan

- [ ] `DraftSplit` / `DraftPool` render tests: a pointer drag from the pack to Sideboard picks and sets the card aside; a drag between zones moves it; the click fallback still works.
- [ ] Preview: shown after the delay beside the card; swaps without delay; hidden on drag, Escape and scroll.
- [ ] Rating reasons: unit tests on the reason builder for a removal spell, a flier, a two-colour card, and a card with no `avg_seen`.
- [ ] Stats: the overlay opens and closes with the button, `S` and `Esc`, returns focus, and locks the page's scroll; unit tests on pip counts, the 17-basics split, type and role counts and type lines.
- [ ] Fuse colour function at 100, 50, 30, 10 and 0 % time left; the flash state under 8 s.
- [ ] Layout at 1440, 1100, 820 and 400 px: one scroll bar, the pairs strip widens the pack, and the corner buttons don't cover the last row.
- [ ] Labels go through `L.<key>` (`labels.ts`), with e2e specs updated for any rename.
- [ ] Copy is proofread and spelled as the rest of the app ("Colours", "Colour pairs"): run a spell check over the new labels and reason strings before review.
- [ ] Manual check on cmd-dev with a bot draft, mouse and touch.
