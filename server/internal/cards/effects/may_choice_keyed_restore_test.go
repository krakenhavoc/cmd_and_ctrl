package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// may_choice_keyed_restore_test.go — #4291: the "you may" a resolving
// effect asks (MayChoice) and a ward's life, sacrifice and discard
// questions are keyed, so a table waiting on one is a restore point. Each
// test takes a STRICT restore (restoreRoundTrip with strict, which fails on
// a table that is not a restore point) while the prompt is open, checks the
// restored prompt still names its key and the data it carries, then answers
// it on the restored table. The *UnknownKey tests are the rollback case: a
// binary that does not know the key refuses the file.

// restoreAtConfirm restores the table strictly while a confirm owed to
// `chooser` is open, and returns the restored game and the restored prompt,
// having checked the prompt names `key`.
func restoreAtConfirm(t *testing.T, g *game.Game, chooser uuid.UUID, key string) (*game.Game, *game.PendingChoice) {
	t.Helper()
	ask := latestConfirmFor(g, chooser)
	if ask == nil || ask.PickThen != key {
		t.Fatalf("before the restore: %+v, want an open confirm keyed %q", ask, key)
	}
	g = restoreRoundTrip(t, g, true)
	ask = latestConfirmFor(g, chooser)
	if ask == nil || ask.PickThen != key {
		t.Fatalf("after the restore: %+v, want an open confirm keyed %q", ask, key)
	}
	return g, ask
}

func seatOf(g *game.Game, id uuid.UUID) *game.Player {
	for _, p := range g.Seats {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// --- the shared keys ---------------------------------------------------

// confirm/draw-cards: Bringer of the Blue Dawn's "draw two cards?", with
// the 2 riding the prompt as its word.
func TestKeyedConfirmMayDrawCardsIsARestorePoint(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		opp := g.Seats[1]
		pushCatalogPermanent(g, opp.ID, "Bringer of the Blue Dawn", "Creature — Bringer", bringerBlueDawnOracle, false)
		advanceToUpkeepOfSeat(t, g, 1)
		passPriorityAroundTable(t, g)
		g, ask := restoreAtConfirm(t, g, opp.ID, mayDrawCards.Key())
		if ask.PickMatchArg != "2" {
			t.Errorf("the restored prompt's word is %q, want 2", ask.PickMatchArg)
		}
		opp = seatOf(g, opp.ID)
		before := opp.Hand.Size()
		if err := g.ResolveConfirm(ask.ID, opp.ID, accept); err != nil {
			t.Fatal(err)
		}
		want := before
		if accept {
			want += 2
		}
		if got := opp.Hand.Size(); got != want {
			t.Errorf("accept=%v: hand %d on the restored table, want %d", accept, got, want)
		}
	}
}

// confirm/gain-life: Deathgreeter.
func TestKeyedConfirmMayGainLifeIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	victim := drcBear(g, opp.ID)
	killCreature(t, g, opp.ID, victim)
	drcSettle(t, g, me.ID)
	g, ask := restoreAtConfirm(t, g, me.ID, mayGainLife.Key())
	me = seatOf(g, me.ID)
	life := me.Life
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Life != life+1 {
		t.Errorf("life %d on the restored table, want %d", me.Life, life+1)
	}
}

// confirm/create-tokens: Darien's number of Soldiers is the damage the
// triggering event carried, which rides the prompt as its word.
func TestKeyedConfirmMayCreateTokensIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b43Catalog(g, me.ID, "Darien, King of Kjeldor", "Legendary Creature — Human Soldier", b43DarienOracle, 3, 3)
	src := b43Creature(g, g.Seats[1].ID, "Some Source", "Creature — Elemental", 1, 1)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, me.ID, 3) })
	passPriorityAroundTable(t, g)
	g, ask := restoreAtConfirm(t, g, me.ID, mayCreateTokens.Key())
	if ask.PickMatchArg != tokenWord(3, "1/1 white Soldier") {
		t.Errorf("the restored prompt's word is %q", ask.PickMatchArg)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := b43TokensNamed(g, me.ID, "Soldier"); n != 3 {
		t.Errorf("%d Soldiers on the restored table, want 3", n)
	}
}

