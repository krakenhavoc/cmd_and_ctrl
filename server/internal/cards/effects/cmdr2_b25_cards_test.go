package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	c25SkinrenderOracle  = "c5c225a6-6a5e-4905-9412-dad15ba645b4"
	c25IncrementalOracle = "159cd205-6afc-4d8c-a4b9-501911a9c345"
	c25SnuffersOracle    = "cdd227fc-0809-45d6-bed0-e85c0c0b6c94"
	c25BansheeOracle     = "dd8b6067-a97f-402f-bba1-70ebc6ee1599"
	c25GrimPoppetOracle  = "43e0dbaf-3c81-47e2-b0c2-364f7ac7295f"
	c25CarnifexOracle    = "527e4eec-0983-46a9-a4f7-451080200aea"
	c25WickerboughOracle = "3d794629-9014-460a-8d3b-785eadc1cdb5"
	c25ChannelerOracle   = "a98ea782-1c49-4207-a130-41329087fdf6"
	c25UrchinsOracle     = "1c2cc166-dc59-44ce-86d5-ed6013b0bf67"
	c25KulrathOracle     = "f3d1b2e9-6ede-416d-937b-2063ea278d2f"
	c25AberrantOracle    = "3a92b235-196b-4f46-9d20-06f4d3653d36"
)

func TestCmdr2B25EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		c25SkinrenderOracle:  "Skinrender",
		c25IncrementalOracle: "Incremental Blight",
		c25SnuffersOracle:    "Soul Snuffers",
		c25BansheeOracle:     "Midnight Banshee",
		c25GrimPoppetOracle:  "Grim Poppet",
		c25CarnifexOracle:    "Carnifex Demon",
		c25WickerboughOracle: "Wickerbough Elder",
		c25ChannelerOracle:   "Channeler Initiate",
		c25UrchinsOracle:     "Dusk Urchins",
		c25KulrathOracle:     "Kulrath Knight",
		c25AberrantOracle:    "Aberrant Return",
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

func TestCmdr2B25SkinrenderPutsThreeCountersOnTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Skinrender", "Creature — Phyrexian Zombie", c25SkinrenderOracle, nil)
	passPriorityAroundTable(t, g)
	pickCard(t, g, g.Seats[0].ID, theirs)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, theirs, game.CounterMinusOne); got != 3 {
		t.Errorf("-1/-1 counters on the target = %d, want 3", got)
	}
	if p, tt := ptOf(t, g, theirs); p != 1 || tt != 1 {
		t.Errorf("the target is %d/%d, want 1/1", p, tt)
	}
}

func TestCmdr2B25IncrementalBlightPlacesOneTwoThree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Ogre", 5, 5)
	b := b12Creature(g, opp.ID, "B", "Creature — Ogre", 5, 5)
	c := b12Creature(g, opp.ID, "C", "Creature — Ogre", 5, 5)
	castCatalogSpell(t, g, "Incremental Blight", "Sorcery", c25IncrementalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}})
	passPriorityAroundTable(t, g)
	for id, want := range map[any]int{a: 1, b: 2, c: 3} {
		_ = id
		_ = want
	}
	if countersOn(g, a, game.CounterMinusOne) != 1 || countersOn(g, b, game.CounterMinusOne) != 2 || countersOn(g, c, game.CounterMinusOne) != 3 {
		t.Errorf("counters = %d/%d/%d, want 1/2/3", countersOn(g, a, game.CounterMinusOne), countersOn(g, b, game.CounterMinusOne), countersOn(g, c, game.CounterMinusOne))
	}
}

func TestCmdr2B25IncrementalBlightRefusesTheSameCreatureTwice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Ogre", 5, 5)
	b := b12Creature(g, opp.ID, "B", "Creature — Ogre", 5, 5)
	err := castCatalogSpellErr(t, g, "Incremental Blight", "Sorcery", c25IncrementalOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	if err == nil {
		t.Fatal("a creature may not fill two target slots")
	}
}

