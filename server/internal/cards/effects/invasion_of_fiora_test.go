package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// invasion_of_fiora_test.go — Invasion of Fiora // Marchesa, Resolute
// Monarch (#5652), both faces, and the since-last-turn memory Marchesa's
// upkeep trigger reads (game/since_last_turn.go).

const marchesaResoluteMonarchKey = invasionOfFioraOracleID + "#1"

func fioraRow() cards.Card {
	return siegeRow(invasionOfFioraOracleID, "Invasion of Fiora", "Marchesa, Resolute Monarch",
		"{4}{B}{B}", "4", "Legendary Creature — Human Noble", "3", "6")
}

func TestInvasionOfFioraBothFacesAreRegisteredFull(t *testing.T) {
	for key, name := range map[string]string{
		invasionOfFioraOracleID:    "Invasion of Fiora",
		marchesaResoluteMonarchKey: "Marchesa, Resolute Monarch",
	} {
		spec, ok := Lookup(key)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
	back, _ := Lookup(marchesaResoluteMonarchKey)
	if len(back.PrintedKeywords) != 2 || back.PrintedKeywords[0] != "menace" || back.PrintedKeywords[1] != "deathtouch" {
		t.Errorf("Marchesa's keywords %v, want menace and deathtouch", back.PrintedKeywords)
	}
	rows := game.CatalogTriggers(invasionOfFioraOracleID)
	if len(rows) == 0 || !rows[0].Purpose.DeclaresSweep() {
		t.Error("the Siege's enters trigger does not declare its sweep")
	}
}

// fioraEnter puts the Siege onto the battlefield from the active seat's
// hand, names its protector, and answers the enters trigger with `modes`.
// It returns the Siege, a legendary creature and a nonlegendary creature
// of each of two players.
func fioraEnter(t *testing.T, g *game.Game, modes []int) (siege uuid.UUID, legends, others []uuid.UUID) {
	t.Helper()
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]
	for _, p := range []*game.Player{owner, protector} {
		legends = append(legends, b12Creature(g, p.ID, "Legend", "Legendary Creature — Human Noble", 3, 3))
		others = append(others, b12Creature(g, p.ID, "Bear", "Creature — Bear", 2, 2))
	}
	siege = importToHand(fioraRow(), owner)
	toMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, siege, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Siege: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, owner.ID)
	if c == nil {
		t.Fatal("the enters trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, owner.ID, modes); err != nil {
		t.Fatalf("ResolveModePick(%v): %v", modes, err)
	}
	passPriorityAroundTable(t, g)
	return siege, legends, others
}

func fioraWant(t *testing.T, g *game.Game, what string, ids []uuid.UUID, alive bool) {
	t.Helper()
	for _, id := range ids {
		if got := g.Battlefield.Contains(id); got != alive {
			t.Errorf("%s on the battlefield = %v, want %v", what, got, alive)
		}
	}
}

func TestInvasionOfFioraDestroysOnlyLegendaryCreatures(t *testing.T) {
	g := newCatalogGame(t)
	siege, legends, others := fioraEnter(t, g, []int{0})
	fioraWant(t, g, "a legendary creature", legends, false)
	fioraWant(t, g, "a nonlegendary creature", others, true)
	battle := battlefieldCardFor(g, siege)
	if battle == nil {
		t.Fatal("the Siege is not on the battlefield")
	}
	if got := battle.Counters[game.CounterDefense]; got != 4 {
		t.Errorf("defense = %d, want the printed 4", got)
	}
}

func TestInvasionOfFioraDestroysOnlyNonlegendaryCreatures(t *testing.T) {
	g := newCatalogGame(t)
	_, legends, others := fioraEnter(t, g, []int{1})
	fioraWant(t, g, "a legendary creature", legends, true)
	fioraWant(t, g, "a nonlegendary creature", others, false)
}

func TestInvasionOfFioraBothModesDestroyEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	siege, legends, others := fioraEnter(t, g, []int{0, 1})
	fioraWant(t, g, "a legendary creature", legends, false)
	fioraWant(t, g, "a nonlegendary creature", others, false)
	if battlefieldCardFor(g, siege) == nil {
		t.Error("the Siege is not a creature and stays")
	}
}

// "Choose one or both": no mode, and the same mode twice, are refused.
func TestInvasionOfFioraRefusesNoModeAndARepeatedMode(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	siege := importToHand(fioraRow(), owner)
	toMainPhaseSameTurn(t, g)
	if err := g.CastSpell(owner.ID, siege, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the Siege: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, owner.ID, g.Seats[(seat+1)%len(g.Seats)].ID)
	passPriorityAroundTable(t, g)
	c := modePickChoiceFor(g, owner.ID)
	if c == nil {
		t.Fatal("the enters trigger did not ask for modes")
	}
	if err := g.ResolveModePick(c.ID, owner.ID, nil); err == nil {
		t.Error("no mode was accepted")
	}
	if err := g.ResolveModePick(c.ID, owner.ID, []int{1, 1}); err == nil {
		t.Error("the same mode twice was accepted")
	}
}

// The Siege's defeat grants the cast of its back face (the "#1" key).
func TestInvasionOfFioraDefeatedGrantsMarchesa(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[g.Turn.ActiveSeat]
	siege, _, _ := fioraEnter(t, g, []int{0})
	defeatBattle(t, g, siege, 4)
	if !inExile(g, siege) {
		t.Fatal("the defeated Siege was not exiled")
	}
	grant := exileGrantFor(g, siege)
	if grant.Player != owner.ID {
		t.Errorf("grant holder = %v, want the Siege's controller", grant.Player)
	}
	if face, ok := grant.NamedFace(); !ok || face != 1 {
		t.Errorf("grant faces = %v, want the back face", grant.Faces)
	}
}

