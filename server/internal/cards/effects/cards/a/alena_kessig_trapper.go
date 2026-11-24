package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alena, Kessig Trapper — Legendary Creature — Human Scout {4}{R}, 4/3
// (Commander Legends):
//
//	"First strike
//	 {T}: Add an amount of {R} equal to the greatest power among creatures
//	 you control that entered this turn.
//	 Partner"
//
// A mana ability whose amount is a pure read of the board when it is
// activated (CR 605.3b), after the {T} is paid: Gaea's Cradle's count with
// a greatest-power measure. The creatures are those the controller controls
// now that entered the battlefield this turn (Alena herself counts if she
// entered this turn, though her {T} makes her summoning-sick anyway); with
// none, or only creatures of power 0 or less, the ability adds nothing.
// Partner is a deck-construction rule (CR 702.124).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ca349ba9-b5ea-48e6-a617-c9b7f2708853",
		Name:            "Alena, Kessig Trapper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{Tap: true},
			ProducedAmount: func(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
				g.RecomputeLayersIfStaleLocked()
				n := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == controller && c.IsCreature() && g.EnteredThisTurn(c.InstanceID) {
						n = max(n, c.CurrentPower())
					}
				}
				return game.ManaAmount{Options: []string{"R"}, N: n}
			},
			Label: "Add an amount of {R} equal to the greatest power among creatures you control that entered this turn",
		}},
	})
}
