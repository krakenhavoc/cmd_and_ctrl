package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const arvinoxOracle = "9abecd23-a2d9-43c7-8d01-f8afdc53026e"

// arvinoxStolen puts a permanent `owner` owns under `controller`'s control.
func arvinoxStolen(g *game.Game, owner, controller uuid.UUID, name string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Artifact",
		Owner: owner, Controller: controller,
	})
}

func TestArvinoxIsACreatureOnlyWithThreePermanentsYouDontOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	arv := b12Push(g, me.ID, "Arvinox, the Mind Flail", "Legendary Enchantment Creature — Horror", arvinoxOracle, 9, 9)
	b12Push(g, me.ID, "My Own Rock", "Artifact", "", 0, 0)
	arvinoxStolen(g, opp.ID, me.ID, "Stolen A")
	arvinoxStolen(g, opp.ID, me.ID, "Stolen B")
	arvinoxStolen(g, opp.ID, opp.ID, "Theirs, Under Their Control")

	c := s58p6View(t, g, arv)
	if c.IsCreature() {
		t.Fatal("two permanents I don't own: Arvinox is a creature")
	}
	if !c.IsEnchantment() || c.HasSubtype("Horror") {
		t.Errorf("not a creature: enchantment=%v Horror=%v, want an enchantment with no creature type (CR 205.1a)", c.IsEnchantment(), c.HasSubtype("Horror"))
	}
	arvinoxStolen(g, g.Seats[2].ID, me.ID, "Stolen C")
	c = s58p6View(t, g, arv)
	if !c.IsCreature() || c.CurrentPower() != 9 || !c.HasSubtype("Horror") {
		t.Errorf("three permanents I don't own: creature=%v power=%d Horror=%v, want a 9/9 Horror", c.IsCreature(), c.CurrentPower(), c.HasSubtype("Horror"))
	}
}

// arvinoxLibrary replaces a player's library with one card per type line,
// the first on the bottom, and returns their ids bottom first.
func arvinoxLibrary(p *game.Player, typeLines ...string) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, 0, len(typeLines))
	for i, tl := range typeLines {
		c := game.NewCard("Library Card "+string(rune('A'+i)), p.ID)
		c.TypeLine = tl
		c.ManaCost = "{1}{G}"
		p.Library.PushTop(c)
		ids = append(ids, c.InstanceID)
	}
	return ids
}

func TestArvinoxExilesTheBottomCardOfEachOpponentsLibraryFaceDown(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Arvinox, the Mind Flail", "Legendary Enchantment Creature — Horror", arvinoxOracle, 9, 9)
	mine := arvinoxLibrary(me, "Creature — Test", "Creature — Test")
	var bottoms, tops []uuid.UUID
	kinds := []string{"Creature — Test", "Instant", "Land"}
	for i, opp := range g.Seats[1:] {
		ids := arvinoxLibrary(opp, kinds[i%len(kinds)], "Creature — Test", "Creature — Test")
		bottoms, tops = append(bottoms, ids[0]), append(tops, ids[2])
	}
	if len(bottoms) < 3 {
		t.Fatalf("want a four-player table, got %d opponents", len(bottoms))
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)

	for i, id := range bottoms {
		if got := b12ZoneOf(g, id); got != game.ZoneExile {
			t.Fatalf("opponent %d's bottom card is in %q, want exile", i+1, got)
		}
		e2AssertFaceDownFor(t, g, id, me.ID)
		// The grant opens a permanent spell only: the instant and the land
		// are looked at and nothing more.
		perm := e2Permission(g, id)
		if i > 0 {
			if perm != nil {
				t.Errorf("opponent %d: a %s is castable (%+v)", i+1, kinds[i], perm)
			}
		} else if perm == nil || perm.Player != me.ID || !perm.CastOnly || !perm.AnyColor || perm.AnyType {
			t.Errorf("opponent %d: permission %+v, want a cast grant spending mana as any color (not any type)", i+1, perm)
		}
		if got := b12ZoneOf(g, tops[i]); got != game.ZoneLibrary {
			t.Errorf("opponent %d's top card is in %q, want library", i+1, got)
		}
	}
	for _, id := range mine {
		if got := b12ZoneOf(g, id); got != game.ZoneLibrary {
			t.Errorf("my own library was touched (%s)", got)
		}
	}

	// My next main phase: the {1}{G} creature is cast off Swamps; the
	// instant is not a permanent spell and the land is not a spell.
	advanceToMainOf(t, g, 0)
	e2Swamps(g, me.ID, 6)
	strict := game.CastSpellParams{FromZone: "exile", AutoTap: true, Strict: true}
	if err := g.CastSpell(me.ID, bottoms[1], strict); err == nil {
		t.Error("an exiled instant was cast: only permanent spells may be")
	}
	if err := g.CastSpell(me.ID, bottoms[2], game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("an exiled land was played: the grant is to cast")
	}
	if err := g.CastSpell(me.ID, bottoms[0], strict); err != nil {
		t.Fatalf("cast the exiled {1}{G} creature off Swamps: %v", err)
	}
}

func TestArvinoxSkipsAnOpponentWithAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, other := g.Seats[0], g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Arvinox, the Mind Flail", "Legendary Enchantment Creature — Horror", arvinoxOracle, 9, 9)
	opp.Library.Cards = nil
	ids := arvinoxLibrary(other, "Creature — Test", "Creature — Test")
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if got := b12ZoneOf(g, ids[0]); got != game.ZoneExile {
		t.Fatalf("the other opponent's bottom card is in %q, want exile", got)
	}
}
