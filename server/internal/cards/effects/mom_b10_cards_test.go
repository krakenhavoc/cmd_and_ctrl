package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_b10_cards_test.go — March of the Machine batch 10 (S82, #5477).

const (
	momB10PreeningOracle  = "e1c458ca-10f6-4da8-a93d-3adfb86fe97d"
	momB10ReinforceOracle = "e7fa2fc2-8542-4069-b75f-5520dd30f74d"
	momB10DelugeOracle    = "706833f7-af50-45bf-8f1d-d3a5a31cf2c8"
	momB10BarrageOracle   = "5e414dc2-d0ff-43e5-9aa2-cd9d96de6d89"
	momB10EarthOracle     = "237b8528-a199-45f4-8efc-96d67119f738"
	momB10ReinsOracle     = "b623e61a-cdb9-46b0-a443-547be05db877"
	momB10FireOracle      = "e7f735ce-e53b-40c0-b449-1f28fdfb2484"
	momB10PhoenixOracle   = "f4490b09-dc55-47ed-a573-4ae0b95725b1"
	momB10DepthguardOracl = "dc8c6803-dd87-4f00-8a92-9613dbe90549"
)

func TestMomB10EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		momB10PreeningOracle:              "Preening Champion",
		momB10ReinforceOracle:             "Ral's Reinforcements",
		momB10DelugeOracle:                "Glistening Deluge",
		momB10BarrageOracle:               "Lithomantic Barrage",
		momB10EarthOracle:                 "Vengeant Earth",
		momB10ReinsOracle:                 "Furnace Reins",
		momB10FireOracle:                  "Into the Fire",
		momB10PhoenixOracle:               "Bloodfeather Phoenix",
		momB10DepthguardOracl:             "Karsus Depthguard",
		invasionOfMercadiaOracleID:        "Invasion of Mercadia",
		invasionOfMercadiaOracleID + "#1": "Kyren Flamewright",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// momB10Elementals counts the 1/1 blue and red Elemental tokens a player controls.
func momB10Elementals(g *game.Game, owner uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Elemental" && c.IsToken() && c.Power == 1 && c.Toughness == 1 &&
			len(c.Colors) == 2 {
			n++
		}
	}
	return n
}

func TestMomB10PreeningChampionMakesAnElemental(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Preening Champion", "Creature — Bird Knight", momB10PreeningOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got := momB10Elementals(g, me.ID); got != 1 {
		t.Errorf("Elemental tokens = %d, want 1", got)
	}
}

func TestMomB10RalsReinforcementsMakesTwoElementals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Ral's Reinforcements", "Sorcery", momB10ReinforceOracle, nil)
	passPriorityAroundTable(t, g)
	if got := momB10Elementals(g, me.ID); got != 2 {
		t.Errorf("Elemental tokens = %d, want 2", got)
	}
}

func momB10Colored(g *game.Game, owner uuid.UUID, name string, p, tough int, colors ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestMomB10GlisteningDelugeShrinksEveryoneAndGreenWhiteMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := momB10Colored(g, me.ID, "Red Bear", 4, 4, "R")
	green := momB10Colored(g, opp.ID, "Green Bear", 5, 5, "G")
	gw := momB10Colored(g, opp.ID, "Selesnya Bear", 5, 5, "G", "W")
	white := momB10Colored(g, opp.ID, "White One", 3, 3, "W")
	blue := momB10Colored(g, opp.ID, "Blue Bear", 3, 3, "U")
	castCatalogSpell(t, g, "Glistening Deluge", "Sorcery", momB10DelugeOracle, nil)
	passPriorityAroundTable(t, g)
	for id, want := range map[uuid.UUID]int{red: 3, green: 2, gw: 2, blue: 2} {
		if got := effectivePower(t, g, id); got != want {
			t.Errorf("power = %d, want %d", got, want)
		}
	}
	if g.Battlefield.Contains(white) {
		t.Error("a 3/3 white creature gets -3/-3 and dies")
	}
}

