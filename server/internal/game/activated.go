package game

import (
	"github.com/google/uuid"
)

// activated.go — S21 sub-PR 2: activated abilities in the catalog
// (CR 602), the fourth ability type after triggered (S19), static
// (S16) and replacement (S17).
//
// Before this, a catalog card's activated ability didn't exist:
// ActivateAbility put a labelled item on the stack with no cost
// check and no effect, and players resolved it in their heads. A
// sacrifice outlet ("Sacrifice a creature: deal 1 damage") can't
// work that way — the cost IS the game action.
//
// The shape mirrors the rest of the catalog surface. A card declares
// its abilities; the engine validates timing and cost, pays the
// cost, and pushes a stack item carrying the same Effect closure
// S19's triggered abilities use. Resolution is therefore already
// written: resolveTopAbilityLocked runs Effect with the live game,
// after the CR 608.2b target re-check.
//
// Mana abilities stay separate (CR 605.3b — they don't use the
// stack) and keep their own ManaAbilityShape path. An ability that
// would be a mana ability but whose cost sacrifices a DIFFERENT
// permanent (Ashnod's Altar) fits neither surface cleanly and is
// deferred; the note is in ADR 0020.

// AbilityCost is what a player pays to activate an ability
// (CR 602.1a). Every field is additive: Goblin Bombardment is
// {SacrificeOther: creature-spec}, Krenko is {Tap: true}, a
// hypothetical "{2}, {T}, Sacrifice a creature:" would set all
// three.
type AbilityCost struct {
	// Tap requires the source to be untapped, taps it, and — for a
	// creature source — enforces summoning sickness (CR 302.6).
	Tap bool

	// TapOthers taps OTHER untapped permanents the activator
	// controls — Earthcraft's "tap an untapped creature you
	// control", Heritage Druid's three Elves, the station ability's
	// "tap another untapped creature you control" (CR 702.184a).
	// Nil means no such component. See TapOthersCost in
	// tap_others_cost.go, whose field name ADR 0071 §2 fixed.
	//
	// The activator names the permanents in
	// ActivateAbilityParams.TapIDs at announce, beside the crew and
	// sacrifice picks. Three rules are enforced in
	// ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 302.6 / 602.5a  this is NOT the {T} symbol, so summoning
	//	                   sickness does not apply and a creature
	//	                   that arrived this turn may pay.
	//	CR 118.3           a permanent already tapped cannot pay,
	//	                   and the source cannot pay twice: when Tap
	//	                   is also set, naming the source here is
	//	                   refused rather than silently half-paid.
	//	CR 601.2h          the named permanents are kept out of the
	//	                   auto-tapper's reach, so one creature
	//	                   cannot both pay this cost and tap for the
	//	                   mana component of the same cost.
	//
	// Paid with the ability already on the stack, like the convoke
	// taps on the cast path, so a "becomes tapped" payoff
	// (Opposition into a tap-watcher) resolves above it.
	TapOthers *TapOthersCost

	// SacrificeSelf sacrifices the source as part of the cost.
	SacrificeSelf bool

	// SacrificeOther sacrifices other permanents the activator
	// controls, chosen at announce and matched against this spec
	// ("Sacrifice a creature"). The activator names them in
	// ActivateAbilityParams.SacrificeIDs. The source itself is a
	// legal choice when the spec admits it — Carrion Feeder can eat
	// itself, as in paper.
	//
	// The spec's Min == Max is how many (#747): "Sacrifice two
	// artifacts" is a clause with a count of 2, built by
	// effects.SacrificeN. Every one-permanent constructor stamps 1.
	SacrificeOther *TargetSpec

	// Mana is a printed cost string ("{1}{B}") paid from the pool
	// under the same strict / permissive rules casting uses. Empty
	// means no mana component.
	Mana string

	// Life is a life payment (CR 119.4). Paying life is legal at any
	// total above the payment; the SBA loop handles the rest.
	Life int

	// Loyalty is the loyalty-counter component of a planeswalker's
	// loyalty ability (CR 606.4): +N adds N counters to the source,
	// −N removes N, and [0] neither. Nil means "this is not a
	// loyalty ability", which is every other ability in the catalog.
	//
	// A POINTER, not a plain int, because [0] is a real printed cost
	// (Jace's [0]: Brainstorm) and has to be distinguishable from
	// "no loyalty component" — the difference decides whether the
	// activation burns the turn's once-per-turn window.
	//
	// ADR 0020 deliberately kept loyalty OUT of AbilityCost and let
	// the S13.1 ActivateLoyalty sandbox action carry it. ADR 0032 §7
	// re-audited that call: the sandbox action moves a counter and
	// runs no effect, so no planeswalker in the game could do
	// anything, which is what #329 and #334 report. Reversing the
	// exclusion is what makes a loyalty ability an ordinary CR 602
	// activation — same stack item, same Effect closure, same
	// targeting.
	//
	// Three rules ride along with a non-nil Loyalty and are enforced
	// in ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 606.2  the source must be a planeswalker you control
	//	CR 606.6  a −N cost needs at least N loyalty counters
	//	CR 606.3  sorcery speed, once per turn per planeswalker
	//
	// SorcerySpeed therefore does not need to be set alongside it.
	Loyalty *int

	// Crew is the crew number of a Vehicle's crew ability (CR
	// 702.122a): "Tap any number of untapped creatures you control
	// with total power N or more". Zero means "not a crew cost",
	// which is every ability that is not printed on a Vehicle.
	//
	// A plain int rather than a pointer, unlike Loyalty: crew 0 is
	// not a printed cost, so zero and absent are the same thing.
	//
	// Three things make crew different from every other cost here
	// and are enforced in ActivateCatalogAbility rather than asked of
	// each card:
	//
	//	CR 702.122b  the creatures tapped are NOT the source, so
	//	             Tap must stay false — a crewed Vehicle is
	//	             untapped and can attack.
	//	CR 702.122b  summoning sickness does not apply. Tapping a
	//	             creature to crew is not paying a {T} cost, so a
	//	             creature that arrived this turn may crew.
	//	CR 702.122a  power is read at the moment the cost is paid,
	//	             from the post-layer effective value, so an
	//	             anthem and +1/+1 counters both count.
	//
	// The activator names the creatures in
	// ActivateAbilityParams.CrewIDs. Any number is legal as long as
	// the total clears the bar, and a single 5-power creature crews a
	// Vehicle that says crew 3 — the printed text is a floor, not an
	// exact amount.
	Crew int

	// RemoveCounters is a "remove N counters" component (#625): from
	// the source ("Remove a gold counter from this artifact"), from
	// another permanent the activator controls ("remove a loyalty
	// counter from a planeswalker you control"), or of any kind
	// ("Remove a counter from a creature you control"). Nil means no
	// counter component. See CounterRemovalCost in counter_cost.go.
	//
	// The activator names the permanent in
	// ActivateAbilityParams.CounterSourceIDs and, for the any-kind
	// form, the kind in CounterKind — both at announce, beside the
	// sacrifice and crew picks. Two rules are enforced in
	// ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 118.3 / 602.2b  paid at activation, after everything
	//	                   else is validated; the removal is not a
	//	                   replaceable event, so nothing doubles or
	//	                   halves it.
	//	CR 606             removing a planeswalker's loyalty counter
	//	                   this way is not activating a loyalty
	//	                   ability — no sorcery-speed window, and the
	//	                   once-per-turn flag is left alone.
	//
	// "Rather than pay" on an activated ability (Heart of Kiran) is
	// modelled as a SECOND ability entry whose cost is this component,
	// not as an alternatives slot on AbilityCost: the client already
	// lists abilities separately, and the entry is only payable while
	// a real planeswalker holds a real counter, which is the whole of
	// what #259 asks.
	//
	// #789 added two more answers to the same component's question:
	// a count the activator announces (Variable — "Remove X storage
	// counters") and a removal split across several permanents
	// (Among — "Remove two +1/+1 counters from among artifacts you
	// control"). Both ride ActivateAbilityParams.CounterCounts.
	//
	// #943 added the last one: an any-kind removal split across
	// permanents (Counter == "" with Among — Tekuthal, Inquiry
	// Dominus), which asks the kind question once per permanent
	// instead of once per payment and rides
	// ActivateAbilityParams.CounterKinds.
	RemoveCounters *CounterRemovalCost

	// AddCounter is a cost that PUTS a counter on the source —
	// Devoted Druid's "Put a -1/-1 counter on this creature: Untap
	// this creature" (#789). Nil means no such component, which is
	// every ability but a handful.
	//
	// There is nothing to choose, so it rides no field on
	// ActivateAbilityParams. Two rules come with it and are enforced
	// in ActivateCatalogAbility rather than asked of each card:
	//
	//	CR 118.3   a cost that cannot be paid stops the activation.
	//	           canPlaceCounterLocked is the one predicate that
	//	           answers "can this permanent have that counter",
	//	           and the place a Solemnity-style prohibition
	//	           plugs in.
	//	CR 121.1   paying a cost is not an effect, so the placement
	//	           is not replaceable: Doubling Season does NOT
	//	           double the Druid's -1/-1. Same rule the loyalty
	//	           cost and the removal half already follow.
	//
	// Paid AFTER the removal half and before the sacrifices, so a
	// cost that both removes and adds (none printed yet, but the
	// order has to be decided somewhere) reads left to right.
	AddCounter *CounterAddCost

	// MinX is the floor the printed text puts on the announced X —
	// Helm of Obedience's "X can't be 0" is MinX: 1. Zero means the
	// ordinary floor of zero, which is what every other {X} cost
	// charges.
	//
	// There is deliberately no MaxX here. A cost has no printed
	// ceiling on X; what the activator can actually pay is the
	// ceiling, and that is the mana check's job, not the card's.
	// The legal-move enumerator's Options.MaxX is a search bound on
	// the BOT, not a rule, and lives there for that reason.
	//
	// MinX is meaningless without an {X} in Mana, and Register
	// panics at boot on a spec that sets one without the other: a
	// floor on a variable that cannot vary is a card-file mistake,
	// and it would silently make the ability unactivatable.
	MinX int

	// DiscardSelf discards the SOURCE card as part of the cost —
	// cycling's "Discard this card" (CR 702.29a, #660). The source
	// leaves whatever zone the ability was activated from, which for
	// every card that prints it is the hand.
	//
	// Separate from DiscardCards below, and deliberately so
	// (ADR 0062 Decision 2): this component names no card, needs no
	// chooser and puts no options on the wire. Folding it into the
	// general one as "N = 1 matching only this card" would ask the
	// activator a question with one answer on every cycling in the
	// game, and would admit a payload naming some OTHER card.
	DiscardSelf bool

	// DiscardCards is the general discard component: "Discard a
	// card" (Cryptbreaker), "Discard a creature card" (Fauna Shaman,
	// Survival of the Fittest, Tortured Existence). Nil means no
	// discard-another component. See DiscardCost in discard_cost.go.
	//
	// The activator names the cards in
	// ActivateAbilityParams.DiscardIDs at announce, beside the
	// sacrifice and crew picks — the SAME announce-time cost pause
	// the client already opens for those, not a new one. CR 602.2b
	// activates an ability in one indivisible step, so a cost may
	// not stop to ask the server a question mid-announce.
	DiscardCards *DiscardCost

	// ReturnToHand returns permanents the activator controls to their
	// OWNERS' hands as part of the cost (#1213) — Quirion Ranger's
	// "Return a Forest you control to its owner's hand", Master
	// Transmuter's artifact, Meloku the Clouded Mirror's land. Nil
	// means no such component. See ReturnToHandCost in
	// return_cost.go.
	//
	// The activator names the permanents in
	// ActivateAbilityParams.ReturnIDs at announce, beside the
	// sacrifice, tap and crew picks. Two rules ride along and are
	// enforced in ActivateCatalogAbility rather than asked of each
	// card:
	//
	//	CR 118.3   a board that cannot produce Count matches cannot
	//	           activate the ability at all.
	//	CR 601.2h  the move settles without a prompt, so a commander
	//	           returned this way is not offered the command zone
	//	           mid-announce — payReturnToHandCostLocked says why.
	//
	// Paid with the sacrifices, BEFORE the ability is on the stack,
	// so the leaves-the-battlefield triggers it queues are drained
	// above the ability (CR 603.3b). The source may be a legal pick
	// when the filter admits it, which is why the activation path
	// drops its `source` pointer afterwards exactly as it does after
	// a sacrifice.
	ReturnToHand *ReturnToHandCost
	// ExileSelf exiles the SOURCE CARD from the zone the ability was
	// activated from, as part of the cost — scavenge's "Exile this
	// card from your graveyard" (CR 702.96a), embalm's and
	// eternalize's (CR 702.128a / CR 702.129a). #1221.
	//
	// DiscardSelf's sibling, one zone over, and written as a second
	// bit rather than as a zone on one "the source pays" component
	// because the two are different rules: discarding is a CR 701.8
	// keyword action that puts the card in a graveyard and fires
	// EventDiscardCard (and, for cycling, EventCycle); exiling as a
	// cost is a plain CR 406 move that fires neither. A card file
	// that wanted "discard this from your graveyard" would be
	// writing a card that does not exist.
	//
	// The source has to be IN A GRAVEYARD — every card that prints
	// the component says "from your graveyard" — and Register
	// refuses the component on an ability that does not declare
	// ZoneGraveyard, exactly as it refuses DiscardSelf off the hand.
	// Paid last, with the discards, because it moves the source and
	// invalidates it; the effect that follows reads the card back
	// out of EXILE by instance ID (LookupCardForEffect), which is
	// where the cost has just put it.
	ExileSelf bool
}

