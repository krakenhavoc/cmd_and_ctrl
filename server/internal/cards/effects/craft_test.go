package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// craft_test.go — CR 702.167, craft (#2124, ADR 0137). The seam is
// pinned here with real cards, through the import road, because the
// return is a transform and a flat fixture has no back face (see
// transform_cards_test.go).
//
// What each test pins:
//
//   - the materials are a cost spanning two zones (CR 702.167b): a
//     permanent you control, a card in your own graveyard, or both;
//   - they are validated before anything is paid (CR 118.3), and the
//     graveyard is the ACTIVATOR's, judged on the card's front face
//     there (CR 712.8a);
//   - the card returns on its back face as a new object, under its
//     owner's control (CR 702.167a, CR 400.7);
//   - the new permanent remembers what was exiled to make it
//     (CR 702.167c), and only what is still in exile.

func tithingBladeRow() cards.Card {
	return transformRow(tithingBladeOracleID, "Tithing Blade", "Artifact", "{1}{B}",
		"Consuming Sepulcher", "Artifact", "", "", []string{"B"})
}

func visageOfDreadRow() cards.Card {
	return transformRow(visageOfDreadOracleID, "Visage of Dread", "Artifact", "{1}{B}",
		"Dread Osseosaur", "Creature — Dinosaur Skeleton Horror", "5", "4", []string{"B"})
}

// pushCraftCard puts an imported transform card onto the battlefield
// under `p`, without running its enters trigger (the craft is the
// subject, not the front face).
func pushCraftCard(g *game.Game, p *game.Player, row cards.Card) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner, c.Controller = p.ID, p.ID
	return pushBattlefieldCardWithTimestamp(g, c)
}

// craftedPermanent finds the one battlefield permanent showing `name`.
func craftedPermanent(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	var found []game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				found = append(found, c)
			}
		}
	})
	if len(found) != 1 {
		t.Fatalf("%d permanents named %q on the battlefield, want 1", len(found), name)
	}
	return found[0]
}

func activateCraft(g *game.Game, p *game.Player, source uuid.UUID, index int, materials ...uuid.UUID) error {
	return g.ActivateCatalogAbility(p.ID, source, index, game.ActivateAbilityParams{
		Strict:            true,
		ExilePermanentIDs: materials,
	})
}

// A creature you control pays "Craft with creature": the Blade and the
// creature are exiled at announce, and on resolution the card returns
// as Consuming Sepulcher — a new object, on its back face, linked to the
// creature it ate.
func TestCraftExilesAPermanentMaterialAndReturnsTransformed(t *testing.T) {
	g, me, _ := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")

	if err := activateCraft(g, me, blade, 0, bear); err != nil {
		t.Fatalf("craft: %v", err)
	}
	if !g.Exile.Contains(blade) || !g.Exile.Contains(bear) {
		t.Fatal("the Blade and the creature must both be in exile once the cost is paid")
	}
	passPriorityAroundTable(t, g)

	sep := craftedPermanent(t, g, "Consuming Sepulcher")
	if sep.InstanceID == blade {
		t.Error("the crafted permanent kept the Blade's instance ID — it must be a new object (CR 400.7)")
	}
	if sep.ActiveFace != 1 {
		t.Errorf("ActiveFace = %d, want the back face", sep.ActiveFace)
	}
	if sep.Controller != me.ID || sep.Owner != me.ID {
		t.Error("the crafted permanent is not under its owner's control")
	}
	if g.Exile.Contains(blade) {
		t.Error("the Blade is still in exile after returning")
	}
	if len(sep.CraftedWith) != 1 || sep.CraftedWith[0].ID != bear {
		t.Fatalf("CraftedWith = %v, want the exiled bear (CR 702.167c)", sep.CraftedWith)
	}
	var mats []game.Card
	g.ReadSnapshot(func() { mats = g.CraftMaterialsForEffect(sep.CraftedWith) })
	if len(mats) != 1 || mats[0].Name != "Grizzly Bears" {
		t.Errorf("CraftMaterialsForEffect = %v, want the bear", mats)
	}
}

