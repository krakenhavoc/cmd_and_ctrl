package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// nghathrod_a_cards_test.go — #4092 slice a, the Captain N'ghathrod deck:
// the commander, Zellix, Fractured Sanity, Memory Plunder and Psionic
// Ritual, with Psionic Ritual's replicate-by-tapping cost (ReplicateTapping).
// Every helper here is prefixed ngha.

const (
	nghaCaptainOracle        = "cfb23c6b-6e4a-4fc9-b3bb-a3ddc2ba06b8"
	nghaZellixOracle         = "4dde484a-8d5b-4eb6-ab4a-e30cc370431c"
	nghaFracturedSanityOracl = "cf32524e-0df9-4af8-9371-84feeb0a86c6"
	nghaMemoryPlunderOracle  = "36b67014-0507-4976-ad52-a2a36676197d"
	nghaPsionicRitualOracle  = "1c3e6980-97fe-4e15-93a0-368280da8608"
)

func TestNghaCardsAreFull(t *testing.T) {
	for _, oracle := range []string{nghaCaptainOracle, nghaZellixOracle, nghaFracturedSanityOracl, nghaMemoryPlunderOracle, nghaPsionicRitualOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// nghaCreature puts an untapped, non-sick creature onto `controller`'s
// battlefield.
func nghaCreature(g *game.Game, controller uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return apaPush(g, controller, controller, game.Card{Name: name, TypeLine: typeLine, Power: power, Toughness: toughness})
}

// nghaLibraryTop puts a card on top of p's library.
func nghaLibraryTop(g *game.Game, p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Library.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Power: 3, Toughness: 3, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// nghaPassUntil passes priority until done reports true.
func nghaPassUntil(t *testing.T, g *game.Game, what string, done func() bool) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if done() {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority waiting for %s: %v", what, err)
		}
	}
	if !done() {
		t.Fatalf("never reached: %s", what)
	}
}

// nghaHorrorTokens counts `controller`'s Horror tokens.
func nghaHorrorTokens(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == "Horror" {
			n++
		}
	}
	return n
}

// Captain N'ghathrod: Horrors you control have menace — the Captain
// itself and another Horror — while a non-Horror and an opponent's Horror
// do not.
func TestNghaCaptainGivesHorrorsYouControlMenace(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	captain := b43Catalog(g, me.ID, "Captain N'ghathrod", "Legendary Creature — Horror Pirate", nghaCaptainOracle, 3, 6)
	horror := nghaCreature(g, me.ID, "Horror", "Creature — Horror", 2, 2)
	bear := nghaCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := nghaCreature(g, opp.ID, "Their Horror", "Creature — Horror", 2, 2)
	for id, want := range map[uuid.UUID]bool{captain: true, horror: true, bear: false, theirs: false} {
		if got := hasEffectiveKeyword(t, g, id, "menace"); got != want {
			t.Errorf("%s: menace %v, want %v", id, got, want)
		}
	}
}

// Captain N'ghathrod: each Horror that connects makes the damaged player
// mill the damage it dealt (the Captain 3, a 2/2 Horror 2); a non-Horror
// mills nothing. At the end step the trigger takes a creature card milled
// this turn onto the battlefield under the Captain's controller; a card
// that was already in that graveyard is not a legal target.
func TestNghaCaptainHorrorsMillAndTheEndStepTakesAMilledCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	captain := b43Catalog(g, me.ID, "Captain N'ghathrod", "Legendary Creature — Horror Pirate", nghaCaptainOracle, 3, 6)
	horror := nghaCreature(g, me.ID, "Horror", "Creature — Horror", 2, 2)
	bear := nghaCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	old := gyCard(opp, "Old Beast", "Creature — Beast", "{3}", "")
	beast := nghaLibraryTop(g, opp, "Milled Beast", "Creature — Beast")
	before := opp.Library.Size()

	attackWith(t, g, opp.ID, captain, horror, bear)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if got := before - opp.Library.Size(); got != 5 {
		t.Fatalf("the defender milled %d, want 5 (3 for the Captain, 2 for the Horror, none for the Bear)", got)
	}
	if !opp.Graveyard.Contains(beast) {
		t.Fatal("the top card was not milled")
	}

	advanceTo(t, g, game.StepEnd)
	p := pickTargetPrompt(t, g)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: old}); err == nil {
		t.Fatal("a creature card already in the graveyard is a legal target")
	}
	answerPickTarget(t, g, beast)
	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, beast); got != me.ID {
		t.Fatalf("the milled Beast is controlled by %s, want the Captain's controller", got)
	}
	if opp.Graveyard.Contains(beast) {
		t.Error("the Beast is still in the graveyard")
	}
}

