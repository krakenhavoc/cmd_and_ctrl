package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB396Hythonia = "48e3d832-2fe9-4ffd-8258-75451100dcfc"
	oldB396Nemesis  = "599fc217-e0a3-45ee-b9aa-cd7f181dcdf3"
	oldB396Colossus = "96d508a6-ecd2-4a97-8802-df099a372084"
	oldB396Crusher  = "ffd6f55f-2592-46a5-94da-cb9424783004"
	oldB396Dissolve = "2f0741f9-802b-4761-8094-7b1f3ea1a424"
	oldB396Gainsay  = "5115eb6b-efae-4561-912a-f7909cc75775"
	oldB396Stymied  = "b081726f-0ed6-4ab0-833e-f510ac6fbeaa"
	oldB396Hydra    = "30b5b764-478a-46fa-8f95-a1f515925a2d"
	oldB396Hunter   = "7b4502bc-9ece-4fd2-978d-bcfdb9358475"
	oldB396Triad    = "fd09875f-71a1-41c2-b404-ff58f4d4cb6c"
	oldB396Lesson   = "1d4605c0-47c5-4561-ab53-9cff677330aa"
	oldB396Bounty   = "c43425c1-ad01-4088-a061-f691fa2f6c20"
	oldB396Bear     = "Creature — Bear"
)

var oldB396Cards = map[string]string{
	oldB396Hythonia: "Hythonia the Cruel", oldB396Nemesis: "Nemesis of Mortals", oldB396Colossus: "Colossus of Akros",
	oldB396Crusher: "Polis Crusher", oldB396Dissolve: "Dissolve", oldB396Gainsay: "Gainsay", oldB396Stymied: "Stymied Hopes",
	oldB396Hydra: "Mistcutter Hydra", oldB396Hunter: "Reverent Hunter", oldB396Triad: "Triad of Fates",
	oldB396Lesson: "Warriors' Lesson", oldB396Bounty: "Thassa's Bounty",
}

func TestOldB396AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range oldB396Cards {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB396Activate(t *testing.T, g *game.Game, me *game.Player, id uuid.UUID, idx int, params game.ActivateAbilityParams) {
	t.Helper()
	params.Strict = true
	if err := g.ActivateCatalogAbility(me.ID, id, idx, params); err != nil {
		t.Fatalf("activate %d: %v", idx, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB396HythoniaMonstrosityDestroysEveryNonGorgon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hythonia := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hythonia the Cruel", TypeLine: "Legendary Creature — Gorgon", OracleID: oldB396Hythonia,
		Power: 4, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	gorgon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Other Gorgon", TypeLine: "Creature — Gorgon", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB396Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB396Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 8)
	oldB396Activate(t, g, me, hythonia, 0, game.ActivateAbilityParams{})
	if !g.Battlefield.Contains(hythonia) || !g.Battlefield.Contains(gorgon) {
		t.Error("a Gorgon was destroyed")
	}
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a non-Gorgon survived")
	}
	if p, tough := effectstest.PtOf(t, g, hythonia); p != 7 || tough != 9 {
		t.Errorf("Hythonia is %d/%d, want 7/9 with three counters", p, tough)
	}
}

func TestOldB396NemesisOfMortalsCostsLessPerCreatureCardInYourGraveyard(t *testing.T) {
	cast := func(inGraveyard, poolG int) error {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		for i := 0; i < inGraveyard; i++ {
			effectstest.PushGraveyardPermanent(me, "Dead", "Creature — Bear", "{1}")
		}
		effectstest.AdvanceToMain(t, g)
		id := effectstest.HandCardOf(g, me, "Nemesis of Mortals", "Creature — Snake", "{4}{G}{G}", oldB396Nemesis)
		effectstest.FillPoolColored(me, "G", poolG)
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
	}
	if err := cast(2, 3); err == nil {
		t.Error("cast for 3 with two creature cards in the graveyard (cost 4)")
	}
	if err := cast(2, 4); err != nil {
		t.Errorf("cast for 4 with two creature cards in the graveyard: %v", err)
	}
	if err := cast(0, 5); err == nil {
		t.Error("cast for 5 with no creature card in the graveyard (cost 6)")
	}
}

func TestOldB396NemesisOfMortalsMonstrosityCostsLessToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	nemesis := effectstest.B12Push(g, me.ID, "Nemesis of Mortals", "Creature — Snake", oldB396Nemesis, 5, 5)
	effectstest.PushGraveyardPermanent(me, "Dead A", "Creature — Bear", "{1}")
	effectstest.PushGraveyardPermanent(me, "Dead B", "Creature — Bear", "{1}")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 6)
	if err := g.ActivateCatalogAbility(me.ID, nemesis, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for 6 with two creature cards in the graveyard (cost 7)")
	}
	effectstest.FillPoolColored(me, "G", 1)
	oldB396Activate(t, g, me, nemesis, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, nemesis); p != 10 || tough != 10 {
		t.Errorf("the Nemesis is %d/%d, want 10/10", p, tough)
	}
}

