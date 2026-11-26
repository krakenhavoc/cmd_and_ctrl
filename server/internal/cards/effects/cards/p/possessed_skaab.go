package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Possessed Skaab — Creature — Zombie {3}{U}{B}, 3/2 (Magic Origins, S82 older
// sets):
//
//	"When this creature enters, return target instant, sorcery, or creature
//	 card from your graveyard to your hand.
//	 If this creature would die, exile it instead."
//
// The enters trigger is Vampire Charmseeker's graveyard-to-hand over your own
// graveyard, the target chosen as the trigger goes on the stack and re-checked
// as it resolves (CR 608.2b). The replacement is Gloomshrieker's: a move of
// this creature from the battlefield to a graveyard becomes a move to exile
// (CR 614.1a, 700.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a42b567d-6bfc-49e7-8f91-917e2bb3046c",
		Name:         "Possessed Skaab",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.OldB321ThisWouldDieExile("Possessed Skaab")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Possessed Skaab — return target instant, sorcery, or creature card from your graveyard to your hand",
					effects.ReturnTargetGraveyardCardToHand),
				effects.TargetCardInGraveyard("target instant, sorcery, or creature card from your graveyard",
					effects.Or(effects.Instant(), effects.Sorcery(), effects.Creature()), effects.YouOwn())),
		},
	})
}
