package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fatestitcher — Creature — Zombie Wizard {3}{U}, 1/2 (Shards of Alara,
// #5588):
//
//	"{T}: You may tap or untap another target permanent.
//	 Unearth {U}"
//
// Trickster Mage's "you may tap or untap" body on Kiora's Follower's
// "another target permanent" clause: the controller picks tap it, untap
// it or leave it as the ability resolves, and a target gone in response
// is illegal, so nothing is asked (CR 608.2b). "Another" is object
// identity (Another), so the Fatestitcher can't pick itself. Unearth is
// the shared constructor (CR 702.84a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55765934-07be-4f03-a2e8-8292c0b601ae",
		Name:         "Fatestitcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: You may tap or untap another target permanent.",
				Cost:    effects.TapCost(),
				Targets: effects.Another(effects.TargetPermanent("another target permanent")),
				Effect:  effects.NemMayTapOrUntapTheTarget("Fatestitcher"),
			},
			effects.Unearth("{U}"),
		},
	})
}
