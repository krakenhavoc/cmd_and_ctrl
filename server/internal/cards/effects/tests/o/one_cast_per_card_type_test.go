package testso

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestMuldrothaCastsOneCreatureAndRefusesASecond(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	wolf := effectstest.PtGraveCard(me, "Wolf", "Creature — Wolf")

	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatalf("first creature from the graveyard: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if err := effectstest.MuldrothaCast(g, me, wolf, ""); err == nil {
		t.Fatal("a second creature spell from the graveyard was allowed")
	}
	if effectstest.EnumeratedFor(g, me.ID, wolf) {
		t.Error("the enumerator still offers the second creature")
	}
}

func TestMuldrothaEachPermanentTypeOnce(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	for _, tl := range []string{"Artifact", "Creature — Bear", "Enchantment", "Legendary Planeswalker — Test", "Battle — Siege"} {
		id := effectstest.PtGraveCard(me, tl, tl)
		if err := effectstest.MuldrothaCast(g, me, id, ""); err != nil {
			t.Fatalf("%s from the graveyard: %v", tl, err)
		}
		effectstest.ResolveAllPT(t, g)
	}
	sorcery := effectstest.PtGraveCard(me, "Divination", "Sorcery")
	if err := effectstest.MuldrothaCast(g, me, sorcery, ""); err == nil {
		t.Error("a sorcery was cast from the graveyard under Muldrotha")
	}
}

func TestMuldrothaArtifactCreatureChoosesOneType(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	first := effectstest.PtGraveCard(me, "Ornithopter", "Artifact Creature — Thopter")
	second := effectstest.PtGraveCard(me, "Myr", "Artifact Creature — Myr")
	third := effectstest.PtGraveCard(me, "Golem", "Artifact Creature — Golem")

	if err := effectstest.MuldrothaCast(g, me, first, ""); !errors.Is(err, game.ErrPermissionTypeRequired) {
		t.Fatalf("no type named: err = %v, want ErrPermissionTypeRequired", err)
	}
	if err := effectstest.MuldrothaCast(g, me, first, "land"); !errors.Is(err, game.ErrPermissionTypeNotOffered) {
		t.Fatalf("a type the spell lacks: err = %v, want ErrPermissionTypeNotOffered", err)
	}
	if err := effectstest.MuldrothaCast(g, me, first, "artifact"); err != nil {
		t.Fatalf("as the artifact: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	// The ruling: another artifact creature as the creature.
	if err := effectstest.MuldrothaCast(g, me, second, "artifact"); !errors.Is(err, game.ErrPermissionTypeNotOffered) {
		t.Fatalf("artifact again: err = %v, want ErrPermissionTypeNotOffered", err)
	}
	if err := effectstest.MuldrothaCast(g, me, second, ""); err != nil {
		t.Fatalf("the second artifact creature, as the creature (the only type left): %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if err := effectstest.MuldrothaCast(g, me, third, ""); err == nil {
		t.Fatal("a third artifact creature was cast with both types spent")
	}
}

// Muldrotha's ruling: "if you cast a creature spell from your graveyard,
// you can cast a card with bestow as an enchantment spell" — and the
// bestow cost is offered from the graveyard ("if it has an alternative
// cost, you may cast it for that cost instead").
func TestMuldrothaBestowCardIsTheEnchantmentOnceTheCreatureIsSpent(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	host := effectstest.PushVanillaCreature(g, me.ID, "Host", 2, 2)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	satyr := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: satyr, Name: "Boon Satyr", TypeLine: "Enchantment Creature — Satyr",
		OracleID: "aedcd2fb-813a-4915-b7b9-ac7479863462", ManaCost: "{1}{G}{G}", Power: 4, Toughness: 2,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})

	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatal(err)
	}
	effectstest.ResolveAllPT(t, g)
	// Printed, it is an enchantment creature: only the enchantment is left.
	effectstest.PayAnyType(t, g, me, "{G}{G}{G}{G}{G}")
	offered := false
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type == legal.TypeCastSpell && m.Source == satyr {
			var p struct {
				AlternativeCost string `json:"alternative_cost"`
			}
			_ = json.Unmarshal(m.Params, &p)
			offered = offered || p.AlternativeCost == game.BestowKey
		}
	}
	if !offered {
		t.Error("the bestow cast from the graveyard is not offered")
	}
	if err := g.CastSpell(me.ID, satyr, game.CastSpellParams{
		FromZone:        "graveyard",
		AlternativeCost: game.BestowKey,
		Targets:         []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("bestowed from the graveyard as the enchantment: %v", err)
	}
}

