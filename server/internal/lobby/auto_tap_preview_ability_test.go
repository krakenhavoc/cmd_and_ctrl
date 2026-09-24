package lobby

import (
	"net/http"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_ability_test.go — #1405. The ?ability= branch of
// the auto-tap preview prices an activation exactly as
// ActivateCatalogAbility charges it: the board's activation modifiers
// (#1184), the ability's own clause (#1296) and, when the request
// names them, the targets. It used to parse the printed cost and
// nothing else, so it disagreed with the payment whenever a modifier
// reached the ability.
//
// Each test asks the preview, then runs the activation it previewed on
// the same board (AutoTap + Strict), and holds the two to one answer:
// "ok" is an activation the engine takes, tapping exactly the plan;
// "not ok" is one it refuses.
//
// The fixture is previewFixture in auto_tap_preview_test.go.

const (
	boomScholarOracle     = "48296cc0-0141-47c6-9ec8-ada6171fee6f"
	dragonfireBladeOracle = "e809b847-b712-4558-95ea-9bb7356cde91"
)

// previewAbility asks about ability 0 of `card`, with `query` appended
// (a "&targets=…" clause, or empty).
func (f *previewFixture) previewAbility(card uuid.UUID, query string) previewBody {
	f.t.Helper()
	return f.previewWith(card, "&ability=0"+query)
}

// lands puts n untapped basics of one type on Alice's battlefield.
func (f *previewFixture) lands(name string, n int) []uuid.UUID {
	f.t.Helper()
	return f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: name, TypeLine: "Basic Land — " + name}, n)
}

// activate runs ability 0 of `card` for Alice with the payment
// enforced and auto-tapped — the activation the preview described.
func (f *previewFixture) activate(card uuid.UUID, targets []game.TargetRef) error {
	f.t.Helper()
	return f.game().ActivateCatalogAbility(f.alice, card, 0, game.ActivateAbilityParams{
		Targets: targets, Strict: true, AutoTap: true,
	})
}

// tappedAmong returns which of `ids` are tapped now, sorted.
func (f *previewFixture) tappedAmong(ids []uuid.UUID) []string {
	f.t.Helper()
	g := f.game()
	var out []string
	for _, id := range ids {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			f.t.Fatalf("card %s left the game", id)
		}
		if c.Tapped {
			out = append(out, id.String())
		}
	}
	sort.Strings(out)
	return out
}

func sortedCopy(ids []string) []string {
	out := append([]string(nil), ids...)
	sort.Strings(out)
	return out
}

// mainPhaseForAlice walks the game to a main phase of Alice's own turn
// with an empty stack, so a sorcery-speed ability (equip) is legal.
func (f *previewFixture) mainPhaseForAlice() {
	f.t.Helper()
	g := f.game()
	for i := 0; i < 40; i++ {
		active := g.Seats[g.Turn.ActiveSeat].ID
		if active == f.alice && (g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain) {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			f.t.Fatalf("AdvanceStep: %v", err)
		}
	}
	f.t.Fatalf("never reached a main phase of Alice's turn")
}

// Two Boom Scholars: each discounts the OTHER's exhaust ability by
// {2} ("Exhaust abilities of other permanents you control cost {2}
// less to activate"), so {4}{R}{G} costs {2}{R}{G}. Four lands pay it.
// The old preview priced the printed six and said no; the activation
// on the same board goes through.
func TestAutoTapPreviewAbilityAppliesABoardDiscount(t *testing.T) {
	f := newPreviewFixture(t)
	scholar := game.Card{Name: "Boom Scholar", TypeLine: "Creature — Human Warrior", OracleID: boomScholarOracle}
	ids := f.spawn(f.alice, game.ZoneBattlefield, scholar, 2)
	lands := append(f.lands("Mountain", 2), f.lands("Forest", 1)...)

	// Three lands for {2}{R}{G}: a miss, and the missing list is the
	// discounted cost's four symbols, not the printed six.
	short := f.previewAbility(ids[0], "")
	if short.OK {
		t.Fatalf("three lands for {2}{R}{G}: preview ok with plan %v, want a miss", short.Plan)
	}
	if len(short.Missing) != 4 {
		t.Errorf("three lands for {2}{R}{G}: missing %v, want the four symbols of the discounted cost", short.Missing)
	}
	if err := f.activate(ids[0], nil); err == nil {
		t.Fatalf("the preview said no and the activation went through")
	}

	lands = append(lands, f.lands("Mountain", 1)...)
	got := f.previewAbility(ids[0], "")
	if !got.OK || len(got.Plan) != 4 {
		t.Fatalf("four lands for {2}{R}{G}: ok=%v plan=%v missing=%v, want ok with four taps", got.OK, got.Plan, got.Missing)
	}
	if got.Cost != "{4}{R}{G}" {
		t.Errorf("cost = %q, want the printed string (the modifiers ride plan / missing)", got.Cost)
	}
	if err := f.activate(ids[0], nil); err != nil {
		t.Fatalf("the preview said ok and the activation was refused: %v", err)
	}
	if tapped, want := f.tappedAmong(lands), sortedCopy(got.Plan); !equalStrings(tapped, want) {
		t.Errorf("the activation tapped %v, the preview planned %v", tapped, want)
	}
}

