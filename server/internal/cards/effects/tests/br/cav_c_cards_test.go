package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The controller of a "tapped and attacking" token chooses what it
// attacks (CR 508.4): Leonin Warleader's two Cats can go at different
// opponents, and the prompt opens on the player the Warleader attacked.
func TestCavCWarleaderCatsAttackWhomeverYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, first, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	warleader := effectstest.B31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", effectstest.B31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	// Attack the SECOND opponent in seat order: the prompt must open on
	// them (what the Warleader attacked), not on the first seat.
	effectstest.DeclareAttack(t, g, second.ID, warleader)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.PayAmountPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("with three opponents to attack the controller is asked where the Cats go")
	}
	if prompt.PayAmount.Max != 2 || prompt.PayAmount.Goal != 2 {
		t.Errorf("first prompt 0..%d goal %d, want 0..2 goal 2 (all at what the Warleader attacked)", prompt.PayAmount.Max, prompt.PayAmount.Goal)
	}
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 1); err != nil {
		t.Fatalf("one at the attacked opponent: %v", err)
	}
	prompt = effectstest.PayAmountPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("asked about the next opponent")
	}
	if err := g.ResolvePayAmount(prompt.ID, me.ID, 0); err != nil {
		t.Fatalf("none at the next: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	at := map[uuid.UUID]int{}
	for _, id := range effectstest.BattlefieldIDsNamed(g, "Cat") {
		at[effectstest.CardByID(g, id).AttackingTarget]++
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
		{"Berserkers' Onslaught", effectstest.CavCBerserkersOnslaughtOracle, "Enchantment"},
		{"Blade Historian", effectstest.B25BladeHistorianOracle, "Creature — Human Cleric"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.PushCatalogPermanent(g, me.ID, tc.name, tc.typeLine, tc.oracle, false)
			warleader := effectstest.B31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", effectstest.B31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
			home := effectstest.PushVanillaCreature(g, me.ID, "Stay Home", 2, 2)
			effectstest.DeclareAttack(t, g, opp.ID, warleader)
			effectstest.PassAndSettleAttackSplit(t, g, me.ID)
			cats := effectstest.BattlefieldIDsNamed(g, "Cat")
			if len(cats) != 2 {
				t.Fatalf("%d Cats, want 2", len(cats))
			}
			for _, id := range cats {
				if !effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
					t.Error("a Cat that entered attacking has double strike")
				}
			}
			if !effectstest.HasEffectiveKeyword(t, g, warleader, "double strike") {
				t.Error("the declared attacker has double strike")
			}
			if effectstest.HasEffectiveKeyword(t, g, home, "double strike") {
				t.Error("a creature that is not attacking does not")
			}
			effectstest.AdvanceTo(t, g, game.StepCombatDamage)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
			for _, id := range cats {
				if effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
					t.Error("combat is over: nothing is attacking, nothing has double strike")
				}
			}
		})
	}
}

