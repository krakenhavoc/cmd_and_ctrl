package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounty of the Luxa — Enchantment {2}{G}{U} (S82 GAP, #5948):
//
//	"At the beginning of your first main phase, remove all flood counters
//	 from this enchantment. If no counters were removed this way, put a flood
//	 counter on this enchantment and draw a card. Otherwise, add {C}{G}{U}."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6b20c30-88c2-4cb0-a5e4-46f1df924a7c",
		Name:         "Bounty of the Luxa",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain("Bounty of the Luxa — remove all flood counters; if none, a flood counter and draw a card, otherwise add {C}{G}{U}",
				effects.Do(effects.Instruction(bountyOfTheLuxaResolve))),
		},
	})
}

const bountyFlood = "flood"

func bountyOfTheLuxaResolve(ctx *effects.Context) error {
	g, self := ctx.Game, ctx.Source()
	if !effects.OnBattlefield(g, self) || effects.SourceIsNewObject(g, ctx.Item) {
		return nil
	}
	c, ok := g.LookupCardForEffect(self)
	if !ok {
		return nil
	}
	if n := c.Counters[bountyFlood]; n > 0 {
		if err := (effects.AddCounter{Target: self, Kind: bountyFlood, N: -n}).Apply(ctx); err != nil {
			return err
		}
		return effects.AddMana{Produced: "{C}{G}{U}"}.Apply(ctx)
	}
	if err := (effects.AddCounter{Target: self, Kind: bountyFlood, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.DrawCards{N: 1}.Apply(ctx)
}
