package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artisan of Kozilek — Creature — Eldrazi {9}, 10/9:
//
//	"When you cast this spell, you may return target creature card from
//	 your graveyard to the battlefield.
//	 Annihilator 2"
//
// The cast trigger is a "you may": the caster is asked as it triggers,
// and on a yes picks a creature card in their own graveyard. It resolves
// above the spell, so it returns the creature even if the Artisan is
// countered; a card that left the graveyard in response is not returned
// (CR 608.2b). Annihilator 2 is the canonical keyword token, and the
// engine makes its attack trigger (ADR 0113 §2).
//
// No simplification.
func init() {
	cast := effects.WhenYouCastThisSpell("Artisan of Kozilek — return target creature card from your graveyard to the battlefield",
		effects.ReturnFirstLegalGraveyardTargetToBattlefield)
	cast.Targets = effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:        "19409704-09c4-4a4b-a5a7-f95120b425db",
		Name:            "Artisan of Kozilek",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"annihilator 2"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(cast, "Return a creature card from your graveyard to the battlefield?"),
		},
	})
}
