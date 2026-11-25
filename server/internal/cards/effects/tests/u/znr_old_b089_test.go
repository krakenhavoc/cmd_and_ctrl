package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_old_b089_test.go — Zendikar Rising, S82 OLD batch 89 (#5508). Every
// test plays the real catalog entry through the engine.

const (
	oldB089SpareSuppliesOracle = "088fb5dc-65eb-42bf-8a32-76fa6b79b79f"
	oldB089CovetedPrizeOracle  = "56204530-7ab5-4610-9570-fba2673725ad"
	oldB089CoralhelmOracle     = "621fc513-b16a-4c09-93f8-424e688471e2"
	oldB089MasterOfWindsOracle = "7accdc91-3140-4096-a71f-0cee38d29e96"
	oldB089RelicVialOracle     = "7e86d4a6-e4cd-418a-8bbc-72cdad753e3a"
	oldB089CleansingWildfire   = "993b92f4-4160-4171-9e67-90dc935030fe"
	oldB089WindrobberOracle    = "be8b9bef-a98d-40a8-bedd-d78f1f08f93b"
	oldB089GlacialGraspOracle  = "c3006be3-a796-4238-95dd-1e8e47c063b5"
	oldB089SpoilsOracle        = "c4f3a6c6-a123-43ad-9513-d65fbff10d02"
	oldB089InfiltratorOracle   = "e6108735-8370-4611-ad59-d7da7fd0323f"
	oldB089TrapfinderOracle    = "e95a9488-3f86-4598-9c6e-6b1b959f8570"
	oldB089ConundrumOracle     = "f86b9a01-62fc-4120-8914-5054ed48d37e"
	oldB089BearType            = "Creature — Bear"
	oldB089ClericType          = "Creature — Human Cleric"
	oldB089RogueType           = "Creature — Human Rogue"
	oldB089WarriorType         = "Creature — Human Warrior"
	oldB089WizardType          = "Creature — Human Wizard"
	oldB089ForestType          = "Basic Land — Forest"
	oldB089IslandType          = "Basic Land — Island"
)

func TestOldB089EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB089SpareSuppliesOracle: "Spare Supplies", oldB089CovetedPrizeOracle: "Coveted Prize",
		oldB089CoralhelmOracle: "Coralhelm Chronicler", oldB089MasterOfWindsOracle: "Master of Winds",
		oldB089RelicVialOracle: "Relic Vial", oldB089CleansingWildfire: "Cleansing Wildfire",
		oldB089WindrobberOracle: "Merfolk Windrobber", oldB089GlacialGraspOracle: "Glacial Grasp",
		oldB089SpoilsOracle: "Spoils of Adventure", oldB089InfiltratorOracle: "Sure-Footed Infiltrator",
		oldB089TrapfinderOracle: "Nimble Trapfinder", oldB089ConundrumOracle: "Confounding Conundrum",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB089Cast puts the card in the active player's hand and casts it for
// free (a {0} cost), then settles the stack.
func oldB089Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.MayhemHandCard(g, me, name, typeLine, "{0}", oracle)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB089Library(p *game.Player, n int) { effectstest.Msc18Library(p, n) }

func TestOldB089SpareSuppliesEntersTappedDrawsAndCashesIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB089Library(me, 6)
	hand := me.Hand.Size()
	id := oldB089Cast(t, g, "Spare Supplies", "Artifact", oldB089SpareSuppliesOracle, game.CastSpellParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the cast card left, the enter trigger drew)", me.Hand.Size(), hand+1)
	}
	if !effectstest.HmlTapped(t, g, id) {
		t.Error("it did not enter tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
	effectstest.P1g3Untap(g, id)
	effectstest.ApaMana(me, "G", "G")
	hand = me.Hand.Size()
	effectstest.ApaActivate(t, g, me, id, 0, game.ActivateAbilityParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after cashing in, want %d", me.Hand.Size(), hand+1)
	}
	if effectstest.OnBattlefieldByID(g, id) {
		t.Error("the artifact was not sacrificed")
	}
}

func TestOldB089CovetedPrizeCostsLessPerPartyMemberAndTutors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Healer", oldB089ClericType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", oldB089RogueType, 1, 1)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Fine Card", TypeLine: "Sorcery"}, game.Card{Name: "Other Card", TypeLine: "Sorcery"})
	id := effectstest.HandCardFull(me, "Coveted Prize", "Sorcery", "{4}{B}", oldB089CovetedPrizeOracle, []string{"B"})
	// Two party members: {4}{B} costs {2}{B}.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for two mana with a party of two")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {2}{B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c == nil {
		t.Fatal("no search prompt")
	}
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Fine Card")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the tutor", me.Hand.Size(), hand+1)
	}
	if effectstest.ResolutionCastFor(g, me.ID) != nil {
		t.Error("a free cast was offered without a full party")
	}
}

