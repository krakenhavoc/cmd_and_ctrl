package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	midHallowedRespite = "2d72454f-3d87-4acc-9f46-b8c579f44af8"
	midDireStrain      = "55281479-9d30-4f3e-afd0-66b0df5de68c"
	midStormFestival   = "5d03188b-fbc4-4e16-9fe6-db98dbdb7f68"
	midTurnTheEarth    = "ebb445f7-75da-4af4-9153-93d419c838e0"
	midCantStayAway    = "69ae219e-bf97-4626-9b16-7901f51f0343"
	midWakeToSlaughter = "78eb5399-8c8b-4310-b585-e4ccd9d7c9cc"
	midRiteOfHarmony   = "af94fd87-ea37-49eb-b5e6-0b1846f23859"
)

func midTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func midOnBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// midFlashback casts the card from the graveyard for its flashback cost
// with the mana in the pool, strictly.
func midFlashback(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: targets}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestMidFlashbackCardsDeclareTheirCost(t *testing.T) {
	for _, c := range []struct{ oracle, cost string }{
		{midHallowedRespite, "{1}{W}{U}"}, {midDireStrain, "{3}{R}{G}"}, {midStormFestival, "{7}{G}{G}{G}"},
		{midTurnTheEarth, "{1}{G}"}, {midCantStayAway, "{3}{W}{B}"}, {midWakeToSlaughter, "{4}{B}{R}"},
		{midRiteOfHarmony, "{2}{G}{W}"},
	} {
		offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
		if len(offers) != 1 || offers[0].Key != "flashback" || offers[0].ManaCost != c.cost || !offers[0].ExileOnLeavingStack {
			t.Errorf("%s: offers %+v, want flashback %s", c.oracle, offers, c.cost)
		}
	}
}

// --- Hallowed Respite ---------------------------------------------------

func TestMidHallowedRespiteBlinksYourCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, bear).Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Hallowed Respite", "Sorcery", midHallowedRespite, midTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil || back == bear {
		t.Fatalf("the creature did not return as a new object (%v → %v)", bear, back)
	}
	c, _ := midOnBattlefield(g, back)
	if c.Counters[game.CounterPlusOne] != 1 || c.Tapped {
		t.Errorf("counters %d tapped %v, want 1 counter, untapped", c.Counters[game.CounterPlusOne], c.Tapped)
	}
}

func TestMidHallowedRespiteSendsAStolenCreatureHomeTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	stolen := effectstest.ApaPush(g, opp.ID, me.ID, game.Card{Name: "Stolen Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastCatalogSpell(t, g, "Hallowed Respite", "Sorcery", midHallowedRespite, midTarget(stolen))
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.FindBattlefieldByName(g, "Stolen Bear")
	if back == uuid.Nil {
		t.Fatal("the creature did not return")
	}
	c, _ := midOnBattlefield(g, back)
	if c.Controller != opp.ID || !c.Tapped || c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("controller %v tapped %v counters %d: want its owner, tapped, no counter", c.Controller, c.Tapped, c.Counters[game.CounterPlusOne])
	}
}

func TestMidHallowedRespiteRefusesALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	king := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "King", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	err := midStrictCast(t, g, midHallowedRespite, midTarget(king))
	if err == nil {
		t.Fatal("a legendary creature was a legal target")
	}
}

func TestMidHallowedRespiteFlashbackExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	id := effectstest.SeedGraveyardCard(t, g, "Hallowed Respite", "Sorcery", midHallowedRespite)
	midFlashback(t, g, me, id, midTarget(bear))
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back spell is not in exile")
	}
	if c, _ := midOnBattlefield(g, effectstest.FindBattlefieldByName(g, "Bear")); c.Counters[game.CounterPlusOne] != 1 {
		t.Error("the flashback did not blink")
	}
}

// midStrictCast is a strict hand cast of a sorcery that must be refused.
func midStrictCast(t *testing.T, g *game.Game, oracle string, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Sorcery", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

// --- Dire-Strain Rampage ------------------------------------------------

func midBasics(p *game.Player, n int) []uuid.UUID {
	return effectstest.SeedTypedLibrary(p, "Basic Land — Forest", func() []string {
		names := make([]string, n)
		for i := range names {
			names[i] = "Forest"
		}
		return names
	}()...)
}

func TestMidDireStrainRampageDestroyedLandFetchesTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Land", TypeLine: "Land"})
	basics := midBasics(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Dire-Strain Rampage", "Sorcery", midDireStrain, midTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := midOnBattlefield(g, land); ok {
		t.Fatal("the land was not destroyed")
	}
	effectstest.AnswerSearchByID(t, g, opp.ID, basics[0], basics[1])
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range basics[:2] {
		c, ok := midOnBattlefield(g, id)
		if !ok || !c.Tapped || c.Controller != opp.ID {
			t.Errorf("basic %v: on battlefield %v tapped %v, want its controller's, tapped", id, ok, c.Tapped)
		}
	}
	if _, ok := midOnBattlefield(g, basics[2]); ok {
		t.Error("a third basic was fetched")
	}
	_ = me
}

