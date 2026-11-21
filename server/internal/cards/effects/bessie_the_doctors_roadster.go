package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bessie, the Doctor's Roadster — Legendary Artifact — Vehicle {2}, 3/4
// (Doctor Who, #5948):
//
//	"Haste
//	 Whenever Bessie attacks, another target legendary creature can't be
//	 blocked this turn.
//	 Crew 2"
//
// Crew 2 is the ordinary CR 702.122 crew ability (Adventurer's Airship). The
// attack trigger targets another legendary creature (any controller's, as
// printed) and gives it CantBeBlocked until end of turn (Access Tunnel's
// body); a target that left is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "776aa036-08aa-4fa0-82bd-c09e6e3e2c5e",
		Name:            "Bessie, the Doctor's Roadster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Bessie — another target legendary creature can't be blocked this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBeBlocked, Label: "Bessie — can't be blocked"}.Apply(ctx)
						}
						return nil
					}),
				Another(TargetCreature("another target legendary creature", Legendary()))),
		},
		Activated: []ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Bessie, the Doctor's Roadster"),
		}},
	})
}
