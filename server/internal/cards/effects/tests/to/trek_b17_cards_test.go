package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b17_cards_test.go — Star Trek, batch 17 (#4973). Each test plays
// the real catalog entry and asserts what its text says.

const (
	trekB17PerilsOracle      = "f1f77fb4-2c83-4cbf-894e-624abe6cd95f"
	trekB17BombardOracle     = "2c2fdeff-a0c6-44cb-9603-4a9865f62985"
	trekB17PatrolOracle      = "fba54c2d-a664-4e4c-b9d4-e24cf2523ab9"
	trekB17CascadeOracle     = "46ce40e9-b7fb-4f40-a7ab-f8a2234d1d63"
	trekB17QuarkOracle       = "61703e94-cb57-44d8-b37b-7b14b3cfd86d"
	trekB17CosmozoanOracle   = "5cbe6066-07f8-45de-b3b4-b8590d111b0d"
	trekB17RedAlertOracle    = "4636afb5-453a-4f27-8916-1cbc32459ba6"
	trekB17SarcophagusOracle = "22b86863-c992-405b-949a-bdf8859884ef"
	trekB17ProfiteerOracle   = "e96a9f9a-9326-4f0b-98ad-23210c97299f"
)

func TestTrekB17EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB17PerilsOracle:      "Perils of the Past",
		trekB17BombardOracle:     "Planetary Bombardment",
		trekB17PatrolOracle:      "Planetary Patrol",
		trekB17CascadeOracle:     "Plasma Cascade",
		trekB17QuarkOracle:       "Quark, Humble Proprietor",
		trekB17CosmozoanOracle:   "Ravenous Cosmozoan",
		trekB17RedAlertOracle:    "Red Alert!",
		trekB17SarcophagusOracle: "Sarcophagus, Ship of the Dead",
		trekB17ProfiteerOracle:   "Savvy Profiteer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- shared helpers ------------------------------------------------------------

// trekB17AbilityIndex is the position of the activated ability labelled
// `label` on a permanent.
func trekB17AbilityIndex(t *testing.T, g *game.Game, id uuid.UUID, label string) int {
	t.Helper()
	c := effectstest.LayeredCard(t, g, id)
	for i, a := range game.ActivatedAbilitiesForCard(c) {
		if a.Label == label {
			return i
		}
	}
	t.Fatalf("%s has no ability %q", c.Name, label)
	return -1
}

// trekB17Treasures is the Treasure tokens p controls.
func trekB17Treasures(g *game.Game, p uuid.UUID) []game.Card {
	var out []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == p && c.IsToken() {
				out = append(out, c)
			}
		}
	})
	return out
}

// trekB17Price is the total mana value a card would cost p to cast now.
func trekB17Price(t *testing.T, g *game.Game, p *game.Player, name, typeLine, cost, oracle string) int {
	t.Helper()
	card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID}
	g.WithWriteLock(func() { p.Hand.PushTop(card) })
	price, err := g.PriceCast(p.ID, card, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("PriceCast %s: %v", name, err)
	}
	return price.Total.ManaValue()
}

// trekB17Graveyard puts a card in p's graveyard.
func trekB17Graveyard(g *game.Game, p *game.Player, name, typeLine string) {
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
}

// --- Perils of the Past ------------------------------------------------------------

func TestTrekB17PerilsUseTheRelicGivesPlusThreePlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Perils of the Past", "Instant", "{1}{R}", trekB17PerilsOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(bear), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, bear); c.CurrentPower() != 5 || c.CurrentToughness() != 4 {
		t.Errorf("the creature is %d/%d, want 5/4", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestTrekB17PerilsDestroyTheRelicDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B27Push(g, opp.ID, "Their Rock", "Artifact", "", "{2}", 0, 0)
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Perils of the Past", "Instant", "{1}{R}", trekB17PerilsOracle,
		game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(rock), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneGraveyard {
		t.Errorf("the artifact is in %v, want its owner's graveyard", effectstest.B12ZoneOf(g, rock))
	}
}

func TestTrekB17PerilsModesKeepTheirOwnTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B27Push(g, opp.ID, "Their Rock", "Artifact", "", "{2}", 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{C}{R}")
	if _, err := effectstest.CastFromHandErr(g, me, "Perils of the Past", "Instant", "{1}{R}", trekB17PerilsOracle,
		game.CastSpellParams{Modes: []int{0}, Targets: effectstest.CardRefs(rock), Strict: true}); err == nil {
		t.Error("Use the Relic targeted a noncreature artifact")
	}
	if _, err := effectstest.CastFromHandErr(g, me, "Perils of the Past", "Instant", "{1}{R}", trekB17PerilsOracle,
		game.CastSpellParams{Modes: []int{1}, Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("Destroy the Relic targeted a creature that is not an artifact")
	}
}

// --- Planetary Bombardment ---------------------------------------------------------

func TestTrekB17BombardmentDealsTwoWhenNoCreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.AddPool(t, g, me.ID, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Planetary Bombardment", "Sorcery", "{R}", trekB17BombardOracle,
		game.CastSpellParams{Targets: effectstest.CardRefs(big), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, big)
	if effectstest.B12ZoneOf(g, big) != game.ZoneBattlefield || c.DamageMarked != 2 {
		t.Errorf("the 4/4 is in %v with %d damage, want alive with 2", effectstest.B12ZoneOf(g, big), c.DamageMarked)
	}
}

func TestTrekB17BombardmentDealsFourAfterACreatureDied(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fodder := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 4, 4)
	effectstest.B27Kill(g, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, fodder) != game.ZoneGraveyard {
		t.Fatal("setup: the bear did not die")
	}
	effectstest.AddPool(t, g, me.ID, "{R}")
	effectstest.CastFromHandForTest(t, g, me, "Planetary Bombardment", "Sorcery", "{R}", trekB17BombardOracle,
		game.CastSpellParams{Targets: effectstest.CardRefs(big), Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, big) != game.ZoneGraveyard {
		t.Errorf("the 4/4 is in %v, want it dead to 4 damage", effectstest.B12ZoneOf(g, big))
	}
}

func TestTrekB17BombardmentTargetsOnlyACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.AddPool(t, g, me.ID, "{R}")
	if _, err := effectstest.CastFromHandErr(g, me, "Planetary Bombardment", "Sorcery", "{R}", trekB17BombardOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true}); err == nil {
		t.Error("Planetary Bombardment targeted a player")
	}
}

// --- Planetary Patrol ---------------------------------------------------------------

func TestTrekB17PatrolGrowsOnLandfallUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B27Push(g, me.ID, "Planetary Patrol", "Creature — Scout", trekB17PatrolOracle, "{1}{R}", 2, 2, "R")
	effectstest.ToMainForCost(t, g)
	effectstest.B12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("after a land it is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.B12PlayFromHand(t, g, "Mountain", "Basic Land — Mountain", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 4 || c.CurrentToughness() != 4 {
		t.Errorf("after two lands it is %d/%d, want 4/4", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AangAdvanceToMain(t, g, 1)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("next turn it is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestTrekB17PatrolIgnoresAnOpponentsLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.B27Push(g, me.ID, "Planetary Patrol", "Creature — Scout", trekB17PatrolOracle, "{1}{R}", 2, 2, "R")
	effectstest.ToMainForCost(t, g)
	effectstest.B12Permanent(g, opp.ID, "Their Mountain", "Basic Land — Mountain")
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(opp.ID, game.Card{Name: "Their Land", TypeLine: "Land"}, 1)
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 2 {
		t.Errorf("an opponent's land pumped it to %d power", c.CurrentPower())
	}
}

// --- Plasma Cascade ------------------------------------------------------------------

func TestTrekB17CascadeDealsThreeThenLootsIfYouDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	trekB17LibraryCard(g, me, "Top Card")
	pitch := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	})
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Plasma Cascade", "Instant", "{2}{R}", trekB17CascadeOracle,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true})
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-3)
	}
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, pitch) != game.ZoneGraveyard {
		t.Error("the discarded card is not in the graveyard")
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d, want %d (one out, one in)", got, hand)
	}
}

