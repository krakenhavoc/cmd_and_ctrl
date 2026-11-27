package game

import "github.com/google/uuid"

// land_mana_colorless.go — Pale Moon (Nemesis, S82 OLD batch 679, #5618):
//
//	"Until end of turn, if a player taps a nonbasic land for mana, it
//	 produces colorless mana instead of any other type."
//
// A CR 614.1a replacement on the mana production event (CR 106.12b), a
// resolved spell's effect (CR 611.2a) and so a ScopedEffect record of its
// own kind. It reaches EVERY player's nonbasic lands, not only the
// caster's. "Instead of any other type" keeps the amount: a nonbasic land
// that makes {W}{U} makes {C}{C} (Contamination's "type and amount" is a
// different card). Mana the land already makes colourless is left alone,
// and a basic land is untouched.

// ModNonbasicLandManaColorless is "if a player taps a nonbasic land for
// mana, it produces colorless mana instead of any other type this turn"
// (Pale Moon). No parameters. Scope ScopeGame. A replacement kind
// (readerReplacement), never the layer pass. An on-disk identity: never
// renamed, never reused.
const ModNonbasicLandManaColorless ModKind = "nonbasicLandManaColorless"

// NonbasicLandsProduceColorlessThisTurnForEffect registers Pale Moon's
// replacement until end of turn, controlled by `controller`. Reports
// whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) NonbasicLandsProduceColorlessThisTurnForEffect(sourceID, controller uuid.UUID, label string) bool {
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModNonbasicLandManaColorless}}, g.UntilEndOfTurnDuration(), label)
}

// nonbasicLandManaColorlessAppliesLocked: a nonbasic land tapped for mana
// (CR 106.12) that is producing some mana that is not already colourless.
func nonbasicLandManaColorlessAppliesLocked(g *Game, _ Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventProduceMana || !ev.ManaFromTap {
		return false
	}
	changes := false
	for _, c := range ev.ManaColors {
		if c != "C" {
			changes = true
			break
		}
	}
	if !changes {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.ManaSource)
	return ok && land.IsLand() && !land.HasSupertype("basic")
}

// applyNonbasicLandManaColorlessLocked turns every mana of the production
// colourless, keeping the amount.
func applyNonbasicLandManaColorlessLocked(ev *ReplacementEvent) {
	out := make([]string, len(ev.ManaColors))
	for i := range out {
		out[i] = "C"
	}
	ev.ManaColors = out
}
