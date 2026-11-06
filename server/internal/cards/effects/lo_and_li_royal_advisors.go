package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lo and Li, Royal Advisors — Legendary Creature — Human Advisor {2}{B}{B},
// 3/3 (Avatar: The Last Airbender Eternal, #5075):
//
//	"Whenever an opponent discards a card or mills one or more cards, put a
//	 +1/+1 counter on each Advisor you control.
//	 {2}{U/B}: Target player mills four cards."
//
// One printed trigger, one row, watching both events. A discard triggers
// once per card discarded; a mill is "one or more cards", so it fires once
// per milling player per event batch (CR 603.2c, Zellix's
// nghaOncePerMiller shape), and the batch key keeps each discard its own
// occurrence. A milled card a replacement sends elsewhere is not milled
// into a graveyard and fires nothing. The Advisors are read as the trigger
// resolves, Lo and Li included. The activated ability is a plain
// target-player mill.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5018037-cf4e-46d2-86c3-fb8eec6e44d1",
		Name:         "Lo and Li, Royal Advisors",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:      []game.EventKind{game.EventDiscardCard, game.EventMill},
			Key:          "Lo and Li, Royal Advisors — put a +1/+1 counter on each Advisor you control",
			AppliesTo:    avB30aLoAndLiOpponentDiscardedOrMilled,
			OncePerBatch: true,
			BatchKey:     avB30aLoAndLiBatchKey,
			Effect:       avB30aPutCounterOnEachOfSubtypeYouControl("Advisor"),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U/B}: Target player mills four cards.",
			Cost:    ManaCost("{2}{U/B}"),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(4),
		}},
	})
}

// avB30aLoAndLiOpponentDiscardedOrMilled is the trigger condition: an
// opponent discarded a card, or milled a card into a graveyard.
func avB30aLoAndLiOpponentDiscardedOrMilled(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	if ev.Kind == game.EventMill {
		return ev.NewZone == game.ZoneGraveyard && ev.Actor != uuid.Nil && ev.Actor != source.Controller
	}
	return b18OpponentDiscarded(ev, source)
}

// avB30aLoAndLiBatchKey collapses a mill batch per milling player and keeps
// every discarded card its own occurrence.
func avB30aLoAndLiBatchKey(ev game.Event, _ *game.Card, _ *game.Game) string {
	if ev.Kind == game.EventMill {
		return "mill|" + ev.Actor.String()
	}
	return "discard|" + ev.CardID.String()
}
