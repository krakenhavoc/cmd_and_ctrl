package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eventide activated-ability cards (S82 older sets, batch 523): every test
// plays the real catalog entry through the engine.

const (
	oldB523IndigoFaerie    = "10ae611f-e56d-47c5-8b6e-d38d95208bba"
	oldB523PatrolSignaler  = "1a8d7714-a566-4195-8605-550929584c61"
	oldB523SwirlingSpriggn = "3b0cea17-9baf-43c7-9f01-44f255e9d687"
	oldB523AltarGolem      = "ac143b8e-f993-4989-8833-e26dc9faffa4"
	oldB523MirrorSheen     = "d6494704-54b2-4c80-acb4-184c9815c5f3"
	oldB523ImpelledGiant   = "eb7c9442-7bfa-4bcf-a394-b5de187f224a"
)

func TestOldB523EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB523IndigoFaerie: "Indigo Faerie", oldB523PatrolSignaler: "Patrol Signaler",
		oldB523SwirlingSpriggn: "Swirling Spriggan", oldB523AltarGolem: "Altar Golem",
		oldB523MirrorSheen: "Mirror Sheen", oldB523ImpelledGiant: "Impelled Giant",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB523IndigoFaerieAddsBlueUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	faerie := effectstest.B27Push(g, me.ID, "Indigo Faerie", "Creature — Faerie Wizard", oldB523IndigoFaerie, "{1}{U}", 1, 1, "U")
	effectstest.AssertKeywords(t, g, faerie, "flying")
	ogre := effectstest.B27Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	effectstest.ApaMana(me, "U")
	effectstest.ApaActivate(t, g, me, faerie, 0, game.ActivateAbilityParams{Targets: effectstest.CardTargets(ogre)})
	if got := effectstest.TmtB9Colors(t, g, ogre); got != "UR" && got != "RU" {
		t.Errorf("colors = %q, want red and blue (in addition)", got)
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.TmtB9Colors(t, g, ogre); got != "R" {
		t.Errorf("after the turn colors = %q, want red again", got)
	}
}

func TestOldB523PatrolSignalerMakesAKithkinForTheUntapSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	signaler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Patrol Signaler", TypeLine: "Creature — Kithkin Soldier", OracleID: oldB523PatrolSignaler,
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	effectstest.ApaMana(me, "W", "W")
	if err := effectstest.InvB1Activate(g, me.ID, signaler, 0); err == nil {
		t.Fatal("an untapped Signaler paid {Q}")
	}
	me.ManaPool = game.ManaPool{}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == signaler {
			g.Battlefield.Cards[i].Tapped = true
		}
	}
	effectstest.ApaMana(me, "W", "W")
	effectstest.ApaActivate(t, g, me, signaler, 0, game.ActivateAbilityParams{})
	if effectstest.FindCardAnywhere(t, g, signaler).Tapped {
		t.Error("the Signaler is still tapped")
	}
	n, _ := effectstest.CountTokensNamed(g, me.ID, "Kithkin Soldier")
	if n != 1 {
		t.Errorf("%d Kithkin Soldier tokens, want 1", n)
	}
}

func TestOldB523SwirlingSprigganRecolorsYourCreatureToTheChosenColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	spriggan := effectstest.B27Push(g, me.ID, "Swirling Spriggan", "Creature — Goblin Shaman", oldB523SwirlingSpriggn, "{2}{G}{G}", 3, 3, "G")
	theirs := effectstest.B27Push(g, opp.ID, "Ogre", "Creature — Ogre", "", "{2}{R}", 3, 3, "R")
	effectstest.ApaMana(me, "G", "U")
	if err := effectstest.InvB1Activate(g, me.ID, spriggan, 0, effectstest.CardTarget(theirs)...); err == nil {
		t.Fatal("the Spriggan targeted an opponent's creature")
	}
	me.ManaPool = game.ManaPool{}
	effectstest.ApaMana(me, "G", "U")
	if err := effectstest.InvB1Activate(g, me.ID, spriggan, 0, effectstest.CardTarget(spriggan)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.PendingOfKind(g, game.PendingChoiceColor)
	if c == nil || !c.ColorMulti {
		t.Fatalf("prompt = %+v, want a multi colour pick", c)
	}
	effectstest.AnswerColor(t, g, me.ID, "UB")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TmtB9Colors(t, g, spriggan); got != "UB" && got != "BU" {
		t.Errorf("colors = %q, want blue and black only", got)
	}
	effectstest.RunCleanup(t, g)
	if got := effectstest.TmtB9Colors(t, g, spriggan); got != "G" {
		t.Errorf("after the turn colors = %q, want green again", got)
	}
}

