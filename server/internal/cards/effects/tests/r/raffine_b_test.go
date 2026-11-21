package testsr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func rbPoolSize(p *game.Player) int { return p.ManaPool.Total() }

func rbUntap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
}

// --- Fetid Pools ----------------------------------------------------

func TestFetidPoolsEntersTappedAndCycles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.PlayLandFromHand(t, g, "Fetid Pools", effectstest.RbFetidPoolsOracle)
	effectstest.Top100AssertEnteredTapped(t, g, id, "Fetid Pools")

	// A second copy cycles from hand for {2}: discard it, draw a card.
	g2 := effectstest.NewCatalogGame(t)
	me := g2.Seats[g2.Turn.ActiveSeat]
	for g2.Turn.Step != game.StepPrecombatMain {
		if _, err := g2.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	cid := effectstest.PushCatalogHandCard(me, "Fetid Pools", "Land — Island Swamp", effectstest.RbFetidPoolsOracle)
	handBefore := len(me.Hand.Cards)
	if err := g2.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.ActivateCatalogAbility(me.ID, cid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if len(me.Hand.Cards) != handBefore {
		t.Errorf("hand %d -> %d, want unchanged (discard one, draw one)", handBefore, len(me.Hand.Cards))
	}
	if !effectstest.InZone(me.Graveyard, cid) {
		t.Error("the cycled land is not in the graveyard")
	}
}

func TestFetidPoolsOffersBlueOrBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.SeedPermanentWithOracle(g, me.ID, "Fetid Pools", "Land — Island Swamp", effectstest.RbFetidPoolsOracle)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.ManaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "R"); err == nil {
		t.Error("red accepted from a U/B land")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "B"); err != nil {
		t.Fatalf("black refused: %v", err)
	}
}

// --- Hall of Heliod's Generosity ---------------------------------------

func TestHallOfHeliodsGenerosityPutsAnEnchantmentOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	hall := effectstest.SeedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", effectstest.RbHallOfHeliodOracle)
	ench := effectstest.PushGraveyardCardTyped(me, "Pacifism", "Enchantment — Aura")
	effectstest.AdvanceToMain(t, g)

	effectstest.DcMana(t, g, me.ID, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{
		Targets: effectstest.B16TargetCard(ench),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InZone(me.Graveyard, ench) {
		t.Error("the enchantment is still in the graveyard")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ench {
		t.Error("the enchantment is not on top of the library")
	}
}

func TestHallOfHeliodsGenerosityRefusesNonEnchantmentAndOpponentsCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hall := effectstest.SeedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", effectstest.RbHallOfHeliodOracle)
	creature := effectstest.PushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	theirs := effectstest.PushGraveyardCardTyped(opp, "Their Ench", "Enchantment")
	effectstest.AdvanceToMain(t, g)

	for _, bad := range []uuid.UUID{creature, theirs} {
		effectstest.DcMana(t, g, me.ID, "{C}{W}")
		err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bad)})
		if err == nil {
			t.Fatal("an illegal target was accepted")
		}
		rbUntap(g, hall)
	}
}

// --- Ondu Inversion ----------------------------------------------------

func TestOnduInversionDestroysEveryNonlandPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.DcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DcSeed(g, opp.ID, "Theirs", "Artifact", "", 0, 0)
	land := effectstest.DcSeed(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)

	effectstest.CastCatalogSpell(t, g, "Ondu Inversion", "Sorcery", effectstest.RbOnduInversionOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, bear) || effectstest.BattlefieldHas(g, theirs) {
		t.Error("a nonland permanent survived Ondu Inversion")
	}
	if !effectstest.BattlefieldHas(g, land) {
		t.Error("a land was destroyed by Ondu Inversion")
	}
	if spec, _ := effects.Lookup(effectstest.RbOnduInversionOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("Ondu Inversion is complete")
	}
}

// --- The Meathook Massacre -----------------------------------------------

