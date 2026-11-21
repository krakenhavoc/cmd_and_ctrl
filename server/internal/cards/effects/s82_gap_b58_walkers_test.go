package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func s82g58Row(oracle, front, frontType, frontCost, back, backType, fp, ft, loyalty string, colors []string) cards.Card {
	row := transformRow(oracle, front, frontType, frontCost, back, backType, "", "", colors)
	row.CardFaces[0].Power, row.CardFaces[0].Toughness = fp, ft
	row.CardFaces[1].Loyalty = loyalty
	return row
}

func s82g58TamiyoRow() cards.Card {
	return s82g58Row(s82g58Tamiyo, "Tamiyo, Inquisitive Student", "Legendary Creature — Moonfolk Wizard", "{U}",
		"Tamiyo, Seasoned Scholar", "Legendary Planeswalker — Tamiyo", "0", "3", "2", []string{"U"})
}

func s82g58RalRow() cards.Card {
	return s82g58Row(s82g58Ral, "Ral, Monsoon Mage", "Legendary Creature — Human Wizard", "{1}{R}",
		"Ral, Leyline Prodigy", "Legendary Planeswalker — Ral", "1", "3", "2", []string{"R"})
}

func s82g58GristRow() cards.Card {
	return s82g58Row(s82g58Grist, "Grist, Voracious Larva", "Legendary Creature — Insect", "{G}",
		"Grist, the Plague Swarm", "Legendary Planeswalker — Grist", "1", "2", "3", []string{"G"})
}

func s82g58AjaniRow() cards.Card {
	return s82g58Row(s82g58Ajani, "Ajani, Nacatl Pariah", "Legendary Creature — Cat Warrior", "{1}{W}",
		"Ajani, Nacatl Avenger", "Legendary Planeswalker — Ajani", "1", "2", "3", []string{"W"})
}

// s82g58Flipped is the transformed permanent of `oracle` that `p` controls.
func s82g58Flipped(g *game.Game, p *game.Player, oracle string) (uuid.UUID, bool) {
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == oracle && c.Controller == p.ID && TransformedPermanent(c) {
				id = c.InstanceID
			}
		}
	})
	return id, id != uuid.Nil
}

// s82g58Walker puts a front face onto the battlefield and turns it over,
// returning the planeswalker with `loyalty` counters on it.
func s82g58Walker(t *testing.T, g *game.Game, row cards.Card, p *game.Player, loyalty int) uuid.UUID {
	t.Helper()
	front := importToBattlefield(t, g, row, p)
	var err error
	g.WithWriteLock(func() { err = g.ExileAndReturnTransformedForEffect(front, uuid.Nil) })
	if err != nil {
		t.Fatal(err)
	}
	id, ok := s82g58Flipped(g, p, row.OracleID.String())
	if !ok {
		t.Fatal("the front face did not come back transformed")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).Counters[game.CounterLoyalty] = loyalty })
	return id
}

func s82g58Library(p *game.Player, n int, typeLine string) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, pushLibraryCardForTest(p, game.Card{Name: "Filler", TypeLine: typeLine, Owner: p.ID, Controller: p.ID}))
	}
	return ids
}

func s82g58Clear(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, id).SummonedThisTurn = false })
}

// --- Tamiyo ----------------------------------------------------------

func TestS82GapB58TamiyoInvestigatesWhenAttackingAndFlipsOnTheThirdDraw(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	s82g58Library(me, 10, "Sorcery")
	tamiyo := importToBattlefield(t, g, s82g58TamiyoRow(), me)
	assertKeywords(t, g, tamiyo, "flying")
	s82g58Clear(g, tamiyo)
	declareAttack(t, g, opp.ID, tamiyo)
	passPriorityAroundTable(t, g)
	if n := countNamed(g, "Clue"); n != 1 {
		t.Fatalf("attacking investigated %d times, want 1", n)
	}
}