// DemandsX reports whether the ability's mana component contains
// {X}, and is therefore an ability the activator announces a value
// for at CR 602.2b.
//
// X used to live in the MANA component and nowhere else. #1213 added
// the second place, and it is the same second place the cast path has
// always had: a spell grows an X outside its printed cost with Toxic
// Deluge's "pay X life" (AdditionalCost.PayLifeX) and with waterbend
// (TapPermanentsCost), and Grim Hireling's "{B}, Sacrifice X
// Treasures" is that on an ability. So the question is asked of the
// COST rather than of the cost string, and every consumer — the view,
// the enumerator, the client — still derives "does this prompt for X"
// from the declared components rather than from a flag a card file
// could forget to set.
//
// XSlots is deliberately NOT widened with it: a sacrifice clause's X
// buys permanents, not generic mana, so Grim Hireling's {B} stays {B}
// at every announced X.
func (c AbilityCost) DemandsX() bool {
	return c.XSlots() > 0 || SacrificeCountFromX(c.SacrificeOther)
}

// XSlots is how many {X} tokens the mana component carries. Usually
// 1; Treasure Vault's "{X}{X}" is 2, and the total generic demand is
// XSlots * the announced X. An unparseable cost reports 0 — the
// activation path rejects it separately rather than guessing.
func (c AbilityCost) XSlots() int {
	if c.Mana == "" {
		return 0
	}
	cost, err := ParseCost(c.Mana)
	if err != nil {
		return 0
	}
	return cost.XSlots
}

// FloorX is the smallest legal announcement for this cost: MinX when
// the card prints a floor, zero otherwise. Zero for a cost with no
// {X} at all, which is the only value the engine accepts there.
func (c AbilityCost) FloorX() int {
	if !c.DemandsX() || c.MinX < 0 {
		return 0
	}
	return c.MinX
}

