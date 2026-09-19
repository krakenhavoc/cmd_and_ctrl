package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_permissions_test.go — S42 / ADR 0066. The game package pins
// the model (cast_permission_test.go there); this file pins the eight
// cards end to end against the REAL catalog, which is the half a hook
// stub cannot check: that the card file actually declared the
// permission, that the price it names is the price the cast path
// charges, and that the window really closes.

const (
	snapcasterOracle   = "2bb2eda7-3b38-4c56-870f-c3218a1056f5"
	pastInFlamesOracle = "37a18736-5fe2-4897-809b-013497bdd890"
	grimLockerOracle   = "29931e18-9dac-46fe-b9a4-72d838d79882"
	breachOracle       = "27e0948b-9916-473b-8d8c-a51bdfbc7457"
	citadelOracle      = "2bd111bb-ce02-414c-b5b7-e0e037d8d96b"
	oracleMulDayaID    = "6c5b02eb-7829-436a-8555-fea200e4b67f"
	courserOracle      = "46779609-4fa7-4fd2-b5b4-7d4d749339e6"
	realmwalkerOracle  = "b81eaa2f-0554-41c6-bdf6-d1cb73b8f56f"
)

// pickTriggerTarget answers the pick_target prompt a targeted trigger
// queues (CR 603.3d) with one card.
func pickTriggerTarget(t *testing.T, g *game.Game, chooser, card uuid.UUID) {
	t.Helper()
	pick := latestPickTarget(g, chooser)
	if pick == nil {
		t.Fatalf("no pick_target prompt for %s", chooser)
	}
	if err := g.ResolvePickTarget(pick.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: card}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// grantedPermissionOn is the one read every test here makes: does an
// effect let `seat` cast this card out of the zone it is sitting in?
func grantedPermissionOn(g *game.Game, seat, card uuid.UUID, zone game.ZoneKind) *game.CastPermission {
	var out *game.CastPermission
	g.ReadSnapshot(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			return
		}
		out = g.CastPermissionForLocked(seat, c, zone)
	})
	return out
}

// seedTopOfLibrary replaces the seat's library with one named card on
// top, so a library-top test knows exactly what is there.
func seedTopOfLibrary(p *game.Player, c game.Card) uuid.UUID {
	if c.InstanceID == uuid.Nil {
		c.InstanceID = uuid.New()
	}
	c.Owner, c.Controller = p.ID, p.ID
	p.Library.Cards = nil
	p.Library.PushTop(c)
	return c.InstanceID
}

// --- Snapcaster Mage -----------------------------------------------

