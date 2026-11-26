package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lifebane Zombie — Creature — Zombie Warrior {1}{B}{B}, 3/1 (Magic 2014, S82
// old sets):
//
//	"Intimidate
//	 When this creature enters, target opponent reveals their hand. You choose a
//	 green or white creature card from it and exile that card."
//
// Covetous Urge's revealed-hand pick without the graveyard, narrowed to a
// creature card that is green, white or both, and exiling the pick rather than
// discarding it. The target is chosen as the trigger goes on the stack (CR
// 603.3d); with no green or white creature card in the hand nothing is exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b31ac7b-7901-484c-b59f-556284fd562c",
		Name:            "Lifebane Zombie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Lifebane Zombie — target opponent reveals their hand; you choose a green or white creature card from it and exile that card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ChooseFromRevealedHand{
						Player: effects.TargetedPlayerOf(ctx),
						Filter: effects.And(effects.Creature(), effects.Or(effects.OfColor("G"), effects.OfColor("W"))),
						Label:  "green or white creature card",
						Exile:  true,
					}.Apply(ctx)
				}),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
