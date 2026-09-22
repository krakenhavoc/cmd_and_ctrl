package protocol

import (
	"encoding/json"
	"fmt"
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
	"instance_id":   true,
	"name":          true,
	"owner":         true,
	"controller":    true,
	"tapped":        true,
	"damage_marked": true,
	// #667: a regeneration shield is a fact about what happens to the
	// permanent next, not about which card it is. A face-down
	// creature with a shield on it is something the table has to be
	// able to see before deciding whether removal is worth casting.
	"regeneration_shields": true,
	"face_down":            true,
	// ADR 0069: WHY it is face down is public — everyone can see
	// that a permanent is a morph and that an exiled card is
	// foretold. The identity of the card under it is not.
	"face_down_kind":        true,
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
	foretoldGrant := game.CastPermission{Player: me.ID, Duration: game.WhileInZoneDuration(), Cost: "{1}{G}"}

	var exiled []uuid.UUID
	g.WithWriteLock(func() {
		me.Library.PushTop(necro)
		var err error
		exiled, err = g.ExileTopFaceDownForEffect(me.ID, 1)
		if err != nil {
			t.Fatalf("exile face down: %v", err)
		}
		g.Exile.PushTop(foretold)
		g.GrantCastPermissionOverCardForEffect(foretold.InstanceID, foretoldGrant)
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
		Colors:              []string{"B", "R"},
		Protection:          []ProtectionView{{Printed: "red", Kind: "color", Value: "R"}},
		NegativePower:       -2,
		Power:               3,
		Toughness:           3,
		Tapped:              true,
		Counters:            map[string]int{"+1/+1": 1},
		IsCommander:         true,
		DamageMarked:        1,
		RegenerationShields: 2,
		FaceDown:            true,
		// An EXILE kind on purpose: a CR 708.2 permanent kind would
		// put the public 2/2 body back after the redaction (decision
		// 6), and this table is the guard for every card that is NOT
		// one. TestFaceDownPermanentShipsItsPublicBody covers that
		// cell.
		FaceDownKind:        "exiled",
		FaceVisible:         true,
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
		// #992: the announce surface is one embedded block now, on
		// CardView and on every castable CardFaceView alike. The
		// reflection guard below recurses into it, so a field added
		// here is still a field this table has to place.
		CastSurfaceView:    everyFieldCastSurface(lt),
		ExilePlay:          &ExilePlayView{Player: owner, CostOverride: "{1}{U}"},
		ActivatedAbilities: []ActivatedAbilityView{{Index: 0, Label: "{T}: Draw", LoyaltyCost: &one}},
		ZoneAbilities:      []ActivatedAbilityView{{Index: 0, Label: "Cycling {2}", DiscardSelf: true, ManaCost: "{2}"}},
		SpecialActions:     []SpecialActionView{{Kind: "foretell", Label: "Foretell {2}", Cost: "{2}", Available: true}},
		SummoningSick:      true,
		LoyaltyActivated:   true,
		ClassLevel:         3,
		Solved:             true,
		// #781. Deliberately NOT added to redactedCardKeys: both are
		// public on a card the viewer can see and both are stripped
		// from one they cannot, because "Elf" names Cavern of Souls
		// and "G" names Coldsteel Heart. They are here so this table
		// is the thing that proves it.
		ChosenColor:   "G",
		NamedTribe:    "Elf",
		ChosenName:    "Sol Ring",
		ManaCost:      "{2}{U}",
		ManaAbilities: []ManaAbilityView{{Index: 0, Label: "Add {U}"}},
		Abilities:     []string{"flying"},
		Restrictions:  []string{"cant_block"},
		Layout:        "modal_dfc",
		// #992: both faces carry the per-face announce block, so the
		// redaction table covers a face's `legal_targets` as well as
		// the card's — a back face's target clause names the card
		// just as loudly as its mana cost does.
		Faces: []CardFaceView{
			{Name: "Hidden Name", TypeLine: "Sorcery", ManaCost: "{6}{U}", OracleText: "Front text", Power: 1, Toughness: 1, Image: "/cards/scryfall/image?face=0", CastSurfaceView: everyFieldCastSurface(lt)},
			{Name: "Hidden Back", TypeLine: "Land", ManaCost: "{1}{U}", OracleText: "Back text", Power: 2, Toughness: 2, Image: "/cards/scryfall/image?face=1", CastSurfaceView: everyFieldCastSurface(lt)},
		},
		ActiveFace: 1,
	}
}

// everyFieldCastSurface is everyFieldCardView's announce block, with
// every field of CastSurfaceView non-zero. One constructor, because
// the block is carried twice on the wire now — by the card, for the
// face that is up, and by each castable face (#992) — and the
// redaction table has to see both filled.
func everyFieldCastSurface(lt *LegalTargetsView) CastSurfaceView {
	return CastSurfaceView{
		TargetMode:   "creature",
		LegalTargets: lt,
		Clauses:      []LegalTargetsView{*lt, *lt},
		// #1172: the nested blocks carry every field of their own too,
		// so the nested half of the allowlist guard below is checking
		// a strip rather than a struct that was empty anyway.
		Modes: &ModeSpecView{Prompt: "Choose one", Min: 1, Max: 1, Options: []ModeOptionView{{
			Label: "mode", TargetMode: "creature", LegalTargets: lt, Clauses: []LegalTargetsView{*lt, *lt},
		}}},
		AdditionalCost: &AdditionalCostView{DiscardCards: 1},
		AlternativeCosts: []AlternativeCostView{{
			Key: "overload", Label: "Overload {6}{U}", ManaCost: "{6}{U}", Life: 1, PayLabel: "a blue card",
			TargetMode: "creature", LegalTargets: lt, PayOptions: lt,
			XLockedAtZero: true, PhyrexianSymbols: 1,
		}},
		// #1012: the flag that says the printed cost is not one of
		// the prices this cast may claim. Redacted with the offer
		// list it is only meaningful beside.
		AlternativeCostRequired: true,
		TapCost:                 &TapCostView{Key: "convoke", Options: lt},
		TargetCostNotes:         []string{"This spell costs {1} more to cast for each target beyond the first."},
		PhyrexianSymbols:        1,
		CastableHere:            true,
		OptionalCosts:           []OptionalCostView{{Index: 0, Key: "kicker", Label: "Kicker {4}", ManaCost: "{4}", MaxTimes: 1}},
		// ADR 0073 §7: the cast gate's stamp. Redacted like the rest
		// of the cost surface — a legendary-sorcery clause says more
		// about a face-down card than its mana cost does.
		CantCast: "Each player can't cast more than one spell each turn.",
	}
}

// assertEveryExportedFieldSet fails when any exported field of `v` is
// left at its zero value, RECURSING into embedded structs — which is
// what keeps the guard honest now that the announce surface is one
// embedded CastSurfaceView (#992) rather than thirteen fields listed
// in line. Without the recursion a field added to that block would be
// covered by "the embedded struct is non-zero" and could leak past
// the allowlist table without the table noticing, which is exactly
// what this guard was written to prevent.
func assertEveryExportedFieldSet(t *testing.T, path string, v reflect.Value) {
	t.Helper()
	for i := 0; i < v.NumField(); i++ {
		f := v.Type().Field(i)
		if !f.IsExported() {
			continue
		}
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			assertEveryExportedFieldSet(t, path+"."+f.Name, v.Field(i))
			continue
		}
		if v.Field(i).IsZero() {
			t.Fatalf("everyFieldCardView leaves %s.%s zero; set it so the allowlist check covers it", path, f.Name)
		}
	}
}

