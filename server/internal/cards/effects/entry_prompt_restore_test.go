package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// entry_prompt_restore_test.go — #4253 slice 2: the entry questions that
// pause through the CR 616 resume frame, and the counter, life and
// keyword-action events with no closure tail, are restore points. Each
// test takes a STRICT restore (RestoreStrict, the boot path) while the
// prompt is open, answers it on the restored table, and checks what a
// player sees afterwards. The frame itself is game/replacement_resume_keyed.go.

const rhoxFaithmenderOracle = "2abe9303-d498-4aad-b6b2-8b5064bd2ffd"

// --- copy (CR 707.2) -------------------------------------------------------

// Clone's "you may have it enter as a copy" is a restore point. The
// restored question copies the creature picked, or declines.
func TestACopyTargetQuestionIsARestorePoint(t *testing.T) {
	for _, copyIt := range []bool{true, false} {
		g := newCatalogGame(t)
		active := g.Seats[g.Turn.ActiveSeat]
		bears := seedCopyableCreature(g, active.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
		seedCopyableCreature(g, active.ID, "Hill Giant", "Creature — Giant", 3, 3)
		cloneID := castCatalogSpell(t, g, "Clone", "Creature — Shapeshifter", oracleClone, nil)
		passPriorityAroundTable(t, g)
		if copyPrompt(g) == nil {
			t.Fatal("Clone did not ask")
		}

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
		c := copyPrompt(g)
		if c == nil || len(c.CopyOptions) != 2 {
			t.Fatalf("the restored question is %+v, want two creatures to copy", c)
		}
		if _, ok := battlefieldCard(g, cloneID); ok {
			t.Fatal("the restored table landed Clone before the answer")
		}
		pick := uuid.Nil
		if copyIt {
			pick = bears
		}
		resolveWithCopyChoice(t, g, pick)
		got, ok := battlefieldCard(g, cloneID)
		if !ok {
			t.Fatal("the restored answer did not land Clone")
		}
		if copyIt && (got.Name != "Grizzly Bears" || !got.IsCopy()) {
			t.Errorf("copied: Clone is %q (copy %v), want a copy of Grizzly Bears", got.Name, got.IsCopy())
		}
		if !copyIt && (got.Name != "Clone" || got.IsCopy()) {
			t.Errorf("declined: Clone is %q (copy %v), want itself", got.Name, got.IsCopy())
		}
	}
}

// --- entry controller (ADR 0102) -------------------------------------------

// Captive Audience's "enters under the control of an opponent of your
// choice" is a restore point, and the restored answer is honoured.
func TestAnEntryControllerQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	id := castCatalogSpell(t, g, "Captive Audience", "Enchantment", captiveAudienceOracle, nil)
	passPriorityAroundTable(t, g)
	if entryControllerPrompt(g) == nil {
		t.Fatal("Captive Audience did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	c := entryControllerPrompt(g)
	if c == nil || len(c.PickOptions) != 3 || c.ControlPurpose != game.ControlForHarm {
		t.Fatalf("the restored question is %+v, want the three opponents, for harm", c)
	}
	victim := g.Seats[(seat+2)%4]
	answerEntryController(t, g, victim.ID)
	card, ok := battlefieldCardByID(g, id)
	if !ok || card.Controller != victim.ID || card.Owner != me.ID {
		t.Fatalf("Captive Audience landed as %+v (ok %v), want controlled by %s and owned by the caster", card, ok, victim.Name)
	}
}

// --- the entry card choices (ADR 0098, ADR 0170) ----------------------------

// A reveal land's "you may reveal" is a restore point: a restored reveal
// lands it untapped, a restored decline lands it tapped.
func TestAnEntryRevealQuestionIsARestorePoint(t *testing.T) {
	for _, reveal := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		swamp := seedRevealLandHand(me, "Basic Land — Swamp")[0]
		id := playLandFromHand(t, g, "Shineshadow Snarl", revealLandOracleIDs["Shineshadow Snarl"])
		if entryRevealChoiceFor(g, me.ID) == nil {
			t.Fatal("no reveal prompt")
		}

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
		me = g.Seats[g.Turn.ActiveSeat]
		if c := entryRevealChoiceFor(g, me.ID); c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != swamp {
			t.Fatalf("the restored reveal prompt is %+v, want the Swamp offered", c)
		}
		if reveal {
			answerEntryReveal(t, g, me.ID, swamp)
		} else {
			answerEntryReveal(t, g, me.ID)
		}
		card, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatal("the restored answer did not land the Snarl")
		}
		if card.Tapped == reveal {
			t.Errorf("reveal %v: the Snarl entered tapped=%v", reveal, card.Tapped)
		}
		if g.LandsPlayedThisTurn[me.ID] != 1 {
			t.Errorf("lands played this turn = %d, want 1", g.LandsPlayedThisTurn[me.ID])
		}
	}
}

