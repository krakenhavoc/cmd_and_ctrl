package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// eerie_cards_test.go — ADR 0103 PR 4: the Duskmourn eerie creatures.
// Every card is driven through BOTH of its real triggers: a real cast
// of an enchantment spell, and a real unlock of a Room's second door.

const eerieRoomOracle = "eerie-test-room-oracle"

// eerieAnswers are the choices eerieSettle gives any prompt it meets.
type eerieAnswers struct {
	Target game.TargetRef // a pick_target prompt; zero means the first candidate
	Card   uuid.UUID      // a choose_cards prompt; zero means the first candidate
	Option int            // an option_pick prompt
}

// eerieSettle resolves the stack, answering trigger order, trigger
// targets, surveils (keep everything on top), option picks and card
// picks for whoever is asked. A trigger prompt (a "you may") is left
// pending for the test to answer.
func eerieSettle(t *testing.T, g *game.Game, a eerieAnswers) {
	t.Helper()
	for i := 0; i < 96; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerPrompt:
				return
			case game.PendingChoiceTriggerOrder:
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
			case game.PendingChoicePickTarget:
				ref := a.Target
				if ref.ID == uuid.Nil {
					switch {
					case len(c.PickTargetCards) > 0:
						ref = game.TargetRef{Kind: game.TargetCard, ID: c.PickTargetCards[0]}
					case len(c.PickTargetPlayers) > 0:
						ref = game.TargetRef{Kind: game.TargetPlayer, ID: c.PickTargetPlayers[0]}
					default:
						t.Fatalf("a target prompt with no candidates: %+v", c)
					}
				}
				if err := g.ResolvePickTarget(c.ID, c.Chooser, ref); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				answered = true
			case game.PendingChoiceSurveil:
				if err := g.ResolveSurveil(c.ID, c.Chooser, nil, c.ScryCards); err != nil {
					t.Fatalf("ResolveSurveil: %v", err)
				}
				answered = true
			case game.PendingChoiceOptionPick:
				if err := g.ResolveOptionPick(c.ID, c.Chooser, a.Option); err != nil {
					t.Fatalf("ResolveOptionPick: %v", err)
				}
				answered = true
			case game.PendingChoiceChooseCards:
				pick := append([]uuid.UUID(nil), c.ChooseCards...)
				for _, id := range c.ChooseCards {
					if id == a.Card {
						pick = []uuid.UUID{id}
					}
				}
				n := c.ChooseMax
				if n <= 0 || n > len(pick) {
					n = len(pick)
				}
				if err := g.ResolveChooseCards(c.ID, c.Chooser, pick[:n]); err != nil {
					t.Fatalf("ResolveChooseCards: %v", err)
				}
				answered = true
			}
			if answered {
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// eerieEnchantment casts a plain enchantment spell and settles.
func eerieEnchantment(t *testing.T, g *game.Game, a eerieAnswers) {
	t.Helper()
	castCatalogSpell(t, g, "Test Aura", "Enchantment", "", nil)
	eerieSettle(t, g, a)
}

// eerieUnlock fully unlocks a Room seat 0 controls by taking the unlock
// special action on its second door, and settles.
func eerieUnlock(t *testing.T, g *game.Game, a eerieAnswers) {
	t.Helper()
	me := g.Seats[0]
	id := roomsDOnBattlefield(g, me.ID, eerieRoomOracle, "Left Door", "Right Door", game.DoorBit(game.DoorLeft))
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorRight}); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	eerieSettle(t, g, a)
}

// forEachEerieSource runs `run` twice on a fresh table in the main
// phase, once per eerie condition. fire() raises that condition.
func forEachEerieSource(t *testing.T, run func(t *testing.T, g *game.Game, fire func(eerieAnswers))) {
	t.Helper()
	t.Run("an enchantment enters", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		run(t, g, func(a eerieAnswers) { eerieEnchantment(t, g, a) })
	})
	t.Run("a Room is fully unlocked", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		run(t, g, func(a eerieAnswers) { eerieUnlock(t, g, a) })
	})
}

// eerieBody puts a catalog creature on seat 0's battlefield.
func eerieBody(g *game.Game, name, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	me := g.Seats[0]
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: me.ID, Controller: me.ID,
	})
}

func eerieDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("DestroyPermanentForEffect: %v", err)
		}
	})
}

func eerieTokens(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.Name == name && IsToken(c) {
			out = append(out, c)
		}
	}
	return out
}

