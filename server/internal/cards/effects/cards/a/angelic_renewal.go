package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelic Renewal — Enchantment {1}{W} (Weatherlight, S82 older sets batch 778, #5629):
//
//	"Whenever a creature is put into your graveyard from the battlefield, you may sacrifice this
//	 enchantment. If you do, return that card to the battlefield."
//
// Mortuary's trigger (a creature card you own went from the battlefield to a graveyard; a token is
// not a card). The "may" is asked as the trigger resolves (CR 603.5); only if the enchantment really was
// sacrificed does the card return, and only if it is still in the graveyard as that card (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f86bcdde-8b1f-4efa-a01d-39e09dc826d0",
		Name:         "Angelic Renewal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && dead.Owner == source.Controller
			}, "Angelic Renewal — you may sacrifice this enchantment; if you do, return that card to the battlefield",
				effects.Do(effects.MayChoice{
					Question: "Angelic Renewal — sacrifice this enchantment to return that card to the battlefield?",
					YesLabel: "Sacrifice",
					NoLabel:  "Decline",
					Then:     effects.S82OldB367SacrificeAuraAnswer("confirm/s82-old-b778-angelic-renewal-sacrifice", effects.S82OldB778ReturnTheDiedCardTail),
				})),
		},
	})
}
