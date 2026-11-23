package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The four new printed-retrace cards declare both halves, at the printed
// cost the oracle text gives. Costs were read off the Scryfall dump
// (2026-10-05).
func TestFollowupRetraceCardsDeclareBothHalves(t *testing.T) {
	for _, c := range []struct{ name, oracle, cost string }{
		{"Cenn's Enlistment", effectstest.CennsOracle, "{3}{W}"},
		{"Call the Skybreaker", effectstest.SkybreakOracle, "{5}{U/R}{U/R}"},
		{"Worm Harvest", effectstest.WormOracle, "{2}{B/G}{B/G}{B/G}"},
		{"Formless Genesis", effectstest.FormlessOracle, "{2}{G}"},
	} {
		if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
			t.Fatalf("%s does not open the graveyard", c.name)
		}
		offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
		if len(offers) != 1 || offers[0].Key != "retrace" || offers[0].ManaCost != c.cost ||
			offers[0].DiscardFromHand == nil || offers[0].ExileOnLeavingStack {
			t.Errorf("%s graveyard offers = %+v, want one retrace for %s", c.name, offers, c.cost)
		}
	}
}

// retraceFromGraveyard puts the card in the graveyard, pays `mana` and
// retraces it discarding the table's hand land.
func retraceFromGraveyard(t *testing.T, name, typeLine, cost, oracle, mana string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, land, _ := effectstest.RetraceTable(t)
	card := effectstest.GraveyardCardOf(g, me, name, typeLine, cost, oracle)
	effectstest.PayMana(t, g, me, mana)
	if err := effectstest.RetraceCast(g, me, card, land); err != nil {
		t.Fatalf("retrace %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InGraveyard(me, card) {
		t.Fatalf("%s is not back in the graveyard to be retraced again", name)
	}
	return g, me, card
}

func TestCennsEnlistmentRetracesIntoTwoKithkinSoldiers(t *testing.T) {
	g, me, _ := retraceFromGraveyard(t, "Cenn's Enlistment", "Sorcery", "{3}{W}", effectstest.CennsOracle, "{W}{W}{W}{W}")
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Kithkin Soldier"); got != 2 {
		t.Fatalf("Kithkin Soldier tokens = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Kithkin Soldier" {
			if !effects.IsToken(c) || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("token = %+v, want a 1/1 white token", c)
			}
		}
	}
}

func TestCallTheSkybreakerRetracesIntoAFlyingElemental(t *testing.T) {
	g, me, _ := retraceFromGraveyard(t, "Call the Skybreaker", "Sorcery", "{5}{U/R}{U/R}", effectstest.SkybreakOracle,
		"{U}{U}{U}{U}{U}{U}{R}")
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if effectstest.EffectivePower(t, g, id) != 5 || effectstest.EffectiveToughness(t, g, id) != 5 {
		t.Error("the Elemental is not 5/5")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("the Elemental has no flying")
	}
	c, _ := effectstest.BattlefieldCard(g, id)
	if len(c.Colors) != 2 {
		t.Errorf("colours = %v, want blue and red", c.Colors)
	}
}

