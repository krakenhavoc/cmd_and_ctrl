package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b605_test.go — Mirrodin, batch 605 (#5606): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB605GrimReminderOracle = "12684e95-2652-4c3c-b2c7-021b62344d77"
	oldB605WoebearerOracle    = "5ed93d71-3f62-49d4-a9ca-5728c44b0633"
	oldB605DevourerOracle     = "89075fbe-cd8a-4054-abab-fcaa2aedddfb"
	oldB605ScavengerOracle    = "fdccdd9f-6bf9-4339-ad7a-e9dd8f771b22"
	oldB605ScaleOracle        = "1ab0a315-0d0d-4247-9bb9-107b35b353d6"
	oldB605ShamblerOracle     = "83d2ff7f-4d6e-4cc0-b8a9-3c5b3236eea0"
	oldB605IrradiateOracle    = "84d45389-a085-44bc-a3fb-1a5f7cc6cbe0"
	oldB605ToothOracle        = "e5e7e7e4-a774-4d68-8767-8ba2bae38250"
	oldB605KaldraOracle       = "35f63c60-b0a1-459d-bf51-a8549fb06930"
	oldB605LashOracle         = "414f1c42-6b0e-49ce-8b44-17a75b996b68"
	oldB605ScytheOracle       = "9dc26b9f-3b02-46f2-88f8-eb6dffa04ee5"
	oldB605GauntletsOracle    = "e8b1672c-e5b1-41e7-9f34-c822f1587369"
)

func TestOldB605CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB605GrimReminderOracle: "Grim Reminder", oldB605WoebearerOracle: "Woebearer",
		oldB605DevourerOracle: "Nim Devourer", oldB605ScavengerOracle: "Moriok Scavenger",
		oldB605ScaleOracle: "Scale of Chiss-Goria", oldB605ShamblerOracle: "Nim Shambler",
		oldB605IrradiateOracle: "Irradiate", oldB605ToothOracle: "Tooth of Chiss-Goria",
		oldB605KaldraOracle: "Sword of Kaldra", oldB605LashOracle: "Nightmare Lash",
		oldB605ScytheOracle: "Scythe of the Wretched", oldB605GauntletsOracle: "Vulshok Gauntlets",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Grim Reminder ---

func oldB605CastGrimReminder(t *testing.T, g *game.Game, me *game.Player, pick uuid.UUID) {
	t.Helper()
	effectstest.CastCatalogSpell(t, g, "Grim Reminder", "Instant", oldB605GrimReminderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, pick)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB605GrimReminderDrainsOnlyTheOpponentWhoCastTheNamedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, caster, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	needle := effectstest.PushLibraryCard(me, game.Card{Name: "Needle", TypeLine: "Instant"})
	effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastFree(t, g, caster, "Needle", "Instant")
	effectstest.CastFree(t, g, bystander, "Other", "Instant")
	before, bystanderBefore := caster.Life, bystander.Life
	oldB605CastGrimReminder(t, g, me, needle)
	if caster.Life != before-6 {
		t.Errorf("the opponent who cast Needle has %d life, want %d", caster.Life, before-6)
	}
	if bystander.Life != bystanderBefore {
		t.Errorf("an opponent who cast a different spell lost life: %d", bystander.Life)
	}
	if !me.Library.Contains(needle) {
		t.Error("the searched card left the library: it is only revealed")
	}
}

func TestOldB605GrimReminderCostsNothingWhenNoOneCastTheName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	needle := effectstest.PushLibraryCard(me, game.Card{Name: "Needle", TypeLine: "Instant"})
	effectstest.CastFree(t, g, opp, "Other", "Instant")
	before := opp.Life
	oldB605CastGrimReminder(t, g, me, needle)
	if opp.Life != before {
		t.Errorf("an opponent lost life: %d, want %d", opp.Life, before)
	}
}

func TestOldB605GrimReminderCannotFindALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.CastCatalogSpell(t, g, "Grim Reminder", "Instant", oldB605GrimReminderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.SearchPromptFor(g, me.ID); p != nil {
		for _, id := range p.SearchCards {
			if id == forest {
				t.Fatal("a land was offered to a search for a nonland card")
			}
		}
	}
}

func TestOldB605GrimReminderReturnsOnlyInYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.PushCatalogGraveyardCard(me, "Grim Reminder", "Instant", oldB605GrimReminderOracle, 0, 0)
	params := game.ActivateAbilityParams{}
	effectstest.FloatMana(t, g, me, "{B}{B}")
	effectstest.AcRefused(t, g, me.ID, card, 0, params) // the draw step
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AcRefused(t, g, me.ID, card, 0, params) // another player's upkeep
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.FloatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, params); err != nil {
		t.Fatalf("activate in your upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) || me.Graveyard.Contains(card) {
		t.Error("Grim Reminder did not return to hand")
	}
}

