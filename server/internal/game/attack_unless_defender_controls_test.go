package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// attack_unless_defender_controls_test.go pins ADR 0107 §2 (#1879): "This
// creature can't attack unless defending player controls an Island" (CR
// 508.1c). CR 508.5 makes the defending player a fact about each target —
// the player attacked, a planeswalker's controller, a battle's protector —
// and CR 508.5a makes it one player, worked out per creature. So the
// restriction narrows the creature's targets per opponent. The enumerator
// half is pinned in internal/legal, the wire half in internal/protocol,
// the card constructor in cards/effects.

const (
	serpentOracle      = "test-island-serpent"       // can't attack unless defending player controls an Island
	needySerpentOracle = "test-needy-island-serpent" // the same, and attacks each combat if able
	octopusOracle      = "test-godhunter-octopus"    // … an enchantment or an enchanted permanent
)

var islandQuery = []PermanentQuery{{Subtypes: []string{"Island"}}}

// withDefenderMustControlStatics installs the test cards' statics in the
// shape effects.CantAttackUnlessDefendingPlayerControls writes them.
func withDefenderMustControlStatics(t *testing.T) {
	t.Helper()
	self := func(target *Card, _ *Game, source *Card) bool {
		return target.InstanceID == source.InstanceID && target.IsCreature()
	}
	restrict := func(qs []PermanentQuery) StaticAbility {
		return StaticAbility{
			Layer:     Layer6Ability,
			AppliesTo: self,
			Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
				c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, AttackTargetRestriction{
					Source: source.InstanceID, SourceName: source.Name, DefenderMustControl: qs,
				})
			},
		}
	}
	mustAttack := StaticAbility{
		Layer:     Layer6Ability,
		AppliesTo: self,
		Apply: func(c *Characteristic, _ *Card, _ *Game, source *Card) {
			c.AttackRequirements = append(c.AttackRequirements, AttackRequirement{Source: source.InstanceID, SourceName: source.Name})
		},
	}
	withStaticAbilities(t, func(oracleID string) []StaticAbility {
		switch oracleID {
		case serpentOracle:
			return []StaticAbility{restrict(islandQuery)}
		case needySerpentOracle:
			return []StaticAbility{mustAttack, restrict(islandQuery)}
		case octopusOracle:
			return []StaticAbility{restrict([]PermanentQuery{{Types: []string{"enchantment"}}, {Enchanted: true}})}
		}
		return nil
	})
}

func pushSerpent(g *Game, controller uuid.UUID, oracle string) uuid.UUID {
	return pushOwnedByAnother(g, controller, controller, "Sea Serpent", oracle)
}

func pushLandForTest(g *Game, controller uuid.UUID, name, typeLine string) uuid.UUID {
	return pushTypedTestCard(g, Card{Name: name, TypeLine: typeLine, Owner: controller, Controller: controller})
}

// TestCantAttackUnlessDefenderControlsIsPerDefendingPlayer — Sea Serpent in
// four seats. It may attack the opponent with an Island, that opponent's
// planeswalker and a battle that opponent protects; it may not attack the
// opponent without one, their planeswalker, or a battle they protect —
// even a battle the Island player controls (CR 508.5: a battle's defending
// player is its protector).
func TestCantAttackUnlessDefenderControlsIsPerDefendingPlayer(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newFourPlayerActiveGame(t)
	me, islands, dry, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	serpent := pushSerpent(g, me.ID, serpentOracle)
	pushLandForTest(g, islands.ID, "Island", "Basic Land — Island")
	pushLandForTest(g, dry.ID, "Mountain", "Basic Land — Mountain")
	// A nonbasic Island counts: it is the subtype, not the name.
	pushLandForTest(g, third.ID, "Volcanic Island", "Land — Island Mountain")
	islandsWalker := pushPlaneswalkerForTest(g, islands.ID, "Islands' Walker", 3)
	dryWalker := pushPlaneswalkerForTest(g, dry.ID, "Dry Walker", 3)
	protectedByIslands := pushBattleForTest(g, dry.ID, islands.ID, "Siege Islands Protect", 3)
	protectedByDry := pushBattleForTest(g, islands.ID, dry.ID, "Siege Dry Protects", 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, serpent)
	for _, want := range []uuid.UUID{islands.ID, third.ID, islandsWalker, protectedByIslands} {
		if !slices.Contains(got, want) {
			t.Errorf("per-attacker targets %v miss %s", got, want)
		}
	}
	for _, refused := range []uuid.UUID{dry.ID, dryWalker, protectedByDry, me.ID} {
		if slices.Contains(got, refused) {
			t.Errorf("per-attacker targets %v include the forbidden %s", got, refused)
		}
	}
	// The per-seat list is unchanged: it is about prices, not one creature.
	var perSeat []uuid.UUID
	g.WithWriteLock(func() { perSeat = attackRefIDs(g.AttackTargetsForEffect(me.ID)) })
	if !slices.Contains(perSeat, dry.ID) {
		t.Errorf("per-seat targets %v lost the dry opponent", perSeat)
	}

	re := restrictionErr(t, g.Clone().DeclareAttacker(serpent, dry.ID))
	if got, want := re.Sentence(), "Sea Serpent can't attack P3: P3 controls no Island."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	re = restrictionErr(t, g.Clone().DeclareAttacker(serpent, dryWalker))
	if got, want := re.Sentence(), "Sea Serpent can't attack Dry Walker: P3 controls no Island."; got != want {
		t.Errorf("planeswalker sentence = %q, want %q", got, want)
	}
	restrictionErr(t, g.Clone().DeclareAttacker(serpent, protectedByDry))
	if _, err := g.Clone().DeclareAttackers([]AttackDeclaration{{Attacker: serpent, Target: dry.ID}}); !errors.Is(err, ErrNoLegalAttackers) {
		t.Errorf("bulk declaration at the dry opponent: err = %v, want ErrNoLegalAttackers", err)
	}
	for _, ok := range []uuid.UUID{islands.ID, third.ID, islandsWalker, protectedByIslands} {
		if err := g.Clone().DeclareAttacker(serpent, ok); err != nil {
			t.Errorf("declare at %s: %v", ok, err)
		}
	}
}

