package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB14ArmixOracle       = "a1464164-fbda-43b5-988e-6e1da9bd9eef"
	rprB14ReshapersOracle   = "525c0e33-f073-4bed-be8a-14e436a0bcf1"
	rprB14CourierOracle     = "a07ceaca-9964-4c24-937e-090b6643b327"
	rprB14DigsiteOracle     = "e57025ed-3de6-4b2f-92bd-5481fd33b7f1"
	rprB14DrossforgeOracle  = "44b83535-fdbf-4307-bf53-ca20470a768d"
	rprB14ChampionOracle    = "7cd7102d-6413-4f46-9434-9130aefa2f88"
	rprB14CapsuleOracle     = "c9cd266c-7ecf-4beb-b9da-69b88f33abd3"
	rprB14AttendantOracle   = "ab7c9e72-8f30-47f7-918c-cd3c0593d704"
	rprB14GethOracle        = "d2310103-ec84-49df-bb12-0f0f175ee801"
	rprB14ArchangelOracle   = "0f9269eb-6a6c-4b32-8b3d-73724c12ca6b"
	rprB14MirrorworksOracle = "fd9c4ebe-cc9c-427c-8105-adb1703c191b"
)

func rprB14Table(t *testing.T) (*game.Game, *game.Player, *game.Player) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func rprB14Artifact(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushCatalogPermanent(g, owner, name, "Artifact", "", false)
}

func rprB14Pay(t *testing.T, g *game.Game, p *game.Player, mana string) {
	t.Helper()
	if err := g.AddManaForEffect(p.ID, uuid.Nil, mana); err != nil {
		t.Fatal(err)
	}
}

func TestRprB14DrossforgeBridgeEntersTappedIndestructibleAndMakesBOrR(t *testing.T) {
	g, me, _ := rprB14Table(t)
	id := effectstest.PlayLandFromHand(t, g, "Drossforge Bridge", rprB14DrossforgeOracle)
	if !effectstest.IsTapped(g, id) {
		t.Error("Drossforge Bridge entered untapped")
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "indestructible") {
		t.Error("Drossforge Bridge is not indestructible")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"R"}}); err != nil {
		t.Fatalf("tap for R: %v", err)
	}
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "R" {
		t.Fatalf("pool %v, want [R]", got)
	}
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err == nil {
		t.Fatal("Drossforge Bridge made G")
	}
}

func TestRprB14ExecutionersCapsuleDestroysANonblackCreature(t *testing.T) {
	g, me, opp := rprB14Table(t)
	capsule := effectstest.PushCatalogPermanent(g, me.ID, "Executioner's Capsule", "Artifact", rprB14CapsuleOracle, false)
	black := effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Black Bear", TypeLine: "Creature — Bear", Colors: []string{"B"}, Power: 2, Toughness: 2})
	red := effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Power: 2, Toughness: 2})

	rprB14Pay(t, g, me, "{C}{B}")
	err := g.ActivateCatalogAbility(me.ID, capsule, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: black}}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a black creature was targetable: %v", err)
	}
	if !g.Battlefield.Contains(capsule) {
		t.Fatal("an illegal target paid the cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, capsule, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: red}}}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(capsule) {
		t.Fatal("the Capsule was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(red) || !opp.Graveyard.Contains(red) {
		t.Fatal("the nonblack creature was not destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Fatal("the black creature was destroyed")
	}
}

func TestRprB14ExecutionersCapsuleNeedsItsMana(t *testing.T) {
	g, me, opp := rprB14Table(t)
	capsule := effectstest.PushCatalogPermanent(g, me.ID, "Executioner's Capsule", "Artifact", rprB14CapsuleOracle, false)
	red := effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"}, Power: 2, Toughness: 2})
	rprB14Pay(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, capsule, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: red}}}); err == nil {
		t.Fatal("activated for {B} alone")
	}
}

