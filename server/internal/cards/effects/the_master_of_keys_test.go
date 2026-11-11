package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const theMasterOfKeysOracle = "37dc566e-a91a-448c-9967-b178a93ffb62"

func TestTheMasterOfKeysIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(theMasterOfKeysOracle)
	if !ok || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("The Master of Keys: %+v %v", spec, ok)
	}
}

func TestTheMasterOfKeysEntersWithXCountersAndMillsTwiceX(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib, gy := me.Library.Size(), me.Graveyard.Size()
	id, err := spmB4Cast(t, g, "The Master of Keys", "Legendary Enchantment Creature — Horror", theMasterOfKeysOracle, "{X}{W}{U}{B}", 3, 3, 2)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, id); n != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", n)
	}
	if got := lib - me.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4 (twice X)", got)
	}
	if me.Graveyard.Size() < gy+4 {
		t.Errorf("graveyard %d → %d, want at least +4", gy, me.Graveyard.Size())
	}
}

func TestTheMasterOfKeysForXZeroDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lib := me.Library.Size()
	id, err := spmB4Cast(t, g, "The Master of Keys", "Legendary Enchantment Creature — Horror", theMasterOfKeysOracle, "{X}{W}{U}{B}", 3, 3, 0)
	if err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := plusOnes(g, id); n != 0 {
		t.Errorf("+1/+1 counters = %d, want 0", n)
	}
	if me.Library.Size() != lib {
		t.Errorf("library %d → %d, want unchanged", lib, me.Library.Size())
	}
}

// The standing escape: enchantment cards only, gone when the Master leaves.
func TestTheMasterOfKeysGrantsEscapeToEnchantmentCardsOnly(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	master := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Master of Keys", TypeLine: "Legendary Enchantment Creature — Horror",
		OracleID: theMasterOfKeysOracle, Power: 3, Toughness: 3, Owner: active.ID, Controller: active.ID,
	})
	ench := seedGraveyardCard(t, g, "Dead Enchantment", "Enchantment", "test-keys-enchantment")
	perm := grantedPermissionOn(g, active.ID, ench, game.ZoneGraveyard)
	if !perm.Granted() {
		t.Fatal("an enchantment card in the graveyard has no escape")
	}
	if perm.AltCostKey != "escape" || perm.ExileOtherFromGraveyard != 3 || perm.Cost != "" {
		t.Errorf("granted escape = key %q, exile %d, cost %q, want escape / 3 / printed cost", perm.AltCostKey, perm.ExileOtherFromGraveyard, perm.Cost)
	}
	creature := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "test-keys-bear")
	if p := grantedPermissionOn(g, active.ID, creature, game.ZoneGraveyard); p.Granted() {
		t.Error("a plain creature card gained escape")
	}
	instant := seedGraveyardCard(t, g, "Dead Bolt", "Instant", "test-keys-bolt")
	if p := grantedPermissionOn(g, active.ID, instant, game.ZoneGraveyard); p.Granted() {
		t.Error("an instant card gained escape")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(master) })
	if p := grantedPermissionOn(g, active.ID, ench, game.ZoneGraveyard); p.Granted() {
		t.Error("escape survived The Master of Keys leaving the battlefield")
	}
}
