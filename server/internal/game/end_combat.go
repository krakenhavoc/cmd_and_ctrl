package game

import "github.com/google/uuid"

// end_combat.go — CR 724.2, "end the combat phase" (#1317, ADR 0045
// amendment 2026-09-23). One card does this: Mandate of Peace.
//
// The rules, verbatim (August 7, 2026):
//
//	724.2   … When an effect ends the combat phase, follow these steps
//	        in order, as they differ from the normal process for
//	        resolving spells and abilities.
//	724.2a  If there are any triggered abilities that triggered before
//	        this process began but haven't been put onto the stack yet,
//	        those abilities cease to exist. …
//	724.2b  Exile every object on the stack, including the object
//	        that's resolving. All objects not on the battlefield or in
//	        the command zone that aren't represented by cards will cease
//	        to exist the next time state-based actions are checked.
//	724.2c  Check state-based actions. No player gets priority, and no
//	        triggered abilities are put onto the stack.
//	724.2d  The current combat phase ends. Remove all creatures and
//	        planeswalkers from combat. Effects that last "until end of
//	        combat" expire. The game skips straight to the next phase,
//	        usually the postcombat main phase; skip any steps between
//	        this step and that phase.
//	724.2e  Even though the combat phase ends, "at end of combat"
//	        triggered abilities don't trigger because the end of combat
//	        step is skipped.
//	724.2f  No player gets priority during this process, so triggered
//	        abilities are not put onto the stack. If any triggered
//	        abilities have triggered since this process began, those
//	        abilities are put onto the stack during the following phase…
//	724.2g  If an effect attempts to end the combat phase at any time
//	        that's not a combat phase, nothing happens.
//
// # How each step maps
//
//   - 724.2a: PendingTriggers is emptied. That queue IS "triggered but
//     not yet put onto the stack"; a trigger that fires during the
//     steps below lands on it afresh and is drained in the postcombat
//     main phase, which is 724.2f.
//   - 724.2b: exileEntireStackLocked. Spells leave through the shared
//     stack-exit door (exitSpellFromStackLocked, with DropStackMeta),
//     so the item leaves StackMeta with its card — the record a plain
//     exile route leaves behind is the table wedge #1318 is about. The
//     resolving spell has no StackMeta entry any more (the resolution
//     frame took it) and goes through the same route directly; the
//     frame then finds it gone and routes nothing (#489's
//     spellMovedItselfLocked). Copies cease to exist: the "not
//     represented by cards" clause, applied at once rather than at the
//     next check, because no zone this engine has can hold a copy.
//     Ability items are deleted, as CR 800.4a's stack cleanup deletes
//     them — exiled, not countered, so no EventCounterSpell.
//   - 724.2c is FOLDED into the post-resolution check. The process only
//     runs inside a resolution, and the resolution frame runs the
//     SBA + trigger loop immediately after it with nobody having
//     received priority in between. Running the SBAs here as well would
//     need a check that does not drain triggers and does not move the
//     turn on for a departed player, and what it would buy is SBAs seen
//     before the jump instead of after it — observable only through a
//     trigger's position in the APNAP drain, which is the same drain
//     either way.
//   - 724.2d: clearCombatLocked, then the cursor walks forward through
//     advanceCursorLocked — the one seam every step transition takes —
//     WITHOUT running the skipped steps' entry hooks, and the
//     postcombat main phase is entered through runStepEntryHooksLocked
//     like any other step. The engine has no "until end of combat"
//     duration to expire (grep says so), so that clause has nothing to
//     act on today.
//   - 724.2e is the walk above: the end of combat step is never
//     entered, so it is never announced and nothing harvested off
//     EventStepBegan{end_combat} fires. A CR 603.7 delayed trigger
//     scheduled for StepEndCombat is not drained either — it stays
//     queued for the next end of combat step that begins.
//   - 724.2g: PhaseOf(g.Turn.Step) != PhaseCombat returns at once.

