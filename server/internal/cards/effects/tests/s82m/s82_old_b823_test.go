package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b823_test.go — Homelands, "Pump and combat tricks", batch 823
// (#5635). Every test plays the real catalog entry through the engine.

const (
	o823Soraya    = "0f1c14c5-7a28-47b0-a603-33cc592b00ad"
	o823HeartWolf = "2f50c45f-058d-41fb-9965-08c2e0edc8ac"
	o823Veldrane  = "39322103-c40e-4d05-8e87-a92d1cbc7b68"
	o823Inquis    = "3c6a32e8-0c2a-45cb-8b79-c7f022fe5129"
	o823Tower     = "3d169bb1-e91b-46d8-9280-797192926314"
	o823Noble     = "6172e5b9-2a01-429c-a140-e9b22fbab56e"
	o823Folk      = "86002526-ed3d-47e5-9f2a-55c4fa220ce8"
	o823Rashka    = "8e1e477a-192a-4e5c-a18b-7e25d9bd47f3"
	o823Anaba     = "9d2e7422-397a-47d4-9a1e-1957c7db9a18"
	o823Spider    = "d8fe0115-5d98-4251-a48c-5e849430a445"
)

func TestS82OldB823EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o823Soraya, o823HeartWolf, o823Veldrane, o823Inquis, o823Tower, o823Noble, o823Folk, o823Rashka, o823Anaba, o823Spider} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o823Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, colors []string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, Colors: colors,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestS82OldB823SorayaPumpsEveryBirdAndGivesBanding(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	sor := o823Perm(g, me.ID, "Soraya the Falconer", "Legendary Creature — Human", o823Soraya, []string{"W"}, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("my Bird is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 3 || tough != 3 {
		t.Errorf("their Bird is %d/%d, want 3/3: the anthem is not limited to the controller's Birds", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the Bear is %d/%d, want 2/2", p, tough)
	}
	effectstest.FloatForTest(g, me, "CW")
	if err := g.ActivateCatalogAbility(me.ID, sor, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("banding was granted to a non-Bird")
	}
	effectstest.ApaActivate(t, g, me, sor, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(mine)})
	if !effectstest.HasEffectiveKeyword(t, g, mine, game.KeywordBanding) {
		t.Error("the Bird did not gain banding")
	}
}

func TestS82OldB823HeartWolfWorksOnlyInCombatAndDiesWithItsDwarf(t *testing.T) {
	for _, tc := range []struct {
		name       string
		dwarfLeave bool
	}{{"the Dwarf leaves", true}, {"the Dwarf stays", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			wolf := effectstest.B12Push(g, me.ID, "Heart Wolf", "Creature — Wolf", o823HeartWolf, 2, 2)
			dwarf := effectstest.B12Creature(g, me.ID, "Dwarf", "Creature — Dwarf", 1, 1)
			bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
			effectstest.AdvanceToMain(t, g)
			if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(dwarf)}); err == nil {
				t.Fatal("activated outside combat")
			}
			effectstest.DeclareAttack(t, g, opp.ID, dwarf)
			if err := g.ActivateCatalogAbility(me.ID, wolf, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
				t.Fatal("targeted a non-Dwarf")
			}
			effectstest.ApaActivate(t, g, me, wolf, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dwarf)})
			if p, _ := effectstest.PtOf(t, g, dwarf); p != 3 {
				t.Errorf("the Dwarf's power is %d, want 3", p)
			}
			if !effectstest.HasEffectiveKeyword(t, g, dwarf, "first strike") {
				t.Error("the Dwarf did not gain first strike")
			}
			if tc.dwarfLeave {
				effectstest.DestroyIt(t, g, dwarf)
				effectstest.PassPriorityAroundTable(t, g)
				if g.Battlefield.Contains(wolf) {
					t.Error("Heart Wolf was not sacrificed when the Dwarf left")
				}
			} else if !g.Battlefield.Contains(wolf) {
				t.Error("Heart Wolf left with its Dwarf still here")
			}
		})
	}
}

func TestS82OldB823VeldraneShrinksAndGainsForestwalk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	vel := o823Perm(g, me.ID, "Veldrane of Sengir", "Legendary Creature — Human Rogue", o823Veldrane, []string{"B"}, 5, 5)
	effectstest.FloatForTest(g, me, "CBB")
	effectstest.ApaActivate(t, g, me, vel, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, vel); p != 2 || tough != 5 {
		t.Errorf("Veldrane is %d/%d, want 2/5", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vel, "forestwalk") {
		t.Error("Veldrane did not gain forestwalk")
	}
}

func TestS82OldB823SerraInquisitorsPumpAgainstBlackOnly(t *testing.T) {
	for _, tc := range []struct {
		name    string
		color   string
		blocked bool
		want    int
	}{
		{"blocked by a black creature", "B", true, 5},
		{"blocked by a red creature", "R", true, 3},
		{"blocking a black creature", "B", false, 5},
		{"blocking a red creature", "R", false, 3},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			if tc.blocked {
				insq := o823Perm(g, me.ID, "Serra Inquisitors", "Creature — Human Cleric", o823Inquis, []string{"W"}, 3, 3)
				foe := o823Perm(g, opp.ID, "Foe", "Creature — Bear", "", []string{tc.color}, 0, 9)
				effectstest.DeclareAttack(t, g, opp.ID, insq)
				effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
				if err := g.DeclareBlocker(foe, insq); err != nil {
					t.Fatalf("block: %v", err)
				}
				effectstest.LockInBlocks(t, g)
				effectstest.PassPriorityAroundTable(t, g)
				if p, _ := effectstest.PtOf(t, g, insq); p != tc.want {
					t.Errorf("power %d, want %d", p, tc.want)
				}
				return
			}
			atk := o823Perm(g, me.ID, "Attacker", "Creature — Bear", "", []string{tc.color}, 0, 9)
			insq := o823Perm(g, opp.ID, "Serra Inquisitors", "Creature — Human Cleric", o823Inquis, []string{"W"}, 3, 3)
			effectstest.DeclareAttack(t, g, opp.ID, atk)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(insq, atk); err != nil {
				t.Fatalf("block: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if p, _ := effectstest.PtOf(t, g, insq); p != tc.want {
				t.Errorf("power %d, want %d", p, tc.want)
			}
		})
	}
}

