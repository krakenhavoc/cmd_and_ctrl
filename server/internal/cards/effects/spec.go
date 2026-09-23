// Package effects is the S14 card-effect catalog: declarative Forge-
// style specifications for the ~30 starter Commander cards that
// resolve automatically on the stack. Entries are keyed by Scryfall
// ID, registered at package init() time in one-file-per-card under
// this directory.
//
// Design pillars (pinned in docs/decisions/0010-card-effect-catalog.md):
//   - Declarative DSL in Go structs, no oracle-text parsing.
//   - Opt-in per Scryfall ID — non-catalog cards keep today's manual
//     sandbox posture.
//   - Effects run under the resolution write lock; primitives call
//     *Locked helpers on game.Game only. A primitive that calls a
//     public locking mutator deadlocks — enforce via review, Go can't
//     encode "locked context" at the type level.
//   - Events emitted per primitive application (see game.EmitEvent).
//
// This file declares the Spec type. Registry, Context, and primitive
// types live in neighbouring files.
package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spec is the canonical declaration for one card in the catalog.
// Populated once at package init() time via Register; looked up at
// resolution time via Lookup. Absence in the registry ⇒ non-catalog
// ⇒ today's manual sandbox behaviour.
type Spec struct {
	// OracleID is the Scryfall oracle-level card identifier — stable
	// across printings (every printing of Lightning Bolt shares one
	// oracle_id). Required. The registry panics if two Specs collide
	// on this key so a copy-paste duplicate fails at server boot
	// (the earliest, most visible failure mode).
	OracleID string

	// Name is the human-readable card name. Present in the Spec so
	// logs / panics / tests have a nice handle without bouncing
	// through the Scryfall index.
	Name string

	// OnResolve runs when the spell's stack item resolves (after
	// target re-check, before zone routing). Nil means "no effect
	// on resolve" — the card still routes to the battlefield /
	// graveyard per type, but no automation fires. Typical usage
	// for vanilla creatures with no ETB triggers (Birds of Paradise
	// in S14).
	//
	// Returning an error emits an EffectError event and allows
	// resolution to continue — partial failure does not wedge the
	// stack.
	OnResolve func(item *game.StackItem, ctx *Context) error

	// AsEnters runs synchronously as the permanent enters the
	// battlefield, inside the entry path and OFF the stack. It is
	// the CR 614.12 slot — "As this permanent enters, choose a
	// creature type" (ChooseCreatureTypeAsEnters) — and nothing
	// else belongs here: an "as enters" clause is not a triggered
	// ability, nobody gets a response window, and that is correct.
	//
	// A printed "When ~ enters" is a trigger and goes in Triggered
	// watching EventETB, so it uses the stack and can be answered.
	// This slot used to be called OnETB and was used for both until
	// #578 (Discussion #560); the rename is the guard against that
	// coming back.
	//
	// The Context carries no stack item (Item is nil), so read the
	// controller off the card, not off ctx.Controller(). Nil for
	// nearly every card.
	AsEnters func(card *game.Card, ctx *Context) error

	// StartingLoyalty is the loyalty counter count a planeswalker
	// enters the battlefield with. 0 means "not a planeswalker" or
	// "planeswalker with 0 starting loyalty" (which the SBA would
	// immediately kill — never the real case). The resolution path
	// stamps this via AddCounter when the card crosses into the
	// battlefield. Used by The Wandering Emperor in S14.
	StartingLoyalty int

	// TargetMode tells the client what to prompt for at cast time.
	// Serialized to CardView.target_mode so the Svelte cast flow
	// can enter a targeting state before firing cast_spell with
	// populated targets[]. Empty string means "no target prompt,
	// cast immediately." Valid values:
	//
	//   ""             no prompt (Pyroclasm, Wrath, vanilla permanents)
	//   "any"          player / creature / planeswalker / battle
	//                  (Lightning Bolt, Shock, Helix)
	//   "player"       seated player only
	//   "creature"     battlefield creature only
	//   "stack_spell"  an item currently on the stack (Counterspell,
	//                  Negate, Swan Song)
	//   "card_in_graveyard" a card in any graveyard (Regrowth, Eternal
	//                       Witness)
	//
	// Target count is always 1 in S14; multi-target (Arcing Lightning
	// style "distribute 3 damage") is S22 territory.
	//
	// S20: prefer Targets. When Targets is set, TargetMode is derived
	// from Targets.Mode and this field is ignored. Cards still on a
	// bare TargetMode get the S13.1 free-form picker with no
	// server-side legality check.
	TargetMode string

	// Targets is the S20 structured targeting clause: which zones and
	// players the single target slot accepts, and the predicate a
	// candidate must pass. Drives the client's legal-target set, the
	// announce-time check (CR 601.2c) and the resolution re-check
	// (CR 608.2b). Build it with the constructors in targets.go —
	// TargetAny(), TargetCreature("target non-black creature",
	// NonBlack()), TargetSpell(…), TargetCardInGraveyard(…). Nil
	// means no structured targeting.
	Targets *game.TargetSpec

	// ManaAbilities is the list of activated mana abilities the card
	// exposes from the battlefield. Each entry is one tap-or-cost-
	// for-mana ability — Sol Ring's "{T}: Add {C}{C}", Birds of
	// Paradise's "{T}: Add one mana of any color", Arcane Signet's
	// commander-identity-restricted variant
	// (NarrowToCommanderIdentity). Mana abilities do NOT use the stack
	// (CR 605.3); they resolve synchronously when the
	// activate_mana_ability action fires. The Index a client sends in
	// the action payload is the position in this slice.
	//
	// Empty / nil for cards that have no mana abilities (the vast
	// majority of S15 catalog). Basic lands get a synthetic default
	// ability derived from their TypeLine when no catalog entry
	// declares ManaAbilities — see Game.ActivateManaAbility for the
	// fallback shape.
	//
	// Added in S15 sub-PR 2.
	ManaAbilities []ManaAbility

	// Static is the list of continuous-effect static abilities the
	// card contributes to the layer engine while on the battlefield
	// (CR 613). Each entry declares its `Layer`, `SubLayer` (for
	// Layer7PT only), an `AppliesTo` predicate evaluated per
	// candidate target on every recompute pass, and an `Apply`
	// function that mutates the candidate's `Characteristic` in
	// place. The engine recomputes from scratch on every relevant
	// event (battlefield zone change, counter change) — so AppliesTo
	// re-runs against the current board state every time.
	//
	// Empty / nil for cards with no static abilities (the S15
	// majority). When non-empty, the card's effects only apply
	// while it's on the battlefield (CR 113.6 default — non-
	// battlefield-zone statics are deferred to a later sprint).
	//
	// Reaches directly into `game.StaticAbility` rather than a
	// per-package adapter type because the function shapes already
	// reference `game.Card` / `game.Game` / `game.Characteristic` —
	// no information would survive a parallel struct.
	//
	// Added in S16 sub-PR 3.
	Static []game.StaticAbility

	// Replacements is the list of CR 614 replacement effects the
	// card contributes while on the battlefield. Each entry
	// declares a `Watches` pre-filter (which event kinds it cares
	// about), an `AppliesTo` predicate evaluated per candidate
	// event, a `Replace` function that mutates or cancels the
	// event, and metadata (Controller, SelfReplacement, Label) for
	// the CR 616 order-choose prompt.
	//
	// Empty / nil for cards with no replacement effects (the S16
	// majority). When non-empty, the replacements fire only while
	// the card is on the battlefield (CR 113.6 default).
	//
	// Reaches directly into `game.ReplacementEffect` rather than a
	// per-package adapter type because the function shapes already
	// reference `game.ReplacementEvent` / `game.Game` / `game.Card` —
	// parallels the S16 `Static []game.StaticAbility` pattern.
	//
	// Added in S17 sub-PR 2. Populated by catalog cards starting
	// in sub-PR 3 (Doubling Season, Hardened Scales, Branching
	// Evolution).
	Replacements []game.ReplacementEffect

	// EntersWithCountersFromCast is the card's printed "this permanent
	// enters with N counters on it" where N is read from the SPELL
	// that became it (CR 614.1c, #1002) — Hangarback Walker's X,
	// Etched Oracle's sunburst. Build each clause with a constructor
	// from entry_counters.go and never by hand:
	//
	//	EntersWithCountersFromCast: []game.EntryCountersFromCast{
	//	    XCounters(game.CounterPlusOne),
	//	},
	//
	// It is NOT the slot for "enters with three +1/+1 counters" or
	// "enters with a counter for each Zombie card in your graveyard".
	// Those read the board, not the announcement, and stay ordinary
	// self-replacements in Replacements — b10EntersWithCounters and
	// b19EntersWithCountersCounted. This slot exists only because the
	// announcement is the one thing a replacement cannot reach on its
	// own: the engine seeds it onto the entry event from the resolving
	// StackItem (game/entry_counters.go).
	//
	// A permanent that did not come from a spell — reanimated, put
	// onto the battlefield, a token — declares nothing and enters with
	// none (CR 107.3b).
	EntersWithCountersFromCast []game.EntryCountersFromCast

	// PrintedKeywords is the list of combat keywords printed on the
	// card — entries like "flying", "reach", "deathtouch", "lifelink",
	// "trample", "vigilance", "first strike", "double strike",
	// "menace", "defender", "haste", "flash". Canonical lowercase
	// tokens; see AGENTS.md §7 "Adding a combat-keyword card" for the
	// complete table.
	//
	// Feeds two consumers:
	//
	//   1. On-battlefield: wire.go synthesizes a self-only Layer 6
	//      StaticAbility per card that appends each entry to the
	//      card's own `Characteristic.Abilities`, so the keyword
	//      lands in `card.Effective().Abilities` alongside grants
	//      from other cards' static abilities (Lord of Atlantis).
	//   2. Off-battlefield: game.HasKeyword (keywords.go) falls back
	//      to `CatalogPrintedKeywords(oracleID)` when the card has
	//      no `effective` cache — needed for flash gating on a
	//      card still in hand.
	//
	// Keyword behaviour itself is engine-side (flying block
	// restriction, trample overflow, etc.); card files just declare
	// the strings.
	//
	// Empty / nil for cards with no printed combat keywords (the
	// S17 majority). Added in S18 sub-PR 2.
	PrintedKeywords []string

	// Triggered is the list of CR 603 triggered abilities the card
	// declares for the S19 auto-fire dispatcher. Each entry watches a
	// set of EventKinds, predicates whether a specific event triggers
	// the source, and builds the StackItem to enqueue onto
	// g.PendingTriggers — same queue the manual S13.1
	// AnnounceTrigger feeds. The harvester (game.triggerHarvester)
	// walks the battlefield + LKI map on every event emit and runs
	// each declared TriggeredAbility's Watches → AppliesTo → Build
	// pipeline.
	//
	// Empty / nil for cards with no auto-fire triggers (the S18
	// majority). Sub-PR 1 ships the dispatcher framework with zero
	// catalog declarations — sub-PRs 3-7 fill in cards. Manual
	// AnnounceTrigger remains for non-catalog cards and "hidden info"
	// triggers (cards in hand with cast-replacement triggers).
	//
	// Reaches directly into `game.TriggeredAbility` rather than a
	// per-package adapter type — same rationale as Static and
	// Replacements.
	//
	// Added in S19 sub-PR 1.
	Triggered []game.TriggeredAbility

	// ManaTriggers are the CR 605.1b TRIGGERED MANA abilities: a
	// trigger that fires when a permanent is TAPPED FOR MANA, adds
	// mana, and therefore does not use the stack at all (CR 605.4a).
	// "Whenever enchanted land is tapped for mana, its controller adds
	// an additional {G}" — Wild Growth, Overgrowth, Utopia Sprawl,
	// Fertile Ground, Mana Flare, Mirari's Wake.
	//
	// NOT Triggered, and the test is one line: if the ability adds
	// mana off a mana ability and does not target, it belongs here. A
	// Triggered entry would go on the stack and give both players a
	// priority window before the mana arrived, which is exactly what
	// CR 605.4a forbids — and by then the spell it was meant to pay
	// for has already been paid for. An "add mana" trigger that fires
	// on a CAST or an ATTACK is an ordinary stack trigger (CR 605.5a)
	// and stays in Triggered.
	//
	// Build them with the constructors in mana_triggers.go
	// (WheneverEnchantedLandTapsForMana and friends). See
	// [ADR 0074](../../../../docs/decisions/0074-triggered-mana-abilities.md)
	// and AGENTS.md §7. Added by #763.
	ManaTriggers []game.ManaTrigger

	// TriggerDoublers are CR 603.2d effects that add one instance to
	// a matching triggered ability when it is harvested. The game
	// package owns the query and applies the predicates at trigger
	// time; the catalog only declares the card's printed condition.
	TriggerDoublers []game.TriggerDoubler

	// Modes is the S20 sub-PR 4 modal-spell clause ("Choose one —").
	// Each option carries its oracle bullet and, when the bullet
	// targets, its own TargetSpec; the engine derives the cast's
	// effective target clause from the chosen options and the client
	// shows a mode picker before targeting. OnResolve reads the
	// choice back with ctx.HasMode(i). Build it with ChooseOne /
	// ChooseN in modes.go. Nil for non-modal cards.
	//
	// Sub-PR 4 limit: at most one chosen option may target, so a
	// "choose two" card may carry targets on at most one option
	// (Register panics otherwise). Per-mode target slots are the
	// deferred multi-target work.
	Modes *game.ModeSpec

	// AdditionalCost is the S21 sub-PR 5 "As an additional cost to
	// cast this spell, …" clause (CR 601.2f). Only the discard shape
	// exists today: Thrill of Possibility, Big Score, Unexpected
	// Windfall. The caster picks the cards in a client prompt that
	// opens BEFORE targeting, matching the order costs are paid in,
	// and the engine pays the cost with the spell already on the
	// stack so discard payoffs (Mary Read and Anne Bonny, Marauding
	// Mako) trigger above it. Build it with DiscardCost(n). Nil for
	// cards with no additional cost.
	AdditionalCost *game.AdditionalCost

	// OptionalCosts are the additional costs the caster may CHOOSE to
	// pay while announcing the spell (CR 601.2b) — kicker
	// (CR 702.33), multikicker (CR 702.33d), buyback (CR 702.27).
	// ADR 0073.
	//
	//	OptionalCosts: []game.AdditionalCost{Kicker("{4}")},                       // Burst Lightning
	//	OptionalCosts: []game.AdditionalCost{Multikicker("{G}")},                  // Wolfbriar Elemental
	//	OptionalCosts: []game.AdditionalCost{Buyback("{3}")},                      // Capsize
	//	OptionalCosts: []game.AdditionalCost{BuybackSacrifice("a land", MatchLand)}, // Constant Mists
	//	OptionalCosts: []game.AdditionalCost{KickerSacrifice("a creature", Creature())}, // Gatekeeper of Malakir
	//
	// A SEPARATE slot from AdditionalCost, not a widening of it,
	// because an optional cost has an INDEX: the announcement names
	// positions in this slice, and so does the record the engine
	// keeps of what was paid. Register cross-checks the two slots, so
	// an entry here without Optional set — or an Optional cost in the
	// mandatory slot — fails at boot rather than silently.
	//
	// Build the entries with the keyword constructors in
	// additional_cost.go, never by hand: each one carries the Key the
	// engine reads (ctx.WasKicked, and buyback's return-to-hand
	// route), and a hand-rolled game.AdditionalCost{Optional: true}
	// compiles and does nothing.
	//
	// OnResolve reads the choice back with ctx.WasKicked() /
	// ctx.KickedTimes(), or ctx.OptionalCostTimes(key) for a cost
	// with another name. A permanent's own "when this enters, if it
	// was kicked" trigger reads game.CardKickedTimes(*source).
	OptionalCosts []game.AdditionalCost

	// CantBeCountered is the S23 "This spell can't be countered"
	// rider (Supreme Verdict). A spell that declares it is still a
	// legal target for Counterspell — the counter resolves and does
	// nothing (CR 701.6a), which is a different and observable thing
	// from the counterspell fizzling.
	//
	// Only a card's OWN printed rider belongs here. A GRANT
	// ("creature spells you control can't be countered", Cavern of
	// Souls) is a continuous effect over the stack and the layer
	// system does not reach the stack; see
	// server/internal/game/cant_be_countered.go.
	CantBeCountered bool

	// AlternativeCosts is the S22 "you may cast this spell for its
	// overload / evoke / cleave cost" clause (CR 118.9) — a cost paid
	// INSTEAD of the mana cost, not alongside it like AdditionalCost.
	// Build the entries with Overload / Evoke / Cleave in
	// alternative_cost.go, never by hand: each keyword bundles a
	// text rewrite with its price, and the rewrite is the half a
	// card file would forget.
	//
	//	AlternativeCosts: []game.AlternativeCost{Overload("{4}{R}")},
	//
	// A slice because a card can offer more than one (spree, the
	// modal-cost cards). Nil for nearly every card. OnResolve reads
	// the choice back with ctx.PaidAltCost("overload").
	AlternativeCosts []game.AlternativeCost

	// CastCondition is "you may cast this spell only if …" — CR
	// 307.6's legendary sorcery ("only if you control a legendary
	// creature or planeswalker"), and the "cast only if" family
	// generally. Checked once, at announce, by the one cast gate
	// (ADR 0073 §7), and never at resolution: a condition that
	// stopped holding while the spell was on the stack does not
	// counter it.
	//
	//	CastCondition:      LegendarySorcery(),                    // Urza's Ruinous Blast
	//	CastConditionLabel: LegendarySorceryLabel,
	//
	// CastConditionLabel is the clause as printed and is REQUIRED
	// with it: the refusal carries the label to the client, and a
	// condition with no label produces a toast that says nothing.
	// Register panics on either without the other.
	//
	// Contract, the same one an ability's Condition has: read-only,
	// runs under g.mu (use *ForEffect accessors), and reads only
	// public information, because the answer reaches every viewer as
	// `cant_cast` on the card.
	CastCondition      func(g *game.Game, controller uuid.UUID, card game.Card) bool
	CastConditionLabel string

	// CastRestrictions are the "can't cast" statics this PERMANENT
	// imposes (CR 101.2) — Rule of Law's "each player can't cast more
	// than one spell each turn", Grafdigger's Cage's "players can't
	// cast spells from graveyards or libraries", Rakdos, Lord of
	// Riots' "you can't cast creature spells unless an opponent lost
	// life this turn".
	//
	// Read from the BATTLEFIELD through CatalogAbilityKey, like a
	// cost modifier and for the same reasons: a permanent that has
	// lost its abilities stops restricting, one whose designation
	// gate is unsatisfied is not there at all, and nothing is stored
	// so the source leaving lifts the restriction on the next query.
	// Build with the constructors in cast_restriction.go.
	CastRestrictions []game.CastRestriction

	// ActivationRestrictions are the "can't be activated" statics
	// this PERMANENT imposes on other objects' activated abilities
	// (CR 602.5a) — Cursed Totem's "activated abilities of creatures
	// can't be activated", Linvala's "…of creatures your opponents
	// control…", Collector Ouphe's "…of artifacts…", Pithing
	// Needle's "…of sources with the chosen name … unless they're
	// mana abilities".
	//
	// Read from the BATTLEFIELD through CatalogAbilityKey, like a
	// cast restriction and for the same reasons: a permanent that
	// has lost its abilities stops restricting, one whose
	// designation gate is unsatisfied is not there at all, and
	// nothing is stored so the source leaving lifts the restriction
	// on the next query. Build with the constructors in
	// activation_restriction.go. #1210, ADR 0073's amendment of
	// 2026-09-22.
	ActivationRestrictions []game.ActivationRestriction

	// TapCost is the S22 "tap permanents you control to help pay"
	// cost component — convoke (CR 702.51) and waterbend, which are
	// the same mechanic under two names. Unlike the other cost slots
	// this one does not add a demand, it SPENDS against one: each
	// permanent tapped pays for {1}, or (convoke only) for one mana
	// of that permanent's colour.
	//
	// Build it with Convoke() or Waterbend("{X}") in tap_cost.go,
	// never by hand: the keyword carries the pool of legal permanents
	// and the colour rule with it, and a card file that got the
	// colour rule wrong would ship a card stronger than printed.
	//
	//	TapCost: Convoke(),
	//	TapCost: Waterbend("{X}"),
	//
	// Nil for nearly every card. The caster's picks ride cast_spell
	// as `tap_ids`; tapping nothing is always legal.
	TapCost *game.TapPermanentsCost

	// CostModifiers is the S28 "spells cost {N} more / {N} less to
	// cast" static (CR 601.2f) the card contributes while it is on
	// the battlefield — Sphere of Resistance, Thalia, Goblin
	// Electromancer, Heartless Summoning, Trinisphere.
	//
	// Deliberately NOT a `Static` entry, for the same reason
	// NoMaxHandSize isn't. The CR 613 layer engine models continuous
	// effects that change a CHARACTERISTIC OF AN OBJECT, and
	// game.StaticAbility's Apply signature is exactly that shape —
	// a *Characteristic and a target *Card. A cost modifier changes
	// neither: it changes what someone PAYS to cast something that
	// is not on the battlefield and has no Characteristic at all.
	// Mana value is explicitly untouched (CR 202.3), so there is
	// no layer for it to sit in.
	//
	// So the engine derives it instead, exactly the way it derives
	// the hand-size answer: the cast path asks the battlefield for
	// every modifier in play and prices the spell through them in
	// CR 601.2f order. Nothing is written anywhere, so nothing has
	// to be unwound when the permanent leaves.
	//
	// Build the entries with CostsMore / CostsLess / CostsAtLeast
	// in cost_modifier.go. Nil for nearly every card.
	CostModifiers []game.CostModifier

	// SelfCostModifiers is "THIS spell costs {N} more / less to cast"
	// (CR 601.2f, 113.6d): Blasphemous Act's "{1} less for each
	// creature on the battlefield", affinity (CR 702.41a), Ghalta's
	// "{X} less where X is the total power of creatures you control",
	// Fireball's "{1} more for each target beyond the first".
	// ADR 0048 addendum §11.
	//
	// The slot, not the constructor, is what makes a modifier apply
	// to its own spell. Write the entries with the same CostsLess /
	// CostsLessEach / CostsMore constructors CostModifiers uses, plus
	// AffinityFor, CostsLessIfItTargets and
	// CostsMorePerTargetBeyondFirst:
	//
	//	SelfCostModifiers: []game.CostModifier{
	//	    AffinityFor("Affinity for artifacts", Artifact()),
	//	},
	//
	// Read only while this card is being priced, from whatever zone it
	// is cast from, for the face being cast; never read from the
	// battlefield. A card can have both slots — Mycosynth Golem has
	// affinity itself and grants it from the battlefield. A CostFloor
	// here panics at Register: no printed card sets a floor on its own
	// cost. Nil for nearly every card.
	SelfCostModifiers []game.CostModifier

	// ExhaustPermissions is "you may activate exhaust abilities as
	// though they haven't been activated" (#1184, CR 609.4) — Elvish
	// Refueler, the one printed card that reads the exhaust record and
	// then tells one player to ignore it.
	//
	// Build the entries with MayActivateExhaustAbilitiesAgain in
	// exhaust_permission.go; the predicate is the printed condition
	// and nothing else:
	//
	//	ExhaustPermissions: []game.ExhaustPermission{
	//	    MayActivateExhaustAbilitiesAgain(
	//	        "During your turn, as long as you haven't activated an "+
	//	            "exhaust ability this turn, you may activate exhaust "+
	//	            "abilities as though they haven't been activated.",
	//	        DuringTheControllersTurn(), ControllerHasActivatedNoExhaustAbilityThisTurn()),
	//	},
	//
	// It suspends the GATE and nothing else: the ability still costs
	// what it costs, still checks its Condition, and activating it
	// still writes the record. Read from the battlefield only
	// (CR 113.6). Nil for every other card.
	ExhaustPermissions []game.ExhaustPermission

	// AttackTaxes is the CR 508.1a attack tax: "creatures can't attack
	// you unless their controller pays {2} for each creature they
	// control that's attacking you" — Propaganda, Ghostly Prison,
	// Windborn Muse, Sphere of Safety. ADR 0080.
	//
	// A static on the DEFENDER's side, read when the attacking player
	// declares and charged as one announce-time payment on the whole
	// declaration. The "you" is the permanent's own controller and is
	// structural, so a card cannot accidentally tax the table.
	//
	// Build the entries with AttackTax / AttackTaxCounting in
	// attack_tax.go. Nil for nearly every card.
	AttackTaxes []game.AttackTax

	// CastableZones is the S29 "you may cast this card from
	// somewhere other than your hand" declaration (CR 601.2, and
	// every keyword in CR 702 that grants an alternative cast
	// path). Nil — nearly every card — means hand only.
	//
	// The zone is only the PLACE. The PRICE rides
	// AlternativeCost.FromZone, and the two are declared together:
	//
	//	CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
	//	AlternativeCosts: []game.AlternativeCost{Flashback("{2}{R}")},
	//
	// Register panics on a zone-bound cost whose zone is not listed
	// here, because such an offer can never be claimed.
	//
	// Hand is implicit and need not be listed — declaring the
	// graveyard ADDS a path, it never removes the ordinary cast.
	// Listing ZoneCommand is pointless (CR 903.4 grants that one to
	// the format, not the card) and listing ZoneExile is for cards
	// whose own text grants the permission; the impulse-exile /
	// airbend / madness family grants it to a single exiled
	// instance instead, through game.CastPermission.
	CastableZones []game.ZoneKind

	// SpecialActions are the CR 116.2 special actions this card
	// offers from its owner's hand — foretell (CR 702.143) and
	// suspend (CR 702.62). A special action does NOT use the stack
	// and is not an ability, which is why it is its own slot rather
	// than an entry in Activated or an AlternativeCost.
	//
	// Declare them with the keyword constructors, never by hand:
	//
	//	SpecialActions: []game.SpecialAction{Foretell("{1}{U}")},
	//	SpecialActions: []game.SpecialAction{Suspend(1, "{R}")},
	//
	// Register refuses a kind the engine cannot carry out, and a
	// suspend declaring no time counters, at boot.
	SpecialActions []game.SpecialAction

	// Madness is the card's madness cost (CR 702.35), as printed:
	//
	//	Madness: "{R}",   // Fiery Temper
	//	Madness: "{B}",   // Big Game Hunter
	//	Madness: "{0}",   // Basking Rootwalla — a real free cost
	//
	// One string, because that is the only thing a madness card says
	// that another madness card does not. Both halves of the keyword
	// — the CR 702.35a discard replacement and the exile-zone trigger
	// that offers the cast — are grown from this by buildDef
	// (game.MadnessReplacement, game.MadnessTrigger), so no card file
	// writes either and none can forget one. Register refuses an
	// unparseable cost at boot.
	//
	// Empty — every card but a handful — means the card has no
	// madness.
	Madness string

	// Activated is the list of CR 602 activated abilities the card
	// offers from the battlefield — the fourth ability type, added
	// in S21 sub-PR 2. Each entry declares its cost (tap, sacrifice
	// this / sacrifice another matching a spec, mana, life), an
	// optional target clause validated like a spell's, and the
	// Effect that runs when the ability resolves off the stack.
	//
	// Mana abilities do NOT belong here — they don't use the stack
	// (CR 605.3b) and keep their own ManaAbilities slot.
	//
	// Build the entries with the constructors in activated.go:
	//
	//	Activated: []ActivatedAbility{{
	//		Label: "Sacrifice a creature: This enchantment deals 1 damage to any target.",
	//		Cost:  SacrificeACreature(),
	//		Targets: TargetAny(),
	//		Effect: ...,
	//	}},
	Activated []ActivatedAbility

	// NoMaxHandSize declares the printed static "You have no maximum
	// hand size" (Thought Vessel, Reliquary Tower, Spellbook,
	// Venser's Journal). True while the permanent is on the
	// battlefield; the controller skips the CR 402.2 cleanup-step
	// discard entirely.
	//
	// This is deliberately NOT a `Static` entry. Every other
	// continuous effect in the catalog modifies a characteristic of
	// an OBJECT, which is what the CR 613 layer engine models —
	// game.StaticAbility's Apply takes a *Characteristic and a
	// target *Card, and there is no seat in that signature. "You
	// have no maximum hand size" modifies a PLAYER, so it has no
	// characteristic to sit in and no layer to sit at.
	//
	// Rather than grow a parallel player-layer pipeline for one
	// clause, the engine DERIVES the answer: at cleanup it asks the
	// battlefield whether the active player controls any permanent
	// with this bit set (game.Game.EffectiveMaxHandSizeLocked, fed
	// by the game.CatalogNoMaxHandSize hook). Nothing is written to
	// Player.MaxHandSize, so nothing has to be restored when the
	// permanent leaves — which is what makes two copies, and one of
	// two leaving, come out right without any bookkeeping.
	//
	// Issue #338.
	NoMaxHandSize bool

	// PlayerKeywords declares a printed static that gives this
	// permanent's CONTROLLER an ability — "You have hexproof"
	// (Leyline of Sanctity, Aegis of the Gods), "You have protection
	// from everything" if a permanent ever prints it. True while the
	// permanent is on the battlefield and nowhere else.
	//
	//	PlayerKeywords: []string{"hexproof"},
	//
	// Engine ability TOKENS, in the vocabulary keywords.go and
	// protection.go already parse — a protection token is built with
	// game.ProtectionFromColor or spelled with the constants in
	// protection.go, never by hand, for the reason that file gives: a
	// token the closed grammar cannot parse grants nothing at all, so
	// a typo ships a card that looks finished and does nothing.
	//
	// NOT a `Static` entry, for exactly the reason NoMaxHandSize
	// above is not: game.StaticAbility's Apply takes a
	// *Characteristic and a target *Card, and a player is neither.
	// The engine derives the answer instead — it asks the battlefield
	// on every query, through the game.CatalogPlayerKeywords hook —
	// so two Leylines compose and one of them leaving cannot revoke
	// the other's grant. The GRANTED half of the same rule, "you gain
	// protection from everything until your next turn", is stored
	// instead and lives on game.Player.Statics.
	//
	// Three consumers read it and there is no fourth: targeting
	// (CR 702.11d / 702.16i), the CR 702.16e damage built-in, and
	// the CR 702.16c attachment check for an "enchant player" Aura.
	//
	// Issue #1197, ADR 0072's 2026-09-22 amendment.
	PlayerKeywords []string

	// PlayerLifeTotalLocked declares the printed static "Your life
	// total can't change" (CR 119.7, CR 119.8 — Platinum Emperion),
	// about this permanent's CONTROLLER. True while the permanent is
	// on the battlefield and nowhere else.
	//
	//	PlayerLifeTotalLocked: true,
	//
	// NOT a `Static` entry, and DERIVED rather than written, for
	// exactly the reasons NoMaxHandSize and PlayerKeywords above give:
	// game.StaticAbility's Apply takes a *Characteristic and a target
	// *Card and a player is neither, and a "set on enter, restore on
	// leave" design cannot answer "restore to what?" when a second
	// copy is out. The engine asks the battlefield on every query,
	// through the game.CatalogPlayerLifeTotalLocked hook.
	//
	// The GRANTED half of the same rule — "until your next turn, your
	// life total can't change" (Teferi's Protection, Teferi's
	// Reproach) — is stored instead, on game.Player.Statics, and is
	// written from a card file with the LockLifeTotal primitive.
	//
	// Issue #1200, ADR 0085.
	PlayerLifeTotalLocked bool

	// WantsDistinctColors declares a spell that READS the colours of
	// the mana that paid for it: converge (CR 702.86 — Painful
	// Truths, Bring to Light) and sunburst (CR 702.44 — Etched
	// Oracle). Nothing else in the game does (#761).
	//
	// What it changes is the PAYMENT, not the effect: the cast gate
	// pays the generic half of the cost with colours it has not spent
	// yet instead of the usual colourless-first order, so a Painful
	// Truths cast out of a five-colour pool converges for five rather
	// than for two. The effect itself reads ctx.ColorsSpent() and
	// does not care how the mana got there.
	//
	// A DECLARATION rather than something inferred from the oracle
	// text, for the reason DerivesFromOtherSources is one: a text
	// scan quietly stops matching when a card words the clause
	// differently, and a converge spell that counts one colour fails
	// silently.
	//
	// Adamant (CR 207.2c — "at least three red mana") deliberately
	// does NOT set it. Spreading colours is the opposite of what
	// adamant wants, and concentrating them is a different strategy
	// again; the honest answer for now is that adamant reads what the
	// player happened to spend.
	WantsDistinctColors bool

	// WantsManaFrom declares the kinds of mana SOURCE this card's own
	// text reads back — game.ManaSourceTreasure for "if mana from a
	// Treasure was spent to cast it" (Hired Hexblade, Jaded
	// Sell-Sword, Devour Intellect), ManaSourceCreature for Inga and
	// Esika's "three or more mana from creatures" (#1212).
	//
	// Like WantsDistinctColors above it changes the PAYMENT and not
	// the effect, and it changes it even more softly: the auto-tapper
	// prefers a matching source when it has a free choice, as a
	// tiebreak after the frozen and restrictiveness orderings and
	// never as a filter. The set of sources it may plan is untouched,
	// so a cast that was payable stays payable and one that was not
	// stays not — see game.autoTapPreferringLocked.
	//
	// It is not the reader. The card still asks
	// ctx.ManaSpent().FromTreasure() (or ManaSpentToCastThis() from
	// an enters trigger) and gets the truth: a Hired Hexblade whose
	// controller tapped two Swamps by hand draws no card, wish or no
	// wish.
	//
	// Zero for every card that does not read its payment's sources,
	// which is all but about twenty of them.
	WantsManaFrom game.ManaSourceKinds

	// AdditionalLandPlays declares the printed static "you may play
	// an additional land on each of your turns" — 1 for Exploration,
	// 2 for Azusa, Lost but Seeking. Counted while the permanent is
	// on the battlefield; its controller's land-play allowance rises
	// by this much (CR 305.2).
	//
	// Deliberately NOT a `Static` entry, for exactly the reasons
	// NoMaxHandSize above is not: it modifies a PLAYER, not an
	// object, so the CR 613 layer engine has no characteristic for
	// it and no layer for it to sit at. The engine DERIVES it
	// instead — game.Game.EffectiveLandDropsLocked sums this over the
	// permanents a player controls, through the
	// game.CatalogAdditionalLandPlays hook — so two Explorations
	// compose and one of them leaving does not take the other's
	// grant with it.
	//
	// Added by #500, which made the land-drop limit enforceable at
	// all. No card sets it yet; Exploration and Azusa are now a
	// one-line Spec each rather than an engine change.
	AdditionalLandPlays int

	// CastPermissions declares the STANDING cast and play permissions
	// this permanent grants its controller while it is on the
	// battlefield (ADR 0066) — "each nonland card in your graveyard
	// has escape" (Underworld Breach), "you may play lands and cast
	// spells from the top of your library" (Bolas's Citadel).
	//
	// Scope is forced to game.ScopeStanding and the window to "while
	// the source remains", because that is what a permanent's static
	// ability means: the engine re-derives these from the battlefield
	// on every query, so two Underworld Breaches compose, one of them
	// leaving does not revoke the other's permission, and there is no
	// duration to expire. A per-INSTANCE permission (Snapcaster's
	// flashback for one card, impulse exile) is granted by an EFFECT
	// instead — game.Game.GrantCastPermissionOverCardForEffect.
	//
	// Narrow which cards qualify with game.PermissionFilter, price the
	// cast with AltCostKey / Cost / LifeEqualToManaValue /
	// ExileOtherFromGraveyard, and set TopOfLibraryOnly for a library
	// permission (CR 401.5). A library permission ALSO needs
	// LibraryTopVisible below: a card you cannot see is a card you
	// cannot play, and every printed card carries both halves.
	CastPermissions []game.CastPermission

	// CastTimings declares the per-player cast-TIMING statements this
	// permanent makes while it is on the battlefield (#1195, ADR 0066's
	// 2026-09-22 amendment) — "you may cast spells as though they had
	// flash" (Vedalken Orrery, Leyline of Anticipation), "you may cast
	// creature spells as though they had flash" (Yeva), "each opponent
	// can cast spells only any time they could cast a sorcery" (Teferi,
	// Time Raveler).
	//
	// Build one with CastAsThoughFlash / OpponentsCastAtSorcerySpeed
	// and friends in cast_timing.go rather than by hand: the
	// constructors carry the Affects clause and the printed label,
	// which are the two halves a card file gets wrong.
	//
	// The window is forced to "while the source remains" for the
	// reason CastPermissions above is: a permanent's static ability is
	// re-derived from the battlefield on every query, so two Orreries
	// compose and one leaving cannot revoke the other's. A statement
	// that OUTLIVES its source — Emergence Zone's "this turn", Teferi's
	// +1 — is granted by an EFFECT instead, with GrantCastTiming.
	CastTimings []game.CastTimingRule

	// LibraryTopVisible declares the printed clause that makes this
	// permanent's controller's top library card visible (CR 401.5) —
	// game.LibraryTopOwner for "you may look at the top card of your
	// library any time" (Realmwalker, Bolas's Citadel),
	// game.LibraryTopRevealed for "play with the top card of your
	// library revealed" (Oracle of Mul Daya, Courser of Kruphix).
	//
	// Deliberately NOT a `Static` entry, for the reason
	// AdditionalLandPlays above is not: it is a fact about a PLAYER
	// and a zone position, not a characteristic of an object, so the
	// CR 613 layer engine has nowhere to put it. Derived from the
	// battlefield on every query instead (game.LibraryTopVisibilityLocked),
	// which also means the answer is always about whatever is on top
	// NOW — no library mutation has to invalidate anything.
	LibraryTopVisible game.LibraryTopVisibility

	// UntapStep declares the printed clause "untap <these> during
	// each other player's untap step" — Seedborn Muse, Unwinding
	// Clock, Drumbellower, Bender's Waterskin, and the second half
	// of Quest for Renewal.
	//
	// This is deliberately NOT a `Triggered` entry, and the
	// distinction is the whole point of the field. Nothing about
	// this clause uses the stack: the untap step grants no priority
	// (CR 502.4), so there is no announce, no response window and
	// nothing to counter. It is a modification of the untap step's
	// TURN-BASED ACTION — CR 502.3's "the active player determines
	// which permanents they control untap" — and the permission
	// widens that set.
	//
	// Written as a trigger instead, the card fires at the following
	// upkeep, a step late, on the stack, where a tap effect in
	// response leaves the permanents tapped. Quest for Renewal
	// shipped exactly that with a caveat saying so; this field is
	// what let the caveat go.
	//
	// It is also NOT a `Static` entry, for the reason NoMaxHandSize
	// isn't: game.StaticAbility modifies a CHARACTERISTIC of an
	// object, and "these permanents untap during that step" is not
	// one — it changes what a turn-based action does, which has no
	// layer to sit in. Like NoMaxHandSize, the answer is DERIVED at
	// the moment the step asks for it, so two Seedborn Muses and one
	// of them dying need no bookkeeping at all.
	//
	// Nil for every card that does not print the clause, which is
	// nearly all of them. Issue #74.
	UntapStep []game.UntapStepPermission
	// UntapStepRestrictions declares permanents that stay tapped during
	// their controller's untap step (Mana Vault, Meekstone, and Auras).
	UntapStepRestrictions []game.UntapStepRestriction
	// UntapCaps declares "players can't untap more than N <kind>
	// during their untap steps" — Winter Orb, Static Orb, Winter Moon
	// (#826, CR 502.3). A ceiling, not a restriction: when more
	// permanents are eligible than the cap allows, the active player
	// is asked which ones untap. See ADR 0070 and untap_choice.go.
	UntapCaps []game.UntapCap
	// UntapOptOuts declares "you may choose not to untap this during
	// your untap step" — Rust Tick, Amber Prison (#826, CR 502.3).
	// Such a permanent joins the same prompt the caps raise, exempt
	// from the rule that untapping is otherwise mandatory.
	UntapOptOuts []game.UntapOptOut

	// Completeness declares how faithfully this spec implements the
	// card as printed — the machine-readable form of the prose
	// "declared simplification" convention in AGENTS.md §7. See
	// completeness.go for the full contract and for why the zero
	// value is CompletenessUnreviewed rather than
	// CompletenessFull.
	//
	// Set it when you add or change a card. Leaving it unset is
	// permitted and is not a failure — it publishes the card as
	// unaudited, which is true.
	// Emblem is the emblem this card's abilities create (CR 114) —
	// "You get an emblem with [ability]". Nil for every card that
	// makes none, which is nearly all of them.
	//
	// Declared once here, next to the ability that creates it; the
	// ability itself is `CreateEmblem{}.Apply(ctx)` and names
	// nothing, because the emblem it makes is this one. Register
	// files the emblem's own CardDef under game.EmblemKey(OracleID),
	// which is how its statics reach the layer pass and its triggers
	// reach the harvester. See emblem.go and ADR 0064.
	Emblem *EmblemSpec

	// Grants are the ability bundles this card's COPY effect can give
	// the copy — "except … it has '<ability>'" (CR 707.9a). Nil for
	// every card that grants none, which is nearly all of them.
	//
	// Declared here, next to the except clause that names one, for
	// the same reason Emblem is: the abilities are static catalog
	// data, and what the copy carries is only the bundle's Key.
	// Register files each bundle's own CardDef under
	// game.GrantKey(Key). See ability_grant.go and
	// server/internal/game/copy_grants.go.
	Grants []AbilityGrant

	Completeness Completeness

	// Caveats names the printed clauses this spec does NOT model,
	// one short player-facing sentence each — "Cycling is not
	// implemented; the land can only be played." Required when
	// Completeness is CompletenessCaveats and rejected otherwise,
	// because a caveat nobody can read is the same as no caveat at
	// all.
	//
	// Write for a player deciding whether to sleeve the card, not
	// for the next engineer: the engineering reason belongs in the
	// file's doc comment, where there is room for it.
	Caveats []string
	// Battle is a battle's printed battle data — its defense and its
	// subtype (CR 310). Nil for every card that is not a battle,
	// which is nearly all of them.
	//
	// A FALLBACK, like StartingLoyalty: the printed value on
	// game.Card.StartingDefense, stamped by the deck importer from
	// Scryfall's per-face `defense`, always wins. See BattleSpec in
	// battles.go for why it is declared anyway.
	//
	// Added in S27.
	Battle *BattleSpec

	// XMatters declares that everything this card does scales with
	// the announced X (CR 601.2b / 602.2b), so an announcement of
	// X=0 does nothing at all: Fireball deals no damage, Soothsaying
	// looks at no cards, Treasure Vault makes no Treasures.
	//
	// It is read by ONE rule, in `internal/legal` (the bot's legal
	// enumerator, ADR 0033 §1): a move whose whole effect is X is
	// not offered at X=0. The engine is unaffected — CR 602.2b makes
	// X=0 a legal announcement and the engine still accepts it; what
	// changes is that the enumerator stops OFFERING an action that
	// does nothing, because a free repeatable no-op is a loop the
	// game does not let run forever (CR 732.2a, #810).
	//
	// Declare it on any card whose resolution reads ctx.X(), and
	// x_matters_guard_test.go fails the build when one does not.
	// The exception it is written to allow is a card with a fixed
	// RIDER — an effect that happens whatever X is — which should
	// leave this unset and say so in its doc comment, because for
	// such a card X=0 is a real move. No card in the catalog is that
	// shape today.
	//
	// Card-level rather than per-ability on purpose: the rule only
	// fires for a cost that actually carries an {X} slot, so the
	// abilities of a card that has both (Soothsaying's {3}{U}{U}
	// shuffle and its {X} look) are never confused by one flag.
	XMatters bool
}

