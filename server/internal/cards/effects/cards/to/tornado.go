package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tornado — Enchantment {4}{G} (Alliances, S82 OLD batch 818):
//
//	"Cumulative upkeep {G}
//	 {2}{G}, Pay 3 life for each velocity counter on this enchantment:
//	 Destroy target permanent and put a velocity counter on this
//	 enchantment. Activate only once each turn."
//
// The life is counted once, as the ability is announced (CR 601.2f, 602.2b),
// from the counters the enchantment has then.
func init() {
	const label = "{2}{G}, Pay 3 life for each velocity counter on this enchantment: Destroy target permanent and put a velocity counter on this enchantment. Activate only once each turn."
	effects.Register(effects.Spec{
		OracleID:     "dd283473-6423-4a1f-b5a3-77234cdce643",
		Name:         "Tornado",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Tornado — cumulative upkeep {G}", "{G}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerRemove},
			Cost:      effects.Plus(effects.ManaCost("{2}{G}"), effects.PayLifeCount(effects.OldB818ThreeLifePerVelocityCounter)),
			Condition: effects.OncePerTurnActivation(label),
			Targets:   effects.TargetPermanent("target permanent"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
					return err
				}
				if !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.AddCounter{Target: item.SourceCardID, Kind: "velocity", N: 1}.Apply(ctx)
			},
		}},
	})
}
