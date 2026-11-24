package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b047_cards_test.go — Strixhaven: School of Mages, batch 047
// (#5502). Every test plays the real catalog entry through the engine.

const (
	o47SquareUp            = "013a465b-7034-45fe-bf62-c3d273131ec5"
	o47ExponentialGrowth   = "18616f70-e30c-4ff7-b62a-55939a96ca1b"
	o47TemptedByTheOriq    = "304450b0-bde6-4d69-bcaf-fb9b9bae2e4b"
	o47AcademicProbation   = "400b2e17-ee48-4e22-a8a6-a607c215d1a2"
	o47Flunk               = "5cd99378-e8e2-43fe-80fc-13d680746a4a"
	o47GoBlank             = "60186579-285b-4034-a782-431f79df5b07"
	o47MercurialTransform  = "c7efe6c6-3da1-47a6-a8ad-30c36502a7ab"
	o47UmbralJuke          = "d56afdf6-51e0-4151-a86c-4c4827a50a0e"
	o47HarnessInfinity     = "efd31076-2bbe-4c65-9166-19ec0e832b57"
	o47Bear                = "Creature — Bear"
	o47Sorcery, o47Instant = "Sorcery", "Instant"
	o47SorceryLesson       = "Sorcery — Lesson"
	o47BoltOracle          = effectstest.LightningBoltOracle
)

func TestS82OldB047EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{
		o47SquareUp, o47ExponentialGrowth, o47TemptedByTheOriq, o47AcademicProbation, o47Flunk,
		o47GoBlank, o47MercurialTransform, o47UmbralJuke, o47HarnessInfinity,
	} {
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

// o47Cast puts the card in the active seat's hand and casts it strictly in
// a main phase, returning the cast's error.
func o47Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	p.Strict = true
	return id, g.CastSpell(me.ID, id, p)
}

func o47MustCast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, p game.CastSpellParams) uuid.UUID {
	t.Helper()
	id, err := o47Cast(t, g, name, typeLine, cost, oracle, p)
	if err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	return id
}

func o47Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o47Pay(t *testing.T, g *game.Game, spec string) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, g.Seats[g.Turn.ActiveSeat], spec)
}

// --- Square Up ----------------------------------------------------------------

func TestS82OldB047SquareUpSetsBasePowerAndToughnessUnderCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	mouse := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Mouse", TypeLine: o47Bear, Power: 1, Toughness: 1,
		Counters: map[string]int{"+1/+1": 1}, Owner: opp.ID, Controller: opp.ID,
	})
	o47Pay(t, g, "GG")
	o47MustCast(t, g, "Square Up", o47Instant, "{1}{G/U}", o47SquareUp, game.CastSpellParams{Targets: o47Target(mouse)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, mouse); p != 5 || tough != 5 {
		t.Errorf("the Mouse is %d/%d, want 5/5 (base 4/4 plus its counter)", p, tough)
	}
}

func TestS82OldB047SquareUpRefusesANoncreatureAndAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	bear := effectstest.B12Creature(g, me.ID, "Bear", o47Bear, 2, 2)
	o47Pay(t, g, "GG")
	if _, err := o47Cast(t, g, "Square Up", o47Instant, "{1}{G/U}", o47SquareUp, game.CastSpellParams{Targets: o47Target(rock)}); err == nil {
		t.Error("Square Up targeted an artifact")
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if _, err := o47Cast(t, g, "Square Up", o47Instant, "{1}{G/U}", o47SquareUp, game.CastSpellParams{Targets: o47Target(bear)}); err == nil {
		t.Error("Square Up was cast with no mana")
	}
}

// --- Exponential Growth ---------------------------------------------------------

func TestS82OldB047ExponentialGrowthDoublesPowerXTimes(t *testing.T) {
	for _, tc := range []struct{ x, wantP int }{{0, 3}, {1, 6}, {3, 24}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		bear := effectstest.B12Creature(g, me.ID, "Bear", o47Bear, 3, 2)
		o47Pay(t, g, "GG"+"CCCCCCCC"[:2*tc.x])
		o47MustCast(t, g, "Exponential Growth", o47Sorcery, "{X}{X}{G}{G}", o47ExponentialGrowth,
			game.CastSpellParams{XValue: tc.x, Targets: o47Target(bear)})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, bear); p != tc.wantP || tough != 2 {
			t.Errorf("X=%d: the Bear is %d/%d, want %d/2", tc.x, p, tough, tc.wantP)
		}
	}
}

