package game

import "github.com/google/uuid"

// blitz.go — blitz (CR 702.152, #3772): "If you cast this spell for its
// blitz cost, it gains haste and 'When this creature dies, draw a card.'
// Sacrifice it at the beginning of the next end step."
//
// The blitz cost is an alternative cost (AlternativeCost.Blitz), dash's
// sibling, and it attaches its clauses to the permanent's ENTRY
// (queueAltCostEntryTriggerLocked) the same way:
//
//   - HASTE and the DRAW are one layer-6 scoped effect for as long as
//     the creature remains on the battlefield (CR 702.152a: "as long as
//     this permanent's blitz cost was paid, it has haste and …"), pinned
//     to the object, so a creature that left and came back has neither.
//     The draw is a granted triggered ability (ADR 0093), the catalog
//     bundle BlitzDrawGrant, so it is the CREATURE's ability: its
//     controller draws, a later "loses all abilities" takes it, and it
//     fires from last-known information as the creature dies.
//   - The SACRIFICE is a CR 603.7 delayed trigger at the next end step.
//     It sacrifices the creature only if it is still the object that was
//     blitzed (CR 400.7) and the trigger's controller still controls it
//     (CR 701.17a: a player can sacrifice only a permanent they control).
//
// The same pair is printed outside the keyword — Jaxis, the
// Troublemaker's token "gains haste and 'When this token dies, draw a
// card.' Sacrifice it at the beginning of the next end step." —
// so it is exported as ApplyBlitzRidersForEffect.

// BlitzDrawGrant is the bundle key of the granted "When this creature
// dies, draw a card." The effects package registers the bundle; the
// engine only names it.
const BlitzDrawGrant = "blitz/draw"

// blitzRidersLabel names the scoped effect on the wire.
const blitzRidersLabel = "Blitz — haste, and draws a card when it dies"

// blitzSacrificeBody is the delayed trigger's body, a registered key
// (ADR 0041 phase 3). Assigned in init, for the reason dashReturnBody's
// comment gives.
var blitzSacrificeBody BodyRef

func init() { blitzSacrificeBody = DelayedBody("blitz/sacrifice", blitzSacrifice) }

// applyBlitzLocked is the alternative cost's entry clause. Caller must
// hold g.mu.
func (g *Game) applyBlitzLocked(card Card, item *StackItem, alt *AlternativeCost) {
	g.ApplyBlitzRidersForEffect(card.InstanceID, item.Controller, card.Name+" — "+alt.Label+", sacrifice it")
}

// ApplyBlitzRidersForEffect gives the permanent `cardID` haste and "When
// this creature dies, draw a card." for as long as it remains on the
// battlefield, and schedules a delayed trigger, controlled by
// `controller`, that sacrifices it at the beginning of the next end
// step. `label` names the delayed trigger. A permanent that is not on
// the battlefield gets nothing (CR 611.2b). Caller must hold g.mu.
func (g *Game) ApplyBlitzRidersForEffect(cardID, controller uuid.UUID, label string) {
	live, ok := g.battlefieldCardLocked(cardID)
	if !ok {
		return
	}
	if d, ok := g.ForAsLongAsOnBattlefieldDuration(live.InstanceID); ok {
		g.RegisterScopedEffectForEffect(live.InstanceID,
			[]AffectedObject{PinObject(live.InstanceID, live.EnteredBattlefieldAt)},
			[]Mod{AddKeywordsMod("haste"), GrantAbilitiesMod(BlitzDrawGrant)}, d, blitzRidersLabel)
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   controller,
		SourceCardID: live.InstanceID,
		Label:        label,
		At:           StepEnd,
		Cards:        []uuid.UUID{live.InstanceID},
		Body:         blitzSacrificeBody,
		// The object that was blitzed (CR 400.7), recorded now.
		Params: EffectParams{Amount: live.ObjectEpoch},
	})
}

// blitzSacrifice is the delayed trigger's body: sacrifice the blitzed
// creature if it is still that object and its controller is the
// trigger's. A creature that already left, or that changed hands, stays
// where it is; nothing to sacrifice is not an error (CR 608.2c).
func blitzSacrifice(g *Game, it *StackItem, p EffectParams) error {
	for _, t := range it.Targets {
		if t.Kind != TargetCard {
			continue
		}
		c, ok := g.battlefieldCardLocked(t.ID)
		if !ok || c.ObjectEpoch != p.Amount || c.Controller != it.Controller {
			continue
		}
		if g.cantBeSacrificedLocked(t.ID) {
			continue
		}
		if err := g.sacrificePermanentLocked(t.ID); err != nil && err != ErrCardNotFound && err != ErrCantBeSacrificed {
			return err
		}
	}
	return nil
}
