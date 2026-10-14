package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// death_recursion_cards_test.go — the All Will Be Black death, discard
// and recursion cards (#2952): Bone Miser, Carnival of Souls,
// Deathgreeter, Endless Cockroaches, Glistening Oil, Necroskitter,
// Nether Traitor, Sifter of Skulls and Vat of Rebirth.

const (
	drcBoneMiser        = "3984e5f6-8643-4104-aa9c-18d474e07801"
	drcCarnivalOfSouls  = "95b10ca7-7360-4da5-bd93-686ae3051833"
	drcDeathgreeter     = "b7ad11dc-c962-4034-b15a-a080c6fff75f"
	drcCockroaches      = "9f31dbb1-c350-46f8-bd1d-9f23a073d2f1"
	drcGlisteningOil    = "645a4dee-a4e0-4850-a781-522aa718827f"
	drcNecroskitter     = "536f7c92-2f1e-49e7-98de-47382a9488f1"
	drcNetherTraitor    = "6f16c3ac-a9b8-47e6-b18b-ae37c74d44a0"
	drcSifterOfSkulls   = "688e5e19-8f28-45c4-ace4-a38144f494f1"
	drcVatOfRebirth     = "10c31317-71e8-42e0-85e0-3e64bd0c3dd3"
	drcRestInPeace      = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
	drcDamnationOracle  = "d57a8f0b-7989-4db5-8756-6f2690097252"
	drcLibraryOfLeng    = "867def48-4be8-4056-bcf1-d6b00450b9a3"
	drcCreatureTypeLine = "Creature — Bear"
)

func TestDeathRecursionCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]string{
		drcBoneMiser:       "Bone Miser",
		drcCarnivalOfSouls: "Carnival of Souls",
		drcDeathgreeter:    "Deathgreeter",
		drcCockroaches:     "Endless Cockroaches",
		drcGlisteningOil:   "Glistening Oil",
		drcNecroskitter:    "Necroskitter",
		drcNetherTraitor:   "Nether Traitor",
		drcSifterOfSkulls:  "Sifter of Skulls",
		drcVatOfRebirth:    "Vat of Rebirth",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// drcSettle passes priority until the stack is empty or a prompt other
// than a trigger-order one is open, ordering simultaneous triggers as
// offered.
func drcSettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if p := triggerOrderPromptFor(g, chooser); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, chooser, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("did not settle")
}

func drcBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b12Creature(g, owner, "Bear", drcCreatureTypeLine, 2, 2)
}

func drcManaOfColor(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool {
		if tok.Color == color {
			n++
		}
	}
	return n
}

// --- Deathgreeter ----------------------------------------------------

func TestDeathgreeterAsksWhenTheTriggerResolvesNotBeforeItIsOnTheStack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	victim := drcBear(g, opp.ID)
	life := me.Life

	killCreature(t, g, opp.ID, victim)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked before the ability was on the stack (the early prompt, #2494)")
	}
	if stackFullyEmpty(g) {
		t.Fatal("the trigger never reached the stack")
	}
	drcSettle(t, g, me.ID)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("no question when the trigger resolved")
	}
	if me.Life != life {
		t.Fatal("gained before answering")
	}
	answerMayChoice(t, g, me.ID, true)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestDeathgreeterMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	killCreature(t, g, opp.ID, drcBear(g, opp.ID))
	life := me.Life
	drcSettle(t, g, me.ID)
	answerMayChoice(t, g, me.ID, false)
	if me.Life != life {
		t.Errorf("declined but life %d -> %d", life, me.Life)
	}
}

func TestDeathgreeterIgnoresItselfAndExileInsteadOfDying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greeter := b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	killCreature(t, g, me.ID, greeter)
	drcSettle(t, g, me.ID)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("Deathgreeter asked about its own death")
	}

	g2 := newCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	b12Push(g2, me2.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	pushTokenReplacementCard(g2, drcRestInPeace, "Rest in Peace", "Enchantment", opp2.ID)
	killCreature(t, g2, opp2.ID, drcBear(g2, opp2.ID))
	drcSettle(t, g2, me2.ID)
	if latestConfirmFor(g2, me2.ID) != nil || !stackFullyEmpty(g2) {
		t.Error("a creature exiled instead of dying set off Deathgreeter")
	}
	_ = opp
}

