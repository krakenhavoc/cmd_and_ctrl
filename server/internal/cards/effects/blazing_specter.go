package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blazing Specter — Creature — Specter {2}{B}{R}, 2/2 (Invasion, #5616):
//
//	"Flying, haste
//	 Whenever this creature deals combat damage to a player, that
//	 player discards a card."
//
// Flying and haste are printed keywords. "That player" is the player
// dealt the damage (the trigger event's target), and they choose the
// card (CR 701.9b: a discard nobody says is random is the discarding
// player's choice), so the payoff is Entropic Specter's discardOne
// prompt, which asks nothing of an empty hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c046368-5899-42ef-9972-3a3a6c8fe3fd",
		Name:            "Blazing Specter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Blazing Specter — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					return discardOne(g, item, item.Trigger.Event.Target)
				}),
		},
	})
}
