package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o613DarkBanishing = "0df450be-9bd2-45be-ae59-fa8d19f6a391"
	o613Boil          = "3a51485b-31b2-4fa5-824a-a919f9d28ca8"
	o613Obliterate    = "47496941-8007-45a0-a981-76c683cbddc1"
	o613SeverSoul     = "577d027a-96c3-46fe-880b-f8b5dd3f3a3d"
	o613Flashfires    = "c281f436-8c77-48f7-b31c-d40cd7f9ed6a"
	o613Dwarven       = "caf3c6ec-d17e-497c-8fe7-6f818ce93f96"
	o613Paladin       = "d2cbdbf9-9b08-419e-bb3e-482972d58e74"
	o613Lyrist        = "e0de1c9b-f71f-4498-a968-d3c635cb4c5c"
)

func o613Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func o613Colored(g *game.Game, owner uuid.UUID, name, typeLine, color string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: []string{color},
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestS82OldB613EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		o613DarkBanishing: "Dark Banishing", o613Boil: "Boil", o613Obliterate: "Obliterate",
		o613SeverSoul: "Sever Soul", o613Flashfires: "Flashfires", o613Dwarven: "Dwarven Demolition Team",
		o613Paladin: "Western Paladin", o613Lyrist: "Elvish Lyrist",
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

func TestS82OldB613DarkBanishingDestroysNonblackOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	black := o613Colored(g, opp.ID, "Zombie", "Creature — Zombie", "B", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Dark Banishing", "Instant", o613DarkBanishing, o613Target(black)); err == nil {
		t.Error("a black creature was accepted")
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Dark Banishing", "Instant", o613DarkBanishing, o613Target(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived")
	}
}

func TestS82OldB613SeverSoulGainsLifeEqualToToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	black := o613Colored(g, opp.ID, "Zombie", "Creature — Zombie", "B", 2, 2)
	if err := effectstest.CastCatalogSpellErr(t, g, "Sever Soul", "Sorcery", o613SeverSoul, o613Target(black)); err == nil {
		t.Error("a black creature was accepted")
	}
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 5)
	before := me.Life
	effectstest.CastCatalogSpell(t, g, "Sever Soul", "Sorcery", o613SeverSoul, o613Target(ogre))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived")
	}
	if me.Life != before+5 {
		t.Errorf("life %d -> %d, want +5", before, me.Life)
	}
}

func TestS82OldB613BoilAndFlashfiresDestroyOnlyTheirLandType(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, kind, other string
	}{
		{"Boil", o613Boil, "Instant", "Island", "Plains"},
		{"Flashfires", o613Flashfires, "Sorcery", "Plains", "Island"},
	} {
		g := effectstest.NewCatalogGame(t)
		mine := effectstest.B12Permanent(g, g.Seats[0].ID, "Mine", "Basic Land — "+tc.kind)
		theirs := effectstest.B12Permanent(g, g.Seats[1].ID, "Theirs", "Basic Land — "+tc.kind)
		kept := effectstest.B12Permanent(g, g.Seats[1].ID, "Kept", "Basic Land — "+tc.other)
		creature := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
		effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
			t.Errorf("%s left a %s", tc.name, tc.kind)
		}
		if !g.Battlefield.Contains(kept) || !g.Battlefield.Contains(creature) {
			t.Errorf("%s destroyed something that is not a %s", tc.name, tc.kind)
		}
	}
}

func TestS82OldB613ObliterateDestroysArtifactsCreaturesAndLandsAndCantBeCountered(t *testing.T) {
	if spec, _ := effects.Lookup(o613Obliterate); !spec.CantBeCountered {
		t.Error("Obliterate can be countered")
	}
	g := effectstest.NewCatalogGame(t)
	land := effectstest.B12Permanent(g, g.Seats[0].ID, "Forest", "Basic Land — Forest")
	art := effectstest.B12Permanent(g, g.Seats[1].ID, "Relic", "Artifact")
	bear := effectstest.B12Creature(g, g.Seats[2].ID, "Bear", "Creature — Bear", 2, 2)
	ench := effectstest.B12Permanent(g, g.Seats[1].ID, "Aura", "Enchantment")
	effectstest.CastCatalogSpell(t, g, "Obliterate", "Sorcery", o613Obliterate, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{land, art, bear} {
		if g.Battlefield.Contains(id) {
			t.Error("a land, artifact or creature survived Obliterate")
		}
	}
	if !g.Battlefield.Contains(ench) {
		t.Error("an enchantment was destroyed")
	}
}

func TestS82OldB613WesternPaladinDestroysAWhiteCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pal := effectstest.PushCatalogPermanent(g, me.ID, "Western Paladin", "Creature — Phyrexian Zombie Knight", o613Paladin, false)
	white := o613Colored(g, opp.ID, "Angel", "Creature — Angel", "W", 2, 2)
	green := o613Colored(g, opp.ID, "Elf", "Creature — Elf", "G", 1, 1)
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	}
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{Strict: true, Targets: o613Target(green)}); err == nil {
		t.Fatal("a green creature was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, pal, 0, game.ActivateAbilityParams{Strict: true, Targets: o613Target(white)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) {
		t.Error("the white creature survived")
	}
}

func TestS82OldB613DwarvenDemolitionTeamDestroysAWall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	team := effectstest.PushCatalogPermanent(g, me.ID, "Dwarven Demolition Team", "Creature — Dwarf", o613Dwarven, false)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, team, 0, game.ActivateAbilityParams{Strict: true, Targets: o613Target(bear)}); err == nil {
		t.Fatal("a non-Wall was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, team, 0, game.ActivateAbilityParams{Strict: true, Targets: o613Target(wall)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall survived")
	}
}

func TestS82OldB613ElvishLyristSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lyrist := effectstest.PushCatalogPermanent(g, me.ID, "Elvish Lyrist", "Creature — Elf", o613Lyrist, false)
	ench := effectstest.B12Permanent(g, opp.ID, "Aura", "Enchantment")
	me.ManaPool.AddMana(game.ManaToken{Color: "G"})
	if err := g.ActivateCatalogAbility(me.ID, lyrist, 0, game.ActivateAbilityParams{Strict: true, Targets: o613Target(ench)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ench) {
		t.Error("the enchantment survived")
	}
	if g.Battlefield.Contains(lyrist) {
		t.Error("the Lyrist was not sacrificed")
	}
}
