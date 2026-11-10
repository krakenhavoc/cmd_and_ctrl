package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redtooth Vanguard — Creature — Elf Warrior {1}{G}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"Trample
//	 Whenever an enchantment you control enters, you may pay {2}. If you
//	 do, return this card from your graveyard to your hand."
//
// A graveyard ability (InGraveyard: it works there and nowhere else, "you"
// being the card's owner). The condition is eerie's enchantment half
// (Rimefur Reindeer's EnchantmentEnteredUnderYourControl, a Role token
// included). The {2} is the ordinary "you may pay" prompt, and the card
// comes back only if it is still in the graveyard once paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9e459e48-d2bb-4c4e-9a87-58efb56cddd1",
		Name:            "Redtooth Vanguard",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventETB, EnchantmentEnteredUnderYourControl,
				"Redtooth Vanguard — you may pay {2}; if you do, return this card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return mayPayForGraveyardReturnToHand(g, item, "{2}",
						"Redtooth Vanguard — pay {2} to return it from your graveyard to your hand?")
				})),
		},
	})
}
