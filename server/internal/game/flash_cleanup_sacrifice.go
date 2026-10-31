package game

import "github.com/google/uuid"

// flash_cleanup_sacrifice.go — the Visions-era rider (#3156):
//
//	"You may cast this spell as though it had flash. If you cast it any
//	 time a sorcery couldn't have been cast, the controller of the
//	 permanent it becomes sacrifices it at the beginning of the next
//	 cleanup step."
//
// Necromancy prints it, and so do nine commons and uncommons from the
// same era (Lightning Reflexes, Spider Climb, Soar, Parapet, …).
//
// THREE PIECES.
//
//   - The PERMISSION is step 1b of castTimingOpenLocked: the card opens
//     the instant-speed window for itself. It is not the flash keyword
//     (CR 702.8), so nothing that asks "does it have flash" sees it, and
//     a per-player restriction (Teferi, Time Raveler) still closes it, as
//     it closes the keyword (CR 101.2).
//   - The MEMORY is StackItem.OwnFlashCast, stamped once at announce by
//     castUsesOwnFlashLocked. The 2022-12-08 ruling: the sacrifice
//     applies only when the spell was cast using its OWN permission, not
//     another one (Vedalken Orrery, Leyline of Anticipation). The engine
//     picks the other permission for the caster whenever one would have
//     opened the window, which is the choice every player makes.
//   - The SACRIFICE is a CR 603.7 delayed trigger at the beginning of the
//     next cleanup step, scheduled as the spell becomes a permanent
//     (runEntryHooksLocked), the last moment the item and the permanent
//     are both in hand. It names the permanent as that object (CR 400.7):
//     a permanent that left and came back is a new object and stays.
//     When it resolves, the permanent's controller AT THAT TIME
//     sacrifices it. A trigger in the cleanup step gives the active
//     player priority and earns another cleanup step (CR 514.3a,
//     cleanup.go), so the table may respond.
//
// A spell that never becomes a permanent (countered, or a copy that is
// not cast) schedules nothing.

// printsFlashWithCleanupSacrifice reports whether the card prints the
// rider. Read off the catalog, so it answers for a card in any zone.
func printsFlashWithCleanupSacrifice(card Card) bool {
	d := catalogDef(CatalogKey(card))
	return d != nil && d.FlashWithCleanupSacrifice
}

// castUsesOwnFlashLocked reports whether a cast of `card` that the
// timing gate has just admitted needed the card's OWN flash permission:
// it prints the rider, a sorcery couldn't be cast now, and no other
// permission (the card's type or keyword, the cast permission, the
// claimed offer, a per-player grant) would have opened the window.
//
// Caller must hold g.mu.
func (g *Game) castUsesOwnFlashLocked(playerID uuid.UUID, card Card, zone ZoneKind, perm *CastPermission, offer *AlternativeCost) bool {
	if !printsFlashWithCleanupSacrifice(card) || g.SorcerySpeedOpenLocked(playerID) {
		return false
	}
	return !g.castTimingOpenLocked(playerID, card, zone, perm, offer.opensOwnCastWindow(), false)
}

// flashCleanupSacrificeBody is the delayed trigger's body, a registered
// key (ADR 0041 phase 3). Assigned in init for the reason dashReturnBody
// is.
var flashCleanupSacrificeBody BodyRef

func init() {
	flashCleanupSacrificeBody = DelayedBody("flash/cleanup-sacrifice", flashCleanupSacrifice)
}

// scheduleOwnFlashCleanupSacrificeLocked schedules "the controller of the
// permanent it becomes sacrifices it at the beginning of the next cleanup
// step" for a spell that resolved into the permanent `entered`, when the
// spell was cast with its own flash at a time a sorcery couldn't have
// been. The delayed trigger is controlled by the spell's controller
// (CR 603.7d).
//
// Caller must hold g.mu.
func (g *Game) scheduleOwnFlashCleanupSacrificeLocked(entered uuid.UUID, item *StackItem) {
	if item == nil || !item.OwnFlashCast || item.IsCopy {
		return
	}
	live, ok := g.battlefieldCardLocked(entered)
	if !ok {
		return
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: live.InstanceID,
		Label:        live.Name + " — cast as though it had flash, its controller sacrifices it",
		At:           StepCleanup,
		Cards:        []uuid.UUID{live.InstanceID},
		Body:         flashCleanupSacrificeBody,
		// The object the spell became (CR 400.7), recorded now.
		Params: EffectParams{Object: ObjectRef{ID: live.InstanceID, Epoch: live.ObjectEpoch}},
	})
}

// flashCleanupSacrifice is the delayed trigger's body: the permanent's
// current controller sacrifices it, if it is still the object the spell
// became. Nothing to sacrifice is not an error (CR 608.2c).
func flashCleanupSacrifice(g *Game, _ *StackItem, p EffectParams) error {
	id := p.Object.ID
	if id == uuid.Nil {
		return nil
	}
	c, ok := g.battlefieldCardLocked(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	return g.SacrificePermanentForEffect(id)
}
