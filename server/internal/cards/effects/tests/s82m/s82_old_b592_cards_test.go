package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b592_cards_test.go — Saviors of Kamigawa, batch 592 (#5601). Each
// test plays the real catalog entry and asserts what its text says.

const (
	b592AdamaroOracle    = "18f34bfd-7675-49df-91ea-695c510cd3c3"
	b592MasumaroOracle   = "8af347b8-89fa-499e-8e5a-77463edd6773"
	b592ShapeOracle      = "fa9a6617-fd89-490e-9ca5-8e6da99d89ed"
	b592SpellbladeOracle = "26e47cc5-9dc7-4d54-80a6-1eb77587df36"
	b592IizukaOracle     = "4f40d66a-1fd8-4b06-b6cb-9cbcf33bc041"
	b592RenegadeOracle   = "63f414ed-01f4-428a-b65f-5c640679d5ef"
	b592CowedOracle      = "10495c22-2456-4b93-afa9-0c3c53075efe"
	b592ClutchOracle     = "f83f6ba0-bb23-4c62-8a16-b17f5fbc0740"
	b592CutOracle        = "127ce189-2758-4789-9e02-4aa71d013431"
	b592WillOracle       = "983a2a50-3634-4783-8576-30916c3d0818"
	b592SekkiOracle      = "1b92501c-80fe-4ae7-834e-72f57ba41a6a"
	b592SeedOracle       = "75b39a67-6720-482f-935a-c142ca0f7891"
)

func TestOldB592EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		b592AdamaroOracle:    "Adamaro, First to Desire",
		b592MasumaroOracle:   "Masumaro, First to Live",
		b592ShapeOracle:      "Shape Stealer",
		b592SpellbladeOracle: "Sokenzan Spellblade",
		b592IizukaOracle:     "Iizuka the Ruthless",
		b592RenegadeOracle:   "Sokenzan Renegade",
		b592CowedOracle:      "Cowed by Wisdom",
		b592ClutchOracle:     "Kagemaro's Clutch",
		b592CutOracle:        "Cut the Earthly Bond",
		b592WillOracle:       "Oppressive Will",
		b592SekkiOracle:      "Sekki, Seasons' Guide",
		b592SeedOracle:       "Seed the Land",
	}
	for oracle, name := range want {
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

func b592Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func b592Hand(p *game.Player, names ...string) {
	p.Hand.Cards = nil
	effectstest.FillHand(p, names...)
}

func b592Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	return effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, targets)
}

func b592Attach(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := effectstest.PushCatalogPermanent(g, owner, name, effectstest.AuraTypeLine, oracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach %s: %v", name, err)
		}
	})
	return aura
}

func TestOldB592AdamaroIsAsBigAsTheLargestOpposingHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	effectstest.AdvanceToMain(t, g)
	third := g.Seats[2]
	for _, p := range g.Seats {
		p.Hand.Cards = nil
	}
	b592Hand(me, "M1", "M2", "M3", "M4", "M5", "M6", "M7", "M8", "M9")
	b592Hand(opp, "A", "B", "C")
	b592Hand(third, "D", "E", "F", "G", "H")
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Adamaro, First to Desire", TypeLine: "Legendary Creature — Spirit",
		OracleID: b592AdamaroOracle, Owner: me.ID, Controller: me.ID,
	})
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 5 || tough != 5 {
		t.Errorf("Adamaro is %d/%d, want 5/5 (the biggest opposing hand, not yours)", p, tough)
	}
	// It follows an opponent's hand live: a draw and nothing else.
	if err := g.DrawCard(third.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != 6 {
		t.Errorf("Adamaro has power %d after that opponent drew, want 6", p)
	}
}

func TestOldB592MasumaroIsTwiceYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b592Seats(g)
	effectstest.AdvanceToMain(t, g)
	b592Hand(me, "A", "B", "C")
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Masumaro, First to Live", TypeLine: "Legendary Creature — Spirit",
		OracleID: b592MasumaroOracle, Owner: me.ID, Controller: me.ID,
	})
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 6 || tough != 6 {
		t.Errorf("Masumaro is %d/%d with three cards in hand, want 6/6", p, tough)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, id); p != 8 || tough != 8 {
		t.Errorf("Masumaro is %d/%d after a draw, want 8/8", p, tough)
	}
}

