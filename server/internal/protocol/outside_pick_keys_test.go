package protocol

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// outside_pick_keys_test.go — #4361 S4 deleted the closure form of an
// outside-the-game pick (OutsidePickPrompt.Match), so the view tests ask
// through a registered sorcery pick that does nothing after the pick.
var testProtocolSorceryPick = game.RegisterOutsidePick("test/4361-protocol-sorcery",
	func(c game.Card, _ string) bool { return c.IsSorcery() },
	func(*game.Game, game.CardsChosen) error { return nil })

// testProtocolDistinctPick is a pick "with different names" (Turtles
// Forever's outside half, #4947), and testProtocolRepeatPick one whose
// names may repeat (Research's "up to four cards you own").
var testProtocolDistinctPick = game.RegisterOutsidePickDistinctNames("test/4947-protocol-distinct",
	nil, nil, func(*game.Game, game.CardsChosen) error { return nil })

var testProtocolRepeatPick = game.RegisterOutsidePick("test/4947-protocol-repeat",
	nil, func(*game.Game, game.CardsChosen) error { return nil })