// Mox Diamond's "you may discard a land card instead" is a restore
// point: the restored discard puts the Mox onto the battlefield. The
// restored prompt still re-checks its picks against the live hand (its
// zone frame is rebuilt from the prompt kind), so a land that has left
// the hand since is refused and the prompt stays open.
func TestAnEntryDiscardQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	forest := handCard(me, "Forest", "Basic Land — Forest")
	island := handCard(me, "Island", "Basic Land — Island")
	mox := castMox(t, g)
	if entryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID) == nil {
		t.Fatal("Mox Diamond did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	me = g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() {
		c, err := me.Hand.Remove(island)
		if err != nil {
			t.Fatalf("setup: %v", err)
		}
		g.Exile.PushTop(c)
	})
	c := entryChoiceOfKind(g, game.PendingChoiceEntryDiscardFromHand, me.ID)
	if c == nil {
		t.Fatal("the restored table has no discard prompt")
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, []uuid.UUID{island}); !errors.Is(err, game.ErrCardNotFound) {
		t.Fatalf("a pick no longer in the hand: err %v, want ErrCardNotFound", err)
	}
	answerEntryChoice(t, g, game.PendingChoiceEntryDiscardFromHand, me.ID, forest)
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, mox); !ok {
		t.Fatal("the restored discard did not land Mox Diamond")
	}
	if !me.Graveyard.Contains(forest) {
		t.Error("the discarded Forest is not in the graveyard")
	}
}

