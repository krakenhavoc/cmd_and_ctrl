package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stormshriek Feral // Flush Out — Tarkir: Dragonstorm (#4463), an omen
// card (CR 720, #5175):
//
//	Stormshriek Feral — Creature — Dragon {4}{R}, 3/3
//	  "Flying, haste
//	   {1}{R}: This creature gets +1/+0 until end of turn."
//	Flush Out — Sorcery — Omen {1}{R}
//	  "Discard a card. If you do, draw two cards. (Then shuffle this card
//	   into its owner's library.)"
//
// The Dragon under the bare oracle ID, the Omen under "<oracle>#1"; the
// shuffle is the engine's (CR 720.3d, game/omen.go). Flush Out's discard
// is chosen as it resolves, and with an empty hand nothing is discarded
// and nothing drawn (Test of Identity's second mode).
//
// No simplification.

const stormshriekFeralOracleID = "e24916ff-7566-4f27-8e31-732311a4a88c"

func init() {
	Register(Spec{
		OracleID:        stormshriekFeralOracleID,
		Name:            "Stormshriek Feral",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}: This creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{R}"),
			Effect:  thisGetsUntilEndOfTurn(1, 0, "Stormshriek Feral — +1/+0"),
		}},
	})
	Register(Spec{
		OracleID:     stormshriekFeralOracleID + "#1",
		Name:         "Flush Out",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2, Discards: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player: ctx.Controller(),
				Source: ctx.Source(),
				N:      1,
			}, drawTwoIfDiscarded.then, nil)
		},
	})
}
