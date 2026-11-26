package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b397_cards_test.go — Theros, S82 older sets batch 397 (#5561): Sea
// God's Revenge, Psychic Intrusion, Tymaret, the Murder King, Abhorrent
// Overlord, Prophet of Kruphix, Satyr Hedonist, Ashiok, Nightmare Weaver and
// Witches' Eye, each through the real cast, activation and trigger paths.

const (
	ths397Revenge  = "8ab71a49-ce81-44f7-b940-50e7aba64df9"
	ths397Intrude  = "fdb10291-b5e0-4da8-a103-8e7ac7f8177b"
	ths397Tymaret  = "0a029c61-b1de-427b-a8e8-36f3a23f001b"
	ths397Overlord = "13f63240-a6bb-4333-a3dd-818725c6fb73"
	ths397Prophet  = "372351fd-58f7-4cb8-8644-5fbc94ccb145"
	ths397Satyr    = "a75bf714-ed3e-4403-a009-3a81f430bc64"
	ths397Ashiok   = "b6f52260-9ad5-4024-acf9-d77f3c08c872"
	ths397Eye      = "c8d479b8-b96c-4382-97d3-68d85cf371ba"
)

func TestS82OldB397EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		ths397Revenge: "Sea God's Revenge", ths397Intrude: "Psychic Intrusion",
		ths397Tymaret: "Tymaret, the Murder King", ths397Overlord: "Abhorrent Overlord",
		ths397Prophet: "Prophet of Kruphix", ths397Satyr: "Satyr Hedonist",
		ths397Ashiok: "Ashiok, Nightmare Weaver", ths397Eye: "Witches' Eye",
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

// --- Sea God's Revenge ---

func TestS82OldB397SeaGodsRevengeBouncesUpToThreeOpposingCreaturesAndScrys(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	var theirs []uuid.UUID
	for _, n := range []string{"A", "B", "C"} {
		theirs = append(theirs, effectstest.B12Creature(g, opp.ID, n, "Creature — Bear", 2, 2))
	}
	spare := effectstest.B12Creature(g, opp.ID, "D", "Creature — Bear", 2, 2)
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.CastCatalogSpell(t, g, "Sea God's Revenge", "Sorcery", ths397Revenge, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs[0]}, {Kind: game.TargetCard, ID: theirs[1]}, {Kind: game.TargetCard, ID: theirs[2]},
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range theirs {
		if effectstest.AppInZone(g, id, game.ZoneBattlefield) {
			t.Errorf("%v was not returned", id)
		}
	}
	if !effectstest.AppInZone(g, spare, game.ZoneBattlefield) {
		t.Error("the untargeted creature left")
	}
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt after the bounce")
	}
}

func TestS82OldB397SeaGodsRevengeCannotTargetYourOwnCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sea God's Revenge", TypeLine: "Sorcery", OracleID: ths397Revenge, Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)})
	if err == nil {
		t.Error("an own creature was a legal target")
	}
}

// --- Psychic Intrusion ---

func TestS82OldB397PsychicIntrusionTakesANonlandCardAndLetsYouCastItWithAnyMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	opp.Graveyard.Cards = nil
	effectstest.CastAt(t, g, "Psychic Intrusion", "Sorcery", ths397Intrude, effectstest.PbPlayer(opp.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{hand[1]}) {
		t.Fatalf("options = %v, want the Bolt only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{hand[1]}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(hand[1]) {
		t.Fatal("the chosen card was not exiled")
	}
	g.Turn.Step = game.StepPrecombatMain
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, hand[1], game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the exiled Bolt with blue mana: %v", err)
	}
}

func TestS82OldB397PsychicIntrusionCanTakeAGraveyardCardAndNotALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.RevealHand(opp)
	opp.Graveyard.Cards = nil
	land, spell := uuid.New(), uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: land, Name: "Swamp", TypeLine: "Basic Land — Swamp", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Dark Ritual", TypeLine: "Instant", ManaCost: "{B}", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	effectstest.CastAt(t, g, "Psychic Intrusion", "Sorcery", ths397Intrude, effectstest.PbPlayer(opp.ID))
	pick := effectstest.OpenVariant(t, g)
	if !effectstest.SameIDs(pick.DiscardOptions, []uuid.UUID{spell}) {
		t.Fatalf("options = %v, want the graveyard Ritual only", pick.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	g.SettleResolution()
	if !g.Exile.Contains(spell) || !opp.Graveyard.Contains(land) {
		t.Error("the Ritual was not exiled or the Swamp left the graveyard")
	}
}

// --- Tymaret, the Murder King ---

func TestS82OldB397TymaretSacrificesAnotherCreatureForTwoDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	tym := effectstest.B27Push(g, me.ID, "Tymaret, the Murder King", "Legendary Creature — Zombie Warrior", ths397Tymaret, "{B}{R}", 2, 2, "B", "R")
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	before := opp.Life
	effectstest.B16Activate(t, g, me.ID, tym, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AppInZone(g, fodder, game.ZoneBattlefield) {
		t.Error("the sacrificed creature stayed")
	}
	if opp.Life != before-2 {
		t.Errorf("life %d → %d, want 2 damage", before, opp.Life)
	}
}

func TestS82OldB397TymaretCannotSacrificeItselfForTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	tym := effectstest.B27Push(g, me.ID, "Tymaret, the Murder King", "Legendary Creature — Zombie Warrior", ths397Tymaret, "{B}{R}", 2, 2, "B", "R")
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	err := g.ActivateCatalogAbility(me.ID, tym, 0, game.ActivateAbilityParams{
		Strict: true, SacrificeIDs: []uuid.UUID{tym}, Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)},
	})
	if err == nil {
		t.Error("Tymaret sacrificed itself to its own first ability")
	}
}

