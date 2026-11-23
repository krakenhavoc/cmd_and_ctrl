package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_neo_b20_cards_test.go — S82 NEO batch 20: planeswalkers and complex
// legends. Every test drives the engine (cast, trigger, stack, prompt).

const (
	neoKuraOracle  = "6426e73a-7e11-48ab-ac5c-15647c6cf9ba"
	neoKairiOracle = "85aa53e8-f062-4d02-bc69-1413f7f0dcf1"
)

func neoB20Push(g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func neoB20Lands(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		neoB20Push(g, owner, "Forest", "Basic Land — Forest", "", "", 0, 0)
	}
}

func neoB20Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func neoB20Kill(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
}

func neoB20Mode(t *testing.T, g *game.Game, seat uuid.UUID, mode int) {
	t.Helper()
	c := effectstest.ModePickChoiceFor(g, seat)
	if c == nil {
		t.Fatal("the dies trigger asks which mode")
	}
	if len(c.ModeOptionIndex) != 2 || c.ModeMin != 1 || c.ModeMax != 1 {
		t.Fatalf("two bullets, choose one: %+v", c)
	}
	if err := g.ResolveModePick(c.ID, seat, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func neoB20Registered(t *testing.T, oracle, name string) effects.Spec {
	t.Helper()
	spec, ok := effects.Lookup(oracle)
	if !ok {
		t.Fatalf("%s is not registered", name)
	}
	if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%s registered as %q / %v / caveats %v, want full with none", name, spec.Name, spec.Completeness, spec.Caveats)
	}
	return spec
}

func TestNeoB20KuraIsRegisteredFull(t *testing.T) {
	neoB20Registered(t, neoKuraOracle, "Kura, the Boundless Sky")
}

func TestNeoB20KuraDiesSearchesUpToThreeLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	p1 := effectstest.PlTop(me, "Island", "Basic Land — Island", "")
	p2 := effectstest.PlTop(me, "Mountain", "Basic Land — Mountain", "")
	p3 := effectstest.PlTop(me, "Plains", "Basic Land — Plains", "")
	p4 := effectstest.PlTop(me, "Swamp", "Basic Land — Swamp", "")
	bear := effectstest.PlTop(me, "Grizzly Bears", "Creature — Bear", "{1}{G}")
	kura := neoB20Push(g, me.ID, "Kura, the Boundless Sky", "Legendary Creature — Dragon Spirit", "{3}{G}{G}", neoKuraOracle, 4, 4)
	neoB20Kill(t, g, kura)
	neoB20Mode(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the first bullet searches the library")
	}
	for _, id := range c.SearchCards {
		if id == bear {
			t.Error("a nonland card was offered")
		}
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{p1, p2, p3, p4}); err == nil {
		t.Fatal("four lands is over the limit of three")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{p1, p2, p3}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	for _, id := range []uuid.UUID{p1, p2, p3} {
		if !effectstest.HandHasID(me, id) {
			t.Errorf("%s did not reach the hand", id)
		}
	}
	if effectstest.HandHasID(me, p4) {
		t.Error("the fourth land was taken")
	}
}

func TestNeoB20KuraDiesMakesASpiritAsBigAsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	neoB20Lands(g, me.ID, 4)
	neoB20Lands(g, opp.ID, 7) // an opponent's lands are not counted
	kura := neoB20Push(g, me.ID, "Kura, the Boundless Sky", "Legendary Creature — Dragon Spirit", "{3}{G}{G}", neoKuraOracle, 4, 4)
	neoB20Kill(t, g, kura)
	neoB20Mode(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	var spirit uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && effects.IsToken(c) && c.Controller == me.ID {
			spirit = c.InstanceID
		}
	}
	if spirit == uuid.Nil {
		t.Fatal("no Spirit token")
	}
	if p := effectstest.EffectivePower(t, g, spirit); p != 4 {
		t.Errorf("Spirit power %d, want 4 (the lands you control)", p)
	}
}

func TestNeoB20KairiIsRegisteredFull(t *testing.T) {
	neoB20Registered(t, neoKairiOracle, "Kairi, the Swirling Sky")
}

