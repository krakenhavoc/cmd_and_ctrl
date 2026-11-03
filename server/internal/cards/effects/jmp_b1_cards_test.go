package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b1_cards_test.go — the Jumpstart Dimir deck request's first batch
// (#4892): every card driven through the engine, with the refusals.

const (
	jmpB1BarterInBloodOracle    = "9167998d-5cac-47d7-99f2-f38122f7b8e7"
	jmpB1BelltowerSphinxOracle  = "b9ed0081-7eda-4ff9-ac2d-bcdc29a22d9f"
	jmpB1BloodDivinationOracle  = "c1cf9d74-e700-456a-9e5a-1c4df22db268"
	jmpB1BloodbondVampireOracle = "a2057118-9fd1-431d-9441-fef198d67ff2"
	jmpB1BogbrewWitchOracle     = "ad138184-ab2d-407a-9f9c-94b0675be19d"
	jmpB1BoneSplintersOracle    = "0c936582-d4c0-4da8-bc7e-49da5a433939"
	jmpB1BubblingCauldronOracle = "ca9dde44-fdee-4bc1-a2ea-ef171737b9e4"
	jmpB1ExcludeOracle          = "a8c9f91a-b1e7-451d-b6db-ae865e2b853c"
	jmpB1FesteringNewtOracle    = "5887922e-d1fb-44e2-9498-2065ce8ac870"
	jmpB1GoremandOracle         = "0f80f061-6304-4c7b-ab53-f406385e39df"
	jmpB1KelsOracle             = "fe2e27a9-437e-4d22-aa22-982978793cff"
)

func TestJmpB1Registered(t *testing.T) {
	want := map[string]string{
		jmpB1BarterInBloodOracle:    "Barter in Blood",
		jmpB1BelltowerSphinxOracle:  "Belltower Sphinx",
		jmpB1BloodDivinationOracle:  "Blood Divination",
		jmpB1BloodbondVampireOracle: "Bloodbond Vampire",
		jmpB1BogbrewWitchOracle:     "Bogbrew Witch",
		jmpB1BoneSplintersOracle:    "Bone Splinters",
		jmpB1BubblingCauldronOracle: "Bubbling Cauldron",
		jmpB1ExcludeOracle:          "Exclude",
		jmpB1FesteringNewtOracle:    "Festering Newt",
		jmpB1GoremandOracle:         "Goremand",
		jmpB1KelsOracle:             "Kels, Fight Fixer",
	}
	for oracle, name := range want {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered as %q, completeness %v, caveats %v", oracle, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// jmpB1Settle drives the table until the stack is empty and nothing is
// pending, answering the prompts these cards ask: a trigger-order prompt
// in the order offered, a pick_target with `target`, and each pay-unless
// offer to `me` with pay(i) (adding the {U} when it pays). It returns how
// many pay-unless offers were made.
func jmpB1Settle(t *testing.T, g *game.Game, me *game.Player, target uuid.UUID, pay func(i int) bool) int {
	t.Helper()
	offers := 0
	for i := 0; i < 96; i++ {
		if ch := triggerOrderPrompt(g); ch != nil {
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if pick := jmpB1AnyPickTarget(g); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, pick.Chooser, game.TargetRef{Kind: game.TargetCard, ID: target}); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			continue
		}
		if hasPayUnlessFor(g, me.ID) {
			p := pay != nil && pay(offers)
			offers++
			if p {
				me.ManaPool.AddMana(game.ManaToken{Color: "U"})
			}
			answerPayUnless(t, g, me.ID, p)
			continue
		}
		if stackFullyEmpty(g) {
			return offers
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				t.Fatalf("an unexpected prompt is pending: %+v", g.PendingChoices[0])
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
	return offers
}

func jmpB1AnyPickTarget(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			return c
		}
	}
	return nil
}

func jmpB1Mana(p *game.Player, colors ...string) {
	for _, c := range colors {
		p.ManaPool.AddMana(game.ManaToken{Color: c})
	}
}

// --- Barter in Blood --------------------------------------------------

// Each player sacrifices two creatures of their own choice; fewer than two
// sacrifices what they have, none is not asked, and nothing else goes.
func TestJmpB1BarterInBloodEachPlayerSacrificesTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := b12Creature(g, me.ID, "Mine A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Mine B", "Creature — Bear", 2, 2)
	keep := b12Creature(g, me.ID, "Mine Kept", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	x := b12Creature(g, opp1.ID, "Theirs X", "Creature — Bear", 2, 2)
	y := b12Creature(g, opp1.ID, "Theirs Y", "Creature — Bear", 2, 2)
	z := b12Creature(g, opp2.ID, "Lone Z", "Creature — Bear", 2, 2)
	land := b12Permanent(g, opp3.ID, "Their Land", "Basic Land — Swamp")

	castCatalogSpell(t, g, "Barter in Blood", "Sorcery", jmpB1BarterInBloodOracle, nil)
	passPriorityAroundTable(t, g)

	c := sacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the caster is asked to choose two of three creatures")
	}
	if edictOffers(sacrificeOptions(c), rock) {
		t.Error("a noncreature is offered")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{a}); err == nil {
		t.Error("choosing one creature of three is refused")
	}
	answerSacrifice(t, g, me.ID, a, b)
	for _, p := range []*game.Player{opp1, opp2, opp3} {
		answerSacrifice(t, g, p.ID, sacrificeOptions(sacrificeChoiceFor(g, p.ID))...)
	}
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{a, b, x, y, z} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived", id)
		}
	}
	for _, id := range []uuid.UUID{keep, rock, land} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed but is not a chosen creature", id)
		}
	}
	if sacrificeChoiceFor(g, opp3.ID) != nil {
		t.Error("a player with no creature is asked")
	}
}

