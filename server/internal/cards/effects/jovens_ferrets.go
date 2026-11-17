package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Joven's Ferrets — Creature — Ferret {G}, 1/1 (Homelands, #5753):
//
//	"Whenever this creature attacks, it gets +0/+2 until end of turn.
//	 At end of combat, tap all creatures that blocked this creature this
//	 turn. They don't untap during their controller's next untap step."
//
// The first ability is the eerie creatures' self pump (thisCreatureUntilEOT).
//
// The second triggers as each end of combat step begins (CR 511.2), on
// any player's turn, while the Ferrets is on the battlefield. The ruling:
// "The second ability will trigger only if Joven's Ferrets is still on the
// battlefield at end of combat." It is not a delayed trigger and has no
// "if", so it goes on the stack even when nothing blocked the Ferrets, and
// then does nothing.
//
// "Creatures that blocked this creature this turn" is read from the turn's
// EventBlock records, not from who is blocking now, so a blocker removed
// from combat still counts (CR 506.4 ends the block, not the fact that it
// blocked). A creature put onto the battlefield blocking never "blocked"
// (CR 509.4) and has no EventBlock. A blocker that has left the battlefield
// since is a new object if it returns, and so is a Ferrets that left and
// came back (CR 400.7): each is dropped from, or restarts, the record. The
// set is fixed as the ability triggers and stamped on the item
// (Params.Objects), since nothing can block the Ferrets between then and
// resolution. As it resolves, each one still on the battlefield as the
// same object and still a creature is tapped and gets ADR 0058's
// controller-keyed one-shot mark. The rulings say the mark "doesn't track
// the creatures' controllers": a creature that changes controllers keeps
// it for its new controller's next untap step, which is what the nil
// UntapSkip.Player means. It is used up by that step even on a creature
// that is untapped by then.
//
// No simplification.
func init() {
	const tapLabel = "Joven's Ferrets — tap all creatures that blocked it this turn; they don't untap during their controller's next untap step"
	tapBlockers := AtEachStep(game.StepEndCombat, tapLabel, jovensFerretsTapBlockers)
	tapBlockers.Build = func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, tapLabel)
		for _, id := range creaturesThatBlockedThisTurn(g, source.InstanceID) {
			item.Params.Objects = append(item.Params.Objects, game.ObjectRef{ID: id, Epoch: g.ObjectEpochForEffect(id)})
		}
		return item
	}
	const pumpLabel = "Joven's Ferrets — gets +0/+2 until end of turn"
	Register(Spec{
		OracleID:     "d2817b5f-f7a2-4254-a00f-e2e45d7a8cc0",
		Name:         "Joven's Ferrets",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks(pumpLabel, thisCreatureUntilEOT(pumpLabel, 0, 2)),
			tapBlockers,
		},
	})
}

// jovensFerretsTapBlockers is the end-of-combat trigger's body: tap the
// stamped blockers that are still the same creatures on the battlefield,
// and mark each for its controller's next untap step.
func jovensFerretsTapBlockers(g *game.Game, item *game.StackItem) error {
	var those []uuid.UUID
	for _, ref := range item.Params.Objects {
		if !onBattlefield(g, ref.ID) || g.ObjectEpochForEffect(ref.ID) != ref.Epoch {
			continue
		}
		if c, ok := g.LookupCardForEffect(ref.ID); ok && c.IsCreature() {
			those = append(those, ref.ID)
		}
	}
	if len(those) == 0 {
		return nil
	}
	return TapAndFreeze{Targets: those, Label: "Joven's Ferrets — doesn't untap during its controller's next untap step"}.Apply(NewContext(g, item))
}

// creaturesThatBlockedThisTurn is "creatures that blocked <attacker> this
// turn", read from the turn's events, in the order they blocked: every
// blocker of the object `attacker` is now, still on the battlefield as
// the object that blocked. A battlefield entry by `attacker` makes it a
// new object that nothing has blocked yet; any zone move by a blocker
// ends that blocker's object. Caller holds g.mu.
func creaturesThatBlockedThisTurn(g *game.Game, attacker uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, ev := range g.EventsThisTurn() {
		switch ev.Kind {
		case game.EventZoneMove:
			if ev.CardID == attacker && ev.NewZone == game.ZoneBattlefield {
				out = out[:0]
				continue
			}
			if i := slices.Index(out, ev.CardID); i >= 0 {
				out = slices.Delete(out, i, i+1)
			}
		case game.EventBlock:
			if ev.Target == attacker && ev.CardID != uuid.Nil && ev.CardID != attacker && !slices.Contains(out, ev.CardID) {
				out = append(out, ev.CardID)
			}
		}
	}
	kept := out[:0]
	for _, id := range out {
		if onBattlefield(g, id) {
			kept = append(kept, id)
		}
	}
	return kept
}
