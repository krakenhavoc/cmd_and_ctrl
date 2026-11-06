package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Crypt Ghast, Mirari's Wake
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_mana_view_test.go is #4770: a mana row's trigger_adds names
// the mana a tap would add through triggered mana abilities, from the
// auto-tap planner's judge, on the catalogued cards. `produced` stays
// what CR 106.7 says the row could produce.

func TestTriggerAddsNamesTheCatalogTriggersMana(t *testing.T) {
	const (
		cryptGhastOracle  = "a3c8d817-7949-4dae-b9f5-f9d952479270"
		mirarisWakeOracle = "852657c0-18a4-4b28-b9ae-7728acdb5044"
	)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	push := func(name, typeLine, oracle string, controller *game.Player) game.Card {
		c := game.NewCard(name, controller.ID)
		c.TypeLine, c.OracleID, c.Controller = typeLine, oracle, controller.ID
		g.Battlefield.PushTop(c)
		return c
	}
	s := push("Swamp", "Basic Land — Swamp", "", me)
	f := push("Forest", "Basic Land — Forest", "", me)
	o := push("Swamp", "Basic Land — Swamp", "", opp)

	row := func(t *testing.T, id game.Card) ManaAbilityView {
		t.Helper()
		c := conditionCardView(t, g, id.InstanceID)
		if len(c.ManaAbilities) != 1 {
			t.Fatalf("%s mana abilities = %+v, want one", id.Name, c.ManaAbilities)
		}
		return c.ManaAbilities[0]
	}
	// No triggered mana ability: nothing is stamped.
	if got := row(t, s); got.TriggerAdds != "" {
		t.Errorf("no trigger on the board, yet the Swamp's trigger_adds = %q", got.TriggerAdds)
	}

	push("Crypt Ghast", "Creature — Spirit", cryptGhastOracle, me)
	if got := row(t, s); got.TriggerAdds != "{B}" || got.Produced != "{B}" {
		t.Errorf("Swamp under Crypt Ghast: produced %q, trigger_adds %q; want {B} and {B}", got.Produced, got.TriggerAdds)
	}
	if got := row(t, f); got.TriggerAdds != "" {
		t.Errorf("Forest under Crypt Ghast: trigger_adds %q, want none", got.TriggerAdds)
	}
	if got := row(t, o); got.TriggerAdds != "" {
		t.Errorf("an opponent's Swamp under my Crypt Ghast: trigger_adds %q, want none", got.TriggerAdds)
	}

	push("Mirari's Wake", "Enchantment", mirarisWakeOracle, me)
	if got := row(t, s); got.TriggerAdds != "{B}{B}" {
		t.Errorf("Swamp under Crypt Ghast and Mirari's Wake: trigger_adds %q, want {B}{B}", got.TriggerAdds)
	}
	if got := row(t, f); got.TriggerAdds != "{G}" {
		t.Errorf("Forest under Mirari's Wake: trigger_adds %q, want {G}", got.TriggerAdds)
	}

	// On the wire it is an omitempty string beside `produced`.
	raw, err := json.Marshal(row(t, f))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"trigger_adds":"{G}"`) {
		t.Errorf("wire row %s does not carry trigger_adds", raw)
	}
	raw, err = json.Marshal(row(t, o))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "trigger_adds") {
		t.Errorf("wire row %s carries an empty trigger_adds", raw)
	}
}
