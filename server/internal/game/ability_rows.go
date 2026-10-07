package game

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ability_rows.go — #2219: the abilities an art tile names in its
// chips. A tile has no rules text, so the server lists what the object
// can do, by kind, with a short label for each, and the client counts
// them (⚡ triggered, ◆ static, ↻ activated). The client never reads
// oracle text.
//
// What a row is. One declaration in the catalog: an entry of a
// CardDef's Triggered or Activated list, an entry of its Static list,
// or an entry of one of the other slots that hold a static ability of
// the object (a cost reduction, a replacement effect, "your opponents
// can't gain life"). staticRowSlots names those slots, and
// TestEveryCardDefSlotIsClassifiedForAbilityRows keeps the list
// complete: a new CardDef field fails until it is placed.
//
// Current, not printed. The rows are the object's abilities right
// now, read through the accessors the engine itself uses:
//
//   - a CR 613.1f removal ("loses all abilities") takes the object's
//     own rows, and a layer-6 grant the layer pass let survive is
//     listed (catalogAbilityKeyOf, the same key the trigger harvest
//     and the activation path read);
//   - an ADR 0071 designation gate is applied, so a Case's solved
//     ability appears when it is solved;
//   - a face-down permanent has no rows (CR 708.2).
//
// Off the battlefield the object has no layered characteristic, so the
// rows are what the card prints — which is what a hand tile wants.
//
// Keyword abilities are left out. Their own chips (`abilities`,
// `protection`) already show them, or the keyword says what the row
// does: a trigger with a Keyword (prowess, cascade, ward), the static
// synthesised from Spec.PrintedKeywords (StaticAbility.Keywords), and
// the activated keyword abilities whose whole meaning is the keyword
// (equip, crew, cycling — IsKeywordActivatedAbility). An ability that
// only borrows a keyword's frame for its own effect ("Exhaust — {4}:
// Earthbend 4", "Boast — …") is the card's own and is counted.
//
// Mana abilities are left out as well: ADR 0105's drop pip and the
// mana menu are their surface, and a mana row on every land would be
// the noise ADR 0105 §4 refused.

// AbilityRowKind is a row's chip: one of the three below.
type AbilityRowKind string

const (
	AbilityRowTriggered AbilityRowKind = "triggered"
	AbilityRowStatic    AbilityRowKind = "static"
	AbilityRowActivated AbilityRowKind = "activated"
)

// AbilityRow is one non-keyword ability of an object.
type AbilityRow struct {
	Kind AbilityRowKind
	// Label is the row's player-facing text: a triggered row's Key and
	// an activated row's Label (the names a stack item carries), a
	// static slot's own Label, with the object's name dropped from the
	// front ("Inferno Titan — 3 damage…" reads "3 damage…"). A row the
	// catalog gives no text is described by what it is ("Replacement
	// effect", "Power/toughness effect").
	Label string
	// Purpose is the row's declared purpose (ADR 0126 §6): a triggered
	// row's TriggeredAbility.Purpose, an activated row's
	// ActivatedAbilityShape.Purpose. Zero for a static row and for a
	// row that declares none.
	Purpose Purpose
}

// AbilityRowsOf lists the object's non-keyword triggered, static and
// activated abilities right now, in that order. Nil for an object with
// none — an uncatalogued card, a vanilla creature, a face-down
// permanent.
func AbilityRowsOf(c Card) []AbilityRow { return abilityRowsOf(&c) }

func abilityRowsOf(c *Card) []AbilityRow {
	if c.faceDownPermanent() {
		return nil
	}
	names := abilityRowNames(c)
	var out []AbilityRow
	for _, t := range triggersOf(c) {
		if t.Keyword != "" {
			continue
		}
		out = append(out, AbilityRow{Kind: AbilityRowTriggered, Label: abilityRowLabel(t.Key, names, "Triggered ability"), Purpose: t.Purpose})
	}
	out = appendStaticRows(out, c, names)
	for _, a := range activatedAbilitiesOf(c) {
		if IsKeywordActivatedAbility(a) {
			continue
		}
		out = append(out, AbilityRow{Kind: AbilityRowActivated, Label: abilityRowLabel(a.Label, names, "Activated ability"), Purpose: a.Purpose})
	}
	return out
}

