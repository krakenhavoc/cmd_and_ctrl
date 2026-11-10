package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const cmdrB13EshkiOracle = "1b162dd3-3be6-406d-bf86-f7cc9eff098d"

func cmdrB13CastCreature(t *testing.T, g *game.Game, me *game.Player, power int) {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Big", TypeLine: "Creature — Beast", Power: power, Toughness: 1, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
}

func TestCmdrB13EshkiScalesWithTheSpellsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	eshki := b12Push(g, me.ID, "Eshki, Temur's Roar", "Legendary Creature — Human Warrior", cmdrB13EshkiOracle, 2, 2)
	advanceTo(t, g, game.StepPrecombatMain)
	hand := me.Hand.Size()
	life := g.Seats[1].Life
	cmdrB13CastCreature(t, g, me, 3)
	if plusOneCounters(g, eshki) != 1 || me.Hand.Size() != hand || g.Seats[1].Life != life {
		t.Fatalf("power 3: counters=%d hand=%d->%d life=%d", plusOneCounters(g, eshki), hand, me.Hand.Size(), g.Seats[1].Life)
	}
	hand = me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 4)
	if plusOneCounters(g, eshki) != 2 || me.Hand.Size() != hand+1 || g.Seats[1].Life != life {
		t.Fatalf("power 4: counters=%d hand=%d->%d life=%d", plusOneCounters(g, eshki), hand, me.Hand.Size(), g.Seats[1].Life)
	}
	cmdrB13CastCreature(t, g, me, 6)
	// Eshki is 2/2 with three counters: 5 damage to each opponent.
	for _, o := range g.Seats[1:] {
		if o.Life != life-5 {
			t.Errorf("opponent life %d, want %d", o.Life, life-5)
		}
	}
	if g.Seats[0].Life != life {
		t.Errorf("controller life changed")
	}
}

const cmdrB13FelotharOracle = "2b3f791a-d672-46fb-a03a-96e19f2c3121"

func TestCmdrB13FelotharAssignsDamageByToughnessAndIgnoresDefender(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	felothar := b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	bear := pushVanillaCreature(g, me.ID, "Glass Cannon", 4, 1)
	wall := pushVanillaCreature(g, me.ID, "Stout Wall", 0, 3)
	before := opp.Life
	attackWith(t, g, opp.ID, felothar, bear, wall)
	passPriorityAroundTable(t, g)
	if want := before - (5 + 1 + 3); opp.Life != want {
		t.Errorf("opponent at %d, want %d (toughness 5 + 1 + 3)", opp.Life, want)
	}
}

func TestCmdrB13FelotharDoesNotReachOpposingCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	theirs := pushVanillaCreature(g, opp.ID, "Their Cannon", 4, 1)
	advanceToNextSeatsTurn(t, g)
	before := me.Life
	attackWith(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if me.Life != before-4 {
		t.Errorf("an opponent's 4/1 still deals 4: %d -> %d", before, me.Life)
	}
}

func TestCmdrB13FelotharSacrificeDrawsToughnessDiscardsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	felothar := b12Push(g, me.ID, "Felothar the Steadfast", "Legendary Creature — Human Warrior", cmdrB13FelotharOracle, 0, 5)
	fed := pushVanillaCreature(g, me.ID, "Fodder", 3, 2)
	for i := 0; i < 6; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	hand := me.Hand.Size()
	b16Activate(t, g, me.ID, felothar, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fed}})
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d, want %d after drawing two", me.Hand.Size(), hand+2)
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand-1 {
		t.Errorf("hand %d, want %d after discarding three", me.Hand.Size(), hand-1)
	}
	if !me.Graveyard.Contains(fed) {
		t.Error("the sacrificed creature is not in the graveyard")
	}
}

const cmdrB13KotisOracle = "0177b410-b559-491f-b393-ac3ed774653c"

func cmdrB13GraveyardCast(g *game.Game, me *game.Player, id uuid.UUID, pay []uuid.UUID) error {
	return g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGraveyardExile, AltCostIDs: pay,
	})
}

func TestCmdrB13KotisCastsFromGraveyardOncePerTurnAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kotis := b12Push(g, me.ID, "Kotis, Sibsig Champion", "Legendary Creature — Zombie Warrior", cmdrB13KotisOracle, 3, 3)
	toMain(t, g)
	first := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	second := seedGraveyardCard(t, g, "Dead Elk", "Creature — Elk", "")
	fodder := seedGraveyardFodder(t, g, 6)

	// Two other cards are not enough.
	if err := cmdrB13GraveyardCast(g, me, first, fodder[:2]); err == nil {
		t.Fatal("cast with only two other cards exiled was accepted")
	}
	if err := cmdrB13GraveyardCast(g, me, first, fodder[:3]); err != nil {
		t.Fatalf("cast from graveyard: %v", err)
	}
	for _, id := range fodder[:3] {
		if me.Graveyard.Contains(id) {
			t.Error("a paid card is still in the graveyard")
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(first) {
		t.Fatal("the creature did not resolve")
	}
	if got := plusOneCounters(g, kotis); got != 2 {
		t.Errorf("Kotis has %d counters, want 2", got)
	}
	// Once during each of your turns.
	if err := cmdrB13GraveyardCast(g, me, second, fodder[3:]); err == nil {
		t.Error("a second graveyard cast in one turn was accepted")
	}
}

func TestCmdrB13KotisIgnoresCreaturesCastFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	kotis := b12Push(g, me.ID, "Kotis, Sibsig Champion", "Legendary Creature — Zombie Warrior", cmdrB13KotisOracle, 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 2)
	if got := plusOneCounters(g, kotis); got != 0 {
		t.Errorf("Kotis grew to %d from a creature cast from hand", got)
	}
}

