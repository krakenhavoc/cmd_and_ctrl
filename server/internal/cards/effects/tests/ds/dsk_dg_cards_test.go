package testsds

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dsk_dg_cards_test.go — Duskmourn: House of Horror, cards D–G (#4466).
// Each test plays the real catalog entry and asserts what its text says.

const (
	dskDGMegalodonOracle     = "dc638413-c4f8-4f6a-a003-5ce8a10a8f94"
	dskDGEmergeOracle        = "fed2127c-7174-4863-ad6a-1847c96bbd3f"
	dskDGCornfieldOracle     = "6ccca5c2-66c3-495a-8d9e-1a9805569e52"
	dskDGExorciseOracle      = "58916455-844c-4d97-908f-8205bdc117c0"
	dskDGHuntedOracle        = "a23293cf-4352-4ce6-8a68-3b30f908e137"
	dskDGLostTeethOracle     = "3724cc42-1bbd-4a11-8469-ab8492ddba52"
	dskDGSurveillanceOracle  = "52478437-bb6c-48bb-855d-54a8a4d42d18"
	dskDGFinalVengeance      = "cdba7f43-df9e-4bfb-9480-721c787ad051"
	dskDGFranticOracle       = "5dbe4648-f4c9-4d26-98f5-58a7e53b149d"
	dskDGGhostOracle         = "e6a93240-8578-479e-b5e7-73f51b141cc8"
	dskDGTeddyOracle         = "02a8b849-e708-427c-92fe-9a41bb0ec7fe"
	dskDGGiveInOracle        = "2bf4811f-3462-4542-acb0-101e772bfec9"
	dskDGNightmareTypeLine   = "Enchantment Creature — Nightmare"
	dskDGStartingLife        = 40
	dskDGEmergeLifeGain      = 3
	dskDGLostTeethLifeChange = 1
)

func TestDskDGEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		dskDGMegalodonOracle:    "Daggermaw Megalodon",
		dskDGEmergeOracle:       "Emerge from the Cocoon",
		dskDGCornfieldOracle:    "Etched Cornfield",
		dskDGExorciseOracle:     "Exorcise",
		dskDGHuntedOracle:       "Fear of Being Hunted",
		dskDGLostTeethOracle:    "Fear of Lost Teeth",
		dskDGSurveillanceOracle: "Fear of Surveillance",
		dskDGFinalVengeance:     "Final Vengeance",
		dskDGFranticOracle:      "Frantic Strength",
		dskDGGhostOracle:        "Friendly Ghost",
		dskDGTeddyOracle:        "Friendly Teddy",
		dskDGGiveInOracle:       "Give In to Violence",
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

// --- Daggermaw Megalodon ----------------------------------------------------------

func TestDskDGDaggermawMegalodonIslandcyclesForAnIsland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shark := effectstest.B12Push(g, me.ID, "Daggermaw Megalodon", "Creature — Shark", dskDGMegalodonOracle, 5, 7)
	if !effectstest.HasKeywordOf(g, shark, "vigilance") {
		t.Error("Daggermaw Megalodon does not have vigilance")
	}

	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := effectstest.CycleFromHand(t, g, "Daggermaw Megalodon", "Creature — Shark", dskDGMegalodonOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search = %+v, want only the Island", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0]}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the Island did not reach the hand")
	}
}

// --- Emerge from the Cocoon ----------------------------------------------------------

func TestDskDGEmergeFromTheCocoonReturnsACreatureAndGainsThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.GyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, effectstest.DskDGTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if c := effectstest.B12Card(t, g, bear); c.Controller != me.ID {
		t.Errorf("the returned creature is controlled by %v, want the caster", c.Controller)
	}
	if want := dskDGStartingLife + dskDGEmergeLifeGain; me.Life != want {
		t.Errorf("life %d, want %d", me.Life, want)
	}
}

func TestDskDGEmergeFromTheCocoonTargetsOnlyYourCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.GyCard(opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, effectstest.DskDGTarget(theirs)); err == nil {
		t.Error("an opponent's creature card was a legal target")
	}
	spell := effectstest.GyCard(me, "Old Bolt", "Instant", "{R}", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, effectstest.DskDGTarget(spell)); err == nil {
		t.Error("a noncreature card was a legal target")
	}
}

// --- Etched Cornfield ------------------------------------------------------------------

func TestDskDGEtchedCornfieldEntersTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12PlayFromHand(t, g, "Etched Cornfield", "Land", dskDGCornfieldOracle, game.CastSpellParams{})
	if !effectstest.B13Tapped(t, g, land) {
		t.Error("everyone at 40: the land entered untapped")
	}
	if got := effectstest.LtProduced(t, g, land); !slices.Equal(got, []string{"{G|W}"}) {
		t.Errorf("the land adds %v, want {G|W}", got)
	}
	g.Seats[3].Life = 13
	if effectstest.B13Tapped(t, g, effectstest.B12PlayFromHand(t, g, "Etched Cornfield", "Land", dskDGCornfieldOracle, game.CastSpellParams{})) {
		t.Error("an opponent at 13: the land entered tapped")
	}
}

// --- Exorcise -----------------------------------------------------------------------------

func TestDskDGExorciseExilesABigCreatureOrAnyArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, effectstest.DskDGTarget(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Exile.Contains(big) {
		t.Error("the 4/4 was not exiled")
	}

	relic := effectstest.B12Permanent(g, opp.ID, "Their Relic", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, effectstest.DskDGTarget(relic))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(relic) {
		t.Error("the artifact was not exiled")
	}
}

func TestDskDGExorciseCantTargetASmallCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	small := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, effectstest.DskDGTarget(small)); err == nil {
		t.Error("a creature with power 3 was a legal target")
	}
}

