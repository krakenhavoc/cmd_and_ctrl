package game

import (
	"testing"

	"github.com/google/uuid"
)

// targets_test.go — S20 sub-PR 1: TargetSpec legality at announce
// and resolution, legal-target enumeration, colour derivation.

// withCatalogTargetSpec swaps the CatalogTargetSpec hook for the
// test's duration.
func withCatalogTargetSpec(t *testing.T, fn func(oracleID string) *TargetSpec) {
	t.Helper()
	prev := CatalogTargetSpec
	CatalogTargetSpec = fn
	t.Cleanup(func() { CatalogTargetSpec = prev })
}

// nonBlackCreatureSpec is Doom Blade's clause.
func nonBlackCreatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature() && !c.HasColor("B")
		},
		Min: 1, Max: 1,
	}
}

func pushColoredCreature(g *Game, owner *Player, name, manaCost string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Creature — Test"
	c.ManaCost = manaCost
	c.Power, c.Toughness = 2, 2
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func TestEffectiveColorsFromManaCostAndStamped(t *testing.T) {
	c := Card{ManaCost: "{1}{W/U}{B}"}
	got := c.EffectiveColors()
	if len(got) != 3 || !c.HasColor("W") || !c.HasColor("U") || !c.HasColor("B") || c.HasColor("R") {
		t.Errorf("derived colours = %v", got)
	}
	stamped := Card{ManaCost: "{2}", Colors: []string{"G"}}
	if !stamped.HasColor("G") || stamped.IsColorless() {
		t.Errorf("stamped Colors should win over the mana-cost derivation")
	}
	if !(Card{ManaCost: "{3}"}).IsColorless() {
		t.Errorf("{3} should be colourless")
	}
}

func TestLegalTargetsFiltersByPredicateAndZone(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	black := pushColoredCreature(g, opp, "Black Knight", "{B}{B}")
	mine := pushColoredCreature(g, me, "My Bear", "{1}{G}")
	rock := NewCard("Sol Ring", opp.ID)
	rock.TypeLine = "Artifact"
	g.Battlefield.PushTop(rock)
	// A creature in a graveyard must not show up for a battlefield spec.
	dead := NewCard("Dead Guy", opp.ID)
	dead.TypeLine = "Creature — Zombie"
	opp.Graveyard.PushTop(dead)

	lt := g.LegalTargetsForEffect(me.ID, nonBlackCreatureSpec())
	has := func(id uuid.UUID) bool {
		for _, c := range lt.Cards {
			if c == id {
				return true
			}
		}
		return false
	}
	if !has(white) || !has(mine) {
		t.Errorf("non-black creatures missing from legal set: %v", lt.Cards)
	}
	if has(black) || has(rock.InstanceID) || has(dead.InstanceID) {
		t.Errorf("illegal candidates present: %v", lt.Cards)
	}
	if len(lt.Players) != 0 {
		t.Errorf("creature spec must not list players")
	}
}

func TestCastSpellRejectsIllegalTargetAndAcceptsLegal(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	white := pushColoredCreature(g, opp, "White Knight", "{W}{W}")
	black := pushColoredCreature(g, opp, "Black Knight", "{B}{B}")
	const oracle = "test-doom-blade"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	blade := NewCard("Doom Blade", me.ID)
	blade.TypeLine = "Instant"
	blade.OracleID = oracle
	me.Hand.PushTop(blade)

	err := g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: black}},
	})
	if err != ErrIllegalTarget {
		t.Fatalf("black creature: got %v, want ErrIllegalTarget", err)
	}
	if !me.Hand.Contains(blade.InstanceID) {
		t.Fatalf("rejected cast must leave the card in hand")
	}
	err = g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetPlayer, ID: opp.ID}},
	})
	if err != ErrIllegalTarget {
		t.Fatalf("player for a creature spec: got %v, want ErrIllegalTarget", err)
	}
	err = g.CastSpell(me.ID, blade.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: white}},
	})
	if err != nil {
		t.Fatalf("legal target rejected: %v", err)
	}
}

func TestCastSpellRejectsWrongTargetCount(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	a := pushColoredCreature(g, opp, "A", "{W}")
	b := pushColoredCreature(g, opp, "B", "{W}")
	const oracle = "test-single-target"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	spell := NewCard("Single", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: a}, {Kind: TargetCard, ID: b}},
	})
	if err != ErrInvalidParam {
		t.Fatalf("two targets for Max 1: got %v, want ErrInvalidParam", err)
	}
}

