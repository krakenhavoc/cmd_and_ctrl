package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b287_test.go — Shadows over Innistrad damage cards, S82 OLD batch
// 287 (#5543), played through the real catalog entries.

const (
	oldB287Wolf         = "453e72bf-a682-4ec8-9cdd-799f2acb3862"
	oldB287Playground   = "ab9e2045-860a-40ab-bec1-0dffced003ef"
	oldB287Forgemaster  = "b0d66546-c6a3-4052-b86e-98c3c2ab44b3"
	oldB287Machinations = "b119e0de-6af4-4b5b-9920-187f0a414231"
	oldB287Struggle     = "c993ebf8-972f-41c7-9870-d7f8aa9a69d9"
	oldB287Castigator   = "d46a6e0b-d39e-42ad-b95f-6902e1b72e62"
	oldB287Angel        = "f912c739-27a2-4068-bf0c-0891309639df"
)

var oldB287ForgemasterCard = effectstest.IwCard{Oracle: oldB287Forgemaster, Front: "Kessig Forgemaster", FrontType: "Creature — Human Shaman Werewolf", Cost: "{1}{R}",
	FrontText: "Whenever this creature blocks or becomes blocked by a creature, this creature deals 1 damage to that creature.\nAt the beginning of each upkeep, if no spells were cast last turn, transform this creature.",
	Fp:        "2", Ft: "1",
	Back: "Flameheart Werewolf", BackText: "Whenever this creature blocks or becomes blocked by a creature, this creature deals 2 damage to that creature.\nAt the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature.",
	Bp: "3", Bt: "2", Colors: []string{"R"}}

func TestOldB287EveryCardIsRegisteredFull(t *testing.T) {
	for id, name := range map[string]string{
		oldB287Wolf: "Wolf of Devil's Breach", oldB287Playground: "Devils' Playground",
		oldB287Forgemaster: "Kessig Forgemaster", oldB287Forgemaster + "#1": "Flameheart Werewolf",
		oldB287Machinations: "Nahiri's Machinations", oldB287Struggle: "Inner Struggle",
		oldB287Castigator: "Goldnight Castigator", oldB287Angel: "Flameblade Angel",
	} {
		spec, ok := effects.Lookup(id)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s registration: %+v", name, spec)
		}
	}
}

func TestOldB287DevilsPlaygroundMakesFourDevils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.CastCatalogSpell(t, g, "Devils' Playground", "Sorcery", oldB287Playground, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Devil"); got != 4 {
		t.Errorf("%d Devil tokens, want 4", got)
	}
}

func TestOldB287InnerStruggleHitsACreatureForItsOwnPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Giant", 3, 5)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Inner Struggle", "Instant", oldB287Struggle, nil); err == nil {
		t.Fatal("cast without a target")
	}
	effectstest.CastCatalogSpell(t, g, "Inner Struggle", "Instant", oldB287Struggle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, big); got != 3 {
		t.Errorf("the big creature has %d damage, want its power, 3", got)
	}
	effectstest.CastCatalogSpell(t, g, "Inner Struggle", "Instant", oldB287Struggle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/2 survived dealing itself 2 damage")
	}
}

func TestOldB287KessigForgemasterPokesBlockPairsAndFlipsBothWays(t *testing.T) {
	g, me := effectstest.IwTable(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	forge := effectstest.IwPut(t, g, oldB287ForgemasterCard, me)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 6)
	effectstest.WoeS2Block(g, me.ID, forge, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, foe); got != 1 {
		t.Errorf("the front face dealt %d to the blocked creature, want 1", got)
	}
	effectstest.IwTransform(t, g, forge)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 6)
	effectstest.WoeS2Block(g, opp.ID, other, forge)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, other); got != 2 {
		t.Errorf("the back face dealt %d to its blocker, want 2", got)
	}
}

func TestOldB287KessigForgemasterTransformsWhenNoSpellsWereCast(t *testing.T) {
	g, me := effectstest.IwTable(t)
	forge := effectstest.IwPut(t, g, oldB287ForgemasterCard, me)
	effectstest.IwNextUpkeep(t, g, map[int]int{})
	if effectstest.IwFace(g, forge) != 1 {
		t.Fatalf("Kessig Forgemaster is on face %d after a turn with no spells, want Flameheart Werewolf", effectstest.IwFace(g, forge))
	}
	effectstest.IwNextUpkeep(t, g, map[int]int{0: 2})
	if effectstest.IwFace(g, forge) != 0 {
		t.Errorf("Flameheart Werewolf is on face %d after a turn with two spells, want Kessig Forgemaster", effectstest.IwFace(g, forge))
	}
}

