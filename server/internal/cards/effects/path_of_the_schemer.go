package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Path of the Schemer — Sorcery {4}{B} (March of the Machine Commander,
// #5648):
//
//	"Each player mills two cards. Then you put a creature card from a
//	 graveyard onto the battlefield under your control. It's an artifact
//	 in addition to its other types.
//	 Will of the Planeswalkers — Starting with you, each player votes for
//	 planeswalk or chaos. If planeswalk gets more votes, planeswalk. If
//	 chaos gets more votes or the vote is tied, chaos ensues."
//
// Extract from Darkness with an artifact rider. The creature card is
// chosen after the mill, from every graveyard, and is not a target, so a
// card the spell milled is a legal pick (CR 608.2c). It enters as an
// artifact in addition to its other types (ADR 0210, game.EntersAs: a
// layer-4 addition that lasts while it is that permanent), so "whenever
// an artifact enters" sees it. With no creature card in any graveyard
// nothing enters. Then the vote (WillOfThePlaneswalkers), after the pick
// and anything the entry asks; it happens at every table, and
// planeswalking and chaos do something only at a table playing
// Planechase.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9cdb1cf6-2bf1-401b-83ae-6f3f7512e5a6",
		Name:         "Path of the Schemer",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return b02EachPlayerMills(ctx.Game, ctx.Item, 2)
			}),
			Instruction(func(ctx *Context) error {
				item := ctx.Item
				candidates := allGraveyardsCardIDs(ctx, game.Card.IsCreature)
				if len(candidates) == 0 {
					return nil
				}
				ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
					Chooser:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Path of the Schemer — put a creature card from a graveyard onto the battlefield under your control (it's also an artifact)",
					Cards:    candidates,
					Min:      1,
					Max:      1,
					ThenKey:  pathOfTheSchemerPicked,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			}),
			WillOfThePlaneswalkers{Card: "Path of the Schemer"},
		),
	})
}

// pathOfTheSchemerPicked is the chosen creature card entering under the
// caster's control as an artifact too. An on-disk key: never renamed,
// never reused.
var pathOfTheSchemerPicked = askedPick("choose-cards/path-of-the-schemer-reanimate", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{r.Picked[0]},
		Controller: ctx.Controller(),
		EntersAs: &game.EntersAs{
			AddTypes: []string{"Artifact"},
			Label:    "Path of the Schemer — an artifact in addition to its other types",
		},
	}.Apply(ctx)
})
