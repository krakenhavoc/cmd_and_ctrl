package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f91c860e-3feb-4a5a-b7ef-4e7d38d61c75",
		Name:         "Rewrite History",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventTapCard, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b32CreatureYouControlBecameTapped(ev, source, g)
				},
				"Rewrite History — draw a card, then discard a card and put a plan counter on this enchantment",
				mshB7RewriteHistoryLoot)),
			On(game.EventCounterPlaced, mshB2PlanCounterReached(4),
				"Rewrite History — sacrifice it; when you do, return up to two target instant and/or sorcery cards from your graveyard to your hand",
				mshB2SacrificeThisThenWhenYouDo("Rewrite History — return up to two target instant and/or sorcery cards from your graveyard to your hand", mshB7RewriteHistoryReturnBody)),
		},
	})
}

func mshB7RewriteHistoryLoot(g *game.Game, item *game.StackItem) error {
	if err := b16DrawThenDiscard(g, item, 1); err != nil {
		return err
	}
	return AddCounter{Target: item.SourceCardID, Kind: "plan", N: 1}.Apply(NewContext(g, item))
}

var mshB7RewriteHistoryReturnBody = game.ReflexiveBody("rewrite-history/return",
	simpleBody(func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("up to two target instant and/or sorcery cards from your graveyard",
			YouOwn(), Or(Instant(), Sorcery())).WithCount(0, 2)
	}))
