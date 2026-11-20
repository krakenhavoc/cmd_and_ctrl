package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Commander-deck wave 2, batch 16 (#5368): land ramp, landfall, untap lands.

const (
	c2b16RejuvenatorOracle = "71d14755-b81d-4aec-b94d-1637884a5844"
	c2b16TanukiOracle      = "b2d19232-9a80-4ea2-9aa7-c47b4719106c"
	c2b16BeanstalkOracle   = "22c75ad5-237c-4dbd-bbc4-3058a3db408f"
	c2b16GrazerOracle      = "d18a0815-59d3-4667-b52b-9acda741215e"
	c2b16CoralhelmOracle   = "c450f6b7-bb2a-4b8c-aec7-d27555137b63"
	c2b16StomperOracle     = "e3fc10d2-2733-4265-8275-112c285fd11a"
	c2b16RangersOracle     = "aece819e-3678-48bc-9677-d06ebfa58b4e"
	c2b16DevourerOracle    = "56314980-5f18-4b93-a9a0-6002ee8a880c"
	c2b16OnTheTrailOracle  = "5a060788-a601-4c01-8341-73398b497a85"
)

// c2b16Lands puts n untapped basic Forests onto a player's battlefield.
func c2b16Lands(g *game.Game, p *game.Player, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: p.ID, Controller: p.ID}))
	}
	return ids
}

func TestC2B16ElvishRejuvenatorPutsALandTappedAndBottomsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	lib := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell(), effectstest.C2b15Forest(), effectstest.C2b15Bear(), effectstest.C2b15Bear())
	effectstest.DftABCast(t, g, me, game.Card{Name: "Elvish Rejuvenator", TypeLine: "Creature — Elf Druid",
		OracleID: c2b16RejuvenatorOracle, Power: 1, Toughness: 1})
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil || ch.ChooseMin != 0 || ch.ChooseMax != 1 {
		t.Fatalf("want an optional pick of one land, got %+v", ch)
	}
	for _, id := range ch.ChooseCards {
		if id != lib[1] && id != lib[3] {
			t.Errorf("a nonland card was offered: %s", id)
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lib[1])
	if !g.Battlefield.Contains(lib[1]) || !effectstest.C2b16IsTapped(g, lib[1]) {
		t.Error("the land is not on the battlefield tapped")
	}
	if me.Library.Contains(lib[5]) == false {
		t.Error("the sixth card left the library")
	}
	for _, i := range []int{0, 2, 3, 4} {
		if !me.Library.Contains(lib[i]) {
			t.Errorf("card %d is not back in the library", i)
		}
	}
	// The five looked at are the bottom five of the library now, the sixth on top.
	if top := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; top != lib[5] {
		t.Errorf("the top card is %s, want the sixth card", top)
	}
}

func TestC2B16ArborealGrazerMayPutALandFromHandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.DftABCast(t, g, me, game.Card{Name: "Arboreal Grazer", TypeLine: "Creature — Sloth Beast",
		OracleID: c2b16GrazerOracle, Power: 0, Toughness: 3})
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !effectstest.C2b16IsTapped(g, land) {
		t.Error("the land is not on the battlefield tapped")
	}
	if g.LandsPlayedThisTurnFor(me.ID) != 0 {
		t.Error("a put used the turn's land drop")
	}
}

func TestC2B16OnTheTrailPutsALandOnTheSecondDrawOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	// Not the active player: the draw step drew nothing for them.
	me := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Push(g, me.ID, "On the Trail", "Enchantment", c2b16OnTheTrailOracle, 0, 0)
	land := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.AvB9Draw(t, g, me.ID, 1)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Fatal("asked after the first draw")
	}
	effectstest.AvB9Draw(t, g, me.ID, 1)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) == nil {
		t.Fatal("not asked after the second draw")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) || !effectstest.C2b16IsTapped(g, land) {
		t.Error("the land is not on the battlefield tapped")
	}
	effectstest.AvB9Draw(t, g, me.ID, 1)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil {
		t.Error("asked again after the third draw")
	}
}

func TestC2B16GreaterTanukiChannelsForABasicLandTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest(), effectstest.C2b15Spell())[1]
	id := effectstest.PushCatalogHandCard(me, "Greater Tanuki", "Enchantment Creature — Dog", c2b16TanukiOracle)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("channel: %v", err)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("the Tanuki was not discarded as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) || !effectstest.C2b16IsTapped(g, forest) {
		t.Error("the basic land is not on the battlefield tapped")
	}
}

func TestC2B16BeanstalkGiantIsAsBigAsYourLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c2b16Lands(g, me, 3)
	giant := effectstest.B12Push(g, me.ID, "Beanstalk Giant", "Creature — Giant", c2b16BeanstalkOracle, 0, 0)
	if p := effectstest.CtrlEffectivePower(t, g, giant); p != 3 {
		t.Errorf("power %d, want 3", p)
	}
	c2b16Lands(g, me, 2)
	if p := effectstest.CtrlEffectivePower(t, g, giant); p != 5 {
		t.Errorf("power %d after two more lands, want 5", p)
	}
	c2b16Lands(g, g.Seats[(g.Turn.ActiveSeat+1)%4], 4)
	if p := effectstest.CtrlEffectivePower(t, g, giant); p != 5 {
		t.Errorf("an opponent's lands counted: power %d", p)
	}
}

