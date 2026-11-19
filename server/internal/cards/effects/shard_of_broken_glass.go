package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shard of Broken Glass — Artifact — Equipment {1}:
//
//	"Equipped creature gets +1/+0.
//	 Whenever equipped creature attacks, you may mill two cards.
//	 Equip {1}"
//
// Bonesplitter's pump and equip with an attack trigger whose optional mill
// is asked as the trigger resolves and only when the library holds two
// cards (a player can't choose to mill more cards than their library has),
// as Ashcoat of the Shadow Swarm's end-step mill is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6c2f370-fc33-4907-975c-a07af3dc664d",
		Name:         "Shard of Broken Glass",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attachedCreatureAttacked(ev, source)
			},
			Key: "Shard of Broken Glass — you may mill two cards",
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if p := g.PlayerByIDForEffect(ctx.Controller()); p == nil || p.Library == nil || p.Library.Size() < 2 {
					return nil
				}
				return MayChoice{
					Question: "Shard of Broken Glass — mill two cards?",
					Then:     shardOfBrokenGlassMill,
				}.Apply(ctx)
			},
		}},
		Activated: []ActivatedAbility{EquipAbility("{1}")},
	})
}

// shardOfBrokenGlassMill is the answer to the mill question. An on-disk key.
var shardOfBrokenGlassMill = MayAnswer("confirm/shard-of-broken-glass-mill-two", func(ctx *Context, r game.ConfirmAnswer) error {
	if !r.Accepted {
		return nil
	}
	return MillCards{N: 2}.Apply(ctx)
})
