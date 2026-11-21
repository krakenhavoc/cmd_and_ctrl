package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Celestial Regulator — Creature — Angel Advisor {1}{W}{U}, 2/3 (Streets
// of New Capenna):
//
//	"Flying
//	 When this creature enters, choose target creature you don't control
//	 and tap it. If you control a creature with a counter on it, the
//	 chosen creature doesn't untap during its controller's next untap
//	 step."
//
// The tap is unconditional. The freeze is the ADR 0058 one-shot mark, put
// on only if, as the ability resolves, its controller controls a creature
// with a counter of any kind on it (the card's own counters included).
//
// No simplification.
func init() {
	const label = "Celestial Regulator — tap target creature you don't control; it doesn't untap during its controller's next untap step if you control a creature with a counter on it"
	effects.Register(effects.Spec{
		OracleID:        "3ee68758-9431-4446-8539-f42f545f3bd0",
		Name:            "Celestial Regulator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					if len(effects.CreaturesWithACounter(g, item.Controller)) == 0 {
						return g.TapTargetForEffect(id)
					}
					return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Celestial Regulator — doesn't untap during its controller's next untap step"}.Apply(ctx)
				}),
				effects.TargetCreature("target creature you don't control", effects.OpponentControls())),
		},
	})
}