// TestCantAttackUnlessDefenderControlsReadsTheBoardLive — the land types
// are effective (an effect that makes a Mountain an Island counts), the
// board is read at declaration, and an Island that leaves afterwards
// changes nothing (CR 506.4a).
func TestCantAttackUnlessDefenderControlsReadsTheBoardLive(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := pushSerpent(g, me.ID, serpentOracle)
	mountain := pushLandForTest(g, opp.ID, "Mountain", "Basic Land — Mountain")
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := attackTargetsOf(g, serpent); len(got) != 0 {
		t.Fatalf("targets = %v, want none: the only opponent controls no Island", got)
	}
	restrictionErr(t, g.Clone().DeclareAttacker(serpent, opp.ID))

	registerScopedEffectForTest(t, g, mountain, []Mod{AddSubtypesMod("Island")}, IndefiniteDuration())
	if got := attackTargetsOf(g, serpent); !slices.Contains(got, opp.ID) {
		t.Fatalf("targets = %v: a Mountain that is also an Island is an Island", got)
	}
	if err := g.DeclareAttacker(serpent, opp.ID); err != nil {
		t.Fatalf("declare once the opponent controls an Island: %v", err)
	}
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: opp.ID}, mountain); err != nil {
		t.Fatalf("move the Island away: %v", err)
	}
	if c := findBattlefieldCard(g, serpent); c == nil || c.AttackingTarget != opp.ID {
		t.Errorf("the Serpent stopped attacking when the Island left: %+v", c)
	}
}

// TestCantAttackUnlessDefenderControlsYieldsToTheRequirement — CR 508.1d
// counts requirements "without disobeying any restrictions". A creature
// that must attack and has no opponent with an Island owes nothing.
func TestCantAttackUnlessDefenderControlsYieldsToTheRequirement(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newFourPlayerActiveGame(t)
	me, islands := g.Seats[0], g.Seats[1]
	serpent := pushSerpent(g, me.ID, needySerpentOracle)
	advanceIntoStep(t, g, StepDeclareAttackers)

	var owed map[uuid.UUID][]uuid.UUID
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if len(owed) != 0 {
		t.Fatalf("owed = %v, want nothing: no opponent controls an Island", owed)
	}
	if err := g.Clone().PassPriority(); err != nil {
		t.Fatalf("pass with nobody to attack: %v", err)
	}

	pushLandForTest(g, islands.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() { owed = g.MustAttackForEffect() })
	if !slices.Equal(owed[serpent], []uuid.UUID{islands.ID}) {
		t.Fatalf("owed answers = %v, want only the opponent with an Island", owed[serpent])
	}
	requirementErr(t, g.Clone().PassPriority())
}

