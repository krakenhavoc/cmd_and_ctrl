package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b750_cards_test.go — Stronghold, pump and combat tricks, enchantments,
// life gain and loss, other creatures, batch 750 (#5627, S82 OLD). Every test
// plays the real catalog entry through the engine.

const (
	oldB750Prowler    = "b56c5b39-2465-4a25-b806-bb84866e576f"
	oldB750Victual    = "bd803b1c-1370-40db-a6c6-abad4b3d1602"
	oldB750Awakening  = "5563427d-ffdc-4bea-ad7d-1605894b8a6c"
	oldB750DreamHalls = "d901e17f-a5f4-49d9-8a6d-ab771949af24"
	oldB750Spined     = "eed9c168-1c78-43c6-a814-9b80054d1373"
	oldB750Taskmaster = "7a2b134d-9d7e-4d43-98fc-c776fc4bad45"
	oldB750Gardens    = "ff99337e-410c-43af-9475-7995d690bcef"
	oldB750Contempl   = "fa7efcef-a688-4e25-a823-4d53b2e96508"
	oldB750Champion   = "765017e5-b43d-4c77-a9a9-0fcf0c0cd098"
	oldB750WalkDream  = "286e7ef1-d763-4b48-8603-a161dcf7efbc"
	oldB750Hesitation = "11eb31d9-cbcb-4ec5-abab-dfa646addbba"
	oldB750Rats       = "0f2504b0-bd5a-4b61-b8bb-c8b20aa1e26e"
)

func TestOldB750CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB750Prowler: "Dream Prowler", oldB750Victual: "Victual Sliver", oldB750Awakening: "Awakening",
		oldB750DreamHalls: "Dream Halls", oldB750Spined: "Spined Sliver", oldB750Taskmaster: "Stronghold Taskmaster",
		oldB750Gardens: "Volrath's Gardens", oldB750Contempl: "Contemplation", oldB750Champion: "Soltari Champion",
		oldB750WalkDream: "Walking Dream", oldB750Hesitation: "Hesitation", oldB750Rats: "Rabid Rats",
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

func TestOldB750SoltariChampionHasShadow(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Soltari Champion", "Creature — Soltari Soldier", oldB750Champion, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, id, "shadow") {
		t.Error("the Champion has no shadow")
	}
}

func oldB750PT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	return effectstest.EffectivePTOf(t, g, id)
}

func TestOldB750SoltariChampionPumpsTheOthersWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champion := effectstest.B12Push(g, me.ID, "Soltari Champion", "Creature — Soltari Soldier", oldB750Champion, 2, 2)
	friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, champion)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB750PT(t, g, friend); p != 3 || tough != 3 {
		t.Errorf("the other creature is %d/%d, want 3/3", p, tough)
	}
	if p, tough := oldB750PT(t, g, champion); p != 2 || tough != 2 {
		t.Errorf("the Champion is %d/%d, want 2/2 (it is not 'other')", p, tough)
	}
	if p, tough := oldB750PT(t, g, foe); p != 2 || tough != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB750TaskmasterShrinksOtherBlackCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := effectstest.B12Push(g, me.ID, "Stronghold Taskmaster", "Creature — Giant Minion", oldB750Taskmaster, 4, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, master).Colors = []string{"B"} })
	black := effectstest.Pr7Creature(g, opp.ID, "Black Bear", 2, "B")
	green := effectstest.Pr7Creature(g, opp.ID, "Green Bear", 2, "G")
	if p, tough := oldB750PT(t, g, black); p != 1 || tough != 3 {
		t.Errorf("a black creature is %d/%d, want 1/3 (a 2/4 shrunk)", p, tough)
	}
	if p, tough := oldB750PT(t, g, green); p != 2 || tough != 4 {
		t.Errorf("a green creature is %d/%d, want 2/4", p, tough)
	}
	if p, tough := oldB750PT(t, g, master); p != 4 || tough != 3 {
		t.Errorf("the Taskmaster is %d/%d, want 4/3 (only other creatures)", p, tough)
	}
}

func TestOldB750ContemplationGainsOneLifePerSpellYouCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Contemplation", "Enchantment", oldB750Contempl, false)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Spell One", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after one spell, want %d", me.Life, life+1)
	}
	effectstest.CastCatalogSpell(t, g, "Spell Two", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d after two spells, want %d", me.Life, life+2)
	}
}

func TestOldB750HesitationSacrificesItselfAndCountersTheSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	halt := effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Hesitation", "Enchantment", oldB750Hesitation, false)
	bolt := effectstest.CastCatalogSpell(t, g, "Spell One", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(halt) {
		t.Error("Hesitation was not sacrificed")
	}
	if !me.Graveyard.Contains(bolt) {
		t.Error("the spell was not countered into its owner's graveyard")
	}
	if len(g.Stack.Cards) != 0 {
		t.Error("the spell is still on the stack")
	}
}

func TestOldB750VolrathsGardensGainsTwoAtSorcerySpeedByTappingACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	gardens := effectstest.PushCatalogPermanent(g, me.ID, "Volrath's Gardens", "Enchantment", oldB750Gardens, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	life := me.Life
	effectstest.RavB2MustActivate(t, g, me.ID, gardens, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if !effectstest.TappedForTest(t, g, bear) {
		t.Error("the creature was not tapped as the cost")
	}
	// Tapped, it can't pay again.
	if err := effectstest.RavB2Activate(t, g, me.ID, gardens, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}, "C", "C"); err == nil {
		t.Error("a tapped creature paid the cost")
	}
	// Sorcery timing: refused at the declare attackers step.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := effectstest.RavB2Activate(t, g, me.ID, gardens, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{bear}}, "C", "C"); err == nil {
		t.Error("activated outside a main phase")
	}
}

