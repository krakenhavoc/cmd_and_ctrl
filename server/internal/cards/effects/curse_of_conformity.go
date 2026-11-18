package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Curse of Conformity — Enchantment — Aura Curse {4}{W}:
//
//	"Enchant player
//	 Nonlegendary creatures enchanted player controls have base power and
//	 toughness 3/3 and lose all creature types."
//
// Two statics over one set (curseOfConformityApplies, read on every
// layer pass, so it follows the Curse moving to another player, leaving,
// or a creature changing hands): a layer 4 loss of every creature type
// (CR 613.1d; a changeling's "is every creature type" goes with them) and
// a layer 7b base-power-and-toughness set (CR 613.4b), under which counters
// and modifiers still apply. A legendary creature is untouched, judged by the
// supertype it has as the layers are applied. Other card types it has stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71de5f9a-73e4-4a93-b199-b280687f71eb",
		Name:         "Curse of Conformity",
		Completeness: CompletenessFull,
		Targets:      EnchantPlayer(),
		Static: []game.StaticAbility{{
			Label:     "Nonlegendary creatures enchanted player controls lose all creature types.",
			Layer:     game.Layer4Type,
			AppliesTo: curseOfConformityApplies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.SetCreatureTypes(nil)
			},
		}, {
			Label:     "Nonlegendary creatures enchanted player controls have base power and toughness 3/3.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: curseOfConformityApplies,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 3
				c.Toughness = 3
			},
		}},
	})
}

// curseOfConformityApplies is "nonlegendary creatures enchanted player
// controls".
func curseOfConformityApplies(target *game.Card, _ *game.Game, source *game.Card) bool {
	return source.AttachedTo.Kind == game.TargetPlayer &&
		target.Controller == source.AttachedTo.ID &&
		target.IsCreature() && !target.HasSupertype("legendary")
}
