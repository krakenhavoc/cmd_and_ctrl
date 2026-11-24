package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaya the Inexorable — Legendary Planeswalker — Kaya {3}{W}{B}, loyalty 5:
//
//	"+1: Put a ghostform counter on up to one target nontoken creature. It
//	 gains 'When this creature dies or is put into exile, return it to its
//	 owner's hand and create a 1/1 white Spirit creature token with flying.'
//	 −3: Exile target nonland permanent.
//	 −7: You get an emblem with 'At the beginning of your upkeep, you may cast a
//	 legendary spell from your hand, from your graveyard, or from among cards
//	 you own in exile without paying its mana cost.'"
//
// The +1 puts the counter and gives the creature the trigger for as long as it
// stays that object on the battlefield (no stated duration, CR 611.2a; the
// counter is a marker and carries no rule). The granted trigger is a
// leaves-the-battlefield look-back (CR 603.10a) over a move to a graveyard or
// to exile; the card returns to its owner's hand only if it is still the object
// that left, in a graveyard or in exile, and the Spirit is made whether or not it
// does. The emblem is Basri Ket's shape with a free cast of one legendary
// spell from the three places, picked first so only one permission exists.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "13a4f88b-6832-43a1-b4fa-c1ee60560bad",
		Name:            "Kaya the Inexorable",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Grants: []effects.AbilityGrant{{
			Key: effects.OldB071KayaGhostformBundle,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B22SelfDiedOrWasExiledFromBattlefield(ev, source)
				}, "Kaya the Inexorable — return it to its owner's hand and create a 1/1 white Spirit creature token with flying",
					effects.OldB071GhostformReturn),
			},
			Text: "When this creature dies or is put into exile, return it to its owner's hand and create a 1/1 white Spirit creature token with flying.",
		}},
		Emblem: &effects.EmblemSpec{
			Label: "Kaya the Inexorable emblem",
			Text:  "At the beginning of your upkeep, you may cast a legendary spell from your hand, from your graveyard, or from among cards you own in exile without paying its mana cost.",
			Triggered: []game.TriggeredAbility{
				effects.AtYourUpkeep("Kaya the Inexorable emblem — you may cast a legendary spell from your hand, your graveyard, or among cards you own in exile without paying its mana cost",
					effects.OldB071KayaEmblemUpkeep),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Put a ghostform counter on up to one target nontoken creature. It gains \"When this creature dies or is put into exile, return it to its owner's hand and create a 1/1 white Spirit creature token with flying.\"",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.UpToOneTargetCreature("up to one target nontoken creature", effects.Not(effects.IsTokenPredicate())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.AddCounter{Target: t.ID, Kind: "ghostform", N: 1}).Apply(ctx); err != nil {
							return err
						}
						return effects.GrantAbilitiesFor{
							Target:   t.ID,
							Keys:     []string{effects.OldB071KayaGhostformBundle},
							Duration: game.IndefiniteDuration(),
							Label:    "Kaya the Inexorable — it returns to hand when it dies or is exiled",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−3: Exile target nonland permanent.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				Effect:  effects.SocB37ExileTheLegalTarget,
			},
			{
				Label:  "−7: You get an emblem with \"At the beginning of your upkeep, you may cast a legendary spell from your hand, from your graveyard, or from among cards you own in exile without paying its mana cost.\"",
				Cost:   effects.LoyaltyCost(-7),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}
