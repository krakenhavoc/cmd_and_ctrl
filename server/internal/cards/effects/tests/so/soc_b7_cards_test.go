package testsso

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSocB7EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		effectstest.SocB7HofriOracle:             "Hofri Ghostforge",
		effectstest.SocB7ProdigyOracle:           "Ingenious Prodigy",
		effectstest.SocB7SkypainterOracle:        "Inspired Skypainter",
		effectstest.SocB7SkypainterOracle + "#1": "Maestro's Gift",
		effectstest.SocB7ChirographyOracle:       "Intermediate Chirography",
		effectstest.SocB7JadarOracle:             "Jadar, Ghoulcaller of Nephalia",
		effectstest.SocB7KamiOracle:              "Kami of Ancient Law",
		effectstest.SocB7DuelistOracle:           "Keen Duelist",
		effectstest.SocB7MentorOracle:            "Killian, Decisive Mentor",
		effectstest.SocB7InkDuelistOracle:        "Killian, Ink Duelist",
		effectstest.SocB7KineticOozeOracle:       "Kinetic Ooze",
	}
	for oracle, name := range want {
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

// socB7CardsNamed lists the battlefield cards with the name.
func socB7CardsNamed(g *game.Game, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

// --- Hofri Ghostforge ---

func TestSocB7HofriExilesTheDeadAndMakesASpiritCopyThatReturnsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", effectstest.SocB7HofriOracle, 4, 5)
	bear := effectstest.B12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", effectstest.SocB7GrizzlyBearsOracle, 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB7Zone(g, bear) != game.ZoneExile {
		t.Fatalf("the dead Bear is in %q, want exile", effectstest.SocB7Zone(g, bear))
	}
	tokens := socB7CardsNamed(g, "Grizzly Bears")
	if len(tokens) != 1 || !tokens[0].IsToken() {
		t.Fatalf("Bear tokens = %d, want one token copy", len(tokens))
	}
	tok := tokens[0].InstanceID
	if c, _ := g.LookupCardForEffect(tok); !c.HasSubtype("Spirit") || !c.HasSubtype("Bear") {
		t.Errorf("the token's types are %q, want a Bear Spirit", c.TypeLine)
	}
	if p, tg := effectstest.EffectivePower(t, g, tok), effectstest.EffectiveToughness(t, g, tok); p != 3 || tg != 3 {
		t.Errorf("the Spirit is %d/%d, want 3/3 under Hofri", p, tg)
	}
	if !effectstest.HasEffectiveKeyword(t, g, tok, "trample") || !effectstest.HasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the Spirit lacks trample or haste")
	}
	// The token dying is not "another nontoken creature": it returns the
	// exiled card and makes nothing.
	effectstest.KillCreature(t, g, me.ID, tok)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB7Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("after the token left, the Bear card is in %q, want its owner's graveyard", effectstest.SocB7Zone(g, bear))
	}
	if n := len(socB7CardsNamed(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d Bear tokens after the token died, want none", n)
	}
}

func TestSocB7HofriMakesNoTokenWhenTheCardHasLeftItsGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", effectstest.SocB7HofriOracle, 4, 5)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB7CardsNamed(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d tokens, want none: the card could not be exiled", n)
	}
	if effectstest.SocB7Zone(g, bear) != game.ZoneHand {
		t.Errorf("the Bear card is in %q, want hand", effectstest.SocB7Zone(g, bear))
	}
}

func TestSocB7HofriIgnoresAnOpponentsCreatureAndPumpsOnlyItsSpirits(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", effectstest.SocB7HofriOracle, 4, 5)
	mySpirit := effectstest.B12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	theirSpirit := effectstest.B12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if effectstest.EffectivePower(t, g, mySpirit) != 2 || effectstest.EffectivePower(t, g, theirSpirit) != 1 {
		t.Errorf("my Spirit %d, theirs %d; want 2 and 1", effectstest.EffectivePower(t, g, mySpirit), effectstest.EffectivePower(t, g, theirSpirit))
	}
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SocB7Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("an opponent's creature went to %q, want its graveyard", effectstest.SocB7Zone(g, bear))
	}
}

// --- Ingenious Prodigy ---

