package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const nivMizzetSupremeOracle = "4c3a865a-88b1-43e0-b326-17c45670da54"

func nivSupreme(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Niv-Mizzet, Supreme", "Legendary Creature — Dragon Avatar", nivMizzetSupremeOracle, false)
}

// nivGraveyardCard seeds an uncatalogued card with a mana cost in the
// active seat's graveyard and walks to a main phase.
func nivGraveyardCard(t *testing.T, g *game.Game, name, typeLine, cost string) uuid.UUID {
	t.Helper()
	id := seedGraveyardCard(t, g, name, typeLine, "")
	setManaCost(g, g.Seats[g.Turn.ActiveSeat], id, cost)
	return id
}

func nivFund(g *game.Game, p *game.Player, colors ...string) {
	g.WithWriteLock(func() {
		p.ManaPool = nil
		for _, c := range colors {
			p.ManaPool.AddMana(game.ManaToken{Color: c})
		}
	})
}

func TestNivMizzetSupremeIsFullWithItsKeywords(t *testing.T) {
	spec, ok := Lookup(nivMizzetSupremeOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("want CompletenessFull with no caveats, got %v %v", spec.Completeness, spec.Caveats)
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	niv := nivSupreme(g, me.ID)
	if !effectiveAbilitiesContain(t, g, niv, "flying") {
		t.Errorf("abilities %v lack flying", effectiveAbilities(t, g, niv))
	}
	if effectiveAbilitiesContain(t, g, niv, "hexproof") {
		t.Error("it is hexproof from monocolored only, not plain hexproof")
	}
	if creatureTargetableFrom(g, opp.ID, niv, srcRed) {
		t.Error("an opponent's monocolored source must not target it")
	}
	if !creatureTargetableFrom(g, opp.ID, niv, srcGold) {
		t.Error("an opponent's multicolored source still can")
	}
	if !creatureTargetableFrom(g, opp.ID, niv, srcColourless) {
		t.Error("an opponent's colorless source still can")
	}
	if !creatureTargetableFrom(g, me.ID, niv, srcRed) {
		t.Error("its controller's own monocolored spells still can")
	}
}

// The grant picks instant and sorcery cards that are exactly two colours,
// in its controller's graveyard only, and ends when Niv-Mizzet leaves.
func TestNivMizzetSupremeGrantsJumpStartToExactlyTwoColourInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	niv := nivSupreme(g, me.ID)
	for _, tc := range []struct {
		name, typeLine, cost string
		want                 bool
	}{
		{"Gold Instant", "Instant", "{U}{R}", true},
		{"Gold Sorcery", "Sorcery", "{2}{G}{W}", true},
		{"Hybrid Sorcery", "Sorcery", "{W/B}", true},
		{"Mono Instant", "Instant", "{1}{U}", false},
		{"Three-Colour Instant", "Instant", "{U}{B}{R}", false},
		{"Colourless Sorcery", "Sorcery", "{4}", false},
		{"Gold Creature", "Creature — Elf Knight", "{G}{W}", false},
	} {
		id := nivGraveyardCard(t, g, tc.name, tc.typeLine, tc.cost)
		got := flashbackOffers(t, g, me.ID, id)
		if _, has := got[game.AltCostKeyJumpStart]; has != tc.want {
			t.Errorf("%s (%s): offers %v, want jump-start=%v", tc.name, tc.cost, got, tc.want)
		}
		if tc.want && got[game.AltCostKeyJumpStart] != tc.cost {
			t.Errorf("%s: jump-start costs %q, want its mana cost %q", tc.name, got[game.AltCostKeyJumpStart], tc.cost)
		}
	}

	// Another player's graveyard is not "your graveyard", and their own
	// cards get nothing from a Niv-Mizzet they do not control.
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Gold Instant", TypeLine: "Instant",
			ManaCost: "{U}{R}", Owner: opp.ID, Controller: opp.ID})
	})
	for _, seat := range []uuid.UUID{me.ID, opp.ID} {
		if got := flashbackOffers(t, g, seat, theirs); len(got) != 0 {
			t.Errorf("offers on an opponent's graveyard card for %s = %v, want none", seat, got)
		}
	}

	gold := nivGraveyardCard(t, g, "Late Gold Instant", "Instant", "{B}{G}")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(niv) })
	if got := flashbackOffers(t, g, me.ID, gold); len(got) != 0 {
		t.Errorf("offers after Niv-Mizzet left = %v, want none", got)
	}
	fodder := handCard(me, "Fodder", "Sorcery")
	nivFund(g, me, "B", "G")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: game.AltCostKeyJumpStart, AltCostIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("a card that lost jump-start was still castable from the graveyard")
	}
}

