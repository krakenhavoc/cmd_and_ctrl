package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b7_cards_test.go — Secrets of Strixhaven Commander, batch 7
// (#4834). Each test plays the real catalog entry and asserts what its
// text says.

const (
	socB7HofriOracle        = "663d451b-f848-41cc-b541-0c6fe37d3d54"
	socB7ProdigyOracle      = "6f498ff7-38c9-44ea-9cb5-994a003825ee"
	socB7SkypainterOracle   = "d751580b-bb72-4d1d-be27-d0410f883b2a"
	socB7ChirographyOracle  = "6c948f71-5c86-450c-8828-df4d8deaf1dd"
	socB7JadarOracle        = "fb3defc0-c93a-4423-9ff9-c399f8a8ef1f"
	socB7KamiOracle         = "e937e0f5-ca35-4dce-9f05-8f2075731628"
	socB7DuelistOracle      = "b8602f19-6fc2-43c4-b0c4-2cd11d5dbda2"
	socB7MentorOracle       = "9c54bd34-667d-415c-ad26-82e0eb0ff409"
	socB7InkDuelistOracle   = "e233cb5b-7237-4271-8157-9179cc897b6c"
	socB7KineticOozeOracle  = "ab74c6b1-2099-4429-94b7-a5656ac04db3"
	socB7DoomBladeOracle    = "59e7f2ae-4535-4191-98be-3e65b6b2befa"
	socB7MindRotOracle      = "ad44cf74-b717-48fb-9fa2-77512024d76a"
	socB7SkypainterPrepared = "Maestro's Gift"
	socB7GrizzlyBearsOracle = "14c8f55d-d177-4c25-a931-ebeb9e6062a0"
)

func TestSocB7EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB7HofriOracle:             "Hofri Ghostforge",
		socB7ProdigyOracle:           "Ingenious Prodigy",
		socB7SkypainterOracle:        "Inspired Skypainter",
		socB7SkypainterOracle + "#1": "Maestro's Gift",
		socB7ChirographyOracle:       "Intermediate Chirography",
		socB7JadarOracle:             "Jadar, Ghoulcaller of Nephalia",
		socB7KamiOracle:              "Kami of Ancient Law",
		socB7DuelistOracle:           "Keen Duelist",
		socB7MentorOracle:            "Killian, Decisive Mentor",
		socB7InkDuelistOracle:        "Killian, Ink Duelist",
		socB7KineticOozeOracle:       "Kinetic Ooze",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// socB7Zone is the zone the card is in now, or "" when it is nowhere.
func socB7Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", socB7HofriOracle, 4, 5)
	bear := b12Push(g, me.ID, "Grizzly Bears", "Creature — Bear", socB7GrizzlyBearsOracle, 2, 2)
	killCreature(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, bear) != game.ZoneExile {
		t.Fatalf("the dead Bear is in %q, want exile", socB7Zone(g, bear))
	}
	tokens := socB7CardsNamed(g, "Grizzly Bears")
	if len(tokens) != 1 || !tokens[0].IsToken() {
		t.Fatalf("Bear tokens = %d, want one token copy", len(tokens))
	}
	tok := tokens[0].InstanceID
	if c, _ := g.LookupCardForEffect(tok); !c.HasSubtype("Spirit") || !c.HasSubtype("Bear") {
		t.Errorf("the token's types are %q, want a Bear Spirit", c.TypeLine)
	}
	if p, tg := effectivePower(t, g, tok), effectiveToughness(t, g, tok); p != 3 || tg != 3 {
		t.Errorf("the Spirit is %d/%d, want 3/3 under Hofri", p, tg)
	}
	if !hasEffectiveKeyword(t, g, tok, "trample") || !hasEffectiveKeyword(t, g, tok, "haste") {
		t.Error("the Spirit lacks trample or haste")
	}
	// The token dying is not "another nontoken creature": it returns the
	// exiled card and makes nothing.
	killCreature(t, g, me.ID, tok)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("after the token left, the Bear card is in %q, want its owner's graveyard", socB7Zone(g, bear))
	}
	if n := len(socB7CardsNamed(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d Bear tokens after the token died, want none", n)
	}
}

func TestSocB7HofriMakesNoTokenWhenTheCardHasLeftItsGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", socB7HofriOracle, 4, 5)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	killCreature(t, g, me.ID, bear)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB7CardsNamed(g, "Grizzly Bears")); n != 0 {
		t.Errorf("%d tokens, want none: the card could not be exiled", n)
	}
	if socB7Zone(g, bear) != game.ZoneHand {
		t.Errorf("the Bear card is in %q, want hand", socB7Zone(g, bear))
	}
}

