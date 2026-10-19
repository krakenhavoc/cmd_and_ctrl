package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// before_attackers_cards_test.go is #4062: "Activate only during your
// turn, before attackers are declared" (Diaochan, Artful Beauty and the
// Portal Three Kingdoms creatures). The window is the condition
// DuringYourTurnBeforeAttackersAreDeclared, read off the turn tally's
// record that a declare attackers step has begun; each card has a test of
// its own effect.

const (
	baAlabornVeteran    = "7d9e905b-5f5c-4dbb-95cb-ccba0a5cbf8d"
	baApprenticeSorc    = "e26bf39f-2d1a-4bc1-aed2-d6d4505f1018"
	baCaoCao            = "81a98b14-e827-4241-a401-64890def51af"
	baCapriciousSorc    = "09fe624f-c66a-46e4-a9af-7e3c3ca1a4e3"
	baCoastalWizard     = "d2a29bbf-bfcd-4e4e-8c0b-95a15b58a612"
	baCunningAdvisor    = "b80b6dd3-320b-4304-8073-f74c54e9a8fe"
	baDiaochan          = "bf20bc37-3205-45af-a3be-05d6835c0d87"
	baFireBowman        = "3f23d7d3-28c6-4388-afb5-8f101b973083"
	baGoblinFirestarter = "18279733-5568-427c-b0a2-044bca28f688"
	baHuaTuo            = "00284955-0a1b-4939-9567-a7d34cbcee51"
	baKingsAssassin     = "4e3f89e2-2be6-4b96-b81c-007b6840af82"
	baLadySun           = "dcc363a7-652d-4d2b-9081-3e5735120f21"
	baLoyalRetainers    = "55b88d43-ec45-4435-a316-1c8aee249ff0"
	baLuSu              = "59238ba1-fde4-4e43-9f24-b24125e59de0"
	baNorwoodPriestess  = "902a7588-602c-456f-8250-c94f35e56d3a"
	baPangTong          = "6d9dc632-414e-480c-85ef-c2b98a6460ad"
	baShuFarmer         = "0d059577-0790-4033-b719-a74b6a62262e"
	baSteamCatapult     = "b6ea0c70-022a-43e4-aa40-3dc6a8b3b21e"
	baSternMarshal      = "00602959-e9a1-4706-a1cb-70156a6fc713"
	baStoneCatapult     = "c3ead850-3f22-4edb-ac96-40abbdf27905"
	baTalasResearcher   = "1ed662d7-65b9-4024-a892-c4559fb6aaf9"
	baTempleElder       = "a0ace0db-1bd6-4b30-b0e2-d266967707d7"
	baWuLongbowman      = "68fc55b9-55cc-4caf-a052-b44ddc0b346f"
	baXiahouDun         = "9602f262-62cb-4c1e-91b6-d94780e571cf"
	baXunYu             = "fb5fea4a-1b04-45c6-a207-e2e8493ea968"
	baZhugeJin          = "ce995d3f-be7c-4d2a-8ec4-c7f7f831f0bf"
)

// baCards is every card that prints the window, with its name.
var baCards = []struct{ oracle, name string }{
	{baAlabornVeteran, "Alaborn Veteran"}, {baApprenticeSorc, "Apprentice Sorcerer"},
	{baCaoCao, "Cao Cao, Lord of Wei"}, {baCapriciousSorc, "Capricious Sorcerer"},
	{baCoastalWizard, "Coastal Wizard"}, {baCunningAdvisor, "Cunning Advisor"},
	{baDiaochan, "Diaochan, Artful Beauty"}, {baFireBowman, "Fire Bowman"},
	{baGoblinFirestarter, "Goblin Firestarter"}, {baHuaTuo, "Hua Tuo, Honored Physician"},
	{baKingsAssassin, "King's Assassin"}, {baLadySun, "Lady Sun"},
	{baLoyalRetainers, "Loyal Retainers"}, {baLuSu, "Lu Su, Wu Advisor"},
	{baNorwoodPriestess, "Norwood Priestess"}, {baPangTong, "Pang Tong, \"Young Phoenix\""},
	{baShuFarmer, "Shu Farmer"}, {baSteamCatapult, "Steam Catapult"},
	{baSternMarshal, "Stern Marshal"}, {baStoneCatapult, "Stone Catapult"},
	{baTalasResearcher, "Talas Researcher"}, {baTempleElder, "Temple Elder"},
	{baWuLongbowman, "Wu Longbowman"}, {baXiahouDun, "Xiahou Dun, the One-Eyed"},
	{baXunYu, "Xun Yu, Wei Advisor"}, {baZhugeJin, "Zhuge Jin, Wu Strategist"},
}

