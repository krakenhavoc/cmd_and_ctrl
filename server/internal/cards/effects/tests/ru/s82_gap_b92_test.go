package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB92Death    = "f2bf5014-677a-4880-b926-d93457348bdc"
	gapB92Toymaker = "8e93ea21-5e3f-448e-a175-96c8438c328e"
	gapB92Blink    = "f5fba11b-7669-40a0-97c7-6cb7c13acedd"
	gapB92City     = "313c6b29-c488-4b4b-af16-f82717f59a5d"
	gapB92Fugitive = "cacda1c2-2fe8-40a3-89f3-38a835f10af3"
	gapB92Devils   = "09d7a2d5-9e6e-4d84-8f97-c36b11ccc6fb"
	gapB92Bigger   = "8f351961-69e3-4637-b010-d13f19daca39"
	gapB92Fenric   = "ef7fce1b-0d95-4f92-a8ad-09d9756fa0e4"
	gapB92Moon     = "7ec3aaa4-8b6a-4046-b599-654f9844ca6c"
	gapB92Return   = "be2b3954-e4d7-438c-a549-252bd7f4a9bd"
	gapB92Flux     = "8beaaf55-e05d-49c3-9320-c222c4d8ad12"
)

var gapB92Names = map[string]string{
	gapB92Blink:    "Blink",
	gapB92City:     "City of Death",
	gapB92Fugitive: "Fugitive of the Judoon",
	gapB92Return:   "Return the Past",
	gapB92Flux:     "The Flux",
}

func TestGapB92CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range gapB92Names {
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

// gapB92Saga puts a Saga with `lore` lore counters on p's side; the next
// precombat main adds one and triggers that chapter.
func gapB92Saga(g *game.Game, owner uuid.UUID, name, oracle string, lore int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Enchantment — Saga", OracleID: oracle,
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLore: lore},
	})
}

func gapB92Count(g *game.Game, owner uuid.UUID, subtype string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype(subtype) {
			n++
		}
	}
	return n
}

func TestGapB92CityOfDeathMakesATreasureThenCopiesATokenEachChapter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB92Saga(g, me.ID, "City of Death", gapB92City, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB92Count(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("Treasures after chapter I: %d, want 1", got)
	}
}

func TestGapB92CityOfDeathChapterTwoCopiesANonSagaTokenOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	goblin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	nonToken := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	gapB92Saga(g, me.ID, "City of Death", gapB92City, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: nonToken}); err == nil {
			t.Error("a nontoken creature was a legal target")
		}
		effectstest.PickCard(t, g, me.ID, goblin)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB92Count(g, me.ID, "Goblin"); got != 2 {
		t.Errorf("Goblins %d after chapter II, want 2", got)
	}
}

func TestGapB92FluxDealsFourThenExilesTheTopCardsAndAddsRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 4, 4)
	gapB92Saga(g, me.ID, "The Flux", gapB92Flux, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) {
		t.Error("the 4/4 survived 4 damage")
	}
}

func TestGapB92FluxChapterTwoExilesTheTopCardAndLetsYouPlayIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	top := effectstest.PutOnTopForTest(me, "Mountain", "Basic Land — Mountain")
	gapB92Saga(g, me.ID, "The Flux", gapB92Flux, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	if p := effectstest.GrantedPermissionOn(g, me.ID, top, game.ZoneExile); !p.Granted() {
		t.Errorf("the exiled card has no permission to be played: %+v", p)
	}
}

func TestGapB92FluxChapterSixAddsSixRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB92Saga(g, me.ID, "The Flux", gapB92Flux, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 6 {
		t.Errorf("mana pool holds %d, want 6", got)
	}
}

func TestGapB92ReturnThePastGivesFlashbackOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Return the Past", "Enchantment", gapB92Return, false)
	sorcery := effectstest.SeedGraveyardCard(t, g, "Dead Sorcery", "Sorcery", effectstest.CsTestGraveyardSorceryOracl)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	})
	perm := effectstest.GrantedPermissionOn(g, me.ID, sorcery, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" || !perm.ExileOnResolution {
		t.Fatalf("the sorcery's granted flashback = %+v, want flashback for its mana cost, exiled after", perm)
	}
	if p := effectstest.GrantedPermissionOn(g, me.ID, bear, game.ZoneGraveyard); p.Granted() {
		t.Error("a creature card gained flashback")
	}
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback the sorcery on my turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(sorcery) {
		t.Error("a sorcery cast with flashback is exiled as it leaves the stack")
	}
	// On an opponent's turn the card has no flashback.
	again := effectstest.SeedGraveyardCard(t, g, "Another Sorcery", "Sorcery", effectstest.CsTestGraveyardSorceryOracl)
	effectstest.EndTurn(t, g)
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("the turn did not pass to another player")
	}
	if err := g.CastSpell(me.ID, again, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err == nil {
		t.Error("cast with flashback during an opponent's turn")
	}
}

func TestGapB92BlinkShufflesTheTargetAwayAndItsOwnerInvestigatesThenTheAngelAppears(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	gapB92Saga(g, me.ID, "Blink", gapB92Blink, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !opp.Library.Contains(foe) {
		t.Error("the creature was not shuffled into its owner's library")
	}
	if got := gapB92Count(g, opp.ID, "Clue"); got != 1 {
		t.Errorf("the owner has %d Clues, want 1", got)
	}
	if got := gapB92Count(g, me.ID, "Clue"); got != 0 {
		t.Errorf("the caster got %d Clues", got)
	}
}

func TestGapB92BlinkChapterTwoMakesAnAlienAngel(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB92Saga(g, me.ID, "Blink", gapB92Blink, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB92Count(g, me.ID, "Angel"); got != 1 {
		t.Errorf("Alien Angels %d, want 1", got)
	}
}

func TestGapB92FugitiveMakesTwoTokensInvestigatesAndFetchesADoctor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB92Saga(g, me.ID, "Fugitive of the Judoon", gapB92Fugitive, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if gapB92Count(g, me.ID, "Human") != 1 || gapB92Count(g, me.ID, "Rhino") != 1 {
		t.Fatalf("tokens: Humans %d, Rhinos %d; want 1 and 1", gapB92Count(g, me.ID, "Human"), gapB92Count(g, me.ID, "Rhino"))
	}
}

func TestGapB92FugitiveChapterTwoInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gapB92Saga(g, me.ID, "Fugitive of the Judoon", gapB92Fugitive, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB92Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("Clues %d, want 1", got)
	}
}

func TestGapB92FugitiveChapterThreeExilesAHumanAndAnArtifactForADoctor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	human := effectstest.B12Creature(g, me.ID, "Human", "Creature — Human", 1, 1)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	other := effectstest.B12Creature(g, me.ID, "Goat", "Creature — Goat", 1, 1)
	doctor := game.Card{InstanceID: uuid.New(), Name: "The Doctor", TypeLine: "Legendary Creature — Time Lord Doctor", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Library.PushTop(doctor) })
	gapB92Saga(g, me.ID, "Fugitive of the Judoon", gapB92Fugitive, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The Human prompt offers only Humans; the artifact prompt only artifacts.
	c := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if c == nil {
		t.Fatalf("no own-permanents prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{other}); err == nil {
		t.Error("a non-Human was accepted for the Human prompt")
	}
	effectstest.AnswerOwnPermanents(t, g, me.ID, human)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(human) || !g.Exile.Contains(rock) {
		t.Error("the Human and the artifact were not exiled")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, doctor.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(doctor.InstanceID) {
		t.Error("the Doctor card was not put onto the battlefield")
	}
}
