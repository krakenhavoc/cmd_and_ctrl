package protocol

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// damage_split_view_test.go — #2956, ADR 0147: a damage-assignment
// prompt carries the canonical split (legal.CanonicalDamageSplit) as
// `suggested`, each blocker's lethal damage, and whether the power
// covers lethal for all of them.
func TestDamageAssignmentCarriesTheCanonicalSplit(t *testing.T) {
	g := threeSeatsInDeclareAttackers(t)
	active := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%3]
	push := func(owner *game.Player, name string, power, tough int, kw []string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: "Creature — Test", Keywords: kw,
			Power: power, Toughness: tough, Owner: owner.ID, Controller: owner.ID,
		})
		return id
	}
	vivi := push(active, "Vivi", 44, 4, []string{"trample"})
	blockers := []uuid.UUID{
		push(def, "Four", 1, 4, nil),
		push(def, "Three", 1, 3, nil),
		push(def, "Two", 1, 2, nil),
	}
	if err := g.DeclareAttacker(vivi, def.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatal(err)
	}
	for _, b := range blockers {
		if err := g.DeclareBlocker(b, vivi); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.FinishBlocks(def.ID); err != nil {
		t.Fatal(err)
	}
	for range 8 {
		if g.Turn.Step == game.StepCombatDamage {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	var da *DamageAssignmentView
	for _, pc := range ViewOfGameFor(g, active.ID.String()).PendingChoices {
		if pc.DamageAssignment != nil {
			da = pc.DamageAssignment
		}
	}
	if da == nil || da.Suggested == nil {
		t.Fatalf("damage assignment on the wire = %+v, want a suggested split", da)
	}
	if !da.CoversLethal || !slices.Equal(da.Lethal, []int{4, 3, 2}) {
		t.Errorf("covers %v, lethal %v; want covers and [4 3 2]", da.CoversLethal, da.Lethal)
	}
	got := map[string]int{}
	for _, a := range da.Suggested.Assignments {
		got[a.BlockerID] = a.Amount
	}
	if got[blockers[0].String()] != 4 || got[blockers[1].String()] != 3 || got[blockers[2].String()] != 2 ||
		da.Suggested.TrampleToPlayer != 35 {
		t.Errorf("suggested %v, trample %d; want 4/3/2 and 35", got, da.Suggested.TrampleToPlayer)
	}
}
