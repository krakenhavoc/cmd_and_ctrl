package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b065_cards_test.go — Kaldheim Auras (S82, set OLD batch 065).

const (
	oldB065GlitteringFrost = "2013df0b-9e8c-4dbf-87b2-0dec983266fa"
	oldB065Arachnoform     = "2e41ae94-db31-4060-881b-a2f6d834c1ee"
	oldB065ValorOfWorthy   = "84534792-7834-4a3c-9940-c1eccbe6e9cf"
	oldB065BindMonster     = "8d1a3935-8a8c-4754-b474-189611188bc5"
	oldB065Withercrown     = "bc700a38-7467-41fd-861f-940e0e5982bd"
	oldB065GiantsGrasp     = "bca3f496-ce2a-4715-ae47-5641c33d629e"
	oldB065MistsOfLittjara = "c0162a6c-12d9-4b84-811a-ee6e39401863"
	oldB065BoundInGold     = "ca597a0d-b510-4b25-9a83-4d4e613546f5"
)

func TestOldB065EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB065GlitteringFrost: "Glittering Frost",
		oldB065Arachnoform:     "Arachnoform",
		oldB065ValorOfWorthy:   "Valor of the Worthy",
		oldB065BindMonster:     "Bind the Monster",
		oldB065Withercrown:     "Withercrown",
		oldB065GiantsGrasp:     "Giant's Grasp",
		oldB065MistsOfLittjara: "Mists of Littjara",
		oldB065BoundInGold:     "Bound in Gold",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func oldB065Pick(t *testing.T, g *game.Game, chooser uuid.UUID, id uuid.UUID) {
	t.Helper()
	for i := 0; i < 4; i++ {
		if p := effectstest.LatestPickTarget(g, chooser); p != nil {
			if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: id}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, chooser) == nil && effectstest.StackFullyEmpty(g) {
			return
		}
	}
}

func TestOldB065GlitteringFrostMakesTheLandSnowAndAddsAnyColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	if slices.Contains(effectstest.LayeredCard(t, g, forest).Effective().Supertypes, "Snow") {
		t.Fatal("the bare Forest is already snow")
	}
	effectstest.PushAuraOnLand(t, g, me.ID, "Glittering Frost", oldB065GlitteringFrost, forest)
	if !slices.Contains(effectstest.LayeredCard(t, g, forest).Effective().Supertypes, "Snow") {
		t.Error("the enchanted land is not snow")
	}
	effectstest.TapForMana(t, g, me.ID, forest)
	pick := effectstest.PendingOfKind(g, game.PendingChoiceMana)
	if pick == nil || len(pick.ColorOptions) != 5 {
		t.Fatalf("no five-colour pick for the additional mana: %+v", pick)
	}
	if err := g.ResolveManaChoice(pick.ID, me.ID, "U"); err != nil {
		t.Fatalf("ResolveManaChoice: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 2 || got[0] != "G" || got[1] != "U" {
		t.Errorf("pool = %v, want {G} and the chosen {U}", got)
	}
}

func TestOldB065GlitteringFrostRefusesANonland(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	err := effectstest.CastCatalogSpellErrStrict(t, g, "Glittering Frost", effectstest.AuraTypeLine, oldB065GlitteringFrost, effectstest.CardTarget(bear))
	if err == nil {
		t.Fatal("Glittering Frost was cast on a creature")
	}
}

func TestOldB065ArachnoformPumpsGivesReachAndEveryType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.Enchant(t, g, "Arachnoform", oldB065Arachnoform, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 4/4", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "reach") {
		t.Error("no reach")
	}
	if !effectstest.LayeredCard(t, g, bear).Effective().AllCreatureTypes {
		t.Error("the creature is not every creature type")
	}
}

func TestOldB065ValorOfTheWorthyMakesASpiritWhenTheCreatureLeaves(t *testing.T) {
	for _, how := range []string{"dies", "exiled"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
		effectstest.Enchant(t, g, "Valor of the Worthy", oldB065ValorOfWorthy, bear)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
			t.Fatalf("P/T = %d/%d, want 3/3", p, tough)
		}
		if how == "dies" {
			effectstest.P1Destroy(t, g, bear)
		} else {
			effectstest.ExileCreature(t, g, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		spirits := effectstest.BattlefieldIDsNamed(g, "Spirit")
		if len(spirits) != 1 {
			t.Fatalf("%s: %d Spirit tokens, want 1", how, len(spirits))
		}
		if !effectstest.HasKeywordOf(g, spirits[0], "flying") {
			t.Errorf("%s: the Spirit has no flying", how)
		}
	}
}

func TestOldB065ValorOfTheWorthyMakesNothingWhenTheAuraLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.PushVanillaCreature(g, g.Seats[0].ID, "Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Valor of the Worthy", oldB065ValorOfWorthy, bear)
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Spirit")); n != 0 {
		t.Errorf("%d Spirits, want none", n)
	}
}

