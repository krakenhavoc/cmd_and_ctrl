package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b06_test.go — the S82 reprint pass, batch 6 (#5908). Each test
// plays the real catalog entry and asserts what its text says.

const (
	rprB06Exterminator = "fd5fc69e-1bba-4f39-88c1-5966fa941f32"
	rprB06Fallowsage   = "b9d018ff-de27-4bf6-9cc6-9d9a46db3eb9"
	rprB06Feldon       = "be094704-5232-4b10-beeb-aa9bbfae063d"
	rprB06Gargantua    = "e101edcd-34d8-4bcd-8ba6-d5941584f9fa"
	rprB06Rider        = "f797a325-ec2a-4142-b0ff-90cac1356da9"
	rprB06Conjuring    = "343460ea-7999-4d9e-b425-97fa1621f9d6"
	rprB06Equenauts    = "5212e707-b487-46b6-aa51-cea3be8cbe00"
	rprB06Powerstone   = "af4f192a-9565-4377-b359-835fe4396e85"
	rprB06Grid         = "05d75b10-4f20-4083-af27-f85c6e207fd0"
	rprB06Medics       = "457bee4c-2332-440c-9336-36f6a69614f9"
	rprB06Bridge       = "c9b7ea9c-3bcb-4538-aa25-cdb82a52037e"
)

func TestRprB06EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB06Exterminator: "Falkenrath Exterminator",
		rprB06Fallowsage:   "Fallowsage",
		rprB06Feldon:       "Feldon of the Third Path",
		rprB06Gargantua:    "First-Sphere Gargantua",
		rprB06Rider:        "Flamerush Rider",
		rprB06Conjuring:    "Flameshadow Conjuring",
		rprB06Equenauts:    "Flight of Equenauts",
		rprB06Powerstone:   "Fractured Powerstone",
		rprB06Grid:         "Ghirapur Aether Grid",
		rprB06Medics:       "Goblin Medics",
		rprB06Bridge:       "Goldmire Bridge",
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

// --- Falkenrath Exterminator -------------------------------------------------

func TestRprB06ExterminatorGrowsOnCombatDamageAndShootsForItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ext := effectstest.B12Push(g, me.ID, "Falkenrath Exterminator", "Creature — Vampire Archer", rprB06Exterminator, 1, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 5, 5)

	// Without counters the ability resolves for nothing.
	effectstest.FloatForTest(g, me, "RRR")
	if err := effectstest.HobB4Activate(t, g, me.ID, ext, 0, "", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 0 {
		t.Errorf("damage with no counters: %d, want 0", got)
	}

	effectstest.DealCombatDamageToPlayer(g, ext, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ext, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after combat damage to a player: %d, want 1", got)
	}
	effectstest.DealCombatDamageToPlayer(g, ext, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ext, game.CounterPlusOne); got != 2 {
		t.Fatalf("counters after a second hit: %d, want 2", got)
	}

	if err := effectstest.HobB4Activate(t, g, me.ID, ext, 0, "", effectstest.CardRefs(victim)); err == nil {
		t.Fatal("activated with an empty pool")
	}
	effectstest.FloatForTest(g, me, "RRR")
	if err := effectstest.HobB4Activate(t, g, me.ID, ext, 0, "", effectstest.CardRefs(victim)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, victim); got != 2 {
		t.Errorf("damage with two counters: %d, want 2", got)
	}
}

// --- Fallowsage --------------------------------------------------------------

func TestRprB06FallowsageDrawsWhenItBecomesTappedIfYouSay(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		sage := effectstest.B12Push(g, me.ID, "Fallowsage", "Creature — Merfolk Wizard", rprB06Fallowsage, 2, 2)
		effectstest.AdvanceToMain(t, g)
		hand := me.Hand.Size()
		effectstest.LciB1Attack(t, g, sage, opp.ID)
		effectstest.AssistCardSettle(t, g, me.ID, nil, yes)
		want := hand
		if yes {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("answer %v: hand %d, want %d", yes, got, want)
		}
	}
}

// --- Goblin Medics -----------------------------------------------------------

