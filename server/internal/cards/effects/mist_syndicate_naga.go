package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "08109241-690f-4d09-ad2f-1cddab9b3c64",
		Name:         "Mist-Syndicate Naga",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Ninjutsu("{2}{U}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
				"Mist-Syndicate Naga — create a token that's a copy of this creature",
				func(g *game.Game, item *game.StackItem) error {
					return CreateTokenCopy{Controller: item.Controller, Copy: item.SourceCardID, N: 1}.Apply(NewContext(g, item))
				}),
		},
	})
}
