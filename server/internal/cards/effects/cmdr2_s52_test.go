package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S52YarokOracle = "95b53836-18aa-451a-992d-2a111deeeab2"
	cmdr2S52FelixOracle = "95175d35-c815-47bd-8e69-427972b567b6"
)

func TestCmdr2S52YarokDoublesAnyPermanentEntering(t *testing.T) {
	for _, withYarok := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withYarok {
			pushCatalogPermanent(g, me.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
		}
		pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
		before := opp.Library.Size()
		// A plain Forest is neither an artifact nor a creature: Panharmonicon
		// would not double it, Yarok does.
		castCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
		passPriorityAroundTable(t, g)
		got := before - opp.Library.Size()
		want := 3
		if withYarok {
			want = 6
		}
		if got != want {
			t.Errorf("withYarok=%v: milled %d, want %d", withYarok, got, want)
		}
	}
}

func TestCmdr2S52YarokOnlyDoublesItsControllersAbilities(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
	pushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
	before := opp.Library.Size()
	castCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("opponent's Yarok doubled my trigger: milled %d, want 3", got)
	}
}

func TestCmdr2S52FelixDoublesCombatDamageToAPlayerTriggers(t *testing.T) {
	for _, withFelix := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		opp := g.Seats[1]
		bidentID := pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
		a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		if withFelix {
			pushCatalogPermanent(g, me.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
		}
		_ = bidentID
		attackWith(t, g, opp.ID, a)
		want := 1
		if withFelix {
			want = 2
		}
		if got := owedMays(g, me.ID); got != want {
			t.Errorf("withFelix=%v: Bident prompts %d, want %d", withFelix, got, want)
		}
	}
}

func TestCmdr2S52FelixIgnoresOpponentsDoublerAndOtherCauses(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushPermanentForTest(g, me.ID, "Bident of Thassa", bidentOracle, "Legendary Enchantment Artifact")
	pushCatalogPermanent(g, opp.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	attackWith(t, g, opp.ID, a)
	if got := owedMays(g, me.ID); got != 1 {
		t.Errorf("opponent's Felix doubled my trigger: %d prompts, want 1", got)
	}
	spec, ok := Lookup(cmdr2S52FelixOracle)
	if !ok || len(spec.TriggerDoublers) != 1 {
		t.Fatalf("Felix registration: ok=%v", ok)
	}
	// Entering is not a combat-damage cause.
	q := game.TriggerDoublingQuery{Event: game.Event{Kind: game.EventETB}, Ability: &game.TriggeredAbility{}, SourceIsPermanent: true, HasSubject: true}
	if spec.TriggerDoublers[0].Applies(g, q) {
		t.Error("Felix doubled an entering cause")
	}
}

const cmdr2S52JunkWinderOracle = "5b677e88-bcc9-4a55-9c89-8bbd0e2344dc"

func TestCmdr2S52JunkWinderHasAffinityForTokens(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	if got := selfPricedMV(t, g, me, cmdr2S52JunkWinderOracle, "Creature — Serpent", "{5}{U}{U}"); got != 7 {
		t.Errorf("no tokens: costs %d, want 7", got)
	}
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	pushToken(g, me.ID, TokenCard("1/1 white Soldier"))
	pushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := selfPricedMV(t, g, me, cmdr2S52JunkWinderOracle, "Creature — Serpent", "{5}{U}{U}"); got != 5 {
		t.Errorf("two tokens and a nontoken: costs %d, want 5", got)
	}
}

func TestCmdr2S52JunkWinderTapsAndFreezesOnTokenEntry(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Junk Winder", "Creature — Serpent", cmdr2S52JunkWinderOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// A token entering under an opponent's control does not trigger it.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 white Soldier"), 1) })
	if p := latestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's token triggered Junk Winder")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, TokenCard("1/1 white Soldier"), 1) })
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt for Junk Winder")
	}
	if hasID(p.PickTargetCards, mine) {
		t.Error("the picker offered my own permanent; the clause says an opponent controls")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	c := layeredCard(t, g, theirs)
	if !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Errorf("target: tapped=%v skips=%d, want tapped with one untap skip", c.Tapped, len(c.NextUntapSkips))
	}
}

const cmdr2S52AnyaOracle = "0665daf5-e12a-4b59-81fc-63fec2807571"

func TestCmdr2S52AnyaScalesWithOpponentsBelowHalfStarting(t *testing.T) {
	g := newCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	anya := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anya, Merciless Angel", TypeLine: "Legendary Creature — Angel",
		OracleID: cmdr2S52AnyaOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	check := func(label string, wantPT int, wantIndestructible bool) {
		t.Helper()
		if p, tg := effectivePower(t, g, anya), effectiveToughness(t, g, anya); p != wantPT || tg != wantPT {
			t.Errorf("%s: %d/%d, want %d/%d", label, p, tg, wantPT, wantPT)
		}
		if got := hasEffectiveKeyword(t, g, anya, "indestructible"); got != wantIndestructible {
			t.Errorf("%s: indestructible = %v, want %v", label, got, wantIndestructible)
		}
	}
	setLife := func(p *game.Player, life int) {
		g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, p.ID, life-p.Life) })
	}
	check("all at 40", 4, false)
	setLife(o1, 20) // exactly half is not less than half
	check("opponent at exactly half", 4, false)
	setLife(o1, 19)
	check("one opponent at 19", 7, true)
	setLife(o2, 1)
	setLife(me, 1)
	check("two opponents low (my own life is irrelevant)", 10, true)
	setLife(o1, 40)
	setLife(o2, 40)
	check("both recovered", 4, false)
}

