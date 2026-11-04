package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spm_b15_cards_test.go — Marvel's Spider-Man family, batch 15 (#4460).
// Each test plays the real catalog entry and asserts what its text says.

const (
	spmB15DontMoveOracle   = "23cf095c-d43e-4a80-9bc5-f0f46ff885c4"
	spmB15EscapeOracle     = "45a8c126-5396-4736-a37e-460a50706a98"
	spmB15ExtinctionOracle = "0a11a3a8-ee5f-4caf-bf84-c8248e07c825"
	spmB15FightOracle      = "7f885ffc-7245-4f00-a903-6e088007c635"
	spmB15VigorOracle      = "5f627e45-deee-4a4f-8128-57b0d8ebc553"
	spmB15FuryOracle       = "c90e2db7-6973-4243-9645-265fd3bc8546"
	spmB15IronManOracle    = "67ab82de-1167-42f5-90bd-21c63edcc712"
	spmB15BindingOracle    = "b49642f6-20bd-468d-9dac-e578f9ac3091"
	spmB15PromiseOracle    = "64c54c9f-0550-41e2-a44a-5892a199f670"
	spmB15LorthosOracle    = "d9fd4ff3-5ada-40f4-b949-6fe65624a0c4"
)

func TestSpmB15EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		spmB15DontMoveOracle:   "Don't Move",
		spmB15EscapeOracle:     "Escape to the Wilds",
		spmB15ExtinctionOracle: "Extinction Event",
		spmB15FightOracle:      "Fight to the Death",
		spmB15VigorOracle:      "Force of Vigor",
		spmB15FuryOracle:       "Fury of the Horde",
		spmB15IronManOracle:    "Iron Man, Titan of Innovation",
		spmB15BindingOracle:    "Leyline Binding",
		spmB15PromiseOracle:    "Light of Promise",
		spmB15LorthosOracle:    "Lorthos, the Tidemaker",
	}
	for oracle, name := range want {
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

// spmB15Cast puts the card in p's hand and casts it with params.
func spmB15Cast(g *game.Game, p *game.Player, name, typeLine, cost, oracle string, colors []string, params game.CastSpellParams) (uuid.UUID, error) {
	id := handCardFull(p, name, typeLine, cost, oracle, colors)
	return id, g.CastSpell(p.ID, id, params)
}

// --- Fight to the Death ---

func TestSpmB15FightDestroysBlockersAndBlockedAttackersOnly(t *testing.T) {
	g, me, opp, atks := bcCombat(t, 2)
	bystander := pushVanillaCreature(g, me.ID, "Bystander", 2, 2)
	theirBystander := pushVanillaCreature(g, opp.ID, "Their Bystander", 2, 2)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 5)
	bcAttack(t, g, opp, atks)
	if err := mbBlock(g, wall, atks[0]); err != nil {
		t.Fatal(err)
	}
	lockInBlocks(t, g)
	if _, err := spmB15Cast(g, me, "Fight to the Death", "Instant", "", spmB15FightOracle, nil, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) || g.Battlefield.Contains(atks[0]) {
		t.Error("the blocking creature and the blocked attacker should both be destroyed")
	}
	for name, id := range map[string]uuid.UUID{"unblocked attacker": atks[1], "my bystander": bystander, "their bystander": theirBystander} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was destroyed", name)
		}
	}
}

func TestSpmB15FightStillDestroysAnAttackerWhoseBlockerLeft(t *testing.T) {
	g, me, opp, atks := bcCombat(t, 1)
	wall := pushVanillaCreature(g, opp.ID, "Wall", 0, 5)
	bcAttack(t, g, opp, atks)
	if err := mbBlock(g, wall, atks[0]); err != nil {
		t.Fatal(err)
	}
	lockInBlocks(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wall) })
	g.RunStateChecksForTest()
	if _, err := spmB15Cast(g, me, "Fight to the Death", "Instant", "", spmB15FightOracle, nil, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(atks[0]) {
		t.Error("a creature stays blocked after its blocker left combat (CR 509.1h)")
	}
}

func TestSpmB15FightDoesNothingOutsideCombat(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, opp.ID, "B", 2, 2)
	castCatalogSpell(t, g, "Fight to the Death", "Instant", spmB15FightOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("with no combat nothing is blocking or blocked")
	}
}

