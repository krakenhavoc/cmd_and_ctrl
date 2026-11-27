package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB451Throne    = "12afe8e6-eaf0-45c9-8086-4c658db7cb7e"
	oldB451Siren     = "5f358544-7319-4f8e-9d60-a410c4feb774"
	oldB451Crown     = "b1298ce8-b318-4d5b-83f2-e274726dfca5"
	oldB451Skin      = "cfa3f2a4-2724-4bc1-bdaf-244051db95ed"
	oldB451Flight    = "6a4068b0-fb4f-429c-a94e-47849f3eb7ef"
	oldB451Fire      = "8603bf74-faab-4910-8e45-0f2e3b318efb"
	oldB451Brink     = "ae150b53-7c4e-48ce-abcf-cfe1e68bb4d5"
	oldB451Cage      = "e37416cc-f343-44cd-8676-4ce4d4068e85"
	oldB451Pharaoh   = "7e465cfe-d2ee-4420-87f4-945728f58f87"
	oldB451Erasure   = "bb7c069b-8d3f-4ba2-874b-8ec575fd6dd4"
	oldB451Treason   = "9d08af23-9f4a-4097-9abc-3b17475ab744"
	oldB451ForestTyp = "Basic Land — Forest"
)

func TestOldB451CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB451Throne: "Throne of Empires", oldB451Siren: "Alluring Siren", oldB451Crown: "Crown of Empires",
		oldB451Skin: "Skinshifter", oldB451Flight: "Flight", oldB451Fire: "Firebreathing",
		oldB451Brink: "Brink of Disaster", oldB451Cage: "Ice Cage", oldB451Pharaoh: "Vengeful Pharaoh",
		oldB451Erasure: "Jace's Erasure",
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

// oldB451Activate pays `mana` into the pool and activates ability 0, Strict.
func oldB451Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, mana string, targets ...game.TargetRef) error {
	t.Helper()
	effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if mana != "" {
		effectstest.DcMana(t, g, me.ID, mana)
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: targets})
	if err == nil {
		effectstest.PassPriorityAroundTable(t, g)
	}
	return err
}

func oldB451Untap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
}

func TestOldB451ThroneMakesOneSoldierOrFiveWithCrownAndScepter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	throne := effectstest.B12Push(g, me.ID, "Throne of Empires", "Artifact", oldB451Throne, 0, 0)
	if err := oldB451Activate(t, g, me, throne, "{C}"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if n := effectstest.NamedOnBattlefield(g, "Soldier", me.ID); n != 1 {
		t.Fatalf("%d Soldiers, want 1", n)
	}
	effectstest.B12Push(g, me.ID, "Crown of Empires", "Artifact", oldB451Crown, 0, 0)
	oldB451Untap(g, throne)
	if err := oldB451Activate(t, g, me, throne, "{C}"); err != nil {
		t.Fatalf("activate with only the Crown: %v", err)
	}
	if n := effectstest.NamedOnBattlefield(g, "Soldier", me.ID); n != 2 {
		t.Fatalf("%d Soldiers with only the Crown, want 2 (one more)", n)
	}
	effectstest.B12Permanent(g, me.ID, "Scepter of Empires", "Artifact")
	oldB451Untap(g, throne)
	if err := oldB451Activate(t, g, me, throne, "{C}"); err != nil {
		t.Fatalf("activate with both: %v", err)
	}
	if n := effectstest.NamedOnBattlefield(g, "Soldier", me.ID); n != 7 {
		t.Fatalf("%d Soldiers with the Crown and Scepter, want 7 (five more)", n)
	}
	if err := oldB451Activate(t, g, me, throne, "{C}"); err == nil {
		t.Fatal("activated a tapped Throne")
	}
}