// The granted jump-start costs the card's mana cost AND a discarded card,
// and the spell is exiled as it leaves the stack (CR 702.133a).
func TestNivMizzetSupremeJumpStartChargesManaAndADiscardThenExiles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nivSupreme(g, me.ID)
	gold := nivGraveyardCard(t, g, "Gold Instant", "Instant", "{U}{R}")
	mono := nivGraveyardCard(t, g, "Mono Instant", "Instant", "{U}")
	fodder := handCard(me, "Fodder", "Creature — Bear")
	cast := func(id uuid.UUID, discard ...uuid.UUID) error {
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard",
			AlternativeCost: game.AltCostKeyJumpStart, AltCostIDs: discard})
	}

	nivFund(g, me, "U", "R", "U")
	if err := cast(mono, fodder); err == nil {
		t.Fatal("a monocolored instant was cast with jump-start")
	}
	if err := cast(gold); err == nil {
		t.Fatal("jump-start was castable without discarding a card")
	}
	nivFund(g, me, "U")
	if err := cast(gold, fodder); err == nil {
		t.Fatal("jump-start was castable without the card's mana cost: the discard is in addition")
	}
	if !inHandOf(me, fodder) {
		t.Fatal("a refused cast discarded the card")
	}
	nivFund(g, me, "U", "R", "U")
	if err := cast(gold, fodder); err != nil {
		t.Fatalf("jump-start cast: %v", err)
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discard is a cost, paid as the spell is cast")
	}
	if n := me.ManaPool.Total(); n != 1 {
		t.Errorf("%d mana left of three, want 1: the spell costs its {U}{R}", n)
	}
	var item *game.StackItem
	g.ReadSnapshot(func() { item = g.StackMeta[gold] })
	if item == nil || item.AltCost != game.AltCostKeyJumpStart {
		t.Fatalf("stack item = %+v, want a jump-start spell", item)
	}

	// The exile rides the spell, so it survives a restore with the spell
	// still on the stack.
	restored := restoreRoundTrip(t, g, true)
	passPriorityAroundTable(t, restored)
	if !restored.Exile.Contains(gold) {
		t.Error("restored: the jump-started instant is not exiled")
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gold) || me.Graveyard.Contains(gold) {
		t.Error("the jump-started instant is not exiled")
	}
}

// A jump-started spell that is countered is exiled too: "any time it
// would leave the stack".
func TestNivMizzetSupremeJumpStartedSpellIsExiledWhenCountered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nivSupreme(g, me.ID)
	gold := nivGraveyardCard(t, g, "Gold Sorcery", "Sorcery", "{R}{G}")
	fodder := handCard(me, "Fodder", "Instant")
	nivFund(g, me, "R", "G")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: game.AltCostKeyJumpStart, AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("jump-start cast: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CounterTargetForEffect(gold); err != nil {
			t.Fatalf("counter: %v", err)
		}
	})
	if !g.Exile.Contains(gold) {
		t.Error("the countered jump-start spell is not in exile")
	}
}

// A card that prints jump-start and is granted it has two (CR 113.2c) at
// one price, so the cast picker lists a single jump-start (#3938: a
// granted instance that only repeats the printed price is not a second
// row). A monocolored one has its printed jump-start alone either way.
func TestNivMizzetSupremeBesideAPrintedJumpStart(t *testing.T) {
	const radicalIdea = "fd51dce2-8e8a-4686-910e-1bbc2825d190"
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nivSupreme(g, me.ID)
	mono := seedGraveyardCard(t, g, "Radical Idea", "Instant", radicalIdea)
	setManaCost(g, me, mono, "{1}{U}")
	gold := seedGraveyardCard(t, g, "Radical Idea", "Instant", radicalIdea)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == gold {
				me.Graveyard.Cards[i].ManaCost = "{1}{U}"
				me.Graveyard.Cards[i].Colors = []string{"U", "R"}
			}
		}
	})
	for name, id := range map[string]uuid.UUID{"monocolored": mono, "two-colour": gold} {
		if got := flashbackOffers(t, g, me.ID, id); len(got) != 1 || got[game.AltCostKeyJumpStart] != "{1}{U}" {
			t.Errorf("%s Radical Idea offers = %v, want one jump-start at {1}{U}", name, got)
		}
	}
	fodder := handCard(me, "Fodder", "Sorcery")
	nivFund(g, me, "U", "U")
	if err := g.CastSpell(me.ID, gold, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: game.AltCostKeyJumpStart, AltCostIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("jump-start: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(gold) {
		t.Error("the jump-started card is not exiled")
	}
}
