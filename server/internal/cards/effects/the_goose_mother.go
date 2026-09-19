package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Goose Mother — Legendary Creature — Bird Hydra {X}{G}{U}, 2/2
// (EDHREC rank 2928):
//
//	"Flying
//	 The Goose Mother enters with X +1/+1 counters on it.
//	 When The Goose Mother enters, create half X Food tokens, rounded
//	 up.
//	 Whenever The Goose Mother attacks, you may sacrifice a Food. If
//	 you do, draw a card."
//
// The Food commander. Flying rides PrintedKeywords. The X counters are
// the printed CR 614.1c entry clause and ride the CR 614 pipeline as
// one (XCounters, #1002). The FOOD still happens at resolution: "create
// half X Food tokens" is a printed enters TRIGGER, not an entry
// replacement, and a trigger's Build is not handed the stack item — so
// resolution stays the last moment X is readable for it, and it makes
// ceil(X/2) real Food tokens a beat early. The attack trigger is
// Springbloom Druid's shape: "you may" is the trigger prompt, the
// Food is chosen through the pick_target prompt when the trigger
// goes on the stack ("a Food you control" — a Food you control can
// never be an illegal pick), and at resolution the Food is
// sacrificed if it is still there and the draw follows only from
// the sacrifice. With no Food the trigger is removed with no prompt
// (CR 603.3d), which matches "you may sacrifice a Food" with
// nothing to sacrifice.
//
// Two declared simplifications, both weaker than printed:
//
//   - The Food is created as the spell resolves rather than by an
//     enters trigger you can respond to.
//   - The Food to sacrifice is chosen when the attack trigger goes
//     on the stack, not on resolution.
func init() {
	Register(Spec{
		OracleID:     "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48",
		Name:         "The Goose Mother",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The Food is created as the spell resolves rather than by an enters trigger you can respond to.",
			"You pick the Food to sacrifice when the attack trigger goes on the stack rather than on resolution.",
		},
		PrintedKeywords:            []string{"flying"},
		EntersWithCountersFromCast: []game.EntryCountersFromCast{XCounters(game.CounterPlusOne)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			return CreateToken{Controller: item.Controller, Template: FoodToken(), N: (x + 1) / 2}.Apply(ctx)
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "The Goose Mother — sacrifice a Food to draw a card?"},
			Targets:        TargetPermanent("a Food you control", And(HasSubtype("Food"), YouControl())),
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, b27GooseMotherAttackLabel, b27SacrificeChosenThenDraw)
			},
		}},
	})
}
