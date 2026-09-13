package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aetheric Amplifier — Artifact {3} (EDHREC rank 2428):
//
//	"{T}: Add one mana of any color.
//	 {4}, {T}: Choose one. Activate only as a sorcery.
//	 • Double the number of each kind of counter on target permanent.
//	 • Double the number of each kind of counter you have."
//
// A Manalith that grows into a Deepglow Skate. The mana ability is
// the Birds shape — any colour, the printed width, so the pipe is not
// narrowed to the commander's identity. The activation is a CR 602
// ability at sorcery speed sharing the tap; its body is Deepglow
// Skate's per-target doubler (each kind snapshotted first, so a
// Doubling Season firing on one kind cannot change what the next
// receives), on one target permanent — anyone's, as printed.
//
// Declared simplification (weaker than printed): the activation has
// no mode picker — Spec.Activated carries a target clause but no
// Modes (a modal activated ability is the Rankle / Aether Channeler
// seam) — so the ability is always its first bullet. The second, the
// counters the controller HAS, is not offered; the engine's player
// counters (poison, energy, experience, rad) are untouched.
func init() {
	Register(Spec{
		OracleID:     "295cd8e1-0830-46c1-9957-556afd4bcee6",
		Name:         "Aetheric Amplifier",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The activated ability always doubles the counters on a target permanent — the option to double the counters you have isn't offered."},
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{W|U|B|R|G}",
			Label:                   "Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label:        "{4}, {T}: Double the number of each kind of counter on target permanent.",
			Cost:         Plus(ManaCost("{4}"), TapCost()),
			Targets:      TargetPermanent("target permanent"),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return b17DoubleCountersOn(ctx, t.ID)
					}
				}
				return nil
			},
		}},
	})
}
