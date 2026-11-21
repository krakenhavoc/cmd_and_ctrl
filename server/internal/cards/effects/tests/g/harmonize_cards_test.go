package testsg

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestHarmonizeCardsAreFullAndDeclareHarmonize(t *testing.T) {
	for oracle, cost := range map[string]string{
		effectstest.NaturesRhythmOracle:      "{X}{G}{G}{G}{G}",
		effectstest.SynchronizedChargeOracle: "{4}{G}",
		effectstest.ChanneledDragonfireOracl: "{5}{R}{R}",
		effectstest.MammothBellowOracle:      "{5}{G}{U}{R}",
		effectstest.WinternightStoriesOracle: "{4}{U}",
		effectstest.UrenisRebuffOracle:       "{5}{U}",
		effectstest.ZenithFestivalOracle:     "{X}{R}{R}",
		effectstest.GlacialDragonhuntOracle:  "{4}{U}{R}",
		effectstest.WildRideOracle:           "{4}{R}",
		effectstest.RoamersRoutineOracle:     "{4}{G}",
		effectstest.UnendingWhisperOracle:    "{5}{U}",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		offers := game.AlternativeCostsFor(oracle)
		if len(offers) != 1 {
			t.Errorf("%s offers %d alternative costs, want 1", spec.Name, len(offers))
			continue
		}
		o := offers[0]
		if o.Key != effects.AltCostKeyHarmonize || o.ManaCost != cost || o.FromZone != game.ZoneGraveyard ||
			!o.ExileOnLeavingStack || !o.ReducedByTappedPower || o.TapOthers == nil || !o.TapOthers.UpTo ||
			o.TapOthers.Count != 1 {
			t.Errorf("%s offers %+v, want harmonize %s from the graveyard", spec.Name, o, cost)
		}
	}
	if spec, ok := effects.Lookup(effectstest.SongcrafterMageOracle); !ok || spec.Completeness != effects.CompletenessFull {
		t.Error("Songcrafter Mage is not registered Full")
	}
}

// castHarmonize casts `card` from its owner's graveyard for its harmonize
// cost, tapping `tap` (none, or one creature), paying mana through the
// auto-tapper.
func castHarmonize(g *game.Game, p *game.Player, card uuid.UUID, x int, targets []game.TargetRef, tap ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: effects.AltCostKeyHarmonize, AltCostIDs: tap,
		XValue: x, Strict: true, AutoTap: true, Targets: targets,
	})
}

// harmonizePrice is what the engine's one pricer charges for `card` cast
// with harmonize, tapping `tap`.
func harmonizePrice(t *testing.T, g *game.Game, p *game.Player, card uuid.UUID, x int, tap ...uuid.UUID) string {
	t.Helper()
	var out string
	g.WithWriteLock(func() {
		c, ok := g.LookupCardForEffect(card)
		if !ok {
			t.Fatal("the harmonize card is gone")
		}
		price, err := g.PriceCastForEffect(p.ID, c, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: effects.AltCostKeyHarmonize, AltCostIDs: tap, XValue: x,
		})
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		out = price.Total.String()
	})
	return out
}

