package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quilled Charger — Creature — Porcupine Mount {3}{R}:
//
//	"Whenever this creature attacks while saddled, it gets +1/+2 and gains menace until end of turn.
//	 Saddle 2"
//
// Saddle is Saddle(2) (CR 702.171, ADR 0071 amendment 2026-10-08); the attack
// trigger reads the saddled designation when the attack is declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6498e3e9-1a7b-4b47-9e7c-cdbd8928c9c8",
		Name:         "Quilled Charger",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Saddle(2)},
		Triggered: []game.TriggeredAbility{
			AttacksWhileSaddled("Quilled Charger — +1/+2 and menace until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 2, Label: "Quilled Charger — +1/+2 and menace until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: "Quilled Charger — +1/+2 and menace until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return nil
			}),
		},
	})
}
