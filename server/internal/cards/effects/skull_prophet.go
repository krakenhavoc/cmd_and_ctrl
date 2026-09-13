package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skull Prophet — Creature — Human Druid {B}{G}, 3/1 (EDHREC rank
// 1389):
//
//	"{T}: Add {B} or {G}.
//	 {T}: Mill two cards. (Put the top two cards of your library into
//	 your graveyard.)"
//
// A two-colour mana dork that is also a self-mill engine. The mana
// is a pipe ability over the two printed colours with the
// commander-identity narrowing off (the text names the colours, not
// the command zone); the mill is a CR 602 activated ability with a
// tap cost, so it uses the stack and can be responded to. Both taps
// are on a creature, so summoning sickness applies to each (CR
// 302.1) — the engine enforces that for mana abilities and activated
// abilities alike.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "39b51f62-8421-42d1-86d3-74fd6e6f31a2",
		Name:         "Skull Prophet",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:                    ManaAbilityCost{Tap: true},
			Produced:                "{B|G}",
			Label:                   "Add {B} or {G}",
			IgnoreCommanderIdentity: true,
		}},
		Activated: []ActivatedAbility{{
			Label: "{T}: Mill two cards.",
			Cost:  TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MillCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
