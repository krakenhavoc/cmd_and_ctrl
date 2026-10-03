package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr7b_cards_test.go — ADR 0108 PR 7b (#1904): the object-source
// and property families of the not-one-use source shield
// (preventFromSource), each card against the damage it must stop and
// the damage it must not.

const (
	pr7bFendOff         = "fa8f3827-8cd9-4896-ab0c-26fecacceb40"
	pr7bKorHaven        = "276cece9-f9f2-46e6-ae76-daddaa2fb9ab"
	pr7bWarning         = "c5ba0f0f-65c5-4ffa-987a-f320b401ec8f"
	pr7bRestrain        = "5aa66cb0-86b9-4e85-9a17-df78416d682d"
	pr7bHorn            = "50a1c14a-003f-424b-bb8e-2e2d51465a90"
	pr7bSafeguard       = "e310c3ab-a729-404d-944f-9b477258495c"
	pr7bLadyEvangela    = "8800d672-424b-4a7b-886f-7eb9d7a56cfe"
	pr7bResistance      = "a95a2d1c-70a9-4eb2-ae30-a07fab26a9c6"
	pr7bSongstitcher    = "eec72bcf-ccbc-43fc-8bdd-7bf4faa1fad7"
	pr7bSubdue          = "81bac4b8-277b-415a-9064-a80a68fd7051"
	pr7bKryShield       = "2c4bd475-b8af-4916-b7a0-68abb8994138"
	pr7bSoulParry       = "8b36335b-119b-4f90-a5d7-b77d8f47f55f"
	pr7bSnare           = "55e3dfcb-0999-401a-ae59-9804e12f4cfe"
	pr7bHallow          = "60b6a31f-043e-4de3-ad37-67eec02114f8"
	pr7bDromoka         = "3bd8ea71-cbcc-4659-b9a8-88cf27ee12d8"
	pr7bAzoriusPloy     = "b9b58c3a-5a76-4b6a-884b-74f54f0ca53c"
	pr7bShieldmageElder = "943bbe5b-76cd-478d-a1fe-d4e557df3469"
	pr7bStonewise       = "ecfa6791-938a-4159-92bf-ff2b0ce69523"
	pr7bIgnoble         = "52a61983-53e2-440f-af03-cebe57b102f2"
	pr7bZealot          = "4107d0e3-608b-4024-b433-40c5d8b63549"
	pr7bFeint           = "1bb8fe05-abb3-40a8-9e80-5d99ed0e4284"
	pr7bFallingTimber   = "4ea19457-97ef-4ac8-b67c-41be1109ca73"
	pr7bSereneSunset    = "e2ef027d-0ef7-4c8b-a594-3918290b02d9"
	pr7bBorosFuryShield = "20eff2ce-f26d-48a2-8a1f-435a7e2968c8"
	pr7bChainOfSilence  = "d7b476ed-fc50-4804-8375-488a9e2ab184"
	pr7bAjanisAid       = "b3779fbe-7701-433b-afad-bebd6096a3c8"
	pr7bFightingChance  = "fece632b-d8d5-4d3d-a4a7-643293b67538"
	pr7bLoafingGiant    = "038dce4c-f754-45ef-98a4-4e16f931a65c"
	pr7bMtendaLion      = "af029853-cfb0-403b-af51-141ba02ae2e4"
	pr7bHeroism         = "74716642-fc8c-4f62-a556-154ed0b4af0e"
	pr7bGuardDogs       = "490af644-2e85-49c7-af63-d5f0a9babff8"
	pr7bRadiantKavu     = "ab46b957-5206-4390-a2f6-aba2e8d1debe"
	pr7bLuminesce       = "7b0f482d-35a3-47f1-8283-0fd47e30cc31"
	pr7bEtherealHaze    = "0cdbd9e7-4941-46ac-99d8-ea227181bdf8"
	pr7bChant           = "e245a736-5f65-4159-8e39-e279e1f8794f"
	pr7bScarecrow       = "58616e15-531f-4680-8294-35ab7e962c96"
	pr7bEerie           = "35ba0e70-da7c-4ed9-b8ec-7dd5c5ce110e"

	pr7bLightningBolt = "4457ed35-7c10-48c8-9776-456485fdf070"
)

// pr7bCreature is pr7Creature with a keyword list and a mana cost.
func pr7bCreature(g *game.Game, owner uuid.UUID, name string, power int, mana string, colors []string, keywords ...string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: 4,
		ManaCost: mana, Colors: colors, Keywords: keywords})
}

