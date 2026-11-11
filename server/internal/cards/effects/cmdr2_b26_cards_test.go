package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	c26PucaOracle      = "9b832f87-0bd5-47bd-ab9c-718bd9a6c441"
	c26ToolsOracle     = "c89cd328-b4cc-4760-8b6c-59e82eed8dac"
	c26GoatOracle      = "63193ad4-bbd2-4cc3-98e2-9f6b985e8ae7"
	c26GnarlbarkOracle = "fcd2142e-ba84-4693-9b29-ecc43feba2b8"
	c26TillerOracle    = "ffcffad7-6c33-4265-82c9-04045fdf371b"
	c26PillagersOracle = "54b44499-398b-47ad-b790-bd5b5994d1a6"
	c26ScorpionOracle  = "aff223d4-6338-460d-a4d5-d45f522864dc"
	c26AuntieOracle    = "e3cd53c1-86bd-4177-adb0-a1fa693b44c6"
	c26ShadowOracle    = "92ac8aa4-e23c-4ed9-8317-c249a17ce3d2"
	c26HapatraOracle   = "a8ddea1c-8d80-49c1-a5b4-630d5e51d66e"
	c26ReaperOracle    = "39b67a4d-6a87-41f0-a86f-b66671ccc20d"
)

func TestCmdr2B26EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		c26PucaOracle: "Puca's Covenant", c26ToolsOracle: "Wickersmith's Tools", c26GoatOracle: "Oft-Nabbed Goat",
		c26GnarlbarkOracle: "Sinister Gnarlbark", c26TillerOracle: "Dread Tiller", c26PillagersOracle: "Village Pillagers",
		c26ScorpionOracle: "The Scorpion God", c26AuntieOracle: "Auntie Ool, Cursewretch", c26ShadowOracle: "Eventide's Shadow",
		c26HapatraOracle: "Hapatra, Vizier of Poisons", c26ReaperOracle: "The Reaper, King No More",
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

func c26Bolt(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(id))
	passPriorityAroundTable(t, g)
}

func TestCmdr2B26SinisterGnarlbarkDrawsAndBlightsAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bark := b12Push(g, me.ID, "Sinister Gnarlbark", "Creature — Treefolk Warlock", c26GnarlbarkOracle, 0, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if blightPrompt(g, me.ID) == nil {
		t.Fatal("no blight prompt")
	}
	if p := blightPrompt(g, me.ID); hasID(p.ChooseCards, theirs) {
		t.Error("an opponent's creature is not a creature you control")
	}
	answerChooseCards(t, g, me.ID, bear)
	monSettle(t, g)
	if got := countersOn(g, bear, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the chosen creature = %d, want 1", got)
	}
	if got := countersOn(g, bark, game.CounterMinusOne); got != 0 {
		t.Errorf("-1/-1 counters on the Gnarlbark = %d, want 0", got)
	}
}

func TestCmdr2B26DreadTillerPutsACounterThenPutsALandWhenACounteredCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Elf", 1, 1)
	other := b12Creature(g, opp.ID, "Other", "Creature — Elf", 3, 3)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: me.ID, Controller: me.ID})
	c25CastCreature(t, g, "Dread Tiller", "Artifact Creature — Scarecrow", c26TillerOracle, 2, 4)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	// The counter killed the 1/1, a creature with a -1/-1 counter dying.
	monSettle(t, g)
	if p := latestChooseCardsFor(g, me.ID); p == nil {
		t.Fatal("no land prompt after a creature with a -1/-1 counter died")
	}
	answerChooseCards(t, g, me.ID, land)
	monSettle(t, g)
	c := battlefieldCardFor(g, land)
	if c == nil || !c.Tapped {
		t.Fatalf("the land is not on the battlefield tapped: %+v", c)
	}
	// A creature with no counter dying asks nothing.
	c26Bolt(t, g, other)
	if p := latestChooseCardsFor(g, me.ID); p != nil {
		t.Error("a creature with no -1/-1 counter died but a prompt opened")
	}
}

