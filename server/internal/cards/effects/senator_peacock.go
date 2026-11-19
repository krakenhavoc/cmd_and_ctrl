package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const senatorPeacockClueGrant = "senator-peacock/clue-draw" // an on-disk identity: never renamed

// Senator Peacock — Legendary Creature — Human Advisor {3}{U}{U}, 3/4
// (Ravnica: Clue Edition):
//
//	"Artifacts you control are Clues in addition to their other types and
//	 have "{2}, Sacrifice this artifact: Draw a card."
//	 Whenever you sacrifice a Clue, target creature can't be blocked this
//	 turn."
//
// A layer-4 subtype addition (CR 205.1b) and a layer-6 grant of the Clue
// ability, both over the artifacts you control as a continuous effect, so an
// artifact that enters later is one too. The trigger sees any Clue you
// sacrifice, for any reason, as the sacrifice happens (the Lonis shape);
// the target is chosen as the trigger goes on the stack and re-checked as
// it resolves (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:     "10d8f29a-d483-4f06-bc13-d41d62847ffe",
		Name:         "Senator Peacock",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: senatorPeacockClueGrant,
			Activated: []ActivatedAbility{{
				Label:   "{2}, Sacrifice this artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Effect:  Do(DrawCards{N: 1}),
			}},
			Text: "{2}, Sacrifice this artifact: Draw a card.",
		}},
		Static: []game.StaticAbility{
			bnsB22AreAlsoClues("Artifacts you control are Clues in addition to their other types.", bnsB22ArtifactsYouControl),
			Labelled(GrantAbilities(bnsB22ArtifactsYouControl, senatorPeacockClueGrant),
				"Artifacts you control have \"{2}, Sacrifice this artifact: Draw a card.\""),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Senator Peacock — target creature can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return RestrictUntilEOT{
						Target:       t.ID,
						Restrictions: game.CantBeBlocked,
						Label:        "Senator Peacock — can't be blocked",
					}.Apply(ctx)
				}),
				TargetCreature("target creature")),
		},
	})
}
