package protocol

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// land_play_castable_view_test.go — #1407. `castable_here` on a LAND in
// a graveyard or on a library top is the land-play rule, not the spell
// rules: CR 305.1's window (the player's own main phase, the stack
// empty) and a land drop left (CR 305.2), through
// game.LandPlayOpenForEffect. #1389 made exile answer it that way; the
// graveyard and the library top asked game.CastTimingOpenLocked, so a
// land stayed lit after the turn's land was played and the click came
// back ErrLandDropUnavailable.
//
// Each zone is pinned three ways: on before the land drop and off
// after it (with the engine's own refusal beside the bit, so the two
// cannot drift), off on an opponent's turn, and a spell in the same
// zone left alone. Every land assertion also asks the bot enumerator,
// whose landPlayMove has read the land rule since #500: the bit and the
// move are one answer.

// publicCard is a card every seat knows, owned and controlled by owner.
func publicCard(g *game.Game, owner uuid.UUID, name, typeLine, cost, oracle string) game.Card {
	c := game.NewCard(name, owner)
	c.TypeLine = typeLine
	c.ManaCost = cost
	c.OracleID = oracle
	c.Controller = owner
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	return c
}

// spendLandDrop plays a land out of p's hand — the real engine path,
// so the tally the view reads is the one CastSpell wrote.
func spendLandDrop(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	for _, c := range p.Hand.Cards {
		if c.IsLand() {
			if err := g.CastSpell(p.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
				t.Fatalf("playing %s from hand: %v", c.Name, err)
			}
			if g.LandDropsRemainingFor(p.ID) != 0 {
				t.Fatalf("land drops left after the play = %d, want 0", g.LandDropsRemainingFor(p.ID))
			}
			return
		}
	}
	t.Fatal("no land in hand to spend the land drop with")
}

// landPlayOffered reports whether the bot enumerator offers seat the
// play of card — the KindLand move, from any zone.
func landPlayOffered(g *game.Game, seat, card uuid.UUID) bool {
	for _, m := range legal.EnumerateLocked(g, seat, legal.Options{}) {
		if m.Kind == legal.KindLand && strings.Contains(string(m.Params), card.String()) {
			return true
		}
	}
	return false
}

// assertLandBit checks a land's `castable_here` against want and
// against the enumerator's land-play move.
func assertLandBit(t *testing.T, g *game.Game, seat uuid.UUID, c *CardView, id uuid.UUID, want bool, label string) {
	t.Helper()
	if c.CastableHere != want {
		t.Errorf("%s: castable_here = %v, want %v", label, c.CastableHere, want)
	}
	if offered := landPlayOffered(g, seat, id); offered != c.CastableHere {
		t.Errorf("%s: the view says castable_here=%v and the enumerator offers the play=%v",
			label, c.CastableHere, offered)
	}
}

// ownGraveyardCard is viewer's own copy of a card in their graveyard.
func ownGraveyardCard(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) *CardView {
	t.Helper()
	v := ViewOfGameFor(g, p.ID.String())
	for i := range v.Seats {
		if v.Seats[i].ID == p.ID.String() {
			if c := cardInZone(v.Seats[i].Graveyard, id); c != nil {
				return c
			}
		}
	}
	t.Fatalf("card %s is not in %s's graveyard view", id, p.Name)
	return nil
}

// libraryTopView is viewer's own copy of the top card of their library.
func libraryTopView(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID) *CardView {
	t.Helper()
	v := ViewOfGameFor(g, p.ID.String())
	for i := range v.Seats {
		if v.Seats[i].ID == p.ID.String() {
			if c := cardInZone(v.Seats[i].Library, id); c != nil {
				return c
			}
		}
	}
	t.Fatalf("card %s is not on %s's library top view", id, p.Name)
	return nil
}

// crucibleOnBoard puts a Crucible of Worlds-shaped permanent under p:
// "You may play lands from your graveyard."
func crucibleOnBoard(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	const oracle = "test-1407-crucible"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:     game.ZoneGraveyard,
		Scope:    game.ScopeStanding,
		Duration: game.WhileInZoneDuration(),
		Filter:   game.PermissionFilter{LandsOnly: true},
		Label:    "Play lands from your graveyard (Crucible of Worlds)",
	})
	g.Battlefield.PushTop(publicCard(g, p.ID, "Test Crucible", "Artifact", "{3}", oracle))
}

