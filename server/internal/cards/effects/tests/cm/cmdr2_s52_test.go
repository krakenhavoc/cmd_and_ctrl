package testscm

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	cmdr2S52YarokOracle = "95b53836-18aa-451a-992d-2a111deeeab2"
	cmdr2S52FelixOracle = "95175d35-c815-47bd-8e69-427972b567b6"
)

func TestCmdr2S52YarokDoublesAnyPermanentEntering(t *testing.T) {
	for _, withYarok := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		if withYarok {
			effectstest.PushCatalogPermanent(g, me.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
		}
		effectstest.PushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
		before := opp.Library.Size()
		// A plain Forest is neither an artifact nor a creature: Panharmonicon
		// would not double it, Yarok does.
		effectstest.CastCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
		effectstest.PassPriorityAroundTable(t, g)
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, opp.ID, "Yarok, the Desecrated", "Legendary Creature — Elemental Horror", cmdr2S52YarokOracle, false)
	effectstest.PushCatalogPermanent(g, me.ID, "Ruin Crab", "Creature — Crab", "8afc00d4-a1c6-4329-af2c-a7f58a0c33e7", false)
	before := opp.Library.Size()
	effectstest.CastCatalogSpell(t, g, "Forest", "Basic Land — Forest", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 3 {
		t.Errorf("opponent's Yarok doubled my trigger: milled %d, want 3", got)
	}
}

func TestCmdr2S52FelixDoublesCombatDamageToAPlayerTriggers(t *testing.T) {
	for _, withFelix := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		opp := g.Seats[1]
		bidentID := effectstest.PushPermanentForTest(g, me.ID, "Bident of Thassa", effectstest.BidentOracle, "Legendary Enchantment Artifact")
		a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
		if withFelix {
			effectstest.PushCatalogPermanent(g, me.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
		}
		_ = bidentID
		effectstest.AttackWith(t, g, opp.ID, a)
		want := 1
		if withFelix {
			want = 2
		}
		if got := effectstest.OwedMays(g, me.ID); got != want {
			t.Errorf("withFelix=%v: Bident prompts %d, want %d", withFelix, got, want)
		}
	}
}

func TestCmdr2S52FelixIgnoresOpponentsDoublerAndOtherCauses(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, me.ID, "Bident of Thassa", effectstest.BidentOracle, "Legendary Enchantment Artifact")
	effectstest.PushCatalogPermanent(g, opp.ID, "Felix Five-Boots", "Legendary Creature — Ooze Rogue", cmdr2S52FelixOracle, false)
	a := effectstest.PushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, a)
	if got := effectstest.OwedMays(g, me.ID); got != 1 {
		t.Errorf("opponent's Felix doubled my trigger: %d prompts, want 1", got)
	}
	spec, ok := effects.Lookup(cmdr2S52FelixOracle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	if got := effectstest.SelfPricedMV(t, g, me, cmdr2S52JunkWinderOracle, "Creature — Serpent", "{5}{U}{U}"); got != 7 {
		t.Errorf("no tokens: costs %d, want 7", got)
	}
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Soldier"))
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if got := effectstest.SelfPricedMV(t, g, me, cmdr2S52JunkWinderOracle, "Creature — Serpent", "{5}{U}{U}"); got != 5 {
		t.Errorf("two tokens and a nontoken: costs %d, want 5", got)
	}
}

func TestCmdr2S52JunkWinderTapsAndFreezesOnTokenEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Junk Winder", "Creature — Serpent", cmdr2S52JunkWinderOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)

	// A token entering under an opponent's control does not trigger it.
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		t.Fatal("an opponent's token triggered Junk Winder")
	}

	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, effects.TokenCard("1/1 white Soldier"), 1) })
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no pick_target prompt for Junk Winder")
	}
	if effectstest.HasID(p.PickTargetCards, mine) {
		t.Error("the picker offered my own permanent; the clause says an opponent controls")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LayeredCard(t, g, theirs)
	if !c.Tapped || len(c.NextUntapSkips) != 1 {
		t.Errorf("target: tapped=%v skips=%d, want tapped with one untap skip", c.Tapped, len(c.NextUntapSkips))
	}
}

const cmdr2S52AnyaOracle = "0665daf5-e12a-4b59-81fc-63fec2807571"

