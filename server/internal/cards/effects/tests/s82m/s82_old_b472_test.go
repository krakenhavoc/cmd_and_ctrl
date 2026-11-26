package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b472_test.go — Scars of Mirrodin enters-and-dies triggers, life,
// recursion, tricks and Auras, S82 OLD batch 472 (#5577): twelve cards, each
// driven through the engine.

const (
	b472HawkIdol   = "575f032a-6ae8-4f4f-ab0b-983e7c76ad7f"
	b472Hawk       = "5d24525f-dbac-434b-bda7-ff8af55245ff"
	b472CloneShell = "94ab7d13-bcd7-419f-9e68-bd8e7ef66d42"
	b472Vigil      = "87a4caa4-cb08-4a0c-b57a-d6d8474b1f5e"
	b472Lifesmith  = "95f5c0ac-8835-4b31-863b-84f5273aac50"
	b472Blistergrb = "f5580a77-a69b-4ee8-9809-4523097cd7ac"
	b472Masticore  = "15fb3a7a-a6c3-4bef-9f72-1d04b623063e"
	b472Reservoir  = "82b327c6-5cd9-4bee-bef4-860522bded9c"
	b472Gaveleer   = "739621e6-e138-4ef9-87e8-c730b5a7f495"
	b472Painsmith  = "fbdbea75-7f80-4704-b2cf-920d00dc1273"
	b472Putrescent = "7e6fd429-f925-4ff7-9c6e-34c2c9d15d04"
	b472Reins      = "bda496f9-5362-46f2-ad4f-cd96410c3c61"
)

func TestS82OldB472EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b472HawkIdol: "Glint Hawk Idol", b472Hawk: "Glint Hawk", b472CloneShell: "Clone Shell",
		b472Vigil: "Vigil for the Lost", b472Lifesmith: "Lifesmith", b472Blistergrb: "Blistergrub",
		b472Masticore: "Molten-Tail Masticore", b472Reservoir: "Myr Reservoir", b472Gaveleer: "Goblin Gaveleer",
		b472Painsmith: "Painsmith", b472Putrescent: "Relic Putrescence", b472Reins: "Volition Reins",
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

// b472Cast casts a card from the active seat's hand (targets optional) and lets
// it resolve, leaving any prompt it asks open.
func b472Cast(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int, targets ...game.TargetRef) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	effectstest.ToMainForCost(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func b472Layered(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	return effectstest.LayeredCard(t, g, id)
}

func TestS82OldB472GlintHawkIdolBecomesABirdForW(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	idol := effectstest.B12Push(g, me.ID, "Glint Hawk Idol", "Artifact", b472HawkIdol, 0, 0)
	if c := b472Layered(t, g, idol); c.IsCreature() {
		t.Fatal("a creature before it is animated")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, idol, 0, game.ActivateAbilityParams{}, "W")
	c := b472Layered(t, g, idol)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtype("Bird") || c.CurrentPower() != 2 || c.CurrentToughness() != 2 ||
		!effectstest.HasKeywordOnBattlefield(t, g, idol, "flying") {
		t.Fatalf("animated: creature=%v artifact=%v %d/%d, want a 2/2 Bird artifact creature with flying", c.IsCreature(), c.IsArtifact(), c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DcCleanup(g)
	if c := b472Layered(t, g, idol); c.IsCreature() {
		t.Error("still a creature after the turn ended")
	}
}

func TestS82OldB472GlintHawkIdolMayAnimateWhenAnotherArtifactEnters(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		idol := effectstest.B12Push(g, me.ID, "Glint Hawk Idol", "Artifact", b472HawkIdol, 0, 0)
		b472Cast(t, g, "Trinket", "Artifact", "", 0, 0)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if got := b472Layered(t, g, idol).IsCreature(); got != yes {
			t.Errorf("yes=%v: the Idol is a creature = %v", yes, got)
		}
	}
	// The Idol's own entry does not count, and a creature that isn't an artifact
	// doesn't either.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	idol := effectstest.B12Push(g, me.ID, "Glint Hawk Idol", "Artifact", b472HawkIdol, 0, 0)
	b472Cast(t, g, "Bear", "Creature — Bear", "", 2, 2)
	if b472Layered(t, g, idol).IsCreature() {
		t.Error("a non-artifact entering animated the Idol")
	}
}

func TestS82OldB472GlintHawkReturnsAnArtifactOrIsSacrificed(t *testing.T) {
	// No artifact: sacrificed at once.
	g := effectstest.NewCatalogGame(t)
	hawk := b472Cast(t, g, "Glint Hawk", "Creature — Bird", b472Hawk, 2, 2)
	if g.Battlefield.Contains(hawk) {
		t.Error("the Hawk stayed with no artifact to return")
	}
	// An artifact: returned, the Hawk stays.
	g = effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	art := effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	hawk = b472Cast(t, g, "Glint Hawk", "Creature — Bird", b472Hawk, 2, 2)
	effectstest.AnswerChooseCards(t, g, me.ID, art)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hawk) || g.Battlefield.Contains(art) || !me.Hand.Contains(art) {
		t.Errorf("hawk=%v artifact on battlefield=%v in hand=%v, want the Hawk kept and the artifact in hand",
			g.Battlefield.Contains(hawk), g.Battlefield.Contains(art), me.Hand.Contains(art))
	}
	// Declining sacrifices it.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	art = effectstest.B12Permanent(g, me.ID, "Trinket", "Artifact")
	hawk = b472Cast(t, g, "Glint Hawk", "Creature — Bird", b472Hawk, 2, 2)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hawk) || !g.Battlefield.Contains(art) {
		t.Error("declining did not sacrifice the Hawk and keep the artifact")
	}
}

