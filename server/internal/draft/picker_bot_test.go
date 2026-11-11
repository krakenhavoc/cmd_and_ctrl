package draft

import (
	"math/rand/v2"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// botPick is the bot's pick from a pack of a 14-card round: the pick
// number is the one at which a pack of 14 is down to len(pack) cards.
func botPick(fx *botFixture, seed uint64, picked, pack []Card, mode Mode) []string {
	return BotPicker{Seed: seed, Rater: &fx.rater}.Pick(PickRequest{
		Pack: pack, Picked: picked, Mode: mode, Take: min(mode.PerPick(), len(pack)),
		Seat: 0, Seats: 8, Round: 1 + len(picked)/14, Pick: 1 + (14-len(pack))/mode.PerPick(),
	})
}

func names(pack []Card, ids []string) []string {
	var out []string
	for _, id := range ids {
		for _, c := range pack {
			if c.ID == id {
				out = append(out, c.Name)
			}
		}
	}
	return out
}

// fullPack pads a pack to 14 cards with basic lands, which no bot wants,
// so the pick reads the pack as a first pick of a round.
func fullPack(fx *botFixture, t testing.TB, cs []Card) []Card {
	for i := len(cs); i < 14; i++ {
		cs = append(cs, fx.card(t, "Swamp", "pad"+string(rune('a'+i))))
	}
	return cs
}

// With nothing drafted the bot takes the best card of the pack, whatever
// its colour and for every seed.
func TestTheBotTakesTheBestCardEarly(t *testing.T) {
	fx := loadBotFixture(t)
	pack := fullPack(fx, t, fx.cards(t, "c", "C2 Fixture White Cub", "Divination", "Grave Titan", "Giant Growth", "Shock", "Scroll Thief"))
	for seed := range uint64(50) {
		if got := names(pack, botPick(fx, seed, nil, pack, PickOne)); !reflect.DeepEqual(got, []string{"Grave Titan"}) {
			t.Fatalf("seed %d: the first pick is %v, want Grave Titan", seed, got)
		}
	}
	// And still, a few picks in: a bomb of another colour is taken over
	// a filler card of the bot's own.
	picked := fx.cards(t, "p", "Shock", "C2 Fixture Red Cub", "Lightning Strike")
	pack = fx.cards(t, "c", "Grave Titan", "C2 Fixture Red Scout", "Dragon Fodder", "Act of Treason",
		"Opt", "Cancel", "Duress", "Mind Rot", "Naturalize", "Unsummon", "Swamp")
	for seed := range uint64(50) {
		if got := names(pack, botPick(fx, seed, picked, pack, PickOne)); !reflect.DeepEqual(got, []string{"Grave Titan"}) {
			t.Fatalf("seed %d: pick four is %v, want Grave Titan", seed, got)
		}
	}
}

// The commitment to the colours taken grows with each pick and is firm
// after about the first pack: the same choice, a much better card of a
// third colour against a filler card of the bot's own, goes one way
// early and the other once a pack has been drafted.
func TestTheColourCommitmentGrowsAndIsFirmAfterAPack(t *testing.T) {
	fx := loadBotFixture(t)
	redWhite := []string{
		"Lightning Bolt", "Serra Angel", "Shock", "Pacifism", "C2 Fixture Red Cub", "C2 Fixture White Cub",
		"Lightning Strike", "Inspiring Overseer", "C2 Fixture Red Scout", "Youthful Knight",
		"Fire Elemental", "Savannah Lions", "Incinerate", "C2 Fixture White Sentry",
	}
	// Murder is black: a far better card than a red three-mana 3/2.
	offer := []string{"Murder", "C2 Fixture Red Scout"}
	lastOff := -1
	for n := 0; n <= len(redWhite); n++ {
		picked := fx.cards(t, "p", redWhite[:n]...)
		pack := fx.cards(t, "c", offer...)
		for i := len(pack); i < 14-n%14; i++ {
			pack = append(pack, fx.card(t, "Island", "pad"+string(rune('a'+i))))
		}
		pack = pack[:max(14-n%14, 2)]
		took := names(pack, botPick(fx, 7, picked, pack, PickOne))[0]
		if took == "Murder" {
			lastOff = n
		} else if n < 4 {
			t.Errorf("after %d picks the bot passed Murder for a filler card: it committed too early", n)
		}
	}
	if lastOff < 4 || lastOff > 12 {
		t.Errorf("the bot last took the off-colour Murder after %d picks, want between 4 and 12 (firm after about a pack of 14)", lastOff)
	}

	// The commitment itself: 0 with nothing drafted, rising with every
	// pick, 1 after a pack.
	prev := -1.0
	for n := 0; n <= 14; n++ {
		st := newPickState(rateAll(fx.rater, fx.cards(t, "p", redWhite[:n]...)), 14)
		if st.commit <= prev {
			t.Errorf("the commitment after %d picks is %.2f, not above %.2f", n, st.commit, prev)
		}
		prev = st.commit
	}
	if prev != 1 {
		t.Errorf("the commitment after a pack is %.2f, want 1", prev)
	}
	st := newPickState(rateAll(fx.rater, fx.cards(t, "p", redWhite...)), 14)
	if got := []int{st.first, st.second}; !reflect.DeepEqual(got, []int{colourIndex("W"), colourIndex("R")}) &&
		!reflect.DeepEqual(got, []int{colourIndex("R"), colourIndex("W")}) {
		t.Errorf("the bot's colours are %v, want white and red", got)
	}

	// Firm means firm: a committed bot takes its own filler over another
	// colour's bomb, and a colourless card is always its own.
	picked := fx.cards(t, "p", redWhite...)
	pack := fullPack(fx, t, fx.cards(t, "c", "Grave Titan", "C2 Fixture White Scout"))
	if got := names(pack, botPick(fx, 7, picked, pack, PickOne)); got[0] != "C2 Fixture White Scout" {
		t.Errorf("a committed red-white bot took %v over its own colour", got)
	}
	pack = fullPack(fx, t, fx.cards(t, "c", "Juggernaut", "Divination", "Giant Growth"))
	if got := names(pack, botPick(fx, 7, picked, pack, PickOne)); got[0] != "Juggernaut" {
		t.Errorf("a committed bot took %v over a colourless creature", got)
	}
}

// What the deck still needs moves a pick: a bot with removal and no
// creatures takes a creature over one more removal spell of about the
// same rating, and a bot with a full top end passes another six-drop.
func TestTheBotWeighsWhatItsDeckNeeds(t *testing.T) {
	fx := loadBotFixture(t)
	spells := fx.cards(t, "p", "Shock", "Lightning Strike", "Incinerate", "Pacifism", "Shock", "Lightning Strike",
		"Incinerate", "Pacifism", "Shock", "Lightning Bolt", "Krenko's Command", "Oblivion Ring", "Lightning Helix", "Lightning Helix",
		"Shock", "Incinerate")
	pack := fullPack(fx, t, fx.cards(t, "c", "Lightning Strike", "C2 Fixture Red Sentry"))
	if got := names(pack, botPick(fx, 3, spells, pack, PickOne)); got[0] != "C2 Fixture Red Sentry" {
		t.Errorf("a bot with sixteen spells and no creature took %v", got)
	}
	// The same pack to a bot that has its creatures: the better card.
	bodies := fx.cards(t, "p", "C2 Fixture Red Cub", "C2 Fixture White Cub", "C2 Fixture Red Scout", "C2 Fixture White Scout",
		"C2 Fixture Red Sentry", "C2 Fixture White Sentry", "C2 Fixture Red Brute", "C2 Fixture White Brute",
		"Savannah Lions", "Youthful Knight", "Inspiring Overseer", "Serra Angel", "Fire Elemental", "C2 Fixture Red Cub",
		"C2 Fixture White Cub", "Dawnwing Marshal")
	if got := names(pack, botPick(fx, 3, bodies, pack, PickOne)); got[0] != "Lightning Strike" {
		t.Errorf("a bot with sixteen creatures and no removal took %v", got)
	}

	// The curve: seven cards of five mana or more, and the next is
	// passed for a cheaper card it would otherwise rate lower.
	heavy := fx.cards(t, "p", "C2 Fixture Red Giant", "C2 Fixture Red Giant", "C2 Fixture White Giant", "C2 Fixture Red Guard",
		"C2 Fixture White Guard", "Serra Angel", "Fire Elemental", "Shock", "Pacifism", "Lightning Strike", "C2 Fixture Red Cub",
		"C2 Fixture White Cub", "Savannah Lions", "Youthful Knight", "C2 Fixture Red Brute", "C2 Fixture White Brute")
	pack = fullPack(fx, t, fx.cards(t, "c", "C2 Fixture White Giant", "C2 Fixture White Sentry"))
	if got := names(pack, botPick(fx, 3, heavy, pack, PickOne)); got[0] != "C2 Fixture White Sentry" {
		t.Errorf("a bot with seven cards of five mana or more took %v", got)
	}
}

// In pick two the bot takes the best pair, not the best card twice. A
// red bot with no second colour yet is shown the best black card, the
// best blue card and a lesser black one. Card by card the two best are
// the black and the blue; as a pair they pull the deck two ways, and the
// two black cards are the better pick.
func TestPickTwoChoosesTheBestPair(t *testing.T) {
	fx := loadBotFixture(t)
	picked := fx.cards(t, "p", "Shock", "Lightning Strike", "C2 Fixture Red Cub", "Incinerate", "C2 Fixture Red Scout", "Fire Elemental", "Lightning Bolt")
	pack := fx.cards(t, "c", "Hero's Downfall", "Claustrophobia", "Read the Bones", "Opt", "Naturalize", "Forest", "Island", "Swamp")
	for seed := range uint64(40) {
		// The best card, and then the best card again from what is left,
		// each chosen as a single pick.
		first := botPick(fx, seed, picked, pack, PickOne)
		var rest []Card
		for _, c := range pack {
			if c.ID != first[0] {
				rest = append(rest, c)
			}
		}
		twice := append(names(pack, first), names(rest, botPick(fx, seed, picked, rest, PickOne))...)
		sort.Strings(twice)
		if !reflect.DeepEqual(twice, []string{"Claustrophobia", "Hero's Downfall"}) {
			t.Fatalf("seed %d: the best card twice is %v, want Claustrophobia and Hero's Downfall", seed, twice)
		}
		got := botPick(fx, seed, picked, pack, PickTwo)
		if len(got) != 2 || got[0] == got[1] {
			t.Fatalf("seed %d: a pick-two pick of %v", seed, got)
		}
		pair := names(pack, got)
		sort.Strings(pair)
		if !reflect.DeepEqual(pair, []string{"Hero's Downfall", "Read the Bones"}) {
			t.Fatalf("seed %d: the pair is %v, want the two black cards", seed, pair)
		}
	}

	// With nothing to choose between colours, the pair is the two best
	// cards.
	pack = fx.cards(t, "c", "Serra Angel", "Pacifism", "C2 Fixture White Cub", "Opt", "Forest", "Island")
	pair := names(pack, botPick(fx, 1, nil, pack, PickTwo))
	sort.Strings(pair)
	if !reflect.DeepEqual(pair, []string{"Pacifism", "Serra Angel"}) {
		t.Errorf("the pair from a white pack is %v", pair)
	}
	// A pack with one card left ends on a pick of one.
	last := fx.cards(t, "c", "Opt")
	if got := botPick(fx, 1, picked, last, PickTwo); !reflect.DeepEqual(got, []string{"c1"}) {
		t.Errorf("the pick from a pack of one is %v", got)
	}
}

// A pick is a function of the seed, where the pick falls and the cards:
// asked again it is the same, in any pack order, and a whole seeded
// draft replays card for card.
func TestTheBotIsDeterministicForASeed(t *testing.T) {
	fx := loadBotFixture(t)
	for _, mode := range []Mode{PickOne, PickTwo} {
		a := botDraft(t, fx, mode, 8, 77, BotPicker{Seed: 77, Rater: &fx.rater})
		b := botDraft(t, fx, mode, 8, 77, BotPicker{Seed: 77, Rater: &fx.rater})
		if !reflect.DeepEqual(a, b) {
			t.Fatalf("%s: two drafts with one seed are not the same draft", mode)
		}
		c := botDraft(t, fx, mode, 8, 78, BotPicker{Seed: 78, Rater: &fx.rater})
		if reflect.DeepEqual(a, c) {
			t.Fatalf("%s: two seeds gave the same draft", mode)
		}
	}
	picked := fx.cards(t, "p", "Shock", "Serra Angel", "C2 Fixture Red Cub")
	pack := fx.cards(t, "c", "Lightning Strike", "Incinerate", "Pacifism", "C2 Fixture White Cub", "C2 Fixture Red Scout",
		"Dawnwing Marshal", "Divination", "Murder", "Giant Spider", "Mountain", "Opt")
	for _, mode := range []Mode{PickOne, PickTwo} {
		want := botPick(fx, 9, picked, pack, mode)
		rng := rand.New(rand.NewPCG(1, 2))
		for range 20 {
			shuffled := append([]Card(nil), pack...)
			rng.Shuffle(len(shuffled), func(i, j int) { shuffled[i], shuffled[j] = shuffled[j], shuffled[i] })
			got := botPick(fx, 9, picked, shuffled, mode)
			sort.Strings(got)
			w := append([]string(nil), want...)
			sort.Strings(w)
			if !reflect.DeepEqual(got, w) {
				t.Fatalf("%s: the pick changed with the order of the pack: %v, then %v", mode, want, got)
			}
		}
	}
}

// The seeded noise is small, and enough that eight bots given the same
// choice do not all make it: a pack of near-equal cards is split, a pack
// with one card clearly best is not.
func TestEightBotsDoNotDraftIdentically(t *testing.T) {
	fx := loadBotFixture(t)
	near := fullPack(fx, t, fx.cards(t, "c", "Incinerate", "Lightning Strike", "C2 Fixture White Giant", "Electrolyze", "Juggernaut"))
	clear := fullPack(fx, t, fx.cards(t, "c", "Grave Titan", "Incinerate", "Lightning Strike"))
	took := map[string]bool{}
	for seat := range 8 {
		p := BotPicker{Seed: 31, Rater: &fx.rater}
		req := PickRequest{Pack: near, Mode: PickOne, Take: 1, Seat: seat, Seats: 8, Round: 1, Pick: 1}
		took[p.Pick(req)[0]] = true
		req.Pack = clear
		if got := names(clear, p.Pick(req)); got[0] != "Grave Titan" {
			t.Errorf("seat %d: the noise moved a pick it should not: %v", seat, got)
		}
	}
	if len(took) < 2 {
		t.Errorf("eight bots given five near-equal cards all took the same one")
	}
	for i := range 1000 {
		n := pickNoise(uint64(i), PickRequest{Seat: i % 8, Round: 1 + i%3, Pick: 1 + i%14}, "c"+string(rune('0'+i%10)))
		if n < -1 || n >= 1 {
			t.Fatalf("pickNoise = %v, outside [-1, 1)", n)
		}
	}
	// In a whole draft, no two of the eight end with the same pool.
	pools := botDraft(t, fx, PickOne, 8, 5, BotPicker{Seed: 5, Rater: &fx.rater})
	seen := map[string]bool{}
	for _, pool := range pools {
		var key []string
		for _, c := range pool {
			key = append(key, c.Name)
		}
		sort.Strings(key)
		k := strings.Join(key, "|")
		if seen[k] {
			t.Error("two bots drafted the same pool")
		}
		seen[k] = true
	}
}

// checkingPicker fails the test on a pick the draft would overrule.
type checkingPicker struct {
	t     testing.TB
	inner Picker
	picks *int
}

func (p checkingPicker) Pick(req PickRequest) []string {
	ids := p.inner.Pick(req)
	*p.picks++
	if len(ids) != req.Take {
		p.t.Errorf("round %d pick %d: %d cards named, the pick takes %d", req.Round, req.Pick, len(ids), req.Take)
	}
	in := map[string]bool{}
	for _, c := range req.Pack {
		in[c.ID] = true
	}
	seen := map[string]bool{}
	for _, id := range ids {
		if !in[id] || seen[id] {
			p.t.Errorf("round %d pick %d: %q is not one card of the pack", req.Round, req.Pick, id)
		}
		seen[id] = true
	}
	return ids
}

// Every pick the bot makes through whole drafts is a legal one: exactly
// Take different cards of the pack, so the draft never overrules it.
func TestEveryBotPickIsLegal(t *testing.T) {
	fx := loadBotFixture(t)
	for _, mode := range []Mode{PickOne, PickTwo} {
		for _, seats := range []int{2, 3, 4, 8} {
			n := 0
			botDraft(t, fx, mode, seats, 11, checkingPicker{t, BotPicker{Seed: 11, Rater: &fx.rater}, &n})
			if want := seats * Rounds * (14 / mode.PerPick()); n != want {
				t.Errorf("%s with %d seats: %d picks, want %d", mode, seats, n, want)
			}
		}
	}
	// And nothing is asked of it that it answers badly.
	p := BotPicker{Seed: 1, Rater: &fx.rater}
	if got := p.Pick(PickRequest{Pack: nil, Take: 1}); len(got) != 0 {
		t.Errorf("a pick from an empty pack: %v", got)
	}
	one := fx.cards(t, "c", "Opt")
	if got := p.Pick(PickRequest{Pack: one, Take: 2, Mode: PickTwo}); !reflect.DeepEqual(got, []string{"c1"}) {
		t.Errorf("asked for two of one card: %v", got)
	}
	three := fx.cards(t, "c", "Opt", "Shock", "Murder", "Forest")
	if got := p.Pick(PickRequest{Pack: three, Take: 3, Mode: PickOne, Pick: 1, Round: 1}); len(got) != 3 || got[0] == got[1] || got[1] == got[2] || got[0] == got[2] {
		t.Errorf("a pick of three: %v", got)
	}
}

// A draft made with no Picker is drafted by the bot, and UseCards is
// what gives that bot the server's card index.
func TestADraftsBotsAreTheBotByDefault(t *testing.T) {
	fx := loadBotFixture(t)
	t.Cleanup(func() { UseCards(nil) })
	run := func() [][]Card {
		d, err := New(Config{Name: "d", Set: botFixtureSet, Mode: PickOne, Seats: 8, Seed: 21}, t0)
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := d.picker.(BotPicker); !ok {
			t.Fatalf("a draft's picker is %T by default, want BotPicker", d.picker)
		}
		host, _ := d.AddSeat("host", uuid.Nil)
		if err := d.Start(host.ID, fx.src, t0); err != nil {
			t.Fatal(err)
		}
		for d.State() == StatePicking {
			d.Expire(t0)
		}
		return d.Pools()
	}
	UseCards(fx.idx)
	withIndex := run()
	want := botDraft(t, fx, PickOne, 8, 21, BotPicker{Seed: 21, Rater: &fx.rater})
	if !reflect.DeepEqual(withIndex, want) {
		t.Error("after UseCards the default bot does not draft as a bot with the index's rater")
	}
	// With no index the bot still drafts, from the cards' own fields.
	UseCards(nil)
	without := run()
	if len(without) != 8 || len(without[0]) != 42 {
		t.Fatalf("with no index the draft gave %d pools of %d", len(without), len(without[0]))
	}
	if reflect.DeepEqual(without, withIndex) {
		t.Error("the index made no difference to the draft")
	}
}
