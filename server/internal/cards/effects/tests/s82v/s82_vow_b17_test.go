package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	vowB17Donal     = "1c9f2b6e-a0eb-493e-8c43-3a14f6d95a53"
	vowB17Haunting  = "c96ed5b2-6286-4f54-becb-c07f54335e63"
	vowB17Dollhouse = "73970475-f899-4e93-8498-286240181295"
	vowB17Olivia    = "8922a91e-3d15-4351-8d77-e0d6bc4de82e"
	vowB17Waltz     = "36a7d1a6-4eb8-4ddc-942c-00fd2acbf2ee"
	vowB17Salvation = "a5b0f351-5536-41ec-aa09-1e5344a0effb"
	vowB17Edgar     = "2c04371a-af69-4e42-8cda-096cf8312b5e"
	vowB17Geralf    = "0a0743b0-07c2-4e11-8b30-6b1b8156bdd0"
)

func TestVowB17CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		vowB17Donal: "Donal, Herald of Wings", vowB17Haunting: "Haunting Imitation",
		vowB17Dollhouse: "Dollhouse of Horrors", vowB17Olivia: "Olivia, Crimson Bride",
		vowB17Waltz: "Sinister Waltz", vowB17Salvation: "Sudden Salvation",
		vowB17Edgar: "Edgar's Awakening", vowB17Geralf: "Geralf, Visionary Stitcher",
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

func vowB17Grave(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func vowB17Targets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}

// --- Edgar's Awakening -------------------------------------------------------

func TestVowB17EdgarsAwakeningReturnsACreatureAndRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := vowB17Grave(me, "Bear", "Creature — Bear")
	bolt := vowB17Grave(me, "Bolt", "Instant")
	if err := effectstest.CastCatalogSpellErr(t, g, "Edgar's Awakening", "Sorcery", vowB17Edgar, vowB17Targets(bolt)); err == nil {
		t.Fatal("Edgar's Awakening targeted an instant card")
	}
	effectstest.CastCatalogSpell(t, g, "Edgar's Awakening", "Sorcery", vowB17Edgar, vowB17Targets(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); !ok {
		t.Fatal("the creature card did not return to the battlefield")
	}
}

func TestVowB17EdgarsAwakeningDiscardPaysBToReturnACreatureToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := vowB17Grave(me, "Bear", "Creature — Bear")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Edgar's Awakening", TypeLine: "Sorcery", OracleID: vowB17Edgar, Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:]
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Fatal("the creature card is not in hand after paying {B}")
	}
}

func TestVowB17EdgarsAwakeningDiscardDeclinedReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := vowB17Grave(me, "Bear", "Creature — Bear")
	me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Edgar's Awakening", TypeLine: "Sorcery", OracleID: vowB17Edgar, Owner: me.ID, Controller: me.ID})
	me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:]
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(me.ID, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "B")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(bear) || !me.Graveyard.Contains(bear) {
		t.Fatal("declining {B} still returned the creature card")
	}
}

// --- Sinister Waltz --------------------------------------------------------------

