package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_vow_b04_test.go — Crimson Vow batch 4 (#5495): Blood tokens,
// sacrificing and counting them.

const (
	vb4BloodHypnotist = "32845ed6-2328-462e-be3a-579589748588"
	vb4Socialite      = "7e165b4e-4c8f-4a0f-bd6e-7b0a15b12d1d"
	vb4GluttonousGst  = "2d338aa5-22bc-4cc4-a1e0-cc5018ce001e"
	vb4WeddingSecur   = "86360908-6b8b-4c1c-9e10-4f3d389aaa3e"
	vb4Statuette      = "35553c2a-0150-49e4-81da-f356c6a17253"
	vb4Forebear       = "940080ce-c504-48c8-a763-326fef907217"
	vb4Bloodseeker    = "7aef5d2e-b66d-4822-b563-44d21d008272"
	vb4Anje           = "1829f1dc-1fa9-4361-b318-d4dee280e6fd"
	vb4Strefan        = "3cc745d2-5021-43ff-ace8-35011011e62a"
	vb4Harvester      = "419a8ab2-cff6-4811-a559-8791c9bb3d8e"
	vb4Purveyor       = "7292272d-6511-466b-aadd-c4fe7ed8df09"
)

func TestVowB04CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		vb4BloodHypnotist:     "Blood Hypnotist",
		vb4Socialite:          "Bloodcrazed Socialite",
		vb4GluttonousGst:      "Gluttonous Guest",
		vb4WeddingSecur:       "Wedding Security",
		vb4Statuette:          "Sanguine Statuette",
		vb4Forebear:           "Falkenrath Forebear",
		vb4Bloodseeker:        "Restless Bloodseeker",
		vb4Bloodseeker + "#1": "Bloodsoaked Reveler",
		vb4Anje:               "Anje, Maid of Dishonor",
		vb4Strefan:            "Strefan, Maurer Progenitor",
		vb4Harvester:          "Bloodtithe Harvester",
		vb4Purveyor:           "Bloodvial Purveyor",
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

// vb4Blood puts n Blood tokens onto the battlefield under owner and returns
// their ids.
func vb4Blood(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, effects.BloodToken(), n) })
	var ids []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Blood" && c.Controller == owner {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

func vb4BloodCount(g *game.Game, owner uuid.UUID) int {
	n, _ := effectstest.CountTokensNamed(g, owner, "Blood")
	return n
}

func vb4Sacrifice(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(id) })
}

// vb4Cast casts a catalog creature from hand for free and resolves it and
// its enters trigger.
func vb4Cast(t *testing.T, g *game.Game, me *game.Player, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.HandCardForTest(me, name, typeLine, oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

// --- Gluttonous Guest ---------------------------------------------------

func TestVowB04GluttonousGuestMakesBloodAndGainsLifeWhenYouSacrificeOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vb4Cast(t, g, me, "Gluttonous Guest", "Creature — Vampire", vb4GluttonousGst)
	blood := vb4Blood(g, me.ID, 0)
	if len(blood) != 1 {
		t.Fatalf("Blood after the Guest entered = %d, want 1", len(blood))
	}
	life := me.Life
	vb4Sacrifice(g, blood[0])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
	// Not a Blood: no life.
	relic := effectstest.B12Permanent(g, me.ID, "Relic", "Artifact")
	vb4Sacrifice(g, relic)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("sacrificing a non-Blood gained life: %d", me.Life)
	}
}

func TestVowB04GluttonousGuestIgnoresAnOpponentsSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Gluttonous Guest", "Creature — Vampire", vb4GluttonousGst, 1, 4)
	theirs := vb4Blood(g, opp.ID, 1)
	life := me.Life
	vb4Sacrifice(g, theirs[0])
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
}

// --- Wedding Security ---------------------------------------------------

func TestVowB04WeddingSecuritySacrificesBloodForACounterAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	sec := effectstest.B12Push(g, me.ID, "Wedding Security", "Creature — Vampire Soldier", vb4WeddingSecur, 4, 4)
	blood := vb4Blood(g, me.ID, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, sec)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, blood[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blood[0]) {
		t.Error("the Blood token was not sacrificed")
	}
	if got := effectstest.CountersOn(g, sec, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters %d, want 1", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestVowB04WeddingSecurityDeclinedOrWithoutBloodDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	sec := effectstest.B12Push(g, me.ID, "Wedding Security", "Creature — Vampire Soldier", vb4WeddingSecur, 4, 4)
	blood := vb4Blood(g, me.ID, 1)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, sec)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(blood[0]) || effectstest.CountersOn(g, sec, game.CounterPlusOne) != 0 || me.Hand.Size() != hand {
		t.Error("declining the sacrifice still paid off")
	}
}

