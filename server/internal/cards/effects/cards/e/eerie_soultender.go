package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a3e4d14b-818a-4a4b-8462-1de9ad844941",
		Name:         "Eerie Soultender",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Eerie Soultender — mill three cards", effects.Do(effects.MillCards{N: 3})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}, Exile this card from your graveyard: Return another target creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{4}{B}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
			Effect:  effects.ReturnTargetGraveyardCardToHand,
		}},
	})
}
