package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b46_cards_test.go — GAP batch 46 (#5948, S82): Modern Horizons 3
// Commander enchantments, instants and lands. Each test plays the real
// catalog entry.

const (
	gapB46CopyLand           = "caa208bd-92e9-40ec-95c0-86e89d4facbb"
	gapB46SufferThePast      = "edf46ef0-159d-4f0b-ae14-14e0c8446382"
	gapB46KozileksReturn     = "6caafe5a-884d-49e8-a62a-7bf609f0ce53"
	gapB46TemptWithMayhem    = "cbf64c58-664f-431f-987e-08ae70d23f2b"
	gapB46CorruptedCrossroad = "6276a985-7630-476d-a94f-6c6adc88f6c4"
	gapB46Glimmerpost        = "92c9aad6-35ec-425d-be7d-393328992820"
	gapB46LazotepQuarry      = "0d2fa39a-9cac-4a1f-bb1e-b6162e6d5169"
	gapB46PlanarNexus        = "26005003-afcb-4c32-a760-be950246ff0f"
	gapB46RuinsOfOranRief    = "7140f396-1bfa-4b28-ba28-fa15eba74652"
	gapB46SpawningBed        = "49e43de3-460b-4562-aef6-da43bd56debc"
	gapB46TendoIceBridge     = "9e006a4b-8dde-4416-8cb4-8401562d0fd5"
	gapB46SunkenPalace       = "c098c507-5154-423a-a70b-f6dfd4959cf6"
)

func TestGapB46CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB46CopyLand:           "Copy Land",
		gapB46SufferThePast:      "Suffer the Past",
		gapB46KozileksReturn:     "Kozilek's Return",
		gapB46TemptWithMayhem:    "Tempt with Mayhem",
		gapB46CorruptedCrossroad: "Corrupted Crossroads",
		gapB46Glimmerpost:        "Glimmerpost",
		gapB46LazotepQuarry:      "Lazotep Quarry",
		gapB46PlanarNexus:        "Planar Nexus",
		gapB46RuinsOfOranRief:    "Ruins of Oran-Rief",
		gapB46SpawningBed:        "Spawning Bed",
		gapB46TendoIceBridge:     "Tendo Ice Bridge",
		gapB46SunkenPalace:       "Sunken Palace",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB46Land(g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	return pushCatalogPermanent(g, owner, name, typeLine, oracle, false)
}

func TestGapB46GlimmerpostGainsOneLifePerLocus(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	m3cb1Push(g, opp.ID, game.Card{Name: "Their Locus", TypeLine: "Land — Locus"})
	m3cb1Push(g, me.ID, game.Card{Name: "Test Locus", TypeLine: "Land — Locus"})
	life := me.Life
	trekB31Enter(t, g, me, game.Card{Name: "Glimmerpost", TypeLine: "Land — Locus", OracleID: gapB46Glimmerpost})
	passPriorityAroundTable(t, g)
	if got := me.Life - life; got != 3 {
		t.Errorf("life gained %d, want 3 (every Locus on the battlefield, the Glimmerpost included)", got)
	}
}

func TestGapB46SpawningBedMakesThreeScions(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bed := gapB46Land(g, me.ID, "Spawning Bed", "Land", gapB46SpawningBed)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, bed, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated the {6} ability for five mana")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, bed, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := len(s46Tokens(g, me.ID, "Eldrazi Scion")); got != 3 {
		t.Errorf("%d Eldrazi Scions, want 3", got)
	}
	if g.Battlefield.Contains(bed) {
		t.Error("the Spawning Bed was not sacrificed")
	}
}

func TestGapB46TendoIceBridgeEntersWithAChargeCounterAndSpendsItForColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bridge := trekB31Enter(t, g, me, game.Card{Name: "Tendo Ice Bridge", TypeLine: "Land", OracleID: gapB46TendoIceBridge})
	if got := counterCount(g, bridge, game.CounterCharge); got != 1 {
		t.Fatalf("%d charge counters on entry, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, bridge, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("the coloured ability: %v", err)
	}
	if got := counterCount(g, bridge, game.CounterCharge); got != 0 {
		t.Errorf("%d charge counters after, want 0", got)
	}
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %v, want one mana", me.ManaPool)
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, bridge).Tapped = false })
	if err := g.ActivateManaAbility(me.ID, bridge, 1, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("the coloured ability worked with no charge counter left")
	}
}

func TestGapB46RuinsOfOranRiefCountersOnlyAColorlessCreatureThatEnteredThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ruins := trekB31Enter(t, g, me, game.Card{Name: "Ruins of Oran-Rief", TypeLine: "Land", OracleID: gapB46RuinsOfOranRief})
	if !layeredCard(t, g, ruins).Tapped {
		t.Fatal("the land did not enter tapped")
	}
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, ruins).Tapped = false })
	colorless := trekB31Enter(t, g, me, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 2, Toughness: 2})
	green := trekB31Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, ruins, 0, game.ActivateAbilityParams{Targets: cardRefs(green), Strict: true}); err == nil {
		t.Fatal("targeted a green creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, ruins, 0, game.ActivateAbilityParams{Targets: cardRefs(colorless), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, colorless); p != 3 {
		t.Errorf("the Golem has power %d, want 3 with a +1/+1 counter", p)
	}
}

