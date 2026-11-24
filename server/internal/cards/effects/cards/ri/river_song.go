package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// River Song — Legendary Creature — Human Time Lord Rogue {1}{U}{R}, 2/2
// (S82 GAP2, #5948, #6089):
//
//	"Meet in Reverse — You draw cards from the bottom of your library
//	 rather than the top.
//	 Spoilers — Whenever an opponent scries, surveils, or searches their
//	 library, put a +1/+1 counter on River Song. Then River Song deals
//	 damage to that player equal to its power."
//
// Meet in Reverse is the engine's Spec.DrawsFromBottom
// (game/draw_from_bottom.go): every draw River Song's controller makes,
// a draw step's or an effect's, takes the bottom card of their library;
// draw replacements still see it, because it is still a draw.
//
// Spoilers is one triggered ability over three events (EventScry,
// EventSurveil, EventSearchLibrary): a completed scry or surveil by an
// opponent, or an opponent's search of their OWN library (an opponent
// searching your library, Bribery, is not "their library";
// AnOpponentSearchesTheirOwnLibrary). As it resolves, the counter goes on
// River Song if it is still the same permanent, and then River Song deals
// damage equal to its power, read after the counter, to that player; a
// River Song that has left deals damage equal to its power as it last
// existed (CR 608.2h). A power of 0 or less deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "418ecbb7-b809-4639-a531-070b58b5b85e",
		Name:            "River Song",
		Completeness:    effects.CompletenessFull,
		DrawsFromBottom: true,
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventScry, game.EventSurveil, game.EventSearchLibrary}, riverSongSpoilersWhen,
				"River Song — Spoilers: put a +1/+1 counter on River Song, then it deals damage to that player equal to its power",
				riverSongSpoilers),
		},
	})
}

// riverSongSpoilersWhen is "whenever an opponent scries, surveils, or
// searches their library".
func riverSongSpoilersWhen(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Actor == uuid.Nil || ev.Actor == source.Controller {
		return false
	}
	switch ev.Kind {
	case game.EventScry, game.EventSurveil:
		return true
	case game.EventSearchLibrary:
		return effects.AnOpponentSearchesTheirOwnLibrary(ev, source, lki, g)
	}
	return false
}

// riverSongSpoilers puts the counter on, then deals the damage.
func riverSongSpoilers(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	src := ctx.Source()
	victim := item.Trigger.Event.Actor
	if !ctx.SourceLeftBattlefield(src) && !ctx.IsNewSourceObject(src) {
		if _, err := g.AddCounterMustSettleNowForEffect(src, game.CounterPlusOne, 1); err != nil {
			return err
		}
	}
	info, ok := ctx.SourcePermanent()
	if !ok || info.Power <= 0 || victim == uuid.Nil {
		return nil
	}
	d := effects.DealDamage{Source: src, Target: victim, Amount: info.Power}
	if ref, ok := ctx.SourceRef(); ok {
		d.SourceObject = &ref
	}
	return d.Apply(ctx)
}
