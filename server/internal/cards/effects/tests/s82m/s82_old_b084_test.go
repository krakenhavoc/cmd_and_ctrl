package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b084_test.go — Commander Legends equipment, S82 OLD batch 084
// (#5507). Each test plays the real catalog entry and asserts what its text says.

const (
	oldB084GreatswordOracle = "14a19102-00ed-4425-8837-1d05bc7ad8ba"
	oldB084ValkasOracle     = "412f57e8-76af-4636-bab0-e0940300ee2e"
	oldB084ArmoryOracle     = "7ac1aab0-9d2d-4fb7-b0bb-0af3e6d665db"
	oldB084ArmamentsOracle  = "c2042b38-1f06-44d4-b56f-2cb4d9197712"
	oldB084ThuneOracle      = "f72e8ccb-a6fb-4161-bfa1-69c94cc5d49a"
	oldB084AxeOracle        = "fe340efa-6b69-4ccc-a13c-3688f1c44c29"
)

func TestOldB084EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB084GreatswordOracle: "Seraphic Greatsword",
		oldB084ValkasOracle:     "Ring of Valkas",
		oldB084ArmoryOracle:     "Armory of Iroas",
		oldB084ArmamentsOracle:  "Angelic Armaments",
		oldB084ThuneOracle:      "Ring of Thune",
		oldB084AxeOracle:        "Amorphous Axe",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

func oldB084Equip(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	eq := effectstest.PushCatalogPermanent(g, owner, name, "Artifact — Equipment", oracle, false)
	effectstest.BrAttach(t, g, eq, host)
	return eq
}

func oldB084Colored(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Colors = colors
			}
		}
	})
}

func oldB084Counters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.LayeredCard(t, g, id).Counters[game.CounterPlusOne]
}

func oldB084SetLife(g *game.Game, p *game.Player, life int) {
	g.WithWriteLock(func() { p.Life = life })
}

// --- Seraphic Greatsword ----------------------------------------------------

func TestOldB084GreatswordPumpsAndMakesAnAngelAttackingTheLifeLeader(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	oldB084Equip(t, g, me.ID, "Seraphic Greatsword", oldB084GreatswordOracle, host)
	if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tough != 4 {
		t.Errorf("equipped creature %d/%d, want 4/4", p, tough)
	}
	oldB084SetLife(g, me, 30)
	oldB084SetLife(g, g.Seats[2], 30)
	oldB084SetLife(g, g.Seats[3], 30)
	oldB084SetLife(g, opp, 40)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Angel", me.ID); n != 1 {
		t.Fatalf("%d Angel tokens, want 1", n)
	}
	var angel game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Angel" && c.Controller == me.ID {
			angel = c
		}
	}
	if !angel.Tapped || angel.AttackingTarget != opp.ID {
		t.Errorf("Angel tapped=%v attacking=%v, want tapped and attacking the life leader", angel.Tapped, angel.AttackingTarget)
	}
	if p := effectstest.EffectivePower(t, g, angel.InstanceID); p != 4 {
		t.Errorf("Angel power %d, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, angel.InstanceID, "flying") {
		t.Error("Angel lacks flying")
	}
}

func TestOldB084GreatswordTiedForMostLifeCounts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	oldB084Equip(t, g, me.ID, "Seraphic Greatsword", oldB084GreatswordOracle, host)
	oldB084SetLife(g, me, 10)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Angel", me.ID); n != 1 {
		t.Errorf("%d Angel tokens attacking a player tied for most life, want 1", n)
	}
}

func TestOldB084GreatswordDoesNothingAgainstAPlayerWhoDoesNotLead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	oldB084Equip(t, g, me.ID, "Seraphic Greatsword", oldB084GreatswordOracle, host)
	oldB084SetLife(g, opp, 12)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Angel", me.ID); n != 0 {
		t.Errorf("%d Angel tokens against a player who trails, want 0", n)
	}
}

func TestOldB084GreatswordIgnoresAnAttackerItDoesNotEquip(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 2, 2)
	oldB084Equip(t, g, me.ID, "Seraphic Greatsword", oldB084GreatswordOracle, host)
	effectstest.DeclareAttack(t, g, opp.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Angel", me.ID); n != 0 {
		t.Errorf("%d Angel tokens from an unequipped attacker, want 0", n)
	}
}

// --- Armory of Iroas --------------------------------------------------------

