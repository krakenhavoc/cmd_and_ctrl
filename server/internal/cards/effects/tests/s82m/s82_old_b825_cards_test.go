package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b825_cards_test.go — Homelands: Auras, batch 825 (S82 older sets,
// #5635). Every test plays the real catalog entry through the engine.

const (
	oldB825Bestiary = "320d8f2b-005d-4769-80d2-f46d96afffe7"
	oldB825Funeral  = "4d8afa5c-cef3-4e67-bd4c-ebd5c68b96a7"
	oldB825Curse    = "51708bd1-23db-4597-84a7-f36abd93c537"
	oldB825Mine     = "a6bfa03e-6317-4550-960b-fbd2a30e521f"
	oldB825Carapace = "b893b554-9156-4d79-aca8-6e09c523760b"
	oldB825Harness  = "f2e106c6-7666-42cd-9a02-e51327fb80e8"

	oldB825Elves    = "68954295-54e3-4303-a6bc-fc4547a4e3a3"
	oldB825Sorcerer = "5e961d15-5972-4e4b-9385-1cd7cd7c6bbe"
	oldB825Dragon   = "711eea87-0fa3-46e0-a42b-fa5a86455f04"
)

func TestOldB825EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB825Bestiary: "Serra Bestiary", oldB825Funeral: "Funeral March", oldB825Curse: "Ironclaw Curse",
		oldB825Mine: "Orcish Mine", oldB825Carapace: "Carapace", oldB825Harness: "Mammoth Harness",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB825AurasRefuseAnIllegalHost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	forest := effectstest.PushLandFor(g, g.Seats[1].ID, "Forest", "Basic Land — Forest")
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	for _, tc := range []struct {
		name, oracle string
		host         uuid.UUID
	}{
		{"Serra Bestiary", oldB825Bestiary, forest}, {"Funeral March", oldB825Funeral, forest},
		{"Ironclaw Curse", oldB825Curse, forest}, {"Carapace", oldB825Carapace, forest},
		{"Mammoth Harness", oldB825Harness, forest}, {"Orcish Mine", oldB825Mine, bear},
	} {
		if err := effectstest.CastCatalogSpellErr(t, g, tc.name, effectstest.AuraTypeLine, tc.oracle, effectstest.B16TargetCard(tc.host)); err == nil {
			t.Errorf("%s was cast on an illegal host", tc.name)
		}
	}
}

func oldB825Upkeep(t *testing.T) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	host := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Serra Bestiary", oldB825Bestiary, host)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	return g, me, host, aura
}

func TestOldB825SerraBestiaryUpkeepPaymentKeepsIt(t *testing.T) {
	g, me, _, aura := oldB825Upkeep(t)
	effectstest.B06AddMana(me, "W", "W")
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("no pay prompt for the upkeep trigger")
	}
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(aura) {
		t.Error("the Aura left after its controller paid {W}{W}")
	}
}

func TestOldB825SerraBestiaryIsSacrificedWhenNotPaid(t *testing.T) {
	g, me, _, aura := oldB825Upkeep(t)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		effectstest.AnswerPayUnless(t, g, me.ID, false)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura stayed without payment")
	}
}

func TestOldB825SerraBestiaryStopsAttackingAndBlocking(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	free := effectstest.B12Creature(g, me.ID, "Free Bear", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Serra Bestiary", oldB825Bestiary, mine)
	effectstest.Enchant(t, g, "Serra Bestiary", oldB825Bestiary, theirs)
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mine, opp.ID); err == nil {
		t.Fatal("the enchanted creature attacked")
	}
	if err := g.DeclareAttacker(free, opp.ID); err != nil {
		t.Fatalf("an unenchanted creature could not attack: %v", err)
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepDeclareBlockers)
	if effectstest.BlockRefusal(t, g, theirs, free) == "" {
		t.Error("the enchanted creature blocked")
	}
}

