package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	dmuB01ElfhameWurm = "9a74a1dd-c36f-4900-b835-fc39dd725bde"
	dmuB01Drake       = "6bc14f1e-1df8-4e72-b05a-a848faa3b080"
	dmuB01Cavalier    = "898763f8-a333-43f5-9f66-52e2adbe5758"
	dmuB01Ancient     = "26afc8c1-c23f-4212-a2aa-386c71c4a7f9"
	dmuB01Abomination = "6358ee1c-3577-4196-bae4-29a3c160a66a"
	dmuB01Librarian   = "eb570331-35da-44b8-b711-ccf7444d8e39"
	dmuB01Swarm       = "1396e7f1-ac91-4778-9f71-493c7a44bd51"
	dmuB01Apparition  = "b38c0575-37f7-413c-9dad-c8d7e6a41b90"
	dmuB01Kavu        = "49475268-be95-4670-a5ca-2f2b6c5cf005"
	dmuB01Vanguard    = "f1861fac-e0e5-4b96-b6e4-6b6573bd02d1"
	dmuB01Drawbridge  = "e49eb21c-2a0d-4fd0-9f69-1bff6e94f9e7"
	dmuB01Turtle      = "ac2668a3-7a26-47b1-9ad4-56e23643de80"
	dmuB01Axe         = "fbdd3ba9-d799-476b-87bc-a689d8e3cff9"
)