func TestSocB7HofriIgnoresAnOpponentsCreatureAndPumpsOnlyItsSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Hofri Ghostforge", "Legendary Creature — Dwarf Cleric", socB7HofriOracle, 4, 5)
	mySpirit := b12Creature(g, me.ID, "My Spirit", "Creature — Spirit", 1, 1)
	theirSpirit := b12Creature(g, opp.ID, "Their Spirit", "Creature — Spirit", 1, 1)
	if effectivePower(t, g, mySpirit) != 2 || effectivePower(t, g, theirSpirit) != 1 {
		t.Errorf("my Spirit %d, theirs %d; want 2 and 1", effectivePower(t, g, mySpirit), effectivePower(t, g, theirSpirit))
	}
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	killCreature(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if socB7Zone(g, bear) != game.ZoneGraveyard {
		t.Errorf("an opponent's creature went to %q, want its graveyard", socB7Zone(g, bear))
	}
}

// --- Ingenious Prodigy ---

func TestSocB7ProdigyEntersWithXAndTradesACounterForACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	prodigy := castFromHandForTest(t, g, me, "Ingenious Prodigy", "Creature — Human Wizard", "{X}{U}", socB7ProdigyOracle,
		game.CastSpellParams{XValue: 2})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, prodigy, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	if !hasEffectiveKeyword(t, g, prodigy, "skulk") {
		t.Error("no skulk")
	}
	advanceToUpkeepOf(t, g, 0)
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, prodigy, game.CounterPlusOne); got != 1 {
		t.Errorf("counters after the upkeep %d, want 1", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestSocB7ProdigyWithNoCounterDoesNotTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	prodigy := b12Push(g, me.ID, "Ingenious Prodigy", "Creature — Human Wizard", socB7ProdigyOracle, 0, 1)
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	if n := triggersOnStackFrom(g, prodigy); n != 0 || latestTriggerPrompt(g, me.ID) != nil {
		t.Errorf("a Prodigy with no counter triggered (%d on the stack)", n)
	}
}

// --- Inspired Skypainter // Maestro's Gift ---

func TestSocB7SkypainterPreparesOnEntryAndOnTokenDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sky := sos1Prep(me.ID, socB7SkypainterOracle, []string{"flying"}, "Inspired Skypainter", "Creature — Lizard Wizard", "{U}{R}", 2, 2,
		socB7SkypainterPrepared, "Sorcery", "{3}{U}{R}")
	castFromHandAndResolve(t, g, me, sky)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(sky.InstanceID) {
		t.Fatal("the Skypainter is not prepared after entering")
	}
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	fraBCastCopy(t, g, me, socB7SkypainterPrepared, game.TargetRef{Kind: game.TargetCard, ID: bear})
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
	if !hasEffectiveKeyword(t, g, token, "haste") {
		t.Fatal("the token has no haste")
	}
	if hasEffectiveKeyword(t, g, bear, "haste") {
		t.Error("the original gained haste")
	}
	// The hasty token attacks and connects: the Skypainter is prepared again.
	attackWith(t, g, opp.ID, token)
	passPriorityAroundTable(t, g)
	if !g.IsPreparedForEffect(sky.InstanceID) {
		t.Error("a creature token's combat damage did not prepare it")
	}
}

func TestSocB7SkypainterIgnoresANontokenHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	sky := b12Push(g, me.ID, "Inspired Skypainter", "Creature — Lizard Wizard", socB7SkypainterOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	attackWith(t, g, opp.ID, bear)
	passPriorityAroundTable(t, g)
	if n := triggersOnStackFrom(g, sky); n != 0 || g.IsPreparedForEffect(sky) {
		t.Error("a nontoken creature's damage prepared it")
	}
}

// --- Intermediate Chirography ---

