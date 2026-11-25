package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b263_cards_test.go — Kaladesh, batch 263 (#5539). Each test plays
// the real catalog entry through the engine and asserts what the printed text
// says.

const (
	oldB263TakeDown    = "0925e112-c089-4aa4-9878-6d6fb0a263c8"
	oldB263Puzzleknot  = "0de18ba2-84a7-4e07-84b2-771761d77aaf"
	oldB263Sabotage    = "28da6443-4acc-4f1c-a54f-9153286f18c4"
	oldB263Firebrand   = "d4b0a6e6-937d-428a-a469-bc09c30ac276"
	oldB263Infiltrator = "f46e9fa0-2ed7-4a94-8ad4-031e87a6da45"
	oldB263Warden      = "388b7b0f-b26d-4de9-bfa6-8c3cbcc2e284"
	oldB263Designs     = "870185fc-7b05-47d3-95ff-97fbf12796d4"
	oldB263Gremlin     = "b88d67c0-ef12-453c-8e58-6642650dffcc"
	oldB263Conclusion  = "d73adc2e-c458-40a9-b10b-7d2133778717"
	oldB263BearType    = "Creature — Bear"
	oldB263FlyerType   = "Creature — Bird"
)

func TestOldB263AllRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB263TakeDown:    "Take Down",
		oldB263Puzzleknot:  "Fireforger's Puzzleknot",
		oldB263Sabotage:    "Incendiary Sabotage",
		oldB263Firebrand:   "Renegade Firebrand",
		oldB263Infiltrator: "Spireside Infiltrator",
		oldB263Warden:      "Fairgrounds Warden",
		oldB263Designs:     "Underhanded Designs",
		oldB263Gremlin:     "Ruinous Gremlin",
		oldB263Conclusion:  "Tidy Conclusion",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

// oldB263Flyer is a battlefield creature with flying and the given body.
func oldB263Flyer(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: oldB263FlyerType, Keywords: []string{"flying"},
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

func TestOldB263TakeDownFourDamageToOneCreatureWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flyer := oldB263Flyer(g, opp.ID, "Their Bird", 2, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 5)
	other := oldB263Flyer(g, me.ID, "My Bird", 2, 5)
	// A creature without flying is not a legal target.
	if _, err := effectstest.CastModalErr(t, g, "Take Down", "Sorcery", oldB263TakeDown, []int{0}, effectstest.CardRefs(bear)); err == nil {
		t.Fatal("Take Down targeted a creature without flying")
	}
	effectstest.CastModal(t, g, "Take Down", "Sorcery", oldB263TakeDown, []int{0}, effectstest.CardRefs(flyer))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, flyer); got != 4 {
		t.Errorf("the targeted flyer has %d damage, want 4", got)
	}
	if got := effectstest.DamageOn(t, g, other); got != 0 {
		t.Errorf("an untargeted flyer has %d damage, want 0", got)
	}
	if got := effectstest.DamageOn(t, g, bear); got != 0 {
		t.Errorf("a creature without flying has %d damage, want 0", got)
	}
}

func TestOldB263TakeDownOneDamageToEachCreatureWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB263Flyer(g, opp.ID, "Their Bird", 2, 5)
	mine := oldB263Flyer(g, me.ID, "My Bird", 2, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 5)
	effectstest.CastModal(t, g, "Take Down", "Sorcery", oldB263TakeDown, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"their flyer": theirs, "my flyer": mine} {
		if got := effectstest.DamageOn(t, g, id); got != 1 {
			t.Errorf("%s has %d damage, want 1", name, got)
		}
	}
	if got := effectstest.DamageOn(t, g, bear); got != 0 {
		t.Errorf("a creature without flying has %d damage, want 0", got)
	}
}