// EndCombatPhaseForEffect ends the combat phase (CR 724.2) from inside
// a resolving spell or ability. See the file header for the mapping.
// A no-op outside the combat phase (CR 724.2g).
//
// It must be the LAST thing the resolving effect does, as it is on the
// one printed card: it exiles the object that is resolving, so any
// instruction after it would be run by a spell that is not on the
// stack.
//
// Caller must hold g.mu.
func (g *Game) EndCombatPhaseForEffect() {
	if PhaseOf(g.Turn.Step) != PhaseCombat {
		return
	}
	// 724.2a.
	g.PendingTriggers = nil
	// 724.2b.
	g.exileEntireStackLocked()
	// 724.2d: remove everything from combat, then skip to the next
	// phase. The clear is explicit rather than left to the walk
	// (advanceCursorLocked clears as the cursor LEAVES end_combat)
	// because 724.2d orders it first, and because from end_combat
	// itself the walk would clear it anyway — calling it twice costs a
	// loop over the battlefield and nothing else.
	g.clearCombatLocked()
	for PhaseOf(g.Turn.Step) == PhaseCombat {
		g.advanceCursorLocked()
	}
	g.runStepEntryHooksLocked()
}

// exileEntireStackLocked is CR 724.2b: every object on the stack is
// exiled, including the one resolving. Deterministic order: spell
// cards from the top of the stack down, then ability items by
// descending Seq.
//
// Caller must hold g.mu.
func (g *Game) exileEntireStackLocked() {
	if g.Stack != nil {
		ids := make([]uuid.UUID, 0, len(g.Stack.Cards))
		for i := len(g.Stack.Cards) - 1; i >= 0; i-- {
			ids = append(ids, g.Stack.Cards[i].InstanceID)
		}
		for _, id := range ids {
			g.exileStackObjectLocked(id)
		}
	}
	var abilities []*StackItem
	for _, item := range g.StackMeta {
		if item != nil && item.Kind != StackItemSpell {
			abilities = append(abilities, item)
		}
	}
	// Highest Seq (the top) first, for a deterministic event order.
	for i := 1; i < len(abilities); i++ {
		for j := i; j > 0 && abilities[j].Seq > abilities[j-1].Seq; j-- {
			abilities[j], abilities[j-1] = abilities[j-1], abilities[j]
		}
	}
	for _, item := range abilities {
		delete(g.StackMeta, item.ID)
	}
	g.recomputeSplitSecondLocked()
}

// exileStackObjectLocked exiles one spell card on the stack, whichever
// of the three states it is in: an ordinary stack item, the spell that
// is resolving right now (no StackMeta entry), or a copy.
//
// A route that PAUSES — a commander, whose owner is offered the
// command zone instead (CR 903.9) — leaves the card on the stack under
// its prompt exactly as a paused counter does, and the answer finishes
// the move. The prompt gates the table, so nothing resolves or is cast
// in the postcombat main phase until it is answered.
//
// Caller must hold g.mu.
func (g *Game) exileStackObjectLocked(id uuid.UUID) {
	item := g.StackMeta[id]
	if item == nil {
		if _, resolving, ok := g.resolvingSpellLocked(id); ok {
			item = resolving
		}
	}
	if item != nil && item.IsCopy {
		delete(g.StackMeta, id)
		g.ceaseToExistLocked(id)
		return
	}
	if item != nil && g.StackMeta[id] != nil {
		_ = g.exitSpellFromStackLocked(id, nil, ZoneExile, false)
		return
	}
	// The resolving spell, or a card on the stack with no record at
	// all (the defensive case resolveTopOfStackLocked also handles).
	owner := uuid.Nil
	if item != nil {
		owner = item.Owner
	}
	_, _ = g.routeCardToZoneLocked(zoneRoute{
		CardID:        id,
		Dst:           ZoneExile,
		DstOwner:      owner,
		DropStackMeta: true,
	})
}
