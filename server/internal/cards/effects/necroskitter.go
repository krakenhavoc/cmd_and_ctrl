package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necroskitter — Creature — Elemental {1}{B}{B}, 1/4:
//
//		"Wither (This deals damage to creatures in the form of -1/-1
//		 counters.)
//		 Whenever a creature an opponent controls with a -1/-1 counter on
//		 it dies, you may return that card to the battlefield under your
//		 control."
//
//	  - Wither is the engine's damage tail (game/infect_wither_toxic.go).
//	  - The condition looks back in time (CR 603.10a). "Had a -1/-1
//	    counter" is read from the counters the exit recorded
//	    (LastKnownCountersForEffect), because the card in the graveyard
//	    has none. "An opponent controls" is the controller the exit
//	    stamped on the event (leftUnderControlOf), never the dead card's
//	    own field: a creature of mine that an opponent stole and that died
//	    under their control counts, one I stole from them does not.
//	  - Only an opponent's creature qualifies, so the Skitter's owner
//	    gets nothing from their own dying creatures. Tokens qualify for
//	    the trigger but cease to exist in the graveyard, so there is
//	    nothing to return (CR 111.7).
//	  - "You may" is a choice made when the ability RESOLVES (MayChoice),
//	    not before it goes on the stack (#2494). The card is not
//	    targeted: it is whichever creature died, and it comes back only
//	    while it is still the graveyard object that arrived from the
//	    battlefield (CR 400.7), checked when the ability resolves and
//	    again once the player says yes. It enters under the controller of
//	    the ability, not its owner.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "536f7c92-2f1e-49e7-98de-47382a9488f1",
		Name:            "Necroskitter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordWither},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, necroskitterCondition,
				"Necroskitter — you may return that card to the battlefield under your control", necroskitterMayReturn),
		},
	})
}

// necroskitterCondition is "a creature an opponent controls with a
// -1/-1 counter on it dies".
func necroskitterCondition(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	if controller := leftUnderControlOf(ev, dead); controller == source.Controller || g.PlayerByIDForEffect(controller) == nil {
		return false
	}
	return g.LastKnownCountersForEffect(ev.CardID)[game.CounterMinusOne] > 0
}

func necroskitterMayReturn(g *game.Game, item *game.StackItem) error {
	dead := item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(NewContext(g, item), dead) {
		return nil
	}
	return MayChoice{
		Question: "Necroskitter — return that card to the battlefield under your control?",
		Then:     necroskitterAnswer,
	}.Apply(NewContext(g, item))
}

// necroskitterAnswer is the "yes": return the card that died (the asking
// item's triggering event) to the battlefield under your control. Keyed
// (#4291).
var necroskitterAnswer = MayYes("confirm/necroskitter-return", func(ctx *Context, _ game.ConfirmAnswer) error {
	if ctx.Item.Trigger == nil {
		return nil
	}
	dead := ctx.Item.Trigger.Event.CardID
	if !diedCardStillInGraveyard(ctx, dead) {
		return nil
	}
	return ReturnFromGraveyard{Target: dead, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
})
