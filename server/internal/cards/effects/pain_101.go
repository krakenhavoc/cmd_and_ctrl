package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pain 101 — {1}{B} Instant (Teenage Mutant Ninja Turtles, #4457):
//
//	"Until end of turn, target creature gains deathtouch and "When this
//	 creature dies, return it to the battlefield tapped under its owner's
//	 control.""
//
// Fake Your Own Death's duration grant (ADR 0093): the keyword and the
// granted trigger are one effect at one timestamp (GrantAbilitiesFor.Also).
// The trigger is the creature's own, so a token that dies does not come
// back (CR 111.7) and a card that has left the graveyard by then is a new
// object the trigger cannot find (CR 400.7).
//
// No simplification.
const pain101Return = "pain-101/return"

func init() {
	Register(Spec{
		OracleID:     "e9975888-c75f-471e-a70d-17449e92226d",
		Name:         "Pain 101",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Grants: []AbilityGrant{{
			Key: pain101Return,
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Pain 101 — return it to the battlefield tapped",
					returnThisCreatureFromGraveyard(nil, nil)),
			},
			Text: "When this creature dies, return it to the battlefield tapped under its owner's control.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return GrantAbilitiesFor{
				Target: target,
				Keys:   []string{pain101Return},
				Also:   []game.Mod{game.AddKeywordsMod("deathtouch")},
				Label:  "Pain 101 — deathtouch, and it returns when it dies",
			}.Apply(ctx)
		},
	})
}
