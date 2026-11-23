package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crawling Barrens — Land (Foundations):
//
//	"{T}: Add {C}.
//	 {4}: Put two +1/+1 counters on this land. Then you may have it
//	 become a 0/0 Elemental creature until end of turn. It's still a land."
//
// The counters go on first, unconditionally. The "you may" is asked
// partway through the resolution (MayChoice), and a "yes" is one
// continuous effect (CR 611.2a) pinned to this object (CR 400.7): the
// creature type and Elemental subtype (layer 4) and a base 0/0 (layer
// 7b), under the counters. A Barrens that left and came back is a new
// object and takes neither the counters nor the animation.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfe1a112-97aa-4e81-8431-81552ba2cdcf",
		Name:         "Crawling Barrens",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: Put two +1/+1 counters on this land. Then you may have it become a 0/0 Elemental creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return effects.MayChoice{
					Question: "Crawling Barrens — have it become a 0/0 Elemental creature until end of turn?",
					YesLabel: "Become a creature",
					NoLabel:  "Stay a land",
					Then:     crawlingBarrensAnimate,
				}.Apply(ctx)
			},
		}},
	})
}

// crawlingBarrensAnimate is the "yes" branch: the land becomes a 0/0
// Elemental creature until end of turn.
var crawlingBarrensAnimate = effects.MayYes("confirm/crawling-barrens-animate", func(ctx *effects.Context, _ game.ConfirmAnswer) error {
	source := ctx.Source()
	if source == uuid.Nil || ctx.IsNewSourceObjectAsThis(source) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target: source,
		Mods: append([]game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
		}, game.SetBasePTMods(0, 0)...),
		Duration: effects.DurationUntilEndOfTurn(ctx),
		Label:    "Crawling Barrens — a 0/0 Elemental creature until end of turn",
	}.Apply(ctx)
})