func TestOldB089CovetedPrizeOffersAFreeCastWithAFullParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Creature(g, me.ID, "Healer", oldB089ClericType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Thief", oldB089RogueType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Fighter", oldB089WarriorType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Mage", oldB089WizardType, 1, 1)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Cheap Spell", TypeLine: "Instant", ManaCost: "{2}"}, game.Card{Name: "Other Card", TypeLine: "Sorcery"})
	id := effectstest.HandCardFull(me, "Coveted Prize", "Sorcery", "{4}{B}", oldB089CovetedPrizeOracle, []string{"B"})
	// Four party members: {4}{B} costs {B}.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Cheap Spell")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ResolutionCastFor(g, me.ID) == nil {
		t.Error("no free cast was offered with a full party")
	}
}

func TestOldB089CoralhelmChroniclerLootsOnAKickedSpellAndFindsAKickerCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	oldB089Library(me, 8)
	// A catalog kicker card on top of the library.
	kicker := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: kicker, Name: "Ghastly Gloomhunter", TypeLine: "Creature — Zombie Bat", OracleID: "70ab65ff-724f-4126-b472-97fa03ca3a34", Owner: me.ID, Controller: me.ID})
	effectstest.CastHoldCreature(t, g, "Coralhelm Chronicler", "Creature — Merfolk Wizard", oldB089CoralhelmOracle, 2, 2)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no choose-cards prompt for the look at the top five")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != kicker {
		t.Fatalf("offered %v, want only the card with kicker", c.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, kicker)
	if !me.Hand.Contains(kicker) {
		t.Error("the kicker card is not in hand")
	}
}

func TestOldB089CoralhelmChroniclerLootsWhenYouCastAKickedSpellOnly(t *testing.T) {
	for _, kicked := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.AdvanceToMain(t, g)
		effectstest.B12Push(g, me.ID, "Coralhelm Chronicler", "Creature — Merfolk Wizard", oldB089CoralhelmOracle, 2, 2)
		oldB089Library(me, 8)
		id := effectstest.HandCardFull(me, "Ghastly Gloomhunter", "Creature — Zombie Bat", "{1}{B}", "70ab65ff-724f-4126-b472-97fa03ca3a34", []string{"B"})
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}{B}{B}"); err != nil {
			t.Fatalf("mana: %v", err)
		}
		params := game.CastSpellParams{}
		if kicked {
			params.OptionalCosts = []int{0}
		}
		hand := me.Hand.Size()
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		drew := me.Hand.Size() == hand // the spell left (-1), the loot drew (+1)
		if kicked != drew {
			t.Errorf("kicked=%v: hand %d -> %d", kicked, hand, me.Hand.Size())
		}
		if kicked {
			effectstest.DiscardFromHand(t, g, me.ID)
		}
	}
}

func TestOldB089MasterOfWindsLoots2ThenChoosesItsBasePowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB089Library(me, 10)
	hand := me.Hand.Size()
	master := effectstest.CastHoldCreature(t, g, "Master of Winds", "Creature — Sphinx Wizard", oldB089MasterOfWindsOracle, 1, 4)
	if !effectstest.HasEffectiveKeyword(t, g, master, "flying") {
		t.Error("no flying")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after draw two and discard one", me.Hand.Size(), hand+1)
	}
	for _, tc := range []struct {
		index int
		p, t  int
	}{{1, 4, 1}, {2, 1, 4}, {0, 1, 4}} {
		oldB089Cast(t, g, "Quick Instant", "Instant", "", game.CastSpellParams{})
		effectstest.AnswerOptionPick(t, g, me.ID, tc.index)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, master); p != tc.p || tough != tc.t {
			t.Errorf("option %d: %d/%d, want %d/%d", tc.index, p, tough, tc.p, tc.t)
		}
	}
}

func TestOldB089MasterOfWindsIgnoresOtherSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Master of Winds", "Creature — Sphinx Wizard", oldB089MasterOfWindsOracle, 1, 4)
	oldB089Cast(t, g, "Plain Bear", oldB089BearType, "", game.CastSpellParams{})
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) != nil {
		t.Error("a creature spell that is no Wizard offered the choice")
	}
	oldB089Cast(t, g, "Plain Wizard", oldB089WizardType, "", game.CastSpellParams{})
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID) == nil {
		t.Error("a Wizard spell did not offer the choice")
	}
}

