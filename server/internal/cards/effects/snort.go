package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Snort — Sorcery {3}{R}:
//
//	"Each player may discard their hand and draw five cards. Then Snort
//	 deals 5 damage to each opponent who discarded their hand this way.
//	 Flashback {5}{R}"
//
// Each player is asked in turn order starting with the caster, and each
// finishes their own discard-and-draw before the next is asked. The
// damage is dealt once the last player has answered ("Then"), to each
// opponent of the caster who said yes. An opponent who chose to discard
// an empty hand still counts as having discarded it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "ab94985c-5a3e-42c7-be10-29c5b638346b",
		Name:             "Snort",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return snortAsk(ctx, tablePlayers(ctx), nil)
		},
	})
}

// snortAsk asks the first remaining player, then continues down the
// list; when it is empty the damage goes out to the opponents who said
// yes.
func snortAsk(ctx *Context, players []uuid.UUID, discarded []uuid.UUID) error {
	if len(players) == 0 {
		// One printed instruction: the 5 damage to each of them is one
		// damage instance (CR 615.8).
		return ctx.Game.DamageInstanceForEffect(func() error {
			for _, p := range discarded {
				if p == ctx.Controller() {
					continue
				}
				if err := (DealDamage{Source: ctx.Item.SourceCardID, Target: p, Amount: 5}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		})
	}
	player, rest := players[0], players[1:]
	return MayChoice{
		Player:   player,
		Question: "Snort — discard your hand and draw five cards?",
		Draws:    5,
		Then:     snortAnswer,
		Carry:    joinCarry(discarded, rest),
	}.Apply(ctx)
}

// What the answer hands the steps behind it: the player who said yes, the
// players still to ask, and who has discarded so far (#4163). The next player
// is asked from the last step, so only one answer is in flight.
const (
	snortPlayerRecord    = "snort/player"
	snortRestRecord      = "snort/rest"
	snortDiscardedRecord = "snort/discarded"
)

// snortWheel is one player's "discard your hand, then draw five cards" and
// the question to the next player: the draw waits for a prompt a discard asks
// (a replacement offer such as Library of Leng's; CR 608.2c, #4155), and the next question waits for the draw's.
// It asks again, so it is registered in init rather than a package-level
// initialiser (Go would see an initialisation cycle).
var snortWheel ThenSteps

func init() {
	snortWheel = RegisterThenSteps("then-steps/snort-wheel",
		Instruction(func(ctx *Context) error {
			ids := ctx.StepRecord(snortPlayerRecord).IDs
			if len(ids) == 0 {
				return nil
			}
			_, err := discardWholeHand(ctx.Game, ids[0])
			return err
		}),
		Instruction(func(ctx *Context) error {
			ids := ctx.StepRecord(snortPlayerRecord).IDs
			if len(ids) == 0 {
				return nil
			}
			return ctx.Game.DrawNForEffect(ids[0], 5)
		}),
		Instruction(func(ctx *Context) error {
			return snortAsk(ctx, ctx.StepRecord(snortRestRecord).IDs, ctx.StepRecord(snortDiscardedRecord).IDs)
		}),
	)
}

// snortAnswer is one player's answer to "discard your hand and draw five
// cards?": a yes runs the wheel for them, a no asks the next player. The
// carried IDs are the players who have discarded, a Nil, then the players
// still to ask. Registered in init because the no asks through snortAsk,
// which names it. Keyed (#4291).
var snortAnswer game.ConfirmThen

func init() {
	snortAnswer = MayAnswer("confirm/snort-wheel", func(c *Context, r game.ConfirmAnswer) error {
		discarded, rest := splitCarry(r)
		if !r.Accepted {
			return snortAsk(c, rest, discarded)
		}
		c.RecordStep(snortPlayerRecord, []uuid.UUID{r.Chooser})
		c.RecordStep(snortRestRecord, rest)
		c.RecordStep(snortDiscardedRecord, append(slices.Clone(discarded), r.Chooser))
		return snortWheel.Run(c)
	})
}
