package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bro_b20_cards_test.go — The Brothers' War, S82 batch 20 (#5482): the
// two meld pairs, the Saga and the planeswalker. Each test plays the real
// catalog entry and asserts what its text says.

const (
	s82BroDragonEngineOracle   = "022b42a6-ee0f-4080-b973-c1b68938c3be"
	s82BroMishraClaimedOracle  = "26a623a1-60c9-47e3-af96-616e29551865"
	s82BroMishraLostOracle     = "3989308a-4c0f-493a-b38f-cb4f5019dfe7"
	s82BroTitaniaVoiceOracle   = "42bccda5-1754-4619-b94f-7a72f934a718"
	s82BroArgothOracle         = "62648946-1708-48f4-ba40-c057563ab11b"
	s82BroTitaniaGaeaOracle    = "43230d70-4d60-4ceb-bb45-fbc043280652"
	s82BroBrothersWarOracle    = "43040c34-acad-4d62-893d-fd606fa9d8fe"
	s82BroSaheeliFiligreOracle = "68cf7415-b434-4284-89b4-c7d44a003eb8"
)

func TestS82BroB20EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		s82BroDragonEngineOracle:   "Phyrexian Dragon Engine",
		s82BroMishraClaimedOracle:  "Mishra, Claimed by Gix",
		s82BroMishraLostOracle:     "Mishra, Lost to Phyrexia",
		s82BroTitaniaVoiceOracle:   "Titania, Voice of Gaea",
		s82BroArgothOracle:         "Argoth, Sanctum of Nature",
		s82BroTitaniaGaeaOracle:    "Titania, Gaea Incarnate",
		s82BroBrothersWarOracle:    "The Brothers' War",
		s82BroSaheeliFiligreOracle: "Saheeli, Filigree Master",
	}
	for oracle, name := range want {
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

// --- Phyrexian Dragon Engine ---

func s82BroDragonEngineUnearthed(t *testing.T, g *game.Game) (uuid.UUID, *game.Player) {
	t.Helper()
	return broB04Unearth(t, g, broB04Card{"Phyrexian Dragon Engine", "Artifact Creature — Phyrexian Dragon",
		s82BroDragonEngineOracle, 2, 2, "{3}{R}{R}"})
}

func TestS82BroB20DragonEngineHasDoubleStrikeAndUnearth(t *testing.T) {
	g := newCatalogGame(t)
	id, _ := s82BroDragonEngineUnearthed(t, g)
	assertKeywords(t, g, id, "double strike", "haste")
}

func TestS82BroB20DragonEngineDiscardsTheHandAndDrawsThreeWhenItEntersFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	hand := me.Hand.Size()
	if hand == 0 {
		t.Fatal("the test needs cards in hand")
	}
	id := pushCatalogGraveyardCard(me, "Phyrexian Dragon Engine", "Artifact Creature — Phyrexian Dragon", s82BroDragonEngineOracle, 2, 2)
	rfCrAAddMana(t, g, me, broB04Pool("{3}{R}{R}"))
	idx := -1
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: s82BroDragonEngineOracle}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no unearth")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	yard := me.Graveyard.Size()
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (the old hand discarded, three drawn)", got)
	}
	if got := me.Graveyard.Size() - yard; got < hand {
		t.Errorf("%d cards reached the graveyard, want at least the hand of %d", got, hand)
	}
}

func TestS82BroB20DragonEngineKeepsTheHandOnNo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	hand := me.Hand.Size()
	id := pushCatalogGraveyardCard(me, "Phyrexian Dragon Engine", "Artifact Creature — Phyrexian Dragon", s82BroDragonEngineOracle, 2, 2)
	rfCrAAddMana(t, g, me, broB04Pool("{3}{R}{R}"))
	idx := -1
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: s82BroDragonEngineOracle}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no unearth")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after a no, want %d", got, hand)
	}
}

