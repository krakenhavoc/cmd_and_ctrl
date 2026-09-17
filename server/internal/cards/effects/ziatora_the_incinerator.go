package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ziatora, the Incinerator — Legendary Creature — Demon Dragon
// {3}{B}{R}{G}, 6/6 (EDHREC rank 3098):
//
//	"Flying
//	 At the beginning of your end step, you may sacrifice another
//	 creature. When you do, Ziatora deals damage equal to that
//	 creature's power to any target and you create three Treasure
//	 tokens."
//
// The Riveteers' Fling engine. Flying rides PrintedKeywords. The
// end-step trigger is Ruthless Technomancer's shape: the "you may"
// is the trigger's optional prompt, the creature is chosen as the
// trigger's target (any other creature you control — the picker
// the engine has for a choice among your own permanents), and on
// resolution it is sacrificed (b29SacrificeChosenCreature).
//
// The printed "when you do" is a CR 603.12 REFLEXIVE trigger, and
// since #636 it is exactly that: the end-step ability creates it as
// it resolves, having actually sacrificed something. It goes on the
// stack above its parent with its own "any target" clause — chosen
// after the sacrifice, as printed — carrying the sacrificed
// creature's instance ID as its payload, and on resolution deals
// damage equal to that creature's power as it last stood
// (b29PowerAsItLastStood) and makes the three Treasures whether or
// not the damage lands.
//
// Two sandbox simplifications, declared, both weaker than printed:
//
//   - The creature to sacrifice is chosen when the end-step trigger
//     goes on the stack, not on resolution (the Technomancer's
//     caveat): an opponent who removes it in response fizzles the
//     trigger, where printed you would pick another. A second
//     Ziatora cannot be chosen — "another" is by name.
//   - A static power bonus from another permanent (an anthem, an
//     Equipment) is not in the sacrificed creature's last-known
//     power; counters are.
//
// What is no longer a simplification: the reflexive half used to be
// a second ability harvested off Ziatora on the battlefield, so a
// Ziatora removed in response to her own end-step trigger took the
// damage and the Treasures with her. The trigger now belongs to the
// resolving ability, not to the permanent, so it happens either way
// — which is what the card says.
func init() {
	Register(Spec{
		OracleID:     "d46c3fb6-f1c7-4a96-ae42-5bce17fc7c1d",
		Name:         "Ziatora, the Incinerator",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature to sacrifice when the end-step trigger goes on the stack rather than on resolution, so opponents can respond to the choice.",
			"The sacrificed creature's power is read from its printed value and its counters only — a bonus from another permanent isn't counted in the damage.",
		},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(
				Targeting(
					AtYourEndStep(b29ZiatoraSacrificeLabel, b29SacrificeChosenCreature),
					TargetCreature("another creature you control", YouControl(), b03NotNamed("Ziatora, the Incinerator"))),
				"Ziatora, the Incinerator — sacrifice another creature to deal damage equal to its power and make three Treasures?"),
		},
	})
}