func TestNeoB20KairiDiesBouncesPermanentsUnderTheTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	two := neoB20Push(g, opp.ID, "Two Drop", "Creature — Bear", "{1}{G}", "", 2, 2)
	four := neoB20Push(g, opp.ID, "Four Drop", "Artifact", "{4}", "", 0, 0)
	three := neoB20Push(g, me.ID, "Three Drop", "Enchantment", "{2}{W}", "", 0, 0)
	land := neoB20Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	kairi := neoB20Push(g, me.ID, "Kairi, the Swirling Sky", "Legendary Creature — Dragon Spirit", "{4}{U}{U}", neoKairiOracle, 6, 6)
	neoB20Kill(t, g, kairi)
	neoB20Mode(t, g, me.ID, 0)
	// 4 + 3 = 7 is over the total; the land is not a nonland permanent.
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the first bullet asks for targets")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(four, three)); err == nil {
		t.Fatal("total mana value 7 must be refused")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(land)); err == nil {
		t.Fatal("a land must be refused")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, effectstest.CardRefs(two, four)); err != nil {
		t.Fatalf("2 + 4 = 6 is legal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(two) || g.Battlefield.Contains(four) {
		t.Error("the targets stayed on the battlefield")
	}
	if !effectstest.HandHasID(opp, two) || !effectstest.HandHasID(opp, four) {
		t.Error("the targets did not go to their owner's hand")
	}
	if !g.Battlefield.Contains(three) {
		t.Error("an untargeted permanent was bounced")
	}
}

func TestNeoB20KairiDiesMillsSixThenReturnsTwoSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Six cards on top of the library, two of them spells.
	bolt := effectstest.PlTop(me, "Shock", "Instant", "{R}")
	for i := 0; i < 3; i++ {
		effectstest.PlTop(me, "Filler", "Creature — Bear", "{2}")
	}
	sorc := effectstest.PlTop(me, "Divination", "Sorcery", "{2}{U}")
	effectstest.PlTop(me, "Filler", "Creature — Bear", "{2}")
	oldInstant := effectstest.PushGraveyardCardTyped(me, "Counterspell", "Instant")
	creature := effectstest.PushGraveyardCardTyped(me, "Old Bear", "Creature — Bear")
	kairi := neoB20Push(g, me.ID, "Kairi, the Swirling Sky", "Legendary Creature — Dragon Spirit", "{4}{U}{U}", neoKairiOracle, 6, 6)
	neoB20Kill(t, g, kairi)
	neoB20Mode(t, g, me.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the second bullet offers the graveyard after the mill")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if !offered[bolt] || !offered[sorc] || !offered[oldInstant] {
		t.Errorf("the milled and older spells must be offered: %v", offered)
	}
	if offered[creature] {
		t.Error("a creature card was offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{bolt, sorc, oldInstant}); err == nil {
		t.Fatal("three cards is more than up to two")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{sorc, oldInstant}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !effectstest.HandHasID(me, sorc) || !effectstest.HandHasID(me, oldInstant) {
		t.Error("the chosen spells did not reach the hand")
	}
}

const neoKyodaiOracle = "d89bd776-8489-4c34-94ea-db1b088d9379"

func TestNeoB20KyodaiIsRegisteredFull(t *testing.T) {
	spec := neoB20Registered(t, neoKyodaiOracle, "Kyodai, Soul of Kamigawa")
	if !effectstest.HasKeyword(spec.PrintedKeywords, "flash") || !effectstest.HasKeyword(spec.PrintedKeywords, "flying") {
		t.Errorf("printed keywords %v, want flash and flying", spec.PrintedKeywords)
	}
}

// The grant lasts while Kyodai's controller controls Kyodai; with Kyodai
// gone it ends (CR 611.2b), and Kyodai itself is never the target.
func TestNeoB20KyodaiGrantsIndestructibleWhileYouControlIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bear := neoB20Push(g, opp.ID, "Opposing Bear", "Creature — Bear", "{1}{G}", "", 2, 2)
	kyodai := effectstest.CastCatalogSpell(t, g, "Kyodai, Soul of Kamigawa", "Legendary Creature — Dragon Spirit", neoKyodaiOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asks for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: kyodai}); err == nil {
		t.Fatal("Kyodai may not target itself: another target permanent")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Fatal("the target did not gain indestructible")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, kyodai, "indestructible") {
		t.Error("Kyodai gained indestructible itself")
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(kyodai); err != nil {
			t.Fatalf("destroy Kyodai: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "indestructible") {
		t.Error("indestructible outlived Kyodai")
	}
}

func TestNeoB20KyodaiPumpNeedsAllFiveColours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	kyodai := neoB20Push(g, me.ID, "Kyodai, Soul of Kamigawa", "Legendary Creature — Dragon Spirit", "{3}{W}", neoKyodaiOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "U"}, game.ManaToken{Color: "B"}, game.ManaToken{Color: "R"})
	if err := g.ActivateCatalogAbility(me.ID, kyodai, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("four colours cannot pay {W}{U}{B}{R}{G}")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, kyodai, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePower(t, g, kyodai), effectstest.EffectiveToughness(t, g, kyodai); p != 8 || tough != 8 {
		t.Errorf("Kyodai is %d/%d, want 8/8", p, tough)
	}
}

const neoRisonaOracle = "44c4839f-c92a-4535-817b-1838c47fcb5e"

