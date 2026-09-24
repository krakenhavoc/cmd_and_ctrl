package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// waterbend_view_test.go — #1310 / #1311: the wire half of waterbend
// as an activated ability's cost and as a ward's payment. The picker
// options come from the walk the engine validates against, so the
// client never offers a tap the server refuses; and the enumerator's
// payment is a subset of what the view offers (#544).

func waterbendClause(extra string) *game.TapPermanentsCost {
	return &game.TapPermanentsCost{
		Key:   "waterbend",
		Label: "Waterbend " + extra,
		Extra: extra,
		Spec: &game.TargetSpec{
			Mode: "permanent", Label: "an untapped artifact or creature you control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsArtifact() || c.IsCreature()
			},
		},
	}
}

func seatWaterbendSource(g *game.Game, owner uuid.UUID, cost game.AbilityCost) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Waterbender",
		TypeLine:   "Creature — Human",
		OracleID:   "00000000-0000-0000-0000-0000000000f1",
		Power:      2, Toughness: 2,
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label:  "Waterbend: mark",
			Cost:   cost,
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	})
	return id
}

func sortedCopy(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	return out
}

// The clause, its budget, and exactly the permanents that could pay:
// my untapped artifacts and creatures — the source among them, since
// its cost prints no {T}, and a hexproof one, since a cost does not
// target — and never a land, a tapped creature or an opponent's.
func TestActivatedAbilityViewCarriesTheWaterbendClause(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	src := seatWaterbendSource(g, owner, game.AbilityCost{Mana: "{8}", Waterbend: waterbendClause("{8}")})
	ready := seatCreature(g, owner, "Ready", 2, false, true)
	rock := seatArtifactFor(g, owner, "Rock")
	seatCreature(g, owner, "Tired", 2, true, false)
	seatLandFor(g, owner, "A Land")
	seatCreature(g, g.Seats[1].ID, "Theirs", 2, false, false)
	g.BumpLayerVersionForTest()

	wb := vehicleView(t, g, src).ActivatedAbilities[0].Waterbend
	if wb == nil {
		t.Fatal("waterbend is absent on an ability that prints the clause")
	}
	if wb.Key != "waterbend" || wb.Label != "Waterbend {8}" || wb.Max != 8 || wb.DemandsX {
		t.Errorf("waterbend = %+v, want key/label/max 8", wb)
	}
	got := sortedCopy(wb.Options.Cards)
	want := sortedCopy([]string{src.String(), ready.String(), rock.String()})
	if !sameStrings(got, want) {
		t.Errorf("options = %v, want the source, the sick creature and the artifact", wb.Options.Cards)
	}
}

// A cost that also prints {T} spends the source on the {T}, so the
// source is not a waterbender; and a Waterbend {X} ships Max 0 with
// demands_x, the client's cue to size the picker from its X.
func TestWaterbendViewExcludesATappingSourceAndDefersX(t *testing.T) {
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	tapSrc := seatWaterbendSource(g, owner, game.AbilityCost{Mana: "{2}", Tap: true, Waterbend: waterbendClause("{2}")})
	xSrc := seatWaterbendSource(g, owner, game.AbilityCost{Mana: "{X}", MinX: 1, Waterbend: waterbendClause("{X}")})
	g.BumpLayerVersionForTest()

	for _, id := range vehicleView(t, g, tapSrc).ActivatedAbilities[0].Waterbend.Options.Cards {
		if id == tapSrc.String() {
			t.Error("a {T} ability offers its own source as a waterbender")
		}
	}
	xv := vehicleView(t, g, xSrc).ActivatedAbilities[0]
	if xv.Waterbend.Max != 0 || !xv.Waterbend.DemandsX || !xv.DemandsX || xv.MinX != 1 {
		t.Errorf("X ability: waterbend %+v demands_x %v min_x %d", xv.Waterbend, xv.DemandsX, xv.MinX)
	}
}

// An ability with no clause ships none.
func TestNoWaterbendClauseShipsNone(t *testing.T) {
	g := buildActiveGame(t)
	src := seatWaterbendSource(g, g.Seats[0].ID, game.AbilityCost{Mana: "{2}"})
	g.BumpLayerVersionForTest()
	if wb := vehicleView(t, g, src).ActivatedAbilities[0].Waterbend; wb != nil {
		t.Errorf("waterbend = %+v on an ability without the clause", wb)
	}
}

