package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b20_more_test.go — the rest of batch 20 (#5479): Jace, Kaya, Lukka,
// Nahiri, Nissa, Goliath Hatchery and Viral Spawning.

// s82CastWalker casts a planeswalker from hand for `cost`, strictly, with
// `life` Phyrexian symbols paid with life, and resolves it.
func s82CastWalker(t *testing.T, g *game.Game, name, oracle, typeLine, cost string, loyalty, life int, pool string) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, StartingLoyalty: loyalty,
		OracleID: oracle, Owner: me.ID, Controller: me.ID,
	})
	toMainForCost(t, g)
	floatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, PhyrexianLife: life}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

// --- compleated, on the hybrid and the plain Phyrexian symbols -----------------------

func TestS82CompleatedWalkersEnterWithTwoFewerPerSymbolPaidWithLife(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
		loyalty, life      int
		pool               string
		want               int
	}{
		{"Jace paid with mana", s82JaceOracle, "{2}{U}{U/P}", 5, 0, "CCUU", 5},
		{"Jace paid with life", s82JaceOracle, "{2}{U}{U/P}", 5, 1, "CCU", 3},
		{"Nahiri paid with {W}", s82NahiriOracle, "{1}{R}{R/W/P}{W}", 5, 0, "CRWW", 5},
		{"Nahiri paid with life", s82NahiriOracle, "{1}{R}{R/W/P}{W}", 5, 1, "CRW", 3},
		{"Lukka paid with life", s82LukkaOracle, "{2}{R}{R/G/P}{G}", 5, 1, "CCRG", 3},
		{"Nissa paid with life twice", s82NissaOracle, "{3}{G}{G}{G/P}{G/P}", 7, 2, "CCCGG", 3},
		{"Nissa paid with life once", s82NissaOracle, "{3}{G}{G}{G/P}{G/P}", 7, 1, "CCCGGG", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			id := s82CastWalker(t, g, tc.name, tc.oracle, "Legendary Planeswalker — Test", tc.cost, tc.loyalty, tc.life, tc.pool)
			if got := loyaltyOf(g, id); got != tc.want {
				t.Errorf("loyalty %d, want %d", got, tc.want)
			}
			if got := life - me.Life; got != 2*tc.life {
				t.Errorf("paid %d life, want %d", got, 2*tc.life)
			}
		})
	}
}

// --- Jace, the Perfected Mind ---------------------------------------------------------

func TestS82JacePlusOneShrinksPowerUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jace := s82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 4, 4)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{Targets: s82TargetCard(bear)})
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, bear); c.Effective().Power != 1 || c.Effective().Toughness != 4 {
		t.Errorf("Bear %d/%d, want 1/4", c.Effective().Power, c.Effective().Toughness)
	}
	powerUpPassTo(t, g, 1)
	if c, _ := battlefieldCard(g, bear); c.Effective().Power != 1 {
		t.Error("the effect ended on the opponent's own turn")
	}
	powerUpPassTo(t, g, 0)
	if c, _ := battlefieldCard(g, bear); c.Effective().Power != 4 {
		t.Errorf("Bear power %d after my next turn began, want 4", c.Effective().Power)
	}
}

func TestS82JaceMinusTwoDrawsOneUnlessAGraveyardHasTwentyCards(t *testing.T) {
	for _, tc := range []struct {
		name string
		pad  int
		draw int
	}{{"small graveyards", 0, 1}, {"an opponent's graveyard of 20 after the mill", 17, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			jace := s82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
			awbPadLibrary(me)
			awbPadLibrary(opp)
			for i := 0; i < tc.pad; i++ {
				opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
			}
			advanceToMain(t, g)
			hand, grave := me.Hand.Size(), opp.Graveyard.Size()
			s82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
			passPriorityAroundTable(t, g)
			if got := opp.Graveyard.Size() - grave; got != 3 {
				t.Errorf("milled %d, want 3", got)
			}
			if got := me.Hand.Size() - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
			if got := loyaltyOf(g, jace); got != 3 {
				t.Errorf("loyalty %d, want 3", got)
			}
		})
	}
}

func TestS82JaceMinusXMillsThreeTimesX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jace := s82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
	awbPadLibrary(opp)
	advanceToMain(t, g)
	grave := opp.Graveyard.Size()
	s82Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - grave; got != 6 {
		t.Errorf("milled %d, want 6", got)
	}
	if got := loyaltyOf(g, jace); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

// --- Kaya, Intangible Slayer ------------------------------------------------------------

