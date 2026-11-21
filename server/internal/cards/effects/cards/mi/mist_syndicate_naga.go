package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mist-Syndicate Naga — Creature — Snake Ninja {2}{U}, 3/1 (Modern Horizons, #5519):
//
//	"Ninjutsu {2}{U} ({2}{U}, Return an unblocked attacker you control to
//	 hand: Put this card onto the battlefield from your hand tapped and
//	 attacking.)
//	 Whenever this creature deals combat damage to a player, create a token
//	 that's a copy of this creature."
//
// Moonblade Shinobi's ninjutsu and Giant Adephage's copy trigger. Each
// copy has the trigger too, so the Nagas multiply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "08109241-690f-4d09-ad2f-1cddab9b3c64",
		Name:         "Mist-Syndicate Naga",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Ninjutsu("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayer,
				"Mist-Syndicate Naga — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
