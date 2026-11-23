package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// top1000_sweep_b_test.go — the second slice of the missing top-1,000
// cards (#3654): Windborn Muse, Aesi, Thrummingbird, Tezzeret's Gambit,
// Reconnaissance Mission, Vault of the Archangel, Flowering of the White
// Tree, Eerie Interlude, Teleportation Circle and Reconnaissance.

const (
	sweepBWindbornMuseOracle           = "75e09ac6-d789-4c4d-9aca-845d68f3899f"
	sweepBAesiOracle                   = "6511f317-bd38-46d0-b800-7125a3f420da"
	sweepBThrummingbirdOracle          = "eac94269-4baa-4b8e-a0fd-d6b227d1cde3"
	sweepBTezzeretsGambitOracle        = "b33d2b77-f068-4fdc-a1a8-0353a35e3973"
	sweepBReconnaissanceMissionOracle  = "11bf211c-95a3-4e7f-bd08-13f979d54e73"
	sweepBVaultOfTheArchangelOracle    = "eeaac65a-3480-475a-bb28-e6375d53f487"
	sweepBFloweringOfTheWhiteTreeOracl = "3cda958a-f777-433d-850e-a6c8f581db4e"
	sweepBEerieInterludeOracle         = "0634091a-a74c-4cea-b6d1-7324a725554a"
	sweepBTeleportationCircleOracle    = "c786a0aa-d86f-42c3-a7ea-5cb6ab72b5ec"
	sweepBReconnaissanceOracle         = "32599d7c-243b-447b-9d5d-e8e29830ac61"
)

func TestSweepBWindbornMuseTaxesEachAttackerTwo(t *testing.T) {
	g, bear, opp := effectstest.TaxTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Windborn Muse", "Creature — Spirit", sweepBWindbornMuseOracle, false)

	if err := g.DeclareAttacker(bear, opp.ID); !errors.Is(err, game.ErrAttackTaxUnpaid) {
		t.Fatalf("attacking into Windborn Muse with no mana: %v, want ErrAttackTaxUnpaid", err)
	}
	effectstest.FillPool(g.Seats[0], 2)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("attacking into Windborn Muse with {2}: %v", err)
	}
}

func TestSweepBAesiGrantsAnExtraLandDropAndDrawsOnLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	aesi := effectstest.SeedPermanentWithOracle(g, me.ID, "Aesi, Tyrant of Gyre Strait", "Legendary Creature — Serpent", sweepBAesiOracle)

	handBefore := me.Hand.Size()
	if err := effectstest.B02cPlayLandRaw(t, g, "Forest"); err != nil {
		t.Fatalf("the turn's ordinary land drop: %v", err)
	}
	// The land left the hand (net -1); the landfall trigger waits on the
	// stack and asks its "you may" as it resolves.
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, aesi) == nil {
		t.Fatal("the landfall trigger is not on the stack")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d, want %d (the land is pushed and played, one card drawn)", got, handBefore+1)
	}

	if err := effectstest.B02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("Aesi's extra land drop was refused: %v", err)
	}
	// Declining draws nothing.
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand %d after declining, want %d", got, handBefore+1)
	}
	if err := effectstest.B02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn was allowed; Aesi grants one extra")
	}
}

func TestSweepBThrummingbirdProliferatesOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bird := effectstest.B12Push(g, me.ID, "Thrummingbird", "Creature — Phyrexian Bird Horror", sweepBThrummingbirdOracle, 1, 1)
	g.WithWriteLock(func() {
		if opp.Counters == nil {
			opp.Counters = map[string]int{}
		}
		opp.Counters[game.CounterPoison] = 1
	})
	effectstest.AttackWith(t, g, opp.ID, bird)
	if opp.Life != game.StartingLife-1 {
		t.Fatalf("opponent life %d, want %d", opp.Life, game.StartingLife-1)
	}
	effectstest.SettleAnsweringProliferate(t, g)
	if effectstest.PoisonOn(opp) != 2 {
		t.Errorf("poison %d after the proliferate, want 2", effectstest.PoisonOn(opp))
	}
}

func TestSweepBTezzeretsGambitDrawsTwoThenProliferates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ally := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Counter Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{"+1/+1": 1},
	})
	handBefore := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Tezzeret's Gambit", "Sorcery", sweepBTezzeretsGambitOracle, nil)
	effectstest.SettleAnsweringProliferate(t, g)
	if got := me.Hand.Size(); got != handBefore+2 {
		t.Errorf("hand %d, want %d (two cards drawn)", got, handBefore+2)
	}
	if c := effectstest.S58p6View(t, g, ally); c.Counters["+1/+1"] != 2 {
		t.Errorf("+1/+1 counters %d, want 2 after proliferate", c.Counters["+1/+1"])
	}
}

