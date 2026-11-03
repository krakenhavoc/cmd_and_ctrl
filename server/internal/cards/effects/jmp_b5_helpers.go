package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jmp_b5_helpers.go — shared pieces of the Jumpstart Selesnya deck
// request's fifth batch (#4893). Its own file so a parallel batch never
// collides on a shared one.

// jmpB5PowerIsCreaturesYouControl is "<this>'s power is equal to the
// number of creatures you control" (Ironroot Warlord): a Layer 7a
// characteristic-defining ability, counted as the layers run, so a
// creature entering or leaving changes it at once.
func jmpB5PowerIsCreaturesYouControl(label string) game.StaticAbility {
	return game.StaticAbility{
		Label:     label,
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7A_CDA,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			c.Power = b04CreaturesControlled(g, source.Controller)
		},
	}
}

// jmpB5OtherCreaturesWithACounter counts the other creatures the
// source's controller controls that have a +1/+1 counter on them (High
// Sentinels of Arashin). Counters are instance state, not a
// characteristic, so reading them inside a layer pass is safe.
func jmpB5OtherCreaturesWithACounter(g *game.Game, source *game.Card) int {
	if g == nil || g.Battlefield == nil || source == nil {
		return 0
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != source.InstanceID && c.Controller == source.Controller && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
			n++
		}
	}
	return n
}

// jmpB5CreaturesYouControlWithACounter snapshots the creatures
// `controller` controls that have a +1/+1 counter on them, so the
// caller may add counters while it walks them (Lifecrafter's Gift).
func jmpB5CreaturesYouControlWithACounter(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && c.Counters[game.CounterPlusOne] > 0 {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// jmpB5NontokenCreaturesYouControl counts the nontoken creatures
// `controller` controls as the effect resolves (Lena, Selfless Champion).
func jmpB5NontokenCreaturesYouControl(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.IsCreature() && !c.IsToken() {
			n++
		}
	}
	return n
}
