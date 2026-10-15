package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// draconic_intervention_test.go — #3537: Draconic Intervention and Relentless
// Skaabs, the two spells whose additional cost is "exile a <kind> card from
// your graveyard" (#3536).

const (
	draconicInterventionOracle = "9fe75f56-c2e4-4c82-8113-039b8c486fae"
	relentlessSkaabsOracle     = "5089cd0f-1338-4ecd-bb1b-664d7e72f8fc"
)

func pushTypedGraveyardCard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return id
}

func castDraconic(t *testing.T, g *game.Game, exile ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Draconic Intervention", "Sorcery", "{2}{R}{R}", draconicInterventionOracle,
		game.CastSpellParams{ExileIDs: exile})
}

// A three-mana-value instant: three damage to each non-Dragon creature, a
// creature that would die is exiled instead, a Dragon is untouched, and the
// spell exiles itself.
func TestDraconicInterventionDealsXAndExilesWhatDies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuel := pushTypedGraveyardCard(me, "Lightning Helix", "Instant", "{1}{R}{W}")
	small := p1Creature(g, opp.ID, 2, 2)
	big := p1Creature(g, opp.ID, 5, 5)
	dragon := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Whelp",
		TypeLine: "Creature — Dragon", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})

	spell, err := castDraconic(t, g, fuel)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(fuel) {
		t.Error("the instant was not exiled by the cost")
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, small, game.ZoneExile, "the 2/2 that took 3")
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 5/5 that took 3")
	p1WantZone(t, g, dragon, game.ZoneBattlefield, "the Dragon")
	p1Destroy(t, g, big)
	p1WantZone(t, g, big, game.ZoneExile, "the damaged 5/5 destroyed later this turn")
	p1WantZone(t, g, spell, game.ZoneExile, "Draconic Intervention itself")
}

// X is the exiled card's mana value: a one-mana sorcery deals one.
func TestDraconicInterventionXFollowsTheExiledCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuel := pushTypedGraveyardCard(me, "Shock", "Sorcery", "{R}")
	bear := p1Creature(g, opp.ID, 2, 2)
	if _, err := castDraconic(t, g, fuel); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneBattlefield, "the 2/2 that took 1")
}

// The cost wants an instant or sorcery: a creature card is refused.
func TestDraconicInterventionRefusesANonSpellCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := pushTypedGraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	spell, err := castDraconic(t, g, fuel)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast naming a creature card = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(spell) || g.Exile.Size() != 0 {
		t.Error("a refused cast moved cards")
	}
}

func castSkaabs(t *testing.T, g *game.Game, exile ...uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	return castWithTapParams(t, g, "Relentless Skaabs", "Creature — Zombie", "{3}{U}{U}", relentlessSkaabsOracle,
		game.CastSpellParams{ExileIDs: exile})
}

// A graveyard with no creature card cannot pay.
func TestRelentlessSkaabsRefusesAGraveyardWithNoCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	instant := pushTypedGraveyardCard(me, "Opt", "Instant", "{U}")
	spell, err := castSkaabs(t, g, instant)
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast naming an instant = %v, want ErrInvalidParam", err)
	}
	if g.Stack.Contains(spell) || g.Exile.Size() != 0 || me.Graveyard.Size() != 1 {
		t.Error("a refused cast moved cards")
	}
}

// Exiling a creature card casts it, and undying returns it with a counter.
func TestRelentlessSkaabsExilesACreatureCardAndHasUndying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fuel := pushTypedGraveyardCard(me, "Bear", "Creature — Bear", "{1}{G}")
	spell, err := castSkaabs(t, g, fuel)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(fuel) {
		t.Error("the creature card was not exiled by the cost")
	}
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, spell) {
		t.Fatal("Skaabs did not resolve")
	}
	b25Destroy(g, spell)
	settleProwess(t, g)
	if !onBattlefield(g, spell) || countersOn(g, spell, game.CounterPlusOne) != 1 {
		t.Fatal("undying did not return Skaabs with a +1/+1 counter")
	}
}
