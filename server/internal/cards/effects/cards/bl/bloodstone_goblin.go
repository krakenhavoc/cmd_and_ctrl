package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodstone Goblin — Creature — Goblin Warrior {1}{R}, 2/2 (Dominaria, S82
// OLD, #5530):
//
//	"Whenever you cast a spell, if that spell was kicked, this creature gets
//	 +1/+1 and gains menace until end of turn."
//
// Risen Riptide's "you cast a kicked spell" check (CR 702.33a), read off the
// optional costs the cast paid, tested as the trigger is checked and again
// as it resolves (CR 603.4). The pump and the keyword go to this permanent
// only, and nothing happens to a new object (CR 400.7).
//
// No simplification.
func init() {
	const label = "Bloodstone Goblin — +1/+1 and menace until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "ec17b9db-dd80-4b46-a1b5-215f26bf6498",
		Name:         "Bloodstone Goblin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.S82OldB089YouCastAKickedSpell, label,
				func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 1, Toughness: 1, Label: label}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"menace"}, Label: label}.Apply(ctx)
				}),
		},
	})
}
