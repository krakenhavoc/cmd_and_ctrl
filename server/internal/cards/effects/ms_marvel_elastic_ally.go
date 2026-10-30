package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ms. Marvel, Elastic Ally — Legendary Creature — Mutant Inhuman Hero
// {2}{G/W}, 2/2 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 When Ms. Marvel enters, target creature gets +2/+0 until end of turn.
//	 Whenever a creature you control with power greater than its base power
//	 deals combat damage to a player, draw a card. This ability triggers
//	 only once each turn."
//
// The enters trigger is Rampart Hunter's targeted pump. The draw is
// Kutzil, Malamet Exemplar's "power greater than its base power" test
// (Card.BasePower, CR 208.4b), but per creature rather than per batch, as
// printed: it is read off the creature as it deals the damage, so a creature
// pumped by the enters trigger, a counter or an anthem counts and one whose
// power was only set does not. "Only once each turn" is the per-turn trigger
// record (b11TriggeredThisTurn, as Ant-Man reads it): a trigger already put
// on the stack this turn keeps every later hit from triggering.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "26a1b36c-3c27-4eb4-b85f-63459653b773",
		Name:            "Ms. Marvel, Elastic Ally",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WhenThisEnters(msMarvelPumpLabel, msMarvelPump), TargetCreature("target creature")),
				ForTargets(BenefitUntilEOT(0, 2, 0))),
			On(game.EventDealDamage, msMarvelHitWithBoostedPower, msMarvelDrawLabel, Do(DrawCards{N: 1})),
		},
	})
}

const (
	msMarvelPumpLabel = "Ms. Marvel, Elastic Ally — target creature gets +2/+0 until end of turn"
	msMarvelDrawLabel = "Ms. Marvel, Elastic Ally — draw a card"
)

func msMarvelPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return nil
	}
	return BoostUntilEOT{Target: ts[0].ID, Power: 2, Label: "Ms. Marvel, Elastic Ally — +2/+0"}.Apply(ctx)
}

// msMarvelHitWithBoostedPower is the When of the draw: one of the source
// controller's creatures with power greater than its base power dealt combat
// damage to a player, and this ability has not been put on the stack yet
// this turn.
func msMarvelHitWithBoostedPower(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !combatDamageToPlayerBy(ev, source.Controller, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Source)
	if !ok || c.PowerForComparison() <= c.BasePower() {
		return false
	}
	return !b11TriggeredThisTurn(g, source.InstanceID, msMarvelDrawLabel)
}
