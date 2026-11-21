package testsg

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGuffbSliceIsRegisteredFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{effectstest.GuffbIchormoonOracle, effectstest.GuffbJayaOracle, effectstest.GuffbLeoriOracle, effectstest.GuffbReverbOracle,
		effectstest.GuffbRowanOracle, effectstest.GuffbOathOracle, effectstest.GuffbNarsetOracle, effectstest.GuffbSaheeliOracle, effectstest.GuffbChainVeilOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// guffbPool is the mana in p's pool.
func guffbPool(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.ManaPool.Total() })
	return n
}

// guffbKeepTargets answers every open "choose new targets" prompt by
// keeping the target, then lets the stack settle.
func guffbKeepTargets(t *testing.T, g *game.Game, keep game.TargetRef) {
	t.Helper()
	for i := 0; i < 8; i++ {
		c := effectstest.LatestChoiceOfKind(g, game.PendingChoicePickTarget)
		if c == nil {
			return
		}
		if err := g.ResolvePickTargets(c.ID, c.Chooser, []game.TargetRef{keep}); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

// guffbRequirementRecords counts the attack-requirement records.
func guffbRequirementRecords(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddAttackRequirement {
				n++
			}
		}
	}
	return n
}

// --- Narset of the Ancient Way --------------------------------------

// +1: two life and one mana of the three colours, spendable only on a
// noncreature spell.
func TestGuffbNarsetPlusOneGainsLifeAndAddsRestrictedMana(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	n := effectstest.GuffbWalker(g, me.ID, "Narset of the Ancient Way", "Legendary Planeswalker — Narset", effectstest.GuffbNarsetOracle, 4)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, n, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil || len(pick.PickOptions) != 3 {
		t.Fatalf("no three-colour pick: %+v", pick)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, 1); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 {
		t.Fatalf("pool %+v, want one mana", units)
	}
	want := map[string]bool{effects.ManaRestrictCast: false, effects.ManaRestrictNotType("Creature"): false}
	for _, r := range units[0].Restrictions {
		want[r] = true
	}
	for r, seen := range want {
		if !seen {
			t.Errorf("the mana lacks restriction %q: %v", r, units[0].Restrictions)
		}
	}
}

// −2: a nonland discard deals its mana value to a target creature; a
// land discard deals nothing.
func TestGuffbNarsetMinusTwoDamagesByTheDiscardedCardsManaValue(t *testing.T) {
	for _, c := range []struct {
		name     string
		typeLine string
		cost     string
		hits     bool
	}{
		{"a nonland card", "Sorcery", "{1}{R}{R}", true},
		{"a land", "Basic Land — Island", "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, opp := effectstest.SpendTable(t)
			n := effectstest.GuffbWalker(g, me.ID, "Narset of the Ancient Way", "Legendary Planeswalker — Narset", effectstest.GuffbNarsetOracle, 4)
			pitch := effectstest.HandCardOf(g, me, "Pitch", c.typeLine, c.cost, "")
			victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 4, 5)
			effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Drawn", "Instant"))
			effectstest.B16Activate(t, g, me.ID, n, 1, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerDiscard(t, g, me.ID, pitch)
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if !c.hits {
					t.Fatal("a land discard created the reflexive trigger")
				}
				effectstest.PickCard(t, g, me.ID, victim)
				effectstest.PassPriorityAroundTable(t, g)
			}
			var damage int
			g.WithWriteLock(func() { damage = effectstest.FindBattlefieldCardForTest(g, victim).DamageMarked })
			want := 0
			if c.hits {
				want = 3
			}
			if damage != want {
				t.Errorf("damage %d, want %d", damage, want)
			}
		})
	}
}

// −6: the emblem deals 2 to any target on each noncreature spell, and
// nothing on a creature spell.
func TestGuffbNarsetEmblemPingsOnNoncreatureSpells(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	n := effectstest.GuffbWalker(g, me.ID, "Narset of the Ancient Way", "Legendary Planeswalker — Narset", effectstest.GuffbNarsetOracle, 6)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 4, 5)
	effectstest.B16Activate(t, g, me.ID, n, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 {
		t.Fatalf("emblems %d, want 1", me.Emblems.Size())
	}

	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	g.RunStateChecksForTest()
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature spell triggered the emblem")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	var damage int
	g.WithWriteLock(func() { damage = effectstest.FindBattlefieldCardForTest(g, victim).DamageMarked })
	if damage != 2 {
		t.Errorf("damage %d, want 2", damage)
	}
}

// --- Saheeli, Sublime Artificer -------------------------------------

