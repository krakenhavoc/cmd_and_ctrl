package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m16_cards_test.go — Marvel Super Heroes Commander, M–Z batch 16
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM16FabricatorOracle = "127f6985-c1f4-487f-b5ee-c84524cd6c55"
	mscM16SpectralOracle   = "1a44261a-2b6a-45dc-91a4-b2eb219ba6fb"
	mscM16AvengerOracle    = "2f4360f6-9ac7-499f-b0a5-769550ea15ba"
	mscM16WakandaOracle    = "b8add484-8ed1-4a1c-aff9-dec65d60e4b1"
	mscM16WandaOracle      = "1448eadb-e8b1-4e8f-9697-3bb919085dab"
	mscM16WarMachineOracle = "b4f35cdd-3857-4a7d-8ef6-df5424dd40e7"
)

func TestMscM16EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscM16FabricatorOracle: "Shuri's Fabricator",
		mscM16SpectralOracle:   "Vision, Spectral Synthezoid",
		mscM16AvengerOracle:    "Vision, Synthezoid Avenger",
		mscM16WakandaOracle:    "Wakanda Forever!",
		mscM16WandaOracle:      "Wanda's Vision",
		mscM16WarMachineOracle: "War Machine, Avenging Arsenal",
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

// --- Shuri's Fabricator ------------------------------------------------------

func TestMscM16FabricatorMakesTwoTappedVibraniumWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	msc17Play(t, g, me, game.Card{Name: "Shuri's Fabricator", TypeLine: "Artifact", OracleID: mscM16FabricatorOracle, ManaCost: "{4}"})
	tokens := msc17NamedOnBattlefield(g, "Vibranium")
	if len(tokens) != 2 {
		t.Fatalf("%d Vibranium tokens, want 2", len(tokens))
	}
	for _, tok := range tokens {
		if !tok.Tapped || tok.Controller != me.ID {
			t.Errorf("token tapped=%v controller=%v, want tapped and mine", tok.Tapped, tok.Controller)
		}
	}
}

func TestMscM16FabricatorReturnsAnArtifactCardWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	fab := pushCatalogPermanent(g, me.ID, "Shuri's Fabricator", "Artifact", mscM16FabricatorOracle, false)
	ring := pushCatalogGraveyardCard(me, "Sol Ring", "Artifact", "", 0, 0)
	bear := pushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)

	floatForTest(g, me, "CCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, fab, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(bear)}); err == nil {
		t.Fatal("targeted a creature card that is not an artifact")
	}
	if err := g.ActivateCatalogAbility(me.ID, fab, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ring)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, ring)
	if !ok {
		t.Fatal("the artifact card did not return")
	}
	if c.Counters[game.CounterFinality] != 1 {
		t.Errorf("finality counters %d, want 1", c.Counters[game.CounterFinality])
	}
	if f, _ := battlefieldCard(g, fab); !f.Tapped {
		t.Error("the Fabricator's {T} was not paid")
	}
}

func TestMscM16FabricatorIsSorcerySpeedAndCostsSix(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	fab := pushCatalogPermanent(g, me.ID, "Shuri's Fabricator", "Artifact", mscM16FabricatorOracle, false)
	ring := pushCatalogGraveyardCard(me, "Sol Ring", "Artifact", "", 0, 0)

	floatForTest(g, me, "CCCCC")
	if err := g.ActivateCatalogAbility(me.ID, fab, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ring)}); err == nil {
		t.Error("activated for five mana: the ability costs {6}")
	}
	floatForTest(g, me, "C")
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, fab, 0, game.ActivateAbilityParams{Strict: true, Targets: cardRefs(ring)}); err == nil {
		t.Error("activated at the beginning of combat: sorcery speed only")
	}
	if _, ok := battlefieldCard(g, ring); ok {
		t.Error("a refused activation returned the card")
	}
}

// --- War Machine, Avenging Arsenal -------------------------------------------