// confirm/copy-this-spell: Chain of Smog's copy, offered to the player it
// hit after their discard.
func TestKeyedConfirmMayCopyThisSpellIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	castCatalogSpell(t, g, "Chain of Smog", "Sorcery", chainOfSmogOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	hand := discardChoiceFor(g, opp.ID)
	if hand == nil {
		t.Fatalf("no discard prompt: %+v", g.PendingChoices)
	}
	answerDiscard(t, g, opp.ID, opp.Hand.Cards[3].InstanceID, opp.Hand.Cards[5].InstanceID)
	g, ask := restoreAtConfirm(t, g, opp.ID, mayCopyThisSpell.Key())
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	if latestPickTarget(g, opp.ID) == nil {
		t.Errorf("the copy's controller is not asked for a new target on the restored table: %+v", g.PendingChoices)
	}
}

// confirm/pay-life-draw: Erebos pays the life the prompt declares and
// draws the number of cards its word holds.
func TestKeyedConfirmMayPayLifeDrawIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Erebos, Bleak-Hearted",
		TypeLine: "Legendary Enchantment Creature — God", OracleID: b39ErebosOracle,
		ManaCost: "{3}{B}", Power: 5, Toughness: 6, Owner: me.ID, Controller: me.ID,
	})
	fodder := b39Creature(g, me.ID, "Fodder", "Creature — Goblin", 1, 1)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	passPriorityAroundTable(t, g)
	g, ask := restoreAtConfirm(t, g, me.ID, mayPayLifeDraw.Key())
	if ask.LifeCost != 2 || ask.PickMatchArg != "1" {
		t.Errorf("the restored prompt declares life %d and word %q, want 2 and 1", ask.LifeCost, ask.PickMatchArg)
	}
	me = seatOf(g, me.ID)
	hand, life := me.Hand.Size(), me.Life
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if me.Hand.Size() != hand+1 || me.Life != life-2 {
		t.Errorf("hand %d → %d, life %d → %d on the restored table, want +1 and -2", hand, me.Hand.Size(), life, me.Life)
	}
}

