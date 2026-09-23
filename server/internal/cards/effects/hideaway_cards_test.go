package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// hideaway_cards_test.go — ADR 0091 (#1331) through the catalog: the
// Hideaway trigger's look-and-choose, and the four proof cards' linked
// free play. The engine half (the FaceDownHidden viewer rule, the link,
// playing a face-down card out of exile) is game/hideaway_test.go.

// stackTopLibrary pushes named cards onto the TOP of p's library, the
// first name ending up on top. Returns their IDs in that order.
func stackTopLibrary(p *game.Player, names ...string) []uuid.UUID {
	ids := make([]uuid.UUID, len(names))
	for i := len(names) - 1; i >= 0; i-- {
		c := game.NewCard(names[i], p.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = "{3}{R}"
		ids[i] = c.InstanceID
		p.Library.PushTop(c)
	}
	return ids
}

func hiddenCardsIn(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Exile.Cards {
		if c.FaceDownKind == game.FaceDownHidden {
			out = append(out, c)
		}
	}
	return out
}

// playHiddenGrant is the free-play permission over a hidden card, or
// nil.
func playHiddenGrant(g *game.Game, id uuid.UUID) *game.CastPermission {
	var perm *game.CastPermission
	g.ReadSnapshot(func() { perm = g.CastPermissionOnCardByIDForEffect(id) })
	return perm
}

// enterWithHideaway puts a hideaway card onto the battlefield from
// seat 0's hand through the real entry — cast, or a land play — lets
// the hideaway trigger resolve, and answers its pick with `pick`.
func enterWithHideaway(t *testing.T, g *game.Game, c game.Card, pick uuid.UUID) {
	t.Helper()
	me := g.Seats[0]
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast / play %s: %v", c.Name, err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, pick)
	passPriorityAroundTable(t, g)
}

// CR 702.75a whole, through Rabble Rousing's Hideaway 5: the controller
// looks at five and nobody else does; exactly one is exiled face down,
// hidden by THIS permanent and readable by its controller only; the
// other four go to the bottom.
func TestHideawayExilesOneOfTheTopNFaceDownAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "B", "C", "D", "E")
	rabble := game.NewCard("Rabble Rousing", me.ID)
	rabble.OracleID = rabbleRousingOracleID
	rabble.TypeLine = "Enchantment"
	rabble.ManaCost = "{4}{W}"
	size := me.Library.Size()

	enterWithHideaway(t, g, rabble, top[2])

	hidden := hiddenCardsIn(g)
	if len(hidden) != 1 || hidden[0].InstanceID != top[2] {
		t.Fatalf("hidden cards = %d, want exactly the one picked", len(hidden))
	}
	h := hidden[0]
	if !h.FaceDown || !h.IsKnownTo(me.ID) || h.IsKnownTo(opp.ID) {
		t.Errorf("hidden card: face down %v, known to me %v, to opponent %v", h.FaceDown, h.IsKnownTo(me.ID), h.IsKnownTo(opp.ID))
	}
	if h.HiddenBy.ID != rabble.InstanceID {
		t.Errorf("HiddenBy = %s, want Rabble Rousing", h.HiddenBy.ID)
	}
	if got := me.Library.Size(); got != size-1 {
		t.Errorf("library = %d cards, want %d", got, size-1)
	}
	bottom := map[uuid.UUID]bool{}
	for _, c := range me.Library.Cards[:4] {
		bottom[c.InstanceID] = true
	}
	for i, id := range top {
		if i != 2 && !bottom[id] {
			t.Errorf("card %d of the five is not among the bottom four", i)
		}
	}
}

// Rabble Rousing's attack trigger: one Citizen per attacking creature,
// then — counting the Citizens — ten creatures opens the hidden card
// for free; nine does not.
func TestRabbleRousingOffersTheHiddenCardAtTenCreatures(t *testing.T) {
	for _, tc := range []struct {
		name      string
		bystander int
		wantGrant bool
	}{
		{"nine creatures after the Citizens", 7, false},
		{"ten creatures after the Citizens", 8, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceTo(t, g, game.StepPrecombatMain)
			top := stackTopLibrary(me, "A", "B", "C", "D", "E")
			rabble := game.NewCard("Rabble Rousing", me.ID)
			rabble.OracleID = rabbleRousingOracleID
			rabble.TypeLine = "Enchantment"
			rabble.ManaCost = "{4}{W}"
			enterWithHideaway(t, g, rabble, top[0])

			attacker := pushVanillaCreature(g, me.ID, "Attacker", 1, 1)
			for i := 0; i < tc.bystander; i++ {
				pushVanillaCreature(g, me.ID, "Bystander", 1, 1)
			}
			declareAttack(t, g, opp.ID, attacker)
			passPriorityAroundTable(t, g)

			citizens := 0
			for _, c := range g.Battlefield.Cards {
				if c.Name == "Citizen" && c.Controller == me.ID {
					citizens++
				}
			}
			if citizens != 1 {
				t.Errorf("Citizens = %d, want one per attacking creature", citizens)
			}
			perm := playHiddenGrant(g, top[0])
			if got := perm != nil; got != tc.wantGrant {
				t.Fatalf("grant over the hidden card = %v, want %v", got, tc.wantGrant)
			}
			if perm != nil && (perm.Player != me.ID || perm.Cost != "{0}" || perm.CastOnly) {
				t.Errorf("grant = %+v, want a free PLAY for the controller", perm)
			}
		})
	}
}

