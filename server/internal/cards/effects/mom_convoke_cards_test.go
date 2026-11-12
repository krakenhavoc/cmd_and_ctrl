package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mom_convoke_cards_test.go — #5647: the March of the Machine cards that
// read or grant "has convoke": Flockchaser Phantom, Joyful Stormsculptor,
// Kasla, the Broken Halo, Saint Traft and Rem Karolus and Invasion of
// Segovia // Caetus, Sea Tyrant of Segovia.

const (
	mcFlockchaserOracle   = "6471335e-e054-4e68-b02e-052d65a2fdfc"
	mcStormsculptorOracle = "39867856-c1f3-442a-abe9-53f8f73777d0"
	mcKaslaOracle         = "b2d4ce3c-00b1-4a72-b652-4c129475080f"
	mcSaintTraftOracle    = "3a1be8b3-16e2-4fae-bea7-3a98f2f9db38"
)

// mcFlockchaserCard is the Phantom as a card: a spell with printed
// convoke, which the cast triggers below are driven with.
func mcFlockchaserCard(owner uuid.UUID) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Flockchaser Phantom", TypeLine: "Creature — Spirit",
		OracleID: mcFlockchaserOracle, ManaCost: "{4}{W}{U}", Colors: []string{"W", "U"}, Power: 5, Toughness: 5,
		Owner: owner, Controller: owner}
}

// mcCastConvokeSpell casts a Flockchaser Phantom from `me`'s hand for
// mana (no creature tapped) and leaves it on the stack.
func mcCastConvokeSpell(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	c := mcFlockchaserCard(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	fillPoolColored(me, "W", 5)
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting a convoke spell: %v", err)
	}
	return c.InstanceID
}

// mcCastPlainSpell casts a {1}{G} creature with no convoke for mana.
func mcCastPlainSpell(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	c := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting a plain spell: %v", err)
	}
	return c.InstanceID
}

func mcTap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(id) })
}

func mcUntap(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
}

// --- Flockchaser Phantom ----------------------------------------------------

func TestFlockchaserPhantomIsConvokedByColour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 5)
	elf := pushTapCostPermanent(g, me.ID, "Elf", "Creature — Elf", "G")
	c := mcFlockchaserCard(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(c) })

	// Five white creatures pay {4}{W}; a green one cannot pay {U}.
	taps := append(append([]uuid.UUID{}, soldiers...), elf)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, TapIDs: taps}); err == nil {
		t.Fatal("a green creature paid {U}")
	}
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, TapIDs: soldiers}); err != nil {
		t.Fatalf("convoking the Phantom: %v", err)
	}
	if !cgTapped(g, soldiers) || !cpHadConvoke(g, c.InstanceID) {
		t.Error("the Phantom should be convoked and recorded as a spell that has convoke")
	}
	passPriorityAroundTable(t, g)
	v := s58p6View(t, g, c.InstanceID)
	for _, kw := range []string{"flying", "vigilance"} {
		if !game.HasKeyword(&v, kw) {
			t.Errorf("the Phantom has no %s", kw)
		}
	}
}

func TestFlockchaserPhantomAttackGivesTheNextSpellConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	phantom := b27Push(g, me.ID, "Flockchaser Phantom", "Creature — Spirit", mcFlockchaserOracle, "{4}{W}{U}", 5, 5, "W", "U")
	elf := pushTapCostPermanent(g, me.ID, "Elf", "Creature — Elf", "G")
	if len(nsPromises(me)) != 0 {
		t.Fatal("a promise before any attack")
	}
	lciB1Attack(t, g, phantom, opp.ID)
	if got := nsPromises(me); len(got) != 1 || !got[0].Convoke {
		t.Fatalf("the promise after the attack: %+v", got)
	}
	if tappedForTest(t, g, phantom) {
		t.Fatal("the Phantom tapped to attack; it has vigilance")
	}

	// The untapped Phantom pays {U} and the Elf {1} of an instant.
	trick := game.Card{InstanceID: uuid.New(), Name: "Trick", TypeLine: "Instant", ManaCost: "{1}{U}",
		Colors: []string{"U"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(trick) })
	if err := g.CastSpell(me.ID, trick.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{phantom, elf}}); err != nil {
		t.Fatalf("convoking the next spell: %v", err)
	}
	if !cpHadConvoke(g, trick.InstanceID) || len(nsPromises(me)) != 0 {
		t.Error("the spell should have convoke and spend the promise")
	}
}

