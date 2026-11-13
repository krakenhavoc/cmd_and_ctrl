package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// apc_b2_cards_test.go — Apocalypse, batch 2 (#5614). Each test plays the
// real catalog entry and asserts what its text says.

const (
	apcB2SkyfolkOracle     = "dd0da475-55d0-462e-bfac-6eeb5d39a099"
	apcB2BarrierOracle     = "dcf65293-14a0-46a5-a0b0-ca2b7fda8bef"
	apcB2ElfOracle         = "9c382817-5c62-44ff-8f29-d283e65712a1"
	apcB2SpringOracle      = "1d7a4e9c-b3c9-421e-b5f4-82d51d7d5125"
	apcB2EstateOracle      = "4d52c4a5-e5c8-4fb4-be50-78d5482dd1ae"
	apcB2LegionnaireOracle = "582fdfbe-c5fe-45d0-b3ff-e66db5183725"
	apcB2TombOracle        = "a0baa9eb-0b1e-4d90-a7a1-f7102f575467"
	apcB2VerdictOracle     = "670b5dfb-ffdf-43ed-acc3-855006e352fc"
	apcB2WarriorOracle     = "6ec24c24-7e87-4c12-b94e-3cd5b184beb5"
	apcB2MongerOracle      = "3913c7f2-86b7-4dd4-84a6-9bd4c1c66781"
	apcB2DeathMutOracle    = "6f705338-b00b-4a6d-924f-443e7c5670b7"
	apcB2AetherMutOracle   = "6697fe5b-90ac-4321-aa2f-cdc6ec283cb4"
	apcB2ShamblerOracle    = "9991941e-436f-42d4-8b0c-1ee84234774b"
	apcB2MinefieldOracle   = "3a8a67ad-e4ff-4b55-9976-a2f13cc9b41b"
	apcB2LifeDeathOracle   = "d476f4b3-db63-4756-ab76-4b35f63c2825"
	apcB2OrderChaosOracle  = "bc540259-459d-4561-9268-c4f323ea62d2"
)

func TestApcB2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		apcB2SkyfolkOracle:           "Gaea's Skyfolk",
		apcB2BarrierOracle:           "Jungle Barrier",
		apcB2ElfOracle:               "Urborg Elf",
		apcB2SpringOracle:            "Temporal Spring",
		apcB2EstateOracle:            "Overgrown Estate",
		apcB2LegionnaireOracle:       "Goblin Legionnaire",
		apcB2TombOracle:              "Martyrs' Tomb",
		apcB2VerdictOracle:           "Gerrard's Verdict",
		apcB2WarriorOracle:           "Putrid Warrior",
		apcB2MongerOracle:            "Spiritmonger",
		apcB2DeathMutOracle:          "Death Mutation",
		apcB2AetherMutOracle:         "Aether Mutation",
		apcB2ShamblerOracle:          "Fungal Shambler",
		apcB2MinefieldOracle:         "Powerstone Minefield",
		apcB2LifeDeathOracle:         "Life",
		apcB2LifeDeathOracle + "#1":  "Death",
		apcB2OrderChaosOracle:        "Order",
		apcB2OrderChaosOracle + "#1": "Chaos",
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

// apcB2Saprolings counts the Saproling tokens a player controls.
func apcB2Saprolings(g *game.Game, controller uuid.UUID) int {
	n, _ := countTokensNamed(g, controller, "Saproling")
	return n
}

// --- Gaea's Skyfolk ----------------------------------------------------------------------

func TestApcB2GaeasSkyfolkHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	folk := apcB1Push(g, g.Seats[0].ID, "Gaea's Skyfolk", "Creature — Elf Merfolk", apcB2SkyfolkOracle, "{G}{U}", 2, 2)
	if !hasKeywordOnBattlefield(t, g, folk, "flying") {
		t.Error("Gaea's Skyfolk does not have flying")
	}
}

// --- Jungle Barrier ----------------------------------------------------------------------

func TestApcB2JungleBarrierIsADefenderThatDrawsAsItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "Top")
	barrier := castCatalogSpell(t, g, "Jungle Barrier", "Creature — Plant Wall", apcB2BarrierOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, barrier) {
		t.Fatal("Jungle Barrier is not on the battlefield")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards as it entered, want 1", got)
	}
	if !hasKeywordOnBattlefield(t, g, barrier, "defender") {
		t.Error("Jungle Barrier does not have defender")
	}
}

