package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// msc_scarlet_quantum_cards_test.go — Scarlet Witch, Chaotic Avenger and
// Quantum Reduction (Marvel Super Heroes Commander, #4455), the two cards
// that needed a small engine extension each: the filter's "Hero or
// noncreature" and the optional cost that opens instant speed.

const (
	mscSqScarletWitchOracle    = "8c72d54e-97a6-408e-ae77-18d3c96b3bd8"
	mscSqQuantumReductionOracl = "4e18d20f-f153-4607-9201-3e22591474db"
)

func TestMscSqCardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		mscSqScarletWitchOracle:    "Scarlet Witch, Chaotic Avenger",
		mscSqQuantumReductionOracl: "Quantum Reduction",
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

// --- Scarlet Witch ---------------------------------------------------------

type mscSqWitchTable struct {
	g       *game.Game
	me, opp *game.Player
	witch   uuid.UUID
}

func mscSqWitch(t *testing.T) mscSqWitchTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	witch := b12Push(g, me.ID, "Scarlet Witch, Chaotic Avenger", "Legendary Creature — Mutant Warlock Hero", mscSqScarletWitchOracle, 3, 3)
	advanceToMain(t, g)
	return mscSqWitchTable{g: g, me: me, opp: opp, witch: witch}
}

// hit makes Scarlet Witch deal combat damage to the opponent and runs the
// trigger up to the resolution_cast offer it raises.
func (w mscSqWitchTable) hit(t *testing.T) {
	t.Helper()
	dealCombatDamageToPlayer(w.g, w.witch, w.opp.ID, 3)
	passPriorityUntilChoice(t, w.g)
}

var (
	mscSqHeroCreature = game.Card{Name: "Hero Creature", TypeLine: "Creature — Human Hero", ManaCost: "{3}{W}", Power: 3, Toughness: 3}
	mscSqBear         = game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2}
	mscSqSorcery      = game.Card{Name: "Big Sorcery", TypeLine: "Sorcery", ManaCost: "{5}{R}"}
	mscSqInstant      = game.Card{Name: "Fast Instant", TypeLine: "Instant", ManaCost: "{2}{U}"}
	mscSqForest       = game.Card{Name: "A Forest", TypeLine: "Basic Land — Forest"}
)

func mscSqExiledWithWitch(g *game.Game, witch uuid.UUID) map[uuid.UUID]bool {
	out := map[uuid.UUID]bool{}
	g.ReadSnapshot(func() {
		for _, c := range g.Exile.Cards {
			if c.ExiledWith.ID == witch {
				out[c.InstanceID] = true
			}
		}
	})
	return out
}

func mscSqInExile(g *game.Game, id uuid.UUID) bool {
	var in bool
	g.ReadSnapshot(func() { in = g.Exile.Contains(id) })
	return in
}

func TestMscSqScarletWitchExilesTheTopTwoFaceDownAndCastsAHeroFree(t *testing.T) {
	w := mscSqWitch(t)
	ids := sunbirdLibrary(w.me, mscSqHeroCreature, mscSqBear, mscSqForest)
	hero, bear := ids[0], ids[1]
	before := w.me.Library.Size()
	w.hit(t)

	linked := mscSqExiledWithWitch(w.g, w.witch)
	if len(linked) != 2 || !linked[hero] || !linked[bear] {
		t.Fatalf("exiled with Scarlet Witch = %v, want the top two cards", linked)
	}
	w.g.ReadSnapshot(func() {
		for _, c := range w.g.Exile.Cards {
			if linked[c.InstanceID] && (!c.FaceDown || c.FaceDownKind != game.FaceDownHidden) {
				t.Errorf("%s was exiled face up (%v, %v), want face down", c.Name, c.FaceDown, c.FaceDownKind)
			}
		}
		if got := w.me.Library.Size(); got != before-2 {
			t.Errorf("library size %d, want the two exiled cards gone from %d", got, before)
		}
	})

	offer := mustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, hero)
	if hasID(offer.DiscardOptions, bear) {
		t.Errorf("a creature that is not a Hero was offered: %v", offer.DiscardOptions)
	}

	// Strict: a non-Hero creature from her exile is refused.
	if err := w.g.CastSpell(w.me.ID, bear, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("a non-Hero creature was cast from among the cards exiled with Scarlet Witch")
	}
	if !mscSqInExile(w.g, bear) {
		t.Error("the refused cast moved the card")
	}

	// The Hero creature is cast for no mana, inside the resolution.
	if err := castThroughResolutionCast(t, w.g, w.me.ID, hero, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the Hero creature: %v", err)
	}
	castInsideTheResolution(t, w.g, hero)
	passPriorityAroundTable(t, w.g)
	if !w.g.Battlefield.Contains(hero) {
		t.Error("the free Hero creature should resolve onto the battlefield")
	}
	if !mscSqInExile(w.g, bear) {
		t.Error("the uncast card should stay in exile")
	}
}