// Heart of Yavimaya's "sacrifice a Forest instead" is a restore point.
// The restored prompt still refuses the empty answer, because its
// live-zone frame comes back with it.
func TestAnEntrySacrificeQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	forest := pushSacrificeCandidate(g, me.ID, "Forest", "Basic Land — Forest", true)
	heart := playLandFromHand(t, g, "Heart of Yavimaya", sacrificeInsteadLandOracleIDs["Heart of Yavimaya"])
	if entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID) == nil {
		t.Fatal("Heart of Yavimaya did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	me = g.Seats[g.Turn.ActiveSeat]
	c := entryChoiceOfKind(g, game.PendingChoiceEntrySacrifice, me.ID)
	if c == nil {
		t.Fatal("the restored table has no sacrifice prompt")
	}
	if err := g.ResolveEntryCardChoice(c.ID, me.ID, nil); err == nil {
		t.Error("the restored prompt accepted an empty answer to a mandatory sacrifice")
	}
	answerEntryChoice(t, g, game.PendingChoiceEntrySacrifice, me.ID, forest)
	if _, ok := battlefieldCard(g, heart); !ok {
		t.Fatal("the restored sacrifice did not land Heart of Yavimaya")
	}
	if !me.Graveyard.Contains(forest) || sacrificeEventsFor(g, forest) != 1 {
		t.Errorf("the Forest was not sacrificed once: in graveyard %v, %d sacrifices",
			me.Graveyard.Contains(forest), sacrificeEventsFor(g, forest))
	}
}

// Phylactery Lich's "as this enters, choose an artifact you control" is
// a restore point: the restored choice puts the counter on the artifact
// picked, and the Lich lands.
func TestAnEntryChoosePermanentsQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	golem := b12Creature(g, me.ID, "Steel Golem", "Artifact Creature — Golem", 3, 4)
	rock := b12Permanent(g, me.ID, "Mind Stone", "Artifact")
	id := castPhylacteryLich(t, g)
	if entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID) == nil {
		t.Fatal("the Lich did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	me = g.Seats[g.Turn.ActiveSeat]
	c := entryChoiceOfKind(g, game.PendingChoiceEntryChoosePermanents, me.ID)
	if c == nil || len(c.ChooseSuggested) != 1 || c.ChooseSuggested[0] != rock {
		t.Fatalf("the restored prompt is %+v, want the noncreature artifact suggested", c)
	}
	answerEntryChoice(t, g, game.PendingChoiceEntryChoosePermanents, me.ID, golem)
	passPriorityAroundTable(t, g)
	if got := b12Counter(t, g, golem, phylacteryCounter); got != 1 {
		t.Errorf("phylactery counters on the golem = %d, want 1", got)
	}
	if _, ok := battlefieldCard(g, id); !ok {
		t.Fatal("the restored choice did not land the Lich")
	}
}

// --- aura host (CR 303.4f, #4260) --------------------------------------------

// Animate Dead returned by Sun Titan asks what it enchants. The restored
// question attaches it to the creature card picked, which returns.
func TestAnAuraHostQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aura := ahAnimateDead(me)
	mine := adGraveyardCreature(me, "My Ghoul")
	theirs := adGraveyardCreature(opp, "Their Zombie")
	sunTitanReturns(t, g, aura)
	if auraHostChoiceFor(g, me.ID) == nil {
		t.Fatal("Animate Dead did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	me = g.Seats[0]
	if c := auraHostChoiceFor(g, me.ID); c == nil || len(c.PickTargetCards) != 2 {
		t.Fatalf("the restored question is %+v, want both creature cards offered", c)
	}
	if adZone(g, aura) != game.ZoneGraveyard {
		t.Fatalf("the restored table moved the Aura before the answer: %v", adZone(g, aura))
	}
	answerAuraHost(t, g, me.ID, cardRef(theirs))
	passPriorityAroundTable(t, g)
	if adZone(g, theirs) != game.ZoneBattlefield || adZone(g, mine) != game.ZoneGraveyard {
		t.Fatalf("the restored answer returns their Zombie: mine=%v theirs=%v", adZone(g, mine), adZone(g, theirs))
	}
	if !adCard(t, g, aura).IsAttachedTo(theirs) {
		t.Error("Animate Dead is not on the creature it returned")
	}
}

// --- read ahead (CR 702.155) --------------------------------------------------

// Read ahead's starting chapter is a restore point: the restored answer
// lands the Saga on the chapter picked, and the chapter rule still holds
// for the turn it entered.
func TestAReadAheadQuestionIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	id := castReadAheadSaga(t, g, "Love Song of Night and Day", loveSongOracle)
	if readAheadPrompt(g) == nil {
		t.Fatal("read ahead did not ask")
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
	me := g.Seats[g.Turn.ActiveSeat]
	if c := readAheadPrompt(g); c == nil || len(c.PickOptions) != 3 {
		t.Fatalf("the restored question is %+v, want three chapters", c)
	}
	chooseChapter(t, g, 2)
	passPriorityAroundTable(t, g)
	if got := counterCount(g, id, game.CounterLore); got != 2 {
		t.Fatalf("lore counters = %d, want 2", got)
	}
	g.WithWriteLock(func() {
		if err := g.ProliferateForEffect(me.ID, uuid.Nil, []uuid.UUID{id}, nil); err != nil {
			t.Fatalf("ProliferateForEffect: %v", err)
		}
	})
	if got := chaptersFired(g, id); len(got) == 0 || !slices.Equal(got[len(got)-1:], []int{3}) {
		t.Fatalf("chapters fired = %v, want chapter III last", got)
	}
}

// --- the entry life payment ---------------------------------------------------

// A shockland's "you may pay 2 life" is a restore point: a restored pay
// lands it untapped for 2 life, a restored decline lands it tapped.
func TestAnEntryLifePaymentIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		life := me.Life
		id := playLandFromHand(t, g, "Steam Vents", steamVentsOracle)
		if entryPayLifeChoiceFor(g, me.ID) == nil {
			t.Fatal("no pay-life prompt")
		}

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeMove)
		me = g.Seats[g.Turn.ActiveSeat]
		if c := entryPayLifeChoiceFor(g, me.ID); c == nil || c.LifeCost != 2 {
			t.Fatalf("the restored prompt is %+v, want a 2-life cost", c)
		}
		answerEntryPayLife(t, g, me.ID, pay)
		card, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatal("the restored answer did not land Steam Vents")
		}
		wantLife := life
		if pay {
			wantLife -= 2
		}
		if card.Tapped == pay || me.Life != wantLife {
			t.Errorf("pay %v: tapped %v, life %d; want tapped %v, life %d", pay, card.Tapped, me.Life, !pay, wantLife)
		}
	}
}

