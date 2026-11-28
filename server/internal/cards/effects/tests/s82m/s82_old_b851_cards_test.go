package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b851_cards_test.go — Ice Age cards, S82 OLD batch 851 (#5636), played
// through the real catalog entries.

const (
	o851Familiar  = "90767f52-fa6d-4ef3-ae54-d8e3881971c1"
	o851Spoils    = "a40c52ee-4b45-4b14-8fbc-6529b42bfc9d"
	o851Tarpan    = "b1e1ebc8-cd09-4280-a169-58ab7ece43ff"
	o851Chaos     = "b6756608-a618-4e26-a090-ea8fc992def1"
	o851Leech     = "b74d3768-74bb-413c-b7c7-41028ff48225"
	o851Machinist = "06600efe-1718-4cc1-b1c0-59c82e2efd73"
	o851Unicorn   = "287cd5e8-51e2-42a4-a139-c112ff89706d"
	o851Songs     = "39419d29-47e7-4d71-953b-96a2e14ae964"
	o851Lord      = "16d0aea2-4ca4-4a9d-8674-2863e085f514"
	o851Mutant    = "540e1371-baa9-412f-a17a-22ea833ea3e8"
	o851Ghoul     = "46fbcbdf-73d0-4fa2-8db7-3b19bfc7ac22"
	o851Hymn      = "f06fca68-593b-4fbd-a5a9-618060311172"
	o851Reserves  = "a2dabd52-0b75-42d7-a07d-0df7b38646e2"
)

func TestS82OldB851CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o851Familiar: "Foul Familiar", o851Spoils: "Spoils of Evil", o851Tarpan: "Tarpan", o851Chaos: "Game of Chaos",
		o851Leech: "Thoughtleech", o851Machinist: "Soldevi Machinist", o851Unicorn: "Adarkar Unicorn",
		o851Songs: "Songs of the Damned", o851Lord: "Chaos Lord", o851Mutant: "Goblin Mutant",
		o851Ghoul: "Ashen Ghoul", o851Hymn: "Hymn of Rebirth",
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

func TestS82OldB851FoulFamiliarCantBlockAndReturnsToHandForManaAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.B12Push(g, me.ID, "Foul Familiar", "Creature — Spirit", o851Familiar, 3, 1)
	if effectstest.AuraRestrictions(t, g, fam)&game.CantBlock == 0 {
		t.Error("Foul Familiar can block")
	}
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("returned the Familiar without paying {B}")
	}
	effectstest.B06AddMana(me, "B")
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, fam, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d (1 paid)", me.Life, life-1)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fam) || !me.Hand.Contains(fam) {
		t.Error("the Familiar did not return to its owner's hand")
	}
}

func TestS82OldB851TarpanGainsALifeWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tarpan := effectstest.B12Push(g, me.ID, "Tarpan", "Creature — Horse", o851Tarpan, 1, 1)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Human", 1, 1)
	life := me.Life
	effectstest.DealDamageTo(t, g, src, tarpan, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tarpan) {
		t.Fatal("Tarpan survived lethal damage")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestS82OldB851ThoughtleechGainsLifeForAnOpponentsIslandBecomingTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Thoughtleech", "Enchantment", o851Leech, 0, 0)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Island", "Basic Land — Island", "", false)
	forest := effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Basic Land — Forest", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	life := me.Life
	for _, id := range []uuid.UUID{forest, mine} {
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
		effectstest.PassPriorityAroundTable(t, g)
	}
	if me.Life != life {
		t.Fatalf("life %d after a Forest and your own Island were tapped, want %d", me.Life, life)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d after an opponent's Island was tapped, want %d", me.Life, life+1)
	}
}

