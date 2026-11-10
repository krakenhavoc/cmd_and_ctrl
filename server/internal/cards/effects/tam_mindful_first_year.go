package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tam, Mindful First-Year — Legendary Creature — Gorgon Wizard {1}{G/U}
// (Lorwyn Eclipsed, #4458):
//
//	"Each other creature you control has hexproof from each of its colors.
//	 {T}: Target creature you control becomes all colors until end of turn."
//
// "Hexproof from each of its colors" is one hexproof-from-<color> per color
// the creature has at layer 6 (colors are settled in layer 5), so a colorless
// creature gets nothing; colorless is not a color (the ruling). The second
// ability sets all five colors, so the creature has hexproof from every color
// until end of turn.
func init() {
	Register(Spec{
		OracleID:     "096a878e-cff0-430a-90b5-031ecb9bfe1b",
		Name:         "Tam, Mindful First-Year",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Each other creature you control has hexproof from each of its colors.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID != source.InstanceID && target.Controller == source.Controller && target.IsCreature()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, col := range c.Colors {
					if kw := game.HexproofFromColor(col); kw != "" {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
					}
				}
			},
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature you control becomes all colors until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return ScopedEffectFor{
					Target:   id,
					Mods:     []game.Mod{game.SetColorsMod(game.AllColors...)},
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Tam, Mindful First-Year — becomes all colors until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
