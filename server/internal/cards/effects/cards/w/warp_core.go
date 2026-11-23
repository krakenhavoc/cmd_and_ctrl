package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6a2df97a-341d-4bc4-9f5c-661b6ca78fd0",
		Name:         "Warp Core",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Warp Core — you may search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:    item.Controller,
					Predicate: effects.IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Warp Core — a basic land card, revealed, to hand",
				}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Put two charge counters on target Spacecraft. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets:      effects.TargetPermanent("target Spacecraft", effects.HasSubtype("Spacecraft")),
			SorcerySpeed: true,
			Effect:       effects.TrekB28TwoChargeCountersOnTheTarget,
		}},
	})
}
