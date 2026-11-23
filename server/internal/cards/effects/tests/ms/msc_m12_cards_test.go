package testsms

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_m12_cards_test.go — Marvel Super Heroes Commander, M–Z half, batch 12
// (#4455). Each test plays the real catalog entry and asserts what its text
// says.

const (
	mscM12ViperOracle     = "6824a6ed-37d0-4db5-b2d4-9c737d0c49f9"
	mscM12ThorAsgardOracl = "44c2fefb-5de3-4406-955a-3fabc46357d0"
	mscM12ThorMidgardOrac = "74acdcbc-5185-422a-a236-0a126edd2eee"
	mscM12WhiteTigerOracl = "0fa01155-638d-44ed-8963-46b202b0e720"
	mscM12StegronOracle   = "5c94b616-8edf-465c-b316-4599df95c52a"
	mscM12WaspOracle      = "002cd9b4-efce-429f-82cd-d37352d2e932"
	mscM12TChakaOracle    = "2a9f2d69-328c-46f4-be37-60b85b197b72"
	mscM12WKabiOracle     = "66950068-adf4-42d7-ad72-fdb4a0e58922"
	mscM12WestCoastOracle = "2e1bf860-22b6-4d08-9b9f-88038cd1dd3f"
	mscM12PhotonOracle    = "c46d6f42-5e50-4c0d-8908-360337285f83"
	mscM12VultureOracle   = "c598d97f-fec6-4238-8eed-8051e9541d31"
)

func TestMscM12EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		mscM12ViperOracle:     "Viper, Cruel Conspirator",
		mscM12ThorAsgardOracl: "Thor, Asgard's Avenger",
		mscM12ThorMidgardOrac: "Thor, Guardian of Midgard",
		mscM12WhiteTigerOracl: "White Tiger, Amulet Keeper",
		mscM12StegronOracle:   "Stegron the Dinosaur Man",
		mscM12WaspOracle:      "Wasp, Shrinking Savior",
		mscM12TChakaOracle:    "T'Chaka, Venerable King",
		mscM12WKabiOracle:     "W'Kabi, Shield of the Nation",
		mscM12WestCoastOracle: "West Coast Expansion",
		mscM12PhotonOracle:    "Photon, Mighty Marvel",
		mscM12VultureOracle:   "Vulture, Feathered Fiend",
	}
	for oracle, name := range want {
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

func mscM12Toughness(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.CurrentToughness()
}

// mscM12RawPower is a battlefield creature's power after the layers settle,
// below zero when it is (hobA4Power clamps at zero).
func mscM12RawPower(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c := effectstest.FindBattlefieldCardByID(g, id)
	if c == nil {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.PowerForComparison()
}

func mscM12SetFlying(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := effectstest.BattlefieldCardPtr(t, g, id)
	c.Keywords = append(c.Keywords, "flying")
}

// --- Viper, Cruel Conspirator ---------------------------------------------------------------------

func mscM12ViperTable(t *testing.T) (g *game.Game, me *game.Player, viper, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	viper = effectstest.PushCatalogPermanent(g, me.ID, "Viper, Cruel Conspirator", "Legendary Creature — Human Assassin Villain", mscM12ViperOracle, false)
	bear = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	return g, me, viper, bear
}

func TestMscM12ViperPumpsTheLoneAttackerForOneBlack(t *testing.T) {
	g, me, viper, bear := mscM12ViperTable(t)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 3 {
		t.Errorf("power %d, want 3", got)
	}
	if got := mscM12Toughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left floating, the {B} should have been spent", me.ManaPool.Total())
	}
}

func TestMscM12ViperNeedsTheManaAndAnAloneAttacker(t *testing.T) {
	g, me, viper, bear := mscM12ViperTable(t)
	if err := g.ActivateCatalogAbility(me.ID, viper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
	wolf := effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Ox", "Creature — Ox", 2, 2)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(idle), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("a creature that is not attacking: err = %v, want ErrIllegalTarget", err)
	}
	_ = wolf
}

func TestMscM12ViperLoneAttackerCheckFailsWithTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	viper := effectstest.PushCatalogPermanent(g, me.ID, "Viper, Cruel Conspirator", "Legendary Creature — Human Assassin Villain", mscM12ViperOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wolf := effectstest.B12Creature(g, me.ID, "My Wolf", "Creature — Wolf", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, wolf)
	effectstest.FloatMana(t, g, me, "{B}{B}")
	for i := 0; i < 2; i++ {
		if err := g.ActivateCatalogAbility(me.ID, viper, i, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
			t.Errorf("ability %d with two attackers: err = %v, want ErrIllegalTarget", i, err)
		}
	}
}

func mscM12ViperChoose(t *testing.T, option int) (g *game.Game, bear uuid.UUID) {
	t.Helper()
	g, me, viper, bear := mscM12ViperTable(t)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestOptionPickFor(g, me.ID)
	if c == nil {
		t.Fatal("no keyword choice was asked")
	}
	if len(c.PickOptions) != 2 {
		t.Fatalf("options %+v, want deathtouch and lifelink", c.PickOptions)
	}
	if err := g.ResolveOptionPick(c.ID, me.ID, option); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	return g, bear
}

func TestMscM12ViperChoosesDeathtouch(t *testing.T) {
	g, bear := mscM12ViperChoose(t, 0)
	got := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.ContainsString(got, "deathtouch") || effectstest.ContainsString(got, "lifelink") {
		t.Errorf("abilities %v, want deathtouch only", got)
	}
}

func TestMscM12ViperChoosesLifelinkAndItEndsWithTheTurn(t *testing.T) {
	g, bear := mscM12ViperChoose(t, 1)
	got := effectstest.EffectiveAbilities(t, g, bear)
	if !effectstest.ContainsString(got, "lifelink") || effectstest.ContainsString(got, "deathtouch") {
		t.Errorf("abilities %v, want lifelink only", got)
	}
	effectstest.PowerUpPassTo(t, g, 1)
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, bear), "lifelink") {
		t.Error("lifelink outlasted the turn")
	}
}

