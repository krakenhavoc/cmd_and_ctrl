package game

import (
	"strings"

	"github.com/google/uuid"
)

// epic.go — CR 702.50, epic (#6141, S82).
//
//	702.50a Epic represents two spell abilities, one of which creates a
//	delayed triggered ability. "Epic" means "For the rest of the game,
//	you can't cast spells," and "At the beginning of each of your
//	upkeeps for the rest of the game, copy this spell except for its
//	epic ability. If the spell has any targets, you may choose new
//	targets for the copy." See rule 707.10.
//
//	702.50b A player can't cast spells once a spell with epic they
//	control resolves, but effects (such as the epic ability itself) can
//	still put copies of spells onto the stack.
//
// Both halves are built from seams that already exist:
//
//   - The first is a per-player cast ban (cast_ban.go) with no end:
//     CastBanOutright for IndefiniteDuration, sourced by the epic spell.
//     It lives on the player's Statics, so it clones and snapshots with
//     every other granted ban, and CastGateLocked already refuses each
//     later cast. A copy is not cast (CR 707.10), so the ban never stops
//     the upkeep copies.
//
//   - The second is a step-conditioned delayed trigger (delayed.go) at
//     the controller's upkeep with an Indefinite duration, which CR
//     603.7b's "stated duration" keeps queued after each firing
//     (repeatsAtStep). What it copies is the spell as it resolved: the
//     card and its stack item, carried on the trigger as their snapshot
//     records (DelayedTrigger.CopiesSpell). Plain data, so the trigger
//     adds no route from Game to a closure and a table waiting on it is
//     a restore point. The original card is in a graveyard by the first
//     upkeep, or exiled, or shuffled away, and the copy does not care:
//     CR 707.10 copies the spell's copiable values and the choices made
//     for it (X, modes, targets, a kicker), all of which the record has.
//
// "Except for its epic ability" (CR 707.9b) is the copy losing the epic
// keyword from its copiable values (PrintedValues.LoseKeyword), so the
// copy's own resolution reaches EpicForEffect, finds the keyword lost and
// does nothing more: a copy grants no second ban and no second trigger.
// A Twincast of the ORIGINAL spell does have epic (a plain copy), and
// sets up its own pair for the Twincast's controller, which is what CR
// 702.50a reads.

// KeywordEpic is the keyword the importer stamps from Scryfall's
// "Epic", compared without case.
const KeywordEpic = "Epic"

// EpicSpell is a spell as it resolved, kept by an epic trigger for the
// rest of the game: the card and its stack item as their snapshot
// records. Plain data only (ADR 0041 phase 3), never mutated after the
// trigger is scheduled.
type EpicSpell struct {
	Card cardSnapshot      `json:"card"`
	Item stackItemSnapshot `json:"item"`
}

// epicCopyBody is the delayed trigger's body: copy the carried spell.
var epicCopyBody = SimpleDelayedBody("epic/copy", resolveEpicCopy)

// EpicForEffect is the epic ability of the spell `spellID`, which is
// resolving now (CR 702.50a): its controller can't cast spells for the
// rest of the game, and at the beginning of each of their upkeeps a copy
// of the spell, as it is now, is put on the stack. Returns false and does
// nothing when the spell is not on the stack or resolving, or when it has
// lost its epic ability (a copy the epic trigger made, CR 707.9b).
//
// Call it from the spell's own resolution, before anything that could
// pause it, so the spell is still the resolving object. Caller must hold
// g.mu (write), which a resolution frame already does.
func (g *Game) EpicForEffect(spellID uuid.UUID, label string) bool {
	src, item, ok := g.stackSpellLocked(spellID)
	if !ok || item == nil || item.Kind != StackItemSpell {
		return false
	}
	for _, lost := range src.LostKeywords {
		if strings.EqualFold(lost, KeywordEpic) {
			return false
		}
	}
	controller := item.Controller
	if label == "" {
		label = src.Name + " — epic"
	}
	g.GrantCastBanForEffect(controller, CastBanRule{Kind: CastBanOutright}, label, src.InstanceID, IndefiniteDuration())
	cen := &ContinuationCensus{}
	rec := &EpicSpell{
		Card: snapshotCard(src, cen),
		Item: snapshotStackItemAs(item, src.OracleID, cen),
	}
	forever := IndefiniteDuration()
	id := uuid.New()
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		ID:                 id,
		Controller:         controller,
		SourceCardID:       src.InstanceID,
		Label:              src.Name + " — epic: copy this spell",
		At:                 StepUpkeep,
		ControllerTurnOnly: true,
		Duration:           &forever,
		Body:               epicCopyBody,
		// The fired item names its trigger, which holds the spell.
		Params:      EffectParams{Object: ObjectRef{ID: id}},
		CopiesSpell: rec,
	})
	return true
}

// resolveEpicCopy is the upkeep trigger resolving: copy the spell it
// carries, except for its epic ability, under the trigger's controller,
// who may choose new targets for the copy (CR 702.50a, CR 707.10c). The
// copy goes through the one copy instruction (copySpellTimesLocked), so a
// "copy it again" replacement sees it as it sees any copy.
func resolveEpicCopy(g *Game, item *StackItem) error {
	rec := g.epicSpellForItemLocked(item)
	if rec == nil {
		return nil
	}
	src := restoreCard(&rec.Card)
	spell, _ := restoreStackItem(&rec.Item)
	if spell == nil {
		return nil
	}
	spell.Controller = item.Controller
	g.copySpellTimesLocked(src, spell, item.Controller, 1, true, func(v *PrintedValues) {
		v.LoseKeyword(KeywordEpic)
	}, nil)
	return nil
}

// epicSpellForItemLocked finds the spell record of the epic trigger that
// put `item` on the stack, which its params name. The trigger stays queued
// for the rest of the game, so it is still there when its item resolves;
// a trigger that has gone (its controller left the game) copies nothing.
// Caller must hold g.mu.
func (g *Game) epicSpellForItemLocked(item *StackItem) *EpicSpell {
	if item == nil {
		return nil
	}
	if id := item.Params.Object.ID; id != uuid.Nil {
		for _, dt := range g.DelayedTriggers {
			if dt != nil && dt.ID == id && dt.CopiesSpell != nil {
				return dt.CopiesSpell
			}
		}
	}
	return nil
}

// cloneEpicSpell copies the record for Clone. It is never mutated after
// it is stored, but a value copy keeps the clone independent anyway.
func cloneEpicSpell(r *EpicSpell) *EpicSpell {
	if r == nil {
		return nil
	}
	out := *r
	return &out
}
