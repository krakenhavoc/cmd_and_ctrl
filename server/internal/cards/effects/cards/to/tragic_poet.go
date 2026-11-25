package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tragic Poet — Creature — Human {W}, 1/1 (Dominaria, S82 OLD, #5530):
//
//	"{T}, Sacrifice this creature: Return target enchantment card from your
//	 graveyard to your hand."
//
// The tap and the sacrifice are paid on announcement (CR 602.2b), so the Poet
// is in the graveyard before the ability resolves and cannot return itself (it
// is not an enchantment). The target is a card in the activator's own
// graveyard, rechecked on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97bdfbe0-7048-4157-8a58-5b17410750ab",
		Name:         "Tragic Poet",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Return target enchantment card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCardInGraveyard("target enchantment card from your graveyard", effects.Enchantment(), effects.YouOwn()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
			},
		}},
	})
}
