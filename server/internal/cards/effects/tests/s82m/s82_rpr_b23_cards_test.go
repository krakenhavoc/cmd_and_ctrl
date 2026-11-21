package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB23ShadeOracle      = "2345b243-ed0d-4556-8017-d533db1fbc74"
	rprB23ReignOracle      = "3e1a6337-bb29-450f-bf29-ff71b357a983"
	rprB23PlundererOracle  = "c5e47644-c72c-40f9-b65f-57fcea075466"
	rprB23PhoenixOracle    = "f34be4b8-126d-4b5c-8ebb-23b5e5370012"
	rprB23VengeanceOracle  = "5f4ff27f-ebc1-4a86-8b0b-eeea470a25fb"
	rprB23SeverOracle      = "b9019332-2f02-48d9-be0a-a2b41b79e136"
	rprB23SabertoothOracle = "8bf8e846-9d10-4dbc-a379-df37bb9fb6d4"
	rprB23OfferingOracle   = "732aaa4b-22ea-49e2-b09f-e6169f726537"
	rprB23RoaleskOracle    = "ca4936cf-48d4-4469-9243-8dbf8ec4cdca"
	rprB23BogbonderOracle  = "2f946c13-cfab-415f-9df7-6d621761b2fc"
)

func rprB23Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func rprB23Pay(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func rprB23Registered(t *testing.T) {
	t.Helper()
	for name, oracle := range map[string]string{
		"Skyclave Shade": rprB23ShadeOracle, "Reign of the Pit": rprB23ReignOracle,
		"Skyship Plunderer": rprB23PlundererOracle, "Rekindling Phoenix": rprB23PhoenixOracle,
		"Stalking Vengeance": rprB23VengeanceOracle, "Sever the Bloodline": rprB23SeverOracle,
		"Temur Sabertooth": rprB23SabertoothOracle, "Sylvan Offering": rprB23OfferingOracle,
		"Roalesk, Apex Hybrid": rprB23RoaleskOracle, "Slippery Bogbonder": rprB23BogbonderOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v registered=%v", name, spec, ok)
		}
	}
}

func TestRprB23AllCardsAreRegisteredFull(t *testing.T) { rprB23Registered(t) }

// --- Skyclave Shade -----------------------------------------------------

func TestRprB23SkyclaveShadeKickedEntersWithTwoCounters(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{{"unkicked", nil, 0}, {"kicked", []int{0}, 2}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id, err := effectstest.CastWithOptionalCosts(t, g, "Skyclave Shade", "Creature — Shade", rprB23ShadeOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if n := effectstest.PlusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.want)
			}
		})
	}
}

func TestRprB23SkyclaveShadeCantBlock(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Skyclave Shade", "Creature — Shade", rprB23ShadeOracle, false)
	if effectstest.AuraRestrictions(t, g, id)&game.CantBlock == 0 {
		t.Fatal("the Shade can block")
	}
}

func rprB23ShadeInGraveyard(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Skyclave Shade", TypeLine: "Creature — Shade",
			OracleID: rprB23ShadeOracle, ManaCost: "{1}{B}", Power: 3, Toughness: 1, Owner: p.ID, Controller: p.ID})
	})
	return id
}

