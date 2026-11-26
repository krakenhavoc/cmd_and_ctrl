package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b465_cards_test.go — Mirrodin Besieged, S82 OLD batch 465 (#5576).
// Every test plays the real catalog entry through the engine.

const (
	oldB465MirranMettle    = "35506b6b-6476-4830-9baa-4051c18322f0"
	oldB465SpireSerpent    = "6c0c5832-416f-405f-bb44-c24be6f9d081"
	oldB465SpiralingDuel   = "cc803d43-cbbe-4964-8077-ba405a741047"
	oldB465ConcussiveBolt  = "e99958bf-458c-403b-aae2-91d60c055ee2"
	oldB465HorrifyingRev   = "36ee4179-83a1-472a-a48c-394c4c0457bc"
	oldB465MitoticManip    = "69cc006d-172f-4e4c-bb75-10d772b061b1"
	oldB465MetallicMastery = "7de034a4-a42a-4e60-9ec9-5e9bb9987cfa"
	oldB465GruesomeEncore  = "f87bf51e-6218-4418-ae3e-98055e9601e4"
	oldB465RustedSlasher   = "098cc111-493b-470b-929c-67deb981240c"
	oldB465Gnathosaur      = "7f52993c-a20a-48fe-b463-2a33873b489a"
	oldB465SpinEngine      = "88e98269-fe0e-42f8-830d-776aa06a58dc"
)

func TestOldB465EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB465MirranMettle:    "Mirran Mettle",
		oldB465SpireSerpent:    "Spire Serpent",
		oldB465SpiralingDuel:   "Spiraling Duelist",
		oldB465ConcussiveBolt:  "Concussive Bolt",
		oldB465HorrifyingRev:   "Horrifying Revelation",
		oldB465MitoticManip:    "Mitotic Manipulation",
		oldB465MetallicMastery: "Metallic Mastery",
		oldB465GruesomeEncore:  "Gruesome Encore",
		oldB465RustedSlasher:   "Rusted Slasher",
		oldB465Gnathosaur:      "Gnathosaur",
		oldB465SpinEngine:      "Spin Engine",
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

func oldB465Artifacts(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.B12Permanent(g, owner, "Rock", "Artifact"))
	}
	return ids
}

func TestOldB465MirranMettleIsPlusTwoUnlessMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Mirran Mettle", "Instant", oldB465MirranMettle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Fatalf("without metalcraft the Bear is %d/%d, want 4/4", p, tough)
	}
	oldB465Artifacts(g, me.ID, 3)
	effectstest.CastCatalogSpell(t, g, "Mirran Mettle", "Instant", oldB465MirranMettle, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 8 || tough != 8 {
		t.Fatalf("with metalcraft the Bear is %d/%d, want 8/8 (4/4 + 4/4)", p, tough)
	}
}

func TestOldB465MirranMettleNeedsACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[0].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Mirran Mettle", "Instant", oldB465MirranMettle, effectstest.CardRefs(rock)); err == nil {
		t.Fatal("Mirran Mettle was cast at a noncreature artifact")
	}
}

func TestOldB465SpireSerpentGrowsAndAttacksOnlyWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := effectstest.B12Push(g, me.ID, "Spire Serpent", "Creature — Serpent", oldB465SpireSerpent, 3, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(serpent, opp.ID); err == nil {
		t.Fatal("a defender attacked without metalcraft")
	}
	if p, tough := effectstest.EffectivePTOf(t, g, serpent); p != 3 || tough != 5 {
		t.Errorf("without metalcraft the Serpent is %d/%d, want 3/5", p, tough)
	}
	oldB465Artifacts(g, me.ID, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, serpent); p != 5 || tough != 7 {
		t.Errorf("with metalcraft the Serpent is %d/%d, want 5/7", p, tough)
	}
	if err := g.DeclareAttacker(serpent, opp.ID); err != nil {
		t.Fatalf("the Serpent could not attack with metalcraft: %v", err)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, serpent, "defender") {
		t.Error("the waiver removed defender")
	}
}

func TestOldB465SpiralingDuelistHasDoubleStrikeOnlyWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	duelist := effectstest.B12Push(g, me.ID, "Spiraling Duelist", "Creature — Human Berserker", oldB465SpiralingDuel, 3, 1)
	if effectstest.HasEffectiveKeyword(t, g, duelist, "double strike") {
		t.Fatal("double strike without metalcraft")
	}
	arts := oldB465Artifacts(g, me.ID, 3)
	if !effectstest.HasEffectiveKeyword(t, g, duelist, "double strike") {
		t.Fatal("no double strike with three artifacts")
	}
	effectstest.P1Destroy(t, g, arts[0])
	if effectstest.HasEffectiveKeyword(t, g, duelist, "double strike") {
		t.Error("double strike stayed after the count fell to two")
	}
}

