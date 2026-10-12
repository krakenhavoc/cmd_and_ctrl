package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const goblinWelderOracle = "9fc28e1f-f4ea-4bf8-9442-a36dba3b3b29"

func welderGraveyardArtifact(p *game.Player, name, oracle string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.Controller, c.TypeLine, c.OracleID = p.ID, "Artifact", oracle
	p.Graveyard.PushTop(c)
	return c.InstanceID
}

func TestGoblinWelderRefusesTheWrongGraveyardBeforeTapping(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := pushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", goblinWelderOracle, false)
	a := pushCatalogPermanent(g, opp.ID, "Their artifact", "Artifact", "", false)
	b := welderGraveyardArtifact(me, "My artifact", "")
	err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("wrong graveyard: %v", err)
	}
	card, _ := battlefieldCard(g, w)
	if card.Tapped || len(g.StackMeta) != 0 {
		t.Fatal("illegal targets paid the cost or put an ability on the stack")
	}
}

func TestGoblinWelderExchangesAnOpponentsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	w := pushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", goblinWelderOracle, false)
	a := pushCatalogPermanent(g, opp.ID, "Their artifact", "Artifact", "", false)
	b := welderGraveyardArtifact(opp, "Their Sol Ring", solRingOracle)
	if err := g.ActivateCatalogAbility(me.ID, w, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b, Slot: 1},
	}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || !opp.Graveyard.Contains(a) {
		t.Fatal("the outgoing artifact was not sacrificed")
	}
	c, ok := battlefieldCard(g, b)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			w := pushCatalogPermanent(g, me.ID, "Goblin Welder", "Creature — Goblin Artificer", goblinWelderOracle, false)
			a := pushCatalogPermanent(g, opp.ID, "Artifact", "Artifact", "", false)
			b := welderGraveyardArtifact(opp, "Return me", "")
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
			passPriorityAroundTable(t, g)
			if !g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
				t.Fatal("Welder performed half an exchange with an illegal target")
			}
		})
	}
}

func TestCopiedArtifactRunsItsOnEnterHook(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	template := welderGraveyardArtifact(me, "Adaptive Automaton", "53c730c6-2f8c-4af8-b400-b9d573a71e60")
	for i := range me.Graveyard.Cards {
		if me.Graveyard.Cards[i].InstanceID == template {
			me.Graveyard.Cards[i].TypeLine = "Artifact Creature — Construct"
			me.Graveyard.Cards[i].Power = 2
			me.Graveyard.Cards[i].Toughness = 2
		}
	}
	g.WithWriteLock(func() {
		if err := (CreateTokenCopy{Controller: me.ID, Copy: template, N: 1}).Apply(NewContext(g, nil)); err != nil {
			t.Fatal(err)
		}
	})
	if len(g.PendingChoices) == 0 {
		t.Fatal("a token copy skipped its catalog entry hook")
	}
}
