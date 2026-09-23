package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cost_components_cards_test.go — #1213's card proofs: the seven cards
// the three seam rows were waiting on, each exercised through its own
// printed cost.
//
//	return-to-hand cost   Meloku the Clouded Mirror, Master Transmuter,
//	                      Quirion Ranger, Wirewood Symbiote
//	variable sacrifice    Radiant Lotus (one or more), Grim Hireling (X)
//	mana-ability discard  Skirge Familiar
//
// The engine contracts are pinned in game/return_cost_test.go,
// game/variable_sacrifice_cost_test.go and
// game/mana_discard_cost_test.go; what is here is that each card
// declares the component it prints and that the cost actually pays.

func ccLand(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Owner: owner, Controller: owner,
	})
	return id
}

func ccPermanent(g *game.Game, owner uuid.UUID, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: 1, Toughness: 1, Owner: owner, Controller: owner,
	})
	return id
}

func ccHandCard(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Hand.PushTop(c)
	return c.InstanceID
}

// --- the return-to-hand cost ----------------------------------------

// Meloku: {1} and a land back to hand, and a 1/1 flier off the stack.
func TestMelokuPaysAManaAndALand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meloku := pushCatalogPermanent(g, me.ID, "Meloku the Clouded Mirror",
		"Legendary Creature — Moonfolk Wizard", "3b96b8c3-b9c8-4712-a3b1-243a67cc013a", false)
	land := ccLand(g, me.ID, "Island", "Basic Land — Island")
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	before := len(g.Battlefield.Cards)

	if err := g.ActivateCatalogAbility(me.ID, meloku, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{land},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(land) {
		t.Error("the land did not go back to its owner's hand")
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want the {1} spent", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)
	if got := len(g.Battlefield.Cards); got != before {
		// one land left, one Illusion arrived
		t.Errorf("battlefield has %d cards, want %d — a 1/1 Illusion replaced the land", got, before)
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Illusion" && c.Controller == me.ID {
			found = true
			hasFlying := false
			for _, k := range c.Keywords {
				if k == "flying" {
					hasFlying = true
				}
			}
			if !hasFlying {
				t.Errorf("the Illusion token has no flying: %v", c.Keywords)
			}
		}
	}
	if !found {
		t.Error("no Illusion token was created")
	}
}

// CR 118.3: with no land on the board Meloku's ability cannot be
// activated at all, however much mana is floating.
func TestMelokuCannotActivateWithNoLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	meloku := pushCatalogPermanent(g, me.ID, "Meloku the Clouded Mirror",
		"Legendary Creature — Moonfolk Wizard", "3b96b8c3-b9c8-4712-a3b1-243a67cc013a", false)
	me.ManaPool.AddMana(game.ManaToken{Color: "C"})

	if err := g.ActivateCatalogAbility(me.ID, meloku, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 1 {
		t.Error("a refused activation spent the mana")
	}
}

// Master Transmuter is an artifact you control, so she is a legal pick
// for her own cost — and the ability still resolves with its source in
// hand.
func TestMasterTransmuterMayReturnHerself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tm := pushCatalogPermanent(g, me.ID, "Master Transmuter",
		"Artifact Creature — Human Artificer", "da46786e-28df-4638-ab3d-121011d2f150", false)
	me.ManaPool.AddMana(game.ManaToken{Color: "U"})

	if err := g.ActivateCatalogAbility(me.ID, tm, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{tm},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(tm) {
		t.Error("Master Transmuter paid for her own ability and is still on the battlefield")
	}
	if !me.Hand.Contains(tm) {
		t.Error("Master Transmuter did not reach her owner's hand")
	}
}

// Quirion Ranger returns a Forest and untaps a creature, once each
// turn — and the gate reads the ACTIVATION tally, so a second attempt
// is refused while the first is still on the stack.
func TestQuirionRangerUntapsOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ranger := pushCatalogPermanent(g, me.ID, "Quirion Ranger",
		"Creature — Elf Ranger", "3ecaefc8-ead2-47a3-a7ea-b030faab65a7", false)
	a := ccLand(g, me.ID, "Forest", "Basic Land — Forest")
	b := ccLand(g, me.ID, "Forest", "Basic Land — Forest")
	mana := ccPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid")
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == mana {
			g.Battlefield.Cards[i].Tapped = true
		}
	}

	if err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{a},
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: mana}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Hand.Contains(a) {
		t.Error("the Forest did not go back to hand")
	}
	// The second activation is refused WHILE THE FIRST IS ON THE STACK
	// — an activation count, not a resolution count.
	err := g.ActivateCatalogAbility(me.ID, ranger, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{b},
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: mana}},
	})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("second activation: err = %v, want ErrConditionNotMet", err)
	}
	if !g.Battlefield.Contains(b) {
		t.Error("the refused second activation returned a Forest anyway")
	}
	passPriorityAroundTable(t, g)
	if c, ok := g.LookupCardForEffect(mana); !ok || c.Tapped {
		t.Error("the targeted creature was not untapped")
	}
}

