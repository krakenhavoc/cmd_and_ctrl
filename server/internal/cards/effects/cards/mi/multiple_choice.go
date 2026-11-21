package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Multiple Choice — Sorcery {X}{U} (Jumpstart 2022, #5948):
//
//	"If X is 1, scry 1, then draw a card.
//	 If X is 2, you may choose a player. They return a creature they control
//	 to its owner's hand.
//	 If X is 3, create a 4/4 blue and red Elemental creature token.
//	 If X is 4 or more, do all of the above."
//
// X is the one announced with the cast, so the three paragraphs are read
// off it: the first runs for X of 1 or 4+, the second for 2 or 4+, the
// third for 3 or 4+. They run in printed order (CR 608.2c) as a chain of
// continuations, each starting when the one before has finished its prompt:
// the scry (Preordain's tail) then the draw, then the "you may" (MayChoice),
// the choice of a player (ChoosePlayer, not a target) and that player's own
// pick of one creature they control to bounce (ChoosePermanents with the
// chosen player as chooser), then the token. A player with no creature, a
// "no", or nobody chosen skips to the next paragraph.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4d23acf-ff1f-4a6a-9ede-2667a99e9665",
		Name:         "Multiple Choice",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return multipleChoiceScryDraw(ctx)
		},
	})
}

func multipleChoiceScryDraw(ctx *effects.Context) error {
	if x := ctx.X(); x != 1 && x < 4 {
		return multipleChoiceBounce(ctx)
	}
	return effects.Scry{Player: ctx.Controller(), N: 1, ThenTail: ctx.ItemTail(multipleChoiceDrew)}.Apply(ctx)
}

// multipleChoiceDrew is "then draw a card", after which the next paragraph
// runs. An on-disk key: never renamed, never reused.
var multipleChoiceDrew = effects.CardTail("multiple-choice/scry-then-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	if err := (effects.DrawCards{N: 1}).Apply(ctx); err != nil {
		return err
	}
	return multipleChoiceBounce(ctx)
})

func multipleChoiceBounce(ctx *effects.Context) error {
	if x := ctx.X(); x != 2 && x < 4 {
		return multipleChoiceToken(ctx)
	}
	return effects.MayChoice{
		Question: "Multiple Choice — choose a player? They return a creature they control to its owner's hand.",
		YesLabel: "Choose a player",
		NoLabel:  "Decline",
		Then:     multipleChoiceMay,
	}.Apply(ctx)
}

var multipleChoiceMay = effects.MayYesNo("multiple-choice/may-choose-player",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return effects.ChoosePlayer{
			Question: "Multiple Choice — choose a player; they return a creature they control to its owner's hand",
			ThenKey:  multipleChoicePlayer,
		}.Apply(ctx)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return multipleChoiceToken(ctx) })

var multipleChoicePlayer = effects.ChoosePlayerThen("option-pick/multiple-choice-player",
	func(ctx *effects.Context, chosen uuid.UUID, _ []uuid.UUID) error {
		if chosen == uuid.Nil {
			return multipleChoiceToken(ctx)
		}
		return effects.ChoosePermanents{
			Player:   chosen,
			Question: "Multiple Choice — return a creature you control to its owner's hand",
			Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller == of && effects.MatchCreature(c) {
						out = append(out, c.InstanceID)
					}
				}
				return out, 1, 1
			},
			Keyed: multipleChoiceBounced,
		}.Apply(ctx)
	})

var multipleChoiceBounced = effects.RegisterPermanentsPick("prompt-run/multiple-choice-bounce", effects.PermanentsDef{
	Then: func(ctx *effects.Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if err := (effects.BounceToHand{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
		return multipleChoiceToken(ctx)
	},
})

func multipleChoiceToken(ctx *effects.Context) error {
	if x := ctx.X(); x != 3 && x < 4 {
		return nil
	}
	return effects.CreateToken{Template: effects.TokenCard("4/4 blue and red Elemental"), N: 1}.Apply(ctx)
}
