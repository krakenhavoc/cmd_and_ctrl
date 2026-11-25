package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b206_test.go — Battlebond, batch 206 (#5529): each test plays the
// real catalog entry through the real cast, activation and trigger paths.

const (
	oldB206WhimOracle     = "01c2b66a-66d1-4282-a078-b1fa737c57f4"
	oldB206BonusOracle    = "40ab2b22-9cf6-4a73-a05f-6ce496d10bf7"
	oldB206ManeuverOracle = "1c2d6b55-653a-41fd-8a5e-6463a992640b"
	oldB206MaggotOracle   = "67d0df0f-37d1-46da-b1c3-4fc9d1e3cc05"
	oldB206InsightOracle  = "3f1f27ca-15e6-4ac6-8bd3-1e6b3ca00857"
	oldB206TalonsOracle   = "4a2a7eca-3477-4858-9dab-ba80d4d704cd"
	oldB206FanaticOracle  = "18543bb3-14f9-45a0-8ed3-5f11832527d9"
	oldB206TowerOracle    = "49b86b51-e1ef-4e27-84d2-66166481e0a5"
	oldB206DrakeOracle    = "57d93683-8739-49ef-856a-5a70f92d8763"
	oldB206FumbleOracle   = "cbf68ddd-596f-4b2d-b93c-3b7976ba4d37"
	oldB206SphinxOracle   = "d23b514f-3081-45da-a2cf-7ed3cf49dc24"
	oldB206TwinOracle     = "da17da9b-f300-4613-978b-ba3026d4e275"
	oldB206RampartOracle  = "ec187304-4c05-44da-8ad5-7c56f6b05212"
)

func TestOldB206CardsAreRegisteredComplete(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB206WhimOracle:     "Pir's Whim",
		oldB206BonusOracle:    "Bonus Round",
		oldB206ManeuverOracle: "Virtus's Maneuver",
		oldB206MaggotOracle:   "Rotfeaster Maggot",
		oldB206InsightOracle:  "Oracle's Insight",
		oldB206TalonsOracle:   "Lightning Talons",
		oldB206FanaticOracle:  "Cheering Fanatic",
		oldB206TowerOracle:    "Sentinel Tower",
		oldB206DrakeOracle:    "Urborg Drake",
		oldB206FumbleOracle:   "Fumble",
		oldB206SphinxOracle:   "Jelenn Sphinx",
		oldB206TwinOracle:     "Evil Twin",
		oldB206RampartOracle:  "Battle Rampart",
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

func oldB206Target(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

// oldB206Resolve casts the spell for the active seat and answers friend or foe
// for each seat in turn order: true is a friend.
func oldB206FriendOrFoe(t *testing.T, g *game.Game, name, typeLine, oracle string, friends ...bool) {
	t.Helper()
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, name, typeLine, oracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, friend := range friends {
		idx := 1
		if friend {
			idx = 0
		}
		effectstest.AnswerOptionPick(t, g, me.ID, idx)
	}
}

// --- Pir's Whim ---

func TestOldB206PirsWhimFriendsFetchALandAndFoesSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	forest := effectstest.PushLibraryCard(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	rock := effectstest.B12Push(g, foe.ID, "Rock", "Artifact", "", 0, 0)
	keep := effectstest.B12Creature(g, foe.ID, "Keeper", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oldB206FriendOrFoe(t, g, "Pir's Whim", "Sorcery", oldB206WhimOracle, true, false, false, false)
	open := effectstest.SearchPromptFor(g, me.ID)
	if open == nil {
		t.Fatal("the friend was not offered a search")
	}
	if err := g.ResolveSearchLibrary(open.ID, me.ID, []uuid.UUID{forest}); err != nil {
		t.Fatalf("search: %v", err)
	}
	effectstest.SlChoose(t, g, foe.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(forest) {
		t.Error("the friend's land was not put onto the battlefield")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the foe kept its artifact")
	}
	if !g.Battlefield.Contains(keep) {
		t.Error("the foe lost a creature, not an artifact or enchantment")
	}
}

// --- Virtus's Maneuver ---

func TestOldB206VirtussManeuverFriendsReturnAndFoesSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	dead := effectstest.GyCard(me, "Dead Bear", "Creature — Bear", "", "")
	victim := effectstest.B12Creature(g, foe.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	oldB206FriendOrFoe(t, g, "Virtus's Maneuver", "Sorcery", oldB206ManeuverOracle, true, false, true, true)
	effectstest.SlChoose(t, g, me.ID, dead)
	effectstest.SlChoose(t, g, foe.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(dead) {
		t.Error("the friend's creature card was not returned to hand")
	}
	if g.Battlefield.Contains(victim) {
		t.Error("the foe's only creature was not sacrificed")
	}
}

// --- Rotfeaster Maggot ---

func TestOldB206RotfeasterMaggotExilesAndGainsToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.GyCard(opp, "Big Wall", "Creature — Wall", "", "")
	g.WithWriteLock(func() {
		for i := range opp.Graveyard.Cards {
			if opp.Graveyard.Cards[i].InstanceID == dead {
				opp.Graveyard.Cards[i].Power, opp.Graveyard.Cards[i].Toughness = 0, 6
			}
		}
	})
	life := me.Life
	effectstest.EnterCard(t, g, me.ID, game.Card{Name: "Rotfeaster Maggot", TypeLine: "Creature — Insect", OracleID: oldB206MaggotOracle, Power: 3, Toughness: 5})
	effectstest.AnswerPickTarget(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(dead) {
		t.Error("the creature card is still in the graveyard")
	}
	if me.Life != life+6 {
		t.Errorf("life %d -> %d, want +6 (the card's toughness)", life, me.Life)
	}
}

// --- Oracle's Insight ---

func TestOldB206OraclesInsightGivesScryThenDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Oracle's Insight", oldB206InsightOracle, bear)
	hand := len(me.Hand.Cards)
	effectstest.AuraSeedLibrary(me, 3)
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, len(me.Hand.Cards))
	}
}

// --- Lightning Talons ---

func TestOldB206LightningTalonsPumpsAndGivesFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.AuraBear(g, me.ID)
	effectstest.AuraCast(t, g, "Lightning Talons", oldB206TalonsOracle, bear)
	if p, tt := effectstest.PtOf(t, g, bear); p != 5 || tt != 2 {
		t.Errorf("the enchanted creature is %d/%d, want 5/2", p, tt)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the enchanted creature lacks first strike")
	}
}

// --- Urborg Drake ---

func TestOldB206UrborgDrakeHasFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drake := effectstest.PushCatalogPermanent(g, me.ID, "Urborg Drake", "Creature — Drake", oldB206DrakeOracle, false)
	if !effectstest.HasEffectiveKeyword(t, g, drake, "flying") {
		t.Error("the Drake lacks flying")
	}
}

// --- Jelenn Sphinx ---

func TestOldB206JelennSphinxPumpsOtherAttackersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := effectstest.B12Push(g, me.ID, "Jelenn Sphinx", "Creature — Sphinx", oldB206SphinxOracle, 1, 5)
	ally := effectstest.B12Creature(g, me.ID, "Ally Bear", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{sphinx, ally} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, ally); p != 3 || tt != 3 {
		t.Errorf("the other attacker is %d/%d, want 3/3", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, sphinx); p != 1 || tt != 5 {
		t.Errorf("the Sphinx is %d/%d, want 1/5", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, idle); p != 2 || tt != 2 {
		t.Errorf("the creature that did not attack is %d/%d, want 2/2", p, tt)
	}
}

// --- Battle Rampart ---

func TestOldB206BattleRampartGivesHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.PushCatalogPermanent(g, me.ID, "Battle Rampart", "Creature — Wall", oldB206RampartOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Fatal("the Bear has haste already")
	}
	if err := g.ActivateCatalogAbility(me.ID, wall, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB206Target(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the Bear did not gain haste")
	}
}

// --- Sentinel Tower ---

func TestOldB206SentinelTowerCountsEarlierSpellsThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Sentinel Tower", "Artifact", oldB206TowerOracle, false)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("first spell: life %d -> %d, want 1 damage", life, opp.Life)
	}
	effectstest.CastCatalogSpell(t, g, "Opt Two", "Sorcery", "", nil)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1-2 {
		t.Errorf("second spell: life is %d, want 2 more damage", opp.Life)
	}
}