func TestS82BroB20DragonEngineDoesNotTriggerWhenItEntersFromAnotherZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	hand := me.Hand.Cards[0].InstanceID
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == hand {
				me.Hand.Cards[i].Name = "Phyrexian Dragon Engine"
				me.Hand.Cards[i].OracleID = s82BroDragonEngineOracle
				me.Hand.Cards[i].TypeLine = "Artifact Creature — Phyrexian Dragon"
				me.Hand.Cards[i].Power, me.Hand.Cards[i].Toughness = 2, 2
			}
		}
	})
	before := me.Hand.Size()
	g.WithWriteLock(func() {
		if _, err := g.PutFromHandOntoBattlefieldForEffect(hand, game.HandEntryOptions{}); err != nil {
			t.Fatalf("put: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if c := latestConfirmFor(g, me.ID); c != nil {
		t.Errorf("a prompt was asked for an entry from the hand: %+v", c)
	}
	if got := me.Hand.Size(); got != before-1 {
		t.Errorf("hand %d, want %d", got, before-1)
	}
}

// --- Mishra, Claimed by Gix and Mishra, Lost to Phyrexia ---

func s82BroMishraMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   s82BroMishraLostOracle,
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Mishra, Lost to Phyrexia", TypeLine: "Legendary Artifact Creature — Phyrexian Artificer",
			Colors: []string{"B", "R"}, Power: 9, Toughness: 9,
		},
	}
}

func s82BroPushMishraPair(g *game.Game, owner uuid.UUID) (mishra, engine uuid.UUID) {
	mishra = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mishra, Claimed by Gix", OracleID: s82BroMishraClaimedOracle,
		TypeLine: "Legendary Creature — Phyrexian Human Artificer", ManaCost: "{2}{B}{R}",
		Power: 3, Toughness: 5, Colors: []string{"B", "R"}, Owner: owner, Controller: owner,
		Layout: game.LayoutMeld, Meld: s82BroMishraMeldPrint(), ScryfallID: uuid.NewString(),
	})
	engine = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Phyrexian Dragon Engine", OracleID: s82BroDragonEngineOracle,
		TypeLine: "Artifact Creature — Phyrexian Dragon", ManaCost: "{3}", Power: 2, Toughness: 2,
		Owner: owner, Controller: owner, Layout: game.LayoutMeld, Meld: s82BroMishraMeldPrint(), ScryfallID: uuid.NewString(),
	})
	return mishra, engine
}

func s82BroMeldedMishra(g *game.Game) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == s82BroMishraLostOracle {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestS82BroB20MishraClaimedDrainsForEveryAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra := b12Push(g, me.ID, "Mishra, Claimed by Gix", "Legendary Creature — Phyrexian Human Artificer", s82BroMishraClaimedOracle, 3, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	foe := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	lifeMe, lifeOpp, lifeOther := me.Life, opp.Life, g.Seats[2].Life
	declareAttack(t, g, opp.ID, mishra, bear)
	passPriorityAroundTable(t, g)
	// X counts every attacking creature: Mishra and the Bear. The defender's own creature is not one.
	_ = foe
	if me.Life != lifeMe+2 || opp.Life != lifeOpp-2 || g.Seats[2].Life != lifeOther-2 {
		t.Errorf("life me %d opp %d other %d, want +2 / -2 / -2 from %d %d %d",
			me.Life, opp.Life, g.Seats[2].Life, lifeMe, lifeOpp, lifeOther)
	}
}

func TestS82BroB20MishraClaimedFiresOncePerAttackDeclaration(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra := b12Push(g, me.ID, "Mishra, Claimed by Gix", "Legendary Creature — Phyrexian Human Artificer", s82BroMishraClaimedOracle, 3, 5)
	a := b12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	lifeOpp := opp.Life
	declareAttack(t, g, opp.ID, mishra, a, b)
	passPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-3 {
		t.Errorf("opponent life %d, want %d: one trigger, X = 3", opp.Life, lifeOpp-3)
	}
}

func TestS82BroB20MishraClaimedMeldsWithAnAttackingDragonEngine(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra, engine := s82BroPushMishraPair(g, me.ID)
	declareAttack(t, g, opp.ID, mishra, engine)
	passPriorityAroundTable(t, g)
	// What it attacks is its controller's choice (CR 508.4).
	if entryAttackPrompt(g) == nil {
		t.Fatal("the melded permanent entered without being asked what it attacks")
	}
	answerEntryAttack(t, g, g.Seats[2].ID)
	m, ok := s82BroMeldedMishra(g)
	if !ok {
		t.Fatalf("no Mishra, Lost to Phyrexia: %+v", g.Battlefield.Cards)
	}
	if m.AttackingTarget != g.Seats[2].ID {
		t.Errorf("attacking %v, want the player chosen, %v", m.AttackingTarget, g.Seats[2].ID)
	}
	// Its enters trigger fires: it was put onto the battlefield, never declared as an attacker.
	if latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID) == nil {
		t.Error("no \"choose three\" prompt from the melded permanent's enters trigger")
	}
	if !m.IsMelded() || len(m.MeldedFrom) != 2 {
		t.Errorf("not melded from two cards: %+v", m.MeldedFrom)
	}
	if !m.Tapped {
		t.Error("the melded Mishra entered untapped")
	}
	if mv, ok := m.ParsedManaValue(); !ok || mv != 7 {
		t.Errorf("mana value %d, want 7 ({2}{B}{R} + {3})", mv)
	}
	for _, id := range []uuid.UUID{mishra, engine} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v is still on the battlefield beside the melded permanent", id)
		}
	}
}

