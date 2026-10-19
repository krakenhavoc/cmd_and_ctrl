package game

import "testing"

// granted_trigger_doubler_test.go — #4137: a CR 603.2d trigger doubler on
// a granted bundle reaches the recipient's composite catalog definition,
// and the entry pre-check that decides whether to recompute the layers
// before a harvest asks the grantor about it.

const (
	gtdGrantorOracle = "fixture-gtd-grantor"
	gtdBundle        = "fixture-gtd/room-doubler"
)

func stubGrantedDoublerCatalog(t testing.TB) {
	t.Helper()
	defs := map[string]*CardDef{
		gtdGrantorOracle: {Static: []StaticAbility{{
			Layer:          Layer6Ability,
			AppliesTo:      creaturesYouControlFixture,
			GrantAbilities: []string{gtdBundle},
		}}},
		GrantKey(gtdBundle): {
			TriggerDoublers: []TriggerDoubler{{
				Label:   "fixture doubler",
				Applies: func(*Game, TriggerDoublingQuery) bool { return true },
			}},
			GrantText: "Abilities trigger an additional time.",
		},
	}
	prev := CatalogLookup
	CatalogLookup = func(key string) *CardDef { return defs[key] }
	t.Cleanup(func() { CatalogLookup = prev })
}

func TestAGrantedTriggerDoublerIsInTheRecipientsDefinition(t *testing.T) {
	stubGrantedDoublerCatalog(t)
	g := newActiveGame(t)
	me, opp := g.Seats[0].ID, g.Seats[1].ID
	bear := gaBear(g, me)
	theirs := gaBear(g, opp)
	gaPush(g, me, "Grantor", "Enchantment", gtdGrantorOracle)

	mine := layeredBattlefieldCard(t, g, bear)
	got := CatalogTriggerDoublers(CatalogAbilityKey(mine))
	if len(got) != 1 || got[0].Label != "fixture doubler" {
		t.Fatalf("the recipient's doublers = %+v, want the granted one", got)
	}
	other := layeredBattlefieldCard(t, g, theirs)
	if got := CatalogTriggerDoublers(CatalogAbilityKey(other)); len(got) != 0 {
		t.Errorf("an opponent's creature has %d granted doublers, want none", len(got))
	}
}

func TestTheEntryPreCheckAsksTheGrantorAboutAGrantedDoubler(t *testing.T) {
	stubGrantedDoublerCatalog(t)
	g := newActiveGame(t)
	me := g.Seats[0].ID
	gaBear(g, me)
	has := func() bool {
		var out bool
		g.ReadSnapshot(func() { out = g.hasTriggerModifierLocked() })
		return out
	}
	if has() {
		t.Fatal("a board with no grantor reports a trigger modifier")
	}
	gaPush(g, me, "Grantor", "Enchantment", gtdGrantorOracle)
	if !has() {
		t.Error("a board with a grantor of a doubler does not report a trigger modifier")
	}
}
