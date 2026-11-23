package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSNC5SNCBatch5CardsAreRegisteredAndWhole(t *testing.T) {
	for _, id := range []string{effectstest.Snc5MakeDisappear, effectstest.Snc5ALittleChat, effectstest.Snc5LightEmUp, effectstest.Snc5GrislySigil, effectstest.Snc5DigUpTheBody,
		effectstest.Snc5RobTheArchives, effectstest.Snc5RooftopNuisance, effectstest.Snc5JoinTheMaestros, effectstest.Snc5IllicitShipment, effectstest.Snc5CutYourLosses,
		effectstest.Snc5CutOfTheProfits, effectstest.Snc5AudaciousSwap} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Fatalf("%s: not registered", id)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v", spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.OptionalCosts) != 1 || spec.OptionalCosts[0].Key != game.CasualtyKey {
			t.Errorf("%s: no casualty cost", spec.Name)
		}
	}
}

func TestSNC5LightEmUpCasualtyCopiesAndRetargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", effectstest.Snc5LightEmUp)
	opp := g.Seats[1]
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Wall", 0, 4)
	fodder := effectstest.Snc5Sac(g, me.ID, 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("casualty creature not sacrificed")
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{Kind: game.TargetCard, ID: victim}, false, false)
	if g.Battlefield.Contains(victim) {
		t.Error("2 damage from the spell and 2 from its copy should kill a 0/4")
	}
}

func TestSNC5LightEmUpWithoutCasualtyDealsTwoOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", effectstest.Snc5LightEmUp)
	victim := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if !g.Battlefield.Contains(victim) {
		t.Error("a 0/4 died to one 2-damage spell")
	}
}

func TestSNC5CasualtyTwoRefusesAPowerOneCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Light 'Em Up", "Sorcery", "{1}{R}", effectstest.Snc5LightEmUp)
	victim := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	small := effectstest.Snc5Sac(g, me.ID, 1)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
		OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{small}, Strict: true,
	})
	if err == nil {
		t.Error("a power-1 creature paid casualty 2")
	}
}

func TestSNC5GrislySigil(t *testing.T) {
	for _, hurt := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, bolt := effectstest.SqBoltInHand(t, g)
		victim := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 9)
		if hurt {
			if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
				t.Fatalf("bolt: %v", err)
			}
			effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
		}
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Grisly Sigil", TypeLine: "Sorcery", ManaCost: "{B}",
				OracleID: effectstest.Snc5GrislySigil, Owner: me.ID, Controller: me.ID})
		})
		before := me.Life
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
			t.Fatalf("sigil: %v", err)
		}
		effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
		want := 1
		if hurt {
			want = 3
		}
		if me.Life-before != want {
			t.Errorf("hurt=%v: gained %d, want %d", hurt, me.Life-before, want)
		}
	}
}

func TestSNC5GrislySigilIgnoresCombatAndRecycledDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	victim := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 9)
	g.WithWriteLock(func() {
		if g.PermanentWasDealtNoncombatDamageThisTurn(victim) {
			t.Error("undamaged creature reads as dealt noncombat damage")
		}
	})
}

func TestSNC5CutYourLossesMillsHalf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Cut Your Losses", "Sorcery", "{4}{U}{U}", effectstest.Snc5CutYourLosses)
	opp := g.Seats[1]
	before := len(opp.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if got := before - len(opp.Library.Cards); got != before/2 {
		t.Errorf("milled %d, want %d", got, before/2)
	}
}

func TestSNC5CutOfTheProfitsDrawsAndLosesX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Cut of the Profits", "Sorcery", "{X}{B}{B}", effectstest.Snc5CutOfTheProfits)
	hand, life := len(me.Hand.Cards)-1, me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+3 || me.Life != life-3 {
		t.Errorf("hand %d (want %d), life %d (want %d)", len(me.Hand.Cards), hand+3, me.Life, life-3)
	}
}

func TestSNC5CutOfTheProfitsCasualtyCopyHasTheSameX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Cut of the Profits", "Sorcery", "{X}{B}{B}", effectstest.Snc5CutOfTheProfits)
	fodder := effectstest.Snc5Sac(g, me.ID, 3)
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, OptionalCosts: []int{0}, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if me.Life != life-4 {
		t.Errorf("life %d, want %d (X=2 twice)", me.Life, life-4)
	}
}

func TestSNC5JoinTheMaestrosMakesAnOgreWarrior(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Join the Maestros", "Sorcery", "{4}{B}", effectstest.Snc5JoinTheMaestros)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Name == "Ogre Warrior" && c.Controller == me.ID && c.Power == 4 && c.Toughness == 3 {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Ogre Warrior tokens, want 1", n)
	}
}

