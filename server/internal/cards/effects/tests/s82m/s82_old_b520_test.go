package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b520_test.go — Eventide, damage: S82 OLD batch 520 (#5590), each card
// driven through the engine.

const (
	b520Noggle    = "16069935-4aeb-481f-98f2-46387775540f"
	b520Crackle   = "59a6cb4e-1646-42b5-b482-da641c0f7f6f"
	b520Woodlurk  = "5f4ddd52-4940-4557-8531-40c66857d4cd"
	b520Hateflyr  = "681d07e6-234f-4544-a2ea-bfbefe6b0b36"
	b520Assailant = "85b52f70-b583-4f04-a2ab-226d7291c03a"
	b520Backlash  = "be6cb3e6-7492-4e30-83a9-9bc413e50cdc"
	b520FireAtWl  = "ca8409ab-d438-441a-861f-c41e307d5621"
	b520Ashling   = "e89e102f-e3bc-4c16-bb27-09db83cec5ec"
	b520Needle    = "ea8acbd4-1ace-4634-9e88-7f91401d2b30"
)

func TestS82OldB520EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b520Noggle: "Noggle Hedge-Mage", b520Crackle: "Crackleburr", b520Woodlurk: "Woodlurker Mimic",
		b520Hateflyr: "Hateflayer", b520Assailant: "Duergar Assailant", b520Backlash: "Chaotic Backlash",
		b520FireAtWl: "Fire at Will", b520Ashling: "Ashling, the Extinguisher", b520Needle: "Needle Specter",
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

func b520Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestS82OldB520NoggleHedgeMageTapsTwoPermanentsWithTwoIslands(t *testing.T) {
	for _, tc := range []struct {
		name    string
		islands int
		yes     bool
		want    bool
	}{{"two Islands, yes", 2, true, true}, {"two Islands, no", 2, false, false}, {"one Island", 1, true, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := b520Seats(g)
			for i := 0; i < tc.islands; i++ {
				effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
			}
			a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 2)
			b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 2)
			effectstest.CastCatalogSpell(t, g, "Noggle Hedge-Mage", "Creature — Noggle Wizard", b520Noggle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.LatestPickTarget(g, me.ID) != nil {
				effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(a, b)...)
			} else if tc.islands >= 2 {
				t.Fatal("no target prompt with two Islands")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if effectstest.B18TriggerPromptCount(g, me.ID) > 0 {
				effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, tc.yes)
				effectstest.PassPriorityAroundTable(t, g)
			}
			if got := effectstest.B16Tapped(t, g, a) && effectstest.B16Tapped(t, g, b); got != tc.want {
				t.Errorf("both targets tapped = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestS82OldB520NoggleHedgeMageHitsAPlayerWithTwoMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b520Seats(g)
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	effectstest.B12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Noggle Hedge-Mage", "Creature — Noggle Wizard", b520Noggle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.PlayerTarget(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want −2", life, opp.Life)
	}
}

func TestS82OldB520CrackleburrDamagesByTappingTwoRedCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	burr := effectstest.B12Push(g, me.ID, "Crackleburr", "Creature — Elemental", b520Crackle, 2, 2)
	effectstest.Unsick(g, burr)
	r1 := effectstest.B31Push(g, me.ID, "Red One", "Creature — Goblin", "", "{R}", 1, 1, "R")
	r2 := effectstest.B31Push(g, me.ID, "Red Two", "Creature — Goblin", "", "{R}", 1, 1, "R")
	green := effectstest.B31Push(g, me.ID, "Green", "Creature — Elf", "", "{G}", 1, 1, "G")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, burr, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, TapIDs: []uuid.UUID{r1, green},
	}); err == nil {
		t.Fatal("a green creature paid for tapping two red creatures")
	}
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, burr, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, TapIDs: []uuid.UUID{r1, r2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want −3", life, opp.Life)
	}
	for _, id := range []uuid.UUID{burr, r1, r2} {
		if !effectstest.B16Tapped(t, g, id) {
			t.Error("the Crackleburr or one of the red creatures is untapped")
		}
	}
}

