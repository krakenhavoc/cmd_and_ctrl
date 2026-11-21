package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brinelin, the Moon Kraken — Legendary Creature — Kraken {6}{U}{U}, 6/8
// (Commander Legends: Battle for Baldur's Gate):
//
//	"When Brinelin enters and whenever you cast a spell with mana value 6 or
//	 greater, you may return target nonland permanent to its owner's hand.
//	 Partner (You can have two commanders if both have partner.)"
//
// One ability with two conditions (Hraesvelgr's shape): it enters, or a spell
// of yours is cast whose mana value, as cast and X included (CR 202.3e, read
// off the cast event), is six or more. The "may" is asked as it resolves
// (Optional), after the target is chosen. Partner is a deck-building rule
// (CR 702.124) that internal/deck reads off the oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f00edd6-4321-48a1-9053-77c77b922aeb",
		Name:         "Brinelin, the Moon Kraken",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(Optional(OnAny([]game.EventKind{game.EventETB, game.EventCast},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					if ev.Kind == game.EventETB {
						return ev.CardID == source.InstanceID
					}
					return ev.Kind == game.EventCast && ev.Actor == source.Controller && ev.Amount >= 6
				},
				"Brinelin, the Moon Kraken — you may return target nonland permanent to its owner's hand",
				bounceTheTargetToHand), "Brinelin, the Moon Kraken — return the target to its owner's hand?"),
				TargetPermanent("target nonland permanent", Nonland())), game.Purpose{Answers: game.AnswerRemove}),
		},
	})
}
