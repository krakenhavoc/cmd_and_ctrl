package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b15_cards_test.go — Kamigawa: Neon Dynasty, batch 15 (S82, #5493).
// Each test plays the real catalog entry through the engine.

const (
	neoB15Voltage   = "1779af8e-38fc-4043-b5fa-ee16a7ea840c"
	neoB15Commune   = "68f8ed23-024e-414f-bbb2-ab73fbe21b69"
	neoB15Seismic   = "6a56e2ef-75dd-47d5-b61c-11faaba8cd9a"
	neoB15Return    = "9c002b4d-bb6c-45a5-88e1-1df2958dff9f"
	neoB15Storyweav = "94389047-4976-4be2-aaca-25dcdc538abe"
	neoB15Kick      = "ed37ce4e-d84e-442a-ac4e-0512f358c138"
	neoB15Access    = "353dd446-2ea6-41c9-a10f-5a03e4de19b4"
	neoB15Discover  = "1c2a3f13-6bcb-42ad-98f6-642151ad54f1"
	neoB15Smoke     = "8fef58ff-4743-4065-9df0-c5edf9aa2502"
	neoB15Collision = "40cbf2c0-041a-44df-a775-580012caab67"
	neoB15Invoke    = "f9ee4c80-20fb-4322-be02-1c78d37d7a99"
)

func TestNeoB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		neoB15Voltage:   "Voltage Surge",
		neoB15Commune:   "Commune with Spirits",
		neoB15Seismic:   "Seismic Wave",
		neoB15Return:    "Return to Action",
		neoB15Storyweav: "Storyweave",
		neoB15Kick:      "Spinning Wheel Kick",
		neoB15Access:    "Access Denied",
		neoB15Discover:  "Discover the Impossible",
		neoB15Smoke:     "Smoke Spirits' Aid",
		neoB15Collision: "Collision of Realms",
		neoB15Invoke:    "Invoke Despair",
	} {
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

// neoB15Cast casts the named card strictly from seat `active`'s hand.
func neoB15Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	params.Strict = true
	return castFromHandForTest(t, g, me, name, typeLine, cost, oracle, params)
}

func neoB15Refused(g *game.Game, name, typeLine, cost, oracle string, params game.CastSpellParams) error {
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	params.Strict = true
	return g.CastSpell(me.ID, id, params)
}

func neoB15DamageOn(g *game.Game, id uuid.UUID) int {
	c, ok := battlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.DamageMarked
}