func TestRprB14EtchedChampionHasProtectionWithMetalcraft(t *testing.T) {
	g, me, _ := rprB14Table(t)
	champ := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Etched Champion", TypeLine: "Artifact Creature — Soldier",
		OracleID: rprB14ChampionOracle, Power: 2, Toughness: 2})
	one := rprB14Artifact(g, me.ID, "Rock 1")
	for _, c := range game.AllColors {
		if effectstest.HasEffectiveKeyword(t, g, champ, game.ProtectionFromColor(c)) {
			t.Fatalf("protection from %s with two artifacts", c)
		}
	}
	rprB14Artifact(g, me.ID, "Rock 2")
	for _, c := range game.AllColors {
		if !effectstest.HasEffectiveKeyword(t, g, champ, game.ProtectionFromColor(c)) {
			t.Fatalf("no protection from %s with three artifacts", c)
		}
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(one) })
	if effectstest.HasEffectiveKeyword(t, g, champ, game.ProtectionFromColor("R")) {
		t.Fatal("kept protection after dropping below three artifacts")
	}
}

func TestRprB14EtchedChampionCountsOnlyYourArtifacts(t *testing.T) {
	g, me, opp := rprB14Table(t)
	champ := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Etched Champion", TypeLine: "Artifact Creature — Soldier",
		OracleID: rprB14ChampionOracle, Power: 2, Toughness: 2})
	rprB14Artifact(g, opp.ID, "Their Rock 1")
	rprB14Artifact(g, opp.ID, "Their Rock 2")
	if effectstest.HasEffectiveKeyword(t, g, champ, game.ProtectionFromColor("R")) {
		t.Fatal("opponents' artifacts counted for metalcraft")
	}
}

func TestRprB14IndomitableArchangelGivesArtifactsShroudWithMetalcraft(t *testing.T) {
	g, me, opp := rprB14Table(t)
	angel := effectstest.PushCatalogPermanent(g, me.ID, "Indomitable Archangel", "Creature — Angel", rprB14ArchangelOracle, false)
	a := rprB14Artifact(g, me.ID, "Rock 1")
	b := rprB14Artifact(g, me.ID, "Rock 2")
	theirs := rprB14Artifact(g, opp.ID, "Their Rock")
	if effectstest.HasEffectiveKeyword(t, g, a, "shroud") {
		t.Fatal("shroud with two artifacts")
	}
	c := rprB14Artifact(g, me.ID, "Rock 3")
	for _, id := range []uuid.UUID{a, b, c} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "shroud") {
			t.Fatal("an artifact lacks shroud with three artifacts")
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, angel, "shroud") {
		t.Error("the Archangel (not an artifact) has shroud")
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "shroud") {
		t.Error("an opponent's artifact has shroud")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(c) })
	if effectstest.HasEffectiveKeyword(t, g, a, "shroud") {
		t.Fatal("shroud stayed after dropping below three artifacts")
	}
}

func TestRprB14FiligreeAttendantPowerIsYourArtifacts(t *testing.T) {
	g, me, opp := rprB14Table(t)
	att := effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Filigree Attendant", TypeLine: "Artifact Creature — Homunculus",
		OracleID: rprB14AttendantOracle, Power: 0, Toughness: 3})
	if got := effectstest.EffectivePower(t, g, att); got != 1 {
		t.Fatalf("power %d, want 1 (itself)", got)
	}
	rprB14Artifact(g, me.ID, "Rock 1")
	rprB14Artifact(g, me.ID, "Rock 2")
	rprB14Artifact(g, opp.ID, "Their Rock")
	if got := effectstest.EffectivePower(t, g, att); got != 3 {
		t.Fatalf("power %d, want 3", got)
	}
	if got := effectstest.EffectiveToughness(t, g, att); got != 3 {
		t.Fatalf("toughness %d, want 3", got)
	}
}

func rprB14Graveyard(p *game.Player, name, typeLine, cost string) uuid.UUID {
	return effectstest.PushGraveyardPermanent(p, name, typeLine, cost)
}