// pr7bCastNow puts a card in the active seat's hand and casts it where
// the game stands — in combat, with priority — then resolves the stack.
func pr7bCastNow(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Owner: active.ID, Controller: active.ID})
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func pr7bTargets(ids ...uuid.UUID) []game.TargetRef {
	out := make([]game.TargetRef, len(ids))
	for i, id := range ids {
		out[i] = game.TargetRef{Kind: game.TargetCard, ID: id}
	}
	return out
}

// pr7bAttack declares the attackers against `defender` and locks the
// declaration in, leaving the game in the declare attackers step with
// any attack triggers resolved.
func pr7bAttack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
}

// pr7bBlock walks to the declare blockers step, declares the blocks
// (blocker → attacker) and locks them in, leaving priority with the
// active player in that step.
func pr7bBlock(t *testing.T, g *game.Game, blocks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareBlockers)
	for b, a := range blocks {
		if err := g.DeclareBlocker(b, a); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	finishBlockDeclarations(t, g)
	passPriorityAroundTable(t, g)
}

// pr7bDamageStep walks into the combat damage step, where the damage is
// dealt, and runs its priority boundary.
func pr7bDamageStep(t *testing.T, g *game.Game) {
	t.Helper()
	advanceTo(t, g, game.StepCombatDamage)
	g.RunStateChecksForTest()
}

func pr7bPrompt(t *testing.T, g *game.Game, kind game.PendingChoiceKind) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == kind {
			return c
		}
	}
	t.Fatalf("no %s prompt open", kind)
	return nil
}

func pr7bSourceShields(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModPreventFromSource {
				n++
			}
		}
	}
	return n
}

// Every "prevent all combat damage that would be dealt by target
// [attacking] creature this turn" card: the targeted attacker deals no
// combat damage, and its non-combat damage is still dealt.
func TestADR0108PR7bCombatShieldsAgainstTheTarget(t *testing.T) {
	spells := []struct{ name, oracle string }{
		{"Fend Off", pr7bFendOff},
		{"Warning", pr7bWarning},
		{"Restrain", pr7bRestrain},
		{"Subdue", pr7bSubdue},
		{"Boros Fury-Shield", pr7bBorosFuryShield},
	}
	for _, s := range spells {
		t.Run(s.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			att := pr7Creature(g, me.ID, "Attacker", 3, "R")
			pr7bAttack(t, g, opp.ID, att)
			pr7bCastNow(t, g, s.name, "Instant", s.oracle, game.CastSpellParams{Targets: pr7bTargets(att)})
			life := opp.Life
			pr7bDamageStep(t, g)
			if opp.Life != life {
				t.Fatalf("defender life %d, want %d: the attacker's combat damage is prevented", opp.Life, life)
			}
			pr7Hit(t, g, att, opp.ID, 1)
			if opp.Life != life-1 {
				t.Fatalf("defender life %d, want %d: non-combat damage is not prevented", opp.Life, life-1)
			}
		})
	}
	rows := []struct {
		name, typeLine, oracle string
		params                 func(g *game.Game, me uuid.UUID) game.ActivateAbilityParams
	}{
		{"Kor Haven", "Legendary Land", pr7bKorHaven, nil},
		{"Horn of Deafening", "Artifact", pr7bHorn, nil},
		{"Safeguard", "Enchantment", pr7bSafeguard, nil},
		{"Lady Evangela", "Legendary Creature — Human Cleric", pr7bLadyEvangela, nil},
		{"Resistance Fighter", "Creature — Human Soldier", pr7bResistance, nil},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushCatalogPermanent(g, me.ID, r.name, r.typeLine, r.oracle, false)
			att := pr7Creature(g, me.ID, "Attacker", 3, "R")
			pr7bAttack(t, g, opp.ID, att)
			pr7Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Targets: pr7bTargets(att)})
			life := opp.Life
			pr7bDamageStep(t, g)
			if opp.Life != life {
				t.Fatalf("defender life %d, want %d", opp.Life, life)
			}
		})
	}
}

