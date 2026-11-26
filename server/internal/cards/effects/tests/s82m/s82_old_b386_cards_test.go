package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 older sets, Commander 2013, batch 386 (#5560).

const (
	oldB386FamineOracle        = "1f82fe6b-1ec7-400b-a80d-40921b359723"
	oldB386FellShepherdOracle  = "5fd78088-53db-453b-90a3-b8426b0a8269"
	oldB386SuddenDemiseOracle  = "b34b5b3f-7f17-4292-814e-634408a5d7a5"
	oldB386AuguryAdeptOracle   = "be5a65fd-0d06-4771-bee7-0e42cc9871da"
	oldB386BladetrapOracle     = "ca21d597-40a5-4724-9d53-1ab6e6c6e767"
	oldB386MysticBarrierOracle = "0caf42f5-abff-48aa-9bbf-df6cba169ef3"
	oldB386JelevaOracle        = "a014f283-c531-415c-ac00-e6773ea5d64d"
	oldB386ReincarnationOracle = "d6bf5e22-8d33-43a9-8824-435068e0a87a"
	oldB386DivinityOracle      = "2c91c236-34d7-4454-a55a-784db7f68bde"
	oldB386WidespreadOracle    = "853a3c2b-3d37-453a-8a77-4d90bd3a1cb7"
	oldB386ArcaneMeleeOracle   = "9967b292-1239-4aed-b871-72e40d1f3e26"
)

func TestOldB386EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB386FamineOracle:        "Famine",
		oldB386FellShepherdOracle:  "Fell Shepherd",
		oldB386SuddenDemiseOracle:  "Sudden Demise",
		oldB386AuguryAdeptOracle:   "Augury Adept",
		oldB386BladetrapOracle:     "Leonin Bladetrap",
		oldB386MysticBarrierOracle: "Mystic Barrier",
		oldB386JelevaOracle:        "Jeleva, Nephalia's Scourge",
		oldB386ReincarnationOracle: "Reincarnation",
		oldB386DivinityOracle:      "Divinity of Pride",
		oldB386WidespreadOracle:    "Widespread Panic",
		oldB386ArcaneMeleeOracle:   "Arcane Melee",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q completeness %v caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB386FamineHitsEveryCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine, theirs := effectstest.SeedBear(g, me.ID), effectstest.SeedBear(g, opp.ID)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}
	effectstest.CastCatalogSpell(t, g, "Famine", "Sorcery", oldB386FamineOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a 2/2 survived 3 damage")
	}
	for _, p := range g.Seats {
		if p.Life != before[p.ID]-3 {
			t.Errorf("%s life = %d, want %d", p.ID, p.Life, before[p.ID]-3)
		}
	}
}

func TestOldB386SuddenDemiseHitsOnlyTheChosenColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	red := effectstest.SeedColoredCreature(g, me.ID, "Red Bear", []string{"R"}, 0)
	green := effectstest.SeedColoredCreature(g, me.ID, "Green Bear", []string{"G"}, 0)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sudden Demise", TypeLine: "Sorcery", OracleID: oldB386SuddenDemiseOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) {
		t.Error("the red creature survived 2 damage")
	}
	if !g.Battlefield.Contains(green) {
		t.Error("the green creature died")
	}
}

func TestOldB386LeoninBladetrapHitsOnlyAttackersWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trap := effectstest.B12Push(g, me.ID, "Leonin Bladetrap", "Artifact", oldB386BladetrapOracle, 0, 0)
	walker := effectstest.SeedBear(g, g.Seats[1].ID)
	flyer := effectstest.PushPermanentForTest(g, g.Seats[1].ID, "Bird", "", "Creature — Bird")
	idle := effectstest.SeedBear(g, g.Seats[1].ID)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			switch c.InstanceID {
			case walker:
				c.AttackingTarget = me.ID
			case flyer:
				c.Power, c.Toughness = 2, 2
				c.Keywords = []string{"flying"}
				c.AttackingTarget = me.ID
			}
		}
	})
	if err := effectstest.ConB1Activate(t, g, me, trap, 0, "CC", game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the attacking ground creature survived 2 damage")
	}
	if !g.Battlefield.Contains(flyer) || !g.Battlefield.Contains(idle) {
		t.Error("a flyer or a non-attacker was hurt")
	}
	if g.Battlefield.Contains(trap) {
		t.Error("the Bladetrap was not sacrificed")
	}
}

func TestOldB386DivinityOfPrideGrowsWithLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	div := effectstest.B12Push(g, me.ID, "Divinity of Pride", "Creature — Spirit Avatar", oldB386DivinityOracle, 4, 4)
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 24-me.Life) })
	if p, tough := effectstest.PtOf(t, g, div); p != 4 || tough != 4 {
		t.Errorf("at 24 life it is %d/%d, want 4/4", p, tough)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	if p, tough := effectstest.PtOf(t, g, div); p != 8 || tough != 8 {
		t.Errorf("at 25 life it is %d/%d, want 8/8", p, tough)
	}
	for _, kw := range []string{"flying", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, div, kw) {
			t.Errorf("lacks %s", kw)
		}
	}
}

func TestOldB386AuguryAdeptRevealsAndGainsLifeEqualToManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	adept := effectstest.B12Push(g, me.ID, "Augury Adept", "Creature — Kithkin Wizard", oldB386AuguryAdeptOracle, 2, 2)
	top := effectstest.B10LibraryTop(me, "Three Drop", "Creature — Elk", "{2}{G}", 2, 2)
	life := me.Life
	effectstest.DealCombatDamageToPlayer(g, adept, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the revealed card did not go to hand")
	}
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d", me.Life, life+3)
	}
}

