package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b691_test.go — Mercadian Masques removal, S82 OLD batch 691 (#5619),
// played through the real catalog entries.

const (
	o691SilentAssassin  = "05b07f1f-42e1-41fb-8858-5e4a311c4ed9"
	o691Bailiff         = "1f861517-ac53-4218-b62e-35e971098b6d"
	o691NotoriousAssass = "455138dd-4f76-4b40-9715-f1b1d61f8f23"
	o691ForcedMarch     = "457dc141-bc72-42ad-b482-c78183462e34"
	o691CrookedScales   = "49c47768-fcc7-4091-b06d-d8a2bd7d1cc9"
	o691WordBlasting    = "8b391f24-992f-4cc0-ac4d-31456983fc03"
	o691HonorFallen     = "db557d5e-eecd-4891-90fa-c5bcb8532015"
	o691PuppetsVerdict  = "e07a7c75-b7b7-4c2d-bd7d-5f056cde871d"
	o691BlasterMage     = "e5b35b54-77d9-4cb0-86a1-efc8e49aea09"
	o691HammerMage      = "f642a43c-f8cc-44ee-ac6a-5d4b3fbc49af"
	o691VenomousBreath  = "fc73bd1b-c713-4825-9f2d-f6ae95dddf09"
	o691CausticWasps    = "fea997a4-3ed1-4f83-b16a-56939f4b2211"
)

func TestS82OldB691CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		o691SilentAssassin: "Silent Assassin", o691Bailiff: "Saprazzan Bailiff", o691NotoriousAssass: "Notorious Assassin",
		o691ForcedMarch: "Forced March", o691CrookedScales: "Crooked Scales", o691WordBlasting: "Word of Blasting",
		o691HonorFallen: "Honor the Fallen", o691PuppetsVerdict: "Puppet's Verdict", o691BlasterMage: "Blaster Mage",
		o691HammerMage: "Hammer Mage", o691VenomousBreath: "Venomous Breath", o691CausticWasps: "Caustic Wasps",
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

func o691Push(g *game.Game, owner uuid.UUID, name, typeLine, cost string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func o691Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o691Alive(g *game.Game, id uuid.UUID) bool {
	return effectstest.FindBattlefieldCardForTest(g, id) != nil
}

// o691Combat has `attacker` (seat 0's) attack seat 1, optionally blocked by
// `blocker`, and stops in the declare blockers step with priority.
func o691Combat(t *testing.T, g *game.Game, attacker, blocker uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatal(err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if blocker != uuid.Nil {
		if err := g.DeclareBlocker(blocker, attacker); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.LockInBlocks(t, g)
}

func TestS82OldB691SilentAssassinDestroysABlockerAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.B12Push(g, me.ID, "Silent Assassin", "Creature — Human Mercenary Assassin", o691SilentAssassin, 2, 1)
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 0, 9)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 0, 9)
	idle := effectstest.PushVanillaCreature(g, opp.ID, "Idle", 0, 9)
	o691Combat(t, g, attacker, blocker)
	effectstest.FloatForTest(g, me, "CCCB")
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Strict: true, Targets: o691Target(idle)}); err == nil {
		t.Fatal("targeted a creature that is not blocking")
	}
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Strict: true, Targets: o691Target(blocker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !o691Alive(g, blocker) {
		t.Fatal("the blocker died at once, not at end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, blocker) {
		t.Error("the blocker survived end of combat")
	}
	if !o691Alive(g, idle) || !o691Alive(g, attacker) {
		t.Error("an unrelated creature died")
	}
}

func TestS82OldB691SaprazzanBailiffExilesOnEnteringAndReturnsOnLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	exiledArtifact := effectstest.PushGraveyardCardWithTypeLineID(me, "Old Relic", "Artifact")
	exiledEnchantment := effectstest.PushGraveyardCardWithTypeLineID(opp, "Old Aura", "Enchantment — Aura")
	creature := effectstest.PushGraveyardCardWithTypeLineID(opp, "Old Bear", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	bailiff := effectstest.HandCardOf(g, me, "Saprazzan Bailiff", "Creature — Merfolk", "", o691Bailiff)
	if err := g.CastSpell(me.ID, bailiff, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(exiledArtifact) || opp.Graveyard.Contains(exiledEnchantment) {
		t.Fatal("an artifact or enchantment card is still in a graveyard")
	}
	if !opp.Graveyard.Contains(creature) {
		t.Error("the creature card was exiled too")
	}
	later := effectstest.PushGraveyardCardWithTypeLineID(opp, "Newer Relic", "Artifact Creature — Golem")
	effectstest.KillCreature(t, g, me.ID, bailiff)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(later) {
		t.Error("the artifact in the graveyard did not return to its owner's hand")
	}
	if me.Hand.Contains(exiledArtifact) || opp.Hand.Contains(exiledEnchantment) {
		t.Error("an exiled card came back")
	}
}

func TestS82OldB691NotoriousAssassinDestroysANonblackCreatureForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	assassin := effectstest.B12Push(g, me.ID, "Notorious Assassin", "Creature — Human Spellshaper Assassin", o691NotoriousAssass, 2, 2)
	black := o691Push(g, opp.ID, "Black Bear", "Creature — Bear", "{1}{B}", 2, 2)
	green := o691Push(g, opp.ID, "Green Bear", "Creature — Bear", "{1}{G}", 2, 2)
	discard := effectstest.HandCardOf(g, me, "Spare", "Creature — Bear", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCB")
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}, Targets: o691Target(black)}); err == nil {
		t.Fatal("targeted a black creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, assassin, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}, Targets: o691Target(green)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, green) {
		t.Error("the nonblack creature survived")
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard was not paid")
	}
}

