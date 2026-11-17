package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cormela, Glamour Thief — Legendary Creature — Vampire Rogue
// {1}{U}{B}{R}, 2/4 (Streets of New Capenna):
//
//	"Haste
//	 {1}, {T}: Add {U}{B}{R}. Spend this mana only to cast instant and/or
//	 sorcery spells.
//	 When Cormela dies, return up to one target instant or sorcery card
//	 from your graveyard to your hand."
//
// The mana ability is a real mana ability with a {1} component in its cost
// (CR 605.1a) and the cast-only, instant-or-sorcery restriction on all
// three mana (CR 106.6). The dies trigger takes "up to one" target from
// your graveyard, so with none chosen it does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3cd9bac9-0abc-43d8-9f84-94f965f7a2e0",
		Name:            "Cormela, Glamour Thief",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{U}{B}{R}",
			Label:    "{1}, {T}: Add {U}{B}{R}. Spend this mana only to cast instant and/or sorcery spells.",
			Restrictions: []string{
				ManaRestrictCast,
				ManaRestrictAnyType("Instant", "Sorcery"),
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Cormela, Glamour Thief — return up to one target instant or sorcery card from your graveyard to your hand",
				returnTargetGraveyardCardToHand),
				TargetCardInGraveyard("up to one target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())).WithCount(0, 1)),
		},
	})
}