func TestC2B16FertileFootstepsPutsABasicLandUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest())[1]
	id := effectstest.CastCatalogSpell(t, g, "Fertile Footsteps", "Sorcery — Adventure", c2b16BeanstalkOracle+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) || effectstest.C2b16IsTapped(g, forest) {
		t.Error("the basic land is not on the battlefield untapped")
	}
	_ = id
}

func TestC2B16TopiaryStomperFetchesALandAndNeedsSevenLandsToFight(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	forest := effectstest.DftMPLibraryTop(me, effectstest.C2b15Bear(), effectstest.C2b15Forest())[1]
	c2b16Lands(g, me, 5)
	stomper := effectstest.DftABCast(t, g, me, game.Card{Name: "Topiary Stomper", TypeLine: "Creature — Plant Dinosaur",
		OracleID: c2b16StomperOracle, Power: 4, Toughness: 4})
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !g.Battlefield.Contains(forest) || !effectstest.C2b16IsTapped(g, forest) {
		t.Fatal("the basic land is not on the battlefield tapped")
	}
	restricted := func() bool {
		card := effectstest.LayeredCard(t, g, stomper)
		return game.Restricted(&card, game.CantAttack) && game.Restricted(&card, game.CantBlock)
	}
	if !restricted() {
		t.Error("six lands: the Stomper can attack or block")
	}
	c2b16Lands(g, me, 1)
	if restricted() {
		t.Error("seven lands: the Stomper is still restricted")
	}
	// An opponent's lands do not count.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	c2b16Lands(g2, g2.Seats[(g2.Turn.ActiveSeat+1)%4], 9)
	c2b16Lands(g2, me2, 6)
	s2 := effectstest.B12Push(g2, me2.ID, "Topiary Stomper", "Creature — Plant Dinosaur", c2b16StomperOracle, 4, 4)
	effectstest.AdvanceToMain(t, g2)
	g2.WithWriteLock(func() { g2.RecomputeLayersIfStaleLocked() })
	if card := effectstest.LayeredCard(t, g2, s2); !game.Restricted(&card, game.CantAttack) {
		t.Error("six own lands and nine opposing ones: the Stomper can attack")
	}
}

func TestC2B16RetreatToCoralhelmLandfallModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Retreat to Coralhelm", "Enchantment", c2b16CoralhelmOracle, 0, 0)
	bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0) // tap it
	if !effectstest.C2b16IsTapped(g, bear) {
		t.Error("the creature was not tapped")
	}
	// Second land: untap it, then scry.
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // untap it
	if effectstest.C2b16IsTapped(g, bear) {
		t.Error("the creature was not untapped")
	}
	// Third land: the may: do neither.
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{0, 1}, 0)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 2)
	if effectstest.C2b16IsTapped(g, bear) {
		t.Error("declining tapped the creature")
	}
}

func TestC2B16RetreatToCoralhelmScryMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Retreat to Coralhelm", "Enchantment", c2b16CoralhelmOracle, 0, 0)
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.ChooseModeNow(t, g, me.ID, []int{1}, 1) // no creature on the board: only the scry
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceScry, me.ID) == nil {
		t.Error("no scry prompt")
	}
}

func TestC2B16VergeRangersPlaysLandsOffTheTopOnlyWhenBehindOnLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.ToMain(t, g)
	effectstest.B12Push(g, me.ID, "Verge Rangers", "Creature — Human Scout Ranger", c2b16RangersOracle, 3, 3)
	if !effectstest.TopVisibleTo(g, me.ID, me.ID) || effectstest.TopVisibleTo(g, me.ID, opp.ID) {
		t.Error("the look should be private to the Rangers' controller")
	}
	land := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})
	if effectstest.GrantedPermissionOn(g, me.ID, land, game.ZoneLibrary).Granted() {
		t.Error("a land on top is open with no opponent ahead")
	}
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a land was played off the top while level on lands")
	}
	c2b16Lands(g, opp, 1)
	if err := g.CastSpell(me.ID, land, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("play a land off the top while an opponent has more lands: %v", err)
	}
	if !g.Battlefield.Contains(land) {
		t.Error("the land is not on the battlefield")
	}
	// Level again (the opponent's one land against the one just played): closed.
	next := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Next Forest", TypeLine: "Basic Land — Forest"})
	me.LandDropsPerTurn++
	if err := g.CastSpell(me.ID, next, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a second land was played off the top once the lands were level")
	}
}

func TestC2B16ShriekwoodDevourerUntapsUpToXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	devourer := effectstest.B12Push(g, me.ID, "Shriekwood Devourer", "Creature — Treefolk", c2b16DevourerOracle, 7, 5)
	lands := c2b16Lands(g, me, 8)
	for _, id := range lands {
		effectstest.FinS3Tap(g, id)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(devourer, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if ch == nil {
		t.Fatal("no untap pick")
	}
	if ch.ChooseMin != 0 || ch.ChooseMax != 7 {
		t.Fatalf("pick is %d..%d, want up to 7", ch.ChooseMin, ch.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, lands[:7]...)
	untapped := 0
	for _, id := range lands {
		if !effectstest.C2b16IsTapped(g, id) {
			untapped++
		}
	}
	if untapped != 7 {
		t.Errorf("%d lands untapped, want 7", untapped)
	}
}
