package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// least_power.go — "destroy the creature with the least power. It can't
// be regenerated. If two or more creatures are tied for least power, you
// choose one of them." (Drop of Honey, Porphyry Nodes; ADR 0107 PR 1,
// #1858).
//
// Append-only, per the shared-vocabulary rule.

// destroyCreatureWithLeastPower is the whole upkeep effect. It reads
// every creature on the battlefield as the ability resolves (CR 608.2h),
// anyone's. With one creature at the least power it is destroyed with no
// question asked; with a tie the ability's controller picks one of the
// tied creatures. Destruction ignores regeneration (CR 701.19c). No
// creature on the battlefield: nothing happens.
func destroyCreatureWithLeastPower(g *game.Game, item *game.StackItem) error {
	tied := creaturesWithLeastPower(g)
	switch len(tied) {
	case 0:
		return nil
	case 1:
		return DestroyTarget{Target: tied[0], CantBeRegenerated: true}.Apply(NewContext(g, item))
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Choose the creature with the least power to destroy",
		Cards:    tied,
		Min:      1,
		Max:      1,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 || !onBattlefield(g, picked[0]) {
				return nil
			}
			return DestroyTarget{Target: picked[0], CantBeRegenerated: true}.Apply(NewContext(g, item))
		},
	})
	return nil
}

// creaturesWithLeastPower lists the creatures on the battlefield tied
// for the least current power, in battlefield order.
func creaturesWithLeastPower(g *game.Game) []uuid.UUID {
	if g.Battlefield == nil {
		return nil
	}
	var out []uuid.UUID
	least := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() {
			continue
		}
		p := c.CurrentPower()
		switch {
		case len(out) == 0 || p < least:
			least = p
			out = append(out[:0], c.InstanceID)
		case p == least:
			out = append(out, c.InstanceID)
		}
	}
	return out
}
