package testso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const ogreBattlecasterOracle = "00ba0c24-a671-493e-ba46-13e45d1818f1"

func TestOgreBattlecasterIsRegisteredFull(t *testing.T) {
	effectstest.RequireFullNoCaveats(t, ogreBattlecasterOracle, "Ogre Battlecaster")
}

// ogreAttack attacks with the Ogre, targets the graveyard sorcery with its
// trigger and resolves it up to the cast offer.
func ogreAttack(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, *game.PendingChoice) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ogre := effectstest.PushCatalogPermanent(g, me.ID, "Ogre Battlecaster", "Creature — Ogre Shaman", ogreBattlecasterOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == ogre {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 3
			}
		}
	})
	spell := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.SpellControlDivinationOracle)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(ogre, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B04WaitForPickOrSkip(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ResolutionCastOfferFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no offer to cast the targeted card")
	}
	return g, me, ogre, spell, prompt
}

func ogrePower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	var p int
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(id)
		info, ok := g.PermanentForEffect(game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
		if !ok {
			t.Fatal("the Ogre is gone")
		}
		p = info.Power
	})
	return p
}

// The spell costs its mana cost plus {R}{R}: {2}{U} alone is refused,
// {2}{U}{R}{R} casts it. Casting it gives the Ogre +3/+0, and the spell is
// exiled after it resolves.
func TestOgreBattlecasterCastsForRRMoreAndPumps(t *testing.T) {
	g, me, ogre, spell, prompt := ogreAttack(t)
	params := game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: game.AltCostKeyGrantedCost}
	effectstest.FloatForTest(g, me, "CCU")
	if effectstest.EnumeratedFor(g, me.ID, spell) {
		t.Error("the bot is offered the cast without the {R}{R}")
	}
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, spell, params); err == nil {
		t.Fatal("cast without the {R}{R}")
	}
	effectstest.FloatForTest(g, me, "RR")
	if !effectstest.EnumeratedFor(g, me.ID, spell) {
		t.Error("the bot is not offered the cast with {2}{U}{R}{R}")
	}
	if err := g.CastForResolutionChoice(me.ID, prompt.ID, spell, params); err != nil {
		t.Fatalf("cast for {2}{U}{R}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g) // the delayed trigger
	if got := ogrePower(t, g, ogre); got != 6 {
		t.Errorf("Ogre power %d, want 6", got)
	}
	effectstest.PassPriorityAroundTable(t, g) // the spell
	if z := effectstest.E2Zone(g, spell); z != game.ZoneExile {
		t.Errorf("the spell is in %q after resolving, want exile", z)
	}
}

// Declining leaves the card in the graveyard and the Ogre unpumped.
func TestOgreBattlecasterDeclineLeavesTheCard(t *testing.T) {
	g, me, ogre, spell, _ := ogreAttack(t)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if z := effectstest.E2Zone(g, spell); z != game.ZoneGraveyard {
		t.Errorf("the card is in %q, want the graveyard", z)
	}
	if got := ogrePower(t, g, ogre); got != 3 {
		t.Errorf("Ogre power %d, want 3", got)
	}
}
