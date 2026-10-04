package effects

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve_test.go — CR 702.66, ADR 0100 sub-PR 1: delve as a payment.
// The pure arithmetic (the budget, X, the any-colour fold) is tested
// next to the code in game/delve_test.go; these are the board-level
// rules — what the announcement may name, when the cards are exiled,
// and what the record keeps.

const (
	treasureCruiseOracle   = "5b6bdf5a-2742-4851-92cd-a857a3852836"
	digThroughTimeOracle   = "f8b17b89-26ce-4208-874a-9e1d66514640"
	murderousCutOracle     = "c0ce7d5a-68fd-41f3-b5ba-a1a178cee9ec"
	deathRattleOracle      = "ae327976-1026-4318-8846-f6b0cac373a7"
	becomeImmenseOracle    = "40e5b83e-5f53-4b15-8bfd-c0c0b8355a6f"
	magmaticSinkholeOracle = "2da7f1b7-316b-4e8e-9fa3-5880b258d601"
	setAdriftOracle        = "f27f51e1-b881-4639-9d36-47cdf2a61117"
	tasigursCrueltyOracle  = "a232daf3-db54-4711-88b0-e5072c05f58e"
	willOfTheNagaOracle    = "520b6637-0a9f-4dc4-846e-f1cd2b868263"
	logicKnotOracle        = "b2da7acb-d80c-414c-9f7d-753a5d6ccad9"
	emptyThePitsOracle     = "089d91cf-ed7a-4859-8967-cad975a5127e"
	riteOfUndoingOracle    = "9dbfa026-e364-4111-a03a-e9b1693bc7b7"
	deadDropOracle         = "09f3d60c-34ee-41ec-a047-fe2140b11950"
	sibsigOracle           = "423c1079-bdba-42fc-8732-59cb83ebafc3"
	divinationDelveOracle  = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
)

// delveFuel puts n instants into p's graveyard and returns their IDs,
// oldest first.
func delveFuel(p *game.Player, n int) []uuid.UUID {
	ids := make([]uuid.UUID, 0, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant",
			Owner: p.ID, Controller: p.ID,
		})
		ids = append(ids, id)
	}
	return ids
}

func castCruise(t *testing.T, g *game.Game, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Treasure Cruise", "Sorcery", "{7}{U}", treasureCruiseOracle, params)
}

func delvedInExile(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- the payment ---------------------------------------------------

// Seven graveyard cards pay Treasure Cruise's {7}; the {U} is paid with
// mana. The cards are in exile with the spell still on the stack
// (CR 601.2h), the record names each one as the object that landed in
// exile, and the spell resolves for three cards.
func TestDelveExilesTheNamedCardsAndPaysTheGeneric(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := delveFuel(me, 7)
	// The pool empties as a step ends (CR 106.4), so walk to the main
	// phase before floating the {U}.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	id, err := castCruise(t, g, game.CastSpellParams{Strict: true, DelveIDs: fuel})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Graveyard.Size() != 0 {
		t.Fatalf("graveyard holds %d cards after delving all seven", me.Graveyard.Size())
	}
	item := g.StackMeta[id]
	if item == nil {
		t.Fatal("Treasure Cruise is not on the stack")
	}
	if len(item.Paid.Delved) != 7 {
		t.Fatalf("Paid.Delved = %v, want seven", item.Paid.Delved)
	}
	for i, ref := range item.Paid.Delved {
		c, ok := delvedInExile(g, ref.ID)
		if !ok || ref.ID != fuel[i] {
			t.Fatalf("delved card %d is not the one named, in exile", i)
		}
		if c.ObjectEpoch != ref.Epoch {
			t.Fatalf("recorded epoch %d, exiled object epoch %d", ref.Epoch, c.ObjectEpoch)
		}
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("Treasure Cruise drew %d, want 3", got)
	}
}

