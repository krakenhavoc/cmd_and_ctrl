package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mkm_std_b42_cards_test.go — Murders at Karlov Manor cards, Standard
// batch 42 (#4452). Each test plays the real catalog entry.

const (
	stdB42SlimeOracle    = "b53597f4-1a0f-4fa8-9c17-29178cdc4d2b"
	stdB42TipsterOracle  = "8d0bf685-e521-4d72-acc6-86794f685ce2"
	stdB42ReconOracle    = "3fbe46fa-88da-439a-994e-c8f1791332a8"
	stdB42CreeperOracle  = "d5b7c283-357b-409f-9afc-985a3b89cc40"
	stdB42ExplosionOracl = "6e3b3a3f-e32f-4c72-a9fb-dd1e877876b6"
	stdB42JudithOracle   = "e86c965b-ac10-4fff-b682-dddd6d9747c6"
	stdB42KyloxOracle    = "37f6bbb8-2136-4f03-966c-85d72180e71f"
	stdB42LeylineOracle  = "6b2e6a51-b9e6-4848-b7d7-98f02f2b897f"
	stdB42ScoundrelOracl = "d69cb63f-98cc-4699-a03d-4f1beeeb7065"
	stdB42ReliveOracle   = "dcc51bcb-a61e-471f-8a01-7fcbc069fcba"
)

func TestStdB42EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		stdB42SlimeOracle:    "Slime Against Humanity",
		stdB42TipsterOracle:  "Tunnel Tipster",
		stdB42ReconOracle:    "Undergrowth Recon",
		stdB42CreeperOracle:  "Vengeful Creeper",
		stdB42ExplosionOracl: "Ill-Timed Explosion",
		stdB42JudithOracle:   "Judith, Carnage Connoisseur",
		stdB42KyloxOracle:    "Kylox, Visionary Inventor",
		stdB42LeylineOracle:  "Leyline of the Guildpact",
		stdB42ScoundrelOracl: "Rakish Scoundrel",
		stdB42ReliveOracle:   "Relive the Past",
	} {
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

// stdB42ToGraveyard puts a card into a player's graveyard.
func stdB42ToGraveyard(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// --- Relive the Past -----------------------------------------------------------

func TestStdB42ReliveThePastReturnsThreeAsFiveFiveElementals(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	art := stdB42ToGraveyard(g, me, "Old Trinket", "Artifact")
	land := stdB42ToGraveyard(g, me, "Old Forest", "Basic Land — Forest")
	ench := stdB42ToGraveyard(g, me, "Old Rite", "Enchantment")
	if _, err := mscM4TryCast(t, g, "Relive the Past", "Sorcery", stdB42ReliveOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: art, Slot: 0},
		{Kind: game.TargetCard, ID: land, Slot: 1},
		{Kind: game.TargetCard, ID: ench, Slot: 2},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"artifact": art, "land": land, "enchantment": ench} {
		c := tmtB1Card(t, g, id)
		if !c.IsCreature() || !c.HasSubtype("Elemental") {
			t.Errorf("the returned %s is not an Elemental creature: %q", name, c.TypeLine)
		}
		if p, tough := mkmMPPT(t, g, id); p != 5 || tough != 5 {
			t.Errorf("the returned %s is %d/%d, want 5/5", name, p, tough)
		}
	}
	if !tmtB1Card(t, g, art).IsArtifact() || !tmtB1Card(t, g, land).IsLand() {
		t.Error("a returned permanent lost its own types")
	}
}

func TestStdB42ReliveThePastClausesAreEachOptionalAndTypeChecked(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := stdB42ToGraveyard(g, me, "Old Forest", "Basic Land — Forest")
	aura := stdB42ToGraveyard(g, me, "Old Aura", "Enchantment — Aura")
	if _, err := mscM4TryCast(t, g, "Relive the Past", "Sorcery", stdB42ReliveOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: aura, Slot: 2},
	}); err == nil {
		t.Fatal("an Aura was accepted as the non-Aura enchantment")
	}
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	land = stdB42ToGraveyard(g, me, "Old Forest", "Basic Land — Forest")
	if _, err := mscM4TryCast(t, g, "Relive the Past", "Sorcery", stdB42ReliveOracle, nil, []game.TargetRef{
		{Kind: game.TargetCard, ID: land, Slot: 1},
	}); err != nil {
		t.Fatalf("CastSpell with only the land clause: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := mkmMPPT(t, g, land); p != 5 || tough != 5 {
		t.Errorf("the land is %d/%d, want 5/5", p, tough)
	}
}

