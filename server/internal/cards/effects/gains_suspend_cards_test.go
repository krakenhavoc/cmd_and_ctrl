package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gains_suspend_cards_test.go — "If it doesn't have suspend, it gains
// suspend" (#4455): Kang Prime and Doom's Time Platform, on the engine's
// Card.SuspendGranted marker. Each test plays the real catalog entries.

const (
	gsKangOracle     = "c8e0b90c-edfb-4024-9ed5-398966f4ac72"
	gsPlatformOracle = "78019626-1154-483f-b862-73982f9ddc7f"
	gsRiftBoltOracle = "2b8afa9f-4236-4c02-a8d5-3c145caecfd6"
)

func TestGainsSuspendCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{gsKangOracle: "Kang Prime", gsPlatformOracle: "Doom's Time Platform"} {
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

func gsKangCard() game.Card {
	return game.Card{Name: "Kang Prime", TypeLine: "Legendary Creature — Human Villain", OracleID: gsKangOracle,
		ManaCost: "{3}{U}{B}", Power: 3, Toughness: 5}
}

// gsSetLibrary replaces the player's library with `cards`, the first listed on top.
func gsSetLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func gsForest() game.Card { return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"} }

func gsGoblin() game.Card {
	return game.Card{Name: "Test Goblin", TypeLine: "Creature — Goblin", ManaCost: "{2}{R}", Power: 2, Toughness: 2}
}

func gsExiled(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

// gsNextUpkeep walks the table to the named seat's NEXT upkeep and
// resolves what the step queued, stopping at a prompt.
func gsNextUpkeep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToStepOf(t, g, seat, game.StepUpkeep)
	passPriorityAroundTable(t, g)
}

// Kang Prime entering exiles past the lands to the nonland card, puts two
// time counters on it and gives it suspend; the lands stay exiled and plain.
// Then over two of its owner's upkeeps the card is offered a free cast, and
// the creature that results has haste.
func TestKangPrimeEntersAndGivesTheNonlandCardSuspendThroughToAHastyCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := gsSetLibrary(me, gsForest(), gsForest(), gsGoblin(), gsForest())
	msc17Play(t, g, me, gsKangCard())

	for _, land := range ids[:2] {
		c := gsExiled(g, land)
		if c == nil {
			t.Fatal("a land passed over is not in exile")
		}
		if c.SuspendGranted || c.Counters[game.CounterTime] != 0 {
			t.Errorf("a land was given counters or suspend: %+v", c)
		}
	}
	goblin := gsExiled(g, ids[2])
	if goblin == nil {
		t.Fatal("the nonland card is not in exile")
	}
	if goblin.Counters[game.CounterTime] != 2 || !goblin.SuspendGranted {
		t.Fatalf("the nonland card has %d time counters, granted suspend %v; want 2 and true",
			goblin.Counters[game.CounterTime], goblin.SuspendGranted)
	}
	if !me.Library.Contains(ids[3]) {
		t.Error("the walk went past the first nonland card")
	}

	// Another seat's upkeep does not count it down.
	gsNextUpkeep(t, g, 1)
	if n := gsExiled(g, ids[2]).Counters[game.CounterTime]; n != 2 {
		t.Fatalf("an opponent's upkeep left %d time counters, want 2", n)
	}
	gsNextUpkeep(t, g, 0)
	if n := gsExiled(g, ids[2]).Counters[game.CounterTime]; n != 1 {
		t.Fatalf("after one upkeep: %d time counters, want 1", n)
	}
	if latestChoiceOfKind(g, game.PendingChoiceResolutionCast) != nil {
		t.Fatal("the cast was offered before the last counter came off")
	}
	gsNextUpkeep(t, g, 0)
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("the last counter offered no cast")
	}
	answerKeywordCast(t, g, offer, true)
	passPriorityAroundTable(t, g)

	var found *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == ids[2] {
			found = &g.Battlefield.Cards[i]
		}
	}
	if found == nil {
		t.Fatal("the creature was not cast from exile")
	}
	if !game.HasKeyword(found, "haste") {
		t.Error("a creature cast this way has no haste")
	}
	if found.SuspendGranted || found.Counters[game.CounterTime] != 0 {
		t.Errorf("the permanent kept the exiled card's state: %+v", found)
	}
}

// A card that already has suspend gets the counters and nothing else: the
// one countdown removes one counter an upkeep, not two.
func TestKangPrimeDoesNotGiveASecondSuspendToACardThatHasOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := gsSetLibrary(me, game.Card{Name: "Rift Bolt", TypeLine: "Sorcery", OracleID: gsRiftBoltOracle, ManaCost: "{2}{R}"})
	msc17Play(t, g, me, gsKangCard())

	bolt := gsExiled(g, ids[0])
	if bolt == nil {
		t.Fatal("Rift Bolt is not in exile")
	}
	if bolt.SuspendGranted {
		t.Error("a card with printed suspend was marked as given suspend")
	}
	if bolt.Counters[game.CounterTime] != 2 {
		t.Fatalf("Rift Bolt has %d time counters, want 2", bolt.Counters[game.CounterTime])
	}
	gsNextUpkeep(t, g, 1)
	gsNextUpkeep(t, g, 0)
	if n := gsExiled(g, ids[0]).Counters[game.CounterTime]; n != 1 {
		t.Errorf("after one upkeep Rift Bolt has %d time counters, want 1 (one countdown)", n)
	}
}

