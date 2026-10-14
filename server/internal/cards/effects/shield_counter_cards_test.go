package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// shield_counter_cards_test.go — #2000, #3246: Protection Magic and
// Summon: Magus Sisters put CR 122.1c shield counters on creatures.

const (
	protectionMagicOracle = "1380f0cc-e875-46ec-a802-c6684bf3965c"
	magusSistersOracle    = "06f11ff8-cc99-411d-b00b-464ba4b7b87d"
)

func TestProtectionMagicShieldsUpToThreeCreatures(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	a := tdCreature(g, me.ID, "Bear", 2, 2)
	b := tdCreature(g, me.ID, "Wolf", 2, 2)
	theirs := tdCreature(g, opp.ID, "Ogre", 3, 3)
	untouched := tdCreature(g, me.ID, "Elf", 1, 1)

	targets := append(append(tdCard(a), tdCard(b)...), tdCard(theirs)...)
	tdCastFromHand(t, g, game.Card{Name: "Protection Magic", TypeLine: "Instant",
		OracleID: protectionMagicOracle, ManaCost: "{1}{W}"}, game.CastSpellParams{Targets: targets})
	passPriorityAroundTable(t, g)

	for _, id := range []struct {
		name string
		id   uuid.UUID
		want int
	}{{"Bear", a, 1}, {"Wolf", b, 1}, {"Ogre", theirs, 1}, {"Elf", untouched, 0}} {
		if got := countersOn(g, id.id, game.CounterShield); got != id.want {
			t.Errorf("%s: %d shield counters, want %d", id.name, got, id.want)
		}
	}

	b25Destroy(g, a)
	if !g.Battlefield.Contains(a) || countersOn(g, a, game.CounterShield) != 0 {
		t.Error("the shield counter didn't stand in for the destruction")
	}
	b25Destroy(g, a)
	if g.Battlefield.Contains(a) {
		t.Error("with its counter spent, the Bear is destroyed")
	}
}

// Each chapter draws one of the three bullets; all three are reachable,
// and each does what it says to an opponent's creature.
func TestSummonMagusSistersChapterDrawsABulletAtRandom(t *testing.T) {
	seen := map[string]bool{}
	for key := byte(0); key < 48 && len(seen) < 3; key++ {
		g := newCatalogGame(t)
		g.SetRNGKeyForTest(randomModesKey(key))
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		ogre := tdCreature(g, opp.ID, "Ogre", 2, 2)

		sisters := tdCastFromHand(t, g, game.Card{Name: "Summon: Magus Sisters", TypeLine: "Enchantment Creature — Saga Faerie",
			OracleID: magusSistersOracle, ManaCost: "{4}{G}", Power: 5, Toughness: 5}, game.CastSpellParams{})
		passPriorityAroundTable(t, g)
		if c := modePickChoiceFor(g, me.ID); c != nil {
			t.Fatalf("a random mode asks nobody to choose: %+v", c)
		}
		life := me.Life
		b04WaitForPick(t, g, me.ID)
		pickCard(t, g, me.ID, ogre)
		passPriorityAroundTable(t, g)

		switch {
		case countersOn(g, ogre, game.CounterPlusOne) == 3:
			seen["combine"] = true
		case countersOn(g, ogre, game.CounterShield) == 1:
			seen["defense"] = true
			if me.Life != life+3 {
				t.Errorf("key %d Defense!: gained %d life, want 3", key, me.Life-life)
			}
		case !g.Battlefield.Contains(ogre):
			seen["fight"] = true
			if dmg := b12Card(t, g, sisters).DamageMarked; dmg != 2 {
				t.Errorf("key %d Fight!: the Sisters took %d damage, want 2", key, dmg)
			}
		default:
			t.Fatalf("key %d: chapter I did none of its three bullets", key)
		}
	}
	if len(seen) != 3 {
		t.Errorf("48 draws reached %v; a uniform pick reaches all three", seen)
	}
}

// A shield counter's prevention watches damage (CR 122.1c), so a board
// with one keeps the trigger-order prompt the two pingers otherwise skip
// (TestTwoPingersOnOneEntryDoNotAsk).
func TestTwoPingersStillAskBesideAShieldCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Impact Tremors", "Enchantment", impactTremorsOracle, false)
	pushCatalogPermanent(g, me.ID, "Molten Gatekeeper", "Creature — Elemental", moltenGatekeeperOracle, false)
	shielded := tdCreature(g, opp.ID, "Ogre", 3, 3)
	g.WithWriteLock(func() {
		if err := g.AddCounterByForEffect(opp.ID, shielded, game.CounterShield, 1); err != nil {
			t.Fatal(err)
		}
	})
	b20CastCreature(t, g, me, "Walking Corpse", "Creature — Zombie", "", 2, 2)
	if ch := settleUntilOrderPrompt(t, g); ch == nil {
		t.Fatal("a shield counter on the board must keep the prompt")
	}
}