// A hideaway land (Windbrisk Heights): enters tapped, hides one of
// four, and its activation plays the card only if its condition holds
// WHEN IT RESOLVES — refused before any attack, granted after three
// creatures attacked. The granted card is then cast for nothing.
func TestWindbriskHeightsPlaysTheHiddenCardAfterThreeAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	top := stackTopLibrary(me, "A", "B", "C", "D")
	heights := game.NewCard("Windbrisk Heights", me.ID)
	heights.OracleID = "3589bcfc-42b0-414a-adce-bc690dc631c8"
	heights.TypeLine = "Land"
	enterWithHideaway(t, g, heights, top[1])
	if c, ok := cardOnBattlefield(g, heights.InstanceID); !ok || !c.Tapped {
		t.Fatal("Windbrisk Heights did not enter tapped")
	}

	activate := func() {
		t.Helper()
		untapForTest(g, heights.InstanceID)
		floatForTest(g, me, "W")
		if err := g.ActivateCatalogAbility(me.ID, heights.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	activate()
	if playHiddenGrant(g, top[1]) != nil {
		t.Fatal("granted with no attack this turn")
	}

	a := pushVanillaCreature(g, me.ID, "A1", 1, 1)
	b := pushVanillaCreature(g, me.ID, "A2", 1, 1)
	c := pushVanillaCreature(g, me.ID, "A3", 1, 1)
	declareAttack(t, g, opp.ID, a, b, c)
	advanceTo(t, g, game.StepPostcombatMain)
	activate()
	if playHiddenGrant(g, top[1]) == nil {
		t.Fatal("no grant after three creatures attacked")
	}
	if err := g.CastSpell(me.ID, top[1], game.CastSpellParams{FromZone: string(game.ZoneExile), Strict: true}); err != nil {
		t.Fatalf("cast the hidden card for free: %v", err)
	}
}

// Mosswort Bridge reads total power at resolution; Spinerock Knoll
// reads damage dealt to ONE opponent this turn.
func TestMosswortBridgeAndSpinerockKnollConditions(t *testing.T) {
	t.Run("Mosswort Bridge", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		top := stackTopLibrary(me, "A", "B", "C", "D")
		bridge := game.NewCard("Mosswort Bridge", me.ID)
		bridge.OracleID = "7cb9e29f-835f-4155-a2a5-4b778866c773"
		bridge.TypeLine = "Land"
		enterWithHideaway(t, g, bridge, top[3])
		activate := func() {
			t.Helper()
			untapForTest(g, bridge.InstanceID)
			floatForTest(g, me, "G")
			if err := g.ActivateCatalogAbility(me.ID, bridge.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
		}
		pushVanillaCreature(g, me.ID, "Big", 9, 9)
		activate()
		if playHiddenGrant(g, top[3]) != nil {
			t.Fatal("granted at total power 9")
		}
		pushVanillaCreature(g, me.ID, "Small", 1, 1)
		activate()
		if playHiddenGrant(g, top[3]) == nil {
			t.Fatal("no grant at total power 10")
		}
	})
	t.Run("Spinerock Knoll", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		advanceTo(t, g, game.StepPrecombatMain)
		top := stackTopLibrary(me, "A", "B", "C", "D")
		knoll := game.NewCard("Spinerock Knoll", me.ID)
		knoll.OracleID = "690c7f8e-fea2-4920-afa7-02ff120701a1"
		knoll.TypeLine = "Land"
		enterWithHideaway(t, g, knoll, top[0])
		activate := func() {
			t.Helper()
			untapForTest(g, knoll.InstanceID)
			floatForTest(g, me, "R")
			if err := g.ActivateCatalogAbility(me.ID, knoll.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)
		}
		g.WithWriteLock(func() {
			// 4 + 3 split across two opponents is not "an opponent was
			// dealt 7".
			_ = g.DealDamageToPlayerForEffect(knoll.InstanceID, opp.ID, 4)
			_ = g.DealDamageToPlayerForEffect(knoll.InstanceID, third.ID, 3)
		})
		activate()
		if playHiddenGrant(g, top[0]) != nil {
			t.Fatal("granted with 7 damage split across two opponents")
		}
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(knoll.InstanceID, opp.ID, 3) })
		activate()
		if playHiddenGrant(g, top[0]) == nil {
			t.Fatal("no grant after one opponent was dealt 7")
		}
	})
}