func TestOldB084ArmoryPutsACounterOnTheEquippedAttackerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 2, 2)
	oldB084Equip(t, g, me.ID, "Armory of Iroas", oldB084ArmoryOracle, host)
	effectstest.DeclareAttack(t, g, opp.ID, host, other)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB084Counters(t, g, host); got != 1 {
		t.Errorf("equipped attacker has %d counters, want 1", got)
	}
	if got := oldB084Counters(t, g, other); got != 0 {
		t.Errorf("unequipped attacker has %d counters, want 0", got)
	}
}

func TestOldB084ArmoryDoesNothingWhenUnattached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	effectstest.PushCatalogPermanent(g, me.ID, "Armory of Iroas", "Artifact — Equipment", oldB084ArmoryOracle, false)
	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB084Counters(t, g, host); got != 0 {
		t.Errorf("%d counters from an unattached Armory, want 0", got)
	}
}

// --- Ring of Valkas / Ring of Thune ----------------------------------------

func oldB084RingUpkeep(t *testing.T, name, oracle, keyword, color, offColor string) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	match := effectstest.B12Creature(g, me.ID, "Matching", "Creature — Bear", 2, 2)
	off := effectstest.B12Creature(g, me.ID, "Off", "Creature — Bear", 2, 2)
	oldB084Colored(g, match, color)
	oldB084Colored(g, off, offColor)
	ring := oldB084Equip(t, g, me.ID, name, oracle, match)
	if !effectstest.HasEffectiveKeyword(t, g, match, keyword) {
		t.Errorf("equipped creature lacks %s", keyword)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB084Counters(t, g, match); got != 1 {
		t.Errorf("%s: matching colour has %d counters, want 1", name, got)
	}
	// Move it to a creature of the wrong colour: nothing.
	effectstest.BrAttach(t, g, ring, off)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB084Counters(t, g, off); got != 0 {
		t.Errorf("%s: wrong colour has %d counters, want 0", name, got)
	}
}

func TestOldB084RingOfValkasGivesHasteAndGrowsARedHostEachUpkeep(t *testing.T) {
	oldB084RingUpkeep(t, "Ring of Valkas", oldB084ValkasOracle, "haste", "R", "W")
}

func TestOldB084RingOfThuneGivesVigilanceAndGrowsAWhiteHostEachUpkeep(t *testing.T) {
	oldB084RingUpkeep(t, "Ring of Thune", oldB084ThuneOracle, "vigilance", "W", "R")
}

func TestOldB084RingsOnlyTriggerOnYourOwnUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2)
	oldB084Colored(g, host, "R")
	oldB084Equip(t, g, me.ID, "Ring of Valkas", oldB084ValkasOracle, host)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB084Counters(t, g, host); got != 0 {
		t.Errorf("%d counters on an opponent's upkeep, want 0", got)
	}
}

// --- Amorphous Axe ----------------------------------------------------------

func TestOldB084AxePumpsPowerAndMakesTheCreatureEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	axe := oldB084Equip(t, g, me.ID, "Amorphous Axe", oldB084AxeOracle, host)
	if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 5 || tough != 2 {
		t.Errorf("equipped creature %d/%d, want 5/2", p, tough)
	}
	if !effectstest.LayeredCard(t, g, host).Effective().AllCreatureTypes {
		t.Error("equipped creature is not every creature type")
	}
	_ = axe
	other := effectstest.B12Creature(g, me.ID, "Cub", "Creature — Bear", 2, 2)
	if effectstest.LayeredCard(t, g, other).Effective().AllCreatureTypes {
		t.Error("an unequipped creature is every creature type")
	}
}

// --- Angelic Armaments ------------------------------------------------------

func TestOldB084ArmamentsAddsFlyingAngelAndWhiteToTheHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	oldB084Colored(g, host, "G")
	oldB084Equip(t, g, me.ID, "Angelic Armaments", oldB084ArmamentsOracle, host)
	if p, tough := effectstest.EffectivePower(t, g, host), effectstest.EffectiveToughness(t, g, host); p != 4 || tough != 4 {
		t.Errorf("equipped creature %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, host, "flying") {
		t.Error("equipped creature lacks flying")
	}
	eff := effectstest.LayeredCard(t, g, host).Effective()
	subs := map[string]bool{}
	for _, s := range eff.Subtypes {
		subs[s] = true
	}
	if !subs["Angel"] || !subs["Bear"] {
		t.Errorf("subtypes %v, want Angel and Bear", eff.Subtypes)
	}
	cols := map[string]bool{}
	for _, c := range effectstest.EffectiveColors(t, g, host) {
		cols[c] = true
	}
	if !cols["W"] || !cols["G"] {
		t.Errorf("colours %v, want white in addition to green", cols)
	}
}

var _ = effects.Lookup
