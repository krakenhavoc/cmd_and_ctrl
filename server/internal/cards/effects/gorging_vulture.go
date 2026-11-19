package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "67330f45-d08c-4fe2-8ae3-8d18939c512a",
		Name:            "Gorging Vulture",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gorging Vulture — mill four cards, then gain 1 life for each creature card milled this way",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{N: 4, ThenTail: ctx.ItemTail(gorgingVultureMilled)}.Apply(ctx)
				}),
		},
	})
}

var gorgingVultureMilled = CardTail("gorging-vulture/after-mill", func(ctx *Context, o game.TailOutcome) error {
	n := 0
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			n++
		}
	}
	if n == 0 {
		return nil
	}
	return GainLife{Player: ctx.Controller(), Amount: n}.Apply(ctx)
})
