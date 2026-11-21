package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lumen-Class Frigate — Artifact — Spacecraft {1}{W}, 3/5 (Edge of
// Eternities, #4461):
//
//	"Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 12+.)
//	 2+ | Other creatures you control get +1/+1.
//	 12+ | Flying, lifelink"
//
// The 2+ line is Angel of Invention's anthem (TribalAnthem over other
// creatures you control) behind Borg Scout Sphere's charge-counter gate,
// read live, so a Frigate that loses its counters stops pumping. The
// Frigate is not a creature until twelve, and "other" leaves it out
// after that. Station, the 12+ creature threshold and its keywords are
// the shared constructors (Galvanizing Sawship).
//
// No simplification.
func init() {
	anthem := effects.TribalAnthem(effects.TribeFilter{Others: true, YoursOnly: true}, 1, 1)
	anthem.ActiveWhen = effects.AtChargeCounters(2)
	static := append(effects.SpacecraftAt(12, 3, 5), effects.ThresholdKeywords(12, "flying", "lifelink"))
	effects.Register(effects.Spec{
		OracleID:     "3aa0cda4-fb2d-4fd4-84f6-482089a85547",
		Name:         "Lumen-Class Frigate",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static:       append(static, anthem),
	})
}
