package effects

import (
	"errors"
	"slices"
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

const cmdrB13TranscendentOracle = "e83a629e-2d74-48e2-ad4d-f390067cc51a"

// cmdrB13CastDragonOverBolt has the active seat cast a Bolt at the opponent
// and then flash in the Dragon over it, targeting the Bolt with the ETB.
func cmdrB13CastDragonOverBolt(t *testing.T, g *game.Game) (bolt uuid.UUID, me *game.Player) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	bolt = castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	castFromSeat(t, g, me, "Transcendent Dragon", "Creature — Dragon", cmdrB13TranscendentOracle, nil)
	if err := g.PassPriority(); err != nil && !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority: %v", err)
	}
	for i := 0; i < 8 && latestPickTarget(g, me.ID) == nil && resolutionCastOfferFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt}); err != nil {
			t.Fatalf("target the Bolt: %v", err)
		}
	}
	for i := 0; i < 8 && resolutionCastOfferFor(g, me.ID) == nil && !stackFullyEmpty(g); i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	return bolt, me
}

func TestCmdrB13TranscendentDragonCountersExilesAndCastsFree(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	bolt, me := cmdrB13CastDragonOverBolt(t, g)
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordExileUntil, bolt)
	if err := castThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, Strict: true,
	}); err != nil {
		t.Fatalf("free cast of the countered Bolt: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d: the exiled Bolt was cast for free", opp.Life, life-3)
	}
	if me.Graveyard.Contains(bolt) && g.Exile.Contains(bolt) {
		t.Error("the Bolt is in two zones")
	}
}

func TestCmdrB13TranscendentDragonDeclinedSpellStaysExiled(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := opp.Life
	bolt, me := cmdrB13CastDragonOverBolt(t, g)
	declineResolutionCast(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(bolt) {
		t.Error("a declined countered spell should stay in exile")
	}
	if me.Graveyard.Contains(bolt) {
		t.Error("the countered spell went to the graveyard instead of exile")
	}
	if opp.Life != life {
		t.Errorf("opponent at %d, want %d", opp.Life, life)
	}
}

const cmdrB13BibliophileOracle = "dcc323d6-ddae-4fd9-9c9e-466b07895d73"

func TestCmdrB13BibliophileDrawsForEachTargetOfASpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Voracious Bibliophile", "Creature — Dragon", cmdrB13BibliophileOracle, 3, 3)
	advanceTo(t, g, game.StepPrecombatMain)
	// A creature spell has no target: nothing is drawn.
	hand := me.Hand.Size()
	cmdrB13CastCreature(t, g, me, 1)
	if me.Hand.Size() != hand {
		t.Fatalf("an untargeted spell drew: hand %d -> %d", hand, me.Hand.Size())
	}
	// A Bolt has one target, and the draw happens even though the Bolt
	// resolves afterwards.
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	handBeforeResolve := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - handBeforeResolve; got != 1 {
		t.Errorf("drew %d cards for a one-target spell, want 1", got)
	}
}

const cmdrB14AfterlifeOracle = "151c867f-cc77-42de-9ae5-5fd2233ccdab"

func TestCmdrB14AfterlifeTakesOneCreatureFromEachGraveyardAsZombies(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	mine := cmdrB13GraveCreature(g, me, "My Bear", 2)
	spare := cmdrB13GraveCreature(g, me, "My Spare", 2)
	theirs := cmdrB13GraveCreature(g, opp, "Their Elk", 3)
	thirds := cmdrB13GraveCreature(g, third, "Third Wolf", 1)
	pick := []game.TargetRef{
		{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: thirds},
	}
	// Two cards from one graveyard is more than "up to one" per player.
	if err := castCatalogSpellErr(t, g, "Afterlife from the Loam", "Sorcery", cmdrB14AfterlifeOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: spare}}); err == nil {
		t.Fatal("two targets from the same graveyard were accepted")
	}
	castCatalogSpell(t, g, "Afterlife from the Loam", "Sorcery", cmdrB14AfterlifeOracle, pick)
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, theirs, thirds} {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("a target did not reach the battlefield")
		}
		if c.Controller != me.ID {
			t.Errorf("%s is controlled by %s, want the caster", c.Name, c.Controller)
		}
		if !containsString(effectiveSubtypes(t, g, id), "Zombie") {
			t.Errorf("%s is not a Zombie: %v", c.Name, effectiveSubtypes(t, g, id))
		}
	}
	if !me.Graveyard.Contains(spare) {
		t.Error("the untargeted card left the graveyard")
	}
}

const cmdrB14BoneDevourerOracle = "0ee41ef9-0b39-4d00-a68e-7914839a7175"

