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
	// Types are the card types the permanent has, in place of its own
	// (CR 205.1a). Empty only for an overwrite that adds types instead
	// (AddTypes).
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

	// AddTypes are card types the permanent has IN ADDITION to its own
	// (CR 205.1b, layer 4): Excava, the Risen Past's "It's a 1/1 Spirit
	// creature with flying in addition to its other types" (#4834). An
	// overwrite either sets its types (Types) or adds them, never both.
	AddTypes []string `json:"addTypes,omitempty"`
	// AddSubtypes are subtypes it has in addition to its own (layer 4),
	// beside AddTypes: Excava's "Spirit".
	AddSubtypes []string `json:"addSubtypes,omitempty"`
	// Keywords are keyword abilities the effect gives it (layer 6):
	// Excava's "with flying". Canonical lowercase tokens.
	Keywords []string `json:"keywords,omitempty"`
	// BasePT, when set, is the base power and toughness the effect gives
	// it (layer 7b, CR 613.4b): Excava's "1/1". Nil keeps the card's own.
	BasePT *EntersAsPT `json:"basePT,omitempty"`
	// CreatureTypes, when set, REPLACE the object's creature types and
	// leave its card types and other subtypes alone (layer 4, CR 205.1b's
	// "[creature type] artifact creature"): assimilate's "It's a Borg
	// artifact creature and loses all other creature types" is
	// AddTypes ["Artifact"] with CreatureTypes ["Borg"]. Beside AddTypes
	// only; an overwrite that sets Types names its subtypes in Subtypes.
	CreatureTypes []string `json:"creatureTypes,omitempty"`
	// Name, when set, is the name the effect gives it (layer 3, CR
	// 612.8): Princess Yue's "She's a land named Moon." Empty keeps the
	// card's own.
	Name string `json:"name,omitempty"`
	// LosesAllAbilities is "and it loses all abilities" (layer 6, CR
	// 613.1f): Retched Wretch's return. It lasts as long as the object is
	// on the battlefield, like the rest of the record, and a card with only
	// this field changes no types.
	LosesAllAbilities bool `json:"losesAllAbilities,omitempty"`
}

// EntersAsPT is an entry overwrite's base power and toughness.
type EntersAsPT struct {
	Power     int `json:"power"`
	Toughness int `json:"toughness"`
}

// mods is the overwrite as the ScopedEffect mods it registers, in one
// record so they share one timestamp and are one effect (CR 613.7).
func (e *EntersAs) mods() []Mod {
	var mods []Mod
	if e.Name != "" {
		mods = append(mods, SetNameMod(e.Name))
	}
	if len(e.Types) > 0 {
		mods = append(mods, SetTypesMod(e.Types, e.Subtypes))
	}
	if len(e.AddTypes) > 0 {
		mods = append(mods, AddTypesMod(e.AddTypes...))
	}
	if len(e.AddSubtypes) > 0 {
		mods = append(mods, AddSubtypesMod(e.AddSubtypes...))
	}
	if len(e.CreatureTypes) > 0 {
		mods = append(mods, SetCreatureTypesMod(e.CreatureTypes...))
	}
	if len(e.Keywords) > 0 {
		mods = append(mods, AddKeywordsMod(e.Keywords...))
	}
	if len(e.Grants) > 0 {
		mods = append(mods, GrantAbilitiesMod(e.Grants...))
	}
	if e.BasePT != nil {
		mods = append(mods, SetBasePowerMod(e.BasePT.Power), SetBaseToughnessMod(e.BasePT.Toughness))
	}
	if e.LosesAllAbilities {
		mods = append(mods, LoseAllAbilitiesMod())
	}
	return mods
}

// Problem is the registration and restore check: "" when the overwrite
// is sound.
func (e *EntersAs) Problem() string {
	if e == nil {
		return ""
	}
	if len(e.Types) == 0 && len(e.AddTypes) == 0 && !e.LosesAllAbilities {
		return "an entry overwrite names no card type"
	}
	if len(e.Types) > 0 && (len(e.AddTypes) > 0 || len(e.AddSubtypes) > 0 || len(e.CreatureTypes) > 0) {
		return "an entry overwrite both sets and adds types"
	}
	if len(e.Types) == 0 && len(e.Subtypes) > 0 {
		return "an entry overwrite that adds types names its subtypes in AddSubtypes"
	}
	for _, m := range e.mods() {
		if p := becomeNamedModProblem(m); p != "" {
			return p
		}
		if p := setCreatureTypesModProblem(m); p != "" {
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
	out := &EntersAs{Types: copyStrings(e.Types), Subtypes: copyStrings(e.Subtypes), Grants: copyStrings(e.Grants), Label: e.Label,
		AddTypes: copyStrings(e.AddTypes), AddSubtypes: copyStrings(e.AddSubtypes), Keywords: copyStrings(e.Keywords),
		CreatureTypes: copyStrings(e.CreatureTypes), Name: e.Name, LosesAllAbilities: e.LosesAllAbilities}
	if e.BasePT != nil {
		pt := *e.BasePT
		out.BasePT = &pt
	}
	return out
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
