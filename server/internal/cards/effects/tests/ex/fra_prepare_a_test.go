package testsex

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func fraPhytomedic(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraPhytomedicOracle, nil,
		game.Face{Name: "Emergency Phytomedic", TypeLine: "Creature — Dryad Cleric", ManaCost: "{G/W}", Colors: []string{"G", "W"}, Power: 1, Toughness: 1},
		game.Face{Name: "Seed Suture", TypeLine: "Sorcery", ManaCost: "{G/W}", Colors: []string{"G", "W"}})
}

func fraBlossom(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraBlossomOracle, []string{"flying", "vigilance"},
		game.Face{Name: "Blossom-Blessed Angel", TypeLine: "Creature — Angel Cleric", ManaCost: "{3}{W}", Colors: []string{"W"}, Power: 2, Toughness: 4},
		game.Face{Name: "Seed Suture", TypeLine: "Sorcery", ManaCost: "{G/W}", Colors: []string{"G", "W"}})
}

func fraKonstrari(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraKonstrariOracle, nil,
		game.Face{Name: "Konstrari Improviser", TypeLine: "Creature — Human Artificer", ManaCost: "{1}{R/G}", Colors: []string{"R", "G"}, Power: 2, Toughness: 2},
		game.Face{Name: "Soul Tether", TypeLine: "Sorcery", ManaCost: "{2}{R/G}", Colors: []string{"R", "G"}})
}

func fraCrafter(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraCrafterOracle, nil,
		game.Face{Name: "Heartwood Crafter", TypeLine: "Creature — Elf Artificer", ManaCost: "{G}", Colors: []string{"G"}, Power: 1, Toughness: 1},
		game.Face{Name: "Soul Tether", TypeLine: "Sorcery", ManaCost: "{2}{R/G}", Colors: []string{"R", "G"}})
}

func fraFatehold(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraFateholdOracle, []string{"flying"},
		game.Face{Name: "Fatehold Chronologist", TypeLine: "Creature — Bird Wizard", ManaCost: "{1}{W/U}", Colors: []string{"W", "U"}, Power: 1, Toughness: 2},
		game.Face{Name: "Peer Review", TypeLine: "Sorcery", ManaCost: "{2}{W/U}", Colors: []string{"W", "U"}})
}

func fraDiviner(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraDivinerOracle, nil,
		game.Face{Name: "Diviner of Victory", TypeLine: "Creature — Dwarf Wizard", ManaCost: "{U}", Colors: []string{"U"}, Power: 1, Toughness: 1},
		game.Face{Name: "Unwind History", TypeLine: "Sorcery", ManaCost: "{1}{U}", Colors: []string{"U"}})
}

func fraHeckler(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraHecklerOracle, nil,
		game.Face{Name: "Hallway Heckler", TypeLine: "Creature — Elemental Sorcerer", ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 3},
		game.Face{Name: "Vicious Verse", TypeLine: "Sorcery", ManaCost: "{B/R}", Colors: []string{"B", "R"}})
}

func fraCultivator(owner uuid.UUID) game.Card {
	return effectstest.FraCard(owner, effectstest.FraCultivatorOracle, []string{"deathtouch"},
		game.Face{Name: "Carnivorous Cultivator", TypeLine: "Creature — Elf Warlock", ManaCost: "{1}{G}", Colors: []string{"G"}, Power: 2, Toughness: 3},
		game.Face{Name: "Enroot", TypeLine: "Sorcery", ManaCost: "{G}", Colors: []string{"G"}})
}

// fraEnter walks to the active seat's main phase and casts the card,
// returning the prepared permanent's id and the copy of its spell.
func fraEnter(t *testing.T, g *game.Game, me *game.Player, c game.Card, spell string) (uuid.UUID, game.Card) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastFromHandAndResolve(t, g, me, c)
	cp, ok := effectstest.PrepareCopyOf(g, spell)
	if !ok {
		t.Fatalf("%s: no copy of %s in exile", c.Name, spell)
	}
	if !g.IsPreparedForEffect(c.InstanceID) {
		t.Fatalf("%s did not enter prepared", c.Name)
	}
	return c.InstanceID, cp
}

func fraCardTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id}
}

func fraPlayerTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
			_, cp := fraEnter(t, g, me, tc.make(me.ID), "Seed Suture")
			life := me.Life
			if err := effectstest.FraCastCopy(t, g, me, cp, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); !errors.Is(err, game.ErrIllegalTarget) {
				t.Fatalf("Seed Suture on a player: err = %v, want ErrIllegalTarget", err)
			}
			if err := effectstest.FraCastCopy(t, g, me, cp, fraCardTarget(bear)); err != nil {
				t.Fatalf("cast Seed Suture: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.FraHasCounters(t, g, bear); got != 1 {
				t.Errorf("bear has %d +1/+1 counters, want 1", got)
			}
			if me.Life != life+1 {
				t.Errorf("life = %d, want %d", me.Life, life+1)
			}
		})
	}
}

func TestBlossomBlessedAngelHasFlyingAndVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id, _ := fraEnter(t, g, me, fraBlossom(me.ID), "Seed Suture")
	abilities := effectstest.EffectiveAbilities(t, g, id)
	for _, kw := range []string{"flying", "vigilance"} {
		if !effectstest.ContainsString(abilities, kw) {
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
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			_, cp := fraEnter(t, g, me, tc.make(me.ID), "Soul Tether")
			if err := effectstest.FraCastCopy(t, g, me, cp); err != nil {
				t.Fatalf("cast Soul Tether: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			id := effectstest.FindBattlefieldByName(g, "Heartwood")
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

// Heartwood Crafter taps for {C} that can't be spent to cast spells
// from your hand (#2811), and ships whole.
func TestHeartwoodCrafterManaIsRestricted(t *testing.T) {
	spec, ok := effects.Lookup(effectstest.FraCrafterOracle)
	if !ok || len(spec.ManaAbilities) != 1 {
		t.Fatalf("Heartwood Crafter spec missing its mana ability: %v %d", ok, len(spec.ManaAbilities))
	}
	ab := spec.ManaAbilities[0]
	if ab.Produced != "{C}" || !ab.Cost.Tap {
		t.Errorf("mana ability = %+v, want {T}: Add {C}", ab)
	}
	if len(ab.Restrictions) != 1 || ab.Restrictions[0] != effects.ManaRestrictNotFromHand {
		t.Errorf("restrictions = %v, want only the not-from-hand restriction", ab.Restrictions)
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Heartwood Crafter ships full, got %v %v", spec.Completeness, spec.Caveats)
	}
}

// Peer Review makes a 2/2 colourless Wizard Soldier named Cadet, then
// the surveil prompt opens.
func TestPeerReviewMakesACadetThenSurveils(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "Fodder", "Keeper")
	_, cp := fraEnter(t, g, me, fraFatehold(me.ID), "Peer Review")
	if err := effectstest.FraCastCopy(t, g, me, cp); err != nil {
		t.Fatalf("cast Peer Review: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	id := effectstest.FindBattlefieldByName(g, "Cadet")
	if id == uuid.Nil {
		t.Fatal("no Cadet token")
	}
	if effectstest.EffectivePower(t, g, id) != 2 {
		t.Errorf("Cadet power = %d, want 2", effectstest.EffectivePower(t, g, id))
	}
	types := effectstest.EffectiveSubtypes(t, g, id)
	if !effectstest.ContainsString(types, "Wizard") || !effectstest.ContainsString(types, "Soldier") {
		t.Errorf("Cadet subtypes = %v, want Wizard Soldier", types)
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt after the token")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if got := effectstest.GraveyardNames(me); len(got) != 1 || got[0] != "Fodder" {
		t.Errorf("graveyard = %v, want [Fodder]", got)
	}
}

// Unwind History returns an opponent's creature of mana value 3 or less
// and then surveils; Diviner of Victory grows when that surveil
// finishes. A creature of mana value 4, and your own, are illegal.
func TestUnwindHistoryBouncesSmallCreatureAndDivinerGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "Fodder", "Keeper")
	smallID := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Small", TypeLine: effectstest.TestCreatureTypeLine, ManaCost: "{2}",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID,
	})
	bigID := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Big", TypeLine: effectstest.TestCreatureTypeLine, ManaCost: "{3}{G}",
		Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID,
	})
	mine := effectstest.PushVanillaCreature(g, me.ID, "Mine", 2, 2)
	diviner, cp := fraEnter(t, g, me, fraDiviner(me.ID), "Unwind History")

	for _, bad := range []uuid.UUID{bigID, mine} {
		if err := effectstest.FraCastCopy(t, g, me, cp, fraCardTarget(bad)); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("Unwind History on an illegal creature: err = %v, want ErrIllegalTarget", err)
		}
	}
	if err := effectstest.FraCastCopy(t, g, me, cp, fraCardTarget(smallID)); err != nil {
		t.Fatalf("cast Unwind History: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(smallID) {
		t.Fatal("the small creature is still on the battlefield")
	}
	if opp.Hand.Size() == 0 {
		t.Error("the small creature did not reach its owner's hand")
	}
	c := effectstest.SurveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no surveil prompt after the bounce")
	}
	if err := g.ResolveSurveil(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, diviner); got != 2 {
		t.Errorf("Diviner of Victory power = %d after a surveil, want 2", got)
	}
}