func TestS82OldB520CrackleburrBouncesByUntappingTwoBlueCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	burr := effectstest.B12Push(g, me.ID, "Crackleburr", "Creature — Elemental", b520Crackle, 2, 2)
	effectstest.Unsick(g, burr)
	effectstest.ArnTap(t, g, burr)
	b1 := effectstest.B31Push(g, me.ID, "Blue One", "Creature — Merfolk", "", "{U}", 1, 1, "U")
	b2 := effectstest.B31Push(g, me.ID, "Blue Two", "Creature — Merfolk", "", "{U}", 1, 1, "U")
	effectstest.ArnTap(t, g, b1)
	effectstest.ArnTap(t, g, b2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "UU")
	if err := g.ActivateCatalogAbility(me.ID, burr, 1, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(theirs), TapIDs: []uuid.UUID{b1, b2},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the target is still on the battlefield")
	}
	if !opp.Hand.Contains(theirs) {
		t.Errorf("the target did not return to its owner's hand (it is in %v)", effectstest.ZoneOfCard(g, theirs))
	}
	for _, id := range []uuid.UUID{burr, b1, b2} {
		if effectstest.B16Tapped(t, g, id) {
			t.Error("the Crackleburr or a blue creature is still tapped")
		}
	}
}

func TestS82OldB520WoodlurkerMimicGrowsOnABlackAndGreenSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b520Seats(g)
	effectstest.AdvanceToMain(t, g)
	mimic := effectstest.B12Push(g, me.ID, "Woodlurker Mimic", "Creature — Shapeshifter", b520Woodlurk, 2, 1)
	effectstest.FloatForTest(g, me, "B")
	effectstest.CastFromHandForTest(t, g, me, "Mono", "Instant", "{B}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mimic); p != 2 || tough != 1 {
		t.Fatalf("a mono-black spell changed it to %d/%d", p, tough)
	}
	effectstest.FloatForTest(g, me, "BG")
	effectstest.CastFromHandForTest(t, g, me, "Both", "Instant", "{B}{G}", "", effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mimic); p != 4 || tough != 5 {
		t.Errorf("after a black-and-green spell: %d/%d, want 4/5", p, tough)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, mimic, "wither") {
		t.Error("no wither")
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.PtOf(t, g, mimic); p != 2 || tough != 1 || effectstest.HasKeywordOnBattlefield(t, g, mimic, "wither") {
		t.Errorf("after cleanup: %d/%d, want 2/1 with no wither", p, tough)
	}
}

func TestS82OldB520HateflayerDealsItsPowerAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fl := effectstest.B12Push(g, me.ID, "Hateflayer", "Creature — Elemental", b520Hateflyr, 5, 5)
	effectstest.Unsick(g, fl)
	if !effectstest.HasKeywordOnBattlefield(t, g, fl, "wither") {
		t.Error("no wither")
	}
	if err := g.ActivateCatalogAbility(me.ID, fl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err == nil {
		t.Fatal("{Q} was paid by an untapped Hateflayer")
	}
	effectstest.ArnTap(t, g, fl)
	effectstest.FloatForTest(g, me, "RRR")
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, fl, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("life %d → %d, want −5", life, opp.Life)
	}
	if effectstest.B16Tapped(t, g, fl) {
		t.Error("the Hateflayer is still tapped after {Q}")
	}
}

