package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b415_test.go — Gatecrash, batch 415 (#5565): every card played
// through the engine.

const (
	oldB415DimirCharm  = "4d99e786-585d-4026-9253-19741fb357de"
	oldB415Grisly      = "8c83d21a-6138-409f-8e00-75dcd14388e2"
	oldB415RustScarab  = "c3bd1d38-e73f-4e30-b22a-dca0a0e4d701"
	oldB415Drake       = "6e33ec9a-e148-4aa6-bff8-b356c45fb369"
	oldB415Realmwright = "bad869dd-e844-46ca-be53-73c4bc19d09c"
	oldB415Ragebeast   = "e5fa93a3-0b2b-4477-ad3f-4fc1b325ca19"
	oldB415Tracer      = "9b38ab81-7ee7-4d89-878b-3507e44836ca"
	oldB415Biovis      = "e99fde0d-ca28-4a49-abe0-d01cf4e322b6"
	oldB415Rogue       = "fa8ae8ac-4ad2-4cda-9f6b-4ec634526658"
	oldB415RiotGear    = "3e150025-51c2-4b92-970c-a89e83210e68"
	oldB415Staff       = "3fea3d1a-067b-4eae-8433-f7e1e7b724a3"
	oldB415Transport   = "0caae56e-5995-4f1b-b735-80580a372707"
	oldB415Bear        = "Creature — Bear"
)

func TestOldB415AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB415DimirCharm:  "Dimir Charm",
		oldB415Grisly:      "Grisly Spectacle",
		oldB415RustScarab:  "Rust Scarab",
		oldB415Drake:       "Mindeye Drake",
		oldB415Realmwright: "Realmwright",
		oldB415Ragebeast:   "Gruul Ragebeast",
		oldB415Tracer:      "Spire Tracer",
		oldB415Biovis:      "Biovisionary",
		oldB415Rogue:       "Deathcult Rogue",
		oldB415RiotGear:    "Riot Gear",
		oldB415Staff:       "Skyblinder Staff",
		oldB415Transport:   "Armored Transport",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// oldB415Attack has seat 0's attacker attack seat 1 and stops in the declare
// blockers step; it returns the error from seat 1's creature blocking it.
func oldB415Block(t *testing.T, g *game.Game, attacker, blocker uuid.UUID) error {
	t.Helper()
	opp := g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	return g.DeclareBlocker(blocker, attacker)
}

// --- Dimir Charm ------------------------------------------------------------------

func oldB415Charm(t *testing.T, g *game.Game, who *game.Player, mode int, target game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: "Dimir Charm", TypeLine: "Instant", OracleID: oldB415DimirCharm, Owner: who.ID, Controller: who.ID})
	return id, g.CastSpell(who.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: []game.TargetRef{target}})
}

func TestOldB415DimirCharmCountersASorceryButNotACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	creature := effectstest.CastCatalogSpell(t, g, "Their Bear", oldB415Bear, "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if _, err := oldB415Charm(t, g, them, 0, game.TargetRef{Kind: game.TargetCard, ID: creature}); err == nil {
		t.Fatal("Dimir Charm countered a creature spell")
	}
}

func TestOldB415DimirCharmCountersTheSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	sorcery := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if _, err := oldB415Charm(t, g, them, 0, game.TargetRef{Kind: game.TargetCard, ID: sorcery}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(sorcery) || !me.Graveyard.Contains(sorcery) {
		t.Error("the sorcery was not countered")
	}
}

func TestOldB415DimirCharmDestroysACreatureWithPowerTwoOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB415Bear, 2, 2)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB415Bear, 3, 3)
	effectstest.ToMainForCost(t, g)
	if _, err := oldB415Charm(t, g, me, 1, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("destroyed a creature with power 3")
	}
	if _, err := oldB415Charm(t, g, me, 1, game.TargetRef{Kind: game.TargetCard, ID: small}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) || !g.Battlefield.Contains(big) {
		t.Error("the wrong creature was destroyed")
	}
}