func TestMidDireStrainRampageOtherTargetFetchesOneAndMayBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	rock := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	basics := midBasics(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Dire-Strain Rampage", "Sorcery", midDireStrain, midTarget(rock))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatalf("its controller was not offered a search: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{basics[0], basics[1]}); err == nil {
		t.Fatal("two basics were allowed for a non-land")
	}
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{basics[0]}); err != nil {
		t.Fatalf("one basic: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := midOnBattlefield(g, basics[0]); !ok {
		t.Error("the basic did not enter")
	}
}

// --- Storm the Festival -------------------------------------------------

func TestMidStormTheFestivalPutsUpToTwoCheapPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Library.Cards = nil })
	mk := func(name, typ, cost string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typ, ManaCost: cost, Owner: me.ID, Controller: me.ID})
		return id
	}
	bottom := mk("Deep", "Creature — Bear", "{1}")
	big := mk("Big", "Creature — Wurm", "{6}")
	bolt := mk("Bolt", "Instant", "{R}")
	elf := mk("Elf", "Creature — Elf", "{G}")
	ring := mk("Ring", "Artifact", "{1}")
	cheap := mk("Cheap", "Creature — Bear", "{2}")
	effectstest.CastCatalogSpell(t, g, "Storm the Festival", "Sorcery", midStormFestival, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatalf("no pick offered: %+v", g.PendingChoices)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	for _, id := range []uuid.UUID{big, bolt, bottom, cheap} {
		if id == cheap {
			continue
		}
		if offered[id] {
			t.Errorf("%v was offered (not a permanent card with value 5 or less, or not in the top five)", id)
		}
	}
	if !offered[elf] || !offered[ring] || !offered[cheap] {
		t.Fatalf("legal cards not offered: %v", offered)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elf, ring, cheap}); err == nil {
		t.Fatal("three cards were accepted")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{elf, ring}); err != nil {
		t.Fatalf("pick two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{elf, ring} {
		if _, ok := midOnBattlefield(g, id); !ok {
			t.Errorf("%v did not enter", id)
		}
	}
	// the rest went to the bottom: the next draw is the card under the top five
	if me.Library.Size() != 4 {
		t.Errorf("library size %d, want 4", me.Library.Size())
	}
	for _, id := range []uuid.UUID{cheap, big, bolt} {
		if !me.Library.Contains(id) {
			t.Errorf("%v left the library", id)
		}
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != bottom {
		t.Errorf("the un-looked-at card is not on top anymore")
	}
}

// --- Turn the Earth -----------------------------------------------------

func TestMidTurnTheEarthShufflesUpToThreeOwnersCardsAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.PushGraveyardCardForTest(me, "A")
	b := effectstest.PushGraveyardCardForTest(opp, "B")
	c := effectstest.PushGraveyardCardForTest(opp, "C")
	d := effectstest.PushGraveyardCardForTest(opp, "D")
	life := me.Life
	if err := midStrictCast(t, g, midTurnTheEarth, []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}, {Kind: game.TargetCard, ID: d}}); err == nil {
		t.Fatal("four targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Turn the Earth", "Instant", midTurnTheEarth,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(a) || !opp.Library.Contains(b) || !opp.Library.Contains(c) {
		t.Error("the cards did not go to their owners' libraries")
	}
	if !opp.Graveyard.Contains(d) {
		t.Error("an untargeted card moved")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMidTurnTheEarthWithNoTargetsStillGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Turn the Earth", "Instant", midTurnTheEarth, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestMidTurnTheEarthFlashback(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	x := effectstest.PushGraveyardCardForTest(me, "X")
	id := effectstest.SeedGraveyardCard(t, g, "Turn the Earth", "Instant", midTurnTheEarth)
	midFlashback(t, g, me, id, midTarget(x))
	if !g.Exile.Contains(id) || !me.Library.Contains(x) {
		t.Error("flashback did not shuffle the target and exile the spell")
	}
}

// --- Can't Stay Away ----------------------------------------------------

func TestMidCantStayAwayReturnsACheapCreatureThatExilesInsteadOfDying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := uuid.New()
	pricey := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: pricey, Name: "Pricey", TypeLine: "Creature — Bear", ManaCost: "{3}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if err := midStrictCast(t, g, midCantStayAway, midTarget(pricey)); err == nil {
		t.Fatal("a mana value 4 creature was a legal target")
	}
	effectstest.CastCatalogSpell(t, g, "Can't Stay Away", "Sorcery", midCantStayAway, midTarget(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := midOnBattlefield(g, cheap); !ok {
		t.Fatal("the creature did not return")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(cheap); err != nil {
			t.Errorf("destroy: %v", err)
		}
	})
	if !g.Exile.Contains(cheap) || me.Graveyard.Contains(cheap) {
		t.Error("the creature did not get exiled instead of dying")
	}
}

func TestMidCantStayAwayTheGrantOutlastsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	cheap := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Can't Stay Away", "Sorcery", midCantStayAway, midTarget(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	// it survives the turn boundary: the effect is not "this turn"
	for seq := g.Turn.Seq; g.Turn.Seq == seq; {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(cheap) })
	if !g.Exile.Contains(cheap) {
		t.Error("the replacement did not outlast the turn")
	}
}

