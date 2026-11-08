package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lci_dl_b2_cards_test.go — The Lost Caverns of Ixalan, cards D–L, batch 2
// (#4470). Each test plays the real catalog entry and asserts what its
// text says.

const (
	lciDL2DismantlerOracle = "7129bdec-83e0-4c1e-a87b-ac0da1a714b3"
	lciDL2SlumberOracle    = "6a381297-ebce-403c-b1dc-b7405fa6902c"
	lciDL2IreOracle        = "781d6981-d389-4677-9eaf-9c46c9192ed6"
	lciDL2PiranhasOracle   = "c37eae65-1afe-4242-b677-24dd40e6f401"
	lciDL2CacheOracle      = "9fa0c9d1-2a63-41a8-8733-39bfaee4eb0a"
	lciDL2MonumentOracle   = "71393988-ad6f-43fd-9978-c0de15ae8e87"
	lciDL2GhaltaOracle     = "b8f16311-be5a-4648-b024-849a61dff0bc"
	lciDL2GishathOracle    = "a311daab-40dc-4433-9ed8-feab7c558bbf"
	lciDL2GlorifierOracle  = "07a8ca73-9890-4c87-8b39-fd9eaea5ef45"
	lciDL2PresenceOracle   = "142a960c-4d8e-4b3b-a0db-435c22df8f92"
	lciDL2IntiOracle       = "176f5ba8-6e7b-4234-a298-b6164f51656f"
	lciDL2LorekeeperOracle = "a3ce0fa9-8742-4e2c-a8a0-00fc009f998a"
	lciDL2KellanOracle     = "7af69a44-e958-42e5-9d65-75e53a601e96"
)

func TestLciDL2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		lciDL2DismantlerOracle:    "Dauntless Dismantler",
		lciDL2SlumberOracle:       "Disturbed Slumber",
		lciDL2IreOracle:           "Dreadmaw's Ire",
		lciDL2PiranhasOracle:      "Eaten by Piranhas",
		lciDL2CacheOracle:         "Explorer's Cache",
		lciDL2MonumentOracle:      "Forgotten Monument",
		lciDL2GhaltaOracle:        "Ghalta, Stampede Tyrant",
		lciDL2GishathOracle:       "Gishath, Sun's Avatar",
		lciDL2GlorifierOracle:     "Glorifier of Suffering",
		lciDL2PresenceOracle:      "In the Presence of Ages",
		lciDL2IntiOracle:          "Inti, Seneschal of the Sun",
		lciDL2LorekeeperOracle:    "Ixalli's Lorekeeper",
		lciDL2KellanOracle:        "Kellan, Daring Traveler",
		lciDL2KellanOracle + "#1": "Journey On",
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

// lciDL2CastNow puts a catalog spell into p's hand and casts it in the
// current step, without walking to a main phase.
func lciDL2CastNow(g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) (uuid.UUID, error) {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id, g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets})
}

// lciDL2PayUnless is the open pay_unless prompt addressed to chooser.
func lciDL2PayUnless(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	c := latestChoiceOfKindFor(g, game.PendingChoicePayUnless, chooser)
	if c == nil {
		t.Fatal("no pay_unless prompt")
	}
	return c
}

// --- Dauntless Dismantler ------------------------------------------------------

func TestLciDL2DauntlessDismantlerTapsOpponentsArtifactsOnEntry(t *testing.T) {
	g, me, opp := spendTable(t)
	b12Push(g, me.ID, "Dauntless Dismantler", "Creature — Human Artificer", lciDL2DismantlerOracle, 1, 4)
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, TreasureToken(), 1)
		_ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1)
	})
	if theirs := trekB33Tokens(g, opp.ID, "Treasure"); len(theirs) != 1 || !theirs[0].Tapped {
		t.Error("an opponent's artifact entered untapped")
	}
	if mine := trekB33Tokens(g, me.ID, "Treasure"); len(mine) != 1 || mine[0].Tapped {
		t.Error("your own artifact entered tapped")
	}
}

