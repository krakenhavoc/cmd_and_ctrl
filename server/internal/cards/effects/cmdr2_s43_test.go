package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S43DazzlingSphinxOracle = "00e8b4c9-3a79-4864-a730-c0081d397c5b"
	cmdr2S43ChaosWandOracle      = "0b95c115-6745-4ce1-896b-9021e93c161e"
	cmdr2S43StolenGoodsOracle    = "9294d4d3-0df0-4070-9261-06f15b758a0d"
)

func TestCmdr2S43CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		cmdr2S43DazzlingSphinxOracle:           "Dazzling Sphinx",
		cmdr2S43ChaosWandOracle:                "Chaos Wand",
		cmdr2S43StolenGoodsOracle:              "Stolen Goods",
		"08bc10fd-8901-4836-ad4d-a867e1bcba77": "Epic Experiment",
		"0e5e46e3-f9af-47cc-a740-bf808d5cb4b1": "Tectonic Giant",
		"b391a4ce-52ce-4a8c-bbe5-6053951ee1d4": "Theater of Horrors",
		"06bb4272-25d1-46db-8585-15f919267e1e": "One with the Multiverse",
		"e330412f-46a1-4e7c-8766-f644606049e0": "Into the Pit",
		"53716e0b-4d4e-40b0-8370-38e11f73d195": "Demon of Fate's Design",
		"5ecf35b1-0bfd-4c60-a791-5a3b918dd630": "Marshland Bloodcaster",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull {
			t.Errorf("%s: registered %v, name %q, completeness %s", name, ok, spec.Name, spec.Completeness)
		}
	}
}

func TestCmdr2S43DazzlingSphinxCombatDamageOffersTheFreeInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), g.Seats[1]
	sphinx := b12Push(g, me.ID, "Dazzling Sphinx", "Creature — Sphinx", cmdr2S43DazzlingSphinxOracle, 4, 5)
	island := hobB10Card(opp, "Island", "Basic Land — Island", "")
	hit := hobB10Card(opp, "Divination", "Sorcery", "{2}{U}")
	deep := hobB10Card(opp, "Deep Card", "Creature — Bear", "{1}{G}")
	hobB10Stack(opp, island, hit, deep)
	declareAttack(t, g, opp.ID, sphinx)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)
	if err := castThroughResolutionCast(t, g, me.ID, hit.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Error("the cast was not free")
	}
	if !opp.Graveyard.Contains(hit.InstanceID) {
		t.Error("the cast spell is not in its owner's graveyard")
	}
	if !opp.Library.Contains(island.InstanceID) || g.Exile.Contains(island.InstanceID) {
		t.Error("the passed-over land is not back in the library")
	}
	if opp.Library.Size() != 2 || opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Errorf("library %d cards, want the untouched card on top of the bottomed one", opp.Library.Size())
	}
}

func TestCmdr2S43DazzlingSphinxHasFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sphinx := b12Push(g, me.ID, "Dazzling Sphinx", "Creature — Sphinx", cmdr2S43DazzlingSphinxOracle, 4, 5)
	c, ok := battlefieldCard(g, sphinx)
	if !ok || !game.HasKeyword(&c, "flying") {
		t.Error("Dazzling Sphinx does not fly")
	}
}

func TestCmdr2S43DazzlingSphinxDecliningBottomsTheHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), g.Seats[2]
	sphinx := b12Push(g, me.ID, "Dazzling Sphinx", "Creature — Sphinx", cmdr2S43DazzlingSphinxOracle, 4, 5)
	hit := hobB10Card(opp, "Shock", "Instant", "{R}")
	hobB10Stack(opp, hit)
	declareAttack(t, g, opp.ID, sphinx)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)
	declineResolutionCast(t, g, me.ID)
	if !opp.Library.Contains(hit.InstanceID) || len(g.Exile.Cards) != 0 {
		t.Error("the declined card is not on the bottom of its owner's library")
	}
}

func cmdr2S43WandGame(t *testing.T, victim int, lib ...game.Card) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[victim]
	wand := pushCatalogPermanent(g, me.ID, "Chaos Wand", "Artifact", cmdr2S43ChaosWandOracle, false)
	for i := range lib {
		lib[i].Owner, lib[i].Controller = opp.ID, opp.ID
	}
	hobB10Stack(opp, lib...)
	advanceToMain(t, g)
	return g, me, opp, wand
}

