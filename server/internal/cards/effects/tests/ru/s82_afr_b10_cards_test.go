package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_afr_b10_cards_test.go — S82 AFR batch 10 (#5499): the simple d20
// results tables. Each card is cast or activated through the engine with
// the die pinned to a natural result, so the table row is read from the
// real roll.

const (
	afrSwarmingGoblinsOracle = "6a038dec-2d9c-4422-a0f3-94bf70e55217"
	afrDjinniWindseerOracle  = "f06cecd4-8447-4351-9dfa-861afb84f1e6"
	afrSylvanShepherdOracle  = "ae9f4b85-46ed-4a1a-ada4-54ce56f41815"
	afrLightfootRogueOracle  = "8a825e4c-4a87-4b8d-84f7-3e50b896d7cb"
	afrContactPlaneOracle    = "a352cc67-eae8-4fa5-abba-507821c4cf47"
	afrFaridehsFireOracle    = "927d74c4-afe0-4856-9a02-c3fca9535c2f"
	afrSpikedPitTrapOracle   = "a0aa1aa1-f011-4c65-9886-95521947beef"
	afrScionStygiaOracle     = "a73e1f0a-8895-4eb2-9780-94951a8f227b"
	afrHeraldHadarOracle     = "25ab42a5-ec1a-4510-b957-1d4f14fa10b0"
	afrArcaneInvestOracle    = "19b229c4-1c21-4c7c-a25e-4e8c1cd4ed4b"
	afrAberrantMindOracle    = "753ead8c-f3f8-42fe-931d-79f5ee3ece54"
	afrEarthCultOracle       = "043d8dee-379f-4706-ad47-0bb562717007"
	afrGoblinMorningOracle   = "61ff247e-a119-4f35-af62-2645d6762bba"
)

// afrCast puts the card in hand under id and casts it (free).
func afrCast(t *testing.T, g *game.Game, id uuid.UUID, name, typeLine, oracle string, targets ...game.TargetRef) {
	t.Helper()
	me := g.Seats[0]
	effectstest.D5681Spell(g, me, id, name, typeLine, "", oracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
}

func afrCount(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() { n = effectstest.RandomTokenCount(g, name) })
	return n
}

func TestAFRSwarmingGoblinsTable(t *testing.T) {
	for _, tc := range []struct{ natural, want int }{{1, 1}, {9, 1}, {10, 2}, {19, 2}, {20, 3}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		afrCast(t, g, id, "Swarming Goblins", "Creature — Goblin", afrSwarmingGoblinsOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if got := afrCount(g, "Goblin"); got != tc.want {
			t.Errorf("natural %d: %d Goblins, want %d", tc.natural, got, tc.want)
		}
	}
}

func TestAFRDjinniWindseerScriesByTable(t *testing.T) {
	for _, tc := range []struct{ natural, want int }{{9, 1}, {10, 2}, {19, 2}, {20, 3}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		afrCast(t, g, id, "Djinni Windseer", "Creature — Djinn", afrDjinniWindseerOracle)
		effectstest.PassPriorityAroundTable(t, g)
		scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
		if scry == nil || len(scry.ScryCards) != tc.want {
			t.Fatalf("natural %d: scry prompt %+v, want %d cards", tc.natural, scry, tc.want)
		}
	}
}

func TestAFRSylvanShepherdGainsLifeByTable(t *testing.T) {
	for _, tc := range []struct{ natural, want int }{{9, 1}, {10, 2}, {19, 2}, {20, 5}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Sylvan Shepherd", "Creature — Human Druid", afrSylvanShepherdOracle, false)
		before := me.Life
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Target: g.Seats[1].ID})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if got := me.Life - before; got != tc.want {
			t.Errorf("natural %d: gained %d, want %d", tc.natural, got, tc.want)
		}
	}
}

