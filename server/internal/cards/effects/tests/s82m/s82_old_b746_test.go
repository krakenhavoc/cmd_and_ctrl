package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b746_test.go — S82 OLD batch 746 (Stronghold activated abilities,
// #5627), through the real engine.

const (
	oldB746Shapeshifter = "16e9a4ee-0398-4e76-bc7e-dd2d52056aaf"
	oldB746Trapper      = "3aa64915-8b1b-4414-b11e-d2078c1732a5"
	oldB746Tempting     = "3b2206ed-55e9-4735-b969-59d9e089bb10"
	oldB746Wyvern       = "4baa8f05-1ef8-4770-8c66-ef5f0f02a717"
	oldB746Corrupting   = "503ddd47-4b0c-4a2a-b77a-fb321044b438"
	oldB746Cannon       = "595ae6f1-e66a-4774-943a-9d5368f4a3b4"
	oldB746Convulsing   = "893c88cb-72ba-4f54-809c-bfac4d6efb74"
	oldB746Calming      = "93d20859-249f-47ae-8f6d-095228e7b6e9"
	oldB746Carnassid    = "97b386da-4bc0-4c39-a5cc-17e441c45127"
	oldB746Gliding      = "a07b8619-224d-46d0-a17a-38d04aba574e"
	oldB746Queen        = "b8376cca-ea96-478a-8e98-c4482031300a"
	oldB746Crawler      = "d47a7aa5-3b64-4ff8-bb10-e46a34d1993c"
)

func TestS82OldB746CardsAreFull(t *testing.T) {
	for _, id := range []string{oldB746Shapeshifter, oldB746Trapper, oldB746Tempting, oldB746Wyvern, oldB746Corrupting,
		oldB746Cannon, oldB746Convulsing, oldB746Calming, oldB746Carnassid, oldB746Gliding, oldB746Queen, oldB746Crawler} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s is not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB746Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, typeLine, oracle, p, tough)
}

func oldB746CanBlock(t *testing.T, g *game.Game, attacker, blocker uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		a, _ := g.LookupCardForEffect(attacker)
		b, _ := g.LookupCardForEffect(blocker)
		if r := g.BlockPairRefusalLocked(&a, &b); r.Legal() != want {
			t.Errorf("blocker %s may block = %v, want %v (%q)", b.Name, r.Legal(), want, r.Reason)
		}
	})
}

func TestS82OldB746DauthiTrapperGivesShadowUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	trapper := oldB746Push(g, me.ID, "Dauthi Trapper", "Creature — Dauthi Minion", oldB746Trapper, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, trapper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "shadow") {
		t.Fatal("the target did not gain shadow")
	}
	if c, _ := effectstest.BattlefieldCard(g, trapper); !c.Tapped {
		t.Error("the Trapper was not tapped for the cost")
	}
	// {T}: it is tapped now, so a second activation is refused.
	if err := effectstest.RavB2Activate(t, g, me.ID, trapper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Error("activated while tapped")
	}
	for i := 0; i < 40 && g.Turn.ActiveSeat == 0; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat == 0 {
		t.Fatal("the turn never ended")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "shadow") {
		t.Error("shadow outlasted the turn")
	}
}

func TestS82OldB746CarnassidTramplesAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	c := oldB746Push(g, me.ID, "Carnassid", "Creature — Beast", oldB746Carnassid, 5, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, c, "trample") {
		t.Error("Carnassid lacks trample")
	}
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, c, 0, game.ActivateAbilityParams{}, "G"); err == nil {
		t.Fatal("regenerated for {G} alone")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, c, 0, game.ActivateAbilityParams{}, "G", "G")
	effectstest.P1Destroy(t, g, c)
	if !g.Battlefield.Contains(c) {
		t.Error("the regeneration shield did not save Carnassid")
	}
}

func TestS82OldB746DuctCrawlerStopsOneCreatureFromBlockingIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crawler := oldB746Push(g, me.ID, "Duct Crawler", "Creature — Insect", oldB746Crawler, 1, 1)
	kept := effectstest.B12Creature(g, opp.ID, "Kept Back", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := effectstest.RavB2Activate(t, g, me.ID, crawler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(kept)}, "R"); err == nil {
		t.Fatal("activated for {R} alone")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, crawler, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(kept)}, "R", "R")
	oldB746CanBlock(t, g, crawler, kept, false)
	oldB746CanBlock(t, g, crawler, free, true)
}

