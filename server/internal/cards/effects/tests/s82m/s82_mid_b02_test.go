package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mid_b02_test.go — Innistrad: Midnight Hunt, plain instants and
// removal (S82, #5497). Each test plays the real catalog entry.

const (
	s82MidBlessedDefianceOracle = "04e3d36f-5dec-422c-a371-15e135fdface"
	s82MidFlareOfFaithOracle    = "13ebe447-1f8f-4b12-9e35-9d334a863c19"
	s82MidPlayWithFireOracle    = "5ffc1e77-152b-4a82-8ea8-2c79cc31d5ef"
	s82MidPacksBetrayalOracle   = "5b7abe71-0be3-4dd4-880e-88b65b47ea4d"
	s82MidFoulPlayOracle        = "773c7a1e-0f59-4606-a2fc-c47bda760ecd"
	s82MidLunarFrenzyOracle     = "d95fc2cb-5e49-4a22-a93e-d21cf64dcd84"
	s82MidDefendOracle          = "c079a636-cfd4-49cd-bcd4-54375b3fa355"
	s82MidFadingHopeOracle      = "64883ef2-c536-42e2-b52f-cfd1984fcdfd"
	s82MidStolenVitalityOracle  = "bd7454f3-57bc-44f4-bd58-54e883c699fd"
	s82MidDefenestrateOracle    = "caf840e8-c1d3-4570-8542-432a765c09ff"
	s82MidRazeOracle            = "9286ddc5-8151-4129-aaf6-2a2abe7047e5"
	s82MidSungoldOracle         = "2a33d0ea-9b19-4fde-8e7a-2a9d17f716ea"
	s82MidGeistwaveOracle       = "14d5ed90-0aa7-48f0-9017-83912b9e89d1"
)

func TestS82MidB02EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82MidBlessedDefianceOracle: "Blessed Defiance",
		s82MidFlareOfFaithOracle:    "Flare of Faith",
		s82MidPlayWithFireOracle:    "Play with Fire",
		s82MidPacksBetrayalOracle:   "Pack's Betrayal",
		s82MidFoulPlayOracle:        "Foul Play",
		s82MidLunarFrenzyOracle:     "Lunar Frenzy",
		s82MidDefendOracle:          "Defend the Celestus",
		s82MidFadingHopeOracle:      "Fading Hope",
		s82MidStolenVitalityOracle:  "Stolen Vitality",
		s82MidDefenestrateOracle:    "Defenestrate",
		s82MidRazeOracle:            "Raze the Effigy",
		s82MidSungoldOracle:         "Sungold Barrage",
		s82MidGeistwaveOracle:       "Geistwave",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// s82MidCastNow puts the card in the active seat's hand and casts it where
// the game stands, with no walk to a main phase (combat spells).
func s82MidCastNow(t *testing.T, g *game.Game, seat *game.Player, name, typeLine, oracle, cost string, params game.CastSpellParams) error {
	t.Helper()
	id := uuid.New()
	seat.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: seat.ID, Controller: seat.ID})
	params.Strict = true
	return g.CastSpell(seat.ID, id, params)
}

func s82MidMe(g *game.Game) *game.Player { return g.Seats[g.Turn.ActiveSeat] }
func s82MidOpp(g *game.Game) *game.Player {
	return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestS82MidB02BlessedDefianceLifelinkAndSpiritWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Blessed Defiance", "Instant", s82MidBlessedDefianceOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("target is %d/%d, want 4/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("target has no lifelink")
	}
	if effectstest.TmtB1Count(g, me.ID, "Spirit") != 0 {
		t.Fatal("a Spirit came before the creature died")
	}
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB1Count(g, me.ID, "Spirit"); got != 1 {
		t.Errorf("%d Spirits, want 1", got)
	}
}

func TestS82MidB02BlessedDefianceRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, s82MidOpp(g).ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Blessed Defiance", "Instant", s82MidBlessedDefianceOracle, effectstest.B16TargetCard(theirs)); err == nil {
		t.Error("an opponent's creature was accepted")
	}
}

func TestS82MidB02FlareOfFaithHumanGetsPlusThreeAndIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	human := effectstest.B12Creature(g, s82MidMe(g).ID, "Villager", "Creature — Human Peasant", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Flare of Faith", "Instant", s82MidFlareOfFaithOracle, effectstest.B16TargetCard(human))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, human); p != 4 || tough != 4 {
		t.Errorf("Human is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, human, "indestructible") {
		t.Error("Human has no indestructible")
	}
}

func TestS82MidB02FlareOfFaithNonHumanGetsPlusTwoOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, s82MidOpp(g).ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Flare of Faith", "Instant", s82MidFlareOfFaithOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("Bear is %d/%d, want 4/4", p, tough)
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "indestructible") {
		t.Error("a non-Human gained indestructible")
	}
}

func TestS82MidB02PlayWithFireScrysOnlyWhenAPlayerIsDealtDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := s82MidMe(g), s82MidOpp(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Play with Fire", "Instant", s82MidPlayWithFireOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-2 {
		t.Errorf("life %d -> %d, want -2", before, opp.Life)
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry after a player was dealt damage")
	}
}

