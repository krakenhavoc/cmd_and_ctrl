package effects

import (
	"math/rand/v2"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cards_test.go exercises each registered catalog card end-to-end:
// seed a game, cast the card targeting an appropriate victim, pass
// priority around, assert the resulting state + graveyard. Grouped
// by family to keep the per-card helpers small.

// newCatalogGame returns a fully-started 4-player game (four seats
// so target-rotation tests have room). 20-card libraries give room
// for mill / discard / draw primitives without bottoming out.
func newCatalogGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i := 0; i < 4; i++ {
		deck := make([]game.Card, 20)
		for j := range deck {
			deck[j] = game.NewCard("basic-filler", uuid.Nil)
		}
		if _, err := g.AddPlayer(string(rune('A'+i)), deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.Start(rand.New(rand.NewPCG(11, 22))); err != nil {
		t.Fatalf("Start: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

// castCatalogSpell seeds a card with the given name, type line, and
// oracle ID into the active seat's hand, advances to the precombat
// main phase, then calls CastSpell with the supplied targets. Pass
// priority around afterwards to resolve.
func castCatalogSpell(t *testing.T, g *game.Game, name, typeLine, oracleID string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		OracleID:   oracleID,
		Owner:      active.ID,
		Controller: active.ID,
	})
	// Advance to main phase if needed.
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// passPriorityAroundTable passes priority until the stack is empty
// (bounded at 16 to catch runaway loops).
func passPriorityAroundTable(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		if g.Stack.Size() == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("stack did not empty after 16 priority passes")
}

func pushCreatureToBattlefieldForTest(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	return id
}

// --- Direct damage ---------------------------------------------

func TestLightningBoltDamagesPlayer(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	beforeLife := target.Life

	castCatalogSpell(t, g, "Lightning Bolt", "Instant",
		"4457ed35-7c10-48c8-9776-456485fdf070",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeLife-3 {
		t.Errorf("target life: got %d, want %d", got, beforeLife-3)
	}
}

func TestShockDamagesPlayer(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	beforeLife := target.Life

	castCatalogSpell(t, g, "Shock", "Instant",
		"a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeLife-2 {
		t.Errorf("target life: got %d, want %d", got, beforeLife-2)
	}
}

func TestLightningHelixDamagesAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	beforeTargetLife := target.Life
	beforeCasterLife := caster.Life

	castCatalogSpell(t, g, "Lightning Helix", "Instant",
		"800c258a-cfc4-4a54-a667-065ea8dea69e",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Life; got != beforeTargetLife-3 {
		t.Errorf("target life: got %d, want %d", got, beforeTargetLife-3)
	}
	if got := caster.Life; got != beforeCasterLife+3 {
		t.Errorf("caster life: got %d, want %d", got, beforeCasterLife+3)
	}
}

func TestPyroclasmDamagesEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Creature A")
	c2 := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Creature B")

	castCatalogSpell(t, g, "Pyroclasm", "Sorcery",
		"e4bcd4ea-e7cd-4471-8f3b-18bb51d3d70c",
		nil,
	)
	passPriorityAroundTable(t, g)

	// Each 2/2 took 2 damage → lethal → SBA routes to graveyard.
	if g.Battlefield.Contains(c1) {
		t.Errorf("creature A should have been destroyed")
	}
	if g.Battlefield.Contains(c2) {
		t.Errorf("creature B should have been destroyed")
	}
}

// --- Mass removal ----------------------------------------------

func TestWrathOfGodDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "Mine")
	c2 := pushCreatureToBattlefieldForTest(g, g.Seats[1].ID, "Theirs")

	castCatalogSpell(t, g, "Wrath of God", "Sorcery",
		"34515b16-c9a4-4f98-8c77-416a7a523407",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) || g.Battlefield.Contains(c2) {
		t.Errorf("Wrath did not clear the battlefield")
	}
}

func TestDamnationDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "One")

	castCatalogSpell(t, g, "Damnation", "Sorcery",
		"d57a8f0b-7989-4db5-8756-6f2690097252",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) {
		t.Errorf("Damnation did not destroy creature")
	}
}

