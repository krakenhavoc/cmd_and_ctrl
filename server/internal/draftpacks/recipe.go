// Package draftpacks builds a booster pack for one set from the Scryfall
// dump the server already loads, and says which sets can be drafted
// (ADR 0222 §4, #5437).
//
// A pack here is an APPROXIMATION of the retail booster a set was
// drafted with: the DRAFTABLE booster of the set's own era (a play
// booster for a set of 2024 or later, a draft booster before that; see
// Eras), in slots by rarity at the slot rates below, drawn from the
// cards the dump marks as found in the set's boosters. Set boosters and
// collector boosters were sold beside draft boosters and were never the
// draft product, so they are not modelled at all. It does not model print sheets, foils,
// showcase or other treatments, tokens or ad cards. One printing stands
// for each card.
//
// Nothing here is stored, owned or scarce. A pack is a value the caller
// throws away.
package draftpacks

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// The play-booster slot rates, as ADR 0222 §4 and issue #5437 describe
// the rates Wizards published for play boosters: 14 playable cards, in
// the slots below.
//
// These numbers were NOT re-checked against a Wizards source when this
// package was written (the session that wrote it could not reach the
// web). They are the defaults; a set whose published sheet differs is
// corrected in overrides.json, as data, not here.
const (
	// PlayBoosterSize is the number of cards in a play booster.
	PlayBoosterSize = 14

	// CommonSlots are the slots that are always a common.
	CommonSlots = 6
	// BonusSlots is the one seventh common slot: a common, or 1 time in
	// BonusSheetOneIn a card from the set's bonus sheet. A set with no
	// bonus sheet in the data (Recipe.BonusSheet empty, or the sheet's
	// set absent from the dump) always gets a common there.
	BonusSlots      = 1
	BonusSheetOneIn = 8

	// UncommonSlots are the slots that are always an uncommon.
	UncommonSlots = 3

	// The one rare slot is a mythic rare about 1 time in MythicOneIn,
	// and a rare otherwise.
	MythicOneIn = 7

	// LandSlots is the one land slot: a basic land.
	LandSlots = 1

	// WildcardSlots are the slots that can be a card of any rarity. A
	// retail pack has one non-foil wildcard and one traditional foil of
	// any rarity; foils are not modelled, so the foil is a second
	// wildcard.
	WildcardSlots = 2
)

// The rarity of a wildcard. Wizards described the wildcard as "a card of
// any rarity" and, as far as this package's author could establish
// without a source to hand, did not publish one split for it. These
// weights are therefore an ASSUMPTION, chosen so that a wildcard is
// usually a common and a rare or mythic about one time in six. Correct
// them here, or per set in overrides.json, when a source is checked.
const (
	WildcardCommonWeight       = 6
	WildcardUncommonWeight     = 4
	WildcardRareOrMythicWeight = 2
)

// The draft-booster slot plans, for the sets sold before play boosters.
// A draft booster was 15 playable cards. From Shards of Alara on: 10
// commons, 3 uncommons, 1 rare that is a mythic rare about 1 time in 8,
// and 1 basic land. Before mythic rares existed: 11 commons, 3 uncommons
// and 1 rare, with no basic land slot. Neither has a wildcard, and
// neither has a bonus-sheet slot by default: a set that had a bonus
// sheet or a guaranteed special slot is a per-set entry in
// overrides.json.
//
// Like the play-booster numbers, these come from the owner session's
// memory and were NOT checked against a Wizards source.
const (
	// DraftBoosterSize is the number of cards in a draft booster.
	DraftBoosterSize = 15

	DraftCommonSlots   = 10
	DraftUncommonSlots = 3
	DraftMythicOneIn   = 8
	DraftLandSlots     = 1

	// PreMythicCommonSlots is the common count of a draft booster from
	// before mythic rares: the eleventh common sits where the basic land
	// later went.
	PreMythicCommonSlots = 11
)

// Kind names the retail booster a recipe approximates. Only the boosters
// a set was drafted with have one: set and collector boosters are not
// modelled.
type Kind string

// The booster kinds. The spellings are on the wire (SetStatus.Booster).
const (
	KindPlay  Kind = "play"
	KindDraft Kind = "draft"
)

