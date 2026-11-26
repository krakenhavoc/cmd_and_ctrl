package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nissa's Defeat — Sorcery {2}{G} (Hour of Devastation, #5534):
//
//	"Destroy target Forest, green enchantment, or green planeswalker. If that
//	 permanent was a Nissa planeswalker, draw a card."
//
// One target clause over three kinds of permanent: a Forest (a land, or any
// permanent that is a Forest), a green enchantment, a green planeswalker
// ("Forest" and "green" are words of the card's text, CR 612.2). The target's
// planeswalker type and Nissa subtype are read from the permanent that was
// destroyed, as it last existed (CR 608.2h), and the draw waits for the CR
// 614 window like any payout on a destroy: a Nissa that is indestructible or
// regenerates is not destroyed and draws nothing. A target that has gone makes
// the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f808e3b-6b69-4b46-8a92-36d960527acc",
		Name:         "Nissa's Defeat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.TargetPermanent("target Forest, green enchantment, or green planeswalker",
			effects.Or(
				effects.HasSubtype("Forest"),
				effects.And(effects.Enchantment(), effects.OfColor("G")),
				effects.And(effects.Planeswalker(), effects.OfColor("G")),
			)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, ctx.ItemTail(nissasDefeatDraw))
		},
	})
}

// nissasDefeatDraw is "if that permanent was a Nissa planeswalker, draw a
// card", told the permanent that was destroyed: its types and subtypes as it
// last existed (CR 608.2h). An on-disk key: never renamed, never reused.
var nissasDefeatDraw = effects.CardTail("destroy/nissas-defeat-draw-for-a-nissa", func(ctx *effects.Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if ok && c.IsPlaneswalker() && c.HasSubtype(ctx.TextSubtype("Nissa")) {
			return effects.DrawCards{N: 1}.Apply(ctx)
		}
	}
	return nil
})