// The headline of #652: a card that never printed flashback gains it,
// for one turn, and the cast exiles it on the way out (CR 702.34a).
func TestSnapcasterMageGrantsFlashbackForOneTurn(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	bolt := seedGraveyardCard(t, g, "Filler Bolt", "Instant", "test-granted-bolt")

	castCatalogSpell(t, g, "Snapcaster Mage", "Creature — Human Wizard", snapcasterOracle, nil)
	passPriorityAroundTable(t, g)
	// The ETB trigger targets; answer the pick with the graveyard card
	// and let the trigger resolve.
	pickTriggerTarget(t, g, active.ID, bolt)
	passPriorityAroundTable(t, g)

	perm := grantedPermissionOn(g, active.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() {
		t.Fatalf("Snapcaster granted nothing to the targeted card")
	}
	if perm.AltCostKey != "flashback" {
		t.Errorf("granted key = %q, want flashback", perm.AltCostKey)
	}
	if !perm.ExileOnResolution {
		t.Errorf("the granted flashback does not carry CR 702.34a's exile clause")
	}
	if perm.Cost != "" {
		t.Errorf("granted cost = %q, want empty — \"equal to its mana cost\"", perm.Cost)
	}
	// "Until end of turn": live now, dark once that seat has begun
	// another turn (ADR 0063's seat-turn counter, #945).
	if !permissionLive(g, perm, active.ID) {
		t.Errorf("the grant is not live on the turn it was made")
	}
	later := g.Clone()
	later.Seats[later.Turn.ActiveSeat].TurnsBegun++
	if permissionLive(later, perm, active.ID) {
		t.Errorf("the grant outlived the turn it was made on")
	}
}

// The other half: the cast actually happens, at the card's own mana
// cost, and the card is EXILED rather than left in the graveyard to be
// cast again next turn.
func TestSnapcasterFlashbackCastExilesOnResolution(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	looting := seedGraveyardCard(t, g, "Faithless Looting", "Sorcery", faithlessLootingOracle)

	castCatalogSpell(t, g, "Snapcaster Mage", "Creature — Human Wizard", snapcasterOracle, nil)
	passPriorityAroundTable(t, g)
	pickTriggerTarget(t, g, active.ID, looting)
	passPriorityAroundTable(t, g)

	// The granted key is claimed exactly like a printed one. Faithless
	// Looting prints flashback {2}{R}; the GRANT is not what prices
	// this cast, because a printed offer wins (ADR 0066 decision 3).
	if err := g.CastSpell(active.ID, looting, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback",
	}); err != nil {
		t.Fatalf("granted flashback cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(looting) {
		t.Errorf("the flashed-back card is not in exile")
	}
	if active.Graveyard.Contains(looting) {
		t.Errorf("the flashed-back card returned to the graveyard")
	}
}

// --- Past in Flames ------------------------------------------------

// CR 611.2c, and the reason ScopeCards names OBJECTS rather than
// re-deriving a set: a card that reaches the graveyard AFTER the
// resolution has no flashback, however much of the turn is left.
func TestPastInFlamesLocksTheSetAtResolution(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	early := seedGraveyardCard(t, g, "Early Instant", "Instant", "test-granted-early")

	castCatalogSpell(t, g, "Past in Flames", "Sorcery", pastInFlamesOracle, nil)
	passPriorityAroundTable(t, g)

	if perm := grantedPermissionOn(g, active.ID, early, game.ZoneGraveyard); !perm.Granted() {
		t.Fatalf("a card already in the graveyard gained no flashback")
	}

	// A card that arrives afterwards is not in the locked set.
	late := uuid.New()
	g.WithWriteLock(func() {
		active.Graveyard.PushTop(game.Card{
			InstanceID: late, Name: "Late Instant", TypeLine: "Instant",
			OracleID: "test-granted-late", Owner: active.ID, Controller: active.ID,
		})
	})
	if perm := grantedPermissionOn(g, active.ID, late, game.ZoneGraveyard); perm.Granted() {
		t.Errorf("a card that reached the graveyard after the resolution gained flashback: %+v", perm)
	}

	// And a land in the graveyard was never in the set either — "each
	// instant and sorcery card", as printed.
	land := uuid.New()
	g.WithWriteLock(func() {
		active.Graveyard.PushTop(game.Card{
			InstanceID: land, Name: "Dead Forest", TypeLine: "Basic Land — Forest",
			Owner: active.ID, Controller: active.ID,
		})
	})
	if perm := grantedPermissionOn(g, active.ID, land, game.ZoneGraveyard); perm.Granted() {
		t.Errorf("a land gained flashback from Past in Flames")
	}
}

// Past in Flames prints flashback as well as granting it, which pins
// the precedence rule: a granted key never overrides an offer the card
// itself makes.
func TestPastInFlamesKeepsItsOwnPrintedFlashback(t *testing.T) {
	if !game.CardCastableFromZone(pastInFlamesOracle, game.ZoneGraveyard) {
		t.Fatalf("Past in Flames does not declare the graveyard castable")
	}
	offers := game.AlternativeCostsOfferedFromZone(pastInFlamesOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != "flashback" || offers[0].ManaCost != "{4}{R}" {
		t.Fatalf("Past in Flames graveyard offers = %+v, want one flashback at {4}{R}", offers)
	}
}

// --- The Grim Captain's Locker -------------------------------------

// Past in Flames' shape, pointed at escape and activated rather than
// cast. The set is still locked when the ABILITY resolves.
func TestGrimCaptainsLockerGrantsEscapeToCreatureCards(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	beast := seedGraveyardCard(t, g, "Dead Beast", "Creature — Beast", "test-granted-beast")
	spell := seedGraveyardCard(t, g, "Dead Spell", "Instant", "test-granted-spell")

	locker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "The Grim Captain's Locker",
		TypeLine: "Legendary Artifact", OracleID: grimLockerOracle,
		Owner: active.ID, Controller: active.ID,
	})
	// Ability index 1 is the escape grant; 0 is the surveil.
	if err := g.ActivateCatalogAbility(active.ID, locker, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate the escape grant: %v", err)
	}
	passPriorityAroundTable(t, g)

	perm := grantedPermissionOn(g, active.ID, beast, game.ZoneGraveyard)
	if !perm.Granted() {
		t.Fatalf("the creature card gained no escape")
	}
	if perm.AltCostKey != "escape" || perm.Cost != "{3}{B}" || perm.ExileOtherFromGraveyard != 4 {
		t.Errorf("granted escape = key %q cost %q exile %d, want escape/{3}{B}/4",
			perm.AltCostKey, perm.Cost, perm.ExileOtherFromGraveyard)
	}
	if perm.ExileOnResolution {
		t.Errorf("escape must NOT carry flashback's exile clause (CR 702.138)")
	}
	if p := grantedPermissionOn(g, active.ID, spell, game.ZoneGraveyard); p.Granted() {
		t.Errorf("a non-creature card gained escape: %+v", p)
	}
}

