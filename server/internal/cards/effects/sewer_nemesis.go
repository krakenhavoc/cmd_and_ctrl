package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sewer Nemesis — Creature — Horror {3}{B}, */* (#4092):
//
//	"As this creature enters, choose a player.
//	 Sewer Nemesis's power and toughness are each equal to the number
//	 of cards in the chosen player's graveyard.
//	 Whenever the chosen player casts a spell, that player mills a
//	 card."
//
// Entropic Specter's shape over a graveyard instead of a hand. The
// player is an as-enters choice stored on the permanent
// (ChoosePlayerAsEnters, Players: you may choose yourself). The size is
// a layer 7a characteristic-defining ability that reads the chosen
// seat's graveyard live; a graveyard only changes by a zone move, and
// every zone move already invalidates the layer cache, so it needs no
// dependency hint (Splinterfright's count does the same). Until the
// choice is made the ability defines nothing and the Nemesis keeps the
// printed `*`, which the state-based actions skip (#690), so it is not
// put into the graveyard as a 0/0 before its controller has answered.
//
// The mill is a cast trigger on the chosen player, whoever controls the
// Nemesis, and "that player" is the caster read off the event. A copy
// of a spell is not cast and does not trigger it (CR 707.10).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c7f1a49a-cc22-432f-bf3e-50ba61982b87",
		Name:         "Sewer Nemesis",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlayerAsEnters("Sewer Nemesis", Players),
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to the number of cards in the chosen player's graveyard.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && ChosenPlayerOf(g, source.InstanceID) != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := sewerNemesisGraveyardSize(g, ChosenPlayerOf(g, source.InstanceID))
				c.Power = n
				c.Toughness = n
			},
		}},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, sewerNemesisChosenPlayerCasts,
				"Sewer Nemesis — the chosen player mills a card", sewerNemesisMill),
		},
	})
}

func sewerNemesisGraveyardSize(g *game.Game, player uuid.UUID) int {
	if p := g.PlayerByIDForEffect(player); p != nil && p.Graveyard != nil {
		return p.Graveyard.Size()
	}
	return 0
}

func sewerNemesisChosenPlayerCasts(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	chosen := ChosenPlayerOf(g, source.InstanceID)
	return chosen != uuid.Nil && ev.Actor == chosen
}

func sewerNemesisMill(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	caster := ctx.Trigger().Event.Actor
	if caster == uuid.Nil {
		return nil
	}
	return MillCards{Player: caster, N: 1}.Apply(ctx)
}
