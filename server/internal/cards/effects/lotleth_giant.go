package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5e0bd136-e7e9-49f0-910b-a9fb0b9c481a",
		Name:         "Lotleth Giant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Lotleth Giant — it deals 1 damage to target opponent for each creature card in your graveyard",
					lotlethGiantDamage),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

func lotlethGiantDamage(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := creatureCardsInGraveyard(g, ctx.Controller())
	if n == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
		}
	}
	return nil
}