func TestMscM12ViperChoiceDoesNothingIfTheTargetLeaves(t *testing.T) {
	g, me, viper, bear := mscM12ViperTable(t)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, viper, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestOptionPickFor(g, me.ID); c != nil {
		t.Error("a keyword choice was asked for a creature that was gone")
	}
}

// --- Thor, Asgard's Avenger --------------------------------------------------------------------------

func TestMscM12ThorAddsOneToDamageFromAnotherSourceYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Thor, Asgard's Avenger", "Legendary Creature — God Warrior Hero", mscM12ThorAsgardOracl, false)
	source := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 2, 2)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 2) })
	if got := before - opp.Life; got != 3 {
		t.Errorf("2 damage became %d, want 3", got)
	}
	victim := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, victim, 2) })
	if c, ok := effectstest.BattlefieldCard(g, victim); !ok || c.DamageMarked != 3 {
		t.Errorf("victim = %+v, want 3 damage marked", c)
	}
}

func TestMscM12ThorAddsToCombatDamageToo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Thor, Asgard's Avenger", "Legendary Creature — God Warrior Hero", mscM12ThorAsgardOracl, false)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	before := opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	if got := before - opp.Life; got != 3 {
		t.Errorf("combat damage was %d, want 3 (2 plus 1)", got)
	}
}

func TestMscM12ThorDoesNotBoostHimselfOrDamageToYou(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	thor := effectstest.PushCatalogPermanent(g, me.ID, "Thor, Asgard's Avenger", "Legendary Creature — God Warrior Hero", mscM12ThorAsgardOracl, false)
	before := opp.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(thor, opp.ID, 2) })
	if got := before - opp.Life; got != 2 {
		t.Errorf("Thor's own damage was %d, want exactly 2", got)
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pinger", "Creature — Bear", 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Wall", "Creature — Wall", 0, 9)
	other := effectstest.B12Creature(g, me.ID, "My Pinger", "Creature — Bear", 2, 2)
	life := me.Life
	g.WithWriteLock(func() {
		_ = g.DealDamageToPlayerForEffect(theirs, me.ID, 2)
		_ = g.DealDamageToCreatureForEffect(other, mine, 2)
	})
	if me.Life != life-2 {
		t.Errorf("damage to me: life %d -> %d, want 2", life, me.Life)
	}
	if c, ok := effectstest.BattlefieldCard(g, mine); !ok || c.DamageMarked != 2 {
		t.Errorf("my creature = %+v, want 2 damage marked", c)
	}
	third := g.Seats[2]
	tl := third.Life
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(theirs, third.ID, 2) })
	if got := tl - third.Life; got != 2 {
		t.Errorf("an opponent's source hitting another opponent was %d, want 2", got)
	}
}

