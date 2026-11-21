package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// activation_restriction_zone_test.go — #5759: "activated abilities of
// artifacts / creatures can't be activated" names a card type and no
// zone, so it means permanents on the battlefield (CR 109.2). An
// artifact or creature CARD in a hand or a graveyard is not one, and
// its cycling, ninjutsu or graveyard ability still works.
//
// Each test asserts the battlefield refusal on the same board first,
// so a restriction that stopped applying at all would fail here too
// rather than pass every hand and graveyard assertion.

// actOffersActivationFrom reports whether the enumerator offers `seat`
// an activation of `source`, the bot's half of the agreement (#544).
func ActOffersActivationFrom(g *game.Game, seat, source uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == "activate_ability" && m.Source == source {
			return true
		}
	}
	return false
}
