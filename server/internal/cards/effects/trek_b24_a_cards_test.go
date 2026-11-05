package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b24_a_cards_test.go — Star Trek, batch 24 (#4973), part A. Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB24DanielsOracle = "2e44a02f-af94-4696-b979-cdf68285c87a"
	trekB24DathonOracle  = "8de259a9-99d0-4f07-8dbc-f065e3af352b"
	trekB24DoctorOracle  = "3b04f5b9-40d9-49aa-a38d-0f946401263c"
	trekB24DurasOracle   = "03973bad-1b3c-4488-a23e-232b192c95cc"
	trekB24TendiOracle   = "6bb42689-b169-43cb-b787-f0e348063cd6"
	trekB24GowronOracle  = "40616029-e1a8-4e7c-bc7b-e381779c57cc"
	trekB24GrayOracle    = "34a39862-c18b-4c1b-8920-207627c40934"
	trekB24HemmerOracle  = "83d8e98d-ab97-40b1-820a-2d193a5ed261"
	trekB24HitItOracle   = "62361438-3da4-4ce9-95b8-fc361b824f58"
	trekB24LightOracle   = "0de3aad2-5f4c-4437-b0b3-4b867309528d"
)

func TestTrekB24EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB24DanielsOracle: "Daniels, Temporal Agent",
		trekB24DathonOracle:  "Dathon, His Heart Open",
		trekB24DoctorOracle:  "The Doctor, EMH",
		trekB24DurasOracle:   "The Duras Sisters",
		trekB24TendiOracle:   "D'Vana Tendi, Optimistic Ensign",
		trekB24GowronOracle:  "Gowron, Eager for Battle",
		trekB24GrayOracle:    "Gray and Adira, Forever Joined",
		trekB24HemmerOracle:  "Hemmer, Fixing What Is Broken",
		trekB24HitItOracle:   "Hit It!",
		trekB24LightOracle:   "The Inner Light",
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

// --- fixtures ------------------------------------------------------------------

// trekB24Graveyard puts a card into p's graveyard and returns its id.
func trekB24Graveyard(g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Graveyard.PushTop(c) })
	return c.InstanceID
}

// trekB24Lands gives p n untapped Forests.
func trekB24Lands(g *game.Game, p uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushLandFor(g, p, "Forest", "Basic Land — Forest")
	}
}

// trekB24Counter reads a battlefield card's counters of one kind.
func trekB24Counter(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := battlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Counters[kind]
}

// trekB24SetCounters sets a battlefield card's counters directly.
func trekB24SetCounters(g *game.Game, id uuid.UUID, counters map[string]int) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = counters
			}
		}
	})
}

// --- Daniels, Temporal Agent --------------------------------------------------

func TestTrekB24DanielsReturnsACreatureCardFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := trekB24Graveyard(g, me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	castHoldCreature(t, g, "Daniels, Temporal Agent", "Legendary Creature — Human Advisor", trekB24DanielsOracle, 3, 1)
	answerPickTargetIfOpen(t, g, bear)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card did not return to your hand")
	}
}

func TestTrekB24DanielsCannotReturnANoncreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := trekB24Graveyard(g, me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	castHoldCreature(t, g, "Daniels, Temporal Agent", "Legendary Creature — Human Advisor", trekB24DanielsOracle, 3, 1)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			for _, id := range c.PickTargetCards {
				if id == bolt {
					t.Fatal("an instant card is offered as a target")
				}
			}
		}
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(bolt) {
		t.Fatal("the instant card left the graveyard")
	}
}

