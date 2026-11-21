package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hithlain Rope — Artifact {2} (Tales of Middle-earth Commander, #5948):
//
//	"This artifact can't be sacrificed.
//	 {1}, {T}: Search your library for a basic land card, put it onto the
//	 battlefield tapped, then shuffle. The player to your right gains
//	 control of this artifact.
//	 {2}, {T}: Draw a card. The player to your right gains control of this
//	 artifact."
//
// Bucknard's Everfull Purse's hand-off: "the player to your right" is the
// nearest player still in the game walking right from the ability's
// controller (game.NearestOpponentForEffect); the control change has no
// duration (CR 611.2a) and does nothing if the Rope has left the battlefield
// or no one is left to take it. The search (Evolving Wilds' basic-land fetch)
// comes first, the hand-off after it, so a search waits for its prompt (CR
// 608.2c). "Can't be sacrificed" is Alexios' static (no cost, edict or
// delayed sacrifice takes it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "76dfa5c2-1ea6-4f89-9903-26998636845e",
		Name:         "Hithlain Rope",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CantBeSacrificed()},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Search your library for a basic land card, put it onto the battlefield tapped, then shuffle. The player to your right gains control of this artifact.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1},
				Effect: Do(
					Instruction(func(ctx *Context) error { return fetchBasicTapped(ctx.Game, ctx.Item) }),
					Instruction(gapB115HandToTheRight("Hithlain Rope")),
				),
			},
			{
				Label:   "{2}, {T}: Draw a card. The player to your right gains control of this artifact.",
				Cost:    Plus(ManaCost("{2}"), TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Effect: Do(
					DrawCards{N: 1},
					Instruction(gapB115HandToTheRight("Hithlain Rope")),
				),
			},
		},
	})
}

// gapB115HandToTheRight is "the player to your right gains control of this
// artifact": the nearest player to the ability controller's right takes the
// source, indefinitely (Bucknard's Everfull Purse).
func gapB115HandToTheRight(name string) func(ctx *Context) error {
	return func(ctx *Context) error {
		right := ctx.Game.NearestOpponentForEffect(ctx.Controller(), game.AttackDirectionRight)
		if right == uuid.Nil || !onBattlefield(ctx.Game, ctx.Source()) {
			return nil
		}
		return GainControl{
			Target:     ctx.Source(),
			Controller: right,
			Duration:   game.IndefiniteDuration(),
			Label:      name + " — the player to your right gains control",
		}.Apply(ctx)
	}
}
