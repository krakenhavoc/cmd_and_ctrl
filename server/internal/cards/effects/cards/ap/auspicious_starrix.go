package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Auspicious Starrix — Creature — Elk Beast {4}{G}, 6/6 (Ikoria: Lair of
// Behemoths, #5512):
//
//	"Mutate {5}{G} (If you cast this spell for its mutate cost, put it over or
//	 under target non-Human creature you own. They mutate into the creature on
//	 top plus all abilities from under it.)
//	 Whenever this creature mutates, exile cards from the top of your library
//	 until you exile X permanent cards, where X is the number of times this
//	 creature has mutated. Put those permanent cards onto the battlefield."
//
// Mutate is the engine's (ADR 0223). The run is Illuna's shared exile-until
// primitive with a registered stop clause that counts permanent cards against
// X (S82OldB111TimesMutated, read as the trigger resolves); the cards exiled
// on the way stay in exile. A library with fewer than X permanent cards exiles
// all of itself and puts the ones it found onto the battlefield. The permanent
// cards enter together under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "8432df19-dbf4-457b-8c0b-1f4b3a0a37d2",
		Name:             "Auspicious Starrix",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Mutate("{5}{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.ThisCreatureMutates,
				"Auspicious Starrix — exile cards from the top of your library until you exile X permanent cards, where X is the number of times this creature has mutated, then put those permanent cards onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    game.MillUntil{Ref: starrixUntilPermanents, N: effects.S82OldB111XOfThisMutation(ctx)},
						ThenTail: ctx.ItemTail(starrixExiled),
					}.Apply(ctx)
				}),
		},
	})
}

// starrixUntilPermanents is the run's stop clause: N permanent cards have
// been exiled. An on-disk key.
var starrixUntilPermanents = game.RegisterMillUntil("until/starrix-n-permanent-cards", func(landed []game.Card, n int) bool {
	count := 0
	for _, c := range landed {
		if c.IsPermanent() && !c.IsToken() {
			count++
		}
	}
	return count >= n
})

// starrixExiled puts the permanent cards the run exiled onto the battlefield
// together. An on-disk key.
var starrixExiled = effects.CardTail("starrix/after-exile", func(ctx *effects.Context, o game.TailOutcome) error {
	var permanents []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsPermanent() && !c.IsToken() {
			permanents = append(permanents, id)
		}
	}
	return effects.ReturnFromExileTogether{Targets: permanents, Controller: ctx.Controller()}.Apply(ctx)
})