func TestS82OldB472VigilForTheLostGainsXWhenACreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Vigil for the Lost", "Enchantment", b472Vigil, 0, 0)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FillPool(me, 3)
	life := me.Life
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2 for X=2", life, me.Life)
	}
	// Another player's creature dying is not "a creature you control".
	other := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, g.Seats[1].ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PayAmountPrompt(g, me.ID) != nil {
		t.Error("an opponent's creature dying asked for a payment")
	}
}

func TestS82OldB472LifesmithGainsThreeOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.B12Push(g, me.ID, "Lifesmith", "Creature — Human Artificer", b472Lifesmith, 2, 1)
		life := me.Life
		b472Cast(t, g, "Trinket", "Artifact", "", 0, 0)
		effectstest.B06AddMana(me, "C")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if pay {
			want = 3
		}
		if me.Life != life+want {
			t.Errorf("pay=%v: life %d → %d, want +%d", pay, life, me.Life, want)
		}
	}
}

func TestS82OldB472BlistergrubDrainsEachOpponentWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	grub := effectstest.B12Push(g, me.ID, "Blistergrub", "Creature — Phyrexian Horror", b472Blistergrb, 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, grub, "swampwalk") {
		t.Error("no swampwalk")
	}
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.KillCreature(t, g, me.ID, grub)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want -= 2
		}
		if p.Life != want {
			t.Errorf("seat %v life %d, want %d", p.ID, p.Life, want)
		}
	}
}

