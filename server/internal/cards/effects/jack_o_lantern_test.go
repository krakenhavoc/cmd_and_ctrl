package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// jack_o_lantern_test.go — #5896: Jack-o'-Lantern, and the mana ability
// that functions from a graveyard (CR 113.6b) that it needed.

const jackOLanternOracle = "192b5dd7-f420-4d5b-86d6-66e86fa67f41"

// jackInGraveyard puts a Jack-o'-Lantern into p's graveyard, known to
// every seat as a graveyard card is.
func jackInGraveyard(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	known := map[uuid.UUID]bool{}
	for _, s := range g.Seats {
		known[s.ID] = true
	}
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: "Jack-o'-Lantern", TypeLine: "Artifact",
		ManaCost: "{1}", OracleID: jackOLanternOracle,
		Owner: p.ID, Controller: p.ID, KnownBy: known,
	})
	return id
}

// jackManaMove reports whether the enumerator offers the graveyard mana
// activation of `jack` to `seat`.
func jackManaMove(g *game.Game, seat, jack uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeActivateManaAbility && m.Source == jack {
			return true
		}
	}
	return false
}

// Stand-alone: the {1} comes off a Mountain the activation auto-taps,
// the card leaves the graveyard for exile, and the chosen colour lands
// in the pool.
func TestJackOLanternGraveyardManaAbilityAddsTheChosenColor(t *testing.T) {
	g, me, _ := spendTable(t)
	mountain := basics(g, me, "Mountain", 1)[0]
	jack := jackInGraveyard(g, me)

	if !jackManaMove(g, me.ID, jack) {
		t.Fatal("the enumerator does not offer the graveyard mana ability")
	}
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{AutoTap: true, Colors: []string{"U"}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	if !isTapped(g, mountain) {
		t.Error("the Mountain did not pay the {1}")
	}
	if me.Graveyard.Contains(jack) || !exileHas(g, jack) {
		t.Error("Jack-o'-Lantern did not move from the graveyard to exile")
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "U" {
		t.Fatalf("pool = %+v, want one {U}", units)
	}
}

// The {1} is paid out of floating mana too, with no auto-tap.
func TestJackOLanternPaysItsCostFromThePool(t *testing.T) {
	g, me, _ := spendTable(t)
	jack := jackInGraveyard(g, me)
	mountain := basics(g, me, "Mountain", 1)[0]
	if err := g.ActivateManaAbility(me.ID, mountain, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Mountain: %v", err)
	}
	if err := g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{Colors: []string{"G"}}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	units := me.ManaPool.Units()
	if len(units) != 1 || units[0].Color != "G" {
		t.Fatalf("pool = %+v, want the {R} spent and one {G} made", units)
	}
}

// Nothing to pay the {1} with: refused, and the card stays put.
func TestJackOLanternRefusedWithoutTheOne(t *testing.T) {
	g, me, _ := spendTable(t)
	jack := jackInGraveyard(g, me)

	if jackManaMove(g, me.ID, jack) {
		t.Error("the enumerator offers an activation nothing can pay for")
	}
	refusedForMana(t, g.ActivateManaAbility(me.ID, jack, 0, game.ManaAbilityParams{AutoTap: true}), "no mana")
	if !me.Graveyard.Contains(jack) || me.ManaPool.Total() != 0 {
		t.Error("the refused activation moved the card or made mana")
	}
}

// CR 113.6b: the ability functions from the graveyard only, and CR
// 108.4: only its owner may activate it there.
func TestJackOLanternManaAbilityOnlyFromItsOwnersGraveyard(t *testing.T) {
	g, me, them := spendTable(t)
	basics(g, me, "Mountain", 2)
	onBoard := pushCatalogPermanent(g, me.ID, "Jack-o'-Lantern", "Artifact", jackOLanternOracle, false)
	if err := g.ActivateManaAbility(me.ID, onBoard, 0, game.ManaAbilityParams{AutoTap: true}); !errors.Is(err, game.ErrActivationZoneNotAllowed) {
		t.Fatalf("activate on the battlefield: %v, want ErrActivationZoneNotAllowed", err)
	}
	jack := jackInGraveyard(g, me)
	basics(g, them, "Mountain", 1)
	if err := g.ActivateManaAbility(them.ID, jack, 0, game.ManaAbilityParams{AutoTap: true}); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("an opponent activating my graveyard card: %v, want ErrCardCallerMismatch", err)
	}
	if !me.Graveyard.Contains(jack) {
		t.Error("a refused activation moved the card")
	}
}

// While paying a cost (CR 605.3a): a {U} spell with one Mountain and
// Jack-o'-Lantern in the graveyard. The auto-tapper's costed pass taps
// the Mountain for the {1} and exiles Jack-o'-Lantern for the {U}.
func TestJackOLanternAutoTapPaysAColoredSpell(t *testing.T) {
	g, me, _ := spendTable(t)
	mountain := basics(g, me, "Mountain", 1)[0]
	jack := jackInGraveyard(g, me)
	spell := handSpell(me, "Opt-ish", "Instant", "{U}")

	if !castMove(g, me.ID, spell) {
		t.Fatal("the enumerator does not offer a {U} cast Jack-o'-Lantern can pay for")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("auto-tapped cast: %v", err)
	}
	if !isTapped(g, mountain) {
		t.Error("the Mountain did not pay Jack-o'-Lantern's {1}")
	}
	if !exileHas(g, jack) {
		t.Error("Jack-o'-Lantern was not exiled for the {U}")
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool has %d left, want 0", me.ManaPool.Total())
	}
}

// The costed pass runs only when no plain plan exists: a {R} spell
// with a Mountain is paid by the Mountain, and the card stays in the
// graveyard.
func TestJackOLanternIsNotSpentWhenALandPays(t *testing.T) {
	g, me, _ := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := jackInGraveyard(g, me)
	spell := handSpell(me, "Shock-ish", "Instant", "{R}")

	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !me.Graveyard.Contains(jack) {
		t.Error("Jack-o'-Lantern was exiled although the Mountain paid")
	}
}

// Jack-o'-Lantern makes as much as it costs: it never adds mana, so a
// {U}{R} spell off one Mountain is not payable.
func TestJackOLanternDoesNotMakeExtraMana(t *testing.T) {
	g, me, _ := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := jackInGraveyard(g, me)
	spell := handSpell(me, "Two-ish", "Instant", "{U}{R}")

	if castMove(g, me.ID, spell) {
		t.Error("a two-mana cast offered off one Mountain")
	}
	refusedForMana(t, g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, AutoTap: true}), "two off one")
	if !me.Graveyard.Contains(jack) {
		t.Error("the refused cast exiled Jack-o'-Lantern")
	}
}

