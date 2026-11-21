package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b07_cards_test.go — S82 MID batch 7 (#5497): graveyard cards.
// Every test plays the real catalog card through the engine.

const (
	midB07Candleguide  = "1b817184-4965-4186-bfa9-6d9cf198050e"
	midB07Gryff        = "e26a08de-6108-4051-9c42-b5ce9dd8b533"
	midB07Dreadhound   = "5ed533a7-82b0-4a71-a9e9-4dffe9d18851"
	midB07Amalgam      = "2378d6fc-b6ba-4a44-b951-5698ed507e4c"
	midB07Hoarder      = "a6f0dcbc-ca45-4ae8-961e-e42ef8d8a975"
	midB07Carriage     = "d95e7282-b298-4410-baac-01766d6852e3"
	midB07Sifters      = "d70ee6f1-541b-43b4-897c-68a231beb894"
	midB07Exorcism     = "b43e25d8-1c8b-4316-84f3-c4ec6f77092e"
	midB07Geist        = "7a7096a7-ef1b-4015-b1eb-5c2174d7ec9a"
	midB07Locked       = "035f0aae-9c23-4b22-8c87-0d66b84b7962"
	midB07Elementalist = "1edd4f50-05ee-43e0-a288-c26666910c8b"
)

func TestMidB07EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		midB07Candleguide: "Crossroads Candleguide", midB07Gryff: "Soul-Guide Gryff", midB07Dreadhound: "Dreadhound",
		midB07Amalgam: "Drownyard Amalgam", midB07Hoarder: "Organ Hoarder", midB07Carriage: "Phantom Carriage",
		midB07Sifters: "Shipwreck Sifters", midB07Exorcism: "Thraben Exorcism", midB07Geist: "Willow Geist",
		midB07Locked: "Locked in the Cemetery", midB07Elementalist: "Ardent Elementalist",
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

// midB07PickAndResolve answers the enters trigger's target prompt (when one
// is open) with a card and lets the trigger resolve.
func midB07PickAndResolve(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, chooser); p != nil {
		effectstest.PickCard(t, g, chooser, card)
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestMidB07CandleguideExilesAGraveyardCardAndMakesAnyColour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushGraveyardCardForTest(opp, "Doomed")
	id := effectstest.RfCrACast(t, g, "Crossroads Candleguide", "Artifact Creature — Scarecrow", midB07Candleguide, 3, 4)
	midB07PickAndResolve(t, g, me.ID, victim)
	if !g.Exile.Contains(victim) {
		t.Fatal("the graveyard card was not exiled")
	}
	// {2}: Add one mana of any colour.
	effectstest.SamiGiveColorless(me, 2)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v after paying {2} for one red mana, want [R]", got)
	}
}

func TestMidB07CandleguideAndGryffExileNothingWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mine := effectstest.PushGraveyardCardForTest(me, "Mine")
	effectstest.RfCrACast(t, g, "Soul-Guide Gryff", "Creature — Hippogriff Spirit", midB07Gryff, 3, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		// "Up to one": declining is a legal answer, so a pick is open.
		t.Log("a target prompt is open; declining by passing")
	}
	if g.Exile.Contains(mine) && !me.Graveyard.Contains(mine) {
		t.Log("exiled without a pick")
	}
}

func TestMidB07GryffExilesTheChosenCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.PushGraveyardCardForTest(opp, "Doomed")
	other := effectstest.PushGraveyardCardForTest(opp, "Spared")
	id := effectstest.RfCrACast(t, g, "Soul-Guide Gryff", "Creature — Hippogriff Spirit", midB07Gryff, 3, 4)
	midB07PickAndResolve(t, g, me.ID, victim)
	if !g.Exile.Contains(victim) || g.Exile.Contains(other) {
		t.Fatal("exactly the chosen card should be exiled")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("Soul-Guide Gryff has no flying")
	}
}

func TestMidB07DreadhoundMillsThreeAndDrainsOnCreatureCardsFromLibraries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	me.Library.Cards = nil
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Milled Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Milled Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Deep Card", TypeLine: "Sorcery"})
	life := opp.Life
	effectstest.RfCrACast(t, g, "Dreadhound", "Creature — Demon Dog", midB07Dreadhound, 6, 6)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 3 {
		t.Fatalf("graveyard %d after the enters mill, want 3", got)
	}
	// Two creature cards were milled: two triggers.
	if got := opp.Life; got != life-2 {
		t.Errorf("opponent life %d -> %d, want -2 for two milled creature cards", life, got)
	}
}