func TestMeathookMassacreShrinksEveryCreatureByCastX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.DcSeed(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	big := effectstest.DcSeed(g, opp.ID, "Big", "Creature — Giant", "", 5, 5)

	massacre := effectstest.CastXSpell(t, g, "The Meathook Massacre", "Legendary Enchantment", effectstest.RbMeathookOracle, "{X}{B}{B}", 2, nil)
	effectstest.WaitForXTrigger(t, g, massacre)
	if effectstest.EffectivePower(t, g, big) != 5 {
		t.Fatal("-X/-X applied before the trigger resolved")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if got := effectstest.EffectiveToughness(t, g, big); got != 3 {
		t.Errorf("the 5/5 has toughness %d after X=2, want 3", got)
	}
}

func TestMeathookMassacreDeathPayoffs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DcSeed(g, me.ID, "The Meathook Massacre", "Legendary Enchantment", effectstest.RbMeathookOracle, 0, 0)
	mine := effectstest.DcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.DcSeed(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	meLife, oppLife := me.Life, opp.Life

	for _, id := range []uuid.UUID{mine, theirs} {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Mine died: each opponent loses 1. Theirs died: I gain 1.
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want -1", oppLife, opp.Life)
	}
	if me.Life != meLife+1 {
		t.Errorf("my life %d -> %d, want +1", meLife, me.Life)
	}
}

func TestMeathookMassacreEnteringUncastShrinksNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.DcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	// Pushed straight onto the battlefield (no cast): X is zero.
	effectstest.DcSeed(g, me.ID, "The Meathook Massacre", "Legendary Enchantment", effectstest.RbMeathookOracle, 0, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, bear) {
		t.Error("an uncast Massacre shrank a creature")
	}
}

// --- Silver Scrutiny ---------------------------------------------------

func TestSilverScrutinyDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastXSpell(t, g, "Silver Scrutiny", "Sorcery", effectstest.RbSilverScrutinyOracl, "{X}{U}{U}", 4, nil)
	before := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
	// #3225: the flash clause is the card's per-offer X ceiling now
	// (flash_if_x_at_most_test.go covers it, and its #5207 caveat).
}

// --- Plaza of Heroes -----------------------------------------------------

func TestPlazaOfHeroesColourlessAndLegendaryRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plaza := effectstest.SeedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", effectstest.RbPlazaOracle)

	if err := g.ActivateManaAbility(me.ID, plaza, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if rbPoolSize(me) != 1 || me.ManaPool.Units()[0].Color != "C" {
		t.Fatalf("pool = %v, want one {C}", me.ManaPool)
	}
	me.ManaPool = nil
	rbUntap(g, plaza)

	if err := g.ActivateManaAbility(me.ID, plaza, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	pick := effectstest.ManaPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no colour pick queued")
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "G"); err != nil {
		t.Fatal(err)
	}
	cost, err := game.ParseCost("{G}")
	if err != nil {
		t.Fatal(err)
	}
	ordinary := game.Card{Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid", ManaCost: "{G}"}
	legendary := game.Card{Name: "Yeva, Nature's Herald", TypeLine: "Legendary Creature — Elf Shaman", ManaCost: "{G}"}
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(ordinary)) {
		t.Error("Plaza mana paid for a non-legendary spell — stronger than printed")
	}
	if !me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForCast(legendary)) {
		t.Error("Plaza mana refused a legendary spell")
	}
}

func TestPlazaOfHeroesColorAmongLegendaryPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plaza := effectstest.SeedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", effectstest.RbPlazaOracle)

	// No legendary permanent: nothing.
	if err := g.ActivateManaAbility(me.ID, plaza, 2, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if rbPoolSize(me) != 0 || effectstest.ManaPickFor(g, me.ID) != nil {
		t.Fatal("produced mana with no legendary permanent out")
	}
	rbUntap(g, plaza)

	// A legendary ARTIFACT (not a creature) switches it on — the
	// difference from Mox Amber.
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Legendary Relic", TypeLine: "Legendary Artifact",
		ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID,
	})
	if err := g.ActivateManaAbility(me.ID, plaza, 2, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if rbPoolSize(me) != 1 || me.ManaPool.Units()[0].Color != "U" {
		t.Errorf("pool = %v, want one {U}", me.ManaPool)
	}
}