func TestDayOfJudgmentDestroysAllCreatures(t *testing.T) {
	g := newCatalogGame(t)
	c1 := pushCreatureToBattlefieldForTest(g, g.Seats[0].ID, "One")

	castCatalogSpell(t, g, "Day of Judgment", "Sorcery",
		"d057289d-5e28-43d5-8ff3-4a1bc723477d",
		nil,
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(c1) {
		t.Errorf("Day of Judgment did not destroy creature")
	}
}

// --- Counter magic ---------------------------------------------

// TestCounterspellCountersLightningBolt is the multi-spell scenario
// that proves the counter chain works: opponent casts Bolt; caster
// casts Counterspell targeting it; priority wraps and resolves
// Counterspell first (top of stack), which pulls Bolt off the stack
// into opponent's graveyard without damage.
func TestCounterspellCountersLightningBolt(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]
	beforeLife := caster.Life

	// Advance to main phase so the active seat can cast a sorcery /
	// non-instant. Lightning Bolt is an instant, so opponent can
	// cast it at instant speed during the active player's priority
	// — but castCatalogSpell seeds into the active seat's hand. For
	// this test we'll skip that helper and place the Bolt manually
	// in opponent's hand + cast it.

	// Put us in main phase.
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Seed + cast opponent's Lightning Bolt. Since opponent isn't
	// the active seat, they'd normally need priority — which they
	// can hold as an instant. But CastSpell applies the sorcery-
	// speed gate to non-instants only, and Bolt is an Instant, so
	// this works.
	boltID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: boltID,
		Name:       "Lightning Bolt",
		TypeLine:   "Instant",
		OracleID:   "4457ed35-7c10-48c8-9776-456485fdf070",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent CastSpell Bolt: %v", err)
	}

	// Active seat casts Counterspell targeting the Bolt stack item.
	counterID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: counterID,
		Name:       "Counterspell",
		TypeLine:   "Instant",
		OracleID:   "cc187110-1148-4090-bbb8-e205694a39f5",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, counterID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: boltID}},
	}); err != nil {
		t.Fatalf("Caster CastSpell Counterspell: %v", err)
	}

	// Pass priority around — Counterspell resolves first (top of
	// stack), pulling Bolt to opponent's graveyard. Then Bolt would
	// try to resolve but is no longer on the stack.
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(boltID) || g.Stack.Contains(boltID) {
		t.Errorf("Bolt did not leave the stack")
	}
	if !opponent.Graveyard.Contains(boltID) {
		t.Errorf("countered Bolt not in opponent graveyard")
	}
	if !caster.Graveyard.Contains(counterID) {
		t.Errorf("Counterspell not in caster graveyard")
	}
	// Life total unchanged — Bolt never dealt damage.
	if caster.Life != beforeLife {
		t.Errorf("caster life changed: got %d, want %d", caster.Life, beforeLife)
	}
}

func TestNegateCountersSpell(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Opponent casts something (use Shock — any spell targeting
	// Negate tests the counter independent of the noncreature
	// predicate that Negate doesn't enforce in S14).
	shockID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shockID,
		Name:       "Shock",
		TypeLine:   "Instant",
		OracleID:   "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shockID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent Shock: %v", err)
	}

	// Caster Negates.
	negateID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: negateID,
		Name:       "Negate",
		TypeLine:   "Instant",
		OracleID:   "3407fe41-fdd3-4119-8f70-4bc4590a379f",
		Owner:      caster.ID,
		Controller: caster.ID,
	})
	if err := g.CastSpell(caster.ID, negateID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shockID}},
	}); err != nil {
		t.Fatalf("Caster Negate: %v", err)
	}

	passPriorityAroundTable(t, g)

	if !opponent.Graveyard.Contains(shockID) {
		t.Errorf("countered Shock not in opponent graveyard")
	}
}