func TestS82OldB851SoldeviMachinistMakesManaForArtifactAbilitiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mach := effectstest.PushCatalogPermanent(g, me.ID, "Soldevi Machinist", "Creature — Human Wizard Artificer", o851Machinist, false)
	icy := effectstest.PushCatalogPermanent(g, me.ID, "Icy Manipulator", "Artifact", icyOracle851(t), false)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	if err := g.ActivateManaAbility(me.ID, mach, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 2 {
		t.Fatalf("the pool holds %d mana, want 2", got)
	}
	bear := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{1}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the restricted mana paid for a spell")
	}
	if err := g.ActivateCatalogAbility(me.ID, icy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(foe)}); err != nil {
		t.Fatalf("the restricted mana could not pay for an artifact's ability: %v", err)
	}
}

func icyOracle851(t *testing.T) string {
	t.Helper()
	for _, s := range effects.All() {
		if s.Name == "Icy Manipulator" {
			return s.OracleID
		}
	}
	t.Fatal("Icy Manipulator is not in the catalog")
	return ""
}

func TestS82OldB851AdarkarUnicornPaysCumulativeUpkeepWithItsRestrictedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	res := effectstest.B12Push(g, me.ID, "Volunteer Reserves", "Creature — Human Soldier", o851Reserves, 2, 4)
	uni := effectstest.B12Push(g, me.ID, "Adarkar Unicorn", "Creature — Unicorn", o851Unicorn, 2, 2)
	// The mana is for cumulative upkeep only: a spell can't use it.
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, uni, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {U}: %v", err)
	}
	spell := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: spell, Name: "Test Spell", TypeLine: "Sorcery", ManaCost: "{U}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the cumulative-upkeep mana paid for a spell")
	}
	g.WithWriteLock(func() {
		me.ManaPool.EmptyPool()
		effectstest.BattlefieldCardFor(g, uni).Tapped = false
	})
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, uni, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {U} in upkeep: %v", err)
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay the first upkeep with {U}: %v", err)
	}
	if !g.Battlefield.Contains(res) {
		t.Fatal("a paid upkeep sacrificed the Reserves")
	}
	// Two age counters: {1}{1}, paid with the second ability's {C}{U}.
	effectstest.LeaveUpkeepThenAdvanceTo(t, g, 1)
	g.WithWriteLock(func() { effectstest.BattlefieldCardFor(g, uni).Tapped = false })
	effectstest.PassPriorityAroundTable(t, g)
	prompt = effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if err := g.ActivateManaAbility(me.ID, uni, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for {C}{U} in upkeep: %v", err)
	}
	if err := g.ResolvePayUnless(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("pay the second upkeep with {C}{U}: %v", err)
	}
	if !g.Battlefield.Contains(res) {
		t.Error("a paid second upkeep sacrificed the Reserves")
	}
}

func TestS82OldB851SongsOfTheDamnedAddsBForEachCreatureCardInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogGraveyardCard(me, "Dead A", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(me, "Dead B", "Artifact Creature — Golem", "", 2, 2)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead C", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.CastCatalogSpell(t, g, "Songs of the Damned", "Instant", o851Songs, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if cols := effectstest.Batch01PoolColors(me); len(cols) != 2 || cols[0] != "B" || cols[1] != "B" {
		t.Errorf("pool %v, want {B}{B}", cols)
	}
}

func TestS82OldB851SpoilsOfEvilAddsManaAndLifeForEachArtifactOrCreatureCardInTheirGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogGraveyardCard(opp, "Dead Bear", "Creature — Bear", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(opp, "Dead Golem", "Artifact Creature — Golem", "", 2, 2)
	effectstest.PushCatalogGraveyardCard(opp, "Dead Rock", "Artifact", "", 0, 0)
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	effectstest.PushCatalogGraveyardCard(me, "My Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Spoils of Evil", "Instant", o851Spoils, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if cols := effectstest.Batch01PoolColors(me); len(cols) != 3 {
		t.Errorf("pool %v, want three {C}", cols)
	}
	if me.Life != life+3 || opp.Life != oppLife {
		t.Errorf("life %d (want %d), opponent %d (want %d)", me.Life, life+3, opp.Life, oppLife)
	}
}

func TestS82OldB851HymnOfRebirthReanimatesFromAnyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dead := effectstest.PushCatalogGraveyardCard(opp, "Dead Bear", "Creature — Bear", "", 2, 2)
	spell := effectstest.PushCatalogGraveyardCard(opp, "Dead Spell", "Sorcery", "", 0, 0)
	_ = spell
	effectstest.CastCatalogSpell(t, g, "Hymn of Rebirth", "Sorcery", o851Hymn, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not enter the battlefield")
	}
	if c.Controller != me.ID {
		t.Errorf("controller %v, want the caster", c.Controller)
	}
}