// --- Belltower Sphinx -------------------------------------------------

// A Bolt at the Sphinx mills the Bolt's controller three; damage elsewhere
// mills nobody.
func TestJmpB1BelltowerSphinxMillsTheDamagersController(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sphinx := b12Push(g, me.ID, "Belltower Sphinx", "Creature — Sphinx", jmpB1BelltowerSphinxOracle, 2, 5)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 5)
	mine, theirs := me.Library.Size(), opp.Library.Size()

	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if opp.Library.Size() != theirs || me.Library.Size() != mine {
		t.Fatal("damage to another creature milled someone")
	}

	batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetCard(sphinx))
	passPriorityAroundTable(t, g)
	if got := theirs - opp.Library.Size(); got != 3 {
		t.Errorf("the Bolt's controller milled %d, want 3", got)
	}
	if me.Library.Size() != mine {
		t.Error("the Sphinx's controller milled")
	}
	if !g.Battlefield.Contains(sphinx) {
		t.Error("3 damage killed a 2/5")
	}
}

// --- Blood Divination -------------------------------------------------

func TestJmpB1BloodDivinationSacrificesThenDrawsThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	if _, err := tryCastWithSacrifice(g, "Blood Divination", "Sorcery", jmpB1BloodDivinationOracle, nil); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	if _, err := tryCastWithSacrifice(g, "Blood Divination", "Sorcery", jmpB1BloodDivinationOracle, []uuid.UUID{theirs}); err == nil {
		t.Error("an opponent's creature is refused")
	}
	if _, err := tryCastWithSacrifice(g, "Blood Divination", "Sorcery", jmpB1BloodDivinationOracle, []uuid.UUID{rock}); err == nil {
		t.Error("a noncreature is refused")
	}
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	castWithSacrifice(t, g, "Blood Divination", "Sorcery", jmpB1BloodDivinationOracle, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrifice is paid while casting")
	}
	passPriorityAroundTable(t, g)
	// castWithSacrifice pushes the spell into the hand and casts it, so
	// the count before it is the baseline.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("drew %d, want 3", got)
	}
}

// --- Bloodbond Vampire and Bubbling Cauldron ------------------------

// The Cauldron's 4 life is one gain event: one counter, and an opponent's
// Vampire gets none.
func TestJmpB1BloodbondVampireGrowsOncePerGainAndCauldronGainsFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vamp := b12Push(g, me.ID, "Bloodbond Vampire", "Creature — Vampire Shaman Ally", jmpB1BloodbondVampireOracle, 3, 3)
	theirVamp := b12Push(g, opp.ID, "Bloodbond Vampire", "Creature — Vampire Shaman Ally", jmpB1BloodbondVampireOracle, 3, 3)
	cauldron := pushCatalogPermanent(g, me.ID, "Bubbling Cauldron", "Artifact", jmpB1BubblingCauldronOracle, false)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	life := me.Life

	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	jmpB1Settle(t, g, me, uuid.Nil, nil)
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4", life, me.Life)
	}
	if n := pr7Counters(g, vamp); n != 1 {
		t.Errorf("%d counters, want 1 (one gain event)", n)
	}
	if n := pr7Counters(g, theirVamp); n != 0 {
		t.Errorf("an opponent's gain grew their Vampire: %d", n)
	}
	// Tapped: the Cauldron can't be activated again.
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{other}, Strict: true}); err == nil {
		t.Error("a tapped Cauldron activated")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a refused activation paid its sacrifice")
	}
}

