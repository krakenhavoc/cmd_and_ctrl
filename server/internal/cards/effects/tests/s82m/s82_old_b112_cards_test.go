package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b112_cards_test.go — Ikoria: Lair of Behemoths, batch 112 (#5512).
// Each mutate trigger is driven through the engine (ADR 0223).

type o112Card struct {
	oracle, name, typeLine, cost, mutate string
	power, toughness                     int
	keywords                             []string
}

var (
	o112Lurker     = o112Card{"88aa5cd0-811a-4682-b2b0-2decbd13b332", "Boneyard Lurker", "Creature — Nightmare Beast", "{2}{B}{G}", "{2}{B/G}{B/G}", 4, 4, []string{"mutate"}}
	o112Cubwarden  = o112Card{"8c1b3f1b-40d8-4cd5-bedf-92c3c534be84", "Cubwarden", "Creature — Cat", "{3}{W}", "{2}{W}{W}", 3, 5, []string{"lifelink", "mutate"}}
	o112Harvester  = o112Card{"9473b445-294b-4572-8267-f1f2fc2a05a7", "Chittering Harvester", "Creature — Nightmare", "{5}{B}", "{4}{B}", 4, 6, []string{"mutate"}}
	o112Archipel   = o112Card{"ae1df2ad-f34c-4569-92f2-98be776c612d", "Archipelagore", "Creature — Leviathan", "{5}{U}{U}", "{5}{U}", 7, 7, []string{"mutate"}}
	o112Snapdax    = o112Card{"b2b78623-5a10-4ef7-b983-3d82478269c3", "Snapdax, Apex of the Hunt", "Legendary Creature — Dinosaur Cat Nightmare", "{1}{R}{W}{B}", "{2}{B/R}{W}{W}", 3, 5, []string{"double strike", "mutate"}}
	o112Whisperer  = o112Card{"b42147e1-cf90-4c12-8abe-7f2931d0607c", "Cavern Whisperer", "Creature — Nightmare", "{4}{B}", "{3}{B}", 4, 4, []string{"menace", "mutate"}}
	o112Parcel     = o112Card{"de83b309-ac31-4a57-95e6-ccd3659e6c1f", "Parcelbeast", "Creature — Elemental Beast", "{2}{G}{U}", "{G}{U}", 2, 4, []string{"mutate"}}
	o112Porcuparr  = o112Card{"e1efd99f-d918-4cd1-b85f-b90b106c3c88", "Porcuparrot", "Creature — Bird Beast", "{3}{R}", "{2}{R}", 3, 4, []string{"mutate"}}
	o112Brokkos    = o112Card{"e9137b54-3464-4f57-95e3-f94bf2ac854e", "Brokkos, Apex of Forever", "Legendary Creature — Nightmare Beast Elemental", "{2}{B}{G}{U}", "{2}{U/B}{G}{G}", 6, 6, []string{"trample", "mutate"}}
	o112Cloudpierc = o112Card{"ed7885bf-c557-4328-b1a4-54780b2315f4", "Cloudpiercer", "Creature — Dinosaur", "{4}{R}", "{3}{R}", 5, 4, []string{"reach", "mutate"}}
	o112Leosaur    = o112Card{"fc445b1d-229d-4cf3-8203-8872d8957a15", "Regal Leosaur", "Creature — Dinosaur Cat", "{R}{W}", "{1}{R/W}{R/W}", 2, 2, []string{"mutate"}}
)

var o112All = []o112Card{o112Lurker, o112Cubwarden, o112Harvester, o112Archipel, o112Snapdax, o112Whisperer, o112Parcel, o112Porcuparr, o112Brokkos, o112Cloudpierc, o112Leosaur}

func TestS82OldB112EveryCardIsFullWithItsMutateOffer(t *testing.T) {
	for _, tc := range o112All {
		spec, ok := effects.Lookup(tc.oracle)
		if !ok || spec.Name != tc.name {
			t.Errorf("%s is not registered: %+v", tc.name, spec.Name)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: want Full with no caveat; got %v %q", tc.name, spec.Completeness, spec.Caveats)
		}
		alt := game.AlternativeCostByKey(tc.oracle, game.MutateKey)
		if alt == nil || alt.ManaCost != tc.mutate || alt.Targets == nil || !alt.Targets.OwnedBySourceOwner {
			t.Errorf("%s: mutate offer %+v, want %s targeting a creature with the spell's owner", tc.name, alt, tc.mutate)
		}
	}
	grave := game.AlternativeCostByKey(o112Brokkos.oracle, game.MutateGraveyardKey)
	hand := game.AlternativeCostByKey(o112Brokkos.oracle, game.MutateKey)
	if grave == nil || grave.FromZone != game.ZoneGraveyard || grave.ManaCost != o112Brokkos.mutate || hand == nil || hand.FromZone != game.ZoneHand {
		t.Errorf("Brokkos's offers: hand %+v graveyard %+v", hand, grave)
	}
}

func (c o112Card) card() game.Card {
	return game.Card{Name: c.name, TypeLine: c.typeLine, OracleID: c.oracle, ManaCost: c.cost,
		Keywords: c.keywords, Power: c.power, Toughness: c.toughness}
}

func o112Mutate(t *testing.T, g *game.Game, me *game.Player, c o112Card, host uuid.UUID, under bool) uuid.UUID {
	t.Helper()
	id := effectstest.RprMutatorInHand(t, g, me.ID, c.card())
	if err := effectstest.CastMutating(g, me.ID, id, host, false); err != nil {
		t.Fatalf("cast %s mutating: %v", c.name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	idx := 0
	if under {
		idx = 1
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	return id
}

func o112Ref(id uuid.UUID) *game.TargetRef { return &game.TargetRef{Kind: game.TargetCard, ID: id} }

func o112Opp(g *game.Game) *game.Player { return g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)] }

func o112Perm(g *game.Game, owner uuid.UUID, name, typeLine string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: p, Toughness: tough, Owner: owner, Controller: owner})
}