func TestNeoB15VoltageSurge(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	rock := b12Push(g, me.ID, "Mind Stone", "Artifact", "", 0, 0)
	land := b12Push(g, me.ID, "Mountain", "Basic Land — Mountain", "", 0, 0)

	b06AddMana(me, "R")
	if err := neoB15Refused(g, "Voltage Surge", "Instant", "{R}", neoB15Voltage,
		game.CastSpellParams{Targets: playerRefs(opp.ID)}); err == nil {
		t.Error("cast at a player")
	}
	b06AddMana(me, "R")
	if err := neoB15Refused(g, "Voltage Surge", "Instant", "{R}", neoB15Voltage,
		game.CastSpellParams{Targets: cardRefs(bear), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{land}}); err == nil {
		t.Error("the additional cost was paid by sacrificing a land")
	}

	// Without the sacrifice: 2 damage, so a 3/3 survives it.
	b06AddMana(me, "R")
	neoB15Cast(t, g, "Voltage Surge", "Instant", "{R}", neoB15Voltage, game.CastSpellParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if got := neoB15DamageOn(g, bear); got != 2 {
		t.Errorf("unpaid: %d damage marked, want 2", got)
	}
	if findBattlefieldCardForTest(g, rock) == nil {
		t.Error("the artifact was sacrificed without being asked")
	}

	// With it: 4 instead (not 2 + 4), which kills the 4/4 outright.
	b06AddMana(me, "R")
	neoB15Cast(t, g, "Voltage Surge", "Instant", "{R}", neoB15Voltage,
		game.CastSpellParams{Targets: cardRefs(ogre), OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{rock}})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if findBattlefieldCardForTest(g, rock) != nil {
		t.Error("the artifact was not sacrificed")
	}
	if findBattlefieldCardForTest(g, ogre) != nil {
		t.Error("4 damage did not kill the 4/4")
	}
}

func TestNeoB15SeismicWave(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := b12Creature(g, opp.ID, "Spirit", "Creature — Spirit", 1, 1)
	big := b12Creature(g, opp.ID, "Giant", "Creature — Giant", 3, 3)
	golem := b12Creature(g, opp.ID, "Golem", "Artifact Creature — Golem", 1, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 1, 1)
	life := opp.Life

	b06AddMana(me, "R", "C", "C")
	if err := neoB15Refused(g, "Seismic Wave", "Instant", "{2}{R}", neoB15Seismic,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetPlayer, ID: me.ID, Slot: 1}}}); err == nil {
		t.Error("the second target was yourself, not an opponent")
	}
	b06AddMana(me, "R", "C", "C")
	neoB15Cast(t, g, "Seismic Wave", "Instant", "{2}{R}", neoB15Seismic, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}}})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 less", life, opp.Life)
	}
	if findBattlefieldCardForTest(g, small) != nil {
		t.Error("the 1/1 nonartifact creature survived")
	}
	if got := neoB15DamageOn(g, big); got != 1 {
		t.Errorf("the 3/3 has %d damage, want 1", got)
	}
	if got := neoB15DamageOn(g, golem); got != 0 {
		t.Errorf("the artifact creature has %d damage, want none", got)
	}
	if got := neoB15DamageOn(g, mine); got != 0 {
		t.Errorf("your own creature has %d damage, want none", got)
	}
}

func TestNeoB15ReturnToAction(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	b06AddMana(me, "B", "C")
	if err := neoB15Refused(g, "Return to Action", "Instant", "{1}{B}", neoB15Return, game.CastSpellParams{}); err == nil {
		t.Error("cast with no target")
	}
	b06AddMana(me, "B", "C")
	neoB15Cast(t, g, "Return to Action", "Instant", "{1}{B}", neoB15Return, game.CastSpellParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("%d/%d, want 3/2", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, bear, "lifelink") {
		t.Error("the creature did not gain lifelink")
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, bear, 5); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	back, ok := battlefieldCard(g, bear)
	if !ok {
		t.Fatal("the creature did not return to the battlefield")
	}
	if !back.Tapped {
		t.Error("it returned untapped")
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("returned %d/%d, want the bare 2/2 (the effect ended with the old object)", p, tough)
	}
}

func TestNeoB15CommuneWithSpirits(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	// Library cards are listed bottom → top: the last is the top card.
	under := discoverLibraryCard(me.ID, "Deep Card", "Creature — Bear", "{1}{G}")
	c4 := discoverLibraryCard(me.ID, "Fourth", "Creature — Bear", "{1}{G}")
	ench := discoverLibraryCard(me.ID, "Pacifism", "Enchantment — Aura", "{1}{W}")
	forest := discoverLibraryCard(me.ID, "Forest", "Basic Land — Forest", "")
	c1 := discoverLibraryCard(me.ID, "First", "Creature — Bear", "{1}{G}")
	me.Library.Cards = []game.Card{under, c4, ench, forest, c1}

	b06AddMana(me, "G")
	neoB15Cast(t, g, "Commune with Spirits", "Sorcery", "{G}", neoB15Commune, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was raised")
	}
	if len(pick.ChooseCards) != 2 || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("offered %v (min %d max %d), want the land and the enchantment, up to one", pick.ChooseCards, pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ench.InstanceID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	if !inHandOf(me, ench.InstanceID) {
		t.Error("the enchantment is not in hand")
	}
	// The other three looked-at cards went to the bottom; the card that
	// was fifth is now the top.
	passPriorityAroundTable(t, g)
	if n := me.Library.Size(); n != 4 {
		t.Fatalf("library %d, want 4", n)
	}
	top := me.Library.Cards[len(me.Library.Cards)-1]
	if top.InstanceID != under.InstanceID {
		t.Errorf("the top of the library is %q, want the card below the four looked at", top.Name)
	}
}

func TestNeoB15CommuneWithSpiritsMayTakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a := discoverLibraryCard(me.ID, "A", "Basic Land — Forest", "")
	b := discoverLibraryCard(me.ID, "B", "Creature — Bear", "{1}{G}")
	c := discoverLibraryCard(me.ID, "C", "Creature — Bear", "{1}{G}")
	d := discoverLibraryCard(me.ID, "D", "Creature — Bear", "{1}{G}")
	me.Library.Cards = []game.Card{a, b, c, d}
	hand := me.Hand.Size()
	b06AddMana(me, "G")
	neoB15Cast(t, g, "Commune with Spirits", "Sorcery", "{G}", neoB15Commune, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was raised")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d → %d, want unchanged", hand, me.Hand.Size())
	}
	if me.Library.Size() != 4 {
		t.Errorf("library %d, want all four back on the bottom", me.Library.Size())
	}
}

