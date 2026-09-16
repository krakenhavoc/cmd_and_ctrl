package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Insidious Fungus — Creature — Fungus {G}, 1/2 (EDHREC rank 1900):
//
//	"{2}, Sacrifice this creature: Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Draw a card. Then you may put a land card from your hand onto
//	   the battlefield tapped."
//
// The one-mana modal utility creature. A modal ACTIVATED ability has
// no shape in the catalog (Spec.Modes is cast-time only), so the
// three modes are three activated abilities with the same cost, one
// per bullet — the mode is chosen at activation either way (CR
// 700.2), and the client's ability menu is the mode picker. The two
// removal modes are targeted; the third is not.
//
// Sandbox simplification, declared (the Stoneforge Mystic posture):
// the third mode's "you may put a land card from your hand onto the
// battlefield tapped" is not implemented — a pick-from-hand prompt
// with a hand-to-battlefield move does not exist (the Growth Spiral
// gap). That mode draws its card and stops. Weaker than printed,
// never stronger.
func init() {
	fungusCost := Plus(ManaCost("{2}"), SacrificeThis())
	Register(Spec{
		OracleID:     "0a8d0217-ff24-4177-b6be-707eb2b6b9e9",
		Name:         "Insidious Fungus",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The third mode only draws the card — it doesn't offer to put a land from your hand onto the battlefield."},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}, Sacrifice Insidious Fungus: Destroy target artifact",
				Cost:    fungusCost,
				Targets: TargetPermanent("target artifact", Artifact()),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label:   "{2}, Sacrifice Insidious Fungus: Destroy target enchantment",
				Cost:    fungusCost,
				Targets: TargetPermanent("target enchantment", Enchantment()),
				Effect:  destroyFirstLegalTarget,
			},
			{
				Label: "{2}, Sacrifice Insidious Fungus: Draw a card",
				Cost:  fungusCost,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
