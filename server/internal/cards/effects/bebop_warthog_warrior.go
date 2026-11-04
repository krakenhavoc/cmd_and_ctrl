package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bebop, Warthog Warrior — {4}{B} Legendary Creature — Boar Mutant Warrior
// 5/4 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Menace
//	 Rhinos you control have menace.
//	 Swampcycling {2}"
//
// Menace is printed (CR 702.111a). The grant is Warg Rider's tribal keyword
// grant, narrowed to Rhinos you control; Bebop is a Boar, so it reaches
// Bebop only if something makes it a Rhino. Swampcycling is Troll of
// Khazad-dûm's typecycling (CR 702.29e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39c813e1-0e26-440a-b9bd-e781f7576b8b",
		Name:            "Bebop, Warthog Warrior",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Rhino"}, YoursOnly: true}, "menace"),
		},
		Activated: []ActivatedAbility{
			Typecycling("Swampcycling", "{2}", "a Swamp card", IsLandWithSubtype("Swamp")),
		},
	})
}
