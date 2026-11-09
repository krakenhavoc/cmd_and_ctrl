package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander deck batches 7 and 8 (#5368): the cards' tests.

const (
	b7RhonasOracle       = "5d2acd3e-4782-42e3-bcd5-80f8b439cc28"
	b7RipjawOracle       = "80bb9ce4-1f3c-42ca-b75f-9a6b2f42073b"
	b7SteelLeafOracle    = "666af637-9282-4121-962a-048e528f9221"
	b7SurrakOracle       = "95be5f72-1f81-4ec6-9bd9-f0b07bfc874b"
	b7TangleweaveOracle  = "b9cec31a-9891-414f-b528-954612168832"
	b7ThickestOracle     = "4d60dae0-7be4-4859-99ac-781a27bcfd93"
	b7WhiptongueOracle   = "53d5d961-d6a3-49b1-b335-2c2b49972008"
	b7WhispererOracle    = "c1e456b6-5d2b-4c42-9464-54a8f9f8f3ba"
	b7YevaOracle         = "2d513c3e-bd27-479e-af32-7f66d96d0de0"
	b7DarksteelJugOracle = "4893a301-29a8-4a0e-845c-8ddb5cb8db25"
	b8GraazOracle        = "aebfdcc0-c5b9-4737-8498-764a5021b087"
	b8HazoretMonOracle   = "f02a4cbf-f0bb-4bc7-99c2-6afa1b41ebb8"
	b8InfernalIdolOracle = "a0802918-e18a-4716-befb-96ae6667f2f3"
	b8NettlecystOracle   = "04c7f4fe-2098-4311-866d-6733c08d5178"
	b8RhonasMonOracle    = "b9a7c722-6b24-4383-ac6e-2a30f7834269"
	b8ShimmerMyrOracle   = "e110ee80-37af-4989-a5fb-95d51dc7f102"
	b8SpireGolemOracle   = "bfdc6bbc-7402-40ae-91be-658afca2f145"
	b8BarrenMoorOracle   = "326ba371-124c-4949-a048-3a0c8962e567"
	b8FoundryOracle      = "0498ea14-53d9-4655-a616-0ff3cf73de4e"
	b8HavenOracle        = "acc9c16a-5e72-43bd-87e1-56a16aa892f5"
)

// b7Put seeds a permanent with printed power and toughness (the catalog
// supplies abilities, not P/T) and fires its enter event.
func b7Put(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, t int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}

func b7Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Counters[game.CounterPlusOne]
}

func TestB7RhonasNeedsAnotherCreatureWithPowerFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rhonas := b7Put(g, me.ID, "Rhonas the Indomitable", "Legendary Creature — God", b7RhonasOracle, 5, 5)
	small := b7Put(g, me.ID, "Small Bear", "Creature — Bear", "test-small", 3, 3)
	_ = small
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(rhonas, opp.ID); err == nil {
		t.Fatal("Rhonas attacked with no other creature of power 4 or greater (his own 5 power doesn't count)")
	}
	b7Put(g, me.ID, "Big Bear", "Creature — Bear", "test-big", 4, 4)
	if err := g.DeclareAttacker(rhonas, opp.ID); err != nil {
		t.Fatalf("Rhonas refused with a power-4 creature out: %v", err)
	}
}

func TestB7RhonasPumpsAnotherCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rhonas := b7Put(g, me.ID, "Rhonas the Indomitable", "Legendary Creature — God", b7RhonasOracle, 5, 5)
	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 2, 2)
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}{G}"); err != nil {
		t.Fatal(err)
	}
	self := []game.TargetRef{{Kind: game.TargetCard, ID: rhonas}}
	if err := g.ActivateCatalogAbility(me.ID, rhonas, 0, game.ActivateAbilityParams{Targets: self, Strict: true}); err == nil {
		t.Fatal("Rhonas targeted himself with 'another target creature'")
	}
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, rhonas, 0, game.ActivateAbilityParams{Targets: tgt, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("Bear %d/%d, want 4/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("Bear did not gain trample")
	}
}

