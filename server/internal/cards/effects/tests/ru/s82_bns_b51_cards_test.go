package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82BnsFumulusOracle  = "c75f8365-71f1-415f-8bbc-9bacb8ddba0b"
	s82BnsGornogOracle   = "adfd6c2d-a7f8-4f5c-bc31-d65c8150ed49"
	s82BnsHurskaOracle   = "8d49b6a1-4ae2-4443-a6b3-9ef806cb6fc4"
	s82BnsIsarethOracle  = "7af0e2da-9163-42d9-bf69-439cc61cd28d"
	s82BnsIvoraOracle    = "d847542b-7522-43ec-91b0-f0aba5d7dddd"
	s82BnsMowuOracle     = "34ceb733-0329-4c1b-8d75-25c34fd4400f"
	s82BnsNabanOracle    = "cffb9b7f-d52c-43e0-a4a6-b06dea83eb53"
	s82BnsNazarOracle    = "beeed1c5-2f51-4e40-87b9-9a6f810ee4cb"
	s82BnsNeerdivOracle  = "b015ec84-95b9-4010-afef-d9d0cacb98dd"
	s82BnsOzoxOracle     = "04e575d1-f843-4330-972f-e77b264073e0"
	s82BnsPlagonOracle   = "1c95b7cf-be4d-4961-9073-3e39487492c8"
	s82BnsOzoxCardName   = "Ozox, the Clattering King"
	s82BnsJumblebonesTok = "Jumblebones"
)

func TestS82BnsB51EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82BnsFumulusOracle: "Fumulus, the Infestation",
		s82BnsGornogOracle:  "Gornog, the Red Reaper",
		s82BnsHurskaOracle:  "Hurska Sweet-Tooth",
		s82BnsIsarethOracle: "Isareth the Awakener",
		s82BnsIvoraOracle:   "Ivora, Insatiable Heir",
		s82BnsMowuOracle:    "Mowu, Loyal Companion",
		s82BnsNabanOracle:   "Naban, Dean of Iteration",
		s82BnsNazarOracle:   "Nazar, the Velvet Fang",
		s82BnsNeerdivOracle: "Neerdiv, Devious Diver",
		s82BnsOzoxOracle:    "Ozox, the Clattering King",
		s82BnsPlagonOracle:  "Plagon, Lord of the Beach",
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

func s82BnsAttack(t *testing.T, g *game.Game, attacker, defender uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
}

func s82BnsLife(g *game.Game, p *game.Player) int {
	n := 0
	g.ReadSnapshot(func() { n = p.Life })
	return n
}

func s82BnsGain(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, n) })
}