// --- Bloodcrazed Socialite ----------------------------------------------

func TestVowB04BloodcrazedSocialiteMakesBloodAndPumpsOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soc := vb4Cast(t, g, me, "Bloodcrazed Socialite", "Creature — Vampire", vb4Socialite)
	blood := vb4Blood(g, me.ID, 0)
	if len(blood) != 1 {
		t.Fatalf("Blood = %d, want 1", len(blood))
	}
	if !effectstest.HasEffectiveKeyword(t, g, soc, "menace") {
		t.Error("no menace")
	}
	effectstest.Unsick(g, soc)
	base := effectstest.EffectivePower(t, g, soc)
	effectstest.DeclareAttack(t, g, opp.ID, soc)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, blood[0])
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blood[0]) {
		t.Error("the Blood token was not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g, soc); got != base+2 {
		t.Errorf("power %d, want %d", got, base+2)
	}
}

// --- Blood Hypnotist ----------------------------------------------------

func TestVowB04BloodHypnotistCantBlockAndStopsATargetOncePerTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hyp := effectstest.B12Push(g, me.ID, "Blood Hypnotist", "Creature — Vampire", vb4BloodHypnotist, 3, 3)
	if effectstest.AuraRestrictions(t, g, hyp)&game.CantBlock == 0 {
		t.Error("the Hypnotist can block")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	blood := vb4Blood(g, me.ID, 3)
	// Two Blood tokens in one batch: one trigger.
	g.WithWriteLock(func() { _ = g.SacrificeAllForEffect(uuid.Nil, blood[:2]) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AuraRestrictions(t, g, bear)&game.CantBlock == 0 {
		t.Fatal("the targeted creature can still block")
	}
	// A second sacrifice the same turn does not trigger again.
	other := effectstest.B12Creature(g, opp.ID, "Other Bear", "Creature — Bear", 2, 2)
	vb4Sacrifice(g, blood[2])
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the Hypnotist triggered twice in a turn")
	}
	if effectstest.AuraRestrictions(t, g, other)&game.CantBlock != 0 {
		t.Error("a second creature can't block")
	}
}

// --- Sanguine Statuette -------------------------------------------------

func TestVowB04SanguineStatuetteBecomesAVampireCreatureOnABloodSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	st := effectstest.B12Push(g, me.ID, "Sanguine Statuette", "Artifact", vb4Statuette, 0, 0)
	blood := vb4Blood(g, me.ID, 1)
	vb4Sacrifice(g, blood[0])
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.BattlefieldCardCopy(t, g, st)
	if !c.IsCreature() || !c.IsArtifact() || !c.HasSubtypeWord("Vampire") {
		t.Fatalf("type line %q, want a Vampire artifact creature", c.TypeLine)
	}
	if got := effectstest.EffectivePower(t, g, st); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, st, "haste") {
		t.Error("no haste")
	}
}

func TestVowB04SanguineStatuetteDeclinedStaysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	st := effectstest.B12Push(g, me.ID, "Sanguine Statuette", "Artifact", vb4Statuette, 0, 0)
	blood := vb4Blood(g, me.ID, 1)
	vb4Sacrifice(g, blood[0])
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldCardCopy(t, g, st).IsCreature() {
		t.Error("declining still animated it")
	}
}

// --- Anje, Maid of Dishonor ---------------------------------------------

func TestVowB04AnjeMakesBloodOncePerTurnForVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vb4Cast(t, g, me, "Anje, Maid of Dishonor", "Legendary Creature — Vampire", vb4Anje)
	if got := vb4BloodCount(g, me.ID); got != 1 {
		t.Fatalf("Blood after Anje entered = %d, want 1", got)
	}
	vb4Cast(t, g, me, "Vampire Friend", "Creature — Vampire", "")
	if got := vb4BloodCount(g, me.ID); got != 1 {
		t.Errorf("a second Vampire the same turn made Blood: %d", got)
	}
}

func TestVowB04AnjeIgnoresNonVampires(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Anje, Maid of Dishonor", "Legendary Creature — Vampire", vb4Anje, 4, 5)
	vb4Cast(t, g, me, "Human Friend", "Creature — Human", "")
	if got := vb4BloodCount(g, me.ID); got != 0 {
		t.Errorf("a Human made Blood: %d", got)
	}
	vb4Cast(t, g, me, "Vampire Friend", "Creature — Vampire", "")
	if got := vb4BloodCount(g, me.ID); got != 1 {
		t.Errorf("another Vampire made %d Blood, want 1", got)
	}
}