// ActivatedAbilityShape is one activated ability on a permanent, as
// the game package consumes it. Mirrors effects.ActivatedAbility,
// which lives in the catalog package (the same import-cycle dodge
// ManaAbilityShape uses).
type ActivatedAbilityShape struct {
	// Label is the oracle text of the ability, shown in the
	// activation menu: "Sacrifice a creature: Goblin Bombardment
	// deals 1 damage to any target."
	Label string

	Cost AbilityCost

	// Targets is the ability's target clause, validated at announce
	// and re-checked at resolution exactly as a spell's is. Nil for
	// untargeted abilities.
	Targets *TargetSpec

	// Modes is the CR 700.2 mode clause of a modal activated ability
	// ("{4}, {T}: Choose one — double the counters on target
	// permanent; double the counters you have"). The same
	// game.ModeSpec a modal spell and a modal triggered ability
	// declare — one struct, three owners (#764, ADR 0065 §3).
	//
	// Chosen at ACTIVATION with the targets (CR 602.2b), in the same
	// message, because activating an ability is one indivisible step
	// and there is nobody to prompt: the player who activates is the
	// player who chooses. Validated exactly as a cast's modes are;
	// each chosen occurrence contributes its option's target clauses
	// to the announcement, in mode order.
	//
	// Nil for every ability that is not modal, which is nearly all of
	// them. Added by #764.
	Modes *ModeSpec

	// SorcerySpeed marks "activate only as a sorcery" (CR 602.5d).
	SorcerySpeed bool

	// Zones is the set of zones this ability functions from
	// (CR 113.6 — an ability works only where it says it does).
	// Nil means the BATTLEFIELD, which is every ability written
	// before this field existed and nearly every ability there will
	// ever be.
	//
	// A slice rather than a FromHand bool, and per-ABILITY rather
	// than per-card, because the cards immediately behind cycling
	// want other zones and want them one entry at a time:
	// Reassembling Skeleton and Drownyard Temple activate from the
	// graveyard, and Eternal Dragon prints a hand ability
	// (plainscycling) and a graveyard ability on one card.
	// ADR 0062 Decision 1.
	//
	// ActivateCatalogAbility checks it AFTER the index lookup and
	// before anything is validated or paid (ErrActivationZoneNotAllowed),
	// so there is one activation path with a zone dimension rather
	// than a second entry point for hand abilities. A cost component
	// that names the source as a permanent — Tap, SacrificeSelf,
	// Crew, Loyalty — is refused at effects.Register on an ability
	// that declares a non-battlefield zone, not silently at runtime.
	Zones []ZoneKind

	// Cycling marks this ability as the card's CYCLING ability
	// (CR 702.29a). Activating it is "cycling a card" (CR 702.29b),
	// which emits EventCycle beside the cost's EventDiscardCard so
	// Astral Slide / Drake Haven-style watchers can see it.
	//
	// A declarative bit rather than a name match on the label: the
	// event has to be a fact about what the card printed, and a
	// typecycling ability ("Basic landcycling {2}") is a cycling
	// ability too even though its effect searches rather than draws.
	Cycling bool

	// Condition is the ability's other activation instructions
	// (CR 602.1b): "Activate only if an opponent controls four or
	// more lands" (Tectonic Edge), "Activate only during your turn"
	// (Sanctum of Eternity). Nil means none, which is nearly every
	// ability. ADR 0020's #743 addendum.
	//
	// The contract is ManaAbilityShape.Condition's, word for word:
	// READ-ONLY, and it runs under g.mu — held for write by
	// ActivateCatalogAbility, for read by the view and the legal
	// enumerator — so it reads *ForEffect accessors and g.Turn /
	// g.Seats directly, never a public locking accessor
	// (g.ActivePlayer deadlocks the activation). `controller` is the
	// activating player, the "you" of the printed text; `source` is
	// the permanent's instance ID, enough to read a per-source count
	// later without changing the signature.
	//
	// Checked once, in ActivateCatalogAbility, right after the timing
	// check and before X, the loyalty checks, the costs and the
	// targets, so a false return is ErrConditionNotMet with nothing
	// paid (CR 602.5). Never re-checked at resolution: an activation
	// instruction is not part of the effect (CR 602.1b).
	//
	// Every viewer receives the view's condition_unmet flag, so a
	// condition must read only public information (board counts,
	// graveyard and hand sizes, life, the turn and the step).
	//
	// SorcerySpeed stays a separate flag: some cards print both
	// (Speaker of the Heavens), and the two fail with different
	// errors.
	Condition func(g *Game, controller, source uuid.UUID) bool

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this ability exists only while the permanent has the
	// designation named — a Case's "Solved — {1}{B}, Sacrifice this:
	// …" is CaseSolved(). The zero value is "no gate".
	//
	// It is NOT Condition. Condition is CR 602.1b, an activation
	// instruction on an ability the permanent HAS ("activate only if
	// an opponent controls four or more lands"), and the view renders
	// it greyed with condition_unmet. A designation gate says the
	// ability is not there at all, so it is absent from
	// ActivatedAbilitiesForCard — and therefore from the activation
	// path, the legal-move enumerator and the wire — rather than
	// offered and refused.
	//
	// The level-up ability itself takes a Condition, not this: "{3}{U}:
	// Level 2" is printed on the Class from the moment it enters, and
	// CR 716.2e's "activate only if this Class is level 1" is an
	// activation instruction word for word.
	//
	// See designations.go and ADR 0071.
	ActiveWhen Designation

	// Exhaust marks an EXHAUST ability (#1181):
	//
	//	Exhaust — {4}: Earthbend 4.
	//	(Activate each exhaust ability only once.)
	//
	// One declarative bit and no per-card logic, exactly as Cycling
	// above is: the keyword IS the rule, and a card that wrote its own
	// "have I done this yet" check would be writing a rule the engine
	// has to enforce in three places anyway.
	//
	// What the bit buys, all of it in game.AbilityExhausted:
	//
	//   - per ABILITY, not per source. A card with three exhaust
	//     abilities (Loot, the Pathfinder) may activate each of them
	//     once, and activating the first must not lock the other two —
	//     which is why the record is keyed by the ability's label and
	//     not by the permanent.
	//   - for the whole GAME, not the turn. Game.Activations.Ever is
	//     never reset (activation_tally.go).
	//   - spent at the ANNOUNCE. An exhaust ability countered on the
	//     stack, or fizzled for want of a target, has been activated.
	//   - CR 400.7 all the way down: a flicker refreshes it, a phase-
	//     out would not, and a copy of the permanent has its own
	//     (CR 707.2).
	//
	// It is NOT Condition and NOT ActiveWhen. An exhausted ability is
	// still printed on the permanent and still enumerated by the card
	// — the view ships `exhausted` and greys the row, and
	// ActivateCatalogAbility refuses with ErrAbilityExhausted — where
	// a designation gate would make it absent entirely. A card that
	// prints BOTH (Bitter Work's "Exhaust — {4}: Earthbend 4. Activate
	// only during your turn.") sets this and Condition, and the two
	// are checked in that order.
	Exhaust bool

	// Effect runs at resolution against the live game. Same contract
	// as TriggeredAbility's stack items: never capture a *Card,
	// read what you need off the item and the game.
	Effect func(g *Game, item *StackItem) error
}

// CatalogActivatedAbilities returns the registered activated
// abilities for an oracle ID, or nil. Populated by the effects
// package at init, alongside the other catalog hooks.
var CatalogActivatedAbilities func(oracleID string) []ActivatedAbilityShape

// ActivatedAbilitiesForCard returns the activated abilities a card
// offers right now. Catalog-only: unlike mana abilities there's no
// synthetic fallback, because there's no ability every card of some
// type implicitly has.
func ActivatedAbilitiesForCard(c Card) []ActivatedAbilityShape {
	// S24 layer 6: a permanent an ability-removing effect applies to
	// offers nothing, and that has to be checked BEFORE the
	// instance-carried list as well as before the catalog. A Clue
	// token's "{2}, Sacrifice this: Draw a card" is an activated
	// ability like any other, and Darksteel Mutation takes it away
	// exactly as it takes away Sol Ring's.
	//
	// This one accessor is why internal/legal needs no change:
	// legal.Move enumeration, the client's context menu
	// (protocol.CardView), the lobby's ability lookup and
	// ActivateCatalogAbility itself all read through here, so the
	// enumerator can never offer a bot a move the engine will refuse
	// — the #544 hung-table failure mode.
	if c.HasLostAllAbilities() {
		return nil
	}
	// S21 sub-PR 4: intrinsic abilities win — a token has no oracle
	// ID for the catalog to key on, and Food / Clue / Blood ARE
	// their activated ability.
	if len(c.ActivatedAbilities) > 0 {
		return c.ActivatedAbilities
	}
	if CatalogActivatedAbilities == nil {
		return nil
	}
	// #521: the guard used to be `c.OracleID == ""` as well, which
	// made a token unreachable here by construction. A token now has
	// a catalog key of its own, so the question is the one the key
	// already answers — an object with no entry has the empty key,
	// whether because it is uncatalogued or because CR 708.2a has
	// silenced it.
	key := CatalogAbilityKey(c)
	if key == "" {
		return nil
	}
	// ADR 0071: an ability gated on a designation the permanent does
	// not have is not on the permanent. Because every consumer reads
	// through this one accessor, a Case's "Solved — {1}{B}, Sacrifice
	// this Case: …" is absent from the activation path, the legal-move
	// enumerator, the lobby lookup and the wire together — greying it
	// in one of them and offering it in another is not representable.
	//
	// The intrinsic list above is deliberately NOT gated: a token
	// carries its own abilities and has no catalog entry to print a
	// designation on.
	return activeOnly(c, CatalogActivatedAbilities(key), func(a ActivatedAbilityShape) Designation {
		return a.ActiveWhen
	})
}

