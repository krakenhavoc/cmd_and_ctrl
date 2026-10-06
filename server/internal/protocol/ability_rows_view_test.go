package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ability_rows_view_test.go — #2219: `ability_rows` on the wire, the
// list behind the art tile's ⚡ / ◆ / ↻ chips.

const (
	rowsFixtureTitan    = "rows-fixture-titan"    // a trigger, a static and an activated ability
	rowsFixtureAngel    = "rows-fixture-angel"    // keywords only
	rowsFixtureMutation = "rows-fixture-mutation" // other creatures lose all abilities
	rowsFixtureRite     = "rows-fixture-rite"     // creatures you control have a dies trigger
	rowsGrantTrigger    = "rows-fixture/dies-draw"
)

func stubAbilityRowFixtures(t *testing.T) {
	t.Helper()
	fixtures := map[string]*game.CardDef{
		rowsFixtureTitan: {
			Triggered: []game.TriggeredAbility{{
				Key:     "Fixture Titan — 3 damage divided among one, two, or three targets",
				Watches: []game.EventKind{game.EventETB},
			}},
			Static: []game.StaticAbility{{
				Layer:     game.Layer7PT,
				AppliesTo: func(*game.Card, *game.Game, *game.Card) bool { return false },
				Apply:     func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
			}},
			Activated: []game.ActivatedAbilityShape{
				{Label: "{R}: Fixture Titan gets +1/+0 until end of turn."},
				{Label: "Equip {2}", Equip: true},
			},
		},
		rowsFixtureAngel: {
			PrintedKeywords: []string{"flying"},
			Static: []game.StaticAbility{{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "flying")
				},
				Keywords: []string{"flying"},
			}},
		},
		rowsFixtureMutation: {Static: []game.StaticAbility{{
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.InstanceID != source.InstanceID
			},
			Apply: func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}}},
		rowsFixtureRite: {Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			GrantAbilities: []string{rowsGrantTrigger},
		}}},
		game.GrantKey(rowsGrantTrigger): {
			Triggered: []game.TriggeredAbility{{Key: "Rite — when this creature dies, draw a card", Watches: []game.EventKind{game.EventLTB}}},
			GrantText: "When this creature dies, draw a card.",
		},
	}
	prev := game.CatalogLookup
	game.CatalogLookup = func(key string) *game.CardDef {
		if d, ok := fixtures[key]; ok {
			return d
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	t.Cleanup(func() { game.CatalogLookup = prev })
}

func rowsFixtureCard(name, typeLine, oracle string, owner uuid.UUID) game.Card {
	c := game.NewCard(name, owner)
	c.TypeLine = typeLine
	c.OracleID = oracle
	return c
}

func findCardView(t *testing.T, zone ZoneView, id uuid.UUID) *CardView {
	t.Helper()
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID == id.String() {
			return &zone.Cards[i]
		}
	}
	return nil
}