func o112Grave(p *game.Player, name, typeLine, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func o112Zone(g *game.Game, id uuid.UUID) game.ZoneKind {
	if z := g.FindCardZoneForEffect(id); z != nil {
		return z.Kind
	}
	return ""
}

// --- Boneyard Lurker --------------------------------------------------------------------

func TestS82OldB112BoneyardLurkerReturnsAPermanentCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := o112Grave(me, "Old Forest", "Basic Land — Forest", "")
	spell := o112Grave(me, "Old Bolt", "Instant", "{R}")
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Lurker, elves, false)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil || effectstest.HasID(pick.PickTargetCards, spell) || !effectstest.HasID(pick.PickTargetCards, land) {
		t.Fatalf("pick %+v: want only the permanent card on offer", pick)
	}
	effectstest.AssistCardSettle(t, g, me.ID, o112Ref(land), true)
	if o112Zone(g, land) != game.ZoneHand || o112Zone(g, spell) != game.ZoneGraveyard {
		t.Errorf("land in %v, instant in %v", o112Zone(g, land), o112Zone(g, spell))
	}
}

// --- Cubwarden --------------------------------------------------------------------------

func TestS82OldB112CubwardenMakesTwoLifelinkCats(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Cubwarden, elves, true)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	cats := effectstest.TokensNamed(g, "Cat")
	if len(cats) != 2 {
		t.Fatalf("%d Cat tokens, want 2", len(cats))
	}
	for _, c := range cats {
		if p, tough := effectstest.PtOf(t, g, c.InstanceID); p != 1 || tough != 1 || !effectstest.HasKeywordOf(g, c.InstanceID, "lifelink") {
			t.Errorf("a Cat is %d/%d, lifelink %v", p, tough, effectstest.HasKeywordOf(g, c.InstanceID, "lifelink"))
		}
	}
	if !effectstest.HasKeywordOf(g, elves, "lifelink") {
		t.Error("the merged creature lacks lifelink")
	}
}

// --- Chittering Harvester ---------------------------------------------------------------

func TestS82OldB112ChitteringHarvesterMakesEachOpponentSacrificeOfTheirChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	mine := o112Perm(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	a := o112Perm(g, opp.ID, "Opp A", "Creature — Bear", 5, 5)
	b := o112Perm(g, opp.ID, "Opp B", "Creature — Bear", 1, 1)
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Harvester, elves, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SacrificeChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerSacrifice(t, g, opp.ID, b)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o112Zone(g, b) != game.ZoneGraveyard || o112Zone(g, a) != game.ZoneBattlefield {
		t.Errorf("the chosen creature is in %v and the other in %v", o112Zone(g, b), o112Zone(g, a))
	}
	if o112Zone(g, mine) != game.ZoneBattlefield {
		t.Error("my own creature was sacrificed")
	}
}

// --- Archipelagore ----------------------------------------------------------------------