func TestS82GapB58TamiyoFlipsWhenYouDrawYourThirdCardInATurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	s82g58Library(me, 10, "Sorcery")
	importToBattlefield(t, g, s82g58TamiyoRow(), me)
	drawn := g.TurnTallyFor(me.ID).CardsDrawn
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 2-drawn) })
	passPriorityAroundTable(t, g)
	if _, ok := s82g58Flipped(g, me, s82g58Tamiyo); ok {
		t.Fatal("flipped on the second card drawn this turn")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	passPriorityAroundTable(t, g)
	walker, ok := s82g58Flipped(g, me, s82g58Tamiyo)
	if !ok {
		t.Fatal("Tamiyo did not flip on the third card")
	}
	c := e2Card(t, g, walker)
	if c.Name != "Tamiyo, Seasoned Scholar" || loyaltyOf(g, walker) != 2 {
		t.Errorf("back face %q with loyalty %d, want Seasoned Scholar with 2", c.Name, loyaltyOf(g, walker))
	}
}

func TestS82GapB58TamiyoPlusTwoShrinksCreaturesAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58TamiyoRow(), me, 2)
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if loyaltyOf(g, walker) != 4 {
		t.Fatalf("loyalty %d after +2, want 4", loyaltyOf(g, walker))
	}
	theirs := p1Creature(g, opp.ID, 3, 3)
	advanceToMainOf(t, g, 1)
	s82g58Clear(g, theirs)
	declareAttack(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, theirs); got != 2 {
		t.Errorf("the attacker has power %d, want 2 (-1/-0)", got)
	}
}

func TestS82GapB58TamiyoMinusThreeReturnsASpellAndAddsManaForAGreenOne(t *testing.T) {
	for _, green := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		walker := s82g58Walker(t, g, s82g58TamiyoRow(), me, 5)
		colors := []string(nil)
		if green {
			colors = []string{"G"}
		}
		spell := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Instant", Colors: colors, Owner: me.ID, Controller: me.ID})
		creature := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(creature), Strict: true}); err == nil {
			t.Fatal("a creature card is not an instant or sorcery card")
		}
		b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(spell)})
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(spell) {
			t.Fatalf("green=%v: the spell did not return to hand", green)
		}
		if loyaltyOf(g, walker) != 2 {
			t.Errorf("loyalty %d after -3, want 2", loyaltyOf(g, walker))
		}
		if green {
			b10ResolveAllManaPicks(t, g, me.ID, "G")
		}
		if got := me.ManaPool.Total(); (got == 1) != green {
			t.Errorf("green=%v: pool total %d", green, got)
		}
	}
}

func TestS82GapB58TamiyoMinusSevenDrawsHalfTheLibraryRoundedUpAndGivesAnEmblem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58TamiyoRow(), me, 8)
	s82g58Library(me, 7, "Sorcery")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 4 {
		t.Errorf("drew %d of a 7-card library, want 4 (half, rounded up)", got)
	}
	if len(emblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(emblemsOf(g, me.ID)))
	}
	var max int
	g.ReadSnapshot(func() { max = g.EffectiveMaxHandSizeLocked(me) })
	if max != game.NoMaxHandSize {
		t.Errorf("maximum hand size = %d, want none", max)
	}
}

// --- Ral -------------------------------------------------------------

// s82g58RalFlip resolves the coin Ral's trigger flipped. When the flip is won
// it answers the "you may exile Ral" prompt with `exile`. Reports the flip.
func s82g58RalFlip(t *testing.T, g *game.Game, me *game.Player, exile bool) (won bool) {
	t.Helper()
	var flip *game.PendingChoice
	for i := 0; i < 8 && flip == nil; i++ {
		if flip = latestChoiceOfKind(g, game.PendingChoiceCoinCall); flip == nil {
			passPriorityAroundTable(t, g)
		}
	}
	if flip == nil {
		t.Fatal("casting an instant on my turn asked for no coin flip")
	}
	if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
		t.Fatalf("coin: %v", err)
	}
	flips := randomEvents(g, game.EventFlipCoin)
	won = len(flips) > 0 && flips[len(flips)-1].Won
	passPriorityAroundTable(t, g)
	if won {
		if latestConfirmFor(g, me.ID) == nil {
			t.Fatal("a won flip did not ask whether to exile Ral")
		}
		answerMayChoice(t, g, me.ID, exile)
		passPriorityAroundTable(t, g)
	}
	return won
}

