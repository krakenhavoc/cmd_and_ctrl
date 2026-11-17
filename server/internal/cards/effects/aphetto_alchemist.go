package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aphetto Alchemist — Creature — Human Wizard {1}{U}, 1/2 (Onslaught):
//
//	"{T}: Untap target artifact or creature.
//	 Morph {U}"
//
// Cat-Owl's untap on a tap ability. The {T} makes it summoning-sick like
// any creature's. Face down it is a 2/2 with no text (CR 708.2a), so the
// ability is not there until it is turned face up for {U}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "867aa0b7-b814-4d33-a677-f6368c8e93c0",
		Name:             "Aphetto Alchemist",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Morph("{U}")},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap target artifact or creature.",
			Cost:    TapCost(),
			Targets: TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
			Effect:  untapTheTarget,
		}},
	})
}