func TestAbilityRowsOnTheBattlefieldForEveryViewer(t *testing.T) {
	stubAbilityRowFixtures(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	angel := rowsFixtureCard("Fixture Angel", "Creature — Angel", rowsFixtureAngel, me.ID)
	bear := rowsFixtureCard("Grizzly Bears", "Creature — Bear", "rows-fixture-uncatalogued", me.ID)
	g.WithWriteLock(func() {
		for _, c := range []game.Card{titan, angel, bear} {
			c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
			g.Battlefield.PushTop(c)
		}
	})
	g.BumpLayerVersionForTest()

	want := []AbilityRowView{
		{Kind: "triggered", Label: "3 damage divided among one, two, or three targets"},
		{Kind: "static", Label: "Power/toughness effect"},
		{Kind: "activated", Label: "{R}: Fixture Titan gets +1/+0 until end of turn."},
	}
	v := ViewOfGame(g)
	for _, viewer := range []struct{ name, id string }{
		{"controller", me.ID.String()}, {"opponent", opp.ID.String()}, {"spectator", SpectatorViewerID}, {"admin", ""},
	} {
		t.Run(viewer.name, func(t *testing.T) {
			fv := FilterViewFor(v, viewer.id)
			got := findCardView(t, fv.Battlefield, titan.InstanceID)
			if got == nil {
				t.Fatal("titan missing")
			}
			if len(got.AbilityRows) != len(want) {
				t.Fatalf("titan ability_rows = %+v, want %+v (equip is a keyword ability and is left out)", got.AbilityRows, want)
			}
			for i := range want {
				if got.AbilityRows[i] != want[i] {
					t.Errorf("row %d = %+v, want %+v", i, got.AbilityRows[i], want[i])
				}
			}
			if a := findCardView(t, fv.Battlefield, angel.InstanceID); a == nil || a.AbilityRows != nil {
				t.Errorf("a keyword-only card carries ability_rows %+v; its keyword chips are the whole story", a)
			}
			b := findCardView(t, fv.Battlefield, bear.InstanceID)
			if b == nil || b.AbilityRows != nil {
				t.Errorf("an uncatalogued card carries ability_rows %+v", b)
			}
		})
	}

	// On the wire: the key, the shape, and omitempty.
	raw, err := json.Marshal(FilterViewFor(v, me.ID.String()))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"ability_rows":[{"kind":"triggered","label":"3 damage divided among one, two, or three targets"}`) {
		t.Errorf("ability_rows not on the wire in the expected shape")
	}
	if strings.Count(string(raw), `"ability_rows"`) != 1 {
		t.Errorf("ability_rows sent for a card that has none (want omitempty)")
	}
}

// Current abilities, not printed: a "loses all abilities" effect takes
// the titan's own rows, and a layer-6 grant is listed on the creature
// it is granted to.
func TestAbilityRowsFollowRemovalAndGrants(t *testing.T) {
	stubAbilityRowFixtures(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	titan := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	bear := rowsFixtureCard("Grizzly Bears", "Creature — Bear", "rows-fixture-uncatalogued", me.ID)
	rite := rowsFixtureCard("Rite", "Enchantment", rowsFixtureRite, me.ID)
	mutation := rowsFixtureCard("Mutation", "Enchantment", rowsFixtureMutation, opp.ID)
	push := func(cs ...game.Card) {
		g.WithWriteLock(func() {
			for _, c := range cs {
				c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
				g.Battlefield.PushTop(c)
			}
		})
		g.BumpLayerVersionForTest()
	}
	push(titan, bear, rite)

	fv := FilterViewFor(ViewOfGame(g), opp.ID.String())
	b := findCardView(t, fv.Battlefield, bear.InstanceID)
	if b == nil || len(b.AbilityRows) != 1 || b.AbilityRows[0] != (AbilityRowView{Kind: "triggered", Label: "Rite — when this creature dies, draw a card"}) {
		t.Fatalf("the bear under the Rite lists %+v, want the granted trigger", b)
	}
	if ti := findCardView(t, fv.Battlefield, titan.InstanceID); ti == nil || len(ti.AbilityRows) != 4 {
		t.Fatalf("the titan under the Rite lists %+v, want its three rows and the granted trigger", ti)
	}

	// The Mutation arrives later: every creature loses all abilities,
	// and a removal with a later timestamp takes the Rite's grant too.
	push(mutation)
	fv = FilterViewFor(ViewOfGame(g), me.ID.String())
	for _, id := range []uuid.UUID{titan.InstanceID, bear.InstanceID} {
		if c := findCardView(t, fv.Battlefield, id); c == nil || c.AbilityRows != nil {
			t.Errorf("a creature that lost all abilities lists %+v, want none", c)
		}
	}
	if r := findCardView(t, fv.Battlefield, rite.InstanceID); r == nil || len(r.AbilityRows) != 1 || r.AbilityRows[0].Label != "Grants abilities" {
		t.Errorf("the Rite itself lists %+v, want its grant static", r)
	}
}

// The field is filtered exactly like the card's other characteristics:
// a hand card reaches its owner and nobody else until it is revealed, a
// face-down permanent lists nothing to anyone (CR 708.2 for its
// controller, the non-knower redaction for everyone else), and a
// graveyard card — never a tile — carries none at all.
func TestAbilityRowsAreHiddenWithTheCard(t *testing.T) {
	stubAbilityRowFixtures(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	inHand := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	revealed := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	faceDown := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	inGrave := rowsFixtureCard("Fixture Titan", "Creature — Giant", rowsFixtureTitan, me.ID)
	g.WithWriteLock(func() {
		inHand.KnownBy = map[uuid.UUID]bool{me.ID: true}
		revealed.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		me.Hand.PushTop(inHand)
		me.Hand.PushTop(revealed)
		faceDown.FaceDown = true
		faceDown.FaceDownKind = game.FaceDownMorphed
		faceDown.KnownBy = map[uuid.UUID]bool{me.ID: true}
		g.Battlefield.PushTop(faceDown)
		inGrave.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
		me.Graveyard.PushTop(inGrave)
	})
	g.BumpLayerVersionForTest()
	v := ViewOfGame(g)

	mine := FilterViewFor(v, me.ID.String())
	var myHand ZoneView
	for _, s := range mine.Seats {
		if s.ID == me.ID.String() {
			myHand = s.Hand
		}
	}
	if c := findCardView(t, myHand, inHand.InstanceID); c == nil || len(c.AbilityRows) != 3 {
		t.Errorf("the owner's hand card lists %+v, want its three rows", c)
	}

	theirs := FilterViewFor(v, opp.ID.String())
	for _, s := range theirs.Seats {
		if s.ID != me.ID.String() {
			continue
		}
		if c := findCardView(t, s.Hand, inHand.InstanceID); c != nil {
			t.Errorf("an unrevealed hand card reached the opponent: %+v", c)
		}
		if c := findCardView(t, s.Hand, revealed.InstanceID); c == nil || len(c.AbilityRows) != 3 {
			t.Errorf("a revealed hand card lists %+v to the opponent, want its rows (they know it)", c)
		}
		if c := findCardView(t, s.Graveyard, inGrave.InstanceID); c == nil || c.AbilityRows != nil {
			t.Errorf("a graveyard card carries ability_rows %+v; a graveyard card is never a tile", c)
		}
	}

	for _, viewer := range []struct{ name, id string }{
		{"controller", me.ID.String()}, {"opponent", opp.ID.String()}, {"spectator", SpectatorViewerID},
	} {
		fv := FilterViewFor(v, viewer.id)
		c := findCardView(t, fv.Battlefield, faceDown.InstanceID)
		if c == nil {
			t.Fatalf("%s: the face-down permanent is missing", viewer.name)
		}
		if c.AbilityRows != nil {
			t.Errorf("%s: a face-down permanent lists %+v", viewer.name, c.AbilityRows)
		}
	}
	// The redaction itself, for a card whose rows were stamped: the
	// zone × viewer table in face_down_view_test.go covers every zone,
	// and this is the one-line version of it.
	stamped := CardView{InstanceID: "x", AbilityRows: []AbilityRowView{{Kind: "static", Label: "Draw a card"}}}
	if got := redactCardForViewer(stamped, false); got.AbilityRows != nil {
		t.Errorf("a non-knower keeps ability_rows %+v", got.AbilityRows)
	}
	if got := redactCardForViewer(stamped, true); len(got.AbilityRows) != 1 {
		t.Errorf("a knower lost ability_rows")
	}
}
