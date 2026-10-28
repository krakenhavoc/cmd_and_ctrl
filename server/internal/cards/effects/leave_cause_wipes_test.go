package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// leave_cause_wipes_test.go — #4395, ADR 0209 §3: one board wipe is one
// occurrence. Every creature it takes off the battlefield carries the
// same Batch and the same LeaveSource (the wipe), and a cause that says
// whether an effect destroyed it or a state-based action did.
//
// CR 701.8b: a permanent is destroyed by an effect that says "destroy" or
// by the state-based actions for lethal damage (704.5g) and deathtouch
// (704.5h). The cue the client plays for the first is "destroy"; the
// second, like zero toughness (704.5f), is a death, so "sba".

func TestBoardWipeExitsShareBatchSourceAndCause(t *testing.T) {
	cases := []struct {
		name  string
		cast  func(t *testing.T, g *game.Game) uuid.UUID
		cause game.LeaveCause
	}{
		{
			// An effect that says "destroy" (CR 701.8a).
			name: "Damnation destroys",
			cast: func(t *testing.T, g *game.Game) uuid.UUID {
				return castCatalogSpell(t, g, "Damnation", "Sorcery",
					"d57a8f0b-7989-4db5-8756-6f2690097252", nil)
			},
			cause: game.LeaveCauseDestroy,
		},
		{
			// 13 damage marks lethal damage; the creatures die to CR 704.5g.
			name: "Blasphemous Act deals lethal damage",
			cast: func(t *testing.T, g *game.Game) uuid.UUID {
				return castCatalogSpell(t, g, "Blasphemous Act", "Sorcery", blasphemousActOracle, nil)
			},
			cause: game.LeaveCauseSBA,
		},
		{
			// -3/-3 leaves a 2/2 with toughness 0 or less: CR 704.5f.
			name: "Toxic Deluge shrinks",
			cast: func(t *testing.T, g *game.Game) uuid.UUID {
				return castWipe(t, g, "Toxic Deluge", "Sorcery", toxicDelugeOracle,
					game.CastSpellParams{XValue: 3})
			},
			cause: game.LeaveCauseSBA,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			var victims []uuid.UUID
			for i, owner := range []uuid.UUID{g.Seats[0].ID, g.Seats[0].ID, g.Seats[1].ID, g.Seats[1].ID} {
				victims = append(victims, pushWipeCreature(g, owner, "Bear "+string(rune('A'+i)), "Creature — Bear", 2, 2))
			}
			// A noncreature stays, and must not appear among the exits.
			rock := seedPermanentFor(g, g.Seats[1].ID, "Sol Ring", "Artifact")

			spell := tc.cast(t, g)
			passPriorityAroundTable(t, g)

			isVictim := map[uuid.UUID]bool{}
			for _, v := range victims {
				isVictim[v] = true
			}
			var exits []game.Event
			for _, ev := range g.Events {
				if ev.Kind == game.EventZoneMove && ev.OldZone == game.ZoneBattlefield {
					if ev.CardID == rock {
						t.Fatalf("the wipe moved a noncreature")
					}
					if isVictim[ev.CardID] {
						exits = append(exits, ev)
					}
				}
			}
			if len(exits) != len(victims) {
				t.Fatalf("%d creatures left the battlefield, want %d", len(exits), len(victims))
			}
			for _, ev := range exits {
				if ev.Batch != exits[0].Batch || ev.Batch == 0 {
					t.Errorf("batch %d, want every exit in the same non-zero batch (%d)", ev.Batch, exits[0].Batch)
				}
				if ev.LeaveSource != spell {
					t.Errorf("source %s, want the wipe %s", ev.LeaveSource, spell)
				}
				if ev.LeaveCause != tc.cause {
					t.Errorf("cause %q, want %q", ev.LeaveCause, tc.cause)
				}
				if ev.NewZone != game.ZoneGraveyard {
					t.Errorf("went to %s, want graveyard", ev.NewZone)
				}
			}

			// And on the wire, for every seat: the same three facts.
			isVictimWire := map[string]bool{}
			for _, v := range victims {
				isVictimWire[v.String()] = true
			}
			for _, seat := range g.Seats {
				wire := 0
				for _, e := range protocol.ViewOfGameFor(g, seat.ID.String()).Log {
					if e.Kind != protocol.LogZone || !isVictimWire[e.CardID] {
						continue
					}
					wire++
					if e.Batch != exits[0].Batch || e.Source != spell.String() || e.Cause != string(tc.cause) {
						t.Errorf("seat %s sees (batch %d, source %s, cause %q), want (%d, %s, %q)",
							seat.Name, e.Batch, e.Source, e.Cause, exits[0].Batch, spell, tc.cause)
					}
				}
				if wire != len(victims) {
					t.Errorf("seat %s sees %d exits in the public log, want %d", seat.Name, wire, len(victims))
				}
			}
		})
	}
}
