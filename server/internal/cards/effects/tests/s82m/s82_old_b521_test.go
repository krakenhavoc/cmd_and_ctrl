package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eventide chroma cards (S82 older sets, batch 521): every test plays the real
// catalog entry through the engine.

const (
	oldB521UmbraStalker  = "51afd09d-7eee-41c9-bc42-f3f61a176479"
	oldB521Primalcrux    = "8fa6e2fb-3b92-466a-a2cd-b1e5e9a4bf23"
	oldB521LightFromWith = "4d6ad734-b8b5-4d91-9b7b-1fa0c55bfa4f"
	oldB521Springjack    = "325b3469-531e-4ca2-bccb-bb95258c1b54"
	oldB521OutrageShaman = "08d77e5b-7de1-4b0a-a5a7-75e60d452d27"
	oldB521Heartlash     = "04560bc2-df00-4f75-95a0-a63e2ce760b0"
	oldB521FieryBomb     = "85842fa9-e7e8-4380-9034-c93a5497d3d8"
	oldB521SanityGrind   = "bfa0ec94-8851-4518-92e7-ab9f1f90a051"
)

func TestOldB521EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB521UmbraStalker: "Umbra Stalker", oldB521Primalcrux: "Primalcrux",
		oldB521LightFromWith: "Light from Within", oldB521Springjack: "Springjack Shepherd",
		oldB521OutrageShaman: "Outrage Shaman", oldB521Heartlash: "Heartlash Cinder",
		oldB521FieryBomb: "Fiery Bombardment", oldB521SanityGrind: "Sanity Grinding",
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

// oldB521Perm seeds a permanent with a mana cost on the battlefield.
func oldB521Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func oldB521Pick(id uuid.UUID) func(*game.PendingChoice) game.TargetRef {
	return func(*game.PendingChoice) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }
}

func TestOldB521UmbraStalkerCountsBlackSymbolsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushGraveyardPermanent(me, "Two", "Creature — Test", "{B}{B}")
	effectstest.PushGraveyardPermanent(me, "Hybrid", "Creature — Test", "{2}{B/R}")
	effectstest.PushGraveyardPermanent(me, "Red", "Creature — Test", "{R}")
	effectstest.PushGraveyardPermanent(opp, "Theirs", "Creature — Test", "{B}")
	stalker := oldB521Perm(g, me.ID, "Umbra Stalker", "Creature — Elemental", oldB521UmbraStalker, "{4}{B}{B}{B}", 0, 0)
	if p, tough := effectstest.EffectivePower(t, g, stalker), effectstest.EffectiveToughness(t, g, stalker); p != 3 || tough != 3 {
		t.Errorf("Umbra Stalker is %d/%d, want 3/3 (three black symbols, hybrid included, theirs not)", p, tough)
	}
}

func TestOldB521PrimalcruxCountsGreenSymbolsOfYourPermanents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crux := oldB521Perm(g, me.ID, "Primalcrux", "Creature — Elemental", oldB521Primalcrux, "{G}{G}{G}{G}{G}{G}", 0, 0)
	if p := effectstest.EffectivePower(t, g, crux); p != 6 {
		t.Fatalf("alone: power %d, want 6", p)
	}
	oldB521Perm(g, me.ID, "Elf", "Creature — Elf", "", "{1}{G}", 1, 1)
	oldB521Perm(g, opp.ID, "Their Elf", "Creature — Elf", "", "{G}{G}", 1, 1)
	if p, tough := effectstest.EffectivePower(t, g, crux), effectstest.EffectiveToughness(t, g, crux); p != 7 || tough != 7 {
		t.Errorf("Primalcrux is %d/%d, want 7/7", p, tough)
	}
}

func TestOldB521LightFromWithinPumpsEachCreatureByItsOwnWhiteSymbols(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB521Perm(g, me.ID, "Light from Within", "Enchantment", oldB521LightFromWith, "{2}{W}{W}", 0, 0)
	two := oldB521Perm(g, me.ID, "Two", "Creature — Test", "", "{1}{W}{W}", 2, 2)
	red := oldB521Perm(g, me.ID, "Red", "Creature — Test", "", "{R}", 2, 2)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Token", TypeLine: "Token Creature — Test", Power: 1, Toughness: 1,
		Owner: me.ID, Controller: me.ID,
	})
	theirs := oldB521Perm(g, opp.ID, "Theirs", "Creature — Test", "", "{W}", 2, 2)
	for id, want := range map[uuid.UUID]int{two: 4, red: 2, token: 1, theirs: 2} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
}

