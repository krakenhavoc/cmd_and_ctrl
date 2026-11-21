package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cytoplast Manipulator — Creature — Human Wizard Mutant {2}{U}{U}, 0/0
// (Ravnica Allegiance reprint, #5908):
//
//	"Graft 2 (This creature enters with two +1/+1 counters on it.
//	 Whenever another creature enters, you may move a +1/+1 counter from
//	 this creature onto it.)
//	 {U}, {T}: Gain control of target creature with a +1/+1 counter on
//	 it for as long as this creature remains on the battlefield."
//
// Graft is the entry replacement plus cmdr2B23Graft (CR 702.58a). The
// control effect lasts "for as long as this creature remains on the
// battlefield" (CR 611.2b): DurationWhileSourceRemains, as Sower of
// Temptation. The target must have a +1/+1 counter as it is chosen and
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "026ef98e-4f0a-4a6d-b74f-d17ac8b10e37",
		Name:         "Cytoplast Manipulator",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Graft 2 — enters with two +1/+1 counters on it."),
		},
		Triggered: []game.TriggeredAbility{effects.Cmdr2B23Graft("Cytoplast Manipulator")},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: Gain control of target creature with a +1/+1 counter on it for as long as this creature remains on the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature with a +1/+1 counter on it", effects.Not(effects.B13WithoutPlusOneCounter())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				d, ok := effects.DurationWhileSourceRemains(ctx, item.SourceCardID)
				if !ok {
					return nil
				}
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					return effects.GainControl{
						Target:     t.ID,
						Controller: item.Controller,
						Duration:   d,
						Label:      "Cytoplast Manipulator — control for as long as the Manipulator remains",
					}.Apply(ctx)
				}
				return nil
			},
		}},
	})
}
