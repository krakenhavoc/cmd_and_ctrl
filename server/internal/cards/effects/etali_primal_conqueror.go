package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Etali, Primal Conqueror — Legendary Creature — Elder Dinosaur
// {5}{R}{R}, 7/7, the front face of a transforming card whose back is
// Etali, Primal Sickness (EDHREC rank 812):
//
//	"Trample
//	 When Etali enters, each player exiles cards from the top of their
//	 library until they exile a nonland card. You may cast any number
//	 of spells from among the nonland cards exiled this way without
//	 paying their mana costs.
//	 {9}{G/P}: Transform Etali. Activate only as a sorcery."
//
// The enters trigger runs Nicol Bolas, God-Pharaoh's exile-until-
// nonland (MillToZone with an Until, so a commander offers its owner the
// command zone and a replacement is honoured) for each player in APNAP
// order, one library after another, then grants Etali's controller a
// free cast of every nonland card that landed (grantFreeCasts): cast
// only, at flash speed because the casts are part of this resolution
// (CR 608.2g), closing on the controller's next pass and leaving an
// uncast card in exile. The lands exiled on the way stay exiled.
//
// The transform is TransformThis at sorcery speed; {G/P} is Phyrexian
// mana, payable with {G} or 2 life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        etaliOracleID,
		Name:            "Etali, Primal Conqueror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Etali, Primal Conqueror — each player exiles until a nonland card; you may cast them free",
				func(g *game.Game, item *game.StackItem) error {
					return etaliExileFrom(NewContext(g, item), apnapLivingSeats(g), nil)
				}),
		},
		Activated: []ActivatedAbility{{
			Label:        "{9}{G/P}: Transform Etali. Activate only as a sorcery.",
			Cost:         ManaCost("{9}{G/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
}

// etaliExileFrom exiles from the first player's library until a nonland
// card, then moves on to the next, and grants the casts at the end.
// `hits` is the nonland cards exiled so far.
func etaliExileFrom(ctx *Context, players, hits []uuid.UUID) error {
	if len(players) == 0 {
		grantFreeCasts(ctx.Game, ctx.Controller(), ctx.Source(), "Etali, Primal Conqueror",
			game.TimingFlash, game.LapseStaysInExile, hits)
		return nil
	}
	rest := players[1:]
	return MillToZone{
		Player: players[0],
		To:     game.ZoneExile,
		Until:  UntilCard(func(c game.Card) bool { return !c.IsLand() }),
		Then: func(ctx *Context, exiled []uuid.UUID) error {
			next := append([]uuid.UUID(nil), hits...)
			for _, id := range exiled {
				if c, ok := ctx.Game.LookupCardForEffect(id); ok && !c.IsLand() {
					next = append(next, id)
				}
			}
			return etaliExileFrom(ctx, rest, next)
		},
	}.Apply(ctx)
}

// Etali, Primal Sickness — Legendary Creature — Phyrexian Elder
// Dinosaur, 11/11, the back face:
//
//	"Trample, indestructible
//	 Whenever Etali deals combat damage to a player, they get that many
//	 poison counters. (A player with ten or more poison counters loses
//	 the game.)"
//
// The poison is the damage event's amount, as it was dealt, given to
// the player it was dealt to, with Etali's controller as the placer
// (so a counter doubler or halver reads the right player).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        etaliOracleID + "#1",
		Name:            "Etali, Primal Sickness",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "indestructible"},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Etali, Primal Sickness — they get that many poison counters",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
						return nil
					}
					ev := item.Trigger.Event
					if p := g.PlayerByIDForEffect(ev.Target); p == nil || p.Eliminated {
						return nil
					}
					return g.AddPlayerCounterByForEffect(item.Controller, ev.Target, game.CounterPoison, ev.Amount)
				}),
		},
	})
}

const etaliOracleID = "7514e401-7aa1-405d-9f7a-312b4e630cc2"