func TestOldB206SentinelTowerIgnoresAnOpponentsTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Sentinel Tower", "Artifact", oldB206TowerOracle, false)
	effectstest.AdvanceToMainOf(t, g, 1)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Opt", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || len(g.PendingChoices) != 0 {
		t.Error("the Tower triggered on an opponent's turn")
	}
}

// --- Bonus Round ---

func TestOldB206BonusRoundCopiesEveryInstantOrSorceryThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, victim := g.Seats[0].ID, g.Seats[1].ID
	before := effectstest.LifeOf(g, victim)
	effectstest.CastCatalogSpell(t, g, "Bonus Round", "Sorcery", oldB206BonusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	effectstest.CastBoltAtAndKeepTheCopysTarget(t, g, me, victim, 1)
	if got := effectstest.LifeOf(g, victim); got != before-12 {
		t.Errorf("victim life = %d, want %d (each Bolt and its copy)", got, before-12)
	}
}

func TestOldB206BonusRoundLetsAnOpponentCopyTheirOwnSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Bonus Round", "Sorcery", oldB206BonusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	before := me.Life
	effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	asked := false
	for i := 0; i < 48; i++ {
		if p := effectstest.TriggerOrderPromptFor(g, me.ID); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, me.ID, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
			t.Fatal("Bonus Round's controller was asked about the opponent's copy")
		}
		if p := effectstest.LatestPickTarget(g, opp.ID); p != nil {
			asked = true
			if err := g.ResolvePickTarget(p.ID, opp.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if effectstest.StackFullyEmpty(g) {
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !asked {
		t.Error("the opponent, who copied the spell, was not asked for new targets")
	}
	if me.Life != before-6 {
		t.Errorf("life %d -> %d, want 6 damage (the Bolt and its copy)", before, me.Life)
	}
}

func TestOldB206BonusRoundIgnoresACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Bonus Round", "Sorcery", oldB206BonusOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Plain Bear", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me.ID && c.Name == "Plain Bear" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("%d Plain Bears, want 1: a creature spell was copied", n)
	}
}

