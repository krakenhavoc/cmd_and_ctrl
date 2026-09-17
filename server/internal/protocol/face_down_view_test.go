package protocol

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_view_test.go — #95. A card the viewer does not know must
// reach them as game state and nothing else. Before this file the
// redaction zeroed the name, the costs and the faces, and left every
// field that is read off the same card text: the catalog flag, the
// target prompt, the mana abilities, the impulse grant. A face-down
// Forest that Necropotence exiled told the whole table it tapped for
// {G}.
//
// Two tests. The first is the real shape end to end: the engine's own
// face-down exile, projected for every seat. The second is the table
// the fix is actually pinned by: every zone a card can be projected
// out of, crossed with every kind of viewer, against a CardView with
// every field populated, so a field added to CardView later is
// redacted-or-allowlisted by construction rather than by memory.

// redactedCardKeys is every JSON key an unknown card may carry. Each
// one describes the game around the card, not the card: who owns and
// controls it, whether it is tapped, the damage on it, the combat
// declarations and attachments that name it, and where it sits.
// `name` is here only because the field has no omitempty; its value
// is checked separately.
var redactedCardKeys = map[string]bool{
	"instance_id":           true,
	"name":                  true,
	"owner":                 true,
	"controller":            true,
	"tapped":                true,
	"damage_marked":         true,
	"face_down":             true,
	"battle_x":              true,
	"battle_y":              true,
	"attacking_target":      true,
	"attacking_target_kind": true,
	"blocking_target":       true,
	"goaded_by":             true,
	"attached_to":           true,
	"no_untap":              true,
}