func TestS82OldB851GoblinMutantCantAttackIntoAnUntappedBigCreature(t *testing.T) {
	for _, tc := range []struct {
		name   string
		power  int
		tapped bool
		can    bool
	}{
		{"untapped power 3", 3, false, false},
		{"untapped power 5", 5, false, false},
		{"untapped power 2", 2, false, true},
		{"tapped power 3", 3, true, true},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		mutant := effectstest.HmlCreature(g, me.ID, "Goblin Mutant", o851Mutant, 5, 3)
		guard := effectstest.PushVanillaCreature(g, opp.ID, "Guard", tc.power, 3)
		if tc.tapped {
			g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, guard).Tapped = true })
		}
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		err := g.DeclareAttacker(mutant, opp.ID)
		if tc.can && err != nil {
			t.Errorf("%s: the Mutant could not attack: %v", tc.name, err)
		}
		if !tc.can && err == nil {
			t.Errorf("%s: the Mutant attacked", tc.name)
		}
	}
}

func TestS82OldB851GoblinMutantCantBlockCreaturesWithPowerThreeOrGreater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mutant := effectstest.B12Push(g, opp.ID, "Goblin Mutant", "Creature — Goblin Mutant", o851Mutant, 5, 3)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
	effectstest.BrAttack(t, g, small, big)
	if err := g.DeclareBlocker(mutant, big); err == nil {
		t.Error("the Mutant blocked a creature with power 3")
	}
	if err := g.DeclareBlocker(mutant, small); err != nil {
		t.Errorf("the Mutant could not block a creature with power 2: %v", err)
	}
}

func TestS82OldB851AshenGhoulReturnsOnlyInYourUpkeepUnderThreeCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.PushCatalogGraveyardCard(me, "Ashen Ghoul", "Creature — Zombie", o851Ghoul, 3, 1)
	effectstest.PushCatalogGraveyardCard(me, "Dead A", "Creature — Bear", "", 2, 2)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.PushCatalogGraveyardCard(me, "Dead B", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("returned with only two creature cards above it (a sorcery does not count)")
	}
	effectstest.PushCatalogGraveyardCard(me, "Dead C", "Creature — Bear", "", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with three creature cards above it: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ghoul) || me.Graveyard.Contains(ghoul) {
		t.Error("the Ghoul did not return to the battlefield")
	}
}

func TestS82OldB851AshenGhoulWillNotReturnOutsideYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ghoul := effectstest.PushCatalogGraveyardCard(me, "Ashen Ghoul", "Creature — Zombie", o851Ghoul, 3, 1)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogGraveyardCard(me, "Dead", "Creature — Bear", "", 2, 2)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the Ghoul returned outside your upkeep")
	}
}

func TestS82OldB851ChaosLordGivesItselfAwayInAnUpkeepWithAnEvenNumberOfPermanents(t *testing.T) {
	for _, tc := range []struct {
		name   string
		extras int
		gone   bool
	}{{"odd count", 0, false}, {"even count", 1, true}, {"odd again", 2, false}, {"even again", 3, true}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lord := effectstest.B12Push(g, me.ID, "Chaos Lord", "Creature — Human", o851Lord, 7, 7)
		for i := 0; i < tc.extras; i++ {
			effectstest.PushLand(g, opp.ID, "Island")
		}
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.BattlefieldCardFor(g, lord)
		if tc.gone && c.Controller != opp.ID {
			t.Errorf("%s: the Lord stayed with its controller", tc.name)
		}
		if !tc.gone && c.Controller != me.ID {
			t.Errorf("%s: the Lord changed hands", tc.name)
		}
	}
}

