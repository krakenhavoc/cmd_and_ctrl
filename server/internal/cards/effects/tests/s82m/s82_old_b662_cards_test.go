package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b662_cards_test.go — Invasion instants and sorceries, S82 OLD batch
// 662 (#5616). Every test plays the real catalog entry through the engine.

const (
	b662Coalition = "e9e25800-9ee7-40c9-b22d-611c7281c125"
	b662Wave      = "f455b731-6ff9-4acb-9447-f524faa49b27"
	b662Ruin      = "4a7b89b2-5fe9-4d39-9d42-3ca78bcf6d40"
	b662Light     = "6b315dc3-c330-4b30-b6ad-4da12ccf6ca3"
	b662Lobotomy  = "a7ebe748-c3de-402f-8a6e-236a6413c0f8"
	b662Research  = "308b49ad-ebab-41c4-9e09-44202549bafc"
	b662Spite     = "68b038ef-71c4-4ca5-9821-57bf85831bb7"
	b662Glory     = "b2ae0968-28ed-4ebb-a4e8-69ff6e6a6f1d"
	b662Bear      = "Creature — Bear"
)

func TestB662EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b662Coalition: "Coalition Victory", b662Wave: "Breaking Wave", b662Ruin: "Global Ruin",
		b662Light: "Blinding Light", b662Lobotomy: "Lobotomy", b662Research: "Desperate Research",
		b662Spite: "Barrin's Spite", b662Glory: "Death or Glory",
	} {
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

func b662Push(g *game.Game, owner uuid.UUID, name, typeLine string, colors []string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: colors,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
}

func b662CastSorcery(t *testing.T, g *game.Game, name, oracle string, p game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Sorcery", OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	p.Strict = true
	if err := g.CastSpell(me.ID, id, p); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Coalition Victory ---------------------------------------------------

func b662Table(g *game.Game, me uuid.UUID, withMountain bool, creatureColors [][]string) {
	for _, l := range []string{"Plains", "Island", "Swamp", "Forest"} {
		b662Push(g, me, l, "Basic Land — "+l, nil)
	}
	if withMountain {
		b662Push(g, me, "Mountain", "Basic Land — Mountain", nil)
	}
	for i, cols := range creatureColors {
		b662Push(g, me, "Creature "+string(rune('A'+i)), b662Bear, cols)
	}
}

func TestB662CoalitionVictoryWinsWithEveryLandTypeAndColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	// Two multicolor creatures cover all five colors.
	b662Table(g, me.ID, true, [][]string{{"W", "U", "B"}, {"R", "G"}})
	b662CastSorcery(t, g, "Coalition Victory", b662Coalition, game.CastSpellParams{})
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Fatalf("outcome %+v, want a win for the caster", g.Outcome)
	}
}

func TestB662CoalitionVictoryNeedsEveryLandType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b662Table(g, me.ID, false, [][]string{{"W", "U", "B", "R", "G"}})
	b662CastSorcery(t, g, "Coalition Victory", b662Coalition, game.CastSpellParams{})
	if g.Outcome != nil {
		t.Fatal("won with no Mountain")
	}
}

func TestB662CoalitionVictoryNeedsEveryColorAmongCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b662Table(g, me.ID, true, [][]string{{"W", "U", "B", "R"}})
	// A green noncreature permanent, and an opponent's green creature, do not count.
	b662Push(g, me.ID, "Green Aura", "Enchantment — Aura", []string{"G"})
	b662Push(g, g.Seats[1].ID, "Their Elf", "Creature — Elf", []string{"G"})
	b662CastSorcery(t, g, "Coalition Victory", b662Coalition, game.CastSpellParams{})
	if g.Outcome != nil {
		t.Fatal("won with no green creature of yours")
	}
}

// --- Breaking Wave -------------------------------------------------------

func TestB662BreakingWaveSwapsTappedAndUntappedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	up := b662Push(g, me.ID, "Untapped", b662Bear, nil)
	down := b662Push(g, opp.ID, "Tapped", b662Bear, nil)
	land := b662Push(g, me.ID, "Plains", "Basic Land — Plains", nil)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == down {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	b662CastSorcery(t, g, "Breaking Wave", b662Wave, game.CastSpellParams{})
	if !effectstest.DftCG2Tapped(g, up) {
		t.Error("the untapped creature was not tapped")
	}
	if effectstest.DftCG2Tapped(g, down) {
		t.Error("the tapped creature was not untapped")
	}
	if effectstest.DftCG2Tapped(g, land) {
		t.Error("a land was tapped")
	}
}

func TestB662BreakingWaveHasFlashOnlyForTwoMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Breaking Wave", TypeLine: "Sorcery", ManaCost: "{2}{U}{U}",
			OracleID: b662Wave, Owner: me.ID, Controller: me.ID})
	})
	fund := func(colors ...string) {
		g.WithWriteLock(func() {
			me.ManaPool = nil
			for _, c := range colors {
				me.ManaPool.AddMana(game.ManaToken{Color: c})
			}
		})
	}
	fund("U", "U", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast Breaking Wave at instant speed for its printed cost")
	}
	fund("U", "U", "C", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast Breaking Wave at instant speed paying {2} more: %v", err)
	}
}