// appendStaticRows adds the object's static rows: its Static list, then
// each slot in staticRowSlots, in CardDef order.
func appendStaticRows(out []AbilityRow, c *Card, names []string) []AbilityRow {
	d := catalogDef(catalogAbilityKeyOf(c))
	if d == nil {
		return out
	}
	statics := activeOnly(c, d.Static, func(s StaticAbility) Designation { return s.ActiveWhen })
	for _, s := range statics {
		// A keyword row is on the keyword chips already, and a
		// ContinuesAfterRemoval row is the later-layer half of an
		// ability another row starts (ADR 0067 §2): one ability, one
		// row.
		if len(s.Keywords) > 0 || s.ContinuesAfterRemoval {
			continue
		}
		out = append(out, AbilityRow{Kind: AbilityRowStatic, Label: abilityRowLabel(s.Label, names, staticLayerLabel(s))})
	}
	def := reflect.ValueOf(d).Elem()
	for i, slot := range staticRowSlots {
		out = slot.appendRows(out, c, def.Field(staticRowSlotIndex[i]), names)
	}
	return out
}

// staticLayerLabel describes a Static row that declares no Label by
// what CR 613.1 calls the effect it generates.
func staticLayerLabel(s StaticAbility) string {
	switch s.Layer {
	case Layer1Copy:
		return "Copy effect"
	case Layer2Control:
		return "Control-changing effect"
	case Layer3Text:
		return "Text-changing effect"
	case Layer4Type:
		return "Type-changing effect"
	case Layer5Color:
		return "Color-changing effect"
	case Layer6Ability:
		switch {
		case s.RemovesAbilities:
			return "Removes abilities"
		case len(s.GrantAbilities) > 0:
			return "Grants abilities"
		}
		return "Ability-changing effect"
	case Layer7PT:
		return "Power/toughness effect"
	}
	return "Static ability"
}

// staticRowSlot is one CardDef slot whose entries are static abilities
// of the object.
type staticRowSlot struct {
	field string
	// fallback describes an entry that has no Label of its own (or
	// whose Label is only the card's name).
	fallback string
	// ownLabel reads each entry's Label field. False for a slot whose
	// Label is not a sentence a player reads (BlockRule's is the
	// clause's tail, "creatures with greater power").
	ownLabel bool
	// label, when set, answers for a slot that is not a list of
	// structs: a bool, a count, a keyword list. It returns one label per
	// row, nil for none.
	label func(v reflect.Value) []string
}