// baTable is a four-seat game at the start of the active player's turn,
// with the active player and the next seat.
func baTable(t *testing.T) (g *game.Game, me, opp *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	return g, me, opp
}

// baPermanent puts one of the cards in play, untapped and not summoning
// sick.
func baPermanent(g *game.Game, me *game.Player, name, oracle string) uuid.UUID {
	return pushCatalogPermanent(g, me.ID, name, "Creature — Test", oracle, false)
}

// baActivate activates ability 0 and resolves it.
func baActivate(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID, targets []game.TargetRef) {
	t.Helper()
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Targets: targets}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// Every card declares the gate on its one ability, and the gate refuses at
// the declare attackers step with nothing paid - the stronger-than-printed
// failure (#259) is a card that lost it.
func TestBeforeAttackersCardsAreGated(t *testing.T) {
	for _, c := range baCards {
		t.Run(c.name, func(t *testing.T) {
			spec, ok := Lookup(c.oracle)
			if !ok {
				t.Fatalf("%s (%s) is not registered", c.name, c.oracle)
			}
			if spec.Name != c.name {
				t.Errorf("registered as %q, want %q", spec.Name, c.name)
			}
			if spec.Completeness != CompletenessFull {
				t.Errorf("completeness %v, want Full", spec.Completeness)
			}
			if len(spec.Activated) != 1 || spec.Activated[0].Condition == nil {
				t.Fatalf("want one activated ability carrying the window condition, got %d", len(spec.Activated))
			}
			g, me, _ := baTable(t)
			card := baPermanent(g, me, c.name, c.oracle)
			// Open at the start of the turn: the condition itself holds.
			if !spec.Activated[0].Condition(g, me.ID, card) {
				t.Error("the window is shut at the start of the turn")
			}
			advanceTo(t, g, game.StepDeclareAttackers)
			acRefused(t, g, me.ID, card, 0, game.ActivateAbilityParams{})
			if spec.Activated[0].Condition(g, me.ID, card) {
				t.Error("the window is open in the declare attackers step")
			}
		})
	}
}

// The window opens with the turn, stays open through the beginning of
// combat, shuts when the declare attackers step begins and stays shut for
// the rest of the turn, and is not open on another player's turn; it
// reopens on your next turn.
func TestBeforeAttackersWindowOpensAndShuts(t *testing.T) {
	g, me, _ := baTable(t)
	seat := g.Turn.ActiveSeat
	talas := baPermanent(g, me, "Talas Researcher", baTalasResearcher)
	untap := func() { rtUntap(g, talas) }
	drew := func(want int) {
		t.Helper()
		if got := me.Hand.Size(); got != want {
			t.Fatalf("hand size %d, want %d", got, want)
		}
	}
	hand := me.Hand.Size()

	// The draw step (where the harness starts).
	baActivate(t, g, me, talas, nil)
	hand++
	drew(hand)
	untap()

	advanceTo(t, g, game.StepPrecombatMain)
	baActivate(t, g, me, talas, nil)
	hand++
	drew(hand)
	untap()

	advanceTo(t, g, game.StepBeginCombat)
	baActivate(t, g, me, talas, nil)
	hand++
	drew(hand)
	untap()

	// With the declare attackers step begun: shut, and nothing paid.
	for _, step := range []game.Step{game.StepDeclareAttackers, game.StepEndCombat, game.StepPostcombatMain, game.StepEnd} {
		advanceTo(t, g, step)
		acRefused(t, g, me.ID, talas, 0, game.ActivateAbilityParams{})
	}
	drew(hand)

	// Another player's turn is not your turn.
	advanceToMainOf(t, g, (seat+1)%len(g.Seats))
	acRefused(t, g, me.ID, talas, 0, game.ActivateAbilityParams{})

	// Your next turn: the window is open again.
	advanceToMainOf(t, g, seat)
	baActivate(t, g, me, talas, nil)
	drew(hand + 1 + 1) // one for the intervening turns' draw step of this one, one for the activation
}

