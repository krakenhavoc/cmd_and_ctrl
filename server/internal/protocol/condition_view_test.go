package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// condition_view_test.go — the wire half of #743. condition_unmet is
// present on an activated or mana ability exactly while its activation
// condition is false, evaluated with the permanent's controller as
// "you", and never on an ability that has no condition.

func seatConditionPermanent(g *game.Game, open *bool) (uuid.UUID, uuid.UUID) {
	owner := g.Seats[0].ID
	id := uuid.New()
	// The closure also checks it is asked about this permanent with its
	// controller as "you"; any other arguments read as false.
	cond := func(_ *game.Game, controller, source uuid.UUID) bool {
		return *open && source == id && controller == owner
	}
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Gate",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-0000000000cc",
		Owner:      owner,
		Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{T}: ungated", Cost: game.AbilityCost{Tap: true}},
			{Label: "{T}: gated. Activate only if …", Cost: game.AbilityCost{Tap: true}, Condition: cond},
		},
		ManaAbilities: []game.ManaAbilityShape{
			{TapCost: true, Produced: "{C}", Label: "Add {C}"},
			{TapCost: true, Produced: "{C}{C}", Label: "Add {C}{C}. Activate only if …", Condition: cond},
		},
	})
	return id, owner
}

func conditionCardView(t *testing.T, g *game.Game, id uuid.UUID) CardView {
	t.Helper()
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			return v
		}
	}
	t.Fatal("permanent missing from the view")
	return CardView{}
}

func TestConditionUnmetTracksTheCondition(t *testing.T) {
	g := buildActiveGame(t)
	open := false
	id, _ := seatConditionPermanent(g, &open)

	c := conditionCardView(t, g, id)
	if len(c.ActivatedAbilities) != 2 || len(c.ManaAbilities) != 2 {
		t.Fatalf("got %d activated / %d mana abilities, want 2 / 2", len(c.ActivatedAbilities), len(c.ManaAbilities))
	}
	if c.ActivatedAbilities[0].ConditionUnmet || c.ManaAbilities[0].ConditionUnmet {
		t.Error("an ability with no condition must never carry condition_unmet")
	}
	if !c.ActivatedAbilities[1].ConditionUnmet {
		t.Error("activated ability: condition false → condition_unmet")
	}
	if !c.ManaAbilities[1].ConditionUnmet {
		t.Error("mana ability: condition false → condition_unmet (owner decision on #743)")
	}

	open = true
	c = conditionCardView(t, g, id)
	if c.ActivatedAbilities[1].ConditionUnmet || c.ManaAbilities[1].ConditionUnmet {
		t.Error("condition true → condition_unmet absent")
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "condition_unmet") {
		t.Errorf("a met condition must leave condition_unmet off the wire: %s", raw)
	}

	open = false
	raw, _ = json.Marshal(conditionCardView(t, g, id))
	if got := strings.Count(string(raw), `"condition_unmet":true`); got != 2 {
		t.Errorf("condition_unmet on the wire %d times, want 2 (one activated, one mana): %s", got, raw)
	}
}
