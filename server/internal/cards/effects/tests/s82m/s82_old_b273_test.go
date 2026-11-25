package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b273_test.go — Conspiracy: Take the Crown, S82 older-sets batch
// 273 (#5540): activated abilities, removal, equipment, life gain and loss,
// pump and combat tricks, other creatures. Each test plays the real catalog
// entry.

const (
	oldB273BrigadeOracle   = "10bc98b0-3fdc-46d1-8d3b-6d160e9dd62f"
	oldB273ChimeraOracle   = "fccc57f6-2555-4ab3-b85a-83b524abe892"
	oldB273AvatarOracle    = "2da5aa33-3794-4f83-936c-9d8640833d86"
	oldB273ExecutionOracle = "6125f993-970c-469a-b557-23d4dc6684f4"
	oldB273SpyKitOracle    = "4d94f5bd-42c6-4d15-a693-8be189ec3746"
	oldB273MatrixOracle    = "dc748941-fd91-4a0e-b0c4-04c9092ff158"
	oldB273GladiatorOracle = "7d8e20e5-6456-43f5-9188-263b54c0a923"
	oldB273SycophantOracle = "8301de9f-3082-400f-92d4-553536f79f49"
	oldB273SingerOracle    = "8aae6f84-02ac-4f8f-9721-b12138588b6b"
	oldB273ChantOracle     = "a9edb891-1a38-4b8a-8539-c459e8a52315"
	oldB273ManaplasmOracle = "9061e2ae-7385-494a-a511-0b6531d1fe0e"
	oldB273HoplditeOracle  = "e15db59f-4477-4139-8e7e-741ab4584e7f"
)

func TestOldB273EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB273BrigadeOracle: "Goblin Balloon Brigade", oldB273ChimeraOracle: "Stormchaser Chimera", oldB273AvatarOracle: "Avatar of Woe",
		oldB273ExecutionOracle: "Public Execution", oldB273SpyKitOracle: "Spy Kit", oldB273MatrixOracle: "Hedron Matrix",
		oldB273GladiatorOracle: "Carnage Gladiator", oldB273SycophantOracle: "Garrulous Sycophant", oldB273SingerOracle: "Shipwreck Singer",
		oldB273ChantOracle: "Gruul War Chant", oldB273ManaplasmOracle: "Manaplasm", oldB273HoplditeOracle: "Akroan Hoplite",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered=%v name=%q completeness=%s caveats=%v", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB273Enter puts a catalog creature into owner's hand and moves it onto
// the battlefield through the real zone move, then clears its summoning
// sickness so it can attack this turn.
func oldB273Enter(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := effectstest.PlayerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner}, game.ZoneRef{Kind: game.ZoneBattlefield}, id); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	return id
}

// oldB273Equip enters an Equipment from the hand and attaches it to host.
func oldB273Equip(t *testing.T, g *game.Game, owner uuid.UUID, name, oracle string, host uuid.UUID) uuid.UUID {
	t.Helper()
	kit := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Artifact — Equipment",
		OracleID: oracle, Owner: owner, Controller: owner})
	effectstest.AttachForTest(t, g, kit, host)
	return kit
}

// --- Goblin Balloon Brigade -----------------------------------------------

func TestOldB273GoblinBalloonBrigadeGainsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	goblin := oldB273Enter(t, g, me.ID, "Goblin Balloon Brigade", "Creature — Goblin Warrior", oldB273BrigadeOracle, "{R}", 1, 1)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, goblin), "flying") {
		t.Fatal("flying before the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, goblin, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, goblin), "flying") {
		t.Error("no flying after the ability resolved")
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, goblin), "flying") {
		t.Error("flying outlasted the turn")
	}
}

// --- Stormchaser Chimera --------------------------------------------------

func TestOldB273StormchaserChimeraGetsTheRevealedCardsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chimera := oldB273Enter(t, g, me.ID, "Stormchaser Chimera", "Creature — Chimera", oldB273ChimeraOracle, "{2}{U}{R}", 2, 3)
	effectstest.P8LibraryTop(g, me, game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, chimera, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, chimera); got != 5 {
		t.Errorf("power %d, want 5 (2 + the revealed card's mana value 3)", got)
	}
}

