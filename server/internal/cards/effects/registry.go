package effects

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// registry is the package-level card-effect catalog. Populated at
// init() time by each per-card file calling Register(Spec{...}).
// Reads (Lookup / All) are zero-lock because Go's init() ordering
// serialises writes before any reader touches the map. Post-init
// mutation is disallowed — tests that need to inject specs call
// registerForTest, which is only visible within the test binary.
var registry = map[string]Spec{}

// Register adds a catalog entry. Called once per card at init()
// time. Panics on duplicate OracleID — a collision means a card
// file was copy-pasted without updating the key, which is a logic
// bug we want to surface loudly at server boot rather than
// silently letting one spec win.
//
// Empty OracleID is rejected for the same reason: a spec without
// a key would shadow Lookup() results for the empty string and
// mask bugs in the caller.
func Register(spec Spec) {
	if spec.OracleID == "" {
		panic(fmt.Sprintf("effects.Register: empty OracleID on %q", spec.Name))
	}
	if existing, ok := registry[spec.OracleID]; ok {
		panic(fmt.Sprintf("effects.Register: duplicate OracleID %s (existing %q, new %q)",
			spec.OracleID, existing.Name, spec.Name))
	}
	if spec.Modes != nil {
		if spec.Targets != nil {
			panic(fmt.Sprintf("effects.Register: %q declares both Targets and Modes — put the target clause on the mode", spec.Name))
		}
		// #764 retired the "one targeted option per cast" panic that
		// stood here: per-mode target slots exist now, so Kolaghan's
		// Command is declarable. What replaced it are the two shapes
		// the new machinery genuinely cannot read.
		if spec.Modes.Repeatable && spec.Modes.Max == 1 {
			panic(fmt.Sprintf("effects.Register: %q is Repeatable with Max 1 — there is nothing to repeat", spec.Name))
		}
		for i, o := range spec.Modes.Options {
			if o.Label == "" {
				panic(fmt.Sprintf("effects.Register: %q mode %d has no label — the bullet is the whole of what the picker shows", spec.Name, i))
			}
			checkFlatClauses(spec.Name, o.Targets)
		}
	}
	checkFlatClauses(spec.Name, spec.Targets)
	checkExhaustAbilities(spec)
	checkPlayerKeywords(spec)
	for _, a := range spec.Activated {
		checkFlatClauses(spec.Name, a.Targets)
		if a.Modes != nil {
			if a.Targets != nil {
				panic(fmt.Sprintf("effects.Register: %q declares an activated ability with both Targets and Modes — put the target clause on the mode", spec.Name))
			}
			for i, o := range a.Modes.Options {
				if o.Label == "" {
					panic(fmt.Sprintf("effects.Register: %q activated mode %d has no label", spec.Name, i))
				}
				checkFlatClauses(spec.Name, o.Targets)
			}
		}
	}
	for _, t := range spec.Triggered {
		checkFlatClauses(spec.Name, t.Targets)
		if t.Modes != nil && t.Targets != nil {
			panic(fmt.Sprintf("effects.Register: %q declares a trigger with both Targets and Modes — put the target clause on the mode", spec.Name))
		}
	}
	// S22: an alternative cost is claimed by name on the wire, so a
	// blank or duplicated key is unaddressable — the cast either
	// can't name it or names two of them. Both are copy-paste
	// mistakes, and both fail loudly at boot rather than as a
	// mysteriously-rejected cast mid-game.
	seenAlt := make(map[string]bool, len(spec.AlternativeCosts))
	for _, ac := range spec.AlternativeCosts {
		if ac.Key == "" {
			panic(fmt.Sprintf("effects.Register: %q declares an alternative cost with no Key", spec.Name))
		}
		if seenAlt[ac.Key] {
			panic(fmt.Sprintf("effects.Register: %q declares two alternative costs keyed %q", spec.Name, ac.Key))
		}
		seenAlt[ac.Key] = true
		// S29: an offer bound to a zone the card cannot be cast from
		// is unclaimable — the cast path rejects the zone before it
		// ever looks at the price. A card file that wrote one meant
		// to list the zone as well, and finding out at boot is far
		// cheaper than finding out when a flashback button never
		// appears.
		if ac.FromZone != "" && ac.FromZone != game.ZoneHand && !zoneDeclared(spec.CastableZones, ac.FromZone) {
			panic(fmt.Sprintf("effects.Register: %q offers %q from %s but does not list that zone in CastableZones",
				spec.Name, ac.Key, ac.FromZone))
		}
		if ac.FaceDown == nil {
			continue
		}
		// ADR 0082, CR 708.4: the face-down cast. Three boot checks,
		// each for a shape that compiles and then behaves as
		// something the card does not print.
		//
		// A kind that is not a CR 708.2 object state would cast the
		// card into an EXILE state on the stack — an object with no
		// characteristics at all, which the resolution has no
		// meaning for.
		if !ac.FaceDown.Kind.IsPermanentState() {
			panic(fmt.Sprintf("effects.Register: %q offers %q with face-down kind %q, which is not a CR 708.2 object state — build it with Morph / Megamorph / Disguise",
				spec.Name, ac.Key, ac.FaceDown.Kind))
		}
		// The face-up cost is the half of the keyword the CARD
		// prints, and the only place the engine can read it from
		// once the permanent is face down and has no text
		// (ADR 0082 decision 4). An unparseable one refuses at boot
		// rather than at the moment a player tries to turn a
		// permanent up they can no longer turn up.
		if ac.FaceDown.FaceUpCost != "" {
			if _, err := game.ParseCost(ac.FaceDown.FaceUpCost); err != nil {
				panic(fmt.Sprintf("effects.Register: %q offers %q with an unparseable face-up cost %q: %v",
					spec.Name, ac.Key, ac.FaceDown.FaceUpCost, err))
			}
		}
		// A face-down cast has no targets, no modes and no
		// additional costs, because the object it produces has no
		// text (CR 708.2a) — CastSpell stamps the state before every
		// one of those gates reads the catalog. An offer that
		// declared a target clause anyway would be a card file
		// expecting a clause the announce path can never reach.
		if ac.Targets != nil || ac.ClearsTargets {
			panic(fmt.Sprintf("effects.Register: %q offers %q with a target clause — a spell cast face down has no text and no targets (CR 708.2a)",
				spec.Name, ac.Key))
		}
	}
	// #659: a card may not declare exile castable. S29 allowed it "for
	// the shape suspend and foretell will use"; they do not use it and
	// could not, because a CARD-level declaration opens exile for
	// every copy of the card at any time, however the copy got there.
	// Every exile cast in this engine is a per-instance
	// game.CastPermission (ADR 0066). Refused at boot so the retired
	// shape cannot come back through a card file.
	for _, z := range spec.CastableZones {
		if z == game.ZoneExile {
			panic(fmt.Sprintf("effects.Register: %q declares ZoneExile in CastableZones — an exile cast is a per-instance game.CastPermission, never a card-level declaration (#659, ADR 0066)",
				spec.Name))
		}
	}
	// ADR 0073: the optional additional costs. Six boot checks, each
	// for a shape that compiles and then behaves as something the card
	// does not print.
	if spec.AdditionalCost != nil && spec.AdditionalCost.Optional {
		panic(fmt.Sprintf("effects.Register: %q puts an Optional cost in AdditionalCost — the mandatory slot is never optional; declare it in OptionalCosts", spec.Name))
	}
	seenOptional := make(map[string]bool, len(spec.OptionalCosts))
	for i, oc := range spec.OptionalCosts {
		// Without the flag the cast path prices it and never offers
		// the choice, which is a card that always kicks itself.
		if !oc.Optional {
			panic(fmt.Sprintf("effects.Register: %q optional cost %d is not marked Optional — build it with Kicker / Multikicker / Buyback, never by hand", spec.Name, i))
		}
		// The Key is what OnResolve and the buyback route ask for. A
		// blank or duplicated one is unaddressable.
		if oc.Key == "" {
			panic(fmt.Sprintf("effects.Register: %q declares an optional cost with no Key", spec.Name))
		}
		if seenOptional[oc.Key] {
			panic(fmt.Sprintf("effects.Register: %q declares two optional costs keyed %q", spec.Name, oc.Key))
		}
		seenOptional[oc.Key] = true
		if oc.ManaCost != "" {
			if _, err := game.ParseCost(oc.ManaCost); err != nil {
				panic(fmt.Sprintf("effects.Register: %q declares an unparseable optional cost %q: %v", spec.Name, oc.ManaCost, err))
			}
		}
		// #1224: an AdditionalCost in OptionalCosts carries the same
		// Sacrifice *TargetSpec the mandatory slot does (Constant Mists'
		// "Buyback—Sacrifice a land"), and it reaches
		// validateAdditionalCostLocked through the same plan and the same
		// flat sacrifice_ids walk — so every shape the guard refuses on
		// the mandatory slot below is refusable here too.
		// #1213: an optional cost is a CAST cost, so neither variable
		// shape has a shape here — the flat payment lists are walked
		// in plan order and need a fixed width, exactly as the
		// mandatory slot above.
		checkSacrificeClause(spec.Name, fmt.Sprintf("optional cost %q", oc.Key), oc.Sacrifice, false, false)
		if oc.Empty() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q demands nothing", spec.Name, oc.Key))
		}
		// ADR 0073 §4: only a mana-only cost may be paid more than
		// once. Every printed multikicker is mana, and N card-shaped
		// payments per cast is a wire shape nothing asks for.
		if oc.MaxPayments() > 1 && oc.CardsDemanded() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q repeats and demands cards or permanents — only a mana-only cost may repeat", spec.Name, oc.Key))
		}
		// PayLifeX rides the shared XValue slot (ADR 0021 §3), and two
		// claimants on one number is a bug waiting to be written.
		if oc.PayLifeX {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q pays X life — an optional cost cannot claim the shared X slot", spec.Name, oc.Key))
		}
	}
	// S22: a tap-permanents cost with no pool of legal permanents can
	// never be paid, and one whose extra cost doesn't parse would
	// silently charge nothing — both are copy-paste mistakes in a
	// card file, and both fail at boot rather than mid-game.
	if tc := spec.TapCost; tc != nil {
		if tc.Key == "" || tc.Spec == nil {
			panic(fmt.Sprintf("effects.Register: %q declares a tap cost with no key or no legal permanents — build it with Convoke() or Waterbend()", spec.Name))
		}
		if tc.Extra != "" {
			if _, err := game.ParseCost(tc.Extra); err != nil {
				panic(fmt.Sprintf("effects.Register: %q declares an unparseable tap cost %q: %v", spec.Name, tc.Extra, err))
			}
		}
	}
	// ADR 0073 §7: a cast condition with no printed clause produces a
	// refusal the client cannot explain, and a clause with no
	// condition refuses nothing while claiming to. Same for a
	// restriction: the Label IS the message the player is shown.
	if (spec.CastCondition == nil) != (spec.CastConditionLabel == "") {
		panic(fmt.Sprintf("effects.Register: %q declares a CastCondition without its printed CastConditionLabel, or the label without the condition", spec.Name))
	}
	for i, r := range spec.CastRestrictions {
		if r.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q cast restriction %d has no printed Label — the refusal carries it to the client", spec.Name, i))
		}
		if r.Forbids == nil {
			panic(fmt.Sprintf("effects.Register: %q cast restriction %q forbids nothing", spec.Name, r.Label))
		}
	}
	// #1210, ADR 0073's amendment of 2026-09-22: the same two checks
	// for the activation twin, and for the same reason — the Label is
	// what the greyed ability row shows the player, and a restriction
	// with no Forbids claims to refuse and refuses nothing.
	for i, r := range spec.ActivationRestrictions {
		if r.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q activation restriction %d has no printed Label — the refusal carries it to the client", spec.Name, i))
		}
		if r.Forbids == nil {
			panic(fmt.Sprintf("effects.Register: %q activation restriction %q forbids nothing", spec.Name, r.Label))
		}
	}
	// #1208, ADR 0066's and ADR 0073's amendments of 2026-09-23: the
	// same two checks again for an activation-timing statement, plus
	// the one CastTimings needs — a statement that says nothing.
	// TimingNormal is the zero value and the read ignores it, so a
	// Spec slot carrying one is a card file that meant to say
	// something and failed SILENTLY. The Label is what the next
	// reader matches against the oracle text; a nil Covers is a
	// statement about nothing.
	for i, t := range spec.ActivationTimings {
		if t.Timing == game.TimingNormal {
			panic(fmt.Sprintf("effects.Register: %q activation timing %d says nothing — set TimingFlash, TimingSorcery or TimingYourTurnOnly", spec.Name, i))
		}
		if t.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q activation timing %d has no printed Label", spec.Name, i))
		}
		if t.Covers == nil {
			panic(fmt.Sprintf("effects.Register: %q activation timing %q covers nothing", spec.Name, t.Label))
		}
	}
	// #1195: the same bargain for a timing statement. TimingNormal is
	// the zero value and says nothing, so a Spec slot carrying one is
	// a card file that meant to say something and did not — and the
	// failure would be silent, because the read ignores it. The Label
	// is what the log prints.
	for i, ct := range spec.CastTimings {
		if ct.Timing == game.TimingNormal {
			panic(fmt.Sprintf("effects.Register: %q cast timing %d says nothing — set TimingFlash, TimingSorcery or TimingYourTurnOnly", spec.Name, i))
		}
		if ct.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q cast timing %d has no printed Label", spec.Name, i))
		}
	}
	// ADR 0048 addendum §11: no printed card sets a floor on its own
	// cost, and an untested kind should not be declarable. A mana Unit
	// belongs on an increase only (open question 3), and carries only
	// generic and single-colour symbols (§16); the engine would refuse
	// every cast of a card that declared any other shape, so say so at
	// boot instead.
	for i, m := range spec.SelfCostModifiers {
		if m.Kind == game.CostFloor {
			panic(fmt.Sprintf("effects.Register: %q self cost modifier %d is a CostFloor — a spell's own cost modifier increases or reduces", spec.Name, i))
		}
	}
	for _, mods := range [][]game.CostModifier{spec.CostModifiers, spec.SelfCostModifiers} {
		for i, m := range mods {
			if why := m.UnitProblem(); why != "" {
				panic(fmt.Sprintf("effects.Register: %q cost modifier %d (%q) declares %s (ADR 0048 addendum §16)", spec.Name, i, m.Label, why))
			}
		}
	}
	// The completeness declaration is published verbatim on the
	// public catalog page, so the two ways of getting it wrong are
	// both caught at boot rather than shipped to a reader.
	//
	// Note what is NOT checked: an absent declaration. The zero
	// value means "unreviewed", which is a legal and honest thing
	// for a spec to say — completeness.go explains why a hard gate
	// would make the catalog less truthful, not more.
	if spec.Completeness == CompletenessCaveats && len(spec.Caveats) == 0 {
		panic(fmt.Sprintf("effects.Register: %q declares CompletenessCaveats with no Caveats — say what the caveat is", spec.Name))
	}
	if spec.Completeness != CompletenessCaveats && len(spec.Caveats) > 0 {
		panic(fmt.Sprintf("effects.Register: %q lists Caveats but declares %s — use CompletenessCaveats", spec.Name, spec.Completeness))
	}
	for _, cv := range spec.Caveats {
		if cv == "" {
			panic(fmt.Sprintf("effects.Register: %q declares an empty caveat", spec.Name))
		}
	}
	// ADR 0071 decision 3: the Room door gate is RESERVED, not built.
	// game.Card has no unlocked state, so Designation.Active answers
	// false for it — a card that declared one would ship with that
	// ability silently switched off forever, which is exactly the
	// half-a-card failure ADR 0037 §5 forbids. #886 lifts this in the
	// same change that adds the state.
	for _, d := range specDesignations(spec) {
		if d.Kind == game.DesignationDoorUnlocked {
			panic(fmt.Sprintf("effects.Register: %q gates an ability on an unlocked Room door, which is designed but not built (ADR 0071 decision 3, #886)", spec.Name))
		}
	}
	// ADR 0062 Decision 4: a special action the engine cannot carry
	// out would take a card out of a hand and do nothing with it, so
	// the declaration fails at boot rather than mid-game. The cost
	// is parsed here for the same reason an ability's is: an
	// unparseable one is refused at payment time, which is after the
	// timing check has already said yes.
	for i, sa := range spec.SpecialActions {
		if !game.SpecialActionKindBuilt(sa.Kind) {
			panic(fmt.Sprintf("effects.Register: %q special action %d declares kind %q, which the engine cannot carry out (ADR 0062 Decision 4)",
				spec.Name, i, sa.Kind))
		}
		if sa.Cost != "" {
			if _, err := game.ParseCost(sa.Cost); err != nil {
				panic(fmt.Sprintf("effects.Register: %q special action %d declares an unparseable cost %q: %v",
					spec.Name, i, sa.Cost, err))
			}
		}
		if sa.CastCost != "" {
			if _, err := game.ParseCost(sa.CastCost); err != nil {
				panic(fmt.Sprintf("effects.Register: %q special action %d declares an unparseable cast cost %q: %v",
					spec.Name, i, sa.CastCost, err))
			}
		}
		if sa.Kind == game.SpecialActionSuspend && sa.Counters <= 0 {
			panic(fmt.Sprintf("effects.Register: %q suspends with %d time counters — suspend N is at least one (CR 702.62a)",
				spec.Name, sa.Counters))
		}
	}
	// #657 / CR 702.35a: the madness cost is the price of a cast the
	// engine will offer, so an unparseable one is refused at boot
	// rather than at the moment the offer is taken — which is after
	// the card has already been exiled and cannot go back.
	if spec.Madness != "" {
		if _, err := game.ParseCost(spec.Madness); err != nil {
			panic(fmt.Sprintf("effects.Register: %q declares an unparseable madness cost %q: %v",
				spec.Name, spec.Madness, err))
		}
	}
	// An activated ability's mana component is the only place an X
	// can live (game.AbilityCost.DemandsX says why), so both ways of
	// getting a variable cost wrong are visible from here, and both
	// fail at boot rather than as a mysteriously-refused activation
	// mid-game.
	for i, ab := range spec.Activated {
		if ab.Cost.Mana != "" {
			if _, err := game.ParseCost(ab.Cost.Mana); err != nil {
				panic(fmt.Sprintf("effects.Register: %q ability %d declares an unparseable mana cost %q: %v",
					spec.Name, i, ab.Cost.Mana, err))
			}
		}
		if ab.Cost.MinX < 0 {
			panic(fmt.Sprintf("effects.Register: %q ability %d sets a negative MinX %d", spec.Name, i, ab.Cost.MinX))
		}
		checkCounterCost(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.RemoveCounters, ab.Cost.AddCounter)
		checkSacrificeClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.SacrificeOther, true, true)
		checkReturnClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.ReturnToHand)
		// #660: a discard clause that discards nothing would make
		// the ability free, the way a zero-counter cost would.
		if dc := ab.Cost.DiscardCards; dc != nil && dc.N <= 0 {
			panic(fmt.Sprintf("effects.Register: %q ability %d discards %d cards — a discard cost discards at least one",
				spec.Name, i, dc.N))
		}
		// CR 113.6 / ADR 0062 Decision 1: an ability that functions
		// somewhere other than the battlefield has no permanent to
		// tap, sacrifice, crew or put loyalty counters on. Such a
		// cost could never be paid, so it fails at boot naming the
		// component rather than as a mysteriously-refused activation
		// mid-game — the treatment MinX-without-{X} already gets.
		for _, z := range ab.Zones {
			if z == game.ZoneBattlefield {
				continue
			}
			if why := game.AbilityNeedsPermanentSource(ab.Cost); why != "" {
				panic(fmt.Sprintf("effects.Register: %q ability %d functions from the %s but declares %s — that component needs a permanent on the battlefield",
					spec.Name, i, z, why))
			}
		}
		// DiscardSelf is cycling's component (CR 702.29a) and it
		// discards the source, so the source has to be a card in a
		// hand. A battlefield ability that declared it would have
		// nothing to discard.
		if ab.Cost.DiscardSelf && !zoneDeclared(ab.Zones, game.ZoneHand) {
			panic(fmt.Sprintf("effects.Register: %q ability %d declares a discard-this cost but does not function from the hand — build it with Cycling / Typecycling",
				spec.Name, i))
		}
		// #1213: two claimants on one announced X. A cost whose mana
		// component carries {X} AND whose sacrifice clause counts
		// from X would have to spend one number on both, and the
		// engine would silently take whichever the first reader
		// asked for. The same refusal ADR 0021 §3 makes for
		// PayLifeX, one component over; no printed card does it.
		if ab.Cost.XSlots() > 0 && game.SacrificeCountFromX(ab.Cost.SacrificeOther) {
			panic(fmt.Sprintf("effects.Register: %q ability %d has {X} in its mana cost %q AND sacrifices X permanents — one announced X cannot pay both",
				spec.Name, i, ab.Cost.Mana))
		}
		// #1221: the same rule one zone over. ExileSelf is scavenge's
		// and embalm's "Exile this card from YOUR GRAVEYARD"
		// (CR 702.96a, CR 702.128a), so an ability that declares it
		// without declaring the graveyard could never pay it — and
		// would look complete on the catalog page while refusing
		// every activation.
		if ab.Cost.ExileSelf && !zoneDeclared(ab.Zones, game.ZoneGraveyard) {
			panic(fmt.Sprintf("effects.Register: %q ability %d declares an exile-this cost but does not function from the graveyard — build it with Scavenge / Embalm / Eternalize",
				spec.Name, i))
		}
		if ab.Cost.MinX > 0 && !ab.Cost.DemandsX() {
			panic(fmt.Sprintf("effects.Register: %q ability %d sets MinX %d but its cost %q has no {X} — a floor on a variable that cannot vary makes the ability unactivatable",
				spec.Name, i, ab.Cost.MinX, ab.Cost.Mana))
		}
	}
	for i, ma := range spec.ManaAbilities {
		// #1213: `false` — a mana ability has no stack item and no
		// announced X (CR 605.3b), so a "Sacrifice X …" clause there
		// has nothing to read its count from.
		checkSacrificeClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.SacrificeOther, true, false)
		// #1228 / CR 113.6: the MANA half of the zone dimension, held
		// to the same three rules the activated half is held to —
		// every one of them a boot-time refusal rather than a
		// mysteriously-dead card.
		checkManaAbilityZones(spec.Name, i, ma)
		// #789: the counter components are one declaration with two
		// owners, so they are checked by one function in both places.
		checkCounterCost(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.RemoveCounters, ma.Cost.AddCounter)
		// #1213 / #1283: a card-picking clause that picks nothing would
		// make the ability free — the refusal the CR 602 discard gets.
		if dc := ma.Cost.DiscardCards; dc != nil && dc.N <= 0 {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d discards %d cards — a discard cost discards at least one",
				spec.Name, i, dc.N))
		}
		if ec := ma.Cost.ExileCards; ec != nil && ec.N <= 0 {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d exiles %d cards from hand — an exile cost exiles at least one",
				spec.Name, i, ec.N))
		}
		if ma.Cost.Mana != "" {
			if _, err := game.ParseCost(ma.Cost.Mana); err != nil {
				panic(fmt.Sprintf("effects.Register: %q mana ability %d declares an unparseable mana cost %q: %v",
					spec.Name, i, ma.Cost.Mana, err))
			}
		}
	}
	if spec.AdditionalCost != nil {
		checkSacrificeClause(spec.Name, "additional cost", spec.AdditionalCost.Sacrifice, false, false)
	}
	// #801: a replacement's per-instance ReplacementEffectID packs the
	// source's battlefield index and its slot in this slice into one
	// number, with game.MaxCatalogReplacementSlots as the stride. A
	// Spec over the budget would mint IDs belonging to the next
	// permanent along — a CR 616 prompt ordering some other card's
	// effect — so it fails at boot rather than aliasing in play. The
	// fullest entry in the catalog today declares two slots.
	if n := len(spec.Replacements); n > game.MaxCatalogReplacementSlots {
		panic(fmt.Sprintf("effects.Register: %q declares %d replacement effects — the ID scheme reserves %d slots per card (game.MaxCatalogReplacementSlots)",
			spec.Name, n, game.MaxCatalogReplacementSlots))
	}
	// #925: a triggered ability may declare the zone it watches from
	// (CR 113.6). A zone the harvest does not walk would be a
	// declaration the engine silently ignored — the card would
	// register, look complete on the catalog page, and never fire —
	// so it fails at boot with the reason, exactly as an
	// unclaimable alternative cost does above. An emblem's triggers
	// are checked too: they are harvested from the command zone by
	// their own walk (#623), so a Zones on one of them would be
	// ignored just as silently.
	checkTriggerZones(spec.Name, "trigger", spec.Triggered)
	// #1221: the same boot-time refusal for the STATIC half of
	// CR 113.6. A zone the layer gather does not walk would be a
	// declaration the engine silently ignored — the card would
	// register, look complete on the catalog page, and never apply.
	checkStaticZones(spec.Name, "static", spec.Static)
	if spec.Emblem != nil {
		checkTriggerZones(spec.Name, "emblem trigger", spec.Emblem.Triggered)
	}
	checkEmblemSpec(spec.Name, spec.Emblem)
	checkGrants(spec.Name, spec.Grants)
	registry[spec.OracleID] = spec
	def := buildDef(spec)
	defs[spec.OracleID] = def
	// #925 + #659: the index has to see the triggers the ENGINE will
	// harvest, not the ones the card file wrote. A suspend
	// declaration grows the exile countdown in buildDef — the keyword
	// owns it, not the card — and an index built from spec.Triggered
	// would never walk exile for it.
	game.IndexTriggerZones(spec.OracleID, def.Triggered)
	// #1221: and the static half of the same index. From the SPEC
	// rather than from the def, because nothing in buildDef grows or
	// rewrites a static the way a suspend declaration grows a
	// trigger — Spec.Static is what the layer pass gathers.
	game.IndexStaticZones(spec.OracleID, spec.Static)
	// #623 / CR 114: a card that makes an emblem files a SECOND def
	// for the emblem object, under "emblem:<this key>". It goes in
	// `defs` and not in `registry`, so the engine finds the emblem's
	// abilities through the ordinary CatalogLookup and the card
	// census keeps counting cards — an emblem is not a card
	// (CR 114.4). See emblem.go and ADR 0064.
	if spec.Emblem != nil {
		defs[game.EmblemKey(spec.OracleID)] = buildEmblemDef(*spec.Emblem)
	}
	// #665 / CR 707.9a: a card whose copy effect GRANTS an ability
	// files that ability's own def, under "grant:<key>". Same map,
	// same reason as the emblem: the copy carries the key in its
	// copiable values and the engine finds the abilities through the
	// ordinary CatalogLookup, while the census keeps counting cards —
	// a granted ability is not one.
	for _, gr := range spec.Grants {
		defs[game.GrantKey(gr.Key)] = buildGrantDef(gr)
	}
}

