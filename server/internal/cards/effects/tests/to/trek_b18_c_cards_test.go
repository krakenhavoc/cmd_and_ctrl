package teststo

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b18_c_cards_test.go — Star Trek, batch 18 (#4973), part C: Worf,
// Intelligence Contractor and The Wrath of Khan. Each test plays the real
// catalog entry and asserts what its text says.

const (
	trekB18CWorfOracle  = "dfff6f8f-b038-4619-8190-7a6918afab1e"
	trekB18CWrathOracle = "22249955-0741-470e-bc2b-f32d4510849b"
)

func TestTrekB18CEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB18CWorfOracle:  "Worf, Intelligence Contractor",
		trekB18CWrathOracle: "The Wrath of Khan",
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

// --- fixtures ------------------------------------------------------------------

// trekB18CCreature seeds a creature with a mana cost on the battlefield.
func trekB18CCreature(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func trekB18CTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

// trekB18COpponents is every seat but the active one, in seat order.
func trekB18COpponents(g *game.Game) (*game.Player, []*game.Player) {
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	return me, opps
}

func trekB18CControllerOf(t *testing.T, g *game.Game, id uuid.UUID) uuid.UUID {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Controller
}

// trekB18CTop stacks cards on top of p's library and returns their ids.
func trekB18CTop(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		id := uuid.New()
		p.Library.PushTop(game.Card{InstanceID: id, Name: names[i], TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		ids[i] = id
	}
	return ids
}

// trekB18CHand puts a card in p's hand and returns its id.
func trekB18CHand(p *game.Player, name string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	return id
}

// --- Worf, Intelligence Contractor ------------------------------------------------

func TestTrekB18CWorfTakesACheapCreatureFromEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := trekB18COpponents(g)
	three := trekB18CCreature(g, opps[0].ID, "Three Drop", "{2}{R}", 3, 3)
	four := trekB18CCreature(g, opps[0].ID, "Four Drop", "{3}{R}", 4, 4)
	one := trekB18CCreature(g, opps[1].ID, "One Drop", "{G}", 1, 1)
	trekB18CTap(g, three)
	effectstest.CastHoldCreature(t, g, "Worf, Intelligence Contractor", "Legendary Creature — Klingon Warrior Spy", trekB18CWorfOracle, 3, 2)

	ch := effectstest.PickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, three) || slices.Contains(ch.PickTargetCards, four) ||
		slices.Contains(ch.PickTargetCards, one) || ch.PickTargetMin != 0 || ch.PickTargetMax != 1 {
		t.Fatalf("the first clause offers only the first opponent's creatures with mana value 3 or less, up to one: %v (%d..%d)",
			ch.PickTargetCards, ch.PickTargetMin, ch.PickTargetMax)
	}
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: four}); err == nil {
		t.Fatal("a mana value 4 creature was accepted as a target")
	}
	effectstest.AnswerPickTarget(t, g, three)
	effectstest.AnswerPickTarget(t, g, one)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{three, one} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s is not controlled by Worf's controller", id)
			continue
		}
		if c.Tapped {
			t.Errorf("%s was not untapped", id)
		}
		if !effectstest.ContainsString(c.Effective().Abilities, "haste") {
			t.Errorf("%s has no haste", id)
		}
		if got := effectstest.EffectiveSubtypes(t, g, id); !effectstest.ContainsString(got, "Warrior") || !effectstest.ContainsString(got, "Bear") {
			t.Errorf("%s should be a Warrior in addition to its other types: %v", id, got)
		}
	}
	if trekB18CControllerOf(t, g, four) != opps[0].ID {
		t.Error("the mana value 4 creature changed hands")
	}

	// Until end of turn: all of it ends at cleanup.
	effectstest.AdvanceToPrecombatMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	c, _ := effectstest.BattlefieldCard(g, three)
	if c.Controller != opps[0].ID || effectstest.ContainsString(c.Effective().Abilities, "haste") || effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, three), "Warrior") {
		t.Errorf("the theft did not end with the turn: controller %v, abilities %v, subtypes %v",
			c.Controller, c.Effective().Abilities, effectstest.EffectiveSubtypes(t, g, three))
	}
}

