package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b681_test.go — Mercadian Masques, "Activated abilities", batch 681
// (#5619). Every test plays the real catalog entry through the engine.

const (
	o681Panacea   = "08ee7879-3c9d-43c6-a44e-f81180f938fd"
	o681Lance     = "09e588c2-0fb8-4c30-aff3-433db7a07b6e"
	o681Instig    = "0fe1a1f7-41c9-415e-a5e3-b5ba0c0271e5"
	o681Balloon   = "10f182a7-ffa3-4128-8b2e-e05b97f92e00"
	o681Detail    = "2a26dee4-b4b6-489d-99e5-6dbd8d19419d"
	o681Herbalist = "2e35bcef-505a-4559-a1fe-f42f4543197f"
	o681Darting   = "2f33e684-a4e4-4528-9a20-3df902bc2cde"
	o681Puffer    = "4f427de6-b551-4699-b78a-44ff668198d8"
	o681Overtaker = "71668016-cefb-43d2-a84c-47956cffd71e"
	o681Henge     = "71744428-65ae-42ac-893d-0802fd41e9b4"
	o681Runner    = "730fbd6a-096c-4d25-9372-c05f90088f83"
	o681Pawnshop  = "9eef1d16-4405-40e0-8b8b-008e08aac4ac"
)

func TestS82OldB681EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o681Panacea, o681Lance, o681Instig, o681Balloon, o681Detail, o681Herbalist, o681Darting, o681Puffer, o681Overtaker, o681Henge, o681Runner, o681Pawnshop} {
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

func o681Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func o681Hand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Spare", TypeLine: "Sorcery", Owner: p.ID, Controller: p.ID})
	})
	return id
}

func o681Activate(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID, params game.ActivateAbilityParams) error {
	t.Helper()
	params.Strict = true
	return g.ActivateCatalogAbility(me.ID, src, 0, params)
}

func TestS82OldB681PanaceaPreventsTheNextXDamageForTwiceXMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pan := o681Perm(g, me.ID, "Panacea", "Artifact", o681Panacea, 0, 0)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 4, 4)
	effectstest.FloatForTest(g, me, "CCC")
	if err := o681Activate(t, g, me, pan, game.ActivateAbilityParams{XValue: 2, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("{X}{X} with X=2 was paid with three mana")
	}
	effectstest.FloatForTest(g, me, "C")
	if err := o681Activate(t, g, me, pan, game.ActivateAbilityParams{XValue: 2, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Cmdr2B08Damage(g, src, victim, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, victim); n != 1 {
		t.Errorf("%d damage marked, want 1 (3 less the 2 prevented)", n)
	}
}

func TestS82OldB681IronLanceGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	lance := o681Perm(g, me.ID, "Iron Lance", "Artifact", o681Lance, 0, 0)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Ogre", 2, 2)
	effectstest.FloatForTest(g, me, "CCC")
	effectstest.ApaActivate(t, g, me, lance, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(foe)})
	if !effectstest.HasEffectiveKeyword(t, g, foe, "first strike") {
		t.Error("the target did not gain first strike")
	}
	if c, _ := effectstest.BattlefieldCard(g, lance); !c.Tapped {
		t.Error("Iron Lance was not tapped as a cost")
	}
}

func TestS82OldB681InstigatorForcesATargetPlayersCreaturesToAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	inst := o681Perm(g, me.ID, "Instigator", "Creature — Human Spellshaper", o681Instig, 1, 1)
	effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Ogre", 2, 2)
	effectstest.FloatForTest(g, me, "CBB")
	if err := o681Activate(t, g, me, inst, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err == nil {
		t.Fatal("Instigator activated without discarding a card")
	}
	spare := o681Hand(g, me)
	effectstest.ApaActivate(t, g, me, inst, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, DiscardIDs: []uuid.UUID{spare}})
	if !me.Graveyard.Contains(spare) {
		t.Error("the discarded card is not in the graveyard")
	}
	if c, _ := effectstest.BattlefieldCard(g, inst); !c.Tapped {
		t.Error("Instigator was not tapped")
	}
}

func TestS82OldB681BalloonPeddlerGrantsFlyingForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	peddler := o681Perm(g, me.ID, "Balloon Peddler", "Creature — Human Spellshaper", o681Balloon, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "U")
	if err := o681Activate(t, g, me, peddler, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("activated without a discard")
	}
	spare := o681Hand(g, me)
	effectstest.ApaActivate(t, g, me, peddler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{spare}})
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the target did not gain flying")
	}
}

func TestS82OldB681SecurityDetailNeedsNoCreaturesAndOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	det := o681Perm(g, me.ID, "Security Detail", "Enchantment", o681Detail, 0, 0)
	guard := effectstest.B12Creature(g, me.ID, "Guard", "Creature — Soldier", 1, 1)
	effectstest.FloatForTest(g, me, "WWWW")
	if err := o681Activate(t, g, me, det, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("activated while controlling a creature")
	}
	g.WithWriteLock(func() { g.Battlefield.Remove(guard) })
	effectstest.ApaActivateNoPoolCheck(t, g, me, det, 0)
	soldiers := 0
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Controller == me.ID {
			soldiers++
		}
	}
	if soldiers != 1 {
		t.Fatalf("%d Soldier tokens, want 1", soldiers)
	}
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.IsToken() {
				g.Battlefield.Remove(c.InstanceID)
				break
			}
		}
	})
	if err := o681Activate(t, g, me, det, game.ActivateAbilityParams{}); err == nil {
		t.Error("activated a second time in the same turn")
	}
}

