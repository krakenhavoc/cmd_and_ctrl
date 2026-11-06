package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dskDGTarget is a card target.
func dskDGTarget(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// --- Daggermaw Megalodon ----------------------------------------------------------

func TestDskDGDaggermawMegalodonIslandcyclesForAnIsland(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	shark := b12Push(g, me.ID, "Daggermaw Megalodon", "Creature — Shark", dskDGMegalodonOracle, 5, 7)
	if !hasKeywordOf(g, shark, "vigilance") {
		t.Error("Daggermaw Megalodon does not have vigilance")
	}

	g.WithWriteLock(func() { me.Hand.Cards = nil })
	ids := seedSearchLibrary(me,
		searchTestLand("Island", "Basic Land — Island"),
		searchTestLand("Forest", "Basic Land — Forest"),
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	id, _ := cycleFromHand(t, g, "Daggermaw Megalodon", "Creature — Shark", dskDGMegalodonOracle, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := gyCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	castCatalogSpell(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, dskDGTarget(bear))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if c := b12Card(t, g, bear); c.Controller != me.ID {
		t.Errorf("the returned creature is controlled by %v, want the caster", c.Controller)
	}
	if want := dskDGStartingLife + dskDGEmergeLifeGain; me.Life != want {
		t.Errorf("life %d, want %d", me.Life, want)
	}
}

func TestDskDGEmergeFromTheCocoonTargetsOnlyYourCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := gyCard(opp, "Their Bear", "Creature — Bear", "{1}{G}", "")
	if err := castCatalogSpellErr(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, dskDGTarget(theirs)); err == nil {
		t.Error("an opponent's creature card was a legal target")
	}
	spell := gyCard(me, "Old Bolt", "Instant", "{R}", "")
	if err := castCatalogSpellErr(t, g, "Emerge from the Cocoon", "Sorcery", dskDGEmergeOracle, dskDGTarget(spell)); err == nil {
		t.Error("a noncreature card was a legal target")
	}
}

// --- Etched Cornfield ------------------------------------------------------------------

func TestDskDGEtchedCornfieldEntersTappedUnlessAPlayerIsAtThirteen(t *testing.T) {
	g := newCatalogGame(t)
	land := b12PlayFromHand(t, g, "Etched Cornfield", "Land", dskDGCornfieldOracle, game.CastSpellParams{})
	if !b13Tapped(t, g, land) {
		t.Error("everyone at 40: the land entered untapped")
	}
	if got := ltProduced(t, g, land); !slices.Equal(got, []string{"{G|W}"}) {
		t.Errorf("the land adds %v, want {G|W}", got)
	}
	g.Seats[3].Life = 13
	if b13Tapped(t, g, b12PlayFromHand(t, g, "Etched Cornfield", "Land", dskDGCornfieldOracle, game.CastSpellParams{})) {
		t.Error("an opponent at 13: the land entered tapped")
	}
}

// --- Exorcise -----------------------------------------------------------------------------

func TestDskDGExorciseExilesABigCreatureOrAnyArtifact(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	big := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	castCatalogSpell(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, dskDGTarget(big))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) || !g.Exile.Contains(big) {
		t.Error("the 4/4 was not exiled")
	}

	relic := b12Permanent(g, opp.ID, "Their Relic", "Artifact")
	castCatalogSpell(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, dskDGTarget(relic))
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(relic) {
		t.Error("the artifact was not exiled")
	}
}

func TestDskDGExorciseCantTargetASmallCreature(t *testing.T) {
	g := newCatalogGame(t)
	small := b12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 3, 3)
	if err := castCatalogSpellErr(t, g, "Exorcise", "Sorcery", dskDGExorciseOracle, dskDGTarget(small)); err == nil {
		t.Error("a creature with power 3 was a legal target")
	}
}

// --- Fear of Being Hunted -------------------------------------------------------------

func TestDskDGFearOfBeingHuntedMustBeBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	fear := b12Push(g, me.ID, "Fear of Being Hunted", dskDGNightmareTypeLine, dskDGHuntedOracle, 4, 2)
	if !hasKeywordOf(g, fear, "haste") {
		t.Error("Fear of Being Hunted does not have haste")
	}
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 4)
	reqToBlockers(t, g, fear)
	reqRefusal(t, g.PassPriority())
	if err := reqBlock(t, g, fear, wall); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the Fear blocked: %v", err)
	}
}

// --- Fear of Lost Teeth -----------------------------------------------------------------