// CR 702.180a: reduced by the tapped creature's power, generic mana only,
// never below {0}; tapping none reduces nothing; a counter counts; a
// creature with no power saves nothing.
func TestHarmonizeIsPricedByTheTappedCreaturesPower(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	fire := effectstest.GraveyardCardOf(g, me, "Channeled Dragonfire", "Sorcery", "{R}", effectstest.ChanneledDragonfireOracl)
	three := effectstest.Pr7Creature(g, me.ID, "Three", 3, "G")
	nine := effectstest.Pr7Creature(g, me.ID, "Nine", 9, "G")
	zero := effectstest.Pr7Creature(g, me.ID, "Zero", 0, "G")
	pumped := effectstest.Pr7Creature(g, me.ID, "Pumped", 1, "G")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, pumped).Counters = map[string]int{game.CounterPlusOne: 2}
		g.RecomputeLayersIfStaleLocked()
	})
	for name, c := range map[string]struct {
		tap  []uuid.UUID
		want string
	}{
		"nothing tapped":          {nil, "{5}{R}{R}"},
		"a 3-power creature":      {[]uuid.UUID{three}, "{2}{R}{R}"},
		"a 9-power creature":      {[]uuid.UUID{nine}, "{R}{R}"},
		"a 0-power creature":      {[]uuid.UUID{zero}, "{5}{R}{R}"},
		"a 1/x with two counters": {[]uuid.UUID{pumped}, "{2}{R}{R}"},
	} {
		if got := harmonizePrice(t, g, me, fire, 0, c.tap...); got != c.want {
			t.Errorf("%s: harmonize costs %s, want %s", name, got, c.want)
		}
	}
	// CR 601.2f, #2701: with {X} in the harmonize cost, the reduction comes
	// off the announced X once there is no printed generic left.
	rhythm := effectstest.GraveyardCardOf(g, me, "Nature's Rhythm", "Sorcery", "{X}{G}{G}", effectstest.NaturesRhythmOracle)
	g.WithWriteLock(func() {
		c, _ := g.LookupCardForEffect(rhythm)
		price, err := g.PriceCastForEffect(me.ID, c, game.CastSpellParams{
			FromZone: "graveyard", AlternativeCost: effects.AltCostKeyHarmonize, AltCostIDs: []uuid.UUID{three}, XValue: 5,
		})
		if err != nil {
			t.Fatalf("PriceCastForEffect: %v", err)
		}
		if price.Total.XReduced != 3 || price.Total.Generic != 0 {
			t.Errorf("Nature's Rhythm at X=5 tapping a 3-power creature: %+v, want 3 off the X", price.Total)
		}
	})
}

// The cast: only from the graveyard; the payment refuses a tapped creature,
// an opponent's, two creatures and a land; a creature that arrived this
// turn and a hexproof one may pay (CR 302.6, 601.2h); it is tapped with
// the spell on the stack; and the spell is exiled when it resolves.
func TestHarmonizeCastsFromTheGraveyardTapsAndExiles(t *testing.T) {
	g, me, opp := effectstest.HarmonizeTable(t)
	whisper := effectstest.GraveyardCardOf(g, me, "Unending Whisper", "Sorcery", "{U}", effectstest.UnendingWhisperOracle)
	for i := 0; i < 3; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	fresh := effectstest.Pr7Creature(g, me.ID, "Fresh", 3, "G")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, fresh)
		c.SummonedThisTurn = true
		c.Keywords = []string{"hexproof"}
	})
	already := effectstest.Pr7Creature(g, me.ID, "Already tapped", 5, "G")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, already).Tapped = true })
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 5, "G")
	other := effectstest.Pr7Creature(g, me.ID, "Other", 1, "G")
	land := effectstest.PushLandFor(g, me.ID, "Wastes", "Basic Land — Wastes")

	inHand := effectstest.HandCardOf(g, me, "Unending Whisper", "Sorcery", "{U}", effectstest.UnendingWhisperOracle)
	if err := g.CastSpell(me.ID, inHand, game.CastSpellParams{AlternativeCost: effects.AltCostKeyHarmonize, AltCostIDs: []uuid.UUID{fresh}}); err == nil {
		t.Fatal("harmonize was claimed from the hand")
	}
	for name, ids := range map[string][]uuid.UUID{
		"an already tapped creature": {already},
		"an opponent's creature":     {theirs},
		"two creatures":              {fresh, other},
		"a land":                     {land},
	} {
		if err := castHarmonize(g, me, whisper, 0, nil, ids...); err == nil {
			t.Fatalf("%s paid harmonize", name)
		}
	}
	// {5}{U} with three Islands and a 1-power creature ({4}{U}) is short.
	if err := castHarmonize(g, me, whisper, 0, nil, other); err == nil {
		t.Fatal("harmonize was cast for {4}{U} with three Islands")
	}
	if effectstest.TappedForTest(t, g, fresh) || effectstest.TappedForTest(t, g, other) || !me.Graveyard.Contains(whisper) {
		t.Fatal("a refused cast paid something")
	}
	if err := castHarmonize(g, me, whisper, 0, nil, fresh); err != nil {
		t.Fatalf("harmonize tapping a 3-power creature that arrived this turn: %v", err)
	}
	if g.Stack.Size() != 1 || !effectstest.TappedForTest(t, g, fresh) || effectstest.TappedForTest(t, g, other) {
		t.Fatal("the creature was not tapped, alone, with the spell on the stack")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after Unending Whisper, want %d", me.Hand.Size(), hand+1)
	}
	if !g.Exile.Contains(whisper) || me.Graveyard.Contains(whisper) {
		t.Error("a harmonized spell was not exiled as it left the stack (CR 702.180a)")
	}
}