func TestPlazaOfHeroesExileProtectsALegendaryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plaza := effectstest.SeedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", effectstest.RbPlazaOracle)
	legend := effectstest.DcSeed(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", 2, 2)
	plain := effectstest.DcSeed(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)

	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, plaza, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(plain)}); err == nil {
		t.Fatal("a non-legendary creature was a legal target")
	}
	rbUntap(g, plaza)
	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, plaza, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(legend)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, plaza) {
		t.Error("the land was not exiled as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	abilities := effectstest.EffectiveAbilities(t, g, legend)
	has := map[string]bool{}
	for _, a := range abilities {
		has[a] = true
	}
	if !has["hexproof"] || !has["indestructible"] {
		t.Errorf("legend abilities = %v, want hexproof and indestructible", abilities)
	}
}

// --- Realmbreaker --------------------------------------------------------

func TestRealmbreakerMillsAndStealsALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := effectstest.SeedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", effectstest.RbRealmbreakerOracle)
	land := effectstest.PushGraveyardCardTyped(opp, "Their Mountain", "Basic Land — Mountain")
	effectstest.AdvanceToMain(t, g)
	libBefore := opp.Library.Size()

	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d cards, want 3", got)
	}
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the activator was not asked to choose a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	c, ok := effectstest.BattlefieldCard(g, land)
	if !ok {
		t.Fatal("the land did not reach the battlefield")
	}
	if c.Controller != me.ID {
		t.Error("the land is not under the activator's control")
	}
	if !c.Tapped || effectstest.TapEventsFor(g, land) != 0 {
		t.Error("the land must enter tapped (not be tapped afterwards)")
	}

	// "If this land would leave the battlefield, exile it instead."
	if err := g.DestroyPermanentForEffect(land); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.InZone(opp.Graveyard, land) || effectstest.InZone(me.Graveyard, land) {
		t.Error("the stolen land went to a graveyard instead of exile")
	}
	if !effectstest.InZone(g.Exile, land) {
		t.Error("the stolen land was not exiled")
	}
}

func TestRealmbreakerAsksNothingWithoutALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := effectstest.SeedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", effectstest.RbRealmbreakerOracle)
	effectstest.PushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	// The library is 20 plain filler cards (no lands), so the mill adds none.
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Error("a land prompt was opened with no land card in the graveyard")
	}
}

func TestRealmbreakerSearchesForPraetors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tree := effectstest.SeedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", effectstest.RbRealmbreakerOracle)
	p1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Praetor One", TypeLine: "Legendary Creature — Phyrexian Praetor", Owner: me.ID, Controller: me.ID})
	p2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Praetor Two", TypeLine: "Legendary Creature — Phyrexian Praetor", Owner: me.ID, Controller: me.ID})
	other := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)

	effectstest.DcMana(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if effectstest.BattlefieldHas(g, tree) {
		t.Error("Realmbreaker was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	// Take whatever search prompt exists: both Praetors, never the Bear.
	if pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); pc != nil {
		if err := g.ResolveSearchLibrary(pc.ID, me.ID, []uuid.UUID{p1, p2}); err != nil {
			t.Fatalf("search answer: %v", err)
		}
	}
	if !effectstest.BattlefieldHas(g, p1) || !effectstest.BattlefieldHas(g, p2) {
		t.Error("the Praetors did not reach the battlefield")
	}
	if effectstest.BattlefieldHas(g, other) {
		t.Error("a non-Praetor was fetched")
	}
}

// --- Thespian's Stage -----------------------------------------------------

func TestThespiansStageBecomesACopyAndKeepsItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stage := effectstest.SeedPermanentWithOracle(g, me.ID, "Thespian's Stage", "Land", effectstest.RbThespiansStageOracl)
	other := effectstest.SeedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", effectstest.RbHallOfHeliodOracle)
	effectstest.AdvanceToMain(t, g)

	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, stage)
	if c.Name != "Hall of Heliod's Generosity" {
		t.Fatalf("the Stage is %q after copying, want the Hall", c.Name)
	}
	if !c.Tapped {
		t.Error("the tap was a cost; the copy stays tapped")
	}

	// It can copy again: the ability came along with it.
	rbUntap(g, stage)
	third := effectstest.SeedPermanentWithOracle(g, me.ID, "Forest", "Basic Land — Forest", "")
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 1, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(third)}); err != nil {
		t.Fatalf("the copy cannot activate the Stage's ability again: %v", err)
	}
}

func TestThespiansStageRefusesANonLandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stage := effectstest.SeedPermanentWithOracle(g, me.ID, "Thespian's Stage", "Land", effectstest.RbThespiansStageOracl)
	bear := effectstest.DcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Error("a creature was a legal target for a land copy")
	}
}

