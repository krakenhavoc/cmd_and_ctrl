package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b409_test.go — Gatecrash, batch 409 (#5565): each test plays the real
// catalog entry through the real cast, activation and trigger paths.

const (
	oldB409MassiveRaid     = "032367fd-e869-463e-aa0f-65e3c6567e60"
	oldB409Perimeter       = "413e880a-bdc8-41bc-923a-566b08e68d0b"
	oldB409Whip            = "44cf7fd5-721f-41f4-ae78-cf93afeff5f1"
	oldB409Specter         = "4ce83f8c-b27e-4e5b-b19b-76b66aaa904b"
	oldB409Phantom         = "af642b8d-a085-4a07-8a1d-ca385d4700f7"
	oldB409Fury            = "ceaa0e02-2dec-412d-86aa-c1129fd0fb08"
	oldB409Champion        = "d1c44856-ca53-4b6e-902e-8d9c6fe1f61c"
	oldB409Collapse        = "de97ec82-d09d-440e-8c7d-a3f0fb552418"
	oldB409GruulCharm      = "e552cab2-cee0-4420-9419-4bada4b2f241"
	oldB409StartingLifeFmt = "life %d → %d, want %d"
)

func TestOldB409CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB409MassiveRaid: "Massive Raid",
		oldB409Perimeter:   "Crackling Perimeter",
		oldB409Whip:        "Razortip Whip",
		oldB409Specter:     "Nightveil Specter",
		oldB409Phantom:     "Leyline Phantom",
		oldB409Fury:        "Aurelia's Fury",
		oldB409Champion:    "Foundry Champion",
		oldB409Collapse:    "Structural Collapse",
		oldB409GruulCharm:  "Gruul Charm",
	} {
		effectstest.StdRegisteredFull(t, oracle, name)
	}
}

// --- Massive Raid ---

func TestOldB409MassiveRaidCountsEveryCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	for _, n := range []string{"A", "B", "C"} {
		effectstest.B12Creature(g, me.ID, n, "Creature — Bear", 1, 1)
	}
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Massive Raid", "Instant", oldB409MassiveRaid, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-3)
	}
}

func TestOldB409MassiveRaidWithNoCreaturesDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Massive Raid", "Instant", oldB409MassiveRaid, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("life %d → %d, want no damage with no creatures", life, opp.Life)
	}
}

// --- Crackling Perimeter ---

func TestOldB409CracklingPerimeterTapsAGateToHitEachOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	perimeter := effectstest.PushCatalogPermanent(g, me.ID, "Crackling Perimeter", "Enchantment", oldB409Perimeter, false)
	if err := g.ActivateCatalogAbility(me.ID, perimeter, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no Gate to tap")
	}
	gate := effectstest.PushCatalogPermanent(g, me.ID, "Test Gate", "Land — Gate", "", false)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	if err := g.ActivateCatalogAbility(me.ID, perimeter, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{gate}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.B16Tapped(t, g, gate) {
		t.Error("the Gate was not tapped")
	}
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %s life %d, want %d", p.ID, p.Life, want)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, perimeter, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{gate}}); err == nil {
		t.Error("tapped the same Gate twice")
	}
}

// --- Razortip Whip ---

func TestOldB409RazortipWhipHitsAnOpponentForOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	whip := effectstest.PushCatalogPermanent(g, me.ID, "Razortip Whip", "Artifact", oldB409Whip, false)
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err == nil {
		t.Fatal("activated with no mana for {1}")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(me.ID)}); err == nil {
		t.Fatal("targeted its own controller")
	}
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, whip, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-1)
	}
	if !effectstest.B16Tapped(t, g, whip) {
		t.Error("the Whip was not tapped")
	}
}

// --- Nightveil Specter ---

