package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const shelobOracle = "7eade76f-92e4-4d26-9eb8-d3d954845c04"

func TestShelobDreadWeaverIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(shelobOracle)
	if !ok || spec.Name != "Shelob, Dread Weaver" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("spec = %+v, %v", spec, ok)
	}
}

// shelobTable puts Shelob on the active player's battlefield in their main
// phase, kills a nontoken creature of an opponent's (mana value `mv`) and
// resolves Shelob's trigger. It returns the game, Shelob's controller,
// Shelob, the opponent and the dead creature's card.
func shelobTable(t *testing.T, mv string) (*game.Game, *game.Player, uuid.UUID, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	shelob := effectstest.B12Push(g, me.ID, "Shelob, Dread Weaver", "Legendary Creature — Spider Demon", shelobOracle, 3, 3)
	victim := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].Name == "Their Bear" {
				g.Battlefield.Cards[i].ManaCost = mv
			}
		}
	})
	effectstest.Destroy(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CardInExile(t, g, victim).ExiledWith.ID; got != shelob {
		t.Fatalf("the dead creature is linked to %s, want Shelob", got)
	}
	return g, me, shelob, opp, victim
}

// A nontoken creature an opponent controls dies: Shelob exiles it, linked.
// One of Shelob's controller's own creatures dying is not exiled.
func TestShelobExilesOpponentsNontokenCreaturesThatDie(t *testing.T) {
	g, me, _, _, _ := shelobTable(t, "{1}{G}")
	mine := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	effectstest.Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(mine) {
		t.Fatal("Shelob's controller's own creature was exiled")
	}
}

// The second ability's cost puts a creature card exiled with Shelob into
// its owner's graveyard; it puts two +1/+1 counters on Shelob and draws.
// A card exiled by something else cannot pay it, and with nothing left
// exiled with Shelob the ability can't be activated.
func TestShelobProcessesAnExiledWithCardForCountersAndACard(t *testing.T) {
	g, me, shelob, opp, victim := shelobTable(t, "{1}{G}")
	other := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.PushTop(game.Card{InstanceID: other, Name: "Unlinked Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.FloatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, shelob, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{other}}); err == nil {
		t.Fatal("a card not exiled with Shelob paid the cost")
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, shelob, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("activate with the linked card: %v", err)
	}
	if !opp.Graveyard.Contains(victim) {
		t.Fatal("the paid card is not in its owner's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOneCounters(g, shelob); got != 2 {
		t.Errorf("Shelob has %d +1/+1 counters, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, shelob, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{victim}}); err == nil {
		t.Fatal("activated with no card exiled with Shelob")
	}
}

// The third ability puts a creature card exiled with Shelob, with mana
// value X, onto the battlefield tapped under Shelob's controller's
// control. A wrong X is not a legal target.
func TestShelobReanimatesAnExiledCreatureWithManaValueX(t *testing.T) {
	g, me, shelob, _, victim := shelobTable(t, "{1}{G}")
	effectstest.FloatMana(t, g, me, "{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, shelob, 1, game.ActivateAbilityParams{Strict: true, XValue: 1,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Fatal("targeted a mana value 2 card with X=1")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, shelob, 1, game.ActivateAbilityParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("activate with X=2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var got *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Their Bear" {
			got = &g.Battlefield.Cards[i]
		}
	}
	if got == nil {
		t.Fatal("the creature is not on the battlefield")
	}
	if !got.Tapped || got.Controller != me.ID {
		t.Errorf("tapped=%v controller=%s, want tapped under Shelob's controller", got.Tapped, got.Controller)
	}
}
