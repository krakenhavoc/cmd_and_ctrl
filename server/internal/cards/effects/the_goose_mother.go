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
// The Food commander. Flying rides PrintedKeywords. The X counters
// go on as the spell resolves (Goldvein Hydra's posture) and, for
// the same reason, so does the Food: an ETB trigger cannot see the
// X announced for the spell, and resolution is the last moment it
// is readable — ceil(X/2) real Food tokens. The attack trigger is
// Springbloom Druid's shape: "you may" is the trigger prompt, the
// Food is chosen through the pick_target prompt when the trigger
// goes on the stack ("a Food you control" — a Food you control can
// never be an illegal pick), and at resolution the Food is
// sacrificed if it is still there and the draw follows only from
// the sacrifice. With no Food the trigger is removed with no prompt
// (CR 603.3d), which matches "you may sacrifice a Food" with
// nothing to sacrifice.
//
// Three declared simplifications, all weaker than printed:
//
//   - The X counters are placed as the spell resolves, a beat
//     before the card enters, so a "whenever you put counters on a
//     permanent" payoff does not see them.
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
			"The X +1/+1 counters are put on The Goose Mother as the spell resolves, a beat before it enters, so effects that watch you put counters on a permanent don't see them.",
			"The Food is created as the spell resolves rather than by an enters trigger you can respond to.",
			"You pick the Food to sacrifice when the attack trigger goes on the stack rather than on resolution.",
		},
		PrintedKeywords: []string{"flying"},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			if err := (AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: x}).Apply(ctx); err != nil {
				return err
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
