package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alacrian Jaguar — Creature — Cat Mount {4}{G}:
//
//	"Vigilance
//	 Whenever this creature attacks while saddled, it gets +2/+2 until end
//	 of turn.
//	 Saddle 1"
//
// Saddle is Saddle(1) (CR 702.171, ADR 0071 amendment 2026-10-08); the attack
// trigger reads the saddled designation when the attack is declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "caeb4ec1-a8ba-4ac4-ad32-edbc671e5a39",
		Name:            "Alacrian Jaguar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.Saddle(1)},
		Triggered: []game.TriggeredAbility{
			effects.AttacksWhileSaddled("Alacrian Jaguar — +2/+2 until end of turn", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Toughness: 2, Label: "Alacrian Jaguar — +2/+2 until end of turn"}).Apply(ctx); err != nil {
					return err
				}
				return nil
			}),
		},
	})
}
