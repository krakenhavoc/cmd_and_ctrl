package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Akroan War — Enchantment — Saga for {3}{R}:
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — Gain control of target creature for as long as this Saga
//	     remains on the battlefield.
//	 II — Until your next turn, creatures your opponents control attack
//	      each combat if able.
//	 III — Each tapped creature deals damage to itself equal to its
//	       power."
//
// Chapter I is Sower of Temptation's shape: GainControl for
// DurationWhileSourceRemains, keyed to the Saga as the object it is now
// (CR 400.7), so the creature goes home when the Saga is sacrificed
// after chapter III — or earlier, if it is destroyed.
//
// Chapter II is the reason the card waited (#1571): a CR 508.1d attack
// requirement on "creatures your opponents control", until your next
// turn. A requirement is not a characteristic change, so CR 611.2c does
// not lock the set — a creature an opponent casts during that turn
// cycle has to attack too — and the engine judges it with every other
// requirement: a tapped or summoning-sick creature, or one that could
// attack only by paying a tax, is not made to (CR 508.1d).
//
// Chapter III reads each creature's power as the chapter resolves and
// has each tapped creature deal that much damage to itself — the
// creature is the source, so its deathtouch or lifelink applies. The
// set is collected before any damage is dealt, and no state-based
// action runs until the chapter has finished resolving, so the damage
// is simultaneous as printed.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8c0b9596-c44f-44fb-89c9-056da0db33d3",
		Name:         "The Akroan War",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTriggerTargeting(1, "The Akroan War — I: gain control of target creature",
				TargetCreature("target creature"), akroanWarSteal),
			ChapterTrigger(2, "The Akroan War — II: creatures your opponents control attack each combat if able",
				akroanWarForceAttacks),
			ChapterTrigger(3, "The Akroan War — III: each tapped creature deals damage to itself",
				akroanWarTappedCreaturesHitThemselves),
		},
	})
}

func akroanWarSteal(g *game.Game, item *game.StackItem) error {
	if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
		return nil
	}
	ctx := NewContext(g, item)
	d, ok := DurationWhileSourceRemains(ctx, item.SourceCardID)
	if !ok {
		// CR 611.2b: the Saga is already gone, so the effect never
		// begins.
		return nil
	}
	return GainControl{
		Target:     item.Targets[0].ID,
		Controller: item.Controller,
		Duration:   d,
		Label:      "The Akroan War — control for as long as the Saga remains",
	}.Apply(ctx)
}

func akroanWarForceAttacks(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return OpponentsCreaturesAttackIfAble{
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "The Akroan War — creatures your opponents control attack each combat if able",
	}.Apply(ctx)
}

func akroanWarTappedCreaturesHitThemselves(g *game.Game, item *game.StackItem) error {
	type hit struct {
		id    uuid.UUID
		power int
	}
	var hits []hit
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() || !c.Tapped {
			continue
		}
		if p := c.CurrentPower(); p > 0 {
			hits = append(hits, hit{id: c.InstanceID, power: p})
		}
	}
	ctx := NewContext(g, item)
	for _, h := range hits {
		if err := (DealDamage{Source: h.id, Target: h.id, Amount: h.power}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}
