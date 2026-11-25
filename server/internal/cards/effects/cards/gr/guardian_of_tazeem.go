package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guardian of Tazeem — Creature — Sphinx {3}{U}{U}, 4/5 (Battle for
// Zendikar):
//
//	"Flying
//	 Landfall — Whenever a land you control enters, tap target creature an
//	 opponent controls. If that land is an Island, that creature doesn't
//	 untap during its controller's next untap step."
//
// The Island check reads the entering land as the ability resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "06432957-decc-4ba2-95b1-cc9536a576d4",
		Name:            "Guardian of Tazeem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Landfall("Guardian of Tazeem — tap target creature an opponent controls; it doesn't untap next untap step if that land is an Island",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					island := effects.S82OldB309EnteredLandIs(g, item, "Island")
					legal := ctx.LegalTargets()
					if len(legal) == 0 {
						return nil
					}
					id := legal[0].ID
					if island {
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Guardian of Tazeem — it doesn't untap during its controller's next untap step"}.Apply(ctx)
					}
					return g.TapTargetForEffect(id)
				}), effects.Harmful(effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()))),
		},
	})
}
