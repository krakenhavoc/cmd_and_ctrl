package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every "prevent all combat damage that would be dealt by target
// [attacking] creature this turn" card: the targeted attacker deals no
// combat damage, and its non-combat damage is still dealt.
func TestADR0108PR7bCombatShieldsAgainstTheTarget(t *testing.T) {
	spells := []struct{ name, oracle string }{
		{"Fend Off", effectstest.Pr7bFendOff},
		{"Warning", effectstest.Pr7bWarning},
		{"Restrain", effectstest.Pr7bRestrain},
		{"Subdue", effectstest.Pr7bSubdue},
		{"Boros Fury-Shield", effectstest.Pr7bBorosFuryShield},
	}
	for _, s := range spells {
		t.Run(s.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
			effectstest.Pr7bAttack(t, g, opp.ID, att)
			effectstest.Pr7bCastNow(t, g, s.name, "Instant", s.oracle, game.CastSpellParams{Targets: effectstest.Pr7bTargets(att)})
			life := opp.Life
			effectstest.Pr7bDamageStep(t, g)
			if opp.Life != life {
				t.Fatalf("defender life %d, want %d: the attacker's combat damage is prevented", opp.Life, life)
			}
			effectstest.Pr7Hit(t, g, att, opp.ID, 1)
			if opp.Life != life-1 {
				t.Fatalf("defender life %d, want %d: non-combat damage is not prevented", opp.Life, life-1)
			}
		})
	}
	rows := []struct {
		name, typeLine, oracle string
		params                 func(g *game.Game, me uuid.UUID) game.ActivateAbilityParams
	}{
		{"Kor Haven", "Legendary Land", effectstest.Pr7bKorHaven, nil},
		{"Horn of Deafening", "Artifact", effectstest.Pr7bHorn, nil},
		{"Safeguard", "Enchantment", effectstest.Pr7bSafeguard, nil},
		{"Lady Evangela", "Legendary Creature — Human Cleric", effectstest.Pr7bLadyEvangela, nil},
		{"Resistance Fighter", "Creature — Human Soldier", effectstest.Pr7bResistance, nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := effectstest.PushCatalogPermanent(g, me.ID, r.name, r.typeLine, r.oracle, false)
			att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
			effectstest.Pr7bAttack(t, g, opp.ID, att)
			effectstest.Pr7Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(att)})
			life := opp.Life
			effectstest.Pr7bDamageStep(t, g)
			if opp.Life != life {
				t.Fatalf("defender life %d, want %d", opp.Life, life)
			}
		})
	}
}

// Songstitcher targets only an attacker with flying.
func TestADR0108PR7bSongstitcherNeedsAFlyingAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Songstitcher", "Creature — Human Cleric", effectstest.Pr7bSongstitcher, false)
	flier := effectstest.Pr7bCreature(g, me.ID, "Flier", 3, "", []string{"W"}, "flying")
	walker := effectstest.Pr7Creature(g, me.ID, "Walker", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, flier, walker)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(walker)}); err == nil {
		t.Fatal("a non-flying attacker was accepted as the target")
	}
	effectstest.Pr7Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(flier)})
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life-2 {
		t.Fatalf("defender life %d, want %d: only the walker's 2 is dealt", opp.Life, life-2)
	}
}