// checkManaAbilityZones is #1228's registration guard for a CR 605
// mana ability's CR 113.6 declaration. Three rules, and each one
// makes an otherwise-silent failure loud at boot:
//
//  1. a declared zone has to be one every consumer walks
//     (game.ManaAbilityZoneUnsupported) — otherwise the card
//     registers, looks complete on the catalog page, and never
//     offers the ability anywhere;
//  2. a non-battlefield declaration may not carry a component only a
//     permanent could pay (game.ManaAbilityNeedsPermanentSource) —
//     the same refusal an activated ability's Zones already gets, and
//     the same reason: the cost could never be paid;
//  3. an exile-this cost and a non-battlefield zone imply each other.
//     Without the zone there is nothing to exile FROM; without the
//     cost the ability is a free, repeatable mana source, which is
//     not a card anybody printed.
func checkManaAbilityZones(card string, i int, ma ManaAbility) {
	offBattlefield := false
	for _, zone := range ma.Zones {
		if why := game.ManaAbilityZoneUnsupported(zone); why != "" {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d functions from %s — %s",
				card, i, zone, why))
		}
		if zone == game.ZoneBattlefield {
			continue
		}
		offBattlefield = true
		if why := game.ManaAbilityNeedsPermanentSource(shapeOfManaAbility(ma)); why != "" {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d functions from the %s but declares %s — that component needs a permanent on the battlefield",
				card, i, zone, why))
		}
	}
	if ma.Cost.ExileSelf && !offBattlefield {
		panic(fmt.Sprintf("effects.Register: %q mana ability %d declares an exile-this cost but does not function from a non-battlefield zone — build it with ExileFromHandForMana",
			card, i))
	}
	if offBattlefield && !ma.Cost.ExileSelf {
		panic(fmt.Sprintf("effects.Register: %q mana ability %d functions off the battlefield but exiles nothing — a mana ability with no cost is a free repeatable source; build it with ExileFromHandForMana",
			card, i))
	}
}