// --- Marchesa, Resolute Monarch ---

func pushMarchesa(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushCatalogPermanent(g, owner, "Marchesa, Resolute Monarch", "Legendary Creature — Human Noble", marchesaResoluteMonarchKey, false)
}

func TestMarchesaDrawsAndLosesOneWhenNotDealtCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	marchesa := pushMarchesa(g, me.ID)
	// Noncombat damage is not what the condition asks about.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: marchesa, Target: me.ID, Amount: 1})
	})
	life, hand := me.Life, me.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, marchesa) == nil {
		t.Fatal("no upkeep trigger, though no combat damage was dealt")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
}

// Real combat: seat 0's creature connects with seat 1, and seat 1's next
// upkeep does not trigger. The upkeep after that, with no combat damage
// in between, does.
func TestMarchesaDoesNotTriggerAfterCombatDamageSinceYourLastTurn(t *testing.T) {
	g := newCatalogGame(t)
	if g.Turn.ActiveSeat != 0 {
		t.Fatalf("setup: seat %d is active, want seat 0", g.Turn.ActiveSeat)
	}
	attackerOwner, me := g.Seats[0], g.Seats[1]
	marchesa := pushMarchesa(g, me.ID)
	// Tapped, so Marchesa cannot block.
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(marchesa) })
	bear := b12Creature(g, attackerOwner.ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	trekB25ToCombatDamage(t, g, me.ID, bear)
	if me.Life != life-2 {
		t.Fatalf("setup: life %d, want %d after the Bear connects", me.Life, life-2)
	}
	life, hand := me.Life, me.Hand.Size()
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, marchesa) != nil {
		t.Fatal("the upkeep trigger went on the stack after combat damage since the last turn")
	}
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d hand %d, want %d and %d", me.Life, me.Hand.Size(), life, hand)
	}

	// A full round with no combat damage: seat 1's turn ended, so the
	// span starts over.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, marchesa) == nil {
		t.Fatal("no upkeep trigger a round later, with no combat damage since the last turn")
	}
	hand = me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// Combat damage dealt to another player, or to a creature you control,
// is not combat damage dealt to you.
func TestMarchesaIgnoresCombatDamageDealtToOthers(t *testing.T) {
	g := newCatalogGame(t)
	me, other := g.Seats[1], g.Seats[2]
	marchesa := pushMarchesa(g, me.ID)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, bear, other.ID, 2)
	dealCombatDamageToPlayer(g, bear, marchesa, 1)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, marchesa) == nil {
		t.Error("no upkeep trigger, though the combat damage was dealt to others")
	}
}

// The memory is carried across a restore: a table saved after the damage
// still knows at the next upkeep.
func TestMarchesaConditionSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	marchesa := pushMarchesa(g, me.ID)
	bear := b12Creature(g, g.Seats[0].ID, "Bear", "Creature — Bear", 2, 2)
	dealCombatDamageToPlayer(g, bear, me.ID, 2)
	restored := restoreRoundTrip(t, g, true)
	advanceToUpkeepOf(t, restored, 1)
	if triggerOnStack(restored, marchesa) != nil {
		t.Error("the restored table forgot the combat damage")
	}
}

// CR 603.4: the condition is asked again as the ability resolves.
func TestMarchesaUpkeepConditionIsRecheckedOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	marchesa := pushMarchesa(g, me.ID)
	advanceToUpkeepOf(t, g, 1)
	if triggerOnStack(g, marchesa) == nil {
		t.Fatal("no upkeep trigger")
	}
	life, hand := me.Life, me.Hand.Size()
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Source: marchesa, Target: me.ID, Amount: 1, Combat: true})
	})
	passPriorityAroundTable(t, g)
	if me.Life != life || me.Hand.Size() != hand {
		t.Errorf("life %d hand %d, want %d and %d: the condition no longer held", me.Life, me.Hand.Size(), life, hand)
	}
}

func marchesaAttack(t *testing.T, g *game.Game) (me *game.Player, marchesa, bear uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	marchesa = pushMarchesa(g, me.ID)
	bear = b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bear {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterPlusOne: 2, "stun": 1}
			}
		}
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(marchesa, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	return me, marchesa, bear
}

func TestMarchesaAttackRemovesEveryCounterFromTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me, _, bear := marchesaAttack(t, g)
	pickTriggerTarget(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	c, ok := lookupBattlefield(g, bear)
	if !ok {
		t.Fatal("the Bear left the battlefield")
	}
	for kind, n := range c.Counters {
		if n != 0 {
			t.Errorf("the Bear still has %d %s counters", n, kind)
		}
	}
}

// "Up to one": the trigger may be given no target, and removes nothing.
func TestMarchesaAttackMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, _, bear := marchesaAttack(t, g)
	b26DeclinePickTarget(t, g, me.ID)
	passPriorityAroundTable(t, g)
	c, _ := lookupBattlefield(g, bear)
	if c.Counters[game.CounterPlusOne] != 2 || c.Counters["stun"] != 1 {
		t.Errorf("counters %v, want them untouched", c.Counters)
	}
}

// A card that is not a permanent is not a legal target.
func TestMarchesaAttackRefusesACardInHand(t *testing.T) {
	g := newCatalogGame(t)
	me, _, _ := marchesaAttack(t, g)
	inHand := pushToHand(g, me.ID, "Held Card", "Creature — Bear")
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no pick_target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: inHand}); err == nil {
		t.Error("a card in hand was accepted as a target permanent")
	}
}
