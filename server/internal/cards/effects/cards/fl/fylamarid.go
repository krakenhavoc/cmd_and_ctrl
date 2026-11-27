package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fylamarid — Creature — Squid Beast {1}{U}{U}, 1/3 (Tempest, S82 OLD b754,
// #5628):
//
//	"Flying
//	 This creature can't be blocked by blue creatures.
//	 {U}: Target creature becomes blue until end of turn."
//
// Mudbrawler Raiders' block rule (a blue blocker, as it is now, is refused,
// CR 509.1b) and Grand Architect's layer-5 colour change (CR 613.1e), which
// replaces the creature's colours with blue.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c5829bff-9687-42b2-8bba-1f036e592cdd",
		Name:            "Fylamarid",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("U")), "blue creatures"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}: Target creature becomes blue until end of turn.",
			Cost:    effects.ManaCost("{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Gap2B05TargetBecomesColor(effects.NewContext(g, item), []string{"U"}, "Fylamarid — blue until end of turn")
			},
		}},
	})
}