func TestNeoB20RisonaIsRegisteredFull(t *testing.T) {
	neoB20Registered(t, neoRisonaOracle, "Risona, Asari Commander")
}

func neoB20Counters(g *game.Game, id uuid.UUID, kind string) int {
	var n int
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.Counters[kind]
		}
	})
	return n
}

// Connecting puts one indestructible counter on Risona; a second hit finds
// the intervening "if" false and puts none more.
func TestNeoB20RisonaGainsAnIndestructibleCounterWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	risona := neoB20Push(g, me.ID, "Risona, Asari Commander", "Legendary Creature — Human Samurai", "{1}{R}{W}", neoRisonaOracle, 3, 3)
	effectstest.AttackWith(t, g, opp.ID, risona)
	effectstest.PassPriorityAroundTable(t, g)
	if n := neoB20Counters(g, risona, game.CounterIndestructible); n != 1 {
		t.Fatalf("%d indestructible counters after connecting, want 1", n)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, risona, "indestructible") {
		t.Error("the counter should make Risona indestructible")
	}
}

// Already holding the counter, Risona's hit adds nothing (CR 603.4).
func TestNeoB20RisonaWithACounterAlreadyGainsNoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	risona := neoB20Push(g, me.ID, "Risona, Asari Commander", "Legendary Creature — Human Samurai", "{1}{R}{W}", neoRisonaOracle, 3, 3)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(risona, game.CounterIndestructible, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	effectstest.AttackWith(t, g, opp.ID, risona)
	effectstest.PassPriorityAroundTable(t, g)
	if n := neoB20Counters(g, risona, game.CounterIndestructible); n != 1 {
		t.Fatalf("%d indestructible counters, want 1", n)
	}
}

// Combat damage dealt to Risona's controller removes one counter, once per
// damage event, and nothing when there is none.
func TestNeoB20RisonaLosesACounterWhenYouAreDealtCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	risona := neoB20Push(g, opp.ID, "Risona, Asari Commander", "Legendary Creature — Human Samurai", "{1}{R}{W}", neoRisonaOracle, 3, 3)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(risona, game.CounterIndestructible, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := neoB20Counters(g, risona, game.CounterIndestructible); n != 0 {
		t.Fatalf("%d counters left, want 0 (two attackers are one damage event)", n)
	}
}

func TestNeoB20RisonaCounterIsNotRemovedWhenSomeoneElseIsDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	risona := neoB20Push(g, opp.ID, "Risona, Asari Commander", "Legendary Creature — Human Samurai", "{1}{R}{W}", neoRisonaOracle, 3, 3)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(risona, game.CounterIndestructible, 1); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	effectstest.AttackWith(t, g, third.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if n := neoB20Counters(g, risona, game.CounterIndestructible); n != 1 {
		t.Fatalf("%d counters, want 1: damage to another player is not damage to you", n)
	}
}

const neoHidetsuguOracle = "d56f24a1-04dc-4769-9941-05bd99508e40"

func TestNeoB20HidetsuguIsRegisteredFull(t *testing.T) {
	neoB20Registered(t, neoHidetsuguOracle, "Hidetsugu, Devouring Chaos")
}

func TestNeoB20HidetsuguSacrificesACreatureToScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := neoB20Push(g, me.ID, "Hidetsugu, Devouring Chaos", "Legendary Creature — Ogre Demon", "{3}{B}", neoHidetsuguOracle, 4, 4)
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	effectstest.AdvanceToMain(t, g)
	// Without the {B} the activation is refused, and nothing is sacrificed.
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("no mana, no activation")
	}
	if !g.Battlefield.Contains(fodder) {
		t.Fatal("a refused activation sacrificed the creature")
	}
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	if err := g.ActivateCatalogAbility(me.ID, h, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the ability scries 2, got %+v", c)
	}
}