func TestS82BroB20MishraClaimedDoesNotMeldUnlessTheDragonEngineIsAttacking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra, engine := s82BroPushMishraPair(g, me.ID)
	declareAttack(t, g, opp.ID, mishra)
	passPriorityAroundTable(t, g)
	if _, ok := s82BroMeldedMishra(g); ok {
		t.Fatal("melded with a Dragon Engine that was not attacking")
	}
	if !g.Battlefield.Contains(engine) || !g.Battlefield.Contains(mishra) {
		t.Error("the two cards should both still be on the battlefield")
	}
}

func TestS82BroB20MishraClaimedDoesNotMeldWithADragonEngineYouDoNotOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra, engine := s82BroPushMishraPair(g, me.ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == engine {
				g.Battlefield.Cards[i].Owner = opp.ID
			}
		}
	})
	declareAttack(t, g, opp.ID, mishra, engine)
	passPriorityAroundTable(t, g)
	if _, ok := s82BroMeldedMishra(g); ok {
		t.Fatal("melded with a Dragon Engine its controller does not own")
	}
}

func TestS82BroB20MeldedMishraLeavesAsTwoCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mishra, engine := s82BroPushMishraPair(g, me.ID)
	declareAttack(t, g, opp.ID, mishra, engine)
	passPriorityAroundTable(t, g)
	if c := entryAttackPrompt(g); c != nil {
		answerEntryAttack(t, g, opp.ID)
	}
	m, ok := s82BroMeldedMishra(g)
	if !ok {
		t.Fatal("no melded Mishra")
	}
	b27Kill(g, m.InstanceID)
	passPriorityAroundTable(t, g)
	names := map[string]bool{}
	for _, c := range me.Graveyard.Cards {
		names[c.Name] = true
	}
	if !names["Mishra, Claimed by Gix"] || !names["Phyrexian Dragon Engine"] {
		t.Errorf("graveyard %v, want both cards", names)
	}
}

func TestS82BroB20MishraLostChoosesExactlyThreeDifferentModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := b12Push(g, me.ID, "Mishra, Lost to Phyrexia", "Legendary Artifact Creature — Phyrexian Artificer", s82BroMishraLostOracle, 9, 9)
	declareAttack(t, g, opp.ID, id)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	for _, bad := range [][]int{{3, 4}, {3, 4, 5, 0}, {3, 3, 4}} {
		if err := g.ResolveModePick(pick.ID, me.ID, bad); err == nil {
			t.Errorf("modes %v were accepted; the ability chooses three different modes", bad)
		}
	}
}

func TestS82BroB20MishraLostCreatureModesAndPowerstones(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := b12Push(g, me.ID, "Mishra, Lost to Phyrexia", "Legendary Artifact Creature — Phyrexian Artificer", s82BroMishraLostOracle, 9, 9)
	declareAttack(t, g, opp.ID, id)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{3, 4, 5}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
	assertKeywords(t, g, mine, "menace", "trample")
	assertKeywords(t, g, id, "menace", "trample")
	if c := layeredCard(t, g, theirs); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if c := layeredCard(t, g, mine); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("my Bear is %d/%d, want 2/2: the -1/-1 is for creatures you don't control", c.CurrentPower(), c.CurrentToughness())
	}
	if n, tapped := countTokensNamed(g, me.ID, "Powerstone"); n != 2 || tapped != 2 {
		t.Errorf("%d Powerstones, %d tapped, want 2 tapped", n, tapped)
	}
}