// staticRowSlots is every CardDef slot, other than Static, that holds
// a static ability of the object, in CardDef order. Each entry of a
// list slot is one row. A slot's entries are designation-gated when
// their type declares ActiveWhen.
var staticRowSlots = []staticRowSlot{
	{field: "Replacements", fallback: "Replacement effect", ownLabel: true},
	{field: "EntersWithCountersFromCast", fallback: "Enters with counters"},
	{field: "TriggerDoublers", fallback: "Abilities trigger an additional time", ownLabel: true},
	{field: "TriggerSuppressors", fallback: "Stops abilities from triggering", ownLabel: true},
	{field: "SpellsYouCastHaveDelve", label: boolRow("Spells you cast have delve.")},
	{field: "CostModifiers", fallback: "Cost-changing effect", ownLabel: true},
	{field: "ExhaustPermissions", fallback: "Exhaust abilities can be activated again", ownLabel: true},
	{field: "AttackTaxes", fallback: "Attack tax", ownLabel: true},
	{field: "BlockRules", fallback: "Blocking restriction"},
	{field: "AttackLimits", fallback: "Limits how many creatures can attack"},
	{field: "ExertOnAttack", fallback: "You may exert this creature as it attacks."},
	{field: "HexproofBypasses", fallback: "Ignores hexproof"},
	{field: "WardSuppressions", fallback: "Stops ward from triggering"},
	{field: "TargetingRestrictions", fallback: "Targeting restriction", ownLabel: true},
	{field: "SpecialActionGrants", fallback: "Grants a special action"},
	{field: "UntapStep", fallback: "Untaps during other untap steps", ownLabel: true},
	{field: "UntapStepRestrictions", fallback: "Untap restriction", ownLabel: true},
	{field: "UntapCaps", fallback: "Limits untapping", ownLabel: true},
	{field: "UntapOptOuts", fallback: "May choose not to untap", ownLabel: true},
	{field: "DrawStep", fallback: "Extra draw", ownLabel: true},
	{field: "CastRestrictions", fallback: "Casting restriction", ownLabel: true},
	{field: "LandPlayRestrictions", fallback: "Land-play restriction", ownLabel: true},
	{field: "ActivationRestrictions", fallback: "Activation restriction", ownLabel: true},
	{field: "ActivationTimings", fallback: "Changes when abilities can be activated", ownLabel: true},
	{field: "SpellsCantBeCountered", fallback: "Spells can't be countered", ownLabel: true},
	{field: "HandSize", fallback: "Changes maximum hand size"},
	{field: "ManaPool", fallback: "Changes how mana empties"},
	{field: "PlayerKeywords", label: playerKeywordRows},
	{field: "PlayerLifeTotalLocked", label: boolRow("Your life total can't change.")},
	{field: "DamageCantBePrevented", fallback: "Damage can't be prevented"},
	{field: "CantGainLife", fallback: "Life can't be gained", ownLabel: true},
	{field: "DamageAsThough", fallback: "Changes how damage is dealt"},
	{field: "AnyColorSpend", fallback: "Spend mana as though it were any color"},
	{field: "LegendRuleExemptions", fallback: "The legend rule doesn't apply"},
	{field: "OpponentEffectProtections", fallback: "Opponents' effects can't make you discard or sacrifice"},
	{field: "GameEndGates", fallback: "Changes who can win or lose"},
	{field: "AdditionalLandPlays", label: additionalLandPlayRows},
	{field: "CastPermissions", fallback: "Lets you play cards from another zone"},
	{field: "GatedCastPermissions", fallback: "Lets you play cards from another zone"},
	{field: "CastTimings", fallback: "Changes when spells can be cast"},
	{field: "GrantedAlternativeCosts", fallback: "Offers another way to pay for spells"},
	{field: "LibraryTopVisible", label: libraryTopRows},
}

// staticRowSlotIndex is each staticRowSlots entry's field index in
// CardDef, resolved once: a slot naming no field is a programming error
// and stops the boot.
var staticRowSlotIndex = func() []int {
	t := reflect.TypeOf(CardDef{})
	idx := make([]int, len(staticRowSlots))
	for i, slot := range staticRowSlots {
		f, ok := t.FieldByName(slot.field)
		if !ok {
			panic("game: staticRowSlots names no CardDef field " + slot.field)
		}
		idx[i] = f.Index[0]
	}
	return idx
}()

var designationType = reflect.TypeOf(Designation{})

// appendRows adds one slot's rows.
func (s staticRowSlot) appendRows(out []AbilityRow, c *Card, v reflect.Value, names []string) []AbilityRow {
	if !v.IsValid() || v.IsZero() {
		return out
	}
	if s.label != nil {
		for _, l := range s.label(v) {
			out = append(out, AbilityRow{Kind: AbilityRowStatic, Label: l})
		}
		return out
	}
	if v.Kind() != reflect.Slice {
		return append(out, AbilityRow{Kind: AbilityRowStatic, Label: s.fallback})
	}
	for i := 0; i < v.Len(); i++ {
		e := v.Index(i)
		label := ""
		if e.Kind() == reflect.Struct {
			if gate := e.FieldByName("ActiveWhen"); gate.IsValid() && gate.Type() == designationType {
				if d := gate.Interface().(Designation); d.IsGate() && !d.Active(*c) {
					continue
				}
			}
			if s.ownLabel {
				if f := e.FieldByName("Label"); f.IsValid() && f.Kind() == reflect.String {
					label = f.String()
				}
			}
		}
		out = append(out, AbilityRow{Kind: AbilityRowStatic, Label: abilityRowLabel(label, names, s.fallback)})
	}
	return out
}

func boolRow(label string) func(reflect.Value) []string {
	return func(v reflect.Value) []string {
		if v.Kind() == reflect.Bool && v.Bool() {
			return []string{label}
		}
		return nil
	}
}