// Restrain draws; Subdue and Kry Shield add the creature's mana value
// to its toughness; Kry Shield prevents ALL the creature's damage.
func TestADR0108PR7bRidersOnTheTargetedShield(t *testing.T) {
	t.Run("Restrain draws", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
		effectstest.Pr7bAttack(t, g, opp.ID, att)
		hand := me.Hand.Size()
		effectstest.Pr7bCastNow(t, g, "Restrain", "Instant", effectstest.Pr7bRestrain, game.CastSpellParams{Targets: effectstest.Pr7bTargets(att)})
		if me.Hand.Size() != hand+1 {
			t.Fatalf("hand %d, want %d (Restrain put in and cast, one card drawn)", me.Hand.Size(), hand+1)
		}
	})
	t.Run("Subdue toughness", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		c := effectstest.Pr7bCreature(g, me.ID, "Three Drop", 3, "{2}{G}", []string{"G"})
		effectstest.CastCatalogSpell(t, g, "Subdue", "Instant", effectstest.Pr7bSubdue, effectstest.Pr7bTargets(c))
		effectstest.PassPriorityAroundTable(t, g)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		if got := effectstest.FindBattlefieldCardForTest(g, c).CurrentToughness(); got != 7 {
			t.Fatalf("toughness %d, want 7 (4 + mana value 3)", got)
		}
	})
	t.Run("Kry Shield", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		shield := effectstest.PushCatalogPermanent(g, me.ID, "Kry Shield", "Artifact", effectstest.Pr7bKryShield, false)
		c := effectstest.Pr7bCreature(g, me.ID, "Two Drop", 2, "{1}{R}", []string{"R"})
		effectstest.Pr7Activate(t, g, me.ID, shield, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(c)})
		life := opp.Life
		effectstest.Pr7Hit(t, g, c, opp.ID, 2)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		if opp.Life != life || effectstest.FindBattlefieldCardForTest(g, c).CurrentToughness() != 6 {
			t.Fatalf("life %d (want %d), toughness %d (want 6)", opp.Life, life, effectstest.FindBattlefieldCardForTest(g, c).CurrentToughness())
		}
	})
}

// Soul Parry shields one or two sources, all their damage; a third
// source is untouched.
func TestADR0108PR7bSoulParryShieldsBothTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, opp.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, opp.ID, "B", 2, "R")
	c := effectstest.Pr7Creature(g, opp.ID, "C", 2, "R")
	effectstest.CastCatalogSpell(t, g, "Soul Parry", "Instant", effectstest.Pr7bSoulParry, effectstest.Pr7bTargets(a, b))
	effectstest.PassPriorityAroundTable(t, g)
	life := me.Life
	effectstest.Pr7Hit(t, g, a, me.ID, 2)
	effectstest.Pr7Hit(t, g, b, me.ID, 2)
	effectstest.Pr7Hit(t, g, c, me.ID, 2)
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d", me.Life, life-2)
	}
}

// Inquisitor's Snare destroys a red target and still prevents its
// damage as it last existed; a green one lives, shielded.
func TestADR0108PR7bInquisitorsSnare(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := effectstest.Pr7Creature(g, me.ID, "Red", 3, "R")
	green := effectstest.Pr7Creature(g, me.ID, "Green", 3, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, red, green)
	effectstest.Pr7bCastNow(t, g, "Inquisitor's Snare", "Instant", effectstest.Pr7bSnare, game.CastSpellParams{Targets: effectstest.Pr7bTargets(red)})
	if effectstest.FindBattlefieldCardForTest(g, red) != nil {
		t.Fatal("the red creature was not destroyed")
	}
	effectstest.Pr7bCastNow(t, g, "Inquisitor's Snare", "Instant", effectstest.Pr7bSnare, game.CastSpellParams{Targets: effectstest.Pr7bTargets(green)})
	if effectstest.FindBattlefieldCardForTest(g, green) == nil {
		t.Fatal("the green creature was destroyed")
	}
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	effectstest.Pr7Hit(t, g, green, opp.ID, 2)
	if opp.Life != life {
		t.Fatalf("defender life %d, want %d", opp.Life, life)
	}
}

// Hallow prevents a spell's damage and gains its caster that much life.
func TestADR0108PR7bHallowStopsTheSpellAndGainsTheLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.Pr7bLightningBolt,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, HoldPriority: true}); err != nil {
		t.Fatal(err)
	}
	life, mine := opp.Life, me.Life
	effectstest.Pr7bCastNow(t, g, "Hallow", "Instant", effectstest.Pr7bHallow, game.CastSpellParams{Targets: effectstest.Pr7bTargets(bolt)})
	if opp.Life != life || me.Life != mine+3 {
		t.Fatalf("opponent %d (want %d), me %d (want %d)", opp.Life, life, me.Life, mine+3)
	}
}