func TestSocB7ProdigyEntersWithXAndTradesACounterForACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	prodigy := effectstest.CastFromHandForTest(t, g, me, "Ingenious Prodigy", "Creature — Human Wizard", "{X}{U}", effectstest.SocB7ProdigyOracle,
		game.CastSpellParams{XValue: 2})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, prodigy, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, prodigy, "skulk") {
		t.Error("no skulk")
	}
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, prodigy, game.CounterPlusOne); got != 1 {
		t.Errorf("counters after the upkeep %d, want 1", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestSocB7ProdigyWithNoCounterDoesNotTrigger(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	prodigy := effectstest.B12Push(g, me.ID, "Ingenious Prodigy", "Creature — Human Wizard", effectstest.SocB7ProdigyOracle, 0, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if n := effectstest.TriggersOnStackFrom(g, prodigy); n != 0 || effectstest.LatestTriggerPrompt(g, me.ID) != nil {
		t.Errorf("a Prodigy with no counter triggered (%d on the stack)", n)
	}
}

// --- Inspired Skypainter // Maestro's Gift ---

func TestSocB7SkypainterPreparesOnEntryAndOnTokenDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sky := sos1Prep(me.ID, effectstest.SocB7SkypainterOracle, []string{"flying"}, "Inspired Skypainter", "Creature — Lizard Wizard", "{U}{R}", 2, 2,
		effectstest.SocB7SkypainterPrepared, "Sorcery", "{3}{U}{R}")
	effectstest.CastFromHandAndResolve(t, g, me, sky)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(sky.InstanceID) {
		t.Fatal("the Skypainter is not prepared after entering")
	}
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.FraBCastCopy(t, g, me, effectstest.SocB7SkypainterPrepared, game.TargetRef{Kind: game.TargetCard, ID: bear})
	if g.IsPreparedForEffect(sky.InstanceID) {
		t.Fatal("casting Maestro's Gift did not unprepare it")
	}
	var token uuid.UUID
	for _, c := range socB7CardsNamed(g, "Grizzly Bears") {
		if c.IsToken() {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("Maestro's Gift made no token copy")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") {
		t.Fatal("the token has no haste")
	}
	if effectstest.HasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the original gained haste")
	}
	// The hasty token attacks and connects: the Skypainter is prepared again.
	effectstest.AttackWith(t, g, opp.ID, token)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(sky.InstanceID) {
		t.Error("a creature token's combat damage did not prepare it")
	}
}

func TestSocB7SkypainterIgnoresANontokenHit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	sky := effectstest.B12Push(g, me.ID, "Inspired Skypainter", "Creature — Lizard Wizard", effectstest.SocB7SkypainterOracle, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TriggersOnStackFrom(g, sky); n != 0 || g.IsPreparedForEffect(sky) {
		t.Error("a nontoken creature's damage prepared it")
	}
}

// --- Intermediate Chirography ---

func TestSocB7ChirographyMakesAnInklingAndLevelTwoCountersTheFirstLoss(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	class := effectstest.CastFromHandForTest(t, g, me, "Intermediate Chirography", "Enchantment — Class", "{1}{B}", effectstest.SocB7ChirographyOracle, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Inkling") == uuid.Nil {
		t.Fatal("no Inkling on entering")
	}
	bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{B}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after the first loss %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.CountersOn(g, bear, game.CounterPlusOne) != 1 {
		t.Error("a second loss in the turn triggered again")
	}
}

func TestSocB7ChirographyLevelThreeNeedsAModifiedCreatureToDie(t *testing.T) {
	for _, tc := range []struct {
		name     string
		counter  bool
		inklings int
	}{{"a modified creature died", true, 1}, {"an unmodified creature died", false, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			effectstest.AdvanceToMain(t, g)
			class := effectstest.PushClass(g, me.ID, "Intermediate Chirography", "Enchantment — Class", effectstest.SocB7ChirographyOracle)
			if err := effectstest.LevelUpTo(t, g, me.ID, class, 0, "{1}{B}"); err != nil {
				t.Fatalf("level 2: %v", err)
			}
			if err := effectstest.LevelUpTo(t, g, me.ID, class, 1, "{2}{B}"); err != nil {
				t.Fatalf("level 3: %v", err)
			}
			bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			if tc.counter {
				g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
			}
			effectstest.KillCreature(t, g, me.ID, bear)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if n := len(socB7CardsNamed(g, "Inkling")); n != tc.inklings {
				t.Errorf("Inklings at the end step %d, want %d", n, tc.inklings)
			}
		})
	}
}

// A creature enchanted by an opponent's Aura is not modified; one wearing
// an opponent's Equipment is (CR 700.9).
func TestSocB7ModifiedCreatureDiedTallyReadsWhoseAura(t *testing.T) {
	for _, tc := range []struct {
		name     string
		subtype  string
		theirs   bool
		modified int
	}{
		{"my Aura", "Aura", false, 1},
		{"an opponent's Aura", "Aura", true, 0},
		{"an opponent's Equipment", "Equipment", true, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.AdvanceToMain(t, g)
			bear := effectstest.B12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			owner := me.ID
			if tc.theirs {
				owner = opp.ID
			}
			typeLine := "Enchantment — Aura"
			if tc.subtype == "Equipment" {
				typeLine = "Artifact — Equipment"
			}
			effectstest.PushPermanent(g, owner, game.Card{Name: "Attachment", TypeLine: typeLine,
				AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
			effectstest.KillCreature(t, g, me.ID, bear)
			if got := g.TurnTallyFor(me.ID).ModifiedCreaturesDied; got != tc.modified {
				t.Errorf("modified creatures died %d, want %d", got, tc.modified)
			}
		})
	}
}

// --- Jadar, Ghoulcaller of Nephalia ---

func TestSocB7JadarMakesADecayedZombieOnlyWithoutOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Jadar, Ghoulcaller of Nephalia", "Legendary Creature — Human Wizard", effectstest.SocB7JadarOracle, 1, 1)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	zombies := socB7CardsNamed(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("Zombies %d, want 1", len(zombies))
	}
	if !effectstest.HasEffectiveKeyword(t, g, zombies[0].InstanceID, game.KeywordDecayed) {
		t.Error("the Zombie has no decayed")
	}
	// Next end step: the decayed Zombie is still here, so no second one.
	effectstest.AdvanceToStepOf(t, g, 0, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(socB7CardsNamed(g, "Zombie")); n != 1 {
		t.Errorf("Zombies %d with a decayed creature already in play, want 1", n)
	}
}

// --- Kami of Ancient Law ---

func TestSocB7KamiSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kami := effectstest.B12Push(g, me.ID, "Kami of Ancient Law", "Creature — Spirit", effectstest.SocB7KamiOracle, 2, 2)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ench := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Some Enchantment", TypeLine: "Enchantment"})
	if err := g.ActivateCatalogAbility(me.ID, kami, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an enchantment")
	}
	effectstest.B16Activate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{Targets: effectstest.B16TargetCard(ench)})
	if g.Battlefield.Contains(ench) || g.Battlefield.Contains(kami) {
		t.Error("the enchantment is not destroyed or the Kami is not sacrificed")
	}
}

