package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// library_pick_validate_test.go — #998. A library-top pick can now
// carry a rule about the picked SET, not just about each card:
// PutFromLibraryOntoBattlefield.Validate and
// TakeFromLibraryToHand.Validate forward
// game.ChooseCardsPrompt.Validate, so "any number of nonland permanent
// cards with TOTAL MANA VALUE 4 OR LESS from among them" (Ao, the Dawn
// Sky) is expressible over a look at the top seven.
//
// The rule itself is enforced in exactly one place — the engine's
// checkChooseCardsPicksLocked, which the submit path and internal/legal
// both go through — so these tests are about the WIRE: that the field
// reaches the prompt, that a refused answer leaves the prompt open, and
// that the forced-answer shortcut steps aside for a set rule. The bot's
// half is in legal/choose_cards_library_validate_test.go.

// testPutBudget4 is the put's total-mana-value-4 set rule, registered
// (#4253: the put's pick is keyed, its set rule a registered rule).
var TestPutBudget4 = effects.RegisterPutRule("test/4253-put-total-mana-value-4", effects.B23TotalManaValueAtMost(4))

// testBudgetTake is the take-to-hand set rule below as a registered take
// (#4253 retired the unkeyed take, whose pick held a closure).
var TestBudgetTake = effects.RegisterTake("test/4253-take-total-mana-value-4", effects.TakeDef{
	Validate: effects.B23TotalManaValueAtMost(4),
	Then:     effects.TakeRestOnBottomInRandomOrder,
})
