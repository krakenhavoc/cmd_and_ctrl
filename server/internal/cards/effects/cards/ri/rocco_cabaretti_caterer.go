package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rocco, Cabaretti Caterer — Legendary Creature — Elf Druid {X}{R}{G}{W},
// 3/1 (Streets of New Capenna):
//
//	"When Rocco enters, if you cast it, you may search your library for a
//	 creature card with mana value X or less, put it onto the battlefield,
//	 then shuffle."
//
// X is the cast's X, stamped on the trigger by the engine (XFromCast, CR
// 107.3m, as Arboreal Alliance's is). "If you cast it" is Annie Flash's
// read of the entering permanent's cast provenance (CR 603.4): a Rocco put
// onto the battlefield another way never triggers, and X is 0 then anyway.
// The search is Chord of Calling's, offered as a "you may" (Optional).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "41c765a4-a9c9-4169-a9f1-74ededeb793c",
		Name:         "Rocco, Cabaretti Caterer",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.AllOf(effects.Self, effects.GeologicalAppraiserWasCast),
			Key:       "Rocco, Cabaretti Caterer — you may search for a creature card with mana value X or less and put it onto the battlefield",
			XFromCast: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.X()
				return effects.SearchLibrary{
					Player:    ctx.Controller(),
					Predicate: func(c game.Card) bool { return c.IsCreature() && c.ManaValue() <= x },
					Dest:      game.ZoneBattlefield,
					Limit:     1,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Rocco, Cabaretti Caterer — a creature card with mana value X or less",
				}.Apply(ctx)
			},
		}},
	})
}