func TestMuldrothaLandPlaySpendsLandAndTheLandDrop(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	forest := effectstest.PtGraveCard(me, "Forest", "Basic Land — Forest")
	island := effectstest.PtGraveCard(me, "Island", "Basic Land — Island")

	if err := effectstest.MuldrothaCast(g, me, forest, ""); err != nil {
		t.Fatalf("land from the graveyard: %v", err)
	}
	if !g.Battlefield.Contains(forest) {
		t.Fatal("the land did not enter")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
		t.Errorf("lands played = %d, want 1 — the graveyard land is the turn's land play", got)
	}
	// Another land drop does not reopen Muldrotha's land.
	me.LandDropsPerTurn = 2
	if err := effectstest.MuldrothaCast(g, me, island, ""); err == nil {
		t.Fatal("a second land was played from the graveyard")
	}
}

func TestMuldrothaLandNeedsALandPlayLeft(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	hand := effectstest.HandCardForTest(me, "Plains", "Basic Land — Plains", "")
	if err := g.CastSpell(me.ID, hand, game.CastSpellParams{}); err != nil {
		t.Fatalf("land from hand: %v", err)
	}
	swamp := effectstest.PtGraveCard(me, "Swamp", "Basic Land — Swamp")
	if err := effectstest.MuldrothaCast(g, me, swamp, ""); !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Fatalf("err = %v, want ErrLandDropUnavailable", err)
	}
}

func TestMuldrothaANewMuldrothaGrantsAFreshSet(t *testing.T) {
	g, me, m := effectstest.MuldrothaTable(t)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	wolf := effectstest.PtGraveCard(me, "Wolf", "Creature — Wolf")
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatal(err)
	}
	effectstest.ResolveAllPT(t, g)
	// The same Muldrotha leaves and comes back: a new object (CR 400.7),
	// which is what an epoch bump is.
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == m {
			g.Battlefield.Cards[i].ObjectEpoch++
		}
	}
	if err := effectstest.MuldrothaCast(g, me, wolf, ""); err != nil {
		t.Fatalf("a creature under the new Muldrotha: %v", err)
	}
}

func TestMuldrothaGrantsNothingOffItsControllersTurn(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	effectstest.AdvanceToMainOf(t, g, 1)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err == nil {
		t.Fatal("cast from the graveyard on another player's turn")
	}
}

func TestMuldrothaBudgetRenewsEachOfYourTurns(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	bear := effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	if err := effectstest.MuldrothaCast(g, me, bear, ""); err != nil {
		t.Fatal(err)
	}
	effectstest.ResolveAllPT(t, g)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.AdvanceToMainOf(t, g, 0)
	wolf := effectstest.PtGraveCard(me, "Wolf", "Creature — Wolf")
	if err := effectstest.MuldrothaCast(g, me, wolf, ""); err != nil {
		t.Fatalf("a creature on your next turn: %v", err)
	}
}

