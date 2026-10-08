package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venomsac Lagac — Creature — Lizard Mount {1}{G}:
//
//	"Deathtouch
//	 Whenever this creature attacks while saddled, it gets +0/+3 until end
//	 of turn.
//	 Saddle 2"
//
// Saddle is Saddle(2) (CR 702.171, ADR 0071 amendment 2026-10-08); the attack
// trigger reads the saddled designation when the attack is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2fac62be-dcbe-409f-a362-f0e519ec14a9",
		Name:            "Venomsac Lagac",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Activated:       []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Venomsac Lagac — +0/+3 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 0, Toughness: 3, Label: "Venomsac Lagac — +0/+3 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return nil
			}),
		},
	})
}
