package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestDoomBladeLegalSetExcludesBlackAndNoncreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := effectstest.PushTypedCard(g, opp.ID, "White Knight", "Creature — Human Knight", "{W}{W}")
	black := effectstest.PushTypedCard(g, opp.ID, "Black Knight", "Creature — Human Knight", "{B}{B}")
	mine := effectstest.PushTypedCard(g, me.ID, "My Elf", "Creature — Elf", "{G}")
	rock := effectstest.PushTypedCard(g, opp.ID, "Sol Ring", "Artifact", "{1}")

	legal := effectstest.LegalCards(g, me.ID, effectstest.DoomBladeOracle)
	if !legal[white] || !legal[mine] {
		t.Errorf("non-black creatures should be legal: %v", legal)
	}
	if legal[black] || legal[rock] {
		t.Errorf("black creature / artifact must not be legal: %v", legal)
	}
}

func TestDoomBladeRejectsBlackDestroysWhite(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	white := effectstest.PushTypedCard(g, opp.ID, "White Knight", "Creature — Human Knight", "{W}{W}")
	black := effectstest.PushTypedCard(g, opp.ID, "Black Knight", "Creature — Human Knight", "{B}{B}")

	// Illegal pick: rejected at announce, card stays in hand.
	active := g.Seats[0]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Doom Blade", TypeLine: "Instant",
		OracleID: effectstest.DoomBladeOracle, Owner: active.ID, Controller: active.ID})
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
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(white) || !opp.Graveyard.Contains(white) {
		t.Errorf("White Knight should have been destroyed")
	}
	if !g.Battlefield.Contains(black) {
		t.Errorf("Black Knight should be untouched")
	}
}

func TestNegateLegalSetIsNoncreatureSpellsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// A creature spell and an instant on the stack.
	creatureSpell := effectstest.CastCatalogSpell(t, g, "Colossal Dreadmaw", "Creature — Dinosaur", effectstest.ColossalDreadmawOracle, nil)
	boltID := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: boltID, Name: "Lightning Bolt", TypeLine: "Instant",
		OracleID: effectstest.LightningBoltOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, boltID, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
	}); err != nil {
		t.Fatalf("CastSpell Bolt: %v", err)
	}
	legal := effectstest.LegalCards(g, g.Seats[1].ID, "3407fe41-fdd3-4119-8f70-4bc4590a379f")
	if !legal[boltID] {
		t.Errorf("Negate must be able to target the Bolt")
	}
	if legal[creatureSpell] {
		t.Errorf("Negate must not be able to target a creature spell")
	}
}

func TestTargetAnyOffersPlayersAndCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.PushTypedCard(g, opp.ID, "Bear", "Creature — Bear", "{1}{G}")
	rock := effectstest.PushTypedCard(g, opp.ID, "Sol Ring", "Artifact", "{1}")
	lt := g.LegalTargetsFor(game.SourceChooser(me.ID), effectstest.LightningBoltOracle)
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
		effectstest.DoomBladeOracle:            "creature",
		effectstest.LightningBoltOracle:        "any",
		"3407fe41-fdd3-4119-8f70-4bc4590a379f": "stack_spell",
	} {
		if got := game.TargetModeFor(oracle); got != want {
			t.Errorf("TargetModeFor(%s) = %q, want %q", oracle, got, want)
		}
	}
}
