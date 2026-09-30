package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// phyrexian_spend_grant_moves_test.go — the enumerator half of #1589.
// Under an any-colour or any-type grant a Phyrexian symbol stays
// Phyrexian and any mana pays its mana half. The enumerator prices
// through the one pricer (game.PriceCastForEffect) and asks the same
// payability predicate the auto-tapper answers, so a stolen
// {1}{B/P}{B/P} is offered off three Mountains, is NOT offered off two,
// and the engine agrees both ways: the offered cast dispatches, and the
// same auto-tapped cast off two Mountains is refused.
func TestEnumeratorAgreesWithTheAutoTapperOnAStolenPhyrexianCard(t *testing.T) {
	for _, grant := range []struct {
		name string
		perm game.CastPermission
	}{
		{"any color", game.CastPermission{AnyColor: true, Duration: game.WhileInZoneDuration()}},
		{"any type", game.CastPermission{AnyType: true, Duration: game.WhileInZoneDuration()}},
	} {
		for _, lands := range []int{3, 2} {
			want := 0
			if lands == 3 {
				want = 1
			}
			t.Run(grant.name+"/"+map[int]string{3: "three", 2: "two"}[lands]+" Mountains", func(t *testing.T) {
				g := newTable(t)
				seat := g.Seats[g.Turn.ActiveSeat]
				victim := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
				clearHand(seat)
				advanceTo(t, g, game.StepPrecombatMain)
				for i := 0; i < lands; i++ {
					g.Battlefield.PushTop(game.Card{
						InstanceID: uuid.New(), Name: "Mountain", TypeLine: "Basic Land — Mountain",
						Owner: seat.ID, Controller: seat.ID,
					})
				}
				loot := game.NewCard("Enum Stolen Dismember", victim.ID)
				loot.TypeLine = "Sorcery"
				loot.ManaCost = "{1}{B/P}{B/P}"
				victim.Library.PushTop(loot)
				g.WithWriteLock(func() {
					if _, err := g.ExileTopWithPermissionForEffect(victim.ID, seat.ID, 1, grant.perm); err != nil {
						t.Fatalf("ExileTopWithPermissionForEffect: %v", err)
					}
				})

				// #1677: the life payments are offered as well, so the
				// agreement is about the MANA payment — phyrexian_life 0
				// — and every move offered, life or not, must dispatch.
				all := castMovesFor(legal.EnumerateFor(g, seat.ID), loot.InstanceID)
				dispatchAll(t, g, seat.ID, all)
				var got []legal.Move
				for _, m := range all {
					if phyrexianLifeOf(t, m) == 0 {
						got = append(got, m)
					}
				}
				if len(got) != want {
					t.Fatalf("offered %d mana casts off %d Mountains, want %d", len(got), lands, want)
				}
				if want > 0 {
					return
				}
				// Not offered, and the auto-tapper refuses the same cast.
				if err := g.Clone().CastSpell(seat.ID, loot.InstanceID, game.CastSpellParams{
					FromZone: "exile", Strict: true, AutoTap: true,
				}); err == nil {
					t.Error("the enumerator offered nothing, but the auto-tapped cast went through")
				}
			})
		}
	}
}
