package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starfleet Security — Creature — Human Officer {R}, 1/1 (Star Trek,
// #4973):
//
//	"Haste
//	 When this creature dies during your turn, investigate. (Create a
//	 Clue token. It's an artifact with "{2}, Sacrifice this token: Draw
//	 a card.")"
//
// Haste is the printed keyword. "During your turn" is part of the
// trigger event, not an intervening "if": the ability triggers only when
// the creature dies while the player who controlled it as it last
// existed on the battlefield (CR 603.10a, leftUnderControlOf) is the
// active player, and nothing is checked again as it resolves.
// Investigate is the printed Clue token (CR 701.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "896077f9-b26e-4964-9279-03ee284a3e89",
		Name:            "Starfleet Security",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, trekB18DiedDuringYourTurn, "Starfleet Security — investigate",
				Do(Investigate{N: 1})),
		},
	})
}
