package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b742_helpers.go — Exodus, S82 OLD batch 742 (#5626).

// WithMoreCardsInHandThanTheChooser narrows a player clause the triggering
// player chooses (OfThatPlayersChoice) to "who has more cards in hand than
// they do and is their opponent", they being that chooser (Oath of Scholars,
// game.TargetSpec.MoreCardsInHandThanChooser).
func WithMoreCardsInHandThanTheChooser(spec *game.TargetSpec) *game.TargetSpec {
	spec.MoreCardsInHandThanChooser = true
	return spec
}

// S82OldB742RevealUntilLandRestToGraveyard is Avenging Druid's "reveal cards
// from the top of your library until you reveal a land card. If you do, put
// that card onto the battlefield and put all other cards revealed this way
// into your graveyard." A library with no land card is revealed whole and all
// of it goes to the graveyard.
func S82OldB742RevealUntilLandRestToGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	run, hit := RevealUntil(ctx, player, game.Card.IsLand, "Avenging Druid — revealed until a land card")
	return PutFromLibraryOntoBattlefield{
		Player:   player,
		Cards:    run,
		Match:    func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return hit != uuid.Nil && c.InstanceID == hit },
		All:      true,
		ThenTail: PutRestIntoGraveyardTail,
	}.Apply(ctx)
}

// S82OldB742TargetOpponentWithAtLeastTwoMoreCardsInHand is Keeper of the
// Mind's "target opponent who has at least two more cards in hand than you do
// as you activate this ability". The hand sizes are judged as the ability is
// activated; the draw does not depend on the opponent, so the ability still
// resolves if the target stops qualifying (StillResolves).
func S82OldB742TargetOpponentWithAtLeastTwoMoreCardsInHand() *game.TargetSpec {
	spec := TargetPlayer("target opponent who has at least two more cards in hand than you do",
		Opponent(),
		func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			me := g.PlayerByIDForEffect(caster)
			return me != nil && me.Hand != nil && p.Hand != nil && p.Hand.Size() >= me.Hand.Size()+2
		})
	return spec.StillResolves()
}

// cataclysmTypes are the four types a player keeps one permanent of.
var cataclysmTypes = []string{"artifact", "creature", "enchantment", "land"}

func cataclysmTypesOf(c game.Card) []string {
	var out []string
	for _, t := range cataclysmTypes {
		if c.HasCardType(t) {
			out = append(out, t)
		}
	}
	return out
}

func cataclysmMask(types []string) int {
	mask := 0
	for i, name := range cataclysmTypes {
		if slices.Contains(types, name) {
			mask |= 1 << i
		}
	}
	return mask
}

// cataclysmLegs are the per-seat set rules, one for each non-empty set of the
// four types a board can have. The keys name the types and are on-disk
// identities: never renamed, never reused.
var cataclysmLegs = func() (legs [1 << 4]game.ChooseCardsThen) {
	for mask := 1; mask < 1<<len(cataclysmTypes); mask++ {
		var present []string
		for i, name := range cataclysmTypes {
			if mask&(1<<i) != 0 {
				present = append(present, name)
			}
		}
		legs[mask] = game.RegisterOrderedPickLegRule(
			"choose-cards/cataclysm-keep-"+strings.Join(present, "-"),
			LilianaKeepsOneOfEachType(present))
	}
	return legs
}()

// cataclysmBoard is every permanent a seat controls that the effect may
// sacrifice, read through the layered view.
func cataclysmBoard(g *game.Game, seat uuid.UUID) []game.Card {
	var out []game.Card
	for _, id := range g.EffectSacrificeCandidatesForEffect(seat, nil) {
		if c, ok := g.LookupCardForEffect(id); ok {
			out = append(out, c)
		}
	}
	return out
}

// cataclysmChoosable is the part of a board that has one of the four types.
func cataclysmChoosable(board []game.Card) []game.Card {
	var out []game.Card
	for _, c := range board {
		if len(cataclysmTypesOf(c)) > 0 {
			out = append(out, c)
		}
	}
	return out
}

func cataclysmPresent(cards []game.Card) []string {
	var present []string
	for _, t := range cataclysmTypes {
		for _, c := range cards {
			if c.HasCardType(t) {
				present = append(present, t)
				break
			}
		}
	}
	return present
}