// --- Slime Against Humanity ----------------------------------------------------

func TestStdB42SlimeAgainstHumanitySizesItselfOffOozesAndItsNamesakes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	stdB42ToGraveyard(g, me, "Gloop", "Creature — Ooze")
	stdB42ToGraveyard(g, me, "Slime Against Humanity", "Sorcery")
	stdB42ToGraveyard(g, me, "Bear", "Creature — Bear")
	stdB42ToGraveyard(g, opp, "Their Ooze", "Creature — Ooze") // not mine
	g.WithWriteLock(func() {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Exiled Ooze", TypeLine: "Creature — Ooze", Owner: me.ID, Controller: me.ID})
	})
	// One exiled Ooze of mine, one face down (not counted).
	exiled := uuid.New()
	hidden := uuid.New()
	g.WithWriteLock(func() {
		g.Exile.Cards = append(g.Exile.Cards,
			game.Card{InstanceID: exiled, Name: "Exiled Ooze", TypeLine: "Creature — Ooze", Owner: me.ID, Controller: me.ID},
			game.Card{InstanceID: hidden, Name: "Hidden Ooze", TypeLine: "Creature — Ooze", Owner: me.ID, Controller: me.ID, FaceDown: true})
	})
	if _, err := mscM4TryCast(t, g, "Slime Against Humanity", "Sorcery", stdB42SlimeOracle, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	var ooze *game.Card
	g.ReadSnapshot(func() {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == me.ID && c.Name == "Ooze" {
				cc := c
				ooze = &cc
			}
		}
	})
	if ooze == nil {
		t.Fatal("no Ooze token was created")
	}
	// 2 + graveyard Ooze + graveyard Slime + exiled face-up Ooze = 5; the
	// spell itself was on the stack when counted.
	if got := ooze.CurrentPower(); got != 5 {
		t.Errorf("the Ooze is %d/%d, want 5/5", got, ooze.CurrentToughness())
	}
	if !game.HasKeyword(ooze, "trample") {
		t.Error("the Ooze has no trample")
	}
}

// --- Undergrowth Recon ---------------------------------------------------------

func TestStdB42UndergrowthReconReturnsALandTappedAtUpkeep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Undergrowth Recon", "Enchantment", stdB42ReconOracle, false)
	land := stdB42ToGraveyard(g, me, "Old Forest", "Basic Land — Forest")
	stdB42ToGraveyard(g, me, "Old Bear", "Creature — Bear")
	advanceToUpkeepOfSeat(t, g, 1)
	advanceToUpkeepOfSeat(t, g, 0)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	c := tmtB1Card(t, g, land)
	if !c.Tapped {
		t.Error("the returned land is untapped")
	}
	if me.Graveyard.Contains(land) {
		t.Error("the land is still in the graveyard")
	}
}

// --- Leyline of the Guildpact --------------------------------------------------

func TestStdB42LeylineOfTheGuildpactMakesNonlandsAllColorsAndLandsEveryBasicType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	pushCatalogPermanent(g, me.ID, "Leyline of the Guildpact", "Enchantment", stdB42LeylineOracle, false)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := pushCatalogPermanent(g, me.ID, "Mystery Land", "Land", "", false)
	theirLand := pushCatalogPermanent(g, opp.ID, "Their Land", "Land", "", false)
	if c := layeredCard(t, g, bear).Effective(); len(c.Colors) != 5 {
		t.Errorf("my nonland permanent has colours %v, want all five", c.Colors)
	}
	if c := layeredCard(t, g, theirs).Effective(); len(c.Colors) != 0 {
		t.Errorf("an opponent's creature has colours %v", c.Colors)
	}
	for _, st := range []string{"Plains", "Island", "Swamp", "Mountain", "Forest"} {
		if !keywordSliceContains(layeredCard(t, g, land).Effective().Subtypes, st) {
			t.Errorf("my land is not a %s", st)
		}
	}
	if keywordSliceContains(layeredCard(t, g, theirLand).Effective().Subtypes, "Forest") {
		t.Error("an opponent's land gained a basic land type")
	}
}

// --- Vengeful Creeper ----------------------------------------------------------