// ActivateAbilityParams carries the announce-time choices for a
// catalog activated ability.
type ActivateAbilityParams struct {
	// SacrificeIDs names the permanents paid to a SacrificeOther
	// cost: exactly the clause's count (SacrificeCostCount), each
	// once. Their order does not matter — the permanents leave as
	// one simultaneous exit (#747).
	SacrificeIDs []uuid.UUID

	// CrewIDs names the creatures tapped to pay a Crew cost, in the
	// order the activator picked them. Any number is legal; what
	// matters is that their total effective power clears the crew
	// number (CR 702.122a).
	CrewIDs []uuid.UUID

	// TapIDs names the permanents tapped to pay a TapOthers cost
	// (#758): exactly the clause's Count, each once, each untapped
	// and controlled by the activator, and never the source when
	// the clause prints "another". Their order does not matter.
	//
	// On the wire as `tap_ids`, the same name the cast path's
	// convoke taps ride under (CastSpellParams.TapIDs) — the same
	// question, asked of an activation instead of a cast.
	TapIDs []uuid.UUID

	// CounterSourceIDs names the permanents a RemoveCounters cost
	// removes from (#625). Exactly one for the "from a planeswalker
	// you control" form; empty (or the source's own ID) for the
	// self form; any number for the "from among artifacts you
	// control" form (#789), which is why it was a slice from the
	// start.
	CounterSourceIDs []uuid.UUID

	// CounterCounts is the per-permanent split, parallel to
	// CounterSourceIDs (#789): how many counters come off each. It
	// is what makes ONE payment shape cover all five printed forms
	// of the component.
	//
	// Empty means "the printed count, off the one permanent named"
	// — the shape every #625 client sends, and still the whole
	// answer for a fixed single-permanent cost. An among cost sends
	// counts summing to exactly N; a variable cost sends the count
	// the activator announced, which is the payment itself (CR
	// 601.2b's "announce the value of X", one component over).
	CounterCounts []int

	// DiscardIDs names the cards paid to a DiscardCards cost
	// (#660): exactly the clause's count, each once, each in the
	// activator's hand and each matching the clause. Never the
	// source of an ability activated FROM hand — a card cannot pay
	// for its own activation twice, and cycling's "Discard this
	// card" is the separate DiscardSelf component.
	//
	// Chosen at announce with the sacrifice and crew picks
	// (CR 602.2b), on the wire as `discard_ids`, exactly as a cast's
	// additional discard cost rides CastSpellParams.DiscardIDs.
	DiscardIDs []uuid.UUID

	// ReturnIDs names the permanents paid to a ReturnToHand cost
	// (#1213): exactly the clause's Count, each once, each on the
	// battlefield under the activator's control and each matched by
	// the clause. Their order does not matter — each takes its own
	// route to its OWNER's hand.
	//
	// On the wire as `return_ids`, beside `sacrifice_ids` and
	// `tap_ids`: the same announce-time cost pause the client
	// already opens for those, not a new one.
	ReturnIDs []uuid.UUID

	// CounterKind is the kind a "remove a counter" cost of ANY kind
	// removes (Fain, the Broker), chosen at announce with the
	// permanent. Optional for a cost that prints its kind — if sent,
	// it must repeat that kind. One kind for the whole payment.
	CounterKind string

	// CounterKinds is the per-permanent kind, parallel to
	// CounterSourceIDs (#943): what an any-kind AMONG cost removes
	// from each permanent, when the parts do not share a kind.
	// Tekuthal, Inquiry Dominus' "Remove three counters from among
	// other artifacts, creatures, and planeswalkers you control" is
	// the only printed cost that needs it — a loyalty counter off a
	// planeswalker and two +1/+1 counters off a creature is one legal
	// payment, and no single CounterKind can say so.
	//
	// Empty is the ordinary case, including an any-kind payment whose
	// parts happen to share a kind: CounterKind says it, and every
	// client that predates this field keeps working. When both are
	// sent they must agree.
	CounterKinds []string

	// Targets are the ability's targets, validated against the
	// ability's clause list — or, for a modal ability, against the
	// clauses of the chosen modes (#764).
	Targets []TargetRef

	// Modes are the mode indexes announced for a modal activated
	// ability (CR 602.2b, CR 700.2), in the order chosen; a repeated
	// index is legal only when the ability's ModeSpec is Repeatable
	// (CR 700.2d). Empty for every non-modal ability, and non-empty
	// for one is rejected rather than ignored. Added by #764.
	Modes []int

	// XValue is the value announced for an {X} in the ability's mana
	// component (CR 602.2b). Chosen as part of ACTIVATING the
	// ability — after the source and the mode, before any cost is
	// paid — and locked from that moment: it rides onto the stack
	// item and nothing later can change it, which is what makes an
	// X ability's effect (Helm of Obedience's mill, Treasure Vault's
	// token count) a fact about the announcement rather than about
	// the board at resolution.
	//
	// The same slot on the cast path is CastSpellParams.XValue, and
	// the two are validated the same way: non-negative, zero unless
	// the cost actually has an {X}, and at least the cost's printed
	// floor.
	XValue int

	// PhyrexianLife is how many of the mana component's Phyrexian
	// symbols the activator is paying with 2 life each instead of
	// mana (CR 107.4c, and CR 107.4f for the ten hybrid Phyrexian
	// symbols) — Birthing Pod's "{1}{G/P}", Solphim's
	// "{1}{R/P}{R/P}".
	//
	// A COUNT, not a list: the engine strikes the symbols a life
	// payment can actually save first (PhyrexianLifePlan), and
	// choosing between two symbols the pool can both pay changes
	// nothing but which colour is left floating.
	//
	// CR 602.2b makes "how do you intend to pay each hybrid and
	// Phyrexian symbol" part of activating the ability, in the same
	// indivisible announcement as the modes, the targets and X — so
	// it is a parameter here rather than a PendingChoice, exactly as
	// CastSpellParams.PhyrexianLife is on the cast path (#917). It is
	// the same field name, the same wire name (`phyrexian_life`) and
	// the same strike-and-pay helper; only the announcement differs.
	// More than the cost prints, or more life than CR 119.4 allows,
	// is refused before anything is paid.
	PhyrexianLife int

	// Strict / AutoTap mirror CastSpellParams: they gate the mana
	// component of the cost the same way a cast is gated.
	Strict  bool
	AutoTap bool
}