func TestDmuB01CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		dmuB01ElfhameWurm: "Elfhame Wurm", dmuB01Drake: "Soaring Drake",
		dmuB01Cavalier: "Mesa Cavalier", dmuB01Ancient: "Mossbeard Ancient",
		dmuB01Abomination: "Toxic Abomination", dmuB01Librarian: "Automatic Librarian",
		dmuB01Swarm: "Battlefly Swarm", dmuB01Apparition: "Tattered Apparition",
		dmuB01Kavu: "Flowstone Kavu", dmuB01Vanguard: "Charismatic Vanguard",
		dmuB01Drawbridge: "Clockwork Drawbridge", dmuB01Turtle: "Tidepool Turtle",
		dmuB01Axe: "Vanquisher's Axe",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestDmuB01PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Elfhame Wurm", "Creature — Wurm", dmuB01ElfhameWurm, []string{"vigilance", "trample"}},
		{"Soaring Drake", "Creature — Drake", dmuB01Drake, []string{"flying"}},
		{"Mesa Cavalier", "Creature — Human Knight", dmuB01Cavalier, []string{"flying"}},
		{"Mossbeard Ancient", "Creature — Treefolk", dmuB01Ancient, []string{"trample"}},
		{"Battlefly Swarm", "Creature — Phyrexian Insect", dmuB01Swarm, []string{"flying"}},
		{"Tattered Apparition", "Creature — Shade", dmuB01Apparition, []string{"flying"}},
		{"Flowstone Kavu", "Creature — Kavu", dmuB01Kavu, []string{"menace"}},
		{"Clockwork Drawbridge", "Artifact Creature — Wall", dmuB01Drawbridge, []string{"defender"}},
	} {
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.kws {
			if !hasKeywordOf(g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestDmuB01EntersLifeChanges(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		delta                  int
	}{
		{"Mesa Cavalier", "Creature — Human Knight", dmuB01Cavalier, 2},
		{"Mossbeard Ancient", "Creature — Treefolk", dmuB01Ancient, 5},
		{"Toxic Abomination", "Creature — Phyrexian Zombie", dmuB01Abomination, -2},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		castCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		passPriorityAroundTable(t, g)
		passPriorityAroundTable(t, g)
		if me.Life != life+tc.delta {
			t.Errorf("%s: life %d -> %d, want %+d", tc.name, life, me.Life, tc.delta)
		}
	}
}

func TestDmuB01LibrarianScriesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	castCatalogSpell(t, g, "Automatic Librarian", "Artifact Creature — Construct", dmuB01Librarian, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil {
		t.Fatal("no scry was asked")
	}
	if len(scry.ScryCards) != 2 {
		t.Errorf("scry looked at %d cards, want 2", len(scry.ScryCards))
	}
}

func TestDmuB01SwarmGainsDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := b12Push(g, me.ID, "Battlefly Swarm", "Creature — Phyrexian Insect", dmuB01Swarm, 1, 1)
	if err := broB01Activate(t, g, me, id, 0, "{C}"); err == nil {
		t.Fatal("activated with colorless mana for {B}")
	}
	if hasKeywordOf(g, id, "deathtouch") {
		t.Fatal("deathtouch before activation")
	}
	if err := broB01Activate(t, g, me, id, 0, "{B}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasKeywordOf(g, id, "deathtouch") {
		t.Error("no deathtouch after the ability resolved")
	}
}

func TestDmuB01ApparitionAndKavuPump(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	app := b12Push(g, me.ID, "Tattered Apparition", "Creature — Shade", dmuB01Apparition, 2, 2)
	kavu := b12Push(g, me.ID, "Flowstone Kavu", "Creature — Kavu", dmuB01Kavu, 2, 3)
	if err := broB01Activate(t, g, me, app, 0, "{B}"); err == nil {
		t.Fatal("Apparition activated for {B} alone")
	}
	if err := broB01Activate(t, g, me, app, 0, "{C}{B}"); err != nil {
		t.Fatalf("Apparition: %v", err)
	}
	if err := broB01Activate(t, g, me, kavu, 0, "{R}"); err != nil {
		t.Fatalf("Kavu: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, app); p != 3 || tough != 3 {
		t.Errorf("Apparition %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, kavu); p != 3 || tough != 2 {
		t.Errorf("Kavu %d/%d, want 3/2", p, tough)
	}
}

func TestDmuB01VanguardPumpsTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	v := b12Push(g, me.ID, "Charismatic Vanguard", "Creature — Dwarf Soldier", dmuB01Vanguard, 3, 2)
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := broB01Activate(t, g, me, v, 0, "{C}{C}{C}{W}"); err == nil {
		t.Fatal("activated for four mana")
	}
	if err := broB01Activate(t, g, me, v, 0, "{C}{C}{C}{C}{W}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{v, mine} {
		if p, tough := effectivePTOf(t, g, id); p < 3 || tough < 3 {
			t.Errorf("own creature %d/%d not pumped", p, tough)
		}
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("opponent's creature %d/%d, want 2/2", p, tough)
	}
}

func TestDmuB01DrawbridgeTapsATarget(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := b12Push(g, me.ID, "Clockwork Drawbridge", "Artifact Creature — Wall", dmuB01Drawbridge, 0, 3)
	target := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err == nil {
		t.Fatal("activated with one mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !tappedNow(g, target) {
		t.Error("target not tapped")
	}
	if !tappedNow(g, w) {
		t.Error("Drawbridge not tapped as a cost")
	}
}

func TestDmuB01TurtleScries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	id := b12Push(g, me.ID, "Tidepool Turtle", "Creature — Turtle", dmuB01Turtle, 2, 5)
	if err := broB01Activate(t, g, me, id, 0, "{U}"); err == nil {
		t.Fatal("activated for {U} alone")
	}
	if err := broB01Activate(t, g, me, id, 0, "{C}{C}{U}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	scry := latestChoiceOfKind(g, game.PendingChoiceScry)
	if scry == nil || len(scry.ScryCards) != 1 {
		t.Fatalf("expected a scry 1 prompt, got %+v", scry)
	}
}

func TestDmuB01AxePumpsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := seedBear(g, me.ID)
	axe := seedEquipment(g, me.ID, "Vanquisher's Axe", dmuB01Axe)
	equipTo(t, g, me.ID, axe, bear)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("equipped bear %d/%d, want 4/2", p, tough)
	}
}