func TestMscSqScarletWitchCastsANoncreatureSpellAndNeverALand(t *testing.T) {
	w := mscSqWitch(t)
	ids := sunbirdLibrary(w.me, mscSqForest, mscSqSorcery, mscSqBear)
	land, sorcery := ids[0], ids[1]
	w.hit(t)

	offer := mustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, sorcery)
	if hasID(offer.DiscardOptions, land) {
		t.Errorf("a land was offered as a spell: %v", offer.DiscardOptions)
	}
	if err := w.g.CastSpell(w.me.ID, land, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("the land was played through a cast permission")
	}
	if err := castThroughResolutionCast(t, w.g, w.me.ID, sorcery, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of a six-mana sorcery: %v", err)
	}
	if offer := resolutionCastOfferFor(w.g, w.me.ID); offer != nil {
		t.Error("the offer outlived its one cast")
	}
}

// "From among cards exiled with Scarlet Witch" is every card she has
// exiled, not the two this hit exiled; and it is one cast.
func TestMscSqScarletWitchOffersEarlierExiledCardsAndOnlyOneCast(t *testing.T) {
	w := mscSqWitch(t)
	first := sunbirdLibrary(w.me, mscSqHeroCreature, mscSqBear)
	hero, bear := first[0], first[1]
	w.hit(t)
	mustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, hero)
	declineResolutionCast(t, w.g, w.me.ID)
	if !mscSqInExile(w.g, hero) {
		t.Fatal("declining the cast should leave the Hero in exile")
	}
	if err := w.g.CastSpell(w.me.ID, hero, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Fatal("the declined offer left the Hero castable")
	}

	second := sunbirdLibrary(w.me, mscSqInstant, mscSqForest)
	instant := second[0]
	w.hit(t)
	offer := mustResolutionCastOffer(t, w.g, w.me.ID, game.MayCastKeywordFreeCast, hero, instant)
	if hasID(offer.DiscardOptions, bear) {
		t.Errorf("the earlier non-Hero creature was offered: %v", offer.DiscardOptions)
	}
	if got := len(mscSqExiledWithWitch(w.g, w.witch)); got != 4 {
		t.Errorf("%d cards exiled with Scarlet Witch, want 4", got)
	}
	if err := castThroughResolutionCast(t, w.g, w.me.ID, hero, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("free cast of the earlier Hero: %v", err)
	}
	if err := w.g.CastSpell(w.me.ID, instant, game.CastSpellParams{FromZone: "exile", Strict: true}); err == nil {
		t.Error("a second spell was cast from the same offer")
	}
	if !mscSqInExile(w.g, instant) {
		t.Error("the uncast instant should stay in exile")
	}
}

func TestMscSqScarletWitchWithNothingToCastAsksNothing(t *testing.T) {
	w := mscSqWitch(t)
	sunbirdLibrary(w.me, mscSqBear, mscSqForest)
	dealCombatDamageToPlayer(w.g, w.witch, w.opp.ID, 3)
	passPriorityAroundTable(t, w.g)
	if resolutionCastOfferFor(w.g, w.me.ID) != nil {
		t.Error("an offer was raised with no Hero or noncreature spell among the cards")
	}
	if got := len(mscSqExiledWithWitch(w.g, w.witch)); got != 2 {
		t.Errorf("%d cards exiled with Scarlet Witch, want 2", got)
	}
}

