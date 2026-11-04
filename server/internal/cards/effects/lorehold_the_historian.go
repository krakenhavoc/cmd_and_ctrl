package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lorehold, the Historian — {3}{R}{W} Legendary Creature — Elder Dragon
// 5/5 (Secrets of Strixhaven, #4685):
//
//	"Flying, haste
//	 Each instant and sorcery card in your hand has miracle {2}. (You may
//	 cast a card for its miracle cost when you draw it if it's the first
//	 card you drew this turn.)
//	 At the beginning of each opponent's upkeep, you may discard a card. If
//	 you do, draw a card."
//
// The miracle is granted to the cards in its controller's hand (CR
// 702.94a) in the printed shape (game/miracle.go): the static is a granted
// alternative cost over instant and sorcery cards
// (MiracleForInstantAndSorceryCardsInYourHand), claimable only under the
// hand permission a revealed trigger's resolution grants, and the reveal
// is this permanent's own trigger, grown from the offer. It fires for the
// first card its controller drew this turn, on any player's turn, when that
// card is an instant or sorcery card still in their hand. The price rides
// the trigger, so Lorehold leaving once the card is revealed does not take
// the {2} cast away (CR 113.7a). A card that prints a miracle of its own
// has both (CR 113.2c). The loot is Cool but Rude's rummage, in each
// opponent's upkeep and not in yours.
//
// No simplifications.
const loreholdTheHistorianOracleID = "61a41cf1-60cc-45ba-aa98-493c14e87d9d"

const loreholdTheHistorianLootLabel = "Lorehold, the Historian — at the beginning of each opponent's upkeep, you may discard a card; if you do, draw a card"

func init() {
	Register(Spec{
		OracleID:                loreholdTheHistorianOracleID,
		Name:                    "Lorehold, the Historian",
		Completeness:            CompletenessFull,
		PrintedKeywords:         []string{"flying", "haste"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{MiracleForInstantAndSorceryCardsInYourHand("{2}")},
		Triggered: []game.TriggeredAbility{
			AtEachOpponentsUpkeep(loreholdTheHistorianLootLabel, func(g *game.Game, item *game.StackItem) error {
				return b39MayDiscardThenDraw(1, false, loreholdTheHistorianLootLabel, drawOneIfDiscarded)(NewContext(g, item))
			}),
		},
	})
}