// --- Virtue of Persistence // Locthwain Scorn -------------------------------

func TestVirtueOfPersistenceReanimatesFromAnyGraveyardAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.DcSeed(g, me.ID, "Virtue of Persistence", "Enchantment", effectstest.RbVirtuePersistOracle, 0, 0)
	dead := effectstest.PushGraveyardCardTyped(opp, "Fallen Giant", "Creature — Giant")
	// Walk to my NEXT upkeep (the trigger fires on my upkeep only).
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature did not come back")
	}
	if c.Controller != me.ID {
		t.Error("the creature is not under the Virtue controller's control")
	}
}

func TestLocthwainScornShrinksAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.DcSeed(g, opp.ID, "Victim", "Creature — Bear", "", 4, 4)
	life := me.Life

	card := rbVirtueCard(me.ID)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: effectstest.B16TargetCard(victim)}); err != nil {
		t.Fatalf("cast Locthwain Scorn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveToughness(t, g, victim); got != 1 {
		t.Errorf("toughness %d after -3/-3 on a 4/4, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Dowsing Dagger // Lost Vale ---------------------------------------------

func TestDowsingDaggerGivesTheOpponentTwoPlants(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Dowsing Dagger", "Artifact — Equipment", effectstest.RbDowsingDaggerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The enters trigger targets: it may need a pick.
	if pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	plants := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Plant" {
			plants++
			if c.Controller != opp.ID {
				t.Error("a Plant is not under the target opponent's control")
			}
		}
	}
	if plants != 2 {
		t.Errorf("Plants = %d, want 2", plants)
	}
}

func TestDowsingDaggerPumpsAndTransformsAfterCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dagger := effectstest.PushBattlefieldCardWithTimestamp(g, rbDowsingDaggerCard(me.ID))
	bear := effectstest.DcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}{C}")
	effectstest.EquipTo(t, g, me.ID, dagger, bear)
	if effectstest.EffectivePower(t, g, bear) != 4 || effectstest.EffectiveToughness(t, g, bear) != 3 {
		t.Fatalf("equipped Bear is %d/%d, want 4/3", effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear))
	}

	// A noncombat hit does not trigger it.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.DcCard(t, g, dagger); c.ActiveFace != 0 {
		t.Fatal("noncombat damage transformed the Dagger")
	}

	effectstest.DealCombatDamageToPlayer(g, bear, opp.ID, 4)
	effectstest.B06AnswerAllTriggerPrompts(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.DcCard(t, g, dagger)
	if c.ActiveFace != 1 || c.Name != "Lost Vale" {
		t.Fatalf("after combat damage the Dagger is face %d %q, want Lost Vale", c.ActiveFace, c.Name)
	}
	if effectstest.EffectivePower(t, g, bear) != 2 {
		t.Error("the Bear kept the +2/+1 after the Equipment became a land")
	}
}

// rbDowsingDaggerCard is the printed transforming card as the deck
// importer builds it: per-face data, face 0 up.
func rbDowsingDaggerCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.RbDowsingDaggerOracle,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Dowsing Dagger", TypeLine: "Artifact — Equipment", ManaCost: "{2}"},
			{Name: "Lost Vale", TypeLine: "Land"},
		},
	}
	c.SetFace(0)
	return c
}

// rbVirtueCard builds the printed adventure card the way deck.toGameCard
// does: per-face data, then SetFace(0).
func rbVirtueCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.RbVirtuePersistOracle,
		Layout:     game.LayoutAdventure,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Virtue of Persistence", TypeLine: "Enchantment", ManaCost: "{5}{B}{B}", Colors: []string{"B"}},
			{Name: "Locthwain Scorn", TypeLine: "Sorcery — Adventure", ManaCost: "{1}{B}", Colors: []string{"B"}},
		},
	}
	c.SetFace(0)
	return c
}

func TestLostValeTapsForThreeOfOneColour(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.RbDowsingDaggerOracle + "#1")
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Lost Vale spec = %+v", spec)
	}
	if spec.ManaAbilities[0].Produced != effects.OneColorOfAmount(3) {
		t.Errorf("Lost Vale produces %q, want three of one colour", spec.ManaAbilities[0].Produced)
	}
}
