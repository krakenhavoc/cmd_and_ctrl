package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Golden Guardian // Gold-Forge Garrison — Artifact Creature — Golem {4}, 4/4 //
// Land:
//
//	Golden Guardian: "Defender
//	 {2}: This creature fights another target creature you control. When this
//	 creature dies this turn, return it to the battlefield transformed under
//	 your control."
//	Gold-Forge Garrison: "{T}: Add two mana of any one color.
//	 {4}, {T}: Create a 4/4 colorless Golem artifact creature token."
//
// The fight is Polukranos' shape (B30SourceFightsFirstLegalTarget's own fight,
// a target that is another creature the controller controls). Each resolution
// that fights also creates a delayed trigger about THIS object (Felonious Rage's
// "when that creature dies this turn", CR 603.7c): if it dies this turn it comes
// back from the graveyard with its back face up, under the ability controller's
// control (CR 712.14a). A trigger that finds the card no longer in a graveyard
// (it came back already, or was exiled in response) does nothing. The back face
// is a land: its first ability is a mana ability, taps for two mana of one
// color (CR 605.1a), and the second makes a 4/4 colorless Golem artifact
// creature token.
//
// No simplification.
func init() {
	const oracle = "58afb897-4d57-4b53-a5c3-b532cb3d5180"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Golden Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This creature fights another target creature you control. When this creature dies this turn, return it to the battlefield transformed under your control.",
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect:  goldenGuardianFight,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Gold-Forge Garrison",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: effects.OneColorOfAmount(2),
			Label:    "{T}: Add two mana of any one color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}, {T}: Create a 4/4 colorless Golem artifact creature token.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{4}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.CreateToken{Template: effects.TokenCard("4/4 colorless Golem artifact creature"), N: 1}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

// goldenGuardianFight fights the chosen creature and, once it has, sets up the
// "when this creature dies this turn" return for the Guardian as it is now.
func goldenGuardianFight(g *game.Game, item *game.StackItem) error {
	if err := effects.B30SourceFightsFirstLegalTarget(g, item); err != nil {
		return err
	}
	ctx := effects.NewContext(g, item)
	return effects.WhenThatCreatureDiesThisTurn(ctx, ctx.Source(),
		"Golden Guardian — return it to the battlefield transformed under your control", goldenGuardianReturn)
}

// goldenGuardianReturn returns the Guardian that died, transformed, under the
// control of the ability's controller. An on-disk key: never renamed.
var goldenGuardianReturn = game.SimpleDelayedBody("golden-guardian/return-transformed", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	obj, ok := effects.PayloadObject(ctx)
	if !ok || obj.ID == uuid.Nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(obj.ID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return effects.ReturnFromGraveyard{Target: obj.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Transformed: true}.Apply(ctx)
})
