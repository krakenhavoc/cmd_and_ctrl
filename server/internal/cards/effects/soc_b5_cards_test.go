package effects

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b5_cards_test.go — Secrets of Strixhaven Commander, fifth batch
// (#4834). Each test plays the real catalog entry through the engine and
// asserts what the printed text says.

const (
	socB5EntrancingMelody = "f9f82958-a7c6-4161-b0d8-fdbf7a3dd911"
	socB5Eriette          = "f43f5855-f297-4433-984b-18263b734cfc"
	socB5Excava           = "77216636-f76d-4c5e-b1ac-ab1844753f05"
	socB5Expansion        = "afbc348a-7b27-4577-8c24-cd6d5b894dc1"
	socB5FallenIdeal      = "4b8183e6-4ff2-4254-9ee0-0ebbf8cef4a3"
	socB5FeralAppetite    = "adf731a1-fc4d-4260-bb85-535edb1d82e9"
	socB5FesteringThicket = "bdb9b2ce-342c-4935-8943-d0c3971b1e38"
	socB5ForumFilibuster  = "6a9cc11b-a223-46f0-bf58-c1e08ad1253d"
	socB5FractalHarness   = "09342fd7-3dad-431e-8b87-de8a9287e757"
)

func TestSocB5EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		socB5EntrancingMelody: "Entrancing Melody",
		socB5Eriette:          "Eriette of the Charmed Apple",
		socB5Excava:           "Excava, the Risen Past",
		socB5Expansion:        "Expansion Algorithm",
		socB5FallenIdeal:      "Fallen Ideal",
		socB5FeralAppetite:    "Feral Appetite",
		socB5FesteringThicket: "Festering Thicket",
		socB5ForumFilibuster:  "Forum Filibuster",
		socB5FractalHarness:   "Fractal Harness",
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

// socB5CastX puts a catalog spell in the active seat's hand, walks to a
// main phase and casts it with the given X and targets.
func socB5CastX(t *testing.T, g *game.Game, name, typeLine, oracle, cost string, x int, targets []game.TargetRef) (uuid.UUID, error) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	return id, g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: targets})
}

// socB5Creature seeds a creature with a real mana cost.
func socB5Creature(g *game.Game, owner uuid.UUID, name, cost string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", ManaCost: cost,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// socB5Named lists the battlefield permanents with this name that
// `controller` controls.
func socB5Named(g *game.Game, controller uuid.UUID, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// --- Entrancing Melody ----------------------------------------------------

func TestSocB5EntrancingMelodyTakesACreatureWithManaValueXForGood(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := socB5Creature(g, opp.ID, "Bear", "{1}{G}", 2, 2)
	if _, err := socB5CastX(t, g, "Entrancing Melody", "Sorcery", socB5EntrancingMelody, "{X}{U}{U}", 2, b16TargetCard(bear)); err != nil {
		t.Fatalf("cast with X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if roonbControllerOf(g, bear) != me.ID {
		t.Fatal("the creature was not taken")
	}
	// No stated duration: it does not end at cleanup (the ruling).
	advanceToUpkeepOf(t, g, 1)
	if roonbControllerOf(g, bear) != me.ID {
		t.Error("control ended at the end of the turn")
	}
}

func TestSocB5EntrancingMelodyRefusesAWrongManaValue(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	ogre := socB5Creature(g, opp.ID, "Ogre", "{2}{R}", 3, 3)
	_, err := socB5CastX(t, g, "Entrancing Melody", "Sorcery", socB5EntrancingMelody, "{X}{U}{U}", 2, b16TargetCard(ogre))
	if !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("a mana value 3 creature with X=2: err = %v, want ErrIllegalTarget", err)
	}
}

// --- Eriette of the Charmed Apple -----------------------------------------

func TestSocB5ErietteBarsCreaturesEnchantedByYourAurasFromAttackingYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	pushCatalogPermanent(g, me.ID, "Eriette of the Charmed Apple", "Legendary Creature — Human Warlock", socB5Eriette, false)
	charmed := b12Creature(g, opp.ID, "Charmed", "Creature — Bear", 2, 2)
	free := b12Creature(g, opp.ID, "Free", "Creature — Bear", 2, 2)
	theirOwn := b12Creature(g, opp.ID, "Their Own", "Creature — Bear", 2, 2)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Aura", TypeLine: auraTypeLine,
		Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: charmed}})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Aura", TypeLine: auraTypeLine,
		Owner: opp.ID, Controller: opp.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: theirOwn}})
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = pushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	advanceToDeclareAttackersOf(t, g, 1)
	for _, target := range []uuid.UUID{me.ID, walker} {
		err := g.Clone().DeclareAttacker(charmed, target)
		if !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("the charmed creature attacking %s: err = %v, want ErrIllegalAttackTarget", target, err)
		}
	}
	if err := g.Clone().DeclareAttacker(charmed, third.ID); err != nil {
		t.Errorf("the charmed creature attacking another player: %v", err)
	}
	if err := g.Clone().DeclareAttacker(free, me.ID); err != nil {
		t.Errorf("an unenchanted creature attacking me: %v", err)
	}
	if err := g.Clone().DeclareAttacker(theirOwn, me.ID); err != nil {
		t.Errorf("a creature enchanted by an opponent's Aura attacking me: %v", err)
	}
}

