package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bruna, the Fading Light — Legendary Creature — Angel Horror {5}{W}{W},
// 5/7 (Innistrad Remastered):
//
//	"When you cast this spell, you may return target Angel or Human creature
//	 card from your graveyard to the battlefield.
//	 Flying, vigilance
//	 (Melds with Gisela, the Broken Blade.)"
//
// The cast trigger is Artisan of Kozilek's: it goes on the stack above the
// spell, picks its target then, and returns it as it resolves. The meld
// itself is printed on Gisela.
func init() {
	cast := effects.WhenYouCastThisSpell("Bruna, the Fading Light — return target Angel or Human creature card from your graveyard to the battlefield",
		effects.ReturnFirstLegalGraveyardTargetToBattlefield)
	cast.Targets = effects.TargetCardInGraveyard("target Angel or Human creature card from your graveyard",
		effects.Creature(), effects.Or(effects.HasSubtype("Angel"), effects.HasSubtype("Human")), effects.YouOwn())
	effects.Register(effects.Spec{
		OracleID:        "bce7e45d-b2c5-43aa-91a6-3a3c24f30850",
		Name:            "Bruna, the Fading Light",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(cast, "Return a target Angel or Human creature card from your graveyard to the battlefield?"),
		},
	})
}