// A library with no nonland card is exiled whole and nothing is counted.
func TestKangPrimeWithOnlyLandsLeftExilesThemAndCountsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := gsSetLibrary(me, gsForest(), gsForest())
	msc17Play(t, g, me, gsKangCard())
	for _, id := range ids {
		c := gsExiled(g, id)
		if c == nil || c.SuspendGranted || c.Counters[game.CounterTime] != 0 {
			t.Fatalf("a land after an all-land walk: %+v", c)
		}
	}
	if me.Library.Size() != 0 {
		t.Errorf("the library has %d cards left, want 0", me.Library.Size())
	}
}

// Kang attacking is the same one ability.
func TestKangPrimeAttackingDoesItAgain(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kang := b12Push(g, me.ID, "Kang Prime", "Legendary Creature — Human Villain", gsKangOracle, 3, 5)
	ids := gsSetLibrary(me, gsForest(), gsGoblin())
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kang, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	goblin := gsExiled(g, ids[1])
	if goblin == nil || goblin.Counters[game.CounterTime] != 2 || !goblin.SuspendGranted {
		t.Fatalf("after attacking: %+v", goblin)
	}
	if c := gsExiled(g, ids[0]); c == nil || c.SuspendGranted {
		t.Errorf("the land: %+v", c)
	}
}

func gsPlatformBoard(t *testing.T) (g *game.Game, me, opp *game.Player, platform, attacker uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	platform = b12Push(g, me.ID, "Doom's Time Platform", "Artifact", gsPlatformOracle, 0, 0)
	attacker = b12Creature(g, me.ID, "Test Attacker", "Creature — Test", 2, 2)
	return g, me, opp, platform, attacker
}

func gsAttack(t *testing.T, g *game.Game, defender, attacker uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}

func gsGraveyardCard(p *game.Player, c game.Card) uuid.UUID {
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

// Whenever you attack, one trigger exiles the chosen nonland card from your
// graveyard with two time counters, and it gains suspend. A land in the
// graveyard and a card in an opponent's graveyard are not legal targets.
func TestDoomsTimePlatformExilesANonlandGraveyardCardWithTwoCountersAndSuspend(t *testing.T) {
	g, me, opp, _, attacker := gsPlatformBoard(t)
	goblin := gsGraveyardCard(me, gsGoblin())
	land := gsGraveyardCard(me, gsForest())
	theirs := gsGraveyardCard(opp, game.Card{Name: "Their Bolt", TypeLine: "Instant", ManaCost: "{R}"})

	gsAttack(t, g, opp.ID, attacker)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		for _, bad := range []uuid.UUID{land, theirs} {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Fatalf("the trigger accepted the illegal target %s", bad)
			}
		}
		pickCard(t, g, me.ID, goblin)
	}
	passPriorityAroundTable(t, g)

	c := gsExiled(g, goblin)
	if c == nil {
		t.Fatal("the targeted card is not in exile")
	}
	if c.Counters[game.CounterTime] != 2 || !c.SuspendGranted {
		t.Errorf("the exiled card has %d time counters, suspend granted %v; want 2 and true",
			c.Counters[game.CounterTime], c.SuspendGranted)
	}
	if !me.Graveyard.Contains(land) || !opp.Graveyard.Contains(theirs) {
		t.Error("a card that was not targeted left its graveyard")
	}
}

// With only a land in the graveyard there is nothing to target, so nothing
// is exiled.
func TestDoomsTimePlatformCannotTargetALandCard(t *testing.T) {
	g, me, opp, _, attacker := gsPlatformBoard(t)
	land := gsGraveyardCard(me, gsForest())

	gsAttack(t, g, opp.ID, attacker)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
			t.Fatal("the trigger accepted a land card")
		}
	}
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(land) || gsExiled(g, land) != nil {
		t.Error("a land card was exiled")
	}
}

// Attacking with several creatures is one attack declaration: one trigger,
// one card.
func TestDoomsTimePlatformTriggersOncePerAttack(t *testing.T) {
	g, me, opp, _, attacker := gsPlatformBoard(t)
	second := b12Creature(g, me.ID, "Second Attacker", "Creature — Test", 2, 2)
	a := gsGraveyardCard(me, gsGoblin())
	b := gsGraveyardCard(me, gsGoblin())

	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{attacker, second} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) != nil; i++ {
		pickCard(t, g, me.ID, a)
	}
	passPriorityAroundTable(t, g)
	exiled := 0
	for _, id := range []uuid.UUID{a, b} {
		if gsExiled(g, id) != nil {
			exiled++
		}
	}
	if exiled != 1 {
		t.Errorf("%d cards were exiled by one attack, want 1", exiled)
	}
}

// The exiled card then counts down and is offered the free cast, as for
// Kang Prime.
func TestDoomsTimePlatformCardCountsDownAndIsOfferedAFreeCast(t *testing.T) {
	g, me, opp, _, attacker := gsPlatformBoard(t)
	goblin := gsGraveyardCard(me, gsGoblin())

	gsAttack(t, g, opp.ID, attacker)
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, goblin)
	}
	passPriorityAroundTable(t, g)
	if c := gsExiled(g, goblin); c == nil || c.Counters[game.CounterTime] != 2 {
		t.Fatalf("the exiled card: %+v", c)
	}
	gsNextUpkeep(t, g, 1)
	gsNextUpkeep(t, g, 0)
	gsNextUpkeep(t, g, 0)
	offer := latestChoiceOfKind(g, game.PendingChoiceResolutionCast)
	if offer == nil {
		t.Fatal("the last counter offered no cast")
	}
	answerKeywordCast(t, g, offer, true)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(goblin) {
		t.Fatal("the creature was not cast from exile")
	}
}