func TestVowB17SinisterWaltzReturnsTwoAtRandomAndBottomsTheThird(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := vowB17Grave(me, "A", "Creature — Bear")
	b := vowB17Grave(me, "B", "Creature — Bear")
	c := vowB17Grave(me, "C", "Creature — Bear")
	if err := effectstest.CastCatalogSpellErr(t, g, "Sinister Waltz", "Sorcery", vowB17Waltz, vowB17Targets(a, b)); err == nil {
		t.Fatal("Sinister Waltz accepted two targets")
	}
	effectstest.CastCatalogSpell(t, g, "Sinister Waltz", "Sorcery", vowB17Waltz, vowB17Targets(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	on, bottom := 0, uuid.Nil
	for _, id := range []uuid.UUID{a, b, c} {
		if _, ok := effectstest.BattlefieldCard(g, id); ok {
			on++
		} else if me.Library.Contains(id) {
			bottom = id
		}
	}
	if on != 2 || bottom == uuid.Nil {
		t.Fatalf("%d returned, bottom card %v, want two returned and one in the library", on, bottom)
	}
	if n := len(me.Library.Cards); me.Library.Cards[n-1].InstanceID != bottom && me.Library.Cards[0].InstanceID != bottom {
		t.Error("the third card is not at an end of the library")
	}
	if got := len(me.Graveyard.Cards); got != 1 { // the resolved Waltz itself
		t.Errorf("graveyard holds %d cards, want only the spell", got)
	}
}

// --- Sudden Salvation ---------------------------------------------------------------

func TestVowB17SuddenSalvationReturnsTappedAndDrawsPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirs2 := effectstest.B12Creature(g, opp.ID, "Their Other Bear", "Creature — Bear", 2, 2)
	old := vowB17Grave(opp, "Old Bear", "Creature — Bear")
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(mine)
		_ = g.DestroyPermanentForEffect(theirs)
		_ = g.DestroyPermanentForEffect(theirs2)
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Sudden Salvation", "Instant", vowB17Salvation, vowB17Targets(old)); err == nil {
		t.Fatal("Sudden Salvation targeted a card that did not die this turn")
	}
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sudden Salvation", "Instant", vowB17Salvation, vowB17Targets(mine, theirs, theirs2))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, theirs2} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || !c.Tapped {
			t.Fatalf("%v: on battlefield %v tapped %v, want returned tapped", id, ok, c.Tapped)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("an opponent's creature came back under the caster's control")
	}
	// One opponent controls two of them: one card, not two.
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d cards, want 1 (one opponent controls the returned permanents)", got)
	}
}

func TestVowB17SuddenSalvationDrawsNothingForOnlyYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Sudden Salvation", "Instant", vowB17Salvation, vowB17Targets(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mine); !ok {
		t.Fatal("the creature did not return")
	}
	if got := me.Hand.Size() - hand; got != 0 {
		t.Errorf("drew %d cards, want 0", got)
	}
}

// --- Haunting Imitation ----------------------------------------------------------------

func vowB17TopOfLibrary(p *game.Player, name, typeLine string, power, toughness int, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: power, Toughness: toughness,
		Keywords: keywords, Owner: p.ID, Controller: p.ID})
	return id
}

func TestVowB17HauntingImitationCopiesEachRevealedCreatureAsAOneOneSpiritWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := vowB17TopOfLibrary(me, "Hill Giant", "Creature — Giant", 3, 3)
	vowB17TopOfLibrary(g.Seats[1], "Forest", "Basic Land — Forest", 0, 0)
	b := vowB17TopOfLibrary(g.Seats[2], "Wurm", "Creature — Wurm", 5, 5)
	effectstest.CastCatalogSpell(t, g, "Haunting Imitation", "Sorcery", vowB17Haunting, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, name := range []string{"Hill Giant", "Wurm"} {
		toks := vowB17Tokens(g, me.ID, name)
		if len(toks) != 1 {
			t.Fatalf("%d %s tokens, want 1", len(toks), name)
		}
		c := effectstest.LayeredCard(t, g, toks[0])
		if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasSubtype("Spirit") || !c.IsCreature() {
			t.Errorf("%s token is %d/%d %v, want a 1/1 Spirit creature", name, c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
		}
		if !effectstest.EffectiveAbilitiesContain(t, g, toks[0], "flying") {
			t.Errorf("%s token has no flying", name)
		}
	}
	if !me.Library.Contains(a) || !g.Seats[2].Library.Contains(b) {
		t.Error("a revealed card left its library")
	}
	if !me.Graveyard.Contains(vowB17SpellID(me, "Haunting Imitation")) {
		t.Error("Haunting Imitation did not go to the graveyard after copying")
	}
}

func TestVowB17HauntingImitationReturnsToHandWhenNoCreatureWasRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, p := range g.Seats {
		vowB17TopOfLibrary(p, "Forest", "Basic Land — Forest", 0, 0)
	}
	effectstest.CastCatalogSpell(t, g, "Haunting Imitation", "Sorcery", vowB17Haunting, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(vowB17SpellID(me, "Haunting Imitation")) {
		t.Fatal("Haunting Imitation is not back in its owner's hand")
	}
	if len(vowB17Tokens(g, me.ID, "Forest")) != 0 {
		t.Error("a land was copied")
	}
}

func vowB17SpellID(p *game.Player, name string) uuid.UUID {
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		for _, c := range z.Cards {
			if c.Name == name {
				return c.InstanceID
			}
		}
	}
	return uuid.Nil
}