func TestCmdrB14BoneDevourerGrowsByDeathsAndPaysOutWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, name := range []string{"Bear One", "Bear Two"} {
		id := b12Creature(g, me.ID, name, "Creature — Bear", 2, 2)
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
		passPriorityAroundTable(t, g)
	}
	devourer := castHoldCreature(t, g, "Bone Devourer", "Creature — Dragon", cmdrB14BoneDevourerOracle, 2, 2)
	if got := plusOnes(g, devourer); got != 2 {
		t.Fatalf("%d +1/+1 counters after two deaths, want 2", got)
	}
	for i := 0; i < 6; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(devourer) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("hand %d->%d, life %d->%d; want to draw 2 and lose 2", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestCmdrB14BoneDevourerWithNoDeathsDrawsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devourer := castHoldCreature(t, g, "Bone Devourer", "Creature — Dragon", cmdrB14BoneDevourerOracle, 2, 2)
	hand, life := me.Hand.Size(), me.Life
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(devourer) })
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d->%d, life %d->%d; want no change", hand, me.Hand.Size(), life, me.Life)
	}
}

const cmdrB14TipOracle = "f7cac691-960f-4e36-8eae-f24bcdc1269b"

func TestCmdrB14TipTheScalesShrinksEveryCreatureByTheSacrificedToughness(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 3)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 5, 5)
	small := b12Creature(g, opp.ID, "Small", "Creature — Elf", 2, 3)
	big := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 6, 6)
	castCatalogSpell(t, g, "Tip the Scales", "Sorcery", cmdrB14TipOracle, nil)
	passPriorityAroundTable(t, g)
	// Sacrifice the Wall (toughness 3).
	answerOwnPermanents(t, g, me.ID, wall)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wall) {
		t.Error("the Wall was not sacrificed")
	}
	if g.Battlefield.Contains(small) {
		t.Error("the 2/3 survived -3/-3")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(big) {
		t.Fatal("a big creature died")
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my 5/5 has power %d, want 2", p)
	}
	if p := effectivePower(t, g, big); p != 3 {
		t.Errorf("their 6/6 has power %d, want 3", p)
	}
}

const cmdrB14WithinRangeOracle = "cc8c1e21-4540-4e74-a513-eb6c17372de1"

func TestCmdrB14WithinRangeMakesWarriorsAndDrainsByAttackers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	castCatalogSpell(t, g, "Within Range", "Enchantment", cmdrB14WithinRangeOracle, nil)
	passPriorityAroundTable(t, g)
	if n := b16CountNamed(g, "Warrior"); n != 2 {
		t.Fatalf("%d Warriors, want 2", n)
	}
	a := pushVanillaCreature(g, me.ID, "A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "B", 1, 1)
	c := pushVanillaCreature(g, me.ID, "C", 1, 1)
	lifeOpp, lifeThird := opp.Life, third.Life
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(c, third.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	// Settle the combat damage too: the trigger drains 2 from the first
	// opponent and 1 from the third, then the unblocked 1/1s deal 2 and 1.
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if got := lifeOpp - opp.Life; got != 4 {
		t.Errorf("first opponent lost %d, want 4 (2 drained, 2 combat)", got)
	}
	if got := lifeThird - third.Life; got != 2 {
		t.Errorf("third opponent lost %d, want 2 (1 drained, 1 combat)", got)
	}
	if g.Seats[3].Life != 40 || me.Life != 40 {
		t.Errorf("controller life %d, fourth player %d; want 40 and 40", me.Life, g.Seats[3].Life)
	}
}

const cmdrB14InfantryShieldOracle = "71c68002-1489-4858-a8fb-1f575641983e"

func TestCmdrB14InfantryShieldGivesMenaceAndMobilizeByPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	bear := pushVanillaCreature(g, me.ID, "Bear", 3, 3)
	other := pushVanillaCreature(g, me.ID, "Other", 2, 2)
	shield := pushCatalogPermanent(g, me.ID, "Infantry Shield", "Artifact — Equipment", cmdrB14InfantryShieldOracle, false)
	equipTo(t, g, me.ID, shield, bear)
	if !hasEffectiveKeyword(t, g, bear, "menace") {
		t.Error("the equipped creature lacks menace")
	}
	if hasEffectiveKeyword(t, g, other, "menace") {
		t.Error("an unequipped creature has menace")
	}
	declareAttack(t, g, opp.ID, bear, other)
	passPriorityAroundTable(t, g)
	for c := latestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = latestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me.ID, c.PayAmount.Max); err != nil {
			t.Fatalf("place the Warriors: %v", err)
		}
	}
	if n := b16CountNamed(g, "Warrior"); n != 3 {
		t.Errorf("%d Warriors, want 3 (mobilize X for the equipped 3-power creature only)", n)
	}
}

const cmdrB14CanopyOracle = "ec1e69b1-f664-494b-8da2-c7199872b08b"