const cmdrB13ShikoOracle = "40ed32a6-ad56-48c7-aecf-b4238c34c212"

func TestCmdrB13ShikoCopiesATargetedSecondSpellAndDrawsOtherwise(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Shiko and Narset, Unified", "Legendary Creature — Human Spirit Dragon", cmdrB13ShikoOracle, 4, 4)
	advanceTo(t, g, game.StepPrecombatMain)
	// The first spell is not a second spell: nothing happens.
	hand := me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 1)
	if me.Hand.Size() != hand {
		t.Fatalf("the first spell drew: hand %d -> %d", hand, me.Hand.Size())
	}
	// The second spell targets a player: copied, so the opponent takes 6.
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life > life-3 {
		t.Fatalf("opponent at %d, want at most %d", opp.Life, life-3)
	}
	if opp.Life != life-6 {
		t.Errorf("opponent at %d, want %d (the bolt and its copy)", opp.Life, life-6)
	}
}

func TestCmdrB13ShikoDrawsWhenTheSecondSpellHasNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Shiko and Narset, Unified", "Legendary Creature — Human Spirit Dragon", cmdrB13ShikoOracle, 4, 4)
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 1)
	hand := me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 1)
	// One card left the hand for the cast, one was drawn.
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the cast is net zero, one card drawn)", me.Hand.Size(), hand+1)
	}
}

const cmdrB13ZurgoOracle = "0f93d88c-9d2e-416d-a10b-99483360b1fb"

func cmdrB13ZurgoAttacks(t *testing.T, g *game.Game, me, opp *game.Player) uuid.UUID {
	t.Helper()
	zurgo := b12Push(g, me.ID, "Zurgo Stormrender", "Legendary Creature — Orc Warrior", cmdrB13ZurgoOracle, 3, 3)
	declareAttack(t, g, opp.ID, zurgo)
	passPriorityAroundTable(t, g)
	for c := latestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = latestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me.ID, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Warrior: %v", err)
		}
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Warrior" && c.Controller == me.ID {
			return c.InstanceID
		}
	}
	t.Fatal("mobilize made no Warrior")
	return uuid.Nil
}

func TestCmdrB13ZurgoDrawsWhenAnAttackingTokenLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	warrior := cmdrB13ZurgoAttacks(t, g, me, opp)
	hand, life := me.Hand.Size(), opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(warrior) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || opp.Life != life {
		t.Errorf("hand %d->%d, opponent life %d->%d; want a draw and no life loss", hand, me.Hand.Size(), life, opp.Life)
	}
}

func TestCmdrB13ZurgoDrainsWhenANonattackingTokenLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Zurgo Stormrender", "Legendary Creature — Orc Warrior", cmdrB13ZurgoOracle, 3, 3)
	token := b12Creature(g, me.ID, "Spirit", "Token Creature — Spirit", 1, 1)
	hand, life := me.Hand.Size(), opp.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || opp.Life != life-1 || g.Seats[2].Life != life-1 {
		t.Errorf("hand %d->%d, opponent life %d->%d; want each opponent to lose 1", hand, me.Hand.Size(), life, opp.Life)
	}
	if me.Life != life {
		t.Errorf("controller life changed")
	}
}

const cmdrB13ReunionOracle = "ad5ff61f-79d6-4dd9-b4fe-17f7a0055c44"

func cmdrB13GraveCreature(g *game.Game, me *game.Player, name string, power int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Power: power, Toughness: 1, Owner: me.ID, Controller: me.ID})
	})
	return id
}

func TestCmdrB13ReunionReturnsCreaturesUpToTenPowerAndExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	a := cmdrB13GraveCreature(g, me, "Six", 6)
	b := cmdrB13GraveCreature(g, me, "Four", 4)
	zero := cmdrB13GraveCreature(g, me, "Zero", 0)
	c := cmdrB13GraveCreature(g, me, "Extra", 1)
	pick := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, id := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
		}
		return out
	}
	if err := castCatalogSpellErr(t, g, "Reunion of the House", "Sorcery", cmdrB13ReunionOracle, pick(a, b, c)); err == nil {
		t.Fatal("targets with total power 11 were accepted")
	}
	id := castCatalogSpell(t, g, "Reunion of the House", "Sorcery", cmdrB13ReunionOracle, pick(a, b, zero))
	passPriorityAroundTable(t, g)
	for _, want := range []uuid.UUID{a, b, zero} {
		if !g.Battlefield.Contains(want) {
			t.Errorf("a target did not return")
		}
	}
	if !me.Graveyard.Contains(c) || g.Battlefield.Contains(c) {
		t.Error("an untargeted creature card moved")
	}
	if me.Graveyard.Contains(id) {
		t.Error("Reunion of the House went to the graveyard instead of exile")
	}
}