func TestOldB089RelicVialSacrificesACreatureToDrawAndDrainsWithACleric(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	oldB089Library(me, 6)
	vial := effectstest.B12Push(g, me.ID, "Relic Vial", "Artifact", oldB089RelicVialOracle, 0, 0)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", oldB089BearType, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other", oldB089BearType, 1, 1)
	effectstest.ApaMana(me, "G", "G")
	hand := me.Hand.Size()
	lifeMe, lifeOpp := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, vial, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Life != lifeMe || opp.Life != lifeOpp {
		t.Error("the Vial drained with no Cleric in play")
	}
	// With a Cleric, a creature dying drains.
	effectstest.B12Creature(g, me.ID, "Healer", oldB089ClericType, 1, 1)
	// A read through the layers, as any state change does in play, so the grant is on the Vial when the creature dies.
	_ = effectstest.GaLayered(t, g, vial)
	effectstest.SbondDie(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != lifeMe+1 || opp.Life != lifeOpp-1 {
		t.Errorf("life %d/%d, want %d/%d after a creature died beside a Cleric", me.Life, opp.Life, lifeMe+1, lifeOpp-1)
	}
}

func TestOldB089CleansingWildfireDestroysALandItsControllerFetchesAndYouDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB089Library(me, 4)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", oldB089ForestType)
	effectstest.SeedSearchLibrary(opp, effectstest.SearchTestLand("Island", oldB089IslandType))
	hand := me.Hand.Size()
	id := effectstest.MayhemHandCard(g, me, "Cleansing Wildfire", "Sorcery", "{0}", oldB089CleansingWildfire)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, land) {
		t.Error("the land survived")
	}
	if c := effectstest.SearchChoiceFor(g, opp.ID); c == nil {
		t.Error("the land's controller was not asked to search")
	} else {
		effectstest.AnswerSearchNamed(t, g, opp.ID, "Island")
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != hand+1 { // the spell was added and left, the draw came
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if got := effectstest.CountBattlefieldNamed(g, opp.ID, "Island"); got != 1 {
		t.Errorf("%d Islands for the land's controller, want 1", got)
	}
}

func TestOldB089MerfolkWindrobberMillsAndSacrificesOnlyAgainstAFullGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB089Library(me, 6)
	oldB089Library(opp, 6)
	rob := effectstest.B12Push(g, me.ID, "Merfolk Windrobber", "Creature — Merfolk Rogue", oldB089WindrobberOracle, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, rob, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("sacrificed with no opponent's graveyard at eight")
	}
	for i := 0; i < 8; i++ {
		opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Old", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	}
	hand := me.Hand.Size()
	effectstest.Pr7Activate(t, g, me.ID, rob, 0, game.ActivateAbilityParams{Strict: true})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if effectstest.OnBattlefieldByID(g, rob) {
		t.Error("the Windrobber was not sacrificed")
	}
}

func TestOldB089MerfolkWindrobberMillsTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB089Library(opp, 6)
	rob := effectstest.B12Push(g, me.ID, "Merfolk Windrobber", "Creature — Merfolk Rogue", oldB089WindrobberOracle, 1, 1)
	library, yard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.AttackWith(t, g, opp.ID, rob)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library-1 || opp.Graveyard.Size() != yard+1 {
		t.Errorf("library %d->%d graveyard %d->%d, want one card milled", library, opp.Library.Size(), yard, opp.Graveyard.Size())
	}
}

func TestOldB089GlacialGraspTapsMillsFreezesAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB089Library(me, 4)
	oldB089Library(opp, 6)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB089BearType, 2, 2)
	hand, yard := me.Hand.Size(), opp.Graveyard.Size()
	id := effectstest.MayhemHandCard(g, me, "Glacial Grasp", "Instant", "{0}", oldB089GlacialGraspOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, bear) {
		t.Error("the creature is untapped")
	}
	if opp.Graveyard.Size() != yard+2 {
		t.Errorf("the creature's controller milled %d, want 2", opp.Graveyard.Size()-yard)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the spell was added and left, one card drawn)", me.Hand.Size(), hand+1)
	}
}

func TestOldB089SpoilsOfAdventureCostsLessPerPartyMemberAndGainsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB089Library(me, 6)
	effectstest.B12Creature(g, me.ID, "Healer", oldB089ClericType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Fighter", oldB089WarriorType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Mage", oldB089WizardType, 1, 1)
	id := effectstest.HandCardFull(me, "Spoils of Adventure", "Instant", "{4}{W}{U}", oldB089SpoilsOracle, []string{"W", "U"})
	// Three party members: {4}{W}{U} costs {1}{W}{U}.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for two mana with a party of three")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	hand, life := me.Hand.Size(), me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {1}{W}{U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1+3 || me.Life != life+3 {
		t.Errorf("hand %d life %d, want %d / %d", me.Hand.Size(), me.Life, hand+2, life+3)
	}
}

