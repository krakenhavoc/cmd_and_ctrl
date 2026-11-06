package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Xande, Dark Mage — Legendary Creature — Human Wizard {2}{U}{B}, 3/3
// (Final Fantasy, #4462):
//
//	"Menace
//	 Xande gets +1/+1 for each noncreature, nonland card in your
//	 graveyard."
//
// Jarad, Golgari Lich Lord's self pump (layer 7c) over a different
// count: the cards in its controller's graveyard that are neither
// creature nor land cards, read live as each recompute runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36b5c793-30f6-4f21-8bf6-c18d1c46d732",
		Name:            "Xande, Dark Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{{
			Label:     "Xande gets +1/+1 for each noncreature, nonland card in your graveyard.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := xandeNoncreatureNonlandCardsInGraveyard(g, source)
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}

// xandeNoncreatureNonlandCardsInGraveyard counts the noncreature,
// nonland cards in the source's controller's graveyard.
func xandeNoncreatureNonlandCardsInGraveyard(g *game.Game, source *game.Card) int {
	p := g.PlayerByIDForEffect(source.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if !c.IsCreature() && !c.IsLand() {
			n++
		}
	}
	return n
}
