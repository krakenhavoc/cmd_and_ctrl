package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 277 (Eldritch Moon): Impetuous Devils, Markov Crusader,
// Distended Mindbender, Foul Emissary, Oath of Liliana, Otherworldly
// Outburst, Tangleclaw Werewolf, Emrakul's Evangel, Kessig Prowler,
// Providence, Lone Rider, Prying Questions.

const (
	oldB277Devils     = "01add64a-205d-4c0e-a393-89e7a1a0ea44"
	oldB277Crusader   = "0a8511b9-ed7d-4e33-9c2f-96b9a4c535e2"
	oldB277Mindbender = "36d32194-d547-4e2d-894f-f4bf13cf7da6"
	oldB277Emissary   = "034c2b40-6bcc-4bb5-be94-4602496ddcc6"
	oldB277Oath       = "0bd7fd12-a0b7-4b4e-8a33-cace6e6bd6d1"
	oldB277Outburst   = "43feb61f-5de0-41b6-b678-9785c476dc08"
	oldB277Tangleclaw = "280d5601-802c-4607-9655-75cc83248210"
	oldB277Evangel    = "99eb572b-85b1-4d78-aea8-4400a5b7e508"
	oldB277Prowler    = "d2aff0ea-5474-4329-ae6a-80ee4908ee8d"
	oldB277Providence = "43b164d5-2bff-4a03-a4fa-16ffad7e3164"
	oldB277LoneRider  = "849741a3-9e64-41a9-a4f4-d146a7d42389"
	oldB277Questions  = "f33ef275-632e-46c8-a735-042dee340a12"
)

func TestOldB277CardsAreFullWithNoCaveats(t *testing.T) {
	for _, oracle := range []string{oldB277Devils, oldB277Crusader, oldB277Mindbender, oldB277Emissary, oldB277Oath, oldB277Outburst,
		oldB277Tangleclaw, oldB277Tangleclaw + "#1", oldB277Evangel, oldB277Prowler, oldB277Prowler + "#1", oldB277Providence,
		oldB277LoneRider, oldB277LoneRider + "#1", oldB277Questions} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %v with caveats %v, want full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB277DevilsForceABlockAndAreSacrificedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	devils := effectstest.B12Push(g, me.ID, "Impetuous Devils", "Creature — Devil", oldB277Devils, 6, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, devils)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p != nil {
		if effectstest.HasID(p.PickTargetCards, mine) {
			t.Error("your own creature is offered")
		}
		effectstest.PickCard(t, g, me.ID, theirs)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BlocksAttackerRecords(g); len(got) != 1 {
		t.Errorf("%d forced-block records, want 1", len(got))
	}
	effectstest.B35Block(t, g, theirs, devils)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(devils) {
		t.Error("the Devils were not sacrificed at the end step")
	}
}

func TestOldB277MarkovCrusaderHasHasteOnlyBesideAnotherVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crusader := effectstest.B12Push(g, me.ID, "Markov Crusader", "Creature — Vampire Knight", oldB277Crusader, 4, 3)
	effectstest.B12Creature(g, opp.ID, "Their Vampire", "Creature — Vampire", 1, 1)
	g.RunStateChecksForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, crusader, "haste") {
		t.Fatal("hasty with only an opponent's Vampire")
	}
	effectstest.B12Creature(g, me.ID, "My Vampire", "Creature — Vampire", 1, 1)
	g.BumpLayerVersionForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, crusader, "haste") {
		t.Error("not hasty beside another Vampire you control")
	}
}