// Dromoka's Command: the counter lands before the fight; the
// sacrifice is the target player's choice.
func TestADR0108PR7bDromokasCommand(t *testing.T) {
	t.Run("counter then fight", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := effectstest.Pr7Creature(g, me.ID, "Mine", 2, "G")
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 1, "B")
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dromoka's Command", TypeLine: "Instant", OracleID: effectstest.Pr7bDromoka,
			Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2, 3}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Mode: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1, Slot: 0},
			{Kind: game.TargetCard, ID: theirs, Mode: 1, Slot: 1},
		}})
		if err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.Pr6Marked(g, theirs); got != 3 {
			t.Fatalf("their creature took %d, want 3 (2 power plus the counter)", got)
		}
	})
	t.Run("spell shield and sacrifice", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ench := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Some Enchantment", TypeLine: "Enchantment"})
		effectstest.AdvanceTo(t, g, game.StepPrecombatMain)
		bolt := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: effectstest.Pr7bLightningBolt,
			Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, HoldPriority: true}); err != nil {
			t.Fatal(err)
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dromoka's Command", TypeLine: "Instant", OracleID: effectstest.Pr7bDromoka,
			Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: bolt, Mode: 0},
			{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1},
		}})
		if err != nil {
			t.Fatal(err)
		}
		life := opp.Life
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceOwnPermanents)
		if c.Chooser != opp.ID {
			t.Fatalf("the sacrifice is asked of %v, want the target player", c.Chooser)
		}
		if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{ench}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.FindBattlefieldCardForTest(g, ench) != nil || opp.Life != life {
			t.Fatalf("enchantment still there %v, life %d (want %d)", effectstest.FindBattlefieldCardForTest(g, ench) != nil, opp.Life, life)
		}
	})
}

// Azorius Ploy: one creature's combat damage, and combat damage to
// another, are prevented.
func TestADR0108PR7bAzoriusPloy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
	blk := effectstest.Pr7Creature(g, opp.ID, "Blocker", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
	effectstest.Pr7bCastNow(t, g, "Azorius Ploy", "Instant", effectstest.Pr7bAzoriusPloy, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: blk, Slot: 0},
		{Kind: game.TargetCard, ID: blk, Slot: 1},
	}})
	effectstest.Pr7bDamageStep(t, g)
	if effectstest.Pr6Marked(g, blk) != 0 || effectstest.Pr6Marked(g, att) != 0 {
		t.Fatalf("blocker %d, attacker %d: both want 0", effectstest.Pr6Marked(g, blk), effectstest.Pr6Marked(g, att))
	}
}

// Shieldmage Elder taps two Clerics (itself among them) to shield
// against a creature.
func TestADR0108PR7bShieldmageElderTapsTwoClerics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Shieldmage Elder", TypeLine: "Creature — Human Cleric Wizard",
		OracleID: effectstest.Pr7bShieldmageElder, Power: 2, Toughness: 3})
	cleric := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Cleric", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.Pr7Activate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(src), TapIDs: []uuid.UUID{elder, cleric}})
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life || !effectstest.FindBattlefieldCardForTest(g, cleric).Tapped {
		t.Fatalf("life %d (want %d), cleric tapped %v", me.Life, life, effectstest.FindBattlefieldCardForTest(g, cleric).Tapped)
	}
}

