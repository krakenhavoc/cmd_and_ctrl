package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Displaced Dinosaurs — Creature — Dinosaur {5}{G}{G}, 7/7 (Doctor Who):
//
//	"As a historic permanent you control enters, it becomes a 7/7 Dinosaur
//	 creature in addition to its other types. (Artifacts, legendaries, and
//	 Sagas are historic.)"
//
// A replacement effect on the entry of each historic permanent under the
// controller's control (CR 614.1c): the permanent arrives already a 7/7
// Dinosaur creature in addition to its other types (CR 205.1b; layers 4 and
// 7b), so no enters trigger or state check sees it without that. It reaches
// artifacts, legendary permanents and Sagas, tokens that are historic
// included; the Dinosaurs themselves are not historic.
//
// No simplification.
func init() {
	const label = "Displaced Dinosaurs: a historic permanent you control enters as a 7/7 Dinosaur creature in addition to its other types"
	Register(Spec{
		OracleID:     "bc0daba2-c51c-44b0-a4a3-806f2acd9aa6",
		Name:         "Displaced Dinosaurs",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{{
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: gapB85HistoricPermanentYouControlEnters,
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.BecomeCreatureAtETB("Dinosaur", 7, 7, label)
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: label,
		}},
	})
}
