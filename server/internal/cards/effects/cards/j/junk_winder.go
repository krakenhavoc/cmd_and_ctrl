package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Junk Winder — Creature — Serpent {5}{U}{U}, 5/6: "Affinity for tokens.
// Whenever a token you control enters, tap target nonland permanent an
// opponent controls. It doesn't untap during its controller's next untap
// step." Affinity is the shared self cost modifier counting tokens you
// control; the trigger fires for any token entering under your control,
// itself excluded only by being a nontoken card.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5b677e88-bcc9-4a55-9c89-8bbd0e2344dc",
		Name:         "Junk Winder",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for tokens", effects.IsTokenPredicate()),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
					return ok && effects.IsToken(c)
				}, "Junk Winder — tap target nonland permanent an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Junk Winder — doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				effects.TargetPermanent("target nonland permanent an opponent controls", effects.Nonland(), effects.OpponentControls())),
		},
	})
}
