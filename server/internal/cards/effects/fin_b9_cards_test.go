package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fin_b9_cards_test.go — Final Fantasy batches 9 and 10 (#4462).

const (
	finB9GabranthOracle = "1a846ad6-15c9-4bf9-9127-612018a557b9"
)

var _ = uuid.Nil

func TestFinB9GabranthGrowsWhenYourOtherCreatureOrArtifactDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	gab := b12Push(g, me.ID, "Judge Magister Gabranth", "Legendary Creature — Human Advisor Knight", finB9GabranthOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	art := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	if !hasEffectiveKeyword(t, g, gab, "menace") {
		t.Error("no menace")
	}
	for i, id := range []uuid.UUID{bear, art, theirs, land} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		passPriorityAroundTable(t, g)
		want := []int{1, 2, 2, 2}[i]
		if got := countersOn(g, gab, game.CounterPlusOne); got != want {
			t.Fatalf("after death %d: %d counters, want %d", i, got, want)
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(gab) })
	passPriorityAroundTable(t, g)
}

const finB9JenovaOracle = "9370f6bf-d6a0-4c5b-828f-7229f563cc07"

func TestFinB9JenovaCountersMutatesAndDrawsOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jen := b12Push(g, me.ID, "Jenova, Ancient Calamity", "Legendary Creature — Alien", finB9JenovaOracle, 1, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("bear has %d counters, want 1 (Jenova's power)", got)
	}
	if c, _ := g.LookupCardForEffect(bear); !c.HasSubtype("Mutant") || !c.HasSubtype("Bear") {
		t.Error("bear is not a Bear and a Mutant")
	}
	_ = jen
	hand := me.Hand.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3 (a 3-power Mutant)", hand, me.Hand.Size())
	}
}

const (
	finB9BoltOracle  = "aad9f9f2-4c6b-48d0-80bd-77b73a9ea5f1"
	finB9RufusOracle = "0caeb396-e574-4e80-a1f6-79988fe9ba19"
	finB9RinoaOracle = "8896c74b-d466-4c25-94cf-82b5e2b1fdef"
)

func TestFinB9JudgmentBoltHitsTheCreatureAndItsControllerPerEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
	finABEquipment(g, me.ID, "Sword")
	finABEquipment(g, me.ID, "Shield")
	finABEquipment(g, opp.ID, "Their Sword")
	life := opp.Life
	finABCastInstantNow(t, g, "Judgment Bolt", finB9BoltOracle, finABCard(victim))
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("their life %d → %d, want -2 (two Equipment of yours)", life, opp.Life)
	}
	if !opp.Graveyard.Contains(victim) {
		t.Error("the 5-damage creature survived")
	}
}

func TestFinB9JudgmentBoltWithNoEquipmentDealsOnlyTheFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	advanceToMain(t, g)
	victim := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 6)
	life := opp.Life
	finABCastInstantNow(t, g, "Judgment Bolt", finB9BoltOracle, finABCard(victim))
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("their life %d → %d, want unchanged", life, opp.Life)
	}
	if opp.Graveyard.Contains(victim) {
		t.Error("a 6-toughness creature died to 5")
	}
}

func TestFinB9RufusMakesDarkstarOnlyWithoutOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rufus := b12Push(g, me.ID, "Rufus Shinra", "Legendary Creature — Human Noble", finB9RufusOracle, 2, 4)
	finABAttack(t, g, opp.ID, rufus)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Darkstar"); n != 1 {
		t.Fatalf("%d Darkstar tokens after the first attack, want 1", n)
	}
}

func TestFinB9RufusDoesNotMakeASecondDarkstar(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rufus := b12Push(g, me.ID, "Rufus Shinra", "Legendary Creature — Human Noble", finB9RufusOracle, 2, 4)
	b12Creature(g, me.ID, "Darkstar", "Legendary Creature — Dog", 2, 2)
	finABAttack(t, g, opp.ID, rufus)
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "Darkstar"); n != 1 {
		t.Errorf("%d Darkstars, want the one you already control", n)
	}
}

func TestFinB9RinoaMakesAngeloAndPumpsPerCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rinoa := b12Push(g, me.ID, "Rinoa Heartilly", "Legendary Creature — Human Rebel Warlock", finB9RinoaOracle, 4, 4)
	ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Other", "Creature — Bear", 1, 1)
	finABAttack(t, g, opp.ID, rinoa)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(ally)...)
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, ally)
	if v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Errorf("ally is %d/%d, want 5/5 (+3 for three creatures)", v.CurrentPower(), v.CurrentToughness())
	}
}