// --- Woebearer ---

func TestOldB605WoebearerReturnsACreatureCardOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	spell := effectstest.GraveyardCardOf(g, me, "Bolt", "Instant", "{R}", "")
	woe := effectstest.B12Push(g, me.ID, "Woebearer", "Creature — Zombie", oldB605WoebearerOracle, 2, 3)
	if !effectstest.HasEffectiveKeyword(t, g, woe, "fear") {
		t.Error("Woebearer lacks fear")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, woe, opp.ID, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, spell) {
		t.Error("an instant card is offered to a creature clause")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card was not returned")
	}
}

func TestOldB605WoebearerMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	woe := effectstest.B12Push(g, me.ID, "Woebearer", "Creature — Zombie", oldB605WoebearerOracle, 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, woe, opp.ID, 2)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) {
		t.Error("the card was returned though the player declined")
	}
}

// --- Moriok Scavenger ---

func TestOldB605ScavengerReturnsOnlyAnArtifactCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem := effectstest.GraveyardCardOf(g, me, "Golem", "Artifact Creature — Golem", "{3}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Bear", "Creature — Bear", "{1}{G}", "")
	rock := effectstest.GraveyardCardOf(g, me, "Rock", "Artifact", "{1}", "")
	effectstest.CastHoldCreature(t, g, "Moriok Scavenger", "Creature — Human Rogue", oldB605ScavengerOracle, 2, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil &&
		(effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, rock)) {
		t.Error("a card that is not an artifact creature card is offered")
	}
	effectstest.PickCard(t, g, me.ID, golem)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(golem) {
		t.Error("the artifact creature card was not returned")
	}
}

// --- Nim Devourer and Nim Shambler ---

func TestOldB605NimDevourerGrowsWithArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dev := effectstest.B12Push(g, me.ID, "Nim Devourer", "Creature — Zombie", oldB605DevourerOracle, 4, 1)
	if p, _ := effectstest.EffectivePTOf(t, g, dev); p != 4 {
		t.Fatalf("power %d, want 4", p)
	}
	effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.B12Push(g, me.ID, "Rock Golem", "Artifact Creature — Golem", "", 1, 1)
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	if p, _ := effectstest.EffectivePTOf(t, g, dev); p != 6 {
		t.Errorf("power %d with two artifacts of mine and one of theirs, want 6", p)
	}
}