func TestRprB14GethReanimatesAnOpponentsCardTappedAndMillsThem(t *testing.T) {
	g, me, opp := rprB14Table(t)
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Lord of the Vault", "Legendary Creature — Phyrexian Zombie", rprB14GethOracle, false)
	target := rprB14Graveyard(opp, "Their Golem", "Artifact Creature — Golem", "{2}{R}")
	wrongMV := rprB14Graveyard(opp, "Pricey", "Artifact", "{5}")
	mine := rprB14Graveyard(me, "My Golem", "Artifact Creature — Golem", "{2}{R}")
	land := rprB14Graveyard(opp, "A Land", "Land", "")
	for i := 0; i < 5; i++ {
		effectstest.PutOnTopForTest(opp, "Filler", "Sorcery")
	}
	libBefore := opp.Library.Size()

	rprB14Pay(t, g, me, "{C}{C}{C}{B}")
	for _, bad := range []uuid.UUID{wrongMV, mine, land} {
		err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Strict: true, XValue: 3,
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bad}}})
		if !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("illegal target accepted: %v", err)
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Strict: true, XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, target)
	if !ok {
		t.Fatal("the card did not reach the battlefield")
	}
	if c.Controller != me.ID || c.Owner != opp.ID || !c.Tapped {
		t.Fatalf("controller %v owner %v tapped %v; want mine, theirs, tapped", c.Controller, c.Owner, c.Tapped)
	}
	if got := libBefore - opp.Library.Size(); got != 3 {
		t.Fatalf("opponent milled %d, want 3", got)
	}
}

func TestRprB14GethTakesACreatureCardToo(t *testing.T) {
	g, me, opp := rprB14Table(t)
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Lord of the Vault", "Legendary Creature — Phyrexian Zombie", rprB14GethOracle, false)
	target := rprB14Graveyard(opp, "Their Bear", "Creature — Bear", "")
	rprB14Pay(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Strict: true, XValue: 0,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, ok := effectstest.BattlefieldCard(g, target); !ok || c.Controller != me.ID {
		t.Fatal("X=0 did not take the mana value 0 creature card")
	}
}

func TestRprB14GethDoesNothingWhenTheCardLeavesInResponse(t *testing.T) {
	g, me, opp := rprB14Table(t)
	geth := effectstest.PushCatalogPermanent(g, me.ID, "Geth, Lord of the Vault", "Legendary Creature — Phyrexian Zombie", rprB14GethOracle, false)
	target := rprB14Graveyard(opp, "Their Golem", "Artifact Creature — Golem", "{2}{R}")
	for i := 0; i < 5; i++ {
		effectstest.PutOnTopForTest(opp, "Filler", "Sorcery")
	}
	libBefore := opp.Library.Size()
	rprB14Pay(t, g, me, "{C}{C}{C}{B}")
	if err := g.ActivateCatalogAbility(me.ID, geth, 0, game.ActivateAbilityParams{Strict: true, XValue: 3,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _, _ = game.MoveCard(opp.Graveyard, g.Exile, target) })
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) || opp.Library.Size() != libBefore {
		t.Fatal("a fizzled Geth still reanimated or milled")
	}
}

func TestRprB14ArmixDiscardsToShrinkADefendersCreature(t *testing.T) {
	g, me, opp := rprB14Table(t)
	armix := effectstest.B12Push(g, me.ID, "Armix, Filigree Thrasher", "Legendary Artifact Creature — Golem", rprB14ArmixOracle, 3, 2)
	rprB14Artifact(g, me.ID, "Rock")
	rprB14Graveyard(me, "Old Rock", "Artifact", "")
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 4)
	spare := effectstest.B12Creature(g, opp.ID, "Their Other Giant", "Creature — Giant", 5, 9)
	own := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	discard := effectstest.PushHandCardOfType(me, "Fodder Artifact", "Artifact")

	effectstest.DeclareAttack(t, g, opp.ID, armix)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LciDL2PayUnless(t, g, me.ID)
	if err := g.ResolvePayUnlessWithCards(c.ID, me.ID, true, []uuid.UUID{discard}); err != nil {
		t.Fatalf("discard: %v", err)
	}
	effectstest.B04WaitForPick(t, g, me.ID)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if effectstest.HasID(pick.PickTargetCards, own) || !effectstest.HasID(pick.PickTargetCards, victim) || !effectstest.HasID(pick.PickTargetCards, spare) {
		t.Errorf("offered %v; want the defending player's creatures only", pick.PickTargetCards)
	}
	effectstest.PickTriggerTarget(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)

	if !me.Graveyard.Contains(discard) {
		t.Error("the card was not discarded")
	}
	// X: Armix + Rock on the battlefield, Old Rock + the discarded
	// artifact in the graveyard = 4.
	if g.Battlefield.Contains(victim) {
		t.Fatalf("a 5/4 survived -X/-X with X = 4 (toughness %d)", effectstest.EffectiveToughness(t, g, victim))
	}
	if got := effectstest.EffectiveToughness(t, g, spare); got != 9 {
		t.Errorf("the untargeted creature changed: toughness %d", got)
	}
}