// Recipe is the slot plan one set's packs are built to: the plan of the
// set's era (Eras), with the set's entry in overrides.json laid over it.
type Recipe struct {
	// Kind is the booster this plan approximates.
	Kind Kind
	// Commons, BonusSlots, Uncommons, Lands and Wildcards are slot
	// counts, and any of them may be zero. A bonus slot is a common slot
	// that may hold a bonus-sheet card instead. The rare slot is not
	// counted here: there is always exactly one.
	Commons    int
	BonusSlots int
	Uncommons  int
	Lands      int
	Wildcards  int
	// BonusSheetOneIn and MythicOneIn are the 1-in-N rates. Zero or
	// less turns the bonus sheet off, or means the set's packs hold no
	// mythic rares at all (the rare slot and a rare wildcard are always
	// a rare).
	BonusSheetOneIn int
	MythicOneIn     int
	// WildcardCommon, WildcardUncommon and WildcardRareOrMythic are the
	// relative weights of a wildcard's rarity.
	WildcardCommon       int
	WildcardUncommon     int
	WildcardRareOrMythic int
	// BonusSheet is the lowercase Scryfall code of the set that holds
	// this set's bonus sheet, empty when the set has none or the data
	// does not carry it.
	BonusSheet string
}

// Size is the number of cards a pack built to r holds.
func (r Recipe) Size() int {
	return r.Commons + r.BonusSlots + r.Uncommons + 1 + r.Lands + r.Wildcards
}

// HasMythics reports whether a pack built to r can hold a mythic rare.
func (r Recipe) HasMythics() bool { return r.MythicOneIn > 0 }

// HasLands reports whether a pack built to r has a basic land slot.
func (r Recipe) HasLands() bool { return r.Lands > 0 }

// UsesBonusSheet reports whether a pack built to r can hold a card of
// the set's bonus sheet: the plan has a bonus slot, a rate for it, and
// names the sheet.
func (r Recipe) UsesBonusSheet() bool {
	return r.BonusSlots > 0 && r.BonusSheetOneIn > 0 && r.BonusSheet != ""
}

// PlayBoosterRecipe is the play-booster slot plan of the constants
// above.
func PlayBoosterRecipe() Recipe {
	return Recipe{
		Kind:                 KindPlay,
		Commons:              CommonSlots,
		BonusSlots:           BonusSlots,
		Uncommons:            UncommonSlots,
		Lands:                LandSlots,
		Wildcards:            WildcardSlots,
		BonusSheetOneIn:      BonusSheetOneIn,
		MythicOneIn:          MythicOneIn,
		WildcardCommon:       WildcardCommonWeight,
		WildcardUncommon:     WildcardUncommonWeight,
		WildcardRareOrMythic: WildcardRareOrMythicWeight,
	}
}

// DraftBoosterRecipe is the draft-booster slot plan of the sets with
// mythic rares, Shards of Alara up to the last set before play boosters.
func DraftBoosterRecipe() Recipe {
	return Recipe{
		Kind:        KindDraft,
		Commons:     DraftCommonSlots,
		Uncommons:   DraftUncommonSlots,
		Lands:       DraftLandSlots,
		MythicOneIn: DraftMythicOneIn,
	}
}

// PreMythicDraftBoosterRecipe is the draft-booster slot plan of the sets
// from before mythic rares: no mythic, no basic land slot.
func PreMythicDraftBoosterRecipe() Recipe {
	return Recipe{
		Kind:      KindDraft,
		Commons:   PreMythicCommonSlots,
		Uncommons: DraftUncommonSlots,
	}
}

// Era is one stretch of release dates whose sets were drafted with the
// same booster.
type Era struct {
	// Name says which booster, for people ("play booster").
	Name string
	// From is the release date, "YYYY-MM-DD", of the first set of the
	// era. The era runs up to the day before the next era's From. The
	// oldest era's From is empty: it has no beginning.
	From string
	// Recipe is the era's slot plan.
	Recipe Recipe
}

// The release dates the eras turn on.
//
// Like the rates, these dates come from the owner session's memory and
// were NOT checked against a Wizards or Scryfall source (the session
// that wrote this could reach neither). A set the date rule gets wrong
// is corrected in overrides.json, as data.
const (
	// MythicRaresFrom is the release date of Shards of Alara, the first
	// set with mythic rares and a basic land in every booster.
	MythicRaresFrom = "2008-10-03"
	// PlayBoostersFrom is the release date of Murders at Karlov Manor,
	// the first set sold in play boosters.
	PlayBoostersFrom = "2024-02-09"
)

// Eras is the table a set's recipe is chosen from by the set's Scryfall
// released_at, oldest era first. It is data: a new booster product is a
// new row.
var Eras = []Era{
	{Name: "draft booster, before mythic rares", From: "", Recipe: PreMythicDraftBoosterRecipe()},
	{Name: "draft booster", From: MythicRaresFrom, Recipe: DraftBoosterRecipe()},
	{Name: "play booster", From: PlayBoostersFrom, Recipe: PlayBoosterRecipe()},
}

