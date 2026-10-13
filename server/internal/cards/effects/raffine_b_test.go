package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// raffine_b_test.go — #3212 slice raffine-b: Dowsing Dagger // Lost
// Vale, Fetid Pools, Hall of Heliod's Generosity, Ondu Inversion,
// Plaza of Heroes, Realmbreaker, Silver Scrutiny, The Meathook
// Massacre, Thespian's Stage, Virtue of Persistence // Locthwain Scorn.

const (
	rbDowsingDaggerOracle = "df34a6ad-ae1c-4470-8c9e-49815bba1973"
	rbFetidPoolsOracle    = "32b03b48-06da-4a74-a7ac-e5ae39a4f428"
	rbHallOfHeliodOracle  = "2fc070dc-f2f7-4648-8069-31d74790a39c"
	rbOnduInversionOracle = "15fc4e74-300e-4c2d-8ed7-004553b2f7c2"
	rbPlazaOracle         = "9c58d241-4d9f-4b46-b8ee-f4587f9acfd6"
	rbRealmbreakerOracle  = "77398866-eebb-4091-b566-7726f8194a59"
	rbSilverScrutinyOracl = "2b0b4fc1-9ad6-4259-a0a2-9a14484c246e"
	rbMeathookOracle      = "127de52b-df75-4342-95a0-20d84c5bf916"
	rbThespiansStageOracl = "b01e698b-608a-4fc7-8073-b01d044743ec"
	rbVirtuePersistOracle = "fae36da0-bf9c-484f-b2fe-dd8ab2effc5c"
)

func rbPoolSize(p *game.Player) int { return len(p.ManaPool) }

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
	g := newCatalogGame(t)
	id := playLandFromHand(t, g, "Fetid Pools", rbFetidPoolsOracle)
	top100AssertEnteredTapped(t, g, id, "Fetid Pools")

	// A second copy cycles from hand for {2}: discard it, draw a card.
	g2 := newCatalogGame(t)
	me := g2.Seats[g2.Turn.ActiveSeat]
	for g2.Turn.Step != game.StepPrecombatMain {
		if _, err := g2.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	cid := pushCatalogHandCard(me, "Fetid Pools", "Land — Island Swamp", rbFetidPoolsOracle)
	handBefore := len(me.Hand.Cards)
	if err := g2.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.ActivateCatalogAbility(me.ID, cid, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if len(me.Hand.Cards) != handBefore {
		t.Errorf("hand %d -> %d, want unchanged (discard one, draw one)", handBefore, len(me.Hand.Cards))
	}
	if !inZone(me.Graveyard, cid) {
		t.Error("the cycled land is not in the graveyard")
	}
}

func TestFetidPoolsOffersBlueOrBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := seedPermanentWithOracle(g, me.ID, "Fetid Pools", "Land — Island Swamp", rbFetidPoolsOracle)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	pick := manaPickFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	hall := seedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", rbHallOfHeliodOracle)
	ench := pushGraveyardCardTyped(me, "Pacifism", "Enchantment — Aura")
	advanceToMain(t, g)

	dcMana(t, g, me.ID, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{
		Targets: b16TargetCard(ench),
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if inZone(me.Graveyard, ench) {
		t.Error("the enchantment is still in the graveyard")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != ench {
		t.Error("the enchantment is not on top of the library")
	}
}

func TestHallOfHeliodsGenerosityRefusesNonEnchantmentAndOpponentsCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hall := seedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", rbHallOfHeliodOracle)
	creature := pushGraveyardCardTyped(me, "Bear", "Creature — Bear")
	theirs := pushGraveyardCardTyped(opp, "Their Ench", "Enchantment")
	advanceToMain(t, g)

	for _, bad := range []uuid.UUID{creature, theirs} {
		dcMana(t, g, me.ID, "{C}{W}")
		err := g.ActivateCatalogAbility(me.ID, hall, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bad)})
		if err == nil {
			t.Fatal("an illegal target was accepted")
		}
		rbUntap(g, hall)
	}
}

// --- Ondu Inversion ----------------------------------------------------

func TestOnduInversionDestroysEveryNonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := dcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := dcSeed(g, opp.ID, "Theirs", "Artifact", "", 0, 0)
	land := dcSeed(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)

	castCatalogSpell(t, g, "Ondu Inversion", "Sorcery", rbOnduInversionOracle, nil)
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, bear) || battlefieldHas(g, theirs) {
		t.Error("a nonland permanent survived Ondu Inversion")
	}
	if !battlefieldHas(g, land) {
		t.Error("a land was destroyed by Ondu Inversion")
	}
	if spec, _ := Lookup(rbOnduInversionOracle); spec.Completeness != CompletenessFull {
		t.Error("Ondu Inversion is complete")
	}
}

// --- The Meathook Massacre -----------------------------------------------

func TestMeathookMassacreShrinksEveryCreatureByCastX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := dcSeed(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	big := dcSeed(g, opp.ID, "Big", "Creature — Giant", "", 5, 5)

	massacre := castXSpell(t, g, "The Meathook Massacre", "Legendary Enchantment", rbMeathookOracle, "{X}{B}{B}", 2, nil)
	waitForXTrigger(t, g, massacre)
	if effectivePower(t, g, big) != 5 {
		t.Fatal("-X/-X applied before the trigger resolved")
	}
	passPriorityAroundTable(t, g)
	if battlefieldHas(g, small) {
		t.Error("the 2/2 survived -2/-2")
	}
	if got := effectiveToughness(t, g, big); got != 3 {
		t.Errorf("the 5/5 has toughness %d after X=2, want 3", got)
	}
}

func TestMeathookMassacreDeathPayoffs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dcSeed(g, me.ID, "The Meathook Massacre", "Legendary Enchantment", rbMeathookOracle, 0, 0)
	mine := dcSeed(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := dcSeed(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	meLife, oppLife := me.Life, opp.Life

	for _, id := range []uuid.UUID{mine, theirs} {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	// Mine died: each opponent loses 1. Theirs died: I gain 1.
	if opp.Life != oppLife-1 {
		t.Errorf("opponent life %d -> %d, want -1", oppLife, opp.Life)
	}
	if me.Life != meLife+1 {
		t.Errorf("my life %d -> %d, want +1", meLife, me.Life)
	}
}

func TestMeathookMassacreEnteringUncastShrinksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := dcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	// Pushed straight onto the battlefield (no cast): X is zero.
	dcSeed(g, me.ID, "The Meathook Massacre", "Legendary Enchantment", rbMeathookOracle, 0, 0)
	passPriorityAroundTable(t, g)
	if !battlefieldHas(g, bear) {
		t.Error("an uncast Massacre shrank a creature")
	}
}

// --- Silver Scrutiny ---------------------------------------------------

func TestSilverScrutinyDrawsX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castXSpell(t, g, "Silver Scrutiny", "Sorcery", rbSilverScrutinyOracl, "{X}{U}{U}", 4, nil)
	before := len(me.Hand.Cards)
	passPriorityAroundTable(t, g)
	if got := len(me.Hand.Cards) - before; got != 4 {
		t.Errorf("drew %d cards, want 4", got)
	}
	spec, _ := Lookup(rbSilverScrutinyOracl)
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) == 0 {
		t.Error("the missing flash clause must be a declared caveat")
	}
}

// --- Plaza of Heroes -----------------------------------------------------