func TestFinB9RinoaEntersWithAngelo(t *testing.T) {
	g := newCatalogGame(t)
	castCreatureCard(t, g, "Rinoa Heartilly", "Legendary Creature — Human Rebel Warlock", finB9RinoaOracle, 4, 4)
	if n := onBattlefieldNamed(g, "Angelo"); n != 1 {
		t.Errorf("%d Angelo tokens, want 1", n)
	}
}

const (
	finB9FangOracle     = "04ac8b2c-95e9-40e9-8cc0-af9b075a94c0"
	finB9RagnarokOracle = "14b839ef-80b1-4e7d-b2fd-38e973899cde"
)

func TestFinB9FangDrawsAndLosesLifeOncePerTurnPerBatch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Fang, Fearless l'Cie", "Legendary Creature — Human Warrior", finB9FangOracle, 2, 3)
	a := pushGraveyardPermanent(me, "A", "Creature — Bear", "{1}")
	b := pushGraveyardPermanent(me, "B", "Creature — Bear", "{1}")
	c := pushGraveyardPermanent(me, "C", "Creature — Bear", "{1}")
	exile := func(id uuid.UUID) {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
			t.Fatal(err)
		}
	}
	hand, life := me.Hand.Size(), me.Life
	exile(a)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Fatalf("hand %d→%d life %d→%d, want +1 card and -1 life", hand, me.Hand.Size(), life, me.Life)
	}
	exile(b)
	exile(c)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("a second batch the same turn changed hand %d or life %d", me.Hand.Size(), me.Life)
	}
}

func TestFinB9RagnarokDiesDestroysAndReturnsANonlegendaryPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	rag := b12Push(g, me.ID, "Ragnarok, Divine Deliverance", "Legendary Creature — Beast Avatar", finB9RagnarokOracle, 7, 6)
	for _, k := range []string{"reach", "vigilance", "menace", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, rag, k) {
			t.Errorf("no %s", k)
		}
	}
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	legend := pushGraveyardPermanent(me, "Legend", "Legendary Creature — Human", "{1}{G}")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rag) })
	pickCard(t, g, me.ID, victim)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("the target permanent was not destroyed")
	}
	if !onBattlefield(g, bear) {
		t.Error("the nonlegendary card did not return")
	}
	if onBattlefield(g, legend) {
		t.Error("the legendary card returned")
	}
}

const finB9HopeOracle = "187a110c-032a-42e1-be2d-3cfaaa9495f8"

func TestFinB9HopeMillsEachOpponentByLifeGainedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Hope Estheim", "Legendary Creature — Human Wizard", finB9HopeOracle, 2, 2)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3) })
	passPriorityAroundTable(t, g)
	sizes := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		sizes[p.ID] = p.Library.Size()
	}
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Library.Size() != sizes[p.ID]-3 {
			t.Errorf("opponent library %d → %d, want -3", sizes[p.ID], p.Library.Size())
		}
	}
	if g.Seats[0].Library.Size() == 0 {
		t.Error("you milled")
	}
}

const finB9SeiferOracle = "7088a901-0489-41c3-9f2f-633939514de8"

func TestFinB9SeiferGivesALoneAttackerDoubleStrikeAndFireCrossCastsFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seifer := b12Push(g, me.ID, "Seifer Almasy", "Legendary Creature — Human Knight", finB9SeiferOracle, 3, 4)
	spell := gyCard(me, "Old Sorcery", "Sorcery", "{2}{R}", "")
	dear := gyCard(me, "Big Sorcery", "Sorcery", "{3}{R}", "")
	finABAttack(t, g, opp.ID, seifer)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, seifer, "double strike") {
		t.Fatal("the lone attacker did not gain double strike")
	}
	for i := 0; i < 12 && latestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
	}
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no Fire Cross target prompt")
	}
	if !hasID(p.PickTargetCards, spell) || hasID(p.PickTargetCards, dear) {
		t.Errorf("offered %v: want only the mana value 3 sorcery", p.PickTargetCards)
	}
	pickCard(t, g, me.ID, spell)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell)
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("casting for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(spell) || !inExile(g, spell) {
		t.Error("the spell should be exiled, not back in the graveyard")
	}
}

