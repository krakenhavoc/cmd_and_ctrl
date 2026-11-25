package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b190_test.go — Guilds of Ravnica, older-sets batch 190 (#5525).
// Every test plays the real catalog entry through the engine.

const (
	oldB190StreetRiot      = "526db754-abb6-450c-bead-81dd1e1c347c"
	oldB190BountyOfMight   = "6f067bfd-1c59-463f-99e1-61161b43b405"
	oldB190ErstwhileTroop  = "a1640d9a-367f-4b90-bd4d-18f2c9244ccf"
	oldB190ArtfulTakedown  = "edc193bf-2987-4c7b-9cf3-69089ff13764"
	oldB190RiskFactor      = "7f61a2a6-2fe2-48ec-8e2e-74d147be0a0a"
	oldB190BeaconBolt      = "884686dd-515c-4484-a324-e7d126903a42"
	oldB190GraviticPunch   = "90ea4c57-e0ff-4697-912c-c6e2dc873529"
	oldB190SonicAssault    = "dfe9e485-c070-4efd-8c35-69d4eb3f58e4"
	oldB190SwarmGuildmage  = "14b8b35e-57b6-4577-97a8-9d9c513db0b8"
	oldB190LedevChampion   = "97b2aa77-122f-41e3-ba28-c7040dae0f01"
	oldB190AttendantVraska = "9a639f82-24b6-4b54-8b9e-1fbc0f0094bf"
)

func TestOldB190EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB190StreetRiot, oldB190BountyOfMight, oldB190ErstwhileTroop, oldB190ArtfulTakedown,
		oldB190RiskFactor, oldB190BeaconBolt, oldB190GraviticPunch, oldB190SonicAssault, oldB190SwarmGuildmage,
		oldB190LedevChampion, oldB190AttendantVraska} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB190StreetRiotPumpsYourCreaturesOnlyDuringYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, active.ID, "Street Riot", "Enchantment", oldB190StreetRiot, false)
	// A Street Riot another player controls does nothing on this turn: it is not their turn.
	effectstest.PushCatalogPermanent(g, other.ID, "Street Riot", "Enchantment", oldB190StreetRiot, false)
	mine := effectstest.B12Push(g, active.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, other.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	if p, tough := effectstest.PtOf(t, g, mine); p != 3 || tough != 2 {
		t.Errorf("my creature is %d/%d on my turn, want 3/2", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Error("my creature has no trample on my turn")
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature got the bonus on a turn that is not theirs")
	}
	// A creature that arrives later in the turn gets it too (CR 611.3a).
	late := effectstest.B12Push(g, active.ID, "Late Bear", "Creature — Bear", "", 2, 2)
	if p, _ := effectstest.PtOf(t, g, late); p != 3 || !effectstest.HasEffectiveKeyword(t, g, late, "trample") {
		t.Error("a creature that arrived later got no bonus")
	}
}

func TestOldB190BountyOfMightPumpsThreeTargetsOrOneThreeTimes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.B12Push(g, me.ID, "A", "Creature — Bear", "", 2, 2)
	b := effectstest.B12Push(g, me.ID, "B", "Creature — Bear", "", 2, 2)
	c := effectstest.B12Push(g, opp.ID, "C", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Bounty of Might", "Instant", oldB190BountyOfMight, effectstest.CardTargets(a, b, c))
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "C": c} {
		if p, tough := effectstest.PtOf(t, g, id); p != 5 || tough != 5 {
			t.Errorf("%s is %d/%d, want 5/5", name, p, tough)
		}
	}
	g2 := effectstest.NewCatalogGame(t)
	x := effectstest.B12Push(g2, g2.Seats[0].ID, "X", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g2, "Bounty of Might", "Instant", oldB190BountyOfMight, effectstest.CardTargets(x, x, x))
	effectstest.PassPriorityAroundTable(t, g2)
	if p, tough := effectstest.PtOf(t, g2, x); p != 11 || tough != 11 {
		t.Errorf("a creature targeted three times is %d/%d, want 11/11", p, tough)
	}
}

