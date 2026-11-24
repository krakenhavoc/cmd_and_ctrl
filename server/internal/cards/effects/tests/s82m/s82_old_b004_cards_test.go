package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b004_cards_test.go — S82 OLD batch 4 (#5462, Secret Lair Drop): card
// draw and library, last batch. Each test plays the real catalog entry.

const (
	oldB004Tadeas  = "b4b5b928-8b98-4f65-9f15-212859874fe4"
	oldB004Shadow  = "b4b724bb-4c1c-4002-8eb8-bc8dd3f88e62"
	oldB004Casal   = "b7bdb688-f8ee-4d22-a679-37a2ecd390a1"
	oldB004Nezahal = "c45e96cb-2539-4e00-9fff-782a182c2f4b"
	oldB004Sway    = "c4fccdd5-5b51-46b9-9937-9167c108743d"
	oldB004Realms  = "e21c8fc6-d4ef-42b6-b11e-d9c931da1387"
	oldB004Demonic = "93056597-b964-421f-be2f-e92abef1c2a4#1"
)

func TestOldB004EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB004Tadeas:       "Tadeas, Juniper Ascendant",
		oldB004Shadow:       "Shadow of Doubt",
		oldB004Casal:        "Casal, Lurkwood Pathfinder",
		oldB004Casal + "#1": "Casal, Pathbreaker Owlbear",
		oldB004Nezahal:      "Nezahal, Primal Tide",
		oldB004Sway:         "Sway of the Stars",
		oldB004Realms:       "Realms Uncharted",
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

func oldB004Library(p *game.Player, n int, name, typeLine string) []uuid.UUID {
	p.Library.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

// --- Shadow of Doubt ---------------------------------------------------------

func TestOldB004ShadowOfDoubtBarsLibrarySearchesForTheTurnAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB004Library(me, 10, "Forest", "Basic Land — Forest")
	before := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Shadow of Doubt", "Instant", "{U/B}{U/B}", oldB004Shadow)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("Shadow of Doubt drew %d net of the cast, want 1", got)
	}
	// Demonic Tutor finds nothing this turn: no prompt, hand unchanged.
	hand := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Demonic Tutor", "Sorcery", "{1}{B}", oldB004Demonic)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchPromptFor(g, me.ID) != nil {
		t.Fatal("a library search was offered after Shadow of Doubt")
	}
	if me.Hand.Size() != hand {
		t.Errorf("the barred search still found a card: hand %d (was %d)", me.Hand.Size(), hand)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventSearchLibrary {
			t.Error("a search event was emitted for a barred search")
		}
	}
}

func TestOldB004ShadowOfDoubtEndsWithTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB004Library(me, 10, "Forest", "Basic Land — Forest")
	effectstest.CastWithCost(t, g, "Shadow of Doubt", "Instant", "{U/B}{U/B}", oldB004Shadow)
	effectstest.PassPriorityAroundTable(t, g)
	// A later turn of mine searches normally.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.CastWithCost(t, g, "Demonic Tutor", "Sorcery", "{1}{B}", oldB004Demonic)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchPromptFor(g, me.ID) == nil {
		t.Error("the search bar outlived the turn")
	}
}

// --- Tadeas, Juniper Ascendant -----------------------------------------------

func TestOldB004TadeasIsHexproofUnlessAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tadeas := effectstest.PushCatalogPermanent(g, me.ID, "Tadeas, Juniper Ascendant", "Legendary Creature — Human Monk", oldB004Tadeas, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, tadeas, "hexproof") {
		t.Error("Tadeas is not hexproof while idle")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(tadeas, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, tadeas, "hexproof") {
		t.Error("Tadeas is hexproof while attacking")
	}
}

func TestOldB004TadeasUntapsAReachAttackerAndGivesItSkulk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Tadeas, Juniper Ascendant", "Legendary Creature — Human Monk", oldB004Tadeas, false)
	reach := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spider", TypeLine: "Creature — Spider",
		Power: 2, Toughness: 4, Keywords: []string{"reach"}, Owner: me.ID, Controller: me.ID})
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	for _, id := range []uuid.UUID{reach, plain} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, reach); c.Tapped {
		t.Error("the reach attacker stayed tapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, plain); !c.Tapped {
		t.Error("the attacker without reach was untapped")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, reach, "skulk") {
		t.Error("the reach attacker did not gain skulk this combat")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, plain, "skulk") {
		t.Error("an attacker without reach gained skulk")
	}
}