// An item on the stack changes nothing: the window is a position in the
// turn, not a sorcery-speed gate, and a response in the beginning of combat
// step is allowed.
func TestBeforeAttackersAllowsAResponseInTheBeginningOfCombat(t *testing.T) {
	g, me, _ := baTable(t)
	first := baPermanent(g, me, "Lu Su, Wu Advisor", baLuSu)
	second := baPermanent(g, me, "Talas Researcher", baTalasResearcher)
	advanceTo(t, g, game.StepBeginCombat)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, first, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("first: %v", err)
	}
	if err := g.ActivateCatalogAbility(me.ID, second, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("with an item on the stack: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
}

// The ruling: a turn with several combat phases opens the window only
// before the declare attackers step of the FIRST combat.
func TestBeforeAttackersStaysShutInAnExtraCombat(t *testing.T) {
	g, me, _ := baTable(t)
	talas := baPermanent(g, me, "Talas Researcher", baTalasResearcher)
	advanceTo(t, g, game.StepDeclareAttackers)
	g.WithWriteLock(func() {
		g.AddPhasesForEffect(uuid.Nil, game.PhaseAnchor{Kind: game.AnchorThisPhase}, game.PhaseKindCombat)
	})
	advanceTo(t, g, game.StepEndCombat)
	// Into the extra combat's beginning of combat step.
	for i := 0; i < 8 && !(g.Turn.Step == game.StepBeginCombat && g.Turn.PhaseOrdinal == 2); i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if g.Turn.Step != game.StepBeginCombat || g.Turn.PhaseOrdinal != 2 {
		t.Fatalf("at %s in combat %d, want the second combat's beginning of combat", g.Turn.Step, g.Turn.PhaseOrdinal)
	}
	acRefused(t, g, me.ID, talas, 0, game.ActivateAbilityParams{})
}

// The record lives in the turn tally a restore point carries: a table
// reloaded after attackers were declared keeps the window shut.
func TestBeforeAttackersWindowSurvivesARestore(t *testing.T) {
	g, me, _ := baTable(t)
	talas := baPermanent(g, me, "Talas Researcher", baTalasResearcher)
	advanceTo(t, g, game.StepDeclareAttackers)
	restored := restoreRoundTrip(t, g, true)
	acRefused(t, restored, me.ID, talas, 0, game.ActivateAbilityParams{})
	advanceTo(t, restored, game.StepPostcombatMain)
	acRefused(t, restored, me.ID, talas, 0, game.ActivateAbilityParams{})

	// And a table reloaded before the step keeps it open.
	g2, me2, _ := baTable(t)
	talas2 := baPermanent(g2, me2, "Talas Researcher", baTalasResearcher)
	advanceTo(t, g2, game.StepBeginCombat)
	restored2 := restoreRoundTrip(t, g2, true)
	baActivate(t, restored2, restored2.PlayerByIDForEffect(me2.ID), talas2, nil)
}

// A turn in which the first combat is skipped never begins a declare
// attackers step, so nothing closes the window: the tally is the reading.
func TestBeforeAttackersDeclareAttackersBegunFactFollowsTheTally(t *testing.T) {
	g, me, _ := baTable(t)
	talas := baPermanent(g, me, "Talas Researcher", baTalasResearcher)
	spec, _ := Lookup(baTalasResearcher)
	g.WithWriteLock(func() {
		if g.TurnTally.StepsBegun[game.StepDeclareAttackers] != 0 {
			t.Fatal("setup: a declare attackers step has already begun")
		}
		if !spec.Activated[0].Condition(g, me.ID, talas) {
			t.Error("closed with no declare attackers step begun")
		}
		g.TurnTally.StepsBegun = map[game.Step]int{game.StepDeclareAttackers: 1}
		if spec.Activated[0].Condition(g, me.ID, talas) {
			t.Error("open although a declare attackers step has begun")
		}
	})
}

// --- the cards ---------------------------------------------------------

func TestBeforeAttackersPumps(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		power, tough int  // expected on a 2/2 after resolution
		mine         bool // the target must be yours
	}{
		{"Alaborn Veteran", baAlabornVeteran, 4, 4, false},
		{"Stern Marshal", baSternMarshal, 4, 4, false},
		{"Pang Tong, \"Young Phoenix\"", baPangTong, 2, 4, false},
		{"Xun Yu, Wei Advisor", baXunYu, 4, 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			pumper := baPermanent(g, me, tc.name, tc.oracle)
			mine := ctrlPushCreature(g, me.ID, "My Bear")
			theirs := ctrlPushCreature(g, opp.ID, "Their Bear")
			if tc.mine {
				if err := g.ActivateCatalogAbility(me.ID, pumper, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs)}); err == nil {
					t.Fatal("the pump targeted a creature you don't control")
				}
				if b16Tapped(t, g, pumper) {
					t.Fatal("a refused activation tapped the source")
				}
			}
			target := theirs
			if tc.mine {
				target = mine
			}
			baActivate(t, g, me, pumper, ltCardTarget(target))
			c := layeredCard(t, g, target)
			if c.CurrentPower() != tc.power || c.CurrentToughness() != tc.tough {
				t.Errorf("%d/%d, want %d/%d", c.CurrentPower(), c.CurrentToughness(), tc.power, tc.tough)
			}
			advancePastCleanupForTest(t, g)
			c = layeredCard(t, g, target)
			if c.CurrentPower() != 2 || c.CurrentToughness() != 2 {
				t.Errorf("after the turn %d/%d, want 2/2", c.CurrentPower(), c.CurrentToughness())
			}
		})
	}
}