func neoB20ActivateHidetsuguExile(t *testing.T, g *game.Game, me *game.Player, h uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	me.ManaPool.AddMana(game.ManaToken{Color: "R"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, h, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

// A nonland card: exiled, playable this turn, and Hidetsugu deals damage
// equal to its mana value to the chosen target.
func TestNeoB20HidetsuguExilesANonlandCardAndDealsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	h := neoB20Push(g, me.ID, "Hidetsugu, Devouring Chaos", "Legendary Creature — Ogre Demon", "{3}{B}", neoHidetsuguOracle, 4, 4)
	top := effectstest.SeedLibraryTop(me, "Top Spell", "Sorcery", "{3}{R}")
	life := effectstest.LifeOf(g, opp.ID)
	neoB20ActivateHidetsuguExile(t, g, me, h)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if perm := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile); perm == nil {
		t.Error("the exiled card is not playable this turn")
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the reflexive trigger asks for a target")
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-4 {
		t.Errorf("opponent life %d → %d, want -4 (mana value 4)", life, got)
	}
}

// A land: exiled and playable, but no reflexive trigger.
func TestNeoB20HidetsuguExilingALandDealsNoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	h := neoB20Push(g, me.ID, "Hidetsugu, Devouring Chaos", "Legendary Creature — Ogre Demon", "{3}{B}", neoHidetsuguOracle, 4, 4)
	top := effectstest.SeedLibraryTop(me, "Mountain", "Basic Land — Mountain", "")
	life := effectstest.LifeOf(g, opp.ID)
	neoB20ActivateHidetsuguExile(t, g, me, h)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a land asks for no damage target")
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life {
		t.Errorf("opponent life changed %d → %d", life, got)
	}
}

const neoTatsunariOracle = "e87be082-eafd-480e-b150-d27fd937e1b1"

func TestNeoB20TatsunariIsRegisteredFull(t *testing.T) {
	neoB20Registered(t, neoTatsunariOracle, "Tatsunari, Toad Rider")
}

func neoB20Keimis(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Keimi" && effects.IsToken(c) && c.Controller == owner {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestNeoB20TatsunariMakesKeimiOnAnEnchantmentCastOnlyWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	neoB20Push(g, me.ID, "Tatsunari, Toad Rider", "Legendary Creature — Human Ninja", "{2}{B}", neoTatsunariOracle, 3, 3)

	// A creature spell does nothing.
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(neoB20Keimis(g, me.ID)); n != 0 {
		t.Fatalf("%d Keimi after a creature spell, want 0", n)
	}

	effectstest.CastCatalogSpell(t, g, "Plain Aura Thing", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	keimis := neoB20Keimis(g, me.ID)
	if len(keimis) != 1 {
		t.Fatalf("%d Keimi after an enchantment spell, want 1", len(keimis))
	}
	k := keimis[0]
	if p, tough := effectstest.EffectivePower(t, g, k), effectstest.EffectiveToughness(t, g, k); p != 3 || tough != 3 {
		t.Errorf("Keimi is %d/%d, want 3/3", p, tough)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == k && (!c.HasSupertype("legendary") || !c.HasSubtype("Frog")) {
			t.Errorf("Keimi should be a legendary Frog: %q", c.TypeLine)
		}
	}

	// With Keimi in play, a second enchantment makes no second Keimi; Keimi's own
	// trigger drains each opponent for 1.
	myLife, oppLife := effectstest.LifeOf(g, me.ID), effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Another Enchantment", "Enchantment", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(neoB20Keimis(g, me.ID)); n != 1 {
		t.Fatalf("%d Keimi, want still 1", n)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != oppLife-1 {
		t.Errorf("opponent life %d → %d, want -1 from Keimi", oppLife, got)
	}
	if got := effectstest.LifeOf(g, me.ID); got != myLife+1 {
		t.Errorf("your life %d → %d, want +1 from Keimi", myLife, got)
	}
}

func TestNeoB20TatsunariUnblockableExceptFlyingOrReachForItAndAFrog(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tats := neoB20Push(g, me.ID, "Tatsunari, Toad Rider", "Legendary Creature — Human Ninja", "{2}{B}", neoTatsunariOracle, 3, 3)
	frog := effectstest.B12Creature(g, me.ID, "Pond Frog", "Creature — Frog", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	ground := effectstest.BrCreature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	ground2 := effectstest.BrCreature(g, opp.ID, "Hill Giant Two", "Creature — Giant", 3, 3)
	ground3 := effectstest.BrCreature(g, opp.ID, "Hill Giant Three", "Creature — Giant", 3, 3)
	flyer := effectstest.BrCreature(g, opp.ID, "Wind Drake", "Creature — Drake", 2, 2, "flying")
	effectstest.AdvanceToMain(t, g)

	// The target must be a Frog you control.
	me.ManaPool.AddMana(game.ManaToken{Color: "G"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, tats, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other), Strict: true}); err == nil {
		t.Fatal("a non-Frog is not a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, tats, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(frog), Strict: true}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, tats, frog, other)
	effectstest.BrRefusal(t, g.DeclareBlocker(ground, tats), game.BlockReasonCantBeBlockedExceptBy)
	effectstest.BrRefusal(t, g.DeclareBlocker(ground2, frog), game.BlockReasonCantBeBlockedExceptBy)
	if err := g.DeclareBlocker(flyer, tats); err != nil {
		t.Fatalf("a flyer may block Tatsunari: %v", err)
	}
	if err := g.DeclareBlocker(ground3, other); err != nil {
		t.Fatalf("the rule did not cover the other creature: %v", err)
	}
}

const neoShadowOracle = "98636623-bbc6-4935-a2bb-2234691eb6d8"

func TestNeoB20ThousandFacedShadowIsRegisteredFull(t *testing.T) {
	spec := neoB20Registered(t, neoShadowOracle, "Thousand-Faced Shadow")
	if len(spec.Activated) != 1 || !spec.Activated[0].Ninjutsu {
		t.Errorf("expected exactly the ninjutsu ability, got %d activated", len(spec.Activated))
	}
}

func TestNeoB20ThousandFacedShadowNinjutsuCopiesAnotherAttackerTappedAndAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.B12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	giant := effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	shadow := effectstest.PushNinjaToHand(me, "Thousand-Faced Shadow", "Creature — Human Ninja", neoShadowOracle, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, rat, giant)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{U}{U}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, shadow, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{rat}, Strict: true}); err != nil {
		t.Fatalf("activate ninjutsu: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger asks for another attacking creature")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: shadow}); err == nil {
		t.Fatal("the Shadow may not target itself: another")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: giant}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	var copyID uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hill Giant" && effects.IsToken(c) && c.Controller == me.ID {
			copyID = c.InstanceID
		}
	}
	if copyID == uuid.Nil {
		t.Fatal("no token copy of Hill Giant")
	}
	c, _ := g.LookupCardForEffect(copyID)
	if !c.Tapped {
		t.Error("the copy did not enter tapped")
	}
	if c.AttackingTarget != opp.ID {
		t.Errorf("the copy attacks %s, want %s", c.AttackingTarget, opp.ID)
	}
	if p, tough := effectstest.EffectivePower(t, g, copyID), effectstest.EffectiveToughness(t, g, copyID); p != 3 || tough != 3 {
		t.Errorf("the copy is %d/%d, want 3/3", p, tough)
	}
	if effectstest.AttackEventsFor(g, copyID) != 0 {
		t.Error("the copy was never declared as an attacker (CR 506.3c)")
	}
}