// --- Keen Duelist ---

func TestSocB7KeenDuelistTradesLifeForTheOtherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	effectstest.B12Push(g, me.ID, "Keen Duelist", "Creature — Human Wizard", effectstest.SocB7DuelistOracle, 2, 2)
	mine := effectstest.SeedLibraryTop(me, "Mine", "Sorcery", "{2}{B}")
	theirs := effectstest.SeedLibraryTop(opp, "Theirs", "Creature — Giant", "{4}{R}")
	myLife, oppLife := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.SettleAnsweringPicks(t, g, me.ID, opp.ID)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a revealed card did not reach its owner's hand")
	}
	if me.Life != myLife-5 || opp.Life != oppLife-3 {
		t.Errorf("life %d / %d, want %d / %d (each loses the other card's mana value)", me.Life, opp.Life, myLife-5, oppLife-3)
	}
}

// --- Killian, Decisive Mentor ---

func TestSocB7MentorGoadsOnAnEnchantmentAndDrawsOnAnEnchantedAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Killian, Decisive Mentor", "Legendary Creature — Human Warlock", effectstest.SocB7MentorOracle, 2, 3)
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Ogre", 3, 3)
	effectstest.CastFromHandForTest(t, g, me, "Plain Enchantment", "Enchantment", "{1}", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerTriggerTargets(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(foe)
	if !c.Tapped || !c.Goaded() {
		t.Fatalf("the target is tapped %v, goaded %v; want both", c.Tapped, c.Goaded())
	}
	a := effectstest.B12Creature(g, me.ID, "Enchanted A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Enchanted B", "Creature — Bear", 2, 2)
	plain := effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{a, b} {
		effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Aura", TypeLine: "Enchantment — Aura",
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	}
	effectstest.SeedLibrary(me, "X", "Y", "Z")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a, b, plain)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one draw for two enchanted attackers)", me.Hand.Size(), hand+1)
	}
}

