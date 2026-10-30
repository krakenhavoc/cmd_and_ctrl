package effects

import (
	"fmt"
	"slices"
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
	// ADR 0126 §6: what a declared Purpose may say, slot by slot.
	checkSpecPurposes(spec)
	// ADR 0208 §5: a sticker sheet's ability stickers.
	registerStickerSheet(spec)
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
		checkRaisedModeMax(spec.Name, "Modes", spec.Modes)
		// ADR 0097: no printed spell says "that hasn't been chosen",
		// and a spell has no permanent object to remember with — the
		// restriction would read the zero identity and never apply.
		if spec.Modes.NotChosen != game.ModeMemoryNone {
			panic(fmt.Sprintf("effects.Register: %q declares \"that hasn't been chosen\" on a spell's Modes — only a triggered or activated ability has an object to remember with", spec.Name))
		}
		if spec.Modes.AtRandom {
			panic(fmt.Sprintf("effects.Register: %q declares \"choose one at random\" on a spell's Modes — only a triggered ability's mode can be picked by the engine (#3399)", spec.Name))
		}
		for i, o := range spec.Modes.Options {
			if o.Label == "" {
				panic(fmt.Sprintf("effects.Register: %q mode %d has no label — the bullet is the whole of what the picker shows", spec.Name, i))
			}
			checkFlatClauses(spec.Name, o.Targets)
			refuseOpponentChoice(spec.Name, "a mode's", o.Targets)
			checkSpellXBound(spec.Name, o.Targets)
			checkModeCost(spec.Name, "mode", i, o.Cost, true)
			checkTargetAtRandom(spec.Name, "a spell's mode", o.Targets, false)
		}
	}
	checkEscalate(spec)
	checkEntwine(spec)
	checkSplice(spec)
	checkFlatClauses(spec.Name, spec.Targets)
	refuseTriggerActorChoice(spec.Name, "a spell's", spec.Targets)
	checkSpellXBound(spec.Name, spec.Targets)
	checkTargetAtRandom(spec.Name, "a spell's", spec.Targets, false)
	checkExhaustAbilities(spec)
	checkBoast(spec)
	checkPowerUp(spec)
	checkPowerUpExtras(spec)
	checkMayAtResolution(spec.Name, spec.Activated)
	checkPlayerKeywords(spec)
	checkHandSize(spec)
	for _, a := range spec.Activated {
		checkFlatClauses(spec.Name, a.Targets)
		refuseTriggerActorChoice(spec.Name, "an activated ability's", a.Targets)
		checkTargetAtRandom(spec.Name, "an activated ability's", a.Targets, true)
		checkActivatedAtRandom(spec.Name, a.Targets)
		// #1723: an X-bound target clause is allowed on an activated
		// ability whose OWN cost announces an X (CR 602.2b) — Lazav,
		// the Multifarious's "{X}: ... with mana value X". Still
		// refused when the cost has none: the flag would be a bound
		// nothing on the wire ever sets.
		//
		// ADR 0109 §9: a clause bounded by the counters removed needs a
		// cost that removes a VARIABLE number of them — a fixed count
		// is the printed number, and no printed card bounds by one.
		counted := a.Cost.RemoveCounters != nil && a.Cost.RemoveCounters.Variable
		checkNoXBound(spec.Name, "an activated ability's", a.Targets, a.Cost.DemandsX(), counted)
		if a.Modes != nil {
			for _, o := range a.Modes.Options {
				checkNoXBound(spec.Name, "an activated mode's", o.Targets, a.Cost.DemandsX(), counted)
			}
		}
		if a.Modes != nil {
			if a.Targets != nil {
				panic(fmt.Sprintf("effects.Register: %q declares an activated ability with both Targets and Modes — put the target clause on the mode", spec.Name))
			}
			checkRaisedModeMax(spec.Name, "an activated ability's Modes", a.Modes)
			checkNotChosen(spec.Name, "an activated ability's Modes", a.Modes, a.Label)
			if a.Modes.AtRandom {
				panic(fmt.Sprintf("effects.Register: %q declares \"choose one at random\" on an activated ability — only a triggered ability's mode can be picked by the engine (#3399)", spec.Name))
			}
			for i, o := range a.Modes.Options {
				if o.Label == "" {
					panic(fmt.Sprintf("effects.Register: %q activated mode %d has no label", spec.Name, i))
				}
				checkFlatClauses(spec.Name, o.Targets)
				refuseOpponentChoice(spec.Name, "an activated mode's", o.Targets)
				checkModeCost(spec.Name, "activated mode", i, o.Cost, false)
				checkTargetAtRandom(spec.Name, "an activated mode's", o.Targets, false)
			}
		}
	}
	for _, t := range spec.Triggered {
		checkStateTrigger(spec.Name, t)
		checkNotPutByThisAbility(spec.Name, t)
		checkFlatClauses(spec.Name, t.Targets)
		// A trigger announces no X, ever (#1559) — unlike an activated
		// ability (#1723), there is no cost to check.
		checkNoXBound(spec.Name, "a trigger's", t.Targets, false, false)
		checkNoDivideX(spec.Name, "a trigger's", t.Targets)
		checkTargetAtRandom(spec.Name, "a trigger's", t.Targets, true)
		if t.Modes != nil {
			for _, o := range t.Modes.Options {
				refuseOpponentChoice(spec.Name, "a trigger mode's", o.Targets)
				checkNoXBound(spec.Name, "a trigger mode's", o.Targets, false, false)
				checkNoDivideX(spec.Name, "a trigger mode's", o.Targets)
				checkTargetAtRandom(spec.Name, "a trigger mode's", o.Targets, true)
			}
		}
		if t.Modes != nil && t.Targets != nil {
			panic(fmt.Sprintf("effects.Register: %q declares a trigger with both Targets and Modes — put the target clause on the mode", spec.Name))
		}
		if t.Modes != nil {
			checkRaisedModeMax(spec.Name, "a trigger's Modes", t.Modes)
			checkNotChosen(spec.Name, "a trigger's Modes", t.Modes, t.Key)
			checkAtRandom(spec.Name, t.Modes)
			for i, o := range t.Modes.Options {
				checkModeCost(spec.Name, "trigger mode", i, o.Cost, false)
			}
		}
	}
	checkCastTimeValues(spec)
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
		// #3225: a per-offer X ceiling bounds an X the offer's own
		// price announces, so it needs an {X} to bound; on a cost
		// without one the engine locks X at 0 (CR 107.3b) and the
		// number would never be read. A negative is a typo.
		if ac.XCeiling < 0 {
			panic(fmt.Sprintf("effects.Register: %q offers %q with a negative XCeiling", spec.Name, ac.Key))
		}
		if ac.XCeiling > 0 {
			if pc, err := game.ParseCost(ac.ManaCost); err != nil || pc.XSlots == 0 {
				panic(fmt.Sprintf("effects.Register: %q offers %q with an XCeiling but a mana cost %q that has no {X} to bound", spec.Name, ac.Key, ac.ManaCost))
			}
		}
		// S29: an offer bound to a zone the card cannot be cast from
		// is unclaimable — the cast path rejects the zone before it
		// ever looks at the price. A card file that wrote one meant
		// to list the zone as well, and finding out at boot is far
		// cheaper than finding out when a flashback button never
		// appears.
		if ac.FromZone != "" && ac.FromZone != game.ZoneHand && !zoneDeclared(spec.CastableZones, ac.FromZone) && !conditionalZoneDeclared(spec.CastableZonesIf, ac.FromZone) {
			panic(fmt.Sprintf("effects.Register: %q offers %q from %s but does not list that zone in CastableZones or CastableZonesIf",
				spec.Name, ac.Key, ac.FromZone))
		}
		// #1727: an offer's card component rides ONE wire list
		// (alt_cost_ids) and is read by one accessor that picks the
		// first component it finds, so an offer that declared two would
		// silently charge only one of them — a cheaper card than the one
		// printed. No printed alternative cost names two card-shaped
		// payments; Demon of Death's Gate's "pay 6 life and sacrifice
		// three black creatures" is life plus ONE.
		if n := altCostCardComponents(ac); n > 1 {
			panic(fmt.Sprintf("effects.Register: %q offers %q with %d card-shaped payments — an alternative cost carries at most one of ExileFromHand, ReturnToHand, ExileFromGraveyard, Sacrifice, DiscardFromHand and TapOthers",
				spec.Name, ac.Key, n))
		}
		// The sacrifice component is the additional cost's clause and
		// is held to the same shape, minus the two variable forms: no
		// printed alternative cost sacrifices "X" or "any number", and
		// a fixed count is what the announce path, the picker and the
		// enumerator all read off it.
		checkSacrificeClause(spec.Name, fmt.Sprintf("alternative cost %q", ac.Key), ac.Sacrifice, false, false, false)
		refuseSacrificeDifferent(spec.Name, fmt.Sprintf("alternative cost %q", ac.Key), ac.Sacrifice)
		checkAltCostSetRule(spec.Name, ac)
		checkAltCostXBound(spec, ac)
		checkAltCostTapOthers(spec.Name, ac)
		checkEmerge(spec.Name, ac)
		checkHarmonize(spec.Name, ac)
		checkAwaken(spec, ac)
		checkCastsFace(spec, ac)
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
		if z == game.ZoneOutside {
			panic(fmt.Sprintf("effects.Register: %q declares ZoneOutside in CastableZones — a cast from outside the game is a per-instance game.CastPermission (#3955, ADR 0198), never a card-level declaration",
				spec.Name))
		}
		if z == game.ZoneExile {
			panic(fmt.Sprintf("effects.Register: %q declares ZoneExile in CastableZones — an exile cast is a per-instance game.CastPermission, never a card-level declaration (#659, ADR 0066)",
				spec.Name))
		}
	}
	// ADR 0066 amendment (2026-10-13): a conditional zone is a card-level
	// declaration too, so the same refusals apply, plus a nil condition,
	// which would be an unconditional zone written the long way round.
	for _, cz := range spec.CastableZonesIf {
		switch {
		case cz.Zone != game.ZoneGraveyard:
			panic(fmt.Sprintf("effects.Register: %q declares %s in CastableZonesIf — only a graveyard is a card-level conditional cast zone (hand and command are the format's, exile and the library are per-instance permissions, ADR 0066)",
				spec.Name, cz.Zone))
		case cz.Cond == nil:
			panic(fmt.Sprintf("effects.Register: %q declares %s in CastableZonesIf with no condition — list it in CastableZones", spec.Name, cz.Zone))
		}
	}
	// ADR 0073: the optional additional costs. Six boot checks, each
	// for a shape that compiles and then behaves as something the card
	// does not print.
	checkTeamworkBlight(spec)
	// #3868: a tap-permanents cost, on the mandatory slot or a branch.
	checkTapOthersCost(spec)
	// ADR 0100 §2: the either/or cost's shapes.
	checkEitherCost(spec)
	// ADR 0129 §5: energy in an additional cost, and replicate's pair.
	checkReplicate(spec)
	// #3999: an optional "you may waterbend {N}".
	checkOptionalWaterbend(spec)
	if spec.AdditionalCost != nil && spec.AdditionalCost.Optional {
		panic(fmt.Sprintf("effects.Register: %q puts an Optional cost in AdditionalCost — the mandatory slot is never optional; declare it in OptionalCosts", spec.Name))
	}
	// ADR 0089: gift is declared once, in Spec.Gift, and buildDef grows
	// its cost. A hand-rolled one in OptionalCosts would have no gift
	// effect behind it and no entry trigger — a promise that gives
	// nothing — and a Gift built outside the constructors would have
	// no effect at all.
	if g := spec.Gift; g != nil {
		if g.give == nil || g.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q declares a Gift built by hand — use GiftACard / GiftAFood / GiftATappedFish / GiftATreasure", spec.Name))
		}
		if g.Targets != nil {
			if spec.Modes != nil {
				panic(fmt.Sprintf("effects.Register: %q is modal and its gift rewrites the target clause — put the clause on the mode", spec.Name))
			}
			checkFlatClauses(spec.Name, g.Targets)
			refuseOpponentChoice(spec.Name, "a gift's", g.Targets)
		}
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
		// An optional cost that rewrites the target clause is
		// read against the card-level clause (ADR 0089 §3); a modal
		// card's clauses live on its modes, where no rewrite reaches.
		if oc.Targets != nil {
			if spec.Modes != nil {
				panic(fmt.Sprintf("effects.Register: %q is modal and optional cost %q rewrites the target clause — put the clause on the mode", spec.Name, oc.Key))
			}
			checkFlatClauses(spec.Name, oc.Targets)
			refuseOpponentChoice(spec.Name, "an optional cost's", oc.Targets)
		}
		if oc.ChoosesOpponent || oc.Key == game.GiftKey {
			panic(fmt.Sprintf("effects.Register: %q declares a gift cost in OptionalCosts — declare Spec.Gift and let buildDef grow the cost (ADR 0089)", spec.Name))
		}
		// #2153: kicker is the one key a card may declare twice — "Kicker
		// {R} and/or {W}" is two kicker abilities (CR 702.33b) — and
		// checkKickers below holds that to its printed shape.
		if seenOptional[oc.Key] && oc.Key != game.KickerKey {
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
		// #1213, ADR 0100 §3: no variable shape here. A cast's plan may
		// hold one variable clause, and only in the mandatory slot, where
		// every printed one sits (checkVariableSacrificePlan); no card
		// prints a variable kicker or buyback.
		checkSacrificeClause(spec.Name, fmt.Sprintf("optional cost %q", oc.Key), oc.Sacrifice, false, false, false)
		refuseSacrificeDifferent(spec.Name, fmt.Sprintf("optional cost %q", oc.Key), oc.Sacrifice)
		if oc.Empty() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q demands nothing", spec.Name, oc.Key))
		}
		// ADR 0073 §4: only a mana-only cost may be paid more than
		// once. Every printed multikicker is mana, and N card-shaped
		// payments per cast is a wire shape nothing asks for.
		// #4092: a replicate cost that taps one permanent per payment is
		// the exception — its N payments are N names on teamwork_ids, one
		// list the cast path sums (game.planTapOthers), so nothing else
		// card-shaped may ride beside it.
		tapless := oc
		if oc.Key == game.ReplicateKey {
			tapless.TapOthers = nil
		}
		if oc.MaxPayments() > 1 && tapless.CardsDemanded() {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q repeats and demands cards or permanents — only a mana-only cost may repeat", spec.Name, oc.Key))
		}
		// #3232: an {X} in an optional cost is the caster's one announced
		// X (CR 107.3a), so it is paid at most once — "X, paid three
		// times" is not a card — and a floor on X is meaningful only on
		// a cost that charges one. A negative floor is a typo.
		if oc.MinX < 0 {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q has a negative MinX", spec.Name, oc.Key))
		}
		if oc.MinX > 0 && !game.OptionalCostHasX(oc) {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q declares MinX but its mana cost %q has no {X}", spec.Name, oc.Key, oc.ManaCost))
		}
		if game.OptionalCostHasX(oc) && oc.MaxPayments() > 1 {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q has an {X} and repeats — X is one announced number, so it may be paid at most once", spec.Name, oc.Key))
		}
		// PayLifeX rides the shared XValue slot (ADR 0021 §3), and two
		// claimants on one number is a bug waiting to be written.
		if oc.PayLifeX {
			panic(fmt.Sprintf("effects.Register: %q optional cost %q pays X life — an optional cost cannot claim the shared X slot", spec.Name, oc.Key))
		}
	}
	checkKickers(spec)
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
	checkSpellSpendOnly(spec)
	// ADR 0073 §7: a cast condition with no printed clause produces a
	// refusal the client cannot explain, and a clause with no
	// condition refuses nothing while claiming to. Same for a
	// restriction: the Label IS the message the player is shown.
	if (spec.CastCondition == nil) != (spec.CastConditionLabel == "") {
		panic(fmt.Sprintf("effects.Register: %q declares a CastCondition without its printed CastConditionLabel, or the label without the condition", spec.Name))
	}
	checkCastRestrictions(spec.Name, spec.CastRestrictions)
	// ADR 0109 §4: the same two checks for the land-play twin.
	checkLandPlayRestrictions(spec.Name, spec.LandPlayRestrictions)
	// CR 116.2d (#1986): an ignorable effect and the restrictions that
	// are it come together or not at all.
	checkIgnorableEffect(spec)
	// #1985: and for the entry twin.
	checkEntryRestrictions(spec.Name, spec.EntryRestrictions)
	// ADR 0109 §6: and for a targeting restriction, which must also say
	// which zone it is about.
	checkTargetingRestrictions(spec.Name, spec.TargetingRestrictions)
	// #3860: a target tax names what it charges, whom and for what.
	checkTargetTaxes(spec.Name, spec.TargetTaxes)
	// ADR 0108 §10: a "dealt as though its source had" static names what
	// the source is treated as having, and wither is about creatures only
	// (CR 702.80a), so it can't be narrowed to damage dealt to a player.
	for i, s := range spec.DamageAsThough {
		if !s.Wither && !s.Infect {
			panic(fmt.Sprintf("effects.Register: %q damage-as-though static %d gives the source neither wither nor infect", spec.Name, i))
		}
		if s.ToYou && !s.Infect {
			panic(fmt.Sprintf("effects.Register: %q damage-as-though static %d is wither on damage dealt to a player, which changes nothing (CR 702.80a)", spec.Name, i))
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
	// statement about nothing. The emblem slot (#1275) runs the same
	// guard from checkEmblemSpec.
	checkActivationTimings(spec.Name, spec.ActivationTimings)
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
	checkGrantedAlternativeCosts(spec.Name, spec.GrantedAlternativeCosts)
	checkOpeningHand(spec.Name, spec.OpeningHand)
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
		if m.Kind == game.CostZeroAlternative {
			panic(fmt.Sprintf("effects.Register: %q self cost modifier %d is a CostZeroAlternative — an alternative on an activation belongs in Spec.CostModifiers", spec.Name, i))
		}
	}
	// #4527: a "{0} rather than" alternative prices activations only;
	// the engine never reads it on a cast or a special action.
	for i, m := range spec.CostModifiers {
		if m.Kind == game.CostZeroAlternative && (!m.Activations || m.Amount != nil) {
			panic(fmt.Sprintf("effects.Register: %q cost modifier %d (%q) is a CostZeroAlternative without Activations or with an Amount — build it with PayZeroRatherThanThePowerUpCost", spec.Name, i, m.Label))
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
	// #1547: a spend rider that could never fire is a card that says
	// something the engine silently does not do.
	for i, ma := range spec.ManaAbilities {
		// #2392: PainToYou BUILDS the rider, so a spec that also
		// declares one would deal its damage one way and tell the
		// auto-tapper another.
		if ma.PainToYou < 0 || (ma.PainToYou > 0 && ma.Rider != nil) {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d: PainToYou must be positive and replaces Rider — set one", spec.Name, i))
		}
		// #4185, #4189: GainLife, OpponentsGainLife and PingOpponents build
		// the rider the same way, so a spec carries at most one of the five.
		if ma.GainLife < 0 || ma.OpponentsGainLife < 0 || ma.PingOpponents < 0 {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d: GainLife, OpponentsGainLife and PingOpponents must be positive", spec.Name, i))
		}
		riders := 0
		for _, set := range []bool{ma.Rider != nil, ma.PainToYou > 0, ma.GainLife > 0, ma.OpponentsGainLife > 0, ma.PingOpponents > 0} {
			if set {
				riders++
			}
		}
		if riders > 1 {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d: Rider, PainToYou, GainLife, OpponentsGainLife and PingOpponents each build the rider — set one", spec.Name, i))
		}
		// #3042: two computed outputs would leave the precedence
		// (ProducedAmount wins) as the only thing deciding which one runs.
		if ma.ProducedAmount != nil && ma.ProducedFunc != nil {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d declares both ProducedAmount and ProducedFunc — set one", spec.Name, i))
		}
		// ADR 0130 §4: as for an activated ability, an exert beside a
		// cost that moves the source is not modelled.
		if ma.Cost.Exert && (ma.Cost.Sacrifice || ma.Cost.ExileSelf) {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d exerts its source AND moves it — not modelled (ADR 0130 §4)", spec.Name, i))
		}
		// ADR 0129 §5: zero is "no energy component".
		if ma.Cost.Energy < 0 {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d declares a negative energy cost %d", spec.Name, i, ma.Cost.Energy))
		}
		for _, r := range ma.SpendRiders {
			if err := validateManaSpendRider(r); err != nil {
				panic(fmt.Sprintf("effects.Register: %q mana ability %d: %v", spec.Name, i, err))
			}
		}
	}
	// ADR 0103 (superseding ADR 0071 decision 3's reservation): a Room
	// gates EVERY ability on a door, and nothing but a Room has doors.
	// An ungated ability on a Room would be active while its door is
	// locked — stronger than printed — and a door gate anywhere else
	// would switch an ability on and off for a reason the card never
	// prints.
	if problem := checkRoomSpec(spec); problem != "" {
		panic(fmt.Sprintf("effects.Register: %q %s (ADR 0103)", spec.Name, problem))
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
	// #1391: a grant the engine cannot carry out would put a row on a
	// card and then refuse it, so it fails at boot. Only plot from the
	// top of the library is built (game.SpecialActionGrantBuilt).
	for i, gr := range spec.SpecialActionGrants {
		if !game.SpecialActionGrantBuilt(gr.Kind, gr.Zone) {
			panic(fmt.Sprintf("effects.Register: %q special action grant %d gives %q in zone %q, which the engine cannot carry out (ADR 0062 amendment 2026-09-24, #1391)",
				spec.Name, i, gr.Kind, gr.Zone))
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
	// #2581: a printed X ceiling with no count would read as "no
	// ceiling", a card stronger than printed.
	if xc := spec.XCeiling; xc != nil && (xc.Count == nil || xc.Label == "") {
		panic(fmt.Sprintf("effects.Register: %q declares an X ceiling with no count or no label", spec.Name))
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
		checkSacrificeClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.SacrificeOther, true, true, false)
		checkReturnClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.ReturnToHand)
		checkExilePermanentsClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.ExilePermanents)
		// ADR 0137: craft's graveyard half and an exile-cards component
		// could both name one graveyard card, and no printed cost has
		// both.
		if ec := ab.Cost.ExilePermanents; ec != nil && ec.FromGraveyard && ab.Cost.ExileCards != nil {
			panic(fmt.Sprintf("effects.Register: %q ability %d exiles materials from the graveyard beside an exile-cards cost — one card could pay both", spec.Name, i))
		}
		checkTapOthersClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.TapOthers, true)
		// #660: a discard clause that discards nothing would make
		// the ability free, the way a zero-counter cost would — unless
		// it is "Discard your hand" (#1600), whose count is the hand.
		checkDiscardClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.DiscardCards)
		checkDiscardHandBesideHandCosts(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost)
		checkDiscardXBesideOtherCosts(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost)
		// #2598: the reveal-cards component (reveal_cards_cost.go).
		checkRevealCardsClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost)
		// ADR 0109 §7: the random discard and the two library
		// components (checkLibraryCosts).
		checkLibraryCosts(spec.Name, i, ab.Cost)
		checkEnergyCost(spec.Name, i, ab.Cost)
		checkLoyaltyX(spec.Name, i, ab.Cost)
		// #1297: the exile-N-cards component, held to the rules its
		// mana owner is held to (checkExileCardsClause).
		checkExileCardsClause(spec.Name, fmt.Sprintf("ability %d", i), ab.Cost.ExileCards)
		checkAbilityCostModifiers(spec.Name, i, ab)
		// ADR 0106 §1 decision 1: "Any player may activate this ability".
		checkAnyPlayerAbility(spec.Name, fmt.Sprintf("ability %d", i), ab)
		checkGrantorOnlyIsGranted(spec.Name, i, ab)
		// #4000: an exile-the-grantor cost is a granted row's alone.
		checkExileGrantor(spec.Name, fmt.Sprintf("ability %d", i), ab, false, "")
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
		// #2028: the source pays one component (CR 118.3), so a cost
		// cannot both return it and sacrifice or exile it. No printed
		// card does; the activation would refuse every attempt.
		if ab.Cost.ReturnSelf && (ab.Cost.SacrificeSelf || ab.Cost.ExileSelf) {
			panic(fmt.Sprintf("effects.Register: %q ability %d returns its source to hand AND sacrifices or exiles it — one permanent pays one cost component",
				spec.Name, i))
		}
		// #2726: the same one-component rule for the bottom-of-library cost.
		if ab.Cost.BottomSelf && (ab.Cost.SacrificeSelf || ab.Cost.ExileSelf || ab.Cost.ReturnSelf) {
			panic(fmt.Sprintf("effects.Register: %q ability %d puts its source on the bottom of the library AND sacrifices, exiles or returns it — one permanent pays one cost component",
				spec.Name, i))
		}
		// ADR 0130 §4: no printed card exerts a source its own cost also
		// moves, and the exert would expire with the object (CR 400.7).
		if ab.Cost.Exert && (ab.Cost.SacrificeSelf || ab.Cost.ExileSelf || ab.Cost.ReturnSelf || ab.Cost.BottomSelf) {
			panic(fmt.Sprintf("effects.Register: %q ability %d exerts its source AND moves it — not modelled (ADR 0130 §4)",
				spec.Name, i))
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
		if ab.Cost.XSlots() > 0 && game.TapOthersCountFromX(ab.Cost.TapOthers) {
			panic(fmt.Sprintf("effects.Register: %q ability %d has {X} in its mana cost %q AND taps X permanents — one announced X cannot pay both",
				spec.Name, i, ab.Cost.Mana))
		}
		if game.SacrificeCountFromX(ab.Cost.SacrificeOther) && game.TapOthersCountFromX(ab.Cost.TapOthers) {
			panic(fmt.Sprintf("effects.Register: %q ability %d both sacrifices X permanents and taps X permanents — one announced X cannot pay both",
				spec.Name, i))
		}
		// #1221 / #1404: the same rule one zone over. ExileSelf
		// follows the ability's zone — scavenge's and embalm's
		// "Exile this card from YOUR GRAVEYARD" (CR 702.96a,
		// CR 702.128a) from the graveyard, Perpetual Timepiece's
		// "Exile this artifact" from the battlefield (the default
		// when Zones is nil). Every zone the ability functions from
		// has to be one the component can be paid from
		// (game.ExileSelfZoneSupported); an ability that declares
		// the hand or exile could never pay it, and would look
		// complete on the catalog page while refusing every
		// activation.
		if ab.Cost.ExileSelf {
			for _, z := range game.AbilityZones(game.ActivatedAbilityShape{Zones: ab.Zones}) {
				if !game.ExileSelfZoneSupported(z) {
					panic(fmt.Sprintf("effects.Register: %q ability %d declares an exile-this cost but functions from the %s — the component is paid from the battlefield or the graveyard",
						spec.Name, i, z))
				}
			}
		}
		// #1310, CR 701.67b: a waterbend clause names part of the
		// mana component — the part its taps may cover — so the mana
		// component must contain it. A clause with no key or no pool
		// could never be paid, one whose cost does not parse would
		// cover nothing, and one larger than the mana would let a
		// tap pay for mana the ability never charged. Build it with
		// WaterbendCost, which cannot get any of these wrong.
		if ab.Cost.Waterbend != nil {
			checkAbilityWaterbend(spec.Name, i, ab.Cost)
		}
		if ab.Cost.MinX > 0 && !ab.Cost.DemandsX() {
			panic(fmt.Sprintf("effects.Register: %q ability %d sets MinX %d but its cost %q has no {X} — a floor on a variable that cannot vary makes the ability unactivatable",
				spec.Name, i, ab.Cost.MinX, ab.Cost.Mana))
		}
		// #1600: a "spend only <colour> mana" clause (spend_only.go).
		checkSpendOnlyClause(spec.Name, i, ab.Cost)
	}
	for i, ma := range spec.ManaAbilities {
		// #1213: `false` — a mana ability has no stack item and no
		// announced X (CR 605.3b), so a "Sacrifice X …" clause there
		// has nothing to read its count from.
		checkSacrificeClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.SacrificeOther, true, false, false)
		refuseSacrificeDifferent(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.SacrificeOther)
		checkTapOthersClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.TapOthers, false)
		// #1228 / CR 113.6: the MANA half of the zone dimension, held
		// to the same three rules the activated half is held to —
		// every one of them a boot-time refusal rather than a
		// mysteriously-dead card.
		checkManaAbilityZones(spec.Name, i, ma)
		// #1864: an any-player mana ability's limits.
		checkAnyPlayerManaAbility(spec.Name, i, ma)
		// #789: the counter components are one declaration with two
		// owners, so they are checked by one function in both places.
		checkCounterCost(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.RemoveCounters, ma.Cost.AddCounter)
		// #1213 / #1283: a card-picking clause that picks nothing would
		// make the ability free — the refusal the CR 602 discard gets,
		// and the same "Discard your hand" exception (#1600).
		checkDiscardClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.DiscardCards)
		checkManaDiscardHandBesideHandCosts(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost)
		// ADR 0109 §7: a random discard is a CR 602 component only. No
		// printed mana ability discards at random, and the mana payer
		// never draws one, so the ability would be free.
		if dc := ma.Cost.DiscardCards; dc != nil && dc.Random {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d discards at random — only an activated ability's cost can",
				spec.Name, i))
		}
		// #2527: and "Discard X cards" needs an announced X, which a
		// mana ability (CR 605.3b) has no stack item to carry.
		if game.DiscardCountFromX(ma.Cost.DiscardCards) {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d discards X cards — a mana ability announces no X (CR 605.3b)",
				spec.Name, i))
		}
		// #2190: and neither can "Discard a card with mana value X".
		if game.DiscardManaValueX(ma.Cost.DiscardCards) {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d discards a card with mana value X — a mana ability announces no X (CR 605.3b)",
				spec.Name, i))
		}
		checkExileCardsClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.ExileCards)
		checkExilePermanentsClause(spec.Name, fmt.Sprintf("mana ability %d", i), ma.Cost.ExilePermanents)
		// ADR 0137: craft materials are a CR 602 ability's cost; no mana
		// ability exiles graveyard cards through this component.
		if ec := ma.Cost.ExilePermanents; ec != nil && (ec.FromGraveyard || ec.OrMore || ec.ShareCardType || len(ec.EachSubtype) > 0) {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d exiles craft materials (the graveyard, an open count or a set rule) — only a CR 602 ability may (ADR 0137)", spec.Name, i))
		}
		if ma.Cost.Mana != "" {
			if _, err := game.ParseCost(ma.Cost.Mana); err != nil {
				panic(fmt.Sprintf("effects.Register: %q mana ability %d declares an unparseable mana cost %q: %v",
					spec.Name, i, ma.Cost.Mana, err))
			}
		}
	}
	if spec.AdditionalCost != nil {
		// ADR 0100 §3: a cast's mandatory additional cost may print
		// "sacrifice X …" or "sacrifice any number of …"; the cross-slot
		// rule that makes the flat payment list unambiguous is
		// checkVariableSacrificePlan's.
		checkSacrificeClause(spec.Name, "additional cost", spec.AdditionalCost.Sacrifice, false, true, true)
		refuseSacrificeDifferent(spec.Name, "additional cost", spec.AdditionalCost.Sacrifice)
	}
	checkVariableSacrificePlan(spec)
	checkSacrificeAllCost(spec)
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
	// ADR 0108 §8 (#1906): a replacement's additional effect (Then) is a
	// prevention effect's (CR 615.5), so it requires Prevention, and its
	// unit (ThenPer) is required with it — the printed subject picks it,
	// and a forgotten one would run once per recipient on a card whose
	// text says once per source.
	for i, r := range spec.Replacements {
		if problem := replacementThenProblem(r); problem != "" {
			panic(fmt.Sprintf("effects.Register: %q replacement %d: %s", spec.Name, i, problem))
		}
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
	// ADR 0093: an ability grant is a layer-6 effect and names a bundle.
	checkStaticGrants(spec.Name, "static", spec.Static)
	// ADR 0067 amendment 2026-10-23: a static that reads an ability.
	checkAbilityReaders(spec.Name, "static", spec.Static)
	if spec.Emblem != nil {
		checkStaticGrants(spec.Name, "emblem static", spec.Emblem.Static)
		checkAbilityReaders(spec.Name, "emblem static", spec.Emblem.Static)
		checkTriggerZones(spec.Name, "emblem trigger", spec.Emblem.Triggered)
	}
	checkEmblemSpec(spec.Name, spec.Emblem)
	checkGrants(spec.Name, spec.Grants)
	registry[spec.OracleID] = spec
	def := buildDef(spec)
	fileDef(spec.OracleID, def)
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
	// ADR 0093 amendment 2026-10-13: and the zone-grant statics, so the
	// layer pass finds a Satoru by one map read per permanent.
	game.IndexZoneGrants(spec.OracleID, spec.Static)
	// #3862: and the counter kinds a static's "as long as it has N or
	// more counters" begins on, so a counter placement on any other
	// kind pays one map read.
	game.IndexCounterThresholds(spec.Static)
	// #3862: and the statics over objects off the battlefield, so the
	// zone layer step finds a Mycosynth Lattice by one map read per
	// permanent.
	game.IndexZoneLayers(spec.OracleID, spec.Static)
	// #623 / CR 114: a card that makes an emblem files a SECOND def
	// for the emblem object, under "emblem:<this key>". It goes in
	// `defs` and not in `registry`, so the engine finds the emblem's
	// abilities through the ordinary CatalogLookup and the card
	// census keeps counting cards — an emblem is not a card
	// (CR 114.4). See emblem.go and ADR 0064.
	if spec.Emblem != nil {
		fileDef(game.EmblemKey(spec.OracleID), buildEmblemDef(*spec.Emblem))
	}
	// #665 / CR 707.9a: a card whose copy effect GRANTS an ability
	// files that ability's own def, under "grant:<key>". Same map,
	// same reason as the emblem: the copy carries the key in its
	// copiable values and the engine finds the abilities through the
	// ordinary CatalogLookup, while the census keeps counting cards —
	// a granted ability is not one.
	for _, gr := range spec.Grants {
		fileDef(game.GrantKey(gr.Key), buildGrantDef(gr))
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
		TapOthers:      ma.Cost.TapOthers,
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
		// ADR 0107 §3: a static over spells is a layer-6 keyword grant
		// from the battlefield — the one thing the stack step applies.
		if s.AffectsSpells {
			switch {
			case s.Layer != game.Layer6Ability:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects spells outside layer 6 — the stack step applies keywords only", card, what, i))
			case s.AppliesTo == nil || s.Apply == nil:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects spells but has no AppliesTo or Apply", card, what, i))
			case len(s.Zones) > 0:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects spells from a declared zone — only the battlefield's statics over spells are gathered", card, what, i))
			case s.RemovesAbilities || len(s.GrantAbilities) > 0 || s.Borrows != nil:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects spells with a removal or an ability bundle — the stack step applies keywords only", card, what, i))
			}
		}
		// #3862: a static over objects off the battlefield is a layer-4
		// or layer-5 effect from the battlefield — what the zone layer
		// step applies (game/zone_layers.go).
		if len(s.AffectsZones) > 0 {
			switch {
			case s.Layer != game.Layer4Type && s.Layer != game.Layer5Color:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects objects off the battlefield in layer %d — the zone layer step applies layers 4 and 5 only", card, what, i, s.Layer))
			case s.AppliesTo == nil || s.Apply == nil:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects objects off the battlefield but has no AppliesTo or Apply", card, what, i))
			case len(s.Zones) > 0 || s.AffectsSpells || s.GrantZone != "":
				panic(fmt.Sprintf("effects.Register: %q %s %d affects objects off the battlefield beside a declared source zone, a spell clause or a zone grant — declare those as a static of their own", card, what, i))
			case s.RemovesAbilities || len(s.GrantAbilities) > 0 || s.Borrows != nil || s.ContinuesAfterRemoval:
				panic(fmt.Sprintf("effects.Register: %q %s %d affects objects off the battlefield with a removal, a grant or a borrow — the zone layer step changes types and colours only", card, what, i))
			}
			for _, zone := range s.AffectsZones {
				if why := game.AffectsZoneUnsupported(zone); why != "" {
					panic(fmt.Sprintf("effects.Register: %q %s %d affects objects in %s — %s", card, what, i, zone, why))
				}
			}
		}
		for _, zone := range s.Zones {
			if why := game.StaticZoneUnsupported(zone); why != "" {
				panic(fmt.Sprintf("effects.Register: %q %s %d functions from %s — %s",
					card, what, i, zone, why))
			}
		}
	}
}