func TestOldB004TadeasDrawsOncePerPlayerDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB004Library(me, 8, "Forest", "Basic Land — Forest")
	tadeas := effectstest.PushCatalogPermanent(g, me.ID, "Tadeas, Juniper Ascendant", "Legendary Creature — Human Monk", oldB004Tadeas, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	effectstest.DealCombatDamageToPlayer(g, tadeas, opp.ID, 1)
	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d for two creatures hitting one player, want 1", got)
	}
}

// --- Nezahal, Primal Tide ----------------------------------------------------

func TestOldB004NezahalStatics(t *testing.T) {
	spec, _ := effects.Lookup(oldB004Nezahal)
	if !spec.CantBeCountered || !spec.NoMaxHandSize {
		t.Errorf("CantBeCountered = %v, NoMaxHandSize = %v; want both", spec.CantBeCountered, spec.NoMaxHandSize)
	}
}

func TestOldB004NezahalDrawsForAnOpponentsNoncreatureSpellOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB004Library(me, 8, "Forest", "Basic Land — Forest")
	effectstest.PushCatalogPermanent(g, me.ID, "Nezahal, Primal Tide", "Legendary Creature — Elder Dinosaur", oldB004Nezahal, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	before := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Their Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before {
		t.Fatal("drew for an opponent's creature spell")
	}
	effectstest.CastWithCost(t, g, "Their Sorcery", "Sorcery", "{1}{R}", "")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d for an opponent's noncreature spell, want 1", got)
	}
	_ = opp
}

func TestOldB004NezahalDiscardThreeExilesItAndItReturnsTappedAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nezahal := effectstest.PushCatalogPermanent(g, me.ID, "Nezahal, Primal Tide", "Legendary Creature — Elder Dinosaur", oldB004Nezahal, false)
	var discards []uuid.UUID
	g.WithWriteLock(func() {
		me.Hand.Cards = nil
		for i := 0; i < 3; i++ {
			c := game.Card{InstanceID: uuid.New(), Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
			me.Hand.PushTop(c)
			discards = append(discards, c.InstanceID)
		}
	})
	// Two cards are not enough.
	if err := g.ActivateCatalogAbility(me.ID, nezahal, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: discards[:2]}); err == nil {
		t.Fatal("two discards paid for three")
	}
	effectstest.Pr7Activate(t, g, me.ID, nezahal, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: discards})
	if g.Battlefield.Contains(nezahal) {
		t.Fatal("Nezahal is still on the battlefield")
	}
	if me.Hand.Size() != 0 {
		t.Errorf("hand = %d after discarding three", me.Hand.Size())
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	var back *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].OracleID == oldB004Nezahal {
			back = &g.Battlefield.Cards[i]
		}
	}
	if back == nil {
		t.Fatal("Nezahal did not return at the next end step")
	}
	if !back.Tapped {
		t.Error("Nezahal returned untapped")
	}
}

// --- Casal -------------------------------------------------------------------

var oldB004CasalCard = effectstest.IwCard{Oracle: oldB004Casal, Front: "Casal, Lurkwood Pathfinder", FrontType: "Legendary Creature — Tiefling Druid",
	Cost: "{3}{G}", FrontText: "Vigilance\nWhen Casal enters, search your library for a Forest card, put it onto the battlefield tapped, then shuffle.\nWhenever Casal attacks, you may pay {1}{G}. If you do, transform her.",
	Fp: "3", Ft: "3", Back: "Casal, Pathbreaker Owlbear", BackText: "Vigilance, trample", Bp: "6", Bt: "6", Keywords: []string{"Vigilance"}}

func TestOldB004CasalFetchesATappedForestWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	forest := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: forest, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.ImportAndCast(t, g, oldB004CasalCard.Row(), me)
	effectstest.MonSettle(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	effectstest.MonSettle(t, g)
	c, ok := effectstest.BattlefieldCard(g, forest)
	if !ok || !c.Tapped {
		t.Errorf("the fetched Forest: on battlefield %v, tapped %v; want both", ok, c.Tapped)
	}
}

func TestOldB004CasalAttackPaymentTransformsHer(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.IwPut(t, g, oldB004CasalCard, me)
		effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"})
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.MonSettle(t, g)
		want := 0
		if pay {
			want = 1
		}
		if got := effectstest.IwFace(g, id); got != want {
			t.Errorf("pay %v: face %d, want %d", pay, got, want)
		}
	}
}

