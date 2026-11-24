package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b002_cards_test.go — S82 OLD batch 2 (#5462, Secret Lair Drop):
// gods, Slivers and multicolour legends. Each test plays the real catalog
// entry.

const (
	oldB002Keranos = "2bd6f13a-ae02-426d-8f99-c20c5c08cf55"
	oldB002Knight  = "393e5331-4456-4ba3-b410-bb630bb053db"
	oldB002Compost = "4681bcc2-8a46-44cd-8b04-6ab140904dea"
	oldB002Anje    = "4dab6a96-4376-4aea-983d-406167993214"
	oldB002Kynaios = "53ee4254-fef7-49ec-aafc-0320987764e6"
	oldB002Sliver  = "55e4d4b2-1bd8-4c29-a21a-c014a974b715"
	oldB002Ephara  = "5968ecb5-5728-429f-9324-0c108820ebe8"
	oldB002Mathise = "61be7df8-cc78-4822-9e70-f226a8a96c9d"
	oldB002Oloro   = "620ff5f2-7d3f-467f-943d-3b62c2135023"
)

func TestOldB002EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB002Keranos: "Keranos, God of Storms",
		oldB002Knight:  "Knight of the Reliquary",
		oldB002Compost: "Compost",
		oldB002Anje:    "Anje Falkenrath",
		oldB002Kynaios: "Kynaios and Tiro of Meletis",
		oldB002Sliver:  "The First Sliver",
		oldB002Ephara:  "Ephara, God of the Polis",
		oldB002Mathise: "Mathise, Surge Channeler",
		oldB002Oloro:   "Oloro, Ageless Ascetic",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB002Library(p *game.Player, n int, name, typeLine string) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func oldB002Emit(g *game.Game, ev game.Event) {
	g.WithWriteLock(func() { g.EmitEvent(ev) })
}

// --- The First Sliver --------------------------------------------------------

func TestOldB002FirstSliverHasCascadeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.CastWithCost(t, g, "The First Sliver", "Legendary Creature — Sliver", "{W}{U}{B}{R}{G}", oldB002Sliver)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("casting The First Sliver got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
}

func TestOldB002FirstSliverGivesCascadeToSliverSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "The First Sliver", "Legendary Creature — Sliver", oldB002Sliver, false)
	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.CastWithCost(t, g, "Plain Sliver", "Creature — Sliver", "{2}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 1 {
		t.Errorf("a Sliver spell got %d cascade offers, want 1", got)
	}
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.SeedCheapLibrary(g, me, 6)
	effectstest.CastWithCost(t, g, "Plain Bear", "Creature — Bear", "{2}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountMayCastPrompts(g); got != 0 {
		t.Errorf("a non-Sliver spell got %d cascade offers, want 0", got)
	}
}

// --- Compost -----------------------------------------------------------------

func TestOldB002CompostDrawsForEachBlackCardPutIntoAnOpponentsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB002Library(me, 6, "Forest", "Basic Land — Forest")
	effectstest.BfCardWithOracle(g, me, "Compost", "Enchantment", oldB002Compost)
	hand := func(color []string, name string) {
		g.WithWriteLock(func() {
			opp.Hand.Cards = nil
			opp.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Sorcery", Colors: color, Owner: opp.ID, Controller: opp.ID})
		})
	}
	discard := func() {
		var err error
		g.WithWriteLock(func() { err = g.DiscardRandomForEffect(opp.ID, 1) })
		if err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	before := me.Hand.Size()
	hand([]string{"R"}, "Red Card")
	discard()
	if me.Hand.Size() != before {
		t.Fatal("drew for a red card")
	}
	hand([]string{"B"}, "Black Card")
	discard()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("hand = %d, want %d after an opponent's black card hit their graveyard", me.Hand.Size(), before+1)
	}
	// Your own black card does nothing.
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "My Black", TypeLine: "Sorcery", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	})
	mine := me.Hand.Size()
	var err error
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:]; err = g.DiscardRandomForEffect(me.ID, 1) })
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID) != nil {
		t.Error("Compost offered a draw for the controller's own black card")
	}
	_ = mine
}

// --- Oloro -------------------------------------------------------------------

