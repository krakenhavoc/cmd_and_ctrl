package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niv-Mizzet, Parun — Legendary Creature — Dragon Wizard
// {U}{U}{U}{R}{R}{R}, 5/5:
//
//	"This spell can't be countered.
//	 Flying
//	 Whenever you draw a card, Niv-Mizzet deals 1 damage to any
//	 target.
//	 Whenever a player casts an instant or sorcery spell, you draw a
//	 card."
//
// The Firemind's draw ping with the tap ability traded for the second
// half of the loop: every instant or sorcery anyone casts draws you a
// card, and every card drawn is a ping. The cast trigger goes on the
// stack above the spell, so the draw — and its ping — resolve first.
// Each drawn card is its own trigger with its own "any target" pick.
// "Can't be countered" is the S23 rider, honoured at the counter
// choke point.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "33666a98-812f-4892-9f8d-33e0cbecc340",
		Name:            "Niv-Mizzet, Parun",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventDrawCard},
				AppliesTo: ByYou,
				Targets:   TargetAny(),
				Key:       "Niv-Mizzet, Parun — deal 1 damage to any target",
				Effect:    sourceDealsDamageToEachLegalTarget(1),
			},
			On(game.EventCast, anyPlayerCastAnInstantOrSorcery, "Niv-Mizzet, Parun — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

// anyPlayerCastAnInstantOrSorcery is "whenever a player casts an
// instant or sorcery spell" — the spell is read off the stack, and
// the caster is anyone.
func anyPlayerCastAnInstantOrSorcery(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}
