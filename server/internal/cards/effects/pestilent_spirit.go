package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pestilent Spirit — Creature — Spirit {2}{B}, 3/2:
//
//	"Menace, deathtouch
//	 Instant and sorcery spells you control have deathtouch. (Any
//	 amount of damage they deal to a creature is enough to destroy
//	 it.)"
//
// The static is a layer-6 grant over SPELLS (CR 613.1f, ADR 0107 §3),
// the same SpellsYouControlHave shape Cast Through Time uses for
// rebound: the stack step of the layer pass gives each of its
// controller's instants and sorceries deathtouch while they are on the
// stack. A spell deals its damage from the stack (CR 702.2d), and the
// damage tail reads the spell's keywords there (effectDamageTailLocked
// → stackSpellLocked), so a Lightning Bolt's 3 to a 5/5 marks it for
// CR 704.5h and the excess-damage readers count one point as lethal.
// A static is never locked in (CR 611.3a): a spell still on the stack
// when the Spirit leaves loses deathtouch.
//
// Damage a spell makes a creature deal (a fight, "target creature you
// control deals damage equal to its power") is the creature's, not the
// spell's, so it gets nothing from this, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83930f98-5596-4713-998c-732d5ecedf72",
		Name:            "Pestilent Spirit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "deathtouch"},
		Static: []game.StaticAbility{
			Labelled(SpellsYouControlHave(Or(Instant(), Sorcery()), "deathtouch"),
				"Instant and sorcery spells you control have deathtouch."),
		},
	})
}
