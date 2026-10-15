package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_if_test.go — #3033: the 21 intervening-"if" cards
// the #3013 ratchet left listed move onto Do(If{…}), so their rows
// declare a footprint, and their batches skip the prompt when nothing
// else in the batch can change the condition's answer.

const (
	if3033ColossalMajesty    = "cac95494-0db0-4bec-8665-998431a6f76b"
	if3033Bugenhagen         = "4228dd75-1fdd-48dc-9259-b841b2e46c64"
	if3033TheGaffer          = "a177295e-3b58-4e46-a1cb-fc003a7a0848"
	if3033IndulgingPatrician = "189131b7-ed59-4750-a4ed-453e4497a5c8"
	if3033UpkeepTokenProbe   = "test-3033-upkeep-token-probe"
)

func init() {
	// "At the beginning of your upkeep, create a 1/1 colorless Construct
	// token": a declared token step that changes the board.
	Register(Spec{
		OracleID: if3033UpkeepTokenProbe,
		Name:     "Upkeep Token Probe",
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginUpkeep, ByYou, "Upkeep Token Probe — create a 1/1 Construct",
				Do(CreateToken{Template: game.Card{
					Name: "Construct", TypeLine: "Token Artifact Creature — Construct", Power: 1, Toughness: 1,
				}, N: 1})),
		},
	})
}

// Every card the ratchet listed for an intervening "if" declares a row
// whose footprint starts with the condition it re-checks.
func TestInterveningIfRowsDeclareTheirCondition(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		cond         game.FootprintCond
	}{
		{"Agent of Treachery", "4cd82c7a-86df-4d29-b578-c009208c0d4b", game.FootprintCondBoard},
		{"Beastbond Outcaster", "19a32e29-45ec-433e-9cb4-b2c32cac8f80", game.FootprintCondBoard},
		{"Colossal Majesty", if3033ColossalMajesty, game.FootprintCondBoard},
		{"Garruk's Uprising", "3127ae9b-a7a7-43ec-89d7-688f8445b33d", game.FootprintCondBoard},
		{"Bugenhagen, Wise Elder", if3033Bugenhagen, game.FootprintCondBoard},
		{"Crested Sunmare", "054fedbb-062e-491e-9a94-594fda33094f", game.FootprintCondLifeGained},
		{"The Gaffer", if3033TheGaffer, game.FootprintCondLifeGained},
		{"Indulging Patrician", if3033IndulgingPatrician, game.FootprintCondLifeGained},
		{"Lunar Convocation", "cb68d1e4-36aa-4a00-b671-8959e7d526d4", game.FootprintCondLifeGained},
		{"Deadeye Brawler", "a7c36b66-4f58-41af-b1c5-704ff0d9eac8", game.FootprintCondCitysBlessing},
		{"Mausoleum Harpy", "ba682da8-50f5-478f-99cc-60a57d229a04", game.FootprintCondCitysBlessing},
		{"Edgar, Moonlit Sovereign", "1917bda6-c0e1-4c80-8009-74c28cf6b8e9", game.FootprintCondNoSpellsCast},
		{"Field of the Dead", "aa959340-c869-4caa-92c7-572bd8d23eef", game.FootprintCondBoard},
		{"Keeper of the Accord", "49b9a3ac-d245-431a-bf40-64d1b8dc2b91", game.FootprintCondBoard},
		{"Ginger, Queen of Sweets", "5ea93956-085d-409b-bd95-dd669fa69eeb", game.FootprintCondMonarch},
		{"Inventors' Fair", "91d4a5fe-fd6d-4b14-a63f-61b4d0ecd9c4", game.FootprintCondBoard},
		{"Lux Artillery", "cbd76b22-d04e-48e7-bcab-c7f5466d67d8", game.FootprintCondBoard},
		{"Padeem, Consul of Innovation", "0c7ba712-6a99-4d2f-9242-a2163a11f69c", game.FootprintCondBoard},
		{"Runaway Steam-Kin", "eb414a8f-566f-4635-80a1-77b2e18ac8dc", game.FootprintCondSource},
		{"Satoru, the Infiltrator", "7555c429-5f2d-4171-b6b0-8e3c8da7f314", game.FootprintCondBoard},
		{"Faramir, Field Commander", "adeed5e8-9019-4740-83ea-67dcd754e194", game.FootprintCondCreatureDied},
	} {
		found := false
		for _, row := range game.CatalogTriggers(c.oracle) {
			if len(row.Footprint) > 1 && row.Footprint[0].Kind == game.FootprintCondition && row.Footprint[0].Cond == c.cond {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no row's footprint starts with the %q condition", c.name, c.cond)
		}
	}
}

