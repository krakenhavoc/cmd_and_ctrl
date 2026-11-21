package testsg

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestGoblinWelderRefusesTheWrongGraveyardBeforeTapping(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", effectstest.GoblinWelderOracle, false)
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Their artifact", "Artifact", "", false)
	b := effectstest.WelderGraveyardArtifact(me, "My artifact", "")
	err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("wrong graveyard: %v", err)
	}
	card, _ := effectstest.BattlefieldCard(g, w)
	if card.Tapped || len(g.StackMeta) != 0 {
		t.Fatal("illegal targets paid the cost or put an ability on the stack")
	}
}

func TestGoblinWelderExchangesAnOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", effectstest.GoblinWelderOracle, false)
	a := effectstest.PushCatalogPermanent(g, opp.ID, "Their artifact", "Artifact", "", false)
	b := effectstest.WelderGraveyardArtifact(opp, "Their Sol Ring", effectstest.SolRingOracle)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !opp.Graveyard.Contains(a) {
		t.Fatal("the outgoing artifact was not sacrificed")
	}
	c, ok := effectstest.BattlefieldCard(g, b)
	if !ok || c.Controller != opp.ID || c.Owner != opp.ID {
		t.Fatalf("returned artifact: %+v, found=%v", c, ok)
	}
	if len(game.ManaAbilitiesForCard(c)) == 0 {
		t.Fatal("returned artifact lost its mana ability")
	}
}

func TestGoblinWelderDoesNothingWhenEitherTargetBecomesIllegal(t *testing.T) {
	for _, changed := range []string{"controller", "graveyard"} {
		t.Run(changed, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			w := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", effectstest.GoblinWelderOracle, false)
			a := effectstest.PushCatalogPermanent(g, opp.ID, "Artifact", "Artifact", "", false)
			b := effectstest.WelderGraveyardArtifact(opp, "Return me", "")
			if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
				{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
			}}); err != nil {
				t.Fatal(err)
			}
			g.WithWriteLock(func() {
				if changed == "controller" {
					for i := range g.Battlefield.Cards {
						if g.Battlefield.Cards[i].InstanceID == a {
							g.Battlefield.Cards[i].Controller = me.ID
						}
					}
				} else {
					_, _ = game.MoveCard(opp.Graveyard, g.Exile, b)
				}
			})
			effectstest.PassPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
				t.Fatal("Welder performed half an exchange with an illegal target")
			}
		})
	}
}

func TestCopiedArtifactRunsItsOnEnterHook(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	template := effectstest.WelderGraveyardArtifact(me, "Adaptive Automaton", "53c730c6-2f8c-4af8-b400-b9d573a71e60")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == template {
			me.Graveyard.Cards[i].TypeLine = "Artifact Creature — Construct"
			me.Graveyard.Cards[i].Power = 2
			me.Graveyard.Cards[i].Toughness = 2
		}
	}
	g.WithWriteLock(func() {
		if err := (effects.CreateTokenCopy{Controller: me.ID, Copy: template, N: 1}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) == 0 {
		t.Fatal("a token copy skipped its catalog entry hook")
	}
}
