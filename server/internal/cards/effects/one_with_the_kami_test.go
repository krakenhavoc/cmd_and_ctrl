package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oneWithTheKamiOracle = "006f4a62-8590-4ca0-9f51-3aea26cd1a54"

func TestOneWithTheKamiIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(oneWithTheKamiOracle)
	if !ok || spec.Name != "One with the Kami" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func bigBear(g *game.Game, owner uuid.UUID, power int, counters bool) uuid.UUID {
	c := game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: power, Toughness: 2, Owner: owner, Controller: owner}
	if counters {
		c.Counters = map[string]int{"+1/+1": 1}
	}
	return pushBattlefieldCardWithTimestamp(g, c)
}

func TestOneWithTheKamiMakesSpiritsEqualToTheEnchantedCreaturesPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := bigBear(g, me.ID, 3, false)
	auraCast(t, g, "One with the Kami", oneWithTheKamiOracle, bear)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Spirit"); n != 3 {
		t.Errorf("Spirits = %d, want 3", n)
	}
}

func TestOneWithTheKamiTriggersForAnotherModifiedCreatureButNotAPlainOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := bigBear(g, me.ID, 1, false)
	auraCast(t, g, "One with the Kami", oneWithTheKamiOracle, host)
	plain := bigBear(g, me.ID, 5, false)
	killCreature(t, g, me.ID, plain)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Spirit"); n != 0 {
		t.Fatalf("a creature with no modification made %d Spirits", n)
	}
	modified := bigBear(g, me.ID, 2, true) // power 2 base; counter is layer 7 so the power as it died is 3
	want := 0
	if c, ok := g.LookupCardForEffect(modified); ok {
		want = c.CurrentPower()
	}
	killCreature(t, g, me.ID, modified)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Spirit"); n != want || n == 0 {
		t.Errorf("Spirits = %d, want %d (its power as it died)", n, want)
	}
}

func TestOneWithTheKamiIgnoresAnOpponentsModifiedCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := bigBear(g, me.ID, 1, false)
	auraCast(t, g, "One with the Kami", oneWithTheKamiOracle, host)
	theirs := bigBear(g, opp.ID, 4, true)
	killCreature(t, g, opp.ID, theirs)
	passPriorityAroundTable(t, g)
	if n := countTokensControlled(g, me.ID, "Spirit"); n != 0 {
		t.Errorf("Spirits = %d, want 0", n)
	}
}

func TestOneWithTheKamiMustEnchantACreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	theirs := bigBear(g, g.Seats[1].ID, 2, false)
	hand := g.Seats[0].Hand
	id := uuid.New()
	hand.PushTop(game.Card{InstanceID: id, Name: "One with the Kami", OracleID: oneWithTheKamiOracle, TypeLine: auraTypeLine,
		ManaCost: "{3}{G}", Owner: g.Seats[0].ID, Controller: g.Seats[0].ID})
	toMainForCost(t, g)
	apaMana(g.Seats[0], "G", "G", "G", "G")
	err := g.CastSpell(g.Seats[0].ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}})
	if err == nil {
		t.Error("cast on an opponent's creature was accepted")
	}
}
