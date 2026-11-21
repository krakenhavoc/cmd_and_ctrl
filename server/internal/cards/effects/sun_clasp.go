package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun Clasp — Enchantment — Aura {1}{W} (Dominaria Remastered, #5948):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+3.
//	 {W}: Return enchanted creature to its owner's hand."
//
// The host is found through the Aura as the ability resolves; an Aura that
// has fallen off by then returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb8f5ee4-285a-4c37-adcb-514b3e4bdad3",
		Name:         "Sun Clasp",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(1, 3)},
		Activated: []ActivatedAbility{{
			Label:   "{W}: Return enchanted creature to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    ManaCost("{W}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				host := attachedHostFor(g, item.SourceCardID)
				if host == nil {
					return nil
				}
				return BounceToHand{Target: host.InstanceID}.Apply(NewContext(g, item))
			},
		}},
	})
}