// With nothing milled this turn the end-step trigger has no legal target
// and is removed (CR 603.3d): no prompt, nothing on the stack.
func TestNghaCaptainEndStepWithNothingMilledAsksNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	b43Catalog(g, me.ID, "Captain N'ghathrod", "Legendary Creature — Horror Pirate", nghaCaptainOracle, 3, 6)
	gyCard(opp, "Old Beast", "Creature — Beast", "{3}", "")
	advanceTo(t, g, game.StepEnd)
	if latestPickTarget(g, me.ID) != nil || !stackFullyEmpty(g) {
		t.Fatal("the end-step trigger asked or went on the stack with no legal target")
	}
}

// Zellix: Hive Mind fires once per player who milled a creature card —
// Fractured Sanity milling two opponents' creature cards is two Horrors,
// the third opponent's creatureless mill none — and {1}, {T} makes a
// target player mill three; two creature cards in that mill are still
// one Horror.
func TestNghaZellixHiveMindAndActivation(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	a, b, c := g.Seats[(g.Turn.ActiveSeat+1)%4], g.Seats[(g.Turn.ActiveSeat+2)%4], g.Seats[(g.Turn.ActiveSeat+3)%4]
	zellix := nghaCreature(g, me.ID, "Zellix, Sanity Flayer", "Legendary Creature — Horror", 2, 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, zellix).OracleID = nghaZellixOracle })
	nghaLibraryTop(g, a, "Beast A", "Creature — Beast")
	nghaLibraryTop(g, a, "Beast A2", "Creature — Beast")
	nghaLibraryTop(g, b, "Beast B", "Creature — Beast")
	before := [3]int{a.Library.Size(), b.Library.Size(), c.Library.Size()}

	castCatalogSpell(t, g, "Fractured Sanity", "Sorcery", nghaFracturedSanityOracl, nil)
	passPriorityAroundTable(t, g)
	for i, p := range []*game.Player{a, b, c} {
		if got := before[i] - p.Library.Size(); got != min(14, before[i]) {
			t.Errorf("opponent %d milled %d, want %d", i, got, min(14, before[i]))
		}
	}
	if n := nghaHorrorTokens(g, me.ID); n != 2 {
		t.Fatalf("%d Horror tokens, want 2 (one per opponent who milled a creature card)", n)
	}

	nghaLibraryTop(g, me, "My Beast", "Creature — Beast")
	nghaLibraryTop(g, me, "My Beast 2", "Creature — Beast")
	g.WithWriteLock(func() { fillPool(me, 1) })
	mine := me.Library.Size()
	if err := g.ActivateCatalogAbility(me.ID, zellix, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err != nil {
		t.Fatalf("activate Zellix: %v", err)
	}
	if !tappedForTest(t, g, zellix) {
		t.Error("the activation did not tap Zellix")
	}
	passPriorityAroundTable(t, g)
	if got := mine - me.Library.Size(); got != 3 {
		t.Errorf("the target player milled %d, want 3", got)
	}
	if n := nghaHorrorTokens(g, me.ID); n != 3 {
		t.Errorf("%d Horror tokens, want 3: two creature cards in one mill is one more", n)
	}
	if tok := nghaFirstHorrorToken(g, me.ID); tok == nil || tok.Power != 1 || tok.Toughness != 1 || len(tok.Colors) != 1 || tok.Colors[0] != "B" {
		t.Errorf("token %+v, want a 1/1 black Horror", tok)
	}
}

func nghaFirstHorrorToken(g *game.Game, controller uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Controller == controller && c.IsToken() && c.Name == "Horror" {
			return c
		}
	}
	return nil
}

