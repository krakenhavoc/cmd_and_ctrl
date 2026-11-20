package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b15_cards_test.go — Phyrexia: All Will Be One, batch 15 (#5479).
// Each test plays the real catalog entry through the engine.

const (
	oneB15BlackSunOracle   = "af105cda-1f7c-4868-9407-a97c46bd762d"
	oneB15GreenSunOracle   = "0221330a-3e7a-40aa-9c78-85c88a4c1a53"
	oneB15OddsOracle       = "f4a73cde-58bf-4719-8eb0-e039fdde4f62"
	oneB15DemolitionOracle = "66465bc1-4a34-4951-be53-57f90002976a"
	oneB15CorruptionOracle = "0b49924b-cf83-46d9-b665-9c718a40e16c"
	oneB15NahirisOracle    = "a84cc828-8d1b-4e61-a13f-60124e2ee35f"
	oneB15PredationOracle  = "478e59b8-03ad-4f8e-8c18-6ab5ad77a6f8"
	oneB15VanishOracle     = "c14e422b-6412-45ce-8e82-d47ba61557a0"
	oneB15DisruptionOracle = "6128679e-7e54-41e4-ab6e-25aeeffed49f"
)

func TestOneB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oneB15BlackSunOracle:   "Black Sun's Twilight",
		oneB15GreenSunOracle:   "Green Sun's Twilight",
		oneB15OddsOracle:       "Against All Odds",
		oneB15DemolitionOracle: "Gleeful Demolition",
		oneB15CorruptionOracle: "Monumental Corruption",
		oneB15NahirisOracle:    "Nahiri's Sacrifice",
		oneB15PredationOracle:  "Ruthless Predation",
		oneB15VanishOracle:     "Vanish into Eternity",
		oneB15DisruptionOracle: "Planar Disruption",
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

func oneB15Damage(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return effectstest.E2Card(t, g, id).DamageMarked
}

// --- Ruthless Predation ------------------------------------------------------

func TestOneB15RuthlessPredationPumpsThenFights(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Ruthless Predation", "Sorcery", oneB15PredationOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine, Slot: 0}, {Kind: game.TargetCard, ID: theirs, Slot: 1}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their 3/3 survived the +1/+2 fighter's 3 damage")
	}
	if !g.Battlefield.Contains(mine) {
		t.Fatal("your 3/4 died to 3 damage")
	}
	if p, tg := effectstest.SocB14PT(t, g, mine); p != 3 || tg != 4 {
		t.Errorf("your creature is %d/%d, want 3/4", p, tg)
	}
	if d := oneB15Damage(t, g, mine); d != 3 {
		t.Errorf("damage on your creature %d, want 3", d)
	}
}

func TestOneB15RuthlessPredationNeedsBothTargetsOnTheirSide(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	ally := effectstest.B12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	cast := func(targets ...game.TargetRef) error {
		card := game.Card{Name: "Ruthless Predation", TypeLine: "Sorcery", OracleID: oneB15PredationOracle}
		card.Owner, card.Controller, card.InstanceID = me.ID, me.ID, uuid.New()
		me.Hand.PushTop(card)
		return g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: targets})
	}
	if err := cast(game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 0}, game.TargetRef{Kind: game.TargetCard, ID: ally, Slot: 1}); err == nil {
		t.Error("a creature you control was accepted as the one you don't control")
	}
	if err := cast(game.TargetRef{Kind: game.TargetCard, ID: theirs, Slot: 0}, game.TargetRef{Kind: game.TargetCard, ID: theirs, Slot: 1}); err == nil {
		t.Error("an opponent's creature was accepted as the one you control")
	}
	if err := cast(game.TargetRef{Kind: game.TargetCard, ID: mine, Slot: 0}); err == nil {
		t.Error("the spell was cast without the second target")
	}
}

// --- Planar Disruption -------------------------------------------------------

