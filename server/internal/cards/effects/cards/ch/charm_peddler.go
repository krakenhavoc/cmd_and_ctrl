package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Charm Peddler — Creature — Human Spellshaper {W}:
//
//	"{W}, {T}, Discard a card: The next time a source of your choice would deal damage to target creature this turn, prevent that damage."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// The card is discarded as part of the cost.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c425fe26-0027-4319-80a9-a053925651cb",
		Name:         "Charm Peddler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.NextDamageShieldRow(
			"{W}, {T}, Discard a card: The next time a source of your choice would deal damage to target creature this turn, prevent that damage.",
			effects.Plus(effects.ManaCost("{W}"), effects.TapCost(), effects.DiscardACard()), effects.TargetCreature("target creature"),
			effects.PreventNextDamageFromChosenSource(effects.ShieldTheTarget))},
	})
}