// --- Thor, Guardian of Midgard -----------------------------------------------------------------------

func mscM12ThorGuardianTable(t *testing.T) (g *game.Game, me, opp *game.Player, source uuid.UUID, top []uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Thor, Guardian of Midgard", "Legendary Creature — God Warrior Hero", mscM12ThorMidgardOrac, false)
	source = effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 2, 2)
	effectstest.PlTop(me, "Deep Card", "Instant", "{U}")
	top = []uuid.UUID{
		effectstest.PlTop(me, "Top Three", "Sorcery", "{R}"),
		effectstest.PlTop(me, "Top Two", "Basic Land — Forest", ""),
		effectstest.PlTop(me, "Top One", "Creature — Bear", "{1}{G}"),
	}
	return g, me, opp, source, top
}

func TestMscM12ThorGuardianExilesThatManyAndTheyArePlayable(t *testing.T) {
	g, me, opp, source, top := mscM12ThorGuardianTable(t)
	lib := me.Library.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 3) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-3 {
		t.Fatalf("library %d -> %d, want three exiled", lib, me.Library.Size())
	}
	for _, id := range top {
		if !effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %s is not exiled and playable by me", id)
		}
	}
	effectstest.PowerUpPassTo(t, g, 1)
	for _, id := range top {
		if effectstest.B12ExiledPlayableBy(t, g, id, me.ID) {
			t.Errorf("card %s is still playable after the turn ended", id)
		}
	}
}

func TestMscM12ThorGuardianExileIsOptional(t *testing.T) {
	g, me, opp, source, _ := mscM12ThorGuardianTable(t)
	lib := me.Library.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(source, opp.ID, 3) })
	effectstest.AssistCardSettle(t, g, me.ID, nil, false)
	if me.Library.Size() != lib {
		t.Errorf("library %d -> %d, want untouched after declining", lib, me.Library.Size())
	}
}

func TestMscM12ThorGuardianCountsASpellsDamageAndIgnoresCombatAndOtherTargets(t *testing.T) {
	g, me, opp, source, _ := mscM12ThorGuardianTable(t)
	lib := me.Library.Size()
	// Combat damage does not trigger it.
	effectstest.AttackWith(t, g, opp.ID, source)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib {
		t.Fatalf("combat damage exiled cards: library %d -> %d", lib, me.Library.Size())
	}
	// Noncombat damage to a creature, or to myself, does not either.
	victim := effectstest.B12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Pinger", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DealDamageToCreatureForEffect(source, victim, 2)
		_ = g.DealDamageToPlayerForEffect(source, me.ID, 1)
		_ = g.DealDamageToPlayerForEffect(theirs, g.Seats[2].ID, 2)
	})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib {
		t.Fatalf("library %d -> %d, want untouched", lib, me.Library.Size())
	}
	// A spell I control is a source I control.
	effectstest.FloatMana(t, g, me, "{R}")
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.BoltOracleCombat, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if me.Library.Size() != lib-3 {
		t.Errorf("library %d -> %d, want the three Bolt damage exiled", lib, me.Library.Size())
	}
}

// --- White Tiger, Amulet Keeper -------------------------------------------------------------------

func mscM12WhiteTigerActivate(t *testing.T, withLand bool) (g *game.Game, me *game.Player, tiger, land uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if withLand {
		land = uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: land, Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	tiger = effectstest.PushCatalogGraveyardCard(me, "White Tiger, Amulet Keeper", "Legendary Creature — Human Hero", mscM12WhiteTigerOracl, 3, 1)
	effectstest.FloatMana(t, g, me, "{G}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tiger, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	return g, me, tiger, land
}

func TestMscM12WhiteTigerDrawsThenPutsALandFromHandOntoTheBattlefield(t *testing.T) {
	g, me, tiger, land := mscM12WhiteTigerActivate(t, true)
	if me.Graveyard.Contains(tiger) {
		t.Error("the card did not leave the graveyard as the cost")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one drawn card", hand, me.Hand.Size())
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil || !effectstest.HasID(pick.ChooseCards, land) {
		t.Fatalf("the land in hand was not offered: %+v", pick)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, land)
	if !g.Battlefield.Contains(land) {
		t.Error("the land did not enter the battlefield")
	}
}

func TestMscM12WhiteTigerLandIsOptionalAndNothingIsAskedWithoutOne(t *testing.T) {
	g, me, _, land := mscM12WhiteTigerActivate(t, true)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if g.Battlefield.Contains(land) || !effectstest.MiInHand(me, land) {
		t.Error("the land moved although the player declined")
	}
	g2, me2, _, _ := mscM12WhiteTigerActivate(t, false)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ChooseCardsChoiceFor(g2, me2.ID) != nil {
		t.Error("a land pick was offered with no land card in hand")
	}
}

func TestMscM12WhiteTigerNeedsFourManaAndTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	tiger := effectstest.PushCatalogGraveyardCard(me, "White Tiger, Amulet Keeper", "Legendary Creature — Human Hero", mscM12WhiteTigerOracl, 3, 1)
	effectstest.FloatMana(t, g, me, "{G}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, tiger, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with three mana")
	}
	if !me.Graveyard.Contains(tiger) {
		t.Error("a refused activation still exiled the card")
	}
	onField := effectstest.PushCatalogPermanent(g, me.ID, "White Tiger, Amulet Keeper", "Legendary Creature — Human Hero", mscM12WhiteTigerOracl, false)
	effectstest.FloatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, onField, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability worked from the battlefield")
	}
}

