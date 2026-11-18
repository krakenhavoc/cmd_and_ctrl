package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mystic Skull // Mystic Monstrosity — Artifact {2} // Artifact Creature —
// Construct 5/6 (Midnight Hunt, Innistrad: Double Feature):
//
//	Mystic Skull
//	  "{1}, {T}: Add one mana of any color.
//	   {5}, {T}: Transform this artifact."
//	Mystic Monstrosity
//	  "Lands you control have "{T}: Add one mana of any color.""
//
// The transform is an ordinary activation that turns the permanent over in
// place (ADR 0079). The back face's grant is Chromatic Lantern's layer-6
// grant to every land you control.
//
// No simplification.
func init() {
	const oracle = "1a7e367e-1d8d-463d-a35f-4bad44a5ebea"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Mystic Skull",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{W|U|B|R|G}",
			Label:    "{1}, {T}: Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}: Transform this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{5}"), TapCost()),
			Effect:  Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Mystic Monstrosity",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(mysticMonstrosityGrant)},
		Static: []game.StaticAbility{Labelled(GrantAbilities(landsYouControl, mysticMonstrosityGrant),
			"Lands you control have \"{T}: Add one mana of any color.\"")},
	})
}

const mysticMonstrosityGrant = "mystic-monstrosity/any-color"