// --- Avatar of Woe --------------------------------------------------------

func oldB273FillGraveyards(g *game.Game, n int) {
	for i := 0; i < n; i++ {
		effectstest.B17GraveyardCard(g.Seats[i%2], "Corpse", "Creature — Bear", "{1}")
	}
}

func oldB273CastAvatar(t *testing.T, g *game.Game) error {
	t.Helper()
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "B", "B")
	id := effectstest.HandCardFull(me, "Avatar of Woe", "Creature — Avatar", "{6}{B}{B}", oldB273AvatarOracle, []string{"B"})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
}

func TestOldB273AvatarOfWoeCostsSixLessWithTenCreatureCardsInGraveyards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB273FillGraveyards(g, 10)
	if err := oldB273CastAvatar(t, g); err != nil {
		t.Errorf("cast for {B}{B} with ten creature cards in graveyards: %v", err)
	}
}

func TestOldB273AvatarOfWoeIsFullPriceWithNineCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB273FillGraveyards(g, 9)
	// A noncreature card does not count toward the ten.
	effectstest.B17GraveyardCard(g.Seats[0], "Bolt", "Instant", "{R}")
	if err := oldB273CastAvatar(t, g); err == nil {
		t.Error("cast for {B}{B} with only nine creature cards in graveyards")
	}
}

func TestOldB273AvatarOfWoeHasFearAndDestroysATargetCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	avatar := oldB273Enter(t, g, me.ID, "Avatar of Woe", "Creature — Avatar", oldB273AvatarOracle, "{6}{B}{B}", 6, 5)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, avatar), "fear") {
		t.Error("no fear")
	}
	if err := g.ActivateCatalogAbility(me.ID, avatar, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the target creature survived")
	}
	if !effectstest.E2Card(t, g, avatar).Tapped {
		t.Error("the ability did not tap the Avatar")
	}
}

// --- Public Execution -----------------------------------------------------

func TestOldB273PublicExecutionDestroysAndShrinksTheirOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := effectstest.B12Creature(g, opp.ID, "Target", "Creature — Ogre", 3, 3)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Ogre", 4, 4)
	third := effectstest.B12Creature(g, g.Seats[2].ID, "Bystander", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 4, 4)
	effectstest.CastCatalogSpell(t, g, "Public Execution", "Instant", oldB273ExecutionOracle, effectstest.CardRefs(target))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(target) {
		t.Error("the target survived")
	}
	for id, want := range map[uuid.UUID]int{other: 2, third: 4, mine: 4} {
		if got := effectstest.EffectivePower(t, g, id); got != want {
			t.Errorf("power %d, want %d", got, want)
		}
	}
}

func TestOldB273PublicExecutionRefusesMyOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := effectstest.B12Creature(g, g.Seats[0].ID, "Mine", "Creature — Ogre", 4, 4)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Public Execution", "Instant", oldB273ExecutionOracle, effectstest.CardRefs(mine)); err == nil {
		t.Error("Public Execution targeted a creature I control")
	}
}

// --- Spy Kit --------------------------------------------------------------

func TestOldB273SpyKitPumpsAndGivesAllNonlegendaryCreatureNames(t *testing.T) {
	prev := game.NonlegendaryCreatureCardName
	t.Cleanup(func() { game.NonlegendaryCreatureCardName = prev })
	game.NonlegendaryCreatureCardName = func(name string) bool { return name == "Grizzly Bears" }
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	needle := effectstest.B21Push(g, them.ID, "Pithing Needle", "Artifact", effectstest.ActPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(them.ID, needle, "Pithing Needle — choose a card name")
	})
	effectstest.ActNameFor(t, g, them.ID, "Grizzly Bears")

	krenko := effectstest.ActPushKrenko(g, me.ID)
	oldB273Equip(t, g, me.ID, "Spy Kit", oldB273SpyKitOracle, krenko)
	if got := effectstest.EffectivePower(t, g, krenko); got != 4 {
		t.Errorf("equipped power %d, want 4 (3 + 1)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, krenko, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the equipped creature with Grizzly Bears among its names: err = %v, want ErrCantActivate", err)
	}
	plain := effectstest.ActPushKrenko(g, them.ID)
	if err := g.ActivateCatalogAbility(them.ID, plain, 0, game.ActivateAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("an unequipped Krenko: err = %v, want the Needle not to stop it", err)
	}
}