// checkAbilityReaders is the registration check for
// StaticAbility.ReadsAbilities (ADR 0067 amendment 2026-10-23): the
// dependency order it asks for is kept inside the layer-6 bucket of the
// battlefield pass, so a reader is a layer-6 static over permanents with a
// predicate to read, and it is not also a removal or a borrow (the bucket
// orders those by rules of their own, ADR 0093 Decision 3 and ADR 0155
// Decision 4).
func checkAbilityReaders(card, what string, statics []game.StaticAbility) {
	for i, s := range statics {
		if !s.ReadsAbilities {
			continue
		}
		switch {
		case s.Layer != game.Layer6Ability:
			panic(fmt.Sprintf("effects.Register: %q %s %d reads abilities outside layer 6: only the layer-6 bucket is ordered for it", card, what, i))
		case s.AppliesTo == nil:
			panic(fmt.Sprintf("effects.Register: %q %s %d reads abilities but has no AppliesTo", card, what, i))
		case s.RemovesAbilities || s.Borrows != nil:
			panic(fmt.Sprintf("effects.Register: %q %s %d reads abilities beside a removal or a borrow: declare them as statics of their own", card, what, i))
		case s.AffectsSpells || len(s.AffectsZones) > 0:
			panic(fmt.Sprintf("effects.Register: %q %s %d reads abilities over spells or cards off the battlefield: only the battlefield's layer-6 bucket is ordered for it", card, what, i))
		}
	}
}

