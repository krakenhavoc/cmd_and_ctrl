package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Urge to Feed — Instant {B}{B} (Innistrad: Crimson Vow Commander, S82):
//
//	"Target creature gets -3/-3 until end of turn. You may tap any number of
//	 untapped Vampire creatures you control. If you do, put a +1/+1 counter
//	 on each of those Vampires."
//
// The -3/-3 is layer 7c until cleanup (CR 514.2); a target that has gone
// by resolution makes the whole spell do nothing (CR 608.2b). The tap is
// Orphans of the Wheat's "tap any number" prompt over the controller's
// untapped Vampires, floor zero, asked as the spell resolves, so the
// shrink is already in place. Each pick is re-read as it is tapped: only a
// Vampire that really became tapped gets its counter ("each of those
// Vampires"). Tapping them is not a cost, so a creature that entered this
// turn may be tapped (CR 302.6 covers only {T} costs).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9f2f0ae-5869-43dc-a225-4dfe3d37c166",
		Name:         "Urge to Feed",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(HarmUntilEOT(0, -3, -3)),
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{Target: item.Targets[0].ID, Power: -3, Toughness: -3,
				Label: "Urge to Feed — -3/-3"}).Apply(ctx); err != nil {
				return err
			}
			var vampires []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsCreature() && !c.Tapped && c.HasSubtypeWord("Vampire") {
					vampires = append(vampires, c.InstanceID)
				}
			}
			if len(vampires) == 0 {
				return nil
			}
			ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
				Chooser:  item.Controller,
				Source:   item.SourceCardID,
				Question: "Urge to Feed — tap any number of untapped Vampires you control to put a +1/+1 counter on each",
				Cards:    vampires,
				Min:      0,
				Max:      0, // "any number"
				LiveZone: game.ZoneBattlefield,
				ThenKey:  urgeToFeedTap,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// urgeToFeedTap taps the picked Vampires and puts a counter on each one
// that really became tapped. An on-disk key: never renamed, never reused.
var urgeToFeedTap = askedPick("choose-cards/s82-rpr-b30-urge-to-feed-tap", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() || !c.HasSubtypeWord("Vampire") {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		if t, ok := ctx.Game.LookupCardForEffect(id); !ok || !t.Tapped {
			continue
		}
		if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
})