func TestDeathgreeterOneQuestionPerSimultaneousDeathAndTokensCount(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	drcBear(g, me.ID)
	drcBear(g, opp.ID)
	makeTokens(t, g, opp.ID, TokenCard("2/2 black Zombie"), 1)
	life := me.Life

	castCatalogSpell(t, g, "Damnation", "Sorcery", drcDamnationOracle, nil)
	asked := 0
	for i := 0; i < 20; i++ {
		drcSettle(t, g, me.ID)
		if latestConfirmFor(g, me.ID) == nil {
			break
		}
		asked++
		answerMayChoice(t, g, me.ID, true)
	}
	// Two bears and a token die with the Greeter: three others.
	if asked != 3 || me.Life != life+3 {
		t.Errorf("asked %d times, life %d -> %d; want 3 and +3", asked, life, me.Life)
	}
}

// --- Carnival of Souls -----------------------------------------------

func TestCarnivalOfSoulsIsAnOrdinaryTriggerForAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTokenReplacementCard(g, drcCarnivalOfSouls, "Carnival of Souls", "Enchantment", me.ID)
	life := me.Life

	makeTokens(t, g, opp.ID, TokenCard("2/2 black Zombie"), 3)
	if stackFullyEmpty(g) {
		t.Fatal("an opponent's creatures entering did not put triggers on the stack")
	}
	if me.Life != life || drcManaOfColor(me, "B") != 0 {
		t.Fatal("the triggers resolved without a chance to respond")
	}
	drcSettle(t, g, me.ID)
	if me.Life != life-3 || drcManaOfColor(me, "B") != 3 {
		t.Errorf("life %d -> %d, {B} in pool %d; want -3 and 3", life, me.Life, drcManaOfColor(me, "B"))
	}
}

func TestCarnivalOfSoulsIgnoresNoncreaturePermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTokenReplacementCard(g, drcCarnivalOfSouls, "Carnival of Souls", "Enchantment", me.ID)
	makeTokens(t, g, me.ID, TreasureToken(), 1)
	drcSettle(t, g, me.ID)
	if me.Life != game.StartingLife || drcManaOfColor(me, "B") != 0 {
		t.Errorf("a Treasure entering cost life %d / gave mana %d", me.Life, drcManaOfColor(me, "B"))
	}
}

// --- Bone Miser ------------------------------------------------------

func drcDiscard(t *testing.T, g *game.Game, who *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	card := handCardFull(who, name, typeLine, "{1}", "", nil)
	lengDiscard(t, g, who, card)
	return card
}

func TestBoneMiserDiscardBranches(t *testing.T) {
	cases := []struct {
		name, typeLine        string
		zombies, black, drawn int
	}{
		{"Grizzly Bears", "Creature — Bear", 1, 0, 0},
		{"Forest", "Basic Land — Forest", 0, 2, 0},
		{"Lightning Bolt", "Instant", 0, 0, 1},
		{"Sol Ring", "Artifact", 0, 0, 1},
		{"Dryad Arbor", "Land Creature — Forest Dryad", 1, 2, 0},
		{"Mutavault", "Land", 0, 2, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			b12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", drcBoneMiser, 4, 4)
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			drcDiscard(t, g, me, tc.name, tc.typeLine)
			drcSettle(t, g, me.ID)
			if got := countBattlefieldByName(g, "Zombie"); got != tc.zombies {
				t.Errorf("%d Zombies, want %d", got, tc.zombies)
			}
			if got := drcManaOfColor(me, "B"); got != tc.black {
				t.Errorf("%d {B}, want %d", got, tc.black)
			}
			if got := me.Hand.Size(); got != tc.drawn {
				t.Errorf("hand %d, want %d", got, tc.drawn)
			}
		})
	}
}