// Hallway Heckler: Vicious Verse pings an opponent for 1 and refuses
// you; the creature rummages by tapping and discarding.
func TestViciousVerseDamagesOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	_, cp := fraEnter(t, g, me, fraHeckler(me.ID), "Vicious Verse")
	if err := effectstest.FraCastCopy(t, g, me, cp, fraPlayerTarget(me.ID)); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("Vicious Verse on yourself: err = %v, want ErrIllegalTarget", err)
	}
	life := opp.Life
	if err := effectstest.FraCastCopy(t, g, me, cp, fraPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast Vicious Verse: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life = %d, want %d", opp.Life, life-1)
	}
}

func TestHallwayHecklerRummages(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	id := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Hallway Heckler", TypeLine: "Creature — Elemental Sorcerer",
		OracleID: effectstest.FraHecklerOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
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
	effectstest.PassPriorityAroundTable(t, g)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		effectstest.CastFromHandAndResolve(t, g, me, effectstest.FraBloodline(me.ID))
		var id uuid.UUID
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.OracleID == effectstest.FraBloodlineOracle {
					id = c.InstanceID
				}
			}
		})
		if g.IsPreparedForEffect(id) {
			t.Fatal("Bloodline Recollector entered prepared; it should not")
		}
		g.WithWriteLock(func() { g.TurnTally.CreaturesDied = tc.deaths })
		effectstest.AdvanceToEndStep(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if got := g.IsPreparedForEffect(id); got != tc.want {
			t.Errorf("%d creatures died: prepared = %v, want %v", tc.deaths, got, tc.want)
		}
	}
}

func TestAncestralCravingDrawsThreeAndLosesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.CastFromHandAndResolve(t, g, me, effectstest.FraBloodline(me.ID))
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == effectstest.FraBloodlineOracle {
				id = c.InstanceID
			}
		}
	})
	if _, err := g.BecomePreparedForEffect(id); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	cp, ok := effectstest.PrepareCopyOf(g, "Ancestral Craving")
	if !ok {
		t.Fatal("no copy of Ancestral Craving")
	}
	hand, life := opp.Hand.Size(), opp.Life
	if err := effectstest.FraCastCopy(t, g, me, cp, fraPlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast Ancestral Craving: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand+3 || opp.Life != life-3 {
		t.Errorf("target hand %d->%d life %d->%d, want +3 cards and -3 life", hand, opp.Hand.Size(), life, opp.Life)
	}
}

// Carnivorous Cultivator: Enroot puts a land from the library into the
// graveyard; connecting returns a land card from the graveyard.
func TestEnrootPutsALandIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: me.ID, Controller: me.ID})
	effectstest.PushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	_, cp := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	if err := effectstest.FraCastCopy(t, g, me, cp); err != nil {
		t.Fatalf("cast Enroot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.SearchChoiceFor(g, me.ID); c != nil {
		if effectstest.SearchOptionNamed(g, c, "Bear") != uuid.Nil {
			t.Error("Enroot offered a nonland card")
		}
		effectstest.AnswerSearchNamed(t, g, me.ID, "Forest")
	}
	if got := effectstest.GraveyardNames(me); len(got) != 1 || got[0] != "Forest" {
		t.Errorf("graveyard = %v, want [Forest]", got)
	}
}

func TestCarnivorousCultivatorReturnsALandWhenItConnects(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushCatalogGraveyardCard(me, "Forest", "Basic Land — Forest", "", 0, 0)
	effectstest.PushCatalogGraveyardCard(me, "Bear", "Creature — Bear", "", 2, 2)
	id, _ := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	effectstest.DealCombatDamageToPlayer(g, id, opp.ID, 2)
	fraSettle(t, g, land)
	if !me.Hand.Contains(land) {
		t.Error("the land card did not return to hand")
	}
}

func TestCarnivorousCultivatorHasDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id, _ := fraEnter(t, g, me, fraCultivator(me.ID), "Enroot")
	if !effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, id), "deathtouch") {
		t.Error("no deathtouch")
	}
}