func TestOldB465ConcussiveBoltDamagesAPlayerAndBarsBlocksWithMetalcraft(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Concussive Bolt", "Sorcery", oldB465ConcussiveBolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("life %d → %d, want 4 damage", life, opp.Life)
	}
	oldB465CanBlock(t, g, attacker, blocker, true)
	oldB465Artifacts(g, me.ID, 3)
	effectstest.CastCatalogSpell(t, g, "Concussive Bolt", "Sorcery", oldB465ConcussiveBolt, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	oldB465CanBlock(t, g, attacker, blocker, false)
	late := effectstest.B12Creature(g, opp.ID, "Latecomer", "Creature — Bear", 1, 1)
	oldB465CanBlock(t, g, attacker, late, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	oldB465CanBlock(t, g, attacker, mine, true)
}

func TestOldB465ConcussiveBoltAtAPlaneswalkerBarsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	life := opp.Life
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: opp.ID, Controller: opp.ID})
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	oldB465Artifacts(g, me.ID, 3)
	effectstest.CastCatalogSpell(t, g, "Concussive Bolt", "Sorcery", oldB465ConcussiveBolt, effectstest.CardRefs(walker))
	effectstest.PassPriorityAroundTable(t, g)
	if effects := effectstest.BattlefieldNamed(g, opp.ID, "Their Walker"); effects != 0 {
		t.Error("4 damage should have killed the 3-loyalty planeswalker")
	}
	if opp.Life != life {
		t.Errorf("the planeswalker's controller lost life: %d → %d", life, opp.Life)
	}
	oldB465CanBlock(t, g, attacker, blocker, false)
}

func oldB465CanBlock(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(attacker)
		b, _ := g.LookupCardForEffect(blocker)
		if r := g.BlockPairRefusalLocked(&a, &b); r.Legal() != want {
			t.Errorf("blocker %s may block = %v, want %v (%q)", b.Name, r.Legal(), want, r.Reason)
		}
	})
}

func TestOldB465HorrifyingRevelationDiscardsThenMills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 3; i++ {
		effectstest.HandSpell(opp, "Card", "Sorcery", "{9}")
	}
	hand, lib, gy := opp.Hand.Size(), opp.Library.Size(), opp.Graveyard.Size()
	effectstest.CastCatalogSpell(t, g, "Horrifying Revelation", "Sorcery", oldB465HorrifyingRev, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) == nil {
		t.Fatal("the target player was not asked to discard")
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	g.SettleResolution()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-1)
	}
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d, want %d (one mill)", opp.Library.Size(), lib-1)
	}
	if opp.Graveyard.Size() != gy+2 {
		t.Errorf("graveyard %d, want %d (the discard and the milled card)", opp.Graveyard.Size(), gy+2)
	}
}

func TestOldB465MetallicMasteryStealsAnArtifactUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Tapped: true, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Metallic Mastery", "Sorcery", oldB465MetallicMastery, effectstest.CardRefs(golem))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, golem)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the Golem is controlled by %v, want the caster", c.Controller)
	}
	if c.Tapped {
		t.Error("the artifact stays tapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, golem, "haste") {
		t.Error("the artifact has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, golem); c.Controller != opp.ID {
		t.Error("control did not revert at cleanup")
	}
}

func TestOldB465MetallicMasteryRefusesANonartifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Metallic Mastery", "Sorcery", oldB465MetallicMastery, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Metallic Mastery was cast at a nonartifact creature")
	}
}