// checkActivationTimings is #1208's guard for a list of activation
// timing statements, shared by Spec.ActivationTimings and
// EmblemSpec.ActivationTimings (#1275) so the two homes cannot drift
// in what they refuse. `card` names the declarer in the panic.
func checkActivationTimings(card string, timings []game.ActivationTiming) {
	for i, t := range timings {
		if t.Timing == game.TimingNormal {
			panic(fmt.Sprintf("effects.Register: %q activation timing %d says nothing — set TimingFlash, TimingSorcery or TimingYourTurnOnly", card, i))
		}
		if t.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q activation timing %d has no printed Label", card, i))
		}
		if t.Covers == nil {
			panic(fmt.Sprintf("effects.Register: %q activation timing %q covers nothing", card, t.Label))
		}
	}
}

// checkCastRestrictions, checkLandPlayRestrictions and
// checkTargetingRestrictions are ADR 0073 §7's guard and its ADR 0109
// twins, shared by a Spec's slot and an EmblemSpec's (ADR 0109 §5) so the
// two homes cannot drift in what they refuse: a restriction with no
// printed Label produces a refusal the client cannot explain, and one
// with no Forbids claims to refuse and refuses nothing. `card` names the
// declarer in the panic.
func checkCastRestrictions(card string, rs []game.CastRestriction) {
	for _, r := range rs {
		checkRestriction(card, "cast", r.Label, r.Forbids == nil)
	}
}