// confirm/search-named-library-or-graveyard: Ajani's Aid's search names the
// card in the prompt's word, and the Ajani in the graveyard is found after
// the restore.
func TestKeyedConfirmMaySearchNamedLibraryOrGraveyardIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	ajani := pushGraveyardCardWithTypeLineID(me, "Ajani, Valiant Protector", "Legendary Planeswalker — Ajani")
	castCatalogSpell(t, g, "Ajani's Aid", "Enchantment", lgAjanisAidOracle, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	g, ask := restoreAtConfirm(t, g, me.ID, maySearchNamedLibraryOrGraveyard.Key())
	if ask.PickMatchArg != "Ajani, Valiant Protector" {
		t.Errorf("the restored prompt's word is %q, want the card's name", ask.PickMatchArg)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	answerSearchByID(t, g, me.ID, ajani)
	passPriorityAroundTable(t, g)
	if !seatOf(g, me.ID).Hand.Contains(ajani) {
		t.Error("Ajani was not put into the hand on the restored table")
	}
}

// confirm/fabricate: Angel of Invention's counters-or-Servos, with the 2 as
// its word. A "no" makes the Servos.
func TestKeyedConfirmMayFabricateIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Angel of Invention", "Creature — Angel", "28c7f2b6-ba67-4ccf-9e1b-99c89a9d1f72", nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	g, ask := restoreAtConfirm(t, g, me.ID, mayFabricate.Key())
	if ask.PickMatchArg != "2" {
		t.Errorf("the restored prompt's word is %q, want 2", ask.PickMatchArg)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if n := countOnBattlefield(g, "Servo", me.ID); n != 2 {
		t.Errorf("%d Servos on the restored table, want 2", n)
	}
}

// confirm/laccolith-blocked-damage is shared by the four Laccolith
// creatures; the card's name rides the prompt as its word and the target as
// its carried ID.
func TestKeyedConfirmMayLaccolithDamageIsARestorePoint(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := ncdAttacker(g, "Laccolith Grunt", ncdGrunt, 2, 2)
	blocker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 20})
	victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Test", Power: 1, Toughness: 20})
	pr7bAttack(t, g, opp.ID, id)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, id); err != nil {
		t.Fatal(err)
	}
	finishBlockDeclarations(t, g)
	for i := 0; i < 24 && latestConfirmFor(g, me.ID) == nil; i++ {
		if open := latestPickTarget(g, me.ID); open != nil {
			if err := g.ResolvePickTargets(open.ID, me.ID, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	g, ask := restoreAtConfirm(t, g, me.ID, mayLaccolithDamage.Key())
	if ask.PickMatchArg != "Laccolith Grunt" || len(ask.OptionCarry) != 2 || ask.OptionCarry[1] != victim {
		t.Errorf("the restored prompt has word %q and carries %v, want the card's name and the target %s", ask.PickMatchArg, ask.OptionCarry, victim)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	pr7bDamageStep(t, g)
	if got := pr6Marked(g, victim); got != 2 {
		t.Errorf("the target took %d on the restored table, want 2", got)
	}
	if got := pr6Marked(g, blocker); got != 0 || !ncdMarked(g, id) {
		t.Errorf("the blocker took %d and the Grunt is marked %v, want 0 and marked", got, ncdMarked(g, id))
	}
}

// --- state that rides the prompt ---------------------------------------

// Necroskitter's answer reads the card that died off the asking item's
// triggering event, which the prompt carries as data.
func TestKeyedConfirmNecroskitterAnswerReadsTheTriggerEventAfterARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Necroskitter", "Creature — Elemental", drcNecroskitter, 1, 4)
	victim := drcWithMinusOne(g, opp.ID, opp.ID, 1)
	killCreature(t, g, opp.ID, victim)
	drcSettle(t, g, me.ID)
	g, ask := restoreAtConfirm(t, g, me.ID, necroskitterAnswer.Key())
	if ask.Asked == nil || ask.Asked.Trigger == nil {
		t.Fatalf("the restored prompt carries no triggering event: %+v", ask.Asked)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	c, ok := battlefieldCard(g, victim)
	if !ok || c.Controller != me.ID {
		t.Errorf("the creature is not on the battlefield under my control on the restored table: %+v ok=%v", c, ok)
	}
}

// The Second Doctor asks each player in turn; the players who have said yes
// and the ones still to ask ride the prompt, so a restore between two
// questions loses neither list.
func TestKeyedConfirmSecondDoctorCarriesItsPlayersAcrossARestore(t *testing.T) {
	g, me := secondDoctorGame(t)
	seatIdx := g.Turn.ActiveSeat
	seats := func(k int) *game.Player { return g.Seats[(seatIdx+k)%len(g.Seats)] }
	a, b, c := seats(1), seats(2), seats(3)
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Hand.Size()
	}
	advanceToEndStepOf(t, g, seatIdx)
	answer := func(g *game.Game, who uuid.UUID, yes bool) {
		t.Helper()
		for i := 0; i < 64 && latestConfirmFor(g, who) == nil; i++ {
			if err := g.PassPriority(); err != nil {
				t.Fatal(err)
			}
		}
		answerMayChoice(t, g, who, yes)
	}
	answer(g, me.ID, true)
	answer(g, a.ID, true)
	g, ask := restoreAtConfirm(t, g, b.ID, secondDoctorAnswer.Key())
	if len(ask.OptionCarry) < 4 {
		t.Errorf("the restored prompt carries %v, want the controller, the two who said yes and the players still to ask", ask.OptionCarry)
	}
	answer(g, b.ID, false)
	answer(g, c.ID, true)
	passPriorityAroundTable(t, g)
	for _, p := range []*game.Player{me, a, c} {
		if got := seatOf(g, p.ID).Hand.Size(); got != before[p.ID]+1 {
			t.Errorf("%s hand = %d on the restored table, want %d", p.Name, got, before[p.ID]+1)
		}
	}
	if got := seatOf(g, b.ID).Hand.Size(); got != before[b.ID] {
		t.Errorf("%s said no and drew: hand %d, want %d", b.Name, got, before[b.ID])
	}
}

// --- the unknown-key refusals ------------------------------------------

func TestKeyedConfirmMayChoiceWithAnUnknownKeyIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Deathgreeter", "Creature — Human Shaman", drcDeathgreeter, 1, 1)
	killCreature(t, g, opp.ID, drcBear(g, opp.ID))
	drcSettle(t, g, me.ID)
	refuseUnknownKey(t, g, mayGainLife.Key())
}

// --- ward ---------------------------------------------------------------

// Ward—Pay 2 life, restored strictly while the caster decides: the life
// the prompt declares and the spell it guards ride the prompt, and a
// decline on the restored table counters the spell.
func TestKeyedConfirmLifeWardConfirmIsARestorePoint(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g, _, opp, elemental := wardTable(t, "Refraction Elemental", "Creature — Elemental", refractionElementalOracle)
		blade := castAtWardedCreature(t, g, opp, elemental)
		if passUntilConfirmFor(t, g, opp.ID) == nil {
			t.Fatal("no ward prompt for the spell's controller")
		}
		g, ask := restoreAtConfirm(t, g, opp.ID, wardPayLifeAnswer.Key())
		if ask.LifeCost != 2 || len(ask.OptionCarry) != 1 {
			t.Fatalf("the restored prompt declares life %d and carries %v, want 2 and the spell", ask.LifeCost, ask.OptionCarry)
		}
		opp = seatOf(g, opp.ID)
		life := opp.Life
		if err := g.ResolveConfirm(ask.ID, opp.ID, pay); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		if pay {
			if opp.Life != life-2 || g.Battlefield.Contains(elemental) {
				t.Errorf("paid: life %d → %d and the Elemental on the battlefield %v, want -2 and the spell resolved",
					life, opp.Life, g.Battlefield.Contains(elemental))
			}
		} else if opp.Life != life || !g.Battlefield.Contains(elemental) || !opp.Graveyard.Contains(blade) {
			t.Errorf("declined: life %d → %d, the Elemental survives %v, the spell in the graveyard %v; want the spell countered",
				life, opp.Life, g.Battlefield.Contains(elemental), opp.Graveyard.Contains(blade))
		}
	}
}

