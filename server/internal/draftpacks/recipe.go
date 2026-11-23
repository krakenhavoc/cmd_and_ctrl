// Package draftpacks builds a booster pack for one set from the Scryfall
// dump the server already loads, and says which sets can be drafted
// (ADR 0222 §4, #5437).
//
// A pack here is an APPROXIMATION of the retail booster a set was
// drafted with: the DRAFTABLE booster of the set's own era (a play
// booster for a set of 2024 or later, a draft booster before that; see
// Eras), in slots by rarity at the slot rates below, drawn from the
// cards the dump marks as found in the set's boosters (or, for a set
// Scryfall has flagged nowhere, the printings of its main run: see
// collect, #5689), plus the set's bonus slots (a bonus sheet, The List,
// Special Guests) at the rates MTGJSON's collation gives them. Set
// boosters and collector boosters were sold beside draft boosters and
// were never the draft product, so they are not modelled at all. It
// does not model print sheets, foils (a play booster's foil slot is
// drawn as a plain card), showcase or other treatments, tokens or ad
// cards. One printing stands for each card.
//
// Nothing here is stored, owned or scarce. A pack is a value the caller
// throws away.
package draftpacks

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// The play-booster slot plan. Every number here is read from MTGJSON v5's
// collation of the play booster, data.booster.play in each set's file
// (https://mtgjson.com/api/v5/<SET>.json.gz, fetched 2026-11-18; the
// copy that was read is draft/boosters.json.gz on the work/s82-cards
// branch). Every play-booster set MTGJSON collates (MKM, OTJ, BLB, DSK,
// FDN, DFT, TDM, FIN, EOE, SPM, TLA) has the same shape: 7 commons, 3
// uncommons, 1 rare or mythic, 1 land, 1 wildcard and 1 traditional foil,
// 14 cards. A set's own rates, rarity splits and bonus slots are its
// entry in overrides.json, as data, not here.
const (
	// PlayBoosterSize is the number of cards in a play booster: the
	// slots below, added up.
	PlayBoosterSize = 14

	// CommonSlots are the slots that are a common. A bonus slot (see
	// BonusSlot) takes the place of one of them when it opens.
	CommonSlots = 7

	// UncommonSlots are the slots that are always an uncommon.
	UncommonSlots = 3

	// The one rare slot is a mythic rare about 1 time in MythicOneIn,
	// and a rare otherwise: 120 of 840 on the rare sheet of MKM, OTJ,
	// BLB, DSK, FDN and DFT.
	MythicOneIn = 7

	// LandSlots is the one land slot: a basic land.
	LandSlots = 1

	// WildcardSlots is the one non-foil card of any rarity.
	WildcardSlots = 1

	// FoilSlots is the one traditional foil of any rarity. Foils are not
	// modelled: the foil slot is drawn as a plain card, by its own
	// rarity split.
	FoilSlots = 1
)

// The rarity split of a play booster's wildcard and of its foil, per
// 10,000: the Foundations (FDN) wildcard and foil sheets in MTGJSON,
// each card's weight summed by rarity. Foundations is the core set, so
// it stands for a set with no entry of its own; every collated play
// booster set has its own split in overrides.json.
const (
	WildcardCommonWeight   = 1850
	WildcardUncommonWeight = 6070
	WildcardRareWeight     = 1783
	WildcardMythicWeight   = 297

	FoilCommonWeight   = 6002
	FoilUncommonWeight = 3146
	FoilRareWeight     = 730
	FoilMythicWeight   = 122
)