// Cast from the hand rather than put in attacking, the Shadow's "if it's
// attacking" is false: no trigger, no target prompt.
func TestNeoB20ThousandFacedShadowHardCastDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Thousand-Faced Shadow", "Creature — Human Ninja", neoShadowOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("a Shadow that is not attacking asks for no target")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Hill Giant" && effects.IsToken(c) {
			t.Fatal("a copy was made though the Shadow was not attacking")
		}
	}
}

const neoHinataOracle = "aee473c7-c659-421c-96b5-05f57ca2ed97"

func TestNeoB20HinataIsRegisteredFull(t *testing.T) {
	spec := neoB20Registered(t, neoHinataOracle, "Hinata, Dawn-Crowned")
	if len(spec.CostModifiers) != 2 {
		t.Fatalf("Hinata declares %d cost modifiers, want 2", len(spec.CostModifiers))
	}
	for i, m := range spec.CostModifiers {
		if !m.ReadsTargets {
			t.Errorf("cost modifier %d does not declare ReadsTargets", i)
		}
	}
}

func neoB20Price(t *testing.T, g *game.Game, seat *game.Player, cost string, targets []game.TargetRef) game.ParsedCost {
	t.Helper()
	return effectstest.SelfPriced(t, g, seat, "", "Sorcery", cost, game.ZoneHand, targets)
}

// Every target takes {1} off a spell its controller casts and puts {1} on a
// spell an opponent casts; a targetless spell costs what it did.
func TestNeoB20HinataPricesSpellsByTheirTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neoB20Push(g, me.ID, "Hinata, Dawn-Crowned", "Legendary Creature — Kirin Spirit", "{1}{U}{R}{W}", neoHinataOracle, 4, 4)
	player := func(id uuid.UUID, slot int) game.TargetRef {
		return game.TargetRef{Kind: game.TargetPlayer, ID: id, Slot: slot}
	}
	one := []game.TargetRef{player(opp.ID, 0)}
	three := []game.TargetRef{player(opp.ID, 0), player(g.Seats[2].ID, 1), player(g.Seats[3].ID, 2)}

	cases := []struct {
		name    string
		seat    *game.Player
		cost    string
		targets []game.TargetRef
		want    int
	}{
		{"mine, no target", me, "{3}{R}", nil, 4},
		{"mine, one target", me, "{3}{R}", one, 3},
		{"mine, three targets", me, "{3}{R}", three, 1},
		{"mine, reduction takes generic mana only", me, "{1}{R}", three, 1},
		{"opponent's, no target", opp, "{3}{R}", nil, 4},
		{"opponent's, one target", opp, "{3}{R}", one, 5},
		{"opponent's, three targets", opp, "{3}{R}", three, 7},
	}
	for _, tc := range cases {
		if got := neoB20Price(t, g, tc.seat, tc.cost, tc.targets).ManaValue(); got != tc.want {
			t.Errorf("%s: mana value %d, want %d", tc.name, got, tc.want)
		}
	}
}