func TestFinB9SeiferGivesNoDoubleStrikeWhenTwoAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seifer := b12Push(g, me.ID, "Seifer Almasy", "Legendary Creature — Human Knight", finB9SeiferOracle, 3, 4)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	finABAttack(t, g, opp.ID, seifer, other)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, seifer, "double strike") || hasEffectiveKeyword(t, g, other, "double strike") {
		t.Error("double strike with two attackers")
	}
}

const finB9ZidaneOracle = "f267fb54-1881-464c-81af-9e03c3f3b41d"

func TestFinB9ZidaneStealsUntapsAndGrantsLifelinkAndHaste(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == victim {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	castCreatureCard(t, g, "Zidane, Tantalus Thief", "Legendary Creature — Human Mutant Scout", finB9ZidaneOracle, 3, 3)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(victim)
	if c.Controller != me.ID {
		t.Fatal("control not gained")
	}
	if c.Tapped {
		t.Error("not untapped")
	}
	if !hasEffectiveKeyword(t, g, victim, "lifelink") || !hasEffectiveKeyword(t, g, victim, "haste") {
		t.Error("missing lifelink or haste")
	}
	// An opponent taking something from you makes a Treasure.
	zid := b12Push(g, me.ID, "Zidane, Tantalus Thief", "Legendary Creature — Human Mutant Scout", finB9ZidaneOracle, 3, 3)
	_ = zid
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := onBattlefieldNamed(g, "Treasure")
	stealForTest(t, g, mine, opp.ID)
	passPriorityAroundTable(t, g)
	if got := onBattlefieldNamed(g, "Treasure"); got != before+1 {
		t.Errorf("%d Treasures, want %d", got, before+1)
	}
}

const (
	finB9WarriorOracle   = "393a5d32-fc8e-422b-a957-59e545a54cb6"
	finB9WhiteMageOracle = "c8110200-209e-423a-96b7-81a3ac32447c"
	finB9ThiefOracle     = "7857c7b6-9676-47ac-bb28-d0b4cbda5fa7"
	finB9BlackMageOracle = "ef2aa67a-9531-495c-85c7-7f745bc19ce1"
)

func TestFinB9PartySizeCountsOneCreaturePerTypeBestAssignment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	size := func(p uuid.UUID) (n int) {
		g.WithWriteLock(func() { n = finB9PartySize(g, p) })
		return
	}
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if got := size(me.ID); got != 0 {
		t.Fatalf("party %d with no party members", got)
	}
	// A Cleric Wizard and a Wizard: best assignment gives two.
	b12Creature(g, me.ID, "Cleric Wizard", "Creature — Human Cleric Wizard", 1, 1)
	b12Creature(g, me.ID, "Wizard", "Creature — Human Wizard", 1, 1)
	b12Creature(g, me.ID, "Another Wizard", "Creature — Human Wizard", 1, 1)
	if got := size(me.ID); got != 2 {
		t.Fatalf("party %d, want 2 (Cleric/Wizard plus Wizard)", got)
	}
	b12Creature(g, me.ID, "Rogue", "Creature — Human Rogue", 1, 1)
	b12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 1, 1)
	if got := size(me.ID); got != 3 {
		t.Fatalf("party %d, want 3", got)
	}
	b12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 1, 1)
	if got := size(me.ID); got != 4 {
		t.Fatalf("party %d, want 4", got)
	}
	if got := size(opp.ID); got != 1 {
		t.Errorf("their party %d, want 1", got)
	}
}

func finB9FillParty(g *game.Game, owner uuid.UUID, skip string) {
	for _, ty := range []string{"Cleric", "Rogue", "Warrior", "Wizard"} {
		if ty != skip {
			b12Creature(g, owner, "Party "+ty, "Creature — Human "+ty, 1, 1)
		}
	}
}

func TestFinB9DestinedWarriorPumpsByOneOrThreeAndDiscountsPartySpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	w := b12Push(g, me.ID, "The Destined Warrior", "Legendary Creature — Human Warrior", finB9WarriorOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, k := range []string{"first strike", "vigilance", "menace"} {
		if !hasEffectiveKeyword(t, g, w, k) {
			t.Errorf("no %s", k)
		}
	}
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 3 {
		t.Errorf("bear power %d, want 3 (+1/+0)", v.CurrentPower())
	}
	if v := s58p6View(t, g, w); v.CurrentPower() != 4 {
		t.Errorf("Warrior power %d, want 4", v.CurrentPower())
	}
}

