package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nashi, Searcher in the Dark — Legendary Creature — Rat Ninja Wizard
// {U}{B}, 2/2 (Duskmourn: House of Horror, #4466):
//
//	"Menace
//	 Whenever Nashi deals combat damage to a player, you mill that many
//	 cards. You may put any number of legendary and/or enchantment cards
//	 from among them into your hand. If you put no cards into your hand
//	 this way, put a +1/+1 counter on Nashi."
//
// Barrowgoyf's trigger: the damage dealt rides the stack item, the mill is
// MillToZone with a keyed tail told what landed in the graveyard, and the
// pick is a free-count choice over the legendary and enchantment cards
// among them still there (any number, including none). The counter goes on
// when nothing reached the hand, whether the player declined, none of the
// cards qualified or a replacement kept them out of it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ade8fc6f-efaa-4a24-96d8-99ef5026cd74",
		Name:            "Nashi, Searcher in the Dark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventDealDamage},
			AppliesTo: ThisDealtCombatDamageToAPlayer,
			Key:       "Nashi, Searcher in the Dark — mill that many cards, then put legendary and/or enchantment cards from among them into your hand",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Nashi, Searcher in the Dark — mill that many cards")
				item.Params.Amount = ev.Amount
				return item
			},
			Effect: stdB26NashiMill,
		}},
	})
}
