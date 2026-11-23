package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b162_cards_test.go — Jumpstart 2022 creatures, GAP batch 162 (#5948).
// Every test plays the real catalog entry through the engine.

const (
	gap162Sage       = "852621bf-9ab7-4027-89fb-b438e4080684"
	gap162Sentinels  = "a93779c5-603e-4a91-aef8-a0cd8fd1b3e9"
	gap162Shaper     = "ee6b0bb4-9c40-435b-8a5d-fd89d7e29ae1"
	gap162Animator   = "6e51a0b2-e07f-4c9d-b055-dd1dcd2e29d4"
	gap162Sprite     = "32e60fb4-e841-4f82-9c83-5e63766e8e6f"
	gap162Angler     = "d29d28c8-4ba9-4fcc-b0d7-5d7b3de89b82"
	gap162Sculptor   = "e43eb15e-e910-4165-b9df-25cbe7054e0a"
	gap162TolSentnl  = "bef15918-c0e7-4f6d-a589-f91f3fa11ef1"
	gap162BoltOracle = effectstest.LightningBoltOracle
)

func TestGapB162EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{gap162Sage, gap162Sentinels, gap162Shaper, gap162Animator, gap162Sprite, gap162Angler, gap162Sculptor, gap162TolSentnl} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func TestGapB162SentinelsAndSpriteKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for _, tc := range []struct {
		name, typeLine, oracle string
		kws                    []string
	}{
		{"Sentinels of Glen Elendra", "Creature — Faerie Soldier", gap162Sentinels, []string{"flash", "flying"}},
		{"Spellstutter Sprite", "Creature — Faerie Wizard", gap162Sprite, []string{"flash", "flying"}},
		{"Tolarian Sentinel", "Creature — Human Spellshaper", gap162TolSentnl, []string{"flying"}},
	} {
		id := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, 1, 1)
		for _, kw := range tc.kws {
			if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
				t.Errorf("%s lacks %s", tc.name, kw)
			}
		}
	}
}

func TestGapB162SageLoopsForNonHumansOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 8)
	effectstest.B12Push(g, me.ID, "Sage of the Falls", "Creature — Merfolk Wizard", gap162Sage, 2, 5)

	// A Human entering: nothing.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Soldier", TypeLine: "Creature — Human Soldier", Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Fatal("the Sage triggered for a Human")
	}

	// A non-Human entering under an opponent's control: nothing.
	effectstest.EnterCard(t, g, g.Seats[1].ID, game.Card{Name: "Their Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Fatal("the Sage triggered for an opponent's creature")
	}

	// A non-Human entering under my control: may draw, then discard.
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil || ask.Chooser != me.ID {
		t.Fatal("no \"draw a card?\" prompt for a non-Human creature")
	}
	hand := me.Hand.Size()
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after drawing, want %d", me.Hand.Size(), hand+1)
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}
}

func TestGapB162SageTriggersForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.Msc18Library(me, 8)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Sage of the Falls", TypeLine: "Creature — Merfolk Wizard", OracleID: gap162Sage, Power: 2, Toughness: 5})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) == nil {
		t.Fatal("the Sage did not trigger as it entered")
	}
}

func TestGapB162ShaperApprenticeFliesOnlyWithAnotherMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	shaper := effectstest.B12Push(g, me.ID, "Shaper Apprentice", "Creature — Merfolk Wizard", gap162Shaper, 2, 1)
	if effectstest.HasKeywordOnBattlefield(t, g, shaper, "flying") {
		t.Fatal("flying with no other Merfolk")
	}
	pal := effectstest.B12Creature(g, me.ID, "Merfolk Pal", "Creature — Merfolk", 1, 1)
	if !effectstest.HasKeywordOnBattlefield(t, g, shaper, "flying") {
		t.Fatal("no flying with another Merfolk")
	}
	effectstest.B27Kill(g, pal)
	g.RunStateChecksForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, shaper, "flying") {
		t.Error("still flying after the other Merfolk left")
	}
}

func TestGapB162SkilledAnimatorAnimatesAnArtifactWhileItStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	animator := effectstest.CastHoldCreature(t, g, "Skilled Animator", "Creature — Human Artificer", gap162Animator, 1, 3)
	effectstest.AnswerTriggerTargets(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	live := effectstest.ApaLive(g, rock)
	if !live.IsCreature() || !live.IsArtifact() {
		t.Fatalf("the artifact is %q, want an artifact creature", live.TypeLine)
	}
	if live.CurrentPower() != 5 || live.CurrentToughness() != 5 {
		t.Errorf("P/T = %d/%d, want 5/5", live.CurrentPower(), live.CurrentToughness())
	}
	effectstest.B27Kill(g, animator)
	g.RunStateChecksForTest()
	if after := effectstest.ApaLive(g, rock); after == nil || after.IsCreature() {
		t.Error("the artifact is still a creature after the Animator left")
	}
}

