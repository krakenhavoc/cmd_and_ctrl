package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s34_helpers.go — shared pieces of batch 34, "Goad and attack
// politics" (#5368): the Lightning Rager and Pirate token templates and the
// "creatures named X can't attack you" static.

// printedCmdr2S34LightningRagerToken is Rite of the Raging Storm's 5/1 red
// Elemental named Lightning Rager, with trample, haste and "At the
// beginning of the end step, sacrifice this creature." Its slug is an
// on-disk identity: never renamed, never reused.
func printedCmdr2S34LightningRagerToken() tokenTemplate {
	return tokenTemplate{
		Slug: "lightning-rager",
		Card: game.Card{
			Name:      "Lightning Rager",
			TypeLine:  "Token Creature — Elemental",
			Power:     5,
			Toughness: 1,
			Colors:    []string{"R"},
			Keywords:  []string{"trample", "haste"},
		},
		Triggered: []game.TriggeredAbility{
			AtEachStep(game.StepEnd, "Lightning Rager — sacrifice this creature", func(g *game.Game, item *game.StackItem) error {
				return SacrificePermanent{Target: item.SourceCardID}.Apply(NewContext(g, item))
			}),
		},
		Text: "At the beginning of the end step, sacrifice this creature.",
	}
}

// cmdr2S34LightningRagerToken is the token's card value for CreateToken.
func cmdr2S34LightningRagerToken() game.Card {
	return tokenFromCatalog(printedCmdr2S34LightningRagerToken)
}

// printedCmdr2S34PirateToken is Nettling Nuisance's 4/2 red Pirate with
// "This token can't block." Its slug is an on-disk identity.
func printedCmdr2S34PirateToken() tokenTemplate {
	return tokenTemplate{
		Slug: "pirate-cant-block",
		Card: game.Card{
			Name:      "Pirate",
			TypeLine:  "Token Creature — Pirate",
			Power:     4,
			Toughness: 2,
			Colors:    []string{"R"},
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
		Text:   "This token can't block.",
	}
}

// cmdr2S34PirateToken is the Pirate's card value for CreateToken.
func cmdr2S34PirateToken() game.Card { return tokenFromCatalog(printedCmdr2S34PirateToken) }

// cmdr2S34NamedCreaturesCantAttackYouOrYourPlaneswalkers is "Creatures
// named <name> can't attack you or planeswalkers you control" (CR 508.1c):
// the restriction is written onto every creature with the printing
// permanent as its source and binds only an attacker with that effective
// name, judged at declaration. "You" is the printing permanent's
// controller, read when the layer pass runs.
func cmdr2S34NamedCreaturesCantAttackYouOrYourPlaneswalkers(label, name string) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
			return target.IsCreature()
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, source *game.Card) {
			if source == nil {
				return
			}
			c.AttackTargetRestrictions = append(c.AttackTargetRestrictions, game.AttackTargetRestriction{
				Source:                source.InstanceID,
				SourceName:            source.Name,
				NotPlayer:             source.Controller,
				OnlyAttackersMatching: []game.PermanentQuery{{Names: []string{name}}},
			})
		},
	}
}

// cmdr2S34GoadTokensForRestOfGame is the keyed "then" of "<token> is goaded
// for the rest of the game" (Nettling Nuisance, Rendmaw, Creaking Nest): it
// goads each token the creation made on behalf of the tail's Player, the
// effect's controller (CR 701.15a). A token that is already gone is skipped.
// The key is an on-disk identity: never renamed, never reused.
var cmdr2S34GoadTokensForRestOfGame = CardTail("cmdr2-s34/goad-tokens-for-rest-of-game", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		ctx.Game.GoadForRestOfGameForEffect(id, o.Tail.Player)
	}
	return nil
})

// cmdr2S34EachPlayerCreatesGoadedBird is Rendmaw, Creaking Nest's "each
// player creates a tapped 2/2 black Bird creature token with flying. The
// tokens are goaded for the rest of the game": the players in turn order
// from the active player each create one, and each token is goaded by the
// effect's controller as it lands. The players go one after another
// through one chain of keyed tails, so a creation paused on a
// replacement-ordering question is still a restore point and the goad
// lands on the token it made.
func cmdr2S34EachPlayerCreatesGoadedBird(ctx *Context) error {
	return cmdr2S34BirdForNext(ctx, seatsFromActive(ctx.Game))
}

func cmdr2S34BirdForNext(ctx *Context, players []uuid.UUID) error {
	for len(players) > 0 && ctx.Game.PlayerByIDForEffect(players[0]) == nil {
		players = players[1:]
	}
	if len(players) == 0 {
		return nil
	}
	player, rest := players[0], append([]uuid.UUID(nil), players[1:]...)
	tail := ctx.ItemTail(cmdr2S34GoadThenNextBird)
	tail.Carry = rest
	return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: player,
		Source:     ctx.Source(),
		Groups: []game.TokenGroup{{
			Template: ctx.TextToken(TokenCard("2/2 black Bird with flying")),
			Count:    1,
			Entry:    game.TokenEntryOptions{Tapped: true},
		}},
	}, tail)
}

// cmdr2S34GoadThenNextBird goads the Bird the last player's creation made,
// then makes the next player's. Assigned in init because the body asks the
// next creation, which names this key again. The key is an on-disk
// identity: never renamed, never reused.
var cmdr2S34GoadThenNextBird game.TailRef

