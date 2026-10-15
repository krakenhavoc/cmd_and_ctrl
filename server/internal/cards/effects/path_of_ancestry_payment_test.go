package effects

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #2033 and #3844 (ADR 0068 amendments 2026-10-12 and 2026-10-15):
// Path of Ancestry's {G} pays whenever it can, in every payment mode —
// a funded cast, an insufficient permissive pool, and Cast anyway — and
// its scry fires whenever it really paid for a creature spell that
// shares a type with the commander. A token the cost cannot use is not
// spent, stays in the pool, and does not scry.
func TestPathOfAncestryScriesWheneverItsManaPays(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, cost string
		params               game.CastSpellParams
		spendsPath           bool
		scry, onPaper        bool
	}{
		{"shared creature type", "Creature — Elf Warrior", "{G}", game.CastSpellParams{}, true, true, false},
		{"different creature type", "Creature — Goblin", "{G}", game.CastSpellParams{}, true, false, false},
		{"not a creature", "Sorcery", "{G}", game.CastSpellParams{}, true, false, false},
		{"insufficient pool", "Creature — Elf Warrior", "{1}{G}", game.CastSpellParams{}, true, true, true},
		{"insufficient pool, generic", "Creature — Elf Warrior", "{3}", game.CastSpellParams{}, true, true, true},
		{"cast anyway, pool covers it", "Creature — Elf Warrior", "{G}", game.CastSpellParams{Strict: true, ForceCast: true}, true, true, false},
		{"cast anyway, pool short", "Creature — Elf Warrior", "{2}{G}", game.CastSpellParams{Strict: true, ForceCast: true}, true, true, true},
		{"cast anyway, Goblin", "Creature — Goblin", "{2}{G}", game.CastSpellParams{Strict: true, ForceCast: true}, true, false, true},
		{"insufficient pool, wrong colour", "Creature — Elf Warrior", "{B}{B}", game.CastSpellParams{}, false, false, true},
		{"mana not spent", "Creature — Elf Warrior", "{0}", game.CastSpellParams{}, false, false, false},
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
			paid := g.StackMeta[spell].Paid
			if paid.OnPaper != tc.onPaper {
				t.Fatalf("payment = %+v, want onPaper %v", paid, tc.onPaper)
			}
			if tc.spendsPath {
				if len(paid.Mana) != 1 || paid.Mana[0].Source != path || len(me.ManaPool) != 0 {
					t.Fatalf("the payment did not record Path's token: %+v (pool %+v)", paid.Mana, me.ManaPool)
				}
			} else {
				if len(paid.Mana) != 0 {
					t.Fatalf("the payment recorded %+v, want nothing", paid.Mana)
				}
				if !reflect.DeepEqual(me.ManaPool, before) {
					t.Fatal("an unspent Path token left the pool")
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

// Path's caveat is gone: its token pays whenever it can, so the scry no
// longer depends on the payment mode (#3844).
func TestPathOfAncestryIsFull(t *testing.T) {
	spec, ok := Lookup(pathOfAncestryOracle)
	if !ok {
		t.Fatal("Path of Ancestry is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Path of Ancestry is %v with caveats %q, want full with none", spec.Completeness, spec.Caveats)
	}
}
