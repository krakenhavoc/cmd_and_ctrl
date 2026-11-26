package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b317_test.go — S82 OLD batch 317 (#5549, Magic Origins): damage.
// Each test plays the real catalog entry.

const (
	oldB317Hellion     = "01c39670-3677-49df-a4f6-e5062e0909d5"
	oldB317Hixus       = "02bff648-fbad-4d61-ab09-93993a5fda58"
	oldB317Alpha       = "11b12954-dd1b-4c50-8433-15d8c1a15e29"
	oldB317Responder   = "1d84fb81-2de5-4be1-9638-9ae28e3d867a"
	oldB317Rambler     = "97f929ce-3eb3-4040-972f-0b21d5f2f61f"
	oldB317Instincts   = "99d9d928-a426-431a-85ee-cdec86a283cd"
	oldB317Prickleboar = "a023deaa-c393-44f9-9c9c-6bb9aae05c1a"
	oldB317Volley      = "ab855992-34da-42a2-b7c7-dec4db849159"
	oldB317Orbs        = "c420d9c8-dc54-46b8-bd22-39c21019c3d7"
	oldB317Vortex      = "da5ee877-143e-4f53-ad20-1bc527a8fff8"
	oldB317Artificer   = "de2c7724-647c-4ca2-be4d-f0e348a59262"
	oldB317Fury        = "eb16f188-fe12-4a13-9947-d43cd6ce270c"
)

func TestOldB317EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB317Hellion:     "Embermaw Hellion",
		oldB317Hixus:       "Hixus, Prison Warden",
		oldB317Alpha:       "Somberwald Alpha",
		oldB317Responder:   "Mage-Ring Responder",
		oldB317Rambler:     "Volcanic Rambler",
		oldB317Instincts:   "Wild Instincts",
		oldB317Prickleboar: "Prickleboar",
		oldB317Volley:      "Aerial Volley",
		oldB317Orbs:        "Orbs of Warding",
		oldB317Vortex:      "Molten Vortex",
		oldB317Artificer:   "Reclusive Artificer",
		oldB317Fury:        "Chandra's Fury",
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

// --- Embermaw Hellion ---------------------------------------------------------

func TestOldB317HellionAddsOneToAnotherRedSourcesDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hellion := effectstest.B21Push(g, me.ID, "Embermaw Hellion", "Creature — Hellion", oldB317Hellion, 4, 5, "R")
	red := effectstest.B21Push(g, me.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	green := effectstest.B21Push(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	theirRed := effectstest.B21Push(g, opp.ID, "Their Red Bear", "Creature — Bear", "", 2, 2, "R")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 30)
	deal := func(src, target uuid.UUID, n int) {
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, target, n) })
	}
	before := opp.Life
	deal(red, opp.ID, 2)
	if got := before - opp.Life; got != 3 {
		t.Errorf("another red source dealt 2 and it became %d, want 3", got)
	}
	before = opp.Life
	deal(hellion, opp.ID, 2)
	if got := before - opp.Life; got != 2 {
		t.Errorf("the Hellion's own damage became %d, want 2: only ANOTHER source", got)
	}
	before = opp.Life
	deal(green, opp.ID, 2)
	if got := before - opp.Life; got != 2 {
		t.Errorf("a green source's damage became %d, want 2", got)
	}
	before = opp.Life
	deal(theirRed, opp.ID, 2)
	if got := before - opp.Life; got != 2 {
		t.Errorf("an opponent's red source's damage became %d, want 2: only sources you control", got)
	}
	// Damage to your own player counts too: "a permanent or player".
	mine := me.Life
	deal(red, me.ID, 1)
	if got := mine - me.Life; got != 2 {
		t.Errorf("damage to its own controller became %d, want 2", got)
	}
	// And to a permanent, combat or not.
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(red, wall, 1) })
	if got := effectstest.DamageMarkedOn(g, wall); got != 2 {
		t.Errorf("1 damage to a creature became %d, want 2", got)
	}
}

// --- Hixus, Prison Warden -----------------------------------------------------

func oldB317Raid(t *testing.T, enteredThisTurn bool) (g *game.Game, me, them *game.Player, hixus, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, them = g.Seats[0], g.Seats[1]
	hixus = effectstest.PushCatalogPermanent(g, them.ID, "Hixus, Prison Warden", "Legendary Creature — Human Soldier", oldB317Hixus, false)
	g.WithWriteLock(func() {
		if g.TurnTally.Entered == nil {
			g.TurnTally.Entered = map[uuid.UUID]int{}
		}
		if enteredThisTurn {
			g.TurnTally.Entered[hixus] = 1
		} else {
			delete(g.TurnTally.Entered, hixus)
		}
	})
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, them.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, them, hixus, bear
}

