package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_cast_offers_view_test.go — #978. Before this, the offer stamps
// walked hand, the command zone, the graveyard and the library top,
// all of them per-seat zones; exile is a shared top-level zone, so an
// exiled card a CastPermission let its owner cast reached the client
// with `target_mode` and `mana_cost` and nothing else. #977 had just
// routed the impulse button through the one cast chain, and that chain
// reads the fields that were not there.
//
// What is pinned here: the stamps appear on an exiled card exactly as
// they do on a hand card, they are the HOLDER's answer and only the
// holder's, and the gate is the engine's own castability predicate
// rather than anything this package re-derives.

// withTargetSpec stubs one oracle ID's target clause.
func withTargetSpec(t *testing.T, oracle string, spec func() *game.TargetSpec) {
	t.Helper()
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id == oracle {
			return spec()
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })
}

// withModeSpec stubs one oracle ID's mode list.
func withModeSpec(t *testing.T, oracle string, ms func() *game.ModeSpec) {
	t.Helper()
	prev := game.CatalogModeSpec
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id == oracle {
			return ms()
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })
}

// anyCreatureSpec is "target creature", the simplest clause with a
// legal set that a bystander must not be handed.
func anyCreatureSpec() *game.TargetSpec {
	return &game.TargetSpec{
		Mode:  "creature",
		Label: "target creature",
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1, CountFromX: false,
	}
}

// exileWithGrant drops a card into the shared exile pile, known to
// both seats, carrying `perm`.
func exileWithGrant(t *testing.T, g *game.Game, c game.Card, perm game.CastPermission) uuid.UUID {
	t.Helper()
	c.KnownBy = map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		c.KnownBy[p.ID] = true
	}
	id := c.InstanceID
	g.Exile.PushTop(c)
	if !g.GrantCastPermissionOverCardForEffect(id, perm) {
		t.Fatalf("GrantCastPermissionOverCardForEffect(%s): card not found", c.Name)
	}
	return id
}

// An impulse-exiled {X} spell that targets: the holder gets the whole
// announce surface, and nobody else gets any of it.
func TestExiledSpellUnderAPrintedCostGrantCarriesItsOffers(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-exiled-stroke"
	withTargetSpec(t, oracle, anyCreatureSpec)

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = opp.ID
	g.Battlefield.PushTop(bear)

	stroke := game.NewCard("Test Stroke", me.ID)
	stroke.TypeLine = "Sorcery"
	stroke.ManaCost = "{X}{U}"
	stroke.OracleID = oracle
	id := exileWithGrant(t, g, stroke, game.CastPermission{Player: me.ID})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine == nil {
		t.Fatal("the exiled card is in the holder's view")
	}
	if mine.ExilePlay == nil || mine.ExilePlay.Player != me.ID.String() {
		t.Fatalf("exile_play = %+v, want the holder's grant", mine.ExilePlay)
	}
	if mine.LegalTargets == nil || len(mine.LegalTargets.Cards) != 1 ||
		mine.LegalTargets.Cards[0] != bear.InstanceID.String() {
		t.Errorf("legal_targets = %+v, want the one creature on the battlefield", mine.LegalTargets)
	}
	// The X half: the printed {X} is genuinely being paid, because the
	// grant names no price, so the picker must open and the wire has
	// to carry what bounds it.
	if mine.ExilePlay.XLockedAtZero {
		t.Error("a printed-cost grant does not lock X at zero (CR 107.3b)")
	}
	if mine.ManaCost != "{X}{U}" {
		t.Errorf("mana_cost = %q, want the printed cost the X picker prices against", mine.ManaCost)
	}

	// A bystander sees the card — exile is public — and none of the
	// answer, because a legal target set is narrowed by who is asking.
	theirs := cardInZone(ViewOfGameFor(g, opp.ID.String()).Exile, id)
	if theirs == nil {
		t.Fatal("exile is public: the card is still in the bystander's view")
	}
	if theirs.ExilePlay == nil {
		t.Error("who may play it is public information and stays")
	}
	if theirs.LegalTargets != nil || theirs.Clauses != nil {
		t.Errorf("a bystander was handed the holder's legal set: %+v", theirs.LegalTargets)
	}
	// And a spectator, who is nobody's seat, gets nothing either.
	watcher := cardInZone(ViewOfGameFor(g, "").Exile, id)
	if watcher == nil || watcher.LegalTargets != nil {
		t.Errorf("a spectator was handed a legal set: %+v", watcher)
	}
}