func TestCmdr2S43ChaosWandWalksTheOpponentsLibraryAndOffersTheFreeCast(t *testing.T) {
	bear := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	hit := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"}
	deep := game.Card{InstanceID: uuid.New(), Name: "Deep", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"}
	g, me, opp, wand := cmdr2S43WandGame(t, 1, bear, hit, deep)
	b06AddMana(me, "C", "C", "C", "C")
	b16Activate(t, g, me.ID, wand, 0, game.ActivateAbilityParams{Targets: b16TargetPlayer(opp.ID)})
	passPriorityAroundTable(t, g)

	if !b16Tapped(t, g, wand) || me.ManaPool.Total() != 0 {
		t.Error("the {4} and {T} were not paid")
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, hit.InstanceID)
	if err := castThroughResolutionCast(t, g, me.ID, hit.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting it free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(hit.InstanceID) {
		t.Error("the cast spell is not in its owner's graveyard")
	}
	if !opp.Library.Contains(bear.InstanceID) || g.Exile.Contains(bear.InstanceID) {
		t.Error("the passed-over card is not under the opponent's library")
	}
	if opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID != deep.InstanceID {
		t.Error("the unexiled card is not on top of the pile bottomed under it")
	}
}

func TestCmdr2S43ChaosWandNeedsFourManaAndAnOpponentTarget(t *testing.T) {
	hit := game.Card{InstanceID: uuid.New(), Name: "Shock", TypeLine: "Instant", ManaCost: "{R}"}
	g, me, opp, wand := cmdr2S43WandGame(t, 1, hit)
	b06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetPlayer(opp.ID)}); err == nil {
		t.Error("activated with three mana")
	}
	b06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, wand, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetPlayer(me.ID)}); err == nil {
		t.Error("targeted myself")
	}
}

func TestCmdr2S43StolenGoodsLeavesEverythingInExileAndGrantsTheNonlandCardForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := top100ActiveSeat(g), g.Seats[1]
	island := hobB10Card(opp, "Island", "Basic Land — Island", "")
	hit := hobB10Card(opp, "Divination", "Sorcery", "{2}{U}")
	deep := hobB10Card(opp, "Deep", "Creature — Bear", "{1}{G}")
	hobB10Stack(opp, island, hit, deep)
	advanceToMain(t, g)
	castCatalogSpell(t, g, "Stolen Goods", "Sorcery", cmdr2S43StolenGoodsOracle, b16TargetPlayer(opp.ID))
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(island.InstanceID) || !g.Exile.Contains(hit.InstanceID) {
		t.Fatal("the cards were not exiled")
	}
	if !opp.Library.Contains(deep.InstanceID) {
		t.Error("the card after the nonland card was exiled")
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("the cast was offered during the resolution")
	}
	// The land stays exiled and cannot be cast; the nonland card can, free.
	if err := g.CastSpell(me.ID, island.InstanceID, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err == nil {
		t.Error("cast the land")
	}
	if err := g.CastSpell(me.ID, hit.InstanceID, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("casting the exiled card free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Error("the cast was not free")
	}
	if !opp.Graveyard.Contains(hit.InstanceID) || !g.Exile.Contains(island.InstanceID) {
		t.Error("the spell did not resolve to its owner's graveyard, or the land moved")
	}
}

func TestCmdr2S43StolenGoodsRefusesAnUntargetedCast(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	b06AddMana(me, "U", "C", "C", "C")
	if err := castCatalogSpellErr(t, g, "Stolen Goods", "Sorcery", cmdr2S43StolenGoodsOracle, b16TargetPlayer(me.ID)); err == nil {
		t.Error("targeted myself")
	}
}

const (
	cmdr2S43EpicExperimentOracle = "08bc10fd-8901-4836-ad4d-a867e1bcba77"
	cmdr2S43DemonOracle          = "53716e0b-4d4e-40b0-8370-38e11f73d195"
)

func TestCmdr2S43EpicExperimentCastsInstantsAndSorceriesUpToXAndBinsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	island := hobB10Card(me, "Island", "Basic Land — Island", "")
	shock := hobB10Card(me, "Shock", "Instant", "{R}")
	big := hobB10Card(me, "Big Spell", "Sorcery", "{3}{U}{U}")
	deep := hobB10Card(me, "Deep Card", "Instant", "{R}")
	hobB10Stack(me, island, shock, big, deep)

	castFromHandForTest(t, g, me, "Epic Experiment", "Sorcery", "{X}{U}{R}", cmdr2S43EpicExperimentOracle,
		game.CastSpellParams{XValue: 3})
	passPriorityAroundTable(t, g)

	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, shock.InstanceID)
	if err := castThroughResolutionCast(t, g, me.ID, big.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast a spell with mana value 5 off X = 3")
	}
	if err := castThroughResolutionCast(t, g, me.ID, island.InstanceID, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("cast a land")
	}
	if err := castThroughResolutionCast(t, g, me.ID, shock.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting Shock free: %v", err)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		declineResolutionCast(t, g, me.ID)
	}
	passPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(island.InstanceID) || !me.Graveyard.Contains(big.InstanceID) {
		t.Error("the cards that were not cast did not go to the graveyard")
	}
	if g.Exile.Contains(island.InstanceID) || g.Exile.Contains(big.InstanceID) {
		t.Error("a card is still in exile")
	}
	if !me.Library.Contains(deep.InstanceID) {
		t.Error("the fourth card was exiled off X = 3")
	}
}

