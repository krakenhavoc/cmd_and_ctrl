package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Queen Mother Ramonda (#4659, Marvel Super Heroes Commander) — Legendary
// Creature — Human Noble {3}{W}{W}, 3/5:
//
//	"When Queen Mother Ramonda enters, you become the monarch.
//	 As long as you're the monarch, creatures with power 2 or less can't
//	 attack you."
//
// The attack line is CreaturesMatchingCantAttackYouWhile: the player alone,
// not their planeswalkers (the text says "you"), gated on the crown. The
// gate is a layer input, read when the layer pass runs, and the crown
// moving invalidates the layers (EventMonarchChanged), so the restriction
// goes to whoever takes the crown with no creature moving. The power is
// judged apart, at declaration (CR 508.1c), on the attacker's power as the
// layers and counters leave it then: a creature pumped to 3 before
// attackers are declared may attack, and one shrunk to 2 may not. Any
// controller's creatures are bound, Ramonda's own included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5619a13-b229-4202-a217-bf6509706e77",
		Name:         "Queen Mother Ramonda",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouBecomeTheMonarch("Queen Mother Ramonda"),
		},
		Static: []game.StaticAbility{CreaturesMatchingCantAttackYouWhile(
			"As long as you're the monarch, creatures with power 2 or less can't attack you.",
			YoureTheMonarch,
			game.PermanentQuery{Types: []string{"creature"}}.WithPowerAtMost(2))},
	})
}