// EraFor is the era of a set released on the given "YYYY-MM-DD" date:
// the last era whose From is on or before it.
//
// A set with no date, or one that is not a date, gets the newest era.
// The data gives no way to tell how old such a set is, and the newest
// era is the booster sold today.
func EraFor(released string) Era {
	return eraFor(Eras, released)
}

func eraFor(table []Era, released string) Era {
	released = strings.TrimSpace(released)
	if _, err := time.Parse(time.DateOnly, released); err != nil {
		return table[len(table)-1]
	}
	era := table[0]
	for _, e := range table[1:] {
		// ISO dates sort as strings.
		if e.From <= released {
			era = e
		}
	}
	return era
}

// Override is one set's corrections to the recipe of its era. A nil
// field keeps the era's value. Note says where the correction comes
// from. Kind is "play" or "draft", for a set the date rule files under
// the wrong booster; it changes only the name, so the slot fields are
// corrected beside it.
type Override struct {
	Note                 string  `json:"note,omitempty"`
	Kind                 *string `json:"kind,omitempty"`
	BonusSheet           *string `json:"bonus_sheet,omitempty"`
	BonusSheetOneIn      *int    `json:"bonus_sheet_one_in,omitempty"`
	MythicOneIn          *int    `json:"mythic_one_in,omitempty"`
	Commons              *int    `json:"commons,omitempty"`
	BonusSlots           *int    `json:"bonus_slots,omitempty"`
	Uncommons            *int    `json:"uncommons,omitempty"`
	Lands                *int    `json:"lands,omitempty"`
	Wildcards            *int    `json:"wildcards,omitempty"`
	WildcardCommon       *int    `json:"wildcard_common,omitempty"`
	WildcardUncommon     *int    `json:"wildcard_uncommon,omitempty"`
	WildcardRareOrMythic *int    `json:"wildcard_rare_or_mythic,omitempty"`
}

// Apply returns r with o's set fields laid over it.
func (o Override) Apply(r Recipe) Recipe {
	set := func(dst *int, src *int) {
		if src != nil && *src >= 0 {
			*dst = *src
		}
	}
	if o.Kind != nil {
		if k := Kind(strings.ToLower(strings.TrimSpace(*o.Kind))); k == KindPlay || k == KindDraft {
			r.Kind = k
		}
	}
	if o.BonusSheet != nil {
		r.BonusSheet = strings.ToLower(strings.TrimSpace(*o.BonusSheet))
	}
	set(&r.BonusSheetOneIn, o.BonusSheetOneIn)
	set(&r.MythicOneIn, o.MythicOneIn)
	set(&r.Commons, o.Commons)
	set(&r.BonusSlots, o.BonusSlots)
	set(&r.Uncommons, o.Uncommons)
	set(&r.Lands, o.Lands)
	set(&r.Wildcards, o.Wildcards)
	set(&r.WildcardCommon, o.WildcardCommon)
	set(&r.WildcardUncommon, o.WildcardUncommon)
	set(&r.WildcardRareOrMythic, o.WildcardRareOrMythic)
	return r
}

// overridesFile is the shape of overrides.json.
type overridesFile struct {
	Sets map[string]Override `json:"sets"`
}

//go:embed overrides.json
var overridesJSON []byte

// overrides is overrides.json by lowercase set code, parsed at init. A
// malformed file is a build-time mistake, so it panics, and
// TestOverridesFileParses keeps that from reaching a deploy.
var overrides = mustParseOverrides(overridesJSON)

func parseOverrides(b []byte) (map[string]Override, error) {
	var f overridesFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("draftpacks: overrides.json: %w", err)
	}
	out := make(map[string]Override, len(f.Sets))
	for code, o := range f.Sets {
		out[strings.ToLower(strings.TrimSpace(code))] = o
	}
	return out, nil
}

func mustParseOverrides(b []byte) map[string]Override {
	m, err := parseOverrides(b)
	if err != nil {
		panic(err)
	}
	return m
}

// RecipeFor is the slot plan for a set released on the given
// "YYYY-MM-DD" date: the recipe of the set's era (EraFor), with the
// set's entry in overrides.json, if it has one, laid over it.
func RecipeFor(set, released string) Recipe {
	return recipeFor(overrides, set, released)
}

func recipeFor(table map[string]Override, set, released string) Recipe {
	r := EraFor(released).Recipe
	if o, ok := table[strings.ToLower(strings.TrimSpace(set))]; ok {
		r = o.Apply(r)
	}
	return r
}
