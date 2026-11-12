package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// whenever_you_proliferate_test.go — #5653: game.EventProliferate, the
// post-action fact of CR 701.34a, and the four Phyrexia: All Will Be One
// cards that watch it.

const (
	ezuriStalkerOracle       = "cffec16f-373b-43e7-a04c-4452987d7e02"
	schemingAspirantOracle   = "cf4a3b1a-4b87-473e-be1f-3aa3da45c28e"
	venserCorpsePuppetOracle = "17c934a7-3480-4309-a217-ca2550538ae7"
	voidwingHybridOracle     = "0dab592c-a230-408d-a383-044e561dbe5d"
)

// startProliferate has `player` proliferate once as an effect of
// `source`, the choice left to them (the catalog's Proliferate{}).
func startProliferate(t *testing.T, g *game.Game, player, source uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.ProliferateChoosingKeyedForEffect(player, source, game.KeyedTail{}); err != nil {
			t.Fatalf("ProliferateChoosingKeyedForEffect: %v", err)
		}
	})
}

// proliferateEvents is every EventProliferate in the log.
func proliferateEvents(g *game.Game) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == game.EventProliferate {
			out = append(out, ev)
		}
	}
	return out
}

// CR 701.34a: the event is raised once the choice is made and the
// counters are placed, naming the proliferating player and the card.
func TestProliferateEventFollowsTheCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	startProliferate(t, g, me.ID, mine)
	if n := len(proliferateEvents(g)); n != 0 {
		t.Fatalf("%d proliferate events while the choice is still open, want 0", n)
	}
	answerProliferate(t, g, mine)

	evs := proliferateEvents(g)
	if len(evs) != 1 {
		t.Fatalf("%d proliferate events, want 1", len(evs))
	}
	if evs[0].Actor != me.ID || evs[0].Source != mine {
		t.Errorf("event actor/source = %v/%v, want the proliferating player and the card", evs[0].Actor, evs[0].Source)
	}
	placed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventCounterPlaced && ev.Target == mine && ev.Seq < evs[0].Seq {
			placed = true
		}
	}
	if !placed || counterCount(g, mine, game.CounterPlusOne) != 2 {
		t.Error("the proliferate event did not follow its counter")
	}
}

// "Any number" includes none: a proliferate that chose nothing, and one
// with nothing on the table to choose, are both proliferates.
func TestProliferateEventForAProliferateOfNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	startProliferate(t, g, me.ID, uuid.Nil)
	if proliferatePrompt(g) != nil {
		t.Fatal("a board with no counters asked what to proliferate")
	}
	if n := len(proliferateEvents(g)); n != 1 {
		t.Fatalf("%d proliferate events on an empty board, want 1", n)
	}

	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	startProliferate(t, g, me.ID, mine)
	answerProliferate(t, g)
	if n := len(proliferateEvents(g)); n != 2 {
		t.Errorf("%d proliferate events after choosing nothing, want 2", n)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("an unchosen permanent has %d counters, want 1", got)
	}
}

// "Proliferate twice" (Tekuthal) is two proliferates: one event after
// each one's counters, on the asked path, the explicit path and an
// empty board alike.
func TestProliferateTwiceRaisesTwoEvents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)

	startProliferate(t, g, me.ID, uuid.Nil)
	if n := len(proliferateEvents(g)); n != 2 {
		t.Fatalf("%d proliferate events for two proliferates of an empty board, want 2", n)
	}

	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	startProliferate(t, g, me.ID, mine)
	answerProliferate(t, g, mine)
	if n := len(proliferateEvents(g)); n != 3 {
		t.Fatalf("%d proliferate events after the first of two answers, want 3", n)
	}
	answerProliferate(t, g, mine)
	if n := len(proliferateEvents(g)); n != 4 {
		t.Fatalf("%d proliferate events after the second answer, want 4", n)
	}

	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, mine, []uuid.UUID{mine}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if n := len(proliferateEvents(g)); n != 6 {
		t.Errorf("%d proliferate events after an explicit doubled proliferate, want 6", n)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 5 {
		t.Errorf("Mine has %d counters, want 5 (one per proliferate that chose it)", got)
	}
}