func TestOldB317HixusExilesACreatureThatHitYouUntilItLeaves(t *testing.T) {
	g, me, them, hixus, bear := oldB317Raid(t, true)
	if !g.Exile.Contains(bear) || g.Battlefield.Contains(bear) {
		t.Fatal("the creature that dealt combat damage to Hixus's controller was not exiled")
	}
	if them.Life >= 40 {
		t.Errorf("life %d: the combat damage was still dealt", them.Life)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(hixus) })
	g.RunStateChecksForTest()
	if effectstest.NamedOnBattlefield(g, "Bear", me.ID) != 1 {
		t.Error("the exiled creature did not return under its owner's control when Hixus left")
	}
}

func TestOldB317HixusDoesNothingIfItDidNotEnterThisTurn(t *testing.T) {
	g, _, _, _, bear := oldB317Raid(t, false)
	if g.Exile.Contains(bear) || !g.Battlefield.Contains(bear) {
		t.Error("exiled a creature although Hixus did not enter this turn (CR 603.4)")
	}
}

func TestOldB317HixusHasFlash(t *testing.T) {
	spec, _ := effects.Lookup(oldB317Hixus)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

// --- Somberwald Alpha ---------------------------------------------------------

func TestOldB317AlphaPumpsEachBlockedCreatureOfYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Somberwald Alpha", "Creature — Wolf", oldB317Alpha, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("blocked bear is %d/%d, want 3/3", p, tough)
	}
}

func TestOldB317AlphaIgnoresAnOpposingBlockedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, them.ID, "Somberwald Alpha", "Creature — Wolf", oldB317Alpha, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	wall := effectstest.PushVanillaCreature(g, them.ID, "Wall", 0, 9)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("an opponent's blocked bear is %d/%d, want 2/2", p, tough)
	}
}

func TestOldB317AlphaGrantsTrampleToACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them := g.Seats[0], g.Seats[1]
	alpha := effectstest.PushCatalogPermanent(g, me.ID, "Somberwald Alpha", "Creature — Wolf", oldB317Alpha, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, them.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, alpha, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.B06AddMana(me, "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, alpha, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err == nil {
		t.Fatal("targeted a creature you don't control")
	}
	effectstest.B16Activate(t, g, me.ID, alpha, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "trample") {
		t.Error("the bear did not gain trample")
	}
}

// --- Mage-Ring Responder ------------------------------------------------------

func TestOldB317ResponderDoesNotUntapAndUntapsForSeven(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	golem := effectstest.PushCatalogPermanent(g, me.ID, "Mage-Ring Responder", "Artifact Creature — Golem", oldB317Responder, false)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, golem).Tapped = true })
	// {7}: Untap this creature.
	effectstest.FillPool(me, 6)
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("untapped for only six mana")
	}
	effectstest.FillPool(me, 7)
	effectstest.B16Activate(t, g, me.ID, golem, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, golem); tapped {
		t.Error("still tapped after {7}")
	}
	spec, _ := effects.Lookup(oldB317Responder)
	if len(spec.UntapStepRestrictions) != 1 {
		t.Errorf("%d untap-step restrictions, want 1", len(spec.UntapStepRestrictions))
	}
}

func TestOldB317ResponderAttackDealsSevenToADefendingPlayersCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, them, third := g.Seats[0], g.Seats[1], g.Seats[2]
	golem := effectstest.PushCatalogPermanent(g, me.ID, "Mage-Ring Responder", "Artifact Creature — Golem", oldB317Responder, false)
	big := effectstest.B12Creature(g, them.ID, "Big", "Creature — Wall", 0, 20)
	other := effectstest.B12Creature(g, third.ID, "Other", "Creature — Wall", 0, 20)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(golem, them.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the attack trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(other)); err == nil {
		t.Fatal("targeted a creature a non-defending player controls")
	}
	effectstest.PickCard(t, g, me.ID, big)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, big); got != 7 {
		t.Errorf("the target has %d damage, want 7", got)
	}
	if got := effectstest.DamageMarkedOn(g, other); got != 0 {
		t.Errorf("a creature of a player who was not attacked took %d", got)
	}
}

// --- Volcanic Rambler ---------------------------------------------------------