func TestOldB065BindTheMonsterTapsLocksAndHurtsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	beast := effectstest.PushVanillaCreature(g, opp.ID, "Beast", 3, 3)
	life := me.Life
	effectstest.Enchant(t, g, "Bind the Monster", oldB065BindMonster, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.BattlefieldCard(g, beast); !c.Tapped {
		t.Fatal("the enchanted creature is not tapped")
	}
	if me.Life != life-3 {
		t.Errorf("life %d, want %d (3 damage from the creature)", me.Life, life-3)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if c, _ := effectstest.BattlefieldCard(g, beast); !c.Tapped {
		t.Error("the enchanted creature untapped in its controller's untap step")
	}
}

func TestOldB065WithercrownZeroesPowerAndTaxesTheUpkeep(t *testing.T) {
	for _, sacrifice := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 4)
		effectstest.Enchant(t, g, "Withercrown", oldB065Withercrown, bear)
		if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 0 || tough != 4 {
			t.Fatalf("P/T = %d/%d, want 0/4", p, tough)
		}
		life := me.Life
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		label := "Lose 1 life"
		if sacrifice {
			label = "Sacrifice this creature"
		}
		effectstest.BlbSTPickOption(t, g, me.ID, label)
		effectstest.PassPriorityAroundTable(t, g)
		if sacrifice {
			if effectstest.BattlefieldHas(g, bear) || me.Life != life {
				t.Errorf("sacrifice: creature on battlefield %v, life %d (was %d)", effectstest.BattlefieldHas(g, bear), me.Life, life)
			}
		} else if !effectstest.BattlefieldHas(g, bear) || me.Life != life-1 {
			t.Errorf("lose life: creature on battlefield %v, life %d (was %d)", effectstest.BattlefieldHas(g, bear), me.Life, life)
		}
	}
}

func TestOldB065GiantsGraspStealsForAsLongAsTheAuraRemains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B31Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "", 3, 3)
	prize := effectstest.PushPermanentForTest(g, opp.ID, "Relic", "", "Artifact")
	aura := effectstest.CastCatalogSpell(t, g, "Giant's Grasp", effectstest.AuraTypeLine, oldB065GiantsGrasp, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	oldB065Pick(t, g, me.ID, prize)
	if got := effectstest.ControllerOf(t, g, prize); got != me.ID {
		t.Fatalf("controller of the stolen permanent = %v, want me", got)
	}
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, prize); got != opp.ID {
		t.Errorf("control did not return when the Aura left")
	}
}

func TestOldB065GiantsGraspTakesNoLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	giant := effectstest.B31Push(g, me.ID, "Hill Giant", "Creature — Giant", "", "", 3, 3)
	forest := effectstest.B31Push(g, opp.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	prize := effectstest.PushPermanentForTest(g, opp.ID, "Relic", "", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Giant's Grasp", effectstest.AuraTypeLine, oldB065GiantsGrasp, effectstest.CardTarget(giant))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: forest}); err == nil {
			t.Error("a land was accepted as the target")
		}
	}
	oldB065Pick(t, g, me.ID, prize)
	if got := effectstest.ControllerOf(t, g, forest); got != opp.ID {
		t.Error("the land changed controllers")
	}
}

func TestOldB065GiantsGraspEnchantsOnlyYourGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	theirs := effectstest.B31Push(g, opp.ID, "Their Giant", "Creature — Giant", "", "", 4, 4)
	for name, host := range map[string]uuid.UUID{"a non-Giant": bear, "an opposing Giant": theirs} {
		if err := effectstest.CastCatalogSpellErrStrict(t, g, "Giant's Grasp", effectstest.AuraTypeLine, oldB065GiantsGrasp, effectstest.CardTarget(host)); err == nil {
			t.Errorf("Giant's Grasp was cast on %s", name)
		}
	}
	_ = me
}

func TestOldB065MistsOfLittjaraShrinksPowerAndEnchantsAVehicle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	spec, _ := effects.Lookup(oldB065MistsOfLittjara)
	if !slices.Contains(spec.PrintedKeywords, "flash") {
		t.Error("Mists of Littjara lacks flash")
	}
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 4, 4)
	effectstest.Enchant(t, g, "Mists of Littjara", oldB065MistsOfLittjara, bear)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 1 || tough != 4 {
		t.Errorf("P/T = %d/%d, want 1/4", p, tough)
	}
	vehicle := effectstest.PushVehicleForTest(g, me.ID, "Skiff", "", 3, 3)
	aura := effectstest.Enchant(t, g, "Mists of Littjara", oldB065MistsOfLittjara, vehicle)
	if got := effectstest.AttachmentHostOf(t, g, aura); got.ID != vehicle {
		t.Errorf("the Aura is on %v, want the Vehicle", got)
	}
}

func TestOldB065MistsOfLittjaraRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	forest := effectstest.B31Push(g, g.Seats[0].ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Mists of Littjara", effectstest.AuraTypeLine, oldB065MistsOfLittjara, effectstest.CardTarget(forest)); err == nil {
		t.Fatal("Mists of Littjara was cast on a land")
	}
}

func TestOldB065BoundInGoldLocksAPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.Enchant(t, g, "Bound in Gold", oldB065BoundInGold, bear)
	eff := effectstest.LayeredCard(t, g, bear).Effective()
	if eff.Restrictions&game.CantAttackOrBlock != game.CantAttackOrBlock || eff.Restrictions&game.CantActivate == 0 {
		t.Errorf("restrictions %v lack attack/block/activate", eff.Restrictions)
	}
	if eff.ExtraRestrictions&game.CantCrew == 0 {
		t.Error("the creature can still crew")
	}
	if eff.Restrictions&game.CantActivateMana != 0 {
		t.Error("mana abilities are locked too")
	}
	forest := effectstest.B31Push(g, me.ID, "Forest", "Basic Land — Forest", "", "", 0, 0)
	effectstest.Enchant(t, g, "Bound in Gold", oldB065BoundInGold, forest)
}