func TestS82MidB02PlayWithFireOnACreatureDoesNotScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, s82MidOpp(g).ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Play with Fire", "Instant", s82MidPlayWithFireOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("scried after damaging a creature")
	}
	if effectstest.TmtB1Count(g, s82MidOpp(g).ID, "Bear") != 0 {
		t.Error("the 2/2 survived 2 damage")
	}
}

func TestS82MidB02PacksBetrayalStealsUntapsHastesAndScrysWithAWolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.PushTypedCard(g, me.ID, "Wolf", "Creature — Wolf", "")
	theirs := effectstest.B12Creature(g, s82MidOpp(g).ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pack's Betrayal", "Sorcery", s82MidPacksBetrayalOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindCardAnywhere(t, g, theirs); c.Controller != me.ID {
		t.Error("control did not change")
	}
	if !effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("no haste")
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry 2 while controlling a Wolf")
	}
}

func TestS82MidB02PacksBetrayalDoesNotScryWithoutAWolfOrWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	theirs := effectstest.B12Creature(g, s82MidOpp(g).ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pack's Betrayal", "Sorcery", s82MidPacksBetrayalOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Error("scried without a Wolf or Werewolf")
	}
}

func TestS82MidB02PacksBetrayalCountsTheStolenWerewolf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	effectstest.SeedLibrary(me, "A", "B", "C")
	theirs := effectstest.B12Creature(g, s82MidOpp(g).ID, "Their Werewolf", "Creature — Werewolf", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Pack's Betrayal", "Sorcery", s82MidPacksBetrayalOracle, effectstest.B16TargetCard(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("the stolen Werewolf did not turn the rider on")
	}
}

func TestS82MidB02FoulPlayDestroysSmallCreatureAndInvestigates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	small := effectstest.B12Creature(g, s82MidOpp(g).ID, "Small", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Foul Play", "Sorcery", s82MidFoulPlayOracle, effectstest.B16TargetCard(small))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, s82MidOpp(g).ID, "Small") != 0 {
		t.Error("the creature survived")
	}
	if got := effectstest.TmtB1Count(g, me.ID, "Clue"); got != 1 {
		t.Errorf("%d Clues, want 1", got)
	}
}

func TestS82MidB02FoulPlayRefusesPowerThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	big := effectstest.B12Creature(g, s82MidOpp(g).ID, "Big", "Creature — Ogre", 3, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Foul Play", "Sorcery", s82MidFoulPlayOracle, effectstest.B16TargetCard(big)); err == nil {
		t.Error("a power-3 creature was accepted")
	}
}

func TestS82MidB02LunarFrenzyGivesXPowerFirstStrikeAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, s82MidMe(g).ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastXSpell(t, g, "Lunar Frenzy", "Instant", s82MidLunarFrenzyOracle, "{X}{R}", 3, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 2 {
		t.Errorf("target is %d/%d, want 5/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") || !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("missing first strike or trample")
	}
}

func TestS82MidB02LunarFrenzyRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, s82MidOpp(g).ID, "Theirs", "Creature — Bear", 2, 2)
	if err := s82MidCastNow(t, g, s82MidMe(g), "Lunar Frenzy", "Instant", s82MidLunarFrenzyOracle, "{X}{R}", game.CastSpellParams{Targets: effectstest.B16TargetCard(theirs)}); err == nil {
		t.Error("an opponent's creature was accepted")
	}
}

func TestS82MidB02DefendTheCelestusDividesThreeCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Defend the Celestus", TypeLine: "Instant", OracleID: s82MidDefendOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 2, b: 1}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, a); p != 3 {
		t.Errorf("A power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, b); p != 2 {
		t.Errorf("B power %d, want 2", p)
	}
}

func TestS82MidB02DefendTheCelestusRefusesABadDistribution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := s82MidMe(g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	err := s82MidCastNow(t, g, me, "Defend the Celestus", "Instant", s82MidDefendOracle, "", game.CastSpellParams{Targets: effectstest.CardRefs(a), Distribution: map[uuid.UUID]int{a: 2}})
	if err == nil {
		t.Error("a distribution of 2 counters was accepted")
	}
}

func TestS82MidB02FadingHopeScrysForCheapCreaturesOnly(t *testing.T) {
	for _, tc := range []struct {
		cost string
		scry bool
	}{{"{2}{G}", true}, {"{3}{G}", false}} {
		g := effectstest.NewCatalogGame(t)
		me, opp := s82MidMe(g), s82MidOpp(g)
		effectstest.SeedLibrary(me, "A", "B", "C")
		id := effectstest.PushTypedCard(g, opp.ID, "Target", "Creature — Bear", tc.cost)
		effectstest.CastCatalogSpell(t, g, "Fading Hope", "Instant", s82MidFadingHopeOracle, effectstest.B16TargetCard(id))
		effectstest.PassPriorityAroundTable(t, g)
		if opp.Hand.Size() == 0 || effectstest.TmtB1Count(g, opp.ID, "Target") != 0 {
			t.Errorf("%s: the creature was not returned to hand", tc.cost)
		}
		if got := effectstest.ScryChoiceFor(g, me.ID) != nil; got != tc.scry {
			t.Errorf("%s: scry = %v, want %v", tc.cost, got, tc.scry)
		}
	}
}

