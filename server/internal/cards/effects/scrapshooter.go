package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scrapshooter — Creature — Raccoon Archer {1}{G}{G}, 4/4 (EDHREC rank 6759):
//
//	"Gift a card (You may promise an opponent a gift as you cast this
//	 spell. If you do, when it enters, they draw a card.)
//	 Reach
//	 When this creature enters, if the gift was promised, destroy
//	 target artifact or enchantment an opponent controls."
//
// The permanent half of gift (CR 702.174b, ADR 0089). On a permanent
// the gift is not a spell ability but an entry trigger — "when this
// permanent enters, if its gift cost was paid, [effect]" — which
// `Gift: GiftACard()` grows for the card. The card's own trigger is a
// second, separate ability with the same intervening if (CR 603.4),
// read off the permanent's cast record (CR 400.7d): a Scrapshooter
// cast without the promise, reanimated, or copied as a token puts
// neither trigger on the stack.
//
// Both triggers go on the stack when it enters, controller's choice of
// order; the opponent's draw and the Naturalize are independent, so
// either order is the printed card.
func init() {
	Register(Spec{
		OracleID:        "235e3231-c5e5-4696-a867-705b2c4158fe",
		Name:            "Scrapshooter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Gift:            GiftACard(),
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID && source.GiftPromised()
			},
			Targets: TargetPermanent("target artifact or enchantment an opponent controls",
				Or(Artifact(), Enchantment()), OpponentControls()),
			Key:    "Scrapshooter — destroy target artifact or enchantment an opponent controls",
			Effect: destroyChosenPermanent,
		}},
	})
}
