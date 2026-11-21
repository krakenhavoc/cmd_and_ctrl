package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Writ of Return — Sorcery {3}{B}{B}:
//
//	"Return target creature card from your graveyard to the battlefield
//	 tapped.
//	 Cipher (Then you may exile this spell card encoded on a creature you
//	 control. Whenever that creature deals combat damage to a player, its
//	 controller may cast a copy of the encoded card without paying its
//	 mana cost.)"
//
// The shared tapped reanimation (returnFirstLegalGraveyardTargetToBattlefieldTapped,
// the CR 614 entry stamped tapped), and cipher (game/cipher.go, CR
// 702.99, #3867): the creature it returns is one cipher may then encode
// the card on. A copy cast later reads "your graveyard" as its caster's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47735ef2-fc4c-4634-83ab-715b1e86da0c",
		Name:            "Writ of Return",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordCipher},
		Targets:         effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefieldTapped(ctx.Game, item)
		},
	})
}