// The real cast: with Hinata, a two-target {1}{R} spell is paid with {R};
// without it the same pool is refused.
func TestNeoB20HinataDiscountIsPaidByARealCast(t *testing.T) {
	cast := func(withHinata bool) error {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withHinata {
			neoB20Push(g, me.ID, "Hinata, Dawn-Crowned", "Legendary Creature — Kirin Spirit", "{1}{U}{R}{W}", neoHinataOracle, 4, 4)
		}
		a := effectstest.PushCostedPermanentForTest(g, opp.ID, "Opponent A", "Creature — Bear", "{1}{G}")
		b := effectstest.PushCostedPermanentForTest(g, opp.ID, "Opponent B", "Creature — Bear", "{1}{G}")
		effectstest.AdvanceToMain(t, g)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Arc Trail", TypeLine: "Sorcery", ManaCost: "{1}{R}", OracleID: effectstest.ArcTrailOracle,
			Owner: me.ID, Controller: me.ID})
		me.ManaPool.AddMana(game.ManaToken{Color: "R"})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1}}})
	}
	if err := cast(true); err != nil {
		t.Fatalf("with Hinata the two-target spell costs {R}: %v", err)
	}
	if err := cast(false); err == nil {
		t.Fatal("without Hinata {R} cannot pay {1}{R}")
	}
}

const neoTamiyoOracle = "021b4745-4a15-462c-a6b0-2dee1dc9d196"

func TestNeoB20TamiyoIsRegisteredFull(t *testing.T) {
	spec := neoB20Registered(t, neoTamiyoOracle, "Tamiyo, Compleated Sage")
	if spec.StartingLoyalty != 5 || len(spec.Activated) != 3 {
		t.Errorf("loyalty %d with %d abilities, want 5 with 3", spec.StartingLoyalty, len(spec.Activated))
	}
}

// {G/U/P} paid with life: two fewer loyalty counters (Compleated), and 2
// life; paid with mana it enters with all five.
func TestNeoB20TamiyoCompleatedEntersWithTwoFewerWhenLifeIsPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	tamiyo := effectstest.S82CastWalker(t, g, "Tamiyo, Compleated Sage", neoTamiyoOracle, "Legendary Planeswalker — Tamiyo",
		"{2}{G}{G/U/P}{U}", 5, 1, "CCGU")
	if got := effectstest.LoyaltyOf(g, tamiyo); got != 3 {
		t.Errorf("loyalty %d after paying life for {G/U/P}, want 3", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d → %d, want 2 paid", life, me.Life)
	}

	g = effectstest.NewCatalogGame(t)
	tamiyo = effectstest.S82CastWalker(t, g, "Tamiyo, Compleated Sage", neoTamiyoOracle, "Legendary Planeswalker — Tamiyo",
		"{2}{G}{G/U/P}{U}", 5, 0, "CCGGU")
	if got := effectstest.LoyaltyOf(g, tamiyo); got != 5 {
		t.Errorf("loyalty %d after paying mana, want 5", got)
	}
}

func TestNeoB20TamiyoPlusOneTapsAndFreezesAnArtifactOrCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	tamiyo := effectstest.S82Walker(g, me.ID, "Tamiyo, Compleated Sage", neoTamiyoOracle, 5)
	bear := effectstest.PushCreatureToBattlefieldForTest(g, opp.ID, "Bear")
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	land := effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(land), Strict: true}); err == nil {
		t.Fatal("a land is neither an artifact nor a creature")
	}
	effectstest.S82Activate(t, g, me.ID, tamiyo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, bear)
	if !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Errorf("creature: tapped=%v skips=%d, want tapped with one untap skip", c.Tapped, len(c.NextUntapSkips))
	}
	if got := effectstest.LoyaltyOf(g, tamiyo); got != 6 {
		t.Errorf("loyalty %d after +1, want 6", got)
	}

	// An artifact is as legal.
	g2 := effectstest.NewCatalogGame(t)
	seat = g2.Turn.ActiveSeat
	me, opp = g2.Seats[seat], g2.Seats[(seat+1)%len(g2.Seats)]
	tamiyo = effectstest.S82Walker(g2, me.ID, "Tamiyo, Compleated Sage", neoTamiyoOracle, 5)
	rock := effectstest.PushPermanent(g2, opp.ID, game.Card{Name: "Mind Stone", TypeLine: "Artifact", ManaCost: "{2}"})
	effectstest.AdvanceToMainOf(t, g2, seat)
	effectstest.S82Activate(t, g2, me.ID, tamiyo, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(rock)})
	effectstest.PassPriorityAroundTable(t, g2)
	if c := effectstest.LayeredCard(t, g2, rock); !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Errorf("artifact: tapped=%v skips=%d, want tapped with one untap skip", c.Tapped, len(c.NextUntapSkips))
	}
}

