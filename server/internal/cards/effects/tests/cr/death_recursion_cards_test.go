package testscr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDeathRecursionCardsAreRegistered(t *testing.T) {
	for oracle, want := range map[string]string{
		effectstest.DrcBoneMiser:       "Bone Miser",
		effectstest.DrcCarnivalOfSouls: "Carnival of Souls",
		effectstest.DrcDeathgreeter:    "Deathgreeter",
		effectstest.DrcCockroaches:     "Endless Cockroaches",
		effectstest.DrcGlisteningOil:   "Glistening Oil",
		effectstest.DrcNecroskitter:    "Necroskitter",
		effectstest.DrcNetherTraitor:   "Nether Traitor",
		effectstest.DrcSifterOfSkulls:  "Sifter of Skulls",
		effectstest.DrcVatOfRebirth:    "Vat of Rebirth",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", want)
			continue
		}
		if spec.Name != want || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: name %q completeness %v", want, spec.Name, spec.Completeness)
		}
	}
}

// --- Deathgreeter ----------------------------------------------------

func TestDeathgreeterAsksWhenTheTriggerResolvesNotBeforeItIsOnTheStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", effectstest.DrcDeathgreeter, 1, 1)
	victim := effectstest.DrcBear(g, opp.ID)
	life := me.Life

	effectstest.KillCreature(t, g, opp.ID, victim)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked before the ability was on the stack (the early prompt, #2494)")
	}
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("the trigger never reached the stack")
	}
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Fatal("no question when the trigger resolved")
	}
	if me.Life != life {
		t.Fatal("gained before answering")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	if me.Life != life+1 {
		t.Errorf("life %d -> %d, want +1", life, me.Life)
	}
}

func TestDeathgreeterMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", effectstest.DrcDeathgreeter, 1, 1)
	effectstest.KillCreature(t, g, opp.ID, effectstest.DrcBear(g, opp.ID))
	life := me.Life
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if me.Life != life {
		t.Errorf("declined but life %d -> %d", life, me.Life)
	}
}

func TestDeathgreeterIgnoresItselfAndExileInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	greeter := effectstest.B12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", effectstest.DrcDeathgreeter, 1, 1)
	effectstest.KillCreature(t, g, me.ID, greeter)
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("Deathgreeter asked about its own death")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	effectstest.B12Push(g2, me2.ID, "Deathgreeter", "Creature — Human Shaman", effectstest.DrcDeathgreeter, 1, 1)
	effectstest.PushTokenReplacementCard(g2, effectstest.DrcRestInPeace, "Rest in Peace", "Enchantment", opp2.ID)
	effectstest.KillCreature(t, g2, opp2.ID, effectstest.DrcBear(g2, opp2.ID))
	effectstest.DrcSettle(t, g2, me2.ID)
	if effectstest.LatestConfirmFor(g2, me2.ID) != nil || !effectstest.StackFullyEmpty(g2) {
		t.Error("a creature exiled instead of dying set off Deathgreeter")
	}
	_ = opp
}

func TestDeathgreeterOneQuestionPerSimultaneousDeathAndTokensCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", effectstest.DrcDeathgreeter, 1, 1)
	effectstest.DrcBear(g, me.ID)
	effectstest.DrcBear(g, opp.ID)
	effectstest.MakeTokens(t, g, opp.ID, effects.TokenCard("2/2 black Zombie"), 1)
	life := me.Life

	effectstest.CastCatalogSpell(t, g, "Damnation", "Sorcery", effectstest.DrcDamnationOracle, nil)
	asked := 0
	for i := 0; i < 20; i++ {
		effectstest.DrcSettle(t, g, me.ID)
		if effectstest.LatestConfirmFor(g, me.ID) == nil {
			break
		}
		asked++
		effectstest.AnswerMayChoice(t, g, me.ID, true)
	}
	// Two bears and a token die with the Greeter: three others.
	if asked != 3 || me.Life != life+3 {
		t.Errorf("asked %d times, life %d -> %d; want 3 and +3", asked, life, me.Life)
	}
}

// --- Carnival of Souls -----------------------------------------------

