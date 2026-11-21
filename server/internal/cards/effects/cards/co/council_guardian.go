package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Council Guardian — Creature — Giant Soldier {5}{W}, 5/5:
//
//	"Will of the council — When this creature enters, starting with
//	 you, each player votes for blue, black, red, or green. This
//	 creature gains protection from each color with the most votes or
//	 tied for most votes."
//
// An enters trigger whose resolution is ADR 0146's vote (CR 701.38),
// over four colour words. Each colour with the most votes, or tied for
// the most, becomes a protection quality on the Guardian (CR 702.16),
// with no stated duration, so it lasts for as long as the Guardian
// stays on the battlefield (CR 611.2a). "This creature" is the object
// that entered (CR 400.7): one that left while the trigger waited gets
// nothing.
//
// The bot hints are a guess: the Guardian's controller favours red,
// then black; an opponent's bot favours blue, then black. Nothing
// about the table is read.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f212675-e34d-4387-a125-e17b25757083",
		Name:         "Council Guardian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Council Guardian — each player votes for blue, black, red, or green", effects.VoteForThisCreature(effects.Vote{
				Question:      "Council Guardian — vote for blue, black, red, or green",
				Words:         []string{"blue", "black", "red", "green"},
				ForController: []int{0, 1, 2, 0},
				ForOpponents:  []int{2, 1, 0, 0},
				Then:          councilGuardianVoted,
			})),
		},
	})
}

// councilGuardianColors maps each option, in printed order, to the
// engine's wire colour letter.
var councilGuardianColors = []string{"U", "B", "R", "G"}

var councilGuardianVoted = effects.VoteResultThen("vote/council-guardian", func(ctx *effects.Context, r game.VoteResult) error {
	self := effects.VotedCreature(r)
	if self == uuid.Nil {
		return nil
	}
	var tokens []string
	for _, i := range r.MostVotes() {
		if t := game.ProtectionFromColor(councilGuardianColors[i]); t != "" {
			tokens = append(tokens, t)
		}
	}
	if len(tokens) == 0 {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   self,
		Mods:     []game.Mod{game.AddKeywordsMod(tokens...)},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), self),
		Label:    "Council Guardian — protection from the colors with the most votes",
	}.Apply(ctx)
})
