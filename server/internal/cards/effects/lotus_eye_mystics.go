package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lotus-Eye Mystics — Creature — Human Monk {3}{W}, 3/2 (Fate Reforged,
// #5552):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, return target enchantment card from your
//	 graveyard to your hand."
//
// Prowess is printed. The entry trigger is Auramancer's without the "may":
// the target is chosen as it goes on the stack and re-checked as it resolves
// (CR 608.2b); with no enchantment card in your graveyard it is removed (CR
// 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "efb11270-4b5a-4c52-862f-1d24c6b95d73",
		Name:            "Lotus-Eye Mystics",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Lotus-Eye Mystics — return target enchantment card from your graveyard to your hand",
					returnFirstLegalGraveyardTargetToHand),
				TargetCardInGraveyard("target enchantment card from your graveyard", YouOwn(), Enchantment())),
		},
	})
}
