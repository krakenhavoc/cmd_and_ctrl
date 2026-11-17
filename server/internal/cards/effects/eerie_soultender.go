package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eerie Soultender — Creature — Spirit Cleric {2}{B}, 3/1 (Dominaria United,
// S82 DMU batch 4):
//
//	"When this creature enters, mill three cards.
//	 {4}{B}, Exile this card from your graveyard: Return another target
//	 creature card from your graveyard to your hand."
//
// Aftermath Analyst's enters-and-mill, and Bonebind Orator's graveyard
// ability with {4}{B}. The target is chosen before the cost is paid
// (CR 602.2b), so "another" keeps it from naming itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a3e4d14b-818a-4a4b-8462-1de9ad844941",
		Name:         "Eerie Soultender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Eerie Soultender — mill three cards", Do(MillCards{N: 3})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{4}{B}, Exile this card from your graveyard: Return another target creature card from your graveyard to your hand.",
			Cost:    Plus(ManaCost("{4}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: Another(TargetCardInGraveyard("another target creature card from your graveyard", YouOwn(), Creature())),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
