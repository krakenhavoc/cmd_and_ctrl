package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death's Caress — Sorcery {3}{B}{B}:
//
//	"Destroy target creature. If that creature was a Human, you gain life
//	 equal to its toughness."
//
// The creature is read as the spell resolves, before it is destroyed: it was a
// Human (a changeling counts) and has that toughness whether or not the
// destruction takes (indestructible, a regeneration shield). The life gain is
// not conditioned on the destruction.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aac711d1-2fa0-42f9-9eda-a1642f3ff016",
		Name:         "Death's Caress",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 1},
		Targets:      effects.Harmful(effects.TargetCreature("target creature")),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			gain := 0
			if c.HasSubtypeWord("Human") {
				gain = max(c.CurrentToughness(), 0)
			}
			// The life gain runs from the destroy's continuation: a creature
			// that stops the destruction to answer a replacement prompt must
			// not have been paid for yet. The tail does not ask whether
			// anything was destroyed; the sentence does not depend on it.
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, effects.WithN(ctx.Tail(deathsCaressGain), gain))
		},
	})
}

// deathsCaressGain is the second sentence: you gain life equal to the
// toughness (N) a Human target had, read before it was destroyed.
var deathsCaressGain = effects.CardTail("deaths-caress/gain-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: o.Tail.N}.Apply(ctx)
})