// cataclysmForced reports that a board needs no choice: everything on it is
// choosable and the whole of it is the one keep-set (nothing is sacrificed).
func cataclysmForced(board []game.Card) bool {
	choosable := cataclysmChoosable(board)
	if len(choosable) != len(board) {
		return false
	}
	return LilianaKeepsOneOfEachType(cataclysmPresent(choosable))(choosable)
}

// S82OldB742Cataclysm is "Each player chooses from among the permanents they
// control an artifact, a creature, an enchantment, and a land, then sacrifices
// the rest." Each player is asked once, in turn order, for the permanents they
// keep (each kept permanent stands for a type of its own, so an artifact
// creature may be both the artifact and the creature, but then nothing else
// need be kept for either); when everyone has answered, everything else they
// control goes at once (CR 701.21a). A permanent with none of the four types
// (a planeswalker, a battle) can't be chosen and is sacrificed. A player who
// would keep everything anyway is not asked.
func S82OldB742Cataclysm(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	var seats []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		board := cataclysmBoard(g, seat)
		if len(board) == 0 {
			continue
		}
		seats = append(seats, seat)
		choosable := cataclysmChoosable(board)
		if cataclysmForced(board) || len(choosable) == 0 {
			continue
		}
		present := cataclysmPresent(choosable)
		ids := make([]uuid.UUID, len(choosable))
		for i, c := range choosable {
			ids[i] = c.InstanceID
		}
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Cataclysm — choose an artifact, a creature, an enchantment, and a land from among the permanents you control to keep; the rest are sacrificed",
			Cards:      ids,
			Min:        1,
			Max:        len(present),
			Zone:       game.ZoneBattlefield,
			ThenKey:    cataclysmLegs[cataclysmMask(present)],
		})
	}
	if len(seats) == 0 {
		return nil
	}
	if len(legs) == 0 {
		return cataclysmSacrificeRest(g, item.SourceCardID, seats, nil)
	}
	return g.ChoosePermanentsInOrderKeyedForEffect(legs, cataclysmKeepRest, item.SourceCardID, game.RunData{Carry: seats})
}

// cataclysmKeepRest is "then sacrifices the rest", keyed. Carry is every seat
// with a permanent, in turn order.
var cataclysmKeepRest = game.RegisterPromptRunThen("prompt-run/s82-old-b742-cataclysm-keep", func(g *game.Game, r game.RunDone) error {
	return cataclysmSacrificeRest(g, r.Source, r.Carry, r.Picks().Cards())
})

// cataclysmSacrificeRest sacrifices, at once, every permanent the seats
// control that is neither among `picked` nor on a board that needed no choice.
func cataclysmSacrificeRest(g *game.Game, source uuid.UUID, seats, picked []uuid.UUID) error {
	kept := map[uuid.UUID]bool{}
	for _, id := range picked {
		kept[id] = true
	}
	var rest []uuid.UUID
	for _, seat := range seats {
		board := cataclysmBoard(g, seat)
		forced := cataclysmForced(board)
		for _, c := range board {
			if kept[c.InstanceID] || (forced && len(cataclysmTypesOf(c)) > 0) {
				continue
			}
			rest = append(rest, c.InstanceID)
		}
	}
	if len(rest) == 0 {
		return nil
	}
	return g.SacrificeAllKeyedForEffect(source, rest, game.KeyedTail{})
}

// S82OldB742FadeAway is "For each creature, its controller sacrifices a
// permanent of their choice unless they pay {1}." The creatures are counted as
// the spell resolves, one decision for each, in turn order of their
// controllers (CR 101.4): the controller pays {1} or sacrifices a permanent of
// their choice, then the next creature's.
func S82OldB742FadeAway(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var owners []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.Controller == seat && c.IsCreature() {
				owners = append(owners, seat)
			}
		}
	}
	for _, seat := range owners {
		if err := (PayUnless{
			Chooser:  seat,
			Cost:     "{1}",
			Question: "Fade Away — pay {1} or sacrifice a permanent?",
			Else:     s82OldB742FadeAwayDecline,
		}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

var s82OldB742FadeAwayDecline = PayUnlessElse("pay-unless/s82-old-b742-fade-away-sacrifice", func(ctx *Context, r game.PayUnlessAnswer) error {
	B17PlayerSacrificesN(ctx.Game, ctx.Source(), r.Chooser, 1, "Fade Away — sacrifice a permanent")
	return nil
})