func TestS82BroB20MishraLostDamageAndDestroyAndDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	id := b12Push(g, me.ID, "Mishra, Lost to Phyrexia", "Legendary Artifact Creature — Phyrexian Artificer", s82BroMishraLostOracle, 9, 9)
	declareAttack(t, g, opp.ID, id)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{0, 1, 2}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	lifeOpp := opp.Life
	hand := opp.Hand.Size()
	if hand < 2 {
		t.Fatalf("the opponent needs two cards in hand, has %d", hand)
	}
	// Targets are announced in printed order: the opponent, any target, the artifact.
	answerPickTargetPlayer(t, g, opp.ID)
	answerPickTargetPlayer(t, g, opp.ID)
	answerPickTarget(t, g, relic)
	passPriorityAroundTable(t, g)
	// The bullets resolve in printed order: the discard is asked before the damage and the destroy.
	if got := discardOwed(g, opp.ID); got != 2 {
		t.Fatalf("the opponent owes %d discards, want 2", got)
	}
	if opp.Life != lifeOpp || !g.Battlefield.Contains(relic) {
		t.Fatal("a later bullet resolved before the discard was answered")
	}
	discardFromHand(t, g, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, lifeOpp-3)
	}
	if g.Battlefield.Contains(relic) {
		t.Error("the artifact was not destroyed")
	}
	if got := opp.Hand.Size(); got != hand-2 {
		t.Errorf("opponent hand %d, want %d", got, hand-2)
	}
}

func TestS82BroB20MishraLostDestroyTakesAPlaneswalkerButNotACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushWalkerForTest(g, opp.ID, "Their Walker", "", 3)
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := b12Push(g, me.ID, "Mishra, Lost to Phyrexia", "Legendary Artifact Creature — Phyrexian Artificer", s82BroMishraLostOracle, 9, 9)
	declareAttack(t, g, opp.ID, id)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatal("no mode prompt")
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{2, 4, 5}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	p := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err == nil {
		t.Fatal("a plain creature was accepted for \"target artifact or planeswalker\"")
	}
	answerPickTarget(t, g, walker)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the planeswalker was not destroyed")
	}
}

// --- Titania, Voice of Gaea; Argoth, Sanctum of Nature; Titania, Gaea Incarnate ---

func s82BroTitaniaMeldPrint() *game.MeldPrint {
	return &game.MeldPrint{
		ResultOracleID:   s82BroTitaniaGaeaOracle,
		ResultScryfallID: uuid.NewString(),
		Result: game.Face{
			Name: "Titania, Gaea Incarnate", TypeLine: "Legendary Creature — Elemental Avatar",
			Colors: []string{"G"}, Keywords: []string{"reach", "vigilance", "trample", "haste"},
		},
	}
}

func s82BroPushTitaniaPair(g *game.Game, owner uuid.UUID) (titania, argoth uuid.UUID) {
	titania = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Titania, Voice of Gaea", OracleID: s82BroTitaniaVoiceOracle,
		TypeLine: "Legendary Creature — Elemental", ManaCost: "{1}{G}{G}", Power: 3, Toughness: 4,
		Colors: []string{"G"}, Owner: owner, Controller: owner,
		Layout: game.LayoutMeld, Meld: s82BroTitaniaMeldPrint(), ScryfallID: uuid.NewString(),
	})
	argoth = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Argoth, Sanctum of Nature", OracleID: s82BroArgothOracle,
		TypeLine: "Land", Owner: owner, Controller: owner,
		Layout: game.LayoutMeld, Meld: s82BroTitaniaMeldPrint(), ScryfallID: uuid.NewString(),
	})
	return titania, argoth
}

func s82BroLandsInYard(me *game.Player, n int) {
	for i := 0; i < n; i++ {
		pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	}
}

func s82BroMeldedTitania(g *game.Game) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == s82BroTitaniaGaeaOracle {
			return c, true
		}
	}
	return game.Card{}, false
}

func TestS82BroB20TitaniaGainsTwoLifeWhenALandCardIsPutIntoYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Titania, Voice of Gaea", "Legendary Creature — Elemental", s82BroTitaniaVoiceOracle, 3, 4)
	assertKeywords(t, g, battlefieldIDNamed(g, me.ID, "Titania, Voice of Gaea"), "reach")
	mine := b12Permanent(g, me.ID, "My Forest", "Basic Land — Forest")
	theirs := b12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	life := me.Life
	b27Kill(g, bear)
	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d: a creature or an opponent's land is not a land card put into your graveyard", me.Life, life)
	}
	b27Kill(g, mine)
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d after your own land card reached your graveyard", me.Life, life+2)
	}
}

func TestS82BroB20TitaniaGainsLifeOnceForLandsPutIntoTheGraveyardTogether(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Titania, Voice of Gaea", "Legendary Creature — Elemental", s82BroTitaniaVoiceOracle, 3, 4)
	a := b12Permanent(g, me.ID, "Forest A", "Basic Land — Forest")
	b := b12Permanent(g, me.ID, "Forest B", "Basic Land — Forest")
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(a)
		_ = g.DestroyPermanentForEffect(b)
	})
	passPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d: two lands dying together are one occurrence", me.Life, life+2)
	}
}

