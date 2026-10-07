package aiseat_test

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// manifest_dread_test.go — #2570 / ADR 0082's 2026-10-07 amendment.
// Manifest dread asks the controller to name one of two looked-at
// cards. A seat that owes that question is enumerated its answers and
// nothing else (#544), so both policies must be able to finish it, and
// the action must end with one manifested creature and one graveyard
// card whichever answer they pick.

func TestBotsAnswerManifestDread(t *testing.T) {
	policies := map[string]aiseat.Policy{
		"heuristic": heuristic.New(),
		"random":    aiseat.NewRandomPolicy(rand.NewPCG(7, 7)),
	}
	for name, pol := range policies {
		t.Run(name, func(t *testing.T) {
			g := newSettledTable(t, 11)
			seat := g.Seats[0]
			libBefore := seat.Library.Size()
			bfBefore := len(g.Battlefield.Cards)
			gyBefore := seat.Graveyard.Size()

			var got game.ManifestDreadResult
			g.WithWriteLock(func() {
				if err := g.ManifestDreadThenForEffect(seat.ID, uuid.Nil, func(_ *game.Game, res game.ManifestDreadResult) error {
					got = res
					return nil
				}); err != nil {
					t.Fatalf("ManifestDreadThenForEffect: %v", err)
				}
			})

			took := driveChoices(t, g, pol, seat.ID, 4)
			if len(took) != 1 {
				t.Fatalf("took %d answers, want exactly the one pick: %v", len(took), took)
			}
			if got.Manifested == uuid.Nil || got.Graveyarded == uuid.Nil {
				t.Fatalf("result = %+v, want a manifested card and a graveyard card", got)
			}
			if seat.Library.Size() != libBefore-2 {
				t.Errorf("library %d, want %d", seat.Library.Size(), libBefore-2)
			}
			if len(g.Battlefield.Cards) != bfBefore+1 || seat.Graveyard.Size() != gyBefore+1 {
				t.Errorf("battlefield +%d, graveyard +%d, want +1 each",
					len(g.Battlefield.Cards)-bfBefore, seat.Graveyard.Size()-gyBefore)
			}
		})
	}
}
