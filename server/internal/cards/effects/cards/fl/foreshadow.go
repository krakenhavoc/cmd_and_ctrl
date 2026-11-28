package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Foreshadow — Instant {1}{U} (Visions, S82 OLD batch 787, #5631):
//
//	"Choose a card name, then target opponent mills a card. If a card with the
//	 chosen name was milled this way, you draw a card.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// The caster is asked for a card name (free text, CR 201.2; a split or
// double-faced card matches under any face) by a keyed card-name prompt, then
// the target opponent mills one card. The draw is for a card that landed in
// the graveyard with the chosen name: a card another effect exiled instead
// was not milled. The delayed draw at the next upkeep (CR 603.7a) is set up
// last, whatever the guess. With its one target gone the spell does not
// resolve, so no name is asked and nothing is drawn.
//
// No simplification.
func init() {
	foreshadowNamed = effects.CardTail("then/s82-old-b787-foreshadow-named", func(ctx *effects.Context, o game.TailOutcome) error {
		if len(o.Tail.Carry) == 0 {
			return foreshadowFinish(ctx)
		}
		t := ctx.ItemTail(foreshadowMilled)
		t.Word = o.Tail.Word
		return ctx.Game.MillToZoneKeyedForEffect(o.Tail.Carry[0], 1, game.ZoneGraveyard, game.MillUntil{}, t)
	})
	foreshadowMilled = effects.CardTail("then/s82-old-b787-foreshadow-milled", func(ctx *effects.Context, o game.TailOutcome) error {
		for _, id := range o.IDs {
			c, ok := ctx.Game.LookupCardForEffect(id)
			if ok && game.CardNameMatches(c, o.Tail.Word) {
				if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				break
			}
		}
		return foreshadowFinish(ctx)
	})
	effects.Register(effects.Spec{
		OracleID:     "4caa97df-1f6c-4806-bf14-e589cc1badb5",
		Name:         "Foreshadow",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			opp := effects.TargetedPlayerOf(ctx)
			if opp == uuid.Nil {
				return nil
			}
			t := ctx.ItemTail(foreshadowNamed)
			t.Carry = []uuid.UUID{opp}
			ctx.Game.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(), "Foreshadow — choose a card name", t)
			return nil
		},
	})
}

var foreshadowNamed, foreshadowMilled game.TailRef

// foreshadowFinish is "Draw a card at the beginning of the next turn's
// upkeep", the last sentence of the spell.
func foreshadowFinish(ctx *effects.Context) error {
	return effects.DrawACardAtTheNextUpkeep("Foreshadow — draw a card").Apply(ctx)
}
