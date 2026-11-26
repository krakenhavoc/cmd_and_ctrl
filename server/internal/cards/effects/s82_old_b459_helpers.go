package effects

import (
	"sort"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b459_helpers.go — New Phyrexia, batch 459 (S82, #5575).

// --- Hex Parasite ------------------------------------------------------------
//
// "{X}{B/P}: Remove up to X counters from target permanent. For each counter
// removed this way, this creature gets +1/+0 until end of turn."
//
// The player decides how many counters of each kind to take, one kind at a
// time in sorted order, never more than X in all; one chain of keyed number
// prompts (the answer carries its place in Words and Nums), so a table
// waiting on the answer is a restore point. Words is {the kind being asked},
// Nums {counters still allowed, counters removed so far} and Carry {the
// Parasite, the target}.

// HexParasiteRemoveCounters is the ability's whole effect.
// x is the announced X, read by the card.
func HexParasiteRemoveCounters(g *game.Game, item *game.StackItem, x int) error {
	ctx := NewContext(g, item)
	target, ok := B16FirstLegalTargetCard(ctx)
	if !ok || x <= 0 {
		return nil
	}
	return hexParasiteAsk(ctx, item.SourceCardID, target, "", x, 0)
}

// hexParasiteAsk asks about the first kind sorting after `after` that still
// has counters, or finishes when there is none or the budget is spent.
func hexParasiteAsk(ctx *Context, parasite, target uuid.UUID, after string, budget, removed int) error {
	if budget > 0 {
		if c, ok := ctx.Game.LookupCardForEffect(target); ok && OnBattlefield(ctx.Game, target) {
			var kinds []string
			for kind, n := range c.Counters {
				if n > 0 && kind > after {
					kinds = append(kinds, kind)
				}
			}
			if len(kinds) > 0 {
				sort.Strings(kinds)
				kind := kinds[0]
				return ChooseNumber{
					Min:      0,
					Max:      min(budget, c.Counters[kind]),
					Question: "Hex Parasite — how many " + kind + " counters do you remove?",
					Unit:     game.PayAmountCounters,
					Then:     hexParasiteStep(),
					Carry:    []uuid.UUID{parasite, target},
					Words:    []string{kind},
					Nums:     []int{budget, removed},
				}.Apply(ctx)
			}
		}
	}
	return hexParasitePump(ctx, parasite, removed)
}

// hexParasiteRemoveKey is the chain's number branch. The chain refers back
// to itself, so hexParasiteAsk looks the branch up by key when it runs
// rather than naming a package variable (which would be an initialization
// cycle), and the core keeps no init() (#5906).
const hexParasiteRemoveKey = "number/s82-old-b459-hex-parasite-remove"

var _ = NumberThen(hexParasiteRemoveKey, hexParasiteRemoved)

// hexParasiteStep is the registered branch.
func hexParasiteStep() game.NumberThen {
	then, _ := game.NumberThenOf(hexParasiteRemoveKey)
	return then
}

func hexParasiteRemoved(ctx *Context, r game.NumberAnswer) error {
	if len(r.Carry) < 2 || len(r.Words) < 1 || len(r.Nums) < 2 {
		return nil
	}
	parasite, target := r.Carry[0], r.Carry[1]
	kind, budget, removed := r.Words[0], r.Nums[0], r.Nums[1]
	if r.Number > 0 && OnBattlefield(ctx.Game, target) {
		got, err := SocB9RemoveCounters(ctx, target, kind, min(r.Number, budget))
		if err != nil {
			return err
		}
		budget -= got
		removed += got
	}
	return hexParasiteAsk(ctx, parasite, target, kind, budget, removed)
}

func hexParasitePump(ctx *Context, parasite uuid.UUID, removed int) error {
	if removed <= 0 || !OnBattlefield(ctx.Game, parasite) || ctx.IsNewSourceObject(parasite) {
		return nil
	}
	return BoostUntilEOT{Target: parasite, Power: removed, Label: "Hex Parasite — +1/+0 for each counter removed"}.Apply(ctx)
}

// --- Apostle's Blessing ------------------------------------------------------

// ApostlesBlessingProtection is "Target artifact or creature you control gains
// protection from artifacts or from the color of your choice until end of
// turn": the choice is made as the spell resolves, among artifacts and the five
// colours, and a target that left in response is asked nothing.
func ApostlesBlessingProtection(item *game.StackItem, ctx *Context) error {
	var target uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetCard {
			target = t.ID
		}
	}
	if target == uuid.Nil {
		return nil
	}
	options := []game.ChoiceOption{{Label: "Artifacts"}}
	for _, c := range game.AllColors {
		options = append(options, game.ChoiceOption{Label: game.ColorName(c)})
	}
	return PickOption{
		Question: "Apostle's Blessing — protection from what?",
		Options:  options,
		ThenKey:  apostlesBlessingPick,
		Carry:    []uuid.UUID{target},
	}.Apply(ctx)
}

var apostlesBlessingPick = OptionPickThen("option-pick/s82-old-b459-apostles-blessing", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 || r.Index > len(game.AllColors) || len(r.Carry) == 0 || !OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	token := "protection from artifacts"
	if r.Index > 0 {
		token = game.ProtectionFromColor(game.AllColors[r.Index-1])
	}
	if token == "" {
		return nil
	}
	return GrantKeywordUntilEOT{Target: r.Carry[0], Keywords: []string{token}, Label: "Apostle's Blessing — " + token}.Apply(ctx)
})

// --- small bodies ------------------------------------------------------------

// WarReportGain is "You gain life equal to the number of creatures on the
// battlefield plus the number of artifacts on the battlefield": an artifact
// creature counts once in each.
func WarReportGain(_ *game.StackItem, ctx *Context) error {
	n := 0
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.IsCreature() {
			n++
		}
		if c.IsArtifact() {
			n++
		}
	}
	return GainLife{Amount: n}.Apply(ctx)
}
