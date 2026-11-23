package cardsex

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fangorn, Tree Shepherd — Legendary Creature — Treefolk {4}{G}{G}{G}, 4/10
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"Treefolk you control have vigilance.
//	 Whenever one or more Treefolk you control attack, add twice that much
//	 {G}.
//	 You don't lose unspent green mana as steps and phases end."
//
// Three independent abilities. The grant is a layer-6 keyword over the
// Treefolk the controller controls, Fangorn included. The trigger is once
// per declaration of attackers that includes a Treefolk (OncePerBatch, CR
// 603.2c); "that much" is the number of Treefolk the controller has attacking
// as it resolves, and it adds twice as many {G}. The mana-pool clause is
// Omnath, Locus of Mana's ManaPoolKeep, covering green only (CR 106.4).
// The creature type is read through the source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "04166d20-88c9-4c89-be9d-3e422f90959a",
		Name:         "Fangorn, Tree Shepherd",
		Completeness: effects.CompletenessFull,
		ManaPool:     []game.ManaPoolStatic{{Kind: game.ManaPoolKeep, Colors: []string{"G"}}},
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller && target.HasSubtype(source.TextSubtype("Treefolk"))
		}, "vigilance")},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Treefolk"))
			}, "Fangorn, Tree Shepherd — add twice that much {G}",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					treefolk := ctx.TextSubtype("Treefolk")
					n := effects.CountControlled(g, item.Controller, func(c game.Card) bool {
						return c.IsCreature() && c.AttackingTarget != uuid.Nil && c.HasSubtype(treefolk)
					})
					if n <= 0 {
						return nil
					}
					return effects.AddMana{Amount: effects.ManaN(2*n, "G")}.Apply(ctx)
				})),
		},
	})
}