func TestOldB089SureFootedInfiltratorTapsAnotherRogueForEvasionAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	oldB089Library(me, 6)
	inf := effectstest.B12Push(g, me.ID, "Sure-Footed Infiltrator", "Creature — Merfolk Rogue", oldB089InfiltratorOracle, 2, 3)
	if err := g.ActivateCatalogAbility(me.ID, inf, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no other Rogue to tap")
	}
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", oldB089RogueType, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB089BearType, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, inf, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{rogue}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, rogue) {
		t.Error("the other Rogue was not tapped")
	}
	hand := me.Hand.Size()
	effectstest.BrAttack(t, g, inf)
	if err := g.DeclareBlocker(blocker, inf); err == nil {
		t.Error("blocked an unblockable Infiltrator")
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after combat damage", me.Hand.Size(), hand+1)
	}
}

func TestOldB089NimbleTrapfinderIsUnblockableAfterAnotherPartyMemberEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	trap := effectstest.B12Push(g, me.ID, "Nimble Trapfinder", "Creature — Human Rogue", oldB089TrapfinderOracle, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB089BearType, 2, 2)
	if oldB089Unblockable(g, trap) {
		t.Error("unblockable before another party creature entered")
	}
	oldB089Cast(t, g, "Plain Bear", oldB089BearType, "", game.CastSpellParams{})
	if oldB089Unblockable(g, trap) {
		t.Error("unblockable after a Bear entered")
	}
	oldB089Cast(t, g, "Another Cleric", oldB089ClericType, "", game.CastSpellParams{})
	if !oldB089Unblockable(g, trap) {
		t.Error("blockable after another Cleric entered")
	}
	effectstest.BrAttack(t, g, trap)
	if err := g.DeclareBlocker(blocker, trap); err == nil {
		t.Error("blocked an unblockable Trapfinder")
	}
}

func TestOldB089NimbleTrapfinderDoesNotCountItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	trap := effectstest.CastHoldCreature(t, g, "Nimble Trapfinder", "Creature — Human Rogue", oldB089TrapfinderOracle, 2, 1)
	if oldB089Unblockable(g, trap) {
		t.Error("the Trapfinder counted its own entry as another Rogue")
	}
	oldB089Cast(t, g, "Another Rogue", oldB089RogueType, "", game.CastSpellParams{})
	if !oldB089Unblockable(g, trap) {
		t.Errorf("another Rogue entering (%s) did not make it unblockable", me.Name)
	}
}

func TestOldB089NimbleTrapfinderGrantsADrawTriggerWithAFullParty(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB089Library(me, 8)
	effectstest.B12Push(g, me.ID, "Nimble Trapfinder", "Creature — Human Rogue", oldB089TrapfinderOracle, 2, 1)
	effectstest.B12Creature(g, me.ID, "Healer", oldB089ClericType, 1, 1)
	effectstest.B12Creature(g, me.ID, "Fighter", oldB089WarriorType, 1, 1)
	mage := effectstest.B12Creature(g, me.ID, "Mage", oldB089WizardType, 1, 1)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mage, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d: the Wizard did not draw on combat damage", me.Hand.Size(), hand+1)
	}
}

func TestOldB089ConfoundingConundrumDrawsWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB089Library(me, 4)
	hand := me.Hand.Size()
	oldB089Cast(t, g, "Confounding Conundrum", "Enchantment", oldB089ConundrumOracle, game.CastSpellParams{})
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the cast card was added and left, one drawn)", me.Hand.Size(), hand+1)
	}
}

func TestOldB089ConfoundingConundrumBouncesALandOnAnOpponentsSecondLandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, owner.ID, "Confounding Conundrum", "Enchantment", oldB089ConundrumOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 1) })
	first := effectstest.B31PlayLand(t, g, "Forest", oldB089ForestType, "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, active.ID) != nil {
		t.Fatal("the first land of the turn triggered it")
	}
	second := effectstest.B31PlayLand(t, g, "Island", oldB089IslandType, "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, active.ID) == nil {
		t.Fatal("the second land did not ask that player to return a land")
	}
	effectstest.AnswerOwnPermanents(t, g, active.ID, first)
	g.SettleResolution()
	if effectstest.OnBattlefieldByID(g, first) || !effectstest.OnBattlefieldByID(g, second) {
		t.Error("the chosen land did not return")
	}
	if !active.Hand.Contains(first) {
		t.Error("the land is not in its owner's hand")
	}
}

func TestOldB089ConfoundingConundrumIgnoresItsControllersLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, active.ID, "Confounding Conundrum", "Enchantment", oldB089ConundrumOracle, 0, 0)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 1) })
	effectstest.B31PlayLand(t, g, "Forest", oldB089ForestType, "")
	effectstest.B31PlayLand(t, g, "Island", oldB089IslandType, "")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, active.ID) != nil {
		t.Error("the Conundrum triggered on its own controller's lands")
	}
}

func oldB089Unblockable(g *game.Game, id uuid.UUID) bool {
	return game.Restricted(effectstest.ApaLive(g, id), game.CantBeBlocked)
}
