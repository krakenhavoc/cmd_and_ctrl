package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Pallimud — Creature — Beast {2}{R}, */3 (Tempest, S82 OLD batch 768, #5628):
//
//	"As this creature enters, choose an opponent.
//	 Pallimud's power is equal to the number of tapped lands the chosen player
//	 controls."
//
// The as-enters choice is Nyxathid's: the controller names an opponent, stored on
// the permanent. The power is a characteristic-defining ability (CR 604.3,
// 613.4a, layer 7a), counted live as the layers are recomputed: tapping or
// untapping a land, or one entering or leaving, moves it, and a creature whose
// opponent has not been chosen yet counts none. Its toughness is the printed 3.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cad0f865-62e0-47ba-84ae-8e97cb52bbd1",
		Name:         "Pallimud",
		Completeness: effects.CompletenessFull,
		AsEnters:     effects.ChoosePlayerAsEnters("Pallimud", effects.Opponents),
		Static: []game.StaticAbility{{
			Label:     "Pallimud's power is equal to the number of tapped lands the chosen player controls.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power = TappedLandsControlledBy(g, effects.ChosenPlayerOf(g, source.InstanceID))
			},
		}},
	})
}

// TappedLandsControlledBy is the number of tapped lands `player` controls;
// zero for no player (uuid.Nil).
func TappedLandsControlledBy(g *game.Game, player uuid.UUID) int {
	if player == uuid.Nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == player && c.IsLand() && c.Tapped {
			n++
		}
	}
	return n
}