func checkLandPlayRestrictions(card string, rs []game.LandPlayRestriction) {
	for _, r := range rs {
		checkRestriction(card, "land-play", r.Label, r.Forbids == nil)
	}
}

// checkIgnorableEffect refuses a card whose ignore switches nothing off
// (no restriction marked Ignorable) or whose Ignorable restriction has no
// ignore to be switched off by, and an ignore with no printed row or no
// rule for whom it bears on.
func checkIgnorableEffect(spec Spec) {
	marked := false
	for _, r := range spec.CastRestrictions {
		marked = marked || r.Ignorable
	}
	for _, r := range spec.LandPlayRestrictions {
		marked = marked || r.Ignorable
	}
	ig := spec.IgnorableEffect
	if ig == nil {
		if marked {
			panic(fmt.Sprintf("effects.Register: %q marks a restriction Ignorable but declares no IgnorableEffect", spec.Name))
		}
		return
	}
	if ig.Label == "" || ig.Applies == nil {
		panic(fmt.Sprintf("effects.Register: %q declares an IgnorableEffect without its Label or Applies", spec.Name))
	}
	if !marked {
		panic(fmt.Sprintf("effects.Register: %q declares an IgnorableEffect but no restriction is Ignorable, so ignoring it would switch nothing off", spec.Name))
	}
}

func checkEntryRestrictions(card string, rs []game.EntryRestriction) {
	for _, r := range rs {
		checkRestriction(card, "entry", r.Label, r.Forbids == nil)
	}
}

// checkTargetingRestrictions also refuses a restriction about no zone,
// which the targeting check would never ask.
func checkTargetingRestrictions(card string, rs []game.TargetingRestriction) {
	for _, r := range rs {
		checkRestriction(card, "targeting", r.Label, r.Forbids == nil)
		if len(r.Zones) == 0 {
			panic(fmt.Sprintf("effects.Register: %q targeting restriction %q is about no zone", card, r.Label))
		}
	}
}

func checkRestriction(card, kind, label string, forbidsNothing bool) {
	if label == "" {
		panic(fmt.Sprintf("effects.Register: %q %s restriction has no printed Label — the refusal carries it to the client", card, kind))
	}
	if forbidsNothing {
		panic(fmt.Sprintf("effects.Register: %q %s restriction %q forbids nothing", card, kind, label))
	}
}

func checkTriggerZones(card, what string, triggers []game.TriggeredAbility) {
	for i, t := range triggers {
		if zone, why := game.TriggerZonesUnsupported(t.Zones); why != "" {
			panic(fmt.Sprintf("effects.Register: %q %s %d watches from %s — %s",
				card, what, i, zone, why))
		}
	}
}

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
//     is no longer refused here. Nor is the same split across the
//     source's own kinds, spelled Among with no From (#3987,
//     Ghost-Spider's "Remove two counters from Ghost-Spider");
//   - Among without a From clause is only that self form: any kinds,
//     a fixed N of at least two. Among with Variable is a shape no
//     card prints except Jetfire's one printed kind;
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
			panic(fmt.Sprintf("effects.Register: %q %s removes %d counters of any kind with one kind choice — a removal that may mix kinds asks the kind per part: RemoveCountersOfAnyKindsFromThis for this permanent (#3987), RemoveCountersAmong across permanents (#943)", card, where, rc.N))
		}
		if rc.Among && rc.From == nil {
			// #3987: "Remove two counters from Ghost-Spider" splits across
			// the SOURCE's kinds. Only that shape: with a printed kind
			// there is one part and RemoveCountersFromThis says it, and no
			// card prints a variable any-kind removal from itself.
			switch {
			case rc.Counter != "":
				panic(fmt.Sprintf("effects.Register: %q %s splits a %s counter removal among nothing — use RemoveCountersFromThis, or give the Among cost its \"from among …\" clause", card, where, rc.Counter))
			case rc.Variable:
				panic(fmt.Sprintf("effects.Register: %q %s removes a variable number of counters of any kinds from itself — no printed card does", card, where))
			case rc.N < 2:
				panic(fmt.Sprintf("effects.Register: %q %s removes one counter of any kind from itself as an Among cost — use RemoveCountersFromThis(\"\", 1)", card, where))
			}
		}
		if rc.Among && rc.From != nil {
			// A variable count split across permanents is a shape since
			// #3276 (Jetfire, Ingenious Scientist: "one or more +1/+1
			// counters from among artifacts you control"). Only the
			// printed form is: one kind, and a floor of at least one.
			if rc.Variable && rc.Counter == "" {
				panic(fmt.Sprintf("effects.Register: %q %s removes a variable number of counters of any kind from among several permanents — no printed card does", card, where))
			}
			if rc.Variable && rc.N < 1 {
				panic(fmt.Sprintf("effects.Register: %q %s removes \"any number\" of counters from among several permanents — no printed card does; a variable among cost prints \"one or more\" (floor 1)", card, where))
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

// checkSacrificeClause is #747's registration guard, narrowed by
// #1213 (ADR 0020 addendum §12, ADR 0073's 2026-09-22 amendment) and
// again by ADR 0100 §3.
//
// #747 accepted exactly one shape — a FIXED count of at least one,
// written Min == Max == N — because a variable count had no announced
// count and no record of what was paid. Both exist now
// (SacrificeCostBounds and PaidCost.Sacrificed), so three more shapes
// are legal where the site can announce them:
//
//   - an OPEN count, Min ≥ 1 with Max 0: "Sacrifice one or more
//     artifacts" (Radiant Lotus). The activator names how many.
//     `allowOpen`: an activated or mana ability.
//   - CountFromX: "Sacrifice X Treasures" (Grim Hireling), "sacrifice
//     X lands" (Devastating Summons). The announced X is the count on
//     both sides. `allowX`: an activated ability, or a cast's
//     mandatory additional cost (CR 107.3a).
//   - ANY NUMBER, Min 0 with Max 0 (game.SacrificeAnyNumber): "sacrifice
//     any number of creatures" (Vicious Betrayal), "you may sacrifice
//     any number of creatures" (Torgaar). `allowZero`: a cast's
//     mandatory additional cost only, where zero is printed.
//
// A cast's plan is walked against ONE flat sacrifice_ids list, so a
// variable clause on a cast is sound only as the plan's one sacrifice;
// checkVariableSacrificePlan holds that across the slots.
//
// Everything still refused is a card-file mistake that would ship the
// card as something it does not print:
//
//   - a floor below one outside a cast: a cost that can be paid with
//     nothing is free (#1213's guard, which stays for abilities).
//   - a ceiling below the floor.
//   - CountFromX where there is no X to announce — `allowX` is false
//     for a mana ability, which has no stack item to carry one
//     (CR 605.3b), and for an optional cost or an either/or branch.
//   - AllowSame: one permanent cannot pay two sacrifices.
//   - Players: a player is not a permanent.
//
// Nil (no sacrifice component) is fine.
//
// Register calls it on the sites the ADRs name: spec.Activated,
// spec.ManaAbilities, spec.AdditionalCost, its either/or branches and
// spec.OptionalCosts. An ability granted at runtime (a static grant,
// an Equipment's "equipped creature has …") is built after
// registration and is not checked here. Every such grant with a
// sacrifice clause today is a count of one, so nothing escapes the
// guard yet; a grant with a variable count would.
func checkSacrificeClause(card, where string, spec *game.TargetSpec, allowOpen, allowX, allowZero bool) {
	if spec == nil {
		return
	}
	anyNumber := game.SacrificeAnyNumber(spec)
	switch {
	case spec.CountFromX && !allowX:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices X permanents, but there is no X to announce here (CR 605.3b; only an activated ability or a cast's mandatory additional cost announces one, #1213, ADR 0100 §3)", card, where))
	case anyNumber && !allowZero:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents — a sacrifice cost pays at least one, so its floor is at least one (SacrificeN, SacrificeOneOrMore); only a cast's additional cost prints \"any number\" (SacrificeAnyNumberCost, ADR 0100 §3)",
			card, where, spec.Min, spec.Max))
	case game.SacrificeCostVariable(spec) && !spec.CountFromX && !anyNumber && !allowOpen:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents, but that variable count has no shape here — a cast's additional cost prints only \"sacrifice X\" (SacrificeXCost) or \"any number\" (SacrificeAnyNumberCost), ADR 0100 §3",
			card, where, spec.Min, spec.Max))
	case !spec.CountFromX && !anyNumber && spec.Min < 1:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents — a sacrifice cost pays at least one, so its floor is at least one (SacrificeN, SacrificeOneOrMore)",
			card, where, spec.Min, spec.Max))
	case !spec.CountFromX && spec.Max != 0 && spec.Max < spec.Min:
		panic(fmt.Sprintf("effects.Register: %q %s sacrifices %d to %d permanents — the ceiling is below the floor",
			card, where, spec.Min, spec.Max))
	case spec.AllowSame:
		panic(fmt.Sprintf("effects.Register: %q %s lets one permanent pay a sacrifice twice (AllowSame)", card, where))
	case spec.Players:
		panic(fmt.Sprintf("effects.Register: %q %s admits players — a sacrifice clause matches permanents only", card, where))
	case spec.Total != nil:
		// #4616: the total rule is read by the target readers only; no
		// sacrifice payer or cost picker sums one.
		panic(fmt.Sprintf("effects.Register: %q %s has a total rule on its sacrifice clause — only a target clause reads one", card, where))
	}
	checkSacrificeSetRule(card, where, spec)
	checkSacrificeDifferentRule(card, where, spec)
}

// checkSacrificeDifferentRule is #3940's guard for TargetSpec.Different
// on a sacrifice clause ("three artifact tokens with different names").
// The readers that honour it (the validator, the enumerator's one
// payment, the view's keys and the client picker) read a FIXED count of
// at least two, with no EachOf beside it: a count of one has nothing to
// differ from, and a variable count or a kind rule as well would need a
// search none of them makes. The rule needs its Key and its Label, as a
// target clause's does.
func checkSacrificeDifferentRule(card, where string, spec *game.TargetSpec) {
	d := spec.Different
	if d == nil {
		return
	}
	switch {
	case d.Key == nil || d.Label == "":
		panic(fmt.Sprintf("effects.Register: %q %s has a difference rule with no Key or no Label — build it with EachDifferentName", card, where))
	case spec.CountFromX || game.SacrificeCostVariable(spec) || game.SacrificeCostCount(spec) < 2:
		panic(fmt.Sprintf("effects.Register: %q %s has a difference rule on %d to %d permanents — it reads a fixed count of two or more (SacrificeNDifferent)", card, where, spec.Min, spec.Max))
	case len(spec.EachOf) > 0:
		panic(fmt.Sprintf("effects.Register: %q %s has both a difference rule and a kind rule (EachOf) — no reader searches for a set under both", card, where))
	}
}

// refuseSacrificeDifferent keeps a difference rule (#3940) off the
// sacrifice clauses whose own payers or offer checks do not read it: a
// mana ability's, a cast's additional cost (its payability check counts
// candidates) and an optional or alternative cost. Only an activated ability's SacrificeOther carries one, which
// is where the one printed card of the family (Transmutation Font) has
// it. Widen this when a card needs it elsewhere, with that site's
// payer.
func refuseSacrificeDifferent(card, where string, spec *game.TargetSpec) {
	if spec != nil && spec.Different != nil {
		panic(fmt.Sprintf("effects.Register: %q %s has a difference rule on its sacrifice clause — only an activated ability's sacrifice cost reads one (#3940)", card, where))
	}
}