func TestOldB409NightveilSpecterExilesTheirTopCardAndYouMayCastIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	prize := effectstest.PushLibraryCard(opp, game.Card{Name: "Prize Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	specter := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Nightveil Specter", OracleID: oldB409Specter,
		TypeLine: "Creature — Specter", Power: 2, Toughness: 3, Keywords: []string{"flying"}})
	_ = specter
	effectstest.AttackWith(t, g, opp.ID, specter)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(prize) {
		t.Fatalf("their top card is in %v, want exile", effectstest.ZoneOf(g, prize))
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(opp.ID, prize, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err == nil {
		t.Error("the card's owner cast the card exiled with the Specter")
	}
	if err := g.CastSpell(me.ID, prize, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile)}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(prize) {
		t.Errorf("the cast card is in %v, want the battlefield", effectstest.ZoneOf(g, prize))
	}
}

func TestOldB409NightveilSpecterWithAnEmptyLibraryExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Library.Cards = nil })
	specter := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Nightveil Specter", OracleID: oldB409Specter,
		TypeLine: "Creature — Specter", Power: 2, Toughness: 3, Keywords: []string{"flying"}})
	effectstest.AttackWith(t, g, opp.ID, specter)
	effectstest.PassPriorityAroundTable(t, g)
	if len(g.Exile.Cards) != 0 {
		t.Errorf("exile holds %d cards, want 0", len(g.Exile.Cards))
	}
}

// --- Leyline Phantom ---

func TestOldB409LeylinePhantomReturnsToHandAfterDealingCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Leyline Phantom", OracleID: oldB409Phantom,
		TypeLine: "Creature — Illusion", Power: 5, Toughness: 5})
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, phantom)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-5)
	}
	if !me.Hand.Contains(phantom) {
		t.Errorf("the Phantom is in %v, want its owner's hand", effectstest.ZoneOf(g, phantom))
	}
}

func TestOldB409LeylinePhantomThatDiedIsNotReturned(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	phantom := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Leyline Phantom", OracleID: oldB409Phantom,
		TypeLine: "Creature — Illusion", Power: 5, Toughness: 5})
	effectstest.AdvanceToMain(t, g)
	effectstest.DealCombatDamageToPlayer(g, phantom, opp.ID, 5)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, phantom); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(phantom) {
		t.Error("a Phantom that left the battlefield was returned to hand")
	}
}

// --- Aurelia's Fury ---

func TestOldB409AureliasFuryTapsCreaturesAndBarsPlayersItDamaged(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 5)
	spared := effectstest.B12Creature(g, third.ID, "Spared", "Creature — Wall", 0, 5)
	life := opp.Life
	effectstest.D1658Cast(t, g, "Aurelia's Fury", "Instant", oldB409Fury, game.CastSpellParams{
		XValue: 3, Targets: append(effectstest.CardRefs(wall), effectstest.PlayerRefs(opp.ID)...),
		Distribution: map[uuid.UUID]int{wall: 1, opp.ID: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-2)
	}
	if !effectstest.B16Tapped(t, g, wall) {
		t.Error("the damaged creature was not tapped")
	}
	if effectstest.B16Tapped(t, g, spared) {
		t.Error("a creature that was not dealt damage was tapped")
	}
	next := effectstest.AvB24Hand(g, opp, game.Card{Name: "Bolt", TypeLine: "Instant"})
	if err := g.CastSpell(opp.ID, next, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a damaged player cast a noncreature spell")
	}
	other := effectstest.AvB24Hand(g, third, game.Card{Name: "Bolt", TypeLine: "Instant"})
	if err := g.CastSpell(third.ID, other, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a player who was not damaged was barred: %v", err)
	}
	_ = me
}

// --- Foundry Champion ---

func TestOldB409FoundryChampionDealsDamageEqualToCreaturesYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 1, 1)
	effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Foundry Champion", "Creature — Elemental Soldier", oldB409Champion, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-3)
	}
}

