package effects

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// power_up_additions_test.go — #4571, the three power-up additions
// (CR 702.193, CR 106.6): Wonder Man, Hollywood Hero's "can be activated
// an additional time", Kang the Conqueror's "During that turn, power-up
// abilities can't be activated" and Quinjet Technician's "Spend this
// mana only to activate power-up abilities". The engine halves are
// game/power_up_additions_test.go; the cards' base cases are in
// power_up_cards_test.go.

const (
	wonderManOracle = "e2c49754-711e-43b3-82f1-5376a4d99701"
	kangOracle      = "da6e24a4-8adc-41a1-b5eb-3e31dc1250c9"
	quinjetOracle   = "e8025c63-80b4-4cce-bf2d-df294450070d"
)

const (
	brawlerType = "Creature — Human Warrior Hero"
	// brawlerPay is Brave Brawler's printed {4}{W}.
	brawlerPay = "{C}{C}{C}{C}{W}"
)

func TestPowerUpAdditionCardsAreFull(t *testing.T) {
	for _, oracle := range []string{wonderManOracle, kangOracle, quinjetOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: Completeness = %v, Caveats = %v, want Full and none", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// powerUpLimit is game.Game.PowerUpLimitFor for a battlefield card.
func powerUpLimit(g *game.Game, id uuid.UUID) int {
	limit := 0
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				limit = g.PowerUpLimitFor(c)
			}
		}
	})
	return limit
}

// activatePowerUp announces row 0 of `id` with `pay` floated and resolves
// it, failing the test if the engine refuses.
func activatePowerUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, pay string) {
	t.Helper()
	floatMana(t, g, p, pay)
	if err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("power-up: %v", err)
	}
	resolveTopOfStack(t, g)
}

// refusedPowerUp announces row 0 of `id` with `pay` floated and returns
// the engine's refusal, emptying the pool after a refusal.
func refusedPowerUp(t *testing.T, g *game.Game, p *game.Player, id uuid.UUID, pay string) error {
	t.Helper()
	floatMana(t, g, p, pay)
	err := g.ActivateCatalogAbility(p.ID, id, 0, game.ActivateAbilityParams{})
	if err == nil {
		t.Fatalf("power-up was accepted, want a refusal")
	}
	g.WithWriteLock(func() { p.ManaPool = nil })
	return err
}

// --- Wonder Man, Hollywood Hero ------------------------------------------

// "Each power-up ability of permanents you control can be activated an
// additional time": twice, not three times, for a permanent I control;
// still once for an opponent's.
func TestWonderManLetsMyPowerUpAbilitiesBeActivatedTwice(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	theirs := pushPowerUpPermanent(g, opp.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	if got := powerUpLimit(g, mine); got != 1 {
		t.Fatalf("limit before Wonder Man = %d, want 1", got)
	}
	pushPowerUpPermanent(g, me.ID, "Wonder Man, Hollywood Hero", "Legendary Creature — Human Performer Hero", "{3}{R}{R}", wonderManOracle)

	if got := powerUpLimit(g, mine); got != 2 {
		t.Errorf("my Brawler's limit under Wonder Man = %d, want 2", got)
	}
	if got := powerUpLimit(g, theirs); got != 1 {
		t.Errorf("an opponent's Brawler's limit = %d, want 1 — the static says permanents YOU control", got)
	}
	activatePowerUp(t, g, me, mine, brawlerPay)
	activatePowerUp(t, g, me, mine, brawlerPay)
	if got := plusOnes(g, mine); got != 4 {
		t.Errorf("+1/+1 counters after two power-ups = %d, want 4", got)
	}
	if err := refusedPowerUp(t, g, me, mine, brawlerPay); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("third activation: %v, want ErrPowerUpUsed", err)
	}
}

