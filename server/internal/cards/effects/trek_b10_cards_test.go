package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b10_cards_test.go — Star Trek, batch 10 (#4973). Each test plays the
// real catalog entry through the engine and asserts what its text says.

const (
	trekB10ArmusOracle        = "42c77c4d-90bb-45d6-a9e0-36704b5a2b11"
	trekB10AssaultDroneOracle = "18832d7a-9a1f-45f1-a0f9-8e768a36af07"
	trekB10WarfareOracle      = "b28b4fe9-1a35-48c1-8818-f6f786b04e3e"
	trekB10BatlethOracle      = "1000d26e-7a29-48df-806f-9eb7d4e1082d"
	trekB10SurvivalistOracle  = "0e0e15d9-bd29-495a-9975-d1bd2ebb4bff"
	trekB10BlastoneuronOracle = "87bcc935-e7e7-4e98-a23c-01c5a9b7cfcd"
	trekB10ChadichOracle      = "cfdef803-6f7a-4055-8d51-8e8c60953928"
	trekB10TrimatrixOracle    = "4042ab14-c579-4d82-a642-576b88508ec7"
	trekB10CollectiveOracle   = "d9806ba7-8af4-4e8e-9238-7bc40c6bee0b"
	trekB10RaffiOracle        = "4f8a70cb-2487-4e6d-9a3f-c6c5313b8e82"
)

func TestTrekB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB10ArmusOracle:        "Armus, Skin of Evil",
		trekB10AssaultDroneOracle: "Assault Drone",
		trekB10WarfareOracle:      "Automated Warfare System",
		trekB10BatlethOracle:      "Bat'leth",
		trekB10SurvivalistOracle:  "Battle-Scarred Survivalist",
		trekB10BlastoneuronOracle: "Blastoneuron",
		trekB10ChadichOracle:      "Cha'DIch Investigator",
		trekB10TrimatrixOracle:    "Collation Trimatrix",
		trekB10CollectiveOracle:   "Collective Drone",
		trekB10RaffiOracle:        "Commander Raffi Musiker",
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

// trekB10Cast puts a catalog card with its printed P/T in seat 0's hand,
// floats its cost and casts it strictly in the precombat main phase.
func trekB10Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID,
		})
	})
	floatMana(t, g, me, trekB10Generic(cost))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

// trekB10Generic turns a printed cost into mana that pays it: each generic
// pip becomes {C}, coloured pips stay.
func trekB10Generic(cost string) string {
	out := ""
	for i := 0; i < len(cost); {
		j := i + 1
		for j < len(cost) && cost[j] != '}' {
			j++
		}
		sym := cost[i+1 : j]
		n := 0
		for _, r := range sym {
			if r < '0' || r > '9' {
				n = -1
				break
			}
			n = n*10 + int(r-'0')
		}
		if n < 0 {
			out += "{" + sym + "}"
		} else {
			for k := 0; k < n; k++ {
				out += "{C}"
			}
		}
		i = j + 1
	}
	return out
}

// trekB10ToHand puts n plain cards in a player's hand.
func trekB10ToHand(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		}
	})
}

// trekB10Tokens is the tokens named `name` that `controller` controls,
// through the layered view.
func trekB10Tokens(t *testing.T, g *game.Game, controller uuid.UUID, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			if v := layeredCard(t, g, c.InstanceID); v.IsToken() {
				out = append(out, v)
			}
		}
	}
	return out
}

func trekB10Artifact(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return b12Creature(g, owner, name, "Artifact Creature — Borg", power, toughness)
}

// --- Armus, Skin of Evil -----------------------------------------------------

func TestTrekB10ArmusMakesTargetOpponentDiscardTwoAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	armus := b27Push(g, me.ID, "Armus, Skin of Evil", "Legendary Creature — Ooze Horror", trekB10ArmusOracle, "{5}{B}{B}", 6, 6, "B")
	if c := layeredCard(t, g, armus); !game.HasKeyword(&c, "menace") {
		t.Error("Armus has no menace")
	}
	trekB10ToHand(g, opp, 3)
	trekB10ToHand(g, other, 3)
	oppHand, otherHand := opp.Hand.Size(), other.Hand.Size()

	advanceTo(t, g, game.StepEnd)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, p := range pick.PickTargetPlayers {
			if p == me.ID {
				t.Error("I was offered as the target opponent")
			}
		}
	}
	answerPickTargetPlayer(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	c := discardChoiceFor(g, opp.ID)
	if c == nil || c.ChooseMax != 2 {
		t.Fatalf("the target opponent was not asked to discard two: %+v", c)
	}
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != oppHand-2 {
		t.Errorf("target opponent's hand %d, want %d", got, oppHand-2)
	}
	if got := other.Hand.Size(); got != otherHand {
		t.Errorf("another opponent's hand %d, want %d", got, otherHand)
	}
}

