package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revival Experiment — Sorcery {4}{B}{G} (Commander 2021, #5504):
//
//	"For each permanent type, return up to one card of that type from your
//	 graveyard to the battlefield. You lose 3 life for each card returned
//	 this way. Exile Revival Experiment."
//
// The six permanent types are artifact, battle, creature, enchantment, land
// and planeswalker (CR 110.4). The pick is one choose-cards prompt over the
// permanent cards in your graveyard, none to six of them, with a set rule
// (RegisterChooseCardsThenRule): the picked cards must be assignable to
// DISTINCT types each has, one card per type, so a card with two types
// (an artifact creature) serves one of them and two cards of the same only
// type are refused. They return together as one entry, and the life is 3
// for each card that really entered. The spell is exiled first, as
// Sheoldred's Restoration's is: it is on the stack, so the order is not
// observable.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "913ec957-5f4e-46b4-a819-9448837f72a3",
		Name:         "Revival Experiment",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.ExileTarget{Target: ctx.Source()}).Apply(ctx); err != nil {
				return err
			}
			you := ctx.Controller()
			candidates := effects.GraveyardCardIDs(ctx, you, func(c game.Card) bool {
				return c.IsPermanent() && !c.IsToken()
			})
			if len(candidates) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:     you,
				FromPlayer:  you,
				Source:      ctx.Source(),
				Question:    "Revival Experiment — return up to one card of each permanent type from your graveyard to the battlefield",
				Cards:       candidates,
				Min:         0,
				Max:         len(effects.OldB054PermanentTypes),
				Destination: game.ChooseOntoBattlefield,
				ThenKey:     revivalExperimentChosen,
				LiveZone:    game.ZoneGraveyard,
				Asked:       game.NewAskedItem(ctx.Item),
			})
			return nil
		},
	})
}

// revivalExperimentChosen is the pick's answer: the cards return together,
// then the life. An on-disk key.
var revivalExperimentChosen = game.RegisterChooseCardsThenRule("choose-cards/revival-experiment-chosen",
	effects.OldB054OnePerPermanentType,
	func(g *game.Game, r game.CardsChosen) error {
		if !r.Answered || len(r.Picked) == 0 {
			return nil
		}
		ctx, _ := effects.AskedContext(g, r.Asked, r.Chooser, r.Source, nil)
		return effects.ReturnFromGraveyardTogether{Targets: r.Picked, ThenTail: revivalExperimentLife}.Apply(ctx)
	})

// revivalExperimentLife is "you lose 3 life for each card returned this
// way", told the cards that entered.
var revivalExperimentLife = effects.TogetherTail("together/revival-experiment-life", func(ctx *effects.Context, entered []uuid.UUID, _ int) error {
	if len(entered) == 0 {
		return nil
	}
	return effects.GainLife{Player: ctx.Controller(), Amount: -3 * len(entered)}.Apply(ctx)
})