// --- Urborg Elf --------------------------------------------------------------------------

func TestApcB2UrborgElfTapsForBlackGreenOrBlue(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	refused := apcB1Push(g, me.ID, "Urborg Elf", "Creature — Elf Druid", apcB2ElfOracle, "{1}{G}", 1, 1)
	if err := g.ActivateManaAbility(me.ID, refused, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err == nil {
		t.Error("Urborg Elf made {R}; it adds {B}, {G}, or {U}")
	}
	for _, color := range []string{"B", "G", "U"} {
		elf := apcB1Push(g, me.ID, "Urborg Elf", "Creature — Elf Druid", apcB2ElfOracle, "{1}{G}", 1, 1)
		if err := g.ActivateManaAbility(me.ID, elf, 0, game.ManaAbilityParams{Colors: []string{color}}); err != nil {
			t.Fatalf("add {%s}: %v", color, err)
		}
		if got := me.ManaPool.CountColor(color); got != 1 {
			t.Errorf("pool holds %d {%s}, want 1", got, color)
		}
		if !findCardAnywhere(t, g, elf).Tapped {
			t.Error("Urborg Elf did not tap")
		}
	}
}

// --- Temporal Spring ---------------------------------------------------------------------

func TestApcB2TemporalSpringPutsAPermanentOnTopOfItsOwnersLibrary(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	seedLibrary(opp, "Under")
	relic := apcB1Push(g, opp.ID, "Relic", "Artifact", "", "{3}", 0, 0)
	castCatalogSpell(t, g, "Temporal Spring", "Sorcery", apcB2SpringOracle, cardTarget(relic))
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, relic) {
		t.Fatal("the target is still on the battlefield")
	}
	if top := libraryTopIDs(opp, 1); top[0] != relic {
		t.Error("the target is not on top of its owner's library")
	}
}

// --- Overgrown Estate --------------------------------------------------------------------

func TestApcB2OvergrownEstateSacrificesALandForThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	estate := apcB1Push(g, me.ID, "Overgrown Estate", "Enchantment", apcB2EstateOracle, "{W}{B}{G}", 0, 0)
	land := apcB1Push(g, me.ID, "My Land", "Land", "", "", 0, 0)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, estate, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{land}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if battlefieldHas(g, land) {
		t.Fatal("the land was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := me.Life - before; got != 3 {
		t.Errorf("gained %d life, want 3", got)
	}
}

// --- Goblin Legionnaire ------------------------------------------------------------------

func TestApcB2GoblinLegionnaireSacrificesToDealTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := apcB1Push(g, me.ID, "Goblin Legionnaire", "Creature — Goblin Soldier", apcB2LegionnaireOracle, "{R}{W}", 2, 2)
	dcMana(t, g, me.ID, "{R}")
	before := opp.Life
	if err := invB1Activate(g, me.ID, gob, 0, playerTarget(opp.ID)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if battlefieldHas(g, gob) {
		t.Fatal("Goblin Legionnaire was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the target lost %d life, want 2", got)
	}
}

func TestApcB2GoblinLegionnaireSacrificesToPreventTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gob := apcB1Push(g, me.ID, "Goblin Legionnaire", "Creature — Goblin Soldier", apcB2LegionnaireOracle, "{R}{W}", 2, 2)
	bear := seedBear(g, me.ID)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	dcMana(t, g, me.ID, "{W}")
	if err := invB1Activate(g, me.ID, gob, 1, cardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, ogre, bear, 3)
	if !battlefieldHas(g, bear) {
		t.Fatal("the shielded 2/2 died to 3 damage; 2 of it should be prevented")
	}
	if got := e2Card(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the target has %d damage marked, want 1", got)
	}
}

// --- Martyrs' Tomb -----------------------------------------------------------------------

