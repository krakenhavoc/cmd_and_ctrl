package game

import (
	"testing"

	"github.com/google/uuid"
)

// target_ability_test.go — CR 115.4 (#1211): a target clause that can
// name an ACTIVATED or TRIGGERED ability on the stack, beside or
// instead of a spell.
//
// The rules pinned here are the ones a reader of targets.go's
// Abilities / AbilityOK pair and of counterAbilityLocked would want
// proved: one clause enumerates both halves of the stack in a stable
// order, a picked ability is an ordinary card-kind ref, the CR 608.2b
// re-check follows the item rather than a zone, and countering one is
// a deletion that leaves its source permanent standing.

const (
	stackItemOracle    = "test-target-spell-or-ability"
	abilityOnlyOracle  = "test-target-ability-only"
	spellOnlyOracle    = "test-target-spell-only"
	plainSpellOracle   = "test-plain-untargeted-spell"
	stackTargetedLabel = "a targeted ability"
)

// spellOrAbilitySpec is Disallow's clause: the stack zone AND the
// ability items, one pick.
func spellOrAbilitySpec() *TargetSpec {
	return &TargetSpec{
		Mode:      "stack_item",
		Label:     "target spell, activated ability, or triggered ability",
		Zones:     []ZoneKind{ZoneStack},
		Abilities: true,
		Min:       1, Max: 1,
	}
}

// abilityOnlySpec is Stifle's clause: no zones at all, so nothing in
// the stack ZONE is a candidate and only StackMeta is walked.
func abilityOnlySpec() *TargetSpec {
	return &TargetSpec{
		Mode:      "stack_ability",
		Label:     "target activated or triggered ability",
		Abilities: true,
		Min:       1, Max: 1,
	}
}

// spellOnlySpec is Counterspell's clause, unchanged by #1211 — proof
// that widening the vocabulary did not widen the old clause.
func spellOnlySpec() *TargetSpec {
	return &TargetSpec{
		Mode:  "stack_spell",
		Label: "target spell",
		Zones: []ZoneKind{ZoneStack},
		Min:   1, Max: 1,
	}
}

func stackTargetSpecs(t *testing.T) {
	t.Helper()
	withCatalogTargetSpec(t, func(oracleID string) *TargetSpec {
		switch oracleID {
		case stackItemOracle:
			return spellOrAbilitySpec()
		case abilityOnlyOracle:
			return abilityOnlySpec()
		case spellOnlyOracle:
			return spellOnlySpec()
		case retargetBoltOracle:
			return anyTargetSpec()
		}
		return nil
	})
}

// pushStackTargetPermanent puts a plain permanent on the battlefield
// to be an ability's source.
func pushStackTargetPermanent(g *Game, owner *Player, name string) uuid.UUID {
	c := NewCard(name, owner.ID)
	c.TypeLine = "Artifact"
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

// announceAbility puts a triggered ability item on the stack and
// returns it.
func announceAbility(t *testing.T, g *Game, controller *Player, source uuid.UUID, label string) *StackItem {
	t.Helper()
	before := make(map[uuid.UUID]bool, len(g.StackMeta))
	for id := range g.StackMeta {
		before[id] = true
	}
	if err := g.AnnounceTrigger(controller.ID, source, AbilityParams{Label: label}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	for id, item := range g.StackMeta {
		if !before[id] && item != nil && item.Kind == StackItemTriggered {
			return item
		}
	}
	t.Fatalf("the announced trigger is not on the stack")
	return nil
}

// A "target spell or ability" clause enumerates BOTH halves of the
// stack: the spell cards in the zone and the ability items in
// StackMeta, with the cards first and the abilities in Seq order.
func TestASpellOrAbilityClauseEnumeratesBothHalvesOfTheStack(t *testing.T) {
	g := newActiveGame(t)
	stackTargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")

	// An untargeted spell, so the stack has a CARD on it that is not
	// this clause's own announcement.
	spell := castTargeted(t, g, opp, plainSpellOracle, nil)
	first := announceAbility(t, g, opp, src, "first ability")
	second := announceAbility(t, g, opp, src, "second ability")

	var lt LegalTargets
	g.WithWriteLock(func() {
		lt = g.legalTargetsLocked(SourceChooser(me.ID), spellOrAbilitySpec())
	})
	if len(lt.Cards) != 3 {
		t.Fatalf("legal targets = %v, want the spell and both abilities", lt.Cards)
	}
	if lt.Cards[0] != spell.ID {
		t.Errorf("the zone walk runs first: %v", lt.Cards)
	}
	// Seq order, which is the stack's order — a legal-target list is
	// public output and two runs of one game must not disagree.
	if lt.Cards[1] != first.ID || lt.Cards[2] != second.ID {
		t.Errorf("abilities out of Seq order: %v, want %v then %v", lt.Cards, first.ID, second.ID)
	}

	// The narrow clauses each see exactly one half.
	var spellsOnly, abilitiesOnly LegalTargets
	g.WithWriteLock(func() {
		spellsOnly = g.legalTargetsLocked(SourceChooser(me.ID), spellOnlySpec())
		abilitiesOnly = g.legalTargetsLocked(SourceChooser(me.ID), abilityOnlySpec())
	})
	if len(spellsOnly.Cards) != 1 || spellsOnly.Cards[0] != spell.ID {
		t.Errorf("target spell = %v, want only the spell", spellsOnly.Cards)
	}
	if len(abilitiesOnly.Cards) != 2 || hasUUID(abilitiesOnly.Cards, spell.ID) {
		t.Errorf("target ability = %v, want only the two abilities", abilitiesOnly.Cards)
	}
}

// A picked ability is announced as an ordinary TargetRef{Kind:
// TargetCard} carrying the STACK ITEM's id, and the announce gate
// (CR 601.2c) accepts it. The same gate refuses an ability for a
// spell-only clause, which is what keeps Counterspell off a trigger.
func TestAnAbilityIsAnnouncedAsAnOrdinaryCardRef(t *testing.T) {
	g := newActiveGame(t)
	stackTargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	ability := announceAbility(t, g, opp, src, "their trigger")

	item := castTargeted(t, g, me, stackItemOracle,
		[]TargetRef{{Kind: TargetCard, ID: ability.ID}})
	if len(item.Targets) != 1 || item.Targets[0].Kind != TargetCard || item.Targets[0].ID != ability.ID {
		t.Fatalf("announced targets = %+v, want a card-kind ref carrying the item id", item.Targets)
	}

	// The spell-only clause refuses the same ref.
	c := NewCard("Counterspell", me.ID)
	c.TypeLine, c.ManaCost, c.OracleID = "Instant", "{0}", spellOnlyOracle
	c.Controller = me.ID
	me.Hand.PushTop(c)
	err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: ability.ID}},
	})
	if err != ErrIllegalTarget {
		t.Errorf("a spell-only clause accepted an ability item: %v", err)
	}
}

