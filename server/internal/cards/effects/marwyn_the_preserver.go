package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "eb03f6c1-c94f-4f09-b2e9-a55ce5533653",
		Name:         "Marwyn, the Preserver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			b16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsLand() && target.Controller == source.Controller
			}, "hexproof"),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}: Return target land card from your graveyard to your hand",
			Cost:    ManaCost("{2}"),
			Targets: TargetCardInGraveyard("target land card from your graveyard", Land(), YouOwn()),
			Effect:  returnTargetGraveyardCardToHand,
		}},
	})
}
