package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Zinnia, Valley's Voice — Legendary Creature — Bird Bard {U}{R}{W}, 1/3
// (Bloomburrow Commander, #5428):
//
//	"Flying
//	 Zinnia gets +X/+0, where X is the number of other creatures you
//	 control with base power 1.
//	 Creature spells you cast gain offspring {2} as you cast them. (You
//	 may pay an additional {2} as you cast a creature spell. If you do,
//	 when that creature enters, create a 1/1 token copy of it.)"
//
// The pump is a layer 7c self-pump that counts the other creatures its
// controller controls by their base power (CR 208.4b, game.Card.BasePower:
// the power after layer 7b, so an anthem or a counter on a 1/1 does not
// stop it counting, and one on a 0/1 does not make it count).
//
// The grant is OffspringForSpellsYouCast (granted_offspring.go): each
// creature spell its controller casts offers "Offspring {2}" beside its
// own optional costs, from any zone, and the spell that paid it becomes
// a permanent with offspring's enters trigger (CR 702.175a). The spell
// gained the keyword as it was cast (CR 610.5), so the token is made even
// if Zinnia has left by then, and a creature that prints offspring may
// pay both costs and gets a token for each (CR 702.175b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d8c16a3a-dd9d-410c-a5d5-06339a89970f",
		Name:            "Zinnia, Valley's Voice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:    "Gets +X/+0, where X is the number of other creatures you control with base power 1.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				c.Power += zinniaOtherCreaturesWithBasePowerOne(g, source)
			},
		}},
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			OffspringForSpellsYouCast("Creature spells you cast gain offspring {2} as you cast them.", Creature(), "{2}"),
		},
	})
}

// zinniaOtherCreaturesWithBasePowerOne counts the creatures other than
// `source` its controller controls whose base power is 1.
func zinniaOtherCreaturesWithBasePowerOne(g *game.Game, source *game.Card) int {
	n := 0
	for _, id := range permanentsControlledByMatching(g, source.Controller, Creature()) {
		if id == source.InstanceID {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.BasePower() == 1 {
			n++
		}
	}
	return n
}
