package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fra_prepare_a_test.go — Reality Fracture slice fra-prepare-a: the
// preparation cards and the cards that make or cast prepared creatures.
// One test per behaviour, through the catalog and the real cast path.

const (
	fraPhytomedicOracle = "7737ccdc-49f2-463e-8ffd-43ab98d11f45"
	fraBlossomOracle    = "cd61c26b-8d3c-4730-92f2-0d94d961bafd"
	fraKonstrariOracle  = "bad2e3b2-ad47-4e17-b35e-0fcc13e82937"
	fraCrafterOracle    = "c8b3a070-408e-4a8f-8597-6476eeec0a5f"
	fraFateholdOracle   = "1063822f-47d3-42e9-8a21-f62b12609fe1"
	fraDivinerOracle    = "e7b78acd-5288-4528-97a8-a5b43de90bf5"
	fraHecklerOracle    = "d1af8018-35b5-4b6f-93c8-1ed6bd8c92b1"
	fraBloodlineOracle  = "c8e9a7e1-28ae-40e3-8715-1bb0aa2057c7"
	fraCultivatorOracle = "db655c5a-06f6-42bc-8180-d9ca1c1d939a"
	fraCodieOracle      = "a17e414c-8de0-47fd-bcbc-540c7fec7642"
	fraArenaOracle      = "2cc075da-f79b-4050-84eb-2c1b9a30260c"
	fraCourseworkOracle = "666b34b5-a0c2-43e3-befd-7c3979313682"
)

func fraCard(owner uuid.UUID, oracle string, keywords []string, front, spell game.Face) game.Card {
	return preparationCard(owner, oracle, keywords, front, spell)
}

func fraPhytomedic(owner uuid.UUID) game.Card {
	return fraCard(owner, fraPhytomedicOracle, nil,
		game.Face{Name: "Emergency Phytomedic", TypeLine: "Creature — Dryad Cleric", ManaCost: "{G/W}", Colors: []string{"G", "W"}, Power: 1, Toughness: 1},
		game.Face{Name: "Seed Suture", TypeLine: "Sorcery", ManaCost: "{G/W}", Colors: []string{"G", "W"}})
}

func fraBlossom(owner uuid.UUID) game.Card {
	return fraCard(owner, fraBlossomOracle, []string{"flying", "vigilance"},
		game.Face{Name: "Blossom-Blessed Angel", TypeLine: "Creature — Angel Cleric", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 2, Toughness: 4},
		game.Face{Name: "Seed Suture", TypeLine: "Sorcery", ManaCost: "{G/W}", Colors: []string{"G", "W"}})
}

func fraKonstrari(owner uuid.UUID) game.Card {
	return fraCard(owner, fraKonstrariOracle, nil,
		game.Face{Name: "Konstrari Improviser", TypeLine: "Creature — Human Artificer", ManaCost: "{1}{R/G}", Colors: []string{"R", "G"}, Power: 2, Toughness: 2},
		game.Face{Name: "Soul Tether", TypeLine: "Sorcery", ManaCost: "{2}{R/G}", Colors: []string{"R", "G"}})
}