// --- Underworld Breach ---------------------------------------------

// The standing half: true of whatever is in the graveyard while Breach
// is there, and gone the moment it leaves (CR 702.138).
func TestUnderworldBreachGrantsEscapeWhileItRemains(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)

	breach := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Underworld Breach", TypeLine: "Enchantment",
		OracleID: breachOracle, Owner: active.ID, Controller: active.ID,
	})
	// A card that reaches the graveyard AFTER Breach still has escape:
	// the permission is a standing rule, not a locked set.
	bolt := seedGraveyardCard(t, g, "Dead Bolt", "Instant", "test-breach-bolt")

	perm := grantedPermissionOn(g, active.ID, bolt, game.ZoneGraveyard)
	if !perm.Granted() {
		t.Fatalf("Underworld Breach granted no escape")
	}
	if perm.AltCostKey != "escape" || perm.ExileOtherFromGraveyard != 3 {
		t.Errorf("granted escape = key %q, exile %d, want escape/3", perm.AltCostKey, perm.ExileOtherFromGraveyard)
	}
	if perm.Cost != "" {
		t.Errorf("Breach's escape cost = %q, want empty — \"equal to the card's mana cost\"", perm.Cost)
	}

	// A land gets nothing — "each NONLAND card".
	land := uuid.New()
	g.WithWriteLock(func() {
		active.Graveyard.PushTop(game.Card{
			InstanceID: land, Name: "Dead Island", TypeLine: "Basic Land — Island",
			Owner: active.ID, Controller: active.ID,
		})
	})
	if p := grantedPermissionOn(g, active.ID, land, game.ZoneGraveyard); p.Granted() {
		t.Errorf("a land gained escape from Underworld Breach")
	}

	// Breach leaves, and the permission goes with it with nothing
	// having to be swept — it was never stored.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(breach) })
	if p := grantedPermissionOn(g, active.ID, bolt, game.ZoneGraveyard); p.Granted() {
		t.Errorf("escape survived Underworld Breach leaving the battlefield: %+v", p)
	}
}

// --- Bolas's Citadel -----------------------------------------------