func TestS82OldB520DuergarAssailantSacrificesToPingAnAttackerOrBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	asl := effectstest.B12Push(g, me.ID, "Duergar Assailant", "Creature — Dwarf Soldier", b520Assailant, 1, 1)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.MscM1Attacking(g, attacker, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, asl, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if !g.Battlefield.Contains(asl) {
		t.Fatal("a refused activation paid its sacrifice")
	}
	if err := g.ActivateCatalogAbility(me.ID, asl, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(asl) {
		t.Error("the Assailant was not sacrificed")
	}
	if d := effectstest.E2Card(t, g, attacker).DamageMarked; d != 1 {
		t.Errorf("the attacker has %d damage, want 1", d)
	}
}

func TestS82OldB520ChaoticBacklashCountsWhiteAndBluePermanentsOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B31Push(g, opp.ID, "White", "Creature — Soldier", "", "{W}", 1, 1, "W")
	effectstest.B31Push(g, opp.ID, "Blue", "Creature — Merfolk", "", "{U}", 1, 1, "U")
	effectstest.B31Push(g, opp.ID, "Both", "Creature — Sphinx", "", "{W}{U}", 1, 1, "W", "U")
	effectstest.B31Push(g, opp.ID, "Red", "Creature — Goblin", "", "{R}", 1, 1, "R")
	effectstest.B31Push(g, me.ID, "My White", "Creature — Soldier", "", "{W}", 1, 1, "W")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Chaotic Backlash", "Instant", b520Backlash, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("life %d → %d, want −6 (three white and/or blue permanents, doubled)", life, opp.Life)
	}
}

func TestS82OldB520FireAtWillDividesThreeAmongAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Wall", 0, 30)
	effectstest.MscM1Attacking(g, a, me.ID)
	effectstest.MscM1Attacking(g, b, me.ID)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fire at Will", TypeLine: "Instant", OracleID: b520FireAtWl, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(idle), Distribution: map[uuid.UUID]int{idle: 3}}); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1}}); err == nil {
		t.Fatal("a division of 2 was accepted")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if d1, d2 := effectstest.E2Card(t, g, a).DamageMarked, effectstest.E2Card(t, g, b).DamageMarked; d1 != 1 || d2 != 2 {
		t.Errorf("damage %d/%d, want 1/2", d1, d2)
	}
}

func TestS82OldB520AshlingMakesTheDamagedPlayerSacrificeTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b520Seats(g)
	ash := effectstest.B12Push(g, me.ID, "Ashling, the Extinguisher", "Legendary Creature — Elemental Shaman", b520Ashling, 4, 4)
	keep := effectstest.B12Creature(g, opp.ID, "Keep", "Creature — Bear", 2, 2)
	doomed := effectstest.B12Creature(g, opp.ID, "Doomed", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.DealCombatDamage(g, me.ID, ash, opp.ID)
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(doomed)...)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(doomed) {
		t.Error("the chosen creature was not sacrificed")
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(mine) {
		t.Error("another creature was sacrificed")
	}
	if effectstest.ZoneOfCard(g, doomed) != game.ZoneGraveyard {
		t.Errorf("the sacrificed creature is in %v", effectstest.ZoneOfCard(g, doomed))
	}
}

func TestS82OldB520NeedleSpecterMakesThePlayerDiscardThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b520Seats(g)
	spec := effectstest.B12Push(g, me.ID, "Needle Specter", "Creature — Specter", b520Needle, 1, 1)
	for _, kw := range []string{"flying", "wither"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, spec, kw) {
			t.Errorf("no %s", kw)
		}
	}
	effectstest.D2dSetHand(opp, 5)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDealDamage, Actor: me.ID, Source: spec, Target: opp.ID, Amount: 3, Combat: true})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DiscardOwed(g, opp.ID); n != 3 {
		t.Fatalf("the damaged player owes %d discards, want 3", n)
	}
	// Noncombat damage does not trigger it.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := b520Seats(g2)
	spec2 := effectstest.B12Push(g2, me2.ID, "Needle Specter", "Creature — Specter", b520Needle, 1, 1)
	effectstest.D2dSetHand(opp2, 5)
	effectstest.DealNoncombatDamage(g2, me2.ID, spec2, opp2.ID)
	effectstest.PassPriorityAroundTable(t, g2)
	if n := effectstest.DiscardOwed(g2, opp2.ID); n != 0 {
		t.Errorf("noncombat damage asked for %d discards", n)
	}
}
