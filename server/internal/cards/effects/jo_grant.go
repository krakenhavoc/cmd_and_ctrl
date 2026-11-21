package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "65ea3638-1766-4096-88a8-662a0d6faf0c",
		Name:         "Jo Grant",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:       "jo-grant/cycling",
			Zone:      game.ZoneHand,
			Activated: []ActivatedAbility{Cycling("{2}{W}")},
			Text:      "Cycling {2}{W}",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneHand, CardsYouOwn(func(card *game.Card) bool { return b09IsHistoric(*card) }), "jo-grant/cycling"),
				"Each historic card in your hand has cycling {2}{W}."),
		},
		Triggered: []game.TriggeredAbility{
			WheneverYouCycle("Jo Grant — put a +1/+1 counter on Jo Grant", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
