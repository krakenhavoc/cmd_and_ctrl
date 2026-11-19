package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Temporal Adept — Creature — Human Wizard {1}{U}{U}, 1/1 (Urza's
// Destiny, #5621):
//
//	"{U}{U}{U}, {T}: Return target permanent to its owner's hand."
//
// Boomerang on a creature: any permanent, any player's, goes to its
// OWNER's hand. {T} needs the Adept to have been under its controller's
// control since their most recent turn began (CR 302.6). A target gone
// by resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fd9b5462-6578-4bf6-8026-23eaf5af3eee",
		Name:         "Temporal Adept",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}{U}, {T}: Return target permanent to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    Plus(ManaCost("{U}{U}{U}"), TapCost()),
			Targets: TargetPermanent("target permanent"),
			Effect:  bounceTheTargetToHand,
		}},
	})
}