func TestS82OldB472MoltenTailMasticoreUpkeepAndAbilities(t *testing.T) {
	// Upkeep: discard keeps it, declining sacrifices it, an empty hand sacrifices it.
	for _, tc := range []struct {
		name          string
		hand          int
		discard, kept bool
	}{{"discard", 1, true, true}, {"decline", 1, false, false}, {"empty hand", 0, false, false}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		g.WithWriteLock(func() { me.Hand.Cards = nil })
		var card uuid.UUID
		if tc.hand > 0 {
			card = effectstest.B13HandCard(me, "Spare", "Sorcery", "")
		}
		mast := effectstest.B12Push(g, me.ID, "Molten-Tail Masticore", "Artifact Creature — Masticore", b472Masticore, 4, 4)
		effectstest.AdvanceToUpkeepOf(t, g, 0)
		effectstest.PassPriorityAroundTable(t, g)
		if tc.hand > 0 {
			if tc.discard {
				effectstest.AnswerDiscard(t, g, me.ID, card)
			} else {
				effectstest.AnswerDiscard(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if got := g.Battlefield.Contains(mast); got != tc.kept {
			t.Errorf("%s: the Masticore on the battlefield = %v, want %v", tc.name, got, tc.kept)
		}
	}
	// {4}, exile a creature card: 4 damage to any target.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mast := effectstest.B12Push(g, me.ID, "Molten-Tail Masticore", "Artifact Creature — Masticore", b472Masticore, 4, 4)
	dead := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Fallen Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	life := opp.Life
	effectstest.RavB2MustActivate(t, g, me.ID, mast, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}, ExileIDs: []uuid.UUID{dead}}, "C", "C", "C", "C")
	if opp.Life != life-4 {
		t.Errorf("life %d → %d, want -4", life, opp.Life)
	}
	if !g.Exile.Contains(dead) {
		t.Error("the creature card was not exiled as a cost")
	}
	// {2}: regenerate.
	effectstest.RavB2MustActivate(t, g, me.ID, mast, 1, game.ActivateAbilityParams{}, "C", "C")
	effectstest.P1Destroy(t, g, mast)
	if !g.Battlefield.Contains(mast) {
		t.Error("the regeneration shield did not save the Masticore")
	}
}

func TestS82OldB472MyrReservoirMakesRestrictedManaAndReturnsAMyr(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	res := effectstest.PushCatalogPermanent(g, me.ID, "Myr Reservoir", "Artifact", b472Reservoir, false)
	if err := g.ActivateManaAbility(me.ID, res, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := effectstest.PoolSize(g, me.ID); got != 2 {
		t.Fatalf("the pool holds %d mana, want 2", got)
	}
	// Restricted to Myr: a non-Myr spell can't be cast with it, a Myr spell can.
	bear := b472HandCard(me, "Bear", "Creature — Bear", "{1}{1}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the restricted mana paid for a non-Myr spell")
	}
	myr := b472HandCard(me, "Tiny Myr", "Artifact Creature — Myr", "{1}{1}")
	if err := g.CastSpell(me.ID, myr, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("the restricted mana could not pay for a Myr spell: %v", err)
	}
}

func TestS82OldB472MyrReservoirReturnsAMyrCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	res := effectstest.PushCatalogPermanent(g, me.ID, "Myr Reservoir", "Artifact", b472Reservoir, false)
	myr := uuid.New()
	other := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: myr, Name: "Dead Myr", TypeLine: "Artifact Creature — Myr", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: other, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if err := effectstest.RavB2Activate(t, g, me.ID, res, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(other)}, "C", "C", "C"); err == nil {
		t.Fatal("a non-Myr card was a legal target")
	}
	effectstest.RavB2MustActivate(t, g, me.ID, res, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(myr)}, "C", "C", "C")
	if !me.Hand.Contains(myr) || me.Graveyard.Contains(myr) {
		t.Error("the Myr card did not return to its owner's hand")
	}
}

func TestS82OldB472GoblinGaveleerGrowsPerEquipment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gav := effectstest.B12Push(g, me.ID, "Goblin Gaveleer", "Creature — Goblin Warrior", b472Gaveleer, 1, 1)
	other := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, gav, "trample") {
		t.Error("no trample")
	}
	attach := func(host uuid.UUID) {
		sword := effectstest.B12Permanent(g, me.ID, "Sword", "Artifact — Equipment")
		g.WithWriteLock(func() {
			if err := g.AttachForEffect(sword, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
				t.Errorf("AttachForEffect: %v", err)
			}
		})
	}
	attach(gav)
	attach(gav)
	attach(other)
	if c := b472Layered(t, g, gav); c.CurrentPower() != 5 || c.CurrentToughness() != 1 {
		t.Errorf("with two Equipment it is %d/%d, want 5/1", c.CurrentPower(), c.CurrentToughness())
	}
	if c := b472Layered(t, g, other); c.CurrentPower() != 2 {
		t.Errorf("another creature's power is %d, want 2 (the Gaveleer pumps only itself)", c.CurrentPower())
	}
}

func TestS82OldB472PainsmithPumpsAndGivesDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.B12Push(g, me.ID, "Painsmith", "Creature — Human Artificer", b472Painsmith, 2, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b472Cast(t, g, "Trinket", "Artifact", "", 0, 0)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the cast trigger")
	}
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if c := b472Layered(t, g, bear); c.CurrentPower() != 4 || !effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Errorf("the target is %d/%d, want 4/2 with deathtouch", c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.DcCleanup(g)
	if c := b472Layered(t, g, bear); c.CurrentPower() != 2 || effectstest.HasKeywordOnBattlefield(t, g, bear, "deathtouch") {
		t.Error("the pump or deathtouch outlasted the turn")
	}
}

