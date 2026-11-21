package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Symbiote Spider-Man — Legendary Creature — Symbiote Spider Hero
// {2}{U/B}, 2/4 (Marvel's Spider-Man promos, #4460):
//
//	"Whenever this creature deals combat damage to a player, look at that
//	 many cards from the top of your library. Put one of them into your
//	 hand and the rest into your graveyard.
//	 Find New Host — {2}{U/B}, Exile this card from your graveyard: Put a
//	 +1/+1 counter on target creature you control. It gains this card's
//	 other abilities. Activate only as a sorcery."
//
// The trigger reads the damage event's amount: "that many" is the damage
// dealt. The look is Wail of the Forgotten's (one into the hand, the rest
// into the graveyard).
//
// Find New Host works from the graveyard (CR 113.6) at sorcery speed; its
// cost exiles the card. "This card's other abilities" is the one other
// ability the card prints, the combat-damage trigger, declared once as a
// catalog bundle (ADR 0093, Feign Death's shape) that the resolution gives
// the target with no duration, for as long as it is that object (CR 611.2a,
// 400.7). On the creature the trigger is the creature's own: "this
// creature" is it, and its controller looks at their library.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc681ecf-2a52-4782-9a47-90f5a9f435d8",
		Name:         "Symbiote Spider-Man",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{symbioteSpiderManTrigger()},
		Grants: []effects.AbilityGrant{{
			Key:       symbioteSpiderManGrant,
			Triggered: []game.TriggeredAbility{symbioteSpiderManTrigger()},
			Text:      "Whenever this creature deals combat damage to a player, look at that many cards from the top of your library. Put one of them into your hand and the rest into your graveyard.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "Find New Host — {2}{U/B}, Exile this card from your graveyard: Put a +1/+1 counter on target creature you control. It gains this card's other abilities. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{U/B}"), effects.ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:       symbioteSpiderManFindNewHost,
		}},
	})
}

// symbioteSpiderManGrant is the bundle key of "this card's other
// abilities". An on-disk identity: never renamed, never reused.
const symbioteSpiderManGrant = "symbiote-spider-man/other-abilities"

func symbioteSpiderManTrigger() game.TriggeredAbility {
	return effects.WheneverThisDealsCombatDamageToAPlayer(
		"Symbiote Spider-Man — look at that many cards; one into your hand, the rest into your graveyard",
		func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
				return nil
			}
			ctx := effects.NewContext(g, item)
			player := item.Controller
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  g.LookAtTopOfLibraryForEffect(player, item.Trigger.Event.Amount),
				Max:    1,
				Label:  "Symbiote Spider-Man — put one into your hand",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		})
}

func symbioteSpiderManFindNewHost(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: target, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantAbilitiesFor{
		Target:   target,
		Keys:     []string{symbioteSpiderManGrant},
		Duration: g.PinnedTo(game.IndefiniteDuration(), target),
		Label:    "Symbiote Spider-Man — Find New Host: it gains the combat-damage look",
	}.Apply(ctx)
}