// shapeOfManaAbility projects the declared cost onto the game-package
// shape the zone rules are written against, so the boot check asks
// game.ManaAbilityNeedsPermanentSource exactly the question the
// engine will ask of the built ability. Only the cost components
// matter here; the produced-mana half is not a zone question.
func shapeOfManaAbility(ma ManaAbility) game.ManaAbilityShape {
	return game.ManaAbilityShape{
		Zones:          ma.Zones,
		TapCost:        ma.Cost.Tap,
		SacrificeCost:  ma.Cost.Sacrifice,
		SacrificeOther: ma.Cost.SacrificeOther,
		RemoveCounters: ma.Cost.RemoveCounters,
		AddCounter:     ma.Cost.AddCounter,
		ExileSelf:      ma.Cost.ExileSelf,
	}
}

// checkTriggerZones is #925's registration guard: every zone a
// triggered ability declares has to be one the harvest actually
// walks, or the ability is dead text the catalog page would still
// call complete.
// checkStaticZones is checkTriggerZones for the layer half of
// CR 113.6 (#1221): a static may declare the zone it functions from,
// and a zone activeStaticAbilitiesLocked does not gather is refused
// at boot with the reason.
func checkStaticZones(card, what string, statics []game.StaticAbility) {
	for i, s := range statics {
		for _, zone := range s.Zones {
			if why := game.StaticZoneUnsupported(zone); why != "" {
				panic(fmt.Sprintf("effects.Register: %q %s %d functions from %s — %s",
					card, what, i, zone, why))
			}
		}
	}
}

