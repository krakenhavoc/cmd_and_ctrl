package testsst

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stx_5502_b46_cards_test.go — batch 46 of Strixhaven: School of Mages
// (#5502): ten cards. Every test plays the real catalog entry through the
// engine.

const (
	stxB46ContainmentBreachOracle = "043dd2e5-21e4-49b9-90b2-831e587b2238"
	stxB46MakeYourMarkOracle      = "1c7ca399-7d85-452b-97c0-5d7f6c0eeb12"
	stxB46ReduceToMemoryOracle    = "661736eb-97d7-4b3c-ab18-708479a5fb29"
	stxB46ValentinOracle          = "6e1aa07a-5d9b-4daf-8dc4-c1717360855f"
	stxB46BiomathematicianOracle  = "70868667-a843-4954-863f-990854188c53"
	stxB46IlluminateHistoryOracle = "c7c1e7cf-0be2-4129-9bdf-0862be2fba7f"
	stxB46CulminationOracle       = "d4e124af-d335-4c42-b60a-fe578f69f7ed"
	stxB46ManifestationSageOracle = "d8bcd896-cae9-4733-a437-165644b8b0e8"
	stxB46IngeniousMasteryOracle  = "dad43a2d-5b30-45f2-a34d-9e68d1fbb716"
	stxB46DramaticFinaleOracle    = "e0155236-13c5-4e01-b4de-b8fa4af846d8"

	stxB46Bear = "Creature — Bear"
)

func stxB46Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, effectstest.Top100ActiveSeat(g), effectstest.Top100Opponent(g)
}

// stxB46Cast puts the card in p's hand, floats `pool` (exactly its cost, plus
// colorless for an announced X) and casts it strictly.
func stxB46Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	g.WithWriteLock(func() { p.Hand.PushTop(c) })
	effectstest.FloatForTest(g, p, pool)
	params.Strict = true
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

func stxB46MustCast(t *testing.T, g *game.Game, p *game.Player, c game.Card, pool string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := stxB46Cast(t, g, p, c, pool, params)
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func stxB46Card(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func stxB46Permanent(g *game.Game, owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost, Owner: owner, Controller: owner,
	})
}

func stxB46Tokens(g *game.Game, controller uuid.UUID, name string) int {
	n, _ := effectstest.CountTokensNamed(g, controller, name)
	return n
}

// --- every card is in the catalog ----------------------------------------

func TestStx5502B46EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		stxB46ContainmentBreachOracle, stxB46MakeYourMarkOracle, stxB46ReduceToMemoryOracle,
		stxB46ValentinOracle, stxB46ValentinOracle + "#1", stxB46BiomathematicianOracle,
		stxB46IlluminateHistoryOracle, stxB46CulminationOracle, stxB46ManifestationSageOracle,
		stxB46IngeniousMasteryOracle, stxB46DramaticFinaleOracle,
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

// --- Containment Breach -------------------------------------------------------

func stxB46Breach(t *testing.T, typeLine, cost string) (*game.Game, *game.Player, *game.Player, uuid.UUID, error) {
	t.Helper()
	g, me, opp := stxB46Table(t)
	target := stxB46Permanent(g, opp.ID, "Target Thing", typeLine, cost)
	_, err := stxB46Cast(t, g, me, game.Card{Name: "Containment Breach", TypeLine: "Sorcery — Lesson", ManaCost: "{2}{G}", OracleID: stxB46ContainmentBreachOracle},
		"CCG", game.CastSpellParams{Targets: stxB46Card(target)})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return g, me, opp, target, err
}

