package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6498e3e9-1a7b-4b47-9e7c-cdbd8928c9c8",
		Name:         "Quilled Charger",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Saddle(2)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Quilled Charger — +1/+2 and menace until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 2, Label: "Quilled Charger — +1/+2 and menace until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: "Quilled Charger — +1/+2 and menace until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return nil
			}),
		},
	})
}
