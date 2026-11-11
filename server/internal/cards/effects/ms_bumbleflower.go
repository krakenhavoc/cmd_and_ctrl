package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ms. Bumbleflower — Legendary Creature — Rabbit Citizen {1}{G}{W}{U},
// 1/5 (Bloomburrow Commander, #5368):
//
//	"Vigilance
//	 Whenever you cast a spell, target opponent draws a card. Put a +1/+1
//	 counter on target creature. It gains flying until end of turn. If
//	 this is the second time this ability has resolved this turn, you
//	 draw two cards."
//
// Two targets, one trigger. Each half works if only its own target is still
// legal. "The second time this ability has resolved this turn" is the
// per-object resolution tally (Game.ResolvedThisTurn, as Zimone reads it),
// which counts the resolution in progress; a third resolution draws
// nothing.
func init() {
	Register(Spec{
		OracleID:        "25636214-c6ea-4655-a5e8-98e735da5b05",
		Name:            "Ms. Bumbleflower",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverYouCast(nil, msBumbleflowerLabel, msBumbleflowerResolve),
				Clauses(TargetPlayer("target opponent", Opponent()), TargetCreature("target creature"))),
		},
	})
}

const msBumbleflowerLabel = "Ms. Bumbleflower — target opponent draws a card; a +1/+1 counter on target creature, which gains flying until end of turn; the second time this turn, you draw two cards"

func msBumbleflowerResolve(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	if opp, ok := ctx.ClauseTarget(0); ok {
		if err := (DrawCards{Player: opp.ID, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	if cr, ok := ctx.ClauseTarget(1); ok {
		if err := (AddCounter{Target: cr.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
		if err := (GrantKeywordUntilEOT{Target: cr.ID, Keywords: []string{"flying"}, Label: "Ms. Bumbleflower — flying"}).Apply(ctx); err != nil {
			return err
		}
	}
	if g.ResolvedThisTurn(item.SourceCardID, msBumbleflowerLabel) == 2 {
		return DrawCards{N: 2}.Apply(ctx)
	}
	return nil
}