// TestResolutionRecheckUsesPredicate: the target was legal at
// announce and still exists at resolution, but no longer satisfies
// the predicate (it stopped being a creature) — CR 608.2b counters
// the spell by game rules.
func TestResolutionRecheckUsesPredicate(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for g.Turn.Step != StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	target := pushColoredCreature(g, opp, "Mutable", "{W}")
	const oracle = "test-recheck"
	withCatalogTargetSpec(t, func(id string) *TargetSpec {
		if id == oracle {
			return nonBlackCreatureSpec()
		}
		return nil
	})
	spell := NewCard("Recheck", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	me.Hand.PushTop(spell)
	if err := g.CastSpell(me.ID, spell.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	// "In response": the target becomes an artifact (loses creature
	// type). Still on the battlefield, so the existence check alone
	// would let the spell resolve.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == target {
				g.Battlefield.Cards[i].TypeLine = "Artifact"
			}
		}
	})
	g.WithWriteLock(func() {
		if err := g.resolveTopOfStackLocked(); err != nil {
			t.Fatalf("resolveTopOfStackLocked: %v", err)
		}
	})
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == EventFizzle && ev.CardID == spell.InstanceID {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("spell whose target failed the predicate at resolution was not countered by game rules")
	}
	if !me.Graveyard.Contains(spell.InstanceID) {
		t.Errorf("fizzled spell should be in its owner's graveyard")
	}
}

func TestSpecWithoutPredicateKeepsExistenceCheck(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src, _ := me.Hand.Top()
	// Free-form (nil spec) announce still accepts any card ID.
	if err := g.ActivateAbility(me.ID, src.InstanceID, AbilityParams{
		Label:   "free-form",
		Targets: []TargetRef{{Kind: TargetCard, ID: src.InstanceID}},
	}); err != nil {
		t.Fatalf("free-form ability with an arbitrary target: %v", err)
	}
}

// --- S20 sub-PR 2: targeted triggers ----------------------------

// artifactSpec is "target artifact an opponent controls".
func opponentArtifactSpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "permanent",
		Label: "target artifact an opponent controls",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, caster uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsArtifact() && c.Controller != caster
		},
		Min: 1, Max: 1,
	}
}

func pushArtifact(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Artifact"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// targetedETB registers a stub "when ~ enters, destroy target
// artifact an opponent controls" trigger under oracle; the Effect
// destroys item.Targets[0] and records it in *destroyed.
func targetedETB(t *testing.T, oracle string, optional bool, destroyed *[]uuid.UUID) {
	t.Helper()
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		ab := TriggeredAbility{
			Watches: []EventKind{EventETB},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: opponentArtifactSpec(),
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "destroy target artifact", func(g *Game, item *StackItem) error {
					if len(item.Targets) == 0 {
						return nil
					}
					*destroyed = append(*destroyed, item.Targets[0].ID)
					return g.DestroyPermanentForEffect(item.Targets[0].ID)
				})
			},
		}
		if optional {
			ab.OptionalPrompt = &TriggerOptionalPrompt{Question: "Destroy an artifact?"}
		}
		return []TriggeredAbility{ab}
	})
}