func TestB7RipjawRaptorDrawsOncePerDamageInstance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	raptor := b7Put(g, me.ID, "Ripjaw Raptor", "Creature — Dinosaur", b7RipjawOracle, 4, 5)
	a := b7Put(g, opp.ID, "Bear A", "Creature — Bear", "test-a", 2, 2)
	b := b7Put(g, opp.ID, "Bear B", "Creature — Bear", "test-b", 2, 2)
	before := me.Hand.Size()
	// Two sources dealing damage at once are one damage event (the ruling).
	g.WithWriteLock(func() {
		_ = g.DamageInstanceForEffect(func() error {
			_ = g.DealDamageToCreatureForEffect(a, raptor, 2)
			return g.DealDamageToCreatureForEffect(b, raptor, 2)
		})
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("hand +%d after two simultaneous sources, want exactly 1", got)
	}
	// Two separate instructions are two events.
	before = me.Hand.Size()
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(a, raptor, 1)
		_ = g.DealDamageToCreatureForEffect(b, raptor, 1)
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand +%d after two separate damage events, want 2", got)
	}
	// Zero damage (prevented) draws nothing.
	before = me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(a, raptor, 0) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 0 {
		t.Errorf("hand +%d after 0 damage, want 0", got)
	}
}

func TestB7SteelLeafChampionCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := b7Put(g, me.ID, "Steel Leaf Champion", "Creature — Elf Knight", b7SteelLeafOracle, 5, 4)
	small := b7Put(g, opp.ID, "Small", "Creature — Bear", "test-small", 2, 2)
	big := b7Put(g, opp.ID, "Big", "Creature — Bear", "test-big", 3, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(champ, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(small, champ); err == nil {
		t.Error("a power-2 creature blocked Steel Leaf Champion")
	}
	if err := g.DeclareBlocker(big, champ); err != nil {
		t.Errorf("a power-3 creature could not block: %v", err)
	}
}

func TestB7SurrakGivesTrampleAndCountersToNontokenEntrants(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	surrak := b7Put(g, me.ID, "Surrak and Goreclaw", "Legendary Creature — Human Bear", b7SurrakOracle, 6, 5)
	other := enterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if !hasEffectiveKeyword(t, g, other, "trample") {
		t.Error("another creature you control lacks trample")
	}
	passPriorityAroundTable(t, g)
	if got := b7Counters(t, g, other); got != 1 {
		t.Errorf("entering nontoken creature has %d +1/+1 counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, other, "haste") {
		t.Error("entering creature did not gain haste")
	}
	if got := b7Counters(t, g, surrak); got != 0 {
		t.Errorf("Surrak put a counter on itself: %d", got)
	}
	// A token does not trigger it.
	tok := enterCard(t, g, me.ID, game.Card{Name: "Bear Token", TypeLine: "Token Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := b7Counters(t, g, tok); got != 0 {
		t.Errorf("a token got %d counters", got)
	}
	// An opponent's creature does not.
	theirs := enterCard(t, g, g.Seats[1].ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if got := b7Counters(t, g, theirs); got != 0 || hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Errorf("an opponent's creature was affected: counters %d", got)
	}
}

func TestB7DarksteelJuggernautCountsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jug := b7Put(g, me.ID, "Darksteel Juggernaut", "Artifact Creature — Juggernaut", b7DarksteelJugOracle, 0, 0)
	if p, tough := effectivePTOf(t, g, jug); p != 1 || tough != 1 {
		t.Errorf("alone: %d/%d, want 1/1 (it is an artifact)", p, tough)
	}
	b7Put(g, me.ID, "Rock", "Artifact", "test-rock", 0, 0)
	b7Put(g, me.ID, "Rock 2", "Artifact", "test-rock2", 0, 0)
	b7Put(g, opp.ID, "Their Rock", "Artifact", "test-their-rock", 0, 0)
	if p, tough := effectivePTOf(t, g, jug); p != 3 || tough != 3 {
		t.Errorf("with two more artifacts: %d/%d, want 3/3 (opponents' don't count)", p, tough)
	}
	if !hasEffectiveKeyword(t, g, jug, "indestructible") {
		t.Error("not indestructible")
	}
}

func TestB7DarksteelJuggernautMustAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	jug := b7Put(g, me.ID, "Darksteel Juggernaut", "Artifact Creature — Juggernaut", b7DarksteelJugOracle, 0, 0)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(jug, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != 40-1 {
		t.Errorf("opponent life %d, want 39", opp.Life)
	}
}

func TestB8GraazMakesJuggernauts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	graaz := b7Put(g, me.ID, "Graaz, Unstoppable Juggernaut", "Legendary Artifact Creature — Juggernaut", b8GraazOracle, 7, 5)
	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 2, 2)
	theirs := b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-their", 2, 2)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("my Bear %d/%d, want 5/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, graaz); p != 7 || tough != 5 {
		t.Errorf("Graaz %d/%d, want 7/5", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear %d/%d, want 2/2", p, tough)
	}
	c, _ := g.LookupCardForEffect(bear)
	if !c.HasSubtype("Juggernaut") || !c.HasSubtype("Bear") {
		t.Errorf("Bear's type line %v, want Bear and Juggernaut", c.TypeLine)
	}
}