// Fractured Sanity's cycling draws a card, and the cycle trigger mills
// each opponent four from the graveyard.
func TestNghaFracturedSanityCyclesAndMillsEachOpponentFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var before [4]int
	for i, p := range g.Seats {
		before[i] = p.Library.Size()
	}
	hand := me.Hand.Size()
	id, _ := cycleFromHand(t, g, "Fractured Sanity", "Sorcery", nghaFracturedSanityOracl, "{C}{U}")
	if !me.Graveyard.Contains(id) {
		t.Fatal("the cycled card is not in the graveyard")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1-1+1 {
		t.Errorf("hand %d → %d: cycling discards this card and draws one", hand, me.Hand.Size())
	}
	for i, p := range g.Seats {
		want := 4
		if p.ID == me.ID {
			want = 1 // the cycling draw
		}
		if got := before[i] - p.Library.Size(); got != want {
			t.Errorf("seat %d library fell by %d, want %d", i, got, want)
		}
	}
}

// Memory Plunder: a card in your own graveyard is not a legal target; an
// opponent's instant is cast for free from their graveyard and, being the
// real card, goes back to its owner's graveyard.
func TestNghaMemoryPlunderCastsAnOpponentsInstantForFree(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	bolt := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	own := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	if err := castCatalogSpellErr(t, g, "Memory Plunder", "Instant", nghaMemoryPlunderOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: own}}); err == nil {
		t.Fatal("a card in your own graveyard is not in an opponent's")
	}
	g.WithWriteLock(func() { me.Hand.Cards = nil })

	castCatalogSpell(t, g, "Memory Plunder", "Instant", nghaMemoryPlunderOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	life := third.Life
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyGrantedCost, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: third.ID}},
	}); err != nil {
		t.Fatalf("casting the opponent's instant for free: %v", err)
	}
	passPriorityAroundTable(t, g)
	if third.Life != life-3 {
		t.Errorf("life %d → %d: the plundered Bolt deals 3", life, third.Life)
	}
	if !opp.Graveyard.Contains(bolt) || inExile(g, bolt) {
		t.Error("the spell goes back to its owner's graveyard")
	}
}

// nghaRitualTable is Psionic Ritual in the active seat's hand with
// {4}{U}{U} floating, two untapped Horrors (one that arrived this turn),
// a tapped Horror, a Bear and an opponent's Horror.
type nghaRitualTable struct {
	g                                    *game.Game
	me, opp, third                       *game.Player
	ritual                               uuid.UUID
	fresh, other, tapped, bear, theirOne uuid.UUID
}

func nghaNewRitualTable(t *testing.T) nghaRitualTable {
	t.Helper()
	g, me, opp := tapTable(t)
	r := nghaRitualTable{g: g, me: me, opp: opp, third: g.Seats[2]}
	r.ritual = handCardOf(g, me, "Psionic Ritual", "Sorcery", "{4}{U}{U}", nghaPsionicRitualOracle)
	r.fresh = nghaCreature(g, me.ID, "Fresh Horror", "Creature — Horror", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, r.fresh).SummonedThisTurn = true })
	r.other = nghaCreature(g, me.ID, "Other Horror", "Creature — Horror", 1, 1)
	r.tapped = nghaCreature(g, me.ID, "Tapped Horror", "Creature — Horror", 1, 1)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, r.tapped).Tapped = true })
	r.bear = nghaCreature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	r.theirOne = nghaCreature(g, opp.ID, "Their Horror", "Creature — Horror", 1, 1)
	g.WithWriteLock(func() {
		fillPoolColored(me, "U", 2)
		fillPool(me, 4)
	})
	return r
}

