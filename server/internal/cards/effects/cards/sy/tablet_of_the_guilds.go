package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tablet of the Guilds — Artifact {2} (Return to Ravnica, #5566):
//
//	"As this artifact enters, choose two colors.
//	 Whenever you cast a spell, if it's at least one of the chosen colors, you gain
//	 1 life for each of the chosen colors it is."
//
// The two (different) colours are chosen as it enters, the second from the five
// minus the first, with Seal of the Guildpact's pair of prompts. The trigger
// counts the chosen colours the spell is as it is cast and gains that much; a
// colourless spell, or one of neither colour, does not trigger it (an intervening
// "if", CR 603.4).
//
// No simplification.
func init() {
	const label = "Tablet of the Guilds — gain 1 life for each of the chosen colors the spell is"
	effects.Register(effects.Spec{
		OracleID:     "71ce7034-1400-4e7c-9ef3-37baed669faf",
		Name:         "Tablet of the Guilds",
		Completeness: effects.CompletenessFull,
		AsEnters: func(card *game.Card, ctx *effects.Context) error {
			effects.ChooseColorTail(game.ColorForBenefit, ctx.Game, card.Controller, card.InstanceID,
				"Tablet of the Guilds — choose the first of two colors",
				game.KeyedTail{Ref: tabletOfTheGuildsFirst, Player: card.Controller, Source: card.InstanceID})
			return nil
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.OldB423SpellChosenColors(ev, source, g) > 0
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				n := effects.OldB423SpellChosenColors(ev, source, g)
				if n == 0 {
					return nil
				}
				item := game.NewTriggeredItem(source, label)
				item.Params.Amount = n
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.GainLife{Amount: item.Params.Amount}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var tabletOfTheGuildsFirst = effects.CardTail("tail/tablet-of-the-guilds-first-color", func(ctx *effects.Context, o game.TailOutcome) error {
	g := ctx.Game
	g.AppendChosenColorForEffect(o.Tail.Source, o.Tail.Word)
	effects.ChooseColorOtherThanTail(game.ColorForBenefit, g, o.Tail.Player, o.Tail.Source,
		"Tablet of the Guilds — choose the second color", o.Tail.Word,
		game.KeyedTail{Ref: tabletOfTheGuildsSecond, Player: o.Tail.Player, Source: o.Tail.Source})
	return nil
})

var tabletOfTheGuildsSecond = effects.CardTail("tail/tablet-of-the-guilds-second-color", func(ctx *effects.Context, o game.TailOutcome) error {
	ctx.Game.AppendChosenColorForEffect(o.Tail.Source, o.Tail.Word)
	return nil
})
