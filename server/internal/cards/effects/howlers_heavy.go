package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Howler's Heavy — Creature — Seal Pirate {3}{U}, 3/4 (Aetherdrift,
// #4464):
//
//	"Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, target creature or Vehicle an opponent
//	 controls gets -3/-0 until end of turn."
//
// Cycling is the hand ability (CR 702.29a), and the cycle trigger fires
// from the graveyard, where the card already is (CR 702.29c), Agonasaur
// Rex's InGraveyard(On(EventCycle, Self, …)). Its target is a creature or
// a Vehicle, crewed or not, that an opponent controls; an uncrewed
// Vehicle keeps the -3/-0 for the turn and feels it once it becomes a
// creature (CR 301.7a). With nothing to target the trigger is removed
// (CR 603.3d); a target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8abeaa77-d455-47e7-bf60-aefb94ab83fb",
		Name:         "Howler's Heavy",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{1}{U}")},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				InGraveyard(On(game.EventCycle, Self,
					"Howler's Heavy — cycled: target creature or Vehicle an opponent controls gets -3/-0 until end of turn",
					boostTargetUntilEOT("Howler's Heavy — -3/-0", -3, 0))),
				TargetPermanent("target creature or Vehicle an opponent controls",
					Or(Creature(), Subtype("Vehicle")), OpponentControls())),
				ForTargets(HarmUntilEOT(0, -3, 0))),
		},
	})
}