// The second ability takes only a Festering Newt, and drains each opponent
// for 4, gaining the total.
func TestJmpB1BubblingCauldronNewtDrainsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cauldron := pushCatalogPermanent(g, me.ID, "Bubbling Cauldron", "Artifact", jmpB1BubblingCauldronOracle, false)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	newt := b12Push(g, me.ID, "Festering Newt", "Creature — Salamander", jmpB1FesteringNewtOracle, 1, 1)
	lives := b17Life(g)

	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{bear}, Strict: true}); err == nil {
		t.Fatal("a creature not named Festering Newt paid the Newt cost")
	}
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the refused activation sacrificed the Bear")
	}
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 1, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{newt}, Strict: true}); err != nil {
		t.Fatalf("activate with the Newt: %v", err)
	}
	// The Newt's own dies trigger wants a target: the Bear is mine, so it
	// takes an opponent's creature or nothing. Seed one.
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	jmpB1Settle(t, g, me, theirs, nil)
	opponents := len(g.Seats) - 1
	for i, p := range g.Seats {
		want := lives[i] - 4
		if i == 0 {
			want = lives[i] + 4*opponents
		}
		if p.Life != want {
			t.Errorf("seat %d life %d, want %d", i, p.Life, want)
		}
	}
}

// --- Bone Splinters ---------------------------------------------------

func TestJmpB1BoneSplintersSacrificesAndDestroys(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	target := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	advanceToMain(t, g)

	id := b13HandCard(me, "Bone Splinters", "Sorcery", jmpB1BoneSplintersOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(target)}); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(rock), SacrificeIDs: []uuid.UUID{victim}}); err == nil {
		t.Error("a noncreature target is refused")
	}
	if !g.Battlefield.Contains(victim) {
		t.Fatal("a refused cast paid its sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(target), SacrificeIDs: []uuid.UUID{victim}}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) || g.Battlefield.Contains(target) {
		t.Error("the victim is sacrificed and the target destroyed")
	}
	if !g.Battlefield.Contains(rock) {
		t.Error("the rock was touched")
	}
}

// --- Exclude ----------------------------------------------------------

func TestJmpB1ExcludeCountersACreatureSpellAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bolt := b13OpponentCasts(t, g, opp, "Their Bolt", "Instant", "", "{R}", nil)
	id := b13HandCard(me, "Exclude", "Instant", jmpB1ExcludeOracle)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bolt)}); err == nil {
		t.Error("a noncreature spell is refused as a target")
	}
	passPriorityAroundTable(t, g)

	bear := b13PlayAs(t, g, 1, "Their Bear", "Creature — Bear", "")
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: b16TargetCard(bear)}); err != nil {
		t.Fatalf("Exclude at a creature spell: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the creature spell is countered")
	}
	if got := me.Hand.Size() - (hand - 1); got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

// --- Festering Newt and Bogbrew Witch ---------------------------------

// jmpB1KillNewt sacrifices the Newt to a Cauldron and settles, aiming the
// dies trigger at `target`.
func jmpB1KillNewt(t *testing.T, g *game.Game, me *game.Player, newt, target uuid.UUID) {
	t.Helper()
	cauldron := pushCatalogPermanent(g, me.ID, "Bubbling Cauldron", "Artifact", jmpB1BubblingCauldronOracle, false)
	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{newt}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	jmpB1Settle(t, g, me, target, nil)
}

func TestJmpB1FesteringNewtShrinksByOneWithoutTheWitch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	newt := b12Push(g, me.ID, "Festering Newt", "Creature — Salamander", jmpB1FesteringNewtOracle, 1, 1)
	ox := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	// An opponent's Witch is not "you control".
	b12Push(g, opp.ID, "Bogbrew Witch", "Creature — Human Wizard", jmpB1BogbrewWitchOracle, 1, 3)
	jmpB1KillNewt(t, g, me, newt, ox)
	if p, tough := effectivePTOf(t, g, ox); p != 2 || tough != 2 {
		t.Errorf("the Ox is %d/%d, want 2/2", p, tough)
	}
}