func TestCmdr2S43EpicExperimentDecliningBinsEverything(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	shock := hobB10Card(me, "Shock", "Instant", "{R}")
	bear := hobB10Card(me, "Bear", "Creature — Bear", "{1}{G}")
	hobB10Stack(me, shock, bear)
	castFromHandForTest(t, g, me, "Epic Experiment", "Sorcery", "{X}{U}{R}", cmdr2S43EpicExperimentOracle,
		game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, shock.InstanceID)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(shock.InstanceID) || !me.Graveyard.Contains(bear.InstanceID) || len(g.Exile.Cards) != 0 {
		t.Error("the exiled cards are not all in the graveyard")
	}
}

func TestCmdr2S43DemonOffersAnEnchantmentForLifeOncePerTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	b12Push(g, me.ID, "Demon of Fate's Design", "Enchantment Creature — Demon", cmdr2S43DemonOracle, 6, 6)
	mk := func(name, typeLine, cost string) uuid.UUID {
		c := hobB10Card(me, name, typeLine, cost)
		me.Hand.PushTop(c)
		return c.InstanceID
	}
	ench := mk("Ward", "Enchantment", "{3}{W}")
	ench2 := mk("Ward Two", "Enchantment", "{1}{W}")
	sorc := mk("Plain Sorcery", "Sorcery", "{1}{R}")
	key := cmdr2S43GrantedAltCostLifeEnchantmentOnceYourTurn

	if offerWithKey(grantedCardOffers(g, me.ID, sorc, game.ZoneHand), key) != nil {
		t.Error("a sorcery was offered the life cast")
	}
	offer := offerWithKey(grantedCardOffers(g, me.ID, ench, game.ZoneHand), key)
	if offer == nil || offer.Life != 4 || offer.ManaCost != "" {
		t.Fatalf("offer = %+v, want 4 life and no mana", offer)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, ench, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("casting for life: %v", err)
	}
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	passPriorityAroundTable(t, g)
	if offerWithKey(grantedCardOffers(g, me.ID, ench2, game.ZoneHand), key) != nil {
		t.Error("the second life cast of the turn was offered")
	}
	if err := g.CastSpell(me.ID, ench2, game.CastSpellParams{Strict: true, AlternativeCost: key}); err == nil {
		t.Error("a second life cast was accepted")
	}
}

func TestCmdr2S43DemonSacrificeGivesPowerEqualToTheEnchantmentsManaValue(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	demon := b12Push(g, me.ID, "Demon of Fate's Design", "Enchantment Creature — Demon", cmdr2S43DemonOracle, 6, 6)
	ench := hobB10Card(me, "Ward", "Enchantment", "{3}{W}")
	g.Battlefield.PushTop(ench)
	b06AddMana(me, "B", "C", "C")
	b16Activate(t, g, me.ID, demon, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{ench.InstanceID}})
	passPriorityAroundTable(t, g)
	c, _ := battlefieldCard(g, demon)
	if p := effectivePower(t, g, c.InstanceID); p != 10 {
		t.Errorf("power %d, want 10", p)
	}
	if _, ok := battlefieldCard(g, ench.InstanceID); ok {
		t.Error("the enchantment was not sacrificed")
	}
}

const cmdr2S43MarshlandBloodcasterOracle = "5ecf35b1-0bfd-4c60-a791-5a3b918dd630"