func fraCrafter(owner uuid.UUID) game.Card {
	return fraCard(owner, fraCrafterOracle, nil,
		game.Face{Name: "Heartwood Crafter", TypeLine: "Creature — Elf Artificer", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
		game.Face{Name: "Soul Tether", TypeLine: "Sorcery", ManaCost: "{2}{R/G}", Colors: []string{"R", "G"}})
}

func fraFatehold(owner uuid.UUID) game.Card {
	return fraCard(owner, fraFateholdOracle, []string{"flying"},
		game.Face{Name: "Fatehold Chronologist", TypeLine: "Creature — Bird Wizard", ManaCost: "{1}{W/U}", Colors: []string{"W", "U"}, Power: 1, Toughness: 2},
		game.Face{Name: "Peer Review", TypeLine: "Sorcery", ManaCost: "{2}{W/U}", Colors: []string{"W", "U"}})
}

func fraDiviner(owner uuid.UUID) game.Card {
	return fraCard(owner, fraDivinerOracle, nil,
		game.Face{Name: "Diviner of Victory", TypeLine: "Creature — Dwarf Wizard", ManaCost: "{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
		game.Face{Name: "Unwind History", TypeLine: "Sorcery", ManaCost: "{1}{U}", Colors: []string{"U"}})
}

func fraHeckler(owner uuid.UUID) game.Card {
	return fraCard(owner, fraHecklerOracle, nil,
		game.Face{Name: "Hallway Heckler", TypeLine: "Creature — Elemental Sorcerer", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 3},
		game.Face{Name: "Vicious Verse", TypeLine: "Sorcery", ManaCost: "{B/R}", Colors: []string{"B", "R"}})
}

func fraBloodline(owner uuid.UUID) game.Card {
	return fraCard(owner, fraBloodlineOracle, nil,
		game.Face{Name: "Bloodline Recollector", TypeLine: "Creature — Vampire Warlock", ManaCost: "{1}{B}", Colors: []string{"B"}, Power: 2, Toughness: 2},
		game.Face{Name: "Ancestral Craving", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}})
}

func fraCultivator(owner uuid.UUID) game.Card {
	return fraCard(owner, fraCultivatorOracle, []string{"deathtouch"},
		game.Face{Name: "Carnivorous Cultivator", TypeLine: "Creature — Elf Warlock", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 3},
		game.Face{Name: "Enroot", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
}

// fraEnter walks to the active seat's main phase and casts the card,
// returning the prepared permanent's id and the copy of its spell.
func fraEnter(t *testing.T, g *game.Game, me *game.Player, c game.Card, spell string) (uuid.UUID, game.Card) {
	t.Helper()
	advanceTo(t, g, game.StepPrecombatMain)
	castFromHandAndResolve(t, g, me, c)
	cp, ok := prepareCopyOf(g, spell)
	if !ok {
		t.Fatalf("%s: no copy of %s in exile", c.Name, spell)
	}
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatalf("%s did not enter prepared", c.Name)
	}
	return c.InstanceID, cp
}

func fraCastCopy(t *testing.T, g *game.Game, me *game.Player, cp game.Card, targets ...game.TargetRef) error {
	t.Helper()
	return g.CastSpell(me.ID, cp.InstanceID, game.CastSpellParams{FromZone: string(game.ZoneExile), Targets: targets})
}

func fraCardTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

func fraPlayerTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

func fraHasCounters(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				n = c.Counters[game.CounterPlusOne]
			}
		}
	})
	return n
}

// Both Seed Suture cards enter prepared and the copy puts a +1/+1
// counter on the target and gains 1 life; the spell needs a creature.
func TestSeedSutureCardsEnterPreparedAndCastTheCopy(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func(uuid.UUID) game.Card
	}{
		{"Emergency Phytomedic", fraPhytomedic},
		{"Blossom-Blessed Angel", fraBlossom},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
			_, cp := fraEnter(t, g, me, tc.make(me.ID), "Seed Suture")
			life := me.Life
			if err := fraCastCopy(t, g, me, cp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("Seed Suture on a player: err = %v, want ErrIllegalTarget", err)
			}
			if err := fraCastCopy(t, g, me, cp, fraCardTarget(bear)); err != nil {
				t.Fatalf("cast Seed Suture: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := fraHasCounters(t, g, bear); got != 1 {
				t.Errorf("bear has %d +1/+1 counters, want 1", got)
			}
			if me.Life != life+1 {
				t.Errorf("life = %d, want %d", me.Life, life+1)
			}
		})
	}
}

func TestBlossomBlessedAngelHasFlyingAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _ := fraEnter(t, g, me, fraBlossom(me.ID), "Seed Suture")
	abilities := effectiveAbilities(t, g, id)
	for _, kw := range []string{"flying", "vigilance"} {
		if !containsString(abilities, kw) {
			t.Errorf("abilities %v lack %q", abilities, kw)
		}
	}
}

// Soul Tether (both cards) makes a red and green Heartwood artifact
// that taps for {R} or {G}.
func TestSoulTetherMakesAHeartwoodToken(t *testing.T) {
	for _, tc := range []struct {
		name string
		make func(uuid.UUID) game.Card
	}{
		{"Konstrari Improviser", fraKonstrari},
		{"Heartwood Crafter", fraCrafter},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			_, cp := fraEnter(t, g, me, tc.make(me.ID), "Soul Tether")
			if err := fraCastCopy(t, g, me, cp); err != nil {
				t.Fatalf("cast Soul Tether: %v", err)
			}
			passPriorityAroundTable(t, g)
			id := findBattlefieldByName(g, "Heartwood")
			if id == uuid.Nil {
				t.Fatal("no Heartwood token on the battlefield")
			}
			var card game.Card
			g.ReadSnapshot(func() {
				for _, c := range g.Battlefield.Cards {
					if c.InstanceID == id {
						card = c
					}
				}
			})
			if !card.IsToken() || !card.IsArtifact() {
				t.Errorf("Heartwood: token=%v artifact=%v, want both", card.IsToken(), card.IsArtifact())
			}
			if len(card.Colors) != 2 {
				t.Errorf("Heartwood colours = %v, want red and green", card.Colors)
			}
			abs := game.ManaAbilitiesForCard(card)
			if len(abs) != 1 || abs[0].Produced != "{R|G}" {
				t.Errorf("Heartwood mana abilities = %+v, want one {R|G}", abs)
			}
		})
	}
}