// Songstitcher targets only an attacker with flying.
func TestADR0108PR7bSongstitcherNeedsAFlyingAttacker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pushCatalogPermanent(g, me.ID, "Songstitcher", "Creature — Human Cleric", pr7bSongstitcher, false)
	flier := pr7bCreature(g, me.ID, "Flier", 3, "", []string{"W"}, "flying")
	walker := pr7Creature(g, me.ID, "Walker", 2, "G")
	pr7bAttack(t, g, opp.ID, flier, walker)
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{Targets: pr7bTargets(walker)}); err == nil {
		t.Fatal("a non-flying attacker was accepted as the target")
	}
	pr7Activate(t, g, me.ID, src, 0, game.ActivateAbilityParams{Targets: pr7bTargets(flier)})
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life-2 {
		t.Fatalf("defender life %d, want %d: only the walker's 2 is dealt", opp.Life, life-2)
	}
}

// Restrain draws; Subdue and Kry Shield add the creature's mana value
// to its toughness; Kry Shield prevents ALL the creature's damage.
func TestADR0108PR7bRidersOnTheTargetedShield(t *testing.T) {
	t.Run("Restrain draws", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		att := pr7Creature(g, me.ID, "Attacker", 3, "R")
		pr7bAttack(t, g, opp.ID, att)
		hand := me.Hand.Size()
		pr7bCastNow(t, g, "Restrain", "Instant", pr7bRestrain, game.CastSpellParams{Targets: pr7bTargets(att)})
		if me.Hand.Size() != hand+1 {
			t.Fatalf("hand %d, want %d (Restrain put in and cast, one card drawn)", me.Hand.Size(), hand+1)
		}
	})
	t.Run("Subdue toughness", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		c := pr7bCreature(g, me.ID, "Three Drop", 3, "{2}{G}", []string{"G"})
		castCatalogSpell(t, g, "Subdue", "Instant", pr7bSubdue, pr7bTargets(c))
		passPriorityAroundTable(t, g)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		if got := findBattlefieldCardForTest(g, c).CurrentToughness(); got != 7 {
			t.Fatalf("toughness %d, want 7 (4 + mana value 3)", got)
		}
	})
	t.Run("Kry Shield", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		shield := pushCatalogPermanent(g, me.ID, "Kry Shield", "Artifact", pr7bKryShield, false)
		c := pr7bCreature(g, me.ID, "Two Drop", 2, "{1}{R}", []string{"R"})
		pr7Activate(t, g, me.ID, shield, 0, game.ActivateAbilityParams{Targets: pr7bTargets(c)})
		life := opp.Life
		pr7Hit(t, g, c, opp.ID, 2)
		g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
		if opp.Life != life || findBattlefieldCardForTest(g, c).CurrentToughness() != 6 {
			t.Fatalf("life %d (want %d), toughness %d (want 6)", opp.Life, life, findBattlefieldCardForTest(g, c).CurrentToughness())
		}
	})
}

// Soul Parry shields one or two sources, all their damage; a third
// source is untouched.
func TestADR0108PR7bSoulParryShieldsBothTargets(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, opp.ID, "A", 2, "R")
	b := pr7Creature(g, opp.ID, "B", 2, "R")
	c := pr7Creature(g, opp.ID, "C", 2, "R")
	castCatalogSpell(t, g, "Soul Parry", "Instant", pr7bSoulParry, pr7bTargets(a, b))
	passPriorityAroundTable(t, g)
	life := me.Life
	pr7Hit(t, g, a, me.ID, 2)
	pr7Hit(t, g, b, me.ID, 2)
	pr7Hit(t, g, c, me.ID, 2)
	if me.Life != life-2 {
		t.Fatalf("life %d, want %d", me.Life, life-2)
	}
}