// TestCantAttackUnlessDefenderControlsGoesWithItsAbilities — CR 613.1f: a
// creature that loses all abilities may attack anyone.
func TestCantAttackUnlessDefenderControlsGoesWithItsAbilities(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	serpent := pushSerpent(g, me.ID, serpentOracle)
	registerScopedEffectForTest(t, g, serpent, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(serpent, opp.ID); err != nil {
		t.Fatalf("declare after losing all abilities: %v", err)
	}
}

// TestCantAttackUnlessDefenderControlsAnyOfItsQueries — Godhunter Octopus's
// "an enchantment or an enchanted permanent": either half lets it attack.
// An Equipment does not enchant.
func TestCantAttackUnlessDefenderControlsAnyOfItsQueries(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newFourPlayerActiveGame(t)
	me, withEnchantment, withEnchanted, withEquipped := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	octopus := pushOwnedByAnother(g, me.ID, me.ID, "Godhunter Octopus", octopusOracle)
	pushTypedTestCard(g, Card{Name: "Test Enchantment", TypeLine: "Enchantment", Owner: withEnchantment.ID, Controller: withEnchantment.ID})
	// An Aura controlled by ME on a creature P3 controls: P3 controls an
	// enchanted permanent, not an enchantment.
	bear := pushTypedTestCard(g, Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: withEnchanted.ID, Controller: withEnchanted.ID})
	pushTypedTestCard(g, Card{Name: "Test Aura", TypeLine: "Enchantment — Aura", Owner: me.ID, Controller: me.ID,
		AttachedTo: TargetRef{Kind: TargetCard, ID: bear}})
	wolf := pushTypedTestCard(g, Card{Name: "Wolf", TypeLine: "Creature — Wolf", Power: 2, Toughness: 2, Owner: withEquipped.ID, Controller: withEquipped.ID})
	pushTypedTestCard(g, Card{Name: "Test Sword", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID,
		AttachedTo: TargetRef{Kind: TargetCard, ID: wolf}})
	advanceIntoStep(t, g, StepDeclareAttackers)

	got := attackTargetsOf(g, octopus)
	if !slices.Contains(got, withEnchantment.ID) || !slices.Contains(got, withEnchanted.ID) || slices.Contains(got, withEquipped.ID) {
		t.Fatalf("targets = %v: want P2 (an enchantment) and P3 (an enchanted permanent), not P4 (an equipped one)", got)
	}
	re := restrictionErr(t, g.Clone().DeclareAttacker(octopus, withEquipped.ID))
	if got, want := re.Sentence(), "Godhunter Octopus can't attack P4: P4 controls no enchantment or enchanted permanent."; got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
}

// TestPermanentQueryMatchesEffectiveCharacteristics — each printed
// condition in the 34 cards, as data, against the permanents that do and
// don't satisfy it.
func TestPermanentQueryMatchesEffectiveCharacteristics(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[1].ID
	snowIsland := pushLandForTest(g, p, "Snow-Covered Island", "Basic Snow Land — Island")
	mountain := pushLandForTest(g, p, "Mountain", "Basic Land — Mountain")
	flier := pushTypedTestCard(g, Card{Name: "Bird", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Owner: p, Controller: p,
		Keywords: []string{"flying"}})
	blue := pushTypedTestCard(g, Card{Name: "Blue Thing", TypeLine: "Artifact", ManaCost: "{U}", Owner: p, Controller: p})

	cases := []struct {
		q    PermanentQuery
		noun string
		yes  []uuid.UUID
		no   []uuid.UUID
	}{
		{PermanentQuery{Subtypes: []string{"Island"}}, "Island", []uuid.UUID{snowIsland}, []uuid.UUID{mountain, flier, blue}},
		{PermanentQuery{Subtypes: []string{"Mountain"}}, "Mountain", []uuid.UUID{mountain}, []uuid.UUID{snowIsland}},
		{PermanentQuery{Supertypes: []string{"snow"}, Types: []string{"land"}}, "snow land", []uuid.UUID{snowIsland}, []uuid.UUID{mountain}},
		{PermanentQuery{Colors: []string{"U"}}, "blue permanent", []uuid.UUID{blue}, []uuid.UUID{snowIsland, flier}},
		{PermanentQuery{Types: []string{"creature"}, Keyword: "flying"}, "creature with flying", []uuid.UUID{flier}, []uuid.UUID{blue}},
		{PermanentQuery{Enchanted: true}, "enchanted permanent", nil, []uuid.UUID{flier}},
	}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, tc := range cases {
			if got := tc.q.Noun(); got != tc.noun {
				t.Errorf("%+v: noun = %q, want %q", tc.q, got, tc.noun)
			}
			for _, id := range tc.yes {
				if !tc.q.matchesLocked(g, findBattlefieldCard(g, id)) {
					t.Errorf("%s does not match %s", tc.noun, findBattlefieldCard(g, id).Name)
				}
			}
			for _, id := range tc.no {
				if tc.q.matchesLocked(g, findBattlefieldCard(g, id)) {
					t.Errorf("%s matches %s", tc.noun, findBattlefieldCard(g, id).Name)
				}
			}
		}
	})
}

// TestDefenderRefusalsNameEachOpponentWithoutAMatch — the chip's data: one
// row per live opponent the creature can't attack right now.
func TestDefenderRefusalsNameEachOpponentWithoutAMatch(t *testing.T) {
	withDefenderMustControlStatics(t)
	g := newFourPlayerActiveGame(t)
	me, islands, dry, gone := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	serpent := pushSerpent(g, me.ID, serpentOracle)
	pushLandForTest(g, islands.ID, "Island", "Basic Land — Island")
	gone.Eliminated = true
	var got []DefenderRefusal
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		got = g.DefenderRefusalsForEffect(findBattlefieldCard(g, serpent))
	})
	if len(got) != 1 || got[0].Player != dry.ID || got[0].Restriction.SourceName != "Sea Serpent" {
		t.Fatalf("refusals = %+v, want only P3", got)
	}
}