func TestCmdr2S43MarshlandBloodcasterLetsTheNextSpellBePaidInLifeOnce(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	caster := b12Push(g, me.ID, "Marshland Bloodcaster", "Creature — Vampire Warlock", cmdr2S43MarshlandBloodcasterOracle, 3, 5)
	hand := func(name, typeLine, cost string) uuid.UUID {
		c := hobB10Card(me, name, typeLine, cost)
		me.Hand.PushTop(c)
		return c.InstanceID
	}
	first := hand("Five Sorcery", "Sorcery", "{3}{B}{B}")
	second := hand("Two Sorcery", "Sorcery", "{1}{B}")
	key := game.GrantedAltCostNextLife

	if offerWithKey(grantedCardOffers(g, me.ID, first, game.ZoneHand), key) != nil {
		t.Fatal("a life cast was offered before the ability")
	}
	b06AddMana(me, "B", "C")
	b16Activate(t, g, me.ID, caster, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	offer := offerWithKey(grantedCardOffers(g, me.ID, first, game.ZoneHand), key)
	if offer == nil || offer.Life != 5 || offer.ManaCost != "" {
		t.Fatalf("offer = %+v, want 5 life and no mana", offer)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{Strict: true, AlternativeCost: key}); err != nil {
		t.Fatalf("casting for life: %v", err)
	}
	if me.Life != life-5 || me.ManaPool.Total() != 0 {
		t.Errorf("life %d (want %d), pool %d", me.Life, life-5, me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if offerWithKey(grantedCardOffers(g, me.ID, second, game.ZoneHand), key) != nil {
		t.Error("the promise outlived the next spell")
	}
}

func TestCmdr2S43MarshlandBloodcasterIsOnlyAnOffer(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	caster := b12Push(g, me.ID, "Marshland Bloodcaster", "Creature — Vampire Warlock", cmdr2S43MarshlandBloodcasterOracle, 3, 5)
	c := hobB10Card(me, "Two Sorcery", "Sorcery", "{1}{B}")
	me.Hand.PushTop(c)
	b06AddMana(me, "B", "C")
	b16Activate(t, g, me.ID, caster, 0, game.ActivateAbilityParams{Strict: true})
	passPriorityAroundTable(t, g)
	life := me.Life
	b06AddMana(me, "B", "C")
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting for mana: %v", err)
	}
	if me.Life != life {
		t.Errorf("life changed to %d though the spell was paid with mana", me.Life)
	}
	// The player with too little life is not offered it.
}

const cmdr2S43IntoThePitOracle = "e330412f-46a1-4e7c-8766-f644606049e0"

func TestCmdr2S43IntoThePitCastsTheTopCardBySacrificingANonlandPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Into the Pit", "Enchantment", cmdr2S43IntoThePitOracle, false)
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	land := seedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	bear := hobB10Card(me, "Bear", "Creature — Bear", "{1}")
	bear.Power, bear.Toughness = 2, 2
	hobB10Stack(me, bear)
	key := intoThePitAltCost
	floatForTest(g, me, "C")

	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("cast from the library without the sacrifice")
	}
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library", AlternativeCost: key}); err == nil {
		t.Fatal("cast without naming a permanent to sacrifice")
	}
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library", AlternativeCost: key, AltCostIDs: []uuid.UUID{land}}); err == nil {
		t.Fatal("sacrificed a land")
	}
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library", AlternativeCost: key, AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear.InstanceID) {
		t.Error("the Bear did not enter")
	}
	if g.Battlefield.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Error("the permanent was not sacrificed")
	}
	if me.ManaPool.Total() != 0 {
		t.Error("the mana cost was not paid")
	}
}

func TestCmdr2S43IntoThePitNeverPlaysALandFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "Into the Pit", "Enchantment", cmdr2S43IntoThePitOracle, false)
	fodder := b12Permanent(g, me.ID, "Fodder", "Artifact")
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, forest)
	if err := g.CastSpell(me.ID, forest.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library", AlternativeCost: intoThePitAltCost, AltCostIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("played a land from the top of the library")
	}
	if g.Battlefield.Contains(forest.InstanceID) {
		t.Error("the land entered the battlefield")
	}
}

const (
	cmdr2S43TheaterOfHorrorsOracle = "b391a4ce-52ce-4a8c-bbe5-6053951ee1d4"
	cmdr2S43MultiverseOracle       = "06bb4272-25d1-46db-8585-15f919267e1e"
	cmdr2S43TectonicGiantOracle    = "0e5e46e3-f9af-47cc-a740-bf808d5cb4b1"
)

