package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b8_helpers.go — shared pieces for Teenage Mutant Ninja Turtles,
// batch 8 (#4457).

// tmtB8RobotCountersOnFirstTarget is "put three +1/+1 counters on target
// artifact you control. If it isn't a creature, it becomes a 0/0 Robot
// creature in addition to its other types" (Donatello, Mutant Mechanic;
// Does Machines' level 3), on the first still-legal battlefield target.
//
// Tezzeret, Cruel Captain's emblem is the precedent. No duration is
// stated, so the type change is CR 611.2a's indefinite one pinned to that
// object. The creature test is made on resolution and before the
// counters land, which gives the printed order's answer: counters never
// make a permanent a creature, and putting them first would run the type
// change while a CR 616 ordering prompt could still be open.
func tmtB8RobotCountersOnFirstTarget(label string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id := FirstLegalBattlefieldTarget(ctx)
		if id == uuid.Nil {
			return nil
		}
		if c, ok := g.LookupCardForEffect(id); ok && !c.IsCreature() {
			robot := []game.Mod{
				game.AddTypesMod("Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Robot")...),
				game.SetBasePowerMod(0),
				game.SetBaseToughnessMod(0),
			}
			duration := g.PinnedTo(game.IndefiniteDuration(), id)
			if err := (ScopedEffectFor{Target: id, Mods: robot, Duration: duration, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 3}.Apply(ctx)
	}
}

// tmtB8SpellTargetsArtifactOrCreatureYouControl is Fugitive Droid's
// "target spell that targets an artifact or creature you control": a
// spell at least one of whose targets is, on the battlefield, an artifact
// or a creature the counterer controls. Rebuff the Wicked's predicate
// (b14SpellTargetsAPermanentYouControl) narrowed to those two types,
// read after the layers.
func tmtB8SpellTargetsArtifactOrCreatureYouControl() CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		item := g.StackItemForEffect(c.InstanceID)
		if item == nil {
			return false
		}
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard || !onBattlefield(g, t.ID) {
				continue
			}
			tc, ok := g.LookupCardForEffect(t.ID)
			if ok && tc.Controller == caster && (tc.IsArtifact() || tc.IsCreature()) {
				return true
			}
		}
		return false
	}
}