func TestApcB2MartyrsTombPaysTwoLifeForAOnePointShield(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tomb := apcB1Push(g, me.ID, "Martyrs' Tomb", "Enchantment", apcB2TombOracle, "{2}{W}{B}", 0, 0)
	bear := seedBear(g, me.ID)
	ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	before := me.Life
	if err := invB1Activate(g, me.ID, tomb, 0, cardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := before - me.Life; got != 2 {
		t.Errorf("paid %d life, want 2", got)
	}
	passPriorityAroundTable(t, g)
	pr6Damage(t, g, ogre, bear, 2)
	if !battlefieldHas(g, bear) {
		t.Fatal("the 1-point shield did not save the 2/2 from 2 damage")
	}
	if got := e2Card(t, g, bear).DamageMarked; got != 1 {
		t.Errorf("the creature has %d damage marked, want 1", got)
	}
}

func TestApcB2MartyrsTombCantPayMoreLifeThanYouHave(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	tomb := apcB1Push(g, me.ID, "Martyrs' Tomb", "Enchantment", apcB2TombOracle, "{2}{W}{B}", 0, 0)
	bear := seedBear(g, me.ID)
	g.WithWriteLock(func() { me.Life = 1 })
	if err := invB1Activate(g, me.ID, tomb, 0, cardTarget(bear)...); err == nil {
		t.Error("activated with 1 life; the cost is 2 life (CR 119.4)")
	}
}

// --- Gerrard's Verdict -------------------------------------------------------------------

func TestApcB2GerrardsVerdictDiscardsTwoAndGainsThreePerLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	land := handCardForTest(opp, "Forest", "Basic Land — Forest", "")
	bear := handCardForTest(opp, "Bear", "Creature — Bear", "")
	keep := handCardForTest(opp, "Keeper", "Sorcery", "")
	before := me.Life
	castCatalogSpell(t, g, "Gerrard's Verdict", "Sorcery", apcB2VerdictOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, land, bear)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(land) || !opp.Graveyard.Contains(bear) || !opp.Hand.Contains(keep) {
		t.Fatal("the target did not discard the two cards it chose")
	}
	if got := me.Life - before; got != 3 {
		t.Errorf("gained %d life for one land discarded, want 3", got)
	}
}

func TestApcB2GerrardsVerdictGainsNothingForNonlands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	a := handCardForTest(opp, "Bear", "Creature — Bear", "")
	b := handCardForTest(opp, "Spell", "Sorcery", "")
	before := me.Life
	castCatalogSpell(t, g, "Gerrard's Verdict", "Sorcery", apcB2VerdictOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, opp.ID) != nil {
		answerDiscard(t, g, opp.ID, a, b)
	}
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Errorf("life %d → %d with no land discarded", before, me.Life)
	}
}

// --- Putrid Warrior ----------------------------------------------------------------------

func TestApcB2PutridWarriorMakesEachPlayerLoseOrGainOne(t *testing.T) {
	for _, tc := range []struct {
		mode  int
		delta int
	}{{0, -1}, {1, 1}} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		warrior := b12Push(g, me.ID, "Putrid Warrior", "Creature — Zombie Soldier Warrior", apcB2WarriorOracle, 2, 2)
		wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 10)
		before := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			before[p.ID] = p.Life
		}
		pr6Damage(t, g, warrior, wall, 2)
		c := modePickChoiceFor(g, me.ID)
		if c == nil {
			t.Fatal("no mode choice for the damage it dealt")
		}
		answerMode(t, g, c, me.ID, tc.mode)
		passPriorityAroundTable(t, g)
		for _, p := range g.Seats {
			if got := p.Life - before[p.ID]; got != tc.delta {
				t.Errorf("mode %d: a player's life changed by %d, want %d", tc.mode, got, tc.delta)
			}
		}
	}
}

// --- Spiritmonger ------------------------------------------------------------------------

func TestApcB2SpiritmongerGrowsWhenItDamagesACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	monger := b12Push(g, me.ID, "Spiritmonger", "Creature — Beast", apcB2MongerOracle, 6, 6)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	pr6Damage(t, g, monger, wall, 2)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, monger); p != 7 || tough != 7 {
		t.Errorf("Spiritmonger is %d/%d after damaging a creature, want 7/7", p, tough)
	}
	// Damage to a player is not damage to a creature.
	dealNoncombatDamage(g, me.ID, monger, opp.ID)
	passPriorityAroundTable(t, g)
	if p, _ := ptOf(t, g, monger); p != 7 {
		t.Errorf("Spiritmonger grew to power %d off damage to a player", p)
	}
}

func TestApcB2SpiritmongerBecomesTheChosenColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	monger := b27Push(g, me.ID, "Spiritmonger", "Creature — Beast", apcB2MongerOracle, "{3}{B}{G}", 6, 6, "B", "G")
	dcMana(t, g, me.ID, "{G}")
	if err := invB1Activate(g, me.ID, monger, 1); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "W")
	passPriorityAroundTable(t, g)
	if got := tmtB9Colors(t, g, monger); got != "W" {
		t.Errorf("colors = %q, want only white", got)
	}
}

