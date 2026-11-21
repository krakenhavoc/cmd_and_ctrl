package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magik, Illyana Rasputin // Darkchild, Demon Uncaged — modal double-faced
// (Marvel's Spider-Man, #4460), both faces legendary creatures:
//
//	Magik, Illyana Rasputin {1}{B}{R}, 3/3:
//	 "At the beginning of your upkeep, exile the top card of your library.
//	  You may play that card this turn.
//	  At the beginning of your end step, if you've sacrificed two or more
//	  permanents this turn, you may transform Magik."
//	Darkchild, Demon Uncaged {3}{B}{R}, 4/4:
//	 "At the beginning of your upkeep, exile cards from the top of your
//	  library until you exile a nonland card. You may play those cards this
//	  turn.
//	  Pay 4 life: Darkchild gains indestructible until end of turn. Tap
//	  her."
//
// The printed layout is modal_dfc and the card carries the transform
// keyword, so the engine's CanTransform (a transform or modal_dfc layout)
// turns it over in place. The end-step trigger checks "two or more
// permanents sacrificed this turn" twice, as the trigger event is read and
// again as it resolves (CR 603.4): the player's PermanentsSacrificed tally.
// "Those cards" is every card the walk exiled, lands included, each playable
// this turn. Darkchild's life payment and tap follow the printed order: the
// indestructible grant, then the tap.
func init() {
	effects.Register(effects.Spec{
		OracleID:     MagikOracle,
		Name:         "Magik, Illyana Rasputin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Magik — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 1,
					}.Apply(effects.NewContext(g, item))
				}),
			effects.Optional(
				effects.InterveningIf(magikSacrificedTwo, effects.On(game.EventBeginEndStep,
					effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return g.TurnTallyFor(source.Controller).PermanentsSacrificed >= 2
					}),
					"Magik — you may transform Magik",
					effects.Do(effects.TransformThis{}))),
				"Transform Magik?"),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     MagikOracle + "#1",
		Name:         "Darkchild, Demon Uncaged",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Darkchild — exile cards from the top of your library until you exile a nonland card; you may play those cards this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    effects.UntilNonland,
						ThenTail: ctx.ItemTail(darkchildExiled),
					}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Pay 4 life: Darkchild gains indestructible until end of turn. Tap her.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.PayLife(4),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Darkchild — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return effects.TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}

const MagikOracle = "72d85d61-6bc3-4028-97c0-d5c34204a1df"

// magikSacrificedTwo is Magik's intervening "if you've sacrificed two or
// more permanents this turn", read as the ability resolves.
var magikSacrificedTwo = effects.CondBoard{
	Name: "you've sacrificed two or more permanents this turn",
	Test: func(ctx *effects.Context) bool {
		return ctx.Game.TurnTallyFor(ctx.Controller()).PermanentsSacrificed >= 2
	},
}

// darkchildExiled grants a play permission, for this turn, over every card
// the walk exiled and that is still in exile.
var darkchildExiled = effects.CardTail("darkchild/play-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	var cards []game.Card
	for _, id := range o.IDs {
		c, ok := ctx.Game.LookupCardForEffect(id)
		z := ctx.Game.FindCardZoneForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneExile {
			continue
		}
		cards = append(cards, c)
	}
	if len(cards) == 0 {
		return nil
	}
	perm := effects.ExileTopWithPermission{}.Permission()
	perm.Player = ctx.Controller()
	perm.Zone = game.ZoneExile
	ctx.Game.GrantCastPermissionToCardsForEffect(perm, cards)
	return nil
})