// --- Joyful Stormsculptor ---------------------------------------------------

func TestJoyfulStormsculptorEntersWithTwoElementals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	c := game.Card{InstanceID: uuid.New(), Name: "Joyful Stormsculptor", TypeLine: "Creature — Human Shaman",
		OracleID: mcStormsculptorOracle, ManaCost: "{3}{U}{R}", Colors: []string{"U", "R"}, Power: 2, Toughness: 3,
		Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	fillPoolColored(me, "U", 4)
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the Stormsculptor: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	toks := tokensNamed(g, "Elemental")
	if len(toks) != 2 {
		t.Fatalf("made %d Elementals, want 2", len(toks))
	}
	for _, tok := range toks {
		if tok.Controller != me.ID || tok.Power != 1 || tok.Toughness != 1 || !tok.IsCreature() ||
			len(tok.Colors) != 2 || !tok.HasColorWord("U") || !tok.HasColorWord("R") {
			t.Errorf("the token is %+v, want a 1/1 blue and red Elemental creature under my control", tok)
		}
	}
}

func TestJoyfulStormsculptorDamagesOpponentsAndTheBattlesTheyProtect(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Joyful Stormsculptor", "Creature — Human Shaman", mcStormsculptorOracle, "{3}{U}{R}", 2, 3, "U", "R")
	var opps []*game.Player
	for _, p := range g.Seats {
		if p.ID != me.ID {
			opps = append(opps, p)
		}
	}
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "My Siege", TypeLine: "Battle — Siege",
		Owner: me.ID, Controller: me.ID, ProtectorPlayerID: opps[0].ID,
		Counters: map[string]int{game.CounterDefense: 5},
	})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Their Siege", TypeLine: "Battle — Siege",
		Owner: opps[1].ID, Controller: opps[1].ID, ProtectorPlayerID: me.ID,
		Counters: map[string]int{game.CounterDefense: 5},
	})
	myLife := me.Life
	life := map[uuid.UUID]int{}
	for _, o := range opps {
		life[o.ID] = o.Life
	}

	// A spell with no convoke: nothing.
	mcCastPlainSpell(t, g, me)
	passPriorityAroundTable(t, g)
	for _, o := range opps {
		if o.Life != life[o.ID] {
			t.Fatalf("a spell with no convoke dealt damage")
		}
	}

	// A spell with convoke, paid for with mana alone.
	mcCastConvokeSpell(t, g, me)
	passPriorityAroundTable(t, g)
	for _, o := range opps {
		if o.Life != life[o.ID]-1 {
			t.Errorf("an opponent's life went %d → %d, want 1 damage", life[o.ID], o.Life)
		}
	}
	if me.Life != myLife {
		t.Error("the Stormsculptor damaged its controller")
	}
	if got := battlefieldCardFor(g, theirs).Counters[game.CounterDefense]; got != 4 {
		t.Errorf("the battle an opponent protects has %d defense, want 4", got)
	}
	if got := battlefieldCardFor(g, mine).Counters[game.CounterDefense]; got != 5 {
		t.Errorf("the battle I protect has %d defense, want 5", got)
	}
}

func TestJoyfulStormsculptorSeesAPromisedConvoke(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b27Push(g, me.ID, "Joyful Stormsculptor", "Creature — Human Shaman", mcStormsculptorOracle, "{3}{U}{R}", 2, 3, "U", "R")
	cpActivateWand(t, g, me)
	before := opp.Life
	mcCastPlainSpell(t, g, me)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("the promised spell has convoke: life %d → %d, want 1 damage", before, opp.Life)
	}
	// The promise is spent: the next plain spell has no convoke.
	mcCastPlainSpell(t, g, me)
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Error("a second spell was treated as having convoke")
	}
}

func TestJoyfulStormsculptorIgnoresAnOpponentsConvokeSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	b27Push(g, opp.ID, "Joyful Stormsculptor", "Creature — Human Shaman", mcStormsculptorOracle, "{3}{U}{R}", 2, 3, "U", "R")
	before := me.Life
	mcCastConvokeSpell(t, g, me)
	passPriorityAroundTable(t, g)
	if me.Life != before {
		t.Error("an opponent's Stormsculptor triggered on my spell")
	}
}

// --- Kasla, the Broken Halo -------------------------------------------------

