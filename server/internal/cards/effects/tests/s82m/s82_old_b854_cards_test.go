package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b854_cards_test.go — Fallen Empires, keywords, removal and combat
// tricks, batch 854 (#5637, S82 OLD). Every test plays the real catalog entry
// through the engine.

const (
	oldB854Flotilla    = "014ca38f-d4ac-4417-bfa8-785a7d6fce43"
	oldB854Skirmishers = "8ddbe82b-7fc9-4b2c-85de-dc747adbb3ca"
	oldB854Phalanx     = "a1160b8d-c4e4-4869-b7a5-c6a826eb1ba4"
	oldB854Orgg        = "f6ea2b1b-1403-4ccd-a23c-48e0e81ceb91"
	oldB854Hand        = "af8eb20a-dfa2-49a4-892b-26d7e276fcd2"
	oldB854Necrite     = "e6ab8a04-e225-4ff0-bb8b-c80d0e6ce50c"
	oldB854WarMachine  = "f72511b0-9993-4998-b646-65a855cb8170"
	oldB854Fortress    = "29f34e18-2102-4ded-912f-4ac3e9e6744c"
	oldB854Champion    = "728f0a3f-722d-4936-ade8-31883499f56d"
	oldB854Captain     = "76c06de7-39ff-4b80-8e54-20910372cffd"
	oldB854Soldier     = "88f9075c-ea96-4f8d-b0d1-e5e244e824c1"
)

func TestOldB854CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB854Flotilla: "Goblin Flotilla", oldB854Skirmishers: "Icatian Skirmishers", oldB854Phalanx: "Icatian Phalanx",
		oldB854Orgg: "Orgg", oldB854Hand: "Hand of Justice", oldB854Necrite: "Necrite", oldB854WarMachine: "Vodalian War Machine",
		oldB854Fortress: "Elven Fortress", oldB854Champion: "Thrull Champion", oldB854Captain: "Orcish Captain",
		oldB854Soldier: "Dwarven Soldier",
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

func TestOldB854PrintedKeywords(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		keywords               []string
	}{
		{"Goblin Flotilla", "Creature — Goblin", oldB854Flotilla, []string{"islandwalk"}},
		{"Icatian Skirmishers", "Creature — Human Soldier", oldB854Skirmishers, []string{"first strike", "banding"}},
		{"Icatian Phalanx", "Creature — Human Soldier", oldB854Phalanx, []string{"banding"}},
		{"Orgg", "Creature — Orgg", oldB854Orgg, []string{"trample"}},
		{"Vodalian War Machine", "Creature — Wall", oldB854WarMachine, []string{"defender"}},
	} {
		g := effectstest.NewCatalogGame(t)
		id := effectstest.B12Push(g, g.Seats[0].ID, tc.name, tc.typeLine, tc.oracle, 2, 2)
		for _, kw := range tc.keywords {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestOldB854FlotillaGivesFirstStrikeUnlessPaid(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		flotilla := effectstest.B12Push(g, me.ID, "Goblin Flotilla", "Creature — Goblin", oldB854Flotilla, 2, 2)
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 10)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.B06AddMana(me, "R")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(flotilla, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(bear, flotilla); err != nil {
			// The Flotilla has islandwalk but the defender controls no Island.
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.HasKeywordOnBattlefield(t, g, bear, "first strike")
		if got == pay {
			t.Errorf("paid %v: the blocker has first strike = %v", pay, got)
		}
	}
}

func TestOldB854SkirmishersGiveTheirBandFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	skirm := effectstest.B12Push(g, me.ID, "Icatian Skirmishers", "Creature — Human Soldier", oldB854Skirmishers, 1, 1)
	phalanx := effectstest.B12Push(g, me.ID, "Icatian Phalanx", "Creature — Human Soldier", oldB854Phalanx, 2, 4)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.HasKeywordOnBattlefield(t, g, phalanx, "first strike") {
		t.Fatal("the Phalanx has first strike before attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	decls := []game.AttackDeclaration{
		{Attacker: skirm, Target: opp.ID, Band: 1},
		{Attacker: phalanx, Target: opp.ID, Band: 1},
		{Attacker: bear, Target: opp.ID, Band: 1},
	}
	if _, err := g.DeclareAttackersWith(decls, game.DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare the band: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{phalanx, bear} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "first strike") {
			t.Errorf("%s did not gain first strike from the band", effectstest.LayeredCard(t, g, id).Name)
		}
	}
}

func TestOldB854OrggCantAttackIntoABigUntappedCreatureNorBlockBigOnes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orgg := effectstest.B12Push(g, me.ID, "Orgg", "Creature — Orgg", oldB854Orgg, 6, 6)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Giant", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(orgg, opp.ID); err == nil {
		t.Fatal("the Orgg attacked a player controlling an untapped 3-power creature")
	}
	effectstest.TapForTest(t, g, big)
	if err := g.DeclareAttacker(orgg, opp.ID); err != nil {
		t.Fatalf("the Orgg could not attack once the big creature was tapped: %v", err)
	}
}

func TestOldB854OrggCanAttackPastSmallCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	orgg := effectstest.B12Push(g, me.ID, "Orgg", "Creature — Orgg", oldB854Orgg, 6, 6)
	effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(orgg, opp.ID); err != nil {
		t.Fatalf("the Orgg could not attack past a 2-power creature: %v", err)
	}
}

func TestOldB854OrggBlocksSmallAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 3)
	orgg := effectstest.B12Push(g, opp.ID, "Orgg", "Creature — Orgg", oldB854Orgg, 6, 6)
	effectstest.DeclareAttack(t, g, opp.ID, small, big)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(orgg, big); err == nil {
		t.Fatal("the Orgg blocked a creature with power 3")
	}
	if err := g.DeclareBlocker(orgg, small); err != nil {
		t.Fatalf("the Orgg could not block a 2-power creature: %v", err)
	}
}

func TestOldB854HandOfJusticeTapsThreeWhiteCreaturesToDestroy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	hand := effectstest.B12Push(g, me.ID, "Hand of Justice", "Creature — Avatar", oldB854Hand, 2, 6)
	white := func(name string) uuid.UUID {
		id := effectstest.B12Creature(g, me.ID, name, "Creature — Soldier", 1, 1)
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Colors = []string{"W"} })
		return id
	}
	w1, w2, w3 := white("W1"), white("W2"), white("W3")
	green := effectstest.B12Creature(g, me.ID, "Green", "Creature — Elf", 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), TapIDs: []uuid.UUID{w1, w2}}); err == nil {
		t.Fatal("activated with two creatures tapped")
	}
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), TapIDs: []uuid.UUID{w1, w2, green}}); err == nil {
		t.Fatal("a non-white creature paid")
	}
	if err := g.ActivateCatalogAbility(me.ID, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), TapIDs: []uuid.UUID{w1, w2, hand}}); err == nil {
		t.Fatal("the Hand tapped itself as one of the three")
	}
	effectstest.P7Activate(t, g, me, hand, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim), TapIDs: []uuid.UUID{w1, w2, w3}})
	if g.Battlefield.Contains(victim) {
		t.Error("the target survived")
	}
	for _, id := range []uuid.UUID{hand, w1, w2, w3} {
		if !effectstest.IsTapped(g, id) {
			t.Errorf("%s is untapped", effectstest.LayeredCard(t, g, id).Name)
		}
	}
}

func TestOldB854NecriteSacrificesToDestroyADefendersCreature(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		necrite := effectstest.B12Push(g, me.ID, "Necrite", "Creature — Thrull", oldB854Necrite, 2, 2)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		effectstest.DeclareAttack(t, g, opp.ID, necrite)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		for i := 0; i < 4; i++ {
			p := effectstest.LatestPickTarget(g, me.ID)
			if p == nil {
				break
			}
			for _, id := range []uuid.UUID{mine, victim} {
				if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
					break
				}
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.Battlefield.Contains(necrite); got == yes {
			t.Errorf("answered %v: the Necrite on the battlefield = %v", yes, got)
		}
		if got := g.Battlefield.Contains(victim); got == yes {
			t.Errorf("answered %v: their creature on the battlefield = %v", yes, got)
		}
		if !g.Battlefield.Contains(mine) {
			t.Errorf("answered %v: the attacker's own creature was destroyed", yes)
		}
	}
}

func TestOldB854WarMachineAttacksAndPumpsByTappingMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	machine := effectstest.B12Push(g, me.ID, "Vodalian War Machine", "Creature — Wall", oldB854WarMachine, 0, 4)
	fish := effectstest.B12Creature(g, me.ID, "Fish", "Creature — Merfolk", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, machine, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("a non-Merfolk paid")
	}
	effectstest.P7Activate(t, g, me, machine, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{fish}})
	if p, tough := effectstest.EffectivePTOf(t, g, machine); p != 2 || tough != 5 {
		t.Errorf("the War Machine is %d/%d, want 2/5", p, tough)
	}
	fish2 := effectstest.B12Creature(g, me.ID, "Fish Two", "Creature — Merfolk", 1, 1)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(machine, opp.ID); err == nil {
		t.Fatal("the War Machine attacked with defender")
	}
	effectstest.P7Activate(t, g, me, machine, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{fish2}})
	if err := g.DeclareAttacker(machine, opp.ID); err != nil {
		t.Fatalf("the War Machine could not attack after paying: %v", err)
	}
}