// Tapping nothing is a legal payment of the full cost.
func TestHarmonizeMayTapNothing(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	whisper := effectstest.GraveyardCardOf(g, me, "Unending Whisper", "Sorcery", "{U}", effectstest.UnendingWhisperOracle)
	for i := 0; i < 6; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	if err := castHarmonize(g, me, whisper, 0, nil); err != nil {
		t.Fatalf("harmonize for {5}{U} tapping nothing: %v", err)
	}
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("tapping nothing tapped a creature")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(whisper) {
		t.Error("the spell was not exiled")
	}
}

// A spell cast with harmonize that is countered on resolution for want of
// a target is exiled, not put into the graveyard to be cast again.
func TestHarmonizedSpellThatFizzlesIsExiled(t *testing.T) {
	g, me, opp := effectstest.HarmonizeTable(t)
	rebuff := effectstest.GraveyardCardOf(g, me, "Ureni's Rebuff", "Sorcery", "{1}{U}", effectstest.UrenisRebuffOracle)
	for i := 0; i < 6; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2, "B")
	if err := castHarmonize(g, me, rebuff, 0, []game.TargetRef{{Kind: game.TargetCard, ID: victim}}); err != nil {
		t.Fatalf("Ureni's Rebuff with harmonize: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(victim); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(rebuff) || me.Graveyard.Contains(rebuff) {
		t.Error("a harmonized spell that fizzled was not exiled")
	}
}

// The creature named to harmonize is tapped for the cost, so the
// auto-tapper never taps it for the spell's mana (CR 601.2g-h).
func TestHarmonizeTappedCreatureCannotAlsoMakeMana(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	whisper := effectstest.GraveyardCardOf(g, me, "Unending Whisper", "Sorcery", "{U}", effectstest.UnendingWhisperOracle)
	elf := effectstest.PushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", effectstest.LlanowarElvesOracle, false)
	for i := 0; i < 4; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	// {5}{U} less the Elf's 1 is {4}{U}: four Islands are short, and the
	// Elf can't make the fifth mana.
	if err := castHarmonize(g, me, whisper, 0, nil, elf); err == nil {
		t.Fatal("the creature tapped for harmonize also paid mana")
	}
	if effectstest.TappedForTest(t, g, elf) {
		t.Error("a refused cast tapped the Elf")
	}
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	if err := castHarmonize(g, me, whisper, 0, nil, elf); err != nil {
		t.Fatalf("harmonize for {4}{U} with five Islands: %v", err)
	}
	if !effectstest.TappedForTest(t, g, elf) {
		t.Error("the Elf was not tapped for harmonize")
	}
}

// harmonizeOfferOnTheWire is the harmonize offer the view stamps on a card
// in the viewer's graveyard.
func harmonizeOfferOnTheWire(t *testing.T, g *game.Game, me *game.Player, card uuid.UUID) *protocol.AlternativeCostView {
	t.Helper()
	v := protocol.ViewOfGameFor(g, me.ID.String())
	for _, s := range v.Seats {
		for i := range s.Graveyard.Cards {
			if s.Graveyard.Cards[i].InstanceID != card.String() {
				continue
			}
			for j := range s.Graveyard.Cards[i].AlternativeCosts {
				if s.Graveyard.Cards[i].AlternativeCosts[j].Key == effects.AltCostKeyHarmonize {
					return &s.Graveyard.Cards[i].AlternativeCosts[j]
				}
			}
		}
	}
	return nil
}

// The wire: tap_options min 0 max 1 (none is a payment), the price per
// candidate, and the offer stays with no creature to tap.
func TestHarmonizeOfferOnTheWireIsPricedPerCandidate(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	fire := effectstest.GraveyardCardOf(g, me, "Channeled Dragonfire", "Sorcery", "{R}", effectstest.ChanneledDragonfireOracl)
	offer := harmonizeOfferOnTheWire(t, g, me, fire)
	if offer == nil {
		t.Fatal("the harmonize offer is not on the wire with no creature to tap")
	}
	three := effectstest.Pr7Creature(g, me.ID, "Three", 3, "R")
	zero := effectstest.Pr7Creature(g, me.ID, "Zero", 0, "R")
	offer = harmonizeOfferOnTheWire(t, g, me, fire)
	if offer == nil || offer.TapOptions == nil {
		t.Fatalf("offer %+v, want a tap offer", offer)
	}
	if !offer.ReducesByPower || offer.TapOptions.Min != 0 || offer.TapOptions.Max != 1 || len(offer.TapOptions.Cards) != 2 {
		t.Fatalf("offer %+v / tap options %+v, want min 0, max 1 over two creatures, reduces by power", offer, offer.TapOptions)
	}
	want := map[uuid.UUID]protocol.AltCostTapPriceView{
		three: {Power: 3, Price: "{2}{R}{R}"},
		zero:  {Power: 0, Price: "{5}{R}{R}"},
	}
	for id, w := range want {
		if got := offer.TapPrices[id.String()]; got != w {
			t.Errorf("price for %s: %+v, want %+v", id, got, w)
		}
	}
}