func TestOldB592ShapeStealerTakesTheOtherCreaturesPowerAndToughness(t *testing.T) {
	// Blocking: it copies the attacker's.
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 5)
	stealer := effectstest.B12Push(g, opp.ID, "Shape Stealer", "Creature — Shapeshifter Spirit", b592ShapeOracle, 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, ogre)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(stealer, ogre); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, stealer); p != 4 || tough != 5 {
		t.Errorf("a blocking Shape Stealer is %d/%d, want 4/5", p, tough)
	}

	// Blocked: it copies the blocker's.
	g = effectstest.NewCatalogGame(t)
	me, opp = b592Seats(g)
	stealer = effectstest.B12Push(g, me.ID, "Shape Stealer", "Creature — Shapeshifter Spirit", b592ShapeOracle, 1, 1)
	bear := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 7)
	effectstest.DeclareAttack(t, g, opp.ID, stealer)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(bear, stealer); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, stealer); p != 0 || tough != 7 {
		t.Errorf("a blocked Shape Stealer is %d/%d, want 0/7", p, tough)
	}
	effectstest.DcCleanup(g)
	if p, tough := effectstest.EffectivePTOf(t, g, stealer); p != 1 || tough != 1 {
		t.Errorf("the change outlived the turn (%d/%d)", p, tough)
	}
}

func TestOldB592SokenzanSpellbladeHasBushidoAndPumpsByHandSize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Sokenzan Spellblade", "Creature — Ogre Samurai Shaman", b592SpellbladeOracle, false)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesBlocked, Actor: opp.ID, Source: id, CardID: id, Target: id})
	})
	effectstest.PassPriorityAroundTable(t, g)
	base, _ := effectstest.EffectivePTOf(t, g, id)
	b592Hand(me, "A", "B", "C")
	effectstest.C2Pay(t, g, me, "{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, id); p != base+3 {
		t.Errorf("power %d → %d, want +3 (three cards in hand)", base, p)
	}
}

func TestOldB592IizukaGivesSamuraiDoubleStrikeForASacrificedSamurai(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	iizuka := effectstest.PushCatalogPermanent(g, me.ID, "Iizuka the Ruthless", "Legendary Creature — Human Samurai", b592IizukaOracle, false)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Human Samurai", 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Other Samurai", "Creature — Fox Samurai", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Samurai", "Creature — Human Samurai", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.C2Pay(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, iizuka, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("sacrificed a non-Samurai to Iizuka")
	}
	if err := g.ActivateCatalogAbility(me.ID, iizuka, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the Samurai was not sacrificed")
	}
	for _, id := range []uuid.UUID{iizuka, other} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") {
			t.Error("a Samurai you control lacks double strike")
		}
	}
	for _, id := range []uuid.UUID{bear, theirs} {
		if effectstest.HasKeywordOnBattlefield(t, g, id, "double strike") {
			t.Error("a creature that is not your Samurai has double strike")
		}
	}
}