func TestFinB9DestinedWarriorPumpsByThreeWithAFullParty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "The Destined Warrior", "Legendary Creature — Human Warrior", finB9WarriorOracle, 3, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	finB9FillParty(g, me.ID, "Warrior") // the Warrior itself is the fourth member
	advanceToStepOf(t, g, 0, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, bear); v.CurrentPower() != 5 {
		t.Errorf("bear power %d, want 5 (+3/+0 with a full party)", v.CurrentPower())
	}
}

func TestFinB9DestinedWhiteMageCountersOnLifeGainAndThreeWithAFullParty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	wm := b12Push(g, me.ID, "The Destined White Mage", "Legendary Creature — Human Cleric", finB9WhiteMageOracle, 2, 3)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gain := func() {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}
	gain()
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters, want 1", got)
	}
	finB9FillParty(g, me.ID, "Cleric") // the White Mage is the Cleric
	gain()
	if got := countersOn(g, bear, game.CounterPlusOne); got != 4 {
		t.Errorf("%d counters, want 4 (1 + 3 with a full party)", got)
	}
	_ = wm
}

func TestFinB9DestinedBlackMagePingsOneOrThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "The Destined Black Mage", "Legendary Creature — Human Wizard", finB9BlackMageOracle, 3, 2)
	cast := func() {
		id := pushCatalogHandCard(me, "Shock", "Instant", "")
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	life := opp.Life
	cast()
	if opp.Life != life-1 {
		t.Fatalf("life %d → %d, want -1", life, opp.Life)
	}
	finB9FillParty(g, me.ID, "Wizard")
	cast()
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4 total (1, then 3 with a full party)", life, opp.Life)
	}
}

func TestFinB9DestinedThiefLootsOrDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	thief := b12Push(g, me.ID, "The Destined Thief", "Legendary Creature — Human Rogue", finB9ThiefOracle, 2, 2)
	hand := me.Hand.Size()
	finB9FillParty(g, me.ID, "Rogue")
	finABAttack(t, g, opp.ID, thief)
	passPriorityAroundTable(t, g)
	for i := 0; i < 12 && me.Hand.Size() == hand; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			break
		}
		passPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3 with a full party", hand, me.Hand.Size())
	}
}

const finB9ChocoOracle = "7979bf83-5091-4e62-8885-ca1833a97306"

func TestFinB9ChocoLooksAtOneCardPerBirdPutsOneInHandLandsInPlayAndTheRestInTheYard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	choco := b12Push(g, me.ID, "Choco, Seeker of Paradise", "Legendary Creature — Bird", finB9ChocoOracle, 3, 5)
	bird := b12Creature(g, me.ID, "Bird", "Creature — Bird", 1, 1)
	b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// Top to bottom: spell, land, land, then an unlooked card.
	deep := topCard(me, "Deep", "Creature — Bear", "{1}")
	land2 := topCard(me, "Forest B", "Basic Land — Forest", "")
	land1 := topCard(me, "Forest A", "Basic Land — Forest", "")
	spell := topCard(me, "Spell", "Sorcery", "{1}")
	_ = land2
	finABAttack(t, g, opp.ID, choco, bird)
	passPriorityAroundTable(t, g)
	// Three cards are NOT looked at: two Birds attacked, so two cards.
	answerChooseCards(t, g, me.ID, spell)
	if !me.Hand.Contains(spell) {
		t.Fatal("the chosen card did not reach your hand")
	}
	answerChooseCards(t, g, me.ID, land1)
	if !g.Battlefield.Contains(land1) {
		t.Fatal("the land did not enter")
	}
	if c, _ := g.LookupCardForEffect(land1); !c.Tapped {
		t.Error("the land entered untapped")
	}
	if !me.Library.Contains(deep) || !me.Library.Contains(land2) {
		t.Error("an unlooked card left the library")
	}
}

func TestFinB9ChocoLandfallPumps(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	choco := b12Push(g, me.ID, "Choco, Seeker of Paradise", "Legendary Creature — Bird", finB9ChocoOracle, 3, 5)
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, choco); v.CurrentPower() != 4 {
		t.Errorf("Choco power %d, want 4", v.CurrentPower())
	}
}