func s82g58CastSpell(t *testing.T, g *game.Game, cost string) uuid.UUID {
	t.Helper()
	return p1g2Cast(t, g, "Shock", "Instant", "", cost, game.CastSpellParams{})
}

func TestS82GapB58RalMonsoonMageReducesInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	importToBattlefield(t, g, s82g58RalRow(), me)
	b06AddMana(me, "R")
	s82g58CastSpell(t, g, "{1}{R}") // {R} alone pays for it
	if me.ManaPool.Total() != 0 {
		t.Errorf("the reduced spell left %+v in the pool", me.ManaPool)
	}
	// A creature spell is not reduced.
	b06AddMana(me, "G")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell cost {1} less")
	}
}

func TestS82GapB58RalMonsoonMageLosingTheFlipDealsOneDamageToYou(t *testing.T) {
	for i := 0; i < 64; i++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		importToBattlefield(t, g, s82g58RalRow(), me)
		life := lifeOf(g, me.ID)
		s82g58CastSpell(t, g, "")
		if won := s82g58RalFlip(t, g, me, false); won {
			continue
		}
		if got := lifeOf(g, me.ID); got != life-1 {
			t.Fatalf("a lost flip: life %d -> %d, want a loss of 1", life, got)
		}
		return
	}
	t.Fatal("64 flips were all won")
}

func TestS82GapB58RalMonsoonMageWinningTheFlipLetsYouTransformIt(t *testing.T) {
	sawDecline, sawFlip := false, false
	for i := 0; i < 80 && !(sawDecline && sawFlip); i++ {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		importToBattlefield(t, g, s82g58RalRow(), me)
		exile := !sawFlip
		s82g58CastSpell(t, g, "")
		if !s82g58RalFlip(t, g, me, exile) {
			continue
		}
		walker, flipped := s82g58Flipped(g, me, s82g58Ral)
		if exile != flipped {
			t.Fatalf("exile=%v but transformed=%v", exile, flipped)
		}
		if flipped {
			sawFlip = true
			// printed 2 loyalty plus one for the instant cast this turn
			if got := loyaltyOf(g, walker); got != 3 {
				t.Errorf("Ral entered with %d loyalty, want 3 (2 + 1 instant cast this turn)", got)
			}
			if e2Card(t, g, walker).Name != "Ral, Leyline Prodigy" {
				t.Error("wrong back face")
			}
		} else {
			sawDecline = true
		}
	}
	if !sawFlip || !sawDecline {
		t.Fatalf("never saw both a won flip that exiles (%v) and one that declines (%v)", sawFlip, sawDecline)
	}
}

func TestS82GapB58RalLeylineProdigyPlusOneReducesInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58RalRow(), me, 3)
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	b06AddMana(me, "R")
	s82g58CastSpell(t, g, "{1}{R}")
	if me.ManaPool.Total() != 0 {
		t.Errorf("the +1 reduction was not applied: pool %+v", me.ManaPool)
	}
}

func TestS82GapB58RalLeylineProdigyMinusTwoDividesDamageAndDrawsOnlyWithBlue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58RalRow(), me, 4)
	ogre := p1Creature(g, opp.ID, 3, 3)
	s82g58Library(me, 3, "Sorcery")
	hand, life := me.Hand.Size(), lifeOf(g, opp.ID)
	err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}, {Kind: game.TargetCard, ID: ogre}},
		Distribution: map[uuid.UUID]int{opp.ID: 1, ogre: 1},
	})
	if err != nil {
		t.Fatalf("-2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := lifeOf(g, opp.ID); got != life-1 {
		t.Errorf("opponent life %d -> %d, want 1 damage", life, got)
	}
	if me.Hand.Size() != hand {
		t.Error("drew a card with no other blue permanent")
	}
}

