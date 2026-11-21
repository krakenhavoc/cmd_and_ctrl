package testsi

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsth "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/th"
	cardsu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/u"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyed_take_attack_restore_test.go — #4253 slice 8: the look-and-take
// callers, the last prompt of the graveyard-return chains and the
// attack-target picks are keyed, so a table waiting on any of them is a
// restore point. Each test takes a STRICT restore (restoreRoundTrip with
// strict, which fails on a table that is not a restore point) while the
// prompt is open and finishes on the restored table. One caller stands for
// each shared key.

// openKeyedPick returns the open card-set prompt of `chooser`, failing unless
// its continuation is the keyed one named `key`.
func openKeyedPick(t *testing.T, g *game.Game, chooser uuid.UUID, key string) *game.PendingChoice {
	t.Helper()
	pick := effectstest.LatestChooseCardsFor(g, chooser)
	if pick == nil {
		t.Fatal("no card-set prompt is open")
	}
	if pick.PickThen != key {
		t.Fatalf("the prompt's continuation is %q, want the keyed %q", pick.PickThen, key)
	}
	return pick
}

// refuseUnknownPickKey rewrites every card-set prompt naming `key` to an
// unregistered key and checks the restore is refused.
func refuseUnknownPickKey(t *testing.T, g *game.Game, key string) {
	t.Helper()
	snap := g.CaptureSnapshot()
	found := false
	for i := range snap.PendingChoices {
		if snap.PendingChoices[i].PickThen == key {
			snap.PendingChoices[i].PickThen = "choose-cards/never-registered"
			found = true
		}
	}
	if !found {
		t.Fatalf("the snapshot holds no prompt naming %q", key)
	}
	if _, err := snap.Restore(); !errors.Is(err, game.ErrUnknownEffectKey) {
		t.Fatalf("restore with an unknown key: err = %v, want ErrUnknownEffectKey", err)
	}
}

// --- the shared takes ---------------------------------------------------

// Stock Up (takeRestBottomAny): take two of five, the rest ordered onto the
// bottom.
func TestStockUpTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "Sixth")
	effectstest.CastCatalogSpell(t, g, "Stock Up", "Sorcery", effectstest.S58nStockUp, nil)
	effectstest.PassPriorityAroundTable(t, g)
	openKeyedPick(t, g, me.ID, effects.TakeRestBottomAny.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1], ids[3])
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[3]) {
		t.Fatal("the two picked cards are not in hand after the restore")
	}
	order := effectstest.PutInLibraryChoiceFor(g, me.ID)
	if order == nil || order.LibraryPlacement != game.LibraryPlaceBottom || len(order.ScryCards) != 3 {
		t.Fatalf("the other three are not ordered onto the bottom: %+v", order)
	}
}

func TestStockUpTakeWithAnUnknownKeyIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "Sixth")
	effectstest.CastCatalogSpell(t, g, "Stock Up", "Sorcery", effectstest.S58nStockUp, nil)
	effectstest.PassPriorityAroundTable(t, g)
	refuseUnknownPickKey(t, g, effects.TakeRestBottomAny.Key())
}

// Memory Deluge (takeRestBottomRandom): take two, the rest go under in a
// random order with no further prompt.
func TestMemoryDelugeTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedLibrary(me, "1", "2", "3", "4", "5", "6", "7", "8", "9", "10")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UUUU")
	effectstest.CastFromHandForTest(t, g, me, "Memory Deluge", "Instant", "{2}{U}{U}", effectstest.S58nMemoryDeluge, effectstest.Strict)
	effectstest.PassPriorityAroundTable(t, g)
	pick := openKeyedPick(t, g, me.ID, effects.TakeRestBottomRandom.Key())
	picked := append([]uuid.UUID(nil), pick.ChooseCards[:2]...)
	libBefore := me.Library.Size()

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerChooseCards(t, g, me.ID, picked...)
	for _, id := range picked {
		if !me.Hand.Contains(id) {
			t.Errorf("%s is not in hand after the restore", id)
		}
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) != nil {
		t.Error("the rest go to the bottom in a random order, with no ordering prompt")
	}
	if me.Library.Size() != libBefore-2 {
		t.Errorf("library = %d, want %d (two taken, the rest went under)", me.Library.Size(), libBefore-2)
	}
}