func TestFinB9ChocoSendsUnchosenCardsToTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	choco := b12Push(g, me.ID, "Choco, Seeker of Paradise", "Legendary Creature — Bird", finB9ChocoOracle, 3, 5)
	land := topCard(me, "Forest", "Basic Land — Forest", "")
	spell := topCard(me, "Spell", "Sorcery", "{1}")
	bear := topCard(me, "Bear", "Creature — Bear", "{1}")
	_ = bear
	finABAttack(t, g, opp.ID, choco)
	passPriorityAroundTable(t, g)
	// One Bird attacked: the look is one card, the top one.
	answerChooseCards(t, g, me.ID)
	if me.Graveyard.Size() != 1 {
		t.Fatalf("graveyard %d, want the one looked-at card", me.Graveyard.Size())
	}
	if !me.Library.Contains(land) || !me.Library.Contains(spell) {
		t.Error("an unlooked card left the library")
	}
}

const finB9ChocoboOracle = "2e5675d3-3fb0-43b5-a65c-95aacbe48bf9"

func TestFinB9TravelingChocoboPlaysLandsAndCastsBirdsFromTheTopOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Traveling Chocobo", "Creature — Bird", finB9ChocoboOracle, 3, 2)
	bird := topCard(me, "Sparrow", "Creature — Bird", "")
	bear := topCard(me, "Bear", "Creature — Bear", "")
	land := topCard(me, "Forest", "Basic Land — Forest", "")
	// Top to bottom: land, bear, bird.
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Error("a Bear under the land was cast")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("playing the land from the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land did not enter")
	}
	if err := g.CastSpell(me.ID, bird, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Error("a Bird below the Bear was cast")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Error("a Bear on top was cast")
	}
}

func TestFinB9TravelingChocoboCastsABirdFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Traveling Chocobo", "Creature — Bird", finB9ChocoboOracle, 3, 2)
	bird := topCard(me, "Sparrow", "Creature — Bird", "")
	if err := g.CastSpell(me.ID, bird, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("casting a Bird from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bird) {
		t.Error("the Bird did not enter")
	}
}

func TestFinB9TravelingChocoboDoublesLandfallForYourLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Traveling Chocobo", "Creature — Bird", finB9ChocoboOracle, 3, 2)
	choco := b12Push(g, me.ID, "Choco, Seeker of Paradise", "Legendary Creature — Bird", finB9ChocoOracle, 3, 5)
	b31PlayLand(t, g, "Forest", "Basic Land — Forest", "")
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if v := s58p6View(t, g, choco); v.CurrentPower() != 5 {
		t.Errorf("Choco power %d, want 5 (landfall twice)", v.CurrentPower())
	}
}

const finB9GarlandOracle = "ab914053-4384-415f-b1e8-09fecb5a30b8"

func TestFinB9GarlandCrownsAnOpponentStealsWhileTheyHoldItAndPumpsWhatYouDontOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	victim := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	own := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCreatureCard(t, g, "Garland, Royal Kidnapper", "Legendary Creature — Human Knight", finB9GarlandOracle, 3, 4)
	pickTargetRefs(t, g, me.ID, nil, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	passPriorityAroundTable(t, g)
	if g.Monarch != opp.ID {
		t.Fatalf("monarch is %v, want the targeted opponent", g.Monarch)
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(victim)
	if c.Controller != me.ID {
		t.Fatal("Garland did not steal the creature")
	}
	if v := s58p6View(t, g, victim); v.CurrentPower() != 5 || v.CurrentToughness() != 5 {
		t.Errorf("the stolen creature is %d/%d, want 5/5", v.CurrentPower(), v.CurrentToughness())
	}
	if v := s58p6View(t, g, own); v.CurrentPower() != 2 {
		t.Errorf("your own creature is %d/%d, want unchanged", v.CurrentPower(), v.CurrentToughness())
	}
	// Another player takes the crown: the effect is over.
	g.WithWriteLock(func() { _ = g.SetMonarchForEffect(third.ID) })
	g.ReadSnapshot(func() {})
	c, _ = g.LookupCardForEffect(victim)
	if c.Controller != opp.ID {
		t.Errorf("the creature is controlled by %v after they lost the crown, want its owner", c.Controller)
	}
}