// The draft-booster slot plans, for the sets sold before play boosters.
// A draft booster is 15 cards: 10 commons, 3 uncommons, 1 rare or mythic
// and 1 basic land, the shape of every draft booster MTGJSON collates
// (data.booster.draft of WOE, LCI, MOM, ONE, BRO, DMU, SNC, NEO, VOW,
// MID and AFR, fetched 2026-11-18). One draft booster in three or so
// swaps a common for a foil; foils are not modelled, so that slot stays
// a common. Neither plan has a wildcard or a bonus slot by default: a
// set with one is a per-set entry in overrides.json.
const (
	// DraftBoosterSize is the number of cards in a draft booster.
	DraftBoosterSize = 15

	DraftCommonSlots   = 10
	DraftUncommonSlots = 3
	DraftLandSlots     = 1

	// DraftMythicOneIn is the mythic rate of a draft booster with no
	// entry of its own: 1 in 8, the rate Wizards announced with mythic
	// rares in Shards of Alara (2008). The file above holds no set that
	// old to check it against. The collated sets read 1 in 6 to 1 in 8
	// from their own sheets, and each has its rate in overrides.json.
	DraftMythicOneIn = 8

	// PreMythicCommonSlots is the common count of a draft booster from
	// before mythic rares: the eleventh common sits where the basic land
	// later went. MTGJSON's file above holds no set that old; this is
	// the 15-card booster of 11 commons, 3 uncommons and 1 rare that
	// Wizards sold before Shards of Alara.
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

// RaritySplit is the relative weight of each rarity in a slot that can
// be any rarity. The weights need not add up to anything; overrides.json
// gives them per 10,000.
type RaritySplit struct {
	Common   int `json:"common"`
	Uncommon int `json:"uncommon"`
	Rare     int `json:"rare"`
	Mythic   int `json:"mythic"`
}

// Source is one set of printings a bonus slot draws from: the whole set
// (From, To and Numbers all empty), the collector numbers From to To of
// it, or the printings of it listed in Numbers.
type Source struct {
	// Set is the lowercase Scryfall code of the set the printings are in.
	Set string `json:"set"`
	// From and To are an inclusive range of plain collector numbers.
	From int `json:"from,omitempty"`
	To   int `json:"to,omitempty"`
	// Numbers are collector numbers as Scryfall spells them ("MKM-12"
	// on The List).
	Numbers []string `json:"numbers,omitempty"`
}

// wholeSet reports whether the source names a set and nothing narrower.
func (s Source) wholeSet() bool { return s.From == 0 && s.To == 0 && len(s.Numbers) == 0 }

// BonusSlot is a slot that holds a card from outside the set's main run
// (a bonus sheet, The List, Special Guests) 1 time in OneIn, in the place
// of a slot of the kind Replaces. OneIn 1 is every pack.
type BonusSlot struct {
	// Name is the sheet's name, for people ("Enchanting Tales").
	Name string `json:"name"`
	// OneIn is the 1-in-N rate. Zero or less turns the slot off.
	OneIn int `json:"one_in"`
	// Replaces is the slot the card takes the place of: "common" when
	// empty, or "uncommon", "land", "wildcard" or "foil". The rare slot
	// is never replaced.
	Replaces Slot `json:"replaces,omitempty"`
	// Sources are the printings the slot draws from, one card each.
	Sources []Source `json:"sources"`
}

// Slot is the slot the bonus card takes the place of.
func (b BonusSlot) Slot() Slot {
	if b.Replaces == "" {
		return SlotCommon
	}
	return b.Replaces
}

// replaceable are the slots a bonus slot may take the place of.
var replaceable = map[Slot]bool{SlotCommon: true, SlotUncommon: true, SlotLand: true, SlotWildcard: true, SlotFoil: true}

// Recipe is the slot plan one set's packs are built to: the plan of the
// set's era (Eras), with the set's entry in overrides.json laid over it.
type Recipe struct {
	// Kind is the booster this plan approximates.
	Kind Kind
	// Commons, Uncommons, Lands, Wildcards and Foils are slot counts, and
	// any of them may be zero. The rare slot is not counted here: there
	// is always exactly one. A bonus slot takes the place of one of these
	// and adds none.
	Commons   int
	Uncommons int
	Lands     int
	Wildcards int
	Foils     int
	// MythicOneIn is the rare slot's 1-in-N mythic rate. Zero or less
	// means the set's packs hold no mythic rares at all (the rare slot is
	// always a rare, and a wildcard or foil never a mythic).
	MythicOneIn int
	// Wildcard and Foil are the rarity splits of the wildcard and the
	// foil slot.
	Wildcard RaritySplit
	Foil     RaritySplit
	// Bonus are the set's bonus slots, in the order they are rolled.
	Bonus []BonusSlot
}

// Size is the number of cards a pack built to r holds.
func (r Recipe) Size() int {
	return r.Commons + r.Uncommons + 1 + r.Lands + r.Wildcards + r.Foils
}

// HasMythics reports whether a pack built to r can hold a mythic rare.
func (r Recipe) HasMythics() bool { return r.MythicOneIn > 0 }

// HasLands reports whether a pack built to r has a basic land slot.
func (r Recipe) HasLands() bool { return r.Lands > 0 }

// UsesBonusSheet reports whether a pack built to r can hold a card from
// outside the set's main run: some bonus slot can open (bonusOn).
func (r Recipe) UsesBonusSheet() bool {
	for _, b := range r.Bonus {
		if r.bonusOn(b) {
			return true
		}
	}
	return false
}

// bonusOn reports whether a bonus slot can open in a pack built to r: it
// has a rate and a source, and r has a slot of the kind it replaces.
func (r Recipe) bonusOn(b BonusSlot) bool {
	return b.OneIn > 0 && len(b.Sources) > 0 && replaceable[b.Slot()] && r.slotCount(b.Slot()) > 0
}

// slotCount is how many slots of a kind r has: what a bonus slot can
// take the place of.
func (r Recipe) slotCount(s Slot) int {
	switch s {
	case SlotCommon:
		return r.Commons
	case SlotUncommon:
		return r.Uncommons
	case SlotLand:
		return r.Lands
	case SlotWildcard:
		return r.Wildcards
	case SlotFoil:
		return r.Foils
	}
	return 0
}

// PlayBoosterRecipe is the play-booster slot plan of the constants
// above, with no bonus slot: each set's is in overrides.json.
func PlayBoosterRecipe() Recipe {
	return Recipe{
		Kind:        KindPlay,
		Commons:     CommonSlots,
		Uncommons:   UncommonSlots,
		Lands:       LandSlots,
		Wildcards:   WildcardSlots,
		Foils:       FoilSlots,
		MythicOneIn: MythicOneIn,
		Wildcard: RaritySplit{
			Common: WildcardCommonWeight, Uncommon: WildcardUncommonWeight,
			Rare: WildcardRareWeight, Mythic: WildcardMythicWeight,
		},
		Foil: RaritySplit{
			Common: FoilCommonWeight, Uncommon: FoilUncommonWeight,
			Rare: FoilRareWeight, Mythic: FoilMythicWeight,
		},
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
// They are the released_at of the two sets in Scryfall's set table (ALA
// and MKM). A set the date rule gets wrong is corrected in
// overrides.json, as data.
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
// corrected beside it. Bonus, when present, replaces the era's bonus
// slots (an empty list removes them).
type Override struct {
	Note        string       `json:"note,omitempty"`
	Kind        *string      `json:"kind,omitempty"`
	MythicOneIn *int         `json:"mythic_one_in,omitempty"`
	Commons     *int         `json:"commons,omitempty"`
	Uncommons   *int         `json:"uncommons,omitempty"`
	Lands       *int         `json:"lands,omitempty"`
	Wildcards   *int         `json:"wildcards,omitempty"`
	Foils       *int         `json:"foils,omitempty"`
	Wildcard    *RaritySplit `json:"wildcard,omitempty"`
	Foil        *RaritySplit `json:"foil,omitempty"`
	Bonus       *[]BonusSlot `json:"bonus,omitempty"`
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
	set(&r.MythicOneIn, o.MythicOneIn)
	set(&r.Commons, o.Commons)
	set(&r.Uncommons, o.Uncommons)
	set(&r.Lands, o.Lands)
	set(&r.Wildcards, o.Wildcards)
	set(&r.Foils, o.Foils)
	if o.Wildcard != nil {
		r.Wildcard = *o.Wildcard
	}
	if o.Foil != nil {
		r.Foil = *o.Foil
	}
	if o.Bonus != nil {
		r.Bonus = normalBonus(*o.Bonus)
	}
	return r
}

// normalBonus lowercases the set codes of bonus slots read from JSON.
func normalBonus(in []BonusSlot) []BonusSlot {
	out := make([]BonusSlot, len(in))
	for i, b := range in {
		b.Replaces = Slot(strings.ToLower(strings.TrimSpace(string(b.Replaces))))
		srcs := make([]Source, len(b.Sources))
		for j, s := range b.Sources {
			s.Set = strings.ToLower(strings.TrimSpace(s.Set))
			srcs[j] = s
		}
		b.Sources = srcs
		out[i] = b
	}
	return out
}

// overridesFile is the shape of overrides.json.
type overridesFile struct {
	Sets map[string]Override `json:"sets"`
}

// validate refuses a bonus slot that could not open: one that replaces a
// slot a bonus card cannot stand in, or names no printings.
func (o Override) validate() error {
	if o.Bonus == nil {
		return nil
	}
	for _, b := range normalBonus(*o.Bonus) {
		if !replaceable[b.Slot()] {
			return fmt.Errorf("bonus slot %q replaces %q, want common, uncommon, land, wildcard or foil", b.Name, b.Replaces)
		}
		if len(b.Sources) == 0 {
			return fmt.Errorf("bonus slot %q names no printings", b.Name)
		}
		for _, s := range b.Sources {
			if s.Set == "" || s.From < 0 || s.To < s.From || (len(s.Numbers) > 0 && s.To > 0) {
				return fmt.Errorf("bonus slot %q has a malformed source %+v", b.Name, s)
			}
		}
	}
	return nil
}

//go:embed overrides.json
var overridesJSON []byte

// overrides is overrides.json by lowercase set code, parsed at init. A
// malformed file is a build-time mistake, so it panics, and
// TestOverridesFileParses keeps that from reaching a deploy.
var overrides = mustParseOverrides(overridesJSON)

func parseOverrides(b []byte) (map[string]Override, error) {
	var f overridesFile
	dec := json.NewDecoder(bytes.NewReader(b))
	// A misspelt field would silently keep the era's value.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("draftpacks: overrides.json: %w", err)
	}
	out := make(map[string]Override, len(f.Sets))
	for code, o := range f.Sets {
		code = strings.ToLower(strings.TrimSpace(code))
		if err := o.validate(); err != nil {
			return nil, fmt.Errorf("draftpacks: overrides.json: %s: %w", code, err)
		}
		out[code] = o
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