// TestSwanSongCountersAndCreatesBirdToken proves the two-step
// composition (CounterTarget → CreateToken) + the "token under the
// countered spell's controller" routing.
func TestSwanSongCountersAndCreatesBirdToken(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}

	// Opponent casts Shock (a valid Swan Song target: instants are
	// enchantment/instant/sorcery).
	shockID := uuid.New()
	opponent.Hand.PushTop(game.Card{
		InstanceID: shockID,
		Name:       "Shock",
		TypeLine:   "Instant",
		OracleID:   "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	if err := g.CastSpell(opponent.ID, shockID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: caster.ID}},
	}); err != nil {
		t.Fatalf("Opponent Shock: %v", err)
	}

	// Caster Swan Songs.
	swanID := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: swanID,
		Name:       "Swan Song",
		TypeLine:   "Instant",
		OracleID:   "8ddfc283-c9b4-41a5-af88-cf0068e986cc",
		Owner:      caster.ID,
		Controller: caster.ID,
	})

	bfBefore := len(g.Battlefield.Cards)

	if err := g.CastSpell(caster.ID, swanID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: shockID}},
	}); err != nil {
		t.Fatalf("Caster Swan Song: %v", err)
	}
	passPriorityAroundTable(t, g)

	// Shock countered into opponent's graveyard.
	if !opponent.Graveyard.Contains(shockID) {
		t.Errorf("countered Shock not in opponent graveyard")
	}
	// One new token on the battlefield under opponent (spell
	// controller) control.
	bfDelta := len(g.Battlefield.Cards) - bfBefore
	if bfDelta != 1 {
		t.Errorf("battlefield delta: got %d, want 1 (Bird token)", bfDelta)
	}
	var tokenCard game.Card
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" {
			tokenCard = c
			found = true
		}
	}
	if !found {
		t.Fatalf("bird token not on battlefield")
	}
	if tokenCard.Controller != opponent.ID {
		t.Errorf("token controller: got %v, want %v (countered spell's controller)", tokenCard.Controller, opponent.ID)
	}
	if tokenCard.Power != 2 || tokenCard.Toughness != 2 {
		t.Errorf("token stats: got %d/%d, want 2/2", tokenCard.Power, tokenCard.Toughness)
	}
}

// --- Opt-in canary ---------------------------------------------

// --- Draw + mill + discard (sub-PR 5) --------------------------

func TestDivinationDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Divination", "Sorcery",
		"273b339c-964b-4a18-8eb5-ceb8abcdfd9e",
		nil,
	)
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 2 {
		t.Errorf("hand delta: got %d, want 2 (Divination spell is now in graveyard, net +2)", got)
	}
}

func TestHarmonizeDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	handBefore := caster.Hand.Size()

	castCatalogSpell(t, g, "Harmonize", "Sorcery",
		"7eff84f1-f772-497a-b350-bbc93d0230f7",
		nil,
	)
	passPriorityAroundTable(t, g)

	if got := caster.Hand.Size() - handBefore; got != 3 {
		t.Errorf("hand delta: got %d, want 3", got)
	}
}

func TestSignInBloodDrawsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	targetHand := target.Hand.Size()
	targetLife := target.Life

	castCatalogSpell(t, g, "Sign in Blood", "Sorcery",
		"c6207f6a-a624-4754-88f5-dbe700c841ff",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := target.Hand.Size() - targetHand; got != 2 {
		t.Errorf("target hand delta: got %d, want 2", got)
	}
	if target.Life != targetLife-2 {
		t.Errorf("target life: got %d, want %d", target.Life, targetLife-2)
	}
}

