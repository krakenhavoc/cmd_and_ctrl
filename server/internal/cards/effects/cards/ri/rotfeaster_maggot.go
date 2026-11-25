package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rotfeaster Maggot — Creature — Insect {4}{B}, 3/5 (Battlebond):
//
//	"When this creature enters, exile target creature card from a graveyard.
//	 You gain life equal to that card's toughness."
//
// Carrion Imp's trigger without the "may": the target is chosen as the trigger
// goes on the stack (CR 603.3d). The toughness is read from the card as it
// sits in the graveyard, before it is exiled (CR 608.2h), and the life follows
// only an exile that happened. A card with no toughness gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67d0df0f-37d1-46da-b1c3-4fc9d1e3cc05",
		Name:         "Rotfeaster Maggot",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Rotfeaster Maggot — exile target creature card from a graveyard; you gain life equal to its toughness",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						t, ok := effects.FirstLegalTarget(ctx)
						if !ok {
							return nil
						}
						tail := ctx.ItemTail(rotfeasterMaggotGain)
						if c, found := g.LookupCardForEffect(t.ID); found {
							tail.N = max(c.CurrentToughness(), 0)
						}
						return effects.ExileTarget{Target: t.ID, ThenTail: tail}.Apply(ctx)
					}),
				effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature())),
		},
	})
}

// rotfeasterMaggotGain is the life, once the card has reached exile; its
// toughness rides the tail as N, read in the graveyard. An on-disk key: never
// renamed, never reused.
var rotfeasterMaggotGain = effects.CardTail("then/s82-old-b206-rotfeaster-maggot-gain", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Amount: o.Tail.N}.Apply(ctx)
})
