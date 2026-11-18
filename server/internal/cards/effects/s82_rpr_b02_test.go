package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82RprB02InspiredChargeOracle = "d465a3d6-5830-456c-8e7a-908b464db846"
	s82RprB02KitesailOracle       = "b079f9db-974d-4525-a894-57b754ba9dcc"
	s82RprB02IngaOracle           = "78afa295-c9f4-4a81-9f7b-5746a3cf6b31"
	s82RprB02JudithOracle         = "c01516e0-eec3-4370-b935-7674982e8850"
	s82RprB02KaheeraOracle        = "35428c77-9731-4d4d-8f75-e939a491cf86"
	s82RprB02KerugaOracle         = "ca5121b2-e6d9-40bc-a2d2-05852c4efbe8"
	s82RprB02KroxaOracle          = "2cb1c213-b3f2-4433-8f89-6a3b4c3d2795"
	s82RprB02LurrusOracle         = "3bc757c1-3adb-4321-8832-8e1cc9e687f7"
	s82RprB02LutriOracle          = "158a6225-a246-4fd6-aa57-0df8067b4383"
	s82RprB02NivOracle            = "10764b35-5c6c-4161-8579-1124d5963a0e"
	s82RprB02OboshOracle          = "6985c8a9-9299-461b-8bef-bd2b2c739c11"
)

func s82RprB02Power(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	return layeredCard(t, g, id).CurrentPower()
}

// --- Inspired Charge ----------------------------------------------------------

func TestS82RprB02InspiredChargePumpsOnlyTheCreaturesYouControlWhenItResolves(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	floatMana(t, g, me, "{W}{W}{C}{C}")
	b12PlayFromHand(t, g, "Inspired Charge", "Instant", s82RprB02InspiredChargeOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if p, tough := rfPT(t, g, mine); p != 4 || tough != 3 {
		t.Errorf("my creature is %d/%d, want 4/3", p, tough)
	}
	if p, tough := rfPT(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
	late := b12Creature(g, me.ID, "Late Bear", "Creature — Bear", 2, 2)
	if p, _ := rfPT(t, g, late); p != 2 {
		t.Errorf("a creature that entered afterwards has power %d, want 2 (CR 611.2c)", p)
	}
}

// --- Kitesail -----------------------------------------------------------------

func TestS82RprB02KitesailGivesPlusOneAndFlyingOnceEquipped(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	sail := pushCatalogPermanent(g, me.ID, "Kitesail", "Artifact — Equipment", s82RprB02KitesailOracle, false)
	toMain(t, g)
	if hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Fatal("the creature flies before it is equipped")
	}
	floatMana(t, g, me, "{C}{C}")
	b16Activate(t, g, me.ID, sail, 0, game.ActivateAbilityParams{Targets: cardRefs(bear), Strict: true})
	passPriorityAroundTable(t, g)
	if !hasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the equipped creature does not fly")
	}
	if got := s82RprB02Power(t, g, bear); got != 3 {
		t.Errorf("equipped power %d, want 3", got)
	}
}

// --- Inga Rune-Eyes -----------------------------------------------------------

func TestS82RprB02IngaScriesThreeWhenItEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	seedLibraryTop(me, "A", "Instant", "")
	seedLibraryTop(me, "B", "Instant", "")
	seedLibraryTop(me, "C", "Instant", "")
	floatMana(t, g, me, "{U}{C}{C}{C}")
	b12PlayFromHand(t, g, "Inga Rune-Eyes", "Legendary Creature — Human Wizard", s82RprB02IngaOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("entering did not scry")
	}
	if len(c.ScryCards) != 3 {
		t.Errorf("scry looks at %d cards, want 3", len(c.ScryCards))
	}
}

