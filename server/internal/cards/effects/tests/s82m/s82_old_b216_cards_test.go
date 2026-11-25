package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b216_cards_test.go — S82 OLD batch 216 (#5531, Rivals of Ixalan:
// card draw and library). Each test plays the real catalog entry.

const (
	oldB216VraskasScorn = "020c26ba-ed46-47c3-925c-b9f32500f503"
	oldB216ForeEmpire   = "113c3024-c3bd-4c3a-8c08-3828646d9ca2"
	oldB216BloodSun     = "1534d90f-fc88-42dc-bf01-8b0986485e78"
	oldB216ForeLegion   = "439537ca-53c1-47cb-9174-5872c3c46288"
	oldB216AngrathsFury = "61f3a549-77b4-4403-afa2-a0f02603f2e6"
	oldB216Amnesia      = "6f01c124-0518-428c-9efe-7f3741f3c109"
	oldB216Azor         = "89812139-9fc1-4fc9-b1d0-d07067d62083"
	oldB216Conquistador = "b20671e6-d740-4ef6-b408-bf8a0465c0da"
	oldB216ForeCoal     = "d6458cf6-eb1f-448b-8205-1b5d887ff2bf"
	oldB216Bear         = "Creature — Bear"
)

func TestOldB216EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB216VraskasScorn: "Vraska's Scorn",
		oldB216ForeEmpire:   "Forerunner of the Empire",
		oldB216BloodSun:     "Blood Sun",
		oldB216ForeLegion:   "Forerunner of the Legion",
		oldB216AngrathsFury: "Angrath's Fury",
		oldB216Amnesia:      "Induced Amnesia",
		oldB216Azor:         "Azor, the Lawbringer",
		oldB216Conquistador: "Legion Conquistador",
		oldB216ForeCoal:     "Forerunner of the Coalition",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB216Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Vraska's Scorn and Angrath's Fury ----------------------------------------

func TestOldB216VraskasScornDrainsFourAndMayFetchVraska(t *testing.T) {
	for _, zone := range []string{"library", "graveyard", "declined"} {
		t.Run(zone, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB216Seats(g)
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			var vraska uuid.UUID
			if zone == "graveyard" {
				vraska = effectstest.PushGraveyardCardWithTypeLineID(me, "Vraska, Scheming Gorgon", "Legendary Planeswalker — Vraska")
			} else {
				vraska = effectstest.PushLibraryCardForTest(me, game.Card{Name: "Vraska, Scheming Gorgon", TypeLine: "Legendary Planeswalker — Vraska"})
			}
			before, mine := opp.Life, me.Life
			effectstest.CastCatalogSpell(t, g, "Vraska's Scorn", "Sorcery", oldB216VraskasScorn, effectstest.PlayerRefs(opp.ID))
			effectstest.PassPriorityAroundTable(t, g)
			if opp.Life != before-4 || me.Life != mine {
				t.Fatalf("life: opponent %d (was %d), me %d (was %d); want a 4-life loss for the opponent only", opp.Life, before, me.Life, mine)
			}
			if zone == "declined" {
				effectstest.AnswerMayChoice(t, g, me.ID, false)
				effectstest.PassPriorityAroundTable(t, g)
				if me.Hand.Contains(vraska) {
					t.Error("declining still took Vraska")
				}
				return
			}
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			if zone == "graveyard" {
				effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			}
			effectstest.AnswerSearchByID(t, g, me.ID, vraska)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(vraska) {
				t.Errorf("Vraska did not come from the %s", zone)
			}
		})
	}
}

func TestOldB216VraskasScornTargetsOnlyAnOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Vraska's Scorn", "Sorcery", oldB216VraskasScorn, effectstest.PlayerRefs(me.ID)); err == nil {
		t.Error("the caster was a legal target")
	}
}

func TestOldB216AngrathsFuryDestroysDamagesAndMayFetchAngrath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB216Bear, 2, 2)
	effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	angrath := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Angrath, Minotaur Pirate", TypeLine: "Legendary Planeswalker — Angrath"})
	before := opp.Life
	effectstest.CastCatalogSpell(t, g, "Angrath's Fury", "Sorcery", oldB216AngrathsFury, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature was not destroyed")
	}
	if opp.Life != before-3 {
		t.Errorf("opponent life %d, want %d", opp.Life, before-3)
	}
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.AnswerSearchByID(t, g, me.ID, angrath)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(angrath) {
		t.Error("Angrath was not fetched")
	}
}

