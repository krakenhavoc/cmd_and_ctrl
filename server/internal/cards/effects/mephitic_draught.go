package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mephitic Draught — Artifact {1}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"When this artifact enters or is put into a graveyard from the
//	 battlefield, you draw a card and you lose 1 life."
//
// One ability with two events (WhenThisEntersOrDies). "Is put into a
// graveyard from the battlefield" is what dies means for any permanent
// (CR 700.4), so the artifact's own death triggers it, and it looks back
// in time (CR 603.10a), triggering from the graveyard. A bounce or an
// exile does not. The life is lost after the draw.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "05da6b06-fb5f-41e8-b0a2-d949646e36d9",
		Name:         "Mephitic Draught",
		Purpose:      game.Purpose{Draws: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Mephitic Draught — you draw a card and you lose 1 life",
				Do(DrawCards{N: 1}, youLoseLife(1))),
		},
	})
}