func playerKeywordRows(v reflect.Value) []string {
	kws, _ := v.Interface().([]string)
	out := make([]string, 0, len(kws))
	for _, kw := range kws {
		out = append(out, "You have "+kw+".")
	}
	return out
}

func additionalLandPlayRows(v reflect.Value) []string {
	n, _ := v.Interface().(int)
	switch {
	case n == 1:
		return []string{"You may play an additional land on each of your turns."}
	case n > 1:
		return []string{fmt.Sprintf("You may play %d additional lands on each of your turns.", n)}
	}
	return nil
}

func libraryTopRows(v reflect.Value) []string {
	switch vis, _ := v.Interface().(LibraryTopVisibility); vis {
	case LibraryTopOwner:
		return []string{"You may look at the top card of your library any time."}
	case LibraryTopRevealed:
		return []string{"Play with the top card of your library revealed."}
	}
	return nil
}

// abilityRowNames is every name a row's label may begin with: the
// object's name now (a copy's is the copied card's), its printed name,
// and each face's.
func abilityRowNames(c *Card) []string {
	names := make([]string, 0, 3+len(c.Faces))
	add := func(n string) {
		if n == "" {
			return
		}
		names = append(names, n)
		// A split card's catalog keys name a half ("Fire // Ice").
		if strings.Contains(n, " // ") {
			names = append(names, strings.Split(n, " // ")...)
		}
	}
	if c.effective != nil {
		add(c.effective.Name)
	}
	add(c.Name)
	for _, f := range c.Faces {
		add(f.Name)
	}
	return names
}

// abilityRowLabelSeparators are the ways a catalog label joins the
// card's name to what the row does.
var abilityRowLabelSeparators = []string{" — ", ": ", " - "}

// abilityRowLabel is raw with the object's own name dropped from the
// front, capitalised, or fallback when nothing is left.
func abilityRowLabel(raw string, names []string, fallback string) string {
	l := stripLeadingName(strings.TrimSpace(raw), names)
	if l == "" {
		return fallback
	}
	r, size := utf8.DecodeRuneInString(l)
	if unicode.IsLower(r) {
		return string(unicode.ToUpper(r)) + l[size:]
	}
	return l
}

func stripLeadingName(l string, names []string) string {
	for _, n := range names {
		if l == n {
			return ""
		}
		for _, sep := range abilityRowLabelSeparators {
			if strings.HasPrefix(l, n+sep) {
				return strings.TrimSpace(l[len(n)+len(sep):])
			}
		}
	}
	return l
}

// activatedKeywordAbilities are the CR 702 keyword abilities that are
// activated abilities whose whole effect the keyword defines. A row
// whose label starts with one of them is that keyword (cycling and its
// typecycling forms are matched by suffix, below).
var activatedKeywordAbilities = []string{
	"equip", "crew", "station", "unearth", "ninjutsu", "commander ninjutsu",
	"embalm", "eternalize", "scavenge", "transmute", "transfigure",
	"reconfigure", "outlast", "level up", "fortify", "reinforce", "encore",
	"saddle", "craft",
}

// IsKeywordActivatedAbility reports whether an activated row is a
// keyword ability whose effect the keyword defines — equip, crew,
// cycling and the rest of activatedKeywordAbilities — read off the
// row's Equip and Cycling bits and the head of its label (the words
// before the cost). An ability word ("Channel — …") or a keyword frame
// around the card's own effect ("Exhaust — {4}: Earthbend 4") is not
// one.
func IsKeywordActivatedAbility(a ActivatedAbilityShape) bool {
	if a.Equip || a.Cycling {
		return true
	}
	head := strings.ToLower(a.Label)
	if i := strings.IndexAny(head, "{(:—"); i >= 0 {
		head = head[:i]
	}
	head = strings.TrimSpace(head)
	if head == "" {
		return false
	}
	if strings.HasSuffix(head, "cycling") {
		return true
	}
	for _, kw := range activatedKeywordAbilities {
		if head == kw || strings.HasPrefix(head, kw+" ") {
			return true
		}
	}
	return false
}
