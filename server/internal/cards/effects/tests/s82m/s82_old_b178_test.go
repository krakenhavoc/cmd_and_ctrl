package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB178Regenesis = "14bc2a2e-a445-4785-9c88-b9ad169c39bc"
	oldB178Consec    = "20e7a93f-77ce-466b-8586-35d390689d0c"
	oldB178Claws     = "4538ec2c-867e-412b-b6e3-a57f8e29ba84"
	oldB178Gutter    = "76ae3779-c821-4cd3-8d0f-8c42d1206ce4"
	oldB178Stage     = "77b6a3db-c098-4b06-bc7f-136eef892d30"
	oldB178Imp       = "bba44fad-b3ca-4892-b7d1-2d03478062af"
	oldB178Angel     = "fcfcac6b-6d26-4ffe-ae9e-b0ee7fa49787"
)

func TestOldB178AreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB178Regenesis: "Regenesis", oldB178Consec: "Consecrate", oldB178Consec + "#1": "Consume",
		oldB178Claws: "Scrabbling Claws", oldB178Gutter: "Gutterbones", oldB178Stage: "Clear the Stage",
		oldB178Imp: "Carrion Imp", oldB178Angel: "Angel of Grace",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB178CastSplit casts a free split card from the active player's hand.
func oldB178CastSplit(t *testing.T, g *game.Game, oracle string, face int, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	card := game.Card{
		InstanceID: uuid.New(), OracleID: oracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Left", TypeLine: "Instant", ManaCost: "{0}"},
			{Name: "Right", TypeLine: "Sorcery", ManaCost: "{0}"},
		},
	}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.AdvanceToMain(t, g)
	return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Face: face, Targets: targets, Strict: true})
}

func oldB178GraveCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestOldB178RegenesisReturnsUpToTwoPermanentCardsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	creature := oldB178GraveCard(me, "Bear", "Creature — Bear", "", "")
	land := oldB178GraveCard(me, "Forest", "Basic Land — Forest", "", "")
	spell := oldB178GraveCard(me, "Opt", "Instant", "", "")
	third := oldB178GraveCard(me, "Ox", "Creature — Ox", "", "")
	if err := effectstest.CastCatalogSpellErr(t, g, "Regenesis", "Instant", oldB178Regenesis, effectstest.CardRefs(spell)); err == nil {
		t.Fatal("an instant card was accepted as a target")
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Regenesis", "Instant", oldB178Regenesis, effectstest.CardRefs(creature, land, third)); err == nil {
		t.Fatal("three targets were accepted")
	}
	effectstest.CastCatalogSpell(t, g, "Regenesis", "Instant", oldB178Regenesis, effectstest.CardRefs(creature, land))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, creature) || !effectstest.InHandOf(me, land) {
		t.Error("the two permanent cards did not return to the hand")
	}
	if effectstest.InHandOf(me, third) {
		t.Error("an untargeted card returned")
	}
}

func TestOldB178ConsecrateExilesTheCardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := oldB178GraveCard(opp, "Their Card", "Creature — Bear", "", "")
	effectstest.SeedLibrary(me, "x", "y")
	hand := me.Hand.Size()
	if err := oldB178CastSplit(t, g, oldB178Consec, 0, effectstest.CardRefs(theirs)); err != nil {
		t.Fatalf("Consecrate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, theirs) != game.ZoneExile {
		t.Errorf("the card is in %v, want exile", effectstest.ZoneOfCard(g, theirs))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestOldB178ConsumeSacrificesAGreatestPowerCreatureAndGainsItsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	alsoBig := effectstest.B12Creature(g, opp.ID, "Also Big", "Creature — Ogre", 5, 2)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 2)
	life := me.Life
	if err := oldB178CastSplit(t, g, oldB178Consec, 1, effectstest.Pr6Player(opp.ID)); err != nil {
		t.Fatalf("Consume: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("no sacrifice prompt for the target player")
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{small}); err == nil {
		t.Error("a creature without the greatest power was accepted")
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, alsoBig)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(alsoBig) || !g.Battlefield.Contains(big) || !g.Battlefield.Contains(small) {
		t.Error("the wrong creature was sacrificed")
	}
	if me.Life != life+5 {
		t.Errorf("life %d, want %d", me.Life, life+5)
	}
}

func TestOldB178ScrabblingClawsBothAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	claws := effectstest.PushCatalogPermanent(g, me.ID, "Scrabbling Claws", "Artifact", oldB178Claws, false)
	mine := oldB178GraveCard(me, "Mine", "Sorcery", "", "")
	other := oldB178GraveCard(me, "Other", "Sorcery", "", "")
	theirs := oldB178GraveCard(opp, "Theirs", "Sorcery", "", "")
	effectstest.SeedLibrary(me, "x", "y")
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, claws, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr6Player(me.ID)}); err != nil {
		t.Fatalf("{T}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, mine)
	if effectstest.ZoneOfCard(g, mine) != game.ZoneExile || effectstest.ZoneOfCard(g, other) != game.ZoneGraveyard {
		t.Error("the target player did not exile exactly the card they chose")
	}
	hand := me.Hand.Size()
	effectstest.FloatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, claws, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(theirs)}); err != nil {
		t.Fatalf("{1}, sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, theirs) != game.ZoneExile {
		t.Error("their card was not exiled")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	if g.Battlefield.Contains(claws) {
		t.Error("the Claws were not sacrificed")
	}
}

func TestOldB178GutterbonesEntersTappedAndReturnsOnlyOnYourTurnAfterAnOpponentLostLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bones := effectstest.CastHoldCreature(t, g, "Gutterbones", "Creature — Skeleton Warrior", oldB178Gutter, 2, 1)
	if c, ok := effectstest.BattlefieldCard(g, bones); !ok || !c.Tapped {
		t.Fatal("Gutterbones did not enter tapped")
	}
	card := oldB178GraveCard(me, "Gutterbones", "Creature — Skeleton Warrior", "{B}", oldB178Gutter)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated before any opponent lost life")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated after only its controller lost life")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{})
	if !effectstest.InHandOf(me, card) {
		t.Error("Gutterbones did not return to the hand")
	}
}

