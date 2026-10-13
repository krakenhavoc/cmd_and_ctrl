package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestCrimeAnnotationSurvivesCloneUndoAndDiskRestore(t *testing.T) {
	g := newActiveGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	announce := func() {
		g.WithWriteLock(func() {
			g.emitAnnouncedTargetsLocked(me.ID, uuid.New(), uuid.New(), []TargetRef{{Kind: TargetPlayer, ID: opponent.ID}})
		})
	}
	announce()
	clone := g.Clone()
	announce()
	if countCrimeEvents(clone.Events) != 1 {
		t.Fatal("a later announcement changed the clone's crime history")
	}
	g.RestoreFrom(clone)
	if countCrimeEvents(g.Events) != 1 {
		t.Fatal("undo did not restore the annotated announcement")
	}
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatal(err)
	}
	var snapshot GameSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	restored, err := snapshot.RestoreStrict()
	if err != nil {
		t.Fatal(err)
	}
	if countCrimeEvents(restored.Events) != 1 {
		t.Fatal("disk restore lost the crime annotation")
	}
}

func crimeTestCard(g *Game, zone *Zone, name string, owner, controller uuid.UUID) uuid.UUID {
	id := uuid.New()
	zone.PushTop(Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", Owner: owner, Controller: controller})
	return id
}

func countCrimeEvents(events []Event) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == EventBecomesTarget && ev.CrimeCommitted {
			n++
		}
	}
	return n
}

func countEventKind(events []Event, kind EventKind) int {
	n := 0
	for _, ev := range events {
		if ev.Kind == kind {
			n++
		}
	}
	return n
}

func TestCrimeTargetClassification(t *testing.T) {
	g := newActiveGame(t)
	actor, opponent := g.Seats[0], g.Seats[1]
	ownPermanent := crimeTestCard(g, g.Battlefield, "Own permanent", actor.ID, actor.ID)
	opponentPermanent := crimeTestCard(g, g.Battlefield, "Opponent permanent", opponent.ID, opponent.ID)
	stolenPermanent := crimeTestCard(g, g.Battlefield, "Stolen permanent", actor.ID, opponent.ID)
	opponentControlledByActor := crimeTestCard(g, g.Battlefield, "Opponent-owned permanent", opponent.ID, actor.ID)
	ownHand := crimeTestCard(g, actor.Hand, "Own hand", actor.ID, actor.ID)
	opponentExile := crimeTestCard(g, g.Exile, "Opponent exile", opponent.ID, opponent.ID)
	ownGraveyard := crimeTestCard(g, actor.Graveyard, "Own graveyard", actor.ID, actor.ID)
	opponentGraveyard := crimeTestCard(g, opponent.Graveyard, "Opponent graveyard", opponent.ID, opponent.ID)
	// Graveyard ownership, unlike battlefield control, determines the
	// opponent relationship for a card there.
	for i := range opponent.Graveyard.Cards {
		if opponent.Graveyard.Cards[i].InstanceID == opponentGraveyard {
			opponent.Graveyard.Cards[i].Controller = actor.ID
		}
	}
	spell := crimeTestCard(g, g.Stack, "Opponent spell", opponent.ID, opponent.ID)
	ability := uuid.New() // An ability is a stack object, not a card in a zone.
	source := crimeTestCard(g, g.Battlefield, "Actor's source", actor.ID, actor.ID)
	opponentSource := crimeTestCard(g, g.Battlefield, "Opponent's source", opponent.ID, opponent.ID)
	g.StackMeta = make(map[uuid.UUID]*StackItem)
	g.StackMeta[spell] = &StackItem{ID: spell, Kind: StackItemSpell, Controller: opponent.ID}
	ownSpell := crimeTestCard(g, g.Stack, "Opponent-owned spell controlled by actor", opponent.ID, actor.ID)
	g.StackMeta[ownSpell] = &StackItem{ID: ownSpell, Kind: StackItemSpell, Controller: actor.ID}
	g.StackMeta[ability] = &StackItem{ID: ability, Kind: StackItemActivated, Controller: opponent.ID, SourceCardID: source}
	ownAbility := uuid.New()
	g.StackMeta[ownAbility] = &StackItem{ID: ownAbility, Kind: StackItemActivated, Controller: actor.ID, SourceCardID: opponentSource}

	cases := []struct {
		name string
		ref  TargetRef
		want bool
	}{
		{"opponent player", TargetRef{Kind: TargetPlayer, ID: opponent.ID}, true},
		{"self player", TargetRef{Kind: TargetPlayer, ID: actor.ID}, false},
		{"opponent permanent", TargetRef{Kind: TargetCard, ID: opponentPermanent}, true},
		{"own permanent", TargetRef{Kind: TargetCard, ID: ownPermanent}, false},
		{"stolen permanent", TargetRef{Kind: TargetCard, ID: stolenPermanent}, true},
		{"opponent-owned permanent controlled by actor", TargetRef{Kind: TargetCard, ID: opponentControlledByActor}, false},
		{"opponent spell", TargetRef{Kind: TargetCard, ID: spell}, true},
		{"opponent-owned spell controlled by actor", TargetRef{Kind: TargetCard, ID: ownSpell}, false},
		{"opponent ability with own source", TargetRef{Kind: TargetCard, ID: ability}, true},
		{"own ability with opponent source", TargetRef{Kind: TargetCard, ID: ownAbility}, false},
		{"own hand", TargetRef{Kind: TargetCard, ID: ownHand}, false},
		{"opponent exile", TargetRef{Kind: TargetCard, ID: opponentExile}, false},
		{"own graveyard", TargetRef{Kind: TargetCard, ID: ownGraveyard}, false},
		{"opponent graveyard", TargetRef{Kind: TargetCard, ID: opponentGraveyard}, true},
		{"source itself", TargetRef{Kind: TargetCard, ID: source}, false},
		{"missing object", TargetRef{Kind: TargetCard, ID: uuid.New()}, false},
		{"no target", TargetRef{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var got bool
			g.WithWriteLock(func() { got = g.crimeTargetLocked(actor.ID, tc.ref) })
			if got != tc.want {
				t.Errorf("crimeTargetLocked(%+v) = %v, want %v", tc.ref, got, tc.want)
			}
		})
	}
}