func TestRedactionZoneByViewer(t *testing.T) {
	ownerID, oppID := uuid.NewString(), uuid.NewString()

	probe := everyFieldCardView(ownerID, nil)
	assertEveryExportedFieldSet(t, "CardView", reflect.ValueOf(probe))
	// #992: and the per-face block, which the walk above cannot
	// reach — `Faces` is a non-zero slice whatever is inside it.
	for i, f := range probe.Faces {
		assertEveryExportedFieldSet(t, fmt.Sprintf("CardView.Faces[%d]", i), reflect.ValueOf(f))
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

// TestFaceDownPermanentShipsItsPublicBody is ADR 0069 decision 6 —
// the one cell the table above deliberately does not cover.
//
// CR 708.2 makes a face-down permanent a 2/2 colourless creature with
// no name, and that body is PUBLIC: an opponent has to see it to
// block it, target it and count it. The #646 redaction strips name,
// type line, colours and P/T because on an ordinary hidden card those
// fields name it — on this one they ARE the projection and name
// nothing, so they are put back after the redaction.
//
// What must still be gone is everything that identifies the card
// underneath: the art, the cost, the faces, the ability lists, the
// catalog flags.
func TestFaceDownPermanentShipsItsPublicBody(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	// A manifested Griselbrand: a 7/7 legendary Demon with keywords
	// and a mana cost, so every field that could leak is loud.
	loud := game.NewCard("Griselbrand", me.ID)
	loud.TypeLine = "Legendary Creature — Demon"
	loud.OracleID = "oracle-grissy"
	loud.ScryfallID = "scryfall-grissy"
	loud.ManaCost = "{4}{B}{B}"
	loud.Colors = []string{"B"}
	loud.Power, loud.Toughness = 7, 7
	loud.Keywords = []string{"flying", "lifelink"}
	id := loud.InstanceID
	g.WithWriteLock(func() {
		me.Library.PushTop(loud)
		if _, err := g.ManifestForEffect(me.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})

	find := func(viewer uuid.UUID) CardView {
		t.Helper()
		for _, c := range ViewOfGameFor(g, viewer.String()).Battlefield.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("manifested card missing from %s's battlefield view", viewer)
		return CardView{}
	}

	theirs := find(opp.ID)
	if !theirs.FaceDown || theirs.FaceDownKind != "manifested" {
		t.Errorf("opponent: face_down=%v kind=%q, want true/manifested", theirs.FaceDown, theirs.FaceDownKind)
	}
	if theirs.FaceVisible || theirs.KnownByYou {
		t.Error("opponent: may look at a face-down permanent they do not control (CR 708.5)")
	}
	if theirs.TypeLine != "Creature" || theirs.Power != 2 || theirs.Toughness != 2 {
		t.Errorf("opponent: %q %d/%d, want the public CR 708.2 body: Creature 2/2",
			theirs.TypeLine, theirs.Power, theirs.Toughness)
	}
	if theirs.Name != "" || len(theirs.Colors) != 0 || len(theirs.Abilities) != 0 {
		t.Errorf("opponent: name %q / colors %v / abilities %v, want a nameless colourless vanilla",
			theirs.Name, theirs.Colors, theirs.Abilities)
	}
	// Identity: still gone, and mostly never stamped, because
	// CatalogKey answers "" for a face-down permanent.
	if theirs.ScryfallID != "" || theirs.ManaCost != "" || len(theirs.Faces) != 0 ||
		theirs.Auto || theirs.TargetMode != "" || theirs.Unimplemented ||
		len(theirs.ManaAbilities) != 0 || len(theirs.ActivatedAbilities) != 0 {
		t.Errorf("opponent: the card under the back leaked: %+v", theirs)
	}

	// The controller may look (CR 708.5): they get the art and the
	// face, and face_visible says so — but the OBJECT is still the
	// nameless 2/2, because that is what it is for both of them.
	mine := find(me.ID)
	if !mine.FaceVisible || !mine.KnownByYou {
		t.Error("controller: may not look at their own face-down permanent (CR 708.5)")
	}
	if mine.ScryfallID != "scryfall-grissy" {
		t.Errorf("controller: scryfall_id = %q, want the art so the client can show them their own card", mine.ScryfallID)
	}
	if mine.Name != "" || mine.TypeLine != "Creature" || mine.Power != 2 {
		t.Errorf("controller: %q %q %d/%d, want the same CR 708.2 object the table sees",
			mine.Name, mine.TypeLine, mine.Power, mine.Toughness)
	}
}

// TestFaceVisibleTracksTheViewersRule pins the per-viewer field against
// ADR 0069's table, for the two exile kinds the engine can make today.
func TestFaceVisibleTracksTheViewersRule(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	necro := game.NewCard("Forest", me.ID)
	necro.TypeLine = "Basic Land — Forest"
	foretold := game.NewCard("Saw It Coming", me.ID)
	foretold.TypeLine = "Instant"
	foretold.SetFaceDown(game.FaceDownForetold)
	foretold.AddKnower(me.ID)

	g.WithWriteLock(func() {
		me.Library.PushTop(necro)
		if _, err := g.ExileTopFaceDownForEffect(me.ID, 1); err != nil {
			t.Fatalf("exile face down: %v", err)
		}
		g.Exile.PushTop(foretold)
	})

	find := func(viewer uuid.UUID, id uuid.UUID) CardView {
		t.Helper()
		for _, c := range ViewOfGameFor(g, viewer.String()).Exile.Cards {
			if c.InstanceID == id.String() {
				return c
			}
		}
		t.Fatalf("card %s missing from %s's exile view", id, viewer)
		return CardView{}
	}

	for _, tc := range []struct {
		name     string
		id       uuid.UUID
		kind     string
		ownerMay bool
	}{
		{"CR 406.3 Necropotence exile: nobody may look", necro.InstanceID, "exiled", false},
		{"CR 702.143d foretold: the owner may look", foretold.InstanceID, "foretold", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mine := find(me.ID, tc.id)
			if mine.FaceDownKind != tc.kind {
				t.Errorf("owner: face_down_kind = %q, want %q", mine.FaceDownKind, tc.kind)
			}
			if mine.FaceVisible != tc.ownerMay {
				t.Errorf("owner: face_visible = %v, want %v", mine.FaceVisible, tc.ownerMay)
			}
			theirs := find(opp.ID, tc.id)
			if theirs.FaceVisible {
				t.Error("opponent: face_visible on a face-down exiled card")
			}
			// An exiled card gets NO CR 708.2 body: it is not a
			// permanent, and inventing a 2/2 in exile would be
			// inventing a creature.
			if theirs.TypeLine != "" || theirs.Power != 0 {
				t.Errorf("opponent: a face-down EXILED card shipped %q %d/%d",
					theirs.TypeLine, theirs.Power, theirs.Toughness)
			}
		})
	}
}

// --- #1169: the hand's public half is an allowlist -------------------
//
// A hand is not a public zone the way a graveyard is: a revealed card
// (Thoughtseize, Telepathy) is ONE card the viewer has been shown, not
// a pile they may read. `keepKnownInHandZone` used to keep that true
// by clearing six announce fields from a list of field NAMES, which
// is a second list in a second place — the drift `castStamps` exists
// to end — and being a list of what to REMOVE it had gone stale
// twice: it never covered `optional_costs`, and since #992 it never
// covered the per-face blocks at all.
//
// The narrowing lives in `castStamps.publicIn` now, as an allowlist,
// and this is the guard on it: every field of CastSurfaceView has to
// be PLACED in the table below, so a field added to the block is a
// decision somebody makes rather than one nobody notices.

// castSurfaceScope is how far one announce field travels: never off
// the asking seat's frame, out to every viewer of a PUBLIC pile, or
// out to a viewer who has merely been shown one card in a hand.
type castSurfaceScope int

const (
	// surfacePrivate answers "what may YOU announce" and reaches the
	// asking seat's frame alone, in every zone (#1055).
	surfacePrivate castSurfaceScope = iota
	// surfacePublicPile is a fact about the card in a PUBLIC zone —
	// a graveyard, a library top, the command zone, exile — which
	// every player may pick up and read there, and which a viewer
	// shown one card of a HAND may not.
	surfacePublicPile
	// surfacePublicAnywhere survives even on a revealed hand card.
	surfacePublicAnywhere
)

// castSurfaceScopes places every field of CastSurfaceView. A field
// missing from this table fails the guard below rather than defaulting
// to anything: "which viewers may read this" is the question a new
// announce field has to answer, and the answer belongs beside the
// other twelve.
var castSurfaceScopes = map[string]castSurfaceScope{
	// #1055: the three that answer "what may YOU announce".
	"CastableHere": surfacePrivate,
	"LegalTargets": surfacePrivate,
	"Clauses":      surfacePrivate,
	// #1169: cost-shaped facts about the card. Public on a public
	// pile — an escape offer is priced by a graveyard everybody can
	// count — and not on a hand card the viewer was shown one of.
	// `alternative_costs` is also a live leak of the REST of that
	// hand: a pitch offer (Force of Will) carries `pay_options`,
	// which is a list of instance IDs out of it.
	"Modes":                   surfacePublicPile,
	"AlternativeCosts":        surfacePublicPile,
	"AlternativeCostRequired": surfacePublicPile,
	"TapCost":                 surfacePublicPile,
	"PhyrexianSymbols":        surfacePublicPile,
	"TargetCostNotes":         surfacePublicPile,
	// #1169: and the four that are not cost-shaped. `target_mode` is
	// the card's printed prompt shape and is already public on a
	// revealed card's FACES (#992); `additional_cost` and
	// `optional_costs` are printed clauses whose pickers read the
	// public battlefield; `cant_cast` is a Rule of Law on the
	// battlefield, visible to everybody in the same words.
	"TargetMode":     surfacePublicAnywhere,
	"AdditionalCost": surfacePublicAnywhere,
	"OptionalCosts":  surfacePublicAnywhere,
	"CantCast":       surfacePublicAnywhere,
}

// survivesIn reports whether a field of this scope is still set after
// the public strip for `kind`.
func (s castSurfaceScope) survivesIn(kind game.ZoneKind) bool {
	switch s {
	case surfacePublicAnywhere:
		return true
	case surfacePublicPile:
		return kind != game.ZoneHand
	default:
		return false
	}
}

// --- #1172: the same question one level down -------------------------
//
// `modes` and `alternative_costs` are surfacePublicPile above — the
// card's printed text, which a bystander reads off a graveyard — and
// each carries board-derived lists of its own that are the ASKING
// SEAT's answer. The two tables below place every field of the nested
// blocks on the same line the outer one is placed on, so a field added
// to either is a decision somebody makes rather than one nobody
// notices.
//
// The rule that decides a row: does this field come off the PRINTED
// CARD (public with its parent) or off the BOARD for one player
// (surfacePrivate, and gone from the public projection).
var modeOptionScopes = map[string]castSurfaceScope{
	// Printed text: the bullet and the shape of its prompt.
	"Label":      surfacePublicPile,
	"TargetMode": surfacePublicPile,
	// #1172: the legal sets. Narrowed by hexproof, shroud, protection
	// and "target opponent", so seat A's is not seat B's to read.
	"LegalTargets": surfacePrivate,
	"Clauses":      surfacePrivate,
}

var alternativeCostScopes = map[string]castSurfaceScope{
	// Printed price: what the offer is called, what it charges, and
	// the two derived facts about the COST string (CR 107.3b's X lock,
	// CR 107.4's symbol count).
	"Key":              surfacePublicPile,
	"Label":            surfacePublicPile,
	"ManaCost":         surfacePublicPile,
	"Life":             surfacePublicPile,
	"PayLabel":         surfacePublicPile,
	"TargetMode":       surfacePublicPile,
	"XLockedAtZero":    surfacePublicPile,
	"PhyrexianSymbols": surfacePublicPile,
	// #1172: the two board-derived lists. `legal_targets` is the
	// clause this offer leaves the spell with, resolved for the asking
	// seat; `pay_options` is "the blue cards in YOUR hand", "the cards
	// in YOUR graveyard" — not a target list (a cost does not target,
	// CR 601.2h) but one seat's all the same.
	"LegalTargets": surfacePrivate,
	"PayOptions":   surfacePrivate,
}

// TestHandPublicCastSurfaceIsAnAllowlist is the #1169 guard, and it is
// the reflection shape TestRedactionZoneByViewer above uses one field
// list over: every field of the block, placed by a table, checked
// against what the strip actually leaves behind.
//
// It fails three ways, and each is a thing that went wrong before:
// a field nobody placed (the list went stale), a field the hand strip
// keeps that the table says it should not (a leak), and a field the
// strip clears that the table says is public (a regression on a
// public pile, where the same function does the stripping).
func TestHandPublicCastSurfaceIsAnAllowlist(t *testing.T) {
	lt := &LegalTargetsView{Cards: []string{"target"}, Min: 1, Max: 1}
	full := everyFieldCastSurface(lt)
	assertEveryExportedFieldSet(t, "CastSurfaceView", reflect.ValueOf(full))

	typ := reflect.TypeOf(CastSurfaceView{})
	for i := 0; i < typ.NumField(); i++ {
		if _, ok := castSurfaceScopes[typ.Field(i).Name]; !ok {
			t.Fatalf("CastSurfaceView.%s is not placed in castSurfaceScopes: say whether a viewer "+
				"shown one card of somebody else's HAND may read it, and whether a bystander at a "+
				"public pile may", typ.Field(i).Name)
		}
	}

	// #1172: and every field of the two NESTED blocks, placed the same
	// way. A field nobody has placed fails here rather than defaulting
	// to public, which is the direction that leaks.
	for name, typ := range map[string]reflect.Type{
		"ModeOptionView":      reflect.TypeOf(ModeOptionView{}),
		"AlternativeCostView": reflect.TypeOf(AlternativeCostView{}),
	} {
		table := modeOptionScopes
		if name == "AlternativeCostView" {
			table = alternativeCostScopes
		}
		for i := 0; i < typ.NumField(); i++ {
			if _, ok := table[typ.Field(i).Name]; !ok {
				t.Fatalf("%s.%s is not placed: say whether it comes off the PRINTED CARD (public "+
					"with its parent on a pile) or off the board for one player (#1172)",
					name, typ.Field(i).Name)
			}
		}
	}

	// A HAND and a GRAVEYARD, through the one function that knows the
	// field list. The graveyard column is not decoration: it is what
	// stops the hand's narrowing being applied to the piles a card in
	// them is genuinely public in.
	for _, kind := range []game.ZoneKind{game.ZoneHand, game.ZoneGraveyard} {
		public := castStamps{CastSurfaceView: full}.publicIn(kind).CastSurfaceView
		got := reflect.ValueOf(public)
		for i := 0; i < typ.NumField(); i++ {
			name := typ.Field(i).Name
			want := castSurfaceScopes[name].survivesIn(kind)
			if set := !got.Field(i).IsZero(); set != want {
				t.Errorf("%s: %s set=%v, want %v", kind, name, set, want)
			}
		}
		// #1172: the nested half. On a HAND the parents are gone
		// altogether, so there is nothing to walk; on a public pile
		// they survive and each nested field has to land where its
		// table put it.
		if kind == game.ZoneHand {
			if public.Modes != nil || public.AlternativeCosts != nil {
				t.Errorf("hand: the parents of the nested blocks survived the strip")
			}
			continue
		}
		if public.Modes == nil || len(public.Modes.Options) != 1 {
			t.Fatalf("%s: `modes` did not survive the public strip; the nested check below is vacuous", kind)
		}
		assertNestedScopes(t, kind, "modes[0]", reflect.ValueOf(public.Modes.Options[0]), modeOptionScopes)
		if len(public.AlternativeCosts) != 1 {
			t.Fatalf("%s: `alternative_costs` did not survive the public strip", kind)
		}
		assertNestedScopes(t, kind, "alternative_costs[0]",
			reflect.ValueOf(public.AlternativeCosts[0]), alternativeCostScopes)

		// And the copy is a COPY: blanking the public projection in
		// place would take the nested sets off the asking seat's own
		// answer as well, because the two come out of ONE
		// castStampsFor call and share the pointer and the slice
		// header.
		if full.Modes.Options[0].LegalTargets == nil || full.Modes.Options[0].Clauses == nil {
			t.Errorf("%s: the public strip reached through into the seat's own `modes`", kind)
		}
		if full.AlternativeCosts[0].LegalTargets == nil || full.AlternativeCosts[0].PayOptions == nil {
			t.Errorf("%s: the public strip reached through into the seat's own `alternative_costs`", kind)
		}
	}
}

// assertNestedScopes checks one nested block against its table: a
// surfacePrivate field must be gone from the public projection and
// every other field must still be there (#1172).
func assertNestedScopes(t *testing.T, kind game.ZoneKind, path string, v reflect.Value, table map[string]castSurfaceScope) {
	t.Helper()
	typ := v.Type()
	for i := 0; i < typ.NumField(); i++ {
		name := typ.Field(i).Name
		want := table[name].survivesIn(kind)
		if set := !v.Field(i).IsZero(); set != want {
			t.Errorf("%s: %s.%s set=%v, want %v", kind, path, name, set, want)
		}
	}
}

// And end to end, which is the half a unit test over publicIn cannot
// answer: the fields the table calls pile-public must actually be gone
// from a NON-OWNER's copy of a revealed hand card — on the card and on
// its FACES, which is where the old hand-rolled list never reached.
//
// The adventure fixture from per_face_announce_view_test.go, because a
// revealed multi-face card is the case that was live: #992 publishes a
// price list per castable face, `keepKnownInHandZone` cleared the
// CARD's fields by name and nothing else, so a knower of one revealed
// adventure card read its owner's whole per-face offer list.
func TestRevealedHandCardAndItsFacesDropThePileOnlyFields(t *testing.T) {
	withFaceCatalog(t)
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		adventureOracle:        {{Key: "overload", Label: "Overload {4}{R}", ManaCost: "{4}{R}", ClearsTargets: true}},
		adventureOracle + "#1": {{Key: "evoke", Label: "Evoke {R}", ManaCost: "{R}"}},
	})
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)

	// Revealed: the opponent is a knower, so the card survives
	// keepKnownInHandZone and the assertions below are answered by
	// the stamp rather than by the redaction of a card nobody sees.
	c := adventureCard(me.ID, me.ID, opp.ID)
	id := c.InstanceID
	me.Hand.PushTop(c)

	ownHand := ViewOfGameFor(g, me.ID.String()).Seats[0].Hand
	own := cardInSeatZone(t, ownHand, id)
	ownBack := faceOf(t, ownHand, id, 1)
	// Non-vacuity, both levels: the owner is offered a price list on
	// the card and a different one on the Adventure half.
	if keys := keysOf(own.AlternativeCosts); !sameStrings(keys, []string{"overload"}) {
		t.Fatalf("the hand's owner lost their own offers: %v", keys)
	}
	if keys := keysOf(ownBack.AlternativeCosts); !sameStrings(keys, []string{"evoke"}) {
		t.Fatalf("the hand's owner lost the Adventure half's offers: %v", keys)
	}

	theirHand := ViewOfGameFor(g, opp.ID.String()).Seats[0].Hand
	theirs := cardInSeatZone(t, theirHand, id)
	if !theirs.KnownByYou {
		t.Fatalf("the fixture hid the card from the knower; every assertion below would be vacuous")
	}
	if len(theirs.Faces) != 2 {
		t.Fatalf("the knower's copy ships %d faces, want both halves", len(theirs.Faces))
	}
	for _, probe := range []struct {
		label string
		got   CastSurfaceView
	}{
		{"the card", theirs.CastSurfaceView},
		{"faces[0]", theirs.Faces[0].CastSurfaceView},
		{"faces[1]", theirs.Faces[1].CastSurfaceView},
	} {
		v := reflect.ValueOf(probe.got)
		typ := v.Type()
		for i := 0; i < typ.NumField(); i++ {
			name := typ.Field(i).Name
			if castSurfaceScopes[name].survivesIn(game.ZoneHand) || v.Field(i).IsZero() {
				continue
			}
			t.Errorf("a knower of a revealed hand card read %s.%s off it: %+v",
				probe.label, name, v.Field(i).Interface())
		}
	}
	// And the allowlisted half really is still there, or the check
	// above would pass on a card that lost everything.
	if theirs.Faces[1].TargetMode != "any" {
		t.Errorf("faces[1].target_mode = %q on a revealed card, want the public \"any\"", theirs.Faces[1].TargetMode)
	}
}
