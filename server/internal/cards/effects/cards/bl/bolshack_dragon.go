package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bolshack Dragon — Creature — Armored Dragon {5}{R}, 6/6 (Mystery Booster 2):
//
//	"Double b̶r̶e̶a̶k̶e̶r̶ strike
//	 While attacking, this creature gets +10̶0̶0̶ power for each f̶i̶r̶e̶ red
//	 card in your graveyard."
//
// The struck-through words are part of the printed joke and not of the rules
// text: the card has double strike and, while attacking, +1 power for each red
// card in its controller's graveyard. The bonus is a layer 7c modification
// (CR 613.4c) read live, so it follows the graveyard as cards come and go and
// is nothing once the Dragon stops attacking. A red card is judged by its
// printed colors in the graveyard (CR 109.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8714939-aaf9-4ede-a04f-61f0df5ea4d4",
		Name:            "Bolshack Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{{
			// The bonus reads attacking status, so the layers are redone when it changes.
			DependsOnAttackingStatus: true,
			Label:                    "While attacking, this creature gets +1/+0 for each red card in your graveyard.",
			Layer:                    game.Layer7PT,
			SubLayer:                 game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil || p.Graveyard == nil {
					return
				}
				for _, gc := range p.Graveyard.Cards {
					if gc.HasColor(source.TextColor("R")) {
						c.Power++
					}
				}
			},
		}},
	})
}