// ActivatedAbility is one activated ability on a permanent. Mirrors
// game.ActivatedAbilityShape; the wire hook converts. Added in S21
// sub-PR 2.
type ActivatedAbility struct {
	Label   string
	Cost    game.AbilityCost
	Targets *game.TargetSpec
	// Modes is the CR 700.2 mode clause of a modal activated ability
	// ("{4}, {T}: Choose one —"). The same game.ModeSpec a modal
	// spell declares in Spec.Modes, built with the same ChooseOne /
	// ChooseN constructors (#764, ADR 0065 §3). Modes and targets are
	// announced together at activation (CR 602.2b); each chosen
	// bullet's ModeOption.Effect runs at resolution in announce
	// order. Declare the target clause on the OPTION, not here.
	Modes        *game.ModeSpec
	SorcerySpeed bool
	// Zones is the set of zones this ability functions from
	// (CR 113.6). Nil — nearly every ability — means the
	// battlefield. Cycling declares ZoneHand; a graveyard activation
	// (Reassembling Skeleton) will declare ZoneGraveyard. Build the
	// entry with Cycling / Typecycling rather than setting this by
	// hand. See game.ActivatedAbilityShape.Zones and ADR 0062.
	Zones []game.ZoneKind
	// Cycling marks the card's cycling ability (CR 702.29a), so
	// activating it emits EventCycle. Set by the Cycling /
	// Typecycling constructors; no card file sets it directly.
	Cycling bool
	// Condition is the "Activate only if …" / "Activate only during
	// your turn" gate (CR 602.1b, #743). Same contract and helpers as
	// ManaAbility.Condition — see game.ActivatedAbilityShape.Condition
	// and activation_conditions.go. Nil means no condition.
	Condition func(g *game.Game, controller, source uuid.UUID) bool
	// ActiveWhen is the CR 716 / 719 / 721 designation gate (ADR
	// 0071): this ability exists only while the permanent is at that
	// level, is solved, or has that many charge counters. Build it
	// with Level / Solved / AtChargeCounters in designations.go. The
	// zero value is "no gate", which is every ability in the catalog
	// but a handful.
	//
	// Distinct from Condition: a Condition greys an ability the
	// permanent HAS, a gate means it is not there at all.
	ActiveWhen game.Designation
	// Exhaust marks an exhaust ability — "Exhaust — {4}: Earthbend 4.
	// (Activate each exhaust ability only once.)" One bit, no card
	// logic: the engine keys the record by (object, this ability's
	// Label) and refuses a second activation itself. See
	// game.ActivatedAbilityShape.Exhaust and ADR 0020's exhaust
	// addendum (#1181).
	//
	// ManaAbility deliberately has no twin of this field: the mana
	// path does not write the activation record, so the combination
	// is unspellable rather than silently ignored.
	Exhaust bool
	Effect  func(g *game.Game, item *game.StackItem) error
}

