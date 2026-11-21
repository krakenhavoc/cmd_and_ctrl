package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b42_test.go — Foundations Jumpstart cards, S82 BNS batch 42.

const (
	bns42Weaver     = "28274c39-19e2-4b4c-bd53-d31d6063c2ba"
	bns42Wildfire   = "76ff6653-f9b9-4bfe-92a9-52bda275b340"
	bns42WindStr    = "9d57cda6-9042-44bd-9785-38de9487ff85"
	bns42Windcaller = "1da1aa95-b587-4bc3-8bc1-7daa1fc33f00"
	bns42Wingsteed  = "9c46adaa-ab39-4c0a-99bd-935b6c4f732f"
	bns42Woodland   = "675321d7-2cf0-4e0b-9517-d711b22865ab"
	bns42Wavecrash  = "30868d69-2ba3-4373-9d47-4bc8978f35d1"
	bns42Zealot     = "2911d52f-7faf-455a-acdf-77269c9ee8ce"
	bns42Conscripts = "1dae6f39-2cbd-485c-b190-017a26401fd4"
	bns42Frontline  = "4c340b82-22e1-445a-b236-82471b442031"
	bns42Fungal     = "0a85402a-e3be-4ebb-8001-9d9a20dde7e8"
	bns42BruteForce = "9880ba09-d5b8-4675-bfb4-2161d86d2d41"
)

func TestBNS42CardsAreRegisteredAndWhole(t *testing.T) {
	for _, id := range []string{bns42Weaver, bns42Wildfire, bns42WindStr, bns42Windcaller, bns42Wingsteed,
		bns42Woodland, bns42Wavecrash, bns42Zealot, bns42Conscripts, bns42Frontline, bns42Fungal} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s: not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func bns42Damage(g *game.Game, id uuid.UUID) int {
	n := 0
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			n = c.DamageMarked
		}
	})
	return n
}

func bns42Float(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
}

// bns42CastCreature casts a creature from hand with `mana` floating and
// lets it and its enter trigger go.
func bns42CastCreature(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle, mana string) uuid.UUID {
	t.Helper()
	bns42Float(t, g, p, mana)
	id := effectstest.CastFromHandForTest(t, g, p, name, typeLine, cost, oracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestBNS42WindStriderHasFlashAndFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	// On an opponent's turn only a flash spell can be cast.
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	bns42Float(t, g, me, "{C}{C}{C}{C}{U}")
	id := effectstest.CastFromHandForTest(t, g, me, "Wind Strider", "Creature — Merfolk Wizard", "{4}{U}", bns42WindStr, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, id) {
		t.Fatal("Wind Strider did not resolve at instant speed")
	}
	if !effectstest.P1g2HasKeyword(g, id, "flying") || !effectstest.P1g2HasKeyword(g, id, "flash") {
		t.Error("Wind Strider lacks flying or flash")
	}
	// A creature without flash is refused on the same turn.
	bns42Float(t, g, me, "{C}{C}{C}{U}")
	hand := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Yarok's Wavecrasher", TypeLine: "Creature — Elemental",
			ManaCost: "{3}{U}", OracleID: bns42Wavecrash, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature without flash was cast on an opponent's turn")
	}
	_ = opp
}

func TestBNS42WeaverOfLightningShootsOnInstantsAndSorceries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Weaver of Lightning", "Creature — Human Shaman", bns42Weaver, false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 4)

	// A creature spell does not trigger it.
	creature := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventCast, Actor: me.ID, CardID: creature}) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered the Weaver")
	}

	effectstest.Std53CastEvent(t, g, me) // an instant
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == mine {
			t.Fatal("the Weaver may target a creature its controller controls")
		}
	}
	effectstest.PickTriggerTarget(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if d := bns42Damage(g, foe); d != 1 {
		t.Errorf("the foe has %d damage, want 1", d)
	}
}

func TestBNS42WindcallerAvenCycleGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	handBefore := me.Hand.Size()
	id, _ := effectstest.CycleFromHand(t, g, "Windcaller Aven", "Creature — Bird Wizard", bns42Windcaller, "{U}")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.P1g2HasKeyword(g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
	if me.Hand.Size() != handBefore+1 {
		t.Errorf("hand = %d, want %d (the card drawn)", me.Hand.Size(), handBefore+1)
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == id {
			return
		}
	}
	t.Error("Windcaller Aven is not in the graveyard")
}

func TestBNS42WingsteedRiderIsHeroic(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	rider := effectstest.PushCatalogPermanent(g, me.ID, "Wingsteed Rider", "Creature — Human Knight", bns42Wingsteed, false)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)

	bns42Float(t, g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Brute Force", "Instant", "{R}", bns42BruteForce,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: other}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, rider); n != 0 {
		t.Fatalf("a spell targeting another creature put %d counters on the Rider", n)
	}

	bns42Float(t, g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Brute Force", "Instant", "{R}", bns42BruteForce,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rider}}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, rider); n != 1 {
		t.Errorf("Rider counters = %d, want 1", n)
	}
}

