package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aven Courier — Creature — Bird Advisor {1}{U}, 1/1 (New Capenna
// Commander, #5491):
//
//	"Flying
//	 Whenever this creature attacks, choose a counter on a permanent you
//	 control. Put a counter of that kind on target permanent you control if
//	 it doesn't have a counter of that kind on it."
//
// The target is chosen as the trigger goes on the stack; the kind is chosen
// as it resolves, from the kinds of counter on the permanents the controller
// has then (the target's own counters included: choosing a kind it already
// has does nothing, as printed). One kind is taken without asking, and with
// no counter anywhere nothing happens. The target is checked again when the
// counter is put, so a counter the target gained in between still blocks it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a01746f3-8e6f-47ce-a95b-e44556d1f73e",
		Name:            "Aven Courier",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Aven Courier — choose a counter on a permanent you control and put one of that kind on target permanent you control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					kinds := effects.SncB07KindsAmongPermanentsYouControl(g, item.Controller)
					switch len(kinds) {
					case 0:
						return nil
					case 1:
						return avenCourierPut(ctx, target, kinds[0])
					}
					options := make([]game.ChoiceOption, len(kinds))
					for i, kind := range kinds {
						options[i] = game.ChoiceOption{Label: "A " + kind + " counter"}
					}
					return effects.PickOption{
						Question: "Aven Courier — choose a counter on a permanent you control",
						Options:  options,
						ThenKey:  avenCourierPicked,
						Carry:    []uuid.UUID{target},
						Words:    kinds,
					}.Apply(ctx)
				}), effects.TargetPermanent("target permanent you control", effects.YouControl())),
		},
	})
}

// avenCourierPut puts a counter of the kind on the target if it has none of
// that kind.
func avenCourierPut(ctx *effects.Context, target uuid.UUID, kind string) error {
	c, ok := ctx.Game.LookupCardForEffect(target)
	if !ok || !effects.OnBattlefield(ctx.Game, target) || c.Counters[kind] > 0 {
		return nil
	}
	return ctx.Game.AddCounterByForEffect(ctx.Controller(), target, kind, 1)
}

// avenCourierPicked is the chosen kind, keyed. An on-disk key.
var avenCourierPicked = effects.OptionPickThen("option-pick/aven-courier-counter", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(r.Words) {
		return nil
	}
	return avenCourierPut(ctx, r.Carry[0], r.Words[r.Index])
})
