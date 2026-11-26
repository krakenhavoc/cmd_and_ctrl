package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b458_cards_test.go — New Phyrexia, damage, S82 older sets, batch 458
// (#5575). Each test plays the real catalog entry.

const (
	oldB458Membrane = "0297eeb2-47ec-4f9a-b1ad-4e7286994f00"
	oldB458Specter  = "18d270bf-d0d3-4fac-a8d2-5f920e789bb5"
	oldB458Betray   = "1d5ebb4b-0248-4255-8e1a-bcba4d5eaf7f"
	oldB458Reaper   = "251cee41-30f5-4bba-93cd-31493a5ea051"
	oldB458Mine     = "33510024-b1ff-428a-8ccd-2ca2a332d53c"
	oldB458Scamp    = "d0868a52-5b19-4429-acf1-399f3ae8308c"
	oldB458Rage     = "e45f0a12-e1ca-432f-a208-9359dd18513d"
	oldB458Arm      = "e6479195-cbdf-4d2b-9f45-5fd0a6ab4a77"
	oldB458Salvo    = "f63894a8-318f-4cfb-9a17-359fb9153e74"
	oldB458Bear     = "Creature — Bear"
)

func TestOldB458EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB458Membrane: "Cathedral Membrane",
		oldB458Specter:  "Whispering Specter",
		oldB458Betray:   "Viridian Betrayers",
		oldB458Reaper:   "Reaper of Sheoldred",
		oldB458Mine:     "Gremlin Mine",
		oldB458Scamp:    "Furnace Scamp",
		oldB458Rage:     "Rage Extractor",
		oldB458Arm:      "Arm with Aether",
		oldB458Salvo:    "Scrapyard Salvo",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB458Poison(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(uuid.Nil, p.ID, game.CounterPoison, n) })
}

func oldB458OnBattlefield(g *game.Game, id uuid.UUID) bool {
	_, ok := effectstest.BattlefieldCard(g, id)
	return ok
}

// --- Cathedral Membrane ------------------------------------------------------------

func TestOldB458CathedralMembraneDealsSixToTheAttackerItBlockedWhenItDiesInCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB458Bear, 3, 9)
	membrane := effectstest.B12Push(g, opp.ID, "Cathedral Membrane", "Artifact Creature — Phyrexian Wall", oldB458Membrane, 0, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{membrane: attacker})
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB458OnBattlefield(g, membrane) {
		t.Fatal("the Membrane survived 3 damage")
	}
	if got := effectstest.Pr6Marked(g, attacker); got < 6 {
		t.Errorf("the attacker has %d damage marked, want at least the 6 from the Membrane", got)
	}
}

func TestOldB458CathedralMembraneDoesNothingWhenItDiesOutsideCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", oldB458Bear, 3, 9)
	membrane := effectstest.B12Push(g, opp.ID, "Cathedral Membrane", "Artifact Creature — Phyrexian Wall", oldB458Membrane, 0, 3)
	effectstest.P1Destroy(t, g, membrane)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, bystander); got != 0 {
		t.Errorf("an unrelated creature has %d damage", got)
	}
}

// --- Whispering Specter and Furnace Scamp -------------------------------------------------

func TestOldB458WhisperingSpecterSacrificesToMakeThemDiscardPerPoisonCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB458Poison(g, opp, 2)
	specter := effectstest.B12Push(g, me.ID, "Whispering Specter", "Creature — Phyrexian Specter", oldB458Specter, 1, 1)
	effectstest.DealCombatDamageToPlayer(g, specter, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB458OnBattlefield(g, specter) {
		t.Fatal("the Specter was not sacrificed")
	}
	hand := len(opp.Hand.Cards)
	if hand < 2 {
		t.Fatalf("the opponent has only %d cards in hand to discard", hand)
	}
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := hand - len(opp.Hand.Cards); got != 2 {
		t.Errorf("the opponent discarded %d, want 2 (one per poison counter)", got)
	}
}

func TestOldB458WhisperingSpecterStaysWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB458Poison(g, opp, 2)
	specter := effectstest.B12Push(g, me.ID, "Whispering Specter", "Creature — Phyrexian Specter", oldB458Specter, 1, 1)
	hand := len(opp.Hand.Cards)
	effectstest.DealCombatDamageToPlayer(g, specter, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB458OnBattlefield(g, specter) || len(opp.Hand.Cards) != hand {
		t.Error("declining still sacrificed the Specter or made them discard")
	}
}

func TestOldB458FurnaceScampSacrificesForThreeDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	scamp := effectstest.B12Push(g, me.ID, "Furnace Scamp", "Creature — Phyrexian Beast", oldB458Scamp, 1, 1)
	life := opp.Life
	effectstest.DealCombatDamageToPlayer(g, scamp, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB458OnBattlefield(g, scamp) {
		t.Error("the Scamp was not sacrificed")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 (the combat damage here was an event only)", life, opp.Life)
	}
}

func TestOldB458FurnaceScampStaysWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	scamp := effectstest.B12Push(g, me.ID, "Furnace Scamp", "Creature — Phyrexian Beast", oldB458Scamp, 1, 1)
	life := opp.Life
	effectstest.DealCombatDamageToPlayer(g, scamp, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB458OnBattlefield(g, scamp) || opp.Life != life {
		t.Error("declining still sacrificed the Scamp or damaged the player")
	}
}

// --- Viridian Betrayers and Reaper of Sheoldred -----------------------------------------------

func TestOldB458ViridianBetrayersHasInfectOnlyWhileAnOpponentIsPoisoned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betrayers := effectstest.B12Push(g, me.ID, "Viridian Betrayers", "Creature — Phyrexian Elf Warrior", oldB458Betray, 3, 1)
	if c := effectstest.LayeredCard(t, g, betrayers); game.HasKeyword(&c, "infect") {
		t.Error("the Betrayers has infect with no poisoned opponent")
	}
	oldB458Poison(g, opp, 1)
	g.RunStateChecksForTest()
	if c := effectstest.LayeredCard(t, g, betrayers); !game.HasKeyword(&c, "infect") {
		t.Error("the Betrayers lacks infect while an opponent is poisoned")
	}
	oldB458Poison(g, me, 3)
	g.WithWriteLock(func() { _ = g.AddPlayerCounterByForEffect(uuid.Nil, opp.ID, game.CounterPoison, -1) })
	g.RunStateChecksForTest()
	if c := effectstest.LayeredCard(t, g, betrayers); game.HasKeyword(&c, "infect") {
		t.Error("the controller's own poison counters gave it infect")
	}
}

func TestOldB458ReaperOfSheoldredPoisonsTheControllerOfEachSourceThatDamagesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	reaper := effectstest.B12Push(g, me.ID, "Reaper of Sheoldred", "Creature — Phyrexian Horror", oldB458Reaper, 2, 5)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", oldB458Bear, 1, 1)
	effectstest.Pr6Damage(t, g, pinger, reaper, 1)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Pr6Damage(t, g, pinger, reaper, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Counters[game.CounterPoison]; got != 2 {
		t.Errorf("the source's controller has %d poison counters, want 2 (one per damage event)", got)
	}
	if got := me.Counters[game.CounterPoison]; got != 0 {
		t.Errorf("the Reaper's controller got %d poison counters", got)
	}
}

// --- Gremlin Mine -----------------------------------------------------------------------------------

func TestOldB458GremlinMineDealsFourToAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Gremlin Mine", "Artifact", oldB458Mine, 0, 0)
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Golem", TypeLine: "Artifact Creature — Golem",
		Power: 2, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.P7Activate(t, g, me, mine, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(golem)})
	if oldB458OnBattlefield(g, mine) {
		t.Error("the Mine was not sacrificed")
	}
	if oldB458OnBattlefield(g, golem) {
		t.Error("the artifact creature survived 4 damage")
	}
}

func TestOldB458GremlinMineCannotTargetANonartifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Gremlin Mine", "Artifact", oldB458Mine, 0, 0)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB458Bear, 2, 2)
	effectstest.FloatMana(t, g, me, "{C}")
	err := g.ActivateCatalogAbility(me.ID, mine, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTargets(bear)})
	if err == nil {
		t.Error("a nonartifact creature was accepted")
	}
}

func TestOldB458GremlinMineRemovesUpToFourChargeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Gremlin Mine", "Artifact", oldB458Mine, 0, 0)
	vial := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Charged Rock", TypeLine: "Artifact",
		Counters: map[string]int{game.CounterCharge: 6}, Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, mine, 1, game.ActivateAbilityParams{Targets: effectstest.CardTargets(vial)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.AnswerNumber(t, g, me.ID, 4)
	if prompt.Max != 4 {
		t.Errorf("the number prompt allows up to %d, want 4", prompt.Max)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := effectstest.BattlefieldCard(g, vial)
	if got := c.Counters[game.CounterCharge]; got != 2 {
		t.Errorf("%d charge counters left, want 2", got)
	}
}

// --- Rage Extractor ----------------------------------------------------------------------------------

func TestOldB458RageExtractorDealsTheManaValueOfAPhyrexianSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rage Extractor", "Artifact", oldB458Rage, 0, 0)
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}")
	effectstest.P1g2Cast(t, g, "Phyrexian Thing", "Sorcery", "", "{2}{R/P}", game.CastSpellParams{})
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d -> %d, want -3 (the spell's mana value)", life, opp.Life)
	}
}

func TestOldB458RageExtractorIgnoresASpellWithoutAPhyrexianSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Rage Extractor", "Artifact", oldB458Rage, 0, 0)
	life := opp.Life
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}")
	effectstest.P1g2Cast(t, g, "Plain Thing", "Sorcery", "", "{2}{R}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || opp.Life != life {
		t.Error("the Extractor triggered for a spell with no Phyrexian mana symbol")
	}
}

// --- Arm with Aether -----------------------------------------------------------------------------------

func TestOldB458ArmWithAetherBouncesACreatureWhenTheirCreatureDealsDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB458Bear, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB458Bear, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Arm with Aether", "Sorcery", oldB458Arm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DealCombatDamageToPlayer(g, attacker, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB458OnBattlefield(g, victim) || !opp.Hand.Contains(victim) {
		t.Error("the creature was not returned to its owner's hand")
	}
}

func TestOldB458ArmWithAetherOnlyCoversCreaturesYouControlWhenItResolves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB458Bear, 1, 1)
	effectstest.CastCatalogSpell(t, g, "Arm with Aether", "Sorcery", oldB458Arm, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.B12Creature(g, me.ID, "Late Arrival", oldB458Bear, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, late, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature that arrived after the spell resolved got the ability")
	}
	if !oldB458OnBattlefield(g, victim) {
		t.Error("their creature was bounced")
	}
}

// --- Scrapyard Salvo --------------------------------------------------------------------------------------

func TestOldB458ScrapyardSalvoCountsArtifactCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardMaterial(me, game.Card{Name: "Junk One", TypeLine: "Artifact"})
	effectstest.PushGraveyardMaterial(me, game.Card{Name: "Junk Two", TypeLine: "Artifact Creature — Golem"})
	effectstest.PushGraveyardMaterial(me, game.Card{Name: "Not Junk", TypeLine: "Sorcery"})
	effectstest.PushGraveyardMaterial(opp, game.Card{Name: "Their Junk", TypeLine: "Artifact"})
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Scrapyard Salvo", "Sorcery", oldB458Salvo, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d -> %d, want -2 (two artifact cards in the caster's graveyard)", life, opp.Life)
	}
}

func TestOldB458ScrapyardSalvoRefusesACreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB458Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Scrapyard Salvo", "Sorcery", oldB458Salvo, effectstest.CardTargets(bear)); err == nil {
		t.Error("a creature was accepted as a target")
	}
}