func TestOldB287NahirisMachinationsGivesIndestructibleAndShootsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mach := effectstest.PushCatalogPermanent(g, me.ID, "Nahiri's Machinations", "Enchantment", oldB287Machinations, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", 2, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PickTriggerTarget(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "indestructible") {
		t.Error("the chosen creature did not gain indestructible")
	}
	effectstest.CrtBlock(g, blocker, attacker, opp.ID)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, mach, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(attacker)}); err == nil {
		t.Fatal("a creature that is not blocking was accepted as the target")
	}
	if err := g.ActivateCatalogAbility(me.ID, mach, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.B16TargetCard(blocker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, blocker); got != 1 {
		t.Errorf("the blocker has %d damage, want 1", got)
	}
}

func TestOldB287GoldnightCastigatorDoublesDamageToYouAndToItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cast := effectstest.PushCatalogPermanent(g, me.ID, "Goldnight Castigator", "Creature — Angel", oldB287Castigator, false)
	shooter := effectstest.DcSeed(g, opp.ID, "Shooter", "Creature — Goblin", "", 1, 1)
	bystander := effectstest.B12Creature(g, me.ID, "Bystander", "Creature — Bear", 2, 9)
	life, oppLife := me.Life, opp.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(shooter, me.ID, 2)
		_ = g.DealDamageToPlayerForEffect(shooter, opp.ID, 2)
		_ = g.DealDamageToCreatureForEffect(shooter, cast, 2)
		_ = g.DealDamageToCreatureForEffect(shooter, bystander, 2)
	})
	if got := life - me.Life; got != 4 {
		t.Errorf("you lost %d from 2 damage, want 4", got)
	}
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d from 2 damage, want 2", got)
	}
	if got := effectstest.Pr6Marked(g, cast); got != 4 {
		t.Errorf("the Castigator has %d damage from 2, want 4", got)
	}
	if got := effectstest.Pr6Marked(g, bystander); got != 2 {
		t.Errorf("another creature has %d damage from 2, want 2", got)
	}
}

func TestOldB287FlamebladeAngelPunishesAnOpponentsSourceForHittingYouOrYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Flameblade Angel", "Creature — Angel", oldB287Angel, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 9)
	shooter := effectstest.DcSeed(g, opp.ID, "Shooter", "Creature — Goblin", "", 1, 1)
	own := effectstest.DcSeed(g, me.ID, "Own Shooter", "Creature — Goblin", "", 1, 1)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 9)
	oppLife := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(shooter, me.ID, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 1 {
		t.Errorf("after damage to you the opponent lost %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(shooter, mine, 1) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("after damage to your permanent the opponent lost %d in all, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(shooter, theirs, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLife - opp.Life; got != 2 {
		t.Errorf("damage to the opponent's own creature cost them %d in all, want 2", got)
	}
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(own, me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, me.ID) != nil {
		t.Error("your own source triggered the Angel")
	}
}

func TestOldB287WolfOfDevilsBreachPaysAndDiscardsForDamageEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wolf := effectstest.PushCatalogPermanent(g, me.ID, "Wolf of Devil's Breach", "Creature — Elemental Wolf", oldB287Wolf, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 9)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	pitch := effectstest.HandCardFull(me, "Pitch", "Sorcery", "{2}{R}", "", nil)
	effectstest.DeclareAttack(t, g, opp.ID, wolf)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt after the Wolf's attack trigger")
	}
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPool(me, 1)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no discard prompt after paying")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{pitch}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, victim); got != 3 {
		t.Errorf("the target has %d damage, want the discarded card's mana value, 3", got)
	}
}

func TestOldB287WolfOfDevilsBreachDoesNothingIfYouDeclineTheOffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	wolf := effectstest.PushCatalogPermanent(g, me.ID, "Wolf of Devil's Breach", "Creature — Elemental Wolf", oldB287Wolf, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 9)
	pitch := effectstest.HandCardFull(me, "Pitch", "Sorcery", "{2}{R}", "", nil)
	effectstest.DeclareAttack(t, g, opp.ID, wolf)
	effectstest.AnswerTriggerTargets(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.Pr6Marked(g, victim); got != 0 {
		t.Errorf("the target has %d damage after a declined offer", got)
	}
	if !me.Hand.Contains(pitch) {
		t.Error("a card was discarded without paying")
	}
}
