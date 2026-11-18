package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vowB09Sharpshooter = "40d6e1ff-995b-42fc-8146-8d5e5f1adcd9"
	vowB09Cadet        = "08a40717-c272-42e6-adb1-0148614b1202"
	vowB09GryffRider   = "e1ca5c24-bbd8-4eed-bb2b-17163f90e72b"
	vowB09Cavalry      = "8e5e6f1d-4945-4074-a33e-c0e8b6b24a47"
	vowB09Trainee      = "e1733392-1402-458b-bbc7-673d733c731c"
	vowB09Recruit      = "fda06d9e-d17f-4bb6-916b-06116e47c8e8"
	vowB09Savior       = "1857dc32-8ff1-4519-b1ca-480a5c2d3b3c"
	vowB09Torens       = "2cb96408-9195-4e41-ad9a-f8c74fbad083"
	vowB09Escort       = "7a8fa68b-8c87-4ecd-ac29-50fd90c62159"
	vowB09Paladin      = "472756f7-7fb8-4630-be14-d78bb9c99157"
	vowB09LaidToRest   = "ae6dc45c-b8c9-4cbf-892f-8c00f4199aaf"
)

func TestVowB09CardsAreFullWithNoCaveats(t *testing.T) {
	want := map[string]string{
		vowB09Sharpshooter: "Apprentice Sharpshooter", vowB09Cadet: "Cloaked Cadet",
		vowB09GryffRider: "Gryff Rider", vowB09Cavalry: "Gryffwing Cavalry",
		vowB09Trainee: "Parish-Blade Trainee", vowB09Recruit: "Rural Recruit",
		vowB09Savior: "Savior of Ollenbock", vowB09Torens: "Torens, Fist of the Angels",
		vowB09Escort: "Ollenbock Escort", vowB09Paladin: "Sigardian Paladin",
		vowB09LaidToRest: "Laid to Rest",
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

func TestVowB09PrintedKeywords(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle, kw string
	}{
		{"Apprentice Sharpshooter", "Creature — Human Archer", vowB09Sharpshooter, "reach"},
		{"Gryff Rider", "Creature — Human Knight", vowB09GryffRider, "flying"},
		{"Gryffwing Cavalry", "Creature — Human Knight", vowB09Cavalry, "flying"},
		{"Ollenbock Escort", "Creature — Human Cleric", vowB09Escort, "vigilance"},
	} {
		id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		if !hasKeywordOf(g, id, tc.kw) {
			t.Errorf("%s has no %s", tc.name, tc.kw)
		}
	}
}

// Every training creature of the batch grows when it attacks beside a bigger
// creature, and not beside an equal one.
func TestVowB09TrainingCreaturesTrain(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Apprentice Sharpshooter", "Creature — Human Archer", vowB09Sharpshooter},
		{"Cloaked Cadet", "Creature — Human Ranger", vowB09Cadet},
		{"Gryff Rider", "Creature — Human Knight", vowB09GryffRider},
		{"Parish-Blade Trainee", "Creature — Human Soldier", vowB09Trainee},
		{"Rural Recruit", "Creature — Human Peasant", vowB09Recruit},
		{"Torens, Fist of the Angels", "Legendary Creature — Human Cleric", vowB09Torens},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			id := b12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
			b12Creature(g, me.ID, "Big", "Creature — Bear", 5, 5)
			attackWith(t, g, opp.ID, id, g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID)
			passPriorityAroundTable(t, g)
			if got := counterCount(g, id, game.CounterPlusOne); got != 1 {
				t.Errorf("+1/+1 counters: %d, want 1", got)
			}
		})
	}
}

func TestVowB09RuralRecruitMakesAThreeOneBoar(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12PlayFromHand(t, g, "Rural Recruit", "Creature — Human Peasant", vowB09Recruit, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Boar" && c.Controller == me.ID {
			found = true
			if c.Power != 3 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "G" {
				t.Errorf("Boar %d/%d %v, want 3/1 green", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if !found {
		t.Fatal("no Boar token")
	}
}

func TestVowB09TorensMakesTrainingSoldiersOnCreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Torens, Fist of the Angels", "Legendary Creature — Human Cleric", vowB09Torens, 2, 2)
	b12PlayFromHand(t, g, "Bear", "Creature — Bear", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	b12PlayFromHand(t, g, "Rock", "Sorcery", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Human Soldier" {
			tokens = append(tokens, c.InstanceID)
			if c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 2 {
				t.Errorf("token %d/%d %v, want 1/1 green and white", c.Power, c.Toughness, c.Colors)
			}
		}
	}
	if len(tokens) != 1 {
		t.Fatalf("Human Soldier tokens: %d, want 1 (a sorcery does not trigger Torens)", len(tokens))
	}
	// The token printed training: a fresh copy (no summoning sickness) trains.
	token := pushToken(g, me.ID, HumanSoldierTrainingToken())
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 4, 4)
	attackWith(t, g, opp.ID, token, big)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, token, game.CounterPlusOne); got != 1 {
		t.Errorf("the token's training: %d counters, want 1", got)
	}
}

func TestVowB09ParishBladeTraineeMovesItsCountersOnDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trainee := b12Push(g, me.ID, "Parish-Blade Trainee", "Creature — Human Soldier", vowB09Trainee, 1, 2)
	other := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	if err := g.AddCounterForEffect(trainee, game.CounterPlusOne, 2); err != nil {
		t.Fatal(err)
	}
	vowB09Sacrifice(t, g, trainee)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: other})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, other, game.CounterPlusOne); got != 2 {
		t.Errorf("counters on the other creature: %d, want 2", got)
	}
}