func TestBeforeAttackersPingers(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Apprentice Sorcerer", baApprenticeSorc},
		{"Capricious Sorcerer", baCapriciousSorc},
		{"Wu Longbowman", baWuLongbowman},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			pinger := baPermanent(g, me, tc.name, tc.oracle)
			life := opp.Life
			baActivate(t, g, me, pinger, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			if opp.Life != life-1 {
				t.Errorf("life %d, want %d", opp.Life, life-1)
			}
			if !b16Tapped(t, g, pinger) {
				t.Error("the ping did not tap the creature")
			}
			// A creature takes the damage too.
			rtUntap(g, pinger)
			bear := ctrlPushCreature(g, opp.ID, "Bear")
			baActivate(t, g, me, pinger, ltCardTarget(bear))
			if c := layeredCard(t, g, bear); c.DamageMarked != 1 {
				t.Errorf("damage %d, want 1", c.DamageMarked)
			}
		})
	}
}

func TestBeforeAttackersSacrificePingers(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Fire Bowman", baFireBowman},
		{"Goblin Firestarter", baGoblinFirestarter},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			pinger := baPermanent(g, me, tc.name, tc.oracle)
			life := opp.Life
			baActivate(t, g, me, pinger, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
			if opp.Life != life-1 {
				t.Errorf("life %d, want %d", opp.Life, life-1)
			}
			if onBattlefield(g, pinger) {
				t.Error("the creature was not sacrificed")
			}
			if !me.Graveyard.Contains(pinger) {
				t.Error("the sacrificed creature is not in the graveyard")
			}
		})
	}
}

func TestBeforeAttackersGainLife(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Temple Elder", baTempleElder},
		{"Shu Farmer", baShuFarmer},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := baTable(t)
			c := baPermanent(g, me, tc.name, tc.oracle)
			life := me.Life
			baActivate(t, g, me, c, nil)
			if me.Life != life+1 {
				t.Errorf("life %d, want %d", me.Life, life+1)
			}
		})
	}
}

func TestBeforeAttackersDrawers(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Talas Researcher", baTalasResearcher},
		{"Lu Su, Wu Advisor", baLuSu},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := baTable(t)
			c := baPermanent(g, me, tc.name, tc.oracle)
			hand := me.Hand.Size()
			baActivate(t, g, me, c, nil)
			if got := me.Hand.Size(); got != hand+1 {
				t.Errorf("hand %d, want %d", got, hand+1)
			}
		})
	}
}