// assertRedacted fails on any key outside redactedCardKeys and on a
// non-empty name.
func assertRedacted(t *testing.T, label string, c CardView) {
	t.Helper()
	if c.KnownByYou {
		t.Errorf("%s: known_by_you is true on a card the viewer does not know", label)
	}
	if c.Name != "" {
		t.Errorf("%s: name %q shipped to a non-knower", label, c.Name)
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("%s: marshal: %v", label, err)
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal(raw, &keys); err != nil {
		t.Fatalf("%s: unmarshal: %v", label, err)
	}
	var leaked []string
	for k := range keys {
		if !redactedCardKeys[k] {
			leaked = append(leaked, k)
		}
	}
	sort.Strings(leaked)
	if len(leaked) > 0 {
		t.Errorf("%s: non-knower received identifying fields %v\n%s", label, leaked, raw)
	}
}

func TestFaceDownExileShipsNoIdentityToAnySeat(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-face-down-forest"

	prevCatalog, prevMode := game.IsCatalogCard, game.CatalogTargetMode
	game.IsCatalogCard = func(id string) bool { return id == oracle }
	game.CatalogTargetMode = func(id string) string {
		if id == oracle {
			return "creature"
		}
		return ""
	}
	t.Cleanup(func() { game.IsCatalogCard, game.CatalogTargetMode = prevCatalog, prevMode })

	// Necropotence's shape: the top card of a library exiled face
	// down, known to nobody. A basic Forest because its mana ability
	// is intrinsic to the type line — no catalog hook is needed for
	// mana_abilities to be populated, which is exactly why it leaked.
	necro := game.NewCard("Forest", me.ID)
	necro.TypeLine = "Basic Land — Forest"
	necro.OracleID = oracle
	// Foretell's shape: face down in exile, known to its owner, with
	// a cast grant whose cost is the card's foretell cost. Built by
	// hand — nothing in the engine makes one yet — because the grant
	// is the field that names the card loudest.
	foretold := game.NewCard("Forest", me.ID)
	foretold.TypeLine = "Basic Land — Forest"
	foretold.OracleID = oracle
	foretold.FaceDown = true
	foretold.KnownBy = map[uuid.UUID]bool{me.ID: true}
	foretold.ExilePlay = game.ExilePlayPermission{Player: me.ID, WhileExiled: true, CostOverride: "{1}{G}"}

	var exiled []uuid.UUID
	g.WithWriteLock(func() {
		me.Library.PushTop(necro)
		var err error
		exiled, err = g.ExileTopFaceDownForEffect(me.ID, 1)
		if err != nil {
			t.Fatalf("exile face down: %v", err)
		}
		g.Exile.PushTop(foretold)
	})
	if len(exiled) != 1 || exiled[0] != necro.InstanceID {
		t.Fatalf("exiled %v, want the seeded top card", exiled)
	}

	find := func(v GameView, id uuid.UUID) CardView {
		t.Helper()
		for _, c := range v.Exile.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("card %s missing from the exile view", id)
		return CardView{}
	}

	for _, seat := range []struct {
		name string
		id   uuid.UUID
	}{{"controller", me.ID}, {"opponent", opp.ID}} {
		v := ViewOfGameFor(g, seat.id.String())
		c := find(v, necro.InstanceID)
		if !c.FaceDown {
			t.Errorf("%s: face_down not set on the Necropotence exile", seat.name)
		}
		assertRedacted(t, seat.name+" / face-down exile", c)
	}

	// The owner of the foretold card reads all of it, grant included;
	// the opponent reads none of it.
	mine := find(ViewOfGameFor(g, me.ID.String()), foretold.InstanceID)
	if !mine.KnownByYou || mine.Name != "Forest" || mine.ExilePlay == nil || !mine.Auto || len(mine.ManaAbilities) == 0 {
		t.Errorf("owner lost their own face-down card's characteristics: %+v", mine)
	}
	theirs := find(ViewOfGameFor(g, opp.ID.String()), foretold.InstanceID)
	if !theirs.FaceDown {
		t.Errorf("opponent: face_down not set on the foretold card")
	}
	assertRedacted(t, "opponent / known-to-owner face-down exile", theirs)
}

// everyFieldCardView returns a CardView with every exported field
// set, so the allowlist check below sees every key the wire could
// carry. The reflection guard in the test fails when a new field is
// added and left zero here — which would otherwise let that field
// leak past the table without the table noticing.
func everyFieldCardView(owner string, knowers map[string]bool) CardView {
	one := 1
	lt := &LegalTargetsView{Cards: []string{"target"}, Min: 1, Max: 1}
	x, y := 0.25, 0.75
	return CardView{
		InstanceID:          uuid.NewString(),
		Name:                "Hidden Name",
		Owner:               owner,
		Controller:          owner,
		ScryfallID:          "scryfall",
		TypeLine:            "Legendary Creature — Test",
		Power:               3,
		Toughness:           3,
		Tapped:              true,
		Counters:            map[string]int{"+1/+1": 1},
		IsCommander:         true,
		DamageMarked:        1,
		FaceDown:            true,
		KnownByYou:          true,
		knowers:             knowers,
		oracleID:            "oracle",
		BattleX:             &x,
		BattleY:             &y,
		AttackingTarget:     "defender",
		AttackingTargetKind: "player",
		ProtectorPlayer:     "protector",
		Defense:             4,
		BlockingTarget:      "attacker",
		GoadedBy:            "goader",
		AttachedTo:          &TargetRefView{Kind: "card", ID: "host"},
		NoUntap:             &NoUntapView{Static: true, Next: []string{"next-player"}},
		Auto:                true,
		Unimplemented:       true,
		TargetMode:          "creature",
		LegalTargets:        lt,
		Modes:               &ModeSpecView{Prompt: "Choose one", Min: 1, Max: 1, Options: []ModeOptionView{{Label: "mode"}}},
		AdditionalCost:      &AdditionalCostView{DiscardCards: 1},
		AlternativeCosts:    []AlternativeCostView{{Key: "overload", Label: "Overload {6}{U}"}},
		TapCost:             &TapCostView{Key: "convoke", Options: lt},
		TargetCostNotes:     []string{"This spell costs {1} more to cast for each target beyond the first."},
		CastableHere:        true,
		ExilePlay:           &ExilePlayView{Player: owner, CostOverride: "{1}{U}"},
		ActivatedAbilities:  []ActivatedAbilityView{{Index: 0, Label: "{T}: Draw", LoyaltyCost: &one}},
		SummoningSick:       true,
		LoyaltyActivated:    true,
		ManaCost:            "{2}{U}",
		ManaAbilities:       []ManaAbilityView{{Index: 0, Label: "Add {U}"}},
		Abilities:           []string{"flying"},
		Restrictions:        []string{"cant_block"},
		Layout:              "modal_dfc",
		Faces:               []CardFaceView{{Name: "Hidden Name"}, {Name: "Hidden Back"}},
		ActiveFace:          1,
	}
}

func TestRedactionZoneByViewer(t *testing.T) {
	ownerID, oppID := uuid.NewString(), uuid.NewString()

	probe := everyFieldCardView(ownerID, nil)
	pv := reflect.ValueOf(probe)
	for i := 0; i < pv.NumField(); i++ {
		f := pv.Type().Field(i)
		if f.IsExported() && pv.Field(i).IsZero() {
			t.Fatalf("everyFieldCardView leaves CardView.%s zero; set it so the allowlist check covers it", f.Name)
		}
	}

	type outcome int
	const (
		full outcome = iota
		redacted
		absent
	)
	zones := []string{"battlefield", "stack", "exile", "graveyard", "command", "hand", "library"}
	viewers := []struct {
		name    string
		viewer  string
		knowers map[string]bool
		want    func(zone string) outcome
	}{
		{
			name: "owner who knows it", viewer: ownerID,
			knowers: map[string]bool{ownerID: true},
			want:    func(string) outcome { return full },
		},
		{
			// Necropotence's exile, and every card in the owner's own
			// library: the owner's seat is not a knower.
			name: "owner who does not know it", viewer: ownerID,
			knowers: map[string]bool{},
			want:    func(string) outcome { return redacted },
		},
		{
			name: "opponent", viewer: oppID,
			knowers: map[string]bool{ownerID: true},
			want: func(zone string) outcome {
				if zone == "hand" || zone == "library" {
					return absent
				}
				return redacted
			},
		},
		{
			// The empty viewer knows every card on the table and sees
			// no hand or library at all (FilterViewFor's contract).
			name: "spectator", viewer: "",
			knowers: map[string]bool{},
			want: func(zone string) outcome {
				if zone == "hand" || zone == "library" {
					return absent
				}
				return full
			},
		},
	}

	for _, zone := range zones {
		for _, vw := range viewers {
			t.Run(zone+"/"+vw.name, func(t *testing.T) {
				card := everyFieldCardView(ownerID, vw.knowers)
				// KnownByYou is stamped by the filter, never trusted
				// from the input.
				card.KnownByYou = false
				zv := func(kind string) ZoneView {
					z := ZoneView{Kind: kind, Cards: []CardView{}}
					if kind == zone {
						z.Cards = []CardView{card}
						z.Count = 1
					}
					return z
				}
				v := GameView{
					Seats: []PlayerView{
						{ID: ownerID, Hand: zv("hand"), Library: zv("library"), Graveyard: zv("graveyard"), Command: zv("command")},
						{ID: oppID, Hand: zv("-"), Library: zv("-"), Graveyard: zv("-"), Command: zv("-")},
					},
					Battlefield: zv("battlefield"),
					Stack:       zv("stack"),
					Exile:       zv("exile"),
				}
				out := FilterViewFor(v, vw.viewer)
				var got []CardView
				switch zone {
				case "battlefield":
					got = out.Battlefield.Cards
				case "stack":
					got = out.Stack.Cards
				case "exile":
					got = out.Exile.Cards
				case "graveyard":
					got = out.Seats[0].Graveyard.Cards
				case "command":
					got = out.Seats[0].Command.Cards
				case "hand":
					got = out.Seats[0].Hand.Cards
				case "library":
					got = out.Seats[0].Library.Cards
				}

				switch vw.want(zone) {
				case absent:
					if len(got) != 0 {
						t.Errorf("card present, want it withheld: %+v", got[0])
					}
				case full:
					if len(got) != 1 {
						t.Fatalf("got %d cards, want 1", len(got))
					}
					want := card
					want.knowers = nil
					want.KnownByYou = true
					if !reflect.DeepEqual(got[0], want) {
						t.Errorf("knower's copy altered:\n got %+v\nwant %+v", got[0], want)
					}
				case redacted:
					if len(got) != 1 {
						t.Fatalf("got %d cards, want 1", len(got))
					}
					assertRedacted(t, zone+"/"+vw.name, got[0])
					// The allowlisted state must survive, or the board
					// stops being observable (S13.5).
					if got[0].InstanceID != card.InstanceID || got[0].Controller != ownerID || !got[0].Tapped || !got[0].FaceDown {
						t.Errorf("public state lost in redaction: %+v", got[0])
					}
				}
			})
		}
	}
}