func TestStdB42VengefulCreeperDestroysAnOpponentsArtifactWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	rock := pushCatalogPermanent(g, opp.ID, "Their Rock", "Artifact", "", false)
	mine := pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	id := mkmMPCastFaceDown(t, g, "Vengeful Creeper", "Creature — Plant Elemental", stdB42CreeperOracle, 5, 5)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Creeper")
	}
	mkmMPTurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C", "C")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(rock) {
		t.Error("the opponent's artifact was not destroyed")
	}
	if me.Graveyard.Contains(mine) {
		t.Error("my own artifact was destroyed")
	}
}

func TestStdB42VengefulCreeperCannotTargetYourOwnPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	pushCatalogPermanent(g, me.ID, "My Rock", "Artifact", "", false)
	id := mkmMPCastFaceDown(t, g, "Vengeful Creeper", "Creature — Plant Elemental", stdB42CreeperOracle, 5, 5)
	mkmMPTurnFaceUp(t, g, me, id, "G", "C", "C", "C", "C", "C")
	passPriorityAroundTable(t, g)
	if latestPickTarget(g, me.ID) != nil {
		t.Error("the trigger asked for a target with only my own permanents on the table")
	}
}

// --- Rakish Scoundrel ----------------------------------------------------------

func TestStdB42RakishScoundrelGrantsIndestructibleOnEnteringAndWhenTurnedFaceUp(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Rakish Scoundrel", "Creature — Elf Rogue", stdB42ScoundrelOracl, 3, 3)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if b := layeredCard(t, g, bear); !game.HasKeyword(&b, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
}

func TestStdB42RakishScoundrelFaceDownDoesNotTriggerButTurningUpDoes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	id := mkmMPCastFaceDown(t, g, "Rakish Scoundrel", "Creature — Elf Rogue", stdB42ScoundrelOracl, 3, 3)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("entering face down triggered the Scoundrel")
	}
	mkmMPTurnFaceUp(t, g, me, id, "B", "G", "C", "C", "B", "G")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if b := layeredCard(t, g, bear); !game.HasKeyword(&b, "indestructible") {
		t.Error("the target did not gain indestructible")
	}
}

// --- Tunnel Tipster ------------------------------------------------------------

func TestStdB42TunnelTipsterTapsForGreen(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	id := b12Push(g, me.ID, "Tunnel Tipster", "Creature — Mole Scout", stdB42TipsterOracle, 1, 1)
	if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap for mana: %v", err)
	}
	if got := me.ManaPool.CountColor("G"); got != 1 {
		t.Errorf("{G} in pool = %d, want 1", got)
	}
}

func TestStdB42TunnelTipsterGrowsAfterAFaceDownEntryOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	tip := b12Push(g, me.ID, "Tunnel Tipster", "Creature — Mole Scout", stdB42TipsterOracle, 1, 1)
	mkmMPCastFaceDown(t, g, "Museum Nightwatch", "Creature — Centaur Soldier", mkmMPNightwatchOracle, 3, 2)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tip); got != 1 {
		t.Errorf("Tipster has %d +1/+1 counters after a face-down creature entered, want 1", got)
	}

	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	tip = b12Push(g, me.ID, "Tunnel Tipster", "Creature — Mole Scout", stdB42TipsterOracle, 1, 1)
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tip); got != 0 {
		t.Errorf("Tipster grew with no face-down entry: %d counters", got)
	}
}

func TestStdB42TunnelTipsterIgnoresAnOpponentsFaceDownEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	tip := b12Push(g, me.ID, "Tunnel Tipster", "Creature — Mole Scout", stdB42TipsterOracle, 1, 1)
	msc18Library(opp, 2)
	g.WithWriteLock(func() {
		if _, err := g.ManifestForEffect(opp.ID); err != nil {
			t.Fatalf("manifest: %v", err)
		}
	})
	advanceToEndStep(t, g)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, tip); got != 0 {
		t.Errorf("an opponent's face-down entry grew the Tipster: %d counters", got)
	}
}

// --- Ill-Timed Explosion -------------------------------------------------------