func TestOldB451CrownTapsOrStealsWithScepterAndThrone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crown := effectstest.B12Push(g, me.ID, "Crown of Empires", "Artifact", oldB451Crown, 0, 0)
	victim := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	target := game.TargetRef{Kind: game.TargetCard, ID: victim}
	if err := oldB451Activate(t, g, me, crown, "{C}{C}{C}", target); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); !tapped {
		t.Error("the creature was not tapped")
	}
	if got := effectstest.ControllerOf(t, g, victim); got != opp.ID {
		t.Fatal("control changed without the Scepter and Throne")
	}
	oldB451Untap(g, crown)
	oldB451Untap(g, victim)
	effectstest.B12Permanent(g, me.ID, "Scepter of Empires", "Artifact")
	effectstest.B12Permanent(g, me.ID, "Throne of Empires", "Artifact")
	if err := oldB451Activate(t, g, me, crown, "{C}{C}{C}", target); err != nil {
		t.Fatalf("activate with both: %v", err)
	}
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Fatalf("controller %s, want the Crown's controller", got)
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); tapped {
		t.Error("the creature was tapped as well as stolen: \"instead\" replaces the tap")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if got := effectstest.ControllerOf(t, g, victim); got != me.ID {
		t.Error("control returned at cleanup, but it has no stated duration")
	}
}

func TestOldB451SirenForcesTheCreatureToAttackYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	siren := effectstest.B12Push(g, me.ID, "Alluring Siren", "Creature — Siren", oldB451Siren, 1, 1)
	oldB451Untap(g, siren)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := oldB451Activate(t, g, me, siren, "", game.TargetRef{Kind: game.TargetCard, ID: effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)}); err == nil {
		t.Fatal("targeted your own creature")
	}
	if err := oldB451Activate(t, g, me, siren, "", game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	c := effectstest.ApaLive(g, bear)
	found := false
	for _, r := range c.Effective().AttackRequirements {
		if r.MustAttack == me.ID {
			found = true
		}
	}
	if !found {
		t.Fatalf("the creature carries no \"attacks you if able\" requirement: %+v", c.Effective().AttackRequirements)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	for _, r := range effectstest.ApaLive(g, bear).Effective().AttackRequirements {
		if r.MustAttack == me.ID {
			t.Error("the requirement outlived the turn")
		}
	}
}

func TestOldB451SkinshifterBecomesEachFormOncePerTurn(t *testing.T) {
	for i, tc := range []struct {
		subtype    string
		power, tou int
		keyword    string
	}{{"Rhino", 4, 4, "trample"}, {"Bird", 2, 2, "flying"}, {"Plant", 0, 8, ""}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		id := effectstest.B12Push(g, me.ID, "Skinshifter", "Creature — Human Shaman", oldB451Skin, 1, 1)
		effectstest.AdvanceToMainOf(t, g, g.Turn.ActiveSeat)
		g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
		effectstest.DcMana(t, g, me.ID, "{G}{G}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{i}}); err != nil {
			t.Fatalf("%s: activate: %v", tc.subtype, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.ApaLive(g, id)
		if p, tough := c.PowerForComparison(), c.CurrentToughness(); p != tc.power || tough != tc.tou {
			t.Errorf("%s: %d/%d, want %d/%d", tc.subtype, p, tough, tc.power, tc.tou)
		}
		if !c.HasSubtype(tc.subtype) || c.HasSubtype("Human") {
			t.Errorf("%s: subtypes %v, want only %s", tc.subtype, c.Effective().Subtypes, tc.subtype)
		}
		if tc.keyword != "" && !effectstest.HasEffectiveKeyword(t, g, id, tc.keyword) {
			t.Errorf("%s: no %s", tc.subtype, tc.keyword)
		}
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Modes: []int{i}}); err == nil {
			t.Errorf("%s: a second activation in one turn was allowed", tc.subtype)
		}
		effectstest.AdvancePastCleanupForTest(t, g)
		c = effectstest.ApaLive(g, id)
		if p := c.PowerForComparison(); p != 1 || !c.HasSubtype("Human") {
			t.Errorf("%s: after the turn it is power %d, Human %v", tc.subtype, p, c.HasSubtype("Human"))
		}
	}
}

func TestOldB451FlightGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Flight", oldB451Flight, bear)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the enchanted creature has no flying")
	}
}

func TestOldB451FirebreathingPumpsEachTimeYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Firebreathing", oldB451Fire, bear)
	for i := 1; i <= 2; i++ {
		if err := oldB451Activate(t, g, me, aura, "{R}"); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		if p := effectstest.EffectivePower(t, g, bear); p != 2+i {
			t.Fatalf("after %d activations power is %d, want %d", i, p, 2+i)
		}
	}
	if toughness := effectstest.EffectiveToughness(t, g, bear); toughness != 2 {
		t.Errorf("toughness %d, want 2", toughness)
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if p := effectstest.EffectivePower(t, g, bear); p != 2 {
		t.Errorf("power %d after the turn, want 2", p)
	}
}