func TestS82MidB02StolenVitalityTrampleOnYourTurnFirstStrikeOtherwise(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, s82MidOpp(g).ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Stolen Vitality", "Instant", s82MidStolenVitalityOracle, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 5 || tough != 3 {
		t.Errorf("target is %d/%d, want 5/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") || effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("on your turn it should gain trample only")
	}
}

func TestS82MidB02StolenVitalityFirstStrikeOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := s82MidOpp(g)
	bear := effectstest.B12Creature(g, s82MidMe(g).ID, "Bear", "Creature — Bear", 2, 2)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := s82MidCastNow(t, g, opp, "Stolen Vitality", "Instant", s82MidStolenVitalityOracle, "", game.CastSpellParams{Targets: effectstest.B16TargetCard(bear)}); err != nil {
		t.Skipf("the opponent cannot cast with priority here: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") || effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("on another player's turn it should gain first strike only")
	}
}

func TestS82MidB02DefenestrateDestroysAGrounderAndRefusesAFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := s82MidOpp(g)
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 1, Toughness: 1,
		Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID,
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Defenestrate", "Instant", s82MidDefenestrateOracle, effectstest.B16TargetCard(flier)); err == nil {
		t.Error("a creature with flying was accepted")
	}
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Defenestrate", "Instant", s82MidDefenestrateOracle, effectstest.B16TargetCard(ground))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, opp.ID, "Ground") != 0 {
		t.Error("the creature survived")
	}
}

func TestS82MidB02SungoldBarrageNeedsToughnessFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := s82MidOpp(g)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 3)
	if err := effectstest.CastCatalogSpellErr(t, g, "Sungold Barrage", "Instant", s82MidSungoldOracle, effectstest.B16TargetCard(small)); err == nil {
		t.Error("a toughness-3 creature was accepted")
	}
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Wall", 0, 4)
	effectstest.CastCatalogSpell(t, g, "Sungold Barrage", "Instant", s82MidSungoldOracle, effectstest.B16TargetCard(big))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, opp.ID, "Big") != 0 {
		t.Error("the toughness-4 creature survived")
	}
}

func TestS82MidB02RazeTheEffigyDestroysAnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := s82MidOpp(g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	if err := s82MidCastNow(t, g, s82MidMe(g), "Raze the Effigy", "Instant", s82MidRazeOracle, "", game.CastSpellParams{Modes: []int{0}, Targets: effectstest.B16TargetCard(rock)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TmtB1Count(g, opp.ID, "Rock") != 0 {
		t.Error("the artifact survived")
	}
}

func TestS82MidB02RazeTheEffigyPumpsAnAttackerAndRefusesANonAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := s82MidMe(g), s82MidOpp(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idler := effectstest.B12Creature(g, me.ID, "Idler", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if err := s82MidCastNow(t, g, me, "Raze the Effigy", "Instant", s82MidRazeOracle, "", game.CastSpellParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(idler)}); err == nil {
		t.Error("a non-attacking creature was accepted")
	}
	if err := s82MidCastNow(t, g, me, "Raze the Effigy", "Instant", s82MidRazeOracle, "", game.CastSpellParams{Modes: []int{1}, Targets: effectstest.B16TargetCard(attacker)}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, attacker); p != 4 || tough != 4 {
		t.Errorf("attacker is %d/%d, want 4/4", p, tough)
	}
}

func TestS82MidB02GeistwaveDrawsOnlyForYourOwnPermanent(t *testing.T) {
	for _, mine := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := s82MidMe(g), s82MidOpp(g)
		owner := opp
		if mine {
			owner = me
		}
		perm := effectstest.B12Permanent(g, owner.ID, "Thing", "Artifact")
		land := effectstest.PushTypedCard(g, opp.ID, "Island", "Basic Land — Island", "")
		if err := effectstest.CastCatalogSpellErr(t, g, "Geistwave", "Instant", s82MidGeistwaveOracle, effectstest.B16TargetCard(land)); err == nil {
			t.Fatal("a land was accepted")
		}
		hand := me.Hand.Size()
		effectstest.CastCatalogSpell(t, g, "Geistwave", "Instant", s82MidGeistwaveOracle, effectstest.B16TargetCard(perm))
		effectstest.PassPriorityAroundTable(t, g)
		want := 0 // the permanent stays with its owner; no draw
		if mine {
			want = 2 // the permanent returns to my hand, and I draw
		}
		if got := me.Hand.Size() - hand; got != want {
			t.Errorf("mine=%v: hand grew by %d, want %d", mine, got, want)
		}
		if effectstest.TmtB1Count(g, owner.ID, "Thing") != 0 {
			t.Errorf("mine=%v: the permanent was not returned", mine)
		}
	}
}
