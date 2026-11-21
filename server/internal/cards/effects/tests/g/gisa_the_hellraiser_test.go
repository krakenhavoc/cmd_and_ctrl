package testsg

import (
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const gisaHellraiserOracle = "696363c9-e8a7-4c65-b65c-4a94891ec43d"

func TestCompoundWardCardsHaveOneManaAndLifeOffer(t *testing.T) {
	for _, card := range []struct {
		name, typeLine, oracle string
		power, toughness       int
	}{
		{"Gisa, the Hellraiser", "Legendary Creature — Human Warlock", gisaHellraiserOracle, 4, 4},
		{"Captain Howler, Sea Scourge", "Legendary Creature — Shark Pirate", "46c99d67-8307-4af0-afd0-377714806614", 5, 4},
	} {
		t.Run(card.name, func(t *testing.T) {
			for _, pay := range []bool{true, false} {
				g := effectstest.NewCatalogGame(t)
				me, payer := g.Seats[0], g.Seats[1]
				gisa := effectstest.B12Push(g, me.ID, card.name, card.typeLine, card.oracle, card.power, card.toughness)
				effectstest.AdvanceToMainOf(t, g, 1)
				// Bolt can target black Gisa; Doom Blade cannot.
				bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetCard(gisa))
				for i := 0; i < 12 && !effectstest.HasPayUnlessFor(g, payer.ID); i++ {
					if err := g.PassPriority(); err != nil {
						t.Fatal(err)
					}
				}
				if !effectstest.HasPayUnlessFor(g, payer.ID) {
					t.Fatal("no compound ward offer")
				}
				var offer *game.PendingChoice
				for _, c := range g.PendingChoices {
					if c.Kind == game.PendingChoicePayUnless && c.Chooser == payer.ID {
						offer = c
					}
				}
				if offer == nil || offer.PayCost != "{2}" || offer.LifeCost != 2 {
					t.Fatal("ward omitted one component")
				}
				before := payer.Life
				effectstest.B06AddMana(payer, "C", "C")
				effectstest.AnswerPayUnless(t, g, payer.ID, pay)
				if pay {
					if payer.Life != before-2 || payer.ManaPool.Total() != 0 || effectstest.CounteredInLog(g, bolt) {
						t.Fatal("ward did not pay both parts")
					}
				} else if payer.Life != before || payer.ManaPool.Total() != 2 || !effectstest.CounteredInLog(g, bolt) {
					t.Fatal("declining ward charged part of its cost")
				}
			}
		})
	}
}

func TestGisaCrimeTokensAndOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gisa, the Hellraiser", "Legendary Creature — Human Warlock", gisaHellraiserOracle, 4, 4)
	effectstest.AdvanceToMain(t, g)
	crime := func() {
		effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.B16TargetPlayer(opponent.ID))
		effectstest.PassPriorityAroundTable(t, g)
	}
	crime()
	count := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Zombie Rogue" {
			continue
		}
		count++
		if !c.Tapped || !c.HasColor("U") || !c.HasColor("B") || !c.HasSubtype("Zombie") || !c.HasSubtype("Rogue") || !game.HasKeyword(&c, "menace") || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
			t.Fatalf("wrong token: %+v", c)
		}
	}
	if count != 2 {
		t.Fatalf("created %d tokens, want 2", count)
	}
	crime()
	if effectstest.B16CountNamed(g, "Zombie Rogue") != 2 {
		t.Fatal("Gisa triggered twice in one turn")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	crime()
	if effectstest.B16CountNamed(g, "Zombie Rogue") != 4 {
		t.Fatal("Gisa did not trigger on the next turn")
	}
}

func TestGisaTribalBuffIncludesTypeChangedSource(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opponent := g.Seats[0], g.Seats[1]
	gisa := effectstest.B12Push(g, me.ID, "Gisa, the Hellraiser", "Legendary Creature — Human Warlock", gisaHellraiserOracle, 4, 4)
	skeleton := effectstest.B16Creature(g, me.ID, "Skeleton", "Creature — Skeleton", 2, 2, "B")
	zombie := effectstest.B16Creature(g, me.ID, "Zombie", "Creature — Zombie", 2, 2, "B")
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	other := effectstest.B16Creature(g, opponent.ID, "Other Zombie", "Creature — Zombie", 2, 2, "B")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	for _, id := range []uuid.UUID{skeleton, zombie, bear, other} {
		c := effectstest.B12Card(t, g, id)
		want := id == skeleton || id == zombie
		wantPower := 2
		if want {
			wantPower = 3
		}
		if game.HasKeyword(&c, "menace") != want || c.CurrentPower() != wantPower || c.CurrentToughness() != wantPower {
			t.Fatalf("wrong tribal buff for %s", c.Name)
		}
	}
	// Maskwood Nexus changes Gisa's types in layer 4; the tribe helpers
	// then include her in both layer 6 and layer 7 effects.
	effectstest.PushCatalogPermanent(g, me.ID, "Maskwood Nexus", "Artifact", "9b2cdbed-c733-409b-b0e4-2c8960c25111", false)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.B12Card(t, g, gisa)
	if !game.HasKeyword(&c, "menace") || c.CurrentPower() != 5 || c.CurrentToughness() != 5 {
		t.Fatal("Gisa's text does not exclude herself when she is a Skeleton/Zombie")
	}
}