// --- Extinction Event ---

func spmB15ParityTable(g *game.Game) map[string]uuid.UUID {
	me, opp := g.Seats[0], g.Seats[1]
	ids := map[string]uuid.UUID{}
	ids["one"] = pushPowerUpPermanent(g, me.ID, "One", "Creature — Bear", "{1}", "")
	ids["two"] = pushPowerUpPermanent(g, opp.ID, "Two", "Creature — Bear", "{1}{G}", "")
	ids["three"] = pushPowerUpPermanent(g, me.ID, "Three", "Creature — Bear", "{2}{G}", "")
	ids["zero"] = b12Creature(g, opp.ID, "Zero", "Creature — Bear", 2, 2)
	ids["rock"] = pushPowerUpPermanent(g, me.ID, "Odd Rock", "Artifact", "{1}", "")
	return ids
}

func spmB15Extinction(t *testing.T, g *game.Game, option int) {
	t.Helper()
	me := g.Seats[0]
	castCatalogSpell(t, g, "Extinction Event", "Sorcery", spmB15ExtinctionOracle, nil)
	passPriorityAroundTable(t, g)
	c := latestChoiceOfKindFor(g, game.PendingChoiceOptionPick, me.ID)
	if c == nil {
		t.Fatal("Extinction Event did not ask odd or even")
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, option); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
}

func TestSpmB15ExtinctionOddExilesOddCreaturesOfEveryPlayer(t *testing.T) {
	g := newCatalogGame(t)
	ids := spmB15ParityTable(g)
	spmB15Extinction(t, g, 0)
	for name, wantAlive := range map[string]bool{"one": false, "two": true, "three": false, "zero": true, "rock": true} {
		if g.Battlefield.Contains(ids[name]) != wantAlive {
			t.Errorf("odd: %s alive = %v, want %v", name, !wantAlive, wantAlive)
		}
	}
	if !g.Exile.Contains(ids["one"]) || !g.Exile.Contains(ids["three"]) {
		t.Error("the odd creatures are exiled, not destroyed")
	}
}

func TestSpmB15ExtinctionEvenExilesEvenCreaturesIncludingZero(t *testing.T) {
	g := newCatalogGame(t)
	ids := spmB15ParityTable(g)
	spmB15Extinction(t, g, 1)
	for name, wantAlive := range map[string]bool{"one": true, "two": false, "three": true, "zero": false, "rock": true} {
		if g.Battlefield.Contains(ids[name]) != wantAlive {
			t.Errorf("even: %s alive = %v, want %v", name, !wantAlive, wantAlive)
		}
	}
}

func TestSpmB15ExtinctionExilesAnIndestructibleCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tough := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tough", TypeLine: "Creature — Bear", ManaCost: "{1}",
		Keywords: []string{"indestructible"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	spmB15Extinction(t, g, 0)
	if g.Battlefield.Contains(tough) {
		t.Error("exile ignores indestructible")
	}
}

// --- Escape to the Wilds ---

