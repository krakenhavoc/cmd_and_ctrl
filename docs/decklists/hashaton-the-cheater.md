# Decklist support — Hashaton the Cheater

Source: [Archidekt](https://archidekt.com/decks/21873193/hashaton_the_cheater) —
Esper reanimator, built around *discard a fatty, copy it, or put the
real one back*.

This is a **triage of one real deck against the catalog**, in the same
shape as [the Pirates list](pirates-mary-read-anne-bonny.md) and
[the Aang list](aang-is-so-flashy.md). A deck is a better forcing
function than a card count, because it says which gaps actually stop a
game from being played.

Counts were recomputed from the deck's **94 entries** resolved against
the Scryfall dump, not from the deck page. All 94 resolve; nothing is
missing from the index. **33 entries are lands** (Island / Plains /
Swamp ×3 each as three of them), so the 100 cards split **39 land / 61
nonland**.

| | before this batch | after |
|---|---|---|
| nonland cards working | 5 | **9** |
| nonbasic lands working | 9 | **29** |
| basics working | 9 | 9 |
| **total of 100** | **23** | **47** |

Two of those 24 new cards are partials that ship with a clause
missing; they are named in "Sandbox simplifications declared in this
batch" below, not hidden in the count.

## Card index

The dump was refreshed **2026-09-09** and every name in the deck
resolves.

**One trap cost real time and is worth writing down.** Several cards
have a placeholder printing in `default-cards.json` whose `type_line`
is the literal string `"Card // Card"` and whose oracle text is empty.
A name→oracle_id resolver that takes the first printing it sees will
pick one of those and conclude the card is a multi-face DFC. In this
deck it hits **Hoarding Broodlord, Late to Dinner, Rottenmouth Viper
and Seachrome Coast** — four ordinary single-faced cards. Skip any
printing whose `type_line` is `"Card // Card"`, or whose `layout` is
`art_series` / `token` / `double_faced_token`, and prefer a printing
that actually carries `oracle_text`.

The dump is 629 MB. Load it **once** and resolve every name in that
pass (~5 s for the whole deck); shelling out per card is minutes of
wall clock for the same answer.

Every oracle ID shipped in this batch was cross-checked a second way:
all non-placeholder printings of the name in the dump must agree on
one `oracle_id`. All 24 did.

## Done (this batch)

| Card | What it exercises |
|---|---|
| Hashaton, Scarab's Fist | the commander — a discard trigger, an optional mana cost, and a token that is a **copy of a specific card** |
| Reanimate | reanimation from **any** graveyard, under the caster's control, with a mana-value life cost |
| Zombify | the same move restricted to your own graveyard |
| Late to Dinner | reanimation plus an unconditional rider (a Food token) |
| Godless Shrine, Hallowed Fountain, Watery Grave | shocklands — *partial*, see below |
| Concealed Courtyard, Darkslick Shores, Seachrome Coast | fastlands — first **conditional** enters-tapped |
| Drowned Catacomb, Glacial Fortress, Isolated Chapel | checklands — the condition reads land **types** |
| Deserted Beach, Shattered Sanctum, Shipwreck Marsh | slowlands — the fastland condition inverted |
| Meticulous Archive, Shadowy Backstreet, Undercity Sewers | surveil lands — *partial*, see below |
| Bleachbone Verge, Floodfarm Verge, Gloomlake Verge | verges — *partial*, see below |
| Raffine's Tower | tri-land, enters tapped — *partial* (no cycling) |
| Prismatic Vista | a fetchland that fetches any basic |

Already in the catalog from earlier sprints: **Arcane Signet,
Consecrated Sphinx, Counterspell, Esper Sentinel, Swords to
Plowshares**, plus **Command Tower** and eight fetchlands (Arid Mesa,
Bloodstained Mire, Flooded Strand, Marsh Flats, Misty Rainforest,
Polluted Delta, Scalding Tarn, Verdant Catacombs) on the land side.

### Three new pieces of machinery

**`CreateTokenCopy` — a token that is a copy of a specific card.**
`CreateToken` builds a token from a hand-written template, which
covers every token that is its own card and none of the ones that are
a copy of something already in the game. The new primitive reads the
copiable values (CR 707.2) off a real card in **any** zone and hands
them to the existing token path.

What makes it cheap is that a token's identity in this engine is
already just a `game.Card`, and every catalog ability hook keys on
`OracleID`. Carry the copied card's oracle ID onto the token and its
triggered abilities, statics, printed keywords, mana abilities and
activated abilities all come along, because each of those hooks does a
catalog lookup rather than reading a field. A copy of Serra Angel
really has flying and vigilance; there is a test that says so.

It deliberately does **not** copy the source `Card` struct wholesale.
Counters, `KnownBy`, `ExilePlay`, battlefield position and the cached
`effective` characteristic are per-instance state, not copiable values
— and the last of those is a pointer, so a struct copy would leave the
token and its original sharing one layer cache.

**`MayPay` — "you may pay {N}. If you do, ...".** This is `PayUnless`
with the consequence hanging off the other answer. Both are one dialog
with two answers; Rhystic Study's rider happens when the payment does
*not*, Hashaton's happens when it does. Rather than a second
`PendingChoice` kind, `payUnlessFrame` grew an `onPay` alongside its
`onDecline` and `ResolvePayUnless` picks the branch on `paid` — not on
`apply`, so a player who says "Pay" and cannot fund it has declined,
which is already the right behaviour for both shapes.

**`EntersTappedUnless` — conditional enters-tapped.** One mechanic
behind three whole land cycles. It is `SelfEntersTapped` with a
predicate read off game state as the permanent enters, phrased so the
condition matches the printed text (which always states when the
drawback does *not* apply).

The condition is evaluated **pre-push**, which is not a compromise but
the reason it is easy: the entering land is not on the battlefield
yet, so a plain walk of `g.Battlefield.Cards` counts exactly the
*other* lands its controller controls. "Two or fewer other lands" and
"two or more other lands" both fall straight out with no exclusion
logic.

### One correction to the Aang doc

The Aang list records that "**a catalog replacement effect cannot fire
on its own source's entry**", verified by probe when Azorius Chancery
was written. **That is no longer true**, and any future work should
not plan around it.

S21's enters-tapped pass added a third gathering block to
`gatherActiveReplacementsLocked` (`server/internal/game/replacements.go`)
that consults the **entering** card's own replacements, keyed off
`LookupCardForEffect(ev.CardID)` and skipped for a card already on the
battlefield. `SelfEntersTapped` is built on it and the whole Temple
cycle uses it. This batch's twelve conditional lands are real CR 614
self-replacements on the back of that, not `OnETB` taps —
`hashaton_test.go` pins the difference by counting tap events rather
than checking `Tapped`, which both approaches satisfy.

`azorius_chancery.go` still carries the stale comment and still uses
the `OnETB` workaround. Not touched here (it is another batch's card
and the behaviour is only subtly wrong), but it is a one-line fix for
whoever is next in that file.

### Sandbox simplifications declared in this batch

- **Shocklands enter tapped and then untap.** "As this land enters,
  you may pay 2 life. If you don't, it enters tapped" is a
  replacement *with a choice inside it*, and the replacement pipeline
  is synchronous — it cannot stop and ask anyone anything. So the land
  enters tapped (a real self-replacement) and an **optional ETB
  trigger** offers "pay 2 life to untap it". The choice survives,
  which is the half that matters: a player at 3 life can decline. What
  is observably wrong is that the land is tapped for the window
  between entering and the trigger resolving, it emits an untap event,
  and opponents get priority in between — none of which happens in
  paper. Picking one branch at build time would have been silently
  wrong in one of the two cases, which is worse. The trigger is not
  offered below 2 life.
- **The surveil lands do not surveil.** Enters-tapped and the dual
  mana are real; "when this land enters, surveil 1" is **not
  implemented**, and in a reanimator deck that is the entire reason to
  play them over a Guildgate. Surveil is scry's structure with "bottom
  of library" replaced by "graveyard" — a new `PendingChoice` kind, a
  new resolver, a new clone leg and a new client dialog. Reusing
  `PendingChoiceScry` would be wrong in exactly the direction that
  matters here, because a card on the bottom is not a card in the
  graveyard.
- **The verges ship with only their unconditional mana ability.** A
  verge is two abilities, one carrying "activate only if you control a
  Plains or a Swamp". `ManaAbilityCost` has no condition slot and
  `ActivateManaAbility` has no hook to evaluate one, so the restricted
  half is **absent** rather than free. Each verge is therefore a
  strictly worse card than printed — a mono-coloured land — which is
  the safe direction, but it means a verge does not fix colours, which
  is its job.
- **Raffine's Tower has no cycling.** Cycling is "{3}, Discard this
  card: Draw a card" — an activated ability with a discard cost,
  activated *from hand*. `game.AbilityCost` has no discard component
  (the `DiscardCost` from #230 is a spell-cast cost), and the CR 602
  path only offers abilities on battlefield permanents.
- **Checkland conditions read printed type lines.** A land that has a
  basic land type only by way of a continuous effect — Urborg making
  everything a Swamp, or Blood Moon the other way — does not switch a
  checkland on or off. Reading `Effective()` would fix it, but the
  checkland is the *entering* card and the layer cache is maintained
  only for battlefield permanents.
- **Hashaton's copy is read from wherever the card now sits**, not
  from last-known information in the graveyard. The two agree in
  practice, because copiable values are printed values and do not
  change by zone; a card exiled in response would be copied out of
  exile rather than failing over to LKI. There is no LKI store
  reachable from an ability's `Effect`.
- **Prismatic Vista** inherits the fetchland family's gap, stated once
  in `fetchland_helpers.go`: `SearchLibrary` takes the **first** match
  in library order, so the player does not choose which basic.

### Two engine bugs found, both in shared code

**1. Reanimation put the creature under the wrong player's control.**
`ReturnFromGraveyardForEffect` picked its destination off `src.Owner`
— the graveyard's owner — which is correct for hand and library
destinations and wrong for the battlefield. "Put target creature card
from a graveyard onto the battlefield **under your control**" is the
single most valuable line in a reanimator deck, and it handed the
creature straight back to the opponent it was taken from.

Fixed by splitting out
`ReturnFromGraveyardUnderControlForEffect(cardID, dest, controller)`,
with the old signature delegating to it with `uuid.Nil` ("under its
owner's control") so existing callers are byte-for-byte unchanged. The
controller is stamped **before** the events fire, so the layer
listener, the trigger harvester and anything watching `EventETB` all
see the permanent under the right control. The primitive gained a
`Controller` field; "from your graveyard" cards pass it anyway rather
than relying on a default that is only correct for half the family.

**2. There are two independently-wired ETB mechanisms, and this path
fired only one.** `ReturnFromGraveyardForEffect` emitted `EventETB`
but never called `fireETBHookLocked`. Triggered abilities harvest off
the event log and always worked; a card whose ETB lives in
`Spec.OnETB` did not fire at all. Reanimating Solemn Simulacrum
fetched nothing, and reanimating a planeswalker gave it no loyalty
counters (`StartingLoyalty` is stamped by the same hook).

Fixed in the same function. The regression test reanimates a Temple —
whose ETB is *only* an `OnETB` scry — and asserts the prompt arrives,
addressed to the new controller rather than the graveyard's owner, so
it pins both fixes at once.

Worth knowing more generally: **any code path that puts a card onto
the battlefield must fire both.** `CreateTokenForEffect` is the other
one that does not, which is why a token copy gets the copied card's
`Triggered` ETB abilities but not its `OnETB` clause. That is declared
on `CreateTokenCopy` and left alone here — fixing it is safe (every
existing token has an empty `OracleID`, so the hook is a no-op for
them) but it is not this deck's bug.

## Blocked, by machinery needed

Grouped by what would unblock them, ordered by how many cards each
buys. 52 of the 61 nonland cards, plus Urborg on the land side.

**Cheap wins — 15 cards blocked on nothing but effort.** This is the
largest group in the deck and it is worth putting first, because
"needs a sprint" and "needs an afternoon" read identically in a list
of card names. Every one of these is expressible with machinery that
already exists:

- **Elesh Norn, Grand Cenobite** — two Layer 7 statics, the anthem
  pattern with one of them negative.
- **Sheoldred, the Apocalypse** — deathtouch plus two `EventDrawCard`
  triggers, the shape Consecrated Sphinx already uses.
- **Sheoldred, Whispering One** — an upkeep reanimation trigger and an
  each-opponent-sacrifices trigger; both primitives exist, and the
  reanimation half now controls correctly. (Swampwalk is cosmetic
  until evasion grows.)
- **Ashen Rider** — flying, plus one ETB and one dies trigger sharing
  a targeted exile.
- **Loran of the Third Path** — vigilance, a targeted ETB destroy
  (the Acidic Slime pattern), and a tap ability.
- **Lotho, Corrupt Shirriff** — "a player casts their second spell
  each turn" is `CastTallyFor`, which already exists for Brineborn
  Cutthroat.
- **Collector's Vault** — `lootOne` plus a Treasure token, on a
  two-part activated cost.
- **Vohar, Vodalian Desecrator** — the loot half with its
  instant-or-sorcery rider. (The sacrifice half needs cast-from-
  graveyard.)
- **Massacre Wurm** — the "a creature an opponent controls dies"
  drain. (The ETB -2/-2 needs until-end-of-turn effects, below.)
- **Archfiend of Ifnir** — flying plus the discard trigger's -1/-1
  counters. (No cycling.)
- **Champion of Wits** — the ETB draw-then-discard. (No eternalize —
  though eternalize is `CreateTokenCopy` plus an alternative cost, so
  it is closer than it was this morning.)
- **Fatal Push** without revolt — `ManaValueLE(2)` is already a
  predicate. Revolt needs a "a permanent left the battlefield under
  your control this turn" flag.
- **Jin-Gitaxias, Core Augur** minus the end-step draw (see the event
  gap below) and the hand-size clause.
- **Portal to Phyrexia** — the upkeep reanimation works now; the ETB
  wants "each opponent sacrifices **three**", which is
  `EachPlayerSacrifices` with a count.
- **Dark Ritual** — needs one small primitive: there is no
  `AddManaForEffect`. Every mana in the engine today comes from a mana
  ability, and "Add {B}{B}{B}" is a spell.

**Discard as an ability cost — 5 cards.** Guardian of New Benalia,
Seasoned Hallowblade, Psychic Frog, Key to the City, Ghostly Pilferer.
`game.AbilityCost` has tap / sacrifice-self / sacrifice-other / mana /
life and no discard, even though `Spec.AdditionalCost`'s
`DiscardCost(n)` already does the picking for spells. One field plus
reusing that picker. Note three of those five *also* want
indestructible, so this alone unblocks two.

**Keywords outside the canonical twelve — 5 cards.** Indestructible
(Avacyn, Guardian of New Benalia, Seasoned Hallowblade), protection
from a chosen card type (Serra's Emissary), ward (Valgavoth, and
Raffine's ward {1}). Same finding the Aang doc recorded: the static
half is cheap, the SBA / targeting behaviour behind it is the real
work. Avacyn additionally *grants* indestructible to other permanents.

**Alternative and choice-shaped costs — 6 cards.** Bitter Triumph and
Bone Shards and Souls of the Lost each print "**X or Y**" as an
additional cost; Toxic Deluge prints "pay X life" with X chosen;
Rottenmouth Viper prints "you may sacrifice any number"; Wash Away has
cleave. `Spec.AdditionalCost` is a single fixed clause with no "or",
no variable and no *replacement* of the mana cost. This is the same
keystone the Aang list put first, seen from a different angle: that
deck wanted evoke / foretell / spree, this one wants a two-branch
additional cost. One cost model serves both.

**Cast from a zone that isn't your hand — 5 cards.** Hoarding
Broodlord and Valgavoth (play from exile, with a life-instead-of-mana
cost), Vohar and Jace's −3 (cast an instant or sorcery from a
graveyard), Rite of the Moth (flashback). #232's `Card.ExilePlay` is
most of the exile half already; the graveyard half is the same idea
one zone over.

Two adjacent cards want a **source zone on `EventCast`**, which the
event does not carry: Ghostly Pilferer ("an opponent casts a spell
from anywhere other than their hand") and Wash Away ("target spell
that wasn't cast from its owner's hand"). The Aang doc flagged the
same one field for Appa. Three cards across two decks for one struct
field is the best ratio on either list.

**Attack triggers — 4 cards.** Raffine (whenever you attack),
Threefold Thunderhulk and Rottenmouth Viper (enters **or attacks**),
Smuggler's Copter (attacks or blocks). The Aang doc called this the
cheapest item on its list and a sibling branch has now built it; these
four should be re-checked against `EventAttack` once it lands rather
than treated as blocked. Raffine also needs connive and Smuggler's
Copter needs crew, so the trigger alone does not finish them.

**Copy effects that are not tokens — 3 cards.** Jin-Gitaxias, Progress
Tyrant and Kitsa, Otterball Elite copy a **spell on the stack**;
Likeness Looter makes a permanent **become** a copy (a Layer 1
effect). `CreateTokenCopy` does not reach either — a spell copy is a
new stack item, and a become-a-copy is a continuous effect on an
existing permanent. What it does give both is the "read the copiable
values off a card" half; the remainder is stack plumbing and Layer 1
respectively. Note the connection is real though: "a copy of a
permanent spell becomes a token" is exactly the token-copy path.

**Multi-face cards — 3 cards.** Altar of Bhaal // Bone Offering
(adventure), Jace, Vryn's Prodigy // Jace, Telepath Unbound
(transform), Rona, Herald of Invasion // Rona, Tolarian Obliterator
(transform). `game.Card` has one name, one type line, one mana cost.
Unchanged from the Aang doc's assessment: this touches import,
protocol and client, so it is bigger than its card count.

**Conditional / choice-driven mana abilities — 4 cards, plus the three
verges.** Chrome Mox (imprint — colours of an exiled card), Coldsteel
Heart (a colour chosen as it enters), Mox Amber (colours among
legendary permanents you control), Springleaf Drum (tap another
creature as a cost). All four want the same missing hook: a per-
activation narrowing of the pipe set, which the engine has *exactly
once* today, hard-wired for commander identity on Arcane Signet.
Generalising that hook is 7 cards in this deck alone.

**Planeswalker loyalty abilities — 2 cards.** Liliana of the Veil and
Jace's back face. `AbilityCost` has no loyalty component, and loyalty
abilities are sorcery-speed and once per turn. Unchanged from the Aang
doc.

**Until-end-of-turn effects — 2 cards.** Massacre Wurm's ETB
("creatures your opponents control get -2/-2 until end of turn") and
Toxic Deluge ("all creatures get -X/-X until end of turn"). The layer
engine recomputes from scratch from *permanent* sources; there is no
place to hang a temporary modification that expires at cleanup. This
is a genuinely common Commander shape and nothing in the catalog has
it yet.

**End-step triggers — 2 cards.** Jin-Gitaxias, Core Augur ("at the
beginning of your end step") and Toxrill ("at the beginning of each
end step"). `events.go` has `EventBeginUpkeep` and no end-step
counterpart. One new `EventKind` emitted at one step boundary.

**Cost reduction — 2 cards.** Overwhelming Remorse and Rottenmouth
Viper. S28 (cost engine), as already scheduled.

**Delayed triggers — 1 card.** Mana Drain's "at the beginning of your
next main phase, add {C} equal to that spell's mana value". The Aang
doc names the same gap for flicker's "return it at the next end step";
`game.TriggeredAbility` has no shape for a one-shot trigger scheduled
for later.

**One-offs.**

- **Hullbreaker Horror** — flash and a modal cast-trigger bounce are
  both expressible; "this spell can't be countered" is not.
- **Serra's Emissary** — "as this enters, choose a card type", a
  choice made during entry, the same shape shocklands could not
  express.
- **Valgavoth** — a replacement redirecting *opponents'* graveyard
  moves to exile, plus play-from-exile with a life cost, plus ward.
- **Toxrill** — end-step triggers, slime counters, a static reading
  counter counts, and a sacrifice-a-Slug ability.
- **Threefold Thunderhulk** — enters-with-counters is expressible
  (`AddCounterAtETB`); the token count reads its own power, which is
  fine, but the attack half is blocked.
- **Solar Transformer** — energy. Counters live on cards; energy is a
  player-scoped resource with no store.
- **Souls of the Lost** — the CDA power/toughness is the Tarmogoyf
  pattern and would work; the "discard a card **or** sacrifice a
  permanent" cost is what blocks it.
- **Raffine, Scheming Seer** — connive (draw N, discard N, then count
  the nonland cards discarded *this way*) needs the discard result fed
  back to the effect. `DiscardChoiceForEffect` queues and forgets.
- **Rottenmouth Viper** — blight counters driving an
  "each opponent loses 4 unless they sacrifice **or** discard" clause;
  a two-branch `PayUnless` on a non-mana cost.
- **Guardian of New Benalia** — enlist, which is an attack-time tap of
  another creature. Blocked twice over.

## Lands

**29 of the 30 nonbasics work after this batch**, plus the 9 basics.

The land base is the part of this deck the catalog was closest to
supporting, and almost all of it turned out to be one mechanic:
*conditional* enters-tapped. Twelve cards across three cycles
(fastlands, checklands, slowlands) differ only in the predicate, and
the predicate is a battlefield walk. Once `EntersTappedUnless` existed,
each cycle was three lines of data.

Ranked by how much of the printed card actually works:

- **Complete**: the three fastlands, three checklands, three slowlands,
  Prismatic Vista, and the eight already-shipped fetchlands + Command
  Tower.
- **Choice preserved, timing wrong**: the three shocklands (enter
  tapped, then an optional trigger untaps them for 2 life).
- **Colour-correct, clause missing**: the three surveil lands (no
  surveil), Raffine's Tower (no cycling).
- **Strictly weaker**: the three verges (only the unconditional half of
  their mana).
- **Not in the catalog**: **Urborg, Tomb of Yawgmoth**.

Urborg deserves its own note because it looks easy and isn't. "Each
land is a Swamp in addition to its other land types" is a
straightforward Layer 4 static and would apply cleanly. It would also
do **nothing**, because the synthetic basic-land mana ability is
derived from the **printed** `TypeLine`, not from the post-layer
effective types — so every land would say "Swamp" and none of them
would tap for {B}. Shipping the static alone would be a card that
looks implemented and is purely cosmetic, which is the failure mode
this doc exists to avoid. It needs `ManaAbilitiesForCard` to consult
`Effective()`, and that same change is what would make the checkland
conditions above read type-changing effects correctly.

## Suggested order

Sequenced by cards-per-unit-of-work across *both* decklists, not by
this deck alone.

1. **The cheap-wins batch.** 15 cards here need no new machinery at
   all, and several (Elesh Norn, both Sheoldreds, Ashen Rider) are
   format staples that will recur in every list. Highest ratio on
   either doc by a wide margin. Do this before building anything.
2. **Discard as an ability cost.** One field on `game.AbilityCost`,
   reusing the picker `DiscardCost` already has. 5 cards here; it is
   also half of cycling, which unblocks Raffine's Tower and Archfiend
   of Ifnir.
3. **A source zone on `EventCast`.** One struct field. Two cards here,
   one in the Aang deck, and "cast from anywhere other than hand" is a
   common Commander clause.
4. **Alternative / two-branch costs.** Still the keystone the Aang doc
   named. 6 cards here, 7 there, plus three catalog cards already
   shipping without their headline mode. The two decks want different
   faces of it (cleave and "X or Y" here, evoke and foretell there),
   which is an argument for designing the cost model once rather than
   bolting on a clause per card.
5. **Conditional mana abilities.** Generalise the per-activation
   narrowing hook that exists hard-wired for Arcane Signet. 7 cards in
   this deck (three verges, Chrome Mox, Coldsteel Heart, Mox Amber,
   Springleaf Drum).
6. **Until-end-of-turn effects.** 2 cards here, but it is one of the
   most common shapes in Magic and the layer engine has no slot for it
   at all. This one is bigger than its count.
7. **Surveil.** 3 cards here, and it is the deck's engine rather than a
   nicety. Scry's structure with a different destination — a new
   `PendingChoice` kind end to end.
8. **An end-step event.** One `EventKind`, 2 cards, and it pairs with
   the delayed-trigger slot the Aang doc wants for flicker.
9. **Play/cast from a graveyard.** 5 cards, and it reuses #232's
   `ExilePlay` permission model one zone over.
10. **Multi-face cards.** 3 cards here, 7 there, and it touches import,
    protocol and client. A sprint of its own, exactly as the Aang doc
    concluded.

Two things worth folding into whichever PR is nearby: **make
`ManaAbilitiesForCard` read `Effective()`** (unblocks Urborg and
sharpens the checkland conditions), and **fire `fireETBHookLocked`
from `CreateTokenForEffect`** so a token copy gets the copied card's
`OnETB` as well as its triggers — safe today, because every existing
token has an empty oracle ID.
