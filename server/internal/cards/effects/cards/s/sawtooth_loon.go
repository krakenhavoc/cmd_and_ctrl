package cardss

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sawtooth Loon — Creature — Bird {2}{W}{U}, 2/2 (Dominaria Remastered,
// #5948):
//
//	"Flying
//	 When this creature enters, return a white or blue creature you control
//	 to its owner's hand.
//	 When this creature enters, draw two cards, then put two cards from your
//	 hand on the bottom of your library."
//
// Two triggers. The first is not a "may": its controller must name a white
// or blue creature they control as it resolves, and the Loon (white and
// blue) is always one unless it has left, so a pick is always offered (it
// is the same pick Loyal Gryff's bounce uses). The second draws two, then
// asks for two cards from the hand, one prompt each, the first card put
// being the one deeper in the library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b917e472-2550-4381-b756-45487c2c187a",
		Name:            "Sawtooth Loon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sawtooth Loon — return a white or blue creature you control to its owner's hand", func(g *game.Game, item *game.StackItem) error {
				text := g.TextReaderForEffect(item)
				white, blue := text.TextColor("W"), text.TextColor("U")
				return effects.ChoosePermanents{
					Question: "Sawtooth Loon — return a white or blue creature you control to its owner's hand",
					Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
						var out []uuid.UUID
						for _, c := range g.BattlefieldCardsForEffect() {
							if c.Controller == of && c.IsCreature() && (c.HasColor(white) || c.HasColor(blue)) {
								out = append(out, c.InstanceID)
							}
						}
						return out, 1, 1
					},
					Keyed: effects.S82MidB01Return,
				}.Apply(effects.NewContext(g, item))
			}),
			effects.WhenThisEnters("Sawtooth Loon — draw two cards, then put two cards from your hand on the bottom of your library",
				effects.Do(effects.DrawCards{N: 2},
					effects.Instruction(sawtoothLoonBottom),
					effects.Instruction(sawtoothLoonBottom))),
		},
	})
}

func sawtoothLoonBottom(ctx *effects.Context) error {
	return effects.RfPutCardFromHandOnBottom(ctx, "Sawtooth Loon — put a card from your hand on the bottom of your library")
}