func TestNeoB15SpinningWheelKick(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	biter := b12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	a := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	b := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	cost := "{X}{X}{G}{G}"
	pay := func() { b06AddMana(me, "G", "G", "C", "C", "C", "C") }

	pay()
	if err := neoB15Refused(g, "Spinning Wheel Kick", "Sorcery", cost, neoB15Kick, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: biter, Slot: 0}, {Kind: game.TargetCard, ID: a, Slot: 1}}}); err == nil {
		t.Error("X was 2 but only one creature was targeted")
	}
	pay()
	if err := neoB15Refused(g, "Spinning Wheel Kick", "Sorcery", cost, neoB15Kick, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}, {Kind: game.TargetCard, ID: biter, Slot: 1}}}); err == nil {
		t.Error("the damage source was a creature you don't control")
	}
	pay()
	neoB15Cast(t, g, "Spinning Wheel Kick", "Sorcery", cost, neoB15Kick, game.CastSpellParams{
		XValue:  2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: biter, Slot: 0}, {Kind: game.TargetCard, ID: a, Slot: 1}, {Kind: game.TargetCard, ID: b, Slot: 1}}})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if got := neoB15DamageOn(g, a); got != 3 {
		t.Errorf("the Wall took %d, want 3 (the Ogre's power)", got)
	}
	if findBattlefieldCardForTest(g, b) != nil {
		t.Error("the Elf survived 3 damage")
	}
	if got := neoB15DamageOn(g, biter); got != 0 {
		t.Errorf("the Ogre took %d, want none (it deals damage, nothing deals it any)", got)
	}
}

func TestNeoB15AccessDenied(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := castFromHandForTest(t, g, me, "Three Bear", "Creature — Bear", "{1}{G}{G}", "", game.CastSpellParams{})
	b06AddMana(me, "U", "U", "C", "C", "C")
	neoB15Cast(t, g, "Access Denied", "Instant", "{3}{U}{U}", neoB15Access, game.CastSpellParams{Targets: cardRefs(spell)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(spell) {
		t.Error("the countered creature entered the battlefield")
	}
	if n := tokensNamedControlledBy(g, "Thopter", me.ID); n != 3 {
		t.Errorf("%d Thopters, want 3 (the countered spell's mana value)", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Thopter" && (c.Power != 1 || c.Toughness != 1 || !c.HasCardType("artifact")) {
			t.Errorf("Thopter is %d/%d %q", c.Power, c.Toughness, c.TypeLine)
		}
	}
}

func TestNeoB15InvokeDespair(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Spare", "Creature — Bear", 2, 2)
	for i := 0; i < 3; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	b06AddMana(me, "B", "B", "B", "B", "C")
	if err := neoB15Refused(g, "Invoke Despair", "Sorcery", "{1}{B}{B}{B}{B}", neoB15Invoke,
		game.CastSpellParams{Targets: playerRefs(me.ID)}); err == nil {
		t.Error("cast at yourself")
	}
	life, hand := opp.Life, me.Hand.Size()
	b06AddMana(me, "B", "B", "B", "B", "C")
	neoB15Cast(t, g, "Invoke Despair", "Sorcery", "{1}{B}{B}{B}{B}", neoB15Invoke, game.CastSpellParams{Targets: playerRefs(opp.ID)})
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not sacrificed")
	}
	// They had a creature, but no enchantment and no planeswalker: two
	// failures, 2 life and a card each.
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want 4 less", life, opp.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two more (a draw for each failure; the spell itself left the hand)", hand, me.Hand.Size())
	}
}

func TestNeoB15InvokeDespairAllThreeSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	shrine := b12Permanent(g, opp.ID, "Shrine", "Enchantment")
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	life, hand := opp.Life, me.Hand.Size()
	b06AddMana(me, "B", "B", "B", "B", "C")
	neoB15Cast(t, g, "Invoke Despair", "Sorcery", "{1}{B}{B}{B}{B}", neoB15Invoke, game.CastSpellParams{Targets: playerRefs(opp.ID)})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{bear, shrine, walker} {
		answerSacrifice(t, g, opp.ID, id)
		passPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{bear, shrine, walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s is still on the battlefield", id)
		}
	}
	if opp.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d → %d, hand %d → %d: nothing should have been owed", life, opp.Life, hand, me.Hand.Size())
	}
}

func TestNeoB15SmokeSpiritsAid(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := b12Creature(g, opp.ID, "Elf", "Creature — Elf", 1, 1)
	b := b12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	cost := "{X}{R}"
	b06AddMana(me, "R", "C", "C", "C")
	if err := neoB15Refused(g, "Smoke Spirits' Aid", "Sorcery", cost, neoB15Smoke, game.CastSpellParams{
		XValue: 1, Targets: cardRefs(a, b)}); err == nil {
		t.Error("two targets for X=1")
	}
	b06AddMana(me, "R", "C", "C", "C")
	neoB15Cast(t, g, "Smoke Spirits' Aid", "Sorcery", cost, neoB15Smoke, game.CastSpellParams{XValue: 2, Targets: cardRefs(a, b)})
	passPriorityAroundTable(t, g)
	blessings := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Smoke Blessing" {
			continue
		}
		blessings++
		if !IsToken(c) || c.Controller != me.ID || !c.HasCardType("enchantment") || !c.HasSubtype("Aura") || len(c.Colors) != 1 || c.Colors[0] != "R" {
			t.Errorf("Smoke Blessing is %q colours %v controller %v", c.TypeLine, c.Colors, c.Controller)
		}
		if c.AttachedTo.ID != a && c.AttachedTo.ID != b {
			t.Errorf("Smoke Blessing attached to %v, want one of the targets", c.AttachedTo)
		}
	}
	if blessings != 2 {
		t.Fatalf("%d Smoke Blessings, want 2", blessings)
	}
	// The Elf dies: 1 damage to its controller (the opponent), a Treasure for me.
	life := opp.Life
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, a, 5); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d → %d, want 1 less", life, opp.Life)
	}
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
		}
	}
	if treasures != 1 {
		t.Errorf("%d Treasures for the caster, want 1", treasures)
	}
}

