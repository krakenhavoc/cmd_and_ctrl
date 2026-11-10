package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sygg, Wanderwine Wisdom // Sygg, Wanderbrine Shield — {1}{U} Legendary
// Creature — Merfolk Wizard 2/2 // Legendary Creature — Merfolk Rogue 2/2
// (Lorwyn Eclipsed, #4458):
//
//	Front: "Sygg can't be blocked.
//	        Whenever this creature enters or transforms into Sygg,
//	        Wanderwine Wisdom, target creature gains "Whenever this
//	        creature deals combat damage to a player or planeswalker, draw
//	        a card" until end of turn.
//	        At the beginning of your first main phase, you may pay {W}. If
//	        you do, transform Sygg."
//	Back:  "Sygg can't be blocked.
//	        Whenever this creature transforms into Sygg, Wanderbrine
//	        Shield, target creature you control gains protection from each
//	        color until your next turn.
//	        At the beginning of your first main phase, you may pay {U}. If
//	        you do, transform Sygg."
//
// Brigid, Clachan's Heart's shape. The front's grant is Requiem
// Monolith's duration grant: the quoted trigger is a catalog bundle the
// target creature has until end of turn, so it is that creature's own
// ability and "draw a card" is its controller's draw, whoever that is. It
// fires on combat damage to a player or to a planeswalker (Grateful
// Apparition's condition), once per combat damage step it deals damage
// in.
//
// The back's protection from each color is five keywords, one per color
// (CR 702.16, Akroma's Will), on the target until the start of the
// trigger's controller's next turn (DurationUntilYourNextTurn).
//
// No simplification.
func init() {
	const oracle = "9cc91ac5-6f2b-4fc0-b810-cdc92eebe6de"
	cantBeBlocked := RestrictSelf(game.CantBeBlocked)
	Register(Spec{
		OracleID:     oracle,
		Name:         "Sygg, Wanderwine Wisdom",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{cantBeBlocked},
		Grants: []AbilityGrant{{
			Key: syggWanderwineGrant,
			Triggered: []game.TriggeredAbility{
				On(game.EventDealDamage, ThisDealtCombatDamageToAPlayerOrPlaneswalker,
					"Sygg, Wanderwine Wisdom — draw a card", Do(DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player or planeswalker, draw a card.",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OnAny([]game.EventKind{game.EventETB, game.EventTransform}, ThisEntersOrTransformsInto(0),
					"Sygg, Wanderwine Wisdom — target creature gains \"draw a card on combat damage\" until end of turn",
					syggWanderwineGrantToTarget),
				TargetCreature("target creature")),
			PayToTransformAtYourFirstMain("Sygg, Wanderwine Wisdom", "Sygg", "{W}"),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Sygg, Wanderbrine Shield",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{cantBeBlocked},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventTransform, ThisTransformedInto(1),
					"Sygg, Wanderbrine Shield — target creature you control gains protection from each color until your next turn",
					syggWanderbrineProtect),
				TargetCreature("target creature you control", YouControl())),
			PayToTransformAtYourFirstMain("Sygg, Wanderbrine Shield", "Sygg", "{U}"),
		},
	})
}

// syggWanderwineGrant is the key of the granted combat-damage draw, an
// on-disk identity.
const syggWanderwineGrant = "sygg-wanderwine-wisdom/draw-on-combat-damage"

// syggWanderwineGrantToTarget gives the target creature the draw trigger
// until end of turn.
func syggWanderwineGrantToTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return GrantAbilitiesFor{
		Target: target,
		Keys:   []string{syggWanderwineGrant},
		Label:  "Sygg, Wanderwine Wisdom — draw a card on combat damage until end of turn",
	}.Apply(ctx)
}

// syggWanderbrineProtect gives the target creature protection from each
// color until the trigger's controller's next turn.
func syggWanderbrineProtect(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := b16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	var colors []string
	for _, c := range game.AllColors {
		colors = append(colors, game.ProtectionFromColor(c))
	}
	return ScopedEffectFor{
		Target:   target,
		Mods:     []game.Mod{game.AddKeywordsMod(colors...)},
		Duration: DurationUntilYourNextTurn(ctx, item.Controller),
		Label:    "Sygg, Wanderbrine Shield — protection from each color until your next turn",
	}.Apply(ctx)
}
