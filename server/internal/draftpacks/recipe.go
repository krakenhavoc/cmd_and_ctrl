// Package draftpacks builds a booster pack for one set from the Scryfall
// dump the server already loads, and says which sets can be drafted
// (ADR 0222 §4, #5437).
//
// A pack here is an APPROXIMATION of a retail play booster: slots by
// rarity at the slot rates below, drawn from the cards the dump marks as
// found in the set's boosters. It does not model print sheets, foils,
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
	// PackSize is the number of cards in a pack.
	PackSize = 14

	// CommonSlots are the slots that are always a common.
	CommonSlots = 6
	// The seventh common slot is a common, or 1 time in
	// BonusSheetOneIn a card from the set's bonus sheet. A set with no
	// bonus sheet in the data (Recipe.BonusSheet empty, or the sheet's
	// set absent from the dump) always gets a common there.
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

// Recipe is the slot plan one set's packs are built to. DefaultRecipe is
// the plan above; overrides.json changes fields of it for a set.
type Recipe struct {
	// Commons, Uncommons, Lands and Wildcards are slot counts. The
	// seventh common slot (the one that may be a bonus-sheet card) and
	// the rare slot are not counted here: there is always one of each.
	Commons   int
	Uncommons int
	Lands     int
	Wildcards int
	// BonusSheetOneIn and MythicOneIn are the 1-in-N rates. Zero or
	// less turns the bonus sheet, or mythics in the rare slot, off.
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
	return r.Commons + 1 + r.Uncommons + 1 + r.Lands + r.Wildcards
}

// DefaultRecipe is the play-booster slot plan of the constants above.
func DefaultRecipe() Recipe {
	return Recipe{
		Commons:              CommonSlots,
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

// Override is one set's corrections to DefaultRecipe. A nil field keeps
// the default. Note says where the correction comes from.
type Override struct {
	Note                 string  `json:"note,omitempty"`
	BonusSheet           *string `json:"bonus_sheet,omitempty"`
	BonusSheetOneIn      *int    `json:"bonus_sheet_one_in,omitempty"`
	MythicOneIn          *int    `json:"mythic_one_in,omitempty"`
	Commons              *int    `json:"commons,omitempty"`
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
	if o.BonusSheet != nil {
		r.BonusSheet = strings.ToLower(strings.TrimSpace(*o.BonusSheet))
	}
	set(&r.BonusSheetOneIn, o.BonusSheetOneIn)
	set(&r.MythicOneIn, o.MythicOneIn)
	set(&r.Commons, o.Commons)
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

// RecipeFor is the slot plan for a set: DefaultRecipe with the set's
// entry in overrides.json, if it has one, laid over it.
func RecipeFor(set string) Recipe {
	return recipeFor(overrides, set)
}

func recipeFor(table map[string]Override, set string) Recipe {
	r := DefaultRecipe()
	if o, ok := table[strings.ToLower(strings.TrimSpace(set))]; ok {
		r = o.Apply(r)
	}
	return r
}
