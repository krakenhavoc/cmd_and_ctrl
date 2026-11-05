package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Assimilate the Wounded — Instant {2}{B} (Star Trek Commander, #4974):
//
//	"For each opponent, assimilate up to one target creature card from
//	 that player's graveyard that was put there from the battlefield this
//	 turn. (Put it onto the battlefield under your control with a +1/+1
//	 counter. It's a Borg artifact creature and loses all other creature
//	 types.)"
//
// One "up to one" clause per possible opponent, each bound to the
// graveyard of the n-th live opponent after the caster in turn order
// (trekB36OwnedByNthOpponent, opponent_clauses.go's seat numbering), so
// two picks can never come from one graveyard and an absent seat offers
// nothing. "Put there from the battlefield this turn" is Samwise's
// reading (putThereFromTheBattlefieldThisTurn): the card's newest arrival
// in the graveyard happened this turn and came from the battlefield, so a
// milled or discarded card, or one that died last turn, is not offered.
// Every clause is optional, so the spell is castable with no target.
//
// Each pick is re-judged as the spell resolves (CR 608.2b): one that left
// its graveyard in response is skipped and the rest are still taken. The
// legal picks are assimilated together, as one entry (CR 603.6a), each
// with a +1/+1 counter and as a Borg artifact creature that keeps its
// other card types (CR 205.1b; trekB11Assimilate's entry).
//
// No simplification.
func init() {
	var clauses []*game.TargetSpec
	for n := 1; n <= game.MaxPlayers-1; n++ {
		clauses = append(clauses, TargetCardInGraveyard(
			"up to one target creature card in the graveyard of the "+nthOpponentWords[n]+" opponent after you in turn order that was put there from the battlefield this turn",
			trekB36OwnedByNthOpponent(n), Creature(), putThereFromTheBattlefieldThisTurn).WithCount(0, 1))
	}
	Register(Spec{
		OracleID:     "2147101f-d24a-45bb-a141-b69ef7841f53",
		Name:         "Assimilate the Wounded",
		Completeness: CompletenessFull,
		Targets:      Clauses(clauses[0], clauses[1:]...),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return trekB36AssimilateTogether(ctx, game.ZoneGraveyard, ids, "Assimilate the Wounded — assimilated: a Borg artifact creature")
		},
	})
}