func TestS82GapB58RalLeylineProdigyMinusTwoDrawsWithAnotherBluePermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58RalRow(), me, 4)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Blue Thing", TypeLine: "Creature — Merfolk",
		Colors: []string{"U"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	s82g58Library(me, 3, "Sorcery")
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		Distribution: map[uuid.UUID]int{opp.ID: 2},
	})
	passPriorityAroundTable(t, g)
	if me.Hand.Size()-hand != 1 {
		t.Errorf("drew %d cards, want 1", me.Hand.Size()-hand)
	}
}

func TestS82GapB58RalLeylineProdigyMinusEightExilesEightAndCastsSpellsFree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58RalRow(), me, 8)
	me.Library.Cards = nil
	var order []uuid.UUID
	for i := 0; i < 10; i++ {
		typeLine, cost := "Instant", "{R}{R}"
		if i%2 == 1 {
			typeLine, cost = "Creature — Bear", "{1}{G}"
		}
		order = append(order, pushLibraryCardForTest(me, game.Card{Name: "Card", TypeLine: typeLine, ManaCost: cost, Owner: me.ID, Controller: me.ID}))
	}
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range order {
		if g.Exile.Contains(id) {
			exiled++
		}
	}
	if exiled != 8 {
		t.Fatalf("exiled %d cards, want 8", exiled)
	}
	// the top of the library is the last pushed: find an exiled instant and creature.
	var instant, bear uuid.UUID
	for i, id := range order {
		if !g.Exile.Contains(id) {
			continue
		}
		if i%2 == 0 {
			instant = id
		} else {
			bear = id
		}
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a creature card was cast from the exile")
	}
	if err := g.CastSpell(me.ID, instant, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("casting an exiled instant without paying: %v", err)
	}
}

// --- Grist -----------------------------------------------------------

func s82g58GraveCreature(p *game.Player, name, typeLine string, colors []string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Colors: colors, Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID, KnownBy: map[uuid.UUID]bool{p.ID: true}})
	return id
}

func s82g58GristPayAnswer(t *testing.T, g *game.Game, me *game.Player, pay bool) {
	t.Helper()
	for i := 0; i < 4 && latestChoiceOfKind(g, game.PendingChoicePayUnless) == nil; i++ {
		passPriorityAroundTable(t, g)
	}
	c := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatal("Grist offered no payment")
	}
	if pay {
		b06AddMana(me, "G")
	}
	if err := g.ResolvePayUnless(c.ID, me.ID, pay); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82GapB58GristTransformsWhenACreatureReturnsFromYourGraveyardAndYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		grist := importToBattlefield(t, g, s82g58GristRow(), me)
		assertKeywords(t, g, grist, "deathtouch")
		bear := s82g58GraveCreature(me, "Bear", "Creature — Bear", []string{"G"})
		g.WithWriteLock(func() { _ = g.ReturnFromGraveyardForEffect(bear, game.ZoneBattlefield) })
		s82g58GristPayAnswer(t, g, me, pay)
		walker, flipped := s82g58Flipped(g, me, s82g58Grist)
		if flipped != pay {
			t.Fatalf("pay=%v but transformed=%v", pay, flipped)
		}
		if pay {
			if e2Card(t, g, walker).Name != "Grist, the Plague Swarm" || loyaltyOf(g, walker) != 3 {
				t.Errorf("wrong back face or loyalty %d", loyaltyOf(g, walker))
			}
			if me.ManaPool.Total() != 0 {
				t.Error("the {G} was not spent")
			}
		}
	}
}

func TestS82GapB58GristIgnoresCreaturesThatDidNotComeFromYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	importToBattlefield(t, g, s82g58GristRow(), me)
	// from the hand
	castAndResolveCreature(t, g, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	// from an opponent's graveyard, under my control... it entered from their graveyard
	theirs := s82g58GraveCreature(opp, "Their Bear", "Creature — Bear", nil)
	g.WithWriteLock(func() { _ = g.ReturnFromGraveyardUnderControlForEffect(theirs, game.ZoneBattlefield, me.ID) })
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("Grist offered to transform for a creature that did not come from your graveyard")
	}
	if _, ok := s82g58Flipped(g, me, s82g58Grist); ok {
		t.Error("Grist transformed")
	}
}

