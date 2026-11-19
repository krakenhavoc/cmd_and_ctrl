package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	crackdownOracle        = "fa5a9083-e6f7-4f61-9327-a6b96db220c7"
	serenityOracle         = "f4f751c0-02dc-45a2-a4b0-980592509944"
	energyFluxOracle       = "7a756cd1-29a8-4edf-bb74-fbb5b4020022"
	futureSightOracle      = "93a0e938-b916-47ec-9803-369e885e253c"
	wisedraftersWillOracle = "7e09daa4-a0e6-447b-b0c5-6b1bf728525f"
	equilibriumOracle      = "cbc052e7-817f-43b0-b8e3-96b1644d798b"
	lifetapOracle          = "52ac09af-2aa7-4d80-be26-3e6a6efd5c23"
)

func TestCrackdownKeepsBigNonwhiteCreaturesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, them.ID, "Crackdown", "Enchantment", crackdownOracle, false)
	mk := func(name string, power int, colors ...string) uuid.UUID {
		id := uuid.New()
		g.Battlefield.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", Power: power, Toughness: 5,
			Colors: colors, Owner: me.ID, Controller: me.ID, Tapped: true})
		return id
	}
	bigGreen := mk("Big Green", 3, "G")
	bigColorless := mk("Big Colorless", 4)
	bigWhite := mk("Big White", 3, "W")
	smallGreen := mk("Small Green", 2, "G")
	advanceToUpkeepOf(t, g, 0)
	for name, want := range map[string]struct {
		id     uuid.UUID
		tapped bool
	}{
		"nonwhite power 3":  {bigGreen, true},
		"colorless power 4": {bigColorless, true},
		"white power 3":     {bigWhite, false},
		"nonwhite power 2":  {smallGreen, false},
	} {
		if got := untapStaticTapped(t, g, want.id); got != want.tapped {
			t.Errorf("%s: tapped = %v, want %v", name, got, want.tapped)
		}
	}
}

func TestSerenityDestroysEveryArtifactAndEnchantmentOnYourUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	serenity := pushCatalogPermanent(g, me.ID, "Serenity", "Enchantment", serenityOracle, false)
	myRock := pushPermanentForTest(g, me.ID, "My Rock", "", "Artifact")
	theirRock := pushPermanentForTest(g, them.ID, "Their Rock", "", "Artifact")
	theirAura := pushPermanentForTest(g, them.ID, "Their Enchantment", "", "Enchantment")
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	land := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	walkToUpkeepOfSeat0(t, g)
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Serenity": serenity, "my artifact": myRock, "their artifact": theirRock, "their enchantment": theirAura} {
		if onBoard(g, id) {
			t.Errorf("%s survived Serenity", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"a creature": bear, "a land": land} {
		if !onBoard(g, id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestEnergyFluxTaxesEveryArtifactOnItsControllersUpkeep(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, opp.ID, "Energy Flux", "Enchantment", energyFluxOracle, false)
		mine := pushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
		bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
		walkToUpkeepOfSeat0(t, g)
		passPriorityAroundTable(t, g)
		if pay {
			dcMana(t, g, me.ID, "{C}{C}")
		}
		answerPayUnless(t, g, me.ID, pay)
		if got := g.Battlefield.Contains(mine); got != pay {
			t.Errorf("paid %v: my artifact on the battlefield = %v", pay, got)
		}
		if !g.Battlefield.Contains(bear) {
			t.Errorf("paid %v: a nonartifact permanent was taxed", pay)
		}
	}
}

func TestFutureSightRevealsTheTopAndPlaysLandsAndSpellsFromIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Future Sight", TypeLine: "Enchantment",
		OracleID: futureSightOracle, Owner: me.ID, Controller: me.ID,
	})
	land := seedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if vis := g.LibraryTopVisibilityForEffect(me.ID); vis != game.LibraryTopRevealed {
		t.Errorf("top visibility = %v, want revealed to everyone", vis)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land never reached the battlefield")
	}
	spell := seedTopOfLibrary(me, game.Card{Name: "Top Rock", TypeLine: "Artifact", ManaCost: "{1}"})
	dcMana(t, g, me.ID, "{C}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("cast the spell off the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(spell) {
		t.Error("the spell cast from the top did not resolve")
	}
}

func TestFutureSightOnlyOpensTheTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Future Sight", TypeLine: "Enchantment",
		OracleID: futureSightOracle, Owner: me.ID, Controller: me.ID,
	})
	second := seedTopOfLibrary(me, game.Card{Name: "Second Forest", TypeLine: "Basic Land — Forest"})
	seedTopOfLibrary(me, game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"})
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("played the second card of the library")
	}
}

