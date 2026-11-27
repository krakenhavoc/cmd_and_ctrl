package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b745_test.go — S82 OLD batch 745 (Exodus, #5626), through the real
// engine: Recurring Nightmare, Oath of Ghouls, Mirozel, Exalted Dragon and
// Memory Crystal.

const (
	oldB745Nightmare = "a6708b11-1bcd-4208-a967-fe91f2e3313c"
	oldB745Ghouls    = "b9490b52-6c3f-437b-b2c2-9e81c4b43b61"
	oldB745Mirozel   = "c35111ce-4325-4547-b954-08275a268060"
	oldB745Dragon    = "c901805e-c196-4b9d-8d5a-e17367fb37bc"
	oldB745Crystal   = "f5fd3141-6fd9-4491-9f29-7fd81209bd57"
	oldB745Mystic    = "a58adf82-d9e2-4b47-933e-68180f280ba2"
)

func TestS82OldB745CardsAreFull(t *testing.T) {
	for _, id := range []string{oldB745Nightmare, oldB745Ghouls, oldB745Mirozel, oldB745Dragon, oldB745Crystal} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82OldB745RecurringNightmareSacrificesReturnsItselfAndReanimates(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	nightmare := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Recurring Nightmare", OracleID: oldB745Nightmare, TypeLine: "Enchantment"})
	fodder := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	dead := uuid.New()
	spell := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Big Dead", TypeLine: "Creature — Test", ManaCost: "{4}{B}", Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dead Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	})

	// A noncreature card is not a legal target; nothing is paid.
	err := g.ActivateCatalogAbility(me.ID, nightmare, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: spell}}})
	if err == nil {
		t.Fatal("a sorcery card was accepted as the target")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(nightmare) {
		t.Fatal("a refused activation paid its costs")
	}

	effectstest.P7Activate(t, g, me, nightmare, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}})
	if !g.Battlefield.Contains(dead) {
		t.Error("the creature card did not return to the battlefield")
	}
	if !me.Hand.Contains(nightmare) {
		t.Error("Recurring Nightmare is not in its owner's hand")
	}
	if g.Battlefield.Contains(fodder) || !me.Graveyard.Contains(fodder) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

func TestS82OldB745RecurringNightmareIsASorceryAndNeedsACreatureToSacrifice(t *testing.T) {
	g, me, _ := effectstest.P7Table(t)
	nightmare := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Recurring Nightmare", OracleID: oldB745Nightmare, TypeLine: "Enchantment"})
	dead := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Big Dead", TypeLine: "Creature — Test", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	})
	target := []game.TargetRef{{Kind: game.TargetCard, ID: dead}}
	if err := g.ActivateCatalogAbility(me.ID, nightmare, 0, game.ActivateAbilityParams{Targets: target}); err == nil {
		t.Fatal("activated with no creature to sacrifice")
	}
	fodder := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Fodder", TypeLine: "Creature — Test", Power: 1, Toughness: 1})
	// Not at sorcery timing: an opponent's turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := g.ActivateCatalogAbility(me.ID, nightmare, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: target}); err == nil {
		t.Fatal("activated during an opponent's turn")
	}
	if !g.Battlefield.Contains(fodder) || !g.Battlefield.Contains(nightmare) {
		t.Fatal("a refused activation paid its costs")
	}
}

func TestS82OldB745MirozelReturnsToHandWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := effectstest.B12Push(g, me.ID, "Mirozel", "Creature — Illusion", oldB745Mirozel, 2, 3)
	if !effectstest.HasKeywordOnBattlefield(t, g, m, "flying") {
		t.Error("Mirozel lacks flying")
	}
	effectstest.CastCatalogSpell(t, g, "Pacifism", "Enchantment — Aura", effectstest.PacifismAuraOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: m}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(m) || !me.Hand.Contains(m) {
		t.Errorf("Mirozel: battlefield=%v hand=%v, want it returned to its owner's hand", g.Battlefield.Contains(m), me.Hand.Contains(m))
	}
}