func TestOldB465GruesomeEncoreReanimatesAnOpponentsCreatureForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "My Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := effectstest.CastCatalogSpellErr(t, g, "Gruesome Encore", "Sorcery", oldB465GruesomeEncore, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("Gruesome Encore was cast at a creature card in the caster's own graveyard")
	}
	effectstest.CastCatalogSpell(t, g, "Gruesome Encore", "Sorcery", oldB465GruesomeEncore, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok || c.Controller != me.ID {
		t.Fatalf("the Ogre is on the battlefield under %v, want the caster", c.Controller)
	}
	if !effectstest.HasEffectiveKeyword(t, g, dead, "haste") {
		t.Error("the Ogre has no haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects := effectstest.BattlefieldNamed(g, me.ID, "Dead Ogre"); effects != 0 {
		t.Fatal("the Ogre is still on the battlefield after the end step")
	}
	inExile := false
	for _, e := range g.Exile.Cards {
		if e.InstanceID == dead {
			inExile = true
		}
	}
	if !inExile {
		t.Error("the Ogre should be exiled, not put in a graveyard")
	}
}

func TestOldB465GruesomeEncoreExilesTheCreatureIfItWouldDie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Gruesome Encore", "Sorcery", oldB465GruesomeEncore, effectstest.CardRefs(dead))
	effectstest.PassPriorityAroundTable(t, g)
	_ = me
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.New(), dead, 9); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	g.RunStateChecksForTest()
	for _, c := range opp.Graveyard.Cards {
		if c.InstanceID == dead {
			t.Fatal("the creature went to the graveyard instead of exile")
		}
	}
	inExile := false
	for _, e := range g.Exile.Cards {
		if e.InstanceID == dead {
			inExile = true
		}
	}
	if !inExile {
		t.Error("the creature is not in exile")
	}
}

func TestOldB465MitoticManipulationPutsASameNamedCardOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Creature(g, g.Seats[1].ID, "Twin", "Creature — Bear", 2, 2)
	me.Library.Cards = nil
	var ids []uuid.UUID
	for _, c := range []game.Card{
		{Name: "Other", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{Name: "Twin", TypeLine: "Creature — Bear", Power: 2, Toughness: 2},
		{Name: "Spell", TypeLine: "Sorcery"},
		{Name: "Third", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{Name: "Fourth", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{Name: "Fifth", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{Name: "Sixth", TypeLine: "Creature — Elf", Power: 1, Toughness: 1},
		{Name: "Twin", TypeLine: "Creature — Bear", Power: 2, Toughness: 2}, // eighth: out of reach
	} {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		ids = append(ids, c.InstanceID)
		g.WithWriteLock(func() { me.Library.PushBottom(c) })
	}
	effectstest.CastCatalogSpell(t, g, "Mitotic Manipulation", "Sorcery", oldB465MitoticManip, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCards(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	if len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != ids[1] {
		t.Fatalf("offered %v, want only the Twin among the top seven", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Error("the chosen card did not enter the battlefield")
	}
	if g.Battlefield.Contains(ids[7]) {
		t.Error("the eighth card was not looked at and must not enter")
	}
}

func TestOldB465GnathosaurSacrificesAnArtifactForTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saur := effectstest.B12Push(g, me.ID, "Gnathosaur", "Creature — Dinosaur", oldB465Gnathosaur, 5, 4)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, saur, "trample") {
		t.Fatal("trample before the ability")
	}
	effectstest.B16Activate(t, g, me.ID, saur, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as a cost")
	}
	if !effectstest.HasEffectiveKeyword(t, g, saur, "trample") {
		t.Error("no trample after the ability resolved")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, saur, "trample") {
		t.Error("trample outlasted the turn")
	}
}

func TestOldB465GnathosaurNeedsAnArtifactToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	saur := effectstest.B12Push(g, me.ID, "Gnathosaur", "Creature — Dinosaur", oldB465Gnathosaur, 5, 4)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, saur, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the ability was activated with nothing to sacrifice")
	}
}

func TestOldB465RustedSlasherRegeneratesBySacrificingAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slasher := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rusted Slasher",
		TypeLine: "Artifact Creature — Phyrexian Horror", OracleID: oldB465RustedSlasher, Power: 4, Toughness: 1, Owner: me.ID, Controller: me.ID})
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, slasher, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact is sacrificed as a cost")
	}
	effectstest.P1Destroy(t, g, slasher)
	if !g.Battlefield.Contains(slasher) {
		t.Fatal("the Slasher was destroyed despite its regeneration shield")
	}
	effectstest.P1Destroy(t, g, slasher)
	if g.Battlefield.Contains(slasher) {
		t.Error("the shield was used once; the second destruction should stick")
	}
}

func TestOldB465SpinEngineStopsOneCreatureFromBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	engine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Spin Engine",
		TypeLine: "Artifact Creature — Construct", OracleID: oldB465SpinEngine, Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	kept := effectstest.B12Creature(g, opp.ID, "Kept Back", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, engine, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(kept)}, "R")
	oldB465CanBlock(t, g, engine, kept, false)
	oldB465CanBlock(t, g, engine, free, true)
}