func TestBoneMiserOnlyCountsYourOwnDiscards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", drcBoneMiser, 4, 4)
	drcDiscard(t, g, opp, "Grizzly Bears", "Creature — Bear")
	drcSettle(t, g, me.ID)
	if countBattlefieldByName(g, "Zombie") != 0 {
		t.Error("an opponent's discard made a Zombie")
	}
}

func TestBoneMiserPaysOnEachCardOfOneDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", drcBoneMiser, 4, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := handCardFull(me, "Bear A", "Creature — Bear", "{1}", "", nil)
	b := handCardFull(me, "Bear B", "Creature — Bear", "{1}", "", nil)
	lengDiscard(t, g, me, a, b)
	drcSettle(t, g, me.ID)
	if got := countBattlefieldByName(g, "Zombie"); got != 2 {
		t.Errorf("%d Zombies, want 2", got)
	}
}

func TestBoneMiserStillPaysWhenLibraryOfLengMovesTheDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", drcBoneMiser, 4, 4)
	pushTokenReplacementCard(g, drcLibraryOfLeng, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := handCardFull(me, "Dryad Arbor", "Land Creature — Forest Dryad", "", "", nil)
	lengDiscard(t, g, me, card)
	answerLengPrompt(t, g, me.ID, true)
	drcSettle(t, g, me.ID)
	if !me.Library.Contains(card) {
		t.Fatal("Leng did not take the card")
	}
	if countBattlefieldByName(g, "Zombie") != 1 || drcManaOfColor(me, "B") != 2 {
		t.Errorf("Zombies %d, {B} %d; want both payoffs even with the card in the library",
			countBattlefieldByName(g, "Zombie"), drcManaOfColor(me, "B"))
	}
}

// --- Endless Cockroaches ----------------------------------------------

func TestEndlessCockroachesReturnsToItsOwnersHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roach := b12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", drcCockroaches, 1, 1)
	killCreature(t, g, me.ID, roach)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(roach) || me.Graveyard.Contains(roach) {
		t.Fatal("the Cockroaches did not return to hand")
	}

	// Stolen: it dies under an opponent's control and still goes home.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Endless Cockroaches", TypeLine: "Creature — Insect",
		OracleID: drcCockroaches, Power: 1, Toughness: 1, Owner: me.ID, Controller: opp.ID,
	})
	killCreature(t, g, me.ID, stolen)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(stolen) || opp.Hand.Contains(stolen) {
		t.Error("a stolen Cockroaches did not go to its OWNER's hand")
	}
}

func TestEndlessCockroachesIsNotFoundAfterItLeftAndCameBack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	roach := b12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", drcCockroaches, 1, 1)
	killCreature(t, g, me.ID, roach)
	if stackFullyEmpty(g) {
		t.Fatal("no trigger")
	}
	// In response the card is exiled and put back into the graveyard:
	// a new object the trigger knows nothing about (CR 400.7).
	for _, to := range []game.ZoneRef{{Kind: game.ZoneExile}, {Kind: game.ZoneGraveyard, Owner: me.ID}} {
		from := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}
		if to.Kind == game.ZoneGraveyard {
			from = game.ZoneRef{Kind: game.ZoneExile}
		}
		if err := g.MoveCardByID(from, to, roach); err != nil {
			t.Fatalf("MoveCardByID: %v", err)
		}
	}
	drcSettle(t, g, me.ID)
	if me.Hand.Contains(roach) || !me.Graveyard.Contains(roach) {
		t.Error("the trigger returned a card that had left and come back")
	}
}

func TestEndlessCockroachesDoesNotDieWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	roach := b12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", drcCockroaches, 1, 1)
	exileCreature(t, g, roach)
	drcSettle(t, g, me.ID)
	if me.Hand.Contains(roach) || !stackFullyEmpty(g) {
		t.Error("an exiled Cockroaches came back")
	}
}

// --- Glistening Oil ---------------------------------------------------

func TestGlisteningOilGivesInfectToItsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := drcBear(g, me.ID)
	auraCast(t, g, "Glistening Oil", drcGlisteningOil, bear)
	assertKeywords(t, g, bear, game.KeywordInfect)

	attackWith(t, g, opp.ID, bear)
	if poisonOn(opp) != 2 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 2 poison and %d life", poisonOn(opp), opp.Life, game.StartingLife)
	}
}