func TestS82OldB745ExaltedDragonCantAttackUnlessYouSacrificeALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.B12Push(g, me.ID, "Exalted Dragon", "Creature — Dragon", oldB745Dragon, 5, 5)
	land := effectstest.B12Permanent(g, me.ID, "Plains", "Basic Land — Plains")
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)

	err := g.DeclareAttacker(dragon, opp.ID)
	if !errors.Is(err, game.ErrCombatCostUnpaid) {
		t.Fatalf("declaring with no land named = %v, want ErrCombatCostUnpaid", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("a refused declaration sacrificed the land")
	}
	// A creature is not a land.
	if err := g.DeclareAttackerWith(dragon, opp.ID, game.DeclareAttackersParams{
		Pay: game.CombatCostPayment{SacrificeIDs: []uuid.UUID{other}}}); err == nil {
		t.Fatal("a creature paid for \"sacrifice a land\"")
	}
	if err := g.DeclareAttackerWith(dragon, opp.ID, game.DeclareAttackersParams{
		Pay: game.CombatCostPayment{SacrificeIDs: []uuid.UUID{land}}}); err != nil {
		t.Fatalf("declaring with the land named: %v", err)
	}
	if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	// Other creatures attack for free.
	if err := g.DeclareAttacker(other, opp.ID); err != nil {
		t.Errorf("a plain creature was taxed: %v", err)
	}
}

func TestS82OldB745MemoryCrystalDiscountsOnlyTheBuybackComponent(t *testing.T) {
	cast := func(crystal bool, mana string, buyback bool) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		if crystal {
			effectstest.PushPermanentForTest(g, g.Seats[1].ID, "Memory Crystal", oldB745Crystal, "Artifact")
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mystic Speculation", TypeLine: "Sorcery", OracleID: oldB745Mystic,
			ManaCost: "{1}{U}", Owner: me.ID, Controller: me.ID})
		effectstest.FloatMana(t, g, me, mana)
		params := game.CastSpellParams{Strict: true}
		if buyback {
			params.OptionalCosts = []int{0}
		}
		return g.CastSpell(me.ID, id, params)
	}
	// Mystic Speculation is {1}{U} with buyback {2}: {3}{U} printed, {1}{U} under the Crystal.
	if err := cast(false, "{U}{C}", true); err == nil {
		t.Error("a bought-back cast with two mana succeeded without the Crystal")
	}
	if err := cast(true, "{U}{C}", true); err != nil {
		t.Errorf("a bought-back cast with two mana under the Crystal (any player's): %v", err)
	}
	// An unbought cast is not discounted: {1}{U} still needs two mana.
	if err := cast(true, "{U}", false); err == nil {
		t.Error("the Crystal discounted a cast that paid no buyback")
	}
}

func TestS82OldB745OathOfGhoulsLetsTheUpkeepPlayerTakeACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep, rich, poor := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Ghouls", oldB745Ghouls, "Enchantment")
	push := func(p *game.Player, name, typeLine string) uuid.UUID {
		id := uuid.New()
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
		return id
	}
	mine := push(upkeep, "Upkeep Bear", "Creature — Bear")
	push(upkeep, "Upkeep Bear Two", "Creature — Bear")
	push(upkeep, "Upkeep Spell", "Sorcery")
	push(rich, "Rich A", "Creature — Bear")
	push(rich, "Rich B", "Creature — Bear")
	push(rich, "Rich C", "Creature — Bear")
	push(owner, "Owner Bear", "Creature — Bear")
	push(poor, "Poor Spell", "Sorcery")
	push(poor, "Poor Spell Two", "Sorcery")
	push(poor, "Poor Spell Three", "Sorcery")

	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetPlayers {
		offered[id] = true
	}
	// Upkeep has 2 creature cards: only players with fewer (owner 1, poor 0).
	if !offered[owner.ID] || !offered[poor.ID] || offered[rich.ID] || offered[upkeep.ID] || len(offered) != 2 {
		t.Fatalf("offered %v, want only the opponents with fewer creature cards in their graveyard", ch.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: rich.ID}); err == nil {
		t.Fatal("a player with more creature cards was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: poor.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, owner.ID); p != nil {
		t.Fatal("the Oath's controller was asked")
	}
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, upkeep.ID)
	if p == nil {
		t.Fatal("the upkeep player was not asked")
	}
	if err := g.ResolveChooseCards(p.ID, upkeep.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !upkeep.Hand.Contains(mine) || upkeep.Graveyard.Contains(mine) {
		t.Error("the chosen creature card did not go to the upkeep player's hand")
	}
}

func TestS82OldB745OathOfGhoulsDoesNothingWithoutAPoorerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Ghouls", oldB745Ghouls, "Enchantment")
	id := uuid.New()
	owner.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Owner Bear", TypeLine: "Creature — Bear", Owner: owner.ID, Controller: owner.ID})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, upkeep.ID); p != nil {
		t.Fatalf("the upkeep player was asked for a target with nobody eligible: %+v", p)
	}
}