func TestCarnivalOfSoulsIsAnOrdinaryTriggerForAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushTokenReplacementCard(g, effectstest.DrcCarnivalOfSouls, "Carnival of Souls", "Enchantment", me.ID)
	life := me.Life

	effectstest.MakeTokens(t, g, opp.ID, effects.TokenCard("2/2 black Zombie"), 3)
	if effectstest.StackFullyEmpty(g) {
		t.Fatal("an opponent's creatures entering did not put triggers on the stack")
	}
	if me.Life != life || effectstest.DrcManaOfColor(me, "B") != 0 {
		t.Fatal("the triggers resolved without a chance to respond")
	}
	effectstest.DrcSettle(t, g, me.ID)
	if me.Life != life-3 || effectstest.DrcManaOfColor(me, "B") != 3 {
		t.Errorf("life %d -> %d, {B} in pool %d; want -3 and 3", life, me.Life, effectstest.DrcManaOfColor(me, "B"))
	}
}

func TestCarnivalOfSoulsIgnoresNoncreaturePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushTokenReplacementCard(g, effectstest.DrcCarnivalOfSouls, "Carnival of Souls", "Enchantment", me.ID)
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	effectstest.DrcSettle(t, g, me.ID)
	if me.Life != game.StartingLife || effectstest.DrcManaOfColor(me, "B") != 0 {
		t.Errorf("a Treasure entering cost life %d / gave mana %d", me.Life, effectstest.DrcManaOfColor(me, "B"))
	}
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.B12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", effectstest.DrcBoneMiser, 4, 4)
			g.WithWriteLock(func() { me.Hand.Cards = nil })
			effectstest.DrcDiscard(t, g, me, tc.name, tc.typeLine)
			effectstest.DrcSettle(t, g, me.ID)
			if got := effectstest.CountBattlefieldByName(g, "Zombie"); got != tc.zombies {
				t.Errorf("%d Zombies, want %d", got, tc.zombies)
			}
			if got := effectstest.DrcManaOfColor(me, "B"); got != tc.black {
				t.Errorf("%d {B}, want %d", got, tc.black)
			}
			if got := me.Hand.Size(); got != tc.drawn {
				t.Errorf("hand %d, want %d", got, tc.drawn)
			}
		})
	}
}

func TestBoneMiserOnlyCountsYourOwnDiscards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", effectstest.DrcBoneMiser, 4, 4)
	effectstest.DrcDiscard(t, g, opp, "Grizzly Bears", "Creature — Bear")
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.CountBattlefieldByName(g, "Zombie") != 0 {
		t.Error("an opponent's discard made a Zombie")
	}
}

func TestBoneMiserPaysOnEachCardOfOneDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", effectstest.DrcBoneMiser, 4, 4)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	a := effectstest.HandCardFull(me, "Bear A", "Creature — Bear", "{1}", "", nil)
	b := effectstest.HandCardFull(me, "Bear B", "Creature — Bear", "{1}", "", nil)
	effectstest.LengDiscard(t, g, me, a, b)
	effectstest.DrcSettle(t, g, me.ID)
	if got := effectstest.CountBattlefieldByName(g, "Zombie"); got != 2 {
		t.Errorf("%d Zombies, want 2", got)
	}
}

func TestBoneMiserStillPaysWhenLibraryOfLengMovesTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Bone Miser", "Creature — Zombie Wizard", effectstest.DrcBoneMiser, 4, 4)
	effectstest.PushTokenReplacementCard(g, effectstest.DrcLibraryOfLeng, "Library of Leng", "Artifact", me.ID)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	card := effectstest.HandCardFull(me, "Dryad Arbor", "Land Creature — Forest Dryad", "", "", nil)
	effectstest.LengDiscard(t, g, me, card)
	effectstest.AnswerLengPrompt(t, g, me.ID, true)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Library.Contains(card) {
		t.Fatal("Leng did not take the card")
	}
	if effectstest.CountBattlefieldByName(g, "Zombie") != 1 || effectstest.DrcManaOfColor(me, "B") != 2 {
		t.Errorf("Zombies %d, {B} %d; want both payoffs even with the card in the library",
			effectstest.CountBattlefieldByName(g, "Zombie"), effectstest.DrcManaOfColor(me, "B"))
	}
}

// --- Endless Cockroaches ----------------------------------------------

func TestEndlessCockroachesReturnsToItsOwnersHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	roach := effectstest.B12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", effectstest.DrcCockroaches, 1, 1)
	effectstest.KillCreature(t, g, me.ID, roach)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(roach) || me.Graveyard.Contains(roach) {
		t.Fatal("the Cockroaches did not return to hand")
	}

	// Stolen: it dies under an opponent's control and still goes home.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Endless Cockroaches", TypeLine: "Creature — Insect",
		OracleID: effectstest.DrcCockroaches, Power: 1, Toughness: 1, Owner: me.ID, Controller: opp.ID,
	})
	effectstest.KillCreature(t, g, me.ID, stolen)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(stolen) || opp.Hand.Contains(stolen) {
		t.Error("a stolen Cockroaches did not go to its OWNER's hand")
	}
}