// --- Stegron the Dinosaur Man ----------------------------------------------------------------------

func mscM12StegronTable(t *testing.T) (g *game.Game, me *game.Player, stegron, bear uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	stegron = effectstest.HandCardFull(me, "Stegron the Dinosaur Man", "Legendary Creature — Dinosaur Villain", "{4}{R}", mscM12StegronOracle, []string{"R"})
	bear = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	return g, me, stegron, bear
}

func TestMscM12StegronDiscardsToPumpAndMakeADinosaurUntilEndOfTurn(t *testing.T) {
	g, me, stegron, bear := mscM12StegronTable(t)
	effectstest.FloatMana(t, g, me, "{R}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stegron, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err != nil {
		t.Fatalf("activate from hand: %v", err)
	}
	if !me.Graveyard.Contains(stegron) {
		t.Error("the discard was not the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.HobA4Power(t, g, bear); got != 5 {
		t.Errorf("power %d, want 5", got)
	}
	if got := mscM12Toughness(t, g, bear); got != 3 {
		t.Errorf("toughness %d, want 3", got)
	}
	subs := effectstest.EffectiveSubtypes(t, g, bear)
	if !effectstest.ContainsString(subs, "Dinosaur") || !effectstest.ContainsString(subs, "Bear") {
		t.Errorf("subtypes %v, want Bear and Dinosaur", subs)
	}
	effectstest.PowerUpPassTo(t, g, 1)
	if got := effectstest.HobA4Power(t, g, bear); got != 2 {
		t.Errorf("power %d after the turn, want 2", got)
	}
	if effectstest.ContainsString(effectstest.EffectiveSubtypes(t, g, bear), "Dinosaur") {
		t.Error("the Dinosaur type outlasted the turn")
	}
}

func TestMscM12StegronOnlyTargetsYourCreaturesAndNeedsTwoMana(t *testing.T) {
	g, me, stegron, bear := mscM12StegronTable(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{C}")
	if err := g.ActivateCatalogAbility(me.ID, stegron, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(theirs), Strict: true}); !errors.Is(err, game.ErrIllegalTarget) {
		t.Errorf("an opposing creature: err = %v, want ErrIllegalTarget", err)
	}
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.ActivateCatalogAbility(me.ID, stegron, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Error("activated with one mana")
	}
	if !effectstest.MiInHand(me, stegron) {
		t.Error("a refused activation still discarded the card")
	}
}

// --- Wasp, Shrinking Savior --------------------------------------------------------------------------

func mscM12WaspAttack(t *testing.T, victimPower int, pick bool) (g *game.Game, me *game.Player, wasp, victim uuid.UUID, hand int) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wasp = effectstest.PushCatalogPermanent(g, me.ID, "Wasp, Shrinking Savior", "Legendary Creature — Human Hero", mscM12WaspOracle, false)
	effectstest.BattlefieldCardPtr(t, g, wasp).Power, effectstest.BattlefieldCardPtr(t, g, wasp).Toughness = 1, 3
	victim = effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", victimPower, 2)
	hand = me.Hand.Size()
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, wasp)
	if pick {
		effectstest.AnswerPickTarget(t, g, victim)
	} else {
		p := effectstest.PickTargetPrompt(t, g)
		if err := g.ResolvePickTargets(p.ID, p.Chooser, nil); err != nil {
			t.Fatalf("ResolvePickTargets: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, wasp, victim, hand
}

func TestMscM12WaspShrinksTheTargetAndDrawsForEachCreatureBelowZero(t *testing.T) {
	g, me, _, victim, hand := mscM12WaspAttack(t, 2, true)
	if got := mscM12RawPower(t, g, victim); got != -1 {
		t.Errorf("power %d, want -1", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card for the one creature below zero", hand, me.Hand.Size())
	}
}

func TestMscM12WaspNoDrawWhenThePowerStaysAtZeroOrMore(t *testing.T) {
	g, me, _, victim, hand := mscM12WaspAttack(t, 3, true)
	if got := effectstest.HobA4Power(t, g, victim); got != 0 {
		t.Errorf("power %d, want 0", got)
	}
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw for a creature at power 0", hand, me.Hand.Size())
	}
}

func TestMscM12WaspCountsEveryCreatureBelowZeroEvenWithNoTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wasp := effectstest.PushCatalogPermanent(g, me.ID, "Wasp, Shrinking Savior", "Legendary Creature — Human Hero", mscM12WaspOracle, false)
	effectstest.B12Creature(g, g.Seats[1].ID, "Shrunken A", "Creature — Bear", -1, 2)
	effectstest.B12Creature(g, me.ID, "Shrunken B", "Creature — Bear", -2, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, wasp)
	p := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(p.ID, p.Chooser, nil); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want two cards for two creatures below zero", hand, me.Hand.Size())
	}
}

