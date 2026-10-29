package game

import "github.com/google/uuid"

// entry_overwrite.go — a card put onto the battlefield as a different
// type (#4506, ADR 0210).
//
//	Supper for Spiders: "Put onto the battlefield under your control all
//	creature cards in your opponents' graveyards that were put there
//	from the battlefield this turn. They are Food artifacts with '{2},
//	{T}, Sacrifice this artifact: You gain 3 life.' (They lose all other
//	types and subtypes.)"
//	Tom, Bert, and William: "… return them to the battlefield. They're
//	an artifact. (They're no longer a creature.)"
//
// The putting effect says what the object IS as it arrives, so the
// object is that from the moment it lands (CR 614.1c, "[This permanent]
// enters as …"): it is never a creature on the battlefield, and an
// "enters" trigger that looks for a creature finds an artifact. The
// overwrite is one ScopedEffect record (ADR 0041 phase 3), registered
// as the permanent lands and before any event about the entry goes out,
// so it is plain data and the existing layer vocabulary:
//
//   - ModSetTypes, layer 4 (CR 205.1a's set, not CR 205.1b's add): the
//     card types become Types, every subtype whose card type is gone
//     goes with it, Subtypes replace their own set, and supertypes stay
//     (CR 205.4b).
//   - ModGrantAbilities, layer 6 (ADR 0093): the ability the effect
//     names ("with '{2}, {T}, Sacrifice this artifact: You gain 3
//     life.'") as a catalog bundle. The object keeps its own abilities.
//
// The record is pinned to the permanent by its ObjectEpoch, taken as it
// lands and before the zone-move event stamps it (PinObjectByEpoch, the
// pin a stolen permanent spell's record takes, ADR 0104), with no stated
// duration (CR 611.2a). It lasts as long as that object is on the
// battlefield: a zone change makes a new object (CR 400.7) the pin does
// not match, and the sweep collects the record. A copy of the permanent
// copies its printed values, not this type change (CR 707.2).
//
// The entry window sees it too (CR 614.12): "To determine which
// replacement effects apply and how they apply, check the
// characteristics of the permanent as it would exist on the
// battlefield, taking into account replacement effects that have
// already modified how it enters". The dry run (entryCharacteristicsLocked)
// applies the overwrite, and while a window gathers its replacements
// LookupCardForEffect answers for the entering card with that dry run's
// characteristics, so "creatures your opponents control enter tapped"
// does not apply to a Food (entryOverwriteViewLocked).

// EntersAs is a putting effect's "it's a <types> with <ability>" for
// the object it puts onto the battlefield (ZoneEntryOptions.EntersAs).
// Plain data: it rides a paused entry's event and its batch's options
// to disk.
type EntersAs struct {
	// Types are the card types the permanent has (at least one).
	Types []string `json:"types"`
	// Subtypes are the subtypes it has, each one of Types' (CR 205.3d):
	// Supper for Spiders' "Food". Empty keeps the subtypes CR 205.1a
	// leaves it.
	Subtypes []string `json:"subtypes,omitempty"`
	// Grants are ADR 0093 bundle keys for the abilities the effect gives
	// it. Empty for an effect that gives none.
	Grants []string `json:"grants,omitempty"`
	// Label names the record for logs and tests.
	Label string `json:"label,omitempty"`
}

// mods is the overwrite as the ScopedEffect mods it registers, in one
// record so they share one timestamp and are one effect (CR 613.7).
func (e *EntersAs) mods() []Mod {
	mods := []Mod{SetTypesMod(e.Types, e.Subtypes)}
	if len(e.Grants) > 0 {
		mods = append(mods, GrantAbilitiesMod(e.Grants...))
	}
	return mods
}

// Problem is the registration and restore check: "" when the overwrite
// is sound.
func (e *EntersAs) Problem() string {
	if e == nil {
		return ""
	}
	if len(e.Types) == 0 {
		return "an entry overwrite names no card type"
	}
	for _, m := range e.mods() {
		if p := becomeNamedModProblem(m); p != "" {
			return p
		}
		if m.Kind == ModGrantAbilities && len(m.Grants) == 0 {
			return "an entry overwrite grants an empty bundle key"
		}
	}
	return ""
}

func copyEntersAs(e *EntersAs) *EntersAs {
	if e == nil {
		return nil
	}
	return &EntersAs{Types: copyStrings(e.Types), Subtypes: copyStrings(e.Subtypes), Grants: copyStrings(e.Grants), Label: e.Label}
}

// entryOverwriteRecord is the overwrite as the record it registers for
// the permanent `id` with ObjectEpoch `epoch`, with no timestamp. Used
// by the landing and by the dry run.
func entryOverwriteRecord(e *EntersAs, id uuid.UUID, epoch int) ScopedEffect {
	return ScopedEffect{
		Affected: []AffectedObject{PinObjectByEpoch(id, epoch)},
		Mods:     e.mods(),
		Label:    e.Label,
	}
}

// applyEntryOverwriteLocked registers the overwrite on the permanent
// `entered` as it lands: after the CR 400.7 reset, and before the
// zone-move event and EventETB in announceEntryLocked, so nothing ever
// sees the object without it. A nil overwrite is every other entry.
//
// Caller must hold g.mu (write).
func (g *Game) applyEntryOverwriteLocked(entered uuid.UUID, ev *ReplacementEvent) {
	e := ev.entersAs
	if e == nil {
		return
	}
	perm, ok := g.battlefieldCardLocked(entered)
	if !ok || perm.ObjectEpoch <= 0 {
		return
	}
	if e.Problem() != "" {
		// Refused when the batch began (startPreparedEntryBatchLocked),
		// so only a hand-edited restore point reaches here: the card
		// enters as itself rather than the table stopping.
		return
	}
	g.registerScopedEffectLocked(ev.Source,
		[]AffectedObject{PinObjectByEpoch(entered, perm.ObjectEpoch)},
		e.mods(),
		g.PinnedToEpoch(IndefiniteDuration(), entered),
		e.Label, timeNowUnixNano())
	perm.effective = nil
}

// entryOverwriteView is the entering card as the CR 614.12 dry run sees
// it, for the span of one replacement gather (entryOverwriteViewLocked).
// Never set between calls, so it is neither snapshotted nor cloned.
type entryOverwriteView struct {
	card      uuid.UUID
	effective Characteristic
}

// entryOverwriteViewLocked sets the view for an entry event that carries
// an overwrite and returns the function that clears it. It does nothing
// for any other event, or for a caller holding only a read lock (the dry
// run writes the battlefield slice for its span).
//
// Caller must hold g.mu (write unless readOnly).
func (g *Game) entryOverwriteViewLocked(ev *ReplacementEvent, readOnly bool) func() {
	if readOnly || ev == nil || ev.entersAs == nil || ev.Kind != RepEventMove || ev.NewZone != ZoneBattlefield ||
		g.entryView != nil || g.Battlefield == nil || g.Battlefield.Contains(ev.CardID) {
		return func() {}
	}
	perm, ok := g.entryCharacteristicsLocked(ev)
	if !ok || perm.effective == nil {
		return func() {}
	}
	g.entryView = &entryOverwriteView{card: ev.CardID, effective: *perm.effective}
	return func() { g.entryView = nil }
}

// viewEnteringCardLocked is LookupCardForEffect's answer for the card
// whose entry window is gathering with an overwrite: the card in its
// zone, with the characteristics it would have on the battlefield.
func (g *Game) viewEnteringCardLocked(c Card) Card {
	v := g.entryView
	if v == nil || v.card != c.InstanceID {
		return c
	}
	ch := v.effective
	c.effective = &ch
	return c
}