func TestS82RprB02IngaDrawsThreeOnlyWithThreeCreatureDeathsThisTurn(t *testing.T) {
	for _, tc := range []struct {
		name   string
		others int
		want   int
	}{{"two died", 1, 0}, {"three died", 2, 3}} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			toMain(t, g)
			inga := pushCatalogPermanent(g, me.ID, "Inga Rune-Eyes", "Legendary Creature — Human Wizard", s82RprB02IngaOracle, false)
			for i := 0; i < tc.others; i++ {
				killCreature(t, g, me.ID, b12Creature(g, me.ID, "Doomed", "Creature — Bear", 1, 1))
			}
			seedLibraryTop(me, "A", "Instant", "")
			seedLibraryTop(me, "B", "Instant", "")
			seedLibraryTop(me, "C", "Instant", "")
			hand := me.Hand.Size()
			killCreature(t, g, me.ID, inga)
			passPriorityAroundTable(t, g)
			if got := me.Hand.Size() - hand; got != tc.want {
				t.Errorf("drew %d, want %d", got, tc.want)
			}
		})
	}
}

// --- Judith, the Scourge Diva -------------------------------------------------

func TestS82RprB02JudithPumpsOnlyYourOtherCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	judith := b12Push(g, me.ID, "Judith, the Scourge Diva", "Legendary Creature — Human Shaman", s82RprB02JudithOracle, 2, 2)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	toMain(t, g)
	if got := s82RprB02Power(t, g, judith); got != 2 {
		t.Errorf("Judith pumped herself: power %d", got)
	}
	if got := s82RprB02Power(t, g, mine); got != 3 {
		t.Errorf("my other creature's power %d, want 3", got)
	}
	if got := s82RprB02Power(t, g, theirs); got != 2 {
		t.Errorf("an opponent's creature's power %d, want 2", got)
	}
}

func TestS82RprB02JudithPingsWhenANontokenCreatureOfYoursDiesButNotATokenOrAnOpponents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Judith, the Scourge Diva", "Legendary Creature — Human Shaman", s82RprB02JudithOracle, 2, 2)
	toMain(t, g)
	token := b12Creature(g, me.ID, "Soldier", "Token Creature — Soldier", 1, 1)
	killCreature(t, g, me.ID, token)
	killCreature(t, g, opp.ID, b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2))
	if latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID) != nil {
		t.Fatal("Judith triggered for a token or an opponent's creature")
	}
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2))
	pickTriggerTargetPlayer(t, g, me.ID, opp.ID)
	before := opp.Life
	passPriorityAroundTable(t, g)
	if got := before - opp.Life; got != 1 {
		t.Errorf("Judith dealt %d damage, want 1", got)
	}
}

// --- Kaheera, the Orphanguard -------------------------------------------------

func TestS82RprB02KaheeraLordsOnlyYourOtherCatsElementalsNightmaresDinosaursAndBeasts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaheera := b12Push(g, me.ID, "Kaheera, the Orphanguard", "Legendary Creature — Cat Beast", s82RprB02KaheeraOracle, 3, 2)
	dino := b12Creature(g, me.ID, "Dino", "Creature — Dinosaur", 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Cat", "Creature — Cat", 2, 2)
	toMain(t, g)
	if p, tough := rfPT(t, g, dino); p != 3 || tough != 3 {
		t.Errorf("my Dinosaur is %d/%d, want 3/3", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, dino, "vigilance") {
		t.Error("my Dinosaur lacks vigilance")
	}
	if p, _ := rfPT(t, g, bear); p != 2 || hasKeywordOnBattlefield(t, g, bear, "vigilance") {
		t.Error("a Bear was lorded")
	}
	if p, _ := rfPT(t, g, theirs); p != 2 || hasKeywordOnBattlefield(t, g, theirs, "vigilance") {
		t.Error("an opponent's Cat was lorded")
	}
	if p, tough := rfPT(t, g, kaheera); p != 3 || tough != 2 {
		t.Errorf("Kaheera lorded herself: %d/%d", p, tough)
	}
	if !hasKeywordOnBattlefield(t, g, kaheera, "vigilance") {
		t.Error("Kaheera lacks her own vigilance")
	}
}

// --- Keruga, the Macrosage ----------------------------------------------------

func TestS82RprB02KerugaDrawsForEachOtherPermanentWithManaValueThreeOrMore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	b12Permanent(g, me.ID, "Three", "Artifact")
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].ManaCost = "{3}"
	big := b12Permanent(g, me.ID, "Five", "Artifact")
	_ = big
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].ManaCost = "{4}{R}"
	small := b12Permanent(g, me.ID, "Two", "Artifact")
	_ = small
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].ManaCost = "{2}"
	theirs := b12Permanent(g, opp.ID, "Theirs", "Artifact")
	_ = theirs
	g.Battlefield.Cards[len(g.Battlefield.Cards)-1].ManaCost = "{6}"
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	for i := 0; i < 4; i++ {
		seedLibraryTop(me, "Card", "Instant", "")
	}
	hand := me.Hand.Size()
	floatMana(t, g, me, "{G}{G}{C}{C}{C}")
	b12PlayFromHand(t, g, "Keruga, the Macrosage", "Legendary Creature — Dinosaur Hippo", s82RprB02KerugaOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// The two permanents with mana value 3 or more (the {3} and the {4}{R}).
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d cards, want 2", got)
	}
}

