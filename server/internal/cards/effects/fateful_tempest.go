package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fateful Tempest — Sorcery {2}{R}:
//
//	"Council's dilemma — Starting with you, each player votes for past
//	 or present. You mill a card for each past vote, then Fateful
//	 Tempest deals damage to each opponent equal to the total mana
//	 value of cards milled this way. Exile the top card of your library
//	 for each present vote. Until the end of your next turn, you may
//	 play the exiled cards."
//
// A council's dilemma (ADR 0146, CR 701.38): every ballot pays the
// caster. The past votes are one mill of that many cards; the damage
// is read off the cards that really reached the graveyard (a card a
// replacement sent elsewhere adds nothing), and it is dealt as the mill
// settles, before the present votes exile anything. The present votes
// exile that many cards off the top, playable by the caster until the
// end of their next turn — lands included, as "play" says.
//
// A bot casting it votes past (damage); an opponent's bot votes
// present, the card advantage, since the mill is a damage roll.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d2f02f7-adcd-4d23-a051-ebca70f22ba5",
		Name:         "Fateful Tempest",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Vote{
				Question:      "Fateful Tempest — vote for past or present",
				Words:         []string{"past", "present"},
				ForController: []int{2, 1},
				ForOpponents:  []int{0, 1},
				Then:          fatefulTempestVoted,
			}.Apply(ctx)
		},
	})
}

var fatefulTempestVoted = VoteResultThen("vote/fateful-tempest", func(ctx *Context, r game.VoteResult) error {
	past, present := r.Votes(0), r.Votes(1)
	item := ctx.Item
	exile := func(g *game.Game) error {
		if present <= 0 {
			return nil
		}
		return ExileTopNUntilYourNextTurn(NewContext(g, item), present)
	}
	if past <= 0 {
		return exile(ctx.Game)
	}
	return ctx.Game.MillToZoneThenForEffect(ctx.Controller(), past, game.ZoneGraveyard, nil,
		func(g *game.Game, milled []uuid.UUID) error {
			total := 0
			for _, id := range milled {
				if c, ok := g.LookupCardForEffect(id); ok {
					total += c.ManaValue()
				}
			}
			if total > 0 {
				if err := (DamageEachOpponent{N: total}).Apply(NewContext(g, item)); err != nil {
					return err
				}
			}
			return exile(g)
		})
})
