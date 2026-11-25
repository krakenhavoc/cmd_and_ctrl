package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa's Revelation — Sorcery {5}{G}{G} (Game Night):
//
//	"Scry 5, then reveal the top card of your library. If it's a creature
//	 card, you draw cards equal to its power and you gain life equal to its
//	 toughness."
//
// The reveal and what follows are the scry's keyed tail, so the card
// revealed is the one the scry left on top. Power and toughness are the
// revealed card's own, read once it is revealed. A non-creature card, or an
// empty library, does nothing. The draw comes before the life gain, as the
// text orders them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba9cd7bc-846f-4b5c-a1dc-9763a60dafc9",
		Name:         "Nissa's Revelation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 5, Draws: 1, LifeGain: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			you := ctx.Controller()
			return effects.Scry{Player: you, N: 5, ThenTail: game.KeyedTail{Ref: nissasRevelationReveal, Player: you}}.Apply(ctx)
		},
	})
}

// nissasRevelationReveal is the clause after the scry. An on-disk key.
var nissasRevelationReveal = effects.CardTail("nissas-revelation/reveal", func(ctx *effects.Context, _ game.TailOutcome) error {
	g, you := ctx.Game, ctx.Controller()
	revealed := g.RevealTopOfLibraryForEffect(you, ctx.Source(), 1, "Nissa's Revelation — reveal the top card of your library")
	if len(revealed) == 0 {
		return nil
	}
	top, ok := g.LookupCardForEffect(revealed[0])
	if !ok || !top.IsCreature() {
		return nil
	}
	if err := (effects.DrawCards{Player: you, N: top.CurrentPower()}).Apply(ctx); err != nil {
		return err
	}
	return effects.GainLife{Player: you, Amount: top.CurrentToughness()}.Apply(ctx)
})