func TestGapB46PlanarNexusIsEveryNonbasicLandType(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	nexus := gapB46Land(g, me.ID, "Planar Nexus", "Land", gapB46PlanarNexus)
	c := layeredCard(t, g, nexus)
	for _, want := range []string{"Cave", "Desert", "Gate", "Lair", "Locus", "Mine", "Power-Plant", "Sphere", "Tower", "Urza's"} {
		if !c.HasSubtypeWord(want) {
			t.Errorf("Planar Nexus is not a %s", want)
		}
	}
	for _, not := range []string{"Forest", "Island", "Plains", "Mountain", "Swamp"} {
		if c.HasSubtypeWord(not) {
			t.Errorf("Planar Nexus is a %s", not)
		}
	}
	if err := g.ActivateManaAbility(me.ID, nexus, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("{T}: Add {C}: %v", err)
	}
}

func TestGapB46CorruptedCrossroadsManaPaysOnlyForDevoidSpells(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	crossroads := gapB46Land(g, me.ID, "Corrupted Crossroads", "Land", gapB46CorruptedCrossroad)
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, crossroads, 1, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want a life paid", life, me.Life)
	}
	devoid := game.Card{Name: "Devoid Drone", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{U}", Colors: []string{}, Keywords: []string{game.KeywordDevoid}}
	plain := game.Card{Name: "Plain Drone", TypeLine: "Creature — Eldrazi Drone", ManaCost: "{U}", Colors: []string{"U"}}
	pay := func(c game.Card) bool {
		var ok bool
		g.WithWriteLock(func() {
			cost, err := game.ParseCost("{U}")
			if err != nil {
				t.Fatal(err)
			}
			_, ok = me.ManaPool.SpendManaFor(cost, 0, game.ManaSpendForCast(c))
		})
		return ok
	}
	if pay(plain) {
		t.Fatal("the mana paid for a spell without devoid")
	}
	if !pay(devoid) {
		t.Error("the mana did not pay for a spell with devoid")
	}
}

func TestGapB46SufferThePastExilesXCardsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	var yard []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		yard = append(yard, id)
		g.WithWriteLock(func() {
			opp.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Spent", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
		})
	}
	oppLife, life := opp.Life, me.Life
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Suffer the Past", TypeLine: "Instant", OracleID: gapB46SufferThePast, ManaCost: "{X}{B}", Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{B}{B}{B}")
	targets := append([]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, cardRefs(yard[0], yard[1])...)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets, XValue: 2}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != 1 {
		t.Errorf("their graveyard holds %d cards, want 1", got)
	}
	if opp.Life != oppLife-2 || me.Life != life+2 {
		t.Errorf("life: they %d -> %d, you %d -> %d; want -2 and +2", oppLife, opp.Life, life, me.Life)
	}
}

func TestGapB46KozileksReturnDamagesEachCreatureAndCanBeExiledForMore(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	small := m3cb1Push(g, opp.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	big := m3cb1Push(g, opp.ID, game.Card{Name: "Big", TypeLine: "Creature — Giant", Power: 5, Toughness: 6})
	castCatalogSpell(t, g, "Kozilek's Return", "Instant", gapB46KozileksReturn, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Fatal("the 5/6 died to 2 damage")
	}
	// Now in the graveyard: casting an Eldrazi creature with mana value 7
	// offers the exile.
	if me.Graveyard.Size() == 0 {
		t.Fatal("Kozilek's Return is not in the graveyard")
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Big Eldrazi", TypeLine: "Creature — Eldrazi", ManaCost: "{7}", Power: 7, Toughness: 7, Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the 5/6 survived the 5 damage")
	}
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Kozilek's Return" {
			t.Error("Kozilek's Return is still in the graveyard")
		}
	}
}