func TestBNS42WildfireElementalAndZealotOfTheGodPharaoh(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Wildfire Elemental", "Creature — Elemental", bns42Wildfire, false)
	zealot := effectstest.PushCatalogPermanent(g, me.ID, "Zealot of the God-Pharaoh", "Creature — Minotaur Archer", bns42Zealot, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	base, _ := effectstest.EffectivePTOf(t, g, bear)

	// Damage to your own face is not damage to an opponent: refused as a target.
	bns42Float(t, g, me, "{C}{C}{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, zealot, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, Strict: true}); err == nil {
		t.Fatal("Zealot may target its controller")
	}
	before := effectstest.LifeOf(g, opp.ID)
	if err := g.ActivateCatalogAbility(me.ID, zealot, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != before-2 {
		t.Errorf("opponent life = %d, want %d", got, before-2)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != base+1 {
		t.Errorf("Bear power = %d, want %d", p, base+1)
	}
}

func TestBNS42WoodlandChampionCountsTokensAndFungalPlotsMakesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	champ := effectstest.PushCatalogPermanent(g, me.ID, "Woodland Champion", "Creature — Elf Scout", bns42Woodland, false)
	plots := effectstest.PushCatalogPermanent(g, me.ID, "Fungal Plots", "Enchantment", bns42Fungal, false)

	// A nontoken creature entering does nothing.
	effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, champ); n != 0 {
		t.Fatalf("a nontoken gave %d counters", n)
	}

	// Two tokens at once are one trigger worth two counters.
	if err := g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 green Saproling"), 2); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, champ); n != 2 {
		t.Fatalf("two tokens gave %d counters, want 2", n)
	}

	// Fungal Plots: exile a creature card from the graveyard for a Saproling.
	bns42Float(t, g, me, "{C}{G}")
	if err := g.ActivateCatalogAbility(me.ID, plots, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no creature card in the graveyard")
	}
	gy := effectstest.PushCatalogGraveyardCard(me, "Dead Bear", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, plots, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{gy}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.PlusOneCounters(g, champ); n != 3 {
		t.Errorf("counters = %d after the Saproling token, want 3", n)
	}
	saps := func() []uuid.UUID {
		var ids []uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == "Saproling" {
					ids = append(ids, c.InstanceID)
				}
			}
		})
		return ids
	}
	ids := saps()
	if len(ids) != 3 {
		t.Fatalf("Saprolings = %d, want 3", len(ids))
	}

	// Sacrifice two Saprolings: gain 2 life, draw a card.
	life, hand := effectstest.LifeOf(g, me.ID), me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, plots, 1, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: ids[:2]}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LifeOf(g, me.ID) != life+2 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d→%d, hand %d→%d", life, effectstest.LifeOf(g, me.ID), hand, me.Hand.Size())
	}
}

func TestBNS42YaroksWavecrasherBouncesAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	// Alone it has nothing to return and stays.
	lone := bns42CastCreature(t, g, me, "Yarok's Wavecrasher", "Creature — Elemental", "{3}{U}", bns42Wavecrash, "{C}{C}{C}{U}")
	if !effectstest.BattlefieldHas(g, lone) {
		t.Fatal("the lone Wavecrasher left the battlefield")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Fatal("asked to return a creature with no other creature")
	}

	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	second := bns42CastCreature(t, g, me, "Yarok's Wavecrasher", "Creature — Elemental", "{3}{U}", bns42Wavecrash, "{C}{C}{C}{U}")
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no prompt to return a creature")
	}
	if effectstest.HasID(pick.ChooseCards, second) {
		t.Error("the new Wavecrasher itself is offered")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("the return is not mandatory: bounds %d..%d", pick.ChooseMin, pick.ChooseMax)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{other}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, other) {
		t.Error("the chosen creature stayed on the battlefield")
	}
}

func TestBNS42ZealousConscriptsStealsAnyPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == land {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	bns42CastCreature(t, g, me, "Zealous Conscripts", "Creature — Human Warrior", "{4}{R}", bns42Conscripts, "{C}{C}{C}{C}{R}")
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, land) != me.ID {
		t.Fatal("control of the land did not change")
	}
	if c, _ := effectstest.CardOnBattlefield(g, land); c.Tapped {
		t.Error("the stolen land is still tapped")
	}
	if !effectstest.P1g2HasKeyword(g, land, "haste") {
		t.Error("the stolen permanent has no haste")
	}
	// It goes home at end of turn.
	for g.Turn.Step != game.StepCleanup {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, land) != opp.ID {
		t.Error("the land did not return at end of turn")
	}
}

func TestBNS42FrontlineHeroismMakesSoldiersAndCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	soldiers := func() []uuid.UUID {
		var ids []uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.IsToken() && c.Name == "Soldier" {
					ids = append(ids, c.InstanceID)
				}
			}
		})
		return ids
	}
	bns42CastCreature(t, g, me, "Frontline Heroism", "Enchantment", "{2}{R}", bns42Frontline, "{C}{C}{R}")
	if n := len(soldiers()); n != 1 {
		t.Fatalf("enter trigger made %d Soldiers, want 1", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	bns42Float(t, g, me, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Brute Force", "Instant", "{R}", bns42BruteForce,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	for i := 0; i < 6; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	ids := soldiers()
	if len(ids) != 2 {
		t.Fatalf("Soldiers = %d, want 2", len(ids))
	}
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 5 {
		t.Errorf("Bear power = %d, want 5 (one Brute Force)", p)
	}
	boosted := 0
	for _, id := range ids {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p == 4 {
			boosted++
		}
		if !effectstest.P1g2HasKeyword(g, id, "haste") {
			t.Error("a Soldier lacks haste")
		}
	}
	if boosted != 1 {
		t.Errorf("%d Soldiers got +3/+3 from the copy, want exactly 1", boosted)
	}

}