func TestGlisteningOilCountsDownOnItsControllersUpkeepOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", drcCreatureTypeLine, 3, 3)
	auraCast(t, g, "Glistening Oil", drcGlisteningOil, bear)

	advanceToUpkeepOf(t, g, 1)
	drcSettle(t, g, me.ID)
	if n := minusOneOn(t, g, bear); n != 0 {
		t.Fatalf("a counter on the enchanted creature's controller's upkeep: %d", n)
	}
	advanceToUpkeepOf(t, g, 0)
	drcSettle(t, g, me.ID)
	if n := minusOneOn(t, g, bear); n != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", n)
	}
	_ = me
}

func TestGlisteningOilReturnsToHandWhenItsCreatureShrinksToNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mouse := b12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	oil := auraCast(t, g, "Glistening Oil", drcGlisteningOil, mouse)

	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	drcSettle(t, g, me.ID)
	if g.Battlefield.Contains(mouse) {
		t.Fatal("the Mouse survived a -1/-1 counter")
	}
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil) || me.Graveyard.Contains(oil) {
		t.Error("the Aura that fell off did not return to its owner's hand")
	}
}

func TestGlisteningOilReturnsWhenItsCreatureDiesOrItIsDestroyed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := drcBear(g, me.ID)
	oil := auraCast(t, g, "Glistening Oil", drcGlisteningOil, bear)
	killCreature(t, g, me.ID, bear)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil) {
		t.Fatal("not back in hand after the creature died")
	}

	bear2 := drcBear(g, me.ID)
	oil2 := auraCast(t, g, "Glistening Oil", drcGlisteningOil, bear2)
	destroy(t, g, oil2)
	drcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil2) {
		t.Error("not back in hand after being destroyed")
	}
}

func TestGlisteningOilStaysAwayWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := drcBear(g, me.ID)
	oil := auraCast(t, g, "Glistening Oil", drcGlisteningOil, bear)
	exileCreature(t, g, oil)
	drcSettle(t, g, me.ID)
	if me.Hand.Contains(oil) {
		t.Error("an exiled Aura came back")
	}
}

// --- Necroskitter -----------------------------------------------------

func drcWithMinusOne(g *game.Game, owner, controller uuid.UUID, counters int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: "Victim", TypeLine: drcCreatureTypeLine,
		Power: 2, Toughness: 4, Owner: owner, Controller: controller,
	}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterMinusOne: counters}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestNecroskitterWitherPutsCountersAndTheDeathReturnsTheCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skitter := b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	assertKeywords(t, g, skitter, game.KeywordWither)
	victim := b12Creature(g, opp.ID, "Victim", drcCreatureTypeLine, 2, 2)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(skitter, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if n := minusOneOn(t, g, victim); n != 1 {
		t.Fatalf("wither put %d -1/-1 counters on the creature, want 1", n)
	}
	killCreature(t, g, opp.ID, victim)
	drcSettle(t, g, me.ID)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("a creature with a -1/-1 counter died and Necroskitter did not ask")
	}
	answerMayChoice(t, g, me.ID, true)
	c, ok := battlefieldCard(g, victim)
	if !ok || c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("the creature is not on the battlefield under my control: %+v ok=%v", c, ok)
	}
}

func TestNecroskitterAsksAtResolutionAndMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	victim := drcWithMinusOne(g, opp.ID, opp.ID, 1)
	killCreature(t, g, opp.ID, victim)
	if latestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked before the trigger was on the stack")
	}
	drcSettle(t, g, me.ID)
	answerMayChoice(t, g, me.ID, false)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("declined, but the creature moved")
	}
}

func TestNecroskitterNeedsAnOpponentsCreatureWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	for _, tc := range []struct {
		name              string
		owner, controller uuid.UUID
		counters          int
	}{
		{"an opponent's creature with no counter", opp.ID, opp.ID, 0},
		{"my own creature with a counter", me.ID, me.ID, 1},
		{"an opponent's creature that I controlled as it died", opp.ID, me.ID, 1},
	} {
		victim := drcWithMinusOne(g, tc.owner, tc.controller, tc.counters)
		killCreature(t, g, tc.owner, victim)
		drcSettle(t, g, me.ID)
		if latestConfirmFor(g, me.ID) != nil {
			t.Errorf("%s: Necroskitter asked", tc.name)
		}
	}
}

