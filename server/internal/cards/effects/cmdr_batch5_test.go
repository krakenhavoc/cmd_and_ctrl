package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdrSoullessOneOracle  = "27d98b94-a8a1-442b-8940-4ffab51b5164"
	cmdrTendrilsCorrOracle = "84ee0da2-d886-4603-9aaa-bc5326a424d7"
	cmdrBreathWeaponOracle = "f3dd1f6d-f7d7-4358-8139-7495404f29c7"
	cmdrMagmaquakeOracle   = "a52037df-0be1-4c99-b9eb-b033157af203"
)

// cmdrB5Push seeds a catalog permanent under owner and returns its id.
func cmdrB5Push(g *game.Game, owner uuid.UUID, name, oracle, typeLine string, p, t int, keywords ...string) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		OracleID:   oracle,
		TypeLine:   typeLine,
		Power:      p,
		Toughness:  t,
		Keywords:   keywords,
		Owner:      owner,
		Controller: owner,
	})
}

func cmdrPT(t *testing.T, g *game.Game, id uuid.UUID) (int, int) {
	t.Helper()
	c := layeredCard(t, g, id)
	return c.CurrentPower(), c.CurrentToughness()
}

func TestSoullessOneCountsZombiesOnBattlefieldAndInGraveyards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	so := cmdrB5Push(g, me.ID, "Soulless One", cmdrSoullessOneOracle, "Creature — Zombie Avatar", 0, 0)
	if p, tt := cmdrPT(t, g, so); p != 1 || tt != 1 {
		t.Fatalf("alone it is %d/%d, want 1/1 (it is a Zombie)", p, tt)
	}
	cmdrB5Push(g, opp.ID, "Walker", "", "Creature — Zombie", 2, 2)
	cmdrB5Push(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Zombie", TypeLine: "Creature — Zombie Knight", Owner: opp.ID, Controller: opp.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	g.RunStateChecksForTest()
	if p, tt := cmdrPT(t, g, so); p != 3 || tt != 3 {
		t.Errorf("Soulless One is %d/%d, want 3/3 (two Zombies on the battlefield, one in a graveyard)", p, tt)
	}
}

func TestTendrilsOfCorruptionDamagesAndGainsPerSwamp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 3; i++ {
		cmdrB5Push(g, me.ID, "Swamp", "", "Basic Land — Swamp", 0, 0)
	}
	cmdrB5Push(g, opp.ID, "Swamp", "", "Basic Land — Swamp", 0, 0)
	big := cmdrB5Push(g, opp.ID, "Ogre", "", "Creature — Ogre", 4, 4)
	before := me.Life
	castCatalogSpell(t, g, "Tendrils of Corruption", "Instant", cmdrTendrilsCorrOracle, []game.TargetRef{{Kind: game.TargetCard, ID: big}})
	passPriorityAroundTable(t, g)
	if got := damageMarkedOn(g, big); got != 3 {
		t.Errorf("damage = %d, want 3 (only your Swamps count)", got)
	}
	if me.Life != before+3 {
		t.Errorf("life %d -> %d, want +3", before, me.Life)
	}
}

func TestTendrilsOfCorruptionRefusesALandTarget(t *testing.T) {
	g := newCatalogGame(t)
	land := cmdrB5Push(g, g.Seats[1].ID, "Swamp", "", "Basic Land — Swamp", 0, 0)
	err := castCatalogSpellErr(t, g, "Tendrils of Corruption", "Instant", cmdrTendrilsCorrOracle, []game.TargetRef{{Kind: game.TargetCard, ID: land}})
	if err == nil {
		t.Fatal("a non-creature target was accepted")
	}
}

func TestBreathWeaponSparesDragons(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := cmdrB5Push(g, opp.ID, "Drake", "", "Creature — Dragon", 3, 2, "flying")
	bear := cmdrB5Push(g, opp.ID, "Bear", "", "Creature — Bear", 2, 2)
	mine := cmdrB5Push(g, me.ID, "Elf", "", "Creature — Elf", 3, 3)
	castCatalogSpell(t, g, "Breath Weapon", "Instant", cmdrBreathWeaponOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear should die")
	}
	if !g.Battlefield.Contains(dragon) || damageMarkedOn(g, dragon) != 0 {
		t.Error("the Dragon is untouched")
	}
	if damageMarkedOn(g, mine) != 2 {
		t.Errorf("your own creature takes 2, has %d", damageMarkedOn(g, mine))
	}
}

func TestMagmaquakeHitsGroundCreaturesAndPlaneswalkers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := cmdrB5Push(g, opp.ID, "Bear", "", "Creature — Bear", 2, 2)
	bird := cmdrB5Push(g, opp.ID, "Bird", "", "Creature — Bird", 2, 2, "flying")
	mine := cmdrB5Push(g, me.ID, "Ogre", "", "Creature — Ogre", 4, 4)
	pw := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 5}, Owner: opp.ID, Controller: opp.ID})
	before := opp.Life
	castXSpell(t, g, "Magmaquake", "Instant", cmdrMagmaquakeOracle, "{X}{R}{R}", 2, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a ground 2/2 dies to X=2")
	}
	if !g.Battlefield.Contains(bird) || damageMarkedOn(g, bird) != 0 {
		t.Error("a flyer is untouched")
	}
	if damageMarkedOn(g, mine) != 2 {
		t.Error("your own ground creature takes it too")
	}
	if got := countersOn(g, pw, game.CounterLoyalty); got != 3 {
		t.Errorf("planeswalker loyalty = %d, want 3", got)
	}
	if opp.Life != before {
		t.Error("players take no damage")
	}
}