func TestMscM12WaspCannotTargetHerselfAndTheShrinkEndsAtYourNextTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wasp := effectstest.PushCatalogPermanent(g, me.ID, "Wasp, Shrinking Savior", "Legendary Creature — Human Hero", mscM12WaspOracle, false)
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, wasp)
	p := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(p.ID, p.Chooser, game.TargetRef{Kind: game.TargetCard, ID: wasp}); err == nil {
		t.Error("Wasp was a legal target for her own trigger")
	}
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PowerUpPassTo(t, g, 1)
	if got := mscM12RawPower(t, g, victim); got != -1 {
		t.Errorf("power %d on the opponent's turn, want it still -1", got)
	}
	effectstest.PowerUpPassTo(t, g, 0)
	if got := effectstest.HobA4Power(t, g, victim); got != 2 {
		t.Errorf("power %d on my next turn, want 2", got)
	}
}

func TestMscM12WaspDoesNothingWhenHerOnlyTargetIsGone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wasp := effectstest.PushCatalogPermanent(g, me.ID, "Wasp, Shrinking Savior", "Legendary Creature — Human Hero", mscM12WaspOracle, false)
	victim := effectstest.B12Creature(g, g.Seats[1].ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, g.Seats[1].ID, "Shrunken", "Creature — Bear", -1, 2)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, g.Seats[1].ID, wasp)
	effectstest.AnswerPickTarget(t, g, victim)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(victim) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw: the ability's only target was illegal", hand, me.Hand.Size())
	}
}

// --- T'Chaka, Venerable King ----------------------------------------------------------------------

func TestMscM12TChakaMillsThreeAndTakesAnArtifactOrLand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PlTop(me, "Deep Card", "Instant", "{U}")
	rock := effectstest.PlTop(me, "Milled Rock", "Artifact", "{2}")
	land := effectstest.PlTop(me, "Milled Land", "Basic Land — Forest", "")
	bear := effectstest.PlTop(me, "Milled Bear", "Creature — Bear", "{1}{G}")
	lib := me.Library.Size()
	effectstest.CastHoldCreature(t, g, "T'Chaka, Venerable King", "Legendary Creature — Human Noble Hero", mscM12TChakaOracle, 2, 2)
	if me.Library.Size() != lib-3 {
		t.Errorf("library %d -> %d, want three milled", lib, me.Library.Size())
	}
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("nothing was offered from among the milled cards")
	}
	if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, rock) || !effectstest.HasID(pick.ChooseCards, land) || effectstest.HasID(pick.ChooseCards, bear) {
		t.Errorf("offered %v, want the milled artifact and land only", pick.ChooseCards)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, rock)
	if !effectstest.MiInHand(me, rock) || effectstest.MiInHand(me, land) || effectstest.MiInHand(me, bear) {
		t.Error("only the chosen artifact should have reached the hand")
	}
}

