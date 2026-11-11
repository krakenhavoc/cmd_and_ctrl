package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prisoner's Dilemma — Sorcery {3}{R}{R} (Murders at Karlov Manor
// Commander):
//
//	"Each opponent secretly chooses silence or snitch, then the choices
//	 are revealed. If each opponent chose silence, Prisoner's Dilemma deals
//	 4 damage to each of them. If each opponent chose snitch, Prisoner's
//	 Dilemma deals 8 damage to each of them. Otherwise, Prisoner's Dilemma
//	 deals 12 damage to each opponent who chose silence.
//	 Flashback {5}{R}{R}"
//
// The secret-council machinery with the vote rules off (a choice, not a
// vote): one hidden prompt per opponent, nothing shown to anyone else
// until the last is in, then one reveal. Only opponents choose, and only
// opponents still in the game at resolution are asked.
func init() {
	Register(Spec{
		OracleID:         "fecbbc79-2f74-46b3-a049-45736f904f58",
		Name:             "Prisoner's Dilemma",
		Completeness:     CompletenessFull,
		Purpose:          game.Purpose{DamageEachOpponent: 4},
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{R}{R}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if len(ctx.Opponents()) == 0 {
				return nil
			}
			return Vote{
				Question:      "Prisoner's Dilemma — secretly choose silence or snitch",
				Words:         []string{"silence", "snitch"},
				Voters:        ctx.Opponents(),
				Choice:        true,
				Secret:        true,
				ForController: []int{1, 0},
				ForOpponents:  []int{0, 1},
				Then:          prisonersDilemmaChosen,
			}.Apply(ctx)
		},
	})
}

var prisonersDilemmaChosen = VoteResultThen("vote/prisoners-dilemma", func(ctx *Context, r game.VoteResult) error {
	silent, snitches := r.VotersFor(0), r.VotersFor(1)
	var victims []uuid.UUID
	amount := 0
	switch {
	case len(snitches) == 0:
		// Each opponent chose silence (vacuously so with no opponents).
		victims, amount = silent, 4
	case len(silent) == 0:
		victims, amount = snitches, 8
	default:
		victims, amount = silent, 12
	}
	for _, p := range victims {
		if err := ctx.Game.DealDamageToPlayerForEffect(ctx.Source(), p, amount); err != nil {
			return err
		}
	}
	return nil
})
