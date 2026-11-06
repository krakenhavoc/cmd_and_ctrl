package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ashiok's Reaper — Creature — Nightmare {3}{B}, 3/3 (Wilds of Eldraine,
// #4471):
//
//	"Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, draw a card."
//
// Starfield Mystic's trigger (socB13EnchantmentYouControlWasPutIntoAGraveyard)
// with a draw for its payout. The enchantment is read as it last existed on
// the battlefield and under whose control, so an Aura that falls off with
// its creature counts, a token enchantment (a Role) counts, and an
// enchantment an opponent controlled does not. The draw does not need the
// Reaper to still be on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f1e5571-ddda-4cea-84ef-36a571d8fd51",
		Name:         "Ashiok's Reaper",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Ashiok's Reaper — draw a card",
				Do(DrawCards{N: 1})),
		},
	})
}
