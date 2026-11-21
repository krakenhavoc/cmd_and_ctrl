package cardsso

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spectral Adversary — Creature — Spirit {1}{U}, 2/1:
//
//	"Flash
//	 Flying
//	 When this creature enters, you may pay {1}{U} any number of times.
//	 When you pay this cost one or more times, put that many +1/+1
//	 counters on this creature, then up to that many other target
//	 artifacts, creatures, and/or enchantments phase out."
//
// The Adversary cycle's shape (Bloodthirsty Adversary). The targets are
// chosen when the reflexive trigger goes on the stack (CR 603.3d); the
// ones still legal phase out together as it resolves (CR 702.26).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0de4171e-5908-408e-9852-25f148d0fb47",
		Name:            "Spectral Adversary",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spectral Adversary — pay {1}{U} any number of times; counters, then phase out other permanents", spectralAdversaryEnters),
		},
	})
}

var spectralAdversaryBody = game.ReflexiveBody("spectral-adversary/phase-out",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		if err := (effects.CounterOnThis{Kind: game.CounterPlusOne, N: p.Amount}).Apply(ctx); err != nil {
			return err
		}
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		return effects.PhaseOut{Targets: ids}.Apply(ctx)
	},
	func(src uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return effects.Another(effects.TargetPermanent("up to "+strconv.Itoa(p.Amount)+" other target artifacts, creatures, and/or enchantments",
			effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()))).WithCount(0, p.Amount)
	})

var spectralAdversaryPaid = effects.PayUnlessBranches("pay-unless/spectral-adversary-times",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if r.Number <= 0 {
			return nil
		}
		return effects.ReflexiveTrigger{
			Label:  "Spectral Adversary — " + strconv.Itoa(r.Number) + " +1/+1 counters, then up to that many other permanents phase out",
			Body:   spectralAdversaryBody,
			Params: game.EffectParams{Amount: r.Number},
		}.Apply(ctx)
	}, nil)

func spectralAdversaryEnters(g *game.Game, item *game.StackItem) error {
	return effects.MayPayX{
		Symbol: "{1}{U}",
		Label:  "Spectral Adversary",
		Buys:   "put X +1/+1 counters on it, then phase out up to X other artifacts, creatures and/or enchantments",
		Goal:   effects.XAsHighAsYouCan,
		Then:   spectralAdversaryPaid,
	}.Apply(effects.NewContext(g, item))
}
