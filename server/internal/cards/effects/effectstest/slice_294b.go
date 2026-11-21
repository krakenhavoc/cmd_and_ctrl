package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// slice_294b_test.go — the granted-mana-ability slice (ADR 0093,
// #294 slice 294-b). Each card gets a main-behaviour test through
// game.ManaAbilitiesWithOrigins + ActivateManaAbility, the shape
// TestGrantAbilitiesGivesACreatureTheBundlesManaAbility
// (ability_grant_test.go) established, plus a refused-case test for
// every card whose grant is conditional.
//
// Chromatic Lantern, Cryptolith Rite, Gemhide Sliver, Great Divide
// Guide, Paradise Mantle, Rishkar Peema Renegade and The World Tree
// were also built in this slice, but #1566 (ADR 0093 PR 2, server
// half) shipped them to develop first with equivalent behaviour and
// its own coverage (granted_ability_cards_test.go) while this branch
// was in flight; this file keeps only the eight cards unique to this
// slice.

const (
	EnduringVitalityOracle      = "3577c47e-76d3-4659-b922-31c4b74be3a0"
	ElvenChorusOracle           = "4d5df2d9-06bf-4bec-9be0-dabe75a858fb"
	NightOfTheSweetsRevengeCard = "4bcf4f12-470e-4f73-a19a-643a20084e16"
	AlchemistsTalentOracle      = "5a2dfff5-9dba-42f5-bcca-918c16f23807"
	AbundantGrowthOracle        = "947a2665-2f4d-4193-8768-118f85334549"
	WrennAndRealmbreakerOracle  = "4566fb92-448e-4b3f-9045-9d74323c35d1"
	MachineGodsEffigyOracle     = "64ebdd6f-acde-4aab-a86b-2798bad5f70c"
	ResonatingLuteOracle        = "9263287d-f782-4318-9ab8-e2e5f8107f5e"
)

// grantedManaAbilityFor finds the first GRANTED mana ability on `id`
// and returns its index and ref, or ok=false when there is none.
func GrantedManaAbilityFor(g *game.Game, id uuid.UUID) (idx int, ref string, ok bool) {
	var abs []game.ManaAbilityShape
	var origins game.AbilityOrigins
	g.ReadSnapshot(func() {
		c, found := g.LookupCardForEffect(id)
		if !found {
			return
		}
		abs, origins = game.ManaAbilitiesWithOrigins(c)
	})
	for i := range abs {
		if origins.At(i).Granted() {
			return i, origins.Ref(i), true
		}
	}
	return 0, "", false
}
