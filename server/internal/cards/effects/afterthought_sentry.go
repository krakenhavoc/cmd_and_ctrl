package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Afterthought Sentry — Artifact Creature — Gargoyle {2}, 2/2:
//
//	"{2}: This creature gains flying until end of turn.
//	 Whenever this creature attacks, exile up to one target card from a
//	 graveyard."
//
// The attack trigger's target is "up to one", so it goes on the stack
// with no target and then does nothing; with a card chosen, it is
// re-checked at resolution (CR 608.2b) and skipped if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f0c533a2-751a-4c5a-8cec-b83da3c8c62f",
		Name:         "Afterthought Sentry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Afterthought Sentry — exile up to one target card from a graveyard",
					b27ExileChosenTarget),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: This creature gains flying until end of turn",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"flying"},
					Label:    "Afterthought Sentry — flying until end of turn",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
