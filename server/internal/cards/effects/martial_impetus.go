package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Martial Impetus — Enchantment — Aura {2}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and is goaded. (It attacks each combat
//	 if able and attacks a player other than you if able.)
//	 Whenever enchanted creature attacks, each other creature that's
//	 attacking one of your opponents gets +1/+1 until end of turn."
//
// Shiny Impetus's goad (GoadAttached, CR 701.15b) and pump. The attack
// trigger is the Aura's, so "your opponents" are the Aura controller's.
// "Attacking one of your opponents" is attacking that player, not a
// planeswalker or battle (CR 508.1b), and the creatures are chosen as the
// trigger resolves and locked in then (CR 611.2c). The enchanted creature
// is the one that attacked, named by the trigger's event.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c8decfdf-1421-4649-acde-5215e7f08bb9",
		Name:         "Martial Impetus",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GoadAttached(),
		},
		Triggered: []game.TriggeredAbility{
			WheneverEnchantedCreatureAttacks("Martial Impetus — each other creature attacking one of your opponents gets +1/+1",
				martialImpetusPump),
		},
	})
}

// martialImpetusPump gives +1/+1 until end of turn to each creature other
// than the enchanted one that is attacking an opponent of the Aura's
// controller.
func martialImpetusPump(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	enchanted := item.Trigger.Event.CardID
	return BoostUntilEOT{
		Match: func(g *game.Game, controller uuid.UUID, c game.Card) bool {
			return c.IsCreature() && c.InstanceID != enchanted && socB8AttackingAnOpponentOf(g, controller, c)
		},
		Power:     1,
		Toughness: 1,
		Label:     "Martial Impetus — +1/+1",
	}.Apply(NewContext(g, item))
}