// ActivateCatalogAbility activates ability `index` on a permanent
// the player controls: validates timing and cost, pays the cost,
// and pushes a stack item carrying the ability's effect (CR 602.2).
// The activator retains priority, as with any announce.
//
// Costs are validated in full before ANY of them is paid, so a
// half-paid activation can't strand the board — the same discipline
// the sacrifice-cost mana ability path uses.
//
// Caller must NOT hold g.mu.
func (g *Game) ActivateCatalogAbility(playerID, cardID uuid.UUID, index int, params ActivateAbilityParams) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	if g.SplitSecondActive {
		return ErrSplitSecondActive
	}
	p := g.playerByIDLocked(playerID)
	if p == nil {
		return ErrPlayerNotFound
	}
	// S24: the CR 602.5 restriction rides on the effective
	// characteristic, so the layers have to be fresh before the
	// source is judged. Hoisted above the controller check rather
	// than tucked in beside the restriction test, because
	// Card.Controller is itself layer 2's materialised output — a
	// Mind Control that resolved a moment ago has to be visible to
	// "do you control this permanent" too. Fast-path no-op when
	// nothing has changed.
	g.RecomputeLayersIfStaleLocked()
	// #660 / ADR 0062 Decision 1: the source is found in WHATEVER
	// zone holds it, not on the battlefield alone. One activation
	// path with a zone dimension — cycling is an ordinary CR 602
	// activation whose ability happens to function from hand
	// (CR 702.29a), not a fork.
	source, srcZone := g.findCardAndZoneLocked(cardID)
	if source == nil {
		return ErrCardNotFound
	}
	if srcZone == ZoneBattlefield {
		if source.Controller != playerID {
			return ErrCardCallerMismatch
		}
		// CR 602.5: "its activated abilities can't be activated"
		// (Arrest, Faith's Fetters). Checked before the index lookup
		// so the answer does not depend on which ability was named,
		// and before any cost validation so nothing is paid. Loyalty
		// abilities are activated abilities (CR 606.1) and are
		// covered here too. Mana abilities take the other entry
		// point and the other bit — see restrictions.go on why that
		// split exists.
		if !CanActivateAbilities(source) {
			return ErrCantActivate
		}
	} else if source.Owner != playerID {
		// CR 108.4: a card outside the battlefield and the stack has
		// no controller, so its owner is the "you" of the printed
		// text. Every hand in this engine holds only its owner's
		// cards, so this is the same check the battlefield arm makes,
		// asked of the field that means something here.
		//
		// CanActivateAbilities is deliberately NOT asked: Arrest and
		// Faith's Fetters apply to a permanent, and layer 6 has
		// nothing to say about a card in a hand.
		return ErrCardCallerMismatch
	}
	abilities := ActivatedAbilitiesForCard(*source)
	if index < 0 || index >= len(abilities) {
		return ErrInvalidParam
	}
	ab := abilities[index]
	// CR 113.6: an ability functions only from the zone it says it
	// functions from. Checked after the index lookup (so the index
	// is the one the view and the enumerator published) and before
	// everything else, so a cycling ability fired from the
	// battlefield — or a sacrifice outlet fired from hand — costs
	// nothing and pays nothing.
	if !AbilityFunctionsFromZone(ab, srcZone) {
		return ErrActivationZoneNotAllowed
	}
	// #1210, CR 602.5a / CR 101.2: the board-wide "can't be
	// activated" gate — Cursed Totem, Linvala, Collector Ouphe,
	// Pithing Needle. ONE function, four callers; see
	// activation_gate.go.
	//
	// After the zone check, so the ability judged is the one the view
	// and the enumerator published, and BEFORE the timing check, X,
	// the targets and every cost — a refused activation costs
	// nothing. Before timing because "can't be activated" is the
	// answer that will still be true next turn where
	// ErrSorcerySpeedRequired will not; that order is ours, not the
	// rules'.
	//
	// Separate from CanActivateAbilities above, which is the
	// per-PERMANENT Arrest bit layer 6 already answered.
	if err := g.ActivationGateLocked(playerID, *source, srcZone, ActivationAbility{Label: ab.Label}); err != nil {
		return err
	}

	// --- timing -------------------------------------------------
	// CR 606.3: a loyalty ability is sorcery-speed whether or not
	// the catalog entry bothered to say so — the loyalty component
	// carries the restriction, so a card can't forget it.
	if (ab.SorcerySpeed || ab.Cost.Loyalty != nil) && !g.sorcerySpeedOpenLocked(playerID) {
		return ErrSorcerySpeedRequired
	}
	// #1181: "Activate each exhaust ability only once". The key is
	// taken HERE, before anything is paid, because a cost that moves
	// the source (SacrificeSelf, DiscardSelf) ends the object and
	// carries Card.ObjectEpoch with it — a key built at the announce
	// would be written against an object that never had the ability.
	// The same key is read now and written below, so the gate and the
	// record cannot address different things.
	activationKey := g.activationTallyKeyLocked(cardID, ab.Label)
	if g.AbilityExhausted(playerID, cardID, ab) {
		return ErrAbilityExhausted
	}
	// CR 602.1b / 602.5: the ability's "Activate only if …" and
	// "Activate only during …" instructions (#743). A player can't
	// begin to activate an ability whose instructions forbid it, so
	// this stops the activation before X, targets or any cost — and
	// it is never checked again at resolution. When both this and the
	// timing check fail, the timing error wins; that order is ours,
	// not the rules'.
	if ab.Condition != nil && !ab.Condition(g, playerID, cardID) {
		return ErrConditionNotMet
	}

	// --- validate every cost before paying any ------------------
	// CR 602.2b: X is announced as part of activating, with the
	// targets and before anything is paid. Same three checks the
	// cast path makes at CR 601.2b, in the same order.
	//
	// The "no {X}, so X must be zero" branch is rejected rather than
	// ignored on purpose, exactly as an unexpected sacrifice_ids is:
	// a client that sends an X for Krenko is confused about which
	// ability it is firing, and silently dropping the field would
	// hide that from whoever has to debug it.
	if params.XValue < 0 {
		return ErrInvalidParam
	}
	if !ab.Cost.DemandsX() {
		if params.XValue != 0 {
			return ErrInvalidParam
		}
	} else if params.XValue < ab.Cost.FloorX() {
		// "X can't be 0" (Helm of Obedience). A floor is part of the
		// cost, so announcing under it is an illegal announcement,
		// not a cheap one.
		return ErrInvalidParam
	}
	// CR 107.4f / CR 602.2b (#917): the Phyrexian half of the
	// announcement. How many symbols the cost prints, and whether
	// CR 119.4 allows the life, are checked against the PARSED cost
	// in the payment step below — by the one strike-and-pay helper
	// the cast path also runs, before a point or a token is spent.
	// What belongs here is the case that step never reaches: an
	// ability with no mana component at all. A claim against it is a
	// client firing the wrong ability rather than a cheap
	// activation, so it is refused rather than dropped, exactly as
	// an X on a costless ability is.
	if params.PhyrexianLife != 0 && ab.Cost.Mana == "" {
		return ErrInvalidParam
	}
	if ab.Cost.Loyalty != nil {
		// CR 606.3 / 606.5 say PERMANENT, not planeswalker, and that
		// wording is load-bearing rather than loose: a loyalty ability
		// is defined by its cost symbol (CR 606.1), and a permanent can
		// carry one without being a planeswalker — the printed case is
		// a planeswalker that a type-setting effect has stopped being
		// one while keeping its abilities (Song of the Dryads on a
		// Teferi), and the catalog can print the same shape on anything.
		// This used to refuse with ErrNotAPlaneswalker on top of the
		// controller check above, which asked a question CR 606 does not.
		// #1157 is what sent us looking: The Aetherspark is a
		// Legendary Artifact Planeswalker — Equipment, so it passed the
		// gate and the gate was never the report's cause, but a gate
		// that only happens not to fire is still a rule the engine has
		// wrong. The SANDBOX ActivateLoyalty verb keeps its own
		// IsPlaneswalker check (mutations.go): that one INVENTS a
		// loyalty ability for a card the catalog cannot read, and its
		// whole premise is the planeswalker row it is offered from.
		//
		// CR 606.3: once per turn per permanent. The flag was S13.1's
		// and only the sandbox action consulted it; this is the path
		// that matters now.
		if g.LoyaltyActivatedThisTurn[cardID] {
			return ErrLoyaltyAlreadyActivated
		}
		// CR 606.5: you can't activate a −N ability with fewer than
		// N loyalty counters. Paying down to exactly 0 is legal and
		// the 704.5i SBA sweeps the permanent afterwards — when it is
		// a planeswalker, which is the one place that rule does ask.
		if n := *ab.Cost.Loyalty; n < 0 && source.Counters[CounterLoyalty] < -n {
			return ErrInsufficientLoyalty
		}
	}
	if ab.Cost.Tap {
		if source.Tapped {
			return ErrAlreadyTapped
		}
		// CR 302.6: a creature's {T} ability needs it to have been
		// under your control since your most recent turn began.
		// Non-creature sources (Krenko is a creature; an artifact
		// with {T} isn't) are never sick.
		if source.IsCreature() && HasSummoningSickness(source) {
			return ErrSummoningSick
		}
	}
	sacrifices, err := g.validateSacrificeCostLocked(playerID, cardID, ab.Cost, params.SacrificeIDs, params.XValue)
	if err != nil {
		return err
	}
	crew, err := g.validateCrewCostLocked(playerID, ab.Cost, params.CrewIDs)
	if err != nil {
		return err
	}
	// #758: the tap-another component. Validated here with every
	// other cost and paid at the foot of the announce, so a refusal
	// leaves the board entirely untapped (ADR 0020 §3).
	if err := g.validateTapOthersCostLocked(playerID, cardID, ab.Cost.TapOthers, params.TapIDs); err != nil {
		return err
	}
	// CR 118.3: the source can only be tapped once. A cost that
	// prints both {T} and "tap N untapped permanents you control"
	// (Jaspera Sentinel) has already spent the source on the {T},
	// so naming it again would pay one of the N with a permanent
	// that is about to be tapped anyway — an underpaid cost. The
	// component itself cannot see this: ExcludeSource is the
	// printed word "another", and this is the interaction between
	// two components of one cost, which only the activation path
	// holds both halves of.
	if ab.Cost.Tap && !ab.Cost.TapOthers.Empty() {
		for _, id := range params.TapIDs {
			if id == cardID {
				return ErrInvalidParam
			}
		}
	}
	// #1213: the return-to-hand component. Validated here with every
	// other cost and paid with the sacrifices below, so a refusal
	// leaves the board entirely untouched (ADR 0020 §3).
	if err := g.validateReturnToHandCostLocked(playerID, cardID, ab.Cost.ReturnToHand, params.ReturnIDs); err != nil {
		return err
	}
	counters, err := g.validateCounterRemovalCostLocked(playerID, cardID, ab.Cost, CounterCostPayment{
		SourceIDs: params.CounterSourceIDs,
		Counts:    params.CounterCounts,
		Kind:      params.CounterKind,
		Kinds:     params.CounterKinds,
	})
	if err != nil {
		return err
	}
	// CR 118.3 (#789): "Put a -1/-1 counter on this creature" is a
	// cost, so an activator who cannot put that counter on cannot
	// activate. Checked here, with everything else, so a refusal
	// leaves the source untapped and the pool untouched.
	if !g.canPlaceCounterLocked(playerID, cardID, ab.Cost.AddCounter) {
		return ErrCantPayCounterCost
	}
	// #660: the discard components. Validated here with everything
	// else and paid last (a discard moves the source out of hand,
	// which invalidates `source` exactly as a sacrifice does).
	discards, err := g.validateDiscardCostLocked(playerID, cardID, srcZone, ab.Cost, params.DiscardIDs)
	if err != nil {
		return err
	}
	// #1221: the discard component's sibling one zone over —
	// scavenge's and embalm's "Exile this card from your graveyard".
	// Nothing to resolve (the source IS the payment), so this is the
	// zone check alone, made here with the rest so a refusal costs
	// nothing.
	if err := g.validateExileSelfCostLocked(srcZone, ab.Cost); err != nil {
		return err
	}
	if ab.Cost.Life > 0 && p.Life < ab.Cost.Life {
		// CR 119.4 forbids paying more life than you have. Paying
		// down to exactly 0 is legal; the SBA loop ends the game
		// after.
		return ErrInvalidParam
	}
	// CR 602.2b / 700.2: the modes are announced with the targets, in
	// that order — the chosen bullets are what decide which target
	// clauses the activation even has (#764).
	if err := validateModes(ab.Modes, params.Modes); err != nil {
		return err
	}
	if ab.Modes == nil && len(params.Modes) > 0 {
		return ErrInvalidParam
	}
	steps := AnnouncedClauses(ab.Targets, ab.Modes, params.Modes)
	if len(steps) == 0 && len(params.Targets) > 0 {
		return ErrInvalidParam
	}
	xSteps := resolveStepCountsFromX(steps, params.XValue)
	params.Targets = assignAnnouncedSlots(steps, params.Targets)
	for _, i := range xSteps {
		if n := stepTargetCount(steps[i], params.Targets); n != params.XValue {
			return ErrInvalidParam
		}
	}
	// CR 702.16b: the SOURCE of an activated ability is the permanent
	// that has it, not its controller. A red player activating a
	// colourless Equipment's ability may still target a pro-red
	// creature; a red permanent's ability may not (#662).
	if err := g.validateAnnouncedTargetsLocked(SourceObject(playerID, source), steps, params.Targets); err != nil {
		return err
	}

	// --- pay ----------------------------------------------------
	//
	// #789 / #761: one record of what this announcement paid, built
	// as the components are charged and stamped onto the stack item
	// below. The effect reads the paid counter count back out of it
	// (Context.CountersRemoved), the same way it reads X.
	paid := PaidCost{}
	if ab.Cost.Mana != "" {
		// S32 (#352): "activate abilities of colorless Eldrazi" is a
		// restriction on the SOURCE permanent, so the spend context
		// is built from it.
		// An ability whose cost includes {T} cannot tap its own
		// source for mana to help pay itself — the {T} and the mana
		// are components of the same cost, and the source can only
		// be tapped once. Without the exclusion the auto-tapper is
		// free to spend Treasure Vault's own "{T}: Add {C}" on its
		// "{X}{X}, {T}" ability, which is a free mana on every
		// activation of every tap ability with a mana component.
		var excluded map[uuid.UUID]bool
		if ab.Cost.Tap {
			excluded = map[uuid.UUID]bool{cardID: true}
		}
		// #758, the same rule one component over: a permanent named
		// to pay the TapOthers half is already spent, so the
		// auto-tapper must not also tap it for mana. Without this
		// the payment below would find it tapped and skip it, and
		// the cost would be paid with one permanent fewer than the
		// clause prints (CR 118.3).
		for _, id := range params.TapIDs {
			if excluded == nil {
				excluded = make(map[uuid.UUID]bool, len(params.TapIDs))
			}
			excluded[id] = true
		}
		// #1184: the CR 601.2f pass over the ability's mana component
		// — "Exhaust abilities of other permanents you control cost
		// {2} less to activate" (Boom Scholar). The same function the
		// legal-move enumerator prices with, so a bot is never
		// offered an activation the engine then refuses for want of
		// mana. A board with no activation-scoped modifier on it
		// returns the printed cost and walks nothing.
		manaCost, err := g.AbilityManaCostForEffect(playerID, *source, srcZone, ab)
		if err != nil {
			return ErrInvalidParam
		}
		spent, err := g.payAbilityManaCostLocked(p, cardID, source.Name, manaCost, params, ManaSpendForAbility(*source), excluded)
		if err != nil {
			return err
		}
		paid.Mana = spent.Mana
		paid.OnPaper = spent.OnPaper
		// CR 107.4f (#917): the life the activator announced for the
		// cost's Phyrexian symbols. It is already paid — the helper
		// pays it with the mana half, in one indivisible step — and
		// the record carries it beside the printed Life component
		// below, because LifePaid is every point this announcement
		// paid.
		paid.LifePaid = spent.LifePaid
	}
	if ab.Cost.Tap {
		source.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: cardID})
	}
	for _, id := range crew {
		// The crewing creatures tap, the Vehicle does not (CR
		// 702.122b) — which is the whole point, since a Vehicle that
		// tapped to crew itself could never attack.
		if c := findBattlefieldCard(g, id); c != nil {
			c.Tapped = true
			g.EmitEvent(Event{Kind: EventTapCard, Actor: playerID, CardID: id})
		}
	}
	// #793: the cost path — CR 602.2b activates an ability in one
	// indivisible step, so the payment runs the CR 614 window
	// (CR 119.4) but never stops to ask a CR 616 ordering question.
	if ab.Cost.Life > 0 {
		if err := g.PayLifeForEffect(cardID, playerID, ab.Cost.Life); err != nil {
			return err
		}
		paid.LifePaid += ab.Cost.Life
	}
	if ab.Cost.Loyalty != nil {
		// applyCounterLocked, not AddCounterForEffect: paying a cost
		// is not an effect (CR 121.1 / 606.2), so counter-doubling
		// replacements do NOT apply to a loyalty ability's + cost.
		// Doubling Season really does nothing here, and routing
		// through the CR 614 pipeline would silently make it.
		if err := g.applyCounterLocked(cardID, CounterLoyalty, *ab.Cost.Loyalty); err != nil {
			return err
		}
		if g.LoyaltyActivatedThisTurn == nil {
			g.LoyaltyActivatedThisTurn = make(map[uuid.UUID]bool)
		}
		g.LoyaltyActivatedThisTurn[cardID] = true
	}
	// #625: a "remove N counters" component. After life and loyalty,
	// before sacrifices — a self-form removal on a source that is
	// also sacrificed has to find the source still on the
	// battlefield. Never replaceable and never a loyalty activation;
	// payCounterRemovalLocked says why.
	if err := g.payCounterRemovalLocked(counters); err != nil {
		return err
	}
	paid.CountersRemoved = counters.total
	// #789: the other direction — "Put a -1/-1 counter on this
	// creature" (Devoted Druid). After the removal so a cost that
	// did both would read left to right, and before the sacrifices
	// for the same reason the removal is: the source has to still be
	// on the battlefield.
	if ac := ab.Cost.AddCounter; ac != nil {
		if err := g.payCounterAddLocked(cardID, ac); err != nil {
			return err
		}
		paid.CountersAdded = ac.N
	}
	// Sacrifices last: they move cards, which invalidates `source`.
	// One payment is one simultaneous exit (#747, CR 603.10a).
	if err := g.payCostSacrificesLocked(sacrifices); err != nil {
		return err
	}
	// #1213: how many this announcement actually sacrificed, recorded
	// before the permanents are unreachable. Radiant Lotus's "for each
	// artifact sacrificed this way" reads it at resolution through
	// Context.Sacrificed().
	paid.Sacrificed = len(sacrifices)
	// #1213: the return-to-hand component, beside the sacrifices and
	// for the same reason — it moves cards, so it goes after every
	// component that needs the source still on the battlefield, and
	// before the stack item is built so the leaves-triggers it queues
	// are drained ABOVE the ability by the closing state-check pass
	// (CR 603.3b).
	if err := g.payReturnToHandCostLocked(playerID, cardID, params.ReturnIDs); err != nil {
		return err
	}
	source = nil
	// Discards last (#660). They move cards out of the hand, which
	// invalidates `source` for a DiscardSelf cost, and they go
	// through the ONE discard helper with cause cost — so
	// EventDiscardCard still fires per card, the CR 614 window still
	// runs over the exit, and CR 601.2h / 602.2b's indivisible step
	// is expressed by MustSettleNow rather than by a second loop
	// that knows not to prompt. See discard.go.
	//
	// Cycling emits EventCycle here too (CR 702.29b): the ability
	// has been paid for, so the card HAS been cycled, and the
	// watchers see it with the card already in the graveyard, which
	// is where CR 702.29c says it is.
	if err := g.payAbilityDiscardsLocked(playerID, cardID, ab, discards); err != nil {
		return err
	}
	// #1221: and the graveyard half. Last of all, because it moves
	// the source out of the graveyard and the effect that follows
	// reads it back out of exile. See exile_cost.go.
	if err := g.payAbilityExileSelfLocked(playerID, cardID, ab); err != nil {
		return err
	}

	// --- announce -----------------------------------------------
	itemID := uuid.New()
	if g.StackMeta == nil {
		g.StackMeta = make(map[uuid.UUID]*StackItem)
	}
	item := &StackItem{
		ID:           itemID,
		Kind:         StackItemActivated,
		Controller:   playerID,
		Owner:        playerID,
		SourceCardID: cardID,
		// CR 400.7: which OBJECT this ability came from, not just
		// which card. Read after the costs have been paid, because a
		// cost that moved the source (sacrifice, discard) has already
		// ended the object the ability belonged to and the stamp must
		// say so. See StackItem.SourceEpoch.
		SourceEpoch: g.cardObjectEpochLocked(cardID),
		Label:       ab.Label,
		Targets:     append([]TargetRef(nil), params.Targets...),
		Modes:       append([]int(nil), params.Modes...),
		// CR 602.2b: X was announced above and is locked here. The
		// effect reads it back through Context.X(), the same
		// accessor an X spell's OnResolve uses, and the wire ships
		// it on the stack item so the table can see what was paid.
		XValue: params.XValue,
		// #789: what the activation actually paid. An announced
		// counter count is a fact about the ANNOUNCEMENT, exactly as
		// X is, so it is locked here and read back at resolution
		// through Context.CountersRemoved — the counters are off the
		// board by then, so nothing downstream could recompute it.
		Paid:       paid,
		Effect:     ab.Effect,
		targetSpec: ab.Targets,
		modeSpec:   ab.Modes,
		Seq:        g.nextStackSeqLocked(),
	}
	g.StackMeta[itemID] = item
	// #628: activating an ability is a player decision, so it
	// restarts the CR 726 loop run. The announce emits EventTrigger
	// rather than an event of its own — the same kind a triggered
	// ability announces with — so the notch is here rather than in
	// turnTallyListener, which cannot tell the two apart.
	//
	// #810: every run but THIS ability's. A free, repeatable
	// activation is a loop whose every iteration is a player
	// decision, and clearing its own run was what kept the breaker
	// from ever seeing it. See notePlayerActivationLocked.
	g.notePlayerActivationLocked(TallyKey(cardID, ab.Label))
	// #1181: the activation record, in both scopes, written at the
	// ANNOUNCE and not at resolution — an exhaust ability countered on
	// the stack is still spent (CR 602.2b: activating an ability is
	// one indivisible step, and it has now finished). The key was
	// taken above, before the costs ran, so a sacrifice-this cost has
	// not moved the object out from under it.
	g.noteAbilityActivationLocked(activationKey)
	g.EmitEvent(Event{
		Kind:   EventTrigger,
		Actor:  playerID,
		Source: cardID,
		CardID: cardID,
	})
	// #1184: the same announcement, said in a kind anything may
	// WATCH. The EventTrigger above is the loop breaker's and the
	// turn tally's breadcrumb and the harvester returns on it
	// immediately (triggers.go), so "whenever you activate an exhaust
	// ability" had nothing to hang on; this event carries the
	// ability's identity — its label, the key the activation record
	// uses — and the exhaust bit, read off the shape rather than
	// re-derived later from a source a SacrificeSelf cost has already
	// ended.
	//
	// Emitted AFTER the record is written, so a watcher that reads
	// Game.Activations back sees this activation counted, and after
	// the item is on StackMeta, so the ability a trigger will sit
	// above already exists.
	//
	// #1223: StackItemID names the ability ITSELF, which Source and
	// CardID cannot — both name the permanent the ability came from,
	// and that permanent stays on the battlefield with any number of
	// its abilities on the stack at once. Rings of Brighthearth's
	// "copy THAT ability" needs a handle on the one that was just
	// activated, and the announcement is the only moment it is
	// unambiguous. Same field, same reason, as the became-target emit
	// one line below (see Event.StackItemID).
	g.EmitEvent(Event{
		Kind:        EventActivateAbility,
		Actor:       playerID,
		Source:      cardID,
		CardID:      cardID,
		StackItemID: itemID,
		Label:       ab.Label,
		Exhaust:     ab.Exhaust,
	})
	// CR 602.2b / 115.7: the ability's targets were chosen as it was
	// put on the stack. S22, for "whenever ~ becomes the target of a
	// spell or ability" (Monk Gyatso) — the clause names abilities as
	// well as spells, so the activated path has to emit too.
	g.emitBecameTargetLocked(playerID, cardID, itemID, params.Targets)
	// #758: the tap-another component's board half, paid HERE
	// rather than in the payment block above and for the reason the
	// cast path pays its convoke taps after the spell is on the
	// stack — a "whenever a permanent becomes tapped" payoff
	// (Opposition feeding a tap-watcher) has to sit ABOVE this
	// ability and resolve first (ADR 0020 §4, CR 603.3b). Every
	// named permanent was validated before anything was paid, so
	// there is nothing left that can fail.
	g.payTapOthersCostLocked(playerID, params.TapIDs)
	// The cost may have queued dies-triggers (a sacrifice outlet
	// feeding Blood Artist). Drain them so they sit ABOVE the
	// ability on the stack, which is where paying a cost puts them.
	g.runStateChecksLocked()
	return nil
}