func TestS82OldB112ArchipelagoreTapsAndFreezesUpToXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	a := o112Perm(g, opp.ID, "Opp A", "Creature — Bear", 2, 2)
	b := o112Perm(g, opp.ID, "Opp B", "Creature — Bear", 2, 2)
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Archipel, elves, false)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if pick.PickTargetMax != 1 {
		t.Errorf("up to %d targets after one mutation, want 1", pick.PickTargetMax)
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a, b)); err == nil {
		t.Fatal("two targets were accepted for X = 1")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, effectstest.CardRefs(a)); err != nil {
		t.Fatalf("ResolvePickTargets: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c, _ := effectstest.AangCardOnBF(g, a); !c.Tapped {
		t.Error("the target was not tapped")
	}
	if c, _ := effectstest.AangCardOnBF(g, b); c.Tapped {
		t.Error("an untargeted creature was tapped")
	}
	// It stays tapped through its controller's next untap step.
	effectstest.AdvanceToNextSeatsTurn(t, g)
	for i := 0; i < 8 && g.Turn.ActiveSeat != opp.Seat; i++ {
		effectstest.AdvanceToNextSeatsTurn(t, g)
	}
	effectstest.AdvanceToMain(t, g)
	if c, _ := effectstest.AangCardOnBF(g, a); !c.Tapped {
		t.Error("the frozen creature untapped in its controller's untap step")
	}
}

// --- Snapdax, Apex of the Hunt ----------------------------------------------------------

func TestS82OldB112SnapdaxDealsFourAndGainsFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	mine := o112Perm(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	victim := o112Perm(g, opp.ID, "Victim", "Creature — Bear", 4, 4)
	elves := effectstest.LlanowarOn(g, me.ID)
	life := me.Life
	o112Mutate(t, g, me, o112Snapdax, elves, true)
	pick := effectstest.LatestPickTarget(g, me.ID)
	if pick == nil || effectstest.HasID(pick.PickTargetCards, mine) {
		t.Fatalf("pick %+v: want only the opponent's permanents", pick)
	}
	effectstest.AssistCardSettle(t, g, me.ID, o112Ref(victim), true)
	if o112Zone(g, victim) != game.ZoneGraveyard {
		t.Errorf("the 4/4 is in %v, want the graveyard", o112Zone(g, victim))
	}
	if me.Life != life+4 {
		t.Errorf("life %d, want %d", me.Life, life+4)
	}
	if !effectstest.HasKeywordOf(g, elves, "double strike") {
		t.Error("the merged creature lacks double strike")
	}
}

// --- Cavern Whisperer -------------------------------------------------------------------

func TestS82OldB112CavernWhispererMakesEachOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	card := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.Cards = nil
		opp.Hand.PushTop(game.Card{InstanceID: card, Name: "Only Card", TypeLine: "Sorcery", Owner: opp.ID, Controller: opp.ID})
	})
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Whisperer, elves, false)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DiscardChoiceFor(g, opp.ID) != nil {
		effectstest.AnswerDiscard(t, g, opp.ID, card)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if o112Zone(g, card) != game.ZoneGraveyard {
		t.Errorf("the card is in %v, want the graveyard", o112Zone(g, card))
	}
}

// --- Parcelbeast ------------------------------------------------------------------------