// The coherence property: the enumerator's waterbend payment is a
// subset of the view's options, and the engine accepts it. Six lands
// cannot pay {8}, so the move MUST tap — the two free artifacts first.
func TestWaterbendViewAndEnumeratorAgree(t *testing.T) {
	g := busyTable(t, 0)
	me := g.Seats[g.Turn.ActiveSeat]
	src := seatWaterbendSource(g, me.ID, game.AbilityCost{Mana: "{8}", Waterbend: waterbendClause("{8}")})
	rockA := seatArtifactFor(g, me.ID, "Rock A")
	rockB := seatArtifactFor(g, me.ID, "Rock B")
	g.BumpLayerVersionForTest()

	view := vehicleView(t, g, src).ActivatedAbilities[0].Waterbend
	shown := map[string]bool{}
	for _, id := range view.Options.Cards {
		shown[id] = true
	}
	var taps []string
	found := false
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type != legal.TypeActivateAbility {
			continue
		}
		var p struct {
			SourceCardID string   `json:"source_card_id"`
			WaterbendIDs []string `json:"waterbend_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if p.SourceCardID == src.String() {
			found, taps = true, p.WaterbendIDs
		}
	}
	if !found {
		t.Fatal("the enumerator offered no waterbend activation the view says is payable")
	}
	if len(taps) < 2 {
		t.Fatalf("waterbend_ids = %v, want at least the two free artifacts", taps)
	}
	tapSet := map[string]bool{}
	for _, id := range taps {
		if !shown[id] {
			t.Errorf("the enumerator taps %s, which the view does not offer", id)
		}
		tapSet[id] = true
	}
	if !tapSet[rockA.String()] || !tapSet[rockB.String()] {
		t.Errorf("waterbend_ids = %v, want the free artifacts tapped first", taps)
	}

	ids := make([]uuid.UUID, 0, len(taps))
	for _, s := range taps {
		ids = append(ids, uuid.MustParse(s))
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{
		WaterbendIDs: ids, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("the engine refused the enumerated payment: %v", err)
	}
}

// The pay-unless half: a "Ward—Waterbend {4}" prompt ships the
// chooser's waterbenders as tap_cost, and an ordinary ward ships none.
func TestPayUnlessViewCarriesTheWaterbendClause(t *testing.T) {
	g := buildActiveGame(t)
	caster := g.Seats[1]
	spell := uuid.New()
	g.WithWriteLock(func() {
		g.Stack.PushTop(game.Card{InstanceID: spell, Name: "Doom Blade", TypeLine: "Instant",
			Owner: caster.ID, Controller: caster.ID})
		if g.StackMeta == nil {
			g.StackMeta = map[uuid.UUID]*game.StackItem{}
		}
		g.StackMeta[spell] = &game.StackItem{ID: spell, Kind: game.StackItemSpell,
			Controller: caster.ID, Owner: caster.ID, SourceCardID: spell}
	})
	mine := seatCreature(g, caster.ID, "Helper", 1, false, false)
	seatCreature(g, g.Seats[0].ID, "Warder's", 1, false, false)
	g.WithWriteLock(func() {
		_ = g.QueueCounterUnlessPaidForEffect(game.CounterUnlessPaidPrompt{
			StackItem: spell, Chooser: caster.ID, Cost: "{4}",
			Question: "Ward — waterbend {4}", Waterbend: waterbendClause("{4}"),
		})
	})

	v := ViewOfGame(g)
	var pc *PendingChoiceView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoicePayUnless) {
			pc = &v.PendingChoices[i]
		}
	}
	if pc == nil {
		t.Fatal("no pay_unless in the view")
	}
	if pc.TapCost == nil || pc.TapCost.Max != 4 || pc.TapCost.Key != "waterbend" {
		t.Fatalf("tap_cost = %+v, want the waterbend clause with max 4", pc.TapCost)
	}
	if !sameStrings(pc.TapCost.Options.Cards, []string{mine.String()}) {
		t.Errorf("options = %v, want only the chooser's creature", pc.TapCost.Options.Cards)
	}
}
