package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestPathOfAncestryScriesForRecordedPermissivePayment(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		params               game.CastSpellParams
		scry, onPaper        bool
	}{
		{"shared creature type", "Creature — Elf Warrior", "{G}", game.CastSpellParams{}, true, false},
		{"different creature type", "Creature — Goblin", "{G}", game.CastSpellParams{}, false, false},
		{"not a creature", "Sorcery", "{G}", game.CastSpellParams{}, false, false},
		{"insufficient pool", "Creature — Elf Warrior", "{1}{G}", game.CastSpellParams{}, false, true},
		{"cast anyway", "Creature — Elf Warrior", "{G}", game.CastSpellParams{Strict: true, ForceCast: true}, false, true},
		{"mana not spent", "Creature — Elf Warrior", "{0}", game.CastSpellParams{}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			g.WithWriteLock(func() {
				me.Command.PushTop(game.Card{
					InstanceID: uuid.New(), Name: "Lathril, Blade of the Elves",
					TypeLine: "Legendary Creature — Elf Noble", ManaCost: "{2}{B}{G}",
					Owner: me.ID, Controller: me.ID, IsCommander: true,
				})
				me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID})
			})
			path := seedPermanentWithOracle(g, me.ID, "Path of Ancestry", "Land", pathOfAncestryOracle)
			activateManaFor(t, g, me.ID, path, 0, game.ManaAbilityParams{Colors: []string{"G"}})
			before := append(game.ManaPool(nil), me.ManaPool...)
			spell := castFromHandForTest(t, g, me, "Test spell", tc.typeLine, tc.cost, "", tc.params)
			if paid := g.StackMeta[spell].Paid; paid.OnPaper != tc.onPaper {
				t.Fatalf("payment = %+v, want onPaper %v", paid, tc.onPaper)
			}
			if (tc.onPaper || tc.cost == "{0}") && !reflect.DeepEqual(me.ManaPool, before) {
				t.Fatal("an unspent Path token left the pool")
			}
			if !tc.onPaper && tc.cost != "{0}" {
				paid := g.StackMeta[spell].Paid.Mana
				if len(paid) != 1 || paid[0].Source != path || len(me.ManaPool) != 0 {
					t.Fatalf("funded payment did not record Path's token: %+v", paid)
				}
			}
			passPriorityAroundTable(t, g)
			scried := false
			for _, choice := range g.PendingChoices {
				if choice != nil && choice.Kind == game.PendingChoiceScry {
					scried = true
				}
			}
			if scried != tc.scry {
				t.Errorf("scry queued = %v, want %v", scried, tc.scry)
			}
		})
	}
}
