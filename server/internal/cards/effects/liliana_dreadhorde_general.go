package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Liliana, Dreadhorde General — Legendary Planeswalker — Liliana
// {4}{B}{B}, loyalty 6:
//
//	"Whenever a creature you control dies, draw a card.
//	 +1: Create a 2/2 black Zombie creature token.
//	 −4: Each player sacrifices two creatures of their choice.
//	 −9: Each opponent chooses a permanent they control of each permanent
//	 type and sacrifices the rest."
//
// −4 gathers each living player's choice in APNAP order, then sacrifices
// all chosen creatures together. A player with fewer than two creatures
// chooses all of them.
//
// −9 is a keep-set: each eligible opponent is asked once, over their
// whole board, for the permanents they keep, and when everyone has
// answered, everything else those opponents control is sacrificed at
// once. The set rule is
// lilianaKeepsOneOfEachType: every permanent type the opponent controls
// has a kept permanent of that type, and every kept permanent stands for
// at least one of those types on its own. A permanent with several
// types may be kept for each of them (an artifact creature can be both
// the artifact and the creature), per the War of the Spark release
// notes. An opponent who would keep everything anyway is not asked.
//
// Battle is a permanent type (CR 110.4), so it counts: permanentTypeNames
// includes it, and the card's "each permanent type" is not limited to the
// five the ruling lists. Supertypes (legendary, basic, snow) never count.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5e958212-6a5b-4288-8d31-f1572619d7dc",
		Name:         "Liliana, Dreadhorde General",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverACreatureYouControlDies("Liliana, Dreadhorde General — draw a card", Do(DrawCards{N: 1})),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Create a 2/2 black Zombie creature token.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Tokens: 1},
				Effect:  Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 1}),
			},
			{
				Label:  "−4: Each player sacrifices two creatures of their choice.",
				Cost:   LoyaltyCost(-4),
				Effect: lilianaEachPlayerSacrificesTwo,
			},
			{
				Label:  "−9: Each opponent chooses a permanent they control of each permanent type and sacrifices the rest.",
				Cost:   LoyaltyCost(-9),
				Effect: lilianaOpponentsKeepOneOfEachType,
			},
		},
	})
}

// apnapLivingSeats lists the players still in the game in APNAP order (CR
// 101.4), starting with the active player.
func apnapLivingSeats(g *game.Game) []uuid.UUID {
	var out []uuid.UUID
	n := len(g.Seats)
	for i := 0; i < n; i++ {
		p := g.Seats[(g.Turn.ActiveSeat+i)%n]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}

// opponentsInTurnOrder lists the living opponents of `controller` in turn
// order starting with the next player after them (the -9 ruling: "the next
// opponent in turn order makes all of their choices, then each other
// opponent in turn order does the same"). This is the controller's own
// seat, not the active player's, so it stays right if the ability is ever
// activated off-turn.
func opponentsInTurnOrder(g *game.Game, controller uuid.UUID) []uuid.UUID {
	n := len(g.Seats)
	start := -1
	for i, p := range g.Seats {
		if p != nil && p.ID == controller {
			start = i
		}
	}
	if start < 0 {
		return nil
	}
	var out []uuid.UUID
	for i := 1; i < n; i++ {
		p := g.Seats[(start+i)%n]
		if p != nil && !p.Eliminated {
			out = append(out, p.ID)
		}
	}
	return out
}

func lilianaEachPlayerSacrificesTwo(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	source := item.SourceCardID
	spec := sacrificeSpec("a creature", Creature())
	for _, seat := range apnapLivingSeats(g) {
		if !g.CanEffectCauseSacrificeForEffect(seat) {
			continue
		}
		candidates := g.EffectSacrificeCandidatesForEffect(seat, spec)
		if len(candidates) == 0 {
			continue
		}
		n := min(2, len(candidates))
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser: seat, FromPlayer: seat, Source: source,
			Question: "Liliana, Dreadhorde General — choose two creatures to sacrifice",
			Cards:    candidates, Min: n, Max: n, Zone: game.ZoneBattlefield,
			Destination: game.ChooseSacrifice,
		})
	}
	return g.ChoosePermanentsInOrderThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
		chosen := picks.Cards()
		if len(chosen) == 0 {
			return nil
		}
		return g.SacrificeAllThenForEffect(source, chosen, nil)
	})
}