func findPickTarget(g *Game, chooser uuid.UUID) *PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoicePickTarget && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func TestTargetedTriggerPromptsPickThenBuildsWithTarget(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-targeted-etb"
	var destroyed []uuid.UUID
	targetedETB(t, oracle, false, &destroyed)
	rockA := pushArtifact(g, opp, "Rock A")
	rockB := pushArtifact(g, opp, "Rock B")
	mine := pushArtifact(g, me, "My Rock")
	srcID := seedTriggerSource(g, me, oracle)

	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: srcID, Actor: me.ID})
	})
	if len(g.PendingTriggers) != 0 || findAbilityOnStack(g, srcID) != nil {
		t.Fatalf("targeted trigger must not build before the target is chosen")
	}
	prompt := findPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt for the controller")
	}
	if prompt.Reason != "target artifact an opponent controls" {
		t.Errorf("prompt label = %q", prompt.Reason)
	}
	legal := map[uuid.UUID]bool{}
	for _, id := range prompt.PickTargetCards {
		legal[id] = true
	}
	if !legal[rockA] || !legal[rockB] || legal[mine] {
		t.Errorf("legal set = %v (want both opponent rocks, not mine)", prompt.PickTargetCards)
	}

	// Illegal pick (my own artifact) is rejected and the prompt stays.
	if err := g.ResolvePickTarget(prompt.ID, me.ID, TargetRef{Kind: TargetCard, ID: mine}); err != ErrIllegalTarget {
		t.Fatalf("own artifact: got %v, want ErrIllegalTarget", err)
	}
	if findPickTarget(g, me.ID) == nil {
		t.Fatalf("rejected pick must leave the prompt open")
	}
	if err := g.ResolvePickTarget(prompt.ID, opp.ID, TargetRef{Kind: TargetCard, ID: rockA}); err != ErrNotTheChooser {
		t.Fatalf("wrong chooser: got %v, want ErrNotTheChooser", err)
	}

	// Legal pick: item built with the target, on the stack.
	if err := g.ResolvePickTarget(prompt.ID, me.ID, TargetRef{Kind: TargetCard, ID: rockB}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	item := findAbilityOnStack(g, srcID)
	if item == nil {
		t.Fatalf("trigger not on the stack after the pick")
	}
	if len(item.Targets) != 1 || item.Targets[0].ID != rockB {
		t.Errorf("item targets = %+v, want rock B", item.Targets)
	}
	g.WithWriteLock(func() {
		if err := g.resolveTopAbilityLocked(); err != nil {
			t.Fatal(err)
		}
	})
	if len(destroyed) != 1 || destroyed[0] != rockB {
		t.Errorf("destroyed = %v, want [rockB]", destroyed)
	}
	if g.Battlefield.Contains(rockB) || !g.Battlefield.Contains(rockA) {
		t.Errorf("wrong artifact destroyed")
	}
}

func TestTargetedTriggerWithNoLegalTargetIsRemoved(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-targeted-etb-none"
	var destroyed []uuid.UUID
	targetedETB(t, oracle, true, &destroyed) // optional: still no yes/no when nothing is legal
	pushArtifact(g, me, "My Rock")           // only my own — not a legal target
	srcID := seedTriggerSource(g, me, oracle)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: srcID, Actor: me.ID})
	})
	if len(g.PendingChoices) != 0 || len(g.PendingTriggers) != 0 {
		t.Errorf("a targeted trigger with no legal target must be removed silently (CR 603.3d): choices=%d pending=%d",
			len(g.PendingChoices), len(g.PendingTriggers))
	}
}

func TestOptionalTargetedTriggerYesThenPick(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-targeted-etb-optional"
	var destroyed []uuid.UUID
	targetedETB(t, oracle, true, &destroyed)
	rock := pushArtifact(g, opp, "Rock")
	srcID := seedTriggerSource(g, me, oracle)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: srcID, Actor: me.ID})
	})
	// Yes/no first.
	var yn *PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == PendingChoiceTriggerPrompt {
			yn = c
		}
	}
	if yn == nil || findPickTarget(g, me.ID) != nil {
		t.Fatalf("optional targeted trigger must ask yes/no before the pick")
	}
	if err := g.ResolveTriggerPrompt(yn.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	prompt := findPickTarget(g, me.ID)
	if prompt == nil {
		t.Fatalf("no pick_target prompt after yes")
	}
	if err := g.ResolvePickTarget(prompt.ID, me.ID, TargetRef{Kind: TargetCard, ID: rock}); err != nil {
		t.Fatal(err)
	}
	if findAbilityOnStack(g, srcID) == nil {
		t.Errorf("trigger not on the stack after yes + pick")
	}
}

func TestTargetedTriggerRecheckFizzlesWhenTargetStopsQualifying(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-targeted-etb-recheck"
	var destroyed []uuid.UUID
	targetedETB(t, oracle, false, &destroyed)
	rock := pushArtifact(g, opp, "Rock")
	srcID := seedTriggerSource(g, me, oracle)
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventETB, CardID: srcID, Actor: me.ID})
	})
	prompt := findPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(prompt.ID, me.ID, TargetRef{Kind: TargetCard, ID: rock}); err != nil {
		t.Fatal(err)
	}
	// In response: the rock stops being an artifact (still on the
	// battlefield, so the old existence check would have let the
	// ability resolve).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == rock {
				g.Battlefield.Cards[i].TypeLine = "Enchantment"
			}
		}
		if err := g.resolveTopAbilityLocked(); err != nil {
			t.Fatal(err)
		}
	})
	if len(destroyed) != 0 {
		t.Errorf("effect ran on a target that no longer satisfies the spec")
	}
	var fizzled bool
	for _, ev := range g.Events {
		if ev.Kind == EventFizzle && ev.Source == srcID {
			fizzled = true
		}
	}
	if !fizzled {
		t.Errorf("no EventFizzle for the ability whose target stopped qualifying")
	}
}
