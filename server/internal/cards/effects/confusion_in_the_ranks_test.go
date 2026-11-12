package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// confusion_in_the_ranks_test.go — #5739: a trigger's target chosen by
// the entering permanent's controller, "another player controls" relative
// to that chooser, and "shares a card type with it" read off the entering
// permanent (Confusion in the Ranks, Mirrodin).

const confusionInTheRanksOracle = "e477abb9-9b99-48a8-8c09-07b120b72540"

// citrEnter puts a non-catalog permanent with the given type line from
// owner's hand onto the battlefield, so the enters triggers fire.
func citrEnter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	playerByIDForTest(g, owner).Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return id
}

// citrConfusion seeds a Confusion in the Ranks under controller.
func citrConfusion(g *game.Game, controller uuid.UUID) uuid.UUID {
	return pushPermanentForTest(g, controller, "Confusion in the Ranks", confusionInTheRanksOracle, "Enchantment")
}

// citrPickSet is the open pick_target prompt's card set.
func citrPickSet(t *testing.T, g *game.Game) (*game.PendingChoice, map[uuid.UUID]bool) {
	t.Helper()
	ch := pickTargetPrompt(t, g)
	set := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetCards {
		set[id] = true
	}
	return ch, set
}

// citrNoPrompt fails if a pick_target prompt is open.
func citrNoPrompt(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatalf("a pick_target prompt was opened: %+v", c)
		}
	}
}

// Each of the three trigger types, and a permanent of two: only another
// player's permanents that share a card type with the entering one are
// offered, the entering permanent's controller chooses, and the two
// permanents change hands.
func TestConfusionInTheRanksEachCardTypeOffersOnlyPermanentsSharingAType(t *testing.T) {
	cases := []struct {
		name, typeLine string
		// want lists which of the seeded permanents are offered.
		artifact, creature, enchantment, land, artifactLand bool
	}{
		{"Ornithopter", "Artifact", true, false, false, false, true},
		{"Bear", "Creature — Bear", false, true, false, false, false},
		{"Intruder Alarm", "Enchantment", false, false, true, false, false},
		{"Steel Wall", "Artifact Creature — Wall", true, true, false, false, true},
		{"Darksteel Citadel", "Artifact Land", true, false, false, true, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			owner, me, other := g.Seats[0], g.Seats[1], g.Seats[2]
			confusion := citrConfusion(g, owner.ID)
			seeded := map[string]uuid.UUID{
				"artifact":     b12Permanent(g, other.ID, "Sol Ring", "Artifact"),
				"creature":     b12Creature(g, other.ID, "Elk", "Creature — Elk", 3, 3),
				"enchantment":  b12Permanent(g, other.ID, "Glorious Anthem", "Enchantment"),
				"land":         b12Permanent(g, other.ID, "Forest", "Basic Land — Forest"),
				"artifactLand": b12Permanent(g, other.ID, "Seat of the Synod", "Artifact Land"),
			}
			// The chooser's own matching permanents are never offered.
			mine := b12Creature(g, me.ID, "Myr", "Artifact Creature — Myr", 1, 1)
			want := map[string]bool{"artifact": tc.artifact, "creature": tc.creature,
				"enchantment": tc.enchantment, "land": tc.land, "artifactLand": tc.artifactLand}

			entered := citrEnter(t, g, me.ID, tc.name, tc.typeLine)
			ch, set := citrPickSet(t, g)
			if ch.Chooser != me.ID || ch.PickTargetFor != owner.ID {
				t.Fatalf("prompt for %s on behalf of %s, want the entering permanent's controller for Confusion's controller", ch.Chooser, ch.PickTargetFor)
			}
			for k, id := range seeded {
				if set[id] != want[k] {
					t.Errorf("%s offered = %v, want %v", k, set[id], want[k])
				}
			}
			if set[mine] || set[entered] {
				t.Errorf("legal set %v offers a permanent the chooser controls", ch.PickTargetCards)
			}
			// Confusion is an enchantment another player controls.
			if set[confusion] != tc.enchantment {
				t.Errorf("Confusion offered = %v, want %v", set[confusion], tc.enchantment)
			}
			var pick uuid.UUID
			for _, k := range []string{"artifact", "creature", "enchantment", "land", "artifactLand"} {
				if want[k] && pick == uuid.Nil {
					pick = seeded[k]
				}
			}
			answerPickTarget(t, g, pick)
			if triggerOnStack(g, confusion) == nil {
				t.Fatal("the trigger is not on the stack after the target was chosen")
			}
			passPriorityAroundTable(t, g)
			if got := controllerOf(t, g, entered); got != other.ID {
				t.Errorf("the entering permanent: controller %s, want the other player %s", got, other.ID)
			}
			if got := controllerOf(t, g, pick); got != me.ID {
				t.Errorf("the chosen permanent: controller %s, want the chooser %s", got, me.ID)
			}
		})
	}
}