// CR 701.5c: an ability that is countered does not GO anywhere. The
// item leaves StackMeta, its source permanent is untouched, and
// nothing lands in a graveyard.
func TestCounteringAnAbilityDeletesTheItemAndLeavesItsSourceAlone(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	opp := g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	ability := announceAbility(t, g, opp, src, stackTargetedLabel)

	if err := g.CounterAbility(ability.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	if g.StackMeta[ability.ID] != nil {
		t.Error("the countered ability is still on the stack")
	}
	if !g.Battlefield.Contains(src) {
		t.Error("the ability's SOURCE moved — a countered ability goes nowhere")
	}
	if opp.Graveyard.Contains(ability.ID) || opp.Graveyard.Contains(src) {
		t.Error("something reached a graveyard; CR 701.5c says nothing does")
	}
	if g.Stack != nil && g.Stack.Contains(ability.ID) {
		t.Error("an ability item has no card and must never be in the stack zone")
	}
}

// The counter's event says what events.go documents it says: Source
// is THE COUNTER (unset here — the emit site does not know it), Target
// is the countered item, and the ability's own source rides CardID
// with the item's label beside it. Before #1211 the source permanent
// sat in Source, which the game log reads as "who countered it".
func TestACounteredAbilitysEventNamesTheItemNotItsSource(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	opp := g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	ability := announceAbility(t, g, opp, src, stackTargetedLabel)

	before := len(g.Events)
	if err := g.CounterAbility(ability.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	var ev *Event
	for i := before; i < len(g.Events); i++ {
		if g.Events[i].Kind == EventCounterSpell {
			ev = &g.Events[i]
		}
	}
	if ev == nil {
		t.Fatal("no EventCounterSpell for a countered ability")
	}
	if ev.Source != uuid.Nil {
		t.Errorf("Source = %v, want unset: it is reserved for the COUNTER", ev.Source)
	}
	if ev.Target != ability.ID {
		t.Errorf("Target = %v, want the countered item %v", ev.Target, ability.ID)
	}
	if ev.CardID != src {
		t.Errorf("CardID = %v, want the permanent whose ability it was %v", ev.CardID, src)
	}
	if ev.Label != stackTargetedLabel {
		t.Errorf("Label = %q, want the item's label", ev.Label)
	}
}

// CR 608.2b: a counterspell aimed at an ability that has already left
// the stack — resolved, or countered in response — fizzles. The
// re-check follows the ITEM, which is the only place an ability
// exists.
func TestASpellTargetingAnAbilityFizzlesWhenTheAbilityLeaves(t *testing.T) {
	g := newActiveGame(t)
	stackTargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	ability := announceAbility(t, g, opp, src, "their trigger")

	item := castTargeted(t, g, me, stackItemOracle,
		[]TargetRef{{Kind: TargetCard, ID: ability.ID}})

	var legalBefore, legalAfter bool
	g.WithWriteLock(func() {
		legalBefore = g.TargetStillLegalForEffect(item, item.Targets[0])
	})
	if !legalBefore {
		t.Fatal("the ability is on the stack and must still be a legal target")
	}
	if err := g.CounterAbility(ability.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	g.WithWriteLock(func() {
		legalAfter = g.TargetStillLegalForEffect(item, item.Targets[0])
	})
	if legalAfter {
		t.Error("an ability that has left the stack is not a legal target")
	}
}

// The existence-only fallback (a free-form announcement with no
// clause behind it) knows an ability item too. Without this an
// engine-level re-check would call every ability target illegal the
// moment it looked, because findCardZoneLocked cannot find a card
// that does not exist.
func TestTheExistenceFallbackFindsAnAbilityItem(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	opp := g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	ability := announceAbility(t, g, opp, src, stackTargetedLabel)

	ref := TargetRef{Kind: TargetCard, ID: ability.ID}
	var found, gone bool
	g.WithWriteLock(func() { found = targetStillExistsLocked(g, ref) })
	if !found {
		t.Error("an ability on the stack exists")
	}
	if err := g.CounterAbility(ability.ID); err != nil {
		t.Fatalf("CounterAbility: %v", err)
	}
	g.WithWriteLock(func() { gone = targetStillExistsLocked(g, ref) })
	if gone {
		t.Error("an ability that left the stack does not exist")
	}
}

// CR 115.7 over an ability: the retarget entry point (#1196) is
// written over StackMeta, so it takes an ability item with no code of
// its own. This is the half Deflecting Swat's and Bolt Bend's caveats
// were waiting on.
func TestAnAbilityOnTheStackCanBeRetargeted(t *testing.T) {
	g := newActiveGame(t)
	stackTargetSpecs(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushStackTargetPermanent(g, opp, "Their Rock")
	first := pushRetargetCreature(g, opp, "First")
	second := pushRetargetCreature(g, opp, "Second")

	if err := g.AnnounceTrigger(opp.ID, src, AbilityParams{
		Label:   "Their Rock — deal 2 damage to any target",
		Targets: []TargetRef{{Kind: TargetCard, ID: first}},
	}); err != nil {
		t.Fatalf("AnnounceTrigger: %v", err)
	}
	var ability *StackItem
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == StackItemTriggered {
			ability = it
		}
	}
	if ability == nil {
		t.Fatal("the trigger is not on the stack")
	}
	// An ability item carries its own clause; nothing is looked up by
	// oracle id for it.
	ability.targetSpec = anyTargetSpec()

	var count int
	g.WithWriteLock(func() { count = g.StackItemTargetCountForEffect(ability.ID) })
	if count != 1 {
		t.Fatalf("an ability's target count = %d, want 1 (Bolt Bend's clause)", count)
	}

	g.WithWriteLock(func() {
		if err := g.OfferRetargetForEffect(RetargetOffer{
			ItemID:  ability.ID,
			Chooser: me.ID,
			Policy:  RetargetChangeOne,
			Reason:  "Bolt Bend — change the target",
		}); err != nil {
			t.Fatalf("OfferRetargetForEffect: %v", err)
		}
	})
	prompt := findRetargetPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no retarget prompt for an ability item")
	}
	if !hasUUID(prompt.PickTargetCards, second) || hasUUID(prompt.PickTargetCards, first) {
		t.Fatalf("alternatives = %v, want the other creature and not the current target", prompt.PickTargetCards)
	}
	if err := g.ResolveRetarget(prompt.ID, me.ID,
		[]TargetRef{{Kind: TargetCard, ID: second}}); err != nil {
		t.Fatalf("ResolveRetarget: %v", err)
	}
	if got := g.StackMeta[ability.ID].Targets; len(got) != 1 || got[0].ID != second {
		t.Fatalf("the ability's targets = %+v, want the second creature", got)
	}
}

// An ability is neither a permanent nor a player, so the CR 702
// keyword gate has no subject and is not run — a clause that admits
// abilities offers one whose SOURCE has hexproof, while the same
// source is off a battlefield clause. Stated as a test because the
// omission is deliberate and a future reader will want to know it was
// not forgotten.
func TestTheKeywordGateDoesNotApplyToAnAbilityItem(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	c := NewCard("Hexproof Rock", opp.ID)
	c.TypeLine = "Artifact Creature — Golem"
	c.Power, c.Toughness = 2, 2
	c.Keywords = append(c.Keywords, "hexproof")
	g.Battlefield.PushTop(c)
	ability := announceAbility(t, g, opp, c.InstanceID, "the hexproof rock's trigger")

	var abilities, permanents LegalTargets
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		abilities = g.legalTargetsLocked(SourceChooser(me.ID), abilityOnlySpec())
		permanents = g.legalTargetsLocked(SourceChooser(me.ID), oneCreatureSpec())
	})
	if !hasUUID(abilities.Cards, ability.ID) {
		t.Error("a hexproof SOURCE does not shield its ability on the stack")
	}
	if hasUUID(permanents.Cards, c.InstanceID) {
		t.Error("the same permanent is still hexproof for a battlefield clause")
	}
}