func TestSpmB15EscapeExilesFivePlayableUntilYourNextTurnAndGrantsALandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Two", TypeLine: "Sorcery"},
		game.Card{Name: "Three", TypeLine: "Sorcery"},
		game.Card{Name: "Four", TypeLine: "Sorcery"},
		game.Card{Name: "Five", TypeLine: "Sorcery"},
		game.Card{Name: "Sixth", TypeLine: "Sorcery"},
	)
	castCatalogSpell(t, g, "Escape to the Wilds", "Sorcery", spmB15EscapeOracle, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 5; i++ {
		if !g.Exile.Contains(ids[i]) {
			t.Errorf("card %d of the top five is not exiled", i+1)
		}
		if g.CastPermissionOnCardByIDForEffect(ids[i]) == nil {
			t.Errorf("card %d has no permission to be played", i+1)
		}
	}
	if !me.Library.Contains(ids[5]) {
		t.Error("the sixth card stays in the library")
	}
	// The ordinary land drop plays the exiled Forest; the extra one a
	// land from hand; a third is refused.
	if err := g.CastSpell(me.ID, ids[0], game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("playing the exiled land: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := b02cPlayLandRaw(t, g, "Island"); err != nil {
		t.Fatalf("the additional land drop was refused: %v", err)
	}
	if err := b02cPlayLandRaw(t, g, "Swamp"); err == nil {
		t.Error("a third land in one turn should be refused")
	}
}

// --- Force of Vigor ---

func spmB15VigorTable(t *testing.T) (g *game.Game, me, opp *game.Player, art, ench uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	art = b12Permanent(g, me.ID, "Rock", "Artifact")
	ench = b12Permanent(g, me.ID, "Aura", "Enchantment")
	toMainForCost(t, g)
	return g, me, opp, art, ench
}

func TestSpmB15VigorPitchesAGreenCardOffYourTurnAndDestroysTwo(t *testing.T) {
	g, _, opp, art, ench := spmB15VigorTable(t)
	pitch := handCardFull(opp, "Their Green Card", "Instant", "{G}", "", []string{"G"})
	if _, err := spmB15Cast(g, opp, "Force of Vigor", "Instant", "{2}{G}{G}", spmB15VigorOracle, []string{"G"}, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{pitch},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: art}, {Kind: game.TargetCard, ID: ench}},
	}); err != nil {
		t.Fatalf("pitching on someone else's turn: %v", err)
	}
	if opp.Hand.Contains(pitch) || !g.Exile.Contains(pitch) {
		t.Error("the green card must be exiled as the cost")
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || g.Battlefield.Contains(ench) {
		t.Error("both targets should be destroyed")
	}
}

func TestSpmB15VigorPitchRefusals(t *testing.T) {
	g, me, opp, art, _ := spmB15VigorTable(t)
	tgt := []game.TargetRef{{Kind: game.TargetCard, ID: art}}
	notGreen := handCardFull(opp, "Blue Card", "Instant", "{U}", "", []string{"U"})
	if _, err := spmB15Cast(g, opp, "Force of Vigor", "Instant", "{2}{G}{G}", spmB15VigorOracle, []string{"G"}, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{notGreen}, Targets: tgt,
	}); err == nil {
		t.Error("a nongreen card can't be pitched")
	}
	green := handCardFull(me, "My Green Card", "Instant", "{G}", "", []string{"G"})
	if _, err := spmB15Cast(g, me, "Force of Vigor", "Instant", "{2}{G}{G}", spmB15VigorOracle, []string{"G"}, game.CastSpellParams{
		Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{green}, Targets: tgt,
	}); err == nil {
		t.Error("the pitch needs it to be not your turn")
	}
}

func TestSpmB15VigorRefusesAThirdTargetAndAllowsOne(t *testing.T) {
	g, me, _, art, ench := spmB15VigorTable(t)
	third := b12Permanent(g, me.ID, "Third", "Artifact")
	refs := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	floatMana(t, g, me, "{C}{C}{G}{G}")
	if _, err := spmB15Cast(g, me, "Force of Vigor", "Instant", "{2}{G}{G}", spmB15VigorOracle, []string{"G"}, game.CastSpellParams{
		Strict: true, Targets: refs(art, ench, third),
	}); err == nil {
		t.Error("three targets are one too many")
	}
	if _, err := spmB15Cast(g, me, "Force of Vigor", "Instant", "{2}{G}{G}", spmB15VigorOracle, []string{"G"}, game.CastSpellParams{
		Strict: true, Targets: refs(art),
	}); err != nil {
		t.Fatalf("one target: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(art) || !g.Battlefield.Contains(ench) {
		t.Error("only the chosen target should be destroyed")
	}
}

// --- Fury of the Horde ---

func TestSpmB15FuryPitchesTwoRedCardsUntapsAttackersAndAddsCombatAndMain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	attackWith(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepPostcombatMain)
	r1 := handCardFull(me, "Red One", "Instant", "{R}", "", []string{"R"})
	r2 := handCardFull(me, "Red Two", "Instant", "{R}", "", []string{"R"})
	blue := handCardFull(me, "Blue", "Instant", "{U}", "", []string{"U"})
	fury := handCardFull(me, "Fury of the Horde", "Sorcery", "{5}{R}{R}", spmB15FuryOracle, []string{"R"})

	if err := g.CastSpell(me.ID, fury, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{r1, blue}}); err == nil {
		t.Fatal("a blue card can't be one of the two red cards")
	}
	if err := g.CastSpell(me.ID, fury, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{r1}}); err == nil {
		t.Fatal("one card is not two")
	}
	if err := g.CastSpell(me.ID, fury, game.CastSpellParams{Strict: true, AlternativeCost: "pitch", AltCostIDs: []uuid.UUID{r1, r2}}); err != nil {
		t.Fatalf("pitching two red cards: %v", err)
	}
	if me.Hand.Contains(r1) || me.Hand.Contains(r2) || !g.Exile.Contains(r1) || !g.Exile.Contains(r2) {
		t.Error("both red cards are exiled as the cost")
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, bear) {
		t.Error("the creature that attacked this turn is still tapped")
	}
	next := stepOnce(t, g)
	if next.Step != game.StepBeginCombat || next.PhaseOrdinal != 2 {
		t.Fatalf("after Fury's main phase: %+v, want the added combat", next)
	}
	advanceTo(t, g, game.StepPostcombatMain)
	if g.Turn.PhaseOrdinal != 3 {
		t.Errorf("the added main phase reads ordinal %d, want 3", g.Turn.PhaseOrdinal)
	}
}

