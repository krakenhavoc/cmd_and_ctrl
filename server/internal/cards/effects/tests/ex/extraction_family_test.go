package testsex

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Surgical Extraction chooses the target without exiling it: the target
// is a candidate like every other card with its name, and "any number"
// lets the caster leave any of them, the target included.
func TestSurgicalExtractionTakesAnyNumberIncludingNone(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	effectstest.CastCatalogSpell(t, x.G, "Surgical Extraction", "Instant", effectstest.SurgicalExtractionOracle, x.Target("gy"))
	effectstest.PassPriorityAroundTable(t, x.G)

	if x.G.Exile.Contains(x.Cards["gy"]) {
		t.Fatal("the target was exiled before the search")
	}
	effectstest.WantKeys(t, x.Candidates(t), "gy", "gy 2", "hand", "lib")
	if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); x.G.SearchFloorLocked(c) != 0 {
		t.Error("any number: finding nothing must be legal")
	}
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["gy 2"], x.Cards["hand"])
	x.Exiled(t, "gy 2", "hand")
	x.NotExiled(t, "gy", "lib", "gy other", "hand other", "lib other")
	if x.Searches() != 1 {
		t.Errorf("%d searches of their library, want 1", x.Searches())
	}
}

// "Other than a basic land card", in any graveyard; and the Phyrexian
// symbol may be paid with 2 life and no mana (CR 107.4f).
func TestSurgicalExtractionTargetsAndPhyrexianMana(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Surgical Extraction", "Instant", effectstest.SurgicalExtractionOracle, x.Target("gy forest")); err == nil {
		t.Fatal("a basic land card was a legal target")
	}
	life := x.Me.Life
	effectstest.TdCastFromHand(t, x.G, game.Card{Name: "Surgical Extraction", TypeLine: "Instant", OracleID: effectstest.SurgicalExtractionOracle,
		ManaCost: "{B/P}"}, game.CastSpellParams{Strict: true, PhyrexianLife: 1, Targets: x.Target("gy")})
	if got := life - x.Me.Life; got != game.PhyrexianLifePerSymbol {
		t.Errorf("paid %d life for {B/P}, want %d", got, game.PhyrexianLifePerSymbol)
	}
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.AnswerSearchFailToFind(t, x.G, x.Me.ID)
	x.NotExiled(t, "gy", "gy 2", "hand", "lib")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1: the library is shuffled whatever was taken", x.Searches())
	}
}

// Extirpate has split second, and "all": both graveyard copies, the target
// among them, are exiled without being offered; the hand and library copies
// are the caster's to take or leave.
func TestExtirpateFindsEveryGraveyardCopyUnderSplitSecond(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	spell := effectstest.CastCatalogSpell(t, x.G, "Extirpate", "Instant", effectstest.ExtirpateOracle, x.Target("gy"))
	effectstest.RequireSplitSecondOn(t, x.G, spell)
	effectstest.PassPriorityAroundTable(t, x.G)

	effectstest.WantKeys(t, x.Candidates(t), "hand", "lib")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["lib"])
	x.Exiled(t, "gy", "gy 2", "lib")
	x.NotExiled(t, "hand", "gy other", "hand other", "lib other", "gy forest")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1", x.Searches())
	}
}

// With no copy in the hand or the library there is nothing to choose:
// the graveyard copies go and the library is still shuffled.
func TestExtirpateWithOnlyGraveyardCopiesDoesNotAsk(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	x.G.WithWriteLock(func() {
		_, _ = x.Opp.Hand.Remove(x.Cards["hand"])
		_, _ = x.Opp.Library.Remove(x.Cards["lib"])
	})
	effectstest.CastCatalogSpell(t, x.G, "Extirpate", "Instant", effectstest.ExtirpateOracle, x.Target("gy"))
	effectstest.PassPriorityAroundTable(t, x.G)
	if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); c != nil {
		t.Fatalf("asked with nothing to choose: %+v", c)
	}
	x.Exiled(t, "gy", "gy 2")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1", x.Searches())
	}
}

// Eradicate exiles the creature, then every graveyard card with its name,
// and offers the hand and library copies.
func TestEradicateExilesTheCreatureAndEveryGraveyardCopy(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.Permanent("bears", "Grizzly Bears", "Creature — Bear", "G")
	effectstest.CastCatalogSpell(t, x.G, "Eradicate", "Sorcery", effectstest.EradicateOracle, x.Target("bears"))
	effectstest.PassPriorityAroundTable(t, x.G)

	x.Exiled(t, "bears")
	effectstest.WantKeys(t, x.Candidates(t), "hand", "lib")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"], x.Cards["lib"])
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
	x.NotExiled(t, "gy other", "hand other", "lib other")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1", x.Searches())
	}
}

