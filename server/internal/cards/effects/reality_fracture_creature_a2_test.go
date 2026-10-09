package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_a2_test.go — the second half of slice
// fra-creature-a's card tests (constants and helpers live in
// reality_fracture_creature_a_test.go).

// --- Arni, Humble Scribe -------------------------------------------

func TestArniHumbleScribeUntapsOnANontokenCreatureOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arni := rfCrAPush(g, me.ID, "Arni, Humble Scribe", "Legendary Creature — Human Wizard", oracleArniScribe, 3, 2)
	b16Tap(g, arni)
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(me.ID, TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{})
	})
	passPriorityAroundTable(t, g)
	if !b16Tapped(t, g, arni) {
		t.Fatal("a creature token untapped Arni")
	}
	rfCrACast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	passPriorityAroundTable(t, g)
	if b16Tapped(t, g, arni) {
		t.Fatal("a nontoken creature did not untap Arni")
	}
}

func TestArniHumbleScribeLoots(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	arni := rfCrAPush(g, me.ID, "Arni, Humble Scribe", "Legendary Creature — Human Wizard", oracleArniScribe, 3, 2)
	before := handSize(me)
	b16Activate(t, g, me.ID, arni, 0, game.ActivateAbilityParams{})
	if !b16Tapped(t, g, arni) {
		t.Fatal("the loot did not tap Arni")
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after the draw, want %d", got, before+1)
	}
	if len(g.PendingChoices) == 0 {
		t.Fatal("no discard prompt followed the draw")
	}
}

// --- Arni, Renowned Champion ---------------------------------------

func TestArniRenownedChampionGainsThePowerOfEachCreatureThatEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	arni := rfCrAPush(g, me.ID, "Arni, Renowned Champion", "Legendary Creature — Human Berserker", oracleArniChampion, 1, 5)
	rfCrACast(t, g, "Ogre", "Creature — Ogre", "", 3, 3)
	passPriorityAroundTable(t, g)
	if got := rfCrAPower(t, g, arni); got != 4 {
		t.Fatalf("power %d after a 3-power creature, want 4", got)
	}
	// An opponent's creature does not trigger it.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	g.WithWriteLock(func() {
		_, _ = g.CreateTokensForEffect(opp.ID, TokenCard("2/2 black Zombie"), 1, game.TokenEntryOptions{})
	})
	passPriorityAroundTable(t, g)
	if got := rfCrAPower(t, g, arni); got != 4 {
		t.Fatalf("power %d after an opponent's creature, want 4", got)
	}
}

// --- Bloombrute ----------------------------------------------------

func TestBloombruteDrawsOncePerTurnOnLifeGain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrAPush(g, me.ID, "Bloombrute", "Creature — Plant Elemental", oracleBloombrute, 4, 4)
	before := handSize(me)
	for i := 0; i < 2; i++ {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
		passPriorityAroundTable(t, g)
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after two life gains, want %d (one draw per turn)", got, before+1)
	}
}

func TestBloombruteGrantsTrampleAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	b := rfCrAPush(g, me.ID, "Bloombrute", "Creature — Plant Elemental", oracleBloombrute, 4, 4)
	bear := rfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrAAddMana(t, g, me, "{C}{C}{C}{C}{G}{W}")
	b16Activate(t, g, me.ID, b, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	for _, kw := range []string{"trample", "lifelink"} {
		if !effectiveAbilitiesContain(t, g, bear, kw) {
			t.Fatalf("Bear lacks %s", kw)
		}
	}
}

// --- Blessed Ghoul -------------------------------------------------

func TestBlessedGhoulReturnsFromTheGraveyardForHybridMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	ghoul := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: ghoul, Name: "Blessed Ghoul", TypeLine: "Creature — Zombie Cleric",
		OracleID: oracleBlessedGhoul, Owner: me.ID, Controller: me.ID})
	rfCrAAddMana(t, g, me, "{C}{C}{B}")
	b16Activate(t, g, me.ID, ghoul, 0, game.ActivateAbilityParams{})
	if !me.Hand.Contains(ghoul) {
		t.Fatal("Blessed Ghoul did not return to hand")
	}
}

// --- Cryotheory Adept ----------------------------------------------

func rfCrAAdeptInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Cryotheory Adept", TypeLine: "Creature — Human Wizard",
		OracleID: oracleCryotheoryAdept, Owner: me.ID, Controller: me.ID})
	return id
}

func TestCryotheoryAdeptTapsAndStunsFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	adept := rfCrAAdeptInGraveyard(g, me)
	target := rfCrAPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrAAddMana(t, g, me, "{C}{C}{C}{U}")
	b16Activate(t, g, me.ID, adept, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}})
	if !g.Exile.Contains(adept) {
		t.Fatal("the Adept was not exiled as a cost")
	}
	if !b16Tapped(t, g, target) {
		t.Fatal("the target was not tapped")
	}
	if got := rfCrACounters(g, target, game.CounterStun); got != 1 {
		t.Fatalf("stun counters %d, want 1", got)
	}
}

func TestCryotheoryAdeptRefusesOutsideAMainPhase(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	adept := rfCrAAdeptInGraveyard(g, me)
	target := rfCrAPush(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	rfCrAAddMana(t, g, me, "{C}{C}{C}{U}")
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, adept, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

// --- Budding Insurgent ---------------------------------------------

func TestBuddingInsurgentDestroysAnArtifactWithoutDrawing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	ins := rfCrAPush(g, me.ID, "Budding Insurgent", "Creature — Dryad Scout", oracleBuddingInsurg, 3, 3)
	rock := rfCrAPush(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	before := handSize(me)
	b16Activate(t, g, me.ID, ins, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}})
	if g.Battlefield.Contains(ins) || g.Battlefield.Contains(rock) {
		t.Fatal("the Insurgent or the artifact is still on the battlefield")
	}
	if got := handSize(me); got != before {
		t.Fatalf("hand %d, want %d: a plain artifact draws nothing", got, before)
	}
}

func TestBuddingInsurgentDrawsForALegendaryEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	ins := rfCrAPush(g, me.ID, "Budding Insurgent", "Creature — Dryad Scout", oracleBuddingInsurg, 3, 3)
	saga := rfCrAPush(g, opp.ID, "Old Wonder", "Legendary Enchantment", "", 0, 0)
	before := handSize(me)
	b16Activate(t, g, me.ID, ins, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: saga}}})
	if g.Battlefield.Contains(saga) {
		t.Fatal("the enchantment survived")
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d, want %d", got, before+1)
	}
}

func TestBuddingInsurgentStillDrawsWhenTheLegendaryEnchantmentIsIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	ins := rfCrAPush(g, me.ID, "Budding Insurgent", "Creature — Dryad Scout", oracleBuddingInsurg, 3, 3)
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Unbreakable", TypeLine: "Legendary Enchantment",
		Keywords: []string{"indestructible"}, Owner: opp.ID, Controller: opp.ID,
	})
	before := handSize(me)
	b16Activate(t, g, me.ID, ins, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tough}}})
	if !g.Battlefield.Contains(tough) {
		t.Fatal("an indestructible enchantment was destroyed")
	}
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d, want %d: it was still a legendary enchantment", got, before+1)
	}
}

func TestBuddingInsurgentIsSorceryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	toMain(t, g)
	ins := rfCrAPush(g, me.ID, "Budding Insurgent", "Creature — Dryad Scout", oracleBuddingInsurg, 3, 3)
	rock := rfCrAPush(g, opp.ID, "Rock", "Artifact", "", 0, 0)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, ins, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}}); err == nil {
		t.Fatal("activated outside a main phase")
	}
}

// --- Danitha, Spear of Agony ---------------------------------------

func TestDanithaSpearGrowsOnSpellsThatTargetOpponentsOrTheirCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d := rfCrAPush(g, me.ID, "Danitha, Spear of Agony", "Legendary Creature — Human Knight", oracleDanithaSpear, 2, 2)
	mine := rfCrAPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrAPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)

	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, d, game.CounterPlusOne); got != 0 {
		t.Fatalf("counters %d after spells aimed at me and my creature, want 0", got)
	}
	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, d, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters %d after a spell aimed at an opponent, want 1", got)
	}
	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, d, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters %d after a spell aimed at an opponent's creature, want 2", got)
	}
}

// --- Danitha, Sword of Hope ----------------------------------------

