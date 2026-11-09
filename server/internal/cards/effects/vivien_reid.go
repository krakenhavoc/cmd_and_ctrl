package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivien Reid — Legendary Planeswalker — Vivien {3}{G}{G}, loyalty 5
// (Foundations, #4465):
//
//	"+1: Look at the top four cards of your library. You may reveal a
//	 creature or land card from among them and put it into your hand. Put
//	 the rest on the bottom of your library in a random order.
//	 −3: Destroy target artifact, enchantment, or creature with flying.
//	 −8: You get an emblem with "Creatures you control get +2/+2 and have
//	 vigilance, trample, and indestructible.""
//
// The +1 is Horn of the Mark's sentence (LookAtTopThenMayTakeToHand): a
// private look, an optional reveal-and-take of one creature or land card, and
// the rest on the bottom in a random order. The −3 destroys the still-legal
// target (CR 608.2b); "with flying" modifies only the creature, and is read
// through the layers. The emblem is Elspeth, Sun's Champion's shape: a layer
// 7c +2/+2 and the keyword grants (layer 6) on one emblem, for the emblem's
// owner's creatures, indestructible included (CR 114).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5f20a6d-8c3e-452f-9d98-4886f0fa052a",
		Name:            "Vivien Reid",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
			Label: "Vivien Reid emblem",
			Text:  "Creatures you control get +2/+2 and have vigilance, trample, and indestructible.",
			Static: []game.StaticAbility{
				{
					Layer:     game.Layer7PT,
					SubLayer:  game.SubLayer7C_Modify,
					AppliesTo: creaturesTheEmblemsOwnerControls,
					Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
						c.Power += 2
						c.Toughness += 2
					},
				},
				KeywordGrant(creaturesTheEmblemsOwnerControls, "vigilance"),
				KeywordGrant(creaturesTheEmblemsOwnerControls, "trample"),
				KeywordGrant(creaturesTheEmblemsOwnerControls, "indestructible"),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Look at the top four cards of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
				Purpose: game.Purpose{Draws: 1},
				Cost:    LoyaltyCost(1),
				Effect: LookAtTopThenMayTakeToHand(4, Or(Creature(), Land()), 1,
					"Vivien Reid — you may reveal a creature or land card and put it into your hand"),
			},
			{
				Label: "−3: Destroy target artifact, enchantment, or creature with flying.",
				Cost:  LoyaltyCost(-3),
				Targets: TargetPermanent("target artifact, enchantment, or creature with flying",
					Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))),
				Effect: destroyFirstLegalTarget,
			},
			{
				Label:  "−8: You get an emblem with \"Creatures you control get +2/+2 and have vigilance, trample, and indestructible.\"",
				Cost:   LoyaltyCost(-8),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}