// The land that pays retrace is discarded as a cost, with the spell on
// the stack, so it counts: with an empty graveyard to start and one land
// discarded, Worm Harvest makes exactly one Worm.
func TestWormHarvestCountsTheLandDiscardedToRetraceIt(t *testing.T) {
	g, me, land, _ := effectstest.RetraceTable(t)
	extraLand := effectstest.GraveyardCardOf(g, me, "Dead Swamp", "Basic Land — Swamp", "", "")
	_ = extraLand
	harvest := effectstest.GraveyardCardOf(g, me, "Worm Harvest", "Sorcery", "{2}{B/G}{B/G}{B/G}", effectstest.WormOracle)
	// A nonland card in the graveyard must not be counted.
	effectstest.GraveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.PayMana(t, g, me, "{B}{B}{B}{B}{B}")
	if err := effectstest.RetraceCast(g, me, harvest, land); err != nil {
		t.Fatalf("retrace: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Worm"); got != 2 {
		t.Fatalf("Worms = %d, want 2 (one land already there, one discarded to retrace)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Worm" && (len(c.Colors) != 2 || c.Power != 1) {
			t.Errorf("Worm = %+v, want a 1/1 black and green token", c)
		}
	}
}

func TestFormlessGenesisMakesAnXByXDeathtouchChangeling(t *testing.T) {
	g, me, land, _ := effectstest.RetraceTable(t)
	effectstest.GraveyardCardOf(g, me, "Dead Forest", "Basic Land — Forest", "", "")
	effectstest.GraveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	genesis := effectstest.GraveyardCardOf(g, me, "Formless Genesis", "Kindred Sorcery — Shapeshifter", "{2}{G}", effectstest.FormlessOracle)
	effectstest.PayMana(t, g, me, "{G}{G}{G}")
	if err := effectstest.RetraceCast(g, me, genesis, land); err != nil {
		t.Fatalf("retrace: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Shapeshifter" && c.Controller == me.ID {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Shapeshifter token")
	}
	if p, tt := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 3 || tt != 3 {
		t.Errorf("token is %d/%d, want 3/3 (two lands there, one discarded)", p, tt)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "deathtouch") {
		t.Error("the token has no deathtouch")
	}
	if c, _ := effectstest.BattlefieldCard(g, id); len(c.Colors) != 0 {
		t.Errorf("the token is coloured %v, want colourless", c.Colors)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the token is not on the battlefield")
	}
}

// "Merfolk AND Druid cards" is a union: either type qualifies, both is
// fine, and neither does not. Pure filter test, no table.
func TestCreatureTypesAnyIsAUnion(t *testing.T) {
	f := game.PermissionFilter{CreatureTypesAny: [2]string{"Merfolk", "Druid"}}
	for _, c := range []struct {
		typeLine string
		want     bool
	}{
		{"Creature — Merfolk", true},
		{"Creature — Elf Druid", true},
		{"Creature — Merfolk Druid", true},
		{"Creature — Merfolk Wizard", true},
		{"Creature — Elf Warrior", false},
		{"Sorcery", false},
		{"Instant", false},
	} {
		if got := f.Matches(game.Card{TypeLine: c.typeLine}); got != c.want {
			t.Errorf("%q: Matches = %v, want %v", c.typeLine, got, c.want)
		}
	}
	// The zero filter constrains nothing; the field alone never narrows.
	if !(game.PermissionFilter{}).Matches(game.Card{TypeLine: "Creature — Elf"}) {
		t.Error("the zero filter must match everything")
	}
}

func TestDeeprootHistorianGrantsRetraceToMerfolkAndDruidCards(t *testing.T) {
	g, me, _, land := effectstest.HistorianTable(t)
	merfolk := effectstest.GraveyardCardOf(g, me, "Gravebound Merfolk", "Creature — Merfolk Wizard", "{1}{U}", "")
	druid := effectstest.GraveyardCardOf(g, me, "Gravebound Druid", "Creature — Elf Druid", "{1}{G}", "")
	elf := effectstest.GraveyardCardOf(g, me, "Gravebound Elf", "Creature — Elf Warrior", "{1}{G}", "")

	effectstest.PayMana(t, g, me, "{G}{G}{U}{U}")
	if err := effectstest.RetraceCast(g, me, elf, land); err == nil {
		t.Fatal("an Elf Warrior card was retraced under Deeproot Historian")
	}
	if len(effectstest.RetraceMoves(g, me.ID, elf)) != 0 {
		t.Error("the enumerator offers an Elf Warrior")
	}
	if len(effectstest.RetraceMoves(g, me.ID, merfolk)) == 0 || len(effectstest.RetraceMoves(g, me.ID, druid)) == 0 {
		t.Fatal("the enumerator does not offer a Merfolk and a Druid card")
	}
	if err := effectstest.RetraceCast(g, me, merfolk, land); err != nil {
		t.Fatalf("retrace of a Merfolk card: %v", err)
	}
	if !effectstest.InGraveyard(me, land) {
		t.Error("the land was not discarded")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, merfolk); !ok {
		t.Fatal("the retraced Merfolk did not resolve onto the battlefield")
	}
}

// The grant is derived from the battlefield: it is gone when the
// Historian is, and it is the controller's alone.
func TestDeeprootHistorianGrantEndsWhenItLeavesAndIsItsControllersAlone(t *testing.T) {
	g, me, historian, land := effectstest.HistorianTable(t)
	merfolk := effectstest.GraveyardCardOf(g, me, "Gravebound Merfolk", "Creature — Merfolk", "{1}{U}", "")
	theirs := effectstest.GraveyardCardOf(g, g.Seats[1], "Their Merfolk", "Creature — Merfolk", "{1}{U}", "")
	effectstest.PayMana(t, g, me, "{U}{U}")
	if len(effectstest.RetraceMoves(g, g.Seats[1].ID, theirs)) != 0 {
		t.Error("an opponent's Merfolk card was offered retrace under MY Historian")
	}
	if len(effectstest.RetraceMoves(g, me.ID, merfolk)) == 0 {
		t.Fatal("control: no retrace offered under the Historian")
	}
	if _, err := g.Battlefield.Remove(historian); err != nil {
		t.Fatal(err)
	}
	if err := effectstest.RetraceCast(g, me, merfolk, land); err == nil {
		t.Error("retrace survived the Historian leaving the battlefield")
	}
	if len(effectstest.RetraceMoves(g, me.ID, merfolk)) != 0 {
		t.Error("the enumerator still offers retrace without the Historian")
	}
}

// ---------------------------------------------------------- the card disclosure

func TestFollowupCardsAndCacklingCounterpartShipFull(t *testing.T) {
	for _, oracle := range []string{effectstest.HistorianOracle, effectstest.CennsOracle, effectstest.SkybreakOracle, effectstest.WormOracle, effectstest.FormlessOracle,
		"9e2adca5-f39c-4a09-bcce-8238ebac2c4a"} {
		s, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if s.Completeness != effects.CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full", s.Name, s.Completeness, s.Caveats)
		}
	}
}