func TestNeoB20TamiyoMinusXExilesAMatchingCardAndCopiesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	tamiyo := effectstest.S82Walker(g, me.ID, "Tamiyo, Compleated Sage", neoTamiyoOracle, 5)
	three := neoB20Graveyard(me, "Three Drop", "Creature — Bear", "{2}{G}")
	two := neoB20Graveyard(me, "Two Drop", "Artifact", "{2}")
	spell := neoB20Graveyard(me, "Three Spell", "Sorcery", "{2}{U}")
	land := neoB20Graveyard(me, "Forest", "Basic Land — Forest", "")
	effectstest.AdvanceToMainOf(t, g, seat)

	// Wrong mana value, a nonpermanent card, and a land are all refused.
	for name, id := range map[string]uuid.UUID{"mana value 2 at X=3": two, "an instant/sorcery": spell} {
		if err := g.ActivateCatalogAbility(me.ID, tamiyo, 1, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(id), Strict: true}); err == nil {
			t.Errorf("%s was accepted", name)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, tamiyo, 1, game.ActivateAbilityParams{XValue: 0, Targets: effectstest.CardRefs(land), Strict: true}); err == nil {
		t.Error("a land card was accepted at X=0")
	}
	if got := effectstest.LoyaltyOf(g, tamiyo); got != 5 {
		t.Fatalf("refused activations moved loyalty to %d", got)
	}

	effectstest.S82Activate(t, g, me.ID, tamiyo, 1, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(three)})
	if got := effectstest.LoyaltyOf(g, tamiyo); got != 2 {
		t.Errorf("loyalty %d after −3, want 2", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(three) {
		t.Error("the card was not exiled")
	}
	var token *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name == "Three Drop" && effects.IsToken(*c) && c.Controller == me.ID {
			token = c
		}
	}
	if token == nil {
		t.Fatal("no token copy of the exiled card")
	}
}

func TestNeoB20TamiyoMinusSevenMakesTheNotebook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	tamiyo := effectstest.S82Walker(g, me.ID, "Tamiyo, Compleated Sage", neoTamiyoOracle, 7)
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.S82Activate(t, g, me.ID, tamiyo, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var book uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Tamiyo's Notebook" && effects.IsToken(c) && c.Controller == me.ID {
			book = c.InstanceID
			if !c.HasSupertype("legendary") || !c.IsArtifact() || !c.HasSubtype("Book") {
				t.Errorf("the Notebook is %q, want a legendary Book artifact", c.TypeLine)
			}
		}
	}
	if book == uuid.Nil {
		t.Fatal("no Tamiyo's Notebook")
	}
	if g.Battlefield.Contains(tamiyo) {
		t.Error("−7 from 7 loyalty leaves Tamiyo at 0, and she dies to state-based actions")
	}

	// {T}: Draw a card.
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, book, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Notebook draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1", hand, me.Hand.Size())
	}

	// "Spells you cast cost {2} less to cast", and only yours.
	opp := g.Seats[(seat+1)%len(g.Seats)]
	if got := neoB20Price(t, g, me, "{3}{R}", nil).ManaValue(); got != 2 {
		t.Errorf("your {3}{R} spell costs %d, want 2", got)
	}
	if got := neoB20Price(t, g, opp, "{3}{R}", nil).ManaValue(); got != 4 {
		t.Errorf("an opponent's {3}{R} spell costs %d, want 4", got)
	}
}

const neoKaitoOracle = "ba48e067-f4d9-45b2-ade1-71cf8cda418c"

func TestNeoB20KaitoShizukiIsRegisteredFull(t *testing.T) {
	spec := neoB20Registered(t, neoKaitoOracle, "Kaito Shizuki")
	if spec.StartingLoyalty != 3 || len(spec.Activated) != 3 || spec.Emblem == nil {
		t.Errorf("loyalty %d, %d abilities, emblem %v; want 3, 3 and an emblem", spec.StartingLoyalty, len(spec.Activated), spec.Emblem != nil)
	}
}

// Cast this turn, Kaito phases out at the beginning of his controller's end
// step, and phases back in at their next untap step.
func TestNeoB20KaitoPhasesOutAtTheEndStepOfTheTurnHeEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kaito := effectstest.S82CastWalker(t, g, "Kaito Shizuki", neoKaitoOracle, "Legendary Planeswalker — Kaito", "{1}{U}{B}", 3, 0, "CUB")
	if !g.Battlefield.Contains(kaito) || effectstest.LoyaltyOf(g, kaito) != 3 {
		t.Fatalf("Kaito did not enter with 3 loyalty")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.PhasedOut.Contains(kaito) {
		t.Fatal("Kaito did not phase out at the end step of the turn he entered")
	}
	if me.Hand.Contains(kaito) || me.Graveyard.Contains(kaito) {
		t.Error("phasing out moved Kaito out of the game's zones")
	}
}