func TestCmdrB14CanopyGargantuanCountersEqualToughnessOnOthers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	canopy := b12Push(g, me.ID, "Canopy Gargantuan", "Creature — Dragon", cmdrB14CanopyOracle, 7, 7)
	wall := b12Creature(g, me.ID, "Wall", "Creature — Wall", 0, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Elf", 1, 1)
	advanceToUpkeepOf(t, g, g.Turn.ActiveSeat)
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, wall); got != 4 {
		t.Errorf("Wall has %d counters, want 4", got)
	}
	if got := plusOneCounters(g, bear); got != 2 {
		t.Errorf("Bear has %d counters, want 2", got)
	}
	if got := plusOneCounters(g, canopy); got != 0 {
		t.Errorf("Canopy put %d counters on itself", got)
	}
	if got := plusOneCounters(g, theirs); got != 0 {
		t.Errorf("an opponent's creature got %d counters", got)
	}
}

const cmdrB14FloralOracle = "4da17e9e-df3e-4c46-a0cf-baafbe67186f"

func TestCmdrB14FloralEvokerLandfallAndRecursion(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	evoker := b12Push(g, me.ID, "Floral Evoker", "Creature — Snake Druid", cmdrB14FloralOracle, 2, 3)
	b12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := plusOneCounters(g, evoker); got != 1 {
		t.Fatalf("%d counters after a land entered, want 1", got)
	}
	// Discard a creature card to return a land card from the graveyard
	// tapped; that entry is landfall too.
	deadLand := seedGraveyardCard(t, g, "Dead Island", "Basic Land — Island", "")
	fodder := handCardForTest(me, "Fodder", "Creature — Bear", "")
	b06AddMana(me, "G")
	b16Activate(t, g, me.ID, evoker, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{fodder},
		Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: deadLand}},
	})
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(deadLand) || !b13Tapped(t, g, deadLand) {
		t.Error("the land did not return tapped")
	}
	if !me.Graveyard.Contains(fodder) {
		t.Error("the discarded creature card is not in the graveyard")
	}
	if got := plusOneCounters(g, evoker); got != 2 {
		t.Errorf("%d counters after the returned land entered, want 2", got)
	}
	// A noncreature card can't pay the cost.
	spell := handCardForTest(me, "Bolt", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, evoker, 0, game.ActivateAbilityParams{
		DiscardIDs: []uuid.UUID{spell}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: deadLand}},
	}); err == nil {
		t.Error("an Instant paid for a creature-card discard")
	}
}

const cmdrB14AlignedHeartOracle = "ae61b2ef-2e60-499a-8708-e37b9bd5620e"

func TestCmdrB14AlignedHeartMakesMoreMonksEachFlurry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	heart := b12Permanent(g, me.ID, "Aligned Heart", "Enchantment")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == heart {
				g.Battlefield.Cards[i].OracleID = cmdrB14AlignedHeartOracle
			}
		}
	})
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 1)
	if n := b16CountNamed(g, "Monk"); n != 0 {
		t.Fatalf("%d Monks after the first spell, want 0", n)
	}
	cmdrB13CastCreature(t, g, me, 1)
	if n := b16CountNamed(g, "Monk"); n != 1 {
		t.Errorf("%d Monks after the second spell, want 1", n)
	}
	if got := countersOn(g, heart, "rally"); got != 1 {
		t.Errorf("%d rally counters, want 1", got)
	}
	// A third spell is not a second spell.
	cmdrB13CastCreature(t, g, me, 1)
	if n := b16CountNamed(g, "Monk"); n != 1 {
		t.Errorf("%d Monks after the third spell, want 1", n)
	}
}

const cmdrB14IronwillOracle = "eee99fdc-b116-4f06-9f9f-7b7dda5743c1"

func cmdrB14IronwillMyriadCopies(t *testing.T, withCommander bool) int {
	t.Helper()
	g := newCatalogGame(t)
	me, opp1 := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Ironwill Forger", "Creature — Orc Artificer", cmdrB14IronwillOracle, 3, 3)
	if withCommander {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
			Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true,
		})
	}
	knight := pushVanillaCreature(g, me.ID, "Knight", 3, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	for i := 0; i < 4 && latestPickTarget(g, me.ID) != nil; i++ {
		pickCard(t, g, me.ID, knight)
		passPriorityAroundTable(t, g)
	}
	declareAttack(t, g, opp1.ID, knight)
	passPriorityAroundTable(t, g)
	for i := 0; i < 3 && len(g.PendingChoices) > 0; i++ {
		answerMayChoice(t, g, me.ID, true)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && c.InstanceID != knight && IsToken(c) {
			n++
		}
	}
	return n
}

func TestCmdrB14IronwillForgerGivesMyriadOnlyWithYourCommander(t *testing.T) {
	if got := cmdrB14IronwillMyriadCopies(t, true); got != 2 {
		t.Errorf("with a commander: %d myriad copies, want 2 (one per other opponent)", got)
	}
	if got := cmdrB14IronwillMyriadCopies(t, false); got != 0 {
		t.Errorf("without a commander: %d myriad copies, want 0", got)
	}
}

