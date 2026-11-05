package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_transform_test.go — #4448, ADR 0209 §3: the client's transform and
// day_night cues read what the public log already says. Nothing on the
// wire is new; this pins, for every viewer, an opponent and a spectator
// included:
//
//   - a double-faced permanent turned over in place (CR 701.27a) is one
//     `transform` entry with its `card_id`, the face it turned from in
//     `label`, and a batch, and the card's view says which face is up
//     (`active_face`) and serves both faces' art; a convert (CR 701.28a)
//     is the same entry with `converted`;
//   - night falling in the untap step (CR 731.2a, 502.2) is one
//     `day_night` entry ("night") followed by one `transform` entry per
//     werewolf it turned over (CR 702.145c), all in one batch, which is
//     how the cue plays one sound for the moment;
//   - a meld (CR 701.42a) is the two cards' exile, the melded permanent's
//     entry from exile, and one `meld` entry naming it, all in one batch,
//     with no `transform` entry (CR 712.4c: a meld card never transforms);
//     the melded permanent's view carries its two cards' art;
//   - none of these is a transform, so none writes a `transform` entry: a
//     daybound card entering at night (it enters transformed, CR
//     702.145b), a disturb cast (CR 702.146a: the back face is cast, CR
//     712.11a, and enters back face up, CR 702.146b), a face-down permanent
//     turned face up by an effect (CR 701.27b, 708.8: a `turn_face_up`
//     entry), and "exile it, then return it transformed" (a new object,
//     CR 712.14a: two `zone` entries).

// transformEntries is every `transform` entry viewer sees.
func transformEntries(g *game.Game, viewer string) []LogEvent {
	return logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogTransform)
}

// dfcCard is a transform-layout double-faced card in the shape the
// importer gives one: two faces, front up, with a printing to serve art
// from. Keywords go on the faces they belong to.
func dfcCard(owner uuid.UUID, front, back string, frontKW, backKW []string) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   uuid.NewString(),
		ScryfallID: uuid.NewString(),
		Owner:      owner,
		Controller: owner,
		Layout:     game.LayoutTransform,
		Faces: []game.Face{
			{Name: front, TypeLine: "Creature — Human Werewolf", ManaCost: "{2}{G}", Colors: []string{"G"}, Power: 2, Toughness: 3, Keywords: frontKW},
			{Name: back, TypeLine: "Creature — Werewolf", Colors: []string{"G"}, Power: 4, Toughness: 5, Keywords: backKW},
		},
	}
	c.SetFace(0)
	return c
}

// enterDFC puts c onto the battlefield from its owner's hand through the
// entry pipeline, so the zone move the day/night rules read is emitted.
func enterDFC(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	var err error
	g.WithWriteLock(func() {
		p.Hand.PushTop(c)
		id, err = g.PutFromHandOntoBattlefieldForEffect(c.InstanceID, game.HandEntryOptions{})
	})
	if err != nil || id == uuid.Nil {
		t.Fatalf("put %q onto the battlefield: id=%v err=%v", c.Name, id, err)
	}
	return id
}

// fieldCardView is the battlefield card id as viewer sees it.
func fieldCardView(t *testing.T, g *game.Game, viewer string, id uuid.UUID) CardView {
	t.Helper()
	for _, c := range ViewOfGameFor(g, viewer).Battlefield.Cards {
		if c.InstanceID == id.String() {
			return c
		}
	}
	t.Fatalf("viewer %q: %v is not on the battlefield", viewer, id)
	return CardView{}
}

// assertTransform checks one `transform` entry for id with the wire
// fields the cue reads, and the card's face as every viewer sees it.
func assertTransform(t *testing.T, g *game.Game, viewer string, e LogEvent, id uuid.UUID, from, to string) {
	t.Helper()
	raw := wireFields(t, e)
	if raw["card_id"] != id.String() || raw["label"] != from || e.Batch == 0 {
		t.Errorf("viewer %q: transform entry %v, want card_id %v, label %q and a batch", viewer, raw, id, from)
	}
	if want := from + " transformed into " + to; e.Text != want {
		t.Errorf("viewer %q: text %q, want %q", viewer, e.Text, want)
	}
	assertNoUUID(t, e.Text)
	c := fieldCardView(t, g, viewer, id)
	if c.ActiveFace != 1 || c.Name != to || len(c.Faces) != 2 {
		t.Errorf("viewer %q: card face %d %q with %d faces, want the back face %q of two", viewer, c.ActiveFace, c.Name, len(c.Faces), to)
		return
	}
	for i, f := range c.Faces {
		if f.Image == "" {
			t.Errorf("viewer %q: face %d (%q) has no image", viewer, i, f.Name)
		}
	}
}

