package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr_b19_test.go — Commander deck cards, batches 19 and 20 (#5368).

const (
	cmdrB19WallOfLimbsOracle      = "1aabb6bd-8061-42a5-9a71-f9eac718bd22"
	cmdrB19ElectrodominanceOracle = "4a59d86e-405d-4c70-b2a5-c64f5cecfd9e"
	cmdrB19GadrakOracle           = "42d7ec93-0943-43e2-9b65-b7edc3349410"
	cmdrB19GrenzoOracle           = "80788290-1569-4c7f-b263-b4da7b31a9d1"
	cmdrB19CourserOracle          = "c5abd58f-3e00-447b-a978-64cb88ad501b"
	cmdrB19NogiOracle             = "3d04aa8c-7bff-42c6-bed3-0b18d2cb1529"
	cmdrB19BattledriverOracle     = "c834c4b5-c122-4c63-81c7-884ca4c0dc3b"
	cmdrB19OppDragonOracle        = "46f46bdd-52b8-436f-9a39-f5360a84b2cd"
	cmdrB19SpitFlameOracle        = "f63278e0-3c67-421d-87ce-67725e5e74df"
	cmdrB19StormsWrathOracle      = "eeb6f5eb-0e16-468e-9dfe-6ec8d8172f86"
	cmdrB19HellkiteOracle         = "1c772df4-bd88-4f7e-853c-5dd06f4cf70e"
	cmdrB19VerixOracle            = "d7315b3b-6e5d-4d28-a503-bfae1daf6c43"
	cmdrB19FormationOracle        = "4d6d3e6d-f958-4c17-b117-960fb8033601"
	cmdrB19AxebaneOracle          = "060e0378-e335-4b3d-82aa-ae36639339f6"
	cmdrB19CaryatidOracle         = "bddf5e9a-7c41-45a8-8fa1-18093c408e15"
	cmdrB19SensationOracle        = "98441834-7da0-4d9f-88a4-a342a9b2a4c3"
	cmdrB19HornetNestOracle       = "123d2b6e-114e-4767-860d-50790382ca4a"
	cmdrB19JaddiOracle            = "20faefbd-059c-4aae-81e3-31683bf9f7bf"
)

func TestCmdrB19StormsWrathHitsEveryCreatureAndPlaneswalkerForFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := pushVanillaCreature(g, me.ID, "Big", 5, 5)
	small := pushVanillaCreature(g, opp.ID, "Small", 3, 3)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID,
	})
	lifeBefore := opp.Life
	castCatalogSpell(t, g, "Storm's Wrath", "Sorcery", cmdrB19StormsWrathOracle, nil)
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, big); got != 4 {
		t.Errorf("my 5/5 has %d damage, want 4", got)
	}
	if _, ok := battlefieldCard(g, small); ok {
		t.Error("their 3/3 should have died")
	}
	if got := counterCount(g, walker, game.CounterLoyalty); got != 1 {
		t.Errorf("planeswalker loyalty %d, want 1", got)
	}
	if opp.Life != lifeBefore {
		t.Error("players take no damage")
	}
}

func TestCmdrB19OgreBattledriverBoostsAndHastesOthersEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	driver := pushTestPermanent(g, me.ID, game.Card{Name: "Ogre Battledriver", OracleID: cmdrB19BattledriverOracle,
		TypeLine: "Creature — Ogre Warrior", Power: 3, Toughness: 3})
	bear := cmdrB17Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, bear); p != 4 {
		t.Errorf("entering creature power %d, want 4", p)
	}
	if !hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("entering creature should have haste")
	}
	if p := effectivePower(t, g, driver); p != 3 {
		t.Errorf("the Battledriver itself got a boost: %d", p)
	}
	theirs := cmdrB17Enter(t, g, opp, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature got a boost: %d", p)
	}
	// Until end of turn.
	advanceToDeclareAttackersOf(t, g, 1)
	advanceToDeclareAttackersOf(t, g, 0)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d after the turn, want 2", p)
	}
}

func TestCmdrB19CarvenCaryatidDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	before := me.Hand.Size()
	castCatalogSpell(t, g, "Carven Caryatid", "Creature — Spirit", cmdrB19CaryatidOracle, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d", got, before+1)
	}
}