func TestMscM12TChakaTakingIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PlTop(me, "Milled Land", "Basic Land — Forest", "")
	effectstest.CastHoldCreature(t, g, "T'Chaka, Venerable King", "Legendary Creature — Human Noble Hero", mscM12TChakaOracle, 2, 2)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if effectstest.MiInHand(me, land) || !me.Graveyard.Contains(land) {
		t.Error("the land left the graveyard although the player declined")
	}
}

func mscM12TChakaGraveyard(t *testing.T, owner, controller int) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	if owner >= 0 {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "The Commander", TypeLine: "Legendary Creature — Human",
			Power: 2, Toughness: 2, IsCommander: true,
			Owner: g.Seats[owner].ID, Controller: g.Seats[controller].ID,
		})
	}
	tchaka := effectstest.PushCatalogGraveyardCard(me, "T'Chaka, Venerable King", "Legendary Creature — Human Noble Hero", mscM12TChakaOracle, 2, 2)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}")
	return g, me, tchaka
}

func TestMscM12TChakaBecomesTheMonarchFromTheGraveyardWithYourCommander(t *testing.T) {
	g, me, tchaka := mscM12TChakaGraveyard(t, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, tchaka, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if me.Graveyard.Contains(tchaka) {
		t.Error("the card did not leave the graveyard as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Monarch != me.ID {
		t.Errorf("monarch = %v, want me", g.Monarch)
	}
}

func TestMscM12TChakaNeedsYourOwnCommanderOnTheBattlefield(t *testing.T) {
	for name, tc := range map[string]struct{ owner, controller int }{
		"no commander":                {-1, -1},
		"an opponent's commander":     {1, 1},
		"my commander, stolen":        {0, 1},
		"their commander, I stole it": {1, 0},
	} {
		t.Run(name, func(t *testing.T) {
			g, me, tchaka := mscM12TChakaGraveyard(t, tc.owner, tc.controller)
			effectstest.AcRefused(t, g, me.ID, tchaka, 0, game.ActivateAbilityParams{Strict: true})
			if !me.Graveyard.Contains(tchaka) {
				t.Error("a refused activation still exiled the card")
			}
		})
	}
}

// --- W'Kabi, Shield of the Nation ------------------------------------------------------------------

func mscM12WKabiTable(t *testing.T, artifactCost string, commander bool) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "W'Kabi, Shield of the Nation", "Legendary Creature — Human Warrior Hero", mscM12WKabiOracle, false)
	if artifactCost != "" {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "Big Rock", TypeLine: "Artifact", ManaCost: artifactCost, Owner: me.ID, Controller: me.ID})
	}
	attacker := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Human",
		Power: 2, Toughness: 2, IsCommander: commander, Owner: me.ID, Controller: me.ID,
	})
	return g, me, attacker
}

func TestMscM12WKabiMakesARhinoWhenYourCommanderAttacksWithABigArtifact(t *testing.T) {
	g, me, cmdr := mscM12WKabiTable(t, "{4}", true)
	effectstest.AttackWith(t, g, g.Seats[1].ID, cmdr)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.NamedOnBattlefield(g, "Rhino", me.ID); got != 1 {
		t.Fatalf("%d Rhino tokens, want 1", got)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rhino" && (c.Power != 4 || c.Toughness != 4) {
			t.Errorf("Rhino is %d/%d, want 4/4", c.Power, c.Toughness)
		}
	}
}

func TestMscM12WKabiNeedsTheArtifactAndTheCommander(t *testing.T) {
	for name, tc := range map[string]struct {
		cost      string
		commander bool
	}{
		"no artifact":          {"", true},
		"artifact worth three": {"{3}", true},
		"not the commander":    {"{4}", false},
	} {
		t.Run(name, func(t *testing.T) {
			g, me, atk := mscM12WKabiTable(t, tc.cost, tc.commander)
			effectstest.AttackWith(t, g, g.Seats[1].ID, atk)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.NamedOnBattlefield(g, "Rhino", me.ID); got != 0 {
				t.Errorf("%d Rhino tokens, want none", got)
			}
		})
	}
}

