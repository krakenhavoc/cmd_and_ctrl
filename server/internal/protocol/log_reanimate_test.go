package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_reanimate_test.go — #4446, ADR 0209 §3: the client's reanimate cue
// ("a card returns from a graveyard to the battlefield or a hand") reads
// what the public log already says. Nothing on the wire is new; this
// pins, for every viewer, an opponent included, that a card leaving a
// graveyard for the battlefield (Reanimate, unearth, CR 702.84a, a land
// put back) or a hand (Regrowth, Eternal Witness) is a `zone` entry with
// `card_id`, `old_zone` "graveyard", `new_zone` where it went and a
// batch:
//
//   - a creature returned under another player's control: `seat` is
//     the new controller, not the graveyard's owner, so the client reads
//     whose graveyard it rose from off the permanent's `owner`;
//   - several cards put onto the battlefield together (Splendid
//     Reclamation, Living Death) share one batch, one entry each, which
//     is how the cue plays its sound once;
//   - a card returned to its owner's hand (CR 400.3) keeps its card_id
//     for a viewer who cannot see that hand: its old zone was public;
//   - a token copy of a card in a graveyard is a `token` entry and the
//     card stays in the graveyard, so nothing about it reads as a return.

// zonesFromGraveyard is every zone entry viewer sees that left a graveyard.
func zonesFromGraveyard(g *game.Game, viewer string) []LogEvent {
	var out []LogEvent
	for _, e := range logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogZone) {
		if e.OldZone == string(game.ZoneGraveyard) {
			out = append(out, e)
		}
	}
	return out
}

// buryCard puts a new card named name, owned by p, on top of p's graveyard.
func buryCard(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// assertReturns checks that every viewer sees one zone entry out of a
// graveyard per card in ids, in order, in one non-zero batch, with the
// wire fields the cue reads.
func assertReturns(t *testing.T, g *game.Game, ids []uuid.UUID, to game.ZoneKind, seat int) {
	t.Helper()
	for _, viewer := range allViewers(g) {
		got := zonesFromGraveyard(g, viewer)
		if len(got) != len(ids) {
			t.Fatalf("viewer %q: %d zone entries out of a graveyard, want %d: %+v", viewer, len(got), len(ids), got)
		}
		for i, e := range got {
			if e.Batch == 0 || e.Batch != got[0].Batch {
				t.Errorf("viewer %q: entry %d batch %d, want one shared non-zero batch (%d)", viewer, i, e.Batch, got[0].Batch)
			}
			raw := wireFields(t, e)
			want := map[string]any{
				"kind":     "zone",
				"card_id":  ids[i].String(),
				"old_zone": "graveyard",
				"new_zone": string(to),
				"seat":     float64(seat),
			}
			for key, w := range want {
				if raw[key] != w {
					t.Errorf("viewer %q: entry %d wire field %q = %v, want %v (entry %v)", viewer, i, key, raw[key], w, raw)
				}
			}
			assertNoUUID(t, e.Text)
		}
	}
}

func TestACardReturnedFromAGraveyardIsAZoneLineForEveryViewer(t *testing.T) {
	t.Run("a creature returned under another player's control", func(t *testing.T) {
		g := keptGame(t, 0)
		me, opp := g.Seats[0], g.Seats[1]
		id := buryCard(g, opp, "Their Bear", "Creature — Bear")
		var err error
		g.WithWriteLock(func() {
			err = g.ReturnFromGraveyardUnderControlForEffect(id, game.ZoneBattlefield, me.ID)
		})
		if err != nil {
			t.Fatalf("ReturnFromGraveyardUnderControlForEffect: %v", err)
		}
		// `seat` is who returned it, the new controller.
		assertReturns(t, g, []uuid.UUID{id}, game.ZoneBattlefield, me.Seat)
		// Whose graveyard it rose from is the permanent's owner, on the
		// board every viewer sees.
		for _, viewer := range allViewers(g) {
			var owner, controller string
			for _, c := range ViewOfGameFor(g, viewer).Battlefield.Cards {
				if c.InstanceID == id.String() {
					owner, controller = c.Owner, c.Controller
				}
			}
			if owner != opp.ID.String() || controller != me.ID.String() {
				t.Errorf("viewer %q: owner %q controller %q, want %q and %q", viewer, owner, controller, opp.ID, me.ID)
			}
		}
	})

	t.Run("several cards put onto the battlefield together", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		a := buryCard(g, me, "Old Forest", "Land — Forest")
		b := buryCard(g, me, "Old Swamp", "Land — Swamp")
		var err error
		g.WithWriteLock(func() {
			err = g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{
				{CardID: a, From: game.ZoneGraveyard},
				{CardID: b, From: game.ZoneGraveyard},
			}, game.ZoneEntryOptions{}, game.KeyedTail{})
		})
		if err != nil {
			t.Fatalf("PutOntoBattlefieldTogetherKeyedForEffect: %v", err)
		}
		assertReturns(t, g, []uuid.UUID{a, b}, game.ZoneBattlefield, me.Seat)
	})

	t.Run("a card returned to its owner's hand", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		id := buryCard(g, me, "Regrown Rites", "Sorcery")
		var err error
		g.WithWriteLock(func() { err = g.ReturnFromGraveyardForEffect(id, game.ZoneHand) })
		if err != nil {
			t.Fatalf("ReturnFromGraveyardForEffect: %v", err)
		}
		// The opponent cannot see the hand, and still gets the card_id:
		// the card's old zone was public.
		assertReturns(t, g, []uuid.UUID{id}, game.ZoneHand, me.Seat)
	})

	t.Run("a token copy of a card in a graveyard is a token entry, not a return", func(t *testing.T) {
		g := keptGame(t, 0)
		me := g.Seats[0]
		orig := buryCard(g, me, "Copied Bear", "Creature — Bear")
		var made []uuid.UUID
		var err error
		g.WithWriteLock(func() {
			tmpl, ok := effects.TokenCopyTemplate(g, orig)
			if !ok {
				t.Fatal("TokenCopyTemplate: no such card")
			}
			made, err = g.CreateTokensForEffect(me.ID, tmpl, 1, game.TokenEntryOptions{})
		})
		if err != nil || len(made) != 1 {
			t.Fatalf("CreateTokensForEffect: %v, %d made", err, len(made))
		}
		for _, viewer := range allViewers(g) {
			if got := zonesFromGraveyard(g, viewer); len(got) != 0 {
				t.Errorf("viewer %q: %d zone entries out of a graveyard, want none: %+v", viewer, len(got), got)
			}
			tokens := logEntriesOfKind(ViewOfGameFor(g, viewer).Log, LogToken)
			if len(tokens) != 1 || tokens[0].CardID != made[0].String() {
				t.Errorf("viewer %q: token entries %+v, want one naming the copy %s", viewer, tokens, made[0])
			}
		}
		g.ReadSnapshot(func() {
			if !me.Graveyard.Contains(orig) {
				t.Error("the copied card left the graveyard")
			}
		})
	})
}