// A modal spell in exile gets its mode picker.
func TestExiledModalSpellCarriesItsModes(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-exiled-modal"
	withModeSpec(t, oracle, func() *game.ModeSpec {
		return &game.ModeSpec{
			Prompt: "Choose one —",
			Min:    1, Max: 1,
			Options: []game.ModeOption{
				{Label: "Draw a card"},
				{Label: "Gain 2 life"},
			},
		}
	})

	charm := game.NewCard("Test Charm", me.ID)
	charm.TypeLine = "Instant"
	charm.ManaCost = "{1}{U}"
	charm.OracleID = oracle
	id := exileWithGrant(t, g, charm, game.CastPermission{Player: me.ID})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine == nil || mine.Modes == nil {
		t.Fatalf("the holder gets the mode picker: %+v", mine)
	}
	if mine.Modes.Min != 1 || mine.Modes.Max != 1 || len(mine.Modes.Options) != 2 {
		t.Errorf("modes = %+v, want choose one of two", mine.Modes)
	}
	theirs := cardInZone(ViewOfGameFor(g, opp.ID.String()).Exile, id)
	if theirs == nil || theirs.Modes != nil {
		t.Errorf("a bystander was handed the mode picker: %+v", theirs)
	}
}

// A permission that names a FACE opens that face and no other
// (ADR 0034), so the clauses stamped are the granted face's — the pile
// is showing the front.
func TestExiledFaceGrantReadsTheGrantedFacesClauses(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-exiled-siege"
	// Only the BACK face targets. If the stamp read the bare oracle
	// ID, as every per-seat zone does, this card would ship nothing.
	withTargetSpec(t, game.CatalogKeyForFace(oracle, 1), anyCreatureSpec)

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = opp.ID
	g.Battlefield.PushTop(bear)

	siege := game.NewCard("Test Siege", me.ID)
	siege.OracleID = oracle
	siege.Layout = game.LayoutTransform
	siege.Faces = []game.Face{
		{Name: "Test Siege", TypeLine: "Battle — Siege", ManaCost: "{2}{R}", StartingDefense: 4},
		{Name: "Test Elemental", TypeLine: "Creature — Elemental", Power: 4, Toughness: 4},
	}
	siege.SetFace(0)
	id := exileWithGrant(t, g, siege, game.CastPermission{Player: me.ID, Cost: "{0}", Faces: []int{1}})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine == nil || mine.ExilePlay == nil ||
		len(mine.ExilePlay.Faces) != 1 || mine.ExilePlay.Faces[0] != 1 {
		t.Fatalf("the grant names the back face: %+v", mine)
	}
	if mine.LegalTargets == nil || len(mine.LegalTargets.Cards) != 1 {
		t.Errorf("legal_targets = %+v, want the BACK face's clause", mine.LegalTargets)
	}
}

// A grant whose window has not opened yet — warp's CR 702.185a "on a
// later turn" — is shown as a greyed button and nothing else: there is
// no cast to target for yet.
func TestExiledCardBeforeItsFloorCarriesNoOffers(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-warped"
	withTargetSpec(t, oracle, anyCreatureSpec)

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = opp.ID
	g.Battlefield.PushTop(bear)

	warped := game.NewCard("Warped Thing", me.ID)
	warped.TypeLine = "Sorcery"
	warped.ManaCost = "{2}{R}"
	warped.OracleID = oracle
	id := exileWithGrant(t, g, warped, game.CastPermission{
		Player:        me.ID,
		Duration:      game.WhileInZoneDuration(),
		NotBeforeTurn: g.Turn.Number + 1,
	})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine == nil || mine.ExilePlay == nil {
		t.Fatalf("the grant is still shown, so the client can grey it: %+v", mine)
	}
	if mine.ExilePlay.NotBeforeTurn != g.Turn.Number+1 {
		t.Errorf("not_before_turn = %d, want %d", mine.ExilePlay.NotBeforeTurn, g.Turn.Number+1)
	}
	if mine.LegalTargets != nil {
		t.Errorf("a window that has not opened carries no target set: %+v", mine.LegalTargets)
	}
}