func TestAFRLightfootRogueTable(t *testing.T) {
	for _, tc := range []struct {
		natural, power int
		first          bool
	}{{9, 1, false}, {10, 2, false}, {19, 2, false}, {20, 4, true}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Lightfoot Rogue", "Creature — Halfling Rogue", afrLightfootRogueOracle, false)
		g.WithWriteLock(func() {
			g.EmitEvent(game.Event{Kind: game.EventAttack, Actor: me.ID, CardID: id, Target: g.Seats[1].ID})
		})
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.EffectivePower(t, g, id); got != tc.power {
			t.Errorf("natural %d: power %d, want %d", tc.natural, got, tc.power)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Errorf("natural %d: no deathtouch", tc.natural)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, id, "first strike"); got != tc.first {
			t.Errorf("natural %d: first strike %v, want %v", tc.natural, got, tc.first)
		}
	}
}

func TestAFRContactOtherPlaneTable(t *testing.T) {
	for _, tc := range []struct{ natural, scry, draw int }{{9, 0, 2}, {10, 2, 2}, {19, 2, 2}, {20, 3, 3}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		before := me.Hand.Size()
		afrCast(t, g, id, "Contact Other Plane", "Instant", afrContactPlaneOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if tc.scry > 0 {
			scry := effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry)
			if scry == nil || len(scry.ScryCards) != tc.scry {
				t.Fatalf("natural %d: scry prompt %+v, want %d", tc.natural, scry, tc.scry)
			}
			if got := me.Hand.Size(); got != before {
				t.Fatalf("natural %d: drew before the scry was answered", tc.natural)
			}
			if err := g.ResolveScry(scry.ID, me.ID, nil, scry.ScryCards); err != nil {
				t.Fatal(err)
			}
		}
		if got := me.Hand.Size() - before; got != tc.draw {
			t.Errorf("natural %d: drew %d, want %d", tc.natural, got, tc.draw)
		}
	}
}

func TestAFRFaridehsFireballTable(t *testing.T) {
	for _, tc := range []struct {
		natural int
		self    int
	}{{9, 2}, {10, 0}, {20, 0}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		bear := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 5)
		afrCast(t, g, id, "Farideh's Fireball", "Instant", afrFaridehsFireOracle, effectstest.P2Card(bear))
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if me.Life != 40-tc.self {
			t.Errorf("natural %d: my life %d, want %d", tc.natural, me.Life, 40-tc.self)
		}
		for _, opp := range g.Seats[1:] {
			if opp.Life != 38 {
				t.Errorf("natural %d: opponent life %d, want 38", tc.natural, opp.Life)
			}
		}
		if !effectstest.D5681InZone(g, bear, game.ZoneGraveyard) {
			t.Errorf("natural %d: the target took 5 and did not die", tc.natural)
		}
	}
}

func TestAFRFaridehsFireballIllegalTargetDoesNothing(t *testing.T) {
	id := uuid.New()
	g := effectstest.D5681Table(t, id, 15)
	bear := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Bear", 2, 2)
	afrCast(t, g, id, "Farideh's Fireball", "Instant", afrFaridehsFireOracle, effectstest.P2Card(bear))
	effectstest.D5681Dies(t, g, g.Seats[1].ID, bear)
	from := effectstest.SfkEventCount(g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.D5681Rolled(g, from); len(got) != 0 {
		t.Errorf("a fizzled spell rolled %+v", got)
	}
	if g.Seats[1].Life != 40 {
		t.Errorf("opponent life %d, want 40", g.Seats[1].Life)
	}
	// Refused at announce: not a creature or planeswalker.
	g2 := effectstest.D5681Table(t, uuid.New(), 15)
	land := effectstest.PushCatalogPermanent(g2, g2.Seats[1].ID, "Forest", "Land", "", false)
	id2 := uuid.New()
	effectstest.D5681Spell(g2, g2.Seats[0], id2, "Farideh's Fireball", "Instant", "", afrFaridehsFireOracle)
	if err := g2.CastSpell(g2.Seats[0].ID, id2, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{effectstest.P2Card(land)}}); err == nil {
		t.Error("targeting a land was allowed")
	}
}