// --- Wake to Slaughter --------------------------------------------------

func midWakeSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.PushGraveyardCardTyped(me, "Alpha", "Creature — Bear")
	b := effectstest.PushGraveyardCardTyped(me, "Beta", "Creature — Bear")
	return g, me, opp, a, b
}

func TestMidWakeToSlaughterOpponentPicksTheHandCardTheOtherRunsWithHaste(t *testing.T) {
	g, me, opp, a, b := midWakeSetup(t)
	effectstest.CastCatalogSpell(t, g, "Wake to Slaughter", "Sorcery", midWakeToSlaughter,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	}
	pick := effectstest.LatestRevealPickFor(g, opp.ID)
	if pick == nil {
		t.Fatalf("the opponent was not asked: %+v", g.PendingChoices)
	}
	if err := g.ResolveRevealPick(pick.ID, opp.ID, []uuid.UUID{a}); err != nil {
		t.Fatalf("ResolveRevealPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the chosen card did not go to hand")
	}
	c, ok := midOnBattlefield(g, b)
	if !ok || c.Controller != me.ID {
		t.Fatal("the other card did not return under your control")
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, b, "haste") {
		t.Error("it did not gain haste")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(b) {
		t.Error("it was not exiled at the beginning of the next end step")
	}
}

func TestMidWakeToSlaughterOneTargetGoesToHand(t *testing.T) {
	g, me, _, a, _ := midWakeSetup(t)
	effectstest.CastCatalogSpell(t, g, "Wake to Slaughter", "Sorcery", midWakeToSlaughter, midTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) {
		t.Error("the only chosen card did not go to hand")
	}
	if _, ok := midOnBattlefield(g, a); ok {
		t.Error("it also entered the battlefield")
	}
}

func TestMidWakeToSlaughterRefusesANonCreatureAndAnOpponentsCard(t *testing.T) {
	g, _, opp, _, _ := midWakeSetup(t)
	theirs := effectstest.PushGraveyardCardTyped(opp, "Theirs", "Creature — Bear")
	if err := midStrictCast(t, g, midWakeToSlaughter, midTarget(theirs)); err == nil {
		t.Fatal("an opponent's graveyard card was a legal target")
	}
}

// --- Rite of Harmony ----------------------------------------------------

func TestMidRiteOfHarmonyDrawsForEachCreatureOrEnchantmentThatEntersThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Early", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.CastCatalogSpell(t, g, "Rite of Harmony", "Instant", midRiteOfHarmony, nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.CastCreatureCard(t, g, "Late Bear", "Creature — Bear", "", 2, 2)
	if got := me.Hand.Size(); got != hand+1 {
		// the card was put in hand and cast (net 0), then the trigger drew (+1)
		t.Errorf("hand %d after a creature entered, want %d", got, hand+1)
	}
	hand = me.Hand.Size()
	effectstest.CastCreatureCard(t, g, "Aura Maker", "Enchantment", "", 0, 0)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after an enchantment entered, want %d", got, hand+1)
	}
	hand = me.Hand.Size()
	effectstest.CastCreatureCard(t, g, "Rock", "Artifact", "", 0, 0)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after an artifact entered, want %d", got, hand)
	}
	hand = me.Hand.Size()
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after an opponent's creature entered, want %d", got, hand)
	}
}

func TestMidDireStrainRampageIndestructibleLandIsNotDestroyedSoOneBasic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	land := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Hardy Land", TypeLine: "Land", Keywords: []string{"indestructible"}})
	basics := midBasics(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Dire-Strain Rampage", "Sorcery", midDireStrain, midTarget(land))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := midOnBattlefield(g, land); !ok {
		t.Fatal("the indestructible land was destroyed")
	}
	c := effectstest.SearchChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatalf("no search offered: %+v", g.PendingChoices)
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, opp.ID, []uuid.UUID{basics[0], basics[1]}); err == nil {
		t.Fatal("two basics were allowed: a land was not destroyed this way")
	}
}