// --- Fear of Being Hunted -------------------------------------------------------------

func TestDskDGFearOfBeingHuntedMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fear := effectstest.B12Push(g, me.ID, "Fear of Being Hunted", dskDGNightmareTypeLine, dskDGHuntedOracle, 4, 2)
	if !effectstest.HasKeywordOf(g, fear, "haste") {
		t.Error("Fear of Being Hunted does not have haste")
	}
	wall := effectstest.B12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	effectstest.ReqToBlockers(t, g, fear)
	effectstest.ReqRefusal(t, g.PassPriority())
	if err := effectstest.ReqBlock(t, g, fear, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Fear blocked: %v", err)
	}
}

// --- Fear of Lost Teeth -----------------------------------------------------------------

func TestDskDGFearOfLostTeethPingsAndGainsOneWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fear := effectstest.B12Push(g, me.ID, "Fear of Lost Teeth", dskDGNightmareTypeLine, dskDGLostTeethOracle, 1, 1)
	effectstest.KillCreature(t, g, me.ID, fear)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if want := dskDGStartingLife - dskDGLostTeethLifeChange; opp.Life != want {
		t.Errorf("opponent life %d, want %d", opp.Life, want)
	}
	if want := dskDGStartingLife + dskDGLostTeethLifeChange; me.Life != want {
		t.Errorf("controller life %d, want %d", me.Life, want)
	}
}

// --- Fear of Surveillance --------------------------------------------------------------

func TestDskDGFearOfSurveillanceSurveilsWhenItAttacksAndStaysUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fear := effectstest.B12Push(g, me.ID, "Fear of Surveillance", dskDGNightmareTypeLine, dskDGSurveillanceOracle, 2, 2)
	top := effectstest.SeedLibrary(me, "Top Card")[0]
	effectstest.DeclareAttack(t, g, opp.ID, fear)
	if effectstest.B13Tapped(t, g, fear) {
		t.Error("the attacking Fear tapped despite vigilance")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("attacking did not surveil")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, []uuid.UUID{top}, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card was not put into the graveyard")
	}
}

// --- Final Vengeance -------------------------------------------------------------------

func TestDskDGFinalVengeanceSacrificesAnEnchantmentToExileACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	charm := effectstest.B12Permanent(g, me.ID, "My Charm", "Enchantment")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastFromHand(t, g, game.Card{Name: "Final Vengeance", TypeLine: "Sorcery", OracleID: dskDGFinalVengeance, ManaCost: "{B}"},
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{charm}, Targets: effectstest.DskDGTarget(victim)})
	if !me.Graveyard.Contains(charm) {
		t.Error("the enchantment was not sacrificed as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the target creature was not exiled")
	}
}

func TestDskDGFinalVengeanceNeedsACreatureOrEnchantmentToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := effectstest.B12Permanent(g, me.ID, "My Relic", "Artifact")
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if _, err := effectstest.CastFromHandErr(g, me, "Final Vengeance", "Sorcery", "{B}", dskDGFinalVengeance,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{relic}, Targets: effectstest.DskDGTarget(victim)}); err == nil {
		t.Error("an artifact paid the sacrifice")
	}
	if !g.Battlefield.Contains(relic) || !g.Battlefield.Contains(victim) {
		t.Error("the refused cast changed the battlefield")
	}
}

// --- Frantic Strength ------------------------------------------------------------------

func TestDskDGFranticStrengthHasFlashAndGivesTwoAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// The table starts in the draw step: only flash lets the Aura in.
	if err := dskDGCastAtDrawStep(t, g, "Frantic Strength", "Enchantment — Aura", dskDGFranticOracle, bear); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted Bear %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "trample") {
		t.Error("the enchanted Bear lacks trample")
	}
}

// --- Friendly Ghost ---------------------------------------------------------------------

func TestDskDGFriendlyGhostPumpsATargetOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ghost := effectstest.CastFromHand(t, g, game.Card{
		Name: "Friendly Ghost", TypeLine: "Creature — Spirit", OracleID: dskDGGhostOracle,
		ManaCost: "{3}{W}", Power: 2, Toughness: 4,
	}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 4/6", p, tough)
	}
	if !effectstest.HasKeywordOf(g, ghost, "flying") {
		t.Error("Friendly Ghost does not have flying")
	}
}

// --- Friendly Teddy ---------------------------------------------------------------------

func TestDskDGFriendlyTeddyMakesEachPlayerDrawWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	teddy := effectstest.B12Push(g, me.ID, "Friendly Teddy", "Artifact Creature — Bear Toy", dskDGTeddyOracle, 2, 2)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	effectstest.KillCreature(t, g, me.ID, teddy)
	effectstest.PassPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != before[i]+1 {
			t.Errorf("seat %d hand %d, want %d", i, got, before[i]+1)
		}
	}
}

// --- Give In to Violence ----------------------------------------------------------------

func TestDskDGGiveInToViolencePumpsAndGrantsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := dskDGCastAtDrawStep(t, g, "Give In to Violence", "Instant", dskDGGiveInOracle, bear); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasKeywordOf(g, bear, "lifelink") {
		t.Error("the Bear did not gain lifelink")
	}
}

// dskDGCastAtDrawStep casts a card from the active seat's hand at
// `target` while the table is still in the draw step, where only an
// instant-speed spell may be cast.
func dskDGCastAtDrawStep(t *testing.T, g *game.Game, name, typeLine, oracle string, target uuid.UUID) error {
	t.Helper()
	if g.Turn.Step != game.StepDraw {
		t.Fatalf("setup: at %s, want the draw step", g.Turn.Step)
	}
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogHandCard(me, name, typeLine, oracle)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: effectstest.DskDGTarget(target)})
}