func TestEndlessCockroachesIsNotFoundAfterItLeftAndCameBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	roach := effectstest.B12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", effectstest.DrcCockroaches, 1, 1)
	effectstest.KillCreature(t, g, me.ID, roach)
	if effectstest.StackFullyEmpty(g) {
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
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Contains(roach) || !me.Graveyard.Contains(roach) {
		t.Error("the trigger returned a card that had left and come back")
	}
}

func TestEndlessCockroachesDoesNotDieWhenExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	roach := effectstest.B12Push(g, me.ID, "Endless Cockroaches", "Creature — Insect", effectstest.DrcCockroaches, 1, 1)
	effectstest.ExileCreature(t, g, roach)
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Contains(roach) || !effectstest.StackFullyEmpty(g) {
		t.Error("an exiled Cockroaches came back")
	}
}

// --- Glistening Oil ---------------------------------------------------

func TestGlisteningOilGivesInfectToItsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.DrcBear(g, me.ID)
	effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, bear)
	effectstest.AssertKeywords(t, g, bear, game.KeywordInfect)

	effectstest.AttackWith(t, g, opp.ID, bear)
	if effectstest.PoisonOn(opp) != 2 || opp.Life != game.StartingLife {
		t.Errorf("opponent: %d poison, %d life; want 2 poison and %d life", effectstest.PoisonOn(opp), opp.Life, game.StartingLife)
	}
}

func TestGlisteningOilCountsDownOnItsControllersUpkeepOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", effectstest.DrcCreatureTypeLine, 3, 3)
	effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, bear)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.MinusOneOn(t, g, bear); n != 0 {
		t.Fatalf("a counter on the enchanted creature's controller's upkeep: %d", n)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.MinusOneOn(t, g, bear); n != 1 {
		t.Errorf("-1/-1 counters = %d, want 1", n)
	}
	_ = me
}

func TestGlisteningOilReturnsToHandWhenItsCreatureShrinksToNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mouse := effectstest.B12Creature(g, me.ID, "Mouse", "Creature — Mouse", 1, 1)
	oil := effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, mouse)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.DrcSettle(t, g, me.ID)
	if g.Battlefield.Contains(mouse) {
		t.Fatal("the Mouse survived a -1/-1 counter")
	}
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil) || me.Graveyard.Contains(oil) {
		t.Error("the Aura that fell off did not return to its owner's hand")
	}
}

func TestGlisteningOilReturnsWhenItsCreatureDiesOrItIsDestroyed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.DrcBear(g, me.ID)
	oil := effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, bear)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil) {
		t.Fatal("not back in hand after the creature died")
	}

	bear2 := effectstest.DrcBear(g, me.ID)
	oil2 := effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, bear2)
	effectstest.Destroy(t, g, oil2)
	effectstest.DrcSettle(t, g, me.ID)
	if !me.Hand.Contains(oil2) {
		t.Error("not back in hand after being destroyed")
	}
}

func TestGlisteningOilStaysAwayWhenExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.DrcBear(g, me.ID)
	oil := effectstest.AuraCast(t, g, "Glistening Oil", effectstest.DrcGlisteningOil, bear)
	effectstest.ExileCreature(t, g, oil)
	effectstest.DrcSettle(t, g, me.ID)
	if me.Hand.Contains(oil) {
		t.Error("an exiled Aura came back")
	}
}