func TestS82OldB746SliverQueenMakesSliverTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	q := oldB746Push(g, me.ID, "Sliver Queen", "Legendary Creature — Sliver", oldB746Queen, 7, 7)
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 2; i++ {
		effectstest.RavB2MustActivate(t, g, me.ID, q, 0, game.ActivateAbilityParams{}, "C", "C")
	}
	if n := effectstest.NamedOnBattlefield(g, "Sliver", me.ID); n != 2 {
		t.Fatalf("%d Sliver tokens, want 2", n)
	}
	if err := effectstest.RavB2Activate(t, g, me.ID, q, 0, game.ActivateAbilityParams{}, "C"); err == nil {
		t.Error("activated for {1}")
	}
}

func TestS82OldB746HornetCannonMakesAHornetDestroyedAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cannon := oldB746Push(g, me.ID, "Hornet Cannon", "Artifact", oldB746Cannon, 0, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, cannon, 0, game.ActivateAbilityParams{}, "C", "C", "C")
	var hornet uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Hornet" && c.Controller == me.ID {
				hornet = c.InstanceID
			}
		}
	})
	if hornet == uuid.Nil {
		t.Fatal("no Hornet token")
	}
	c, _ := effectstest.BattlefieldCard(g, hornet)
	if !c.IsToken() || !c.IsCreature() || !c.IsArtifact() || !c.HasSubtypeWord("Insect") || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("the token is %+v, want a colorless 1/1 Insect artifact creature", c)
	}
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, hornet, kw) {
			t.Errorf("the Hornet lacks %s", kw)
		}
	}
	effectstest.EndTheTurn(t, g)
	if g.Battlefield.Contains(hornet) {
		t.Error("the Hornet survived the end step")
	}
}

func TestS82OldB746SilverWyvernRedirectsASpellThatTargetsOnlyItToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wyvern := oldB746Push(g, opp.ID, "Silver Wyvern", "Creature — Drake", oldB746Wyvern, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		effectstest.CardRefs(wyvern))
	if err := effectstest.RavB2Activate(t, g, opp.ID, wyvern, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bolt)}, "U"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestRetarget(g, opp.ID)
	if p == nil {
		t.Fatal("the Wyvern's controller was not asked for a new target")
	}
	if len(p.PickTargetPlayers) != 0 || len(p.PickTargetCards) != 1 || p.PickTargetCards[0] != bear {
		t.Fatalf("offered players %v cards %v, want only the Bear (a creature, not the player)", p.PickTargetPlayers, p.PickTargetCards)
	}
	if err := g.ResolveRetarget(p.ID, opp.ID, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("a player was accepted as the new target")
	}
	if err := g.ResolveRetarget(p.ID, opp.ID, []game.TargetRef{{Kind: game.TargetCard, ID: bear}}); err != nil {
		t.Fatalf("choosing the Bear: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bolt did not hit the Bear")
	}
	if got := effectstest.DamageMarkedOn(g, wyvern); got != 0 {
		t.Errorf("the Wyvern took %d damage, want 0", got)
	}
}

func TestS82OldB746SilverWyvernNeedsASpellTargetingOnlyItAndAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wyvern := oldB746Push(g, opp.ID, "Silver Wyvern", "Creature — Drake", oldB746Wyvern, 4, 3)
	face := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	if err := effectstest.RavB2Activate(t, g, opp.ID, wyvern, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(face)}, "U"); err == nil {
		t.Fatal("targeted a spell that does not target the Wyvern")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	w2 := oldB746Push(g2, opp2.ID, "Silver Wyvern", "Creature — Drake", oldB746Wyvern, 4, 3)
	_ = me2
	bolt := effectstest.CastCatalogSpell(t, g2, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardRefs(w2))
	if err := effectstest.RavB2Activate(t, g2, opp2.ID, w2, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bolt)}, "U"); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.LatestRetarget(g2, opp2.ID) != nil {
		t.Error("asked for a new target with no other creature to choose")
	}
	if g2.Battlefield.Contains(w2) {
		t.Error("the Bolt's target changed with no creature to move to: the Wyvern should have taken the 3")
	}
}

// oldB746Licid puts a licid and a host bear under me and attaches the licid.
func oldB746Licid(t *testing.T, g *game.Game, name, oracle string, mana ...string) (licid, bear uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	licid = oldB746Push(g, me.ID, name, "Creature — Licid", oracle, 2, 2)
	bear = effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, mana...)
	l, _ := effectstest.BattlefieldCard(g, licid)
	if !l.IsAura() || l.IsCreature() || l.AttachedTo.ID != bear {
		t.Fatalf("%s is not an Aura on the Bear (aura %v, creature %v, attached %+v)", name, l.IsAura(), l.IsCreature(), l.AttachedTo)
	}
	return licid, bear
}

