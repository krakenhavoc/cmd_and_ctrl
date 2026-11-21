package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cartographer's Hawk — Creature — Bird {1}{W}, 2/1 (Commander Legends:
// Battle for Baldur's Gate):
//
//	"Flying
//	 When this creature deals combat damage to a player who controls more lands
//	 than you, return it to its owner's hand. If you do, you may search your
//	 library for a Plains card, put it onto the battlefield tapped, then
//	 shuffle."
//
// The damaged player is the event's target, and "controls more lands than you"
// is an intervening if (CR 603.4): read as the trigger goes on the stack and
// again as it resolves, on that player's lands against the Hawk's controller's
// now. The return is BounceToHand with the search as its "if you do" tail, run
// only if the Hawk really reached a hand (CR 400.7), with Kor Cartographer's
// optional search for a Plains card, tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ac8cf699-f67b-44da-bb79-eb95cfc0026d",
		Name:            "Cartographer's Hawk",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			cartographersHawkTrigger(),
		},
	})
}

// cartographersHawkTrigger is the combat damage trigger with its intervening
// "if" (CR 603.4) read twice: the trigger only goes on the stack if the damaged
// player controls more lands than the Hawk's controller (the AppliesTo), and
// the condition is checked again as it resolves (InterveningIf).
func cartographersHawkTrigger() game.TriggeredAbility {
	t := effects.InterveningIf(effects.CondBoard{
		Name: "the damaged player controls more lands than you",
		Test: func(ctx *effects.Context) bool {
			return ctx.Item != nil && ctx.Item.Trigger != nil &&
				gapB108ControlsMoreLands(ctx.Game, ctx.Item.Trigger.Event.Target, ctx.Controller())
		},
	}, effects.WheneverThisDealsCombatDamageToAPlayer("Cartographer's Hawk — return it to its owner's hand; if you do, you may search for a Plains card",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			if !effects.OnBattlefield(g, item.SourceCardID) || ctx.IsNewSourceObject(item.SourceCardID) {
				return nil
			}
			return effects.BounceToHand{Target: item.SourceCardID, ThenTail: ctx.ItemTail(cartographersHawkSearch)}.Apply(ctx)
		}))
	connects := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return connects(ev, source, lki, g) && gapB108ControlsMoreLands(g, ev.Target, source.Controller)
	}
	return t
}

// gapB108ControlsMoreLands reports whether `player` controls more lands than
// `than`.
func gapB108ControlsMoreLands(g *game.Game, player, than uuid.UUID) bool {
	theirs, mine := 0, 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if !c.IsLand() {
			continue
		}
		switch c.Controller {
		case player:
			theirs++
		case than:
			mine++
		}
	}
	return theirs > mine
}

// cartographersHawkSearch is "If you do, you may search your library for a
// Plains card, put it onto the battlefield tapped, then shuffle", run only when
// the Hawk reached a hand. An on-disk key: never renamed, never reused.
var cartographersHawkSearch = effects.CardTail("cartographers-hawk/search-for-a-plains", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.OneWentThisWay(o) {
		return nil
	}
	return effects.SearchLibrary{
		Player:        ctx.Controller(),
		Predicate:     effects.B39IsPlainsCard,
		Dest:          game.ZoneBattlefield,
		Limit:         1,
		TappedOnEntry: true,
		Optional:      true,
		Shuffle:       true,
		Source:        ctx.Source(),
		Reason:        "Cartographer's Hawk — search for a Plains card",
	}.Apply(ctx)
})