func TestS82BroB20TitaniaMeldsWithArgothAtUpkeepWithFourLandCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	titania, argoth := s82BroPushTitaniaPair(g, me.ID)
	// Two lands already out, or the melded Titania (as big as your lands) would die before her trigger resolves.
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	s82BroLandsInYard(me, 4)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	m, ok := s82BroMeldedTitania(g)
	if !ok {
		t.Fatalf("no Titania, Gaea Incarnate: %+v", g.Battlefield.Cards)
	}
	if !m.IsMelded() || len(m.MeldedFrom) != 2 {
		t.Errorf("not melded from two cards: %+v", m.MeldedFrom)
	}
	for _, id := range []uuid.UUID{titania, argoth} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v is still on the battlefield beside the melded permanent", id)
		}
	}
	// Its enters trigger returned every land card from the graveyard, tapped.
	lands, tapped := 0, 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Forest" {
			lands++
			if c.Tapped {
				tapped++
			}
		}
	}
	if lands != 4 || tapped != 4 {
		t.Errorf("%d Forests returned, %d tapped, want 4 tapped", lands, tapped)
	}
	for _, c := range me.Graveyard.Cards {
		if c.IsLand() {
			t.Errorf("land card %q is still in the graveyard", c.Name)
		}
	}
	if c := layeredCard(t, g, m.InstanceID); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Errorf("Titania is %d/%d, want 6/6: one for each land you control", c.CurrentPower(), c.CurrentToughness())
	}
	assertKeywords(t, g, m.InstanceID, "reach", "vigilance", "trample", "haste")
}

func TestS82BroB20TitaniaDoesNotMeldWithThreeLandCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	titania, argoth := s82BroPushTitaniaPair(g, me.ID)
	s82BroLandsInYard(me, 3)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, ok := s82BroMeldedTitania(g); ok {
		t.Fatal("melded with only three land cards in the graveyard")
	}
	if !g.Battlefield.Contains(titania) || !g.Battlefield.Contains(argoth) {
		t.Error("both cards should still be on the battlefield")
	}
}

func TestS82BroB20TitaniaDoesNotMeldWithoutArgoth(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	titania, argoth := s82BroPushTitaniaPair(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(argoth) })
	s82BroLandsInYard(me, 4)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	if _, ok := s82BroMeldedTitania(g); ok || !g.Battlefield.Contains(titania) {
		t.Fatal("Titania melded without a land named Argoth, Sanctum of Nature")
	}
}

func TestS82BroB20TitaniaGaeaIncarnateAnimatesALandWithFourCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	s82BroPushTitaniaPair(g, me.ID)
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	s82BroLandsInYard(me, 4)
	advanceToUpkeepOfSeat(t, g, 0)
	passPriorityAroundTable(t, g)
	m, ok := s82BroMeldedTitania(g)
	if !ok {
		t.Fatal("no melded Titania")
	}
	var forest uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Forest" {
			forest = c.InstanceID
			break
		}
	}
	advanceToMainOf(t, g, 0)
	floatForTest(g, me, "CCCG")
	b16Activate(t, g, me.ID, m.InstanceID, 0, game.ActivateAbilityParams{Strict: true, Targets: cardTarget(forest)})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, forest, game.CounterPlusOne); got != 4 {
		t.Fatalf("%d +1/+1 counters, want 4", got)
	}
	c := layeredCard(t, g, forest)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Elemental") || c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("the land is %q %d/%d, want a 4/4 Elemental land creature", c.TypeLine, c.CurrentPower(), c.CurrentToughness())
	}
	assertKeywords(t, g, forest, "haste")
}

