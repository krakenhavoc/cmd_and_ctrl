package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// world_permanents_test.go — ADR 0109 §8 (#1862): the world enchantments
// the world rule (CR 704.5k) unblocked, and the rule itself through a
// real entry.

const (
	livingPlaneOracle       = "30e99293-3212-4e3f-b543-b1c7c416575d"
	pillarTombsOracle       = "b8cdf79f-b247-4346-8122-9a2d7d23b3c3"
	mysticDecreeOracle      = "06eb8b90-aad8-41c1-bbc5-85ded2fe76bc"
	netherVoidOracle        = "53bab610-1d27-4897-a4ad-cfecca82b811"
	serraAviaryOracle       = "7a06483b-e71c-4d11-86ea-3b48e2a9cdaf"
	inTheEyeOfChaosOracle   = "6a0da9f3-cb06-42b1-ae73-b142c0eedaff"
	bazaarOfWondersOracle   = "7fca65b8-01fe-4858-97c6-f97ae82cd801"
	gravitySphereOracle     = "8ddf93fe-980b-4dc4-b56f-6a2ee50100a6"
	teferisRealmOracle      = "9851d934-2e07-49c8-b08b-15f96d0f3f0c"
	stormWorldOracle        = "868f4ab2-a846-4ad0-8720-95fd234dd36b"
	chaosphereOracle        = "ed01d5d7-8f34-47b3-9ca8-d82c242d38b4"
	eyeOfSingularityOracle  = "e13edf92-cdba-4c24-b81b-088b1554fde6"
	concordantCrossroadsOrc = "ff01b408-6d17-40a3-9efd-a1b341ec1307"
)

// wpCard is a battlefield card as the layers leave it now.
func wpCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	var out game.Card
	found := false
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				out, found = c, true
			}
		}
	})
	if !found {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return out
}

func wpHas(g *game.Game, c game.Card, kw string) bool {
	return HasKeyword(kw)(g, uuid.Nil, c)
}

func wpOnBattlefield(g *game.Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func wpInGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

// The world rule through a real entry: two world enchantments cast and
// resolved one after the other — the older one goes to its owner's
// graveyard, the newer stays. The three formerly caveated cards are
// Full.
func TestWorldEnchantmentsObeyTheWorldRule(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphere := castCatalogSpell(t, g, "Gravity Sphere", "World Enchantment", gravitySphereOracle, nil)
	passPriorityAroundTable(t, g)
	if !wpOnBattlefield(g, sphere) {
		t.Fatal("Gravity Sphere did not resolve onto the battlefield")
	}
	aviary := castCatalogSpell(t, g, "Serra Aviary", "World Enchantment", serraAviaryOracle, nil)
	passPriorityAroundTable(t, g)
	if !wpOnBattlefield(g, aviary) {
		t.Error("the newer world enchantment left the battlefield")
	}
	if wpOnBattlefield(g, sphere) || !wpInGraveyard(me, sphere) {
		t.Error("the older world enchantment is not in its owner's graveyard")
	}
	for _, oracle := range []string{concordantCrossroadsOrc, cavernsOfDespairOracle, forsakenWastesOracle} {
		spec, _ := Lookup(oracle)
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s still carries a caveat: %v", spec.Name, spec.Caveats)
		}
	}
}

func TestLivingPlaneMakesEveryLandA1x1Creature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Living Plane", "World Enchantment", livingPlaneOracle, 0, 0)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	c := wpCard(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Forest") {
		t.Fatalf("the Forest is not a land creature that is still a Forest: %+v", c.Effective())
	}
	if c.PowerForComparison() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the Forest is %d/%d, want 1/1", c.PowerForComparison(), c.CurrentToughness())
	}
}

func TestGravitySphereAndMysticDecreeTakeKeywordsAway(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk Flier", TypeLine: "Creature — Merfolk",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"flying", "islandwalk", "vigilance"},
	})
	b12Push(g, me.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, 0, 0)
	c := wpCard(t, g, flier)
	if wpHas(g, c, "flying") || !wpHas(g, c, "islandwalk") {
		t.Fatalf("under Gravity Sphere: flying=%v islandwalk=%v, want false/true", wpHas(g, c, "flying"), wpHas(g, c, "islandwalk"))
	}
	b12Push(g, me.ID, "Mystic Decree", "World Enchantment", mysticDecreeOracle, 0, 0)
	c = wpCard(t, g, flier)
	if wpHas(g, c, "flying") || wpHas(g, c, "islandwalk") || !wpHas(g, c, "vigilance") {
		t.Error("under Mystic Decree: flying and islandwalk go, vigilance stays")
	}
}

func TestSerraAviaryPumpsOnlyFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	ground := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b12Push(g, me.ID, "Serra Aviary", "World Enchantment", serraAviaryOracle, 0, 0)
	if p, tt := ptOf(t, g, flier); p != 2 || tt != 2 {
		t.Errorf("the opponent's flier is %d/%d, want 2/2", p, tt)
	}
	if p, tt := ptOf(t, g, ground); p != 2 || tt != 2 {
		t.Errorf("a creature without flying is %d/%d, want 2/2", p, tt)
	}
}

func TestChaosphereFliersBlockOnlyFliersAndGroundCreaturesHaveReach(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Chaosphere", "World Enchantment", chaosphereOracle, 0, 0)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	ground := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	attackerFlier := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"},
	})
	attackerGround := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	if !wpHas(g, wpCard(t, g, ground), "reach") || wpHas(g, wpCard(t, g, flier), "reach") {
		t.Fatal("creatures without flying have reach; creatures with flying do not get it")
	}
	refused := func(attacker, blocker uuid.UUID) bool {
		var r game.BlockRefusal
		g.WithWriteLock(func() {
			g.RecomputeLayersIfStaleLocked()
			var a, b *game.Card
			for i := range g.Battlefield.Cards {
				switch g.Battlefield.Cards[i].InstanceID {
				case attacker:
					a = &g.Battlefield.Cards[i]
				case blocker:
					b = &g.Battlefield.Cards[i]
				}
			}
			r = g.BlockPairRefusalLocked(a, b)
		})
		return r.Reason != ""
	}
	if !refused(attackerGround, flier) {
		t.Error("a creature with flying blocked a creature without flying")
	}
	if refused(attackerFlier, flier) {
		t.Error("a creature with flying could not block a flier")
	}
	if refused(attackerFlier, ground) {
		t.Error("a ground creature with reach could not block a flier")
	}
	if refused(attackerGround, ground) {
		t.Error("a ground creature could not block a ground creature")
	}
}

// Nether Void: every spell, its controller's too, is countered unless
// its caster pays {3}.
func TestNetherVoidTaxesEverySpell(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		b12Push(g, me.ID, "Nether Void", "World Enchantment", netherVoidOracle, 0, 0)
		start := opp.Life
		castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		passPriorityAroundTable(t, g)
		if pay {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
				t.Fatal(err)
			}
		}
		answerPayUnless(t, g, me.ID, pay)
		passPriorityAroundTable(t, g)
		if hit := opp.Life == start-3; hit != pay {
			t.Errorf("paid=%v: the Bolt resolved=%v", pay, hit)
		}
	}
}

// In the Eye of Chaos: only instants, and the price is the spell's mana
// value.
func TestInTheEyeOfChaosTaxesInstantsByManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	b12Push(g, opp.ID, "In the Eye of Chaos", "World Enchantment", inTheEyeOfChaosOracle, 0, 0)
	advanceToMain(t, g)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := answerPayUnless(t, g, me.ID, false)
	if c.PayCost != "{1}" {
		t.Errorf("the price is %q, want the Bolt's mana value {1}", c.PayCost)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Error("the unpaid Bolt resolved")
	}
	// A sorcery is not taxed.
	castCatalogSpell(t, g, "Lava Spike", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		t.Error("a sorcery was taxed")
	}
}

// Bazaar of Wonders: entering exiles every graveyard; a spell is
// countered when a card in a graveyard or a nontoken permanent shares
// its name, and not otherwise.
func TestBazaarOfWonders(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	old := pushGraveyardCardTyped(opp, "Old Card", "Sorcery")
	castCatalogSpell(t, g, "Bazaar of Wonders", "World Enchantment", bazaarOfWondersOracle, nil)
	passPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(old) {
		t.Fatal("the Bazaar's entry did not exile all graveyards")
	}
	start := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Fatalf("a Bolt with no namesake was countered (life %d)", opp.Life)
	}
	// The first Bolt is now in the graveyard: the second is countered.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Error("a Bolt with a Bolt in a graveyard resolved")
	}
	// A nontoken permanent on the battlefield with the spell's name.
	b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bears := castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if wpOnBattlefield(g, bears) {
		t.Error("a Grizzly Bears was cast past a Grizzly Bears on the battlefield")
	}
}