func TestOldB415DimirCharmKeepsOneOfTheTopThreeAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(opp, 6)
	c := effectstest.SeedLibraryTop(opp, "Third", "Sorcery", "")
	b := effectstest.SeedLibraryTop(opp, "Second", "Sorcery", "")
	a := effectstest.SeedLibraryTop(opp, "First", "Sorcery", "")
	effectstest.ToMainForCost(t, g)
	if _, err := oldB415Charm(t, g, me, 2, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsPrompt(t, g)
	if pick.Chooser != me.ID || len(pick.ChooseCards) != 3 {
		t.Fatalf("prompt for %v over %d cards, want me over 3", pick.Chooser, len(pick.ChooseCards))
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if top := effectstest.LibraryTopIDs(opp, 1)[0]; top != b {
		t.Errorf("the kept card is not on top of the library")
	}
	if !opp.Graveyard.Contains(a) || !opp.Graveyard.Contains(c) {
		t.Error("the other two cards did not go into their owner's graveyard")
	}
	if me.Graveyard.Contains(a) || me.Graveyard.Contains(c) {
		t.Error("a card went to the caster's graveyard")
	}
}

// --- Grisly Spectacle -------------------------------------------------------------------

func TestOldB415GrislySpectacleDestroysANonartifactCreatureAndMillsItsControllerForItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.Msc18Library(opp, 10)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", oldB415Bear, 4, 4)
	rock := effectstest.B12Creature(g, opp.ID, "Rock", "Artifact Creature — Golem", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Grisly Spectacle", "Instant", oldB415Grisly, effectstest.CardTargets(rock)); err == nil {
		t.Fatal("targeted an artifact creature")
	}
	lib := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Grisly Spectacle", "Instant", oldB415Grisly, effectstest.CardTargets(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the creature survived")
	}
	if got := opp.Library.Size(); got != lib-4 {
		t.Errorf("library %d -> %d, want -4 (the creature's power)", lib, got)
	}
}

// --- Rust Scarab ---------------------------------------------------------------------------

// oldB415ScarabCombat attacks seat 1 with the Scarab, has `blocker` block it
// (when not uuid.Nil), and answers the Scarab's trigger: target `target`,
// then yes or no to the "you may".
func oldB415ScarabCombat(t *testing.T, g *game.Game, scarab, blocker uuid.UUID, yes bool, target uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Pr7bAttack(t, g, opp.ID, scarab)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if blocker != uuid.Nil {
		if err := g.DeclareBlocker(blocker, scarab); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.FinishBlockDeclarations(t, g)
	for i := 0; i < 24; i++ {
		var open *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil {
				open = c
				break
			}
		}
		switch {
		case open == nil && len(g.StackMeta) == 0 && len(g.PendingTriggers) == 0:
			return
		case open == nil:
			if err := g.PassPriority(); err != nil {
				return
			}
		case open.Kind == game.PendingChoicePickTarget:
			if err := g.ResolvePickTarget(open.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("pick target: %v", err)
			}
		case open.Kind == game.PendingChoiceTriggerPrompt:
			if err := g.ResolveTriggerPrompt(open.ID, me.ID, yes); err != nil {
				t.Fatalf("trigger prompt: %v", err)
			}
		default:
			t.Fatalf("unexpected prompt %s", open.Kind)
		}
	}
}

func TestOldB415RustScarabDestroysAnArtifactOrEnchantmentWhenItBecomesBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scarab := effectstest.B12Push(g, me.ID, "Rust Scarab", "Creature — Insect", oldB415RustScarab, 4, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB415Bear, 1, 1)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	oldB415ScarabCombat(t, g, scarab, blocker, true, rock)
	if g.Battlefield.Contains(rock) {
		t.Error("the defending player's artifact survived")
	}
	if !g.Battlefield.Contains(mine) {
		t.Error("my own artifact was destroyed")
	}
}

func TestOldB415RustScarabMayDeclineAndDoesNothingUnblocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scarab := effectstest.B12Push(g, me.ID, "Rust Scarab", "Creature — Insect", oldB415RustScarab, 4, 5)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB415Bear, 1, 1)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	oldB415ScarabCombat(t, g, scarab, blocker, false, rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed though the controller declined")
	}
}

func TestOldB415RustScarabDoesNothingWhenNobodyBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scarab := effectstest.B12Push(g, me.ID, "Rust Scarab", "Creature — Insect", oldB415RustScarab, 4, 5)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	oldB415ScarabCombat(t, g, scarab, uuid.Nil, true, rock)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was destroyed though nothing blocked the Scarab")
	}
}