func TestS82OldB047ExponentialGrowthRefusesAnX2ShortPoolAndANoncreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", o47Bear, 3, 2)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	o47Pay(t, g, "GGCCC") // X=2 costs six generic plus GG: one short
	if _, err := o47Cast(t, g, "Exponential Growth", o47Sorcery, "{X}{X}{G}{G}", o47ExponentialGrowth,
		game.CastSpellParams{XValue: 2, Targets: o47Target(bear)}); err == nil {
		t.Error("X=2 was paid for with five mana")
	}
	o47Pay(t, g, "CCCC")
	if _, err := o47Cast(t, g, "Exponential Growth", o47Sorcery, "{X}{X}{G}{G}", o47ExponentialGrowth,
		game.CastSpellParams{XValue: 1, Targets: o47Target(rock)}); err == nil {
		t.Error("Exponential Growth targeted an artifact")
	}
}

// --- Tempted by the Oriq --------------------------------------------------------

func TestS82OldB047TemptedByTheOriqTakesOnePerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var picks []uuid.UUID
	var others []*game.Player
	for i := 1; i < len(g.Seats); i++ {
		others = append(others, g.Seats[(g.Turn.ActiveSeat+i)%len(g.Seats)])
	}
	for _, s := range others {
		picks = append(picks, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Small", TypeLine: o47Bear, ManaCost: "{2}",
			Power: 2, Toughness: 2, Owner: s.ID, Controller: s.ID,
		}))
	}
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: o47Bear, ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, Owner: others[0].ID, Controller: others[0].ID,
	})
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Jace Jr.", TypeLine: "Legendary Planeswalker — Jace", ManaCost: "{1}{U}",
		Counters: map[string]int{game.CounterLoyalty: 3}, Owner: others[1].ID, Controller: others[1].ID,
	})
	// A 4-mana creature is not a legal pick.
	o47Pay(t, g, "UUUU")
	if _, err := o47Cast(t, g, "Tempted by the Oriq", o47Sorcery, "{1}{U}{U}{U}", o47TemptedByTheOriq,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}}}); err == nil {
		t.Error("a mana value 4 creature was taken")
	}
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: picks[0]},
		{Kind: game.TargetCard, ID: walker},
		{Kind: game.TargetCard, ID: picks[2]},
	}
	o47MustCast(t, g, "Tempted by the Oriq", o47Sorcery, "{1}{U}{U}{U}", o47TemptedByTheOriq, game.CastSpellParams{Targets: targets})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{picks[0], walker, picks[2]} {
		c, ok := effectstest.BattlefieldCard(g, id)
		if !ok || c.Controller != me.ID {
			t.Errorf("%s is not under my control", c.Name)
		}
	}
	if c, _ := effectstest.BattlefieldCard(g, picks[1]); c.Controller == me.ID {
		t.Error("an untargeted creature changed hands")
	}
	if c, _ := effectstest.BattlefieldCard(g, big); c.Controller == me.ID {
		t.Error("the mana value 4 creature changed hands")
	}
}

func TestS82OldB047TemptedByTheOriqRefusesTwoPicksFromOneOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	a := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "A", TypeLine: o47Bear,
		ManaCost: "{1}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	b := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "B", TypeLine: o47Bear,
		ManaCost: "{1}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	o47Pay(t, g, "UUUU")
	if _, err := o47Cast(t, g, "Tempted by the Oriq", o47Sorcery, "{1}{U}{U}{U}", o47TemptedByTheOriq,
		game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}}); err == nil {
		t.Error("two creatures of one opponent were taken")
	}
}

// --- Flunk -------------------------------------------------------------------------

func TestS82OldB047FlunkCountsTheTargetControllersHand(t *testing.T) {
	for _, tc := range []struct{ hand, wantP int }{{0, 2}, {3, 5}, {6, 8}, {7, 9}, {9, 9}} {
		g := effectstest.NewCatalogGame(t)
		opp := effectstest.Top100Opponent(g)
		g.WithWriteLock(func() { opp.Hand.Cards = nil })
		for i := 0; i < tc.hand; i++ {
			effectstest.HandCardForTest(opp, "Filler", o47Sorcery, "")
		}
		ogre := effectstest.B12Creature(g, opp.ID, "Ogre", o47Bear, 9, 9)
		o47Pay(t, g, "BB")
		o47MustCast(t, g, "Flunk", o47Instant, "{1}{B}", o47Flunk, game.CastSpellParams{Targets: o47Target(ogre)})
		effectstest.PassPriorityAroundTable(t, g)
		if p, tough := effectstest.PtOf(t, g, ogre); p != tc.wantP || tough != tc.wantP {
			t.Errorf("hand %d: the Ogre is %d/%d, want %d/%d", tc.hand, p, tough, tc.wantP, tc.wantP)
		}
	}
}