func TestOldB750RabidRatsShrinksABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rats := effectstest.PushCatalogPermanent(g, opp.ID, "Rabid Rats", "Creature — Rat", oldB750Rats, false)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 3, 3)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	if err := effectstest.RavB2Activate(t, g, opp.ID, rats, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Error("a creature that is not blocking was accepted as the target")
	}
	effectstest.RavB2MustActivate(t, g, opp.ID, rats, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(blocker)})
	if p, tough := oldB750PT(t, g, blocker); p != 2 || tough != 2 {
		t.Errorf("the blocker is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB750DreamProwlerIsUnblockableOnlyWhileAttackingAlone(t *testing.T) {
	for _, alone := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		prowler := effectstest.PushCatalogPermanent(g, me.ID, "Dream Prowler", "Creature — Illusion", oldB750Prowler, false)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		friend := effectstest.B12Creature(g, me.ID, "Friend", "Creature — Bear", 2, 2)
		attackers := []uuid.UUID{prowler}
		if !alone {
			attackers = append(attackers, friend)
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		for _, a := range attackers {
			if err := g.DeclareAttacker(a, opp.ID); err != nil {
				t.Fatalf("DeclareAttacker: %v", err)
			}
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		err := g.DeclareBlocker(bear, prowler)
		if alone {
			effectstest.BrRefusal(t, err, game.BlockReasonCantBeBlocked)
		} else if err != nil {
			t.Errorf("attacking with another creature, the Prowler could not be blocked: %v", err)
		}
	}
}

func TestOldB750WalkingDreamCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	dream := effectstest.PushCatalogPermanent(g, me.ID, "Walking Dream", "Creature — Illusion", oldB750WalkDream, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(dream, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.BrRefusal(t, g.DeclareBlocker(bear, dream), game.BlockReasonCantBeBlocked)
}

func TestOldB750WalkingDreamStaysTappedWhileAnOpponentControlsTwoCreatures(t *testing.T) {
	for _, theirs := range []int{1, 2} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dream := effectstest.PushCatalogPermanent(g, me.ID, "Walking Dream", "Creature — Illusion", oldB750WalkDream, false)
		for i := 0; i < theirs; i++ {
			effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		}
		g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, dream).Tapped = true })
		effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		stillTapped := effectstest.TappedForTest(t, g, dream)
		if want := theirs >= 2; stillTapped != want {
			t.Errorf("opponent controls %d creatures: Walking Dream tapped = %v, want %v", theirs, stillTapped, want)
		}
	}
}

func TestOldB750VictualSliverGivesEverySliverASacrificeForFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Victual Sliver", "Creature — Sliver", oldB750Victual, false)
	theirs := effectstest.GaPut(g, opp.ID, "Their Sliver", "Creature — Sliver", "")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, theirs)
		c.Power, c.Toughness = 1, 1
	})
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, opp.ID, theirs, 0, game.ActivateAbilityParams{}, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life+4 {
		t.Errorf("their life %d, want %d", opp.Life, life+4)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("the Sliver was not sacrificed as the cost")
	}
}

func TestOldB750AwakeningUntapsEveryCreatureAndLandAtEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Awakening", "Enchantment", oldB750Awakening, false)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.GaPut(g, opp.ID, "Island", "Basic Land — Island", "")
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{mine, theirs, land, rock} {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = true
		}
	})
	// An opponent's upkeep triggers it as well as mine.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"my creature": mine, "their creature": theirs, "their land": land} {
		if effectstest.TappedForTest(t, g, id) {
			t.Errorf("%s is still tapped", name)
		}
	}
	if !effectstest.TappedForTest(t, g, rock) {
		t.Error("a tapped artifact that is neither a creature nor a land was untapped")
	}
}

func TestOldB750SpinedSliverGrowsABlockedSliverPerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Spined Sliver", "Creature — Sliver", oldB750Spined, false)
	sliver := effectstest.GaPut(g, me.ID, "Plain Sliver", "Creature — Sliver", "")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, sliver)
		c.Power, c.Toughness = 1, 1
	})
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 4)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 4)
	effectstest.DeclareAttack(t, g, opp.ID, sliver)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := effectstest.ReqBlock(t, g, sliver, b1, b2); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := oldB750PT(t, g, sliver); p != 3 || tough != 3 {
		t.Errorf("the blocked Sliver is %d/%d, want 3/3 (+1/+1 for each of two blockers)", p, tough)
	}
}

func TestOldB750DreamHallsLetsAPlayerDiscardAColorSharingCardInsteadOfPaying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID, "Dream Halls", "Enchantment", oldB750DreamHalls, false)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	mk := func(name, cost, colour string) uuid.UUID {
		id := effectstest.HandCardOf(g, me, name, "Instant", cost, "")
		g.WithWriteLock(func() {
			for i := range me.Hand.Cards {
				if me.Hand.Cards[i].InstanceID == id {
					me.Hand.Cards[i].Colors = []string{colour}
				}
			}
		})
		return id
	}
	spell := mk("Blue Spell", "{U}{U}", "U")
	red := mk("Red Card", "{R}", "R")
	blue := mk("Blue Card", "{U}", "U")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostDreamHalls, AltCostIDs: []uuid.UUID{red}}); err == nil {
		t.Fatal("a red card was accepted for a blue spell")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AlternativeCost: effects.GrantedAltCostDreamHalls, AltCostIDs: []uuid.UUID{blue}}); err != nil {
		t.Fatalf("cast with Dream Halls: %v", err)
	}
	if me.Hand.Contains(blue) || !me.Graveyard.Contains(blue) {
		t.Error("the blue card was not discarded as the cost")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool %v, want nothing spent", me.ManaPool)
	}
}
