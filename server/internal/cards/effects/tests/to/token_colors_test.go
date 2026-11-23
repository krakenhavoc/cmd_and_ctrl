package teststo

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// token_colors_test.go — #1127. Twenty-one token templates declared no
// colour where the printed token is coloured, which is a RULES bug and
// not an art one: colour is a characteristic the engine reads. These
// four tests are the behavioural half of the fix — one per colour
// family corrected — and each one is about a colour-sensitive effect
// meeting a token through the card that makes it, not about the
// literal in tokens_table.go. Flip a template back to colourless and
// the effect stops seeing it, which is exactly what these fail on.
//
// The literal half is two other tests: TestTokenTableRowsAreWellFormed
// pins key-against-Colors hermetically, and
// TestEveryTokenTemplateMatchesAPrintedToken (dump-gated) pins every
// row against the printed token card.

// White — Elspeth, Sun's Champion's Soldiers are white, so Mass
// Calcify ("destroy all nonwhite creatures") spares them. A colourless
// Soldier was destroyed by its own controller's sweeper.
func TestElspethsSoldiersAreWhiteAndSurviveMassCalcify(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	pw := effectstest.PushWalkerForTest(g, me.ID, "Elspeth, Sun's Champion", effectstest.ElspethSunsChampionOracle, 4)
	colourless := effectstest.B12Creature(g, opp.ID, "Ornithopter", "Artifact Creature — Thopter", 0, 2)
	effectstest.AdvanceToMainOf(t, g, seat)

	if err := g.ActivateCatalogAbility(me.ID, pw, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("Elspeth +1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountNamed(g, "Soldier"); got != 3 {
		t.Fatalf("Soldiers after +1 = %d, want 3", got)
	}

	effectstest.CastCatalogSpell(t, g, "Mass Calcify", "Sorcery", effectstest.B40MassCalcifyOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountNamed(g, "Soldier"); got != 3 {
		t.Errorf("Soldiers after Mass Calcify = %d, want 3 — a white Soldier is not a nonwhite creature", got)
	}
	if g.Battlefield.Contains(colourless) {
		t.Error("the colourless creature should have been destroyed, so the sweeper did run")
	}
}

// Black — Ophiomancer's Snake is black, so Doom Blade ("destroy target
// nonblack creature") cannot even be pointed at it. A colourless Snake
// was a legal target for every nonblack removal spell in the format.
func TestOphiomancersSnakeIsBlackAndDoomBladeCannotTargetIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Ophiomancer", "Creature — Human Shaman", effectstest.B05OphiomancerOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	snake := effectstest.FindBattlefieldByName(g, "Snake")
	if snake == uuid.Nil {
		t.Fatal("Ophiomancer made no Snake")
	}

	caster := g.Seats[g.Turn.ActiveSeat]
	legal := effectstest.LegalCards(g, caster.ID, effectstest.DoomBladeOracle)
	if legal[snake] {
		t.Error("a black Snake must not be a legal Doom Blade target")
	}
	if !legal[bear] {
		t.Fatal("the colourless Bear should be legal, so the predicate is live")
	}
	err := effectstest.CastCatalogSpellErr(t, g, "Doom Blade", "Instant", effectstest.DoomBladeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: snake}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("Doom Blade at the Snake: got %v, want ErrIllegalTarget", err)
	}
}

// Green — Beast Within's Beast is green, so its controller's Sylvan
// Anthem ("green creatures you control get +1/+1", and a scry when
// another green creature enters) both pumps it and triggers on it.
func TestBeastWithinsBeastIsGreenAndSylvanAnthemSeesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Sylvan Anthem", "Enchantment", effectstest.B26SylvanAnthemOracle, 0, 0)
	mine := effectstest.SeedPermanentFor(g, me.ID, "Sol Ring", "Artifact")

	effectstest.CastCatalogSpell(t, g, "Beast Within", "Instant", effectstest.BeastWithinOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)

	beast := effectstest.FindBattlefieldByName(g, "Beast")
	if beast == uuid.Nil {
		t.Fatal("Beast Within made no Beast")
	}
	if p := effectstest.EffectivePower(t, g, beast); p != 4 {
		t.Errorf("Beast power under Sylvan Anthem = %d, want 4 — a 3/3 GREEN Beast gets the +1/+1", p)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("a green creature entering under your control scries 1")
	} else {
		effectstest.B26AnswerScryAllTop(t, g, me.ID)
	}
}

// Blue — Swan Song's Bird is blue, so Pyroblast's "destroy target
// permanent if it's blue" destroys it. Pyroblast reads the colour on
// RESOLUTION, so a colourless Bird survived a spell aimed straight at
// it.
func TestSwanSongsBirdIsBlueAndPyroblastDestroysIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, opponent := g.Seats[0], g.Seats[1]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// The opponent casts a Shock; the caster Swan Songs it, so the
	// BIRD lands under the opponent's control (the countered spell's
	// controller gets it).
	shock := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shock, Name: "Shock", TypeLine: "Instant",
		OracleID: "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner:    opponent.ID, Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shock, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("opponent Shock: %v", err)
	}
	swan := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: swan, Name: "Swan Song", TypeLine: "Instant",
		OracleID: swanSongOracle, Owner: caster.ID, Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, swan, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shock}},
	}); err != nil {
		t.Fatalf("Swan Song: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	bird := effectstest.FindBattlefieldByName(g, "Bird")
	if bird == uuid.Nil {
		t.Fatal("Swan Song made no Bird")
	}
	effectstest.CastModal(t, g, "Pyroblast", "Instant", effectstest.B03PyroblastOracle, []int{1},
		[]game.TargetRef{{Kind: game.TargetCard, ID: bird}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bird) {
		t.Error("a blue Bird is destroyed by Pyroblast's second mode")
	}
}

const swanSongOracle = "8ddfc283-c9b4-41a5-af88-cf0068e986cc"