func TestAFRSpikedPitTrapTable(t *testing.T) {
	for _, tc := range []struct {
		natural, treasures int
	}{{9, 0}, {10, 1}, {20, 1}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Spiked Pit Trap", "Artifact", afrSpikedPitTrapOracle, false)
		wall := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 5)
		if err := effectstest.P2Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}", effectstest.P2Card(wall)); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		g.RunStateChecksForTest()
		if !effectstest.D5681InZone(g, wall, game.ZoneGraveyard) {
			t.Errorf("natural %d: the creature survived 5 damage", tc.natural)
		}
		if !effectstest.D5681InZone(g, id, game.ZoneGraveyard) {
			t.Errorf("natural %d: the trap was not sacrificed", tc.natural)
		}
		if got := afrCount(g, "Treasure"); got != tc.treasures {
			t.Errorf("natural %d: %d Treasures, want %d", tc.natural, got, tc.treasures)
		}
	}
}

func TestAFRSpikedPitTrapRefusals(t *testing.T) {
	id := uuid.New()
	g := effectstest.D5681Table(t, id, 15)
	me := g.Seats[0]
	effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Spiked Pit Trap", "Artifact", afrSpikedPitTrapOracle, false)
	wall := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Wall", 0, 5)
	effectstest.FloatForTest(g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.P2Card(wall)}, Strict: true}); err == nil {
		t.Error("activated with only four mana")
	}
	land := effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Land", "", false)
	if err := effectstest.P2Activate(t, g, me, id, 0, "{C}", effectstest.P2Card(land)); err == nil {
		t.Error("targeted a land")
	}
}

func TestAFRScionOfStygiaTable(t *testing.T) {
	for _, tc := range []struct {
		natural int
		stays   bool
	}{{9, false}, {10, true}, {20, true}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		bear := effectstest.PushVanillaCreature(g, g.Seats[1].ID, "Bear", 2, 2)
		afrCast(t, g, id, "Scion of Stygia", "Creature — Tiefling Shaman", afrScionStygiaOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, g.Seats[0].ID) != nil {
			effectstest.PickTriggerTarget(t, g, g.Seats[0].ID, bear)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.IsTapped(g, bear) {
			t.Errorf("natural %d: the creature was not tapped", tc.natural)
		}
		_ = tc.stays
	}
}

func TestAFRHeraldOfHadarTable(t *testing.T) {
	for _, tc := range []struct {
		natural, gain, treasures int
	}{{9, 0, 0}, {10, 2, 0}, {19, 2, 0}, {20, 2, 2}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Herald of Hadar", "Creature — Human Warlock", afrHeraldHadarOracle, false)
		before := me.Life
		if err := effectstest.P2Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}{B}"); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		for _, opp := range g.Seats[1:] {
			if opp.Life != 38 {
				t.Errorf("natural %d: opponent life %d, want 38", tc.natural, opp.Life)
			}
		}
		if me.Life-before != tc.gain {
			t.Errorf("natural %d: gained %d, want %d", tc.natural, me.Life-before, tc.gain)
		}
		if got := afrCount(g, "Treasure"); got != tc.treasures {
			t.Errorf("natural %d: %d Treasures, want %d", tc.natural, got, tc.treasures)
		}
	}
}

func TestAFRArcaneInvestigatorTable(t *testing.T) {
	for _, natural := range []int{9, 10, 20} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, natural)
		me := g.Seats[0]
		effectstest.PushCatalogPermanentWithID(g, id, me.ID, "Arcane Investigator", "Creature — Elf Wizard", afrArcaneInvestOracle, false)
		before := me.Hand.Size()
		if err := effectstest.P2Activate(t, g, me, id, 0, "{C}{C}{C}{C}{C}{U}"); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if natural < 10 {
			if got := me.Hand.Size() - before; got != 1 {
				t.Errorf("natural %d: drew %d, want 1", natural, got)
			}
			continue
		}
		pick := effectstest.LatestChooseCardsFor(g, me.ID)
		if pick == nil {
			t.Fatalf("natural %d: no look-at-three prompt", natural)
		}
	}
}