func TestS82GapB58GristTriggersOnACreatureCastFromYourGraveyard(t *testing.T) {
	g, me, _ := muldrothaTable(t)
	importToBattlefield(t, g, s82g58GristRow(), me)
	bear := ptGraveCard(me, "Bear", "Creature — Bear")
	if err := muldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	passPriorityAroundTable(t, g)
	s82g58GristPayAnswer(t, g, me, true)
	if _, ok := s82g58Flipped(g, me, s82g58Grist); !ok {
		t.Error("Grist did not transform")
	}
}

func TestS82GapB58GristPlagueSwarmPlusOneMakesAnInsectAndMills(t *testing.T) {
	for _, black := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		walker := s82g58Walker(t, g, s82g58GristRow(), me, 3)
		me.Library.Cards = nil
		colors := []string{"G"}
		if black {
			colors = []string{"B"}
		}
		// the first card pushed is the top of the library
		pushLibraryCardForTest(me, game.Card{Name: "Milled A", TypeLine: "Sorcery", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
		pushLibraryCardForTest(me, game.Card{Name: "Milled B", TypeLine: "Sorcery", Colors: colors, Owner: me.ID, Controller: me.ID})
		pushLibraryCardForTest(me, game.Card{Name: "Deep", TypeLine: "Sorcery", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
		gy := me.Graveyard.Size()
		b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		if got := me.Graveyard.Size() - gy; got != 2 {
			t.Fatalf("milled %d, want 2", got)
		}
		var insect *game.Card
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Name == "Insect" && c.Controller == me.ID {
				c := c
				insect = &c
			}
		}
		if insect == nil || insect.Power != 1 || insect.Toughness != 1 {
			t.Fatalf("no 1/1 Insect token: %+v", insect)
		}
		if got := insect.Counters[game.CounterDeathtouch] > 0; got != black {
			t.Errorf("black card milled=%v but deathtouch counter=%v", black, got)
		}
		if got := loyaltyOf(g, walker); got != 4 {
			t.Errorf("loyalty %d after +1, want 4", got)
		}
	}
}

func TestS82GapB58GristPlagueSwarmMinusTwoDestroysAnArtifactOrEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58GristRow(), me, 3)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := g.ActivateCatalogAbility(me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true}); err == nil {
		t.Fatal("a creature is not an artifact or enchantment")
	}
	b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{Targets: cardRefs(rock)})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact survived")
	}
}

func TestS82GapB58GristPlagueSwarmMinusSixCopiesEveryCreatureCardAsAnInsect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58GristRow(), me, 6) // dies to the cost, its front face is a creature card in the graveyard
	a := s82g58GraveCreature(me, "Elf Warrior", "Legendary Creature — Elf Warrior", []string{"G"})
	b := s82g58GraveCreature(me, "Golem", "Artifact Creature — Golem", nil)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	got := map[string]game.Card{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.IsToken() {
			got[c.Name] = c
		}
	}
	// the two creature cards, and Grist's own front face (CR 712.8a): he died to the cost
	if len(got) != 3 {
		names := []string{}
		for n := range got {
			names = append(names, n)
		}
		t.Fatalf("created tokens %v, want one per creature card", names)
	}
	for _, name := range []string{"Elf Warrior", "Golem", "Grist, Voracious Larva"} {
		c, ok := got[name]
		if !ok {
			t.Fatalf("no token copy of %s", name)
		}
		if c.Power != 1 || c.Toughness != 1 || !c.HasSubtype("Insect") || c.HasSubtype("Elf") || c.HasSubtype("Golem") {
			t.Errorf("%s token: %d/%d subtypes %v, want a 1/1 Insect", name, c.Power, c.Toughness, c.TypeLine)
		}
		if cols := c.EffectiveColors(); len(cols) != 2 {
			t.Errorf("%s token colors %v, want black and green", name, cols)
		}
	}
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the original cards should stay in the graveyard")
	}
}

