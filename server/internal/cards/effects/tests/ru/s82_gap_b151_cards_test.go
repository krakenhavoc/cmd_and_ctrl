package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b151_cards_test.go — GAP batch 151, Jumpstart 2022 creatures
// (#5948). Every card is played through the engine.

const (
	gap151WickedGuardianOracle    = "dda70e25-b3c5-444d-9dff-1366771bf881"
	gap151ZaskOracle              = "3c4966b9-26f9-46ff-8d41-7fd01c70986d"
	gap151BlisterpodOracle        = "2a62b226-d317-4212-9781-9c1d72056f11"
	gap151BroodMonitorOracle      = "9a08bb4b-46df-4b99-8f86-670d07e900ba"
	gap151EngulfingSlagwurmOracle = "0340cfcb-fc96-40d2-8175-31a0093615bb"
)

func gap151Seats(g *game.Game) (me, opp *game.Player) {
	seat := g.Turn.ActiveSeat
	return g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
}

func gap151Scions(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Eldrazi Scion" && c.IsToken() {
			n++
		}
	}
	return n
}

func TestGap151EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		gap151WickedGuardianOracle, gap151ZaskOracle, gap151BlisterpodOracle,
		gap151BroodMonitorOracle, gap151EngulfingSlagwurmOracle,
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

func TestGap151DevoidCardsAreColorless(t *testing.T) {
	for _, tc := range []struct{ name, typeLine, oracle string }{
		{"Blisterpod", "Creature — Eldrazi Drone", gap151BlisterpodOracle},
		{"Brood Monitor", "Creature — Eldrazi Drone", gap151BroodMonitorOracle},
	} {
		g := effectstest.NewCatalogGame(t)
		me, _ := gap151Seats(g)
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), game.KeywordDevoid) {
			t.Errorf("%s has no devoid", tc.name)
		}
	}
}

func TestGap151BlisterpodLeavesAScionWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	pod := effectstest.B12Push(g, me.ID, "Blisterpod", "Creature — Eldrazi Drone", gap151BlisterpodOracle, 1, 1)
	effectstest.B27Kill(g, pod)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap151Scions(g, me.ID); got != 1 {
		t.Errorf("%d Eldrazi Scion tokens, want 1", got)
	}
}

func TestGap151BroodMonitorMakesThreeScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	effectstest.CastHoldCreature(t, g, "Brood Monitor", "Creature — Eldrazi Drone", gap151BroodMonitorOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if got := gap151Scions(g, me.ID); got != 3 {
		t.Errorf("%d Eldrazi Scion tokens, want 3", got)
	}
}

func gap151OwnPick(t *testing.T, g *game.Game, me *game.Player, picks ...uuid.UUID) {
	t.Helper()
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents && pc.Chooser == me.ID {
			if err := g.ResolveOwnPermanents(pc.ID, me.ID, picks); err != nil {
				t.Fatalf("ResolveOwnPermanents: %v", err)
			}
			return
		}
	}
	t.Fatal("no own_permanents prompt is open")
}