func TestS82OldB691ForcedMarchDestroysCreaturesWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	one := o691Push(g, me.ID, "One", "Creature — Bear", "{1}", 1, 1)
	two := o691Push(g, opp.ID, "Two", "Creature — Bear", "{1}{G}", 2, 2)
	three := o691Push(g, opp.ID, "Three", "Creature — Bear", "{2}{G}", 3, 3)
	token := o691Push(g, opp.ID, "Token", "Token Creature — Bear", "", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCBBB")
	id := effectstest.HandCardOf(g, me, "Forced March", "Sorcery", "{X}{B}{B}{B}", o691ForcedMarch)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for who, alive := range map[uuid.UUID]bool{one: false, two: false, token: false, three: true} {
		if o691Alive(g, who) != alive {
			t.Errorf("creature %v alive = %v, want %v", who, o691Alive(g, who), alive)
		}
	}
}

func TestS82OldB691HonorTheFallenExilesEveryCreatureCardAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.PushGraveyardCardWithTypeLineID(me, "A", "Creature — Bear")
	b := effectstest.PushGraveyardCardWithTypeLineID(opp, "B", "Artifact Creature — Golem")
	c := effectstest.PushGraveyardCardWithTypeLineID(opp, "C", "Sorcery")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CW")
	id := effectstest.HandCardOf(g, me, "Honor the Fallen", "Instant", "{1}{W}", o691HonorFallen)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Graveyard.Contains(a) || opp.Graveyard.Contains(b) {
		t.Error("a creature card is still in a graveyard")
	}
	if !opp.Graveyard.Contains(c) {
		t.Error("the sorcery card was exiled")
	}
	if me.Life != 42 {
		t.Errorf("life = %d, want 42", me.Life)
	}
}

// o691Flip casts or activates through `start`, calls heads and reports whether
// the flip was won.
func o691CallFlip(t *testing.T, g *game.Game, me *game.Player) bool {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
	if ask == nil {
		t.Fatal("no coin call was asked")
	}
	flips := len(effectstest.RandomEvents(g, game.EventFlipCoin))
	if err := g.ResolveCoinCall(ask.ID, me.ID, "heads"); err != nil {
		t.Fatalf("ResolveCoinCall: %v", err)
	}
	evs := effectstest.RandomEvents(g, game.EventFlipCoin)
	if len(evs) != flips+1 {
		t.Fatalf("%d coins were flipped, want 1", len(evs)-flips)
	}
	return evs[len(evs)-1].Won
}

func TestS82OldB691PuppetsVerdictDestroysByPowerOnEitherFlip(t *testing.T) {
	won, lost := 0, 0
	for table := 0; table < 60 && (won == 0 || lost == 0); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		small := effectstest.PushVanillaCreature(g, me.ID, "Small", 2, 9)
		big := effectstest.PushVanillaCreature(g, opp.ID, "Big", 3, 9)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, "CRR")
		id := effectstest.HandCardOf(g, me, "Puppet's Verdict", "Instant", "{1}{R}{R}", o691PuppetsVerdict)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		w := o691CallFlip(t, g, me)
		effectstest.PassPriorityAroundTable(t, g)
		// 9 toughness: a destroyed creature must be gone only if the flip chose it.
		if w {
			won++
			if o691Alive(g, small) || !o691Alive(g, big) {
				t.Fatal("a won flip must destroy power 2 or less only")
			}
		} else {
			lost++
			if !o691Alive(g, small) || o691Alive(g, big) {
				t.Fatal("a lost flip must destroy power 3 or greater only")
			}
		}
	}
	if won == 0 || lost == 0 {
		t.Fatalf("saw %d won and %d lost flips; want both", won, lost)
	}
}