func checkTriggerZones(card, what string, triggers []game.TriggeredAbility) {
	for i, t := range triggers {
		for _, zone := range t.Zones {
			if why := game.TriggerZoneUnsupported(zone); why != "" {
				panic(fmt.Sprintf("effects.Register: %q %s %d watches from %s — %s",
					card, what, i, zone, why))
			}
		}
	}
}

// checkSacrificeClause is #747's registration guard, narrowed by
// #1213 (ADR 0020 addendum §12, ADR 0073's 2026-09-22 amendment).
//
// #747 accepted exactly one shape — a FIXED count of at least one,
// written Min == Max == N — because a variable count had no announced
// count and no record of what was paid. Both exist now
// (SacrificeCostBounds and PaidCost.Sacrificed), so two more shapes
// are legal:
//
//   - an OPEN count, Min ≥ 1 with Max 0: "Sacrifice one or more
//     artifacts" (Radiant Lotus). The activator names how many.
//   - CountFromX: "Sacrifice X Treasures" (Grim Hireling). The
//     announced X is the count on both sides.
//
// Everything still refused is a card-file mistake that would ship the
// card as something it does not print:
//
//   - a floor below one: a cost that can be paid with nothing is free.
//   - a ceiling below the floor.
//   - CountFromX where there is no X to announce — `allowX` is false
//     for a mana ability, which has no stack item to carry one
//     (CR 605.3b).
//   - AllowSame: one permanent cannot pay two sacrifices.
//   - Players: a player is not a permanent.
//
// Nil (no sacrifice component) is fine.
//
// Register calls it on the three sites the ADR names: spec.Activated,
// spec.ManaAbilities and spec.AdditionalCost. An ability granted at
// runtime (a static grant, an Equipment's "equipped creature has …")
// is built after registration and is not checked here. Every such
// grant with a sacrifice clause today is a count of one, so nothing
// escapes the guard yet; a grant with a variable count would.
// checkCounterCost holds a counter cost to the shapes the engine can
// actually pay, at boot rather than as a mysteriously-refused
// activation mid-game. One function for both owners (#789), because
// it is one component: an activated ability's and a mana ability's
// counter costs are the same struct and must be declared the same
// way.
//
// Four refusals, each naming a card file mistake that would
// otherwise ship a card stronger or weaker than printed:
//
//   - a fixed cost that removes nothing makes the ability free;
//   - "a counter" of any kind with N > 1, off ONE permanent, has no
//     single kind to name at announce (#625). Split across
//     permanents it has one: #943 asks the kind per part, so an
//     any-kind AMONG cost is a shape (Tekuthal, Inquiry Dominus) and
//     is no longer refused here;
//   - Among without a From clause names no permanents to split
//     across, and Among with Variable is a shape no card prints;
//   - an add-a-counter cost with no kind, or none to add, would put
//     nothing on and make the ability free.
func checkCounterCost(card, where string, rc *game.CounterRemovalCost, ac *game.CounterAddCost) {
	if rc != nil {
		switch {
		case rc.Variable && rc.N < 0:
			panic(fmt.Sprintf("effects.Register: %q %s sets a negative floor %d on a variable counter cost", card, where, rc.N))
		case !rc.Variable && rc.N <= 0:
			panic(fmt.Sprintf("effects.Register: %q %s removes %d counters — a counter cost removes at least one", card, where, rc.N))
		}
		if rc.Counter == "" && !rc.Among && (rc.N > 1 || rc.Variable) {
			panic(fmt.Sprintf("effects.Register: %q %s removes %d counters of any kind off one permanent — only \"a counter\" (N = 1) has an any-kind shape there; a removal that may mix kinds is spread across permanents (Among, #943)", card, where, rc.N))
		}
		if rc.Among {
			if rc.From == nil {
				panic(fmt.Sprintf("effects.Register: %q %s splits a counter removal among nothing — an Among cost needs its \"from among …\" clause", card, where))
			}
			if rc.Variable {
				panic(fmt.Sprintf("effects.Register: %q %s removes a variable number of counters from among several permanents — no printed card does, and it has no shape", card, where))
			}
		}
	}
	if ac != nil {
		if ac.Counter == "" {
			panic(fmt.Sprintf("effects.Register: %q %s puts a counter of no kind on as a cost", card, where))
		}
		if ac.N <= 0 {
			panic(fmt.Sprintf("effects.Register: %q %s puts %d counters on as a cost — it has to put at least one on", card, where, ac.N))
		}
	}
}