func TestSocB7ChirographyMakesAnInklingAndLevelTwoCountersTheFirstLoss(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	class := castFromHandForTest(t, g, me, "Intermediate Chirography", "Enchantment — Class", "{1}{B}", socB7ChirographyOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Inkling") == uuid.Nil {
		t.Fatal("no Inkling on entering")
	}
	bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	if err := levelUpTo(t, g, me.ID, class, 0, "{1}{B}"); err != nil {
		t.Fatalf("level 2: %v", err)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -2) })
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Fatalf("counters after the first loss %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, -1) })
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil || countersOn(g, bear, game.CounterPlusOne) != 1 {
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			advanceToMain(t, g)
			class := pushClass(g, me.ID, "Intermediate Chirography", "Enchantment — Class", socB7ChirographyOracle)
			if err := levelUpTo(t, g, me.ID, class, 0, "{1}{B}"); err != nil {
				t.Fatalf("level 2: %v", err)
			}
			if err := levelUpTo(t, g, me.ID, class, 1, "{2}{B}"); err != nil {
				t.Fatalf("level 3: %v", err)
			}
			bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			if tc.counter {
				g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
			}
			killCreature(t, g, me.ID, bear)
			passPriorityAroundTable(t, g)
			advanceTo(t, g, game.StepEnd)
			passPriorityAroundTable(t, g)
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
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			advanceToMain(t, g)
			bear := b12Creature(g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
			owner := me.ID
			if tc.theirs {
				owner = opp.ID
			}
			typeLine := "Enchantment — Aura"
			if tc.subtype == "Equipment" {
				typeLine = "Artifact — Equipment"
			}
			pushPermanent(g, owner, game.Card{Name: "Attachment", TypeLine: typeLine,
				AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: bear}})
			killCreature(t, g, me.ID, bear)
			if got := g.TurnTallyFor(me.ID).ModifiedCreaturesDied; got != tc.modified {
				t.Errorf("modified creatures died %d, want %d", got, tc.modified)
			}
		})
	}
}

// --- Jadar, Ghoulcaller of Nephalia ---

func TestSocB7JadarMakesADecayedZombieOnlyWithoutOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Jadar, Ghoulcaller of Nephalia", "Legendary Creature — Human Wizard", socB7JadarOracle, 1, 1)
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	zombies := socB7CardsNamed(g, "Zombie")
	if len(zombies) != 1 {
		t.Fatalf("Zombies %d, want 1", len(zombies))
	}
	if !hasEffectiveKeyword(t, g, zombies[0].InstanceID, game.KeywordDecayed) {
		t.Error("the Zombie has no decayed")
	}
	// Next end step: the decayed Zombie is still here, so no second one.
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if n := len(socB7CardsNamed(g, "Zombie")); n != 1 {
		t.Errorf("Zombies %d with a decayed creature already in play, want 1", n)
	}
}

// --- Kami of Ancient Law ---

func TestSocB7KamiSacrificesToDestroyAnEnchantment(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	kami := b12Push(g, me.ID, "Kami of Ancient Law", "Creature — Spirit", socB7KamiOracle, 2, 2)
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	ench := pushPermanent(g, opp.ID, game.Card{Name: "Some Enchantment", TypeLine: "Enchantment"})
	if err := g.ActivateCatalogAbility(me.ID, kami, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a creature is not an enchantment")
	}
	b16Activate(t, g, me.ID, kami, 0, game.ActivateAbilityParams{Targets: b16TargetCard(ench)})
	if g.Battlefield.Contains(ench) || g.Battlefield.Contains(kami) {
		t.Error("the enchantment is not destroyed or the Kami is not sacrificed")
	}
}

// --- Keen Duelist ---

func TestSocB7KeenDuelistTradesLifeForTheOtherCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[1], g.Seats[2]
	b12Push(g, me.ID, "Keen Duelist", "Creature — Human Wizard", socB7DuelistOracle, 2, 2)
	mine := seedLibraryTop(me, "Mine", "Sorcery", "{2}{B}")
	theirs := seedLibraryTop(opp, "Theirs", "Creature — Giant", "{4}{R}")
	myLife, oppLife := me.Life, opp.Life
	advanceToUpkeepOf(t, g, 1)
	settleAnsweringPicks(t, g, me.ID, opp.ID)
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a revealed card did not reach its owner's hand")
	}
	if me.Life != myLife-5 || opp.Life != oppLife-3 {
		t.Errorf("life %d / %d, want %d / %d (each loses the other card's mana value)", me.Life, opp.Life, myLife-5, oppLife-3)
	}
}

// --- Killian, Decisive Mentor ---

func TestSocB7MentorGoadsOnAnEnchantmentAndDrawsOnAnEnchantedAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Killian, Decisive Mentor", "Legendary Creature — Human Warlock", socB7MentorOracle, 2, 3)
	foe := b12Creature(g, opp.ID, "Foe", "Creature — Ogre", 3, 3)
	castFromHandForTest(t, g, me, "Plain Enchantment", "Enchantment", "{1}", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	answerTriggerTargets(t, g, me.ID, foe)
	passPriorityAroundTable(t, g)
	c, _ := g.LookupCardForEffect(foe)
	if !c.Tapped || !c.Goaded() {
		t.Fatalf("the target is tapped %v, goaded %v; want both", c.Tapped, c.Goaded())
	}
	a := b12Creature(g, me.ID, "Enchanted A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Enchanted B", "Creature — Bear", 2, 2)
	plain := b12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{a, b} {
		pushPermanent(g, me.ID, game.Card{Name: "My Aura", TypeLine: "Enchantment — Aura",
			AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	}
	seedLibrary(me, "X", "Y", "Z")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a, b, plain)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (one draw for two enchanted attackers)", me.Hand.Size(), hand+1)
	}
}