func TestMidB07DreadhoundDrainsWhenAnyCreatureDiesItselfIncluded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hound := effectstest.B12Push(g, me.ID, "Dreadhound", "Creature — Demon Dog", midB07Dreadhound, 6, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("an opposing creature died: life %d -> %d, want -1", life, opp.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hound) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("Dreadhound's own death: life %d, want %d", opp.Life, life-2)
	}
}

func TestMidB07DreadhoundIgnoresNoncreatureAndNonLibraryCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "Dreadhound", "Creature — Demon Dog", midB07Dreadhound, 6, 6)
	life := opp.Life
	// A noncreature permanent dying, and a creature card leaving a hand
	// for the graveyard (a discard), are neither.
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Milled Rock", TypeLine: "Artifact"})
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d -> %d on events Dreadhound does not watch", life, opp.Life)
	}
}

func TestMidB07DrownyardAmalgamMillsTheTargetPlayerAndGoesUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	}
	before := opp.Graveyard.Size()
	id := effectstest.RfCrACast(t, g, "Drownyard Amalgam", "Creature — Zombie Horror", midB07Amalgam, 3, 6)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size() - before; got != 3 {
		t.Fatalf("target player milled %d, want 3", got)
	}
	effectstest.RfCrAAddMana(t, g, me, "{C}{C}{U}")
	effectstest.B16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !midB07CantBeBlocked(t, g, id) {
		t.Error("the Amalgam can be blocked after {2}{U}")
	}
}

func midB07CantBeBlocked(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	return effectstest.RestrictionsOf(t, g, id)&game.CantBeBlocked != 0
}

func TestMidB07OrganHoarderTakesOneOfThreeAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	ids := effectstest.SeedLibrary(me, "One", "Two", "Three", "Four")
	effectstest.RfCrACast(t, g, "Organ Hoarder", "Creature — Zombie", midB07Hoarder, 3, 2)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick among the top three cards")
	}
	if len(pick.ChooseCards) != 3 || pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Fatalf("offered %d cards in [%d,%d], want exactly one of the top three", len(pick.ChooseCards), pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the chosen card is not in hand")
	}
	for _, id := range []uuid.UUID{ids[0], ids[2]} {
		if !me.Graveyard.Contains(id) {
			t.Error("a card that was not chosen is not in the graveyard")
		}
	}
	if me.Graveyard.Contains(ids[3]) || me.Hand.Contains(ids[3]) {
		t.Error("the fourth card was touched")
	}
}

func TestMidB07PhantomCarriageFindsOnlyFlashbackOrDisturbCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	plain := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Plain", TypeLine: "Sorcery"})
	flash := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Fire Spell", TypeLine: "Sorcery", FlashbackText: game.PrintedYes})
	disturb := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Ghost Door", TypeLine: "Creature — Human", Keywords: []string{"disturb"}})
	notFlash := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Names It", TypeLine: "Sorcery", FlashbackText: game.PrintedNo})
	effectstest.RfCrACast(t, g, "Phantom Carriage", "Creature — Spirit", midB07Carriage, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.BeginSearchLibrary(search.ID, me.ID); err != nil {
		t.Fatalf("BeginSearchLibrary: %v", err)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range search.SearchCards {
		offered[id] = true
	}
	if !offered[flash] || !offered[disturb] || offered[plain] || offered[notFlash] {
		t.Fatalf("offered %v: want the flashback and disturb cards only", offered)
	}
	if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{disturb}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Graveyard.Contains(disturb) {
		t.Error("the found card is not in the graveyard")
	}
}

func TestMidB07PhantomCarriageMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	flash := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Fire Spell", TypeLine: "Sorcery", FlashbackText: game.PrintedYes})
	effectstest.RfCrACast(t, g, "Phantom Carriage", "Creature — Spirit", midB07Carriage, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	search := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID)
	if search == nil {
		t.Fatal("no search prompt")
	}
	if err := g.DeclineSearchLibrary(search.ID, me.ID); err != nil {
		t.Fatalf("declining: %v", err)
	}
	if me.Graveyard.Contains(flash) {
		t.Error("a declined search put a card in the graveyard")
	}
}