func TestCmdr2S52AnyaScalesWithOpponentsBelowHalfStarting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, o1, o2 := g.Seats[0], g.Seats[1], g.Seats[2]
	anya := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Anya, Merciless Angel", TypeLine: "Legendary Creature — Angel",
		OracleID: cmdr2S52AnyaOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
	})
	check := func(label string, wantPT int, wantIndestructible bool) {
		t.Helper()
		if p, tg := effectstest.EffectivePower(t, g, anya), effectstest.EffectiveToughness(t, g, anya); p != wantPT || tg != wantPT {
			t.Errorf("%s: %d/%d, want %d/%d", label, p, tg, wantPT, wantPT)
		}
		if got := effectstest.HasEffectiveKeyword(t, g, anya, "indestructible"); got != wantIndestructible {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	price := func() int {
		return effectstest.SelfPricedMV(t, g, me, cmdr2S52KorvoldOracle, "Legendary Creature — Dragon Noble", "{5}{B}{R}{G}")
	}
	if got := price(); got != 8 {
		t.Fatalf("nothing sacrificed: %d, want 8", got)
	}
	golem := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1})
	other := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Clue", TypeLine: "Artifact", Power: 0, Toughness: 0})
	land := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	korvold := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Korvold, Gleeful Glutton", TypeLine: "Legendary Creature — Dragon Noble",
		OracleID: cmdr2S52KorvoldOracle, Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID,
		Keywords: []string{"flying", "trample", "haste"},
	})
	// An artifact creature is two types; the land a third; an instant none.
	effectstest.Std51Graveyard(me, "Artifact Creature — Golem", "Basic Land — Forest", "Instant")
	hand := me.Hand.Size()
	effectstest.AttackWith(t, g, opp.ID, korvold)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePowerWithCounters(t, g, korvold); got != 7 {
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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	derevi := cmdr2S52DereviInCommandZone(me, 2)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.Std51Pay(g, me, "{1}{G}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, derevi, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the command zone: %v", err)
	}
	if !me.Command.Contains(derevi) {
		t.Error("Derevi left the command zone at activation; she enters on resolution")
	}
	effectstest.PassPriorityAroundTable(t, g)
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
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for Derevi's enters trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.LayeredCard(t, g, victim).Tapped {
		t.Error("the Bear should be tapped")
	}
}

func TestCmdr2S52DereviAbilityWorksOnlyFromTheCommandZone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	derevi := effectstest.PushCatalogPermanent(g, me.ID, "Derevi, Empyrial Tactician", "Legendary Creature — Bird Wizard", cmdr2S52DereviOracle, false)
	effectstest.Std51Pay(g, me, "{1}{G}{W}{U}")
	if err := g.ActivateCatalogAbility(me.ID, derevi, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the put-from-command-zone ability activated from the battlefield")
	}
}

func TestCmdr2S52DereviTriggersOnCombatDamageByAnyCreatureYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Derevi, Empyrial Tactician", "Legendary Creature — Bird Wizard", cmdr2S52DereviOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Tapped Thing", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, mine).Tapped = true })
	effectstest.AttackWith(t, g, opp.ID, bear)
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt after my creature connected")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: mine}}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 1) // untap it
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LayeredCard(t, g, mine).Tapped {
		t.Error("the permanent should be untapped")
	}
}

// cmdr2S52Attack declares the attack and locks it in, stopping in the
// declare attackers step so an attack trigger's prompt can be answered.
func cmdr2S52Attack(t *testing.T, g *game.Game, defender, attacker uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
}

const cmdr2S52GrothamaOracle = "34fd52c1-3d81-4005-b826-eeefbd7fd874"

func TestCmdr2S52GrothamaFightGrantedToEveryOtherCreatureAndDrawsPerController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	groth := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grothama, All-Devouring", TypeLine: "Legendary Creature — Wurm",
		OracleID: cmdr2S52GrothamaOracle, Power: 10, Toughness: 8, Owner: opp.ID, Controller: opp.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)
	effectstest.PushVanillaCreature(g, third.ID, "Their Bear", 1, 1)
	// Damage by my sources (3 from the fight) and by the opponent's own
	// source (Grothama's controller): each player draws their own total.
	before := [3]int{me.Hand.Size(), opp.Hand.Size(), third.Hand.Size()}
	cmdr2S52Attack(t, g, opp.ID, bear)
	// "you may have it fight": accept.
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, groth); got != 3 {
		t.Fatalf("Grothama has %d damage, want 3 from the fighting Bear", got)
	}
	// Grothama leaves: I drew for my 3 damage; nobody else dealt any.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(groth) })
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before[0]; got != 3 {
		t.Errorf("I drew %d, want 3", got)
	}
	if got := opp.Hand.Size() - before[1]; got != 0 {
		t.Errorf("Grothama's controller drew %d, want 0", got)
	}
	if got := third.Hand.Size() - before[2]; got != 0 {
		t.Errorf("a bystander drew %d, want 0", got)
	}
}

func TestCmdr2S52GrothamaFightIsOptional(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	groth := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Grothama, All-Devouring", TypeLine: "Legendary Creature — Wurm",
		OracleID: cmdr2S52GrothamaOracle, Power: 10, Toughness: 8, Owner: opp.ID, Controller: opp.ID,
	})
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 3, 3)
	cmdr2S52Attack(t, g, opp.ID, bear)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, groth); got != 0 {
		t.Errorf("declined fight, Grothama has %d damage, want 0", got)
	}
}