func TestCmdr2B26VillagePillagersDamagesOpponentsAndMakesATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Elf", 2, 2)
	small := b12Creature(g, opp.ID, "Small", "Creature — Elf", 1, 1)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	c25CastCreature(t, g, "Village Pillagers", "Creature — Goblin Warrior", c26PillagersOracle, 5, 5)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if countersOn(g, mine, game.CounterMinusOne) != 0 || damageMarkedOn(g, mine) != 0 {
		t.Error("my own creature is untouched")
	}
	if countersOn(g, big, game.CounterMinusOne) != 1 {
		t.Errorf("the Ogre has %d -1/-1 counters, want 1 (wither)", countersOn(g, big, game.CounterMinusOne))
	}
	if g.Battlefield.Contains(small) {
		t.Fatal("the 1/1 with a -1/-1 counter survived")
	}
	// The 1/1 died with a counter: a tapped Treasure.
	n, _ := countTokensNamed(g, me.ID, "Treasure")
	if n != 1 {
		t.Fatalf("Treasures = %d, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && !c.Tapped {
			t.Error("the Treasure is not tapped")
		}
	}
	// My creature dying with a counter makes none.
	addCountersForTest(g, mine, game.CounterPlusOne, 1)
	c26Bolt(t, g, mine)
	if n2, _ := countTokensNamed(g, me.ID, "Treasure"); n2 != 1 {
		t.Errorf("Treasures = %d, want still 1: it was my creature", n2)
	}
}

func TestCmdr2B26ScorpionGodDrawsPutsCountersAndComesBack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	god := b12Push(g, me.ID, "The Scorpion God", "Legendary Creature — God", c26ScorpionOracle, 6, 5)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Elf", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, god, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: god}}}); err == nil {
		t.Error("the target must be another creature")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, god, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the 1/1 survived a -1/-1 counter")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	// The God dies and returns at the next end step.
	addCountersForTest(g, god, game.CounterMinusOne, 1)
	c26Bolt(t, g, god)
	c26Bolt(t, g, god)
	monSettle(t, g)
	if g.Battlefield.Contains(god) {
		t.Fatal("the God survived")
	}
	advanceToEndStepOf(t, g, 0)
	monSettle(t, g)
	if !me.Hand.Contains(god) {
		t.Error("The Scorpion God did not return to its owner's hand at the next end step")
	}
}

func TestCmdr2B26WickersmithsToolsChargeAndScarecrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tools := b12Push(g, me.ID, "Wickersmith's Tools", "Artifact", c26ToolsOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	putCounters(t, g, theirs, game.CounterMinusOne, 2)
	putCounters(t, g, theirs, game.CounterMinusOne, 1)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if got := countersOn(g, tools, game.CounterCharge); got != 2 {
		t.Fatalf("charge counters = %d, want 2 (one per placement event)", got)
	}
	putCounters(t, g, theirs, game.CounterPlusOne, 1)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, tools, game.CounterCharge); got != 2 {
		t.Errorf("a +1/+1 counter charged it: %d", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, tools, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tools) {
		t.Error("the artifact is not sacrificed")
	}
	n, _ := countTokensNamed(g, me.ID, "Scarecrow")
	if n != 2 {
		t.Fatalf("Scarecrows = %d, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Scarecrow" && (!c.Tapped || c.Power != 2 || c.Toughness != 2) {
			t.Errorf("Scarecrow %+v: want a tapped 2/2", c)
		}
	}
}

func TestCmdr2B26HapatraCountersOnCombatDamageAndMakesASnake(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hap := b12Push(g, me.ID, "Hapatra, Vizier of Poisons", "Legendary Creature — Human Cleric", c26HapatraOracle, 2, 2)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	attackWith(t, g, opp.ID, hap)
	passPriorityAroundTable(t, g)
	if p := latestPickTarget(g, me.ID); p != nil {
		pickCard(t, g, me.ID, victim)
		passPriorityAroundTable(t, g)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if got := countersOn(g, victim, game.CounterMinusOne); got != 1 {
		t.Fatalf("victim counters = %d, want 1", got)
	}
	n, _ := countTokensNamed(g, me.ID, "Snake")
	if n != 1 {
		t.Errorf("Snakes = %d, want 1 (I put the counter)", n)
	}
	// An opponent putting counters on their own creature makes me no Snake.
	g.WithWriteLock(func() { _, _ = g.AddCounterByMustSettleNowForEffect(opp.ID, victim, game.CounterMinusOne, 1) })
	passPriorityAroundTable(t, g)
	if n2, _ := countTokensNamed(g, me.ID, "Snake"); n2 != 1 {
		t.Errorf("Snakes = %d, want still 1", n2)
	}
}

func TestCmdr2B26AuntieOolDrawsForYourCreatureAndDrainsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Auntie Ool, Cursewretch", "Legendary Creature — Goblin Warlock", c26AuntieOracle, 4, 4)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Elf", 3, 3)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Elf", 3, 3)
	hand, life := me.Hand.Size(), opp.Life
	putCounters(t, g, mine, game.CounterMinusOne, 2)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card for my creature", hand, me.Hand.Size())
	}
	putCounters(t, g, theirs, game.CounterMinusOne, 2)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if opp.Life != life-1 {
		t.Errorf("their life %d → %d, want 1 lost", life, opp.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want no more cards for their creature", me.Hand.Size())
	}
}