func TestTrekB10ArmusDoesNotTriggerAtAnOpponentsEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	armus := b27Push(g, me.ID, "Armus, Skin of Evil", "Legendary Creature — Ooze Horror", trekB10ArmusOracle, "{5}{B}{B}", 6, 6, "B")
	powerUpPassTo(t, g, 1)
	advanceTo(t, g, game.StepEnd)
	if n := triggersOnStackFrom(g, armus); n != 0 || latestPickTarget(g, me.ID) != nil {
		t.Errorf("Armus triggered at an opponent's end step (triggers %d)", n)
	}
}

func TestTrekB10ArmusDiscardsFromHandToGiveMinusFourMinusFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	armus := handCardFull(me, "Armus, Skin of Evil", "Legendary Creature — Ooze Horror", "{5}{B}{B}", trekB10ArmusOracle, []string{"B"})
	big := b12Creature(g, g.Seats[1].ID, "Their Giant", "Creature — Giant", 5, 5)

	floatMana(t, g, me, "{B}{C}")
	if err := g.ActivateCatalogAbility(me.ID, armus, 0, game.ActivateAbilityParams{Targets: cardRefs(big), Strict: true}); err == nil {
		t.Fatal("activated with two mana")
	}
	if !me.Hand.Contains(armus) {
		t.Fatal("a refused activation discarded Armus")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, armus, 0, game.ActivateAbilityParams{Targets: cardRefs(big), Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(armus) {
		t.Error("the discard was not the cost")
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, big)
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the target is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Assault Drone ---------------------------------------------------------------

func TestTrekB10AssaultDronePumpsYourArtifactCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drone := b27Push(g, me.ID, "Assault Drone", "Artifact Creature — Borg", trekB10AssaultDroneOracle, "{3}{B}", 2, 4, "B")
	mine := trekB10Artifact(g, me.ID, "My Drone", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := trekB10Artifact(g, opp.ID, "Their Drone", 1, 1)
	for id, want := range map[uuid.UUID]int{drone: 3, mine: 2, bear: 2, theirs: 1} {
		c := layeredCard(t, g, id)
		if c.CurrentPower() != want {
			t.Errorf("%s has power %d, want %d", c.Name, c.CurrentPower(), want)
		}
	}
	if c := layeredCard(t, g, drone); c.CurrentToughness() != 4 {
		t.Errorf("Assault Drone toughness %d, want 4", c.CurrentToughness())
	}
}

// --- Automated Warfare System -----------------------------------------------------

func TestTrekB10WarfareSystemSacrificesAnotherArtifactToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aws := b27Push(g, me.ID, "Automated Warfare System", "Artifact Creature — Construct", trekB10WarfareOracle, "{2}{B}", 2, 2, "B")
	if c := layeredCard(t, g, aws); !game.HasKeyword(&c, "flying") {
		t.Error("no flying")
	}
	relic := b12Permanent(g, me.ID, "Relic", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")

	declareAttack(t, g, opp.ID, aws)
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if prompt == nil || prompt.ChooseMin != 0 || !hasID(prompt.ChooseCards, relic) || !hasID(prompt.ChooseCards, bear) ||
		hasID(prompt.ChooseCards, aws) || hasID(prompt.ChooseCards, land) {
		t.Fatalf("want a may-sacrifice over another creature or artifact: %+v", prompt)
	}
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID, relic)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not sacrificed")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestTrekB10WarfareSystemDecliningDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aws := b27Push(g, me.ID, "Automated Warfare System", "Artifact Creature — Construct", trekB10WarfareOracle, "{2}{B}", 2, 2, "B")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, aws)
	passPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("declining still sacrificed")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d", got, hand)
	}
}

// --- Bat'leth -------------------------------------------------------------------------