func oldB746EndEffect(t *testing.T, g *game.Game, licid uuid.UUID, mana string) {
	t.Helper()
	me := g.Seats[0]
	effectstest.B06AddMana(me, mana)
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	l, _ := effectstest.BattlefieldCard(g, licid)
	if l.IsAura() || !l.IsCreature() || l.IsAttached() {
		t.Errorf("after ending, %s is not an unattached creature", l.Name)
	}
}

func TestS82OldB746GlidingLicidGivesFlyingAndEnds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	licid, bear := oldB746Licid(t, g, "Gliding Licid", oldB746Gliding, "U")
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the enchanted creature lacks flying")
	}
	oldB746EndEffect(t, g, licid, "U")
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("flying outlasted the effect")
	}
}

func TestS82OldB746CorruptingLicidGivesFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	licid, bear := oldB746Licid(t, g, "Corrupting Licid", oldB746Corrupting, "B")
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Fatal("the enchanted creature lacks fear")
	}
	oldB746EndEffect(t, g, licid, "B")
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "fear") {
		t.Error("fear outlasted the effect")
	}
}

func TestS82OldB746CalmingLicidStopsTheEnchantedCreatureAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	licid, bear := oldB746Licid(t, g, "Calming Licid", oldB746Calming, "W")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err == nil {
		t.Fatal("the enchanted creature attacked")
	}
	// Ending the effect frees it (the licid is an untapped creature again).
	effectstest.B06AddMana(me, "W")
	if err := g.PerformSpecialAction(me.ID, licid, game.SpecialActionEndEffect, game.SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Errorf("the freed creature could not attack: %v", err)
	}
}

func TestS82OldB746ConvulsingLicidStopsTheEnchantedCreatureBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	licid := oldB746Push(g, me.ID, "Convulsing Licid", "Creature — Licid", oldB746Convulsing, 2, 2)
	blocker := effectstest.B12Creature(g, me.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oldB746CanBlock(t, g, attacker, blocker, true)
	effectstest.RavB2MustActivate(t, g, me.ID, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(blocker)}, "R")
	oldB746CanBlock(t, g, attacker, blocker, false)
	oldB746EndEffect(t, g, licid, "R")
	oldB746CanBlock(t, g, attacker, blocker, true)
}

func TestS82OldB746TemptingLicidMakesEveryAbleCreatureBlockTheEnchantedOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	licid := oldB746Push(g, me.ID, "Tempting Licid", "Creature — Licid", oldB746Tempting, 2, 2)
	bear := effectstest.ReqCreature(g, me.ID, "Bear", "", 2, 2, "haste")
	w1 := effectstest.ReqCreature(g, opp.ID, "Wall One", "", 0, 4)
	w2 := effectstest.ReqCreature(g, opp.ID, "Wall Two", "", 0, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, licid, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)}, "G")
	effectstest.ReqToBlockers(t, g, bear)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Wall One must block Bear if able (Tempting Licid)." {
		t.Errorf("sentence = %q", got)
	}
	if err := effectstest.ReqBlock(t, g, bear, w1, w2); err != nil {
		t.Fatalf("both walls on the Bear: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with both blocking: %v", err)
	}
}

func TestS82OldB746VolrathsShapeshifterTakesTheTopCreatureCardAndKeepsDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shifter := oldB746Push(g, me.ID, "Volrath's Shapeshifter", "Creature — Phyrexian Shapeshifter", oldB746Shapeshifter, 0, 1)
	if c, _ := effectstest.BattlefieldCard(g, shifter); c.CurrentPower() != 0 || c.CurrentToughness() != 1 {
		t.Fatalf("with no creature on top it is %d/%d, want 0/1", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c, _ := effectstest.BattlefieldCard(g, shifter); c.Name != "Dead Bear" {
		t.Errorf("name %q; the top card of the graveyard is a creature card, want its name", c.Name)
	}
	effectstest.HandSpell(me, "Card", "Sorcery", "{9}")
	before := me.Hand.Size()
	effectstest.AdvanceToMain(t, g)
	effectstest.RavB2MustActivate(t, g, me.ID, shifter, 0, game.ActivateAbilityParams{}, "C", "C")
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("the discard prompt was not opened for the activator")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != before-1 {
		t.Errorf("hand %d -> %d, want one card discarded", before, me.Hand.Size())
	}
}
