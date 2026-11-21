package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brass Squire — Artifact Creature — Myr, {3}, 1/3:
//
//	"{T}: Attach target Equipment you control to target creature you
//	 control."
//
// The activated-ability twin of Magnetic Theft's spell: two target
// clauses of DIFFERENT kinds — clause 0 "target Equipment you
// control", clause 1 "target creature you control" — read back
// positionally with AttachClauseTargets rather than as one clause
// widened to fit both (ADR 0065, #764). Nothing on the printed card
// restricts activation to sorcery speed, unlike the equip ability
// itself, so this is instant-speed and stays that way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a10c3a5-6724-4e5a-ae4f-b27dde12735a",
		Name:         "Brass Squire",
		Completeness: effects.CompletenessFull,
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
