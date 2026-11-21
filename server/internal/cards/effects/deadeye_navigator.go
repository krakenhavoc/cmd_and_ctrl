package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deadeye Navigator — Creature — Spirit {4}{U}{U}, 5/5 (Innistrad
// Remastered, S82 GAP, #5948):
//
//	"Soulbond (You may pair this creature with another unpaired creature
//	 when either enters. They remain paired for as long as you control both
//	 of them.)
//	 As long as Deadeye Navigator is paired with another creature, each of
//	 those creatures has "{1}{U}: Exile this creature, then return it to the
//	 battlefield under your control.""
//
// Soulbond is the keyword's two triggers (CR 702.95a, soulbond.go). The
// second line is a layer-6 grant of an activated ability to both creatures
// of the pair (ADR 0093), so each has its own instance: activating it
// blinks the creature it is on (Aethergeode Miner's blink, returning under
// the activator's control), and the returning creature is a new object
// (CR 400.7) that is unpaired, ending the pair (CR 702.95e). A creature
// that left and came back before the ability resolves is not blinked.
//
// No simplification.
const deadeyeNavigatorGrant = "deadeye-navigator/blink"

func init() {
	Register(Spec{
		OracleID:        "ebef6d84-b724-4a84-bcde-8eb6dc06717d",
		Name:            "Deadeye Navigator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordSoulbond},
		Triggered:       Soulbond("Deadeye Navigator"),
		Grants: []AbilityGrant{{
			Key: deadeyeNavigatorGrant,
			Activated: []ActivatedAbility{{
				Label:   "{1}{U}: Exile this creature, then return it to the battlefield under your control.",
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Cost:    ManaCost("{1}{U}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					return Flicker{Target: item.SourceCardID, Controller: item.Controller}.Apply(NewContext(g, item))
				},
			}},
			Text: "{1}{U}: Exile this creature, then return it to the battlefield under your control.",
		}},
		Static: []game.StaticAbility{
			PairedCreaturesHave("As long as Deadeye Navigator is paired with another creature, each of those creatures has \"{1}{U}: Exile this creature, then return it to the battlefield under your control.\"",
				deadeyeNavigatorGrant),
		},
	})
}