// --- a counter placement and a life change with no tail -----------------------

// Doubling Season and Hardened Scales over one +1/+1 counter: the CR 616
// ordering prompt is a restore point, and each restored order gives its
// own count — (1+1)×2 = 4 with the Scales first, 1×2+1 = 3 with the
// Season first.
func TestACounterOrderingPromptIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		scalesFirst bool
		want        int
	}{{true, 4}, {false, 3}} {
		g := newCatalogGame(t)
		p0 := g.Seats[0].ID
		hs := seedReplacementPermanent(g, hardenedScalesOracle, "Hardened Scales", p0)
		ds := seedReplacementPermanent(g, doublingSeasonOracle, "Doubling Season", p0)
		bear := seedCreature(g, "Bears", p0)
		if err := g.AddCounter(bear, "+1/+1", 1); err != nil {
			t.Fatalf("AddCounter: %v", err)
		}

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeCounter)
		if got := countersOn(g, bear, "+1/+1"); got != 0 {
			t.Fatalf("the restored table placed %d counters before the order", got)
		}
		if tc.scalesFirst {
			answerOrderBySource(t, g, hs, ds)
		} else {
			answerOrderBySource(t, g, ds, hs)
		}
		if got := countersOn(g, bear, "+1/+1"); got != tc.want {
			t.Errorf("scales first %v: %d counters, want %d", tc.scalesFirst, got, tc.want)
		}
	}
}

// Rhox Faithmender and Alhammarret's Archive over one life gain: two
// different declared doublers, so the gainer orders them. The ordering
// prompt is a restore point, and the restored answer gains ×4.
func TestALifeOrderingPromptIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	faith := seedReplacementPermanent(g, rhoxFaithmenderOracle, "Rhox Faithmender", me.ID)
	archive := seedReplacementPermanent(g, alhammarretsArchiveOracle, "Alhammarret's Archive", me.ID)
	life := me.Life
	if _, err := g.ChangePlayerLife(me.ID, 3); err != nil {
		t.Fatalf("ChangePlayerLife: %v", err)
	}
	if me.Life != life {
		t.Fatalf("life %d with the order open, want %d", me.Life, life)
	}

	g = strictRestoreWithFrame(t, g, game.ReplacementResumeLife)
	me = g.Seats[0]
	answerOrderBySource(t, g, archive, faith)
	if me.Life != life+12 {
		t.Errorf("life %d after the restored order, want %d (3 doubled twice)", me.Life, life+12)
	}
	if len(g.PendingChoices) != 0 {
		t.Errorf("questions left open: %+v", g.PendingChoices)
	}
}

// --- a keyword action whose tail is data ----------------------------------------

// A scry with no "then", under a scry doubler and a scry-plus-one: the
// ordering prompt is a restore point, and the restored order queues the
// scry with its own count (×2 then +1 over 2 is 5, +1 then ×2 is 6).
func TestAKeywordActionOrderingPromptIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		order []string
		cards int
	}{
		{[]string{"Doubling Probe — scry twice as many", "Scrying Probe — scry one more"}, 5},
		{[]string{"Scrying Probe — scry one more", "Doubling Probe — scry twice as many"}, 6},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		seedReplacementPermanent(g, scryPlusOneProbeOracle, "Scrying Probe", me.ID)
		seedReplacementPermanent(g, scryDoubledProbeOracle, "Doubling Probe", me.ID)
		g.WithWriteLock(func() {
			if err := (Scry{N: 2}).Apply(NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
				t.Fatalf("Scry: %v", err)
			}
		})

		g = strictRestoreWithFrame(t, g, game.ReplacementResumeKeywordAction)
		me = g.Seats[0]
		prompt := latestChoiceOfKindFor(g, game.PendingChoiceReplacementOrder, me.ID)
		if prompt == nil {
			t.Fatal("the restored table has no ordering prompt")
		}
		if err := g.ResolveReplacementOrder(prompt.ID, me.ID,
			replacementIDsByLabel(t, g, prompt.ReplacementEffectIDs, tc.order)); err != nil {
			t.Fatalf("ResolveReplacementOrder: %v", err)
		}
		scry := scryChoiceFor(g, me.ID)
		if scry == nil || len(scry.ScryCards) != tc.cards {
			t.Fatalf("order %v: the restored scry is %+v, want one over %d cards", tc.order, scry, tc.cards)
		}
	}
}