func TestB8GraazJuggernautsAttackAndCantBeBlockedByWalls(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	graaz := b7Put(g, me.ID, "Graaz, Unstoppable Juggernaut", "Legendary Artifact Creature — Juggernaut", b8GraazOracle, 7, 5)
	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 2, 2)
	wall := b7Put(g, opp.ID, "Wall", "Creature — Wall", "test-wall", 0, 5)
	adv := b7Put(g, opp.ID, "Guard", "Creature — Human", "test-guard", 1, 5)
	advanceTo(t, g, game.StepDeclareAttackers)
	// Graaz and the Bear are Juggernauts: neither may stay home.
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatal(err)
	}
	if err := g.DeclareAttacker(graaz, opp.ID); err != nil {
		t.Fatal(err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err == nil {
		t.Error("a Wall blocked a Juggernaut")
	}
	if err := g.DeclareBlocker(adv, bear); err != nil {
		t.Errorf("a non-Wall could not block: %v", err)
	}
}

func TestB7YevaGivesGreenCreaturesFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b7Put(g, me.ID, "Yeva, Nature's Herald", "Legendary Creature — Elf Shaman", b7YevaOracle, 4, 4)
	advanceTo(t, g, game.StepUpkeep)
	cast := func(name, typeLine, cost string, colors []string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
			Owner: me.ID, Controller: me.ID,
		})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := cast("Green Bear", "Creature — Bear", "{1}{G}", []string{"G"}); err != nil {
		t.Errorf("green creature in the upkeep: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}"); err != nil {
		t.Fatal(err)
	}
	if err := cast("Blue Bear", "Creature — Bear", "{1}{U}", []string{"U"}); err == nil {
		t.Error("a blue creature was cast in the upkeep")
	}
}

func TestB7TangleweaveArmorReadsTheGreatestCommanderManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cmdr One", TypeLine: "Legendary Creature — Elf", ManaCost: "{2}{G}",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
		me.Command.PushTop(game.Card{
			InstanceID: uuid.New(), Name: "Cmdr Two", TypeLine: "Legendary Creature — Elf", ManaCost: "{4}{G}{G}",
			Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
	})
	_, germ := castLivingWeapon(t, g, "Tangleweave Armor", b7TangleweaveOracle)
	if p, tough := effectivePTOf(t, g, germ); p != 6 || tough != 6 {
		t.Errorf("Germ %d/%d, want 6/6 (the greatest of 3 and 6, from the command zone)", p, tough)
	}
}

func TestB7TangleweaveArmorWithNoCommanderLeavesADeadGerm(t *testing.T) {
	g := newCatalogGame(t)
	g.WithWriteLock(func() {
		for _, s := range g.Seats {
			s.Command.Cards = nil
		}
	})
	castCatalogSpell(t, g, "Tangleweave Armor", "Artifact — Equipment", b7TangleweaveOracle, nil)
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if findBattlefieldByName(g, "Phyrexian Germ") != uuid.Nil {
		t.Error("the 0/0 Germ survived with no commander")
	}
}

