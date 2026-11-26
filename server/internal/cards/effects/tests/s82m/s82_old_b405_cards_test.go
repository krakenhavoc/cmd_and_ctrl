package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b405_cards_test.go — Dragon's Maze, batch 405 (S82 older sets,
// #5564): twelve cards of damage, combat, counters and Auras. Every test plays
// the real catalog entry through the engine.

const (
	oldB405Showstopper = "0441a901-31ce-419f-a898-e9a9ad4ab351"
	oldB405Punish      = "09f73f77-a9bc-4718-be7d-ba6e08576e3a"
	oldB405Mirko       = "469da932-8186-4e54-a858-cfc72c4ded96"
	oldB405Gorger      = "0e741c79-5ea4-457d-a7d3-b11154459246"
	oldB405Snare       = "2a174adc-0536-4e12-a6fe-c44305822714"
	oldB405Battleshape = "cf5bf1ef-e40b-4fb5-8148-d4ca7a307501"
	oldB405Gleam       = "19f23d9e-0c05-4028-9c64-60b9a20020f9"
	oldB405Prey        = "3836cf33-bd1d-485d-94a2-85ea210ddf7a"
	oldB405Hydra       = "fa8e5291-1a50-4c05-820c-8ae756364d33"
	oldB405Bane        = "31c30c4f-12f4-4cbc-b7c1-123f36b8df75"
	oldB405Incubation  = "66e2e816-5197-4151-a6c0-352c34c2eec7"
	oldB405Possession  = "6ee431f3-64ec-4015-8f5a-9c2ac6004ebd"

	oldB405Bear = "Creature — Bear"
)

var oldB405Cards = map[string]string{
	oldB405Showstopper: "Showstopper", oldB405Punish: "Punish the Enemy", oldB405Mirko: "Mirko Vosk, Mind Drinker",
	oldB405Gorger: "Species Gorger", oldB405Snare: "Haazda Snare Squad", oldB405Battleshape: "Boros Battleshaper",
	oldB405Gleam: "Gleam of Battle", oldB405Prey: "Mutant's Prey", oldB405Hydra: "Savageborn Hydra",
	oldB405Bane: "Runner's Bane", oldB405Incubation: "Krasis Incubation", oldB405Possession: "Sinister Possession",
}

func TestOldB405EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB405Cards {
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

func TestOldB405ShowstopperGivesTheDiesTriggerToCreaturesYouControlForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB405Bear, 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB405Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Showstopper", "Instant", oldB405Showstopper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, mine)
	effectstest.MscM2PickTargets(t, g, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's creature survived 2 damage from the dying creature")
	}
}

func TestOldB405ShowstopperDoesNotReachAnOpponentsCreatureOrALaterOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB405Bear, 2, 2)
	mineBefore := effectstest.B12Creature(g, me.ID, "Before", oldB405Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Showstopper", "Instant", oldB405Showstopper, nil)
	effectstest.PassPriorityAroundTable(t, g)
	later := effectstest.B12Creature(g, me.ID, "Later", oldB405Bear, 2, 2)
	effectstest.P1Destroy(t, g, theirs)
	effectstest.P1Destroy(t, g, later)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a creature that was not a creature you controlled as the spell resolved asked for a dies target")
	}
	if !g.Battlefield.Contains(mineBefore) {
		t.Error("the granted creature should be untouched")
	}
}

func TestOldB405PunishTheEnemyHitsAPlayerAndACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB405Bear, 3, 3)
	life := effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Punish the Enemy", "Instant", oldB405Punish,
		append(effectstest.PlayerRefs(opp.ID), effectstest.CardTarget(bear)...))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, opp.ID); got != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, got)
	}
	if g.Battlefield.Contains(bear) {
		t.Error("the 3/3 survived 3 damage")
	}
}

func TestOldB405PunishTheEnemyRefusesACreatureAsItsFirstTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB405Bear, 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB405Bear, 3, 3)
	card := effectstest.HandCardFull(g.Seats[g.Turn.ActiveSeat], "Punish the Enemy", "Instant", "", oldB405Punish, nil)
	effectstest.ToMainForCost(t, g)
	err := g.CastSpell(g.Seats[g.Turn.ActiveSeat].ID, card, game.CastSpellParams{Strict: true,
		Targets: append(effectstest.CardTarget(other), effectstest.CardTarget(bear)...)})
	if err == nil {
		t.Error("a creature was accepted as the player-or-planeswalker target")
	}
}