// checkSacrificeSetRule is #2526's guard for TargetSpec.EachOf on a
// sacrifice clause ("a Swamp and a Forest"). The rule is a one-to-one
// matching of picks to entries, so it only means something on a FIXED
// count equal to the entry count, with at least two entries (one entry
// is an ordinary clause) that each name something to match. Anything else would register a cost the
// validator and the picker read differently.
func checkSacrificeSetRule(card, where string, spec *game.TargetSpec) {
	if len(spec.EachOf) == 0 {
		return
	}
	if len(spec.EachOf) < 2 {
		panic(fmt.Sprintf("effects.Register: %q %s has a one-entry set rule — that is an ordinary sacrifice clause (SacrificeN), not EachOf", card, where))
	}
	if spec.CountFromX || game.SacrificeCostVariable(spec) || spec.Min != len(spec.EachOf) || spec.Max != len(spec.EachOf) {
		panic(fmt.Sprintf("effects.Register: %q %s has a set rule over %d entries but sacrifices %d to %d permanents — the count must be exactly the entry count (SacrificeEach)",
			card, where, len(spec.EachOf), spec.Min, spec.Max))
	}
	checkSetRuleEntries(card, where, spec)
}

// checkSetRuleEntries refuses a set-rule entry that names nothing: no
// subtype, no card type and not "any card" (ADR 0135 §2's Any). Such an
// entry would match nothing, and the cost could never be paid.
func checkSetRuleEntries(card, where string, spec *game.TargetSpec) {
	for i, k := range spec.EachOf {
		if len(k.Subtypes) == 0 && len(k.CardTypes) == 0 && !k.Any {
			panic(fmt.Sprintf("effects.Register: %q %s set-rule entry %d names no subtype or card type — it would match nothing", card, where, i))
		}
	}
}

// checkAltCostSetRule is ADR 0135 §2's guard for a set rule on an
// alternative cost's card component. It is allowed on a sacrifice
// (checkSacrificeSetRule, #2526) and on a hand discard (Foil's "an
// Island card and another card"), held there to the same shape: at least
// two entries, a fixed count equal to their number. On the other
// components (exile from hand or graveyard, return to hand) nothing
// reads it, so the cost would be charged without its rule.
func checkAltCostSetRule(card string, ac game.AlternativeCost) {
	where := fmt.Sprintf("alternative cost %q", ac.Key)
	for _, spec := range []*game.TargetSpec{ac.ExileFromHand, ac.ReturnToHand, ac.ExileFromGraveyard} {
		if spec != nil && len(spec.EachOf) > 0 {
			panic(fmt.Sprintf("effects.Register: %q %s has a set rule (EachOf) on a component that does not read one — only a sacrifice or a hand discard does", card, where))
		}
	}
	spec := ac.DiscardFromHand
	if spec == nil || len(spec.EachOf) == 0 {
		return
	}
	if len(spec.EachOf) < 2 {
		panic(fmt.Sprintf("effects.Register: %q %s has a one-entry set rule — that is an ordinary discard clause (DiscardInstead), not EachOf", card, where))
	}
	if spec.CountFromX || spec.Min != len(spec.EachOf) || spec.Max != len(spec.EachOf) {
		panic(fmt.Sprintf("effects.Register: %q %s has a set rule over %d entries but discards %d to %d cards — the count must be exactly the entry count (DiscardEachInstead)",
			card, where, len(spec.EachOf), spec.Min, spec.Max))
	}
	checkSetRuleEntries(card, where, spec)
}

