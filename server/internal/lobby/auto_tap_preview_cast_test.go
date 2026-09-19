package lobby

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_cast_test.go — #696. The preview PRICES THE
// ANNOUNCEMENT, and every test here asks the same claim once: a
// preview that says "ok" is a cast CastSpell{AutoTap: true} accepts
// on the same board, and one that says "not ok" is a cast it refuses.
//
// The endpoint used to parse `card.ManaCost` and apply the commander
// tax and the cost modifiers itself, so it knew nothing about the
// alternative cost claimed at announce, a granted permission's flat
// override, the "spend mana as though any colour" fold, the mana half
// of an optional additional cost or the face being cast. Each of
// those was a disagreement with the engine in one direction or the
// other: a disabled "Auto-tap & cast" on a cast that would have gone
// through, or a plan for one that would fail.
//
// The fixture is previewFixture in auto_tap_preview_test.go.

// previewWith asks the endpoint about an announced cast — the same
// query string the client now builds from the cast payload.
func (f *previewFixture) previewWith(card uuid.UUID, query string) previewBody {
	f.t.Helper()
	path := "/games/" + f.gameID.String() + "/auto-tap-preview?card=" + card.String() + query
	resp := f.get(path)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		f.t.Fatalf("GET %s: status %d", path, resp.StatusCode)
	}
	var body previewBody
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		f.t.Fatalf("decode preview: %v", err)
	}
	return body
}

// game reaches the engine behind the lobby entry.
func (f *previewFixture) game() *game.Game {
	f.t.Helper()
	g, err := f.lobby.LookupGame(f.gameID)
	if err != nil {
		f.t.Fatalf("LookupGame: %v", err)
	}
	return g
}

// castsWithAutoTap runs the cast the preview was asked about and
// reports whether the engine took it. Destructive, so it is the last
// thing a test does to a board.
func (f *previewFixture) castsWithAutoTap(card uuid.UUID, params game.CastSpellParams) bool {
	f.t.Helper()
	params.Strict = true
	params.AutoTap = true
	return f.game().CastSpell(f.alice, card, params) == nil
}

// grant hands Alice a cast permission over one card.
func (f *previewFixture) grant(card uuid.UUID, perm game.CastPermission) {
	f.t.Helper()
	perm.Player = f.alice
	if !f.game().GrantCastPermissionOverCardForEffect(card, perm) {
		f.t.Fatalf("the grant did not attach to %s", card)
	}
}

// mainPhase walks the turn to a main phase so a sorcery-speed cast is
// legal. The preview does not care; the CastSpell half does.
func (f *previewFixture) mainPhase() {
	f.t.Helper()
	g := f.game()
	for i := 0; i < 20; i++ {
		if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			f.t.Fatalf("AdvanceStep: %v", err)
		}
	}
	f.t.Fatalf("never reached a main phase")
}

// A granted exile cast with a {0} override: one Mountain is plenty,
// and the printed {5}{R}{R} the endpoint used to price is not.
func TestAutoTapPreviewPricesAGrantedExileCastAtTheOverride(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	card := f.spawn(f.alice, game.ZoneExile, game.Card{
		Name: "Cascade Hit", TypeLine: "Sorcery", ManaCost: "{5}{R}{R}",
	}, 1)[0]
	f.grant(card, game.CastPermission{CastOnly: true, Duration: game.WhileInZoneDuration(), Cost: "{0}"})
	f.mountains(1)

	// Priced as a hand cast — no from_zone — the override is invisible
	// and the preview reports the printed cost's shortfall. That is
	// exactly the reported bug, and the same board takes the cast the
	// moment the announcement is passed through.
	if asHand := f.previewWith(card, ""); asHand.OK {
		t.Errorf("priced without from_zone the preview was ok; the grant is what makes this castable")
	}

	got := f.previewWith(card, "&from_zone=exile")
	if !got.OK {
		t.Fatalf("granted exile cast: ok=%v missing=%v cost=%q, want ok at the {0} override", got.OK, got.Missing, got.Cost)
	}
	if got.Cost != "{0}" {
		t.Errorf("reported cost = %q, want the override {0} rather than the printed cost", got.Cost)
	}
	if !f.castsWithAutoTap(card, game.CastSpellParams{FromZone: "exile"}) {
		t.Errorf("the preview said ok and CastSpell refused the same cast")
	}
}