func TestSNC5IllicitShipmentTutors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Illicit Shipment", "Sorcery", "{3}{B}{B}", effectstest.Snc5IllicitShipment)
	hand := len(me.Hand.Cards) - 1
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestSNC5RooftopNuisanceTapsDrawsAndFreezes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Rooftop Nuisance", "Sorcery", "{2}{U}", effectstest.Snc5RooftopNuisance)
	victim := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 4)
	hand := len(me.Hand.Cards) - 1
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: victim}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	c, _ := g.LookupCardForEffect(victim)
	if !c.Tapped {
		t.Error("creature not tapped")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d, want %d", len(me.Hand.Cards), hand+1)
	}
}

func TestSNC5RobTheArchivesExilesTwoYouMayPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Rob the Archives", "Sorcery", "{1}{R}", effectstest.Snc5RobTheArchives)
	lib := len(me.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Library.Cards) != lib-2 {
		t.Errorf("library %d, want %d", len(me.Library.Cards), lib-2)
	}
}

func TestSNC5DigUpTheBodyMillsTwoAndReturnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Dig Up the Body", "Instant", "{2}{B}", effectstest.Snc5DigUpTheBody)
	bear := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bear, Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
			Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	inHand := false
	for _, c := range me.Hand.Cards {
		if c.InstanceID == bear {
			inHand = true
		}
	}
	if !inHand {
		t.Error("the creature card did not return to hand")
	}
}

func TestSNC5ALittleChatTakesOneOfTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "A Little Chat", "Instant", "{1}{U}", effectstest.Snc5ALittleChat)
	hand, lib := len(me.Hand.Cards)-1, len(me.Library.Cards)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if len(me.Hand.Cards) != hand+1 || len(me.Library.Cards) != lib-1 {
		t.Errorf("hand %d (want %d), library %d (want %d)", len(me.Hand.Cards), hand+1, len(me.Library.Cards), lib-1)
	}
}

func TestSNC5MakeDisappearCountersUnlessTwoPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	me, bolt := effectstest.SqBoltInHand(t, g)
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("bolt: %v", err)
	}
	id := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: id, Name: "Make Disappear", TypeLine: "Instant", ManaCost: "{1}{U}",
			OracleID: effectstest.Snc5MakeDisappear, Owner: opp.ID, Controller: opp.ID})
	})
	before := opp.Life
	if err := g.CastSpell(opp.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}}); err != nil {
		t.Fatalf("Make Disappear: %v", err)
	}
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	if opp.Life != before {
		t.Errorf("the Bolt resolved (life %d -> %d); it should have been countered", before, opp.Life)
	}
}

// snc5SwapSetup targets an opponent's permanent `target`, with the
// owner's library holding exactly `top`. The target is shuffled into that
// library, so which of the two is exiled is random: the tests below make
// both cards the same kind, so the outcome does not depend on the shuffle.
func snc5SwapSetup(t *testing.T, target, top game.Card) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, id := effectstest.Snc5Hand(t, g, "Audacious Swap", "Instant", "{3}{R}", effectstest.Snc5AudaciousSwap)
	opp := g.Seats[1]
	tok := uuid.New()
	g.WithWriteLock(func() {
		target.InstanceID = tok
		target.Owner, target.Controller = opp.ID, opp.ID
		g.Battlefield.PushTop(target)
		top.InstanceID = uuid.New()
		top.Owner, top.Controller = opp.ID, opp.ID
		opp.Library.Cards = nil
		opp.Library.PushTop(top)
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: tok}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	return g, me, opp
}

func TestSNC5AudaciousSwapLandEntersUnderTheOwner(t *testing.T) {
	g, _, opp := snc5SwapSetup(t,
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.Snc5Drive(t, g, game.TargetRef{}, false, false)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsLand() && c.Controller == opp.ID {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d lands under the owner, want 1 (the other was shuffled into the library)", n)
	}
	if len(opp.Library.Cards) != 1 {
		t.Errorf("library %d, want 1", len(opp.Library.Cards))
	}
}

func TestSNC5AudaciousSwapNonlandIsOfferedToTheOwnerNotTheCaster(t *testing.T) {
	g, me, opp := snc5SwapSetup(t,
		game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4},
		game.Card{Name: "Divination", TypeLine: "Sorcery", ManaCost: "{2}{U}"})
	for i := 0; i < 8 && effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, opp.ID) == nil; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	pc := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, opp.ID)
	if pc == nil {
		t.Fatal("the owner was not offered the free cast")
	}
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceResolutionCast, me.ID) != nil {
		t.Error("the caster was offered the free cast")
	}
}
