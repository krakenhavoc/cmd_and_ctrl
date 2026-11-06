package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Splinter, Aging Champion — Legendary Creature — Mutant Ninja Rat
// {2}{W}, 2/2 (TMNT Commander, #5011):
//
//	"When Splinter enters, destroy up to one target tapped creature.
//	 When Splinter leaves the battlefield, you and another target player
//	 each draw a card."
//
// Two triggers. The first takes zero or one tapped creature, re-checked
// as it resolves, so a creature untapped in response is not destroyed
// (CR 608.2b). The second's "another target player" is any player but
// you; if that player is an illegal target as it resolves the ability
// does nothing at all and you don't draw either (the card's ruling, CR
// 608.2b), and otherwise you draw and then they do.
//
// No simplification.
func init() {
	leaves := WhenThisLeaves("Splinter, Aging Champion — you and another target player each draw a card",
		youAndTargetOpponentEachDraw())
	leaves.Targets = TargetPlayer("another target player", Opponent())
	leaves.Purpose = game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 1})}
	Register(Spec{
		OracleID:     "45a87131-ae8a-4e6d-9ac0-e52b4795691b",
		Name:         "Splinter, Aging Champion",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Splinter, Aging Champion — destroy up to one target tapped creature",
					func(g *game.Game, item *game.StackItem) error {
						return destroyEachLegalTarget(item, NewContext(g, item))
					}),
				TargetCreature("up to one target tapped creature", tappedPermanent()).WithCount(0, 1)),
			leaves,
		},
	})
}