// Wonder Man's own power-up is one of "each power-up ability of
// permanents you control": the turn he entered it costs {5}{R}{R} less
// {3}{R}{R} each time, which is {2}, and it can be activated twice.
func TestWonderManMayPowerUpTwiceHimself(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	wm := castPowerUpCreature(t, g, me, "Wonder Man, Hollywood Hero", "Legendary Creature — Human Performer Hero",
		"{3}{R}{R}", "{C}{C}{C}{R}{R}", wonderManOracle, 4)

	if got := chargedPowerUp(t, g, me.ID, wm); got != "{2}" {
		t.Fatalf("his power-up the turn he entered: %q, want {2}", got)
	}
	activatePowerUp(t, g, me, wm, "{C}{C}")
	activatePowerUp(t, g, me, wm, "{C}{C}")
	if got := plusOnes(g, wm); got != 4 {
		t.Errorf("+1/+1 counters = %d, want 4 (two activations of two)", got)
	}
	if err := refusedPowerUp(t, g, me, wm, "{C}{C}"); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("third activation: %v, want ErrPowerUpUsed", err)
	}
}

// "An additional time" adds: two Wonder Men make it three, and one
// leaving takes one back.
func TestTwoWonderMenMakeItThreeAndEachLeavingTakesOneBack(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mine := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	// Not legendary, so the legend rule leaves both: a real second copy
	// would be a clone, which the engine treats the same way.
	first := pushPowerUpPermanent(g, me.ID, "Wonder Man, Hollywood Hero", "Creature — Human Performer Hero", "{3}{R}{R}", wonderManOracle)
	second := pushPowerUpPermanent(g, me.ID, "Wonder Man, Hollywood Hero", "Creature — Human Performer Hero", "{3}{R}{R}", wonderManOracle)
	if got := powerUpLimit(g, mine); got != 3 {
		t.Fatalf("limit under two Wonder Men = %d, want 3", got)
	}
	for i := 0; i < 3; i++ {
		activatePowerUp(t, g, me, mine, brawlerPay)
	}
	if err := refusedPowerUp(t, g, me, mine, brawlerPay); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("fourth activation: %v, want ErrPowerUpUsed", err)
	}
	leaveBattlefield(t, g, me.ID, first)
	leaveBattlefield(t, g, me.ID, second)
	if got := powerUpLimit(g, mine); got != 1 {
		t.Errorf("limit with no Wonder Man = %d, want 1", got)
	}
}

// The static is the duration: once Wonder Man leaves, an ability that has
// been activated once is spent again.
func TestAPowerUpActivatedUnderWonderManIsSpentWhenHeLeaves(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mine := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	wm := pushPowerUpPermanent(g, me.ID, "Wonder Man, Hollywood Hero", "Legendary Creature — Human Performer Hero", "{3}{R}{R}", wonderManOracle)
	activatePowerUp(t, g, me, mine, brawlerPay)

	leaveBattlefield(t, g, me.ID, wm)
	if err := refusedPowerUp(t, g, me, mine, brawlerPay); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("after Wonder Man left: %v, want ErrPowerUpUsed", err)
	}
}

// A restore with one activation recorded keeps the record and the
// static: the second still works, the third is refused.
func TestWonderManSurvivesARestoreBetweenActivations(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	mine := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	pushPowerUpPermanent(g, me.ID, "Wonder Man, Hollywood Hero", "Legendary Creature — Human Performer Hero", "{3}{R}{R}", wonderManOracle)
	activatePowerUp(t, g, me, mine, brawlerPay)

	restored := restoreThroughJSON(t, g)
	rme := restored.PlayerByIDForEffect(me.ID)
	activatePowerUp(t, restored, rme, mine, brawlerPay)
	if err := refusedPowerUp(t, restored, rme, mine, brawlerPay); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("third activation after the restore: %v, want ErrPowerUpUsed", err)
	}
}

// --- Kang the Conqueror --------------------------------------------------

const kangType = "Legendary Creature — Human Villain"

