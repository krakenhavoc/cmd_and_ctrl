package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b321_cards_test.go — Magic Origins, batch 321 (S82 older sets,
// #5549). Every test plays the real catalog entry through the engine.

const (
	oldB321TaintedRemedy   = "51b114cd-c174-4d74-a854-07c68b41fc9d"
	oldB321ShadowsPast     = "a5ba544f-b223-47c8-a408-b00512772107"
	oldB321ThornbowArcher  = "a9b35457-0f1c-495c-a508-6f6693835051"
	oldB321Moonglove       = "d3a6bff6-ea85-4e8a-b256-7427d15326b6"
	oldB321BloodCursed     = "fdf51294-6270-4c94-a2a9-8df37dfcaa0d"
	oldB321DisplacementWav = "2b3fc609-7b19-441a-ab7e-58ccd70d955b"
	oldB321InfiniteObl     = "3bea62dc-14cd-4cbc-b55d-3791c444bdc7"
	oldB321Dreadwaters     = "e6e70eb3-3da5-4052-9f62-2960a30d4b41"
	oldB321CelestialFlare  = "efdaa3eb-e5d1-4438-b21d-2ec0fff44721"
	oldB321ErebosTitan     = "1341b64e-1f8f-4613-abea-a337cd0169de"
	oldB321PossessedSkaab  = "a42b567d-6bfc-49e7-8f91-917e2bb3046c"
	oldB321Despoiler       = "cabb0cf3-124a-4f5f-9ca7-2d45a72159ca"
	oldB321Bear            = "Creature — Bear"
)

func TestOldB321EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		oldB321TaintedRemedy, oldB321ShadowsPast, oldB321ThornbowArcher, oldB321Moonglove, oldB321BloodCursed,
		oldB321DisplacementWav, oldB321InfiniteObl, oldB321Dreadwaters, oldB321CelestialFlare, oldB321ErebosTitan,
		oldB321PossessedSkaab, oldB321Despoiler,
	} {
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

func TestOldB321TaintedRemedyTurnsAnOpponentsGainIntoLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Tainted Remedy", "Enchantment", oldB321TaintedRemedy, false)
	mine, theirs := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, 4)
		_ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 3)
	})
	if opp.Life != theirs-4 {
		t.Errorf("the opponent went %d → %d, want a loss of 4", theirs, opp.Life)
	}
	if me.Life != mine+3 {
		t.Errorf("its controller went %d → %d, want a gain of 3", mine, me.Life)
	}
}

func TestOldB321BloodCursedKnightNeedsAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	knight := effectstest.B12Push(g, me.ID, "Blood-Cursed Knight", "Creature — Vampire Knight", oldB321BloodCursed, 3, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 3 || tough != 2 {
		t.Errorf("alone it is %d/%d, want 3/2", p, tough)
	}
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, knight), "lifelink") {
		t.Error("it has lifelink with no enchantment")
	}
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Some Enchantment", "Enchantment", "", false)
	if p, tough := effectstest.EffectivePTOf(t, g, knight); p != 4 || tough != 3 {
		t.Errorf("with an enchantment it is %d/%d, want 4/3", p, tough)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, knight), "lifelink") {
		t.Error("it lacks lifelink with an enchantment")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(aura) })
	if p, _ := effectstest.EffectivePTOf(t, g, knight); p != 3 {
		t.Errorf("after the enchantment left it is %d power, want 3", p)
	}
}

func TestOldB321BloodCursedKnightIgnoresAnOpponentsEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	knight := effectstest.B12Push(g, me.ID, "Blood-Cursed Knight", "Creature — Vampire Knight", oldB321BloodCursed, 3, 2)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Enchantment", "Enchantment", "", false)
	if p, _ := effectstest.EffectivePTOf(t, g, knight); p != 3 {
		t.Errorf("an opponent's enchantment made it %d power, want 3", p)
	}
}

func TestOldB321DisplacementWaveBouncesByManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	push := func(owner uuid.UUID, name, typeLine, cost string) uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: cost,
			Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	}
	cheap := push(me.ID, "Cheap", oldB321Bear, "{1}")
	edge := push(opp.ID, "Edge", oldB321Bear, "{1}{G}")
	pricey := push(opp.ID, "Pricey", oldB321Bear, "{2}{G}")
	land := push(opp.ID, "Island", "Basic Land — Island", "")
	free := push(me.ID, "Free", "Artifact", "")
	effectstest.CastXSpell(t, g, "Displacement Wave", "Sorcery", oldB321DisplacementWav, "{X}{U}{U}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"cheap": cheap, "edge": edge, "free": free} {
		if effectstest.ZoneOf(g, id) != game.ZoneHand {
			t.Errorf("the %s permanent is in %v, want a hand", name, effectstest.ZoneOf(g, id))
		}
	}
	if effectstest.ZoneOf(g, pricey) != game.ZoneBattlefield {
		t.Error("a mana value 3 permanent was returned by X=2")
	}
	if effectstest.ZoneOf(g, land) != game.ZoneBattlefield {
		t.Error("a land was returned")
	}
}

