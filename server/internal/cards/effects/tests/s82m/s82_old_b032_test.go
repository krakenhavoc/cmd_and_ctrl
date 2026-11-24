package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OLD batch 32 (S82, #5501): Modern Horizons 2 — Chainer, Nightmare Adept;
// Young Necromancer; The Underworld Cookbook; Echoing Return; Said // Done;
// Territorial Kavu; Scour the Desert; Fairgrounds Patrol; Kaleidoscorch;
// Dihada's Ploy; Calibrated Blast; Road // Ruin.

const (
	oldB032Chainer  = "14f6fcb0-bed4-4a38-b7df-1e438d81bedb"
	oldB032Necro    = "17455491-c295-4130-99b4-1b5c6ac1cd25"
	oldB032Cookbook = "37d5c720-d090-44bc-91df-0a7bab79a03e"
	oldB032Echoing  = "591fe685-767e-4e6e-905e-cc8b423caa19"
	oldB032SaidDone = "5d1d76e0-049c-4a33-96ec-8e60deed16d5"
	oldB032Kavu     = "5f0be4ce-898e-4849-ab8f-a85fb05c3017"
	oldB032Scour    = "86c9e08a-397a-4c24-8f4c-48a648889399"
	oldB032Patrol   = "8b0689b5-7c68-46b3-a72a-4f97ef1d0a15"
	oldB032Kaleido  = "b02bd695-ec7f-47be-b797-e2b4b309cc2a"
	oldB032Ploy     = "cc929343-db65-43d5-9bf6-9f0bd77c2cd8"
	oldB032Blast    = "d9f5fc77-40df-4872-b633-6bf5a22a8831"
	oldB032RoadRuin = "f18d8939-7ee8-4a44-a1c8-6ae3a08f52fd"
)

func TestOldB032CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB032Chainer:         "Chainer, Nightmare Adept",
		oldB032Necro:           "Young Necromancer",
		oldB032Cookbook:        "The Underworld Cookbook",
		oldB032Echoing:         "Echoing Return",
		oldB032SaidDone:        "Said",
		oldB032SaidDone + "#1": "Done",
		oldB032Kavu:            "Territorial Kavu",
		oldB032Scour:           "Scour the Desert",
		oldB032Patrol:          "Fairgrounds Patrol",
		oldB032Kaleido:         "Kaleidoscorch",
		oldB032Ploy:            "Dihada's Ploy",
		oldB032Blast:           "Calibrated Blast",
		oldB032RoadRuin:        "Road",
		oldB032RoadRuin + "#1": "Ruin",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v", name, spec)
		}
	}
}

func oldB032Card(name, typeLine, cost, oracle string, p, tough int) game.Card {
	return game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Layout: "normal", Power: p, Toughness: tough}
}

func oldB032Library(p *game.Player, n int) {
	p.Library.Cards = nil
	for i := 0; i < n; i++ {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	}
}

// oldB032Graveyard puts a card with its characteristics into a graveyard.
func oldB032Graveyard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	if c.Layout == "" {
		c.Layout = "normal"
	}
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func oldB032Pick(t *testing.T, g *game.Game, chooser uuid.UUID, kind game.PendingChoiceKind, ids ...uuid.UUID) {
	t.Helper()
	c := effectstest.LatestChoiceOfKindFor(g, kind, chooser)
	if c == nil {
		t.Fatalf("no %s prompt for %s", kind, chooser)
	}
	if err := g.ResolveChooseCards(c.ID, chooser, ids); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
}

// --- Chainer, Nightmare Adept ------------------------------------------------

func TestOldB032ChainerOpensOneCreatureCastFromYourGraveyardPerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	chainer := effectstest.B12Push(g, me.ID, "Chainer, Nightmare Adept", "Legendary Creature — Human Minion", oldB032Chainer, 3, 2)
	bear := oldB032Graveyard(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2})
	bolt := oldB032Graveyard(me, game.Card{Name: "Bolt", TypeLine: "Instant", ManaCost: "{R}"})
	effectstest.FloatForTest(g, me, "GG")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast a creature from the graveyard before activating Chainer")
	}
	pitch := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, chainer, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("cast an instant from the graveyard")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast the creature from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature did not resolve onto the battlefield")
	}
	pitch2 := effectstest.HandCardFull(me, "Fodder Two", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, chainer, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{pitch2}}); err == nil {
		t.Error("activated twice in one turn")
	}
}

