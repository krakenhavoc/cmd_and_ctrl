package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lavaclaw Reaches — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B} or {R}.
//	 {1}{B}{R}: Until end of turn, this land becomes a 2/2 black and red
//	 Elemental creature with "{X}: This creature gets +X/+0 until end of
//	 turn." It's still a land."
//
// Raging Ravine's shape: one animation that grants an ability bundle (here
// the activated pump) with the effect, belonging to the land (ADR 0093). The
// granted {X} pump is Kessig Wolf Run's BoostUntilEOT on the source.
//
// No simplification.
func init() {
	const grantKey = "lavaclaw-reaches/x-pump"
	effects.Register(effects.Spec{
		OracleID:     "340de307-982a-4b26-9dc0-99113dc766cd",
		Name:         "Lavaclaw Reaches",
		XMatters:     true,
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Grants: []effects.AbilityGrant{{
			Key: grantKey,
			Activated: []effects.ActivatedAbility{{
				Label:   "{X}: This creature gets +X/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{X}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  ctx.X(),
						Label:  "Lavaclaw Reaches — +X/+0",
					}.Apply(ctx)
				},
			}},
			Text: "{X}: This creature gets +X/+0 until end of turn.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}{R}: Until end of turn, this land becomes a 2/2 black and red Elemental creature with \"{X}: This creature gets +X/+0 until end of turn.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Cost:    effects.ManaCost("{1}{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.GrantAbilitiesFor{
					Target: item.SourceCardID,
					Keys:   []string{grantKey},
					Also: []game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						game.SetColorsMod(ctx.TextColors("B", "R")...),
						game.SetBasePowerMod(2),
						game.SetBaseToughnessMod(2),
					},
					Label: "Lavaclaw Reaches — a 2/2 black and red Elemental creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}