// Inquisitor's Snare destroys a red target and still prevents its
// damage as it last existed; a green one lives, shielded.
func TestADR0108PR7bInquisitorsSnare(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := pr7Creature(g, me.ID, "Red", 3, "R")
	green := pr7Creature(g, me.ID, "Green", 3, "G")
	pr7bAttack(t, g, opp.ID, red, green)
	pr7bCastNow(t, g, "Inquisitor's Snare", "Instant", pr7bSnare, game.CastSpellParams{Targets: pr7bTargets(red)})
	if findBattlefieldCardForTest(g, red) != nil {
		t.Fatal("the red creature was not destroyed")
	}
	pr7bCastNow(t, g, "Inquisitor's Snare", "Instant", pr7bSnare, game.CastSpellParams{Targets: pr7bTargets(green)})
	if findBattlefieldCardForTest(g, green) == nil {
		t.Fatal("the green creature was destroyed")
	}
	life := opp.Life
	pr7bDamageStep(t, g)
	pr7Hit(t, g, green, opp.ID, 2)
	if opp.Life != life {
		t.Fatalf("defender life %d, want %d", opp.Life, life)
	}
}

// Hallow prevents a spell's damage and gains its caster that much life.
func TestADR0108PR7bHallowStopsTheSpellAndGainsTheLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceTo(t, g, game.StepPrecombatMain)
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: pr7bLightningBolt,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, HoldPriority: true}); err != nil {
		t.Fatal(err)
	}
	life, mine := opp.Life, me.Life
	pr7bCastNow(t, g, "Hallow", "Instant", pr7bHallow, game.CastSpellParams{Targets: pr7bTargets(bolt)})
	if opp.Life != life || me.Life != mine+3 {
		t.Fatalf("opponent %d (want %d), me %d (want %d)", opp.Life, life, me.Life, mine+3)
	}
}

// Dromoka's Command: the counter lands before the fight; the
// sacrifice is the target player's choice.
func TestADR0108PR7bDromokasCommand(t *testing.T) {
	t.Run("counter then fight", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		mine := pr7Creature(g, me.ID, "Mine", 2, "G")
		theirs := pr7Creature(g, opp.ID, "Theirs", 1, "B")
		advanceTo(t, g, game.StepPrecombatMain)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dromoka's Command", TypeLine: "Instant", OracleID: pr7bDromoka,
			Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{2, 3}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: mine, Mode: 0},
			{Kind: game.TargetCard, ID: mine, Mode: 1, Slot: 0},
			{Kind: game.TargetCard, ID: theirs, Mode: 1, Slot: 1},
		}})
		if err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if got := pr6Marked(g, theirs); got != 3 {
			t.Fatalf("their creature took %d, want 3 (2 power plus the counter)", got)
		}
	})
	t.Run("spell shield and sacrifice", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ench := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Some Enchantment", TypeLine: "Enchantment"})
		advanceTo(t, g, game.StepPrecombatMain)
		bolt := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", OracleID: pr7bLightningBolt,
			Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, HoldPriority: true}); err != nil {
			t.Fatal(err)
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Dromoka's Command", TypeLine: "Instant", OracleID: pr7bDromoka,
			Owner: me.ID, Controller: me.ID})
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: bolt, Mode: 0},
			{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1},
		}})
		if err != nil {
			t.Fatal(err)
		}
		life := opp.Life
		passPriorityAroundTable(t, g)
		c := pr7bPrompt(t, g, game.PendingChoiceOwnPermanents)
		if c.Chooser != opp.ID {
			t.Fatalf("the sacrifice is asked of %v, want the target player", c.Chooser)
		}
		if err := g.ResolveOwnPermanents(c.ID, opp.ID, []uuid.UUID{ench}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if findBattlefieldCardForTest(g, ench) != nil || opp.Life != life {
			t.Fatalf("enchantment still there %v, life %d (want %d)", findBattlefieldCardForTest(g, ench) != nil, opp.Life, life)
		}
	})
}

// Azorius Ploy: one creature's combat damage, and combat damage to
// another, are prevented.
func TestADR0108PR7bAzoriusPloy(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	blk := pr7Creature(g, opp.ID, "Blocker", 2, "G")
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: att})
	pr7bCastNow(t, g, "Azorius Ploy", "Instant", pr7bAzoriusPloy, game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: blk, Slot: 0},
		{Kind: game.TargetCard, ID: blk, Slot: 1},
	}})
	pr7bDamageStep(t, g)
	if pr6Marked(g, blk) != 0 || pr6Marked(g, att) != 0 {
		t.Fatalf("blocker %d, attacker %d: both want 0", pr6Marked(g, blk), pr6Marked(g, att))
	}
}

