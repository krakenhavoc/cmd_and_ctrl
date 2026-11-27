package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b736_test.go — Portal Second Age, batch 736 (#5625): damage, card
// draw and counterspell-style bounce. Each test plays the real catalog entry.

const (
	oldB736Hurricane = "9c021685-4017-49c7-9f58-2ae0243361a0"
	oldB736Kiss      = "c2c25406-e84a-47b0-8a46-fe60f9b0cffa"
	oldB736Theft     = "008011e2-7b82-4962-af6e-be627112f37f"
	oldB736Remove    = "a1ed0f7e-56c3-4866-89c4-77b81a952955"
)

func TestOldB736EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB736Hurricane: "Hurricane", oldB736Kiss: "Kiss of Death",
		oldB736Theft: "Theft of Dreams", oldB736Remove: "Remove",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v; want Full with none", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB736Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestOldB736HurricaneHitsOnlyFlyingCreaturesAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB736Seats(g)
	bird := effectstest.BrCreature(g, opp.ID, "Bird", "Creature — Bird", 1, 5, "flying")
	myBird := effectstest.BrCreature(g, me.ID, "My Bird", "Creature — Bird", 1, 5, "flying")
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 5)
	life, oppLife := me.Life, opp.Life
	effectstest.Pr8CastX(t, g, "Hurricane", "Sorcery", oldB736Hurricane, 3, nil)
	effectstest.P1g2WantDamage(t, g, bird, 3, "the opposing flier")
	effectstest.P1g2WantDamage(t, g, myBird, 3, "your own flier")
	effectstest.P1g2WantDamage(t, g, bear, 0, "the creature without flying")
	if me.Life != life-3 || opp.Life != oppLife-3 {
		t.Errorf("life %d and %d, want both players down 3 (%d, %d)", me.Life, opp.Life, life-3, oppLife-3)
	}
}

func TestOldB736KissOfDeathDamagesAnOpponentAndGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB736Seats(g)
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Kiss of Death", "Sorcery", oldB736Kiss,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-4 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, oppLife-4)
	}
	if me.Life != life+4 {
		t.Errorf("the caster is at %d, want %d", me.Life, life+4)
	}
}

func TestOldB736KissOfDeathRefusesYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB736Seats(g)
	if err := effectstest.CastCatalogSpellErr(t, g, "Kiss of Death", "Sorcery", oldB736Kiss,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Error("Kiss of Death targeted its own controller")
	}
}

func TestOldB736TheftOfDreamsDrawsForEachTappedCreatureTheOpponentControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB736Seats(g)
	for i := 0; i < 5; i++ {
		effectstest.SeedLibraryTop(me, "Card", "Sorcery", "")
	}
	tapped1 := effectstest.B12Creature(g, opp.ID, "Tapped 1", "Creature — Bear", 2, 2)
	tapped2 := effectstest.B12Creature(g, opp.ID, "Tapped 2", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			switch g.Battlefield.Cards[i].InstanceID {
			case tapped1, tapped2, mine:
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Theft of Dreams", "Sorcery", oldB736Theft,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2 (one per tapped creature the opponent controls)", got)
	}
}

func TestOldB736RemoveBouncesAnAttackerOnlyWhenYouWereAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	attacker := effectstest.SfPush(g, me.ID, "Attacker", "Creature — Test", 3, []string{"R"})
	home := effectstest.SfPush(g, me.ID, "Home", "Creature — Test", 1, []string{"R"})
	if err := effectstest.CastCatalogSpellErr(t, g, "Remove", "Instant", oldB736Remove, effectstest.CardTarget(attacker)); err == nil {
		t.Fatal("cast outside the declare attackers step")
	}
	effectstest.SfAttackHold(t, g, opp.ID, attacker)
	other := effectstest.SfInstant("Remove", oldB736Remove)
	other.InstanceID, other.Owner, other.Controller = uuid.New(), third.ID, third.ID
	third.Hand.PushTop(other)
	if err := g.CastSpell(third.ID, other.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(attacker)}); err == nil {
		t.Fatal("cast by a player nobody attacked")
	}
	card := effectstest.SfInstant("Remove", oldB736Remove)
	card.InstanceID, card.Owner, card.Controller = uuid.New(), opp.ID, opp.ID
	opp.Hand.PushTop(card)
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(home), Strict: true}); err == nil {
		t.Fatal("Remove targeted a creature that isn't attacking")
	}
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Targets: effectstest.CardTarget(attacker), Strict: true}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, attacker) {
		t.Error("the attacker is still on the battlefield")
	}
	back := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == attacker {
			back = true
		}
	}
	if !back {
		t.Error("the attacker did not return to its owner's hand")
	}
	if !effectstest.OnBattlefieldByID(g, home) {
		t.Error("another creature was bounced")
	}
}