func TestOldB521SpringjackShepherdMakesAGoatPerWhiteSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB521Perm(g, me.ID, "White", "Creature — Test", "", "{W}{W}", 2, 2)
	oldB521Perm(g, opp.ID, "Theirs", "Creature — Test", "", "{W}{W}", 2, 2)
	effectstest.PlsB1Cast(t, g, "Springjack Shepherd", "Creature — Kithkin Wizard", oldB521Springjack, "{3}{W}", 1, 2)
	effectstest.PlsB1Settle(t, g, nil, nil)
	goats := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goat" && c.Controller == me.ID && c.Power == 0 && c.Toughness == 1 {
			goats++
		}
	}
	if goats != 3 {
		t.Errorf("%d Goats, want 3 (two from the other permanent, one from the Shepherd)", goats)
	}
}

func TestOldB521OutrageShamanDealsDamagePerRedSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB521Perm(g, me.ID, "Red", "Creature — Test", "", "{R}{R}", 2, 2)
	ogre := oldB521Perm(g, opp.ID, "Ogre", "Creature — Ogre", "", "", 5, 5)
	shaman := effectstest.PlsB1Cast(t, g, "Outrage Shaman", "Creature — Goblin Shaman", oldB521OutrageShaman, "{3}{R}{R}", 2, 2)
	effectstest.PlsB1Settle(t, g, oldB521Pick(ogre), nil)
	if got := effectstest.FindCardAnywhere(t, g, ogre).DamageMarked; got != 4 {
		t.Errorf("the Ogre has %d damage, want 4 (the Shaman's two symbols and the other creature's two)", got)
	}
	if !g.Battlefield.Contains(shaman) {
		t.Error("the Shaman left the battlefield")
	}
}

func TestOldB521HeartlashCinderGetsPowerPerRedSymbol(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	oldB521Perm(g, me.ID, "Red", "Creature — Test", "", "{R}{R}", 2, 2)
	cinder := effectstest.PlsB1Cast(t, g, "Heartlash Cinder", "Creature — Elemental Warrior", oldB521Heartlash, "{1}{R}", 1, 1)
	effectstest.PlsB1Settle(t, g, nil, nil)
	if p, tough := effectstest.EffectivePower(t, g, cinder), effectstest.EffectiveToughness(t, g, cinder); p != 4 || tough != 1 {
		t.Errorf("Heartlash Cinder is %d/%d, want 4/1 (+3/+0)", p, tough)
	}
}

func TestOldB521FieryBombardmentDamagesByTheSacrificedCreaturesRedSymbols(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bomb := effectstest.PushCatalogPermanent(g, me.ID, "Fiery Bombardment", "Enchantment", oldB521FieryBomb, false)
	fodder := oldB521Perm(g, me.ID, "Fodder", "Creature — Goblin", "", "{1}{R}{R}", 1, 1)
	life := opp.Life
	effectstest.ApaMana(me, "R", "R")
	effectstest.ApaActivate(t, g, me, bomb, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
		SacrificeIDs: []uuid.UUID{fodder},
	})
	if opp.Life != life-2 {
		t.Errorf("the opponent is at %d, want %d", opp.Life, life-2)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	// A colourless creature deals none.
	plain := oldB521Perm(g, me.ID, "Plain", "Creature — Test", "", "{3}", 1, 1)
	life = opp.Life
	effectstest.ApaMana(me, "R", "R")
	effectstest.ApaActivate(t, g, me, bomb, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
		SacrificeIDs: []uuid.UUID{plain},
	})
	if opp.Life != life {
		t.Errorf("the opponent is at %d after a colourless sacrifice, want %d", opp.Life, life)
	}
}

func TestOldB521SanityGrindingMillsPerBlueSymbolAndBottomsTheReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var ids []uuid.UUID
	// Ten revealed cards: 5 blue symbols between them.
	for i, cost := range []string{"{U}{U}", "{1}{U}", "{U/R}", "{R}", "{2}", "", "{G}", "{W}", "{B}", "{1}{U}"} {
		_ = i
		ids = append(ids, effectstest.SeedLibraryTop(me, "Card", "Creature — Test", cost))
	}
	oppLib := opp.Library.Size()
	myLib := me.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Sanity Grinding", "Sorcery", oldB521SanityGrind,
		[]game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := oppLib - opp.Library.Size(); got != 5 {
		t.Errorf("the opponent milled %d, want 5", got)
	}
	if order := effectstest.PutInLibraryChoiceFor(g, me.ID); order != nil {
		if err := g.ResolvePutInLibrary(order.ID, me.ID, ids, nil); err != nil {
			t.Fatalf("ResolvePutInLibrary: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != myLib {
		t.Errorf("my library is %d cards, want %d", me.Library.Size(), myLib)
	}
	for _, id := range ids {
		if me.Library.Cards[len(me.Library.Cards)-1].InstanceID == id {
			t.Error("a revealed card is still on top of the library")
		}
	}
}
