package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b225_test.go — Ixalan, batch 225 (#5532): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB225Keeper    = "14537df5-e2c8-4906-99ac-742fb628fe08"
	oldB225Avatar    = "248d7759-aa73-4430-aa86-d45dc607944b"
	oldB225Firecan   = "6e61ae7b-f2e8-421c-870b-a48c08d07503"
	oldB225Hunters   = "9e6e6458-9e0e-404c-af07-225c4665eaec"
	oldB225Archer    = "aed10a6c-ef93-46a7-a5d7-c9cd2a4fd91a"
	oldB225Armasaur  = "dc0bd3a3-f34e-4d81-b9e8-b172b005d671"
	oldB225DualShot  = "e56056dc-637e-4229-ae2c-dfd06d24086b"
	oldB225SlashTalo = "ef331f13-bdd0-4123-9e61-b2e4b6b6d8b8"
)

func TestOldB225CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB225Keeper:    "Fire Shrine Keeper",
		oldB225Avatar:    "Burning Sun's Avatar",
		oldB225Firecan:   "Elaborate Firecannon",
		oldB225Hunters:   "Sun-Crowned Hunters",
		oldB225Archer:    "Atzocan Archer",
		oldB225Armasaur:  "Steadfast Armasaur",
		oldB225DualShot:  "Dual Shot",
		oldB225SlashTalo: "Slash of Talons",
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

func oldB225Card(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

// oldB225Pick answers chooser's open pick-target prompt (if any) with ref and
// settles the stack.
func oldB225Pick(t *testing.T, g *game.Game, chooser uuid.UUID, ref game.TargetRef) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, chooser) == nil && effectstest.StackFullyEmpty(g) {
			return
		}
	}
}

// oldB225PickEach answers chooser's pick-target prompts, one per ref in the
// order the clauses are asked, then settles the stack.
func oldB225PickEach(t *testing.T, g *game.Game, chooser uuid.UUID, refs ...game.TargetRef) {
	t.Helper()
	for _, ref := range refs {
		p := effectstest.LatestPickTarget(g, chooser)
		if p == nil {
			effectstest.PassPriorityAroundTable(t, g)
			p = effectstest.LatestPickTarget(g, chooser)
		}
		if p == nil {
			t.Fatalf("no pick-target prompt for %v", ref)
		}
		if err := g.ResolvePickTarget(p.ID, chooser, ref); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB225Damage(g *game.Game, id uuid.UUID) int {
	return effectstest.DamageMarkedOn(g, id)
}

// --- Dual Shot ---

func TestOldB225DualShotHitsBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 3)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 3)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 3)
	effectstest.CastCatalogSpell(t, g, "Dual Shot", "Instant", oldB225DualShot, []game.TargetRef{oldB225Card(a), oldB225Card(b)})
	effectstest.PassPriorityAroundTable(t, g)
	if oldB225Damage(g, a) != 1 || oldB225Damage(g, b) != 1 || oldB225Damage(g, c) != 0 {
		t.Errorf("damage a=%d b=%d c=%d, want 1 1 0", oldB225Damage(g, a), oldB225Damage(g, b), oldB225Damage(g, c))
	}
}

func TestOldB225DualShotRefusesThreeTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	var refs []game.TargetRef
	for i := 0; i < 3; i++ {
		refs = append(refs, oldB225Card(effectstest.B12Creature(g, opp.ID, "X", "Creature — Bear", 2, 3)))
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Dual Shot", "Instant", oldB225DualShot, refs); err == nil {
		t.Error("three targets were accepted")
	}
}

// --- Slash of Talons ---

func TestOldB225SlashOfTalonsHitsAnAttackerAndRefusesAnIdleCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	spell := effectstest.B13HandCard(me, "Slash of Talons", "Instant", oldB225SlashTalo)
	effectstest.BrAttack(t, g, attacker)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{oldB225Card(idle)}}); err == nil {
		t.Error("a creature that is neither attacking nor blocking was accepted")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{oldB225Card(attacker)}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB225Damage(g, attacker); got != 2 {
		t.Errorf("the attacker has %d damage, want 2", got)
	}
}

// --- Sun-Crowned Hunters ---

func TestOldB225SunCrownedHuntersHitTheOpponentForThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunters := effectstest.B12Push(g, me.ID, "Sun-Crowned Hunters", "Creature — Dinosaur", oldB225Hunters, 5, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 1, 1)
	life := opp.Life
	effectstest.Cmdr2B08Damage(g, src, hunters, 1)
	oldB225Pick(t, g, me.ID, effectstest.PlayerTarget(opp.ID))
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
}

func TestOldB225SunCrownedHuntersStillHitWhenTheDamageKillsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunters := effectstest.B12Push(g, me.ID, "Sun-Crowned Hunters", "Creature — Dinosaur", oldB225Hunters, 5, 4)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 4, 4)
	life := opp.Life
	effectstest.Cmdr2B08Damage(g, src, hunters, 4)
	oldB225Pick(t, g, me.ID, effectstest.PlayerTarget(opp.ID))
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
}

// --- Burning Sun's Avatar ---

func TestOldB225BurningSunsAvatarHitsAnOpponentAndACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 1, 9)
	life := opp.Life
	effectstest.AvB16CastCreature(t, g, "Burning Sun's Avatar", "Creature — Dinosaur Avatar", oldB225Avatar, 6, 6)
	oldB225PickEach(t, g, me.ID, effectstest.PlayerTarget(opp.ID), oldB225Card(victim))
	if opp.Life != life-3 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-3)
	}
	if got := oldB225Damage(g, victim); got != 3 {
		t.Errorf("the creature has %d damage, want 3", got)
	}
}