func TestOldB032ChainerGivesHasteToCreaturesNotCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Chainer, Nightmare Adept", "Legendary Creature — Human Minion", oldB032Chainer, 3, 2)
	hand, err := effectstest.Msc17Cast(t, g, me, oldB032Card("Hand Bear", "Creature — Bear", "{1}", "", 2, 2), game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HobB10Has(effectstest.GaLayered(t, g, hand), "haste") {
		t.Error("a creature cast from hand gained haste")
	}
	dead := effectstest.PushGraveyardPermanent(me, "Dead Bear", "Creature — Bear", "{2}")
	effectstest.B30Reanimate(t, g, dead, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HobB10Has(effectstest.GaLayered(t, g, dead), "haste") {
		t.Error("a reanimated creature did not gain haste")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	if !effectstest.HobB10Has(effectstest.GaLayered(t, g, dead), "haste") {
		t.Error("the haste ended before your next turn")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	if effectstest.HobB10Has(effectstest.GaLayered(t, g, dead), "haste") {
		t.Error("the haste outlasted until your next turn")
	}
}

// --- Young Necromancer -------------------------------------------------------

func TestOldB032YoungNecromancerExilesTwoToReanimate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := oldB032Graveyard(me, game.Card{Name: "Fodder A", TypeLine: "Sorcery"})
	b := oldB032Graveyard(me, game.Card{Name: "Fodder B", TypeLine: "Sorcery"})
	bear := oldB032Graveyard(me, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.EnterFromHand(t, g, me.ID, "Young Necromancer", "Creature — Human Warlock", oldB032Necro)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	oldB032Pick(t, g, me.ID, game.PendingChoiceChooseCards, a, b)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(b) {
		t.Error("the two cards were not exiled")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature card was not returned to the battlefield")
	}
}

func TestOldB032YoungNecromancerDecliningDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	a := oldB032Graveyard(me, game.Card{Name: "Fodder A", TypeLine: "Sorcery"})
	oldB032Graveyard(me, game.Card{Name: "Fodder B", TypeLine: "Sorcery"})
	bear := oldB032Graveyard(me, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.EnterFromHand(t, g, me.ID, "Young Necromancer", "Creature — Human Warlock", oldB032Necro)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Error("declining changed the graveyard")
	}
}

// --- The Underworld Cookbook -------------------------------------------------

func TestOldB032CookbookDiscardsForAFoodAndSacrificesToReturnACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	book := effectstest.PushCatalogPermanent(g, me.ID, "The Underworld Cookbook", "Artifact — Book", oldB032Cookbook, false)
	fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("food: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discard was not paid")
	}
	if effectstest.CountOnBattlefield(g, "Food", me.ID) != 1 {
		t.Errorf("%d Food tokens, want 1", effectstest.CountOnBattlefield(g, "Food", me.ID))
	}
	dead := oldB032Graveyard(me, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear"})
	effectstest.UntapForTest(g, book)
	if err := effectstest.HobB4Activate(t, g, me.ID, book, 1, "{C}{C}{C}{C}", effectstest.CardTargets(dead)); err != nil {
		t.Fatalf("return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) || g.Battlefield.Contains(book) {
		t.Errorf("creature in hand %v, book still on the battlefield %v", me.Hand.Contains(dead), g.Battlefield.Contains(book))
	}
}

// --- Echoing Return ----------------------------------------------------------

func TestOldB032EchoingReturnTakesEveryCardWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	one := oldB032Graveyard(me, game.Card{Name: "Rat", TypeLine: "Creature — Rat"})
	two := oldB032Graveyard(me, game.Card{Name: "Rat", TypeLine: "Creature — Rat"})
	other := oldB032Graveyard(me, game.Card{Name: "Bat", TypeLine: "Creature — Bat"})
	theirs := oldB032Graveyard(opp, game.Card{Name: "Rat", TypeLine: "Creature — Rat"})
	effectstest.CastCatalogSpell(t, g, "Echoing Return", "Sorcery", oldB032Echoing, effectstest.CardTargets(one))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(one) || !me.Hand.Contains(two) {
		t.Errorf("both Rats should return: first %v, second %v", me.Hand.Contains(one), me.Hand.Contains(two))
	}
	if me.Hand.Contains(other) || !opp.Graveyard.Contains(theirs) {
		t.Error("a differently named card, or an opponent's Rat, moved")
	}
}

// --- Said // Done --------------------------------------------------------------

func TestOldB032SaidReturnsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bolt := oldB032Graveyard(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	bear := oldB032Graveyard(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Said", "Sorcery", oldB032SaidDone, effectstest.CardTargets(bear)); err == nil {
		t.Fatal("Said targeted a creature card")
	}
	effectstest.CastCatalogSpell(t, g, "Said", "Sorcery", oldB032SaidDone, effectstest.CardTargets(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return")
	}
}

func TestOldB032DoneTapsTwoCreaturesAndTheyStayTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 2, 2)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 2, 2)
	c := effectstest.PushVanillaCreature(g, opp.ID, "C", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Done", "Instant", oldB032SaidDone+"#1", effectstest.CardTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
			t.Error("a target is untapped")
		}
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, c); tapped {
		t.Error("a creature that was not targeted was tapped")
	}
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	for _, id := range []uuid.UUID{a, b} {
		if tapped, _ := effectstest.BattlefieldCardTapped(g, id); !tapped {
			t.Error("a target untapped during its controller's next untap step")
		}
	}
}