// --- Hedron Matrix --------------------------------------------------------

func TestOldB273HedronMatrixGivesTheEquippedCreatureItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	four := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Four Drop", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Token", TypeLine: "Creature — Bear",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	oldB273Equip(t, g, me.ID, "Hedron Matrix", oldB273MatrixOracle, four)
	if got := effectstest.EffectivePower(t, g, four); got != 6 {
		t.Errorf("equipped 2/2 with mana value 4 has power %d, want 6", got)
	}
	if got := effectstest.E2Card(t, g, four).CurrentToughness(); got != 6 {
		t.Errorf("toughness %d, want 6", got)
	}
	// A creature with no mana cost has mana value 0: nothing is added.
	oldB273Equip(t, g, me.ID, "Hedron Matrix", oldB273MatrixOracle, token)
	if got := effectstest.EffectivePower(t, g, token); got != 1 {
		t.Errorf("a mana value 0 creature has power %d, want 1", got)
	}
}

// --- Carnage Gladiator ----------------------------------------------------

func TestOldB273CarnageGladiatorMakesAnyBlockersControllerLoseALife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	oldB273Enter(t, g, me.ID, "Carnage Gladiator", "Creature — Skeleton Warrior", oldB273GladiatorOracle, "{2}{B}{R}", 4, 2)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Wall", 0, 4)
	life := opp.Life
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 1 {
		t.Errorf("the blocker's controller lost %d life, want 1", got)
	}
}

func TestOldB273CarnageGladiatorRegeneratesForOneBR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	gladiator := oldB273Enter(t, g, me.ID, "Carnage Gladiator", "Creature — Skeleton Warrior", oldB273GladiatorOracle, "{2}{B}{R}", 4, 2)
	if err := g.ActivateCatalogAbility(me.ID, gladiator, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := (effects.DestroyTarget{Target: gladiator}).Apply(effects.NewContext(g, nil)); err != nil {
			t.Errorf("destroy: %v", err)
		}
	})
	if !g.Battlefield.Contains(gladiator) {
		t.Error("the regeneration shield did not save the Gladiator")
	}
}

// --- Garrulous Sycophant --------------------------------------------------

func TestOldB273GarrulousSycophantDrainsOnlyForTheMonarch(t *testing.T) {
	for _, monarch := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		oldB273Enter(t, g, me.ID, "Garrulous Sycophant", "Creature — Human Advisor", oldB273SycophantOracle, "{2}{B}", 1, 4)
		if monarch {
			effectstest.MonCrown(t, g, me.ID)
		}
		lives := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			lives[p.ID] = p.Life
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) // the crown's own draw trigger beside the Sycophant's
		effectstest.PassPriorityAroundTable(t, g)
		for i, p := range g.Seats {
			want := lives[p.ID]
			if monarch && i == 0 {
				want++
			} else if monarch {
				want--
			}
			if p.Life != want {
				t.Errorf("monarch=%v seat %d life %d, want %d", monarch, i, p.Life, want)
			}
		}
	}
}

// --- Shipwreck Singer -----------------------------------------------------

