package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b296_cards_test.go — Oath of the Gatewatch devoid cards, S82 OLD
// batch 296, played through the real catalog entries.

const (
	oldB296Cultivator   = "6e62d553-d0d0-400e-a423-f0c3c1fdc911"
	oldB296Inverter     = "941350ba-30e2-4549-865f-ca1bd7626782"
	oldB296FlayerDrone  = "9e275428-3562-4ed6-8a49-3bb202a5130a"
	oldB296Aggressor    = "aa0c6fea-c90c-4b65-aa1e-18e5b34ea6e3"
	oldB296Infiltrator  = "b93732e4-7aeb-4c37-a415-2b0ff46e2056"
	oldB296Obligator    = "d09402cb-cf64-493e-a82d-244b31700710"
	oldB296Bearer       = "d69c8c55-6433-4cc5-bdc7-9722f121b8e2"
	oldB296SkyScourer   = "d816df03-c79b-4c63-913f-41b38afc173c"
	oldB296Mindmelter   = "dbcceed6-51a9-4ab4-b9a8-d65140661195"
	oldB296Abstruse     = "e8ebf5cb-3a26-4bbd-812b-414c129e4bb7"
	oldB296DreadDefiler = "ebdd0027-3340-459e-a598-dd416e76744f"
)

func TestOldB296CardsAreFullWithNoCaveats(t *testing.T) {
	for id, name := range map[string]string{
		oldB296Cultivator: "Cultivator Drone", oldB296Inverter: "Inverter of Truth", oldB296FlayerDrone: "Flayer Drone",
		oldB296Aggressor: "Eldrazi Aggressor", oldB296Infiltrator: "Dimensional Infiltrator", oldB296Obligator: "Eldrazi Obligator",
		oldB296Bearer: "Bearer of Silence", oldB296SkyScourer: "Sky Scourer", oldB296Mindmelter: "Mindmelter",
		oldB296Abstruse: "Abstruse Interference", oldB296DreadDefiler: "Dread Defiler",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
		if !effectstest.HasKeyword(spec.PrintedKeywords, game.KeywordDevoid) {
			t.Errorf("%s does not declare devoid", name)
		}
	}
}

func oldB296PickTarget(t *testing.T, g *game.Game, me uuid.UUID, ref game.TargetRef) {
	t.Helper()
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me, ref); err != nil {
			t.Fatalf("pick target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
}

func TestOldB296CultivatorDroneManaPaysOnlyColorlessThings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.SeedPermanentWithOracle(g, me.ID, "Cultivator Drone", "Creature — Eldrazi Drone", oldB296Cultivator)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, drone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "C" || len(units[0].Restrictions) != 1 {
		t.Fatalf("pool %v, want one restricted {C}", me.ManaPool)
	}
	blue := effectstest.HandCardFull(me, "Blue Spell", "Instant", "{C}", "", []string{"U"})
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the mana could not pay a cost that contains {C}: %v", err)
	}
}

func TestOldB296CultivatorDroneManaCannotPayAColoredSpellWithoutC(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.SeedPermanentWithOracle(g, me.ID, "Cultivator Drone", "Creature — Eldrazi Drone", oldB296Cultivator)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateManaAbility(me.ID, drone, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility: %v", err)
	}
	blue := effectstest.HandCardFull(me, "Blue Spell", "Instant", "{1}", "", []string{"U"})
	if err := g.CastSpell(me.ID, blue, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the restricted mana paid a blue spell whose cost has no {C}")
	}
	rock := effectstest.HandCardFull(me, "Rock", "Artifact", "{1}", "", nil)
	if err := g.CastSpell(me.ID, rock, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the restricted mana could not pay a colorless spell: %v", err)
	}
}

func TestOldB296InverterOfTruthExilesTheLibraryAndShufflesTheGraveyardIn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushGraveyardCardTyped(me, "Old One", "Creature — Bear")
	effectstest.PushGraveyardCardTyped(me, "Old Two", "Creature — Bear")
	effectstest.AdvanceToMain(t, g)
	lib := me.Library.Size()
	exile := len(g.Exile.Cards)
	effectstest.CastCatalogSpell(t, g, "Inverter of Truth", "Creature — Eldrazi", oldB296Inverter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Size(); got != 2 {
		t.Errorf("library %d, want the 2 graveyard cards (was %d)", got, lib)
	}
	if got := len(g.Exile.Cards) - exile; got != lib {
		t.Errorf("%d cards exiled, want %d", got, lib)
	}
	if me.Graveyard.Size() != 0 {
		t.Errorf("graveyard still holds %d cards", me.Graveyard.Size())
	}
}

func TestOldB296FlayerDroneDrainsForOtherColorlessCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Flayer Drone", "Creature — Eldrazi Drone", oldB296FlayerDrone, 3, 1)
	effectstest.AdvanceToMain(t, g)
	before := opp.Life
	bear := effectstest.HandCardFull(me, "Green Bear", "Creature — Bear", "", "", []string{"G"})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the green creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Fatalf("a colored creature drained %d", before-opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Golem", "Artifact Creature — Golem", "", nil)
	oldB296PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if opp.Life != before-1 {
		t.Errorf("life %d, want %d after a colorless creature entered", opp.Life, before-1)
	}
}