func TestVowB04AnjeDrainsBySacrificingABloodOrACreatureButNotHerself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	anje := effectstest.B12Push(g, me.ID, "Anje, Maid of Dishonor", "Legendary Creature — Vampire", vb4Anje, 4, 5)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, anje, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with nothing else to sacrifice")
	}
	blood := vb4Blood(g, me.ID, 1)
	myLife, theirLife := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, anje, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: blood}); err != nil {
		t.Fatalf("activate with a Blood token: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife+2 || opp.Life != theirLife-2 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, myLife+2, theirLife-2)
	}
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	effectstest.B06AddMana(me, "B", "B")
	if err := g.ActivateCatalogAbility(me.ID, anje, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate with a creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != theirLife-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, theirLife-4)
	}
}

// --- Bloodtithe Harvester -----------------------------------------------

func TestVowB04BloodtitheHarvesterShrinksByTwiceYourBloodAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hv := vb4Cast(t, g, me, "Bloodtithe Harvester", "Creature — Vampire", vb4Harvester)
	if got := vb4BloodCount(g, me.ID); got != 1 {
		t.Fatalf("Blood = %d, want 1", got)
	}
	vb4Blood(g, me.ID, 1) // two Blood: X = 4
	effectstest.Unsick(g, hv)
	bear := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 6, 6)
	if err := g.ActivateCatalogAbility(me.ID, hv, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(hv) {
		t.Error("the Harvester was not sacrificed")
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2 (6-4)", got)
	}
}

func TestVowB04BloodtitheHarvesterRefusesOutsideTheMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hv := effectstest.B12Push(g, me.ID, "Bloodtithe Harvester", "Creature — Vampire", vb4Harvester, 3, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, hv, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}}); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Bloodvial Purveyor -------------------------------------------------

func TestVowB04BloodvialPurveyorGivesTheCasterABloodToken(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Bloodvial Purveyor", "Creature — Vampire", vb4Purveyor, 5, 6)
	effectstest.AdvanceToMainOf(t, g, 1)
	id := effectstest.HandCardForTest(opp, "Test Sorcery", "Sorcery", "")
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := vb4BloodCount(g, opp.ID); got != 1 {
		t.Errorf("the caster's Blood = %d, want 1", got)
	}
	if got := vb4BloodCount(g, me.ID); got != 0 {
		t.Errorf("the Purveyor's controller got %d Blood", got)
	}
	// Its controller casting a spell does not trigger it.
	effectstest.AdvanceToMainOf(t, g, 0)
	mine := effectstest.HandCardForTest(me, "My Sorcery", "Sorcery", "")
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := vb4BloodCount(g, me.ID); got != 0 {
		t.Errorf("my own spell made %d Blood", got)
	}
}

func TestVowB04BloodvialPurveyorGainsPowerPerBloodTheDefenderControls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pv := effectstest.B12Push(g, me.ID, "Bloodvial Purveyor", "Creature — Vampire", vb4Purveyor, 5, 6)
	vb4Blood(g, opp.ID, 3)
	vb4Blood(g, me.ID, 2) // not counted
	effectstest.DeclareAttack(t, g, opp.ID, pv)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, pv); got != 8 {
		t.Errorf("power %d, want 8", got)
	}
}

// --- Falkenrath Forebear ------------------------------------------------

func TestVowB04FalkenrathForebearReturnsFromTheGraveyardForTwoBlood(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	fb := effectstest.GyCardWithOracle(me, "Falkenrath Forebear", "Creature — Vampire", vb4Forebear)
	blood := vb4Blood(g, me.ID, 2)
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, fb, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: blood[:1]}); err == nil {
		t.Fatal("activated with one Blood token")
	}
	effectstest.B06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, fb, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: blood}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fb) {
		t.Error("the Forebear did not return")
	}
	if got := vb4BloodCount(g, me.ID); got != 0 {
		t.Errorf("Blood left %d, want 0", got)
	}
}

func TestVowB04FalkenrathForebearCantBlockAndMakesBloodOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fb := effectstest.B12Push(g, me.ID, "Falkenrath Forebear", "Creature — Vampire", vb4Forebear, 3, 1)
	if effectstest.AuraRestrictions(t, g, fb)&game.CantBlock == 0 {
		t.Error("the Forebear can block")
	}
	effectstest.DeclareAttack(t, g, opp.ID, fb)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vb4BloodCount(g, me.ID); got != 1 {
		t.Errorf("Blood %d, want 1", got)
	}
}

// --- Restless Bloodseeker // Bloodsoaked Reveler ------------------------