// checkAltCostXBound is #3175's guard for an alternative cost whose card
// component is bound to the announced X (Shining Shoal's "exile a white
// card with mana value X from your hand"). Only the hand pitch reads the
// bound (validateAlternativeCostPaymentLocked), only "mana value X" is a
// printed form, the pitch names one card, and the card must read its X
// (XMatters) — a bound on an X the spell never uses binds nothing.
func checkAltCostXBound(spec Spec, ac game.AlternativeCost) {
	where := fmt.Sprintf("alternative cost %q", ac.Key)
	for _, s := range []*game.TargetSpec{ac.ReturnToHand, ac.ExileFromGraveyard, ac.Sacrifice, ac.DiscardFromHand} {
		if s != nil && (s.ManaValueAtMostX || s.ManaValueEqualsX || s.PowerAtMostX || s.ToughnessAtMostX) {
			panic(fmt.Sprintf("effects.Register: %q %s bounds a component by X that does not read the bound — only a hand pitch (ExileFromHand) does", spec.Name, where))
		}
	}
	s := ac.ExileFromHand
	if s == nil {
		return
	}
	if s.ManaValueAtMostX || s.PowerAtMostX || s.ToughnessAtMostX || s.BoundByCountersRemoved {
		panic(fmt.Sprintf("effects.Register: %q %s bounds its pitch by a statistic other than \"mana value X\" — no printed pitch does", spec.Name, where))
	}
	if !s.ManaValueEqualsX {
		return
	}
	if s.Min > 1 || s.Max > 1 || s.CountFromX {
		panic(fmt.Sprintf("effects.Register: %q %s binds X to a pitch of more than one card — the card's mana value is the X", spec.Name, where))
	}
	if !spec.XMatters {
		panic(fmt.Sprintf("effects.Register: %q %s binds its pitch to X but the card does not read X (XMatters) — the bound would bind nothing the spell uses", spec.Name, where))
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

// checkExilePermanentsClause is checkReturnClause for the
// exile-a-permanent component (#1600), on both owners: a count below
// one would make the ability free (game.ExilePermanentsCost.Empty treats
// it as inert so the engine can ask without a guard, and a catalog
// declaration cannot be allowed to turn that into a free ability), and
// a card type that is not a permanent's would match nothing, leaving a
// card that registers, looks complete and can never be activated.
func checkExilePermanentsClause(card, where string, ec *game.ExilePermanentsCost) {
	if ec == nil {
		return
	}
	if ec.Count < 1 || ec.Label == "" {
		panic(fmt.Sprintf("effects.Register: %q %s exiles %d permanents you control with label %q — an exile-a-permanent cost exiles at least one and prints its clause (ExileAPermanentYouControl)",
			card, where, ec.Count, ec.Label))
	}
	if ec.CardType != "" && !slices.Contains(game.PermanentCardTypes, strings.ToLower(ec.CardType)) {
		panic(fmt.Sprintf("effects.Register: %q %s exiles a %q you control — not a permanent card type (game.PermanentCardTypes)",
			card, where, ec.CardType))
	}
	// ADR 0137's amendment: the craft variants. Each refusal is a
	// declaration that would register and then match nothing, or match
	// what the card does not print.
	if ec.GraveyardOnly && !ec.FromGraveyard {
		panic(fmt.Sprintf("effects.Register: %q %s is graveyard-only without FromGraveyard — it could never be paid", card, where))
	}
	if ec.CardType != "" && len(ec.CardTypes) > 0 {
		panic(fmt.Sprintf("effects.Register: %q %s names both CardType and CardTypes — use one", card, where))
	}
	for _, t := range ec.CardTypes {
		lt := strings.ToLower(t)
		ok := slices.Contains(game.PermanentCardTypes, lt)
		if !ok && ec.GraveyardOnly {
			ok = slices.ContainsFunc(game.ChoosableCardTypes, func(ct string) bool { return strings.EqualFold(ct, lt) })
		}
		if !ok {
			panic(fmt.Sprintf("effects.Register: %q %s exiles a %q — not a card type a material there can have", card, where, t))
		}
	}
	if ec.Color != "" && !slices.Contains([]string{"W", "U", "B", "R", "G"}, ec.Color) {
		panic(fmt.Sprintf("effects.Register: %q %s names colour %q — use a one-letter code (W, U, B, R, G)", card, where, ec.Color))
	}
	if len(ec.EachSubtype) > 0 && (ec.Count != len(ec.EachSubtype) || ec.OrMore || ec.ShareCardType) {
		panic(fmt.Sprintf("effects.Register: %q %s fills %d subtypes one-to-one with a count of %d (OrMore %v, ShareCardType %v) — the count is the list's length, alone",
			card, where, len(ec.EachSubtype), ec.Count, ec.OrMore, ec.ShareCardType))
	}
	if ec.ShareCardType && ec.Count < 2 {
		panic(fmt.Sprintf("effects.Register: %q %s shares a card type among %d material — a set rule needs two", card, where, ec.Count))
	}
}

// checkTapOthersClause is the boot-time refusal for #758's fixed-count
// and #1421's X-count tap-others component. TapOthersCost.Empty
// intentionally treats a malformed zero value as inert so engine call
// sites can be nil-safe; a catalog declaration cannot be allowed to
// turn that into a free ability.
func checkTapOthersClause(card, where string, tc *game.TapOthersCost, allowX bool) {
	if tc == nil {
		return
	}
	if tc.Filter == nil || tc.Label == "" {
		panic(fmt.Sprintf("effects.Register: %q %s taps %d permanents — a tap-others cost needs a clause and label",
			card, where, tc.Count))
	}
	switch {
	case tc.AnyNumber:
		panic(fmt.Sprintf("effects.Register: %q %s taps any number of permanents — only a spell's additional cost may (#3961)", card, where))
	case tc.UpTo:
		panic(fmt.Sprintf("effects.Register: %q %s taps up to a count of permanents — only harmonize's alternative cost may (#3931)", card, where))
	case tc.Filter.CountFromX && !allowX:
		panic(fmt.Sprintf("effects.Register: %q %s taps X permanents but a mana ability has no X announcement", card, where))
	case tc.Filter.CountFromX && tc.Count != 0:
		panic(fmt.Sprintf("effects.Register: %q %s taps both fixed %d and X permanents — choose one count", card, where, tc.Count))
	case !tc.Filter.CountFromX && tc.Count < 1:
		panic(fmt.Sprintf("effects.Register: %q %s taps %d permanents — a fixed tap-others cost needs a positive count", card, where, tc.Count))
	case tc.Filter.AllowSame:
		panic(fmt.Sprintf("effects.Register: %q %s lets one permanent pay a tap-others cost twice (AllowSame)", card, where))
	case tc.Filter.Players:
		panic(fmt.Sprintf("effects.Register: %q %s admits players — a tap-others clause matches permanents only", card, where))
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

// conditionalZoneDeclared is zoneDeclared for Spec.CastableZonesIf.
func conditionalZoneDeclared(zones []game.ConditionalCastZone, zone game.ZoneKind) bool {
	for _, z := range zones {
		if z.Zone == zone {
			return true
		}
	}
	return false
}

// checkCastsFace holds an offer that casts another face (disturb, ADR
// 0107 §4) to the one shape the cast path reads. Four refusals, each a
// card file that compiles and then casts something the card does not
// print:
//
//   - a negative face, which names nothing;
//   - an offer on a BACK face's entry ("<oracle_id>#N"). CR 702.146a
//     prints disturb on the front face, and the cast path reads the
//     offer off the card as it sits in its zone, which is front face
//     up (CR 712.8a) — an offer on the back would never be claimable;
//   - a face-down cast as well, two answers to what the spell is;
//   - a target clause of its own. The spell is the back face, and its
//     clause is the back face entry's (CR 712.8c); a rewrite on the
//     front face's offer would be read against the wrong face.
//
// The card's layout is not visible here (a Spec carries no layout), so
// "this is a transform card with that face" is the cast path's check:
// faceForClaimLocked refuses the claim on any other card.
func checkCastsFace(spec Spec, ac game.AlternativeCost) {
	if ac.CastsFace == 0 {
		return
	}
	switch {
	case ac.CastsFace < 0:
		panic(fmt.Sprintf("effects.Register: %q offers %q casting face %d — a face index is never negative", spec.Name, ac.Key, ac.CastsFace))
	case strings.Contains(spec.OracleID, "#"):
		panic(fmt.Sprintf("effects.Register: %q offers %q on a back face's entry — an offer that casts a face is printed on the front face (CR 702.146a) and read off the card in its zone, front face up (CR 712.8a)", spec.Name, ac.Key))
	case ac.FaceDown != nil:
		panic(fmt.Sprintf("effects.Register: %q offers %q casting face %d face down — pick one", spec.Name, ac.Key, ac.CastsFace))
	case ac.Targets != nil || ac.ClearsTargets:
		panic(fmt.Sprintf("effects.Register: %q offers %q casting face %d with a target rewrite — the spell's clause is that face's own entry's (CR 712.8c)", spec.Name, ac.Key, ac.CastsFace))
	}
}

// altCostCardComponents counts an alternative cost's card-shaped
// payments (#1727) — the ones whose cards ride alt_cost_ids. ADR 0135
// §1's TapOthers is one of them.
func altCostCardComponents(ac game.AlternativeCost) int {
	n := 0
	for _, spec := range []*game.TargetSpec{ac.ExileFromHand, ac.ReturnToHand, ac.ExileFromGraveyard, ac.Sacrifice, ac.DiscardFromHand} {
		if spec != nil {
			n++
		}
	}
	if ac.TapOthers != nil {
		n++
	}
	return n
}

// checkAltCostTapOthers holds a tap alternative cost (ADR 0135 §1) to the
// shape the cast path reads: the ability clause's own checks (a filter, a
// label, a positive fixed count, no AllowSame, no players), and two more.
// No "another" (ExcludeSource): the spell is not on the battlefield, so
// there is nothing for the word to exclude. No X: no printed alternative
// cost taps X permanents, and the announce path has no X to size it by.
func checkAltCostTapOthers(card string, ac game.AlternativeCost) {
	tc := ac.TapOthers
	if tc == nil {
		return
	}
	where := fmt.Sprintf("alternative cost %q", ac.Key)
	if tc.ExcludeSource {
		panic(fmt.Sprintf("effects.Register: %q %s taps \"another\" permanent — a spell being cast is not on the battlefield, so ExcludeSource names nothing", card, where))
	}
	if tc.Filter != nil && tc.Filter.CountFromX {
		panic(fmt.Sprintf("effects.Register: %q %s taps X permanents — an alternative cost taps a fixed number", card, where))
	}
	if tc.UpTo {
		// #3931: harmonize's "up to one", judged by checkHarmonize; the
		// rest of the clause is held to the ordinary shape.
		plain := *tc
		plain.UpTo = false
		tc = &plain
	}
	checkTapOthersClause(card, where, tc, false)
}

// checkHarmonize holds ReducedByTappedPower (#3931, CR 702.180a) and
// TapOthersCost.UpTo to the one shape harmonize prints: the two
// together, over up to ONE creature. The pricer reads the one permanent
// AltCostIDs names, so the flag beside anything else would price a
// discount the card does not print, and "up to" without the reduction
// is no printed alternative cost.
func checkHarmonize(card string, ac game.AlternativeCost) {
	upTo := ac.TapOthers != nil && ac.TapOthers.UpTo
	if !ac.ReducedByTappedPower && !upTo {
		return
	}
	if !ac.ReducedByTappedPower || !upTo || ac.TapOthers.Count != 1 {
		panic(fmt.Sprintf("effects.Register: %q offers %q reduced by a tapped creature's power without tapping up to one creature, or the reverse (CR 702.180a) — build it with Harmonize", card, ac.Key))
	}
}

// checkEmerge holds ReducedBySacrificedManaValue (ADR 0135 §4) to the
// shape CR 702.119a prints: beside a Sacrifice of exactly one permanent.
// The pricer reads the one permanent AltCostIDs names, so the flag on a
// cost that sacrifices nothing, or several, would price a discount the
// card does not print.
func checkEmerge(card string, ac game.AlternativeCost) {
	if !ac.ReducedBySacrificedManaValue {
		return
	}
	if ac.Sacrifice == nil || game.SacrificeCostCount(ac.Sacrifice) != 1 || game.SacrificeCostVariable(ac.Sacrifice) {
		panic(fmt.Sprintf("effects.Register: %q offers %q reduced by the sacrificed permanent's mana value without a sacrifice of exactly one permanent (CR 702.119a)", card, ac.Key))
	}
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
	for i := 0; i < spec.ClauseCount(); i++ {
		if r := spec.Clause(i).Relation; r != nil {
			if r.Slot < 0 || r.Slot >= i || spec.Clause(r.Slot).Min != 1 || spec.Clause(r.Slot).Max != 1 {
				panic(fmt.Sprintf("effects.Register: %q target clause %d must relate to one required earlier target", name, i))
			}
			if spec.Clause(i).Max != 1 || spec.Clause(i).Min > 1 {
				panic(fmt.Sprintf("effects.Register: %q target clause %d must relate at most one pick", name, i))
			}
			switch r.Kind {
			case game.TargetInControllersGraveyard, game.TargetAttachedTo, game.TargetControlledBySameController, game.TargetSharesCardType:
			default:
				panic(fmt.Sprintf("effects.Register: %q target clause %d has an unknown target relation", name, i))
			}
		}
		checkDivide(name, i, spec.Clause(i))
		checkOpponentChoice(name, i, spec.Clause(i))
		if d := spec.Clause(i).Different; d != nil && (d.Key == nil || d.Label == "") {
			panic(fmt.Sprintf("effects.Register: %q target clause %d has a set rule with no Key or no Label — build it with EachDifferentManaValue / EachDifferentController / EachDifferentName", name, i))
		}
		// #1807: the sameness rule, likewise — a rule with no key
		// constrains nothing, which is a card that reaches several
		// graveyards at once.
		if s := spec.Clause(i).Same; s != nil && (s.Key == game.TargetShareNone || s.Label == "") {
			panic(fmt.Sprintf("effects.Register: %q target clause %d has a sameness rule with no Key or no Label — build it with FromASingleGraveyard", name, i))
		}
		checkTotalRule(name, i, spec.Clause(i))
		// #1723: "mana value X or less" and "mana value X" are
		// different clauses — no printed card is both, and a spec
		// that set both would have the second WithManaValue...X()
		// call silently mean nothing (xBoundAdmits reads
		// ManaValueEqualsX first).
		if c := spec.Clause(i); c.ManaValueAtMostX && c.ManaValueEqualsX {
			panic(fmt.Sprintf("effects.Register: %q target clause %d sets both ManaValueAtMostX and ManaValueEqualsX", name, i))
		}
		// ADR 0109 §9: one statistic per bound — no printed clause
		// bounds two by one X — and a counters-removed input bounds
		// nothing without one.
		c := spec.Clause(i)
		stats := 0
		for _, set := range []bool{c.ManaValueAtMostX || c.ManaValueEqualsX, c.PowerAtMostX, c.ToughnessAtMostX} {
			if set {
				stats++
			}
		}
		if stats > 1 {
			panic(fmt.Sprintf("effects.Register: %q target clause %d bounds more than one statistic by X", name, i))
		}
		if c.BoundByCountersRemoved && stats == 0 {
			panic(fmt.Sprintf("effects.Register: %q target clause %d reads the counters removed but bounds no statistic — chain BoundByTheCountersRemoved after WithPowerAtMostX or its siblings", name, i))
		}
	}
}

// checkTotalRule is #4616's guard for TargetSpec.Total, the "total mana
// value N or less" set rule. The rule needs a statistic, a Label and a
// ceiling that is not negative. It sums CARDS, so a clause that can
// pick a player is refused (a player has no mana value). And it stands
// alone: the fillability count and the enumerator answer one set rule
// per clause, and no printed card has a total beside a key rule or an
// X bound.
func checkTotalRule(name string, i int, c *game.TargetClause) {
	r := c.Total
	if r == nil {
		return
	}
	switch {
	case r.Stat == game.TargetTotalNone || r.Label == "":
		panic(fmt.Sprintf("effects.Register: %q target clause %d has a total rule with no statistic or no Label — build it with TotalManaValueAtMost", name, i))
	case r.Max < 0:
		panic(fmt.Sprintf("effects.Register: %q target clause %d has a total rule with a negative ceiling", name, i))
	case c.Players:
		panic(fmt.Sprintf("effects.Register: %q target clause %d sums a total over a clause that can pick a player", name, i))
	case c.Different != nil || c.Same != nil || len(c.EachOf) > 0:
		panic(fmt.Sprintf("effects.Register: %q target clause %d has a total rule beside another set rule — no reader searches for a set under both", name, i))
	case c.CountFromX || c.ManaValueAtMostX || c.ManaValueEqualsX || c.PowerAtMostX || c.ToughnessAtMostX:
		panic(fmt.Sprintf("effects.Register: %q target clause %d has a total rule and an X — no printed card has both yet", name, i))
	}
}

// checkOpponentChoice validates an "of an opponent's choice" clause
// (#2055, ADR 0019's 2026-10-16 amendment). "They control" means nothing
// without an opponent choosing; and the clause is a plain one: chosen by
// one opponent, at least one target, with no division, relation or set
// rule tying it to the other clauses' picks and no X bound -- no printed
// card has one, and the walk that asks the opponent judges one clause
// alone.
func checkOpponentChoice(name string, i int, c *game.TargetClause) {
	if c.ControlledByChooser && !c.ChosenByOpponent && !c.ChosenByTriggerActor {
		panic(fmt.Sprintf("effects.Register: %q target clause %d says \"they control\" with no opponent choosing it — wrap it as TheyControl(OfAnOpponentsChoice(...))", name, i))
	}
	if c.Harmful && !slices.Contains(c.Zones, game.ZoneBattlefield) {
		panic(fmt.Sprintf("effects.Register: %q target clause %d is marked Harmful but cannot pick a permanent (#4249)", name, i))
	}
	if c.ChosenByTriggerActor {
		// #1997: "of their choice" is one plain clause, like an
		// opponent's.
		switch {
		case c.ChosenByOpponent:
			panic(fmt.Sprintf("effects.Register: %q target clause %d is chosen both by an opponent and by the triggering player", name, i))
		case c.AtRandom, c.Divide != nil, c.Relation != nil, c.Different != nil, c.Same != nil, c.Total != nil, len(c.EachOf) > 0:
			panic(fmt.Sprintf("effects.Register: %q target clause %d is the triggering player's choice and is random, divided or related — no printed card is", name, i))
		}
	}
	if !c.ChosenByOpponent {
		return
	}
	switch {
	case c.Divide != nil, c.Relation != nil, c.Different != nil, c.Same != nil, c.Total != nil, len(c.EachOf) > 0:
		panic(fmt.Sprintf("effects.Register: %q target clause %d is an opponent's choice and divides or relates its picks — no printed card does", name, i))
	case c.CountFromX, c.ManaValueAtMostX, c.ManaValueEqualsX, c.PowerAtMostX, c.ToughnessAtMostX:
		panic(fmt.Sprintf("effects.Register: %q target clause %d is an opponent's choice bounded by X — no printed card is", name, i))
	case c.Min < 1:
		panic(fmt.Sprintf("effects.Register: %q target clause %d is an opponent's choice of \"up to\" targets — no printed card is", name, i))
	}
}

// refuseOpponentChoice refuses an "of an opponent's choice" clause where
// the engine does not ask an opponent: a mode's clause, and a clause an
// optional or gift cost swaps in (#2055).
func refuseOpponentChoice(name, owner string, spec *game.TargetSpec) {
	for i := 0; i < spec.ClauseCount(); i++ {
		if spec.Clause(i).ChosenByOpponent {
			panic(fmt.Sprintf("effects.Register: %q declares an opponent's-choice target on %s clause %d — only a spell's, an activated ability's or a trigger's own clause list is asked of an opponent", name, owner, i))
		}
	}
	refuseTriggerActorChoice(name, owner, spec)
}

// refuseTriggerActorChoice refuses a "that player controls of their
// choice" clause (#1997, OfThatPlayersChoice) anywhere but a trigger's
// own clause list: only a trigger's CR 603.3d walk has an event whose
// player could be asked.
func refuseTriggerActorChoice(name, owner string, spec *game.TargetSpec) {
	for i := 0; i < spec.ClauseCount(); i++ {
		if spec.Clause(i).ChosenByTriggerActor {
			panic(fmt.Sprintf("effects.Register: %q declares a triggering player's-choice target on %s clause %d — only a trigger's own clause list has a triggering player", name, owner, i))
		}
	}
}

// checkDivide validates a divided clause (#1563): the amount is a
// positive constant, the announced X, or a registered amount rule
// standing alone (#1657), a doubling threshold only
// means something on an X amount, and the clause may not let one
// object fill two of its slots — the division is keyed by target id,
// so two picks of the same object could not be told apart (and CR
// 601.2d's "each target" is about distinct targets anyway).
func checkDivide(name string, i int, c *game.TargetClause) {
	d := c.Divide
	if d == nil {
		return
	}
	switch {
	case !d.AmountKey.IsZero() && (d.Total != 0 || d.FromX || d.DoubleFromX != 0):
		panic(fmt.Sprintf("effects.Register: %q target clause %d names an amount rule AND a fixed or X amount — the rule replaces them; use DivideBy(rule)", name, i))
	case !d.AmountKey.IsZero():
		// The rule sizes the division at announce (#1657); nothing
		// here to check but AllowSame, below.
		if c.AllowSame {
			panic(fmt.Sprintf("effects.Register: %q target clause %d divides among picks that may repeat — a division is keyed by target", name, i))
		}
	case !d.FromX && d.Total < 1:
		panic(fmt.Sprintf("effects.Register: %q target clause %d divides a fixed amount of %d — use Divide(n) with n ≥ 1, or DivideX()", name, i, d.Total))
	case d.DoubleFromX > 0 && !d.FromX:
		panic(fmt.Sprintf("effects.Register: %q target clause %d doubles a fixed divided amount — DoubleFromX only applies to DivideX", name, i))
	case c.AllowSame:
		panic(fmt.Sprintf("effects.Register: %q target clause %d divides among picks that may repeat — a division is keyed by target", name, i))
	}
}

// checkNoDivideX refuses DivideX on a trigger's clause: a trigger
// announces no X, so the amount would always be 0 and no target could
// ever be chosen (#1563).
func checkNoDivideX(name, owner string, spec *game.TargetSpec) {
	for i := 0; i < spec.ClauseCount(); i++ {
		if d := spec.Clause(i).Divide; d != nil && d.FromX {
			panic(fmt.Sprintf("effects.Register: %q divides X on %s target clause %d — a trigger announces no X", name, owner, i))
		}
	}
}

// checkNoXBound refuses an X-bound target clause (mana value, power or
// toughness against X) on an owner that announces nothing the engine
// can bind it to (#1559, #1723, ADR 0109 §9). `xAvailable` is the
// caller's answer to "does this owner announce an X":
// `a.Cost.DemandsX()` for an activated ability (CR 602.2b's X is that
// ability's own cost) and always false for a trigger, which announces
// none. `countersAvailable` is whether its cost removes a variable
// number of counters, the input a BoundByCountersRemoved clause reads.
// A spell's own clauses are checked by checkSpellXBound.
func checkNoXBound(name, owner string, spec *game.TargetSpec, xAvailable, countersAvailable bool) {
	for i := 0; i < spec.ClauseCount(); i++ {
		c := spec.Clause(i)
		if !c.ManaValueAtMostX && !c.ManaValueEqualsX && !c.PowerAtMostX && !c.ToughnessAtMostX {
			continue
		}
		if c.BoundByCountersRemoved {
			if !countersAvailable {
				panic(fmt.Sprintf("effects.Register: %q bounds %s target clause %d by the counters removed — %s removes no variable number of counters", name, owner, i, owner))
			}
			continue
		}
		if xAvailable {
			continue
		}
		switch {
		case c.ManaValueAtMostX:
			panic(fmt.Sprintf("effects.Register: %q declares \"mana value X or less\" on %s target clause %d — %s announces no X", name, owner, i, owner))
		case c.ManaValueEqualsX:
			panic(fmt.Sprintf("effects.Register: %q declares \"mana value X\" on %s target clause %d — %s announces no X", name, owner, i, owner))
		default:
			panic(fmt.Sprintf("effects.Register: %q bounds a statistic by X on %s target clause %d — %s announces no X", name, owner, i, owner))
		}
	}
}

// checkSpellXBound refuses a spell clause bounded by the counters
// removed (ADR 0109 §9): a spell's cost removes none.
func checkSpellXBound(name string, spec *game.TargetSpec) {
	for i := 0; i < spec.ClauseCount(); i++ {
		if spec.Clause(i).BoundByCountersRemoved {
			panic(fmt.Sprintf("effects.Register: %q bounds spell target clause %d by the counters removed — a spell removes none", name, i))
		}
	}
}

// checkRaisedModeMax validates a conditional mode count (#1590, ADR
// 0065's 2026-09-27 amendment): RaisedMax and RaiseMaxIf come as a
// pair, the raise must actually raise a bounded Max, and — unless the
// spec is Repeatable — it may not promise more distinct bullets than
// the card prints; a raised minimum (#1655) must sit between the
// printed Min and the raised Max. Each is a card that would otherwise
// register silently and ship a mode count the printed card does not
// have.
func checkRaisedModeMax(name, owner string, ms *game.ModeSpec) {
	if ms.RaiseMaxIf.IsZero() && ms.RaisedMax == 0 {
		return
	}
	if ms.RaiseMaxIf.IsZero() || ms.RaisedMax == 0 {
		panic(fmt.Sprintf("effects.Register: %q %s declares half a conditional mode count — use OrUpToIf / InsteadIf / AnyNumberIf", name, owner))
	}
	if ms.Max <= 0 || ms.RaisedMax <= ms.Max {
		panic(fmt.Sprintf("effects.Register: %q %s raises Max %d to %d — a conditional count must raise a bounded Max", name, owner, ms.Max, ms.RaisedMax))
	}
	if !ms.Repeatable && ms.RaisedMax > len(ms.Options) {
		panic(fmt.Sprintf("effects.Register: %q %s raises Max to %d with only %d bullets", name, owner, ms.RaisedMax, len(ms.Options)))
	}
	// #1655: a forced count ("choose both instead", InsteadIf) raises
	// the minimum too, and never past the raised maximum or below the
	// printed minimum — either would be a count no printed card has.
	if ms.RaisedMin != 0 && (ms.RaisedMin <= ms.Min || ms.RaisedMin > ms.RaisedMax) {
		panic(fmt.Sprintf("effects.Register: %q %s raises Min %d to %d with a raised Max of %d", name, owner, ms.Min, ms.RaisedMin, ms.RaisedMax))
	}
}

// checkTargetAtRandom validates "target … chosen at random" (#1815,
// game.TargetSpec.AtRandom): only a triggered ability's clause, or an
// activated ability's own (#4051, checkActivatedAtRandom), may carry
// it (`allowed`), since only those targets are drawn by the engine
// (CR 603.3d; CR 602.2b / 601.2c), and only on a clause that picks
// exactly one target and divides nothing — the shape every printed
// card has, and the only one the draw honours.
func checkTargetAtRandom(name, owner string, spec *game.TargetSpec, allowed bool) {
	for i := 0; i < spec.ClauseCount(); i++ {
		c := spec.Clause(i)
		if !c.AtRandom {
			continue
		}
		if !allowed {
			panic(fmt.Sprintf("effects.Register: %q declares a target chosen at random on %s target clause %d — only a triggered or an activated ability's target is drawn by the engine (#1815, #4051)", name, owner, i))
		}
		if c.Min != 1 || c.Max != 1 || c.Divide != nil {
			panic(fmt.Sprintf("effects.Register: %q declares a target chosen at random on %s target clause %d that is not exactly one undivided target", name, owner, i))
		}
	}
}

// checkActivatedAtRandom narrows a random clause on an activated
// ability (#4051) to the one printed shape, Goblin Test Pilot's: the
// ability's only clause, chosen by nobody else. The payload names its
// other clauses before the draw is made, so a second clause could not
// be judged against the random one (a relation, Distinct), and an
// "of an opponent's choice" clause parks the announcement in a way the
// draw's carry is not built for.
func checkActivatedAtRandom(name string, spec *game.TargetSpec) {
	random := false
	for i := 0; i < spec.ClauseCount(); i++ {
		random = random || spec.Clause(i).AtRandom
	}
	if !random {
		return
	}
	if spec.ClauseCount() != 1 || spec.ChosenByOpponent {
		panic(fmt.Sprintf("effects.Register: %q declares a target chosen at random on an activated ability with another clause — only a lone random clause is drawn at activation (#4051)", name))
	}
}

// checkAtRandom validates "choose one at random" (CR 700.2, #3399) on
// a trigger's modes: exactly one bullet, never repeated, and no count
// raise or "hasn't been chosen" memory, none of which a random pick
// can honour.
func checkAtRandom(name string, ms *game.ModeSpec) {
	if ms == nil || !ms.AtRandom {
		return
	}
	if ms.Min != 1 || ms.Max != 1 || ms.Repeatable || ms.RaisedMax != 0 || ms.NotChosen != game.ModeMemoryNone {
		panic(fmt.Sprintf("effects.Register: %q declares \"choose one at random\" that is not a plain choose-one — build it with ChooseOneAtRandom", name))
	}
}

// checkNotChosen validates a "that hasn't been chosen" restriction
// (ADR 0097) on a triggered or activated ability's modes. Refused on a
// Repeatable spec — CR 700.2d's "you may choose the same mode more
// than once" says the opposite — and on an ability with no label,
// because the label is half of the memory's key and an empty one
// would never be remembered.
func checkNotChosen(name, owner string, ms *game.ModeSpec, label string) {
	if ms == nil || ms.NotChosen == game.ModeMemoryNone {
		return
	}
	if ms.Repeatable {
		panic(fmt.Sprintf("effects.Register: %q %s is both Repeatable and \"that hasn't been chosen\" — CR 700.2d and the restriction contradict each other", name, owner))
	}
	if label == "" {
		panic(fmt.Sprintf("effects.Register: %q %s declares \"that hasn't been chosen\" on an ability with no label — the label keys the memory", name, owner))
	}
}

// checkModeCost validates one mode option's Spree cost (CR 702.172a,
// ADR 0065's 2026-09-23 amendment) at boot. `owner` names what is
// being checked in the panic message.
//
// `allowed` is false for an activated ability's or a trigger's mode:
// CR 702.172a is a static ability found on modal SPELLS, and no
// printed activated or triggered ability prices a chosen mode. A
// silent no-op cost on either of those owners would be a card that
// compiles and never charges what it means to — the same failure
// mode ADR 0073 guards against everywhere else in this file.
func checkModeCost(name, owner string, i int, cost string, allowed bool) {
	if cost == "" {
		return
	}
	if !allowed {
		panic(fmt.Sprintf("effects.Register: %q %s %d declares a Cost — Spree (CR 702.172a) prices a spell's own modes; only Spec.Modes may", name, owner, i))
	}
	if _, err := game.ParseCost(cost); err != nil {
		panic(fmt.Sprintf("effects.Register: %q %s %d declares an unparseable Cost %q: %v", name, owner, i, cost, err))
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
	// point (ActivateManaAbility, CR 605.3) and the same record, so
	// it answers to the same three rules.
	for i, a := range spec.ManaAbilities {
		checkOneExhaustAbility(spec.Name, "mana ability", i, a.Label, a.Exhaust, seen)
		// #1621: "Activate only once each turn" reads the same record
		// by the same key, so it has the same need of a label.
		if a.OncePerTurn && a.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q mana ability %d is once-each-turn with no Label — the label is the record's key", spec.Name, i))
		}
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

// checkBoast is the boot-time contract of CR 702.142 (#2697). The
// ability label and the bit must agree in both directions, for the same
// reason checkExhaustAbilities holds them to: a label that prints
// "Boast —" with no bit is repeatable every turn and attackless, and a
// bit with no printed keyword puts a rule on the card the player cannot
// read. And a boast limit without Applies reaches every creature.
func checkBoast(spec Spec) {
	seen := map[string]bool{}
	for i, a := range spec.Activated {
		if a.Boast {
			if seen[a.Label] {
				panic(fmt.Sprintf("effects.Register: %q has two boast abilities labelled %q — the activation record is keyed by the label, so they would share one use", spec.Name, a.Label))
			}
			seen[a.Label] = true
		}
		printed := strings.HasPrefix(a.Label, "Boast — ")
		if a.Boast && !printed {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d sets Boast but its label does not start with \"Boast — \" — build it with effects.Boast", spec.Name, i))
		}
		if printed && !a.Boast {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d prints \"Boast —\" and does not set Boast: true — without the bit it needs no attack and can be activated every turn", spec.Name, i))
		}
		if a.Boast && (a.Cost.Loyalty != nil || a.Equip || a.Cycling) {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d is a boast ability with a loyalty, equip or cycling shape — boast is its own ability", spec.Name, i))
		}
	}
	for i, l := range spec.BoastLimits {
		if l.Applies == nil {
			panic(fmt.Sprintf("effects.Register: %q boast limit %d has no Applies — it would reach every creature of every player, always", spec.Name, i))
		}
		if l.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q boast limit %d has no Label — the label is the printed clause", spec.Name, i))
		}
		if l.Limit < 2 {
			panic(fmt.Sprintf("effects.Register: %q boast limit %d is %d — the printed limit is already 1, so a limit below 2 changes nothing", spec.Name, i, l.Limit))
		}
	}
}

// checkPowerUpExtras refuses a PowerUpExtras entry that would reach every
// permanent of every player, has no printed clause, or adds nothing
// (#4571). checkBoast's reasons.
func checkPowerUpExtras(spec Spec) {
	for i, x := range spec.PowerUpExtras {
		if x.Applies == nil {
			panic(fmt.Sprintf("effects.Register: %q power-up extra %d has no Applies — it would reach every permanent of every player, always", spec.Name, i))
		}
		if x.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q power-up extra %d has no Label — the label is the printed clause", spec.Name, i))
		}
		if x.Times < 1 {
			panic(fmt.Sprintf("effects.Register: %q power-up extra %d adds %d activations — it must add at least one", spec.Name, i, x.Times))
		}
	}
}

