package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b841MonsoonOracle  = "a86e7de8-8448-406c-b162-22c62bdc3813"
	b841SupplicantOrcl = "c28a7641-49e3-43b1-a366-8d6e9c8dab9c"
	b841WispsOracle    = "ce65c6ed-e5d4-48d8-b569-4e2774b2912c"
	b841ZombiesOracle  = "d0d844f2-6cf4-49f2-95db-64e932a15b17"
	b841MeteorOracle   = "d2110083-b30f-43f1-9352-7536f5e19d3c"
	b841YetiOracle     = "d6d84884-4a58-4187-aa5c-20b39b379590"
	b841CatapultOracle = "eef931d8-4048-4c33-bd8c-0f67d1083ee6"
	b841SnowSwampType  = "Basic Snow Land — Swamp"
	b841PlainSwampType = "Basic Land — Swamp"
	b841IslandType     = "Basic Land — Island"
	b841ActivateNoMana = ""
)

func b841Activate(t *testing.T, g *game.Game, who *game.Player, id uuid.UUID, mana string, params game.ActivateAbilityParams) error {
	t.Helper()
	effectstest.FloatForTest(g, who, mana)
	params.Strict = true
	return g.ActivateCatalogAbility(who.ID, id, 0, params)
}

// Monsoon: at each player's end step that player's untapped Islands are tapped and the
// player takes one damage per Island tapped this way; a tapped Island and a non-Island
// are not counted.
func TestB841MonsoonTapsTheActivePlayersIslandsAndDamagesThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, g.Seats[1].ID, "Monsoon", "Enchantment", b841MonsoonOracle, false)
	i1 := effectstest.B12Push(g, me.ID, "Island", b841IslandType, "", 0, 0)
	i2 := effectstest.B12Push(g, me.ID, "Island", b841IslandType, "", 0, 0)
	tapped := effectstest.B12Push(g, me.ID, "Island", b841IslandType, "", 0, 0)
	forest := effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(tapped) })
	life := me.Life
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 (two Islands tapped this way)", life, me.Life)
	}
	if !effectstest.B13Tapped(t, g, i1) || !effectstest.B13Tapped(t, g, i2) {
		t.Error("untapped Islands were not tapped")
	}
	if effectstest.B13Tapped(t, g, forest) {
		t.Error("a Forest was tapped")
	}
}

// Freyalise Supplicant: only a red or white creature can be sacrificed; the damage is half its power, rounded down.
func TestB841FreyaliseSupplicantDamagesHalfTheSacrificedPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	supp := effectstest.B12Push(g, me.ID, "Freyalise Supplicant", "Creature — Human Cleric", b841SupplicantOrcl, 1, 1)
	red := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Red Ogre", TypeLine: "Creature — Ogre", Power: 5, Toughness: 5, Colors: []string{"R"}})
	green := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	if err := b841Activate(t, g, me, supp, b841ActivateNoMana, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), SacrificeIDs: []uuid.UUID{green}}); err == nil {
		t.Fatal("a green creature paid the cost")
	}
	life := opp.Life
	if err := b841Activate(t, g, me, supp, b841ActivateNoMana, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), SacrificeIDs: []uuid.UUID{red}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 (half of 5, rounded down)", life, opp.Life)
	}
}

// Withering Wisps: the cap is the number of snow Swamps, and it dies at an end step with no creatures.
func TestB841WitheringWispsIsCappedBySnowSwamps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wisps := effectstest.PushCatalogPermanent(g, me.ID, "Withering Wisps", "Enchantment", b841WispsOracle, false)
	effectstest.B12Push(g, me.ID, "Swamp", b841PlainSwampType, "", 0, 0)
	if err := b841Activate(t, g, me, wisps, "B", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated with no snow Swamp")
	}
	effectstest.B12Push(g, me.ID, "Snow-Covered Swamp", b841SnowSwampType, "", 0, 0)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	life := me.Life
	if err := b841Activate(t, g, me, wisps, "BB", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first activation: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want -1", life, me.Life)
	}
	if !g.Battlefield.Contains(bear) {
		t.Error("a 2/2 died to one damage")
	}
	if err := b841Activate(t, g, me, wisps, "B", game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a second activation with one snow Swamp")
	}
}

