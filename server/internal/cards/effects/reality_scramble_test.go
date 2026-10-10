package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_scramble_test.go — #2550. Reality Scramble against the real
// catalog: the target's card types are read before it is tucked, the
// reveal stops on the first card sharing one, and the rest go under.

const realityScrambleOracle = "afad2e76-4b53-4884-8f96-ba68b0808990"

// scrambleTable clears seat 0's library and stocks it top-first.
func scrambleTable(g *game.Game, me *game.Player, topFirst ...game.Card) []uuid.UUID {
	me.Library.Cards = nil
	ids := make([]uuid.UUID, len(topFirst))
	for i := len(topFirst) - 1; i >= 0; i-- {
		c := topFirst[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		ids[i] = c.InstanceID
		me.Library.PushTop(c)
	}
	return ids
}

func scrambleCast(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	castCatalogSpell(t, g, "Reality Scramble", "Sorcery", realityScrambleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	passPriorityAroundTable(t, g)
}

func scrambleIndex(p *game.Player, id uuid.UUID) int {
	for i, c := range p.Library.Cards {
		if c.InstanceID == id {
			return i
		}
	}
	return -1
}

func TestRealityScrambleRevealsUntilASharedTypeAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clockwork Bear",
		TypeLine: "Artifact Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	ids := scrambleTable(g, me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sol Ring", TypeLine: "Artifact"}, // shares "artifact": the stop
		game.Card{Name: "Deeper Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
	)
	scrambleCast(t, g, target)

	if !g.Battlefield.Contains(ids[2]) {
		t.Fatal("the first card sharing a type did not enter the battlefield")
	}
	if g.Battlefield.Contains(target) {
		t.Fatal("the target stayed on the battlefield")
	}
	if g.Battlefield.Contains(ids[3]) || !me.Library.Contains(ids[3]) {
		t.Error("the reveal ran past its stop: the deeper creature must stay in the library")
	}
	ti := scrambleIndex(me, target)
	if ti < 0 {
		t.Fatal("the target is not in the library")
	}
	// Bottom is index 0. The target went under first, then the revealed
	// rest went under it: bottom to top is rest, target, deeper bear.
	for _, rest := range ids[:2] {
		ri := scrambleIndex(me, rest)
		if ri < 0 || ri >= ti {
			t.Errorf("a revealed non-match is at %d, target at %d: the rest go UNDER the target", ri, ti)
		}
	}
	if scrambleIndex(me, ids[3]) <= ti {
		t.Error("the unrevealed creature was already in the library, so the tucked target goes under it")
	}
}

// With no other card sharing a type the reveal reaches the tucked
// target itself on the way down and puts IT back, which is the printed
// result (it was put on the bottom first).
func TestRealityScrambleWithNoMatchFindsTheTuckedTargetAgain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Lone Wolf",
		TypeLine: "Creature — Wolf", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	ids := scrambleTable(g, me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
	)
	scrambleCast(t, g, target)

	if !g.Battlefield.Contains(target) {
		t.Fatal("the tucked target was not revealed and returned")
	}
	for _, id := range ids {
		if !me.Library.Contains(id) {
			t.Error("a revealed non-match left the library")
		}
	}
}

// A token target is tucked (and ceases to exist, CR 111.8) but still
// had its types; and a token revealed on the way never stops the run.
func TestRealityScrambleOnATokenUsesItsTypesAndSkipsTokensInTheLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin",
		TypeLine: "Token Creature — Goblin", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	ids := scrambleTable(g, me,
		game.Card{Name: "Stray Token", TypeLine: "Token Creature — Saproling", Power: 1, Toughness: 1},
		game.Card{Name: "Wall of Omens", TypeLine: "Creature — Wall", Power: 0, Toughness: 4},
	)
	scrambleCast(t, g, target)

	if g.Battlefield.Contains(target) || scrambleIndex(me, target) >= 0 {
		t.Error("the tucked token should have ceased to exist")
	}
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the real creature card after the token did not enter")
	}
	if g.Battlefield.Contains(ids[0]) {
		t.Error("a token in the library stopped the reveal")
	}
}

func TestRealityScrambleOnlyTargetsPermanentsYouOwn(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Ogre",
		TypeLine: "Creature — Ogre", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	// Casting needs a card to cast; reuse the helper's setup by
	// expecting the announce to refuse the clause.
	err := castCatalogSpellErr(t, g, "Reality Scramble", "Sorcery", realityScrambleOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	if err == nil {
		t.Fatal("Reality Scramble accepted a permanent the caster does not own")
	}
}

// A permanent you own but another player controls is a legal target and
// the types are read off it all the same.
func TestRealityScrambleTakesAPermanentYouOwnButDoNotControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Stolen Elf",
		TypeLine: "Creature — Elf", Power: 1, Toughness: 1, Owner: me.ID, Controller: opp.ID})
	ids := scrambleTable(g, me, game.Card{Name: "Fresh Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	scrambleCast(t, g, target)
	if g.Battlefield.Contains(target) || !me.Library.Contains(target) {
		t.Error("the stolen permanent did not go to its owner's library")
	}
	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the matching creature did not enter")
	}
}

// Retrace: Reality Scramble returns to the graveyard after it resolves
// and is castable from there for its printed cost plus a land.
func TestRealityScrambleRetraces(t *testing.T) {
	g, me, land, _ := retraceTable(t)
	addForests(g, me, 0)
	if !game.CardCastableFromZone(realityScrambleOracle, game.ZoneGraveyard) {
		t.Fatal("the graveyard is not open to Reality Scramble")
	}
	offers := game.AlternativeCostsOfferedFromZone(realityScrambleOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "retrace" || offers[0].ManaCost != "{2}{R}{R}" {
		t.Fatalf("graveyard offers = %+v, want retrace for the printed {2}{R}{R}", offers)
	}
	target := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bear",
		TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	scrambleTable(g, me, game.Card{Name: "Elf", TypeLine: "Creature — Elf", Power: 1, Toughness: 1})
	spell := graveyardCardOf(g, me, "Reality Scramble", "Sorcery", "{2}{R}{R}", realityScrambleOracle)
	payMana(t, g, me, "{R}{R}{R}{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "retrace", AltCostIDs: []uuid.UUID{land}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("retrace cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inGraveyard(me, spell) {
		t.Error("Reality Scramble did not return to the graveyard to be retraced again")
	}
	if g.Battlefield.Contains(target) {
		t.Error("the target stayed on the battlefield")
	}
}
