package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Animation Module — Artifact {1} (Kaladesh, #5539):
//
//	"Whenever one or more +1/+1 counters are put on a permanent you control,
//	 you may pay {1}. If you do, create a 1/1 colorless Servo artifact creature
//	 token.
//	 {3}, {T}: Choose a counter on target permanent or player. Give that
//	 permanent or player another counter of that kind."
//
// The trigger is Berta's "one or more +1/+1 counters are put on" reading: one
// counter-placement event that raised a permanent's +1/+1 total (a removal does
// not count and a batch is one trigger), here on any permanent whose controller
// is the Module's, with Myrsmith's "you may pay {1}. If you do" (MayPay). The
// activated ability is Dathon's choose-a-kind-of-counter over a target
// permanent or player (Skyship Plunderer's target clause): its controller
// chooses one kind of counter on it as the ability resolves, with one kind
// there is nothing to ask and with none nothing happens. It is not proliferate
// (CR 701.34). A counter on a player is a poison, energy, experience or other
// player counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "af42079b-a3c0-448c-9bb2-b915252e87a9",
		Name:         "Animation Module",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if effects.B33CountersPlacedDelta(ev, game.CounterPlusOne, g) <= 0 {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.Target)
				return ok && c.Controller == source.Controller && effects.OnBattlefield(g, ev.Target)
			}, "Animation Module — you may pay {1}; if you do, create a 1/1 colorless Servo artifact creature token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Animation Module — pay {1} to create a 1/1 colorless Servo artifact creature token?",
						Then:     animationModulePaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Choose a counter on target permanent or player. Give that permanent or player another counter of that kind.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.ArdennTargetPermanentOrPlayer(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				t, ok := effects.FirstLegalTarget(ctx)
				if !ok {
					return nil
				}
				return animationModuleChoose(ctx, t.ID)
			},
		}},
	})
}

// animationModulePaid is the "if you do": the Servo. An on-disk key.
var animationModulePaid = effects.PayUnlessBranches("pay-unless/animation-module-servo",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1}.Apply(ctx)
	}, nil)

// animationModuleKinds lists, in a stable order, the kinds of counter on the
// target, which is a player or a permanent.
func animationModuleKinds(ctx *effects.Context, target uuid.UUID) []string {
	if p := ctx.PlayerByID(target); p != nil {
		return effects.MshA6KindsWithCounters(p.Counters)
	}
	if c, ok := ctx.Game.LookupCardForEffect(target); ok {
		return effects.MshA6KindsWithCounters(c.Counters)
	}
	return nil
}

// animationModuleGive gives the target another counter of the kind.
func animationModuleGive(ctx *effects.Context, target uuid.UUID, kind string) error {
	if p := ctx.PlayerByID(target); p != nil {
		return ctx.Game.AddPlayerCounterForEffect(target, kind, 1)
	}
	return effects.AddCounter{Target: target, Kind: kind, N: 1}.Apply(ctx)
}

// animationModuleChoose asks for one kind of counter on the target and gives
// it another of that kind.
func animationModuleChoose(ctx *effects.Context, target uuid.UUID) error {
	kinds := animationModuleKinds(ctx, target)
	switch len(kinds) {
	case 0:
		return nil
	case 1:
		return animationModuleGive(ctx, target, kinds[0])
	}
	opts := make([]game.ChoiceOption, len(kinds))
	for i, k := range kinds {
		opts[i] = game.ChoiceOption{Label: "A " + k + " counter"}
	}
	return effects.PickOption{
		Question: "Animation Module — choose a counter on the target",
		Options:  opts,
		ThenKey:  animationModuleKindPicked,
		Carry:    []uuid.UUID{target},
		Words:    kinds,
	}.Apply(ctx)
}

// animationModuleKindPicked gives the target the chosen kind. An on-disk key.
var animationModuleKindPicked = effects.OptionPickThen("option-pick/animation-module-counter-kind", func(ctx *effects.Context, r game.OptionPicked) error {
	if len(r.Carry) == 0 || r.Index < 0 || r.Index >= len(r.Words) {
		return nil
	}
	return animationModuleGive(ctx, r.Carry[0], r.Words[r.Index])
})