func TestB8NettlecystCountsArtifactsAndEnchantmentsOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, germ := castLivingWeapon(t, g, "Nettlecyst", b8NettlecystOracle)
	if p, tough := effectivePTOf(t, g, germ); p != 1 || tough != 1 {
		t.Errorf("Germ %d/%d, want 1/1 (Nettlecyst is an artifact)", p, tough)
	}
	b7Put(g, me.ID, "Rock", "Artifact", "test-rock", 0, 0)
	b7Put(g, me.ID, "Aura Thing", "Enchantment", "test-ench", 0, 0)
	b7Put(g, me.ID, "Both", "Enchantment Artifact", "test-both", 0, 0)
	b7Put(g, me.ID, "Plain Bear", "Creature — Bear", "test-bear", 2, 2)
	b7Put(g, opp.ID, "Their Rock", "Artifact", "test-their", 0, 0)
	if p, tough := effectivePTOf(t, g, germ); p != 4 || tough != 4 {
		t.Errorf("Germ %d/%d, want 4/4 (Nettlecyst, Rock, Aura Thing, Both once)", p, tough)
	}
}

func TestB8InfernalIdolMakesBlackMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := b7Put(g, me.ID, "Infernal Idol", "Artifact", b8InfernalIdolOracle, 0, 0)
	if err := g.ActivateManaAbility(me.ID, idol, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("pool %v, want [B]", got)
	}
}

func TestB8InfernalIdolSacrificesForTwoCardsAndTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	idol := b7Put(g, me.ID, "Infernal Idol", "Artifact", b8InfernalIdolOracle, 0, 0)
	advanceToMain(t, g)
	hand, life := me.Hand.Size(), me.Life
	// Not enough mana: refused, nothing sacrificed.
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying {1}{B}{B}")
	}
	if !g.Battlefield.Contains(idol) {
		t.Fatal("a refused activation sacrificed the Idol")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(idol) {
		t.Error("the Idol was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand +%d, want 2", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestB8HazoretsMonumentDiscountsRedCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	b7Put(g, me.ID, "Hazoret's Monument", "Legendary Artifact", b8HazoretMonOracle, 0, 0)
	if got := priceInHand(t, g, me, "Red Bear", "Creature — Bear", "{1}{R}"); got != 1 {
		t.Errorf("red creature spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Green Bear", "Creature — Bear", "{1}{G}"); got != 2 {
		t.Errorf("green creature spell: %d, want 2", got)
	}
	if got := priceInHand(t, g, me, "Red Instant", "Instant", "{1}{R}"); got != 2 {
		t.Errorf("red noncreature spell: %d, want 2", got)
	}
	if got := priceInHand(t, g, them, "Their Red Bear", "Creature — Bear", "{1}{R}"); got != 2 {
		t.Errorf("opponent's red creature spell: %d, want 2", got)
	}
}

func TestB8HazoretsMonumentLootsOnAnyCreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b7Put(g, me.ID, "Hazoret's Monument", "Legendary Artifact", b8HazoretMonOracle, 0, 0)
	castCatalogSpell(t, g, "Green Bear", "Creature — Bear", "test-green-bear", nil)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the Monument offered no discard")
	}
	hand := me.Hand.Size()
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want a discard and a draw", hand, me.Hand.Size())
	}
}

func TestB8RhonasMonumentDiscountsGreenAndPumpsAnyCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b7Put(g, me.ID, "Rhonas's Monument", "Legendary Artifact", b8RhonasMonOracle, 0, 0)
	bear := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 2, 2)
	if got := priceInHand(t, g, me, "Green Bear", "Creature — Bear", "{1}{G}"); got != 1 {
		t.Errorf("green creature spell: %d, want 1", got)
	}
	if got := priceInHand(t, g, me, "Red Bear", "Creature — Bear", "{1}{R}"); got != 2 {
		t.Errorf("red creature spell: %d, want 2", got)
	}
	b7Put(g, me.ID, "Other Bear", "Creature — Bear", "test-other", 1, 1)
	castCatalogSpell(t, g, "Red Bear", "Creature — Bear", "test-red-bear", nil)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("Bear has no trample")
	}
}