// A Kaito that has been on the battlefield since an earlier turn does not
// phase out: the intervening "if" is false.
func TestNeoB20KaitoThatEnteredEarlierStaysPut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito Shizuki", neoKaitoOracle, 3)
	// Move to the next turn of this seat so Kaito has been there since an earlier turn.
	start := g.Turn.Seq
	for i := 0; i < 400; i++ {
		if g.Turn.Seq > start && g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepPrecombatMain {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if g.Turn.Seq <= start || g.Turn.ActiveSeat != seat {
		t.Fatalf("never reached a later turn of the seat (turn %d, started at %d)", g.Turn.Seq, start)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.PhasedOut.Contains(kaito) {
		t.Fatal("Kaito phased out though he did not enter this turn")
	}
}

func TestNeoB20KaitoPlusOneDrawsThenDiscardsUnlessYouAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito Shizuki", neoKaitoOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, kaito, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, kaito); got != 4 {
		t.Errorf("loyalty %d after +1, want 4", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 before the discard", hand, me.Hand.Size())
	}
	if c := effectstest.DiscardChoiceFor(g, me.ID); c == nil {
		t.Fatal("you did not attack this turn, so you discard a card")
	}
}

func TestNeoB20KaitoPlusOneAfterAttackingSkipsTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito Shizuki", neoKaitoOracle, 3)
	atk := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	hand := me.Hand.Size()
	effectstest.S82Activate(t, g, me.ID, kaito, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want +1 with no discard", hand, me.Hand.Size())
	}
	if c := effectstest.DiscardChoiceFor(g, me.ID); c != nil {
		t.Fatal("you attacked this turn, so no discard is asked")
	}
}

func TestNeoB20KaitoMinusTwoMakesAnUnblockableNinja(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito Shizuki", neoKaitoOracle, 3)
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.S82Activate(t, g, me.ID, kaito, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, kaito); got != 1 {
		t.Errorf("loyalty %d after −2, want 1", got)
	}
	var ninja uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ninja" && effects.IsToken(c) && c.Controller == me.ID {
			ninja = c.InstanceID
		}
	}
	if ninja == uuid.Nil {
		t.Fatal("no Ninja token")
	}
	if p, tough := effectstest.EffectivePower(t, g, ninja), effectstest.EffectiveToughness(t, g, ninja); p != 1 || tough != 1 {
		t.Errorf("the Ninja is %d/%d, want 1/1", p, tough)
	}
	if effectstest.RestrictionsOf(t, g, ninja)&game.CantBeBlocked == 0 {
		t.Error("the Ninja can be blocked")
	}
}

func TestNeoB20KaitoMinusSevenEmblemFetchesABlueOrBlackCreatureOnEveryHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	kaito := effectstest.S82Walker(g, me.ID, "Kaito Shizuki", neoKaitoOracle, 7)
	green := effectstest.PlTop(me, "Green Bear", "Creature — Bear", "{1}{G}")
	spell := effectstest.PlTop(me, "Blue Spell", "Instant", "{U}")
	blue := effectstest.PlTop(me, "Blue Drake", "Creature — Drake", "{1}{U}")
	black := effectstest.PlTop(me, "Black Bat", "Creature — Bat", "{1}{B}")
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.S82Activate(t, g, me.ID, kaito, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(effectstest.EmblemsOf(g, me.ID)) != 1 {
		t.Fatalf("emblems = %d, want 1", len(effectstest.EmblemsOf(g, me.ID)))
	}
	a := effectstest.PushVanillaCreature(g, me.ID, "Attacker A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Attacker B", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the first creature to connect starts a search")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if offered[green] || offered[spell] {
		t.Errorf("only blue or black creature cards are offered: %v", offered)
	}
	if !offered[blue] || !offered[black] {
		t.Errorf("the blue and the black creature are offered: %v", offered)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, blue)
	effectstest.PassPriorityAroundTable(t, g)
	// The second attacker's damage is its own event: a second search.
	if c2 := effectstest.SearchChoiceFor(g, me.ID); c2 == nil {
		t.Fatal("a second creature connecting triggers the emblem again")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, black)
	if !g.Battlefield.Contains(blue) || !g.Battlefield.Contains(black) {
		t.Error("the found creatures are not on the battlefield")
	}
	if g.Battlefield.Contains(green) {
		t.Error("the green creature came in")
	}
}