func TestMscM16WarMachineGivesAttackingModifiedCreaturesDoubleStrike(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wm := b12Push(g, me.ID, "War Machine, Avenging Arsenal", "Legendary Artifact Creature — Human Hero", mscM16WarMachineOracle, 3, 5)
	countered := b12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	equipped := b12Creature(g, me.ID, "Armed Bear", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain Bear", "Creature — Bear", 2, 2)
	idle := b12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(countered, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
		if err := g.AddCounterForEffect(idle, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
		if err := g.AddCounterForEffect(theirs, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Blade", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID,
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: equipped},
	})

	declareAttack(t, g, opp.ID, wm, countered, equipped, plain)
	passPriorityAroundTable(t, g)

	for name, id := range map[string]uuid.UUID{"counter": countered, "equipment": equipped} {
		if !hasKeywordOnBattlefield(t, g, id, "double strike") {
			t.Errorf("the attacking modified creature (%s) did not gain double strike", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"unmodified attacker": plain, "modified non-attacker": idle, "War Machine herself": wm, "an opponent's": theirs} {
		if hasKeywordOnBattlefield(t, g, id, "double strike") {
			t.Errorf("%s gained double strike", name)
		}
	}
}

func TestMscM16WarMachineCountsHerselfWhenSheHasACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wm := b12Push(g, me.ID, "War Machine, Avenging Arsenal", "Legendary Artifact Creature — Human Hero", mscM16WarMachineOracle, 3, 5)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(wm, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	declareAttack(t, g, opp.ID, wm)
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, wm, "double strike") {
		t.Error("an attacking War Machine with a counter did not gain double strike")
	}
}

func TestMscM16WarMachineGrantsNothingWhenSheDoesNotAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "War Machine, Avenging Arsenal", "Legendary Artifact Creature — Human Hero", mscM16WarMachineOracle, 3, 5)
	bear := b12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(bear, game.CounterPlusOne, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	declareAttack(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if hasKeywordOnBattlefield(t, g, bear, "double strike") {
		t.Error("double strike without War Machine attacking")
	}
}

// --- Wanda's Vision ----------------------------------------------------------

func mscM16CastFree(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	id := handCardOf(g, p, "Spell", typeLine, "", "")
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func mscM16LibraryWithHit(me *game.Player) (hit, land uuid.UUID) {
	hit, land = uuid.New(), uuid.New()
	me.Library.Cards = nil
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: hit, Name: "Two Drop", TypeLine: "Creature — Test", ManaCost: "{1}{G}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	return hit, land
}

func TestMscM16WandasVisionCastsTheNonlandCardFreeOnYourSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Wanda's Vision", "Enchantment", mscM16WandaOracle, false)
	hit, land := mscM16LibraryWithHit(me)

	mscM16CastFree(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil || g.Exile.Contains(land) {
		t.Fatal("the first spell of the turn triggered Wanda's Vision")
	}

	mscM16CastFree(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, hit)
	if err := castThroughResolutionCast(t, g, me.ID, hit, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)

	if _, ok := battlefieldCard(g, hit); !ok {
		t.Error("the nonland card was not cast")
	}
	if !g.Exile.Contains(land) {
		t.Error("the land exiled on the way left exile")
	}
}

func TestMscM16WandasVisionDecliningLeavesTheCardInExile(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Wanda's Vision", "Enchantment", mscM16WandaOracle, false)
	hit, _ := mscM16LibraryWithHit(me)

	mscM16CastFree(t, g, me, "Instant")
	mscM16CastFree(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordPaidCast, hit)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(hit) || g.Battlefield.Contains(hit) {
		t.Error("the declined card left exile or was cast")
	}
}

func TestMscM16WandasVisionOnlyTriggersOnYourSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Wanda's Vision", "Enchantment", mscM16WandaOracle, false)
	hit, _ := mscM16LibraryWithHit(me)

	// An opponent's second spell is not yours.
	mscM16CastFree(t, g, opp, "Instant")
	mscM16CastFree(t, g, opp, "Instant")
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil || g.Exile.Contains(hit) {
		t.Fatal("an opponent's second spell triggered Wanda's Vision")
	}

	// Mine: the second triggers, the third does not.
	mscM16CastFree(t, g, me, "Instant")
	mscM16CastFree(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	library := me.Library.Size()
	mscM16CastFree(t, g, me, "Instant")
	passPriorityAroundTable(t, g)
	if resolutionCastOfferFor(g, me.ID) != nil || me.Library.Size() != library {
		t.Error("a third spell triggered Wanda's Vision")
	}
}

// --- Vision, Synthezoid Avenger ----------------------------------------------

func mscM16Avenger(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Vision, Synthezoid Avenger", "Legendary Artifact Creature — Robot Hero", mscM16AvengerOracle, 3, 3)
}

func TestMscM16AvengerGrowsWhenAnotherPlayerCastsOnTheirOffTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vision := mscM16Avenger(g, me)

	mscM16CastFree(t, g, opp, "Instant")
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("an opponent casting on my turn did not trigger Vision")
	}
	answerMode(t, g, c, me.ID, 0)
	passPriorityAroundTable(t, g)
	if n := counterCount(g, vision, game.CounterPlusOne); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	if phasedOutInCatalogGame(g, vision) {
		t.Error("Vision phased out on the counter bullet")
	}
}

func TestMscM16AvengerPhasesOutOnTheSecondBullet(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	vision := mscM16Avenger(g, me)

	mscM16CastFree(t, g, opp, "Instant")
	c := modePickChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no mode prompt")
	}
	answerMode(t, g, c, me.ID, 1)
	passPriorityAroundTable(t, g)
	if !phasedOutInCatalogGame(g, vision) {
		t.Fatal("Vision did not phase out")
	}
	if n := counterCount(g, vision, game.CounterPlusOne); n > 0 {
		t.Errorf("a counter was also put: %d", n)
	}
}

