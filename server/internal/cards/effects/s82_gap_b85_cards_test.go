package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB85RassilonOracle     = "ebec9957-b777-42ae-a394-85c7ed84d00a"
	gapB85CyberControllerOr  = "06c25d71-f3f4-4089-b9bc-de96fbad73a5"
	gapB85MasterNewOracle    = "e7828284-543d-4f8b-9c5e-a8f082010398"
	gapB85MasterMesmerOracle = "306b6b9b-f488-4ea8-9f5e-9e8e7354df64"
	gapB85AceOracle          = "50d0d870-fe3d-44a2-ad6e-93307a1fb468"
	gapB85DinosaursOracle    = "bc0daba2-c51c-44b0-a4a3-806f2acd9aa6"
	gapB85GrahamOracle       = "7986908d-8af2-465a-8597-26e92ca2f1c9"
	gapB85JamieOracle        = "c9f074cb-6dcb-4c25-8332-9b15e4881d33"
	gapB85KarvanistaOracle   = "6c370299-9803-4cb5-97ba-c17831797b54"
	gapB85LeelaOracle        = "4ae972ea-d5d8-4064-9d91-5b2040d7dcbf"
	gapB85SisterhoodOracle   = "b32484ee-c7d8-472d-818f-2c37783d028a"
)

func TestGapB85EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB85RassilonOracle:          "Rassilon, the War President",
		gapB85CyberControllerOr:       "The Cyber-Controller",
		gapB85MasterNewOracle:         "The Master, Formed Anew",
		gapB85MasterMesmerOracle:      "The Master, Mesmerist",
		gapB85AceOracle:               "Ace, Fearless Rebel",
		gapB85DinosaursOracle:         "Displaced Dinosaurs",
		gapB85GrahamOracle:            "Graham O'Brien",
		gapB85JamieOracle:             "Jamie McCrimmon",
		gapB85KarvanistaOracle:        "Karvanista, Loyal Lupari",
		gapB85LeelaOracle:             "Leela, Sevateem Warrior",
		gapB85SisterhoodOracle:        "Sisterhood of Karn",
		gapB85KarvanistaOracle + "#1": "Lupari Shield",
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

// gapB85CastFromExile puts a spell card with the given type and cost into
// exile with a permission for p and casts it from there.
func gapB85CastFromExile(t *testing.T, g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: id, Name: name, TypeLine: typeLine,
			Owner: p.ID, Controller: p.ID})
		_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: p.ID, Duration: g.UntilEndOfTurnDuration()})
	})
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("cast %s from exile: %v", name, err)
	}
	return id
}

func gapB85Foods(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.HasSubtype("Food") {
			n++
		}
	}
	return n
}

func gapB85Effective(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var c game.Card
	var ok bool
	g.ReadSnapshot(func() { c, ok = g.LookupCardForEffect(id) })
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c
}

func TestGapB85GrahamMakesAFoodForACastFromOutsideTheHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Graham O'Brien", "Legendary Creature — Human Pilot", gapB85GrahamOracle, false)
	// A cast from the hand makes nothing.
	castCatalogSpell(t, g, "Hand Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := gapB85Foods(g, me.ID); n != 0 {
		t.Fatalf("a cast from hand made %d Foods", n)
	}
	// A cast from exile makes one.
	gapB85CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	if n := gapB85Foods(g, me.ID); n != 1 {
		t.Fatalf("a cast from exile made %d Foods, want 1", n)
	}
	// An opponent's cast from exile makes none for Graham's controller.
	gapB85CastFromExile(t, g, g.Seats[1], "Their Spell", "Instant")
	passPriorityAroundTable(t, g)
	if n := gapB85Foods(g, me.ID); n != 1 {
		t.Fatalf("an opponent's cast made a Food: %d", n)
	}
}

func TestGapB85SisterhoodEntersWithACounterAndDoublesThemOnParadox(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Sisterhood of Karn", "Creature — Cleric", gapB85SisterhoodOracle, nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
		t.Fatalf("entered with %d +1/+1 counters, want 1", n)
	}
	gapB85CastFromExile(t, g, me, "Exile Spell", "Instant")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterPlusOne); n != 2 {
		t.Fatalf("after one Paradox cast: %d counters, want 2", n)
	}
	gapB85CastFromExile(t, g, me, "Exile Spell Two", "Instant")
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterPlusOne); n != 4 {
		t.Fatalf("after two Paradox casts: %d counters, want 4", n)
	}
	castCatalogSpell(t, g, "Hand Spell", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, id, game.CounterPlusOne); n != 4 {
		t.Fatalf("a cast from hand changed the counters: %d", n)
	}
}