// CR 607.2a + CR 400.7: a hideaway land that is bounced and replayed is
// a new object. It hides a new card, and its ability plays THAT card —
// never the one its earlier self hid.
func TestAReplayedHideawayLandHasNoClaimOnItsOldCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	first := stackTopLibrary(me, "A", "B", "C", "D")
	bridge := game.NewCard("Mosswort Bridge", me.ID)
	bridge.OracleID = "7cb9e29f-835f-4155-a2a5-4b778866c773"
	bridge.TypeLine = "Land"
	enterWithHideaway(t, g, bridge, first[0])

	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bridge.InstanceID); err != nil {
			t.Fatalf("bounce: %v", err)
		}
	})
	// A new turn's land drop, so the replay is legal.
	g.WithWriteLock(func() { g.LandsPlayedThisTurn[me.ID] = 0 })
	second := stackTopLibrary(me, "E", "F", "G", "H")
	if err := g.CastSpell(me.ID, bridge.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("replay: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID, second[0])
	passPriorityAroundTable(t, g)

	pushVanillaCreature(g, me.ID, "Big", 10, 10)
	untapForTest(g, bridge.InstanceID)
	floatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, bridge.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if playHiddenGrant(g, first[0]) != nil {
		t.Error("the replayed land played the card its earlier self hid")
	}
	if playHiddenGrant(g, second[0]) == nil {
		t.Error("the replayed land did not play the card it hid")
	}
}

// ADR 0091's 2026-09-23 amendment: a hidden LAND is played as part of
// the ability's resolution (CR 608.2g) — during combat, with the stack
// in use, spending a land drop (CR 305.2a). With no land drop left the
// instruction is ignored (CR 305.2b): nobody is asked, and the land
// stays hidden.
func TestAHiddenLandIsPlayedDuringTheResolution(t *testing.T) {
	for _, tc := range []struct {
		name      string
		dropsLeft bool
	}{
		{"with a land drop left", true},
		{"with no land drop left", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceTo(t, g, game.StepPrecombatMain)
			top := stackTopLibrary(me, "A", "Hidden Mountain", "C", "D")
			for i := range me.Library.Cards {
				if me.Library.Cards[i].InstanceID == top[1] {
					me.Library.Cards[i].TypeLine = "Basic Land — Mountain"
					me.Library.Cards[i].ManaCost = ""
				}
			}
			heights := game.NewCard("Windbrisk Heights", me.ID)
			heights.OracleID = "3589bcfc-42b0-414a-adce-bc690dc631c8"
			heights.TypeLine = "Land"
			enterWithHideaway(t, g, heights, top[1]) // spends the turn's land drop
			if tc.dropsLeft {
				g.WithWriteLock(func() { g.LandsPlayedThisTurn[me.ID] = 0 })
			}

			a := pushVanillaCreature(g, me.ID, "A1", 1, 1)
			b := pushVanillaCreature(g, me.ID, "A2", 1, 1)
			c := pushVanillaCreature(g, me.ID, "A3", 1, 1)
			declareAttack(t, g, opp.ID, a, b, c)
			step := g.Turn.Step
			untapForTest(g, heights.InstanceID)
			floatForTest(g, me, "W")
			if err := g.ActivateCatalogAbility(me.ID, heights.InstanceID, 0, game.ActivateAbilityParams{}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			passPriorityAroundTable(t, g)

			answered := answerAllMayCast(t, g, me.ID, true)
			if !tc.dropsLeft {
				if answered != 0 || g.Battlefield.Contains(top[1]) || !g.Exile.Contains(top[1]) {
					t.Fatal("CR 305.2b: with no land drop the land was offered or played")
				}
				return
			}
			if answered != 1 {
				t.Fatalf("answered %d play offers, want 1", answered)
			}
			played, ok := cardOnBattlefield(g, top[1])
			if !ok {
				t.Fatal("the hidden land is not on the battlefield")
			}
			if played.FaceDown || played.Controller != me.ID {
				t.Errorf("played land: face down %v, controller %v", played.FaceDown, played.Controller)
			}
			if g.Turn.Step != step {
				t.Errorf("step moved to %v — the land was played during the resolution, in %v", g.Turn.Step, step)
			}
			if got := g.LandsPlayedThisTurnFor(me.ID); got != 1 {
				t.Errorf("lands played = %d, want 1 — it is a land play (CR 305.2a)", got)
			}
		})
	}
}

func cardOnBattlefield(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