func TestMscM12WKabiDoesNotCountAnOpponentsArtifactOrABigNonartifact(t *testing.T) {
	g, me, cmdr := mscM12WKabiTable(t, "", true)
	effectstest.PushPermanent(g, g.Seats[1].ID, game.Card{Name: "Their Rock", TypeLine: "Artifact", ManaCost: "{5}", Owner: g.Seats[1].ID, Controller: g.Seats[1].ID})
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Big Enchantment", TypeLine: "Enchantment", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})
	effectstest.AttackWith(t, g, g.Seats[1].ID, cmdr)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.NamedOnBattlefield(g, "Rhino", me.ID); got != 0 {
		t.Errorf("%d Rhino tokens, want none", got)
	}
}

// --- West Coast Expansion -------------------------------------------------------------------------

func mscM12WestCoast(t *testing.T, x int) (g *game.Game, me *game.Player, hero, other uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	hero = effectstest.HandCardFull(me, "Test Hero", "Creature — Human Hero", "{3}{R}", "", []string{"R"})
	other = effectstest.HandCardFull(me, "Test Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.B19CastXModal(t, g, "West Coast Expansion", "Sorcery", mscM12WestCoastOracle, "{X}{U}{U}", x, nil, nil)
	return g, me, hero, other
}

func TestMscM12WestCoastExpansionDrawsXAndOffersAFreeHeroAtFive(t *testing.T) {
	g, me, hero, other := mscM12WestCoast(t, 5)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+5 {
		t.Errorf("hand %d -> %d, want five drawn", hand, me.Hand.Size())
	}
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	if !effectstest.HasID(offer.DiscardOptions, hero) || effectstest.HasID(offer.DiscardOptions, other) {
		t.Fatalf("the offer is %v, want the Hero only", offer.DiscardOptions)
	}
	if err := g.CastSpell(me.ID, hero, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(hero) {
		t.Error("the Hero did not enter the battlefield")
	}
}

func TestMscM12WestCoastExpansionOffersNothingBelowFiveOrWithNoHero(t *testing.T) {
	g, me, _, _ := mscM12WestCoast(t, 4)
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+4 {
		t.Errorf("hand %d -> %d, want four drawn", hand, me.Hand.Size())
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered at X = 4")
	}
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	effectstest.HandCardFull(me2, "Test Bear", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.B19CastXModal(t, g2, "West Coast Expansion", "Sorcery", mscM12WestCoastOracle, "{X}{U}{U}", 5, nil, nil)
	effectstest.PassPriorityAroundTable(t, g2)
	if effectstest.ResolutionCastOfferFor(g2, me2.ID) != nil {
		t.Error("a free cast was offered with no Hero spell in hand")
	}
}

func TestMscM12WestCoastExpansionDeclineKeepsTheHeroInHand(t *testing.T) {
	g, me, hero, _ := mscM12WestCoast(t, 5)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	if !effectstest.MiInHand(me, hero) {
		t.Error("the Hero left the hand although the player declined")
	}
}

// --- Photon, Mighty Marvel ----------------------------------------------------------------------------

func TestMscM12PhotonAddsThatMuchManaOfOneColorKeptThroughTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	photon := effectstest.PushCatalogPermanent(g, me.ID, "Photon, Mighty Marvel", "Legendary Creature — Human Hero", mscM12PhotonOracle, false)
	effectstest.BattlefieldCardPtr(t, g, photon).Power, effectstest.BattlefieldCardPtr(t, g, photon).Toughness = 2, 4
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(photon, game.CounterPlusOne, 1) })
	effectstest.AttackWith(t, g, opp.ID, photon)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	colors := effectstest.PoolColors(me)
	if len(colors) != 3 {
		t.Fatalf("pool %v, want three mana (Photon is a 2/4 with a counter)", colors)
	}
	for _, c := range colors {
		if c != "U" {
			t.Errorf("pool %v, want only blue", colors)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	if me.ManaPool.Total() != 3 {
		t.Errorf("pool %d at the end step, want the mana kept", me.ManaPool.Total())
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d after the cleanup step, want it empty", me.ManaPool.Total())
	}
}

func TestMscM12PhotonOnlyTriggersOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	photon := effectstest.PushCatalogPermanent(g, me.ID, "Photon, Mighty Marvel", "Legendary Creature — Human Hero", mscM12PhotonOracle, false)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(photon, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceColor) != nil || me.ManaPool.Total() != 0 {
		t.Error("noncombat damage produced mana")
	}
}

// --- Vulture, Feathered Fiend -------------------------------------------------------------------------

func mscM12VultureTable(t *testing.T) (g *game.Game, me *game.Player, vulture, flyer, ground uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vulture = effectstest.PushCatalogPermanent(g, me.ID, "Vulture, Feathered Fiend", "Legendary Creature — Human Artificer Villain", mscM12VultureOracle, false)
	effectstest.BattlefieldCardPtr(t, g, vulture).Power, effectstest.BattlefieldCardPtr(t, g, vulture).Toughness = 2, 4
	mscM12SetFlying(t, g, vulture)
	flyer = effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 1)
	mscM12SetFlying(t, g, flyer)
	ground = effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	return g, me, vulture, flyer, ground
}