func TestS82OldB397TymaretReturnsFromTheGraveyardByPayingAndSacrificing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Tymaret, the Murder King", TypeLine: "Legendary Creature — Zombie Warrior", OracleID: ths397Tymaret, ManaCost: "{B}{R}", Owner: me.ID, Controller: me.ID})
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BB")
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("Tymaret did not return to hand")
	}
	if effectstest.AppInZone(g, fodder, game.ZoneBattlefield) {
		t.Error("the sacrificed creature stayed")
	}
}

// --- Abhorrent Overlord ---

func overlord397InHand(g *game.Game, me *game.Player) uuid.UUID {
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Abhorrent Overlord", TypeLine: "Creature — Demon", OracleID: ths397Overlord,
		ManaCost: "{5}{B}{B}", Colors: []string{"B"}, Power: 6, Toughness: 6, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	return id
}

func TestS82OldB397AbhorrentOverlordMakesHarpiesEqualToYourDevotionToBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := effectstest.SeatsForCombat(g)
	effectstest.B27Push(g, me.ID, "Black Permanent", "Creature — Bear", "", "{1}{B}{B}", 2, 2, "B")
	id := overlord397InHand(g, me)
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// {B}{B} from the other permanent and {B}{B} from the Overlord itself.
	if n := effectstest.CountOnBattlefield(g, "Harpy", me.ID); n != 4 {
		t.Errorf("%d Harpies, want 4", n)
	}
}

func TestS82OldB397AbhorrentOverlordSacrificesACreatureEachUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	overlord := effectstest.MirB1Push(g, me.ID, "Abhorrent Overlord", "Creature — Demon", ths397Overlord, 6, 6, "flying")
	snack := effectstest.PushVanillaCreature(g, me.ID, "Snack", 1, 1)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.AnswerSacrifice(t, g, me.ID, snack)
	if g.Battlefield.Contains(snack) || !g.Battlefield.Contains(overlord) {
		t.Error("the Snack was not the sacrifice")
	}
}

// --- Prophet of Kruphix ---

func TestS82OldB397ProphetUntapsYourCreaturesAndLandsDuringOtherPlayersUntapSteps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Prophet of Kruphix", "Creature — Human Wizard", ths397Prophet, "{3}{G}{U}", 2, 3, "G", "U")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushLand(g, me.ID, "Forest")
	effectstest.B16Tap(g, bear)
	effectstest.B16Tap(g, land)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if effectstest.B16Tapped(t, g, bear) || effectstest.B16Tapped(t, g, land) {
		t.Error("my creature or land stayed tapped through the other player's untap step")
	}
}

func TestS82OldB397ProphetGivesCreatureSpellsFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	effectstest.B27Push(g, me.ID, "Prophet of Kruphix", "Creature — Human Wizard", ths397Prophet, "{3}{G}{U}", 2, 3, "G", "U")
	g.Turn.ActiveSeat = 0
	g.Turn.Step = game.StepPrecombatMain
	creature, sorcery := uuid.New(), uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: creature, Name: "Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Hand.PushTop(game.Card{InstanceID: sorcery, Name: "Rite", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a creature spell on another player's turn was refused: %v", err)
	}
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a sorcery was castable on another player's turn")
	}
}

// --- Satyr Hedonist ---

func TestS82OldB397SatyrHedonistSacrificesForThreeRedMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	satyr := effectstest.B27Push(g, me.ID, "Satyr Hedonist", "Creature — Satyr", ths397Satyr, "{1}{G}", 2, 1, "G")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatForTest(g, me, "R")
	effectstest.ActivateManaFor(t, g, me.ID, satyr, 0, game.ManaAbilityParams{})
	if got := effectstest.PoolCount(me, "R"); got != 3 {
		t.Errorf("%d red mana, want 3 (the {R} paid is spent)", got)
	}
	if effectstest.AppInZone(g, satyr, game.ZoneBattlefield) {
		t.Error("the Satyr was not sacrificed")
	}
}

func TestS82OldB397SatyrHedonistCannotActivateWithoutTheRed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	satyr := effectstest.B27Push(g, me.ID, "Satyr Hedonist", "Creature — Satyr", ths397Satyr, "{1}{G}", 2, 1, "G")
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	if err := g.ActivateManaAbility(me.ID, satyr, 0, game.ManaAbilityParams{}); err == nil {
		t.Error("activated with no mana to pay")
	}
}

// --- Witches' Eye ---