func TestOldB592SokenzanRenegadeGoesToThePlayerWithTheMostCards(t *testing.T) {
	for _, tc := range []struct {
		name         string
		mine, theirs int
		wantTheirs   bool
	}{
		{"an opponent has more", 2, 5, true},
		{"a tie at the top", 5, 5, false},
		{"you have the most", 5, 2, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := b592Seats(g)
		seat := g.Turn.ActiveSeat
		id := effectstest.PushCatalogPermanent(g, me.ID, "Sokenzan Renegade", "Creature — Ogre Samurai Mercenary", b592RenegadeOracle, false)
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
		names := func(n int) []string {
			out := make([]string, n)
			for i := range out {
				out[i] = "X"
			}
			return out
		}
		// Every other player draws once on the way round to my next upkeep
		// (mine comes after it), so the opponent starts one card short.
		effectstest.FillHand(me, names(tc.mine)...)
		effectstest.FillHand(opp, names(tc.theirs-1)...)
		effectstest.AdvanceToUpkeepOfSeat(t, g, seat)
		if me.Hand.Size() != tc.mine || opp.Hand.Size() != tc.theirs {
			t.Fatalf("%s: hands are %d and %d at the upkeep, want %d and %d", tc.name, me.Hand.Size(), opp.Hand.Size(), tc.mine, tc.theirs)
		}
		effectstest.PassPriorityAroundTable(t, g)
		got := effectstest.ControllerOf(t, g, id)
		if tc.wantTheirs && got != opp.ID {
			t.Errorf("%s: the Renegade stayed with its controller", tc.name)
		}
		if !tc.wantTheirs && got != me.ID {
			t.Errorf("%s: the Renegade changed hands", tc.name)
		}
	}
}

func TestOldB592CowedByWisdomTaxesAttackingAndBlockingByYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, opp.ID, "Free", "Creature — Bear", 2, 2)
	b592Hand(me, "A", "B", "C")
	b592Attach(t, g, me.ID, "Cowed by Wisdom", b592CowedOracle, bear)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: me.ID}}); p.Total.Generic != 3 {
		t.Errorf("attacking with the enchanted creature costs %d, want 3 (cards in the Aura controller's hand)", p.Total.Generic)
	}
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: free, Target: me.ID}}); !p.IsFree() {
		t.Errorf("another creature's attack costs %+v, want nothing", p.Total)
	}
	var blockPrice game.BlockTaxPrice
	g.WithWriteLock(func() { blockPrice = g.PriceBlockersForEffect([]uuid.UUID{bear}) })
	if blockPrice.Total.Generic != 3 {
		t.Errorf("blocking with the enchanted creature costs %d, want 3", blockPrice.Total.Generic)
	}
	// The count follows the hand.
	b592Hand(me)
	if p := g.PriceAttackDeclaration([]game.AttackDeclaration{{Attacker: bear, Target: me.ID}}); !p.IsFree() {
		t.Errorf("with an empty hand the attack costs %+v, want nothing", p.Total)
	}
}

func TestOldB592KagemarosClutchShrinksByYourHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	effectstest.AdvanceToMain(t, g)
	b592Hand(me, "A", "B")
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 6, 6)
	b592Attach(t, g, me.ID, "Kagemaro's Clutch", b592ClutchOracle, ogre)
	if p, tough := effectstest.EffectivePTOf(t, g, ogre); p != 4 || tough != 4 {
		t.Errorf("the Ogre is %d/%d, want 4/4", p, tough)
	}
	if err := g.DrawCard(me.ID); err != nil {
		t.Fatalf("draw: %v", err)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, ogre); p != 3 || tough != 3 {
		t.Errorf("the Ogre is %d/%d after a draw, want 3/3", p, tough)
	}
}

func TestOldB592CutTheEarthlyBondBouncesOnlyEnchantedPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	enchanted := effectstest.B12Creature(g, opp.ID, "Enchanted", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, opp.ID, "Plain", "Creature — Bear", 2, 2)
	b592Attach(t, g, me.ID, "Some Aura", "", enchanted)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Cut the Earthly Bond", "Instant — Arcane", b592CutOracle, effectstest.CardTarget(plain)); err == nil {
		t.Error("targeted a permanent with no Aura on it")
	}
	b592Cast(t, g, "Cut the Earthly Bond", "Instant — Arcane", b592CutOracle, effectstest.CardTarget(enchanted))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(enchanted) || !opp.Hand.Contains(enchanted) {
		t.Error("the enchanted creature was not returned to its owner's hand")
	}
}

