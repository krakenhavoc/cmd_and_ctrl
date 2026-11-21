package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			life := me.Life
			id := effectstest.S82CastWalker(t, g, tc.name, tc.oracle, "Legendary Planeswalker — Test", tc.cost, tc.loyalty, tc.life, tc.pool)
			if got := effectstest.LoyaltyOf(g, id); got != tc.want {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jace := effectstest.S82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 4, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, jace, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Effective().Power != 1 || c.Effective().Toughness != 4 {
		t.Errorf("Bear %d/%d, want 1/4", c.Effective().Power, c.Effective().Toughness)
	}
	effectstest.PowerUpPassTo(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Effective().Power != 1 {
		t.Error("the effect ended on the opponent's own turn")
	}
	effectstest.PowerUpPassTo(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, bear); c.Effective().Power != 4 {
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
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			jace := effectstest.S82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
			effectstest.AwbPadLibrary(me)
			effectstest.AwbPadLibrary(opp)
			for i := 0; i < tc.pad; i++ {
				opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
			}
			effectstest.AdvanceToMain(t, g)
			hand, grave := me.Hand.Size(), opp.Graveyard.Size()
			effectstest.S82Activate(t, g, me.ID, jace, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := opp.Graveyard.Size() - grave; got != 3 {
				t.Errorf("milled %d, want 3", got)
			}
			if got := me.Hand.Size() - hand; got != tc.draw {
				t.Errorf("drew %d, want %d", got, tc.draw)
			}
			if got := effectstest.LoyaltyOf(g, jace); got != 3 {
				t.Errorf("loyalty %d, want 3", got)
			}
		})
	}
}

func TestS82JaceMinusXMillsThreeTimesX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	jace := effectstest.S82Walker(g, me.ID, "Jace, the Perfected Mind", s82JaceOracle, 5)
	effectstest.AwbPadLibrary(opp)
	effectstest.AdvanceToMain(t, g)
	grave := opp.Graveyard.Size()
	effectstest.S82Activate(t, g, me.ID, jace, 2, game.ActivateAbilityParams{XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - grave; got != 6 {
		t.Errorf("milled %d, want 6", got)
	}
	if got := effectstest.LoyaltyOf(g, jace); got != 3 {
		t.Errorf("loyalty %d, want 3", got)
	}
}

// --- Kaya, Intangible Slayer ------------------------------------------------------------

func TestS82KayaPlusTwoDrainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	effectstest.AdvanceToMain(t, g)
	mine := me.Life
	others := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		others[p.ID] = p.Life
	}
	effectstest.S82Activate(t, g, me.ID, kaya, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	effectstest.AwbPadLibrary(me)
	effectstest.AwbPadLibrary(opp)
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, kaya, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, opp.ID)
	if c == nil {
		t.Fatal("the first opponent was not asked whether to scry")
	}
	if err := g.ResolveConfirm(c.ID, opp.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, g.Seats[2].ID) == nil {
		t.Error("the next opponent in turn order was not asked")
	}
}

func TestS82KayaMinusThreeExilesAndMakesASpiritCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	ogre := effectstest.B12Creature(g, opp.ID, "Big Ogre", "Creature — Ogre", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(ogre)})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Pacifism-ish", "Enchantment — Aura", "", false)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, kaya, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(aura)})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kaya := effectstest.S82Walker(g, me.ID, "Kaya, Intangible Slayer", s82KayaOracle, 6)
	c, _ := effectstest.BattlefieldCard(g, kaya)
	if !game.HasKeyword(&c, "hexproof") {
		t.Error("no hexproof")
	}
}

// --- Lukka, Bound to Ruin ---------------------------------------------------------------

func TestS82LukkaMinusOneMakesAToxicBeast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lukka := effectstest.S82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, lukka, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lukka := effectstest.S82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, lukka, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	lukka := effectstest.S82Walker(g, me.ID, "Lukka, Bound to Ruin", s82LukkaOracle, 5)
	effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Large", "Creature — Giant", 5, 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 1, 6)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 1, 6)
	effectstest.AdvanceToMain(t, g)
	// X is 5, read as the ability is activated: splits that do not add up are refused.
	err := g.ActivateCatalogAbility(me.ID, lukka, 2, game.ActivateAbilityParams{Strict: true,
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 2, b: 2}})
	if err == nil {
		t.Fatal("a split of 4 was accepted for X = 5")
	}
	effectstest.S82Activate(t, g, me.ID, lukka, 2, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}},
		Distribution: map[uuid.UUID]int{a: 3, b: 2}})
	effectstest.PassPriorityAroundTable(t, g)
	ca, _ := effectstest.BattlefieldCard(g, a)
	cb, _ := effectstest.BattlefieldCard(g, b)
	if ca.DamageMarked != 3 || cb.DamageMarked != 2 {
		t.Errorf("damage %d and %d, want 3 and 2", ca.DamageMarked, cb.DamageMarked)
	}
	if got := effectstest.LoyaltyOf(g, lukka); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// --- Nahiri, the Unforgiving ------------------------------------------------------------