func TestB841WitheringWispsIsSacrificedWithNoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wisps := effectstest.PushCatalogPermanent(g, me.ID, "Withering Wisps", "Enchantment", b841WispsOracle, false)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wisps) {
		t.Error("no creatures on the battlefield at the end step: sacrificed")
	}
}

// Gangrenous Zombies: 1 damage to each creature and player, 2 with a snow Swamp; the Zombies are sacrificed.
func TestB841GangrenousZombiesDamageDependsOnSnowSwamp(t *testing.T) {
	for _, snow := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		z := effectstest.B12Push(g, me.ID, "Gangrenous Zombies", "Creature — Zombie", b841ZombiesOracle, 2, 2)
		bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		want := 1
		if snow {
			effectstest.B12Push(g, me.ID, "Snow-Covered Swamp", b841SnowSwampType, "", 0, 0)
			want = 2
		}
		life := opp.Life
		if err := b841Activate(t, g, me, z, b841ActivateNoMana, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Life != life-want {
			t.Errorf("snow=%v: life %d -> %d, want -%d", snow, life, opp.Life, want)
		}
		if g.Battlefield.Contains(z) {
			t.Error("the Zombies were not sacrificed")
		}
		if g.Battlefield.Contains(bear) != !snow {
			t.Errorf("snow=%v: a 2/2 should die only to 2 damage", snow)
		}
	}
}

// Meteor Shower: X plus 1 damage divided; a split that does not total X+1 is refused.
func TestB841MeteorShowerDividesXPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	if err := effectstest.D1658CastErr(t, g, "Meteor Shower", "Sorcery", b841MeteorOracle, game.CastSpellParams{
		XValue: 2, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1},
	}); err == nil {
		t.Fatal("a split of 2 was accepted for X=2 (3 to divide)")
	}
	effectstest.D1658Cast(t, g, "Meteor Shower", "Sorcery", b841MeteorOracle, game.CastSpellParams{
		XValue: 2, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.E2Card(t, g, a).DamageMarked; d != 1 {
		t.Errorf("a: %d damage, want 1", d)
	}
	if d := effectstest.E2Card(t, g, b).DamageMarked; d != 2 {
		t.Errorf("b: %d damage, want 2", d)
	}
}

// Karplusan Yeti: mutual damage equal to each creature's power.
func TestB841KarplusanYetiTradesBlowsWithItsTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sick := effectstest.PushCatalogPermanent(g, me.ID, "Karplusan Yeti", "Creature — Yeti", b841YetiOracle, true)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
	if err := b841Activate(t, g, me, sick, b841ActivateNoMana, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("a summoning-sick Yeti paid {T}")
	}
	yeti := effectstest.B12Push(g, me.ID, "Karplusan Yeti", "Creature — Yeti", b841YetiOracle, 3, 3)
	if err := b841Activate(t, g, me, yeti, b841ActivateNoMana, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 2/3 took 3 and survived")
	}
	if d := effectstest.E2Card(t, g, yeti).DamageMarked; d != 2 {
		t.Errorf("Yeti took %d, want 2", d)
	}
}

// Skull Catapult: {1}, {T}, sacrifice a creature: 2 damage to any target; usable the turn it arrives.
func TestB841SkullCatapultSacrificesACreatureForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cat := effectstest.PushCatalogPermanent(g, me.ID, "Skull Catapult", "Artifact", b841CatapultOracle, true)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Rat", 1, 1)
	if err := b841Activate(t, g, me, cat, b841ActivateNoMana, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated without paying {1}")
	}
	life := opp.Life
	if err := b841Activate(t, g, me, cat, "C", game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 || g.Battlefield.Contains(fodder) {
		t.Errorf("life %d -> %d (want -2), fodder on battlefield %v", life, opp.Life, g.Battlefield.Contains(fodder))
	}
}
