package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_total_mana_value_view_test.go — #5401. "Exile one or more other
// artifacts you control with total mana value X" ships the
// exile_permanent_total_mana_value_x flag and demands_x (so the
// activation flow knows an X is owed, which the picked artifacts then
// answer), the other artifacts the controller could exile as an open
// count, and a refusal reason when there is nothing to exile.

func seatExileTotalSource(g *game.Game, owner uuid.UUID) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Exile Total Source", TypeLine: "Artifact",
		OracleID: "00000000-0000-0000-0000-000000005401",
		Owner:    owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Exile one or more other artifacts you control with total mana value X: Draw a card.",
			Cost: game.AbilityCost{ExilePermanents: &game.ExilePermanentsCost{
				Count: 1, OrMore: true, CardType: "artifact", ExcludeSource: true, TotalManaValueX: true,
				Label: "one or more other artifacts you control with total mana value X",
			}},
		}},
	})
	return id
}

func TestExileTotalManaValueViewStampsTheFlagAndTheOpenCount(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src, relic uuid.UUID
	g.WithWriteLock(func() {
		src = seatExileTotalSource(g, me.ID)
		relic = uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: relic, Name: "Relic", TypeLine: "Artifact", ManaCost: "{2}",
			Owner: me.ID, Controller: me.ID})
	})
	g.BumpLayerVersionForTest()

	row := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if !row.ExilePermanentTotalManaValueX || !row.DemandsX {
		t.Errorf("total_mana_value_x=%v demands_x=%v, want both", row.ExilePermanentTotalManaValueX, row.DemandsX)
	}
	opts := row.ExilePermanentOptions
	if opts == nil || opts.Min != 1 || opts.Max != 0 || len(opts.Cards) != 1 || opts.Cards[0] != relic.String() {
		t.Errorf("options = %+v, want the other artifact alone, min 1 and no ceiling", opts)
	}
	if row.CantActivate != "" {
		t.Errorf("cant_activate = %q with an artifact to exile", row.CantActivate)
	}
}

func TestExileTotalManaValueViewSaysWhyWithNothingToExile(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	var src uuid.UUID
	g.WithWriteLock(func() { src = seatExileTotalSource(g, me.ID) })
	g.BumpLayerVersionForTest()

	row := controllerFrameCard(t, g, src).ActivatedAbilities[0]
	if row.CantActivate == "" {
		t.Error("no cant_activate with nothing but the source to exile")
	}
}
