package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zabaz, the Glimmerwasp — Legendary Artifact Creature — Insect {1}, 0/0
// (Modern Horizons 2, #5501):
//
//	"Modular 1
//	 If a modular triggered ability would put one or more +1/+1 counters on
//	 a creature you control, that many plus one +1/+1 counters are put on it
//	 instead.
//	 {R}: Destroy target artifact you control.
//	 {W}: Zabaz gains flying until end of turn."
//
// The replacement is Hardened Scales' with one more condition: the
// placement is made by the resolving modular dies trigger
// (game.ResolvingModularTriggerForEffect), so a counter from any other
// source is left alone. It stacks with Hardened Scales in CR 616 order.
// Zabaz's own modular entry counter is not a triggered ability and gets no
// bonus.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "549dc9f3-1fda-4ad3-87bc-8a990800380b",
		Name:            "Zabaz, the Glimmerwasp",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"modular 1"},
		Replacements: []game.ReplacementEffect{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
				return effects.PlusOneCounterPlacementOnYourCreature(ev, g, src) && g.ResolvingModularTriggerForEffect()
			},
			Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
				ev.CounterDelta++
				return nil
			},
			Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
				return src.Controller
			},
			Label: "Zabaz, the Glimmerwasp: a modular trigger puts one more +1/+1 counter",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: Destroy target artifact you control.",
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Cost:    effects.ManaCost("{R}"),
				Targets: effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DestroyTheTargetPermanent(item, effects.NewContext(g, item))
				},
			},
			{
				Label:   "{W}: Zabaz gains flying until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant, Pump: &game.Pump{Keywords: []string{"flying"}}},
				Cost:    effects.ManaCost("{W}"),
				Effect:  effects.ThisGainsKeywordUntilEndOfTurn("flying", "Zabaz, the Glimmerwasp — flying until end of turn"),
			},
		},
	})
}