func TestOneB15PlanarDisruptionLocksCombatAndEveryActivatedAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	feeder := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Carrion Feeder", TypeLine: "Creature — Zombie",
		Power: 1, Toughness: 1, OracleID: effectstest.CarrionFeederOracle, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Planar Disruption", "Enchantment — Aura", oneB15DisruptionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: feeder}})
	effectstest.PassPriorityAroundTable(t, g)
	got := effectstest.RestrictionsOf(t, g, feeder)
	want := game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana
	if got&want != want {
		t.Errorf("restrictions = %v, want can't attack/block and can't activate, mana abilities included", got.Names())
	}
	if err := g.ActivateCatalogAbility(me.ID, feeder, 0, game.ActivateAbilityParams{}); err != game.ErrCantActivate {
		t.Errorf("an enchanted permanent activated an ability: %v", err)
	}
}

func TestOneB15PlanarDisruptionEnchantsArtifactCreatureOrPlaneswalkerOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	land := effectstest.PushCatalogPermanent(g, me.ID, "Plains", "Basic Land — Plains", "", false)
	rock := effectstest.PushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	walker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Walker", TypeLine: "Legendary Planeswalker — Test",
		Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 4},
	})
	if err := effectstest.CastCatalogSpellErr(t, g, "Planar Disruption", "Enchantment — Aura", oneB15DisruptionOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: land}}); err == nil {
		t.Error("the Aura was cast on a land")
	}
	for _, id := range []uuid.UUID{rock, walker} {
		effectstest.CastCatalogSpell(t, g, "Planar Disruption", "Enchantment — Aura", oneB15DisruptionOracle,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.RestrictionsOf(t, g, id)&game.CantActivate == 0 {
			t.Errorf("%s is not locked", effectstest.E2Card(t, g, id).Name)
		}
	}
}

// --- Vanish into Eternity ----------------------------------------------------

func TestOneB15VanishCostsThreeMoreAtACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	if got := effectstest.SelfPriced(t, g, me, oneB15VanishOracle, "Instant", "{2}{W}", game.ZoneHand, effectstest.CardRefs(bear)).ManaValue(); got != 6 {
		t.Errorf("Vanish at a creature costs %d, want 6", got)
	}
	if got := effectstest.SelfPriced(t, g, me, oneB15VanishOracle, "Instant", "{2}{W}", game.ZoneHand, effectstest.CardRefs(rock)).ManaValue(); got != 3 {
		t.Errorf("Vanish at an artifact costs %d, want 3", got)
	}
}

