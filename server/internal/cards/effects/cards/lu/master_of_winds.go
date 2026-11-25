package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Master of Winds — Creature — Sphinx Wizard {2}{U}{U}, 1/4 (Zendikar
// Rising, S82 OLD #5508):
//
//	"Flying
//	 When this creature enters, draw two cards, then discard a card.
//	 Whenever you cast an instant, sorcery, or Wizard spell, you may have
//	 this creature's base power and toughness become 4/1 or 1/4 until end
//	 of turn."
//
// The cast trigger asks for a choice of three: leave it alone, 4/1 or 1/4.
// The new base power and toughness is a layer 7b effect on the creature
// that lasts until end of turn (CR 613.4b); a Master that has left by the
// time the trigger resolves is not changed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7accdc91-3140-4096-a71f-0cee38d29e96",
		Name:            "Master of Winds",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Master of Winds — draw two cards, then discard a card",
				func(g *game.Game, item *game.StackItem) error {
					return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, 2, 1, "")
				}),
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery(), effects.HasSubtype("Wizard")),
				"Master of Winds — you may have its base power and toughness become 4/1 or 1/4 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.PickOption{
						Question: "Master of Winds — base power and toughness 4/1 or 1/4 until end of turn?",
						Options: []game.ChoiceOption{
							{Label: "Don't change it"},
							{Label: "4/1"},
							{Label: "1/4"},
						},
						ThenKey: masterOfWindsPicked,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var masterOfWindsPicked = effects.OptionPickThen("option-pick/master-of-winds-base-pt", func(ctx *effects.Context, r game.OptionPicked) error {
	var p, t int
	switch r.Index {
	case 1:
		p, t = 4, 1
	case 2:
		p, t = 1, 4
	default:
		return nil
	}
	src := ctx.Source()
	if src == uuid.Nil {
		return nil
	}
	if _, ok := ctx.Game.LookupCardForEffect(src); !ok {
		return nil
	}
	return effects.UntilEndOfTurn(ctx, src, nil, "Master of Winds — base power and toughness", game.SetBasePTMods(p, t)...)
})
