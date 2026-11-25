package cardsap

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ascendant Spirit — Snow Creature — Spirit {U}, 1/1 (Kaldheim, #5505):
//
//	"{S}{S}: This creature becomes a Spirit Warrior with base power and
//	 toughness 2/3.
//	 {S}{S}{S}: If this creature is a Warrior, put a flying counter on it and
//	 it becomes a Spirit Warrior Angel with base power and toughness 4/4.
//	 {S}{S}{S}{S}: If this creature is an Angel, put two +1/+1 counters on it
//	 and it gains "Whenever this creature deals combat damage to a player,
//	 draw a card.""
//
// Figure of Destiny's shape, paid in {S} (CR 107.4h: only mana from a snow
// source pays it). No duration: each effect lasts while it is on the
// battlefield (CR 611.2a), a later type or base P/T set overwrites an
// earlier one, and the "if this creature is a …" is read as the ability
// resolves. The granted trigger is a bundle on the Spirit itself.
const ascendantSpiritDraw = "ascendant-spirit/draw-on-combat-damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "784c039d-a9b2-4b22-a937-c8f7b475e96a",
		Name:         "Ascendant Spirit",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: ascendantSpiritDraw,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Ascendant Spirit — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Activated: []effects.ActivatedAbility{
			ascendantSpiritStage("{S}{S}: This creature becomes a Spirit Warrior with base power and toughness 2/3.",
				"{S}{S}", "", []string{"Spirit", "Warrior"}, 2, 3, nil),
			ascendantSpiritStage("{S}{S}{S}: If this creature is a Warrior, put a flying counter on it and it becomes a Spirit Warrior Angel with base power and toughness 4/4.",
				"{S}{S}{S}", "Warrior", []string{"Spirit", "Warrior", "Angel"}, 4, 4, func(ctx *effects.Context) error {
					return effects.AddCounter{Target: ctx.Item.SourceCardID, Kind: game.CounterFlying, N: 1}.Apply(ctx)
				}),
			{
				Label:   "{S}{S}{S}{S}: If this creature is an Angel, put two +1/+1 counters on it and it gains \"Whenever this creature deals combat damage to a player, draw a card.\"",
				Purpose: game.Purpose{Answers: game.AnswerPump},
				Cost:    effects.ManaCost("{S}{S}{S}{S}"),
				Effect:  ascendantSpiritAngel,
			},
		},
	})
}

// ascendantSpiritStage is one of the two type-changing abilities. needs is
// the creature type the source must be as it resolves ("" for none); also
// is the ability's own extra step, run first.
func ascendantSpiritStage(label, cost, needs string, types []string, power, toughness int, also func(*effects.Context) error) effects.ActivatedAbility {
	return effects.ActivatedAbility{
		Label:   label,
		Purpose: game.Purpose{Answers: game.AnswerPump},
		Cost:    effects.ManaCost(cost),
		Effect: func(g *game.Game, item *game.StackItem) error {
			if !effects.SourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := effects.NewContext(g, item)
			if needs != "" && !ascendantSpiritIs(g, item, needs) {
				return nil
			}
			if also != nil {
				if err := also(ctx); err != nil {
					return err
				}
			}
			mods := append([]game.Mod{game.SetCreatureTypesMod(ctx.TextSubtypes(types...)...)},
				game.SetBasePTMods(power, toughness)...)
			return effects.ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     mods,
				Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
				Label:    "Ascendant Spirit — becomes a " + types[len(types)-1] + " " + fmt.Sprintf("%d/%d", power, toughness),
			}.Apply(ctx)
		},
	}
}

func ascendantSpiritIs(g *game.Game, item *game.StackItem, subtype string) bool {
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	return ok && src.HasSubtype(effects.NewContext(g, item).TextSubtype(subtype))
}

func ascendantSpiritAngel(g *game.Game, item *game.StackItem) error {
	if !effects.SourceIsStillThisPermanent(g, item) || !ascendantSpiritIs(g, item, "Angel") {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if err := (effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantAbilitiesFor{
		Target:   item.SourceCardID,
		Keys:     []string{ascendantSpiritDraw},
		Duration: g.PinnedTo(game.IndefiniteDuration(), item.SourceCardID),
		Label:    "Ascendant Spirit — draws on combat damage to a player",
	}.Apply(ctx)
}
