package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Wasp, Winsome Avenger — {1}{U} Legendary Creature — Human Hero 2/1
// (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 Flying
//	 When The Wasp enters, target Hero gains hexproof until end of turn.
//	 Whenever The Wasp attacks, tap target creature defending player
//	 controls."
//
// The enters target is any Hero on the battlefield, The Wasp included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4b706959-ea38-4808-870a-7c35d83f7ab2",
		Name:            "The Wasp, Winsome Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(
					effects.WhenThisEnters("The Wasp — target Hero gains hexproof until end of turn",
						effects.EbTargetCreatureGainsUntilEOT("hexproof", "The Wasp, Winsome Avenger — hexproof")),
					effects.TargetCreature("target Hero", effects.Subtype("Hero"))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof"))),
			effects.DistractingGeistTap("The Wasp — tap target creature defending player controls"),
		},
	})
}