func TestS82OldB397WitchesEyeGrantsTapAndPayOneToScry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	eye := effectstest.B27Push(g, me.ID, "Witches' Eye", "Artifact — Equipment", ths397Eye, "{1}", 0, 0)
	host := effectstest.B16Creature(g, me.ID, "Host", "Creature — Bear", 2, 2, "G")
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	if idx, _ := effectstest.GrantedActivatedIndex(t, g, host); idx >= 0 {
		t.Fatal("the creature has the ability before it is equipped")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	effectstest.EquipTo(t, g, me.ID, eye, host)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, host)
	if idx < 0 {
		t.Fatal("the equipped creature has no granted ability")
	}
	effectstest.B16Activate(t, g, me.ID, host, idx, game.ActivateAbilityParams{Ref: ref})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Error("no scry prompt")
	}
	if !effectstest.LayeredCard(t, g, host).Tapped {
		t.Error("the ability's {T} cost was not paid")
	}
}

// --- Ashiok, Nightmare Weaver ---

// ashiok397Library stacks the cards top first.
func ashiok397Library(opp *game.Player, cards ...game.Card) []uuid.UUID {
	opp.Library.Cards = nil
	var ids []uuid.UUID
	for _, c := range cards {
		c.Owner, c.Controller = opp.ID, opp.ID
		ids = append(ids, effectstest.PushLibraryCardForTest(opp, c))
	}
	return ids
}

func TestS82OldB397AshiokPlusTwoExilesTheTopThreeOfTheTargetOpponentsLibrary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	walker := effectstest.S82Walker(g, me.ID, "Ashiok, Nightmare Weaver", ths397Ashiok, 3)
	ids := ashiok397Library(opp,
		game.Card{Name: "First", TypeLine: "Sorcery"}, game.Card{Name: "Second", TypeLine: "Sorcery"},
		game.Card{Name: "Third", TypeLine: "Sorcery"}, game.Card{Name: "Deep", TypeLine: "Sorcery"})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range ids[:3] {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled", id)
		}
	}
	if g.Exile.Contains(ids[3]) || !opp.Library.Contains(ids[3]) {
		t.Error("the fourth card left the library")
	}
	if got := effectstest.LoyaltyOf(g, walker); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestS82OldB397AshiokMinusXPutsAMatchingExiledCreatureOntoTheBattlefieldAsANightmare(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	walker := effectstest.S82Walker(g, me.ID, "Ashiok, Nightmare Weaver", ths397Ashiok, 3)
	ids := ashiok397Library(opp,
		game.Card{Name: "Spell", TypeLine: "Sorcery", ManaCost: "{2}"},
		game.Card{Name: "Wrong Size", TypeLine: "Creature — Ogre", ManaCost: "{3}", Power: 3, Toughness: 3},
		game.Card{Name: "Fit", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Filler", TypeLine: "Sorcery", ManaCost: "{1}"})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	g.LoyaltyActivatedThisTurn = nil
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{XValue: 2})
	if got := effectstest.LoyaltyOf(g, walker); got != 3 {
		t.Errorf("loyalty %d after +2 then −2, want 3", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, ids[2])
	g.SettleResolution()
	var fit *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Fit" {
			fit = &g.Battlefield.Cards[i]
		}
	}
	if fit == nil {
		t.Fatal("the creature card did not enter")
	}
	if fit.Controller != me.ID {
		t.Errorf("controller %v, want me", fit.Controller)
	}
	if !effectstest.LayeredCard(t, g, fit.InstanceID).HasSubtype("Nightmare") {
		t.Error("the creature is not a Nightmare")
	}
	if !g.Exile.Contains(ids[1]) {
		t.Error("the card of the wrong mana value left exile")
	}
}

func TestS82OldB397AshiokMinusXWithNoMatchingCardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	walker := effectstest.S82Walker(g, me.ID, "Ashiok, Nightmare Weaver", ths397Ashiok, 3)
	ashiok397Library(opp, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", ManaCost: "{3}", Power: 3, Toughness: 3})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{effectstest.PbPlayer(opp.ID)}})
	effectstest.PassPriorityAroundTable(t, g)
	g.LoyaltyActivatedThisTurn = nil
	effectstest.S82Activate(t, g, me.ID, walker, 1, game.ActivateAbilityParams{XValue: 1})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Ogre" {
			t.Error("a creature of the wrong mana value entered")
		}
	}
}

func TestS82OldB397AshiokMinusTenExilesEveryOpposingHandAndGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	walker := effectstest.S82Walker(g, me.ID, "Ashiok, Nightmare Weaver", ths397Ashiok, 10)
	hand := effectstest.RevealHand(opp, effectstest.RhForest(), effectstest.RhBolt())
	opp.Graveyard.Cards = nil
	grave := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Dead", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	mine := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: mine, Name: "Mine", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	effectstest.AdvanceToMain(t, g)
	effectstest.S82Activate(t, g, me.ID, walker, 2, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range append(hand, grave) {
		if !g.Exile.Contains(id) {
			t.Errorf("%v was not exiled", id)
		}
	}
	if !me.Graveyard.Contains(mine) {
		t.Error("my own graveyard card was exiled")
	}
}