// Kang's power-up: the price the turn he entered, the counter, and the
// extra turn it queues.
func TestKangPowerUpPutsACounterAndQueuesAnExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	kang := castPowerUpCreature(t, g, me, "Kang the Conqueror", kangType, "{2}{U}{U}", "{C}{C}{U}{U}", kangOracle, 4)
	if got := chargedPowerUp(t, g, me.ID, kang); got != "{3}{U}" {
		t.Fatalf("his power-up the turn he entered: %q, want {3}{U} ({5}{U}{U}{U} less {2}{U}{U})", got)
	}
	activatePowerUp(t, g, me, kang, "{C}{C}{C}{U}")
	if got := plusOnes(g, kang); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
	queued := g.ExtraTurnsQueued()
	if len(queued) != 1 || queued[0].Seat != 0 {
		t.Fatalf("extra turns queued = %+v, want one for seat 0", queued)
	}
	if err := refusedPowerUp(t, g, me, kang, "{C}{C}{C}{U}"); !errors.Is(err, game.ErrPowerUpUsed) {
		t.Errorf("second activation: %v, want ErrPowerUpUsed", err)
	}
}

// "During that turn, power-up abilities can't be activated": not in the
// turn the ability resolved in, not in the turn after the extra one, and
// for nobody during it.
func TestKangBansEveryPowerUpAbilityDuringTheExtraTurnOnly(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	kang := pushPowerUpPermanent(g, me.ID, "Kang the Conqueror", kangType, "{2}{U}{U}", kangOracle)
	early := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	banned := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	theirs := pushPowerUpPermanent(g, opp.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	activatePowerUp(t, g, me, kang, "{C}{C}{C}{C}{C}{U}{U}{U}")

	// The turn Kang's ability resolved in is untouched.
	activatePowerUp(t, g, me, early, brawlerPay)
	if g.Turn.Extra {
		t.Fatal("setup: this is already the extra turn")
	}

	endTurn(t, g)
	assertTurnOf(t, g, 0, true, "the extra turn")
	powerUpMain(t, g)
	err := refusedPowerUp(t, g, me, banned, brawlerPay)
	if !errors.Is(err, game.ErrCantActivate) {
		t.Fatalf("my power-up during the extra turn: %v, want ErrCantActivate", err)
	}
	if !strings.Contains(err.Error(), "Power-up abilities can't be activated") {
		t.Errorf("the refusal %q does not say which clause refused", err)
	}
	// Anyone's, and a non-power-up ability, asked of the gate directly.
	var mineGate, theirGate, plainGate error
	g.WithWriteLock(func() {
		for _, c := range g.Battlefield.Cards {
			switch c.InstanceID {
			case banned:
				mineGate = g.ActivationGateLocked(me.ID, c, game.ZoneBattlefield, game.ActivationAbility{Label: "x", PowerUp: true})
				plainGate = g.ActivationGateLocked(me.ID, c, game.ZoneBattlefield, game.ActivationAbility{Label: "x"})
			case theirs:
				theirGate = g.ActivationGateLocked(opp.ID, c, game.ZoneBattlefield, game.ActivationAbility{Label: "x", PowerUp: true})
			}
		}
	})
	if !errors.Is(mineGate, game.ErrCantActivate) || !errors.Is(theirGate, game.ErrCantActivate) {
		t.Errorf("gate: mine = %v, opponent's = %v, want both ErrCantActivate", mineGate, theirGate)
	}
	if plainGate != nil {
		t.Errorf("an ability that is not a power-up was refused: %v", plainGate)
	}

	// The turn after the extra one is a plain turn again.
	endTurn(t, g)
	assertTurnOf(t, g, 1, false, "the turn after the extra turn")
	powerUpMain(t, g)
	activatePowerUp(t, g, opp, theirs, brawlerPay)
	if got := plusOnes(g, theirs); got != 2 {
		t.Errorf("the opponent's power-up after the extra turn: %d counters, want 2", got)
	}
}

// The extra turn's ban is queued state: a restore taken between the
// activation and the extra turn still bans it.
func TestKangBanSurvivesARestoreBeforeTheExtraTurn(t *testing.T) {
	g := newCatalogGame(t)
	powerUpMain(t, g)
	me := g.Seats[0]
	kang := pushPowerUpPermanent(g, me.ID, "Kang the Conqueror", kangType, "{2}{U}{U}", kangOracle)
	banned := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	activatePowerUp(t, g, me, kang, "{C}{C}{C}{C}{C}{U}{U}{U}")

	restored := restoreThroughJSON(t, g)
	rme := restored.PlayerByIDForEffect(me.ID)
	endTurn(t, restored)
	assertTurnOf(t, restored, 0, true, "the restored extra turn")
	powerUpMain(t, restored)
	if err := refusedPowerUp(t, restored, rme, banned, brawlerPay); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("a restored game lost the ban: %v, want ErrCantActivate", err)
	}
}

