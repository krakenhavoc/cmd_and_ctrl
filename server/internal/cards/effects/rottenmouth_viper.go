package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rottenmouth Viper — Creature — Elemental Snake {5}{B}, 6/6:
//
//	"As an additional cost to cast this spell, you may sacrifice any
//	 number of nonland permanents. This spell costs {1} less to cast for
//	 each permanent sacrificed this way.
//	 Whenever this creature enters or attacks, put a blight counter on
//	 it. Then for each blight counter on it, each opponent loses 4 life
//	 unless that player sacrifices a nonland permanent of their choice or
//	 discards a card."
//
// The variable sacrifice (ADR 0100 §3) with a {1} per-sacrifice discount
// read at CR 601.2f (CostsLessPerSacrificed), so the preview, the bot
// and CastSpell all charge the same price.
//
// The trigger puts the counter on first and counts after — through the
// counter path's continuation, so a replacement that changes the
// placement (a counter doubler) is settled before the count is read.
// The count is the Viper's blight counters as they stand then; if the
// Viper has left the battlefield, no counter goes on and the count is
// the counters it had as it left (CR 608.2h, last-known information).
// Each blight counter is one run of Torment of Hailfire's punisher, at
// 4 life (punisherRepeat): every opponent in turn order chooses to lose
// 4 life, sacrifice a nonland permanent, or discard a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "2c75623c-59f4-4449-ab43-9d1225185ad9",
		Name:           "Rottenmouth Viper",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeAnyNumberCost("any number of nonland permanents", Nonland()),
		SelfCostModifiers: []game.CostModifier{
			CostsLessPerSacrificed("{1}", "This spell costs {1} less to cast for each permanent sacrificed this way"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Rottenmouth Viper — put a blight counter on it, then each opponent loses 4 life for each unless they sacrifice or discard", rottenmouthViperTrigger),
		},
	})
}

// rottenmouthViper is the printed punisher: 4 life a refusal.
var rottenmouthViper = punisherRepeat{Name: "Rottenmouth Viper", Life: 4}

// rottenmouthBlight is the blight counter's kind.
const rottenmouthBlight = "blight"

// rottenmouthViperTrigger puts the counter on, then runs the punisher
// once per blight counter the Viper has. ctx.SourcePermanent is the
// CR 608.2h read: the Viper as it is now while the same object is on
// the battlefield, and as it last was there once it has left (Left).
func rottenmouthViperTrigger(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	perm, ok := ctx.SourcePermanent()
	if !ok {
		return nil
	}
	if perm.Left {
		return rottenmouthViper.start(ctx, perm.Counters[rottenmouthBlight])
	}
	return g.AddCounterByThenForEffect(item.Controller, item.SourceCardID, rottenmouthBlight, 1, func(g *game.Game, _ int) error {
		ctx := NewContext(g, item)
		perm, ok := ctx.SourcePermanent()
		if !ok {
			return nil
		}
		return rottenmouthViper.start(ctx, perm.Counters[rottenmouthBlight])
	})
}