const cmdrB14ProtectorOracle = "4534e260-06fd-4054-8077-505c2f6c5767"

func TestCmdrB14ProtectorOfTheWastesExilesOnePerControllerAndAgainWhenMonstrous(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[g.Turn.ActiveSeat], g.Seats[1], g.Seats[2]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	rock2 := b12Permanent(g, opp.ID, "Their Other Rock", "Artifact")
	aura := b12Permanent(g, third.ID, "Their Shrine", "Enchantment")
	mine := b12Permanent(g, me.ID, "My Shrine", "Enchantment")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Protector of the Wastes", TypeLine: "Creature — Dragon",
		OracleID: cmdrB14ProtectorOracle, Power: 5, Toughness: 5, Owner: me.ID, Controller: me.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the enters trigger")
	}
	// Two permanents of one player are refused.
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: rock2},
	}); err == nil {
		t.Fatal("two targets with the same controller were accepted")
	}
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the prompt went away after a refused answer")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{
		{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetCard, ID: aura},
	}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) || g.Battlefield.Contains(aura) {
		t.Error("a target was not exiled")
	}
	if !g.Battlefield.Contains(rock2) || !g.Battlefield.Contains(mine) {
		t.Error("an untargeted permanent was exiled")
	}
	// Becoming monstrous triggers it again.
	b06AddMana(me, "W", "C", "C", "C", "C")
	activateMonstrosity(t, g, me.ID, id, 0)
	passPriorityAroundTable(t, g)
	p = latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt for the monstrous trigger")
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: rock2}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock2) {
		t.Error("the monstrous trigger did not exile its target")
	}
	if plusOneCounters(g, id) != 3 {
		t.Errorf("%d +1/+1 counters after monstrosity 3, want 3", plusOneCounters(g, id))
	}
}

const cmdrB15PostOracle = "a6657fcf-f08c-4b03-8ec8-cb0b194eb553"

func TestCmdrB15AdaptiveTrainingPostChargesToThreeThenCopiesTheNextSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	post := pushCatalogPermanent(g, me.ID, "Adaptive Training Post", "Artifact", cmdrB15PostOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	for i := 1; i <= 4; i++ {
		castCatalogSpell(t, g, "Filler", "Instant", "", nil)
		passPriorityAroundTable(t, g)
		want := i
		if want > 3 {
			want = 3
		}
		if got := countersOn(g, post, "charge"); got != want {
			t.Fatalf("after %d spells: %d charge counters, want %d", i, got, want)
		}
	}
	// A creature spell doesn't charge it.
	cmdrB13CastCreature(t, g, me, 1)
	if got := countersOn(g, post, "charge"); got != 3 {
		t.Fatalf("a creature spell changed the count to %d", got)
	}
	b16Activate(t, g, me.ID, post, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if got := countersOn(g, post, "charge"); got != 0 {
		t.Fatalf("%d charge counters left after paying three", got)
	}
	life := opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if c := latestChoiceOfKind(g, game.PendingChoiceTriggerOrder); c != nil {
		if err := g.ResolveTriggerOrder(c.ID, c.Chooser, c.TriggerOrderIDs); err != nil {
			t.Fatalf("ResolveTriggerOrder: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	b16PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if opp.Life != life-6 {
		t.Errorf("opponent at %d, want %d: the Bolt and its copy", opp.Life, life-6)
	}
	// Only the next spell: a second Bolt isn't copied.
	life = opp.Life
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent at %d, want %d: a second Bolt must not be copied", opp.Life, life-3)
	}
}

const cmdrB15DivinerOracle = "abf4dcd8-5176-4f4a-b7c1-5ce24ff181bf"

func cmdrB15DivinerAttack(t *testing.T) (g *game.Game, me *game.Player, bolt, big uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[0]
	diviner := b12Push(g, me.ID, "Diviner of Mist", "Creature — Dragon", cmdrB15DivinerOracle, 4, 5)
	bolt = uuid.New()
	big = uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
		me.Graveyard.PushTop(game.Card{InstanceID: big, Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{4}{R}", Owner: me.ID, Controller: me.ID})
	})
	declareAttack(t, g, g.Seats[1].ID, diviner)
	passPriorityAroundTable(t, g)
	return g, me, bolt, big
}

func TestCmdrB15DivinerMillsFourAndCastsAnInstantFreeThenExilesIt(t *testing.T) {
	g, me, bolt, big := cmdrB15DivinerAttack(t)
	opp := g.Seats[1]
	c := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, bolt)
	if slices.Contains(c.DiscardOptions, big) {
		t.Error("a mana value 5 sorcery was offered")
	}
	life := opp.Life
	if err := castThroughResolutionCast(t, g, me.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{playerTarget(opp.ID)}, Strict: true, AlternativeCost: game.AltCostKeyGrantedCost,
	}); err != nil {
		t.Fatalf("free cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life > life-3 {
		t.Errorf("opponent at %d, want at most %d", opp.Life, life-3)
	}
	if slices.Contains(c.DiscardOptions, big) || me.Graveyard.Contains(bolt) {
		t.Error("the Bolt returned to the graveyard instead of being exiled")
	}
	if !g.Exile.Contains(bolt) {
		t.Error("the Bolt is not in exile")
	}
}

func TestCmdrB15DivinerMillsFourCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	diviner := b12Push(g, me.ID, "Diviner of Mist", "Creature — Dragon", cmdrB15DivinerOracle, 4, 5)
	before := me.Library.Size()
	declareAttack(t, g, g.Seats[1].ID, diviner)
	passPriorityAroundTable(t, g)
	if got := before - me.Library.Size(); got != 4 {
		t.Errorf("milled %d, want 4", got)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("an offer was made with no instant or sorcery in the graveyard")
	}
}

