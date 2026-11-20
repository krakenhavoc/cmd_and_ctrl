package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	cast := WhenYouCastThisSpell("Bruna, the Fading Light — return target Angel or Human creature card from your graveyard to the battlefield",
		returnFirstLegalGraveyardTargetToBattlefield)
	cast.Targets = TargetCardInGraveyard("target Angel or Human creature card from your graveyard",
		Creature(), Or(HasSubtype("Angel"), HasSubtype("Human")), YouOwn())
	Register(Spec{
		OracleID:        "bce7e45d-b2c5-43aa-91a6-3a3c24f30850",
		Name:            "Bruna, the Fading Light",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			Optional(cast, "Return a target Angel or Human creature card from your graveyard to the battlefield?"),
		},
	})
}
