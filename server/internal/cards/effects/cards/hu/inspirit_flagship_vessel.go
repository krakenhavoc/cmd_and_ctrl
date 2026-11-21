package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inspirit, Flagship Vessel — Legendary Artifact — Spacecraft {U}{R}{W},
// 5/5 (Edge of Eternities Commander):
//
//	"Station (Tap another creature you control: Put charge counters equal
//	 to its power on this Spacecraft. Station only as a sorcery. It's an
//	 artifact creature at 8+.)
//	 1+ | At the beginning of combat on your turn, put your choice of a
//	 +1/+1 counter or two charge counters on up to one other target
//	 artifact.
//	 8+ | Flying
//	 Other artifacts you control have hexproof and indestructible."
//
// The 1+ trigger is Additive Evolution's beginning-of-combat watcher with
// an "up to one other target artifact" clause (Another, minimum zero) and
// a registered option pick for the counters (energy_resolution.go's
// shape), behind the charge-counter gate. The 8+ grant is a layer-6
// static over the controller's other artifacts under the same gate.
//
// No simplification.
func init() {
	spec := effects.Another(effects.TargetPermanent("up to one other target artifact", effects.Artifact()))
	spec.Min = 0
	combat := effects.Targeting(
		effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
			"Inspirit, Flagship Vessel — put your choice of a +1/+1 counter or two charge counters on up to one other target artifact",
			inspiritCounters),
		spec)
	combat.ActiveWhen = effects.AtChargeCounters(1)
	grant := game.StaticAbility{
		Label:      "Other artifacts you control have hexproof and indestructible.",
		Layer:      game.Layer6Ability,
		ActiveWhen: effects.AtChargeCounters(8),
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			effects.AppendKeywordsTo(c, []string{"hexproof", "indestructible"})
		},
		Keywords: []string{"hexproof", "indestructible"},
	}
	effects.Register(effects.Spec{
		OracleID:     "554df866-3dbb-4811-8573-6033481591aa",
		Name:         "Inspirit, Flagship Vessel",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Triggered:    []game.TriggeredAbility{combat},
		Static: append(append(
			effects.SpacecraftAt(8, 5, 5),
			effects.ThresholdKeywords(8, "flying")),
			grant),
	})
}

const (
	inspiritPlusOne = "A +1/+1 counter"
	inspiritCharge  = "Two charge counters"
)

// inspiritCounters asks which counters the still-legal target gets.
func inspiritCounters(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return effects.PickOption{
			Question: "Inspirit, Flagship Vessel — put your choice of a +1/+1 counter or two charge counters on the artifact",
			Options:  []game.ChoiceOption{{Label: inspiritPlusOne}, {Label: inspiritCharge}},
			ThenKey:  inspiritChoice,
			Carry:    []uuid.UUID{t.ID},
		}.Apply(ctx)
	}
	return nil
}

// inspiritChoice places the chosen counters. An on-disk key.
var inspiritChoice = effects.OptionPickThen("option-pick/inspirit-flagship-vessel-counters", func(ctx *effects.Context, r game.OptionPicked) error {
	if r.Option == nil || len(r.Carry) < 1 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	if r.Option.Label == inspiritCharge {
		return effects.AddCounter{Target: r.Carry[0], Kind: game.CounterCharge, N: 2}.Apply(ctx)
	}
	return effects.AddCounter{Target: r.Carry[0], Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
