package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nebelgast Intruder — Creature — Spirit {2}{U}, 2/1 (Innistrad: Midnight Hunt, S82):
//
//	"Flash
//	 Flying
//	 When this creature enters, up to one target creature an opponent controls gets -2/-0 until end of turn."
//
// Up to one target: with none chosen nothing happens; a target that is illegal at resolution is left alone (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c591977-f2af-4a6e-a0be-3dfb45a3598b",
		Name:            "Nebelgast Intruder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Nebelgast Intruder — up to one target creature an opponent controls gets -2/-0 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return BoostUntilEOT{Target: id, Power: -2, Label: "Nebelgast Intruder — -2/-0 until end of turn"}.Apply(ctx)
					}),
				UpToOneTargetCreature("up to one target creature an opponent controls", OpponentControls())),
		},
	})
}