func rprB23Land(t *testing.T, g *game.Game, owner uuid.UUID) {
	t.Helper()
	effectstest.EnterCard(t, g, owner, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
}

func TestRprB23SkyclaveShadeLandfallLetsYouCastItFromYourGraveyard(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := rprB23ShadeInGraveyard(g, me)
	rprB23Pay(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("cast from the graveyard with no landfall")
	}
	rprB23Land(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("cast from the graveyard after landfall: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("the Shade did not enter")
	}
}

func TestRprB23SkyclaveShadeCanBeKickedFromTheGraveyard(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := rprB23ShadeInGraveyard(g, me)
	rprB23Land(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	rprB23Pay(t, g, me, "{B}{B}{B}{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("kicked cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, id); n != 2 {
		t.Fatalf("+1/+1 counters %d, want 2", n)
	}
}

func TestRprB23SkyclaveShadeIgnoresAnOpponentsLand(t *testing.T) {
	g, me, opp := rprB23Table(t)
	id := rprB23ShadeInGraveyard(g, me)
	rprB23Land(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	rprB23Pay(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("an opponent's land gave a cast from the graveyard")
	}
}

func TestRprB23SkyclaveShadeOnlyOnYourTurn(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := rprB23ShadeInGraveyard(g, me)
	effectstest.AdvanceToMainOf(t, g, 1)
	rprB23Land(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	rprB23Pay(t, g, me, "{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("landfall on another player's turn gave a cast from the graveyard")
	}
}

// --- Sever the Bloodline ------------------------------------------------

func TestRprB23SeverExilesTheCreatureAndEveryOtherWithItsName(t *testing.T) {
	g, me, opp := rprB23Table(t)
	a := effectstest.B12Creature(g, opp.ID, "Grizzly", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Grizzly", "Creature — Bear", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Wolf", "Creature — Wolf", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sever the Bloodline", "Sorcery", rprB23SeverOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if !g.Exile.Contains(id) {
			t.Errorf("Grizzly %s was not exiled", id)
		}
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a differently named creature was exiled")
	}
}

func TestRprB23SeverLeavesANonCreatureWithTheSameName(t *testing.T) {
	g, me, opp := rprB23Table(t)
	a := effectstest.B12Creature(g, opp.ID, "Grizzly", "Creature — Bear", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly", "Artifact", "", false)
	effectstest.CastCatalogSpell(t, g, "Sever the Bloodline", "Sorcery", rprB23SeverOracle, effectstest.B16TargetCard(a))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Battlefield.Contains(rock) {
		t.Fatal("only creatures with the name go")
	}
}

func TestRprB23SeverRefusesANonCreatureTargetAndHasFlashback(t *testing.T) {
	g, me, _ := rprB23Table(t)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Sever the Bloodline", "Sorcery", rprB23SeverOracle, effectstest.B16TargetCard(rock)); err == nil {
		t.Fatal("targeted an artifact")
	}
	g2, _, opp2 := rprB23Table(t)
	foe := effectstest.B12Creature(g2, opp2.ID, "Grizzly", "Creature — Bear", 2, 2)
	id := effectstest.MidB03Flash(t, g2, "Sever the Bloodline", "Sorcery", rprB23SeverOracle, effectstest.B16TargetCard(foe))
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Exile.Contains(foe) {
		t.Error("the flashback did not exile the creature")
	}
	effectstest.MidB03Exiled(t, g2, id, "Sever the Bloodline")
}

// --- Temur Sabertooth ---------------------------------------------------

func TestRprB23SabertoothBouncesAnotherAndGainsIndestructible(t *testing.T) {
	for _, decline := range []bool{false, true} {
		g, me, _ := rprB23Table(t)
		saber := effectstest.PushCatalogPermanent(g, me.ID, "Temur Sabertooth", "Creature — Cat", rprB23SabertoothOracle, false)
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
		rprB23Pay(t, g, me, "{C}{G}")
		if err := g.ActivateCatalogAbility(me.ID, saber, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
		if pick == nil {
			t.Fatal("no creature choice")
		}
		if effectstest.HasID(pick.ChooseCards, saber) || effectstest.HasID(pick.ChooseCards, foe) || !effectstest.HasID(pick.ChooseCards, other) {
			t.Fatalf("offered %v", pick.ChooseCards)
		}
		picks := []uuid.UUID{other}
		if decline {
			picks = nil
		}
		if err := g.ResolveOwnPermanents(pick.ID, me.ID, picks); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Hand.Contains(other); got == decline {
			t.Errorf("decline=%v: the creature is in hand = %v", decline, got)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, saber, "indestructible"); got == decline {
			t.Errorf("decline=%v: indestructible = %v", decline, got)
		}
	}
}

func TestRprB23SabertoothNeedsItsMana(t *testing.T) {
	g, me, _ := rprB23Table(t)
	saber := effectstest.PushCatalogPermanent(g, me.ID, "Temur Sabertooth", "Creature — Cat", rprB23SabertoothOracle, false)
	rprB23Pay(t, g, me, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, saber, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {G}")
	}
}

// --- Stalking Vengeance -------------------------------------------------

func TestRprB23StalkingVengeanceDealsTheDeadCreaturesPowerToAPlayer(t *testing.T) {
	g, me, opp := rprB23Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Stalking Vengeance", "Creature — Avatar", rprB23VengeanceOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	before := opp.Life
	mb1Kill(g, bear)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Fatalf("damage %d, want 3 (2 power plus a counter)", got)
	}
}

func TestRprB23StalkingVengeanceIgnoresOtherCreaturesAndItself(t *testing.T) {
	g, me, opp := rprB23Table(t)
	vengeance := effectstest.PushCatalogPermanent(g, me.ID, "Stalking Vengeance", "Creature — Avatar", rprB23VengeanceOracle, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mb1Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("an opponent's creature dying triggered it")
	}
	mb1Kill(g, vengeance)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("the Vengeance's own death triggered it")
	}
}

func TestRprB23StalkingVengeanceHasHaste(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Stalking Vengeance", "Creature — Avatar", rprB23VengeanceOracle, true)
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("no haste")
	}
}

// --- Rekindling Phoenix -------------------------------------------------

func TestRprB23RekindlingPhoenixLeavesAnElementalThatBringsItBack(t *testing.T) {
	g, me, _ := rprB23Table(t)
	phoenix := effectstest.PushCatalogPermanent(g, me.ID, "Rekindling Phoenix", "Creature — Phoenix", rprB23PhoenixOracle, false)
	mb1Kill(g, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	var token game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no Elemental token")
	}
	if p, tough := mb1Power(t, g, token.InstanceID); p != 0 || tough != 1 || !token.HasColor("R") {
		t.Fatalf("token is %d/%d, colours %v; want a 0/1 red Elemental", p, tough, token.Colors)
	}
	if !me.Graveyard.Contains(phoenix) {
		t.Fatal("the Phoenix is not in the graveyard")
	}
	effectstest.AdvanceToStepOf(t, g, rprB23SeatOf(g, me), game.StepUpkeep)
	for i := 0; i < 3 && g.Battlefield.Contains(token.InstanceID); i++ {
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: phoenix}); err != nil {
				t.Fatal(err)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Fatal("the Elemental was not sacrificed")
	}
	if !g.Battlefield.Contains(phoenix) {
		t.Fatal("the Phoenix did not return")
	}
	if !effectstest.HasEffectiveKeyword(t, g, phoenix, "haste") {
		t.Fatal("the returned Phoenix has no haste")
	}
}

func TestRprB23RekindlingElementalNeedsAPhoenixInYourGraveyard(t *testing.T) {
	g, me, _ := rprB23Table(t)
	phoenix := effectstest.PushCatalogPermanent(g, me.ID, "Rekindling Phoenix", "Creature — Phoenix", rprB23PhoenixOracle, false)
	mb1Kill(g, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	// Exile the Phoenix before the upkeep: the trigger has no target, so it
	// is removed and the Elemental stays.
	g.WithWriteLock(func() { _, _ = game.MoveCard(me.Graveyard, g.Exile, phoenix) })
	effectstest.AdvanceToStepOf(t, g, rprB23SeatOf(g, me), game.StepUpkeep)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" && c.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("%d Elementals, want the one to stay", n)
	}
}

func rprB23SeatOf(g *game.Game, p *game.Player) int {
	for i, s := range g.Seats {
		if s.ID == p.ID {
			return i
		}
	}
	return -1
}

// --- Skyship Plunderer --------------------------------------------------

func rprB23CombatHit(g *game.Game, me *game.Player, source, target uuid.UUID, combat bool) {
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: source, Target: target, Amount: 2, Combat: combat})
	})
}

