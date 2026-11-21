package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Radiant Kavu — Creature — Kavu {R}{G}{W}, 3/3:
//
//	"{R}{G}{W}: Prevent all combat damage blue creatures and black creatures would deal this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): a preventFromSource record with no
// source and two properties (a blue creature, a black creature), checked
// as the damage would be dealt (CR 615.9), so a creature that turns blue
// after the ability resolves is caught and one that stops being blue is
// not. Combat damage only, to anything, for the rest of the turn.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab46b957-5206-4390-a2f6-aba2e8d1debe",
		Name:         "Radiant Kavu",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.NextDamageShieldRow(
			"{R}{G}{W}: Prevent all combat damage blue creatures and black creatures would deal this turn.",
			effects.ManaCost("{R}{G}{W}"), nil,
			effects.PreventDamageFromSource{CombatOnly: true, Protect: effects.ShieldAnything, Queries: []game.PermanentQuery{
				{Types: []string{"creature"}, Colors: []string{"U"}},
				{Types: []string{"creature"}, Colors: []string{"B"}},
			}})},
	})
}