func TestTrekB10BatlethShrinksAnOpposingCreatureAndEquipsForTwoPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	batleth := trekB10Cast(t, g, "Bat'leth", "Artifact — Equipment", "{2}{B}", trekB10BatlethOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if hasID(pick.PickTargetCards, mine) {
			t.Error("a creature I control was offered")
		}
	}
	answerPickTargetIfOpen(t, g, big)
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, big); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("the target is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("an untargeted creature changed: power %d", c.CurrentPower())
	}

	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, batleth, 0, game.ActivateAbilityParams{Targets: cardRefs(mine), Strict: true}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, mine); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("the equipped creature is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Battle-Scarred Survivalist ----------------------------------------------

func TestTrekB10SurvivalistGrowsOnlyOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	surv := b27Push(g, me.ID, "Battle-Scarred Survivalist", "Creature — Klingon Scout", trekB10SurvivalistOracle, "{1}{B}", 1, 1, "B")
	if c := layeredCard(t, g, surv); !game.HasKeyword(&c, "menace") {
		t.Error("no menace")
	}
	a := b12Creature(g, opp.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	c := b12Creature(g, me.ID, "Bear C", "Creature — Bear", 2, 2)
	d := b12Creature(g, opp.ID, "Bear D", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	// Two at once: one trigger.
	b27Kill(g, a)
	b27Kill(g, b)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := counterOn(g, surv, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after two deaths at once, want 1", got)
	}
	// Another later this turn: none.
	b27Kill(g, c)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := counterOn(g, surv, game.CounterPlusOne); got != 1 {
		t.Fatalf("%d counters after a second death this turn, want 1", got)
	}
	// Next turn it triggers again.
	powerUpPassTo(t, g, 1)
	b27Kill(g, d)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := counterOn(g, surv, game.CounterPlusOne); got != 2 {
		t.Errorf("%d counters after a death next turn, want 2", got)
	}
}

// --- Blastoneuron -------------------------------------------------------------

func TestTrekB10BlastoneuronCopiesASmallCreatureAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	small := pushToGraveyard(g, me, graveBear())
	big := pushToGraveyard(g, me, game.Card{Name: "Grave Giant", TypeLine: "Creature — Giant", ManaCost: "{3}{G}", Power: 4, Toughness: 4})
	theirs := pushToGraveyard(g, g.Seats[1], graveBear())
	lives := []int{}
	for _, s := range g.Seats {
		lives = append(lives, s.Life)
	}

	trekB10Cast(t, g, "Blastoneuron", "Creature — Jellyfish Leech", "{3}{B}", trekB10BlastoneuronOracle, 2, 2)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enter trigger")
	}
	if !hasID(pick.PickTargetCards, small) || hasID(pick.PickTargetCards, big) || hasID(pick.PickTargetCards, theirs) {
		t.Fatalf("offered %v: want only the mana value 2 creature card in my graveyard", pick.PickTargetCards)
	}
	answerPickTarget(t, g, small)
	passPriorityAroundTable(t, g)

	if z := g.FindCardZoneForEffect(small); z == nil || z.Kind != game.ZoneExile {
		t.Fatal("the card was not exiled")
	}
	copies := trekB10Tokens(t, g, me.ID, "Grave Bear")
	if len(copies) != 1 {
		t.Fatalf("%d Grave Bear tokens, want 1", len(copies))
	}
	tok := copies[0]
	if !game.HasKeyword(&tok, "haste") || tok.CurrentPower() != 2 {
		t.Errorf("the copy: haste=%v power=%d, want a hasty 2/2", game.HasKeyword(&tok, "haste"), tok.CurrentPower())
	}
	// The exile is a card leaving my graveyard: each opponent lost 1, I gained 1.
	for i, s := range g.Seats {
		want := lives[i] - 1
		if i == 0 {
			want = lives[i] + 1
		}
		if s.Life != want {
			t.Errorf("seat %d life %d, want %d", i, s.Life, want)
		}
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(trekB10Tokens(t, g, me.ID, "Grave Bear")); n != 0 {
		t.Errorf("%d copies left after the end step, want 0", n)
	}
}

func TestTrekB10BlastoneuronDrainsOnceForSeveralCardsLeavingTogether(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Blastoneuron", "Creature — Jellyfish Leech", trekB10BlastoneuronOracle, "{3}{B}", 2, 2, "B")
	x := pushToGraveyard(g, me, graveBear())
	y := pushToGraveyard(g, me, graveBear())
	advanceToMain(t, g)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(x)
		_ = g.ExileCardForEffect(y)
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("life %d / opponent %d, want %d / %d", me.Life, opp.Life, life+1, oppLife-1)
	}
	// An opponent's graveyard does not count.
	z := pushToGraveyard(g, opp, graveBear())
	g.WithWriteLock(func() {
		_ = g.ExileCardForEffect(z)
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("an opponent's card leaving their graveyard drained: life %d", me.Life)
	}
}

// --- Cha'DIch Investigator ----------------------------------------------------

func TestTrekB10ChadichMakesAHastyKlingonWarrior(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB10Cast(t, g, "Cha'DIch Investigator", "Creature — Klingon Warrior", "{4}{B}", trekB10ChadichOracle, 4, 4)
	passPriorityAroundTable(t, g)
	toks := trekB10Tokens(t, g, me.ID, "Klingon Warrior")
	if len(toks) != 1 {
		t.Fatalf("%d Klingon Warrior tokens, want 1", len(toks))
	}
	tok := toks[0]
	cols := tok.EffectiveColors()
	if tok.CurrentPower() != 2 || tok.CurrentToughness() != 1 || len(cols) != 1 || cols[0] != "R" ||
		!game.HasKeyword(&tok, "haste") || !tok.IsCreature() {
		t.Errorf("token %d/%d %v haste=%v, want a 2/1 red creature with haste",
			tok.CurrentPower(), tok.CurrentToughness(), cols, game.HasKeyword(&tok, "haste"))
	}
}

// --- Collation Trimatrix --------------------------------------------------------

func TestTrekB10TrimatrixTapsThreeArtifactCreaturesToDrawAndLoseOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tri := b27Push(g, me.ID, "Collation Trimatrix", "Artifact Creature — Borg", trekB10TrimatrixOracle, "{B}", 1, 1, "B")
	a := trekB10Artifact(g, me.ID, "Drone A", 1, 1)
	b := trekB10Artifact(g, me.ID, "Drone B", 1, 1)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{tri, a, bear}, Strict: true}); err == nil {
		t.Fatal("a non-artifact creature paid the cost")
	}
	hand, life := me.Hand.Size(), me.Life
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{tri, a, b}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d life %d, want %d and %d", me.Hand.Size(), me.Life, hand+1, life-1)
	}
	for _, id := range []uuid.UUID{tri, a, b} {
		if !layeredCard(t, g, id).Tapped {
			t.Errorf("%s is untapped", id)
		}
	}
	// All three are tapped now: it can't be activated again.
	if err := g.ActivateCatalogAbility(me.ID, tri, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{tri, a, b}, Strict: true}); err == nil {
		t.Error("activated again with tapped artifact creatures")
	}
}