func TestNeoB15Storyweave(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	saga := b12Permanent(g, me.ID, "Test Saga", "Enchantment — Saga")

	b06AddMana(me, "G", "C", "C")
	neoB15Cast(t, g, "Storyweave", "Instant", "{2}{G}", neoB15Storyweav, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if got := battlefieldCountersOf(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the creature has %d +1/+1 counters, want 2", got)
	}

	b06AddMana(me, "G", "C", "C")
	if err := neoB15Refused(g, "Storyweave", "Instant", "{2}{G}", neoB15Storyweav, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(bear)}); err == nil {
		t.Error("lore counters aimed at a nonSaga creature")
	}
	b06AddMana(me, "G", "C", "C")
	neoB15Cast(t, g, "Storyweave", "Instant", "{2}{G}", neoB15Storyweav, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(saga)})
	passPriorityAroundTable(t, g)
	if got := battlefieldCountersOf(g, saga, game.CounterLore); got != 2 {
		t.Errorf("the Saga has %d lore counters, want 2", got)
	}

	enter := func(name, typeLine string, n int) []uuid.UUID {
		var ids []uuid.UUID
		g.WithWriteLock(func() {
			tmpl := game.Card{Name: name, TypeLine: typeLine, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}
			var err error
			ids, err = g.CreateTokensForEffect(me.ID, tmpl, n, game.TokenEntryOptions{})
			if err != nil {
				t.Fatalf("tokens: %v", err)
			}
		})
		return ids
	}
	// A creature that is not an enchantment does not use the grant up.
	plain := enter("Plain", "Token Creature — Spirit", 1)
	if got := battlefieldCountersOf(g, plain[0], game.CounterPlusOne); got != 0 {
		t.Errorf("a nonenchantment creature got %d counters", got)
	}
	// Two enchantment creatures entering together each get two counters.
	pair := enter("Pair", "Token Enchantment Creature — Spirit", 2)
	for _, id := range pair {
		if got := battlefieldCountersOf(g, id, game.CounterPlusOne); got != 2 {
			t.Errorf("an enchantment creature of the first entry has %d counters, want 2", got)
		}
	}
	// "The next time": a later entry gets nothing.
	later := enter("Later", "Token Enchantment Creature — Spirit", 1)
	if got := battlefieldCountersOf(g, later[0], game.CounterPlusOne); got != 0 {
		t.Errorf("a later enchantment creature got %d counters, want none", got)
	}
}

func TestNeoB15StoryweaveGrantEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	saga := b12Permanent(g, me.ID, "Test Saga", "Enchantment — Saga")
	b06AddMana(me, "G", "C", "C")
	neoB15Cast(t, g, "Storyweave", "Instant", "{2}{G}", neoB15Storyweav, game.CastSpellParams{Modes: []int{1}, Targets: cardRefs(saga)})
	passPriorityAroundTable(t, g)
	turn := g.Turn.Seq
	for i := 0; i < 40 && g.Turn.Seq == turn; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	var ids []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		ids, err = g.CreateTokensForEffect(me.ID, game.Card{Name: "Late", TypeLine: "Token Enchantment Creature — Spirit", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID}, 1, game.TokenEntryOptions{})
		if err != nil {
			t.Fatalf("tokens: %v", err)
		}
	})
	if got := battlefieldCountersOf(g, ids[0], game.CounterPlusOne); got != 0 {
		t.Errorf("an enchantment creature entering next turn got %d counters, want none", got)
	}
}

