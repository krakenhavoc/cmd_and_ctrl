package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tiger Shark, Abyssal Hunter — Legendary Creature — Shark Human Villain
// {2}{U/B}{U/B}, 2/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Tiger Shark enters or attacks, he connives.
//	 {4}{U/B}: Tiger Shark can't be blocked this turn."
//
// One trigger on two events (Avengers Quinjet's WhenThisEntersOrAttacks)
// with the shared sourceConnives body. The activated ability is Ant-Man's
// RestrictUntilEOT on the source itself, which a source that has left the
// battlefield by resolution does not get (a new object, CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "286be139-1ddf-4b04-86b2-cf17a06c262a",
		Name:         "Tiger Shark, Abyssal Hunter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Tiger Shark — he connives", Do(Instruction(sourceConnives))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{U/B}: Tiger Shark can't be blocked this turn.",
			Cost:    ManaCost("{4}{U/B}"),
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				return RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked,
					Label: "Tiger Shark — can't be blocked"}.Apply(NewContext(g, item))
			},
		}},
	})
}