// Monumental Henge (takeRevealRestBottomRandom): the historic card taken is
// revealed, the rest go under in a random order.
func TestMonumentalHengeTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	henge := effectstest.SeedPermanentWithOracle(g, me.ID, "Monumental Henge", "Land", effectstest.SamiMonumentalHengeOracle)
	ids := effectstest.SamiLibraryTopFirst(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Wurm", TypeLine: "Artifact Creature — Wurm"},
		game.Card{Name: "Lone Hero", TypeLine: "Legendary Creature — Human"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Sixth", TypeLine: "Instant"},
	)
	effectstest.SamiMainPhase(t, g)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{W}{W}"); err != nil {
		t.Fatalf("AddManaForEffect: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, henge, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	openKeyedPick(t, g, me.ID, effects.TakeRevealRestBottomRandom.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerChooseCards(t, g, me.ID, ids[3])
	if !me.Hand.Contains(ids[3]) {
		t.Fatal("the chosen legendary did not reach the hand after the restore")
	}
	if top := me.Library.Cards[me.Library.Size()-1].InstanceID; top != ids[5] {
		t.Error("the sixth card must now be on top; the other four went under it")
	}
}

// Oath of Nissa (takeRevealRestBottomAny): the creature, land or
// planeswalker taken is revealed, the rest are ordered onto the bottom.
func TestOathOfNissaTakeIsARestorePoint(t *testing.T) {
	g, me, _ := effectstest.SpendTable(t)
	top := make([]uuid.UUID, 3)
	for i, c := range []struct{ Name, TypeLine string }{
		{"Garruk Relentless", "Legendary Planeswalker — Garruk"},
		{"Shock", "Instant"},
		{"Forest", "Basic Land — Forest"},
	} {
		top[i] = uuid.New()
		me.Library.PushTop(game.Card{InstanceID: top[i], Name: c.Name, TypeLine: c.TypeLine, Owner: me.ID, Controller: me.ID})
	}
	// PushTop leaves the last pushed on top: the Forest.
	effectstest.CastCatalogSpell(t, g, "Oath of Nissa", "Legendary Enchantment", effectstest.OathOfNissaOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	openKeyedPick(t, g, me.ID, effects.TakeRevealRestBottomAny.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerChooseCards(t, g, me.ID, top[2])
	if !me.Hand.Contains(top[2]) {
		t.Fatal("the Forest is not in hand after the restore")
	}
	if effectstest.PutInLibraryChoiceFor(g, me.ID) == nil {
		t.Fatal("the rest are not ordered onto the bottom")
	}
}

// Wail of the Forgotten's look (takeRestGraveyard): one into hand, the rest
// into the graveyard.
func TestWailOfTheForgottenTakeIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.CastCatalogSpellWithModes(t, g, "Wail of the Forgotten", "Sorcery", effectstest.WailOfTheForgottenOracle, []int{2})
	effectstest.PassPriorityAroundTable(t, g)
	openKeyedPick(t, g, me.ID, effects.TakeRestGraveyard.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	g.SettleResolution()
	if !me.Hand.Contains(ids[2]) {
		t.Fatal("the chosen card is not in hand after the restore")
	}
	if !me.Graveyard.Contains(ids[0]) || !me.Graveyard.Contains(ids[1]) {
		t.Error("the rest of the look did not reach the graveyard")
	}
}

// --- the graveyard-return chains' last prompt ----------------------------

// Rise of the Witch-king: each player sacrifices a creature and, if you
// did, you may return another permanent card. The return pick is a keyed
// card-set prompt, a restore point of its own.
func TestRiseOfTheWitchKingReturnPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushVanillaCreature(g, me.ID, "Mine", 1, 1)
	effectstest.PushVanillaCreature(g, me.ID, "Mine Too", 1, 1)
	for _, p := range g.Seats[1:] {
		effectstest.PushVanillaCreature(g, p.ID, p.Name+" A", 1, 1)
		effectstest.PushVanillaCreature(g, p.ID, p.Name+" B", 1, 1)
	}
	gone := effectstest.PushGraveyardCardForTest(me, "Gone Land")
	g.WithWriteLock(func() {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == gone {
				c.TypeLine = "Land"
			}
		}
	})
	effectstest.CastCatalogSpell(t, g, "Rise of the Witch-king", "Sorcery", "3c86541c-3601-4a38-8872-39705e41303a", nil)
	effectstest.PassPriorityAroundTable(t, g)
	// Answer every sacrifice leg until the return pick (the one that names
	// the shared return key) is the open prompt.
	for range g.Seats {
		if c := effectstest.LatestChooseCardsFor(g, me.ID); c != nil && c.PickThen == effects.ReturnChosenFromGraveyardThen.Key() {
			break
		}
		answered := false
		for _, p := range effectstest.ApnapSeats(g) {
			if c := effectstest.ChooseCardsChoiceFor(g, p.ID); c != nil && c.PickThen != effects.ReturnChosenFromGraveyardThen.Key() && len(c.ChooseCards) > 0 {
				if err := g.ResolveChooseCards(c.ID, p.ID, []uuid.UUID{c.ChooseCards[0]}); err != nil {
					t.Fatalf("answer: %v", err)
				}
				answered = true
				break
			}
		}
		if !answered {
			break
		}
	}
	pick := openKeyedPick(t, g, me.ID, effects.ReturnChosenFromGraveyardThen.Key())
	if !effectstest.ContainsID(pick.ChooseCards, gone) {
		t.Fatalf("the return pick offers %v, want the Land card %s", pick.ChooseCards, gone)
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerChooseCards(t, g, me.ID, gone)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(gone) {
		t.Error("the Land card did not return to the battlefield after the restore")
	}
	if me.Graveyard.Contains(gone) {
		t.Error("the Land card is still in the graveyard")
	}
}

func TestRiseOfTheWitchKingReturnPickWithAnUnknownKeyIsRefused(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gone := effectstest.PushGraveyardCardForTest(me, "Gone Land")
	g.WithWriteLock(func() {
		for _, c := range me.Graveyard.Cards {
			if c.InstanceID == gone {
				c.TypeLine = "Land"
			}
		}
	})
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Kind: game.StackItemSpell, Controller: me.ID, Owner: me.ID})
		if err := (effects.ReturnChosenFromGraveyard{Question: "Return a card", Max: 1}).Apply(ctx); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	})
	openKeyedPick(t, g, me.ID, effects.ReturnChosenFromGraveyardThen.Key())
	refuseUnknownPickKey(t, g, effects.ReturnChosenFromGraveyardThen.Key())
}