// futureSightOnBoard puts a Future Sight-shaped permanent under p:
// "Play with the top card of your library revealed. You may play lands
// and cast spells from the top of your library."
func futureSightOnBoard(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	const oracle = "test-1407-future-sight"
	withTopVisibility(t, oracle, game.LibraryTopRevealed)
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:             game.ZoneLibrary,
		Scope:            game.ScopeStanding,
		Duration:         game.WhileInZoneDuration(),
		TopOfLibraryOnly: true,
		Label:            "Play the top card of your library (Future Sight)",
	})
	g.Battlefield.PushTop(publicCard(g, p.ID, "Test Future Sight", "Enchantment", "{2}{U}{U}{U}", oracle))
}

// Crucible of Worlds: a graveyard land is playable while the land drop
// is unspent, and not once it is — the bit and CastSpell agree.
func TestGraveyardLandCastableHereFollowsTheLandDrop(t *testing.T) {
	g, me, _ := stripTable(t)
	crucibleOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Dead Forest", "Basic Land — Forest", "", "")
	me.Graveyard.PushTop(land)

	assertLandBit(t, g, me.ID, ownGraveyardCard(t, g, me, land.InstanceID), land.InstanceID, true,
		"a Crucible land in its owner's main phase with a land drop left")

	spendLandDrop(t, g, me)
	assertLandBit(t, g, me.ID, ownGraveyardCard(t, g, me, land.InstanceID), land.InstanceID, false,
		"a Crucible land after the land drop is spent (#1407)")
	err := g.CastSpell(me.ID, land.InstanceID, game.CastSpellParams{FromZone: "graveyard"})
	if !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Errorf("playing the graveyard land with no drop left: err = %v, want ErrLandDropUnavailable", err)
	}
}

// Crucible of Worlds on an opponent's turn: the seat has a land drop
// (a fresh turn's worth) but no CR 305.1 window.
func TestGraveyardLandIsNotCastableOnAnOpponentsTurn(t *testing.T) {
	g, me, _ := stripTable(t)
	crucibleOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Dead Island", "Basic Land — Island", "", "")
	me.Graveyard.PushTop(land)

	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepPrecombatMain)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("fixture: expected an opponent's main phase")
	}
	assertLandBit(t, g, me.ID, ownGraveyardCard(t, g, me, land.InstanceID), land.InstanceID, false,
		"a graveyard land on an opponent's turn (CR 305.1)")
}

// A timing STATEMENT does not reach a land (#1407's second half): an
// Orrery-shaped "you may cast spells as though they had flash" leaves a
// graveyard land shut on an opponent's turn — playing a land is not
// casting, and CR 305.1's window is the whole of its timing.
func TestGraveyardLandIgnoresAnOrrery(t *testing.T) {
	const oracleOrrery = "test-1407-orrery"
	g, me, _ := stripTable(t)
	crucibleOnBoard(t, g, me)
	withCastTimings(t, map[string][]game.CastTimingRule{
		oracleOrrery: {{Timing: game.TimingFlash, Label: "You may cast spells as though they had flash."}},
	})
	timingPermanent(g, me, "Test Orrery", oracleOrrery)
	land := publicCard(g, me.ID, "Dead Wastes", "Basic Land — Wastes", "", "")
	me.Graveyard.PushTop(land)

	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepPrecombatMain)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("fixture: expected an opponent's main phase")
	}
	if c := ownGraveyardCard(t, g, me, land.InstanceID); c.CastableHere {
		t.Error("an Orrery lit a graveyard land on an opponent's turn; a land play is not a cast")
	}
	err := g.CastSpell(me.ID, land.InstanceID, game.CastSpellParams{FromZone: "graveyard"})
	if !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("playing the land on an opponent's turn: err = %v, want ErrSorcerySpeedRequired", err)
	}
}

// A cast-only grant over a graveyard land (Realmwalker's verb, pointed
// at the graveyard) never lights it: playing a land is not casting.
func TestGraveyardLandUnderACastOnlyGrantIsNeverCastable(t *testing.T) {
	g, me, _ := stripTable(t)
	land := publicCard(g, me.ID, "Stranded Mountain", "Basic Land — Mountain", "", "")
	me.Graveyard.PushTop(land)
	if !g.GrantCastPermissionOverCardForEffect(land.InstanceID, game.CastPermission{
		Player: me.ID, Zone: game.ZoneGraveyard, CastOnly: true,
		Duration: game.WhileInZoneDuration(), Label: "Cast it from your graveyard",
	}) {
		t.Fatal("grant: card not found")
	}

	if g.LandDropsRemainingFor(me.ID) == 0 {
		t.Fatal("fixture: the land drop should be unspent")
	}
	assertLandBit(t, g, me.ID, ownGraveyardCard(t, g, me, land.InstanceID), land.InstanceID, false,
		"a graveyard land under a cast-only grant (CR 305.1)")
	err := g.CastSpell(me.ID, land.InstanceID, game.CastSpellParams{FromZone: "graveyard"})
	if !errors.Is(err, game.ErrNoPlayPermission) {
		t.Errorf("playing a land under a cast-only grant: err = %v, want ErrNoPlayPermission", err)
	}
}