func checkSacrificeClause(card, where string, spec *game.TargetSpec, allowOpen, allowX bool) {
	if spec == nil {
		return
	}
	switch {
	case spec.CountFromX && !allowX:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices X permanents, but there is no X to announce here (CR 605.3b / the flat cast payment list, #1213)", card, where))
	case game.SacrificeCostVariable(spec) && !spec.CountFromX && !allowOpen:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents, but a variable count has no shape here — a cast's payment lists are walked in plan order and need a fixed width (#1213)",
			card, where, spec.Min, spec.Max))
	case !spec.CountFromX && spec.Min < 1:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents — a sacrifice cost pays at least one, so its floor is at least one (SacrificeN, SacrificeOneOrMore)",
			card, where, spec.Min, spec.Max))
	case !spec.CountFromX && spec.Max != 0 && spec.Max < spec.Min:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents — the ceiling is below the floor",
			card, where, spec.Min, spec.Max))
	case spec.AllowSame:
		panic(fmt.Sprintf("effects.Register: %q %s lets one permanent pay a sacrifice twice (AllowSame)", card, where))
	case spec.Players:
		panic(fmt.Sprintf("effects.Register: %q %s admits players — a sacrifice clause matches permanents only", card, where))
	}
}

// checkReturnClause is #1213's registration guard for the
// return-to-hand cost component, and it is checkSacrificeClause's
// shape one verb over: a fixed count of at least one, over a clause
// that matches permanents.
//
// Refused at boot, each naming a card-file mistake:
//
//   - a count below one, or no clause at all: the component would
//     demand nothing and the ability would be free.
//   - a variable count: "Return X permanents you control" is not
//     printed, and it would need the announce path
//     SacrificeCostBounds uses.
//   - AllowSame: one permanent cannot pay two returns.
//   - Players: a player is not a permanent.
func checkReturnClause(card, where string, rc *game.ReturnToHandCost) {
	if rc == nil {
		return
	}
	if rc.Count < 1 || rc.Filter == nil {
		panic(fmt.Sprintf("effects.Register: %q %s returns %d permanents to hand — a return cost returns at least one and needs its clause (ReturnAPermanentToHand)",
			card, where, rc.Count))
	}
	switch {
	case rc.Filter.CountFromX:
		panic(fmt.Sprintf("effects.Register: %q %s returns X permanents to hand — a variable return count has no shape (#1213)", card, where))
	case rc.Filter.AllowSame:
		panic(fmt.Sprintf("effects.Register: %q %s lets one permanent pay a return twice (AllowSame)", card, where))
	case rc.Filter.Players:
		panic(fmt.Sprintf("effects.Register: %q %s admits players — a return clause matches permanents only", card, where))
	}
}

