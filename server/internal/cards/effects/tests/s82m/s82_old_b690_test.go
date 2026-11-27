package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b690_test.go — Mercadian Masques damage cards, S82 OLD batch 690
// (#5619), played through the real catalog entries.

const (
	o690Warpath       = "413520c7-5227-406f-8861-e844c153e5b0"
	o690Negotiations  = "50090188-651d-4402-968f-3f6bf9276029"
	o690BallistaSquad = "5973725c-6e1f-4d7e-bcef-3d867ea4c244"
	o690Lunge         = "5e642112-9592-475f-ba03-d71544270cfa"
	o690Brawl         = "6d53f58c-6cf3-4ecc-8227-8c264fecc836"
	o690Sniper        = "7661dda6-4e7b-42d0-8374-0a1fe65b83e4"
	o690CragSaurian   = "7779e3be-5c3d-454f-88c1-9c0a97ebff09"
	o690BloodOath     = "877c34ca-6eff-42ce-b4d7-8a2071d0290a"
	o690Barrier       = "b56f375f-35b3-4d96-9749-c059c2d29c8b"
	o690VolcanicWind  = "baa85b14-1c9f-4a19-9f07-93e3192c597d"
	o690CloseQuarters = "c3f841c4-6e12-46fc-802a-3025182e7ff4"
	o690Wumpus        = "ef220824-fab4-4d8e-9ba4-65ff5ba1db66"
	o690FuriousAssaut = "f5d1a004-5598-4201-ac07-bca6cd23008f"
)

func TestS82OldB690CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o690Warpath: "Warpath", o690Negotiations: "Kyren Negotiations", o690BallistaSquad: "Ballista Squad",
		o690Lunge: "Lunge", o690Brawl: "Brawl", o690Sniper: "Kyren Sniper", o690CragSaurian: "Crag Saurian",
		o690BloodOath: "Blood Oath", o690Barrier: "Stinging Barrier", o690VolcanicWind: "Volcanic Wind",
		o690CloseQuarters: "Close Quarters", o690Wumpus: "Thrashing Wumpus", o690FuriousAssaut: "Furious Assault",
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

func o690Marked(g *game.Game, id uuid.UUID) int {
	if c := effectstest.FindBattlefieldCardForTest(g, id); c != nil {
		return c.DamageMarked
	}
	return -1
}

func o690Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o690Player(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetPlayer, ID: id} }

func TestS82OldB690WarpathHitsBlockersAndBlockedOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blocked := effectstest.PushVanillaCreature(g, me.ID, "Blocked", 0, 9)
	unblocked := effectstest.PushVanillaCreature(g, me.ID, "Unblocked", 0, 9)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 0, 9)
	bystander := effectstest.PushVanillaCreature(g, opp.ID, "Bystander", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range []uuid.UUID{blocked, unblocked} {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, blocked); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.FloatForTest(g, me, "CCCR")
	id := effectstest.HandCardOf(g, me, "Warpath", "Instant", "{3}{R}", o690Warpath)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for who, want := range map[uuid.UUID]int{blocked: 3, blocker: 3, unblocked: 0, bystander: 0} {
		if got := o690Marked(g, who); got != want {
			t.Errorf("damage marked on %v = %d, want %d", who, got, want)
		}
	}
}

func TestS82OldB690KyrenNegotiationsTapsACreatureToPing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Kyren Negotiations", "Enchantment", o690Negotiations, false)
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Fresh Bear", "Creature — Bear", "", true)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{sick}, Targets: []game.TargetRef{o690Player(opp.ID)}}); err != nil {
		t.Fatalf("activate tapping a summoning-sick creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("life = %d, want 39", opp.Life)
	}
	if !effectstest.TappedOnBattlefield(t, g, sick) {
		t.Error("the creature was not tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{sick}, Targets: []game.TargetRef{o690Player(opp.ID)}}); err == nil {
		t.Error("a tapped creature paid the cost again")
	}
}