func TestOldB321DisplacementWaveWithXZeroStillReturnsFreePermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	free := effectstest.B12Creature(g, me.ID, "Free", oldB321Bear, 1, 1)
	effectstest.CastXSpell(t, g, "Displacement Wave", "Sorcery", oldB321DisplacementWav, "{X}{U}{U}", 0, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, free) != game.ZoneHand {
		t.Errorf("a mana value 0 creature is in %v, want a hand", effectstest.ZoneOf(g, free))
	}
}

func TestOldB321DreadwatersMillsByYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Island", "Basic Land — Island", "", false)
	before, mine := opp.Library.Size(), me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Dreadwaters", "Sorcery", oldB321Dreadwaters, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("the target milled %d, want the caster's 3 lands", got)
	}
	if me.Library.Size() != mine {
		t.Error("the caster milled")
	}
}

func TestOldB321ShadowsOfThePastScriesWhenAnyCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.PushCatalogPermanent(g, me.ID, "Shadows of the Past", "Enchantment", oldB321ShadowsPast, false)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB321Bear, 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Fatal("an opponent's creature dying did not scry its controller 1")
	}
}

func TestOldB321ShadowsOfThePastDrainsOnlyWithFourCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	shadows := effectstest.PushCatalogPermanent(g, me.ID, "Shadows of the Past", "Enchantment", oldB321ShadowsPast, false)
	for i := 0; i < 3; i++ {
		effectstest.PushGraveyardCardTyped(me, "Dead Bear", oldB321Bear)
	}
	effectstest.PushGraveyardCardTyped(me, "Dead Spell", "Sorcery")
	effectstest.FloatForTest(g, me, "BBBBB")
	if err := g.ActivateCatalogAbility(me.ID, shadows, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with three creature cards in the graveyard")
	}
	effectstest.PushGraveyardCardTyped(me, "Dead Bear", oldB321Bear)
	mine, theirs := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, shadows, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate with four creature cards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirs-2 || me.Life != mine+2 {
		t.Errorf("life %d → %d and %d → %d, want -2 and +2", theirs, opp.Life, mine, me.Life)
	}
	for _, other := range g.Seats[2:] {
		if other.Life != 40 && other.Life != theirs-2 {
			t.Errorf("another opponent is at %d", other.Life)
		}
	}
}

func TestOldB321ThornbowArcherDrainsOpponentsWithoutAnElf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	archer := effectstest.B12Push(g, me.ID, "Thornbow Archer", "Creature — Elf Archer", oldB321ThornbowArcher, 1, 2)
	elfOwner := g.Seats[2]
	effectstest.B12Creature(g, elfOwner.ID, "Their Elf", "Creature — Elf Warrior", 1, 1)
	plain := g.Seats[3]
	lifeOpp, lifeElf, lifePlain := opp.Life, elfOwner.Life, plain.Life
	effectstest.AttackWithTrigger(t, g, archer, opp.ID, uuid.Nil)
	if opp.Life != lifeOpp-1 {
		t.Errorf("an opponent with no Elf is at %d, want %d", opp.Life, lifeOpp-1)
	}
	if elfOwner.Life != lifeElf {
		t.Errorf("an opponent with an Elf lost life: %d, want %d", elfOwner.Life, lifeElf)
	}
	if plain.Life != lifePlain-1 {
		t.Errorf("another opponent with no Elf is at %d, want %d", plain.Life, lifePlain-1)
	}
}

func TestOldB321TouchOfMoongloveShrinksWhatTheTargetDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", oldB321Bear, 2, 2)
	hit := effectstest.B12Creature(g, opp.ID, "Hit Bear", oldB321Bear, 5, 5)
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", oldB321Bear, 5, 5)
	effectstest.CastCatalogSpell(t, g, "Touch of Moonglove", "Instant", oldB321Moonglove, []game.TargetRef{{Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 3 {
		t.Errorf("the target has %d power, want 3", p)
	}
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, mine), "deathtouch") {
		t.Error("the target lacks deathtouch")
	}
	effectstest.DealDamageTo(t, g, mine, hit, 1)
	life := opp.Life
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(other)
		_ = g.DestroyPermanentForEffect(hit)
	})
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the controller of the damaged creature is at %d, want %d (one creature was damaged, one was not)", opp.Life, life-2)
	}
}