func vowB17Tokens(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name && c.Controller == controller && effects.IsToken(c) {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Geralf, Visionary Stitcher ----------------------------------------------------------

func TestVowB17GeralfGivesZombiesFlyingAndSacrificesForAToughnessSizedZombie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	geralf := effectstest.B12Push(g, me.ID, "Geralf, Visionary Stitcher", "Legendary Creature — Human Wizard", vowB17Geralf, 1, 4)
	zombie := effectstest.B12Creature(g, me.ID, "Plain Zombie", "Creature — Zombie", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Zombie", "Creature — Zombie", 2, 2)
	wall := effectstest.B12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 6)
	if !effectstest.EffectiveAbilitiesContain(t, g, zombie, "flying") || effectstest.EffectiveAbilitiesContain(t, g, theirs, "flying") {
		t.Fatal("flying should reach only the Zombies Geralf's controller controls")
	}
	effectstest.B06AddMana(me, "U")
	if err := g.ActivateCatalogAbility(me.ID, geralf, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{wall}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := vowB17Tokens(g, me.ID, "Zombie")
	if len(toks) != 1 {
		t.Fatalf("%d Zombie tokens, want 1", len(toks))
	}
	c := effectstest.LayeredCard(t, g, toks[0])
	if c.CurrentPower() != 6 || c.CurrentToughness() != 6 || !c.HasColor("U") {
		t.Errorf("token is %d/%d, want a blue 6/6 (the Wall's toughness)", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, toks[0], "flying") {
		t.Error("the new Zombie token has no flying")
	}
	if _, ok := effectstest.BattlefieldCard(g, wall); ok {
		t.Error("the Wall was not sacrificed")
	}
}

func TestVowB17GeralfCannotSacrificeATokenOrItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	geralf := effectstest.B12Push(g, me.ID, "Geralf, Visionary Stitcher", "Legendary Creature — Human Wizard", vowB17Geralf, 1, 4)
	tok := effectstest.PushCatalogPermanent(g, me.ID, "Servo", "Token Artifact Creature — Servo", "", false)
	effectstest.B06AddMana(me, "U", "U")
	if err := g.ActivateCatalogAbility(me.ID, geralf, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{tok}}); err == nil {
		t.Error("Geralf sacrificed a token")
	}
	if err := g.ActivateCatalogAbility(me.ID, geralf, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{geralf}}); err == nil {
		t.Error("Geralf sacrificed itself")
	}
}

// --- Dollhouse of Horrors ----------------------------------------------------------------

func TestVowB17DollhouseCopiesTheExiledCreatureAsAConstructThatGrowsAndHasHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	house := effectstest.PushCatalogPermanent(g, me.ID, "Dollhouse of Horrors", "Artifact", vowB17Dollhouse, false)
	dead := vowB17Grave(me, "Hill Giant", "Creature — Giant")
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, house, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := vowB17Tokens(g, me.ID, "Hill Giant")
	if len(toks) != 1 {
		t.Fatalf("%d Hill Giant tokens, want 1", len(toks))
	}
	c := effectstest.LayeredCard(t, g, toks[0])
	if !c.IsArtifact() || !c.IsCreature() || !c.HasSubtype("Giant") || !c.HasSubtype("Construct") {
		t.Errorf("token types %v %v, want an artifact Giant Construct creature", c.Effective().Types, c.Effective().Subtypes)
	}
	// 0/0 plus one for the one Construct (itself).
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("token is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, toks[0], "haste") {
		t.Error("the token has no haste this turn")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled as the cost")
	}
	// A second Construct makes both bigger.
	vowB17Construct := effectstest.PushCatalogPermanent(g, me.ID, "Other Construct", "Artifact Creature — Construct", "", false)
	_ = vowB17Construct
	c = effectstest.LayeredCard(t, g, toks[0])
	if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
		t.Errorf("with two Constructs the token is %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestVowB17DollhouseIsSorceryOnlyAndNeedsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	house := effectstest.PushCatalogPermanent(g, me.ID, "Dollhouse of Horrors", "Artifact", vowB17Dollhouse, false)
	bolt := vowB17Grave(me, "Bolt", "Instant")
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, house, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{bolt}}); err == nil {
		t.Error("Dollhouse exiled a noncreature card")
	}
	dead := vowB17Grave(me, "Bear", "Creature — Bear")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	if err := g.ActivateCatalogAbility(me.ID, house, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}}); err == nil {
		t.Error("Dollhouse activated outside a main phase")
	}
}