func TestS82NahiriLootsOnThePlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nahiri := effectstest.S82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	effectstest.AwbPadLibrary(me)
	effectstest.HandCardForTest(me, "Keep", "Sorcery", "")
	effectstest.HandCardForTest(me, "Keep Two", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, nahiri, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCardsOrDiscard(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want the same size after discarding and drawing", hand, me.Hand.Size())
	}
	if got := effectstest.LoyaltyOf(g, nahiri); got != 6 {
		t.Errorf("loyalty %d, want 6", got)
	}
}

func TestS82NahiriPlusOneForcesAnAttackOnAPlayerNotAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	nahiri := effectstest.S82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	walker := effectstest.S82Walker(g, third.ID, "Other Walker", "", 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, nahiri, 0, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PowerUpPassTo(t, g, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nahiri := effectstest.S82Walker(g, me.ID, "Nahiri, the Unforgiving", s82NahiriOracle, 5)
	cheap := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap Elf", TypeLine: "Creature — Elf", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	dear := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dear, Name: "Dear Giant", TypeLine: "Creature — Giant", ManaCost: "{4}{G}", Power: 5, Toughness: 5, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	sword := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sword, Name: "Cheap Blade", TypeLine: "Artifact — Equipment", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	// Mana value 5 is not less than loyalty 5.
	if err := g.ActivateCatalogAbility(me.ID, nahiri, 2, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(dear)}); err == nil {
		t.Fatal("a card with mana value equal to Nahiri's loyalty was a legal target")
	}
	effectstest.S82Activate(t, g, me.ID, nahiri, 2, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(cheap)})
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tokenID) {
		t.Error("the token was not exiled at the beginning of the end step")
	}
	_ = sword
}

// --- Nissa, Ascended Animist ------------------------------------------------------------

func TestS82NissaPlusOneMakesAHorrorAsBigAsHerLoyalty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := effectstest.S82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, nissa, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	nissa := effectstest.S82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 7)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, nissa, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.S82TargetCard(bear)}); err == nil {
		t.Fatal("a plain creature was a legal target")
	}
	effectstest.S82Activate(t, g, me.ID, nissa, 1, game.ActivateAbilityParams{Targets: effectstest.S82TargetCard(rock)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

func TestS82NissaMinusSevenPumpsByForestsAndGrantsTramplyOnlyToTheCreaturesYouHad(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nissa := effectstest.S82Walker(g, me.ID, "Nissa, Ascended Animist", s82NissaOracle, 7)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, nissa, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	c, _ := effectstest.BattlefieldCard(g, bear)
	if e := c.Effective(); e.Power != 5 || e.Toughness != 5 || !game.HasKeyword(&c, "trample") {
		t.Errorf("Bear %d/%d trample=%v, want 5/5 with trample (three Forests)", e.Power, e.Toughness, game.HasKeyword(&c, "trample"))
	}
	l, _ := effectstest.BattlefieldCard(g, late)
	if e := l.Effective(); e.Power != 2 || game.HasKeyword(&l, "trample") {
		t.Errorf("a creature that arrived later got %d/%d trample=%v", e.Power, e.Toughness, game.HasKeyword(&l, "trample"))
	}
}

// --- Goliath Hatchery -------------------------------------------------------------------

func TestS82GoliathHatcheryMakesTwoBeastsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Goliath Hatchery", TypeLine: "Enchantment", ManaCost: "{4}{G}{G}", OracleID: s82HatcheryOracle, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "CCCCGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Goliath Hatchery", "Enchantment", s82HatcheryOracle, false)
		toxic := effectstest.PushCatalogPermanent(g, me.ID, "Toxic Thing", "Creature — Phyrexian", "", false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == toxic {
					g.Battlefield.Cards[i].Keywords = []string{"toxic 2", "toxic 1"}
				}
			}
		})
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AwbPadLibrary(me)
		g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, poison) })
		effectstest.PowerUpPassTo(t, g, 1)
		hand := me.Hand.Size()
		for i := 0; i < 8 && g.Turn.ActiveSeat != 0; i++ {
			if err := g.PassTurn(); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
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
		effectstest.PassPriorityAroundTable(t, g)
		// One card for the draw step, three for toxic 2 + toxic 1.
		if got := me.Hand.Size() - hand; got != 3 {
			t.Errorf("hand grew by %d, want 3 (a total toxic value of 3)", got)
		}
	}
}

// --- Viral Spawning ---------------------------------------------------------------------

func TestS82ViralSpawningMakesABeastAndFlashbackNeedsCorrupted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Viral Spawning", TypeLine: "Sorcery", ManaCost: "{2}{G}", Colors: []string{"G"}, OracleID: s82SpawningOracle, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	effectstest.FloatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	effectstest.FloatForTest(g, me, "CCG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err == nil {
		t.Fatal("flashback was offered without corrupted")
	}
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, game.CounterPoison, 3) })
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback refused while corrupted: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if beasts() != 2 {
		t.Errorf("%d Beasts after the flashback, want 2", beasts())
	}
	if me.Graveyard.Contains(id) || !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}