func TestOldB854WarMachineDiesAndDestroysTheMerfolkThatPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	machine := effectstest.B12Push(g, me.ID, "Vodalian War Machine", "Creature — Wall", oldB854WarMachine, 0, 4)
	payer := effectstest.B12Creature(g, me.ID, "Payer", "Creature — Merfolk", 1, 1)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Merfolk", 1, 1)
	effectstest.P7Activate(t, g, me, machine, 1, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{payer}})
	effectstest.DestroyIt(t, g, machine)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(machine) {
		t.Fatal("the War Machine survived")
	}
	if g.Battlefield.Contains(payer) {
		t.Error("the Merfolk that paid survived")
	}
	if !g.Battlefield.Contains(bystander) {
		t.Error("a Merfolk that never paid was destroyed")
	}
}

func TestOldB854ElvenFortressPumpsABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fortress := effectstest.PushCatalogPermanent(g, me.ID, "Elven Fortress", "Enchantment", oldB854Fortress, false)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, atk); err != nil {
		t.Fatal(err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.B06AddMana(me, "G", "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, fortress, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(idle)}); err == nil {
		t.Fatal("a non-blocking creature was targeted")
	}
	effectstest.P7Activate(t, g, me, fortress, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(blocker)})
	if p, tough := effectstest.EffectivePTOf(t, g, blocker); p != 2 || tough != 3 {
		t.Errorf("the blocker is %d/%d, want 2/3", p, tough)
	}
}

func TestOldB854ThrullChampionPumpsThrullsAndStealsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	champ := effectstest.B12Push(g, me.ID, "Thrull Champion", "Creature — Thrull", oldB854Champion, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Thrull", "Creature — Thrull", 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 1, 1)
	if p, tough := effectstest.EffectivePTOf(t, g, champ); p != 3 || tough != 3 {
		t.Errorf("the Champion is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Thrull is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 1 {
		t.Errorf("their Bear is %d/%d, want 1/1", p, tough)
	}
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Fatal("a non-Thrull was targeted")
	}
	effectstest.P7Activate(t, g, me, champ, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(theirs)})
	if got := effectstest.LayeredCard(t, g, theirs).Controller; got != me.ID {
		t.Errorf("their Thrull is controlled by %s, want its thief", got)
	}
	effectstest.DestroyIt(t, g, champ)
	if got := effectstest.LayeredCard(t, g, theirs).Controller; got != opp.ID {
		t.Errorf("the Thrull stayed stolen after the Champion died")
	}
}

func TestOldB854OrcishCaptainFlipsToPumpOrShrinkAnOrc(t *testing.T) {
	sawWin, sawLose := false, false
	for i := 0; i < 40 && !(sawWin && sawLose); i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		captain := effectstest.B12Push(g, me.ID, "Orcish Captain", "Creature — Orc Warrior", oldB854Captain, 1, 4)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		if err := g.ActivateCatalogAbility(me.ID, captain, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
			t.Fatal("a non-Orc was targeted")
		}
		effectstest.B06AddMana(me, "C")
		if err := g.ActivateCatalogAbility(me.ID, captain, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(captain)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if ask == nil {
			t.Fatal("no coin call was asked")
		}
		if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
			t.Fatalf("ResolveCoinCall: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.EffectivePTOf(t, g, captain)
		switch {
		case p == 3 && tough == 4:
			sawWin = true
		case p == 1 && tough == 2:
			sawLose = true
		default:
			t.Fatalf("the Captain is %d/%d, want 3/4 (won) or 1/2 (lost)", p, tough)
		}
	}
	if !sawWin || !sawLose {
		t.Errorf("forty flips never showed both results (won %v, lost %v)", sawWin, sawLose)
	}
}

func TestOldB854DwarvenSoldierGrowsAgainstOrcsOnly(t *testing.T) {
	for _, tc := range []struct {
		subtype string
		want    int
	}{{"Orc", 12}, {"Bear", 10}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		soldier := effectstest.B12Push(g, me.ID, "Dwarven Soldier", "Creature — Dwarf Soldier", oldB854Soldier, 2, 10)
		foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — "+tc.subtype, 1, 1)
		effectstest.DeclareAttack(t, g, opp.ID, soldier)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(foe, soldier); err != nil {
			t.Fatal(err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if _, tough := effectstest.EffectivePTOf(t, g, soldier); tough != tc.want {
			t.Errorf("blocked by a %s: toughness %d, want %d", tc.subtype, tough, tc.want)
		}
	}
}
