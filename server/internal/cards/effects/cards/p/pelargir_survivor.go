package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ce0b3c22-3f77-4a6c-a126-9ed3c06b4482",
		Name:         "Pelargir Survivor",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{effects.ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			Label:        "Add one mana of any color. Spend this mana only to cast an instant or sorcery spell",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{U}, {T}: Target player mills three cards.",
			Cost:    effects.Plus(effects.ManaCost("{5}{U}"), effects.TapCost()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetPlayerMills(3),
		}},
	})
}
