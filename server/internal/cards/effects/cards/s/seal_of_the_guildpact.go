package cardss

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seal of the Guildpact — Artifact {5} (Ravnica Remastered, #5948):
//
//	"As this artifact enters, choose two colors.
//	 Each spell you cast costs {1} less to cast for each of the chosen
//	 colors it is."
//
// The two colours are asked one after another, the second prompt without
// the first answer, so they are different colours, and both are kept as
// the letters of the permanent's ChosenColor (Game.AppendChosenColorForEffect).
// The discount is one reducer over every spell its controller casts
// (CR 601.2f, generic mana only): it counts the chosen colours the spell is,
// so a spell that is both chosen colours costs {2} less and one that is
// neither costs nothing less. Until both answers are in, the Seal discounts
// only the colours already chosen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a2d60bc-3024-4a1b-ad74-81645fd227c4",
		Name:         "Seal of the Guildpact",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, card.Controller, card.InstanceID,
				"Seal of the Guildpact — choose the first of two colors",
				game.KeyedTail{Ref: sealOfTheGuildpactFirst, Player: card.Controller, Source: card.InstanceID})
			return nil
		},
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				n := 0
				for _, c := range q.Card.Colors {
					if c != "" && strings.Contains(q.Source.ChosenColor, strings.ToUpper(c)) {
						n++
					}
				}
				return n
			}, "Each spell you cast costs {1} less to cast for each of the chosen colors it is.", effects.YourSpell()),
		},
	})
}

// sealOfTheGuildpactFirst records the first colour and asks for the second,
// which may not be the first. An on-disk key: never renamed, never reused.
var sealOfTheGuildpactFirst = effects.CardTail("tail/seal-of-the-guildpact-first-color", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	g.AppendChosenColorForEffect(o.Tail.Source, o.Tail.Word)
	effects.ChooseColorOtherThanTail(game.ColorForBenefit, g, o.Tail.Player, o.Tail.Source,
		"Seal of the Guildpact — choose the second color", o.Tail.Word,
		game.KeyedTail{Ref: sealOfTheGuildpactSecond, Player: o.Tail.Player, Source: o.Tail.Source})
	return nil
})

// sealOfTheGuildpactSecond records the second colour. An on-disk key.
var sealOfTheGuildpactSecond = effects.CardTail("tail/seal-of-the-guildpact-second-color", func(ctx *effects.Context, o game.TailOutcome) error {
	ctx.Game.AppendChosenColorForEffect(o.Tail.Source, o.Tail.Word)
	return nil
})
