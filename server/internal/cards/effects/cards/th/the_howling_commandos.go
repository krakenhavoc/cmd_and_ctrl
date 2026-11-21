package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Howling Commandos — Legendary Creature — Human Soldier Hero {1}{W},
// 3/1 (Marvel Super Heroes Commander, #4455):
//
//	"{5}: Creatures you control get +1/+1 until end of turn. Soldiers you
//	 control gain vigilance until end of turn."
//
// Dwarven Provisioner's mass pump, plus a vigilance grant over the Soldiers.
// Both sets are locked as the ability resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49325445-0825-405e-be86-2a814699ce03",
		Name:         "The Howling Commandos",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: Creatures you control get +1/+1 until end of turn. Soldiers you control gain vigilance until end of turn.",
			Purpose: game.Purpose{TeamPump: &game.Pump{Power: 1, Toughness: 1}, Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1,
					Label: "The Howling Commandos — creatures you control get +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl(), effects.Subtype("Soldier")), Keywords: []string{"vigilance"},
					Label: "The Howling Commandos — Soldiers you control gain vigilance"}.Apply(ctx)
			},
		}},
	})
}
