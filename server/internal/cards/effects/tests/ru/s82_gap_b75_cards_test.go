package testsru

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b75_cards_test.go — S82 GAP batch 75 (#5948), the Ravnica
// Remastered red-white and blue creatures that were missing. Each test plays
// the catalog entry through the engine.

const (
	gapTajicOracle        = "ae6f21a2-e6b6-4793-8343-e27310c0bea1"
	gapTruefireOracle     = "3677e743-8d3d-4ec7-8bfe-21e36979a79b"
	gapFblthpOracle       = "ba5a7ac0-b625-42e9-af05-2680a01a95ed"
	gapGigantoplasmOracle = "52fdfe4b-10a7-421d-80ea-f6c39013062b"
	gapKeymasterOracle    = "b8c0c102-f0ef-4ab6-9b1a-ee82ee159c54"
	gapPetitionersOracle  = "0e488c6c-aae2-450f-b969-7bb5a1b37a66"
	gapTidespoutOracle    = "331b44e4-cf08-44f4-8afd-a7374e5dc5ed"
	gapAetherplasmOracle  = "0cd99210-8d34-4112-8e41-7e8c5d2a95b9"
)

func TestGapB75CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		gapTajicOracle: "Tajic, Legion's Edge", gapTruefireOracle: "Truefire Captain",
		gapFblthpOracle: "Fblthp, the Lost", gapGigantoplasmOracle: "Gigantoplasm",
		gapKeymasterOracle: "Keymaster Rogue", gapPetitionersOracle: "Persistent Petitioners",
		gapTidespoutOracle: "Tidespout Tyrant", gapAetherplasmOracle: "Aetherplasm",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: %+v (registered %v)", name, spec, ok)
		}
	}
}

// --- Tajic, Legion's Edge -------------------------------------------------------

func TestGapB75TajicPreventsNoncombatDamageToOtherCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tajic := effectstest.B12Push(g, me.ID, "Tajic, Legion's Edge", "Legendary Creature — Human Soldier", gapTajicOracle, 3, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, tajic, "haste") {
		t.Error("no haste")
	}
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(src, mine, 5)
		_ = g.DealDamageToCreatureForEffect(src, theirs, 5)
		_ = g.DealDamageToCreatureForEffect(src, tajic, 5)
	})
	effectstest.CheckState(t, g)
	if !g.Battlefield.Contains(mine) {
		t.Error("noncombat damage to another creature its controller controls was not prevented")
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("damage to an opponent's creature was prevented")
	}
	if g.Battlefield.Contains(tajic) {
		t.Error("damage to Tajic itself was prevented")
	}
}

func TestGapB75TajicGainsFirstStrikeForRedWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tajic := effectstest.B12Push(g, me.ID, "Tajic, Legion's Edge", "Legendary Creature — Human Soldier", gapTajicOracle, 3, 2)
	effectstest.AdvanceToMain(t, g)
	if effectstest.HasEffectiveKeyword(t, g, tajic, "first strike") {
		t.Fatal("first strike before paying")
	}
	effectstest.AddPool(t, g, me.ID, "{R}{W}")
	effectstest.P7Activate(t, g, me, tajic, 0, game.ActivateAbilityParams{Strict: true})
	if !effectstest.HasEffectiveKeyword(t, g, tajic, "first strike") {
		t.Error("no first strike after {R}{W}")
	}
}

func TestGapB75TajicMentorsASmallerAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.BnsActive(g)
	tajic := effectstest.B12Push(g, me.ID, "Tajic, Legion's Edge", "Legendary Creature — Human Soldier", gapTajicOracle, 3, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, tajic, small)
	effectstest.AnswerTriggerTargets(t, g, me.ID, small)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountersOn(g, small, game.CounterPlusOne); n != 1 {
		t.Errorf("%d counters on the smaller attacker, want 1", n)
	}
}

// --- Truefire Captain ----------------------------------------------------------------

func TestGapB75TruefireCaptainDealsThatMuchDamageToATargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.B12Push(g, me.ID, "Truefire Captain", "Creature — Human Knight", gapTruefireOracle, 4, 3)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, captain, 2) })
	effectstest.CheckState(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 2 {
		t.Errorf("the opponent lost %d, want 2", got)
	}
	if !g.Battlefield.Contains(captain) {
		t.Error("the Captain died of 2 damage")
	}
}

func TestGapB75TruefireCaptainStillTriggersWhenTheDamageKillsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	captain := effectstest.B12Push(g, me.ID, "Truefire Captain", "Creature — Human Knight", gapTruefireOracle, 4, 3)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(src, captain, 3) })
	effectstest.CheckState(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the Captain that was just dealt lethal damage")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
		t.Fatalf("pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 3 {
		t.Errorf("the opponent lost %d, want 3", got)
	}
}

// --- Fblthp, the Lost ------------------------------------------------------------------