// --- Blinding Light ------------------------------------------------------

func TestB662BlindingLightTapsOnlyNonwhiteCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	white := b662Push(g, me.ID, "White", b662Bear, []string{"W"})
	gold := b662Push(g, me.ID, "Gold", b662Bear, []string{"W", "U"})
	red := b662Push(g, opp.ID, "Red", b662Bear, []string{"R"})
	colorless := b662Push(g, opp.ID, "Construct", "Artifact Creature — Construct", nil)
	land := b662Push(g, me.ID, "Plains", "Basic Land — Plains", nil)
	b662CastSorcery(t, g, "Blinding Light", b662Light, game.CastSpellParams{})
	for name, id := range map[string]uuid.UUID{"red": red, "colorless": colorless} {
		if !effectstest.DftCG2Tapped(g, id) {
			t.Errorf("the %s creature was not tapped", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"white": white, "white-blue": gold, "land": land} {
		if effectstest.DftCG2Tapped(g, id) {
			t.Errorf("the %s permanent was tapped", name)
		}
	}
}

// --- Global Ruin ---------------------------------------------------------

func TestB662GlobalRuinKeepsOneLandOfEachType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f1 := b662Push(g, me.ID, "Forest", "Basic Land — Forest", nil)
	f2 := b662Push(g, me.ID, "Forest", "Basic Land — Forest", nil)
	i1 := b662Push(g, me.ID, "Island", "Basic Land — Island", nil)
	waste := b662Push(g, me.ID, "Wasteland", "Land", nil)
	creature := b662Push(g, me.ID, "Bear", b662Bear, nil)
	oppOnly := b662Push(g, opp.ID, "Mountain", "Basic Land — Mountain", nil)
	b662CastSorcery(t, g, "Global Ruin", b662Ruin, game.CastSpellParams{})
	prompt := effectstest.LatestChooseCardsFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked")
	}
	for _, id := range prompt.ChooseCards {
		if id == waste || id == creature {
			t.Fatal("a land with no basic land type, or a creature, was offered")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{f1}); err == nil {
		t.Fatal("keeping a Forest alone, with an Island left out, was accepted")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{f1, f2, i1}); err == nil {
		t.Fatal("two Forests were accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, f2, i1)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"kept Forest": f2, "Island": i1, "creature": creature, "their only Mountain": oppOnly} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("the %s was sacrificed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"other Forest": f1, "typeless land": waste} {
		if g.Battlefield.Contains(id) {
			t.Errorf("the %s was kept", name)
		}
	}
}

func TestB662GlobalRuinLetsADualStandForBothTypes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dual := b662Push(g, me.ID, "Dual", "Land — Forest Island", nil)
	extra := b662Push(g, me.ID, "Island", "Basic Land — Island", nil)
	f := b662Push(g, me.ID, "Forest", "Basic Land — Forest", nil)
	b662CastSorcery(t, g, "Global Ruin", b662Ruin, game.CastSpellParams{})
	if effectstest.LatestChooseCardsFor(g, me.ID) == nil {
		t.Fatal("the caster was not asked")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, dual)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dual) {
		t.Error("the Dual was sacrificed")
	}
	if g.Battlefield.Contains(extra) || g.Battlefield.Contains(f) {
		t.Error("a land chosen for nothing was kept")
	}
}

func TestB662GlobalRuinAsksNothingWhenEveryTypeIsAlreadyOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	f := b662Push(g, me.ID, "Forest", "Basic Land — Forest", nil)
	i := b662Push(g, opp.ID, "Island", "Basic Land — Island", nil)
	b662CastSorcery(t, g, "Global Ruin", b662Ruin, game.CastSpellParams{})
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil || effectstest.LatestChooseCardsFor(g, opp.ID) != nil {
		t.Fatal("a player with one land of one type was asked")
	}
	if !g.Battlefield.Contains(f) || !g.Battlefield.Contains(i) {
		t.Error("a lone land was sacrificed")
	}
}

// --- Barrin's Spite ------------------------------------------------------

func TestB662BarrinsSpiteTheirControllerSacrificesOneAndTheOtherReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b662Push(g, opp.ID, "Their Bear", b662Bear, nil)
	b := b662Push(g, opp.ID, "Their Ox", "Creature — Ox", nil)
	mine := b662Push(g, me.ID, "My Bear", b662Bear, nil)
	effectstest.AdvanceToMain(t, g)
	bad := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bad, Name: "Barrin's Spite", TypeLine: "Sorcery", OracleID: b662Spite, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, bad, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, mine)}); err == nil {
		t.Fatal("targeted creatures with different controllers")
	}
	b662CastSorcery(t, g, "Barrin's Spite", b662Spite, game.CastSpellParams{Targets: effectstest.CardRefs(a, b)})
	if effectstest.LatestChooseCardsFor(g, me.ID) != nil {
		t.Fatal("the caster was asked, not their controller")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, b)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(b) {
		t.Error("the chosen creature was not sacrificed")
	}
	if g.Battlefield.Contains(a) || !opp.Hand.Contains(a) {
		t.Error("the other creature did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("an unrelated creature left")
	}
}