func TestSocB5ErietteDrainsForEachAuraYouControlAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Eriette of the Charmed Apple", "Legendary Creature — Human Warlock", socB5Eriette, false)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	for _, host := range []uuid.UUID{theirs, mine} {
		pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Aura", TypeLine: auraTypeLine,
			Owner: me.ID, Controller: me.ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: host}})
	}
	// An opponent's Aura is not one you control.
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Aura", TypeLine: auraTypeLine,
		Owner: g.Seats[1].ID, Controller: g.Seats[1].ID, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: theirs}})
	lives := make([]int, len(g.Seats))
	for i, p := range g.Seats {
		lives[i] = p.Life
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if me.Life != lives[0]+2 {
		t.Errorf("my life %d -> %d, want +2", lives[0], me.Life)
	}
	for i := 1; i < len(g.Seats); i++ {
		if g.Seats[i].Life != lives[i]-2 {
			t.Errorf("seat %d life %d -> %d, want -2", i, lives[i], g.Seats[i].Life)
		}
	}
}

// --- Excava, the Risen Past -----------------------------------------------

func TestSocB5ExcavaReturnsAnArtifactAsAOneOneSpiritWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	excava := pushCatalogPermanent(g, me.ID, "Excava, the Risen Past", "Legendary Creature — Spirit Horse", socB5Excava, false)
	pushCatalogPermanent(g, me.ID, "Soul Warden", "Creature — Human Cleric", e4506SoulWardenOracle, false)
	stone := pushGraveyardPermanent(me, "Stone", "Artifact", "{2}")
	advanceToDeclareAttackersOf(t, g, 0)
	life := me.Life
	if err := g.DeclareAttacker(excava, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	pickTriggerTarget(t, g, me.ID, stone)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(stone) {
		t.Fatal("the artifact did not come back")
	}
	c := layeredCard(t, g, stone)
	eff := c.Effective()
	if !slices.Contains(eff.Types, "Artifact") || !slices.Contains(eff.Types, "Creature") {
		t.Errorf("types %v, want an artifact creature", eff.Types)
	}
	if !c.HasSubtype("Spirit") {
		t.Errorf("subtypes %v, want a Spirit", eff.Subtypes)
	}
	if !slices.Contains(eff.Abilities, "flying") {
		t.Errorf("abilities %v, want flying", eff.Abilities)
	}
	if eff.Power != 1 || eff.Toughness != 1 {
		t.Errorf("%d/%d, want 1/1", eff.Power, eff.Toughness)
	}
	if countersOn(g, stone, game.CounterFinality) != 1 {
		t.Errorf("finality counters %d, want 1", countersOn(g, stone, game.CounterFinality))
	}
	// A creature as it entered: Soul Warden saw it (CR 614.1c).
	if me.Life != life+1 {
		t.Errorf("Soul Warden: life %d -> %d, want +1", life, me.Life)
	}
	// The finality counter exiles it (CR 122.1h).
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(stone) })
	if !g.Exile.Contains(stone) || me.Graveyard.Contains(stone) {
		t.Error("the returned card was not exiled instead of dying")
	}
}