// Ward—Sacrifice a creature, restored strictly at BOTH legs: the confirm,
// then the pick it chains.
func TestKeyedConfirmSacrificeWardIsARestorePointAtBothLegs(t *testing.T) {
	g, _, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	bear := b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castAtWardedCreature(t, g, opp, ripper)
	if passUntilConfirmFor(t, g, opp.ID) == nil {
		t.Fatal("no ward prompt for the spell's controller")
	}
	key := wardSacrificeKeysFor(*veinRipperSacrifice()).confirm.Key()
	g, ask := restoreAtConfirm(t, g, opp.ID, key)
	if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	pickKey := wardSacrificeKeysFor(*veinRipperSacrifice()).pick.Key()
	pick := chooseCardsChoiceFor(g, opp.ID)
	if pick == nil || pick.PickThen != pickKey {
		t.Fatalf("the pick after the confirm is %+v, want one keyed %q", pick, pickKey)
	}
	g = restoreRoundTrip(t, g, true)
	opp = seatOf(g, opp.ID)
	pick = chooseCardsChoiceFor(g, opp.ID)
	if pick == nil || pick.PickThen != pickKey || len(pick.OptionCarry) != 1 {
		t.Fatalf("the restored pick is %+v, want one keyed %q carrying the spell", pick, pickKey)
	}
	if err := g.ResolveChooseCards(pick.ID, opp.ID, []uuid.UUID{bear}); err != nil {
		t.Fatal(err)
	}
	if g.Battlefield.Contains(bear) || !opp.Graveyard.Contains(bear) {
		t.Error("the chosen creature was not sacrificed on the restored table")
	}
	settleDrainingAt(t, g, g.Seats[0].ID, opp.ID)
	if g.Battlefield.Contains(ripper) {
		t.Error("the ward was paid, so the removal spell should have resolved")
	}
}

