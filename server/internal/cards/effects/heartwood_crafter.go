package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartwood Crafter // Soul Tether — Creature — Elf Artificer {G}, 1/1
// // Sorcery {2}{R/G} (preparation card, CR 722):
//
//	"This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)
//	 {T}: Add {C}. This mana can't be spent to cast spells from your
//	 hand."
//
//	Soul Tether — "Create a Heartwood token."
//
// The mana carries ManaRestrictNotFromHand (#2811): the spend context
// knows the zone a spell is cast from, so the mana pays for activated
// abilities and for spells cast from anywhere but the hand.
//
// No simplification.
func init() {
	const id = "c8b3a070-408e-4a8f-8597-6476eeec0a5f"
	Register(Spec{
		OracleID:     id,
		Name:         "Heartwood Crafter",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{C}",
			Label:        "Add {C} (can't be spent to cast spells from your hand)",
			Restrictions: []string{ManaRestrictNotFromHand},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Soul Tether",
		Completeness: CompletenessFull,
		OnResolve:    soulTetherResolve,
	})
}