// Psionic Ritual's replicate cost: one untapped Horror you control per
// payment, on teamwork_ids. Refused: a payment with nothing named, a
// non-Horror, a tapped Horror, an opponent's Horror, too few Horrors for
// the payments, and Horrors named with no payment — and a refused cast
// pays nothing. Accepted for two payments naming two Horrors, one of them
// arrived this turn: both tapped with the spell on the stack, recorded on
// the payment, and the replicate trigger counts two.
func TestNghaPsionicRitualReplicateTapsOneHorrorPerPayment(t *testing.T) {
	r := nghaNewRitualTable(t)
	g, me := r.g, r.me
	bolt := gyCard(r.opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}
	cast := func(times int, taps ...uuid.UUID) error {
		var opt []int
		for range times {
			opt = append(opt, 0)
		}
		return g.CastSpell(me.ID, r.ritual, game.CastSpellParams{Targets: target, OptionalCosts: opt, TeamworkIDs: taps, Strict: true})
	}
	for name, c := range map[string]struct {
		times int
		taps  []uuid.UUID
	}{
		"nothing named":         {1, nil},
		"a non-Horror":          {1, []uuid.UUID{r.bear}},
		"a tapped Horror":       {1, []uuid.UUID{r.tapped}},
		"an opponent's Horror":  {1, []uuid.UUID{r.theirOne}},
		"one Horror for two":    {2, []uuid.UUID{r.fresh}},
		"a Horror with no paid": {0, []uuid.UUID{r.fresh}},
		"the same Horror twice": {2, []uuid.UUID{r.fresh, r.fresh}},
	} {
		if err := cast(c.times, c.taps...); err == nil {
			t.Fatalf("%s paid Psionic Ritual's replicate", name)
		}
	}
	if tappedForTest(t, g, r.fresh) || tappedForTest(t, g, r.other) || !me.Hand.Contains(r.ritual) {
		t.Fatal("a refused cast paid something")
	}
	if err := cast(2, r.fresh, r.other); err != nil {
		t.Fatalf("replicate twice for two Horrors: %v", err)
	}
	if !tappedForTest(t, g, r.fresh) || !tappedForTest(t, g, r.other) {
		t.Fatal("the Horrors were not tapped with the spell on the stack")
	}
	if item := g.StackMeta[r.ritual]; item == nil || len(item.Paid.TappedOthers) != 2 {
		t.Errorf("the payment does not record two tapped Horrors: %+v", g.StackMeta[r.ritual])
	}
	if it := triggerOnStack(g, r.ritual); it == nil || it.XValue != 2 {
		t.Fatalf("replicate trigger = %+v, want one counting two payments", it)
	}
}

// Psionic Ritual, replicated once: the copy is retargeted to a card in
// another player's graveyard and resolves first, exiling that card and
// offering a free copy of it — the original Ritual still on the stack.
// Then the original exiles its own target, offers that copy, and exiles
// itself. Each offered copy casts for free and resolves as the card.
func TestNghaPsionicRitualCopiesFromAnyGraveyardAndExilesItself(t *testing.T) {
	r := nghaNewRitualTable(t)
	g, me, opp := r.g, r.me, r.opp
	bolt1 := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	bolt2 := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	if err := g.CastSpell(me.ID, r.ritual, game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt1}}, OptionalCosts: []int{0}, TeamworkIDs: []uuid.UUID{r.other}, Strict: true,
	}); err != nil {
		t.Fatalf("cast Psionic Ritual: %v", err)
	}
	nghaPassUntil(t, g, "the copy's new-target prompt", func() bool { return latestPickTarget(g, me.ID) != nil })
	p := latestPickTarget(g, me.ID)
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bolt2}); err != nil {
		t.Fatalf("retarget the copy: %v", err)
	}
	nghaPassUntil(t, g, "the copy to resolve", func() bool { return inExile(g, bolt2) })
	if !g.Stack.Contains(r.ritual) || inExile(g, r.ritual) {
		t.Fatal("the original Psionic Ritual left the stack when its copy resolved")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	life := opp.Life
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the first copy for free: %v", err)
	}
	nghaPassUntil(t, g, "the first Bolt copy", func() bool { return opp.Life == life-3 })

	nghaPassUntil(t, g, "the original to resolve", func() bool { return inExile(g, bolt1) })
	if !inExile(g, r.ritual) || me.Graveyard.Contains(r.ritual) {
		t.Error("Exile Psionic Ritual: the card is exiled, not put into the graveyard")
	}
	cp = copyOf(t, g, "Lightning Bolt")
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the second copy for free: %v", err)
	}
	nghaPassUntil(t, g, "the second Bolt copy", func() bool { return opp.Life == life-6 })
}

