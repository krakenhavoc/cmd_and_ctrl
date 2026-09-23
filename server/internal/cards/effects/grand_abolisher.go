package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grand Abolisher — Creature — Human Cleric {W}{W}, 2/2
// (EDHREC rank 250):
//
//	"During your turn, your opponents can't cast spells or activate
//	 abilities of artifacts, creatures, or enchantments."
//
// Two turn-scoped restriction statics, not one:
//
//   - The cast half is Dragonlord Dromoka's own clause,
//     OpponentsCantCastDuringYourTurn(label, nil) — nil because Grand
//     Abolisher restricts every spell, not a narrowed subset.
//   - The activation half needed a turn-scoped sibling of
//     OpponentsSourcesCantActivate (#1210), added alongside this card
//     in activation_restriction.go: OpponentsSourcesCantActivateDuringYourTurn.
//     No mana exemption — the printed text names artifacts, creatures
//     and enchantments with no "unless they're mana abilities"
//     carve-out, so a Birds of Paradise an opponent controls is
//     silenced on the controller's turn like everything else the
//     clause names.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c749f23c-40c0-4159-b84c-a70cbb062c14",
		Name:         "Grand Abolisher",
		Completeness: CompletenessFull,
		CastRestrictions: []game.CastRestriction{
			OpponentsCantCastDuringYourTurn(
				"During your turn, your opponents can't cast spells.", nil),
		},
		ActivationRestrictions: []game.ActivationRestriction{
			OpponentsSourcesCantActivateDuringYourTurn(
				"During your turn, your opponents can't activate abilities of artifacts, creatures, or enchantments.",
				Or(Artifact(), Creature(), Enchantment()), false),
		},
	})
}
