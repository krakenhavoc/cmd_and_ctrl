package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gorging Vulture — Creature — Bird {2}{B}, 2/2 (Ravnica: Clue Edition,
// #5929):
//
//	"Flying
//	 When this creature enters, mill four cards. You gain 1 life for each
//	 creature card milled this way."
//
// Dawnhand Eulogist's mill with a continuation: the tail is told the
// cards that actually reached the graveyard (CR 400.7), so a replacement
// that sent one elsewhere is not counted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67330f45-d08c-4fe2-8ae3-8d18939c512a",
		Name:            "Gorging Vulture",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gorging Vulture — mill four cards, then gain 1 life for each creature card milled this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{N: 4, ThenTail: ctx.ItemTail(gorgingVultureMilled)}.Apply(ctx)
				}),
		},
	})
}

var gorgingVultureMilled = effects.CardTail("gorging-vulture/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
})