func TestOldB277MindbenderEmergesAndTakesOneCardPerFilter(t *testing.T) {
	g, me, opp := effectstest.EmergeTable(t)
	opp.Hand.Cards = nil
	cheap := effectstest.HandCardOf(g, opp, "Cheap", "Sorcery", "{1}{B}", "")
	big := effectstest.HandCardOf(g, opp, "Big", "Creature — Giant", "{3}{R}{R}", "")
	land := effectstest.HandCardOf(g, opp, "Forest", "Basic Land — Forest", "", "")
	mind := effectstest.HandCardOf(g, me, "Distended Mindbender", "Creature — Eldrazi Insect", "{8}", oldB277Mindbender)
	fodder := effectstest.EmergeFodder(g, me.ID, "Fodder", "{3}", 1, 1)
	if err := effectstest.CastEmerge(g, me, mind, fodder, false); err != nil {
		t.Fatal(err)
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.OpenVariant(t, g)
	if c.Count != 2 || effectstest.HasID(c.DiscardOptions, land) {
		t.Fatalf("want two picks and no land offered: count %d options %v", c.Count, c.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{cheap, big}); err != nil {
		t.Fatal(err)
	}
	for _, id := range []uuid.UUID{cheap, big} {
		if !opp.Graveyard.Contains(id) {
			t.Errorf("%s was not discarded", id)
		}
	}
	if !opp.Hand.Contains(land) {
		t.Error("the land left the hand")
	}
}

func TestOldB277FoulEmissaryMakesAHorrorOnlyWhenSacrificedToEmerge(t *testing.T) {
	g, me, _ := effectstest.EmergeTable(t)
	emissary := effectstest.B12Push(g, me.ID, "Foul Emissary", "Creature — Human Horror", oldB277Emissary, 1, 1)
	other := effectstest.EmergeFodder(g, me.ID, "Fodder", "{3}", 1, 1)
	gryff := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	if err := effectstest.CastEmerge(g, me, gryff, other, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 0 {
		t.Fatalf("sacrificing another creature made %d Horrors", n)
	}
	gryff2 := effectstest.HandCardOf(g, me, "Wretched Gryff", "Creature — Eldrazi Hippogriff", "{7}", effectstest.WretchedGryffOracle)
	if err := effectstest.CastEmerge(g, me, gryff2, emissary, false); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Eldrazi Horror tokens after sacrificing the Emissary to emerge, want 1", n)
	}
}

func TestOldB277OathOfLilianaMakesAZombieAtEachEndStepAfterAPlaneswalkerEntered(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Oath of Liliana", "Legendary Enchantment", oldB277Oath, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the opponent's only creature was not sacrificed")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Zombie"); n != 0 {
		t.Fatalf("%d Zombies with no planeswalker entered", n)
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g2, me2.ID, "Oath of Liliana", "Legendary Enchantment", oldB277Oath, false)
	pw := effectstest.B12Permanent(g2, me2.ID, "Test Walker", "Legendary Planeswalker — Test")
	g2.WithWriteLock(func() { g2.EmitEvent(game.Event{Kind: game.EventETB, CardID: pw, Actor: me2.ID}) })
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if n, _ := effectstest.CountTokensNamed(g2, me2.ID, "Zombie"); n != 1 {
		t.Errorf("%d Zombies after a planeswalker entered, want 1", n)
	}
}

func TestOldB277OtherworldlyOutburstMakesAHorrorWhenThatCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Otherworldly Outburst", "Instant", oldB277Outburst, effectstest.CardRefs(victim))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, victim); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	effectstest.KillCreature(t, g, opp.ID, bystander)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 0 {
		t.Fatalf("another creature dying made %d Horrors", n)
	}
	effectstest.KillCreature(t, g, opp.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Horrors for the caster, want 1", n)
	}
}

func oldB277Transforms(t *testing.T, oracle, front, frontType, cost, p, tough, back, backType, bp, bt string, mana func(*game.Player)) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.VowB15Put(t, g, effectstest.VowB15Row(oracle, front, frontType, cost, p, tough, back, backType, bp, bt, []string{"G"}), me)
	effectstest.AdvanceToMain(t, g)
	mana(me)
	return g, me, card
}