// zoneDeclared reports whether `zone` appears in a Spec's
// CastableZones. S29's Register guard, kept out of the loop body so
// the panic message above reads as one thought.
func zoneDeclared(zones []game.ZoneKind, zone game.ZoneKind) bool {
	for _, z := range zones {
		if z == zone {
			return true
		}
	}
	return false
}

// Lookup returns the Spec for a given oracle ID. The second return
// is false when the ID is not in the catalog — that's the signal
// for the resolution path to fall back to manual sandbox behaviour.
// Callers MUST check the second return; the zero Spec{} is
// semantically distinct from a real registered spec (no OnResolve,
// no AsEnters), which means a missed check would silently apply
// nothing rather than triggering the manual fallback correctly.
func Lookup(oracleID string) (Spec, bool) {
	s, ok := registry[oracleID]
	return s, ok
}

// All returns a snapshot slice of every registered Spec. Order is
// undefined (map iteration). Used by the "auto"-bit serialiser on
// CardView (sub-PR 3) and by tests that want to iterate the whole
// catalog. The returned slice is freshly allocated — callers can
// mutate it freely without leaking into the registry.
func All() []Spec {
	out := make([]Spec, 0, len(registry))
	for _, s := range registry {
		out = append(out, s)
	}
	return out
}

// Has reports whether the catalog knows the given oracle ID.
// Convenience wrapper for the auto-badge bit (sub-PR 3).
func Has(oracleID string) bool {
	_, ok := registry[oracleID]
	return ok
}