func TestBeforeAttackersOpponentDiscards(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		n            int
	}{
		{"Cao Cao, Lord of Wei", baCaoCao, 2},
		{"Cunning Advisor", baCunningAdvisor, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			c := baPermanent(g, me, tc.name, tc.oracle)
			hand := flipFillHand(g, opp, 4)
			if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
				t.Fatal(err)
			}
			passPriorityAroundTable(t, g)
			// The opponent chooses, and picks the LAST cards to prove it.
			picks := hand[len(hand)-tc.n:]
			answerDiscard(t, g, opp.ID, picks...)
			passPriorityAroundTable(t, g)
			if got := opp.Hand.Size(); got != 4-tc.n {
				t.Errorf("opponent holds %d cards, want %d", got, 4-tc.n)
			}
			for _, id := range picks {
				if !opp.Graveyard.Contains(id) {
					t.Error("a card the opponent chose is not in their graveyard")
				}
			}
			// Only an opponent is a legal target.
			rtUntap(g, c)
			if err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
				t.Error("targeted yourself")
			}
		})
	}
}

func TestBeforeAttackersCaoCaoWithAOneCardHandDiscardsIt(t *testing.T) {
	g, me, opp := baTable(t)
	c := baPermanent(g, me, "Cao Cao, Lord of Wei", baCaoCao)
	hand := flipFillHand(g, opp, 1)
	baActivate(t, g, me, c, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	if discardChoiceFor(g, opp.ID) != nil {
		answerDiscard(t, g, opp.ID, hand...)
		passPriorityAroundTable(t, g)
	}
	if opp.Hand.Size() != 0 || !opp.Graveyard.Contains(hand[0]) {
		t.Errorf("the lone card should have been discarded (hand %d)", opp.Hand.Size())
	}
}

func TestBeforeAttackersDestroyTappedCreature(t *testing.T) {
	for _, tc := range []struct {
		name, oracle string
		blackOK      bool
	}{
		{"King's Assassin", baKingsAssassin, true},
		{"Steam Catapult", baSteamCatapult, true},
		{"Stone Catapult", baStoneCatapult, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			killer := baPermanent(g, me, tc.name, tc.oracle)
			untapped := ctrlPushCreature(g, opp.ID, "Untapped Bear")
			if err := g.ActivateCatalogAbility(me.ID, killer, 0, game.ActivateAbilityParams{Targets: ltCardTarget(untapped)}); err == nil {
				t.Fatal("targeted an untapped creature")
			}
			tapped := ctrlPushCreature(g, opp.ID, "Tapped Bear")
			rtRetap(g, tapped)
			black := acPermanent(g, opp.ID, "Black Bear", "Creature — Test", 2, 2, "B")
			rtRetap(g, black)
			err := g.ActivateCatalogAbility(me.ID, killer, 0, game.ActivateAbilityParams{Targets: ltCardTarget(black)})
			if tc.blackOK && err != nil || !tc.blackOK && err == nil {
				t.Fatalf("targeting a tapped black creature: err = %v, want allowed = %v", err, tc.blackOK)
			}
			if err == nil {
				passPriorityAroundTable(t, g)
				if onBattlefield(g, black) {
					t.Fatal("the black creature survived")
				}
				return
			}
			baActivate(t, g, me, killer, ltCardTarget(tapped))
			if onBattlefield(g, tapped) || !onBattlefield(g, untapped) {
				t.Error("only the tapped creature should have died")
			}
		})
	}
}

func TestBeforeAttackersHuaTuoPutsACreatureOnTopOfYourLibrary(t *testing.T) {
	g, me, opp := baTable(t)
	hua := baPermanent(g, me, "Hua Tuo, Honored Physician", baHuaTuo)
	mine := pushGraveyardCardForTest(me, "Fallen Ally")
	theirs := pushGraveyardCardForTest(opp, "Their Fallen")
	if err := g.ActivateCatalogAbility(me.ID, hua, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs)}); err == nil {
		t.Fatal("targeted a card in an opponent's graveyard")
	}
	baActivate(t, g, me, hua, ltCardTarget(mine))
	top, err := me.Library.Top()
	if err != nil || top.InstanceID != mine {
		t.Errorf("top of the library is %v (%v), want the creature card", top.Name, err)
	}
	if me.Graveyard.Contains(mine) {
		t.Error("the card is still in the graveyard")
	}
}