// --- Donal, Herald of Wings ---------------------------------------------------------------------

func TestVowB17DonalCopiesANonlegendaryFlyerAsAOneOneSpiritOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Donal, Herald of Wings", "Legendary Creature — Human Wizard", vowB17Donal, 3, 3)
	vowB17CastCreature(t, g, "Sky Wurm", "Creature — Wurm", 4, 4, "flying")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	toks := vowB17Tokens(g, me.ID, "Sky Wurm")
	if len(toks) != 1 {
		t.Fatalf("%d Sky Wurm tokens, want 1", len(toks))
	}
	c := effectstest.LayeredCard(t, g, toks[0])
	if c.CurrentPower() != 1 || c.CurrentToughness() != 1 || !c.HasSubtype("Spirit") || !c.HasSubtype("Wurm") {
		t.Errorf("copy is %d/%d %v, want a 1/1 Spirit Wurm", c.CurrentPower(), c.CurrentToughness(), c.Effective().Subtypes)
	}
	// Once each turn: a second flyer raises no prompt.
	vowB17CastCreature(t, g, "Second Flyer", "Creature — Bird", 2, 2, "flying")
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Error("a second flying spell raised Donal's prompt in the same turn")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(vowB17Tokens(g, me.ID, "Second Flyer")); got != 0 {
		t.Errorf("%d copies of the second flyer, want 0", got)
	}
}

func TestVowB17DonalIgnoresLegendaryAndNonFlyingSpellsAndDecliningIsFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Donal, Herald of Wings", "Legendary Creature — Human Wizard", vowB17Donal, 3, 3)
	vowB17CastCreature(t, g, "Bear", "Creature — Bear", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	vowB17CastCreature(t, g, "Legend Bird", "Legendary Creature — Bird", 2, 2, "flying")
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a nonflying or legendary spell raised Donal's prompt")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Fatal("a nonflying or legendary spell raised Donal's prompt")
	}
	vowB17CastCreature(t, g, "Bird One", "Creature — Bird", 1, 1, "flying")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	vowB17CastCreature(t, g, "Bird Two", "Creature — Bird", 1, 1, "flying")
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if len(vowB17Tokens(g, me.ID, "Bird One")) != 0 || len(vowB17Tokens(g, me.ID, "Bird Two")) != 1 {
		t.Error("declining must not use up the turn: want no copy of Bird One and one of Bird Two")
	}
}

func vowB17CastCreature(t *testing.T, g *game.Game, name, typeLine string, p, tough int, keywords ...string) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: p, Toughness: tough,
		Keywords: keywords, Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// --- Olivia, Crimson Bride ------------------------------------------------------------------------

func TestVowB17OliviaReturnsACreatureTappedAndAttackingThatIsExiledWithoutAVampire(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	olivia := effectstest.B12Push(g, me.ID, "Olivia, Crimson Bride", "Legendary Creature — Vampire Noble", vowB17Olivia, 3, 4)
	dead := vowB17Grave(me, "Hill Giant", "Creature — Giant")
	effectstest.DeclareAttack(t, g, opp.ID, olivia)
	effectstest.PickTriggerTarget(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if r := effectstest.LatestOptionPickFor(g, me.ID); r != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature card did not return")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Controller != me.ID {
		t.Errorf("tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	// Olivia is a legendary Vampire, so the returned creature stays.
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, dead); !ok {
		t.Fatal("the creature was exiled while Olivia was on the battlefield")
	}
	// Olivia leaves: the state trigger exiles it.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(olivia) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, dead); ok {
		t.Fatal("the creature stayed after its controller lost their legendary Vampire")
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature is not in exile")
	}
}
