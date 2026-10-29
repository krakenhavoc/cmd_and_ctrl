package effects

import (
	"github.com/google/uuid"

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
// #4250: the casts are made during the ability's resolution (CR 608.2g),
// each its own answer to the resolution_cast prompt over the exiled
// cards, until the player declines or none is left. "Then each player
// who owns a spell you cast this way loses life" is the prompt's
// continuation (kefkaAfterTheCasts), run once it is answered, so the life
// is lost AFTER every cast, as printed. The loser is the spell card's
// owner (an opponent, since the cards came from their graveyards) and
// the amount is its mana value.
//
// kefkaLifeLoss is the per-cast FollowUp the permission carried before
// #4250. It stays registered (a table saved with such a permission open
// still restores), and it is still the life loss outside an open
// resolution, where no prompt is asked and the permissions alone are
// the window.
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
	var picks []uuid.UUID
	for _, opp := range ctx.Opponents() {
		picks = append(picks, g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, opp, nil), 1)...)
	}
	if len(picks) == 0 {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	perm := game.CastPermission{
		Player:      controller,
		Duration:    g.UntilEndOfTurnDuration(),
		Cost:        "{0}",
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInExile,
		Source:      source,
		SourceName:  "Kefka, Dancing Mad",
		Label:       "Kefka, Dancing Mad — cast it without paying its mana cost",
	}
	inline := 0
	if g.ResolutionOpenForEffect() {
		inline = 1
	} else {
		perm.FollowUp = kefkaLifeLoss
	}
	// Keyed (#4253 slice 10a): the permission rides the exile's tail as
	// data, and N is whether the casts are offered inside the resolution.
	return g.ExileCardsKeyedForEffect(picks, game.KeyedTail{Ref: kefkaExiled, Player: controller, Source: source,
		N: inline, Permission: &perm})
}

// kefkaExiled grants the free cast over each card that reached exile and,
// inside the resolution, offers the casts.
var kefkaExiled = game.RegisterTail("exile/kefka-dancing-mad-grant", func(g *game.Game, o game.TailOutcome) error {
	if o.Tail.Permission == nil {
		return nil
	}
	var landed []uuid.UUID
	for _, id := range o.IDs {
		grant := *o.Tail.Permission
		grant.Zone = game.ZoneExile
		if g.GrantCastPermissionOverCardForEffect(id, grant) {
			landed = append(landed, id)
		}
	}
	if o.Tail.N == 1 {
		g.OfferResolutionCastForEffect(game.ResolutionCastOffer{
			Chooser:      o.Tail.Player,
			Source:       o.Tail.Source,
			Reason:       "Kefka, Dancing Mad",
			Cards:        landed,
			Keyword:      game.MayCastKeywordFreeCast,
			AcceptLabel:  "Cast it free",
			DeclineLabel: declineLeaveInExile,
			Then:         kefkaAfterTheCasts,
			Carry:        landed,
		})
	}
	return nil
})

// kefkaAfterTheCasts is "Then each player who owns a spell you cast this
// way loses life equal to its mana value", once the resolution_cast prompt
// is answered: Picked are the exiled cards cast through it, still on the
// stack. An on-disk key, never renamed.
var kefkaAfterTheCasts = game.RegisterChooseCardsThen("resolution-cast/kefka-life-loss",
	func(g *game.Game, r game.CardsChosen) error {
		for _, id := range r.Picked {
			c, ok := g.LookupCardForEffect(id)
			if !ok {
				continue
			}
			mv := c.ManaValue()
			if mv <= 0 || g.PlayerByIDForEffect(c.Owner) == nil {
				continue
			}
			if err := g.ChangePlayerLifeForEffect(r.Source, c.Owner, -mv); err != nil {
				return err
			}
		}
		return nil
	})
