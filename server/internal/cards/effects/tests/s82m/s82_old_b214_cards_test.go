package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b214_cards_test.go — S82 OLD batch 214 (#5530, Dominaria): Sentinel
// of the Pearl Trident, Corrosive Ooze, Seismic Shift, Dauntless Bodyguard,
// Guardians of Koilos, Merfolk Trickster, Thallid Omnivore, Blessing of
// Belzenlok, Evra, Halcyon Witness, Sorcerer's Wand and Shield of the Realm.
// Each test plays the real catalog entry through the engine.

const (
	oldB214Sentinel  = "2bddbd7d-9b6b-4166-88de-368d91c7fe4f"
	oldB214Ooze      = "416cca4a-b1b2-49d5-9cac-ec347056bdb0"
	oldB214Seismic   = "b81894bc-d5ca-44ee-b837-bdb8fec32881"
	oldB214Bodyguard = "70973a2d-66e6-4662-ab10-99678775dda1"
	oldB214Guardians = "8f593bed-4be8-403d-bce2-fb769658e673"
	oldB214Trickster = "f712aed5-d6e0-45d3-ae20-518b765ce9da"
	oldB214Omnivore  = "b9bb1b0e-38ce-40f1-882b-85f6c95610ff"
	oldB214Blessing  = "c4a03986-4db6-4114-9a58-5a8a3a7a2461"
	oldB214Evra      = "cf4b538d-7115-406d-94bf-bfa25197b6ef"
	oldB214Wand      = "0edc8ad3-2920-4b25-8a86-98ce216111d1"
	oldB214Shield    = "d669ed30-efe9-4455-813a-c747b123333f"
)

func TestOldB214EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB214Sentinel:  "Sentinel of the Pearl Trident",
		oldB214Ooze:      "Corrosive Ooze",
		oldB214Seismic:   "Seismic Shift",
		oldB214Bodyguard: "Dauntless Bodyguard",
		oldB214Guardians: "Guardians of Koilos",
		oldB214Trickster: "Merfolk Trickster",
		oldB214Omnivore:  "Thallid Omnivore",
		oldB214Blessing:  "Blessing of Belzenlok",
		oldB214Evra:      "Evra, Halcyon Witness",
		oldB214Wand:      "Sorcerer's Wand",
		oldB214Shield:    "Shield of the Realm",
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

func oldB214Attach(g *game.Game, equipment, host uuid.UUID) {
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(equipment, game.TargetRef{Kind: game.TargetCard, ID: host})
	})
}

// --- Sentinel of the Pearl Trident -------------------------------------------

func TestOldB214SentinelBlinksAHistoricPermanentAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Shield of the Realm", "Artifact — Equipment", oldB214Shield, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Sentinel of the Pearl Trident", "Creature — Merfolk Soldier", oldB214Sentinel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, bear) {
		t.Errorf("a non-historic creature is a legal target: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneExile {
		t.Fatalf("the artifact is in %v, want exile", effectstest.B12ZoneOf(g, rock))
	}
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	// It comes back as a new object (CR 400.7).
	if !effectstest.OnBattlefieldByName(g, "Shield of the Realm") || g.Battlefield.Contains(rock) {
		t.Error("the artifact did not return at the next end step as a new object")
	}
}

func TestOldB214SentinelMayDeclineTheExile(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	effectstest.CastCatalogSpell(t, g, "Sentinel of the Pearl Trident", "Creature — Merfolk Soldier", oldB214Sentinel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact was exiled although the player declined")
	}
}

// --- Guardians of Koilos -----------------------------------------------------

func TestOldB214GuardiansReturnsAnotherHistoricPermanentToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	guardians := effectstest.CastCatalogSpell(t, g, "Guardians of Koilos", "Artifact Creature — Construct", oldB214Guardians, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil &&
		(effectstest.HasID(p.PickTargetCards, bear) || effectstest.HasID(p.PickTargetCards, guardians)) {
		t.Errorf("only another historic permanent is a legal target: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !me.Hand.Contains(rock) {
		t.Error("the historic permanent did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(guardians) || !g.Battlefield.Contains(bear) {
		t.Error("the Guardians or the Bear left the battlefield")
	}
}

// --- Merfolk Trickster -------------------------------------------------------

func TestOldB214TricksterTapsAndStripsAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.Pr7bCreature(g, opp.ID, "Their Flyer", 2, "{2}", nil, "flying")
	effectstest.CastCatalogSpell(t, g, "Merfolk Trickster", "Creature — Merfolk Wizard", oldB214Trickster, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("your own creature is a legal target: %v", p.PickTargetCards)
	}
	effectstest.PickCard(t, g, me.ID, flyer)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, flyer) {
		t.Error("the creature did not tap")
	}
	if effectstest.HasEffectiveKeyword(t, g, flyer, "flying") {
		t.Error("the creature kept flying")
	}
	effectstest.EndTurn(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, flyer, "flying") {
		t.Error("the creature did not get its abilities back at end of turn")
	}
}