// --- Quinjet Technician --------------------------------------------------

func quinjetTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, _ := spendTable(t)
	q := pushCatalogPermanent(g, me.ID, "Quinjet Technician", "Creature — Human Artificer", quinjetOracle, false)
	return g, me, q
}

func tapQuinjetForPowerUpMana(t *testing.T, g *game.Game, me *game.Player, q uuid.UUID) {
	t.Helper()
	if err := g.ActivateManaAbility(me.ID, q, 1, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Quinjet Technician for {R}{R}: %v", err)
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool = %v, want {R}{R}", me.ManaPool)
	}
	for _, tok := range me.ManaPool {
		if len(tok.Restrictions) != 1 || tok.Restrictions[0] != game.ManaRestrictPowerUp {
			t.Fatalf("token %+v carries restrictions %v, want [%s]", tok, tok.Restrictions, game.ManaRestrictPowerUp)
		}
	}
}

// The {R}{R} pays a power-up ability's cost: {R}{R} and {C}{C}{W} are the
// {4}{W} of a Brave Brawler's power-up.
func TestQuinjetTechnicianManaPaysAPowerUpAbility(t *testing.T) {
	g, me, q := quinjetTable(t)
	brawler := pushPowerUpPermanent(g, me.ID, "Brave Brawler", brawlerType, "{1}{W}", braveBrawlerOracle)
	tapQuinjetForPowerUpMana(t, g, me, q)
	activatePowerUp(t, g, me, brawler, "{C}{C}{W}")
	if got := plusOnes(g, brawler); got != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", got)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool = %v, want everything spent", me.ManaPool)
	}
}

// It pays nothing else: not a spell, not a creature's ordinary ability.
func TestQuinjetTechnicianPowerUpManaPaysNothingElse(t *testing.T) {
	g, me, q := quinjetTable(t)
	rats := pushCatalogPermanent(g, me.ID, "Crypt Rats", "Creature — Rat", cryptRatsOracle, false)
	tapQuinjetForPowerUpMana(t, g, me, q)

	cost, _ := game.ParseCost("{R}")
	if me.ManaPool.CanPayFor(cost, 0, game.ManaSpendForAbility(mustCard(t, g, rats))) {
		t.Error("power-up-only mana pays an ordinary activated ability — stronger than printed")
	}
	refusedForMana(t, g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true, XValue: 1}),
		"{R} for Crypt Rats' {X}")
	bolt := handSpell(me, "Lightning Bolt", "Instant", "{R}")
	refusedForMana(t, g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}), "{R} for a spell")
	if me.ManaPool.Total() != 2 {
		t.Errorf("pool = %v, want both mana still floating after the refusals", me.ManaPool)
	}
}

// Its plain {R} is ordinary mana.
func TestQuinjetTechnicianPlainRowIsOrdinaryMana(t *testing.T) {
	g, me, q := quinjetTable(t)
	if err := g.ActivateManaAbility(me.ID, q, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap Quinjet Technician for {R}: %v", err)
	}
	bolt := handSpell(me, "Lightning Bolt", "Instant", "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("{R} from the plain row does not cast a spell: %v", err)
	}
}

// mustCard is the battlefield card named by id.
func mustCard(t *testing.T, g *game.Game, id uuid.UUID) game.Card {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == id {
			return c
		}
	}
	t.Fatalf("%s is not on the battlefield", id)
	return game.Card{}
}
