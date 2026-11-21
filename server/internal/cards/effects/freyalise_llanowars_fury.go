package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Freyalise, Llanowar's Fury — Legendary Planeswalker — Freyalise
// {3}{G}{G}, loyalty 3 (Commander 2014, #5553):
//
//	"+2: Create a 1/1 green Elf Druid creature token with "{T}: Add {G}."
//	 −2: Destroy target artifact or enchantment.
//	 −6: Draw a card for each green creature you control.
//	 Freyalise, Llanowar's Fury can be your commander."
//
// The Elf Druid's mana ability rides its token template (token_catalog.go,
// the Human Monk's shape), so the token taps for {G} like a printed one. The
// −2 destroys the one target still legal as it resolves. The −6 counts the
// green creatures you control as it resolves, the Elf Druids included. "Can
// be your commander" is deck construction (CR 903.3a), read off the printed
// text by the deck validator.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "659dfdbf-5e9d-4828-a967-348cb3bcf05a",
		Name:            "Freyalise, Llanowar's Fury",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Create a 1/1 green Elf Druid creature token with \"{T}: Add {G}.\"",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:    LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenFromCatalog(printedC14ElfDruidToken), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "−2: Destroy target artifact or enchantment.",
				Cost:    LoyaltyCost(-2),
				Targets: TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				Effect:  DestroyFirstLegalTarget,
			},
			{
				Label: "−6: Draw a card for each green creature you control.",
				Cost:  LoyaltyCost(-6),
				Effect: Do(Instruction(func(ctx *Context) error {
					n := len(MatchingBattlefield(ctx, And(Creature(), YouControl(), OfColor("G"))))
					if n <= 0 {
						return nil
					}
					return DrawCards{N: n}.Apply(ctx)
				})),
			},
		},
	})
}

// printedC14ElfDruidToken is Freyalise's 1/1 green Elf Druid creature token
// with "{T}: Add {G}." (Commander 2014's token). Its slug is an on-disk
// identity.
func printedC14ElfDruidToken() TokenTemplate {
	return TokenTemplate{
		Slug: "elf-druid-tap-g",
		Card: game.Card{
			Name:      "Elf Druid",
			TypeLine:  "Token Creature — Elf Druid",
			Power:     1,
			Toughness: 1,
			Colors:    []string{"G"},
		},
		Mana: []game.ManaAbilityShape{{
			TapCost:  true,
			Produced: "{G}",
			Label:    "{T}: Add {G}",
		}},
		Text: "{T}: Add {G}.",
	}
}