func TestS82OldB681RushwoodHerbalistRegeneratesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	herb := o681Perm(g, me.ID, "Rushwood Herbalist", "Creature — Human Spellshaper", o681Herbalist, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 5, 5)
	effectstest.FloatForTest(g, me, "G")
	spare := o681Hand(g, me)
	effectstest.ApaActivate(t, g, me, herb, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), DiscardIDs: []uuid.UUID{spare}})
	effectstest.Cmdr2B08Damage(g, src, bear, 5)
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	c, ok := effectstest.BattlefieldCard(g, bear)
	if !ok {
		t.Fatal("the shielded creature died")
	}
	if !c.Tapped {
		t.Error("a regenerated creature should be tapped")
	}
}

func TestS82OldB681DartingMerfolkReturnsItselfToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	merf := o681Perm(g, me.ID, "Darting Merfolk", "Creature — Merfolk", o681Darting, 1, 1)
	effectstest.FloatForTest(g, me, "U")
	effectstest.ApaActivate(t, g, me, merf, 0, game.ActivateAbilityParams{})
	if effectstest.FindBattlefieldCardForTest(g, merf) != nil {
		t.Error("Darting Merfolk is still on the battlefield")
	}
	if !me.Hand.Contains(merf) {
		t.Error("Darting Merfolk is not in its owner's hand")
	}
}

func TestS82OldB681PufferExtractPumpsThenDestroysAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	ext := o681Perm(g, me.ID, "Puffer Extract", "Artifact", o681Puffer, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CCC")
	effectstest.ApaActivate(t, g, me, ext, 0, game.ActivateAbilityParams{XValue: 3, Targets: effectstest.CardRefs(bear)})
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the creature is %d/%d, want 5/5", p, tough)
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Error("the pumped creature was not destroyed at the end step")
	}
}

func TestS82OldB681OvertakerUntapsAndStealsUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ov := o681Perm(g, me.ID, "Overtaker", "Creature — Merfolk Spellshaper", o681Overtaker, 1, 1)
	foe := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Tapped Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3,
		Owner: opp.ID, Controller: opp.ID, Tapped: true})
	effectstest.FloatForTest(g, me, "CCCU")
	spare := o681Hand(g, me)
	effectstest.ApaActivate(t, g, me, ov, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(foe), DiscardIDs: []uuid.UUID{spare}})
	c, _ := effectstest.BattlefieldCard(g, foe)
	if c.Tapped {
		t.Error("the stolen creature is still tapped")
	}
	if effectstest.ControllerOf(t, g, foe) != me.ID {
		t.Error("Overtaker's controller did not gain control")
	}
	if !effectstest.HasEffectiveKeyword(t, g, foe, "haste") {
		t.Error("the stolen creature did not gain haste")
	}
}

func TestS82OldB681HengeGuardianGainsTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	hg := o681Perm(g, me.ID, "Henge Guardian", "Artifact Creature — Dragon Wurm", o681Henge, 3, 4)
	effectstest.FloatForTest(g, me, "CC")
	effectstest.ApaActivate(t, g, me, hg, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, hg, "trample") {
		t.Error("Henge Guardian did not gain trample")
	}
}

func TestS82OldB681TrapRunnerMakesAnUnblockedAttackerBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	runner := o681Perm(g, opp.ID, "Trap Runner", "Creature — Human Soldier", o681Runner, 2, 3)
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(opp.ID, runner, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)}); err == nil {
		t.Fatal("Trap Runner activated outside combat")
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(atk, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	effectstest.FinishBlockDeclarations(t, g)
	unblocked := false
	g.ReadSnapshot(func() { unblocked = g.UnblockedAttackerForEffect(atk) })
	if !unblocked {
		t.Fatal("the attacker should be unblocked")
	}
	if err := g.ActivateCatalogAbility(opp.ID, runner, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(atk)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.ReadSnapshot(func() { unblocked = g.UnblockedAttackerForEffect(atk) })
	if unblocked {
		t.Error("the attacker is still unblocked")
	}
}

func TestS82OldB681RishadanPawnshopShufflesANontokenPermanentIntoItsOwnersLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	shop := o681Perm(g, me.ID, "Rishadan Pawnshop", "Artifact", o681Pawnshop, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CC")
	if err := o681Activate(t, g, me, shop, game.ActivateAbilityParams{Targets: effectstest.CardRefs(token)}); err == nil {
		t.Fatal("a token was a legal target")
	}
	before := len(me.Library.Cards)
	effectstest.ApaActivate(t, g, me, shop, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if effectstest.FindBattlefieldCardForTest(g, bear) != nil {
		t.Error("the Bear is still on the battlefield")
	}
	if len(me.Library.Cards) != before+1 {
		t.Errorf("the library holds %d cards, want %d", len(me.Library.Cards), before+1)
	}
}