const cmdrB15HammerheadOracle = "7830f600-c7c7-48f8-8020-3493c5745eed"

func TestCmdrB15HammerheadTyrantBouncesOnlyWhatTheSpellCanMatch(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	b12Push(g, me.ID, "Hammerhead Tyrant", "Creature — Dragon", cmdrB15HammerheadOracle, 6, 6)
	cheap := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Creature — Elf", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	dear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Creature — Ogre", ManaCost: "{3}{R}", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Opp Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	advanceTo(t, g, game.StepPrecombatMain)
	// A mana value 1 spell: the 1-drop may be returned, the 4-drop and the
	// land may not be.
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "One Drop", TypeLine: "Instant", ManaCost: "{R}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no target prompt")
	}
	if !slices.Contains(p.PickTargetCards, cheap) || slices.Contains(p.PickTargetCards, dear) || slices.Contains(p.PickTargetCards, land) {
		t.Fatalf("candidates %v: want only the 1-drop", p.PickTargetCards)
	}
	if err := g.ResolvePickTargets(p.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: cheap}}); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cheap) || !opp.Hand.Contains(cheap) {
		t.Error("the 1-drop was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(dear) {
		t.Error("the 4-drop was returned")
	}
}

const cmdrB15TemurOracle = "ebe1448e-21c5-4b73-8296-03d3e5097c50"

func TestCmdrB15WillOfTheTemurCopyIsAFourFourFlyingDragon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Rock", TypeLine: "Artifact", ManaCost: "{6}", Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Will of the Temur", "Sorcery", cmdrB15TemurOracle, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: rock}})
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Big Rock" && c.InstanceID != rock {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy")
	}
	if p, tt := effectivePower(t, g, token), effectiveToughness(t, g, token); p != 4 || tt != 4 {
		t.Errorf("copy is %d/%d, want 4/4", p, tt)
	}
	if !hasEffectiveKeyword(t, g, token, "flying") {
		t.Error("the copy lacks flying")
	}
	types := effectiveTypes(t, g, token)
	if !slices.Contains(types, "Creature") || !slices.Contains(types, "Artifact") {
		t.Errorf("types %v, want Artifact and Creature", types)
	}
	if !slices.Contains(effectiveSubtypes(t, g, token), "Dragon") {
		t.Errorf("subtypes %v, want Dragon", effectiveSubtypes(t, g, token))
	}
}

func TestCmdrB15WillOfTheTemurDrawsGreatestManaValueAndBothWithACommander(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Five", TypeLine: "Artifact", ManaCost: "{5}", Owner: me.ID, Controller: me.ID})
	for i := 0; i < 12; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	both := []game.TargetRef{{Kind: game.TargetCard, ID: rock}, {Kind: game.TargetPlayer, ID: me.ID}}
	if err := b22TryModal(g, cmdrB15TemurOracle, []int{0, 1}, both); err == nil {
		t.Fatal("no commander: choosing both modes was accepted")
	}
	// Only the draw: five cards for the Five.
	hand := me.Hand.Size()
	castModal(t, g, "Will of the Temur", "Sorcery", cmdrB15TemurOracle, []int{1}, []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 5 {
		t.Fatalf("drew %d, want 5 (the greatest mana value is 5)", got)
	}
	// With a commander on the battlefield, both: the copy of the Five and then
	// six cards for the six-drop commander.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		ManaCost: "{4}{G}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
	advanceToPrecombatMainOf(t, g, 0)
	hand = me.Hand.Size()
	castModal(t, g, "Will of the Temur", "Sorcery", cmdrB15TemurOracle, []int{0, 1}, both)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 6 {
		t.Errorf("drew %d, want 6", got)
	}
	if n := b16CountNamed(g, "Five"); n != 2 {
		t.Errorf("%d permanents named Five, want the original and the token copy", n)
	}
}

const cmdrB15JawsOracle = "46327e52-b04a-4155-bda6-9559e46fb9ed"

func TestCmdrB15JawsOfDefeatDrainsByThePowerToughnessDifference(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Jaws of Defeat", "Enchantment", cmdrB15JawsOracle, false)
	advanceTo(t, g, game.StepPrecombatMain)
	cast := func(power, toughness int) {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Body", TypeLine: "Creature — Wall", Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		for i := 0; i < 3 && latestPickTarget(g, me.ID) != nil; i++ {
			b16PickPlayer(t, g, me.ID, opp.ID)
			passPriorityAroundTable(t, g)
		}
	}
	life := opp.Life
	cast(1, 4) // 3
	if opp.Life != life-3 {
		t.Fatalf("a 1/4 drained %d, want 3", life-opp.Life)
	}
	cast(5, 2) // 3
	if opp.Life != life-6 {
		t.Fatalf("a 5/2 drained %d in total, want 6", life-opp.Life)
	}
	cast(3, 3) // 0
	if opp.Life != life-6 {
		t.Errorf("a 3/3 changed the life total: %d", life-opp.Life)
	}
}

