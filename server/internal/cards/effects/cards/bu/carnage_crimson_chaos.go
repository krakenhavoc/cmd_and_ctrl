package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Carnage, Crimson Chaos — Legendary Creature — Symbiote Villain
// {2}{B}{R}, 4/3 (Marvel's Spider-Man):
//
//	"Trample
//	 When Carnage enters, return target creature card with mana value 3 or
//	 less from your graveyard to the battlefield. It gains "This creature
//	 attacks each combat if able" and "When this creature deals combat
//	 damage to a player, sacrifice it."
//	 Mayhem {B}{R} (You may cast this card from your graveyard for {B}{R}
//	 if you discarded it this turn. Timing rules still apply.)"
//
// Mayhem is the engine's (CR 702.187, Rocket-Powered Goblin Glider). The
// return is Realmbreaker's: the card enters from the graveyard, and the two
// quoted abilities are one granted bundle (ADR 0093) on that new object
// for as long as it stays on the battlefield (an indefinite layer-6 grant,
// CR 611.2a), so it dies carrying them but a later reanimation is a new
// object without them (CR 400.7). The requirement is AttacksEachCombat
// (CR 508.1d) with the creature as its own source.
//
// No simplification.
const carnageGrant = "carnage-crimson-chaos/reanimated"

func init() {
	effects.Register(effects.Spec{
		OracleID:         "c0efafc6-8ffb-4908-ae64-3f391722370a",
		Name:             "Carnage, Crimson Chaos",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"trample"},
		CastableZonesIf:  []game.ConditionalCastZone{effects.MayhemCastableZone()},
		AlternativeCosts: []game.AlternativeCost{effects.Mayhem("{B}{R}")},
		Grants: []effects.AbilityGrant{{
			Key:    carnageGrant,
			Static: []game.StaticAbility{effects.AttacksEachCombat()},
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Carnage — sacrifice this creature",
					func(g *game.Game, item *game.StackItem) error {
						return effects.SacrificePermanent{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
					}),
			},
			Text: "This creature attacks each combat if able. When this creature deals combat damage to a player, sacrifice it.",
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Carnage — return target creature card with mana value 3 or less from your graveyard to the battlefield", carnageReturn),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", effects.Creature(), effects.ManaValueLE(3), effects.YouOwn()),
			),
		},
	})
}

// carnageReturn puts the chosen card onto the battlefield and gives the new
// object the granted bundle for as long as it remains.
func carnageReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
		return err
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.GrantAbilitiesMod(carnageGrant)},
		Duration: game.IndefiniteDuration(),
		Label:    "Carnage — gains \"attacks each combat if able\" and the sacrifice trigger",
	}.Apply(ctx)
}