func TestSocB7MentorDrawsNothingForAnOpponentsAura(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Killian, Decisive Mentor", "Legendary Creature — Human Warlock", effectstest.SocB7MentorOracle, 2, 3)
	a := effectstest.B12Creature(g, me.ID, "Enchanted", "Creature — Bear", 2, 2)
	effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment — Aura",
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: a}})
	effectstest.SeedLibrary(me, "X")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

// --- Killian, Ink Duelist ---

func TestSocB7InkDuelistDiscountsSpellsThatTargetACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	killian := effectstest.B12Push(g, me.ID, "Killian, Ink Duelist", "Legendary Creature — Human Warlock", effectstest.SocB7InkDuelistOracle, 2, 2)
	for _, kw := range []string{"lifelink", "menace"} {
		if !effectstest.HasEffectiveKeyword(t, g, killian, kw) {
			t.Errorf("no %s", kw)
		}
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	at := func(kind game.TargetRefKind, id uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: kind, ID: id}}
	}
	if got := effectstest.SelfPriced(t, g, me, effectstest.SocB7MindRotOracle, "Sorcery", "{2}{B}", game.ZoneHand, at(game.TargetPlayer, opp.ID)).ManaValue(); got != 3 {
		t.Errorf("a spell at a player costs %d, want 3", got)
	}
	if got := effectstest.SelfPriced(t, g, me, effectstest.SocB7DoomBladeOracle, "Instant", "{1}{B}", game.ZoneHand, at(game.TargetCard, bear)).ManaValue(); got != 1 {
		t.Errorf("a spell at a creature costs %d, want 1", got)
	}
	if got := effectstest.SelfPriced(t, g, opp, effectstest.SocB7DoomBladeOracle, "Instant", "{1}{B}", game.ZoneHand, at(game.TargetCard, killian)).ManaValue(); got != 2 {
		t.Errorf("an opponent's spell at a creature costs %d, want 2", got)
	}
	// A real cast: Doom Blade for {B} alone.
	effectstest.FloatMana(t, g, me, "{B}")
	blade := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: blade, Name: "Doom Blade", TypeLine: "Instant", ManaCost: "{1}{B}",
			OracleID: effectstest.SocB7DoomBladeOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{Strict: true, Targets: at(game.TargetCard, bear)}); err != nil {
		t.Fatalf("Doom Blade for {B}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived")
	}
}

// --- Kinetic Ooze ---

func TestSocB7KineticOozeDestroysUpToXAndDrawsAtFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cheap := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	dear := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{6}"})
	effectstest.SeedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	ooze := effectstest.CastFromHandForTest(t, g, me, "Kinetic Ooze", "Creature — Ooze", "{X}{G}", effectstest.SocB7KineticOozeOracle,
		game.CastSpellParams{XValue: 5})
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == dear {
			t.Error("an artifact with mana value 6 is offered at X=5")
		}
	}
	effectstest.AnswerTriggerTargets(t, g, me.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(dear) {
		t.Error("want the MV 5 artifact destroyed and the MV 6 one kept")
	}
	if got := effectstest.CountersOn(g, ooze, game.CounterPlusOne); got != 5 {
		t.Errorf("Ooze counters %d, want 5", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (X is 5)", me.Hand.Size(), hand+1)
	}
}

func TestSocB7KineticOozeAtTenDoublesOtherCreaturesCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 3) })
	effectstest.SeedLibrary(me, "A", "B")
	ooze := effectstest.CastFromHandForTest(t, g, me, "Kinetic Ooze", "Creature — Ooze", "{X}{G}", effectstest.SocB7KineticOozeOracle,
		game.CastSpellParams{XValue: 10})
	effectstest.PassPriorityAroundTable(t, g)
	// No artifact or enchantment: the first clause has nothing to offer
	// and is skipped; the second takes the Bear.
	effectstest.AnswerTriggerTargets(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 6 {
		t.Errorf("Bear counters %d, want 6", got)
	}
	if got := effectstest.CountersOn(g, ooze, game.CounterPlusOne); got != 10 {
		t.Errorf("Ooze counters %d, want 10 (not a target: \"other\")", got)
	}
}