func TestMscM16AvengerIgnoresASpellCastOnItsCastersOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vision := mscM16Avenger(g, me)

	mscM16CastFree(t, g, me, "Instant")
	if modePickChoiceFor(g, me.ID) != nil {
		t.Fatal("a spell cast on its caster's own turn triggered Vision")
	}
	passPriorityAroundTable(t, g)
	if n := counterCount(g, vision, game.CounterPlusOne); n != 0 {
		t.Errorf("%d counters, want 0", n)
	}

	// And on the opponent's turn, my own cast triggers it too (it is not my turn).
	powerUpPassTo(t, g, 1)
	mscM16CastFree(t, g, me, "Instant")
	if modePickChoiceFor(g, me.ID) == nil {
		t.Error("my spell cast on an opponent's turn did not trigger Vision")
	}
}

// --- Vision, Spectral Synthezoid ---------------------------------------------

func mscM16Spectral(g *game.Game, me *game.Player) uuid.UUID {
	return b12Push(g, me.ID, "Vision, Spectral Synthezoid", "Legendary Artifact Creature — Robot Hero", mscM16SpectralOracle, 2, 5)
}

func mscM16HandCard(me *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	return id
}

func TestMscM16SpectralCastsANoncreatureSpellFreeOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	mscM16Spectral(g, me)
	first := mscM16HandCard(me, "Big Sorcery", "Sorcery", "{6}{U}")
	second := mscM16HandCard(me, "Other Sorcery", "Sorcery", "{6}{U}")
	key := grantedAltCostVisionSpectral

	if offerWithKey(grantedCardOffers(g, me.ID, first, game.ZoneHand), key) == nil {
		t.Fatal("no free offer for a noncreature spell")
	}
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if offerWithKey(grantedCardOffers(g, me.ID, second, game.ZoneHand), key) != nil {
		t.Error("offered a second free cast this turn")
	}
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Error("the cast gate accepted a second free cast this turn")
	}
}

func TestMscM16SpectralCastsARobotCreatureButNotOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	mscM16Spectral(g, me)
	robot := mscM16HandCard(me, "Test Robot", "Artifact Creature — Robot", "{5}")
	bear := mscM16HandCard(me, "Test Bear", "Creature — Bear", "{5}")
	key := grantedAltCostVisionSpectral

	if offerWithKey(grantedCardOffers(g, me.ID, bear, game.ZoneHand), key) != nil {
		t.Error("offered a free cast for a creature that is not a Robot")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Error("the cast gate accepted a free non-Robot creature")
	}
	if got := offerWithKey(grantedCardOffers(g, me.ID, robot, game.ZoneHand), key); got == nil || got.ManaCost != "" {
		t.Fatalf("Robot offer = %+v, want a free one", got)
	}
	if err := g.CastSpell(me.ID, robot, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free Robot cast: %v", err)
	}
}