func TestOldB386FellShepherdReturnsCreaturesPutIntoTheGraveyardFromTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shep := effectstest.B12Push(g, me.ID, "Fell Shepherd", "Creature — Avatar", oldB386FellShepherdOracle, 8, 6)
	dead := effectstest.SeedBear(g, me.ID)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(dead) })
	old := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if !me.Graveyard.Contains(dead) {
		t.Fatal("the destroyed bear is not in the graveyard")
	}
	effectstest.DealCombatDamageToPlayer(g, shep, opp.ID, 8)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the creature that died this turn was not returned")
	}
	if me.Hand.Contains(old) {
		t.Error("a card that was already in the graveyard was returned")
	}
}

func TestOldB386FellShepherdSacrificesForMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shep := effectstest.B12Push(g, me.ID, "Fell Shepherd", "Creature — Avatar", oldB386FellShepherdOracle, 8, 6)
	fodder := effectstest.SeedBear(g, me.ID)
	victim := effectstest.SeedBear(g, g.Seats[1].ID)
	if err := effectstest.ConB1Activate(t, g, me, shep, 0, "B", game.ActivateAbilityParams{
		Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{shep},
	}); err == nil {
		t.Fatal("it sacrificed itself")
	}
	if err := effectstest.ConB1Activate(t, g, me, shep, 0, "B", game.ActivateAbilityParams{
		Targets: effectstest.CardTarget(victim), SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || g.Battlefield.Contains(fodder) {
		t.Error("the 2/2 target or the sacrificed creature is still here")
	}
}

func TestOldB386ReincarnationReturnsACreatureCardToItsOwner(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := effectstest.SeedBear(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Reincarnation", "Instant", oldB386ReincarnationOracle, effectstest.CardTarget(victim))
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPromptFrom(g, g.Seats[0].ID)
	if prompt == nil {
		for _, p := range g.PendingChoices {
			if p != nil && p.Kind == game.PendingChoiceChooseCards {
				prompt = p
			}
		}
	}
	if prompt == nil {
		t.Fatal("no creature card pick was offered")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{victim}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(victim) || effectstest.ControllerOf(t, g, victim) != opp.ID {
		t.Error("the creature card did not return under its owner's control")
	}
}

func TestOldB386MysticBarrierChoosesOnEnteringAndEachUpkeep(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	barrier := effectstest.CastCatalogSpell(t, g, "Mystic Barrier", "Enchantment", oldB386MysticBarrierOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnchorWord(t, g, me.ID, "Right")
	if got := g.ChosenOptionOf(barrier); got != "Right" {
		t.Fatalf("chosen = %q, want Right", got)
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerAnchorWord(t, g, me.ID, "Left")
	if got := g.ChosenOptionOf(barrier); got != "Left" {
		t.Errorf("after the upkeep choice = %q, want Left (the last chosen direction)", got)
	}
}

func TestOldB386WidespreadPanicMakesTheShufflerPutACardOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Widespread Panic", "Enchantment", oldB386WidespreadOracle, false)
	hand := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: hand, Name: "Only Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	for _, c := range me.Hand.Cards {
		if c.InstanceID != hand {
			_ = c
		}
	}
	g.WithWriteLock(func() { me.Hand.Cards = me.Hand.Cards[len(me.Hand.Cards)-1:] })
	g.WithWriteLock(func() { _ = g.ShuffleLibraryForEffect(me.ID) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(hand) {
		t.Error("the card stayed in hand")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != hand {
		t.Error("the hand card is not on top of the library")
	}
}

func TestOldB386ArcaneMeleeReducesInstantsAndSorceriesForEveryone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Arcane Melee", "Enchantment", oldB386ArcaneMeleeOracle, false)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "R")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("a {2}{R} sorcery should cost {R}: %v", err)
	}
}

func TestOldB386JelevaExilesXFromEachLibraryAndCastsFromThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	spell := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: spell, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	mineBefore, theirsBefore := len(me.Library.Cards), len(opp.Library.Cards)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Jeleva, Nephalia's Scourge", TypeLine: "Legendary Creature — Vampire Wizard", OracleID: oldB386JelevaOracle,
		ManaCost: "{1}{U}{B}{R}", Power: 1, Toughness: 3, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "CUBR")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := mineBefore - len(me.Library.Cards); got != 4 {
		t.Errorf("my library lost %d cards, want 4", got)
	}
	if got := theirsBefore - len(opp.Library.Cards); got != 4 {
		t.Errorf("their library lost %d cards, want 4", got)
	}
	var linked int
	g.WithWriteLock(func() {
		ref, _ := g.PermanentRefForEffect(id)
		linked = len(g.ExiledWithForEffect(game.PermissionCardRef(ref)))
	})
	if want := 4 * len(g.Seats); linked != want {
		t.Errorf("%d cards are exiled with Jeleva, want %d", linked, want)
	}
}

func TestOldB386JelevaOffersAFreeCastOfAnExiledInstantOrSorceryWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jeleva := effectstest.B12Push(g, me.ID, "Jeleva, Nephalia's Scourge", "Legendary Creature — Vampire Wizard", oldB386JelevaOracle, 1, 3)
	sorcery := uuid.New()
	elk := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: elk, Name: "Elk", TypeLine: "Creature — Elk", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: sorcery, Name: "Test Sorcery", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	g.WithWriteLock(func() {
		ref, _ := g.PermanentRefForEffect(jeleva)
		_ = g.ExileLinkedForEffect(ref, sorcery)
		_ = g.ExileLinkedForEffect(ref, elk)
	})
	effectstest.AdvanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(jeleva, g.Seats[1].ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityUntilChoice(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, sorcery)
	for _, id := range offer.DiscardOptions {
		if id == elk {
			t.Error("a creature card was offered")
		}
	}
}