func TestPlazaOfHeroesColourlessAndLegendaryRestrictedMana(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	plaza := seedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", rbPlazaOracle)

	if err := g.ActivateManaAbility(me.ID, plaza, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if rbPoolSize(me) != 1 || me.ManaPool[0].Color != "C" {
		t.Fatalf("pool = %v, want one {C}", me.ManaPool)
	}
	me.ManaPool = nil
	rbUntap(g, plaza)

	if err := g.ActivateManaAbility(me.ID, plaza, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	pick := manaPickFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	plaza := seedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", rbPlazaOracle)

	// No legendary permanent: nothing.
	if err := g.ActivateManaAbility(me.ID, plaza, 2, game.ManaAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	if rbPoolSize(me) != 0 || manaPickFor(g, me.ID) != nil {
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
	if rbPoolSize(me) != 1 || me.ManaPool[0].Color != "U" {
		t.Errorf("pool = %v, want one {U}", me.ManaPool)
	}
}

func TestPlazaOfHeroesExileProtectsALegendaryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plaza := seedPermanentWithOracle(g, me.ID, "Plaza of Heroes", "Land", rbPlazaOracle)
	legend := dcSeed(g, opp.ID, "Their Legend", "Legendary Creature — Human", "", 2, 2)
	plain := dcSeed(g, me.ID, "Plain Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)

	dcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, plaza, 0, game.ActivateAbilityParams{Targets: b16TargetCard(plain)}); err == nil {
		t.Fatal("a non-legendary creature was a legal target")
	}
	rbUntap(g, plaza)
	dcMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, plaza, 0, game.ActivateAbilityParams{Targets: b16TargetCard(legend)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if battlefieldHas(g, plaza) {
		t.Error("the land was not exiled as a cost")
	}
	passPriorityAroundTable(t, g)
	abilities := effectiveAbilities(t, g, legend)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := seedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", rbRealmbreakerOracle)
	land := pushGraveyardCardTyped(opp, "Their Mountain", "Basic Land — Mountain")
	advanceToMain(t, g)
	libBefore := opp.Library.Size()

	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Errorf("milled %d cards, want 3", got)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the activator was not asked to choose a land")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{land}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	c, ok := battlefieldCard(g, land)
	if !ok {
		t.Fatal("the land did not reach the battlefield")
	}
	if c.Controller != me.ID {
		t.Error("the land is not under the activator's control")
	}
	if !c.Tapped || tapEventsFor(g, land) != 0 {
		t.Error("the land must enter tapped (not be tapped afterwards)")
	}

	// "If this land would leave the battlefield, exile it instead."
	if err := g.DestroyPermanentForEffect(land); err != nil {
		t.Fatalf("destroy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if inZone(opp.Graveyard, land) || inZone(me.Graveyard, land) {
		t.Error("the stolen land went to a graveyard instead of exile")
	}
	if !inZone(g.Exile, land) {
		t.Error("the stolen land was not exiled")
	}
}

func TestRealmbreakerAsksNothingWithoutALandCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tree := seedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", rbRealmbreakerOracle)
	pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	advanceToMain(t, g)
	// The library is 20 plain filler cards (no lands), so the mill adds none.
	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if latestChooseCardsFor(g, me.ID) != nil {
		t.Error("a land prompt was opened with no land card in the graveyard")
	}
}

func TestRealmbreakerSearchesForPraetors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tree := seedPermanentWithOracle(g, me.ID, "Realmbreaker, the Invasion Tree", "Legendary Artifact", rbRealmbreakerOracle)
	p1 := pushLibraryCardForTest(me, game.Card{Name: "Praetor One", TypeLine: "Legendary Creature — Phyrexian Praetor", Owner: me.ID, Controller: me.ID})
	p2 := pushLibraryCardForTest(me, game.Card{Name: "Praetor Two", TypeLine: "Legendary Creature — Phyrexian Praetor", Owner: me.ID, Controller: me.ID})
	other := pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)

	dcMana(t, g, me.ID, "{C}{C}{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tree, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if battlefieldHas(g, tree) {
		t.Error("Realmbreaker was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	// Take whatever search prompt exists: both Praetors, never the Bear.
	if pc := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); pc != nil {
		if err := g.ResolveSearchLibrary(pc.ID, me.ID, []uuid.UUID{p1, p2}); err != nil {
			t.Fatalf("search answer: %v", err)
		}
	}
	if !battlefieldHas(g, p1) || !battlefieldHas(g, p2) {
		t.Error("the Praetors did not reach the battlefield")
	}
	if battlefieldHas(g, other) {
		t.Error("a non-Praetor was fetched")
	}
}

// --- Thespian's Stage -----------------------------------------------------

func TestThespiansStageBecomesACopyAndKeepsItsAbility(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stage := seedPermanentWithOracle(g, me.ID, "Thespian's Stage", "Land", rbThespiansStageOracl)
	other := seedPermanentWithOracle(g, me.ID, "Hall of Heliod's Generosity", "Legendary Land", rbHallOfHeliodOracle)
	advanceToMain(t, g)

	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 0, game.ActivateAbilityParams{Targets: b16TargetCard(other)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, stage)
	if c.Name != "Hall of Heliod's Generosity" {
		t.Fatalf("the Stage is %q after copying, want the Hall", c.Name)
	}
	if !c.Tapped {
		t.Error("the tap was a cost; the copy stays tapped")
	}

	// It can copy again: the ability came along with it.
	rbUntap(g, stage)
	third := seedPermanentWithOracle(g, me.ID, "Forest", "Basic Land — Forest", "")
	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 1, game.ActivateAbilityParams{Targets: b16TargetCard(third)}); err != nil {
		t.Fatalf("the copy cannot activate the Stage's ability again: %v", err)
	}
}

func TestThespiansStageRefusesANonLandTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	stage := seedPermanentWithOracle(g, me.ID, "Thespian's Stage", "Land", rbThespiansStageOracl)
	bear := dcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	advanceToMain(t, g)
	dcMana(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stage, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Error("a creature was a legal target for a land copy")
	}
}

// --- Virtue of Persistence // Locthwain Scorn -------------------------------

func TestVirtueOfPersistenceReanimatesFromAnyGraveyardAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dcSeed(g, me.ID, "Virtue of Persistence", "Enchantment", rbVirtuePersistOracle, 0, 0)
	dead := pushGraveyardCardTyped(opp, "Fallen Giant", "Creature — Giant")
	// Walk to my NEXT upkeep (the trigger fires on my upkeep only).
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	if pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: dead}); err != nil {
			t.Fatalf("pick: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature did not come back")
	}
	if c.Controller != me.ID {
		t.Error("the creature is not under the Virtue controller's control")
	}
}

func TestLocthwainScornShrinksAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := dcSeed(g, opp.ID, "Victim", "Creature — Bear", "", 4, 4)
	life := me.Life

	card := rbVirtueCard(me.ID)
	me.Hand.PushTop(card)
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: 1, Targets: b16TargetCard(victim)}); err != nil {
		t.Fatalf("cast Locthwain Scorn: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectiveToughness(t, g, victim); got != 1 {
		t.Errorf("toughness %d after -3/-3 on a 4/4, want 1", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Dowsing Dagger // Lost Vale ---------------------------------------------

func TestDowsingDaggerGivesTheOpponentTwoPlants(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Dowsing Dagger", "Artifact — Equipment", rbDowsingDaggerOracle, nil)
	passPriorityAroundTable(t, g)
	// The enters trigger targets: it may need a pick.
	if pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		passPriorityAroundTable(t, g)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	dagger := pushBattlefieldCardWithTimestamp(g, rbDowsingDaggerCard(me.ID))
	bear := dcSeed(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	opp := g.Seats[1]
	advanceToMain(t, g)
	dcMana(t, g, me.ID, "{C}{C}")
	equipTo(t, g, me.ID, dagger, bear)
	if effectivePower(t, g, bear) != 4 || effectiveToughness(t, g, bear) != 3 {
		t.Fatalf("equipped Bear is %d/%d, want 4/3", effectivePower(t, g, bear), effectiveToughness(t, g, bear))
	}

	// A noncombat hit does not trigger it.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(bear, opp.ID, 2) })
	passPriorityAroundTable(t, g)
	if c := dcCard(t, g, dagger); c.ActiveFace != 0 {
		t.Fatal("noncombat damage transformed the Dagger")
	}

	dealCombatDamageToPlayer(g, bear, opp.ID, 4)
	b06AnswerAllTriggerPrompts(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	c := dcCard(t, g, dagger)
	if c.ActiveFace != 1 || c.Name != "Lost Vale" {
		t.Fatalf("after combat damage the Dagger is face %d %q, want Lost Vale", c.ActiveFace, c.Name)
	}
	if effectivePower(t, g, bear) != 2 {
		t.Error("the Bear kept the +2/+1 after the Equipment became a land")
	}
}

// rbDowsingDaggerCard is the printed transforming card as the deck
// importer builds it: per-face data, face 0 up.
func rbDowsingDaggerCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   rbDowsingDaggerOracle,
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
		OracleID:   rbVirtuePersistOracle,
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
	spec, ok := Lookup(rbDowsingDaggerOracle + "#1")
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Lost Vale spec = %+v", spec)
	}
	if spec.ManaAbilities[0].Produced != OneColorOfAmount(3) {
		t.Errorf("Lost Vale produces %q, want three of one colour", spec.ManaAbilities[0].Produced)
	}
}