// checkFlatClauses refuses a nested clause list at boot (#764, ADR
// 0065 §1). A TargetSpec IS its first clause and hangs the rest off
// it in Rest; an entry of Rest with its own Rest is a statement
// nothing reads, because every walk over a clause list is flat. The
// constructor (Clauses) flattens, so reaching this is a hand-built
// literal, and finding out at boot is far cheaper than finding out
// when a second target silently never gets asked for.
func checkFlatClauses(name string, spec *game.TargetSpec) {
	if spec == nil {
		return
	}
	for i := range spec.Rest {
		if len(spec.Rest[i].Rest) > 0 {
			panic(fmt.Sprintf("effects.Register: %q target clause %d nests further clauses — the list is flat; build it with Clauses(...)", name, i+1))
		}
	}
}

// checkExhaustAbilities holds the exhaust declaration to the two
// things the engine's record needs from it (#1181), over BOTH ability
// lists (#1183).
//
// The record is keyed by (object, ability LABEL) — see
// game/activation_tally.go on why the label and not the index — so a
// blank label is unaddressable and two exhaust abilities sharing one
// label would share one use between them. Loot, the Pathfinder prints
// three exhaust abilities on one card and each of them is separately
// activatable; a copy-paste that gave two of them the same label would
// silently take one away, which is exactly the class of bug a boot
// panic is cheap insurance against.
//
// ONE `seen` set across the two lists, and that is the #1183 half: a
// mana ability and an activated ability on the same card write to the
// same key space on the same object, so Loot's "Exhaust — {G}, {T}:
// Add three mana of any one color" and a hypothetical activated
// ability with the same label would share one use across the two
// kinds. Two per-list sets would not have caught it.
//
// The third check is the other direction: an ability whose label says
// "exhaust" and does not set the bit gets no gate at all and is
// repeatable forever. Nothing else in an ability label mentions the
// word — the cards that talk ABOUT exhaust abilities (Rangers'
// Refueler's trigger, Boom Scholar's cost modifier, Elvish Refueler's
// static) say so somewhere other than an ability Label.
func checkExhaustAbilities(spec Spec) {
	seen := make(map[string]bool, len(spec.Activated)+len(spec.ManaAbilities))
	for i, a := range spec.Activated {
		checkOneExhaustAbility(spec.Name, "activated ability", i, a.Label, a.Exhaust, seen)
	}
	// #1183: the mana half. A mana ability takes the other entry
	// point (ActivateManaAbility, CR 605.3a) and the same record, so
	// it answers to the same three rules.
	for i, a := range spec.ManaAbilities {
		checkOneExhaustAbility(spec.Name, "mana ability", i, a.Label, a.Exhaust, seen)
	}
	// #1184: the permission that suspends the gate. A nil Applies
	// would grant it to every player at every moment, which no card
	// prints and which nothing downstream could tell apart from a
	// card whose condition simply happened to hold — so it is a boot
	// panic rather than a silent grant. A blank Label is the same
	// argument as the one above: it is what a log line names.
	for i, p := range spec.ExhaustPermissions {
		if p.Applies == nil {
			panic(fmt.Sprintf("effects.Register: %q exhaust permission %d has no Applies — it would suspend the gate for every player, always", spec.Name, i))
		}
		if p.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q exhaust permission %d has no Label — the label is the printed clause", spec.Name, i))
		}
	}
}

