package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bleeding Effect — Enchantment {2}{W}{B} (Assassin's Creed, #5948):
//
//	"At the beginning of combat on your turn, creatures you control gain flying
//	 until end of turn if a creature card in your graveyard has flying. The same is
//	 true for first strike, double strike, deathtouch, hexproof, indestructible,
//	 lifelink, menace, reach, trample, and vigilance."
//
// Each of the eleven keywords is its own condition, all read as the trigger
// resolves from the creature cards in your graveyard (their printed keywords, where
// a card in a graveyard has no other source): the creatures that gain are the
// creatures you control as it resolves, the set locked then (CR 611.2c), each
// gaining every keyword some graveyard creature has. A keyword nothing in the
// graveyard has is not given.
//
// No simplification.
func init() {
	const label = "Bleeding Effect — creatures you control gain each keyword a creature card in your graveyard has until end of turn"
	Register(Spec{
		OracleID:     "689c9a49-d0f5-494a-9fa2-948cd5517e74",
		Name:         "Bleeding Effect",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat(label, func(g *game.Game, item *game.StackItem) error {
				p := g.PlayerByIDForEffect(item.Controller)
				if p == nil || p.Graveyard == nil {
					return nil
				}
				var gained []string
				for _, kw := range bleedingEffectKeywords {
					for i := range p.Graveyard.Cards {
						c := p.Graveyard.Cards[i]
						if c.IsCreature() && game.HasKeyword(&c, kw) {
							gained = append(gained, kw)
							break
						}
					}
				}
				if len(gained) == 0 {
					return nil
				}
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: gained,
					Label:    label,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}

// bleedingEffectKeywords are the eleven keywords Bleeding Effect names, in
// the printed order.
var bleedingEffectKeywords = []string{
	"flying", "first strike", "double strike", "deathtouch", "hexproof", "indestructible",
	"lifelink", "menace", "reach", "trample", "vigilance",
}
