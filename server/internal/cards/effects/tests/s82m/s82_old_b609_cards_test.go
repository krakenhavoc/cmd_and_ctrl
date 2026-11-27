package testss82m

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"testing"

	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

// s82_old_b609_cards_test.go — Eighth Edition, life gain and loss (S82 OLD
// batch 609, #5607). Every test plays the real catalog entry through the engine.

const (
	oldB609DeepwoodGhoul    = "d99cd2d0-32b8-4edc-8089-557c48dd4be0"
	oldB609VampiricSpirit   = "dbd1fc9c-6316-4c89-a971-f2e08ca291a4"
	oldB609AvatarOfHope     = "e2f22cef-9494-4356-aa5c-619aecd7e74d"
	oldB609MaggotCarrier    = "e4ad4b81-3685-4f95-84c0-755263b9d3b1"
	oldB609CrystalRod       = "e68bc048-1009-46a5-97d6-ec77a18067da"
	oldB609IronStar         = "e9ec67e1-7064-44d4-a1ed-04b7893ffb15"
	oldB609StaunchDefenders = "ec8bf245-4b2c-432b-9b2a-8d7b9224258c"
	oldB609ThroneOfBone     = "f73c7edf-ed2c-41e8-ac83-c83ddd543f14"
)

func TestS82OldB609EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB609DeepwoodGhoul: "Deepwood Ghoul", oldB609VampiricSpirit: "Vampiric Spirit",
		oldB609AvatarOfHope: "Avatar of Hope", oldB609MaggotCarrier: "Maggot Carrier",
		oldB609CrystalRod: "Crystal Rod", oldB609IronStar: "Iron Star",
		oldB609StaunchDefenders: "Staunch Defenders", oldB609ThroneOfBone: "Throne of Bone",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB609Cast casts a card from the active seat's hand and lets it resolve.
func oldB609Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestS82OldB609DeepwoodGhoulRegeneratesForTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ghoul := effectstest.B12Push(g, me.ID, "Deepwood Ghoul", "Creature — Zombie", oldB609DeepwoodGhoul, 2, 1)
	life := me.Life
	effectstest.B16Activate(t, g, me.ID, ghoul, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	effectstest.B27Kill(g, ghoul)
	g.RunStateChecksForTest()
	if !g.Battlefield.Contains(ghoul) {
		t.Error("the regeneration shield did not save the Ghoul")
	}
	me.Life = 1
	if err := g.ActivateCatalogAbility(me.ID, ghoul, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the ability was activated at 1 life with 2 to pay")
	}
}

func TestS82OldB609VampiricSpiritFliesAndCostsItsCasterFourLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	spirit := oldB609Cast(t, g, "Vampiric Spirit", "Creature — Spirit", oldB609VampiricSpirit, 4, 3)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d", me.Life, life-4)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, spirit, "flying") {
		t.Error("no flying")
	}
}

func TestS82OldB609MaggotCarrierDrainsEveryPlayerOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	oldB609Cast(t, g, "Maggot Carrier", "Creature — Zombie", oldB609MaggotCarrier, 1, 1)
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-1 {
			t.Errorf("seat %v life %d, want %d", p.ID, p.Life, lives[p.ID]-1)
		}
	}
}

func TestS82OldB609StaunchDefendersGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	oldB609Cast(t, g, "Staunch Defenders", "Creature — Human Soldier", oldB609StaunchDefenders, 3, 4)
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
}

func TestS82OldB609AvatarOfHopeCostsSixLessAtThreeOrLessLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	if got := effectstest.SelfPricedMV(t, g, me, oldB609AvatarOfHope, "Creature — Avatar", "{6}{W}{W}"); got != 8 {
		t.Errorf("at %d life: %d, want 8", me.Life, got)
	}
	me.Life = 4
	if got := effectstest.SelfPricedMV(t, g, me, oldB609AvatarOfHope, "Creature — Avatar", "{6}{W}{W}"); got != 8 {
		t.Errorf("at 4 life: %d, want 8", got)
	}
	me.Life = 3
	if got := effectstest.SelfPricedMV(t, g, me, oldB609AvatarOfHope, "Creature — Avatar", "{6}{W}{W}"); got != 2 {
		t.Errorf("at 3 life: %d, want 2", got)
	}
	// It is the caster's life, not an opponent's.
	me.Life, opp.Life = 20, 1
	if got := effectstest.SelfPricedMV(t, g, me, oldB609AvatarOfHope, "Creature — Avatar", "{6}{W}{W}"); got != 8 {
		t.Errorf("opponent at 1 life: %d, want 8", got)
	}
}

func TestS82OldB609AvatarOfHopeFliesAndBlocksAnyNumber(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	avatar := effectstest.PushCatalogPermanent(g, me.ID, "Avatar of Hope", "Creature — Avatar", oldB609AvatarOfHope, false)
	if !effectstest.HasKeywordOnBattlefield(t, g, avatar, "flying") {
		t.Error("no flying")
	}
	if got := effectstest.MbCapacity(t, g, avatar); got != 0 {
		t.Errorf("capacity = %d, want any number", got)
	}
}

func TestS82OldB609LifeArtifactsGainOnlyOnTheirColourAndOnlyWhenPaid(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, color, cost, otherColor, otherCost string
	}{
		{"Crystal Rod", oldB609CrystalRod, "U", "{U}", "R", "{R}"},
		{"Iron Star", oldB609IronStar, "R", "{R}", "B", "{B}"},
		{"Throne of Bone", oldB609ThroneOfBone, "B", "{B}", "U", "{U}"},
	} {
		for _, pay := range []bool{true, false} {
			t.Run(tc.name, func(t *testing.T) {
				g := effectstest.NewCatalogGame(t)
				me, opp := g.Seats[0], g.Seats[1]
				effectstest.AdvanceToMain(t, g)
				effectstest.Mrd2Catalog(g, opp.ID, tc.name, "Artifact", tc.oracle, "{1}", 0, 0)
				life := opp.Life

				off := effectstest.DstHand(me, "Off Colour", "Sorcery", tc.otherCost, tc.otherColor)
				if err := g.CastSpell(me.ID, off, game.CastSpellParams{}); err != nil {
					t.Fatalf("cast the off-colour spell: %v", err)
				}
				if effectstest.TriggerOnStackFrom(g, tc.name) {
					t.Errorf("a spell of another colour triggered %s", tc.name)
				}
				effectstest.PassPriorityAroundTable(t, g)

				on := effectstest.DstHand(me, "On Colour", "Sorcery", tc.cost, tc.color)
				if err := g.CastSpell(me.ID, on, game.CastSpellParams{}); err != nil {
					t.Fatalf("cast the on-colour spell: %v", err)
				}
				effectstest.PassPriorityAroundTable(t, g)
				effectstest.B06AddMana(opp, "C")
				effectstest.AnswerPayUnless(t, g, opp.ID, pay)
				effectstest.PassPriorityAroundTable(t, g)
				want := life
				if pay {
					want++
				}
				if opp.Life != want {
					t.Errorf("an opponent's %s spell, pay=%v: life %d, want %d", tc.color, pay, opp.Life, want)
				}
			})
		}
	}
}
