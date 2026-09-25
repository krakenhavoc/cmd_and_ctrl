package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// verges.go — the Duskmourn / Foundations "verge" cycle:
//
//	"{T}: Add {A}."
//	"{T}: Add {B}. Activate only if you control a <X> or a <Y>."
//
// Six of the ten, the six played across the decks that reach this
// file.
//
// A verge is two separate mana abilities, the second carrying a CR
// 602.1b activation restriction. That restriction is ManaAbility.
// Condition (the S32 mana-pipeline gate Temple of the False God uses):
// checked before the tap is paid, so a verge with no Plains or Island
// beside it offers only its first colour and a failed activation taps
// nothing. The land types are read post-layer (MatchLandSubtype), so
// a shockland or a Urborg-granted Swamp counts, and the verge's own
// type line — it has no land types — never satisfies it.
//
// These first shipped with only the unconditional ability, because
// mana abilities had no condition slot; the slot landed in S32 (#352)
// and the caveat outlived it.
//
// No simplification.
func init() {
	// `free` is the colour of the unconditional ability; `gated` the
	// colour behind "Activate only if you control an <x> or a <y>".
	for _, t := range []struct {
		oracleID string
		name     string
		free     string
		gated    string
		x, y     string
		gate     string // the printed restriction, for the label
	}{
		{"2b8144a0-08d2-4c28-9fd7-5d90f90105e4", "Bleachbone Verge", "B", "W", "Plains", "Swamp", "a Plains or a Swamp"},
		{"f1e9abfb-c3c8-483e-b446-5c2afc9f6394", "Floodfarm Verge", "W", "U", "Plains", "Island", "a Plains or an Island"},
		{"d71bda4c-3dee-4398-8fd0-f77d8743b887", "Gloomlake Verge", "U", "B", "Island", "Swamp", "an Island or a Swamp"},
		{"977c2f33-b622-4172-9efb-7f523becd32b", "Blazemire Verge", "B", "R", "Swamp", "Mountain", "a Swamp or a Mountain"},
		{"510a6ac5-f098-4145-ac07-771b1b6f7cdf", "Riverpyre Verge", "R", "U", "Island", "Mountain", "an Island or a Mountain"},
		{"cce328b9-6100-417e-9ddf-808bbe3e3bc5", "Hushwood Verge", "G", "W", "Forest", "Plains", "a Forest or a Plains"},
	} {
		x, y := MatchLandSubtype(t.x), MatchLandSubtype(t.y)
		Register(Spec{
			OracleID:     t.oracleID,
			Name:         t.name,
			Completeness: CompletenessFull,
			ManaAbilities: []ManaAbility{
				{
					Cost:     ManaAbilityCost{Tap: true},
					Produced: "{" + t.free + "}",
					Label:    "Add {" + t.free + "}",
				},
				{
					Cost:     ManaAbilityCost{Tap: true},
					Produced: "{" + t.gated + "}",
					Label:    "Add {" + t.gated + "} (only if you control " + t.gate + ")",
					Condition: ControlsAtLeast(1, func(c game.Card) bool {
						return x(c) || y(c)
					}),
				},
			},
		})
	}
}