func TestOldB317RamblerPingsAPlayerOrPlaneswalkerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rambler := effectstest.PushCatalogPermanent(g, me.ID, "Volcanic Rambler", "Creature — Elemental", oldB317Rambler, true)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "R", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, rambler, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("targeted a creature")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, rambler, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("the ping dealt %d, want 1", got)
	}
	// No tap: it can be activated while summoning sick, and again.
	effectstest.B06AddMana(me, "R", "C", "C")
	walker := effectstest.PushCatalogWalkerWithLine(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", 5)
	effectstest.B16Activate(t, g, me.ID, rambler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardCopy(t, g, walker).Counters[game.CounterLoyalty]; got != 4 {
		t.Errorf("planeswalker loyalty %d, want 4", got)
	}
}

// --- Wild Instincts -----------------------------------------------------------

func TestOldB317WildInstinctsPumpsThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 9)
	effectstest.D1658Cast(t, g, "Wild Instincts", "Sorcery", oldB317Instincts, game.CastSpellParams{Targets: effectstest.CardRefs(mine, theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	// The pump comes first: the 2/2 fights as a 4/4 and survives the Ogre's 3.
	if got := effectstest.DamageMarkedOn(g, theirs); got != 4 {
		t.Errorf("the opposing creature took %d, want 4 (the pumped power)", got)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 4 || tough != 4 {
		t.Errorf("my creature is %d/%d, want 4/4 until end of turn", p, tough)
	}
	if got := effectstest.DamageMarkedOn(g, mine); got != 3 {
		t.Errorf("my creature took %d from the fight, want 3", got)
	}
}

func TestOldB317WildInstinctsRefusesIllegalTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	mine2 := effectstest.B12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	if err := effectstest.D1658CastErr(t, g, "Wild Instincts", "Sorcery", oldB317Instincts, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(mine, mine2)}); err == nil {
		t.Error("the second target may not be a creature you control")
	}
	if err := effectstest.D1658CastErr(t, g, "Wild Instincts", "Sorcery", oldB317Instincts, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirs, theirs)}); err == nil {
		t.Error("the first target may not be an opponent's creature")
	}
}

// --- Prickleboar --------------------------------------------------------------

func TestOldB317PrickleboarIsBiggerWithFirstStrikeOnlyOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	boar := effectstest.B12Push(g, me.ID, "Prickleboar", "Creature — Boar", oldB317Prickleboar, 3, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, boar); p != 5 || tough != 3 {
		t.Errorf("on its controller's turn it is %d/%d, want 5/3", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, boar, "first strike") {
		t.Error("no first strike on its controller's turn")
	}
	for i := 0; i < 60 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != 1 {
		t.Fatalf("never reached %s's turn", opp.ID)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, boar); p != 3 || tough != 3 {
		t.Errorf("on another player's turn it is %d/%d, want 3/3", p, tough)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, boar, "first strike") {
		t.Error("first strike on another player's turn")
	}
}

// --- Aerial Volley ------------------------------------------------------------

func TestOldB317AerialVolleyDividesThreeAmongCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bird A", "Creature — Bird", 1, 10)
	b := effectstest.B12Creature(g, opp.ID, "Bird B", "Creature — Bird", 1, 10)
	for _, id := range []uuid.UUID{a, b} {
		g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, id).Keywords = []string{"flying"} })
	}
	effectstest.D1658Cast(t, g, "Aerial Volley", "Instant", oldB317Volley, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, a); got != 2 {
		t.Errorf("a: %d damage, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, b); got != 1 {
		t.Errorf("b: %d damage, want 1", got)
	}
}

func TestOldB317AerialVolleyRefusesAGroundCreatureAndABadSplit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bird := effectstest.B12Creature(g, opp.ID, "Bird", "Creature — Bird", 1, 10)
	g.WithWriteLock(func() { effectstest.BattlefieldCardPtr(t, g, bird).Keywords = []string{"flying"} })
	ground := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := effectstest.D1658CastErr(t, g, "Aerial Volley", "Instant", oldB317Volley, game.CastSpellParams{
		Strict: true, Targets: effectstest.CardRefs(ground), Distribution: map[uuid.UUID]int{ground: 3},
	}); err == nil {
		t.Error("targeted a creature without flying")
	}
	if err := effectstest.D1658CastErr(t, g, "Aerial Volley", "Instant", oldB317Volley, game.CastSpellParams{
		Strict: true, Targets: effectstest.CardRefs(bird), Distribution: map[uuid.UUID]int{bird: 2},
	}); err == nil {
		t.Error("divided only 2 of the 3 damage")
	}
	effectstest.D1658Cast(t, g, "Aerial Volley", "Instant", oldB317Volley, game.CastSpellParams{
		Targets: effectstest.CardRefs(bird), Distribution: map[uuid.UUID]int{bird: 3},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bird); got != 3 {
		t.Errorf("a single target took %d, want 3", got)
	}
}

// --- Orbs of Warding ----------------------------------------------------------