func TestOldB263FireforgersPuzzleknotPingsOnEntryAndAgainWhenSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	knot := effectstest.B13HandCard(me, "Fireforger's Puzzleknot", "Artifact", oldB263Puzzleknot)
	if err := g.CastSpell(me.ID, knot, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(ch.ID, ch.Chooser, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Fatalf("the opponent is at %d after the entry trigger, want %d", opp.Life, life-1)
	}
	// {2}{R}, sacrifice: 1 damage to any target; unpayable without the {R}.
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 1, 1)
	if err := effectstest.UsgB2Activate(t, g, me, knot, "CCC", []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err == nil {
		t.Fatal("activated without {R}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, knot, "RCC", []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(knot) || !me.Graveyard.Contains(knot) {
		t.Error("the Puzzleknot is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the 1/1 survived the sacrificed Puzzleknot's damage")
	}
}

func TestOldB263IncendiarySabotageNeedsAnArtifactToSacrificeAndHitsEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB263BearType, 2, 4)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 3)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	id := effectstest.B13HandCard(me, "Incendiary Sabotage", "Instant", oldB263Sabotage)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with no sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{mine}}); err == nil {
		t.Fatal("a creature that is not an artifact paid the sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{SacrificeIDs: []uuid.UUID{rock}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed")
	}
	if got := effectstest.DamageOn(t, g, mine); got != 3 {
		t.Errorf("my creature has %d damage, want 3 (the caster's own is hit too)", got)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("their 2/3 survived 3 damage")
	}
}

func TestOldB263RenegadeFirebrandGetsPumpedAndFirstStrikeWithAChandra(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fire := effectstest.B12Push(g, me.ID, "Renegade Firebrand", "Creature — Human Warrior", oldB263Firebrand, 3, 2)
	if p := effectstest.EffectivePower(t, g, fire); p != 3 || effectstest.HasEffectiveKeyword(t, g, fire, "first strike") {
		t.Fatalf("power %d with no Chandra, want 3 and no first strike", p)
	}
	// A planeswalker that is not a Chandra, and a Chandra an opponent controls,
	// do nothing; a creature merely named Chandra is not a planeswalker.
	effectstest.B12Push(g, me.ID, "Jace", "Legendary Planeswalker — Jace", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Chandra, Test", "Legendary Planeswalker — Chandra", "", 0, 0)
	effectstest.B12Creature(g, me.ID, "Chandra Impostor", "Creature — Human", 1, 1)
	if p := effectstest.EffectivePower(t, g, fire); p != 3 || effectstest.HasEffectiveKeyword(t, g, fire, "first strike") {
		t.Fatalf("power %d without my Chandra planeswalker, want 3 and no first strike", p)
	}
	chandra := effectstest.B12Push(g, me.ID, "Chandra, Test", "Legendary Planeswalker — Chandra", "", 0, 0)
	if p := effectstest.EffectivePower(t, g, fire); p != 4 {
		t.Errorf("power %d with a Chandra, want 4", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, fire, "first strike") {
		t.Error("no first strike with a Chandra")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(chandra) })
	g.RunStateChecksForTest()
	if p := effectstest.EffectivePower(t, g, fire); p != 3 || effectstest.HasEffectiveKeyword(t, g, fire, "first strike") {
		t.Errorf("power %d after the Chandra left, want 3 and no first strike", p)
	}
}

func TestOldB263SpiresideInfiltratorPingsEachOpponentWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spire := effectstest.B12Push(g, me.ID, "Spireside Infiltrator", "Creature — Human Rogue", oldB263Infiltrator, 3, 2)
	effectstest.AdvanceToMain(t, g)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats[1:] {
		before[p.ID] = p.Life
	}
	selfLife := me.Life
	effectstest.TapForTest(t, g, spire)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != before[p.ID]-1 {
			t.Errorf("opponent at %d, want %d", p.Life, before[p.ID]-1)
		}
	}
	if me.Life != selfLife {
		t.Errorf("the controller lost life: %d, want %d", me.Life, selfLife)
	}
	// Already tapped: tapping it again is not a new "becomes tapped".
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if p.Life != before[p.ID]-1 {
			t.Errorf("a second ping landed: opponent at %d", p.Life)
		}
	}
}

func TestOldB263SpiresideInfiltratorPingsWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spire := effectstest.B12Push(g, me.ID, "Spireside Infiltrator", "Creature — Human Rogue", oldB263Infiltrator, 3, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, spire)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("the defender is at %d after the attack tapped it, want %d", opp.Life, life-1)
	}
}