func TestDanithaSwordDrawsOncePerTurnOnEquipmentOrSpellsAtYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rfCrAPush(g, me.ID, "Danitha, Sword of Hope", "Legendary Creature — Human Knight", oracleDanithaSword, 2, 2)
	mine := rfCrAPush(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrAPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	before := handSize(me)

	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != before {
		t.Fatalf("hand %d after a spell at their creature, want %d", got, before)
	}
	castCatalogSpell(t, g, "Sword of Testing", "Artifact — Equipment", "", nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after an Equipment spell, want %d", got, before+1)
	}
	castCatalogSpell(t, g, "Zap", "Instant", "", []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	passPriorityAroundTable(t, g)
	if got := handSize(me); got != before+1 {
		t.Fatalf("hand %d after a second qualifying spell, want %d (once each turn)", got, before+1)
	}
}

// --- Denzilore Fatehold --------------------------------------------

func TestDenziloreFateholdCountersEachCreatureOnScryAndSurveil(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	d := rfCrAPush(g, me.ID, "Denzilore Fatehold", "Legendary Creature — Elder Sphinx", oracleDenzilore, 3, 4)
	bear := rfCrAPush(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := rfCrAPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	seedLibrary(me, "A", "B", "C")

	g.WithWriteLock(func() { g.ScryForEffect(me.ID, uuid.Nil, 1) })
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no scry prompt")
	}
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{d, bear} {
		if got := rfCrACounters(g, id, game.CounterPlusOne); got != 1 {
			t.Fatalf("counters %d after a scry, want 1", got)
		}
	}
	if got := rfCrACounters(g, theirs, game.CounterPlusOne); got != 0 {
		t.Fatalf("an opponent's creature got %d counters", got)
	}

	g.WithWriteLock(func() { g.SurveilThenForEffect(me.ID, uuid.Nil, 1, nil) })
	s := surveilChoiceFor(g, me.ID)
	if s == nil {
		t.Fatal("no surveil prompt")
	}
	if err := g.ResolveSurveil(s.ID, me.ID, nil, s.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := rfCrACounters(g, d, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters %d after a scry and a surveil, want 2", got)
	}
}

// --- Darklight Phoenix ---------------------------------------------

func rfCrADarklightInGraveyard(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Darklight Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: oracleDarklightPhoen, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
	return id
}

func TestDarklightPhoenixReturnsAfterTwoCreaturesDied(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	phoenix := rfCrADarklightInGraveyard(g, me)
	a := rfCrAPush(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := rfCrAPush(g, me.ID, "B", "Creature — Bear", "", 2, 2)
	rfCrADestroy(t, g, a)
	rfCrADestroy(t, g, b)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not return")
	}
}

func TestDarklightPhoenixStaysDownAfterOneDeath(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	phoenix := rfCrADarklightInGraveyard(g, me)
	a := rfCrAPush(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	rfCrADestroy(t, g, a)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix returned after only one creature died")
	}
}

// --- Darksteel Angel -----------------------------------------------

func TestDarksteelAngelStopsItsControllerLosing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushPermanentForTest(g, me.ID, "Darksteel Angel", oracleDarksteelAngel, "Artifact Creature — Angel")
	me.Life = -3
	checkState(t, g)
	checkState(t, g)
	if me.Eliminated {
		t.Fatal("the controller lost at -3 life")
	}
}

func TestDarksteelAngelStopsMinusCountersOnYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushPermanentForTest(g, me.ID, "Darksteel Angel", oracleDarksteelAngel, "Artifact Creature — Angel")
	mine := rfCrAPush(g, me.ID, "My Bear", "Creature — Bear", "", 3, 3)
	theirs := rfCrAPush(g, opp.ID, "Their Bear", "Creature — Bear", "", 3, 3)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterMinusOne, 2)
		_ = g.AddCounterForEffect(theirs, game.CounterMinusOne, 2)
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
	})
	if got := rfCrACounters(g, mine, game.CounterMinusOne); got != 0 {
		t.Fatalf("my creature has %d -1/-1 counters, want 0", got)
	}
	if got := rfCrACounters(g, mine, game.CounterPlusOne); got != 1 {
		t.Fatalf("my creature has %d +1/+1 counters, want 1", got)
	}
	if got := rfCrACounters(g, theirs, game.CounterMinusOne); got != 2 {
		t.Fatalf("their creature has %d -1/-1 counters, want 2", got)
	}
}