func TestPermissionFilterNoncreatureOrCreatureType(t *testing.T) {
	f := game.PermissionFilter{NoncreatureOrCreatureType: "Hero"}
	for _, tc := range []struct {
		card game.Card
		want bool
	}{
		{mscSqHeroCreature, true},
		{mscSqBear, false},
		{mscSqSorcery, true},
		{mscSqInstant, true},
		{game.Card{Name: "Hero Aura", TypeLine: "Enchantment — Aura"}, true},
		{game.Card{Name: "Hero Artifact Creature", TypeLine: "Artifact Creature — Robot Hero"}, true},
	} {
		if got := f.Matches(tc.card); got != tc.want {
			t.Errorf("%s (%s) matches = %v, want %v", tc.card.Name, tc.card.TypeLine, got, tc.want)
		}
	}
	if !(game.PermissionFilter{}).Matches(mscSqBear) {
		t.Error("the zero filter should match everything")
	}
}

// --- Quantum Reduction -----------------------------------------------------

type mscSqAuraTable struct {
	g          *game.Game
	me         *game.Player
	aura       uuid.UUID
	target     uuid.UUID
	helper     uuid.UUID
	targetRefs []game.TargetRef
}

func mscSqQuantum(t *testing.T) mscSqAuraTable {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	target := pushVanillaCreature(g, opp.ID, "Big Bear", 5, 5)
	helper := pushVanillaCreature(g, me.ID, "Helper", 2, 2)
	aura := handCardFull(me, "Quantum Reduction", "Enchantment — Aura", "{1}{U}", mscSqQuantumReductionOracl, []string{"U"})
	return mscSqAuraTable{g: g, me: me, aura: aura, target: target, helper: helper,
		targetRefs: []game.TargetRef{{Kind: game.TargetCard, ID: target}}}
}

func (a mscSqAuraTable) cast(optional []int, team []uuid.UUID) error {
	b06AddMana(a.me, "U", "U")
	return a.g.CastSpell(a.me.ID, a.aura, game.CastSpellParams{
		Strict: true, Targets: a.targetRefs, OptionalCosts: optional, TeamworkIDs: team,
	})
}

func TestMscSqQuantumReductionHasFlashOnlyWhenCastUsingTeamwork(t *testing.T) {
	a := mscSqQuantum(t)
	advanceToEndStep(t, a.g)

	// Instant speed without teamwork: refused.
	if err := a.cast(nil, nil); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("an Aura without flash in an end step: got %v, want ErrSorcerySpeedRequired", err)
	}
	// Announcing the teamwork cost opens the window.
	if err := a.cast([]int{0}, []uuid.UUID{a.helper}); err != nil {
		t.Fatalf("cast using teamwork in an end step: %v", err)
	}
	if c := findBattlefieldCardByID(a.g, a.helper); c == nil || !c.Tapped {
		t.Error("the helper should be tapped to pay teamwork")
	}
	passPriorityAroundTable(t, a.g)
	if got := effectivePower(t, a.g, a.target); got != 0 {
		t.Errorf("enchanted creature has power %d, want 5 - 5 = 0", got)
	}
}

func TestMscSqQuantumReductionStaysASorceryWithoutTeamwork(t *testing.T) {
	a := mscSqQuantum(t)
	advanceToMain(t, a.g)
	if err := a.cast(nil, nil); err != nil {
		t.Fatalf("an ordinary cast in the main phase: %v", err)
	}
	passPriorityAroundTable(t, a.g)
	if got := effectivePower(t, a.g, a.target); got != 0 {
		t.Errorf("enchanted creature has power %d, want 0", got)
	}
}