// The material may instead be a card in your graveyard (CR 702.167b),
// which leaves the graveyard for exile.
func TestCraftExilesAGraveyardMaterial(t *testing.T) {
	g, me, _ := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	dead := pushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")

	var options []uuid.UUID
	g.ReadSnapshot(func() {
		options = g.ExilePermanentsOptionsForEffect(me.ID, blade, CraftWith("creature").ExilePermanents)
	})
	if len(options) != 1 || options[0] != dead {
		t.Fatalf("craft options = %v, want the graveyard creature card", options)
	}

	if err := activateCraft(g, me, blade, 0, dead); err != nil {
		t.Fatalf("craft from the graveyard: %v", err)
	}
	if me.Graveyard.Contains(dead) || !g.Exile.Contains(dead) {
		t.Fatal("the graveyard material must move to exile at announce")
	}
	passPriorityAroundTable(t, g)
	if sep := craftedPermanent(t, g, "Consuming Sepulcher"); len(sep.CraftedWith) != 1 || sep.CraftedWith[0].ID != dead {
		t.Errorf("CraftedWith = %v, want the graveyard card", sep.CraftedWith)
	}
}

// Everything a craft cost refuses, each refused before anything moves.
func TestCraftRefusesMaterialsTheClauseDoesNotAdmit(t *testing.T) {
	g, me, opp := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	rock := pushCatalogPermanent(g, me.ID, "Mind Stone", "Artifact", "", false)
	theirs := pushGraveyardCardTyped(opp, "Their Bear", "Creature — Bear")
	theirBear := pushCatalogPermanent(g, opp.ID, "Their Live Bear", "Creature — Bear", "", false)
	// A double-faced card whose BACK is a creature is not a creature card
	// in a graveyard: it is its front face there (CR 712.8a).
	visage := deck.ToGameCard(visageOfDreadRow(), false)
	visage.Owner, visage.Controller = me.ID, me.ID
	me.Graveyard.PushTop(visage)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")

	for _, tc := range []struct {
		name string
		ids  []uuid.UUID
		want error
	}{
		{"no material", nil, game.ErrInvalidParam},
		{"a noncreature permanent", []uuid.UUID{rock}, game.ErrIllegalTarget},
		{"the source itself", []uuid.UUID{blade}, game.ErrInvalidParam},
		{"a card in another player's graveyard", []uuid.UUID{theirs}, game.ErrCardNotFound},
		{"another player's creature", []uuid.UUID{theirBear}, game.ErrCardCallerMismatch},
		{"a DFC whose back face is a creature", []uuid.UUID{visage.InstanceID}, game.ErrIllegalTarget},
	} {
		err := activateCraft(g, me, blade, 0, tc.ids...)
		if !errors.Is(err, tc.want) {
			t.Errorf("%s: err = %v, want %v", tc.name, err, tc.want)
		}
	}
	if !battlefieldHas(g, blade) || !me.Graveyard.Contains(visage.InstanceID) || len(me.ManaPool) == 0 {
		t.Error("a refused craft moved something or spent mana")
	}
}

// "Craft only as a sorcery": not with a spell on the stack.
func TestCraftIsSorcerySpeed(t *testing.T) {
	g, me, _ := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	pushCatalogPermanent(g, me.ID, "Prodigal Pyromancer", "Creature — Human Wizard", "", false)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}{R}")
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})

	if err := activateCraft(g, me, blade, 0, bear); err == nil {
		t.Fatal("craft was activated with a spell on the stack")
	}
	if !battlefieldHas(g, blade) || !battlefieldHas(g, bear) {
		t.Error("a refused craft moved something")
	}
}

