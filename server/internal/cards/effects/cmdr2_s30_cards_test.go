package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s30_cards_test.go — batch 30 of the wave-2 Commander-deck cards
// (#5368): reanimation and graveyard creatures.

const (
	s30StormOfSoulsOracle = "7022ae9a-2589-493b-a050-a230fd02f80b"
)

func TestCmdr2S30StormOfSoulsReturnsEveryCreatureAsAOneOneSpirit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	golem := pushGraveyardCardTyped(me, "Dead Golem", "Artifact Creature — Golem")
	stone := pushGraveyardCardTyped(me, "Dead Stone", "Artifact")
	theirs := pushGraveyardCardTyped(opp, "Their Elf", "Creature — Elf")
	storm := castCatalogSpell(t, g, "Storm of Souls", "Sorcery", s30StormOfSoulsOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, golem} {
		if !g.Battlefield.Contains(id) {
			t.Fatalf("creature card %s did not return", id)
		}
		c := layeredCard(t, g, id)
		eff := c.Effective()
		if !c.HasSubtype("Spirit") || !slices.Contains(eff.Abilities, "flying") || eff.Power != 1 || eff.Toughness != 1 {
			t.Errorf("%s: subtypes %v abilities %v %d/%d, want a 1/1 Spirit with flying", c.Name, eff.Subtypes, eff.Abilities, eff.Power, eff.Toughness)
		}
	}
	if !layeredCard(t, g, bear).HasSubtype("Bear") {
		t.Error("it keeps its other types")
	}
	if g.Battlefield.Contains(stone) || g.Battlefield.Contains(theirs) {
		t.Error("a noncreature card and an opponent's creature card stay put")
	}
	if !inExile(g, storm) {
		t.Error("Storm of Souls is exiled")
	}
}

const (
	s30MarshalsAnthemOracle = "41fb2b0e-d0a0-4e47-a6f3-fc9860ade65a"
	s30RotHulkOracle        = "36117e99-a2bc-46f9-8e17-bd2e2a9c31f4"
)

// s30PickTargets answers the latest pick_target prompt with every card given.
func s30PickTargets(t *testing.T, g *game.Game, chooser uuid.UUID, cards ...uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	var refs []game.TargetRef
	for _, c := range cards {
		refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: c})
	}
	if err := g.ResolvePickTargets(pick.ID, chooser, refs); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
}

func TestCmdr2S30RotHulkReturnsUpToOneZombiePerOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opponents := len(g.Seats) - 1
	var zombies []uuid.UUID
	for i := 0; i < opponents+1; i++ {
		zombies = append(zombies, pushGraveyardCardTyped(me, "Zombie", "Creature — Zombie"))
	}
	elf := pushGraveyardCardTyped(me, "Dead Elf", "Creature — Elf")
	castCatalogSpell(t, g, "Rot Hulk", "Creature — Zombie", s30RotHulkOracle, nil)
	passPriorityAroundTable(t, g)
	// One more pick than there are opponents is refused.
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	var tooMany []game.TargetRef
	for _, z := range zombies {
		tooMany = append(tooMany, game.TargetRef{Kind: game.TargetCard, ID: z})
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, tooMany); err == nil {
		t.Error("more targets than opponents was accepted")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: elf}}); err == nil {
		t.Error("a non-Zombie was accepted")
	}
	s30PickTargets(t, g, me.ID, zombies[:opponents]...)
	passPriorityAroundTable(t, g)
	for _, z := range zombies[:opponents] {
		if !g.Battlefield.Contains(z) {
			t.Errorf("Zombie %s did not return", z)
		}
	}
	if g.Battlefield.Contains(zombies[opponents]) || g.Battlefield.Contains(elf) {
		t.Error("only the chosen Zombies return")
	}
}

func TestCmdr2S30MarshalsAnthemPumpsAndReturnsOnePerKick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardTyped(me, "Dead A", "Creature — Bear")
	b := pushGraveyardCardTyped(me, "Dead B", "Creature — Bear")
	c := pushGraveyardCardTyped(me, "Dead C", "Creature — Bear")
	anthem, err := castWithOptionalCosts(t, g, "Marshal's Anthem", "Enchantment", s30MarshalsAnthemOracle, nil, []int{0, 0}, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Kicked twice: three targets are too many.
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Error("three targets accepted for a twice-kicked Anthem")
	}
	s30PickTargets(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(anthem) || !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) || g.Battlefield.Contains(c) {
		t.Fatal("the Anthem and exactly the two chosen creatures are on the battlefield")
	}
	if e := layeredCard(t, g, a).Effective(); e.Power != 3 || e.Toughness != 3 {
		t.Errorf("returned creature is %d/%d, want 3/3 with the Anthem", e.Power, e.Toughness)
	}
}

