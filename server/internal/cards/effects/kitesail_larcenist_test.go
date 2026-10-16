package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kitesailLarcenistOracle = "2452be47-cc23-47f7-a3a1-fec900bb0119"

// Flying is a printed keyword and reaches the effective
// characteristic like any other.
func TestKitesailLarcenistFlies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := pushCatalogPermanent(g, me.ID, "Kitesail Larcenist",
		"Creature — Human Pirate", kitesailLarcenistOracle, false)
	if !hasAbility(effectiveAbilities(t, g, id), "flying") {
		t.Errorf("Kitesail Larcenist has flying, got %v", effectiveAbilities(t, g, id))
	}
}

// Ward {1} is a triggered ability (CR 702.21a), so an opponent's spell
// is a LEGAL target and the tax lands on the caster as a payment
// prompt once the ward trigger resolves.
func TestKitesailLarcenistTaxesAnOpponentsSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	larcenist := pushCatalogPermanent(g, me.ID, "Kitesail Larcenist",
		"Creature — Human Pirate", kitesailLarcenistOracle, false)

	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	id := uuid.New()
	opp.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: opp.ID, Controller: opp.ID,
	})
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: larcenist}},
	}); err != nil {
		t.Fatalf("a warded permanent is a legal target: %v", err)
	}
	for i := 0; i < 8 && !hasPayUnlessFor(g, opp.ID); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !hasPayUnlessFor(g, opp.ID) {
		t.Fatalf("ward {1} asks the CASTER to pay: %+v", g.PendingChoices)
	}
}

// kitesailAnswerPicks answers the enters trigger's clauses one by one:
// each takes the first of `want` it offers, or nothing ("up to one").
func kitesailAnswerPicks(t *testing.T, g *game.Game, chooser uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	for i := 0; i < game.MaxPlayers; i++ {
		ch := latestPickTarget(g, chooser)
		if ch == nil {
			return
		}
		var refs []game.TargetRef
		for _, w := range want {
			if hasID(ch.PickTargetCards, w) {
				refs = append(refs, game.TargetRef{Kind: game.TargetCard, ID: w})
				break
			}
		}
		if err := g.ResolvePickTargets(ch.ID, chooser, refs); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
}

// kitesailIsTreasure reports whether the permanent is now only a
// Treasure artifact with the Treasure's one mana ability.
func kitesailIsTreasure(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	ch := gaLayered(t, g, id)
	shapes, _ := game.ManaAbilitiesWithOrigins(ch)
	return ch.IsArtifact() && ch.HasSubtype("Treasure") && !ch.IsCreature() && len(shapes) == 1
}

// #3863: for each player, up to one other artifact or creature that
// player controls becomes a Treasure for as long as the Larcenist
// remains. The Larcenist itself is never offered, a pick is bound to
// its player, and the creatures come back when it leaves.
func TestKitesailLarcenistTurnsOnePermanentPerPlayerIntoATreasure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	other := g.Seats[(g.Turn.ActiveSeat+2)%4]
	mine := seedBear(g, me.ID)
	theirs := pushDiesCreatureForTest(g, opp.ID, "Legendary Nighthawk", vampireNighthawkOracle, "Legendary Creature — Vampire Shaman", 2, 3)
	rock := b12Permanent(g, other.ID, "Mind Stone", "Artifact")
	spare := seedBear(g, opp.ID)

	larcenist := castHoldCreature(t, g, "Kitesail Larcenist", "Creature — Human Pirate", kitesailLarcenistOracle, 2, 3)
	first := latestPickTarget(g, me.ID)
	if first == nil {
		t.Fatal("the enters trigger asks for its targets")
	}
	if hasID(first.PickTargetCards, larcenist) {
		t.Error("the Larcenist is offered as one of its own targets; the clause says other")
	}
	if hasID(first.PickTargetCards, theirs) {
		t.Error("your clause offers an opponent's creature")
	}
	kitesailAnswerPicks(t, g, me.ID, mine, theirs, rock)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{mine, theirs, rock} {
		if !kitesailIsTreasure(t, g, id) {
			t.Errorf("%s is not a Treasure artifact with one mana ability: %q", id, gaLayered(t, g, id).TypeLine)
		}
	}
	if ch := gaLayered(t, g, theirs); !ch.HasSupertype("Legendary") || game.HasKeyword(&ch, "flying") {
		t.Errorf("the Nighthawk keeps its supertype and loses its abilities: %q", ch.TypeLine)
	}
	if kitesailIsTreasure(t, g, spare) {
		t.Error("only one permanent per player is chosen")
	}

	// It leaves: every chosen permanent is itself again (CR 611.2b).
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, larcenist); err != nil {
		t.Fatalf("move the Larcenist: %v", err)
	}
	for _, id := range []uuid.UUID{mine, theirs} {
		if ch := gaLayered(t, g, id); !ch.IsCreature() || ch.HasSubtype("Treasure") {
			t.Errorf("%s is still a Treasure after the Larcenist left: %q", id, ch.TypeLine)
		}
	}
}

// A Larcenist that leaves before its trigger resolves changes nothing:
// the duration has already ended (CR 611.2b).
func TestKitesailLarcenistGoneBeforeResolutionChangesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	theirs := seedBear(g, opp.ID)
	larcenist := castHoldCreature(t, g, "Kitesail Larcenist", "Creature — Human Pirate", kitesailLarcenistOracle, 2, 3)
	kitesailAnswerPicks(t, g, me.ID, theirs)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, larcenist); err != nil {
		t.Fatalf("move the Larcenist: %v", err)
	}
	passPriorityAroundTable(t, g)
	if kitesailIsTreasure(t, g, theirs) {
		t.Error("the trigger resolved after the Larcenist left and still made a Treasure")
	}
}
