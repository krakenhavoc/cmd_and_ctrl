package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Picard, Leading by Example — Legendary Creature — Human Officer {1}{W},
// 2/2 (Star Trek, #4973):
//
//	"{1}, {T}: Another target creature gets +X/+X until end of turn,
//	 where X is Picard's power. Activate only as a sorcery."
//
// X is read as the ability resolves (CR 608.2h): Picard's current power
// while he is on the battlefield, his last-known power once he has left
// (sourcePowerNow, Storm, Queen of Wakanda's reading). "Another" is
// object identity (Another), so Picard can't pump himself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b36f8e2-ba7d-4bc7-890f-dd6ab73eef17",
		Name:         "Picard, Leading by Example",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}, {T}: Another target creature gets +X/+X until end of turn, where X is Picard's power. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets:      effects.Another(effects.TargetCreature("another target creature")),
			SorcerySpeed: true,
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
			Effect:       picardPumpByHisPower,
		}},
	})
}

// picardPumpByHisPower gives the target +X/+X where X is Picard's power.
func picardPumpByHisPower(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	x := effects.SourcePowerNow(ctx)
	return effects.BoostUntilEOT{Target: id, Power: x, Toughness: x, Label: "Picard, Leading by Example — +X/+X"}.Apply(ctx)
}
