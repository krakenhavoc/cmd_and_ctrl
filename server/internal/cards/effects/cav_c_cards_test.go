package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cav_c_cards_test.go — the proof for #3172 slice cav-c: each test fails
// on the engine or card behaviour the slice added.

const (
	cavCBerserkersOnslaughtOracle = "85d2948c-1a79-418d-80c4-bc1012a4d313"
	cavCArchmageOracle            = "8305d576-21d8-4ce7-8eda-a7cd9793aca5"
	cavCSedgemoorWitchOracle      = "25dce517-ac0a-4577-89ed-04296c7c4069"
)

// The controller of a "tapped and attacking" token chooses what it
// attacks (CR 508.4): Leonin Warleader's two Cats can go at different
// opponents, and the prompt opens on the player the Warleader attacked.
func TestCavCWarleaderCatsAttackWhomeverYouChoose(t *testing.T) {
	g := newCatalogGame(t)
	me, first, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	warleader := b31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", b31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	// Attack the SECOND opponent in seat order: the prompt must open on
	// them (what the Warleader attacked), not on the first seat.
	declareAttack(t, g, second.ID, warleader)
	passPriorityAroundTable(t, g)
	prompt := payAmountPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("with three opponents to attack the controller is asked where the Cats go")
	}
	if prompt.PayAmount.Max != 2 || prompt.PayAmount.Goal != 2 {
		t.Errorf("first prompt 0..%d goal %d, want 0..2 goal 2 (all at what the Warleader attacked)", prompt.PayAmount.Max, prompt.PayAmount.Goal)
	}
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 1); err != nil {
		t.Fatalf("one at the attacked opponent: %v", err)
	}
	prompt = payAmountPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("asked about the next opponent")
	}
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 0); err != nil {
		t.Fatalf("none at the next: %v", err)
	}
	passPriorityAroundTable(t, g)
	at := map[uuid.UUID]int{}
	for _, id := range battlefieldIDsNamed(g, "Cat") {
		at[cardByID(g, id).AttackingTarget]++
	}
	if at[second.ID] != 1 || at[first.ID] != 0 || at[third.ID] != 1 {
		t.Errorf("Cats attacking %v: want one at the attacked opponent and one at the last seat", at)
	}
}

// "Attacking creatures you control have double strike" is a layer 6
// static now: a Cat put onto the battlefield attacking has it the moment
// it arrives (the old per-attacker trigger only reached creatures that
// were DECLARED), and loses it when combat ends.
func TestCavCOnslaughtReachesCreaturesThatEnterAttacking(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
	}{
		{"Berserkers' Onslaught", cavCBerserkersOnslaughtOracle, "Enchantment"},
		{"Blade Historian", b25BladeHistorianOracle, "Creature — Human Cleric"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			pushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			warleader := b31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", b31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
			home := pushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
			declareAttack(t, g, opp.ID, warleader)
			passAndSettleAttackSplit(t, g, me.ID)
			cats := battlefieldIDsNamed(g, "Cat")
			if len(cats) != 2 {
				t.Fatalf("%d Cats, want 2", len(cats))
			}
			for _, id := range cats {
				if !hasEffectiveKeyword(t, g, id, "double strike") {
					t.Error("a Cat that entered attacking has double strike")
				}
			}
			if !hasEffectiveKeyword(t, g, warleader, "double strike") {
				t.Error("the declared attacker has double strike")
			}
			if hasEffectiveKeyword(t, g, home, "double strike") {
				t.Error("a creature that is not attacking does not")
			}
			advanceTo(t, g, game.StepCombatDamage)
			passPriorityAroundTable(t, g)
			advanceTo(t, g, game.StepPostcombatMain)
			for _, id := range cats {
				if hasEffectiveKeyword(t, g, id, "double strike") {
					t.Error("combat is over: nothing is attacking, nothing has double strike")
				}
			}
		})
	}
}

// Neyali's static covers an attacking TOKEN the moment it appears, and
// only a token: the Warleader that made the Cats is not one.
func TestCavCNeyaliGivesAttackingTokensDoubleStrikeAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", b32NeyaliOracle, 3, 3)
	warleader := b31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", b31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	declareAttack(t, g, opp.ID, warleader)
	passAndSettleAttackSplit(t, g, me.ID)
	cats := battlefieldIDsNamed(g, "Cat")
	if len(cats) != 2 {
		t.Fatalf("%d Cats, want 2", len(cats))
	}
	for _, id := range cats {
		if !hasEffectiveKeyword(t, g, id, "double strike") {
			t.Error("an attacking token has double strike as soon as it exists")
		}
	}
	if hasEffectiveKeyword(t, g, warleader, "double strike") {
		t.Error("the Warleader is not a token")
	}
}