// CR 401.5 plus CR 119.4: the top card is castable for life equal to
// its mana value, and a player without the life cannot claim the offer
// at all.
func TestBolassCitadelCastsFromTheTopForLife(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bolas's Citadel", TypeLine: "Legendary Artifact",
		OracleID: citadelOracle, Owner: active.ID, Controller: active.ID,
	})
	top := seedTopOfLibrary(active, game.Card{
		Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{B}",
	})

	perm := grantedPermissionOn(g, active.ID, top, game.ZoneLibrary)
	if !perm.Granted() {
		t.Fatalf("Citadel opened nothing on the top of the library")
	}
	var offer *game.AlternativeCost
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(top)
		offer = perm.AlternativeCostFor(c)
	})
	if offer == nil || offer.Life != 3 {
		t.Fatalf("Citadel's offer = %+v, want 3 life (the card's mana value)", offer)
	}

	lifeBefore := active.Life
	if err := g.CastSpell(active.ID, top, game.CastSpellParams{
		FromZone: "library", AlternativeCost: "bolas_citadel",
	}); err != nil {
		t.Fatalf("cast from the top for life: %v", err)
	}
	if got := lifeBefore - active.Life; got != 3 {
		t.Errorf("life paid = %d, want 3", got)
	}
	if active.Library.Contains(top) {
		t.Errorf("the cast card is still in the library")
	}
}

// CR 118.4 / 119.4: paying life is a COST, so a player who cannot pay
// it cannot claim the offer — the cast is refused rather than
// resolving and killing them.
func TestBolassCitadelRefusesACastYouCannotPayFor(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bolas's Citadel", TypeLine: "Legendary Artifact",
		OracleID: citadelOracle, Owner: active.ID, Controller: active.ID,
	})
	top := seedTopOfLibrary(active, game.Card{
		Name: "Expensive Sorcery", TypeLine: "Sorcery", ManaCost: "{9}{B}",
	})
	g.WithWriteLock(func() { active.Life = 4 })

	err := g.CastSpell(active.ID, top, game.CastSpellParams{
		FromZone: "library", AlternativeCost: "bolas_citadel",
	})
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("cast at 4 life for a mana value of 10: got %v, want ErrInvalidParam", err)
	}
	if active.Life != 4 {
		t.Errorf("a refused cast still paid life: %d", active.Life)
	}
	if !active.Library.Contains(top) {
		t.Errorf("a refused cast moved the card out of the library")
	}
}

// --- Oracle of Mul Daya --------------------------------------------

// A land played off the top is still a land PLAY: it spends the turn's
// drop (CR 305.2), and Oracle's own first line is what pays for it.
func TestOracleOfMulDayaPlaysALandFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Oracle of Mul Daya", TypeLine: "Creature — Elf Shaman",
		OracleID: oracleMulDayaID, Owner: active.ID, Controller: active.ID,
	})
	top := seedTopOfLibrary(active, game.Card{Name: "Top Forest", TypeLine: "Basic Land — Forest"})

	if got := g.LandDropsRemainingFor(active.ID); got != 2 {
		t.Fatalf("land drops with an Oracle out = %d, want 2", got)
	}
	if err := g.CastSpell(active.ID, top, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Errorf("the land never reached the battlefield")
	}
	if got := g.LandsPlayedThisTurnFor(active.ID); got != 1 {
		t.Errorf("land plays this turn = %d, want 1 — the drop was not spent", got)
	}
}

// "You may play LANDS from the top", so a spell on top is visible and
// unplayable. This is the filter doing its job, and the direction that
// would otherwise be a free Future Sight.
func TestOracleOfMulDayaOpensLandsOnly(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Oracle of Mul Daya", TypeLine: "Creature — Elf Shaman",
		OracleID: oracleMulDayaID, Owner: active.ID, Controller: active.ID,
	})
	top := seedTopOfLibrary(active, game.Card{
		Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{R}",
	})
	if perm := grantedPermissionOn(g, active.ID, top, game.ZoneLibrary); perm.Granted() {
		t.Fatalf("Oracle of Mul Daya opened a spell on top of the library: %+v", perm)
	}
	if err := g.CastSpell(active.ID, top, game.CastSpellParams{FromZone: "library"}); err == nil {
		t.Errorf("a sorcery on top was castable under a lands-only permission")
	}
}

// --- Courser of Kruphix --------------------------------------------