func TestMomB10LithomanticBarrageIsColourConditional(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := momB10Colored(g, opp.ID, "Red Bear", 9, 9, "R")
	blue := momB10Colored(g, opp.ID, "Blue Bear", 9, 9, "U")
	green := momB10Colored(g, opp.ID, "Green Bear", 9, 9, "G")
	castCatalogSpell(t, g, "Lithomantic Barrage", "Sorcery", momB10BarrageOracle, cardRefs(red))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, red); got != 1 {
		t.Errorf("damage to a red creature = %d, want 1", got)
	}
	g.Turn.ActiveSeat = 0
	castCatalogSpell(t, g, "Lithomantic Barrage", "Sorcery", momB10BarrageOracle, cardRefs(blue))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, blue); got != 5 {
		t.Errorf("damage to a blue creature = %d, want 5", got)
	}
	castCatalogSpell(t, g, "Lithomantic Barrage", "Sorcery", momB10BarrageOracle, cardRefs(green))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, green); got != 1 {
		t.Errorf("damage to a green creature = %d, want 1", got)
	}
	_ = me
}

func TestMomB10LithomanticBarrageCannotBeCounteredAndNeedsACreatureOrPlaneswalker(t *testing.T) {
	spec, _ := Lookup(momB10BarrageOracle)
	if !spec.CantBeCountered {
		t.Error("Lithomantic Barrage can't be countered")
	}
	g := newCatalogGame(t)
	land := b12Permanent(g, g.Seats[1].ID, "Island", "Basic Land — Island")
	advanceToMain(t, g)
	id := uuid.New()
	me := g.Seats[g.Turn.ActiveSeat]
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Lithomantic Barrage", TypeLine: "Sorcery", OracleID: momB10BarrageOracle,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(land)}); err == nil {
		t.Error("a land is not a legal target")
	}
}

func TestMomB10VengeantEarthAnimatesALandAndForcesABlock(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Vengeant Earth", "Instant", momB10EarthOracle, cardRefs(land))
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, land)
	if !ok || !c.IsCreature() {
		t.Fatal("the land did not become a creature")
	}
	if got := effectivePower(t, g, land); got != 4 {
		t.Errorf("power = %d, want 4", got)
	}
	if got := effectiveToughness(t, g, land); got != 4 {
		t.Errorf("toughness = %d, want 4", got)
	}
	if !c.IsLand() || !c.HasSubtype("Elemental") {
		t.Errorf("it stays a land and is an Elemental: %q", c.TypeLine)
	}
	if !hasEffectiveKeyword(t, g, land, "haste") {
		t.Error("it has haste")
	}
	_ = opp
}

func TestMomB10VengeantEarthRefusesAnOpponentsPermanent(t *testing.T) {
	g := newCatalogGame(t)
	foe := b12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Vengeant Earth", TypeLine: "Instant", OracleID: momB10EarthOracle,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: cardRefs(foe)}); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
}

func TestMomB10FurnaceReinsStealsAndPaysATreasureOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Furnace Reins", "Sorcery", momB10ReinsOracle, cardRefs(foe))
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, foe)
	if c.Controller != me.ID {
		t.Fatal("control did not change")
	}
	if !hasEffectiveKeyword(t, g, foe, "haste") {
		t.Error("it has haste")
	}
	lciB1Attack(t, g, foe, opp.ID)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	n := 0
	for _, x := range g.Battlefield.Cards {
		if x.Name == "Treasure" && x.Controller == me.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("Treasures = %d, want 1", n)
	}
}

func TestMomB10IntoTheFireDamagesCreaturesPlaneswalkersAndBattles(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 5, 5)
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID,
	})
	battle := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Test Siege", TypeLine: "Battle — Siege",
		Counters: map[string]int{game.CounterDefense: 5}, Owner: opp.ID, Controller: opp.ID,
	})
	castModal(t, g, "Into the Fire", "Sorcery", momB10FireOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, bear); got != 2 {
		t.Errorf("creature damage = %d, want 2", got)
	}
	if c, _ := battlefieldCard(g, pw); c.Counters[game.CounterLoyalty] != 3 {
		t.Errorf("loyalty = %d, want 3", c.Counters[game.CounterLoyalty])
	}
	if c, _ := battlefieldCard(g, battle); c.Counters[game.CounterDefense] != 3 {
		t.Errorf("defense = %d, want 3", c.Counters[game.CounterDefense])
	}
	_ = me
}

func TestMomB10IntoTheFireSecondModeWithNoCardsDrawsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "One", "Two", "Three")
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	castModal(t, g, "Into the Fire", "Sorcery", momB10FireOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 1 {
		t.Errorf("hand = %d, want 1 (zero put back, plus one)", got)
	}
}

