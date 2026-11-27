package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB552Harpoon  = "2856844d-f727-48aa-8387-cb6079afb40c"
	oldB552Crush    = "81686da0-85fc-4938-962b-24443de3ada1"
	oldB552Brigid   = "929ad5c3-4efe-44c3-93ac-eda432814385"
	oldB552Spitfire = "b6f2bec3-cbff-40a5-a2ed-ec078f9fabf5"
	oldB552Fodder   = "bec0127a-ab3c-4923-93c2-c43ba8998091"
	oldB552Bonfire  = "e1feea85-2f06-43aa-9b3c-eb63085276f5"
	oldB552Pitcher  = "e6991ab8-36f7-40c4-ad37-06e847d92664"
)

func TestOldB552CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB552Harpoon: "Harpoon Sniper", oldB552Crush: "Crush Underfoot", oldB552Brigid: "Brigid, Hero of Kinsbaile",
		oldB552Spitfire: "Flamekin Spitfire", oldB552Fodder: "Fodder Launch", oldB552Bonfire: "Consuming Bonfire",
		oldB552Pitcher: "Tar Pitcher",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

func TestOldB552TarPitcherSacrificesAGoblinForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pitcher := effectstest.B12Push(g, me.ID, "Tar Pitcher", "Creature — Goblin Shaman", oldB552Pitcher, 2, 2)
	goblin := effectstest.B12Push(g, me.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, pitcher, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Error("a Bear was accepted as the Goblin to sacrifice")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, pitcher, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{goblin}})
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want 2 damage", life, opp.Life)
	}
	if !effectstest.FindBattlefieldCardForTest(g, pitcher).Tapped {
		t.Error("the Pitcher did not tap")
	}
}

func TestOldB552TarPitcherMaySacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	pitcher := effectstest.B12Push(g, me.ID, "Tar Pitcher", "Creature — Goblin Shaman", oldB552Pitcher, 2, 2)
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, pitcher, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, SacrificeIDs: []uuid.UUID{pitcher}})
	if g.Battlefield.Contains(pitcher) || opp.Life != life-2 {
		t.Errorf("the Pitcher on the battlefield = %v, life %d → %d; want it sacrificed and 2 damage", g.Battlefield.Contains(pitcher), life, opp.Life)
	}
}

func TestOldB552FlamekinSpitfirePingsForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	spit := effectstest.B12Push(g, me.ID, "Flamekin Spitfire", "Creature — Elemental Shaman", oldB552Spitfire, 1, 1)
	life := opp.Life
	for range 2 {
		effectstest.RavB2MustActivate(t, g, me.ID, spit, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}}, "R", "R", "R", "R")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d → %d, want two pings", life, opp.Life)
	}
}

func TestOldB552HarpoonSniperDealsDamageEqualToYourMerfolkToACombatant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	sniper := effectstest.B12Push(g, opp.ID, "Harpoon Sniper", "Creature — Merfolk Archer", oldB552Harpoon, 2, 2)
	effectstest.B12Push(g, opp.ID, "Other Merfolk", "Creature — Merfolk", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	big := effectstest.B12Push(g, me.ID, "Big", "Creature — Ogre", "", 2, 3)
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Ogre", "", 2, 3)
	effectstest.DeclareAttack(t, g, opp.ID, big)
	if err := g.ActivateCatalogAbility(opp.ID, sniper, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: idle}}}); err == nil {
		t.Error("a creature that is not attacking or blocking was a legal target")
	}
	effectstest.RavB2MustActivate(t, g, opp.ID, sniper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}, "W")
	// Two Merfolk (the Sniper and the other one): 2 damage, a 2/3 survives.
	if !g.Battlefield.Contains(big) {
		t.Fatal("two Merfolk should deal 2, not enough to kill a 2/3")
	}
	effectstest.B12Push(g, opp.ID, "Third Merfolk", "Creature — Merfolk", "", 1, 1)
	effectstest.UntapForTest(g, sniper)
	effectstest.RavB2MustActivate(t, g, opp.ID, sniper, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}, "W")
	if g.Battlefield.Contains(big) {
		t.Error("three Merfolk deal 3 damage, which with the first 2 should have killed the 2/3")
	}
}