func TestS82KayaPlusTwoDrainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := s82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	advanceToMain(t, g)
	mine := me.Life
	others := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		others[p.ID] = p.Life
	}
	s82Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			continue
		}
		if p.Life != others[p.ID]-3 {
			t.Errorf("opponent life %d, want %d", p.Life, others[p.ID]-3)
		}
	}
	if me.Life != mine+3 {
		t.Errorf("my life %d, want %d (gained once)", me.Life, mine+3)
	}
}

func TestS82KayaZeroDrawsTwoThenEachOpponentMayScry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := s82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	awbPadLibrary(me)
	awbPadLibrary(opp)
	advanceToMain(t, g)
	hand := me.Hand.Size()
	s82Activate(t, g, me.ID, kaya, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	c := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if c == nil {
		t.Fatal("the first opponent was not asked whether to scry")
	}
	if err := g.ResolveConfirm(c.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceConfirm, g.Seats[2].ID) == nil {
		t.Error("the next opponent in turn order was not asked")
	}
}

func TestS82KayaMinusThreeExilesAndMakesASpiritCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := s82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	ogre := b12Creature(g, opp.ID, "Big Ogre", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{Targets: s82TargetCard(ogre)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) || !g.Exile.Contains(ogre) {
		t.Fatal("the Ogre was not exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.IsToken() && c.Name == "Big Ogre" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy")
	}
	e := token.Effective()
	if token.Controller != me.ID {
		t.Error("the copy is not mine")
	}
	if e.Power != 1 || e.Toughness != 1 || len(e.Colors) != 1 || e.Colors[0] != "W" ||
		!token.HasSubtype("Ogre") || !token.HasSubtype("Spirit") || !game.HasKeyword(token, "flying") {
		t.Errorf("copy is %d/%d %v %v, want a 1/1 white Ogre Spirit with flying", e.Power, e.Toughness, e.Colors, e.Subtypes)
	}
}

func TestS82KayaMinusThreeOnAnAuraOnlyExiles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := s82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	aura := pushCatalogPermanent(g, opp.ID, "Pacifism-ish", "Enchantment — Aura", "", false)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{Targets: s82TargetCard(aura)})
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(aura) {
		t.Fatal("the Aura was not exiled")
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Errorf("a token %q was made from an Aura", c.Name)
		}
	}
}

func TestS82KayaIsHexproof(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := s82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	c, _ := battlefieldCard(g, kaya)
	if !game.HasKeyword(&c, "hexproof") {
		t.Error("no hexproof")
	}
}

// --- Lukka, Bound to Ruin ---------------------------------------------------------------

func TestS82LukkaMinusOneMakesAToxicBeast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lukka := s82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, lukka, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Phyrexian Beast" {
			continue
		}
		n++
		e := c.Effective()
		if !c.IsToken() || e.Power != 3 || e.Toughness != 3 || len(c.Colors) != 1 || c.Colors[0] != "G" ||
			!c.HasSubtype("Beast") || !c.HasSubtype("Phyrexian") || game.ToxicTotal(&c) != 1 {
			t.Errorf("Beast %d/%d %v %v toxic %d, want a green 3/3 Phyrexian Beast with toxic 1", e.Power, e.Toughness, c.Colors, e.Subtypes, game.ToxicTotal(&c))
		}
	}
	if n != 1 {
		t.Errorf("%d Beasts, want 1", n)
	}
}

func TestS82LukkaPlusOneAddsRedGreenOnlyForCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lukka := s82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, lukka, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool holds %d mana, want {R}{G}", got)
	}
	// A noncreature spell can't be paid with it.
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lava Spike", TypeLine: "Sorcery", ManaCost: "{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Error("restricted mana paid for a noncreature spell")
	}
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Grizzly Bears", TypeLine: "Creature — Bear", ManaCost: "{R}{G}", Colors: []string{"R", "G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the mana could not cast a creature spell: %v", err)
	}
}

func TestS82LukkaMinusFourDividesTheGreatestPowerAmongTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	lukka := s82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "Large", "Creature — Giant", 5, 5)
	a := b12Creature(g, opp.ID, "A", "Creature — Bear", 1, 6)
	b := b12Creature(g, opp.ID, "B", "Creature — Bear", 1, 6)
	advanceToMain(t, g)
	// X is 5, read as the ability is activated: splits that do not add up are refused.
	err := g.ActivateCatalogAbility(me.ID, lukka, 2, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 2, b: 2}})
	if err == nil {
		t.Fatal("a split of 4 was accepted for X = 5")
	}
	s82Activate(t, g, me.ID, lukka, 2, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 3, b: 2}})
	passPriorityAroundTable(t, g)
	ca, _ := battlefieldCard(g, a)
	cb, _ := battlefieldCard(g, b)
	if ca.DamageMarked != 3 || cb.DamageMarked != 2 {
		t.Errorf("damage %d and %d, want 3 and 2", ca.DamageMarked, cb.DamageMarked)
	}
	if got := loyaltyOf(g, lukka); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// --- Nahiri, the Unforgiving ------------------------------------------------------------

func TestS82NahiriLootsOnThePlusOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nahiri := s82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	awbPadLibrary(me)
	handCardForTest(me, "Keep", "Sorcery", "")
	handCardForTest(me, "Keep Two", "Sorcery", "")
	advanceToMain(t, g)
	hand := me.Hand.Size()
	s82Activate(t, g, me.ID, nahiri, 1, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerChooseCardsOrDiscard(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after discarding and drawing", hand, me.Hand.Size())
	}
	if got := loyaltyOf(g, nahiri); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82NahiriPlusOneForcesAnAttackOnAPlayerNotAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	nahiri := s82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	walker := s82Walker(g, third.ID, "Other Walker", "", 4)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, nahiri, 0, game.ActivateAbilityParams{Targets: s82TargetCard(bear)})
	passPriorityAroundTable(t, g)
	powerUpPassTo(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	var must map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { must = g.MustAttackForEffect() })
	got := must[bear]
	if len(got) == 0 {
		t.Fatal("the Bear is not required to attack")
	}
	for _, id := range got {
		if id == walker {
			t.Error("attacking a planeswalker would satisfy the requirement")
		}
	}
	if err := g.DeclareAttacker(bear, walker); err == nil {
		// Declaring at a planeswalker leaves a player attack possible, so the lock-in refuses it.
		if perr := g.PassPriority(); perr == nil {
			t.Error("attacking only a planeswalker was allowed")
		}
	}
}

func TestS82NahiriZeroCopiesALowerCostCreatureWithHasteAndExilesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nahiri := s82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	cheap := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Giant", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	sword := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sword, Name: "Cheap Blade", TypeLine: "Artifact — Equipment", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	// Mana value 5 is not less than loyalty 5.
	if err := g.ActivateCatalogAbility(me.ID, nahiri, 2, game.ActivateAbilityParams{Strict: true, Targets: s82TargetCard(dear)}); err == nil {
		t.Fatal("a card with mana value equal to Nahiri's loyalty was a legal target")
	}
	s82Activate(t, g, me.ID, nahiri, 2, game.ActivateAbilityParams{Targets: s82TargetCard(cheap)})
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(cheap) || !g.Exile.Contains(cheap) {
		t.Fatal("the Elf was not exiled from the graveyard")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.IsToken() && c.Name == "Cheap Elf" {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy")
	}
	if !game.HasKeyword(token, "haste") {
		t.Error("the token lacks haste")
	}
	tokenID := token.InstanceID
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tokenID) {
		t.Error("the token was not exiled at the beginning of the end step")
	}
	_ = sword
}

// --- Nissa, Ascended Animist ------------------------------------------------------------

func TestS82NissaPlusOneMakesAHorrorAsBigAsHerLoyalty(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := s82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 5)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, nissa, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	var horror *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Phyrexian Horror" {
			horror = c
		}
	}
	if horror == nil {
		t.Fatal("no Horror")
	}
	e := horror.Effective()
	if e.Power != 6 || e.Toughness != 6 || len(horror.Colors) != 1 || horror.Colors[0] != "G" || !horror.HasSubtype("Horror") || !horror.HasSubtype("Phyrexian") {
		t.Errorf("Horror %d/%d %v %v, want a green 6/6 Phyrexian Horror (loyalty after the +1)", e.Power, e.Toughness, horror.Colors, e.Subtypes)
	}
}