// --- Mindeye Drake ------------------------------------------------------------------------------

func TestOldB415MindeyeDrakeMillsTheTargetPlayerFiveWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Msc18Library(opp, 10)
	drake := effectstest.B12Push(g, me.ID, "Mindeye Drake", "Creature — Drake", oldB415Drake, 2, 5)
	effectstest.AdvanceToMain(t, g)
	lib := opp.Library.Size()
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(drake) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets2(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-5 {
		t.Errorf("library %d -> %d, want -5", lib, got)
	}
}

// --- Realmwright ----------------------------------------------------------------------------------

func TestOldB415RealmwrightAddsTheChosenTypeToLandsYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedLand(g, me.ID, "Island", "Basic Land — Island", "")
	nonbasic := effectstest.SeedLand(g, me.ID, "Test Cave", "Land — Cave", "")
	theirs := effectstest.SeedLand(g, opp.ID, "Island", "Basic Land — Island", "")
	effectstest.CastCatalogSpell(t, g, "Realmwright", "Creature — Vedalken Wizard", oldB415Realmwright, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectiveSubtypes(t, g, mine); slices.Contains(got, "Forest") {
		t.Errorf("a land is a Forest before a type was chosen: %v", got)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 4) // Forest
	if got := effectstest.EffectiveSubtypes(t, g, mine); !slices.Contains(got, "Forest") || !slices.Contains(got, "Island") {
		t.Errorf("my Island is %v, want Island and Forest", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, nonbasic); !slices.Contains(got, "Forest") || !slices.Contains(got, "Cave") {
		t.Errorf("my nonbasic land is %v, want Cave and Forest", got)
	}
	if got := effectstest.EffectiveSubtypes(t, g, theirs); slices.Contains(got, "Forest") {
		t.Errorf("an opponent's land is %v, want no Forest", got)
	}
}

// --- Gruul Ragebeast ---------------------------------------------------------------------------------

func TestOldB415GruulRagebeastMakesItselfFightWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB415Bear, 2, 2)
	effectstest.RfdCast(t, g, "Gruul Ragebeast", "Creature — Beast", oldB415Ragebeast, 6, 6)
	effectstest.PickCard(t, g, g.Seats[0].ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the opponent's creature survived the fight")
	}
}

func TestOldB415GruulRagebeastMakesAnotherCreatureYouControlFightNotItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.B12Push(g, me.ID, "Gruul Ragebeast", "Creature — Beast", oldB415Ragebeast, 6, 6)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Ogre", oldB415Bear, 3, 3)
	effectstest.ToMainForCost(t, g)
	entering := effectstest.TdCastCreature(t, g, "Goblin", 1, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(entering) {
		t.Error("the entering 1/1 survived a fight with a 3/3, so it did not fight")
	}
	if !g.Battlefield.Contains(theirs) || !g.Battlefield.Contains(beast) {
		t.Error("the Ogre or the Ragebeast left the battlefield")
	}
}

func TestOldB415GruulRagebeastCannotTargetItsOwnersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB415Bear, 2, 2)
	effectstest.RfdCast(t, g, "Gruul Ragebeast", "Creature — Beast", oldB415Ragebeast, 6, 6)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil && effectstest.HasID(pick.PickTargetCards, mine) {
		t.Error("my own creature was offered as the fight's target")
	}
}

// --- evasion creatures ---------------------------------------------------------------------------------

func TestOldB415SpireTracerCanBeBlockedOnlyByFlyingOrReach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tracer := effectstest.B12Push(g, me.ID, "Spire Tracer", "Creature — Elf Scout", oldB415Tracer, 1, 1)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", oldB415Bear, 2, 2)
	flier := effectstest.PushDiesCreatureForTest(g, opp.ID, "Flier", "", oldB415Bear, 2, 2)
	reach := effectstest.PushDiesCreatureForTest(g, opp.ID, "Reacher", "", oldB415Bear, 2, 2)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, flier).Keywords = []string{"flying"}
		effectstest.FindBattlefieldCardForTest(g, reach).Keywords = []string{"reach"}
	})
	if err := oldB415Block(t, g, tracer, ground); err == nil {
		t.Fatal("a creature with neither flying nor reach blocked the Tracer")
	}
	if err := g.DeclareBlocker(flier, tracer); err != nil {
		t.Errorf("a flier could not block the Tracer: %v", err)
	}
	if err := g.DeclareBlocker(reach, tracer); err != nil {
		t.Errorf("a reach creature could not block the Tracer: %v", err)
	}
}