func TestOldB552BrigidHitsEachCombatantOfTheTargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	brigid := effectstest.B12Push(g, opp.ID, "Brigid, Hero of Kinsbaile", "Legendary Creature — Kithkin Archer", oldB552Brigid, 2, 3)
	a1 := effectstest.B12Push(g, me.ID, "Attacker One", "Creature — Bear", "", 2, 2)
	a2 := effectstest.B12Push(g, me.ID, "Attacker Two", "Creature — Bear", "", 2, 2)
	idle := effectstest.B12Push(g, me.ID, "Idle", "Creature — Bear", "", 2, 2)
	mine := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a1, a2)
	effectstest.AssertKeywords(t, g, brigid, "first strike")
	effectstest.RavB2MustActivate(t, g, opp.ID, brigid, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
	for _, id := range []uuid.UUID{a1, a2} {
		if g.Battlefield.Contains(id) {
			t.Error("an attacker survived 2 damage")
		}
	}
	if !g.Battlefield.Contains(idle) || !g.Battlefield.Contains(mine) {
		t.Error("a creature that was not attacking, or the other player's, was hit")
	}
}

func TestOldB552ConsumingBonfireBulletsHaveTheirOwnTargetsAndDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	elemental := effectstest.B12Creature(g, opp.ID, "Elemental", "Creature — Elemental", 5, 8)
	treefolk := effectstest.B12Creature(g, opp.ID, "Treefolk", "Creature — Treefolk", 6, 7)
	spell := func(mode int, target uuid.UUID) error {
		id := uuid.New()
		g.Seats[g.Turn.ActiveSeat].Hand.PushTop(game.Card{InstanceID: id, Name: "Consuming Bonfire", TypeLine: "Kindred Sorcery — Elemental", OracleID: oldB552Bonfire,
			Owner: g.Seats[g.Turn.ActiveSeat].ID, Controller: g.Seats[g.Turn.ActiveSeat].ID})
		effectstest.AdvanceToMain(t, g)
		return g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode},
			Targets: []game.TargetRef{effectstest.ModeRef(game.TargetCard, target, 0, 0)}})
	}
	if err := spell(0, elemental); err == nil {
		t.Error("an Elemental was a legal target of the 4-damage bullet")
	}
	if err := spell(1, ogre); err == nil {
		t.Error("an Ogre was a legal target of the Treefolk bullet")
	}
	if err := spell(0, ogre); err != nil {
		t.Fatalf("4 damage at an Ogre: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("4 damage should kill a 4/4")
	}
	if err := spell(1, treefolk); err != nil {
		t.Fatalf("7 damage at a Treefolk: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(treefolk) {
		t.Error("7 damage should kill a 6/7")
	}
	if !g.Battlefield.Contains(elemental) {
		t.Error("the Elemental was hit")
	}
}

func TestOldB552FodderLaunchSacrificesAGoblinShrinksAndHitsTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	goblin := effectstest.B12Push(g, me.ID, "Goblin", "Creature — Goblin", "", 1, 1)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Ogre", "", 5, 5)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fodder Launch", TypeLine: "Kindred Sorcery — Goblin", OracleID: oldB552Fodder, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{bear},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err == nil {
		t.Error("a Bear was accepted as the Goblin to sacrifice")
	}
	life := opp.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, SacrificeIDs: []uuid.UUID{goblin},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if g.Battlefield.Contains(goblin) {
		t.Error("the Goblin was not sacrificed as the spell was cast")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("-5/-5 should kill a 5/5")
	}
	if opp.Life != life-5 {
		t.Errorf("the creature's controller's life %d → %d, want 5 damage", life, opp.Life)
	}
}

func TestOldB552CrushUnderfootDamageComesFromTheChosenGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	giant := effectstest.B12Push(g, me.ID, "Giant", "Creature — Giant Warrior", "", 4, 4)
	effectstest.B12Push(g, me.ID, "Big Bear", "Creature — Bear", "", 9, 9)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Ogre", "", 3, 4)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Crush Underfoot", TypeLine: "Kindred Instant — Giant", OracleID: oldB552Crush, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pc := oldB552OpenPermanentPick(g)
	if pc == nil {
		t.Fatal("no prompt to choose a Giant")
	}
	if len(pc.ChooseCards) != 1 || pc.ChooseCards[0] != giant {
		t.Errorf("the offer is %v, want only the Giant", pc.ChooseCards)
	}
	effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {giant}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("4 damage from the Giant should kill a 3/4")
	}
}

func TestOldB552CrushUnderfootWithNoGiantDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Big Bear", "Creature — Bear", "", 9, 9)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Ogre", "", 3, 4)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Crush Underfoot", TypeLine: "Kindred Instant — Giant", OracleID: oldB552Crush, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if oldB552OpenPermanentPick(g) != nil || !g.Battlefield.Contains(victim) {
		t.Error("with no Giant the spell should do nothing")
	}
}

// oldB552OpenPermanentPick is the open "choose a permanent you control" prompt.
func oldB552OpenPermanentPick(g *game.Game) *game.PendingChoice {
	for _, pc := range g.PendingChoices {
		if pc != nil && (pc.Kind == game.PendingChoiceOwnPermanents || pc.Kind == game.PendingChoiceTheirPermanents) {
			return pc
		}
	}
	return nil
}