func TestGapB85LeelaGrowsOnEveryExtraOpponentDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leela := pushCatalogPermanent(g, me.ID, "Leela, Sevateem Warrior", "Legendary Creature — Human Warrior", gapB85LeelaOracle, false)
	advanceToMain(t, g)
	// Outside the opponent's own draw step, every draw counts.
	if err := g.DrawCard(opp.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, leela, game.CounterPlusOne); n != 1 {
		t.Fatalf("one opponent draw: %d counters, want 1", n)
	}
	// Our own draws never count.
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countersOn(g, leela, game.CounterPlusOne); n != 1 {
		t.Fatalf("our own draw: %d counters, want 1", n)
	}
}

func TestGapB85JamiePumpsByAHistoricSpellsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	jamie := b21Push(g, me.ID, "Jamie McCrimmon", "Legendary Creature — Human Warrior", gapB85JamieOracle, 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { fillPool(me, 5) })
	rock := handCardOf(g, me, "Big Rock", "Artifact", "{3}", "")
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the artifact: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := gapB85Effective(t, g, jamie).CurrentPower(); p != 5 {
		t.Fatalf("after a {3} artifact: power %d, want 5", p)
	}
	// A nonhistoric spell does nothing.
	bolt := handCardOf(g, me, "Plain Spell", "Instant", "{2}", "")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p := gapB85Effective(t, g, jamie).CurrentPower(); p != 5 {
		t.Fatalf("after a nonhistoric spell: power %d, want 5", p)
	}
}

func gapB85Human(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return b16Creature(g, owner, name, "Creature — Human", 1, 1)
}

func TestGapB85KarvanistaPutsACounterOnEachHumanWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	karv := b21Push(g, me.ID, "Karvanista, Loyal Lupari", "Legendary Creature — Alien Dog Soldier", gapB85KarvanistaOracle, 5, 5)
	h1, h2 := gapB85Human(g, me.ID, "Human One"), gapB85Human(g, me.ID, "Human Two")
	theirs := gapB85Human(g, opp.ID, "Their Human")
	declareAttack(t, g, opp.ID, karv)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{h1, h2} {
		if n := countersOn(g, id, game.CounterPlusOne); n != 1 {
			t.Errorf("our Human has %d counters, want 1", n)
		}
	}
	if n := countersOn(g, theirs, game.CounterPlusOne); n != 0 {
		t.Errorf("their Human has %d counters", n)
	}
	if n := countersOn(g, karv, game.CounterPlusOne); n != 0 {
		t.Errorf("Karvanista is not a Human but has %d counters", n)
	}
}

func TestGapB85LupariShieldGivesOnlyTheHumansYouHaveNowIndestructible(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := gapB85Human(g, me.ID, "Human One")
	bear := b16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := gapB85Human(g, opp.ID, "Their Human")
	castCatalogSpell(t, g, "Lupari Shield", "Sorcery — Adventure", gapB85KarvanistaOracle+"#1", nil)
	passPriorityAroundTable(t, g)
	late := gapB85Human(g, me.ID, "Late Human")
	for id, want := range map[uuid.UUID]bool{h: true, bear: false, theirs: false, late: false} {
		c := gapB85Effective(t, g, id)
		if got := HasKeyword("indestructible")(g, me.ID, c); got != want {
			t.Errorf("%s indestructible = %v, want %v", c.Name, got, want)
		}
	}
}

func TestGapB85DisplacedDinosaursMakesEachHistoricPermanentASevenSevenDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Displaced Dinosaurs", "Creature — Dinosaur", gapB85DinosaursOracle, false)
	rock := castCatalogSpell(t, g, "Plain Rock", "Artifact", "", nil)
	passPriorityAroundTable(t, g)
	c := gapB85Effective(t, g, rock)
	if !c.IsCreature() || !c.HasSubtype("Dinosaur") || !c.HasCardType("artifact") || c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Fatalf("an artifact entered as %q %d/%d, want a 7/7 Dinosaur artifact creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	legend := castCatalogSpell(t, g, "Legend", "Legendary Creature — Elf", "", nil)
	passPriorityAroundTable(t, g)
	if c := gapB85Effective(t, g, legend); c.CurrentPower() != 7 || !c.HasSubtype("Dinosaur") || !c.HasSubtype("Elf") {
		t.Fatalf("a legendary creature entered as %q %d/%d", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	// A nonhistoric permanent is left alone, and so is an opponent's artifact.
	plain := castCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if c := gapB85Effective(t, g, plain); c.HasSubtype("Dinosaur") {
		t.Errorf("a nonhistoric creature became a Dinosaur")
	}
	theirs := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if c := gapB85Effective(t, g, theirs); c.IsCreature() {
		t.Errorf("an opponent's artifact became a creature")
	}
}

func TestGapB85AceSacrificesAnArtifactToGrowAndFight(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ace := b21Push(g, me.ID, "Ace, Fearless Rebel", "Legendary Creature — Human Rebel", gapB85AceOracle, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 3)
	declareAttack(t, g, opp.ID, ace)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, rock)
	b04WaitForPick(t, g, me.ID)
	if g.Battlefield.Contains(rock) {
		t.Fatal("the artifact is sacrificed before the reflexive trigger asks for a target")
	}
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, ace, game.CounterPlusOne); n != 1 {
		t.Errorf("Ace has %d counters, want 1", n)
	}
	if g.Battlefield.Contains(victim) {
		t.Error("a 3/3 Ace fights and kills a 1/3")
	}
}

