package lobby

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// auto_tap_preview_sources_test.go — #1285. The preview's `plan` is a
// bare ID list the client looked up on the battlefield, so a planned
// source that is a card in HAND — a Spirit Guide, #1228 — was simply
// not shown: the preview said "ok" and listed fewer sources than the
// payment would spend, and the one it hid is the one that does not
// come back. `sources` describes each entry: where it is, and whether
// paying with it taps, sacrifices or exiles it.

const simianSpiritGuideOracle = "44e0ffa3-8915-4c1f-8f1a-4aeea1365f07"

type previewSource struct {
	CardID    string `json:"card_id"`
	Name      string `json:"name"`
	Zone      string `json:"zone"`
	Tap       bool   `json:"tap"`
	Sacrifice bool   `json:"sacrifice"`
	Exile     bool   `json:"exile"`
}

func (f *previewFixture) previewSources(card uuid.UUID, query ...string) (bool, []string, []previewSource) {
	f.t.Helper()
	path := "/games/" + f.gameID.String() + "/auto-tap-preview?card=" + card.String()
	for _, q := range query {
		path += "&" + q
	}
	resp := f.get(path)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		f.t.Fatalf("GET %s: status %d", path, resp.StatusCode)
	}
	var body struct {
		OK      bool            `json:"ok"`
		Plan    []string        `json:"plan"`
		Sources []previewSource `json:"sources"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		f.t.Fatalf("decode preview: %v", err)
	}
	return body.OK, body.Plan, body.Sources
}

// {R}{R}{R} off a Mountain, a Gold and a Simian Spirit Guide in hand:
// three sources, three different payments, and the one in hand is
// named — zone "hand", exiled — rather than silently missing.
func TestAutoTapPreviewNamesAHandSourceAndACrackedToken(t *testing.T) {
	f := newPreviewFixture(t)
	spell := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Red Thing", TypeLine: "Creature — Elemental", ManaCost: "{R}{R}{R}",
	}, 1)[0]
	mountain := f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}, 1)[0]
	gold := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Gold", TypeLine: "Token Artifact — Gold", TokenKey: game.TokenKey("gold"),
	}, 1)[0]
	guide := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Simian Spirit Guide", TypeLine: "Creature — Ape Spirit", ManaCost: "{2}{R}", OracleID: simianSpiritGuideOracle,
	}, 1)[0]

	ok, plan, sources := f.previewSources(spell)
	if !ok || len(plan) != 3 || len(sources) != 3 {
		t.Fatalf("Mountain + Gold + Spirit Guide for {R}{R}{R}: ok=%v plan=%v sources=%+v", ok, plan, sources)
	}
	for i := range plan {
		if sources[i].CardID != plan[i] {
			t.Errorf("sources[%d] = %s, plan[%d] = %s — the two lists must be one order", i, sources[i].CardID, i, plan[i])
		}
	}
	by := map[string]previewSource{}
	for _, s := range sources {
		by[s.CardID] = s
	}
	if s := by[mountain.String()]; s.Zone != "battlefield" || !s.Tap || s.Sacrifice || s.Exile {
		t.Errorf("Mountain: %+v, want a battlefield tap", s)
	}
	if s := by[gold.String()]; s.Zone != "battlefield" || s.Tap || !s.Sacrifice || s.Exile {
		t.Errorf("Gold: %+v, want sacrificed WITHOUT a tap", s)
	}
	if s := by[guide.String()]; s.Zone != "hand" || s.Tap || s.Sacrifice || !s.Exile || s.Name != "Simian Spirit Guide" {
		t.Errorf("Spirit Guide: %+v, want named, from the hand, exiled", s)
	}
}

// #1242: a permanent the cast names to its additional cost is not the
// preview's to spend on mana, exactly as CastSpell's auto-tap will not
// spend it. Deadly Dispute ({1}{B}, sacrifice an artifact or creature)
// off a Swamp and one Eldrazi Spawn: the preview without the
// sacrifice named says yes, and with it named says no — the answer the
// cast itself will give.
func TestAutoTapPreviewDoesNotSpendTheNamedSacrifice(t *testing.T) {
	f := newPreviewFixture(t)
	dispute := f.spawn(f.alice, game.ZoneHand, game.Card{
		Name: "Deadly Dispute", TypeLine: "Instant", ManaCost: "{1}{B}", OracleID: "457af74a-02b3-4659-846d-63e482667f34",
	}, 1)[0]
	f.spawn(f.alice, game.ZoneBattlefield, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"}, 1)
	spawn := f.spawn(f.alice, game.ZoneBattlefield, game.Card{
		Name: "Eldrazi Spawn", TypeLine: "Token Creature — Eldrazi Spawn", TokenKey: game.TokenKey("eldrazi-spawn"),
	}, 1)[0]

	if ok, plan, _ := f.previewSources(dispute); !ok || len(plan) != 2 {
		t.Fatalf("no sacrifice named: ok=%v plan=%v, want the Swamp and the Spawn", ok, plan)
	}
	if ok, plan, _ := f.previewSources(dispute, "sacrifice_ids="+spawn.String()); ok {
		t.Errorf("the Spawn named to the sacrifice was also planned for mana: %v", plan)
	}
}