// validateCrewCostLocked resolves a Crew cost into the concrete list
// of creatures to tap, without tapping anything (CR 702.122a).
//
// What it enforces, and what it deliberately does not:
//
//   - Every named card must be an untapped creature the activator
//     controls, and must be named once. Naming the same creature
//     twice would let one 3-power creature crew a 6.
//   - Total EFFECTIVE power must reach the crew number. Effective,
//     so an anthem and +1/+1 counters both count (CurrentPower), and
//     read at payment time (CR 702.122a) rather than at declaration.
//   - Summoning sickness is NOT checked. Tapping a creature to crew
//     is not paying a {T} cost, so a creature that arrived this turn
//     may crew (CR 702.122b). This is the rule most implementations
//     get wrong in the strict direction; getting it wrong the other
//     way is impossible here, because the check simply isn't made.
//   - Overshooting is legal. The printed number is a floor.
//
// The clause is "creatures you control", with no "another", so a
// Vehicle that something else has already animated is a legal
// crewer — of a different Vehicle, or in principle of itself. That
// is deliberately not special-cased: the rule says what it says, and
// self-crewing taps the Vehicle and strands it, which is a bad play
// rather than an illegal one.
//
// Caller must hold g.mu.
func (g *Game) validateCrewCostLocked(playerID uuid.UUID, cost AbilityCost, chosen []uuid.UUID) ([]uuid.UUID, error) {
	if cost.Crew <= 0 {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return nil, nil
	}
	if len(chosen) == 0 {
		return nil, ErrInsufficientCrew
	}
	// Effective characteristics have to be fresh: a creature that
	// gained +2/+2 this turn crews for its current power, not its
	// printed one.
	g.RecomputeLayersIfStaleLocked()
	seen := make(map[uuid.UUID]bool, len(chosen))
	total := 0
	out := make([]uuid.UUID, 0, len(chosen))
	for _, id := range chosen {
		if seen[id] {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return nil, ErrCardNotFound
		}
		if c.Controller != playerID {
			return nil, ErrCardCallerMismatch
		}
		if !c.IsCreature() {
			return nil, ErrNotACreature
		}
		if c.Tapped {
			return nil, ErrAlreadyTapped
		}
		total += c.CurrentPower()
		out = append(out, id)
	}
	if total < cost.Crew {
		return nil, ErrInsufficientCrew
	}
	return out, nil
}

