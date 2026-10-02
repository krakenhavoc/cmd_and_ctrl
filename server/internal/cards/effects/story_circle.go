package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Story Circle — Enchantment {1}{W}{W}:
//
//	"As this enchantment enters, choose a color.
//	 {W}: The next time a source of your choice of the chosen color would deal damage to you this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as the ability resolves (CR 615.8, 609.7a),
// offered only among sources of the colour chosen as the enchantment
// entered (CR 105.4) and checked again when the source would deal the
// damage (CR 615.9). The colour is read as Story Circle last existed if
// it has left by then (CR 608.2h).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "7071aee8-b5ca-4be5-9ba0-2df7e3af303b",
		Name:         "Story Circle",
		Completeness: CompletenessFull,
		AsEnters:     ChooseColorAsEnters(game.ColorForProtection, "Story Circle"),
		Activated: []ActivatedAbility{chosenColorShieldRow(
			"{W}: The next time a source of your choice of the chosen color would deal damage to you this turn, prevent that damage.", ManaCost("{W}"))},
	})
}
