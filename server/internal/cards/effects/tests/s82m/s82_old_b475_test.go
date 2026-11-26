package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB475Idol      = "89b1eab9-4d24-4367-99f7-263349f69c3a"
	oldB475Wurm      = "93514874-41e8-49e7-8935-d06e852dd1e3"
	oldB475Knight    = "9456c5b6-946d-403a-8ed0-dff9f921d98c"
	oldB475Force     = "b0967525-554b-41eb-b639-75fe36e0fe57"
	oldB475Cyclops   = "ca5711d1-8cc9-4540-a567-778b56d51f50"
	oldB475Servant   = "dbf64afa-ace6-44b9-b47f-750c7e52cc29"
	oldB475Hellkite  = "de672a4b-07eb-4be8-a03f-23dfbb7ac0fa"
	oldB475BoltOracl = "4457ed35-7c10-48c8-9776-456485fdf070"
)

func TestOldB475CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB475Idol: "Jinxed Idol", oldB475Wurm: "Yavimaya Wurm", oldB475Knight: "Black Knight",
		oldB475Force: "Destructive Force", oldB475Cyclops: "Cyclops Gladiator", oldB475Servant: "Fire Servant",
		oldB475Hellkite: "Ancient Hellkite",
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

func TestOldB475WurmAndKnightKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wurm := effectstest.B12Push(g, me.ID, "Yavimaya Wurm", "Creature — Wurm", oldB475Wurm, 6, 4)
	knight := effectstest.B12Push(g, me.ID, "Black Knight", "Creature — Human Knight", oldB475Knight, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, wurm, "trample") {
		t.Error("Yavimaya Wurm has no trample")
	}
	for _, kw := range []string{"first strike", "protection from white"} {
		if !effectstest.HasEffectiveKeyword(t, g, knight, kw) {
			t.Errorf("Black Knight has no %s", kw)
		}
	}
	_ = opp
}

func TestOldB475IdolHurtsItsControllerAndChangesHands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	idol := effectstest.B12Permanent(g, me.ID, "Jinxed Idol", "Artifact")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, idol).OracleID = oldB475Idol })
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != 0 || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	before := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage from the Idol", before, me.Life)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, idol, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Error("the Idol's controller was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, idol, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, SacrificeIDs: []uuid.UUID{fodder},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the creature was not sacrificed")
	}
	if got := effectstest.ApaLive(g, idol).Controller; got != opp.ID {
		t.Errorf("the Idol is controlled by %s, want the opponent", got)
	}
	// It stays with them past the turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if got := effectstest.ApaLive(g, idol).Controller; got != opp.ID {
		t.Error("control of the Idol went back")
	}
}

func TestOldB475DestructiveForceTakesFiveLandsThenDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var mine []uuid.UUID
	for i := 0; i < 6; i++ {
		mine = append(mine, effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}))
	}
	effectstest.PushLands(g, opp.ID, 5)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 6, 6)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 2, 2)
	effectstest.CastWithCost(t, g, "Destructive Force", "Sorcery", "{5}{R}{R}", oldB475Force)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine[:5]...)
	effectstest.PassPriorityAroundTable(t, g)
	lands := func(id uuid.UUID) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == id && c.IsLand() {
				n++
			}
		}
		return n
	}
	if lands(me.ID) != 1 || lands(opp.ID) != 0 {
		t.Errorf("lands left: me %d (want 1), opponent %d (want 0)", lands(me.ID), lands(opp.ID))
	}
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the 2/2 survived 5 damage")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 5 {
		t.Errorf("the 6/6 has %d damage, want 5", got)
	}
}

func TestOldB475DestructiveForceStillDamagesWhenNobodyHasLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastWithCost(t, g, "Destructive Force", "Sorcery", "{5}{R}{R}", oldB475Force)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear survived with nobody having a land to sacrifice")
	}
}

func TestOldB475CyclopsGladiatorFightsOnAttackWhenYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		cyclops := effectstest.B12Push(g, me.ID, "Cyclops Gladiator", "Creature — Cyclops Warrior", oldB475Cyclops, 4, 4)
		victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(cyclops, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerPickTarget(t, g, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		_, alive := effectstest.BattlefieldCard(g, victim)
		hurt := effectstest.DamageMarkedOn(g, cyclops)
		if yes && (alive || hurt != 2) {
			t.Errorf("yes: victim alive = %v, Gladiator damage = %d, want a dead victim and 2 damage", alive, hurt)
		}
		if !yes && (!alive || hurt != 0) {
			t.Errorf("no: victim alive = %v, Gladiator damage = %d, want nothing to happen", alive, hurt)
		}
	}
}

func TestOldB475FireServantDoublesOnlyYourRedInstantsAndSorceries(t *testing.T) {
	for _, tc := range []struct {
		name  string
		mine  bool // the Servant is mine
		cost  string
		spell string
		want  int
	}{
		{"red instant with my Servant", true, "{R}", "Instant", 6},
		{"nonred instant with my Servant", true, "{1}", "Instant", 3},
		{"red instant with their Servant", false, "{R}", "Instant", 3},
		{"red sorcery with my Servant", true, "{R}", "Sorcery", 6},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
			owner := me.ID
			if !tc.mine {
				owner = opp.ID
			}
			effectstest.B27Push(g, owner, "Fire Servant", "Creature — Elemental", oldB475Servant, "{3}{R}{R}", 4, 3, "R")
			before := opp.Life
			effectstest.CastSpellWithCost(t, g, "Lightning Bolt", tc.spell, oldB475BoltOracl, tc.cost,
				[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			effectstest.PassPriorityAroundTable(t, g)
			if got := before - opp.Life; got != tc.want {
				t.Errorf("the spell dealt %d, want %d", got, tc.want)
			}
		})
	}
}

func TestOldB475FireServantLeavesCreatureDamageAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B27Push(g, me.ID, "Fire Servant", "Creature — Elemental", oldB475Servant, "{3}{R}{R}", 4, 3, "R")
	// A red creature of mine hits for its power, undoubled.
	hit := effectstest.B16Creature(g, me.ID, "Red Bear", "Creature — Bear", 2, 2, "R")
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(hit, opp.ID, 2) })
	if got := before - opp.Life; got != 2 {
		t.Errorf("a red creature dealt %d, want 2", got)
	}
}

func TestOldB475HellkitePingsOnlyTheDefendersCreaturesWhileAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	kite := effectstest.B12Push(g, me.ID, "Ancient Hellkite", "Creature — Dragon", oldB475Hellkite, 6, 6)
	aimed := effectstest.B12Creature(g, defender.ID, "Aimed Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, third.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(aimed)}); err == nil {
		t.Fatal("the Hellkite pinged while it was not attacking")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(kite, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, kite, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bystander)}); err == nil {
		t.Fatal("a creature the attacked player does not control was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, kite, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(aimed)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, aimed); got != 1 {
		t.Errorf("the target has %d damage, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, bystander); got != 0 {
		t.Errorf("the bystander has %d damage, want 0", got)
	}
}
