package effects

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestFieryAnnihilationRelatedEquipmentAndDeathRider(t *testing.T) {
	for _, change := range []string{"none", "detach", "hexproof", "omit"} {
		t.Run(change, func(t *testing.T) {
			g := newCatalogGame(t)
			opp := g.Seats[1]
			creature := pushCreatureToBattlefieldForTest(g, opp.ID, "Bear")
			equipment := seedPermanentFor(g, opp.ID, "Equipment", "Artifact — Equipment")
			g.WithWriteLock(func() {
				if err := g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: creature}); err != nil {
					t.Fatal(err)
				}
			})
			targets := []game.TargetRef{{Kind: game.TargetCard, ID: creature}}
			if change != "omit" {
				targets = append(targets, game.TargetRef{Kind: game.TargetCard, ID: equipment, Slot: 1})
			}
			castCatalogSpell(t, g, "Fiery Annihilation", "Instant", "d44f3724-17b7-48f0-885d-75292669f971", targets)
			g.WithWriteLock(func() {
				if change == "hexproof" {
					if err := (ScopedEffectFor{Target: creature, Mods: []game.Mod{game.AddKeywordsMod("hexproof")}, Duration: game.IndefiniteDuration()}).Apply(NewContext(g, nil)); err != nil {
						t.Fatal(err)
					}
				}
				for i := range g.Battlefield.Cards {
					c := &g.Battlefield.Cards[i]
					if c.InstanceID == equipment && change == "detach" {
						c.AttachedTo = game.TargetRef{}
					}
				}
			})
			passPriorityAroundTable(t, g)
			if change == "hexproof" {
				if !g.Battlefield.Contains(creature) || !g.Battlefield.Contains(equipment) {
					t.Fatal("an illegal creature target still affected the creature or Equipment")
				}
				return
			}
			if !g.Exile.Contains(creature) || opp.Graveyard.Contains(creature) {
				t.Fatal("lethal damage missed the exile-instead-of-dying rider")
			}
			if got := g.Exile.Contains(equipment); got != (change == "none") {
				t.Fatalf("Equipment exile for %s: %v", change, got)
			}
		})
	}
}

func TestFieryAnnihilationRejectsUnattachedEquipment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	creature := pushCreatureToBattlefieldForTest(g, opp.ID, "Bear")
	equipment := seedPermanentFor(g, opp.ID, "Equipment", "Artifact — Equipment")
	spell := game.NewCard("Fiery Annihilation", me.ID)
	spell.TypeLine, spell.OracleID = "Instant", "d44f3724-17b7-48f0-885d-75292669f971"
	me.Hand.PushTop(spell)
	advanceToMain(t, g)
	err := g.CastSpell(me.ID, spell.InstanceID, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: creature}, {Kind: game.TargetCard, ID: equipment, Slot: 1}}})
	if !errors.Is(err, game.ErrIllegalTarget) || !me.Hand.Contains(spell.InstanceID) {
		t.Fatalf("unattached Equipment was accepted: %v", err)
	}
}
