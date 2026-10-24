package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sephiroth, Fallen Hero — Legendary Creature — Human Avatar Soldier
// {3}{R}{W}, 7/5 (#3656):
//
//	"Jenova Cells — Whenever Sephiroth attacks, you may put a cell
//	 counter on target creature. Until end of turn, each modified
//	 creature you control has base power and toughness 7/5. (Equipment,
//	 Auras you control, and counters are modifications.)
//	 The Reunion — {3}, Sacrifice a modified creature: Return this card
//	 from your graveyard to the battlefield tapped."
//
// The attack trigger is mandatory and targets; the "may" is the counter
// alone: the question is asked as the trigger resolves, and the base-P/T
// sentence happens either way. The set of modified creatures is fixed as the effect applies
// (CR 611.2c), after the counter, so the creature that just got the
// counter is in it. The base-P/T set sits in layer 7b, so counters and
// Equipment still add to the 7/5. The Reunion is Reassembling Skeleton's
// return from the graveyard with a sacrifice cost, paid at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca9ab230-51e2-4ef2-a7b9-4e325d9f9657",
		Name:         "Sephiroth, Fallen Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Sephiroth, Fallen Hero — Jenova Cells: you may put a cell counter on target creature; each modified creature you control has base power and toughness 7/5 until end of turn",
					sephirothJenovaCells),
				TargetCreature("target creature")),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice a modified creature: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet | game.AnswerSurvive},
			Cost: Plus(ManaCost("{3}"), game.AbilityCost{
				SacrificeOther: sacrificeSpec("a modified creature", Creature(), YouControl(), modifiedCreature),
			}),
			Zones:  []game.ZoneKind{game.ZoneGraveyard},
			Effect: returnThisFromGraveyardTapped,
		}},
	})
}

func sephirothJenovaCells(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, hasTarget := b16FirstLegalTargetCard(ctx)
	if !hasTarget {
		return sephirothBaseStats(ctx)
	}
	// Keyed (#4298): the target is found while the item still has its clauses
	// and rides the prompt as its carried ID.
	return MayChoice{
		Question: "Sephiroth, Fallen Hero — put a cell counter on target creature?",
		Then:     sephirothCounter,
		Carry:    []uuid.UUID{target},
	}.Apply(ctx)
}

// sephirothBaseStats is the second sentence: each modified creature you
// control has base power and toughness 7/5 until end of turn, counter or not.
func sephirothBaseStats(ctx *Context) error {
	return ScopedEffectFor{
		Match: And(Creature(), YouControl(), modifiedCreature),
		Mods: []game.Mod{
			game.SetBasePowerMod(7),
			game.SetBaseToughnessMod(5),
		},
		Duration: DurationUntilEndOfTurn(ctx),
		Label:    "Sephiroth, Fallen Hero — each modified creature you control has base power and toughness 7/5 until end of turn",
	}.Apply(ctx)
}

// sephirothCounter is "you may put a cell counter on target creature": the
// counter on a yes, the base-P/T sentence either way.
var sephirothCounter = MayAnswer("confirm/sephiroth-cell-counter", func(ctx *Context, r game.ConfirmAnswer) error {
	if r.Accepted {
		if err := (AddCounter{Target: carryAt(r, 0), Kind: "cell", N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return sephirothBaseStats(ctx)
})