func TestOldB002OloroGainsTwoAtUpkeepAndOffersTheLifegainPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB002Library(me, 8, "Forest", "Basic Land — Forest")
	effectstest.BfCardWithOracle(g, me, "Oloro, Ageless Ascetic", "Legendary Creature — Giant Soldier", oldB002Oloro)
	life, oppLife, hand := me.Life, opp.Life, me.Hand.Size()
	oldB002Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	// The gain triggers the payment offer; decline it first.
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Fatalf("life = %d, want %d", me.Life, life+2)
	}
	if me.Hand.Size() != hand || opp.Life != oppLife {
		t.Fatal("declining the payment still drew or drained")
	}
	// Gain again and pay {1}.
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || opp.Life != oppLife-1 {
		t.Errorf("paying: hand %d (want %d), opponent life %d (want %d)", me.Hand.Size(), hand+1, opp.Life, oppLife-1)
	}
}

func TestOldB002OloroGainsTwoFromTheCommandZoneOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB002Library(me, 4, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		me.Command.PushTop(game.Card{InstanceID: uuid.New(), Name: "Oloro, Ageless Ascetic", OracleID: oldB002Oloro,
			TypeLine: "Legendary Creature — Giant Soldier", IsCommander: true, Owner: me.ID, Controller: me.ID})
	})
	life := me.Life
	oldB002Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d from the command zone", me.Life, life+2)
	}
	// Not on opponents' upkeeps.
	oldB002Emit(g, game.Event{Kind: game.EventBeginUpkeep, Actor: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("gained life on an opponent's upkeep: %d", me.Life)
	}
}

// --- Knight of the Reliquary -------------------------------------------------

func TestOldB002KnightGrowsWithLandsInTheGraveyardAndFetchesAfterASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	knight := effectstest.PushCatalogPermanent(g, me.ID, "Knight of the Reliquary", "Creature — Human Knight", oldB002Knight, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == knight {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 2
			}
		}
	})
	forest := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.PushGraveyardPermanent(me, "Old Land", "Land", "")
	effectstest.PushGraveyardPermanent(me, "Old Land Two", "Land", "")
	effectstest.PushGraveyardPermanent(me, "Old Bear", "Creature — Bear", "{1}{G}")
	if pw, tg := effectstest.B38PT(t, g, knight); pw != 4 || tg != 4 {
		t.Errorf("Knight is %d/%d, want 4/4 with two land cards in the graveyard", pw, tg)
	}
	me.Library.Cards = nil
	wanted := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: wanted, Name: "Volcanic Island", TypeLine: "Land — Island Mountain", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	// A non-Forest, non-Plains cannot be sacrificed.
	island := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	if err := g.ActivateCatalogAbility(me.ID, knight, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{island}}); err == nil {
		t.Error("an Island paid the cost")
	}
	effectstest.Pr7Activate(t, g, me.ID, knight, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{forest}})
	effectstest.AnswerSearchByID(t, g, me.ID, wanted)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(forest) {
		t.Error("the Forest was not sacrificed")
	}
	if !g.Battlefield.Contains(wanted) {
		t.Error("the fetched land did not enter")
	}
	if c, _ := effectstest.BattlefieldCard(g, knight); !c.Tapped {
		t.Error("the Knight was not tapped for the cost")
	}
	// The sacrificed Forest is now a land card in the graveyard too.
	if pw, _ := effectstest.B38PT(t, g, knight); pw != 5 {
		t.Errorf("Knight power = %d, want 5 with three land cards in the graveyard", pw)
	}
}

// --- Kynaios and Tiro of Meletis ---------------------------------------------