func TestRprB06MedicsPingsAnyTargetWhenItBecomesTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	medics := effectstest.B12Push(g, me.ID, "Goblin Medics", "Creature — Goblin Shaman", rprB06Medics, 1, 1)
	victim := effectstest.B12Creature(g, g.Seats[2].ID, "Third Bear", "Creature — Bear", 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(medics, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: victim}, true)
	if got := effectstest.DamageMarkedOn(g, victim); got != 1 {
		t.Errorf("damage on the target: %d, want 1", got)
	}
}

// --- Goldmire Bridge ---------------------------------------------------------

func TestRprB06GoldmireBridgeEntersTappedIndestructibleAndMakesWOrB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.B12Push(g, me.ID, "Goldmire Bridge", "Artifact Land", rprB06Bridge, 0, 0)
	c, ok := g.LookupCardForEffect(id)
	if !ok || !c.IsLand() || !c.IsArtifact() {
		t.Fatalf("not an artifact land: %+v", c)
	}
	spec, _ := effects.Lookup(rprB06Bridge)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "indestructible" || len(spec.Replacements) != 1 {
		t.Errorf("indestructible and enters-tapped are not declared: %+v", spec)
	}
	if len(spec.ManaAbilities) != 1 {
		t.Fatalf("mana abilities: %d, want 1", len(spec.ManaAbilities))
	}
	if got := spec.ManaAbilities[0].Produced; got != "{W|B}" {
		t.Errorf("produces %q, want {W|B}", got)
	}
}

// --- Flight of Equenauts -----------------------------------------------------

func TestRprB06EquenautsConvokesAndFlies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	taps := s82Creatures(g, me.ID, 3)
	_, id, err := s82Cast(t, g, "Flight of Equenauts", "Creature — Human Knight", "{7}{W}", rprB06Equenauts,
		[]string{"W"}, "WWWWW", taps, nil)
	if err != nil {
		t.Fatalf("cast with three convokers and five mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, tap := range taps {
		if !effectstest.TappedNow(g, tap) {
			t.Error("a convoking creature stayed untapped")
		}
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, id, "flying") {
		t.Error("no flying")
	}
	g2 := effectstest.NewCatalogGame(t)
	if _, _, err := s82Cast(t, g2, "Flight of Equenauts", "Creature — Human Knight", "{7}{W}", rprB06Equenauts,
		[]string{"W"}, "WWWWW", s82Creatures(g2, g2.Seats[g2.Turn.ActiveSeat].ID, 2), nil); err == nil {
		t.Error("cast for eight with only seven paid")
	}
}

// --- Ghirapur Aether Grid ----------------------------------------------------

func TestRprB06GridTapsTwoArtifactsForOneDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	grid := effectstest.B12Push(g, me.ID, "Ghirapur Aether Grid", "Enchantment", rprB06Grid, 0, 0)
	a := effectstest.B12Permanent(g, me.ID, "Rock A", "Artifact")
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, grid, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a}, Targets: target}); err == nil {
		t.Fatal("activated with one artifact")
	}
	b := effectstest.B12Permanent(g, me.ID, "Rock B", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	_ = theirs
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, grid, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{a, b}, Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, a) || !effectstest.TappedNow(g, b) {
		t.Error("the two artifacts were not tapped")
	}
	if opp.Life != life-1 {
		t.Errorf("life %d, want %d", opp.Life, life-1)
	}
}

// --- First-Sphere Gargantua --------------------------------------------------