// checkPowerUp is the boot-time contract of CR 702.193 (#4527), for
// checkBoast's reasons: the label and the bit agree in both directions
// (a printed "Power-up —" without the bit could be activated every turn
// at full price, and a bit without the printed keyword is a rule the
// player cannot read), two power-up abilities on one card may not share
// a label (the record is keyed by it), and the keyword is its own
// ability, so it is never a loyalty, equip or cycling shape. A
// power-up cost always has mana in it, or there is nothing to reduce.
func checkPowerUp(spec Spec) {
	seen := map[string]bool{}
	for i, a := range spec.Activated {
		printed := strings.HasPrefix(a.Label, powerUpPrefix)
		if a.PowerUp && !printed {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d sets PowerUp but its label does not start with %q — build it with effects.PowerUp", spec.Name, i, powerUpPrefix))
		}
		if printed && !a.PowerUp {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d prints %q and does not set PowerUp: true — without the bit it could be activated again and never costs less", spec.Name, i, powerUpPrefix))
		}
		if !a.PowerUp {
			continue
		}
		if seen[a.Label] {
			panic(fmt.Sprintf("effects.Register: %q has two power-up abilities labelled %q — the activation record is keyed by the label, so they would share one use", spec.Name, a.Label))
		}
		seen[a.Label] = true
		if a.Cost.Loyalty != nil || a.Equip || a.Cycling || a.Exhaust || a.Boast {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d is a power-up ability with a loyalty, equip, cycling, exhaust or boast shape — power-up is its own ability", spec.Name, i))
		}
		if a.Cost.Mana == "" {
			panic(fmt.Sprintf("effects.Register: %q activated ability %d is a power-up ability with no mana in its cost", spec.Name, i))
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
//
// PlayerKeywordsIf (#2013) is held to the same vocabulary, and each
// entry must carry its condition and its printed sentence: a nil While
// would grant nothing, silently, and an empty Label would leave the
// tile's row saying nothing about when the ability holds.
func checkPlayerKeywords(spec Spec) {
	for i, kw := range spec.PlayerKeywords {
		if !playerKeywordHonoured(kw) {
			panic(fmt.Sprintf("effects.Register: %q PlayerKeywords[%d] = %q is not a player ability the engine honours — "+
				"use %q or a \"protection from <quality>\" token the closed grammar parses (ADR 0072, #1197)",
				spec.Name, i, kw, game.KeywordHexproof))
		}
	}
	for i, kw := range spec.PlayerKeywordsIf {
		if !playerKeywordHonoured(kw.Keyword) {
			panic(fmt.Sprintf("effects.Register: %q PlayerKeywordsIf[%d].Keyword = %q is not a player ability the engine honours — "+
				"use %q or a \"protection from <quality>\" token the closed grammar parses (ADR 0072, #2013)",
				spec.Name, i, kw.Keyword, game.KeywordHexproof))
		}
		if kw.While == nil || kw.Label == "" {
			panic(fmt.Sprintf("effects.Register: %q PlayerKeywordsIf[%d] needs a condition and a label — "+
				"build it with YouHaveDuringYourTurn or YouHaveWhile (#2013)", spec.Name, i))
		}
	}
}

// playerKeywordHonoured reports whether a token is one the player walk's
// consumers read: bare hexproof, the chosen-card-type placeholder, or a
// protection or hexproof-from token the closed grammar parses.
func playerKeywordHonoured(kw string) bool {
	if kw == game.KeywordHexproof || kw == game.ProtectionFromTheChosenCardType {
		return true
	}
	if _, ok := game.ParseProtectionQuality(kw); ok {
		return true
	}
	_, ok := game.ParseHexproofFromQuality(kw)
	return ok
}

// checkAbilityWaterbend refuses, at boot, an activated ability whose
// waterbend clause (#1310) could not be paid as printed: no key or no
// pool of permanents, an unparseable clause, or a clause asking for
// more generic (or more {X}) than the mana component charges.
func checkAbilityWaterbend(name string, i int, cost game.AbilityCost) {
	wb := cost.Waterbend
	if wb.Key == "" || wb.Spec == nil || wb.Extra == "" {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares a waterbend clause with no key, pool or cost — build it with WaterbendCost", name, i))
	}
	clause, err := game.ParseCost(wb.Extra)
	if err != nil {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares an unparseable waterbend cost %q: %v", name, i, wb.Extra, err))
	}
	mana, err := game.ParseCost(cost.Mana)
	if err != nil || cost.Mana == "" || clause.Generic > mana.Generic || clause.XSlots > mana.XSlots {
		panic(fmt.Sprintf("effects.Register: %q ability %d waterbends %q but its mana component %q does not contain it — CR 701.67b lets the taps pay only mana the ability charges",
			name, i, wb.Extra, cost.Mana))
	}
}

// checkEnergyCost is the boot-time refusal for ADR 0129 §2's energy
// component:
//
//   - a negative Energy is a card-file mistake (zero is "no such
//     component");
//   - "Pay X {E}" beside another component that claims the announced X
//     as a count (sacrifice X, tap X) is refused, the same rule #1213
//     makes for {X} in the mana: one announced X cannot be a count of
//     permanents and an amount of energy, and no printed card asks it to
//     be.
func checkEnergyCost(name string, i int, cost game.AbilityCost) {
	if cost.Energy < 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares a negative energy cost %d", name, i, cost.Energy))
	}
	if cost.EnergyX && (game.SacrificeCountFromX(cost.SacrificeOther) || game.TapOthersCountFromX(cost.TapOthers)) {
		panic(fmt.Sprintf("effects.Register: %q ability %d pays X energy AND counts permanents from X — one announced X cannot pay both",
			name, i))
	}
}

// checkLoyaltyX is the boot-time refusal for a −X loyalty cost (#1944):
//
//   - LoyaltyX with no loyalty component, or beside a printed loyalty
//     other than 0, is a card-file mistake: every printed −X is the
//     whole cost, and LoyaltyMinusX builds it;
//   - beside any other component that claims the announced X ({X} in
//     the mana, "Pay X {E}", a count of sacrificed, tapped, discarded or
//     revealed cards) it is refused: one announced X cannot pay both,
//     and no printed card asks it to.
func checkLoyaltyX(name string, i int, cost game.AbilityCost) {
	if !cost.LoyaltyX {
		return
	}
	if cost.Loyalty == nil || *cost.Loyalty != 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d sets LoyaltyX without a zero loyalty cost — build it with LoyaltyMinusX", name, i))
	}
	if cost.XSlots() > 0 || cost.EnergyX || game.SacrificeCountFromX(cost.SacrificeOther) || game.TapOthersCountFromX(cost.TapOthers) ||
		game.DiscardCountFromX(cost.DiscardCards) || game.DiscardManaValueX(cost.DiscardCards) || game.RevealCardsCountFromX(cost.RevealCards) {
		panic(fmt.Sprintf("effects.Register: %q ability %d removes X loyalty AND pays X another way — one announced X cannot pay both", name, i))
	}
}

// checkLibraryCosts is the boot-time refusal for ADR 0109 §7's three
// components (#1902, owner decision 3):
//
//   - a negative count on either library component is a card-file
//     mistake (zero is "no such component");
//   - both library components on one cost are refused: the put would
//     change which cards the exile takes, and no printed card has both;
//   - a random discard with a predicate is refused: every printed
//     clause discards "a card" or "two cards", and a random draw from
//     a filtered hand is not a rule anything prints;
//   - a random discard beside the card the ability is activated from
//     being discarded (cycling) is refused for the same reason.
func checkLibraryCosts(name string, i int, cost game.AbilityCost) {
	if cost.PutFromHandOnLibraryTop < 0 || cost.ExileFromLibraryTop < 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares a negative library cost (%d on top, %d exiled)",
			name, i, cost.PutFromHandOnLibraryTop, cost.ExileFromLibraryTop))
	}
	// #3845: and the mill component, which is the same kind of
	// mistake and one more thing: a mill beside an exile of the top
	// would have to be paid from a library that must hold both counts,
	// which no printed cost asks.
	if cost.Mill < 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d declares a negative mill cost %d", name, i, cost.Mill))
	}
	if cost.Mill > 0 && cost.ExileFromLibraryTop > 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d both mills and exiles the top of the library as a cost — no printed cost does both", name, i))
	}
	if cost.PutFromHandOnLibraryTop > 0 && cost.ExileFromLibraryTop > 0 {
		panic(fmt.Sprintf("effects.Register: %q ability %d both puts a card on top of the library and exiles the top of it — no printed cost does both",
			name, i))
	}
	if dc := cost.DiscardCards; dc != nil && dc.Random {
		if dc.Match != nil {
			panic(fmt.Sprintf("effects.Register: %q ability %d discards at random with a predicate — a random discard takes any card",
				name, i))
		}
		if cost.DiscardSelf {
			panic(fmt.Sprintf("effects.Register: %q ability %d discards itself and discards at random — no printed cost does both",
				name, i))
		}
	}
}