func TestGapB162SkilledAnimatorOnlyTargetsYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	effectstest.B12Permanent(g, me.ID, "My Other Rock", "Artifact")
	effectstest.CastHoldCreature(t, g, "Skilled Animator", "Creature — Human Artificer", gap162Animator, 1, 3)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("an opponent's artifact was accepted as a target")
	}
}

func gap162BoltOnTheStack(t *testing.T, g *game.Game, cost string) (me *game.Player, bolt uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1; g.Turn.PriorityHolder = 1 })
	bolt = uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: gap162BoltOracle,
		ManaCost: cost, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err != nil {
		t.Fatalf("cast Bolt: %v", err)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = 0 })
	return me, bolt
}

func TestGapB162SpellstutterSpriteCountersASpellWithinFaerieCount(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := gap162BoltOnTheStack(t, g, "{R}")
	life := me.Life
	sprite := effectstest.CastFromSeat(t, g, me, "Spellstutter Sprite", "Creature — Faerie Wizard", gap162Sprite, nil)
	effectstest.PassUntilOnBattlefield(t, g, sprite)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d: the Bolt resolved", life, me.Life)
	}
	if !g.Seats[1].Graveyard.Contains(bolt) {
		t.Error("the countered Bolt is not in its owner's graveyard")
	}
}

func TestGapB162SpellstutterSpriteCannotTargetAMoreExpensiveSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := gap162BoltOnTheStack(t, g, "{1}{R}")
	sprite := effectstest.CastFromSeat(t, g, me, "Spellstutter Sprite", "Creature — Faerie Wizard", gap162Sprite, nil)
	effectstest.PassUntilOnBattlefield(t, g, sprite)
	if g.Seats[1].Graveyard.Contains(bolt) {
		t.Fatal("a mana value 2 spell was countered with one Faerie")
	}
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err == nil {
			t.Fatal("a mana value 2 spell was accepted with one Faerie")
		}
	}
}

func TestGapB162SpellstutterSpriteCountsEveryFaerie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, bolt := gap162BoltOnTheStack(t, g, "{1}{R}")
	effectstest.B12Creature(g, me.ID, "Faerie Friend", "Creature — Faerie", 1, 1)
	life := me.Life
	sprite := effectstest.CastFromSeat(t, g, me, "Spellstutter Sprite", "Creature — Faerie Wizard", gap162Sprite, nil)
	effectstest.PassUntilOnBattlefield(t, g, sprite)
	effectstest.PickCard(t, g, me.ID, bolt)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d → %d: the Bolt resolved with two Faeries", life, me.Life)
	}
}

func TestGapB162AnglerTapsOrUntapsATarget(t *testing.T) {
	for _, tc := range []struct {
		name       string
		startTaped bool
		option     int
		wantTapped bool
	}{
		{"tap an untapped creature", false, 0, true},
		{"untap a tapped creature", true, 1, false},
		{"leave it as it is", true, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			angler := effectstest.B12Push(g, me.ID, "Stonybrook Angler", "Creature — Merfolk Wizard", gap162Angler, 1, 2)
			victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
			if tc.startTaped {
				g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, victim).Tapped = true })
			}
			effectstest.ApaMana(me, "U", "U")
			effectstest.ApaActivate(t, g, me, angler, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}})
			effectstest.AnswerOptionPick(t, g, me.ID, tc.option)
			if got := effectstest.HmlTapped(t, g, victim); got != tc.wantTapped {
				t.Errorf("tapped = %v, want %v", got, tc.wantTapped)
			}
			if !effectstest.IsTapped(g, angler) {
				t.Error("the Angler did not tap for its cost")
			}
		})
	}
}

func TestGapB162StormSculptorIsUnblockableAndMustReturnACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	sculptor := effectstest.CastHoldCreature(t, g, "Storm Sculptor", "Creature — Merfolk Wizard", gap162Sculptor, 3, 2)
	effectstest.AssertRestrictions(t, g, sculptor, game.CantBeBlocked)
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if p == nil {
		t.Fatal("no choice of which creature to return")
	}
	if p.ChooseMin != 1 {
		t.Error("the return is mandatory")
	}
	if err := g.ResolveOwnPermanents(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	if g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Error("the Bear did not return to hand")
	}
	if !g.Battlefield.Contains(sculptor) {
		t.Error("the Sculptor left the battlefield")
	}
}

func TestGapB162TolarianSentinelDiscardsToReturnYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	sentinel := effectstest.B12Push(g, me.ID, "Tolarian Sentinel", "Creature — Human Spellshaper", gap162TolSentnl, 1, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.ApaMana(me, "U")
	discard := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: discard, Name: "Chaff", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, sentinel, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}}, DiscardIDs: []uuid.UUID{discard},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) {
		t.Error("the Mind Stone did not return to hand")
	}
	if !me.Graveyard.Contains(discard) {
		t.Error("the discarded card is not in the graveyard")
	}
}
