package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mass of Mysteries — Legendary Creature — Elemental {W}{U}{B}{R}{G}, 5/5
// (Lorwyn Eclipsed Commander):
//
//	"First strike, vigilance, trample
//	 At the beginning of combat on your turn, another target Elemental you
//	 control gains myriad until end of turn. (...)"
//
// The target is chosen as the trigger goes on the stack and re-checked as it
// resolves (CR 608.2b). Myriad is the shared trigger (myriad.go) granted to
// it in layer 6 until end of turn (ADR 0093), so it is that creature's own
// ability and fires when it is declared as an attacker.
//
// No simplification.
const massOfMysteriesGrant = "mass-of-mysteries/myriad"

func init() {
	Register(Spec{
		OracleID:        "3ceffadd-c32d-4508-b307-e03bf3c8e7b8",
		Name:            "Mass of Mysteries",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike", "vigilance", "trample"},
		Grants: []AbilityGrant{{
			Key:       massOfMysteriesGrant,
			Triggered: []game.TriggeredAbility{GrantedMyriad("Mass of Mysteries")},
			Text:      "Myriad",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(AtBeginningOfYourCombat("Mass of Mysteries — another target Elemental you control gains myriad until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := ctx.ClauseTarget(0)
					if !ok {
						return nil
					}
					return GrantAbilitiesFor{
						Target: t.ID,
						Keys:   []string{massOfMysteriesGrant},
						Label:  "Mass of Mysteries — gains myriad until end of turn",
					}.Apply(ctx)
				}),
				Another(TargetCreature("another target Elemental you control", Subtype("Elemental"), YouControl()))),
		},
	})
}