// "Nonblack", and "its controller": a creature the opponent controls but
// another player owns sends the search through the controller's zones.
func TestEradicateSearchesTheControllersZonesAndSparesBlack(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.Permanent("black", "Vampire", "Creature — Vampire", "B")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Eradicate", "Sorcery", effectstest.EradicateOracle, x.Target("black")); err == nil {
		t.Fatal("a black creature was a legal target")
	}
	stolen := effectstest.PushBattlefieldCardWithTimestamp(x.G, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: x.Opp2.ID, Controller: x.Opp.ID,
	})
	ownerCopy := uuid.New()
	x.Opp2.Graveyard.PushTop(game.Card{InstanceID: ownerCopy, Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Owner: x.Opp2.ID, Controller: x.Opp2.ID})
	effectstest.CastCatalogSpell(t, x.G, "Eradicate", "Sorcery", effectstest.EradicateOracle, []game.TargetRef{{Kind: game.TargetCard, ID: stolen}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.AnswerSearchFailToFind(t, x.G, x.Me.ID)
	if !x.G.Exile.Contains(stolen) {
		t.Error("the creature was not exiled")
	}
	x.Exiled(t, "gy", "gy 2")
	if x.G.Exile.Contains(ownerCopy) {
		t.Error("the owner's graveyard was searched instead of the controller's")
	}
}

// Scour on an enchantment, with nothing in the hidden zones: no prompt.
func TestScourExilesTheEnchantmentAndItsGraveyardCopies(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Pacifism", "Enchantment — Aura")
	x.G.WithWriteLock(func() {
		_, _ = x.Opp.Hand.Remove(x.Cards["hand"])
		_, _ = x.Opp.Library.Remove(x.Cards["lib"])
	})
	x.Permanent("ench", "Pacifism", "Enchantment")
	x.Permanent("art", "Sol Ring", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Scour", "Instant", effectstest.ScourOracle, x.Target("art")); err == nil {
		t.Fatal("an artifact was a legal target")
	}
	effectstest.CastCatalogSpell(t, x.G, "Scour", "Instant", effectstest.ScourOracle, x.Target("ench"))
	effectstest.PassPriorityAroundTable(t, x.G)
	if c := effectstest.SearchChoiceFor(x.G, x.Me.ID); c != nil {
		t.Fatalf("asked with nothing to choose: %+v", c)
	}
	x.Exiled(t, "ench", "gy", "gy 2")
	x.NotExiled(t, "art")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1", x.Searches())
	}
}

func TestSplinterExilesTheArtifactAndEveryCopy(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Sol Ring", "Artifact")
	x.Permanent("art", "Sol Ring", "Artifact")
	x.Permanent("ench", "Pacifism", "Enchantment")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Splinter", "Sorcery", effectstest.SplinterOracle, x.Target("ench")); err == nil {
		t.Fatal("an enchantment was a legal target")
	}
	effectstest.CastCatalogSpell(t, x.G, "Splinter", "Sorcery", effectstest.SplinterOracle, x.Target("art"))
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "art")
	effectstest.WantKeys(t, x.Candidates(t), "hand", "lib")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"], x.Cards["lib"])
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
}

// Sowing Salt: a basic land is not a legal target; a nonbasic land goes,
// with every graveyard copy.
func TestSowingSaltExilesANonbasicLandAndItsCopies(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Command Tower", "Land")
	x.Permanent("tower", "Command Tower", "Land")
	x.Permanent("forest", "Forest", "Basic Land — Forest")
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Sowing Salt", "Sorcery", effectstest.SowingSaltOracle, x.Target("forest")); err == nil {
		t.Fatal("a basic land was a legal target")
	}
	effectstest.CastCatalogSpell(t, x.G, "Sowing Salt", "Sorcery", effectstest.SowingSaltOracle, x.Target("tower"))
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "tower")
	effectstest.WantKeys(t, x.Candidates(t), "hand", "lib")
	effectstest.AnswerSearchFailToFind(t, x.G, x.Me.ID)
	x.Exiled(t, "gy", "gy 2")
	x.NotExiled(t, "forest", "hand", "lib", "gy forest")
}

// Quash counters the spell, which lands in the graveyard and is exiled
// with the other graveyard copies; the hidden copies are offered.
func TestQuashCountersAndExilesEveryGraveyardCopy(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	bolt := uuid.New()
	x.Opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant",
		Owner: x.Opp.ID, Controller: x.Opp.ID})
	x.Cards["cast"] = bolt
	effectstest.AdvanceToMain(t, x.G)
	if err := x.G.CastSpell(x.Opp.ID, bolt, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Bolt: %v", err)
	}
	effectstest.CastCatalogSpell(t, x.G, "Quash", "Instant", effectstest.QuashOracle, x.Target("cast"))
	effectstest.PassPriorityAroundTable(t, x.G)

	if !x.Opp.Graveyard.Contains(x.Cards["cast"]) {
		t.Fatal("the Bolt was not countered into its owner's graveyard")
	}
	effectstest.WantKeys(t, x.Candidates(t), "hand", "lib")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"])
	x.Exiled(t, "cast", "gy", "gy 2", "hand")
	x.NotExiled(t, "lib", "gy other", "hand other")
	if x.Searches() != 1 {
		t.Errorf("%d searches, want 1", x.Searches())
	}
}

// "Target instant or sorcery spell": a creature spell is not one.
func TestQuashRefusesACreatureSpell(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	x.Cards["cast"] = effectstest.TdCastFromHand(t, x.G, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2}, game.CastSpellParams{})
	if err := effectstest.CastCatalogSpellErr(t, x.G, "Quash", "Instant", effectstest.QuashOracle, x.Target("cast")); err == nil {
		t.Error("a creature spell was a legal target")
	}
}

// Every card of the family is catalogued as Full.
func TestExtractionFamilyIsComplete(t *testing.T) {
	for _, oid := range []string{effectstest.SurgicalExtractionOracle, effectstest.ExtirpateOracle, effectstest.EradicateOracle, effectstest.ScourOracle,
		effectstest.SplinterOracle, effectstest.SowingSaltOracle, effectstest.QuashOracle} {
		spec, ok := effects.Lookup(oid)
		if !ok || spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: registered %v, completeness %v", oid, ok, spec.Completeness)
		}
	}
}