// Shieldmage Elder taps two Clerics (itself among them) to shield
// against a creature.
func TestADR0108PR7bShieldmageElderTapsTwoClerics(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	elder := apaPush(g, me.ID, me.ID, game.Card{Name: "Shieldmage Elder", TypeLine: "Creature — Human Cleric Wizard",
		OracleID: pr7bShieldmageElder, Power: 2, Toughness: 3})
	cleric := apaPush(g, me.ID, me.ID, game.Card{Name: "Cleric", TypeLine: "Creature — Human Cleric", Power: 1, Toughness: 1})
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	pr7Activate(t, g, me.ID, elder, 0, game.ActivateAbilityParams{Targets: pr7bTargets(src), TapIDs: []uuid.UUID{elder, cleric}})
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life || !findBattlefieldCardForTest(g, cleric).Tapped {
		t.Fatalf("life %d (want %d), cleric tapped %v", me.Life, life, findBattlefieldCardForTest(g, cleric).Tapped)
	}
}

// Stonewise Fortifier protects only itself from only the target.
func TestADR0108PR7bStonewiseFortifier(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fort := apaPush(g, me.ID, me.ID, game.Card{Name: "Stonewise Fortifier", TypeLine: "Creature — Human Wizard",
		OracleID: pr7bStonewise, Power: 2, Toughness: 4})
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	other := pr7Creature(g, opp.ID, "Other", 3, "R")
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	pr7Activate(t, g, me.ID, fort, 0, game.ActivateAbilityParams{Targets: pr7bTargets(src)})
	pr6Damage(t, g, src, fort, 2)
	pr6Damage(t, g, src, bear, 1)
	pr6Damage(t, g, other, fort, 1)
	if pr6Marked(g, fort) != 1 || pr6Marked(g, bear) != 1 {
		t.Fatalf("fortifier %d (want 1), bear %d (want 1)", pr6Marked(g, fort), pr6Marked(g, bear))
	}
}

// Ignoble Soldier, blocked, deals no combat damage and still takes it.
func TestADR0108PR7bIgnobleSoldierBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	soldier := apaPush(g, me.ID, me.ID, game.Card{Name: "Ignoble Soldier", TypeLine: "Creature — Human Soldier",
		OracleID: pr7bIgnoble, Power: 3, Toughness: 4})
	blk := pr7Creature(g, opp.ID, "Blocker", 2, "G")
	pr7bAttack(t, g, opp.ID, soldier)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{blk: soldier})
	pr7bDamageStep(t, g)
	if pr6Marked(g, blk) != 0 || pr6Marked(g, soldier) != 2 {
		t.Fatalf("blocker %d (want 0), soldier %d (want 2)", pr6Marked(g, blk), pr6Marked(g, soldier))
	}
}

// Zealot il-Vec, unblocked, may ping a creature; if it does, its own
// combat damage is prevented.
func TestADR0108PR7bZealotIlVec(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		zealot := apaPush(g, me.ID, me.ID, game.Card{Name: "Zealot il-Vec", TypeLine: "Creature — Human Rebel",
			OracleID: pr7bZealot, Power: 1, Toughness: 1})
		victim := pr7Creature(g, opp.ID, "Victim", 2, "G")
		pr7bAttack(t, g, opp.ID, zealot)
		advanceTo(t, g, game.StepDeclareBlockers)
		finishBlockDeclarations(t, g)
		for i := 0; i < 8; i++ {
			if c := firstPendingTarget(g); c != nil {
				break
			}
			if err := g.PassPriority(); err != nil {
				break
			}
		}
		if c := firstPendingTarget(g); c != nil {
			if err := g.ResolvePickTarget(c.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}); err != nil {
				t.Fatal(err)
			}
		}
		passPriorityAroundTable(t, g)
		c := pr7bPrompt(t, g, game.PendingChoiceConfirm)
		if err := g.ResolveConfirm(c.ID, me.ID, yes); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		life := opp.Life
		pr7bDamageStep(t, g)
		switch {
		case yes && (pr6Marked(g, victim) != 1 || opp.Life != life):
			t.Fatalf("yes: victim %d (want 1), life %d (want %d)", pr6Marked(g, victim), opp.Life, life)
		case !yes && (pr6Marked(g, victim) != 0 || opp.Life != life-1):
			t.Fatalf("no: victim %d (want 0), life %d (want %d)", pr6Marked(g, victim), opp.Life, life-1)
		}
	}
}

