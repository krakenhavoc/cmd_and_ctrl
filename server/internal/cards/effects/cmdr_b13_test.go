package effects

import (
	"errors"
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