func TestGapB85AceMayDeclineTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ace := b21Push(g, me.ID, "Ace, Fearless Rebel", "Legendary Creature — Human Rebel", gapB85AceOracle, 2, 2)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, ace)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) || !g.Battlefield.Contains(victim) {
		t.Fatal("declining sacrificed or fought")
	}
	if n := countersOn(g, ace, game.CounterPlusOne); n != 0 {
		t.Errorf("Ace has %d counters after declining", n)
	}
}

func TestGapB85MasterMesmeristGrantsSkulkAndGoads(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b21Push(g, me.ID, "The Master, Mesmerist", "Legendary Creature — Time Lord Rogue", gapB85MasterMesmerOracle, 3, 3)
	small := b12Creature(g, opp.ID, "Small", "Creature — Rat", 3, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, master, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(big)}); err == nil {
		t.Fatal("targeted a creature with greater power")
	}
	b16Activate(t, g, me.ID, master, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(small)})
	passPriorityAroundTable(t, g)
	c := gapB85Effective(t, g, small)
	if !HasKeyword("skulk")(g, me.ID, c) {
		t.Error("the target did not gain skulk")
	}
	if !c.IsGoadedBy(me.ID) {
		t.Error("the target was not goaded")
	}
}

func TestGapB85MasterMesmeristGrowsAndDrawsWhenASkulkCreatureConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := b21Push(g, me.ID, "The Master, Mesmerist", "Legendary Creature — Time Lord Rogue", gapB85MasterMesmerOracle, 3, 3)
	sneak := apaPush(g, me.ID, me.ID, game.Card{Name: "Sneak", TypeLine: "Creature — Rogue", Power: 2, Toughness: 2, Keywords: []string{"skulk"}})
	bystander := b12Creature(g, me.ID, "Bystander", "Creature — Rogue", 2, 2)
	hand := me.Hand.Size()
	life := opp.Life
	declareAttack(t, g, opp.ID, sneak, bystander)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("opponent life %d → %d, want 4 combat damage", life, opp.Life)
	}
	if n := countersOn(g, master, game.CounterPlusOne); n != 1 {
		t.Errorf("The Master has %d counters, want 1 (only the skulk creature counts)", n)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// --- The Cyber-Controller -------------------------------------------------

func TestGapB85CyberControllerMillsXAndTurnsCreaturesIntoCybermen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The top of the opponent's library: a creature, a land, a creature.
	seedLibraryTop(opp, "Land", "Basic Land — Forest", "")
	seedLibraryTop(opp, "Beast", "Creature — Beast", "{2}")
	seedLibraryTop(opp, "Ogre", "Creature — Ogre", "{3}")
	paCast(t, g, "The Cyber-Controller", "Legendary Artifact Creature — Cyberman", "{X}{U}{U}{B}", gapB85CyberControllerOr,
		game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)
	var cybermen int
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || !c.FaceDown {
			continue
		}
		e := gapB85Effective(t, g, c.InstanceID)
		if e.IsCreature() && e.HasCardType("artifact") && e.HasSubtype("Cyberman") {
			cybermen++
		}
	}
	if cybermen != 2 {
		t.Fatalf("%d face-down Cyberman artifact creatures, want 2", cybermen)
	}
	if got := opp.Graveyard.Size(); got != 1 {
		t.Errorf("the opponent's graveyard holds %d cards, want only the milled land", got)
	}
	// The Controller itself is not an "other" artifact creature; the Cybermen are
	// 2/2 plus the anthem.
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.FaceDown {
			if e := gapB85Effective(t, g, c.InstanceID); e.CurrentPower() != 3 {
				t.Errorf("a Cyberman is %d/%d, want 3/3 under the anthem", e.CurrentPower(), e.CurrentToughness())
			}
		}
	}
}

// --- Rassilon, the War President ---------------------------------------------

