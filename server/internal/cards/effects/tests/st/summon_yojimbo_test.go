package testsst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// summon_yojimbo_test.go — Summon: Yojimbo and the duration attack tax
// (#3398, the Tidus deck request #3246).

const yojimboOracle = "ff70d743-280e-476f-a3b8-92026419419d"

func TestSummonYojimboIsRegisteredAndWhole(t *testing.T) {
	spec, ok := effects.Lookup(yojimboOracle)
	if !ok || spec.Name != "Summon: Yojimbo" {
		t.Fatalf("Summon: Yojimbo is not registered: %+v", spec)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

func yojimboAttackPrice(g *game.Game, attacker, defender uuid.UUID) game.AttackTaxPrice {
	return g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: attacker, Target: defender}})
}

func tapBattlefieldCard(g *game.Game, id uuid.UUID) {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
}

func TestSummonYojimboChapterIExilesATappedCreatureOrAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	tapped := effectstest.TdCreature(g, opp.ID, "Tapped", 2, 2)
	tapBattlefieldCard(g, tapped)
	idle := effectstest.TdCreature(g, opp.ID, "Idle", 2, 2)

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Summon: Yojimbo", TypeLine: "Enchantment Creature — Saga Samurai",
		OracleID: yojimboOracle, ManaCost: "{3}{W}", Power: 3, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: idle}); err == nil {
		t.Error("an untapped creature was a legal target")
	}
	effectstest.PickCard(t, g, me.ID, tapped)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(tapped) {
		t.Error("chapter I did not exile the tapped creature")
	}
}

func TestSummonYojimboTaxesAttacksOnItsControllerUntilTheirNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
	raider := effectstest.TdCreature(g, opp.ID, "Raider", 2, 2)

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Summon: Yojimbo", TypeLine: "Enchantment Creature — Saga Samurai",
		OracleID: yojimboOracle, ManaCost: "{3}{W}", Power: 3, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !yojimboAttackPrice(g, raider, me.ID).IsFree() {
		t.Fatal("chapter I taxed attacks")
	}

	effectstest.AdvanceToPrecombatMainOf(t, g, seat) // chapter II
	effectstest.PassPriorityAroundTable(t, g)
	price := yojimboAttackPrice(g, raider, me.ID)
	if price.Cost != "{2}" || len(price.Lines) != 1 {
		t.Fatalf("after chapter II an attack on its controller costs %q (%d lines), want {2}", price.Cost, len(price.Lines))
	}
	// Only its controller is defended.
	other := g.Seats[(seat+2)%len(g.Seats)]
	if !yojimboAttackPrice(g, raider, other.ID).IsFree() {
		t.Error("an attack on another player was taxed")
	}
	if !yojimboAttackPrice(g, raider, opp.ID).IsFree() {
		t.Error("an attack on the raider's own controller was taxed")
	}

	// It lasts through the opponents' turns, and ends as the controller's
	// next turn begins.
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if got := yojimboAttackPrice(g, raider, me.ID).Cost; got != "{2}" {
		t.Errorf("on an opponent's turn the tax is %q, want {2}", got)
	}
	for !(g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepUpkeep) {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if !yojimboAttackPrice(g, raider, me.ID).IsFree() {
		t.Error("the tax outlived 'until your next turn'")
	}
}

func TestSummonYojimboChapterIVMakesATreasureForEachOpponentWithABigCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	a, b, c := g.Seats[(seat+1)%4], g.Seats[(seat+2)%4], g.Seats[(seat+3)%4]
	effectstest.TdCreature(g, a.ID, "A1", 4, 4)
	effectstest.TdCreature(g, a.ID, "A2", 5, 5)
	effectstest.TdCreature(g, b.ID, "B1", 4, 1)
	effectstest.TdCreature(g, c.ID, "C1", 3, 9)
	effectstest.TdCreature(g, me.ID, "Mine", 6, 6)

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Summon: Yojimbo", TypeLine: "Enchantment Creature — Saga Samurai",
		OracleID: yojimboOracle, ManaCost: "{3}{W}", Power: 3, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 3; i++ { // II, III, IV
		effectstest.AdvanceToPrecombatMainOf(t, g, seat)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountBattlefieldByName(g, "Treasure"); got != 2 {
		t.Errorf("chapter IV made %d Treasures, want 2 (two opponents have a power 4+ creature)", got)
	}
}