func TestS82BroB20ArgothEntersTappedUnlessYouControlALegendaryGreenCreature(t *testing.T) {
	for _, tc := range []struct {
		name       string
		setup      func(g *game.Game, me *game.Player)
		wantTapped bool
	}{
		{"nothing", func(*game.Game, *game.Player) {}, true},
		{"a legendary green creature", func(g *game.Game, me *game.Player) {
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Legend",
				TypeLine: "Legendary Creature — Elf", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		}, false},
		{"a legendary red creature", func(g *game.Game, me *game.Player) {
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Legend",
				TypeLine: "Legendary Creature — Elf", Colors: []string{"R"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		}, true},
		{"a nonlegendary green creature", func(g *game.Game, me *game.Player) {
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear",
				TypeLine: "Creature — Bear", Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			tc.setup(g, me)
			id := handCardFull(me, "Argoth, Sanctum of Nature", "Land", "", s82BroArgothOracle, nil)
			putFromHand(t, g, id, me.ID)
			if got := b27Tapped(t, g, id); got != tc.wantTapped {
				t.Errorf("tapped = %v, want %v", got, tc.wantTapped)
			}
		})
	}
}

func TestS82BroB20ArgothMakesABearAndMillsThreeAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	argoth := b12Push(g, me.ID, "Argoth, Sanctum of Nature", "Land", s82BroArgothOracle, 0, 0)
	advanceToMainOf(t, g, 0)
	floatForTest(g, me, "CCGG")
	lib := me.Library.Size()
	b16Activate(t, g, me.ID, argoth, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	n, _ := countTokensNamed(g, me.ID, "Bear")
	if n != 1 {
		t.Errorf("%d Bear tokens, want 1", n)
	}
	if got := lib - me.Library.Size(); got != 3 {
		t.Errorf("milled %d cards, want 3", got)
	}
	if !b27Tapped(t, g, argoth) {
		t.Error("Argoth's ability taps it")
	}
}

func TestS82BroB20ArgothsBearAbilityIsSorcerySpeedOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	argoth := b12Push(g, me.ID, "Argoth, Sanctum of Nature", "Land", s82BroArgothOracle, 0, 0)
	advanceToMainOf(t, g, 0)
	advanceTo(t, g, game.StepBeginCombat)
	floatForTest(g, me, "CCGG")
	if err := g.ActivateCatalogAbility(me.ID, argoth, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("the Bear ability was activated outside a main phase")
	}
}

// --- Saheeli, Filigree Master ---

func s82BroSaheeli(t *testing.T, g *game.Game, me *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	id := pushWalkerForTest(g, me.ID, "Saheeli, Filigree Master", s82BroSaheeliFiligreOracle, loyalty)
	advanceToMainOf(t, g, 0)
	return id
}

func s82BroSaheeliPlusOne(t *testing.T, g *game.Game, me *game.Player, walker uuid.UUID) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, walker, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerScryKeepAll(t, g, me.ID)
}

func TestS82BroB20SaheeliPlusOneScriesThenMayTapAnArtifactToDraw(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := s82BroSaheeli(t, g, me, 3)
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	hand := me.Hand.Size()
	s82BroSaheeliPlusOne(t, g, me, walker)
	if got := loyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}
	c := chooseCardsPrompt(t, g)
	if c.ChooseMin != 0 || c.ChooseMax != 1 || len(c.ChooseCards) != 1 || c.ChooseCards[0] != relic {
		t.Fatalf("prompt min %d max %d cards %v, want up to one of the Relic", c.ChooseMin, c.ChooseMax, c.ChooseCards)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{relic}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !b27Tapped(t, g, relic) {
		t.Error("the Relic was not tapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d: one card drawn", got, hand+1)
	}
}

func TestS82BroB20SaheeliPlusOneDrawsNothingIfYouDeclineToTap(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := s82BroSaheeli(t, g, me, 3)
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	hand := me.Hand.Size()
	s82BroSaheeliPlusOne(t, g, me, walker)
	c := chooseCardsPrompt(t, g)
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if b27Tapped(t, g, relic) || me.Hand.Size() != hand {
		t.Errorf("declined: tapped=%v hand %d, want untapped and %d", b27Tapped(t, g, relic), me.Hand.Size(), hand)
	}
}

func TestS82BroB20SaheeliPlusOneOffersOnlyUntappedArtifactsYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := s82BroSaheeli(t, g, me, 3)
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	b12Permanent(g, me.ID, "My Island", "Basic Land — Island")
	tappedRelic := b12Permanent(g, me.ID, "Tapped Relic", "Artifact")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tappedRelic {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hand := me.Hand.Size()
	s82BroSaheeliPlusOne(t, g, me, walker)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			t.Fatalf("an artifact prompt was asked with no untapped artifact of yours: %+v", c.ChooseCards)
		}
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d: nothing to tap, nothing drawn", me.Hand.Size(), hand)
	}
}

func TestS82BroB20SaheeliMinusTwoMakesTwoHastyThopters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := s82BroSaheeli(t, g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := loyaltyOf(g, walker); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller != me.ID || c.Name != "Thopter" {
			continue
		}
		n++
		if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
			t.Errorf("Thopter %+v, want a colorless 1/1 artifact creature token", c)
		}
		assertKeywords(t, g, c.InstanceID, "flying", "haste")
	}
	if n != 2 {
		t.Errorf("%d Thopters, want 2", n)
	}
}

