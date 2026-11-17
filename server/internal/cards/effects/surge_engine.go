package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surge Engine — Artifact Creature — Construct {2}, 3/2 (The Brothers'
// War, #5482):
//
//	"Defender
//	 {U}: This creature loses defender and gains 'This creature can't be
//	 blocked.'
//	 {2}{U}: This creature becomes blue and has base power and toughness
//	 5/4. Activate only if this creature doesn't have defender.
//	 {4}{U}{U}: Draw three cards. Activate only if this creature is blue
//	 and only once."
//
// The first two print no duration, so each lasts for the rest of the game
// (CR 611.2a), pinned to this object (CR 400.7): the first is one effect
// removing defender (layer 6) and adding "can't be blocked"; the second
// sets blue (layer 5) and base 5/4 (layer 7b). The activation conditions
// read the layered view, so losing defender (or being made blue by
// something else) counts. "Only once" is the OnlyOnce bit (CR 602.5b),
// once per object.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "864d9db7-34ca-4440-8d6d-096f9eddedd4",
		Name:            "Surge Engine",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []ActivatedAbility{
			{
				Label:   "{U}: This creature loses defender and gains \"This creature can't be blocked.\"",
				Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
				Cost:    ManaCost("{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: []game.Mod{
							game.RemoveKeywordsMod("defender"),
							game.AddRestrictionsMod(game.CantBeBlocked),
						},
						Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
						Label:    "Surge Engine — loses defender and can't be blocked",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{2}{U}: This creature becomes blue and has base power and toughness 5/4. Activate only if this creature doesn't have defender.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    ManaCost("{2}{U}"),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && !game.HasKeyword(&c, "defender")
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := NewContext(g, item)
					return ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{game.SetColorsMod(ctx.TextColors("U")...)},
							game.SetBasePTMods(5, 4)...),
						Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
						Label:    "Surge Engine — becomes blue with base power and toughness 5/4",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{4}{U}{U}: Draw three cards. Activate only if this creature is blue and only once.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    ManaCost("{4}{U}{U}"),
				Condition: func(g *game.Game, _, source uuid.UUID) bool {
					c, ok := g.LookupCardForEffect(source)
					return ok && c.HasColorWord("U")
				},
				OnlyOnce: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
				},
			},
		},
	})
}
