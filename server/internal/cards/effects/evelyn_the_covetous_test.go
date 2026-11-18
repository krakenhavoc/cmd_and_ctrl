package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// evelyn_the_covetous_test.go — #5876: collection counters, "exiled by an
// ability you controlled" and the once-each-turn play from exile.

const evelynOracle = "4ab6f615-1f04-4ed2-88b2-49faa8356dd0"

const evelynTypeLine = "Legendary Creature — Vampire Rogue"

func TestEvelynTheCovetousIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(evelynOracle)
	if !ok || spec.Name != "Evelyn, the Covetous" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// evelynTops seeds the top card of every seat's library and returns them
// by seat.
func evelynTops(g *game.Game, typeLine, cost string) map[uuid.UUID]uuid.UUID {
	out := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		out[p.ID] = topOfLibrary(p, "Top of "+p.Name, typeLine, cost)
	}
	return out
}

func evelynPlay(g *game.Game, who *game.Player, id uuid.UUID) error {
	return g.CastSpell(who.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true})
}

// Evelyn entering exiles the top card of each player's library, each with
// a collection counter and marked as exiled by her controller's ability.
// One of them, an opponent's, is cast with mana of the wrong colour; a
// second play the same turn is refused.
func TestEvelynCollectsEachLibraryAndPlaysOncePerTurnWithAnyColor(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	tops := evelynTops(g, "Creature — Bear", "{G}{G}")

	enterFromHand(t, g, me.ID, "Evelyn, the Covetous", evelynTypeLine, evelynOracle)
	passPriorityAroundTable(t, g)
	for seat, id := range tops {
		c := cardInExile(t, g, id)
		if c.Counters[evelynCollectionCounter] != 1 {
			t.Errorf("seat %s's top card: collection counters = %v, want 1", seat, c.Counters)
		}
		if c.ExiledByAbilityOf != me.ID {
			t.Errorf("seat %s's top card: exiled by %s, want Evelyn's controller", seat, c.ExiledByAbilityOf)
		}
	}
	g.Turn.Step = game.StepPrecombatMain

	// {G}{G} paid with black mana: "as though it were mana of any color".
	payAnyType(t, g, me, "{B}{B}")
	theirs := tops[opp.ID]
	if err := evelynPlay(g, me, theirs); err != nil {
		t.Fatalf("play an opponent's collected card with any-colour mana: %v", err)
	}
	passPriorityAroundTable(t, g)
	got, ok := battlefieldCard(g, theirs)
	if !ok || got.Controller != me.ID {
		t.Fatal("the collected creature did not resolve under Evelyn's controller")
	}
	if got.Counters[evelynCollectionCounter] != 0 {
		t.Errorf("the collection counter followed the card out of exile: %v", got.Counters)
	}

	// Once each turn: my own collected card is refused now.
	g.Turn.Step = game.StepPrecombatMain
	mine := tops[me.ID]
	payAnyType(t, g, me, "{G}{G}")
	if err := evelynPlay(g, me, mine); err == nil {
		t.Fatal("a second play from exile in the same turn")
	}
	if enumeratedFor(g, me.ID, mine) {
		t.Error("the spent permission is still offered")
	}
	if e2Zone(g, mine) != game.ZoneExile {
		t.Error("the refused card left exile")
	}
}

// Cards an OPPONENT's Evelyn exiled carry collection counters but were
// not exiled by an ability I controlled, so my Evelyn does not open them;
// nor does a collection-countered card nobody's ability exiled.
func TestEvelynRefusesCardsAnOpponentsAbilityExiled(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Evelyn, the Covetous", evelynTypeLine, evelynOracle, false)
	tops := evelynTops(g, "Sorcery", "{B}")

	enterFromHand(t, g, opp.ID, "Evelyn, the Covetous", evelynTypeLine, evelynOracle)
	passPriorityAroundTable(t, g)
	g.Turn.Step = game.StepPrecombatMain
	for _, id := range tops {
		c := cardInExile(t, g, id)
		if c.Counters[evelynCollectionCounter] != 1 || c.ExiledByAbilityOf != opp.ID {
			t.Fatalf("card exiled by the opponent's Evelyn: counters %v, by %s", c.Counters, c.ExiledByAbilityOf)
		}
	}
	for _, id := range []uuid.UUID{tops[me.ID], tops[opp.ID]} {
		payAnyType(t, g, me, "{B}")
		if err := evelynPlay(g, me, id); err == nil {
			t.Fatal("played a card an opponent's ability exiled")
		}
		if enumeratedFor(g, me.ID, id) || viewCastable(t, g, me.ID, id) {
			t.Error("enumerator or view offers a card an opponent's ability exiled")
		}
	}

	// A collection counter alone is not enough.
	loose := stashed(g, opp.ID, "Loose Card", "Sorcery", "{B}", 0)
	g.WithWriteLock(func() {
		if err := g.AddCounterForEffect(loose, evelynCollectionCounter, 1); err != nil {
			t.Fatal(err)
		}
	})
	if err := evelynPlay(g, me, loose); err == nil {
		t.Fatal("played a collected card no ability of mine exiled")
	}
}

// Another Vampire entering under my control triggers Evelyn too, and a
// collected land is PLAYED from exile, which spends the turn's use.
func TestEvelynAnotherVampireTriggersAndALandIsPlayable(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	pushCatalogPermanent(g, me.ID, "Evelyn, the Covetous", evelynTypeLine, evelynOracle, false)
	land := topOfLibrary(opp, "Their Island", "Basic Land — Island", "")
	other := topOfLibrary(me, "My Sorcery", "Sorcery", "{B}")

	// A non-Vampire entering does nothing.
	enterFromHand(t, g, me.ID, "Bear", "Creature — Bear", "")
	passPriorityAroundTable(t, g)
	if e2Zone(g, land) != game.ZoneLibrary {
		t.Fatal("a non-Vampire entering triggered Evelyn")
	}

	enterFromHand(t, g, me.ID, "Vampire Friend", "Creature — Vampire", "")
	passPriorityAroundTable(t, g)
	g.Turn.Step = game.StepPrecombatMain
	if c := cardInExile(t, g, land); c.Counters[evelynCollectionCounter] != 1 {
		t.Fatalf("the land was not collected: %v", c.Counters)
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("play the collected land: %v", err)
	}
	if e2Zone(g, land) != game.ZoneBattlefield {
		t.Fatalf("the land is in %q, want the battlefield", e2Zone(g, land))
	}
	payAnyType(t, g, me, "{B}")
	if err := evelynPlay(g, me, other); err == nil {
		t.Fatal("the land play did not spend the turn's use")
	}
}
