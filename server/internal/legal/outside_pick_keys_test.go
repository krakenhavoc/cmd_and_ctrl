package legal_test

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// outside_pick_keys_test.go — #4361 S4 deleted the closure form of an
// outside-the-game pick (OutsidePickPrompt.Match and Validate), so these
// tests ask through registered picks: a sorcery filter, and Extrapolate
// the Impossible's "exactly two cards with different names". Neither does
// anything after the pick; the tests read the enumerated moves.
var (
	testLegalSorceryPick = game.RegisterOutsidePick("test/4361-legal-sorcery",
		func(c game.Card, _ string) bool { return c.IsSorcery() },
		func(*game.Game, game.CardsChosen) error { return nil })
	testLegalTwoNamesPick = game.RegisterOutsidePickRule("test/4361-legal-two-names", nil,
		func(picked []game.Card) bool { return len(picked) == 2 && picked[0].Name != picked[1].Name },
		func(*game.Game, game.CardsChosen) error { return nil })
)
