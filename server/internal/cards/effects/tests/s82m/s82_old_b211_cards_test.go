package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b211_cards_test.go — Dominaria cards, S82 OLD batch 211 (#5530),
// played through the real catalog entries.

const (
	oldB211ChainwhirlerOracle = "0342b085-be20-475a-bc38-59b9cb6012e8"
	oldB211AdeptOracle        = "0f513242-b175-4b4d-a05f-78b1b457a020"
	oldB211JourneymageOracle  = "5202c00a-5e47-43fe-9ed7-52c276120df3"
	oldB211HoundOracle        = "7af53dd7-3994-48f1-9ae1-44380ac56ee0"
	oldB211PierceOracle       = "7e4504c1-4349-408d-8fdf-0c44efdd3f56"
	oldB211AdelizOracle       = "850eb3c3-029b-49b0-91a9-6daeb1b3a9e8"
	oldB211SentinelOracle     = "bf24c063-dd9a-4381-b94e-682b0cb0c03d"
	oldB211CyclopsOracle      = "cee8ad22-12a7-4da6-a7d9-e205ada16489"
	oldB211CandleOracle       = "da9b3edc-bc2d-4ff5-b614-3626f504c4e9"
	oldB211BloodstoneOracle   = "ec17b9db-dd80-4b46-a1b5-215f26bf6498"
	oldB211KickerSerpent      = "c39c1604-3bae-454d-9985-85101e51ec6e"
)

func TestOldB211CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB211ChainwhirlerOracle: "Goblin Chainwhirler",
		oldB211AdeptOracle:        "Firefist Adept",
		oldB211JourneymageOracle:  "Ghitu Journeymage",
		oldB211HoundOracle:        "Karplusan Hound",
		oldB211PierceOracle:       "Pierce the Sky",
		oldB211AdelizOracle:       "Adeliz, the Cinder Wind",
		oldB211SentinelOracle:     "Teferi's Sentinel",
		oldB211CyclopsOracle:      "Rampaging Cyclops",
		oldB211CandleOracle:       "Bloodtallow Candle",
		oldB211BloodstoneOracle:   "Bloodstone Goblin",
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

func oldB211Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB211PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Goblin Chainwhirler", "Creature — Goblin Warrior", oldB211ChainwhirlerOracle, 3, 3, []string{"first strike"}},
		{"Adeliz, the Cinder Wind", "Legendary Creature — Human Wizard", oldB211AdelizOracle, 2, 2, []string{"flying", "haste"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

func TestOldB211GoblinChainwhirlerHitsEachOpponentAndTheirCreaturesAndPlaneswalkersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	thirds := effectstest.PushVanillaCreature(g, third.ID, "Third Bear", 3, 3)
	pw := effectstest.PushCatalogPermanent(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", false)
	effectstest.SetCounters(g, pw, map[string]int{game.CounterLoyalty: 4})
	myPW := effectstest.PushCatalogPermanent(g, me.ID, "My Walker", "Legendary Planeswalker — Test", "", false)
	effectstest.SetCounters(g, myPW, map[string]int{game.CounterLoyalty: 4})
	myLife := me.Life
	oppLife, thirdLife := opp.Life, third.Life

	effectstest.CastCatalogSpell(t, g, "Goblin Chainwhirler", "Creature — Goblin Warrior", oldB211ChainwhirlerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if opp.Life != oppLife-1 || third.Life != thirdLife-1 {
		t.Errorf("opponent lives %d, %d, want %d, %d", opp.Life, third.Life, oppLife-1, thirdLife-1)
	}
	if me.Life != myLife {
		t.Errorf("the controller took damage: %d -> %d", myLife, me.Life)
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 3 {
		t.Errorf("opposing planeswalker loyalty = %d, want 3", got)
	}
	if got := effectstest.CounterCount(g, myPW, game.CounterLoyalty); got != 4 {
		t.Errorf("own planeswalker loyalty = %d, want 4 (untouched)", got)
	}
	if effectstest.BattlefieldHas(g, theirs) == false || effectstest.BattlefieldHas(g, mine) == false || effectstest.BattlefieldHas(g, thirds) == false {
		t.Error("a 2/2 or 3/3 died to 1 damage")
	}
	// 1 damage kills an X/1 opposing creature and spares the controller's own.
}

func TestOldB211GoblinChainwhirlerKillsOpposingXOneCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Mouse", 1, 1)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Mouse", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Goblin Chainwhirler", "Creature — Goblin Warrior", oldB211ChainwhirlerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, theirs) {
		t.Error("the opponent's 1/1 survived")
	}
	if !effectstest.BattlefieldHas(g, mine) {
		t.Error("the controller's own 1/1 died")
	}
}

func TestOldB211FirefistAdeptDealsDamageEqualToYourWizards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	effectstest.B12Push(g, me.ID, "Other Wizard", "Creature — Human Wizard", "", 1, 1)
	effectstest.B12Push(g, me.ID, "Not a Wizard", "Creature — Bear", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Their Wizard", "Creature — Human Wizard", "", 1, 1)
	target := effectstest.PushVanillaCreature(g, opp.ID, "Big Target", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Firefist Adept", "Creature — Human Wizard", oldB211AdeptOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, target); got != 2 {
		t.Fatalf("damage on the target = %d, want 2 (the Adept and one other Wizard of mine)", got)
	}
}

func TestOldB211GhituJourneymageNeedsAnotherWizard(t *testing.T) {
	for _, other := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB211Seats(g)
		if other {
			effectstest.B12Push(g, me.ID, "Other Wizard", "Creature — Human Wizard", "", 1, 1)
		}
		life := opp.Life
		effectstest.CastCatalogSpell(t, g, "Ghitu Journeymage", "Creature — Human Wizard", oldB211JourneymageOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if other {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("another Wizard %v: opponent life %d, want %d", other, opp.Life, want)
		}
	}
}

func TestOldB211KarplusanHoundShootsOnlyWithAChandraPlaneswalker(t *testing.T) {
	for _, chandra := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB211Seats(g)
		hound := effectstest.B12Push(g, me.ID, "Karplusan Hound", "Creature — Dog", oldB211HoundOracle, 3, 3)
		if chandra {
			pw := effectstest.PushCatalogPermanent(g, me.ID, "Chandra, Test", "Legendary Planeswalker — Chandra", "", false)
			effectstest.SetCounters(g, pw, map[string]int{game.CounterLoyalty: 4})
		} else {
			pw := effectstest.PushCatalogPermanent(g, me.ID, "Chandra Lookalike", "Legendary Planeswalker — Jace", "", false)
			effectstest.SetCounters(g, pw, map[string]int{game.CounterLoyalty: 4})
		}
		life := opp.Life
		effectstest.DeclareAttack(t, g, opp.ID, hound)
		if chandra {
			pick := effectstest.LatestPickTarget(g, me.ID)
			if pick == nil {
				t.Fatal("no target prompt with a Chandra planeswalker")
			}
			effectstest.PickTriggerTargetPlayer(t, g, me.ID, opp.ID)
		} else if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("the trigger asked for a target with no Chandra planeswalker")
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := life
		if chandra {
			want = life - 2
		}
		if opp.Life != want {
			t.Errorf("Chandra %v: opponent life %d (before combat damage), want %d", chandra, opp.Life, want)
		}
	}
}

func TestOldB211PierceTheSkyHitsOnlyCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB211Seats(g)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Test Flyer", TypeLine: "Creature — Bird",
		Power: 3, Toughness: 8, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	ground := effectstest.PushVanillaCreature(g, opp.ID, "Ground Bear", 3, 8)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Pierce the Sky", "Instant", oldB211PierceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ground}}); err == nil {
		t.Fatal("Pierce the Sky targeted a creature without flying")
	}
	effectstest.CastCatalogSpell(t, g, "Pierce the Sky", "Instant", oldB211PierceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: flyer}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, flyer); got != 7 {
		t.Fatalf("damage on the flyer = %d, want 7", got)
	}
}

