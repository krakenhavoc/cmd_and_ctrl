package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vorel of the Hull Clade — Legendary Creature — Human Merfolk
// {1}{G}{U}, 1/4 (EDHREC rank 3223):
//
//	"{G}{U}, {T}: Double the number of each kind of counter on target
//	 artifact, creature, or land."
//
// The counters commander. Every kind of counter on the target is
// read once, before the first placement, and each kind is then
// added again through AddCounter — so a Hardened Scales or Doubling
// Season applies to the doubling, as printed, and a doubled kind is
// never re-read mid-loop. A creature source, so the tap waits out
// summoning sickness (CR 302.6). The target is any artifact,
// creature or land, the controller's or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3001f971-c4db-4176-89b5-e7f4a8890c0f",
		Name:         "Vorel of the Hull Clade",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{U}, {T}: Double the number of each kind of counter on target artifact, creature, or land",
			Cost:    effects.Plus(effects.ManaCost("{G}{U}"), effects.TapCost()),
			Targets: effects.TargetPermanent("target artifact, creature, or land", effects.Or(effects.Artifact(), effects.Creature(), effects.Land())),
			Effect:  effects.B30DoubleCountersOnFirstLegalTarget,
		}},
	})
}