func TestTrekB17CascadeDrawsNothingWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	effectstest.CastFromHandForTest(t, g, me, "Plasma Cascade", "Instant", "{2}{R}", trekB17CascadeOracle,
		game.CastSpellParams{Targets: effectstest.CardRefs(bear), Strict: true})
	hand, library := me.Hand.Size(), me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, bear) != game.ZoneGraveyard {
		t.Error("3 damage did not kill the 3/3")
	}
	if me.Hand.Size() != hand || me.Library.Size() != library {
		t.Errorf("hand %d library %d, want %d and %d (no discard, no draw)", me.Hand.Size(), me.Library.Size(), hand, library)
	}
}

// trekB17LibraryCard puts a plain card on top of p's library.
func trekB17LibraryCard(g *game.Game, p *game.Player, name string) {
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	})
}

// --- Quark, Humble Proprietor --------------------------------------------------------

func TestTrekB17QuarkStopsYouSacrificingTreasureTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B27Push(g, me.ID, "Quark, Humble Proprietor", "Legendary Creature — Ferengi Rogue", trekB17QuarkOracle, "{2}{R}", 1, 2, "R")
	g.WithWriteLock(func() {
		_ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1)
		_ = g.CreateTokenForEffect(opp.ID, effects.TreasureToken(), 1)
	})
	mine, theirs := trekB17Treasures(g, me.ID), trekB17Treasures(g, opp.ID)
	if len(mine) != 1 || len(theirs) != 1 {
		t.Fatalf("setup: %d and %d Treasures", len(mine), len(theirs))
	}
	if err := g.ActivateManaAbility(me.ID, mine[0].InstanceID, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("Quark's controller sacrificed a Treasure token for mana")
	}
	if effectstest.B12ZoneOf(g, mine[0].InstanceID) != game.ZoneBattlefield {
		t.Error("the Treasure left the battlefield")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(mine[0].InstanceID) })
	if effectstest.B12ZoneOf(g, mine[0].InstanceID) != game.ZoneBattlefield {
		t.Error("an effect sacrificed Quark's controller's Treasure token")
	}
	var oppBarred bool
	g.ReadSnapshot(func() { oppBarred = g.CantBeSacrificedForEffect(theirs[0].InstanceID) })
	if oppBarred {
		t.Error("an opponent's Treasure token can't be sacrificed")
	}
	if err := g.ActivateManaAbility(opp.ID, theirs[0].InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("the opponent could not crack their own Treasure: %v", err)
	}
}

func TestTrekB17QuarkLeavingFreesTheTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	quark := effectstest.B27Push(g, me.ID, "Quark, Humble Proprietor", "Legendary Creature — Ferengi Rogue", trekB17QuarkOracle, "{2}{R}", 1, 2, "R")
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
	effectstest.B27Kill(g, quark)
	effectstest.PassPriorityAroundTable(t, g)
	mine := trekB17Treasures(g, me.ID)
	if len(mine) != 1 {
		t.Fatalf("setup: %d Treasures", len(mine))
	}
	if err := g.ActivateManaAbility(me.ID, mine[0].InstanceID, 0, game.ManaAbilityParams{}); err != nil {
		t.Errorf("with Quark gone the Treasure still can't be sacrificed: %v", err)
	}
}

func TestTrekB17QuarkMakesATreasureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Quark, Humble Proprietor", "Legendary Creature — Ferengi Rogue", trekB17QuarkOracle, "{2}{R}", 1, 2, "R")
	effectstest.AdvanceToMain(t, g)
	before := len(trekB17Treasures(g, me.ID))
	// The next upkeep is the next player's.
	effectstest.AangAdvanceToMain(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	got := trekB17Treasures(g, me.ID)
	if len(got) != before+1 {
		t.Fatalf("Treasures after another player's upkeep = %d, want %d", len(got), before+1)
	}
	for _, c := range got {
		if !c.IsArtifact() || !c.HasSubtype("Treasure") {
			t.Errorf("token %s is not a Treasure artifact", c.TypeLine)
		}
	}
	for _, p := range g.Seats[1:] {
		if n := len(trekB17Treasures(g, p.ID)); n != 0 {
			t.Errorf("another player got %d Treasures", n)
		}
	}
}

// --- Ravenous Cosmozoan ---------------------------------------------------------------

