package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phelia, Exuberant Shepherd — Legendary Creature — Dog, 2/2 for {1}{W}:
//
//	"Flash
//	 Whenever Phelia attacks, exile up to one other target nonland
//	 permanent. At the beginning of the next end step, return that
//	 card to the battlefield under its owner's control. If it entered
//	 under your control, put a +1/+1 counter on Phelia."
//
// The delayed blink is the flicker.go family's (Waterbender's
// Restoration): the exile happens as the trigger resolves, and the
// return is a CR 603.7 delayed trigger that names the card by the ID
// it had in exile — exiling does not re-mint an InstanceID, only the
// return does (CR 400.7).
//
// "IF IT ENTERED UNDER YOUR CONTROL" is the half that waited on #1327.
// The return is an entry, and an entry can stop to ask something — a
// returning Clone asks what to copy, a blinked shockland asks for its
// life — so the synchronous exile return reported "nothing entered"
// while the question was open, and a counter written on the next line
// was either never placed or placed before an entry that could still
// be cancelled. ReturnFromExile.Then runs from the landing, with the
// NEW object's ID, so the question is asked of the permanent that
// actually arrived.
//
// "Under its owner's control" makes the answer usually "it's yours if
// you own it": your own creature blinked comes back to you and grows
// Phelia, an opponent's comes home and does not. The test is on the
// arrived permanent's controller rather than on ownership, because a
// replacement could change who it enters under and the printed words
// are about where it entered.
//
// "Phelia" is the object that made the trigger: if Phelia has left the
// battlefield (or left and come back, a new object) by the end step,
// there is nothing to put the counter on, and it is skipped.
//
// "Other" is AnotherTarget: the trigger's clause is built from its
// source (TargetsFrom), so Phelia herself is never a legal target and a
// second Phelia would be. The resolution also skips her own ID, belt
// and braces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d86a59a-ba1f-45f7-b829-dca6f9f3e624",
		Name:            "Phelia, Exuberant Shepherd",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclared(ev, source)
			},
			TargetsFrom: AnotherTarget(func(other CardPredicate) *game.TargetSpec {
				return TargetPermanent("up to one other target nonland permanent", Nonland(), other).WithCount(0, 1)
			}),
			Key:    "Phelia, Exuberant Shepherd — exile up to one other nonland permanent",
			Effect: pheliaExile,
		}},
	})
}

// pheliaExile exiles the trigger's target (not Phelia itself) and, from
// the exile's continuation, schedules the return for the card that
// actually reached exile. Package-level so the stack item captures
// nothing.
func pheliaExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var ids []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil || t.ID == item.SourceCardID {
			continue
		}
		ids = append(ids, t.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		if len(exiled) == 0 {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:    game.StepEnd,
			Label: "Phelia, Exuberant Shepherd — return the exiled card",
			Cards: exiled,
			Body:  pheliaReturnBody,
		}.Apply(NewContext(g, item))
	})
}

// pheliaReturn is the delayed trigger: return the card under its
// owner's control and, once it has entered, grow Phelia if it entered
// under the trigger controller's control. A package-level func, so the
// delayed trigger shares nothing with a snapshot but this code.
func pheliaReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	you, phelia := item.Controller, item.SourceCardID
	for _, t := range item.Targets {
		if t.Kind != game.TargetCard || t.ID == uuid.Nil {
			continue
		}
		err := ReturnFromExile{
			Target: t.ID,
			Then: func(g *game.Game, entered uuid.UUID) error {
				if entered == uuid.Nil {
					return nil
				}
				c, ok := g.LookupCardForEffect(entered)
				if !ok || c.Controller != you {
					return nil
				}
				if z := g.FindCardZoneForEffect(phelia); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return g.AddCounterForEffect(phelia, "+1/+1", 1)
			},
		}.Apply(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}