func TestGuffbSaheeliMakesAServoOnEachNoncreatureSpell(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.GuffbWalker(g, me.ID, "Saheeli, Sublime Artificer", "Legendary Planeswalker — Saheeli", effectstest.GuffbSaheeliOracle, 5)
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Servo"); n != 0 {
		t.Fatalf("%d Servos after a creature spell, want 0", n)
	}
	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountBattlefieldNamed(g, me.ID, "Servo"); n != 1 {
		t.Errorf("%d Servos after a sorcery, want 1", n)
	}
}

// −2: the artifact becomes a copy of the other target until end of
// turn and is still an artifact; the same permanent can't be both
// targets.
func TestGuffbSaheeliMinusTwoMakesAnArtifactCopy(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	s := effectstest.GuffbWalker(g, me.ID, "Saheeli, Sublime Artificer", "Legendary Planeswalker — Saheeli", effectstest.GuffbSaheeliOracle, 5)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 3, 3)

	same := []game.TargetRef{{Kind: game.TargetCard, ID: rock, Slot: 0}, {Kind: game.TargetCard, ID: rock, Slot: 1}}
	if err := g.ActivateCatalogAbility(me.ID, s, 0, game.ActivateAbilityParams{Targets: same}); err == nil {
		t.Fatal("−2 took the same artifact as both targets")
	}
	effectstest.B16Activate(t, g, me.ID, s, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: rock, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1},
	}})
	effectstest.PassPriorityAroundTable(t, g)
	var name string
	var creature, artifact bool
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, rock)
		name, creature, artifact = c.Name, c.IsCreature(), c.IsArtifact()
	})
	if name != "My Bear" || !creature || !artifact {
		t.Errorf("the Rock is %q creature=%v artifact=%v, want an artifact copy of My Bear", name, creature, artifact)
	}
}

// --- Rowan Kenrith --------------------------------------------------

// +2 on an opponent: their creatures must attack on their next turn,
// yours are not bound this turn, and it is over after their turn.
func TestGuffbRowanPlusTwoForcesTheTargetsNextTurn(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	r := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 4)
	effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, r, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasRequirementRecord(g) {
		t.Fatal("no attack requirement registered")
	}
	if effectstest.FaMustAttack(t, g) {
		t.Fatal("your own creature was bound to attack")
	}
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if !effectstest.FaMustAttack(t, g) {
		t.Fatal("the opponent's creature was not required to attack on their turn")
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("the opponent passed without attacking")
	}
}

// +2 on yourself: nothing binds this turn; a delayed trigger at your
// next upkeep makes the requirement for that turn.
func TestGuffbRowanPlusTwoOnYourselfWaitsForYourNextTurn(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	r := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 4)
	effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.B16Activate(t, g, me.ID, r, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasRequirementRecord(g) {
		t.Fatal("the requirement binds the rest of this turn")
	}
	if effectstest.FaMustAttack(t, g) {
		t.Fatal("your creature must attack this turn")
	}
	// Round the table to your next upkeep and let the trigger resolve.
	effectstest.AdvanceTo(t, g, game.StepEnd)
	for i := 0; i < 300 && !(g.Turn.ActiveSeat == 0 && g.Turn.Step == game.StepUpkeep); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasRequirementRecord(g) {
		t.Fatal("no requirement on your next turn")
	}
	if !effectstest.FaMustAttack(t, g) {
		t.Error("your creature was not required to attack on your next turn")
	}
}

// −2: 3 damage to each TAPPED creature the target player controls.
func TestGuffbRowanMinusTwoHitsTheTargetsTappedCreatures(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	r := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 4)
	tapped := effectstest.PushVanillaCreature(g, opp.ID, "Tapped", 4, 4)
	untapped := effectstest.PushVanillaCreature(g, opp.ID, "Untapped", 4, 4)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 4, 4)
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true
		effectstest.FindBattlefieldCardForTest(g, mine).Tapped = true
	})
	effectstest.B16Activate(t, g, me.ID, r, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		for id, want := range map[uuid.UUID]int{tapped: 3, untapped: 0, mine: 0} {
			if got := effectstest.FindBattlefieldCardForTest(g, id).DamageMarked; got != want {
				t.Errorf("%v: damage %d, want %d", id, got, want)
			}
		}
	})
}

// −8: the TARGET player gets the emblem, and it copies their next
// activated ability.
func TestGuffbRowanEmblemGoesToTheTargetAndCopiesAbilities(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	r := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 8)
	effectstest.B16Activate(t, g, me.ID, r, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 1 || opp.Emblems.Size() != 0 {
		t.Fatalf("emblems: me %d, opp %d; want 1 and 0", me.Emblems.Size(), opp.Emblems.Size())
	}
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := guffbPool(g, me); got != 4 {
		t.Errorf("pool %d, want 4: the +1 and its copy", got)
	}
}