func TestATransformIsOneTransformEntryForEveryViewer(t *testing.T) {
	t.Run("an effect turns a permanent over", func(t *testing.T) {
		g, me, _ := attachTable(t)
		id := enterDFC(t, g, me, dfcCard(me.ID, "Village Ironsmith", "Ironfang", nil, nil))
		var err error
		g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
		if err != nil {
			t.Fatalf("TransformPermanentForEffect: %v", err)
		}
		for _, viewer := range allViewers(g) {
			got := transformEntries(g, viewer)
			if len(got) != 1 {
				t.Fatalf("viewer %q: %d transform entries, want 1: %+v", viewer, len(got), got)
			}
			assertTransform(t, g, viewer, got[0], id, "Village Ironsmith", "Ironfang")
			if got[0].Converted {
				t.Errorf("viewer %q: a transform is marked converted", viewer)
			}
		}
	})

	t.Run("a convert is the same entry, marked converted", func(t *testing.T) {
		g, me, _ := attachTable(t)
		id := enterDFC(t, g, me, dfcCard(me.ID, "Front Bot", "Back Car", nil, nil))
		var err error
		g.WithWriteLock(func() { err = g.ConvertPermanentForEffect(id) })
		if err != nil {
			t.Fatalf("ConvertPermanentForEffect: %v", err)
		}
		for _, viewer := range allViewers(g) {
			got := transformEntries(g, viewer)
			if len(got) != 1 || !got[0].Converted || got[0].CardID != id.String() {
				t.Fatalf("viewer %q: transform entries %+v, want one marked converted", viewer, got)
			}
		}
	})
}