// --- Collective Drone --------------------------------------------------------------

func TestTrekB10CollectiveDroneCountsYourArtifactCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	drone := b27Push(g, me.ID, "Collective Drone", "Artifact Creature — Borg", trekB10CollectiveOracle, "{1}{B}", 0, 2, "B")
	if c := layeredCard(t, g, drone); c.CurrentPower() != 1 || c.CurrentToughness() != 2 {
		t.Fatalf("alone it is %d/%d, want 1/2", c.CurrentPower(), c.CurrentToughness())
	}
	trekB10Artifact(g, me.ID, "Drone A", 1, 1)
	trekB10Artifact(g, me.ID, "Drone B", 1, 1)
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "Relic", "Artifact")
	trekB10Artifact(g, opp.ID, "Their Drone", 1, 1)
	if c := layeredCard(t, g, drone); c.CurrentPower() != 3 {
		t.Errorf("power %d, want 3 (itself and two other artifact creatures)", c.CurrentPower())
	}
}

// --- Commander Raffi Musiker ---------------------------------------------------

func trekB10Raffi(t *testing.T, g *game.Game, mode int) {
	t.Helper()
	me := g.Seats[0]
	trekB10Cast(t, g, "Commander Raffi Musiker", "Legendary Creature — Human Officer", "{2}{B}", trekB10RaffiOracle, 3, 2)
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				break
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestTrekB10RaffiExilesACardFromAGraveyardAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := pushToGraveyard(g, opp, graveBear())
	life := me.Life
	trekB10Raffi(t, g, 0)
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Error("the graveyard card was not exiled")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestTrekB10RaffiInvestigatesAndLosesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	trekB10Raffi(t, g, 1)
	passPriorityAroundTable(t, g)
	clues := trekB10Tokens(t, g, me.ID, "Clue")
	if len(clues) != 1 || !clues[0].IsArtifact() {
		t.Fatalf("%d Clue tokens, want 1 artifact", len(clues))
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}
