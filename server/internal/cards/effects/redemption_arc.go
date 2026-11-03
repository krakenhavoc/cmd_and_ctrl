package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redemption Arc — Enchantment — Aura {2}{W} (Secrets of Strixhaven
// Commander, #4834):
//
//	"Enchant creature
//	 Enchanted creature has indestructible and is goaded. (It attacks
//	 each combat if able and attacks a player other than you if able.)
//	 {1}{W}: Exile enchanted creature."
//
// Coercive Impetus's goad pair (GoadAttached, which ends when the Aura
// falls off) beside Darksteel Plate's indestructible grant. The activated
// ability names the creature through the Aura's attachment as it resolves,
// or as it last was if the Aura has already gone (CR 608.2h), the way
// Convenient Target's trigger does, and exiles it. It does not target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c22e5fc6-9b4e-4197-b472-70d168eb4330",
		Name:         "Redemption Arc",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			Labelled(GrantToAttached("indestructible"), "Enchanted creature has indestructible."),
			GoadAttached(),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{W}: Exile enchanted creature.",
			Cost:  ManaCost("{1}{W}"),
			// Removes a permanent with no target (ADR 0142's AnswerRemove).
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  exileEnchantedCreature,
		}},
	})
}

// exileEnchantedCreature is "exile enchanted creature" for an Aura's own
// ability.
func exileEnchantedCreature(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	return ExileTarget{Target: host}.Apply(ctx)
}
