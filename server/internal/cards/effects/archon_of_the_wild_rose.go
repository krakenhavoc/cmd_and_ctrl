package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of the Wild Rose — Creature — Archon {2}{W}{W}, 3/3 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Other creatures you control that are enchanted by Auras you control have
//	 base power and toughness 4/4 and have flying."
//
// Two layers off one test (archonEnchantedByYourAura): a layer 7b base-P/T set
// (CR 613.4b), which later set effects overwrite and counters and modifiers
// still sit on top of (the ruling), and a layer 6 flying grant (CR 613.1f).
// "Enchanted by Auras you control" is read from the battlefield on every pass,
// so it follows an Aura moving or leaving. The Archon itself is excluded
// ("other"), and an Aura its controller does not control does not count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "180efbd4-0f77-41cc-8ae2-2e6ccf853d10",
		Name:            "Archon of the Wild Rose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Other creatures you control that are enchanted by Auras you control have base power and toughness 4/4.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: archonEnchantedByYourAura,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power = 4
				c.Toughness = 4
			},
		}, {
			Label:     "Other creatures you control that are enchanted by Auras you control have flying.",
			Layer:     game.Layer6Ability,
			AppliesTo: archonEnchantedByYourAura,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
			},
		}},
	})
}

// archonEnchantedByYourAura is "other creature you control that is enchanted
// by an Aura you control".
func archonEnchantedByYourAura(target *game.Card, g *game.Game, source *game.Card) bool {
	if g == nil || g.Battlefield == nil || target.InstanceID == source.InstanceID ||
		target.Controller != source.Controller || !target.IsCreature() {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.Controller == source.Controller && a.IsAura() && a.IsAttachedTo(target.InstanceID) {
			return true
		}
	}
	return false
}