// firstPendingTarget is a trigger's open target pick, if any.
func firstPendingTarget(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}

// Feint taps the blockers and prevents the attacker's and theirs.
func TestADR0108PR7bFeint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	b1 := pr7Creature(g, opp.ID, "B1", 2, "G")
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: att})
	pr7bCastNow(t, g, "Feint", "Instant", pr7bFeint, game.CastSpellParams{Targets: pr7bTargets(att)})
	if !findBattlefieldCardForTest(g, b1).Tapped {
		t.Fatal("the blocker is not tapped")
	}
	pr7bDamageStep(t, g)
	if pr6Marked(g, att)+pr6Marked(g, b1) != 0 {
		t.Fatalf("damage marked: attacker %d, blocker %d — want none", pr6Marked(g, att), pr6Marked(g, b1))
	}
}

// Falling Timber kicked shields two creatures; unkicked takes one.
func TestADR0108PR7bFallingTimberKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 2, "R")
	b := pr7Creature(g, me.ID, "B", 3, "R")
	land := apaPush(g, me.ID, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	pr7bAttack(t, g, opp.ID, a, b)
	pr7bCastNow(t, g, "Falling Timber", "Instant", pr7bFallingTimber, game.CastSpellParams{
		OptionalCosts: []int{0},
		SacrificeIDs:  []uuid.UUID{land},
		Targets:       []game.TargetRef{{Kind: game.TargetCard, ID: a, Slot: 0}, {Kind: game.TargetCard, ID: b, Slot: 1}},
	})
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life || findBattlefieldCardForTest(g, land) != nil {
		t.Fatalf("life %d (want %d), land still there %v", opp.Life, life, findBattlefieldCardForTest(g, land) != nil)
	}
}

// Serene Sunset takes X targets.
func TestADR0108PR7bSereneSunset(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pr7Creature(g, me.ID, "A", 2, "R")
	b := pr7Creature(g, me.ID, "B", 3, "R")
	c := pr7Creature(g, me.ID, "C", 1, "R")
	pr7bAttack(t, g, opp.ID, a, b, c)
	pr7bCastNow(t, g, "Serene Sunset", "Instant", pr7bSereneSunset, game.CastSpellParams{XValue: 2, Targets: pr7bTargets(a, b)})
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life-1 {
		t.Fatalf("life %d, want %d (only C's 1)", opp.Life, life-1)
	}
}

// Chain of Silence shields the creature and offers its controller the
// land-for-a-copy chain.
func TestADR0108PR7bChainOfSilence(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	src := pr7Creature(g, opp.ID, "Src", 3, "R")
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	castCatalogSpell(t, g, "Chain of Silence", "Instant", pr7bChainOfSilence, pr7bTargets(src))
	passPriorityAroundTable(t, g)
	if c := pr7bPrompt(t, g, game.PendingChoiceConfirm); c.Chooser != opp.ID {
		t.Fatalf("the chain is offered to %v, want the creature's controller", c.Chooser)
	}
	life := me.Life
	pr7Hit(t, g, src, me.ID, 3)
	if me.Life != life {
		t.Fatalf("life %d, want %d", me.Life, life)
	}
}

// Ajani's Aid: sacrificed, a creature of your choice deals no combat
// damage; only creatures are offered.
func TestADR0108PR7bAjanisAid(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	aid := pushCatalogPermanent(g, me.ID, "Ajani's Aid", "Enchantment", pr7bAjanisAid, false)
	art := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Pinger", TypeLine: "Artifact", Colors: nil})
	att := pr7Creature(g, me.ID, "Attacker", 3, "R")
	pr7bAttack(t, g, opp.ID, att)
	pr7Activate(t, g, me.ID, aid, 0, game.ActivateAbilityParams{})
	c := pr7SourcePrompt(t, g)
	if pr7Offers(c, art) || !pr7Offers(c, att) {
		t.Fatalf("offered %v: want creatures only", c.ChooseCards)
	}
	pr7Choose(t, g, me.ID, att)
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life {
		t.Fatalf("life %d, want %d", opp.Life, life)
	}
}

