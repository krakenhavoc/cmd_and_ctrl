package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pelargir Survivor — Creature — Human Peasant {1}{U}, 1/3 (The Hobbit):
//
//	"{T}: Add one mana of any color. Spend this mana only to cast an
//	 instant or sorcery spell.
//	 {5}{U}, {T}: Target player mills three cards."
//
// The mana ability is the any-colour tap of Beastcaller Savant with the
// Ore-Rich Stalactite restriction: it may be spent only to cast
// (ManaRestrictCast) a spell of one of two types (ManaRestrictAnyType),
// so it cannot pay an activation cost or a creature spell. The mill is
// an ordinary targeted ability whose body is TargetPlayerMills.
// Summoning sickness gates both taps (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce0b3c22-3f77-4a6c-a126-9ed3c06b4482",
		Name:         "Pelargir Survivor",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			Label:        "Add one mana of any color. Spend this mana only to cast an instant or sorcery spell",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}{U}, {T}: Target player mills three cards.",
			Cost:    Plus(ManaCost("{5}{U}"), TapCost()),
			Targets: TargetPlayer("target player"),
			Effect:  TargetPlayerMills(3),
		}},
	})
}
