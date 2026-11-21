package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestCryptolithRiteGivesYourCreaturesAnyColourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	theirs := effectstest.GaPut(g, opp.ID, "Bear", "Creature — Bear", "")
	effectstest.GaPut(g, me.ID, "Cryptolith Rite", "Enchantment", effectstest.GaCryptolithRiteOracle)
	if n := effectstest.GrantedManaRowsOf(g, bear); n != 1 {
		t.Fatalf("your bear has %d granted mana rows, want 1", n)
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's bear has %d granted rows, want 0", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, bear, "U"); err != nil {
		t.Fatalf("tap the bear: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "U" {
		t.Errorf("pool %v, want [U]", got)
	}
	if !effectstest.GaLayered(t, g, bear).Tapped {
		t.Error("the granted {T} did not tap the bear")
	}
	// CR 302.6: a creature that entered this turn cannot use it.
	fresh := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Fresh Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fresh {
				g.Battlefield.Cards[i].SummonedThisTurn = true
			}
		}
	})
	if err := effectstest.TapGrantedMana(t, g, me.ID, fresh, "G"); err == nil {
		t.Error("a summoning-sick creature tapped for the Rite's mana")
	}
}

func TestChromaticLanternGivesYourLandsAnyColourMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	lantern := effectstest.GaPut(g, me.ID, "Chromatic Lantern", "Artifact", effectstest.GaChromaticLanternOracle)
	if n := effectstest.GrantedManaRowsOf(g, forest); n != 1 {
		t.Fatalf("the Forest has %d granted rows, want 1", n)
	}
	abs := game.ManaAbilitiesForCard(effectstest.GaLayered(t, g, forest))
	if len(abs) != 2 || abs[0].Produced != "{G}" {
		t.Errorf("Forest mana = %+v, want its own {G} then the granted any colour", abs)
	}
	if n := effectstest.GrantedManaRowsOf(g, lantern); n != 0 {
		t.Errorf("the Lantern is not a land but has %d granted rows", n)
	}
	if err := effectstest.TapGrantedMana(t, g, me.ID, forest, "R"); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("pool %v, want [R]", got)
	}
}

func TestGemhideSliverGrantsEverySliverAndManaweftOnlyYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.GaPut(g, me.ID, "Muscle Sliver", "Creature — Sliver", "")
	theirs := effectstest.GaPut(g, opp.ID, "Muscle Sliver", "Creature — Sliver", "")
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	gemhide := effectstest.GaPut(g, me.ID, "Gemhide Sliver", "Creature — Sliver", effectstest.GaGemhideSliverOracle)
	if effectstest.GrantedManaRowsOf(g, mine) != 1 || effectstest.GrantedManaRowsOf(g, theirs) != 1 || effectstest.GrantedManaRowsOf(g, gemhide) != 1 {
		t.Error("Gemhide gives ALL Slivers — yours, theirs and itself — the ability")
	}
	if effectstest.GrantedManaRowsOf(g, bear) != 0 {
		t.Error("a Bear is not a Sliver")
	}
	// The opponent's Sliver is theirs to tap, not yours (CR 602.2).
	if err := effectstest.TapGrantedMana(t, g, me.ID, theirs, "G"); err == nil {
		t.Error("you tapped an opponent's Sliver for mana")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	mine2 := effectstest.GaPut(g2, me2.ID, "Muscle Sliver", "Creature — Sliver", "")
	theirs2 := effectstest.GaPut(g2, opp2.ID, "Muscle Sliver", "Creature — Sliver", "")
	effectstest.GaPut(g2, me2.ID, "Manaweft Sliver", "Creature — Sliver", effectstest.GaManaweftSliverOracle)
	if effectstest.GrantedManaRowsOf(g2, mine2) != 1 || effectstest.GrantedManaRowsOf(g2, theirs2) != 0 {
		t.Error("Manaweft gives only YOUR Sliver creatures the ability")
	}
}