func TestRprB14ArmixDecliningDiscardsNothingAndAsksNoTarget(t *testing.T) {
	g, me, opp := rprB14Table(t)
	armix := effectstest.B12Push(g, me.ID, "Armix, Filigree Thrasher", "Legendary Artifact Creature — Golem", rprB14ArmixOracle, 3, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 4)
	keep := effectstest.PushHandCardOfType(me, "Keep Me", "Artifact")
	effectstest.DeclareAttack(t, g, opp.ID, armix)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(keep) || !g.Battlefield.Contains(victim) {
		t.Fatal("declining discarded a card or shrank a creature")
	}
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for after declining")
	}
}

func TestRprB14MirrorworksCopiesANontokenArtifactForTwo(t *testing.T) {
	g, me, _ := rprB14Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Mirrorworks", "Artifact", rprB14MirrorworksOracle, false)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Odd Gizmo", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an artifact entering did not offer the {2} payment")
	}
	rprB14Pay(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Odd Gizmo" && c.Controller == me.ID {
			n++
			if c.IsToken() && !c.IsArtifact() {
				t.Error("the copy is not an artifact")
			}
		}
	}
	if n != 2 {
		t.Fatalf("%d Odd Gizmos, want the original and one token", n)
	}
	// The token is a token: its own entry offers nothing.
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("the token copy triggered Mirrorworks again")
	}
}

func TestRprB14MirrorworksDeclineAndIgnoredEntries(t *testing.T) {
	g, me, opp := rprB14Table(t)
	works := effectstest.PushCatalogPermanent(g, me.ID, "Mirrorworks", "Artifact", rprB14MirrorworksOracle, false)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Odd Gizmo", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "Odd Gizmo", me.ID) != 1 {
		t.Fatal("declining still made a copy")
	}
	effectstest.EnterCard(t, g, opp.ID, game.Card{Name: "Their Gizmo", TypeLine: "Artifact"})
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("an opponent's artifact or a non-artifact offered the payment")
	}
	_ = works
}

func TestRprB14MirrorworksNeedsThePaymentToBeAffordable(t *testing.T) {
	g, me, _ := rprB14Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Mirrorworks", "Artifact", rprB14MirrorworksOracle, false)
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Odd Gizmo", TypeLine: "Artifact"})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, true)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CountOnBattlefield(g, "Odd Gizmo", me.ID) != 1 {
		t.Fatal("a copy was made with no mana to pay")
	}
}

func TestRprB14DigsiteEngineerMakesAConstructForTwo(t *testing.T) {
	g, me, _ := rprB14Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Digsite Engineer", "Creature — Dwarf Artificer", rprB14DigsiteOracle, false)
	effectstest.CastCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("casting an artifact spell did not offer the {2} payment")
	}
	rprB14Pay(t, g, me, "{C}{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTokensControlled(g, me.ID, "Construct"); got != 1 {
		t.Fatalf("%d Constructs, want 1", got)
	}
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.HasSubtype("Construct") {
			id = c.InstanceID
		}
	}
	// Test Rock and the Construct itself.
	if got := effectstest.EffectivePower(t, g, id); got != 2 {
		t.Fatalf("Construct power %d, want 2 (artifacts you control)", got)
	}
}

