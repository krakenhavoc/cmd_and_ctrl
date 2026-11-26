package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b340_cards_test.go — OLD batch 340 (Fate Reforged: other creatures,
// enters and dies triggers, counterspells and bounce, instants and sorceries,
// token makers, #5552): the cards played through the engine.

const (
	oldB340YetiOracle       = "5c179fae-f767-4246-88d3-aa38da5bf6f4"
	oldB340EnforcersOracle  = "5e4facd9-d416-4ade-ba71-6da32b12a891"
	oldB340ConstructOracle  = "6b87a933-b0d2-4120-b4a0-65ab0f3dbe48"
	oldB340SureshotOracle   = "c56726a6-d424-46a0-b3b3-2b44fbcbc63a"
	oldB340AvengersOracle   = "97a1b49e-d6bc-46ba-88c1-991057e64170"
	oldB340DisdainOracle    = "a8fe6204-9f55-45e5-99b3-43f2761ce8a3"
	oldB340LoyaltiesOracle  = "0ee9a743-c678-4a76-b106-53f57a84848d"
	oldB340SupplantOracle   = "7e49f689-996a-477e-a327-32b69fbc0d85"
	oldB340RetainersOracle  = "a27938d2-9a15-4a87-9e77-dea37347cf62"
	oldB340DragonrageOracle = "fb16c4b0-9ac4-4664-9b57-97bfc22e5de9"
)

func TestOldB340EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB340YetiOracle:       "Hungering Yeti",
		oldB340EnforcersOracle:  "Bloodfire Enforcers",
		oldB340ConstructOracle:  "Ugin's Construct",
		oldB340SureshotOracle:   "Orc Sureshot",
		oldB340AvengersOracle:   "Sage-Eye Avengers",
		oldB340DisdainOracle:    "Rakshasa's Disdain",
		oldB340LoyaltiesOracle:  "Shifting Loyalties",
		oldB340SupplantOracle:   "Supplant Form",
		oldB340RetainersOracle:  "Hewed Stone Retainers",
		oldB340DragonrageOracle: "Dragonrage",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name=%q completeness=%v caveats=%v, want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB340Color(g *game.Game, id uuid.UUID, colors ...string) {
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).Colors = colors })
}

func TestOldB340YetiHasFlashOnlyWithAGreenOrBluePermanent(t *testing.T) {
	for _, tc := range []struct {
		name   string
		colors []string
		flash  bool
	}{{"green", []string{"G"}, true}, {"blue", []string{"U"}, true}, {"red", []string{"R"}, false}, {"colorless", nil, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		p := effectstest.B12Permanent(g, me.ID, "Other", "Creature — Bear")
		oldB340Color(g, p, tc.colors...)
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		yeti := effectstest.HandCardFull(me, "Hungering Yeti", "Creature — Yeti", "{4}{R}", oldB340YetiOracle, []string{"R"})
		effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}{C}{R}")
		err := g.CastSpell(me.ID, yeti, game.CastSpellParams{Strict: true})
		if (err == nil) != tc.flash {
			t.Errorf("%s permanent: cast at the beginning of combat err=%v, want flash=%v", tc.name, err, tc.flash)
		}
	}
}

func TestOldB340EnforcersNeedAnInstantAndASorceryInTheGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name  string
		cards []string
		want  bool
	}{
		{"both", []string{"Instant", "Sorcery"}, true},
		{"only an instant", []string{"Instant", "Instant"}, false},
		{"only a sorcery", []string{"Sorcery"}, false},
		{"neither", []string{"Creature — Bear"}, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		e := effectstest.B12Push(g, me.ID, "Bloodfire Enforcers", "Creature — Human Monk", oldB340EnforcersOracle, 5, 2)
		for _, tl := range tc.cards {
			effectstest.PushGraveyardCardTyped(me, "Card", tl)
		}
		g.RunStateChecksForTest()
		for _, kw := range []string{"first strike", "trample"} {
			if got := effectstest.HasEffectiveKeyword(t, g, e, kw); got != tc.want {
				t.Errorf("%s: %s = %v, want %v", tc.name, kw, got, tc.want)
			}
		}
	}
}

func TestOldB340ConstructSacrificesAColoredPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	colored := effectstest.B12Permanent(g, me.ID, "Colored", "Creature — Bear")
	oldB340Color(g, colored, "G")
	other := effectstest.B12Permanent(g, me.ID, "Colorless", "Artifact")
	construct := effectstest.CastAndResolveCreature(t, g, "Ugin's Construct", "Artifact Creature — Construct", oldB340ConstructOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.SacrificeChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no sacrifice prompt")
	}
	opts := effectstest.SacrificeOptions(prompt)
	if !effectstest.HasID(opts, colored) || effectstest.HasID(opts, other) || effectstest.HasID(opts, construct) {
		t.Fatalf("options %v: only the colored permanent is a legal sacrifice", opts)
	}
	effectstest.AnswerSacrifice(t, g, me.ID, colored)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(colored) || !g.Battlefield.Contains(construct) {
		t.Error("the colored permanent was not the one sacrificed")
	}
}

func TestOldB340ConstructWithNoColoredPermanentSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	artifact := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	construct := effectstest.CastAndResolveCreature(t, g, "Ugin's Construct", "Artifact Creature — Construct", oldB340ConstructOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, me.ID) != nil {
		t.Error("a sacrifice prompt was asked with nothing colored to sacrifice")
	}
	if !g.Battlefield.Contains(artifact) || !g.Battlefield.Contains(construct) {
		t.Error("a colorless permanent was sacrificed")
	}
}

func TestOldB340SureshotShrinksOnEachOtherCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Orc Sureshot", "Creature — Orc Archer", oldB340SureshotOracle, 4, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Bear", "", 3, 3)
	effectstest.CastAndResolveCreature(t, g, "Hill Giant", "Creature — Giant", "")
	effectstest.MscM2PickTargets(t, g, theirs)
	if c := effectstest.GaLayered(t, g, theirs); c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB340SureshotDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Bear", "", 3, 3)
	effectstest.CastAndResolveCreature(t, g, "Orc Sureshot", "Creature — Orc Archer", oldB340SureshotOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
		t.Error("the Sureshot's own entry triggered it")
	}
	if c := effectstest.GaLayered(t, g, theirs); c.CurrentPower() != 3 {
		t.Errorf("their creature is %d/%d, want it untouched", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestOldB340AvengersBouncesOnlyACreatureWithLesserPower(t *testing.T) {
	for _, tc := range []struct {
		name  string
		power int
		want  bool
	}{{"lesser", 3, true}, {"equal", 4, false}, {"greater", 6, false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		av := effectstest.B12Push(g, me.ID, "Sage-Eye Avengers", "Creature — Djinn Monk", oldB340AvengersOracle, 4, 5)
		target := effectstest.B12Push(g, opp.ID, "Target", "Creature — Giant", "", tc.power, 4)
		effectstest.DeclareAttack(t, g, opp.ID, av)
		effectstest.MscM2PickTargets(t, g, target)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if bounced := !g.Battlefield.Contains(target); bounced != tc.want {
			t.Errorf("%s: bounced=%v, want %v", tc.name, bounced, tc.want)
		}
	}
}

func TestOldB340DisdainTaxesByYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Old Card", "Sorcery")
	}
	victim := effectstest.CastCatalogSpell(t, g, "Divination", "Sorcery", effectstest.DivinationDelveOracle, nil)
	effectstest.CastCatalogSpell(t, g, "Rakshasa's Disdain", "Instant", oldB340DisdainOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PendingOfKind(g, game.PendingChoicePayUnless)
	if c == nil {
		t.Fatalf("no pay-unless prompt: %+v", g.PendingChoices)
	}
	if c.PayCost != "{3}" {
		t.Errorf("the prompt asks for %q, want {3}", c.PayCost)
	}
}

func TestOldB340ShiftingLoyaltiesExchangesPermanentsThatShareAType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "Mine", "Artifact Creature — Golem")
	theirs := effectstest.B12Permanent(g, opp.ID, "Theirs", "Creature — Bear")
	land := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	refs := func(a, b uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}}
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Shifting Loyalties", "Sorcery", oldB340LoyaltiesOracle, refs(mine, land)); err == nil {
		t.Error("a creature and a land, which share no card type, were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Shifting Loyalties", "Sorcery", oldB340LoyaltiesOracle, refs(mine, theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ControllerOf(t, g, mine) != opp.ID || effectstest.ControllerOf(t, g, theirs) != me.ID {
		t.Error("control was not exchanged")
	}
}

func TestOldB340SupplantFormBouncesAndCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Giant", "", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Supplant Form", "Instant", oldB340SupplantOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !oldB340InHand(opp, victim) {
		t.Error("the creature did not return to its owner's hand")
	}
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Victim")
	if n != 1 {
		t.Fatalf("%d token copies under my control, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Victim" && c.Controller == me.ID && (c.Power != 5 || c.Toughness != 5 || !c.IsToken()) {
			t.Errorf("the copy is %d/%d token=%v, want a 5/5 token", c.Power, c.Toughness, c.IsToken())
		}
	}
}

func oldB340InHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}

func TestOldB340RetainersNeedAnotherSpellThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	retainers := effectstest.HandCardFull(me, "Hewed Stone Retainers", "Artifact Creature — Golem", "{3}", oldB340RetainersOracle, nil)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, retainers, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no other spell this turn")
	}
	spell := effectstest.HandCardFull(me, "Cheap Spell", "Instant", "", "", nil)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the first spell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{C}")
	if err := g.CastSpell(me.ID, retainers, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("cast after another spell: %v", err)
	}
}

func TestOldB340DragonrageAddsManaAndGrantsFirebreathingToAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "B", "Creature — Bear", "", 2, 2)
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	rage := effectstest.HandCardFull(me, "Dragonrage", "Instant", "{2}{R}", oldB340DragonrageOracle, []string{"R"})
	if err := g.CastSpell(me.ID, rage, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Dragonrage in the declare attackers step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 2 {
		t.Errorf("pool has %d mana, want 2 ({R} for each of two attackers)", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate the granted ability: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for _, id := range []uuid.UUID{a, b} {
		if c := effectstest.GaLayered(t, g, id); c.CurrentPower() != 3 {
			t.Errorf("%s has power %d, want 3 after {R}: +1/+0", c.Name, c.CurrentPower())
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, idle, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a creature that was not attacking got the ability")
	}
}

const (
	oldB340SoulfireOracle = "e65769f4-50ae-482b-933b-ac5cf3843f2a"
	oldB340BoltOracle     = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func oldB340Zone(z *game.Zone, name string) int {
	n := 0
	for _, c := range z.Cards {
		if c.Name == name {
			n++
		}
	}
	return n
}

func TestOldB340SoulfireIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(oldB340SoulfireOracle)
	if !ok || spec.Name != "Soulfire Grand Master" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("Soulfire Grand Master: registered=%v spec=%+v", ok, spec)
	}
}

func TestOldB340SoulfireGivesYourInstantsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Soulfire Grand Master", "Creature — Human Monk", oldB340SoulfireOracle, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB340BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d → %d, want 3 gained from a lifelinked Lightning Bolt", life, me.Life)
	}
}

func TestOldB340SoulfireReturnsTheNextInstantOrSorceryToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sg := effectstest.B12Push(g, me.ID, "Soulfire Grand Master", "Creature — Human Monk", oldB340SoulfireOracle, 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.Acrb1AddMana(t, g, me.ID, "{C}{C}{U}{R}")
	if err := g.ActivateCatalogAbility(me.ID, sg, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bolt := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB340BoltOracle, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB340Zone(me.Hand, "Lightning Bolt") != 1 || oldB340Zone(me.Graveyard, "Lightning Bolt") != 0 {
		t.Fatalf("first Bolt: hand=%d graveyard=%d, want it back in hand",
			oldB340Zone(me.Hand, "Lightning Bolt"), oldB340Zone(me.Graveyard, "Lightning Bolt"))
	}
	// The promise was for the next spell only: the second one goes to the graveyard.
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB340BoltOracle, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB340Zone(me.Graveyard, "Lightning Bolt") != 1 {
		t.Errorf("second Bolt: graveyard=%d, want 1", oldB340Zone(me.Graveyard, "Lightning Bolt"))
	}
}

func TestOldB340SoulfireWithoutActivatingSendsSpellsToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Soulfire Grand Master", "Creature — Human Monk", oldB340SoulfireOracle, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB340BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB340Zone(me.Graveyard, "Lightning Bolt") != 1 || oldB340Zone(me.Hand, "Lightning Bolt") != 0 {
		t.Error("a spell cast with no promise left the graveyard")
	}
}