func TestS82BroB20SaheeliThopterHasteEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := s82BroSaheeli(t, g, me, 3)
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-2: %v", err)
	}
	passPriorityAroundTable(t, g)
	thopter := battlefieldIDNamed(g, me.ID, "Thopter")
	advanceToMainOf(t, g, 0)
	for g.Turn.Step != game.StepCleanup {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	for _, kw := range effectiveAbilities(t, g, thopter) {
		if kw == "haste" {
			t.Error("the Thopter still has haste in the cleanup step")
		}
	}
}

func TestS82BroB20SaheeliMinusFourEmblemPumpsAndDiscountsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := s82BroSaheeli(t, g, me, 4)
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	plain := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-4: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(g.EmblemsForPlayer(me.ID)); got != 1 {
		t.Fatalf("%d emblems, want 1", got)
	}
	for id, want := range map[uuid.UUID]int{mine: 3, plain: 2, theirs: 2} {
		if c := layeredCard(t, g, id); c.CurrentPower() != want || c.CurrentToughness() != want {
			t.Errorf("%s is %d/%d, want %d/%d", c.Name, c.CurrentPower(), c.CurrentToughness(), want, want)
		}
	}
	// Artifact spells you cast cost {1} less: a {2} artifact is cast with one mana,
	// a {2} creature that is not an artifact still costs two.
	relic := handCardFull(me, "Relic", "Artifact", "{2}", "", nil)
	bearCard := handCardFull(me, "Bear Cub", "Creature — Bear", "{2}", "", nil)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, bearCard, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("a nonartifact {2} spell was cast with one mana")
	}
	if err := g.CastSpell(me.ID, relic, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("an artifact {2} spell could not be cast with one mana: %v", err)
	}
}

func TestS82BroB20SaheeliEmblemDoesNotDiscountAnOpponentsArtifactSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	walker := s82BroSaheeli(t, g, me, 4)
	if err := g.ActivateCatalogAbility(me.ID, walker, 2, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("-4: %v", err)
	}
	passPriorityAroundTable(t, g)
	opp := g.Seats[1]
	relic := handCardFull(opp, "Relic", "Artifact", "{2}", "", nil)
	floatForTest(g, opp, "C")
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	if err := g.CastSpell(opp.ID, relic, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a {2} artifact spell with one mana: the emblem is yours alone")
	}
	// The same cast is legal with the full price, so the refusal above was the price.
	floatForTest(g, opp, "C")
	if err := g.CastSpell(opp.ID, relic, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the cast was refused for a reason other than its price: %v", err)
	}
}

// --- The Brothers' War ---

func TestS82BroB20BrothersWarChapterOneMakesTwoTappedPowerstones(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "The Brothers' War", "Enchantment — Saga", s82BroBrothersWarOracle, nil)
	passPriorityAroundTable(t, g)
	if n, tapped := countTokensNamed(g, me.ID, "Powerstone"); n != 2 || tapped != 2 {
		t.Errorf("%d Powerstones, %d tapped, want 2 tapped", n, tapped)
	}
}