// --- Kroxa, Titan of Death's Hunger -------------------------------------------

// s82RprB02Drive plays the table until the stack is empty: trigger orders
// are kept as offered, each owed discard is answered with the card in
// picks (or the first in hand), everything else is passed.
func s82RprB02Drive(t *testing.T, g *game.Game, picks map[uuid.UUID]uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				if err := g.ResolveTriggerOrder(c.ID, c.Chooser, append([]uuid.UUID(nil), c.TriggerOrderIDs...)); err != nil {
					t.Fatalf("ResolveTriggerOrder: %v", err)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		for _, s := range g.Seats {
			if discardOwed(g, s.ID) > 0 {
				if id, ok := picks[s.ID]; ok {
					answerDiscard(t, g, s.ID, id)
				} else {
					discardFromHand(t, g, s.ID)
				}
				answered = true
				break
			}
		}
		if answered {
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the table never settled")
}

func s82RprB02KroxaTable(t *testing.T) (*game.Game, *game.Player, []*game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	toMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	var opps []*game.Player
	for _, s := range g.Seats {
		if s.ID != me.ID {
			for s.Hand.Size() > 0 {
				s.Hand.Cards = s.Hand.Cards[:len(s.Hand.Cards)-1]
			}
			opps = append(opps, s)
		}
	}
	return g, me, opps
}

func TestS82RprB02KroxaIsSacrificedWhenCastFromHandAndPunishesWhoDidNotDiscardANonland(t *testing.T) {
	g, me, opps := s82RprB02KroxaTable(t)
	spell := handCardFull(opps[0], "Spell", "Sorcery", "", "", nil)
	handCardFull(opps[1], "Plains", "Basic Land — Plains", "", "", nil)
	// opps[2] has an empty hand.
	lives := []int{opps[0].Life, opps[1].Life, opps[2].Life}
	floatMana(t, g, me, "{B}{R}")
	kroxa := b12PlayFromHand(t, g, "Kroxa, Titan of Death's Hunger", "Legendary Creature — Elder Giant", s82RprB02KroxaOracle, game.CastSpellParams{})
	s82RprB02Drive(t, g, map[uuid.UUID]uuid.UUID{opps[0].ID: spell})
	if g.Battlefield.Contains(kroxa) {
		t.Error("Kroxa stayed on the battlefield after being cast from hand")
	}
	if opps[0].Life != lives[0] {
		t.Errorf("the opponent who discarded a nonland lost %d life", lives[0]-opps[0].Life)
	}
	if lives[1]-opps[1].Life != 3 {
		t.Errorf("the opponent who discarded a land lost %d life, want 3", lives[1]-opps[1].Life)
	}
	if lives[2]-opps[2].Life != 3 {
		t.Errorf("the opponent with no hand lost %d life, want 3", lives[2]-opps[2].Life)
	}
}

func TestS82RprB02KroxaEscapesAndStays(t *testing.T) {
	g, me, opps := s82RprB02KroxaTable(t)
	id := seedGraveyardCard(t, g, "Kroxa, Titan of Death's Hunger", "Legendary Creature — Elder Giant", s82RprB02KroxaOracle)
	pay := seedGraveyardFodder(t, g, 5)
	lives := map[uuid.UUID]int{}
	for _, o := range opps {
		lives[o.ID] = o.Life
	}
	floatMana(t, g, me, "{B}{B}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay}); err != nil {
		t.Fatalf("escape cast: %v", err)
	}
	s82RprB02Drive(t, g, nil)
	if !g.Battlefield.Contains(id) {
		t.Error("an escaped Kroxa was sacrificed")
	}
	for _, o := range opps {
		if o.Life != lives[o.ID]-3 {
			t.Errorf("an opponent with an empty hand is at %d, want %d", o.Life, lives[o.ID]-3)
		}
	}
}

func TestS82RprB02KroxaRefusesEscapeWithTooFewCards(t *testing.T) {
	g, me, _ := s82RprB02KroxaTable(t)
	id := seedGraveyardCard(t, g, "Kroxa, Titan of Death's Hunger", "Legendary Creature — Elder Giant", s82RprB02KroxaOracle)
	pay := seedGraveyardFodder(t, g, 4)
	floatMana(t, g, me, "{B}{B}{R}{R}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "escape", AltCostIDs: pay, Strict: true}); err == nil {
		t.Fatal("escape was paid with four cards")
	}
}

// --- Lurrus of the Dream-Den --------------------------------------------------

func TestS82RprB02LurrusCastsOneCheapPermanentFromTheGraveyardEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Lurrus of the Dream-Den", "Legendary Creature — Cat Nightmare", s82RprB02LurrusOracle, false)
	cheap := pushGraveyardPermanent(me, "Cheap Bear", "Creature — Bear", "{1}")
	cheap2 := pushGraveyardPermanent(me, "Cheap Gear", "Artifact", "")
	pricey := pushGraveyardPermanent(me, "Pricey Bear", "Creature — Bear", "{3}")
	spell := me.Graveyard.PushTop
	_ = spell
	sorcery := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: sorcery, Name: "Cheap Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})

	if err := g.CastSpell(me.ID, pricey, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a mana value 3 permanent was cast from the graveyard")
	}
	if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a sorcery was cast from the graveyard")
	}
	floatMana(t, g, me, "{C}")
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("a mana value 1 permanent was refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the cheap permanent did not resolve")
	}
	if err := g.CastSpell(me.ID, cheap2, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("a second permanent was cast from the graveyard in the same turn")
	}
}