func TestCmdr2B25SoulSnuffersShrinksEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Elf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Elf", 3, 3)
	snuffers := c25CastCreature(t, g, "Soul Snuffers", "Creature — Elemental Shaman", c25SnuffersOracle, 3, 3)
	passPriorityAroundTable(t, g)
	for _, id := range []any{mine, theirs} {
		_ = id
	}
	if countersOn(g, mine, game.CounterMinusOne) != 1 || countersOn(g, theirs, game.CounterMinusOne) != 1 {
		t.Error("every creature gets a -1/-1 counter")
	}
	if countersOn(g, snuffers, game.CounterMinusOne) != 1 {
		t.Error("the Snuffers counts itself")
	}
}

func TestCmdr2B25MidnightBansheeShrinksNonblackOnYourUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Midnight Banshee", "Creature — Spirit", c25BansheeOracle, 5, 5)
	green := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Elf", TypeLine: "Creature — Elf",
		Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID})
	black := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Zombie", TypeLine: "Creature — Zombie",
		Power: 2, Toughness: 2, Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	advanceToUpkeepOfSeat(t, g, 1)
	passPriorityAroundTable(t, g)
	if countersOn(g, green, game.CounterMinusOne) != 0 {
		t.Error("an opponent's upkeep does not trigger it")
	}
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if countersOn(g, green, game.CounterMinusOne) != 1 {
		t.Error("the nonblack creature gets a counter")
	}
	if countersOn(g, black, game.CounterMinusOne) != 0 {
		t.Error("a black creature gets none")
	}
}

// c25CastCreature casts a catalog creature with its printed power and
// toughness from the active seat's hand, with no mana check.
func c25CastCreature(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func c25Activate(t *testing.T, g *game.Game, id uuid.UUID, idx int, target uuid.UUID) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	p := game.ActivateAbilityParams{CounterKind: game.CounterMinusOne}
	if target != uuid.Nil {
		p.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: target}}
	}
	return g.ActivateCatalogAbility(me.ID, id, idx, p)
}

func TestCmdr2B25GrimPoppetEntersWithThreeAndSpendsThem(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	poppet := c25CastCreature(t, g, "Grim Poppet", "Artifact Creature — Scarecrow", c25GrimPoppetOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, poppet, game.CounterMinusOne); got != 3 {
		t.Fatalf("Grim Poppet entered with %d counters, want 3", got)
	}
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	if err := c25Activate(t, g, poppet, 0, poppet); err == nil {
		t.Error("the target must be another creature")
	}
	if err := c25Activate(t, g, poppet, 0, victim); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, poppet, game.CounterMinusOne); got != 2 {
		t.Errorf("Poppet counters = %d, want 2", got)
	}
	if got := countersOn(g, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("victim counters = %d, want 1", got)
	}
}

func TestCmdr2B25CarnifexDemonShrinksEveryOtherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	demon := c25CastCreature(t, g, "Carnifex Demon", "Creature — Phyrexian Demon", c25CarnifexOracle, 6, 6)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, demon, game.CounterMinusOne); got != 2 {
		t.Fatalf("Carnifex Demon entered with %d counters, want 2", got)
	}
	mine := b12Creature(g, me.ID, "Mine", "Creature — Elf", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Elf", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, demon, 0, game.ActivateAbilityParams{CounterKind: game.CounterMinusOne}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, mine, game.CounterMinusOne) != 1 || countersOn(g, theirs, game.CounterMinusOne) != 1 {
		t.Error("each other creature gets a counter")
	}
	if got := countersOn(g, demon, game.CounterMinusOne); got != 1 {
		t.Errorf("the Demon has %d counters, want 1 (one paid, none added)", got)
	}
}

func TestCmdr2B25WickerboughElderDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := c25CastCreature(t, g, "Wickerbough Elder", "Creature — Treefolk Shaman", c25WickerboughOracle, 4, 4)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, elder, game.CounterMinusOne); got != 1 {
		t.Fatalf("entered with %d counters, want 1", got)
	}
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := c25Activate(t, g, elder, 0, bear); err == nil {
		t.Fatal("a plain creature is not a legal target")
	}
	if err := c25Activate(t, g, elder, 0, rock); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
	if got := countersOn(g, elder, game.CounterMinusOne); got != 0 {
		t.Errorf("counters left = %d, want 0", got)
	}
}