func TestB7WhispererOfTheWildsFerociousNeedsABigCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := b7Put(g, me.ID, "Whisperer of the Wilds", "Creature — Human Shaman", b7WhispererOracle, 0, 2)
	if err := g.ActivateManaAbility(me.ID, w, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := batch01PoolColors(me); len(got) != 1 || got[0] != "G" {
		t.Errorf("pool %v, want [G]", got)
	}
	me.ManaPool = nil
	w2 := b7Put(g, me.ID, "Whisperer of the Wilds", "Creature — Human Shaman", b7WhispererOracle, 0, 2)
	if err := g.ActivateManaAbility(me.ID, w2, 1, game.ManaAbilityParams{}); !errors.Is(err, game.ErrConditionNotMet) {
		t.Fatalf("err = %v, want ErrConditionNotMet", err)
	}
	if c, ok := g.LookupCardForEffect(w2); !ok || c.Tapped {
		t.Fatal("a failed gate tapped the source")
	}
	b7Put(g, me.ID, "Big Bear", "Creature — Bear", "test-big", 4, 4)
	if err := g.ActivateManaAbility(me.ID, w2, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if got := batch01PoolColors(me); len(got) != 2 {
		t.Errorf("pool %v, want {G}{G}", got)
	}
}

func TestB8ShimmerMyrGivesArtifactSpellsFlash(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b7Put(g, me.ID, "Shimmer Myr", "Artifact Creature — Myr", b8ShimmerMyrOracle, 2, 2)
	advanceTo(t, g, game.StepUpkeep)
	cast := func(name, typeLine string) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := cast("Rock", "Artifact"); err != nil {
		t.Errorf("artifact in the upkeep: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatal(err)
	}
	if err := cast("Sorcery", "Sorcery"); err == nil {
		t.Error("a sorcery was cast in the upkeep")
	}
}

func TestB8SpireGolemHasAffinityForIslands(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	price := func() int {
		base, _ := game.ParseCost("{6}")
		out, err := g.ApplyCostModifiers(base, game.CostQuery{
			Card: game.Card{InstanceID: uuid.New(), Name: "Spire Golem", TypeLine: "Artifact Creature — Golem",
				ManaCost: "{6}", OracleID: b8SpireGolemOracle, Owner: me.ID, Controller: me.ID},
			Controller: me.ID, FromZone: game.ZoneHand,
		})
		if err != nil {
			t.Fatal(err)
		}
		return out.ManaValue()
	}
	if got := price(); got != 6 {
		t.Fatalf("no Islands: %d, want 6", got)
	}
	b7Put(g, me.ID, "Island", "Basic Land — Island", "test-island", 0, 0)
	b7Put(g, me.ID, "Island 2", "Basic Land — Island", "test-island2", 0, 0)
	b7Put(g, me.ID, "Forest", "Basic Land — Forest", "test-forest", 0, 0)
	b7Put(g, g.Seats[1].ID, "Their Island", "Basic Land — Island", "test-their-island", 0, 0)
	if got := price(); got != 4 {
		t.Errorf("two Islands: %d, want 4", got)
	}
}

func TestB8BarrenMoorEntersTappedAndCycles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {})
	id, _ := cycleFromHand(t, g, "Barren Moor", "Land", b8BarrenMoorOracle, "{B}")
	_ = id
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: b8BarrenMoorOracle})
	if len(abilities) != 1 {
		t.Errorf("%d activated abilities, want the one cycling ability", len(abilities))
	}
	_ = me
}

func TestB8FoundryOfTheConsulsMakesTwoThopters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	foundry := b7Put(g, me.ID, "Foundry of the Consuls", "Land", b8FoundryOracle, 0, 0)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {5}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, foundry, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(foundry) {
		t.Error("Foundry not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := countTokensControlled(g, me.ID, "Thopter"); got != 2 {
		t.Fatalf("%d Thopters, want 2", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" {
			if c.Power != 1 || c.Toughness != 1 || !c.HasCardType("artifact") || !c.IsColorless() {
				t.Errorf("Thopter %d/%d %s, want a colorless 1/1 artifact creature", c.Power, c.Toughness, c.TypeLine)
			}
			if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("Thopter has no flying")
			}
		}
	}
}