func TestRprB06GargantuaDrawsAndLosesOnEntryAndUnearths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	hand, life := me.Hand.Size(), me.Life
	id := effectstest.HandCardFull(me, "First-Sphere Gargantua", "Creature — Phyrexian Horror", "{4}{B}{B}", rprB06Gargantua, []string{"B"})
	effectstest.FloatForTest(g, me, "BBBBBB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the card added, cast, and one drawn)", me.Hand.Size(), hand+1)
	}
	if me.Life != life-1 {
		t.Errorf("life %d, want %d", me.Life, life-1)
	}
	if !g.Battlefield.Contains(id) {
		t.Fatal("not on the battlefield")
	}

	// Unearth from the graveyard: back with haste, exiled at the end step.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g2)
	gy := effectstest.PushCatalogGraveyardCard(me2, "First-Sphere Gargantua", "Creature — Phyrexian Horror", rprB06Gargantua, 5, 4)
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: rprB06Gargantua})
	idx := -1
	for i, ab := range abilities {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no unearth")
	}
	effectstest.FloatForTest(g2, me2, "BBB")
	if err := g2.ActivateCatalogAbility(me2.ID, gy, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(gy) || !effectstest.EffectiveAbilitiesContain(t, g2, gy, "haste") {
		t.Fatal("unearthed card is not a hasty permanent")
	}
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(gy) || !g2.Exile.Contains(gy) {
		t.Error("the unearthed card was not exiled at the end step")
	}
}

// --- Fractured Powerstone ----------------------------------------------------

func TestRprB06PowerstoneAddsCAndRollsThePlanarDieAtSorcerySpeed(t *testing.T) {
	g, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAplanMortarium)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	stone := effectstest.B12Push(g, me.ID, "Fractured Powerstone", "Artifact", rprB06Powerstone, 0, 0)
	from := len(g.Events)
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("roll: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlanarRollsSince(g, from); len(got) != 1 {
		t.Fatalf("planar rolls: %v, want one", got)
	}
	if g.Planechase.DieRolls != 0 {
		t.Errorf("the roll counted toward the die's cost: %d", g.Planechase.DieRolls)
	}
	if !effectstest.TappedNow(g, stone) {
		t.Error("the Powerstone did not tap")
	}

	// The mana ability is the other {T} ability.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g2)
	stone2 := effectstest.B12Push(g2, me2.ID, "Fractured Powerstone", "Artifact", rprB06Powerstone, 0, 0)
	effectstest.B28TapForMana(t, g2, me2.ID, stone2, "")
	if got := effectstest.PoolColors(me2); len(got) != 1 || got[0] != "C" {
		t.Errorf("pool %v, want {C}", got)
	}

	// Not at sorcery speed: refused on an opponent's turn.
	g3, _ := effectstest.NewPlaneGame(t, effectstest.PlaneAplanMortarium)
	effectstest.AdvanceToMain(t, g3)
	other := g3.Seats[(g3.Turn.ActiveSeat+1)%len(g3.Seats)]
	stone3 := effectstest.B12Push(g3, other.ID, "Fractured Powerstone", "Artifact", rprB06Powerstone, 0, 0)
	if err := g3.ActivateCatalogAbility(other.ID, stone3, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("rolled the planar die on an opponent's turn")
	}
}

// At a table with no planar die the roll is nothing.
func TestRprB06PowerstoneRollDoesNothingWithoutPlanechase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	stone := effectstest.B12Push(g, me.ID, "Fractured Powerstone", "Artifact", rprB06Powerstone, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountEvents(g, game.EventRollDie) != 0 {
		t.Error("a die was rolled with no Planechase")
	}
}

// --- Feldon of the Third Path ------------------------------------------------

func TestRprB06FeldonCopiesAGraveyardCreatureAsAnArtifactWithHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	feldon := effectstest.B12Push(g, me.ID, "Feldon of the Third Path", "Legendary Creature — Human Artificer", rprB06Feldon, 2, 3)
	bear := effectstest.PushCatalogGraveyardCard(me, "Grizzly Bears", "Creature — Bear", "", 2, 2)
	land := effectstest.PushCatalogGraveyardCard(me, "Old Land", "Land", "", 0, 0)
	effectstest.FloatForTest(g, me, "RRR")
	if err := effectstest.HobB4Activate(t, g, me.ID, feldon, 0, "", effectstest.CardRefs(land)); err == nil {
		t.Fatal("targeted a noncreature card")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, feldon, 0, "", effectstest.CardRefs(bear)); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Grizzly Bears") {
		token = id
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of the Bears")
	}
	c, _ := g.LookupCardForEffect(token)
	if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() {
		t.Errorf("the copy is %q; want an artifact creature token", c.TypeLine)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, token, "haste") {
		t.Error("the copy has no haste")
	}
	if !effectstest.TappedNow(g, feldon) {
		t.Error("Feldon did not tap")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the card left the graveyard; it is not exiled")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the end step")
	}
}