// The view: the owner's frame carries the graveyard row, and the other
// seats' frames do not (the graveyard is public, the row is not).
func TestJackOLanternGraveyardRowIsTheOwners(t *testing.T) {
	g, me, them := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := jackInGraveyard(g, me)

	find := func(viewer uuid.UUID) *protocol.CardView {
		v := protocol.ViewOfGameFor(g, viewer.String())
		for _, s := range v.Seats {
			if s.ID != me.ID.String() {
				continue
			}
			for i := range s.Graveyard.Cards {
				if s.Graveyard.Cards[i].InstanceID == jack.String() {
					return &s.Graveyard.Cards[i]
				}
			}
		}
		return nil
	}
	mine := find(me.ID)
	if mine == nil || len(mine.ZoneManaAbilities) != 1 || mine.ZoneManaAbilities[0].Index != 0 {
		t.Fatalf("owner's graveyard card rows = %+v, want the one mana row", mine)
	}
	if theirs := find(them.ID); theirs == nil || theirs.Name != "Jack-o'-Lantern" || len(theirs.ZoneManaAbilities) != 0 {
		t.Errorf("an opponent's frame carries the owner's mana row: %+v", theirs)
	}
}

// The first ability: exile a card from an opponent's graveyard, draw.
func TestJackOLanternSacrificeExilesATargetAndDraws(t *testing.T) {
	g, me, them := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := pushCatalogPermanent(g, me.ID, "Jack-o'-Lantern", "Artifact", jackOLanternOracle, false)
	victim := pushGraveyardCard(g, them, "Gravecard")
	hand := len(me.Hand.Cards)

	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), AutoTap: true, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(jack) {
		t.Error("Jack-o'-Lantern was not sacrificed")
	}
	resolveTopOfStack(t, g)
	if !exileHas(g, victim) {
		t.Error("the target was not exiled")
	}
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, len(me.Hand.Cards))
	}
	// And the sacrificed card is now a graveyard mana source.
	basics(g, me, "Mountain", 1)
	if !jackManaMove(g, me.ID, jack) {
		t.Error("the sacrificed Jack-o'-Lantern does not offer its graveyard ability")
	}
}

// "Up to one": with no target it still draws.
func TestJackOLanternWithNoTargetJustDraws(t *testing.T) {
	g, me, _ := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := pushCatalogPermanent(g, me.ID, "Jack-o'-Lantern", "Artifact", jackOLanternOracle, false)
	hand := len(me.Hand.Cards)

	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{AutoTap: true, Strict: true}); err != nil {
		t.Fatalf("activate with no target: %v", err)
	}
	resolveTopOfStack(t, g)
	if len(me.Hand.Cards) != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, len(me.Hand.Cards))
	}
}

// CR 608.2b: the one target left the graveyard in response, so the
// ability does not resolve and draws nothing.
func TestJackOLanternWithItsTargetGoneDrawsNothing(t *testing.T) {
	g, me, them := spendTable(t)
	basics(g, me, "Mountain", 1)
	jack := pushCatalogPermanent(g, me.ID, "Jack-o'-Lantern", "Artifact", jackOLanternOracle, false)
	victim := pushGraveyardCard(g, them, "Gravecard")
	hand := len(me.Hand.Cards)

	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), AutoTap: true, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: them.ID}, game.ZoneRef{Kind: game.ZoneExile}, victim); err != nil {
		t.Fatalf("move the target away: %v", err)
	}
	resolveTopOfStack(t, g)
	if len(me.Hand.Cards) != hand {
		t.Errorf("hand %d → %d, want no draw from a fizzled ability", hand, len(me.Hand.Cards))
	}
}

// Strict refusals: the sacrifice ability needs its {1}, and a sick
// Jack-o'-Lantern is an artifact, so {T} is fine — but a tapped one is
// not.
func TestJackOLanternSacrificeRefusals(t *testing.T) {
	g, me, them := spendTable(t)
	jack := pushCatalogPermanent(g, me.ID, "Jack-o'-Lantern", "Artifact", jackOLanternOracle, true)
	victim := pushGraveyardCard(g, them, "Gravecard")
	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), AutoTap: true, Strict: true}); err == nil {
		t.Fatal("activated with nothing to pay the {1}")
	}
	if !battlefieldHas(g, jack) {
		t.Error("the refused activation sacrificed Jack-o'-Lantern")
	}
	basics(g, me, "Mountain", 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == jack {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, jack, 0, game.ActivateAbilityParams{Targets: cardRefs(victim), AutoTap: true, Strict: true}); err == nil {
		t.Fatal("a tapped Jack-o'-Lantern paid {T}")
	}
}