// A flashback cost HIGHER than the printed cost. The old preview
// called this affordable and handed the player a cast that failed.
func TestAutoTapPreviewPricesAFlashbackCostAboveThePrintedOne(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	const oracle = "test-preview-flashback"
	prevZones, prevAlts := game.CatalogCastableZones, game.CatalogAlternativeCosts
	game.CatalogCastableZones = func(id string) []game.ZoneKind {
		if id == oracle {
			return []game.ZoneKind{game.ZoneGraveyard}
		}
		return prevZones(id)
	}
	game.CatalogAlternativeCosts = func(id string) []game.AlternativeCost {
		if id == oracle {
			return []game.AlternativeCost{{
				Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}",
				FromZone: game.ZoneGraveyard,
			}}
		}
		return prevAlts(id)
	}
	t.Cleanup(func() {
		game.CatalogCastableZones, game.CatalogAlternativeCosts = prevZones, prevAlts
	})

	card := f.spawn(f.alice, game.ZoneGraveyard, game.Card{
		Name: "Test Think Twice", TypeLine: "Instant", ManaCost: "{R}", OracleID: oracle,
	}, 1)[0]
	f.mountains(2)

	// Two Mountains pay the printed {R} and not the flashback {2}{R}.
	if got := f.previewWith(card, "&from_zone=graveyard&alternative_cost=flashback"); got.OK {
		t.Errorf("two Mountains against a {2}{R} flashback: preview ok with plan %v, want a miss", got.Plan)
	}
	f.mountains(1)
	got := f.previewWith(card, "&from_zone=graveyard&alternative_cost=flashback")
	if !got.OK {
		t.Fatalf("three Mountains against a {2}{R} flashback: ok=%v missing=%v", got.OK, got.Missing)
	}
	if got.Cost != "{2}{R}" {
		t.Errorf("reported cost = %q, want the flashback cost being paid", got.Cost)
	}
	if !f.castsWithAutoTap(card, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}) {
		t.Errorf("the preview said ok and CastSpell refused the same cast")
	}
}

// "You may spend mana as though it were mana of any color" (Breeches):
// two Mountains pay a {G}{G} card under the grant, and the printed
// pips the old preview reported would have read as unpayable.
func TestAutoTapPreviewFoldsAnAnyColorGrant(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	card := f.spawn(f.alice, game.ZoneExile, game.Card{
		Name: "Impulsed Bear", TypeLine: "Creature — Bear", ManaCost: "{G}{G}",
	}, 1)[0]
	f.grant(card, game.CastPermission{CastOnly: true, Duration: game.WhileInZoneDuration(), AnyColor: true})
	f.mountains(2)

	got := f.previewWith(card, "&from_zone=exile")
	if !got.OK {
		t.Fatalf("two Mountains against {G}{G} under an any-colour grant: ok=%v missing=%v", got.OK, got.Missing)
	}
	if !f.castsWithAutoTap(card, game.CastSpellParams{FromZone: "exile"}) {
		t.Errorf("the preview said ok and CastSpell refused the same cast")
	}
}

// The commander tax, which the endpoint DID know about — asserted so
// that routing through the engine's pricer did not lose it, and so
// the from_zone the client now sends is honoured rather than
// re-derived by scanning the command zone.
func TestAutoTapPreviewChargesTheCommanderTax(t *testing.T) {
	f := newPreviewFixture(t)
	f.mainPhase()
	card := f.spawn(f.alice, game.ZoneCommand, game.Card{
		Name: "Test Commander", TypeLine: "Legendary Creature — Bear", ManaCost: "{1}{R}",
	}, 1)[0]
	seat := f.game().PlayerByIDForEffect(f.alice)
	seat.CommanderCasts[card] = 1
	f.mountains(2)

	if got := f.previewWith(card, "&from_zone=command"); got.OK {
		t.Errorf("two Mountains against a taxed {1}{R}: preview ok with plan %v, want a miss ({3}{R})", got.Plan)
	}
	f.mountains(2)
	if got := f.previewWith(card, "&from_zone=command"); !got.OK {
		t.Fatalf("four Mountains against a taxed {1}{R}: ok=%v missing=%v", got.OK, got.Missing)
	}
	if !f.castsWithAutoTap(card, game.CastSpellParams{FromZone: "command"}) {
		t.Errorf("the preview said ok and CastSpell refused the same cast")
	}
}

// A malformed announcement is a 400 rather than a silent default: a
// preview that quietly priced a different cast from the one the
// button will send is worse than no preview.
func TestAutoTapPreviewRefusesAMalformedAnnouncement(t *testing.T) {
	f := newPreviewFixture(t)
	card := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}",
	}, 1)[0]
	for _, q := range []string{
		"&alternative_cost=overload",
		"&from_zone=nowhere",
		"&face=-1",
		"&optional_costs=kicker",
		"&tap_ids=not-a-uuid",
	} {
		path := "/games/" + f.gameID.String() + "/auto-tap-preview?card=" + card.String() + q
		resp := f.get(path)
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("GET %s: status %d, want 400", path, resp.StatusCode)
		}
		resp.Body.Close()
	}
}