// Delve pays generic mana only. Seven cards exiled still leave the {U},
// and a strict cast with an empty pool is refused with nothing exiled.
func TestDelveCannotPayTheColoredSymbol(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := delveFuel(me, 7)
	_, err := castCruise(t, g, game.CastSpellParams{Strict: true, DelveIDs: fuel})
	var short *game.InsufficientManaError
	if !errors.As(err, &short) {
		t.Fatalf("CastSpell = %v, want InsufficientManaError for the {U}", err)
	}
	if me.Graveyard.Size() != 7 {
		t.Fatalf("a refused cast exiled %d cards", 7-me.Graveyard.Size())
	}
}

// Naming more cards than the budget is a refusal, not a truncation: a
// card exiled for nothing is a card lost for nothing.
func TestDelveRefusesMoreThanTheBudget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := delveFuel(me, 8)
	if _, err := castCruise(t, g, game.CastSpellParams{DelveIDs: fuel}); err == nil {
		t.Fatal("delving eight at a seven-generic spell was accepted")
	}
	if me.Graveyard.Size() != 8 {
		t.Fatal("a refused cast exiled cards")
	}
}

// Delve says "your graveyard": an opponent's card is not a candidate.
func TestDelveRefusesACardFromAnotherGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	theirs := delveFuel(g.Seats[1], 1)
	if _, err := castCruise(t, g, game.CastSpellParams{DelveIDs: theirs}); err == nil {
		t.Fatal("delving an opponent's graveyard card was accepted")
	}
}

// A card without delve that arrives with delve_ids is a client bug and
// is refused, as a stray discard is.
func TestDelveIDsOnACardWithoutDelveAreRefused(t *testing.T) {
	g := newCatalogGame(t)
	fuel := delveFuel(g.Seats[0], 1)
	if _, err := castWithTapParams(t, g, "Divination", "Sorcery", "{2}{U}", divinationDelveOracle,
		game.CastSpellParams{DelveIDs: fuel}); err == nil {
		t.Fatal("delve_ids on Divination were accepted")
	}
}

// A card named twice pays once (CR 118.3).
func TestDelveRefusesTheSameCardTwice(t *testing.T) {
	g := newCatalogGame(t)
	fuel := delveFuel(g.Seats[0], 1)
	if _, err := castCruise(t, g, game.CastSpellParams{DelveIDs: []uuid.UUID{fuel[0], fuel[0]}}); err == nil {
		t.Fatal("one card delved twice was accepted")
	}
}

