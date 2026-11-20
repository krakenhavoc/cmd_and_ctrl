package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Marwyn, the Preserver — Legendary Creature — Elf Druid {1}{G}, 3/2:
//
//	"Lands you control have hexproof.
//	 {2}: Return target land card from your graveyard to your hand."
//
// The hexproof is a plain Layer 6 grant to the controller's lands, the
// shape Avatar Kyoshi uses. The return is a targeted ability, so with no
// land card in the graveyard it can't be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eb03f6c1-c94f-4f09-b2e9-a55ce5533653",
		Name:         "Marwyn, the Preserver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			}, "hexproof"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Return target land card from your graveyard to your hand",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCardInGraveyard("target land card from your graveyard", effects.Land(), effects.YouOwn()),
			Effect:  effects.ReturnTargetGraveyardCardToHand,
		}},
	})
}
