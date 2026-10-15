package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aesi, Tyrant of Gyre Strait — Legendary Creature — Serpent {4}{G}{U},
// 5/5 (EDHREC top 1,000, #3654):
//
//	"You may play an additional land on each of your turns.
//	 Landfall — Whenever a land you control enters, you may draw a card."
//
// Exploration's extra land drop (Spec.AdditionalLandPlays, which the bot
// reads through Purpose.ExtraLandDrops) plus Tatyova's landfall draw as
// a "you may" (CR 603.5): the prompt is asked before the trigger goes on
// the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:            "6511f317-bd38-46d0-b800-7125a3f420da",
		Name:                "Aesi, Tyrant of Gyre Strait",
		Completeness:        CompletenessFull,
		AdditionalLandPlays: 1,
		Purpose:             game.Purpose{ExtraLandDrops: 1},
		Triggered: []game.TriggeredAbility{
			Optional(
				LandfallPaying("Aesi — you may draw a card (landfall)", game.LandfallPayoff{Draws: 1},
					Do(DrawCards{N: 1})),
				"Aesi, Tyrant of Gyre Strait — draw a card?"),
		},
	})
}
