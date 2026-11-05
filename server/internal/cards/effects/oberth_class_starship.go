package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oberth-Class Starship — Artifact — Spacecraft {1}{W}, 2/2 (Star Trek,
// #4973):
//
//	"When this Spacecraft enters, search your library for a basic Plains
//	 card, reveal it, put it into your hand, then shuffle.
//	 Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 6+.)
//	 6+ | Flying"
//
// Galvanizing Sawship's station shape (Station, SpacecraftAt,
// ThresholdKeywords) with Gloryheath Lynx's search: "a basic Plains
// card" admits a Snow-Covered Plains and not a nonbasic Plains
// (IsBasicLandWithSubtype reads the Basic supertype). The search is not
// a "may"; it can find nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fb5e53cf-e12e-469b-b9ea-9debc967e77e",
		Name:         "Oberth-Class Starship",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 2, 2),
			ThresholdKeywords(6, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Oberth-Class Starship — search for a basic Plains", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLandWithSubtype("plains"),
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Reason:    "Oberth-Class Starship — a basic Plains card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
