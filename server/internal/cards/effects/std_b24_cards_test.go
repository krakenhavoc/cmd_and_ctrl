package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b24_cards_test.go — Standard batch 24, Duskmourn: House of Horror
// (#4466). Each test plays the real catalog entry and asserts what its text
// says.

const (
	b24FearOracle     = "0ae74149-60fc-407a-8782-8865458606b5"
	b24GremlinOracle  = "44594004-52e7-4ce6-8eb0-61f72fd9c73b"
	b24LeylineOracle  = "f69574dd-097f-4908-9cab-344b0eb39c62"
	b24NorinOracle    = "425d755b-7a15-423d-9640-494debc79dd9"
	b24TrialOracle    = "4ef07e5a-3dc3-41dd-a3a9-e8812b4d40d3"
	b24WaltzOracle    = "78df47c3-b771-4377-8963-ae3065fdcf8a"
	b24AltanakOracle  = "6591f243-f94c-4a83-bb16-1fb838d9eb14"
	b24AnthropedeOrcl = "452d98ae-0fa2-4cff-9e18-eecd76ddb20b"
	b24WurmOracle     = "2dd32fec-5bc3-4cc0-a9b4-814229a64914"
	b24ParTingOracle  = "0ea80bd9-ab7e-4a8f-85b0-167fe7fae43a"

	b24FearType       = "Enchantment Creature — Nightmare"
	b24GremlinType    = "Creature — Gremlin"
	b24LeylineType    = "Enchantment"
	b24NorinType      = "Legendary Creature — Human Coward"
	b24AltanakType    = "Legendary Creature — Insect Beast"
	b24AnthropedeType = "Creature — Insect"
	b24WurmType       = "Creature — Wurm"
	b24StartingLife   = 40

	b24GiantGrowthOracle = "5748ebf1-24e3-499d-ab7c-c2cebd462a24"
)

func TestStdB24EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b24FearOracle:     "Fear of Burning Alive",
		b24GremlinOracle:  "Irreverent Gremlin",
		b24LeylineOracle:  "Leyline of Resonance",
		b24NorinOracle:    "Norin, Swift Survivalist",
		b24TrialOracle:    "Trial of Agony",
		b24WaltzOracle:    "Waltz of Rage",
		b24AltanakOracle:  "Altanak, the Thrice-Called",
		b24AnthropedeOrcl: "Anthropede",
		b24WurmOracle:     "Balustrade Wurm",
		b24ParTingOracle:  "Cathartic Parting",
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

// b24SeedTypes puts the first n of four card types into `p`'s graveyard: four
// is delirium, three is not.
func b24SeedTypes(p *game.Player, n int) {
	cards := []struct{ name, typeLine, cost string }{
		{"Old Bear", "Creature — Bear", "{1}{G}"},
		{"Old Bolt", "Instant", "{R}"},
		{"Old Ritual", "Sorcery", "{B}"},
		{"Old Forest", "Basic Land — Forest", ""},
	}
	for _, c := range cards[:n] {
		gyCard(p, c.name, c.typeLine, c.cost, "")
	}
}

func b24Damage(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return b12Card(t, g, id).DamageMarked
}

// b24Enter casts a free creature spell with the given power from `p`'s hand
// and resolves it (and what it triggers, up to a prompt), so it really enters.
func b24Enter(t *testing.T, g *game.Game, p *game.Player, name, typeLine string, power int) uuid.UUID {
	t.Helper()
	id := putInHand(p, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: power + 1})
	toMain(t, g)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func b24HandCards(p *game.Player, names ...string) []uuid.UUID {
	var ids []uuid.UUID
	for _, n := range names {
		ids = append(ids, putInHand(p, game.Card{Name: n, TypeLine: "Sorcery"}))
	}
	return ids
}

// --- Fear of Burning Alive -------------------------------------------------------

func TestStdB24FearOfBurningAliveDealsFourToEachOpponentWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b24SeedTypes(me, 3)
	castAndResolveCreature(t, g, "Fear of Burning Alive", b24FearType, b24FearOracle)
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != b24StartingLife-4 {
			t.Errorf("%s has %d life, want %d", p.Name, p.Life, b24StartingLife-4)
		}
	}
	if me.Life != b24StartingLife {
		t.Errorf("the caster lost life: %d", me.Life)
	}
}

