package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Niblis of Frost — Creature — Spirit {2}{U}{U}, 3/3 (Commander 2020, S82 OLD,
// #5513):
//
//	"Flying
//	 Prowess
//	 Whenever you cast an instant or sorcery spell, tap target creature an
//	 opponent controls. That creature doesn't untap during its controller's
//	 next untap step."
//
// Flying and prowess ride PrintedKeywords. The trigger is White Dragon's tap
// and freeze (TapAndFreeze) behind Murmuring Mystic's cast condition: one
// trigger per instant or sorcery spell you cast (a copy is not cast, CR
// 707.10), the target chosen as it goes on the stack and judged again as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ed2f3339-5c3e-473f-b08e-58288c4ad9c2",
		Name:            "Niblis of Frost",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "prowess"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.InstantOrSorceryCastByYou(ev, source, g)
			}, "Niblis of Frost — tap target creature an opponent controls; it doesn't untap during its controller's next untap step",
				func(g *game.Game, item *game.StackItem) error {
					return effects.B751TapFreezeTargets(effects.NewContext(g, item), uuid.Nil, "Niblis of Frost")
				}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
