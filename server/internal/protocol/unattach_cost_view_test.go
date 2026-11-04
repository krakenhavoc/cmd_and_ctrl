package protocol

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// unattach_cost_view_test.go — #4799: the wire half of "Unattach an
// Equipment from <this>" (Captain America, First Avenger). The row carries
// `unattach_label` / `unattach_options`, the engine's own candidate walk,
// and a division read from the pick carries `divide.by_unattach`, the
// amount each candidate would announce, so the client sizes the split from
// the Equipment it picked.

const capFirstAvengerOracleForView = "7d038055-035d-4698-b6b6-e2859bf2f1d1"

func seatUnattachEquipment(g *game.Game, controller, host uuid.UUID, name, cost string) uuid.UUID {
	id := uuid.New()
	c := game.Card{
		InstanceID: id, Name: name, TypeLine: "Artifact — Equipment", ManaCost: cost,
		Owner: controller, Controller: controller,
	}
	if host != uuid.Nil {
		c.AttachedTo = game.TargetRef{Kind: game.TargetCard, ID: host}
	}
	g.Battlefield.PushTop(c)
	return id
}

func TestAbilityViewCarriesUnattachOptionsAndADivisionPerPick(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	cap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cap, Name: "Captain America, First Avenger", TypeLine: "Legendary Creature — Human Soldier Hero",
		OracleID: capFirstAvengerOracleForView, Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	knife := seatUnattachEquipment(g, me, cap, "Knife", "{1}")
	hammer := seatUnattachEquipment(g, opp, cap, "Hammer", "{2}{R}{R}")
	seatUnattachEquipment(g, me, uuid.Nil, "Loose", "{3}")
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, cap).ActivatedAbilities[0]
	if ab.UnattachLabel != "an Equipment from Captain America, First Avenger" {
		t.Errorf("unattach_label = %q", ab.UnattachLabel)
	}
	if ab.UnattachOptions == nil {
		t.Fatal("unattach_options absent on an ability that prints the clause")
	}
	if ab.UnattachOptions.Min != 1 || ab.UnattachOptions.Max != 1 {
		t.Errorf("bounds %d/%d, want 1/1", ab.UnattachOptions.Min, ab.UnattachOptions.Max)
	}
	got := append([]string(nil), ab.UnattachOptions.Cards...)
	sort.Strings(got)
	want := []string{knife.String(), hammer.String()}
	sort.Strings(want)
	if !sameStrings(got, want) {
		t.Errorf("unattach_options = %v, want the two Equipment attached to him", ab.UnattachOptions.Cards)
	}
	if ab.LegalTargets == nil || ab.LegalTargets.Divide == nil {
		t.Fatal("the divided clause ships no divide")
	}
	by := ab.LegalTargets.Divide.ByUnattach
	if by[knife.String()] != 1 || by[hammer.String()] != 4 || len(by) != 2 {
		t.Errorf("divide.by_unattach = %v, want the Knife 1 and the Hammer 4", by)
	}
	raw, err := json.Marshal(ab)
	if err != nil {
		t.Fatal(err)
	}
	var wire map[string]any
	if err := json.Unmarshal(raw, &wire); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"unattach_label", "unattach_options"} {
		if _, ok := wire[k]; !ok {
			t.Errorf("wire row has no %q", k)
		}
	}
}

// With nothing attached the list is present and empty: the cost cannot
// be paid (CR 118.3), and the client greys the row on that.
func TestAbilityViewUnattachOptionsEmptyWithNothingAttached(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	cap := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: cap, Name: "Captain America, First Avenger", TypeLine: "Legendary Creature — Human Soldier Hero",
		OracleID: capFirstAvengerOracleForView, Power: 4, Toughness: 4, Owner: me, Controller: me,
	})
	seatUnattachEquipment(g, me, uuid.Nil, "Loose", "{3}")
	g.BumpLayerVersionForTest()

	ab := vehicleView(t, g, cap).ActivatedAbilities[0]
	if ab.UnattachOptions == nil || len(ab.UnattachOptions.Cards) != 0 {
		t.Errorf("unattach_options = %+v, want present and empty", ab.UnattachOptions)
	}
}