// The card comes back only if it is still in exile as the object the
// cost put there, and only if it can transform. A copy that is not a
// double-faced card stays in exile (ADR 0079 Decision 5), and nothing is linked
// for a material that is no longer in exile.
func TestReturnCraftedOnlyReturnsADoubleFacedCardStillInExile(t *testing.T) {
	g, me, _ := spendTable(t)
	flat := game.Card{InstanceID: uuid.New(), Name: "Copied Blade", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID}
	g.Exile.PushTop(flat)
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() { entered, err = g.ReturnCraftedFromExileForEffect(flat.InstanceID, nil) })
	if err != nil || entered != uuid.Nil || !g.Exile.Contains(flat.InstanceID) {
		t.Errorf("a single-faced card returned (entered %v, err %v) — it must stay in exile", entered, err)
	}

	blade := deck.ToGameCard(tithingBladeRow(), false)
	blade.Owner, blade.Controller = me.ID, me.ID
	g.Exile.PushTop(blade)
	gone := uuid.New() // a material that left exile
	g.WithWriteLock(func() { entered, err = g.ReturnCraftedFromExileForEffect(blade.InstanceID, []uuid.UUID{gone}) })
	if err != nil || entered == uuid.Nil {
		t.Fatalf("the Blade did not return: entered %v, err %v", entered, err)
	}
	if c, _ := battlefieldCard(g, entered); c.Name != "Consuming Sepulcher" || len(c.CraftedWith) != 0 {
		t.Errorf("returned %q with link %v, want Consuming Sepulcher linked to nothing", c.Name, c.CraftedWith)
	}

	g.WithWriteLock(func() { entered, err = g.ReturnCraftedFromExileForEffect(blade.InstanceID, nil) })
	if err != nil || entered != uuid.Nil {
		t.Errorf("a card no longer in exile returned again (entered %v, err %v)", entered, err)
	}
}

// Register refuses the graveyard half where no printed card puts it: on
// a mana ability, and beside an exile-cards cost that could name the
// same graveyard card.
func TestRegisterRefusesCraftMaterialsWhereNoCardPutsThem(t *testing.T) {
	gy := CraftWith("creature").ExilePermanents
	if msg := registerPanics(Spec{
		OracleID: "00000000-0000-0000-0000-00000000c4a1", Name: "Graveyard Mana",
		ManaAbilities: []ManaAbility{{Cost: ManaAbilityCost{ExilePermanents: gy}, Produced: "{G}"}},
	}); !strings.Contains(msg, "mana ability") {
		t.Errorf("a mana ability with craft materials registered (panic %q)", msg)
	}
	if msg := registerPanics(Spec{
		OracleID: "00000000-0000-0000-0000-00000000c4a2", Name: "Double Exile",
		Activated: []ActivatedAbility{{
			Label:  "Exile a card from your graveyard, craft: nothing.",
			Cost:   Plus(ExileFromGraveyard(1, "a card from your graveyard", nil), CraftWith("creature")),
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}); !strings.Contains(msg, "exile-cards") {
		t.Errorf("craft materials beside an exile-cards cost registered (panic %q)", msg)
	}
}

// The link is per object: the crafted permanent that leaves the
// battlefield forgets it (CR 400.7), and its last-known record keeps it
// for a trigger that resolves afterwards (CR 608.2h).
func TestCraftLinkIsForgottenOnLeavingAndKeptInLastKnownInformation(t *testing.T) {
	g, me, _ := spendTable(t)
	blade := pushCraftCard(g, me, tithingBladeRow())
	bear := pushCatalogPermanent(g, me.ID, "Grizzly Bears", "Creature — Bear", "", false)
	floatMana(t, g, me, "{C}{C}{C}{C}{B}")
	if err := activateCraft(g, me, blade, 0, bear); err != nil {
		t.Fatalf("craft: %v", err)
	}
	passPriorityAroundTable(t, g)
	sep := craftedPermanent(t, g, "Consuming Sepulcher")
	ref := game.ObjectRef{ID: sep.InstanceID, Epoch: sep.ObjectEpoch}

	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(sep.InstanceID); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	if !me.Graveyard.Contains(sep.InstanceID) {
		t.Fatal("the crafted artifact did not go to the graveyard")
	}
	for _, c := range me.Graveyard.Cards {
		if c.InstanceID == sep.InstanceID && (len(c.CraftedWith) != 0 || c.ActiveFace != 0) {
			t.Errorf("in the graveyard it shows face %d with link %v — a card there is its front face and no crafted object", c.ActiveFace, c.CraftedWith)
		}
	}
	var info game.PermanentInfo
	var ok bool
	g.ReadSnapshot(func() { info, ok = g.PermanentForEffect(ref) })
	if !ok || len(info.CraftedWith) != 1 || info.CraftedWith[0].ID != bear {
		t.Errorf("last-known CraftedWith = %v (ok %v), want the bear", info.CraftedWith, ok)
	}
}
