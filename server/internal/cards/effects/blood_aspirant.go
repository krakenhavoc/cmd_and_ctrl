package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Aspirant — Creature — Satyr Berserker {1}{R}, 1/1 (Commander
// Masters):
//
//	"Whenever you sacrifice a permanent, put a +1/+1 counter on this
//	 creature.
//	 {1}{R}, {T}, Sacrifice a creature or enchantment: This creature deals 1
//	 damage to target creature. That creature can't block this turn."
//
// The counter is Bloodbriar's trigger without "another" (the Aspirant can be
// the permanent sacrificed, and then has nothing to put the counter on).
// The activated ability is Ertai the Corrupted's sacrifice cost with a
// target chosen before the cost is paid (CR 602.2b), so the creature
// sacrificed may be the target and the Aspirant itself may be sacrificed; the
// damage is then dealt by the Aspirant as it last existed (CR 608.2h) and the
// can't-block rider still applies to a target that survives it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6db51087-6995-4ed4-978a-3d17d60019af",
		Name:         "Blood Aspirant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, b23YouSacrificedAPermanent,
				"Blood Aspirant — a +1/+1 counter on it", Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, {T}, Sacrifice a creature or enchantment: This creature deals 1 damage to target creature. That creature can't block this turn.",
			Purpose: game.Purpose{Answers: game.AnswerRemove | game.AnswerSacOutlet},
			Cost: Plus(ManaCost("{1}{R}"), TapCost(),
				SacrificeN(1, "a creature or enchantment", Or(Creature(), Enchantment()))),
			Targets: TargetCreature("target creature"),
			Effect:  gapB104AspirantPing,
		}},
	})
}

func gapB104AspirantPing(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	victim := uuid.Nil
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			victim = t.ID
			break
		}
	}
	if victim == uuid.Nil {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: victim, Amount: 1}).Apply(ctx); err != nil {
		return err
	}
	return targetCreatureCantBlockThisTurn("Blood Aspirant — that creature can't block this turn")(g, item)
}
