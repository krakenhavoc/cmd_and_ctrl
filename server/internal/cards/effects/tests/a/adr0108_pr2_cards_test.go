package testsa

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func p2Multipliers(g *game.Game) []game.Mod {
	var out []game.Mod
	g.WithWriteLock(func() {
		for _, e := range g.ScopedEffects {
			for _, m := range e.Mods {
				if m.Kind == game.ModMultiplyDamage {
					out = append(out, m)
				}
			}
		}
	})
	return out
}

func p2InsultInjuryCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   effectstest.P2InsultInjuryOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Insult", TypeLine: "Sorcery", ManaCost: "{2}{R}",
				OracleText: "Damage can't be prevented this turn. If a source you control would deal damage this turn, it deals double that damage instead."},
			{Name: "Injury", TypeLine: "Sorcery", ManaCost: "{2}{R}",
				OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)\nInjury deals 2 damage to target creature and 2 damage to target player or planeswalker."},
		},
	}
	c.SettleImported()
	return c
}

// Insult: your sources deal double damage, through a prevention shield
// (its own "can't be prevented"); an opponent's source is untouched.
func TestP2InsultDoublesYourSourcesAndBeatsShields(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.P2Tough(g, me.ID, "Mine")
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
	effectstest.AdvanceToMain(t, g)
	c := p2InsultInjuryCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Insult: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(p2Multipliers(g)) != 1 {
		t.Fatalf("Insult made %d multipliers", len(p2Multipliers(g)))
	}
	g.WithWriteLock(func() { g.PreventNextDamageThisTurnForEffect(uuid.Nil, opp.ID, 20, false, "shield") })
	start, mine0 := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, me.ID)
	effectstest.P2Ping(t, g, mine, opp.ID, 3)
	effectstest.P2Ping(t, g, theirs, me.ID, 3)
	if got := effectstest.LifeOf(g, opp.ID); got != start-6 {
		t.Errorf("opponent life %d, want %d: doubled, and can't be prevented", got, start-6)
	}
	if got := effectstest.LifeOf(g, me.ID); got != mine0-3 {
		t.Errorf("your life %d, want %d: an opponent's source is not doubled", got, mine0-3)
	}
}

// Injury: aftermath, from the graveyard — 2 to the creature and 2 to the
// player, then exiled.
func TestP2InjuryFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	bear := effectstest.P2Tough(g, opp.ID, "Bear")
	effectstest.AdvanceToMain(t, g)
	c := p2InsultInjuryCard(me.ID)
	me.Graveyard.PushTop(c)
	start := effectstest.LifeOf(g, opp.ID)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, FromZone: "graveyard",
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetPlayer, ID: opp.ID, Slot: 1}}}); err != nil {
		t.Fatalf("cast Injury: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("creature has %d damage, want 2", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != start-2 {
		t.Errorf("player life %d, want %d", got, start-2)
	}
	if !g.Exile.Contains(c.InstanceID) {
		t.Error("Injury is not exiled after resolving from the graveyard (aftermath)")
	}
}