// validateSacrificeCostLocked resolves the SacrificeSelf /
// SacrificeOther components into the concrete list of permanents to
// sacrifice, without moving anything. Caller must hold g.mu.
//
// #747 (ADR 0020 addendum §13): the clause's count is the spec's
// Min == Max — "Sacrifice two artifacts" is a clause with Max 2 — and
// the payment must name exactly that many permanents, each once, each
// on the battlefield under the payer's control (CR 701.21a) and each
// matching the clause. None may be the source when SacrificeSelf pays
// the source too. Any failure refuses the whole payment with nothing
// moved; all three cost sites (an activated ability, a mana ability,
// an additional cost to cast) share this one check.
//
// #1213: `x` is the value announced with the activation or the cast,
// and the count is read through SacrificeCostBounds rather than off
// the clause — so "Sacrifice one or more artifacts" accepts any number
// at or above its floor, and "Sacrifice X Treasures" accepts exactly
// the X that was announced. A FIXED clause reads exactly as it did and
// ignores `x` entirely; pass 0 from a site that has no X to announce.
func (g *Game) validateSacrificeCostLocked(playerID, sourceID uuid.UUID, cost AbilityCost, chosen []uuid.UUID, x int) ([]uuid.UUID, error) {
	var out []uuid.UUID
	if cost.SacrificeSelf {
		out = append(out, sourceID)
	}
	if cost.SacrificeOther == nil {
		if len(chosen) > 0 {
			return nil, ErrInvalidParam
		}
		return out, nil
	}
	// #1213: one predicate, shared with the view's picker bounds and
	// the enumerator's payment walk, so none of the three can offer
	// or accept a count the others refuse (#544).
	if !SacrificeCountLegal(cost.SacrificeOther, x, len(chosen)) {
		return nil, ErrInvalidParam
	}
	seen := make(map[uuid.UUID]bool, len(chosen))
	for _, id := range chosen {
		// One permanent pays one sacrifice. Naming it twice would let
		// a single Food pay "Sacrifice two Foods".
		if seen[id] {
			return nil, ErrInvalidParam
		}
		seen[id] = true
		c := findBattlefieldCard(g, id)
		if c == nil {
			return nil, ErrCardNotFound
		}
		// CR 701.21a — you can only sacrifice what you control.
		if c.Controller != playerID {
			return nil, ErrCardCallerMismatch
		}
		// specMatchLocked, not targetLegalLocked: sacrificing a
		// permanent to pay a cost does not target it (CR 601.2h), so
		// the CR 702 keyword gate must not apply — Carrion Feeder can
		// still eat your own hexproof creature.
		if !g.specMatchLocked(SourceChooser(playerID), cost.SacrificeOther, TargetRef{Kind: TargetCard, ID: id}, false) {
			return nil, ErrIllegalTarget
		}
		// Paying the same permanent twice (self-sacrifice plus the
		// same card as one of the "other" picks) isn't a legal cost
		// payment.
		if cost.SacrificeSelf && id == sourceID {
			return nil, ErrInvalidParam
		}
	}
	return append(out, chosen...), nil
}