// Heartwood Crafter taps for {C} but only toward activated abilities.
func TestHeartwoodCrafterManaIsRestricted(t *testing.T) {
	spec, ok := Lookup(fraCrafterOracle)
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Heartwood Crafter spec missing its mana ability: %v %d", ok, len(spec.ManaAbilities))
	}
	ab := spec.ManaAbilities[0]
	if ab.Produced != "{C}" || !ab.Cost.Tap {
		t.Errorf("mana ability = %+v, want {T}: Add {C}", ab)
	}
	if len(ab.Restrictions) != 1 || ab.Restrictions[0] != ManaRestrictActivate {
		t.Errorf("restrictions = %v, want only activated-ability costs", ab.Restrictions)
	}
	if spec.Completeness != CompletenessCaveats || len(spec.Caveats) != 1 {
		t.Errorf("Heartwood Crafter must declare its one caveat, got %v %v", spec.Completeness, spec.Caveats)
	}
}

// Peer Review makes a 2/2 colourless Wizard Soldier named Cadet, then
// the surveil prompt opens.
func TestPeerReviewMakesACadetThenSurveils(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "Fodder", "Keeper")
	_, cp := fraEnter(t, g, me, fraFatehold(me.ID), "Peer Review")
	if err := fraCastCopy(t, g, me, cp); err != nil {
		t.Fatalf("cast Peer Review: %v", err)
	}
	passPriorityAroundTable(t, g)
	id := findBattlefieldByName(g, "Cadet")
	if id == uuid.Nil {
		t.Fatal("no Cadet token")
	}
	if effectivePower(t, g, id) != 2 {
		t.Errorf("Cadet power = %d, want 2", effectivePower(t, g, id))
	}
	types := effectiveSubtypes(t, g, id)
	if !containsString(types, "Wizard") || !containsString(types, "Soldier") {
		t.Errorf("Cadet subtypes = %v, want Wizard Soldier", types)
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt after the token")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Fodder" {
		t.Errorf("graveyard = %v, want [Fodder]", got)
	}
}

// Unwind History returns an opponent's creature of mana value 3 or less
// and then surveils; Diviner of Victory grows when that surveil
// finishes. A creature of mana value 4, and your own, are illegal.
func TestUnwindHistoryBouncesSmallCreatureAndDivinerGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "Fodder", "Keeper")
	smallID := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Small", TypeLine: testCreatureTypeLine, ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	bigID := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: testCreatureTypeLine, ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	mine := pushVanillaCreature(g, me.ID, "Mine", 2, 2)
	diviner, cp := fraEnter(t, g, me, fraDiviner(me.ID), "Unwind History")

	for _, bad := range []uuid.UUID{bigID, mine} {
		if err := fraCastCopy(t, g, me, cp, fraCardTarget(bad)); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("Unwind History on an illegal creature: err = %v, want ErrIllegalTarget", err)
		}
	}
	if err := fraCastCopy(t, g, me, cp, fraCardTarget(smallID)); err != nil {
		t.Fatalf("cast Unwind History: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(smallID) {
		t.Fatal("the small creature is still on the battlefield")
	}
	if opp.Hand.Size() == 0 {
		t.Error("the small creature did not reach its owner's hand")
	}
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt after the bounce")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := effectivePower(t, g, diviner); got != 2 {
		t.Errorf("Diviner of Victory power = %d after a surveil, want 2", got)
	}
}

// Hallway Heckler: Vicious Verse pings an opponent for 1 and refuses
// you; the creature rummages by tapping and discarding.
func TestViciousVerseDamagesOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, cp := fraEnter(t, g, me, fraHeckler(me.ID), "Vicious Verse")
	if err := fraCastCopy(t, g, me, cp, fraPlayerTarget(me.ID)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Vicious Verse on yourself: err = %v, want ErrIllegalTarget", err)
	}
	life := opp.Life
	if err := fraCastCopy(t, g, me, cp, fraPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast Vicious Verse: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-1)
	}
}