func TestOldB592OppressiveWillCountersUnlessYouPayPerCardInHand(t *testing.T) {
	board := func(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
		t.Helper()
		g = effectstest.NewCatalogGame(t)
		caster, them := g.Seats[0], g.Seats[1]
		effectstest.ToMainForCost(t, g)
		victim = effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
		b592Hand(them, "A", "B")
		will := effectstest.HandCardFull(them, "Oppressive Will", "Instant", "", b592WillOracle, []string{"U"})
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority to the responder: %v", err)
		}
		if err := g.CastSpell(them.ID, will, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
			t.Fatalf("casting Oppressive Will: %v", err)
		}
		effectstest.PassUntilTaxed(t, g, caster.ID)
		return g, victim, caster
	}
	g, victim, caster := board(t)
	prompt := effectstest.AnswerPayUnless(t, g, caster.ID, false)
	if prompt.PayCost != "{2}" {
		t.Errorf("the tax is %q, want {2} (two cards in the other hand)", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}
	g, victim, caster = board(t)
	effectstest.FillPool(caster, 2)
	effectstest.AnswerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

func TestOldB592OppressiveWillWithAnEmptyHandCountersNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	them.Hand.Cards = nil
	will := effectstest.HandCardFull(them, "Oppressive Will", "Instant", "", b592WillOracle, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(them.ID, will, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 6 && g.Stack.Contains(will); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if effectstest.HasPayUnlessFor(g, caster.ID) {
		t.Error("asked for a payment of nothing")
	}
	if !g.Stack.Contains(victim) {
		t.Error("the spell was countered with an empty hand")
	}
}

func TestOldB592SekkiEntersWithEightCountersAndTurnsDamageIntoSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastCatalogSpell(t, g, "Sekki, Seasons' Guide", "Legendary Creature — Spirit", b592SekkiOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Sekki is not on the battlefield")
	}
	if n := effectstest.Pr8Counters(g, id, game.CounterPlusOne); n != 8 {
		t.Fatalf("Sekki has %d +1/+1 counters, want 8", n)
	}
	if p, tough := effectstest.PtOf(t, g, id); p != 8 || tough != 8 {
		t.Errorf("Sekki is %d/%d, want 8/8", p, tough)
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(opp.ID, id, 3); err != nil {
			t.Fatalf("damage: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, id); got != 0 {
		t.Errorf("Sekki has %d damage, want it all prevented", got)
	}
	if n := effectstest.Pr8Counters(g, id, game.CounterPlusOne); n != 5 {
		t.Errorf("Sekki has %d counters after 3 damage, want 5", n)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Spirit"); n != 3 {
		t.Errorf("%d Spirit tokens, want 3", n)
	}
}

func TestOldB592SekkiReturnsFromTheGraveyardForEightSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := b592Seats(g)
	effectstest.AdvanceToMain(t, g)
	sekki := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sekki, Name: "Sekki, Seasons' Guide", TypeLine: "Legendary Creature — Spirit",
		OracleID: b592SekkiOracle, Owner: me.ID, Controller: me.ID})
	var spirits []uuid.UUID
	for i := 0; i < 8; i++ {
		spirits = append(spirits, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Spirit", TypeLine: "Token Creature — Spirit", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
		}))
	}
	if err := g.ActivateCatalogAbility(me.ID, sekki, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: spirits[:7]}); err == nil {
		t.Fatal("activated with seven Spirits")
	}
	if err := g.ActivateCatalogAbility(me.ID, sekki, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: spirits}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(sekki) {
		t.Error("Sekki did not return to the battlefield")
	}
	for _, id := range spirits {
		if g.Battlefield.Contains(id) {
			t.Error("a Spirit was not sacrificed")
		}
	}
	if n := effectstest.Pr8Counters(g, sekki, game.CounterPlusOne); n != 8 {
		t.Errorf("the returned Sekki has %d counters, want 8", n)
	}
}

func TestOldB592SeedTheLandGivesTheLandsControllerASnake(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := b592Seats(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Seed the Land", "Enchantment", b592SeedOracle, false)
	effectstest.AdvanceToMain(t, g)
	// An opponent's land entering gives the opponent the Snake.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventETB, CardID: effectstest.AangPushLand(g, opp.ID, "Forest", false), Actor: opp.ID})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, opp.ID, "Snake"); n != 1 {
		t.Errorf("the opponent controls %d Snakes, want 1", n)
	}
	if n := effectstest.CountTokensControlled(g, me.ID, "Snake"); n != 0 {
		t.Errorf("you control %d Snakes, want 0", n)
	}
	// Your own land, played for real.
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountTokensControlled(g, me.ID, "Snake"); n != 1 {
		t.Errorf("you control %d Snakes after playing a land, want 1", n)
	}
}