func TestJmpB1FesteringNewtShrinksByFourWithYourWitch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	newt := b12Push(g, me.ID, "Festering Newt", "Creature — Salamander", jmpB1FesteringNewtOracle, 1, 1)
	b12Push(g, me.ID, "Bogbrew Witch", "Creature — Human Wizard", jmpB1BogbrewWitchOracle, 1, 3)
	ox := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 4, 5)
	jmpB1KillNewt(t, g, me, newt, ox)
	if p, tough := effectivePTOf(t, g, ox); p != 0 || tough != 1 {
		t.Errorf("the Ox is %d/%d, want 0/1", p, tough)
	}
}

// The trigger targets only a creature an opponent controls.
func TestJmpB1FesteringNewtRefusesYourOwnCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	newt := b12Push(g, me.ID, "Festering Newt", "Creature — Salamander", jmpB1FesteringNewtOracle, 1, 1)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	ox := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	cauldron := pushCatalogPermanent(g, me.ID, "Bubbling Cauldron", "Artifact", jmpB1BubblingCauldronOracle, false)
	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, cauldron, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{newt}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for i := 0; i < 8 && jmpB1AnyPickTarget(g) == nil; i++ {
		if ch := triggerOrderPrompt(g); ch != nil {
			if err := g.ResolveTriggerOrder(ch.ID, ch.Chooser, ch.TriggerOrderIDs); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
			t.Fatal(err)
		}
	}
	pick := jmpB1AnyPickTarget(g)
	if pick == nil {
		t.Fatal("the dies trigger asked for no target")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: mine}); err == nil {
		t.Error("your own creature was accepted as the target")
	}
	jmpB1Settle(t, g, me, ox, nil)
	if p, tough := effectivePTOf(t, g, mine); p != 2 || tough != 2 {
		t.Errorf("your creature changed: %d/%d", p, tough)
	}
}

// The Witch finds a Newt or a Cauldron, and only those, and puts it onto
// the battlefield tapped.
func TestJmpB1BogbrewWitchFetchesANamedCardTapped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	witch := b12Push(g, me.ID, "Bogbrew Witch", "Creature — Human Wizard", jmpB1BogbrewWitchOracle, 1, 3)
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Grizzly Bears", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Bubbling Cauldron", TypeLine: "Artifact", OracleID: jmpB1BubblingCauldronOracle, Owner: me.ID, Controller: me.ID})
	newt := pushLibraryCardForTest(me, game.Card{InstanceID: uuid.New(), Name: "Festering Newt", TypeLine: "Creature — Salamander", OracleID: jmpB1FesteringNewtOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})

	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without {2}")
	}
	jmpB1Mana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if searchOptionNamed(g, c, "Grizzly Bears") != uuid.Nil {
		t.Error("a card with another name is offered")
	}
	if searchOptionNamed(g, c, "Bubbling Cauldron") == uuid.Nil {
		t.Error("the Cauldron is not offered")
	}
	answerSearchByID(t, g, me.ID, newt)
	passPriorityAroundTable(t, g)
	card, ok := battlefieldCard(g, newt)
	if !ok || !card.Tapped {
		t.Errorf("the Newt is on the battlefield tapped: present %v, tapped %v", ok, card.Tapped)
	}
	if w, _ := battlefieldCard(g, witch); !w.Tapped {
		t.Error("the Witch did not tap")
	}
}

func TestJmpB1BogbrewWitchNeedsToBeUntappedAndNotSick(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	witch := pushCatalogPermanent(g, me.ID, "Bogbrew Witch", "Creature — Human Wizard", jmpB1BogbrewWitchOracle, true)
	jmpB1Mana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, witch, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("a summoning-sick Witch tapped for its ability")
	}
}

// --- Goremand ---------------------------------------------------------