func TestSocB5ExcavaMakesACreatureCardOneOneAndOffersOnlyLegalCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	excava := pushCatalogPermanent(g, me.ID, "Excava, the Risen Past", "Legendary Creature — Spirit Horse", socB5Excava, false)
	ogre := pushGraveyardPermanent(me, "Ogre", "Creature — Ogre", "{2}{R}")
	big := pushGraveyardPermanent(me, "Big", "Creature — Giant", "{3}{R}")
	aura := pushGraveyardPermanent(me, "Aura", auraTypeLine, "{W}")
	theirs := pushGraveyardPermanent(opp, "Theirs", "Artifact", "{1}")
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(excava, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	for name, id := range map[string]uuid.UUID{"mana value 4": big, "an Aura": aura, "an opponent's card": theirs} {
		if slices.Contains(pick.PickTargetCards, id) {
			t.Errorf("%s card is offered", name)
		}
	}
	if !slices.Contains(pick.PickTargetCards, ogre) {
		t.Fatal("the mana value 3 creature card is not offered")
	}
	pickTriggerTarget(t, g, me.ID, ogre)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ogre) {
		t.Fatal("the creature card did not come back")
	}
	if p, tough := effectivePower(t, g, ogre), effectiveToughness(t, g, ogre); p != 1 || tough != 1 {
		t.Errorf("returned creature is %d/%d, want 1/1", p, tough)
	}
	c := layeredCard(t, g, ogre)
	if !c.HasSubtype("Spirit") || !c.HasSubtype("Ogre") {
		t.Errorf("subtypes %v, want Ogre Spirit", c.Effective().Subtypes)
	}
}

// --- Expansion Algorithm --------------------------------------------------

func TestSocB5ExpansionAlgorithmProliferatesXTimesEachItsOwnChoice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	ox := b12Creature(g, me.ID, "Ox", "Creature — Ox", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(ox, game.CounterPlusOne, 1)
	})
	if _, err := socB5CastX(t, g, "Expansion Algorithm", "Sorcery", socB5Expansion, "{X}{U}{U}", 3, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerProliferate(t, g, bear, ox)
	answerProliferate(t, g, bear)
	answerProliferate(t, g, bear)
	passPriorityAroundTable(t, g)
	if proliferatePrompt(g) != nil {
		t.Error("a fourth proliferate was asked")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 4 {
		t.Errorf("bear has %d counters, want 4", n)
	}
	if n := countersOn(g, ox, game.CounterPlusOne); n != 2 {
		t.Errorf("ox has %d counters, want 2", n)
	}
}

func TestSocB5ExpansionAlgorithmForZeroAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	if _, err := socB5CastX(t, g, "Expansion Algorithm", "Sorcery", socB5Expansion, "{X}{U}{U}", 0, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if proliferatePrompt(g) != nil {
		t.Error("X=0 asked for a proliferate")
	}
	if n := countersOn(g, bear, game.CounterPlusOne); n != 1 {
		t.Errorf("bear has %d counters, want 1", n)
	}
}

// --- Fallen Ideal ---------------------------------------------------------

func TestSocB5FallenIdealGrantsFlyingAndASacrificePumpThenReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", testCreatureTypeLine, 2, 2)
	fodder := b12Creature(g, me.ID, "Fodder", testCreatureTypeLine, 1, 1)
	ideal := castCatalogSpell(t, g, "Fallen Ideal", auraTypeLine, socB5FallenIdeal, b16TargetCard(bear))
	passPriorityAroundTable(t, g)
	if host := attachmentHostOf(t, g, ideal); host.ID != bear {
		t.Fatalf("Fallen Ideal is attached to %+v", host)
	}
	if !containsString(effectiveAbilities(t, g, bear), "flying") {
		t.Error("the creature lacks flying")
	}
	idx, ref := grantedActivatedIndex(t, g, bear)
	if idx < 0 {
		t.Fatal("the creature has no granted activated ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrifice was not paid")
	}
	if p, tough := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("pumped creature %d/%d, want 4/3", p, tough)
	}
	// Into a graveyard from the battlefield: back to its owner's hand.
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, bear); err != nil {
		t.Fatalf("MoveCardByID: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ideal) {
		t.Error("Fallen Ideal did not return to its owner's hand")
	}
}

