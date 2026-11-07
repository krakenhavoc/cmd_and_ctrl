package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Knight of Doves — Creature — Human Knight {2}{W}, 1/3 (Wilds of
// Eldraine, #4471):
//
//	"Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, create a 1/1 white Bird creature token with flying."
//
// Ashiok's Reaper's trigger (socB13EnchantmentYouControlWasPutIntoAGraveyard)
// with a Bird for its payout. The enchantment is read as it last existed on
// the battlefield and under whose control (CR 603.10a), so an Aura that
// falls off with its creature counts, a token enchantment (a Role) counts,
// and an enchantment an opponent controlled does not. The Bird is the
// token table's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2e6fbcb-ebbd-40e8-8a01-b675fcacaf8e",
		Name:         "Knight of Doves",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Knight of Doves — create a 1/1 white Bird creature token with flying",
				Do(CreateToken{Template: TokenCard("1/1 white Bird with flying"), N: 1})),
		},
	})
}