// The pricer answers the budget, so the preview, the view and the bot
// read the same number the validator refuses against (#696).
func TestPriceCastReportsTheDelveBudget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cruise := game.Card{InstanceID: uuid.New(), Name: "Treasure Cruise", TypeLine: "Sorcery",
		ManaCost: "{7}{U}", OracleID: treasureCruiseOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(cruise)
	price, err := g.PriceCast(me.ID, cruise, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 7 {
		t.Fatalf("Treasure Cruise budget = %d, want 7", price.DelveBudget)
	}
	fuel := delveFuel(me, 3)
	price, err = g.PriceCast(me.ID, cruise, game.CastSpellParams{DelveIDs: fuel})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 7 || price.Total.Generic != 4 {
		t.Fatalf("with three delved: budget %d, generic %d; want 7 and 4", price.DelveBudget, price.Total.Generic)
	}

	knot := game.Card{InstanceID: uuid.New(), Name: "Logic Knot", TypeLine: "Instant",
		ManaCost: "{X}{U}{U}", OracleID: logicKnotOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(knot)
	price, err = g.PriceCast(me.ID, knot, game.CastSpellParams{XValue: 3})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 3 {
		t.Fatalf("Logic Knot at X=3 budget = %d, want 3", price.DelveBudget)
	}

	plain := game.Card{InstanceID: uuid.New(), Name: "Divination", TypeLine: "Sorcery",
		ManaCost: "{2}{U}", OracleID: divinationDelveOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(plain)
	if price, _ := g.PriceCast(me.ID, plain, game.CastSpellParams{}); price.DelveBudget != 0 {
		t.Fatalf("a card with no delve has budget %d", price.DelveBudget)
	}
}

// A commander in the graveyard named to delve is exiled as the cost is
// paid, like any other card (ADR 0115 narrowed #1397's ask-first to
// hand and library costs). It was put into exile, so the CR 903.9a
// state-based action then asks its owner about the command zone.
func TestDelveExilesACommanderThenAsksItsOwner(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmdr := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cmdr, Name: "My Commander", TypeLine: "Legendary Creature — Test",
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	id, err := castCruise(t, g, game.CastSpellParams{DelveIDs: []uuid.UUID{cmdr}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Hand.Contains(id) || !g.Stack.Contains(id) {
		t.Fatal("the cast was not paid: the spell is not on the stack")
	}
	if !g.Exile.Contains(cmdr) {
		t.Fatal("the delved commander was not exiled")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID); c != nil {
		t.Fatalf("the delve asked the CR 903.9b replacement first: %+v", c)
	}
	answerCommanderReturn(t, g, me.ID, true)
	if !me.Command.Contains(cmdr) {
		t.Error("the delved commander did not take the command zone after the yes")
	}
}

// --- the cards -------------------------------------------------------

// The five creatures have delve and nothing but their printed keywords.
func TestDelveCreaturesDeclareDelveAndKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		keyword      string
	}{
		{"Gurmag Angler", "315772f2-abd2-4681-b8c8-1db4b0ccbbcd", ""},
		{"Hooting Mandrills", "70e35385-6129-4bd1-861c-df04469566b7", "trample"},
		{"Sultai Scavenger", "837f72a9-9522-46e9-a48c-c6ff6d0bbe67", "flying"},
		{"Tombstalker", "944735ad-3f17-421a-b8fa-40f3bb8456f2", "flying"},
		{"Shambling Attendants", "cd17bd00-90b1-4b89-9368-2d8735159e47", "deathtouch"},
	} {
		c := game.Card{Name: tc.name, OracleID: tc.oracle}
		if !game.DelveFor(c) {
			t.Errorf("%s has no delve", tc.name)
		}
		if tc.keyword != "" && !game.HasKeyword(&c, tc.keyword) {
			t.Errorf("%s lacks %s", tc.name, tc.keyword)
		}
	}
}

// Dig Through Time looks at seven and asks for exactly two.
func TestDigThroughTimeTakesTwoOfSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Dig Through Time", "Instant", digThroughTimeOracle, nil)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 7 {
		t.Fatalf("prompt picks %d..%d of %d, want 2..2 of 7", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
}

func TestMurderousCutDestroysTargetCreature(t *testing.T) {
	g := newCatalogGame(t)
	victim := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Bear")
	castCatalogSpell(t, g, "Murderous Cut", "Instant", murderousCutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("Murderous Cut did not destroy its target")
	}
}

// Death Rattle targets a NONGREEN creature.
func TestDeathRattleRefusesAGreenCreature(t *testing.T) {
	g := newCatalogGame(t)
	green := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: green, Name: "Elf", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	if err := castCatalogSpellErr(t, g, "Death Rattle", "Instant", deathRattleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: green}}); err == nil {
		t.Fatal("Death Rattle targeted a green creature")
	}
	victim := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Bear")
	castCatalogSpell(t, g, "Death Rattle", "Instant", deathRattleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("Death Rattle did not destroy a nongreen creature")
	}
}

func TestBecomeImmenseGivesSixSix(t *testing.T) {
	g := newCatalogGame(t)
	bear := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")
	castCatalogSpell(t, g, "Become Immense", "Instant", becomeImmenseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, bear)
	if p := c.Effective().Power; p != 8 {
		t.Fatalf("power = %d, want 8", p)
	}
}

func TestMagmaticSinkholeDealsFive(t *testing.T) {
	g := newCatalogGame(t)
	victim := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: victim, Name: "Big", TypeLine: "Creature — Test",
		Power: 6, Toughness: 6, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	castCatalogSpell(t, g, "Magmatic Sinkhole", "Instant", magmaticSinkholeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCardByID(g, victim)
	if c.DamageMarked != 5 {
		t.Fatalf("damage = %d, want 5", c.DamageMarked)
	}
}

func TestSetAdriftPutsThePermanentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	victim := pushCreatureToBattlefieldForTest(g, owner.ID, "Bear")
	castCatalogSpell(t, g, "Set Adrift", "Sorcery", setAdriftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	top, err := owner.Library.Top()
	if err != nil || top.InstanceID != victim {
		t.Fatal("Set Adrift did not put the permanent on top of its owner's library")
	}
}

// Tasigur's Cruelty asks each opponent — not its caster — to discard two.
func TestTasigursCrueltyAsksEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Tasigur's Cruelty", "Sorcery", tasigursCrueltyOracle, nil)
	passPriorityAroundTable(t, g)
	if c := discardChoiceFor(g, g.Seats[0].ID); c != nil {
		t.Fatal("the caster was asked to discard")
	}
	asked := 0
	for _, p := range g.Seats[1:] {
		if c := discardChoiceFor(g, p.ID); c != nil {
			asked++
			if c.ChooseMax != 2 {
				t.Errorf("discard prompt asks for %d, want 2", c.ChooseMax)
			}
		}
	}
	if asked == 0 {
		t.Fatalf("no opponent was asked to discard: %+v", g.PendingChoices)
	}
}

func TestWillOfTheNagaTapsTwo(t *testing.T) {
	g := newCatalogGame(t)
	a := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "A")
	b := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "B")
	castCatalogSpell(t, g, "Will of the Naga", "Instant", willOfTheNagaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := battlefieldCardByID(g, id); !c.Tapped {
			t.Fatal("Will of the Naga did not tap a target")
		}
	}
}