func TestB8HavenOfTheSpiritDragonReturnsDragonsAndUgin(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	haven := b7Put(g, me.ID, "Haven of the Spirit Dragon", "Land", b8HavenOracle, 0, 0)
	dragon := seedGraveyardCard(t, g, "Shiv Dragon", "Creature — Dragon", "test-dragon")
	bear := seedGraveyardCard(t, g, "Bear", "Creature — Bear", "test-bear")
	ugin := seedGraveyardCard(t, g, "Ugin, the Spirit Dragon", "Legendary Planeswalker — Ugin", "test-ugin")
	advanceToMain(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	bad := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	if err := g.ActivateCatalogAbility(me.ID, haven, 0, game.ActivateAbilityParams{Targets: bad, Strict: true}); err == nil {
		t.Fatal("targeted a non-Dragon creature card")
	}
	_ = dragon
	_ = ugin
	good := []game.TargetRef{{Kind: game.TargetCard, ID: ugin}}
	if err := g.ActivateCatalogAbility(me.ID, haven, 0, game.ActivateAbilityParams{Targets: good, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ugin) {
		t.Error("Ugin did not return to hand")
	}
	if g.Battlefield.Contains(haven) {
		t.Error("Haven not sacrificed")
	}
}

func TestB7WhiptongueHydraDestroysFliersAndCountsThem(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f1 := enterCard(t, g, opp.ID, game.Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	f2 := enterCard(t, g, me.ID, game.Card{Name: "Own Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	ground := enterCard(t, g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	hydra := enterCard(t, g, me.ID, game.Card{Name: "Whiptongue Hydra", TypeLine: "Creature — Lizard Hydra", OracleID: b7WhiptongueOracle, Power: 4, Toughness: 4})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(f1) || g.Battlefield.Contains(f2) {
		t.Error("a flier survived")
	}
	if !g.Battlefield.Contains(ground) || !g.Battlefield.Contains(hydra) {
		t.Error("a non-flier was destroyed")
	}
	if got := b7Counters(t, g, hydra); got != 2 {
		t.Errorf("Hydra has %d counters, want 2", got)
	}
}

func TestB7ThickestPutsCountersEqualToPowerAndDrawsForTheBiggest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 3, 3)
	theirs := b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-their", 5, 5)
	castCatalogSpell(t, g, "Thickest in the Thicket", "Enchantment", b7ThickestOracle, nil)
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, mine)
	}
	passPriorityAroundTable(t, g)
	if got := b7Counters(t, g, mine); got != 3 {
		t.Fatalf("Bear has %d counters, want 3 (its power)", got)
	}
	_ = theirs
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand +%d at the end step with a 6-power creature against a 5, want 2", got)
	}
}

func TestB7ThickestDrawsNothingWhenAnOpponentHasTheBiggest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b7Put(g, me.ID, "Thickest in the Thicket", "Enchantment", b7ThickestOracle, 0, 0)
	b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 3, 3)
	b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-their", 5, 5)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand +%d, want 0 (an opponent has the greatest power)", got)
	}
}

func TestB7ThickestDrawsOnATieAndNotWithoutCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b7Put(g, me.ID, "Thickest in the Thicket", "Enchantment", b7ThickestOracle, 0, 0)
	b7Put(g, me.ID, "Bear", "Creature — Bear", "test-bear", 4, 4)
	b7Put(g, opp.ID, "Their Bear", "Creature — Bear", "test-their", 4, 4)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand +%d on a tie for greatest power, want 2", got)
	}
}

func TestB7ThickestDrawsNothingWithNoCreatureOfYours(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b7Put(g, me.ID, "Thickest in the Thicket", "Enchantment", b7ThickestOracle, 0, 0)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("hand +%d with no creatures anywhere, want 0", got)
	}
}