func TestOldB216AngrathsFuryHitsAPlaneswalkerAndStillDamagesWhenTheCreatureIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", oldB216Bear, 2, 2)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{"loyalty": 5}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Angrath's Fury", "Sorcery", oldB216AngrathsFury, []game.TargetRef{
		{Kind: game.TargetCard, ID: victim, Slot: 0}, {Kind: game.TargetCard, ID: walker, Slot: 1}})
	effectstest.B27Kill(g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.AangCounterCount(g, walker, "loyalty"); n != 2 {
		t.Errorf("loyalty %d, want 2 (3 damage: the other target's loss does not stop it, CR 608.2b)", n)
	}
	_ = me
}

// --- Blood Sun ----------------------------------------------------------------

func TestOldB216BloodSunDrawsAndStripsLandsOfAllButManaAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Warded Land", TypeLine: "Land",
		Keywords: []string{"hexproof"}, Owner: opp.ID, Controller: opp.ID})
	if !effectstest.EffectiveAbilitiesContain(t, g, land, "hexproof") {
		t.Fatal("setup: the land should have hexproof before Blood Sun")
	}
	before := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Blood Sun", "Enchantment", oldB216BloodSun, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d, want %d (the ETB draw)", got, before+1)
	}
	if effectstest.EffectiveAbilitiesContain(t, g, land, "hexproof") {
		t.Error("an opponent's land kept its keyword under Blood Sun")
	}
	spec, _ := effects.Lookup(oldB216BloodSun)
	if len(spec.Static) != 1 || !spec.Static[0].RemovesAbilities || !spec.Static[0].KeepsManaAbilities {
		t.Fatalf("Blood Sun's static is not the lose-all-but-mana ability: %+v", spec.Static)
	}
}

// --- the Forerunners and Legion Conquistador -----------------------------------

func TestOldB216ForerunnersTutorTheirTribeToTheTop(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, match, other string
	}{
		{"Forerunner of the Empire", "Creature — Human Soldier", oldB216ForeEmpire, "Creature — Dinosaur", "Creature — Goblin"},
		{"Forerunner of the Legion", "Creature — Vampire Knight", oldB216ForeLegion, "Creature — Vampire", "Creature — Goblin"},
		{"Forerunner of the Coalition", "Creature — Human Pirate", oldB216ForeCoal, "Creature — Pirate", "Creature — Goblin"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := oldB216Seats(g)
			match := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Match", TypeLine: tc.match})
			effectstest.PushLibraryCardForTest(me, game.Card{Name: "Other", TypeLine: tc.other})
			effectstest.CastHoldCreature(t, g, tc.name, tc.typeLine, tc.oracle, 2, 2)
			effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt after saying yes")
			}
			if effectstest.SearchOptionNamed(g, c, "Other") != uuid.Nil {
				t.Error("a card of another type was offered")
			}
			effectstest.AnswerSearchByID(t, g, me.ID, match)
			effectstest.PassPriorityAroundTable(t, g)
			top, err := me.Library.Top()
			if err != nil || top.InstanceID != match {
				t.Fatalf("top of library is %q, want the tutored card", top.Name)
			}
			if !top.IsKnownTo(opp.ID) {
				t.Error("the tutored card was not revealed")
			}
		})
	}
}

func TestOldB216ForerunnerOfTheEmpireDeclinedDoesNotSearch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Match", TypeLine: "Creature — Dinosaur"})
	effectstest.CastHoldCreature(t, g, "Forerunner of the Empire", "Creature — Human Soldier", oldB216ForeEmpire, 1, 3)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("a declined Forerunner still searched")
	}
}

func TestOldB216ForerunnerOfTheEmpireMayDamageEachCreatureWhenADinosaurEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	empire := effectstest.B12Push(g, me.ID, "Forerunner of the Empire", "Creature — Human Soldier", oldB216ForeEmpire, 1, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", oldB216Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", oldB216Bear, 1, 1)
	effectstest.CastHoldCreature(t, g, "Raptor", "Creature — Dinosaur", "", 3, 3)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the opponent's 1/1 survived 1 damage to each creature")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(empire) {
		t.Error("a creature with toughness above 1 died to 1 damage")
	}
	if got := effectstest.DamageOn(t, g, mine); got != 1 {
		t.Errorf("my own creature took %d damage, want 1 (each creature, not just theirs)", got)
	}
	// A non-Dinosaur entering does not trigger it.
	effectstest.CastHoldCreature(t, g, "Goblin", "Creature — Goblin", "", 1, 1)
	if lt := effectstest.LatestTriggerPrompt(g, me.ID); lt != nil {
		t.Error("a Goblin entering offered the damage")
	}
}