func TestCmdrB19JaddiOffshootGainsLifeOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushTestPermanent(g, me.ID, game.Card{Name: "Jaddi Offshoot", OracleID: cmdrB19JaddiOracle,
		TypeLine: "Creature — Plant", Power: 0, Toughness: 3})
	before := me.Life
	cmdrB17Enter(t, g, me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d, want %d", me.Life, before+1)
	}
	cmdrB17Enter(t, g, opp, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	passPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("an opponent's land gave life: %d", me.Life)
	}
}

func TestCmdrB19HornetNestMakesAsManyInsectsAsDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	nest := pushTestPermanent(g, me.ID, game.Card{Name: "Hornet Nest", OracleID: cmdrB19HornetNestOracle,
		TypeLine: "Creature — Insect", Power: 0, Toughness: 2})
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), nest, 5); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest() // "dealt damage" triggers wait for the batch to settle
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Insect"); got != 5 {
		t.Fatalf("%d Insects, want 5 (the Nest died to lethal damage and still triggered)", got)
	}
	tok := findBattlefieldByName(g, "Insect")
	if !hasEffectiveKeyword(t, g, tok, "flying") || !hasEffectiveKeyword(t, g, tok, "deathtouch") {
		t.Error("the Insects have flying and deathtouch")
	}
}

func TestCmdrB19AxebaneGuardianTapsForManaPerDefender(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	guardian := pushTestPermanent(g, me.ID, game.Card{Name: "Axebane Guardian", OracleID: cmdrB19AxebaneOracle,
		TypeLine: "Creature — Human Druid", Power: 0, Toughness: 3})
	pushTestPermanent(g, me.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
		Keywords: []string{"defender"}})
	pushTestPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := g.ActivateManaAbility(me.ID, guardian, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("mana ability: %v", err)
	}
	// Two defenders (the Guardian and the Wall): two colour picks.
	if got := b10ResolveAllManaPicks(t, g, me.ID, "G"); got != 2 {
		t.Fatalf("%d colour picks, want 2 (one per creature with defender)", got)
	}
	total := 0
	for _, tok := range me.ManaPool {
		total += tok.N()
	}
	if total != 2 {
		t.Errorf("mana produced %d (%v), want 2", total, me.ManaPool)
	}
}

func TestCmdrB19WallOfLimbsGrowsOnLifeGainAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := pushTestPermanent(g, me.ID, game.Card{Name: "Wall of Limbs", OracleID: cmdrB19WallOfLimbsOracle,
		TypeLine: "Creature — Zombie Wall", Power: 0, Toughness: 3})
	for i := 0; i < 3; i++ {
		g.WithWriteLock(func() {
			if err := g.ChangePlayerLifeForEffect(wall, me.ID, 4); err != nil {
				t.Fatal(err)
			}
		})
		passPriorityAroundTable(t, g)
	}
	if got := counterCount(g, wall, "+1/+1"); got != 3 {
		t.Fatalf("%d counters, want 3 (once per life-gain event, however much)", got)
	}
	before := opp.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"},
		game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("opponent lost %d, want 3 (the Wall's power as it last existed)", got)
	}
}

func TestCmdrB19ElectrodominanceDealsXAndOffersAFreeCastUpToX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cheap := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: cheap, Name: "Cheap", TypeLine: "Sorcery", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	dear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: dear, Name: "Dear", TypeLine: "Sorcery", ManaCost: "{4}", Owner: me.ID, Controller: me.ID})
	before := opp.Life
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Electrodominance", TypeLine: "Instant", OracleID: cmdrB19ElectrodominanceOracle, Owner: me.ID, Controller: me.ID})
	advanceToMainOf(t, g, 0)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		XValue:  3,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 3 {
		t.Errorf("opponent lost %d, want 3", got)
	}
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, cheap)
	for _, c := range offer.DiscardOptions {
		if c == dear {
			t.Error("a spell with mana value 4 was offered at X = 3")
		}
	}
	if err := castThroughResolutionCast(t, g, me.ID, cheap, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
}