func TestOldB321CelestialFlareMakesTheTargetSacrificeAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacking Bear", oldB321Bear, 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", oldB321Bear, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.CastFromHandForTest(t, g, me, "Celestial Flare", "Instant", "", oldB321CelestialFlare,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}})
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.SacrificeChoiceFor(g, me.ID)
	if choice == nil {
		if effectstest.ZoneOf(g, attacker) == game.ZoneBattlefield {
			t.Fatal("no sacrifice was asked and the attacker stayed")
		}
	} else {
		options := effectstest.SacrificeOptions(choice)
		if len(options) != 1 || options[0] != attacker {
			t.Fatalf("sacrifice options %v, want only the attacking creature", options)
		}
		effectstest.AnswerSacrifice(t, g, me.ID, attacker)
	}
	if effectstest.ZoneOf(g, attacker) == game.ZoneBattlefield {
		t.Error("the attacking creature was not sacrificed")
	}
	if effectstest.ZoneOf(g, idle) != game.ZoneBattlefield {
		t.Error("a creature that was not attacking was sacrificed")
	}
}

func TestOldB321ErebosTitanIsIndestructibleWhileOpponentsHaveNoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	titan := effectstest.B12Push(g, me.ID, "Erebos's Titan", "Creature — Giant", oldB321ErebosTitan, 5, 5)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, titan), "indestructible") {
		t.Error("no indestructible while the opponents control no creatures")
	}
	effectstest.B12Creature(g, me.ID, "My Bear", oldB321Bear, 1, 1)
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, titan), "indestructible") {
		t.Error("my own creature turned the ability off")
	}
	foe := effectstest.B12Creature(g, opp.ID, "Their Bear", oldB321Bear, 1, 1)
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, titan), "indestructible") {
		t.Error("indestructible while an opponent controls a creature")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(foe) })
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, titan), "indestructible") {
		t.Error("indestructible did not come back")
	}
}

// oldB321TitanInGraveyard puts the Titan in my graveyard, a creature card in an
// opponent's graveyard, and a card in my hand to discard.
func oldB321TitanInGraveyard(g *game.Game) (me, opp *game.Player, titan, theirCard, spare uuid.UUID) {
	me, opp = g.Seats[0], g.Seats[1]
	titan = effectstest.PushCatalogGraveyardCard(me, "Erebos's Titan", "Creature — Giant", oldB321ErebosTitan, 5, 5)
	theirCard = effectstest.PushGraveyardCardTyped(opp, "Their Dead Bear", oldB321Bear)
	spare = effectstest.PushHandCardOfType(me, "Spare", "Sorcery")
	return
}

func oldB321ExileFromGraveyard(t *testing.T, g *game.Game, owner *game.Player, id uuid.UUID) {
	t.Helper()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatalf("exile the graveyard card: %v", err)
	}
	g.RunStateChecksForTest()
}

func TestOldB321ErebosTitanReturnsFromTheGraveyardForADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, titan, theirCard, spare := oldB321TitanInGraveyard(g)
	oldB321ExileFromGraveyard(t, g, opp, theirCard)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, spare)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, titan) != game.ZoneHand {
		t.Errorf("the Titan is in %v, want my hand", effectstest.ZoneOf(g, titan))
	}
	if !me.Graveyard.Contains(spare) {
		t.Error("the spare card was not discarded")
	}
}

func TestOldB321ErebosTitanStaysWhenYouDeclineTheDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, titan, theirCard, spare := oldB321TitanInGraveyard(g)
	oldB321ExileFromGraveyard(t, g, opp, theirCard)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, titan) != game.ZoneGraveyard {
		t.Errorf("the Titan is in %v, want the graveyard", effectstest.ZoneOf(g, titan))
	}
	if !me.Hand.Contains(spare) {
		t.Error("the spare card was discarded anyway")
	}
}

func TestOldB321ErebosTitanIgnoresYourOwnGraveyardAndNonCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, titan, theirCard, _ := oldB321TitanInGraveyard(g)
	mine := effectstest.PushGraveyardCardTyped(me, "My Dead Bear", oldB321Bear)
	spell := effectstest.PushGraveyardCardTyped(opp, "Their Dead Spell", "Sorcery")
	oldB321ExileFromGraveyard(t, g, me, mine)
	oldB321ExileFromGraveyard(t, g, opp, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a card leaving my graveyard, or a noncreature card, asked for a discard")
	}
	if effectstest.ZoneOf(g, titan) != game.ZoneGraveyard || !opp.Graveyard.Contains(theirCard) {
		t.Error("something moved")
	}
}