func TestKaslaScriesTwoThenDrawsOnAConvokeSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	kasla := b27Push(g, me.ID, "Kasla, the Broken Halo", "Legendary Creature — Angel Ally", mcKaslaOracle, "{3}{U}{R}{W}", 5, 4, "U", "R", "W")
	ids := seedLibrary(me, "One", "Two", "Three")

	mcCastPlainSpell(t, g, me)
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("a spell with no convoke triggered Kasla")
	}

	mcCastConvokeSpell(t, g, me)
	hand := me.Hand.Size()
	resolveTopOfStack(t, g)
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("the scry prompt: %+v, want two cards", c)
	}
	if me.Hand.Size() != hand {
		t.Fatal("the draw did not wait for the scry")
	}
	// Bottom the top card, keep the second: the card drawn is the second.
	if err := g.ResolveScry(c.ID, me.ID, []uuid.UUID{ids[0]}, []uuid.UUID{ids[1]}); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if me.Hand.Size() != hand+1 || !me.Hand.Contains(ids[1]) {
		t.Errorf("Kasla should draw the card the scry left on top")
	}
	v := s58p6View(t, g, kasla)
	for _, kw := range []string{"flying", "vigilance", "haste"} {
		if !game.HasKeyword(&v, kw) {
			t.Errorf("Kasla has no %s", kw)
		}
	}
}

func TestKaslaIsConvokedAndDoesNotTriggerOnHerOwnCast(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	soldiers := pushTapCostSoldiers(g, me.ID, 4)
	c := game.Card{InstanceID: uuid.New(), Name: "Kasla, the Broken Halo", TypeLine: "Legendary Creature — Angel Ally",
		OracleID: mcKaslaOracle, ManaCost: "{3}{U}{R}{W}", Colors: []string{"U", "R", "W"}, Power: 5, Toughness: 4,
		Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "R", 1)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, TapIDs: soldiers}); err != nil {
		t.Fatalf("convoking Kasla: %v", err)
	}
	if len(g.StackMeta) != 1 {
		t.Errorf("Kasla's own cast put %d items on the stack, want only the spell", len(g.StackMeta))
	}
	passPriorityAroundTable(t, g)
	if scryChoiceFor(g, me.ID) != nil {
		t.Error("Kasla triggered on her own cast")
	}
}

// --- Saint Traft and Rem Karolus --------------------------------------------

func TestSaintTraftAndRemMakeAHumanThenASpiritThenAnAngel(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pair := b27Push(g, me.ID, "Saint Traft and Rem Karolus", "Legendary Creature — Spirit Human", mcSaintTraftOracle, "{U}{R}{W}", 3, 4, "U", "R", "W")
	want := []struct {
		name, color string
		pt          int
		flying      bool
	}{
		{"Human", "R", 1, false},
		{"Spirit", "U", 1, true},
		{"Angel", "W", 4, true},
	}
	for i, w := range want {
		mcTap(g, pair)
		passPriorityAroundTable(t, g)
		mcUntap(g, pair)
		toks := tokensNamed(g, w.name)
		if len(toks) != 1 {
			t.Fatalf("resolution %d: %d %s tokens, want 1", i+1, len(toks), w.name)
		}
		tok := toks[0]
		if tok.Controller != me.ID || tok.Power != w.pt || tok.Toughness != w.pt || len(tok.Colors) != 1 ||
			!tok.HasColorWord(w.color) || game.HasKeyword(&tok, "flying") != w.flying {
			t.Errorf("resolution %d made %+v", i+1, tok)
		}
	}
	// A fourth resolution this turn makes nothing.
	mcTap(g, pair)
	passPriorityAroundTable(t, g)
	if n := len(tokensNamed(g, "Human")) + len(tokensNamed(g, "Spirit")) + len(tokensNamed(g, "Angel")); n != 3 {
		t.Errorf("%d tokens after the fourth resolution, want 3", n)
	}
}