// chapter II is put on the stack at my next precombat main; it chooses `a` and `b`.
func s82BroBrothersWarChapterTwo(t *testing.T, g *game.Game, a, b uuid.UUID) uuid.UUID {
	t.Helper()
	saga := castCatalogSpell(t, g, "The Brothers' War", "Enchantment — Saga", s82BroBrothersWarOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToPrecombatMainOf(t, g, 0)
	c := pickTargetPromptOpen(g)
	if c == nil {
		t.Fatal("no target prompt for chapter II")
	}
	if err := g.ResolvePickTargets(c.ID, c.Chooser, []game.TargetRef{
		{Kind: game.TargetPlayer, ID: a}, {Kind: game.TargetPlayer, ID: b}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	return saga
}

func pickTargetPromptOpen(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}

func TestS82BroB20BrothersWarChapterTwoPointsEachChosenPlayersCreaturesAtTheOther(t *testing.T) {
	g := newCatalogGame(t)
	me, a, b, c := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	hasty := func(owner *game.Player, name string) uuid.UUID {
		return pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear",
			Power: 2, Toughness: 2, Owner: owner.ID, Controller: owner.ID, Keywords: []string{"haste"}})
	}
	creatureA, creatureB, creatureC := hasty(a, "A's Bear"), hasty(b, "B's Bear"), hasty(c, "C's Bear")
	saga := s82BroBrothersWarChapterTwo(t, g, a.ID, b.ID)
	// The requirement is an effect of the resolved chapter: it outlives the Saga.
	b27Kill(g, saga)

	// A's turn: A's creature must attack B, and may not attack me or C.
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err == nil {
		t.Fatal("A passed with its creature at home; it must attack B")
	}
	for _, wrong := range []uuid.UUID{me.ID, c.ID} {
		if err := g.DeclareAttacker(creatureA, wrong); err == nil {
			t.Fatalf("A's creature attacked %v instead of the other chosen player", wrong)
		}
	}
	if err := g.DeclareAttacker(creatureA, b.ID); err != nil {
		t.Fatalf("A's creature could not attack B: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with A's creature attacking B: %v", err)
	}
	// B's turn: B's creature must attack A.
	advanceToDeclareAttackersOf(t, g, 2)
	if err := g.DeclareAttacker(creatureB, me.ID); err == nil {
		t.Fatal("B's creature attacked someone other than A")
	}
	if err := g.DeclareAttacker(creatureB, a.ID); err != nil {
		t.Fatalf("B's creature could not attack A: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with B's creature attacking A: %v", err)
	}
	// C was not chosen: it may stay home.
	advanceToDeclareAttackersOf(t, g, 3)
	_ = creatureC
	if err := g.PassPriority(); err != nil {
		t.Fatalf("an unchosen player's creature was forced to attack: %v", err)
	}
	// On my next turn the effect is over.
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("on my next turn: %v", err)
	}
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("after my next turn began, A's creature is still forced to attack: %v", err)
	}
}

func TestS82BroB20BrothersWarChapterTwoReachesCreaturesThatEnterLater(t *testing.T) {
	g := newCatalogGame(t)
	a, b := g.Seats[1], g.Seats[2]
	s82BroBrothersWarChapterTwo(t, g, a.ID, b.ID)
	late := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Late Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: a.ID, Controller: a.ID, Keywords: []string{"haste"}})
	advanceToDeclareAttackersOf(t, g, 1)
	if err := g.PassPriority(); err == nil {
		t.Fatal("a creature that entered after the chapter resolved was not forced to attack")
	}
	if err := g.DeclareAttacker(late, b.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestS82BroB20BrothersWarChapterTwoCanChooseYou(t *testing.T) {
	g := newCatalogGame(t)
	me, a := g.Seats[0], g.Seats[1]
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, Keywords: []string{"haste"}})
	s82BroBrothersWarChapterTwo(t, g, me.ID, a.ID)
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.PassPriority(); err == nil {
		t.Fatal("my creature was not forced to attack")
	}
	if err := g.DeclareAttacker(mine, g.Seats[2].ID); err == nil {
		t.Fatal("my creature attacked a player other than the other chosen player")
	}
	if err := g.DeclareAttacker(mine, a.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
}

func TestS82BroB20BrothersWarChapterThreeDealsXToTwoDifferentTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Permanent(g, me.ID, "Relic One", "Artifact")
	b12Permanent(g, me.ID, "Relic Two", "Artifact")
	b12Permanent(g, me.ID, "Relic Three", "Artifact")
	b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 5, 5)
	// The Saga starts on chapter III.
	saga := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "The Brothers' War",
		TypeLine: "Enchantment — Saga", OracleID: s82BroBrothersWarOracle, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLore: 2}})
	advanceToPrecombatMainOf(t, g, 0)
	first := pickTargetPromptOpen(g)
	if first == nil {
		t.Fatal("no target prompt for chapter III")
	}
	if err := g.ResolvePickTarget(first.ID, first.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("first target: %v", err)
	}
	second := pickTargetPromptOpen(g)
	if second == nil {
		t.Fatal("no second target prompt")
	}
	if err := g.ResolvePickTarget(second.ID, second.Chooser, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err == nil {
		t.Fatal("the same player was accepted for \"any other target\"")
	}
	if err := g.ResolvePickTarget(second.ID, second.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("second target: %v", err)
	}
	life := opp.Life
	passPriorityAroundTable(t, g)
	// X is the artifacts you control as it resolves: three.
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	if got := layeredCard(t, g, bear); got.DamageMarked != 3 {
		t.Errorf("the Bear has %d damage, want 3", got.DamageMarked)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga should be sacrificed after chapter III")
	}
}