func TestOldB263FairgroundsWardenExilesAnOpposingCreatureUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB263BearType, 2, 2)
	warden := effectstest.CastCatalogSpell(t, g, "Fairgrounds Warden", "Creature — Dwarf Soldier", oldB263Warden, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(mine)); err == nil {
		t.Fatal("the Warden targeted its controller's own creature")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(theirs) || g.Battlefield.Contains(theirs) {
		t.Fatal("the opposing creature was not exiled")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(warden) })
	g.RunStateChecksForTest()
	if n := effectstest.NamedOnBattlefield(g, "Their Bear", opp.ID); n != 1 {
		t.Errorf("after the Warden left, %d of the exiled creature under its owner's control, want 1", n)
	}
}

func TestOldB263UnderhandedDesignsDrainsWhenYouPayForAnArtifact(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Underhanded Designs", "Enchantment", oldB263Designs, 0, 0)
		effectstest.AdvanceToMain(t, g)
		myLife, oppLife := me.Life, opp.Life
		effectstest.FillPool(me, 1)
		// An opponent's artifact does nothing.
		effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Rock", TypeLine: "Artifact"})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("asked to pay for an opponent's artifact")
		}
		g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TreasureToken(), 1) })
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("my artifact did not offer the {1} payment")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		wantMine, wantTheirs := myLife, oppLife
		if pay {
			wantMine, wantTheirs = myLife+1, oppLife-1
		}
		if me.Life != wantMine || opp.Life != wantTheirs {
			t.Errorf("pay=%v: life %d/%d, want %d/%d", pay, me.Life, opp.Life, wantMine, wantTheirs)
		}
	}
}

func TestOldB263UnderhandedDesignsDestroysOnlyWithTwoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	designs := effectstest.B12Push(g, me.ID, "Underhanded Designs", "Enchantment", oldB263Designs, 0, 0)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	if err := effectstest.UsgB2Activate(t, g, me, designs, "BCC", target); err == nil {
		t.Fatal("activated with one artifact")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	// Two artifacts, but one belongs to an opponent: still one of mine.
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	if err := effectstest.UsgB2Activate(t, g, me, designs, "BCC", target); err == nil {
		t.Fatal("activated with only one artifact of mine")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	if err := effectstest.UsgB2Activate(t, g, me, designs, "CC", target); err == nil {
		t.Fatal("activated without {B}")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, designs, "BC", target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(designs) {
		t.Error("the enchantment is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
}

func TestOldB263RuinousGremlinSacrificesToDestroyAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	gremlin := effectstest.B12Push(g, me.ID, "Ruinous Gremlin", "Creature — Gremlin", oldB263Gremlin, 1, 1)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 2)
	at := func(id uuid.UUID) []game.TargetRef { return []game.TargetRef{{Kind: game.TargetCard, ID: id}} }
	if err := effectstest.UsgB2Activate(t, g, me, gremlin, "RC", at(rock)); err == nil {
		t.Fatal("activated with only two mana")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, gremlin, "RCC", at(bear)); err == nil {
		t.Fatal("a creature that is not an artifact was targeted")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := effectstest.UsgB2Activate(t, g, me, gremlin, "RCC", at(rock)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(gremlin) || !me.Graveyard.Contains(gremlin) {
		t.Error("the Gremlin is sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("the creature was never a target")
	}
}

func TestOldB263TidyConclusionDestroysAndGainsLifePerArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Two artifacts of mine, one artifact creature of mine and one of theirs.
	effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	golem := effectstest.B12Creature(g, me.ID, "My Golem", "Artifact Creature — Golem", 2, 2)
	life := me.Life
	// Destroying my own artifact creature leaves two artifacts to count.
	effectstest.CastCatalogSpell(t, g, "Tidy Conclusion", "Instant", oldB263Conclusion, effectstest.CardRefs(golem))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(golem) {
		t.Error("the creature was not destroyed")
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d (two artifacts left, the dead one and the opponent's not counted)", me.Life, life+2)
	}
}

func TestOldB263TidyConclusionGainsNothingWithNoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB263BearType, 2, 2)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Tidy Conclusion", "Instant", oldB263Conclusion, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not destroyed")
	}
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}