func TestStx5502B46ContainmentBreachGivesAPestForACheapArtifact(t *testing.T) {
	for _, tc := range []struct{ typeLine, cost string }{{"Artifact", "{1}"}, {"Enchantment", "{1}{W}"}, {"Artifact", ""}} {
		g, me, _, target, err := stxB46Breach(t, tc.typeLine, tc.cost)
		if err != nil {
			t.Fatalf("%s %q: %v", tc.typeLine, tc.cost, err)
		}
		if got := effectstest.B12ZoneOf(g, target); got != game.ZoneGraveyard {
			t.Errorf("%s %q is in %v, want the graveyard", tc.typeLine, tc.cost, got)
		}
		if got := stxB46Tokens(g, me.ID, "Pest"); got != 1 {
			t.Errorf("%s %q: %d Pests, want 1", tc.typeLine, tc.cost, got)
		}
	}
}

func TestStx5502B46ContainmentBreachGivesNoPestForAnExpensiveEnchantment(t *testing.T) {
	g, me, _, target, err := stxB46Breach(t, "Enchantment", "{2}{W}")
	if err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12ZoneOf(g, target); got != game.ZoneGraveyard {
		t.Errorf("the enchantment is in %v, want the graveyard", got)
	}
	if got := stxB46Tokens(g, me.ID, "Pest"); got != 0 {
		t.Errorf("%d Pests for a mana value 3 target, want 0", got)
	}
}

func TestStx5502B46ContainmentBreachRefusesACreature(t *testing.T) {
	_, _, _, _, err := stxB46Breach(t, stxB46Bear, "{1}")
	if err == nil {
		t.Error("a creature was a legal target")
	}
}