func TestCmdr2S30MarshalsAnthemUnkickedReturnsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := pushGraveyardCardTyped(me, "Dead A", "Creature — Bear")
	if _, err := castWithOptionalCosts(t, g, "Marshal's Anthem", "Enchantment", s30MarshalsAnthemOracle, nil, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || g.Battlefield.Contains(a) {
		t.Error("an unkicked Anthem returns nothing")
	}
}

const s30MoldgrafMonstrosityOracle = "3c4d5392-27cf-492e-9ae1-dee330b82223"

func TestCmdr2S30MoldgrafMonstrosityExilesItselfAndReturnsTwoAtRandom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := pushCatalogPermanent(g, me.ID, "Moldgraf Monstrosity", "Creature — Insect", s30MoldgrafMonstrosityOracle, false)
	var dead []uuid.UUID
	for i := 0; i < 3; i++ {
		dead = append(dead, pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear"))
	}
	land := pushGraveyardCardTyped(me, "Dead Land", "Land")
	killCreature(t, g, me.ID, hulk)
	passPriorityAroundTable(t, g)
	if !inExile(g, hulk) {
		t.Error("the Monstrosity is exiled")
	}
	back := 0
	for _, id := range dead {
		if g.Battlefield.Contains(id) {
			back++
		}
	}
	if back != 2 {
		t.Errorf("%d creature cards returned, want exactly 2", back)
	}
	if g.Battlefield.Contains(land) || g.Battlefield.Contains(hulk) {
		t.Error("a land card and the Monstrosity itself stay out")
	}
}

func TestCmdr2S30MoldgrafMonstrosityWithOneCreatureCardReturnsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	hulk := pushCatalogPermanent(g, me.ID, "Moldgraf Monstrosity", "Creature — Insect", s30MoldgrafMonstrosityOracle, false)
	only := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	killCreature(t, g, me.ID, hulk)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(only) || !inExile(g, hulk) {
		t.Error("the single creature card returns and the Monstrosity is exiled")
	}
}

const (
	s30CarrionGrubOracle       = "c54b375d-a456-4886-96f2-8dc7a0024a5b"
	s30MoldgrafMillipedeOracle = "48edac92-0dd8-4274-bf1b-122243770540"
)

func TestCmdr2S30CarrionGrubReadsTheGreatestPowerInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	grub := pushCatalogPermanent(g, me.ID, "Carrion Grub", "Creature — Insect", s30CarrionGrubOracle, false)
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == grub {
			g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 0, 5
		}
	}
	if e := layeredCard(t, g, grub).Effective(); e.Power != 0 {
		t.Fatalf("empty graveyard: power %d, want 0", e.Power)
	}
	toGraveyard := func(p *game.Player, name, typeLine string, power int) {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: 1, Owner: p.ID, Controller: p.ID})
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: p.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: p.ID}, id); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	toGraveyard(me, "Small", "Creature — Bear", 2)
	toGraveyard(me, "Big", "Creature — Bear", 6)
	toGraveyard(me, "Huge Artifact", "Artifact", 9)
	toGraveyard(opp, "Their Giant", "Creature — Giant", 9)
	if e := layeredCard(t, g, grub).Effective(); e.Power != 6 || e.Toughness != 5 {
		t.Errorf("got %d/%d, want 6/5: only creature cards in my graveyard count", e.Power, e.Toughness)
	}
}

func TestCmdr2S30CarrionGrubMillsFourOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 6; i++ {
		pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery"})
	}
	before, gy := me.Library.Size(), me.Graveyard.Size()
	castCatalogSpell(t, g, "Carrion Grub", "Creature — Insect", s30CarrionGrubOracle, nil)
	passPriorityAroundTable(t, g)
	if me.Library.Size() != before-4 || me.Graveyard.Size() != gy+4 {
		t.Errorf("library %d->%d graveyard %d->%d, want four milled", before, me.Library.Size(), gy, me.Graveyard.Size())
	}
}

// s30OnTop puts a card on top of a library (pushLibraryCardForTest goes to the bottom).
func s30OnTop(p *game.Player, name, typeLine string) {
	p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
}

func TestCmdr2S30MoldgrafMillipedeCountsTheCardsItJustMilled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	s30OnTop(me, "Deep", "Creature — Bear")
	for i := 0; i < 2; i++ {
		s30OnTop(me, "Milled Bear", "Creature — Bear")
	}
	s30OnTop(me, "Spell", "Sorcery")
	id := castCatalogSpell(t, g, "Moldgraf Millipede", "Creature — Insect Horror", s30MoldgrafMillipedeOracle, nil)
	passPriorityAroundTable(t, g)
	// The top three (two creatures and a sorcery) are milled; with the
	// creature already there that is three creature cards.
	if got := countersOn(g, id, game.CounterPlusOne); got != 3 {
		t.Errorf("+1/+1 counters = %d, want 3", got)
	}
}