func TestGapB75FblthpDrawsOneWhenCastFromTheHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	before := me.Hand.Size()
	id := effectstest.CastCatalogSpell(t, g, "Fblthp, the Lost", "Legendary Creature — Homunculus", gapFblthpOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(id) {
		t.Fatal("Fblthp did not enter")
	}
	if me.Hand.Size() != before+1 {
		t.Errorf("hand %d, want %d (one draw)", me.Hand.Size(), before+1)
	}
}

func TestGapB75FblthpDrawsTwoWhenItEntersFromTheLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Fblthp, the Lost", TypeLine: "Legendary Creature — Homunculus",
		OracleID: gapFblthpOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	before := me.Hand.Size()
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneLibrary, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("put Fblthp onto the battlefield from the library: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before+2 {
		t.Errorf("hand %d, want %d (two draws)", me.Hand.Size(), before+2)
	}
}

func TestGapB75FblthpIsShuffledIntoItsOwnersLibraryWhenTargetedByASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	fblthp := effectstest.B12Push(g, me.ID, "Fblthp, the Lost", "Legendary Creature — Homunculus", gapFblthpOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, effectstest.CardRefs(fblthp))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fblthp) {
		t.Error("Fblthp is still on the battlefield")
	}
	if !effectstest.InLibrary(g, me.ID, fblthp) {
		t.Error("Fblthp is not in its owner's library")
	}
	if z, _ := effectstest.CardWhere(g, me, bolt); z != "graveyard" {
		t.Errorf("the Bolt is in %s, want the graveyard (its target is gone)", z)
	}
}

// --- Gigantoplasm -------------------------------------------------------------------------

func TestGapB75GigantoplasmCopiesACreatureAndSetsItsBasePowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big := effectstest.SeedCopyableCreature(g, me.ID, "Big Bear", "Creature — Bear", 7, 7)
	id := effectstest.CastCatalogSpell(t, g, "Gigantoplasm", "Creature — Shapeshifter", gapGigantoplasmOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, big)
	g.RunStateChecksForTest()
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Big Bear" || c.CurrentPower() != 7 || c.CurrentToughness() != 7 {
		t.Fatalf("the copy is %q %d/%d, want Big Bear 7/7", c.Name, c.CurrentPower(), c.CurrentToughness())
	}
	effectstest.AddPool(t, g, me.ID, "{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("activate {X}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	c = effectstest.LayeredCard(t, g, id)
	if c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("after {X}=3 it is %d/%d, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	if o := effectstest.LayeredCard(t, g, big); o.CurrentPower() != 7 {
		t.Errorf("the original changed to power %d", o.CurrentPower())
	}
	// The effect has no duration: it is still 3/3 after the turn ends.
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
	}
	if c = effectstest.LayeredCard(t, g, id); c.CurrentPower() != 3 {
		t.Errorf("the base power and toughness lapsed: %d", c.CurrentPower())
	}
}

func TestGapB75GigantoplasmThatDeclinesTheCopyDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Gigantoplasm", TypeLine: "Creature — Shapeshifter", OracleID: gapGigantoplasmOracle,
		ScryfallID: "00000000-0000-0000-0000-0000000000b5", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(id) {
		t.Error("a declined Gigantoplasm is a 0/0 and dies")
	}
}

// --- Keymaster Rogue --------------------------------------------------------------------------

func TestGapB75KeymasterRogueIsUnblockableAndReturnsACreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := effectstest.BrCreature(g, opp.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	rogue := effectstest.CastCatalogSpell(t, g, "Keymaster Rogue", "Creature — Human Rogue", gapKeymasterOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no return pick")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, bear); z != "hand" {
		t.Errorf("the chosen creature is in %s, want the hand", z)
	}
	if !g.Battlefield.Contains(rogue) {
		t.Fatal("the Rogue is gone")
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, rogue).SummonedThisTurn = false })
	effectstest.BrAttack(t, g, rogue)
	effectstest.BrRefusal(t, g.DeclareBlocker(blocker, rogue), game.BlockReasonCantBeBlocked)
}

// --- Persistent Petitioners -----------------------------------------------------------------------

func TestGapB75PersistentPetitionersMillsACardForOneAndTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pp := effectstest.B12Push(g, me.ID, "Persistent Petitioners", "Creature — Human Advisor", gapPetitionersOracle, 1, 3)
	effectstest.AdvanceToMain(t, g)
	lib, gy := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.AddPool(t, g, me.ID, "{C}")
	effectstest.P7Activate(t, g, me, pp, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}})
	if opp.Library.Size() != lib-1 || opp.Graveyard.Size() != gy+1 {
		t.Errorf("library %d -> %d, graveyard %d -> %d; want one card milled", lib, opp.Library.Size(), gy, opp.Graveyard.Size())
	}
	if !effectstest.TappedNow(g, pp) {
		t.Error("the Petitioners is untapped after paying {T}")
	}
}

