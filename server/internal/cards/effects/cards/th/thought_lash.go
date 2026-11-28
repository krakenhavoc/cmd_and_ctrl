package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thought Lash — Enchantment {2}{U}{U} (Alliances, S82 OLD batch 818):
//
//	"Cumulative upkeep—Exile the top card of your library.
//	 When a player doesn't pay this enchantment's cumulative upkeep, that
//	 player exiles all cards from their library.
//	 Exile the top card of your library: Prevent the next 1 damage that would
//	 be dealt to you this turn."
//
// The "doesn't pay" trigger fires before the sacrifice (CR 702.24a) and
// names the player who was asked; with a library too short to pay, the
// player has not paid and loses it all.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "81fe7c5e-503e-488c-b06c-1615ed38e5a3",
		Name:         "Thought Lash",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Thought Lash — cumulative upkeep: exile the top card of your library", effects.ExileLibraryTopPayment(1), "", ""),
			effects.WhenCumulativeUpkeepIsNotPaid("Thought Lash — that player exiles all cards from their library", effects.OldB818ThatPlayerExilesTheirLibrary),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Exile the top card of your library: Prevent the next 1 damage that would be dealt to you this turn.",
			Purpose: effects.PreventsDamage(),
			Cost:    effects.ExileTopOfLibrary(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventNextDamage{Target: item.Controller, Amount: 1, Label: "Thought Lash: prevent the next 1 damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