func TestTrekB18CWorfMayTakeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opps := trekB18COpponents(g)
	big := trekB18CCreature(g, opps[0].ID, "Big", "{4}{R}", 5, 5)
	effectstest.CastHoldCreature(t, g, "Worf, Intelligence Contractor", "Legendary Creature — Klingon Warrior Spy", trekB18CWorfOracle, 3, 2)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatalf("a prompt was opened with no legal creature: %v", c.PickTargetCards)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if trekB18CControllerOf(t, g, big) != opps[0].ID {
		t.Error("a mana value 5 creature was taken")
	}
}

// --- The Wrath of Khan -----------------------------------------------------------

func TestTrekB18CWrathOfKhanChapterOneDamagesThenStealsAtCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := trekB18COpponents(g)
	big := trekB18CCreature(g, opps[0].ID, "Big", "{4}{G}", 4, 6)
	small := trekB18CCreature(g, opps[1].ID, "Small", "{G}", 2, 2)
	mine := trekB18CCreature(g, me.ID, "Mine", "{G}", 2, 2)
	trekB18CTap(g, big)

	effectstest.CastCatalogSpell(t, g, "The Wrath of Khan", "Enchantment — Saga", trekB18CWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if slices.Contains(ch.PickTargetCards, mine) || !slices.Contains(ch.PickTargetCards, big) || !slices.Contains(ch.PickTargetCards, small) {
		t.Fatalf("chapter I offers only creatures an opponent controls: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, big)
	effectstest.PassPriorityAroundTable(t, g)

	c, ok := effectstest.BattlefieldCard(g, big)
	if !ok || c.DamageMarked != 4 {
		t.Fatalf("Big should have survived with 4 damage marked: on battlefield %v, damage %d", ok, c.DamageMarked)
	}
	if c.Controller != opps[0].ID {
		t.Fatal("control changed before combat")
	}

	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ = effectstest.BattlefieldCard(g, big)
	if c.Controller != me.ID || c.Tapped || !effectstest.ContainsString(c.Effective().Abilities, "haste") {
		t.Fatalf("at the beginning of combat: controller mine %v, tapped %v, abilities %v",
			c.Controller == me.ID, c.Tapped, c.Effective().Abilities)
	}
	if trekB18CControllerOf(t, g, small) != opps[1].ID {
		t.Error("an untargeted creature changed hands")
	}

	// Until end of turn, and the delayed trigger is used up.
	effectstest.AdvanceToPrecombatMainOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	if trekB18CControllerOf(t, g, big) != opps[0].ID {
		t.Error("control did not end with the turn")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers outlived the turn", len(g.DelayedTriggers))
	}
}

func TestTrekB18CWrathOfKhanDeadCreatureIsNotTaken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := trekB18COpponents(g)
	small := trekB18CCreature(g, opps[0].ID, "Small", "{G}", 2, 2)
	other := trekB18CCreature(g, opps[0].ID, "Other", "{G}", 2, 6)

	effectstest.CastCatalogSpell(t, g, "The Wrath of Khan", "Enchantment — Saga", trekB18CWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Fatal("4 damage did not kill a 2/2")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if trekB18CControllerOf(t, g, other) != opps[0].ID {
		t.Error("the delayed trigger took a creature it does not name")
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Owner != me.ID {
			t.Errorf("%s was taken", c.Name)
		}
	}
}

func TestTrekB18CWrathOfKhanAfterCombatWaitsForNoOtherTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opps := trekB18COpponents(g)
	big := trekB18CCreature(g, opps[0].ID, "Big", "{4}{G}", 4, 6)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.CastCatalogSpell(t, g, "The Wrath of Khan", "Enchantment — Saga", trekB18CWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, big)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, big); c.DamageMarked != 4 {
		t.Fatalf("damage %d, want 4", c.DamageMarked)
	}
	// The next combat is on the next player's turn: not "this turn".
	seat := g.Turn.ActiveSeat
	for g.Turn.ActiveSeat == seat || g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if trekB18CControllerOf(t, g, big) == me.ID {
		t.Error("the delayed trigger fired at a combat on another turn")
	}
	if len(g.DelayedTriggers) != 0 {
		t.Errorf("%d delayed triggers outlived the turn", len(g.DelayedTriggers))
	}
}

