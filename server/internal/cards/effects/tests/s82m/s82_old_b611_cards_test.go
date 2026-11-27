package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b611_cards_test.go — Eighth Edition, damage (S82 OLD batch 611,
// #5607). Every test plays the real catalog entry through the engine.

const (
	oldB611Sizzle     = "938cc97f-3908-401e-94cd-32004aff1ef1"
	oldB611ManaClash  = "a2f5a5fd-14dc-48e8-bdb0-984a83288023"
	oldB611Artillery  = "c000811e-bde7-4840-a3bb-9714d5c977eb"
	oldB611Crossbow   = "dd99d791-3305-4195-ad12-96d958a28764"
	oldB611UrzasArmor = "e5fa232a-c8f2-4532-88bf-60d1223181b0"
	oldB611Larceny    = "f356087e-e1d2-4c25-aa1d-215d09535e7c"
	oldB611Karma      = "fac4da47-f0b2-4b57-9703-e0ed100d3499"
	oldB611ShockTroop = "fdf2a8da-2933-44c5-b483-f035144da752"
)

func TestS82OldB611EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB611Sizzle: "Sizzle", oldB611ManaClash: "Mana Clash", oldB611Artillery: "Orcish Artillery",
		oldB611Crossbow: "Crossbow Infantry", oldB611UrzasArmor: "Urza's Armor", oldB611Larceny: "Larceny",
		oldB611Karma: "Karma", oldB611ShockTroop: "Shock Troops",
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

func TestS82OldB611SizzleHitsEachOpponentForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Sizzle", "Sorcery", oldB611Sizzle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID] - 3
		if p.ID == me.ID {
			want = lives[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %v life %d, want %d", p.ID, p.Life, want)
		}
	}
}

func TestS82OldB611ManaClashRepeatsUntilBothFlipHeadsTogether(t *testing.T) {
	for i := 0; i < 12; i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		bystander := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
		myLife, oppLife, byLife := me.Life, opp.Life, bystander.Life
		before := len(effectstest.RandomEvents(g, game.EventFlipCoin))
		effectstest.CastCatalogSpell(t, g, "Mana Clash", "Sorcery", oldB611ManaClash, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)[before:]
		if len(flips) < 2 || len(flips)%2 != 0 {
			t.Fatalf("%d coins flipped, want whole rounds of two", len(flips))
		}
		myTails, oppTails := 0, 0
		for r := 0; r < len(flips); r += 2 {
			a, b := flips[r], flips[r+1]
			if a.Actor != me.ID || b.Actor != opp.ID {
				t.Fatalf("round %d: flips by %v and %v, want the caster then the opponent", r/2, a.Actor, b.Actor)
			}
			bothHeads := a.Label == "heads" && b.Label == "heads"
			last := r+2 == len(flips)
			if bothHeads != last {
				t.Fatalf("round %d of %d: both heads = %v, want the process to stop exactly on the first such round", r/2, len(flips)/2, bothHeads)
			}
			if a.Label == "tails" {
				myTails++
			}
			if b.Label == "tails" {
				oppTails++
			}
		}
		if got := myLife - me.Life; got != myTails {
			t.Errorf("the caster took %d, want %d (one per tails)", got, myTails)
		}
		if got := oppLife - opp.Life; got != oppTails {
			t.Errorf("the opponent took %d, want %d (one per tails)", got, oppTails)
		}
		if bystander.Life != byLife {
			t.Errorf("an uninvolved player lost %d", byLife-bystander.Life)
		}
	}
}

func TestS82OldB611ManaClashOnlyTargetsAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	if err := effectstest.CastCatalogSpellErr(t, g, "Mana Clash", "Sorcery", oldB611ManaClash, []game.TargetRef{effectstest.PlayerTarget(me.ID)}); err == nil {
		t.Error("Mana Clash targeted its own caster")
	}
}