// --- Ajani -----------------------------------------------------------

func TestS82GapB58AjaniMakesACatWarriorAndTransformsWhenAnotherCatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ajani := importToBattlefield(t, g, s82g58AjaniRow(), me)
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Cat Warrior" && c.Power == 2 && c.Toughness == 1 && c.IsToken() {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no 2/1 Cat Warrior token")
	}
	// A Cat that is not mine, and a creature that is not a Cat, don't count.
	theirCat := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	for _, id := range []uuid.UUID{theirCat, bear} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	}
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if owedMays(g, me.ID) != 0 {
		t.Fatal("Ajani prompted for a creature that was not another Cat I control")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	walker, ok := s82g58Flipped(g, me, s82g58Ajani)
	if !ok {
		t.Fatal("Ajani did not transform")
	}
	if g.Battlefield.Contains(ajani) {
		t.Error("the front face is still the same object")
	}
	if e2Card(t, g, walker).Name != "Ajani, Nacatl Avenger" || loyaltyOf(g, walker) != 3 {
		t.Errorf("back face wrong, loyalty %d", loyaltyOf(g, walker))
	}
}

func TestS82GapB58AjaniAvengerPlusTwoCountersEachCat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58AjaniRow(), me, 3)
	cat := b12Creature(g, me.ID, "Cat", "Creature — Cat", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirCat := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	b16Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	for name, want := range map[string]struct {
		id uuid.UUID
		n  int
	}{"my Cat": {cat, 1}, "my Bear": {bear, 0}, "their Cat": {theirCat, 0}} {
		if got := e2Card(t, g, want.id).Counters[game.CounterPlusOne]; got != want.n {
			t.Errorf("%s has %d +1/+1 counters, want %d", name, got, want.n)
		}
	}
}

func TestS82GapB58AjaniAvengerZeroMakesACatAndDealsDamageWithARedPermanent(t *testing.T) {
	for _, red := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		walker := pushWalkerForTest(g, me.ID, "Ajani, Nacatl Avenger", s82g58Ajani+"#1", 3)
		b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if red {
			pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Red Thing", TypeLine: "Creature — Goblin",
				Colors: []string{"R"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
		}
		countCats := func() int {
			n := 0
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == me.ID && c.Name == "Cat Warrior" {
					n++
				}
			}
			return n
		}
		before := countCats()
		life := lifeOf(g, opp.ID)
		b16Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{})
		passPriorityAroundTable(t, g)
		if pick := latestPickTarget(g, me.ID); (pick != nil) != red {
			t.Fatalf("red=%v: reflexive target prompt open = %v", red, pick != nil)
		}
		if red {
			pickTriggerTargetPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
		}
		creatures := 0
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID && c.IsCreature() {
				creatures++
			}
		}
		if got := countCats() - before; got != 1 {
			t.Fatalf("red=%v: %d new Cat Warrior tokens, want 1", red, got)
		}
		want := 0
		if red {
			want = creatures
		}
		if got := life - lifeOf(g, opp.ID); got != want {
			t.Errorf("red=%v: damage %d, want %d (the creatures I control)", red, got, want)
		}
	}
}

func TestS82GapB58AjaniAvengerMinusFourLeavesEachOpponentOneOfEachType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	walker := s82g58Walker(t, g, s82g58AjaniRow(), me, 5)
	mine1 := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	mine2 := b12Creature(g, me.ID, "My Ox", "Creature — Ox", 2, 2)
	a1 := b12Permanent(g, opp.ID, "Rock", "Artifact")
	a2 := b12Permanent(g, opp.ID, "Stone", "Artifact")
	c1 := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c2 := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 2, 2)
	land := pushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	b16Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the controller was asked: it is each OPPONENT who chooses")
	}
	if latestChooseCardsFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked")
	}
	answerChooseCards(t, g, opp.ID, a1, c2)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"my Bear": mine1, "my Ox": mine2, "their kept artifact": a1, "their kept creature": c2, "their land": land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"their other artifact": a2, "their other creature": c1} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was kept", name)
		}
	}
}
