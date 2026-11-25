package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret the Schemer — Legendary Planeswalker — Tezzeret {2}{U}{B},
// loyalty 5 (Aether Revolt, S82 older sets, #5536):
//
//	"+1: Create a colorless artifact token named Etherium Cell with "{T},
//	 Sacrifice this token: Add one mana of any color."
//	 −2: Target creature gets +X/−X until end of turn, where X is the number
//	 of artifacts you control.
//	 −7: You get an emblem with "At the beginning of combat on your turn,
//	 target artifact you control becomes an artifact creature with base power
//	 and toughness 5/5.""
//
// The −2 counts artifacts as it resolves (CR 608.2h). The emblem's trigger
// targets as it goes on the stack and its effect has no duration (CR 611.2a):
// the artifact stays a 5/5 artifact creature, in the layers a type-adding
// effect (layer 4) and a base power and toughness setting effect (layer 7b)
// at the emblem trigger's timestamp.
//
// No simplification.
func init() {
	const emblemText = "At the beginning of combat on your turn, target artifact you control becomes an artifact creature with base power and toughness 5/5."
	effects.Register(effects.Spec{
		OracleID:        "49d64ecb-3a4f-4013-9a34-8b1539a91e96",
		Name:            "Tezzeret the Schemer",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Tezzeret the Schemer emblem",
			Text:  emblemText,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.AtBeginningOfYourCombat("Tezzeret the Schemer emblem — target artifact you control becomes an artifact creature with base power and toughness 5/5",
					tezzeretTheSchemerAnimate),
					effects.TargetPermanent("target artifact you control", effects.Artifact(), effects.YouControl())),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Create a colorless artifact token named Etherium Cell with \"{T}, Sacrifice this token: Add one mana of any color.\"",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  effects.Do(effects.CreateToken{Template: effects.S82OldB254EtheriumCellToken(), N: 1}),
			},
			{
				Label:   "−2: Target creature gets +X/−X until end of turn, where X is the number of artifacts you control.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetCreature("target creature"),
				Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					x := effects.CountControlled(g, ctx.Controller(), effects.MatchArtifact)
					return effects.BoostUntilEOT{Target: id, Power: x, Toughness: -x, Label: "Tezzeret the Schemer — +X/-X"}.Apply(ctx)
				},
			},
			{
				Label:  "−7: You get an emblem with \"" + emblemText + "\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// tezzeretTheSchemerAnimate makes the target an artifact creature with base
// power and toughness 5/5, for no stated duration.
func tezzeretTheSchemerAnimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok || !effects.OnBattlefield(g, id) {
		return nil
	}
	mods := append([]game.Mod{game.AddTypesMod("Creature")}, game.SetBasePTMods(5, 5)...)
	g.RegisterScopedEffectForEffect(ctx.Source(),
		[]game.AffectedObject{game.PinObject(id, c.EnteredBattlefieldAt)}, mods,
		g.PinnedTo(game.IndefiniteDuration(), id),
		"Tezzeret the Schemer emblem — becomes an artifact creature with base power and toughness 5/5")
	return nil
}
