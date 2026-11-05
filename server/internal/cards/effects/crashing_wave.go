package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crashing Wave — Sorcery {U}{U} (Avatar: The Last Airbender, #4459):
//
//	"As an additional cost to cast this spell, waterbend {X}. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures
//	 to help. Each one pays for {1}.)
//	 Tap up to X target creatures, then distribute three stun counters
//	 among any number of tapped creatures your opponents control. (If a
//	 permanent with a stun counter would become untapped, remove one
//	 from it instead.)"
//
// The waterbend {X} is Waterbender's Restoration's mandatory
// Spec.TapCost, and "up to X target creatures" is a clause counted by
// the announced X with zero allowed (UpToX).
//
// The counters are not targeted: the ruling says the split is decided
// as the spell resolves, among ANY tapped creatures opponents control,
// not only the ones it just tapped. So, after the taps, the caster
// chooses which of those creatures get counters (one to three of them),
// and the split follows from how many were chosen: one creature takes
// all three, three take one each, and of two the caster names the one
// that takes two. Every distribution of three among any number is one
// answer; each chosen creature gets at least one (CR 601.2d's
// "distribute" reading). With a single candidate there is no choice to
// make and it takes all three. With none, the counters go nowhere.
//
// If every target is illegal as it resolves the spell does nothing at
// all (CR 608.2b), the ruling; a cast with no targets still resolves.
//
// No simplification.
func init() {
	targets := TargetCreature("up to X target creatures")
	targets.CountFromX = true
	targets.UpToX = true
	Register(Spec{
		OracleID:     "c8d3936f-4c08-4ffe-83c9-7ed26fb8f9cd",
		Name:         "Crashing Wave",
		Completeness: CompletenessFull,
		TapCost:      Waterbend("{X}"),
		Targets:      targets,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return crashingWaveDistribute(ctx)
		},
	})
}

// crashingWaveCandidates are the tapped creatures controlled by an
// opponent of `you`, through the layered view.
func crashingWaveCandidates(g *game.Game, you uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you && c.Tapped && c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// crashingWaveDistribute asks which tapped creatures get the counters.
func crashingWaveDistribute(ctx *Context) error {
	candidates := crashingWaveCandidates(ctx.Game, ctx.Controller())
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return crashingWaveStun(ctx, map[uuid.UUID]int{candidates[0]: 3})
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Crashing Wave — choose one, two or three tapped creatures your opponents control to distribute three stun counters among",
		Cards:    candidates,
		Min:      1,
		Max:      min(3, len(candidates)),
		ThenKey:  crashingWavePick,
		LiveZone: game.ZoneBattlefield,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

// crashingWavePick splits the three counters over the chosen creatures:
// three each take one, one takes all three, and two ask which takes two.
var crashingWavePick = askedPick("choose-cards/crashing-wave-stun-among", func(ctx *Context, r game.CardsChosen) error {
	picked := crashingWaveStillCandidates(ctx, r.Picked)
	switch len(picked) {
	case 0:
		return nil
	case 1:
		return crashingWaveStun(ctx, map[uuid.UUID]int{picked[0]: 3})
	case 2:
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Crashing Wave — choose the creature that gets two stun counters (the other gets one)",
			Cards:    picked,
			Min:      1,
			Max:      1,
			ThenKey:  crashingWaveTwo,
			Carry:    picked,
			LiveZone: game.ZoneBattlefield,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}
	shares := map[uuid.UUID]int{}
	for _, id := range picked[:3] {
		shares[id] = 1
	}
	return crashingWaveStun(ctx, shares)
})

// crashingWaveTwo is the split of two creatures: the one named takes
// two counters, the other one.
var crashingWaveTwo = askedPick("choose-cards/crashing-wave-stun-two", func(ctx *Context, r game.CardsChosen) error {
	pair := crashingWaveStillCandidates(ctx, r.Carry)
	if len(pair) == 0 {
		return nil
	}
	two := pair[0]
	if len(r.Picked) == 1 && slices.Contains(pair, r.Picked[0]) {
		two = r.Picked[0]
	}
	shares := map[uuid.UUID]int{two: 2}
	for _, id := range pair {
		if id != two {
			shares[id] = 1
		}
	}
	return crashingWaveStun(ctx, shares)
})

// crashingWaveStillCandidates keeps the chosen creatures that are still
// tapped creatures an opponent controls, in the order chosen, once each.
func crashingWaveStillCandidates(ctx *Context, chosen []uuid.UUID) []uuid.UUID {
	live := crashingWaveCandidates(ctx.Game, ctx.Controller())
	var out []uuid.UUID
	for _, id := range chosen {
		if slices.Contains(live, id) && !slices.Contains(out, id) {
			out = append(out, id)
		}
	}
	return out
}

// crashingWaveStun puts each creature's share of stun counters on it,
// in battlefield order so the log reads the same on every replay.
func crashingWaveStun(ctx *Context, shares map[uuid.UUID]int) error {
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		n := shares[c.InstanceID]
		if n <= 0 {
			continue
		}
		if err := (AddCounter{Target: c.InstanceID, Kind: game.CounterStun, N: n}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}
