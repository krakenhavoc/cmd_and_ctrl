package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #6167 — Yasharn, Implacable Earth: "Players can't pay life or
// sacrifice nonland permanents to cast spells or activate abilities."
// CR 101.2: "can't" beats "can"; CR 118.3: an unpayable cost can't be
// paid; CR 601.2h and CR 602.2b: the costs of casting and activating.

const yasharnOracle = "6dc23410-4d8f-4c8a-92b4-a8b3fae1af4e"

func yasharnFor(g *game.Game, owner uuid.UUID) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Yasharn, Implacable Earth", OracleID: yasharnOracle,
		TypeLine: "Legendary Creature — Elemental Boar", ManaCost: "{2}{G}{W}", Power: 4, Toughness: 4,
		Owner: owner, Controller: owner,
	})
}

func yasharnRemove(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("removing Yasharn: %v", err)
	}
}

func yasharnOffersActivation(g *game.Game, seat, source uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Kind == legal.KindActivate && m.Source == source {
			return true
		}
	}
	return false
}

func TestYasharnIsCompleteWithNoCaveat(t *testing.T) {
	spec, ok := effects.Lookup(yasharnOracle)
	if !ok {
		t.Fatal("Yasharn is not in the catalog")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("completeness %v, caveats %v; want Full with none", spec.Completeness, spec.Caveats)
	}
}

// A sacrifice-a-creature activation (Viscera Seer) is neither accepted
// nor offered under an opponent's Yasharn, and works again once it is
// gone.
func TestYasharnStopsSacrificingACreatureToActivate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	seer := effectstest.PushCatalogPermanent(g, me.ID, "Viscera Seer", "Creature — Vampire Wizard", effectstest.VisceraSeerOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Human", "", false)
	effectstest.SeedLibrary(me, "Top", "Second")
	yasharn := yasharnFor(g, opp.ID)

	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{fodder},
	}); err == nil {
		t.Fatal("Viscera Seer sacrificed a creature under Yasharn")
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); !ok {
		t.Fatal("the refused activation sacrificed the creature")
	}
	if yasharnOffersActivation(g, me.ID, seer) {
		t.Error("the enumerator offers Viscera Seer's sacrifice under Yasharn")
	}

	yasharnRemove(t, g, yasharn)
	if !yasharnOffersActivation(g, me.ID, seer) {
		t.Error("with Yasharn gone the enumerator does not offer Viscera Seer")
	}
	if err := g.ActivateCatalogAbility(me.ID, seer, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("with Yasharn gone Viscera Seer is refused: %v", err)
	}
}

// A spell with a sacrifice-a-creature additional cost (Village Rites)
// can't be cast.
func TestYasharnStopsASacrificeAdditionalCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Human", "", false)
	yasharnFor(g, me.ID)
	hand := me.Hand.Size()
	if _, err := effectstest.TryCastWithSacrifice(g, "Village Rites", "Instant", effectstest.VillageRitesOracle, []uuid.UUID{fodder}); err == nil {
		t.Fatal("Village Rites was cast by sacrificing a creature under Yasharn")
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); !ok {
		t.Error("the refused cast sacrificed the creature")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("the refused cast left the hand at %d, want %d", me.Hand.Size(), hand+1)
	}
}

// A land may still be sacrificed: Evolving Wilds works under Yasharn.
func TestYasharnLetsALandBeSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wilds := effectstest.PushCatalogPermanent(g, me.ID, "Evolving Wilds", "Land", effectstest.EvolvingWildsOracle, false)
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	yasharnFor(g, me.ID)
	if !yasharnOffersActivation(g, me.ID, wilds) {
		t.Error("the enumerator does not offer Evolving Wilds under Yasharn")
	}
	if err := g.ActivateCatalogAbility(me.ID, wilds, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("Evolving Wilds under Yasharn: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, wilds); ok {
		t.Error("Evolving Wilds was not sacrificed")
	}
}

// "Activate abilities" names mana abilities too: a Treasure can't be
// sacrificed for mana, and Mana Confluence's "Pay 1 life" can't be paid.
func TestYasharnStopsManaAbilityCosts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	treasure := effectstest.PushTreasure(g, me)
	confluence := effectstest.SeedPermanentWithOracle(g, me.ID, "Mana Confluence", "Land", effectstest.ManaConfluenceOracle)
	yasharn := yasharnFor(g, me.ID)

	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("a Treasure was sacrificed for mana under Yasharn")
	}
	if _, ok := effectstest.BattlefieldCard(g, treasure); !ok {
		t.Error("the refused Treasure left the battlefield")
	}
	life := me.Life
	if err := g.ActivateManaAbility(me.ID, confluence, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Error("Mana Confluence paid 1 life under Yasharn")
	}
	if me.Life != life {
		t.Errorf("the refused Mana Confluence cost %d life", life-me.Life)
	}

	yasharnRemove(t, g, yasharn)
	if err := g.ActivateManaAbility(me.ID, confluence, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("with Yasharn gone Mana Confluence is refused: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, treasure, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("with Yasharn gone the Treasure is refused: %v", err)
	}
}

// The life half on a cast: Snuff Out's "pay 4 life" alternative cost is
// refused.
func TestYasharnStopsPayingLifeToCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	snuff := effectstest.VisibleHandCard(me, "Snuff Out", "Instant", "{3}{B}", effectstest.SnuffOutOracle, []string{"B"})
	effectstest.FreeSpellPermanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	bear := effectstest.FreeSpellPermanent(g, them.ID, "Grizzly Bears", "Creature — Bear")
	yasharnFor(g, them.ID)
	if keys := effectstest.OfferedKeys(t, g, me, snuff); effectstest.HasKey(keys, "pay_life") {
		t.Errorf("under Yasharn the pay_life offer is still shown: %v", keys)
	}
	life := me.Life
	if err := g.CastSpell(me.ID, snuff, game.CastSpellParams{
		Strict: true, AlternativeCost: "pay_life", Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Error("Snuff Out was cast for 4 life under Yasharn")
	}
	if me.Life != life {
		t.Errorf("the refused cast cost %d life", life-me.Life)
	}
}

// "When Yasharn enters, search your library for a basic Forest card and
// a basic Plains card, reveal those cards, put them into your hand, then
// shuffle."
func TestYasharnFetchesABasicForestAndABasicPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.StapleLibraryCard(me, "Forest", "Basic Land — Forest")
	effectstest.StapleLibraryCard(me, "Plains", "Basic Land — Plains")
	effectstest.StapleLibraryCard(me, "Swamp", "Basic Land — Swamp")
	id := yasharnFor(g, me.ID)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: id, Actor: me.ID})
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt from Yasharn's ETB")
	}
	if effectstest.SearchOptionNamed(g, c, "Swamp") != uuid.Nil {
		t.Error("the search offers a Swamp")
	}
	hand := me.Hand.Size()
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest", "Plains")
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("Yasharn put %d cards into hand, want 2", got)
	}
}