func TestS82OldB472RelicPutrescenceGivesItsControllerPoisonWhenTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Mind Stone", "Artifact", "", false)
	b472Cast(t, g, "Relic Putrescence", "Enchantment — Aura", b472Putrescent, 0, 0, game.TargetRef{Kind: game.TargetCard, ID: rock})
	if got := opp.Poison; got != 0 {
		t.Fatalf("poison %d before the tap", got)
	}
	g.WithWriteLock(func() {
		if err := (effects.TapTarget{Target: rock}).Apply(effects.NewContext(g, &game.StackItem{Controller: opp.ID, SourceCardID: rock})); err != nil {
			t.Errorf("tap: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Poison != 1 || me.Poison != 0 {
		t.Errorf("poison: its controller %d, the Aura's controller %d; want 1 and 0", opp.Poison, me.Poison)
	}
}

func TestS82OldB472VolitionReinsStealsAndUntaps(t *testing.T) {
	for _, tapped := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
		bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
		if tapped {
			effectstest.TapCard(g, bear)
		}
		b472Cast(t, g, "Volition Reins", "Enchantment — Aura", b472Reins, 0, 0, game.TargetRef{Kind: game.TargetCard, ID: bear})
		effectstest.PassPriorityAroundTable(t, g)
		c := b472Layered(t, g, bear)
		if c.Controller != me.ID {
			t.Errorf("tapped=%v: the creature is controlled by %v, want %v", tapped, c.Controller, me.ID)
		}
		if effectstest.FindBattlefieldCardForTest(g, bear).Tapped {
			t.Errorf("tapped=%v: the creature is still tapped", tapped)
		}
	}
}

func TestS82OldB472CloneShellImprintsAndReturnsACreatureWhenItDies(t *testing.T) {
	for _, creature := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		typeLine := "Sorcery"
		if creature {
			typeLine = "Creature — Bear"
		}
		below := b472Lib(me, "Below", "Sorcery")
		pick := b472Lib(me, "Imprinted", typeLine)
		top := []uuid.UUID{pick, b472Lib(me, "Top A", "Sorcery"), b472Lib(me, "Top B", "Sorcery"), b472Lib(me, "Top C", "Sorcery")}
		shell := b472Cast(t, g, "Clone Shell", "Artifact Creature — Shapeshifter", b472CloneShell, 2, 2)
		_ = top
		effectstest.AnswerChooseCards(t, g, me.ID, pick)
		if order := effectstest.PutInLibraryPrompt(g); order != nil {
			if err := g.ResolvePutInLibrary(order.ID, me.ID, order.ScryCards, nil); err != nil {
				t.Fatalf("order the rest: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(pick) {
			t.Fatalf("creature=%v: the picked card was not exiled", creature)
		}
		if c := b472Exiled(g, pick); !c.FaceDown {
			t.Error("the imprinted card is face up")
		}
		if me.Library.Contains(pick) {
			t.Error("the imprinted card is still in the library")
		}
		// The other three went to the bottom: the card that was below them is now on top.
		if tops := effectstest.LibraryTopIDs(me, 1); len(tops) != 1 || tops[0] != below {
			t.Errorf("library top %v, want the card that was below the four", tops)
		}
		effectstest.KillCreature(t, g, me.ID, shell)
		effectstest.PassPriorityAroundTable(t, g)
		// A card entering from exile is a new object, so it is found by name.
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Imprinted") == 1; got != creature {
			t.Errorf("creature=%v: the imprinted card is under your control on the battlefield = %v", creature, got)
		}
		if !creature {
			if !g.Exile.Contains(pick) || b472Exiled(g, pick).FaceDown {
				t.Error("a non-creature imprint should stay in exile, turned face up")
			}
		}
	}
}

// b472Lib puts a card on top of p's library.
func b472Lib(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

// b472Exiled is the exiled card with the given id.
func b472Exiled(g *game.Game, id uuid.UUID) game.Card {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	return game.Card{}
}

// b472HandCard puts a card with a mana cost into p's hand.
func b472HandCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, Power: 1, Toughness: 1, Owner: p.ID, Controller: p.ID})
	return id
}
