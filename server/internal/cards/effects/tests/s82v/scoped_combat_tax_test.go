package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// scoped_combat_tax_test.go — #6191's constructors, asked the way the
// engine's pricer asks them (CR 508.1h, CR 509.1d).

func TestScopedCombatTaxConstructors(t *testing.T) {
	me, them := uuid.New(), uuid.New()
	myr := game.Card{InstanceID: uuid.New(), Controller: me, Counters: map[string]int{game.CounterPlusOne: 2}}
	other := game.Card{InstanceID: uuid.New(), Controller: me}

	atk := effects.CantAttackUnlessYouPayPerCounter(game.CounterPlusOne, "Myr attack")
	if !atk.OnAttacker {
		t.Error("the Myr's attack tax is not the attacker's own")
	}
	q := game.AttackTaxQuery{Source: myr, Attacker: myr}
	if !atk.AppliesTo(q) || atk.ManaCost(q) != "{2}" {
		t.Errorf("Myr with two counters: applies %v, cost %q; want {2}", atk.AppliesTo(q), atk.ManaCost(q))
	}
	if atk.AppliesTo(game.AttackTaxQuery{Source: myr, Attacker: other}) {
		t.Error("the Myr's attack tax binds another creature")
	}
	if c := atk.ManaCost(game.AttackTaxQuery{Source: other, Attacker: other}); c != "" {
		t.Errorf("no counters costs %q, want nothing", c)
	}

	blk := effects.CantBlockUnlessYouPayPerCounter(game.CounterPlusOne, "Myr block")
	bq := game.BlockTaxQuery{Source: myr, Blocker: myr}
	if !blk.AppliesTo(bq) || blk.ManaCost(bq) != "{2}" {
		t.Errorf("Myr blocking: %q, want {2}", blk.ManaCost(bq))
	}

	hip := effects.CantBlockAttackersUnlessYouPay("{1}", effects.PowerAtLeast(3), "Hipparion")
	small := game.Card{InstanceID: uuid.New(), Controller: them, Power: 2, Toughness: 2}
	big := game.Card{InstanceID: uuid.New(), Controller: them, Power: 3, Toughness: 3}
	h := game.Card{InstanceID: uuid.New(), Controller: me}
	if !hip.PerAttacker {
		t.Error("Hipparion's tax is not per attacker")
	}
	if hip.AppliesTo(game.BlockTaxQuery{Source: h, Blocker: h, Attacker: small}) {
		t.Error("Hipparion pays to block a 2/2")
	}
	if !hip.AppliesTo(game.BlockTaxQuery{Source: h, Blocker: h, Attacker: big}) {
		t.Error("Hipparion blocks a 3/3 for free")
	}

	aura := game.Card{InstanceID: uuid.New(), Controller: me, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: big.InstanceID}}
	pres := effects.EnchantedCantBeBlockedUnlessPay("{3}", "Awesome Presence")
	if !pres.PerAttacker || !pres.AppliesTo(game.BlockTaxQuery{Source: aura, Blocker: h, Attacker: big}) || pres.ManaCost(game.BlockTaxQuery{}) != "{3}" {
		t.Error("blocking the enchanted creature is not {3} per blocker")
	}
	if pres.AppliesTo(game.BlockTaxQuery{Source: aura, Blocker: h, Attacker: small}) {
		t.Error("blocking another creature pays Awesome Presence")
	}

	heat := effects.CantBlockYourCreaturesUnlessPayLife(1, func(c game.Card) bool { return !c.HasColor("U") }, "Heat Wave")
	wave := game.Card{InstanceID: uuid.New(), Controller: them}
	red := game.Card{InstanceID: uuid.New(), Controller: me, Colors: []string{"R"}}
	blue := game.Card{InstanceID: uuid.New(), Controller: me, Colors: []string{"U"}}
	if !heat.AppliesTo(game.BlockTaxQuery{Source: wave, Blocker: red, Blocking: []game.Card{small}}) || heat.LifeCost(game.BlockTaxQuery{}) != 1 {
		t.Error("a nonblue creature blocking the Heat Wave controller's creature pays no life")
	}
	if heat.AppliesTo(game.BlockTaxQuery{Source: wave, Blocker: blue, Blocking: []game.Card{small}}) {
		t.Error("a blue creature pays Heat Wave's life")
	}
	if heat.AppliesTo(game.BlockTaxQuery{Source: wave, Blocker: red, Blocking: []game.Card{h}}) {
		t.Error("blocking someone else's creature pays Heat Wave's life")
	}
}
