package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- the probe: one trigger per player -----------------------------

// TestB784KeeperOfFablesDrawsOncePerPlayerHit is the issue's first
// case. Three non-Humans, three opponents, one damage step: three
// draws, one per player.
func TestB784KeeperOfFablesDrawsOncePerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	keeper := effectstest.B12Push(g, me.ID, "Keeper of Fables", "Creature — Cat", effectstest.B30KeeperOfFablesOracle, 4, 5)
	a := effectstest.B12Creature(g, me.ID, "Cat A", "Creature — Cat", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Cat B", "Creature — Cat", 2, 2)
	c := effectstest.B12Creature(g, me.ID, "Cat C", "Creature — Cat", 2, 2)
	hand := me.Hand.Size()

	effectstest.B784AttackEach(t, g,
		[2]uuid.UUID{a, g.Seats[1].ID},
		[2]uuid.UUID{b, g.Seats[2].ID},
		[2]uuid.UUID{c, g.Seats[3].ID})

	if n := effectstest.B784Triggers(g, keeper); n != 3 {
		t.Fatalf("three opponents hit at once: %d triggers, want 3 (CR 603.2c)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d cards, want 3 — one per player the non-Humans hit", got)
	}
}

// TestB784KeeperOfFablesDrawsOnceForTwoCreaturesOnOnePlayer is the
// half that must not move: the batch still collapses, so two
// creatures connecting with ONE opponent are one draw.
func TestB784KeeperOfFablesDrawsOnceForTwoCreaturesOnOnePlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := effectstest.B12Push(g, me.ID, "Keeper of Fables", "Creature — Cat", effectstest.B30KeeperOfFablesOracle, 4, 5)
	a := effectstest.B12Creature(g, me.ID, "Cat A", "Creature — Cat", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Cat B", "Creature — Cat", 2, 2)
	hand := me.Hand.Size()

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, opp.ID}, [2]uuid.UUID{b, opp.ID})

	if n := effectstest.B784Triggers(g, keeper); n != 1 {
		t.Fatalf("two creatures on one opponent: %d triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

// TestB784FaceBreakerMakesATreasurePerStepAndPlayer is the issue's
// second case, and the CR 510.4 half of the fix: a combat with first
// strike in it has TWO combat damage steps, which are two batches
// even though the turn cursor sees one step. A first-striker and a
// regular attacker on the same opponent are two Treasures; a third
// attacker on a second opponent in the regular step is a third.
//
// #914 moved where the FIRST of the three is counted, not how many
// there are. b784AttackEach walks the cursor with advance_step, and a
// step can no longer end owing what is on its stack (CR 117.4), so the
// first-strike step's trigger resolves inside that step — one Treasure
// already on the battlefield by the time the cursor reaches the
// regular step, and two triggers left waiting there. Three either way.
func TestB784FaceBreakerMakesATreasurePerStepAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	breaker := effectstest.B12Push(g, me.ID, "Professional Face-Breaker", "Creature — Human Warrior",
		effectstest.ProfessionalFaceBreakerOracle, 2, 3)
	striker := effectstest.PushTribalCreature(g, me.ID, "Striker", "Creature — Knight", 2, 2, "first strike")
	regular := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)

	effectstest.B784AttackEach(t, g,
		[2]uuid.UUID{striker, g.Seats[1].ID},
		[2]uuid.UUID{regular, g.Seats[1].ID},
		[2]uuid.UUID{other, g.Seats[2].ID})

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 1 {
		t.Fatalf("%d Treasures on entering the regular step, want 1 — (first strike, A) resolved in its own step", got)
	}
	if n := effectstest.B784Triggers(g, breaker); n != 2 {
		t.Fatalf("%d triggers, want 2 — (regular, A) and (regular, B)", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 3 {
		t.Errorf("%d Treasures, want 3 — one per (damage step, player) pair", got)
	}
}

// --- one assertion per converted card ------------------------------

// TestB784GrazilaxxDrawsPerPlayerHit — Grazilaxx, Illithid Scholar.
func TestB784GrazilaxxDrawsPerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	graz := effectstest.B12Push(g, me.ID, "Grazilaxx, Illithid Scholar", "Legendary Creature — Horror", effectstest.B18GrazilaxxOracle, 3, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[2].ID})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 — one per player hit", got)
	}
	_ = graz
}

// TestB784RapaciousGuestMakesAFoodPerPlayerHit — Rapacious Guest.
func TestB784RapaciousGuestMakesAFoodPerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Rapacious Guest", "Creature — Halfling Citizen", effectstest.B31RapaciousGuestOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[2].ID})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Food"); got != 2 {
		t.Errorf("%d Foods, want 2 — one per player hit", got)
	}
}

// TestB784ThopterSpyNetworkDrawsPerPlayerHit — Thopter Spy Network.
func TestB784ThopterSpyNetworkDrawsPerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Thopter Spy Network", "Enchantment", effectstest.B08ThopterSpyNetworkOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Myr A", "Artifact Creature — Myr", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Myr B", "Artifact Creature — Myr", 2, 2)
	hand := me.Hand.Size()

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[2].ID})
	effectstest.PassPriorityAroundTable(t, g)

	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2 — one per player the artifact creatures hit", got)
	}
}