func TestS82OldB047FlunkRefusesANoncreatureTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	o47Pay(t, g, "BB")
	if _, err := o47Cast(t, g, "Flunk", o47Instant, "{1}{B}", o47Flunk, game.CastSpellParams{Targets: o47Target(rock)}); err == nil {
		t.Error("Flunk targeted an artifact")
	}
}

// --- Go Blank ------------------------------------------------------------------------

func TestS82OldB047GoBlankDiscardsTwoThenExilesTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	a := effectstest.HandCardForTest(opp, "Alpha", o47Sorcery, "")
	b := effectstest.HandCardForTest(opp, "Beta", o47Sorcery, "")
	keep := effectstest.HandCardForTest(opp, "Gamma", o47Sorcery, "")
	old := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old", TypeLine: o47Sorcery, Owner: opp.ID, Controller: opp.ID})
	})
	o47Pay(t, g, "BCC")
	o47MustCast(t, g, "Go Blank", o47Sorcery, "{2}{B}", o47GoBlank, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(keep) || len(opp.Hand.Cards) != 1 {
		t.Errorf("their hand is %d cards, want only the card kept", len(opp.Hand.Cards))
	}
	if len(opp.Graveyard.Cards) != 0 {
		t.Errorf("their graveyard holds %d cards, want none", len(opp.Graveyard.Cards))
	}
	for _, id := range []uuid.UUID{a, b, old} {
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
	}
}

func TestS82OldB047GoBlankWithAnEmptyHandStillExilesTheGraveyardAndNeedsATargetPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	old := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: old, Name: "Old", TypeLine: o47Sorcery, Owner: opp.ID, Controller: opp.ID})
	})
	if _, err := o47Cast(t, g, "Go Blank", o47Sorcery, "{2}{B}", o47GoBlank, game.CastSpellParams{}); err == nil {
		t.Error("Go Blank was cast with no target player")
	}
	o47Pay(t, g, "BCC")
	o47MustCast(t, g, "Go Blank", o47Sorcery, "{2}{B}", o47GoBlank, game.CastSpellParams{Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(old) || len(opp.Graveyard.Cards) != 0 {
		t.Error("an empty hand stopped the graveyard being exiled")
	}
}

// --- Umbral Juke -----------------------------------------------------------------------

func TestS82OldB047UmbralJukeEdictLetsThePlayerChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	keep := effectstest.B12Creature(g, opp.ID, "Keeper", o47Bear, 3, 3)
	lose := effectstest.B12Creature(g, opp.ID, "Loser", o47Bear, 1, 1)
	o47Pay(t, g, "BCC")
	o47MustCast(t, g, "Umbral Juke", o47Instant, "{2}{B}", o47UmbralJuke, game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, lose)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(lose) || !g.Battlefield.Contains(keep) {
		t.Error("the player did not sacrifice the creature they chose")
	}
}

func TestS82OldB047UmbralJukeInklingToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o47Pay(t, g, "BCC")
	o47MustCast(t, g, "Umbral Juke", o47Instant, "{2}{B}", o47UmbralJuke, game.CastSpellParams{Modes: []int{1}})
	effectstest.PassPriorityAroundTable(t, g)
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Inkling")
	if n != 1 {
		t.Fatalf("%d Inkling tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Inkling" {
			if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 2 || tough != 1 {
				t.Errorf("Inkling is %d/%d, want 2/1", p, tough)
			}
			if !effectstest.EffectiveAbilitiesContain(t, g, c.InstanceID, "flying") {
				t.Error("the Inkling has no flying")
			}
		}
	}
}