func TestOldB002KynaiosDrawsAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB002Library(me, 6, "Forest", "Basic Land — Forest")
	effectstest.BfCardWithOracle(g, me, "Kynaios and Tiro of Meletis", "Legendary Creature — Human Soldier", oldB002Kynaios)
	before := me.Hand.Size()
	oldB002Emit(g, game.Event{Kind: game.EventBeginEndStep, Actor: me.ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() < before+1 {
		t.Errorf("hand = %d, want at least %d", me.Hand.Size(), before+1)
	}
	// An opponent's end step does nothing.
	h := me.Hand.Size()
	oldB002Emit(g, game.Event{Kind: game.EventBeginEndStep, Actor: g.Seats[1].ID})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != h {
		t.Error("drew at an opponent's end step")
	}
}

// --- Mathise -----------------------------------------------------------------

func TestOldB002MathiseTriggersOnlyOnBigInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB002Library(me, 10, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Mathise, Surge Channeler", "Legendary Creature — Human Elf Shaman Sorcerer", oldB002Mathise, false)
	rolled := func() bool {
		for _, ev := range g.Events {
			if ev.Kind == game.EventRollDie {
				return true
			}
		}
		return false
	}
	effectstest.CastWithCost(t, g, "Cheap Spell", "Instant", "{1}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if rolled() {
		t.Fatal("rolled for a mana value 2 instant")
	}
	effectstest.CastWithCost(t, g, "Big Creature", "Creature — Bear", "{3}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if rolled() {
		t.Fatal("rolled for a creature spell")
	}
	effectstest.CastWithCost(t, g, "Big Sorcery", "Sorcery", "{2}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if !rolled() {
		t.Error("did not roll for a mana value 3 sorcery")
	}
}

// --- Ephara ------------------------------------------------------------------

func TestOldB002EpharaDrawsAfterAnotherCreatureEnteredLastTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB002Library(me, 8, "Forest", "Basic Land — Forest")
	effectstest.BfCardWithOracle(g, me, "Ephara, God of the Polis", "Legendary Enchantment Creature — God", oldB002Ephara)
	// A creature enters under my control this turn.
	bear := effectstest.SeedCreature(g, "Bear", me.ID)
	oldB002Emit(g, game.Event{Kind: game.EventETB, CardID: bear, Actor: me.ID})
	hand := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand = %d, want %d: a creature entered last turn", me.Hand.Size(), hand+1)
	}
	// The turn after that nothing entered, so the next upkeep draws nothing.
	hand = me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew with no creature entering last turn: hand %d (was %d)", me.Hand.Size(), hand)
	}
}

func TestOldB002EpharaIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB002Library(me, 8, "Forest", "Basic Land — Forest")
	effectstest.BfCardWithOracle(g, me, "Ephara, God of the Polis", "Legendary Enchantment Creature — God", oldB002Ephara)
	bear := effectstest.SeedCreature(g, "Their Bear", opp.ID)
	oldB002Emit(g, game.Event{Kind: game.EventETB, CardID: bear, Actor: opp.ID})
	hand := me.Hand.Size()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew for an opponent's creature: hand %d (was %d)", me.Hand.Size(), hand)
	}
}

// --- Keranos -----------------------------------------------------------------

// oldB002KeranosTable puts Keranos under seat 0, fills its library with n
// copies of one card and walks to seat 0's next draw step: the card the draw
// step draws is the first it draws that turn.
func oldB002KeranosTable(t *testing.T, n int, name, typeLine string) (*game.Game, *game.Player, *game.Player, int) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.BfCardWithOracle(g, me, "Keranos, God of Storms", "Legendary Enchantment Creature — God", oldB002Keranos)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	oldB002Library(me, n, name, typeLine)
	before := me.Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, before
}

func TestOldB002KeranosRevealsTheFirstDrawAndDrawsOnALand(t *testing.T) {
	g, me, _, before := oldB002KeranosTable(t, 8, "Forest", "Basic Land — Forest")
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("a revealed land: drew %d cards, want 2 (the draw step's card and one more)", got)
	}
	revealed := false
	for _, ev := range g.Events {
		if ev.Kind == game.EventRevealCards {
			revealed = true
		}
	}
	if !revealed {
		t.Error("the drawn card was not revealed")
	}
	// The second card drawn this turn is not revealed or counted.
	n := me.Hand.Size()
	effectstest.DrawN(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != n+1 {
		t.Errorf("the second draw triggered Keranos: hand %d (was %d)", me.Hand.Size(), n)
	}
}

func TestOldB002KeranosDealsThreeOnANonlandFirstDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.BfCardWithOracle(g, me, "Keranos, God of Storms", "Legendary Enchantment Creature — God", oldB002Keranos)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	oldB002Library(me, 8, "Grizzly Bears", "Creature — Bear")
	life, hand := opp.Life, me.Hand.Size()
	effectstest.AdvanceToDrawStepOfSeat(t, g, 0)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent took %d, want 3", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("a nonland first card drew %d cards, want 1", got)
	}
}

