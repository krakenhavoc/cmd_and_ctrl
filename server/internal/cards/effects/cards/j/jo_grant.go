package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jo Grant — Legendary Creature — Human Soldier {2}{W}, 3/2 (Doctor Who,
// #5948):
//
//	"Each historic card in your hand has cycling {2}{W}. ({2}{W}, Discard that
//	 card: Draw a card.)
//	 Whenever you cycle a card, put a +1/+1 counter on Jo Grant.
//	 Doctor's companion"
//
// The first line is a grant of cycling to each historic card the controller
// owns in their hand (Tectonic Reformation's shape, read as the card is in the
// hand, CR 700.6). Any cycling, granted or printed, triggers the second
// ability (Cycling). Doctor's companion is a deck-construction rule.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65ea3638-1766-4096-88a8-662a0d6faf0c",
		Name:         "Jo Grant",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       "jo-grant/cycling",
			Zone:      game.ZoneHand,
			Activated: []effects.ActivatedAbility{effects.Cycling("{2}{W}")},
			Text:      "Cycling {2}{W}",
		}},
		Static: []game.StaticAbility{
			effects.Labelled(effects.GrantAbilitiesInZone(game.ZoneHand, effects.CardsYouOwn(func(card *game.Card) bool { return effects.B09IsHistoric(*card) }), "jo-grant/cycling"),
				"Each historic card in your hand has cycling {2}{W}."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCycle("Jo Grant — put a +1/+1 counter on Jo Grant", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