// A spell in the same graveyard keeps the spell rules: spending the
// land drop does not touch a granted flashback instant.
func TestGraveyardSpellIsUnaffectedByTheLandDrop(t *testing.T) {
	g, me, _ := stripTable(t)
	crucibleOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Dead Plains", "Basic Land — Plains", "", "")
	me.Graveyard.PushTop(land)
	spell := publicCard(g, me.ID, "Granted Think", "Instant", "{2}{U}", "test-1407-think")
	me.Graveyard.PushTop(spell)
	if !g.GrantCastPermissionOverCardForEffect(spell.InstanceID, game.CastPermission{
		Player: me.ID, Zone: game.ZoneGraveyard, AltCostKey: "flashback",
		ExileOnResolution: true, Label: "Flashback — its mana cost",
	}) {
		t.Fatal("grant: card not found")
	}

	if c := ownGraveyardCard(t, g, me, spell.InstanceID); !c.CastableHere {
		t.Fatal("a granted-flashback instant in its owner's main phase is castable")
	}
	spendLandDrop(t, g, me)
	if c := ownGraveyardCard(t, g, me, spell.InstanceID); !c.CastableHere {
		t.Error("spending the land drop turned off a graveyard SPELL's castable_here")
	}
	if c := ownGraveyardCard(t, g, me, land.InstanceID); c.CastableHere {
		t.Error("the land beside it should be off")
	}
}

// Future Sight: a land on top of the library is playable while the
// land drop is unspent, and not once it is.
func TestLibraryTopLandCastableHereFollowsTheLandDrop(t *testing.T) {
	g, me, _ := stripTable(t)
	futureSightOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Top Forest", "Basic Land — Forest", "", "")
	me.Library.PushTop(land)

	assertLandBit(t, g, me.ID, libraryTopView(t, g, me, land.InstanceID), land.InstanceID, true,
		"a Future Sight land on top in its owner's main phase with a land drop left")

	spendLandDrop(t, g, me)
	assertLandBit(t, g, me.ID, libraryTopView(t, g, me, land.InstanceID), land.InstanceID, false,
		"a library-top land after the land drop is spent (#1407)")
	err := g.CastSpell(me.ID, land.InstanceID, game.CastSpellParams{FromZone: "library"})
	if !errors.Is(err, game.ErrLandDropUnavailable) {
		t.Errorf("playing the library-top land with no drop left: err = %v, want ErrLandDropUnavailable", err)
	}
}

// Future Sight on an opponent's turn: the top land is revealed to the
// whole table and playable by nobody.
func TestLibraryTopLandIsNotCastableOnAnOpponentsTurn(t *testing.T) {
	g, me, _ := stripTable(t)
	futureSightOnBoard(t, g, me)
	land := publicCard(g, me.ID, "Top Island", "Basic Land — Island", "", "")

	advanceTo(t, g, game.StepEnd)
	advanceTo(t, g, game.StepPrecombatMain)
	if g.Seats[g.Turn.ActiveSeat].ID == me.ID {
		t.Fatal("fixture: expected an opponent's main phase")
	}
	// Pushed after the turn passes, so no draw step takes it.
	me.Library.PushTop(land)
	assertLandBit(t, g, me.ID, libraryTopView(t, g, me, land.InstanceID), land.InstanceID, false,
		"a library-top land on an opponent's turn (CR 305.1)")
}

// A spell on the same library top keeps the spell rules: with the land
// drop spent, a sorcery on top is still castable in the main phase.
func TestLibraryTopSpellIsUnaffectedByTheLandDrop(t *testing.T) {
	g, me, _ := stripTable(t)
	futureSightOnBoard(t, g, me)
	spendLandDrop(t, g, me)

	spell := publicCard(g, me.ID, "Top Sorcery", "Sorcery", "{1}{U}", "test-1407-top-sorcery")
	me.Library.PushTop(spell)
	if c := libraryTopView(t, g, me, spell.InstanceID); !c.CastableHere {
		t.Error("spending the land drop turned off a library-top SPELL's castable_here")
	}
}