func TestOldB405MirkoVoskMillsUntilFourLandsFromTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mirko := effectstest.B12Push(g, me.ID, "Mirko Vosk, Mind Drinker", "Legendary Creature — Vampire", oldB405Mirko, 2, 4)
	// Top to bottom: spell, land, spell, land, land, spell, land, land, spell.
	order := []string{"Instant", "Basic Land — Forest", "Sorcery", "Basic Land — Forest", "Basic Land — Forest",
		"Sorcery", "Basic Land — Forest", "Basic Land — Forest", "Instant"}
	opp.Library.Cards = nil
	for i := len(order) - 1; i >= 0; i-- {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: order[i], Owner: opp.ID, Controller: opp.ID})
	}
	mine := me.Library.Size()
	effectstest.DealCombatDamage(g, me.ID, mirko, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Graveyard.Size(); got != 7 {
		t.Errorf("the damaged player's graveyard holds %d, want 7 (through the fourth land)", got)
	}
	if got := opp.Library.Size(); got != 2 {
		t.Errorf("the library holds %d, want the last 2 cards", got)
	}
	if me.Library.Size() != mine || me.Graveyard.Size() != 0 {
		t.Error("the controller's cards moved")
	}
}

func TestOldB405MirkoVoskMillsTheWholeLibraryShortOfFourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mirko := effectstest.B12Push(g, me.ID, "Mirko Vosk, Mind Drinker", "Legendary Creature — Vampire", oldB405Mirko, 2, 4)
	opp.Library.Cards = nil
	for _, tl := range []string{"Sorcery", "Basic Land — Forest", "Instant", "Basic Land — Forest"} {
		opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: tl, Owner: opp.ID, Controller: opp.ID})
	}
	effectstest.DealCombatDamage(g, me.ID, mirko, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != 0 || opp.Graveyard.Size() != 4 {
		t.Errorf("library %d, graveyard %d, want the whole library milled", opp.Library.Size(), opp.Graveyard.Size())
	}
}

func TestOldB405SpeciesGorgerReturnsACreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Species Gorger", "Creature — Frog Beast", oldB405Gorger, 6, 6)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB405Bear, 2, 2)
	effectstest.WalkToUpkeepOfSeat0(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("Species Gorger queued no return pick")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to hand")
	}
}

func oldB405SnareAttack(t *testing.T, g *game.Game, target uuid.UUID) *game.Player {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	squad := effectstest.B12Push(g, me.ID, "Haazda Snare Squad", "Creature — Human Soldier", oldB405Snare, 1, 4)
	effectstest.DeclareAttack(t, g, opp.ID, squad)
	effectstest.MscM2PickTargets(t, g, target)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the Squad did not ask whether to pay {W}")
	}
	return me
}

func TestOldB405HaazdaSnareSquadTapsTheTargetWhenYouPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", oldB405Bear, 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB405Bear, 2, 2)
	me := oldB405SnareAttack(t, g, bear)
	effectstest.FillPoolColored(me, "W", 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedNow(g, bear) {
		t.Error("the target was not tapped")
	}
	if effectstest.TappedNow(g, other) {
		t.Error("an untargeted creature was tapped")
	}
}

func TestOldB405HaazdaSnareSquadDoesNothingWhenYouDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", oldB405Bear, 2, 2)
	me := oldB405SnareAttack(t, g, bear)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, bear) {
		t.Error("the target was tapped although nothing was paid")
	}
}

func TestOldB405BorosBattleshaperAimsBothClauses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Boros Battleshaper", "Creature — Minotaur Soldier", oldB405Battleshape, 5, 5)
	forced := effectstest.B12Creature(g, opp.ID, "Forced", oldB405Bear, 2, 2)
	barred := effectstest.B12Creature(g, opp.ID, "Barred", oldB405Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	for _, id := range []uuid.UUID{forced, barred} {
		ch := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTargets(ch.ID, ch.Chooser, effectstest.CardTarget(id)); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RestrictionsOf(t, g, barred); got&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("the second target's restrictions = %v, want can't attack or block", got.Names())
	}
	if got := effectstest.RestrictionsOf(t, g, forced); got&(game.CantAttack|game.CantBlock) != 0 {
		t.Errorf("the first target is restricted: %v", got.Names())
	}
}

func TestOldB405GleamOfBattleCountersEachAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gleam of Battle", "Enchantment", oldB405Gleam, false)
	a := effectstest.B12Creature(g, me.ID, "A", oldB405Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB405Bear, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", oldB405Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB405Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b} {
		if got := effectstest.CounterOn(g, id, game.CounterPlusOne); got != 1 {
			t.Errorf("attacker %s has %d +1/+1 counters, want 1", name, got)
		}
	}
	if effectstest.CounterOn(g, idle, game.CounterPlusOne) != 0 || effectstest.CounterOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("a creature that did not attack got a counter")
	}
}

func TestOldB405MutantsPreyFightsOnlyFromACreatureWithAPlusOneCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	plain := effectstest.B12Creature(g, me.ID, "Plain", oldB405Bear, 3, 3)
	grown := effectstest.B12Creature(g, me.ID, "Grown", oldB405Bear, 2, 2)
	effectstest.PutCounters(t, g, grown, game.CounterPlusOne, 1)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB405Bear, 3, 3)
	spell := effectstest.HandCardFull(me, "Mutant's Prey", "Instant", "", oldB405Prey, nil)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true,
		Targets: append(effectstest.CardTarget(plain), effectstest.CardTarget(victim)...)}); err == nil {
		t.Fatal("a creature with no +1/+1 counter was accepted as the fighter")
	}
	effectstest.CastCatalogSpell(t, g, "Mutant's Prey", "Instant", oldB405Prey,
		append(effectstest.CardTarget(grown), effectstest.CardTarget(victim)...))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the 3/3 (counter-bearing fighter has power 3) survived the fight")
	}
}

func TestOldB405SavagebornHydraEntersWithXCountersAndGrowsAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	hydra := effectstest.B12Push(g, me.ID, "Savageborn Hydra", "Creature — Hydra", oldB405Hydra, 0, 0)
	effectstest.PutCounters(t, g, hydra, game.CounterPlusOne, 2)
	if !effectstest.HasEffectiveKeyword(t, g, hydra, "double strike") {
		t.Error("the Hydra has no double strike")
	}
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.ActivateCatalogAbility(me.ID, hydra, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate at sorcery speed: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, hydra, game.CounterPlusOne); got != 3 {
		t.Errorf("counters = %d, want 3", got)
	}
}

func TestOldB405RunnersBaneTapsAndHoldsAPowerThreeOrLessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	small := effectstest.B12Creature(g, opp.ID, "Small", oldB405Bear, 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", oldB405Bear, 4, 4)
	me := g.Seats[g.Turn.ActiveSeat]
	card := effectstest.HandCardFull(me, "Runner's Bane", "Enchantment — Aura", "", oldB405Bane, nil)
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(big)}); err == nil {
		t.Fatal("the Aura was cast on a creature with power 4")
	}
	effectstest.AuraCast(t, g, "Runner's Bane", oldB405Bane, small)
	if !effectstest.TappedNow(g, small) {
		t.Error("the enchanted creature was not tapped as the Aura entered")
	}
}

func TestOldB405KrasisIncubationLocksTheCreatureAndReturnsForCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB405Bear, 2, 2)
	aura := effectstest.AuraCast(t, g, "Krasis Incubation", oldB405Incubation, bear)
	if got := effectstest.RestrictionsOf(t, g, bear); got&(game.CantAttackOrBlock|game.CantActivate) != game.CantAttackOrBlock|game.CantActivate {
		t.Errorf("restrictions = %v, want can't attack, block or activate", got.Names())
	}
	effectstest.FloatMana(t, g, me, "{G}{U}{G}")
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(aura) {
		t.Error("the Aura did not return to its owner's hand")
	}
	if got := effectstest.CounterOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the creature has %d +1/+1 counters, want 2", got)
	}
	if got := effectstest.RestrictionsOf(t, g, bear); got != 0 {
		t.Errorf("restrictions %v remain after the Aura left", got.Names())
	}
}

func TestOldB405SinisterPossessionDrainsTheEnchantedCreaturesControllerOnAttackAndBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.SeedBear(g, me.ID)
	theirs := effectstest.SeedBear(g, opp.ID)
	effectstest.AuraCast(t, g, "Sinister Possession", oldB405Possession, mine)
	effectstest.AuraCast(t, g, "Sinister Possession", oldB405Possession, theirs)
	myLife, theirLife := me.Life, opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-2 {
		t.Errorf("attacking: my life %d, want %d", me.Life, myLife-2)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(theirs, mine); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-2 {
		t.Errorf("blocking: the blocker's controller at %d, want %d", opp.Life, theirLife-2)
	}
	if me.Life != myLife-2 {
		t.Errorf("my life %d, want %d: only the block triggered the second time", me.Life, myLife-2)
	}
}
