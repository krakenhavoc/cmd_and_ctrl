package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Auriok Windwalker — Creature — Human Wizard {3}{W}, 2/3 (Fifth Dawn,
// #5604):
//
//	"Flying
//	 {T}: Attach target Equipment you control to target creature you
//	 control."
//
// Brass Squire's ability on a flier: two target clauses read back with
// AttachClauseTargets, at instant speed, since nothing on the card
// limits it to sorcery speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8a06210-de30-4e87-835e-80a6df5716ca",
		Name:            "Auriok Windwalker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Attach target Equipment you control to target creature you control.",
			Cost:  effects.TapCost(),
			Targets: effects.TwoSlotAttachTargets(
				"target Equipment you control", []effects.CardPredicate{effects.YouControl()},
				"target creature you control", []effects.CardPredicate{effects.YouControl()},
			),
			Effect: effects.AttachClauseTargets,
		}},
	})
}
