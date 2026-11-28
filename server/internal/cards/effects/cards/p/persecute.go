package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Persecute — Sorcery {2}{B}{B} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"Choose a color. Target player reveals their hand and discards all cards of
//	 that color."
//
// Addle's colour choice, then the target's hand is revealed and every card of
// the chosen colour in it goes at once (no choice for the discarder). The
// colour is asked as the spell resolves; the target is re-checked first
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "676c4cfa-f929-43aa-95cc-df81a596c2a2",
		Name:         "Persecute",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			victim := effects.TargetedPlayer(ctx)
			if victim == uuid.Nil {
				return nil
			}
			effects.ChooseColorTail(game.ColorForFilter, ctx.Game, ctx.Controller(), ctx.Source(), "Persecute — choose a color",
				effects.WithCarry(ctx.ItemTail(persecuteColorChosen), victim))
			return nil
		},
	})
}

// persecuteColorChosen reveals the target's hand and discards every card of
// the chosen colour. An on-disk key: never renamed, never reused.
var persecuteColorChosen = effects.CardTail("then/persecute-color-chosen", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.Tail.Carry) == 0 {
		return nil
	}
	victim := o.Tail.Carry[0]
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var ids []uuid.UUID
	for _, c := range p.Hand.Cards {
		if c.HasColor(o.Tail.Word) {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(victim, ids, ctx.Source(), game.KeyedTail{})
})
