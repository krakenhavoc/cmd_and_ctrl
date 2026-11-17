package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b14_helpers.go — March of the Machine Commander, S82 batch 14
// (#5477): the pieces Sandsteppe War Riders and Gimbal, Gremlin Prodigy
// share.

// s82MomB14DifferentlyNamedArtifactTokens counts the distinct names among
// the artifact tokens `player` controls, read through the layered view
// (Audience with Trostani's count for artifact tokens). A token with no
// name has none to differ and adds nothing.
func s82MomB14DifferentlyNamedArtifactTokens(g *game.Game, player uuid.UUID) int {
	g.RecomputeLayersIfStaleLocked()
	names := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != player || !c.IsToken() {
			continue
		}
		if !c.IsArtifact() {
			continue
		}
		if name := strings.ToLower(strings.TrimSpace(c.Effective().Name)); name != "" {
			names[name] = true
		}
	}
	return len(names)
}

// s82MomB14BolsterByArtifactTokenNames is Sandsteppe War Riders' body:
// "bolster X, where X is the number of differently named artifact tokens
// you control", with X read as the ability resolves. Zero bolsters
// nothing (CR 701.39a puts no counters, and no creature is chosen). It
// embeds Bolster for the footprint the printed one-sentence line needs.
type s82MomB14BolsterByArtifactTokenNames struct {
	Bolster
}

func (s s82MomB14BolsterByArtifactTokenNames) Apply(ctx *Context) error {
	x := s82MomB14DifferentlyNamedArtifactTokens(ctx.Game, ctx.Controller())
	if x <= 0 {
		return nil
	}
	return Bolster{N: x}.Apply(ctx)
}

// s82MomB14GremlinBeforeKey files the event cursor before Gimbal's token
// is made. An on-disk key.
const s82MomB14GremlinBeforeKey = "gimbal-gremlin-prodigy/before-gremlin"

// s82MomB14CreateGremlin is Gimbal's first step: record where the event
// log is, then create the 0/0 red Gremlin artifact creature token. It
// embeds CreateToken for the footprint.
type s82MomB14CreateGremlin struct {
	CreateToken
}

func (s s82MomB14CreateGremlin) Apply(ctx *Context) error {
	ctx.RecordStep(s82MomB14GremlinBeforeKey, nil, int(b25LastEventSeq(ctx.Game)))
	_, err := ctx.Game.CreateTokensForEffect(ctx.Controller(), ctx.TextToken(TokenCard("0/0 red Gremlin artifact")), 1, game.TokenEntryOptions{})
	return err
}

// s82MomB14PutGremlinCounters is the second step: X +1/+1 counters on the
// Gremlin the first step made, X being the number of differently named
// artifact tokens you control counted now, so the new Gremlin counts
// itself. The counters are PUT on the token, as printed, rather than
// entering with it (Deekah's shape): a creation that waits on a question
// still gets them. With a replaced or prevented creation there is no
// token and nothing is put. It embeds CounterOnThis for the footprint.
type s82MomB14PutGremlinCounters struct {
	CounterOnThis
}

func (s s82MomB14PutGremlinCounters) Apply(ctx *Context) error {
	rec := ctx.StepRecord(s82MomB14GremlinBeforeKey)
	if len(rec.Ns) < 1 {
		return nil
	}
	for _, id := range b27TokensCreatedByAfter(ctx.Game, ctx.Controller(), uint64(rec.Ns[0])) {
		if !onBattlefield(ctx.Game, id) {
			continue
		}
		x := s82MomB14DifferentlyNamedArtifactTokens(ctx.Game, ctx.Controller())
		if x <= 0 {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: x}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

// s82MomB14SharesCreatureType is "a card that shares a creature type with"
// the creature a cmdr2B03TypesWord describes: any card, not only a creature
// card, so a Kindred spell's creature types count. An empty
// word shares with nothing, and a changeling shares with every card that
// has any creature type (CR 702.73a). A card that is itself a changeling
// has every creature type and shares with any creature that has one.
func s82MomB14SharesCreatureType(word string) func(game.Card) bool {
	return func(c game.Card) bool {
		if word == "" {
			return false
		}
		theirs := game.CreatureTypesOf(&c)
		if word == cmdr2B03AllTypesWord {
			return len(theirs) > 0
		}
		for _, t := range strings.Split(word, "|") {
			for _, u := range theirs {
				if t == u {
					return true
				}
			}
		}
		return false
	}
}

// s82MomB14TreasureForTheCaster is Pain Distributor's first ability body:
// "they create a Treasure token", "they" being the player who cast the
// spell that triggered it, not the Distributor's controller. The caster is
// the event's actor. It embeds CreateToken for the footprint: it makes the
// same token with the same events.
type s82MomB14TreasureForTheCaster struct {
	CreateToken
}

func (s s82MomB14TreasureForTheCaster) Apply(ctx *Context) error {
	caster := ctx.Trigger().Event.Actor
	if caster == uuid.Nil {
		return nil
	}
	return CreateToken{Controller: caster, Template: TreasureToken(), N: 1}.Apply(ctx)
}

// s82MomB14DamageTheArtifactsController is Pain Distributor's second
// ability body: "this creature deals 1 damage to that player", the player
// who controlled the artifact as it left the battlefield (CR 603.10a).
// The Distributor deals it as it last existed if it has gone (CR 608.2h).
// It embeds DamageEachOpponent for the footprint (a damage event).
type s82MomB14DamageTheArtifactsController struct {
	DamageEachOpponent
}

func (s s82MomB14DamageTheArtifactsController) Apply(ctx *Context) error {
	ev := ctx.Trigger().Event
	c, ok := ctx.Game.LookupCardForEffect(ev.CardID)
	if !ok {
		return nil
	}
	controller := leftUnderControlOf(ev, c)
	if controller == uuid.Nil {
		return nil
	}
	return DealDamage{Source: ctx.Source(), Target: controller, Amount: 1}.Apply(ctx)
}

// s82MomB14OpponentArtifactPutIntoGraveyard is "an artifact an opponent
// controls is put into a graveyard from the battlefield" (Party Dude's
// level 2, Pain Distributor): an artifact, token or card, that left for a
// graveyard while controlled by someone other than the source's
// controller, read as it last existed there (CR 603.10a).
func s82MomB14OpponentArtifactPutIntoGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if !b23ArtifactPutIntoGraveyardFromBattlefield(ev, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	controller := leftUnderControlOf(ev, c)
	return controller != uuid.Nil && controller != source.Controller
}
