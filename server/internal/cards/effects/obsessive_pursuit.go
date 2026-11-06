package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Obsessive Pursuit — Enchantment {1}{B}:
//
//	"When this enchantment enters and at the beginning of your upkeep,
//	 you lose 1 life and create a Clue token.
//	 Whenever you attack, put X +1/+1 counters on target attacking
//	 creature, where X is the number of permanents you've sacrificed this
//	 turn. If X is three or more, that creature gains lifelink until end
//	 of turn."
//
// The first ability is one trigger with two conditions (Grell
// Philosopher's thisEntersOrYourUpkeep) and Azula, On the Hunt's body:
// lose 1 life, then a Clue (creating it is not investigating, the
// ruling).
//
// The second is Araña's "whenever you attack" (once per declaration,
// targeting an attacking creature). X is read as it resolves (CR
// 608.2h) from the turn's sacrifice tally, PermanentsSacrificed, which
// counts every permanent you sacrificed this turn, tokens included. X
// of zero puts no counters and grants nothing.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:     "7e7abb44-47b0-4d80-b8a1-36f0c1032aa3",
		Name:         "Obsessive Pursuit",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventBeginUpkeep}, thisEntersOrYourUpkeep,
				"Obsessive Pursuit — you lose 1 life and create a Clue token", Do(
					GainLife{Amount: -1},
					CreateToken{Template: ClueToken(), N: 1},
				)),
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Obsessive Pursuit — put X +1/+1 counters on target attacking creature, where X is the number of permanents you've sacrificed this turn",
					obsessivePursuitCounters),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

func obsessivePursuitCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target := FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	x := g.TurnTallyFor(ctx.Controller()).PermanentsSacrificed
	if x <= 0 {
		return nil
	}
	if err := (AddCounter{Target: target, Kind: game.CounterPlusOne, N: x}).Apply(ctx); err != nil {
		return err
	}
	if x < 3 {
		return nil
	}
	return GrantKeywordUntilEOT{
		Target:   target,
		Keywords: []string{"lifelink"},
		Label:    "Obsessive Pursuit — lifelink until end of turn",
	}.Apply(ctx)
}
