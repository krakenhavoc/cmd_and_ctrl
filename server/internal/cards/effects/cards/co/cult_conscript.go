package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cult Conscript — Creature — Skeleton Warrior {B}, 2/1 (Dominaria United,
// S82 DMU batch 4):
//
//	"This creature enters tapped.
//	 {1}{B}: Return this card from your graveyard to the battlefield.
//	 Activate only if a non-Skeleton creature died under your control this
//	 turn."
//
// An enters-tapped self replacement and Cauldron Familiar's graveyard return
// gated by an activation condition (CR 602.5b) read off the turn's per-player
// death tally, less the Skeletons among them. The return puts it onto the
// battlefield, where the replacement taps it as it enters.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5762815-5fc3-4ae7-abee-69f27c5e4e90",
		Name:         "Cult Conscript",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{B}: Return this card from your graveyard to the battlefield. Activate only if a non-Skeleton creature died under your control this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:      effects.ManaCost("{1}{B}"),
			Zones:     []game.ZoneKind{game.ZoneGraveyard},
			Condition: effects.DmuB04NonSkeletonDiedUnderYourControl,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ReturnThisFromGraveyardToBattlefield(g, item)
			},
		}},
	})
}
