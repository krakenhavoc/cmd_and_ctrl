package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Korvold and the Noble Thief — Enchantment — Saga {3}{R} (Wilds of
// Eldraine, #4471):
//
//	"(As this Saga enters and after your draw step, add a lore counter.
//	 Sacrifice after III.)
//	 I, II — Create a Treasure token. (It's an artifact with "{T},
//	 Sacrifice this token: Add one mana of any color.")
//	 III — Exile the top three cards of target opponent's library. You may
//	 play those cards this turn."
//
// Three chapter abilities (CR 714.2). I and II share one body, the
// catalog's Treasure (CR 111.10a). III is Inside Information's impulse
// exile of another player's library at a fixed three, with no price
// attached: the cards are the opponent's, the Saga's controller may play
// them this turn paying their costs as usual, and a land among them is
// played rather than cast. With no opponent to target, chapter III is
// removed (CR 603.3d) and the Saga is still sacrificed after it.
//
// No simplification.
func init() {
	treasure := func(g *game.Game, item *game.StackItem) error {
		return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(effects.NewContext(g, item))
	}
	effects.Register(effects.Spec{
		OracleID:     "bcae7f20-b601-4669-b3a7-c32ed0c6b9d3",
		Name:         "Korvold and the Noble Thief",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.ChapterTrigger(1, "Korvold and the Noble Thief — I: create a Treasure token", treasure),
			effects.ChapterTrigger(2, "Korvold and the Noble Thief — II: create a Treasure token", treasure),
			effects.ChapterTriggerTargeting(3, "Korvold and the Noble Thief — III: exile the top three cards of target opponent's library; you may play them this turn",
				effects.TargetPlayer("target opponent", effects.Opponent()),
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetPlayer {
							continue
						}
						return effects.ExileTopWithPermission{From: t.ID, GrantTo: item.Controller, N: 3}.Apply(ctx)
					}
					return nil
				}),
		},
	})
}
