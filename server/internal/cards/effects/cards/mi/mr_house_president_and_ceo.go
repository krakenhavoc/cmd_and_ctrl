package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mr. House, President and CEO — Legendary Artifact Creature — Human
// {R}{W}{B}, 0/4 (Fallout, #5948):
//
//	"Whenever you roll a 4 or higher, create a 3/3 colorless Robot artifact
//	 creature token. If you rolled 6 or higher, instead create that token
//	 and a Treasure token.
//	 {4}, {T}: Roll a six-sided die plus an additional six-sided die for
//	 each mana from Treasures spent to activate this ability."
//
// The trigger is once per roll instruction (the last die of the batch), and
// it fires only when the roll's result, the total of its dice (CR 706.2),
// is 4 or higher, so a lower roll puts nothing on the stack. The total is
// read again as it resolves from the dice of that instruction
// (RolledDice), 6 or higher adding the Treasure to the Robot. The activated
// ability reads the payment of its activation as it resolves
// (ReadsOwnActivations, ReadsSourceKinds, Jetmir's Fixer's reading): one
// die, and one more for each mana that came from a Treasure. A payment the
// engine did not see counts as no Treasure mana: weaker, never stronger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "d4e54b48-956f-4467-8485-a5fdb44779fd",
		Name:           "Mr. House, President and CEO",
		ReadsManaSpent: game.ReadsOwnActivations | game.ReadsSourceKinds,
		Completeness:   effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventRollDie},
			Key:     "Mr. House — create a 3/3 colorless Robot artifact creature token (and a Treasure token on a 6 or higher)",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventRollDie && ev.Actor == source.Controller && ev.Seq == ev.BatchSeq &&
					effects.GapB62RollTotal(g, ev.BatchSeq) >= 4
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				total := 0
				for _, rolled := range effects.RolledDice(g, item) {
					total += rolled.RollResult()
				}
				if err := (effects.CreateToken{Template: effects.TokenCard("3/3 colorless Robot artifact"), N: 1}).Apply(ctx); err != nil {
					return err
				}
				if total >= 6 {
					return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
				}
				return nil
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Roll a six-sided die plus an additional six-sided die for each mana from Treasures spent to activate this ability.",
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return effects.RollDice{Sides: 6, N: 1 + ctx.ManaSpent().CountFrom(game.ManaSourceTreasure)}.Apply(ctx)
			},
		}},
	})
}
