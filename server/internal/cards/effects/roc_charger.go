package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Roc Charger — Creature — Bird {2}{W}, 1/3 (Ravnica: Clue Edition, #5929):
//
//	"Flying
//	 Whenever this creature attacks, target attacking creature without flying
//	 gains flying until end of turn."
//
// Gryffwing Cavalry's target clause and grant, without the payment. A target that stopped attacking is skipped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "67789c0c-3f05-4b0b-a56d-0757ac801d22",
		Name:            "Roc Charger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Roc Charger — target attacking creature without flying gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					target := firstLegalCardTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: "Roc Charger — flying until end of turn"}.Apply(ctx)
				}),
				TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying"))),
		},
	})
}
