package effects

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// spend_only_spells_test.go — #2556, ADR 0040's 2026-10-07 amendment:
// the cast side of "spend only …". Drain Life, Consume Spirit and Soul
// Burn restrict the mana paid for X by colour; Imperiosaur and Myr
// Superion restrict the mana's SOURCE. The agreement tests at the
// bottom ask one board three questions — the view's legal_actions, the
// bot's enumerator and the real payment — and require one answer.

const (
	drainLifeOracle     = "e75ba79f-4cc2-4ede-8641-559ab94e7e36"
	consumeSpiritOracle = "861fa80d-99e0-4332-a2b8-5aa959fd41a4"
	soulBurnOracle      = "063b0f5d-af27-4681-87b9-b553f6887061"
	imperiosaurOracle   = "e9ced5d8-8337-403f-86a3-bddb9c77d658"
	myrSuperionOracle   = "43dac6fa-4cdf-49ec-9e11-2e5e8f9f928b"
)

// catalogHandSpell puts a catalog spell with its printed mana cost into
// p's hand.
func catalogHandSpell(p *game.Player, name, typeLine, oracle, cost string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: cost,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func atPlayer(p *game.Player) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetPlayer, ID: p.ID}}
}

// castXOffered is every X the enumerator offers for casting `spell`.
func castXOffered(t *testing.T, g *game.Game, seat, spell uuid.UUID) []int {
	t.Helper()
	var xs []int
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != "cast_spell" || m.Source != spell {
			continue
		}
		var p struct {
			XValue int `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("move params: %v", err)
		}
		if !slices.Contains(xs, p.XValue) {
			xs = append(xs, p.XValue)
		}
	}
	slices.Sort(xs)
	return xs
}

func tappedCount(t *testing.T, g *game.Game, ids []uuid.UUID) int {
	t.Helper()
	n := 0
	for _, id := range ids {
		if b31Tapped(t, g, id) {
			n++
		}
	}
	return n
}

// --- Drain Life: "Spend only black mana on X" ------------------------

// Three Swamps and three Mountains make six mana but only three black:
// X plus the {B} needs X+1 of them, so the bot is offered X=2 and no
// more, the engine refuses X=3, and X=2 taps all three Swamps and
// exactly one Mountain (for the {1}).
func TestDrainLifeSpendsOnlyBlackManaOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	swamps := basics(g, me, "Swamp", 3)
	mountains := basics(g, me, "Mountain", 3)
	spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
	at := atPlayer(opp)

	if got := castXOffered(t, g, me.ID, spell); !slices.Equal(got, []int{2}) {
		t.Errorf("enumerator offered X = %v, want [2] — the largest X three black sources pay, not the four any colour would", got)
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 3, Targets: at}),
		"X=3 with three black sources")
	if n := tappedCount(t, g, append(append([]uuid.UUID{}, swamps...), mountains...)); n != 0 {
		t.Fatalf("a refused cast tapped %d lands", n)
	}
	myLife, oppLife := me.Life, opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 2, Targets: at}); err != nil {
		t.Fatalf("X=2 off three Swamps and a Mountain: %v", err)
	}
	if n := tappedCount(t, g, swamps); n != 3 {
		t.Errorf("%d Swamps tapped, want all three — X and {B} are black", n)
	}
	if n := tappedCount(t, g, mountains); n != 1 {
		t.Errorf("%d Mountains tapped, want exactly the one that pays {1}", n)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("life %d / %d, want %d / %d — 2 damage and 2 life", me.Life, opp.Life, myLife+2, oppLife-2)
	}
}

// The pool is held to it too: {B}{B}{B} and a red pay X=1 ({B}{B} for X
// and {B}) with the red on {1}; {R}{R}{B} cannot, whatever the total.
func TestDrainLifeFromThePool(t *testing.T) {
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
	at := atPlayer(opp)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{B}"); err != nil {
		t.Fatal(err)
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 1, Targets: at}), "{R}{R}{B} for X=1")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 1, Targets: at}); err != nil {
		t.Fatalf("{R}{R}{B}{B} for X=1: %v", err)
	}
	if got := poolColors(me); !slices.Equal(got, []string{"R"}) {
		t.Errorf("pool = %v, want one red left", got)
	}
}

// Under Chromatic Orrery any mana pays X: the restriction is about how
// the cost may be paid, which CR 609.4b lets the grant change.
func TestDrainLifeUnderChromaticOrrery(t *testing.T) {
	g, me, opp := spendTable(t)
	pushCatalogPermanent(g, me.ID, "Chromatic Orrery", "Legendary Artifact", chromaticOrreryOracle, false)
	spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}{R}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, Targets: atPlayer(opp)}); err != nil {
		t.Fatalf("four red for X=2 under the Orrery: %v", err)
	}
}

// The price the player sees is the printed one, and the cast pricer
// still reports it: {X}{1}{B}.
func TestDrainLifeKeepsThePrintedPrice(t *testing.T) {
	g, me, _ := spendTable(t)
	spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
	var card game.Card
	for _, c := range me.Hand.Cards {
		if c.InstanceID == spell {
			card = c
		}
	}
	price, err := g.PriceCast(me.ID, card, game.CastSpellParams{XValue: 2})
	if err != nil {
		t.Fatal(err)
	}
	if got := price.Total.String(); got != "{X}{1}{B}" {
		t.Errorf("priced %q, want the printed {X}{1}{B}", got)
	}
	if price.Total.SpendOnly == nil || !price.Total.SpendOnly.XOnly || !slices.Equal(price.Total.SpendOnly.Colors, []string{"B"}) {
		t.Errorf("the price carries %+v, want a black-only X restriction", price.Total.SpendOnly)
	}
}

// The gain is the damage dealt, capped by the recipient as it stood
// before the damage: a 3-toughness creature takes 5 and gives 3, a
// player at 2 life takes 4 and gives 2, a planeswalker at 2 loyalty
// takes 4 and gives 2.
func TestDrainLifeGainsNoMoreThanTheRecipientHad(t *testing.T) {
	cast := func(t *testing.T, g *game.Game, me *game.Player, x int, at []game.TargetRef) {
		t.Helper()
		spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
		if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}{B}{B}{B}"); err != nil {
			t.Fatal(err)
		}
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: x, Targets: at}); err != nil {
			t.Fatalf("cast X=%d: %v", x, err)
		}
		passPriorityAroundTable(t, g)
	}
	t.Run("creature", func(t *testing.T) {
		g, me, opp := spendTable(t)
		bear := b31Push(g, opp.ID, "Grizzly Bears", "Creature — Bear", "", "{1}{G}", 2, 3, "G")
		before := me.Life
		cast(t, g, me, 5, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
		if me.Life != before+3 {
			t.Errorf("life %d → %d, want +3 — the creature's toughness", before, me.Life)
		}
	})
	t.Run("player", func(t *testing.T) {
		g, me, opp := spendTable(t)
		g.WithWriteLock(func() { opp.Life = 2 })
		before := me.Life
		cast(t, g, me, 4, atPlayer(opp))
		if me.Life != before+2 {
			t.Errorf("life %d → %d, want +2 — the player's life total before the damage", before, me.Life)
		}
	})
	t.Run("planeswalker", func(t *testing.T) {
		g, me, opp := spendTable(t)
		walker := pushCatalogPermanent(g, opp.ID, "Test Walker", "Planeswalker — Test", "", false)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == walker {
					g.Battlefield.Cards[i].Counters = map[string]int{game.CounterLoyalty: 2}
				}
			}
		})
		before := me.Life
		cast(t, g, me, 4, []game.TargetRef{{Kind: game.TargetCard, ID: walker}})
		if me.Life != before+2 {
			t.Errorf("life %d → %d, want +2 — the planeswalker's loyalty before the damage", before, me.Life)
		}
	})
}

// K'rrik (ADR 0131) pays the printed {B} with 2 life, but life never
// pays the black mana restricted to X: {R}{R} and 2 life cover the {1}
// and the {B}, and X=1 still wants a black mana.
func TestDrainLifeUnderKrrikPaysOnlyItsPrintedBlackWithLife(t *testing.T) {
	g, me, _, victim := krrikTable(t)
	spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
	at := []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}{R}"); err != nil {
		t.Fatal(err)
	}
	life := me.Life
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 1, PhyrexianLife: 1, Targets: at}),
		"{R}{R} and life for X=1")
	if me.Life != life {
		t.Fatalf("a refused cast cost %d life", life-me.Life)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 1, PhyrexianLife: 1, Targets: at}); err != nil {
		t.Fatalf("{R}{R}{B} and 2 life for X=1: %v", err)
	}
	if me.Life != life-2 {
		t.Errorf("life = %d, want %d — the printed {B} paid with life", me.Life, life-2)
	}
}

// --- Consume Spirit -------------------------------------------------

// Consume Spirit gains X whether or not the damage lands, and spends
// only black on X.
func TestConsumeSpiritSpendsOnlyBlackManaOnXAndGainsX(t *testing.T) {
	g, me, opp := spendTable(t)
	swamps := basics(g, me, "Swamp", 2)
	mountains := basics(g, me, "Mountain", 3)
	spell := catalogHandSpell(me, "Consume Spirit", "Sorcery", consumeSpiritOracle, "{X}{1}{B}")
	at := atPlayer(opp)

	if got := castXOffered(t, g, me.ID, spell); !slices.Equal(got, []int{1}) {
		t.Errorf("enumerator offered X = %v, want [1]", got)
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 2, Targets: at}),
		"X=2 with two black sources")
	myLife, oppLife := me.Life, opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 1, Targets: at}); err != nil {
		t.Fatalf("X=1: %v", err)
	}
	if n := tappedCount(t, g, swamps); n != 2 {
		t.Errorf("%d Swamps tapped, want both", n)
	}
	if n := tappedCount(t, g, mountains); n != 1 {
		t.Errorf("%d Mountains tapped, want one for {1}", n)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife+1 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, myLife+1, oppLife-1)
	}
}

// --- Soul Burn: "black and/or red" ----------------------------------

// Red pays X as well as black, the {B} is still black, and the gain is
// capped by the black mana that can have paid X.
func TestSoulBurnSpendsBlackAndOrRedOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	swamps := basics(g, me, "Swamp", 1)
	mountains := basics(g, me, "Mountain", 2)
	forests := basics(g, me, "Forest", 3)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	at := atPlayer(opp)

	// The Swamp pays {B}, the two Mountains pay X, the Forests pay {2}:
	// six lands could make X=3 if green paid X, and green does not.
	if got := castXOffered(t, g, me.ID, spell); !slices.Equal(got, []int{2}) {
		t.Errorf("enumerator offered X = %v, want [2] — red and black pay X, green does not", got)
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 3, Targets: at}),
		"X=3 with two red sources and a Swamp that pays {B}")
	oppLife, myLife := opp.Life, me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: 2, Targets: at}); err != nil {
		t.Fatalf("X=2: %v", err)
	}
	if n := tappedCount(t, g, swamps) + tappedCount(t, g, mountains); n != 3 {
		t.Errorf("%d black and red lands tapped, want all three", n)
	}
	if n := tappedCount(t, g, forests); n != 2 {
		t.Errorf("%d Forests tapped, want the two that pay {2}", n)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("opponent life %d → %d, want -2", oppLife, opp.Life)
	}
	// One black spent in total and it paid {B}: none was spent on X.
	if me.Life != myLife {
		t.Errorf("life %d → %d, want no gain — no black mana was spent on X", myLife, me.Life)
	}
}

// All black: the life gained is the black mana spent on X, no more.
func TestSoulBurnGainsAsMuchAsTheBlackSpentOnX(t *testing.T) {
	g, me, opp := spendTable(t)
	spell := catalogHandSpell(me, "Soul Burn", "Sorcery", soulBurnOracle, "{X}{2}{B}")
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{B}{B}{B}{B}{R}"); err != nil {
		t.Fatal(err)
	}
	before := me.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 3, Targets: atPlayer(opp)}); err != nil {
		t.Fatalf("X=3 off five black and a red: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Six mana, five of them black, X=3 and three more for {2}{B}: at
	// least two black paid X in the worst case, and the engine counts
	// exactly that (the caveat) — never more than the printed three.
	if gain := me.Life - before; gain != 2 {
		t.Errorf("gained %d, want 2 — the black that must have paid X", gain)
	}
}

// --- Imperiosaur and Myr Superion: "produced by …" --------------------

// Four basic Forests pay for it; a Command Tower-shaped nonbasic, a Sol
// Ring and a Treasure do not, however much mana they make.
func TestImperiosaurSpendsOnlyManaFromBasicLands(t *testing.T) {
	g, me, _ := spendTable(t)
	spell := catalogHandSpell(me, "Imperiosaur", "Creature — Dinosaur", imperiosaurOracle, "{2}{G}{G}")
	forests := basics(g, me, "Forest", 2)
	nonbasic := []uuid.UUID{
		pushCatalogPermanent(g, me.ID, "Overgrown Tomb", "Land — Swamp Forest", "", false),
		pushCatalogPermanent(g, me.ID, "Breeding Pool", "Land — Forest Island", "", false),
	}
	if got := castXOffered(t, g, me.ID, spell); len(got) != 0 {
		t.Fatalf("offered with two basics and two nonbasic lands: %v", got)
	}
	if offered := spellCastOffered(g, me.ID, spell); offered {
		t.Fatal("the enumerator offered Imperiosaur off two basic lands and two nonbasic ones")
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}), "two basic and two nonbasic lands")
	if n := tappedCount(t, g, append(append([]uuid.UUID{}, forests...), nonbasic...)); n != 0 {
		t.Fatalf("a refused cast tapped %d lands", n)
	}
	forests = append(forests, basics(g, me, "Mountain", 2)...)
	if !spellCastOffered(g, me.ID, spell) {
		t.Fatal("the enumerator hid Imperiosaur from four basic lands")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("four basic lands: %v", err)
	}
	if n := tappedCount(t, g, nonbasic); n != 0 {
		t.Errorf("%d nonbasic lands were tapped for a basic-lands-only cost", n)
	}
	if n := tappedCount(t, g, forests); n != 4 {
		t.Errorf("%d basic lands tapped, want 4", n)
	}
}

// Mana already floating from something that is not a basic land cannot
// pay, but mana from a basic land can.
func TestImperiosaurPoolManaMustComeFromBasicLands(t *testing.T) {
	g, me, _ := spendTable(t)
	spell := catalogHandSpell(me, "Imperiosaur", "Creature — Dinosaur", imperiosaurOracle, "{2}{G}{G}")
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
	if err := g.ActivateManaAbility(me.ID, ring, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Sol Ring: %v", err)
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}{G}"); err != nil {
		t.Fatal(err)
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}), "Sol Ring mana and spell mana")
	forest := basics(g, me, "Forest", 1)[0]
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	// {C}{C} from the ring and {G}{G} from a spell are not basic-land
	// mana, so one Forest's {G} is still three short.
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}), "one basic land's mana")
}

// Mana a basic land made itself, floating, pays: the token remembers
// where it came from.
func TestImperiosaurPaysFromFloatingBasicLandMana(t *testing.T) {
	g, me, _ := spendTable(t)
	spell := catalogHandSpell(me, "Imperiosaur", "Creature — Dinosaur", imperiosaurOracle, "{2}{G}{G}")
	for _, id := range basics(g, me, "Forest", 4) {
		if err := g.ActivateManaAbility(me.ID, id, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap a Forest: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("four floating basic-land mana: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool = %v, want it spent", poolColors(me))
	}
}

func TestMyrSuperionSpendsOnlyManaFromCreatures(t *testing.T) {
	g, me, _ := spendTable(t)
	spell := catalogHandSpell(me, "Myr Superion", "Artifact Creature — Myr", myrSuperionOracle, "{2}")
	lands := basics(g, me, "Forest", 3)
	ring := pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
	if spellCastOffered(g, me.ID, spell) {
		t.Fatal("the enumerator offered Myr Superion off three lands and a Sol Ring")
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}), "lands and a Sol Ring")
	elves := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	elves2 := pushCatalogPermanent(g, me.ID, "Llanowar Elves", "Creature — Elf Druid", llanowarElvesOracle, false)
	if !spellCastOffered(g, me.ID, spell) {
		t.Fatal("the enumerator hid Myr Superion from two mana creatures")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("two Llanowar Elves: %v", err)
	}
	if !b31Tapped(t, g, elves) || !b31Tapped(t, g, elves2) {
		t.Error("the Elves were not tapped")
	}
	if n := tappedCount(t, g, append(lands, ring)); n != 0 {
		t.Errorf("%d non-creature sources were tapped", n)
	}
}

func spellCastOffered(g *game.Game, seat, spell uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == "cast_spell" && m.Source == spell {
			return true
		}
	}
	return false
}

// --- agreement: the view, the bot and the payment ----------------------

// castAgreement asks one board three questions about casting `spell`
// and fails unless they give one answer: the view's legal_actions
// digest, the bot's enumerator, and the payment (strict, auto-tapped).
func castAgreement(t *testing.T, name string, g *game.Game, me, opp *game.Player, spell uuid.UUID, want bool) {
	t.Helper()
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			if c.KnownBy == nil {
				c.KnownBy = map[uuid.UUID]bool{}
			}
			for _, p := range g.Seats {
				c.KnownBy[p.ID] = true
			}
		}
	})
	view := false
	if v := protocol.ViewOfGameFor(g, me.ID.String()); v.LegalActions != nil {
		if src := v.LegalActions.Sources[spell.String()]; src != nil {
			view = slices.Contains(src.Kinds, legal.KindCast)
		}
	}
	bot := spellCastOffered(g, me.ID, spell)
	// The X the payment tries: the largest the enumerator offers, or 1
	// for a spell it hides — Drain Life is never offered at X=0.
	x := 1
	if xs := castXOffered(t, g, me.ID, spell); len(xs) > 0 {
		x = xs[len(xs)-1]
	}
	err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true, XValue: x, Targets: atPlayer(opp)})
	if view != want || bot != want || (err == nil) != want {
		t.Errorf("%s: view %v, bot %v, payment %v — want all %v", name, view, bot, err, want)
	}
}

func TestDrainLifeViewBotAndPaymentAgree(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lands map[string]int
		want  bool
	}{
		{"two Swamps and a Mountain", map[string]int{"Swamp": 2, "Mountain": 1}, true},
		{"one Swamp and five Mountains", map[string]int{"Swamp": 1, "Mountain": 5}, false},
		{"three Islands and a Swamp", map[string]int{"Island": 3, "Swamp": 1}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp := spendTable(t)
			for land, n := range tc.lands {
				basics(g, me, land, n)
			}
			spell := catalogHandSpell(me, "Drain Life", "Sorcery", drainLifeOracle, "{X}{1}{B}")
			castAgreement(t, tc.name, g, me, opp, spell, tc.want)
		})
	}
}

func TestImperiosaurViewBotAndPaymentAgree(t *testing.T) {
	t.Run("four basic lands", func(t *testing.T) {
		g, me, opp := spendTable(t)
		basics(g, me, "Forest", 4)
		spell := catalogHandSpell(me, "Imperiosaur", "Creature — Dinosaur", imperiosaurOracle, "{2}{G}{G}")
		castAgreement(t, "four basic lands", g, me, opp, spell, true)
	})
	t.Run("two basic lands and a Sol Ring", func(t *testing.T) {
		g, me, opp := spendTable(t)
		basics(g, me, "Forest", 2)
		pushCatalogPermanent(g, me.ID, "Sol Ring", "Artifact", solRingOracle, false)
		spell := catalogHandSpell(me, "Imperiosaur", "Creature — Dinosaur", imperiosaurOracle, "{2}{G}{G}")
		castAgreement(t, "two basic lands and a Sol Ring", g, me, opp, spell, false)
	})
}

// --- registration --------------------------------------------------

// Every shape the cast pricer cannot pay is refused at boot.
func TestRegisterRefusesASpellSpendOnlyClauseItCannotPay(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		spec       Spec
	}{
		{"no colour", "names no colour", Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyOnX()}},
		{"colorless", "not one of W U B R G", Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyOnX("C")}},
		{"whole cost", "not the \"on X\" form", Spec{Name: "Test Card", SpendOnly: &game.ManaSpendOnly{Colors: []string{"B"}}}},
		{"chosen colour", "chosen color", Spec{Name: "Test Card", SpendOnly: &game.ManaSpendOnly{ChosenColor: true, XOnly: true}}},
		{"delve", "delve or a tap cost", Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyOnX("B"), Delve: true}},
		{"convoke", "delve or a tap cost", Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyOnX("B"), TapCost: Convoke()}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mustPanic(t, tc.want, func() { checkSpellSpendOnly(tc.spec) })
		})
	}
	checkSpellSpendOnly(Spec{Name: "Test Card", SpendOnly: SpellSpendOnlyOnX("B", "R")})
	checkSpellSpendOnly(Spec{Name: "Test Card", SpendOnlySources: game.ManaSourceBasicLand})
}
