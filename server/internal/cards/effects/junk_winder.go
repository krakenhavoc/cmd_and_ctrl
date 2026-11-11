package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Junk Winder — Creature — Serpent {5}{U}{U}, 5/6: "Affinity for tokens.
// Whenever a token you control enters, tap target nonland permanent an
// opponent controls. It doesn't untap during its controller's next untap
// step." Affinity is the shared self cost modifier counting tokens you
// control; the trigger fires for any token entering under your control,
// itself excluded only by being a nontoken card.
func init() {
	Register(Spec{
		OracleID:     "5b677e88-bcc9-4a55-9c89-8bbd0e2344dc",
		Name:         "Junk Winder",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for tokens", IsTokenPredicate()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && IsToken(c)
				}, "Junk Winder — tap target nonland permanent an opponent controls; it doesn't untap during its controller's next untap step",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						id, ok := b16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return TapAndFreeze{Targets: []uuid.UUID{id}, Label: "Junk Winder — doesn't untap during its controller's next untap step"}.Apply(ctx)
					}),
				TargetPermanent("target nonland permanent an opponent controls", Nonland(), OpponentControls())),
		},
	})
}
