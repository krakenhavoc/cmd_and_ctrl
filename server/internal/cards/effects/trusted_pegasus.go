package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trusted Pegasus — Creature — Pegasus {2}{W}, 2/2 (Ravnica: Clue Edition,
// S82):
//
//	"Flying
//	 Whenever this creature attacks, target attacking creature without flying
//	 gains flying until end of turn."
//
// The target is chosen as the trigger goes on the stack; a creature that has
// stopped attacking or gained flying by resolution is no longer legal (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18729459-8669-4c56-a78d-fd5982c3b2ac",
		Name:            "Trusted Pegasus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WheneverThisAttacks("Trusted Pegasus — target attacking creature without flying gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, id := range legalTargetCardIDs(ctx) {
						if err := (GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"}, Label: "Trusted Pegasus — flying"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
				TargetCreature("target attacking creature without flying", AttackingCreature(), WithoutKeyword("flying"))),
				ForTargets(BenefitUntilEOT(0, 0, 0, "flying"))),
		},
	})
}