// Fighting Chance shields exactly the blockers whose flips were won.
func TestADR0108PR7bFightingChance(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	att := pr7Creature(g, me.ID, "Attacker", 9, "R")
	b1 := pr7Creature(g, opp.ID, "B1", 2, "G")
	b2 := pr7Creature(g, opp.ID, "B2", 2, "G")
	b3 := pr7Creature(g, opp.ID, "B3", 2, "G")
	pr7bAttack(t, g, opp.ID, att)
	pr7bBlock(t, g, map[uuid.UUID]uuid.UUID{b1: att, b2: att, b3: att})
	pr7bCastNow(t, g, "Fighting Chance", "Instant", pr7bFightingChance, game.CastSpellParams{})
	c := pr7bPrompt(t, g, game.PendingChoiceCoinCall)
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
	if got := pr7bSourceShields(g); got != won {
		t.Fatalf("%d shields for %d won flips", got, won)
	}
}

// Loafing Giant milling a land prevents its own combat damage.
func TestADR0108PR7bLoafingGiantMillsALand(t *testing.T) {
	for _, land := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		giant := apaPush(g, me.ID, me.ID, game.Card{Name: "Loafing Giant", TypeLine: "Creature — Giant",
			OracleID: pr7bLoafingGiant, Power: 4, Toughness: 6})
		top := game.Card{InstanceID: uuid.New(), Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if land {
			top.Name, top.TypeLine = "Forest", "Basic Land — Forest"
		}
		me.Library.PushTop(top)
		pr7bAttack(t, g, opp.ID, giant)
		life := opp.Life
		pr7bDamageStep(t, g)
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
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		lion := apaPush(g, me.ID, me.ID, game.Card{Name: "Mtenda Lion", TypeLine: "Creature — Cat",
			OracleID: pr7bMtendaLion, Power: 2, Toughness: 1})
		pr7bAttack(t, g, opp.ID, lion)
		opp.ManaPool.AddMana(game.ManaToken{Color: "U"})
		answerPayUnless(t, g, opp.ID, pay)
		passPriorityAroundTable(t, g)
		life := opp.Life
		pr7bDamageStep(t, g)
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
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	heroism := pushCatalogPermanent(g, opp.ID, "Heroism", "Enchantment", pr7bHeroism, false)
	white := pr7Creature(g, opp.ID, "White", 1, "W")
	red := pr7Creature(g, me.ID, "Red", 3, "R")
	green := pr7Creature(g, me.ID, "Green", 2, "G")
	pr7bAttack(t, g, opp.ID, red, green)
	pr7Activate(t, g, opp.ID, heroism, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{white}})
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	life := opp.Life
	pr7bDamageStep(t, g)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d (only the green 2)", opp.Life, life-2)
	}
}

// Guard Dogs: a shared colour makes the shield, read as it resolves.
func TestADR0108PR7bGuardDogs(t *testing.T) {
	for _, shares := range []bool{true, false} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dogs := pushCatalogPermanent(g, me.ID, "Guard Dogs", "Creature — Dog", pr7bGuardDogs, false)
		color := "R"
		if !shares {
			color = "U"
		}
		mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Enchantment", Colors: []string{color}})
		src := pr7Creature(g, opp.ID, "Src", 3, "R")
		pr7Activate(t, g, me.ID, dogs, 0, game.ActivateAbilityParams{Targets: pr7bTargets(src)})
		c := pr7bPrompt(t, g, game.PendingChoiceOwnPermanents)
		if err := g.ResolveOwnPermanents(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		want := 0
		if shares {
			want = 1
		}
		if got := pr7bSourceShields(g); got != want {
			t.Fatalf("shares %v: %d shields, want %d", shares, got, want)
		}
	}
}