// --- Lutri, the Spellchaser ---------------------------------------------------

func TestS82RprB02LutriCopiesYourSpellOnlyWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	// A Lutri that was put onto the battlefield rather than cast asks nothing.
	pushCatalogPermanent(g, opp.ID, "Lutri, the Spellchaser", "Legendary Creature — Elemental Otter", s82RprB02LutriOracle, false)
	if latestPickTarget(g, opp.ID) != nil {
		t.Fatal("a Lutri that was not cast asked for a target")
	}
	// Cast: Lightning Bolt on the stack, Lutri flashed in response.
	floatMana(t, g, me, "{R}{R}{R}{R}{R}")
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	lutri := handCardFull(me, "Lutri, the Spellchaser", "Legendary Creature — Elemental Otter", "{1}{U/R}{U/R}", s82RprB02LutriOracle, []string{"U", "R"})
	if err := g.CastSpell(me.ID, lutri, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash Lutri in: %v", err)
	}
	before := opp.Life
	picked, copyRetargeted := false, false
	for i := 0; i < 32 && !(picked && stackFullyEmpty(g)); i++ {
		if pick := latestPickTarget(g, me.ID); pick != nil {
			// The first prompt is Lutri's target; the second is the copy's
			// "choose new targets" (CR 707.10c), kept on the opponent.
			ref := game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}
			if !picked {
				ref = game.TargetRef{Kind: game.TargetCard, ID: bolt}
			}
			if err := g.ResolvePickTarget(pick.ID, me.ID, ref); err != nil {
				t.Fatalf("ResolvePickTarget: %v", err)
			}
			if picked {
				copyRetargeted = true
			}
			picked = true
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if !picked {
		t.Fatal("Lutri's enter trigger did not ask for a target")
	}
	if !copyRetargeted {
		t.Error("the copy was not offered new targets")
	}
	if got := before - opp.Life; got != 6 {
		t.Errorf("the opponent lost %d life, want 6 (Bolt and its copy)", got)
	}
}

