package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func wpInGraveyard(p *game.Player, id uuid.UUID) bool { return p.Graveyard.Contains(id) }

// The world rule through a real entry: two world enchantments cast and
// resolved one after the other — the older one goes to its owner's
// graveyard, the newer stays. The three formerly caveated cards are
// Full.
func TestWorldEnchantmentsObeyTheWorldRule(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sphere := effectstest.CastCatalogSpell(t, g, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.WpOnBattlefield(g, sphere) {
		t.Fatal("Gravity Sphere did not resolve onto the battlefield")
	}
	aviary := effectstest.CastCatalogSpell(t, g, "Serra Aviary", "World Enchantment", effectstest.SerraAviaryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.WpOnBattlefield(g, aviary) {
		t.Error("the newer world enchantment left the battlefield")
	}
	if effectstest.WpOnBattlefield(g, sphere) || !wpInGraveyard(me, sphere) {
		t.Error("the older world enchantment is not in its owner's graveyard")
	}
	for _, oracle := range []string{effectstest.ConcordantCrossroadsOrc, effectstest.CavernsOfDespairOracle, effectstest.ForsakenWastesOracle} {
		spec, _ := effects.Lookup(oracle)
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s still carries a caveat: %v", spec.Name, spec.Caveats)
		}
	}
}

func TestLivingPlaneMakesEveryLandA1x1Creature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Living Plane", "World Enchantment", effectstest.LivingPlaneOracle, 0, 0)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	c := effectstest.WpCard(t, g, land)
	if !c.IsCreature() || !c.IsLand() || !c.HasSubtype("Forest") {
		t.Fatalf("the Forest is not a land creature that is still a Forest: %+v", c.Effective())
	}
	if c.PowerForComparison() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the Forest is %d/%d, want 1/1", c.PowerForComparison(), c.CurrentToughness())
	}
}

func TestGravitySphereAndMysticDecreeTakeKeywordsAway(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Merfolk Flier", TypeLine: "Creature — Merfolk",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"flying", "islandwalk", "vigilance"},
	})
	effectstest.B12Push(g, me.ID, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, 0, 0)
	c := effectstest.WpCard(t, g, flier)
	if effectstest.WpHas(g, c, "flying") || !effectstest.WpHas(g, c, "islandwalk") {
		t.Fatalf("under Gravity Sphere: flying=%v islandwalk=%v, want false/true", effectstest.WpHas(g, c, "flying"), effectstest.WpHas(g, c, "islandwalk"))
	}
	effectstest.B12Push(g, me.ID, "Mystic Decree", "World Enchantment", effectstest.MysticDecreeOracle, 0, 0)
	c = effectstest.WpCard(t, g, flier)
	if effectstest.WpHas(g, c, "flying") || effectstest.WpHas(g, c, "islandwalk") || !effectstest.WpHas(g, c, "vigilance") {
		t.Error("under Mystic Decree: flying and islandwalk go, vigilance stays")
	}
}

func TestSerraAviaryPumpsOnlyFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	ground := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B12Push(g, me.ID, "Serra Aviary", "World Enchantment", effectstest.SerraAviaryOracle, 0, 0)
	if p, tt := effectstest.PtOf(t, g, flier); p != 2 || tt != 2 {
		t.Errorf("the opponent's flier is %d/%d, want 2/2", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, ground); p != 2 || tt != 2 {
		t.Errorf("a creature without flying is %d/%d, want 2/2", p, tt)
	}
}

func TestChaosphereFliersBlockOnlyFliersAndGroundCreaturesHaveReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Chaosphere", "World Enchantment", effectstest.ChaosphereOracle, 0, 0)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID, Keywords: []string{"flying"},
	})
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	attackerFlier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hawk", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID, Keywords: []string{"flying"},
	})
	attackerGround := effectstest.B12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	if !effectstest.WpHas(g, effectstest.WpCard(t, g, ground), "reach") || effectstest.WpHas(g, effectstest.WpCard(t, g, flier), "reach") {
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.B12Push(g, me.ID, "Nether Void", "World Enchantment", effectstest.NetherVoidOracle, 0, 0)
		start := opp.Life
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}"); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		if hit := opp.Life == start-3; hit != pay {
			t.Errorf("paid=%v: the Bolt resolved=%v", pay, hit)
		}
	}
}

// In the Eye of Chaos: only instants, and the price is the spell's mana
// value.
func TestInTheEyeOfChaosTaxesInstantsByManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, opp.ID, "In the Eye of Chaos", "World Enchantment", effectstest.InTheEyeOfChaosOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
		OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID,
	})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.AnswerPayUnless(t, g, me.ID, false)
	if c.PayCost != "{1}" {
		t.Errorf("the price is %q, want the Bolt's mana value {1}", c.PayCost)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Error("the unpaid Bolt resolved")
	}
	// A sorcery is not taxed.
	effectstest.CastCatalogSpell(t, g, "Lava Spike", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a sorcery was taxed")
	}
}