func TestDskDGFearOfLostTeethPingsAndGainsOneWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fear := b12Push(g, me.ID, "Fear of Lost Teeth", dskDGNightmareTypeLine, dskDGLostTeethOracle, 1, 1)
	killCreature(t, g, me.ID, fear)
	pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("the dies trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if want := dskDGStartingLife - dskDGLostTeethLifeChange; opp.Life != want {
		t.Errorf("opponent life %d, want %d", opp.Life, want)
	}
	if want := dskDGStartingLife + dskDGLostTeethLifeChange; me.Life != want {
		t.Errorf("controller life %d, want %d", me.Life, want)
	}
}

// --- Fear of Surveillance --------------------------------------------------------------

func TestDskDGFearOfSurveillanceSurveilsWhenItAttacksAndStaysUntapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fear := b12Push(g, me.ID, "Fear of Surveillance", dskDGNightmareTypeLine, dskDGSurveillanceOracle, 2, 2)
	top := seedLibrary(me, "Top Card")[0]
	declareAttack(t, g, opp.ID, fear)
	if b13Tapped(t, g, fear) {
		t.Error("the attacking Fear tapped despite vigilance")
	}
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	charm := b12Permanent(g, me.ID, "My Charm", "Enchantment")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castFromHand(t, g, game.Card{Name: "Final Vengeance", TypeLine: "Sorcery", OracleID: dskDGFinalVengeance, ManaCost: "{B}"},
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{charm}, Targets: dskDGTarget(victim)})
	if !me.Graveyard.Contains(charm) {
		t.Error("the enchantment was not sacrificed as the spell was cast")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(victim) {
		t.Error("the target creature was not exiled")
	}
}

func TestDskDGFinalVengeanceNeedsACreatureOrEnchantmentToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	relic := b12Permanent(g, me.ID, "My Relic", "Artifact")
	victim := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	if _, err := castFromHandErr(g, me, "Final Vengeance", "Sorcery", "{B}", dskDGFinalVengeance,
		game.CastSpellParams{SacrificeIDs: []uuid.UUID{relic}, Targets: dskDGTarget(victim)}); err == nil {
		t.Error("an artifact paid the sacrifice")
	}
	if !g.Battlefield.Contains(relic) || !g.Battlefield.Contains(victim) {
		t.Error("the refused cast changed the battlefield")
	}
}

// --- Frantic Strength ------------------------------------------------------------------

func TestDskDGFranticStrengthHasFlashAndGivesTwoAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	// The table starts in the draw step: only flash lets the Aura in.
	if err := dskDGCastAtDrawStep(t, g, "Frantic Strength", "Enchantment — Aura", dskDGFranticOracle, bear); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("enchanted Bear %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOf(g, bear, "trample") {
		t.Error("the enchanted Bear lacks trample")
	}
}

// --- Friendly Ghost ---------------------------------------------------------------------

func TestDskDGFriendlyGhostPumpsATargetOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ghost := castFromHand(t, g, game.Card{
		Name: "Friendly Ghost", TypeLine: "Creature — Spirit", OracleID: dskDGGhostOracle,
		ManaCost: "{3}{W}", Power: 2, Toughness: 4,
	}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("the enters trigger did not ask for a target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 4/6", p, tough)
	}
	if !hasKeywordOf(g, ghost, "flying") {
		t.Error("Friendly Ghost does not have flying")
	}
}

// --- Friendly Teddy ---------------------------------------------------------------------

func TestDskDGFriendlyTeddyMakesEachPlayerDrawWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	teddy := b12Push(g, me.ID, "Friendly Teddy", "Artifact Creature — Bear Toy", dskDGTeddyOracle, 2, 2)
	before := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		before[i] = p.Hand.Size()
	}
	killCreature(t, g, me.ID, teddy)
	passPriorityAroundTable(t, g)
	for i, p := range g.Seats {
		if got := p.Hand.Size(); got != before[i]+1 {
			t.Errorf("seat %d hand %d, want %d", i, got, before[i]+1)
		}
	}
}

// --- Give In to Violence ----------------------------------------------------------------

func TestDskDGGiveInToViolencePumpsAndGrantsLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := dskDGCastAtDrawStep(t, g, "Give In to Violence", "Instant", dskDGGiveInOracle, bear); err != nil {
		t.Fatalf("cast in the draw step: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the Bear is %d/%d, want 4/4", p, tough)
	}
	if !hasKeywordOf(g, bear, "lifelink") {
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
	id := pushCatalogHandCard(me, name, typeLine, oracle)
	return g.CastSpell(me.ID, id, game.CastSpellParams{Targets: dskDGTarget(target)})
}
