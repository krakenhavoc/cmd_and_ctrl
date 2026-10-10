package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Allosaurus Shepherd — Creature — Elf Shaman {G}, 1/1:
//
//	"This spell can't be countered.
//	 Green spells you control can't be countered.
//	 {4}{G}{G}: Until end of turn, each Elf creature you control has base
//	 power and toughness 5/5 and becomes a Dinosaur in addition to its
//	 other creature types."
//
// The rider is Spec.CantBeCountered; the second line is ADR 0106 §4's
// battlefield static (#1806), and "green" is the spell's colour as it
// stands on the stack.
//
// The activated ability is one ScopedEffectFor record (ADR 0041 phase
// 3): base P/T at layer 7b and the added subtype at layer 4, all data.
// It changes characteristics, so the Elves it affects are fixed as it
// resolves (CR 611.2c): an Elf that enters later in the turn stays as
// it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "98295bdc-db58-4e1f-ad4c-63092aea207e",
		Name:            "Allosaurus Shepherd",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouControlCantBeCountered("Green spells you control can't be countered.", OfColor("G")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{G}{G}: Until end of turn, each Elf creature you control has base power and toughness 5/5 and becomes a Dinosaur in addition to its other creature types.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Cost:    ManaCost("{4}{G}{G}"),
			Effect:  allosaurusShepherdElvesBecomeDinosaurs,
		}},
	})
}

// allosaurusShepherdElvesBecomeDinosaurs makes each Elf creature the
// activator controls a 5/5 Dinosaur until end of turn.
func allosaurusShepherdElvesBecomeDinosaurs(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return ScopedEffectFor{
		Match: And(OfCreatureType("Elf"), YouControl()),
		Mods: []game.Mod{
			game.SetBasePowerMod(5),
			game.SetBaseToughnessMod(5),
			game.AddSubtypesMod("Dinosaur"),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Allosaurus Shepherd — each Elf creature you control is a 5/5 Dinosaur until end of turn",
	}.Apply(ctx)
}
