package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elephant-Mandrill — Creature — Elephant Monkey {2}{G}, 3/2 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"Reach
//	 When this creature enters, each player creates a Food token.
//	 At the beginning of combat on your turn, this creature gets +1/+1
//	 until end of turn for each artifact your opponents control."
//
// The enters trigger is Party Dude's (avB30cEachPlayerCreatesAFood). The
// combat trigger counts the artifacts your opponents control once, as it
// resolves (CR 608.2h), and that bonus lasts the turn: an artifact that
// arrives or leaves later changes nothing. A Mandrill that has left the
// battlefield gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8695c7e2-464e-464c-9d3f-a7cf025f3eb2",
		Name:            "Elephant-Mandrill",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Elephant-Mandrill — each player creates a Food token", avB30cEachPlayerCreatesAFood),
			AtBeginningOfYourCombat("Elephant-Mandrill — +1/+1 for each artifact your opponents control", elephantMandrillGrows),
		},
	})
}

// elephantMandrillGrows is the combat trigger's resolution.
func elephantMandrillGrows(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	n := hobB7ArtifactsByOpponents(g, item.Controller)
	return BoostUntilEOT{Target: item.SourceCardID, Power: n, Toughness: n,
		Label: "Elephant-Mandrill — +1/+1 for each artifact your opponents control"}.Apply(NewContext(g, item))
}
