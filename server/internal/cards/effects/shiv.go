package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shiv — Plane — Dominaria (Planechase Anthology, #4870):
//
//	"All creatures have "{R}: This creature gets +1/+0 until end of turn."
//	 Whenever chaos ensues, create a 5/5 red Dragon creature token with
//	 flying."
//
// A plane (ADR 0216 §6). The static is a layer-6 grant of a catalog bundle
// (ADR 0093) to every creature, each creature's controller activating its
// own ability; the Dragon is the planar controller's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "70edae19-e278-401d-9772-508bd66320a6",
		Name:         "Shiv",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: shivPump,
			Activated: []ActivatedAbility{{
				Label:   "{R}: This creature gets +1/+0 until end of turn.",
				Cost:    ManaCost("{R}"),
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{Target: item.SourceCardID, Power: 1, Label: "Shiv — +1/+0"}.Apply(NewContext(g, item))
				},
			}},
			Text: "{R}: This creature gets +1/+0 until end of turn.",
		}},
		Static: []game.StaticAbility{GrantAbilities(allCreatures, shivPump)},
		Triggered: []game.TriggeredAbility{
			WheneverChaosEnsues("Shiv — create a 5/5 red Dragon creature token with flying",
				Do(CreateToken{Template: TokenCard("5/5 red Dragon with flying"), N: 1})),
		},
	})
}

// shivPump is the granted bundle's key. An on-disk key.
const shivPump = "shiv/red-pump"