// --- Cheering Fanatic ---

func TestOldB206CheeringFanaticDiscountsTheChosenNameForEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fanatic := effectstest.B12Push(g, me.ID, "Cheering Fanatic", "Creature — Goblin", oldB206FanaticOracle, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(fanatic, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ActNameFor(t, g, me.ID, "Divination")
	effectstest.PassPriorityAroundTable(t, g)
	price := func(who uuid.UUID, name, typeLine string) string {
		card := game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, ManaCost: "{2}{U}", Owner: who, Controller: who}
		base, err := game.ParseCost("{2}{U}")
		if err != nil {
			t.Fatalf("ParseCost: %v", err)
		}
		var out game.ParsedCost
		g.WithWriteLock(func() {
			out, err = g.ApplyCostModifiersForEffect(base, game.CostQuery{Game: g, Card: card, Controller: who, FromZone: game.ZoneHand})
			if err != nil {
				t.Fatalf("ApplyCostModifiers: %v", err)
			}
		})
		return out.String()
	}
	for _, who := range []*game.Player{me, opp} {
		if got := price(who.ID, "Divination", "Sorcery"); got != "{1}{U}" {
			t.Errorf("%s pays %s for the named spell, want {1}{U}", who.Name, got)
		}
		if got := price(who.ID, "Opt", "Instant"); got != "{2}{U}" {
			t.Errorf("%s pays %s for another spell, want {2}{U}", who.Name, got)
		}
	}
}

