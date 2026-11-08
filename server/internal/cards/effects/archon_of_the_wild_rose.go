package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Archon of the Wild Rose — Creature — Archon {2}{W}{W}, 4/4 (Wilds of
// Eldraine, #4471):
//
//	"Flying
//	 Other creatures you control that are enchanted by Auras you control
//	 have base power and toughness 4/4 and have flying."
//
// Two layered statics over one group, re-read on every layer pass so a
// creature joins it when your Aura is attached and leaves it when the Aura
// falls off or changes hands: a layer 7b base-size set (CR 613.4b), the
// shape Zoetic Glyph's SetAttachedBasePT writes for one creature, and a
// layer 6 flying grant (CR 613.1f). The group is Eriette of the Charmed
// Apple's reading of "enchanted by an Aura you control" (an opponent's
// Aura on your creature does not count, nor your Aura on theirs), less
// the Archon itself. Counters and pumps still apply on top of the 4/4
// (7c and later).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "180efbd4-0f77-41cc-8ae2-2e6ccf853d10",
		Name:            "Archon of the Wild Rose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			{
				Label:     "Other creatures you control that are enchanted by Auras you control have base power and toughness 4/4.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: archonOfTheWildRoseGroup,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power = 4
					c.Toughness = 4
				},
			},
			Labelled(KeywordGrant(archonOfTheWildRoseGroup, "flying"),
				"Other creatures you control that are enchanted by Auras you control have flying."),
		},
	})
}

// archonOfTheWildRoseGroup is "other creatures you control that are
// enchanted by Auras you control".
func archonOfTheWildRoseGroup(target *game.Card, g *game.Game, source *game.Card) bool {
	if source == nil || g == nil || g.Battlefield == nil || target.InstanceID == source.InstanceID ||
		!target.IsCreature() || target.Controller != source.Controller {
		return false
	}
	for i := range g.Battlefield.Cards {
		a := &g.Battlefield.Cards[i]
		if a.Controller == source.Controller && a.IsAttachedTo(target.InstanceID) && a.IsAura() {
			return true
		}
	}
	return false
}
