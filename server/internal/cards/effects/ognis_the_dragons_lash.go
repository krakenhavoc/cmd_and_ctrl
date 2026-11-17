package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ognis, the Dragon's Lash — Legendary Creature — Lizard Warrior
// {1}{B/R}{R}{R/G}, 3/3 (Streets of New Capenna):
//
//	"Haste
//	 Whenever a creature you control with haste attacks, create a tapped
//	 Treasure token."
//
// One trigger per attacking creature you control that has haste as it is
// declared, read through the layers (CR 508.2a), Ognis included. The
// Treasure enters tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a6e50e06-6d1d-457d-8fbb-ec9a866960c3",
		Name:            "Ognis, the Dragon's Lash",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !attackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&attacker, "haste")
			}, "Ognis, the Dragon's Lash — create a tapped Treasure token",
				Do(CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