// Stonewise Fortifier protects only itself from only the target.
func TestADR0108PR7bStonewiseFortifier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fort := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Stonewise Fortifier", TypeLine: "Creature — Human Wizard",
		OracleID: effectstest.Pr7bStonewise, Power: 2, Toughness: 4})
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	other := effectstest.Pr7Creature(g, opp.ID, "Other", 3, "R")
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.Pr7Activate(t, g, me.ID, fort, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(src)})
	effectstest.Pr6Damage(t, g, src, fort, 2)
	effectstest.Pr6Damage(t, g, src, bear, 1)
	effectstest.Pr6Damage(t, g, other, fort, 1)
	if effectstest.Pr6Marked(g, fort) != 1 || effectstest.Pr6Marked(g, bear) != 1 {
		t.Fatalf("fortifier %d (want 1), bear %d (want 1)", effectstest.Pr6Marked(g, fort), effectstest.Pr6Marked(g, bear))
	}
}

// Ignoble Soldier, blocked, deals no combat damage and still takes it.
func TestADR0108PR7bIgnobleSoldierBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Ignoble Soldier", TypeLine: "Creature — Human Soldier",
		OracleID: effectstest.Pr7bIgnoble, Power: 3, Toughness: 4})
	blk := effectstest.Pr7Creature(g, opp.ID, "Blocker", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, soldier)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: soldier})
	effectstest.Pr7bDamageStep(t, g)
	if effectstest.Pr6Marked(g, blk) != 0 || effectstest.Pr6Marked(g, soldier) != 2 {
		t.Fatalf("blocker %d (want 0), soldier %d (want 2)", effectstest.Pr6Marked(g, blk), effectstest.Pr6Marked(g, soldier))
	}
}

// Zealot il-Vec, unblocked, may ping a creature; if it does, its own
// combat damage is prevented.
func TestADR0108PR7bZealotIlVec(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		zealot := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Zealot il-Vec", TypeLine: "Creature — Human Rebel",
			OracleID: effectstest.Pr7bZealot, Power: 1, Toughness: 1})
		victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2, "G")
		effectstest.Pr7bAttack(t, g, opp.ID, zealot)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		effectstest.FinishBlockDeclarations(t, g)
		for i := 0; i < 8; i++ {
			if c := effectstest.FirstPendingTarget(g); c != nil {
				break
			}
			if err := g.PassPriority(); err != nil {
				break
			}
		}
		if c := effectstest.FirstPendingTarget(g); c != nil {
			if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
				t.Fatal(err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceConfirm)
		if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		life := opp.Life
		effectstest.Pr7bDamageStep(t, g)
		switch {
		case yes && (effectstest.Pr6Marked(g, victim) != 1 || opp.Life != life):
			t.Fatalf("yes: victim %d (want 1), life %d (want %d)", effectstest.Pr6Marked(g, victim), opp.Life, life)
		case !yes && (effectstest.Pr6Marked(g, victim) != 0 || opp.Life != life-1):
			t.Fatalf("no: victim %d (want 0), life %d (want %d)", effectstest.Pr6Marked(g, victim), opp.Life, life-1)
		}
	}
}

// Feint taps the blockers and prevents the attacker's and theirs.
func TestADR0108PR7bFeint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
	b1 := effectstest.Pr7Creature(g, opp.ID, "B1", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: att})
	effectstest.Pr7bCastNow(t, g, "Feint", "Instant", effectstest.Pr7bFeint, game.CastSpellParams{Targets: effectstest.Pr7bTargets(att)})
	if !effectstest.FindBattlefieldCardForTest(g, b1).Tapped {
		t.Fatal("the blocker is not tapped")
	}
	effectstest.Pr7bDamageStep(t, g)
	if effectstest.Pr6Marked(g, att)+effectstest.Pr6Marked(g, b1) != 0 {
		t.Fatalf("damage marked: attacker %d, blocker %d — want none", effectstest.Pr6Marked(g, att), effectstest.Pr6Marked(g, b1))
	}
}

// Falling Timber kicked shields two creatures; unkicked takes one.
func TestADR0108PR7bFallingTimberKicked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, me.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, me.ID, "B", 3, "R")
	land := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	effectstest.Pr7bAttack(t, g, opp.ID, a, b)
	effectstest.Pr7bCastNow(t, g, "Falling Timber", "Instant", effectstest.Pr7bFallingTimber, game.CastSpellParams{
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{land},
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}},
	})
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life || effectstest.FindBattlefieldCardForTest(g, land) != nil {
		t.Fatalf("life %d (want %d), land still there %v", opp.Life, life, effectstest.FindBattlefieldCardForTest(g, land) != nil)
	}
}