// --- Territorial Kavu ----------------------------------------------------------

func TestOldB032TerritorialKavuIsAsBigAsYourBasicLandTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	kavu := effectstest.PushCatalogPermanent(g, me.ID, "Territorial Kavu", "Creature — Kavu", oldB032Kavu, false)
	if c := effectstest.GaLayered(t, g, kavu); c.CurrentPower() != 0 || c.CurrentToughness() != 0 {
		t.Fatalf("with no lands it is %d/%d, want 0/0", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Second Forest", "Basic Land — Forest", "", false)
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Their Swamp", "Basic Land — Swamp", "", false)
	if c := effectstest.GaLayered(t, g, kavu); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("with Forest and Island it is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB032TerritorialKavuAttackRummagesOrExilesAGraveyardCard(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		oldB032Library(me, 10)
		effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
		effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
		kavu := effectstest.B12Push(g, me.ID, "Territorial Kavu", "Creature — Kavu", oldB032Kavu, 2, 2)
		victim := oldB032Graveyard(opp, game.Card{Name: "Their Card", TypeLine: "Sorcery"})
		fodder := effectstest.HandCardFull(me, "Fodder", "Sorcery", "{1}", "", nil)
		hand := me.Hand.Size()
		effectstest.DeclareAttack(t, g, opp.ID, kavu)
		effectstest.Cmdr2S43ChooseMode(t, g, me.ID, mode)
		if mode == 1 {
			effectstest.PickTriggerTarget(t, g, me.ID, victim)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if mode == 0 {
			effectstest.AnswerDiscard(t, g, me.ID, fodder)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(fodder) || me.Hand.Size() != hand {
				t.Errorf("rummage: discarded %v, hand %d (was %d), want it discarded and one drawn", me.Graveyard.Contains(fodder), me.Hand.Size(), hand)
			}
		} else if opp.Graveyard.Contains(victim) {
			t.Error("the graveyard card was not exiled")
		}
	}
}

// --- Scour the Desert ----------------------------------------------------------

func TestOldB032ScourTheDesertMakesBirdsEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := oldB032Graveyard(me, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4})
	effectstest.CastCatalogSpell(t, g, "Scour the Desert", "Sorcery", oldB032Scour, effectstest.CardTargets(big))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(big) {
		t.Error("the creature card was not exiled")
	}
	if n := effectstest.CountOnBattlefield(g, "Bird", me.ID); n != 4 {
		t.Errorf("%d Birds, want 4 (the exiled card's toughness)", n)
	}
}

// --- Fairgrounds Patrol --------------------------------------------------------

func TestOldB032FairgroundsPatrolMakesAThopterFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := oldB032Graveyard(me, game.Card{Name: "Fairgrounds Patrol", TypeLine: "Creature — Human Soldier", OracleID: oldB032Patrol, Power: 2, Toughness: 1})
	if err := effectstest.HobB4Activate(t, g, me.ID, id, 0, "{C}{W}", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(id) {
		t.Error("the card was not exiled as a cost")
	}
	if n := effectstest.CountOnBattlefield(g, "Thopter", me.ID); n != 1 {
		t.Errorf("%d Thopters, want 1", n)
	}
	id2 := oldB032Graveyard(me, game.Card{Name: "Fairgrounds Patrol", TypeLine: "Creature — Human Soldier", OracleID: oldB032Patrol, Power: 2, Toughness: 1})
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.HobB4Activate(t, g, me.ID, id2, 0, "{C}{W}", nil); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Kaleidoscorch -------------------------------------------------------------

func TestOldB032KaleidoscorchDealsDamageEqualToColorsSpent(t *testing.T) {
	for _, tc := range []struct {
		pool string
		want int
	}{{"RR", 1}, {"RG", 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, tc.pool)
		life := opp.Life
		effectstest.CastFromHandForTest(t, g, me, "Kaleidoscorch", "Sorcery", "{1}{R}", oldB032Kaleido,
			game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
		effectstest.PassPriorityAroundTable(t, g)
		if got := life - opp.Life; got != tc.want {
			t.Errorf("paid %s: %d damage, want %d", tc.pool, got, tc.want)
		}
	}
}

func TestOldB032KaleidoscorchFlashbackCountsTheColorsOfTheFlashbackPayment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	id := oldB032Graveyard(me, game.Card{Name: "Kaleidoscorch", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: oldB032Kaleido})
	effectstest.FloatForTest(g, me, "RGWUB")
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 5 {
		t.Errorf("%d damage from five colours, want 5", got)
	}
}

// --- Dihada's Ploy ---------------------------------------------------------------

func TestOldB032DihadasPloyGainsLifePerCardDiscardedThisTurn(t *testing.T) {
	for _, earlier := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		oldB032Library(me, 10)
		if earlier {
			effectstest.HandCardFull(me, "Earlier", "Sorcery", "{1}", "", nil)
			effectstest.B30DiscardOne(t, g, me.ID)
		}
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Dihada's Ploy", "Instant", oldB032Ploy, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.DiscardChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no discard prompt after drawing two")
		}
		effectstest.AnswerDiscard(t, g, me.ID, me.Hand.Cards[0].InstanceID)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if earlier {
			want = 2
		}
		if got := me.Life - life; got != want {
			t.Errorf("earlier=%v: gained %d, want %d", earlier, got, want)
		}
	}
}

func TestOldB032DihadasPloyJumpStartDiscardsAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB032Library(me, 10)
	id := oldB032Graveyard(me, game.Card{Name: "Dihada's Ploy", TypeLine: "Instant", ManaCost: "{1}{U}{B}", OracleID: oldB032Ploy})
	pitch := effectstest.HandCardFull(me, "Pitch", "Sorcery", "{1}", "", nil)
	effectstest.FloatForTest(g, me, "CUB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "jump-start", AltCostIDs: []uuid.UUID{pitch}, Strict: true}); err != nil {
		t.Fatalf("jump-start: %v", err)
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the jump-start discard was not paid")
	}
}

// --- Calibrated Blast --------------------------------------------------------------

func TestOldB032CalibratedBlastDealsDamageEqualToTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Bottom", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}{G}", Layout: "normal", Owner: me.ID, Controller: me.ID})
	land := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Layout: "normal", Owner: me.ID, Controller: me.ID})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Calibrated Blast", "Instant", oldB032Blast, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("%d damage, want 3 (the revealed Spell's mana value)", got)
	}
	n := len(me.Library.Cards)
	if n != 3 {
		t.Fatalf("library holds %d cards, want 3", n)
	}
	for _, id := range []uuid.UUID{land, spell} {
		in := false
		for _, c := range me.Library.Cards[:2] {
			if c.InstanceID == id {
				in = true
			}
		}
		if !in {
			t.Errorf("a revealed card is not at the bottom of the library")
		}
	}
}

// --- Road // Ruin ------------------------------------------------------------------

func TestOldB032RoadPutsABasicLandOntoTheBattlefieldTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	effectstest.CastCatalogSpell(t, g, "Road", "Instant", oldB032RoadRuin, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, ids[1])
	if !ok || !c.Tapped {
		t.Errorf("Forest on battlefield %v, tapped %v: want it there and tapped", ok, c.Tapped)
	}
}

func TestOldB032RuinDamagesTheTargetCreatureByTheLandsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	}
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Ruin", "Sorcery", oldB032RoadRuin+"#1", effectstest.CardTargets(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage (three lands) did not kill the 3/3")
	}
}