// Isengard Unleashed: triple to an opponent or an opponent's permanent;
// your own permanent is dealt damage as printed. Flashback casts it again.
func TestP2IsengardTriplesOnlyTowardOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := effectstest.P2Tough(g, me.ID, "Mine")
	other := effectstest.P2Tough(g, me.ID, "Other")
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
	isengard := effectstest.CastCatalogSpell(t, g, "Isengard Unleashed", "Sorcery", effectstest.P2IsengardOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	start := effectstest.LifeOf(g, opp.ID)
	effectstest.P2Ping(t, g, mine, opp.ID, 2)
	effectstest.P2Ping(t, g, mine, theirs, 2)
	effectstest.P2Ping(t, g, mine, other, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != start-6 {
		t.Errorf("opponent life %d, want %d", got, start-6)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 6 {
		t.Errorf("opponent's creature has %d damage, want 6", got)
	}
	if got := effectstest.DamageMarkedOn(g, other); got != 2 {
		t.Errorf("your creature has %d damage, want 2", got)
	}
	if !me.Graveyard.Contains(isengard) {
		t.Fatal("Isengard Unleashed is not in the graveyard")
	}
	if err := g.CastSpell(me.ID, isengard, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := len(p2Multipliers(g)); got != 2 {
		t.Errorf("%d multipliers after the flashback, want 2", got)
	}
	if !g.Exile.Contains(isengard) {
		t.Error("the flashed-back Isengard is not exiled")
	}
}

// Goblin Goliath: a Goblin per opponent; the ability doubles your
// sources' damage to opponents, not to their creatures.
func TestP2GoblinGoliath(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.CastCatalogSpell(t, g, "Goblin Goliath", "Creature — Goblin Mutant", effectstest.P2GoblinGoliathOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	goblins := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.Name == "Goblin" {
			goblins++
		}
	}
	if want := len(g.Seats) - 1; goblins != want {
		t.Fatalf("%d Goblins, want %d (one per opponent)", goblins, want)
	}

	goliath := effectstest.PushCatalogPermanent(g, me.ID, "Goblin Goliath", "Creature — Goblin Mutant", effectstest.P2GoblinGoliathOracle, false)
	effectstest.FillPoolColored(me, "R", 4)
	if err := g.ActivateCatalogAbility(me.ID, goliath, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	mine := effectstest.P2Tough(g, me.ID, "Mine")
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
	start := effectstest.LifeOf(g, opp.ID)
	effectstest.P2Ping(t, g, mine, opp.ID, 2)
	effectstest.P2Ping(t, g, mine, theirs, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != start-4 {
		t.Errorf("opponent life %d, want %d", got, start-4)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 2 {
		t.Errorf("opponent's creature has %d damage, want 2", got)
	}
}

// Blind Fury: every creature loses trample (a locked set — one that
// enters later keeps it), and combat damage creature to creature is
// doubled; combat damage to a player is not.
func TestP2BlindFury(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	trampler := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Trampler", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 30, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID,
	})
	effectstest.CastCatalogSpell(t, g, "Blind Fury", "Instant", effectstest.P2BlindFuryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	late := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Late Trampler", TypeLine: "Creature — Beast",
		Power: 3, Toughness: 3, Keywords: []string{"trample"}, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		if c, ok := g.LookupCardForEffect(trampler); !ok || game.HasKeyword(&c, "trample") {
			t.Error("the trampler kept trample")
		}
		if c, ok := g.LookupCardForEffect(late); !ok || !game.HasKeyword(&c, "trample") {
			t.Error("a creature that entered later lost trample (CR 611.2c locks the set)")
		}
	})
	mods := p2Multipliers(g)
	if len(mods) != 1 || !mods[0].CombatOnly || mods[0].Sources != game.DamageSourcesCreatures || mods[0].Recipients != game.DamageRecipientsCreatures {
		t.Fatalf("multiplier %+v", mods)
	}
	blocker := effectstest.P2Tough(g, opp.ID, "Blocker")
	effectstest.P2Ping(t, g, trampler, blocker, 2)
	if got := effectstest.DamageMarkedOn(g, blocker); got != 2 {
		t.Errorf("non-combat damage doubled: %d", got)
	}
}

// Desperate Gambit: the source is one you control, then a flip — a
// doubled next instance on a win, a prevented one on a loss.
func TestP2DesperateGambitBothBranches(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		mine := effectstest.P2Tough(g, me.ID, "Mine")
		theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
		effectstest.CastCatalogSpell(t, g, "Desperate Gambit", "Instant", effectstest.P2DesperateOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		pick := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseSource)
		if pick == nil {
			t.Fatal("no choose_source prompt")
		}
		if slices.Contains(pick.ChooseCards, theirs) || !slices.Contains(pick.ChooseCards, mine) {
			t.Fatalf("candidates %v: want your sources only", pick.ChooseCards)
		}
		if err := g.ResolveChooseSource(pick.ID, me.ID, []uuid.UUID{mine}); err != nil {
			t.Fatalf("choose: %v", err)
		}
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		start := effectstest.LifeOf(g, opp.ID)
		effectstest.P2Ping(t, g, mine, opp.ID, 3)
		effectstest.P2Ping(t, g, mine, opp.ID, 3)
		got := effectstest.LifeOf(g, opp.ID)
		if won {
			sawWin = true
			if got != start-9 {
				t.Errorf("won: life %d, want %d (6, then 3)", got, start-9)
			}
		} else {
			sawLoss = true
			if got != start-3 {
				t.Errorf("lost: life %d, want %d (prevented, then 3)", got, start-3)
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("did not see both branches (win %v, loss %v)", sawWin, sawLoss)
	}
}

// Impulsive Maneuvers: an attacking creature's next COMBAT damage is
// doubled or prevented; its non-combat damage is left alone.
func TestP2ImpulsiveManeuversAttackerFlip(t *testing.T) {
	sawWin, sawLoss := false, false
	for i := 0; i < 64 && !(sawWin && sawLoss); i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		effectstest.B12Push(g, me.ID, "Impulsive Maneuvers", "Enchantment", effectstest.P2ManeuversOracle, 0, 0)
		attacker := effectstest.PushVanillaCreature(g, me.ID, "Attacker", 3, 3)
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		var flip *game.PendingChoice
		for j := 0; j < 6 && flip == nil; j++ {
			if flip = effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall); flip == nil {
				effectstest.PassPriorityAroundTable(t, g)
			}
		}
		if flip == nil {
			t.Fatal("the attack asked for no coin call")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		start := effectstest.LifeOf(g, opp.ID)
		effectstest.P2Ping(t, g, attacker, opp.ID, 1)
		if got := effectstest.LifeOf(g, opp.ID); got != start-1 {
			t.Fatalf("non-combat damage changed: life %d, want %d", got, start-1)
		}
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.AdvanceTo(t, g, game.StepEndCombat)
		got := effectstest.LifeOf(g, opp.ID)
		if won {
			sawWin = true
			if got != start-7 {
				t.Errorf("won: life %d, want %d (1, then 6 combat)", got, start-7)
			}
		} else {
			sawLoss = true
			if got != start-1 {
				t.Errorf("lost: life %d, want %d (combat damage prevented)", got, start-1)
			}
		}
	}
	if !sawWin || !sawLoss {
		t.Fatalf("did not see both branches (win %v, loss %v)", sawWin, sawLoss)
	}
}

// Quest for Pure Flame: a counter for each time a source you control
// deals damage to an opponent (not to an opponent's creature, not to
// you); four counters and a sacrifice double your sources this turn.
func TestP2QuestForPureFlame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	quest := effectstest.B12Push(g, me.ID, "Quest for Pure Flame", "Enchantment", effectstest.P2QuestOracle, 0, 0)
	mine := effectstest.P2Tough(g, me.ID, "Mine")
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
	effectstest.P2Ping(t, g, mine, theirs, 1)
	effectstest.P2Ping(t, g, mine, me.ID, 1)
	if effectstest.OwedMays(g, me.ID) > 0 {
		t.Fatal("damage to a creature or to you asked for a quest counter")
	}
	for i := 0; i < 4; i++ {
		effectstest.P2Ping(t, g, mine, opp.ID, 1)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, quest, "quest"); got != 4 {
		t.Fatalf("%d quest counters, want 4", got)
	}
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, quest, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, quest) {
		t.Error("the Quest was not sacrificed")
	}
	start := effectstest.LifeOf(g, opp.ID)
	effectstest.P2Ping(t, g, mine, theirs, 2)
	if got := effectstest.DamageMarkedOn(g, theirs); got != 1+4 {
		t.Errorf("creature has %d damage, want 5 (1 earlier, then 2 doubled)", got)
	}
	if got := effectstest.LifeOf(g, opp.ID); got != start {
		t.Errorf("life moved without damage: %d, want %d", got, start)
	}
}

// Lightning, Army of One: its combat damage to a player doubles all
// damage to that player and their permanents until your next turn.
func TestP2LightningStagger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	lightning := effectstest.PushCatalogPermanent(g, me.ID, "Lightning, Army of One", "Legendary Creature — Human Soldier", effectstest.P2LightningOracle, false)
	effectstest.AttackWith(t, g, opp.ID, lightning)
	effectstest.PassPriorityAroundTable(t, g)
	mods := p2Multipliers(g)
	if len(mods) != 1 || mods[0].Player != opp.ID || mods[0].Recipients != game.DamageRecipientsPlayerAndTheirPermanents {
		t.Fatalf("multiplier %+v", mods)
	}
	theirs := effectstest.P2Tough(g, opp.ID, "Theirs")
	thirds := effectstest.P2Tough(g, third.ID, "Third's")
	start, thirdStart := effectstest.LifeOf(g, opp.ID), effectstest.LifeOf(g, third.ID)
	effectstest.P2Ping(t, g, thirds, opp.ID, 2)
	effectstest.P2Ping(t, g, thirds, theirs, 2)
	effectstest.P2Ping(t, g, theirs, third.ID, 2)
	if got := effectstest.LifeOf(g, opp.ID); got != start-4 {
		t.Errorf("that player's life %d, want %d", got, start-4)
	}
	if got := effectstest.DamageMarkedOn(g, theirs); got != 4 {
		t.Errorf("their creature has %d damage, want 4", got)
	}
	if got := effectstest.LifeOf(g, third.ID); got != thirdStart-2 {
		t.Errorf("another player's life %d, want %d", got, thirdStart-2)
	}
}