func TestS82OldB611OrcishArtilleryHitsATargetForTwoAndYouForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	art := effectstest.B12Push(g, me.ID, "Orcish Artillery", "Creature — Orc Warrior", oldB611Artillery, 1, 3)
	myLife, oppLife := me.Life, opp.Life
	effectstest.B16Activate(t, g, me.ID, art, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife-3 {
		t.Errorf("lives me %d→%d, them %d→%d; want −3 and −2", myLife, me.Life, oppLife, opp.Life)
	}
	// It tapped as the cost, so it cannot ping again this turn.
	if err := g.ActivateCatalogAbility(me.ID, art, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestS82OldB611CrossbowInfantryPingsOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bow := effectstest.B12Push(g, me.ID, "Crossbow Infantry", "Creature — Human Soldier Archer", oldB611Crossbow, 1, 1)
	attacker := effectstest.B12Creature(g, me.ID, "My Attacker", "Creature — Bear", 1, 1)
	idle := effectstest.B12Creature(g, opp.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Fatal("pinged a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, bow, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: attacker}}}); err != nil {
		t.Fatalf("activate on an attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, attacker); ok {
		t.Error("1 damage did not kill the attacking 1/1")
	}
}

func TestS82OldB611UrzasArmorPreventsOneFromAnySourceToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Urza's Armor", "Artifact", oldB611UrzasArmor, false)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	deal := func(src uuid.UUID, target *game.Player, n int) int {
		before := target.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, target.ID, n) })
		return before - target.Life
	}
	if got := deal(ogre, me, 4); got != 3 {
		t.Errorf("4 from a creature became %d, want 3", got)
	}
	if got := deal(mine, me, 1); got != 0 {
		t.Errorf("1 from my own creature became %d, want 0", got)
	}
	if got := deal(uuid.Nil, me, 4); got != 3 {
		t.Errorf("4 from a noncreature source became %d, want 3", got)
	}
	if got := deal(ogre, opp, 4); got != 4 {
		t.Errorf("damage to another player became %d, want 4", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(ogre, mine, 2) })
	if got := effectstest.DamageMarkedOn(g, mine); got != 2 {
		t.Errorf("damage to my creature: %d marked, want 2 (not prevented)", got)
	}
}

func TestS82OldB611LarcenyMakesTheHurtPlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Larceny", "Enchantment", oldB611Larceny, false)
	thief := effectstest.B12Creature(g, me.ID, "Thief", "Creature — Rogue", 2, 2)
	opp.Hand.Cards = nil
	keep := effectstest.PushCatalogHandCard(opp, "Keep", "Instant", "")
	toss := effectstest.PushCatalogHandCard(opp, "Toss", "Instant", "")
	effectstest.DealCombatDamageToPlayer(g, thief, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, toss)
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Hand.Cards) != 1 || !effectstest.HandHasID(opp, keep) {
		t.Errorf("hand after the discard: %d cards, want just the one they kept", len(opp.Hand.Cards))
	}
}

func TestS82OldB611LarcenyIgnoresAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Larceny", "Enchantment", oldB611Larceny, false)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Thief", "Creature — Rogue", 2, 2)
	effectstest.PushCatalogHandCard(me, "Mine", "Instant", "")
	hand := len(me.Hand.Cards)
	effectstest.DealCombatDamageToPlayer(g, theirs, me.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("an opponent's creature made me discard: %d → %d", hand, len(me.Hand.Cards))
	}
}

func TestS82OldB611KarmaHurtsEachPlayerForTheirSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Karma", "Enchantment", oldB611Karma, false)
	effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, opp.ID, "Swamp", "Basic Land — Swamp")
	effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.PushLandFor(g, me.ID, "Swamp", "Basic Land — Swamp")
	oppLife, myLife := opp.Life, me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("their upkeep: life %d → %d, want −2 for two Swamps", oppLife, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("my life moved on their upkeep: %d → %d", myLife, me.Life)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 {
		t.Errorf("my upkeep: life %d → %d, want −1 for my one Swamp", myLife, me.Life)
	}
}

func TestS82OldB611KarmaDealsNothingWithoutSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Karma", "Enchantment", oldB611Karma, false)
	effectstest.PushLandFor(g, opp.ID, "Island", "Basic Land — Island")
	life := opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d with no Swamps", life, opp.Life)
	}
}

func TestS82OldB611ShockTroopsSacrificesForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	troops := effectstest.B12Push(g, me.ID, "Shock Troops", "Creature — Human Soldier", oldB611ShockTroop, 2, 2)
	life := opp.Life
	effectstest.B16Activate(t, g, me.ID, troops, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(troops) {
		t.Error("the creature was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
}
