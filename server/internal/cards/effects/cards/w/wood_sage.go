package cardsw

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wood Sage — Creature — Human Druid {G}{U}, 1/1 (Conspiracy, S82 OLD batch
// 363, #5557):
//
//	"{T}: Choose a creature card name. Reveal the top four cards of your
//	 library and put all of them with that name into your hand. Put the rest
//	 into your graveyard."
//
// The name is asked as the ability resolves (Day of the Moon's prompt) and
// then the top four cards are revealed: every one that is a creature card with
// that name goes to hand and the rest are put (not milled) into the graveyard
// (Mulch's reveal, S82's Cmdr2B15RevealTopTakeAllMatching). A name that is not
// a creature card's takes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fde7bb3-2539-411f-8733-36ce9595dfd5",
		Name:         "Wood Sage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Choose a creature card name. Reveal the top four cards of your library and put all of them with that name into your hand. Put the rest into your graveyard.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
					"Wood Sage — choose a creature card name", ctx.Tail(woodSageNamed))
				return nil
			},
		}},
	})
}

// woodSageNamed is the rest of the ability, told the chosen name. An on-disk
// key: never renamed, never reused.
var woodSageNamed = effects.CardTail("card-name/wood-sage-reveal", func(ctx *effects.Context, o game.TailOutcome) error {
	name := o.Tail.Word
	return effects.Cmdr2B15RevealTopTakeAllMatching(ctx.Game, ctx.Controller(), ctx.Source(), 4,
		"Wood Sage — reveal the top four cards; creature cards named "+name+" to your hand, the rest to your graveyard",
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return effects.CreatureWithChosenCreatureCardName(c, []string{name})
		})
})
