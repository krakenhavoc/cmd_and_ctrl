package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Atlas — Artifact {3}:
//
//	"{T}: Add one mana of any color.
//	 Corrupted — Whenever this artifact becomes tapped, each opponent who has
//	 three or more poison counters loses 1 life."
//
// "Becomes tapped" is any tap (ThisBecameTapped, #4534), the mana ability's
// included. The poison is read per opponent as the trigger resolves, so
// only the opponents who are corrupted then lose life; with several, each
// of them does. The trigger itself does not need the controller to be
// corrupted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b85d54e9-4f21-4440-9fd4-7b73b560c99d",
		Name:         "Phyrexian Atlas",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack}, effects.ThisBecameTapped,
				"Phyrexian Atlas — each opponent with three or more poison counters loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, opp := range ctx.Opponents() {
						if !effects.PlayerIsCorrupted(g, opp) {
							continue
						}
						if err := g.ChangePlayerLifeForEffect(ctx.Source(), opp, -1); err != nil {
							return err
						}
					}
					return nil
				}),
		},
	})
}
