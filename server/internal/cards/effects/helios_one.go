package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// HELIOS One — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: You get {E} (an energy counter).
//	 {3}, {T}, Pay X {E}, Sacrifice this land: Destroy target nonland
//	 permanent with mana value X. Activate only as a sorcery."
//
// ADR 0129 §2: "Pay X {E}" is PayXEnergy — X is announced with the
// activation (CR 107.3a), may not exceed the activator's energy
// (CR 118.3), and binds the target clause's "with mana value X", which
// the engine checks at announce (CR 601.2c via CR 602.2b) and again at
// resolution (CR 608.2b). X = 0 is a real announcement: a token or a
// land-less zero-cost permanent has mana value 0, so XMatters is not
// declared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cfb1a656-0bf1-484d-b099-33087914250b",
		Name:         "HELIOS One",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: You get {E} (an energy counter).",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Purpose: game.Purpose{Energy: 1},
				Effect:  Do(GetEnergy{N: 1}),
			},
			{
				Label: "{3}, {T}, Pay X {E}, Sacrifice this land: Destroy target nonland permanent with mana value X. Activate only as a sorcery.",
				Cost:  Plus(ManaCost("{3}"), TapCost(), PayXEnergy(), SacrificeThis()),
				Targets: TargetPermanent("target nonland permanent with mana value X",
					Nonland()).WithManaValueEqualsX(),
				SorcerySpeed: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return destroyEachLegalTarget(item, NewContext(g, item))
				},
			},
		},
	})
}
