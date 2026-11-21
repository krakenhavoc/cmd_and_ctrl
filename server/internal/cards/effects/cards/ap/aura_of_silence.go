package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aura of Silence — Enchantment {1}{W}{W}:
//
//	"Artifact and enchantment spells your opponents cast cost {2}
//	 more to cast.
//	 Sacrifice this enchantment: Destroy target artifact or
//	 enchantment."
//
// The first ASYMMETRIC cost modifier in the catalog: OpponentsSpell()
// excludes the controller's own board, which is what separates this
// from the Sphere / Thorn / Thalia group and is the half a card file
// would most easily get wrong. A version without the predicate would
// tax its own controller's mana rocks and look like it was working.
//
// The second ability is an ordinary sacrifice-this activated ability
// — no mana, no tap, so it fires at instant speed the turn it lands
// and the tax is really "pay {2} more, and I still get to blow up
// the thing you paid for".
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7faf8eb-e829-4109-8dfe-42865a23ba86",
		Name:         "Aura of Silence",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsMore(2, "Artifact and enchantment spells your opponents cast cost {2} more to cast.",
				effects.OpponentsSpell(), effects.ArtifactOrEnchantmentSpell()),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Destroy target artifact or enchantment.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
