package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Seedpod Squire — Creature — Bird Scout {3}{W/U}, 3/3 (Bloomburrow,
// #4467):
//
//	"Flying
//	 Whenever this creature attacks, target creature you control without
//	 flying gets +1/+1 until end of turn."
//
// Mayhem Patrol's attack trigger with Pileated Provisioner's target: a
// creature you control without flying, so never the Squire itself. With
// no such creature the trigger is removed (CR 603.3d); a target that
// gained flying in response gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e2f57e48-ff5b-4a42-9b7c-1bd23b0d78ed",
		Name:            "Seedpod Squire",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Seedpod Squire — target creature you control without flying gets +1/+1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 1, Label: "Seedpod Squire — +1/+1"}.Apply(ctx)
						}
					}
					return nil
				}),
				TargetCreature("target creature you control without flying", YouControl(), WithoutKeyword("flying"))),
		},
	})
}
