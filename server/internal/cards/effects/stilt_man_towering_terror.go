package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stilt-Man, Towering Terror — {2}{R} Legendary Creature — Human Rogue Villain
// 4/2 (Marvel Super Heroes Commander, #4455):
//
//	"Reach
//	 Whenever one or more Villains you control deal combat damage to a player,
//	 gain control of target noncreature, nonland permanent that player
//	 controls until the end of your next turn. It gains 'This permanent can't
//	 be sacrificed' until the end of your next turn."
//
// The trigger collapses per damaged player (CR 603.2c, The Thing, Ben Grimm's
// batch). Its target is chosen from the permanents THAT player controls as it
// goes on the stack. GainControl lasts through the end of the controller's
// next turn (CR 611.2); the sentence after it is a granted ABILITY, so it is
// a layer-6 bundle (Alexios's CantBeSacrificed static) that an ability-removal
// effect takes away, unlike the rule-form CantBeSacrificedThisTurn.
//
// No simplification.
const stiltManGrantKey = "stilt-man-towering-terror/cant-be-sacrificed"

func init() {
	Register(Spec{
		OracleID:        "8ae1bda5-8f70-48ec-9df0-89e58a4ab21c",
		Name:            "Stilt-Man, Towering Terror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Grants: []AbilityGrant{{
			Key:    stiltManGrantKey,
			Static: []game.StaticAbility{CantBeSacrificed()},
			Text:   "This permanent can't be sacrificed.",
		}},
		Triggered: []game.TriggeredAbility{
			withTargetsFrom(
				OncePerBatchPerPlayer(On(game.EventDealDamage, stiltManVillainsHitAPlayer,
					"Stilt-Man, Towering Terror — gain control of target noncreature, nonland permanent that player controls until the end of your next turn; it can't be sacrificed",
					stiltManSteal)),
				func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
					return TargetPermanent("target noncreature, nonland permanent that player controls",
						Noncreature(), Nonland(), ControlledBy(tc.Event.Target))
				}),
		},
	})
}

// stiltManVillainsHitAPlayer is "a Villain you control dealt combat damage to
// a player".
func stiltManVillainsHitAPlayer(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
		return false
	}
	dealer, ok := g.LookupCardForEffect(ev.Source)
	return ok && dealer.Controller == source.Controller && dealer.HasSubtype(source.TextSubtype("Villain"))
}

// stiltManSteal takes the permanent and gives it the granted ability, both
// through the end of the controller's next turn.
func stiltManSteal(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		until := g.UntilEndOfYourNextTurnDuration(item.Controller)
		if err := (GainControl{Target: t.ID, Duration: until, Label: "Stilt-Man, Towering Terror — gain control"}).Apply(ctx); err != nil {
			return err
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.GrantAbilitiesMod(stiltManGrantKey)},
			Duration: until,
			Label:    "Stilt-Man, Towering Terror — gains \"This permanent can't be sacrificed\"",
		}.Apply(ctx)
	}
	return nil
}
