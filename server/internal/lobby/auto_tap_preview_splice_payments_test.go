package lobby

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_splice_payments_test.go — #4159. A splice card's cost
// that is not mana (#4055) names permanents the cast spends, and the cast
// keeps them out of the mana plan (CastAutoTapExclusions). The preview
// takes the same `splice_payments`, so the line a player reads never
// names a source the cast will not use.

const hundredTalonStrikeOracle = "fbfaf009-8276-4687-a3db-5a71dc13f174"

// Two Mountains pay a {1} Arcane spell. A splice cost naming one of them
// leaves the plan the other, whichever one it names.
func TestAutoTapPreviewKeepsWhatASpliceCostSpendsOutOfThePlan(t *testing.T) {
	f := newPreviewFixture(t)
	spell := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Arcane Probe", TypeLine: "Instant — Arcane", ManaCost: "{1}",
	}, 1)[0]
	splice := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Hundred-Talon Strike", TypeLine: "Instant — Arcane", ManaCost: "{W}",
		OracleID: hundredTalonStrikeOracle,
	}, 1)[0]
	lands := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Mountain", TypeLine: "Basic Land — Mountain",
	}, 2)

	for i, spent := range lands {
		other := lands[1-i]
		q := fmt.Sprintf("&splices=%s&splice_payments=%s", splice,
			url.QueryEscape(fmt.Sprintf(`[{"tap_ids":["%s"]}]`, spent)))
		got := f.previewWith(spell, q)
		if !got.OK || len(got.Plan) != 1 || got.Plan[0] != other.String() {
			t.Errorf("splice cost taps %s: plan = %v ok=%v, want only %s", spent, got.Plan, got.OK, other)
		}
	}
}

// A malformed splice_payments is a 400, like every other malformed id.
func TestAutoTapPreviewRefusesMalformedSplicePayments(t *testing.T) {
	f := newPreviewFixture(t)
	spell := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Arcane Probe", TypeLine: "Instant — Arcane", ManaCost: "{1}",
	}, 1)[0]
	for _, bad := range []string{"nope", `[{"tap_ids":["x"]}]`} {
		resp := f.get("/games/" + f.gameID.String() + "/auto-tap-preview?card=" + spell.String() +
			"&splice_payments=" + url.QueryEscape(bad))
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("splice_payments=%q: status %d, want 400", bad, resp.StatusCode)
		}
	}
}