func TestSaintTraftAndRemUntapWhenYouCastAConvokeSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pair := b27Push(g, me.ID, "Saint Traft and Rem Karolus", "Legendary Creature — Spirit Human", mcSaintTraftOracle, "{U}{R}{W}", 3, 4, "U", "R", "W")

	// Tapped to convoke the spell itself: the tap makes a Human, and the
	// cast untaps the pair.
	c := mcFlockchaserCard(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(c) })
	fillPoolColored(me, "W", 4)
	fillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{pair}}); err != nil {
		t.Fatalf("convoking with the pair: %v", err)
	}
	if !tappedForTest(t, g, pair) {
		t.Fatal("the pair was not tapped for convoke")
	}
	// Two triggers of one controller at once: the tap and the cast.
	settleOrderingTriggers(t, g)
	if tappedForTest(t, g, pair) {
		t.Error("casting a spell that has convoke should untap the pair")
	}
	if n := len(tokensNamed(g, "Human")); n != 1 {
		t.Errorf("becoming tapped for convoke made %d Humans, want 1", n)
	}

	// A spell with no convoke leaves them tapped.
	mcTap(g, pair)
	passPriorityAroundTable(t, g)
	toMainPhaseSameTurn(t, g)
	mcCastPlainSpell(t, g, me)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, pair) {
		t.Error("a spell with no convoke untapped the pair")
	}
}

// --- Invasion of Segovia // Caetus, Sea Tyrant of Segovia --------------------

func segoviaRow() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     "Invasion of Segovia // Caetus, Sea Tyrant of Segovia",
		Layout:   "transform",
		TypeLine: "Battle — Siege // Legendary Creature — Serpent",
		OracleID: uuid.MustParse(invasionOfSegoviaOracleID),
		CardFaces: []cards.CardFace{
			{Name: "Invasion of Segovia", TypeLine: "Battle — Siege", ManaCost: "{2}{U}", Defense: "4", Colors: []string{"U"}},
			{Name: "Caetus, Sea Tyrant of Segovia", TypeLine: "Legendary Creature — Serpent", Power: "3", Toughness: "3", Colors: []string{"U"}},
		},
	}
}