func TestNecroskitterTakesAStolenCreatureThatDiedUnderAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	// Mine, stolen by an opponent, dead under their control.
	victim := drcWithMinusOne(g, me.ID, opp.ID, 1)
	killCreature(t, g, me.ID, victim)
	drcSettle(t, g, me.ID)
	answerMayChoice(t, g, me.ID, true)
	c, ok := battlefieldCard(g, victim)
	if !ok || c.Controller != me.ID {
		t.Error("a creature that died under an opponent's control did not come back under mine")
	}
}

func TestNecroskitterDoesNothingForATokenOrACardThatLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	tok := TokenCard("2/2 black Zombie")
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = opp.ID, opp.ID
	tok.Counters = map[string]int{game.CounterMinusOne: 1}
	tokenID := pushBattlefieldCardWithTimestamp(g, tok)
	killCreature(t, g, opp.ID, tokenID)
	drcSettle(t, g, me.ID)
	if latestConfirmFor(g, me.ID) != nil {
		if c := latestConfirmFor(g, me.ID); c != nil {
			answerMayChoice(t, g, me.ID, true)
		}
	}
	if g.Battlefield.Contains(tokenID) {
		t.Error("a dead token came back")
	}

	victim := drcWithMinusOne(g, opp.ID, opp.ID, 1)
	killCreature(t, g, opp.ID, victim)
	drcSettle(t, g, me.ID)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, victim); err != nil {
		t.Fatalf("exile in response: %v", err)
	}
	if c := latestConfirmFor(g, me.ID); c != nil {
		answerMayChoice(t, g, me.ID, true)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("a card exiled from the graveyard came back")
	}
}

// --- Nether Traitor ---------------------------------------------------

func drcTraitorInGraveyard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Nether Traitor", TypeLine: "Creature — Spirit", OracleID: drcNetherTraitor,
			Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID, ManaCost: "{B}{B}",
		})
	})
	return id
}

func TestNetherTraitorIsAskedToPayWhenAnotherCreatureArrivesInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := drcTraitorInGraveyard(g, me)
	killCreature(t, g, me.ID, drcBear(g, me.ID))
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to pay before the ability was on the stack (#2494)")
	}
	drcSettle(t, g, me.ID)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment offered")
	}
	b06AddMana(me, "B")
	answerPayUnless(t, g, me.ID, true)
	c, ok := battlefieldCard(g, traitor)
	if !ok || c.Owner != me.ID || c.Controller != me.ID {
		t.Fatal("Nether Traitor did not return to the battlefield")
	}
	assertKeywords(t, g, traitor, "haste", "shadow")
}

func TestNetherTraitorStaysPutWhenTheCostIsDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := drcTraitorInGraveyard(g, me)
	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	answerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(traitor) || !me.Graveyard.Contains(traitor) {
		t.Error("declined but the Traitor moved")
	}
}

func TestNetherTraitorOnlyWatchesItsOwnersGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drcTraitorInGraveyard(g, me)
	killCreature(t, g, opp.ID, drcBear(g, opp.ID))
	drcSettle(t, g, me.ID)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a creature in an opponent's graveyard woke the Traitor")
	}
	// A creature I own dies under an opponent's control: it goes to MY
	// graveyard, so it counts.
	stolen := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: drcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	killCreature(t, g, me.ID, stolen)
	drcSettle(t, g, me.ID)
	if !hasPayUnlessFor(g, me.ID) {
		t.Error("my creature dying under an opponent's control did not count")
	}
}

func TestNetherTraitorMustAlreadyBeInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	traitor := b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
	drcBear(g, me.ID)

	castCatalogSpell(t, g, "Damnation", "Sorcery", drcDamnationOracle, nil)
	drcSettle(t, g, me.ID)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("the Traitor died in the same wipe and still saw the other creature")
	}
	if !me.Graveyard.Contains(traitor) {
		t.Fatal("the Traitor did not die")
	}

	// Later creatures do wake it.
	advanceTo(t, g, game.StepBeginCombat)
	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	if !hasPayUnlessFor(g, me.ID) {
		t.Error("a later death did not wake the Traitor in the graveyard")
	}
}

func TestNetherTraitorDoesNothingOnTheBattlefieldOrIfItLeftTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", drcNetherTraitor, 1, 1)
	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	if hasPayUnlessFor(g, me.ID) {
		t.Fatal("a Traitor on the battlefield offered to return itself")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	traitor := drcTraitorInGraveyard(g2, me2)
	killCreature(t, g2, me2.ID, drcBear(g2, me2.ID))
	// Exiled and put back before the trigger resolves: a new object.
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, game.ZoneRef{Kind: game.ZoneExile}, traitor); err != nil {
		t.Fatal(err)
	}
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, traitor); err != nil {
		t.Fatal(err)
	}
	drcSettle(t, g2, me2.ID)
	if hasPayUnlessFor(g2, me2.ID) {
		b06AddMana(me2, "B")
		answerPayUnless(t, g2, me2.ID, true)
	}
	if g2.Battlefield.Contains(traitor) {
		t.Error("the trigger returned a card that had left and come back")
	}
}

// --- Sifter of Skulls -------------------------------------------------

func TestSifterOfSkullsMakesAScionForEachNontokenCreatureYouControlThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sifter := b12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", drcSifterOfSkulls, 4, 3)
	assertKeywords(t, g, sifter, game.KeywordDevoid)

	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Fatalf("%d Scions after one creature died, want 1", n)
	}

	// Excluded: a token, an opponent's creature, the Sifter itself.
	tok := TokenCard("2/2 black Zombie")
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	killCreature(t, g, me.ID, pushBattlefieldCardWithTimestamp(g, tok))
	killCreature(t, g, opp.ID, drcBear(g, opp.ID))
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after excluded deaths, want still 1", n)
	}
	killCreature(t, g, me.ID, sifter)
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after the Sifter itself died, want still 1", n)
	}
}

func TestSifterOfSkullsWatchesTheControllerAtDeathNotTheOwner(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", drcSifterOfSkulls, 4, 3)
	// Owned by an opponent, but mine while it lived.
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: drcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	killCreature(t, g, opp.ID, mine)
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions for a creature I controlled, want 1", n)
	}
	// Owned by me, controlled by an opponent: not mine when it died.
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: drcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	killCreature(t, g, me.ID, theirs)
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after a creature I owned died under an opponent, want still 1", n)
	}
}

func TestSifterOfSkullsScionHasARealSacrificeManaAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", drcSifterOfSkulls, 4, 3)
	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	scion := findBattlefieldByName(g, "Eldrazi Scion")
	if scion == uuid.Nil {
		t.Fatal("no Scion")
	}
	if err := g.ActivateManaAbility(me.ID, scion, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(scion) {
		t.Error("the Scion was not sacrificed")
	}
	if drcManaOfColor(me, "C") != 1 {
		t.Errorf("{C} in pool = %d, want 1", drcManaOfColor(me, "C"))
	}
}

func TestSifterOfSkullsOneScionPerCreatureInAWipe(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", drcSifterOfSkulls, 4, 3)
	drcBear(g, me.ID)
	drcBear(g, me.ID)
	castCatalogSpell(t, g, "Damnation", "Sorcery", drcDamnationOracle, nil)
	drcSettle(t, g, me.ID)
	if n := countBattlefieldByName(g, "Eldrazi Scion"); n != 2 {
		t.Errorf("%d Scions, want 2 (one for each Bear; the Sifter is not another)", n)
	}
}

// --- Vat of Rebirth ---------------------------------------------------

func drcOilOn(t *testing.T, g *game.Game, vat uuid.UUID) int {
	t.Helper()
	c := battlefieldCardCopy(t, g, vat)
	return c.Counters["oil"]
}

