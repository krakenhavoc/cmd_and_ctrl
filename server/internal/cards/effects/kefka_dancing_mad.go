package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kefka, Dancing Mad — Legendary Creature — Human Wizard {5}{B}{R},
// 6/6:
//
//	"During your turn, Kefka has indestructible.
//	 At the beginning of your end step, exile a card at random from each
//	 opponent's graveyard. You may cast any number of spells from among
//	 cards exiled this way without paying their mana costs. Then each
//	 player who owns a spell you cast this way loses life equal to its
//	 mana value."
//
// Indestructible is Zurgo Helmsmasher's turn-gated layer 6 grant,
// re-answered on every recompute.
//
// The end-step ability exiles one random card from each opponent's
// graveyard (an opponent with an empty graveyard contributes nothing)
// and grants Etali's free-cast permission over each card: cast only,
// {0} so X is locked at 0 (CR 107.3b), at instant timing and lapsing
// on the caster's next priority pass, which is the engine's spelling of
// "as part of resolving this ability" (CR 608.2g, Conduit of Worlds).
// An uncast card stays in exile and a land is stranded, since the text
// names spells.
//
// The life loss is the permission's FollowUp, which runs once the cast
// is made through THIS grant. It is read when the spell goes on the
// stack rather than after every cast has finished; because every
// spell cast this way waits on the stack until the ability has
// finished resolving, the table sees the same life totals the printed
// order would produce before any of those spells resolve. The loser is
// the spell card's owner (an opponent, since the cards came from their
// graveyards) and the amount is its mana value.
//
// No simplification.
var kefkaLifeLoss = game.RegisterCastFollowUp("kefka-dancing-mad-life-loss",
	func(g *game.Game, f game.CastFollowUp) error {
		c, ok := g.LookupCardForEffect(f.Spell)
		if !ok {
			return nil
		}
		mv := c.ManaValue()
		if mv <= 0 || g.PlayerByIDForEffect(c.Owner) == nil {
			return nil
		}
		return g.ChangePlayerLifeForEffect(f.Source, c.Owner, -mv)
	})

func init() {
	Register(Spec{
		OracleID:     "fe5690d3-547b-4ce9-8e94-77fdc0e9c5c6",
		Name:         "Kefka, Dancing Mad",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "During your turn, has indestructible.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = append(c.Abilities, "indestructible")
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtYourEndStep("Kefka — exile a card at random from each opponent's graveyard; cast them free", kefkaEndStep),
		},
	})
}

func kefkaEndStep(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, opp := range ctx.Opponents() {
		pick := g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, opp, nil), 1)
		if len(pick) == 0 {
			continue
		}
		if err := g.ExileCardWithPermissionForEffect(pick[0], game.CastPermission{
			Player:      item.Controller,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			Timing:      game.TimingFlash,
			CastOnly:    true,
			LapseOnPass: game.LapseStaysInExile,
			FollowUp:    kefkaLifeLoss,
			Source:      item.SourceCardID,
			SourceName:  "Kefka, Dancing Mad",
			Label:       "Kefka, Dancing Mad — cast it without paying its mana cost",
		}); err != nil {
			return err
		}
	}
	return nil
}