// Serene Sunset takes X targets.
func TestADR0108PR7bSereneSunset(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := effectstest.Pr7Creature(g, me.ID, "A", 2, "R")
	b := effectstest.Pr7Creature(g, me.ID, "B", 3, "R")
	c := effectstest.Pr7Creature(g, me.ID, "C", 1, "R")
	effectstest.Pr7bAttack(t, g, opp.ID, a, b, c)
	effectstest.Pr7bCastNow(t, g, "Serene Sunset", "Instant", effectstest.Pr7bSereneSunset, game.CastSpellParams{XValue: 2, Targets: effectstest.Pr7bTargets(a, b)})
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life-1 {
		t.Fatalf("life %d, want %d (only C's 1)", opp.Life, life-1)
	}
}

// Chain of Silence shields the creature and offers its controller the
// land-for-a-copy chain.
func TestADR0108PR7bChainOfSilence(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	effectstest.CastCatalogSpell(t, g, "Chain of Silence", "Instant", effectstest.Pr7bChainOfSilence, effectstest.Pr7bTargets(src))
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceConfirm); c.Chooser != opp.ID {
		t.Fatalf("the chain is offered to %v, want the creature's controller", c.Chooser)
	}
	life := me.Life
	effectstest.Pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Ajani's Aid: sacrificed, a creature of your choice deals no combat
// damage; only creatures are offered.
func TestADR0108PR7bAjanisAid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aid := effectstest.PushCatalogPermanent(g, me.ID, "Ajani's Aid", "Enchantment", effectstest.Pr7bAjanisAid, false)
	art := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Pinger", TypeLine: "Artifact", Colors: nil})
	att := effectstest.Pr7Creature(g, me.ID, "Attacker", 3, "R")
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7Activate(t, g, me.ID, aid, 0, game.ActivateAbilityParams{})
	c := effectstest.Pr7SourcePrompt(t, g)
	if effectstest.Pr7Offers(c, art) || !effectstest.Pr7Offers(c, att) {
		t.Fatalf("offered %v: want creatures only", c.ChooseCards)
	}
	effectstest.Pr7Choose(t, g, me.ID, att)
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("life %d, want %d", opp.Life, life)
	}
}

// Fighting Chance shields exactly the blockers whose flips were won.
func TestADR0108PR7bFightingChance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := effectstest.Pr7Creature(g, me.ID, "Attacker", 9, "R")
	b1 := effectstest.Pr7Creature(g, opp.ID, "B1", 2, "G")
	b2 := effectstest.Pr7Creature(g, opp.ID, "B2", 2, "G")
	b3 := effectstest.Pr7Creature(g, opp.ID, "B3", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, att)
	effectstest.Pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: att, b2: att, b3: att})
	effectstest.Pr7bCastNow(t, g, "Fighting Chance", "Instant", effectstest.Pr7bFightingChance, game.CastSpellParams{})
	c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceCoinCall)
	if c.CoinCount != 3 {
		t.Fatalf("coins %d, want one per blocker", c.CoinCount)
	}
	if err := g.ResolveCoinCall(c.ID, me.ID, "heads"); err != nil {
		t.Fatal(err)
	}
	won := 0
	for _, ev := range g.Events {
		if ev.Kind == game.EventFlipCoin && ev.Label == "heads" {
			won++
		}
	}
	if got := effectstest.Pr7bSourceShields(g); got != won {
		t.Fatalf("%d shields for %d won flips", got, won)
	}
}

