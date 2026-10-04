package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulamog, the Infinite Gyre — Legendary Creature — Eldrazi {11}, 10/10:
//
//	"When you cast this spell, destroy target permanent.
//	 Indestructible
//	 Annihilator 4 (Whenever this creature attacks, defending player
//	 sacrifices four permanents of their choice.)
//	 When Ulamog is put into a graveyard from anywhere, its owner
//	 shuffles their graveyard into their library."
//
// The cast trigger targets as Ulamog is cast and resolves above the
// spell, so it destroys its target even if Ulamog is countered (the
// 2018-12-07 ruling); a target that is gone or no longer legal by then
// is left alone (CR 608.2b). Indestructible and annihilator 4 are
// canonical keyword tokens, and annihilator's attack trigger is the
// engine's (CR 702.86, ADR 0113 §2). The shuffle watches from the
// graveyard, so it fires however Ulamog got there — indestructible
// stops destruction, not a sacrifice, the legend rule or a discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b817bc56-9b4d-4c50-bafa-3c652b99578f",
		Name:            "Ulamog, the Infinite Gyre",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible", "annihilator 4"},
		Triggered: []game.TriggeredAbility{
			ulamogInfiniteGyreCastTrigger(),
			WhenThisIsPutIntoAGraveyardFromAnywhere(
				"Ulamog, the Infinite Gyre — its owner shuffles their graveyard into their library",
				ShuffleYourGraveyardIntoYourLibrary),
		},
	})
}

// ulamogInfiniteGyreCastTrigger is "When you cast this spell, destroy
// target permanent."
func ulamogInfiniteGyreCastTrigger() game.TriggeredAbility {
	t := WhenYouCastThisSpell("Ulamog, the Infinite Gyre — destroy target permanent", destroyFirstLegalCardTarget)
	t.Targets = TargetPermanent("target permanent")
	return t
}
