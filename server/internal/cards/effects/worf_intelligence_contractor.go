package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Worf, Intelligence Contractor — Legendary Creature — Klingon Warrior
// Spy {2}{R}, 3/2 (Star Trek, #4973):
//
//	"When Worf enters, for each opponent, gain control of up to one
//	 target creature that player controls with mana value 3 or less
//	 until end of turn. Untap those creatures. Until end of turn, they
//	 become Warriors in addition to their other types and gain haste."
//
// The ETB is Molten Primordial's shape: one "up to one target creature
// that player controls" clause per opponent of Worf's controller as he
// entered, in seat order, each pick bound to its player and re-checked
// at resolution (CR 608.2b), here narrowed to mana value 3 or less
// (ManaValueLE, read off the permanent's mana cost, CR 202.3). Each
// still-legal pick is Coercive Recruiter's four primitives
// (msc13ThreatenAndPaint): gain control until end of turn, untap, haste
// until end of turn, and a layer-4 Warrior subtype add pinned to that
// creature's own entry (CR 400.7, CR 611.2c). "Warrior" is read off the
// card's own text (ctx.TextSubtype).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Worf, Intelligence Contractor — gain control of up to one creature with mana value 3 or less each opponent controls",
		worfIntelligenceContractorSteal)
	// ADR 0041 P9: the clause list reads only the controller as Worf
	// entered (tc.Object) and the seat list, neither of which a restore
	// can change, so the row is not TargetsFromReadsBoard.
	enters.TargetsFrom = worfIntelligenceContractorClauses
	Register(Spec{
		OracleID:     "dfff6f8f-b038-4619-8190-7a6918afab1e",
		Name:         "Worf, Intelligence Contractor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}

// worfIntelligenceContractorClauses is one "up to one target creature
// that player controls with mana value 3 or less" clause per opponent
// seat, in seat order, the label naming the player. "Opponent" is
// relative to Worf's controller as he entered (moltenPrimordialClauses
// gives the reason).
func worfIntelligenceContractorClauses(tc game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
	controller := source.Controller
	if obj := tc.Object; obj != nil && obj.ID == source.InstanceID && obj.Controller != uuid.Nil {
		controller = obj.Controller
	}
	var clauses []*game.TargetSpec
	for _, p := range g.Seats {
		if p == nil || p.ID == controller {
			continue
		}
		clauses = append(clauses, TargetCreature("up to one target creature "+p.Name+" controls with mana value 3 or less",
			ControlledBy(p.ID), ManaValueLE(3)).WithCount(0, 1))
	}
	if len(clauses) == 0 {
		return nil
	}
	return Clauses(clauses[0], clauses[1:]...)
}

// worfIntelligenceContractorSteal takes each still-legal pick until end
// of turn, untaps it, and makes it a Warrior with haste until end of
// turn.
func worfIntelligenceContractorSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := msc13ThreatenAndPaint(ctx, t.ID, "Worf, Intelligence Contractor", ctx.TextSubtype("Warrior")); err != nil {
			return err
		}
	}
	return nil
}
