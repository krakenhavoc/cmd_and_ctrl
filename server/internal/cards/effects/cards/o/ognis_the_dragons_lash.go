package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "a6e50e06-6d1d-457d-8fbb-ec9a866960c3",
		Name:            "Ognis, the Dragon's Lash",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				attacker, ok := g.LookupCardForEffect(ev.CardID)
				return ok && game.HasKeyword(&attacker, "haste")
			}, "Ognis, the Dragon's Lash — create a tapped Treasure token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: 1})),
		},
	})
}
