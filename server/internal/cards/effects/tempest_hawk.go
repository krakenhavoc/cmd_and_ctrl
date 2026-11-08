package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tempest Hawk — Creature — Bird {2}{W}, 2/2 (Tarkir: Dragonstorm, #4463):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, you may
//	 search your library for a card named Tempest Hawk, reveal it, put it
//	 into your hand, then shuffle.
//	 A deck can have any number of cards named Tempest Hawk."
//
// Flying is a printed keyword. The combat damage trigger is a CR 603.5
// "you may": declining searches nothing and does not shuffle; agreeing
// searches for a card named Tempest Hawk (game.HasName, Shadowborn
// Apostle's match), reveals it and puts it into your hand, then shuffles.
// The deck-construction line needs nothing here: the deck validator reads
// "A deck can have any number of cards named …" from the oracle text (CR
// 113.6n).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7423b3b9-56eb-4cf2-8ada-135918219c4b",
		Name:            "Tempest Hawk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(WheneverThisDealsCombatDamageToAPlayer("Tempest Hawk — search your library for a card named Tempest Hawk",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Tempest Hawk") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Tempest Hawk — a card named Tempest Hawk",
					}.Apply(NewContext(g, item))
				}),
				"Search your library for a card named Tempest Hawk?"),
		},
	})
}