func TestVowB09CloakedCadetDrawsOncePerTurnForHumanCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C", "D")
	b12Push(g, me.ID, "Cloaked Cadet", "Creature — Human Ranger", vowB09Cadet, 2, 4)
	human := b12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("a counter on a non-Human drew %d cards", got)
	}
	for i := 0; i < 2; i++ {
		if err := g.AddCounterForEffect(human, game.CounterPlusOne, 1); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("cards drawn for two Human counter placements in one turn: %d, want 1", got)
	}
}

func TestVowB09LaidToRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D")
	b12Push(g, me.ID, "Laid to Rest", "Enchantment", vowB09LaidToRest, 0, 0)
	human := b12Creature(g, me.ID, "Villager", "Creature — Human Peasant", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Human Soldier", 1, 1)
	hand, life := me.Hand.Size(), me.Life

	killCreature(t, g, me.ID, human)
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 || me.Life != life {
		t.Errorf("a plain Human dying: drew %d, life %d -> %d; want 1 card and no life", me.Hand.Size()-hand, life, me.Life)
	}
	killCreature(t, g, opp.ID, foe)
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 {
		t.Error("an opponent's Human dying drew a card")
	}
	if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 || me.Hand.Size()-hand != 1 {
		t.Errorf("a non-Human with a counter dying: life %d -> %d, drew %d; want +2 life and no draw", life, me.Life, me.Hand.Size()-hand)
	}
}

func TestVowB09EscortGrantsLifelinkAndIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	escort := b12Push(g, me.ID, "Ollenbock Escort", "Creature — Human Cleric", vowB09Escort, 1, 1)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	if err := g.AddCounterForEffect(grown, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	if err := activateOn(t, g, me.ID, escort, 0, plain); err == nil {
		t.Fatal("targeted a creature with no +1/+1 counter")
	}
	if err := activateOn(t, g, me.ID, escort, 0, grown); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"lifelink", "indestructible"} {
		if !hasKeywordOf(g, grown, kw) {
			t.Errorf("the target has no %s", kw)
		}
	}
	if g.Battlefield.Contains(escort) {
		t.Error("the Escort was not sacrificed")
	}
}

func TestVowB09PaladinStaticAndActivation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	paladin := b12Push(g, me.ID, "Sigardian Paladin", "Creature — Human Knight", vowB09Paladin, 4, 4)
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	if hasKeywordOf(g, paladin, "trample") || hasKeywordOf(g, paladin, "lifelink") {
		t.Fatal("the Paladin has trample/lifelink before any counter was put")
	}
	if err := g.AddCounterForEffect(grown, game.CounterPlusOne, 1); err != nil {
		t.Fatal(err)
	}
	for _, kw := range []string{"trample", "lifelink"} {
		if !hasKeywordOf(g, paladin, kw) {
			t.Errorf("the Paladin has no %s after a +1/+1 counter was put this turn", kw)
		}
	}
	if hasKeywordOf(g, grown, "trample") {
		t.Error("the static grant leaked to another creature")
	}
	if err := activateOn(t, g, me.ID, paladin, 0, grown); err == nil {
		t.Fatal("activated without mana")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{W}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := activateOn(t, g, me.ID, paladin, 0, grown); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"trample", "lifelink"} {
		if !hasKeywordOf(g, grown, kw) {
			t.Errorf("the target has no %s", kw)
		}
	}
}

func TestVowB09GryffwingCavalryPaysToGiveFlying(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		cav := b12Push(g, me.ID, "Gryffwing Cavalry", "Creature — Human Knight", vowB09Cavalry, 2, 2)
		walker := b12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
		advanceTo(t, g, game.StepDeclareAttackers)
		for _, id := range []uuid.UUID{cav, walker} {
			if err := g.DeclareAttacker(id, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		lockInAttacks(t, g)
		pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: walker})
		for i := 0; i < 8 && !hasPayUnlessFor(g, me.ID); i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatalf("PassPriority: %v", err)
			}
		}
		if !hasPayUnlessFor(g, me.ID) {
			t.Fatal("no {1}{W} offer")
		}
		// Mana empties between steps, so fund the offer now.
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{C}"); err != nil {
			t.Fatal(err)
		}
		answerPayUnless(t, g, me.ID, pay)
		if got := hasKeywordOf(g, walker, "flying"); got != pay {
			t.Errorf("pay=%v: the walker has flying = %v", pay, got)
		}
	}
}

func TestVowB09SaviorOfOllenbockExilesOnTrainingAndReturnsOnLeaving(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	savior := b12Push(g, me.ID, "Savior of Ollenbock", "Creature — Human Soldier", vowB09Savior, 1, 2)
	big := b12Creature(g, me.ID, "Big", "Creature — Bear", 5, 5)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Goblin", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{savior, big} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetCard, ID: victim})
	passPriorityAroundTable(t, g)
	if got := counterCount(g, savior, game.CounterPlusOne); got != 1 {
		t.Fatalf("the Savior did not train: %d counters", got)
	}
	if g.Battlefield.Contains(victim) {
		t.Fatal("the opponent's creature was not exiled when the Savior trained")
	}
	vowB09Sacrifice(t, g, savior)
	passPriorityAroundTable(t, g)
	// It comes back as a new object (CR 400.7), under its owner's control.
	if got := countOnBattlefield(g, "Victim", opp.ID); got != 1 {
		t.Errorf("the exiled creature on the battlefield under its owner: %d, want 1", got)
	}
}

// vowB09Sacrifice sends a permanent to the graveyard through the effect door
// that records the last-known information a leaves-the-battlefield ability reads.
func vowB09Sacrifice(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.SacrificePermanentForEffect(id) })
	if err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
}