func TestOldB825SerraBestiaryStopsTapAbilitiesAndNothingElse(t *testing.T) {
	cases := []struct {
		name, oracle, typeLine string
		power, toughness       int
		activate               func(g *game.Game, me *game.Player, id uuid.UUID) error
		wantRefused            bool
	}{
		{"Prodigal Sorcerer", oldB825Sorcerer, "Creature — Human Wizard", 1, 1, func(g *game.Game, me *game.Player, id uuid.UUID) error {
			return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}})
		}, true},
		{"Llanowar Elves", oldB825Elves, "Creature — Elf Druid", 1, 1, func(g *game.Game, me *game.Player, id uuid.UUID) error {
			return g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{})
		}, true},
		{"Shivan Dragon", oldB825Dragon, "Creature — Dragon", 5, 5, func(g *game.Game, me *game.Player, id uuid.UUID) error {
			effectstest.B06AddMana(me, "R")
			return g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true})
		}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			host := effectstest.B12Push(g, me.ID, tc.name, tc.typeLine, tc.oracle, tc.power, tc.toughness)
			aura := effectstest.Enchant(t, g, "Serra Bestiary", oldB825Bestiary, host)
			err := tc.activate(g, me, host)
			if tc.wantRefused && !errors.Is(err, game.ErrCantActivate) {
				t.Fatalf("activate under the Aura: err = %v, want ErrCantActivate", err)
			}
			if !tc.wantRefused && err != nil {
				t.Fatalf("an ability without {T} in its cost was refused: %v", err)
			}
			if !tc.wantRefused {
				return
			}
			// The lock ends with the Aura.
			effectstest.P1Destroy(t, g, aura)
			effectstest.PassPriorityAroundTable(t, g)
			me.ManaPool = nil
			if err := tc.activate(g, me, host); err != nil {
				t.Fatalf("activate after the Aura left: %v", err)
			}
		})
	}
}

func TestOldB825FuneralMarchMakesItsControllerSacrificeWhenItLeaves(t *testing.T) {
	for _, how := range []string{"dies", "bounced"} {
		t.Run(how, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
			keep := effectstest.B12Creature(g, opp.ID, "Keep", "Creature — Bear", 1, 1)
			fodder := effectstest.B12Creature(g, opp.ID, "Fodder", "Creature — Bear", 1, 1)
			mine := effectstest.B12Creature(g, g.Seats[0].ID, "Mine", "Creature — Bear", 1, 1)
			effectstest.Enchant(t, g, "Funeral March", oldB825Funeral, host)
			if how == "dies" {
				effectstest.P1Destroy(t, g, host)
			} else if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, host); err != nil {
				t.Fatalf("bounce: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerSacrifice(t, g, opp.ID, fodder)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(fodder) {
				t.Error("the chosen creature was not sacrificed")
			}
			if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(mine) {
				t.Error("a creature other than the chosen one was sacrificed")
			}
		})
	}
}

func TestOldB825FuneralMarchDoesNothingWhileTheCreatureStays(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, opp.ID, "Other", "Creature — Bear", 1, 1)
	aura := effectstest.Enchant(t, g, "Funeral March", oldB825Funeral, host)
	// The Aura itself leaving is not the enchanted creature leaving.
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(host) || !g.Battlefield.Contains(other) {
		t.Error("a creature was sacrificed though the enchanted creature never left")
	}
}

func TestOldB825CarapaceBuffsAndRegeneratesBySacrificingTheAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Carapace", oldB825Carapace, bear)
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("toughness %d, want 2 + 2", got)
	}
	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("power %d, want 2 (+0)", got)
	}
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura was not sacrificed as the cost")
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 2 {
		t.Errorf("toughness %d after the Aura left, want 2", got)
	}
	effectstest.P1Destroy(t, g, bear)
	if !g.Battlefield.Contains(bear) {
		t.Error("the regeneration shield did not save the enchanted creature")
	}
}

func TestOldB825CarapaceCantBeActivatedWithoutItsAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	aura := effectstest.Enchant(t, g, "Carapace", oldB825Carapace, bear)
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, aura, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated from the graveyard")
	}
	effectstest.P1Destroy(t, g, bear)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature survived with no shield")
	}
}

func TestOldB825IronclawCurseShrinksAndRefusesBigAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Wall", 0, 3)
	big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Ogre", 3, 3)
	equal := effectstest.B12Creature(g, me.ID, "Equal", "Creature — Ogre", 2, 2)
	small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Rat", 1, 1)
	effectstest.Enchant(t, g, "Ironclaw Curse", oldB825Curse, host)
	if got := effectstest.EffectiveToughness(t, g, host); got != 2 {
		t.Fatalf("toughness %d, want 3 - 1", got)
	}
	effectstest.AttackIntoBlocks(t, g, opp.ID, big, equal, small)
	if effectstest.BlockRefusal(t, g, host, big) == "" {
		t.Error("blocked a creature with power 3 (>= toughness 2)")
	}
	if effectstest.BlockRefusal(t, g, host, equal) == "" {
		t.Error("blocked a creature with power equal to its toughness")
	}
	if reason := effectstest.BlockRefusal(t, g, host, small); reason != "" {
		t.Errorf("refused to block a creature with power 1: %v", reason)
	}
}

