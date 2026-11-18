package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// One with the Kami — Enchantment — Aura {3}{G} (Kamigawa: Neon Dynasty):
//
//	"Flash
//	 Enchant creature you control
//	 Whenever enchanted creature or another modified creature you control
//	 dies, create X 1/1 colorless Spirit creature tokens, where X is that
//	 creature's power. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// One trigger on a creature dying. It fires for the enchanted creature
// (attached as it died, Decomposition's read) and for any other creature its
// controller controlled that was modified as it last existed (CR 700.9,
// 603.10a: msc17WasModified reads the departure record, so an Aura or
// Equipment that fell off with it still counts). X is the dead creature's
// last-known power (CR 608.2h), never below zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "006f4a62-8590-4ca0-9f51-3aea26cd1a54",
		Name:            "One with the Kami",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(YouControl()),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, oneWithTheKamiTriggers,
				"One with the Kami — create X 1/1 colorless Spirit creature tokens, where X is that creature's power",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					x := b17LastKnownPowerOffBattlefield(g, item.Trigger.Event.CardID)
					if x <= 0 {
						return nil
					}
					return CreateToken{Controller: item.Controller, Template: TokenCard("1/1 colorless Spirit"), N: x}.Apply(NewContext(g, item))
				}),
		},
	})
}

// oneWithTheKamiTriggers is "enchanted creature or another modified creature
// you control dies".
func oneWithTheKamiTriggers(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	if source.IsAttachedTo(dead.InstanceID) {
		return true
	}
	return dead.InstanceID != source.InstanceID && leftUnderControlOf(ev, dead) == source.Controller &&
		msc17WasModified(g, dead.InstanceID)
}
