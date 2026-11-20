package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Scuttling Sliver — Creature — Sliver Trilobite {2}{U}, 2/2 (Modern Horizons, #5519):
//
//	"Sliver creatures you control have "{2}: Untap this creature.""
//
// A granted activated ability (ADR 0093) on your own Slivers: Blizzard
// Elemental's untap row at {2}. "This creature" is the Sliver that has
// the ability, not the Scuttling Sliver.
//
// No simplification.
const scuttlingSliverGrant = "scuttling-sliver/untap"

func init() {
	Register(Spec{
		OracleID:     "7c3dcfea-f42c-4515-9b7d-6e9d8cb4d7f9",
		Name:         "Scuttling Sliver",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: scuttlingSliverGrant,
			Activated: []ActivatedAbility{{
				UntapSelf: true,
				Label:     "{2}: Untap this creature.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      ManaCost("{2}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return UntapTarget{Target: item.SourceCardID}.Apply(NewContext(g, item))
				},
			}},
			Text: "{2}: Untap this creature.",
		}},
		Static: []game.StaticAbility{
			TribalAbilityGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, scuttlingSliverGrant),
		},
	})
}