func TestOldB396ColossusOfAkrosAttacksOnlyOnceMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.B12Push(g, me.ID, "Colossus of Akros", "Artifact Creature — Golem", oldB396Colossus, 10, 10)
	if !effectstest.HasEffectiveKeyword(t, g, colossus, "defender") || !effectstest.HasEffectiveKeyword(t, g, colossus, "indestructible") {
		t.Fatal("no defender and indestructible")
	}
	if effectstest.HasEffectiveKeyword(t, g, colossus, "trample") {
		t.Fatal("trample before it is monstrous")
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 10)
	oldB396Activate(t, g, me, colossus, 0, game.ActivateAbilityParams{})
	if p, tough := effectstest.PtOf(t, g, colossus); p != 20 || tough != 20 {
		t.Errorf("the Colossus is %d/%d, want 20/20", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, colossus, "trample") || !effectstest.HasEffectiveKeyword(t, g, colossus, "defender") {
		t.Error("monstrous: it should have trample and keep defender")
	}
	effectstest.DeclareAttack(t, g, opp.ID, colossus)
	if c, ok := effectstest.BattlefieldCard(g, colossus); !ok || c.AttackingTarget != opp.ID {
		t.Error("the monstrous Colossus did not attack")
	}
}

func TestOldB396ColossusOfAkrosCantAttackBeforeItIsMonstrous(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	colossus := effectstest.B12Push(g, me.ID, "Colossus of Akros", "Artifact Creature — Golem", oldB396Colossus, 10, 10)
	_ = me
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(colossus, opp.ID); err == nil {
		t.Fatal("a defender that is not monstrous attacked")
	}
}

func TestOldB396PolisCrusherDestroysAnEnchantmentOnlyWhileMonstrous(t *testing.T) {
	run := func(monstrous bool) (destroyed bool) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		crusher := effectstest.B12Push(g, me.ID, "Polis Crusher", "Creature — Cyclops", oldB396Crusher, 4, 4)
		mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Enchantment", TypeLine: "Enchantment", Owner: me.ID, Controller: me.ID})
		theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Enchantment", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
		if monstrous {
			effectstest.AdvanceToMain(t, g)
			effectstest.FillPoolColored(me, "R", 5)
			effectstest.FillPoolColored(me, "G", 1)
			oldB396Activate(t, g, me, crusher, 0, game.ActivateAbilityParams{})
		}
		effectstest.AttackWith(t, g, opp.ID, crusher)
		effectstest.InitSettle(t, g)
		if monstrous {
			p := effectstest.LatestPickTarget(g, me.ID)
			if p == nil {
				t.Fatal("the monstrous Crusher's trigger asks for no target")
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
				t.Fatal("targeted my own enchantment")
			}
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err != nil {
				t.Fatalf("target: %v", err)
			}
		} else if effectstest.LatestPickTarget(g, me.ID) != nil {
			t.Fatal("a Crusher that is not monstrous triggered")
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(mine) {
			t.Error("my enchantment was destroyed")
		}
		return !g.Battlefield.Contains(theirs)
	}
	if run(false) {
		t.Error("the enchantment was destroyed by a Crusher that is not monstrous")
	}
	if !run(true) {
		t.Error("the monstrous Crusher did not destroy the enchantment")
	}
}