// --- Seismic Shift -----------------------------------------------------------

func TestOldB214SeismicShiftDestroysALandAndStopsTwoBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	land := effectstest.B12Permanent(g, them.ID, "Mountain", "Basic Land — Mountain")
	a := effectstest.B12Push(g, them.ID, "Blocker A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, them.ID, "Blocker B", "Creature — Bear", "", 2, 2)
	c := effectstest.B12Push(g, them.ID, "Blocker C", "Creature — Bear", "", 2, 2)
	attacker := effectstest.B12Push(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Seismic Shift", "Sorcery", oldB214Seismic, effectstest.CardRefs(land, a, b))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the land was not destroyed")
	}
	effectstest.AttackIntoBlocks(t, g, them.ID, attacker)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if effectstest.BlockRefusal(t, g, id, attacker) == "" {
			t.Errorf("targeted creature %s blocked", name)
		}
	}
	if got := effectstest.BlockRefusal(t, g, c, attacker); got != "" {
		t.Errorf("an untargeted creature was refused: %q", got)
	}
}

func TestOldB214SeismicShiftNeedsALandTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	_, them := effectstest.SeatsForCombat(g)
	bear := effectstest.B12Push(g, them.ID, "Blocker", "Creature — Bear", "", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Seismic Shift", "Sorcery", oldB214Seismic, effectstest.CardRefs(bear)); err == nil {
		t.Error("Seismic Shift was cast without a land target")
	}
}

// --- Dauntless Bodyguard -----------------------------------------------------

func TestOldB214BodyguardSacrificesToProtectTheChosenCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	guard := effectstest.CastCatalogSpell(t, g, "Dauntless Bodyguard", "Creature — Human Knight", oldB214Bodyguard, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no choose-cards prompt for the chosen creature")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{guard}); err == nil {
		t.Fatal("the Bodyguard itself was a legal pick")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, guard, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(guard) {
		t.Error("the Bodyguard was not sacrificed")
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("the chosen creature did not gain indestructible")
	}
}

// --- Thallid Omnivore --------------------------------------------------------

func TestOldB214OmnivorePumpsAndGainsLifeOnlyForASaproling(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		wantLife       int
	}{{"Saproling", "Creature — Saproling", 2}, {"Bear", "Creature — Bear", 0}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		omni := effectstest.B12Push(g, me.ID, "Thallid Omnivore", "Creature — Fungus", oldB214Omnivore, 3, 3)
		fodder := effectstest.B12Creature(g, me.ID, tc.name, tc.typeLine, 1, 1)
		effectstest.ApaMana(me, "C")
		before := me.Life
		if err := g.ActivateCatalogAbility(me.ID, omni, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(fodder) {
			t.Errorf("%s: the creature was not sacrificed", tc.name)
		}
		if p, to := effectstest.PtOf(t, g, omni); p != 5 || to != 5 {
			t.Errorf("%s: Omnivore is %d/%d, want 5/5", tc.name, p, to)
		}
		if got := me.Life - before; got != tc.wantLife {
			t.Errorf("%s: life changed by %d, want %d", tc.name, got, tc.wantLife)
		}
	}
}

// --- Blessing of Belzenlok ---------------------------------------------------

func TestOldB214BlessingPumpsAndGivesALegendaryCreatureLifelink(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		lifelink bool
	}{{"Creature — Bear", false}, {"Legendary Creature — Bear", true}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Target", tc.typeLine, 2, 2)
		effectstest.CastCatalogSpell(t, g, "Blessing of Belzenlok", "Instant", oldB214Blessing, effectstest.CardRefs(bear))
		effectstest.PassPriorityAroundTable(t, g)
		if p, to := effectstest.PtOf(t, g, bear); p != 4 || to != 3 {
			t.Errorf("%s: %d/%d, want 4/3", tc.typeLine, p, to)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, bear, "lifelink"); got != tc.lifelink {
			t.Errorf("%s: lifelink = %v, want %v", tc.typeLine, got, tc.lifelink)
		}
	}
}

// --- Evra, Halcyon Witness ---------------------------------------------------