func TestNecroskitterWitherPutsCountersAndTheDeathReturnsTheCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skitter := effectstest.B12Push(g, me.ID, "Necroskitter", "Creature — Elemental", effectstest.DrcNecroskitter, 1, 4)
	effectstest.AssertKeywords(t, g, skitter, game.KeywordWither)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", effectstest.DrcCreatureTypeLine, 2, 2)

	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(skitter, victim, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	if n := effectstest.MinusOneOn(t, g, victim); n != 1 {
		t.Fatalf("wither put %d -1/-1 counters on the creature, want 1", n)
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.LatestConfirmFor(g, me.ID) == nil {
		t.Fatal("a creature with a -1/-1 counter died and Necroskitter did not ask")
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.Controller != me.ID || c.Owner != opp.ID {
		t.Fatalf("the creature is not on the battlefield under my control: %+v ok=%v", c, ok)
	}
}

func TestNecroskitterAsksAtResolutionAndMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Necroskitter", "Creature — Elemental", effectstest.DrcNecroskitter, 1, 4)
	victim := effectstest.DrcWithMinusOne(g, opp.ID, opp.ID, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("asked before the trigger was on the stack")
	}
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	if g.Battlefield.Contains(victim) || !opp.Graveyard.Contains(victim) {
		t.Error("declined, but the creature moved")
	}
}

func TestNecroskitterNeedsAnOpponentsCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Necroskitter", "Creature — Elemental", effectstest.DrcNecroskitter, 1, 4)
	for _, tc := range []struct {
		name              string
		owner, controller uuid.UUID
		counters          int
	}{
		{"an opponent's creature with no counter", opp.ID, opp.ID, 0},
		{"my own creature with a counter", me.ID, me.ID, 1},
		{"an opponent's creature that I controlled as it died", opp.ID, me.ID, 1},
	} {
		victim := effectstest.DrcWithMinusOne(g, tc.owner, tc.controller, tc.counters)
		effectstest.KillCreature(t, g, tc.owner, victim)
		effectstest.DrcSettle(t, g, me.ID)
		if effectstest.LatestConfirmFor(g, me.ID) != nil {
			t.Errorf("%s: Necroskitter asked", tc.name)
		}
	}
}

func TestNecroskitterTakesAStolenCreatureThatDiedUnderAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Necroskitter", "Creature — Elemental", effectstest.DrcNecroskitter, 1, 4)
	// Mine, stolen by an opponent, dead under their control.
	victim := effectstest.DrcWithMinusOne(g, me.ID, opp.ID, 1)
	effectstest.KillCreature(t, g, me.ID, victim)
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	c, ok := effectstest.BattlefieldCard(g, victim)
	if !ok || c.Controller != me.ID {
		t.Error("a creature that died under an opponent's control did not come back under mine")
	}
}

func TestNecroskitterDoesNothingForATokenOrACardThatLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Necroskitter", "Creature — Elemental", effectstest.DrcNecroskitter, 1, 4)
	tok := effects.TokenCard("2/2 black Zombie")
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = opp.ID, opp.ID
	tok.Counters = map[string]int{game.CounterMinusOne: 1}
	tokenID := effectstest.PushBattlefieldCardWithTimestamp(g, tok)
	effectstest.KillCreature(t, g, opp.ID, tokenID)
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		if c := effectstest.LatestConfirmFor(g, me.ID); c != nil {
			effectstest.AnswerMayChoice(t, g, me.ID, true)
		}
	}
	if g.Battlefield.Contains(tokenID) {
		t.Error("a dead token came back")
	}

	victim := effectstest.DrcWithMinusOne(g, opp.ID, opp.ID, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.DrcSettle(t, g, me.ID)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, victim); err != nil {
		t.Fatalf("exile in response: %v", err)
	}
	if c := effectstest.LatestConfirmFor(g, me.ID); c != nil {
		effectstest.AnswerMayChoice(t, g, me.ID, true)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("a card exiled from the graveyard came back")
	}
}

func TestNetherTraitorIsAskedToPayWhenAnotherCreatureArrivesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	traitor := effectstest.DrcTraitorInGraveyard(g, me)
	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("asked to pay before the ability was on the stack (#2494)")
	}
	effectstest.DrcSettle(t, g, me.ID)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no payment offered")
	}
	effectstest.B06AddMana(me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	c, ok := effectstest.BattlefieldCard(g, traitor)
	if !ok || c.Owner != me.ID || c.Controller != me.ID {
		t.Fatal("Nether Traitor did not return to the battlefield")
	}
	effectstest.AssertKeywords(t, g, traitor, "haste", "shadow")
}

func TestNetherTraitorStaysPutWhenTheCostIsDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	traitor := effectstest.DrcTraitorInGraveyard(g, me)
	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(traitor) || !me.Graveyard.Contains(traitor) {
		t.Error("declined but the Traitor moved")
	}
}

func TestNetherTraitorOnlyWatchesItsOwnersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DrcTraitorInGraveyard(g, me)
	effectstest.KillCreature(t, g, opp.ID, effectstest.DrcBear(g, opp.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a creature in an opponent's graveyard woke the Traitor")
	}
	// A creature I own dies under an opponent's control: it goes to MY
	// graveyard, so it counts.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.DrcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	effectstest.KillCreature(t, g, me.ID, stolen)
	effectstest.DrcSettle(t, g, me.ID)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("my creature dying under an opponent's control did not count")
	}
}

func TestNetherTraitorMustAlreadyBeInTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	traitor := effectstest.B12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", effectstest.DrcNetherTraitor, 1, 1)
	effectstest.DrcBear(g, me.ID)

	effectstest.CastCatalogSpell(t, g, "Damnation", "Sorcery", effectstest.DrcDamnationOracle, nil)
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the Traitor died in the same wipe and still saw the other creature")
	}
	if !me.Graveyard.Contains(traitor) {
		t.Fatal("the Traitor did not die")
	}

	// Later creatures do wake it.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a later death did not wake the Traitor in the graveyard")
	}
}

func TestNetherTraitorDoesNothingOnTheBattlefieldOrIfItLeftTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Nether Traitor", "Creature — Spirit", effectstest.DrcNetherTraitor, 1, 1)
	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a Traitor on the battlefield offered to return itself")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	traitor := effectstest.DrcTraitorInGraveyard(g2, me2)
	effectstest.KillCreature(t, g2, me2.ID, effectstest.DrcBear(g2, me2.ID))
	// Exiled and put back before the trigger resolves: a new object.
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, game.ZoneRef{Kind: game.ZoneExile}, traitor); err != nil {
		t.Fatal(err)
	}
	if err := g2.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me2.ID}, traitor); err != nil {
		t.Fatal(err)
	}
	effectstest.DrcSettle(t, g2, me2.ID)
	if effectstest.HasPayUnlessFor(g2, me2.ID) {
		effectstest.B06AddMana(me2, "B")
		effectstest.AnswerPayUnless(t, g2, me2.ID, true)
	}
	if g2.Battlefield.Contains(traitor) {
		t.Error("the trigger returned a card that had left and come back")
	}
}

// --- Sifter of Skulls -------------------------------------------------

func TestSifterOfSkullsMakesAScionForEachNontokenCreatureYouControlThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sifter := effectstest.B12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", effectstest.DrcSifterOfSkulls, 4, 3)
	effectstest.AssertKeywords(t, g, sifter, game.KeywordDevoid)

	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Fatalf("%d Scions after one creature died, want 1", n)
	}

	// Excluded: a token, an opponent's creature, the Sifter itself.
	tok := effects.TokenCard("2/2 black Zombie")
	tok.InstanceID = uuid.New()
	tok.Owner, tok.Controller = me.ID, me.ID
	effectstest.KillCreature(t, g, me.ID, effectstest.PushBattlefieldCardWithTimestamp(g, tok))
	effectstest.KillCreature(t, g, opp.ID, effectstest.DrcBear(g, opp.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after excluded deaths, want still 1", n)
	}
	effectstest.KillCreature(t, g, me.ID, sifter)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after the Sifter itself died, want still 1", n)
	}
}

func TestSifterOfSkullsWatchesTheControllerAtDeathNotTheOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", effectstest.DrcSifterOfSkulls, 4, 3)
	// Owned by an opponent, but mine while it lived.
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.DrcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: me.ID,
	})
	effectstest.KillCreature(t, g, opp.ID, mine)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions for a creature I controlled, want 1", n)
	}
	// Owned by me, controlled by an opponent: not mine when it died.
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: effectstest.DrcCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: me.ID, Controller: opp.ID,
	})
	effectstest.KillCreature(t, g, me.ID, theirs)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 1 {
		t.Errorf("%d Scions after a creature I owned died under an opponent, want still 1", n)
	}
}

func TestSifterOfSkullsScionHasARealSacrificeManaAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", effectstest.DrcSifterOfSkulls, 4, 3)
	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	scion := effectstest.FindBattlefieldByName(g, "Eldrazi Scion")
	if scion == uuid.Nil {
		t.Fatal("no Scion")
	}
	if err := g.ActivateManaAbility(me.ID, scion, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	if g.Battlefield.Contains(scion) {
		t.Error("the Scion was not sacrificed")
	}
	if effectstest.DrcManaOfColor(me, "C") != 1 {
		t.Errorf("{C} in pool = %d, want 1", effectstest.DrcManaOfColor(me, "C"))
	}
}

func TestSifterOfSkullsOneScionPerCreatureInAWipe(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sifter of Skulls", "Creature — Eldrazi", effectstest.DrcSifterOfSkulls, 4, 3)
	effectstest.DrcBear(g, me.ID)
	effectstest.DrcBear(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Damnation", "Sorcery", effectstest.DrcDamnationOracle, nil)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.CountBattlefieldByName(g, "Eldrazi Scion"); n != 2 {
		t.Errorf("%d Scions, want 2 (one for each Bear; the Sifter is not another)", n)
	}
}

