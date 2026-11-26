package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bant Battlemage — Creature — Human Wizard {2}{W}, 2/2 (Shards of Alara,
// S82 OLD batch 510, #5588):
//
//	"{G}, {T}: Target creature gains trample until end of turn.
//	 {U}, {T}: Target creature gains flying until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "92d8b956-87db-43e4-bbb5-436ba01b5714",
		Name:         "Bant Battlemage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.S82OldB510TargetGainsRow(effects.Plus(effects.ManaCost("{G}"), effects.TapCost()), "{G}, {T}", "", "trample"),
			effects.S82OldB510TargetGainsRow(effects.Plus(effects.ManaCost("{U}"), effects.TapCost()), "{U}, {T}", "", "flying"),
		},
	})
}