func TestGapB85RassilonUpkeepLosesLifeAndExilesAPlayableCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Rassilon, the War President", "Legendary Creature — Time Lord Noble", gapB85RassilonOracle, false)
	top := seedLibraryTop(me, "Top Card", "Sorcery", "")
	life := me.Life
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want a loss of 2", life, me.Life)
	}
	inExile := false
	for _, c := range g.Exile.Cards {
		if c.InstanceID == top {
			inExile = true
		}
	}
	if !inExile {
		t.Fatal("the top card was not exiled")
	}
}

func TestGapB85RassilonGrantsConspireToNoncreatureSpellsCastFromExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Rassilon, the War President", "Legendary Creature — Time Lord Noble", gapB85RassilonOracle, false)
	a := pr7Creature(g, me.ID, "Rat A", 1, "U")
	b := pr7Creature(g, me.ID, "Rat B", 1, "U")
	cast := func(from string, typeLine string, conspire bool) (uuid.UUID, error) {
		id := uuid.New()
		card := game.Card{InstanceID: id, Name: "Blue Spell", TypeLine: typeLine, Colors: []string{"U"}, Owner: me.ID, Controller: me.ID}
		params := game.CastSpellParams{Strict: true}
		if from == "exile" {
			g.WithWriteLock(func() {
				g.Exile.Cards = append(g.Exile.Cards, card)
				_ = g.ExileCardWithPermissionForEffect(id, game.CastPermission{Player: me.ID, Duration: g.UntilEndOfTurnDuration()})
			})
			params.FromZone = "exile"
		} else {
			g.WithWriteLock(func() { me.Hand.PushTop(card) })
		}
		if conspire {
			params.OptionalCosts = []int{0}
			params.ConspireIDs = []uuid.UUID{a, b}
		}
		return id, g.CastSpell(me.ID, id, params)
	}
	if _, err := cast("hand", "Instant", true); err == nil {
		t.Fatal("a spell cast from the hand was offered conspire")
	}
	if _, err := cast("exile", "Creature — Test", true); err == nil {
		t.Fatal("a creature spell cast from exile was offered conspire")
	}
	spell, err := cast("exile", "Instant", true)
	if err != nil {
		t.Fatalf("conspire on a noncreature spell cast from exile: %v", err)
	}
	if !tappedForTest(t, g, a) || !tappedForTest(t, g, b) {
		t.Error("the two creatures were not tapped")
	}
	if n := rsConspireTriggers(g, spell); n != 1 {
		t.Errorf("%d conspire triggers, want 1", n)
	}
}

// --- The Master, Formed Anew ---------------------------------------------------

func TestGapB85MasterFormedAnewTakesOverAnExiledCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := b21Push(g, me.ID, "Stolen Ogre", "Creature — Ogre", "", 4, 4)
	// A creature card in exile with no takeover counter is not a candidate.
	plainExiled := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards, game.Card{InstanceID: plainExiled, Name: "Plain Elf", TypeLine: "Creature — Elf",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	advanceToMain(t, g)
	master := paCast(t, g, "The Master, Formed Anew", "Legendary Creature — Time Lord Rogue", "{U}{B}", gapB85MasterNewOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the creature was not exiled")
	}
	var tagged bool
	for _, c := range g.Exile.Cards {
		if c.InstanceID == victim && c.Counters[gapB85TakeoverCounter] == 1 {
			tagged = true
		}
	}
	if !tagged {
		t.Fatal("the exiled creature has no takeover counter")
	}
	prompt := waitForCopyPrompt(t, g)
	if len(prompt.CopyOptions) != 1 || prompt.CopyOptions[0] != victim {
		t.Fatalf("copy options %v, want only the creature with the takeover counter", prompt.CopyOptions)
	}
	resolveWithCopyChoice(t, g, victim)
	got := copyBattlefieldCard(t, g, master)
	if got.Name != "Stolen Ogre" || got.CurrentPower() != 4 || got.CurrentToughness() != 4 {
		t.Errorf("The Master entered as %q %d/%d, want a 4/4 Stolen Ogre", got.Name, got.CurrentPower(), got.CurrentToughness())
	}
}

func TestGapB85MasterFormedAnewMayDeclineTheCopyAndTheExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := b21Push(g, me.ID, "Ogre", "Creature — Ogre", "", 4, 4)
	advanceToMain(t, g)
	master := paCast(t, g, "The Master, Formed Anew", "Legendary Creature — Time Lord Rogue", "{U}{B}", gapB85MasterNewOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("declining the exile exiled the creature")
	}
	if p := copyPrompt(g); p != nil {
		t.Fatalf("with nothing to copy a prompt opened: %v", p.CopyOptions)
	}
	if got := copyBattlefieldCard(t, g, master); got.Name != "The Master, Formed Anew" {
		t.Errorf("the Master entered as %q", got.Name)
	}
}
