package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// licid_test.go — #5781: a licid end to end through the real engine. Its
// ability makes it an Aura on the target with its static applying, the
// player who activated it may pay to end the effect (CR 116.2c), and all
// of it survives a snapshot and restore.

const licidTestOracle = "test-licid"

// licidCreatureSpec is "enchant creature" / "target creature".
func licidCreatureSpec() *TargetSpec {
	return &TargetSpec{
		Mode: "creature", Label: "target creature",
		Zones: []ZoneKind{ZoneBattlefield},
		CardOK: func(_ *Game, _ uuid.UUID, c Card, _ ZoneKind) bool {
			return c.IsCreature()
		},
		Min: 1, Max: 1,
	}
}

// licidDef is a licid as effects.LicidAbility and effects.LicidForm build
// one, with "Enchanted creature gets +1/+1" as what the Aura does.
func licidDef() *CardDef {
	self := func(target *Card, _ *Game, source *Card) bool {
		return source != nil && target.InstanceID == source.InstanceID && source.ReturnedAura
	}
	removal := OwnAbilityRemoval{Slot: AbilitySlotActivated, Row: 0}
	return &CardDef{
		ReturnedAura: licidCreatureSpec(),
		Static: []StaticAbility{
			{Label: "It's an Aura enchantment.", Layer: Layer4Type, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Types = []string{"Enchantment"}
					c.SetSubtypes([]string{"Aura"})
				}},
			{Label: "It loses this ability.", Layer: Layer6Ability, AppliesTo: self,
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.RemovedOwnAbilities = append(c.RemovedOwnAbilities, removal)
				}},
			{Label: "Enchanted creature gets +1/+1.", Layer: Layer7PT, SubLayer: SubLayer7C_Modify,
				AppliesTo: func(target *Card, _ *Game, source *Card) bool {
					return source != nil && source.AttachedTo.Kind == TargetCard && source.AttachedTo.ID == target.InstanceID
				},
				Apply: func(c *Characteristic, _ *Card, _ *Game, _ *Card) {
					c.Power++
					c.Toughness++
				}},
		},
		Activated: []ActivatedAbilityShape{{
			Label:   "{B}, {T}: This creature loses this ability and becomes an Aura enchantment with enchant creature. Attach it to target creature. You may pay {B} to end this effect.",
			Cost:    AbilityCost{Mana: "{B}", Tap: true},
			Targets: licidCreatureSpec(),
			Effect: func(g *Game, item *StackItem) error {
				for _, t := range item.Targets {
					if t.Kind == TargetCard {
						_, err := g.BecomeLicidAuraForEffect(item.SourceCardID, item.Controller, "{B}", t.ID)
						return err
					}
				}
				return nil
			},
		}},
	}
}

type licidTable struct {
	g        *Game
	me, opp  *Player
	licid    uuid.UUID
	bear     uuid.UUID
	bearBase int
}

func newLicidTable(t *testing.T) licidTable {
	t.Helper()
	withCatalogDefs(t, map[string]*CardDef{licidTestOracle: licidDef()})
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lc := NewCard("Test Licid", me.ID)
	lc.TypeLine = "Creature — Licid"
	lc.OracleID = licidTestOracle
	lc.Controller = me.ID
	lc.Power, lc.Toughness = 1, 1
	licid := pushPermanent(g, lc)
	bear := pushPermanent(g, bearNamed(opp.ID, "Bear"))
	return licidTable{g: g, me: me, opp: opp, licid: licid, bear: bear, bearBase: 2}
}

func (tb licidTable) card(t *testing.T, g *Game, id uuid.UUID) Card {
	t.Helper()
	var out Card
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		c := findBattlefieldCard(g, id)
		if c == nil {
			t.Fatalf("%v is not on the battlefield", id)
		}
		out = *c
	})
	return out
}