const (
	entityTrackerOracle      = "c9db5293-4866-46ec-9ae8-059d27bd50fc"
	gremlinTamerOracle       = "a4e34165-4e53-4c31-bf61-eb772bc3c054"
	balemurkLeechOracle      = "a511772c-3739-469d-a5a2-5827b7ce9b5c"
	skullsnapNuisanceOracle  = "99348e1d-e95e-4911-aae0-8fd1a7345a82"
	scrabblingSkullcrabOID   = "1448489c-dfb8-43aa-ae3d-ed832381c1d9"
	optimisticScavengerOID   = "0180c67b-c5c8-4a67-9562-d51f9e7ffff5"
	cultHealerOracle         = "c51da69b-7fd7-43a3-be29-c44fc0c36130"
	erraticApparitionOracle  = "23d5d43c-ec78-42dd-a53e-a1ee91c94b1d"
	dashingBloodsuckerOracle = "211dca11-b633-4deb-9d51-210fd3843995"
	infernalPhantomOracle    = "a8225b2e-62bf-45b7-b573-e3ce301a1eab"
	unwillingVesselOracle    = "d2593334-ce0b-43e1-97ad-b00950bfae2b"
	stalkedResearcherOracle  = "8ec3c334-8a53-46d8-8cfb-c647c3a1ef74"
	fearOfInfinityOracle     = "018327d1-a3ba-4912-9c88-f0f0c54a1750"
	victorOracle             = "92bf8f05-cbfd-42d8-8bc1-d7f2e46f9872"
)

func TestEntityTrackerDrawsOnEitherCondition(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me := g.Seats[0]
		eerieBody(g, "Entity Tracker", entityTrackerOracle, 2, 3)
		before := me.Hand.Size()
		fire(eerieAnswers{})
		if got := me.Hand.Size(); got != before+1 {
			t.Fatalf("hand %d -> %d, want one card drawn", before, got)
		}
	})
}

// Nothing else triggers eerie: a creature entering, or the first door
// of a Room (which does not fully unlock it).
func TestEerieIgnoresACreatureAndAHalfUnlock(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	eerieBody(g, "Entity Tracker", entityTrackerOracle, 2, 3)
	before := me.Hand.Size()

	castCatalogSpell(t, g, "Some Bear", "Creature — Bear", "", nil)
	eerieSettle(t, g, eerieAnswers{})
	id := roomsDOnBattlefield(g, me.ID, eerieRoomOracle, "Left Door", "Right Door", 0)
	if err := g.PerformSpecialAction(me.ID, id, game.SpecialActionUnlock, game.SpecialActionParams{Door: game.DoorLeft}); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	eerieSettle(t, g, eerieAnswers{})
	if got := me.Hand.Size(); got != before {
		t.Fatalf("hand %d -> %d: nothing should have triggered eerie", before, got)
	}
}

func TestGremlinTamerMakesAGremlin(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me := g.Seats[0]
		eerieBody(g, "Gremlin Tamer", gremlinTamerOracle, 2, 2)
		fire(eerieAnswers{})
		gremlins := eerieTokens(g, me.ID, "Gremlin")
		if len(gremlins) != 1 || gremlins[0].Power != 1 || gremlins[0].Toughness != 1 || gremlins[0].Colors[0] != "R" {
			t.Fatalf("want one 1/1 red Gremlin, got %+v", gremlins)
		}
	})
}

func TestBalemurkLeechDrainsEachOpponent(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		eerieBody(g, "Balemurk Leech", balemurkLeechOracle, 2, 2)
		before := b30Lives(g)
		fire(eerieAnswers{})
		for i, p := range g.Seats {
			want := before[i]
			if i != 0 {
				want--
			}
			if p.Life != want {
				t.Errorf("seat %d life %d -> %d, want %d", i, before[i], p.Life, want)
			}
		}
	})
}

func TestSkullsnapNuisanceSurveilsOne(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me := g.Seats[0]
		eerieBody(g, "Skullsnap Nuisance", skullsnapNuisanceOracle, 1, 4, "flying")
		grave := me.Graveyard.Size()
		topCard, _ := me.Library.Top()
		top := topCard.InstanceID
		fire(eerieAnswers{}) // eerieSettle answers the surveil, keeping the card on top
		if me.Graveyard.Size() != grave {
			t.Fatalf("keeping the card on top must not bin it")
		}
		if now, _ := me.Library.Top(); now.InstanceID != top {
			t.Fatalf("the surveiled card should have stayed on top")
		}
	})
}