// checkOneExhaustAbility is the body of the three checks above, for
// one ability of either kind. `seen` is shared across the kinds
// because the record's key space is.
func checkOneExhaustAbility(name, kind string, i int, label string, exhaust bool, seen map[string]bool) {
	mentions := strings.Contains(strings.ToLower(label), "exhaust")
	if exhaust && !mentions {
		panic(fmt.Sprintf("effects.Register: %q %s %d sets Exhaust but its label does not print the keyword — the label is what the player reads", name, kind, i))
	}
	if mentions && !exhaust {
		panic(fmt.Sprintf("effects.Register: %q %s %d prints \"Exhaust\" and does not set Exhaust: true — without the bit it can be activated every turn", name, kind, i))
	}
	if !exhaust {
		return
	}
	if label == "" {
		panic(fmt.Sprintf("effects.Register: %q %s %d is an exhaust ability with no Label — the label is the record's key", name, kind, i))
	}
	if seen[label] {
		panic(fmt.Sprintf("effects.Register: %q declares two exhaust abilities labelled %q — they would share one use", name, label))
	}
	seen[label] = true
}

// checkPlayerKeywords guards Spec.PlayerKeywords (#1197): every entry
// must be a token the engine actually honours on a PLAYER.
//
// A boot panic rather than a silent no-op, for the reason
// protection.go gives about its closed grammar: a token nothing can
// parse grants NOTHING, so a card with a typo in it ships looking
// finished and doing nothing at all, and the badge would promise a
// rule the engine does not enforce. The grammar is closed, so the
// list of what a player may be granted is short and checkable.
//
// Two shapes are accepted, and they are exactly the two the three
// consumers read:
//
//   - "hexproof" (CR 702.11d), the bare token;
//   - any "protection from <quality>" the closed grammar parses
//     (CR 702.16i) — "protection from everything" is the only one a
//     catalogued card prints, but the grammar is the grammar.
//
// Shroud is refused: no card prints shroud on a player, and a token
// with no card behind it is what game.CanonicalKeywords refuses a
// bare "protection" for.
func checkPlayerKeywords(spec Spec) {
	for i, kw := range spec.PlayerKeywords {
		if kw == game.KeywordHexproof {
			continue
		}
		if _, ok := game.ParseProtectionQuality(kw); ok {
			continue
		}
		panic(fmt.Sprintf("effects.Register: %q PlayerKeywords[%d] = %q is not a player ability the engine honours — "+
			"use %q or a \"protection from <quality>\" token the closed grammar parses (ADR 0072, #1197)",
			spec.Name, i, kw, game.KeywordHexproof))
	}
}
