package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Alloy Animist — Creature — Human Druid {G}, 1/1 (The Brothers' War,
// #5482):
//
//	"{2}{G}: Until end of turn, target noncreature artifact you control
//	 becomes a 4/4 artifact creature."
//
// "Becomes an artifact creature" sets the card types (CR 205.1a): an
// artifact land loses Land, an artifact enchantment loses Enchantment
// (s82BroB11BecomesArtifactCreature). The base 4/4 is layer 7b, so
// counters and pumps still add. The target is checked as the ability
// resolves (CR 608.2b), and it must still be a noncreature artifact you
// control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4acacdd2-a736-43a2-ad67-891b9bd124f8",
		Name:         "Alloy Animist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}: Until end of turn, target noncreature artifact you control becomes a 4/4 artifact creature.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{2}{G}"),
			Targets: TargetPermanent("target noncreature artifact you control", And(Artifact(), Not(Creature()), YouControl())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
					return nil
				}
				return s82BroB11BecomesArtifactCreature(ctx, t.ID, 4, 4, DurationUntilEndOfTurn(ctx),
					"Alloy Animist — a 4/4 artifact creature until end of turn")
			},
		}},
	})
}