// Magecraft is "cast OR COPY": a copy of an instant is created, not
// cast, and still triggers it (CR 207.2c).
func TestCavCMagecraftTriggersOnCopies(t *testing.T) {
	cases := []struct {
		name  string
		build func(g *game.Game, me *game.Player) func() int
	}{
		{"Storm-Kiln Artist", func(g *game.Game, me *game.Player) func() int {
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Storm-Kiln Artist", TypeLine: "Creature — Dwarf Shaman",
				OracleID: stormKilnArtistOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			return func() int { return countBattlefieldNamed(g, me.ID, "Treasure") }
		}},
		{"Sedgemoor Witch", func(g *game.Game, me *game.Player) func() int {
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Sedgemoor Witch", TypeLine: "Creature — Human Warlock",
				OracleID: cavCSedgemoorWitchOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			return func() int { return countBattlefieldNamed(g, me.ID, "Pest") }
		}},
		{"Archmage Emeritus", func(g *game.Game, me *game.Player) func() int {
			pushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Archmage Emeritus", TypeLine: "Creature — Human Wizard",
				OracleID: cavCArchmageOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			start := me.Hand.Size()
			return func() int { return me.Hand.Size() - start }
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, victim := g.Seats[0], g.Seats[1]
			made := tc.build(g, me)
			// Doublecast is a sorcery (one trigger); the Bolt is cast (a
			// second); its copy is created, not cast (the third).
			castCatalogSpell(t, g, "Doublecast", "Sorcery", doublecastOracle, nil)
			passPriorityAroundTable(t, g)
			if got := made(); got != 1 {
				t.Fatalf("after casting Doublecast: %d, want 1", got)
			}
			castBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 1)
			if got := made(); got != 3 {
				t.Errorf("after a cast Bolt and its copy: %d, want 3 (Doublecast, the Bolt, the copy)", got)
			}
		})
	}
}

// Ghalta the Immovable: a creature you control with toughness greater
// than its power assigns combat damage equal to its toughness. A 2/5
// Wall deals 5, Ghalta himself (0/7) deals 7, and a 4/4 still deals 4.
func TestCavCGhaltaAssignsDamageByToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := b12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", rfcbGhaltaImmovable, 0, 7)
	wall := pushVanillaCreature(g, me.ID, "Stout Wall", 2, 5)
	bear := pushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	before := opp.Life
	attackWith(t, g, opp.ID, ghalta, wall, bear)
	passPriorityAroundTable(t, g)
	if want := before - (7 + 5 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d (7 from Ghalta, 5 from the Wall, 4 from the Bear)", opp.Life, want)
	}
}

// The static is "you control": an opponent's 1/6 still deals 1.
func TestCavCGhaltaDoesNotReachOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", rfcbGhaltaImmovable, 0, 7)
	theirs := pushVanillaCreature(g, opp.ID, "Their Wall", 1, 6)
	advanceToNextSeatsTurn(t, g)
	before := me.Life
	attackWith(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("an opponent's 1/6 still deals 1: %d → %d", before, me.Life)
	}
}

// A blocker assigns by toughness too: a 1/5 Wall kills a 4/4 it could not
// have by power alone.
func TestCavCGhaltaWallBlocksForItsToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", rfcbGhaltaImmovable, 0, 7)
	wall := pushVanillaCreature(g, me.ID, "Stout Wall", 1, 5)
	advanceToNextSeatsTurn(t, g)
	raider := pushVanillaCreature(g, opp.ID, "Raider", 4, 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	lockInBlocks(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the Wall blocks for 5 (its toughness), enough to kill a 4/4; by power alone it deals 1")
	}
}

// Dalkovan Encampment's delayed trigger belongs to the turn, not to the
// land: destroy the land after activating it and the attack still brings
// its two Warriors, once however many creatures attack.
func TestCavCDalkovanWarriorsComeAfterTheLandIsGone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushCatalogPermanent(g, me.ID, "Dalkovan Encampment", "Land", b33DalkovanEncampmentOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := b12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	b06AddMana(me, "W", "C", "C")
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
	if g.Battlefield.Contains(land) {
		t.Fatal("the land is gone")
	}
	declareAttack(t, g, opp.ID, bear, elf)
	passAndSettleAttackSplit(t, g, me.ID)
	if got := len(b33AttackingTokensNamed(g, me.ID, opp.ID, "Warrior")); got != 2 {
		t.Errorf("%d Warriors attacking, want 2 (once for the attack, however many attackers)", got)
	}
}
