package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "49325445-0825-405e-be86-2a814699ce03",
		Name:         "The Howling Commandos",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}: Creatures you control get +1/+1 until end of turn. Soldiers you control gain vigilance until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1,
					Label: "The Howling Commandos — creatures you control get +1/+1"}).Apply(ctx); err != nil {
					return err
				}
				return GrantKeywordUntilEOT{Match: And(Creature(), YouControl(), Subtype("Soldier")), Keywords: []string{"vigilance"},
					Label: "The Howling Commandos — Soldiers you control gain vigilance"}.Apply(ctx)
			},
		}},
	})
}