// The same two static halves as Oracle, plus the landfall its own
// permission feeds: the land off the top gains the life.
func TestCourserOfKruphixGainsLifeFromItsOwnLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Courser of Kruphix",
		TypeLine: "Enchantment Creature — Centaur", OracleID: courserOracle,
		Owner: active.ID, Controller: active.ID,
	})
	top := seedTopOfLibrary(active, game.Card{Name: "Top Island", TypeLine: "Basic Land — Island"})
	lifeBefore := active.Life

	if err := g.CastSpell(active.ID, top, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := active.Life - lifeBefore; got != 1 {
		t.Errorf("life gained from the landfall = %d, want 1", got)
	}
}

// --- Realmwalker ---------------------------------------------------

// The permission that reads a per-instance CHOICE: a Realmwalker
// naming Elf opens an Elf on top and nothing else, and one that has
// named nothing yet opens nothing at all.
func TestRealmwalkerCastsTheChosenTypeFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	toMain(t, g)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Realmwalker", TypeLine: "Creature — Shapeshifter",
		OracleID: realmwalkerOracle, Owner: active.ID, Controller: active.ID,
	})
	elf := seedTopOfLibrary(active, game.Card{
		Name: "Top Elf", TypeLine: "Creature — Elf", ManaCost: "{G}",
	})

	// No type named yet (CR 614.12 has not been answered): nothing.
	if perm := grantedPermissionOn(g, active.ID, elf, game.ZoneLibrary); perm.Granted() {
		t.Fatalf("a Realmwalker with no chosen type granted a permission: %+v", perm)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards[i].NamedTribe = "Elf"
			}
		}
	})
	perm := grantedPermissionOn(g, active.ID, elf, game.ZoneLibrary)
	if !perm.Granted() {
		t.Fatalf("a Realmwalker naming Elf opened nothing on an Elf")
	}
	if perm.Filter.CreatureType != "Elf" {
		t.Errorf("filter type = %q, want Elf — the choice is read off the source", perm.Filter.CreatureType)
	}

	// A Goblin on top of the same library gets nothing.
	goblin := seedTopOfLibrary(active, game.Card{
		Name: "Top Goblin", TypeLine: "Creature — Goblin", ManaCost: "{R}",
	})
	if p := grantedPermissionOn(g, active.ID, goblin, game.ZoneLibrary); p.Granted() {
		t.Errorf("a Realmwalker naming Elf opened a Goblin: %+v", p)
	}
}

// --- the declarations, as a set ------------------------------------

// A library permission without the matching visibility opens a card
// nobody can see, which is not a play at all. Every card in the family
// prints both halves; this is the guard against a card file that
// declares one and forgets the other.
func TestLibraryPermissionCardsDeclareTheirVisibility(t *testing.T) {
	for _, tc := range []struct {
		name     string
		oracle   string
		want     game.LibraryTopVisibility
		landOnly bool
	}{
		{"Bolas's Citadel", citadelOracle, game.LibraryTopOwner, false},
		{"Oracle of Mul Daya", oracleMulDayaID, game.LibraryTopRevealed, true},
		{"Courser of Kruphix", courserOracle, game.LibraryTopRevealed, true},
		{"Realmwalker", realmwalkerOracle, game.LibraryTopOwner, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := game.CatalogLibraryTopVisible(tc.oracle); got != tc.want {
				t.Errorf("library top visibility = %v, want %v", got, tc.want)
			}
			perms := game.CatalogCastPermissions(tc.oracle)
			if len(perms) != 1 {
				t.Fatalf("declared %d permissions, want 1", len(perms))
			}
			if !perms[0].TopOfLibraryOnly || perms[0].Zone != game.ZoneLibrary {
				t.Errorf("permission = %+v, want a top-of-library one", perms[0])
			}
			if perms[0].Filter.LandsOnly != tc.landOnly {
				t.Errorf("lands-only = %v, want %v", perms[0].Filter.LandsOnly, tc.landOnly)
			}
		})
	}
}