// Lunar Convocation's Bat needs life gained AND lost: one If inside
// another, both conditions declared before the token.
func TestANestedIfDeclaresBothConditions(t *testing.T) {
	steps, ok := stepFootprints(If{Cond: CondLifeGained{}, Then: []Applier{
		If{Cond: CondLifeLost{}, Then: []Applier{CreateToken{Template: TokenCard("1/1 black Bat with flying"), N: 1}}},
	}})
	if !ok || len(steps) != 3 || steps[0].Cond != game.FootprintCondLifeGained ||
		steps[1].Cond != game.FootprintCondLifeLost || steps[2].Kind != game.FootprintCreateToken {
		t.Errorf("footprint %+v, want gained, lost, token", steps)
	}
}

// An If whose guarded step declares nothing leaves the row without a
// footprint, as a plain Do would.
func TestAnIfAroundAnUndeclaredStepHasNoFootprint(t *testing.T) {
	if _, ok := stepFootprints(If{Cond: CondMonarch{}, Then: []Applier{GainLife{Player: uuid.New(), Amount: 1}}}); ok {
		t.Error("a GainLife with a fixed player declares nothing, so neither does the If")
	}
	if _, ok := stepFootprints(If{Cond: CondBoard{Name: "no test"}, Then: []Applier{DrawCards{N: 1}}}); ok {
		t.Error("a board condition with no test declares nothing")
	}
}

// The condition is re-checked as the item resolves (CR 603.4).
func TestIfRunsItsStepsOnlyWhenTheConditionHolds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	src := pushBattlefieldCreature(g, me.ID)
	var source game.Card
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == src {
			source = c
		}
	}
	item := game.NewTriggeredItem(&source, "If probe")
	step := If{Cond: CondMonarch{}, Then: []Applier{GainLife{Amount: 2}}}
	life := me.Life
	g.WithWriteLock(func() { _ = step.Apply(NewContext(g, item)) })
	if me.Life != life {
		t.Fatalf("gained life while not the monarch: %d → %d", life, me.Life)
	}
	g.Monarch = me.ID
	g.WithWriteLock(func() { _ = step.Apply(NewContext(g, item)) })
	if me.Life != life+2 {
		t.Errorf("as the monarch: %d → %d, want +2", life, me.Life)
	}
}

// Colossal Majesty and Bugenhagen both draw at the upkeep if a big
// creature is there. Draws change no board, so the order cannot show.
func TestTwoBoardConditionDrawsSkipThePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Colossal Majesty", "Enchantment", if3033ColossalMajesty, false)
	pushCatalogPermanent(g, me.ID, "Bugenhagen, Wise Elder", "Legendary Creature — Human Shaman", if3033Bugenhagen, false)
	b21Push(g, me.ID, "Big Beast", "Creature — Beast", "", 7, 7)
	toMain(t, g)
	advanceToUpkeepOf(t, g, 0)
	hand := me.Hand.Size()
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("two draws behind a power condition asked for an order: %+v", ch)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
}

// A token entering changes the board Colossal Majesty's "if" reads, so
// that batch still asks, counted under condition.
func TestABoardConditionBesideATokenMakerAsks(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Colossal Majesty", "Enchantment", if3033ColossalMajesty, false)
	pushCatalogPermanent(g, me.ID, "Upkeep Token Probe", "Enchantment", if3033UpkeepTokenProbe, false)
	b21Push(g, me.ID, "Big Beast", "Creature — Beast", "", 4, 4)
	toMain(t, g)
	before := promptsByReason(t, game.TriggerOrderReasonCondition)
	advanceToUpkeepOf(t, g, 0)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("a token maker beside a board condition must keep the prompt")
	}
	if got := promptsByReason(t, game.TriggerOrderReasonCondition) - before; got != 1 {
		t.Errorf("counted %v under condition, want 1", got)
	}
}

// The Gaffer and Indulging Patrician both re-check "you gained 3 or more
// life this turn", which nothing can undo: with the life gained, the
// draw and the drain go on with no prompt.
func TestTwoLifeGainedConditionsSkipThePrompt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "The Gaffer", "Legendary Creature — Halfling Peasant", if3033TheGaffer, false)
	gafferless := pushCatalogPermanent(g, me.ID, "Indulging Patrician", "Creature — Vampire Noble", if3033IndulgingPatrician, false)
	toMain(t, g)
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(gafferless, me.ID, 3); err != nil {
			t.Fatal(err)
		}
	})
	advanceToEndStepOf(t, g, 0)
	hand, oppLife := me.Hand.Size(), opp.Life
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a draw and a drain behind a life-gained condition asked for an order: %+v", ch)
	}
	if me.Hand.Size()-hand != 1 || oppLife-opp.Life != 3 {
		t.Errorf("drew %d and the opponent lost %d, want 1 and 3", me.Hand.Size()-hand, oppLife-opp.Life)
	}
}