func TestOldB211AdelizPumpsWizardsWhenYouCastAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	adeliz := effectstest.B12Push(g, me.ID, "Adeliz, the Cinder Wind", "Legendary Creature — Human Wizard", oldB211AdelizOracle, 2, 2)
	wizard := effectstest.B12Push(g, me.ID, "Other Wizard", "Creature — Human Wizard", "", 1, 1)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirWizard := effectstest.B12Push(g, opp.ID, "Their Wizard", "Creature — Human Wizard", "", 1, 1)

	effectstest.CastCatalogSpell(t, g, "Test Creature Spell", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, wizard); p != 1 {
		t.Fatalf("a creature spell pumped the Wizards: power %d", p)
	}
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID][2]int{adeliz: {3, 3}, wizard: {2, 2}, bear: {2, 2}, theirWizard: {1, 1}} {
		if p, tt := effectstest.PtOf(t, g, id); p != want[0] || tt != want[1] {
			t.Errorf("creature %s is %d/%d, want %d/%d", id, p, tt, want[0], want[1])
		}
	}
	effectstest.EndTurn(t, g)
	if p, _ := effectstest.PtOf(t, g, adeliz); p != 2 {
		t.Errorf("the pump outlasted the turn: Adeliz power %d", p)
	}
}

func TestOldB211TeferisSentinelGetsPlusFourOnlyWithATeferiPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	sentinel := effectstest.B12Push(g, me.ID, "Teferi's Sentinel", "Artifact Creature — Golem", oldB211SentinelOracle, 2, 6)
	if p, tt := effectstest.PtOf(t, g, sentinel); p != 2 || tt != 6 {
		t.Fatalf("Sentinel alone is %d/%d, want 2/6", p, tt)
	}
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Teferi, Theirs", "Legendary Planeswalker — Teferi", "", false)
	effectstest.SetCounters(g, theirs, map[string]int{game.CounterLoyalty: 4})
	if p, _ := effectstest.PtOf(t, g, sentinel); p != 2 {
		t.Fatalf("an opponent's Teferi pumped the Sentinel: power %d", p)
	}
	notTeferi := effectstest.PushCatalogPermanent(g, me.ID, "Teferi Lookalike", "Legendary Planeswalker — Jace", "", false)
	effectstest.SetCounters(g, notTeferi, map[string]int{game.CounterLoyalty: 4})
	if p, _ := effectstest.PtOf(t, g, sentinel); p != 2 {
		t.Fatalf("a non-Teferi planeswalker pumped the Sentinel: power %d", p)
	}
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Teferi, Mine", "Legendary Planeswalker — Teferi", "", false)
	effectstest.SetCounters(g, mine, map[string]int{game.CounterLoyalty: 4})
	if p, tt := effectstest.PtOf(t, g, sentinel); p != 6 || tt != 6 {
		t.Fatalf("Sentinel with my Teferi is %d/%d, want 6/6", p, tt)
	}
}

