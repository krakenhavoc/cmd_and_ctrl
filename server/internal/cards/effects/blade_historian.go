package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade Historian — Creature — Human Cleric {R/W}{R/W}{R/W}{R/W}, 2/3
// (EDHREC rank 2711):
//
//	"Attacking creatures you control have double strike."
//
// Berserkers' Onslaught on a body, and the same body of code: the
// printed static is written as an attack trigger because the layer
// engine does not recompute on an attack declaration (no tap event,
// no zone move, no counter), so a layer 6 static gated on "is
// attacking" stays cached from before combat. The trigger fires once
// per attacking creature the controller controls — the Historian
// itself included when it attacks — and grants double strike until
// end of turn through the turn-scoped static registry, which does
// invalidate the cache (b25AttackersHaveDoubleStrike).
//
// Declared simplification (Berserkers' Onslaught's): the grant
// lasts the turn rather than the attack, and lands only on creatures
// DECLARED as attackers. A creature put onto the battlefield
// attacking is not covered (weaker than printed); a creature removed
// from combat keeps double strike until end of turn, which changes
// nothing outside combat. Never stronger. The seam is a layer-cache
// bump on attack declaration, after which this is a one-line static.
func init() {
	Register(Spec{
		OracleID:     "314f0a96-5e91-42a3-9d75-3f641f22a9ee",
		Name:         "Blade Historian",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Double strike is granted as each of your creatures is declared as an attacker and lasts until end of turn, so a creature that enters the battlefield already attacking doesn't get it."},
		Triggered: []game.TriggeredAbility{
			b25AttackersHaveDoubleStrike("Blade Historian"),
		},
	})
}