const cmdrB15WelcomeOracle = "affa0b6c-a9fd-49fc-b9ce-9a3486af8320"

func TestCmdrB15WelcomeTheDeadLootsPaysLifeAndCountsGraveyardEntries(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 12; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	advanceTo(t, g, game.StepPrecombatMain)
	// Two cards milled earlier this turn count.
	g.WithWriteLock(func() { _ = g.MillNForEffect(me.ID, 2) })
	life := me.Life
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Welcome the Dead", "Sorcery", cmdrB15WelcomeOracle, nil)
	passPriorityAroundTable(t, g)
	discardFromHand(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("hand grew by %d, want 1 (drew two, discarded one; the spell's own card was cast from hand)", got)
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie Druid" && c.Controller == me.ID {
			n++
			if !c.Tapped {
				t.Error("a Zombie Druid entered untapped")
			}
		}
	}
	if n != 3 {
		t.Errorf("%d Zombie Druids, want 3 (two milled and one discarded)", n)
	}
}

const cmdrB15GoldlustOracle = "86d76078-9af3-45a7-90d6-72946afdc78b"

func TestCmdrB15GoldlustTriadMyriadAndTreasuresFromEachHit(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	triad := b12Push(g, me.ID, "Goldlust Triad", "Creature — Dragon", cmdrB15GoldlustOracle, 4, 3)
	declareAttack(t, g, opp1.ID, triad)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	answerMayChoice(t, g, me.ID, true)
	copies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goldlust Triad" && c.InstanceID != triad && IsToken(c) {
			copies++
		}
	}
	if copies != 2 {
		t.Fatalf("%d myriad copies, want 2", copies)
	}
	l1, l2, l3 := opp1.Life, opp2.Life, opp3.Life
	advanceTo(t, g, game.StepEndCombat)
	passPriorityAroundTable(t, g)
	if opp1.Life != l1-4 || opp2.Life != l2-4 || opp3.Life != l3-4 {
		t.Errorf("each opponent takes 4: %d %d %d", l1-opp1.Life, l2-opp2.Life, l3-opp3.Life)
	}
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
		}
	}
	if treasures != 3 {
		t.Errorf("%d Treasures, want 3 (the Triad and each copy hit a player)", treasures)
	}
}

const cmdrB15ThundermaneOracle = "a1602bd0-f346-43cf-86ec-1dd300a7f2ed"

