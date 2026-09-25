package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ad01df89-29fe-44c7-a133-91425f8ff09c",
		Name:         "Fake Your Own Death",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: fakeYourOwnDeathReturn,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Fake Your Own Death — return it and make a Treasure",
					returnThisCreatureFromGraveyard(nil, func(ctx *Context) error {
						return CreateToken{Template: TreasureToken(), N: 1}.Apply(ctx)
					})),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control and you create a Treasure token.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{fakeYourOwnDeathReturn},
				Also:   []game.Mod{game.ModifyPTMod(2, 0)},
				Label:  "Fake Your Own Death — +2/+0 and it returns when it dies",
			}.Apply(ctx)
		},
	})
}