func TestGap151WickedGuardianDamagesAnotherCreatureAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 5)
	hand := me.Hand.Size()
	guardian := effectstest.CastHoldCreature(t, g, "Wicked Guardian", "Creature — Human Noble", gap151WickedGuardianOracle, 4, 2)
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceOwnPermanents {
			for _, id := range pc.ChooseCards {
				if id == guardian {
					t.Fatal("the Guardian itself was offered")
				}
			}
		}
	}
	gap151OwnPick(t, g, me, ally)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ally).DamageMarked; got != 2 {
		t.Errorf("the ally has %d damage, want 2", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestGap151WickedGuardianMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 5)
	hand := me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Wicked Guardian", "Creature — Human Noble", gap151WickedGuardianOracle, 4, 2)
	gap151OwnPick(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindCardAnywhere(t, g, ally).DamageMarked; got != 0 {
		t.Errorf("the ally has %d damage after declining", got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("hand %d after declining, want %d", got, hand)
	}
}

func TestGap151EngulfingSlagwurmDestroysWhatBlocksOrIsBlockedByIt(t *testing.T) {
	// Becomes blocked.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Engulfing Slagwurm", "Creature — Wurm", gap151EngulfingSlagwurmOracle, 7, 7)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 4)
	life := me.Life
	effectstest.BrAttack(t, g, wurm)
	if err := g.DeclareBlocker(blocker, wurm); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("the blocker survived")
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}

	// Blocks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	wurm = effectstest.B12Push(g, opp.ID, "Engulfing Slagwurm", "Creature — Wurm", gap151EngulfingSlagwurmOracle, 7, 7)
	life = opp.Life
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(wurm, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacker survived")
	}
	if opp.Life != life+3 {
		t.Errorf("life %d, want %d", opp.Life, life+3)
	}
}

func TestGap151ZaskPlaysLandsAndCastsInsectsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	effectstest.B12Push(g, me.ID, "Zask, Skittering Swarmlord", "Legendary Creature — Insect", gap151ZaskOracle, 5, 5)
	spec, _ := effects.Lookup(gap151ZaskOracle)
	var land, insect bool
	for _, p := range spec.CastPermissions {
		if p.Zone == game.ZoneGraveyard && p.Filter.LandsOnly {
			land = true
		}
		if p.Zone == game.ZoneGraveyard && p.Filter.Subtype == "Insect" {
			insect = true
		}
	}
	if !land || !insect {
		t.Fatalf("graveyard permissions: lands=%v insects=%v", land, insect)
	}
	effectstest.AdvanceToMain(t, g)
	gyInsect := effectstest.PushGraveyardCardTyped(me, "Graveyard Insect", "Creature — Insect")
	gyBear := effectstest.PushGraveyardCardTyped(me, "Graveyard Bear", "Creature — Bear")
	if err := g.CastSpell(me.ID, gyBear, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Error("cast a non-Insect creature from the graveyard")
	}
	if err := g.CastSpell(me.ID, gyInsect, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err != nil {
		t.Fatalf("casting an Insect from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gyInsect) {
		t.Error("the Insect did not resolve onto the battlefield")
	}
}

func TestGap151ZaskTucksAnotherInsectAndMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := gap151Seats(g)
	effectstest.B12Push(g, me.ID, "Zask, Skittering Swarmlord", "Legendary Creature — Insect", gap151ZaskOracle, 5, 5)
	bug := effectstest.B12Creature(g, me.ID, "Bug", "Creature — Insect", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	library := me.Library.Size()
	// A non-Insect does nothing.
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != library {
		t.Fatalf("a non-Insect dying changed the library (%d)", me.Library.Size())
	}
	effectstest.B27Kill(g, bug)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(bug) {
		t.Error("the Insect stayed in the graveyard")
	}
	// +1 for the tucked card, -2 for the mill.
	if got := me.Library.Size(); got != library-1 {
		t.Errorf("library %d, want %d", got, library-1)
	}
	if got := me.Library.Cards[0].InstanceID; got != bug {
		t.Errorf("the bottom card is %s, want the Insect", got)
	}
}

func TestGap151ZaskPumpGivesAnInsectPowerAndDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := gap151Seats(g)
	zask := effectstest.B12Push(g, me.ID, "Zask, Skittering Swarmlord", "Legendary Creature — Insect", gap151ZaskOracle, 5, 5)
	bug := effectstest.B12Creature(g, me.ID, "Bug", "Creature — Insect", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.ApaMana(me, "G", "C")
	if err := g.ActivateCatalogAbility(me.ID, zask, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	}); err == nil {
		t.Fatal("targeted a non-Insect")
	}
	effectstest.ApaActivate(t, g, me, zask, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bug}}})
	if p := effectstest.EffectivePower(t, g, bug); p != 2 {
		t.Errorf("power %d, want 2", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bug, "deathtouch") {
		t.Error("no deathtouch")
	}
}