// oldB396SpellOnStack has the active player cast a spell and passes priority to
// the responder, who then casts a counterspell.
func oldB396SpellOnStack(t *testing.T, g *game.Game, colors ...string) (me, opp *game.Player, victim uuid.UUID) {
	t.Helper()
	me, opp = effectstest.GptB1Seats(g)
	effectstest.ToMainForCost(t, g)
	victim = effectstest.HandCardFull(me, "Their Spell", "Sorcery", "", "", colors)
	if err := g.CastSpell(me.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the victim: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	return me, opp, victim
}

func TestOldB396DissolveCountersAndScries(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, victim := oldB396SpellOnStack(t, g)
	effectstest.CastFromSeat(t, g, opp, "Dissolve", "Instant", oldB396Dissolve, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) || !g.Seats[0].Graveyard.Contains(victim) {
		t.Error("the spell was not countered")
	}
	if effectstest.ScryChoiceFor(g, opp.ID) == nil {
		t.Fatal("no scry prompt for the caster")
	}
	effectstest.B26AnswerScryAllTop(t, g, opp.ID)
}

func TestOldB396GainsayCountersOnlyBlueSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp, plain := oldB396SpellOnStack(t, g)
	card := effectstest.HandCardFull(opp, "Gainsay", "Instant", "", oldB396Gainsay, nil)
	if err := g.CastSpell(opp.ID, card, game.CastSpellParams{Targets: effectstest.CardRefs(plain), Strict: true}); err == nil {
		t.Fatal("Gainsay targeted a spell that is not blue")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp, blue := oldB396SpellOnStack(t, g, "U")
	effectstest.CastFromSeat(t, g, opp, "Gainsay", "Instant", oldB396Gainsay, effectstest.CardRefs(blue))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Stack.Contains(blue) || !me.Graveyard.Contains(blue) {
		t.Error("the blue spell was not countered")
	}
}

func TestOldB396StymiedHopesCountersUnlessPaidThenScries(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, victim := oldB396SpellOnStack(t, g)
		effectstest.CastFromSeat(t, g, opp, "Stymied Hopes", "Instant", oldB396Stymied, effectstest.CardRefs(victim))
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			me.ManaPool.AddMana(game.ManaToken{Color: "C"})
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		countered := me.Graveyard.Contains(victim) && !g.Stack.Contains(victim)
		if countered == pay {
			t.Errorf("paid %v: countered %v", pay, countered)
		}
		if effectstest.ScryChoiceFor(g, opp.ID) == nil {
			t.Fatalf("paid %v: no scry prompt", pay)
		}
		effectstest.B26AnswerScryAllTop(t, g, opp.ID)
	}
}