func TestOneCrimePerTargetAnnouncementAndSeparateAnnouncements(t *testing.T) {
	g := newActiveGame(t)
	actor, opponent := g.Seats[0], g.Seats[1]
	first := crimeTestCard(g, g.Battlefield, "First", opponent.ID, opponent.ID)
	second := crimeTestCard(g, g.Battlefield, "Second", opponent.ID, opponent.ID)
	targets := []TargetRef{
		{Kind: TargetCard, ID: first},
		{Kind: TargetPlayer, ID: opponent.ID},
		{Kind: TargetCard, ID: second},
	}
	start := len(g.Events)
	g.WithWriteLock(func() {
		g.emitAnnouncedTargetsLocked(actor.ID, uuid.New(), uuid.New(), targets)
	})
	if got := countCrimeEvents(g.Events[start:]); got != 1 {
		t.Fatalf("announcement with three qualifying targets emitted %d crimes, want one", got)
	}
	if got := countEventKind(g.Events[start:], EventBecomesTarget); got != len(targets) {
		t.Fatalf("becomes-target events = %d, want %d", got, len(targets))
	}
	g.WithWriteLock(func() {
		g.emitAnnouncedTargetsLocked(actor.ID, uuid.New(), uuid.New(), targets[:1])
	})
	if got := countCrimeEvents(g.Events[start:]); got != 2 {
		t.Fatalf("two separate announcements emitted %d crimes, want two", got)
	}
}

func TestCrimeEventsAcrossAnnouncementEntryPoints(t *testing.T) {
	for _, path := range []string{"cast", "sandbox activation", "catalog activation", "manual trigger"} {
		t.Run(path, func(t *testing.T) {
			g := newActiveGame(t)
			advanceTo(t, g, StepPrecombatMain)
			me, opponent := g.Seats[0], g.Seats[1]
			target := TargetRef{Kind: TargetPlayer, ID: opponent.ID}
			switch path {
			case "cast":
				retargetSpecs(t)
				castTargeted(t, g, me, retargetBoltOracle, []TargetRef{target})
			case "sandbox activation":
				source := pushPermanent(g, NewCard("Manual source", me.ID))
				if err := g.ActivateAbility(me.ID, source, AbilityParams{Label: "manual", Targets: []TargetRef{target}}); err != nil {
					t.Fatalf("ActivateAbility: %v", err)
				}
			case "catalog activation":
				const oracle = "test-crime-catalog-activation"
				prev := CatalogActivatedAbilities
				CatalogActivatedAbilities = func(id string) []ActivatedAbilityShape {
					if id == oracle {
						return []ActivatedAbilityShape{{Label: "target opponent", Targets: anyTargetSpec()}}
					}
					if prev != nil {
						return prev(id)
					}
					return nil
				}
				t.Cleanup(func() { CatalogActivatedAbilities = prev })
				source := pushBattlefieldForTest(g, me.ID, "Catalog source", "Artifact", oracle)
				if err := g.ActivateCatalogAbility(me.ID, source, 0, ActivateAbilityParams{Targets: []TargetRef{target}}); err != nil {
					t.Fatalf("ActivateCatalogAbility: %v", err)
				}
			case "manual trigger":
				source := pushPermanent(g, NewCard("Manual trigger source", me.ID))
				if err := g.AnnounceTrigger(me.ID, source, AbilityParams{Label: "manual trigger", Targets: []TargetRef{target}}); err != nil {
					t.Fatalf("AnnounceTrigger: %v", err)
				}
			}
			if got := countCrimeEvents(g.Events); got != 1 {
				t.Errorf("%s emitted %d crimes, want one", path, got)
			}
		})
	}
}