func TestCmdr2B26AuntieOolWardBlightTwo(t *testing.T) {
	// Paying: the caster blights a creature of theirs and the spell resolves.
	g, _, opp, auntie := wardTable(t, "Auntie Ool, Cursewretch", "Legendary Creature — Goblin Warlock", c26AuntieOracle)
	mine := b12Creature(g, opp.ID, "Caster's Bear", "Creature — Bear", 4, 4)
	blade := castAtWardedCreature(t, g, opp, auntie)
	passPriorityAroundTable(t, g)
	p := blightPrompt(g, opp.ID)
	if p == nil {
		t.Fatal("the caster is not asked to blight")
	}
	if !hasID(p.ChooseCards, mine) {
		t.Error("the caster's creature is not offered")
	}
	answerChooseCards(t, g, opp.ID, mine)
	monSettle(t, g)
	if got := countersOn(g, mine, game.CounterMinusOne); got != 2 {
		t.Errorf("the blighted creature has %d counters, want 2", got)
	}
	if g.Battlefield.Contains(auntie) {
		t.Error("the ward was paid, so the removal spell resolves")
	}
	_ = blade

	// Declining counters the spell.
	g2, _, opp2, auntie2 := wardTable(t, "Auntie Ool, Cursewretch", "Legendary Creature — Goblin Warlock", c26AuntieOracle)
	b12Creature(g2, opp2.ID, "Caster's Bear", "Creature — Bear", 4, 4)
	blade2 := castAtWardedCreature(t, g2, opp2, auntie2)
	passPriorityAroundTable(t, g2)
	if blightPrompt(g2, opp2.ID) == nil {
		t.Fatal("no blight prompt")
	}
	answerChooseCards(t, g2, opp2.ID)
	monSettle(t, g2)
	if !g2.Battlefield.Contains(auntie2) || !opp2.Graveyard.Contains(blade2) {
		t.Error("declining must counter the spell")
	}

	// No creature to blight: it can't be paid, and the spell is countered.
	g3, _, opp3, auntie3 := wardTable(t, "Auntie Ool, Cursewretch", "Legendary Creature — Goblin Warlock", c26AuntieOracle)
	blade3 := castAtWardedCreature(t, g3, opp3, auntie3)
	passPriorityAroundTable(t, g3)
	monSettle(t, g3)
	if !g3.Battlefield.Contains(auntie3) || !opp3.Graveyard.Contains(blade3) {
		t.Error("a caster with no creature to blight can't pay: the spell is countered")
	}
}

func TestCmdr2B26OftNabbedGoatIsStolenByOpponentsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goat := b12Push(g, me.ID, "Oft-Nabbed Goat", "Creature — Goat", c26GoatOracle, 0, 5)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, goat, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("its own controller may not activate it")
	}
	if err := g.ActivateCatalogAbility(opp.ID, goat, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("an opponent may only activate it as a sorcery: not in my turn")
	}
	advanceToMainOf(t, g, 1)
	hand := opp.Hand.Size()
	if err := g.ActivateCatalogAbility(opp.ID, goat, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("an opponent activates: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := battlefieldCardFor(g, goat)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("the opponent did not gain control: %+v", c)
	}
	if opp.Hand.Size() != hand+1 {
		t.Errorf("the activator hand %d → %d, want one card", hand, opp.Hand.Size())
	}
	if got := countersOn(g, goat, game.CounterMinusOne); got != 1 {
		t.Errorf("counters on the Goat = %d, want 1", got)
	}
}