func TestOldB409FoundryChampionPumpsWithRedAndWhiteMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	champ := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Foundry Champion", OracleID: oldB409Champion,
		TypeLine: "Creature — Elemental Soldier", Power: 4, Toughness: 4})
	for i, mana := range []string{"{R}", "{R}", "{W}"} {
		if err := g.AddManaForEffect(me.ID, uuid.Nil, mana); err != nil {
			t.Fatalf("mana: %v", err)
		}
		idx := 0
		if mana == "{W}" {
			idx = 1
		}
		if err := g.ActivateCatalogAbility(me.ID, champ, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	c := effectstest.E2Card(t, g, champ)
	if c.CurrentPower() != 6 || c.CurrentToughness() != 5 {
		t.Errorf("Champion is %d/%d, want 6/5", c.CurrentPower(), c.CurrentToughness())
	}
}

// --- Structural Collapse ---

func TestOldB409StructuralCollapseSacrificesAnArtifactAndALandThenDeals2(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	otherRock := effectstest.B12Permanent(g, opp.ID, "Other Rock", "Artifact")
	forest := effectstest.B12Permanent(g, opp.ID, "Forest", "Basic Land — Forest")
	keepLand := effectstest.B12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Structural Collapse", "Sorcery", oldB409Collapse, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	for i := 0; i < 2; i++ {
		if c := effectstest.SacrificeChoiceFor(g, opp.ID); c != nil {
			effectstest.AnswerSacrifice(t, g, opp.ID, effectstest.SacrificeOptions(c)[0])
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	gone := 0
	for _, id := range []uuid.UUID{rock, otherRock} {
		if !g.Battlefield.Contains(id) {
			gone++
		}
	}
	if gone != 1 {
		t.Errorf("%d artifacts were sacrificed, want exactly 1", gone)
	}
	gone = 0
	for _, id := range []uuid.UUID{forest, keepLand} {
		if !g.Battlefield.Contains(id) {
			gone++
		}
	}
	if gone != 1 {
		t.Errorf("%d lands were sacrificed, want exactly 1", gone)
	}
	if opp.Life != life-2 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-2)
	}
}

func TestOldB409StructuralCollapseStillDealsDamageWithNothingToSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Structural Collapse", "Sorcery", oldB409Collapse, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf(oldB409StartingLifeFmt, life, opp.Life, life-2)
	}
}

// --- Gruul Charm ---

func TestOldB409GruulCharmGroundsBlockersWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
	effectstest.CastCatalogSpellWithModes(t, g, "Gruul Charm", "Instant", oldB409GruulCharm, []int{0})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.BrAttack(t, g, walker)
	if err := g.DeclareBlocker(bear, walker); err == nil {
		t.Error("a creature without flying blocked")
	}
	if err := g.DeclareBlocker(flyer, walker); err != nil {
		t.Errorf("a creature with flying can't block: %v", err)
	}
}

func TestOldB409GruulCharmTakesBackEveryPermanentYouOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	stolen := effectstest.ApaPush(g, me.ID, opp.ID, game.Card{Name: "Stolen", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Theirs", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.CastCatalogSpellWithModes(t, g, "Gruul Charm", "Instant", oldB409GruulCharm, []int{1})
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.E2Card(t, g, stolen); c.Controller != me.ID {
		t.Error("the permanent I own is still controlled by an opponent")
	}
	if c := effectstest.E2Card(t, g, theirs); c.Controller != opp.ID {
		t.Error("a permanent an opponent owns changed controller")
	}
}

func TestOldB409GruulCharmDamagesOnlyCreaturesWithFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	flyer := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 4, Keywords: []string{"flying"}})
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Flyer", TypeLine: "Creature — Bird", Power: 1, Toughness: 4, Keywords: []string{"flying"}})
	ground := effectstest.B12Creature(g, opp.ID, "Ground", "Creature — Bear", 1, 4)
	effectstest.CastCatalogSpellWithModes(t, g, "Gruul Charm", "Instant", oldB409GruulCharm, []int{2})
	effectstest.PassPriorityAroundTable(t, g)
	if d := effectstest.DamageMarkedOn(g, flyer); d != 3 {
		t.Errorf("their flyer has %d damage, want 3", d)
	}
	if d := effectstest.DamageMarkedOn(g, mine); d != 3 {
		t.Errorf("my flyer has %d damage, want 3 (each creature with flying)", d)
	}
	if d := effectstest.DamageMarkedOn(g, ground); d != 0 {
		t.Errorf("a creature without flying has %d damage", d)
	}
}
