package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sydri, Galvanic Genius — Legendary Creature — Human Artificer {W}{U}{B}, 2/2
// (Commander 2016, #5537):
//
//	"{U}: Target noncreature artifact becomes an artifact creature with power
//	 and toughness each equal to its mana value until end of turn.
//	 {W}{B}: Target artifact creature gains deathtouch and lifelink until end of
//	 turn."
//
// The first is Xenic Poltergeist's animation (ArtifactBecomesCreatureOfItsManaValue,
// mana value read as it resolves, CR 608.2h) until end of turn; the second grants
// two keywords to a still-legal target (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4e92d36b-1a35-4fa9-87ea-10eace5a3cc7",
		Name:         "Sydri, Galvanic Genius",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Target noncreature artifact becomes an artifact creature with power and toughness each equal to its mana value until end of turn.",
				Cost:    effects.ManaCost("{U}"),
				Targets: effects.TargetPermanent("target noncreature artifact", effects.And(effects.Artifact(), effects.Noncreature())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ArtifactBecomesCreatureOfItsManaValue(ctx, effects.DurationUntilEndOfTurn(ctx),
						"Sydri, Galvanic Genius — an artifact creature until end of turn")
				},
			},
			{
				Label:   "{W}{B}: Target artifact creature gains deathtouch and lifelink until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch", "lifelink")), game.AnswerPump),
				Cost:    effects.ManaCost("{W}{B}"),
				Targets: effects.TargetCreature("target artifact creature", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					target := effects.FirstLegalBattlefieldTarget(ctx)
					if target == uuid.Nil {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch", "lifelink"},
						Label: "Sydri, Galvanic Genius — deathtouch and lifelink until end of turn"}.Apply(ctx)
				},
			},
		},
	})
}