func TestOldB216ForerunnerOfTheLegionPumpsATargetWhenAnotherVampireEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	legion := effectstest.B12Push(g, me.ID, "Forerunner of the Legion", "Creature — Vampire Knight", oldB216ForeLegion, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB216Bear, 2, 2)
	effectstest.CastHoldCreature(t, g, "Bat", "Creature — Vampire", "", 1, 1)
	effectstest.AnswerPickTargetIfOpen(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("the target is %d/%d, want 3/3", p, tough)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, legion); p != 2 {
		t.Errorf("the Forerunner itself is %d power, want 2", p)
	}
}

func TestOldB216ForerunnerOfTheLegionDoesNotTriggerForItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB216Bear, 2, 2)
	effectstest.CastHoldCreature(t, g, "Forerunner of the Legion", "Creature — Vampire Knight", oldB216ForeLegion, 2, 2)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, bear); p != 2 {
		t.Errorf("the Bear has power %d, want 2: the Forerunner is not 'another' Vampire", p)
	}
}

func TestOldB216ForerunnerOfTheCoalitionDrainsEachOpponentForAnotherPirate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	effectstest.B12Push(g, me.ID, "Forerunner of the Coalition", "Creature — Human Pirate", oldB216ForeCoal, 2, 2)
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.CastHoldCreature(t, g, "Buccaneer", "Creature — Human Pirate", "", 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		want := lives[p.ID]
		if p.ID != me.ID {
			want--
		}
		if p.Life != want {
			t.Errorf("seat %s has %d life, want %d", p.ID, p.Life, want)
		}
	}
}

func TestOldB216LegionConquistadorFetchesEveryCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB216Seats(g)
	a := effectstest.TrekB1Library(g, me, "Legion Conquistador", "Creature — Vampire Soldier", oldB216Conquistador)
	b := effectstest.TrekB1Library(g, me, "Legion Conquistador", "Creature — Vampire Soldier", oldB216Conquistador)
	effectstest.TrekB1Library(g, me, "Other Soldier", "Creature — Human Soldier", "")
	effectstest.CastHoldCreature(t, g, "Legion Conquistador", "Creature — Vampire Soldier", oldB216Conquistador, 2, 2)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search was offered when the Conquistador entered")
	}
	if effectstest.SearchOptionNamed(g, c, "Other Soldier") != uuid.Nil {
		t.Error("the search offered a card with another name")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, a) != game.ZoneHand || effectstest.B12ZoneOf(g, b) != game.ZoneHand {
		t.Errorf("the namesakes are in %v and %v, want both in hand", effectstest.B12ZoneOf(g, a), effectstest.B12ZoneOf(g, b))
	}
}

// --- Induced Amnesia ------------------------------------------------------------

func oldB216Hand(p *game.Player, n int) []uuid.UUID {
	p.Hand.Cards = nil
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		id := uuid.New()
		p.Hand.PushTop(game.Card{InstanceID: id, Name: "Held", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
		ids = append(ids, id)
	}
	return ids
}

func TestOldB216InducedAmnesiaExilesTheHandFaceDownAndDrawsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	held := oldB216Hand(opp, 3)
	effectstest.CastCatalogSpell(t, g, "Induced Amnesia", "Enchantment", oldB216Amnesia, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range held {
		if z, _ := effectstest.CardWhere(g, opp, id); z != "exile" {
			t.Errorf("a held card is in %s, want exile", z)
		}
	}
	if got := opp.Hand.Size(); got != 3 {
		t.Errorf("the target has %d cards in hand, want 3 (drawn)", got)
	}
	_ = me
}

func TestOldB216InducedAmnesiaReturnsTheExiledCardsWhenItGoesToTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	held := oldB216Hand(opp, 2)
	amnesia := effectstest.CastCatalogSpell(t, g, "Induced Amnesia", "Enchantment", oldB216Amnesia, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	drawn := opp.Hand.Size()
	effectstest.P1Destroy(t, g, amnesia)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range held {
		if !opp.Hand.Contains(id) {
			t.Error("an exiled card did not return to its owner's hand")
		}
	}
	if got := opp.Hand.Size(); got != drawn+2 {
		t.Errorf("hand %d, want %d", got, drawn+2)
	}
	_ = me
}

func TestOldB216InducedAmnesiaLeavingAnotherWayKeepsTheCardsExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB216Seats(g)
	held := oldB216Hand(opp, 2)
	amnesia := effectstest.CastCatalogSpell(t, g, "Induced Amnesia", "Enchantment", oldB216Amnesia, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(amnesia); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range held {
		if z, _ := effectstest.CardWhere(g, opp, id); z != "exile" {
			t.Errorf("a held card is in %s, want it still exiled (the enchantment was exiled, not put into a graveyard)", z)
		}
	}
}

// --- Azor, the Lawbringer ---------------------------------------------------------

func oldB216ToTurnOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; i < 400 && !(g.Turn.ActiveSeat == seat && g.Turn.Step == game.StepPrecombatMain); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("never reached seat %d's main phase", seat)
	}
}