// --- Leyline Binding ---

func TestSpmB15BindingCostsOneLessPerBasicLandTypeAmongYourLands(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const tl = "Enchantment"
	if got := selfPricedMV(t, g, me, spmB15BindingOracle, tl, "{5}{W}"); got != 6 {
		t.Fatalf("with no lands it costs %d, want 6", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	pushPermanent(g, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	pushPermanent(g, me.ID, game.Card{Name: "Island", TypeLine: "Basic Land — Island"})
	pushPermanent(g, me.ID, game.Card{Name: "Cave", TypeLine: "Land — Cave"})
	pushPermanent(g, opp.ID, game.Card{Name: "Swamp", TypeLine: "Basic Land — Swamp"})
	if got := selfPricedMV(t, g, me, spmB15BindingOracle, tl, "{5}{W}"); got != 4 {
		t.Errorf("with Plains and Island (and an opponent's Swamp) it costs %d, want 4", got)
	}
	pushPermanent(g, me.ID, game.Card{Name: "Dual", TypeLine: "Land — Swamp Mountain Forest"})
	if got := selfPricedMV(t, g, me, spmB15BindingOracle, tl, "{5}{W}"); got != 1 {
		t.Errorf("with all five types it costs %d, want 1", got)
	}
}

func TestSpmB15BindingHasFlash(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	if _, err := spmB15Cast(g, opp, "Plain Enchantment", "Enchantment", "", "", nil, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("control: an enchantment without flash can't be cast on my turn")
	}
	if _, err := spmB15Cast(g, opp, "Leyline Binding", "Enchantment", "", spmB15BindingOracle, nil, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("flash during my turn: %v", err)
	}
	_ = me
}

func TestSpmB15BindingExilesAnOpponentsNonlandPermanentUntilItLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := b12Creature(g, opp.ID, "Their Threat", "Creature — Bear", 4, 4)
	light := castCatalogSpell(t, g, "Leyline Binding", "Enchantment", spmB15BindingOracle, nil)
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, victim)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Fatal("the creature is not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(light) })
	g.RunStateChecksForTest()
	if n := namedOnBattlefield(g, "Their Threat", opp.ID); n != 1 {
		t.Errorf("after Leyline Binding left, %d copies under its owner's control, want 1", n)
	}
}

func TestSpmB15BindingOnlyTargetsOpponentsNonlandPermanents(t *testing.T) {
	spec, _ := Lookup(spmB15BindingOracle)
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirLand := b12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rfPermAExpectLegal(t, g, me.ID, spec.Triggered[0].Targets, map[uuid.UUID]bool{mine: false, theirLand: false, theirs: true})
}

// --- Light of Promise ---

func spmB15Gain(g *game.Game, who uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, who, n) })
}

