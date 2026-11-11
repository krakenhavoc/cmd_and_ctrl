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
		cmdr2S43DazzlingSphinxOracle: "Dazzling Sphinx",
		cmdr2S43ChaosWandOracle:      "Chaos Wand",
		cmdr2S43StolenGoodsOracle:    "Stolen Goods",
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