// Night falls in the untap step: the change and every werewolf it turns
// over are one batch, the change first.
func TestNightFallingIsOneBatchWithEveryWerewolfItTurnsOver(t *testing.T) {
	g, me, opp := attachTable(t)
	daybound := []string{game.KeywordDaybound}
	nightbound := []string{game.KeywordNightbound}
	mine := enterDFC(t, g, me, dfcCard(me.ID, "Fixture Human", "Fixture Wolf", daybound, nightbound))
	theirs := enterDFC(t, g, opp, dfcCard(opp.ID, "Other Human", "Other Wolf", daybound, nightbound))
	if !g.IsDay() {
		t.Fatalf("a daybound permanent entering left it %q, want day (CR 702.145d)", g.DayNightDesignation())
	}
	// The active player casts no spell this turn: at the next untap step
	// it becomes night (CR 731.2a).
	start := g.Turn.ActiveSeat
	for i := 0; g.Turn.ActiveSeat == start || g.Turn.Step == game.StepUntap; i++ {
		if i > 40 {
			t.Fatal("the turn never passed")
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if !g.IsNight() {
		t.Fatalf("a turn with no spells left it %q, want night", g.DayNightDesignation())
	}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		var nights []LogEvent
		for _, e := range logEntriesOfKind(log, LogDayNight) {
			if e.Label == "night" {
				nights = append(nights, e)
			}
		}
		if len(nights) != 1 {
			t.Fatalf("viewer %q: %d \"night\" entries, want 1: %+v", viewer, len(nights), nights)
		}
		night := nights[0]
		if night.Text != "It becomes night" || night.Batch == 0 || night.CardID != "" {
			t.Errorf("viewer %q: day_night entry %+v", viewer, night)
		}
		got := transformEntries(g, viewer)
		if len(got) != 2 {
			t.Fatalf("viewer %q: %d transform entries, want one per werewolf: %+v", viewer, len(got), got)
		}
		assertTransform(t, g, viewer, got[0], mine, "Fixture Human", "Fixture Wolf")
		assertTransform(t, g, viewer, got[1], theirs, "Other Human", "Other Wolf")
		for _, e := range got {
			if e.Batch != night.Batch {
				t.Errorf("viewer %q: a werewolf's transform is batch %d, night's %d: want one moment", viewer, e.Batch, night.Batch)
			}
			if e.Seq < night.Seq {
				t.Errorf("viewer %q: a werewolf turned over (seq %d) before night fell (seq %d)", viewer, e.Seq, night.Seq)
			}
		}
	}
}

// A daybound permanent that enters at night enters transformed: it never
// transformed, so it is its `zone` entry and nothing more.
func TestADayboundCardEnteringAtNightIsNoTransform(t *testing.T) {
	g, me, _ := attachTable(t)
	g.WithWriteLock(func() { g.BecomeNightForEffect() })
	id := enterDFC(t, g, me, dfcCard(me.ID, "Fixture Human", "Fixture Wolf",
		[]string{game.KeywordDaybound}, []string{game.KeywordNightbound}))
	for _, viewer := range allViewers(g) {
		if got := transformEntries(g, viewer); len(got) != 0 {
			t.Errorf("viewer %q: an entry at night logged a transform: %+v", viewer, got)
		}
		if c := fieldCardView(t, g, viewer, id); c.ActiveFace != 1 {
			t.Errorf("viewer %q: entered at night as face %d, want the back", viewer, c.ActiveFace)
		}
	}
}

// A meld is the two cards' exile, the melded permanent's arrival and one
// `meld` entry, in one batch, and never a transform.
func TestAMeldIsOneBatchAndNoTransform(t *testing.T) {
	g, me, opp := attachTable(t)
	print := &game.MeldPrint{
		ResultOracleID:   uuid.NewString(),
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Chittering Host", TypeLine: "Creature — Eldrazi Horror",
			Colors: []string{"B"}, Power: 5, Toughness: 6,
		},
	}
	half := func(name string) uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID: id, Name: name, OracleID: uuid.NewString(), ScryfallID: uuid.NewString(),
				TypeLine: "Creature — Rat", Power: 1, Toughness: 1,
				Owner: me.ID, Controller: me.ID, Layout: game.LayoutMeld, Meld: print,
				KnownBy: map[uuid.UUID]bool{me.ID: true, opp.ID: true},
			})
		})
		return id
	}
	a, b := half("Midnight Scavengers"), half("Graf Rats")
	var err error
	g.WithWriteLock(func() { err = g.MeldForEffect(a, b, uuid.Nil) })
	if err != nil {
		t.Fatalf("MeldForEffect: %v", err)
	}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		melds := logEntriesOfKind(log, LogMeld)
		if len(melds) != 1 || melds[0].Batch == 0 {
			t.Fatalf("viewer %q: meld entries %+v, want one with a batch", viewer, melds)
		}
		m := melds[0]
		if m.Text != "Midnight Scavengers and Graf Rats melded into Chittering Host" {
			t.Errorf("viewer %q: meld text %q", viewer, m.Text)
		}
		var exiled, entered int
		for _, z := range logEntriesOfKind(log, LogZone) {
			switch {
			case z.OldZone == "battlefield" && z.NewZone == "exile" && (z.CardID == a.String() || z.CardID == b.String()):
				exiled++
				if z.Batch != m.Batch {
					t.Errorf("viewer %q: an exile in batch %d, the meld's %d", viewer, z.Batch, m.Batch)
				}
			case z.OldZone == "exile" && z.NewZone == "battlefield" && z.CardID == m.CardID:
				entered++
				if z.Batch != m.Batch || z.Seq > m.Seq {
					t.Errorf("viewer %q: the arrival (batch %d, seq %d) and the meld (batch %d, seq %d)", viewer, z.Batch, z.Seq, m.Batch, m.Seq)
				}
			}
		}
		if exiled != 2 || entered != 1 {
			t.Errorf("viewer %q: %d exiles and %d arrivals, want 2 and 1", viewer, exiled, entered)
		}
		if got := transformEntries(g, viewer); len(got) != 0 {
			t.Errorf("viewer %q: a meld logged a transform: %+v", viewer, got)
		}
		c := fieldCardView(t, g, viewer, uuid.MustParse(m.CardID))
		if c.Name != "Chittering Host" || len(c.MeldedFrom) != 2 {
			t.Fatalf("viewer %q: melded view %q with %d cards", viewer, c.Name, len(c.MeldedFrom))
		}
		for _, part := range c.MeldedFrom {
			if part.Image == "" {
				t.Errorf("viewer %q: melded card %q has no image", viewer, part.Name)
			}
		}
	}
}