// stdB42ExplosionFixture seeds a hand of two discardable cards (mana
// values 3 and 1), a 2/2 and a 4/4 on the table and a library to draw
// from, then casts Ill-Timed Explosion and resolves it up to the "you
// may" prompt.
func stdB42ExplosionFixture(t *testing.T) (g *game.Game, me *game.Player, big, small, hi, lo uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	msc18Library(me, 6)
	big = b12Creature(g, me.ID, "Big", "Creature — Ox", 4, 4)
	small = b12Creature(g, me.ID, "Small", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for me.Hand.Size() > 0 {
			_, _ = me.Hand.PopTop()
		}
		hi = uuid.New()
		lo = uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: hi, Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Owner: me.ID, Controller: me.ID})
		me.Hand.PushTop(game.Card{InstanceID: lo, Name: "One Drop", TypeLine: "Creature — Bear", ManaCost: "{G}", Owner: me.ID, Controller: me.ID})
	})
	if _, err := mscM4TryCast(t, g, "Ill-Timed Explosion", "Sorcery", stdB42ExplosionOracl, nil, nil); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	return
}

func TestStdB42IllTimedExplosionDealsGreatestManaValueToEveryCreature(t *testing.T) {
	g, me, big, small, hi, lo := stdB42ExplosionFixture(t)
	if me.Hand.Size() != 4 {
		t.Fatalf("hand = %d after drawing two, want 4", me.Hand.Size())
	}
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no \"you may discard two cards\" prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	answerDiscard(t, g, me.ID, hi, lo)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(hi) || !me.Graveyard.Contains(lo) {
		t.Fatal("the two cards were not discarded")
	}
	if me.Graveyard.Contains(big) {
		t.Error("the 4/4 died to 3 damage")
	}
	if !me.Graveyard.Contains(small) {
		t.Error("the 2/2 survived 3 damage")
	}
	var dmg int
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(big)
		dmg = c.DamageMarked
	})
	if dmg != 3 {
		t.Errorf("the 4/4 has %d damage, want 3 (the greatest mana value discarded)", dmg)
	}
}

func TestStdB42IllTimedExplosionDecliningDealsNothing(t *testing.T) {
	g, me, big, small, _, _ := stdB42ExplosionFixture(t)
	ask := latestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(small) || me.Graveyard.Contains(big) {
		t.Error("declining the discard still dealt damage")
	}
	if me.Hand.Size() != 4 {
		t.Errorf("hand = %d, want 4 (two drawn, none discarded)", me.Hand.Size())
	}
}

// --- Judith, Carnage Connoisseur -----------------------------------------------

const stdB42BoltOracle = "4457ed35-7c10-48c8-9776-456485fdf070"

// stdB42JudithBolt casts Lightning Bolt at `target` with Judith out and
// answers the trigger's mode prompt with `mode`.
func stdB42JudithBolt(t *testing.T, g *game.Game, target uuid.UUID, mode int) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", stdB42BoltOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: target}})
	c := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if c == nil {
		t.Fatal("no mode prompt for Judith's trigger")
	}
	if err := g.ResolveModePick(c.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestStdB42JudithGivesTheSpellDeathtouchAndLifelink(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	b12Push(g, me.ID, "Judith, Carnage Connoisseur", "Legendary Creature — Human Shaman", stdB42JudithOracle, 3, 4)
	wall := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	life := me.Life
	stdB42JudithBolt(t, g, wall, 0)
	if !opp.Graveyard.Contains(wall) {
		t.Error("the 3 damage from a deathtouch spell did not destroy the 0/9")
	}
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d (lifelink on 3 damage)", me.Life, life+3)
	}
}

func TestStdB42JudithWithoutTheModeTheSpellIsPlain(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	b12Push(g, me.ID, "Judith, Carnage Connoisseur", "Legendary Creature — Human Shaman", stdB42JudithOracle, 3, 4)
	wall := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	life := me.Life
	stdB42JudithBolt(t, g, wall, 1)
	if opp.Graveyard.Contains(wall) {
		t.Error("the Bolt destroyed the 0/9 without deathtouch")
	}
	if me.Life != life {
		t.Errorf("life changed to %d without lifelink", me.Life)
	}
}