// The ability is the creature's: its controller activates it, not the
// Aura's (the ruling).
func TestSocB5FallenIdealOnAnOpponentsCreatureIsTheirs(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", testCreatureTypeLine, 2, 2)
	mine := b12Creature(g, me.ID, "Mine", testCreatureTypeLine, 1, 1)
	castCatalogSpell(t, g, "Fallen Ideal", auraTypeLine, socB5FallenIdeal, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	idx, ref := grantedActivatedIndex(t, g, theirs)
	if idx < 0 {
		t.Fatal("the creature has no granted activated ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, theirs, idx, game.ActivateAbilityParams{Ref: ref, SacrificeIDs: []uuid.UUID{mine}}); err == nil {
		t.Error("the Aura's controller activated the creature's ability")
	}
}

// --- Feral Appetite -------------------------------------------------------

func TestSocB5FeralAppetitePumpsOnlyAttackingPests(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Feral Appetite", "Enchantment", socB5FeralAppetite, false)
	g.WithWriteLock(func() { _, _ = g.CreateTokensForEffect(me.ID, PestToken(), 1, game.TokenEntryOptions{}) })
	pest := socB5Named(g, me.ID, "Pest")[0]
	bear := b12Creature(g, me.ID, "Bear", testCreatureTypeLine, 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == pest {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if p := effectivePower(t, g, pest); p != 1 {
		t.Errorf("an idle Pest has power %d, want 1", p)
	}
	advanceToDeclareAttackersOf(t, g, 0)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: pest, Target: opp.ID}, {Attacker: bear, Target: opp.ID}}); err != nil {
		t.Fatalf("attack: %v", err)
	}
	if p := effectivePower(t, g, pest); p != 2 {
		t.Errorf("an attacking Pest has power %d, want 2", p)
	}
	if !containsString(effectiveAbilities(t, g, pest), "deathtouch") {
		t.Error("the attacking Pest lacks deathtouch")
	}
	if p := effectivePower(t, g, bear); p != 2 || containsString(effectiveAbilities(t, g, bear), "deathtouch") {
		t.Error("a non-Pest attacker was pumped")
	}
}

func TestSocB5FeralAppetiteMakesAPestOnlyForACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	appetite := pushCatalogPermanent(g, me.ID, "Feral Appetite", "Enchantment", socB5FeralAppetite, false)
	advanceToMain(t, g)
	rock := pushGraveyardPermanent(opp, "Rock", "Artifact", "{1}")
	bear := pushGraveyardPermanent(opp, "Bear", "Creature — Bear", "{1}{G}")
	for _, id := range []uuid.UUID{rock, bear} {
		if err := g.ActivateCatalogAbility(me.ID, appetite, 0, game.ActivateAbilityParams{Targets: b16TargetCard(id)}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !g.Exile.Contains(id) {
			t.Errorf("%s was not exiled", id)
		}
		want := 0
		if id == bear {
			want = 1
		}
		if n := len(socB5Named(g, me.ID, "Pest")); n != want {
			t.Errorf("after exiling %s: %d Pests, want %d", id, n, want)
		}
	}
}

// --- Festering Thicket ----------------------------------------------------

func TestSocB5FesteringThicketEntersTappedTapsForBlackOrGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := playLandFromHand(t, g, "Festering Thicket", socB5FesteringThicket)
	if c, ok := battlefieldCard(g, id); !ok || !c.Tapped {
		t.Fatal("the land did not enter tapped")
	}
	for _, colour := range []string{"B", "G"} {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(id) })
		me.ManaPool = nil
		b28TapForMana(t, g, me.ID, id, colour)
		if got := poolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("pool = %v, want {%s}", got, colour)
		}
	}
}

func TestSocB5FesteringThicketCycles(t *testing.T) {
	g := newCatalogGame(t)
	id, me := cycleFromHand(t, g, "Festering Thicket", "Land — Swamp Forest", socB5FesteringThicket, "{C}{C}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled land is not in the graveyard")
	}
	before := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != before+1 {
		t.Errorf("hand %d -> %d, want the cycling draw", before, got)
	}
}

// --- Forum Filibuster -----------------------------------------------------