func TestNecroticSliverGivesEverySliverADestroyAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sliver := effectstest.GaPut(g, me.ID, "Muscle Sliver", "Creature — Sliver", "")
	effectstest.GaPut(g, me.ID, "Necrotic Sliver", "Creature — Sliver", effectstest.GaNecroticSliverOracle)
	target := effectstest.GaPut(g, opp.ID, "Sol Ring", "Artifact", "")
	idx, ref := effectstest.GrantedActivatedIndex(t, g, sliver)
	if idx < 0 {
		t.Fatal("the Sliver has no granted activated ability")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.B16Activate(t, g, me.ID, sliver, idx, game.ActivateAbilityParams{
		Ref:     ref,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	})
	if g.Battlefield.Contains(sliver) {
		t.Error("the Sliver that activated the ability was not sacrificed")
	}
	if g.Battlefield.Contains(target) {
		t.Error("the target was not destroyed")
	}
}

func TestRishkarPutsCountersAndGrantsGreenToCreaturesWithCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.GaPut(g, me.ID, "Bear A", "Creature — Bear", "")
	b := effectstest.GaPut(g, me.ID, "Bear B", "Creature — Bear", "")
	plain := effectstest.GaPut(g, me.ID, "Bear C", "Creature — Bear", "")
	effectstest.CastCatalogSpell(t, g, "Rishkar, Peema Renegade", "Legendary Creature — Elf Druid", effectstest.GaRishkarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.GaLayered(t, g, a).Counters[game.CounterPlusOne] != 1 || effectstest.GaLayered(t, g, b).Counters[game.CounterPlusOne] != 1 {
		t.Fatal("the two targets should each have a +1/+1 counter")
	}
	if effectstest.GrantedManaRowsOf(g, a) != 1 || effectstest.GrantedManaRowsOf(g, b) != 1 {
		t.Error("a creature you control with a counter has {T}: Add {G}")
	}
	if effectstest.GrantedManaRowsOf(g, plain) != 0 {
		t.Error("a creature with no counter does not")
	}
	// Losing its last counter loses the ability.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(a, game.CounterPlusOne, -1) })
	if effectstest.GrantedManaRowsOf(g, a) != 0 {
		t.Error("a creature whose counter came off kept the ability")
	}
}

func TestJaheiraGivesEveryTokenYouControlGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	treasure := effects.TokenCard("1/1 red Goblin")
	treasure.InstanceID, treasure.Owner, treasure.Controller = uuid.New(), me.ID, me.ID
	effectstest.PushBattlefieldCardWithTimestamp(g, treasure)
	nontoken := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	effectstest.GaPut(g, me.ID, "Jaheira, Friend of the Forest", "Legendary Creature — Human Elf Druid", effectstest.GaJaheiraOracle)
	if effectstest.GrantedManaRowsOf(g, treasure.InstanceID) != 1 {
		t.Error("a token you control taps for {G}")
	}
	if effectstest.GrantedManaRowsOf(g, nontoken) != 0 {
		t.Error("a nontoken does not")
	}
	if spec, _ := effects.Lookup(effectstest.GaJaheiraOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Choose a Background is supported since #2874: Jaheira is Full")
	}
}

func TestInsidiousRootsGrantsAndGrowsPlants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	token := effects.TokenCard("0/1 green Plant")
	token.InstanceID, token.Owner, token.Controller = uuid.New(), me.ID, me.ID
	effectstest.PushBattlefieldCardWithTimestamp(g, token)
	effectstest.GaPut(g, me.ID, "Insidious Roots", "Enchantment", effectstest.GaInsidiousRootsOracle)
	if effectstest.GrantedManaRowsOf(g, token.InstanceID) != 1 {
		t.Fatal("a creature token you control taps for any colour")
	}
	// A creature card leaves your graveyard: one new Plant, and a
	// counter on each Plant — the old one and the new one.
	dead := game.NewCard("Dead Bear", me.ID)
	dead.TypeLine = "Creature — Bear"
	me.Graveyard.PushTop(dead)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, dead.InstanceID); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.B16CountNamed(g, "Plant"); n != 2 {
		t.Fatalf("%d Plants, want 2", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plant" && c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("a Plant has %d +1/+1 counters, want 1", c.Counters[game.CounterPlusOne])
		}
	}
}

