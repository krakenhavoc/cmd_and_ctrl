package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// D'Vana Tendi, Optimistic Ensign — Legendary Creature — Orion Scientist
// {1}{G/U}{G/U}, 2/4 (Star Trek, #4973):
//
//	"{T}: Untap another target permanent.
//	 {3}: Surveil 1."
//
// Ioreth of the Healing House's first ability, and a plain surveil sink.
// The untap is a CR 602 activation with a tap cost, so summoning
// sickness applies (CR 302.6); "another" keeps Tendi out of its own
// targets. The surveil needs no tap and can be activated any number of
// times.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bb42689-b169-43cb-b787-f0e348063cd6",
		Name:         "D'Vana Tendi, Optimistic Ensign",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent.",
				Cost:    TapCost(),
				Targets: Another(TargetPermanent("another target permanent")),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:   "{3}: Surveil 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{3}"),
				Effect:  Do(Surveil{N: 1}),
			},
		},
	})
}