func TestStdB42JudithMakesAnImpThatBurnsEachOpponentWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := mkmMPOpponents(g)[0]
	b12Push(g, me.ID, "Judith, Carnage Connoisseur", "Legendary Creature — Human Shaman", stdB42JudithOracle, 3, 4)
	wall := b12Creature(g, opp.ID, "Big Wall", "Creature — Wall", 0, 9)
	stdB42JudithBolt(t, g, wall, 1)
	imp := findBattlefieldByName(g, "Imp")
	if imp == uuid.Nil {
		t.Fatal("no Imp token was created")
	}
	if p, tough := mkmMPPT(t, g, imp); p != 2 || tough != 2 {
		t.Errorf("the Imp is %d/%d, want 2/2", p, tough)
	}
	before := mkmMPLives(mkmMPOpponents(g))
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(imp, game.DestroyOptions{}); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	after := mkmMPLives(mkmMPOpponents(g))
	for i := range after {
		if after[i] != before[i]-2 {
			t.Errorf("opponent %d life = %d, want %d", i, after[i], before[i]-2)
		}
	}
}

// --- Kylox, Visionary Inventor -------------------------------------------------

func TestStdB42KyloxSacrificesForPowerExilesThatManyAndCastsOnlyInstantsAndSorceries(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kylox := pushDiesCreatureForTest(g, me.ID, "Kylox, Visionary Inventor", stdB42KyloxOracle, "Legendary Creature — Lizard Artificer", 4, 4)
	two := pushVanillaCreature(g, me.ID, "Two", 2, 2)
	three := pushVanillaCreature(g, me.ID, "Three", 3, 3)
	keep := pushVanillaCreature(g, me.ID, "Keeper", 1, 1)
	g.WithWriteLock(func() {
		for me.Library.Size() > 0 {
			_, _ = me.Library.PopTop()
		}
	})
	filler := func(name string) game.Card {
		return game.Card{InstanceID: uuid.New(), Name: name, TypeLine: "Land", Owner: me.ID, Controller: me.ID}
	}
	// Top of the library, first to last: sorcery, creature, instant, then
	// two lands, then one more card that must stay put (X is 5).
	spell := uuid.New()
	creature := uuid.New()
	instant := uuid.New()
	rest := uuid.New()
	g.WithWriteLock(func() {
		for _, c := range []game.Card{
			{InstanceID: spell, Name: "Free Sorcery", TypeLine: "Sorcery", ManaCost: "{3}{R}", Owner: me.ID, Controller: me.ID},
			{InstanceID: creature, Name: "Free Creature", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Owner: me.ID, Controller: me.ID},
			{InstanceID: instant, Name: "Free Instant", TypeLine: "Instant", ManaCost: "{U}", Owner: me.ID, Controller: me.ID},
			filler("Land A"), filler("Land B"),
			{InstanceID: rest, Name: "Untouched", TypeLine: "Land", Owner: me.ID, Controller: me.ID},
		} {
			me.Library.PushBottom(c)
		}
	})
	stdB42Attack(t, g, opp.ID, kylox)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID, two, three)
	for _, id := range []uuid.UUID{two, three} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("sacrificed creature %s is not in the graveyard", id)
		}
	}
	if !g.Battlefield.Contains(keep) || !g.Battlefield.Contains(kylox) {
		t.Fatal("an unchosen creature, or Kylox, was sacrificed")
	}
	if got := me.Library.Size(); got != 1 {
		t.Fatalf("library has %d cards, want 1 (5 exiled for total power 2+3)", got)
	}
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, spell, instant)
	if err := castThroughResolutionCast(t, g, me.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature card was cast from among the exiled cards")
	}
	if err := castThroughResolutionCast(t, g, me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the sorcery: %v", err)
	}
	// Any number: the instant is still on offer after the first cast.
	mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, instant)
	if err := castThroughResolutionCast(t, g, me.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the instant: %v", err)
	}
}

func TestStdB42KyloxMaySacrificeNothingAndExileNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kylox := pushDiesCreatureForTest(g, me.ID, "Kylox, Visionary Inventor", stdB42KyloxOracle, "Legendary Creature — Lizard Artificer", 4, 4)
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	library := me.Library.Size()
	stdB42Attack(t, g, opp.ID, kylox)
	passPriorityAroundTable(t, g)
	answerOwnPermanents(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("a creature was sacrificed though none was chosen")
	}
	if me.Library.Size() != library {
		t.Errorf("library %d -> %d: cards were exiled for total power 0", library, me.Library.Size())
	}
}

// stdB42Attack declares one attacker and locks the declaration in, leaving
// the attack trigger's prompt open.
func stdB42Attack(t *testing.T, g *game.Game, defender, attacker uuid.UUID) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
}