func TestApcB2SpiritmongerRegenerates(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	monger := apcB1Push(g, me.ID, "Spiritmonger", "Creature — Beast", apcB2MongerOracle, "{3}{B}{G}", 6, 6)
	dcMana(t, g, me.ID, "{B}")
	if err := invB1Activate(g, me.ID, monger, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(monger, game.DestroyOptions{}) })
	if !battlefieldHas(g, monger) {
		t.Fatal("the regeneration shield did not save Spiritmonger")
	}
	if !findCardAnywhere(t, g, monger).Tapped {
		t.Error("a regenerated Spiritmonger is tapped")
	}
}

// --- Death Mutation ----------------------------------------------------------------------

func TestApcB2DeathMutationDestroysAndMakesSaprolingsPerManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := apcB1Push(g, opp.ID, "Beast", "Creature — Beast", "", "{2}{G}{G}", 4, 4)
	black := b27Push(g, opp.ID, "Shade", "Creature — Shade", "", "{B}", 1, 1, "B")
	if err := castCatalogSpellErr(t, g, "Death Mutation", "Sorcery", apcB2DeathMutOracle, cardTarget(black)); err == nil {
		t.Error("a black creature was accepted as the target")
	}
	castCatalogSpell(t, g, "Death Mutation", "Sorcery", apcB2DeathMutOracle, cardTarget(beast))
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, beast) || !opp.Graveyard.Contains(beast) {
		t.Error("the target was not destroyed")
	}
	if got := apcB2Saprolings(g, me.ID); got != 4 {
		t.Errorf("the caster has %d Saprolings, want 4", got)
	}
}

// --- Aether Mutation ---------------------------------------------------------------------

func TestApcB2AetherMutationBouncesAndMakesSaprolingsPerManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ogre := apcB1Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3)
	castCatalogSpell(t, g, "Aether Mutation", "Sorcery", apcB2AetherMutOracle, cardTarget(ogre))
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, ogre) || !opp.Hand.Contains(ogre) {
		t.Error("the target was not returned to its owner's hand")
	}
	if got := apcB2Saprolings(g, me.ID); got != 3 {
		t.Errorf("the caster has %d Saprolings, want 3", got)
	}
	if got := apcB2Saprolings(g, opp.ID); got != 0 {
		t.Errorf("the creature's owner has %d Saprolings, want 0", got)
	}
}

// --- Fungal Shambler ---------------------------------------------------------------------

func TestApcB2FungalShamblerDrawsAndMakesTheOpponentDiscard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Top")
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	a := handCardForTest(opp, "A", "Sorcery", "")
	b := handCardForTest(opp, "B", "Sorcery", "")
	shambler := b12Push(g, me.ID, "Fungal Shambler", "Creature — Fungus Beast", apcB2ShamblerOracle, 6, 4)
	if !hasKeywordOnBattlefield(t, g, shambler, "trample") {
		t.Error("Fungal Shambler does not have trample")
	}
	hand := me.Hand.Size()
	dealCombatDamage(g, me.ID, shambler, opp.ID)
	passPriorityAroundTable(t, g)
	answerDiscard(t, g, opp.ID, b)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	if !opp.Graveyard.Contains(b) || !opp.Hand.Contains(a) {
		t.Error("the opponent did not discard the card it chose")
	}
}

// --- Powerstone Minefield ----------------------------------------------------------------

func TestApcB2PowerstoneMinefieldDamagesAttackersAndBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	apcB1Push(g, them.ID, "Powerstone Minefield", "Enchantment", apcB2MinefieldOracle, "{2}{R}{W}", 0, 0)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Giant", 5, 5)
	home := b12Creature(g, me.ID, "Home", "Creature — Giant", 5, 5)
	blocker := b12Creature(g, them.ID, "Blocker", "Creature — Wall", 0, 10)
	declareAttack(t, g, them.ID, attacker)
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, attacker).DamageMarked; got != 2 {
		t.Errorf("the attacker has %d damage marked, want 2", got)
	}
	if got := e2Card(t, g, home).DamageMarked; got != 0 {
		t.Errorf("a creature that did not attack has %d damage marked", got)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	passPriorityAroundTable(t, g)
	if got := e2Card(t, g, blocker).DamageMarked; got != 2 {
		t.Errorf("the blocker has %d damage marked, want 2", got)
	}
}

// --- Life // Death -----------------------------------------------------------------------