// --- Niv-Mizzet Reborn --------------------------------------------------------

func TestS82RprB02NivMizzetRebornTakesOneCardPerColorPairAndBottomsTheRest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMain(t, g)
	mk := func(name, cost string, colors []string) uuid.UUID {
		c := game.NewCard(name, me.ID)
		c.TypeLine = "Creature — Test"
		c.ManaCost = cost
		c.Colors = colors
		me.Library.PushTop(c)
		return c.InstanceID
	}
	wu1 := mk("WU one", "{W}{U}", []string{"W", "U"})
	wu2 := mk("WU two", "{1}{W}{U}", []string{"W", "U"})
	br := mk("BR", "{B}{R}", []string{"B", "R"})
	mono := mk("Mono", "{G}", []string{"G"})
	three := mk("Three", "{W}{U}{B}", []string{"W", "U", "B"})
	for i := 0; i < 5; i++ {
		mk("Filler", "{1}", nil)
	}
	hand := me.Hand.Size()
	floatMana(t, g, me, "{W}{U}{B}{R}{G}")
	b12PlayFromHand(t, g, "Niv-Mizzet Reborn", "Legendary Creature — Dragon Avatar", s82RprB02NivOracle, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	// White-blue is the first pair asked: two candidates, one must be taken.
	c := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("want one prompt over the two white-blue cards, got %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{wu2}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	// Black-red is next (the pairs between hold no card).
	c = latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if c == nil || len(c.ChooseCards) != 1 {
		t.Fatalf("want a prompt over the one black-red card, got %+v", c)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{br}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	passPriorityAroundTable(t, g)
	if latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID) != nil {
		t.Fatal("a third prompt went up")
	}
	for _, id := range []uuid.UUID{wu2, br} {
		if !me.Hand.Contains(id) {
			t.Errorf("a chosen card is not in the hand")
		}
	}
	for _, id := range []uuid.UUID{wu1, mono, three} {
		if me.Hand.Contains(id) {
			t.Errorf("an unchosen card went to the hand")
		}
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
	bottom := map[uuid.UUID]bool{wu1: true, mono: true, three: true}
	for i := 0; i < 8; i++ {
		delete(bottom, me.Library.Cards[i].InstanceID)
	}
	if len(bottom) != 0 {
		t.Errorf("%d unchosen revealed cards are not at the bottom of the library", len(bottom))
	}
}

// --- Obosh, the Preypiercer ---------------------------------------------------

func TestS82RprB02OboshDoublesDamageFromYourOddManaValueSourcesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMain(t, g)
	pushCatalogPermanent(g, me.ID, "Obosh, the Preypiercer", "Legendary Creature — Hellion Horror", s82RprB02OboshOracle, false)
	mk := func(owner uuid.UUID, cost string) uuid.UUID {
		id := b12Creature(g, owner, "Source "+cost, "Creature — Bear", 2, 2)
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].ManaCost = cost
			}
		}
		return id
	}
	odd, even, theirs := mk(me.ID, "{3}"), mk(me.ID, "{2}"), mk(opp.ID, "{1}")
	hit := func(src uuid.UUID) int {
		before := opp.Life
		g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 2) })
		return before - opp.Life
	}
	if got := hit(odd); got != 4 {
		t.Errorf("an odd source dealt %d, want 4", got)
	}
	if got := hit(even); got != 2 {
		t.Errorf("an even source dealt %d, want 2", got)
	}
	if got := hit(theirs); got != 2 {
		t.Errorf("an opponent's odd source dealt %d, want 2", got)
	}
}
