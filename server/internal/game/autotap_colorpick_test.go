package game

import (
	"testing"

	"github.com/google/uuid"
)

// autotap_colorpick_test.go — regression cover for issue #273.
//
// The auto-tapper plans a payment against the full option set of
// every source it picks, then materializePlanLocked walks the plan
// and turns each slot into a concrete colour. The two halves have to
// agree: if the solver reserved Command Tower for the {U} in
// {1}{W}{U}, the materialiser must not hand back {W}.
//
// Reported board (game fe34c746, turn 2): Plains + Sol Ring +
// Command Tower, casting Teferi, Time Raveler ({1}{W}{U}) with an
// Azorius commander. The pool came out {W}{W}{C}{C} and the strict
// gate refused the cast — the Tower had re-covered the {W} the
// Plains already paid.

// commandTowerOracleID is Command Tower's Scryfall oracle ID, mirrored
// from cards/effects/command_tower.go (the game package can't import
// the catalog).
const commandTowerOracleID = "0895c9b7-ae7d-4bb3-af17-3b75deb50a25"

// commandTowerHook returns Command Tower's catalog ability: the same
// "any colour in your commander's identity" pipe set Arcane Signet
// uses, on a land.
func commandTowerHook(oracleID string) []ManaAbilityShape {
	if oracleID == commandTowerOracleID {
		return []ManaAbilityShape{{
			TapCost:  true,
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color in your commander's color identity",
			// The printed clause, as the catalog declares it.
			NarrowToCommanderIdentity: true,
		}}
	}
	return nil
}

// setCommanderCostForTest stamps a printed mana cost on the player's
// commander so commanderIdentityFor derives a real colour identity
// from it (the shared test deck's commander has no cost).
func setCommanderCostForTest(t *testing.T, p *Player, cost string) {
	t.Helper()
	for i := range p.Command.Cards {
		if p.Command.Cards[i].IsCommander {
			p.Command.Cards[i].ManaCost = cost
			return
		}
	}
	t.Fatalf("no commander in %s's command zone", p.Name)
}

// setCommanderIdentityForTest stamps a transform-DFC commander into
// the seat's command zone: the shape Scryfall actually delivers for
// `layout: "transform"` / `"modal_dfc"`, where the top-level
// mana_cost and colors are null and the real values live on
// card_faces[0]. Only color_identity survives at the top level.
func setCommanderIdentityForTest(t *testing.T, p *Player, identity []string) {
	t.Helper()
	for i := range p.Command.Cards {
		if p.Command.Cards[i].IsCommander {
			p.Command.Cards[i].Name = "Aang, Swift Savior // Aang and La, Ocean's Fury"
			p.Command.Cards[i].ManaCost = ""
			p.Command.Cards[i].Colors = nil
			p.Command.Cards[i].ColorIdentity = identity
			return
		}
	}
	t.Fatalf("no commander in %s's command zone", p.Name)
}

// TestTransformDFCCommanderHasColorIdentity is the issue #276 repro.
// A transform-DFC commander arrives as game.Card{ManaCost: "",
// Colors: nil} because Scryfall puts both on card_faces[0], so
// commanderIdentityFor's Effective().Colors →
// distinctColorsInManaCost chain returned EMPTY — and an empty
// identity makes filterPipeByCommanderIdentity skip narrowing
// entirely. Command Tower, Arcane Signet and Fellwar Stone then
// offered all five colours to an Azorius deck (game fe34c746).
//
// game.Card now carries ColorIdentity, copied straight from the
// Scryfall record at deck import, and commanderIdentityFor prefers
// it. The narrowing is permissive-to-restrictive only: it can remove
// colours that should never have been on offer, never add any.
func TestTransformDFCCommanderHasColorIdentity(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	setCommanderIdentityForTest(t, p, []string{"W", "U"})

	got := commanderIdentityFor(p)
	want := []string{"W", "U"}
	if len(got) != len(want) {
		t.Fatalf("commanderIdentityFor: got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("commanderIdentityFor: got %v, want %v", got, want)
		}
	}

	// The consequence the issue actually reported: the any-colour
	// pipe must narrow to Azorius.
	pipe := filterPipeByCommanderIdentity([]string{"W", "U", "B", "R", "G"}, p)
	if len(pipe) != 2 {
		t.Fatalf("Command Tower pipe: got %v, want [W U]", pipe)
	}
	for _, c := range pipe {
		if c != "W" && c != "U" {
			t.Errorf("Command Tower offered %q to an Azorius commander: %v", c, pipe)
		}
	}
}

// TestCommanderIdentityFallsBackToManaCost guards the fix's blast
// radius. Cards with no stamped ColorIdentity — the demo seed,
// tokens, test fixtures — must keep the pre-#276 behaviour of
// deriving identity from Effective().Colors / the printed cost.
func TestCommanderIdentityFallsBackToManaCost(t *testing.T) {
	g := newActiveGame(t)
	p := g.Seats[0]
	setCommanderCostForTest(t, p, "{1}{W}{U}")

	got := commanderIdentityFor(p)
	if len(got) != 2 {
		t.Fatalf("identity from printed cost: got %v, want two colours", got)
	}
	pipe := filterPipeByCommanderIdentity([]string{"W", "U", "B", "R", "G"}, p)
	if len(pipe) != 2 {
		t.Errorf("pipe from printed-cost identity: got %v, want [W U]", pipe)
	}
}