// Neyali's static covers an attacking TOKEN the moment it appears, and
// only a token: the Warleader that made the Cats is not one.
func TestCavCNeyaliGivesAttackingTokensDoubleStrikeAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Neyali, Suns' Vanguard", "Legendary Creature — Human Rebel", effectstest.B32NeyaliOracle, 3, 3)
	warleader := effectstest.B31Push(g, me.ID, "Leonin Warleader", "Creature — Cat Soldier", effectstest.B31LeoninWarleaderOracle, "{2}{W}{W}", 4, 4, "W")
	effectstest.DeclareAttack(t, g, opp.ID, warleader)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	cats := effectstest.BattlefieldIDsNamed(g, "Cat")
	if len(cats) != 2 {
		t.Fatalf("%d Cats, want 2", len(cats))
	}
	for _, id := range cats {
		if !effectstest.HasEffectiveKeyword(t, g, id, "double strike") {
			t.Error("an attacking token has double strike as soon as it exists")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, warleader, "double strike") {
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
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Storm-Kiln Artist", TypeLine: "Creature — Dwarf Shaman",
				OracleID: effectstest.StormKilnArtistOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			return func() int { return effectstest.CountBattlefieldNamed(g, me.ID, "Treasure") }
		}},
		{"Sedgemoor Witch", func(g *game.Game, me *game.Player) func() int {
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Sedgemoor Witch", TypeLine: "Creature — Human Warlock",
				OracleID: effectstest.CavCSedgemoorWitchOracle, Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			return func() int { return effectstest.CountBattlefieldNamed(g, me.ID, "Pest") }
		}},
		{"Archmage Emeritus", func(g *game.Game, me *game.Player) func() int {
			effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
				InstanceID: uuid.New(), Name: "Archmage Emeritus", TypeLine: "Creature — Human Wizard",
				OracleID: effectstest.CavCArchmageOracle, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
			})
			start := me.Hand.Size()
			return func() int { return me.Hand.Size() - start }
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, victim := g.Seats[0], g.Seats[1]
			made := tc.build(g, me)
			// Doublecast is a sorcery (one trigger); the Bolt is cast (a
			// second); its copy is created, not cast (the third).
			effectstest.CastCatalogSpell(t, g, "Doublecast", "Sorcery", effectstest.DoublecastOracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			if got := made(); got != 1 {
				t.Fatalf("after casting Doublecast: %d, want 1", got)
			}
			effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me.ID, victim.ID, 1)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ghalta := effectstest.B12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", effectstest.RfcbGhaltaImmovable, 0, 7)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Stout Wall", 2, 5)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Big Bear", 4, 4)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, ghalta, wall, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if want := before - (7 + 5 + 4); opp.Life != want {
		t.Errorf("opponent at %d, want %d (7 from Ghalta, 5 from the Wall, 4 from the Bear)", opp.Life, want)
	}
}

// The static is "you control": an opponent's 1/6 still deals 1.
func TestCavCGhaltaDoesNotReachOpposingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", effectstest.RfcbGhaltaImmovable, 0, 7)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Wall", 1, 6)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	before := me.Life
	effectstest.AttackWith(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-1 {
		t.Errorf("an opponent's 1/6 still deals 1: %d → %d", before, me.Life)
	}
}

// A blocker assigns by toughness too: a 1/5 Wall kills a 4/4 it could not
// have by power alone.
func TestCavCGhaltaWallBlocksForItsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Ghalta the Immovable", "Legendary Creature — Elder Dinosaur", effectstest.RfcbGhaltaImmovable, 0, 7)
	wall := effectstest.PushVanillaCreature(g, me.ID, "Stout Wall", 1, 5)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	raider := effectstest.PushVanillaCreature(g, opp.ID, "Raider", 4, 4)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(raider, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, raider); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raider) {
		t.Error("the Wall blocks for 5 (its toughness), enough to kill a 4/4; by power alone it deals 1")
	}
}

// Dalkovan Encampment's delayed trigger belongs to the turn, not to the
// land: destroy the land after activating it and the attack still brings
// its two Warriors, once however many creatures attack.
func TestCavCDalkovanWarriorsComeAfterTheLandIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogPermanent(g, me.ID, "Dalkovan Encampment", "Land", effectstest.B33DalkovanEncampmentOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elf := effectstest.B12Creature(g, me.ID, "Elf", "Creature — Elf", 1, 1)
	effectstest.B06AddMana(me, "W", "C", "C")
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(land) })
	if g.Battlefield.Contains(land) {
		t.Fatal("the land is gone")
	}
	effectstest.DeclareAttack(t, g, opp.ID, bear, elf)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if got := len(effectstest.B33AttackingTokensNamed(g, me.ID, opp.ID, "Warrior")); got != 2 {
		t.Errorf("%d Warriors attacking, want 2 (once for the attack, however many attackers)", got)
	}
}
