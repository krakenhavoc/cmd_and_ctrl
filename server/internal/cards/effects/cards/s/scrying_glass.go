package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scrying Glass — Artifact {2} (Urza's Destiny, S82 #5621):
//
//	"{3}, {T}: Choose a number greater than 0 and a color. Target opponent
//	 reveals their hand. If that opponent reveals exactly the chosen number of
//	 cards of the chosen color, you draw a card."
//
// The number and the colour are chosen as the ability resolves (CR 608.2c), in
// the printed order and before the hand is looked at; the hand is then revealed
// to the table (CR 701.20a) and its cards of the chosen colour counted as each
// card has it in hand, a multicoloured card counting for each of its colours
// (CR 105.2). A hand of exactly that many draws the activator a card; a
// target that left the game, or an illegal target, does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca16b32c-3f1e-4f8d-821a-064f922c45d8",
		Name:         "Scrying Glass",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Choose a number greater than 0 and a color. Target opponent reveals their hand. If that opponent reveals exactly the chosen number of cards of the chosen color, you draw a card.",
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim := effects.TargetedPlayer(ctx)
				if victim == uuid.Nil {
					return nil
				}
				return effects.ChooseNumber{
					Min:      1,
					NoMax:    true,
					Question: "Scrying Glass — choose a number greater than 0",
					Unit:     game.PayAmountOther,
					Goal:     func(*effects.Context) int { return 1 },
					Carry:    []uuid.UUID{victim},
					Then:     scryingGlassNumber,
				}.Apply(ctx)
			},
		}},
	})
}

// scryingGlassNumber is the number's answer: the colour is asked next, the
// target player carried and the number held on the tail. An on-disk key.
var scryingGlassNumber = effects.NumberThen("number/s82-old-b705-scrying-glass", func(ctx *effects.Context, r game.NumberAnswer) error {
	if len(r.Carry) == 0 {
		return nil
	}
	effects.ChooseColorTail(game.ColorForFilter, ctx.Game, ctx.Controller(), ctx.Source(), "Scrying Glass — choose a color",
		effects.WithCarry(effects.WithN(ctx.ItemTail(scryingGlassColor), r.Number), r.Carry[0]))
	return nil
})

// scryingGlassColor is the colour's answer (the tail's Word): the target
// opponent's hand is revealed and its cards of that colour counted against the
// number (the tail's N). An on-disk key.
var scryingGlassColor = effects.CardTail("then/s82-old-b705-scrying-glass-color", func(ctx *effects.Context, o game.TailOutcome) error {
	victim := effects.TailCarry0(o)
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil || p.Hand == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	n := 0
	for _, c := range p.Hand.Cards {
		if c.HasColorWord(o.Tail.Word) {
			n++
		}
	}
	if n != o.Tail.N {
		return nil
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
})