// The bot's moves: one per type an artifact creature may spend, the
// type the rest of the graveyard needs least first — a plain creature
// waits behind it, so the artifact goes first.
func TestMuldrothaEnumeratesOneMovePerTypeRankedByNeed(t *testing.T) {
	g, me, _ := effectstest.MuldrothaTable(t)
	effectstest.PtGraveCard(me, "Bear", "Creature — Bear")
	thopter := effectstest.PtGraveCard(me, "Ornithopter", "Artifact Creature — Thopter")
	var types []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != thopter {
			continue
		}
		var p struct {
			PermissionType string `json:"permission_type"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		types = append(types, p.PermissionType)
	}
	if len(types) != 2 || types[0] != "artifact" || types[1] != "creature" {
		t.Fatalf("permission types offered = %v, want [artifact creature]", types)
	}
	// And the view hands the holder the same list, in the same order.
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		if s.ID != me.ID.String() {
			continue
		}
		for _, c := range s.Graveyard.Cards {
			if c.InstanceID == thopter.String() {
				if len(c.PermissionTypes) != 2 || c.PermissionTypes[0] != "artifact" || !c.CastableHere {
					t.Errorf("view: permission_types = %v, castable_here = %v", c.PermissionTypes, c.CastableHere)
				}
				return
			}
		}
	}
	t.Fatal("the artifact creature is not in the viewer's graveyard view")
}

// augurTable stacks seat 0's library so its top eight are the cards
// named, top first, and resolves Aminatou's Augury.
func augurTable(t *testing.T, typeLines ...string) (*game.Game, *game.Player, map[string]uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	ids := map[string]uuid.UUID{}
	for i := len(typeLines) - 1; i >= 0; i-- {
		id := uuid.New()
		ids[typeLines[i]] = id
		me.Library.PushTop(game.Card{InstanceID: id, Name: typeLines[i], TypeLine: typeLines[i], ManaCost: "{5}",
			Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	}
	effectstest.CastCatalogSpell(t, g, "Aminatou's Augury", "Sorcery", effectstest.AminatousAuguryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, ids
}

func augurCast(g *game.Game, p *game.Player, id uuid.UUID, permType string) error {
	return g.CastSpell(p.ID, id, game.CastSpellParams{FromZone: "exile", Strict: true, PermissionType: permType})
}

func TestAminatousAuguryCastsOneOfEachNonlandTypeFree(t *testing.T) {
	g, me, ids := augurTable(t,
		"Artifact Creature — Golem", "Creature — Bear", "Instant", "Sorcery", "Enchantment",
		"Basic Land — Forest", "Artifact", "Creature — Wolf")
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick == nil {
		t.Fatal("no land offer")
	} else {
		effectstest.AnswerChooseCards(t, g, me.ID, ids["Basic Land — Forest"])
	}
	if !effectstest.BattlefieldHasName(g, "Basic Land — Forest") {
		t.Fatal("the chosen land did not enter")
	}
	if got := g.LandsPlayedThisTurnFor(me.ID); got != 0 {
		t.Errorf("lands played = %d — putting the land onto the battlefield is not a land play", got)
	}
	// Strict, and no mana in the pool: every cast below is free.
	if err := augurCast(g, me, ids["Creature — Bear"], ""); err != nil {
		t.Fatalf("the creature, free: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if err := augurCast(g, me, ids["Creature — Wolf"], ""); err == nil {
		t.Fatal("a second creature spell was cast")
	}
	// The artifact creature can only be the artifact now.
	if err := augurCast(g, me, ids["Artifact Creature — Golem"], "creature"); !errors.Is(err, game.ErrPermissionTypeNotOffered) {
		t.Fatalf("as the creature again: err = %v, want ErrPermissionTypeNotOffered", err)
	}
	if err := augurCast(g, me, ids["Artifact Creature — Golem"], ""); err != nil {
		t.Fatalf("the artifact creature as the artifact: %v", err)
	}
	effectstest.ResolveAllPT(t, g)
	if err := augurCast(g, me, ids["Artifact"], ""); err == nil {
		t.Fatal("a second artifact spell was cast")
	}
	for _, tl := range []string{"Instant", "Sorcery", "Enchantment"} {
		if err := augurCast(g, me, ids[tl], ""); err != nil {
			t.Fatalf("%s: %v", tl, err)
		}
		effectstest.ResolveAllPT(t, g)
	}
}

func TestAminatousAuguryLandIsNeverCastAndTheOfferIsOptional(t *testing.T) {
	g, me, ids := augurTable(t, "Artifact Land", "Basic Land — Island", "Creature — Bear",
		"Instant", "Sorcery", "Enchantment", "Artifact", "Creature — Wolf")
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.BattlefieldHasName(g, "Artifact Land") || effectstest.BattlefieldHasName(g, "Basic Land — Island") {
		t.Fatal("a land entered though none was chosen")
	}
	if err := augurCast(g, me, ids["Artifact Land"], ""); err == nil {
		t.Fatal("an artifact land was played from among the exiled cards")
	}
	if err := augurCast(g, me, ids["Basic Land — Island"], ""); err == nil {
		t.Fatal("a land was played from among the exiled cards")
	}
}

func TestAminatousAuguryEndsAtEndOfTurn(t *testing.T) {
	g, me, ids := augurTable(t, "Creature — Bear", "Instant", "Sorcery", "Enchantment",
		"Artifact", "Creature — Wolf", "Creature — Elk", "Creature — Ox")
	effectstest.AdvanceToMainOf(t, g, 1)
	if err := augurCast(g, me, ids["Instant"], ""); err == nil {
		t.Fatal("cast from among the exiled cards on a later turn")
	}
}