// --- Atzocan Archer ---

func TestOldB225AtzocanArcherFightsWhenYouSayYes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 2, 9)
	archer := effectstest.AvB16CastCreature(t, g, "Atzocan Archer", "Creature — Human Archer", oldB225Archer, 1, 4)
	oldB225PickEach(t, g, me.ID, oldB225Card(victim))
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB225Damage(g, victim); got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
	if got := oldB225Damage(g, archer); got != 2 {
		t.Errorf("the Archer has %d damage, want 2", got)
	}
}

func TestOldB225AtzocanArcherDoesNotFightWhenYouSayNo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 2, 9)
	archer := effectstest.AvB16CastCreature(t, g, "Atzocan Archer", "Creature — Human Archer", oldB225Archer, 1, 4)
	oldB225PickEach(t, g, me.ID, oldB225Card(victim))
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB225Damage(g, victim) != 0 || oldB225Damage(g, archer) != 0 {
		t.Error("a fight happened after the player declined")
	}
}

// --- Fire Shrine Keeper ---

func TestOldB225FireShrineKeeperSacrificesAndHitsTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Fire Shrine Keeper", "Creature — Elemental", oldB225Keeper, false)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 5)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Bear", 2, 5)
	c := effectstest.B12Creature(g, opp.ID, "C", "Creature — Bear", 2, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB225Card(a), oldB225Card(b)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if _, ok := effectstest.BattlefieldCard(g, keeper); ok {
		t.Error("the Keeper was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB225Damage(g, a) != 3 || oldB225Damage(g, b) != 3 || oldB225Damage(g, c) != 0 {
		t.Errorf("damage a=%d b=%d c=%d, want 3 3 0", oldB225Damage(g, a), oldB225Damage(g, b), oldB225Damage(g, c))
	}
}

func TestOldB225FireShrineKeeperRefusesAPoolThatCantPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Fire Shrine Keeper", "Creature — Elemental", oldB225Keeper, false)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Bear", 2, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "R", "R")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{oldB225Card(a)}}); err == nil {
		t.Error("activated without the mana")
	}
}

// --- Steadfast Armasaur ---

func TestOldB225SteadfastArmasaurDealsItsToughnessToACreatureItBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saur := effectstest.B12Push(g, opp.ID, "Steadfast Armasaur", "Creature — Dinosaur", oldB225Armasaur, 2, 3)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 9)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 9)
	effectstest.CrtBlock(g, saur, attacker, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, saur, 0, effectstest.LtCardTarget(bystander), "W", "W"); err == nil {
		t.Fatal("a creature not in combat with the Armasaur was accepted")
	}
	g.PlayerByIDForEffect(opp.ID).ManaPool = nil
	if err := effectstest.CrtActivateRaw(t, g, opp.ID, saur, 0, effectstest.LtCardTarget(attacker), "W", "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB225Damage(g, attacker); got != 3 {
		t.Errorf("the attacker has %d damage, want 3 (the Armasaur's toughness)", got)
	}
	if got := oldB225Damage(g, bystander); got != 0 {
		t.Errorf("the bystander has %d damage, want 0", got)
	}
}

func TestOldB225SteadfastArmasaurHitsABlockerOfItsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saur := effectstest.B12Push(g, me.ID, "Steadfast Armasaur", "Creature — Dinosaur", oldB225Armasaur, 2, 3)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 1, 6)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == saur {
				g.Battlefield.Cards[i].AttackingTarget = opp.ID
			}
		}
	})
	effectstest.CrtBlock(g, blocker, saur, opp.ID)
	if err := effectstest.CrtActivateRaw(t, g, me.ID, saur, 0, effectstest.LtCardTarget(blocker), "W", "W"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := oldB225Damage(g, blocker); got != 3 {
		t.Errorf("the blocker has %d damage, want 3 (the Armasaur's toughness)", got)
	}
}

// --- Elaborate Firecannon ---

func TestOldB225ElaborateFirecannonShootsAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cannon := effectstest.PushCatalogPermanent(g, me.ID, "Elaborate Firecannon", "Artifact", oldB225Firecan, false)
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, cannon, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	if !effectstest.IsTapped(g, cannon) {
		t.Error("the Firecannon did not tap")
	}
}

func TestOldB225ElaborateFirecannonDoesNotUntapInTheUntapStep(t *testing.T) {
	spec, _ := effects.Lookup(oldB225Firecan)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Fatalf("the Firecannon declares %d untap restrictions, want 1", len(spec.UntapStepRestrictions))
	}
}

func TestOldB225ElaborateFirecannonUntapsOnlyIfYouDiscard(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		cannon := effectstest.PushCatalogPermanent(g, me.ID, "Elaborate Firecannon", "Artifact", oldB225Firecan, false)
		g.WithWriteLock(func() { _ = g.TapTargetForEffect(cannon) })
		spare := effectstest.PushHandCard(g, me)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventBeginUpkeep, Actor: me.ID, Step: game.StepUpkeep})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if discard {
			effectstest.AnswerDiscard(t, g, me.ID, spare)
		} else {
			effectstest.AnswerDiscard(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := !effectstest.IsTapped(g, cannon); got != discard {
			t.Errorf("discard=%v: untapped=%v", discard, got)
		}
		if got := me.Graveyard.Contains(spare); got != discard {
			t.Errorf("discard=%v: the card is in the graveyard=%v", discard, got)
		}
	}
}