func TestGapB46CopyLandEntersAsACopyAndIsAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := m3cb1Push(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	m3cb1Push(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	id := castCatalogSpell(t, g, "Copy Land", "Enchantment", gapB46CopyLand, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if len(pc.CopyOptions) != 1 || !hasID(pc.CopyOptions, forest) {
		t.Errorf("CopyOptions = %v, want only the land", pc.CopyOptions)
	}
	resolveWithCopyChoice(t, g, forest)
	c := layeredCard(t, g, id)
	if c.Name != "Forest" {
		t.Errorf("the permanent is %q, want a copy of the Forest", c.Name)
	}
	if !c.IsLand() || !c.IsEnchantment() || !c.HasSubtypeWord("Forest") {
		t.Errorf("type line %q, want a land that is also an enchantment", c.TypeLine)
	}
}

// gapB46SpellCopies counts the spell copies on the stack.
func gapB46SpellCopies(g *game.Game) int {
	n := 0
	g.WithWriteLock(func() {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemSpell && it.IsCopy {
				n++
			}
		}
	})
	return n
}

func TestGapB46TemptWithMayhemCopiesOncePlusOncePerAcceptingOpponent(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := m3cb1Opponent(g)
	victim := m3cb1Push(g, opp.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Tempt with Mayhem", TypeLine: "Instant", OracleID: gapB46TemptWithMayhem, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Resolve the offer: the first opponent asked accepts, the others
	// decline; every copy keeps the same target.
	accepted := false
	for i := 0; i < 80 && gapB46SpellCopies(g) < 3; i++ {
		if pc := latestChoiceOfKind(g, game.PendingChoicePickTarget); pc != nil {
			if err := g.ResolvePickTarget(pc.ID, pc.Chooser, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if pc := latestChoiceOfKind(g, game.PendingChoiceConfirm); pc != nil {
			yes := !accepted
			accepted = true
			if err := g.ResolveConfirm(pc.ID, pc.Chooser, yes); err != nil {
				t.Fatalf("ResolveConfirm: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// You copy it once, plus once for the opponent who copied it, and that
	// opponent has a copy of their own: three copies in all.
	if got := gapB46SpellCopies(g); got != 3 {
		t.Errorf("%d copies of the spell on the stack, want 3", got)
	}
}

func TestGapB46LazotepQuarryReturnsAZombieCopy(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	quarry := gapB46Land(g, me.ID, "Lazotep Quarry", "Land — Desert", gapB46LazotepQuarry)
	dead := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Knight", TypeLine: "Creature — Human Knight", ManaCost: "{1}{W}", Colors: []string{"W"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, quarry, 0, game.ActivateAbilityParams{
		Targets: cardRefs(dead), XValue: 2, SacrificeIDs: []uuid.UUID{quarry}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	tokens := s46Tokens(g, me.ID, "Fallen Knight")
	if len(tokens) != 1 {
		t.Fatalf("%d Fallen Knight copies, want 1", len(tokens))
	}
	id := tokens[0].InstanceID
	if p, tg := ptOf(t, g, id); p != 4 || tg != 4 {
		t.Errorf("the copy is %d/%d, want 4/4", p, tg)
	}
	c := layeredCard(t, g, id)
	if !c.HasColor("B") || c.HasColor("W") || !c.HasSubtypeWord("Zombie") || c.HasSubtypeWord("Knight") {
		t.Errorf("the copy is %v %q, want a black Zombie", c.EffectiveColors(), c.TypeLine)
	}
}

func TestGapB46SunkenPalaceCopiesTheSpellItsManaPaidFor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	palace := gapB46Land(g, me.ID, "Sunken Palace", "Land — Cave", gapB46SunkenPalace)
	var yard []uuid.UUID
	for i := 0; i < 7; i++ {
		id := uuid.New()
		yard = append(yard, id)
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Spent", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		})
	}
	floatMana(t, g, me, "{C}{U}")
	if err := g.ActivateManaAbility(me.ID, palace, 1, game.ManaAbilityParams{ExileIDs: yard}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := me.Graveyard.Size(); got != 0 {
		t.Errorf("%d cards left in the graveyard, want 0", got)
	}
	// The {U} is spent on Preordain ({U}); the {C}{U} used for the cost is gone.
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Preordain", TypeLine: "Sorcery", ManaCost: "{U}",
			OracleID: "ac641490-ca14-48d7-8cc4-b69ce984befa", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if got := gapB46SpellCopies(g); got != 0 {
		t.Fatalf("%d copies before the trigger resolved, want 0", got)
	}
	// The trigger is above the spell; resolving it copies the spell, and the
	// copy and the spell each scry 2: two prompts, not one.
	scries := 0
	for i := 0; i < 30 && scries < 2; i++ {
		if pc := latestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID); pc != nil {
			scries++
			if err := g.ResolveScry(pc.ID, me.ID, nil, pc.ScryCards); err != nil {
				t.Fatalf("ResolveScry: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if scries != 2 {
		t.Errorf("%d scry prompts, want 2 (the copy and the spell)", scries)
	}
}

func TestGapB46SunkenPalaceNeedsSevenCardsInTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	palace := gapB46Land(g, me.ID, "Sunken Palace", "Land — Cave", gapB46SunkenPalace)
	floatMana(t, g, me, "{C}{U}")
	if err := g.ActivateManaAbility(me.ID, palace, 1, game.ManaAbilityParams{}); err == nil {
		t.Fatal("activated the second ability with an empty graveyard")
	}
}