// Logic Knot at X = 2 asks the targeted spell's controller for {2}.
func TestLogicKnotTaxesByX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	victim := castCatalogSpell(t, g, "Divination", "Sorcery", divinationDelveOracle, nil)
	if _, err := castWithTapParams(t, g, "Logic Knot", "Instant", "{X}{U}{U}", logicKnotOracle, game.CastSpellParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("CastSpell Logic Knot: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := pendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.Chooser != me.ID || c.PayCost != "{2}" {
		t.Fatalf("prompt asks %v for %q, want the caster for {2}", c.Chooser, c.PayCost)
	}
}

func TestEmptyThePitsMakesXTappedZombies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if _, err := castWithTapParams(t, g, "Empty the Pits", "Instant", "{X}{X}{B}{B}{B}{B}", emptyThePitsOracle,
		game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Zombie" {
			n++
			if !c.Tapped {
				t.Error("a Zombie entered untapped")
			}
		}
	}
	if n != 2 {
		t.Fatalf("Zombies = %d, want 2", n)
	}
}

func TestRiteOfUndoingReturnsBothPermanents(t *testing.T) {
	g := newCatalogGame(t)
	mine := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Mine")
	theirs := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Theirs")
	castCatalogSpell(t, g, "Rite of Undoing", "Instant", riteOfUndoingOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if !g.Seats[0].Hand.Contains(mine) || !g.Seats[1].Hand.Contains(theirs) {
		t.Fatal("Rite of Undoing did not return both permanents to their owners' hands")
	}
}

// Dead Drop's target player chooses two of their creatures, and both
// are sacrificed.
func TestDeadDropSacrificesTwoOfTheirChoice(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	a := pushCreatureToBattlefieldForTest(g, opp.ID, "A")
	b := pushCreatureToBattlefieldForTest(g, opp.ID, "B")
	c := pushCreatureToBattlefieldForTest(g, opp.ID, "C")
	castCatalogSpell(t, g, "Dead Drop", "Sorcery", deadDropOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
	if prompt == nil {
		t.Fatalf("no sacrifice prompt for the target player: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(prompt.ID, opp.ID, []uuid.UUID{a, c}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(c) || !g.Battlefield.Contains(b) {
		t.Fatal("Dead Drop did not sacrifice exactly the two chosen creatures")
	}
}

// Sibsig Muckdraggers returns a creature card from its controller's
// graveyard on entry.
func TestSibsigMuckdraggersReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardCardForTest(me, "Dead Bear")
	id := castCatalogSpell(t, g, "Sibsig Muckdraggers", "Creature — Zombie", sibsigOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Sibsig Muckdraggers did not resolve")
	}
	if c := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(dead) {
		t.Fatal("the creature card did not return to hand")
	}
}
