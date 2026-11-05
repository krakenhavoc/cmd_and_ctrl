package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hidden Hideout — Land (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"This land enters tapped.
//	 {T}: Add one mana of any color in your commander's color identity.
//	 {2}, {T}: Target creature you control with a counter on it gains
//	 lifelink until end of turn."
//
// The mana ability is Command Tower's (NarrowToCommanderIdentity): two
// commanders pool their identities, and a colourless commander or none at
// all makes it produce nothing (the rulings). The lifelink target must be
// a creature you control with a counter of any kind on it, both as it is
// chosen and as the ability resolves (CR 608.2b); a creature whose last
// counter is gone by then is an illegal target and gains nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b639e1fe-d099-4cab-a0d0-a1b33c7f31dd",
		Name:         "Hidden Hideout",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:                      ManaAbilityCost{Tap: true},
			Produced:                  "{W|U|B|R|G}",
			Label:                     "Add one mana of any color in your commander's color identity",
			NarrowToCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Target creature you control with a counter on it gains lifelink until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink")),
			Cost:    Plus(ManaCost("{2}"), TapCost()),
			Targets: TargetCreature("target creature you control with a counter on it", YouControl(), b41HasACounter()),
			Effect:  ebTargetCreatureGainsUntilEOT("lifelink", "Hidden Hideout — lifelink until end of turn"),
		}},
	})
}