func TestCmdrB15ThundermaneDragonCastsBigCreaturesFromTheTopWithHaste(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Thundermane Dragon", "Creature — Dragon", cmdrB15ThundermaneOracle, 4, 4)
	advanceToMain(t, g)
	small := pushLibraryCard(me, game.Card{Name: "Small Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 3, Toughness: 3})
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.CastSpell(me.ID, small, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a power 3 creature was cast from the top")
	}
	spell := pushLibraryCard(me, game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{1}"})
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Fatal("a sorcery was cast from the top")
	}
	big := pushLibraryCard(me, game.Card{Name: "Big Ogre", TypeLine: "Creature — Ogre", ManaCost: "{1}", Power: 4, Toughness: 4})
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{FromZone: "library", Strict: true}); err != nil {
		t.Fatalf("cast the 4-power creature from the top: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(big) {
		t.Fatal("the creature did not resolve")
	}
	if !hasEffectiveKeyword(t, g, big, "haste") {
		t.Error("the creature cast this way has no haste")
	}
	// Until end of turn: next turn it is no longer granted haste.
	advanceToNextSeatsTurn(t, g)
	advanceToUpkeepOf(t, g, 0)
	if hasEffectiveKeyword(t, g, big, "haste") {
		t.Error("haste lasted past the turn it was cast")
	}
}

func TestCmdrB15ThundermaneDragonDoesNotHelpOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Thundermane Dragon", "Creature — Dragon", cmdrB15ThundermaneOracle, 4, 4)
	advanceToMain(t, g)
	theirs := pushLibraryCard(opp, game.Card{Name: "Their Ogre", TypeLine: "Creature — Ogre", ManaCost: "{1}", Power: 4, Toughness: 4})
	floatMana(t, g, opp, "{C}")
	if err := g.CastSpell(opp.ID, theirs, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("an opponent cast from the top of their library")
	}
}

const cmdrB15FlourishOracle = "9fb7e7ca-ad2f-4677-bb4d-50808b01b4f9"

func TestCmdrB15TransformingFlourishDestroysThenGivesTheOwnerAFreeCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	// Their library: a land, then a Bolt on top of the land.
	opp.Library.Cards = nil
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	bolt := uuid.New()
	opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Top Land", TypeLine: "Basic Land — Island", Owner: opp.ID, Controller: opp.ID})
	opp.Library.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle, Owner: opp.ID, Controller: opp.ID})
	// Demonstrate asks first; decline it.
	castCatalogSpell(t, g, "Transforming Flourish", "Instant", cmdrB15FlourishOracle, b16TargetCard(ogre))
	answerLatestTriggerPrompt(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Fatal("the Ogre survived")
	}
	// The Bolt was the first nonland card from the top.
	mustResolutionCastOffer(t, g, opp.ID, game.MayCastKeywordFreeCast, bolt)
	life := me.Life
	if err := castThroughResolutionCast(t, g, opp.ID, bolt, game.CastSpellParams{
		Targets: []game.TargetRef{playerTarget(me.ID)},
	}); err != nil {
		t.Fatalf("free cast by the destroyed permanent's controller: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("caster at %d, want %d", me.Life, life-3)
	}
}

const cmdrB15JeskaiOracle = "2c3e6377-d09c-46b6-8a83-ada4a2930b14"

func TestCmdrB15WillOfTheJeskaiWheelsForFiveAndGrantsFlashback(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 20; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	deadBolt := seedGraveyardCard(t, g, "Lightning Bolt", "Instant", lightningBoltOracle)
	creature := seedGraveyardCard(t, g, "Dead Bear", "Creature — Bear", "")
	if err := b22TryModal(g, cmdrB15JeskaiOracle, []int{0, 1}, nil); err == nil {
		t.Fatal("no commander: choosing both modes was accepted")
	}
	castModal(t, g, "Will of the Jeskai", "Sorcery", cmdrB15JeskaiOracle, []int{1}, nil)
	passPriorityAroundTable(t, g)
	perm := grantedPermissionOn(g, me.ID, deadBolt, game.ZoneGraveyard)
	if !perm.Granted() || perm.AltCostKey != "flashback" {
		t.Fatalf("the Bolt gained no flashback: %+v", perm)
	}
	if p := grantedPermissionOn(g, me.ID, creature, game.ZoneGraveyard); p.Granted() {
		t.Error("a creature card gained flashback")
	}
}

func TestCmdrB15WillOfTheJeskaiWheelOnlyForThoseWhoAgree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	for _, p := range g.Seats {
		for i := 0; i < 12; i++ {
			stapleLibraryCard(p, "Spare", "Sorcery")
		}
	}
	oppHand := opp.Hand.Size()
	myCards := append([]game.Card(nil), me.Hand.Cards...)
	castModal(t, g, "Will of the Jeskai", "Sorcery", cmdrB15JeskaiOracle, []int{0}, nil)
	passPriorityAroundTable(t, g)
	// Everyone is asked in turn order: I say yes, the others decline.
	for i := 0; i < 8; i++ {
		c := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if c == nil {
			break
		}
		if err := g.ResolveConfirm(c.ID, c.Chooser, c.Chooser == me.ID); err != nil {
			t.Fatalf("answer: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if me.Hand.Size() != 5 {
		t.Errorf("my hand is %d cards, want 5", me.Hand.Size())
	}
	for _, c := range myCards {
		if me.Hand.Contains(c.InstanceID) {
			t.Errorf("an old card stayed in my hand")
			break
		}
	}
	if opp.Hand.Size() != oppHand {
		t.Errorf("a player who declined changed hands: %d -> %d", oppHand, opp.Hand.Size())
	}
}

func TestCmdrB15WillOfTheJeskaiBothModesGrantsFlashbackToTheDiscardedInstant(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		for i := 0; i < 12; i++ {
			stapleLibraryCard(p, "Spare", "Sorcery")
		}
	}
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Commander", TypeLine: "Legendary Creature — Elf",
		ManaCost: "{2}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID, IsCommander: true})
	bolt := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}", OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
	castModal(t, g, "Will of the Jeskai", "Sorcery", cmdrB15JeskaiOracle, []int{0, 1}, nil)
	passPriorityAroundTable(t, g)
	for i := 0; i < 8; i++ {
		c := latestChoiceOfKind(g, game.PendingChoiceConfirm)
		if c == nil {
			break
		}
		if err := g.ResolveConfirm(c.ID, c.Chooser, c.Chooser == me.ID); err != nil {
			t.Fatalf("answer: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	if !me.Graveyard.Contains(bolt) {
		t.Fatal("the Bolt was not discarded by the wheel")
	}
	if perm := grantedPermissionOn(g, me.ID, bolt, game.ZoneGraveyard); !perm.Granted() {
		t.Error("the discarded Bolt did not gain flashback: the grant ran before the wheel finished")
	}
}

const (
	cmdrB16ShadowSummoningOracle = "b9225046-03aa-4729-89fa-32c049407229"
	cmdrB16GoldnightOracle       = "3dc51dfc-7a60-41bf-b944-6afad6b06c22"
	cmdrB16AvalancheOracle       = "e5fcbbaf-9431-4fa8-ac11-12ace1f2054c"
)

func TestCmdrB16ShadowSummoningMakesTwoTappedFlyingSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Shadow Summoning", "Sorcery", cmdrB16ShadowSummoningOracle, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Spirit" && c.Controller == me.ID {
			n++
			if !c.Tapped {
				t.Error("a Spirit entered untapped")
			}
			if !hasEffectiveKeyword(t, g, c.InstanceID, "flying") {
				t.Error("a Spirit lacks flying")
			}
		}
	}
	if n != 2 {
		t.Errorf("%d Spirits, want 2", n)
	}
}