func TestOldB321PossessedSkaabReturnsACardAndExilesWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	bolt := effectstest.PushGraveyardCardTyped(me, "Old Bolt", "Instant")
	land := effectstest.PushGraveyardCardTyped(me, "Old Land", "Basic Land — Island")
	skaab := effectstest.RfdCast(t, g, "Possessed Skaab", "Creature — Zombie", oldB321PossessedSkaab, 3, 2)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bolt) {
		t.Error("the instant did not return to hand")
	}
	if !me.Graveyard.Contains(land) {
		t.Error("the land card left the graveyard")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(skaab) })
	g.RunStateChecksForTest()
	if effectstest.ZoneOf(g, skaab) != game.ZoneExile {
		t.Errorf("the Skaab is in %v, want exile", effectstest.ZoneOf(g, skaab))
	}
}

func TestOldB321PossessedSkaabCannotTargetALandCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	land := effectstest.PushGraveyardCardTyped(me, "Old Land", "Basic Land — Island")
	effectstest.RfdCast(t, g, "Possessed Skaab", "Creature — Zombie", oldB321PossessedSkaab, 3, 2)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: land}); err == nil {
			t.Fatal("a land card was accepted as the target")
		}
	}
}

func TestOldB321DespoilerCantBlockAndReturnsByExilingTwoOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fodderA := effectstest.PushGraveyardCardTyped(me, "Fodder A", oldB321Bear)
	fodderB := effectstest.PushGraveyardCardTyped(me, "Fodder B", oldB321Bear)
	spell := effectstest.PushGraveyardCardTyped(me, "A Spell", "Sorcery")
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Despoiler of Souls", "Creature — Horror", oldB321Despoiler, 3, 1, "{B}{B}",
		game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{fodderA, fodderB}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOf(g, id) != game.ZoneBattlefield {
		t.Fatalf("the Despoiler is in %v, want the battlefield", effectstest.ZoneOf(g, id))
	}
	if !g.Exile.Contains(fodderA) || !g.Exile.Contains(fodderB) {
		t.Error("the two creature cards were not exiled")
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("the noncreature card left the graveyard")
	}
}

func TestOldB321DespoilerNeedsTwoOtherCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	fodder := effectstest.PushGraveyardCardTyped(me, "Fodder", oldB321Bear)
	id := effectstest.PushCatalogGraveyardCard(me, "Despoiler of Souls", "Creature — Horror", oldB321Despoiler, 3, 1)
	effectstest.FloatForTest(g, me, "BB")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{fodder, id}}); err == nil {
		t.Fatal("activated by exiling itself as the second card")
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("activated with one other creature card")
	}
}

func TestOldB321DespoilerCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	d := effectstest.B12Push(g, opp.ID, "Despoiler of Souls", "Creature — Horror", oldB321Despoiler, 3, 1)
	if !effectstest.MidB03CantBlock(t, g, d) {
		t.Error("the Despoiler can block")
	}
	_ = me
}

func TestOldB321InfiniteObliterationExilesTheNamedCreatureCardFromEveryZone(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Grizzly Bears", "Creature — Bear")
	effectstest.CastCatalogSpell(t, x.G, "Infinite Obliteration", "Sorcery", oldB321InfiniteObl, []game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "grizzly bears")
	effectstest.AnswerSearchByID(t, x.G, x.Me.ID, x.Cards["hand"], x.Cards["lib"], x.Cards["gy"], x.Cards["gy 2"])
	effectstest.PassPriorityAroundTable(t, x.G)
	x.Exiled(t, "gy", "gy 2", "hand", "lib")
	x.NotExiled(t, "gy other", "hand other", "lib other", "gy forest")
	if x.Searches() != 1 {
		t.Errorf("the library was searched and shuffled %d times, want 1", x.Searches())
	}
}

func TestOldB321InfiniteObliterationFindsNothingForANonCreatureName(t *testing.T) {
	x := effectstest.ExtractionSetUp(t, "Lightning Bolt", "Instant")
	effectstest.CastCatalogSpell(t, x.G, "Infinite Obliteration", "Sorcery", oldB321InfiniteObl, []game.TargetRef{{Kind: game.TargetPlayer, ID: x.Opp.ID}})
	effectstest.PassPriorityAroundTable(t, x.G)
	effectstest.ActNameFor(t, x.G, x.Me.ID, "Lightning Bolt")
	if c := effectstest.SearchPromptFor(x.G, x.Me.ID); c != nil {
		if len(c.SearchCards) != 0 {
			t.Errorf("a search offered %d candidates for a name no creature card has", len(c.SearchCards))
		}
	}
	x.NotExiled(t, "gy", "gy 2", "hand", "lib")
}
