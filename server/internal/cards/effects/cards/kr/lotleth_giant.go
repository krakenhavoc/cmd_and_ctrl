package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lotleth Giant — Creature — Zombie Giant {6}{B}, 6/5 (Commander Masters):
//
//	"Undergrowth — When this creature enters, it deals 1 damage to target
//	 opponent for each creature card in your graveyard."
//
// Kraul Harpooner's creature-card count (creatureCardsInGraveyard, read as
// the ability resolves, CR 608.2) as damage from the Giant to the targeted
// opponent. "Undergrowth" is an ability word with no rules meaning. With no
// creature cards no damage is dealt, and an illegal target does nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e0bd136-e7e9-49f0-910b-a9fb0b9c481a",
		Name:         "Lotleth Giant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Lotleth Giant — it deals 1 damage to target opponent for each creature card in your graveyard",
					lotlethGiantDamage),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}

func lotlethGiantDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	n := effects.CreatureCardsInGraveyard(g, ctx.Controller())
	if n == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
	}
	return nil
}