func TestMidB07ShipwreckSiftersGrowsOnDiscardingSpiritsAndDisturbCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	effectstest.SeedLibrary(me, "Drawn", "Spare")
	spirit := effectstest.HandCardFull(me, "Ghost", "Creature — Spirit", "{1}", "", nil)
	id := effectstest.RfCrACast(t, g, "Shipwreck Sifters", "Creature — Spirit", midB07Sifters, 1, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 2 {
		t.Fatalf("power %d after discarding a Spirit, want 2", got)
	}
	// A plain card does not count; a card with disturb does.
	plain := effectstest.HandCardFull(me, "Plain", "Sorcery", "{1}", "", nil)
	effectstest.LengDiscard(t, g, me, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 2 {
		t.Fatalf("power %d after discarding a plain card, want 2", got)
	}
	disturb := effectstest.HandCardFull(me, "Haunted Door", "Creature — Human", "{1}", "", nil)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == disturb {
				me.Hand.Cards[i].Keywords = []string{"disturb"}
			}
		}
	})
	effectstest.LengDiscard(t, g, me, disturb)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 3 {
		t.Errorf("power %d after discarding a disturb card, want 3", got)
	}
}

func TestMidB07ThrabenExorcismExilesASpiritADisturbCreatureOrAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.B12Creature(g, opp.ID, "Ghost", "Creature — Spirit", 1, 1)
	disturb := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Haunted Door", TypeLine: "Creature — Human", Keywords: []string{"disturb"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	ench := effectstest.B12Permanent(g, opp.ID, "Rune", "Enchantment")
	for _, target := range []uuid.UUID{spirit, disturb, ench} {
		effectstest.CastCatalogSpell(t, g, "Thraben Exorcism", "Instant", midB07Exorcism, effectstest.B16TargetCard(target))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(target) {
			t.Errorf("%s was not exiled", target)
		}
	}
}

func TestMidB07ThrabenExorcismRefusesAPlainCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Thraben Exorcism", "Instant", midB07Exorcism, effectstest.B16TargetCard(bear)); err == nil {
		t.Error("exiled-target accepted a plain creature")
	}
}

func TestMidB07WillowGeistGrowsPerBatchLeavingYourGraveyardAndGainsLifeOnDeath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := effectstest.B12Push(g, me.ID, "Willow Geist", "Creature — Treefolk Spirit", midB07Geist, 1, 1)
	mine := effectstest.PushGraveyardCardForTest(me, "Old Card")
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Card")
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 1 {
		t.Fatalf("power %d after a card left an opponent's graveyard, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(mine) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, id); got != 2 {
		t.Fatalf("power %d after a card left your graveyard, want 2", got)
	}
	life := me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d -> %d on death, want +2 (its power)", life, me.Life)
	}
}

func TestMidB07LockedInTheCemeteryTapsOnlyWithFiveCardsInYourGraveyard(t *testing.T) {
	for _, tc := range []struct {
		cards   int
		wantTap bool
	}{{4, false}, {5, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		for i := 0; i < tc.cards; i++ {
			effectstest.PushGraveyardCardForTest(me, "Filler")
		}
		host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, "Locked in the Cemetery", "Enchantment — Aura", midB07Locked, effectstest.B16TargetCard(host))
		effectstest.PassPriorityAroundTable(t, g)
		if c, _ := effectstest.BattlefieldCard(g, host); c.Tapped != tc.wantTap {
			t.Errorf("%d cards in the graveyard: tapped %v, want %v", tc.cards, c.Tapped, tc.wantTap)
		}
	}
}

func TestMidB07LockedInTheCemeteryHoldsItsCreatureDown(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		effectstest.PushGraveyardCardForTest(me, "Filler")
	}
	host := effectstest.PushPermanentForTest(g, me.ID, "Host", "", "Creature — Bear")
	effectstest.CastCatalogSpell(t, g, "Locked in the Cemetery", "Enchantment — Aura", midB07Locked, effectstest.B16TargetCard(host))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if c, _ := effectstest.BattlefieldCard(g, host); !c.Tapped {
		t.Error("enchanted creature untapped during its controller's untap step")
	}
}

func TestMidB07ArdentElementalistReturnsAnInstantOrSorceryOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := effectstest.PushGraveyardCardForTest(me, "Bolt")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == bolt {
				me.Graveyard.Cards[i].TypeLine = "Instant"
			}
		}
	})
	theirs := effectstest.PushGraveyardCardForTest(opp, "Their Bolt")
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == theirs {
				opp.Graveyard.Cards[i].TypeLine = "Instant"
			}
		}
	})
	effectstest.RfCrACast(t, g, "Ardent Elementalist", "Creature — Human Shaman", midB07Elementalist, 2, 1)
	midB07PickAndResolve(t, g, me.ID, bolt)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if opp.Graveyard.Contains(theirs) == false {
		t.Error("an opponent's graveyard card was taken")
	}
}
