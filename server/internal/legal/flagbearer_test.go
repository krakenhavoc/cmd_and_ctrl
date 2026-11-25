package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// flagbearer_test.go — #6236, the enumerator half of the Flagbearer rule
// (CR 601.2c): while an opponent's Standard Bearer is on the battlefield,
// the bot is offered its Lightning Bolt only at the Flagbearer, and every
// move it is offered is one the engine accepts.

const testFlagbearerRuleOracle = "test-legal-flagbearer-rule"

func TestEnumeratorOffersOnlyTheFlagbearerUnderTheRule(t *testing.T) {
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key == testFlagbearerRuleOracle {
			return []game.StaticAbility{{
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) { c.FlagbearerTargetRule = true },
			}}
		}
		if prev == nil {
			return nil
		}
		return prev(key)
	}
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	bolt := handCard(active, game.Card{
		Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		Colors: []string{"R"}, OracleID: oracleLightningBolt,
	})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	bear := battlefieldCard(g, opp, creature("Plain Bear", "{1}{G}", 2, 2))
	bearerCard := creature("Standard Bearer", "{W}", 1, 1)
	bearerCard.TypeLine = "Creature — Human Flagbearer"
	bearerCard.OracleID = testFlagbearerRuleOracle
	bearer := enteredCard(g, opp, bearerCard)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	dispatchAll(t, g, active.ID, moves)

	targeted := map[string]bool{}
	for _, m := range moves {
		if m.Source != bolt {
			continue
		}
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("bad Bolt params: %s", string(m.Params))
		}
		for _, tr := range p.Targets {
			targeted[tr.ID] = true
		}
	}
	if !targeted[bearer.String()] {
		t.Error("the Flagbearer was not offered as the Bolt's target")
	}
	if targeted[bear.String()] {
		t.Error("the bot was offered a Bolt at another creature while a Flagbearer could be targeted")
	}
	for _, p := range g.Seats {
		if targeted[p.ID.String()] {
			t.Error("the bot was offered a Bolt at a player while a Flagbearer could be targeted")
		}
	}
}