func TestVatOfRebirthCountsArtifactsAndCreaturesYouControlOncePerPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vat := b12Push(g, me.ID, "Vat of Rebirth", "Artifact", drcVatOfRebirth, 0, 0)

	killCreature(t, g, me.ID, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	artifact := b12Permanent(g, me.ID, "Rock", "Artifact")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, artifact); err != nil {
		t.Fatal(err)
	}
	drcSettle(t, g, me.ID)
	if n := drcOilOn(t, g, vat); n != 2 {
		t.Fatalf("%d oil counters after a creature and an artifact, want 2", n)
	}

	// One artifact creature is one occurrence.
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	killCreature(t, g, me.ID, golem)
	drcSettle(t, g, me.ID)
	if n := drcOilOn(t, g, vat); n != 3 {
		t.Errorf("%d oil counters after an artifact creature, want 3", n)
	}

	// Excluded: an opponent's, a noncreature nonartifact, exile.
	killCreature(t, g, opp.ID, drcBear(g, opp.ID))
	ench := b12Permanent(g, me.ID, "Aura", "Enchantment")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, ench); err != nil {
		t.Fatal(err)
	}
	exileCreature(t, g, drcBear(g, me.ID))
	drcSettle(t, g, me.ID)
	if n := drcOilOn(t, g, vat); n != 3 {
		t.Errorf("%d oil counters after excluded departures, want 3", n)
	}
}

func TestVatOfRebirthIgnoresItselfAndCountsTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vat := b12Push(g, me.ID, "Vat of Rebirth", "Artifact", drcVatOfRebirth, 0, 0)
	makeTokens(t, g, me.ID, TreasureToken(), 1)
	treasure := findBattlefieldByName(g, "Treasure")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, treasure); err != nil {
		t.Fatal(err)
	}
	drcSettle(t, g, me.ID)
	if n := drcOilOn(t, g, vat); n != 1 {
		t.Errorf("%d oil counters after a Treasure token went to the graveyard, want 1", n)
	}
}

func TestVatOfRebirthPaysExactlyFourCountersAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: drcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 3},
	})
	dead := drcGraveyardCard(me, "Dead Bear", drcCreatureTypeLine)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}

	advanceToMain(t, g)
	b06AddMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated with only three oil counters")
	}
	if n := drcOilOn(t, g, vat); n != 3 {
		t.Fatalf("a refused activation spent counters: %d", n)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vat {
				g.Battlefield.Cards[i].Counters["oil"] = 5
			}
		}
	})
	b16Activate(t, g, me.ID, vat, 0, game.ActivateAbilityParams{Targets: target})
	if n := drcOilOn(t, g, vat); n != 1 {
		t.Errorf("%d oil counters left after paying four of five, want 1", n)
	}
	drcSettle(t, g, me.ID)
	c, ok := battlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestVatOfRebirthCannotBeActivatedOutsideYourMainPhaseOrWithAFullStack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: drcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 4},
	})
	dead := drcGraveyardCard(me, "Dead Bear", drcCreatureTypeLine)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}
	b06AddMana(me, "B", "C", "C")

	// The harness opens on the turn-1 draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated in the draw step")
	}

	// A spell on the stack makes it a non-empty stack.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if g.Stack.Size() == 0 {
		t.Fatal("setup: expected a spell on the stack")
	}
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Error("activated with a spell on the stack")
	}
	if n := drcOilOn(t, g, vat); n != 4 {
		t.Errorf("a refused activation spent counters: %d", n)
	}
}

func TestVatOfRebirthTargetsOnlyCreatureCardsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vat := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: drcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 4},
	})
	advanceToMain(t, g)
	b06AddMana(me, "B", "C", "C")
	theirs := drcGraveyardCard(opp, "Their Bear", drcCreatureTypeLine)
	spell := drcGraveyardCard(me, "Shock", "Instant")
	for name, id := range map[string]uuid.UUID{"opponent's graveyard": theirs, "noncreature card": spell} {
		if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		}); err == nil {
			t.Errorf("%s was accepted as a target", name)
		}
	}
}

// pushGraveyardCard puts a card of the given type line into p's graveyard.
func drcGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}
