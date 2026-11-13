package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Auriok Champion — Creature — Human Cleric {W}{W}, 1/1 (Fifth Dawn,
// #5604):
//
//	"Protection from black and from red
//	 Whenever another creature enters, you may gain 1 life."
//
// Paladin en-Vec's two protections and Soul's Attendant's optional
// trigger, which watches every other creature entering, under any
// player's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "633fd158-19d0-4d97-a076-2ad897bbed19",
		Name:         "Auriok Champion",
		Completeness: CompletenessFull,
		PrintedKeywords: []string{
			game.ProtectionFromColor("B"), game.ProtectionFromColor("R"),
		},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventETB, AnotherCreatureEntered, "Auriok Champion — gain 1 life", Do(GainLife{Amount: 1})), "Auriok Champion — gain 1 life?"),
		},
	})
}
