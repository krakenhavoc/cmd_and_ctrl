package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aerid Konstrari — Legendary Creature — Elder Sphinx {1}{R}{G}{G}, 5/4:
//
//	"Flying
//	 When Aerid Konstrari enters or dies, create a Heartwood token.
//	 (It's a red and green artifact with "{T}: Add {R} or {G}.")
//	 {6}: Create a Heartwood token. Then Aerid Konstrari gets +X/+0
//	 until end of turn, where X is the number of artifacts you
//	 control."
//
// X is counted after the token exists, so the new Heartwood is one of
// the artifacts, as the "Then" orders it. The pump is skipped if Aerid
// has left the battlefield by the time the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "172a8f58-c420-4497-b962-c5853f923667",
		Name:            "Aerid Konstrari",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Aerid Konstrari — create a Heartwood token",
				Do(CreateToken{Template: HeartwoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{6}: Create a Heartwood token. Then Aerid Konstrari gets +X/+0 until end of turn, where X is the number of artifacts you control",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{6}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (CreateToken{Template: HeartwoodToken(), N: 1}).Apply(ctx); err != nil {
					return err
				}
				if !b15OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return BoostUntilEOT{
					Target: item.SourceCardID,
					Power:  rfCreatureAArtifactsYouControl(g, item.Controller),
					Label:  "Aerid Konstrari — +X/+0",
				}.Apply(ctx)
			},
		}},
	})
}