// Radiant Kavu: blue and black creatures' combat damage only.
func TestADR0108PR7bRadiantKavu(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kavu := pushCatalogPermanent(g, me.ID, "Radiant Kavu", "Creature — Kavu", pr7bRadiantKavu, false)
	blue := pr7Creature(g, me.ID, "Blue", 2, "U")
	black := pr7Creature(g, me.ID, "Black", 3, "B")
	green := pr7Creature(g, me.ID, "Green", 1, "G")
	pr7bAttack(t, g, opp.ID, blue, black, green)
	pr7Activate(t, g, me.ID, kavu, 0, game.ActivateAbilityParams{})
	life := opp.Life
	pr7bDamageStep(t, g)
	pr7Hit(t, g, blue, opp.ID, 1)
	if opp.Life != life-2 {
		t.Fatalf("life %d, want %d (green's 1 and blue's non-combat 1)", opp.Life, life-2)
	}
}

// Luminesce, Ethereal Haze, Chant of Vitu-Ghazi, Scarecrow and Eerie
// Interference: a property shield, checked as the damage is dealt.
func TestADR0108PR7bPropertyShields(t *testing.T) {
	t.Run("Luminesce", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		black := pr7Creature(g, opp.ID, "Black", 2, "B")
		green := pr7Creature(g, opp.ID, "Green", 2, "G")
		castCatalogSpell(t, g, "Luminesce", "Instant", pr7bLuminesce, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, black, me.ID, 2)
		pr7Hit(t, g, green, me.ID, 2)
		if me.Life != life-2 {
			t.Fatalf("life %d, want %d", me.Life, life-2)
		}
	})
	t.Run("Ethereal Haze", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := pr7Creature(g, opp.ID, "Creature", 2, "R")
		art := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Artifact", TypeLine: "Artifact"})
		castCatalogSpell(t, g, "Ethereal Haze", "Instant — Arcane", pr7bEtherealHaze, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, c, me.ID, 2)
		pr7Hit(t, g, art, me.ID, 1)
		if me.Life != life-1 {
			t.Fatalf("life %d, want %d", me.Life, life-1)
		}
	})
	t.Run("Chant of Vitu-Ghazi", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := pr7Creature(g, opp.ID, "Creature", 3, "R")
		castCatalogSpell(t, g, "Chant of Vitu-Ghazi", "Instant", pr7bChant, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, c, me.ID, 3)
		if me.Life != life+3 {
			t.Fatalf("life %d, want %d (the 3 prevented, gained as life)", me.Life, life+3)
		}
	})
	t.Run("Scarecrow", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		crow := pushCatalogPermanent(g, me.ID, "Scarecrow", "Artifact Creature — Scarecrow", pr7bScarecrow, false)
		flier := pr7bCreature(g, opp.ID, "Flier", 2, "", []string{"U"}, "flying")
		walker := pr7Creature(g, opp.ID, "Walker", 2, "G")
		bear := pr7Creature(g, me.ID, "Bear", 2, "G")
		pr7Activate(t, g, me.ID, crow, 0, game.ActivateAbilityParams{})
		life := me.Life
		pr7Hit(t, g, flier, me.ID, 2)
		pr7Hit(t, g, walker, me.ID, 1)
		pr6Damage(t, g, flier, bear, 1)
		if me.Life != life-1 || pr6Marked(g, bear) != 1 {
			t.Fatalf("life %d (want %d), bear %d (want 1)", me.Life, life-1, pr6Marked(g, bear))
		}
	})
	t.Run("Eerie Interference", func(t *testing.T) {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		c := pr7Creature(g, opp.ID, "Creature", 2, "R")
		bear := pr7Creature(g, me.ID, "Bear", 2, "G")
		theirs := pr7Creature(g, opp.ID, "Theirs", 2, "G")
		art := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Artifact", TypeLine: "Artifact"})
		castCatalogSpell(t, g, "Eerie Interference", "Instant", pr7bEerie, nil)
		passPriorityAroundTable(t, g)
		life := me.Life
		pr7Hit(t, g, c, me.ID, 2)
		pr6Damage(t, g, c, bear, 2)
		pr6Damage(t, g, c, theirs, 1)
		pr7Hit(t, g, art, me.ID, 1)
		if me.Life != life-1 || pr6Marked(g, bear) != 0 || pr6Marked(g, theirs) != 1 {
			t.Fatalf("life %d (want %d), bear %d (want 0), theirs %d (want 1)", me.Life, life-1, pr6Marked(g, bear), pr6Marked(g, theirs))
		}
	})
}