func o112ParcelbeastReady(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Parcel, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		me.ManaPool.EmptyPool()
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elves {
				g.Battlefield.Cards[i].SummonedThisTurn = false
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	effectstest.FloatForTest(g, me, "G")
	return elves
}

func TestS82OldB112ParcelbeastPutsALandOntoTheBattlefieldOrIntoYourHand(t *testing.T) {
	for _, tc := range []struct {
		name     string
		typeLine string
		accept   bool
		want     game.ZoneKind
	}{
		{"Top Land", "Basic Land — Forest", true, game.ZoneBattlefield},
		{"Top Land", "Basic Land — Forest", false, game.ZoneHand},
		{"Top Spell", "Sorcery", false, game.ZoneHand},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		elves := o112ParcelbeastReady(t, g, me)
		top := effectstest.PushLibraryTyped(me, tc.name, tc.typeLine)
		if err := g.ActivateCatalogAbility(me.ID, elves, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("%s: activate: %v", tc.name, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if tc.typeLine != "Sorcery" {
			if tc.accept {
				effectstest.AnswerChooseCards(t, g, me.ID, top)
			} else {
				effectstest.AnswerChooseCards(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		got := o112Zone(g, top)
		if tc.want == game.ZoneBattlefield {
			if _, ok := effectstest.AangCardOnBF(g, top); !ok {
				// The card is a new object on the battlefield; look it up by name.
				found := false
				for _, c := range g.Battlefield.Cards {
					if c.Name == tc.name && !c.Tapped {
						found = true
					}
				}
				if !found {
					t.Errorf("%s accept=%v: the land is not on the battlefield untapped", tc.name, tc.accept)
				}
			}
			continue
		}
		if got != tc.want {
			t.Errorf("%s accept=%v: the card is in %v, want %v", tc.name, tc.accept, got, tc.want)
		}
	}
}

// --- Porcuparrot ------------------------------------------------------------------------

func TestS82OldB112PorcuparrotDealsXDamageForTheTimesItHasMutated(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Porcuparr, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == elves {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, elves, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d (X = 1 after one mutation)", opp.Life, life-1)
	}
}

// --- Brokkos, Apex of Forever -----------------------------------------------------------

func TestS82OldB112BrokkosMutatesFromTheGraveyardAndOnlyThatWay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	effectstest.AdvanceToMain(t, g)
	// Plenty of mana, so each refusal below is the cast path's and not a short pool.
	effectstest.FloatForTest(g, me, "UGGGGGGGGG")
	brokkos := uuid.New()
	c := o112Brokkos.card()
	c.InstanceID, c.Owner, c.Controller = brokkos, me.ID, me.ID
	g.WithWriteLock(func() { me.Graveyard.PushTop(c) })
	target := []game.TargetRef{{Kind: game.TargetCard, ID: elves}}
	// For its mana cost, or with the hand's offer, it cannot be cast from the graveyard.
	if err := g.CastSpell(me.ID, brokkos, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true}); err == nil {
		t.Fatal("Brokkos was cast from the graveyard for its mana cost")
	}
	if err := g.CastSpell(me.ID, brokkos, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true,
		AlternativeCost: game.MutateKey, Targets: target}); err == nil {
		t.Fatal("the hand's mutate offer was claimed from the graveyard")
	}
	if err := g.CastSpell(me.ID, brokkos, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), Strict: true,
		AlternativeCost: game.MutateGraveyardKey, Targets: target}); err != nil {
		t.Fatalf("mutate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOptionPick(t, g, me.ID, 0)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	f := effectstest.ReadMerged(t, g, elves)
	if f.Name != "Brokkos, Apex of Forever" || len(f.Components) != 2 || f.Power != 6 {
		t.Errorf("the Elves did not mutate into Brokkos: %+v", f)
	}
	if !effectstest.HasKeywordOf(g, elves, "trample") {
		t.Error("the merged creature lacks trample")
	}
}

func TestS82OldB112BrokkosFromHandMutatesAndRefusesTheGraveyardOffer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	elves := effectstest.LlanowarOn(g, me.ID)
	id := effectstest.RprMutatorInHand(t, g, me.ID, o112Brokkos.card())
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "UGGGGGGGGG")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{AlternativeCost: game.MutateGraveyardKey, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: elves}}}); err == nil {
		t.Fatal("the graveyard's mutate offer was claimed from the hand")
	}
	if err := effectstest.CastMutating(g, me.ID, id, elves, true); err != nil {
		t.Fatalf("mutate from the hand: %v", err)
	}
}

// --- Cloudpiercer -----------------------------------------------------------------------

func TestS82OldB112CloudpiercerLootsOnlyIfYouDiscard(t *testing.T) {
	for _, discard := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		elves := effectstest.LlanowarOn(g, me.ID)
		extra := uuid.New()
		g.WithWriteLock(func() {
			me.Hand.PushTop(game.Card{InstanceID: extra, Name: "Spare", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		})
		effectstest.SeedLibrary(me, "A", "B", "C")
		o112Mutate(t, g, me, o112Cloudpierc, elves, false)
		effectstest.PassPriorityAroundTable(t, g)
		hand := me.Hand.Size()
		if discard {
			effectstest.AnswerDiscard(t, g, me.ID, extra)
		} else {
			effectstest.AnswerDiscard(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if discard && (o112Zone(g, extra) != game.ZoneGraveyard || me.Hand.Size() != hand) {
			t.Errorf("discard: card in %v, hand %d -> %d (want the same size after discard and draw)", o112Zone(g, extra), hand, me.Hand.Size())
		}
		if !discard && (o112Zone(g, extra) != game.ZoneHand || me.Hand.Size() != hand) {
			t.Errorf("decline: card in %v, hand %d -> %d (want no change)", o112Zone(g, extra), hand, me.Hand.Size())
		}
	}
}

// --- Regal Leosaur ----------------------------------------------------------------------

func TestS82OldB112RegalLeosaurPumpsTheOthersOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], o112Opp(g)
	bear := o112Perm(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := o112Perm(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	elves := effectstest.LlanowarOn(g, me.ID)
	o112Mutate(t, g, me, o112Leosaur, elves, false)
	effectstest.AssistCardSettle(t, g, me.ID, nil, true)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 3 {
		t.Errorf("my Bear is %d/%d, want 4/3", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their Bear is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectstest.PtOf(t, g, elves); p != 2 || tough != 2 {
		t.Errorf("the Leosaur is %d/%d, want its own 2/2", p, tough)
	}
}