// --- Flameshadow Conjuring ---------------------------------------------------

func TestRprB06ConjuringCopiesANontokenCreatureForR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Flameshadow Conjuring", "Enchantment", rprB06Conjuring, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	g.WithWriteLock(func() { _ = g.AddManaForEffect(me.ID, uuid.Nil, "{R}") })
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)

	bears := effectstest.BattlefieldIDsNamed(g, "Test Bear")
	if len(bears) != 2 {
		t.Fatalf("%d Test Bears, want the card and its copy", len(bears))
	}
	var token game.Card
	for _, id := range bears {
		if c, _ := effectstest.BattlefieldCard(g, id); c.IsToken() {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil || !effectstest.EffectiveAbilitiesContain(t, g, token.InstanceID, "haste") {
		t.Fatalf("the copy is missing or has no haste: %+v", token)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, token.InstanceID); ok {
		t.Error("the copy was not exiled at the end step")
	}
	if n := len(effectstest.BattlefieldIDsNamed(g, "Test Bear")); n != 1 {
		t.Errorf("%d Test Bears after the end step, want 1", n)
	}
}

func TestRprB06ConjuringDeclinedOrATokenMakesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Flameshadow Conjuring", "Enchantment", rprB06Conjuring, 0, 0)
	effectstest.CastCatalogSpell(t, g, "Test Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityUntilChoice(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Test Bear")); n != 1 {
		t.Errorf("%d Test Bears after declining, want 1", n)
	}
}

// --- Flamerush Rider ---------------------------------------------------------

func TestRprB06RiderCopiesAnotherAttackerTappedAndAttackingThenExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rider := effectstest.B12Push(g, me.ID, "Flamerush Rider", "Creature — Human Warrior", rprB06Rider, 3, 3)
	ogre := effectstest.B12Creature(g, me.ID, "Big Ogre", "Creature — Ogre", 4, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{rider, ogre} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	// The Rider itself is no target ("another"), nor is a non-attacker.
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		for _, bad := range []uuid.UUID{rider, idle} {
			if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
				t.Fatalf("%s was accepted as the target", bad)
			}
		}
	}
	effectstest.AssistCardSettle(t, g, me.ID, &game.TargetRef{Kind: game.TargetCard, ID: ogre}, true)
	effectstest.AnswerAttackSplit(t, g, me, 0)
	var token game.Card
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Big Ogre") {
		if c, _ := effectstest.BattlefieldCard(g, id); c.IsToken() {
			token = c
		}
	}
	if token.InstanceID == uuid.Nil {
		t.Fatal("no token copy of the Ogre")
	}
	if !token.Tapped || token.AttackingTarget != opp.ID {
		t.Errorf("the copy: tapped=%v attacking=%s, want tapped and attacking %s", token.Tapped, token.AttackingTarget, opp.ID)
	}
	attacking := false
	for _, a := range effects.AttackingCreatures(g) {
		if a == token.InstanceID {
			attacking = true
		}
	}
	if !attacking {
		t.Error("the copy is not attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token.InstanceID) {
		t.Error("the copy was not exiled at end of combat")
	}
	if !g.Battlefield.Contains(ogre) {
		t.Error("the original Ogre left")
	}
}

func TestRprB06RiderDashesBackToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Flamerush Rider", "Creature — Human Warrior", "{4}{R}", rprB06Rider, []string{"R"})
	effectstest.FloatForTest(g, me, "RRRR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "dash"}); err != nil {
		t.Fatalf("dash: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) || !effectstest.EffectiveAbilitiesContain(t, g, id, "haste") {
		t.Fatal("dashed Rider is not a hasty permanent")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(id) || !me.Hand.Contains(id) {
		t.Error("the dashed Rider did not return to hand")
	}
}