func TestGuffbRowanEmblemToAnOpponent(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	r := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 8)
	effectstest.B16Activate(t, g, me.ID, r, 2, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Emblems.Size() != 0 || opp.Emblems.Size() != 1 {
		t.Fatalf("emblems: me %d, opp %d; want 0 and 1", me.Emblems.Size(), opp.Emblems.Size())
	}
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := guffbPool(g, me); got != 2 {
		t.Errorf("pool %d, want 2: the opponent's emblem copies only their abilities", got)
	}
}

// --- Repeated Reverberation -----------------------------------------

// The next loyalty ability is copied twice; a creature spell before it
// does not use the trigger up.
func TestGuffbReverberationCopiesTheNextLoyaltyAbilityTwice(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.CastCatalogSpell(t, g, "Repeated Reverberation", "Instant", effectstest.GuffbReverbOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := guffbPool(g, me); got != 6 {
		t.Errorf("pool %d, want 6: the +1 and two copies", got)
	}
	if got := effectstest.LoyaltyOf(g, c); got != 5 {
		t.Errorf("loyalty %d, want 5: copies pay nothing", got)
	}
}

// The next instant is copied twice, and the trigger is used up.
func TestGuffbReverberationCopiesTheNextInstantTwice(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.CastCatalogSpell(t, g, "Repeated Reverberation", "Instant", effectstest.GuffbReverbOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	at := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", []game.TargetRef{at})
	effectstest.PassPriorityAroundTable(t, g)
	guffbKeepTargets(t, g, at)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-9 {
		t.Errorf("opponent life %d, want %d: the Bolt and two copies", opp.Life, life-9)
	}
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "4457ed35-7c10-48c8-9776-456485fdf070", []game.TargetRef{at})
	effectstest.PassPriorityAroundTable(t, g)
	guffbKeepTargets(t, g, at)
	if opp.Life != life-12 {
		t.Errorf("opponent life %d, want %d: the trigger was used up", opp.Life, life-12)
	}
	_ = me
}

// --- Jaya's Phoenix -------------------------------------------------

// Combat damage sets up the copy of your next loyalty ability this
// turn.
func TestGuffbJayasPhoenixCopiesTheNextLoyaltyAbility(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	phoenix := effectstest.PushCatalogPermanent(g, me.ID, "Jaya's Phoenix", "Creature — Phoenix", effectstest.GuffbJayaOracle, false)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.AttackWith(t, g, opp.ID, phoenix)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := guffbPool(g, me); got != 4 {
		t.Errorf("pool %d, want 4: the +1 and its copy", got)
	}
}

// From the graveyard, a planeswalker spell offers the return; another
// spell does not.
func TestGuffbJayasPhoenixReturnsOnAPlaneswalkerSpell(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	phoenix := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: phoenix, Name: "Jaya's Phoenix", TypeLine: "Creature — Phoenix",
		OracleID: effectstest.GuffbJayaOracle, Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})

	const label = "Jaya's Phoenix — return it from your graveyard to the battlefield"
	onStack := func() bool {
		for _, it := range g.StackMeta {
			if it != nil && it.Label == label {
				return true
			}
		}
		return false
	}
	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	if onStack() {
		t.Fatal("a sorcery triggered the return")
	}
	effectstest.PassPriorityAroundTable(t, g)

	effectstest.CastCatalogSpell(t, g, "A Walker", "Legendary Planeswalker — Teyo", "", nil)
	g.RunStateChecksForTest()
	if !onStack() {
		t.Fatal("a planeswalker spell did not trigger the return")
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt)
	if p == nil {
		for _, c := range g.PendingChoices {
			t.Logf("pending %s %q", c.Kind, c.Reason)
		}
		t.Fatal("the return asked nothing as it resolved")
	}
	if err := g.ResolveTriggerPrompt(p.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Jaya's Phoenix") || me.Graveyard.Contains(phoenix) {
		t.Error("the Phoenix did not return to the battlefield")
	}
}

// --- Leori, Sparktouched Hunter -------------------------------------

