package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warp Core — {2} Artifact (Star Trek, #4973):
//
//	"When this artifact enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle.
//	 {1}, {T}: Put two charge counters on target Spacecraft. Activate
//	 only as a sorcery."
//
// The enters trigger is Environmental Scientist's optional basic-land
// search to hand. The activation targets any Spacecraft, whoever
// controls it and whether or not it is a creature yet; the counters are
// placed by the ability, not by station, so a "whenever charge counters
// are put" reader sees them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a2df97a-341d-4bc4-9f5c-661b6ca78fd0",
		Name:         "Warp Core",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Warp Core — you may search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Warp Core — a basic land card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{1}, {T}: Put two charge counters on target Spacecraft. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{1}"), TapCost()),
			Targets:      TargetPermanent("target Spacecraft", HasSubtype("Spacecraft")),
			SorcerySpeed: true,
			Effect:       trekB28TwoChargeCountersOnTheTarget,
		}},
	})
}
