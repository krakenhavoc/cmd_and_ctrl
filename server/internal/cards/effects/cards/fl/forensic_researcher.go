package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Forensic Researcher — Creature — Merfolk Detective {2}{U}, 1/3 (Murders
// at Karlov Manor, #5193):
//
//	"{T}: Untap another target permanent you control.
//	 {T}, Collect evidence 3: Tap target creature you don't control. (To
//	 collect evidence 3, exile cards with total mana value 3 or greater from
//	 your graveyard.)"
//
// The first ability is Kelpie Guide's untapper. The second's cost is
// CollectEvidenceN: the activator names graveyard cards with total mana
// value 3 or more as they activate, and with less than that in the
// graveyard the ability can't be activated (CR 701.59b). Both are {T}
// abilities of a creature, so summoning sickness applies (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75e7f7c7-fa98-497b-bf5e-8f74571e190c",
		Name:         "Forensic Researcher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent you control.",
				Cost:    effects.TapCost(),
				Targets: effects.Another(effects.TargetPermanent("another target permanent you control", effects.YouControl())),
				Effect:  effects.B22UntapEachLegalTarget,
			},
			{
				Label:   "{T}, Collect evidence 3: Tap target creature you don't control.",
				Cost:    effects.Plus(effects.TapCost(), effects.CollectEvidenceN(3)),
				Targets: effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
				Effect:  effects.B36TapChosenCreature,
			},
		},
	})
}
