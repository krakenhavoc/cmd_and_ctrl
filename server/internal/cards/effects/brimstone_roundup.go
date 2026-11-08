package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brimstone Roundup — Enchantment {1}{R} (Outlaws of Thunder Junction,
// #4468):
//
//	"Whenever you cast your second spell each turn, create a 1/1 red
//	 Mercenary creature token with '{T}: Target creature you control gets
//	 +1/+0 until end of turn. Activate only as a sorcery.'
//	 Plot {2}{R} (You may pay {2}{R} and exile this card from your hand.
//	 Cast it as a sorcery on a later turn without paying its mana cost.
//	 Plot only as a sorcery.)"
//
// The trigger is Cosmogrand Zenith's second-spell read
// (YouCastYourSecondSpellEachTurn, the per-turn cast tally), and the
// token is the set's Mercenary. Casting a plotted Roundup is a spell
// cast, so it counts toward the tally like any other. Plot is the
// ordinary special action (CR 702.170).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "e79deeee-c14d-496d-9b0b-475a6a24274c",
		Name:           "Brimstone Roundup",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{2}{R}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Brimstone Roundup — create a 1/1 red Mercenary (second spell this turn)", createAMercenary()),
		},
	})
}
