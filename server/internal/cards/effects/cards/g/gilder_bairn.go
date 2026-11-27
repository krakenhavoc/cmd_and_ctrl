package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilder Bairn — {1}{G/U}{G/U} Creature — Ouphe, 1/3:
//
//	"{2}{G/U}, {Q}: Double the number of each kind of counter on target
//	 permanent. ({Q} is the untap symbol.)"
func init() {
	cost := effects.ManaCost("{2}{G/U}")
	cost.UntapSelf = true
	effects.Register(effects.Spec{
		OracleID:     "a4441c7d-e22a-4e19-8365-34d52da9fe86",
		Name:         "Gilder Bairn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G/U}, {Q}: Double the number of each kind of counter on target permanent.",
			Cost:    cost,
			Targets: effects.TargetPermanent("target permanent"),
			Effect:  effects.S82OldB522DoubleEachKindOfCounter,
		}},
	})
}