func TestOldB190ErstwhileTrooperPumpsOncePerTurnForACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	trooper := effectstest.B12Push(g, me.ID, "Erstwhile Trooper", "Creature — Zombie Soldier", oldB190ErstwhileTroop, 2, 2)
	sorcery := effectstest.HandCard(me, "Not a Creature", "Sorcery")
	if err := g.ActivateCatalogAbility(me.ID, trooper, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{sorcery}}); err == nil {
		t.Fatal("the Trooper discarded a noncreature card")
	}
	bear := effectstest.HandCard(me, "Bear", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, trooper, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the discard is a cost, paid on activation")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, trooper); p != 4 || tough != 4 {
		t.Errorf("the Trooper is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, trooper, "trample") {
		t.Error("the Trooper has no trample")
	}
	bear2 := effectstest.HandCard(me, "Bear Two", "Creature — Bear")
	if err := g.ActivateCatalogAbility(me.ID, trooper, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear2}}); err == nil {
		t.Fatal("a second activation in the same turn was accepted")
	}
}

func TestOldB190ArtfulTakedownRunsEitherOrBothModes(t *testing.T) {
	for _, tc := range []struct {
		name       string
		modes      []int
		wantTapped bool
		wantP      int
		wantTou    int
	}{
		{"tap only", []int{0}, true, 3, 5},
		{"shrink only", []int{1}, false, 1, 1},
		{"both on the same creature", []int{0, 1}, true, 1, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			opp := g.Seats[1]
			giant := effectstest.B12Push(g, opp.ID, "Giant", "Creature — Giant", "", 3, 5)
			var targets []game.TargetRef
			for i := range tc.modes {
				targets = append(targets, effectstest.P1g3ModeCard(giant, i))
			}
			effectstest.B02cCastModalSpell(t, g, "Artful Takedown", "Instant", oldB190ArtfulTakedown, tc.modes, targets)
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.HmlTapped(t, g, giant); got != tc.wantTapped {
				t.Errorf("tapped=%v, want %v", got, tc.wantTapped)
			}
			if p, tough := effectstest.PtOf(t, g, giant); p != tc.wantP || tough != tc.wantTou {
				t.Errorf("the Giant is %d/%d, want %d/%d", p, tough, tc.wantP, tc.wantTou)
			}
		})
	}
}

func TestOldB190ArtfulTakedownTargetsEachModeSeparately(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Push(g, opp.ID, "A", "Creature — Giant", "", 3, 5)
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Giant", "", 3, 5)
	effectstest.B02cCastModalSpell(t, g, "Artful Takedown", "Instant", oldB190ArtfulTakedown, []int{0, 1},
		[]game.TargetRef{effectstest.P1g3ModeCard(a, 0), effectstest.P1g3ModeCard(b, 1)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, a) || effectstest.HmlTapped(t, g, b) {
		t.Error("only the first creature should be tapped")
	}
	if p, tough := effectstest.PtOf(t, g, a); p != 3 || tough != 5 {
		t.Errorf("A is %d/%d, want 3/5", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, b); p != 1 || tough != 1 {
		t.Errorf("B is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB190RiskFactorOffersTheDamageToTheTargetOpponentOnly(t *testing.T) {
	for _, takes := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
		hand, life, thirdLife := me.Hand.Size(), opp.Life, third.Life
		effectstest.CastCatalogSpell(t, g, "Risk Factor", "Instant", oldB190RiskFactor, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
		effectstest.PassPriorityAroundTable(t, g)
		link := effectstest.AppOpenLink(g)
		if link == nil || link.Chooser != opp.ID || link.PayAction() == nil || link.PayAction().Kind != game.PayActionDamage {
			t.Fatalf("takes=%v: the open link is %+v, want a damage offer to the target opponent", takes, link)
		}
		var payers []uuid.UUID
		if takes {
			payers = []uuid.UUID{opp.ID}
		}
		asked := effectstest.AppAnswerOffer(t, g, payers...)
		if len(asked) != 1 || asked[0] != opp.ID {
			t.Errorf("takes=%v: asked %v, want only the target opponent", takes, asked)
		}
		if takes {
			if opp.Life != life-4 {
				t.Errorf("the opponent is at %d life, want %d", opp.Life, life-4)
			}
			if me.Hand.Size() != hand {
				t.Errorf("hand %d, want %d: taking the damage stops the draw", me.Hand.Size(), hand)
			}
		} else {
			if opp.Life != life {
				t.Errorf("the opponent lost life although they declined")
			}
			if me.Hand.Size() != hand+3 {
				t.Errorf("hand %d, want %d: declining draws three", me.Hand.Size(), hand+3)
			}
		}
		if third.Life != thirdLife {
			t.Error("a player who was not the target was affected")
		}
	}
}

func TestOldB190JumpStartCardsCastFromTheGraveyardAndAreExiled(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		targets                func(me, opp *game.Player, mine, theirs uuid.UUID) []game.TargetRef
	}{
		{"Risk Factor", "Instant", oldB190RiskFactor, func(_, opp *game.Player, _, _ uuid.UUID) []game.TargetRef {
			return []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
		}},
		{"Beacon Bolt", "Sorcery", oldB190BeaconBolt, func(_, _ *game.Player, _, theirs uuid.UUID) []game.TargetRef {
			return effectstest.CardTargets(theirs)
		}},
		{"Gravitic Punch", "Sorcery", oldB190GraviticPunch, func(_, opp *game.Player, mine, _ uuid.UUID) []game.TargetRef {
			return []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetPlayer, ID: opp.ID}}
		}},
		{"Sonic Assault", "Instant", oldB190SonicAssault, func(_, _ *game.Player, _, theirs uuid.UUID) []game.TargetRef {
			return effectstest.CardTargets(theirs)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			effectstest.AdvanceToMain(t, g)
			mine := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
			theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
			id := effectstest.SeedGraveyardCard(t, g, tc.name, tc.typeLine, tc.oracle)
			fodder := effectstest.HandCard(me, "Fodder", "Sorcery")
			effectstest.S39Fund(g, me)
			params := game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "jump-start", Targets: tc.targets(me, opp, mine, theirs)}
			if err := g.CastSpell(me.ID, id, params); err == nil {
				t.Fatal("jump-start was castable without discarding a card")
			}
			params.AltCostIDs = []uuid.UUID{fodder}
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("jump-start cast: %v", err)
			}
			if !me.Graveyard.Contains(fodder) {
				t.Error("the discard is a cost, paid as the spell is cast")
			}
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AppAnswerOffer(t, g, opp.ID)
			if !g.Exile.Contains(id) {
				t.Error("the jump-started spell is not exiled")
			}
		})
	}
}

