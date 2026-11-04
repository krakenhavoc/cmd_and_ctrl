package protocol

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// log_commander_test.go — #4443, ADR 0209 §3: the client's
// commander_cast and commander_return cues read what the public log
// already says about a commander. Nothing on the wire is new; this pins
// that every viewer, an opponent included, gets:
//
//   - a `cast` entry naming the commander, with `old_zone` "command",
//     when it is cast from the command zone (CR 903.8);
//   - a `zone` entry naming it, with `new_zone` "command" and the zone it
//     left in `old_zone`, when it goes home: from its graveyard by the
//     state-based action (CR 903.9a), and from the battlefield instead of
//     its owner's hand by the replacement (CR 903.9b);
//
// each with a batch, which is how the cue plays its sound once.

// entriesNaming is the log `viewer` sees, cut to the entries about id.
func entriesNaming(g *game.Game, viewer *game.Player, id uuid.UUID) []LogEvent {
	var out []LogEvent
	for _, e := range ViewOfGameFor(g, viewer.ID.String()).Log {
		if e.CardID == id.String() {
			out = append(out, e)
		}
	}
	return out
}

// wireFields is an entry as JSON keys, the names the client reads.
func wireFields(t *testing.T, e LogEvent) map[string]any {
	t.Helper()
	buf, err := json.Marshal(e)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(buf, &raw); err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestACommanderCastFromTheCommandZoneIsLoggedWithItsOrigin(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	advanceTo(t, g, game.StepPrecombatMain)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1-g.Turn.ActiveSeat]
	var id uuid.UUID
	g.WithWriteLock(func() { id = seatTaxedCommander(g, active, "", 0) })

	if err := g.CastSpell(active.ID, id, game.CastSpellParams{FromZone: "command"}); err != nil {
		t.Fatalf("cast the commander from the command zone: %v", err)
	}
	var casts []LogEvent
	for _, e := range entriesNaming(g, opp, id) {
		if e.Kind == LogCast {
			casts = append(casts, e)
		}
	}
	if len(casts) != 1 {
		t.Fatalf("the opponent sees %d cast entries for the commander, want 1", len(casts))
	}
	if casts[0].Batch == 0 {
		t.Error("the cast entry has no batch")
	}
	raw := wireFields(t, casts[0])
	for key, want := range map[string]any{"kind": "cast", "card_id": id.String(), "old_zone": "command"} {
		if raw[key] != want {
			t.Errorf("wire field %q = %v, want %v (entry %v)", key, raw[key], want, raw)
		}
	}
}

func TestACommanderGoingHomeIsLoggedAsAMoveIntoTheCommandZone(t *testing.T) {
	// A commander of seat 0's on the battlefield, public to every seat.
	setup := func(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
		t.Helper()
		g := buildActiveGame(t)
		for _, p := range g.Seats {
			if err := g.KeepHand(p.ID); err != nil {
				t.Fatalf("KeepHand: %v", err)
			}
		}
		owner, opp := g.Seats[0], g.Seats[1]
		id := uuid.New()
		g.WithWriteLock(func() {
			g.Battlefield.PushTop(game.Card{
				InstanceID:  id,
				Name:        "Homing Commander",
				TypeLine:    "Legendary Creature — Elf Druid",
				Power:       2,
				Toughness:   2,
				Owner:       owner.ID,
				Controller:  owner.ID,
				IsCommander: true,
				KnownBy:     map[uuid.UUID]bool{owner.ID: true, opp.ID: true},
			})
		})
		return g, owner, opp, id
	}
	prompt := func(t *testing.T, g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
		t.Helper()
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == kind {
				return c
			}
		}
		t.Fatalf("no %s prompt (pending: %+v)", kind, g.PendingChoices)
		return nil
	}
	homeEntry := func(t *testing.T, entries []LogEvent, from string) LogEvent {
		t.Helper()
		for _, e := range entries {
			if e.Kind == LogZone && e.NewZone == string(game.ZoneCommand) {
				if e.OldZone != from {
					t.Errorf("the move home left %q, want %q", e.OldZone, from)
				}
				if e.Batch == 0 {
					t.Error("the move home has no batch")
				}
				raw := wireFields(t, e)
				for key, want := range map[string]any{"kind": "zone", "new_zone": "command", "old_zone": from} {
					if raw[key] != want {
						t.Errorf("wire field %q = %v, want %v (entry %v)", key, raw[key], want, raw)
					}
				}
				return e
			}
		}
		t.Fatalf("no zone entry into the command zone: %+v", entries)
		return LogEvent{}
	}

	t.Run("from its graveyard, by the state-based action (CR 903.9a)", func(t *testing.T) {
		g, owner, opp, id := setup(t)
		var err error
		g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(id) })
		if err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
		g.RunStateChecksForTest()
		p := prompt(t, g, game.PendingChoiceCommanderReturn)
		if err := g.ResolveCommanderReturn(p.ID, owner.ID, true); err != nil {
			t.Fatalf("ResolveCommanderReturn: %v", err)
		}
		entries := entriesNaming(g, opp, id)
		// It died first, its own cue, then went home: two entries.
		died := false
		for _, e := range entries {
			if e.Kind == LogZone && e.OldZone == string(game.ZoneBattlefield) && e.NewZone == string(game.ZoneGraveyard) {
				died = true
			}
		}
		if !died {
			t.Errorf("no zone entry battlefield → graveyard before the return: %+v", entries)
		}
		homeEntry(t, entries, string(game.ZoneGraveyard))
	})

	t.Run("from the battlefield, instead of its owner's hand (CR 903.9b)", func(t *testing.T) {
		g, owner, opp, id := setup(t)
		var err error
		g.WithWriteLock(func() { err = g.BounceToHandForEffect(id) })
		if err != nil {
			t.Fatalf("BounceToHandForEffect: %v", err)
		}
		p := prompt(t, g, game.PendingChoiceOptionalReplacement)
		if p.Chooser != owner.ID {
			t.Fatalf("the replacement asks %s, want the owner %s", p.Chooser, owner.ID)
		}
		if err := g.ResolveOptionalReplacement(p.ID, owner.ID, true); err != nil {
			t.Fatalf("ResolveOptionalReplacement: %v", err)
		}
		entries := entriesNaming(g, opp, id)
		for _, e := range entries {
			if e.Kind == LogZone && e.NewZone == string(game.ZoneHand) {
				t.Errorf("a zone entry into the hand as well: %+v", e)
			}
		}
		homeEntry(t, entries, string(game.ZoneBattlefield))
	})
}