func TestTrekB24DanielsGraveyardAbilityPutsItOnTheBottomAndGainsTwoLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	daniels := trekB24Graveyard(g, me, game.Card{Name: "Daniels, Temporal Agent", TypeLine: "Legendary Creature — Human Advisor",
		OracleID: trekB24DanielsOracle, Power: 3, Toughness: 1})
	trekB24Lands(g, me.ID, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, daniels, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if me.Graveyard.Contains(daniels) {
		t.Fatal("Daniels is still in the graveyard after paying the cost")
	}
	if n := len(me.Library.Cards); n == 0 || me.Library.Cards[0].InstanceID != daniels {
		t.Fatal("Daniels is not the bottom card of the library")
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

func TestTrekB24DanielsGraveyardAbilityNeedsTwoMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	daniels := trekB24Graveyard(g, me, game.Card{Name: "Daniels, Temporal Agent", TypeLine: "Legendary Creature — Human Advisor",
		OracleID: trekB24DanielsOracle, Power: 3, Toughness: 1})
	trekB24Lands(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, daniels, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("activated {2} with one land")
	}
	if !me.Graveyard.Contains(daniels) {
		t.Fatal("a refused activation moved Daniels")
	}
}

func TestTrekB24DanielsGraveyardAbilityDoesNotWorkFromTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	daniels := b12Push(g, me.ID, "Daniels, Temporal Agent", "Legendary Creature — Human Advisor", trekB24DanielsOracle, 3, 1)
	trekB24Lands(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, daniels, 0, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("the graveyard ability was activated from the battlefield")
	}
	if !g.Battlefield.Contains(daniels) {
		t.Fatal("Daniels left the battlefield")
	}
}

// --- Dathon, His Heart Open ---------------------------------------------------

func TestTrekB24DathonDrawsForAnEnchantmentSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hand, life := me.Hand.Size(), opp.Life
	castCatalogSpell(t, g, "Plain Enchantment", "Enchantment", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if opp.Life != life {
		t.Errorf("opponent life %d -> %d: a non-Saga enchantment deals no damage", life, opp.Life)
	}
}

func TestTrekB24DathonSagaSpellAlsoDealsTwoToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Plain Saga", "Enchantment — Saga", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got < 1 {
		t.Errorf("drew %d, want at least 1", got)
	}
	for _, p := range g.Seats {
		want := lives[p.ID] - 2
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %d life = %d, want %d", p.Seat, p.Life, want)
		}
	}
}

func TestTrekB24DathonIgnoresACreatureSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d for a creature spell, want 0", got)
	}
}

func TestTrekB24DathonPutsAnotherCounterOfTheChosenKind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	dathon := b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	ench := b12Permanent(g, me.ID, "Charged Relic", "Enchantment")
	trekB24SetCounters(g, ench, map[string]int{game.CounterCharge: 2, "verse": 1})
	trekB24Lands(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, dathon, 0, game.ActivateAbilityParams{
		Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: ench}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// The kinds are offered sorted: charge, then verse.
	answerOptionPick(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	if got := trekB24Counter(t, g, ench, "verse"); got != 2 {
		t.Errorf("verse counters = %d, want 2", got)
	}
	if got := trekB24Counter(t, g, ench, game.CounterCharge); got != 2 {
		t.Errorf("charge counters = %d, want 2 (unchanged)", got)
	}
}

func TestTrekB24DathonAddsALoreCounterThatTriggersTheChapter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	dathon := b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Inner Light", TypeLine: "Enchantment — Saga", OracleID: trekB24LightOracle,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLore: 2},
	})
	b16Creature(g, me.ID, "Big Beast", "Creature — Beast", 4, 4)
	trekB24Lands(g, me.ID, 2)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, dathon, 0, game.ActivateAbilityParams{
		Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: saga}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life+4 {
		t.Errorf("life = %d, want %d: the third lore counter should run chapter III", me.Life, life+4)
	}
}

func TestTrekB24DathonTargetsOnlyAnEnchantmentYouControlAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	dathon := b12Push(g, me.ID, "Dathon, His Heart Open", "Legendary Creature — Tamarian Officer", trekB24DathonOracle, 3, 3)
	theirs := b12Permanent(g, opp.ID, "Their Relic", "Enchantment")
	trekB24SetCounters(g, theirs, map[string]int{game.CounterCharge: 1})
	mine := b12Permanent(g, me.ID, "My Relic", "Enchantment")
	trekB24SetCounters(g, mine, map[string]int{game.CounterCharge: 1})
	trekB24Lands(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, dathon, 0, game.ActivateAbilityParams{
		Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}},
	}); err == nil {
		t.Fatal("targeted an opponent's enchantment")
	}
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, dathon, 0, game.ActivateAbilityParams{
		Strict: true, AutoTap: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}},
	}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

// --- The Doctor, EMH ----------------------------------------------------------

