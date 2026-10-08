package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// retrace_followups_test.go — #2550. The retrace cards #2528 left out
// (Deeproot Historian's two-type grant and four token makers) and the
// Cackling Counterpart caveat that had gone stale.

const (
	historianOracle = "4b8d82fe-571d-4fd6-8d0c-77b7e5ef3e23"
	cennsOracle     = "49fe9f5a-5821-4586-b913-7d8aef1f8669"
	skybreakOracle  = "a7213bef-0e9c-44e1-97cb-d1da0ef370bf"
	wormOracle      = "b2223f1c-e607-43e0-86dd-5e3225330066"
	formlessOracle  = "b978bdfe-a73e-4e99-90d7-4f99182c45ee"
)

// The four new printed-retrace cards declare both halves, at the printed
// cost the oracle text gives. Costs were read off the Scryfall dump
// (2026-10-05).
func TestFollowupRetraceCardsDeclareBothHalves(t *testing.T) {
	for _, c := range []struct{ name, oracle, cost string }{
		{"Cenn's Enlistment", cennsOracle, "{3}{W}"},
		{"Call the Skybreaker", skybreakOracle, "{5}{U/R}{U/R}"},
		{"Worm Harvest", wormOracle, "{2}{B/G}{B/G}{B/G}"},
		{"Formless Genesis", formlessOracle, "{2}{G}"},
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
	g, me, land, _ := retraceTable(t)
	card := graveyardCardOf(g, me, name, typeLine, cost, oracle)
	payMana(t, g, me, mana)
	if err := retraceCast(g, me, card, land); err != nil {
		t.Fatalf("retrace %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, card) {
		t.Fatalf("%s is not back in the graveyard to be retraced again", name)
	}
	return g, me, card
}

func TestCennsEnlistmentRetracesIntoTwoKithkinSoldiers(t *testing.T) {
	g, me, _ := retraceFromGraveyard(t, "Cenn's Enlistment", "Sorcery", "{3}{W}", cennsOracle, "{W}{W}{W}{W}")
	if got := countBattlefieldNamed(g, me.ID, "Kithkin Soldier"); got != 2 {
		t.Fatalf("Kithkin Soldier tokens = %d, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Kithkin Soldier" {
			if !IsToken(c) || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" {
				t.Errorf("token = %+v, want a 1/1 white token", c)
			}
		}
	}
}

func TestCallTheSkybreakerRetracesIntoAFlyingElemental(t *testing.T) {
	g, me, _ := retraceFromGraveyard(t, "Call the Skybreaker", "Sorcery", "{5}{U/R}{U/R}", skybreakOracle,
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
	if effectivePower(t, g, id) != 5 || effectiveToughness(t, g, id) != 5 {
		t.Error("the Elemental is not 5/5")
	}
	if !effectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("the Elemental has no flying")
	}
	c, _ := battlefieldCard(g, id)
	if len(c.Colors) != 2 {
		t.Errorf("colours = %v, want blue and red", c.Colors)
	}
}

// The land that pays retrace is discarded as a cost, with the spell on
// the stack, so it counts: with an empty graveyard to start and one land
// discarded, Worm Harvest makes exactly one Worm.
func TestWormHarvestCountsTheLandDiscardedToRetraceIt(t *testing.T) {
	g, me, land, _ := retraceTable(t)
	extraLand := graveyardCardOf(g, me, "Dead Swamp", "Basic Land — Swamp", "", "")
	_ = extraLand
	harvest := graveyardCardOf(g, me, "Worm Harvest", "Sorcery", "{2}{B/G}{B/G}{B/G}", wormOracle)
	// A nonland card in the graveyard must not be counted.
	graveyardCardOf(g, me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	payMana(t, g, me, "{B}{B}{B}{B}{B}")
	if err := retraceCast(g, me, harvest, land); err != nil {
		t.Fatalf("retrace: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Worm"); got != 2 {
		t.Fatalf("Worms = %d, want 2 (one land already there, one discarded to retrace)", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Worm" && (len(c.Colors) != 2 || c.Power != 1) {
			t.Errorf("Worm = %+v, want a 1/1 black and green token", c)
		}
	}
}

func TestFormlessGenesisMakesAnXByXDeathtouchChangeling(t *testing.T) {
	g, me, land, _ := retraceTable(t)
	graveyardCardOf(g, me, "Dead Forest", "Basic Land — Forest", "", "")
	graveyardCardOf(g, me, "Dead Island", "Basic Land — Island", "", "")
	genesis := graveyardCardOf(g, me, "Formless Genesis", "Kindred Sorcery — Shapeshifter", "{2}{G}", formlessOracle)
	payMana(t, g, me, "{G}{G}{G}")
	if err := retraceCast(g, me, genesis, land); err != nil {
		t.Fatalf("retrace: %v", err)
	}
	passPriorityAroundTable(t, g)
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Shapeshifter" && c.Controller == me.ID {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("no Shapeshifter token")
	}
	if p, tt := effectivePower(t, g, id), effectiveToughness(t, g, id); p != 3 || tt != 3 {
		t.Errorf("token is %d/%d, want 3/3 (two lands there, one discarded)", p, tt)
	}
	if !effectiveAbilitiesContain(t, g, id, "deathtouch") {
		t.Error("the token has no deathtouch")
	}
	if c, _ := battlefieldCard(g, id); len(c.Colors) != 0 {
		t.Errorf("the token is coloured %v, want colourless", c.Colors)
	}
	if !g.Battlefield.Contains(id) {
		t.Error("the token is not on the battlefield")
	}
}

// ---------------------------------------------------------- Deeproot Historian

func historianTable(t *testing.T) (g *game.Game, me *game.Player, historian, land uuid.UUID) {
	t.Helper()
	g, me, land, _ = retraceTable(t)
	historian = pushCatalogPermanent(g, me.ID, "Deeproot Historian", "Creature — Merfolk Druid", historianOracle, false)
	return
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
	g, me, _, land := historianTable(t)
	merfolk := graveyardCardOf(g, me, "Gravebound Merfolk", "Creature — Merfolk Wizard", "{1}{U}", "")
	druid := graveyardCardOf(g, me, "Gravebound Druid", "Creature — Elf Druid", "{1}{G}", "")
	elf := graveyardCardOf(g, me, "Gravebound Elf", "Creature — Elf Warrior", "{1}{G}", "")

	payMana(t, g, me, "{G}{G}{U}{U}")
	if err := retraceCast(g, me, elf, land); err == nil {
		t.Fatal("an Elf Warrior card was retraced under Deeproot Historian")
	}
	if len(retraceMoves(g, me.ID, elf)) != 0 {
		t.Error("the enumerator offers an Elf Warrior")
	}
	if len(retraceMoves(g, me.ID, merfolk)) == 0 || len(retraceMoves(g, me.ID, druid)) == 0 {
		t.Fatal("the enumerator does not offer a Merfolk and a Druid card")
	}
	if err := retraceCast(g, me, merfolk, land); err != nil {
		t.Fatalf("retrace of a Merfolk card: %v", err)
	}
	if !inGraveyard(me, land) {
		t.Error("the land was not discarded")
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, merfolk); !ok {
		t.Fatal("the retraced Merfolk did not resolve onto the battlefield")
	}
}

// The grant is derived from the battlefield: it is gone when the
// Historian is, and it is the controller's alone.
func TestDeeprootHistorianGrantEndsWhenItLeavesAndIsItsControllersAlone(t *testing.T) {
	g, me, historian, land := historianTable(t)
	merfolk := graveyardCardOf(g, me, "Gravebound Merfolk", "Creature — Merfolk", "{1}{U}", "")
	theirs := graveyardCardOf(g, g.Seats[1], "Their Merfolk", "Creature — Merfolk", "{1}{U}", "")
	payMana(t, g, me, "{U}{U}")
	if len(retraceMoves(g, g.Seats[1].ID, theirs)) != 0 {
		t.Error("an opponent's Merfolk card was offered retrace under MY Historian")
	}
	if len(retraceMoves(g, me.ID, merfolk)) == 0 {
		t.Fatal("control: no retrace offered under the Historian")
	}
	if _, err := g.Battlefield.Remove(historian); err != nil {
		t.Fatal(err)
	}
	if err := retraceCast(g, me, merfolk, land); err == nil {
		t.Error("retrace survived the Historian leaving the battlefield")
	}
	if len(retraceMoves(g, me.ID, merfolk)) != 0 {
		t.Error("the enumerator still offers retrace without the Historian")
	}
}

// ---------------------------------------------------------- the card disclosure

func TestFollowupCardsAndCacklingCounterpartShipFull(t *testing.T) {
	for _, oracle := range []string{historianOracle, cennsOracle, skybreakOracle, wormOracle, formlessOracle,
		"9e2adca5-f39c-4a09-bcce-8238ebac2c4a"} {
		s, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not in the catalog", oracle)
		}
		if s.Completeness != CompletenessFull || len(s.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full", s.Name, s.Completeness, s.Caveats)
		}
	}
}