func TestOneB15VanishExilesANonlandPermanentAndRefusesALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Rock", "Artifact", "", false)
	land := effectstest.PushCatalogPermanent(g, opp.ID, "Plains", "Basic Land — Plains", "", false)
	if err := effectstest.CastCatalogSpellErr(t, g, "Vanish into Eternity", "Instant", oneB15VanishOracle, effectstest.CardRefs(land)); err == nil {
		t.Error("Vanish into Eternity targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Vanish into Eternity", "Instant", oneB15VanishOracle, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || !effectstest.InExile(g, rock) {
		t.Error("the artifact was not exiled")
	}
}

// --- Gleeful Demolition ------------------------------------------------------

func TestOneB15GleefulDemolitionMakesGoblinsOnlyForYourOwnArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	effectstest.CastCatalogSpell(t, g, "Gleeful Demolition", "Sorcery", oneB15DemolitionOracle, effectstest.CardRefs(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("their artifact survived")
	}
	if n := effectstest.SocB8Named(g, me.ID, "Phyrexian Goblin"); len(n) != 0 {
		t.Errorf("%d Goblins for destroying an opponent's artifact, want 0", len(n))
	}
	mine := effectstest.PushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	effectstest.CastCatalogSpell(t, g, "Gleeful Demolition", "Sorcery", oneB15DemolitionOracle, effectstest.CardRefs(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("your artifact survived")
	}
	if n := effectstest.SocB8Named(g, me.ID, "Phyrexian Goblin"); len(n) != 3 {
		t.Errorf("%d Goblins for destroying your own artifact, want 3", len(n))
	}
}

// --- Monumental Corruption ---------------------------------------------------

func TestOneB15MonumentalCorruptionDrawsAndDrainsByYourArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Rock A", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Rock B", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	hand, life := opp.Hand.Size(), effectstest.LifeOf(g, opp.ID)
	effectstest.CastCatalogSpell(t, g, "Monumental Corruption", "Sorcery", oneB15CorruptionOracle, effectstest.SoPlayer(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - hand; got != 2 {
		t.Errorf("opponent drew %d, want 2 (your artifacts, not theirs)", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != life-2 {
		t.Errorf("opponent's life %d → %d, want -2", life, got)
	}
}

// --- Against All Odds --------------------------------------------------------

func TestOneB15AgainstAllOddsBlinksAndReturnsBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	dead := effectstest.SeedGraveyardCreature(me, "Small Dead", "{2}")
	big := effectstest.SeedGraveyardCreature(me, "Big Dead", "{4}")
	effectstest.AdvanceToMain(t, g)
	card := game.Card{Name: "Against All Odds", TypeLine: "Sorcery", OracleID: oneB15OddsOracle}
	card.Owner, card.Controller, card.InstanceID = me.ID, me.ID, uuid.New()
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{
		Strict: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	}); err == nil {
		t.Error("a mana value 4 card was accepted for the return mode")
	}
	me.Hand.PushTop(card)
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{
		Strict: true, Modes: []int{0, 1},
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: dead}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) {
		t.Error("the blinked creature kept its old object")
	}
	if len(effectstest.SocB8Named(g, me.ID, "Mine")) != 1 {
		t.Error("the blinked creature did not return")
	}
	if !g.Battlefield.Contains(dead) {
		t.Error("the small creature card was not returned from the graveyard")
	}
}

func TestOneB15AgainstAllOddsReturnOnlyMode(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	rock := effectstest.SeedGraveyardCard(t, g, "Rock", "Artifact", "")
	castCatalogSpellWithModesTargets(t, g, oneB15OddsOracle, []int{1}, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rock) {
		t.Error("the artifact card was not returned")
	}
}

func castCatalogSpellWithModesTargets(t *testing.T, g *game.Game, oracle string, modes []int, targets []game.TargetRef) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Against All Odds", TypeLine: "Sorcery", OracleID: oracle,
		Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
}

// --- Black Sun's Twilight ----------------------------------------------------

func TestOneB15BlackSunsTwilightShrinksAndKills(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.D1658Cast(t, g, "Black Sun's Twilight", "Instant", oneB15BlackSunOracle, game.CastSpellParams{
		XValue: 2, Targets: effectstest.CardRefs(bear),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("a 2/2 survived -2/-2")
	}
}

func TestOneB15BlackSunsTwilightBelowFiveReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 9, 9)
	dead := effectstest.SeedGraveyardCreature(me, "Dead", "{2}")
	effectstest.D1658Cast(t, g, "Black Sun's Twilight", "Instant", oneB15BlackSunOracle, game.CastSpellParams{
		XValue: 4, Targets: effectstest.CardRefs(bear),
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || g.Battlefield.Contains(dead) {
		t.Error("X of 4 returned a creature card")
	}
}

func TestOneB15BlackSunsTwilightAtFiveReturnsATappedCreatureWithoutATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	small := effectstest.SeedGraveyardCreature(me, "Small", "{3}")
	tooBig := effectstest.SeedGraveyardCreature(me, "Too Big", "{6}")
	effectstest.D1658Cast(t, g, "Black Sun's Twilight", "Instant", oneB15BlackSunOracle, game.CastSpellParams{XValue: 5})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no creature card was offered")
	}
	for _, id := range pick.ChooseCards {
		if id == tooBig {
			t.Error("a mana value 6 card was offered at X = 5")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, small)
	c := effectstest.E2Card(t, g, small)
	if !g.Battlefield.Contains(small) || !c.Tapped {
		t.Errorf("returned card on battlefield %v, tapped %v, want both", g.Battlefield.Contains(small), c.Tapped)
	}
}

// --- Green Sun's Twilight ----------------------------------------------------

func oneB15SeedTop(me *game.Player, cards ...[2]string) []uuid.UUID {
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		ids[i] = effectstest.SeedLibraryTop(me, cards[i][0], cards[i][1], "")
	}
	return ids
}