func TestTrekB18CWrathOfKhanChaptersTwoAndThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opps := trekB18COpponents(g)
	// Chapter I needs a target to go on the stack; it is not the point here.
	victim := trekB18CCreature(g, opps[2].ID, "Victim", "{G}", 1, 1)
	saga := effectstest.CastCatalogSpell(t, g, "The Wrath of Khan", "Enchantment — Saga", trekB18CWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)

	// Chapter II: exile a card from hand, draw a card.
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	keep := trekB18CHand(me, "Keep")
	toss := trekB18CHand(me, "Toss")
	top := trekB18CTop(me, "Top Card")
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Fatalf("chapter II should ask for up to one card of your hand: %+v", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{toss}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(toss) || !me.Hand.Contains(keep) || !me.Hand.Contains(top[0]) {
		t.Fatalf("chapter II: tossed in exile %v, kept in hand %v, drew the top card %v",
			g.Exile.Contains(toss), me.Hand.Contains(keep), me.Hand.Contains(top[0]))
	}

	// Chapter III: each opponent sacrifices a nontoken permanent.
	mineLand := effectstest.SeedPermanentFor(g, me.ID, "My Land", "Land")
	theirLand := effectstest.SeedPermanentFor(g, opps[0].ID, "Their Land", "Land")
	theirOther := effectstest.SeedPermanentFor(g, opps[0].ID, "Their Relic", "Artifact")
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Clue", TypeLine: "Token Artifact — Clue",
		Owner: opps[1].ID, Controller: opps[1].ID})
	twoLand := effectstest.SeedPermanentFor(g, opps[1].ID, "Second Land", "Land")
	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SacrificeChoiceFor(g, opps[1].ID); c != nil && slices.Contains(c.ChooseCards, token) {
		t.Error("a token was offered for a nontoken sacrifice")
	}
	effectstest.AnswerSacrifice(t, g, opps[0].ID, theirLand)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opps[1].ID, twoLand)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirLand) || !g.Battlefield.Contains(theirOther) {
		t.Error("the first opponent did not sacrifice exactly the chosen nontoken permanent")
	}
	if g.Battlefield.Contains(twoLand) || !g.Battlefield.Contains(token) {
		t.Error("the second opponent's nontoken permanent should go and their token stay")
	}
	if !g.Battlefield.Contains(mineLand) {
		t.Error("the Saga's controller sacrificed a permanent")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

func TestTrekB18CWrathOfKhanChapterTwoMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opps := trekB18COpponents(g)
	victim := trekB18CCreature(g, opps[2].ID, "Victim", "{G}", 1, 1)
	effectstest.CastCatalogSpell(t, g, "The Wrath of Khan", "Enchantment — Saga", trekB18CWrathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.AdvanceToPrecombatMainOf(t, g, seat)
	keep := trekB18CHand(me, "Keep")
	handBefore := len(me.Hand.Cards)
	top := trekB18CTop(me, "Top Card")
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II asked nothing")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("declining: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) || me.Hand.Contains(top[0]) || len(me.Hand.Cards) != handBefore {
		t.Errorf("declining exiled or drew: hand %d (was %d), drew the top card %v", len(me.Hand.Cards), handBefore, me.Hand.Contains(top[0]))
	}
}