func TestRprB23SkyshipPlundererGivesAnotherCounterOfEachKind(t *testing.T) {
	g, me, opp := rprB23Table(t)
	plunderer := effectstest.PushCatalogPermanent(g, me.ID, "Skyship Plunderer", "Creature — Human Pirate", rprB23PlundererOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, plunderer, "flying") {
		t.Fatal("no flying")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Counter Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(bear, "charge", 1)
	})
	rprB23CombatHit(g, me, plunderer, opp.ID, true)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, bear); got != 3 {
		t.Errorf("+1/+1 counters %d, want 3", got)
	}
	if got := effectstest.CounterOn(g, bear, "charge"); got != 2 {
		t.Errorf("charge counters %d, want 2", got)
	}
}

func TestRprB23SkyshipPlundererCanTargetAPlayerAndOnlyTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g, me, opp := rprB23Table(t)
	plunderer := effectstest.PushCatalogPermanent(g, me.ID, "Skyship Plunderer", "Creature — Human Pirate", rprB23PlundererOracle, false)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterForEffect(opp.ID, "poison", 2) })
	rprB23CombatHit(g, me, plunderer, opp.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("noncombat damage triggered it")
	}
	rprB23CombatHit(g, me, plunderer, opp.ID, true)
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters["poison"]; got != 3 {
		t.Fatalf("poison counters %d, want 3", got)
	}
}