// A disturb cast puts the back face on the stack and the permanent enters
// back face up: a `cast` and a `zone` entry, and no transform.
func TestADisturbCastIsNoTransform(t *testing.T) {
	g, me, _ := attachTable(t)
	face := func(name, typeLine, cost, p, tgh string) cards.CardFace {
		return cards.CardFace{Name: name, TypeLine: typeLine, ManaCost: cost, Power: p, Toughness: tgh, Colors: []string{"U"}}
	}
	// Baithook Angler // Hook-Haunt Drifter, in the shape the bulk dump
	// ships a transform card (the catalog row's oracle ID).
	row := cards.Card{
		ID:       uuid.New(),
		Name:     "Baithook Angler // Hook-Haunt Drifter",
		Layout:   "transform",
		TypeLine: "Creature — Human Peasant // Creature — Spirit",
		OracleID: uuid.MustParse("c6bb4b41-8dae-429a-b928-ae9d39c74711"),
		CardFaces: []cards.CardFace{
			face("Baithook Angler", "Creature — Human Peasant", "{1}{U}", "2", "1"),
			face("Hook-Haunt Drifter", "Creature — Spirit", "", "1", "2"),
		},
	}
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = me.ID, me.ID
	g.WithWriteLock(func() { me.Graveyard.PushTop(c) })
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "disturb"}); err != nil {
		t.Fatalf("disturb: %v", err)
	}
	for i := 0; len(g.StackMeta) > 0; i++ {
		if i > 8 {
			t.Fatal("the disturbed spell never resolved")
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	for _, viewer := range allViewers(g) {
		log := ViewOfGameFor(g, viewer).Log
		if got := transformEntries(g, viewer); len(got) != 0 {
			t.Errorf("viewer %q: a disturb cast logged a transform: %+v", viewer, got)
		}
		var arrived bool
		for _, z := range logEntriesOfKind(log, LogZone) {
			if z.CardID == c.InstanceID.String() && z.NewZone == "battlefield" {
				arrived = true
			}
		}
		if !arrived {
			t.Errorf("viewer %q: no zone entry for the disturbed permanent", viewer)
		}
		if v := fieldCardView(t, g, viewer, c.InstanceID); v.ActiveFace != 1 || v.Name != "Hook-Haunt Drifter" {
			t.Errorf("viewer %q: entered as face %d %q, want the back face", viewer, v.ActiveFace, v.Name)
		}
	}
}

// A face-down permanent turned face up is a `turn_face_up` entry when an
// effect did it, and the special action's own line when its controller
// did (the permanent is its own source): turning over is not
// transforming (CR 701.27b), so neither writes a transform.
func TestTurningAPermanentFaceUpIsNoTransform(t *testing.T) {
	for _, byEffect := range []bool{true, false} {
		g, me, opp := attachTable(t)
		var id uuid.UUID
		var err error
		g.WithWriteLock(func() {
			me.Library.PushTop(dfcCard(me.ID, "Hidden Human", "Hidden Wolf", nil, nil))
			id, err = g.ManifestForEffect(me.ID)
		})
		if err != nil || id == uuid.Nil {
			t.Fatalf("ManifestForEffect: id=%v err=%v", id, err)
		}
		source := uuid.Nil // the permanent itself, as for the special action
		if byEffect {
			source = pushPermanent(g, me, opp, "Hauntwoods Shrieker", "Creature — Beast Horror")
		}
		var up bool
		g.WithWriteLock(func() { up = g.TurnFaceUpForEffect(source, me.ID, id) })
		if !up {
			t.Fatal("TurnFaceUpForEffect turned nothing face up")
		}
		for _, viewer := range allViewers(g) {
			if got := transformEntries(g, viewer); len(got) != 0 {
				t.Errorf("viewer %q: turning face up logged a transform: %+v", viewer, got)
			}
			ups := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogTurnFaceUp)
			if want := map[bool]int{true: 1, false: 0}[byEffect]; len(ups) != want {
				t.Errorf("viewer %q (by an effect %v): turn_face_up entries %+v, want %d", viewer, byEffect, ups, want)
			} else if byEffect && ups[0].CardID != id.String() {
				t.Errorf("viewer %q: turn_face_up entry %+v", viewer, ups[0])
			}
			if v := fieldCardView(t, g, viewer, id); v.ActiveFace != 0 {
				t.Errorf("viewer %q: turned up as face %d, want the front (CR 712.15a)", viewer, v.ActiveFace)
			}
		}
	}
}

// "Exile it, then return it transformed" is a new object (CR 400.7) that
// enters back face up (CR 712.14a): an exile and an arrival, no transform.
func TestExileAndReturnTransformedIsNoTransform(t *testing.T) {
	g, me, _ := attachTable(t)
	id := enterDFC(t, g, me, dfcCard(me.ID, "Front Mage", "Back Walker", nil, nil))
	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(id, me.ID) })
	if err != nil {
		t.Fatalf("ExileAndReturnTransformedForEffect: %v", err)
	}
	for _, viewer := range allViewers(g) {
		if got := transformEntries(g, viewer); len(got) != 0 {
			t.Errorf("viewer %q: exile and return logged a transform: %+v", viewer, got)
		}
		var out, in int
		for _, z := range logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogZone) {
			if z.OldZone == "battlefield" && z.NewZone == "exile" {
				out++
			}
			if z.OldZone == "exile" && z.NewZone == "battlefield" {
				in++
			}
		}
		if out != 1 || in != 1 {
			t.Errorf("viewer %q: %d exiles and %d arrivals, want one each", viewer, out, in)
		}
	}
}