func TestTrekB17CosmozoanCostsLessPerNoncreatureNonlandCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const name, typeLine, cost = "Ravenous Cosmozoan", "Creature — Fish", "{5}{R}"
	if got := trekB17Price(t, g, me, name, typeLine, cost, trekB17CosmozoanOracle); got != 6 {
		t.Fatalf("price with an empty graveyard = %d, want 6", got)
	}
	trekB17Graveyard(g, me, "Shock", "Instant")
	trekB17Graveyard(g, me, "Ponder", "Sorcery")
	trekB17Graveyard(g, me, "Sol Ring", "Artifact")
	trekB17Graveyard(g, me, "Grizzly Bears", "Creature — Bear")
	trekB17Graveyard(g, me, "Ornithopter", "Artifact Creature — Thopter")
	trekB17Graveyard(g, me, "Mountain", "Basic Land — Mountain")
	trekB17Graveyard(g, opp, "Their Shock", "Instant")
	if got := trekB17Price(t, g, me, name, typeLine, cost, trekB17CosmozoanOracle); got != 3 {
		t.Errorf("price = %d, want 3 (three noncreature, nonland cards of mine)", got)
	}
	for _, n := range []string{"a", "b", "c", "d", "e", "f"} {
		trekB17Graveyard(g, me, "Spell "+n, "Instant")
	}
	if got := trekB17Price(t, g, me, name, typeLine, cost, trekB17CosmozoanOracle); got != 1 {
		t.Errorf("price with nine = %d, want 1 (the {R} is never reduced)", got)
	}
}

func TestTrekB17CosmozoanIsACheapHastyFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	for _, n := range []string{"a", "b", "c"} {
		trekB17Graveyard(g, me, "Spell "+n, "Sorcery")
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{R}")
	id := effectstest.CastFromHandForTest(t, g, me, "Ravenous Cosmozoan", "Creature — Fish", "{5}{R}", trekB17CosmozoanOracle,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, id) != game.ZoneBattlefield {
		t.Fatal("the Cosmozoan did not resolve")
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Ravenous Cosmozoan lacks %s", kw)
		}
	}
}

// --- Red Alert! -----------------------------------------------------------------------

func TestTrekB17RedAlertCostsTwoLessOnlyWhileACreatureAttacksYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	if got := trekB17Price(t, g, me, "Red Alert!", "Instant", "{2}{R}", trekB17RedAlertOracle); got != 3 {
		t.Fatalf("price with nothing attacking = %d, want 3", got)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.Msc15AttackedBy(t, g, bear, g.Seats[2].ID)
	if got := trekB17Price(t, g, me, "Red Alert!", "Instant", "{2}{R}", trekB17RedAlertOracle); got != 3 {
		t.Errorf("price while a creature attacks someone else = %d, want 3", got)
	}
	effectstest.Msc15AttackedBy(t, g, bear, me.ID)
	if got := trekB17Price(t, g, me, "Red Alert!", "Instant", "{2}{R}", trekB17RedAlertOracle); got != 1 {
		t.Errorf("price while a creature attacks me = %d, want 1", got)
	}
}

func TestTrekB17RedAlertPumpsOnlyYourCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddPool(t, g, me.ID, "{R}")
	if _, err := effectstest.CastFromHandErr(g, me, "Red Alert!", "Instant", "{2}{R}", trekB17RedAlertOracle,
		game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for {R} with nothing attacking me")
	}
	effectstest.Msc15AttackedBy(t, g, theirs, me.ID)
	effectstest.CastFromHandForTest(t, g, me, "Red Alert!", "Instant", "{2}{R}", trekB17RedAlertOracle,
		game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.GaLayered(t, g, mine); c.CurrentPower() != 4 || c.CurrentToughness() != 2 {
		t.Errorf("my creature is %d/%d, want 4/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.GaLayered(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("the opponent's creature has %d power, want 2", c.CurrentPower())
	}
}

// --- Sarcophagus, Ship of the Dead ------------------------------------------------------

func trekB17Sarcophagus(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sarcophagus, Ship of the Dead", TypeLine: "Legendary Artifact — Spacecraft",
		OracleID: trekB17SarcophagusOracle, ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID,
	})
}

