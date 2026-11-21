package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aggressive Biomancy — Sorcery {X}{X}{G}{U} (Modern Horizons 3 Commander,
// #5948):
//
//	"Create X tokens that are copies of target creature you control, except
//	 they have "When this token enters, it fights up to one target creature
//	 you don't control.""
//
// Doppelgang's creation (CreateTokenCopiesTogether): X copies of the one
// target creature, all created and entering together (CR 608.2f), so each
// sees the others enter and a "one or more" trigger sees one occurrence. The
// quoted ability is part of the copiable values (CR 707.9a), a bundle in
// Spec.Grants granted through the copy's exception, so a token copy of one of
// these tokens fights too. Each token has its own "up to one target creature
// you don't control" (Apex Altisaur's clause), chosen as its trigger goes on
// the stack; with no target it fights nothing. The cost has two {X} (the
// engine charges 2X generic), and X = 0 makes nothing. A target that left in
// response makes the spell fizzle (CR 608.2b).
//
// No simplification.
func init() {
	fights := func() *game.TargetSpec {
		return effects.TargetCreature("up to one target creature you don't control", effects.OpponentControls()).WithCount(0, 1)
	}
	effects.Register(effects.Spec{
		OracleID:     "9080a7bb-2b8b-42af-9f5b-a9f374ecb801",
		Name:         "Aggressive Biomancy",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Grants: []effects.AbilityGrant{{
			Key: aggressiveBiomancyGrant,
			Triggered: []game.TriggeredAbility{
				effects.Targeting(effects.WhenThisEnters("Aggressive Biomancy — it fights up to one target creature you don't control", effects.ItFightsTheFirstTarget), fights()),
			},
			Text: "When this token enters, it fights up to one target creature you don't control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			if x <= 0 {
				return nil
			}
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				return effects.CreateTokenCopiesTogether{
					Controller: ctx.Controller(),
					Copies:     []uuid.UUID{t.ID},
					N:          x,
					Except: func(tok *game.Card) {
						v := game.PrintedValues{TypeLine: tok.TypeLine, GrantedAbilities: tok.GrantedAbilities}
						v.GrantAbility(aggressiveBiomancyGrant)
						tok.GrantedAbilities = v.GrantedAbilities
					},
				}.Apply(ctx)
			}
			return nil
		},
	})
}

const aggressiveBiomancyGrant = "aggressive-biomancy/token-fights"