func TestOldB415DeathcultRogueCanBeBlockedOnlyByRogues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := effectstest.B12Push(g, me.ID, "Deathcult Rogue", "Creature — Human Rogue", oldB415Rogue, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB415Bear, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Their Rogue", "Creature — Human Rogue", 2, 2)
	if err := oldB415Block(t, g, rogue, bear); err == nil {
		t.Fatal("a non-Rogue blocked the Rogue")
	}
	if err := g.DeclareBlocker(other, rogue); err != nil {
		t.Errorf("a Rogue could not block: %v", err)
	}
}

// --- Biovisionary --------------------------------------------------------------------------------------

func oldB415Biovisionaries(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		effectstest.B12Push(g, owner, "Biovisionary", "Creature — Human Wizard", oldB415Biovis, 2, 3)
	}
}

func TestOldB415BiovisionaryWinsTheGameWithFourAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB415Biovisionaries(g, me.ID, 4)
	effectstest.AdvanceToEndStep(t, g)
	// Four triggers at once: the controller orders them first (CR 603.3b).
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	// The game ends while the stack drains, so passing stops with an error.
	for i := 0; i < 16 && g.Outcome == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if g.Outcome == nil || g.Outcome.Winner != me.ID {
		t.Errorf("outcome %+v, want a win for the controller", g.Outcome)
	}
}

func TestOldB415BiovisionaryDoesNothingWithThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB415Biovisionaries(g, g.Seats[0].ID, 3)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Outcome != nil {
		t.Errorf("outcome %+v with three Biovisionaries, want none", g.Outcome)
	}
}

// --- the Equipment ---------------------------------------------------------------------------------------

func TestOldB415RiotGearGivesPlusOnePlusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB415Bear, 2, 2)
	gear := effectstest.SeedEquipment(g, me.ID, "Riot Gear", oldB415RiotGear)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, gear, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 4 {
		t.Errorf("the equipped Bear is %d/%d, want 3/4", p, tough)
	}
}

func TestOldB415SkyblinderStaffPumpsAndCantBeBlockedByFliers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB415Bear, 2, 2)
	staff := effectstest.SeedEquipment(g, me.ID, "Skyblinder Staff", oldB415Staff)
	flier := effectstest.PushDiesCreatureForTest(g, opp.ID, "Flier", "", oldB415Bear, 2, 2)
	ground := effectstest.B12Creature(g, opp.ID, "Ground", oldB415Bear, 2, 2)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flier).Keywords = []string{"flying"} })
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, staff, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 3 || tough != 2 {
		t.Errorf("the equipped Bear is %d/%d, want 3/2", p, tough)
	}
	if err := oldB415Block(t, g, bear, flier); err == nil {
		t.Fatal("a flier blocked the equipped creature")
	}
	if err := g.DeclareBlocker(ground, bear); err != nil {
		t.Errorf("a ground creature could not block it: %v", err)
	}
}

// --- Armored Transport ---------------------------------------------------------------------------------------

func TestOldB415ArmoredTransportTakesNoCombatDamageFromItsBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	transport := effectstest.B12Push(g, me.ID, "Armored Transport", "Artifact Creature — Construct", oldB415Transport, 2, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB415Bear, 5, 5)
	effectstest.NcdPlayBlocked(t, g, transport, blocker, true)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if !g.Battlefield.Contains(transport) {
		t.Error("the Transport died to the creature blocking it")
	}
}

func TestOldB415ArmoredTransportStillDiesToOtherDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	transport := effectstest.B12Push(g, me.ID, "Armored Transport", "Artifact Creature — Construct", oldB415Transport, 2, 1)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", oldB415Bear, 5, 5)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { _ = g.DealDamageEachKeyedForEffect(attacker, []uuid.UUID{transport}, 3, game.KeyedTail{}) })
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(transport) {
		t.Error("noncombat damage was prevented")
	}
}