func TestSweepBReconnaissanceMissionDrawsPerCreatureThatConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rm := effectstest.PushPermanentForTest(g, me.ID, "Reconnaissance Mission", sweepBReconnaissanceMissionOracle, "Enchantment")
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	handBefore := me.Hand.Size()

	effectstest.AttackWith(t, g, opp.ID, a, b)
	g.RunStateChecksForTest()
	if effectstest.TriggerOnStack(g, rm) == nil {
		t.Fatal("the triggers are not on the stack")
	}
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBefore; got != 1 {
		t.Errorf("hand delta %d, want 1 (one yes, one no)", got)
	}
}

func TestSweepBReconnaissanceMissionCyclesFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id, _ := effectstest.CycleFromHand(t, g, "Reconnaissance Mission", "Enchantment", sweepBReconnaissanceMissionOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
}

func TestSweepBVaultOfTheArchangelGivesDeathtouchAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vault := effectstest.PushCatalogPermanent(g, me.ID, "Vault of the Archangel", "Land", sweepBVaultOfTheArchangelOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)

	effectstest.B28TapForMana(t, g, me.ID, vault, "")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
		t.Fatalf("tapped for %v, want {C}", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })
	me.ManaPool = nil

	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	for _, c := range []string{"C", "C", "W", "B"} {
		me.ManaPool.AddMana(game.ManaToken{Color: c})
	}
	if err := g.ActivateCatalogAbility(me.ID, vault, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.S58p6View(t, g, bear)
	if !game.HasKeyword(&c, "deathtouch") || !game.HasKeyword(&c, "lifelink") {
		t.Error("your creature should have deathtouch and lifelink")
	}
	o := effectstest.S58p6View(t, g, theirs)
	if game.HasKeyword(&o, "deathtouch") || game.HasKeyword(&o, "lifelink") {
		t.Error("an opponent's creature got the keywords")
	}
}

func TestSweepBFloweringOfTheWhiteTreePumpsByLegendaryStatus(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Flowering of the White Tree", "Legendary Enchantment", sweepBFloweringOfTheWhiteTreeOracl, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Legendary Creature — Human", 2, 2)

	if c := effectstest.S58p6View(t, g, legend); c.CurrentPower() != 4 || c.CurrentToughness() != 3 {
		t.Errorf("legendary creature %d/%d, want 4/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, plain); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("nonlegendary creature %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.S58p6View(t, g, theirs); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("opponent's creature %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestSweepBFloweringOfTheWhiteTreeGivesLegendaryCreaturesWard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Flowering of the White Tree", "Legendary Enchantment", sweepBFloweringOfTheWhiteTreeOracl, 0, 0)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	effectstest.CastAtWardedCreature(t, g, opp, legend)
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Fatal("targeting a legendary creature raised no ward {1} payment for the caster")
	}
}

func TestSweepBEerieInterludeBlinksAnyNumberAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	b := effectstest.PushVanillaCreature(g, me.ID, "Bear B", 2, 2)
	untouched := effectstest.PushVanillaCreature(g, me.ID, "Bear C", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Eerie Interlude", "Instant", sweepBEerieInterludeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, a) || effects.OnBattlefield(g, b) {
		t.Fatal("the targeted creatures are still on the battlefield")
	}
	if !effects.OnBattlefield(g, untouched) {
		t.Error("an untargeted creature was exiled")
	}

	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Bear A", "Bear B"} {
		back := effectstest.FindBattlefieldByName(g, name)
		if back == uuid.Nil {
			t.Errorf("%s did not return at the end step", name)
		}
	}
}

func TestSweepBEerieInterludeWithNoTargetsDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Eerie Interlude", "Instant", sweepBEerieInterludeOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, bear) {
		t.Error("a bear was exiled by a spell with no targets")
	}
}

func TestSweepBTeleportationCircleBlinksAnArtifactOrCreatureAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Teleportation Circle", "Enchantment", sweepBTeleportationCircleOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AdvanceToEndStepOf(t, g, 0)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if !effectstest.HasID(pick.PickTargetCards, bear) {
			t.Fatal("a creature you control is not offered")
		}
		effectstest.PickCard(t, g, me.ID, bear)
	}
	effectstest.PassPriorityAroundTable(t, g)

	back := effectstest.FindBattlefieldByName(g, "Bear")
	if back == uuid.Nil || back == bear {
		t.Fatalf("blinked creature back=%v (old %v): want a new object on the battlefield", back, bear)
	}
}

func TestSweepBReconnaissanceRemovesAnAttackerFromCombatAndUntapsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	recon := effectstest.PushCatalogPermanent(g, me.ID, "Reconnaissance", "Enchantment", sweepBReconnaissanceOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if c, _ := effectstest.BattlefieldCard(g, bear); !c.Tapped {
		t.Fatal("the attacker did not tap")
	}
	if err := g.ActivateCatalogAbility(me.ID, recon, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err != nil {
		t.Fatalf("ActivateCatalogAbility: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	c, _ := effectstest.BattlefieldCard(g, bear)
	if c.AttackingTarget != uuid.Nil {
		t.Error("the creature is still attacking")
	}
	if c.Tapped {
		t.Error("the creature was not untapped")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	if opp.Life != game.StartingLife {
		t.Errorf("opponent life %d, want %d (the removed attacker deals no damage)", opp.Life, game.StartingLife)
	}
}