// A sacrifice ward's confirm decline on the restored table counters the
// spell, and its pick answers a card the payer no longer controls the same
// way.
func TestKeyedConfirmSacrificeWardDeclineAfterARestoreCountersTheSpell(t *testing.T) {
	g, _, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	blade := castAtWardedCreature(t, g, opp, ripper)
	passUntilConfirmFor(t, g, opp.ID)
	g, ask := restoreAtConfirm(t, g, opp.ID, wardSacrificeKeysFor(*veinRipperSacrifice()).confirm.Key())
	if err := g.ResolveConfirm(ask.ID, opp.ID, false); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ripper) || !seatOf(g, opp.ID).Graveyard.Contains(blade) {
		t.Error("a declined ward must counter the spell on the restored table")
	}
}

// Ward—Discard a card (Graveyard Trespasser) and Ward—Discard a card at
// random (Alpharael): the count and the "at random" ride the prompt as its
// word.
func TestKeyedConfirmDiscardWardIsARestorePoint(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, word string
		random                       bool
	}{
		{"Graveyard Trespasser", "Creature — Human Werewolf", "0bbd6cad-9b6f-45a6-9f2e-d7b4853586ae", "1", false},
		{"Alpharael, Stonechosen", "Legendary Creature — Human Cleric", alpharaelOracle, "1 random", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, _, opp, warded := wardTable(t, tc.name, tc.typeLine, tc.oracle)
			fillHand(opp, "Filler A", "Filler B", "Filler C")
			boltWardedCreature(t, g, opp, warded)
			if passUntilConfirmFor(t, g, opp.ID) == nil {
				t.Fatal("no ward prompt for the spell's controller")
			}
			g, ask := restoreAtConfirm(t, g, opp.ID, wardDiscardAnswer.Key())
			if ask.PickMatchArg != tc.word {
				t.Errorf("the restored prompt's word is %q, want %q", ask.PickMatchArg, tc.word)
			}
			opp = seatOf(g, opp.ID)
			hand := opp.Hand.Size()
			if err := g.ResolveConfirm(ask.ID, opp.ID, true); err != nil {
				t.Fatal(err)
			}
			if tc.random {
				if got := opp.Hand.Size(); got != hand-1 {
					t.Errorf("hand %d → %d, want one card discarded at random", hand, got)
				}
			} else if n := discardOwed(g, opp.ID); n != 1 {
				t.Errorf("the payer owes %d discards on the restored table, want 1", n)
			}
		})
	}
}

func TestKeyedConfirmWardConfirmWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, opp, elemental := wardTable(t, "Refraction Elemental", "Creature — Elemental", refractionElementalOracle)
	castAtWardedCreature(t, g, opp, elemental)
	passUntilConfirmFor(t, g, opp.ID)
	refuseUnknownKey(t, g, wardPayLifeAnswer.Key())
}

func veinRipperSacrifice() *WardSacrificeCost {
	spec, ok := Lookup(veinRipperOracle)
	if !ok {
		panic("Vein Ripper is not registered")
	}
	_ = spec
	cost := WardSacrifice("a creature", Creature())
	return cost.Sacrifice
}