func TestCmdrB19SpitFlameDealsFourAndComesBackWhenADragonEnters(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := pushVanillaCreature(g, opp.ID, "Victim", 5, 4)
	flame := castCatalogSpell(t, g, "Spit Flame", "Instant", cmdrB19SpitFlameOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, victim); ok {
		t.Fatal("4 damage should kill a toughness-4 creature")
	}
	if !me.Graveyard.Contains(flame) {
		t.Fatal("Spit Flame is not in the graveyard")
	}
	// An opponent's Dragon does nothing; a non-Dragon of mine does nothing.
	cmdrB17Enter(t, g, opp, game.Card{Name: "Their Dragon", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	cmdrB17Enter(t, g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	passPriorityAroundTable(t, g)
	if triggerOnStack(g, flame) != nil {
		t.Fatal("only a Dragon you control triggers it")
	}
	cmdrB17Enter(t, g, me, game.Card{Name: "My Dragon", TypeLine: "Creature — Dragon", Power: 2, Toughness: 2})
	if triggerOnStack(g, flame) == nil {
		t.Fatal("a Dragon entering under my control should trigger Spit Flame from the graveyard")
	}
	// Without {R} the payment can't be made.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if hasPayUnlessFor(g, me.ID) {
		answerPayUnless(t, g, me.ID, true)
	}
	if !me.Hand.Contains(flame) {
		t.Errorf("Spit Flame should be back in hand after paying {R}")
	}
}

func TestCmdrB19VerixBladewingMakesKaroxOnlyWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Verix Bladewing", "Legendary Creature — Dragon", cmdrB19VerixOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Karox Bladewing"); got != 0 {
		t.Fatalf("unkicked Verix made %d Karox", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	advanceToMain(t, g2)
	id := uuid.New()
	me2.Hand.PushTop(game.Card{InstanceID: id, Name: "Verix Bladewing", TypeLine: "Legendary Creature — Dragon",
		OracleID: cmdrB19VerixOracle, ManaCost: "{2}{R}{R}", Owner: me2.ID, Controller: me2.ID})
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{R}{R}{C}{C}{C}{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.CastSpell(me2.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	passPriorityAroundTable(t, g2)
	karox := findBattlefieldByName(g2, "Karox Bladewing")
	if karox == uuid.Nil {
		t.Fatal("kicked Verix should make Karox Bladewing")
	}
	c := layeredCard(t, g2, karox)
	if !c.IsLegendary() || !c.HasSubtype("Dragon") || !hasEffectiveKeyword(t, g2, karox, "flying") {
		t.Errorf("Karox should be a legendary flying Dragon: %v", c.TypeLine)
	}
	if p, tough := effectivePower(t, g2, karox), effectiveToughness(t, g2, karox); p != 4 || tough != 4 {
		t.Errorf("Karox is %d/%d, want 4/4", p, tough)
	}
}

func TestCmdrB19HellkiteCourserPutsYourCommanderInAndSendsItHome(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmd := uuid.New()
	me.Command.PushTop(game.Card{InstanceID: cmd, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	castsBefore := me.CommanderCasts[cmd]
	castCatalogSpell(t, g, "Hellkite Courser", "Creature — Dragon", cmdrB19CourserOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 {
		t.Fatalf("want an optional pick over the commander, got %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{cmd}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, cmd)
	if !ok {
		t.Fatal("the commander did not enter the battlefield")
	}
	if me.CommanderCasts[cmd] != castsBefore {
		t.Error("putting it onto the battlefield is not casting it: no commander tax")
	}
	if !hasEffectiveKeyword(t, g, cmd, "haste") {
		t.Error("it gains haste")
	}
	_ = c
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, cmd); ok {
		t.Error("the commander should return to the command zone at the beginning of the next end step")
	}
	if !me.Command.Contains(cmd) {
		t.Error("the commander is not in the command zone")
	}
}

func TestCmdrB19HellkiteCourserCanDecline(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cmd := uuid.New()
	me.Command.PushTop(game.Card{InstanceID: cmd, Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: true, Owner: me.ID, Controller: me.ID})
	castCatalogSpell(t, g, "Hellkite Courser", "Creature — Dragon", cmdrB19CourserOracle, nil)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no optional pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, cmd); ok || !me.Command.Contains(cmd) {
		t.Error("declining leaves the commander in the command zone")
	}
}