// --- the attack-target picks ---------------------------------------------

// Chorale of the Void (putAttackingFromGraveyard): the raised Clone is put
// onto the battlefield tapped and attacking what its controller picks; the
// "what does it attack?" pick is a restore point.
func TestChoraleOfTheVoidAttackPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.PushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	aura := effectstest.PushCatalogPermanent(g, me.ID, "Chorale of the Void", "Enchantment — Aura", effectstest.ChoraleOfTheVoidOracle, false)
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(aura, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("attach: %v", err)
		}
	})
	clone := uuid.New()
	opp.Graveyard.PushTop(game.Card{
		InstanceID: clone, Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.SeedCopyableCreature(g, me.ID, "Hill Giant", "Creature — Giant", 3, 3)

	effectstest.DeclareAttack(t, g, opp.ID, host)
	effectstest.PickCard(t, g, me.ID, clone)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no \"what does it attack?\" pick is open")
	}
	if pick.PickThen != effects.PutAttackingFromGraveyard.Key() {
		t.Fatalf("the pick's continuation is %q, want the keyed %q", pick.PickThen, effects.PutAttackingFromGraveyard.Key())
	}
	if len(pick.PickOptions) < 2 {
		t.Fatalf("%d options, want several defenders", len(pick.PickOptions))
	}

	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	if g.Battlefield.Contains(clone) {
		t.Fatal("the restored table put the Clone onto the battlefield before the answer")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	// The Clone asks what to copy as it enters; any copy will do.
	if effectstest.CopyPrompt(g) != nil {
		effectstest.AnswerCloneCopy(t, g, host)
	}
	var raised game.Card
	var found bool
	g.ReadSnapshot(func() {
		if g.Battlefield.Contains(clone) {
			raised, found = g.LookupCardForEffect(clone)
		}
	})
	if !found {
		t.Fatal("the Clone did not enter after the restored answer")
	}
	if raised.Controller != me.ID || !raised.Tapped || raised.AttackingTarget == uuid.Nil {
		t.Errorf("raised controller=%v tapped=%v attacking=%v, want a tapped, attacking creature of mine",
			raised.Controller, raised.Tapped, raised.AttackingTarget)
	}
	if raised.AttackingTarget == opp.ID {
		t.Errorf("attacking %v, want the second option (not the first defender)", raised.AttackingTarget)
	}
}

// Ultra Magnus, Tactician: the card pick and the defender pick are both
// restore points, and the convert tail still reads the resolving item.
func TestUltraMagnusPicksAreRestorePoints(t *testing.T) {
	g, me, opp := effectstest.FlipMain(t)
	id := effectstest.TransformerIn(t, g, effectstest.UltraMagnusRow(), me)
	robot := effectstest.HandCardForTest(me, "Hired Robot", "Artifact Creature — Robot", "")

	effectstest.TransformerAttacks(t, g, id, opp.ID)
	openKeyedPick(t, g, me.ID, cardsu.UltraMagnusPutter.Picked.Key())
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerChooseCards(t, g, me.ID, robot)

	pick := effectstest.LatestOptionPickFor(g, me.ID)
	if pick == nil {
		t.Fatal("no defender pick is open")
	}
	if pick.PickThen != cardsu.UltraMagnusPutter.Attacks.Key() {
		t.Fatalf("the defender pick's continuation is %q, want %q", pick.PickThen, cardsu.UltraMagnusPutter.Attacks.Key())
	}
	g = effectstest.RestoreRoundTrip(t, g, true)
	me = g.Seats[0]
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.TfSettle(t, g)

	c := effectstest.FlipState(t, g, robot)
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Fatalf("the put creature is tapped=%v attacking %v, want tapped and attacking the opponent", c.Tapped, c.AttackingTarget)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.TfSettle(t, g)
	effectstest.WantFace(t, g, id, 1, "Ultra Magnus, Armored Carrier")
}