// checkExileCardsClause is the boot-time refusal for an ExileCards cost
// component, shared by its two owners (#1283's mana ability, #1297's CR
// 602 ability) so the rules cannot drift between them:
//
//   - a clause that exiles no cards makes the ability free, the refusal
//     a zero-card discard gets;
//   - a clause that names a pile other than the hand or the graveyard
//     would never find a card to pay with, and would look complete on
//     the catalogue page while refusing every activation.
func checkExileCardsClause(name, where string, ec *game.ExileCost) {
	if ec == nil {
		return
	}
	if ec.CountFromX {
		// #1948: "exile X cards" exists on a cast's mandatory slot only
		// (checkCastExileCardsClause); no ability prints it.
		panic(fmt.Sprintf("effects.Register: %q %s exiles X cards — the X form is a cast's additional cost only (#1948)",
			name, where))
	}
	if ec.N <= 0 {
		panic(fmt.Sprintf("effects.Register: %q %s exiles %d cards — an exile cost exiles at least one",
			name, where, ec.N))
	}
	if !game.ExileCostZoneSupported(ec.Zone()) {
		panic(fmt.Sprintf("effects.Register: %q %s exiles cards from the %s — an exile cost reads the hand or the graveyard",
			name, where, ec.Zone()))
	}
}

// checkAbilityCostModifiers is the boot-time refusal for an activated
// ability's OWN cost clause (ActivatedAbility.CostModifiers, #1296).
// Every shape it refuses is one the engine would silently ignore or
// refuse at every activation, so it fails here instead:
//
//   - a clause on an ability with no mana component prices nothing —
//     the CR 601.2f pass never runs for it — and would read as a real
//     discount on the catalogue page;
//   - a CostFloor: no printed ability sets a floor on its own cost
//     ("can't reduce … to less than one mana" is Power Artifact's, a
//     board clause about OTHER abilities);
//   - SpecialActions or a designation gate, which mean nothing in a
//     slot that prices exactly one ability;
//   - a mana Unit the engine would refuse (ADR 0048 addendum §16).
func checkAbilityCostModifiers(name string, i int, ab ActivatedAbility) {
	for j, m := range ab.CostModifiers {
		where := fmt.Sprintf("ability %d cost modifier %d (%q)", i, j, m.Label)
		if ab.Cost.Mana == "" {
			panic(fmt.Sprintf("effects.Register: %q %s modifies an ability with no mana cost — there is nothing to price", name, where))
		}
		if m.Kind == game.CostFloor || m.Kind == game.CostZeroAlternative {
			panic(fmt.Sprintf("effects.Register: %q %s is a CostFloor or CostZeroAlternative — an ability's own cost clause increases or reduces", name, where))
		}
		if m.SpecialActions {
			panic(fmt.Sprintf("effects.Register: %q %s sets SpecialActions — an ability's own clause prices that ability", name, where))
		}
		if m.ActiveWhen != (game.Designation{}) {
			panic(fmt.Sprintf("effects.Register: %q %s declares a designation gate — gate the ability (ActivatedAbility.ActiveWhen) instead", name, where))
		}
		if why := m.UnitProblem(); why != "" {
			panic(fmt.Sprintf("effects.Register: %q %s declares %s (ADR 0048 addendum §16)", name, where, why))
		}
	}
}

// checkNotPutByThisAbility holds TriggeredAbility.NotPutByThisAbility
// (#3933, Kodama of the East Tree) to the shape it reads: an entry
// trigger (it watches EventETB, the event the engine stamps with the
// ability that made the entry) with a declared Effect, so the row has a
// catalog identity to compare with. Anything else would silently
// trigger off its own puts.
func checkNotPutByThisAbility(card string, t game.TriggeredAbility) {
	if !t.NotPutByThisAbility {
		return
	}
	watchesETB := false
	for _, k := range t.Watches {
		if k == game.EventETB {
			watchesETB = true
		}
	}
	if !watchesETB || t.Effect == nil || t.Build != nil {
		panic(fmt.Sprintf("effects.Register: %q declares %q not put by this ability without an EventETB watch and a declared Effect (#3933)", card, t.Key))
	}
}