func TestBeforeAttackersLoyalRetainersReturnsALegendFromYourGraveyard(t *testing.T) {
	g, me, opp := baTable(t)
	retainers := baPermanent(g, me, "Loyal Retainers", baLoyalRetainers)
	legend := me.Graveyard
	legendID := uuid.New()
	legend.PushTop(game.Card{InstanceID: legendID, Name: "Fallen Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	plain := pushGraveyardCardForTest(me, "Plain Creature")
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Legend", TypeLine: "Legendary Creature — Human",
		Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	for _, bad := range []uuid.UUID{plain, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, retainers, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bad)}); err == nil {
			t.Fatal("targeted a card that is not a legendary creature card in your graveyard")
		}
		if !onBattlefield(g, retainers) {
			t.Fatal("a refused activation sacrificed the Retainers")
		}
	}
	baActivate(t, g, me, retainers, ltCardTarget(legendID))
	if onBattlefield(g, retainers) {
		t.Error("the Retainers were not sacrificed")
	}
	if !onBattlefield(g, legendID) {
		t.Fatal("the legend did not return to the battlefield")
	}
	if got := controllerOf(t, g, legendID); got != me.ID {
		t.Errorf("controller %s, want %s", got, me.ID)
	}
}

func TestBeforeAttackersXiahouDunReturnsABlackCardAndHasHorsemanship(t *testing.T) {
	g, me, opp := baTable(t)
	dun := baPermanent(g, me, "Xiahou Dun, the One-Eyed", baXiahouDun)
	if !game.HasKeyword(findBattlefieldCardForTest(g, dun), "horsemanship") {
		t.Error("Xiahou Dun lost horsemanship")
	}
	black := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: black, Name: "Black Spell", TypeLine: "Sorcery", Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	green := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: green, Name: "Green Spell", TypeLine: "Sorcery", Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	theirs := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: theirs, Name: "Their Black Spell", TypeLine: "Sorcery", Colors: []string{"B"}, Owner: opp.ID, Controller: opp.ID})
	for _, bad := range []uuid.UUID{green, theirs} {
		if err := g.ActivateCatalogAbility(me.ID, dun, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bad)}); err == nil {
			t.Fatal("targeted a card that is not a black card in your graveyard")
		}
	}
	baActivate(t, g, me, dun, ltCardTarget(black))
	if !me.Hand.Contains(black) {
		t.Error("the black card is not in hand")
	}
	if onBattlefield(g, dun) {
		t.Error("Xiahou Dun was not sacrificed")
	}
}

func TestBeforeAttackersNorwoodPriestessPutsAGreenCreatureOntoTheBattlefield(t *testing.T) {
	g, me, _ := baTable(t)
	priestess := baPermanent(g, me, "Norwood Priestess", baNorwoodPriestess)
	red := handCardForTest(me, "Red Creature", "Creature — Test", "")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == red {
			me.Hand.Cards[i].Colors = []string{"R"}
		}
	}
	green := handCardForTest(me, "Green Creature", "Creature — Test", "")
	for i := range me.Hand.Cards {
		if me.Hand.Cards[i].InstanceID == green {
			me.Hand.Cards[i].Colors = []string{"G"}
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, priestess, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no prompt for the green creature card")
	}
	for _, id := range pick.ChooseCards {
		if id == red {
			t.Error("the red creature card was offered")
		}
	}
	answerChooseCards(t, g, me.ID, green)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, green) || onBattlefield(g, red) {
		t.Errorf("green on battlefield %v, red on battlefield %v", onBattlefield(g, green), onBattlefield(g, red))
	}
}

func TestBeforeAttackersZhugeJinGrantsEvasion(t *testing.T) {
	g, me, opp := baTable(t)
	zhuge := baPermanent(g, me, "Zhuge Jin, Wu Strategist", baZhugeJin)
	bear := ctrlPushCreature(g, me.ID, "Bear")
	blocker := ctrlPushCreature(g, opp.ID, "Blocker")
	baActivate(t, g, me, zhuge, ltCardTarget(bear))
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, bear); err == nil {
		t.Error("a creature that can't be blocked was blocked")
	}
}

