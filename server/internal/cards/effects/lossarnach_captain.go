package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lossarnach Captain — Creature — Human Soldier {3}{W}, 3/1 (EDHREC
// rank 3798):
//
//	"First strike
//	 Whenever this creature or another Human you control enters, tap
//	 target creature an opponent controls.
//	 At the beginning of your upkeep, create a 1/1 white Human Soldier
//	 creature token."
//
// The Human-typal tapper. First strike rides PrintedKeywords. The tap
// is one printed ability with two conditions — the Captain's own
// entry, or another Human (token or not) entering under his
// controller's control — targeted at a creature an opponent
// controls, so it is removed when no opponent has a creature (CR
// 603.3d). The upkeep token is a Human, so it fires the tap trigger
// as it enters, which is the printed loop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1251691e-3c4c-4e45-9f59-f156852f2268",
		Name:            "Lossarnach Captain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			{
				Watches: []game.EventKind{game.EventETB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b36SelfOrAnotherHumanYouControlEntered(ev, source, g)
				},
				Targets: TargetCreature("target creature an opponent controls", OpponentControls()),
				Key:     "Lossarnach Captain — tap target creature an opponent controls",
				Effect:  b36TapChosenCreature,
			},
			AtYourUpkeep("Lossarnach Captain — create a 1/1 white Human Soldier", b34CreateTokens(b36WhiteHumanSoldierToken, 1)),
		},
	})
}