func s82BnsPay(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

func s82BnsCounters(g *game.Game, id uuid.UUID, kind string) int {
	return effectstest.CountersOn(g, id, kind)
}

func TestS82BnsB51MowuAddsOneCounterToItselfOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mowu := effectstest.B12Push(g, me.ID, "Mowu, Loyal Companion", "Legendary Creature — Dog", s82BnsMowuOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.AddCounter(mowu, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := s82BnsCounters(g, mowu, "+1/+1"); got != 2 {
		t.Errorf("Mowu has %d +1/+1 counters, want 2", got)
	}
	if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
		t.Fatal(err)
	}
	if got := s82BnsCounters(g, bear, "+1/+1"); got != 1 {
		t.Errorf("a Bear has %d +1/+1 counters, want 1", got)
	}
	if err := g.AddCounter(mowu, "feeding", 1); err != nil {
		t.Fatal(err)
	}
	if got := s82BnsCounters(g, mowu, "feeding"); got != 1 {
		t.Errorf("Mowu has %d feeding counters, want 1", got)
	}
}

func TestS82BnsB51NabanDoublesAWizardsEnterTrigger(t *testing.T) {
	for _, withNaban := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if withNaban {
			effectstest.B12Push(g, me.ID, "Naban, Dean of Iteration", "Legendary Creature — Human Wizard", s82BnsNabanOracle, 2, 1)
		}
		before := me.Hand.Size()
		s82BnsCastPlagon(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 1
		if withNaban {
			want = 2
		}
		if got := me.Hand.Size() - before; got != want {
			t.Errorf("Naban=%v: drew %d cards (net of the cast), want %d", withNaban, got, want)
		}
	}
}

func TestS82BnsB51PlagonDrawsPerToughnessGreaterAndAssignsByToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	before := me.Hand.Size()
	plagon := s82BnsCastPlagon(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("drew %d cards (net of the cast), want 2 (the Wall and Plagon)", got)
	}
	wall := effectstest.B12Creature(g, me.ID, "Tall Wall", "Creature — Wall", 1, 5)
	s82BnsPay(me, "U")
	effectstest.B16Activate(t, g, me.ID, plagon, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(wall)})
	if !effectstest.EffectiveOf(t, g, wall).DamageByToughnessAlways {
		t.Error("the target does not assign combat damage by toughness")
	}
	if effectstest.EffectiveOf(t, g, plagon).DamageByToughnessAlways {
		t.Error("the Starfish gained it too")
	}
	if err := g.ActivateCatalogAbility(me.ID, plagon, 0, game.ActivateAbilityParams{
		Strict: true, Targets: effectstest.B16TargetCard(effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Ogre", 3, 3))}); err == nil {
		t.Error("a creature you don't control was a legal target")
	}
}

func TestS82BnsB51OzoxCantBlockAndLeavesJumblebonesWhichReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ozox := effectstest.B12Push(g, me.ID, s82BnsOzoxCardName, "Legendary Creature — Skeleton Noble", s82BnsOzoxOracle, 3, 2)
	if effectstest.RestrictionsOf(t, g, ozox)&game.CantBlock == 0 {
		t.Error("Ozox can block")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ozox) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, s82BnsJumblebonesTok, me.ID); n != 1 {
		t.Fatalf("%d Jumblebones tokens, want 1", n)
	}
	var bones uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == s82BnsJumblebonesTok {
				bones = c.InstanceID
			}
		}
	})
	if c := effectstest.EffectiveOf(t, g, bones); c.Power != 2 || c.Toughness != 1 || !slices.Contains(c.Subtypes, "Skeleton") {
		t.Errorf("Jumblebones is %d/%d %v, want a 2/1 Skeleton", c.Power, c.Toughness, c.Subtypes)
	}
	if effectstest.RestrictionsOf(t, g, bones)&game.CantBlock == 0 {
		t.Error("Jumblebones can block")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bones) })
	effectstest.AnswerPickTargetIfOpen(t, g, ozox)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ozox) {
		t.Error("Ozox did not return to hand")
	}
}

func TestS82BnsB51IvoraMakesBloodAndGrowsOnDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ivora := effectstest.CastAndResolveCreature(t, g, "Ivora, Insatiable Heir", "Legendary Creature — Vampire Warrior", s82BnsIvoraOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Blood", me.ID); n != 1 {
		t.Errorf("%d Blood tokens after entering, want 1", n)
	}
	effectstest.DealCombatDamageToPlayer(g, ivora, opp.ID, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Blood", me.ID); n != 2 {
		t.Errorf("%d Blood tokens after combat damage, want 2", n)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventDiscardCard, Actor: me.ID, CardID: uuid.New()})
		g.EmitEvent(game.Event{Kind: game.EventDiscardCard, Actor: opp.ID, CardID: uuid.New()})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82BnsCounters(g, ivora, "+1/+1"); got != 1 {
		t.Errorf("Ivora has %d counters, want 1 (only your own discard counts)", got)
	}
}