func TestGreatDivideGuideGrantsLandsAndAllies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	ally := effectstest.GaPut(g, me.ID, "Ally", "Creature — Human Ally", "")
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	guide := effectstest.GaPut(g, me.ID, "Great Divide Guide", "Creature — Human Scout Ally", effectstest.GaGreatDivideGuideOracle)
	if effectstest.GrantedManaRowsOf(g, forest) != 1 || effectstest.GrantedManaRowsOf(g, ally) != 1 || effectstest.GrantedManaRowsOf(g, guide) != 1 {
		t.Error("each land and Ally you control, the Guide included, has the ability")
	}
	if effectstest.GrantedManaRowsOf(g, bear) != 0 {
		t.Error("a non-Ally creature does not")
	}
}

func TestTheWorldTreeFixesAtSixLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tree := effectstest.GaPut(g, me.ID, "The World Tree", "Land", effectstest.GaTheWorldTreeOracle)
	var forests []uuid.UUID
	for i := 0; i < 4; i++ {
		forests = append(forests, effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", ""))
	}
	if effectstest.GrantedManaRowsOf(g, forests[0]) != 0 || effectstest.GrantedManaRowsOf(g, tree) != 0 {
		t.Fatal("five lands: no grant")
	}
	effectstest.GaPut(g, me.ID, "Forest", "Basic Land — Forest", "")
	if effectstest.GrantedManaRowsOf(g, forests[0]) != 1 || effectstest.GrantedManaRowsOf(g, tree) != 1 {
		t.Error("six lands: every land you control, the Tree included, taps for any colour")
	}
	own := game.ManaAbilitiesForCard(effectstest.GaLayered(t, g, tree))
	if len(own) != 2 || own[0].Produced != "{G}" {
		t.Errorf("the Tree's mana = %+v, want its own {G} then the grant", own)
	}
}

func TestParadiseMantleGrantFollowsTheEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.GaPut(g, me.ID, "Bear", "Creature — Bear", "")
	other := effectstest.GaPut(g, me.ID, "Other Bear", "Creature — Bear", "")
	mantle := effectstest.GaPut(g, me.ID, "Paradise Mantle", "Artifact — Equipment", effectstest.GaParadiseMantleOracle)
	if effectstest.GrantedManaRowsOf(g, bear) != 0 {
		t.Fatal("an unattached Mantle grants nothing")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, mantle, bear)
	if effectstest.GrantedManaRowsOf(g, bear) != 1 || effectstest.GrantedManaRowsOf(g, other) != 0 {
		t.Fatal("the equipped creature has the ability, the other does not")
	}
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.EquipTo(t, g, me.ID, mantle, other)
	if effectstest.GrantedManaRowsOf(g, bear) != 0 || effectstest.GrantedManaRowsOf(g, other) != 1 {
		t.Error("the ability moves with the Mantle")
	}
}

func TestSquirrelNestLetsTheLandsControllerMakeSquirrels(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The Nest on an OPPONENT's land feeds the opponent.
	theirLand := effectstest.GaPut(g, opp.ID, "Forest", "Basic Land — Forest", "")
	effectstest.PushAuraOnLand(t, g, me.ID, "Squirrel Nest", effectstest.GaSquirrelNestOracle, theirLand)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, theirLand)
	if idx < 0 {
		t.Fatal("the enchanted land has no granted ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, theirLand, idx, game.ActivateAbilityParams{Ref: ref}); err == nil {
		t.Error("the Aura's controller activated an ability of a land they do not control")
	}
	g.Turn.PriorityHolder = 1
	g.Turn.ActiveSeat = 1
	if err := g.ActivateCatalogAbility(opp.ID, theirLand, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("the land's controller activates it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	squirrels := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Squirrel" {
			squirrels++
			if c.Controller != opp.ID {
				t.Error("the Squirrel belongs to the land's controller")
			}
		}
	}
	if squirrels != 1 {
		t.Errorf("%d Squirrels, want 1", squirrels)
	}
}
