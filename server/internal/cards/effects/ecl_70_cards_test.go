package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ecl_70_cards_test.go — Lorwyn Eclipsed, batch 70 (#4458).

const (
	ecl70LoyalistOracle  = "b2c30683-cc7f-4341-add1-b08cb9b93eeb"
	ecl70NoggleOracle    = "a2f33166-6cfc-4dc5-8edc-088bf40f06b5"
	ecl70FeastlingOracle = "faf2f368-a7e7-4427-bc03-929f294b1e03"
	ecl70SanarOracle     = "2676ba6d-e7a7-48cf-9c86-6e2a11f03096"
	ecl70UrchinOracle    = "86bd332f-ba61-49a4-ad31-de4d7e3638c6"
	ecl70GroveOracle     = "96c4b51c-8a10-428c-a948-e1ba9c652bfd"
	ecl70CommandOracle   = "7fe5baab-b182-4cd9-bea5-912ce9b1464a"
	ecl70TamOracle       = "096a878e-cff0-430a-90b5-031ecb9bfe1b"
	ecl70LieutOracle     = "50dd9e17-6dae-4ff6-8a38-50088c71283f"
	ecl70TravelersOracle = "341cf69a-246c-4eb5-910c-18c13a4c3dce"
)

func TestEcl70Batch70AreRegisteredFull(t *testing.T) {
	ecl67Registered(t, map[string]string{
		ecl70LoyalistOracle:  "Morcant's Loyalist",
		ecl70NoggleOracle:    "Noggle Robber",
		ecl70FeastlingOracle: "Prideful Feastling",
		ecl70SanarOracle:     "Sanar, Innovative First-Year",
		ecl70UrchinOracle:    "Shadow Urchin",
		ecl70GroveOracle:     "Stoic Grove-Guide",
		ecl70CommandOracle:   "Sygg's Command",
		ecl70TamOracle:       "Tam, Mindful First-Year",
		ecl70LieutOracle:     "Thoughtweft Lieutenant",
		ecl70TravelersOracle: "Twinflame Travelers",
	})
}

// --- Morcant's Loyalist ------------------------------------------------------

func TestEcl70MorcantsLoyalistPumpsOtherElvesAndReturnsAnElfWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	loyalist := pushCatalogPermanent(g, me.ID, "Morcant's Loyalist", "Creature — Elf Warrior", ecl70LoyalistOracle, false)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf Scout", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirElf := b12Creature(g, opp.ID, "Their Elf", "Creature — Elf Scout", 1, 1)
	if p, tt := ptOf(t, g, elf); p != 2 || tt != 2 {
		t.Errorf("your other Elf is %d/%d, want 2/2", p, tt)
	}
	if p, _ := ptOf(t, g, bear); p != 2 {
		t.Error("a non-Elf was pumped")
	}
	if p, _ := ptOf(t, g, theirElf); p != 1 {
		t.Error("an opponent's Elf was pumped")
	}
	if p, _ := ptOf(t, g, loyalist); p > 1 {
		// The catalog test card's own power, nothing added by its anthem.
		t.Error("the Loyalist pumped itself")
	}
	deadElf := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: deadElf, Name: "Dead Elf", TypeLine: "Creature — Elf Scout", Owner: me.ID, Controller: me.ID})
	})
	b27Kill(g, loyalist)
	pickCard(t, g, me.ID, deadElf)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(deadElf) {
		t.Error("the Elf card was not returned to hand")
	}
}

// --- Noggle Robber -----------------------------------------------------------

func TestEcl70NoggleRobberMakesATreasureWhenItEntersAndWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Noggle Robber", "Creature — Noggle Rogue", ecl70NoggleOracle, 3, 3, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 1 {
		t.Fatalf("%d Treasures after entering, want 1", n)
	}
	b27Kill(g, id)
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Treasure"); n != 2 {
		t.Errorf("%d Treasures after dying, want 2", n)
	}
}

// --- Prideful Feastling ------------------------------------------------------

func TestEcl70PridefulFeastlingHasChangelingAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	advanceToMain(t, g)
	id := ecl6Cast(t, g, me, "Prideful Feastling", "Creature — Shapeshifter", ecl70FeastlingOracle, 2, 2, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"changeling", "lifelink"} {
		if !hasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Sanar -------------------------------------------------------------------

func TestEcl70SanarExilesOneCardPerColorAndAllowsCastingThem(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Sanar, Innovative First-Year",
		TypeLine: "Legendary Creature — Goblin Sorcerer", OracleID: ecl70SanarOracle, Colors: []string{"U", "R"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	red := game.Card{Name: "Red Spell", TypeLine: "Instant", ManaCost: "{R}", Colors: []string{"R"}}
	blue := game.Card{Name: "Blue Spell", TypeLine: "Instant", ManaCost: "{U}", Colors: []string{"U"}}
	green := game.Card{Name: "Green Spell", TypeLine: "Instant", ManaCost: "{G}", Colors: []string{"G"}}
	// Top first: red, a land, blue, green, deep. X is 2 (blue and red among
	// Sanar's colors): nonland cards revealed are red and blue; the land is
	// revealed too; the green card stays unrevealed.
	ids := mm2Library(me, red, game.Card{Name: "Island", TypeLine: "Basic Land — Island"}, blue, green, mm2Spell("Deep"))
	advanceTo(t, g, game.StepPrecombatMain)
	monSettle(t, g)
	// One prompt per color in WUBRG order: blue first, then red.
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the first color")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[2] {
		t.Fatalf("blue prompt offered %v, want the blue card only", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, ids[2])
	monSettle(t, g)
	answerChooseCards(t, g, me.ID, ids[0])
	monSettle(t, g)
	for _, id := range []uuid.UUID{ids[0], ids[2]} {
		if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneExile {
			t.Errorf("card %v is not exiled", id)
		}
	}
	// A normal cast: with no mana it is refused, with the mana it works.
	me.ManaPool = nil
	if err := g.CastSpell(me.ID, ids[2], game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("an exiled card was cast for free")
	}
	b06AddMana(me, "U")
	if err := g.CastSpell(me.ID, ids[2], game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("casting the exiled blue card with {U}: %v", err)
	}
}

// --- Shadow Urchin -----------------------------------------------------------

func TestEcl70ShadowUrchinBlightsWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	urchin := pushCatalogPermanent(g, me.ID, "Shadow Urchin", "Creature — Ouphe", ecl70UrchinOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == urchin {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			}
		}
	})
	advanceToMain(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(urchin, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	monSettle(t, g)
	passPriorityAroundTable(t, g)
	if c, _ := battlefieldCard(g, urchin); c.Counters[game.CounterMinusOne] != 1 {
		t.Errorf("%d -1/-1 counters, want 1", c.Counters[game.CounterMinusOne])
	}
}

func TestEcl70ShadowUrchinExilesAsManyCardsAsTheDeadCreatureHadCounters(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Shadow Urchin", "Creature — Ouphe", ecl70UrchinOracle, false)
	counted := b12Creature(g, me.ID, "Counted", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	if err := g.AddCounter(counted, game.CounterPlusOne, 2); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	ids := mm2Library(me, mm2Spell("A"), mm2Spell("B"), mm2Spell("C"), mm2Spell("D"))
	advanceToMain(t, g)
	b15Destroy(t, g, plain)
	passPriorityAroundTable(t, g)
	if zone := g.FindCardZoneForEffect(ids[0]); zone == nil || zone.Kind != game.ZoneLibrary {
		t.Fatal("a creature with no counters made it exile")
	}
	b15Destroy(t, g, counted)
	passPriorityAroundTable(t, g)
	for i, id := range ids {
		zone := g.FindCardZoneForEffect(id)
		wantExile := i < 2
		if zone == nil || (zone.Kind == game.ZoneExile) != wantExile {
			t.Errorf("card %d: in zone %v, want exiled=%v", i, ecl70ZoneKind(zone), wantExile)
		}
	}
}

func ecl70ZoneKind(z *game.Zone) game.ZoneKind {
	if z == nil {
		return ""
	}
	return z.Kind
}

// --- Stoic Grove-Guide -------------------------------------------------------

func TestEcl70StoicGroveGuideMakesAnElfFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Stoic Grove-Guide", TypeLine: "Creature — Elf Druid",
			OracleID: ecl70GroveOracle, Owner: me.ID, Controller: me.ID})
	})
	advanceToMain(t, g)
	b06AddMana(me, "C", "G")
	b16Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Elf"); n != 1 {
		t.Errorf("%d Elf tokens, want 1", n)
	}
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneExile {
		t.Error("the card was not exiled as a cost")
	}
}

// --- Sygg's Command ----------------------------------------------------------

func TestEcl70SyggsCommandDrawsAndTapsAndStuns(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	victim := b12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	seedLibrary(me, "a", "b")
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	ecl6Cast(t, g, me, "Sygg's Command", "Kindred Sorcery — Merfolk", ecl70CommandOracle, 0, 0, game.CastSpellParams{
		Modes: []int{2, 3},
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: me.ID, Mode: 0},
			{Kind: game.TargetCard, ID: victim, Mode: 1},
		}})
	passPriorityAroundTable(t, g)
	if got := handSizeOf(g, me) - before; got != 1 {
		t.Errorf("hand grew by %d, want the one drawn card", got)
	}
	c, _ := battlefieldCard(g, victim)
	if !c.Tapped || c.Counters[game.CounterStun] != 1 {
		t.Errorf("the victim tapped=%v stun=%d, want tapped with a stun counter", c.Tapped, c.Counters[game.CounterStun])
	}
}