// TestAutoTapCommandTowerPaysTheColorTheFixedSourceCannot is the
// issue #273 repro. One Plains, one Sol Ring, one Command Tower, an
// Azorius commander, and Teferi, Time Raveler ({1}{W}{U}). The only
// blue in the pool can come from the Tower, so the materialiser must
// spend the Tower on {U} and let the Plains cover {W}.
func TestAutoTapCommandTowerPaysTheColorTheFixedSourceCannot(t *testing.T) {
	withCatalogHook(t, commandTowerHook)
	withCatalogHook(t, solRingHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	setCommanderCostForTest(t, p, "{1}{W}{U}")

	plains := pushBattlefieldForTest(g, p.ID, "Plains", "Basic Land — Plains", "")
	solRing := pushBattlefieldForTest(g, p.ID, "Sol Ring", "Artifact", "6ad8011d-3471-4369-9d68-b264cc027487")
	tower := pushBattlefieldForTest(g, p.ID, "Command Tower", "Land", commandTowerOracleID)

	teferi := pushTypedCardToHandWithCost(p, "Teferi, Time Raveler", "Legendary Planeswalker — Teferi", "{1}{W}{U}")

	if err := g.CastSpell(p.ID, teferi, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast of {1}{W}{U} off Plains + Sol Ring + Command Tower: %v (pool %+v)", err, p.ManaPool)
	}
	if !g.Stack.Contains(teferi) {
		t.Errorf("Teferi did not reach the stack")
	}
	for _, id := range []uuid.UUID{plains, solRing, tower} {
		if !cardTappedForTest(g, id) {
			t.Errorf("source %v should be tapped after the auto-tap cast", id)
		}
	}
	// Sol Ring always makes two {C} and the cost wanted one generic,
	// so exactly one colourless floats. Anything else — a stray {W},
	// or a second {C} — means a slot paid the wrong requirement.
	if len(p.ManaPool) != 1 || p.ManaPool[0].Color != "C" {
		t.Errorf("post-spend pool: got %+v, want one floating {C} from Sol Ring", p.ManaPool)
	}
}

// TestAutoTapFlexibleSourceCoversTheUnpaidRequirement is the same
// failure in its smallest form and without the commander-identity
// machinery: a Mountain and a Birds of Paradise paying {R}{G}. The
// Mountain is fixed on {R}, so Birds owes the {G}.
func TestAutoTapFlexibleSourceCoversTheUnpaidRequirement(t *testing.T) {
	withCatalogHook(t, birdsHook)
	g := newActiveGame(t)
	advanceTo(t, g, StepPrecombatMain)
	p := g.Seats[0]
	pushBattlefieldForTest(g, p.ID, "Mountain", "Basic Land — Mountain", "")
	pushBattlefieldForTest(g, p.ID, "Birds of Paradise", "Creature — Bird", "d3a0b660-358c-41bd-9cd2-41fbf3491b1a")
	id := pushTypedCardToHandWithCost(p, "Manamorphose", "Instant", "{R}{G}")

	if err := g.CastSpell(p.ID, id, CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tap cast of {R}{G} off Mountain + Birds: %v (pool %+v)", err, p.ManaPool)
	}
	if len(p.ManaPool) != 0 {
		t.Errorf("post-spend pool: got %+v, want empty", p.ManaPool)
	}
}

// TestPickColorForSlotConsumesSingleOptionRequirements pins the unit
// that caused #273: a one-option slot has to tick off the requirement
// it satisfies, or a later multi-option slot will pay that same
// requirement a second time and the restrictive one goes unpaid.
func TestPickColorForSlotConsumesSingleOptionRequirements(t *testing.T) {
	cost := costFor(t, "{1}{W}{U}")
	pending := append([]ColorRequirement(nil), cost.Required...)

	if got := pickColorForSlot([]string{"W"}, &pending); got != "W" {
		t.Fatalf("Plains slot: got %q, want \"W\"", got)
	}
	if len(pending) != 1 {
		t.Fatalf("after the Plains slot, pending: got %+v, want just the {U} requirement", pending)
	}
	if got := pickColorForSlot([]string{"W", "U"}, &pending); got != "U" {
		t.Errorf("Command Tower slot: got %q, want \"U\" — the {W} was already paid by the Plains", got)
	}
	if len(pending) != 0 {
		t.Errorf("after both slots, pending: got %+v, want empty", pending)
	}
}

// cardTappedForTest reports whether the battlefield card is tapped.
func cardTappedForTest(g *Game, id uuid.UUID) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c.Tapped
		}
	}
	return false
}