// Wirewood Symbiote is an Insect, so she can never pay with herself —
// the clause says Elf and the filter is the clause.
func TestWirewoodSymbioteCannotReturnHerself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sym := pushCatalogPermanent(g, me.ID, "Wirewood Symbiote",
		"Creature — Insect", "67a52a74-9474-4f4e-8785-6ff54078a8ca", false)
	elf := ccPermanent(g, me.ID, "Priest of Titania", "Creature — Elf Druid")

	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{sym},
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: elf}},
	}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("returning herself: err = %v, want ErrIllegalTarget", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, sym, 0, game.ActivateAbilityParams{
		ReturnIDs: []uuid.UUID{elf},
		Targets:   []game.TargetRef{{Kind: game.TargetCard, ID: sym}},
	}); err != nil {
		t.Fatalf("returning the Elf: %v", err)
	}
	if !me.Hand.Contains(elf) {
		t.Error("the Elf did not go back to hand")
	}
}

// --- the variable sacrifice count -----------------------------------

// Radiant Lotus: the activator names how many artifacts, and the
// effect reads the count back out of the announcement — three mana
// per artifact, to a target player.
func TestRadiantLotusAddsThreeManaPerArtifactSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lotus := pushCatalogPermanent(g, me.ID, "Radiant Lotus", "Artifact",
		"307be184-176a-40a8-944c-48aa00cfdd29", false)
	a := ccPermanent(g, me.ID, "Rock A", "Artifact")
	b := ccPermanent(g, me.ID, "Rock B", "Artifact")

	if err := g.ActivateCatalogAbility(me.ID, lotus, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{a, b},
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a named artifact survived the announcement")
	}
	passPriorityAroundTable(t, g)
	answerColor(t, g, me.ID, "B")
	if got := len(me.ManaPool); got != 6 {
		t.Errorf("mana pool = %d, want 6 — three per artifact sacrificed this way", got)
	}
}

// The floor is one: an announcement that names nothing is refused
// rather than made free.
func TestRadiantLotusNeedsAtLeastOneArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lotus := pushCatalogPermanent(g, me.ID, "Radiant Lotus", "Artifact",
		"307be184-176a-40a8-944c-48aa00cfdd29", false)

	if err := g.ActivateCatalogAbility(me.ID, lotus, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if !g.Battlefield.Contains(lotus) {
		t.Error("a refused activation sacrificed the Lotus")
	}
}

// Grim Hireling: the count IS the announced X, the mana cost is {B}
// whatever X is, and the effect shrinks by that X.
func TestGrimHirelingSacrificesXTreasuresForMinusXMinusX(t *testing.T) {
	g := newCatalogGame(t)
	// "Activate only as a sorcery": the seat needs its own main phase
	// with an empty stack.
	advanceTo(t, g, game.StepPrecombatMain)
	me := g.Seats[0]
	hireling := pushCatalogPermanent(g, me.ID, "Grim Hireling",
		"Creature — Tiefling Rogue", "89738595-dafb-400a-bfba-91a53a37e717", false)
	t1 := ccPermanent(g, me.ID, "Treasure", "Token Artifact — Treasure")
	t2 := ccPermanent(g, me.ID, "Treasure", "Token Artifact — Treasure")
	victim := ccPermanent(g, me.ID, "Bear", "Creature — Bear")
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})

	// The announced X and the payment have to agree.
	if err := g.ActivateCatalogAbility(me.ID, hireling, 0, game.ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{t1},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("X=2 with one Treasure: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, hireling, 0, game.ActivateAbilityParams{
		XValue:       2,
		SacrificeIDs: []uuid.UUID{t1, t2},
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: victim}},
	}); err != nil {
		t.Fatalf("activate at X=2: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("mana pool = %d, want the one {B} spent — X buys permanents, not mana", len(me.ManaPool))
	}
	passPriorityAroundTable(t, g)
	// A 1/1 with -2/-2 is dead, which is the printed outcome.
	if g.Battlefield.Contains(victim) {
		t.Error("the targeted 1/1 survived -2/-2")
	}
}

// --- the mana-ability discard ----------------------------------------

// Skirge Familiar: a card out of hand, a {B} in the pool, and one
// discard event.
func TestSkirgeFamiliarPitchesACardForBlack(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	imp := pushCatalogPermanent(g, me.ID, "Skirge Familiar",
		"Creature — Phyrexian Imp", "ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5", false)
	pitch := ccHandCard(me, "Pitch Me", "Instant", "{3}{R}")

	before := len(g.Events)
	if err := g.ActivateManaAbility(me.ID, imp, 0, game.ManaAbilityParams{
		DiscardIDs: []uuid.UUID{pitch},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the named card is still in hand")
	}
	if !me.Graveyard.Contains(pitch) {
		t.Error("the discarded card did not reach the graveyard")
	}
	if got := len(me.ManaPool); got != 1 || me.ManaPool[0].Color != "B" {
		t.Errorf("mana pool = %+v, want one {B}", me.ManaPool)
	}
	n := 0
	for _, ev := range g.Events[before:] {
		if ev.Kind == game.EventDiscardCard && ev.CardID == pitch {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d EventDiscardCard, want exactly one", n)
	}
}

// With an empty hand there is nothing to pay with, so the ability is
// not activatable and the pool stays empty.
func TestSkirgeFamiliarCannotPayFromAnEmptyHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Hand.Cards = nil
	imp := pushCatalogPermanent(g, me.ID, "Skirge Familiar",
		"Creature — Phyrexian Imp", "ba95f24d-42da-48ce-bcf1-1b7c4b3c45b5", false)

	if err := g.ActivateManaAbility(me.ID, imp, 0, game.ManaAbilityParams{}); !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("err = %v, want ErrInvalidParam", err)
	}
	if len(me.ManaPool) != 0 {
		t.Error("a refused activation minted mana")
	}
}