func TestCmdr2B25ChannelerInitiatePutsThreeOnYourCreatureAndMakesMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Elf", 5, 5)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Elf", 5, 5)
	_ = c25CastCreature(t, g, "Channeler Initiate", "Creature — Human Druid", c25ChannelerOracle, 3, 4)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		if hasID(p.PickTargetCards, theirs) {
			t.Error("an opponent's creature is not a creature you control")
		}
		pickCard(t, g, me.ID, mine)
		passPriorityAroundTable(t, g)
	}
	if got := countersOn(g, mine, game.CounterMinusOne); got != 3 {
		t.Errorf("target has %d counters, want 3", got)
	}
	if countersOn(g, theirs, game.CounterMinusOne) != 0 {
		t.Error("the opponent's creature is untouched")
	}
	// A creature that has been under its controller's control since the
	// turn began pays {T}.
	vet := b12Push(g, me.ID, "Channeler Initiate", "Creature — Human Druid", c25ChannelerOracle, 3, 4)
	addCountersForTest(g, vet, game.CounterMinusOne, 1)
	initiate := vet
	if err := g.ActivateManaAbility(me.ID, initiate, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := countersOn(g, initiate, game.CounterMinusOne); got != 0 {
		t.Errorf("counters left = %d, want 0", got)
	}
	if err := g.ActivateManaAbility(me.ID, initiate, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("no counter and tapped: it cannot be paid again")
	}
}

func TestCmdr2B25DuskUrchinsShrinksWhenItAttacksAndDrawsWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	urchins := b12Push(g, me.ID, "Dusk Urchins", "Creature — Ouphe", c25UrchinsOracle, 4, 3)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(urchins, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, urchins, game.CounterMinusOne); got != 1 {
		t.Fatalf("counters after attacking = %d, want 1", got)
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(urchins))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(urchins) {
		t.Fatal("the Urchins survived a Bolt")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for its one counter", hand, got)
	}
}

func TestCmdr2B25KulrathKnightHoldsBackOpponentsCreaturesWithCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, opp.ID, "Kulrath Knight", "Creature — Elemental Knight", c25KulrathOracle, 3, 3)
	marked := pushVanillaCreature(g, me.ID, "Marked", 2, 2)
	free := pushVanillaCreature(g, me.ID, "Free", 2, 2)
	addCountersForTest(g, marked, game.CounterPlusOne, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(marked, opp.ID); err == nil {
		t.Error("a creature with a counter cannot attack")
	}
	if err := g.DeclareAttacker(free, opp.ID); err != nil {
		t.Errorf("a creature without counters attacks: %v", err)
	}
}

func TestCmdr2B25KulrathKnightDoesNotHoldBackItsOwnersCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kulrath Knight", "Creature — Elemental Knight", c25KulrathOracle, 3, 3)
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	addCountersForTest(g, mine, game.CounterPlusOne, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Errorf("your own creature with a counter attacks: %v", err)
	}
}

func TestCmdr2B25AberrantReturnReanimatesWithAMinusCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Creature — Elf", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	castCatalogSpell(t, g, "Aberrant Return", "Sorcery", c25AberrantOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		c := battlefieldCardFor(g, id)
		if c == nil {
			t.Fatalf("%s did not return", id)
		}
		if c.Controller != me.ID {
			t.Errorf("%s is under %s, want the caster", c.Name, c.Controller)
		}
		if got := countersOn(g, id, game.CounterMinusOne); got != 1 {
			t.Errorf("%s has %d -1/-1 counters, want 1", c.Name, got)
		}
	}
	if p, tt := ptOf(t, g, b); p != 3 || tt != 3 {
		t.Errorf("the Ogre is %d/%d, want 3/3", p, tt)
	}
}

func TestCmdr2B25AberrantReturnNeedsOneToThreeCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	var ids []game.TargetRef
	for i := 0; i < 4; i++ {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "X", TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		ids = append(ids, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	if err := castCatalogSpellErr(t, g, "Aberrant Return", "Sorcery", c25AberrantOracle, ids); err == nil {
		t.Error("four targets are one too many")
	}
	if err := castCatalogSpellErr(t, g, "Aberrant Return", "Sorcery", c25AberrantOracle, nil); err == nil {
		t.Error("it needs at least one target")
	}
}