func TestOldB190BeaconBoltCountsInstantsAndSorceriesInYourGraveyardAndExile(t *testing.T) {
	for _, tc := range []struct {
		name    string
		exiled  int
		dies    bool
		graveIS int
	}{
		{"two cards is not enough for toughness 3", 0, false, 2},
		{"graveyard and exile together make three", 1, true, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			effectstest.AdvanceToMain(t, g)
			victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Ogre", "", 3, 3)
			for i := 0; i < tc.graveIS; i++ {
				effectstest.PushGraveyardCardTyped(me, "Old Spell", "Instant")
			}
			effectstest.PushGraveyardCardTyped(me, "Old Creature", "Creature — Bear") // not counted
			effectstest.PushGraveyardCardTyped(opp, "Their Spell", "Sorcery")         // not yours
			for i := 0; i < tc.exiled; i++ {
				g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Exiled Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			}
			g.Exile.PushTop(game.Card{InstanceID: uuid.New(), Name: "Their Exiled Spell", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
			effectstest.CastCatalogSpell(t, g, "Beacon Bolt", "Sorcery", oldB190BeaconBolt, effectstest.CardTarget(victim))
			effectstest.PassPriorityAroundTable(t, g)
			if effects.OnBattlefield(g, victim) == tc.dies {
				t.Errorf("victim on battlefield = %v, want %v", !tc.dies, tc.dies)
			}
		})
	}
}

func TestOldB190GraviticPunchHitsAPlayerForTheCreaturesPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Push(g, me.ID, "Ogre", "Creature — Ogre", "", 4, 4)
	theirs := effectstest.B12Push(g, opp.ID, "Their Ogre", "Creature — Ogre", "", 5, 5)
	if err := effectstest.CastCatalogSpellErr(t, g, "Gravitic Punch", "Sorcery", oldB190GraviticPunch,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetPlayer, ID: opp.ID}}); err == nil {
		t.Fatal("Gravitic Punch used a creature you don't control")
	}
	life, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Gravitic Punch", "Sorcery", oldB190GraviticPunch,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("the player is at %d life, want %d", opp.Life, life-4)
	}
	if me.Life != myLife {
		t.Error("the caster lost life")
	}
}

func TestOldB190SonicAssaultTapsTheCreatureAndHurtsItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	life, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Sonic Assault", "Instant", oldB190SonicAssault, effectstest.CardTarget(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HmlTapped(t, g, theirs) {
		t.Error("the creature is not tapped")
	}
	if opp.Life != life-2 {
		t.Errorf("the controller is at %d life, want %d", opp.Life, life-2)
	}
	if me.Life != myLife {
		t.Error("the caster lost life")
	}
}