// Eye of Singularity: its entry destroys every non-basic permanent that
// shares a name; a later entry destroys the others with its name.
func TestEyeOfSingularity(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Twin", "Creature — Test", 1, 1)
	b := b12Creature(g, opp.ID, "Twin", "Creature — Test", 1, 1)
	single := b12Creature(g, opp.ID, "Single", "Creature — Test", 1, 1)
	forestA := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	forestB := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	castCatalogSpell(t, g, "Eye of Singularity", "World Enchantment", eyeOfSingularityOracle, nil)
	passPriorityAroundTable(t, g)
	// Its own entry triggers both abilities; order them as offered.
	if !answerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("the Eye's two entry triggers were not ordered")
	}
	passPriorityAroundTable(t, g)
	if wpOnBattlefield(g, a) || wpOnBattlefield(g, b) {
		t.Error("two permanents named Twin survived the entry")
	}
	if !wpOnBattlefield(g, single) || !wpOnBattlefield(g, forestA) || !wpOnBattlefield(g, forestB) {
		t.Fatal("a unique permanent or a basic land was destroyed")
	}
	// A new Single enters: the old one is destroyed, the new one stays.
	newSingle := castCatalogSpell(t, g, "Single", "Creature — Test", "", nil)
	passPriorityAroundTable(t, g)
	if wpOnBattlefield(g, single) {
		t.Error("the older Single survived a namesake's entry")
	}
	if !wpOnBattlefield(g, newSingle) {
		t.Error("the entering Single was destroyed by its own trigger")
	}
}

// Storm World: 4 minus the hand size, to the player whose upkeep it is.
func TestStormWorldPunishesASmallHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Storm World", stormWorldOracle, "World Enchantment")
	advanceToUpkeepOf(t, g, 1)
	hand := len(opp.Hand.Cards)
	start := opp.Life
	passPriorityAroundTable(t, g)
	want := 4 - hand
	if want < 0 {
		want = 0
	}
	if opp.Life != start-want {
		t.Errorf("hand of %d: life %d, want %d", hand, opp.Life, start-want)
	}
}

// Pillar Tombs of Aku: a player who sacrifices a creature keeps their
// life and the Tombs; one who declines loses 5 and its controller
// sacrifices it.
func TestPillarTombsOfAku(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tombs := pushPermanentForTest(g, me.ID, "Pillar Tombs of Aku", pillarTombsOracle, "World Enchantment")
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt := pendingOfKind(g, game.PendingChoicePayUnless)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	start := opp.Life
	if err := g.ResolvePayUnlessWithCards(prompt.ID, opp.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	passPriorityAroundTable(t, g)
	if wpOnBattlefield(g, bear) || opp.Life != start || !wpOnBattlefield(g, tombs) {
		t.Fatal("a sacrifice should keep the life and the Tombs")
	}
	advanceToUpkeepOf(t, g, 2)
	passPriorityAroundTable(t, g)
	third := g.Seats[2]
	start = third.Life
	answerPayUnless(t, g, third.ID, false)
	passPriorityAroundTable(t, g)
	if third.Life != start-5 {
		t.Errorf("a decline: life %d, want %d", third.Life, start-5)
	}
	if wpOnBattlefield(g, tombs) || !wpInGraveyard(me, tombs) {
		t.Error("a decline should sacrifice the Tombs")
	}
}

// Teferi's Realm: the upkeep player's chosen type phases out, every
// player's nontoken permanents of it.
func TestTeferisRealmPhasesOutTheChosenType(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	realm := pushPermanentForTest(g, me.ID, "Teferi's Realm", teferisRealmOracle, "World Enchantment")
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	advanceToUpkeepOf(t, g, 1)
	passPriorityAroundTable(t, g)
	prompt := pendingOfKind(g, game.PendingChoiceOptionPick)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	if err := g.ResolveOptionPick(prompt.ID, opp.ID, 1); err != nil { // Creature
		t.Fatalf("pick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if wpOnBattlefield(g, mine) || wpOnBattlefield(g, theirs) {
		t.Error("the creatures did not phase out")
	}
	if !wpOnBattlefield(g, land) || !wpOnBattlefield(g, realm) {
		t.Error("a land or the Realm phased out on a creature choice")
	}
}