// activate activates the licid's ability at the bear and resolves it.
func (tb licidTable) activate(t *testing.T) {
	t.Helper()
	if err := tb.g.ActivateCatalogAbility(tb.me.ID, tb.licid, 0, ActivateAbilityParams{
		Targets: []TargetRef{{Kind: TargetCard, ID: tb.bear}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	settleStack(t, tb.g)
}

// assertAura checks the licid is an Aura on the bear, not a creature, with
// its ability gone and its static on the bear.
func (tb licidTable) assertAura(t *testing.T, g *Game) {
	t.Helper()
	l := tb.card(t, g, tb.licid)
	if !l.IsAura() || l.IsCreature() || l.HasSubtype("Licid") {
		t.Errorf("the licid is not an Aura enchantment that is no creature")
	}
	if l.AttachedTo.Kind != TargetCard || l.AttachedTo.ID != tb.bear {
		t.Errorf("the licid is attached to %+v, want the bear", l.AttachedTo)
	}
	if b := tb.card(t, g, tb.bear); b.CurrentPower() != tb.bearBase+1 {
		t.Errorf("the bear's power is %d, want %d", b.CurrentPower(), tb.bearBase+1)
	}
	var rows int
	g.WithWriteLock(func() { rows = licidRows(g, tb.licid) })
	if rows != 0 {
		t.Errorf("the licid still has %d activated abilities, want 0 (it lost this ability)", rows)
	}
}

func TestLicidBecomesAnAuraAndPayingEndsTheEffect(t *testing.T) {
	tb := newLicidTable(t)
	g := tb.g
	tb.activate(t)
	tb.assertAura(t, g)

	// The offer: to the player who activated it, and to nobody else.
	var mine, theirs []SpecialAction
	g.WithWriteLock(func() {
		l := *findBattlefieldCard(g, tb.licid)
		mine = g.SpecialActionsOfferedLocked(tb.me.ID, l, ZoneBattlefield)
		theirs = g.SpecialActionsOfferedLocked(tb.opp.ID, l, ZoneBattlefield)
	})
	if len(mine) != 1 || mine[0].Kind != SpecialActionEndEffect || mine[0].Cost != "{B}" {
		t.Fatalf("offers to the activator = %+v, want end_effect for {B}", mine)
	}
	if len(theirs) != 0 {
		t.Errorf("offers to the opponent = %+v, want none", theirs)
	}
	if err := g.PerformSpecialAction(tb.opp.ID, tb.licid, SpecialActionEndEffect, SpecialActionParams{}); err == nil {
		t.Fatal("the opponent ended the licid's effect")
	}

	// A snapshot and restore keeps the Aura form and the offer.
	restored, err := g.CaptureSnapshot().RestoreStrict()
	if err != nil {
		t.Fatalf("RestoreStrict: %v", err)
	}
	tb.assertAura(t, restored)
	if l := tb.card(t, restored, tb.licid); l.AuraFormEndBy != tb.me.ID || l.AuraFormEndCost != "{B}" {
		t.Errorf("restored offer %v %q, want %v {B}", l.AuraFormEndBy, l.AuraFormEndCost, tb.me.ID)
	}

	// Paying ends it, on the restored table: a creature again, unattached,
	// with its ability back, and the bear back to its own size.
	if err := restored.PerformSpecialAction(tb.me.ID, tb.licid, SpecialActionEndEffect, SpecialActionParams{}); err != nil {
		t.Fatalf("end_effect: %v", err)
	}
	l := tb.card(t, restored, tb.licid)
	if !l.IsCreature() || l.IsAura() || !l.HasSubtype("Licid") {
		t.Errorf("after ending, the licid is not the Licid creature")
	}
	if l.IsAttached() {
		t.Errorf("after ending, the licid is still attached to %+v", l.AttachedTo)
	}
	if b := tb.card(t, restored, tb.bear); b.CurrentPower() != tb.bearBase {
		t.Errorf("after ending, the bear's power is %d, want %d", b.CurrentPower(), tb.bearBase)
	}
	var rows int
	restored.WithWriteLock(func() { rows = licidRows(restored, tb.licid) })
	if rows != 1 {
		t.Errorf("after ending, the licid has %d activated abilities, want its 1 back", rows)
	}
	if err := restored.PerformSpecialAction(tb.me.ID, tb.licid, SpecialActionEndEffect, SpecialActionParams{}); !errors.Is(err, ErrCardNotFound) && !errors.Is(err, ErrSpecialActionNotOffered) {
		t.Errorf("a second end_effect = %v, want it refused", err)
	}
}

// An Aura whose creature leaves is put into its owner's graveyard
// (CR 704.5m): the licid does not become a creature again.
func TestLicidAuraGoesToTheGraveyardWithItsCreature(t *testing.T) {
	tb := newLicidTable(t)
	g := tb.g
	tb.activate(t)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(tb.bear); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(tb.licid) {
		t.Fatal("the licid stayed on the battlefield when its creature died")
	}
	if !tb.me.Graveyard.Contains(tb.licid) {
		t.Error("the licid is not in its owner's graveyard")
	}
}

// CR 116.2c: any time the player has priority, so only with priority.
func TestLicidEndEffectNeedsPriority(t *testing.T) {
	tb := newLicidTable(t)
	g := tb.g
	tb.activate(t)
	var ok bool
	g.WithWriteLock(func() {
		l := *findBattlefieldCard(g, tb.licid)
		ok = g.SpecialActionTimingOKLocked(tb.me.ID, l, SpecialActionEndEffect)
	})
	if !ok {
		t.Fatal("the activator holds priority and may end the effect")
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.PerformSpecialAction(tb.me.ID, tb.licid, SpecialActionEndEffect, SpecialActionParams{}); !errors.Is(err, ErrSpecialActionTiming) {
		t.Errorf("end_effect without priority = %v, want ErrSpecialActionTiming", err)
	}
}

// licidRows is how many activated rows the permanent has now.
// Caller must hold g.mu.
func licidRows(g *Game, id uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	rows, _ := activatedAbilityRows(findBattlefieldCard(g, id), false)
	return len(rows)
}