func TestEcl70SyggsCommandCopiesAMerfolkAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	merfolk := b12Creature(g, me.ID, "Merrow", "Creature — Merfolk Soldier", 2, 2)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Sygg's Command", "Kindred Sorcery — Merfolk", ecl70CommandOracle, 0, 0, game.CastSpellParams{
		Modes: []int{0, 1},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: merfolk, Mode: 0},
			{Kind: game.TargetPlayer, ID: me.ID, Mode: 1},
		}})
	passPriorityAroundTable(t, g)
	if n := len(battlefieldIDsNamed(g, "Merrow")); n != 2 {
		t.Errorf("%d Merrow on the battlefield, want the original and a copy", n)
	}
	if !hasEffectiveKeyword(t, g, other, "lifelink") {
		t.Error("the Bear has no lifelink")
	}
}

// --- Tam ---------------------------------------------------------------------

func TestEcl70TamGivesHexproofFromEachOfAnotherCreaturesColors(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	tam := pushCatalogPermanent(g, me.ID, "Tam, Mindful First-Year", "Legendary Creature — Gorgon Wizard", ecl70TamOracle, false)
	rg := b27Push(g, me.ID, "Red Green", "Creature — Bear", "", "{R}{G}", 2, 2, "R", "G")
	colorless := b12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	advanceToMain(t, g)
	for _, kw := range []string{"hexproof from red", "hexproof from green"} {
		if !hasEffectiveKeyword(t, g, rg, kw) {
			t.Errorf("the two-color creature lacks %q", kw)
		}
	}
	if hasEffectiveKeyword(t, g, rg, "hexproof from blue") {
		t.Error("hexproof from a color the creature is not")
	}
	if hasEffectiveKeyword(t, g, colorless, "hexproof from red") {
		t.Error("a colorless creature got hexproof from a color")
	}
	if hasEffectiveKeyword(t, g, tam, "hexproof from white") {
		t.Error("Tam gave itself hexproof")
	}
	// {T}: the colorless creature becomes all colors: hexproof from every color.
	b16Activate(t, g, me.ID, tam, 0, game.ActivateAbilityParams{Targets: b16TargetCard(colorless)})
	passPriorityAroundTable(t, g)
	for _, kw := range []string{"hexproof from white", "hexproof from blue", "hexproof from black", "hexproof from red", "hexproof from green"} {
		if !hasEffectiveKeyword(t, g, colorless, kw) {
			t.Errorf("after becoming all colors it lacks %q", kw)
		}
	}
	ecl6NextTurn(t, g)
	if hasEffectiveKeyword(t, g, colorless, "hexproof from white") {
		t.Error("all colors outlasted the turn")
	}
}

// --- Thoughtweft Lieutenant --------------------------------------------------

func TestEcl70ThoughtweftLieutenantTriggersForItselfAndOtherKithkin(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := ecl6Seats(g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	ecl6Cast(t, g, me, "Thoughtweft Lieutenant", "Creature — Kithkin Soldier", ecl70LieutOracle, 2, 2, game.CastSpellParams{})
	ecl6ResolveSpell(t, g)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if p, tt := ptOf(t, g, bear); p != 3 || tt != 3 {
		t.Errorf("the Bear is %d/%d, want 3/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("no trample")
	}
	// Another Kithkin of mine entering triggers it; an opponent's does not.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 green and white Kithkin"), 1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's Kithkin entering triggered the Lieutenant")
	}
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 green and white Kithkin"), 1) })
	monSettle(t, g)
	if latestPickTarget(g, me.ID) == nil {
		t.Error("my own Kithkin entering did not trigger the Lieutenant")
	}
}

// --- Twinflame Travelers -----------------------------------------------------

func TestEcl70TwinflameTravelersDoublesAnotherElementalsTriggers(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := ecl6Seats(g)
	pushCatalogPermanent(g, me.ID, "Twinflame Travelers", "Creature — Elemental Sorcerer", ecl70TravelersOracle, false)
	advanceToMain(t, g)
	before := handSizeOf(g, me)
	seedLibrary(me, "a", "b", "c", "d")
	// Another Elemental with an enters trigger: Mistmeadow-style draw via the
	// Noggle Robber is not an Elemental, so use a real Elemental: Luminollusk.
	ecl6Cast(t, g, me, "Luminollusk", "Creature — Elemental", ecl67LuminolluskOracle, 3, 3, game.CastSpellParams{})
	b27Push(g, me.ID, "Red", "Creature — Bear", "", "{1}{R}", 2, 2, "R")
	life := lifeOf(g, me.ID)
	passPriorityAroundTable(t, g)
	// Luminollusk's life gain triggers twice: 2 x (colors among my permanents).
	colors := len(colorsAmongPermanentsYouControl(g, me.ID))
	if got := lifeOf(g, me.ID) - life; got != 2*colors {
		t.Errorf("gained %d life with %d colors, want the trigger twice (%d)", got, colors, 2*colors)
	}
	_ = before
}