func TestS82OldB047UmbralJukeRefusesTheEdictWithoutATargetAndAnEmptyBoardSacrificesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	o47Pay(t, g, "BCC")
	if _, err := o47Cast(t, g, "Umbral Juke", o47Instant, "{2}{B}", o47UmbralJuke, game.CastSpellParams{Modes: []int{0}}); err == nil {
		t.Error("the edict mode was cast with no target player")
	}
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	o47MustCast(t, g, "Umbral Juke", o47Instant, "{2}{B}", o47UmbralJuke, game.CastSpellParams{
		Modes: []int{0}, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil || !g.Battlefield.Contains(land) {
		t.Error("a player with no creature or planeswalker was asked to sacrifice")
	}
}

// --- Mercurial Transformation ----------------------------------------------------------

func TestS82OldB047MercurialTransformationBecomesTheChosenShape(t *testing.T) {
	for _, tc := range []struct {
		option  int
		subtype string
		pt      int
	}{{0, "Frog", 1}, {1, "Octopus", 4}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		opp := effectstest.Top100Opponent(g)
		rock := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Rock",
			TypeLine: "Artifact", OracleID: "", Owner: opp.ID, Controller: opp.ID})
		o47Pay(t, g, "UU")
		o47MustCast(t, g, "Mercurial Transformation", o47SorceryLesson, "{1}{U}", o47MercurialTransform,
			game.CastSpellParams{Targets: o47Target(rock)})
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerOptionPick(t, g, me.ID, tc.option)
		effectstest.PassPriorityAroundTable(t, g)
		types := effectstest.EffectiveTypes(t, g, rock)
		if len(types) != 1 || types[0] != "Creature" {
			t.Errorf("%s: types %v, want just Creature", tc.subtype, types)
		}
		var sub, col []string
		g.ReadSnapshot(func() {
			c, _ := g.LookupCardForEffect(rock)
			sub, col = c.Effective().Subtypes, c.Effective().Colors
		})
		if len(sub) != 1 || sub[0] != tc.subtype {
			t.Errorf("subtypes %v, want [%s]", sub, tc.subtype)
		}
		if len(col) != 1 || col[0] != "U" {
			t.Errorf("colours %v, want blue", col)
		}
		if p, tough := effectstest.PtOf(t, g, rock); p != tc.pt || tough != tc.pt {
			t.Errorf("%s is %d/%d, want %d/%d", tc.subtype, p, tough, tc.pt, tc.pt)
		}
	}
}

func TestS82OldB047MercurialTransformationStripsAbilitiesAndWearsOffAtEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	dragon := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dragon",
		TypeLine: "Creature — Dragon", Power: 5, Toughness: 5, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	o47Pay(t, g, "UU")
	o47MustCast(t, g, "Mercurial Transformation", o47SorceryLesson, "{1}{U}", o47MercurialTransform,
		game.CastSpellParams{Targets: o47Target(dragon)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectiveAbilitiesContain(t, g, dragon, "flying") {
		t.Error("the Frog still has flying")
	}
	if p, _ := effectstest.PtOf(t, g, dragon); p != 1 {
		t.Errorf("the Frog has power %d, want 1", p)
	}
	effectstest.EndTurn(t, g)
	if p, tough := effectstest.PtOf(t, g, dragon); p != 5 || tough != 5 {
		t.Errorf("after the turn the Dragon is %d/%d, want 5/5", p, tough)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, dragon, "flying") {
		t.Error("the Dragon did not get its flying back")
	}
}

func TestS82OldB047MercurialTransformationRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	o47Pay(t, g, "UU")
	if _, err := o47Cast(t, g, "Mercurial Transformation", o47SorceryLesson, "{1}{U}", o47MercurialTransform,
		game.CastSpellParams{Targets: o47Target(land)}); err == nil {
		t.Error("Mercurial Transformation targeted a land")
	}
}

// --- Academic Probation ------------------------------------------------------------------

func TestS82OldB047AcademicProbationBansTheNamedSpellForOpponentsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	o47Pay(t, g, "WW")
	o47MustCast(t, g, "Academic Probation", o47SorceryLesson, "{1}{W}", o47AcademicProbation, game.CastSpellParams{Modes: []int{0}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked for a card name")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)

	boltFor := func(p *game.Player, name string) error {
		id := uuid.New()
		g.WithWriteLock(func() {
			p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: o47Instant, OracleID: o47BoltOracle,
				Owner: p.ID, Controller: p.ID})
		})
		return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}})
	}
	if err := boltFor(opp, "Lightning Bolt"); !errors.Is(err, game.ErrCantCast) {
		t.Errorf("an opponent cast the named spell: %v", err)
	}
	if err := boltFor(me, "Lightning Bolt"); errors.Is(err, game.ErrCantCast) {
		t.Errorf("the caster is banned from the name too: %v", err)
	}
	if err := boltFor(opp, "Shock"); errors.Is(err, game.ErrCantCast) {
		t.Errorf("an unnamed spell was banned: %v", err)
	}
}