// Scheming Aspirant: each proliferate its controller finishes drains
// each opponent for 2, a proliferate of nothing included; an opponent's
// proliferate is not "you".
func TestSchemingAspirantDrainsOnEachProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aspirant := b12Push(g, me.ID, "Scheming Aspirant", "Creature — Phyrexian Advisor", schemingAspirantOracle, 1, 3)
	myLife, oppLife := me.Life, opp.Life

	startProliferate(t, g, opp.ID, uuid.Nil)
	passPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("an opponent's proliferate triggered it: life %d/%d", me.Life, opp.Life)
	}

	startProliferate(t, g, me.ID, aspirant)
	passPriorityAroundTable(t, g)
	if me.Life != myLife+2 {
		t.Errorf("my life = %d, want %d (gain 2, not 2 per opponent)", me.Life, myLife+2)
	}
	for _, p := range g.Seats[1:] {
		if p.Life != oppLife-2 {
			t.Errorf("%s life = %d, want %d", p.Name, p.Life, oppLife-2)
		}
	}
}

// The seam's restore point: between the two proliferates of "proliferate
// twice" the first trigger is already owed and the second proliferate is
// still a prompt. Restored there, the second answer still triggers.
func TestSchemingAspirantTriggersTwiceAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushTekuthal(g, me.ID)
	b12Push(g, me.ID, "Scheming Aspirant", "Creature — Phyrexian Advisor", schemingAspirantOracle, 1, 3)
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	myLife, oppLife := me.Life, g.Seats[1].Life

	castSteadyProgressToPrompt(t, g)
	answerProliferate(t, g, mine)
	if proliferatePrompt(g) == nil {
		t.Fatal("the second proliferate asked nothing")
	}
	g = restoreRoundTrip(t, g, true)
	me = g.Seats[0]
	answerProliferate(t, g)
	passPriorityAroundTable(t, g)

	if me.Life != myLife+4 {
		t.Errorf("my life = %d, want %d (two triggers)", me.Life, myLife+4)
	}
	if got := g.Seats[1].Life; got != oppLife-4 {
		t.Errorf("an opponent's life = %d, want %d (two triggers)", got, oppLife-4)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("Mine has %d counters, want 2 (chosen once)", got)
	}
}

// castEzuri casts Ezuri and resolves it up to its enters trigger's
// payment prompt.
func castEzuri(t *testing.T, g *game.Game) {
	t.Helper()
	castCatalogSpell(t, g, "Ezuri, Stalker of Spheres", "Legendary Creature — Phyrexian Elf Warrior", ezuriStalkerOracle, nil)
	passPriorityAroundTable(t, g)
}

// Ezuri: paying {3} proliferates twice, each its own choice, and each
// one draws a card off the second ability.
func TestEzuriPaysToProliferateTwiceAndDrawsForEach(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	handBefore := len(me.Hand.Cards)

	castEzuri(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	if answerPayUnless(t, g, me.ID, true) == nil {
		t.Fatal("Ezuri did not offer to pay {3}")
	}
	answerProliferate(t, g, mine)
	if proliferatePrompt(g) == nil {
		t.Fatal("the second proliferate asked nothing")
	}
	answerProliferate(t, g, mine)
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mine, game.CounterPlusOne); got != 3 {
		t.Errorf("Mine has %d counters, want 3", got)
	}
	if got := len(me.Hand.Cards); got != handBefore+2 {
		t.Errorf("hand = %d, want %d (one card per proliferate)", got, handBefore+2)
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("%d mana left in the pool, want the {3} spent", got)
	}
}

