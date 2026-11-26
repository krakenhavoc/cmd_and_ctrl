package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Skarrg Guildmage — Creature — Human Shaman {R}{G}, 2/2 (Gatecrash,
// #5565):
//
//	"{R}{G}: Creatures you control gain trample until end of turn.
//	 {1}{R}{G}: Target land you control becomes a 4/4 Elemental creature
//	 until end of turn. It's still a land."
//
// The first ability locks the set of creatures as it resolves (CR 611.2c),
// so a creature that arrives later this turn does not gain trample. The
// second adds the Creature type and the Elemental subtype (layer 4) and sets
// base power and toughness 4/4 (layer 7b) until end of turn on the targeted
// land (Mishra's Foundry's animation); the land keeps its land type and
// abilities. The target must still be a land you control as the ability
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a49f3a1-abdb-4bf3-b5ed-f54349ab40fd",
		Name:         "Skarrg Guildmage",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}{G}: Creatures you control gain trample until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, TeamPump: &game.Pump{Keywords: []string{"trample"}}},
				Cost:    effects.ManaCost("{R}{G}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.GrantKeywordUntilEOT{
						Match:    effects.And(effects.Creature(), effects.YouControl()),
						Keywords: []string{"trample"},
						Label:    "Skarrg Guildmage — creatures you control gain trample",
					}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "{1}{R}{G}: Target land you control becomes a 4/4 Elemental creature until end of turn. It's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{1}{R}{G}"),
				Targets: effects.TargetPermanent("target land you control", effects.Land(), effects.YouControl()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					land := effects.FirstLegalBattlefieldTarget(ctx)
					if land == uuid.Nil {
						return nil
					}
					return effects.ScopedEffectFor{
						Target: land,
						Mods: append([]game.Mod{
							game.AddTypesMod("Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Elemental")...),
						}, game.SetBasePTMods(4, 4)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Skarrg Guildmage — becomes a 4/4 Elemental creature until end of turn",
					}.Apply(ctx)
				},
			},
		},
	})
}