func TestSocB7MentorDrawsNothingForAnOpponentsAura(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Killian, Decisive Mentor", "Legendary Creature — Human Warlock", socB7MentorOracle, 2, 3)
	a := b12Creature(g, me.ID, "Enchanted", "Creature — Bear", 2, 2)
	pushPermanent(g, opp.ID, game.Card{Name: "Their Aura", TypeLine: "Enchantment — Aura",
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: a}})
	seedLibrary(me, "X")
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, a)
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
}

// --- Killian, Ink Duelist ---

func TestSocB7InkDuelistDiscountsSpellsThatTargetACreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	killian := b12Push(g, me.ID, "Killian, Ink Duelist", "Legendary Creature — Human Warlock", socB7InkDuelistOracle, 2, 2)
	for _, kw := range []string{"lifelink", "menace"} {
		if !hasEffectiveKeyword(t, g, killian, kw) {
			t.Errorf("no %s", kw)
		}
	}
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	at := func(kind game.TargetRefKind, id uuid.UUID) []game.TargetRef {
		return []game.TargetRef{{Kind: kind, ID: id}}
	}
	if got := selfPriced(t, g, me, socB7MindRotOracle, "Sorcery", "{2}{B}", game.ZoneHand, at(game.TargetPlayer, opp.ID)).ManaValue(); got != 3 {
		t.Errorf("a spell at a player costs %d, want 3", got)
	}
	if got := selfPriced(t, g, me, socB7DoomBladeOracle, "Instant", "{1}{B}", game.ZoneHand, at(game.TargetCard, bear)).ManaValue(); got != 1 {
		t.Errorf("a spell at a creature costs %d, want 1", got)
	}
	if got := selfPriced(t, g, opp, socB7DoomBladeOracle, "Instant", "{1}{B}", game.ZoneHand, at(game.TargetCard, killian)).ManaValue(); got != 2 {
		t.Errorf("an opponent's spell at a creature costs %d, want 2", got)
	}
	// A real cast: Doom Blade for {B} alone.
	floatMana(t, g, me, "{B}")
	blade := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: blade, Name: "Doom Blade", TypeLine: "Instant", ManaCost: "{1}{B}",
			OracleID: socB7DoomBladeOracle, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, blade, game.CastSpellParams{Strict: true, Targets: at(game.TargetCard, bear)}); err != nil {
		t.Fatalf("Doom Blade for {B}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived")
	}
}

// --- Kinetic Ooze ---

func TestSocB7KineticOozeDestroysUpToXAndDrawsAtFive(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	cheap := pushPermanent(g, opp.ID, game.Card{Name: "Cheap Rock", TypeLine: "Artifact", ManaCost: "{5}"})
	dear := pushPermanent(g, opp.ID, game.Card{Name: "Dear Rock", TypeLine: "Artifact", ManaCost: "{6}"})
	seedLibrary(me, "A", "B")
	hand := me.Hand.Size()
	ooze := castFromHandForTest(t, g, me, "Kinetic Ooze", "Creature — Ooze", "{X}{G}", socB7KineticOozeOracle,
		game.CastSpellParams{XValue: 5})
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for _, id := range pick.PickTargetCards {
		if id == dear {
			t.Error("an artifact with mana value 6 is offered at X=5")
		}
	}
	answerTriggerTargets(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !g.Battlefield.Contains(dear) {
		t.Error("want the MV 5 artifact destroyed and the MV 6 one kept")
	}
	if got := countersOn(g, ooze, game.CounterPlusOne); got != 5 {
		t.Errorf("Ooze counters %d, want 5", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (X is 5)", me.Hand.Size(), hand+1)
	}
}

func TestSocB7KineticOozeAtTenDoublesOtherCreaturesCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 3) })
	seedLibrary(me, "A", "B")
	ooze := castFromHandForTest(t, g, me, "Kinetic Ooze", "Creature — Ooze", "{X}{G}", socB7KineticOozeOracle,
		game.CastSpellParams{XValue: 10})
	passPriorityAroundTable(t, g)
	// No artifact or enchantment: the first clause has nothing to offer
	// and is skipped; the second takes the Bear.
	answerTriggerTargets(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 6 {
		t.Errorf("Bear counters %d, want 6", got)
	}
	if got := countersOn(g, ooze, game.CounterPlusOne); got != 10 {
		t.Errorf("Ooze counters %d, want 10 (not a target: \"other\")", got)
	}
}