func TestMscM16SpectralOneCastOfEitherKindSpendsTheTurnsUse(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	mscM16Spectral(g, me)
	robot := mscM16HandCard(me, "Test Robot", "Artifact Creature — Robot", "{5}")
	sorcery := mscM16HandCard(me, "Big Sorcery", "Sorcery", "{6}{U}")
	key := grantedAltCostVisionSpectral

	if err := g.CastSpell(me.ID, robot, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("free Robot cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if offerWithKey(grantedCardOffers(g, me.ID, sorcery, game.ZoneHand), key) != nil {
		t.Error("a noncreature spell was offered after the Robot used the turn's cast")
	}
}

func TestMscM16SpectralOnlyOnYourOwnTurnAndNotForLands(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	toMain(t, g)
	mscM16Spectral(g, me)
	sorcery := mscM16HandCard(me, "Big Sorcery", "Sorcery", "{6}{U}")
	land := mscM16HandCard(me, "Forest", "Basic Land — Forest", "")
	key := grantedAltCostVisionSpectral

	if offerWithKey(grantedCardOffers(g, me.ID, land, game.ZoneHand), key) != nil {
		t.Error("a land was offered a free cast")
	}
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	if offerWithKey(grantedCardOffers(g, me.ID, sorcery, game.ZoneHand), key) != nil {
		t.Error("offered a free cast during an opponent's turn")
	}
	advanceToMainOf(t, g, seat)
	if offerWithKey(grantedCardOffers(g, me.ID, sorcery, game.ZoneHand), key) == nil {
		t.Error("the free cast did not come back on my next turn")
	}
}

// --- Wakanda Forever! --------------------------------------------------------

type mscM16Six struct {
	creature, otherCreature, land, sorcery, artifact, enchantment, below uuid.UUID
}

// mscM16StackSix puts six known cards on top of the library (a seventh below).
func mscM16StackSix(me *game.Player) mscM16Six {
	s := mscM16Six{
		creature: uuid.New(), otherCreature: uuid.New(), land: uuid.New(),
		sorcery: uuid.New(), artifact: uuid.New(), enchantment: uuid.New(), below: uuid.New(),
	}
	me.Library.Cards = nil
	push := func(id uuid.UUID, name, typeLine string) {
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
	}
	push(s.below, "Below", "Sorcery")
	push(s.enchantment, "Enchant", "Enchantment")
	push(s.artifact, "Rock", "Artifact")
	push(s.sorcery, "Spell", "Sorcery")
	push(s.land, "Forest", "Basic Land — Forest")
	push(s.otherCreature, "Other Bear", "Creature — Bear")
	push(s.creature, "Bear", "Creature — Bear")
	return s
}

func TestMscM16WakandaPutsOneOnTheBattlefieldWithAnIndestructibleCounterOneInHandRestToGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s := mscM16StackSix(me)

	castCatalogSpell(t, g, "Wakanda Forever!", "Sorcery", mscM16WakandaOracle, nil)
	passPriorityAroundTable(t, g)

	first := chooseCardsChoiceFor(g, me.ID)
	if first == nil {
		t.Fatal("no battlefield pick")
	}
	if first.ChooseMin != 0 || first.ChooseMax != 1 {
		t.Errorf("battlefield bounds %d..%d, want 0..1", first.ChooseMin, first.ChooseMax)
	}
	for _, id := range first.ChooseCards {
		if id == s.sorcery {
			t.Error("a sorcery card was offered for the battlefield")
		}
		if id == s.below {
			t.Error("the seventh card was offered")
		}
	}
	answerChooseCards(t, g, me.ID, s.creature)
	passPriorityAroundTable(t, g)

	c, ok := battlefieldCard(g, s.creature)
	if !ok {
		t.Fatal("the picked permanent did not enter")
	}
	if c.Counters[game.CounterIndestructible] != 1 {
		t.Errorf("indestructible counters %d, want 1", c.Counters[game.CounterIndestructible])
	}

	second := chooseCardsChoiceFor(g, me.ID)
	if second == nil {
		t.Fatal("no hand pick")
	}
	for _, id := range second.ChooseCards {
		if id == s.creature || id == s.sorcery {
			t.Errorf("hand pick offered %s", id)
		}
	}
	answerChooseCards(t, g, me.ID, s.artifact)
	passPriorityAroundTable(t, g)

	if !me.Hand.Contains(s.artifact) {
		t.Error("the artifact did not go to hand")
	}
	for _, id := range []uuid.UUID{s.otherCreature, s.land, s.sorcery, s.enchantment} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s did not go to the graveyard", id)
		}
	}
	if !me.Library.Contains(s.below) {
		t.Error("the seventh card left the library")
	}
}

func TestMscM16WakandaDecliningBothPutsEverythingInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s := mscM16StackSix(me)
	castCatalogSpell(t, g, "Wakanda Forever!", "Sorcery", mscM16WakandaOracle, nil)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{s.creature, s.otherCreature, s.land, s.sorcery, s.artifact, s.enchantment} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s is not in the graveyard", id)
		}
	}
	if me.Hand.Contains(s.artifact) || g.Battlefield.Contains(s.creature) {
		t.Error("declining both still put a card somewhere")
	}
}

func TestMscM16WakandaWithNoPermanentCardsPutsAllSixInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := make([]uuid.UUID, 6)
	me.Library.Cards = nil
	for i := range ids {
		ids[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: ids[i], Name: "Spell", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	}
	castCatalogSpell(t, g, "Wakanda Forever!", "Sorcery", mscM16WakandaOracle, nil)
	passPriorityAroundTable(t, g)
	for _, id := range ids {
		if !me.Graveyard.Contains(id) {
			t.Errorf("card %s is not in the graveyard", id)
		}
	}
}