func TestHallwayHecklerRummages(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	id := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hallway Heckler", TypeLine: "Creature — Elemental Sorcerer",
		OracleID: fraHecklerOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	pitch := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: pitch, Name: "Pitch", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{pitch}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (discard one, draw one)", me.Hand.Size(), hand)
	}
	if me.Hand.Contains(pitch) {
		t.Error("the discarded card is still in hand")
	}
}

// Bloodline Recollector becomes prepared at an end step only after
// three creatures have died that turn; Ancestral Craving draws and
// drains three.
func TestBloodlineRecollectorPreparesAfterThreeDeaths(t *testing.T) {
	for _, tc := range []struct {
		deaths int
		want   bool
	}{{2, false}, {3, true}} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceTo(t, g, game.StepPrecombatMain)
		castFromHandAndResolve(t, g, me, fraBloodline(me.ID))
		var id uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.OracleID == fraBloodlineOracle {
					id = c.InstanceID
				}
			}
		})
		if g.IsPreparedForEffect(id) {
			t.Fatal("Bloodline Recollector entered prepared; it should not")
		}
		g.WithWriteLock(func() { g.TurnTally.CreaturesDied = tc.deaths })
		advanceToEndStep(t, g)
		passPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(id); got != tc.want {
			t.Errorf("%d creatures died: prepared = %v, want %v", tc.deaths, got, tc.want)
		}
	}
}

func TestAncestralCravingDrawsThreeAndLosesThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	castFromHandAndResolve(t, g, me, fraBloodline(me.ID))
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == fraBloodlineOracle {
				id = c.InstanceID
			}
		}
	})
	if _, err := g.BecomePreparedForEffect(id); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	cp, ok := prepareCopyOf(g, "Ancestral Craving")
	if !ok {
		t.Fatal("no copy of Ancestral Craving")
	}
	hand, life := opp.Hand.Size(), opp.Life
	if err := fraCastCopy(t, g, me, cp, fraPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast Ancestral Craving: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 || opp.Life != life-3 {
		t.Errorf("target hand %d->%d life %d->%d, want +3 cards and -3 life", hand, opp.Hand.Size(), life, opp.Life)
	}
}

// Carnivorous Cultivator: Enroot puts a land from the library into the
// graveyard; connecting returns a land card from the graveyard.
func TestEnrootPutsALandIntoTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	_, cp := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	if err := fraCastCopy(t, g, me, cp); err != nil {
		t.Fatalf("cast Enroot: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := searchChoiceFor(g, me.ID); c != nil {
		if searchOptionNamed(g, c, "Bear") != uuid.Nil {
			t.Error("Enroot offered a nonland card")
		}
		answerSearchNamed(t, g, me.ID, "Forest")
	}
	if got := graveyardNames(me); len(got) != 1 || got[0] != "Forest" {
		t.Errorf("graveyard = %v, want [Forest]", got)
	}
}

func TestCarnivorousCultivatorReturnsALandWhenItConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	pushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	id, _ := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	dealCombatDamageToPlayer(g, id, opp.ID, 2)
	fraSettle(t, g, land)
	if !me.Hand.Contains(land) {
		t.Error("the land card did not return to hand")
	}
}

func TestCarnivorousCultivatorHasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id, _ := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	if !containsString(effectiveAbilities(t, g, id), "deathtouch") {
		t.Error("no deathtouch")
	}
}

// Codie copies a prepared spell cast from exile, and only that.
func TestCodieCopiesAPreparedSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	pushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", fraCodieOracle, false)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	castFromHandAndResolve(t, g, me, fraPhytomedic(me.ID))
	cp, ok := prepareCopyOf(g, "Seed Suture")
	if !ok {
		t.Fatal("no Seed Suture copy")
	}
	life := me.Life
	if err := fraCastCopy(t, g, me, cp, fraCardTarget(bear)); err != nil {
		t.Fatalf("cast Seed Suture: %v", err)
	}
	// Codie's trigger asks about new targets for the copy; keep the bear.
	fraSettle(t, g, bear)
	if got := fraHasCounters(t, g, bear); got != 2 {
		t.Errorf("bear has %d counters, want 2 (the spell and Codie's copy)", got)
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

func TestCodieIgnoresAnOrdinarySpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	pushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", fraCodieOracle, false)
	castFromHandAndResolve(t, g, me, fraBloodline(me.ID)) // not prepared on entry
	fraSettle(t, g, uuid.Nil)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == fraBloodlineOracle {
				n++
			}
		}
	})
	if n != 1 {
		t.Errorf("%d Bloodline Recollectors on the battlefield, want 1 — Codie copied an ordinary spell", n)
	}
}