func TestOldB523ImpelledGiantPumpsByTheTappedRedCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	giant := effectstest.B27Push(g, me.ID, "Impelled Giant", "Creature — Giant Warrior", oldB523ImpelledGiant, "{4}{R}{R}", 3, 3, "R")
	red := effectstest.B27Push(g, me.ID, "Red Ogre", "Creature — Ogre", "", "{3}{R}", 4, 4, "R")
	green := effectstest.B27Push(g, me.ID, "Green Bear", "Creature — Bear", "", "{1}{G}", 2, 2, "G")
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{green}}); err == nil {
		t.Error("a green creature paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{giant}}); err == nil {
		t.Error("the Giant tapped itself to pay")
	}
	effectstest.ApaActivate(t, g, me, giant, 0, game.ActivateAbilityParams{TapIDs: []uuid.UUID{red}})
	if p := effectstest.EffectivePower(t, g, giant); p != 7 {
		t.Errorf("Giant power %d, want 7 (3 + the tapped creature's 4)", p)
	}
	if !effectstest.FindCardAnywhere(t, g, red).Tapped {
		t.Error("the red creature was not tapped")
	}
}

func TestOldB523AltarGolemCountsEveryCreatureAndUntapsForFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Altar Golem", TypeLine: "Artifact Creature — Golem", OracleID: oldB523AltarGolem,
		Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	effectstest.AssertKeywords(t, g, golem, "trample")
	var mine []uuid.UUID
	for i := 0; i < 4; i++ {
		mine = append(mine, effectstest.B27Push(g, me.ID, "Mine", "Creature — Bear", "", "", 1, 1))
	}
	effectstest.B27Push(g, opp.ID, "Theirs", "Creature — Bear", "", "", 1, 1)
	effectstest.B27Push(g, opp.ID, "Theirs", "Creature — Bear", "", "", 1, 1)
	effectstest.B27Push(g, me.ID, "Relic", "Artifact", "", "{1}", 0, 0)
	if p, tough := effectstest.EffectivePower(t, g, golem), effectstest.EffectiveToughness(t, g, golem); p != 7 || tough != 7 {
		t.Errorf("Altar Golem is %d/%d, want 7/7 (seven creatures, itself included)", p, tough)
	}
	// Four creatures are not five.
	if err := g.ActivateCatalogAbility(me.ID, golem, 0, game.ActivateAbilityParams{Strict: true, TapIDs: mine}); err == nil {
		t.Error("four creatures paid a cost of five")
	}
	// Five untapped creatures: the four plus the Golem itself would need it untapped; use five others.
	fifth := effectstest.B27Push(g, me.ID, "Fifth", "Creature — Bear", "", "", 1, 1)
	effectstest.ApaActivate(t, g, me, golem, 0, game.ActivateAbilityParams{TapIDs: append(append([]uuid.UUID{}, mine...), fifth)})
	if effectstest.FindCardAnywhere(t, g, golem).Tapped {
		t.Error("the Golem is still tapped")
	}
	for _, id := range append(mine, fifth) {
		if !effectstest.FindCardAnywhere(t, g, id).Tapped {
			t.Error("a creature was not tapped to pay")
		}
	}
}

func TestOldB523AltarGolemDoesNotUntapInTheUntapStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	golem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Altar Golem", TypeLine: "Artifact Creature — Golem", OracleID: oldB523AltarGolem,
		Power: 0, Toughness: 0, Owner: me.ID, Controller: me.ID, Tapped: true,
	})
	effectstest.B27Push(g, me.ID, "Bear", "Creature — Bear", "", "", 1, 1)
	for i := 0; i < 200; i++ {
		if g.Turn.ActiveSeat != seat && g.Turn.Step == game.StepUpkeep {
			break
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	if !effectstest.FindCardAnywhere(t, g, golem).Tapped {
		t.Error("the Golem untapped during its controller's untap step")
	}
}

func TestOldB523MirrorSheenCopiesASpellThatTargetsYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	sheen := effectstest.PushCatalogPermanent(g, me.ID, "Mirror Sheen", "Enchantment", oldB523MirrorSheen, false)
	elsewhere := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}})
	effectstest.ApaMana(me, "U", "U", "U")
	if err := effectstest.InvB1Activate(g, me.ID, sheen, 0, effectstest.CardTarget(elsewhere)...); err == nil {
		t.Fatal("Mirror Sheen targeted a spell that does not target its controller")
	}
	// A spell that does target you.
	g2 := effectstest.NewCatalogGame(t)
	me, opp, third = g2.Seats[0], g2.Seats[1], g2.Seats[2]
	sheen = effectstest.PushCatalogPermanent(g2, me.ID, "Mirror Sheen", "Enchantment", oldB523MirrorSheen, false)
	bolt := effectstest.Batch01OpponentCasts(t, g2, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	effectstest.ApaMana(me, "U", "U", "U")
	if err := effectstest.InvB1Activate(g2, me.ID, sheen, 0, effectstest.CardTarget(bolt)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	life, thirdLife := me.Life, third.Life
	for i := 0; i < 8 && effectstest.LatestPickTarget(g2, me.ID) == nil; i++ {
		if err := g2.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	retarget := effectstest.LatestPickTarget(g2, me.ID)
	if retarget == nil {
		t.Fatal("no re-target prompt for the copy")
	}
	if err := g2.ResolvePickTarget(retarget.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: third.ID}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if third.Life != thirdLife-3 {
		t.Errorf("the copy's new target is at %d, want %d", third.Life, thirdLife-3)
	}
	if me.Life != life-3 {
		t.Errorf("the original bolt left me at %d, want %d", me.Life, life-3)
	}
}