// --- Evil Twin ---

func TestOldB206EvilTwinCopiesAndDestroysACreatureWithTheSameName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	theirs := effectstest.SeedCopyableCreature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	different := effectstest.SeedCopyableCreature(g, opp.ID, "Runeclaw Bear", "Creature — Bear", 2, 2)
	twin := effectstest.CastCatalogSpell(t, g, "Evil Twin", "Creature — Shapeshifter", oldB206TwinOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, mine)
	if got := effectstest.CopyBattlefieldCard(t, g, twin); got.Name != "Grizzly Bears" {
		t.Fatalf("the Twin is %q, want a copy of Grizzly Bears", got.Name)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == twin {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	effectstest.B06AddMana(me, "U", "B")
	if err := g.ActivateCatalogAbility(me.ID, twin, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB206Target(different)}); err == nil {
		t.Error("a creature with a different name was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, twin, 0, game.ActivateAbilityParams{Strict: true, Targets: oldB206Target(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the same-named creature was not destroyed")
	}
	if !g.Battlefield.Contains(different) || !g.Battlefield.Contains(twin) {
		t.Error("another creature, or the Twin, was destroyed")
	}
}

func TestOldB206EvilTwinDeclinedIsAZeroZeroThatDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedCopyableCreature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	twin := effectstest.CastCatalogSpell(t, g, "Evil Twin", "Creature — Shapeshifter", oldB206TwinOracle, nil)
	effectstest.ResolveWithCopyChoice(t, g, uuid.Nil)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(twin) && effectstest.CopyBattlefieldCard(t, g, twin).Name == "Grizzly Bears" {
		t.Error("a declined copy still copied a creature")
	}
}

// --- Fumble ---

func TestOldB206FumbleBouncesAndMovesAurasAndEquipmentToAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	from := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	to := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Lightning Talons", effectstest.AuraTypeLine, oldB206TalonsOracle, false)
	effectstest.NcdAttach(t, g, aura, from)
	sword := effectstest.PushCatalogPermanent(g, opp.ID, "Test Blade", "Artifact — Equipment", "", false)
	effectstest.NcdAttach(t, g, sword, from)
	effectstest.CastCatalogSpell(t, g, "Fumble", "Instant", oldB206FumbleOracle, oldB206Target(from))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(from) {
		t.Error("the target creature was not returned to its owner's hand")
	}
	if !effectstest.AttachedTo(t, g, aura, to) || !effectstest.AttachedTo(t, g, sword, to) {
		t.Error("the Aura and the Equipment were not attached to the other creature")
	}
	for _, id := range []uuid.UUID{aura, sword} {
		if c, ok := effectstest.BattlefieldCard(g, id); !ok || c.Controller != me.ID {
			t.Errorf("%v is not under the caster's control", id)
		}
	}
}

func TestOldB206FumbleAnAuraTheChosenCreatureCannotTakeIsPutIntoTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	from := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	to := effectstest.B12Creature(g, me.ID, "Pro-Red Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == to {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "protection from red")
			}
		}
	})
	aura := effectstest.PushCatalogPermanent(g, opp.ID, "Lightning Talons", effectstest.AuraTypeLine, oldB206TalonsOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == aura {
				g.Battlefield.Cards[i].Colors = []string{"R"}
			}
		}
	})
	effectstest.NcdAttach(t, g, aura, from)
	effectstest.CastCatalogSpell(t, g, "Fumble", "Instant", oldB206FumbleOracle, oldB206Target(from))
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.SlChoose(t, g, me.ID, to)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) || !opp.Graveyard.Contains(aura) {
		t.Error("the Aura that could not enchant the chosen creature was not put into its owner's graveyard")
	}
}
