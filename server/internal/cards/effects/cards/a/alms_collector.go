package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Alms Collector — Creature — Cat Cleric {3}{W}, 3/4 (S82 GAP2, #5948,
// #6089):
//
//	"Flash
//	 If an opponent would draw two or more cards, instead you and that
//	 player each draw a card."
//
// A replacement of the whole draw instruction (CR 121.2, 614.1a), the
// engine's game.ReplacementEffect.DrawInstructionEachDrawOne
// (draw_instruction_rewrite.go): an opponent's "draw two" (or three, or
// seven) becomes one draw for them and one for Alms Collector's controller,
// each its own draw that other replacements still see, made in APNAP order
// (CR 101.4). A "draw a card" is untouched, and so is each of a draw step's
// single draws. The instruction's "then" ("draw two, then discard a card")
// still follows the drawing player's one card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "278e51fe-3cd9-4c82-bd81-2164771f8611",
		Name:            "Alms Collector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Replacements: []game.ReplacementEffect{{
			Watches:                    []game.EventKind{game.EventDrawCard},
			DrawInstructionEachDrawOne: true,
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				// A Commander table has no teams: every other player is an opponent.
				return ev.Kind == game.RepEventDraw && ev.DrawInstruction && ev.DrawCount >= 2 &&
					src != nil && ev.DrawPlayer != src.Controller
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Alms Collector: instead you and that player each draw a card",
		}},
	})
}
