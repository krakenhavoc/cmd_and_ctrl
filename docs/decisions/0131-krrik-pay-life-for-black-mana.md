# ADR 0131 — K'rrik: paying 2 life for {B} in any cost

**Status:** Accepted (owner answers 2026-10-07) · 2026-10-07 · S68 — Cost components and alternative costs
**Issues:** [#2531](https://github.com/krakenhavoc/cmd_and_ctrl/issues/2531) (K'rrik, Son of Yawgmoth's payment grant; part of the "Life is just a resource" Betor deck goal). Earlier triage: [#1117](https://github.com/krakenhavoc/cmd_and_ctrl/issues/1117).
**Owner decisions:** the owner answered this ADR's six questions on 2026-10-07, each with the recommended option. The answers are quoted under [Owner decisions](#owner-decisions-2026-10-07) and are binding. The options not chosen are kept under [Questions for the owner (answered)](#questions-for-the-owner-answered).
**Numbering:** the AGENTS.md §4 sweep on 2026-10-07 found 0130 (`0130-exert.md`) as the highest number on any remote head. This ADR takes **0131**.
**Builds on:** [ADR 0011](0011-mana-pool-and-auto-tapper.md) (the auto-tapper), [ADR 0085](0085-life-total-cant-change.md) (a life total that can't change), [ADR 0118](0118-strict-payment-by-default-and-alternative-costs-for-every-spell.md) (strict payment and the pool top-up), [ADR 0126](0126-bots-that-play-their-decks.md) (the heuristic), [ADR 0127](0127-answering-repeated-prompts-for-you.md) (repeated prompts) and [ADR 0129](0129-energy-getting-and-paying-it.md) §5 (the auto-tapper spends energy last, before life). It reuses the Phyrexian life payment of #787, #916 and #917, and the payer-scoped spend grant of #1600.

This ADR was written plan-first. No code changed with it. The changes land in the PRs listed under [Delivery](#delivery).

---

## Context

K'rrik is catalogued with a caveat: "only its own printed Phyrexian mana symbols can be paid with life." Its third ability is missing. Every claim below was checked on `origin/develop` at `bdbc0cbad`. Every rule was checked against the pinned Comprehensive Rules (`MagicCompRules 20260925.txt`).

### Owner decisions (2026-10-07)

The owner answered this ADR's six questions on 2026-10-07, each with the recommended option:

1. **A player static that marks granted {B} at payment, before the spend-only fold** (question 1, §1).
2. **Every mana payment its controller makes**, including the {B} half of hybrid symbols and never generic mana (question 2, §2).
3. **Auto-tap never pays life** (question 3, §3).
4. **The client asks only when mana is short, and a "Pay life for {B}…" entry is always available** (question 4, §4).
5. **The bot gets the same life options as for a printed Phyrexian symbol, with the floor of 10** (question 5, §5).
6. **Two PRs** (question 6, [Delivery](#delivery)).

### The card and its rulings

> For each {B} in a cost, you may pay 2 life rather than pay that mana.

The Scryfall rulings for oracle ID `cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd` settle most of the scope:

- (2024-06-07) It applies to "{B} in any cost you pay, including the mana costs of spells, activation costs, and even costs for special actions (such as morph). Any time you pay mana, that's a cost."
- (2019-08-23) It "doesn't modify or reduce costs you pay. It changes only how you may pay those costs."
- (2019-08-23) "You can't pay 2 life to pay for generic mana in costs you pay, even if an effect says that you must spend black mana to pay that generic mana."
- (2019-08-23) For "{B/R}, {B/P}, or {2/B}, you choose how you'll pay it before you do so. If you choose to pay {B} this way, K'rrik's ability allows you to pay life rather than pay that mana."
- (2024-06-07) A Phyrexian symbol counts toward mana value even if life is paid for it. K'rrik's mana value is always 7.

### The rules

- **CR 107.4e:** a hybrid symbol "can be paid in one of two ways"; "{2/B} can be paid with either one black mana or two mana of any type." So {B/G} and {2/B} each contain a {B} that K'rrik reaches.
- **CR 107.4f:** a Phyrexian symbol "can be paid either with one mana of its color or by paying 2 life", and a hybrid Phyrexian symbol "with one mana of either of its component colors or by paying 2 life." A {B} under K'rrik pays exactly like a {B/P}.
- **CR 118.3:** "A player can't pay a cost without having the necessary resources to pay it fully." **CR 118.3b:** paying life subtracts it from the life total.
- **CR 119.4:** a player may pay life "only if their life total is greater than or equal to the amount of the payment." **CR 119.8:** if a player can't lose life, "a cost that involves having that player pay life can't be paid."
- **CR 601.2b:** the caster announces how they will pay hybrid and Phyrexian symbols. **CR 602.2b** applies 601.2b–i to activated abilities. **CR 601.2f** fixes the total cost, reductions included, before **CR 601.2h** pays it.
- **CR 118.9** and **118.9a:** an alternative cost is paid "rather than paying the spell's mana cost", and only one applies to a spell. K'rrik's grant is not one (the 2019 ruling: it changes how a cost is paid, not the cost), so it combines with evoke or dash.
- **CR 202.3:** mana value is "the total amount of mana in its mana cost". Paying life does not change it.

### What exists

- **The life payment.** `CastSpellParams.PhyrexianLife` (`game/mutations.go:442-465`, #787) and `ActivateAbilityParams.PhyrexianLife` (`game/activated.go:1001-1021`, #917) carry the announced count of symbols paid with life, wire `phyrexian_life`. One helper pair serves both paths: `strikePhyrexianLifeLocked` checks the count against `ParsedCost.PhyrexianSymbols()` and `CanPayLifeLocked` (CR 119.4, and CR 119.8 through ADR 0085's lock), and `payPhyrexianLifeLocked` pays through `PayLifeForEffect` (`game/phyrexian_mana.go:134-202`). `PhyrexianLifePlan` (`:94`) chooses which symbols to strike, starting with the ones the pool can't pay. Attack taxes use it too (`attack_tax.go:456-457`).
- **A payer-scoped grant, read at the payment.** "Spend mana as though it were mana of any color" (#1600, `game/spend_any_color.go`) is a player static, read live off the battlefield. `costAsPaidByLocked` (`:176`) applies it at the top of every payment and every affordability probe: casts (`mutations.go:1816`, `:1934`), activations (`activated.go:2185`), mana abilities (`mutations.go:6598`), attack taxes, pay-unless (`pending_choice.go:4023`), the costed auto-tapper (`autotap_costed.go:685`), the enumerator (`legal/cast.go:1667`, `legal/abilities.go:1844`, `legal/combat.go:148`) and the preview (`lobby/http.go:2268`). It marks `ColorRequirement.AnyMana` and keeps `Options`, so the price shown stays printed. It runs after the spend-only fold (`spend_only.go`), the fold that turns Crypt Rats' "Spend only black mana on X" into black requirements.
- **The gaps.** `ManaAbilityParams` (`mutations.go:6221`) has no `PhyrexianLife`, so a mana ability's mana cost can't be paid with life. The `pay_unless` prompt (`payCostLocked`) pays mana only. The view counts `phyrexian_symbols` by parsing the printed string (`protocol/view.go:5439`, read at `:5289`, `:5928`, `:9499`), so it can't see a grant. The enumerator's life loops (`legal/cast.go:1738`, `:1759`, `legal/abilities.go:1270`) strike the printed cost before `CostAsPaidByForEffect` widens it.
- **Auto-tap and life.** The auto-tapper never pays a Phyrexian symbol with life. The life is spent only when the announcer claims it. Its only life spending is the pain tier for mana abilities with a life cost (#2392), which ADR 0129 §5 put after the energy tier.
- **The client.** `shouldAskPhyrexianLife` (`client/src/lib/phyrexianLife.ts`) opens the life stepper whenever the cost has a Phyrexian symbol and the player can pay for one. `previewDecidesMana` (`client/src/lib/dragCast.ts:323`) skips the drag preview for such a card.
- **The bot.** The enumerator offers the cheapest life count that makes a move payable, plus an all-life variant. The heuristic refuses a Phyrexian life payment below 10 life (`phyrexianLifeFloor`, `aiseat/heuristic/moves.go:167`).

---

## Decision

### 1. The grant: a player static read at the payment

`game` gains `LifeForManaStatic{Label string; Color string}` beside `AnyColorSpendStatic`. It is catalog data on `CardDef.LifeForMana`, read live off the battlefield through `CatalogLifeForMana(key)`, and covers its source's controller. A reader `paysLifeForColorLocked(payer) []string` returns the colours granted to that payer. Two K'rriks grant the same colour once. A K'rrik that loses its abilities grants nothing.

`ColorRequirement` gains `LifeGranted bool`. A new method, `PaysWithLife() bool`, returns `r.Phyrexian || r.LifeGranted`. `PhyrexianSymbols()` and `PhyrexianLifePlan` read `PaysWithLife()` instead of `Phyrexian`. `String()` ignores the new field, so a price still renders as printed. A new function, `grantLifeForManaLocked(payer, cost)`, sets `LifeGranted` on each requirement whose `Options` contain a granted colour and that is not already Phyrexian. It is idempotent, and it returns the cost unchanged when nothing is granted.

**Where it runs.** `costAsPaidByLocked` calls it first, before `foldSpendOnly`. The fold then adds Crypt Rats' black requirements without the flag, which is what the generic-mana ruling requires. The any-colour widening runs after it, so a widened {B} keeps its life half (as #1589 did for {B/P}). Every payment and probe already calls `costAsPaidByLocked`, so they all see the grant, and the price shown stays printed (the 2019 ruling). Cost reductions and increases are applied when the cost is priced, before this runs (CR 601.2f). A reduced-away {B} is never offered, and a {B} that a tax adds is.

The enumerator's three life loops, and the view's count (§4), call the exported `LifeGrantedCostForEffect(payer, cost)` before they count symbols.

Considered: **rewriting the cost at pricing**, turning {B} into {B/P} in the CR 601.2f pass. It changes the price shown, and it pulls every black mana ability out of the costed auto-tapper, which refuses Phyrexian costs (`autotap_costed.go:123-127`). **A K'rrik alternative cost** is wrong under CR 118.9a, because it wouldn't combine with evoke or dash, and it can't reach activations. Both are question 1's alternatives.

### 2. Which costs it covers

The grant covers every mana payment its controller makes, as the 2024 ruling says:

- spells (mana cost, alternative cost, additional costs such as kicker, cost increases);
- activated abilities;
- mana abilities;
- attack taxes;
- resolution payments (`pay_unless`: ward, "counter unless its controller pays").

It reaches {B}, the {B} half of a hybrid symbol ({B/G}, {2/B}; CR 107.4e) and the {B} half of a hybrid Phyrexian symbol. {B/P} is unchanged: it is already 2 life. It never reaches generic mana, {C}, {S}, or a requirement the spend-only fold made. Other players' costs are untouched.

Two paths need new plumbing for this (PR 2):

- `ManaAbilityParams.PhyrexianLife`, the same field and wire name as the other two. It is refused when the cost has no symbol it could apply to.
- A life answer on `pay_unless`. `ResolvePayUnless` takes `phyrexianLife`, and `payCostLocked` strikes and pays through the same helper pair.

Face-down special actions pay no mana through a gated path today (`face_down.go`), so they need nothing.

### 3. Auto-tap: life is never paid unasked

**No change.** The auto-tapper never pays life for a granted {B}. A granted symbol is paid with mana unless the announcement claims it. CR 601.2b makes "how they intend to pay" part of the player's announcement, and the engine keeps that rule for printed Phyrexian symbols today. Paying life automatically on a mono-black K'rrik deck would quietly cost life on almost every spell. ADR 0129 §5's ordering (mana, then energy, then life) is about mana abilities with a printed life cost. It is not about choosing to pay life for a symbol, so it doesn't apply.

A spell that only life can pay is still offered. The enumerator counts the life variants (§5), so the card stays lit and the click opens the stepper (§4). Question 3 keeps the alternatives.

### 4. What the client shows and asks

- **The wire.** The three view sites stop parsing the printed string. They report the viewer's own count through `LifeGrantedCostForEffect`. `phyrexian_symbols` now includes granted symbols, and a new `phyrexian_granted` (`int`, omitted at zero) says how many of them are granted. `docs/protocol.md` documents both.
- **When to ask.** A printed Phyrexian symbol still opens the stepper, as today. If every symbol is granted, the client asks the auto-tap preview with `phyrexian=0` first. If mana pays the whole cost, the cast goes out claiming nothing. If not, the stepper opens, set to the smallest count that pays. The card menu and the dock gain **Pay life for {B}…**, a new label registered in `labels.ts` (ADR 0125 §2), which opens the stepper whenever the player wants it. `previewDecidesMana` treats granted-only symbols as decided by the preview.
- The stepper's readout and its CR 119.4 cap (`maxPhyrexianLife`) are unchanged.

Question 4 keeps the alternatives.

### 5. The bot

**No new code.** Once §1's helper feeds the enumerator's life loops, an `assisted`, `strong` or `heuristic` seat behind K'rrik is offered the same two variants it gets for a printed Phyrexian symbol. The heuristic prices them as it already does, with the life floor of 10 and no payoff credited for life. `MoveCost.PhyrexianLife` carries the count. Question 5 keeps the alternatives.

### 6. Interactions

- **Mana value** is unchanged (CR 202.3). The grant changes only the copy made at payment. A spell paid with life keeps its colour and its mana value, so K'rrik's own cast trigger still sees a black spell.
- **A life lock** (Platinum Emperion, ADR 0085) and CR 119.8: `CanPayLifeLocked` already refuses the claim, the enumerator already asks the same predicate, and the stepper's cap is 0. Nothing new is needed.
- **Too little life** is refused by CR 119.4 at the same gate.
- **Snapshot:** no change. `ParsedCost` isn't captured (`v7.txt` has no requirement fields), and the static is catalog data.

---

## Tests

All test-first. Each file name is a new file unless stated.

- `game/life_for_mana_test.go`:
  - a strict cast of {2}{B}{B} with two Swamps' worth of generic mana, no black and K'rrik: `phyrexian_life: 2` pays 4 life;
  - hybrid {B/G} and {2/B} accept life, while {G}, generic and {C} refuse it;
  - Crypt Rats' X under "Spend only black mana on X" refuses life for X;
  - an opponent's cost gets nothing, and neither does a K'rrik that changed controller or lost its abilities;
  - two K'rriks grant once;
  - Platinum Emperion, and life below 2, refuse the claim before any mana is spent;
  - a "costs {B} less" reducer removes the pip before the grant reads it;
  - evoke, dash and kicker {B} are covered;
  - the mana value on the stack is unchanged;
  - with Chromatic Orrery, the widened {B} keeps its life half;
  - the auto-tapper with `phyrexian_life: 0` pays mana and no life.
- `game/activated_test.go` (extend): an activation's {B} paid with life, and the error text for an over-claim.
- `legal/phyrexian_grant_test.go`: the enumerator offers a life-only cast under K'rrik and not without it, and the move's `MoveCost.PhyrexianLife` equals the claim.
- `protocol/view_test.go` (extend): `phyrexian_symbols` and `phyrexian_granted` count for the controller only.
- `cards/effects/krrik_son_of_yawgmoth_test.go`: the full card.
- PR 2: `game/mana_ability_phyrexian_test.go` (a {B}, {T} filter cost paid with life) and `game/pay_unless_life_test.go` (a ward {B} paid with life, and refused under a life lock).
- Client: `phyrexianLife.test.ts` and `dragCast.test.ts` cover granted-only asking, and `labels.test.ts` covers the new label.

## Delivery

| PR | Changes | Files |
|---|---|---|
| 1 | §1, §3–§6 for casts, activations and attack taxes; K'rrik's declaration; caveat narrowed to "mana abilities and costs paid while a spell or ability resolves" | `game/life_for_mana.go` (new), `game/mana_cost.go` (`LifeGranted`, `PaysWithLife`), `game/phyrexian_mana.go` (read `PaysWithLife`, error text), `game/spend_any_color.go` (`costAsPaidByLocked`), `game/carddef.go` + `cards/effects/carddef.go` (`LifeForMana`), `cards/effects/krrik_son_of_yawgmoth.go`, `legal/cast.go`, `legal/abilities.go`, `legal/combat.go`, `protocol/view.go`, `docs/protocol.md`, `client/src/lib/{protocol.ts,phyrexianLife.ts,dragCast.ts,labels.ts}`, `Board.svelte`, `docs/labels.md` |
| 2 | §2's two new paths; K'rrik's caveat removed and `Completeness` set to what is true | `game/mutations.go` (`ManaAbilityParams.PhyrexianLife`), `actions/actions.go`, `game/pending_choice.go` (`ResolvePayUnless…`, `payCostLocked`), `ws` / `actions` dispatch, the client's mana-ability menu and `pay_unless` dialog |

Each PR regenerates only K'rrik's oracle fixture (`-update-oracle -oracle-ids=cbe3a4e7-5dbe-4f58-8ee6-a1762b65acfd`). PR 1 adds a `pay-life-for-mana` row to `roadmap/registry.go` as partial, through `go test ./internal/roadmap/ -update`. PR 2 flips the row and adds its fragment under `docs/engine-seams/closed/`. PR 1 also adds a short "Paying life for coloured mana" recipe to `docs/adding-cards.md`. PR 2 follows PR 1.

## Consequences

- K'rrik plays as printed, and a black spell or ability its controller can't pay with mana stays castable.
- The life half of a symbol has one owner, `PaysWithLife`. Printed Phyrexian and granted symbols share the strike, the gate, the preview and the enumerator.
- No price shown changes, no snapshot changes, and the auto-tapper spends no life it didn't spend before.

## Out of scope

- Other payment-method grants (convoke-like or delve-like statics). `LifeForManaStatic` takes a colour, so a future "for each {G}" card is data.
- Face-down special-action costs, which aren't mana-gated (§2).

---

## Questions for the owner (answered)

Each question lists the most CR-faithful option first, which was also the recommendation. The owner chose the recommended option for all six on 2026-10-07 (owner decisions 1–6). The questions are kept with the options not chosen.

1. **How the grant is modelled (§1; CR 107.4f, 601.2f, 118.9a).**
   - (a) **Recommended:** a player static that marks granted {B} requirements `LifeGranted` at payment, in `costAsPaidByLocked`, before the spend-only fold. The price shown stays printed and the generic ruling holds, and every payment path picks it up, because they all already call this function.
   - (b) Rewrite {B} to {B/P} at pricing. It needs less plumbing, but the price shown changes and black mana abilities drop out of the costed auto-tapper.
   - (c) A K'rrik alternative cost. Simplest for spells, but it breaks CR 118.9a (no evoke or dash with it) and can't reach activations.

   **Answered: (a), as recommended (owner decision 1).**

2. **Which costs it covers (§2; the 2024 ruling, CR 107.4e).**
   - (a) **Recommended:** every mana payment its controller makes: spells with their alternative and additional costs, activations, mana abilities, attack taxes and resolution payments. That includes the {B} half of hybrid symbols, and never generic mana. It plays as printed, but it needs PR 2's two new paths.
   - (b) Casts, activations and attack taxes only, with a caveat for the rest. One PR, but a ward {B} or a filter land's {B} can't be paid with life.
   - (c) Spells only. Smallest, and most unlike the card.

   **Answered: (a), as recommended (owner decision 2).**

3. **Does auto-tap ever pay life for {B} (§3; CR 601.2b)?**
   - (a) **Recommended:** never. Life is paid only when the announcement claims it, as for printed Phyrexian symbols today. No life is lost without a click, and a spell only life can pay still lights up and opens the stepper.
   - (b) As a last tier, after mana and energy (ADR 0129 §5), shown in the preview. One fewer click when the player is short, but life is spent that the player never chose to spend.
   - (c) Always prefer life above a floor. It saves mana, but it decides a strategic choice for the player.

   **Answered: (a), as recommended (owner decision 3).**

4. **How the client asks (§4).**
   - (a) **Recommended:** if every symbol is granted, ask only when mana can't pay the cost, and add a **Pay life for {B}…** entry the player can use at any time. The choice is always there, without an extra click on every black spell.
   - (b) Always open the stepper, as for printed Phyrexian symbols. It is consistent, but it adds a click to almost every spell in a K'rrik deck.
   - (c) A seat setting (Ask, Only when short, Never), synced like ADR 0127's rules. Most flexible, but it adds a setting for one card.

   **Answered: (a), as recommended (owner decision 4).**

5. **How the bot uses it (§5; ADR 0126).**
   - (a) **Recommended:** no new code. It gets the same two life variants as a printed Phyrexian symbol, with the same floor of 10. The bot plays K'rrik as printed and nothing needs tuning.
   - (b) Lower the life price while the seat controls a lifelinker such as K'rrik. Sharper play, but it adds a board read and a weight to tune under ADR 0126 §8.
   - (c) Hide granted symbols from the bot. It needs no testing, but the bot plays K'rrik weaker than printed.

   **Answered: (a), as recommended (owner decision 5).**

6. **Delivery (Delivery).**
   - (a) **Recommended:** two PRs. PR 1 covers casts, activations and attack taxes, and PR 2 adds mana abilities and resolution payments. Each is small enough to review, and K'rrik is complete after PR 2.
   - (b) One PR with everything. A single review, but it touches the cast path, the mana-ability path, the prompt and the client at once.
   - (c) PR 1 only, with the rest on a registry `Waiting` row. Fastest, but K'rrik keeps a narrower caveat.

   **Answered: (a), as recommended (owner decision 6).**