func TestSocB5ForumFilibusterReturnsAnAuraAttachedToTheInkling(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Forum Filibuster", "Enchantment", socB5ForumFilibuster, false)
	rancor := pushGraveyardPermanent(me, "Rancor", auraTypeLine, "{G}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == rancor {
				me.Graveyard.Cards[i].OracleID = rancorOracle
			}
		}
	})
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	inklings := socB5Named(g, me.ID, "Inkling")
	if len(inklings) != 1 {
		t.Fatalf("Inklings: %d, want 1", len(inklings))
	}
	ink := layeredCard(t, g, inklings[0])
	if ink.Effective().Power != 2 || ink.Effective().Toughness != 1 || !slices.Contains(ink.Effective().Abilities, "flying") {
		t.Errorf("Inkling %d/%d %v, want a 2/1 flier", ink.Effective().Power, ink.Effective().Toughness, ink.Effective().Abilities)
	}
	// The reflexive trigger targets as it goes on the stack.
	pickTriggerTarget(t, g, me.ID, rancor)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rancor) {
		t.Fatal("Rancor did not return")
	}
	if host := attachmentHostOf(t, g, rancor); host.ID != inklings[0] {
		t.Errorf("Rancor is attached to %+v, want the Inkling", host)
	}
	if p := effectivePower(t, g, inklings[0]); p != 4 {
		t.Errorf("enchanted Inkling power %d, want 4", p)
	}
}

func TestSocB5ForumFilibusterReturnsAnEquipmentAndMayReturnNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Forum Filibuster", "Enchantment", socB5ForumFilibuster, false)
	sword := pushGraveyardPermanent(me, "Sword", "Artifact — Equipment", "{1}")
	bear := pushGraveyardPermanent(me, "Bear", "Creature — Bear", "{1}{G}")
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if slices.Contains(pick.PickTargetCards, bear) {
		t.Error("a creature card is offered")
	}
	pickTriggerTarget(t, g, me.ID, sword)
	passPriorityAroundTable(t, g)
	inklings := socB5Named(g, me.ID, "Inkling")
	if len(inklings) != 1 || !g.Battlefield.Contains(sword) {
		t.Fatalf("Inklings %d, Equipment returned %v", len(inklings), g.Battlefield.Contains(sword))
	}
	if host := attachmentHostOf(t, g, sword); host.ID != inklings[0] {
		t.Errorf("the Equipment is attached to %+v, want the Inkling", host)
	}

	// Next upkeep: "up to one", so the controller may choose nothing.
	advanceToUpkeepOf(t, g, 1)
	advanceToUpkeepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if pick := latestPickTarget(g, me.ID); pick != nil {
		if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
			t.Fatalf("choose no target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if n := len(socB5Named(g, me.ID, "Inkling")); n != 2 {
		t.Errorf("Inklings %d, want 2", n)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the creature card left the graveyard")
	}
}

// --- Fractal Harness ------------------------------------------------------

func TestSocB5FractalHarnessMakesAnXFractalAndDoublesItOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harness, err := socB5CastX(t, g, "Fractal Harness", "Artifact — Equipment", socB5FractalHarness, "{X}{2}{G}", 3, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	fractals := socB5Named(g, me.ID, "Fractal")
	if len(fractals) != 1 {
		t.Fatalf("Fractals: %d, want 1", len(fractals))
	}
	fractal := fractals[0]
	if n := countersOn(g, fractal, game.CounterPlusOne); n != 3 {
		t.Errorf("Fractal has %d counters, want 3", n)
	}
	if host := attachmentHostOf(t, g, harness); host.ID != fractal {
		t.Errorf("the Harness is attached to %+v, want the Fractal", host)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == fractal {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	advanceToDeclareAttackersOf(t, g, 0)
	if err := g.DeclareAttacker(fractal, opp.ID); err != nil {
		t.Fatalf("attack: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, fractal, game.CounterPlusOne); n != 6 {
		t.Errorf("after the attack the Fractal has %d counters, want 6", n)
	}
}

func TestSocB5FractalHarnessForZeroMakesAFractalThatDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	harness, err := socB5CastX(t, g, "Fractal Harness", "Artifact — Equipment", socB5FractalHarness, "{X}{2}{G}", 0, nil)
	if err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := len(socB5Named(g, me.ID, "Fractal")); n != 0 {
		t.Errorf("a 0/0 Fractal survived: %d", n)
	}
	if !g.Battlefield.Contains(harness) {
		t.Fatal("the Harness left the battlefield")
	}
	if host := attachmentHostOf(t, g, harness); host.ID != uuid.Nil {
		t.Errorf("the Harness is still attached to %+v", host)
	}
}