func TestOldB277TangleclawTransformsAndMustBeBlocked(t *testing.T) {
	g, me, card := oldB277Transforms(t, oldB277Tangleclaw, "Tangleclaw Werewolf", "Creature — Werewolf Horror", "{2}{G}{G}", "2", "4",
		"Fibrous Entangler", "Creature — Eldrazi Werewolf", "4", "6", func(p *game.Player) { effectstest.FillPoolColored(p, "G", 7) })
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, card); f != 1 {
		t.Fatalf("face %d after the activation, want 1", f)
	}
	g.BumpLayerVersionForTest()
	if !effectstest.HasKeywordOnBattlefield(t, g, card, "vigilance") {
		t.Error("the back face lacks vigilance")
	}
}

func TestOldB277KessigProwlerTransforms(t *testing.T) {
	g, me, card := oldB277Transforms(t, oldB277Prowler, "Kessig Prowler", "Creature — Werewolf Horror", "{G}", "2", "1",
		"Sinuous Predator", "Creature — Eldrazi Werewolf", "4", "4", func(p *game.Player) { effectstest.FillPoolColored(p, "G", 5) })
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if f := effectstest.IwFace(g, card); f != 1 {
		t.Fatalf("face %d after the activation, want 1", f)
	}
}

func TestOldB277EvangelSacrificesItselfAndNonEldraziForAHorrorEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evangel := effectstest.B12Push(g, me.ID, "Emrakul's Evangel", "Creature — Human Horror", oldB277Evangel, 3, 2)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	eldrazi := effectstest.B12Creature(g, me.ID, "Eldrazi", "Creature — Eldrazi", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, evangel, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{a, eldrazi}}); err == nil {
		t.Fatal("sacrificed an Eldrazi creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, evangel, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(evangel) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the Evangel and the two creatures were not all sacrificed")
	}
	if !g.Battlefield.Contains(eldrazi) {
		t.Error("the Eldrazi was sacrificed")
	}
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 3 {
		t.Errorf("%d Horrors, want 3 (the Evangel and two others)", n)
	}
}

func TestOldB277EvangelAloneMakesOneHorror(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	evangel := effectstest.B12Push(g, me.ID, "Emrakul's Evangel", "Creature — Human Horror", oldB277Evangel, 3, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, evangel, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Eldrazi Horror"); n != 1 {
		t.Errorf("%d Horrors, want 1", n)
	}
}

func TestOldB277ProvidenceSetsYourLifeTo26AndDeclaresItsOpeningHandReveal(t *testing.T) {
	spec, _ := effects.Lookup(oldB277Providence)
	if spec.OpeningHand == nil || spec.OpeningHand.Reveal == nil || spec.OpeningHand.Reveal.At != game.RevealAtTheFirstUpkeep {
		t.Fatalf("opening-hand reveal at the first upkeep missing: %+v", spec.OpeningHand)
	}
	for _, start := range []int{40, 12} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		g.WithWriteLock(func() { me.Life = start })
		effectstest.CastCatalogSpell(t, g, "Providence", "Sorcery", oldB277Providence, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if me.Life != 26 {
			t.Errorf("life %d → %d, want 26", start, me.Life)
		}
	}
}

func TestOldB277LoneRiderTransformsAtTheEndStepAfterGainingThree(t *testing.T) {
	for _, gain := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		card := effectstest.VowB15Put(t, g, effectstest.VowB15Row(oldB277LoneRider, "Lone Rider", "Creature — Human Knight", "{1}{W}", "1", "1",
			"It That Rides as One", "Creature — Eldrazi Horror", "4", "4", []string{"W"}), me)
		effectstest.AdvanceToMain(t, g)
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, gain) })
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if gain >= 3 {
			want = 1
		}
		if f := effectstest.IwFace(g, card); f != want {
			t.Errorf("gained %d: face %d, want %d", gain, f, want)
		}
	}
}

func TestOldB277PryingQuestionsCostsThreeLifeAndAPutBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	opp.Hand.Cards = nil
	only := effectstest.HandCardOf(g, opp, "Only Card", "Instant", "{R}", "")
	_ = me
	life := opp.Life
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Prying Questions", "Sorcery", oldB277Questions,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d → %d, want -3", life, opp.Life)
	}
	if opp.Hand.Contains(only) || !opp.Library.Contains(only) {
		t.Error("the opponent's only card did not go on top of their library")
	}
}