func TestS82BnsB51NazarFeedsOnLifeGainAndCashesInThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	nazar := effectstest.B12Push(g, me.ID, "Nazar, the Velvet Fang", "Legendary Creature — Vampire Warlock", s82BnsNazarOracle, 3, 3)
	for range 3 {
		s82BnsGain(g, me, 2)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := s82BnsCounters(g, nazar, "feeding"); got != 3 {
		t.Fatalf("%d feeding counters, want 3 (one per gain)", got)
	}
	hand, life := me.Hand.Size(), s82BnsLife(g, me)
	s82BnsAttack(t, g, nazar, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	s82BnsAnswerYes(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82BnsCounters(g, nazar, "feeding"); got != 0 {
		t.Errorf("%d feeding counters left, want 0", got)
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
	if got := s82BnsLife(g, me); got != life-3 {
		t.Errorf("life %d, want %d", got, life-3)
	}
}

func TestS82BnsB51NeerdivMillsOnTapAndDrawsFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	neerdiv := effectstest.B12Push(g, me.ID, "Neerdiv, Devious Diver", "Legendary Creature — Merfolk Rogue", s82BnsNeerdivOracle, 2, 2)
	lib := opp.Library.Size()
	if err := g.TapCard(neerdiv, true); err != nil {
		t.Fatal(err)
	}
	if pt := effectstest.LatestPickTarget(g, me.ID); pt != nil {
		if err := g.ResolvePickTarget(pt.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := lib - opp.Library.Size(); got != 2 {
		t.Errorf("the opponent milled %d, want 2 (Neerdiv's power)", got)
	}
	hand := me.Hand.Size()
	zombie := effectstest.PushGraveyardCreature(g, me.ID, "Zombie", "{1}")
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventActivateAbility, Actor: me.ID, CardID: zombie, ActivatedFrom: game.ZoneGraveyard})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d on a graveyard activation, want 1", got)
	}
	if got := s82BnsCounters(g, neerdiv, "+1/+1"); got != 1 {
		t.Errorf("%d counters, want 1", got)
	}
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventActivateAbility, Actor: me.ID, CardID: uuid.New()})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82BnsCounters(g, neerdiv, "+1/+1"); got != 1 {
		t.Errorf("an ordinary activation grew Neerdiv to %d counters", got)
	}
}

func s82BnsAnswerYes(t *testing.T, g *game.Game, who uuid.UUID) {
	t.Helper()
	ask := effectstest.AssistCardTriggerPrompt(g, who)
	if ask == nil {
		t.Fatal("no optional-trigger prompt is open")
	}
	if err := g.ResolveTriggerPrompt(ask.ID, who, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

// s82BnsCastPlagon casts a real 0/3 Plagon from hand (the shared cast helper
// builds a 0/0) and lets the spell resolve.
func s82BnsCastPlagon(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Plagon, Lord of the Beach", TypeLine: "Legendary Creature — Starfish Wizard",
		OracleID: s82BnsPlagonOracle, ManaCost: "{2}{U}", Power: 0, Toughness: 3,
		Owner: active.ID, Controller: active.ID,
	})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func TestS82BnsB51FumulusMakesInsectsOnNontokenCreatureSacrificesAndDrainsOnInsectAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fumulus := effectstest.B12Push(g, me.ID, "Fumulus, the Infestation", "Legendary Creature — Vampire Insect", s82BnsFumulusOracle, 2, 2)
	insects := func() int { return effectstest.TokensNamedControlledBy(g, "Insect", me.ID) }

	theirs := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := insects(); got != 1 {
		t.Fatalf("%d Insects after an opponent sacrificed a creature, want 1 (made for you)", got)
	}
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(rock) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := insects(); got != 1 {
		t.Errorf("a sacrificed artifact made an Insect (%d)", got)
	}
	var token uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Insect" && c.Controller == me.ID {
				token = c.InstanceID
			}
		}
	})
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := insects(); got != 0 {
		t.Errorf("sacrificing a token left %d Insects, want 0 (the token itself gone, none made)", got)
	}

	life, oppLife := s82BnsLife(g, me), s82BnsLife(g, opp)
	s82BnsAttack(t, g, fumulus, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := s82BnsLife(g, opp); got != oppLife-1 {
		t.Errorf("defending player life %d, want %d", got, oppLife-1)
	}
	if got := s82BnsLife(g, me); got != life+1 {
		t.Errorf("your life %d, want %d", got, life+1)
	}
}

