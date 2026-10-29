package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// choose_cards_hook_s7b_test.go — #4361 S7b deleted the closure forms of the
// card-set pick (ChooseCardsPrompt.Then and .Validate). This package's tests
// ask through the keyed form and hear the answer through chooseCardsHooked.

// chooseCardsHooked is p with fn as its keyed continuation, registered under
// a fresh test key and handed the picks (nil when nobody chose), and
// validate, when non-nil, as the set rule registered with it. A Zone the
// test set is the pick's live-zone re-check, as it was for the closure pick.
// A test's continuation may capture: a test process keeps the registry for
// its whole run.
func chooseCardsHooked(p game.ChooseCardsPrompt, fn func(g *game.Game, picked []uuid.UUID) error, validate func([]game.Card) bool) game.ChooseCardsPrompt {
	key := "test/4361-s7b-pick-" + uuid.NewString()
	body := func(g *game.Game, r game.CardsChosen) error {
		if fn == nil {
			return nil
		}
		return fn(g, r.Picked)
	}
	if validate != nil {
		p.ThenKey = game.RegisterChooseCardsThenRule(key, validate, body)
	} else {
		p.ThenKey = game.RegisterChooseCardsThen(key, body)
	}
	if p.LiveZone == "" {
		p.LiveZone = p.Zone
	}
	return p
}