func TestOldB396MistcutterHydraEntersWithXCountersAndCantBeCountered(t *testing.T) {
	spec, _ := effects.Lookup(oldB396Hydra)
	if !spec.CantBeCountered {
		t.Error("the Hydra can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mistcutter Hydra", TypeLine: "Creature — Hydra", OracleID: oldB396Hydra, ManaCost: "{X}{G}",
		Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "G", 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("cast X=3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, id); p != 3 || tough != 3 {
		t.Errorf("the Hydra is %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "haste") || !effectstest.HasEffectiveKeyword(t, g, id, "protection from blue") {
		t.Error("the Hydra should have haste and protection from blue")
	}
}

func TestOldB396ReverentHunterGetsCountersEqualToYourDevotionToGreen(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Green Rite", TypeLine: "Enchantment", ManaCost: "{G}{G}", Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Reverent Hunter", TypeLine: "Creature — Human Archer", OracleID: oldB396Hunter, ManaCost: "{2}{G}",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "G", 3)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// {G}{G} on the board plus the Hunter's own {G}: three counters on a 1/1.
	if p, tough := effectstest.PtOf(t, g, id); p != 4 || tough != 4 {
		t.Errorf("the Hunter is %d/%d, want 4/4", p, tough)
	}
}

func TestOldB396TriadOfFatesMarksBlinksAndExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	triad := effectstest.B12Push(g, me.ID, "Triad of Fates", "Legendary Creature — Human Wizard", oldB396Triad, 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", oldB396Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	// The fate counter goes on another creature, never on the Triad.
	effectstest.FillPool(me, 1)
	if err := g.ActivateCatalogAbility(me.ID, triad, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(triad), Strict: true}); err == nil {
		t.Fatal("the Triad targeted itself")
	}
	oldB396Activate(t, g, me, triad, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)})
	if n := effectstest.E2Card(t, g, other).Counters["fate"]; n != 1 {
		t.Fatalf("fate counters %d, want 1", n)
	}
	// Only a creature with a fate counter is a legal target of the others.
	effectstest.P1g3Untap(g, triad)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, triad, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(triad), Strict: true}); err == nil {
		t.Fatal("blinked a creature with no fate counter")
	}
	oldB396Activate(t, g, me, triad, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)})
	if !oldB396OnBattlefieldAs(g, opp.ID, "Other") {
		t.Fatal("the blinked creature did not return under its owner's control")
	}
	if n := effectstest.E2Card(t, g, oldB396FindByName(g, "Other")).Counters["fate"]; n != 0 {
		t.Errorf("the returned creature kept %d fate counters", n)
	}
}

func oldB396FindByName(g *game.Game, name string) uuid.UUID {
	var id uuid.UUID
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				id = c.InstanceID
			}
		}
	})
	return id
}

func oldB396OnBattlefieldAs(g *game.Game, controller uuid.UUID, name string) bool {
	found := false
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				found = true
			}
		}
	})
	return found
}

func TestOldB396TriadOfFatesExileMakesItsControllerDrawTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	triad := effectstest.B12Push(g, me.ID, "Triad of Fates", "Legendary Creature — Human Wizard", oldB396Triad, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB396Bear, 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, victim).Counters = map[string]int{"fate": 1} })
	hand := opp.Hand.Size()
	effectstest.FillPoolColored(me, "B", 1)
	oldB396Activate(t, g, me, triad, 2, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	if g.Battlefield.Contains(victim) || !effectstest.ExileHas(g, victim) {
		t.Error("the creature was not exiled")
	}
	if opp.Hand.Size() != hand+2 {
		t.Errorf("its controller's hand %d, want %d", opp.Hand.Size(), hand+2)
	}
}

func TestOldB396WarriorsLessonGivesTwoCreaturesADrawOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	a := effectstest.B12Creature(g, me.ID, "A", oldB396Bear, 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", oldB396Bear, 2, 2)
	c := effectstest.B12Creature(g, me.ID, "C", oldB396Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Warriors' Lesson", "Instant", oldB396Lesson, effectstest.CardRefs(a, b, c)); err == nil {
		t.Fatal("three targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Warriors' Lesson", "Instant", oldB396Lesson, effectstest.CardRefs(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, a, b, c)
	effectstest.InitSettle(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want +2 (the two target creatures, not the third)", hand, me.Hand.Size())
	}
}

func TestOldB396ThassasBountyDrawsThreeAndMillsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hand, lib := me.Hand.Size(), opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Thassa's Bounty", "Sorcery", oldB396Bounty, effectstest.Pr6Player(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d → %d, want +3", hand, me.Hand.Size())
	}
	if opp.Library.Size() != lib-3 {
		t.Errorf("their library %d → %d, want -3", lib, opp.Library.Size())
	}
}