// X=2 costs {4}{W} and destroys every artifact with mana value 2, yours
// and theirs; mana value 1 and a token's 0 survive.
func TestLciDL2DauntlessDismantlerDestroysEachArtifactWithManaValueX(t *testing.T) {
	g, me, opp := spendTable(t)
	dis := b12Push(g, me.ID, "Dauntless Dismantler", "Creature — Human Artificer", lciDL2DismantlerOracle, 1, 4)
	artifact := func(owner uuid.UUID, name, cost string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: "Artifact", ManaCost: cost, Owner: owner, Controller: owner,
		})
	}
	myTwo := artifact(me.ID, "My Signet", "{2}")
	theirTwo := artifact(opp.ID, "Their Signet", "{2}")
	theirOne := artifact(opp.ID, "Their Trinket", "{1}")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TreasureToken(), 1) })

	apaMana(me, "W", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, dis, 0, game.ActivateAbilityParams{XValue: 2, Strict: true}); err == nil {
		t.Fatal("activated with X=2 for {3}{W}; {X}{X} is two X")
	}
	apaMana(me, "C")
	p7Activate(t, g, me, dis, 0, game.ActivateAbilityParams{XValue: 2, Strict: true})
	if g.Battlefield.Contains(myTwo) || g.Battlefield.Contains(theirTwo) {
		t.Error("an artifact with mana value 2 survived")
	}
	if !g.Battlefield.Contains(theirOne) {
		t.Error("an artifact with mana value 1 was destroyed")
	}
	if len(trekB33Tokens(g, me.ID, "Treasure")) != 1 {
		t.Error("a Treasure (mana value 0) was destroyed")
	}
	if g.Battlefield.Contains(dis) {
		t.Error("the Dismantler was not sacrificed")
	}
}

// --- Disturbed Slumber ---------------------------------------------------------

func TestLciDL2DisturbedSlumberAnimatesALandThatMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	theirLand := pushLandFor(g, opp.ID, "Their Forest", "Basic Land — Forest")
	wall := reqCreature(g, opp.ID, "Wall A", "", 0, 4)
	if err := castCatalogSpellErr(t, g, "Disturbed Slumber", "Instant", lciDL2SlumberOracle, rfCardTarget(theirLand)); err == nil {
		t.Fatal("targeted a land an opponent controls")
	}
	castCatalogSpell(t, g, "Disturbed Slumber", "Instant", lciDL2SlumberOracle, rfCardTarget(land))
	passPriorityAroundTable(t, g)

	c := gaLayered(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Dinosaur") || !c.HasSubtype("Forest") {
		t.Errorf("the land is %q; want a Forest land that is also a Dinosaur creature", c.TypeLine)
	}
	if p, tough := effectivePTOf(t, g, land); p != 4 || tough != 4 {
		t.Errorf("the land is %d/%d, want 4/4", p, tough)
	}
	if !hasEffectiveKeyword(t, g, land, "reach") || !hasEffectiveKeyword(t, g, land, "haste") {
		t.Error("the land lacks reach or haste")
	}

	reqToBlockers(t, g, land)
	br := reqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); !strings.Contains(got, "must be blocked") {
		t.Errorf("refusal = %q, want the must-be-blocked requirement", got)
	}
	if err := reqBlock(t, g, land, wall); err != nil {
		t.Fatal(err)
	}
}

// --- Dreadmaw's Ire ------------------------------------------------------------