func TestWisedraftersWillRevealsOpponentsHandsAndSacrificesForEffects(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	will := pushCatalogPermanent(g, me.ID, "Wisedrafter's Will", "Enchantment", wisedraftersWillOracle, false)
	if got := hrRevealed(g, me.ID); len(got) != 0 {
		t.Errorf("revealed its controller's hand: %v", got)
	}
	if got := hrRevealed(g, them.ID); len(got) != 1 {
		t.Errorf("opponent's hand revealed by %v, want [Wisedrafter's Will]", got)
	}
	handBefore := len(me.Hand.Cards)
	dcMana(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, will, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("draw ability: %v", err)
	}
	if g.Battlefield.Contains(will) {
		t.Error("Wisedrafter's Will was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - handBefore; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if got := hrRevealed(g, them.ID); len(got) != 0 {
		t.Errorf("the hand is still revealed after the sacrifice: %v", got)
	}
}

func TestWisedraftersWillCountersATargetSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	will := pushCatalogPermanent(g, me.ID, "Wisedrafter's Will", "Enchantment", wisedraftersWillOracle, false)
	advanceToMain(t, g)
	rock := game.NewCard("Test Rock", them.ID)
	rock.TypeLine = "Artifact"
	rock.Keywords = []string{"flash"}
	them.Hand.PushTop(rock)
	if err := g.CastSpell(them.ID, rock.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	dcMana(t, g, me.ID, "{U}{U}")
	if err := g.ActivateCatalogAbility(me.ID, will, 1, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock.InstanceID}},
	}); err != nil {
		t.Fatalf("counter ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock.InstanceID) || !them.Graveyard.Contains(rock.InstanceID) {
		t.Error("the targeted spell was not countered")
	}
}

func TestEquilibriumBouncesATargetCreatureWhenYouPay(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me, them := g.Seats[0], g.Seats[1]
		pushCatalogPermanent(g, me.ID, "Equilibrium", "Enchantment", equilibriumOracle, false)
		victim := pushVanillaCreature(g, them.ID, "Their Bear", 2, 2)
		advanceToMain(t, g)
		bear := game.NewCard("Runeclaw Bear", me.ID)
		bear.TypeLine = "Creature — Bear"
		bear.Power, bear.Toughness = 2, 2
		me.Hand.PushTop(bear)
		if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		answerPickTarget(t, g, victim)
		passPriorityAroundTable(t, g)
		if pay {
			dcMana(t, g, me.ID, "{C}")
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if back := them.Hand.Contains(victim); back != pay {
			t.Errorf("paid %v: their creature returned to hand = %v", pay, back)
		}
		if onBoard(g, victim) == pay {
			t.Errorf("paid %v: their creature on the battlefield = %v", pay, onBoard(g, victim))
		}
	}
}

func TestEquilibriumIgnoresNoncreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Equilibrium", "Enchantment", equilibriumOracle, false)
	pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	advanceToMain(t, g)
	rock := game.NewCard("Test Rock", me.ID)
	rock.TypeLine = "Artifact"
	me.Hand.PushTop(rock)
	if err := g.CastSpell(me.ID, rock.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if len(g.StackMeta) > 0 {
		for _, m := range g.StackMeta {
			if m.Kind == game.StackItemTriggered {
				t.Fatal("Equilibrium triggered off a noncreature spell")
			}
		}
	}
}

func TestLifetapGainsLifeWhenAnOpponentsForestBecomesTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Lifetap", "Enchantment", lifetapOracle, false)
	theirForest := pushPermanentForTest(g, them.ID, "Forest", "", "Basic Land — Forest")
	theirIsland := pushPermanentForTest(g, them.ID, "Island", "", "Basic Land — Island")
	myForest := pushPermanentForTest(g, me.ID, "Forest", "", "Basic Land — Forest")
	tap := func(id uuid.UUID, who uuid.UUID) {
		g.WithWriteLock(func() {
			if err := (TapTarget{Target: id}).Apply(NewContext(g, &game.StackItem{Controller: who})); err != nil {
				t.Fatalf("tap: %v", err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	before := me.Life
	tap(theirIsland, them.ID)
	tap(myForest, me.ID)
	if me.Life != before {
		t.Fatalf("life changed by %d for an Island or my own Forest", me.Life-before)
	}
	tap(theirForest, them.ID)
	if me.Life != before+1 {
		t.Errorf("life = %d, want %d after an opponent's Forest became tapped", me.Life, before+1)
	}
}