// Dragonfire Blade's equip is {4}, "{1} less to activate for each
// color of the creature it targets" — the ability's own clause, read
// off the target. With no target in the request the preview prices the
// no-target cost ({4}); with Vivi Ornitier ({U}{R}) it prices {2}; with
// a colourless creature, {4} again. The activation charges the same.
func TestAutoTapPreviewAbilityPricesDragonfireBladeByTarget(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhaseForAlice()
	blade := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Dragonfire Blade", TypeLine: "Artifact — Equipment", OracleID: dragonfireBladeOracle,
	}, 1)[0]
	vivi := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Vivi Ornitier", TypeLine: "Legendary Creature — Wizard", Colors: []string{"U", "R"}, Power: 0, Toughness: 3,
	}, 1)[0]
	golem := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2,
	}, 1)[0]
	lands := f.lands("Mountain", 2)

	if none := f.previewAbility(blade, ""); none.OK || len(none.Missing) != 4 {
		t.Errorf("no targets: ok=%v missing=%v, want the no-target price {4} (four missing)", none.OK, none.Missing)
	}
	if onGolem := f.previewAbility(blade, "&targets=card:"+golem.String()); onGolem.OK || len(onGolem.Missing) != 4 {
		t.Errorf("targeting a colourless creature: ok=%v missing=%v, want the full {4}", onGolem.OK, onGolem.Missing)
	}
	if err := f.activate(blade, []game.TargetRef{{Kind: game.TargetCard, ID: golem}}); err == nil {
		t.Fatalf("the preview said no to the Golem and the activation went through")
	}

	onVivi := f.previewAbility(blade, "&targets=card:"+vivi.String())
	if !onVivi.OK || len(onVivi.Plan) != 2 {
		t.Fatalf("targeting a two-colour creature: ok=%v plan=%v missing=%v, want ok with two taps ({2})", onVivi.OK, onVivi.Plan, onVivi.Missing)
	}
	if err := f.activate(blade, []game.TargetRef{{Kind: game.TargetCard, ID: vivi}}); err != nil {
		t.Fatalf("the preview said ok for Vivi and the activation was refused: %v", err)
	}
	if tapped, want := f.tappedAmong(lands), sortedCopy(onVivi.Plan); !equalStrings(tapped, want) {
		t.Errorf("the activation tapped %v, the preview planned %v", tapped, want)
	}
}

// A malformed ?targets= is a 400, not a silent no-target price: a
// preview that quietly answered a different question from the one
// asked is worse than none.
func TestAutoTapPreviewAbilityRejectsMalformedTargets(t *testing.T) {
	f := newPreviewFixture(t)
	blade := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Dragonfire Blade", TypeLine: "Artifact — Equipment", OracleID: dragonfireBladeOracle,
	}, 1)[0]
	for _, q := range []string{"targets=" + blade.String(), "targets=stack:" + blade.String(), "targets=card:nope"} {
		resp := f.get("/games/" + f.gameID.String() + "/auto-tap-preview?card=" + blade.String() + "&ability=0&" + q)
		resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("?%s: status %d, want 400", q, resp.StatusCode)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