func TestRprB14DigsiteEngineerIgnoresNonartifactSpellsAndDeclines(t *testing.T) {
	g, me, _ := rprB14Table(t)
	effectstest.PushCatalogPermanent(g, me.ID, "Digsite Engineer", "Creature — Dwarf Artificer", rprB14DigsiteOracle, false)
	effectstest.CastCatalogSpell(t, g, "Test Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("a sorcery offered the payment")
	}
	effectstest.CastCatalogSpell(t, g, "Test Rock", "Artifact", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountTokensControlled(g, me.ID, "Construct"); got != 0 {
		t.Fatalf("declining still made %d Constructs", got)
	}
}

func TestRprB14AudaciousReshapersFindsAnArtifactAndHurtsYou(t *testing.T) {
	g, me, _ := rprB14Table(t)
	reshapers := effectstest.PushCatalogPermanent(g, me.ID, "Audacious Reshapers", "Creature — Human Artificer", rprB14ReshapersOracle, false)
	fodder := rprB14Artifact(g, me.ID, "Fodder")
	for i := 0; i < 4; i++ {
		effectstest.PutOnTopForTest(me, "Deep Filler", "Sorcery")
	}
	prize := effectstest.PutOnTopForTest(me, "Prize Artifact", "Artifact — Equipment")
	// Top to bottom: Land, Bear, Prize.
	bear := effectstest.PutOnTopForTest(me, "Revealed Bear", "Creature — Bear")
	land := effectstest.PutOnTopForTest(me, "Revealed Land", "Land")
	_ = bear
	_ = land
	libBefore := me.Library.Size()
	lifeBefore := me.Life

	if err := g.ActivateCatalogAbility(me.ID, reshapers, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Fatal("the artifact was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prize) {
		t.Fatal("the artifact card did not reach the battlefield")
	}
	if got := lifeBefore - me.Life; got != 3 {
		t.Fatalf("took %d damage, want 3 (cards revealed)", got)
	}
	// Three left the library, the two non-artifacts went back to the bottom.
	if me.Library.Size() != libBefore-1 {
		t.Fatalf("library %d, want %d", me.Library.Size(), libBefore-1)
	}
	if !me.Library.Contains(bear) || !me.Library.Contains(land) {
		t.Fatal("a revealed non-artifact was not put on the bottom")
	}
}

func TestRprB14AudaciousReshapersWithNoArtifactRevealsTheWholeLibrary(t *testing.T) {
	g, me, _ := rprB14Table(t)
	reshapers := effectstest.PushCatalogPermanent(g, me.ID, "Audacious Reshapers", "Creature — Human Artificer", rprB14ReshapersOracle, false)
	fodder := rprB14Artifact(g, me.ID, "Fodder")
	g.WithWriteLock(func() {
		for me.Library.Size() > 0 {
			c, _ := me.Library.Top()
			_, _ = game.MoveCard(me.Library, g.Exile, c.InstanceID)
		}
	})
	for i := 0; i < 3; i++ {
		effectstest.PutOnTopForTest(me, "Filler", "Sorcery")
	}
	lifeBefore := me.Life
	if err := g.ActivateCatalogAbility(me.ID, reshapers, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := lifeBefore - me.Life; got != 3 {
		t.Fatalf("took %d damage, want 3", got)
	}
	if me.Library.Size() != 3 {
		t.Fatalf("library %d, want all 3 back", me.Library.Size())
	}
}

func TestRprB14ChromeCourierPutsAnArtifactInHandAndGainsLife(t *testing.T) {
	g, me, _ := rprB14Table(t)
	rock := effectstest.PutOnTopForTest(me, "Revealed Rock", "Artifact")
	land := effectstest.PutOnTopForTest(me, "Revealed Land", "Land")
	life := me.Life
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Chrome Courier", TypeLine: "Artifact Creature — Thopter",
		OracleID: rprB14CourierOracle, Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rock) || !me.Graveyard.Contains(land) {
		t.Fatal("the artifact did not go to hand and the land to the graveyard")
	}
	if me.Life != life+3 {
		t.Fatalf("life %d, want %d", me.Life, life+3)
	}
}

func TestRprB14ChromeCourierGainsNothingForANonartifact(t *testing.T) {
	g, me, _ := rprB14Table(t)
	rock := effectstest.PutOnTopForTest(me, "Revealed Rock", "Artifact")
	land := effectstest.PutOnTopForTest(me, "Revealed Land", "Land")
	life := me.Life
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Chrome Courier", TypeLine: "Artifact Creature — Thopter",
		OracleID: rprB14CourierOracle, Power: 1, Toughness: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) || !me.Graveyard.Contains(rock) {
		t.Fatal("the land did not go to hand and the artifact to the graveyard")
	}
	if me.Life != life {
		t.Fatalf("life %d, want %d: the artifact went to the graveyard", me.Life, life)
	}
}