func TestSpmB15PromisePutsThatManyCountersWheneverYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Light of Promise", "Enchantment — Aura", spmB15PromiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Fatal("the Aura did not enter")
	}
	spmB15Gain(g, me.ID, 3)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 3 {
		t.Fatalf("after gaining 3: %d counters, want 3", got)
	}
	spmB15Gain(g, opp.ID, 5)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 3 {
		t.Errorf("an opponent's life gain gave counters: %d", got)
	}
	spmB15Gain(g, me.ID, 2)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 5 {
		t.Errorf("after gaining 2 more: %d counters, want 5", got)
	}
}

func TestSpmB15PromiseStopsWhenTheAuraLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	aura := castCatalogSpell(t, g, "Light of Promise", "Enchantment — Aura", spmB15PromiseOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	g.RunStateChecksForTest()
	spmB15Gain(g, me.ID, 4)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 0 {
		t.Errorf("without the Aura the creature has %d counters", got)
	}
}

func TestSpmB15PromiseRefusesANonCreatureTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	toMainForCost(t, g)
	if _, err := spmB15Cast(g, me, "Light of Promise", "Enchantment — Aura", "", spmB15PromiseOracle, nil, game.CastSpellParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Error("Enchant creature can't enchant an artifact")
	}
}

// --- Lorthos, the Tidemaker ---

func spmB15LorthosAttack(t *testing.T, g *game.Game, targets ...uuid.UUID) (lorthos uuid.UUID, payPrompt *game.PendingChoice) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	lorthos = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Lorthos, the Tidemaker", TypeLine: "Legendary Creature — Octopus",
		OracleID: spmB15LorthosOracle, Power: 8, Toughness: 8, Owner: me.ID, Controller: me.ID,
	})
	declareAttack(t, g, opp.ID, lorthos)
	mscM2PickTargets(t, g, targets...)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == me.ID {
			payPrompt = c
		}
	}
	if payPrompt == nil {
		t.Fatal("Lorthos did not ask whether to pay {8}")
	}
	return lorthos, payPrompt
}

func TestSpmB15LorthosPayingTapsTheTargetsAndKeepsThemTapped(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	bystander := b12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	_, prompt := spmB15LorthosAttack(t, g, mine, theirs, rock)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("paying {8}: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "rock": rock} {
		if !isTapped(g, id) {
			t.Errorf("%s was not tapped", name)
		}
	}
	if isTapped(g, bystander) {
		t.Error("an untargeted permanent was tapped")
	}
	advanceToUpkeepOf(t, g, 1)
	if !isTapped(g, theirs) || !isTapped(g, rock) {
		t.Error("the opponent's tapped permanents untapped during their next untap step")
	}
}

func TestSpmB15LorthosDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	_, prompt := spmB15LorthosAttack(t, g, theirs)
	if err := g.ResolvePayUnless(prompt.ID, g.Seats[0].ID, false); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if isTapped(g, theirs) {
		t.Error("the target was tapped without the payment")
	}
}

// --- Iron Man, Titan of Innovation ---

func spmB15IronManAttack(t *testing.T, g *game.Game) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	iron := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Iron Man, Titan of Innovation", TypeLine: "Legendary Artifact Creature — Human Hero",
		OracleID: spmB15IronManOracle, ManaCost: "{3}{U}{R}", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	declareAttack(t, g, opp.ID, iron)
	passPriorityAroundTable(t, g)
}

func spmB15Treasure(g *game.Game, me *game.Player) uuid.UUID {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Treasure" {
			return c.InstanceID
		}
	}
	return uuid.Nil
}

func TestSpmB15IronManSacrificingTheTreasureFetchesAOneDropTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Zero", TypeLine: "Artifact", ManaCost: ""},
		game.Card{Name: "Pennant", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Bauble", TypeLine: "Artifact", ManaCost: "{2}"},
	)
	spmB15IronManAttack(t, g)
	treasure := spmB15Treasure(g, me)
	if treasure == uuid.Nil {
		t.Fatal("no Treasure was created")
	}
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Iron Man did not offer the sacrifice")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{treasure}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if g.Battlefield.Contains(treasure) {
		t.Fatal("the Treasure was not sacrificed")
	}
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, ids[1])
	if !ok {
		t.Fatal("the mana value 1 artifact did not enter")
	}
	if !got.Tapped {
		t.Error("it enters tapped")
	}
	if g.Battlefield.Contains(ids[0]) || g.Battlefield.Contains(ids[2]) {
		t.Error("only the mana value 1 artifact is fetched")
	}
}