// Confusion's controller's own permanents are "another player's" when
// someone else chooses, and Confusion's controller cannot answer the
// prompt for them. Four players: only the chooser's permanents are left out.
func TestConfusionInTheRanksChooserIsTheEnteringPermanentsControllerAtAFourPlayerTable(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, west, east := g.Seats[0], g.Seats[2], g.Seats[1], g.Seats[3]
	confusion := citrConfusion(g, owner.ID)
	ownerBear := b12Creature(g, owner.ID, "Owner's Bear", "Creature — Bear", 2, 2)
	westElk := b12Creature(g, west.ID, "West's Elk", "Creature — Elk", 3, 3)
	eastOx := b12Creature(g, east.ID, "East's Ox", "Creature — Ox", 2, 4)
	myWolf := b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)

	entered := citrEnter(t, g, me.ID, "Hill Giant", "Creature — Giant")
	ch, set := citrPickSet(t, g)
	if ch.Chooser != me.ID || ch.PickTargetFor != owner.ID {
		t.Fatalf("prompt for %s on behalf of %s, want %s for %s", ch.Chooser, ch.PickTargetFor, me.ID, owner.ID)
	}
	if !set[ownerBear] || !set[westElk] || !set[eastOx] || set[myWolf] || set[entered] || len(set) != 3 {
		t.Fatalf("legal set %v, want the three other players' creatures", ch.PickTargetCards)
	}
	if err := g.ResolvePickTarget(ch.ID, owner.ID, game.TargetRef{Kind: game.TargetCard, ID: westElk}); !errors.Is(err, game.ErrNotTheChooser) {
		t.Fatalf("Confusion's controller answered the entering permanent's controller's prompt: %v", err)
	}
	answerPickTarget(t, g, ownerBear)
	if triggerOnStack(g, confusion) == nil {
		t.Fatal("the trigger is not on the stack")
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, entered); got != owner.ID {
		t.Errorf("Hill Giant: controller %s, want Confusion's controller %s", got, owner.ID)
	}
	if got := controllerOf(t, g, ownerBear); got != me.ID {
		t.Errorf("Owner's Bear: controller %s, want %s", got, me.ID)
	}
	if got := controllerOf(t, g, westElk); got != west.ID {
		t.Errorf("an unchosen creature changed hands: controller %s", got)
	}
}

// Confusion entering triggers itself (the 2004 ruling), and its own
// controller chooses another player's enchantment.
func TestConfusionInTheRanksTriggersOnItselfEntering(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	anthem := b12Permanent(g, opp.ID, "Glorious Anthem", "Enchantment")
	b12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Confusion in the Ranks", OracleID: confusionInTheRanksOracle,
		TypeLine: "Enchantment", ManaCost: "{3}{R}{R}", Owner: me.ID, Controller: me.ID})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter: %v", err)
	}
	ch, set := citrPickSet(t, g)
	if ch.Chooser != me.ID || ch.PickTargetFor != uuid.Nil {
		t.Fatalf("prompt for %s on behalf of %s, want the controller for itself", ch.Chooser, ch.PickTargetFor)
	}
	if !set[anthem] || len(set) != 1 {
		t.Fatalf("legal set %v, want only the opponent's enchantment", ch.PickTargetCards)
	}
	answerPickTarget(t, g, anthem)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, id); got != opp.ID {
		t.Errorf("Confusion: controller %s, want %s", got, opp.ID)
	}
	if got := controllerOf(t, g, anthem); got != me.ID {
		t.Errorf("the enchantment: controller %s, want %s", got, me.ID)
	}
}

