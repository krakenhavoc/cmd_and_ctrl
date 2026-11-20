package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Theoretical Necromancer — Creature — Vampire Warlock {2}{B}, 4/1:
//
//	"{3}{B}, Exile this card from your graveyard: Return another target
//	 creature card from your graveyard to your hand."
//
// An instant-speed graveyard ability. The target is chosen before the
// cost is paid, so "another" keeps this card from naming itself even
// though the cost then exiles it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6c5ed756-b999-49f7-9a52-fb8196202f0f",
		Name:         "Theoretical Necromancer",
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
