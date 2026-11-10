package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Two Streams Facility — Plane — Apalapucia (Doctor Who, #4870):
//
//	"Whenever you planeswalk to Two Streams Facility and at the beginning
//	 of the first upkeep of the game, each player chooses green anchor or
//	 red waterfall.
//	 Each player who last chose green anchor may play an additional land
//	 during each of their turns.
//	 Creatures controlled by players who last chose red waterfall get +2/+0
//	 and have haste.
//	 Whenever chaos ensues, each player who last chose green anchor
//	 chooses red waterfall, and vice versa."
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6). The
// plane remembers each player's last choice on itself, one counter of kind
// "<choice>:<player ID>" per player (game.SetFaceUpPlanarCounterForEffect):
// the memory is the plane object's, so a planeswalk away ends it (CR 901.7a)
// and the planeswalk back asks everyone again. The first ability is one ability
// with two trigger conditions (the planeswalk to it, and the first upkeep of
// the game, when it is the starting plane); every player is asked in turn order
// from the planar controller. The extra land play is Spec.PlanarLandPlays,
// read by the land-drop count (game/land_drops.go), and the creature bonus is
// two statics reading the plane's memory, so a chaos roll's flip takes effect
// at once. Chaos flips every player's choice (a player who never chose has none
// to flip).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8aa7f2d-2280-42d4-b082-1b5d45df77c5",
		Name:         "Two Streams Facility",
		Completeness: CompletenessFull,
		PlanarLandPlays: &PlanarLandPlays{
			Choice: twoStreamsGreenAnchor + ":",
			Plays:  1,
			Label:  "Each player who last chose green anchor may play an additional land during each of their turns.",
		},
		Static: []game.StaticAbility{
			{
				Label:     "Creatures controlled by players who last chose red waterfall get +2/+0.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7C_Modify,
				AppliesTo: twoStreamsRedWaterfallCreature,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
				},
			},
			b16GrantKeywords(twoStreamsRedWaterfallCreature, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventBeginUpkeep}, twoStreamsAsksForChoices,
				"Two Streams Facility — each player chooses green anchor or red waterfall", twoStreamsAsk),
			WheneverChaosEnsues("Two Streams Facility — each player who last chose green anchor chooses red waterfall, and vice versa", twoStreamsFlip),
		},
	})
}

const (
	twoStreamsGreenAnchor  = "green anchor"
	twoStreamsRedWaterfall = "red waterfall"
)

func twoStreamsKey(choice string, player uuid.UUID) string { return choice + ":" + player.String() }

// twoStreamsRedWaterfallCreature is a creature whose controller last chose
// red waterfall, as the plane (the static's source) remembers it.
func twoStreamsRedWaterfallCreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && source != nil && source.Counters[twoStreamsKey(twoStreamsRedWaterfall, target.Controller)] > 0
}

// twoStreamsAsksForChoices is the first ability's condition: the planar
// controller planeswalked to this plane, or it is the first upkeep of the
// game.
func twoStreamsAsksForChoices(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventPlaneswalkTo:
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller
	case game.EventBeginUpkeep:
		return g.Turn.Seq == 1 && g.Turn.StepOrdinal <= 1
	}
	return false
}

// twoStreamsAsk asks every player, in turn order from the planar controller.
func twoStreamsAsk(g *game.Game, item *game.StackItem) error {
	return twoStreamsAskNext(NewContext(g, item), apnapPlayersFrom(g, item.Controller))
}

func twoStreamsAskNext(ctx *Context, players []uuid.UUID) error {
	if len(players) == 0 {
		return nil
	}
	return PickOption{
		Player:   players[0],
		Question: "Two Streams Facility — choose green anchor or red waterfall",
		Options: []game.ChoiceOption{
			{Label: "Green anchor"},
			{Label: "Red waterfall"},
		},
		ThenKey:  twoStreamsChose,
		Carry:    players,
		WithItem: true,
	}.Apply(ctx)
}

// twoStreamsChose records one player's choice on the plane and asks the next.
// An on-disk key: never renamed, never reused.
var twoStreamsChose game.OptionPickThen

func init() {
	twoStreamsChose = OptionPickThen("option-pick/two-streams-facility-choice", func(ctx *Context, r game.OptionPicked) error {
		if len(r.Carry) == 0 {
			return nil
		}
		player, rest := r.Carry[0], r.Carry[1:]
		choice := twoStreamsGreenAnchor
		if r.Index == 1 {
			choice = twoStreamsRedWaterfall
		}
		if r.Option != nil {
			twoStreamsSet(ctx.Game, ctx.Source(), player, choice)
		}
		return twoStreamsAskNext(ctx, rest)
	})
}

// twoStreamsSet makes `choice` the player's last choice on the plane.
func twoStreamsSet(g *game.Game, plane, player uuid.UUID, choice string) {
	for _, other := range []string{twoStreamsGreenAnchor, twoStreamsRedWaterfall} {
		g.SetFaceUpPlanarCounterForEffect(plane, twoStreamsKey(other, player), 0)
	}
	g.SetFaceUpPlanarCounterForEffect(plane, twoStreamsKey(choice, player), 1)
}

// twoStreamsFlip is the chaos ability: each player's last choice becomes the
// other one, read as the ability resolves.
func twoStreamsFlip(g *game.Game, item *game.StackItem) error {
	var plane *game.Card
	for i := range g.PlanarFaceUp.Cards {
		if g.PlanarFaceUp.Cards[i].InstanceID == item.SourceCardID {
			plane = &g.PlanarFaceUp.Cards[i]
		}
	}
	if plane == nil {
		return nil
	}
	type flip struct {
		player uuid.UUID
		to     string
	}
	var flips []flip
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		switch {
		case plane.Counters[twoStreamsKey(twoStreamsGreenAnchor, p.ID)] > 0:
			flips = append(flips, flip{p.ID, twoStreamsRedWaterfall})
		case plane.Counters[twoStreamsKey(twoStreamsRedWaterfall, p.ID)] > 0:
			flips = append(flips, flip{p.ID, twoStreamsGreenAnchor})
		}
	}
	for _, f := range flips {
		twoStreamsSet(g, item.SourceCardID, f.player, f.to)
	}
	return nil
}