func TestAFRAberrantMindSorcererTable(t *testing.T) {
	for _, natural := range []int{9, 10, 20} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, natural)
		me := g.Seats[0]
		spell := effectstest.PushGraveyardCreature(g, me.ID, "Shock", "{R}")
		g.WithWriteLock(func() {
			for i := range me.Graveyard.Cards {
				if me.Graveyard.Cards[i].InstanceID == spell {
					me.Graveyard.Cards[i].TypeLine = "Instant"
				}
			}
		})
		afrCast(t, g, id, "Aberrant Mind Sorcerer", "Creature — Human Elf Shaman Sorcerer", afrAberrantMindOracle)
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.LatestPickTarget(g, me.ID) != nil {
			effectstest.PickTriggerTarget(t, g, me.ID, spell)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if natural >= 10 {
			if !effectstest.D5681InZone(g, spell, game.ZoneHand) {
				t.Errorf("natural %d: the instant was not returned to hand", natural)
			}
			continue
		}
		var ask *game.PendingChoice
		g.ReadSnapshot(func() { ask = effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) })
		if ask == nil {
			t.Fatalf("natural %d: no may prompt", natural)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		if !effectstest.D5681InZone(g, spell, game.ZoneLibrary) {
			t.Errorf("natural %d: the instant is not on the library", natural)
		}
	}
}

func TestAFRGoblinMorningstarTable(t *testing.T) {
	for _, tc := range []struct {
		natural  int
		attached bool
	}{{9, false}, {10, true}, {20, true}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		afrCast(t, g, id, "Goblin Morningstar", "Artifact — Equipment", afrGoblinMorningOracle)
		effectstest.PassPriorityAroundTable(t, g)
		goblin := effectstest.FindBattlefieldByName(g, "Goblin")
		if goblin == uuid.Nil {
			t.Fatalf("natural %d: no Goblin", tc.natural)
		}
		eq := effectstest.FindBattlefieldCardByID(g, id)
		if eq == nil || eq.IsAttachedTo(goblin) != tc.attached {
			t.Errorf("natural %d: attached=%v, want %v", tc.natural, eq != nil && eq.IsAttachedTo(goblin), tc.attached)
		}
		wantPower := 1
		if tc.attached {
			wantPower = 2
		}
		if got := effectstest.EffectivePower(t, g, goblin); got != wantPower {
			t.Errorf("natural %d: Goblin power %d, want %d", tc.natural, got, wantPower)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, goblin, "trample"); got != tc.attached {
			t.Errorf("natural %d: trample %v, want %v", tc.natural, got, tc.attached)
		}
	}
}

func TestAFREarthCultElementalTable(t *testing.T) {
	for _, tc := range []struct {
		natural, oppPerms, wantOpp int
		meGone                     bool
	}{{9, 1, 0, true}, {10, 1, 0, false}, {20, 2, 0, false}, {20, 3, 1, false}} {
		id := uuid.New()
		g := effectstest.D5681Table(t, id, tc.natural)
		for _, opp := range g.Seats[1:] {
			for i := 0; i < tc.oppPerms; i++ {
				effectstest.PushCatalogPermanent(g, opp.ID, "Forest", "Land", "", false)
			}
		}
		afrCast(t, g, id, "Earth-Cult Elemental", "Creature — Elemental", afrEarthCultOracle)
		effectstest.PassPriorityAroundTable(t, g)
		for i := 0; i < 12; i++ {
			var pick *game.PendingChoice
			g.ReadSnapshot(func() {
				for _, c := range g.PendingChoices {
					if c != nil && c.Kind == game.PendingChoiceChooseCards {
						pick = c
					}
				}
			})
			if pick == nil {
				break
			}
			n := pick.ChooseMin
			if err := g.ResolveChooseCards(pick.ID, pick.Chooser, pick.ChooseCards[:n]); err != nil {
				t.Fatalf("natural %d: choose: %v", tc.natural, err)
			}
		}
		g.RunStateChecksForTest()
		for _, opp := range g.Seats[1:] {
			left := 0
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.Controller == opp.ID {
						left++
					}
				}
			})
			if left != tc.wantOpp {
				t.Errorf("natural %d (%d perms): an opponent kept %d permanents, want %d", tc.natural, tc.oppPerms, left, tc.wantOpp)
			}
		}
		if gone := !effectstest.D5681InZone(g, id, game.ZoneBattlefield); gone != tc.meGone {
			t.Errorf("natural %d: Elemental gone=%v, want %v", tc.natural, gone, tc.meGone)
		}
	}
}