// Loafing Giant milling a land prevents its own combat damage.
func TestADR0108PR7bLoafingGiantMillsALand(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		giant := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Loafing Giant", TypeLine: "Creature — Giant",
			OracleID: effectstest.Pr7bLoafingGiant, Power: 4, Toughness: 6})
		top := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if land {
			top.Name, top.TypeLine = "Forest", "Basic Land — Forest"
		}
		me.Library.PushTop(top)
		effectstest.Pr7bAttack(t, g, opp.ID, giant)
		life := opp.Life
		effectstest.Pr7bDamageStep(t, g)
		want := life - 4
		if land {
			want = life
		}
		if opp.Life != want {
			t.Fatalf("land milled %v: life %d, want %d", land, opp.Life, want)
		}
	}
}

// Mtenda Lion: the defending player pays {U} and the Lion's combat
// damage is prevented; declining lets it through.
func TestADR0108PR7bMtendaLion(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lion := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mtenda Lion", TypeLine: "Creature — Cat",
			OracleID: effectstest.Pr7bMtendaLion, Power: 2, Toughness: 1})
		effectstest.Pr7bAttack(t, g, opp.ID, lion)
		opp.ManaPool.AddMana(game.ManaToken{Color: "U"})
		effectstest.AnswerPayUnless(t, g, opp.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		life := opp.Life
		effectstest.Pr7bDamageStep(t, g)
		want := life - 2
		if pay {
			want = life
		}
		if opp.Life != want {
			t.Fatalf("paid %v: life %d, want %d", pay, opp.Life, want)
		}
	}
}

// Heroism: each red attacker's controller pays {2}{R} or its combat
// damage is prevented.
func TestADR0108PR7bHeroism(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heroism := effectstest.PushCatalogPermanent(g, opp.ID, "Heroism", "Enchantment", effectstest.Pr7bHeroism, false)
	white := effectstest.Pr7Creature(g, opp.ID, "White", 1, "W")
	red := effectstest.Pr7Creature(g, me.ID, "Red", 3, "R")
	green := effectstest.Pr7Creature(g, me.ID, "Green", 2, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, red, green)
	effectstest.Pr7Activate(t, g, opp.ID, heroism, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{white}})
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d (only the green 2)", opp.Life, life-2)
	}
}

// Guard Dogs: a shared colour makes the shield, read as it resolves.
func TestADR0108PR7bGuardDogs(t *testing.T) {
	for _, shares := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dogs := effectstest.PushCatalogPermanent(g, me.ID, "Guard Dogs", "Creature — Dog", effectstest.Pr7bGuardDogs, false)
		color := "R"
		if !shares {
			color = "U"
		}
		mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Enchantment", Colors: []string{color}})
		src := effectstest.Pr7Creature(g, opp.ID, "Src", 3, "R")
		effectstest.Pr7Activate(t, g, me.ID, dogs, 0, game.ActivateAbilityParams{Targets: effectstest.Pr7bTargets(src)})
		c := effectstest.Pr7bPrompt(t, g, game.PendingChoiceOwnPermanents)
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if shares {
			want = 1
		}
		if got := effectstest.Pr7bSourceShields(g); got != want {
			t.Fatalf("shares %v: %d shields, want %d", shares, got, want)
		}
	}
}

// Radiant Kavu: blue and black creatures' combat damage only.
func TestADR0108PR7bRadiantKavu(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kavu := effectstest.PushCatalogPermanent(g, me.ID, "Radiant Kavu", "Creature — Kavu", effectstest.Pr7bRadiantKavu, false)
	blue := effectstest.Pr7Creature(g, me.ID, "Blue", 2, "U")
	black := effectstest.Pr7Creature(g, me.ID, "Black", 3, "B")
	green := effectstest.Pr7Creature(g, me.ID, "Green", 1, "G")
	effectstest.Pr7bAttack(t, g, opp.ID, blue, black, green)
	effectstest.Pr7Activate(t, g, me.ID, kavu, 0, game.ActivateAbilityParams{})
	life := opp.Life
	effectstest.Pr7bDamageStep(t, g)
	effectstest.Pr7Hit(t, g, blue, opp.ID, 1)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d (green's 1 and blue's non-combat 1)", opp.Life, life-2)
	}
}

