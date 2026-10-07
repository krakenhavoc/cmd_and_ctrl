package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Targets described relative to their source (#2146): "target creature
// with lesser power" (Rangers of Ithilien, mentor), "with power less
// than or equal to this creature's power" (Earthshaker Khenra), "with
// greater toughness".
//
// A CardPredicate is handed the caster and the candidate, never the
// object whose ability it is, so these are data on the clause instead:
// game.TargetSpec.RelativeToSource, which the engine evaluates at announce (the
// legal set and the CR 601.2c check) and again at resolution
// (CR 608.2b). The source is read through TargetSource.PowerToughness,
// so it is the permanent's EFFECTIVE power and toughness — layers and
// counters, unclamped — while it stands and its last-known information
// once it has left (CR 608.2h). A TargetsFrom that reads
// source.CurrentPower() instead freezes the bound when the trigger is
// built, which is wrong the moment the source is pumped or shrunk in
// response.

// RelativeToSource narrows a clause with source-relative comparisons;
// every one must hold. Wrap the whole clause, like Another:
//
//	Targets: RelativeToSource(
//	    TargetCreature("up to one target creature with lesser power").WithCount(0, 1),
//	    LesserPower())
func RelativeToSource(spec *game.TargetSpec, comparisons ...game.SourceComparison) *game.TargetSpec {
	spec.RelativeToSource = append(spec.RelativeToSource, comparisons...)
	return spec
}

// LesserPower is "with lesser power" / "with power less than this
// creature's power".
func LesserPower() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourcePower, Cmp: game.CmpLess}
}

// PowerNoGreater is "with power less than or equal to this creature's
// power". The source itself qualifies.
func PowerNoGreater() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourcePower, Cmp: game.CmpLessEq}
}

// GreaterPower is "with greater power".
func GreaterPower() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourcePower, Cmp: game.CmpGreater}
}

// LesserToughness is "with lesser toughness".
func LesserToughness() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourceToughness, Cmp: game.CmpLess}
}

// GreaterToughness is "with greater toughness".
func GreaterToughness() game.SourceComparison {
	return game.SourceComparison{Stat: game.SourceToughness, Cmp: game.CmpGreater}
}

// Mentor is CR 702.136's keyword ability: "Whenever this creature
// attacks, put a +1/+1 counter on target attacking creature with
// lesser power." The clause is relative to the mentor, so it is
// re-judged at resolution against the mentor's power then (or its
// last-known power if it has left), and CR 603.3d drops the trigger
// when no attacker qualifies.
func Mentor(key string) game.TriggeredAbility {
	return game.TriggeredAbility{
		Watches:   []game.EventKind{game.EventAttack},
		AppliesTo: ThisAttacked,
		Key:       key,
		Targets: RelativeToSource(
			TargetCreature("target attacking creature with lesser power", attackingPred()),
			LesserPower()),
		Effect: b36CounterOnChosenAnimal,
	}
}

func attackingPred() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.AttackingTarget != uuid.Nil }
}