// cmdr2S43ToMyNextUpkeep walks the table round to seat's next upkeep, with
// its beginning-of-upkeep triggers on the stack.
func cmdr2S43ToMyNextUpkeep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	for i := 0; i < 400; i++ {
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat == seat {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	t.Fatal("never reached my upkeep")
}

func TestCmdr2S43TheaterOfHorrorsExilesAtUpkeepAndPlaysThemOnlyAfterAnOpponentLostLife(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Theater of Horrors", "Enchantment", cmdr2S43TheaterOfHorrorsOracle, false)
	bear := hobB10Card(me, "Bear", "Creature — Bear", "{1}")
	bear.Power, bear.Toughness = 2, 2
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, bear, forest, hobB10Card(me, "Filler", "Sorcery", "{9}"))

	cmdr2S43ToMyNextUpkeep(t, g, seat)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bear.InstanceID) {
		t.Fatal("the top card was not exiled at the upkeep")
	}
	advanceToMainOf(t, g, seat)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("cast the exiled card before an opponent lost life")
	}
	g.WithWriteLock(func() {
		if err := g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1); err != nil {
			t.Fatalf("lose life: %v", err)
		}
	})
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast after an opponent lost life: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear.InstanceID) {
		t.Error("the Bear did not enter")
	}
}

func TestCmdr2S43TheaterOfHorrorsPlaysAnExiledLandButNotOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Theater of Horrors", "Enchantment", cmdr2S43TheaterOfHorrorsOracle, false)
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, forest, hobB10Card(me, "Filler", "Sorcery", "{9}"))
	cmdr2S43ToMyNextUpkeep(t, g, seat)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(forest.InstanceID) {
		t.Fatal("the land was not exiled")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	advanceToMainOf(t, g, seat)
	if err := g.CastSpell(me.ID, forest.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	if !g.Battlefield.Contains(forest.InstanceID) {
		t.Error("the land did not enter")
	}
}

func TestCmdr2S43TheaterOfHorrorsPingsAnOpponentForFourMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theater := pushCatalogPermanent(g, me.ID, "Theater of Horrors", "Enchantment", cmdr2S43TheaterOfHorrorsOracle, false)
	advanceToMain(t, g)
	life := opp.Life
	b06AddMana(me, "R", "C", "C", "C")
	b16Activate(t, g, me.ID, theater, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetPlayer(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
	if err := g.ActivateCatalogAbility(me.ID, theater, 0, game.ActivateAbilityParams{Strict: true, Targets: b16TargetPlayer(me.ID)}); err == nil {
		t.Error("targeted myself")
	}
}

func TestCmdr2S43OneWithTheMultiverseCastsFromTheTopAndOnceForFree(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "One with the Multiverse", "Enchantment", cmdr2S43MultiverseOracle, false)
	big := hobB10Card(me, "Big", "Creature — Bear", "{5}{G}")
	big.Power, big.Toughness = 2, 2
	hobB10Stack(me, big)
	key := cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn
	handBig := hobB10Card(me, "Hand Big", "Creature — Bear", "{4}{G}")
	handBig.Power, handBig.Toughness = 2, 2
	me.Hand.PushTop(handBig)

	if offerWithKey(grantedCardOffers(g, me.ID, big.InstanceID, game.ZoneLibrary), key) == nil {
		t.Fatal("the top card has no free offer")
	}
	// Without the free offer the top card is cast for its mana cost: too dear.
	if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library"}); err == nil {
		t.Fatal("cast a six-drop with no mana")
	}
	if err := g.CastSpell(me.ID, big.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library", AlternativeCost: key}); err != nil {
		t.Fatalf("free cast from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big.InstanceID) || me.ManaPool.Total() != 0 {
		t.Error("the free cast did not put the creature onto the battlefield")
	}
	if offerWithKey(grantedCardOffers(g, me.ID, handBig.InstanceID, game.ZoneHand), key) != nil {
		t.Error("a second free cast this turn was offered")
	}
}

func TestCmdr2S43OneWithTheMultiverseFreeCastReachesTheHandAndPlaysLandsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	advanceToMain(t, g)
	pushCatalogPermanent(g, me.ID, "One with the Multiverse", "Enchantment", cmdr2S43MultiverseOracle, false)
	forest := hobB10Card(me, "Forest", "Basic Land — Forest", "")
	hobB10Stack(me, forest)
	handBig := hobB10Card(me, "Hand Big", "Sorcery", "{4}{G}")
	me.Hand.PushTop(handBig)
	if err := g.CastSpell(me.ID, forest.InstanceID, game.CastSpellParams{Strict: true, FromZone: "library"}); err != nil {
		t.Fatalf("playing a land from the top: %v", err)
	}
	if offerWithKey(grantedCardOffers(g, me.ID, handBig.InstanceID, game.ZoneHand), cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn) == nil {
		t.Fatal("a hand card has no free offer")
	}
	if err := g.CastSpell(me.ID, handBig.InstanceID, game.CastSpellParams{Strict: true, AlternativeCost: cmdr2S43GrantedAltCostFreeHandOrTopOnceYourTurn}); err != nil {
		t.Fatalf("free cast from hand: %v", err)
	}
}

