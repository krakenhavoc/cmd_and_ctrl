package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rugged Prairie — Land (EDHREC rank 282):
//
//	"{T}: Add {C}.
//	 {R/W}, {T}: Add {R}{R}, {R}{W}, or {W}{W}."
//
// The Guildpact filter land, Boros half. Batch 02's own triage filed
// the whole five-card remainder of this cycle (Rugged Prairie,
// Cascade Bluffs, Flooded Grove, Fetid Heath, Twilight Mire) under
// "a mana component inside a mana ability's cost" — the Signet gap,
// closed by #356. Mystic Gate and its four Shadowmoor siblings
// registered the shape first; this is the same body with Boros
// colours.
//
// The hybrid {R/W} parses as a requirement either colour satisfies,
// paid from the pool (the engine has no auto-tap into a mana
// ability's cost). The three-way output is two independent pipe slots —
// "{R|W}{R|W}" resolves to RR, RW, WR or WW, exactly the printed
// three outcomes (WR and RW are the same choice twice), one colour
// pick per slot.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8e7641e1-e814-4d5a-9cb3-71ad2f4ceee8",
		Name:         "Rugged Prairie",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{R/W}"},
				Produced: "{R|W}{R|W}",
				Label:    "{R/W}, {T}: Add {R}{R}, {R}{W}, or {W}{W}",
			},
		},
	})
}
