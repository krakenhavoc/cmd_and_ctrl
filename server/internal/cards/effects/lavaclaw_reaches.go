package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "340de307-982a-4b26-9dc0-99113dc766cd",
		Name:         "Lavaclaw Reaches",
		XMatters:     true,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|R}",
			Label:    "Add {B} or {R}",
		}},
		Grants: []AbilityGrant{{
			Key: grantKey,
			Activated: []ActivatedAbility{{
				Label:   "{X}: This creature gets +X/+0 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    ManaCost("{X}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return BoostUntilEOT{
						Target: item.SourceCardID,
						Power:  ctx.X(),
						Label:  "Lavaclaw Reaches — +X/+0",
					}.Apply(ctx)
				},
			}},
			Text: "{X}: This creature gets +X/+0 until end of turn.",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}{R}: Until end of turn, this land becomes a 2/2 black and red Elemental creature with \"{X}: This creature gets +X/+0 until end of turn.\" It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerAnimate},
			Cost:    ManaCost("{1}{B}{R}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !sourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := NewContext(g, item)
				return GrantAbilitiesFor{
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