func TestStdB24FearOfBurningAliveHitsThatPlayersCreatureForThatAmountWithDelirium(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	fear := b12Push(g, me.ID, "Fear of Burning Alive", b24FearType, b24FearOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	others := b12Creature(g, other.ID, "Other Wall", "Creature — Wall", 0, 9)
	pinger := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
	b24SeedTypes(me, 4)

	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, opp.ID, 3) })
	answerPickTargetIfOpen(t, g, theirs)
	passPriorityAroundTable(t, g)
	if got := b24Damage(t, g, theirs); got != 3 {
		t.Errorf("that player's creature has %d damage, want 3 (the amount dealt)", got)
	}
	if got := b24Damage(t, g, others); got != 0 {
		t.Errorf("another opponent's creature was dealt %d", got)
	}
	if got := b24Damage(t, g, fear); got != 0 {
		t.Errorf("Fear of Burning Alive was dealt %d", got)
	}
	// Damage to a different opponent aims at that player's creatures only.
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, other.ID, 2) })
	answerPickTargetIfOpen(t, g, others)
	passPriorityAroundTable(t, g)
	if got := b24Damage(t, g, others); got != 2 {
		t.Errorf("the other opponent's creature has %d damage, want 2", got)
	}
}

func TestStdB24FearOfBurningAliveTriggersOncePerOpponentDamaged(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b24SeedTypes(me, 4)
	var walls []uuid.UUID
	for _, p := range g.Seats[1:] {
		walls = append(walls, b12Creature(g, p.ID, "Wall", "Creature — Wall", 0, 9))
	}
	castAndResolveCreature(t, g, "Fear of Burning Alive", b24FearType, b24FearOracle)
	passPriorityAroundTable(t, g)
	// The enters trigger dealt 4 to each of three opponents, and each of those
	// is a delirium trigger of its own aimed at that player's creature.
	for i := 0; i < 6 && latestPickTarget(g, me.ID) != nil; i++ {
		p := latestPickTarget(g, me.ID)
		var pick uuid.UUID
		for _, w := range walls {
			if hasID(p.PickTargetCards, w) {
				pick = w
			}
		}
		if pick == uuid.Nil {
			t.Fatal("no wall among the candidates")
		}
		pickCard(t, g, me.ID, pick)
		passPriorityAroundTable(t, g)
	}
	passPriorityAroundTable(t, g)
	for i, w := range walls {
		if got := b24Damage(t, g, w); got != 4 {
			t.Errorf("opponent %d's wall has %d damage, want 4", i+1, got)
		}
	}
}

func TestStdB24FearOfBurningAliveNeedsDeliriumNoncombatDamageAndAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Fear of Burning Alive", b24FearType, b24FearOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	mine := b12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	pinger := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)

	b24SeedTypes(me, 3)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, opp.ID, 3) })
	if latestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Error("three card types: the trigger fired")
	}
	passPriorityAroundTable(t, g)
	gyCard(me, "Old Forest", "Basic Land — Forest", "", "")
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, me.ID, 2) })
	if latestPickTarget(g, me.ID) != nil || len(g.PendingTriggers) != 0 {
		t.Error("damage to yourself triggered it")
	}
	passPriorityAroundTable(t, g)
	if b24Damage(t, g, theirs) != 0 || b24Damage(t, g, mine) != 0 {
		t.Error("a creature was dealt damage with the condition unmet")
	}
}

func TestStdB24FearOfBurningAliveRechecksDeliriumOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Fear of Burning Alive", b24FearType, b24FearOracle, 4, 4)
	theirs := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	pinger := b12Creature(g, me.ID, "Pinger", "Creature — Wizard", 1, 1)
	b24SeedTypes(me, 4)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(pinger, opp.ID, 3) })
	answerPickTargetIfOpen(t, g, theirs)
	// The land leaves the graveyard in response: three types remain.
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Old Forest" {
			me.Graveyard.Remove(c.InstanceID)
			break
		}
	}
	passPriorityAroundTable(t, g)
	if got := b24Damage(t, g, theirs); got != 0 {
		t.Errorf("the ability did %d with delirium gone on resolution (CR 603.4)", got)
	}
}

// --- Irreverent Gremlin ----------------------------------------------------------

func TestStdB24IrreverentGremlinLoopsOncePerTurnForSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	gremlin := b12Push(g, me.ID, "Irreverent Gremlin", b24GremlinType, b24GremlinOracle, 2, 2)
	if !hasKeywordOf(g, gremlin, "menace") {
		t.Error("Irreverent Gremlin does not have menace")
	}
	hand := b24HandCards(me, "Pitch Me", "Keep Me")
	seedLibrary(me, "A", "B", "C", "D")

	before := me.Hand.Size()
	b24Enter(t, g, me, "Elf One", "Creature — Elf", 2)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("a power-2 creature entering did not offer the loot")
	}
	answerDiscard(t, g, me.ID, hand[0])
	passPriorityAroundTable(t, g)
	if me.Hand.Contains(hand[0]) || !me.Hand.Contains(hand[1]) {
		t.Error("the chosen card was not the one discarded")
	}
	// The creature was put in hand and cast; one card discarded, one drawn.
	if me.Hand.Size() != before {
		t.Errorf("hand %d -> %d, want the discard and the draw to cancel", before, me.Hand.Size())
	}

	// Once each turn: the next small creature neither triggers nor asks.
	b24Enter(t, g, me, "Elf Two", "Creature — Elf", 1)
	if discardChoiceFor(g, me.ID) != nil || len(g.PendingTriggers) != 0 || g.Stack.Size() != 0 {
		t.Error("the loot was offered a second time this turn")
	}
}

