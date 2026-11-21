package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wild Dogs — Creature — Dog {G}, 2/1 (Dominaria Remastered, #5948):
//
//	"At the beginning of your upkeep, if a player has more life than each
//	 other player, the player with the most life gains control of this
//	 creature.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// An intervening "if" (CR 603.4), checked as the upkeep begins and again as
// the trigger resolves. A tie for the most life is nobody having more life
// than each other player, so nothing happens. The control change has no
// duration (CR 611.2a), so it lasts until someone else takes the Dogs; the
// new controller will see the trigger on their own upkeeps, and the effect
// is pinned to the Dogs as the object they are now (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "711cda97-bf57-4b22-b9c1-854ef448d27a",
		Name:         "Wild Dogs",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Triggered: []game.TriggeredAbility{
			InterveningIf(gapB137LifeLeaderCondition(), On(game.EventBeginUpkeep,
				AllOf(ByYou, gapB137ALifeLeaderExists),
				"Wild Dogs — the player with the most life gains control of this creature",
				Do(Instruction(func(ctx *Context) error {
					leader, ok := gapB137LifeLeader(ctx.Game)
					if !ok {
						return nil
					}
					return GainControl{
						Target:     ctx.Source(),
						Controller: leader,
						Duration:   game.IndefiniteDuration(),
						Label:      "Wild Dogs — gain control",
					}.Apply(ctx)
				})))),
		},
	})
}
