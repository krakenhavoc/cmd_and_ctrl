package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Solemn Doomguide — Creature — Tiefling Cleric {3}{B}{B}:
//
//	"Flying
//	 Each creature card in your graveyard that's a Cleric, Rogue, Warrior,
//	 and/or Wizard has unearth {1}{B}. ({1}{B}: Return the card to the
//	 battlefield. The creature gains haste. Exile it at the beginning of
//	 the next end step or if it would leave the battlefield. Unearth only
//	 as a sorcery.)"
//
// A graveyard grant (#3233, ADR 0093's amendment of 2026-10-13) of the
// ordinary Unearth row, to the creature cards with one of four subtypes.
// A card in a graveyard has its printed subtypes, which is what
// HasSubtype reads there (a changeling is every type, CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c3e4a6ce-4a36-43a4-8e80-3c16a9101b79",
		Name:            "Solemn Doomguide",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Grants: []AbilityGrant{{
			Key:       "solemn-doomguide/unearth",
			Zone:      game.ZoneGraveyard,
			Activated: []ActivatedAbility{Unearth("{1}{B}")},
			Text:      "Unearth {1}{B}",
		}},
		Static: []game.StaticAbility{
			Labelled(GrantAbilitiesInZone(game.ZoneGraveyard, doomguideRecipient, "solemn-doomguide/unearth"),
				"Each creature card in your graveyard that's a Cleric, Rogue, Warrior, and/or Wizard has unearth {1}{B}."),
		},
	})
}

// doomguideRecipient is a creature card you own that's a Cleric, Rogue,
// Warrior and/or Wizard (CardsYouOwn's owner test). The four types are
// words of the source's text (CR 612.2, #3984).
func doomguideRecipient(card *game.Card, _ *game.Game, source *game.Card) bool {
	if source == nil || card.Owner != source.Controller || !card.IsCreature() {
		return false
	}
	for _, t := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		if card.HasSubtype(source.TextSubtype(t)) {
			return true
		}
	}
	return false
}