// Bazaar of Wonders: entering exiles every graveyard; a spell is
// countered when a card in a graveyard or a nontoken permanent shares
// its name, and not otherwise.
func TestBazaarOfWonders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	old := effectstest.PushGraveyardCardTyped(opp, "Old Card", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Bazaar of Wonders", "World Enchantment", effectstest.BazaarOfWondersOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(old) {
		t.Fatal("the Bazaar's entry did not exile all graveyards")
	}
	start := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Fatalf("a Bolt with no namesake was countered (life %d)", opp.Life)
	}
	// The first Bolt is now in the graveyard: the second is countered.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != start-3 {
		t.Error("a Bolt with a Bolt in a graveyard resolved")
	}
	// A nontoken permanent on the battlefield with the spell's name.
	effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	bears := effectstest.CastCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.WpOnBattlefield(g, bears) {
		t.Error("a Grizzly Bears was cast past a Grizzly Bears on the battlefield")
	}
}

// Eye of Singularity: its entry destroys every non-basic permanent that
// shares a name; a later entry destroys the others with its name.
func TestEyeOfSingularity(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "Twin", "Creature — Test", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Twin", "Creature — Test", 1, 1)
	single := effectstest.B12Creature(g, opp.ID, "Single", "Creature — Test", 1, 1)
	forestA := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	forestB := effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Eye of Singularity", "World Enchantment", effectstest.EyeOfSingularityOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Its own entry triggers both abilities; order them as offered.
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("the Eye's two entry triggers were not ordered")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.WpOnBattlefield(g, a) || effectstest.WpOnBattlefield(g, b) {
		t.Error("two permanents named Twin survived the entry")
	}
	if !effectstest.WpOnBattlefield(g, single) || !effectstest.WpOnBattlefield(g, forestA) || !effectstest.WpOnBattlefield(g, forestB) {
		t.Fatal("a unique permanent or a basic land was destroyed")
	}
	// A new Single enters: the old one is destroyed, the new one stays.
	newSingle := effectstest.CastCatalogSpell(t, g, "Single", "Creature — Test", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.WpOnBattlefield(g, single) {
		t.Error("the older Single survived a namesake's entry")
	}
	if !effectstest.WpOnBattlefield(g, newSingle) {
		t.Error("the entering Single was destroyed by its own trigger")
	}
}

// Storm World: 4 minus the hand size, to the player whose upkeep it is.
func TestStormWorldPunishesASmallHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Storm World", effectstest.StormWorldOracle, "World Enchantment")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	hand := len(opp.Hand.Cards)
	start := opp.Life
	effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tombs := effectstest.PushPermanentForTest(g, me.ID, "Pillar Tombs of Aku", effectstest.PillarTombsOracle, "World Enchantment")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PendingOfKind(g, game.PendingChoicePayUnless)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	start := opp.Life
	if err := g.ResolvePayUnlessWithCards(prompt.ID, opp.ID, true, []uuid.UUID{bear}); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.WpOnBattlefield(g, bear) || opp.Life != start || !effectstest.WpOnBattlefield(g, tombs) {
		t.Fatal("a sacrifice should keep the life and the Tombs")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 2)
	effectstest.PassPriorityAroundTable(t, g)
	third := g.Seats[2]
	start = third.Life
	effectstest.AnswerPayUnless(t, g, third.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if third.Life != start-5 {
		t.Errorf("a decline: life %d, want %d", third.Life, start-5)
	}
	if effectstest.WpOnBattlefield(g, tombs) || !wpInGraveyard(me, tombs) {
		t.Error("a decline should sacrifice the Tombs")
	}
}

// Teferi's Realm: the upkeep player's chosen type phases out, every
// player's nontoken permanents of it.
func TestTeferisRealmPhasesOutTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	realm := effectstest.PushPermanentForTest(g, me.ID, "Teferi's Realm", effectstest.TeferisRealmOracle, "World Enchantment")
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PendingOfKind(g, game.PendingChoiceOptionPick)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the upkeep player was not asked: %+v", prompt)
	}
	if err := g.ResolveOptionPick(prompt.ID, opp.ID, 1); err != nil { // Creature
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.WpOnBattlefield(g, mine) || effectstest.WpOnBattlefield(g, theirs) {
		t.Error("the creatures did not phase out")
	}
	if !effectstest.WpOnBattlefield(g, land) || !effectstest.WpOnBattlefield(g, realm) {
		t.Error("a land or the Realm phased out on a creature choice")
	}
}