func TestS82NissaMinusOneDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nissa := s82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 7)
	rock := pushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true, Targets: s82TargetCard(bear)}); err == nil {
		t.Fatal("a plain creature was a legal target")
	}
	s82Activate(t, g, me.ID, nissa, 1, game.ActivateAbilityParams{Targets: s82TargetCard(rock)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestS82NissaMinusSevenPumpsByForestsAndGrantsTramplyOnlyToTheCreaturesYouHad(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := s82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 7)
	for i := 0; i < 3; i++ {
		pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	pushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	s82Activate(t, g, me.ID, nissa, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	c, _ := battlefieldCard(g, bear)
	if e := c.Effective(); e.Power != 5 || e.Toughness != 5 || !game.HasKeyword(&c, "trample") {
		t.Errorf("Bear %d/%d trample=%v, want 5/5 with trample (three Forests)", e.Power, e.Toughness, game.HasKeyword(&c, "trample"))
	}
	l, _ := battlefieldCard(g, late)
	if e := l.Effective(); e.Power != 2 || game.HasKeyword(&l, "trample") {
		t.Errorf("a creature that arrived later got %d/%d trample=%v", e.Power, e.Toughness, game.HasKeyword(&l, "trample"))
	}
}

// --- Goliath Hatchery -------------------------------------------------------------------

func TestS82GoliathHatcheryMakesTwoBeastsWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Goliath Hatchery", TypeLine: "Enchantment", ManaCost: "{4}{G}{G}", OracleID: s82HatcheryOracle, Owner: me.ID, Controller: me.ID})
	toMainForCost(t, g)
	floatForTest(g, me, "CCCCGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Phyrexian Beast" && c.IsToken() && game.ToxicTotal(&c) == 1 {
			n++
		}
	}
	if n != 2 {
		t.Errorf("%d toxic Beast tokens, want 2", n)
	}
}

func TestS82GoliathHatcheryCorruptedDrawsTheChosenCreaturesToxic(t *testing.T) {
	for _, poison := range []int{2, 3} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Goliath Hatchery", "Enchantment", s82HatcheryOracle, false)
		toxic := pushCatalogPermanent(g, me.ID, "Toxic Thing", "Creature — Phyrexian", "", false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == toxic {
					g.Battlefield.Cards[i].Keywords = []string{"toxic 2", "toxic 1"}
				}
			}
		})
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		awbPadLibrary(me)
		g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, poison) })
		hand := me.Hand.Size()
		powerUpPassTo(t, g, 1)
		hand = me.Hand.Size()
		for i := 0; i < 8 && g.Turn.ActiveSeat != 0; i++ {
			if err := g.PassTurn(); err != nil {
				t.Fatal(err)
			}
		}
		passPriorityAroundTable(t, g)
		p := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
		if poison < 3 {
			if p != nil {
				t.Errorf("poison %d: the trigger fired below three poison counters", poison)
			}
			continue
		}
		if p == nil {
			t.Fatalf("poison %d: no creature choice", poison)
		}
		if len(p.ChooseCards) != 2 {
			t.Errorf("offered %v, want my two creatures (%v, %v)", p.ChooseCards, toxic, bear)
		}
		if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{toxic}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		// One card for the draw step, three for toxic 2 + toxic 1.
		if got := me.Hand.Size() - hand; got != 3 {
			t.Errorf("hand grew by %d, want 3 (a total toxic value of 3)", got)
		}
	}
}

// --- Viral Spawning ---------------------------------------------------------------------

func TestS82ViralSpawningMakesABeastAndFlashbackNeedsCorrupted(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Viral Spawning", TypeLine: "Sorcery", ManaCost: "{2}{G}", Colors: []string{"G"}, OracleID: s82SpawningOracle, Owner: me.ID, Controller: me.ID})
	toMainForCost(t, g)
	floatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	beasts := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Phyrexian Beast" && c.IsToken() && game.ToxicTotal(&c) == 1 {
				n++
			}
		}
		return n
	}
	if beasts() != 1 || !me.Graveyard.Contains(id) {
		t.Fatalf("beasts %d, in graveyard %v", beasts(), me.Graveyard.Contains(id))
	}
	// Not corrupted: no flashback.
	floatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashback was offered without corrupted")
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3) })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback refused while corrupted: %v", err)
	}
	passPriorityAroundTable(t, g)
	if beasts() != 2 {
		t.Errorf("%d Beasts after the flashback, want 2", beasts())
	}
	if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}

// answerChooseCardsOrDiscard answers the open hand-discard prompt with the
// first card offered.
func answerChooseCardsOrDiscard(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == p.ID {
			t.Logf("prompt kind %v", c.Kind)
			opts := c.DiscardOptions
			if len(opts) == 0 {
				opts = c.ChooseCards
			}
			if len(opts) == 0 {
				opts = []uuid.UUID{p.Hand.Cards[0].InstanceID}
			}
			if err := g.ResolvePendingChoice(c.ID, p.ID, opts[:1]); err != nil {
				t.Fatalf("discard: %v", err)
			}
			return
		}
	}
	t.Fatal("no discard prompt is open")
}