func TestOldB825IronclawCurseRuleEndsWithTheAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Wall", 0, 3)
	two := effectstest.B12Creature(g, me.ID, "Two", "Creature — Ogre", 2, 2)
	aura := effectstest.Enchant(t, g, "Ironclaw Curse", oldB825Curse, host)
	// With the Aura gone the toughness is 3 again and a power-2 attacker may be blocked.
	effectstest.P1Destroy(t, g, aura)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, opp.ID, two)
	if reason := effectstest.BlockRefusal(t, g, host, two); reason != "" {
		t.Errorf("still refused after the Aura left: %v", reason)
	}
}

func oldB825MineTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	aura := effectstest.Enchant(t, g, "Orcish Mine", oldB825Mine, land)
	return g, me, opp, land, aura
}

func TestOldB825OrcishMineEntersWithThreeOreCounters(t *testing.T) {
	g, _, _, _, aura := oldB825MineTable(t)
	if got := effectstest.CounterCount(g, aura, "ore"); got != 3 {
		t.Errorf("%d ore counters, want 3", got)
	}
}

func TestOldB825OrcishMineCountsDownOnUpkeepAndTapThenDestroysTheLand(t *testing.T) {
	g, me, opp, land, aura := oldB825MineTable(t)
	life, myLife := opp.Life, me.Life
	// Whenever enchanted land becomes tapped: 3 -> 2.
	if err := g.ActivateManaAbility(opp.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, aura, "ore"); got != 2 {
		t.Fatalf("%d ore counters after the land was tapped, want 2", got)
	}
	// The opponent's upkeep is not "your upkeep": the land untaps, nothing is removed.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, aura, "ore"); got != 2 {
		t.Fatalf("%d ore counters after another player's upkeep, want 2", got)
	}
	if err := g.ActivateManaAbility(opp.ID, land, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana again: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, aura, "ore"); got != 1 {
		t.Fatalf("%d ore counters, want 1", got)
	}
	if !g.Battlefield.Contains(land) {
		t.Fatal("the land was destroyed with a counter left")
	}
	// At the beginning of your upkeep: 1 -> 0, the last one.
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(land) {
		t.Error("the enchanted land survived the last ore counter")
	}
	if g.Battlefield.Contains(aura) {
		t.Error("the Aura stayed on the battlefield with no land")
	}
	if opp.Life != life-2 {
		t.Errorf("the land's controller's life %d -> %d, want -2", life, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("the Aura's controller's life %d -> %d, want unchanged", myLife, me.Life)
	}
}

func TestOldB825MammothHarnessRemovesFlyingAndGivesTheOtherCreatureFirstStrike(t *testing.T) {
	// The enchanted flier attacks; the creature that blocks it gains first strike.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	flier := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Flier", TypeLine: "Creature — Bird", Power: 2, Toughness: 2,
		Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	bystander := effectstest.B12Creature(g, opp.ID, "Bystander", "Creature — Bear", 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, flier, "flying") {
		t.Fatal("the test flier has no flying")
	}
	effectstest.Enchant(t, g, "Mammoth Harness", oldB825Harness, flier)
	if effectstest.HasEffectiveKeyword(t, g, flier, "flying") {
		t.Fatal("the enchanted creature kept flying")
	}
	effectstest.Pr7bAttack(t, g, opp.ID, flier)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{bear: flier})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("the blocker did not gain first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, bystander, "first strike") {
		t.Error("a creature that did not block gained first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, flier, "first strike") {
		t.Error("the enchanted creature gained first strike")
	}
	// Until end of turn.
	effectstest.AdvanceToMainOf(t, g, 1)
	if effectstest.HasEffectiveKeyword(t, g, bear, "first strike") {
		t.Error("first strike outlasted the turn")
	}
}

func TestOldB825MammothHarnessGivesTheAttackerFirstStrikeWhenTheEnchantedCreatureBlocks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	host := effectstest.B12Creature(g, opp.ID, "Host", "Creature — Bear", 2, 2)
	effectstest.Enchant(t, g, "Mammoth Harness", oldB825Harness, host)
	effectstest.Pr7bAttack(t, g, opp.ID, attacker, other)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{host: attacker})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "first strike") {
		t.Error("the blocked attacker did not gain first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, other, "first strike") {
		t.Error("an unblocked attacker gained first strike")
	}
	if effectstest.HasEffectiveKeyword(t, g, host, "first strike") {
		t.Error("the enchanted creature gained first strike")
	}
}