// The chosen type's abilities are copied for the rest of the turn;
// another type's are not.
func TestGuffbLeoriCopiesAbilitiesOfTheChosenPlaneswalkerType(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	leori := effectstest.PushCatalogPermanent(g, me.ID, "Leori, Sparktouched Hunter", "Legendary Creature — Elemental Cat", effectstest.GuffbLeoriOracle, false)
	c := effectstest.GuffbChandra(g, me.ID)
	other := effectstest.GuffbWalker(g, me.ID, "Rowan Kenrith", "Legendary Planeswalker — Rowan", effectstest.GuffbRowanOracle, 4)
	effectstest.AttackWith(t, g, opp.ID, leori)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil {
		t.Fatal("no planeswalker-type prompt")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "Chandra" {
			idx = i
		}
	}
	if idx < 0 || idx > 1 {
		t.Fatalf("Chandra is option %d, want one of the first two (types you control first)", idx)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)

	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if got := guffbPool(g, me); got != 4 {
		t.Errorf("pool %d, want 4: the Chandra +1 and its copy", got)
	}
	// A Rowan ability is not copied: +2 makes one requirement, not two.
	effectstest.B16Activate(t, g, me.ID, other, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	guffbKeepTargets(t, g, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if n := guffbRequirementRecords(g); n != 1 {
		t.Errorf("%d attack requirements after a Rowan +2, want 1 (no copy)", n)
	}
}

// --- Oath of Teferi -------------------------------------------------

// Each of your planeswalkers may activate twice, not three times; an
// opponent's Oath gives you nothing.
func TestGuffbOathOfTeferiLetsYourWalkersActivateTwice(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, opp.ID, "Oath of Teferi", "Legendary Enchantment", effectstest.GuffbOathOracle, false)
	c := effectstest.GuffbChandra(g, me.ID)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("an opponent's Oath gave your walker a second activation")
	}

	effectstest.PushCatalogPermanent(g, me.ID, "Oath of Teferi", "Legendary Enchantment", effectstest.GuffbOathOracle, false)
	effectstest.B16Activate(t, g, me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, c, effectstest.GuffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed")
	}
	if got := effectstest.LoyaltyOf(g, c); got != 6 {
		t.Errorf("loyalty %d, want 6 after two +1s", got)
	}
}

// The enters trigger exiles another permanent you control until the
// next end step, returning it under its owner's control.
func TestGuffbOathOfTeferiBlinksAnotherPermanentUntilTheEndStep(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	stolen := effectstest.RoonaStolen(g, opp.ID, me.ID, "Their Bear")
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Not Mine", 2, 2)
	oath := effectstest.CastCatalogSpell(t, g, "Oath of Teferi", "Legendary Enchantment", effectstest.GuffbOathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	for what, id := range map[string]uuid.UUID{"an opponent's permanent": theirs, "the Oath itself": oath} {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: id}); err == nil {
			t.Fatalf("%s was accepted as the target", what)
		}
	}
	effectstest.PickCard(t, g, me.ID, stolen)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, stolen) {
		t.Fatal("the permanent was not exiled")
	}
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	back := effectstest.RoonaOnly(t, g, "Their Bear")
	if back.Controller != opp.ID {
		t.Errorf("returned under %v, want its owner %v", back.Controller, opp.ID)
	}
}

// --- Ichormoon Gauntlet ---------------------------------------------

// Your planeswalkers have the −12 (an extra turn) and the [0].
func TestGuffbIchormoonGrantsTheRowsToYourWalkers(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Ichormoon Gauntlet", "Artifact", effectstest.GuffbIchormoonOracle, false)
	walker := effectstest.TalentWalker(g, me.ID, 12)
	theirs := effectstest.TalentWalker(g, opp.ID, 12)
	effectstest.GrantedLoyaltyRow(t, g, walker, "0:")
	idx, ref := effectstest.GrantedLoyaltyRow(t, g, walker, "−12")
	abs, origins := game.ActivatedAbilitiesWithOrigins(effectstest.GaLayered(t, g, theirs))
	for i := range abs {
		if origins.At(i).Granted() {
			t.Errorf("an opponent's walker has the granted row %q", abs[i].Label)
		}
	}
	turns := len(g.ExtraTurns)
	if err := g.ActivateCatalogAbility(me.ID, walker, idx, game.ActivateAbilityParams{Ref: ref}); err != nil {
		t.Fatalf("activate −12: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.ExtraTurns) != turns+1 {
		t.Errorf("extra turns %d, want %d", len(g.ExtraTurns), turns+1)
	}
}

// A noncreature spell puts one more counter of a kind already on the
// target: at once with one kind, after a choice with two.
func TestGuffbIchormoonAddsACounterOfAChosenKind(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Ichormoon Gauntlet", "Artifact", effectstest.GuffbIchormoonOracle, false)
	walker := effectstest.TalentWalker(g, me.ID, 3)
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(bear, "charge", 2)
	})

	effectstest.CastCatalogSpell(t, g, "Some Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LoyaltyOf(g, walker); got != 4 {
		t.Errorf("loyalty %d, want 4", got)
	}

	effectstest.CastCatalogSpell(t, g, "Another Sorcery", "Sorcery", "", nil)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceOptionPick)
	if pick == nil {
		t.Fatal("no counter-kind prompt for two kinds")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if o.Label == "A charge counter" {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no charge option in %+v", pick.PickOptions)
	}
	if err := g.ResolveOptionPick(pick.ID, me.ID, idx); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, bear)
		if c.Counters["charge"] != 3 || c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("counters %v, want 3 charge and 1 +1/+1", c.Counters)
		}
	})

	effectstest.CastCatalogSpell(t, g, "A Bear", "Creature — Bear", "", nil)
	g.RunStateChecksForTest()
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature spell triggered the Gauntlet")
	}
}
