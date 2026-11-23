package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rewrite History — Enchantment — Plan {2}{U} (Marvel Super Heroes, #4455):
//
//	"Whenever one or more creatures you control become tapped, draw a card,
//	 then discard a card and put a plan counter on this enchantment.
//	 When the fourth plan counter is put on this enchantment, sacrifice it.
//	 When you do, return up to two target instant and/or sorcery cards from
//	 your graveyard to your hand."
//
// Deeproot Pilgrimage's once-per-batch "become tapped" watch (a tap or an
// attack that taps a non-vigilant attacker) plus Claim the Kingdom's plan
// counter and sacrifice-then-reflexive recipe. The loot is a draw whose
// discard opens after the draw lands (CR 608.2c); the counter goes on at the
// same resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f91c860e-3feb-4a5a-b7ef-4e7d38d61c75",
		Name:         "Rewrite History",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return effects.B32CreatureYouControlBecameTapped(ev, source, g)
				},
				"Rewrite History — draw a card, then discard a card and put a plan counter on this enchantment",
				mshB7RewriteHistoryLoot)),
			effects.On(game.EventCounterPlaced, effects.MshB2PlanCounterReached(4),
				"Rewrite History — sacrifice it; when you do, return up to two target instant and/or sorcery cards from your graveyard to your hand",
				effects.MshB2SacrificeThisThenWhenYouDo("Rewrite History — return up to two target instant and/or sorcery cards from your graveyard to your hand", mshB7RewriteHistoryReturnBody)),
		},
	})
}

func mshB7RewriteHistoryLoot(g *game.Game, item *game.StackItem) error {
	if err := effects.B16DrawThenDiscard(g, item, 1); err != nil {
		return err
	}
	return effects.AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(effects.NewContext(g, item))
}

var mshB7RewriteHistoryReturnBody = game.ReflexiveBody("rewrite-history/return",
	effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := effects.NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	effects.ConstTargets(func() *game.TargetSpec {
		return effects.TargetCardInGraveyard("up to two target instant and/or sorcery cards from your graveyard",
			effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())).WithCount(0, 2)
	}))
