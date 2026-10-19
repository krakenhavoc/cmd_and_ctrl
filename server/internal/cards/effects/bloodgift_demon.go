package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodgift Demon — Creature — Demon 5/4 for {3}{B}{B}:
//
//	"Flying
//	 At the beginning of your upkeep, target player draws a card and
//	 loses 1 life."
//
// The politics half of the life-for-cards family: the trigger names
// "target player", not "you", so the draw can be handed to whoever
// the controller wants something from — and the 1 life comes off
// whoever draws, not off the Demon's controller. Pointing it at an
// opponent is a real line, which is why the target clause is
// declared rather than hard-wired to the controller.
//
// "Draws a card and loses 1 life" is one instruction and both halves
// hit the same player, so a target that has left the game between
// the trigger going on the stack and resolving does neither
// (CR 608.2b — the engine re-checks the ref and the effect reads it
// back off the item).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "e64b184d-9746-4723-a928-d459b5c3ee6c",
		Name:            "Bloodgift Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Targets: TargetPlayer("target player"),
			Purpose: ForTargets(game.TargetPurpose{Slot: 0, Draws: 1, LifeLoss: 1}),
			Key:     "Bloodgift Demon — target player draws a card and loses 1 life",
			Effect: Do(
				// The life loss waits for a prompt the draw asks (CR 608.2c, #4155).
				Instruction(func(ctx *Context) error {
					if t := ctx.Item.Targets; len(t) > 0 && t[0].Kind == game.TargetPlayer {
						return DrawCards{Player: t[0].ID, N: 1}.Apply(ctx)
					}
					return nil
				}),
				Instruction(func(ctx *Context) error {
					if t := ctx.Item.Targets; len(t) > 0 && t[0].Kind == game.TargetPlayer {
						return GainLife{Player: t[0].ID, Amount: -1}.Apply(ctx)
					}
					return nil
				}),
			),
		}},
	})
}
