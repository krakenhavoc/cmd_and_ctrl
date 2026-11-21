package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ral and the Implicit Maze — Enchantment — Saga {3}{R}{R} (Modern Horizons
// 3):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I — This Saga deals 2 damage to each creature and planeswalker your
//	 opponents control.
//	 II — You may discard a card. If you do, exile the top two cards of your
//	 library. You may play them until the end of your next turn.
//	 III — Create a Spellgorger Weird token. (It's a {2}{R} 2/2 Weird
//	 creature with "Whenever you cast a noncreature spell, put a +1/+1
//	 counter on Spellgorger Weird.")"
//
// I is Fulminous Forte's sweep over opponents' creatures and planeswalkers
// (it names no target). II is a discard through the run's own prompt, and
// only once a card has really left the hand are the top two cards exiled with
// a play window to the end of your next turn (Alania's Pathmaker's window).
// III makes the printed token, whose trigger is the Weird card's own and
// whose mana cost makes its mana value 3.
//
// No simplification.
func init() {
	creatureOrWalker := effects.Or(effects.Creature(), effects.Planeswalker())
	sweep := effects.ChapterTrigger(1, "Ral and the Implicit Maze — 2 damage to each creature and planeswalker your opponents control",
		func(g *game.Game, item *game.StackItem) error {
			return effects.DamageEachMatching(effects.NewContext(g, item), effects.And(creatureOrWalker, effects.OpponentControls()), 2)
		})
	sweep.Purpose = game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true, Partial: true}}
	rummage := effects.ChapterTrigger(2, "Ral and the Implicit Maze — you may discard a card; if you do, exile the top two cards of your library and play them until the end of your next turn",
		func(g *game.Game, item *game.StackItem) error {
			return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   item.Controller,
				Source:   item.SourceCardID,
				N:        1,
				UpTo:     true,
				Question: "Ral and the Implicit Maze — you may discard a card to exile the top two cards of your library",
				Asked:    game.NewAskedItem(item),
			}, effects.GapB56RalExileTwo, nil)
		})
	weird := effects.ChapterTrigger(3, "Ral and the Implicit Maze — create a Spellgorger Weird token",
		effects.Do(effects.CreateToken{Template: effects.SpellgorgerWeirdToken(), N: 1}))
	effects.Register(effects.Spec{
		OracleID:     "3affd37a-d4f6-488a-b30f-8e824b640687",
		Name:         "Ral and the Implicit Maze",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{sweep, rummage, weird},
	})
}