func TestMscM12VultureCountersEachFlyingDamagerOnceAndDrawsOnce(t *testing.T) {
	g, me, vulture, flyer, ground := mscM12VultureTable(t)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, g.Seats[1].ID, vulture, flyer, ground)
	effectstest.MscM12Settle(t, g)
	if got := effectstest.HobA4Power(t, g, vulture); got != 3 {
		t.Errorf("Vulture power %d, want 3 (one counter)", got)
	}
	if got := effectstest.HobA4Power(t, g, flyer); got != 2 {
		t.Errorf("Bird power %d, want 2 (one counter)", got)
	}
	if got := effectstest.HobA4Power(t, g, ground); got != 2 {
		t.Errorf("Bear power %d, want 2 (no counter: it has no flying)", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want exactly one card for the batch", hand, me.Hand.Size())
	}
}

func TestMscM12VultureTriggersOncePerPlayerHit(t *testing.T) {
	g, me, vulture, flyer, _ := mscM12VultureTable(t)
	hand := me.Hand.Size()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(vulture, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(flyer, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.MscM12Settle(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want a card for each player hit", hand, me.Hand.Size())
	}
	if got := effectstest.HobA4Power(t, g, vulture); got != 3 {
		t.Errorf("Vulture power %d, want 3 (only the creatures that hit that player)", got)
	}
	if got := effectstest.HobA4Power(t, g, flyer); got != 2 {
		t.Errorf("Bird power %d, want 2", got)
	}
}

func TestMscM12VultureIgnoresGroundDamageAndNoncombatDamage(t *testing.T) {
	g, me, _, _, ground := mscM12VultureTable(t)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, g.Seats[1].ID, ground)
	effectstest.MscM12Settle(t, g)
	flyer := effectstest.B12Creature(g, me.ID, "Second Bird", "Creature — Bird", 1, 1)
	mscM12SetFlying(t, g, flyer)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(flyer, g.Seats[1].ID, 1) })
	effectstest.MscM12Settle(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d -> %d, want no draw", hand, me.Hand.Size())
	}
	if got := effectstest.HobA4Power(t, g, flyer); got != 1 {
		t.Errorf("Bird power %d, want 1", got)
	}
}

func TestMscM12VultureGivesNoCounterToACreatureThatLeftButStillDraws(t *testing.T) {
	g, me, vulture, flyer, _ := mscM12VultureTable(t)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, g.Seats[1].ID, vulture, flyer)
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(flyer) })
	effectstest.MscM12Settle(t, g)
	if got := effectstest.HobA4Power(t, g, vulture); got != 3 {
		t.Errorf("Vulture power %d, want 3", got)
	}
	if g.Battlefield.Contains(flyer) {
		t.Fatal("the Bird is still on the battlefield")
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d -> %d, want the draw plus the bounced Bird", hand, me.Hand.Size())
	}
}

func TestMscM12VultureCounterGoesToTheCreatureThatHadFlyingWhenItDealtDamage(t *testing.T) {
	g, me, _, flyer, ground := mscM12VultureTable(t)
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, g.Seats[1].ID, flyer, ground)
	// The triggers are on the stack; the Bird loses flying in response.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(uuid.Nil, g.PinnedObjectsLocked(flyer),
			[]game.Mod{game.RemoveKeywordsMod("flying")}, g.UntilEndOfTurnDuration(), "test — loses flying")
	})
	if effectstest.ContainsString(effectstest.EffectiveAbilities(t, g, flyer), "flying") {
		t.Fatal("the Bird still has flying")
	}
	effectstest.MscM12Settle(t, g)
	if got := effectstest.HobA4Power(t, g, flyer); got != 2 {
		t.Errorf("Bird power %d, want 2: it had flying when it dealt the damage", got)
	}
	if got := effectstest.HobA4Power(t, g, ground); got != 2 {
		t.Errorf("Bear power %d, want 2: it never had flying", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want one card", hand, me.Hand.Size())
	}
}