// The library top gets the same treatment, through the same function —
// Bolas's Citadel opens the top card of your library (CR 401.5), and
// its controller is the only viewer who can even see the card.
func TestLibraryTopUnderCitadelCarriesItsOffers(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const citadel = "test-view-citadel"
	const spell = "test-view-top-bolt"
	withStandingPermission(t, citadel, game.CastPermission{
		Zone:                 game.ZoneLibrary,
		Scope:                game.ScopeStanding,
		TopOfLibraryOnly:     true,
		LifeEqualToManaValue: true,
		AltCostKey:           "bolas_citadel",
		Label:                "Bolas's Citadel — pay life equal to its mana value",
	})
	withTopVisibility(t, citadel, game.LibraryTopOwner)
	withTargetSpec(t, spell, anyCreatureSpec)

	source := game.NewCard("Test Citadel", me.ID)
	source.TypeLine = "Legendary Artifact"
	source.OracleID = citadel
	source.Controller = me.ID
	g.Battlefield.PushTop(source)

	bear := game.NewCard("Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	bear.Controller = opp.ID
	g.Battlefield.PushTop(bear)

	bolt := game.NewCard("Test Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	bolt.OracleID = spell
	me.Library.PushTop(bolt)

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Seats[0].Library, bolt.InstanceID)
	if mine == nil {
		t.Fatal("the controller can see the top card under Citadel")
	}
	if !mine.CastableHere {
		t.Error("castable_here marks a library top a permission opens")
	}
	if mine.LegalTargets == nil || len(mine.LegalTargets.Cards) != 1 {
		t.Errorf("legal_targets = %+v, want the one creature", mine.LegalTargets)
	}
	if len(mine.AlternativeCosts) != 1 || mine.AlternativeCosts[0].Key != "bolas_citadel" {
		t.Errorf("alternative_costs = %+v, want the Citadel offer", mine.AlternativeCosts)
	}

	// The bystander cannot see into the library at all, so there is
	// nothing to strip — which is the same answer by a shorter route.
	theirs := ViewOfGameFor(g, opp.ID.String())
	if c := cardInZone(theirs.Seats[0].Library, bolt.InstanceID); c != nil {
		t.Errorf("a bystander sees the top card of somebody else's library: %+v", c)
	}
}

// The ADR 0073 §7 cast gate reaches exile too, because exile reaches
// stampCastOffers. A grant opens a ZONE; a restriction shuts the cast
// anyway (CR 101.2), and the client has to be told which.
func TestExiledCardCarriesTheCastGatesRefusal(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const clause = "Test Warden — each player can't cast more than one spell each turn."
	stubViewCastRestrictions(t, viewRestrictionOracle, []game.CastRestriction{{
		Label: clause,
		Forbids: func(q game.CastQuery) bool {
			return q.Game.CastTallyFor(q.Controller).Total >= 1
		},
	}})
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Warden",
		TypeLine:   "Enchantment",
		OracleID:   viewRestrictionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})

	bolt := game.NewCard("Impulsed Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	id := exileWithGrant(t, g, bolt, game.CastPermission{Player: me.ID})

	if got := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id).CantCast; got != "" {
		t.Errorf("before any cast, cant_cast = %q, want empty", got)
	}

	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[me.ID] = game.CastTally{Total: 1}
	})

	mine := cardInZone(ViewOfGameFor(g, me.ID.String()).Exile, id)
	if mine.CantCast == "" {
		t.Fatalf("a restricted exiled card carries no cant_cast")
	}
	if !strings.Contains(mine.CantCast, "more than one spell") {
		t.Errorf("cant_cast = %q, want the printed clause", mine.CantCast)
	}
	// And the engine agrees, which is what makes the stamp worth
	// anything.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "exile"}); err == nil {
		t.Errorf("the engine allowed a cast the view greyed out")
	}
	// The reason is the holder's answer, so it goes with the rest.
	if theirs := cardInZone(ViewOfGameFor(g, g.Seats[1].ID.String()).Exile, id); theirs.CantCast != "" {
		t.Errorf("a bystander was handed the holder's refusal: %q", theirs.CantCast)
	}
}

// A permission opens the graveyard; the cast gate shuts the cast.
// stampGrantedPermissions runs after stampLegalTargets and used to
// paint castable_here straight back over the gate's answer.
func TestGrantedGraveyardCardUnderARestrictionIsNotACastSurface(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-restricted-yard"
	stubViewCastRestrictions(t, viewRestrictionOracle, []game.CastRestriction{{
		Label: "Test Warden — each player can't cast more than one spell each turn.",
		Forbids: func(q game.CastQuery) bool {
			return q.Game.CastTallyFor(q.Controller).Total >= 1
		},
	}})
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:       game.ZoneGraveyard,
		Scope:      game.ScopeStanding,
		AltCostKey: "escape",
		Label:      "Escape — its mana cost",
	})

	source := game.NewCard("Test Breach", me.ID)
	source.TypeLine = "Enchantment"
	source.OracleID = oracle
	source.Controller = me.ID
	g.Battlefield.PushTop(source)
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(),
		Name:       "Test Warden",
		TypeLine:   "Enchantment",
		OracleID:   viewRestrictionOracle,
		Owner:      me.ID,
		Controller: me.ID,
	})

	bolt := game.NewCard("Yard Bolt", me.ID)
	bolt.TypeLine = "Instant"
	bolt.ManaCost = "{R}"
	bolt.KnownBy = map[uuid.UUID]bool{me.ID: true, g.Seats[1].ID: true}
	me.Graveyard.PushTop(bolt)

	if c := cardInZone(ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, bolt.InstanceID); !c.CastableHere {
		t.Fatalf("setup: the permission opens the graveyard, so it is a cast surface")
	}

	g.WithWriteLock(func() {
		if g.SpellsCastThisTurn == nil {
			g.SpellsCastThisTurn = make(map[uuid.UUID]game.CastTally)
		}
		g.SpellsCastThisTurn[me.ID] = game.CastTally{Total: 1}
	})

	c := cardInZone(ViewOfGameFor(g, me.ID.String()).Seats[0].Graveyard, bolt.InstanceID)
	if c.CantCast == "" {
		t.Fatalf("the gate's refusal is missing from a granted graveyard card")
	}
	if c.CastableHere {
		t.Error("a permission opened the zone and the gate shut the cast: castable_here must be false")
	}
}