// The Grim Captain: the exiled material pick and the defender pick are
// restore points.
func TestGrimCaptainPutPicksAreRestorePoints(t *testing.T) {
	g, me, opp := effectstest.SpendTable(t)
	throne := effectstest.PushCraftCard(g, me, effectstest.TransformRow(cardsth.ThroneOfTheGrimCaptainOracleID, "Throne of the Grim Captain", "Legendary Artifact", "{2}",
		"The Grim Captain", "Legendary Creature — Skeleton Spirit Pirate", "7", "7", []string{"B"}))
	dino := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Dino", TypeLine: "Creature — Dinosaur", Power: 4, Toughness: 4})
	merfolk := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Merrow", TypeLine: "Creature — Merfolk", Power: 1, Toughness: 1})
	pirate := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Pirate", TypeLine: "Creature — Human Pirate", Power: 1, Toughness: 1})
	vampire := effectstest.PushGraveyardMaterial(me, game.Card{Name: "Vampire", TypeLine: "Creature — Vampire", Power: 1, Toughness: 1})
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Bear", "Creature — Bear", "", false)
	captain := effectstest.CraftInto(t, g, me, throne, 1, "{C}{C}{C}{C}", "The Grim Captain", dino, merfolk, pirate, vampire)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == captain.InstanceID {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.DeclareAttack(t, g, opp.ID, captain.InstanceID)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirs)
	openKeyedPick(t, g, me.ID, cardsth.GrimCaptainPut.Key())

	g = effectstest.RestoreRoundTrip(t, g, true)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AnswerChooseCards(t, g, me.ID, dino)
	if pick := effectstest.LatestOptionPickFor(g, me.ID); pick != nil {
		if pick.PickThen != effects.PutAttackingFromExile.Key() {
			t.Fatalf("the defender pick's continuation is %q, want %q", pick.PickThen, effects.PutAttackingFromExile.Key())
		}
		g = effectstest.RestoreRoundTrip(t, g, true)
		me = g.Seats[0]
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	ids := effectstest.BattlefieldIDsNamed(g, "Dino")
	if len(ids) != 1 {
		t.Fatalf("%d Dinos on the battlefield, want the material put in attacking", len(ids))
	}
	var dinoNow game.Card
	g.ReadSnapshot(func() { dinoNow, _ = g.LookupCardForEffect(ids[0]) })
	if !dinoNow.Tapped || dinoNow.AttackingTarget != opp.ID || dinoNow.Controller != me.ID {
		t.Errorf("Dino tapped=%v attacking=%v controller=%v, want tapped, attacking the opponent, mine",
			dinoNow.Tapped, dinoNow.AttackingTarget, dinoNow.Controller)
	}
}

// --- amass ----------------------------------------------------------------

// Widespread Brutality amasses with the Army's creation held on the order
// prompt, then asks which Army to amass onto: that pick is a restore point
// too, and the restored answer puts the counters on the chosen one and runs
// the card's own tail with it.
func TestAnAmassArmyPickIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ds, pl := effectstest.TokenOrderBoard(g, me.ID)
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear",
		Power: 2, Toughness: 2, PrintedPTKnown: true, Owner: opp.ID, Controller: opp.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Widespread Brutality", "Sorcery", effectstest.WidespreadBrutalityOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeCreateTokens)
	effectstest.AnswerOrderBySource(t, g, ds, pl)
	openKeyedPick(t, g, me.ID, "choose-cards/amass-army")

	g = effectstest.RestoreRoundTrip(t, g, true)
	var armies []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if game.IsArmy(c) {
			armies = append(armies, c.InstanceID)
		}
	}
	if len(armies) != 4 {
		t.Fatalf("%d Armies on the restored table, want one doubled twice", len(armies))
	}
	effectstest.AnswerChooseCards(t, g, me.ID, armies[0])
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, armies[0], game.CounterPlusOne); got < 2 {
		t.Errorf("the chosen Army has %d counters, want the amass's", got)
	}
	if effectstest.FindBattlefieldCardByID(g, bear) != nil {
		t.Error("the restored pick did not run the sweep: the 2/2 Bear survived the amassed Army")
	}
}