// neoB15DiscoverTable casts Discover the Impossible with `top` as the top of
// a five-card library and leaves the pick prompt open.
func neoB15DiscoverTable(t *testing.T, typeLine, cost string) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	top := discoverLibraryCard(me.ID, "The Pick", typeLine, cost)
	filler := func(n string) game.Card { return discoverLibraryCard(me.ID, n, "Creature — Bear", "{1}{G}") }
	me.Library.Cards = []game.Card{filler("Deep"), filler("Four"), filler("Three"), filler("Two"), top} // the last is the top
	b06AddMana(me, "U", "C", "C")
	neoB15Cast(t, g, "Discover the Impossible", "Instant", "{2}{U}", neoB15Discover, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pk := latestChooseCards(g, me.ID)
	if pk == nil || len(pk.ChooseCards) != 5 || pk.ChooseMin != 1 || pk.ChooseMax != 1 {
		t.Fatalf("pick: %+v, want exactly one of the five", pk)
	}
	if err := g.ResolveChooseCards(pk.ID, me.ID, []uuid.UUID{top.InstanceID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	return g, me, top.InstanceID
}

func TestNeoB15DiscoverTheImpossibleCastsACheapInstantFree(t *testing.T) {
	g, me, id := neoB15DiscoverTable(t, "Instant", "{1}")
	if !inExile(g, id) {
		t.Fatal("the picked card is not in exile")
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeOrHand, id)
	if err := castThroughResolutionCast(t, g, me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	if inHandOf(me, id) {
		t.Error("the cast card is also in hand")
	}
	passPriorityAroundTable(t, g)
	if inHandOf(me, id) || inExile(g, id) {
		t.Error("the cast card went to hand or stayed in exile")
	}
	if me.Library.Size() != 4 {
		t.Errorf("library %d, want the four others on the bottom", me.Library.Size())
	}
}

func TestNeoB15DiscoverTheImpossibleDecliningPutsItIntoHand(t *testing.T) {
	g, me, id := neoB15DiscoverTable(t, "Instant", "{2}")
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeOrHand, id)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !inHandOf(me, id) {
		t.Error("the declined card is not in hand")
	}
}

func TestNeoB15DiscoverTheImpossibleBiggerInstantGoesToHand(t *testing.T) {
	g, me, id := neoB15DiscoverTable(t, "Instant", "{2}{U}")
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Fatal("a mana value 3 instant was offered for free")
	}
	if !inHandOf(me, id) {
		t.Error("the card is not in hand")
	}
}

func TestNeoB15DiscoverTheImpossibleSorceryAndLandGoToHand(t *testing.T) {
	for _, tc := range []struct{ typeLine, cost string }{{"Sorcery", "{U}"}, {"Basic Land — Island", ""}, {"Creature — Bear", "{U}"}} {
		g, me, id := neoB15DiscoverTable(t, tc.typeLine, tc.cost)
		if resolutionCastOfferFor(g, me.ID) != nil {
			t.Errorf("a %s was offered for free", tc.typeLine)
		}
		if !inHandOf(me, id) {
			t.Errorf("a %s is not in hand", tc.typeLine)
		}
	}
}

func TestNeoB15CollisionOfRealms(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	// Their only creature is a token: it is shuffled away (and ceases to
	// exist), but the token alone doesn't earn a creature off the top.
	g.WithWriteLock(func() {
		if _, err := g.CreateTokensForEffect(third.ID, game.Card{Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Owner: third.ID, Controller: third.ID}, 1, game.TokenEntryOptions{}); err != nil {
			t.Fatalf("token: %v", err)
		}
	})
	myHit := discoverLibraryCard(me.ID, "My Hit", "Creature — Elf", "{G}")
	theirHit := discoverLibraryCard(opp.ID, "Their Hit", "Creature — Elf", "{G}")
	thirdTop := discoverLibraryCard(third.ID, "Third Hit", "Creature — Elf", "{G}")
	// Each library is a lone creature card: whatever the shuffle does, the
	// first creature revealed is among the two cards there are.
	me.Library.Cards = []game.Card{myHit}
	opp.Library.Cards = []game.Card{theirHit}
	third.Library.Cards = []game.Card{thirdTop}

	b06AddMana(me, "R", "C", "C", "C", "C", "C", "C")
	neoB15Cast(t, g, "Collision of Realms", "Sorcery", "{6}{R}", neoB15Collision, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()

	owned := func(p *game.Player) []game.Card {
		var out []game.Card
		for _, c := range g.Battlefield.Cards {
			if c.IsCreature() && c.Controller == p.ID {
				out = append(out, c)
			}
		}
		return out
	}
	// A player who shuffled in a nontoken creature reveals down to a creature
	// card and puts it onto the battlefield: with only creature cards in the
	// library, that is the first card of the shuffled pair, whichever it is.
	for _, p := range []*game.Player{me, opp} {
		got := owned(p)
		if len(got) != 1 {
			t.Errorf("%d creatures on the battlefield for the player who shuffled one in, want exactly 1", len(got))
			continue
		}
		if p.Library.Size() != 1 {
			t.Errorf("library %d, want the other card of the pair left on the bottom", p.Library.Size())
		}
		if got[0].Owner != p.ID {
			t.Error("the creature entered under someone else's control")
		}
	}
	if got := owned(third); len(got) != 0 {
		t.Errorf("the player who shuffled in only a token has %d creatures, want none", len(got))
	}
	if third.Library.Size() != 1 {
		t.Errorf("their library %d, want it untouched", third.Library.Size())
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" {
			t.Error("the token is still on the battlefield")
		}
	}
	_ = []uuid.UUID{mine, theirs, myHit.InstanceID, theirHit.InstanceID, thirdTop.InstanceID}
}