func TestStdB24IrreverentGremlinIgnoresACreatureWithPowerThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Irreverent Gremlin", b24GremlinType, b24GremlinOracle, 2, 2)
	b24HandCards(me, "Pitch Me")
	seedLibrary(me, "A", "B")
	b24Enter(t, g, me, "Big Bear", "Creature — Bear", 3)
	if discardChoiceFor(g, me.ID) != nil || g.Stack.Size() != 0 || len(g.PendingTriggers) != 0 {
		t.Fatal("a power-3 creature offered the loot")
	}
}

func TestStdB24IrreverentGremlinItsOwnEntryIsNotAnotherCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b24HandCards(me, "Pitch Me")
	seedLibrary(me, "A", "B")
	id := putInHand(me, game.Card{Name: "Irreverent Gremlin", TypeLine: b24GremlinType, OracleID: b24GremlinOracle, Power: 2, Toughness: 2})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("the Gremlin's own entry offered the loot")
	}
}

func TestStdB24IrreverentGremlinPowerIsReadAsTheCreatureEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Irreverent Gremlin", b24GremlinType, b24GremlinOracle, 2, 2)
	hand := b24HandCards(me, "Pitch Me")
	seedLibrary(me, "A", "B")
	id := putInHand(me, game.Card{Name: "Elf One", TypeLine: "Creature — Elf", Power: 2, Toughness: 3})
	toMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// Resolve the creature spell only, so the ability is waiting.
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	putCounters(t, g, id, game.CounterPlusOne, 2)
	passPriorityAroundTable(t, g)
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("the loot was not offered after the creature grew in response")
	}
	answerDiscard(t, g, me.ID, hand[0])
}

func TestStdB24IrreverentGremlinDecliningLeavesTheTurnsLoot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Irreverent Gremlin", b24GremlinType, b24GremlinOracle, 2, 2)
	hand := b24HandCards(me, "Pitch Me")
	seedLibrary(me, "A", "B")
	b24Enter(t, g, me, "Elf One", "Creature — Elf", 1)
	answerDiscard(t, g, me.ID) // "you may": discard nothing
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(hand[0]) {
		t.Fatal("declining discarded a card")
	}
	b24Enter(t, g, me, "Elf Two", "Creature — Elf", 1)
	if discardChoiceFor(g, me.ID) == nil {
		t.Error("a declined loot used up the turn's one")
	}
}

// --- Leyline of Resonance --------------------------------------------------------

func TestStdB24LeylineOfResonanceCopiesASpellTargetingOnlyYourCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if spec, _ := Lookup(b24LeylineOracle); spec.OpeningHand == nil {
		t.Error("Leyline of Resonance does not declare the opening-hand clause")
	}
	b12Push(g, me.ID, "Leyline of Resonance", b24LeylineType, b24LeylineOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	// The bear is the only creature to aim at, so the copy keeps it; if the
	// engine does ask, the answer is the bear.
	if latestPickTarget(g, me.ID) != nil {
		pickCard(t, g, me.ID, bear)
		passPriorityAroundTable(t, g)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 8 || tough != 8 {
		t.Errorf("the bear is %d/%d, want 8/8 after the spell and its copy", p, tough)
	}
}

func TestStdB24LeylineOfResonanceCopyMayTargetADifferentCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Leyline of Resonance", b24LeylineType, b24LeylineOracle, 0, 0)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wolf := b12Creature(g, me.ID, "Wolf", "Creature — Wolf", 3, 3)
	castCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, cardRefs(bear))
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, wolf)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, bear); p != 5 {
		t.Errorf("the bear is %d power, want 5 (the original)", p)
	}
	if p, _ := effectivePTOf(t, g, wolf); p != 6 {
		t.Errorf("the wolf is %d power, want 6 (the copy)", p)
	}
}

func TestStdB24LeylineOfResonanceIgnoresOtherTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Leyline of Resonance", b24LeylineType, b24LeylineOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Giant Growth", "Instant", b24GiantGrowthOracle, cardRefs(theirs))
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("a spell targeting an opponent's creature was copied")
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 5 {
		t.Errorf("their bear is %d power, want 5 (no copy)", p)
	}
	// A creature spell is not an instant or a sorcery, and an opponent's
	// spell is not yours to copy.
	if g.Stack.Size() != 0 {
		t.Error("something is left on the stack")
	}
}
