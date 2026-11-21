package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fake Your Own Death — Instant {1}{B}:
//
//	"Until end of turn, target creature gets +2/+0 and gains "When
//	 this creature dies, return it to the battlefield tapped under its
//	 owner's control and you create a Treasure token.""
//
// A duration grant (ADR 0093 PR 4, #1584). "Gets +2/+0 and gains …"
// is ONE effect, so the boost and the grant are one ScopedEffect
// record at one timestamp (GrantAbilitiesFor.Also).
//
// "You" in the granted trigger is the trigger's controller — the
// creature's controller as it died (ADR 0093 Decision 4), not the
// spell's caster. That is who gets the Treasure: an opponent's
// creature you cast this on makes a Treasure for its controller. The
// Treasure is made whether or not the creature came back ("and", not
// "if you do"), so a token that dies still makes one.
//
// No simplification.
const fakeYourOwnDeathReturn = "fake-your-own-death/return"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad01df89-29fe-44c7-a133-91425f8ff09c",
		Name:         "Fake Your Own Death",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Grants: []effects.AbilityGrant{{
			Key: fakeYourOwnDeathReturn,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Fake Your Own Death — return it and make a Treasure",
					effects.ReturnThisCreatureFromGraveyard(nil, func(ctx *effects.Context) error {
						return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
					})),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control and you create a Treasure token.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{fakeYourOwnDeathReturn},
				Also:   []game.Mod{game.ModifyPTMod(2, 0)},
				Label:  "Fake Your Own Death — +2/+0 and it returns when it dies",
			}.Apply(ctx)
		},
	})
}