// vb4Seeker puts the double-faced card onto the battlefield as the importer
// builds it: two faces, `face` up.
func vb4Seeker(g *game.Game, owner uuid.UUID, face int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), OracleID: vb4Bloodseeker, Layout: game.LayoutTransform,
		Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Restless Bloodseeker", TypeLine: "Creature — Vampire", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 1, Toughness: 3},
			{Name: "Bloodsoaked Reveler", TypeLine: "Creature — Vampire", Colors: []string{"B"}, Power: 3, Toughness: 3},
		},
	}
	c.SetFace(face)
	return effectstest.PushBattlefieldCardWithTimestamp(g, c)
}

func TestVowB04RestlessBloodseekerMakesBloodOnlyIfYouGainedLife(t *testing.T) {
	for _, gained := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		vb4Seeker(g, me.ID, 0)
		if gained {
			g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 2) })
		}
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if gained {
			want = 1
		}
		if got := vb4BloodCount(g, me.ID); got != want {
			t.Errorf("gained life %v: Blood %d, want %d", gained, got, want)
		}
	}
}

func TestVowB04RestlessBloodseekerTransformsForTwoBloodAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	rb := vb4Seeker(g, me.ID, 0)
	blood := vb4Blood(g, me.ID, 2)
	if err := g.ActivateCatalogAbility(me.ID, rb, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: blood[:1]}); err == nil {
		t.Fatal("transformed for one Blood token")
	}
	if err := g.ActivateCatalogAbility(me.ID, rb, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: blood}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCardCopy(t, g, rb); got.ActiveFace != 1 {
		t.Errorf("not transformed: %+v", got.Name)
	}
	if got := vb4BloodCount(g, me.ID); got != 0 {
		t.Errorf("Blood left %d", got)
	}
	// Not at instant speed.
	rb2 := vb4Seeker(g, me.ID, 0)
	more := vb4Blood(g, me.ID, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.ActivateCatalogAbility(me.ID, rb2, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: more}); err == nil {
		t.Error("transformed outside a main phase")
	}
}

func TestVowB04BloodsoakedRevelerDrainsForFiveMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	me, opp := g.Seats[0], g.Seats[1]
	rv := vb4Seeker(g, me.ID, 1)
	effectstest.B06AddMana(me, "B", "B", "B", "B", "B")
	mine, theirs := me.Life, opp.Life
	if err := g.ActivateCatalogAbility(me.ID, rv, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != mine+2 || opp.Life != theirs-2 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, mine+2, theirs-2)
	}
}

// --- Strefan, Maurer Progenitor -----------------------------------------

func TestVowB04StrefanMakesBloodForEachPlayerWhoLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Strefan, Maurer Progenitor", "Legendary Creature — Vampire Noble", vb4Strefan, 3, 2)
	effectstest.LoseLife(t, g, me, 1)
	effectstest.LoseLife(t, g, opp, 2)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := vb4BloodCount(g, me.ID); got != 2 {
		t.Errorf("Blood %d, want 2 (two players lost life)", got)
	}
}

func TestVowB04StrefanSacrificesTwoBloodToPutAVampireOntoTheBattlefieldAttacking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	st := effectstest.B12Push(g, me.ID, "Strefan, Maurer Progenitor", "Legendary Creature — Vampire Noble", vb4Strefan, 3, 2)
	blood := vb4Blood(g, me.ID, 2)
	vamp := effectstest.HandCardForTest(me, "Hand Vampire", "Creature — Vampire", "")
	effectstest.DeclareAttack(t, g, opp.ID, st)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, blood...)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, vamp)
	if effectstest.LatestOptionPickFor(g, me.ID) != nil {
		effectstest.AnswerOptionPick(t, g, me.ID, 0)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if vb4BloodCount(g, me.ID) != 0 {
		t.Error("the Blood tokens were not sacrificed")
	}
	c := effectstest.BattlefieldCardCopy(t, g, vamp)
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the Vampire is tapped=%v attacking=%v, want tapped and attacking", c.Tapped, c.AttackingTarget)
	}
	if !effectstest.HasEffectiveKeyword(t, g, vamp, "indestructible") {
		t.Error("no indestructible")
	}
}

func TestVowB04StrefanWithoutTwoBloodAsksNothingMore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	st := effectstest.B12Push(g, me.ID, "Strefan, Maurer Progenitor", "Legendary Creature — Vampire Noble", vb4Strefan, 3, 2)
	vb4Blood(g, me.ID, 1)
	vamp := effectstest.HandCardForTest(me, "Hand Vampire", "Creature — Vampire", "")
	effectstest.DeclareAttack(t, g, opp.ID, st)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(vamp) || effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		t.Error("Strefan carried on with fewer than two Blood tokens")
	}
}
