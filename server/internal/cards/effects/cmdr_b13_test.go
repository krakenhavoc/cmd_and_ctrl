package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cmdrB13EshkiOracle = "1b162dd3-3be6-406d-bf86-f7cc9eff098d"

func cmdrB13CastCreature(t *testing.T, g *game.Game, me *game.Player, power int) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Big", TypeLine: "Creature — Beast", Power: power, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdrB13EshkiScalesWithTheSpellsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eshki := b12Push(g, me.ID, "Eshki, Temur's Roar", "Legendary Creature — Human Warrior", cmdrB13EshkiOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	life := g.Seats[1].Life
	cmdrB13CastCreature(t, g, me, 3)
	if plusOneCounters(g, eshki) != 1 || me.Hand.Size() != hand || g.Seats[1].Life != life {
		t.Fatalf("power 3: counters=%d hand=%d->%d life=%d", plusOneCounters(g, eshki), hand, me.Hand.Size(), g.Seats[1].Life)
	}
	hand = me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 4)
	if plusOneCounters(g, eshki) != 2 || me.Hand.Size() != hand+1 || g.Seats[1].Life != life {
		t.Fatalf("power 4: counters=%d hand=%d->%d life=%d", plusOneCounters(g, eshki), hand, me.Hand.Size(), g.Seats[1].Life)
	}
	cmdrB13CastCreature(t, g, me, 6)
	// Eshki is 2/2 with three counters: 5 damage to each opponent.
	for _, o := range g.Seats[1:] {
		if o.Life != life-5 {
			t.Errorf("opponent life %d, want %d", o.Life, life-5)
		}
	}
	if g.Seats[0].Life != life {
		t.Errorf("controller life changed")
	}
}

const cmdrB13FelotharOracle = "2b3f791a-d672-46fb-a03a-96e19f2c3121"

func TestCmdrB13FelotharAssignsDamageByToughnessAndIgnoresDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	felothar := b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	bear := pushVanillaCreature(g, me.ID, "Glass Cannon", 4, 1)
	wall := pushVanillaCreature(g, me.ID, "Stout Wall", 0, 3)
	before := opp.Life
	attackWith(t, g, opp.ID, felothar, bear, wall)
	passPriorityAroundTable(t, g)
	if want := before - (5 + 1 + 3); opp.Life != want {
		t.Errorf("opponent at %d, want %d (toughness 5 + 1 + 3)", opp.Life, want)
	}
}

func TestCmdrB13FelotharDoesNotReachOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	theirs := pushVanillaCreature(g, opp.ID, "Their Cannon", 4, 1)
	advanceToNextSeatsTurn(t, g)
	before := me.Life
	attackWith(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if me.Life != before-4 {
		t.Errorf("an opponent's 4/1 still deals 4: %d -> %d", before, me.Life)
	}
}

func TestCmdrB13FelotharSacrificeDrawsToughnessDiscardsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	felothar := b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	fed := pushVanillaCreature(g, me.ID, "Fodder", 3, 2)
	for i := 0; i < 6; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, felothar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fed}})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d, want %d after drawing two", me.Hand.Size(), hand+2)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d after discarding three", me.Hand.Size(), hand-1)
	}
	if !me.Graveyard.Contains(fed) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

const cmdrB13KotisOracle = "0177b410-b559-491f-b393-ac3ed774653c"

func cmdrB13GraveyardCast(g *game.Game, me *game.Player, id uuid.UUID, pay []uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGraveyardExile, AltCostIDs: pay,
	})
}

func TestCmdrB13KotisCastsFromGraveyardOncePerTurnAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kotis := b12Push(g, me.ID, "Kotis, Sibsig Champion", "Legendary Creature — Zombie Warrior", cmdrB13KotisOracle, 3, 3)
	toMain(t, g)
	first := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	second := seedGraveyardCard(t, g, "Dead Elk", "Creature — Elk", "")
	fodder := seedGraveyardFodder(t, g, 6)

	// Two other cards are not enough.
	if err := cmdrB13GraveyardCast(g, me, first, fodder[:2]); err == nil {
		t.Fatal("cast with only two other cards exiled was accepted")
	}
	if err := cmdrB13GraveyardCast(g, me, first, fodder[:3]); err != nil {
		t.Fatalf("cast from graveyard: %v", err)
	}
	for _, id := range fodder[:3] {
		if me.Graveyard.Contains(id) {
			t.Error("a paid card is still in the graveyard")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the creature did not resolve")
	}
	if got := plusOneCounters(g, kotis); got != 2 {
		t.Errorf("Kotis has %d counters, want 2", got)
	}
	// Once during each of your turns.
	if err := cmdrB13GraveyardCast(g, me, second, fodder[3:]); err == nil {
		t.Error("a second graveyard cast in one turn was accepted")
	}
}

func TestCmdrB13KotisIgnoresCreaturesCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kotis := b12Push(g, me.ID, "Kotis, Sibsig Champion", "Legendary Creature — Zombie Warrior", cmdrB13KotisOracle, 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 2)
	if got := plusOneCounters(g, kotis); got != 0 {
		t.Errorf("Kotis grew to %d from a creature cast from hand", got)
	}
}

const cmdrB13ShikoOracle = "40ed32a6-ad56-48c7-aecf-b4238c34c212"

func TestCmdrB13ShikoCopiesATargetedSecondSpellAndDrawsOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Shiko and Narset, Unified", "Legendary Creature — Human Spirit Dragon", cmdrB13ShikoOracle, 4, 4)
	advanceTo(t, g, game.StepPrecombatMain)
	// The first spell is not a second spell: nothing happens.
	hand := me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 1)
	if me.Hand.Size() != hand {
		t.Fatalf("the first spell drew: hand %d -> %d", hand, me.Hand.Size())
	}
	// The second spell targets a player: copied, so the opponent takes 6.
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life > life-3 {
		t.Fatalf("opponent at %d, want at most %d", opp.Life, life-3)
	}
	if opp.Life != life-6 {
		t.Errorf("opponent at %d, want %d (the bolt and its copy)", opp.Life, life-6)
	}
}

func TestCmdrB13ShikoDrawsWhenTheSecondSpellHasNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Shiko and Narset, Unified", "Legendary Creature — Human Spirit Dragon", cmdrB13ShikoOracle, 4, 4)
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 1)
	hand := me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 1)
	// One card left the hand for the cast, one was drawn.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the cast is net zero, one card drawn)", me.Hand.Size(), hand+1)
	}
}