func init() {
	cmdr2S34GoadThenNextBird = CardTail("cmdr2-s34/goad-then-next-player-bird", func(ctx *Context, o game.TailOutcome) error {
		for _, id := range o.IDs {
			ctx.Game.GoadForRestOfGameForEffect(id, o.Tail.Player)
		}
		return cmdr2S34BirdForNext(ctx, o.Tail.Carry)
	})
}

// The "each player may put two +1/+1 counters on a creature they control"
// fan-out of Agitator Ant and Orzhov Advokist. Every player in turn order
// from the active player is asked in turn and may pick one creature they
// control or none (a choice, not a target); once the last has answered the
// counters go on every chosen creature together and the card's own rider
// runs. One keyed pick serves both cards, told which rider by its first
// number.
const (
	cmdr2S34RiderGoad     = 0 // Agitator Ant: goad each creature that got counters
	cmdr2S34RiderCantAtk  = 1 // Orzhov Advokist: that player's creatures can't attack you
	cmdr2S34CounterAmount = 2
)

// cmdr2S34EachPlayerMayPutCounters starts the fan-out for `rider`.
func cmdr2S34EachPlayerMayPutCounters(ctx *Context, rider int) error {
	return cmdr2S34AskCounters(ctx, seatsFromActive(ctx.Game), nil, nil, rider)
}

// cmdr2S34AskCounters asks the first of `players`, then the rest; the carry
// is the players still to ask, then the creatures chosen so far, then who
// chose each.
func cmdr2S34AskCounters(ctx *Context, players, chosen, choosers []uuid.UUID, rider int) error {
	if len(players) == 0 {
		return cmdr2S34ApplyCounters(ctx, chosen, choosers, rider)
	}
	p, rest := players[0], append([]uuid.UUID(nil), players[1:]...)
	return ChoosePermanents{
		Player:   p,
		Of:       []uuid.UUID{p},
		Question: "You may put two +1/+1 counters on a creature you control",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var ids []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() {
					ids = append(ids, c.InstanceID)
				}
			}
			if len(ids) == 0 {
				return nil, 0, 0
			}
			return ids, 0, 1
		},
		Keyed: cmdr2S34CountersPicked,
		Carry: packIDLists(rest, chosen, choosers),
		Nums:  []int{rider},
	}.Apply(ctx)
}

// cmdr2S34CountersPicked is one player's answer, then the next player's.
// Registered in init because the answer asks the next player, which names
// the key again. The key is an on-disk identity: never renamed, never
// reused.
var cmdr2S34CountersPicked PermanentsThen

func init() {
	cmdr2S34CountersPicked = RegisterPermanentsPick("prompt-run/cmdr2-s34-each-player-may-put-counters", PermanentsDef{
		ThenRun: func(ctx *Context, picked game.PromptedPicks, r game.RunDone) error {
			lists := unpackIDLists(r.Carry)
			var rest, chosen, choosers []uuid.UUID
			if len(lists) > 0 {
				rest = lists[0]
			}
			if len(lists) > 1 {
				chosen = lists[1]
			}
			if len(lists) > 2 {
				choosers = lists[2]
			}
			rider := 0
			if len(r.Nums) > 0 {
				rider = r.Nums[0]
			}
			// The asked player is the one whose creature was just picked, or,
			// when they picked nothing, the one this leg asked: it is the
			// leg's own player, which is the pick's controller at the time.
			for _, id := range picked.Cards() {
				chosen = append(chosen, id)
				if c, ok := ctx.Game.LookupCardForEffect(id); ok {
					choosers = append(choosers, c.Controller)
				} else {
					choosers = append(choosers, uuid.Nil)
				}
			}
			return cmdr2S34AskCounters(ctx, rest, chosen, choosers, rider)
		},
	})
}

// cmdr2S34ApplyCounters puts two +1/+1 counters on each chosen creature
// that is still on the battlefield, then runs the rider for those that
// really got counters.
func cmdr2S34ApplyCounters(ctx *Context, chosen, choosers []uuid.UUID, rider int) error {
	g := ctx.Game
	var got, gotBy []uuid.UUID
	for i, id := range chosen {
		if !onBattlefield(g, id) {
			continue
		}
		before := counterCountOn(g, id, game.CounterPlusOne)
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: cmdr2S34CounterAmount}).Apply(ctx); err != nil {
			return err
		}
		if counterCountOn(g, id, game.CounterPlusOne) > before {
			got = append(got, id)
			if i < len(choosers) {
				gotBy = append(gotBy, choosers[i])
			}
		}
	}
	switch rider {
	case cmdr2S34RiderGoad:
		// "Goad each creature that had counters put on it this way."
		if len(got) == 0 {
			return nil
		}
		return goadAndScheduleClear(ctx, got)
	case cmdr2S34RiderCantAtk:
		// "If a player does, creatures that player controls can't attack you
		// or planeswalkers you control until your next turn."
		you := ctx.Controller()
		seen := map[uuid.UUID]bool{}
		for _, player := range gotBy {
			if player == uuid.Nil || seen[player] {
				continue
			}
			seen[player] = true
			g.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, player,
				[]game.Mod{game.CantAttackPlayerMod(you)}, DurationUntilYourNextTurn(ctx, you),
				"Orzhov Advokist — creatures that player controls can't attack you or your planeswalkers until your next turn")
		}
	}
	return nil
}