func TestHarvestedTriggerTargetPlacementCanCommitCrime(t *testing.T) {
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	me, opponent := g.Seats[0], g.Seats[1]
	const watcherOracle = "test-crime-targeted-trigger"
	pushBattlefieldForTest(g, me.ID, "Targeted trigger source", "Enchantment", watcherOracle)
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != watcherOracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches:   []EventKind{EventCast},
			AppliesTo: func(Event, *Card, Characteristic, *Game) bool { return true },
			Targets: &TargetSpec{
				Mode: "player", Players: true, Min: 1, Max: 1,
			},
			Build: func(_ Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return newTriggeredItemForTest(source, "Test targeted trigger", func(*Game, *StackItem) error { return nil })
			},
		}}
	})
	retargetSpecs(t)
	// The spell's own target is its controller, so only the trigger's
	// later choice can produce the qualifying opponent target.
	castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetPlayer, ID: me.ID}})
	choice := openPickTarget(g, me.ID)
	if choice == nil {
		t.Fatal("targeted cast trigger did not open its target choice")
	}
	if err := g.ResolvePickTarget(choice.ID, me.ID, TargetRef{Kind: TargetPlayer, ID: opponent.ID}); err != nil {
		t.Fatalf("ResolvePickTarget for harvested trigger: %v", err)
	}
	if got := countCrimeEvents(g.Events); got != 1 {
		t.Errorf("placing the targeted trigger emitted %d crimes, want one", got)
	}
}

func TestCopyAndRetargetDoNotCommitAnotherCrime(t *testing.T) {
	t.Run("copy", func(t *testing.T) {
		g := newActiveGame(t)
		advanceTo(t, g, StepPrecombatMain)
		me, opponent := g.Seats[0], g.Seats[1]
		retargetSpecs(t)
		item := castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetPlayer, ID: opponent.ID}})
		if got := countCrimeEvents(g.Events); got != 1 {
			t.Fatalf("cast emitted %d crime events, want one", got)
		}
		becomes := countEventKind(g.Events, EventBecomesTarget)
		g.WithWriteLock(func() {
			if err := g.CopySpellForEffect(item.ID, me.ID, false, nil); err != nil {
				t.Errorf("CopySpellForEffect: %v", err)
			}
		})
		if got := countCrimeEvents(g.Events); got != 1 {
			t.Errorf("copy emitted a new crime; got %d total, want one", got)
		}
		if got := countEventKind(g.Events, EventBecomesTarget); got <= becomes {
			t.Errorf("copy emitted no becomes-target event (before %d, after %d)", becomes, got)
		}
	})

	t.Run("retarget", func(t *testing.T) {
		g := newActiveGameWithSeats(t, 3)
		advanceTo(t, g, StepPrecombatMain)
		me, opponent, secondOpponent := g.Seats[0], g.Seats[1], g.Seats[2]
		retargetSpecs(t)
		item := castTargeted(t, g, me, retargetBoltOracle, []TargetRef{{Kind: TargetPlayer, ID: opponent.ID}})
		becomes := countEventKind(g.Events, EventBecomesTarget)
		g.WithWriteLock(func() {
			if err := g.RetargetStackItemForEffect(item.ID, me.ID, RetargetChangeOne, []TargetRef{{Kind: TargetPlayer, ID: secondOpponent.ID}}); err != nil {
				t.Errorf("RetargetStackItemForEffect: %v", err)
			}
		})
		if got := countCrimeEvents(g.Events); got != 1 {
			t.Errorf("retarget emitted a new crime; got %d total, want one", got)
		}
		if got := countEventKind(g.Events, EventBecomesTarget); got <= becomes {
			t.Errorf("retarget emitted no becomes-target event (before %d, after %d)", becomes, got)
		}
	})
}

func TestCopyingTargetedAbilityDoesNotCommitCrime(t *testing.T) {
	g := newActiveGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	source := pushPermanent(g, NewCard("Ability source", me.ID))
	item := pushAbilityItem(g, StackItemActivated, me.ID, source, "target an opponent", nil)
	item.Targets = []TargetRef{{Kind: TargetPlayer, ID: opponent.ID}}
	becomes := countEventKind(g.Events, EventBecomesTarget)

	g.WithWriteLock(func() {
		if err := g.CopyAbilityForEffect(item.ID, me.ID, false); err != nil {
			t.Errorf("CopyAbilityForEffect: %v", err)
		}
	})
	if got := countCrimeEvents(g.Events); got != 0 {
		t.Errorf("copying an ability emitted %d crime events, want none", got)
	}
	if got := countEventKind(g.Events, EventBecomesTarget); got <= becomes {
		t.Errorf("ability copy emitted no ordinary becomes-target event (before %d, after %d)", becomes, got)
	}
}
