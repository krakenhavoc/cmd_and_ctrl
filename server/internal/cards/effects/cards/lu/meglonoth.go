package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meglonoth — Creature — Beast {3}{R}{G}{W}, 6/6 (Conflux):
//
//	"Vigilance, trample
//	 Whenever this creature blocks a creature, this creature deals damage to
//	 that creature's controller equal to this creature's power."
//
// One trigger for each attacker blocked (CR 509.3b, Brimaz's shape). The
// attacker is stamped as the object it is when the block is declared
// (CR 400.7); as the trigger resolves the damage goes to that creature's
// controller, as it last was if the attacker has left (CR 608.2h), and the
// amount is Meglonoth's power then (its last-known power if it has left too).
//
// No simplification.
func init() {
	const label = "Meglonoth — it deals damage to that creature's controller equal to its power"
	t := effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return effects.B28SelfBlocked(ev, source)
	}, label, func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		attacker, ok := g.PermanentForEffect(item.Params.Object)
		if !ok {
			return nil
		}
		self, ok := ctx.SourcePermanent()
		if !ok || self.Power <= 0 || attacker.Controller == uuid.Nil {
			return nil
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: attacker.Controller, Amount: self.Power}.Apply(ctx)
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, label)
		item.Params.Object = game.ObjectRef{ID: ev.Target, Epoch: g.ObjectEpochForEffect(ev.Target)}
		return item
	}
	effects.Register(effects.Spec{
		OracleID:        "92e287b1-7cb1-4536-8942-d1ad059bf9bc",
		Name:            "Meglonoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
		Triggered:       []game.TriggeredAbility{t},
	})
}