// TestB784OliviaMakesATreasurePerPlayerHit — Olivia, Opulent Outlaw,
// whose declared caveat this retires.
func TestB784OliviaMakesATreasurePerPlayerHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Olivia, Opulent Outlaw", "Legendary Creature — Vampire Assassin", effectstest.OliviaOpulentOutlawOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Rogue A", "Creature — Human Rogue", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Pirate B", "Creature — Human Pirate", 2, 2)

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[2].ID})
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Treasure"); got != 2 {
		t.Errorf("%d Treasures, want 2 — one per player the outlaws hit", got)
	}
	if spec, _ := effects.Lookup(effectstest.OliviaOpulentOutlawOracle); spec.Completeness != effects.CompletenessFull {
		t.Error("the once-per-damage-step caveat is retired")
	}
}

// TestB784AlelaGoadsForEachPlayerHerFaeriesHit — Alela, Cunning
// Conqueror. Her target clause is per player ("target creature THAT
// PLAYER controls"), so the second player hit is a second trigger
// with its own pick.
func TestB784AlelaGoadsForEachPlayerHerFaeriesHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	alela := effectstest.B12Push(g, me.ID, "Alela, Cunning Conqueror", "Legendary Creature — Faerie Warlock", effectstest.B33AlelaOracle, 2, 4)
	a := effectstest.B12Creature(g, me.ID, "Faerie A", "Creature — Faerie Rogue", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Faerie B", "Creature — Faerie Rogue", 1, 1)
	effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[2].ID, "Other Bear", "Creature — Bear", 2, 2)

	effectstest.B784AttackEach(t, g, [2]uuid.UUID{a, g.Seats[1].ID}, [2]uuid.UUID{b, g.Seats[2].ID})

	if n := effectstest.B784Triggers(g, alela) + b784PickPrompts(g, me.ID, alela); n != 2 {
		t.Errorf("%d triggers for two players hit, want 2 — one goad each", n)
	}
}

// b784PickPrompts counts the open pick_target prompts `chooser` owes
// for triggers from `source` — a targeted trigger waits on its pick
// before it reaches PendingTriggers.
func b784PickPrompts(g *game.Game, chooser, source uuid.UUID) int {
	n := 0
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget && c.Chooser == chooser && c.Source == source {
			n++
		}
	}
	return n
}

// --- "whenever you attack a player" --------------------------------

// TestB784HorizonExplorerMakesALanderPerPlayerAttacked — "Horizon
// Explorer's last ability will trigger once for each player you
// attack" (ruling). Its declared caveat said one Lander for a
// two-player attack; this is the caveat retired.
func TestB784HorizonExplorerMakesALanderPerPlayerAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	explorer := effectstest.B12Push(g, me.ID, "Horizon Explorer", "Creature — Insect Scout", effectstest.B16HorizonExplorerOracle, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(explorer, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(bear, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Lander"); got != 2 {
		t.Errorf("%d Landers, want 2 — one per player attacked", got)
	}
	spec, _ := effects.Lookup(effectstest.B16HorizonExplorerOracle)
	for _, cav := range spec.Caveats {
		if cav == "Attacking two players at once makes one Lander, not two." {
			t.Error("the per-player caveat is retired")
		}
	}
}

// TestB784NeyaliExilesPerPlayerAttackedWithTokens — "The second
// ability of Neyali, Suns' Vanguard triggers for each player you are
// attacking with one or more tokens" (ruling).
func TestB784NeyaliExilesPerPlayerAttackedWithTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	neyali := effectstest.B12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", effectstest.B32NeyaliOracle, 3, 3)
	a := effectstest.B32Token(g, me.ID, "Soldier A", 1, 1)
	b := effectstest.B32Token(g, me.ID, "Soldier B", 1, 1)

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)

	if n := effectstest.B784Triggers(g, neyali); n != 2 {
		t.Errorf("%d triggers for two players attacked with tokens, want 2", n)
	}
}

// TestB784BreenaTriggersForBothOpponentsOfASplitAttack — Breena, the
// Demagogue, whose per-opponent dedup was the last caller of the
// deleted TriggerInFlightForEffect. The declared retreat that went
// with it — a second opponent swallowed while the first's creature
// pick was open — is gone with it.
func TestB784BreenaTriggersForBothOpponentsOfASplitAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second, poorest := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	breena := effectstest.B12Push(g, me.ID, "Breena, the Demagogue", "Legendary Creature — Bird Warlock", effectstest.B17BreenaOracle, 1, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	// Both attacked opponents are ahead of the fourth seat, so the
	// intervening if (CR 603.4) holds for each of them.
	first.Life, second.Life, poorest.Life = 40, 40, 20

	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(a, first.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(b, second.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)

	if n := effectstest.B784Triggers(g, breena) + b784PickPrompts(g, me.ID, breena); n != 2 {
		t.Errorf("%d triggers for a split attack on two opponents, want 2 — one per opponent", n)
	}
}