// Keranos reads the first card drawn on ITS CONTROLLER'S turns only.
func TestOldB002KeranosIgnoresADrawOnAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.BfCardWithOracle(g, me, "Keranos, God of Storms", "Legendary Enchantment Creature — God", oldB002Keranos)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	oldB002Library(me, 8, "Forest", "Basic Land — Forest")
	before := me.Hand.Size()
	effectstest.DrawN(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+1 {
		t.Errorf("Keranos triggered on a draw during an opponent's turn: hand %d (was %d)", me.Hand.Size(), before)
	}
}

// --- Anje Falkenrath ---------------------------------------------------------

func TestOldB002AnjeLootsAndUntapsOnAMadnessDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB002Library(me, 6, "Forest", "Basic Land — Forest")
	anje := effectstest.PushCatalogPermanent(g, me.ID, "Anje Falkenrath", "Legendary Creature — Vampire", oldB002Anje, false)
	fodder := game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(fodder) })
	hand := me.Hand.Size()
	effectstest.Pr7Activate(t, g, me.ID, anje, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder.InstanceID}})
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (discard one, draw one)", me.Hand.Size(), hand)
	}
	if c, _ := effectstest.BattlefieldCard(g, anje); !c.Tapped {
		t.Error("Anje did not tap for the cost")
	}
}

func TestOldB002AnjeUntapsWhenYouDiscardACardWithMadness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB002Library(me, 6, "Forest", "Basic Land — Forest")
	anje := effectstest.PushCatalogPermanent(g, me.ID, "Anje Falkenrath", "Legendary Creature — Vampire", oldB002Anje, false)
	hunter := game.Card{InstanceID: uuid.New(), Name: "Big Game Hunter", OracleID: "ab55834f-c935-4773-89c6-bec9712284eb",
		TypeLine: "Creature — Human Rogue", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(hunter) })
	effectstest.Pr7Activate(t, g, me.ID, anje, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{hunter.InstanceID}})
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == me.ID && c.Kind != game.PendingChoiceConfirm {
			t.Logf("open choice: %v", c.Kind)
		}
	}
	for i := 0; i < 4 && effectstest.TriggerOrderPrompt(g) != nil; i++ {
		effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAllMayCast(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, anje); c.Tapped {
		t.Error("Anje stayed tapped after a madness card was discarded")
	}
}

// The results table: every roll draws what its band says. The die is the
// game's own seeded stream, so the loop casts until it has seen rolls from the
// bands that draw; the 20 band (a copy of the spell) draws nothing.
func TestOldB002MathiseRollBandsDrawAsPrinted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB002Library(me, 200, "Forest", "Basic Land — Forest")
	oldB002Library(opp, 200, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Mathise, Surge Channeler", "Legendary Creature — Human Elf Shaman Sorcerer", oldB002Mathise, false)
	seen := map[string]int{}
	for i := 0; i < 60; i++ {
		rolls := len(effectstest.RollEventsOf(g))
		mine, theirs := me.Hand.Size(), opp.Hand.Size()
		effectstest.CastWithCost(t, g, "Big Sorcery", "Sorcery", "{2}{R}", "")
		effectstest.PassPriorityAroundTable(t, g)
		evs := effectstest.RollEventsOf(g)
		if len(evs) != rolls+1 {
			t.Fatalf("cast %d: %d new rolls, want 1", i, len(evs)-rolls)
		}
		// The cast card is added to the hand and cast again, a net zero.
		dm, do := me.Hand.Size()-mine, opp.Hand.Size()-theirs
		result := evs[len(evs)-1].Amount
		switch {
		case result >= 20:
			seen["20"]++
			if dm != 0 || do != 0 {
				t.Errorf("a roll of %d drew (%d, %d)", result, dm, do)
			}
		case result >= 10:
			seen["10-19"]++
			if dm != 1 || do != 0 {
				t.Errorf("a roll of %d: you drew %d, the opponent %d; want 1 and 0", result, dm, do)
			}
		default:
			seen["1-9"]++
			if dm != 1 || do != 1 {
				t.Errorf("a roll of %d: you drew %d, the opponent %d; want 1 and 1", result, dm, do)
			}
		}
		if g.Stack.Size() != 0 {
			effectstest.PassPriorityAroundTable(t, g)
		}
		g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	}
	if seen["1-9"] == 0 || seen["10-19"] == 0 {
		t.Errorf("60 rolls never reached both drawing bands: %v", seen)
	}
}