func TestOldB211RampagingCyclopsShrinksWhileDoubleBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB211Seats(g)
	cyclops := effectstest.B12Push(g, me.ID, "Rampaging Cyclops", "Creature — Cyclops", oldB211CyclopsOracle, 4, 4)
	_, opp := oldB211Seats(g)
	b1 := effectstest.PushVanillaCreature(g, opp.ID, "Blocker One", 0, 9)
	b2 := effectstest.PushVanillaCreature(g, opp.ID, "Blocker Two", 0, 9)
	effectstest.ReqToBlockers(t, g, cyclops)
	if p, _ := effectstest.PtOf(t, g, cyclops); p != 4 {
		t.Fatalf("unblocked Cyclops power = %d, want 4", p)
	}
	if err := effectstest.ReqBlock(t, g, cyclops, b1, b2); err != nil {
		t.Fatalf("double block: %v", err)
	}
	// The blocks are staged until the declaration is locked in (CR 509.1).
	if err := g.PassPriority(); err != nil {
		t.Fatalf("lock in the blocks: %v", err)
	}
	if p, tt := effectstest.PtOf(t, g, cyclops); p != 2 || tt != 4 {
		t.Fatalf("double-blocked Cyclops is %d/%d, want 2/4", p, tt)
	}
}

func TestOldB211RampagingCyclopsKeepsItsPowerWhenSingleBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB211Seats(g)
	cyclops := effectstest.B12Push(g, me.ID, "Rampaging Cyclops", "Creature — Cyclops", oldB211CyclopsOracle, 4, 4)
	_, opp := oldB211Seats(g)
	b1 := effectstest.PushVanillaCreature(g, opp.ID, "Blocker One", 0, 9)
	effectstest.ReqToBlockers(t, g, cyclops)
	if err := effectstest.ReqBlock(t, g, cyclops, b1); err != nil {
		t.Fatalf("single block: %v", err)
	}
	if p, _ := effectstest.PtOf(t, g, cyclops); p != 4 {
		t.Fatalf("single-blocked Cyclops power = %d, want 4", p)
	}
}

func TestOldB211BloodtallowCandleSpendsItselfToShrinkACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB211Seats(g)
	effectstest.ToMain(t, g)
	candle := effectstest.PushCatalogPermanent(g, me.ID, "Bloodtallow Candle", "Artifact", oldB211CandleOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 5, 5)
	sturdy := effectstest.PushVanillaCreature(g, opp.ID, "Sturdy", 6, 6)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, candle, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, candle) {
		t.Error("the Candle was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("the 5/5 survived -5/-5")
	}
	if !effectstest.BattlefieldHas(g, sturdy) {
		t.Error("an untargeted creature died")
	}
}

func TestOldB211BloodstoneGoblinGrowsOnlyForKickedSpells(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, _ := oldB211Seats(g)
		effectstest.AdvanceToMain(t, g)
		goblin := effectstest.B12Push(g, me.ID, "Bloodstone Goblin", "Creature — Goblin Warrior", oldB211BloodstoneOracle, 2, 2)
		effectstest.AddPool(t, g, me.ID, "{U}{U}{U}{U}")
		var opt []int
		if kicked {
			opt = []int{0}
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Serpent", TypeLine: "Creature — Serpent", ManaCost: "{U}",
			OracleID: oldB211KickerSerpent, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: opt}); err != nil {
			t.Fatalf("kicked=%v: cast: %v", kicked, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		wantP := 2
		if kicked {
			wantP = 3
		}
		if p, tt := effectstest.PtOf(t, g, goblin); p != wantP || tt != wantP {
			t.Errorf("kicked=%v: Goblin is %d/%d, want %d/%d", kicked, p, tt, wantP, wantP)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, goblin, "menace"); got != kicked {
			t.Errorf("kicked=%v: menace = %v", kicked, got)
		}
		if kicked {
			effectstest.EndTurn(t, g)
			if p, _ := effectstest.PtOf(t, g, goblin); p != 2 || effectstest.HasEffectiveKeyword(t, g, goblin, "menace") {
				t.Errorf("the pump outlasted the turn: power %d", p)
			}
		}
	}
}
