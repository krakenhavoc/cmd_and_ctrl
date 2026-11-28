package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chaos Harlequin — Creature — Human {2}{R}{R} 2/4 (Alliances, S82 OLD batch
// 820):
//
//	"{R}: Exile the top card of your library. If that card is a land card,
//	 this creature gets -4/-0 until end of turn. Otherwise, this creature
//	 gets +2/+0 until end of turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c8e96a51-24c9-4955-8bf0-6ef448fa8011",
		Name:         "Chaos Harlequin",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Exile the top card of your library. If that card is a land card, this creature gets -4/-0 until end of turn. Otherwise, this creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 2}},
			Cost:    effects.ManaCost("{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				return g.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 1, ctx.ItemTail(chaosHarlequinPump))
			},
		}},
	})
}

var chaosHarlequinPump = effects.CardTail("chaos-harlequin/pump", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 || !effects.OnBattlefield(ctx.Game, ctx.Item.SourceCardID) || effects.SourceIsNewObject(ctx.Game, ctx.Item) {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok {
		return nil
	}
	power := 2
	if c.IsLand() {
		power = -4
	}
	return effects.BoostUntilEOT{Target: ctx.Item.SourceCardID, Power: power, Label: "Chaos Harlequin — power until end of turn"}.Apply(ctx)
})
