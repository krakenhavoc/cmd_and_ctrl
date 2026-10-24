package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ziatora's Envoy — Creature — Lizard Warrior {1}{B}{R}{G}, 5/4:
//
//	"Trample
//	 Whenever this creature deals combat damage to a player, look at the
//	 top card of your library. You may play a land from the top of your
//	 library or cast a spell with mana value less than or equal to the
//	 damage dealt from the top of your library without paying its mana
//	 cost. If you don't, put that card into your hand.
//	 Blitz {2}{B}{R}{G}"
//
// Blitz (CR 702.152, #3772). The trigger is
// Game.OfferFreePlayFromTopOrHandForEffect (#3801, ADR 0099's 2026-10-15
// amendment): the look is private, a land the player can play now is a
// land play, a spell within the cap gets a free cast from the library made
// inside the trigger's resolution (a resolution_cast prompt, CR 608.2g,
// #4250), and every other outcome puts the card into the hand. The cap is
// "the damage dealt", read off the trigger's event, so a hit that was
// partly prevented counts what was actually dealt.
func init() {
	Register(Spec{
		OracleID:         "93fe1d5b-3ab3-431a-9d78-175550ba9d43",
		Name:             "Ziatora's Envoy",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		AlternativeCosts: []game.AlternativeCost{Blitz("{2}{B}{R}{G}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Ziatora's Envoy — look at the top card, play it free or put it into your hand",
				func(g *game.Game, item *game.StackItem) error {
					return g.OfferFreePlayFromTopOrHandForEffect(game.FreePlayFromTop{
						Player:       item.Controller,
						Source:       item.SourceCardID,
						SourceName:   "Ziatora's Envoy",
						MaxManaValue: thatDamage(item),
						Land:         true,
					})
				}),
		},
	})
}