func TestCmdr2B26OftNabbedGoatDiesForItsOwnerAndDrainsEveryoneElse(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	goat := b12Push(g, me.ID, "Oft-Nabbed Goat", "Creature — Goat", c26GoatOracle, 0, 5)
	// The opponent holds it, with two -1/-1 counters on it.
	stealForTest(t, g, goat, opp.ID)
	addCountersForTest(g, goat, game.CounterMinusOne, 2)
	hand, theirs, mine := me.Hand.Size(), opp.Life, me.Life
	third := g.Seats[2].Life
	c26Bolt(t, g, goat)
	monSettle(t, g)
	if g.Battlefield.Contains(goat) {
		t.Fatal("the Goat survived")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("owner hand %d → %d, want two cards drawn", hand, me.Hand.Size())
	}
	if opp.Life != theirs-2 || g.Seats[2].Life != third-2 {
		t.Errorf("each other player should lose 2: opp %d→%d, third %d→%d", theirs, opp.Life, third, g.Seats[2].Life)
	}
	if me.Life != mine {
		t.Errorf("the owner lost life: %d → %d", mine, me.Life)
	}
}

func TestCmdr2B26PucasCovenantReturnsACheapPermanentOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Puca's Covenant", "Enchantment", c26PucaOracle, 0, 0)
	cheap, dear, third := uuid.New(), uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{3}", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: third, Name: "Spell", TypeLine: "Instant", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	a := b12Creature(g, me.ID, "A", "Creature — Elf", 4, 4)
	b := b12Creature(g, me.ID, "B", "Creature — Elf", 4, 4)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Elf", 2, 2)
	addCountersForTest(g, a, game.CounterMinusOne, 2)
	addCountersForTest(g, b, game.CounterMinusOne, 1)

	c26Bolt(t, g, plain)
	monSettle(t, g)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a creature with no counter on it died but the Covenant triggered")
	}

	c26Bolt(t, g, a)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if hasID(p.PickTargetCards, dear) || hasID(p.PickTargetCards, third) || hasID(p.PickTargetCards, a) {
		t.Errorf("offered %v: want only the permanent card of mana value 2 or less", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, cheap)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(cheap) {
		t.Fatal("the card did not return to hand")
	}
	// Only once each turn: the next creature's death does nothing.
	c26Bolt(t, g, b)
	monSettle(t, g)
	if latestPickTarget(g, me.ID) != nil || latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a second trigger this turn")
	}
}

func TestCmdr2B26TheReaperCountersTwoAndStealsADyingCounteredCreatureOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, opp.ID, "A", "Creature — Elf", 1, 1)
	b := b12Creature(g, opp.ID, "B", "Creature — Ogre", 4, 4)
	c := b12Creature(g, opp.ID, "C", "Creature — Ogre", 4, 4)
	c25CastCreature(t, g, "The Reaper, King No More", "Legendary Artifact Creature — Scarecrow", c26ReaperOracle, 3, 3)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	// A dies to its counter, an opponent's creature with a -1/-1 counter.
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	monSettle(t, g)
	if got := countersOn(g, b, game.CounterMinusOne); got != 1 {
		t.Errorf("the second target has %d counters, want 1", got)
	}
	if cc := battlefieldCardFor(g, a); cc == nil || cc.Controller != me.ID {
		t.Fatalf("the dead creature did not return under my control: %+v", cc)
	}
	// Once each turn: the next one stays dead.
	addCountersForTest(g, c, game.CounterMinusOne, 1)
	c26Bolt(t, g, c)
	monSettle(t, g)
	if latestTriggerPrompt(g, me.ID) != nil {
		t.Error("a second return was offered this turn")
	}
	if battlefieldCardFor(g, c) != nil {
		t.Error("the second creature returned")
	}
}

func TestCmdr2B26EventidesShadowRemovesCountersAndDrawsAndLosesThatMany(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "A", "Creature — Elf", 4, 4)
	b := b12Creature(g, opp.ID, "B", "Creature — Elf", 4, 4)
	addCountersForTest(g, a, game.CounterPlusOne, 3)
	addCountersForTest(g, b, game.CounterMinusOne, 2)
	hand, life := me.Hand.Size(), me.Life
	castCatalogSpell(t, g, "Eventide's Shadow", "Sorcery", c26ShadowOracle, nil)
	passPriorityAroundTable(t, g)
	// The questions come one per kind per permanent, in battlefield order.
	answerOptionPick(t, g, me.ID, 2)
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if countersOn(g, a, game.CounterPlusOne) != 1 || countersOn(g, b, game.CounterMinusOne) != 1 {
		t.Errorf("counters left: A %d, B %d; want 1 and 1", countersOn(g, a, game.CounterPlusOne), countersOn(g, b, game.CounterMinusOne))
	}
	if got := me.Hand.Size(); got != hand+3 {
		t.Errorf("hand %d → %d, want three cards drawn", hand, got)
	}
	if me.Life != life-3 {
		t.Errorf("life %d → %d, want 3 lost", life, me.Life)
	}
}