func TestS82OldB691CrookedScalesWinDestroysTheirCreatureAndLossCanBePaidOff(t *testing.T) {
	sawWin, sawLossDecline, sawLossPay := false, false, false
	for table := 0; table < 80 && !(sawWin && sawLossDecline && sawLossPay); table++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		scales := effectstest.PushCatalogPermanent(g, me.ID, "Crooked Scales", "Artifact", o691CrookedScales, false)
		theirs := effectstest.PushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
		mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.FloatForTest(g, me, "CCCC")
		targets := []game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: mine}}
		if err := g.ActivateCatalogAbility(me.ID, scales, 0, game.ActivateAbilityParams{Strict: true, Targets: targets}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		if o691CallFlip(t, g, me) {
			effectstest.PassPriorityAroundTable(t, g)
			sawWin = true
			if o691Alive(g, theirs) || !o691Alive(g, mine) {
				t.Fatal("a won flip must destroy only the opponent's creature")
			}
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("a lost flip did not offer to pay {3}")
		}
		if !sawLossDecline {
			effectstest.AnswerPayUnless(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			sawLossDecline = true
			if o691Alive(g, mine) || !o691Alive(g, theirs) {
				t.Fatal("declining the payment must destroy only their own creature")
			}
			continue
		}
		effectstest.FloatForTest(g, me, "CCC")
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		sawLossPay = true
		if !o691Alive(g, mine) {
			t.Fatal("paying {3} must spare their own creature")
		}
		if effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall) == nil {
			t.Fatal("paying {3} must repeat the process with another flip")
		}
	}
	if !sawWin || !sawLossDecline || !sawLossPay {
		t.Fatalf("outcomes seen: win %v, lost+decline %v, lost+pay %v", sawWin, sawLossDecline, sawLossPay)
	}
}

func TestS82OldB691WordOfBlastingDestroysAWallAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := o691Push(g, opp.ID, "Test Wall", "Creature — Wall", "{2}{W}", 0, 5)
	bear := o691Push(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CR")
	id := effectstest.HandCardOf(g, me, "Word of Blasting", "Instant", "{1}{R}", o691WordBlasting)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: o691Target(bear)}); err == nil {
		t.Fatal("targeted a non-Wall")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: o691Target(wall)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, wall) {
		t.Error("the Wall survived")
	}
	if opp.Life != 37 {
		t.Errorf("the Wall's controller's life = %d, want 37", opp.Life)
	}
}

func TestS82OldB691BlasterMageDestroysAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Blaster Mage", "Creature — Human Spellshaper", o691BlasterMage, 2, 2)
	wall := o691Push(g, opp.ID, "Test Wall", "Creature — Wall", "{1}", 0, 5)
	bear := o691Push(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}", 2, 2)
	discard := effectstest.HandCardOf(g, me, "Spare", "Creature — Bear", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}, Targets: o691Target(bear)}); err == nil {
		t.Fatal("targeted a non-Wall")
	}
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}, Targets: o691Target(wall)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, wall) {
		t.Error("the Wall survived")
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discard was not paid")
	}
}

func TestS82OldB691HammerMageDestroysArtifactsWithManaValueXOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mage := effectstest.B12Push(g, me.ID, "Hammer Mage", "Creature — Human Spellshaper", o691HammerMage, 1, 1)
	cheap := o691Push(g, opp.ID, "Cheap", "Artifact", "{1}", 0, 0)
	dear := o691Push(g, opp.ID, "Dear", "Artifact", "{3}", 0, 0)
	token := o691Push(g, me.ID, "Treasure", "Token Artifact — Treasure", "", 0, 0)
	discard := effectstest.HandCardOf(g, me, "Spare", "Creature — Bear", "", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CR")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true, XValue: 1, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for who, alive := range map[uuid.UUID]bool{cheap: false, token: false, dear: true} {
		if o691Alive(g, who) != alive {
			t.Errorf("artifact %v alive = %v, want %v", who, o691Alive(g, who), alive)
		}
	}
}

func TestS82OldB691VenomousBreathDestroysTheCreaturesThatFoughtTheTargetAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 0, 9)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 0, 9)
	bystander := effectstest.PushVanillaCreature(g, opp.ID, "Bystander", 0, 9)
	o691Combat(t, g, attacker, blocker)
	effectstest.FloatForTest(g, me, "CCCG")
	id := effectstest.HandCardOf(g, me, "Venomous Breath", "Instant", "{3}{G}", o691VenomousBreath)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: o691Target(attacker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !o691Alive(g, blocker) {
		t.Fatal("the blocker died before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, blocker) {
		t.Error("the creature that blocked the target survived end of combat")
	}
	if !o691Alive(g, attacker) || !o691Alive(g, bystander) {
		t.Error("the target or a creature that never fought it was destroyed")
	}
}

func TestS82OldB691CausticWaspsMayDestroyAnArtifactOfThePlayerItHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wasps := effectstest.PushCatalogPermanent(g, me.ID, "Caustic Wasps", "Creature — Insect", o691CausticWasps, false)
	relic := o691Push(g, opp.ID, "Relic", "Artifact", "{2}", 0, 0)
	mine := o691Push(g, me.ID, "My Relic", "Artifact", "{2}", 0, 0)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == wasps {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 1, 1
			}
		}
	})
	o691Combat(t, g, wasps, uuid.Nil)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.InitSettle(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, relic)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if o691Alive(g, relic) {
		t.Error("the damaged player's artifact survived")
	}
	if !o691Alive(g, mine) {
		t.Error("the controller's own artifact was destroyed")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wasps, "flying") {
		t.Error("no flying")
	}
}