func TestStx5502B46ContainmentBreachPestGainsLifeWhenItDies(t *testing.T) {
	g, me, _, _, err := stxB46Breach(t, "Artifact", "{1}")
	if err != nil {
		t.Fatal(err)
	}
	pests := effectstest.TokensNamed(g, "Pest")
	if len(pests) != 1 {
		t.Fatalf("%d Pests, want 1", len(pests))
	}
	before := me.Life
	effectstest.KillCreature(t, g, me.ID, pests[0].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before+1 {
		t.Errorf("life %d after the Pest died, want %d", me.Life, before+1)
	}
}

// --- Make Your Mark -----------------------------------------------------------

func stxB46MakeYourMark(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID) {
	t.Helper()
	stxB46MustCast(t, g, me, game.Card{Name: "Make Your Mark", TypeLine: "Instant", ManaCost: "{R/W}", OracleID: stxB46MakeYourMarkOracle},
		"R", game.CastSpellParams{Targets: stxB46Card(target)})
	effectstest.PassPriorityAroundTable(t, g)
}

func TestStx5502B46MakeYourMarkPumpsAndMakesASpiritWhenItDies(t *testing.T) {
	g, me, opp := stxB46Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	stxB46MakeYourMark(t, g, me, theirs)
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 3 || th != 2 {
		t.Errorf("the target is %d/%d, want 3/2", p, th)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 0 {
		t.Fatalf("%d Spirits before the creature died, want 0", got)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("I have %d Spirits, want 1 (the spell's controller makes it)", got)
	}
	if got := stxB46Tokens(g, opp.ID, "Spirit"); got != 0 {
		t.Errorf("the creature's controller has %d Spirits, want 0", got)
	}
	spirits := effectstest.TokensNamed(g, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 3 || spirits[0].Toughness != 2 {
		t.Errorf("Spirits %+v, want one 3/2", spirits)
	}
}

func TestStx5502B46MakeYourMarkMakesNothingIfTheCreatureSurvives(t *testing.T) {
	g, me, _ := stxB46Table(t)
	mine := effectstest.B12Creature(g, me.ID, "Mine", stxB46Bear, 2, 2)
	stxB46MakeYourMark(t, g, me, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("%d Spirits though the creature lived, want 0", got)
	}
	// "This turn": a death in a later turn makes nothing.
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("%d Spirits for a death in a later turn, want 0", got)
	}
}

func TestStx5502B46MakeYourMarkRefusesALand(t *testing.T) {
	g, me, _ := stxB46Table(t)
	land := stxB46Permanent(g, me.ID, "Forest", "Basic Land — Forest", "")
	if _, err := stxB46Cast(t, g, me, game.Card{Name: "Make Your Mark", TypeLine: "Instant", ManaCost: "{R/W}", OracleID: stxB46MakeYourMarkOracle},
		"R", game.CastSpellParams{Targets: stxB46Card(land)}); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Reduce to Memory ---------------------------------------------------------

func stxB46Reduce(g *game.Game, t *testing.T, me *game.Player, target uuid.UUID) error {
	t.Helper()
	_, err := stxB46Cast(t, g, me, game.Card{Name: "Reduce to Memory", TypeLine: "Sorcery — Lesson", ManaCost: "{1}{W}{W}", OracleID: stxB46ReduceToMemoryOracle},
		"CWW", game.CastSpellParams{Targets: stxB46Card(target)})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func TestStx5502B46ReduceToMemoryExilesAndGivesItsControllerASpirit(t *testing.T) {
	g, me, opp := stxB46Table(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	if err := stxB46Reduce(g, t, me, theirs); err != nil {
		t.Fatal(err)
	}
	if got := effectstest.B12ZoneOf(g, theirs); got != game.ZoneExile {
		t.Errorf("the target is in %v, want exile", got)
	}
	if got := stxB46Tokens(g, opp.ID, "Spirit"); got != 1 {
		t.Errorf("its controller has %d Spirits, want 1", got)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("the caster has %d Spirits, want 0", got)
	}
	spirits := effectstest.TokensNamed(g, "Spirit")
	if len(spirits) != 1 || spirits[0].Power != 3 || spirits[0].Toughness != 2 {
		t.Errorf("Spirits %+v, want one 3/2", spirits)
	}
}

func TestStx5502B46ReduceToMemoryGivesYouTheSpiritForYourOwnPermanent(t *testing.T) {
	g, me, _ := stxB46Table(t)
	mine := stxB46Permanent(g, me.ID, "My Trinket", "Artifact", "{4}")
	if err := stxB46Reduce(g, t, me, mine); err != nil {
		t.Fatal(err)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("I have %d Spirits, want 1", got)
	}
}

func TestStx5502B46ReduceToMemoryRefusesALand(t *testing.T) {
	g, me, opp := stxB46Table(t)
	land := stxB46Permanent(g, opp.ID, "Forest", "Basic Land — Forest", "")
	if err := stxB46Reduce(g, t, me, land); err == nil {
		t.Error("a land was a legal target")
	}
}

// --- Valentin, Dean of the Vein // Lisette, Dean of the Root -----------------

func stxB46Valentin(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := stxB46Table(t)
	val := effectstest.B12Push(g, me.ID, "Valentin, Dean of the Vein", "Legendary Creature — Vampire Warlock", stxB46ValentinOracle, 1, 1)
	return g, me, opp, val
}

func TestStx5502B46ValentinHasMenaceAndLifelink(t *testing.T) {
	g, _, _, val := stxB46Valentin(t)
	for _, kw := range []string{"menace", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, val, kw) {
			t.Errorf("Valentin has no %s", kw)
		}
	}
}

func TestStx5502B46ValentinExilesAnOpponentsCreatureAndPayingMakesAPest(t *testing.T) {
	g, me, opp, _ := stxB46Valentin(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	if got := effectstest.B12ZoneOf(g, theirs); got != game.ZoneExile {
		t.Fatalf("the dying creature is in %v, want exile", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "C", 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Pest"); got != 1 {
		t.Errorf("%d Pests, want 1", got)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want the {2} spent", n)
	}
}

func TestStx5502B46ValentinDecliningTheOfferMakesNoPest(t *testing.T) {
	g, me, opp, _ := stxB46Valentin(t)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Pest"); got != 0 {
		t.Errorf("%d Pests after declining, want 0", got)
	}
	if got := effectstest.B12ZoneOf(g, theirs); got != game.ZoneExile {
		t.Errorf("the creature is in %v, want exile all the same", got)
	}
}

func TestStx5502B46ValentinLeavesTokensAndYourOwnCreaturesAlone(t *testing.T) {
	g, me, opp, _ := stxB46Valentin(t)
	token := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 white Spirit"))
	mine := effectstest.B12Creature(g, me.ID, "Mine", stxB46Bear, 2, 2)
	effectstest.KillCreature(t, g, opp.ID, token)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a token or my own creature dying offered the Pest")
	}
	if got := effectstest.B12ZoneOf(g, mine); got != game.ZoneGraveyard {
		t.Errorf("my creature is in %v, want the graveyard", got)
	}
}

func stxB46Lisette(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := stxB46Table(t)
	lis := effectstest.B12Push(g, me.ID, "Lisette, Dean of the Root", "Legendary Creature — Human Druid", stxB46ValentinOracle+"#1", 4, 4)
	return g, me, opp, lis
}

func TestStx5502B46LisettePaidCountersEveryCreatureYouControlAndGivesTrample(t *testing.T) {
	g, me, opp, lis := stxB46Lisette(t)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxB46Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	effectstest.AfrB07Gain(g, me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.FillPoolColored(me, "C", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{lis, bear} {
		if got := effectstest.PlusOneCounters(g, id); got != 1 {
			t.Errorf("%s has %d counters, want 1", id, got)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("%s has no trample", id)
		}
	}
	if got := effectstest.PlusOneCounters(g, theirs); got != 0 {
		t.Errorf("an opponent's creature has %d counters, want 0", got)
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature gained trample")
	}
}

func TestStx5502B46LisetteDecliningDoesNothing(t *testing.T) {
	g, me, _, lis := stxB46Lisette(t)
	effectstest.AfrB07Gain(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, lis); got != 0 {
		t.Errorf("%d counters after declining, want 0", got)
	}
}

func TestStx5502B46LisetteIgnoresLifeLostAndOpponentsLifeGain(t *testing.T) {
	g, me, opp, _ := stxB46Lisette(t)
	effectstest.AfrB07Gain(g, me, -2)
	effectstest.AfrB07Gain(g, opp, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("losing life or an opponent gaining life triggered Lisette")
	}
}

// --- Biomathematician and Manifestation Sage ------------------------------------

// stxB46Creature casts a creature spell from hand with an explicit P/T and
// resolves it.
func stxB46Creature(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle string, pool string) {
	t.Helper()
	stxB46MustCast(t, g, me, game.Card{Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Power: 2, Toughness: 2}, pool, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

func stxB46FractalToken() game.Card { return effects.TokenCard("0/0 green and blue Fractal") }

func TestStx5502B46BiomathematicianPutsACounterOnEachFractalYouControl(t *testing.T) {
	g, me, opp := stxB46Table(t)
	old := effectstest.PushToken(g, me.ID, stxB46FractalToken())
	effectstest.AddCounters(t, g, old, game.CounterPlusOne, 2)
	theirs := effectstest.PushToken(g, opp.ID, stxB46FractalToken())
	effectstest.AddCounters(t, g, theirs, game.CounterPlusOne, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxB46Bear, 2, 2)
	stxB46Creature(t, g, me, "Biomathematician", "Creature — Human Wizard", "{1}{G}{U}", stxB46BiomathematicianOracle, "CGU")
	if got := effectstest.PlusOneCounters(g, old); got != 3 {
		t.Errorf("my old Fractal has %d counters, want 3", got)
	}
	if got := effectstest.PlusOneCounters(g, theirs); got != 2 {
		t.Errorf("an opponent's Fractal has %d counters, want 2", got)
	}
	if got := effectstest.PlusOneCounters(g, bear); got != 0 {
		t.Errorf("a non-Fractal has %d counters, want 0", got)
	}
	fractals := 0
	for _, c := range effectstest.TokensNamed(g, "Fractal") {
		if c.Controller == me.ID && c.InstanceID != old {
			fractals++
			if got := effectstest.PlusOneCounters(g, c.InstanceID); got != 1 {
				t.Errorf("the new Fractal has %d counters, want 1", got)
			}
		}
	}
	if fractals != 1 {
		t.Errorf("%d new Fractals, want 1", fractals)
	}
}

func TestStx5502B46ManifestationSageCountsTheCardsInYourHand(t *testing.T) {
	g, me, _ := stxB46Table(t)
	me.Hand.Cards = nil
	effectstest.FillHand(me, "H1", "H2", "H3")
	stxB46Creature(t, g, me, "Manifestation Sage", "Creature — Human Wizard", "{G/U}{G/U}{G/U}{G/U}", stxB46ManifestationSageOracle, "GGGG")
	fractals := effectstest.TokensNamed(g, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("%d Fractals, want 1", len(fractals))
	}
	if got := effectstest.PlusOneCounters(g, fractals[0].InstanceID); got != 3 {
		t.Errorf("the Fractal has %d counters, want 3 (the cards left in hand)", got)
	}
}

func TestStx5502B46ManifestationSageWithAnEmptyHandMakesADeadFractal(t *testing.T) {
	g, me, _ := stxB46Table(t)
	me.Hand.Cards = nil
	stxB46Creature(t, g, me, "Manifestation Sage", "Creature — Human Wizard", "{G/U}{G/U}{G/U}{G/U}", stxB46ManifestationSageOracle, "GGUU")
	if got := len(effectstest.TokensNamed(g, "Fractal")); got != 0 {
		t.Errorf("%d Fractals survive an empty hand, want 0 (a 0/0 dies)", got)
	}
}

// --- Illuminate History -------------------------------------------------------

func stxB46Illuminate(t *testing.T, graveyard, hand int) (*game.Game, *game.Player) {
	t.Helper()
	g, me, _ := stxB46Table(t)
	for i := 0; i < graveyard; i++ {
		effectstest.GyCard(me, "Old", "Instant", "{R}", "")
	}
	me.Hand.Cards = nil
	effectstest.FillHand(me, []string{"H1", "H2", "H3", "H4"}[:hand]...)
	stxB46MustCast(t, g, me, game.Card{Name: "Illuminate History", TypeLine: "Sorcery — Lesson", ManaCost: "{2}{R}{R}", OracleID: stxB46IlluminateHistoryOracle},
		"CCRR", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	return g, me
}

func TestStx5502B46IlluminateHistoryDiscardsDrawsThatManyAndMakesASpiritWithSevenInGraveyard(t *testing.T) {
	g, me := stxB46Illuminate(t, 4, 3)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (discarded 3, drew 3)", got)
	}
	if got := me.Graveyard.Size(); got != 8 {
		t.Errorf("graveyard %d, want 8 (seven and the spell itself)", got)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirits with seven cards in the graveyard, want 1", got)
	}
}

func TestStx5502B46IlluminateHistoryMakesNoSpiritWithSixInGraveyard(t *testing.T) {
	g, me := stxB46Illuminate(t, 4, 2)
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Graveyard.Size(); got != 7 {
		t.Fatalf("graveyard %d, want 7 (six and the spell itself)", got)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 0 {
		t.Errorf("%d Spirits with six cards in the graveyard, want 0 (the spell itself is still on the stack)", got)
	}
}

func TestStx5502B46IlluminateHistoryAnyNumberIncludesNone(t *testing.T) {
	g, me := stxB46Illuminate(t, 7, 3)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != 3 {
		t.Errorf("hand %d, want 3 (nothing discarded, nothing drawn)", got)
	}
	if got := stxB46Tokens(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirits with seven already in the graveyard, want 1", got)
	}
}

// --- Culmination of Studies -----------------------------------------------------

func stxB46Culmination(t *testing.T, x int, top ...game.Card) (*game.Game, *game.Player, *game.Player, int) {
	t.Helper()
	g, me, opp := stxB46Table(t)
	for _, c := range top {
		effectstest.SeedLibraryTop(me, c.Name, c.TypeLine, c.ManaCost)
	}
	before := me.Hand.Size()
	stxB46MustCast(t, g, me, game.Card{Name: "Culmination of Studies", TypeLine: "Sorcery", ManaCost: "{X}{U}{R}", OracleID: stxB46CulminationOracle},
		strings.Repeat("C", x)+"UR", game.CastSpellParams{XValue: x})
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, opp, before
}

func TestStx5502B46CulminationPaysForLandsBlueAndRedCards(t *testing.T) {
	// Seeded bottom to top: the top three cards are the Island, Blue Bear and Red Bear.
	g, me, _, before := stxB46Culmination(t, 3,
		game.Card{Name: "Deep Card", TypeLine: stxB46Bear, ManaCost: "{G}"},
		game.Card{Name: "Red Bear", TypeLine: stxB46Bear, ManaCost: "{1}{R}"},
		game.Card{Name: "Blue Bear", TypeLine: stxB46Bear, ManaCost: "{1}{U}"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 1 {
		t.Errorf("%d Treasures, want 1 (one land card)", got)
	}
	if got := g.Exile.Size(); got != 3 {
		t.Errorf("%d cards exiled, want 3", got)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (one blue card)", got, before+1)
	}
	for _, s := range g.Seats {
		if s.ID == me.ID {
			continue
		}
		if s.Life != 39 {
			t.Errorf("%s has %d life, want 39 (one red card)", s.Name, s.Life)
		}
	}
	if me.Life != 40 {
		t.Errorf("I have %d life, want 40", me.Life)
	}
}

func TestStx5502B46CulminationCountsAGoldCardForBothColours(t *testing.T) {
	g, me, opp, before := stxB46Culmination(t, 2,
		game.Card{Name: "Deep Card", TypeLine: stxB46Bear, ManaCost: "{G}"},
		game.Card{Name: "Gold Bear", TypeLine: stxB46Bear, ManaCost: "{U}{R}"},
		game.Card{Name: "Red Bear 2", TypeLine: stxB46Bear, ManaCost: "{R}"},
	)
	if opp.Life != 38 {
		t.Errorf("an opponent has %d life, want 38 (two red cards)", opp.Life)
	}
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (one blue card)", got, before+1)
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("%d Treasures with no land card exiled, want 0", got)
	}
}

func TestStx5502B46CulminationWithXZeroExilesNothing(t *testing.T) {
	g, me, opp, _ := stxB46Culmination(t, 0, game.Card{Name: "Top", TypeLine: "Basic Land — Island"})
	if got := g.Exile.Size(); got != 0 {
		t.Errorf("%d cards exiled for X=0, want 0", got)
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 0 || opp.Life != 40 {
		t.Errorf("X=0 did something: %d Treasures, opponent life %d", got, opp.Life)
	}
}

// --- Ingenious Mastery ------------------------------------------------------------

func TestStx5502B46IngeniousMasteryPaidAlternativeCostDrawsThreeAndGivesAnOpponentTreasures(t *testing.T) {
	g, me, opp := stxB46Table(t)
	before := me.Hand.Size()
	stxB46MustCast(t, g, me, game.Card{Name: "Ingenious Mastery", TypeLine: "Sorcery", ManaCost: "{X}{2}{U}", OracleID: stxB46IngeniousMasteryOracle},
		"CCU", game.CastSpellParams{AlternativeCost: "ingenious-mastery"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+3 {
		t.Errorf("hand %d, want %d (three drawn)", got, before+3)
	}
	if got := effectstest.CountTreasuresControlledBy(g, opp.ID); got != 2 {
		t.Errorf("the chosen opponent has %d Treasures, want 2", got)
	}
	if got := effectstest.CountTreasuresControlledBy(g, me.ID); got != 0 {
		t.Errorf("I have %d Treasures, want 0", got)
	}
	for _, s := range g.Seats {
		if s.ID != me.ID && s.ID != opp.ID && effectstest.CountTreasuresControlledBy(g, s.ID) != 0 {
			t.Errorf("%s, not chosen, has Treasures", s.Name)
		}
	}
}

func TestStx5502B46IngeniousMasteryFullCostDrawsXAndGivesNothing(t *testing.T) {
	g, me, opp := stxB46Table(t)
	before := me.Hand.Size()
	stxB46MustCast(t, g, me, game.Card{Name: "Ingenious Mastery", TypeLine: "Sorcery", ManaCost: "{X}{2}{U}", OracleID: stxB46IngeniousMasteryOracle},
		"CCCCU", game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("an opponent was asked for without the alternative cost")
	}
	if got := me.Hand.Size(); got != before+2 {
		t.Errorf("hand %d, want %d (X = 2 drawn)", got, before+2)
	}
	if got := effectstest.CountTreasuresControlledBy(g, opp.ID); got != 0 {
		t.Errorf("an opponent has %d Treasures, want 0", got)
	}
}

func TestStx5502B46IngeniousMasteryRefusesAShortPool(t *testing.T) {
	g, me, _ := stxB46Table(t)
	if _, err := stxB46Cast(t, g, me, game.Card{Name: "Ingenious Mastery", TypeLine: "Sorcery", ManaCost: "{X}{2}{U}", OracleID: stxB46IngeniousMasteryOracle},
		"CU", game.CastSpellParams{XValue: 0}); err == nil {
		t.Error("cast for {1}{U} without the alternative cost")
	}
}

// --- Dramatic Finale ------------------------------------------------------------

func stxB46Finale(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g, me, opp := stxB46Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Dramatic Finale", "Enchantment", stxB46DramaticFinaleOracle, false)
	return g, me, opp
}

func TestStx5502B46DramaticFinalePumpsOnlyYourCreatureTokens(t *testing.T) {
	g, me, opp := stxB46Finale(t)
	mine := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Spirit"))
	theirs := effectstest.PushToken(g, opp.ID, effects.TokenCard("1/1 white Spirit"))
	bear := effectstest.B12Creature(g, me.ID, "Bear", stxB46Bear, 2, 2)
	if p, th := effectstest.EffectivePTOf(t, g, mine); p != 2 || th != 2 {
		t.Errorf("my token is %d/%d, want 2/2", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 1 || th != 1 {
		t.Errorf("an opponent's token is %d/%d, want 1/1", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 2 || th != 2 {
		t.Errorf("my nontoken creature is %d/%d, want 2/2", p, th)
	}
}

func TestStx5502B46DramaticFinaleMakesOneInklingPerTurnWhenNontokenCreaturesDie(t *testing.T) {
	g, me, _ := stxB46Finale(t)
	a := effectstest.B12Creature(g, me.ID, "A", stxB46Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", stxB46Bear, 2, 2)
	effectstest.KillCreature(t, g, me.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Inkling"); got != 1 {
		t.Fatalf("%d Inklings after a nontoken creature died, want 1", got)
	}
	inklings := effectstest.TokensNamed(g, "Inkling")
	if len(inklings) != 1 || inklings[0].Power != 2 || inklings[0].Toughness != 1 {
		t.Errorf("Inklings %+v, want one printed 2/1", inklings)
	}
	if p, th := effectstest.EffectivePTOf(t, g, inklings[0].InstanceID); p != 3 || th != 2 {
		t.Errorf("the Inkling is %d/%d with the static, want 3/2", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, inklings[0].InstanceID, "flying") {
		t.Error("the Inkling has no flying")
	}
	effectstest.KillCreature(t, g, me.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Inkling"); got != 1 {
		t.Errorf("%d Inklings after a second death this turn, want still 1", got)
	}
}

func TestStx5502B46DramaticFinaleIgnoresTokensAndOpponentsCreatures(t *testing.T) {
	g, me, opp := stxB46Finale(t)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Spirit"))
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", stxB46Bear, 2, 2)
	effectstest.KillCreature(t, g, me.ID, token)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := stxB46Tokens(g, me.ID, "Inkling"); got != 0 {
		t.Errorf("%d Inklings for a token or an opponent's creature dying, want 0", got)
	}
}