func TestJmpB1GoremandCostsACreatureAndEdictsEachOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	if _, err := tryCastWithSacrifice(g, "Goremand", "Creature — Demon", jmpB1GoremandOracle, nil); err == nil {
		t.Error("cast with no sacrifice is refused")
	}
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 2, 2)
	keep := b12Creature(g, me.ID, "Kept", "Creature — Bear", 2, 2)
	x := b12Creature(g, opp1.ID, "Theirs X", "Creature — Bear", 2, 2)
	y := b12Creature(g, opp1.ID, "Theirs Y", "Creature — Bear", 2, 2)
	z := b12Creature(g, opp2.ID, "Lone Z", "Creature — Bear", 2, 2)
	goremand := castWithSacrifice(t, g, "Goremand", "Creature — Demon", jmpB1GoremandOracle, victim)
	if g.Battlefield.Contains(victim) {
		t.Error("the sacrifice is paid while casting")
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if sacrificeChoiceFor(g, me.ID) != nil {
		t.Error("the Goremand's controller is asked")
	}
	answerSacrifice(t, g, opp1.ID, x)
	answerSacrifice(t, g, opp2.ID, sacrificeOptions(sacrificeChoiceFor(g, opp2.ID))...)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(x) || !g.Battlefield.Contains(y) || g.Battlefield.Contains(z) {
		t.Error("each opponent sacrifices exactly one creature of their choice")
	}
	if sacrificeChoiceFor(g, opp3.ID) != nil {
		t.Error("an opponent with no creature is asked")
	}
	if !g.Battlefield.Contains(goremand) || !g.Battlefield.Contains(keep) {
		t.Error("your own creatures stay")
	}
	if !effectiveAbilitiesContain(t, g, goremand, "flying") || !effectiveAbilitiesContain(t, g, goremand, "trample") {
		t.Error("Goremand has flying and trample")
	}
}

// --- Kels, Fight Fixer ------------------------------------------------

// The sacrifice-a-creature cost is a sacrifice: Kels offers {U/B}, a paid
// offer draws, and the ability makes Kels indestructible.
func TestJmpB1KelsPaysToDrawOffItsOwnOutlet(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kels := b12Push(g, me.ID, "Kels, Fight Fixer", "Legendary Creature — Azra Warlock", jmpB1KelsOracle, 4, 3)
	fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, kels, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err == nil {
		t.Fatal("activated without {1}")
	}
	jmpB1Mana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, kels, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	hand := me.Hand.Size()
	if offers := jmpB1Settle(t, g, me, uuid.Nil, func(int) bool { return true }); offers != 1 {
		t.Fatalf("%d offers, want 1", offers)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
	if !effectiveAbilitiesContain(t, g, kels, "indestructible") {
		t.Error("Kels did not gain indestructible")
	}
	if !effectiveAbilitiesContain(t, g, kels, "menace") {
		t.Error("Kels has menace")
	}
}

// Barter in Blood: each creature you sacrifice is its own offer, an
// opponent's sacrifice is none, and a declined offer draws nothing.
func TestJmpB1KelsTriggersOncePerCreatureYouSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Kels, Fight Fixer", "Legendary Creature — Azra Warlock", jmpB1KelsOracle, 4, 3)
	a := b12Creature(g, me.ID, "Mine A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Mine B", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs X", "Creature — Bear", 2, 2)
	b12Creature(g, opp.ID, "Theirs Y", "Creature — Bear", 2, 2)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")

	castCatalogSpell(t, g, "Barter in Blood", "Sorcery", jmpB1BarterInBloodOracle, nil)
	passPriorityAroundTable(t, g)
	answerSacrifice(t, g, me.ID, a, b)
	for _, p := range g.Seats[1:] {
		if c := sacrificeChoiceFor(g, p.ID); c != nil {
			answerSacrifice(t, g, p.ID, sacrificeOptions(c)...)
		}
	}
	hand := me.Hand.Size()
	offers := jmpB1Settle(t, g, me, uuid.Nil, func(i int) bool { return i == 0 })
	if offers != 2 {
		t.Fatalf("%d offers, want 2 (two of your creatures, none of theirs)", offers)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1 (one paid, one declined)", got)
	}
	_ = rock
}

// A noncreature sacrifice is not "a creature".
func TestJmpB1KelsIgnoresANoncreatureSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Kels, Fight Fixer", "Legendary Creature — Azra Warlock", jmpB1KelsOracle, 4, 3)
	treasure := pushCatalogPermanent(g, me.ID, "Treasure", "Token Artifact — Treasure", "", false)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventSacrifice, Actor: me.ID, CardID: treasure})
	})
	if offers := jmpB1Settle(t, g, me, uuid.Nil, nil); offers != 0 {
		t.Errorf("%d offers for an artifact sacrifice, want 0", offers)
	}
}
