package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ashling, Flame Dancer — Legendary Creature — Elemental Shaman {2}{R}{R}, 4/4
// (Modern Horizons 3):
//
//	"You don't lose unspent red mana as steps and phases end.
//	 Magecraft — Whenever you cast or copy an instant or sorcery spell, discard
//	 a card, then draw a card. If this is the second time this ability has
//	 resolved this turn, Ashling deals 2 damage to each opponent and each
//	 creature they control. If it's the third time, add {R}{R}{R}{R}."
//
// The draw happens whether or not a card was discarded (an empty hand discards
// nothing, CR 701.9a); the count of resolutions is read once the discard is done.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fe9fd59-9cac-4572-8f94-bc19bea44f94",
		Name:         "Ashling, Flame Dancer",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, OpponentsOnly: true}},
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolKeep, Colors: []string{"R"}},
		},
		Triggered: []game.TriggeredAbility{
			effects.Magecraft(gapB52AshlingLabel, func(g *game.Game, item *game.StackItem) error {
				return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player: item.Controller,
					Source: item.SourceCardID,
					N:      1,
					Asked:  game.NewAskedItem(item),
				}, gapB52AshlingThen, nil)
			}),
		},
	})
}

const gapB52AshlingLabel = "Ashling, Flame Dancer — discard a card, then draw a card; the second time this turn 2 damage to each opponent and their creatures, the third time add {R}{R}{R}{R}"

// gapB52AshlingThen is the rest of the ability once the discard is done. An
// on-disk key: never renamed, never reused.
var gapB52AshlingThen = game.RegisterDiscardThen("choose-cards/ashling-flame-dancer-then", func(g *game.Game, r game.DiscardsDone) error {
	if err := g.DrawNForEffect(r.Player, 1); err != nil {
		return err
	}
	ctx := effects.NewContext(g, r.Item)
	switch g.ResolvedThisTurn(r.Source, gapB52AshlingLabel) {
	case 2:
		return effects.DamageEachMatchingAndEachOpponent(ctx, effects.And(effects.Creature(), effects.OpponentControls()), 2)
	case 3:
		return effects.AddMana{Player: r.Player, Produced: "{R}{R}{R}{R}"}.Apply(ctx)
	}
	return nil
})