func TestTrekB17SarcophagusPumpsOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := trekB17Sarcophagus(t, g, me)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if c := effectstest.GaLayered(t, g, mine); c.CurrentPower() != 3 || c.CurrentToughness() != 2 {
		t.Errorf("my creature is %d/%d, want 3/2", c.CurrentPower(), c.CurrentToughness())
	}
	if c := effectstest.GaLayered(t, g, theirs); c.CurrentPower() != 2 {
		t.Errorf("the opponent's creature has %d power, want 2", c.CurrentPower())
	}
	if effectstest.GaLayered(t, g, ship).IsCreature() {
		t.Error("with no charge counters the Spacecraft is a creature")
	}
}

func TestTrekB17SarcophagusChargesWhenAnyCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := trekB17Sarcophagus(t, g, me)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B27Kill(g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 2 {
		t.Errorf("charge counters = %d, want 2 (one per creature that died)", got)
	}
	// A creature that is exiled has not died.
	other := effectstest.B12Creature(g, opp.ID, "Their Elk", "Creature — Elk", 2, 2)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 2 {
		t.Errorf("charge counters = %d after an exile, want 2", got)
	}
}

func TestTrekB17SarcophagusStationsToAFlyingCreatureAndStopsCharging(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := trekB17Sarcophagus(t, g, me)
	crew := effectstest.B12Creature(g, me.ID, "Big Crew", "Creature — Giant", 9, 9)
	// The anthem makes the crew a 10-power creature.
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.GaLayered(t, g, ship)
	if c.Counters[game.CounterCharge] != 10 {
		t.Fatalf("charge counters = %d, want 10", c.Counters[game.CounterCharge])
	}
	if !c.IsCreature() || c.CurrentPower() != 4 || c.CurrentToughness() != 4 || !game.HasKeyword(&c, "flying") {
		t.Errorf("at 10+ it is creature=%v %d/%d flying=%v, want a 4/4 flier", c.IsCreature(), c.CurrentPower(), c.CurrentToughness(), game.HasKeyword(&c, "flying"))
	}
	if c.CurrentPower() != 4 {
		t.Error("the anthem pumps itself")
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 10 {
		t.Errorf("charge counters = %d, want 10: it is a creature now", got)
	}
}

func TestTrekB17SarcophagusRechecksItsIfOnResolution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := trekB17Sarcophagus(t, g, me)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 9) })
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B27Kill(g, victim)
	g.RunStateChecksForTest()
	// With the trigger waiting, the ship reaches 10 and becomes a creature.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(ship, game.CounterCharge, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 10 {
		t.Errorf("charge counters = %d, want 10: the trigger rechecks and does nothing", got)
	}
}

// --- Savvy Profiteer --------------------------------------------------------------------

func TestTrekB17ProfiteerTapsForRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Savvy Profiteer", "Creature — Ferengi Rogue", trekB17ProfiteerOracle, false)
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{})
	if me.ManaPool.CountColor("R") != 1 || me.ManaPool.Total() != 1 {
		t.Errorf("pool = %v, want {R}", me.ManaPool)
	}
}

func TestTrekB17ProfiteerMakesATreasureForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Savvy Profiteer", "Creature — Ferengi Rogue", trekB17ProfiteerOracle, false)
	idx := trekB17AbilityIndex(t, g, id, "{3}, {T}: Create a Treasure token.")
	effectstest.AddPool(t, g, me.ID, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with only two mana")
	}
	effectstest.AddPool(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := trekB17Treasures(g, me.ID)
	if len(got) != 1 {
		t.Fatalf("Treasures = %d, want 1", len(got))
	}
	if !got[0].IsArtifact() || !got[0].HasSubtype("Treasure") {
		t.Errorf("token %s is not a Treasure artifact", got[0].TypeLine)
	}
	if !effectstest.LayeredCard(t, g, id).Tapped {
		t.Error("the ability did not tap the Profiteer")
	}
}

func TestTrekB17ProfiteerIsSummoningSick(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Savvy Profiteer", "Creature — Ferengi Rogue", trekB17ProfiteerOracle, true)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("a summoning-sick Profiteer tapped for mana")
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	idx := trekB17AbilityIndex(t, g, id, "{3}, {T}: Create a Treasure token.")
	if err := g.ActivateCatalogAbility(me.ID, id, idx, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a summoning-sick Profiteer made a Treasure")
	}
}
