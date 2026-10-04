package protocol

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// paused_exit_cast_view_test.go — #1474. A commander an effect is
// shuffling out of its owner's graveyard into their library waits there
// while its owner answers CR 903.9b, and CastSpell refuses to cast or
// play it until the answer is in. (Bojuka Bog's exile was the vehicle
// before ADR 0115; an exiled commander no longer pauses.) `castable_here` is the zone browser's cast button,
// so it has to say the same thing: off while the exit is paused, and
// the bot enumerator must not offer the move either.

// offersCastOf reports whether the seat's legal moves include a cast or
// land play of `id`.
func offersCastOf(g *game.Game, seat, id uuid.UUID) bool {
	for _, m := range legal.EnumerateLocked(g, seat, legal.Options{}) {
		if m.Type == legal.TypeCastSpell && m.Source == id {
			return true
		}
	}
	return false
}

func TestCastableHereIsClearedWhileTheCardsExitIsPaused(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
	}{
		{"a Gravecrawler-shaped creature", "Legendary Creature — Zombie", "{B}"},
		{"a land its own text lets you play from the graveyard", "Legendary Land", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			const oracle = "test-paused-view-cast"
			g := buildActiveGame(t)
			for _, p := range g.Seats {
				if err := g.KeepHand(p.ID); err != nil {
					t.Fatalf("KeepHand: %v", err)
				}
			}
			castWindowOpen(t, g)
			me := g.Seats[g.Turn.ActiveSeat]
			me.Graveyard.Cards = nil
			me.ManaPool.AddMana(game.ManaToken{Color: "B"})
			withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})

			cmdr := publicCard(g, me.ID, "Test Commander", tc.typeLine, tc.cost, oracle)
			cmdr.IsCommander = true
			me.Graveyard.PushTop(cmdr)
			id := cmdr.InstanceID

			// Premise: before the exile, the view and the enumerator
			// both offer it.
			if c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[g.Turn.ActiveSeat].Graveyard, id); !c.CastableHere {
				t.Fatal("premise: castable_here is clear before anything paused the card")
			}
			if !offersCastOf(g, me.ID, id) {
				t.Fatal("premise: the enumerator does not offer the card before anything paused it")
			}

			if err := g.TuckToLibraryForEffect(id, false); err != nil {
				t.Fatalf("TuckToLibraryForEffect: %v", err)
			}
			if len(g.PendingChoices) != 1 || !me.Graveyard.Contains(id) {
				t.Fatalf("premise: the tuck did not pause on the owner's CR 903.9b prompt (%d pending)", len(g.PendingChoices))
			}

			c := cardInSeatZone(t, ViewOfGameFor(g, me.ID.String()).Seats[g.Turn.ActiveSeat].Graveyard, id)
			if c.CastableHere {
				t.Error("castable_here is set on a card whose exit is paused")
			}
			if offersCastOf(g, me.ID, id) {
				t.Error("the enumerator offers a card whose exit is paused")
			}
			// …and the engine refuses it, which is what the bit reports.
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard"}); !errors.Is(err, game.ErrChoicePending) {
				t.Errorf("CastSpell on the paused card: err = %v, want ErrChoicePending", err)
			}
		})
	}
}