func oldB216TryCast(g *game.Game, p *game.Player, name, typeLine, cost string) error {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Owner: p.ID, Controller: p.ID})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true})
}

func TestOldB216AzorBansInstantsAndSorceriesDuringEachOpponentsNextTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	meSeat := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Azor, the Lawbringer", "Legendary Creature — Sphinx", oldB216Azor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// During my own turn the ban reaches nothing (it is about THEIR turn).
	effectstest.AddPool(t, g, opp.ID, "{R}")
	if err := oldB216TryCast(g, opp, "Opt", "Instant", "{R}"); err != nil {
		t.Errorf("an opponent could not cast an instant on my turn, before their next turn: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// During their next turn instants and sorceries are barred, creatures not.
	oppSeat := (meSeat + 1) % len(g.Seats)
	oldB216ToTurnOf(t, g, oppSeat)
	effectstest.AddPool(t, g, opp.ID, "{R}{R}{R}")
	if err := oldB216TryCast(g, opp, "Shock", "Instant", "{R}"); err == nil {
		t.Error("an opponent cast an instant during their next turn")
	}
	if err := oldB216TryCast(g, opp, "Divination", "Sorcery", "{R}"); err == nil {
		t.Error("an opponent cast a sorcery during their next turn")
	}
	if err := oldB216TryCast(g, opp, "Bear", "Creature — Bear", "{R}"); err != nil {
		t.Errorf("a creature spell was barred: %v", err)
	}
	// The third opponent's next turn is banned too, and mine is not.
	third := (meSeat + 2) % len(g.Seats)
	oldB216ToTurnOf(t, g, third)
	effectstest.AddPool(t, g, g.Seats[third].ID, "{R}")
	if err := oldB216TryCast(g, g.Seats[third], "Shock", "Instant", "{R}"); err == nil {
		t.Error("the second opponent cast an instant during their next turn")
	}
	_ = me
}

func TestOldB216AzorsBanEndsWithThatTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := oldB216Seats(g)
	meSeat := g.Turn.ActiveSeat
	effectstest.CastCatalogSpell(t, g, "Azor, the Lawbringer", "Legendary Creature — Sphinx", oldB216Azor, nil)
	effectstest.PassPriorityAroundTable(t, g)
	oppSeat := (meSeat + 1) % len(g.Seats)
	oldB216ToTurnOf(t, g, oppSeat)
	// Walk through that opponent's whole turn and into the table's next round
	// until their following turn.
	oldB216ToTurnOf(t, g, (oppSeat+1)%len(g.Seats))
	oldB216ToTurnOf(t, g, (oppSeat+2)%len(g.Seats))
	oldB216ToTurnOf(t, g, (oppSeat+3)%len(g.Seats))
	oldB216ToTurnOf(t, g, oppSeat)
	effectstest.AddPool(t, g, opp.ID, "{R}")
	if err := oldB216TryCast(g, opp, "Shock", "Instant", "{R}"); err != nil {
		t.Errorf("the ban outlived their next turn: %v", err)
	}
}

func TestOldB216AzorAttackPaysXToGainAndDrawX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	effectstest.ToMain(t, g)
	azor := effectstest.B12Push(g, me.ID, "Azor, the Lawbringer", "Legendary Creature — Sphinx", oldB216Azor, 6, 6)
	before, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(azor, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B06AddMana(me, "W", "U", "U", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	pa := effectstest.AnswerNumber(t, g, me.ID, 2)
	if pa.Min != 0 || pa.Max != 3 {
		t.Errorf("X prompt = %+v, want 0..3", pa)
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if left := me.ManaPool.Total(); left != 1 {
		t.Errorf("%d mana left, want 1 (paid {2}{W}{U}{U})", left)
	}
}

func TestOldB216AzorAttackDeclinedPaysNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB216Seats(g)
	effectstest.ToMain(t, g)
	azor := effectstest.B12Push(g, me.ID, "Azor, the Lawbringer", "Legendary Creature — Sphinx", oldB216Azor, 6, 6)
	before, life := me.Hand.Size(), me.Life
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(azor, opp.ID); err != nil {
		t.Fatal(err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.B06AddMana(me, "W", "U", "U", "C", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerNumber(t, g, me.ID, 2)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != before || me.Life != life || me.ManaPool.Total() != 6 {
		t.Errorf("declining changed hand %d→%d, life %d→%d or mana %d", before, me.Hand.Size(), life, me.Life, me.ManaPool.Total())
	}
}