var apcB2LifeDeathFaces = [2]game.Face{
	{Name: "Life", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}},
	{Name: "Death", TypeLine: "Sorcery", ManaCost: "{1}{B}", Colors: []string{"B"}},
}

func TestApcB2LifeMakesYourLandsOneOneCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := apcB1Push(g, me.ID, "My Land", "Land", "", "", 0, 0)
	theirs := apcB1Push(g, opp.ID, "Their Land", "Land", "", "", 0, 0)
	apcB1CastHalf(t, g, apcB2LifeDeathOracle, apcB2LifeDeathFaces, game.CastSpellParams{Face: 0})
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, mine)
	if !c.IsCreature() || !c.IsLand() {
		t.Fatalf("your land is not a land creature: %q", c.TypeLine)
	}
	if p, tough := ptOf(t, g, mine); p != 1 || tough != 1 {
		t.Errorf("your land is %d/%d, want 1/1", p, tough)
	}
	if layeredCard(t, g, theirs).IsCreature() {
		t.Error("another player's land became a creature")
	}
}

func TestApcB2DeathReanimatesAndCostsLifeEqualToManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := pushGraveyardPermanent(me, "Ogre", "Creature — Ogre", "{2}{R}")
	before := me.Life
	apcB1CastHalf(t, g, apcB2LifeDeathOracle, apcB2LifeDeathFaces, game.CastSpellParams{Face: 1, Targets: cardTarget(dead)})
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, dead) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if got := before - me.Life; got != 3 {
		t.Errorf("lost %d life, want 3 (its mana value)", got)
	}
}

// --- Order // Chaos ----------------------------------------------------------------------

var apcB2OrderChaosFaces = [2]game.Face{
	{Name: "Order", TypeLine: "Instant", ManaCost: "{3}{W}", Colors: []string{"W"}},
	{Name: "Chaos", TypeLine: "Instant", ManaCost: "{2}{R}", Colors: []string{"R"}},
}

// apcB2CastHalfFrom casts one half of a split card from `seat`'s hand at
// the current step, without resolving it.
func apcB2CastHalfFrom(t *testing.T, g *game.Game, seat *game.Player, oracle string, faces [2]game.Face, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: seat.ID, Controller: seat.ID,
		Faces: faces[:],
	}
	card.SetFace(0)
	seat.Hand.PushTop(card)
	if err := g.CastSpell(seat.ID, card.InstanceID, params); err != nil {
		t.Fatalf("cast %s: %v", faces[params.Face].Name, err)
	}
	return card.InstanceID
}

func TestApcB2OrderExilesAnAttackingCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Giant", 5, 5)
	idle := b12Creature(g, me.ID, "Idle", "Creature — Giant", 5, 5)
	declareAttack(t, g, them.ID, attacker)
	refused := game.Card{
		InstanceID: uuid.New(), OracleID: apcB2OrderChaosOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: apcB2OrderChaosFaces[:],
	}
	refused.SetFace(0)
	me.Hand.PushTop(refused)
	if err := g.CastSpell(me.ID, refused.InstanceID, game.CastSpellParams{Face: 0, Targets: cardTarget(idle)}); err == nil {
		t.Error("a creature that is not attacking was accepted as Order's target")
	}
	apcB2CastHalfFrom(t, g, me, apcB2OrderChaosOracle, apcB2OrderChaosFaces, game.CastSpellParams{Face: 0, Targets: cardTarget(attacker)})
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, attacker) || !g.Exile.Contains(attacker) {
		t.Error("the attacking creature was not exiled")
	}
}

func TestApcB2ChaosStopsEveryCreatureBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, them := seatsForCombat(g)
	attacker := b12Creature(g, me.ID, "Attacker", "Creature — Giant", 5, 5)
	blocker := b12Creature(g, them.ID, "Blocker", "Creature — Wall", 0, 10)
	advanceToMain(t, g)
	apcB2CastHalfFrom(t, g, me, apcB2OrderChaosOracle, apcB2OrderChaosFaces, game.CastSpellParams{Face: 1})
	passPriorityAroundTable(t, g)
	late := b12Creature(g, them.ID, "Late", "Creature — Wall", 0, 10)
	declareAttack(t, g, them.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, b := range []uuid.UUID{blocker, late} {
		if err := g.DeclareBlocker(b, attacker); err == nil {
			t.Error("a creature blocked after Chaos resolved")
		}
	}
}