// --- Death or Glory ------------------------------------------------------

func TestB662DeathOrGlorySplitsTheGraveyardCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var cards []uuid.UUID
	for i := 0; i < 3; i++ {
		id := uuid.New()
		cards = append(cards, id)
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Dead Bear", TypeLine: b662Bear, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	}
	spell := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Old Bolt", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	b662CastSorcery(t, g, "Death or Glory", b662Glory, game.CastSpellParams{})
	effectstest.AnswerChoosePlayer(t, g, me.ID, opp)
	effectstest.PassPriorityAroundTable(t, g)
	split := effectstest.LatestRevealPickFor(g, me.ID)
	if split == nil {
		t.Fatal("no pile split offered to the caster")
	}
	if err := g.ResolveRevealPick(split.ID, me.ID, cards[:1]); err != nil {
		t.Fatalf("ResolvePileSplit: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, opp.ID) == nil {
		t.Fatal("the opponent was not asked to choose a pile")
	}
	// Option 0 is the first pile (the one card): it is exiled, the other two return.
	effectstest.AnswerOptionPick(t, g, opp.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	exiled, back := 0, 0
	for _, id := range cards {
		switch {
		case g.Exile.Contains(id):
			exiled++
		case g.Battlefield.Contains(id):
			back++
		}
	}
	if exiled+back != 3 || exiled == 0 && back == 0 {
		t.Errorf("exiled %d, returned %d: the piles were not settled", exiled, back)
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a noncreature card left the graveyard")
	}
}

// --- Lobotomy ------------------------------------------------------------

func TestB662LobotomyExilesEveryCardWithTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	push := func(z *game.Zone, name, typeLine string) uuid.UUID {
		id := uuid.New()
		z.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: opp.ID, Controller: opp.ID})
		return id
	}
	opp.Hand.Cards = nil
	opp.Library.Cards = nil
	hand1 := push(opp.Hand, "Twin Bear", b662Bear)
	hand2 := push(opp.Hand, "Twin Bear", b662Bear)
	other := push(opp.Hand, "Lone Spell", "Sorcery")
	land := push(opp.Hand, "Forest", "Basic Land — Forest")
	gy := push(opp.Graveyard, "Twin Bear", b662Bear)
	lib := push(opp.Library, "Twin Bear", b662Bear)
	libOther := push(opp.Library, "Unrelated", b662Bear)
	b662CastSorcery(t, g, "Lobotomy", b662Lobotomy, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	pick := effectstest.OpenVariant(t, g)
	for _, id := range pick.DiscardOptions {
		if id == land {
			t.Fatal("a basic land card was offered")
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand1, other}); err == nil {
		t.Fatal("two cards were accepted")
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand1}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.PendingChoices) > 0 {
		effectstest.AnswerSearchByID(t, g, me.ID, hand1, hand2, lib)
	}
	for _, id := range []uuid.UUID{hand1, hand2, gy, lib} {
		if !g.Exile.Contains(id) {
			t.Errorf("a Twin Bear was not exiled: %s", id)
		}
	}
	if g.Exile.Contains(other) || g.Exile.Contains(land) || g.Exile.Contains(libOther) {
		t.Error("a card with another name was exiled")
	}
}

// --- Desperate Research --------------------------------------------------

func TestB662DesperateResearchTakesTheNamedCardsAndExilesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	me.Library.Cards = nil
	push := func(name, typeLine string) uuid.UUID {
		id := uuid.New()
		me.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID})
		return id
	}
	deep := push("Deep Card", "Sorcery")
	var top7 []uuid.UUID
	for _, n := range []string{"Bolt", "Other", "Bolt", "Another", "Bolt", "Misc", "Last"} {
		top7 = append(top7, push(n, "Instant"))
	}
	b662CastSorcery(t, g, "Desperate Research", b662Research, game.CastSpellParams{})
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Forest"); err == nil {
		t.Fatal("a basic land card's name was accepted")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range top7 {
		bolt := i == 0 || i == 2 || i == 4
		if bolt && !me.Hand.Contains(id) {
			t.Errorf("a Bolt (%d) is not in hand", i)
		}
		if !bolt && !g.Exile.Contains(id) {
			t.Errorf("card %d of the top seven was not exiled", i)
		}
	}
	if z := g.FindCardZoneForEffect(deep); z == nil || z.Kind != game.ZoneLibrary {
		t.Error("the eighth card left the library")
	}
}