func TestTrekB24TheDoctorCountersTheTargetThenEachCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	stunned := b16Creature(g, me.ID, "Stunned Bear", "Creature — Bear", 2, 2)
	trekB24SetCounters(g, stunned, map[string]int{game.CounterStun: 1})
	plain := b16Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirsWithCounter := b16Creature(g, opp.ID, "Their Big Bear", "Creature — Bear", 2, 2)
	trekB24SetCounters(g, theirsWithCounter, map[string]int{game.CounterPlusOne: 1})
	doctor := castHoldCreature(t, g, "The Doctor, EMH", "Legendary Artifact Creature — Illusion Doctor", trekB24DoctorOracle, 3, 3)
	answerPickTarget(t, g, mine)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{mine: 2, stunned: 1, plain: 0, doctor: 0, theirs: 0, theirsWithCounter: 1} {
		if got := trekB24Counter(t, g, id, game.CounterPlusOne); got != want {
			t.Errorf("%s: %d +1/+1 counters, want %d", id, got, want)
		}
	}
}

func TestTrekB24TheDoctorOnAnOpponentsCreatureGivesItOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	doctor := castHoldCreature(t, g, "The Doctor, EMH", "Legendary Artifact Creature — Illusion Doctor", trekB24DoctorOracle, 3, 3)
	_ = me
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if got := trekB24Counter(t, g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("the opponent's creature has %d counters, want 1", got)
	}
	if got := trekB24Counter(t, g, doctor, game.CounterPlusOne); got != 0 {
		t.Errorf("The Doctor has %d counters, want 0", got)
	}
}

// --- The Duras Sisters --------------------------------------------------------

func TestTrekB24TheDurasSistersMakeAKlingonWarriorWhenTheyDie(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sisters := b12Push(g, me.ID, "The Duras Sisters", "Legendary Creature — Klingon Warrior", trekB24DurasOracle, 3, 2)
	b15Destroy(t, g, sisters)
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Klingon Warrior" && c.Controller == me.ID {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no Klingon Warrior token")
	}
	if token.Power != 2 || token.Toughness != 1 || !token.IsToken() {
		t.Errorf("token %d/%d token=%v, want a 2/1 token", token.Power, token.Toughness, token.IsToken())
	}
	if !hasEffectiveKeyword(t, g, token.InstanceID, "haste") {
		t.Error("the token has no haste")
	}
}

func TestTrekB24TheDurasSistersExiledMakeNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sisters := b12Push(g, me.ID, "The Duras Sisters", "Legendary Creature — Klingon Warrior", trekB24DurasOracle, 3, 2)
	g.WithWriteLock(func() {
		_ = ExileTarget{Target: sisters}.Apply(NewContext(g, &game.StackItem{Controller: me.ID}))
	})
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Klingon Warrior"); n != 0 {
		t.Errorf("%d tokens after an exile, want 0", n)
	}
}

// --- D'Vana Tendi, Optimistic Ensign ------------------------------------------

func TestTrekB24TendiUntapsAnotherPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tendi := b12Push(g, me.ID, "D'Vana Tendi, Optimistic Ensign", "Legendary Creature — Orion Scientist", trekB24TendiOracle, 2, 4)
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, tendi, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tendi}},
	}); err == nil {
		t.Fatal("Tendi targeted itself")
	}
	if err := g.ActivateCatalogAbility(me.ID, tendi, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, land) {
		t.Error("the land is still tapped")
	}
	if !b16Tapped(t, g, tendi) {
		t.Error("Tendi did not tap")
	}
}

func TestTrekB24TendiSurveilsForThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	tendi := b12Push(g, me.ID, "D'Vana Tendi, Optimistic Ensign", "Legendary Creature — Orion Scientist", trekB24TendiOracle, 2, 4)
	trekB24Lands(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, tendi, 1, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err == nil {
		t.Fatal("surveiled with two lands")
	}
	trekB24Lands(g, me.ID, 1)
	if err := g.ActivateCatalogAbility(me.ID, tendi, 1, game.ActivateAbilityParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("pass: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("no surveil choice opened")
	}
	if b16Tapped(t, g, tendi) {
		t.Error("the surveil ability tapped Tendi")
	}
}