func TestOneB15GreenSunsTwilightSmallTakesACreatureAndALandToHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oneB15SeedTop(me,
		[2]string{"Bear", "Creature — Bear"}, [2]string{"Elk", "Creature — Elk"},
		[2]string{"Forest", "Basic Land — Forest"}, [2]string{"Rock", "Artifact"})
	effectstest.D1658Cast(t, g, "Green Sun's Twilight", "Sorcery", oneB15GreenSunOracle, game.CastSpellParams{XValue: 3})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[1], ids[2])
	if !me.Hand.Contains(ids[1]) || !me.Hand.Contains(ids[2]) {
		t.Error("the chosen creature and land did not reach the hand")
	}
	if me.Hand.Contains(ids[0]) || me.Hand.Contains(ids[3]) {
		t.Error("an unchosen card reached the hand")
	}
	if me.Library.Contains(ids[0]) == false || me.Library.Contains(ids[3]) == false {
		t.Error("the rest left the library")
	}
}

func TestOneB15GreenSunsTwilightSmallRefusesTwoCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oneB15SeedTop(me, [2]string{"Bear", "Creature — Bear"}, [2]string{"Elk", "Creature — Elk"}, [2]string{"Forest", "Basic Land — Forest"})
	effectstest.D1658Cast(t, g, "Green Sun's Twilight", "Sorcery", oneB15GreenSunOracle, game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick was offered")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Error("two creature cards were accepted")
	}
}

func TestOneB15GreenSunsTwilightFiveSendsChosenCardsToBattlefieldOrHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := oneB15SeedTop(me,
		[2]string{"Bear", "Creature — Bear"}, [2]string{"Forest", "Basic Land — Forest"},
		[2]string{"Rock", "Artifact"}, [2]string{"Rock 2", "Artifact"}, [2]string{"Rock 3", "Artifact"}, [2]string{"Rock 4", "Artifact"})
	effectstest.D1658Cast(t, g, "Green Sun's Twilight", "Sorcery", oneB15GreenSunOracle, game.CastSpellParams{XValue: 5})
	effectstest.PassPriorityAroundTable(t, g)
	// One creature and one land are the only legal set, so the pick is forced
	// and the placement question comes straight away.
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no placement question was asked")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, ids[0])
	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the creature did not enter the battlefield")
	}
	if !me.Hand.Contains(ids[1]) {
		t.Error("the land not named went to the hand")
	}
	for _, id := range ids[2:] {
		if !me.Library.Contains(id) {
			t.Error("a revealed card was not left in the library")
		}
	}
}

// --- Nahiri's Sacrifice ------------------------------------------------------

func TestOneB15NahirisSacrificeDividesDamageEqualToTheSacrificedValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PaPermanent(g, me.ID, "Fodder", "Artifact Creature — Construct", "{3}", 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	b := effectstest.B12Creature(g, opp.ID, "B", "Creature — Wall", 0, 30)
	effectstest.D1658Cast(t, g, "Nahiri's Sacrifice", "Sorcery", oneB15NahirisOracle, game.CastSpellParams{
		XValue: 3, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(a, b),
		Distribution: map[uuid.UUID]int{a: 1, b: 2},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed permanent stayed")
	}
	if d := oneB15Damage(t, g, a); d != 1 {
		t.Errorf("a took %d, want 1", d)
	}
	if d := oneB15Damage(t, g, b); d != 2 {
		t.Errorf("b took %d, want 2", d)
	}
}

func TestOneB15NahirisSacrificeRefusesAPermanentOfTheWrongValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fodder := effectstest.PaPermanent(g, me.ID, "Fodder", "Creature — Construct", "{2}", 1, 1)
	a := effectstest.B12Creature(g, opp.ID, "A", "Creature — Wall", 0, 30)
	err := effectstest.D1658CastErr(t, g, "Nahiri's Sacrifice", "Sorcery", oneB15NahirisOracle, game.CastSpellParams{
		Strict: true, XValue: 3, SacrificeIDs: []uuid.UUID{fodder}, Targets: effectstest.CardRefs(a),
		Distribution: map[uuid.UUID]int{a: 3},
	})
	if err == nil {
		t.Error("a mana value 2 permanent paid for X = 3")
	}
}
