package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bonebind Orator — Creature — Squirrel Warlock Bard {1}{B}, 2/2
// (Bloomburrow, #4467):
//
//	"{3}{B}, Exile this card from your graveyard: Return another target
//	 creature card from your graveyard to your hand."
//
// Theoretical Necromancer's ability, word for word. It works only from
// the graveyard, at instant speed. The target is chosen before the cost
// is paid (CR 602.2b), so "another" keeps the Orator from naming itself
// even though the cost then exiles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c79cf0e5-52dd-4af3-ab06-6cdaab526c23",
		Name:         "Bonebind Orator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}, Exile this card from your graveyard: Return another target creature card from your graveyard to your hand.",
			Cost:    effects.Plus(effects.ManaCost("{3}{B}"), effects.ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Targets: effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.YouOwn(), effects.Creature())),
			Effect:  effects.ReturnTargetGraveyardCardToHand,
		}},
	})
}