func TestSpmB15IronManSacrificingATwoDropFetchesAThreeDropAndNoCreatureIsOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	two := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Two Rock", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID,
	})
	golem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem", ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	ids := seedSearchLibrary(me,
		game.Card{Name: "Pennant", TypeLine: "Artifact", ManaCost: "{1}"},
		game.Card{Name: "Three Creature", TypeLine: "Artifact Creature — Golem", ManaCost: "{3}", Power: 3, Toughness: 3},
	)
	spmB15IronManAttack(t, g)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Iron Man did not offer the sacrifice")
	}
	if hasID(c.ChooseCards, golem) {
		t.Error("a creature is not a noncreature artifact")
	}
	if !hasID(c.ChooseCards, two) {
		t.Error("the noncreature artifact is offered")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{two}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the mana value 3 artifact (a creature is fine) did not enter")
	}
	if g.Battlefield.Contains(ids[0]) {
		t.Error("the mana value 1 artifact should stay in the library")
	}
}

func TestSpmB15IronManMayDeclineTheSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me, game.Card{Name: "Pennant", TypeLine: "Artifact", ManaCost: "{1}"})
	spmB15IronManAttack(t, g)
	treasure := spmB15Treasure(g, me)
	c := chooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("Iron Man did not offer the sacrifice")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(treasure) {
		t.Error("the Treasure should remain")
	}
	if g.Battlefield.Contains(ids[0]) || searchChoiceFor(g, me.ID) != nil {
		t.Error("no search without the sacrifice")
	}
}

// --- Don't Move ---

func spmB15TappedCreature(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	id := b12Creature(g, owner, name, "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
	return id
}

func TestSpmB15DontMoveDestroysTappedCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myTapped := spmB15TappedCreature(g, me.ID, "My Tapped")
	theirTapped := spmB15TappedCreature(g, opp.ID, "Their Tapped")
	theirUntapped := b12Creature(g, opp.ID, "Their Untapped", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Don't Move", "Sorcery", spmB15DontMoveOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myTapped) || g.Battlefield.Contains(theirTapped) {
		t.Error("tapped creatures, mine included, are destroyed")
	}
	if !g.Battlefield.Contains(theirUntapped) {
		t.Error("an untapped creature survives the sweep")
	}
}

func TestSpmB15DontMoveDestroysACreatureThatBecomesTappedUntilYourNextTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	castCatalogSpell(t, g, "Don't Move", "Sorcery", spmB15DontMoveOracle, nil)
	passPriorityAroundTable(t, g)
	tapped := b12Creature(g, opp.ID, "Tapped By Effect", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("a creature that becomes tapped is destroyed")
	}
	// The caster's own creature is not spared, and a noncreature is not hit.
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(rock) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("a tapped artifact that is not a creature is not destroyed")
	}
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(mine) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the caster's own creature is destroyed too")
	}
}

func TestSpmB15DontMoveDestroysAnAttackerButNotAVigilantOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushVanillaCreature(g, opp.ID, "Plain", 2, 2)
	vigil := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Vigilant", TypeLine: "Creature — Bear", Keywords: []string{"vigilance"},
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	castCatalogSpell(t, g, "Don't Move", "Sorcery", spmB15DontMoveOracle, nil)
	passPriorityAroundTable(t, g)
	// It is still my turn: the opponent's creatures attack me next turn.
	advanceToUpkeepOf(t, g, 1)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{plain, vigil} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(plain) {
		t.Error("an attacker becomes tapped and is destroyed")
	}
	if !g.Battlefield.Contains(vigil) {
		t.Error("a vigilant attacker never becomes tapped")
	}
}

func TestSpmB15DontMoveEndsWhenYourNextTurnBegins(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Don't Move", "Sorcery", spmB15DontMoveOracle, nil)
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 0)
	late := b12Creature(g, me.ID, "Late", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(late) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(late) {
		t.Error("the effect ended as its controller's next turn began")
	}
}