func TestMomB10KarsusDepthguardAttacksOnlyAtPowerFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	d := pushCatalogPermanent(g, me.ID, "Karsus Depthguard", "Creature — Lizard Warrior", momB10DepthguardOracl, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == d {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 4, 3
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(d, opp.ID); err == nil {
		t.Fatal("a 4-power defender cannot attack")
	}
	putCounters(t, g, d, game.CounterPlusOne, 1)
	if got := effectivePowerWithCounters(t, g, d); got != 5 {
		t.Fatalf("power = %d, want 5", got)
	}
	if err := g.DeclareAttacker(d, opp.ID); err != nil {
		t.Errorf("a 5-power Depthguard can attack: %v", err)
	}
}

func TestMomB10BloodfeatherPhoenixReturnsOnSpellDamageToAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ph := pushCatalogGraveyardCard(me, "Bloodfeather Phoenix", "Creature — Phoenix", momB10PhoenixOracle, 2, 2)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", acPlayerRef(opp.ID))
	fillPoolColored(me, "R", 1)
	passPriorityAroundTable(t, g)
	pay := latestChoiceOfKind(g, game.PendingChoicePayUnless)
	if pay == nil {
		t.Fatal("the Phoenix did not offer the {R}")
	}
	if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
		t.Fatalf("ResolvePayUnless: %v", err)
	}
	if !g.Battlefield.Contains(ph) {
		t.Fatal("the Phoenix did not return")
	}
	if !hasEffectiveKeyword(t, g, ph, "haste") {
		t.Error("it gains haste")
	}
}

func TestMomB10BloodfeatherPhoenixIgnoresDamageToYourselfAndToCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ph := pushCatalogGraveyardCard(me, "Bloodfeather Phoenix", "Creature — Phoenix", momB10PhoenixOracle, 2, 2)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Bear", 5, 5)
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", cardRefs(foe))
	fillPoolColored(me, "R", 1)
	passPriorityAroundTable(t, g)
	if latestChoiceOfKind(g, game.PendingChoicePayUnless) != nil {
		t.Error("damage to a creature does not offer the payment")
	}
	if g.Battlefield.Contains(ph) {
		t.Error("damage to a creature does not trigger the Phoenix")
	}
}

func TestMomB10InvasionOfMercadiaLootsOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]
	seedLibrary(owner, "One", "Two", "Three")
	siege := importToHand(siegeRow(invasionOfMercadiaOracleID, "Invasion of Mercadia", "Kyren Flamewright",
		"{1}{R}", "4", "Creature — Goblin Spellshaper", "3", "3"), owner)
	toMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, siege, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Siege: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	passPriorityAroundTable(t, g)
	hand := owner.Hand.Size()
	pitch := owner.Hand.Cards[0].InstanceID
	answerDiscard(t, g, owner.ID, pitch)
	passPriorityAroundTable(t, g)
	if got := owner.Hand.Size() - hand; got != 1 {
		t.Errorf("hand changed by %d, want +1 (discard one, draw two)", got)
	}
	if c := battlefieldCardFor(g, siege); c == nil || c.Counters[game.CounterDefense] != 4 {
		t.Error("the Siege should be on the battlefield with 4 defense")
	}
}

func TestMomB10KyrenFlamewrightMakesTokensAndPumpsTheTeam(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	k := pushCatalogPermanent(g, me.ID, "Kyren Flamewright", "Creature — Goblin Spellshaper", invasionOfMercadiaOracleID+"#1", false)
	other := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	b06AddMana(me, "R", "R", "R")
	b16Activate(t, g, me.ID, k, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}})
	passPriorityAroundTable(t, g)
	if got := momB10Elementals(g, me.ID); got != 2 {
		t.Errorf("Elementals = %d, want 2", got)
	}
	if got := effectivePower(t, g, other); got != 3 {
		t.Errorf("Bear power = %d, want 3", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elemental" {
			if got := effectivePower(t, g, c.InstanceID); got != 2 {
				t.Errorf("token power = %d, want 2", got)
			}
			if !hasEffectiveKeyword(t, g, c.InstanceID, "haste") {
				t.Error("the new tokens gain haste too")
			}
		}
	}
}