func TestOldB317OrbsGiveHexproofAndPreventOneDamageFromEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Orbs of Warding", "Artifact", oldB317Orbs, false)
	if !effectstest.PlayerHasHexproof(g, me) {
		t.Error("the controller does not have hexproof")
	}
	if effectstest.PlayerHasHexproof(g, opp) {
		t.Error("an opponent has hexproof")
	}
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 1, 1)
	deal := func(src uuid.UUID, target *game.Player, n int) int {
		before := target.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, target.ID, n) })
		return before - target.Life
	}
	if got := deal(ogre, me, 4); got != 3 {
		t.Errorf("4 damage from a creature became %d, want 3", got)
	}
	if got := deal(bear, me, 1); got != 0 {
		t.Errorf("1 damage from a creature became %d, want 0", got)
	}
	if got := deal(uuid.Nil, me, 4); got != 4 {
		t.Errorf("4 damage from a noncreature source became %d, want 4", got)
	}
	if got := deal(ogre, opp, 4); got != 4 {
		t.Errorf("damage to another player became %d, want 4", got)
	}
}

// --- Molten Vortex ------------------------------------------------------------

func TestOldB317MoltenVortexDiscardsALandToDealTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vortex := effectstest.PushCatalogPermanent(g, me.ID, "Molten Vortex", "Enchantment", oldB317Vortex, false)
	land := effectstest.PushCatalogHandCard(me, "Mountain", "Basic Land — Mountain", "")
	spell := effectstest.PushCatalogHandCard(me, "Shock", "Instant", "")
	effectstest.B06AddMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, vortex, 0, game.ActivateAbilityParams{Strict: true,
		Targets: effectstest.PlayerRefs(opp.ID), DiscardIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("paid with a nonland discard")
	}
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, vortex, 0, game.ActivateAbilityParams{
		Targets: effectstest.PlayerRefs(opp.ID), DiscardIDs: []uuid.UUID{land}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 2 {
		t.Errorf("dealt %d, want 2", got)
	}
	if effectstest.HandHasID(me, land) || !effectstest.HandHasID(me, spell) {
		t.Error("the land was not the card discarded")
	}
}

// --- Reclusive Artificer ------------------------------------------------------

func oldB317CastArtificer(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, yes bool) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Reclusive Artificer", "Creature — Human Artificer", oldB317Artificer, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB317ArtificerDealsDamageEqualToYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Rock 2", "Artifact Creature — Golem")
	effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	oldB317CastArtificer(t, g, me, wall, true)
	if got := effectstest.DamageMarkedOn(g, wall); got != 2 {
		t.Errorf("took %d, want 2 (two artifacts you control; an opponent's does not count)", got)
	}
}

func TestOldB317ArtificerMayDeclineAndWithNoArtifactsDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 20)
	oldB317CastArtificer(t, g, me, wall, false)
	if got := effectstest.DamageMarkedOn(g, wall); got != 0 {
		t.Errorf("declined, yet it took %d", got)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	wall2 := effectstest.B12Creature(g2, opp2.ID, "Wall", "Creature — Wall", 0, 20)
	oldB317CastArtificer(t, g2, me2, wall2, true)
	if got := effectstest.DamageMarkedOn(g2, wall2); got != 0 {
		t.Errorf("with no artifacts it took %d", got)
	}
}

// --- Chandra's Fury -----------------------------------------------------------

func TestOldB317FuryHitsThePlayerForFourAndTheirCreaturesForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 5)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 5)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Wall", 0, 5)
	other := effectstest.B12Creature(g, third.ID, "Other", "Creature — Wall", 0, 5)
	before := opp.Life
	effectstest.D1658Cast(t, g, "Chandra's Fury", "Instant", oldB317Fury, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 4 {
		t.Errorf("the player lost %d, want 4", got)
	}
	for _, id := range []uuid.UUID{a, b} {
		if got := effectstest.DamageMarkedOn(g, id); got != 1 {
			t.Errorf("a target player's creature took %d, want 1", got)
		}
	}
	for _, id := range []uuid.UUID{mine, other} {
		if got := effectstest.DamageMarkedOn(g, id); got != 0 {
			t.Errorf("another player's creature took %d, want 0", got)
		}
	}
}

func TestOldB317FuryAtAPlaneswalkerHitsItsControllersCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	walker := effectstest.PushCatalogWalkerWithLine(g, opp.ID, "Test Walker", "Legendary Planeswalker — Test", "", 5)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 5)
	before := opp.Life
	effectstest.D1658Cast(t, g, "Chandra's Fury", "Instant", oldB317Fury, game.CastSpellParams{Targets: effectstest.CardRefs(walker)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, a); got != 1 {
		t.Errorf("the planeswalker controller's creature took %d, want 1", got)
	}
	if opp.Life != before {
		t.Error("the planeswalker's controller lost life")
	}
}