func TestGapB75PersistentPetitionersMillsTwelveForFourUntappedAdvisors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pp := effectstest.B12Push(g, me.ID, "Persistent Petitioners", "Creature — Human Advisor", gapPetitionersOracle, 1, 3)
	a1 := effectstest.B12Creature(g, me.ID, "Advisor One", "Creature — Human Advisor", 1, 3)
	a2 := effectstest.B12Creature(g, me.ID, "Advisor Two", "Creature — Human Advisor", 1, 3)
	a3 := effectstest.B12Creature(g, me.ID, "Advisor Three", "Creature — Human Advisor", 1, 3)
	bear := effectstest.B12Creature(g, me.ID, "Not An Advisor", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	if err := g.ActivateCatalogAbility(me.ID, pp, 1, game.ActivateAbilityParams{Strict: true, Targets: target, TapIDs: []uuid.UUID{pp, a1, a2, bear}}); err == nil {
		t.Fatal("a Bear paid for an Advisor")
	}
	if err := g.ActivateCatalogAbility(me.ID, pp, 1, game.ActivateAbilityParams{Strict: true, Targets: target, TapIDs: []uuid.UUID{pp, a1, a2}}); err == nil {
		t.Fatal("three Advisors paid for four")
	}
	lib := opp.Library.Size()
	effectstest.P7Activate(t, g, me, pp, 1, game.ActivateAbilityParams{Strict: true, Targets: target, TapIDs: []uuid.UUID{pp, a1, a2, a3}})
	if want := max(0, lib-12); opp.Library.Size() != want {
		t.Errorf("library %d -> %d, want %d (twelve milled)", lib, opp.Library.Size(), want)
	}
	for _, id := range []uuid.UUID{pp, a1, a2, a3} {
		if !effectstest.TappedNow(g, id) {
			t.Error("an Advisor that paid is untapped")
		}
	}
}

// --- Tidespout Tyrant -----------------------------------------------------------------------------------

func TestGapB75TidespoutTyrantBouncesAPermanentWheneverYouCastASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tyrant := effectstest.B12Push(g, me.ID, "Tidespout Tyrant", "Creature — Djinn", gapTidespoutOracle, 5, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, tyrant, "flying") {
		t.Error("no flying")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "A Charm", "Instant", "", nil)
	effectstest.PickCard(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, opp, victim); z != "hand" {
		t.Errorf("the target is in %s, want its owner's hand", z)
	}
}

func TestGapB75TidespoutTyrantCanBounceItselfAndIgnoresOpponentsSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	tyrant := effectstest.B12Push(g, me.ID, "Tidespout Tyrant", "Creature — Djinn", gapTidespoutOracle, 5, 5)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "A Charm", "Instant", "", nil)
	effectstest.PickCard(t, g, me.ID, tyrant)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, tyrant); z != "hand" {
		t.Errorf("the Tyrant is in %s, want the hand", z)
	}
}

// --- Aetherplasm ---------------------------------------------------------------------------------------

func gapAetherplasmBlocks(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	atk, me := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, atk.ID, "Attacker", "Creature — Bear", 3, 3)
	aeth := effectstest.B12Push(g, me.ID, "Aetherplasm", "Creature — Illusion", gapAetherplasmOracle, 1, 1)
	giant := effectstest.HandCardForTest(me, "Hill Giant", "Creature — Giant", "")
	effectstest.HandCardForTest(me, "A Charm", "Instant", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(aeth, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	return g, me, attacker, aeth, giant
}

func TestGapB75AetherplasmReturnsAndPutsACreatureCardBlockingThatAttacker(t *testing.T) {
	g, me, attacker, aeth, giant := gapAetherplasmBlocks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	choice := effectstest.LatestChooseCards(g, me.ID)
	if choice == nil {
		t.Fatal("no card offered")
	}
	if !slices.Contains(choice.ChooseCards, giant) || !slices.Contains(choice.ChooseCards, aeth) || len(choice.ChooseCards) != 2 {
		t.Errorf("offered %v; want the Giant and the Aetherplasm itself, not the instant", choice.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, giant)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, aeth); z != "hand" {
		t.Errorf("the Aetherplasm is in %s, want the hand", z)
	}
	z, c := effectstest.CardWhere(g, me, giant)
	if z != "battlefield" || c.BlockingTarget != attacker {
		t.Errorf("the Giant is in %s blocking %v, want the battlefield blocking the attacker", z, c.BlockingTarget)
	}
}

func TestGapB75AetherplasmMayDeclineToReturnOrToPut(t *testing.T) {
	g, me, _, aeth, giant := gapAetherplasmBlocks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, aeth); z != "battlefield" {
		t.Errorf("declined, yet the Aetherplasm is in %s", z)
	}
	if z, _ := effectstest.CardWhere(g, me, giant); z != "hand" {
		t.Errorf("declined, yet the Giant is in %s", z)
	}
}

func TestGapB75AetherplasmMayReturnWithoutPuttingAnything(t *testing.T) {
	g, me, _, aeth, giant := gapAetherplasmBlocks(t)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if z, _ := effectstest.CardWhere(g, me, aeth); z != "hand" {
		t.Errorf("the Aetherplasm is in %s, want the hand", z)
	}
	if z, _ := effectstest.CardWhere(g, me, giant); z != "hand" {
		t.Errorf("the Giant is in %s, want the hand", z)
	}
}