func TestCmdr2S43TectonicGiantAttackingDealsThreeToEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	giant := b12Push(g, me.ID, "Tectonic Giant", "Creature — Elemental Giant", cmdr2S43TectonicGiantOracle, 3, 4)
	lives := b28Lives(g)
	declareAttack(t, g, g.Seats[1].ID, giant)
	lockInAttacks(t, g)
	cmdr2S43ChooseMode(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		want := lives[i]
		if p.ID != me.ID {
			want -= 3
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
}

// cmdr2S43ChooseMode answers chooser's open mode_pick prompt with one mode.
func cmdr2S43ChooseMode(t *testing.T, g *game.Game, chooser uuid.UUID, mode int) {
	t.Helper()
	c := modePickChoiceFor(g, chooser)
	if c == nil {
		t.Fatal("no mode prompt is open")
	}
	if err := g.ResolveModePick(c.ID, chooser, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
}

func TestCmdr2S43TectonicGiantSecondModeExilesTwoAndPlaysTheChosenOne(t *testing.T) {
	g := newCatalogGame(t)
	me := top100ActiveSeat(g)
	giant := b12Push(g, me.ID, "Tectonic Giant", "Creature — Elemental Giant", cmdr2S43TectonicGiantOracle, 3, 4)
	a := hobB10Card(me, "A", "Creature — Bear", "{1}")
	b := hobB10Card(me, "B", "Creature — Bear", "{1}")
	c := hobB10Card(me, "C", "Creature — Bear", "{1}")
	hobB10Stack(me, a, b, c)
	declareAttack(t, g, g.Seats[1].ID, giant)
	lockInAttacks(t, g)
	cmdr2S43ChooseMode(t, g, me.ID, 1)
	passPriorityAroundTable(t, g)
	p := chooseCardsChoiceFor(g, me.ID)
	if p == nil || len(p.ChooseCards) != 2 {
		t.Fatalf("want a choice among the two exiled cards, got %+v", p)
	}
	answerChooseCards(t, g, me.ID, a.InstanceID)
	if !g.Exile.Contains(a.InstanceID) || !g.Exile.Contains(b.InstanceID) {
		t.Error("the top two cards are not both exiled")
	}
	if !me.Library.Contains(c.InstanceID) {
		t.Error("a third card was exiled")
	}
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.CastSpell(me.ID, b.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Error("cast the card that was not chosen")
	}
	if err := g.CastSpell(me.ID, a.InstanceID, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Errorf("cast the chosen card: %v", err)
	}
}

func TestCmdr2S43TectonicGiantTriggersOnlyOffAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := b12Push(g, me.ID, "Tectonic Giant", "Creature — Elemental Giant", cmdr2S43TectonicGiantOracle, 3, 4)
	advanceToMain(t, g)
	// My own spell targeting it does nothing.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, b16TargetCard(giant))
	passPriorityAroundTable(t, g)
	if modePickChoiceFor(g, me.ID) != nil {
		t.Error("my own spell triggered the Giant")
	}
	// An opponent's spell does.
	advanceToMainOf(t, g, 1)
	bolt := hobB10Card(opp, "Their Bolt", "Instant", "{R}")
	bolt.OracleID = lightningBoltOracle
	opp.Hand.PushTop(bolt)
	floatForTest(g, opp, "R")
	if err := g.CastSpell(opp.ID, bolt.InstanceID, game.CastSpellParams{Strict: true, Targets: b16TargetCard(giant)}); err != nil {
		t.Fatalf("their Bolt: %v", err)
	}
	if modePickChoiceFor(g, me.ID) == nil {
		t.Error("an opponent's spell did not trigger the Giant")
	}
}