func TestS82OldB851ChaosLordAttacksAsThoughItHadHasteUnlessItEnteredThisTurn(t *testing.T) {
	// A Lord that came under your control this turn (not entered) may attack.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	lord := effectstest.PushCatalogPermanent(g, me.ID, "Chaos Lord", "Creature — Human", o851Lord, true)
	g.WithWriteLock(func() {
		c := effectstest.BattlefieldCardFor(g, lord)
		c.Power, c.Toughness = 7, 7
	})
	if effectstest.HasEffectiveKeyword(t, g, lord, "haste") {
		t.Error("the Lord has haste; it only attacks as though it did")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(lord, opp.ID); err != nil {
		t.Errorf("a Lord that did not enter this turn could not attack: %v", err)
	}

	// A Lord cast this turn may not.
	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	effectstest.AdvanceToMain(t, g2)
	id := effectstest.CastCatalogSpell(t, g2, "Chaos Lord", "Creature — Human", o851Lord, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(id) {
		t.Fatal("the Lord did not resolve")
	}
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(id, opp2.ID); err == nil {
		t.Error("a Lord that entered this turn attacked")
	}
}

func TestS82OldB851GameOfChaosMovesDoublingStakesAndAsksTheRightPlayer(t *testing.T) {
	reachedTwo, reachedWin, reachedLoss := false, false, false
	for attempt := 0; attempt < 120 && !(reachedTwo && reachedWin && reachedLoss); attempt++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		myLife, oppLife := me.Life, opp.Life
		effectstest.CastCatalogSpell(t, g, "Game of Chaos", "Sorcery", o851Chaos, effectstest.PlayerRefs(opp.ID))
		effectstest.PassPriorityAroundTable(t, g)
		wantMe, wantOpp := myLife, oppLife
		flips := 0
		for i := 0; i < 6; i++ {
			call := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
			if call == nil {
				break
			}
			before := len(g.Events)
			if err := g.ResolveCoinCall(call.ID, me.ID, "heads"); err != nil {
				t.Fatal(err)
			}
			won := false
			for _, ev := range g.Events[before:] {
				if ev.Kind == game.EventFlipCoin {
					won = ev.Won
				}
			}
			stakes := 1 << flips
			flips++
			decider := me
			if won {
				wantMe, wantOpp = wantMe+stakes, wantOpp-stakes
				reachedWin = true
			} else {
				wantMe, wantOpp = wantMe-stakes, wantOpp+stakes
				decider = opp
				reachedLoss = true
			}
			if me.Life != wantMe || opp.Life != wantOpp {
				t.Fatalf("flip %d (won %v): life %d/%d, want %d/%d", flips, won, me.Life, opp.Life, wantMe, wantOpp)
			}
			if effectstest.LatestConfirmFor(g, decider.ID) == nil {
				t.Fatalf("flip %d (won %v): the %s was not asked whether to flip again", flips, won, map[bool]string{true: "controller", false: "opponent"}[decider == me])
			}
			again := flips < 2+attempt%2
			effectstest.AnswerMayChoice(t, g, decider.ID, again)
			if !again {
				break
			}
		}
		if flips >= 2 {
			reachedTwo = true
		}
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != wantMe || opp.Life != wantOpp {
			t.Fatalf("attempt %d: final life %d/%d, want %d/%d", attempt, me.Life, opp.Life, wantMe, wantOpp)
		}
	}
	if !reachedTwo || !reachedWin || !reachedLoss {
		t.Errorf("coverage: two-flip chain %v, a win %v, a loss %v", reachedTwo, reachedWin, reachedLoss)
	}
}