func TestOldB451BrinkDestroysTheEnchantedCreatureOrLandWhenItTaps(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
	}{{"Bear", "Creature — Bear"}, {"Forest", oldB451ForestTyp}} {
		g := effectstest.NewCatalogGame(t)
		opp := g.Seats[1]
		host := effectstest.B12Permanent(g, opp.ID, tc.name, tc.typeLine)
		if tc.name == "Bear" {
			host = effectstest.B12Creature(g, opp.ID, tc.name, tc.typeLine, 2, 2)
		}
		effectstest.Enchant(t, g, "Brink of Disaster", oldB451Brink, host)
		if effectstest.ZoneOfCard(g, host) != game.ZoneBattlefield {
			t.Fatalf("%s: the host left on its own", tc.name)
		}
		effectstest.TapForTest(t, g, host)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.ZoneOfCard(g, host) != game.ZoneGraveyard {
			t.Errorf("%s: it is in zone %v after becoming tapped, want the graveyard", tc.name, effectstest.ZoneOfCard(g, host))
		}
	}
}

func TestOldB451BrinkRefusesAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.B12Permanent(g, g.Seats[1].ID, "Rock", "Artifact")
	if err := effectstest.CastCatalogSpellErr(t, g, "Brink of Disaster", effectstest.AuraTypeLine, oldB451Brink, []game.TargetRef{{Kind: game.TargetCard, ID: rock}}); err == nil {
		t.Fatal("enchanted a plain artifact")
	}
}

func TestOldB451CageLocksTheCreatureAndBreaksWhenItIsTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	cage := effectstest.Enchant(t, g, "Ice Cage", oldB451Cage, bear)
	r := effectstest.AuraRestrictions(t, g, bear)
	if r&game.CantAttackOrBlock != game.CantAttackOrBlock || r&game.CantActivate == 0 {
		t.Fatalf("restrictions %v, want no attack, block or activation", r)
	}
	if effectstest.ZoneOfCard(g, cage) != game.ZoneBattlefield {
		t.Fatal("the Aura is not on the battlefield")
	}
	effectstest.CastCatalogSpell(t, g, "Act of Treason", "Sorcery", oldB451Treason, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, cage) != game.ZoneGraveyard {
		t.Errorf("the Aura is in zone %v after the creature was targeted, want the graveyard", effectstest.ZoneOfCard(g, cage))
	}
	if got := effectstest.ControllerOf(t, g, bear); got != me.ID {
		t.Error("the spell that targeted the creature did not resolve")
	}
}

func TestOldB451PharaohRisesWhenCombatDamageHitsYouFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pharaoh := effectstest.PushGraveyardPermanent(me, "Vengeful Pharaoh", "Creature — Zombie", "{2}{B}{B}{B}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == pharaoh {
				me.Graveyard.Cards[i].OracleID = oldB451Pharaoh
			}
		}
	})
	attacker := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.AnswerTriggerTargets(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, attacker) != game.ZoneGraveyard {
		t.Errorf("the attacker is in zone %v, want the graveyard", effectstest.ZoneOfCard(g, attacker))
	}
	if effectstest.ZoneOfCard(g, other) != game.ZoneBattlefield {
		t.Error("a creature that was not attacking was destroyed")
	}
	if effectstest.ZoneOfCard(g, pharaoh) != game.ZoneLibrary || me.Library.Cards[len(me.Library.Cards)-1].InstanceID != pharaoh && me.Library.Cards[0].InstanceID != pharaoh {
		t.Errorf("the Pharaoh is in zone %v, want the top of the library", effectstest.ZoneOfCard(g, pharaoh))
	}
}

func TestOldB451ErasureMillsTheTargetPlayerOnYourDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Jace's Erasure", "Enchantment", oldB451Erasure, false)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Lib", TypeLine: "Instant"})
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	lib := opp.Library.Size()
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 1) })
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.B16PickPlayer(t, g, me.ID, opp.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-1 {
		t.Errorf("the target's library is %d, want %d", got, lib-1)
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(opp.ID, 1) })
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("Jace's Erasure triggered on an opponent's draw")
	}
}