func TestScrabblingSkullcrabMillsTheChosenPlayerTwo(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		victim := g.Seats[2]
		eerieBody(g, "Scrabbling Skullcrab", scrabblingSkullcrabOID, 0, 3)
		lib, grave := victim.Library.Size(), victim.Graveyard.Size()
		fire(eerieAnswers{Target: game.TargetRef{Kind: game.TargetPlayer, ID: victim.ID}})
		if victim.Library.Size() != lib-2 || victim.Graveyard.Size() != grave+2 {
			t.Fatalf("library %d -> %d, graveyard %d -> %d, want two milled", lib, victim.Library.Size(), grave, victim.Graveyard.Size())
		}
	})
}

func TestOptimisticScavengerPutsACounterOnTheChosenCreature(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		eerieBody(g, "Optimistic Scavenger", optimisticScavengerOID, 1, 1)
		other := pushTypedCard(g, g.Seats[1].ID, "Opposing Bear", "Creature — Bear", "{1}{G}")
		fire(eerieAnswers{Target: game.TargetRef{Kind: game.TargetCard, ID: other}})
		c, ok := g.LookupCardForEffect(other)
		if !ok || c.Counters[game.CounterPlusOne] != 1 {
			t.Fatalf("the chosen creature should hold one +1/+1 counter, got %+v", c.Counters)
		}
	})
}

func TestCultHealerGainsLifelink(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		id := eerieBody(g, "Cult Healer", cultHealerOracle, 3, 3)
		if hasKeywordOnBattlefield(t, g, id, "lifelink") {
			t.Fatal("no lifelink before the trigger")
		}
		fire(eerieAnswers{})
		if !hasKeywordOnBattlefield(t, g, id, "lifelink") {
			t.Fatal("the Healer should have lifelink")
		}
	})
}

func TestErraticApparitionGetsPlusOnePlusOneAndStacks(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		id := eerieBody(g, "Erratic Apparition", erraticApparitionOracle, 1, 3, "flying", "vigilance")
		fire(eerieAnswers{})
		if got := b31CurrentPowerOf(g, id); got != 2 {
			t.Fatalf("power %d, want 2", got)
		}
		fire(eerieAnswers{})
		if got := b31CurrentPowerOf(g, id); got != 3 {
			t.Fatalf("power %d after a second trigger, want 3", got)
		}
	})
}

func TestDashingBloodsuckerGetsPlusTwoAndLifelink(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		id := eerieBody(g, "Dashing Bloodsucker", dashingBloodsuckerOracle, 2, 5)
		fire(eerieAnswers{})
		if got := b31CurrentPowerOf(g, id); got != 4 {
			t.Fatalf("power %d, want 4", got)
		}
		if !hasKeywordOnBattlefield(t, g, id, "lifelink") {
			t.Fatal("the Bloodsucker should have lifelink")
		}
	})
}

func TestInfernalPhantomPumpsThenDiesForItsPower(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		opp := g.Seats[1]
		id := eerieBody(g, "Infernal Phantom", infernalPhantomOracle, 2, 3)
		fire(eerieAnswers{})
		if got := b31CurrentPowerOf(g, id); got != 4 {
			t.Fatalf("power %d, want 4", got)
		}
		before := opp.Life
		eerieDestroy(t, g, id)
		eerieSettle(t, g, eerieAnswers{Target: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}})
		if opp.Life != before-4 {
			t.Fatalf("life %d -> %d, want the pumped power (4) dealt", before, opp.Life)
		}
	})
}

func TestInfernalPhantomUnpumpedDealsItsPrintedPower(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	opp := g.Seats[1]
	id := eerieBody(g, "Infernal Phantom", infernalPhantomOracle, 2, 3)
	before := opp.Life
	eerieDestroy(t, g, id)
	eerieSettle(t, g, eerieAnswers{Target: game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}})
	if opp.Life != before-2 {
		t.Fatalf("opponent life %d -> %d, want 2 damage", before, opp.Life)
	}
}

func TestUnwillingVesselGrowsCountersThenLeavesASpiritOfThatSize(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me := g.Seats[0]
		id := eerieBody(g, "Unwilling Vessel", unwillingVesselOracle, 3, 2, "vigilance")
		fire(eerieAnswers{})
		fire(eerieAnswers{})
		c, _ := g.LookupCardForEffect(id)
		if c.Counters["possession"] != 2 {
			t.Fatalf("possession counters %d, want 2", c.Counters["possession"])
		}
		eerieDestroy(t, g, id)
		eerieSettle(t, g, eerieAnswers{})
		spirits := eerieTokens(g, me.ID, "Spirit")
		if len(spirits) != 1 {
			t.Fatalf("want one Spirit token, got %d", len(spirits))
		}
		if got := b31CurrentPowerOf(g, spirits[0].InstanceID); got != 2 {
			t.Errorf("Spirit power %d, want 2", got)
		}
		if !hasKeywordOnBattlefield(t, g, spirits[0].InstanceID, "flying") || spirits[0].Colors[0] != "U" {
			t.Errorf("the Spirit should be blue with flying: %+v", spirits[0])
		}
	})
}

