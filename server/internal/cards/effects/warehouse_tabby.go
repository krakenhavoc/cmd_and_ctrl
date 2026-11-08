package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warehouse Tabby — Creature — Cat {B}, 1/1 (Wilds of Eldraine, #4471):
//
//	"Whenever an enchantment you control is put into a graveyard from the
//	 battlefield, create a 1/1 black Rat creature token with "This token
//	 can't block."
//	 {1}{B}: This creature gains deathtouch until end of turn."
//
// The trigger is Knight of Doves' (socB13EnchantmentYouControlWasPutIntoAGraveyard):
// the enchantment is read as it last existed on the battlefield and under
// whose control (CR 603.10a), so an Aura that falls off its creature and a
// Role token both count, and an opponent's enchantment does not. The payout
// is the catalog's can't-block Rat. The deathtouch is Endling's grant until
// end of turn, on this object only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7736614-cdd3-43bd-ab3e-461e3e24ed40",
		Name:         "Warehouse Tabby",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, socB13EnchantmentYouControlWasPutIntoAGraveyard,
				"Warehouse Tabby — create a 1/1 black Rat creature token with \"This token can't block.\"",
				Do(CreateToken{Template: CantBlockRatToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}: This creature gains deathtouch until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{1}{B}"),
			Effect:  thisGainsKeywordUntilEndOfTurn("deathtouch", "Warehouse Tabby — deathtouch until end of turn"),
		}},
	})
}