func TestLciDL2DreadmawsIrePumpsAndDestroysAnArtifactOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	raptor := b12Creature(g, me.ID, "Raptor", "Creature — Dinosaur", 3, 3)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	declareAttack(t, g, opp.ID, raptor)
	if _, err := lciDL2CastNow(g, me, "Dreadmaw's Ire", "Instant", lciDL2IreOracle, rfCardTarget(idle)); err == nil {
		t.Fatal("targeted a creature that is not attacking")
	}
	if _, err := lciDL2CastNow(g, me, "Dreadmaw's Ire", "Instant", lciDL2IreOracle, rfCardTarget(raptor)); err != nil {
		t.Fatalf("cast at the attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, raptor); p != 5 || tough != 5 {
		t.Errorf("the attacker is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, raptor, "trample") {
		t.Error("the attacker did not gain trample")
	}
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("defender life %d, want %d", opp.Life, life-5)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the damaged player's artifact was not destroyed")
	}
}

// --- Eaten by Piranhas ---------------------------------------------------------

func TestLciDL2EatenByPiranhasMakesABlackSkeletonOneOne(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	host := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Elf Flier", TypeLine: "Artifact Creature — Elf Warrior", Colors: []string{"G"},
		Power: 4, Toughness: 4, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	auraCast(t, g, "Eaten by Piranhas", lciDL2PiranhasOracle, host)
	if p, tough := effectivePTOf(t, g, host); p != 1 || tough != 1 {
		t.Errorf("enchanted creature is %d/%d, want 1/1", p, tough)
	}
	if hasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the creature kept flying")
	}
	c := gaLayered(t, g, host)
	ch := c.Effective()
	if len(ch.Colors) != 1 || ch.Colors[0] != "B" {
		t.Errorf("colors %v, want black only", ch.Colors)
	}
	if !c.IsCreature() || c.IsArtifact() {
		t.Errorf("types %v, want a creature and nothing else", ch.Types)
	}
	if !c.HasSubtype("Skeleton") || c.HasSubtype("Elf") {
		t.Errorf("subtypes %v, want Skeleton alone", ch.Subtypes)
	}
	if spec, _ := Lookup(lciDL2PiranhasOracle); !hasFold(spec.PrintedKeywords, "flash") {
		t.Error("Eaten by Piranhas has no flash")
	}
}

// --- Explorer's Cache ----------------------------------------------------------

func TestLciDL2ExplorersCacheMovesCountersAndGrowsOnDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cache := msc17Play(t, g, me, game.Card{Name: "Explorer's Cache", TypeLine: "Artifact", OracleID: lciDL2CacheOracle})
	if n := b12Counter(t, g, cache, game.CounterPlusOne); n != 2 {
		t.Fatalf("enters with %d counters, want 2", n)
	}
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, cache, 0, game.ActivateAbilityParams{Targets: cardRefs(theirs)}); err != nil {
		t.Fatalf("move onto any target creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if b12Counter(t, g, cache, game.CounterPlusOne) != 1 || b12Counter(t, g, theirs, game.CounterPlusOne) != 1 {
		t.Fatal("the counter did not move from the Cache to the creature")
	}

	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	b27Kill(g, plain)
	passPriorityAroundTable(t, g)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, cache, game.CounterPlusOne); n != 1 {
		t.Fatalf("a creature with no counter, or an opponent's, dying changed the Cache to %d", n)
	}
	grown := b12Creature(g, me.ID, "Grown", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterByForEffect(me.ID, grown, game.CounterPlusOne, 1) })
	b27Kill(g, grown)
	passPriorityAroundTable(t, g)
	if n := b12Counter(t, g, cache, game.CounterPlusOne); n != 2 {
		t.Errorf("your creature with a +1/+1 counter died: the Cache has %d, want 2", n)
	}
}

func TestLciDL2ExplorersCacheMovesOnlyAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cache := msc17Play(t, g, me, game.Card{Name: "Explorer's Cache", TypeLine: "Artifact", OracleID: lciDL2CacheOracle})
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, cache, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err == nil {
		t.Error("activated in the beginning of combat step")
	}
}

// --- Forgotten Monument --------------------------------------------------------