func TestOldB273ShipwreckSingerForcesTheTargetToAttackThisTurnIfAble(t *testing.T) {
	for _, forced := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		singer := oldB273Enter(t, g, me.ID, "Shipwreck Singer", "Creature — Siren", oldB273SingerOracle, "{U}{B}", 1, 2)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Ogre", 3, 3)
		effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
		if forced {
			if err := g.ActivateCatalogAbility(me.ID, singer, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err != nil {
				t.Fatalf("activate: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
		_, err := g.AdvanceStep()
		if forced && err == nil {
			t.Error("the opponent's creature was allowed to skip attacking")
		}
		if !forced && err != nil {
			t.Errorf("without the ability the creature was forced to attack: %v", err)
		}
	}
}

func TestOldB273ShipwreckSingerRefusesMyOwnCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	singer := oldB273Enter(t, g, me.ID, "Shipwreck Singer", "Creature — Siren", oldB273SingerOracle, "{U}{B}", 1, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Ogre", 3, 3)
	if err := g.ActivateCatalogAbility(me.ID, singer, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(mine)}); err == nil {
		t.Error("the ability targeted a creature I control")
	}
}

func TestOldB273ShipwreckSingerShrinksEveryAttackingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	singer := oldB273Enter(t, g, me.ID, "Shipwreck Singer", "Creature — Siren", oldB273SingerOracle, "{U}{B}", 1, 2)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 3, 3)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 3, 3)
	effectstest.Pr7bAttack(t, g, opp.ID, a)
	if err := g.ActivateCatalogAbility(me.ID, singer, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, a); got != 2 {
		t.Errorf("an attacker has power %d, want 2", got)
	}
	if got := effectstest.EffectivePower(t, g, idle); got != 3 {
		t.Errorf("a creature that is not attacking has power %d, want 3", got)
	}
}

// --- Gruul War Chant ------------------------------------------------------

func TestOldB273GruulWarChantPumpsAndGivesMenaceToYourAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Gruul War Chant", TypeLine: "Enchantment",
		OracleID: oldB273ChantOracle, Owner: me.ID, Controller: me.ID})
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if effectstest.HasString(effectstest.EffectiveAbilities(t, g, attacker), "menace") || effectstest.EffectivePower(t, g, attacker) != 2 {
		t.Fatal("a creature that is not attacking is already pumped")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, attacker)
	if got := effectstest.EffectivePower(t, g, attacker); got != 3 {
		t.Errorf("the attacker has power %d, want 3", got)
	}
	if !effectstest.HasString(effectstest.EffectiveAbilities(t, g, attacker), "menace") {
		t.Error("the attacker has no menace")
	}
	if effectstest.EffectivePower(t, g, idle) != 2 || effectstest.HasString(effectstest.EffectiveAbilities(t, g, idle), "menace") {
		t.Error("a creature of mine that is not attacking was affected")
	}
	if effectstest.EffectivePower(t, g, theirs) != 2 {
		t.Error("an opponent's creature was affected")
	}
}

// --- Manaplasm ------------------------------------------------------------

func TestOldB273ManaplasmGrowsByEachSpellsManaValueUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	plasm := oldB273Enter(t, g, me.ID, "Manaplasm", "Creature — Ooze", oldB273ManaplasmOracle, "{2}{G}", 1, 1)
	effectstest.ToMainForCost(t, g)
	effectstest.B10AddMana(me, "G", "G", "G")
	id := effectstest.HandCardFull(me, "Growth", "Sorcery", "{1}{G}", "", []string{"G"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, plasm); got != 3 {
		t.Errorf("power %d after a mana value 2 spell, want 3", got)
	}
	effectstest.AdvanceToStepOf(t, g, 1, game.StepPrecombatMain)
	if got := effectstest.EffectivePower(t, g, plasm); got != 1 {
		t.Errorf("power %d next turn, want 1", got)
	}
}

// --- Akroan Hoplite -------------------------------------------------------

func TestOldB273AkroanHoplitePumpsByTheNumberOfAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hoplite := oldB273Enter(t, g, me.ID, "Akroan Hoplite", "Creature — Human Soldier", oldB273HoplditeOracle, "{R}{W}", 1, 2)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, hoplite, a, b)
	if got := effectstest.EffectivePower(t, g, hoplite); got != 4 {
		t.Errorf("Hoplite power %d with three attackers, want 4 (1 + 3)", got)
	}
}