// Luminesce, Ethereal Haze, Chant of Vitu-Ghazi, Scarecrow and Eerie
// Interference: a property shield, checked as the damage is dealt.
func TestADR0108PR7bPropertyShields(t *testing.T) {
	t.Run("Luminesce", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		black := effectstest.Pr7Creature(g, opp.ID, "Black", 2, "B")
		green := effectstest.Pr7Creature(g, opp.ID, "Green", 2, "G")
		effectstest.CastCatalogSpell(t, g, "Luminesce", "Instant", effectstest.Pr7bLuminesce, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, black, me.ID, 2)
		effectstest.Pr7Hit(t, g, green, me.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("life %d, want %d", me.Life, life-2)
		}
	})
	t.Run("Ethereal Haze", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.Pr7Creature(g, opp.ID, "Creature", 2, "R")
		art := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Artifact", TypeLine: "Artifact"})
		effectstest.CastCatalogSpell(t, g, "Ethereal Haze", "Instant — Arcane", effectstest.Pr7bEtherealHaze, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, c, me.ID, 2)
		effectstest.Pr7Hit(t, g, art, me.ID, 1)
		if me.Life != life-1 {
			t.Fatalf("life %d, want %d", me.Life, life-1)
		}
	})
	t.Run("Chant of Vitu-Ghazi", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.Pr7Creature(g, opp.ID, "Creature", 3, "R")
		effectstest.CastCatalogSpell(t, g, "Chant of Vitu-Ghazi", "Instant", effectstest.Pr7bChant, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, c, me.ID, 3)
		if me.Life != life+3 {
			t.Fatalf("life %d, want %d (the 3 prevented, gained as life)", me.Life, life+3)
		}
	})
	t.Run("Scarecrow", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		crow := effectstest.PushCatalogPermanent(g, me.ID, "Scarecrow", "Artifact Creature — Scarecrow", effectstest.Pr7bScarecrow, false)
		flier := effectstest.Pr7bCreature(g, opp.ID, "Flier", 2, "", []string{"U"}, "flying")
		walker := effectstest.Pr7Creature(g, opp.ID, "Walker", 2, "G")
		bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
		effectstest.Pr7Activate(t, g, me.ID, crow, 0, game.ActivateAbilityParams{})
		life := me.Life
		effectstest.Pr7Hit(t, g, flier, me.ID, 2)
		effectstest.Pr7Hit(t, g, walker, me.ID, 1)
		effectstest.Pr6Damage(t, g, flier, bear, 1)
		if me.Life != life-1 || effectstest.Pr6Marked(g, bear) != 1 {
			t.Fatalf("life %d (want %d), bear %d (want 1)", me.Life, life-1, effectstest.Pr6Marked(g, bear))
		}
	})
	t.Run("Eerie Interference", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := effectstest.Pr7Creature(g, opp.ID, "Creature", 2, "R")
		bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
		theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
		art := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Artifact", TypeLine: "Artifact"})
		effectstest.CastCatalogSpell(t, g, "Eerie Interference", "Instant", effectstest.Pr7bEerie, nil)
		effectstest.PassPriorityAroundTable(t, g)
		life := me.Life
		effectstest.Pr7Hit(t, g, c, me.ID, 2)
		effectstest.Pr6Damage(t, g, c, bear, 2)
		effectstest.Pr6Damage(t, g, c, theirs, 1)
		effectstest.Pr7Hit(t, g, art, me.ID, 1)
		if me.Life != life-1 || effectstest.Pr6Marked(g, bear) != 0 || effectstest.Pr6Marked(g, theirs) != 1 {
			t.Fatalf("life %d (want %d), bear %d (want 0), theirs %d (want 1)", me.Life, life-1, effectstest.Pr6Marked(g, bear), effectstest.Pr6Marked(g, theirs))
		}
	})
}