// The enumerator prices each tap on its own and offers only affordable
// payments that save mana, best first, then none when it is affordable.
func TestHarmonizeEnumeratorOffersOnlyAffordablePayments(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	whisper := effectstest.GraveyardCardOf(g, me, "Unending Whisper", "Sorcery", "{U}", effectstest.UnendingWhisperOracle)
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	one := effectstest.Pr7Creature(g, me.ID, "One", 1, "G")
	five := effectstest.Pr7Creature(g, me.ID, "Five", 5, "G")
	six := effectstest.Pr7Creature(g, me.ID, "Six", 6, "G")
	offered := func() (map[uuid.UUID]bool, bool) {
		out, none := map[uuid.UUID]bool{}, false
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != legal.TypeCastSpell || m.Source != whisper {
				continue
			}
			var p struct {
				AlternativeCost string      `json:"alternative_cost"`
				AltCostIDs      []uuid.UUID `json:"alt_cost_ids"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil || p.AlternativeCost != effects.AltCostKeyHarmonize {
				continue
			}
			if len(p.AltCostIDs) == 0 {
				none = true
			}
			for _, id := range p.AltCostIDs {
				out[id] = true
			}
		}
		return out, none
	}
	got, none := offered()
	if !got[five] || !got[six] || got[one] || none {
		t.Errorf("offered %v (none %v), want the five- and six-power creatures only", got, none)
	}
	for i := 0; i < 5; i++ {
		effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	}
	if _, none = offered(); !none {
		t.Error("with {5}{U} available, tapping nothing is not offered")
	}
}

// effects.Register holds the flag and "up to" together, over one creature.
func TestRegisterRefusesHarmonizeShapes(t *testing.T) {
	noFlag := effects.Harmonize("{4}")
	noFlag.ReducedByTappedPower = false
	noUpTo := effects.Harmonize("{4}")
	noUpTo.TapOthers = effects.TapOthersPrice(1, "an untapped creature you control", effects.Creature())
	two := effects.Harmonize("{4}")
	two.TapOthers.Count = 2
	for name, ac := range map[string]game.AlternativeCost{"no flag": noFlag, "no up to": noUpTo, "two": two} {
		msg := effectstest.RegisterPanics(effects.Spec{OracleID: "test-3931-harmonize-" + strings.ReplaceAll(name, " ", "-"), Name: "Test " + name,
			CastableZones: []game.ZoneKind{game.ZoneGraveyard}, AlternativeCosts: []game.AlternativeCost{ac}})
		if !strings.Contains(msg, "CR 702.180a") {
			t.Errorf("%s: panic %q, want the harmonize guard", name, msg)
		}
	}
	msg := effectstest.RegisterPanics(effects.Spec{OracleID: "test-3931-harmonize-additional", Name: "Test additional",
		AdditionalCost: &game.AdditionalCost{TapOthers: game.HarmonizeTapCost(), Label: "Tap"}})
	if !strings.Contains(msg, "#3931") {
		t.Errorf("an additional cost tapping up to one: panic %q, want the up-to guard", msg)
	}
}

// Nature's Rhythm: X is announced with the harmonize cost and the tapped
// creature's power pays for X; the search reads that X.
func TestNaturesRhythmHarmonizeFindsByTheAnnouncedX(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	rhythm := effectstest.GraveyardCardOf(g, me, "Nature's Rhythm", "Sorcery", "{X}{G}{G}", effectstest.NaturesRhythmOracle)
	for i := 0; i < 5; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	big := effectstest.Pr7Creature(g, me.ID, "Big", 4, "G")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Five-drop", TypeLine: "Creature — Test", ManaCost: "{4}{G}", Power: 5, Toughness: 5},
		game.Card{Name: "Six-drop", TypeLine: "Creature — Test", ManaCost: "{5}{G}", Power: 6, Toughness: 6},
	)
	// X=5 is {5}{G}{G}{G}{G}; the 4-power creature pays four of it.
	if err := castHarmonize(g, me, rhythm, 5, nil, big); err != nil {
		t.Fatalf("Nature's Rhythm, X=5, tapping a 4-power creature with five Forests: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if len(c.SearchCards) != 1 || c.SearchCards[0] != ids[0] {
		t.Fatalf("search offers %v, want the five-drop only", c.SearchCards)
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[0]) || !g.Exile.Contains(rhythm) {
		t.Error("the five-drop is not on the battlefield, or Nature's Rhythm is not exiled")
	}
}

// Zenith Festival exiles X with the window to the end of your next turn.
func TestZenithFestivalExilesXToPlay(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	fest := effectstest.GraveyardCardOf(g, me, "Zenith Festival", "Sorcery", "{X}{R}{R}", effectstest.ZenithFestivalOracle)
	for i := 0; i < 2; i++ {
		effectstest.PushLandFor(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	two := effectstest.Pr7Creature(g, me.ID, "Two", 2, "R")
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("A", "Instant"), effectstest.SearchTestLand("B", "Instant"), effectstest.SearchTestLand("C", "Instant"))
	if err := castHarmonize(g, me, fest, 2, nil, two); err != nil {
		t.Fatalf("Zenith Festival, X=2, tapping a 2-power creature with two Mountains: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Size() != 1 {
		t.Errorf("library %d after exiling two of three, want 1", me.Library.Size())
	}
	if !g.Exile.Contains(fest) {
		t.Error("Zenith Festival is not exiled")
	}
}

// Songcrafter Mage gives a card in your graveyard harmonize at its own
// mana cost, with the tap and the reduction, until end of turn.
func TestSongcrafterMageGrantsHarmonize(t *testing.T) {
	g, me, _ := effectstest.HarmonizeTable(t)
	div := effectstest.GraveyardCardOf(g, me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	effectstest.CastCatalogSpell(t, g, "Songcrafter Mage", "Creature — Human Bard", effectstest.SongcrafterMageOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, div)
	}
	effectstest.PassPriorityAroundTable(t, g)
	o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, div, game.ZoneGraveyard), effects.AltCostKeyHarmonize)
	if o == nil {
		t.Fatal("Divination did not gain harmonize")
	}
	if o.ManaCost != "{2}{U}" || !o.ReducedByTappedPower || o.TapOthers == nil || !o.TapOthers.UpTo || !o.ExileOnLeavingStack {
		t.Fatalf("granted offer %+v, want harmonize {2}{U} with the tap and the reduction", o)
	}
	effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	two := effectstest.Pr7Creature(g, me.ID, "Two", 2, "U")
	if err := castHarmonize(g, me, div, 0, nil, two); err != nil {
		t.Fatalf("Divination with granted harmonize for {U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(div) {
		t.Error("the granted harmonize did not exile the spell")
	}
}

// Synchronized Charge: the split, then vigilance and trample for every
// creature you control with a counter, including one not targeted.
func TestSynchronizedChargeDistributesAndGrants(t *testing.T) {
	g, me, opp := effectstest.HarmonizeTable(t)
	a := effectstest.Pr7Creature(g, me.ID, "A", 2, "G")
	b := effectstest.Pr7Creature(g, me.ID, "B", 2, "G")
	veteran := effectstest.Pr7Creature(g, me.ID, "Veteran", 2, "G")
	plain := effectstest.Pr7Creature(g, me.ID, "Plain", 2, "G")
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "G")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, veteran).Counters = map[string]int{"stun": 1}
		effectstest.FindBattlefieldCardForTest(g, theirs).Counters = map[string]int{game.CounterPlusOne: 1}
	})
	for i := 0; i < 2; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	charge := effectstest.HandCardOf(g, me, "Synchronized Charge", "Sorcery", "{1}{G}", effectstest.SynchronizedChargeOracle)
	if err := g.CastSpell(me.ID, charge, game.CastSpellParams{
		Targets: effectstest.CardRefs(a, b), Distribution: map[uuid.UUID]int{a: 1, b: 1}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("Synchronized Charge: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	has := func(id uuid.UUID, kw string) bool {
		var ok bool
		g.WithWriteLock(func() { ok = game.HasKeyword(effectstest.FindBattlefieldCardForTest(g, id), kw) })
		return ok
	}
	for name, id := range map[string]uuid.UUID{"A": a, "B": b, "the veteran": veteran} {
		if !has(id, "vigilance") || !has(id, "trample") {
			t.Errorf("%s lacks vigilance and trample", name)
		}
	}
	if has(plain, "trample") || has(theirs, "trample") {
		t.Error("a creature without a counter, or an opponent's, gained trample")
	}
	if !me.Graveyard.Contains(charge) {
		t.Error("cast from hand, Synchronized Charge did not go to the graveyard")
	}
}

// Glacial Dragonhunt: a nonland discard makes the reflexive trigger, a
// land or nothing does not.
func TestGlacialDragonhuntDamagesOnANonlandDiscard(t *testing.T) {
	for _, c := range []struct {
		name    string
		discard string // "" declines
		hits    bool
	}{
		{"a nonland card", "Instant", true},
		{"a land", "Basic Land — Island", false},
		{"nothing", "", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g, me, opp := effectstest.HarmonizeTable(t)
			hunt := effectstest.HandCardOf(g, me, "Glacial Dragonhunt", "Sorcery", "{U}{R}", effectstest.GlacialDragonhuntOracle)
			var pitch uuid.UUID
			if c.discard != "" {
				pitch = effectstest.HandCardOf(g, me, "Pitch", c.discard, "", "")
			}
			victim := effectstest.Pr7Creature(g, opp.ID, "Victim", 2, "B")
			effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Drawn", "Instant"))
			if err := g.CastSpell(me.ID, hunt, game.CastSpellParams{}); err != nil {
				t.Fatalf("Glacial Dragonhunt: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if c.discard != "" {
				effectstest.AnswerDiscard(t, g, me.ID, pitch)
			} else {
				effectstest.AnswerDiscard(t, g, me.ID)
			}
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				effectstest.PickCard(t, g, me.ID, victim)
				effectstest.PassPriorityAroundTable(t, g)
			}
			var damage int
			g.WithWriteLock(func() {
				if v := effectstest.FindBattlefieldCardForTest(g, victim); v != nil {
					damage = v.DamageMarked
				} else {
					damage = 3
				}
			})
			if c.hits != (damage == 3) {
				t.Errorf("damage %d on the creature, want hit=%v", damage, c.hits)
			}
		})
	}
}

// The rest resolve as printed when cast from hand.
func TestHarmonizeSpellsResolveFromHand(t *testing.T) {
	g, me, opp := effectstest.HarmonizeTable(t)
	before := len(g.Battlefield.Cards)
	effectstest.CastCatalogSpell(t, g, "Mammoth Bellow", "Sorcery", effectstest.MammothBellowOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var elephant *game.Card
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].Name == "Elephant" {
			elephant = &g.Battlefield.Cards[i]
		}
	}
	if len(g.Battlefield.Cards) != before+1 || elephant == nil || elephant.Power != 5 || elephant.Toughness != 5 {
		t.Fatal("Mammoth Bellow did not make a 5/5 Elephant")
	}
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "R")
	effectstest.CastCatalogSpell(t, g, "Wild Ride", "Sorcery", effectstest.WildRideOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, bear)
		if c.CurrentPower() != 5 || !game.HasKeyword(c, "haste") {
			t.Errorf("Wild Ride: power %d, haste %v; want 5 and haste", c.CurrentPower(), game.HasKeyword(c, "haste"))
		}
	})
	theirs := effectstest.Pr7Creature(g, opp.ID, "Theirs", 2, "B")
	effectstest.CastCatalogSpell(t, g, "Ureni's Rebuff", "Sorcery", effectstest.UrenisRebuffOracle, []game.TargetRef{{Kind: game.TargetCard, ID: theirs}})
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Error("Ureni's Rebuff did not return the creature to its owner's hand")
	}
	hand := me.Hand.Size()
	effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("A", "Instant"), effectstest.SearchTestLand("B", "Creature — Test"), effectstest.SearchTestLand("C", "Instant"))
	effectstest.CastCatalogSpell(t, g, "Winternight Stories", "Sorcery", effectstest.WinternightStoriesOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	var creature uuid.UUID
	for _, c := range me.Hand.Cards {
		if c.Name == "B" {
			creature = c.InstanceID
		}
	}
	effectstest.AnswerDiscard(t, g, me.ID, creature)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d after drawing three and discarding one creature card, want %d", me.Hand.Size(), hand+2)
	}
}
