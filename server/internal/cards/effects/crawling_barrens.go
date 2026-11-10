package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "dfe1a112-97aa-4e81-8431-81552ba2cdcf",
		Name:         "Crawling Barrens",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}: Put two +1/+1 counters on this land. Then you may have it become a 0/0 Elemental creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				if err := (AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
					return err
				}
				return MayChoice{
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
var crawlingBarrensAnimate = MayYes("confirm/crawling-barrens-animate", func(ctx *Context, _ game.ConfirmAnswer) error {
	source := ctx.Source()
	if source == uuid.Nil || ctx.isNewSourceObjectAsThis(source) {
		return nil
	}
	return ScopedEffectFor{
		Target: source,
		Mods: append([]game.Mod{
			game.AddTypesMod("Creature"),
			game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
		}, game.SetBasePTMods(0, 0)...),
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Crawling Barrens — a 0/0 Elemental creature until end of turn",
	}.Apply(ctx)
})
