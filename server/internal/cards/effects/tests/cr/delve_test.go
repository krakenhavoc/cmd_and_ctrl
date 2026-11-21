package testscr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fuel := effectstest.DelveFuel(me, 7)
	// The pool empties as a step ends (CR 106.4), so walk to the main
	// phase before floating the {U}.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})
	id, err := effectstest.CastCruise(t, g, game.CastSpellParams{Strict: true, DelveIDs: fuel})
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
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Fatalf("Treasure Cruise drew %d, want 3", got)
	}
}

// Delve pays generic mana only. Seven cards exiled still leave the {U},
// and a strict cast with an empty pool is refused with nothing exiled.
func TestDelveCannotPayTheColoredSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fuel := effectstest.DelveFuel(me, 7)
	_, err := effectstest.CastCruise(t, g, game.CastSpellParams{Strict: true, DelveIDs: fuel})
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fuel := effectstest.DelveFuel(me, 8)
	if _, err := effectstest.CastCruise(t, g, game.CastSpellParams{DelveIDs: fuel}); err == nil {
		t.Fatal("delving eight at a seven-generic spell was accepted")
	}
	if me.Graveyard.Size() != 8 {
		t.Fatal("a refused cast exiled cards")
	}
}

// Delve says "your graveyard": an opponent's card is not a candidate.
func TestDelveRefusesACardFromAnotherGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.DelveFuel(g.Seats[1], 1)
	if _, err := effectstest.CastCruise(t, g, game.CastSpellParams{DelveIDs: theirs}); err == nil {
		t.Fatal("delving an opponent's graveyard card was accepted")
	}
}

// A card without delve that arrives with delve_ids is a client bug and
// is refused, as a stray discard is.
func TestDelveIDsOnACardWithoutDelveAreRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	fuel := effectstest.DelveFuel(g.Seats[0], 1)
	if _, err := effectstest.CastWithTapParams(t, g, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationDelveOracle,
		game.CastSpellParams{DelveIDs: fuel}); err == nil {
		t.Fatal("delve_ids on Divination were accepted")
	}
}

// A card named twice pays once (CR 118.3).
func TestDelveRefusesTheSameCardTwice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	fuel := effectstest.DelveFuel(g.Seats[0], 1)
	if _, err := effectstest.CastCruise(t, g, game.CastSpellParams{DelveIDs: []uuid.UUID{fuel[0], fuel[0]}}); err == nil {
		t.Fatal("one card delved twice was accepted")
	}
}

// The pricer answers the budget, so the preview, the view and the bot
// read the same number the validator refuses against (#696).
func TestPriceCastReportsTheDelveBudget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cruise := game.Card{InstanceID: uuid.New(), Name: "Treasure Cruise", TypeLine: "Sorcery",
		ManaCost: "{7}{U}", OracleID: effectstest.TreasureCruiseOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(cruise)
	price, err := g.PriceCast(me.ID, cruise, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 7 {
		t.Fatalf("Treasure Cruise budget = %d, want 7", price.DelveBudget)
	}
	fuel := effectstest.DelveFuel(me, 3)
	price, err = g.PriceCast(me.ID, cruise, game.CastSpellParams{DelveIDs: fuel})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 7 || price.Total.Generic != 4 {
		t.Fatalf("with three delved: budget %d, generic %d; want 7 and 4", price.DelveBudget, price.Total.Generic)
	}

	knot := game.Card{InstanceID: uuid.New(), Name: "Logic Knot", TypeLine: "Instant",
		ManaCost: "{X}{U}{U}", OracleID: effectstest.LogicKnotOracle, Owner: me.ID, Controller: me.ID}
	me.Hand.PushTop(knot)
	price, err = g.PriceCast(me.ID, knot, game.CastSpellParams{XValue: 3})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if price.DelveBudget != 3 {
		t.Fatalf("Logic Knot at X=3 budget = %d, want 3", price.DelveBudget)
	}

	plain := game.Card{InstanceID: uuid.New(), Name: "Divination", TypeLine: "Sorcery",
		ManaCost: "{2}{U}", OracleID: effectstest.DivinationDelveOracle, Owner: me.ID, Controller: me.ID}
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cmdr := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cmdr, Name: "My Commander", TypeLine: "Legendary Creature — Test",
		Owner: me.ID, Controller: me.ID, IsCommander: true})
	id, err := effectstest.CastCruise(t, g, game.CastSpellParams{DelveIDs: []uuid.UUID{cmdr}})
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if me.Hand.Contains(id) || !g.Stack.Contains(id) {
		t.Fatal("the cast was not paid: the spell is not on the stack")
	}
	if !g.Exile.Contains(cmdr) {
		t.Fatal("the delved commander was not exiled")
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionalReplacement, me.ID); c != nil {
		t.Fatalf("the delve asked the CR 903.9b replacement first: %+v", c)
	}
	effectstest.AnswerCommanderReturn(t, g, me.ID, true)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Dig Through Time", "Instant", effectstest.DigThroughTimeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no choose-cards prompt: %+v", g.PendingChoices)
	}
	if pick.ChooseMin != 2 || pick.ChooseMax != 2 || len(pick.ChooseCards) != 7 {
		t.Fatalf("prompt picks %d..%d of %d, want 2..2 of 7", pick.ChooseMin, pick.ChooseMax, len(pick.ChooseCards))
	}
}

func TestMurderousCutDestroysTargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Murderous Cut", "Instant", effectstest.MurderousCutOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("Murderous Cut did not destroy its target")
	}
}

// Death Rattle targets a NONGREEN creature.
func TestDeathRattleRefusesAGreenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	green := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: green, Name: "Elf", TypeLine: "Creature — Elf",
		Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Death Rattle", "Instant", effectstest.DeathRattleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: green}}); err == nil {
		t.Fatal("Death Rattle targeted a green creature")
	}
	victim := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Death Rattle", "Instant", effectstest.DeathRattleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Fatal("Death Rattle did not destroy a nongreen creature")
	}
}

func TestBecomeImmenseGivesSixSix(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Become Immense", "Instant", effectstest.BecomeImmenseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCardByID(g, bear)
	if p := c.Effective().Power; p != 8 {
		t.Fatalf("power = %d, want 8", p)
	}
}

func TestMagmaticSinkholeDealsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := uuid.New()
	g.Battlefield.PushTop(game.Card{InstanceID: victim, Name: "Big", TypeLine: "Creature — Test",
		Power: 6, Toughness: 6, Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	effectstest.CastCatalogSpell(t, g, "Magmatic Sinkhole", "Instant", effectstest.MagmaticSinkholeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCardByID(g, victim)
	if c.DamageMarked != 5 {
		t.Fatalf("damage = %d, want 5", c.DamageMarked)
	}
}

func TestSetAdriftPutsThePermanentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	victim := effectstest.PushCreatureToBattlefieldForTest(g, owner.ID, "Bear")
	effectstest.CastCatalogSpell(t, g, "Set Adrift", "Sorcery", effectstest.SetAdriftOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	effectstest.PassPriorityAroundTable(t, g)
	top, err := owner.Library.Top()
	if err != nil || top.InstanceID != victim {
		t.Fatal("Set Adrift did not put the permanent on top of its owner's library")
	}
}

// Tasigur's Cruelty asks each opponent — not its caster — to discard two.
func TestTasigursCrueltyAsksEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Tasigur's Cruelty", "Sorcery", effectstest.TasigursCrueltyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.DiscardChoiceFor(g, g.Seats[0].ID); c != nil {
		t.Fatal("the caster was asked to discard")
	}
	asked := 0
	for _, p := range g.Seats[1:] {
		if c := effectstest.DiscardChoiceFor(g, p.ID); c != nil {
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
	g := effectstest.NewCatalogGame(t)
	a := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "A")
	b := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "B")
	effectstest.CastCatalogSpell(t, g, "Will of the Naga", "Instant", effectstest.WillOfTheNagaOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if c, _ := effectstest.BattlefieldCardByID(g, id); !c.Tapped {
			t.Fatal("Will of the Naga did not tap a target")
		}
	}
}

// Logic Knot at X = 2 asks the targeted spell's controller for {2}.
func TestLogicKnotTaxesByX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	victim := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationDelveOracle, nil)
	if _, err := effectstest.CastWithTapParams(t, g, "Logic Knot", "Instant", "{X}{U}{U}", effectstest.LogicKnotOracle, game.CastSpellParams{
		XValue: 2, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("CastSpell Logic Knot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.Chooser != me.ID || c.PayCost != "{2}" {
		t.Fatalf("prompt asks %v for %q, want the caster for {2}", c.Chooser, c.PayCost)
	}
}

func TestEmptyThePitsMakesXTappedZombies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if _, err := effectstest.CastWithTapParams(t, g, "Empty the Pits", "Instant", "{X}{X}{B}{B}{B}{B}", effectstest.EmptyThePitsOracle,
		game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Mine")
	theirs := effectstest.PushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Theirs")
	effectstest.CastCatalogSpell(t, g, "Rite of Undoing", "Instant", effectstest.RiteOfUndoingOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Seats[0].Hand.Contains(mine) || !g.Seats[1].Hand.Contains(theirs) {
		t.Fatal("Rite of Undoing did not return both permanents to their owners' hands")
	}
}

// Dead Drop's target player chooses two of their creatures, and both
// are sacrificed.
func TestDeadDropSacrificesTwoOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "A")
	b := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "B")
	c := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "C")
	effectstest.CastCatalogSpell(t, g, "Dead Drop", "Sorcery", effectstest.DeadDropOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, opp.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dead := effectstest.PushGraveyardCardForTest(me, "Dead Bear")
	id := effectstest.CastCatalogSpell(t, g, "Sibsig Muckdraggers", "Creature — Zombie", effectstest.SibsigOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Sibsig Muckdraggers did not resolve")
	}
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); c != nil {
		if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !me.Hand.Contains(dead) {
		t.Fatal("the creature card did not return to hand")
	}
}