// Declining the {3} proliferates nothing and draws nothing; a later
// proliferate from another card still draws.
func TestEzuriDeclinedDoesNothingAndStillWatches(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)
	handBefore := len(me.Hand.Cards)

	castEzuri(t, g)
	floatMana(t, g, me, "{C}{C}{C}")
	if answerPayUnless(t, g, me.ID, false) == nil {
		t.Fatal("Ezuri did not offer to pay {3}")
	}
	passPriorityAroundTable(t, g)
	if proliferatePrompt(g) != nil || len(proliferateEvents(g)) != 0 {
		t.Fatal("declining the payment proliferated")
	}
	if got := len(me.Hand.Cards); got != handBefore {
		t.Fatalf("hand = %d, want %d: no proliferate, no card", got, handBefore)
	}

	startProliferate(t, g, opp.ID, mine)
	answerProliferate(t, g)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore {
		t.Fatalf("an opponent's proliferate drew a card: hand %d", got)
	}

	startProliferate(t, g, me.ID, mine)
	answerProliferate(t, g)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d (a proliferate that chose nothing still draws)", got, handBefore+1)
	}
}

// Saying yes to the {3} with a pool that cannot pay it is a decline:
// nothing is proliferated.
func TestEzuriUnpaidDoesNotProliferate(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCounterCreature(g, me.ID, "Mine", game.CounterPlusOne, 1)

	castEzuri(t, g)
	floatMana(t, g, me, "{C}{C}")
	if answerPayUnless(t, g, me.ID, true) == nil {
		t.Fatal("Ezuri did not offer to pay {3}")
	}
	passPriorityAroundTable(t, g)
	if proliferatePrompt(g) != nil || len(proliferateEvents(g)) != 0 {
		t.Error("an unpaid trigger proliferated")
	}
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("%d mana in the pool, want the 2 that could not pay {3} untouched", got)
	}
}

// venserModePick has Venser's controller proliferate nothing and returns
// the mode prompt of the trigger.
func venserModePick(t *testing.T, g *game.Game, me uuid.UUID, venser uuid.UUID) *game.PendingChoice {
	t.Helper()
	startProliferate(t, g, me, venser)
	if c := proliferatePrompt(g); c != nil {
		answerProliferate(t, g)
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me)
	if c == nil {
		t.Fatal("Venser's trigger asked for no mode")
	}
	return c
}

// Venser's first mode makes The Hollow Sentinel as printed, and makes no
// second one while its controller has one.
func TestVenserCreatesTheHollowSentinelOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	venser := b12Push(g, me.ID, "Venser, Corpse Puppet", "Legendary Creature — Phyrexian Zombie Wizard", venserCorpsePuppetOracle, 1, 3)

	mode := venserModePick(t, g, me.ID, venser)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err == nil {
		t.Fatal("the second mode was taken with no artifact creature to target")
	}
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)

	var token *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "The Hollow Sentinel" {
			token = &g.Battlefield.Cards[i]
		}
	}
	if token == nil {
		t.Fatal("no The Hollow Sentinel was created")
	}
	if !token.IsToken() || token.Controller != me.ID || token.Power != 3 || token.Toughness != 3 ||
		token.TypeLine != "Token Legendary Artifact Creature — Phyrexian Golem" || len(token.Colors) != 0 {
		t.Errorf("the token is %+v, want a legendary 3/3 colorless Phyrexian Golem artifact creature", *token)
	}

	mode = venserModePick(t, g, me.ID, venser)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := onBattlefieldNamed(g, "The Hollow Sentinel"); n != 1 {
		t.Errorf("%d Hollow Sentinels, want 1: the first mode does nothing while you control one", n)
	}
}