func TestS82OldB047AcademicProbationNameBanEndsAtYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := effectstest.Top100Opponent(g)
	o47Pay(t, g, "WW")
	o47MustCast(t, g, "Academic Probation", o47SorceryLesson, "{1}{W}", o47AcademicProbation, game.CastSpellParams{Modes: []int{0}})
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceCardName, me.ID)
	if prompt == nil {
		t.Fatal("no card name prompt")
	}
	if err := g.ResolveCardNameChoice(prompt.ID, me.ID, "Lightning Bolt"); err != nil {
		t.Fatalf("ResolveCardNameChoice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	banned := func() bool {
		var forbidden bool
		g.WithWriteLock(func() {
			err := g.CastGateLocked(opp.ID, game.Card{Name: "Lightning Bolt", TypeLine: o47Instant, OracleID: o47BoltOracle}, game.ZoneHand, game.CastSpellParams{})
			forbidden = errors.Is(err, game.ErrCantCast)
		})
		return forbidden
	}
	if !banned() {
		t.Fatal("the name was not banned this turn")
	}
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
		if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
			break
		}
	}
	if banned() {
		t.Error("the ban outlived the caster's next turn")
	}
}

func TestS82OldB047AcademicProbationLocksAPermanentUntilYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := effectstest.Top100Opponent(g)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", o47Bear, 2, 2)
	land := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	o47Pay(t, g, "WW")
	if _, err := o47Cast(t, g, "Academic Probation", o47SorceryLesson, "{1}{W}", o47AcademicProbation,
		game.CastSpellParams{Modes: []int{1}, Targets: o47Target(land)}); err == nil {
		t.Error("Academic Probation targeted a land")
	}
	o47MustCast(t, g, "Academic Probation", o47SorceryLesson, "{1}{W}", o47AcademicProbation,
		game.CastSpellParams{Modes: []int{1}, Targets: o47Target(theirs)})
	effectstest.PassPriorityAroundTable(t, g)
	var r game.Restriction
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(theirs); r = c.Effective().Restrictions })
	want := game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana
	if r&want != want {
		t.Errorf("restrictions %v, want can't attack, block or activate", r)
	}
	// It lasts through the opponent's own turn and ends as the caster's begins.
	for i := 0; i < len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
		if g.Seats[g.Turn.ActiveSeat].ID == g.Seats[0].ID && i > 0 {
			break
		}
	}
	g.ReadSnapshot(func() { c, _ := g.LookupCardForEffect(theirs); r = c.Effective().Restrictions })
	if r&want != 0 {
		t.Errorf("the restrictions %v outlived the caster's next turn", r)
	}
}

// --- Harness Infinity --------------------------------------------------------------------

func TestS82OldB047HarnessInfinityExchangesHandAndGraveyardThenExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	g.WithWriteLock(func() { me.Hand.Cards = nil; me.Graveyard.Cards = nil })
	h1 := effectstest.HandCardForTest(me, "Hand One", o47Sorcery, "")
	h2 := effectstest.HandCardForTest(me, "Hand Two", o47Sorcery, "")
	var yard []uuid.UUID
	for _, n := range []string{"Yard One", "Yard Two", "Yard Three"} {
		id := uuid.New()
		yard = append(yard, id)
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: n, TypeLine: o47Sorcery, Owner: me.ID, Controller: me.ID})
		})
	}
	o47Pay(t, g, "BBBGGGC")
	spell := o47MustCast(t, g, "Harness Infinity", o47Instant, "{1}{B}{B}{B}{G}{G}{G}", o47HarnessInfinity, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != 3 || len(me.Graveyard.Cards) != 2 {
		t.Fatalf("hand %d, graveyard %d, want 3 and 2", len(me.Hand.Cards), len(me.Graveyard.Cards))
	}
	for _, id := range yard {
		if !me.Hand.Contains(id) {
			t.Errorf("graveyard card %s is not in the hand", id)
		}
	}
	for _, id := range []uuid.UUID{h1, h2} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("hand card %s is not in the graveyard", id)
		}
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("Harness Infinity did not exile itself")
	}
}

func TestS82OldB047HarnessInfinityRefusesAPoolThatCannotPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	o47Pay(t, g, "BBGGGC") // one black short of {1}{B}{B}{B}{G}{G}{G}
	if _, err := o47Cast(t, g, "Harness Infinity", o47Instant, "{1}{B}{B}{B}{G}{G}{G}", o47HarnessInfinity, game.CastSpellParams{}); err == nil {
		t.Error("Harness Infinity was cast without its mana")
	}
}