func TestOldB214EvraExchangesYourLifeTotalWithItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	evra := effectstest.B12Push(g, me.ID, "Evra, Halcyon Witness", "Legendary Creature — Avatar", oldB214Evra, 4, 4)
	g.WithWriteLock(func() { me.Life = 9 })
	effectstest.ApaMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, evra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != 4 {
		t.Errorf("life is %d, want 4 (Evra's old power)", me.Life)
	}
	if p, _ := effectstest.PtOf(t, g, evra); p != 9 {
		t.Errorf("Evra's power is %d, want 9 (the old life total)", p)
	}
}

// --- Sorcerer's Wand ---------------------------------------------------------

func TestOldB214WandPingsOneUnlessTheCreatureIsAWizard(t *testing.T) {
	for _, tc := range []struct {
		typeLine string
		want     int
	}{{"Creature — Bear", 1}, {"Creature — Human Wizard", 2}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		holder := effectstest.B12Creature(g, me.ID, "Holder", tc.typeLine, 1, 1)
		wand := effectstest.PushCatalogPermanent(g, me.ID, "Sorcerer's Wand", "Artifact — Equipment", oldB214Wand, false)
		oldB214Attach(g, wand, holder)
		idx, ref := effectstest.GrantedActivatedIndex(t, g, holder)
		if idx < 0 {
			t.Fatal("the equipped creature has no granted ability")
		}
		if err := g.ActivateCatalogAbility(me.ID, holder, idx, game.ActivateAbilityParams{Ref: ref, Strict: true}); err == nil {
			t.Error("the ability was activated without a target")
		}
		before := opp.Life
		if err := g.ActivateCatalogAbility(me.ID, holder, idx, game.ActivateAbilityParams{Ref: ref, Strict: true, Targets: effectstest.P8Player(opp.ID)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := before - opp.Life; got != tc.want {
			t.Errorf("%s: dealt %d, want %d", tc.typeLine, got, tc.want)
		}
		if !effectstest.B16Tapped(t, g, holder) {
			t.Errorf("%s: the creature did not tap", tc.typeLine)
		}
	}
}

// --- Shield of the Realm -----------------------------------------------------

func TestOldB214ShieldPrevents2OfEachDamageToTheEquippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	holder := effectstest.B12Creature(g, me.ID, "Holder", "Creature — Bear", 2, 9)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 9)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Bear", 2, 2)
	shield := effectstest.PushCatalogPermanent(g, me.ID, "Shield of the Realm", "Artifact — Equipment", oldB214Shield, false)
	oldB214Attach(g, shield, holder)
	effectstest.Damage(t, g, src, holder, 2)
	if got := effectstest.DamageOn(t, g, holder); got != 0 {
		t.Errorf("2 damage left %d marked, want 0", got)
	}
	effectstest.Damage(t, g, src, holder, 5)
	if got := effectstest.DamageOn(t, g, holder); got != 3 {
		t.Errorf("5 damage left %d marked, want 3", got)
	}
	effectstest.Damage(t, g, src, bare, 5)
	if got := effectstest.DamageOn(t, g, bare); got != 5 {
		t.Errorf("an unequipped creature took %d, want 5", got)
	}
}

// --- Corrosive Ooze ----------------------------------------------------------

func TestOldB214OozeDestroysTheEquipmentOfAnEquippedCreatureItFightsAtEndOfCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Holder", 1, 4)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Sword", "Artifact — Equipment", "", false)
	other := effectstest.PushCatalogPermanent(g, me.ID, "Other Sword", "Artifact — Equipment", "", false)
	oldB214Attach(g, sword, attacker)
	oldB214Attach(g, other, attacker)
	ooze := effectstest.PushCatalogPermanent(g, opp.ID, "Corrosive Ooze", "Creature — Ooze", oldB214Ooze, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ooze, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Fatal("the Equipment was destroyed before end of combat")
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(sword) || g.Battlefield.Contains(other) {
		t.Error("an Equipment attached to the creature survived end of combat")
	}
	if !g.Battlefield.Contains(attacker) {
		t.Error("the equipped creature was destroyed")
	}
}

func TestOldB214OozeIgnoresAnUnequippedCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.PushVanillaCreature(g, me.ID, "Holder", 1, 4)
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Sword", "Artifact — Equipment", "", false)
	ooze := effectstest.PushCatalogPermanent(g, opp.ID, "Corrosive Ooze", "Creature — Ooze", oldB214Ooze, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(ooze, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	// The Equipment is attached only after the block: not equipped when it was declared.
	oldB214Attach(g, sword, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sword) {
		t.Error("the Ooze destroyed Equipment although the creature was not equipped when it blocked")
	}
}