func TestOldB296EldraziAggressorHasHasteOnlyBesideAnotherColorlessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	agg := effectstest.B12Push(g, me.ID, "Eldrazi Aggressor", "Creature — Eldrazi Drone", oldB296Aggressor, 2, 3)
	if effectstest.HasEffectiveKeyword(t, g, agg, "haste") {
		t.Fatal("haste with no other colorless creature")
	}
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear",
		Colors: []string{"G"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	if effectstest.HasEffectiveKeyword(t, g, agg, "haste") {
		t.Fatal("haste beside a colored creature")
	}
	effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, agg, "haste") {
		t.Error("no haste beside another colorless creature")
	}
}

func TestOldB296DimensionalInfiltratorExilesAndMayBounceOnALand(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		inf := effectstest.B12Push(g, me.ID, "Dimensional Infiltrator", "Creature — Eldrazi", oldB296Infiltrator, 2, 1)
		effectstest.AdvanceToMain(t, g)
		top := "Spell"
		typeLine := "Sorcery"
		if land {
			top, typeLine = "Forest", "Basic Land — Forest"
		}
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: top, TypeLine: typeLine, Owner: opp.ID, Controller: opp.ID})
		effectstest.FloatMana(t, g, me, "{C}")
		effectstest.FloatMana(t, g, me, "{C}")
		before := opp.Library.Size()
		effectstest.S82Activate(t, g, me.ID, inf, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Library.Size(); got != before-1 {
			t.Fatalf("land=%v: library %d, want %d", land, got, before-1)
		}
		asked := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceConfirm {
				asked = true
			}
		}
		if asked != land {
			t.Errorf("land=%v: asked whether to return the Infiltrator = %v", land, asked)
		}
	}
}

func TestOldB296ObligatorStealsOnlyWhenTheCostIsPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		effectstest.AdvanceToMain(t, g)
		effectstest.CastCatalogSpell(t, g, "Eldrazi Obligator", "Creature — Eldrazi", oldB296Obligator, nil)
		effectstest.FloatMana(t, g, me, "{C}")
		effectstest.FloatMana(t, g, me, "{C}")
		oldB296PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim})
		if effectstest.HasPayUnlessFor(g, me.ID) {
			effectstest.AnswerPayUnless(t, g, me.ID, pay)
		}
		c, _ := g.LookupCardForEffect(victim)
		if (c.Controller == me.ID) != pay {
			t.Errorf("pay=%v: the victim's controller is %v", pay, c.Controller)
		}
		if pay && !effectstest.HasEffectiveKeyword(t, g, victim, "haste") {
			t.Error("the stolen creature has no haste")
		}
	}
}

func TestOldB296BearerOfSilenceMakesAnOpponentSacrificeWhenPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Bearer of Silence", "Creature — Eldrazi", oldB296Bearer, nil)
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.FloatMana(t, g, me, "{C}")
	oldB296PickTarget(t, g, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID})
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's only creature was not sacrificed")
	}
}

func TestOldB296SkyScourerGrowsOnlyForColorlessSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sky := effectstest.B12Push(g, me.ID, "Sky Scourer", "Creature — Eldrazi Drone", oldB296SkyScourer, 1, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.PowerOf(t, g, sky); p != 2 {
		t.Errorf("power %d after a colorless spell, want 2", p)
	}
}

func TestOldB296MindmelterIsUnblockableAndExilesFromHandAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mm := effectstest.B12Push(g, me.ID, "Mindmelter", "Creature — Eldrazi Drone", oldB296Mindmelter, 2, 2)
	effectstest.AdvanceToMain(t, g)
	spare := effectstest.HandCard(opp, "Spare", "Sorcery")
	before := opp.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.S82Activate(t, g, me.ID, mm, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, spare)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size(); got != before-1 {
		t.Errorf("hand %d, want %d", got, before-1)
	}
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.BrAttack(t, g, mm)
	if err := g.DeclareBlocker(blocker, mm); err == nil {
		t.Error("the Mindmelter was blocked")
	}
}

func TestOldB296AbstruseInterferenceCountersUnlessPaidAndMakesAScion(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim := effectstest.CastCatalogSpell(t, g, "Their Bear", "Creature — Bear", "", nil)
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Abstruse Interference", "Instant", oldB296Abstruse, []game.TargetRef{{Kind: game.TargetCard, ID: victim}})
		if pay {
			effectstest.B06AddMana(me, "C")
		}
		effectstest.PassUntilTaxed(t, g, me.ID)
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		_, resolved := effectstest.BattlefieldCard(g, victim)
		if pay && !resolved {
			t.Error("the spell was countered although its controller paid {1}")
		}
		if !pay && (resolved || !me.Graveyard.Contains(victim)) {
			t.Error("the spell was not countered")
		}
		if n := effectstest.CountBattlefieldNamed(g, opp.ID, "Eldrazi Scion"); n != 1 {
			t.Errorf("pay=%v: the caster controls %d Scions, want 1", pay, n)
		}
	}
}

func TestOldB296DreadDefilerDrainsByTheExiledCardsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dd := effectstest.B12Push(g, me.ID, "Dread Defiler", "Creature — Eldrazi", oldB296DreadDefiler, 6, 8)
	id := effectstest.PushGraveyardPermanent(me, "Big Bear", "Creature — Bear", "{3}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].Power = 5
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 4; i++ {
		effectstest.FloatMana(t, g, me, "{C}")
	}
	before := opp.Life
	effectstest.S82Activate(t, g, me.ID, dd, 0, game.ActivateAbilityParams{Targets: effectstest.PlayerRefs(opp.ID), ExileIDs: []uuid.UUID{id}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-5 {
		t.Errorf("life %d, want %d", opp.Life, before-5)
	}
}