func TestS82OldB690BallistaSquadDealsXToAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squad := effectstest.PushCatalogPermanent(g, me.ID, "Ballista Squad", "Creature — Human Rebel", o690BallistaSquad, false)
	raider := effectstest.PushVanillaCreature(g, me.ID, "Raider", 0, 9)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.FloatForTest(g, me, "CCCW")
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: o690Target(idle)}); err == nil {
		t.Fatal("Ballista Squad targeted a creature that is neither attacking nor blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, squad, 0, game.ActivateAbilityParams{Strict: true, XValue: 3, Targets: o690Target(raider)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := o690Marked(g, raider); got != 3 {
		t.Errorf("damage on the attacker = %d, want 3", got)
	}
}

func TestS82OldB690LungeDamagesACreatureAndAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, opp.ID, "Bear", 2, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCR")
	id := effectstest.HandCardOf(g, me, "Lunge", "Instant", "{2}{R}", o690Lunge)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: bear}, o690Player(opp.ID)}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := o690Marked(g, bear); got != 2 {
		t.Errorf("creature damage = %d, want 2", got)
	}
	if opp.Life != 38 {
		t.Errorf("life = %d, want 38", opp.Life)
	}
}

func TestS82OldB690BrawlGivesEveryCreatureThePowerPingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 3, 3)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 1, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCRR")
	id := effectstest.HandCardOf(g, me, "Brawl", "Instant", "{3}{R}{R}", o690Brawl)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushVanillaCreature(g, me.ID, "Late", 5, 5)
	if err := g.ActivateCatalogAbility(me.ID, late, 0, game.ActivateAbilityParams{Strict: true, Targets: o690Target(theirs)}); err == nil {
		t.Error("a creature that arrived after Brawl resolved got the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Strict: true, Targets: o690Target(theirs)}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := o690Marked(g, theirs); got != 3 {
		t.Errorf("damage = %d, want 3 (the creature's power)", got)
	}
	if !effectstest.TappedOnBattlefield(t, g, mine) {
		t.Error("the {T} cost was not paid")
	}
}

func TestS82OldB690KyrenSniperMayPingAtUpkeep(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "Kyren Sniper", "Creature — Goblin", o690Sniper, false)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		effectstest.PassPriorityUntilChoice(t, g)
		effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		want := 40
		if yes {
			want = 39
		}
		if opp.Life != want {
			t.Errorf("yes=%v: life = %d, want %d", yes, opp.Life, want)
		}
	}
}

func TestS82OldB690StingingBarrierPingsAnyTargetAndHasDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := effectstest.PushCatalogPermanent(g, me.ID, "Stinging Barrier", "Creature — Wall", o690Barrier, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "U")
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{o690Player(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("life = %d, want 39", opp.Life)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Error("no defender")
	}
}

func TestS82OldB690CloseQuartersPingsWhenYourCreatureBecomesBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Close Quarters", "Enchantment", o690CloseQuarters, false)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 0, 9)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.InitSettle(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("life = %d, want 39", opp.Life)
	}
}

func TestS82OldB690ThrashingWumpusDamagesEverythingIncludingItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wumpus := effectstest.B12Push(g, me.ID, "Thrashing Wumpus", "Creature — Beast", o690Wumpus, 3, 3)
	other := effectstest.PushVanillaCreature(g, opp.ID, "Other", 0, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "B")
	if err := g.ActivateCatalogAbility(me.ID, wumpus, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{wumpus, other} {
		if got := o690Marked(g, id); got != 1 {
			t.Errorf("damage on %v = %d, want 1", id, got)
		}
	}
	if me.Life != 39 || opp.Life != 39 {
		t.Errorf("life %d / %d, want 39 / 39", me.Life, opp.Life)
	}
}

