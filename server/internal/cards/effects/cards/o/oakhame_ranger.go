package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oakhame Ranger // Bring Back — Creature — Elf Knight Ranger {G/W}{G/W}{G/W}{G/W}, 2/2 // Sorcery — Adventure {G/W}{G/W}{G/W}{G/W} (Throne of Eldraine, #5516):
//
//	Oakhame Ranger — "{T}: Creatures you control get +1/+1 until end of turn."
//	Bring Back — "Create two 1/1 white Human creature tokens."
//
// Both faces register: the creature under the oracle ID and the Adventure
// under "<oracle_id>#1". The Adventure lifecycle (exile on resolution, cast
// the creature later, CR 715) is the engine's.
//
// No simplification.
const oakhameRangerOracleID = "9c247a17-361c-435e-90c8-ce9c6b71f7db"

func init() {
	effects.Register(effects.Spec{
		OracleID:     oakhameRangerOracleID,
		Name:         "Oakhame Ranger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Creatures you control get +1/+1 until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "Oakhame Ranger — creatures you control get +1/+1"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oakhameRangerOracleID + "#1",
		Name:         "Bring Back",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 2}),
	})
}
