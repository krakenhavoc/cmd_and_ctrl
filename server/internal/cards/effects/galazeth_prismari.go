package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Galazeth Prismari — Legendary Creature — Elder Dragon {2}{U}{R}, 3/4
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying
//	 When Galazeth Prismari enters, create a Treasure token.
//	 Artifacts you control have "{T}: Add one mana of any color. Spend
//	 this mana only to cast an instant or sorcery spell.""
//
// The grant is Chromatic Lantern's layer-6 bundle (ADR 0093) over every
// artifact you control, its Treasure included, with Ore-Rich
// Stalactite's spend restriction on the mana: only to cast a spell, and
// only an instant or sorcery one. An artifact keeps its own abilities
// beside the granted one, so a Treasure can still be sacrificed for
// unrestricted mana.
//
// No simplification.
const galazethPrismariGrant = "galazeth-prismari/any-color-for-instants-and-sorceries"

func init() {
	Register(Spec{
		OracleID:        "f9217532-f7cf-4436-b476-72d97a3147dc",
		Name:            "Galazeth Prismari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Galazeth Prismari — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Grants: []AbilityGrant{{
			Key: galazethPrismariGrant,
			Mana: []ManaAbility{{
				Cost:         ManaAbilityCost{Tap: true},
				Produced:     "{W|U|B|R|G}",
				Label:        "Add one mana of any color. Spend this mana only to cast an instant or sorcery spell",
				Restrictions: []string{ManaRestrictCast, game.ManaRestrictAnyType("Instant", "Sorcery")},
			}},
			Text: "{T}: Add one mana of any color. Spend this mana only to cast an instant or sorcery spell.",
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.Controller == source.Controller
		}, galazethPrismariGrant), "Artifacts you control have \"{T}: Add one mana of any color. Spend this mana only to cast an instant or sorcery spell.\"")},
	})
}