// ManaAbility is one mana-producing activated ability on a permanent.
// Cost expresses what the controller pays to activate (today: tap +
// optional sacrifice; future: pay-life, sub-mana). Produced is the
// mana the ability adds to the controller's pool; uses the same
// brace-notation grammar the Scryfall mana_cost field does, plus a
// pipe (`|`) inside a single brace pair to mean "controller picks
// one of these colors" — Birds of Paradise prints `"{W|U|B|R|G}"`.
// Label is the menu copy the client renders ("Add {C}{C}",
// "Add one mana of any color"); empty falls back to a generated
// label.
type ManaAbility struct {
	Cost     ManaAbilityCost
	Produced string
	Label    string

	// Exhaust marks an exhaust mana ability — "Exhaust — {G}, {T}:
	// Add three mana of any one color. (Activate each exhaust ability
	// only once.)" (#1183). The twin of ActivatedAbility.Exhaust, and
	// one declarative bit for the same reason: the keyword IS the
	// rule, and a card that wrote its own "have I done this yet"
	// check would be writing a rule the engine enforces in five
	// places anyway.
	//
	// Loot, the Pathfinder is the one printed card that wants it, and
	// prints exhaust three times — once here and twice on ordinary
	// activated abilities, which is exactly why the record is keyed by
	// the ability's LABEL and not by the permanent.
	//
	// Register enforces the same two rules it enforces on the
	// activated list, ACROSS BOTH LISTS: the label must print the
	// keyword (and a label that prints it must set the bit), and no
	// two exhaust abilities on one card may share a label — they would
	// share one use.
	//
	// See game.ManaAbilityShape.Exhaust and ADR 0020's exhaust
	// addendum.
	Exhaust bool

	// Rider is everything the oracle text says AFTER the "Add …"
	// clause, as one callback: the painland cycle's "This land deals
	// 1 damage to you", Ancient Tomb's "deals 2 damage to you". It
	// runs immediately after the produced mana lands in the pool,
	// inside the same atomic mana-ability resolution (CR 605.3b).
	//
	// Build one with PainRider(n) rather than by hand — that helper
	// is the whole reason this slot exists so far.
	//
	// A rider is NOT a cost: it happens whether or not the player
	// could "afford" it, and a source with a damage rider stays
	// activatable at 1 life. Use ManaAbilityCost.Life for a real
	// cost ("{T}, Pay 1 life:").
	//
	// Runs under the resolution write lock — *ForEffect helpers
	// only, never a public locking mutator.
	//
	// Added in the S22 mana-ability-rider pass.
	Rider func(g *game.Game, controller, source uuid.UUID) error

	// NarrowToCommanderIdentity intersects a pipe-syntax Produced
	// string ("{W|U|B|R|G}") with the controller's commander colour
	// identity before the colour pick is offered.
	//
	// Set it ONLY when the printed text says "any color in your
	// commander's color identity" — Command Tower, Arcane Signet,
	// Commander's Sphere, Path of Ancestry.
	// TestNarrowToCommanderIdentityMatchesOracleText holds the
	// catalog to exactly that. Every other pipe — Birds of Paradise,
	// Treasure, City of Brass, the painland and guildgate duals —
	// leaves it off and offers its printed width, with the
	// commander's identity listed first (owner decision 2026-09-17).
	//
	// CR 903.4f (#844): a narrowing ability adds NO mana for a
	// controller with no commander, or a colourless one — the
	// intersection is empty, no pick is offered, and nothing
	// enumerates the activation.
	//
	// Replaced IgnoreCommanderIdentity (S22), its inverse, when the
	// default flipped from narrowing to ordering.
	NarrowToCommanderIdentity bool

	// ProducedFunc computes Produced at activation time instead of
	// declaring it. Two card families need it and they are the same
	// mechanism (#352 sub-gaps 3 and 4):
	//
	//   - DERIVED colours: Exotic Orchard ("one mana of any color
	//     that a land an opponent controls could produce"),
	//     Reflecting Pool, Fellwar Stone, Mox Amber. Return a pipe
	//     string — "{W|U|G}".
	//   - SCALED amounts: Cabal Coffers ("{B} for each Swamp you
	//     control"), Gaea's Cradle. Return the slot repeated —
	//     "{B}{B}{B}".
	//
	// Wins over Produced when non-nil. Returning "" adds no mana,
	// which is the printed behaviour when the derivation finds
	// nothing — Exotic Orchard on an empty opposing board, Gaea's
	// Cradle with no creatures. The source still taps.
	//
	// READ-ONLY, and it runs under g.mu: ActivateManaAbility holds
	// the write lock and the auto-tapper holds the read lock. Read
	// g.Battlefield and the *ForEffect accessors; a public locking
	// mutator deadlocks. Build one with ProducedFromLands,
	// ProducedRepeated or a sibling in mana_derivation.go rather
	// than by hand.
	ProducedFunc func(g *game.Game, controller, source uuid.UUID) string

	// ProducedForPaid computes Produced from what the cost actually
	// PAID, for an ability whose printed text derives its output
	// from the payment rather than from the board (#789):
	//
	//	Mage-Ring Network  "Add {C} for each storage counter removed
	//	                    this way"  →  ProducedPerCounterRemoved("{C}")
	//
	// Wins over ProducedFunc, which wins over Produced. It is handed
	// the same game.PaidCost a stack item carries — a mana ability
	// has no stack item (CR 605.3b), so the record lives only for the
	// length of the activation.
	//
	// Same read-only-under-the-lock contract as ProducedFunc. CR
	// 106.7's "could produce" reader evaluates it with the largest
	// payment the source could make right now, so a Network with
	// three counters could produce {C} and one with none could not.
	ProducedForPaid func(g *game.Game, controller, source uuid.UUID, paid game.PaidCost) string

	// DerivesFromOtherSources marks a ProducedFunc that asks OTHER
	// permanents what THEY could produce — Exotic Orchard, Reflecting
	// Pool, Fellwar Stone, and nothing else in the catalog. It is the
	// recursion guard, and it is a declaration rather than something
	// inferred because the alternative is a re-entrancy counter on a
	// snapshotted struct.
	//
	// CR 106.7's "could produce" reader (game.ProducibleManaLocked)
	// evaluates every OTHER ProducedFunc — a chosen colour, a board
	// count, a devotion — and skips these, because two Exotic Orchards
	// facing each other would otherwise recurse until the stack ran
	// out. CR 106.6b answers the circular case with "no mana" and so
	// does the guard.
	//
	// Pair it with ProducedFromOpponentLands / ProducedFromOwnLands
	// and nothing else; TestDerivedManaAbilitiesDeclareTheGuard holds
	// the catalog to that in both directions.
	//
	// Added in S44 (#782).
	DerivesFromOtherSources bool

	// Condition gates activation — "Activate only if you control
	// five or more lands" (Temple of the False God), "…three or
	// more artifacts" (Mox Opal). Checked before any cost is
	// validated, so a failed gate taps nothing and spends nothing
	// (CR 602.5). Same read-only-under-the-lock contract as
	// ProducedFunc.
	//
	// Added in the S32 mana-pipeline pass (#352 sub-gap 5).
	Condition func(g *game.Game, controller, source uuid.UUID) bool

	// Restrictions are the "spend this mana only on …" tags stamped
	// onto every token this ability produces — Ancient Ziggurat,
	// Eldrazi Temple, Shrine of the Forsaken Gods, the coloured half
	// of Delighted Halfling. Build them with the game package's
	// ManaRestrict* constructors; the spend-time matcher lives in
	// game/mana_restriction.go.
	//
	// All tags must hold for the token to be spendable (AND), and an
	// unknown tag denies. A restricted ability drops out of auto-tap
	// planning — see autoTapAbilityFor.
	//
	// Declaring this WITHOUT the engine honouring it would ship
	// every card in the group stronger than printed, which is the
	// #259 rule; the two halves landed together in #352.
	Restrictions []string

	// RestrictionsFunc computes Restrictions at activation time, for
	// an ability whose restriction names a CHOSEN thing rather than a
	// printed one — Cavern of Souls' "spend this mana only to cast a
	// creature spell of the chosen type". Wins over Restrictions when
	// non-nil.
	//
	// Returning nil produces unrestricted mana. A card that cannot
	// compute its restriction yet must return an impossible tag
	// instead, so the mana is unspendable rather than free: weaker
	// than printed is acceptable, stronger is not. Added in S26.
	RestrictionsFunc func(g *game.Game, controller, source uuid.UUID) []string
}