func TestS82BnsB51GornogMakesACowardAndPumpsAttackingWarriors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gornog := effectstest.B12Push(g, me.ID, "Gornog, the Red Reaper", "Legendary Creature — Minotaur Warrior", s82BnsGornogOracle, 2, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Goblin", "Creature — Goblin", 1, 1)
	s82BnsAttack(t, g, gornog, opp.ID)
	effectstest.PickTriggerTarget(t, g, me.ID, ogre)
	effectstest.PassPriorityAroundTable(t, g)
	if !slices.Contains(effectstest.EffectiveOf(t, g, ogre).Subtypes, "Coward") || !slices.Contains(effectstest.EffectiveOf(t, g, ogre).Subtypes, "Ogre") {
		t.Errorf("the Ogre is %v, want a Coward in addition to being an Ogre", effectstest.EffectiveOf(t, g, ogre).Subtypes)
	}
	if slices.Contains(effectstest.EffectiveOf(t, g, other).Subtypes, "Coward") {
		t.Error("the other creature became a Coward")
	}
	if got := effectstest.EffectivePower(t, g, gornog); got != 3 {
		t.Errorf("attacking Gornog has power %d, want 3 (2 + one Coward)", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("a non-Warrior Bear has power %d, want 2", got)
	}
	if effectstest.RestrictionsOf(t, g, ogre)&game.CantBlock != 0 {
		t.Log("(the Coward's block rule is a pairing rule, not a restriction)")
	}
}

func TestS82BnsB51HurskaMakesFoodAndPumpsWhenYouPayAfterGainingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hurska := effectstest.B12Push(g, me.ID, "Hurska Sweet-Tooth", "Legendary Creature — Bear", s82BnsHurskaOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	s82BnsAttack(t, g, hurska, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Food", me.ID); n != 1 {
		t.Fatalf("%d Food tokens after attacking, want 1", n)
	}
	s82BnsGain(g, me, 4)
	effectstest.PassPriorityAroundTable(t, g)
	s82BnsPay(me, "W")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 6 || tough != 6 {
		t.Errorf("the Bear is %d/%d, want 6/6 (+4/+4 for 4 life)", p, tough)
	}
	// Declining the payment does nothing.
	s82BnsGain(g, me, 2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 6 {
		t.Errorf("the Bear's power is %d after a declined payment, want 6", p)
	}
}

func TestS82BnsB51IsarethPaysXToReanimateAndExilesItInstead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	isareth := effectstest.B12Push(g, me.ID, "Isareth the Awakener", "Legendary Creature — Human Wizard", s82BnsIsarethOracle, 3, 3)
	two := effectstest.GyCard(me, "Two Bear", "Creature — Bear", "{1}{G}", "")
	g.WithWriteLock(func() {
		me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Power, me.Graveyard.Cards[len(me.Graveyard.Cards)-1].Toughness = 2, 2
	})
	three := effectstest.GyCard(me, "Three Ogre", "Creature — Ogre", "{2}{R}", "")
	theirs := effectstest.GyCard(opp, "Their Two", "Creature — Bear", "{1}{G}", "")
	s82BnsAttack(t, g, isareth, opp.ID)
	s82BnsPay(me, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	ch := effectstest.PickTargetPrompt(t, g)
	if !slices.Contains(ch.PickTargetCards, two) || slices.Contains(ch.PickTargetCards, three) || slices.Contains(ch.PickTargetCards, theirs) {
		t.Fatalf("only a creature card of mana value 2 in YOUR graveyard: %v", ch.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, two)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(two) {
		t.Fatal("the creature did not return")
	}
	if got := s82BnsCounters(g, two, "corpse"); got != 1 {
		t.Errorf("%d corpse counters, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(two) })
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(two); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the creature should be exiled instead of dying, it is in %v", z)
	}
}
