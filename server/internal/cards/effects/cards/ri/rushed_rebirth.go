package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rushed Rebirth — Instant {B}{G} (Strixhaven: School of Mages, #5502):
//
//	"Choose target creature. When that creature dies this turn, search your
//	 library for a creature card with lesser mana value, put it onto the
//	 battlefield tapped, then shuffle."
//
// Fatal Fissure's shape: the spell creates a delayed triggered ability about
// the chosen creature as the object it is now (CR 603.7a, 603.7c), controlled
// by the caster (CR 603.7d), whoever controls the creature. "Lesser" is read
// against the creature's mana value as it last existed on the battlefield (CR
// 608.2h, CR 603.10a), so a creature that died as a copy or with a changed
// mana value is compared as it was. The search is mandatory in the sense that
// the trigger is not optional, but a player may fail to find (CR 701.23b);
// the shuffle follows either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8ee64f9d-54c4-4a75-a8ac-6bb0a060b960",
		Name:         "Rushed Rebirth",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerProtect},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			creature, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, creature,
				"Rushed Rebirth — search your library for a creature card with lesser mana value, put it onto the battlefield tapped, then shuffle",
				rushedRebirthDiesBody)
		},
	})
}

// rushedRebirthDiesBody is the fired trigger's body: the search, read against
// the dead creature's last-known mana value. An on-disk key.
var rushedRebirthDiesBody = game.DelayedBody("old-b042/rushed-rebirth-search",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		obj, ok := effects.PayloadObject(ctx)
		if !ok {
			return nil
		}
		info, ok := g.PermanentForEffect(obj)
		if !ok {
			return nil
		}
		mv := info.ManaValue
		return effects.SearchLibrary{
			Player: ctx.Controller(),
			Predicate: func(c game.Card) bool {
				return c.IsCreature() && c.ManaValue() < mv
			},
			Dest: game.ZoneBattlefield, Limit: 1, Shuffle: true, TappedOnEntry: true,
			Reason: "Rushed Rebirth — a creature card with lesser mana value, onto the battlefield tapped",
		}.Apply(ctx)
	})
