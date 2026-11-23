package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draugr Recruiter — Creature — Zombie Cleric {3}{B}, 3/3:
//
//	"Boast — {3}{B}: Return target creature card from your graveyard to your hand. (Activate only if
//	 this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f854ea0d-aae3-4b6a-9767-6ac1788f2190",
		Name:         "Draugr Recruiter",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastTargeting("{3}{B}: Return target creature card from your graveyard to your hand.",
				effects.ManaCost("{3}{B}"),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
				effects.ReturnTargetCardToHand),
		},
	})
}
