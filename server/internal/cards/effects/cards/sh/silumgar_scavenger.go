package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Silumgar Scavenger — Creature — Zombie Bird {4}{B}, 2/3 (Modern Horizons):
//
//	"Flying
//	 Exploit (When this creature enters, you may sacrifice a creature.)
//	 Whenever another creature you control dies, put a +1/+1 counter on this
//	 creature. It gains haste until end of turn if it exploited that
//	 creature."
//
// Flying is a printed keyword and exploit the shared constructor. The dies
// trigger looks back at the creature that died (CR 603.10a), and puts the
// counter on the Scavenger if it is still the permanent that triggered. "If it
// exploited that creature" is read as the ability resolves off the turn's
// events: an exploited event (CR 702.110b) whose exploiter is this Scavenger
// and whose sacrificed creature is the one that died. A creature sacrificed to
// another creature's exploit, or that died any other way, gives no haste, and
// a Scavenger that exploited itself never triggers (it is not "another").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ce068b46-9301-42fb-9db1-9c0da5275d89",
		Name:            "Silumgar Scavenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Exploit("Silumgar Scavenger", game.BodyRef{}),
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Silumgar Scavenger — put a +1/+1 counter on it; it gains haste until end of turn if it exploited that creature",
				silumgarScavengerDies),
		},
	})
}

func silumgarScavengerDies(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	if !silumgarScavengerExploited(g, item.SourceCardID, ctx.Trigger().Event.CardID) {
		return nil
	}
	return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"haste"},
		Label: "Silumgar Scavenger — haste until end of turn"}.Apply(ctx)
}

// silumgarScavengerExploited is whether `exploiter` exploited `dead` this turn.
func silumgarScavengerExploited(g *game.Game, exploiter, dead uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventExploited && ev.CardID == exploiter && ev.Target == dead {
			return true
		}
	}
	return false
}
