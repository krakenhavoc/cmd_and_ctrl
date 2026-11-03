package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Postmortem Professor — Creature — Zombie Warlock {1}{B}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"This creature can't block.
//	 Whenever this creature attacks, each opponent loses 1 life and you
//	 gain 1 life.
//	 {1}{B}, Exile an instant or sorcery card from your graveyard: Return
//	 this card from your graveyard to the battlefield."
//
// "Can't block" is a restriction the card prints on itself (RestrictSelf).
// The attack trigger is the drain pair (EachOpponentLosesLife, GainLife),
// one per attack declaration of this creature.
//
// The third ability is activated from the graveyard (CR 113.6, Zones): the
// exile is a cost, chosen by the activator at announcement, and an instant
// or sorcery card only. The Professor itself is never a legal pick. It
// returns untapped, under its owner's control; if it has left the
// graveyard by the time the ability resolves, nothing is returned
// (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0859d16e-2a1f-4b48-8687-99af8a8bdb38",
		Name:         "Postmortem Professor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			RestrictSelf(game.CantBlock),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, ThisAttacked,
				"Postmortem Professor — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{B}, Exile an instant or sorcery card from your graveyard: Return this card from your graveyard to the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: Plus(ManaCost("{1}{B}"),
				ExileFromGraveyard(1, "an instant or sorcery card", MatchInstantOrSorcery)),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			Effect: func(g *game.Game, item *game.StackItem) error {
				// CR 602.5 / 608.2a: the card may have left the graveyard
				// while the ability waited; it then does nothing.
				if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return ReturnFromGraveyard{
					Target:     item.SourceCardID,
					Dest:       game.ZoneBattlefield,
					Controller: item.Controller,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}
