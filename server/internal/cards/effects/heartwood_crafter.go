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
// THE SIMPLIFICATION. Mana restrictions are decided from the object
// being paid for, and the spend context carries no "cast from" zone, so
// "can't be spent to cast spells from your hand" has no tag. The mana is
// restricted to activated-ability costs instead, which is a strict
// subset of what is printed: it also can't cast a spell from the
// graveyard, exile or command zone, as the printed text allows.
func init() {
	const id = "c8b3a070-408e-4a8f-8597-6476eeec0a5f"
	Register(Spec{
		OracleID:     id,
		Name:         "Heartwood Crafter",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Its mana can only pay for activated abilities, so it can't help cast a spell from your graveyard, exile or command zone either.",
		},
		Replacements: []game.ReplacementEffect{SelfEntersPrepared()},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{C}",
			Label:        "Add {C} (can't be spent to cast spells from your hand)",
			Restrictions: []string{ManaRestrictActivate},
		}},
	})
	Register(Spec{
		OracleID:     id + "#1",
		Name:         "Soul Tether",
		Completeness: CompletenessFull,
		OnResolve:    soulTetherResolve,
	})
}
