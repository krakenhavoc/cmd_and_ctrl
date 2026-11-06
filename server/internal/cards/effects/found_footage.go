package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Found Footage — Artifact — Clue {1} (Duskmourn: House of Horror,
// #4466):
//
//	"You may look at face-down creatures your opponents control any time.
//	 {2}, Sacrifice this artifact: Surveil 2, then draw a card. (To
//	 surveil 2, look at the top two cards of your library, then put any
//	 number of them into your graveyard and the rest on top of your
//	 library in any order.)"
//
// The first line is LooksAtFaceDownCreatures (game/face_down_look.go,
// #5251): while Found Footage is on the battlefield with its abilities,
// its controller's view shows the card under every face-down creature
// another player controls, and nobody else's view changes (a spectator,
// another opponent and the replay keep seeing a face-down 2/2). It is
// read on every frame and never stored, so the look ends the moment the
// Clue is sacrificed or leaves. Every other player is an opponent in a
// free-for-all (CR 806.1).
//
// The sacrifice is Test of Priorities' "Surveil 2, then draw a card":
// the draw rides the surveil prompt's keyed continuation, so the card
// drawn comes from the library the player arranged (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                 "ad37551c-9d20-4902-8190-6a0aa32a7947",
		Name:                     "Found Footage",
		Completeness:             CompletenessFull,
		LooksAtFaceDownCreatures: true,
		Activated: []ActivatedAbility{{
			Label:   "{2}, Sacrifice this artifact: Surveil 2, then draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
			Effect: Do(Instruction(func(ctx *Context) error {
				return Surveil{Player: ctx.Controller(), N: 2, ThenTail: ctx.Tail(thenDrawACard)}.Apply(ctx)
			})),
		}},
	})
}