// payAbilityManaCostLocked charges an activated ability's mana
// component through the same pool / auto-tap machinery a cast uses.
// Permissive mode (the default) leaves the pool alone and emits a
// cost warning, matching how S15 treats an unaffordable cast.
// Caller must hold g.mu.
// `spendCtx` describes the ability's SOURCE permanent, which is what
// a restricted token is matched against when the restriction says
// "activate abilities of …" (#352).
// `sourceName` is the permanent the ability is printed on, read only
// for the error message a malformed Phyrexian claim returns.
//
// It returns the PaidCost's mana half (#761): the tokens that left
// the pool, or OnPaper when permissive mode waived the charge, plus
// LifePaid for the Phyrexian symbols paid with life (#917). An
// ability item records the same fact a spell does, so a "for each
// colour of mana spent" ability would read it the same way — and
// Jeweled Amulet's "spend this mana only to cast" rider will, when
// the rider half lands.
// `cost` arrives already parsed and already priced through the
// CR 601.2f pass (#1184: AbilityManaCostForEffect), because the
// number an activation pays and the number the legal-move enumerator
// checks affordability against have to be computed by one function,
// and that function needs the source card and the ability shape this
// one no longer has.
func (g *Game) payAbilityManaCostLocked(p *Player, sourceID uuid.UUID, sourceName string, cost ParsedCost, params ActivateAbilityParams, spendCtx ManaSpendContext, excluded map[uuid.UUID]bool) (PaidCost, error) {
	var paid PaidCost
	// CR 107.4f / CR 602.2b (#917): the Phyrexian symbols the
	// activator announced they are paying with life leave the mana
	// cost here, through the SAME helper the cast path runs, and the
	// life is paid below — after the mana half is known to be
	// payable, so a refused activation never costs a point. Validated
	// in every mode, because an over-claim is a malformed announce
	// rather than a mana-gate failure.
	//
	// It happens before the auto-tap branch for the reason
	// applyAutoTapLocked strikes before planning: tapping a land for
	// a pip the activator said they would pay with 2 life is
	// stranding it.
	cost, phyrexianLife, err := g.strikePhyrexianLifeLocked(p, sourceName, cost, spendCtx, params.PhyrexianLife, &paid)
	if err != nil {
		return paid, err
	}
	// The announced X multiplies into the generic demand exactly as
	// it does for a cast: cost.Generic + cost.XSlots*x. Treasure
	// Vault's "{X}{X}" has two slots, so X=3 costs six.
	x := params.XValue
	if !params.Strict && !params.AutoTap {
		if !p.ManaPool.CanPayFor(cost, x, spendCtx) {
			g.EmitEvent(Event{
				Kind:   EventCostWarning,
				Actor:  p.ID,
				Source: sourceID,
			})
			// #761: the charge was waived, and the record says so
			// rather than reading as "nothing was spent". The LIFE
			// half is still paid — a life total is engine state in
			// every mode, and permissive mode's bargain is only about
			// the mana the player tracks on paper.
			paid.OnPaper = true
			return paid, g.payPhyrexianLifeLocked(sourceID, p.ID, phyrexianLife)
		}
		if err := g.payPhyrexianLifeLocked(sourceID, p.ID, phyrexianLife); err != nil {
			return paid, err
		}
		spent, _ := p.ManaPool.SpendManaFor(cost, x, spendCtx)
		paid.Mana = spent
		g.EmitEvent(manaSpentEvent(p.ID, sourceID, spent))
		return paid, nil
	}
	if params.AutoTap && !p.ManaPool.CanPayFor(cost, x, spendCtx) {
		plan, ok := g.autoTapLocked(p.ID, cost, x, excluded)
		if !ok {
			return paid, &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, x, spendCtx)}
		}
		g.materializePlanLocked(p, plan, cost)
	}
	if !p.ManaPool.CanPayFor(cost, x, spendCtx) {
		return paid, &InsufficientManaError{Missing: p.ManaPool.MissingFor(cost, x, spendCtx)}
	}
	// CR 602.2b pays every component of an activation together, so
	// the order is an engine-safety choice, not a rules one: the
	// FALLIBLE half goes first. PayLifeForEffect can still refuse
	// after the life total was checked (a replacement that stops the
	// player losing life at all, CR 119.8), and refusing after the
	// pool was emptied would charge for an activation that did not
	// happen.
	if err := g.payPhyrexianLifeLocked(sourceID, p.ID, phyrexianLife); err != nil {
		return paid, err
	}
	spent, _ := p.ManaPool.SpendManaFor(cost, x, spendCtx)
	paid.Mana = spent
	g.EmitEvent(manaSpentEvent(p.ID, sourceID, spent))
	return paid, nil
}