func TestOldB190SwarmGuildmageRalliesYourCreaturesAndGainsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	mage := effectstest.B12Push(g, me.ID, "Swarm Guildmage", "Creature — Elf Shaman", oldB190SwarmGuildmage, 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.FloatForTest(g, me, "RRRRR")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("paid {4}{B} with no black mana")
	}
	effectstest.FloatForTest(g, me, "BBBBB")
	if err := g.ActivateCatalogAbility(me.ID, mage, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mage, bear} {
		if p, _ := effectstest.PtOf(t, g, id); p != 3 || !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
			t.Errorf("my creature has power %d / menace %v, want 3 and menace", p, effectstest.HasEffectiveKeyword(t, g, id, "menace"))
		}
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "menace") {
		t.Error("an opponent's creature got the bonus")
	}
	if !effectstest.HmlTapped(t, g, mage) {
		t.Error("the Guildmage is not tapped")
	}
}

func TestOldB190SwarmGuildmageGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	mage := effectstest.B12Push(g, me.ID, "Swarm Guildmage", "Creature — Elf Shaman", oldB190SwarmGuildmage, 2, 2)
	life := me.Life
	effectstest.FloatForTest(g, me, "GG")
	if err := g.ActivateCatalogAbility(me.ID, mage, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
}

func TestOldB190LedevChampionGrowsForEachCreatureYouTap(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	champ := effectstest.B12Push(g, me.ID, "Ledev Champion", "Creature — Elf Knight", oldB190LedevChampion, 2, 2)
	one := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 1, 1)
	two := effectstest.B12Push(g, me.ID, "Bear Two", "Creature — Bear", "", 1, 1)
	skipped := effectstest.B12Push(g, me.ID, "Bear Three", "Creature — Bear", "", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, champ)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, one, two)
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, champ); p != 4 || tough != 4 {
		t.Errorf("the Champion is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HmlTapped(t, g, one) || !effectstest.HmlTapped(t, g, two) {
		t.Error("the chosen creatures are not tapped")
	}
	if effectstest.HmlTapped(t, g, skipped) {
		t.Error("an unchosen creature was tapped")
	}
}

func TestOldB190LedevChampionMakesALifelinkSoldier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	champ := effectstest.B12Push(g, me.ID, "Ledev Champion", "Creature — Elf Knight", oldB190LedevChampion, 2, 2)
	effectstest.FloatForTest(g, me, "GWGGG")
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	soldiers := effectstest.TokensNamed(g, "Soldier")
	if len(soldiers) != 1 || soldiers[0].Controller != me.ID {
		t.Fatalf("soldiers: %d, want one under your control", len(soldiers))
	}
	if p, tough := effectstest.PtOf(t, g, soldiers[0].InstanceID); p != 1 || tough != 1 || !effectstest.HasEffectiveKeyword(t, g, soldiers[0].InstanceID, "lifelink") {
		t.Errorf("the token is %d/%d, want a 1/1 with lifelink", p, tough)
	}
}

func TestOldB190AttendantOfVraskaGainsLifeOnlyWithAVraskaPlaneswalker(t *testing.T) {
	for _, tc := range []struct {
		name    string
		walker  func(g *game.Game, me, opp *game.Player)
		wantGet int
	}{
		{"no planeswalker", func(*game.Game, *game.Player, *game.Player) {}, 0},
		{"a Vraska you control", func(g *game.Game, me, _ *game.Player) {
			effectstest.B12Permanent(g, me.ID, "Vraska, Test", "Legendary Planeswalker — Vraska")
		}, 3},
		{"a different planeswalker", func(g *game.Game, me, _ *game.Player) {
			effectstest.B12Permanent(g, me.ID, "Jace, Test", "Legendary Planeswalker — Jace")
		}, 0},
		{"an opponent's Vraska", func(g *game.Game, _, opp *game.Player) {
			effectstest.B12Permanent(g, opp.ID, "Vraska, Test", "Legendary Planeswalker — Vraska")
		}, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			effectstest.AdvanceToMain(t, g)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
			att := effectstest.B12Push(g, me.ID, "Attendant of Vraska", "Creature — Zombie Soldier", oldB190AttendantVraska, 3, 3)
			tc.walker(g, me, opp)
			life := me.Life
			effectstest.UsgB3Destroy(t, g, att)
			effectstest.PassPriorityAroundTable(t, g)
			if got := me.Life - life; got != tc.wantGet {
				t.Errorf("gained %d life, want %d", got, tc.wantGet)
			}
		})
	}
}