func TestInvasionOfSegoviaMakesTwoKrakensAndBecomesCaetus(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	owner := g.Seats[seat]
	protector := g.Seats[(seat+1)%len(g.Seats)]
	id := importToHand(segoviaRow(), owner)
	toMainPhaseSameTurn(t, g)
	fillPoolColored(owner, "U", 3)
	if err := g.CastSpell(owner.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting the Siege: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProtectorPrompt(t, g, owner.ID, protector.ID)
	passPriorityAroundTable(t, g)

	battle := battlefieldCardFor(g, id)
	if battle == nil || battle.Counters[game.CounterDefense] != 4 {
		t.Fatalf("the Siege: %+v, want it on the battlefield with 4 defense", battle)
	}
	krakens := tokensNamed(g, "Kraken")
	if len(krakens) != 2 {
		t.Fatalf("made %d Krakens, want 2", len(krakens))
	}
	for _, k := range krakens {
		if k.Controller != owner.ID || k.Power != 1 || k.Toughness != 1 || len(k.Colors) != 1 ||
			!k.HasColorWord("U") || !game.HasKeyword(&k, "trample") {
			t.Errorf("the token is %+v, want a 1/1 blue Kraken with trample", k)
		}
	}

	defeatBattle(t, g, id, 4)
	if !inExile(g, id) {
		t.Fatal("the defeated Siege was not exiled")
	}
	if err := castThroughResolutionCast(t, g, owner.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("casting Caetus transformed: %v", err)
	}
	passPriorityAroundTable(t, g)
	toMainPhaseSameTurn(t, g)
	landed := battlefieldCardFor(g, id)
	if landed == nil || landed.Name != "Caetus, Sea Tyrant of Segovia" || !landed.IsCreature() || landed.IsBattle() {
		t.Fatalf("landed %+v, want Caetus as a creature", landed)
	}

	// Caetus's static reaches a noncreature spell, paid by the Krakens.
	rite := game.Card{InstanceID: uuid.New(), Name: "Rite", TypeLine: "Sorcery", ManaCost: "{1}{U}",
		Colors: []string{"U"}, Owner: owner.ID, Controller: owner.ID}
	g.WithWriteLock(func() { owner.Hand.PushTop(rite) })
	taps := []uuid.UUID{krakens[0].InstanceID, krakens[1].InstanceID}
	if err := g.CastSpell(owner.ID, rite.InstanceID, game.CastSpellParams{Strict: true, TapIDs: taps}); err != nil {
		t.Fatalf("convoking a noncreature spell under Caetus: %v", err)
	}
	if !cgTapped(g, taps) || !cpHadConvoke(g, rite.InstanceID) {
		t.Error("the Krakens should be tapped and the sorcery recorded as having convoke")
	}
}

// mcPushCaetus puts the back face on the battlefield under its "#1" key.
func mcPushCaetus(g *game.Game, owner uuid.UUID) uuid.UUID {
	return b27Push(g, owner, "Caetus, Sea Tyrant of Segovia", "Legendary Creature — Serpent",
		invasionOfSegoviaOracleID+"#1", "", 3, 3, "U")
}

func TestCaetusGivesNoncreatureSpellsConvokeAndNotCreatureSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	mcPushCaetus(g, me.ID)
	b27Push(g, me.ID, "Joyful Stormsculptor", "Creature — Human Shaman", mcStormsculptorOracle, "{3}{U}{R}", 2, 3, "U", "R")
	elves := cgTwoGreenCreatures(g, me.ID)
	before := opp.Life

	bear := cgBear(me.ID)
	g.WithWriteLock(func() { me.Hand.PushTop(bear) })
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err == nil {
		t.Fatal("a creature spell was convoked under Caetus")
	}
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	if cpHadConvoke(g, bear.InstanceID) {
		t.Error("a creature spell was recorded as having convoke")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before {
		t.Error("a creature spell under Caetus triggered the Stormsculptor")
	}

	rite := game.Card{InstanceID: uuid.New(), Name: "Rite", TypeLine: "Sorcery", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rite) })
	if err := g.CastSpell(me.ID, rite.InstanceID, game.CastSpellParams{Strict: true, TapIDs: elves}); err != nil {
		t.Fatalf("convoking a sorcery under Caetus: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-1 {
		t.Errorf("a noncreature spell under Caetus has convoke: life %d → %d, want 1 damage", before, opp.Life)
	}

	// An opponent's noncreature spell gains nothing from my Caetus.
	theirElf := pushTapCostPermanent(g, opp.ID, "Their Elf", "Creature — Elf", "G")
	theirs := game.Card{InstanceID: uuid.New(), Name: "Trick", TypeLine: "Instant", ManaCost: "{G}",
		Colors: []string{"G"}, Owner: opp.ID, Controller: opp.ID}
	g.WithWriteLock(func() { opp.Hand.PushTop(theirs) })
	rite2 := game.Card{InstanceID: uuid.New(), Name: "Rite", TypeLine: "Sorcery", ManaCost: "{1}{G}",
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(rite2) })
	fillPoolColored(me, "G", 2)
	if err := g.CastSpell(me.ID, rite2.InstanceID, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatal(err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(opp.ID, theirs.InstanceID, game.CastSpellParams{Strict: true, TapIDs: []uuid.UUID{theirElf}}); err == nil {
		t.Error("an opponent's spell was convoked under my Caetus")
	}
}

func TestCaetusUntapsUpToFourTargetCreaturesAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	mcPushCaetus(g, me.ID)
	var tapped []uuid.UUID
	for i := 0; i < 4; i++ {
		tapped = append(tapped, b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2))
	}
	tapped = append(tapped, b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2))
	for _, id := range tapped {
		tmcB7Tap(g, id)
	}
	advanceTo(t, g, game.StepEnd)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTargets(p.ID, me.ID, cardRefs(tapped...)); err == nil {
		t.Fatal("five targets were accepted for up to four")
	}
	// Three of mine and the opponent's: "target creatures", anyone's.
	chosen := []uuid.UUID{tapped[0], tapped[1], tapped[2], tapped[4]}
	pickTargetRefs(t, g, me.ID, nil, cardRefs(chosen...)...)
	passPriorityAroundTable(t, g)
	for _, id := range chosen {
		if tappedForTest(t, g, id) {
			t.Errorf("a chosen creature is still tapped")
		}
	}
	if !tappedForTest(t, g, tapped[3]) {
		t.Error("an unchosen creature was untapped")
	}
}

func TestCaetusEndStepTriggerMayTargetNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	mcPushCaetus(g, me.ID)
	bear := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	tmcB7Tap(g, bear)
	advanceTo(t, g, game.StepEnd)
	b04WaitForPick(t, g, me.ID)
	pickTargetRefs(t, g, me.ID, nil)
	passPriorityAroundTable(t, g)
	if !tappedForTest(t, g, bear) {
		t.Error("a creature that was not targeted was untapped")
	}
}
