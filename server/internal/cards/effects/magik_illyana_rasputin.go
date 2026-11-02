package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     magikOracle,
		Name:         "Magik, Illyana Rasputin",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Magik — exile the top card of your library; you may play it this turn",
				func(g *game.Game, item *game.StackItem) error {
					return ExileTopWithPermission{
						From: item.Controller, GrantTo: item.Controller, N: 1,
					}.Apply(NewContext(g, item))
				}),
			Optional(
				InterveningIf(magikSacrificedTwo, On(game.EventBeginEndStep,
					AllOf(ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return g.TurnTallyFor(source.Controller).PermanentsSacrificed >= 2
					}),
					"Magik — you may transform Magik",
					Do(TransformThis{}))),
				"Transform Magik?"),
		},
	})
	Register(Spec{
		OracleID:     magikOracle + "#1",
		Name:         "Darkchild, Demon Uncaged",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Darkchild — exile cards from the top of your library until you exile a nonland card; you may play those cards this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return MillToZone{
						Player:   ctx.Controller(),
						To:       game.ZoneExile,
						Until:    UntilNonland,
						ThenTail: ctx.ItemTail(darkchildExiled),
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:   "Pay 4 life: Darkchild gains indestructible until end of turn. Tap her.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    PayLife(4),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Target:   item.SourceCardID,
					Keywords: []string{"indestructible"},
					Label:    "Darkchild — indestructible",
				}).Apply(ctx); err != nil {
					return err
				}
				return TapTarget{Target: item.SourceCardID}.Apply(ctx)
			},
		}},
	})
}

const magikOracle = "72d85d61-6bc3-4028-97c0-d5c34204a1df"

// magikSacrificedTwo is Magik's intervening "if you've sacrificed two or
// more permanents this turn", read as the ability resolves.
var magikSacrificedTwo = CondBoard{
	Name: "you've sacrificed two or more permanents this turn",
	Test: func(ctx *Context) bool {
		return ctx.Game.TurnTallyFor(ctx.Controller()).PermanentsSacrificed >= 2
	},
}

// darkchildExiled grants a play permission, for this turn, over every card
// the walk exiled and that is still in exile.
var darkchildExiled = CardTail("darkchild/play-exiled", func(ctx *Context, o game.TailOutcome) error {
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
	perm := ExileTopWithPermission{}.permission()
	perm.Player = ctx.Controller()
	perm.Zone = game.ZoneExile
	ctx.Game.GrantCastPermissionToCardsForEffect(perm, cards)
	return nil
})
