package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coral Fighters — Creature — Merfolk Soldier {1}{U}, 1/1 (Mirage, S82 OLD
// batch 810, #5633):
//
//	"Whenever this creature attacks and isn't blocked, look at the top card of
//	 defending player's library. You may put that card on the bottom of that
//	 player's library."
//
// Jace, the Mind Sculptor's +2 as an unblocked-attacker trigger: the top card
// of the defending player's library is looked at by this creature's controller,
// who chooses top or bottom.
//
// No simplification.
func init() {
	const label = "Coral Fighters — look at the top card of defending player's library; you may put it on the bottom"
	effects.Register(effects.Spec{
		OracleID:     "8a397a64-542f-43fb-bf54-27ac651e1819",
		Name:         "Coral Fighters",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenAttacksAndIsNotBlockedEffect(label,
				func(g *game.Game, item *game.StackItem, defender uuid.UUID) error {
					if !effects.DefendingPlayerStillIn(g, defender) {
						return nil
					}
					return effects.LookAtLibraryThenPlace{
						Owner:     defender,
						N:         1,
						Placement: game.LibraryPlaceTopOrBottom,
						Label:     label,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
