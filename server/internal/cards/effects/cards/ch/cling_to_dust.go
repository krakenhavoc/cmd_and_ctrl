package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cling to Dust — Instant {B}:
//
//	"Exile target card from a graveyard. If it was a creature card,
//	 you gain 3 life. Otherwise, you draw a card.
//	 Escape—{3}{B}, Exile five other cards from your graveyard."
//
// Graveyard hate that never runs out, and the escape card that best
// shows the cost and the effect pulling in opposite directions: the
// spell exiles ONE card from ANY graveyard, and escaping it exiles
// five more from YOUR OWN. The second cast is therefore a real
// decision rather than a free repeat.
//
// The "if it was a creature card" clause reads the type BEFORE the
// exile — same shape as Scavenging Ooze — because after the move the
// card is in exile and the question is about what it was. It also
// WAITS for the exile (#911, ADR 0013 §5t): the clause is about the
// card the first sentence moved, so a card the CR 614 window kept in
// its graveyard gains nobody life and draws nobody a card. Both
// halves live in ExileThenIfItWas.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9c44e556-4c7a-48bf-a108-c584f69c2cfa",
		Name:             "Cling to Dust",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetCardInGraveyard("target card from a graveyard"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Escape("{3}{B}", 5)},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.ExileThenIfItWas{
					Target: t.ID,
					Was:    effects.WasCreatureCard,
					Keyed:  clingToDustIfCreature,
				}.Apply(ctx)
			}
			return nil
		},
	})
}

// clingToDustIfCreature is the exile's keyed "if it was a creature card,
// you gain 3 life. Otherwise, you draw a card" (#4253): the spell rides the
// tail, so "you" is its controller. An on-disk key: never renamed, never
// reused.
var clingToDustIfCreature = effects.IfItWasTail("if-it-was/cling-to-dust-life-or-draw", func(ctx *effects.Context) error {
	return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
}, func(ctx *effects.Context) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