func TestLciDL2ForgottenMonumentGivesOtherCavesPayLifeMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monument := gaPut(g, me.ID, "Forgotten Monument", "Land — Cave", lciDL2MonumentOracle)
	cave := gaPut(g, me.ID, "Hidden Cave", "Land — Cave", "")
	forest := gaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	theirCave := gaPut(g, opp.ID, "Their Cave", "Land — Cave", "")
	if grantedManaRowsOf(g, cave) != 1 {
		t.Fatal("another Cave you control lacks the ability")
	}
	if grantedManaRowsOf(g, monument) != 0 || grantedManaRowsOf(g, forest) != 0 || grantedManaRowsOf(g, theirCave) != 0 {
		t.Error("the Monument itself, a non-Cave, or an opponent's Cave has the ability")
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, cave, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err != nil {
		t.Fatalf("tap the Cave for any colour: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if !tappedNow(g, cave) {
		t.Error("the Cave did not tap")
	}
}

// --- Ghalta, Stampede Tyrant ---------------------------------------------------

func TestLciDL2GhaltaPutsAnyNumberOfCreatureCardsFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := pushHandCardOfType(me, "Bear A", "Creature — Bear")
	b := pushHandCardOfType(me, "Bear B", "Creature — Bear")
	sorcery := pushHandCardOfType(me, "A Sorcery", "Sorcery")
	ghalta := castHoldCreature(t, g, "Ghalta, Stampede Tyrant", "Legendary Creature — Elder Dinosaur", lciDL2GhaltaOracle, 12, 12)
	if chooseCardsChoiceFor(g, me.ID) == nil {
		passPriorityAroundTable(t, g)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-from-hand prompt")
	}
	if !hasID(pick.ChooseCards, a) || !hasID(pick.ChooseCards, b) || hasID(pick.ChooseCards, sorcery) {
		t.Errorf("offered %v; want the creature cards only", pick.ChooseCards)
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != len(pick.ChooseCards) {
		t.Errorf("bounds %d..%d over %d cards; want any number", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
	answerChooseCards(t, g, me.ID, a, b)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("the chosen creature cards did not enter")
	}
	if !me.Hand.Contains(sorcery) {
		t.Error("the sorcery left the hand")
	}
	if !hasEffectiveKeyword(t, g, ghalta, "trample") {
		t.Error("Ghalta has no trample")
	}
}

// --- Gishath, Sun's Avatar -----------------------------------------------------

func TestLciDL2GishathRevealsThatManyAndPutsDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gishath := b12Push(g, me.ID, "Gishath, Sun's Avatar", "Legendary Creature — Dinosaur Avatar", lciDL2GishathOracle, 7, 6)
	for _, kw := range []string{"vigilance", "trample", "haste"} {
		if !hasEffectiveKeyword(t, g, gishath, kw) {
			t.Errorf("Gishath has no %s", kw)
		}
	}
	lib := trekB48Library(me,
		game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Brontodon", TypeLine: "Creature — Dinosaur"},
		game.Card{Name: "Untouched", TypeLine: "Sorcery"},
	)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: gishath, Target: opp.ID, Amount: 3, Combat: true})
	})
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put prompt")
	}
	if !hasID(pick.ChooseCards, lib[0]) || !hasID(pick.ChooseCards, lib[2]) || hasID(pick.ChooseCards, lib[1]) || hasID(pick.ChooseCards, lib[3]) {
		t.Errorf("offered %v; want the two Dinosaurs among the top three", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, lib[0], lib[2])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(lib[0]) || !g.Battlefield.Contains(lib[2]) {
		t.Error("the Dinosaurs did not enter")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != lib[3] {
		t.Error("the fourth card is not on top: only that many cards were revealed")
	}
	if me.Library.Cards[0].InstanceID != lib[1] {
		t.Error("the rest did not go to the bottom")
	}
}

// --- Glorifier of Suffering ----------------------------------------------------

func TestLciDL2GlorifierSacrificesThenCountersUpToTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	land := b12Permanent(g, me.ID, "My Plains", "Basic Land — Plains")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	glorifier := castHoldCreature(t, g, "Glorifier of Suffering", "Creature — Vampire Soldier", lciDL2GlorifierOracle, 3, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{land}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{glorifier}); err == nil {
		t.Fatal("sacrificed the Glorifier itself")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{rock}); err != nil {
		t.Fatalf("sacrifice an artifact: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	b17PickCards(t, g, me.ID, bear, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if countersOn(g, bear, game.CounterPlusOne) != 1 || countersOn(g, theirs, game.CounterPlusOne) != 1 {
		t.Error("each target should have one +1/+1 counter")
	}
	if countersOn(g, glorifier, game.CounterPlusOne) != 0 {
		t.Error("the Glorifier was not a target and got a counter")
	}
}

func TestLciDL2GlorifierDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	castHoldCreature(t, g, "Glorifier of Suffering", "Creature — Vampire Soldier", lciDL2GlorifierOracle, 3, 2)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt")
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for after declining")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact left after declining")
	}
}