func TestS82OldB823RashkaGrowsOnlyWhenBlockingABlackCreature(t *testing.T) {
	for _, tc := range []struct {
		color        string
		wantP, wantT int
	}{{"B", 4, 5}, {"R", 3, 3}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		atk := o823Perm(g, me.ID, "Attacker", "Creature — Bear", "", []string{tc.color}, 0, 9)
		rashka := o823Perm(g, opp.ID, "Rashka the Slayer", "Legendary Creature — Human Archer", o823Rashka, []string{"W"}, 3, 3)
		effectstest.DeclareAttack(t, g, opp.ID, atk)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(rashka, atk); err != nil {
			t.Fatalf("block: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, rashka); p != tc.wantP || tough != tc.wantT {
			t.Errorf("against a %s attacker Rashka is %d/%d, want %d/%d", tc.color, p, tough, tc.wantP, tc.wantT)
		}
	}
}

func TestS82OldB823FolkOfAnHavvaAndRootSpiderPumpWhenTheyBlock(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine string
		p, tough, wantP        int
		firstStrike            bool
	}{
		{"Folk of An-Havva", o823Folk, "Creature — Human", 1, 1, 3, false},
		{"Root Spider", o823Spider, "Creature — Spider", 2, 2, 3, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 0, 9)
			blk := o823Perm(g, opp.ID, tc.name, tc.typeLine, tc.oracle, []string{"G"}, tc.p, tc.tough)
			if p, _ := effectstest.PtOf(t, g, blk); p != tc.p {
				t.Fatalf("power %d before blocking, want %d", p, tc.p)
			}
			effectstest.DeclareAttack(t, g, opp.ID, atk)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(blk, atk); err != nil {
				t.Fatalf("block: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if p, _ := effectstest.PtOf(t, g, blk); p != tc.wantP {
				t.Errorf("power %d after blocking, want %d", p, tc.wantP)
			}
			if got := effectstest.HasEffectiveKeyword(t, g, blk, "first strike"); got != tc.firstStrike {
				t.Errorf("first strike = %v, want %v", got, tc.firstStrike)
			}
		})
	}
}

func TestS82OldB823AnabaAncestorPumpsAnotherMinotaurOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	anc := o823Perm(g, me.ID, "Anaba Ancestor", "Creature — Minotaur Spirit", o823Anaba, []string{"R"}, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Minotaur", "Creature — Minotaur", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, anc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a non-Minotaur was targeted")
	}
	if err := g.ActivateCatalogAbility(me.ID, anc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(anc)}); err == nil {
		t.Fatal("the Ancestor targeted itself")
	}
	effectstest.ApaActivate(t, g, me, anc, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)})
	if p, tough := effectstest.PtOf(t, g, other); p != 3 || tough != 3 {
		t.Errorf("the Minotaur is %d/%d, want 3/3", p, tough)
	}
}

func TestS82OldB823AlibansTowerNeedsABlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Ogre", 0, 9)
	blk := effectstest.B12Creature(g, me.ID, "Blocker", "Creature — Bear", 2, 2)
	card := effectstest.HandCardForTest(me, "Aliban's Tower", "Instant", o823Tower)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, me.ID, atk)
	effectstest.FloatForTest(g, me, "RC")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(blk)}); err == nil {
		t.Fatal("cast on a creature that is not blocking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blk, atk); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.FloatForTest(g, me, "RC")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(blk)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, blk); p != 5 || tough != 3 {
		t.Errorf("the blocker is %d/%d, want 5/3", p, tough)
	}
}

func TestS82OldB823FaerieNobleBoostsOtherFaeriesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	noble := o823Perm(g, me.ID, "Faerie Noble", "Creature — Faerie Noble", o823Noble, []string{"G"}, 1, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Faerie", "Creature — Faerie", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Faerie", "Creature — Faerie", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 1 || tough != 2 {
		t.Errorf("my Faerie is %d/%d, want 1/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their Faerie is %d/%d, want 1/1", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, noble); p != 1 || tough != 2 {
		t.Errorf("the Noble is %d/%d, want 1/2: it is not its own other Faerie", p, tough)
	}
	effectstest.ApaActivate(t, g, me, noble, 0, game.ActivateAbilityParams{})
	if p, _ := effectstest.PtOf(t, g, mine); p != 2 {
		t.Errorf("my Faerie's power is %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 1 {
		t.Errorf("their Faerie's power is %d, want 1", p)
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 2 {
		t.Errorf("the Bear's power is %d, want 2", p)
	}
	if p, _ := effectstest.PtOf(t, g, noble); p != 1 {
		t.Errorf("the Noble's power is %d, want 1", p)
	}
}