// CR 101.2: "can't" beats "can". A restriction to sorcery timing closes
// the teamwork cast as it closes the keyword.
func TestMscSqQuantumReductionTeamworkFlashDoesNotBeatARestriction(t *testing.T) {
	a := mscSqQuantum(t)
	advanceToEndStep(t, a.g)
	a.g.WithWriteLock(func() {
		a.g.GrantCastTimingForEffect(a.me.ID, game.CastTimingRule{Timing: game.TimingSorcery},
			"You can cast spells only any time you could cast a sorcery.", uuid.Nil, a.g.UntilEndOfTurnDuration())
	})
	if err := a.cast([]int{0}, []uuid.UUID{a.helper}); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("teamwork flash under a sorcery-timing restriction: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// The enumerator agrees with the cast check: an end step offers only the
// teamwork cast, a main phase offers both, and with nothing to tap the end
// step offers nothing.
func TestMscSqQuantumReductionEnumeratorAgreesWithTheCast(t *testing.T) {
	moves := func(a mscSqAuraTable) (all, teamwork int) {
		for _, m := range legal.EnumerateFor(a.g, a.me.ID) {
			if m.Kind != legal.KindCast || m.Source != a.aura {
				continue
			}
			var p struct {
				OptionalCosts []int    `json:"optional_costs"`
				TeamworkIDs   []string `json:"teamwork_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatalf("move params: %v", err)
			}
			all++
			if len(p.OptionalCosts) > 0 && len(p.TeamworkIDs) > 0 {
				teamwork++
			}
		}
		return all, teamwork
	}

	a := mscSqQuantum(t)
	advanceToEndStep(t, a.g)
	b06AddMana(a.me, "U", "U")
	all, teamwork := moves(a)
	if all == 0 || all != teamwork {
		t.Errorf("end step moves: %d casts, %d with teamwork; want only teamwork casts", all, teamwork)
	}
	for _, m := range legal.EnumerateFor(a.g, a.me.ID) {
		if m.Kind == legal.KindCast && m.Source == a.aura {
			if err := actions.Dispatch(a.g, actions.Action{Type: actions.Type(m.Type), Player: a.me.ID, Caller: a.me.ID, Params: m.Params}); err != nil {
				t.Fatalf("the engine refused a cast the enumerator offered: %v", err)
			}
			break
		}
	}

	b := mscSqQuantum(t)
	advanceToEndStep(t, b.g)
	b06AddMana(b.me, "U", "U")
	b.g.WithWriteLock(func() {
		for i := range b.g.Battlefield.Cards {
			if b.g.Battlefield.Cards[i].InstanceID == b.helper {
				b.g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if all, _ := moves(b); all != 0 {
		t.Errorf("%d casts offered with no way to pay teamwork at instant speed", all)
	}

	c := mscSqQuantum(t)
	advanceToMain(t, c.g)
	b06AddMana(c.me, "U", "U")
	if all, teamwork := moves(c); all < 2 || teamwork == 0 || teamwork == all {
		t.Errorf("main phase moves: %d casts, %d with teamwork; want the plain cast and the teamwork one", all, teamwork)
	}
}

// The view's castable_here, for a card cast out of exile, agrees as well.
func TestMscSqQuantumReductionExileViewAgreesWithTheCast(t *testing.T) {
	setup := func() (mscSqAuraTable, uuid.UUID) {
		a := mscSqQuantum(t)
		id := uuid.New()
		c := game.Card{
			InstanceID: id, Name: "Quantum Reduction", TypeLine: "Enchantment — Aura", ManaCost: "{1}{U}",
			OracleID: mscSqQuantumReductionOracl, Colors: []string{"U"}, Owner: a.me.ID, Controller: a.me.ID,
			KnownBy: map[uuid.UUID]bool{a.me.ID: true},
		}
		a.g.Exile.PushTop(c)
		a.g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
			Player: a.me.ID, Zone: game.ZoneExile, Duration: a.g.UntilEndOfTurnDuration(), Cost: "{0}", CastOnly: true,
			Label: "Test — cast it without paying its mana cost",
		})
		return a, id
	}
	castable := func(a mscSqAuraTable, id uuid.UUID) bool {
		v := protocol.ViewOfGameFor(a.g, a.me.ID.String())
		for _, c := range v.Exile.Cards {
			if c.InstanceID == id.String() {
				return c.CastableHere
			}
		}
		t.Fatal("Quantum Reduction is not in the exile view")
		return false
	}

	a, id := setup()
	advanceToEndStep(t, a.g)
	if !castable(a, id) {
		t.Error("castable_here is false in an end step with a teamwork cast open")
	}
	params := game.CastSpellParams{Strict: true, FromZone: "exile", Targets: a.targetRefs}
	if err := a.g.CastSpell(a.me.ID, id, params); err != game.ErrSorcerySpeedRequired {
		t.Fatalf("the plain cast from exile in an end step: got %v, want ErrSorcerySpeedRequired", err)
	}
	params.OptionalCosts, params.TeamworkIDs = []int{0}, []uuid.UUID{a.helper}
	if err := a.g.CastSpell(a.me.ID, id, params); err != nil {
		t.Fatalf("the teamwork cast from exile in an end step: %v", err)
	}

	b, id := setup()
	advanceToEndStep(t, b.g)
	b.g.WithWriteLock(func() {
		for i := range b.g.Battlefield.Cards {
			if b.g.Battlefield.Cards[i].InstanceID == b.helper {
				b.g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if castable(b, id) {
		t.Error("castable_here is true with no creature to pay teamwork")
	}
}