func TestVatOfRebirthCountsArtifactsAndCreaturesYouControlOncePerPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vat := effectstest.B12Push(g, me.ID, "Vat of Rebirth", "Artifact", effectstest.DrcVatOfRebirth, 0, 0)

	effectstest.KillCreature(t, g, me.ID, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	artifact := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, artifact); err != nil {
		t.Fatal(err)
	}
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, vat); n != 2 {
		t.Fatalf("%d oil counters after a creature and an artifact, want 2", n)
	}

	// One artifact creature is one occurrence.
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	effectstest.KillCreature(t, g, me.ID, golem)
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, vat); n != 3 {
		t.Errorf("%d oil counters after an artifact creature, want 3", n)
	}

	// Excluded: an opponent's, a noncreature nonartifact, exile.
	effectstest.KillCreature(t, g, opp.ID, effectstest.DrcBear(g, opp.ID))
	ench := effectstest.B12Permanent(g, me.ID, "Aura", "Enchantment")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, ench); err != nil {
		t.Fatal(err)
	}
	effectstest.ExileCreature(t, g, effectstest.DrcBear(g, me.ID))
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, vat); n != 3 {
		t.Errorf("%d oil counters after excluded departures, want 3", n)
	}
}

func TestVatOfRebirthIgnoresItselfAndCountsTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vat := effectstest.B12Push(g, me.ID, "Vat of Rebirth", "Artifact", effectstest.DrcVatOfRebirth, 0, 0)
	effectstest.MakeTokens(t, g, me.ID, effects.TreasureToken(), 1)
	treasure := effectstest.FindBattlefieldByName(g, "Treasure")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, treasure); err != nil {
		t.Fatal(err)
	}
	effectstest.DrcSettle(t, g, me.ID)
	if n := effectstest.DrcOilOn(t, g, vat); n != 1 {
		t.Errorf("%d oil counters after a Treasure token went to the graveyard, want 1", n)
	}
}

func TestVatOfRebirthPaysExactlyFourCountersAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vat := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: effectstest.DrcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 3},
	})
	dead := effectstest.DrcGraveyardCard(me, "Dead Bear", effectstest.DrcCreatureTypeLine)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}

	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated with only three oil counters")
	}
	if n := effectstest.DrcOilOn(t, g, vat); n != 3 {
		t.Fatalf("a refused activation spent counters: %d", n)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == vat {
				g.Battlefield.Cards[i].Counters["oil"] = 5
			}
		}
	})
	effectstest.B16Activate(t, g, me.ID, vat, 0, game.ActivateAbilityParams{Targets: target})
	if n := effectstest.DrcOilOn(t, g, vat); n != 1 {
		t.Errorf("%d oil counters left after paying four of five, want 1", n)
	}
	effectstest.DrcSettle(t, g, me.ID)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Error("the creature card did not return to the battlefield")
	}
}

func TestVatOfRebirthCannotBeActivatedOutsideYourMainPhaseOrWithAFullStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vat := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: effectstest.DrcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 4},
	})
	dead := effectstest.DrcGraveyardCard(me, "Dead Bear", effectstest.DrcCreatureTypeLine)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}
	effectstest.B06AddMana(me, "B", "C", "C")

	// The harness opens on the turn-1 draw step: not a main phase.
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated in the draw step")
	}

	// A spell on the stack makes it a non-empty stack.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if g.Stack.Size() == 0 {
		t.Fatal("setup: expected a spell on the stack")
	}
	if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Error("activated with a spell on the stack")
	}
	if n := effectstest.DrcOilOn(t, g, vat); n != 4 {
		t.Errorf("a refused activation spent counters: %d", n)
	}
}

func TestVatOfRebirthTargetsOnlyCreatureCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vat := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vat of Rebirth", TypeLine: "Artifact", OracleID: effectstest.DrcVatOfRebirth,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"oil": 4},
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "C", "C")
	theirs := effectstest.DrcGraveyardCard(opp, "Their Bear", effectstest.DrcCreatureTypeLine)
	spell := effectstest.DrcGraveyardCard(me, "Shock", "Instant")
	for name, id := range map[string]uuid.UUID{"opponent's graveyard": theirs, "noncreature card": spell} {
		if err := g.ActivateCatalogAbility(me.ID, vat, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: id}},
		}); err == nil {
			t.Errorf("%s was accepted as a target", name)
		}
	}
}