const cmdr2S52KorvoldOracle = "99e89098-bc55-4297-9407-ab9256b609da"

func TestCmdr2S52KorvoldCostsLessPerCardTypeSacrificedThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	price := func() int {
		return selfPricedMV(t, g, me, cmdr2S52KorvoldOracle, "Legendary Creature — Dragon Noble", "{5}{B}{R}{G}")
	}
	if got := price(); got != 8 {
		t.Fatalf("nothing sacrificed: %d, want 8", got)
	}
	golem := pushPermanent(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1})
	other := pushPermanent(g, me.ID, game.Card{Name: "Clue", TypeLine: "Artifact", Power: 0, Toughness: 0})
	land := pushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(golem) })
	if got := price(); got != 6 {
		t.Errorf("an artifact creature sacrificed (two types): %d, want 6", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(other) })
	if got := price(); got != 6 {
		t.Errorf("a second artifact adds no new type: %d, want 6", got)
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(land) })
	if got := price(); got != 5 {
		t.Errorf("a land adds a third type: %d, want 5", got)
	}
}

func TestCmdr2S52KorvoldGrowsAndDrawsPerPermanentTypeInGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	korvold := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Korvold, Gleeful Glutton", TypeLine: "Legendary Creature — Dragon Noble",
		OracleID: cmdr2S52KorvoldOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying", "trample", "haste"},
	})
	// An artifact creature is two types; the land a third; an instant none.
	std51Graveyard(me, "Artifact Creature — Golem", "Basic Land — Forest", "Instant")
	hand := me.Hand.Size()
	attackWith(t, g, opp.ID, korvold)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := effectivePowerWithCounters(t, g, korvold); got != 7 {
		t.Errorf("power %d, want 7 (three counters)", got)
	}
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

const cmdr2S52DereviOracle = "afa49a09-146f-4439-850e-dd1938c93cef"

func cmdr2S52DereviInCommandZone(p *game.Player, casts int) uuid.UUID {
	id := uuid.New()
	p.Command.PushTop(game.Card{
		InstanceID: id, Name: "Derevi, Empyrial Tactician", TypeLine: "Legendary Creature — Bird Wizard",
		OracleID: cmdr2S52DereviOracle, ManaCost: "{G}{W}{U}", Power: 2, Toughness: 3,
		Owner: p.ID, Controller: p.ID, IsCommander: true,
	})
	if casts > 0 {
		p.CommanderCasts[id] = casts
	}
	return id
}

func TestCmdr2S52DereviEntersFromTheCommandZoneUntaxedAndTapsOrUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	derevi := cmdr2S52DereviInCommandZone(me, 2)
	victim := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	std51Pay(g, me, "{1}{G}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, derevi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the command zone: %v", err)
	}
	if !me.Command.Contains(derevi) {
		t.Error("Derevi left the command zone at activation; she enters on resolution")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(derevi) {
		t.Fatal("Derevi is not on the battlefield")
	}
	if got := me.CommanderCasts[derevi]; got != 2 {
		t.Errorf("CommanderCasts = %d, want 2 unchanged (put, not cast)", got)
	}
	if c, _ := g.LookupCardForEffect(derevi); !c.IsCommander {
		t.Error("Derevi lost the commander designation")
	}
	// Her enters trigger: tap their Bear.
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Derevi's enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	passPriorityAroundTable(t, g)
	if !layeredCard(t, g, victim).Tapped {
		t.Error("the Bear should be tapped")
	}
}

func TestCmdr2S52DereviAbilityWorksOnlyFromTheCommandZone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	derevi := pushCatalogPermanent(g, me.ID, "Derevi, Empyrial Tactician", "Legendary Creature — Bird Wizard", cmdr2S52DereviOracle, false)
	std51Pay(g, me, "{1}{G}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, derevi, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the put-from-command-zone ability activated from the battlefield")
	}
}

func TestCmdr2S52DereviTriggersOnCombatDamageByAnyCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Derevi, Empyrial Tactician", "Legendary Creature — Bird Wizard", cmdr2S52DereviOracle, false)
	bear := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	mine := pushVanillaCreature(g, me.ID, "Tapped Thing", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, mine).Tapped = true })
	attackWith(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt after my creature connected")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 1) // untap it
	passPriorityAroundTable(t, g)
	if layeredCard(t, g, mine).Tapped {
		t.Error("the permanent should be untapped")
	}
}
