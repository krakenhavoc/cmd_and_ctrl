package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bowie Base One — Plane — Mars (Doctor Who, #4870):
//
//	"At the beginning of your end step, goad target creature controlled
//	 by the player to your left. (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you if
//	 able.)
//	 Whenever chaos ensues, target creature gains islandwalk until end of
//	 turn. (It can't be blocked as long as defending player controls an
//	 Island.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// "The player to your left" is the next player in the game walking left
// round the table from the planar controller's seat (the seat order turns
// pass in), read as the trigger is put on the stack and again as it
// resolves (CR 608.2b). Goad is CR 701.15; islandwalk is CR 702.14.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7076bc0c-9cc7-4cd7-aba5-f2f236aab0f7",
		Name:         "Bowie Base One",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Bowie Base One — goad target creature controlled by the player to your left", GoadTarget),
				TargetCreature("target creature controlled by the player to your left", ControlledByThePlayerToYourLeft())),
			Targeting(WheneverChaosEnsues("Bowie Base One — target creature gains islandwalk until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t, Keywords: []string{"islandwalk"}, Label: "Bowie Base One — islandwalk until end of turn"}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}

// ControlledByThePlayerToYourLeft is "controlled by the player to your
// left": the first player still in the game walking left round the table
// from the caster's seat (game.NearestOpponentForEffect with
// AttackDirectionLeft; every other player is an opponent in Commander).
func ControlledByThePlayerToYourLeft() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		left := g.NearestOpponentForEffect(caster, game.AttackDirectionLeft)
		return left != uuid.Nil && c.Controller == left
	}
}
