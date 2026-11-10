package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prime Minister's Cabinet Room — Plane — Earth (Doctor Who, #4870):
//
//	"At the beginning of combat on your turn, up to one target creature
//	 you control becomes a copy of target creature an opponent controls.
//	 Will of the council — Whenever chaos ensues, starting with you, each
//	 player votes for a creature you don't control. Exile each creature
//	 with the most votes or tied for most votes."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// first ability has two target clauses, the first "up to one"; the copy has no
// stated duration (CR 611.2a) and takes the copiable values of the other
// creature as the ability resolves (CR 707.2), not its counters or other
// effects, and it does not become a token (the rulings). With either target
// illegal as it resolves nothing happens (CR 608.2b). The chaos ability
// (CR 311.7) is Council's Judgment's object vote (CR 701.38b, ADR 0146) over the
// creatures the planar controller doesn't control as the vote begins, starting
// with the planar controller in turn order; nothing is targeted, and every
// creature tied for the most votes is exiled at once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b3acbacd-7e97-4bcf-9281-fffa5ea3e70b",
		Name:         "Prime Minister's Cabinet Room",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Prime Minister's Cabinet Room — up to one target creature you control becomes a copy of target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					mine, ok1 := ctx.ClauseTarget(0)
					theirs, ok2 := ctx.ClauseTarget(1)
					if !ok1 || !ok2 || mine.Kind != game.TargetCard || theirs.Kind != game.TargetCard {
						return nil
					}
					return BecomeCopy{Targets: []uuid.UUID{mine.ID}, Of: theirs.ID, Duration: CopyIndefinite,
						Label: "Prime Minister's Cabinet Room — becomes a copy"}.Apply(ctx)
				}),
				Clauses(TargetCreature("up to one target creature you control", YouControl()).WithCount(0, 1),
					TargetCreature("target creature an opponent controls", OpponentControls()))),
			WheneverChaosEnsues("Prime Minister's Cabinet Room — Will of the council: each player votes for a creature you don't control; exile each creature with the most votes or tied",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.Battlefield.Cards {
						if c.Controller != ctx.Controller() && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					return Vote{
						Question: "Prime Minister's Cabinet Room — vote for a creature its controller doesn't control",
						Options:  VoteForPermanents(g, ids),
						Then:     primeMinistersCabinetRoomVoted,
					}.Apply(ctx)
				}),
		},
	})
}

// primeMinistersCabinetRoomVoted exiles every creature with the most votes or
// tied for most. An on-disk key: never renamed, never reused.
var primeMinistersCabinetRoomVoted = VoteResultThen("vote/prime-ministers-cabinet-room", func(ctx *Context, r game.VoteResult) error {
	var exile []uuid.UUID
	for _, i := range r.MostVotes() {
		if id := VotedPermanent(r, i); id != uuid.Nil && onBattlefield(ctx.Game, id) {
			exile = append(exile, id)
		}
	}
	if len(exile) > 0 {
		ctx.Game.ExileCardsForEffect(exile)
	}
	return nil
})