const cmdrB13TempestOracle = "e86bd848-d712-47c5-a82b-8fd5ce1de76a"

func TestCmdrB13TempestTechniqueCountsEnchantmentsAndOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	mine := auraBear(g, me.ID)
	theirs := auraBear(g, opp.ID)
	b12Permanent(g, me.ID, "Pacifism Shrine", "Enchantment")
	b12Permanent(g, opp.ID, "Their Shrine", "Enchantment")
	if err := castCatalogSpellErr(t, g, "Tempest Technique", auraTypeLine, cmdrB13TempestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("enchanted a creature the caster doesn't control")
	}
	auraCast(t, g, "Tempest Technique", cmdrB13TempestOracle, mine)
	// The shrine and the Aura: +2/+2. An opponent's enchantment doesn't count.
	if p, tt := effectivePower(t, g, mine), effectiveToughness(t, g, mine); p != 4 || tt != 4 {
		t.Errorf("enchanted creature is %d/%d, want 4/4", p, tt)
	}
}

func TestCmdrB13TempestTechniqueStormCopiesEnchantTheirOwnCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	first := auraBear(g, me.ID)
	second := auraBear(g, me.ID)
	stormFiller(t, g)
	castCatalogSpell(t, g, "Tempest Technique", auraTypeLine, cmdrB13TempestOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: first}})
	for i := 0; i < 64 && !(stackFullyEmpty(g) && latestPickTarget(g, me.ID) == nil); i++ {
		if p := latestPickTarget(g, me.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: second}); err != nil {
				t.Fatalf("re-target the copy: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	// Two Auras on the table (the card and its token copy): +2/+2 each.
	for name, id := range map[string]uuid.UUID{"original target": first, "copy's target": second} {
		if p := effectivePower(t, g, id); p != 4 {
			t.Errorf("%s has power %d, want 4", name, p)
		}
	}
}

const cmdrB13FrostkiteOracle = "96d7b761-b0b2-4ba3-98bd-3a9191ba544e"

func cmdrB13CastFrostkite(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Deceptive Frostkite", TypeLine: "Creature — Dragon", ManaCost: "{U}{U}",
		OracleID: cmdrB13FrostkiteOracle, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	return id
}

func TestCmdrB13FrostkiteCopiesAYourBigCreatureAndIsAFlyingDragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ogre := seedCopyableCreature(g, me.ID, "Ogre", "Creature — Ogre", 5, 5)
	seedCopyableCreature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	theirs := seedCopyableCreature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	kite := cmdrB13CastFrostkite(t, g, me)
	c := copyPrompt(g)
	if c == nil {
		t.Fatal("no copy prompt")
	}
	_ = theirs
	if err := g.ResolveCopyTarget(c.ID, c.Chooser, ogre); err != nil {
		t.Fatalf("copy the Ogre: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePower(t, g, kite), effectiveToughness(t, g, kite); p != 5 || tt != 5 {
		t.Errorf("copy is %d/%d, want 5/5", p, tt)
	}
	subs := effectiveSubtypes(t, g, kite)
	if !containsString(subs, "Ogre") || !containsString(subs, "Dragon") {
		t.Errorf("subtypes %v, want Ogre and Dragon", subs)
	}
	if !hasEffectiveKeyword(t, g, kite, "flying") {
		t.Error("the copy lacks flying")
	}
}

func TestCmdrB13FrostkiteIsOnlyOfferedBigCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	seedCopyableCreature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	seedCopyableCreature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	kite := cmdrB13CastFrostkite(t, g, me)
	if c := copyPrompt(g); c != nil {
		if err := g.ResolveCopyTarget(c.ID, c.Chooser, uuid.Nil); err != nil {
			t.Fatalf("decline: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if p, tt := effectivePower(t, g, kite), effectiveToughness(t, g, kite); p != 1 || tt != 1 {
		t.Errorf("declined Frostkite is %d/%d, want 1/1", p, tt)
	}
	if !hasEffectiveKeyword(t, g, kite, "flying") {
		t.Error("Frostkite lost its printed flying")
	}
}

// An illegal pick is treated as declining: an opponent's creature is not
// a creature you control.
func TestCmdrB13FrostkiteCannotCopyAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	theirs := seedCopyableCreature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
	kite := cmdrB13CastFrostkite(t, g, me)
	if c := copyPrompt(g); c != nil {
		_ = g.ResolveCopyTarget(c.ID, c.Chooser, theirs)
		passPriorityAroundTable(t, g)
	}
	if p := effectivePower(t, g, kite); p != 1 {
		t.Errorf("Frostkite has power %d, want 1: it copied an opponent's creature", p)
	}
}
