package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rootwater Mystic — Creature — Merfolk Wizard {U}, 1/1 (Exodus, S82 OLD b737,
// #5626):
//
//	"{1}{U}: Look at the top card of target player's library."
//
// Mishra's Bauble's look without the draw: only the controller sees the
// card (CR 701.20 is not a reveal) and it stays on top.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32d8f7c8-8680-4901-b31b-e0cb66e37548",
		Name:         "Rootwater Mystic",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Look at the top card of target player's library.",
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.LookAtLibraryThenPlace{
					Owner:     effects.TargetedPlayer(ctx),
					N:         1,
					Placement: game.LibraryPlaceTop,
					Label:     "Rootwater Mystic — look at the top card",
				}.Apply(ctx)
			},
		}},
	})
}
