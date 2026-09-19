package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// targets_test.go — S20 sub-PR 1 catalog coverage: the predicate
// constructors produce the right legal sets, announce rejects
// illegal picks, and the migrated cards keep working.

const doomBladeOracle = "59e7f2ae-4535-4191-98be-3e65b6b2befa"

func pushTypedCard(g *game.Game, owner uuid.UUID, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
	})
	return id
}

func legalCards(g *game.Game, caster uuid.UUID, oracle string) map[uuid.UUID]bool {
	lt := g.LegalTargetsFor(game.SourceChooser(caster), oracle)
	out := map[uuid.UUID]bool{}
	if lt == nil {
		return out
	}
	for _, id := range lt.Cards {
		out[id] = true
	}
	return out
}

func TestDoomBladeLegalSetExcludesBlackAndNoncreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := pushTypedCard(g, opp.ID, "White Knight", "Creature — Human Knight", "{W}{W}")
	black := pushTypedCard(g, opp.ID, "Black Knight", "Creature — Human Knight", "{B}{B}")
	mine := pushTypedCard(g, me.ID, "My Elf", "Creature — Elf", "{G}")
	rock := pushTypedCard(g, opp.ID, "Sol Ring", "Artifact", "{1}")

	legal := legalCards(g, me.ID, doomBladeOracle)
	if !legal[white] || !legal[mine] {
		t.Errorf("non-black creatures should be legal: %v", legal)
	}
	if legal[black] || legal[rock] {
		t.Errorf("black creature / artifact must not be legal: %v", legal)
	}
}

func TestDoomBladeRejectsBlackDestroysWhite(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	white := pushTypedCard(g, opp.ID, "White Knight", "Creature — Human Knight", "{W}{W}")
	black := pushTypedCard(g, opp.ID, "Black Knight", "Creature — Human Knight", "{B}{B}")

	// Illegal pick: rejected at announce, card stays in hand.
	active := g.Seats[0]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: doomBladeOracle, Owner: active.ID, Controller: active.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: black}},
	}); err != game.ErrIllegalTarget {
		t.Fatalf("Doom Blade on a black creature: got %v, want ErrIllegalTarget", err)
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: white}},
	}); err != nil {
		t.Fatalf("Doom Blade on a white creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || !opp.Graveyard.Contains(white) {
		t.Errorf("White Knight should have been destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Errorf("Black Knight should be untouched")
	}
}

func TestNegateLegalSetIsNoncreatureSpellsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// A creature spell and an instant on the stack.
	creatureSpell := castCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", colossalDreadmawOracle, nil)
	boltID := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bolt: %v", err)
	}
	legal := legalCards(g, g.Seats[1].ID, "3407fe41-fdd3-4119-8f70-4bc4590a379f")
	if !legal[boltID] {
		t.Errorf("Negate must be able to target the Bolt")
	}
	if legal[creatureSpell] {
		t.Errorf("Negate must not be able to target a creature spell")
	}
}

func TestTargetAnyOffersPlayersAndCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := pushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	rock := pushTypedCard(g, opp.ID, "Sol Ring", "Artifact", "{1}")
	lt := g.LegalTargetsFor(game.SourceChooser(me.ID), lightningBoltOracle)
	if lt == nil {
		t.Fatalf("Lightning Bolt has no legal-target set")
	}
	if len(lt.Players) != 4 {
		t.Errorf("any-target players = %d, want 4 (all seats, self included)", len(lt.Players))
	}
	cards := map[uuid.UUID]bool{}
	for _, id := range lt.Cards {
		cards[id] = true
	}
	if !cards[bear] || cards[rock] {
		t.Errorf("any-target cards: bear=%v rock=%v", cards[bear], cards[rock])
	}
}

func TestTargetModeDerivedFromSpec(t *testing.T) {
	for oracle, want := range map[string]string{
		doomBladeOracle:                        "creature",
		lightningBoltOracle:                    "any",
		"3407fe41-fdd3-4119-8f70-4bc4590a379f": "stack_spell",
	} {
		if got := game.TargetModeFor(oracle); got != want {
			t.Errorf("TargetModeFor(%s) = %q, want %q", oracle, got, want)
		}
	}
}