func TestS82OldB690FuriousAssaultPingsOnlyForCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Furious Assault", "Enchantment", o690FuriousAssaut, false)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.HandCardOf(g, me, "Cantrip", "Sorcery", "", "")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Fatalf("a sorcery pinged: life = %d", opp.Life)
	}
	bear := effectstest.HandCardOf(g, me, "Bear", "Creature — Bear", "", "")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 39 {
		t.Errorf("life = %d, want 39", opp.Life)
	}
}

func TestS82OldB690CragSaurianChangesHandsWhenDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saurian := effectstest.B12Push(g, me.ID, "Crag Saurian", "Creature — Lizard", o690CragSaurian, 4, 4)
	pinger := effectstest.PushVanillaCreature(g, opp.ID, "Pinger", 1, 1)
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(pinger, saurian, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.FindBattlefieldCardForTest(g, saurian)
	if c == nil || c.Controller != opp.ID {
		t.Fatalf("the damage's controller did not gain control of the Saurian: %+v", c)
	}
	// Damage from its new controller's own source changes nothing.
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(pinger, saurian, 1); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindBattlefieldCardForTest(g, saurian); c == nil || c.Controller != opp.ID {
		t.Error("control moved when its controller's own source damaged it")
	}
}

func TestS82OldB690VolcanicWindDividesDamageByTheCreaturesOnTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushVanillaCreature(g, opp.ID, "A", 0, 9)
	b := effectstest.PushVanillaCreature(g, opp.ID, "B", 0, 9)
	effectstest.PushVanillaCreature(g, me.ID, "C", 0, 9)
	effectstest.PushVanillaCreature(g, me.ID, "D", 0, 9)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCRR")
	id := effectstest.HandCardOf(g, me, "Volcanic Wind", "Sorcery", "{4}{R}{R}", o690VolcanicWind)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets, Distribution: map[uuid.UUID]int{a: 3, b: 1}}); err != nil {
		t.Fatalf("cast with X = 4 creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o690Marked(g, a) != 3 || o690Marked(g, b) != 1 {
		t.Errorf("damage %d / %d, want 3 / 1", o690Marked(g, a), o690Marked(g, b))
	}
	// A split that does not total the creatures on the battlefield is refused.
	effectstest.FloatForTest(g, me, "CCCCRR")
	again := effectstest.HandCardOf(g, me, "Volcanic Wind", "Sorcery", "{4}{R}{R}", o690VolcanicWind)
	if err := g.CastSpell(me.ID, again, game.CastSpellParams{Strict: true, Targets: targets, Distribution: map[uuid.UUID]int{a: 1, b: 1}}); err == nil {
		t.Error("a split of 2 was accepted with 4 creatures on the battlefield")
	}
}

func TestS82OldB690BloodOathCountsTheChosenTypeInTheRevealedHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	for _, c := range []struct{ n, tl string }{{"Wolf", "Creature — Wolf"}, {"Bear", "Artifact Creature — Bear"}, {"Forest", "Basic Land — Forest"}} {
		effectstest.HandCardOf(g, opp, c.n, c.tl, "", "")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCR")
	id := effectstest.HandCardOf(g, me, "Blood Oath", "Instant", "{3}{R}", o690BloodOath)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{o690Player(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 2) // Creature
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40-6 {
		t.Errorf("life = %d, want 34 (two creature cards)", opp.Life)
	}
	for _, c := range opp.Hand.Cards {
		if !c.IsKnownTo(me.ID) {
			t.Errorf("%s was not revealed", c.Name)
		}
	}
}

func TestS82OldB690BloodOathWithNoCardOfTheTypeDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	opp.Hand.Cards = nil
	effectstest.HandCardOf(g, opp, "Forest", "Basic Land — Forest", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCR")
	id := effectstest.HandCardOf(g, me, "Blood Oath", "Instant", "{3}{R}", o690BloodOath)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{o690Player(opp.ID)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 4) // Instant
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != 40 {
		t.Errorf("life = %d, want 40", opp.Life)
	}
}