func TestOldB605NimDevourerReturnsInYourUpkeepThenSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	dev := effectstest.PushCatalogGraveyardCard(me, "Nim Devourer", "Creature — Zombie", oldB605DevourerOracle, 4, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	params := game.ActivateAbilityParams{}
	effectstest.FloatMana(t, g, me, "{B}{B}")
	effectstest.AcRefused(t, g, me.ID, dev, 0, params)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AcRefused(t, g, me.ID, dev, 0, params)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.FloatMana(t, g, me, "{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, dev, 0, params); err != nil {
		t.Fatalf("activate in your upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dev) {
		t.Error("Nim Devourer did not return to the battlefield")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
}

func TestOldB605NimShamblerRegeneratesBySacrificingACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sham := effectstest.B12Push(g, me.ID, "Nim Shambler", "Creature — Zombie", oldB605ShamblerOracle, 2, 1)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	if p, _ := effectstest.EffectivePTOf(t, g, sham); p != 3 {
		t.Errorf("power %d with one artifact, want 3", p)
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.B16Activate(t, g, me.ID, sham, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.DealDamageTo(t, g, src, sham, 5)
	if !g.Battlefield.Contains(sham) {
		t.Error("Nim Shambler died though it was regenerated")
	}
}

// --- Scale and Tooth of Chiss-Goria ---

func TestOldB605ChissGoriaPumpsATargetCreature(t *testing.T) {
	for _, c := range []struct {
		oracle, name string
		dp, dt       int
	}{
		{oldB605ScaleOracle, "Scale of Chiss-Goria", 0, 1},
		{oldB605ToothOracle, "Tooth of Chiss-Goria", 1, 0},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		art := effectstest.PushCatalogPermanent(g, me.ID, c.name, "Artifact", c.oracle, true)
		bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.B16Activate(t, g, me.ID, art, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
		effectstest.PassPriorityAroundTable(t, g)
		p, tough := effectstest.EffectivePTOf(t, g, bear)
		if p != 2+c.dp || tough != 2+c.dt {
			t.Errorf("%s: %d/%d, want %d/%d", c.name, p, tough, 2+c.dp, 2+c.dt)
		}
		if !effectstest.IsTapped(g, art) {
			t.Errorf("%s did not tap", c.name)
		}
	}
}

func TestOldB605ChissGoriaAffinityAndFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	for i := 0; i < 2; i++ {
		effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	}
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	for _, c := range []struct{ oracle string }{{oldB605ScaleOracle}, {oldB605ToothOracle}} {
		if got := effectstest.SelfPricedMV(t, g, me, c.oracle, "Artifact", "{3}"); got != 1 {
			t.Errorf("%s costs %d with two artifacts of mine, want 1", c.oracle, got)
		}
		if got := effectstest.SelfPricedMV(t, g, opp, c.oracle, "Artifact", "{3}"); got != 2 {
			t.Errorf("%s costs %d for the player with one artifact, want 2", c.oracle, got)
		}
	}
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogHandCard(me, "Tooth of Chiss-Goria", "Artifact", oldB605ToothOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Error("the artifact did not enter")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "flash") {
		t.Error("no flash")
	}
}

// --- Irradiate ---

func TestOldB605IrradiateShrinksByTheArtifactsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Rock", "Artifact", "", 0, 0)
	effectstest.B12Push(g, me.ID, "Rock Two", "Artifact", "", 0, 0)
	effectstest.B12Push(g, opp.ID, "Their Rock", "Artifact", "", 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Irradiate", "Instant", oldB605IrradiateOracle, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, victim); p != 2 || tough != 2 {
		t.Errorf("%d/%d, want 2/2 (two artifacts of mine)", p, tough)
	}
}

// --- Equipment ---

func TestOldB605VulshokGauntletsPumpsAndKeepsTheCreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gaunt := effectstest.PushCatalogPermanent(g, me.ID, "Vulshok Gauntlets", "Artifact — Equipment", oldB605GauntletsOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, gaunt, bear, "{C}{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 4 {
		t.Errorf("%d/%d, want 6/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(bear) })
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.IsTapped(g, bear) {
		t.Error("the equipped creature untapped")
	}
}

func TestOldB605NightmareLashCountsSwampsAndEquipsForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lash := effectstest.PushCatalogPermanent(g, me.ID, "Nightmare Lash", "Artifact — Equipment", oldB605LashOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1)
	effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.B12Push(g, me.ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.B12Push(g, g.Seats[1].ID, "Swamp", "Basic Land — Swamp", "", 0, 0)
	effectstest.B12Push(g, me.ID, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.AdvanceToMain(t, g)
	before := me.Life
	if err := g.ActivateCatalogAbility(me.ID, lash, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(bear), Strict: true,
	}); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != before-3 {
		t.Errorf("life %d, want %d", me.Life, before-3)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("%d/%d with two Swamps, want 3/3", p, tough)
	}
}

func TestOldB605SwordOfKaldraExilesACreatureItDamages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Sword of Kaldra", "Legendary Artifact — Equipment", oldB605KaldraOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Giant", 3, 9)
	other := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Giant", 1, 9)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, sword, bear, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 7 || tough != 7 {
		t.Fatalf("%d/%d, want 7/7", p, tough)
	}
	// Damage dealt by something else does not trigger it.
	effectstest.DealDamageTo(t, g, other, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("a creature was exiled by damage the equipped creature did not deal")
	}
	effectstest.DealDamageTo(t, g, bear, victim, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || !g.Exile.Contains(victim) {
		t.Error("the damaged creature was not exiled")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a creature the equipped creature never damaged was exiled")
	}
}

func TestOldB605ScytheReturnsADamagedCreatureAndMovesOnto(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	scythe := effectstest.PushCatalogPermanent(g, me.ID, "Scythe of the Wretched", "Artifact — Equipment", oldB605ScytheOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Elf", 1, 3)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Elf", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.StrictEquip(t, g, me.ID, scythe, bear, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("equip: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A creature the equipped one never damaged dies: nothing happens.
	effectstest.KillCreature(t, g, opp.ID, bystander)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bystander) {
		t.Fatal("a creature the equipped creature did not damage came back")
	}
	effectstest.DealDamageTo(t, g, bear, victim, 1)
	effectstest.KillCreature(t, g, opp.ID, victim)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) {
		t.Fatal("the damaged creature did not return")
	}
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Errorf("the returned creature is controlled by %s, want its returner", got)
	}
	if !effectstest.AttachedTo(t, g, scythe, victim) {
		t.Error("the Scythe was not attached to the returned creature")
	}
}