// The wire and the bot: the replicate offer ships tap_others_options (the
// caster's untapped Horrors, one per payment) with max_times capped at
// what they cover; the enumerator offers one and two payments naming that
// many Horrors, never three, and the engine accepts the two-payment move.
// An opponent's view of the hand card carries none of it.
func TestNghaPsionicRitualWireAndEnumerator(t *testing.T) {
	r := nghaNewRitualTable(t)
	g, me := r.g, r.me
	bolt := gyCard(r.opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	_ = bolt

	offer := func(viewer uuid.UUID) *protocol.OptionalCostView {
		v := protocol.ViewOfGameFor(g, viewer.String())
		for _, s := range v.Seats {
			for i := range s.Hand.Cards {
				if c := &s.Hand.Cards[i]; c.InstanceID == r.ritual.String() && len(c.OptionalCosts) > 0 {
					return &c.OptionalCosts[0]
				}
			}
		}
		return nil
	}
	o := offer(me.ID)
	if o == nil || o.Key != game.ReplicateKey || o.TapOthersLabel != "an untapped Horror you control" {
		t.Fatalf("replicate offer %+v", o)
	}
	if to := o.TapOthersOptions; to == nil || to.Min != 1 || to.Max != 1 || len(to.Cards) != 2 {
		t.Fatalf("tap_others_options = %+v, want 1/1 over the two untapped Horrors", o.TapOthersOptions)
	}
	if o.MaxTimes != 2 {
		t.Errorf("max_times = %d, want 2: two Horrors pay two replicates", o.MaxTimes)
	}
	for _, other := range []*game.Player{r.opp, r.third} {
		if leaked := offer(other.ID); leaked != nil {
			t.Errorf("seat %s sees the hidden card's replicate offer: %+v", other.ID, leaked)
		}
	}

	type castMove struct {
		OptionalCosts []int    `json:"optional_costs"`
		TeamworkIDs   []string `json:"teamwork_ids"`
	}
	var two *legal.Move
	seen := map[int]bool{}
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != r.ritual {
			continue
		}
		var p castMove
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		n := len(p.OptionalCosts)
		seen[n] = true
		if len(p.TeamworkIDs) != n {
			t.Errorf("move %q pays %d replicates naming %d Horrors", m.Label, n, len(p.TeamworkIDs))
		}
		if n == 2 && two == nil {
			mv := m
			two = &mv
		}
	}
	if !seen[0] || !seen[1] || !seen[2] || seen[3] {
		t.Fatalf("enumerated replicate counts %v, want 0, 1 and 2 only", seen)
	}
	if !strings.Contains(two.Label, "tapping") || !strings.Contains(two.Label, "Fresh Horror") || !strings.Contains(two.Label, "Other Horror") {
		t.Errorf("move label %q does not name the tapped Horrors", two.Label)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(two.Type), Player: me.ID, Caller: me.ID, Params: two.Params}); err != nil {
		t.Fatalf("the enumerated two-replicate cast is refused: %v", err)
	}
	if !tappedForTest(t, g, r.fresh) || !tappedForTest(t, g, r.other) {
		t.Error("the enumerated move did not tap both Horrors")
	}
}

// Register refuses a tapping optional cost that is not replicate, a
// replicate tap with no fixed count, and one beside a mandatory tap cost.
func TestNghaRegisterRefusesMalformedReplicateTaps(t *testing.T) {
	kicker := ReplicateTapping("an untapped Horror you control", 3, OfCreatureType("Horror"))
	kicker.Key = "kicker"
	open := ReplicateTapping("an untapped Horror you control", 3, OfCreatureType("Horror"))
	open.TapOthers.AnyNumber, open.TapOthers.Count = true, 0
	ok := ReplicateTapping("an untapped Horror you control", 3, OfCreatureType("Horror"))
	for _, c := range []struct {
		name string
		spec Spec
		want string
	}{
		{"kicker", Spec{OptionalCosts: []game.AdditionalCost{kicker}}, "only the mandatory slot"},
		{"open", Spec{OptionalCosts: []game.AdditionalCost{open}, Triggered: []game.TriggeredAbility{Replicate()}}, "no fixed number"},
		{"beside mandatory", Spec{
			AdditionalCost: TapUntappedCost(1, "an untapped creature you control", Creature()),
			OptionalCosts:  []game.AdditionalCost{ok}, Triggered: []game.TriggeredAbility{Replicate()},
		}, "mandatory tap cost with a replicate"},
	} {
		c.spec.OracleID, c.spec.Name = "test-4092-replicate-"+c.name, "Test "+c.name
		if msg := registerPanics(c.spec); !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}