func TestOldB004CasalOwlbearPumpsOtherLegendsAndTransformsBackAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	legend := effectstest.B12Push(g, me.ID, "Some Legend", "Legendary Creature — Test", "", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	casal := effectstest.IwPut(t, g, oldB004CasalCard, me)
	effectstest.IwTransform(t, g, casal)
	effectstest.MonSettle(t, g)
	if p, tg := effectstest.B38PT(t, g, legend); p != 4 || tg != 4 {
		t.Errorf("the other legend is %d/%d, want 4/4", p, tg)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, legend, "trample") {
		t.Error("the other legend did not gain trample")
	}
	if p, _ := effectstest.B38PT(t, g, plain); p != 2 {
		t.Errorf("a nonlegendary creature got the bonus: power %d", p)
	}
	if p, _ := effectstest.B38PT(t, g, casal); p != 6 {
		t.Errorf("Casal herself is %d power, want 6", p)
	}
	// At my next upkeep she turns back.
	for i := 0; i < 8 && effectstest.IwFace(g, casal) == 1; i++ {
		effectstest.IwNextUpkeep(t, g, nil)
	}
	if effectstest.IwFace(g, casal) != 0 {
		t.Error("Casal did not transform back at her controller's upkeep")
	}
}

// --- Realms Uncharted --------------------------------------------------------

func TestOldB004RealmsUnchartedOpponentBinsTwoTheRestGoToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me.Library.Cards = nil
	lands := effectstest.SeedTypedLibrary(me, "Land", "Grove", "Marsh", "Peak", "Strand", "Grove")
	hand := me.Hand.Size()
	effectstest.CastWithCost(t, g, "Realms Uncharted", "Instant", "{2}{G}", oldB004Realms)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{lands[0], lands[4]}); err == nil {
		t.Fatal("two lands with the same name were found")
	}
	found := lands[:4]
	if err := g.ResolveSearchLibrary(c.ID, me.ID, found); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	}
	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked to choose: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{found[0], found[1]}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range found[:2] {
		if !me.Graveyard.Contains(id) {
			t.Errorf("chosen land %s is not in the graveyard", id)
		}
	}
	for _, id := range found[2:] {
		if !me.Hand.Contains(id) {
			t.Errorf("land %s did not go to the hand", id)
		}
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d net of the cast, want 2", got)
	}
	if !me.Library.Contains(lands[4]) {
		t.Error("the unfound land left the library")
	}
}

// --- Sway of the Stars -------------------------------------------------------

func TestOldB004SwayOfTheStarsShufflesEverythingBackDrawsSevenAndSetsLifeToSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB004Library(me, 12, "Forest", "Basic Land — Forest")
	oldB004Library(opp, 12, "Island", "Basic Land — Island")
	for _, p := range []*game.Player{me, opp} {
		p := p
		g.WithWriteLock(func() {
			p.Hand.Cards = nil
			p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: p.Name + " Hand", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
		})
		effectstest.PushGraveyardPermanent(p, p.Name+" Yard", "Creature — Test", "{1}")
		effectstest.B12Creature(g, p.ID, p.Name+" Creature", "Creature — Test", 2, 2)
	}
	// A permanent I own but an opponent controls goes to MY library.
	stolen := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen", TypeLine: "Artifact", Owner: me.ID, Controller: opp.ID})
	g.WithWriteLock(func() { me.Life, opp.Life = 13, 40 })
	myLibrary, oppLibrary := me.Library.Size(), opp.Library.Size()
	effectstest.CastWithCost(t, g, "Sway of the Stars", "Sorcery", "{8}{U}{U}", oldB004Sway)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, opp} {
		if p.Hand.Size() != 7 {
			t.Errorf("%s holds %d cards, want 7", p.Name, p.Hand.Size())
		}
		if p.Life != 7 {
			t.Errorf("%s has %d life, want 7", p.Name, p.Life)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.IsCreature() || c.InstanceID == stolen {
			t.Errorf("%s stayed on the battlefield", c.Name)
		}
	}
	// 12 library + 1 hand + 1 graveyard + 1 creature (+ the stolen artifact for me) - 7 drawn.
	if got := me.Library.Size(); got != myLibrary+3+1-7 {
		// the cast Sway itself was in my hand and cast: it is on the stack, not shuffled
		t.Errorf("my library holds %d cards, want %d", got, myLibrary+3+1-7)
	}
	if got := opp.Library.Size(); got != oppLibrary+3-7 {
		t.Errorf("the opponent's library holds %d cards, want %d", got, oppLibrary+3-7)
	}
}