// Codie's activated ability prepares every creature that can be.
func TestCodieMakesEveryCreatureYouControlPrepared(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	codie := pushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", fraCodieOracle, false)
	bloodline := fraBloodline(me.ID)
	castFromHandAndResolve(t, g, me, bloodline)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}{B}{R}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, codie, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Codie: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(bloodline.InstanceID) {
		t.Error("Bloodline Recollector did not become prepared")
	}
	if g.IsPreparedForEffect(bear) || g.IsPreparedForEffect(codie) {
		t.Error("a creature with no prepare spell became prepared")
	}
	if _, ok := prepareCopyOf(g, "Ancestral Craving"); !ok {
		t.Error("no copy of Ancestral Craving in exile")
	}
}

// Hexhaven Dueling Arena: the {2} ability needs a creature that attacked
// this turn and sorcery timing; the {4} ability takes any creature.
func TestHexhavenDuelingArenaPreparesAnAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	arena := pushCatalogPermanent(g, me.ID, "Hexhaven Dueling Arena", "Land", fraArenaOracle, false)
	bloodline := fraBloodline(me.ID)
	castFromHandAndResolve(t, g, me, bloodline)
	id := bloodline.InstanceID

	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{fraCardTarget(id)}})
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("{2} ability on a creature that did not attack: err = %v, want ErrIllegalTarget", err)
	}

	// A creature that attacked this turn is legal at sorcery timing.
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	declareAttack(t, g, opp.ID, id)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{fraCardTarget(id)}}); err != nil {
		t.Fatalf("{2} ability on an attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(id) {
		t.Error("the attacker did not become prepared")
	}
}

func TestHexhavenDuelingArenaFourManaPreparesAnyCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceTo(t, g, game.StepPrecombatMain)
	arena := pushCatalogPermanent(g, me.ID, "Hexhaven Dueling Arena", "Land", fraArenaOracle, false)
	bloodline := fraBloodline(me.ID)
	castFromHandAndResolve(t, g, me, bloodline)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, arena, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{fraCardTarget(bloodline.InstanceID)}}); err != nil {
		t.Fatalf("{4} ability: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(bloodline.InstanceID) {
		t.Error("the creature did not become prepared")
	}
}

// Infinite Coursework taps the enchanted creature, unprepares it (the
// copy leaves exile), strips its abilities and keeps it from untapping.
func TestInfiniteCourseworkTapsUnpreparesAndSilences(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	angel := fraBlossom(opp.ID)
	pushBattlefieldCardWithTimestamp(g, angel)
	if _, err := g.BecomePreparedForEffect(angel.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if _, ok := prepareCopyOf(g, "Seed Suture"); !ok {
		t.Fatal("no Seed Suture copy before the Aura")
	}
	castCatalogSpell(t, g, "Infinite Coursework", "Enchantment — Aura", fraCourseworkOracle,
		[]game.TargetRef{fraCardTarget(angel.InstanceID)})
	passPriorityAroundTable(t, g)
	_ = me

	if g.IsPreparedForEffect(angel.InstanceID) {
		t.Error("the enchanted creature is still prepared")
	}
	if _, ok := prepareCopyOf(g, "Seed Suture"); ok {
		t.Error("the copy of its prepare spell is still in exile")
	}
	tapped := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == angel.InstanceID {
				tapped = c.Tapped
			}
		}
	})
	if !tapped {
		t.Error("the enchanted creature is not tapped")
	}
	if abs := effectiveAbilities(t, g, angel.InstanceID); containsString(abs, "flying") || containsString(abs, "vigilance") {
		t.Errorf("abilities = %v, want none", abs)
	}
}

// fraSettle passes priority until the stack is empty, answering any
// pick_target prompt with `pick`.
func fraSettle(t *testing.T, g *game.Game, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 32; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoicePickTarget {
				if err := g.ResolvePickTarget(c.ID, c.Chooser, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}