func lilianaOpponentsKeepOneOfEachType(g *game.Game, item *game.StackItem) error {
	var legs []game.ChooseCardsPrompt
	asked := map[uuid.UUID]bool{}
	for _, seat := range opponentsInTurnOrder(g, item.Controller) {
		if !g.CanEffectCauseSacrificeForEffect(seat) {
			continue
		}
		board := g.EffectSacrificeCandidatesForEffect(seat, nil)
		if len(board) == 0 {
			continue
		}
		cards := make([]game.Card, 0, len(board))
		for _, id := range board {
			if c, ok := g.LookupCardForEffect(id); ok {
				cards = append(cards, c)
			}
		}
		present := permanentTypesAmong(cards)
		valid := lilianaKeepsOneOfEachType(present)
		if valid(cards) {
			continue
		}
		asked[seat] = true
		legs = append(legs, game.ChooseCardsPrompt{
			Chooser:    seat,
			FromPlayer: seat,
			Source:     item.SourceCardID,
			Question:   "Liliana, Dreadhorde General — choose a permanent you control of each permanent type to keep; the rest are sacrificed",
			Cards:      board,
			Min:        1,
			Max:        len(present),
			Zone:       game.ZoneBattlefield,
			Validate:   valid,
		})
	}
	if len(legs) == 0 {
		return nil
	}
	source := item.SourceCardID
	return g.ChoosePermanentsInOrderThenForEffect(legs, func(g *game.Game, picks game.PromptedPicks) error {
		kept := map[uuid.UUID]bool{}
		for _, id := range picks.Cards() {
			kept[id] = true
		}
		var rest []uuid.UUID
		for _, seat := range opponentsInTurnOrder(g, item.Controller) {
			if !asked[seat] {
				continue
			}
			for _, id := range permanentsControlledByMatching(g, seat, nil) {
				if !kept[id] {
					rest = append(rest, id)
				}
			}
		}
		if len(rest) == 0 {
			return nil
		}
		return g.SacrificeAllThenForEffect(source, rest, nil)
	})
}

// permanentTypeNames are the permanent types (CR 110.4a), lowercase.
var permanentTypeNames = []string{"artifact", "battle", "creature", "enchantment", "land", "planeswalker"}

// permanentTypesOf lists the permanent types card c has.
func permanentTypesOf(c game.Card) []string {
	var out []string
	for _, t := range permanentTypeNames {
		if c.HasCardType(t) {
			out = append(out, t)
		}
	}
	return out
}

// permanentTypesAmong lists the permanent types any of `cards` has.
func permanentTypesAmong(cards []game.Card) []string {
	seen := map[string]bool{}
	for _, c := range cards {
		for _, t := range permanentTypesOf(c) {
			seen[t] = true
		}
	}
	var out []string
	for _, t := range permanentTypeNames {
		if seen[t] {
			out = append(out, t)
		}
	}
	return out
}

// lilianaKeepsOneOfEachType is the −9's set rule over `present`, the
// permanent types the choosing opponent controls: every type in it is
// covered by a kept permanent, and the kept permanents can each be
// assigned a different type they have (each was chosen for something).
func lilianaKeepsOneOfEachType(present []string) func(picked []game.Card) bool {
	return func(picked []game.Card) bool {
		if len(picked) == 0 || len(picked) > len(present) {
			return false
		}
		types := make([][]string, len(picked))
		for i, c := range picked {
			types[i] = permanentTypesOf(c)
		}
		for _, t := range present {
			covered := false
			for _, ts := range types {
				for _, have := range ts {
					if have == t {
						covered = true
					}
				}
			}
			if !covered {
				return false
			}
		}
		used := map[string]bool{}
		var assign func(i int) bool
		assign = func(i int) bool {
			if i == len(types) {
				return true
			}
			for _, t := range types[i] {
				if used[t] {
					continue
				}
				used[t] = true
				if assign(i + 1) {
					return true
				}
				used[t] = false
			}
			return false
		}
		return assign(0)
	}
}
