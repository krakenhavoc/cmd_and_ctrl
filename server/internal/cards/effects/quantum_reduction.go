package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quantum Reduction — Enchantment — Aura {1}{U} (Marvel Super Heroes
// Commander, #4455):
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 You may cast this spell as though it had flash if it's cast using
//	 teamwork.
//	 Enchant creature
//	 Enchanted creature gets -5/-0 and loses all abilities."
//
// The static half is the shared attachment shape (PumpAttached and
// LoseAllAbilities, Amphibian Downpour's removal), which lasts exactly as
// long as the Aura is attached.
//
// The flash clause is CR 601.3c's "an effect allows a player to cast a
// spell as though it had flash only if an alternative or additional cost
// is paid": the player may BEGIN the cast at instant speed, and the
// teamwork they announce is the price of that. It is the optional cost's
// own flag (game.AdditionalCost.AsThoughFlash), read by the one timing
// read CastSpell, the legal-move enumerator and the view's castable_here
// share, so an ordinary cast of the Aura still waits for a sorcery's
// window while a teamwork cast does not, and a per-player restriction
// (Teferi, Time Raveler) still closes both (CR 101.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "4e18d20f-f153-4607-9201-3e22591474db",
		Name:          "Quantum Reduction",
		Completeness:  CompletenessFull,
		Targets:       EnchantCreature(),
		Purpose:       game.Purpose{Aura: game.AuraHarm},
		OptionalCosts: []game.AdditionalCost{TeamworkOpeningFlash(2)},
		Static: []game.StaticAbility{
			PumpAttached(-5, 0),
			LoseAllAbilities(),
		},
	})
}

// checkCostOpensFlash is Register's guard for AdditionalCost.AsThoughFlash
// (CR 601.3c): the clause is "if it's cast using <an optional cost>", so a
// mandatory cost carrying it would open instant speed for every cast, and
// an either/or branch has no announcement of its own to open it.
func checkCostOpensFlash(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil {
		for _, b := range append([]game.AdditionalCost{*ac}, ac.Either...) {
			if b.AsThoughFlash {
				panic(fmt.Sprintf("effects.Register: %q puts AsThoughFlash on a mandatory additional cost; it belongs on an optional cost", spec.Name))
			}
		}
	}
	for _, oc := range spec.OptionalCosts {
		if oc.AsThoughFlash && !oc.Optional {
			panic(fmt.Sprintf("effects.Register: %q puts AsThoughFlash on a cost that is not Optional", spec.Name))
		}
	}
}

// TeamworkOpeningFlash is Teamwork(n) whose announcement also lets the
// spell be cast as though it had flash: "You may cast this spell as though
// it had flash if it's cast using teamwork" (CR 601.3c, 702.194b).
func TeamworkOpeningFlash(n int) game.AdditionalCost {
	c := Teamwork(n)
	c.AsThoughFlash = true
	return c
}