// Codie copies a prepared spell cast from exile, and only that.
func TestCodieCopiesAPreparedSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.PushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", effectstest.FraCodieOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.CastFromHandAndResolve(t, g, me, fraPhytomedic(me.ID))
	cp, ok := effectstest.PrepareCopyOf(g, "Seed Suture")
	if !ok {
		t.Fatal("no Seed Suture copy")
	}
	life := me.Life
	if err := effectstest.FraCastCopy(t, g, me, cp, fraCardTarget(bear)); err != nil {
		t.Fatalf("cast Seed Suture: %v", err)
	}
	// Codie's trigger asks about new targets for the copy; keep the bear.
	fraSettle(t, g, bear)
	if got := effectstest.FraHasCounters(t, g, bear); got != 2 {
		t.Errorf("bear has %d counters, want 2 (the spell and Codie's copy)", got)
	}
	if me.Life != life+2 {
		t.Errorf("life = %d, want %d", me.Life, life+2)
	}
}

func TestCodieIgnoresAnOrdinarySpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	effectstest.PushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", effectstest.FraCodieOracle, false)
	effectstest.CastFromHandAndResolve(t, g, me, effectstest.FraBloodline(me.ID)) // not prepared on entry
	fraSettle(t, g, uuid.Nil)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.OracleID == effectstest.FraBloodlineOracle {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	codie := effectstest.PushCatalogPermanent(g, me.ID, "Codie, Ravenous Codex", "Legendary Artifact Creature — Book Construct", effectstest.FraCodieOracle, false)
	bloodline := effectstest.FraBloodline(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, bloodline)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{U}{B}{R}{G}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, codie, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate Codie: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(bloodline.InstanceID) {
		t.Error("Bloodline Recollector did not become prepared")
	}
	if g.IsPreparedForEffect(bear) || g.IsPreparedForEffect(codie) {
		t.Error("a creature with no prepare spell became prepared")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Ancestral Craving"); !ok {
		t.Error("no copy of Ancestral Craving in exile")
	}
}

// Hexhaven Dueling Arena: the {2} ability needs a creature that attacked
// this turn and sorcery timing; the {4} ability takes any creature.
func TestHexhavenDuelingArenaPreparesAnAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	arena := effectstest.PushCatalogPermanent(g, me.ID, "Hexhaven Dueling Arena", "Land", effectstest.FraArenaOracle, false)
	bloodline := effectstest.FraBloodline(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, bloodline)
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
	effectstest.DeclareAttack(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, arena, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{fraCardTarget(id)}}); err != nil {
		t.Fatalf("{2} ability on an attacker: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(id) {
		t.Error("the attacker did not become prepared")
	}
}

func TestHexhavenDuelingArenaFourManaPreparesAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	arena := effectstest.PushCatalogPermanent(g, me.ID, "Hexhaven Dueling Arena", "Land", effectstest.FraArenaOracle, false)
	bloodline := effectstest.FraBloodline(me.ID)
	effectstest.CastFromHandAndResolve(t, g, me, bloodline)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{C}{C}"); err != nil {
		t.Fatalf("mana: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, arena, 1, game.ActivateAbilityParams{Targets: []game.TargetRef{fraCardTarget(bloodline.InstanceID)}}); err != nil {
		t.Fatalf("{4} ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(bloodline.InstanceID) {
		t.Error("the creature did not become prepared")
	}
}

// Infinite Coursework taps the enchanted creature, unprepares it (the
// copy leaves exile), strips its abilities and keeps it from untapping.
func TestInfiniteCourseworkTapsUnpreparesAndSilences(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	opp := g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	angel := fraBlossom(opp.ID)
	effectstest.PushBattlefieldCardWithTimestamp(g, angel)
	if _, err := g.BecomePreparedForEffect(angel.InstanceID); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Seed Suture"); !ok {
		t.Fatal("no Seed Suture copy before the Aura")
	}
	effectstest.CastCatalogSpell(t, g, "Infinite Coursework", "Enchantment — Aura", effectstest.FraCourseworkOracle,
		[]game.TargetRef{fraCardTarget(angel.InstanceID)})
	effectstest.PassPriorityAroundTable(t, g)
	_ = me

	if g.IsPreparedForEffect(angel.InstanceID) {
		t.Error("the enchanted creature is still prepared")
	}
	if _, ok := effectstest.PrepareCopyOf(g, "Seed Suture"); ok {
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
	if abs := effectstest.EffectiveAbilities(t, g, angel.InstanceID); effectstest.ContainsString(abs, "flying") || effectstest.ContainsString(abs, "vigilance") {
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
		if effectstest.StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}