func TestOldB178ClearTheStageShrinksAndReturnsOnlyWithAPowerFourCreature(t *testing.T) {
	for _, withBig := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 3)
		if withBig {
			effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 4, 4)
		}
		back := oldB178GraveCard(me, "Back", "Creature — Bear", "", "")
		effectstest.CastCatalogSpell(t, g, "Clear the Stage", "Instant", oldB178Stage, effectstest.CardRefs(victim, back))
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) {
			t.Error("the -3/-3 did not kill the victim")
		}
		if got := effectstest.InHandOf(me, back); got != withBig {
			t.Errorf("withBig=%v: the creature card returned = %v", withBig, got)
		}
	}
}

func TestOldB178CarrionImpMayExileACreatureCardAndGainTwo(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		target := oldB178GraveCard(opp, "Their Bear", "Creature — Bear", "", "")
		life := me.Life
		effectstest.CastCatalogSpell(t, g, "Carrion Imp", "Creature — Imp", oldB178Imp, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			effectstest.PickCard(t, g, me.ID, target)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, accept)
		effectstest.PassPriorityAroundTable(t, g)
		exiled := effectstest.ZoneOfCard(g, target) == game.ZoneExile
		if exiled != accept {
			t.Errorf("accept=%v: exiled=%v", accept, exiled)
		}
		want := life
		if accept {
			want += 2
		}
		if me.Life != want {
			t.Errorf("accept=%v: life %d, want %d", accept, me.Life, want)
		}
	}
}

func TestOldB178AngelOfGraceFloorsLifeAtOneThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Pinger", 1, 1)
	effectstest.CastCatalogSpell(t, g, "Angel of Grace", "Creature — Angel", oldB178Angel, nil)
	effectstest.PassPriorityAroundTable(t, g)
	me.Life = 5
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 9) })
	if me.Life != 1 {
		t.Errorf("life %d after lethal damage, want 1", me.Life)
	}
}

func TestOldB178AngelOfGraceSetsLifeToTenFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := oldB178GraveCard(me, "Angel of Grace", "Creature — Angel", "{3}{W}{W}", oldB178Angel)
	me.Life = 3
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "CCCCWW")
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{})
	if me.Life != 10 {
		t.Errorf("life %d, want 10", me.Life)
	}
	if effectstest.ZoneOfCard(g, card) != game.ZoneExile {
		t.Error("the card was not exiled as a cost")
	}
	// From above ten the ability lowers it too: "becomes", not "gains".
	card2 := oldB178GraveCard(me, "Angel of Grace", "Creature — Angel", "{3}{W}{W}", oldB178Angel)
	me.Life = 30
	effectstest.FloatForTest(g, me, "CCCCWW")
	effectstest.ApaActivate(t, g, me, card2, 0, game.ActivateAbilityParams{})
	if me.Life != 10 {
		t.Errorf("life %d from 30, want 10", me.Life)
	}
}