// Venser's second mode: an artifact creature its controller controls
// gains flying and lifelink until end of turn; nothing else is a legal
// target.
func TestVenserGivesAnArtifactCreatureFlyingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	venser := b12Push(g, me.ID, "Venser, Corpse Puppet", "Legendary Creature — Phyrexian Zombie Wizard", venserCorpsePuppetOracle, 1, 3)
	golem := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Golem", "Artifact Creature — Golem", 3, 3)
	rock := b12Creature(g, me.ID, "Rock", "Artifact", 0, 0)

	mode := venserModePick(t, g, me.ID, venser)
	if err := g.ResolveModePick(mode.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("the second mode asked for no target")
	}
	for _, bad := range []uuid.UUID{venser, theirs, rock} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("an illegal target was accepted: %v", bad)
		}
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: golem}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)

	for _, kw := range []string{"flying", "lifelink"} {
		if !hasKeywordOnBattlefield(t, g, golem, kw) {
			t.Errorf("the target did not gain %s", kw)
		}
	}
	if hasKeywordOnBattlefield(t, g, theirs, "flying") {
		t.Error("an untargeted creature gained flying")
	}
	if n := onBattlefieldNamed(g, "The Hollow Sentinel"); n != 0 {
		t.Error("the second mode made a token")
	}
}

// Venser prints lifelink and toxic 1.
func TestVenserPrintsLifelinkAndToxic(t *testing.T) {
	g := newCatalogGame(t)
	venser := b12Push(g, g.Seats[0].ID, "Venser, Corpse Puppet", "Legendary Creature — Phyrexian Zombie Wizard", venserCorpsePuppetOracle, 1, 3)
	for _, kw := range []string{"lifelink", "toxic 1"} {
		if !hasKeywordOnBattlefield(t, g, venser, kw) {
			t.Errorf("Venser has no %s", kw)
		}
	}
}

// pushVoidwingInGraveyard puts a Voidwing Hybrid in `owner`'s graveyard.
func pushVoidwingInGraveyard(owner *game.Player) uuid.UUID {
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Voidwing Hybrid", TypeLine: "Creature — Phyrexian Bat", OracleID: voidwingHybridOracle,
		Power: 2, Toughness: 1, Owner: owner.ID, Controller: owner.ID,
	})
	return id
}

// Voidwing Hybrid comes back from its owner's graveyard when they
// proliferate, and only then: not on an opponent's proliferate.
func TestVoidwingHybridReturnsFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hybrid := pushVoidwingInGraveyard(me)

	startProliferate(t, g, opp.ID, uuid.Nil)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hybrid) {
		t.Fatal("an opponent's proliferate returned it")
	}

	startProliferate(t, g, me.ID, uuid.Nil)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hybrid) || me.Graveyard.Contains(hybrid) {
		t.Error("the Hybrid did not return to its owner's hand")
	}
}

// The ability works only from the graveyard, and it returns the object
// that triggered: one that left and came back in response stays put.
func TestVoidwingHybridOnlyFromTheGraveyardAndOnlyThatObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	onField := b12Push(g, me.ID, "Voidwing Hybrid", "Creature — Phyrexian Bat", voidwingHybridOracle, 2, 1)
	for _, kw := range []string{"flying", "toxic 1"} {
		if !hasKeywordOnBattlefield(t, g, onField, kw) {
			t.Errorf("Voidwing Hybrid has no %s", kw)
		}
	}
	hybrid := pushVoidwingInGraveyard(me)

	startProliferate(t, g, me.ID, uuid.Nil)
	yard, exile := game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneExile}
	if err := g.MoveCardByID(yard, exile, hybrid); err != nil {
		t.Fatalf("MoveCardByID to exile: %v", err)
	}
	if err := g.MoveCardByID(exile, yard, hybrid); err != nil {
		t.Fatalf("MoveCardByID back: %v", err)
	}
	passPriorityAroundTable(t, g)

	if findBattlefieldCardByID(g, onField) == nil {
		t.Error("the Hybrid on the battlefield moved: the ability works only from the graveyard")
	}
	if me.Hand.Contains(hybrid) || !me.Graveyard.Contains(hybrid) {
		t.Error("a Hybrid that left the graveyard and came back in response was returned: it is a new object (CR 400.7)")
	}
}
