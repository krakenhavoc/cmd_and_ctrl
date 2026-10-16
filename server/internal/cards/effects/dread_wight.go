package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dread Wight — Creature — Zombie (3/4) for {3}{B}{B}:
//
//	"At end of combat, put a paralyzation counter on each creature
//	 blocking or blocked by this creature and tap those creatures. Each
//	 of those creatures doesn't untap during its controller's untap step
//	 for as long as it has a paralyzation counter on it. Each of those
//	 creatures gains "{4}: Remove a paralyzation counter from this
//	 creature.""
//
// #3923, ADR 0109 §2: the untap hold (CR 502.3, #1313) is timed by
// game.WhilePinnedHasCounter{paralyzation}, pinned to each creature the
// counter went on. It ends the moment the last paralyzation counter
// leaves and for good (CR 611.2b): a later paralyzation counter from
// another end of combat starts a hold of its own. A creature the counter
// could not be put on is still tapped, and its hold never starts.
//
// The granted ability prints no duration, so it lasts for as long as the
// creature stays on the battlefield as the same object (CR 611.2a), even
// after its last counter is gone. Whoever controls the creature may
// activate it (CR 602.2).
//
// "At end of combat" is every combat, not only the controller's: it
// triggers as each end of combat step begins (CR 511.2), when the
// creatures are still in combat, and finds its creatures as it
// resolves. A Wight that has left the battlefield in response is read
// from its last-known blocks (CR 608.2h). A creature removed from combat
// is neither blocking nor blocked (CR 506.4).
//
// No simplification.
const (
	dreadWightCounter = "paralyzation"
	dreadWightGrant   = "dread-wight/remove-paralyzation"
)

func init() {
	Register(Spec{
		OracleID:     "f9078d79-3250-44e8-aba1-ab52e0ba4425",
		Name:         "Dread Wight",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  dreadWightGrant,
			Text: "{4}: Remove a paralyzation counter from this creature.",
			Activated: []ActivatedAbility{{
				Label: "{4}: Remove a paralyzation counter from this creature.",
				Cost:  ManaCost("{4}"),
				// It answers nothing an opponent does: the counter only
				// holds the creature through its untap steps.
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return AddCounter{Target: ctx.Source(), Kind: dreadWightCounter, N: -1}.Apply(ctx)
				},
			}},
		}},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEndCombat,
				"Dread Wight — paralyze each creature blocking or blocked by it",
				dreadWightParalyze),
		},
	})
}

// dreadWightParalyze is the end-of-combat trigger's body, in printed
// order: the counters, the taps, then each creature's hold and grant.
func dreadWightParalyze(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	those := dreadWightCombatPartners(ctx)
	member := ctx.asGroupMember()
	for _, id := range those {
		if err := (AddCounter{Target: id, Kind: dreadWightCounter, N: 1}).Apply(member); err != nil {
			return err
		}
	}
	if err := tapEach(ctx, those); err != nil {
		return err
	}
	for _, id := range those {
		if d, ok := DurationWhileItHasCounter(ctx, id, dreadWightCounter); ok {
			if err := g.HoldUntappedForEffect(id, d); err != nil {
				return err
			}
		}
		if !onBattlefield(g, id) {
			continue
		}
		if err := (GrantAbilitiesFor{
			Target:   id,
			Keys:     []string{dreadWightGrant},
			Duration: g.PinnedTo(game.IndefiniteDuration(), id),
			Label:    "Dread Wight — {4}: remove a paralyzation counter",
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// dreadWightCombatPartners is "each creature blocking or blocked by this
// creature", read as the trigger resolves: the Wight as it is now, or as
// it last existed on the battlefield (CR 608.2h).
func dreadWightCombatPartners(ctx *Context) []uuid.UUID {
	self, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	wight := ctx.Source()
	var out []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.InstanceID == wight || !c.IsCreature() {
			continue
		}
		blocking := self.Attacking && c.IsBlockingAttacker(wight)
		blocked := c.AttackingTarget != uuid.Nil && slices.Contains(self.Blocking, c.InstanceID)
		if blocking || blocked {
			out = append(out, c.InstanceID)
		}
	}
	return out
}