// ManaAbilityCost names the activation cost of one mana ability.
// Tap is the canonical cost ({T}). Sacrifice is reserved for future
// mana rocks like Lotus Petal — declared on the struct so the wire
// shape is stable, NOT exercised by S15's catalog. Mana / life /
// counter sub-costs land with later sprints when a catalog card
// demands them.
type ManaAbilityCost struct {
	Tap bool
	// Sacrifice sacrifices the SOURCE (Treasure, Lotus Petal).
	Sacrifice bool
	// SacrificeOther sacrifices OTHER permanents the activator
	// controls, matched against this spec — Ashnod's Altar's
	// "Sacrifice a creature: Add {C}{C}". Build it with the same
	// constructors an activated ability's cost uses
	// (SacrificeACreature().SacrificeOther, or
	// SacrificeN(n, …).SacrificeOther for "Sacrifice two …", #747),
	// so the two ability kinds share one clause vocabulary and one
	// client picker.
	//
	// Added in the S21 mana-cost pass, which is also what closed
	// the S15 note that sacrifice costs were "reserved for future
	// mana rocks" — they are all live now.
	SacrificeOther *game.TargetSpec

	// Life is a "Pay N life" component of the activation cost (CR
	// 119.4) — Mana Confluence's "{T}, Pay 1 life: Add one mana of
	// any color". Mirrors game.AbilityCost.Life, which CR 602
	// activated abilities have carried since S21 and the fetchlands
	// already use.
	//
	// This is the S15 note above finally coming true: life was the
	// one "later sprints when a catalog card demands them" sub-cost
	// left, and the S22 mana-rider pass is the batch that demanded
	// it.
	//
	// A COST, not a rider: validated before anything is paid, so an
	// activation at a life total below N is rejected outright and
	// the source does not tap. "Add {R}. This land deals 1 damage to
	// you" is the other thing — see ManaAbility.Rider.
	Life int

	// Mana is a mana component of the activation cost — the Signet
	// cycle's "{1}, {T}: Add {W}{U}", Cabal Coffers' "{2}, {T}",
	// the filter lands' "{W/U}, {T}". Scryfall brace grammar.
	//
	// Paid out of the controller's pool before the source taps, and
	// validated with every other component first, so an unaffordable
	// activation fails with the source untouched. No auto-tap: the
	// player floats the mana first, which is how a Signet is played
	// on paper — see ActivateManaAbility.
	//
	// This is the last of S15's "mana / life / counter sub-costs
	// land with later sprints when a catalog card demands them"
	// note. #267 took life; #352 takes mana, and #789 took the
	// counter case — see RemoveCounters below, which finally empties
	// that sentence.
	Mana string

	// RemoveCounters is a "remove N counters" component (#789):
	// Vivid Creek's "{T}, Remove a charge counter from this land",
	// Ramos's "Remove five +1/+1 counters from Ramos", Mage-Ring
	// Network's "Remove any number of storage counters from this
	// land".
	//
	// Build it with the SAME constructors an activated ability's
	// cost uses, reading the component off the returned AbilityCost:
	//
	//	RemoveCountersFromThis("charge", 1).RemoveCounters
	//	RemoveCountersXFromThis("storage", 0).RemoveCounters
	//
	// One game.CounterRemovalCost with two owners, so the validator,
	// the candidate walk, the enumerator, the view and the client's
	// picker are each written once — the same "one clause
	// vocabulary" reasoning SacrificeOther above was built on.
	//
	// The auto-tapper only plans a source whose counter cost it can
	// decide and pay: the self form, a printed kind, a fixed count,
	// and enough counters right now. A Vivid land with no charge
	// counters left is not a five-colour source and is not planned
	// as one.
	RemoveCounters *game.CounterRemovalCost

	// AddCounter is a cost that puts a counter on the source. No
	// printed mana ability has one; the slot exists because the
	// component is declared once and owned by both ability kinds.
	// Build it with AddCounterToThis(kind, n).AddCounter.
	AddCounter *game.CounterAddCost

	// DiscardCards is a "discard N cards" component of the
	// activation cost (#1213) — Skirge Familiar's "Discard a card:
	// Add {B}", the shape `ManaAbilityCost` was missing while
	// `AbilityCost` grew one with #660.
	//
	// Build it with the SAME constructors an activated ability's
	// cost uses, reading the component off the returned AbilityCost:
	//
	//	DiscardACard().DiscardCards
	//	DiscardCardsMatching(1, "a land card", isLand).DiscardCards
	//
	// One game.DiscardCost with two owners, so the validator, the
	// candidate walk, the enumerator, the view and the client's
	// picker are each written once — the same "one clause
	// vocabulary" reasoning SacrificeOther and RemoveCounters above
	// were built on.
	//
	// The auto-tapper never plans a source that has one: which card
	// to pitch is a decision, and the planner makes none.
	DiscardCards *game.DiscardCost
}

// ZeroUUID is an alias for uuid.Nil. Mostly used in tests to
// distinguish "no target / self" fields from "uninitialised" —
// callers don't need this but having it available keeps the
// effects package self-contained for tests that don't import uuid.
var ZeroUUID = uuid.Nil