// --- In the Presence of Ages ---------------------------------------------------

func TestLciDL2InThePresenceOfAgesTakesACreatureAndALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := trekB48Library(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Other Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Spell", TypeLine: "Sorcery"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	castCatalogSpell(t, g, "In the Presence of Ages", "Instant", lciDL2PresenceOracle, nil)
	passPriorityAroundTable(t, g)
	ch := latestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 2 {
		t.Fatalf("want an optional pick of up to two: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0], lib[2]}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Errorf("two creature cards: err = %v, want a set refusal", err)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[3]}); err == nil {
		t.Error("a sorcery was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0], lib[1]}); err != nil {
		t.Fatalf("a creature and a land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[0]) || !me.Hand.Contains(lib[1]) {
		t.Error("the picks are not in hand")
	}
	for _, i := range []int{2, 3} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[4]) {
		t.Error("the fifth card left the library")
	}
}

// --- Inti, Seneschal of the Sun ------------------------------------------------

func TestLciDL2IntiDiscardsForACounterAndTrampleThenExilesTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	inti := b12Push(g, me.ID, "Inti, Seneschal of the Sun", "Legendary Creature — Human Knight", lciDL2IntiOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	home := b12Creature(g, me.ID, "Stay-Home Bear", "Creature — Bear", 2, 2)
	discard := pushHandCardOfType(me, "Discard Me", "Sorcery")
	top := putOnTopForTest(me, "Top Card", "Sorcery")

	declareAttack(t, g, opp.ID, inti, bear)
	passPriorityAroundTable(t, g)
	c := lciDL2PayUnless(t, g, me.ID)
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{discard}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	b04WaitForPick(t, g, me.ID)
	pick := latestPickTarget(g, me.ID)
	if hasID(pick.PickTargetCards, home) || !hasID(pick.PickTargetCards, bear) {
		t.Errorf("offered %v; want attacking creatures only", pick.PickTargetCards)
	}
	pickTriggerTarget(t, g, me.ID, bear)
	// The reflexive trigger and the discard trigger wait together: order
	// them, then let both resolve.
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(discard) {
		t.Error("the card was not discarded")
	}
	if countersOn(g, bear, game.CounterPlusOne) != 1 || !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the target did not get a +1/+1 counter and trample")
	}
	if !g.Exile.Contains(top) {
		t.Fatal("discarding did not exile the top card of the library")
	}
	if grantedPermissionOn(g, me.ID, top, game.ZoneExile) == nil {
		t.Error("the exiled card can't be played")
	}
}

func TestLciDL2IntiDecliningDiscardsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	inti := b12Push(g, me.ID, "Inti, Seneschal of the Sun", "Legendary Creature — Human Knight", lciDL2IntiOracle, 2, 2)
	keep := pushHandCardOfType(me, "Keep Me", "Sorcery")
	top := putOnTopForTest(me, "Top Card", "Sorcery")
	declareAttack(t, g, opp.ID, inti)
	passPriorityAroundTable(t, g)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) || g.Exile.Contains(top) {
		t.Error("declining discarded or exiled a card")
	}
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a target was asked for after declining")
	}
}