func TestStalkedResearcherCanAttackThisTurn(t *testing.T) {
	attack := func(g *game.Game, id uuid.UUID) error {
		advanceTo(t, g, game.StepDeclareAttackers)
		return g.DeclareAttacker(id, g.Seats[1].ID)
	}
	t.Run("a defender cannot attack", func(t *testing.T) {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		id := eerieBody(g, "Stalked Researcher", stalkedResearcherOracle, 3, 3, "defender")
		if err := attack(g, id); err == nil {
			t.Fatal("a creature with defender attacked")
		}
	})
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		id := eerieBody(g, "Stalked Researcher", stalkedResearcherOracle, 3, 3, "defender")
		fire(eerieAnswers{})
		if err := attack(g, id); err != nil {
			t.Fatalf("after eerie the Researcher should attack: %v", err)
		}
	})
}

func TestFearOfInfinityReturnsFromTheGraveyardOnEitherCondition(t *testing.T) {
	forEachEerieSource(t, func(t *testing.T, g *game.Game, fire func(eerieAnswers)) {
		me := g.Seats[0]
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Fear of Infinity", TypeLine: "Enchantment Creature — Nightmare",
			OracleID: fearOfInfinityOracle, Owner: me.ID, Controller: me.ID,
		})
		fire(eerieAnswers{})
		answerLatestTriggerPrompt(t, g, me.ID, true)
		eerieSettle(t, g, eerieAnswers{})
		if inZone(me.Graveyard, id) || !inZone(me.Hand, id) {
			t.Fatal("Fear of Infinity should have returned to its owner's hand")
		}
	})
}

func TestFearOfInfinityDeclinedStaysInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Fear of Infinity", TypeLine: "Enchantment Creature — Nightmare",
		OracleID: fearOfInfinityOracle, Owner: me.ID, Controller: me.ID,
	})
	eerieEnchantment(t, g, eerieAnswers{})
	answerLatestTriggerPrompt(t, g, me.ID, false)
	eerieSettle(t, g, eerieAnswers{})
	if !inZone(me.Graveyard, id) {
		t.Fatal("declining the may must leave it in the graveyard")
	}
}

// Victor counts his own resolutions: surveil 2, then each opponent
// discards, then a creature card from any graveyard comes back under
// his controller's control, then nothing.
func TestVictorEscalatesOverThreeResolutions(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	eerieBody(g, "Victor", victorOracle, 3, 3)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: dead, Name: "Fallen Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: opp.ID,
	})

	handsBefore := []int{}
	for _, p := range g.Seats {
		handsBefore = append(handsBefore, p.Hand.Size())
	}
	libBefore := me.Library.Size()

	eerieEnchantment(t, g, eerieAnswers{}) // first: surveil 2 (kept on top)
	if me.Library.Size() != libBefore {
		t.Fatalf("surveil keeping both cards must not shrink the library")
	}
	for i, p := range g.Seats {
		if p.Hand.Size() != handsBefore[i] {
			t.Fatalf("seat %d hand changed on the first resolution", i)
		}
	}

	eerieEnchantment(t, g, eerieAnswers{}) // second: each opponent discards a card
	for i, p := range g.Seats[1:] {
		if p.Hand.Size() != handsBefore[i+1]-1 {
			t.Fatalf("opponent %d should have discarded one card: %d -> %d", i+1, handsBefore[i+1], p.Hand.Size())
		}
	}
	if me.Hand.Size() != handsBefore[0] {
		t.Fatalf("the controller discards nothing")
	}

	eerieEnchantment(t, g, eerieAnswers{Card: dead}) // third: reanimate
	c, ok := g.LookupCardForEffect(dead)
	if !ok || c.Controller != me.ID || !onBattlefield(g, dead) {
		t.Fatalf("the Ogre should be on the battlefield under seat 0's control: %+v ok=%v", c, ok)
	}

	sizes := []int{}
	for _, p := range g.Seats {
		sizes = append(sizes, p.Hand.Size()+p.Graveyard.Size()+p.Library.Size())
	}
	eerieEnchantment(t, g, eerieAnswers{}) // fourth: nothing
	for i, p := range g.Seats {
		if p.Hand.Size()+p.Graveyard.Size()+p.Library.Size() != sizes[i] {
			t.Fatalf("a fourth resolution must do nothing (seat %d)", i)
		}
	}
}