// --- Roalesk, Apex Hybrid -----------------------------------------------

func TestRprB23RoaleskPutsTwoCountersOnAnotherCreatureYouControl(t *testing.T) {
	g, me, opp := rprB23Table(t)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	roalesk := mb1Enter(t, g, me, "Roalesk, Apex Hybrid", "Legendary Creature — Human Mutant", rprB23RoaleskOracle, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if effectstest.HasID(p.PickTargetCards, roalesk) || effectstest.HasID(p.PickTargetCards, foe) {
			t.Fatalf("offered %v", p.PickTargetCards)
		}
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: other}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.PlusOneCounters(g, other); got != 2 {
		t.Errorf("+1/+1 counters on the other creature %d, want 2", got)
	}
	if effectstest.PlusOneCounters(g, roalesk) != 0 || effectstest.PlusOneCounters(g, foe) != 0 {
		t.Error("counters landed on Roalesk or an opponent's creature")
	}
	for _, kw := range []string{"flying", "trample"} {
		if !effectstest.HasEffectiveKeyword(t, g, roalesk, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

func TestRprB23RoaleskAloneHasNoTargetForItsCounters(t *testing.T) {
	g, me, _ := rprB23Table(t)
	roalesk := mb1Enter(t, g, me, "Roalesk, Apex Hybrid", "Legendary Creature — Human Mutant", rprB23RoaleskOracle, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PlusOneCounters(g, roalesk) != 0 {
		t.Fatal("Roalesk put counters on itself")
	}
}

func TestRprB23RoaleskProliferatesTwiceWhenItDies(t *testing.T) {
	g, me, _ := rprB23Table(t)
	roalesk := effectstest.PushCatalogPermanent(g, me.ID, "Roalesk, Apex Hybrid", "Legendary Creature — Human Mutant", rprB23RoaleskOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Counter Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	mb1Kill(g, roalesk)
	asked := 0
	for i := 0; i < 6; i++ {
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.ProliferatePrompt(g)
		if c == nil {
			break
		}
		asked++
		effectstest.AnswerProliferate(t, g, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if asked != 2 {
		t.Errorf("proliferate asked %d times, want 2", asked)
	}
	if got := effectstest.PlusOneCounters(g, bear); got != 3 {
		t.Errorf("+1/+1 counters %d, want 3 (1 plus two proliferates)", got)
	}
}

// --- Slippery Bogbonder -------------------------------------------------

func rprB23AnswerNumber(t *testing.T, g *game.Game, who uuid.UUID, n int) {
	t.Helper()
	c := effectstest.PayAmountPrompt(g, who)
	if c == nil {
		t.Fatal("no number prompt")
	}
	if err := g.ResolvePayAmount(c.ID, who, n); err != nil {
		t.Fatal(err)
	}
}

func rprB23EnterBogbonder(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := mb1Enter(t, g, me, "Slippery Bogbonder", "Creature — Human Druid", rprB23BogbonderOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id
}

func TestRprB23BogbonderPutsAHexproofCounterAndMovesChosenCounters(t *testing.T) {
	g, me, _ := rprB23Table(t)
	dest := effectstest.B12Creature(g, me.ID, "Destination", "Creature — Bear", 2, 2)
	srcA := effectstest.B12Creature(g, me.ID, "Source A", "Creature — Bear", 2, 2)
	srcB := effectstest.B12Creature(g, me.ID, "Source B", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, srcA, game.CounterPlusOne, 3)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(srcB, "charge", 2) })
	rprB23EnterBogbonder(t, g, me, dest)
	// Source A: move two of its three +1/+1 counters.
	rprB23AnswerNumber(t, g, me.ID, 2)
	// Source B: move one of its two charge counters.
	rprB23AnswerNumber(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, dest, game.CounterHexproof); got != 1 {
		t.Errorf("hexproof counters %d, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, dest); got != 2 {
		t.Errorf("+1/+1 counters on the destination %d, want 2", got)
	}
	if got := effectstest.CounterOn(g, dest, "charge"); got != 1 {
		t.Errorf("charge counters on the destination %d, want 1", got)
	}
	if got := effectstest.PlusOneCounters(g, srcA); got != 1 {
		t.Errorf("Source A kept %d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, srcB, "charge"); got != 1 {
		t.Errorf("Source B kept %d charge counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dest, "hexproof") {
		t.Error("the hexproof counter gives no hexproof")
	}
}

func TestRprB23BogbonderMayMoveNothingAndCanHelpAnOpponentsCreature(t *testing.T) {
	g, me, opp := rprB23Table(t)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, me.ID, "Source", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, src, game.CounterPlusOne, 2)
	rprB23EnterBogbonder(t, g, me, foe)
	rprB23AnswerNumber(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, foe, game.CounterHexproof); got != 1 {
		t.Errorf("hexproof counters on the opponent's creature %d, want 1", got)
	}
	if effectstest.PlusOneCounters(g, foe) != 0 || effectstest.PlusOneCounters(g, src) != 2 {
		t.Error("counters moved though the answer was none")
	}
}

func TestRprB23BogbonderDoesNotMoveFromOpponentsCreatures(t *testing.T) {
	g, me, opp := rprB23Table(t)
	dest := effectstest.B12Creature(g, me.ID, "Destination", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, foe, game.CounterPlusOne, 2)
	rprB23EnterBogbonder(t, g, me, dest)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Fatal("asked about an opponent's creature")
	}
	if effectstest.PlusOneCounters(g, foe) != 2 || effectstest.PlusOneCounters(g, dest) != 0 {
		t.Fatal("counters moved off an opponent's creature")
	}
}

func TestRprB23BogbonderHasFlashAndHexproof(t *testing.T) {
	g, me, _ := rprB23Table(t)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Slippery Bogbonder", "Creature — Human Druid", rprB23BogbonderOracle, false)
	for _, kw := range []string{"flash", "hexproof"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

// --- Reign of the Pit ---------------------------------------------------

func rprB23Demons(g *game.Game, controller uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Demon" && c.Controller == controller {
			out = append(out, c)
		}
	}
	return out
}

func TestRprB23ReignOfThePitMakesADemonOfTheTotalPowerSacrificed(t *testing.T) {
	g, me, opp := rprB23Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 3, 3)
	effectstest.PutCounters(t, g, mine, game.CounterPlusOne, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Reign of the Pit", "Sorcery", rprB23ReignOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Fatal("a creature was not sacrificed")
	}
	demons := rprB23Demons(g, me.ID)
	if len(demons) != 1 {
		t.Fatalf("%d Demons, want 1", len(demons))
	}
	p, tough := mb1Power(t, g, demons[0].InstanceID)
	if p != 9 || tough != 9 {
		t.Fatalf("Demon is %d/%d, want 9/9 (4 + 5)", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, demons[0].InstanceID, "flying") {
		t.Error("the Demon has no flying")
	}
	if !demons[0].HasColor("B") {
		t.Error("the Demon is not black")
	}
}

func TestRprB23ReignOfThePitLetsEachPlayerChoose(t *testing.T) {
	g, me, opp := rprB23Table(t)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 6, 6)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirBig := effectstest.B12Creature(g, opp.ID, "Their Big", "Creature — Ogre", 7, 7)
	effectstest.CastCatalogSpell(t, g, "Reign of the Pit", "Sorcery", rprB23ReignOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, small)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(theirBear) {
		t.Fatal("a chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(theirBig) {
		t.Fatal("a creature that was not chosen was sacrificed")
	}
	demons := rprB23Demons(g, me.ID)
	if len(demons) != 1 {
		t.Fatalf("%d Demons, want 1", len(demons))
	}
	if p, _ := mb1Power(t, g, demons[0].InstanceID); p != 3 {
		t.Fatalf("Demon power %d, want 3 (1 + 2)", p)
	}
}

func TestRprB23ReignOfThePitWithNoCreaturesMakesNothingThatLives(t *testing.T) {
	g, me, _ := rprB23Table(t)
	effectstest.CastCatalogSpell(t, g, "Reign of the Pit", "Sorcery", rprB23ReignOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if n := len(rprB23Demons(g, me.ID)); n != 0 {
		t.Fatalf("%d 0/0 Demons survived", n)
	}
}

// --- Sylvan Offering ----------------------------------------------------

func TestRprB23SylvanOfferingMakesTreefolkAndElvesForYouAndTheChosenOpponents(t *testing.T) {
	g, me, opp := rprB23Table(t)
	third := g.Seats[2]
	effectstest.CastXSpell(t, g, "Sylvan Offering", "Sorcery", rprB23OfferingOracle, "{X}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, third)
	effectstest.PassPriorityAroundTable(t, g)
	count := func(who uuid.UUID, name string) (n, power int) {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == who && c.IsToken() {
				n++
				power = effectstest.EffectivePower(t, g, c.InstanceID)
			}
		}
		return
	}
	for _, who := range []uuid.UUID{me.ID, opp.ID} {
		if n, p := count(who, "Treefolk"); n != 1 || p != 3 {
			t.Errorf("%s has %d Treefolk of power %d, want one 3/3", who, n, p)
		}
	}
	if n, _ := count(third.ID, "Treefolk"); n != 0 {
		t.Errorf("the second-chosen opponent got %d Treefolk", n)
	}
	if n, _ := count(me.ID, "Elf Warrior"); n != 3 {
		t.Errorf("you have %d Elf Warriors, want 3", n)
	}
	if n, _ := count(third.ID, "Elf Warrior"); n != 3 {
		t.Errorf("the second chosen opponent has %d Elf Warriors, want 3", n)
	}
	if n, _ := count(opp.ID, "Elf Warrior"); n != 0 {
		t.Errorf("the first chosen opponent got %d Elf Warriors", n)
	}
}

func TestRprB23SylvanOfferingWithXZeroMakesNothingThatLives(t *testing.T) {
	g, me, opp := rprB23Table(t)
	effectstest.CastXSpell(t, g, "Sylvan Offering", "Sorcery", rprB23OfferingOracle, "{X}{G}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() {
			t.Fatalf("a %s token survived X = 0", c.Name)
		}
	}
}