// --- Ixalli's Lorekeeper -------------------------------------------------------

func TestLciDL2IxallisLorekeeperManaIsForDinosaurs(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	keeper := b12Push(g, me.ID, "Ixalli's Lorekeeper", "Creature — Human Druid", lciDL2LorekeeperOracle, 1, 1)
	if err := g.ActivateManaAbility(me.ID, keeper, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	cost, _ := game.ParseCost("{R}")
	dino := game.Card{Name: "Test Dinosaur", TypeLine: "Creature — Dinosaur", ManaCost: "{R}"}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{R}"}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(dino)) || !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(dino)) {
		t.Error("the mana was refused for a Dinosaur spell or ability")
	}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(bear)) || me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(bear)) {
		t.Error("the mana paid for a non-Dinosaur: stronger than printed")
	}
}

// --- Kellan, Daring Traveler // Journey On -------------------------------------

func TestLciDL2KellanTakesASmallCreatureIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	kellan := b12Push(g, me.ID, "Kellan, Daring Traveler", "Legendary Creature — Human Faerie Scout", lciDL2KellanOracle, 2, 3)
	small := plTop(me, "Small Bear", "Creature — Bear", "{1}{G}{G}")
	declareAttack(t, g, opp.ID, kellan)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(small) {
		t.Error("a creature card with mana value 3 did not go to hand")
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID) != nil {
		t.Error("a choice was offered for a creature with mana value 3 or less")
	}
}

func TestLciDL2KellanMayBinAnythingElse(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		bin                  bool
	}{
		{"Big Bear", "Creature — Bear", "{3}{G}", true},
		{"Big Bear", "Creature — Bear", "{3}{G}", false},
		{"Forest", "Basic Land — Forest", "", true},
	} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		kellan := b12Push(g, me.ID, "Kellan, Daring Traveler", "Legendary Creature — Human Faerie Scout", lciDL2KellanOracle, 2, 3)
		top := plTop(me, tc.name, tc.typeLine, tc.cost)
		declareAttack(t, g, opp.ID, kellan)
		passPriorityAroundTable(t, g)
		answerMayChoice(t, g, me.ID, tc.bin)
		if me.Hand.Contains(top) {
			t.Errorf("%s: went to hand", tc.name)
		}
		if tc.bin != me.Graveyard.Contains(top) {
			t.Errorf("%s, answer %v: in graveyard = %v", tc.name, tc.bin, me.Graveyard.Contains(top))
		}
		if got, err := me.Library.Top(); !tc.bin && (err != nil || got.InstanceID != top) {
			t.Errorf("%s: declined, but the card is not on top", tc.name)
		}
	}
}

// Two of three opponents control an artifact: one plus two Maps.
func TestLciDL2JourneyOnMakesAMapPerOpponentWithAnArtifactPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	others := []*game.Player{}
	for _, p := range g.Seats {
		if p.ID != me.ID {
			others = append(others, p)
		}
	}
	b12Permanent(g, others[0].ID, "Rock A", "Artifact")
	b12Permanent(g, others[1].ID, "Rock B", "Artifact")
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	card := game.Card{
		InstanceID: uuid.New(), OracleID: lciDL2KellanOracle, Layout: game.LayoutAdventure,
		Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Kellan, Daring Traveler", TypeLine: "Legendary Creature — Human Faerie Scout", ManaCost: "{1}{W}", Power: 2, Toughness: 3},
			{Name: "Journey On", TypeLine: "Sorcery — Adventure", ManaCost: "{G}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1}); err != nil {
		t.Fatalf("cast Journey On: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(trekB33Tokens(g, me.ID, "Map")); got != 3 {
		t.Errorf("%d Map tokens, want 3", got)
	}
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("the Adventure did not go to exile")
	}
}
