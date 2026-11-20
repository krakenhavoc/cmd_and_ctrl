package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bone Pit Brute — Creature — Cyclops {4}{R}{R}, 4/5 (Core Set 2021, #5511):
//
//	"Menace
//	 When this creature enters, target creature gets +4/+0 until end of
//	 turn."
//
// Guardian of Pilgrims' targeted enters pump at +4/+0; a target that left
// in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6de38910-8d0d-4fac-b9db-c80ffd1850ea",
		Name:            "Bone Pit Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Bone Pit Brute — target creature gets +4/+0 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{Target: id, Power: 4, Label: "Bone Pit Brute — +4/+0"}.Apply(ctx)
				}), TargetCreature("target creature")),
		},
	})
}