func TestBeforeAttackersBounceThisAndAnotherCreature(t *testing.T) {
	for _, tc := range []struct{ name, oracle string }{
		{"Coastal Wizard", baCoastalWizard},
		{"Lady Sun", baLadySun},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := baTable(t)
			bouncer := baPermanent(g, me, tc.name, tc.oracle)
			bear := ctrlPushCreature(g, opp.ID, "Bear")
			if err := g.ActivateCatalogAbility(me.ID, bouncer, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bouncer)}); err == nil {
				t.Fatal("targeted itself: the second creature must be another")
			}
			baActivate(t, g, me, bouncer, ltCardTarget(bear))
			if onBattlefield(g, bouncer) || onBattlefield(g, bear) {
				t.Error("both creatures should have left the battlefield")
			}
			if !me.Hand.Contains(bouncer) {
				t.Error("the source is not in its owner's hand")
			}
			if !opp.Hand.Contains(bear) {
				t.Error("the target is not in its owner's hand")
			}
		})
	}
}

func TestBeforeAttackersBounceThisAndAnotherWhenTheTargetLeaves(t *testing.T) {
	g, me, opp := baTable(t)
	bouncer := baPermanent(g, me, "Coastal Wizard", baCoastalWizard)
	bear := ctrlPushCreature(g, opp.ID, "Bear")
	if err := g.ActivateCatalogAbility(me.ID, bouncer, 0, game.ActivateAbilityParams{Targets: ltCardTarget(bear)}); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear) })
	passPriorityAroundTable(t, g)
	// Its only target is illegal, so the ability does not resolve at all
	// (CR 608.2b) and the Wizard stays.
	if !onBattlefield(g, bouncer) {
		t.Error("with its only target gone the ability still returned the Wizard")
	}
}

func TestBeforeAttackersDiaochanDestroysYourPickAndAnOpponentsPick(t *testing.T) {
	g, me, opp := baTable(t)
	diaochan := baPermanent(g, me, "Diaochan, Artful Beauty", baDiaochan)
	mine := ctrlPushCreature(g, me.ID, "My Bear")
	theirs := ctrlPushCreature(g, opp.ID, "Their Bear")
	chosenByOpp := ctrlPushCreature(g, g.Seats[2].ID, "Third Bear")
	if err := g.ActivateCatalogAbility(me.ID, diaochan, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// Three opponents could choose: you pick which one does.
	otChooseOpponent(t, g, me.ID, opp.ID)
	otPickFor(t, g, opp.ID, me.ID)
	otAnswer(t, g, opp.ID, otCard(mine))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Error("the creature you chose survived")
	}
	if onBattlefield(g, mine) {
		t.Error("the creature the opponent chose survived (they may pick one of yours)")
	}
	if !onBattlefield(g, chosenByOpp) {
		t.Error("an unchosen creature died")
	}
}

func TestBeforeAttackersDiaochanTheSameCreatureTwiceDestroysItOnce(t *testing.T) {
	g, me, opp := baTable(t)
	diaochan := baPermanent(g, me, "Diaochan, Artful Beauty", baDiaochan)
	theirs := ctrlPushCreature(g, opp.ID, "Their Bear")
	if err := g.ActivateCatalogAbility(me.ID, diaochan, 0, game.ActivateAbilityParams{Targets: ltCardTarget(theirs)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	otChooseOpponent(t, g, me.ID, opp.ID)
	otAnswer(t, g, opp.ID, otCard(theirs))
	passPriorityAroundTable(t, g)
	if onBattlefield(g, theirs) {
		t.Error("the creature survived")
	}
}

// The error the view and the enumerator share when the window is shut.
func TestBeforeAttackersRefusalIsAConditionRefusal(t *testing.T) {
	g, me, _ := baTable(t)
	c := baPermanent(g, me, "Shu Farmer", baShuFarmer)
	advanceTo(t, g, game.StepDeclareAttackers)
	err := g.ActivateCatalogAbility(me.ID, c, 0, game.ActivateAbilityParams{})
	if !errors.Is(err, game.ErrConditionNotMet) {
		t.Errorf("err = %v, want ErrConditionNotMet", err)
	}
}
