package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b728_cards_test.go — Urza's Saga pump and combat tricks, batch 728
// (#5624). Each test plays the real catalog entry through the engine and
// asserts what its text says.

const (
	oldB728MobileFort         = "0547e426-e23e-4c60-ba6b-cfd4b5ff8485"
	oldB728Morphling          = "19f233b2-8be6-4172-abf9-b132fc3393d9"
	oldB728Retaliation        = "51571fa8-fd52-4439-8331-4eb4f651e764"
	oldB728HoppingAutomaton   = "7848e040-e1b4-4f81-922e-b1eca8fe4398"
	oldB728Symbiosis          = "787b494b-731b-4e99-8e46-157bc86f6eea"
	oldB728KarnSilverGolem    = "8a7c6391-020a-4d7a-8abb-702631bbee56"
	oldB728RavenousSkirge     = "b8ae71b3-8ee7-49a0-87ea-7e938166ccb3"
	oldB728ViashinoWeaponsmth = "f4d17f41-efd6-4173-ab6e-28b6fbcee10d"
)

var oldB728Cards = map[string]string{
	oldB728MobileFort: "Mobile Fort", oldB728Morphling: "Morphling", oldB728Retaliation: "Retaliation",
	oldB728HoppingAutomaton: "Hopping Automaton", oldB728Symbiosis: "Symbiosis",
	oldB728KarnSilverGolem: "Karn, Silver Golem", oldB728RavenousSkirge: "Ravenous Skirge",
	oldB728ViashinoWeaponsmth: "Viashino Weaponsmith",
}

func TestOldB728EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB728Cards {
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

func oldB728Permanent(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Keywords: keywords, Owner: owner, Controller: owner,
	})
}

func oldB728ActivateStrict(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, idx int, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	if mana != "" {
		effectstest.FloatMana(t, g, who, mana)
	}
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, idx, params)
}

func TestOldB728KarnGetsMinusFourPlusFourWhenItBlocksOrBecomesBlocked(t *testing.T) {
	for _, blocking := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		karn := oldB728Permanent(g, me.ID, "Karn, Silver Golem", "Legendary Artifact Creature — Golem", oldB728KarnSilverGolem, 4, 4)
		other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
		if blocking {
			effectstest.WoeS2Block(g, me.ID, karn, other)
		} else {
			a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
			b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
			effectstest.WoeS2Block(g, opp.ID, a, karn)
			effectstest.WoeS2Block(g, opp.ID, b, karn)
			// The combat engine also emits one becomes-blocked event for the attacker.
			g.WithWriteLock(func() {
				g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: karn, CardID: karn, Target: karn})
			})
		}
		effectstest.PassPriorityAroundTable(t, g)
		if p, th := effectstest.EffectivePTOf(t, g, karn); p != 0 || th != 8 {
			t.Errorf("blocking=%v: Karn is %d/%d, want 0/8 (-4/+4 once)", blocking, p, th)
		}
	}
}

func TestOldB728KarnAnimatesANoncreatureArtifactToItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	karn := oldB728Permanent(g, me.ID, "Karn, Silver Golem", "Legendary Artifact Creature — Golem", oldB728KarnSilverGolem, 4, 4)
	rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Costly Rock", TypeLine: "Artifact", ManaCost: "{3}",
		Owner: me.ID, Controller: me.ID,
	})
	if err := oldB728ActivateStrict(t, g, me, karn, 0, "{C}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.LayeredCard(t, g, rock)
	if !got.IsCreature() || !got.HasCardType("artifact") {
		t.Fatalf("the artifact is %q, want an artifact creature", got.TypeLine)
	}
	if p, th := effectstest.EffectivePTOf(t, g, rock); p != 3 || th != 3 {
		t.Errorf("the artifact is %d/%d, want 3/3", p, th)
	}
}

func TestOldB728KarnCannotTargetACreatureArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	karn := oldB728Permanent(g, me.ID, "Karn, Silver Golem", "Legendary Artifact Creature — Golem", oldB728KarnSilverGolem, 4, 4)
	if err := oldB728ActivateStrict(t, g, me, karn, 0, "{C}", game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: karn}},
	}); err == nil {
		t.Error("targeted an artifact creature")
	}
}

func TestOldB728MorphlingActivatedAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := oldB728Permanent(g, me.ID, "Morphling", "Creature — Shapeshifter", oldB728Morphling, 3, 3)
	run := func(idx int, mana string) {
		t.Helper()
		if err := oldB728ActivateStrict(t, g, me, m, idx, mana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("ability %d: %v", idx, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.TapForTest(t, g, m)
	run(0, "{U}")
	if effectstest.LayeredCard(t, g, m).Tapped {
		t.Error("{U}: Untap did not untap the Morphling")
	}
	run(1, "{U}")
	if !effectstest.HasEffectiveKeyword(t, g, m, "flying") {
		t.Error("no flying")
	}
	run(2, "{U}")
	if !effectstest.HasEffectiveKeyword(t, g, m, "shroud") {
		t.Error("no shroud")
	}
	run(3, "{C}")
	if p, th := effectstest.EffectivePTOf(t, g, m); p != 4 || th != 2 {
		t.Errorf("after +1/-1 the Morphling is %d/%d, want 4/2", p, th)
	}
	run(4, "{C}")
	run(4, "{C}")
	if p, th := effectstest.EffectivePTOf(t, g, m); p != 2 || th != 4 {
		t.Errorf("after two -1/+1 the Morphling is %d/%d, want 2/4", p, th)
	}
}

func TestOldB728MorphlingPumpCanKillIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	m := oldB728Permanent(g, me.ID, "Morphling", "Creature — Shapeshifter", oldB728Morphling, 3, 3)
	for i := 0; i < 3; i++ {
		if err := oldB728ActivateStrict(t, g, me, m, 3, "{C}", game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("+1/-1 #%d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(m) {
		t.Error("a Morphling at 6/0 stayed on the battlefield")
	}
}

func TestOldB728SymbiosisPumpsTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 3, 3)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Symbiosis", "Instant", oldB728Symbiosis,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, tc := range []struct {
		id         uuid.UUID
		power, tou int
	}{{a, 4, 4}, {b, 5, 5}, {c, 1, 1}} {
		if p, th := effectstest.EffectivePTOf(t, g, tc.id); p != tc.power || th != tc.tou {
			t.Errorf("a creature is %d/%d, want %d/%d", p, th, tc.power, tc.tou)
		}
	}
}

func TestOldB728SymbiosisNeedsTwoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	a := effectstest.B12Creature(g, g.Seats[0].ID, "A", "Creature — Bear", 2, 2)
	err := effectstest.CastCatalogSpellErr(t, g, "Symbiosis", "Instant", oldB728Symbiosis,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}})
	if err == nil {
		t.Error("cast Symbiosis with one target")
	}
}

func TestOldB728HoppingAutomatonShrinksAndFliesForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := oldB728Permanent(g, me.ID, "Hopping Automaton", "Artifact Creature — Construct", oldB728HoppingAutomaton, 2, 2)
	if err := oldB728ActivateStrict(t, g, me, h, 0, "", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, h); p != 1 || th != 1 {
		t.Errorf("the Automaton is %d/%d, want 1/1", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, h, "flying") {
		t.Error("no flying")
	}
}

func TestOldB728MobileFortPumpsOncePerTurnAndLosesDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fort := oldB728Permanent(g, me.ID, "Mobile Fort", "Artifact Creature — Wall", oldB728MobileFort, 0, 6, "defender")
	if err := oldB728ActivateStrict(t, g, me, fort, 0, "{C}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, fort); p != 3 || th != 5 {
		t.Errorf("the Fort is %d/%d, want 3/5", p, th)
	}
	if err := oldB728ActivateStrict(t, g, me, fort, 0, "{C}{C}{C}", game.ActivateAbilityParams{}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

func TestOldB728MobileFortCanAttackAfterItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fort := oldB728Permanent(g, me.ID, "Mobile Fort", "Artifact Creature — Wall", oldB728MobileFort, 0, 6, "defender")
	effectstest.AdvanceToMain(t, g)
	if err := oldB728ActivateStrict(t, g, me, fort, 0, "{C}{C}{C}", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fort, opp.ID); err != nil {
		t.Errorf("the Fort could not attack after its ability: %v", err)
	}
}

func TestOldB728MobileFortCannotAttackWithoutItsAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fort := oldB728Permanent(g, me.ID, "Mobile Fort", "Artifact Creature — Wall", oldB728MobileFort, 0, 6, "defender")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fort, opp.ID); err == nil {
		t.Error("a defender attacked")
	}
}

func TestOldB728RetaliationGivesEveryCreatureYouControlTheTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 1)
	mineBlocker := effectstest.B12Creature(g, me.ID, "My Blocker", "Creature — Bear", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Retaliation", "Enchantment", oldB728Retaliation, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.WoeS2Block(g, opp.ID, blocker, mine)
	effectstest.WoeS2Block(g, me.ID, mineBlocker, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, mine); p != 3 || th != 3 {
		t.Errorf("my blocked creature is %d/%d, want 3/3", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's blocked creature is %d/%d, want 2/2", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, mineBlocker); p != 1 || th != 1 {
		t.Errorf("my blocking creature is %d/%d, want 1/1 (blocking is not becoming blocked)", p, th)
	}
}

func TestOldB728RavenousSkirgeGetsPlusTwoWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	skirge := oldB728Permanent(g, me.ID, "Ravenous Skirge", "Creature — Phyrexian Imp", oldB728RavenousSkirge, 1, 1, "flying")
	effectstest.DeclareAttack(t, g, opp.ID, skirge)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, skirge); p != 3 || th != 1 {
		t.Errorf("the Skirge is %d/%d, want 3/1", p, th)
	}
}

func TestOldB728ViashinoWeaponsmithGetsPlusTwoPerBlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smith := oldB728Permanent(g, me.ID, "Viashino Weaponsmith", "Creature — Lizard", oldB728ViashinoWeaponsmth, 2, 2)
	a := effectstest.B12Creature(g, opp.ID, "Blocker A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, opp.ID, "Blocker B", "Creature — Bear", 1, 1)
	effectstest.WoeS2Block(g, opp.ID, a, smith)
	effectstest.WoeS2Block(g, opp.ID, b, smith)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, smith); p != 6 || th != 6 {
		t.Errorf("blocked by two creatures it is %d/%d, want 6/6", p, th)
	}
}

func TestOldB728ViashinoWeaponsmithDoesNotGrowWhenBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	smith := oldB728Permanent(g, opp.ID, "Viashino Weaponsmith", "Creature — Lizard", oldB728ViashinoWeaponsmth, 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.WoeS2Block(g, opp.ID, smith, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, smith); p != 2 || th != 2 {
		t.Errorf("blocking it is %d/%d, want 2/2", p, th)
	}
}