// Myriad offers a token copy per other opponent. The opponents who said yes,
// the one asked and the ones still to ask ride the prompt, and the attacker
// rides it as an object with its epoch, so a restore between two offers
// loses none of them and the copy's template is read again from the attacker.
func TestKeyedConfirmMyriadCarriesItsOpponentsAndAttackerAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp1, opp2, opp3 := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	b21Push(g, me.ID, "Legion Loyalty", "Enchantment", b27LegionLoyaltyOracle, 0, 0)
	knight := pushTribalCreature(g, me.ID, "Knight", "Creature — Human Knight", 3, 3, "vigilance")
	declareAttack(t, g, opp1.ID, knight)
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, me.ID, true)
	g, ask := restoreAtConfirm(t, g, me.ID, myriadAnswer.Key())
	if len(ask.OptionObjects) != 1 || ask.OptionObjects[0].ID != knight {
		t.Fatalf("the restored prompt's objects are %+v, want the attacking Knight", ask.OptionObjects)
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	attacking := map[uuid.UUID]int{}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Knight" && c.InstanceID != knight {
			if !c.Tapped || !IsToken(c) || c.Controller != me.ID {
				t.Errorf("copy %+v should be a tapped token you control", c.Name)
			}
			attacking[c.AttackingTarget]++
		}
	}
	if attacking[opp2.ID] != 1 || attacking[opp3.ID] != 1 || attacking[opp1.ID] != 0 || len(attacking) != 2 {
		t.Errorf("copies attack %v on the restored table, want one each at the two non-defending opponents", attacking)
	}
}

// Kwain asks each player in turn; the players who have drawn and the ones
// still to ask ride the prompt, with the count of the first list as its word.
func TestKeyedConfirmKwainCarriesItsPlayersAcrossARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	kwain := b12Push(g, me.ID, "Kwain, Itinerant Meddler", "Legendary Creature — Rabbit Wizard", b20KwainOracle, 1, 3)
	advanceToMain(t, g)
	hands, lives := b20Hands(g), b20Lives(g)
	b16Activate(t, g, me.ID, kwain, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	answerMayChoice(t, g, g.Seats[0].ID, true)
	answerMayChoice(t, g, g.Seats[1].ID, false)
	g, ask := restoreAtConfirm(t, g, g.Seats[2].ID, b20KwainAnswer.Key())
	if ask.PickMatchArg != "1" {
		t.Errorf("the restored prompt's word is %q, want the one player who has drawn", ask.PickMatchArg)
	}
	answerMayChoice(t, g, g.Seats[2].ID, true)
	answerMayChoice(t, g, g.Seats[3].ID, true)
	want := []struct{ hand, life int }{{1, 1}, {0, 0}, {1, 1}, {1, 1}}
	for i, p := range g.Seats {
		if p.Hand.Size() != hands[i]+want[i].hand || p.Life != lives[i]+want[i].life {
			t.Errorf("seat %d: hand %d life %d on the restored table, want hand %d life %d", i, p.Hand.Size(), p.Life,
				hands[i]+want[i].hand, lives[i]+want[i].life)
		}
	}
}

// Both legs of a sacrifice ward name a key, and a binary that lacks either
// refuses the file.
func TestKeyedConfirmSacrificeWardWithAnUnknownKeyIsRefused(t *testing.T) {
	g, _, opp, ripper := wardTable(t, "Vein Ripper", "Creature — Vampire Assassin", veinRipperOracle)
	b16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	castAtWardedCreature(t, g, opp, ripper)
	passUntilConfirmFor(t, g, opp.ID)
	keys := wardSacrificeKeysFor(*veinRipperSacrifice())
	refuseUnknownKey(t, g, keys.confirm.Key())
	prompt := confirmChoiceFor(g, opp.ID)
	if err := g.ResolveConfirm(prompt.ID, opp.ID, true); err != nil {
		t.Fatal(err)
	}
	refuseUnknownKey(t, g, keys.pick.Key())
}