func TestGlimpseMillsTen(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	libBefore := target.Library.Size()
	gyBefore := target.Graveyard.Size()

	castCatalogSpell(t, g, "Glimpse the Unthinkable", "Sorcery",
		"552f0163-a19d-4671-888f-044fc0354875",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := libBefore - target.Library.Size(); got != 10 {
		t.Errorf("library delta: got %d, want 10", got)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 10 {
		t.Errorf("graveyard delta: got %d, want 10", got)
	}
}

func TestMindRotDiscardsTwo(t *testing.T) {
	g := newCatalogGame(t)
	target := g.Seats[1]
	// Seed extra hand cards so we have at least 2 to discard.
	pushHandCard(g, target)
	pushHandCard(g, target)
	handBefore := target.Hand.Size()
	gyBefore := target.Graveyard.Size()

	castCatalogSpell(t, g, "Mind Rot", "Sorcery",
		"ad44cf74-b717-48fb-9fa2-77512024d76a",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := handBefore - target.Hand.Size(); got != 2 {
		t.Errorf("hand delta: got %d, want 2", got)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 2 {
		t.Errorf("graveyard delta: got %d, want 2", got)
	}
}

func TestThoughtseizeRevealsAndDiscards(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	handBefore := target.Hand.Size()
	gyBefore := target.Graveyard.Size()
	casterLife := caster.Life

	castCatalogSpell(t, g, "Thoughtseize", "Sorcery",
		"edd8d1e8-be43-4c38-bb3a-83081fbaf0b5",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if got := handBefore - target.Hand.Size(); got != 1 {
		t.Errorf("target hand delta: got %d, want 1", got)
	}
	if got := target.Graveyard.Size() - gyBefore; got != 1 {
		t.Errorf("target graveyard delta: got %d, want 1", got)
	}
	if caster.Life != casterLife-2 {
		t.Errorf("caster life: got %d, want %d", caster.Life, casterLife-2)
	}
	// Remaining hand cards should be known to the caster post-reveal.
	// (Sticky knowledge — S13.5 doesn't drop it after discard.)
	for _, c := range target.Hand.Cards {
		if !c.IsKnownTo(caster.ID) {
			t.Errorf("caster not a knower of %v after Thoughtseize reveal", c.InstanceID)
		}
	}
}

// --- Targeted removal ------------------------------------------

func TestSwordsToPlowsharesExilesAndGainsLife(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	opponent := g.Seats[1]
	// Put a 4/4 on opponent's battlefield.
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Big Thing",
		TypeLine:   "Creature — Giant",
		Power:      4,
		Toughness:  4,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	oppLifeBefore := opponent.Life

	castCatalogSpell(t, g, "Swords to Plowshares", "Instant",
		"b1544f21-7e98-461b-aed5-e748b0168c52",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(creatureID) {
		t.Errorf("creature not in exile")
	}
	if opponent.Life != oppLifeBefore+4 {
		t.Errorf("opponent life: got %d, want %d", opponent.Life, oppLifeBefore+4)
	}
	_ = caster
}

func TestPathToExileExilesAndFetchesBasic(t *testing.T) {
	g := newCatalogGame(t)
	opponent := g.Seats[1]
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Beater",
		TypeLine:   "Creature — Beast",
		Power:      3,
		Toughness:  3,
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})
	// Seed a Forest in opponent's library so Path's search has
	// something to find.
	forestID := uuid.New()
	opponent.Library.PushTop(game.Card{
		InstanceID: forestID,
		Name:       "Forest",
		TypeLine:   "Basic Land — Forest",
		Owner:      opponent.ID,
		Controller: opponent.ID,
	})

	castCatalogSpell(t, g, "Path to Exile", "Instant",
		"d683d985-9888-4d21-8b5f-69e69ce4a03b",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if !g.Exile.Contains(creatureID) {
		t.Errorf("creature not exiled")
	}
	if !g.Battlefield.Contains(forestID) {
		t.Errorf("fetched Forest not on battlefield")
	}
}

func TestUnsummonBouncesToHand(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	creatureID := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: creatureID,
		Name:       "Bouncable",
		TypeLine:   "Creature — Bird",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})

	castCatalogSpell(t, g, "Unsummon", "Instant",
		"837182db-1bf3-4a2c-bd01-1af9d9873561",
		[]game.TargetRef{{Kind: game.TargetCard, ID: creatureID}},
	)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(creatureID) {
		t.Errorf("creature still on battlefield")
	}
	if !owner.Hand.Contains(creatureID) {
		t.Errorf("creature not in owner's hand")
	}
}

// TestNonCatalogSpellStaysSandbox proves a spell without a
// catalog-matching OracleID resolves to graveyard with no effect —
// the Cockatrice-style manual fallback is intact. Canary for
// future regressions that might accidentally activate auto-resolve
// on non-catalog cards.
func TestNonCatalogSpellStaysSandbox(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[0]
	target := g.Seats[1]
	beforeLife := target.Life

	spellID := castCatalogSpell(t, g, "Random Spell", "Instant",
		"", // empty oracle ID → cannot match the catalog
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: target.ID}},
	)
	passPriorityAroundTable(t, g)

	if target.Life != beforeLife {
		t.Errorf("non-catalog spell applied damage: target life %d, want %d", target.Life, beforeLife)
	}
	if !caster.Graveyard.Contains(spellID) {
		t.Errorf("non-catalog spell did not route to graveyard")
	}
}