// With no permanent another player controls that shares a type, the
// trigger is removed (CR 603.3d) and nobody is asked; a land entering
// does not trigger at all.
func TestConfusionInTheRanksWithNoLegalTargetDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, opp := g.Seats[0], g.Seats[1], g.Seats[2]
	confusion := citrConfusion(g, owner.ID)
	b12Permanent(g, opp.ID, "Sol Ring", "Artifact")
	b12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	hexed := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Slippery Bogle",
		TypeLine: "Creature — Beast", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID,
		Keywords: []string{"hexproof"}})

	entered := citrEnter(t, g, me.ID, "Bear", "Creature — Bear")
	citrNoPrompt(t, g)
	if triggerOnStack(g, confusion) != nil {
		t.Fatal("the trigger reached the stack with no legal target (hexproof is judged for Confusion's controller)")
	}
	if got := controllerOf(t, g, entered); got != me.ID {
		t.Errorf("the entering creature changed hands: %s", got)
	}
	if got := controllerOf(t, g, hexed); got != opp.ID {
		t.Errorf("the hexproof creature changed hands: %s", got)
	}

	citrEnter(t, g, me.ID, "Forest", "Basic Land — Forest")
	citrNoPrompt(t, g)
	if triggerOnStack(g, confusion) != nil {
		t.Error("a land entering triggered Confusion")
	}
}

// The entering permanent leaves before the ability resolves: there is no
// exchange (CR 701.12a, the 2004 ruling), and the chosen permanent stays
// put.
func TestConfusionInTheRanksNoExchangeWhenTheEnteringPermanentLeft(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, opp := g.Seats[0], g.Seats[1], g.Seats[2]
	citrConfusion(g, owner.ID)
	elk := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	entered := citrEnter(t, g, me.ID, "Bear", "Creature — Bear")
	answerPickTarget(t, g, elk)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(entered) })
	if _, ok := battlefieldCard(g, entered); ok {
		t.Fatal("setup: the bear is still on the battlefield")
	}
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, elk); got != opp.ID {
		t.Errorf("the chosen creature changed hands in a half-exchange: controller %s, want %s", got, opp.ID)
	}
}

// The types are checked again as the ability resolves (the 2004 ruling,
// CR 608.2b): an entering artifact creature that stops being a creature
// no longer shares a type with the chosen creature, so nothing changes.
func TestConfusionInTheRanksRechecksTheSharedTypeOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, opp := g.Seats[0], g.Seats[1], g.Seats[2]
	citrConfusion(g, owner.ID)
	elk := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	entered := citrEnter(t, g, me.ID, "Steel Wall", "Artifact Creature — Wall")
	answerPickTarget(t, g, elk)
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, entered)
		c.TypeLine = "Artifact — Wall"
		g.BumpLayerVersionForTest()
	})
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, elk); got != opp.ID {
		t.Errorf("the creature changed hands though the entering permanent is no longer a creature: %s", got)
	}
	if got := controllerOf(t, g, entered); got != me.ID {
		t.Errorf("the entering permanent changed hands: %s", got)
	}
}

// A restore taken while the entering permanent's controller's prompt is
// open keeps the prompt with its chooser, its legal set and whose ability
// it is, and answering it on the restored game makes the exchange.
func TestConfusionInTheRanksPromptSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	owner, me, opp := g.Seats[0], g.Seats[3], g.Seats[1]
	confusion := citrConfusion(g, owner.ID)
	ownerBear := b12Creature(g, owner.ID, "Owner's Bear", "Creature — Bear", 2, 2)
	elk := b12Creature(g, opp.ID, "Elk", "Creature — Elk", 3, 3)
	entered := citrEnter(t, g, me.ID, "Hill Giant", "Creature — Giant")
	before, beforeSet := citrPickSet(t, g)

	restored := restoreRoundTrip(t, g, true)
	ch, set := citrPickSet(t, restored)
	if ch.Chooser != me.ID || ch.PickTargetFor != owner.ID || ch.Chooser != before.Chooser {
		t.Fatalf("restored prompt for %s on behalf of %s, want %s for %s", ch.Chooser, ch.PickTargetFor, me.ID, owner.ID)
	}
	if len(set) != len(beforeSet) || !set[ownerBear] || !set[elk] {
		t.Fatalf("restored legal set %v, want %v", ch.PickTargetCards, before.PickTargetCards)
	}
	answerPickTarget(t, restored, elk)
	if triggerOnStack(restored, confusion) == nil {
		t.Fatal("restored: the trigger is not on the stack after the target was chosen")
	}
	// A second restore with the trigger on the stack: the CR 608.2b
	// re-check reads the same chooser off the carried event.
	again := restoreRoundTrip(t, restored, true)
	passPriorityAroundTable(t, again)
	if got := controllerOf(t, again, entered); got != opp.ID {
		t.Errorf("Hill Giant: controller %s, want %s", got, opp.ID)
	}
	if got := controllerOf(t, again, elk); got != me.ID {
		t.Errorf("Elk: controller %s, want %s", got, me.ID)
	}
	if !slices.Contains(before.PickTargetCards, ownerBear) {
		t.Error("Confusion's controller's own creature was not offered to another chooser")
	}
}