func TestCmdrB16GoldnightCommanderPumpsOnAnotherCreatureEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	commander := b12Push(g, me.ID, "Goldnight Commander", "Creature — Human Cleric Soldier", cmdrB16GoldnightOracle, 2, 2)
	bear := auraBear(g, me.ID)
	theirs := auraBear(g, g.Seats[1].ID)
	advanceTo(t, g, game.StepPrecombatMain)
	cmdrB13CastCreature(t, g, me, 1)
	for name, id := range map[string]uuid.UUID{"Goldnight Commander": commander, "my Bear": bear} {
		want := map[string]int{"Goldnight Commander": 3, "my Bear": 3}[name]
		if p := effectivePower(t, g, id); p != want {
			t.Errorf("%s has power %d, want %d", name, p, want)
		}
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
}

func TestCmdrB16BecomeTheAvalancheDrawsPerBigCreatureThenPumpsByHandSize(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	big1 := b12Creature(g, me.ID, "Big One", "Creature — Ogre", 4, 4)
	big2 := b12Creature(g, me.ID, "Big Two", "Creature — Ogre", 6, 6)
	small := b12Creature(g, me.ID, "Small", "Creature — Elf", 3, 3)
	for i := 0; i < 10; i++ {
		stapleLibraryCard(me, "Spare", "Sorcery")
	}
	castCatalogSpell(t, g, "Become the Avalanche", "Sorcery", cmdrB16AvalancheOracle, nil)
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("drew %d, want 2 (two creatures with power 4 or greater)", got)
	}
	x := me.Hand.Size()
	if p := effectivePower(t, g, big1); p != 4+x {
		t.Errorf("Big One has power %d, want %d", p, 4+x)
	}
	if p := effectivePower(t, g, small); p != 3+x {
		t.Errorf("Small has power %d, want %d", p, 3+x)
	}
	_ = big2
}

const cmdrB16RampartOracle = "b0edba6f-8f54-4dc4-b952-20b57a797d60"

func TestCmdrB16RampartArchitectMakesWallsAndFetchesWhenADefenderDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := seedSearchLibrary(me,
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Hallowed Fountain", TypeLine: "Land — Plains Island"},
	)
	castCatalogSpell(t, g, "Rampart Architect", "Creature — Elephant Advisor", cmdrB16RampartOracle, nil)
	passPriorityAroundTable(t, g)
	var wall uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Wall" && c.Controller == me.ID {
			wall = c.InstanceID
		}
	}
	if wall == uuid.Nil {
		t.Fatal("no Wall token on entering")
	}
	if p, tt := effectivePower(t, g, wall), effectiveToughness(t, g, wall); p != 1 || tt != 3 {
		t.Errorf("the Wall is %d/%d, want 1/3", p, tt)
	}
	if !hasEffectiveKeyword(t, g, wall, "defender") {
		t.Error("the Wall has no defender")
	}
	// A creature without defender dying does not fetch.
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("a creature without defender dying prompted a search")
	}
	// The Wall dying does.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wall) })
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Wall dying did not offer a search")
	}
	if searchOptionNamed(g, c, "Hallowed Fountain") != uuid.Nil {
		t.Error("a nonbasic land was offered")
	}
	answerSearchNamed(t, g, me.ID, "Forest")
	if !g.Battlefield.Contains(ids[0]) || !b16Tapped(t, g, ids[0]) {
		t.Error("the Forest did not enter tapped")
	}
}
